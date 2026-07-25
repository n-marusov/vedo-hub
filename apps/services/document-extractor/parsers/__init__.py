"""Parsers package — document format parsers for the document-extractor service.

Provides a registry for file extension → parser lookup.
"""

from __future__ import annotations

import logging
from typing import TYPE_CHECKING

from parsers.base import BaseParser, ParseError

if TYPE_CHECKING:
    from parsers.models import ParsedDocument

logger = logging.getLogger("document-extractor.parsers")

# Lazy-loaded parser registry
_parsers: dict[str, BaseParser] | None = None


def _init_registry() -> dict[str, BaseParser]:
    """Initialize the parser registry mapping extensions to parser instances."""
    from parsers.csv_parser import CsvParser
    from parsers.docx_parser import DocxParser
    from parsers.json_parser import JsonParser
    from parsers.pdf_parser import PdfParser
    from parsers.text_parser import TextParser
    from parsers.xlsx_parser import XlsxParser
    from parsers.xml_parser import XmlParser

    registry: dict[str, BaseParser] = {}
    parsers = [
        TextParser(),
        PdfParser(),
        DocxParser(),
        JsonParser(),
        XmlParser(),
        CsvParser(),
        XlsxParser(),
    ]

    for parser in parsers:
        for ext in parser.SUPPORTED_EXTENSIONS:
            registry[ext.lower()] = parser

    logger.debug("Parser registry initialized: %d extensions mapped", len(registry))
    return registry


def get_parser(file_path: str) -> BaseParser:
    """Get the appropriate parser for a file based on its extension.

    Args:
        file_path: Path to the file to parse.

    Returns:
        A BaseParser instance for the file format.

    Raises:
        ValueError: If no parser is available for the file extension.
    """
    global _parsers
    if _parsers is None:
        _parsers = _init_registry()

    import os

    ext = os.path.splitext(file_path)[1].lower()
    parser = _parsers.get(ext)
    if parser is None:
        supported = sorted(set(_parsers.keys()))
        raise ValueError(
            f"No parser available for extension '{ext}'. Supported: {', '.join(supported)}"
        )
    return parser


async def parse_file(file_path: str) -> ParsedDocument:
    """Parse a file by auto-detecting its format.

    Args:
        file_path: Path to the file.

    Returns:
        ParsedDocument with extracted content.

    Raises:
        ValueError: Unsupported format.
        ParseError: On parse failures.
    """
    parser = get_parser(file_path)
    logger.debug("Parsing %s with %s", file_path, type(parser).__name__)
    return await parser.parse(file_path)


__all__ = [
    "BaseParser",
    "ParseError",
    "get_parser",
    "parse_file",
]
