"""FastAPI routes for document extraction endpoints.

Provides:
- POST /api/v1/documents/extract — single file extraction (parse → LLM → validate)
- POST /api/v1/documents/extract/batch — batch upload with deduplication
"""

from __future__ import annotations

import json
import logging
import os
import tempfile
from typing import Any

from fastapi import APIRouter, File, Form, HTTPException, UploadFile
from pydantic import BaseModel

from config import settings
from llm.client import LlmClient
from llm.prompts import build_extraction_prompt
from parsers import parse_file as parse_document
from parsers.base import validate_file_signature
from sequence.validator import validate_llm_response

logger = logging.getLogger(settings.SERVICE_NAME)

router = APIRouter()

# ─── Request/Response Models ───────────────────────────────────────────────────


class ExtractResponse(BaseModel):
    """Response from a document extraction request."""

    filename: str = ""
    format: str = ""
    file_size_bytes: int = 0
    parsed_sections: int = 0
    raw_text_length: int = 0
    sequence: list[dict] = []
    step_count: int = 0
    warnings: list[str] = []
    error: str | None = None


class ApplyRequest(BaseModel):
    """Request to apply a validated sequence to the ontology service."""

    ontology_id: str
    branch_id: str = "main"
    steps: list[dict] = []
    commit_message: str = "AI-assisted ontology extraction"
    source_filename: str | None = None


class ApplyResponse(BaseModel):
    """Response from applying a sequence."""

    success: bool = True
    commit_id: str = ""
    steps_applied: int = 0
    steps_skipped: int = 0
    steps_failed: int = 0
    errors: list[str] = []
    message: str = ""


class BatchExtractRequest(BaseModel):
    """Request body for batch extraction (re-parsing stored files)."""

    filenames: list[str] = []


class BatchExtractResponse(BaseModel):
    """Response from a batch extraction request."""

    results: list[ExtractResponse] = []
    total_files: int = 0
    successful: int = 0
    failed: int = 0


# ─── Helpers ───────────────────────────────────────────────────────────────────


def _get_file_extension(filename: str) -> str:
    """Get the lowercase file extension from a filename."""
    _, ext = os.path.splitext(filename)
    return ext.lower()


def _is_supported_format(filename: str) -> bool:
    """Check if a file format is supported for extraction."""
    try:
        from parsers import get_parser

        get_parser(filename)
        return True
    except ValueError:
        return False


SUPPORTED_FORMATS = {
    ".md",
    ".txt",
    ".markdown",
    ".pdf",
    ".docx",
    ".docm",
    ".json",
    ".xml",
    ".csv",
    ".tsv",
    ".xlsx",
    ".xlsm",
}


# ─── Routes ────────────────────────────────────────────────────────────────────


