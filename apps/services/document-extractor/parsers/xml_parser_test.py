"""Tests for the XML parser."""

from __future__ import annotations

import os
import tempfile

import pytest

from parsers.xml_parser import XmlParser


@pytest.fixture
def parser() -> XmlParser:
    return XmlParser()


@pytest.mark.asyncio
async def test_parse_simple_xml(parser: XmlParser) -> None:
    """Parse a simple XML document."""
    content = """<?xml version="1.0"?>
<root>
    <person id="1">
        <name>John</name>
        <age>30</age>
    </person>
    <person id="2">
        <name>Jane</name>
        <age>25</age>
    </person>
</root>"""
    with tempfile.NamedTemporaryFile(mode="w", suffix=".xml", delete=False) as f:
        f.write(content)
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.format == "xml"
        assert len(doc.sections) > 0
        assert doc.metadata.get("root_tag") is not None
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_xml_with_namespaces(parser: XmlParser) -> None:
    """Parse XML with namespace declarations."""
    content = """<?xml version="1.0"?>
<root xmlns:ns="http://example.com/ns">
    <ns:item id="1">
        <ns:name>Item One</ns:name>
    </ns:item>
</root>"""
    with tempfile.NamedTemporaryFile(mode="w", suffix=".xml", delete=False) as f:
        f.write(content)
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.format == "xml"
        assert doc.metadata.get("namespace_count", 0) > 0
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_invalid_xml(parser: XmlParser) -> None:
    """Attempt to parse malformed XML."""
    with tempfile.NamedTemporaryFile(mode="w", suffix=".xml", delete=False) as f:
        f.write("<root><unclosed>")
        path = f.name
    try:
        with pytest.raises(Exception):
            await parser.parse(path)
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_empty_xml(parser: XmlParser) -> None:
    """Parse an XML document with only root."""
    content = """<?xml version="1.0"?>
<root/>"""
    with tempfile.NamedTemporaryFile(mode="w", suffix=".xml", delete=False) as f:
        f.write(content)
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.format == "xml"
    finally:
        os.unlink(path)
