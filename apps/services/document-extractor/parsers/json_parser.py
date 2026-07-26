"""Parser for JSON documents.

Traverses JSON recursively, inferring types from values:
- string → datatype property
- number → datatype property
- object → class with nested properties
- array → check element types for property range
"""

from __future__ import annotations

import json
import logging
import os
from typing import Any

from parsers.base import BaseParser, ParseError
from parsers.models import ParsedDocument, Section

logger = logging.getLogger("document-extractor.parsers.json")


class JsonParser(BaseParser):
    """Parser for JSON (.json) files — ontology structure inference."""

    SUPPORTED_EXTENSIONS = [".json"]

    async def parse(self, file_path: str) -> ParsedDocument:
        """Parse a JSON file and infer ontology structure."""
        self.validate_extension(file_path)
        file_size = os.path.getsize(file_path)

        logger.debug(
            "Parsing JSON: %s (size=%d bytes)",
            os.path.basename(file_path),
            file_size,
        )

        try:
            with open(file_path, encoding="utf-8") as f:
                data = json.load(f)
        except json.JSONDecodeError as exc:
            raise ParseError(f"Invalid JSON in {file_path}: {exc}", original=exc)
        except Exception as exc:
            raise ParseError(f"Failed to read JSON file {file_path}: {exc}", original=exc)

        # Extract structure from JSON
        structure_lines: list[str] = []
        self._traverse(data, structure_lines, depth=0)

        filename = os.path.basename(file_path)
        section = Section(
            heading=f"JSON Structure: {filename}",
            level=1,
            paragraphs=structure_lines,
        )

        logger.debug(
            "Parsed JSON %s: inferred %d structural elements",
            filename,
            len(structure_lines),
        )

        return ParsedDocument(
            title=os.path.splitext(filename)[0].replace("_", " ").replace("-", " ").title(),
            format="json",
            sections=[section],
            metadata={
                "file_size_bytes": file_size,
                "top_level_type": type(data).__name__,
                "element_count": len(structure_lines),
                "root_is_array": isinstance(data, list),
            },
        )

    def _traverse(self, value: Any, lines: list[str], depth: int = 0) -> str:
        """Traverse a JSON value and infer ontology structure.

        Returns the inferred type name for property range resolution.
        """
        indent = "  " * depth

        if isinstance(value, dict):
            if not value:
                lines.append(f"{indent}(empty object)")
                return "object"

            # First pass: determine the type name from context
            type_name = value.get("type") or value.get("class") or value.get("$type", "Object")
            inferred_type = type_name if isinstance(type_name, str) else "Object"

            lines.append(f"{indent}Class: {inferred_type}")

            for key, val in value.items():
                # Skip metadata keys
                if key in ("type", "class", "$type", "id", "@id", "ontology_id"):
                    lines.append(f"{indent}  Property (datatype): {key} — {type(val).__name__}")
                    continue

                if isinstance(val, dict):
                    sub_type = self._traverse(val, lines, depth + 1)
                    lines.append(f"{indent}  Property (object): {key} → {sub_type}")
                elif isinstance(val, list):
                    if val:
                        elem_type = self._traverse(val[0], lines, depth + 1)
                        lines.append(f"{indent}  Property (collection): {key} → {elem_type}[]")
                    else:
                        lines.append(f"{indent}  Property (datatype): {key} — array (empty)")
                elif val is None:
                    lines.append(f"{indent}  Property (datatype): {key} — null")
                else:
                    lines.append(f"{indent}  Property (datatype): {key} — {type(val).__name__}")

            return inferred_type

        if isinstance(value, list):
            if not value:
                lines.append(f"{indent}(empty array)")
                return "array"

            # Check if all elements have the same type
            types = {type(v).__name__ for v in value}
            if len(types) == 1:
                elem_type = next(iter(types))
                lines.append(f"{indent}List<{elem_type}> ({len(value)} items)")
            else:
                types_str = ", ".join(sorted(types))
                lines.append(f"{indent}List<mixed> ({len(value)} items, types: {types_str})")

            # Traverse first few elements as examples
            for _i, item in enumerate(value[:3]):
                self._traverse(item, lines, depth + 1)
            if len(value) > 3:
                lines.append(f"{indent}  ... and {len(value) - 3} more items")

            return "array"

        if isinstance(value, bool):
            lines.append(f"{indent}{json.dumps(value)} — boolean")
            return "xsd:boolean"

        if isinstance(value, int):
            lines.append(f"{indent}{value} — integer")
            return "xsd:integer"

        if isinstance(value, float):
            lines.append(f"{indent}{value} — float")
            return "xsd:float"

        if isinstance(value, str):
            text = value[:80] + ("..." if len(value) > 80 else "")
            lines.append(f'{indent}"{text}" — xsd:string')
            return "xsd:string"

        lines.append(f"{indent}{value!r} — {type(value).__name__}")
        return "unknown"
