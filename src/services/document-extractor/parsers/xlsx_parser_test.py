"""Tests for the XLSX parser (openpyxl-based)."""

from __future__ import annotations

import pytest

from parsers.xlsx_parser import XlsxParser


@pytest.fixture
def parser() -> XlsxParser:
    return XlsxParser()


@pytest.mark.asyncio
async def test_parse_non_xlsx_file(parser: XlsxParser) -> None:
    """Attempting to parse a non-XLSX file should raise an error."""
    with pytest.raises(Exception):
        await parser.parse("/tmp/nonexistent.xlsx")


@pytest.mark.asyncio
async def test_unsupported_extension(parser: XlsxParser) -> None:
    """Reject unsupported file extensions."""
    with pytest.raises(ValueError, match="Unsupported extension"):
        await parser.parse("test.csv")


@pytest.mark.asyncio
async def test_validate_extension(parser: XlsxParser) -> None:
    """Validate that only .xlsx/.xlsm files pass extension check."""
    # Should not raise
    parser.validate_extension("test.xlsx")
    parser.validate_extension("macro.xlsm")
    # Should raise
    with pytest.raises(ValueError):
        parser.validate_extension("test.xml")
    with pytest.raises(ValueError):
        parser.validate_extension("test.csv")


@pytest.mark.asyncio
async def test_ontology_column_detection(parser: XlsxParser) -> None:
    """Test ontology column pattern detection logic directly."""
    headers = ["Class", "Parent", "Property", "Type"]
    mapping = parser._detect_ontology_columns(headers)
    assert "class" in mapping
    assert mapping["class"] == 0
    assert "parent" in mapping
    assert "property" in mapping
    assert mapping["type"] == 3


@pytest.mark.asyncio
async def test_column_detection_case_insensitive(parser: XlsxParser) -> None:
    """Test that column detection is case-insensitive."""
    headers = ["CLASS", "PARENT CLASS", "PROPERTY NAME", "DATATYPE"]
    mapping = parser._detect_ontology_columns(headers)
    assert "class" in mapping


@pytest.mark.asyncio
async def test_column_detection_no_match(parser: XlsxParser) -> None:
    """Test that non-matching headers return empty mapping."""
    headers = ["A", "B", "C", "D"]
    mapping = parser._detect_ontology_columns(headers)
    assert mapping == {}
