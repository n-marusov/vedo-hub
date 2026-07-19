"""Parser for CSV documents.

Detects delimiter (`,`/`;`/`\t`), extracts headers as column names,
and prepares data for LLM-based column→ontology mapping.
"""

from __future__ import annotations

import csv
import io
import logging
import os
from collections import Counter

from parsers.base import BaseParser, ParseError
from parsers.models import ParsedDocument, Section, Table, TableCell, TableRow

logger = logging.getLogger("document-extractor.parsers.csv")

# Delimiters to try, in priority order
DELIMITERS = [",", ";", "\t", "|"]


class CsvParser(BaseParser):
    """Parser for CSV (.csv) files — ontology structure inference."""

    SUPPORTED_EXTENSIONS = [".csv", ".tsv"]

    async def parse(self, file_path: str) -> ParsedDocument:
        """Parse a CSV file and infer ontology structure."""
        self.validate_extension(file_path)
        file_size = os.path.getsize(file_path)

        logger.debug(
            "Parsing CSV: %s (size=%d bytes)",
            os.path.basename(file_path),
            file_size,
        )

        try:
            with open(file_path, encoding="utf-8", newline="") as f:
                raw = f.read()
        except UnicodeDecodeError:
            with open(file_path, encoding="latin-1", newline="") as f:
                raw = f.read()
        except Exception as exc:
            raise ParseError(f"Failed to read CSV file {file_path}: {exc}", original=exc)

        if not raw.strip():
            return ParsedDocument(
                title=os.path.splitext(os.path.basename(file_path))[0],
                format="csv",
                sections=[],
                metadata={"file_size_bytes": file_size, "empty": True},
            )

        # Auto-detect delimiter
        delimiter = self._detect_delimiter(raw)

        lines = raw.split("\n")
        total_rows = len([l for l in lines if l.strip()])

        # Parse with detected delimiter
        reader = csv.reader(io.StringIO(raw), delimiter=delimiter)
        rows = [r for r in reader if any(c.strip() for c in r)]

        if not rows:
            return ParsedDocument(
                title=os.path.splitext(os.path.basename(file_path))[0],
                format="csv",
                sections=[],
                metadata={"file_size_bytes": file_size, "empty": True},
            )

        # First row is header
        headers = rows[0]
        data_rows = rows[1:] if len(rows) > 1 else []

        # Build analysis
        filename = os.path.basename(file_path)
        analysis_lines = [
            f"CSV File: {filename}",
            f"Delimiter: {delimiter!r}",
            f"Columns: {len(headers)}",
            f"Data rows: {len(data_rows)}",
            "",
            "Column Analysis:",
        ]

        for col_idx, header in enumerate(headers):
            col_values = [
                row[col_idx] for row in data_rows if col_idx < len(row) and row[col_idx].strip()
            ]
            unique_count = len(set(col_values))
            sample_values = col_values[:3] if col_values else []

            analysis_lines.append(
                f'  Column {col_idx}: "{header}" — '
                f"{len(col_values)} non-empty values, "
                f"{unique_count} unique"
            )
            if sample_values:
                samples = ", ".join(repr(v) for v in sample_values)
                analysis_lines.append(f"    Samples: {samples}")

        # Build table representation
        table = Table(caption=f"CSV: {filename}")
        table.rows.append(TableRow(cells=[TableCell(text=h) for h in headers], is_header=True))
        for row in data_rows[:100]:  # Limit to first 100 rows
            table.rows.append(TableRow(cells=[TableCell(text=c) for c in row]))
        if len(data_rows) > 100:
            table.rows.append(
                TableRow(cells=[TableCell(text=f"... {len(data_rows) - 100} more rows")])
            )

        section = Section(
            heading=f"CSV Data: {filename}",
            level=1,
            paragraphs=analysis_lines,
            tables=[table],
        )

        logger.debug(
            "Parsed CSV %s: delimiter=%r columns=%d rows=%d",
            filename,
            delimiter,
            len(headers),
            len(data_rows),
        )

        return ParsedDocument(
            title=os.path.splitext(filename)[0].replace("_", " ").replace("-", " ").title(),
            format="csv",
            sections=[section],
            metadata={
                "file_size_bytes": file_size,
                "delimiter": delimiter,
                "columns": len(headers),
                "data_rows": len(data_rows),
                "column_names": headers,
            },
        )

    def _detect_delimiter(self, raw: str) -> str:
        """Auto-detect the CSV delimiter by trying common ones.

        Tries `,` first, then `;` and `\t`. The delimiter that produces
        the most consistent row lengths (>50% of non-empty rows) wins.
        """
        lines = [l for l in raw.split("\n") if l.strip()]
        if not lines:
            return ","

        best_delimiter = ","
        best_consistency = 0.0

        for delim in DELIMITERS:
            col_counts: list[int] = []

            for line in lines:
                reader = csv.reader(io.StringIO(line), delimiter=delim)
                try:
                    row = next(reader)
                    col_counts.append(len(row))
                except StopIteration:
                    continue

            if not col_counts:
                continue

            # Check consistency — do most rows have the same column count?
            count_freq = Counter(col_counts)
            most_common_count, most_common_freq = count_freq.most_common(1)[0]
            consistency = most_common_freq / len(col_counts)

            if consistency > best_consistency and most_common_count > 1:
                best_consistency = consistency
                best_delimiter = delim

            if consistency >= 0.5 and most_common_count > 1:
                break  # Good enough — delimiter produces multiple columns

        if best_delimiter != ",":
            logger.info(
                "Detected CSV delimiter: %r (consistency=%.0f%%)",
                best_delimiter,
                best_consistency * 100,
            )

        return best_delimiter