@router.post("/documents/extract", response_model=ExtractResponse)
async def extract_document(
    file: UploadFile = File(...),
    ontology_context: str | None = Form(None),
):
    """Extract ontology structure from a single uploaded document.

    The pipeline:
    1. Save uploaded file to temp location
    2. Parse the file using the appropriate format parser
    3. Send parsed content to LLM for ontology extraction
    4. Validate the LLM response sequence
    5. Return the validated sequence as a preview
    """
    if not file.filename:
        raise HTTPException(status_code=400, detail="No file provided")

    ext = _get_file_extension(file.filename)
    if ext not in SUPPORTED_FORMATS:
        raise HTTPException(
            status_code=400,
            detail=f"Unsupported file format '{ext}'. "
            f"Supported: {', '.join(sorted(SUPPORTED_FORMATS))}",
        )

    # Check file size
    file_size = 0
    content_chunks: list[bytes] = []
    async for chunk in file.file:
        content_chunks.append(chunk)
        file_size += len(chunk)

    if file_size == 0:
        raise HTTPException(status_code=400, detail="Empty file")

    max_bytes = settings.MAX_FILE_SIZE_MB * 1024 * 1024
    if file_size > max_bytes:
        return ExtractResponse(
            filename=file.filename,
            error=f"File size ({file_size / 1024 / 1024:.1f} MB) exceeds "
            f"maximum of {settings.MAX_FILE_SIZE_MB} MB for synchronous processing. "
            "Use async mode for larger files.",
        )

    # Save to temp file for parsing
    content_bytes = b"".join(content_chunks)
    with tempfile.NamedTemporaryFile(delete=False, suffix=ext) as tmp:
        tmp.write(content_bytes)
        tmp_path = tmp.name

    try:
        # Step 0: Magic bytes validation
        logger.debug("Validating file signature: %s", file.filename)
        try:
            validate_file_signature(tmp_path, ext)
        except ValueError as exc:
            logger.warning("File signature mismatch for %s: %s", file.filename, exc)
            return ExtractResponse(
                filename=file.filename,
                error=f"File signature mismatch: {exc}",
            )

        # Step 1: Parse
        logger.debug("Parsing file: %s", file.filename)
        parsed = await parse_document(tmp_path)

        if parsed.error:
            logger.warning("Parse warning for %s: %s", file.filename, parsed.error)
            return ExtractResponse(
                filename=file.filename,
                format=parsed.format,
                file_size_bytes=file_size,
                error=parsed.error,
            )

        # Step 2: Prepare content for LLM
        raw_text = _parsed_to_text(parsed)
        if not raw_text.strip():
            return ExtractResponse(
                filename=file.filename,
                format=parsed.format,
                file_size_bytes=file_size,
                error="No extractable text content found in document",
            )

        # Step 3: LLM extraction with retry
        sections_dict = []
        for s in parsed.sections:
            sections_dict.append(
                {
                    "heading": s.heading,
                    "level": s.level,
                    "paragraphs": s.paragraphs,
                    "tables": [t.model_dump() for t in s.tables],
                }
            )

        messages = build_extraction_prompt(
            content=raw_text,
            title=parsed.title or file.filename or "Untitled",
            fmt=parsed.format,
            sections=sections_dict,
        )

        llm_response = ""
        attempts = 0
        validation_result = None
        async with LlmClient() as llm:
            for attempt in range(1, settings.LLM_MAX_RETRIES + 1):
                try:
                    llm_response, attempts = await llm.extract_with_retry(
                        messages,
                        temperature=0.3,
                        max_tokens=4096,
                    )

                    # Step 4: Validate
                    validation_result = validate_llm_response(llm_response)
                    if validation_result.valid:
                        break

                    logger.warning(
                        "Validation failed on attempt %d: %s",
                        attempt,
                        "; ".join(validation_result.errors),
                    )

                    # Update messages for retry with validation feedback
                    messages = build_extraction_prompt(
                        content=raw_text,
                        title=parsed.title or file.filename or "Untitled",
                        fmt=parsed.format,
                        sections=sections_dict,
                        retry_count=attempt,
                        validation_errors=validation_result.errors,
                        validation_warnings=validation_result.warnings,
                    )

                except Exception as exc:
                    logger.error("LLM extraction attempt %d failed: %s", attempt, exc)
                    if attempt == settings.LLM_MAX_RETRIES:
                        return ExtractResponse(
                            filename=file.filename,
                            format=parsed.format,
                            file_size_bytes=file_size,
                            parsed_sections=parsed.section_count,
                            raw_text_length=len(raw_text),
                            error=f"LLM extraction failed after {attempt} attempts: {exc}",
                        )

        # Build response
        steps = []
        if validation_result and validation_result.steps:
            steps = [s.model_dump() for s in validation_result.steps]

        warnings = (validation_result.warnings if validation_result else []) + (
            [] if validation_result and validation_result.valid else ["Not all steps validated"]
        )

        logger.info(
            "Extraction complete: file=%s steps=%d warnings=%d attempts=%d",
            file.filename,
            len(steps),
            len(warnings),
            attempts,
        )

        return ExtractResponse(
            filename=file.filename,
            format=parsed.format,
            file_size_bytes=file_size,
            parsed_sections=parsed.section_count,
            raw_text_length=len(raw_text),
            sequence=steps,
            step_count=len(steps),
            warnings=warnings,
        )

    except Exception as exc:
        logger.error("Extraction failed for %s: %s", file.filename, exc, exc_info=True)
        raise HTTPException(
            status_code=500,
            detail=f"Extraction failed: {exc}",
        )
    finally:
        # Clean up temp file
        try:
            os.unlink(tmp_path)
        except Exception:
            pass


