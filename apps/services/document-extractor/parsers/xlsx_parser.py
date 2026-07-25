"""Parser for XLSX (Excel) documents using openpyxl.

Detects the "Class/Parent/Property/Type" pattern for ontology mapping,
otherwise treats data as generic tables. Supports sheet selection.
"""

from __future__ import annotations

import logging
import os

from parsers.base import BaseParser, ParseError
from parsers.models import ParsedDocument, Section, Table, TableCell, TableRow

logger = logging.getLogger("document-extractor.parsers.xlsx")

# Known ontology column header patterns (case-insensitive)
ONTOLOGY_COLUMN_PATTERNS = {
    "class": ["class", "class name", "class_label", "class_name", "concept", "entity"],
    "parent": ["parent", "parent class", "parent_class", "superclass", "super_class", "subclassof"],
    "property": ["property", "property name", "property_name", "attribute", "field", "column"],
    "type": ["type", "property type", "property_type", "datatype", "range", "value type"],
    "description": ["description", "comment", "definition", "notes", "note"],
    "label": ["label", "rdfs:label", "display name", "display_name", "name"],
}


class XlsxParser(BaseParser):
    """Parser for XLSX (.xlsx) files — ontology structure inference."""

    SUPPORTED_EXTENSIONS = [".xlsx", ".xlsm"]

    async def parse(self, file_path: str) -> ParsedDocument:
        """Parse an XLSX file and infer ontology structure."""
        self.validate_extension(file_path)
        file_size = os.path.getsize(file_path)

        logger.debug(
            "Parsing XLSX: %s (size=%d bytes)",
            os.path.basename(file_path),
            file_size,
        )

        try:
            from openpyxl import load_workbook
        except ImportError as exc:
            raise ParseError(
                "openpyxl not installed; run 'pip install openpyxl'",
                original=exc,
            )

        try:
            wb = load_workbook(file_path, read_only=True, data_only=True)
        except Exception as exc:
            raise ParseError(f"Failed to open XLSX {file_path}: {exc}", original=exc)

        sheet_names = wb.sheetnames
        sections: list[Section] = []
        total_rows = 0
        ontology_detected = False

        for sheet_name in sheet_names:
            ws = wb[sheet_name]
            sheet_rows: list[list[str]] = []

            for row in ws.iter_row():
                row_data = [str(cell.value or "") for cell in row]
                if any(c.strip() for c in row_data):
                    sheet_rows.append(row_data)

            if not sheet_rows:
                logger.info("Empty sheet '%s' in %s", sheet_name, file_path)
                continue

            total_rows += len(sheet_rows)
            headers = sheet_rows[0]
            data_rows = sheet_rows[1:] if len(sheet_rows) > 1 else []

            # Detect ontology column mapping
            col_mapping = self._detect_ontology_columns(headers)

            section_paragraphs: list[str] = []
            section_tables: list[Table] = []

            if col_mapping:
                ontology_detected = True
                section_paragraphs.append(f"Sheet: {sheet_name} — Ontology mapping detected")
                section_paragraphs.append(f"Rows: {len(data_rows)}")
                section_paragraphs.append("Column mapping:")

                for role, col_idx in sorted(col_mapping.items(), key=lambda x: x[1]):
                    section_paragraphs.append(f"  {role}: column {col_idx} ({headers[col_idx]})")

                # Build ontology table
                table = Table(caption=f"Ontology mapping — {sheet_name}")
                table.rows.append(
                    TableRow(
                        cells=[TableCell(text=h) for h in headers],
                        is_header=True,
                    )
                )
                for row in data_rows[:200]:
                    table.rows.append(TableRow(cells=[TableCell(text=c) for c in row]))
                if len(data_rows) > 200:
                    table.rows.append(
                        TableRow(cells=[TableCell(text=f"... {len(data_rows) - 200} more rows")])
                    )
                section_tables.append(table)
            else:
                # Generic table
                section_paragraphs.append(
                    f"Sheet: {sheet_name} — Generic table ({len(data_rows)} rows, "
                    f"{len(headers)} columns)"
                )
                table = Table(caption=f"Data — {sheet_name}")
                table.rows.append(
                    TableRow(cells=[TableCell(text=h) for h in headers], is_header=True)
                )
                for row in data_rows[:100]:
                    table.rows.append(TableRow(cells=[TableCell(text=c) for c in row]))
                if len(data_rows) > 100:
                    table.rows.append(
                        TableRow(cells=[TableCell(text=f"... {len(data_rows) - 100} more rows")])
                    )
                section_tables.append(table)

            section_paragraphs.append("")
            sections.append(
                Section(
                    heading=f"Sheet: {sheet_name}",
                    level=1,
                    paragraphs=section_paragraphs,
                    tables=section_tables,
                )
            )

        wb.close()

        if not sections:
            return ParsedDocument(
                title=os.path.splitext(os.path.basename(file_path))[0],
                format="xlsx",
                sections=[],
                metadata={"file_size_bytes": file_size, "empty": True},
            )

        filename = os.path.basename(file_path)
        logger.debug(
            "Parsed XLSX %s: %d sheets, %d rows, ontology=%s",
            filename,
            len(sections),
            total_rows,
            ontology_detected,
        )

        return ParsedDocument(
            title=os.path.splitext(filename)[0].replace("_", " ").replace("-", " ").title(),
            format="xlsx",
            sections=sections,
            metadata={
                "file_size_bytes": file_size,
                "sheets": sheet_names,
                "total_rows": total_rows,
                "ontology_mapping_detected": ontology_detected,
            },
        )

    def _detect_ontology_columns(self, headers: list[str]) -> dict[str, int]:
        """Detect ontology column mapping from header names.

        Returns a dict mapping role → column index, e.g.,
        ``{"class": 0, "parent": 1, "property": 2, "type": 3}``
        """
        mapping: dict[str, int] = {}

        for col_idx, header in enumerate(headers):
            header_lower = header.strip().lower()

            for role, patterns in ONTOLOGY_COLUMN_PATTERNS.items():
                if role in mapping:
                    continue
                if header_lower in patterns:
                    mapping[role] = col_idx
                    break

        # Only return mapping if at least 'class' is found
        if "class" in mapping:
            return mapping
        return {}
