"""Parser for XML documents.

Parses XML tags as classes, attributes as properties, and nesting as hierarchy.
Preserves namespace information for ontology mapping.
"""

from __future__ import annotations

import logging
import os
import re
import xml.etree.ElementTree as ET

from parsers.base import BaseParser, ParseError
from parsers.models import ParsedDocument, Section

logger = logging.getLogger("document-extractor.parsers.xml")


class XmlParser(BaseParser):
    """Parser for XML (.xml) files — ontology structure inference."""

    SUPPORTED_EXTENSIONS = [".xml"]

    async def parse(self, file_path: str) -> ParsedDocument:
        """Parse an XML file and infer ontology structure."""
        self.validate_extension(file_path)
        file_size = os.path.getsize(file_path)

        logger.debug(
            "Parsing XML: %s (size=%d bytes)",
            os.path.basename(file_path),
            file_size,
        )

        try:
            tree = ET.parse(file_path)
            root = tree.getroot()
        except ET.ParseError as exc:
            raise ParseError(f"Invalid XML in {file_path}: {exc}", original=exc)
        except Exception as exc:
            raise ParseError(f"Failed to read XML file {file_path}: {exc}", original=exc)

        # Extract structure from XML
        structure_lines: list[str] = []
        namespaces: dict[str, str] = {}

        # Collect namespace declarations from raw XML content
        # ElementTree does not expose xmlns attributes in elem.attrib,
        # so we scan the raw text for xmlns:prefix="uri" patterns.
        try:
            with open(file_path, encoding="utf-8") as f:
                raw_content = f.read()
            for match in re.finditer(
                r'\sxmlns(?::(\w+))?\s*=\s*["\']([^"\']+)["\']',
                raw_content,
            ):
                prefix = match.group(1) or "__default__"
                namespaces[prefix] = match.group(2)
        except Exception:
            pass

        # Add namespace info
        if namespaces:
            structure_lines.append(f"Namespaces ({len(namespaces)}):")
            for prefix, uri in namespaces.items():
                structure_lines.append(f"  {prefix}: {uri}")
            structure_lines.append("")

        self._traverse_element(root, structure_lines, depth=0)

        filename = os.path.basename(file_path)
        section = Section(
            heading=f"XML Structure: {filename}",
            level=1,
            paragraphs=structure_lines,
        )

        logger.debug(
            "Parsed XML %s: inferred %d structural elements",
            filename,
            len(structure_lines),
        )

        return ParsedDocument(
            title=os.path.splitext(filename)[0].replace("_", " ").replace("-", " ").title(),
            format="xml",
            sections=[section],
            metadata={
                "file_size_bytes": file_size,
                "root_tag": root.tag,
                "namespace_count": len(namespaces),
                "element_count": len(structure_lines),
            },
        )

    def _traverse_element(self, elem: ET.Element, lines: list[str], depth: int = 0) -> None:
        """Recursively traverse an XML element and infer structure."""
        indent = "  " * depth
        tag = self._clean_tag(elem.tag)

        # Determine if this is a leaf or container
        children = list(elem)
        if children:
            lines.append(f"{indent}Class: {tag}")
            if depth > 0:
                lines.append(f"{indent}  (children of parent class)")

            # Attributes as datatype properties
            for attr_name, attr_value in elem.attrib.items():
                if attr_name.startswith("{"):
                    # Prefixed attribute
                    clean_name = self._clean_tag(attr_name)
                    lines.append(f'{indent}  Property (datatype): {clean_name} = "{attr_value}"')
                elif attr_name == "xmlns":
                    continue
                else:
                    lines.append(f'{indent}  Property (datatype): {attr_name} = "{attr_value}"')

            # Text content as property
            if elem.text and elem.text.strip():
                text_preview = elem.text.strip()[:60]
                lines.append(f'{indent}  Property (datatype): value = "{text_preview}"')

            # Children as object properties
            child_tags: dict[str, int] = {}
            for child in children:
                child_tag = self._clean_tag(child.tag)
                child_tags[child_tag] = child_tags.get(child_tag, 0) + 1

            for child_tag, count in child_tags.items():
                lines.append(
                    f"{indent}  Property (object): has_{child_tag} → {child_tag}"
                    f"{' (x' + str(count) + ')' if count > 1 else ''}"
                )

            # Recurse into children (limit depth)
            if depth < 5:
                for child in children[:5]:
                    self._traverse_element(child, lines, depth + 1)
                if len(children) > 5:
                    lines.append(f"{indent}  ... and {len(children) - 5} more children")

        else:
            # Leaf element
            text_val = elem.text.strip() if elem.text and elem.text.strip() else ""
            text_preview = text_val[:60] + ("..." if len(text_val) > 60 else "")

            if elem.attrib:
                lines.append(f"{indent}Class: {tag}")
                for attr_name, attr_value in elem.attrib.items():
                    if not attr_name.startswith("xmlns"):
                        lines.append(f'{indent}  Property (datatype): {attr_name} = "{attr_value}"')
                if text_preview:
                    lines.append(f'{indent}  Property (datatype): value = "{text_preview}"')
            else:
                lines.append(f'{indent}Property (datatype): {tag} = "{text_preview}"')

    @staticmethod
    def _clean_tag(tag: str) -> str:
        """Remove namespace URI from a tag name and simplify."""
        if tag.startswith("{"):
            idx = tag.find("}")
            return tag[idx + 1 :] if idx > 0 else tag
        return tag
