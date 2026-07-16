"""Tests for the DOCX parser (python-docx-based)."""

from __future__ import annotations

import pytest

from parsers.base import ParseError
from parsers.docx_parser import DocxParser


@pytest.fixture
def parser() -> DocxParser:
    return DocxParser()


@pytest.mark.asyncio
async def test_parse_non_docx_file(parser: DocxParser) -> None:
    """Attempting to parse a non-DOCX file should raise ParseError."""
    with pytest.raises((ParseError, ValueError)):
        await parser.parse("/tmp/nonexistent.docx")


@pytest.mark.asyncio
async def test_unsupported_extension(parser: DocxParser) -> None:
    """Reject unsupported file extensions."""
    with pytest.raises(ValueError, match="Unsupported extension"):
        await parser.parse("test.pdf")


@pytest.mark.asyncio
async def test_validate_extension(parser: DocxParser) -> None:
    """Validate that only .docx/.docm files pass extension check."""
    # Should not raise
    parser.validate_extension("test.docx")
    parser.validate_extension("document.DOCX")
    parser.validate_extension("macro.docm")
    # Should raise
    with pytest.raises(ValueError):
        parser.validate_extension("test.txt")
