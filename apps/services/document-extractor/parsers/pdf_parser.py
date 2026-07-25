"""Parser for PDF documents using pdfplumber.

Extracts text content with structure detection (font size for headings),
table detection, password detection, and scanned document detection.
"""

from __future__ import annotations

import logging
import os

from parsers.base import BaseParser, ParseError
from parsers.models import ParsedDocument, Section, Table, TableCell, TableRow

logger = logging.getLogger("document-extractor.parsers.pdf")

# Minimum font size ratio to consider a text block as heading
HEADING_SIZE_RATIO = 1.15


class PdfParser(BaseParser):
    """Parser for PDF files using pdfplumber."""

    SUPPORTED_EXTENSIONS = [".pdf"]

    async def parse(self, file_path: str) -> ParsedDocument:
        """Parse a PDF file and extract structured content."""
        self.validate_extension(file_path)
        file_size = os.path.getsize(file_path)

        logger.debug(
            "Parsing PDF: %s (size=%d bytes)",
            os.path.basename(file_path),
            file_size,
        )

        # Import pdfplumber here to avoid startup dependency issues
        try:
            import pdfplumber
        except ImportError as exc:
            raise ParseError(
                "pdfplumber not installed; run 'pip install pdfplumber'",
                original=exc,
            )

        try:
            with pdfplumber.open(file_path) as pdf:
                # Check for encrypted PDF
                if hasattr(pdf, "metadata") and pdf.metadata:
                    if pdf.metadata.get("Encrypted") is True:
                        logger.warning("Password-protected PDF: %s", file_path)
                        return ParsedDocument(
                            title=os.path.basename(file_path),
                            format="pdf",
                            sections=[],
                            metadata={
                                "file_size_bytes": file_size,
                                "encrypted": True,
                                "page_count": len(pdf.pages),
                            },
                            error="Password-protected PDF — text extraction not possible",
                        )

                page_count = len(pdf.pages)

                if page_count == 0:
                    return ParsedDocument(
                        title=os.path.basename(file_path),
                        format="pdf",
                        sections=[],
                        metadata={"file_size_bytes": file_size, "page_count": 0, "empty": True},
                    )

                # Extract text from all pages
                sections: list[Section] = []
                current_section = Section(heading="", level=0)
                current_paragraphs: list[str] = []

                # Determine base font size from first page
                base_font_size = 12.0
                try:
                    first_page_text = pdf.pages[0]
                    chars = first_page_text.chars
                    if chars:
                        sizes = [c.get("size", 12.0) for c in chars if c.get("size")]
                        if sizes:
                            base_font_size = max(sizes) if max(sizes) > 12.0 else 12.0
                except Exception:
                    pass

                for page_num, page in enumerate(pdf.pages):
                    # Extract tables
                    tables = page.extract_tables()
                    if tables:
                        page_tables = []
                        for table_data in tables:
                            table = Table()
                            for i, row_data in enumerate(table_data):
                                row = TableRow(is_header=(i == 0))
                                for cell in row_data:
                                    row.cells.append(TableCell(text=str(cell or "")))
                                table.rows.append(row)
                            page_tables.append(table)

                    # Extract text with structure
                    text = page.extract_text()
                    if not text:
                        continue

                    lines = text.split("\n")
                    for line in lines:
                        stripped = line.strip()
                        if not stripped:
                            continue

                        # Detect heading by font size analysis (approximate)
                        is_heading = False
                        heading_level = 1

                        try:
                            chars = page.chars
                            line_chars = [
                                c
                                for c in chars
                                if abs(c.get("top", 0) - chars[0].get("top", 0)) < 5
                                if c.get("text", "").strip() == stripped[:30]
                            ]
                            # Simplified: check char sizes near this line
                            matching = [
                                c
                                for c in chars
                                if c.get("text", "").strip()
                                and stripped.startswith(c.get("text", "").strip())
                            ]
                            if matching:
                                avg_size = sum(c.get("size", 10) for c in matching) / len(matching)
                                if avg_size >= base_font_size * HEADING_SIZE_RATIO:
                                    is_heading = True
                                    if avg_size >= base_font_size * 1.5:
                                        heading_level = 1
                                    elif avg_size >= base_font_size * 1.3:
                                        heading_level = 2
                                    else:
                                        heading_level = 3
                        except Exception:
                            pass

                        if is_heading:
                            # Flush current paragraphs into section
                            if current_paragraphs:
                                if current_section.heading:
                                    current_section.paragraphs = current_paragraphs
                                    sections.append(current_section)
                                current_paragraphs = []

                            current_section = Section(heading=stripped, level=heading_level)
                        # Shorter lines may be headings without font data
                        elif (
                            len(stripped) < 80
                            and not stripped.endswith(".")
                            and stripped.isupper()
                            and len(stripped) > 3
                        ):
                            if current_paragraphs:
                                if current_section.heading:
                                    current_section.paragraphs = current_paragraphs
                                    sections.append(current_section)
                                current_paragraphs = []
                            current_section = Section(heading=stripped, level=2)
                        else:
                            current_paragraphs.append(stripped)

                # Flush last section
                if current_paragraphs:
                    if current_section.heading:
                        current_section.paragraphs = current_paragraphs
                        sections.append(current_section)
                    elif not sections:
                        # No headings found — all content in root section
                        root_section = Section(
                            heading="",
                            level=0,
                            paragraphs=current_paragraphs,
                        )
                        sections.append(root_section)

                # Check if no text was extracted (scanned document)
                total_chars = sum(len(p.extract_text() or "") for p in pdf.pages)
                if total_chars == 0 and page_count > 0:
                    logger.warning("Scanned PDF detected (no text layer): %s", file_path)
                    return ParsedDocument(
                        title=os.path.basename(file_path),
                        format="pdf",
                        sections=[],
                        metadata={
                            "file_size_bytes": file_size,
                            "page_count": page_count,
                            "scanned": True,
                        },
                        error="Scanned PDF — no extractable text layer",
                    )

        except ParseError:
            raise
        except Exception as exc:
            raise ParseError(f"Failed to parse PDF {file_path}: {exc}", original=exc)

        title = os.path.splitext(os.path.basename(file_path))[0]
        logger.debug(
            "Parsed PDF %s: %d pages, %d sections",
            os.path.basename(file_path),
            page_count,
            len(sections),
        )

        return ParsedDocument(
            title=title,
            format="pdf",
            sections=sections,
            metadata={
                "file_size_bytes": file_size,
                "page_count": page_count,
                "encrypted": False,
                "scanned": False,
            },
        )
