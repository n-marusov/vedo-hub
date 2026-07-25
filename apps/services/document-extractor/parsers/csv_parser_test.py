"""Tests for the CSV parser."""

from __future__ import annotations

import os
import tempfile

import pytest

from parsers.csv_parser import CsvParser


@pytest.fixture
def parser() -> CsvParser:
    return CsvParser()


@pytest.mark.asyncio
async def test_parse_simple_csv(parser: CsvParser) -> None:
    """Parse a simple comma-delimited CSV file."""
    content = "name,age,city\nJohn,30,NYC\nJane,25,LA\nBob,35,Chicago\n"
    with tempfile.NamedTemporaryFile(mode="w", suffix=".csv", delete=False) as f:
        f.write(content)
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.format == "csv"
        assert doc.metadata.get("data_rows", 0) == 3
        assert doc.metadata.get("columns", 0) == 3
        assert doc.metadata.get("delimiter") == ","
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_semicolon_csv(parser: CsvParser) -> None:
    """Parse a semicolon-delimited CSV file (common in EU locales)."""
    content = "name;age;city\nJohn;30;NYC\nJane;25;LA\n"
    with tempfile.NamedTemporaryFile(mode="w", suffix=".csv", delete=False) as f:
        f.write(content)
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.format == "csv"
        assert doc.metadata.get("delimiter") == ";"
        assert doc.metadata.get("data_rows", 0) == 2
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_tab_csv(parser: CsvParser) -> None:
    """Parse a tab-delimited file (.tsv)."""
    content = "name\tage\tcity\nJohn\t30\tNYC\nJane\t25\tLA\n"
    with tempfile.NamedTemporaryFile(mode="w", suffix=".tsv", delete=False) as f:
        f.write(content)
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.format == "csv"
        assert doc.metadata.get("data_rows", 0) >= 2
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_empty_csv(parser: CsvParser) -> None:
    """Parse an empty CSV file."""
    with tempfile.NamedTemporaryFile(mode="w", suffix=".csv", delete=False) as f:
        f.write("")
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.metadata.get("empty") is True
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_csv_with_only_headers(parser: CsvParser) -> None:
    """Parse CSV with headers but no data rows."""
    content = "name,age,city\n"
    with tempfile.NamedTemporaryFile(mode="w", suffix=".csv", delete=False) as f:
        f.write(content)
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.format == "csv"
        assert doc.metadata.get("data_rows", 0) == 0
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_csv_with_quoted_fields(parser: CsvParser) -> None:
    """Parse CSV with quoted fields containing commas."""
    content = 'name,description\nJohn,"Engineer, Senior"\nJane,"Designer, Lead"\n'
    with tempfile.NamedTemporaryFile(mode="w", suffix=".csv", delete=False) as f:
        f.write(content)
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.format == "csv"
        assert doc.metadata.get("data_rows", 0) == 2
    finally:
        os.unlink(path)
