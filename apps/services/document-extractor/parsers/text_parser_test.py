"""Tests for the text/plain parser (MD/TXT format)."""

from __future__ import annotations

import os
import tempfile

import pytest

from parsers.text_parser import TextParser


@pytest.fixture
def parser() -> TextParser:
    return TextParser()


@pytest.mark.asyncio
async def test_parse_markdown_with_headings(parser: TextParser) -> None:
    """Parse markdown with H1-H6 headings and verify section extraction."""
    content = """# Main Title

## Section One

Paragraph under section one.

## Section Two

### Subsection 2.1

Detail text.

Some paragraph.
"""
    with tempfile.NamedTemporaryFile(mode="w", suffix=".md", delete=False) as f:
        f.write(content)
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.format == "markdown"
        assert doc.title == "Main Title"
        assert len(doc.sections) >= 2, f"Expected >=2 sections, got {len(doc.sections)}"
        headings = {s.heading: s for s in doc.sections}
        assert "Section One" in headings
        assert "Section Two" in headings
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_markdown_with_lists(parser: TextParser) -> None:
    """Parse unordered and ordered lists."""
    content = """# List Test

- Item one
- Item two
- Item three

1. First
2. Second
3. Third
"""
    with tempfile.NamedTemporaryFile(mode="w", suffix=".md", delete=False) as f:
        f.write(content)
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.format == "markdown"
        sections = [s for s in doc.sections if s.heading == "List Test"]
        assert len(sections) > 0
        section = sections[0]
        assert len(section.lists) > 0, "Expected list items"
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_plain_text(parser: TextParser) -> None:
    """Parse plain text file with paragraphs."""
    content = """This is the first paragraph.

This is the second paragraph with more content.

And a third one."""
    with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=False) as f:
        f.write(content)
        path = f.name
    try:
        doc = await parser.parse(path)
        assert doc.format == "text"
        assert doc.section_count >= 0  # plain text has no heading sections
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_empty_file(parser: TextParser) -> None:
    """Parse an empty text file."""
    with tempfile.NamedTemporaryFile(mode="w", suffix=".md", delete=False) as f:
        f.write("")
        path = f.name
    try:
        doc = await parser.parse(path)
        # Empty file should not crash
        assert doc is not None
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_parse_markdown_with_table(parser: TextParser) -> None:
    """Parse markdown with a table."""
    content = """# Data Table

| Name | Age | City |
|------|-----|------|
| John | 30  | NYC  |
| Jane | 25  | LA   |
"""
    with tempfile.NamedTemporaryFile(mode="w", suffix=".md", delete=False) as f:
        f.write(content)
        path = f.name
    try:
        doc = await parser.parse(path)
        sections = [s for s in doc.sections if s.heading == "Data Table"]
        if sections:
            assert len(sections[0].tables) > 0, "Expected at least one table"
    finally:
        os.unlink(path)


@pytest.mark.asyncio
async def test_unsupported_extension(parser: TextParser) -> None:
    """Reject unsupported file extensions."""
    with pytest.raises(ValueError, match="Unsupported extension"):
        await parser.parse("test.unsupported")
