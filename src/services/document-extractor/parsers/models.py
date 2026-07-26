"""Pydantic models for parsed document representation.

All parsers return a uniform ParsedDocument with sections preserving hierarchy
for downstream LLM processing.
"""

from __future__ import annotations

from typing import Any

from pydantic import BaseModel, Field


class TableCell(BaseModel):
    """A single cell in a table row."""

    text: str = ""
    colspan: int = 1
    rowspan: int = 1


class TableRow(BaseModel):
    """A row in a table."""

    cells: list[TableCell] = Field(default_factory=list)
    is_header: bool = False


class Table(BaseModel):
    """A table parsed from a document."""

    caption: str = ""
    rows: list[TableRow] = Field(default_factory=list)


class ListItem(BaseModel):
    """A list item."""

    text: str = ""
    level: int = 0
    ordered: bool = False
    number: int = 0
    children: list[ListItem] = Field(default_factory=list)


class Section(BaseModel):
    """A hierarchical section within a parsed document."""

    heading: str = ""
    level: int = 1
    paragraphs: list[str] = Field(default_factory=list)
    tables: list[Table] = Field(default_factory=list)
    lists: list[ListItem] = Field(default_factory=list)
    code_blocks: list[str] = Field(default_factory=list)
    subsections: list[Section] = Field(default_factory=list)


class ParsedDocument(BaseModel):
    """Uniform parsed document output from all parsers."""

    title: str = ""
    format: str = ""
    sections: list[Section] = Field(default_factory=list)
    metadata: dict[str, Any] = Field(default_factory=dict)
    """Arbitrary metadata extracted from the source (author, date,
    file_size_bytes, page_count, etc.)."""

    error: str | None = None
    """If set, parsing encountered a non-fatal issue."""

    @property
    def section_count(self) -> int:
        """Total number of non-empty sections (recursive)."""
        count = 0

        def _count(sections: list[Section]) -> None:
            nonlocal count
            for s in sections:
                if s.heading:
                    count += 1
                _count(s.subsections)

        _count(self.sections)
        return count

    @property
    def total_paragraphs(self) -> int:
        """Total paragraphs across all sections."""
        total = 0

        def _count(sections: list[Section]) -> None:
            nonlocal total
            for s in sections:
                total += len(s.paragraphs)
                _count(s.subsections)

        _count(self.sections)
        return total

    @property
    def total_tables(self) -> int:
        """Total tables across all sections."""
        total = 0

        def _count(sections: list[Section]) -> None:
            nonlocal total
            for s in sections:
                total += len(s.tables)
                _count(s.subsections)

        _count(self.sections)
        return total