@router.post("/documents/apply", response_model=ApplyResponse)
async def apply_sequence(request: ApplyRequest):
    """Apply a validated ontology sequence to the ontology-service via gRPC."""
    if not request.steps:
        raise HTTPException(status_code=400, detail="No steps provided in sequence")

    logger.info(
        "Applying sequence: ontology=%s branch=%s steps=%d",
        request.ontology_id,
        request.branch_id,
        len(request.steps),
    )

    try:
        from grpc_client.ontology import get_client

        client = await get_client()
        result = await client.apply_sequence(
            ontology_id=request.ontology_id,
            branch_id=request.branch_id,
            steps=request.steps,
            commit_message=request.commit_message,
            source_filename=request.source_filename,
        )

        success = result.get("steps_failed", 0) == 0
        return ApplyResponse(
            success=success,
            commit_id=result.get("commit_id", ""),
            steps_applied=result.get("steps_applied", 0),
            steps_skipped=result.get("steps_skipped", 0),
            steps_failed=result.get("steps_failed", 0),
            errors=result.get("errors", []),
            message=f"Applied {result.get('steps_applied', 0)} steps"
            if success
            else f"Failed {result.get('steps_failed', 0)} steps",
        )

    except Exception as exc:
        logger.error("Apply sequence failed: %s", exc)
        return ApplyResponse(
            success=False,
            errors=[str(exc)],
            message=f"Failed to apply sequence: {exc}",
        )


@router.post("/documents/validate", response_model=ExtractResponse)
async def validate_sequence_only(sequence_json: str = Form(...)):
    """Validate a sequence JSON without parsing or LLM extraction.

    Useful for testing and debugging sequence formats.
    """
    try:
        data = json.loads(sequence_json)
    except json.JSONDecodeError as exc:
        raise HTTPException(status_code=400, detail=f"Invalid JSON: {exc}")

    # Normalize to list
    if isinstance(data, dict):
        steps_data = data.get("steps", data.get("sequence", []))
    elif isinstance(data, list):
        steps_data = data
    else:
        raise HTTPException(status_code=400, detail="Expected JSON array or object with 'steps'")

    from sequence.validator import validate_sequence

    result = validate_sequence(steps_data)

    return ExtractResponse(
        filename="(validation)",
        format="json",
        sequence=[s.model_dump() for s in result.steps],
        step_count=len(result.steps),
        warnings=result.warnings,
        error="; ".join(result.errors) if result.errors else None,
    )


# ─── Internal Helpers ──────────────────────────────────────────────────────────


def _parsed_to_text(parsed: Any) -> str:
    """Convert a ParsedDocument to plain text for LLM consumption."""
    parts: list[str] = []

    if parsed.title:
        parts.append(f"# {parsed.title}")

    for section in parsed.sections:
        if section.heading:
            level = min(section.level, 6)
            parts.append(f"{'#' * level} {section.heading}")

        for para in section.paragraphs:
            parts.append(para)

        for table in section.tables:
            table_lines: list[str] = []
            for row in table.rows:
                cells = [c.text for c in row.cells]
                table_lines.append(" | ".join(cells))
            if table_lines:
                parts.append("\n" + "\n".join(table_lines))

        for lst in section.lists:
            prefix = "1. " if lst.ordered else "- "
            parts.append(f"{prefix}{lst.text}")
            for child in lst.children:
                parts.append(f"  - {child.text}")

        for code_block in section.code_blocks:
            parts.append(f"\n```\n{code_block}\n```\n")

    return "\n\n".join(parts)
