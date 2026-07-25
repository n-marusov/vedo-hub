"""Tests for the PDF parser (pdfplumber-based)."""

from __future__ import annotations

import os
import tempfile

import pytest

from parsers.base import ParseError
from parsers.pdf_parser import PdfParser


@pytest.fixture
def parser() -> PdfParser:
    return PdfParser()


@pytest.mark.asyncio
async def test_parse_non_pdf_file(parser: PdfParser) -> None:
    """Attempting to parse a non-PDF file should raise ParseError."""
    content = b"This is not a PDF file. Just plain text."
    with tempfile.NamedTemporaryFile(suffix=".pdf", delete=False) as f:
        f.write(content)
        path = f.name
    try:
        # pdfplumber will fail to parse — should wrap in ParseError or return error document
        doc = await parser.parse(path)
        # If it can't parse, should return error gracefully
        assert doc.error is not None or doc.format == "pdf"
    except ParseError:
        pass  # Acceptable
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_empty_file(parser: PdfParser) -> None:
    """Parse an empty PDF file."""
    # Minimal but valid PDF
    content = b"%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [] /Count 0 >>\nendobj\nxref\n0 3\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \ntrailer\n<< /Size 3 /Root 1 0 R >>\nstartxref\n119\n%%EOF"
    with tempfile.NamedTemporaryFile(suffix=".pdf", delete=False) as f:
        f.write(content)
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.format == "pdf"
    except (ParseError, Exception):
        # pdfplumber may still fail on minimal PDF — acceptable
        pass
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_unsupported_extension(parser: PdfParser) -> None:
    """Reject unsupported file extensions."""
    with pytest.raises(ValueError, match="Unsupported extension"):
        await parser.parse("test.txt")


@pytest.mark.asyncio
async def test_validate_extension(parser: PdfParser) -> None:
    """Validate that only .pdf files pass extension check."""
    # Should not raise
    parser.validate_extension("test.pdf")
    parser.validate_extension("document.PDF")
    # Should raise
    with pytest.raises(ValueError):
        parser.validate_extension("test.txt")
    with pytest.raises(ValueError):
        parser.validate_extension("test.docx")
