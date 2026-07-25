"""Parser for plain text and Markdown documents (.md, .txt).

Extracts headings (H1-H6), paragraphs, bullet lists, numbered lists,
code blocks, and tables from Markdown. Plain text is treated as a
single-section document with paragraphs separated by blank lines.
"""

from __future__ import annotations

import logging
import os
import re

from parsers.base import BaseParser, ParseError
from parsers.models import (
    ListItem,
    ParsedDocument,
    Section,
    Table,
    TableCell,
    TableRow,
)

logger = logging.getLogger("document-extractor.parsers.text")

# Regex patterns for Markdown structure detection
HEADING_RE = re.compile(r"^(#{1,6})\s+(.+)$", re.MULTILINE)
BULLET_LIST_RE = re.compile(r"^(\s*)[-*+]\s+(.+)$", re.MULTILINE)
NUMBERED_LIST_RE = re.compile(r"^(\s*)\d+[.)]\s+(.+)$", re.MULTILINE)
CODE_BLOCK_RE = re.compile(r"```(\w*)\n(.*?)```", re.DOTALL)
TABLE_ROW_RE = re.compile(r"^\|(.+)\|$", re.MULTILINE)
TABLE_SEP_RE = re.compile(r"^\|[-| :]+\|$", re.MULTILINE)


class TextParser(BaseParser):
    """Parser for Markdown (.md) and plain text (.txt) files."""

    SUPPORTED_EXTENSIONS = [".md", ".txt", ".markdown"]

    async def parse(self, file_path: str) -> ParsedDocument:
        """Parse a Markdown or plain text file."""
        self.validate_extension(file_path)
        file_size = os.path.getsize(file_path)

        logger.debug(
            "Parsing text file: %s (size=%d bytes)",
            os.path.basename(file_path),
            file_size,
        )

        try:
            with open(file_path, encoding="utf-8") as f:
                content = f.read()
        except UnicodeDecodeError:
            # Fallback to latin-1 for binary-ish text
            with open(file_path, encoding="latin-1") as f:
                content = f.read()
        except Exception as exc:
            raise ParseError(f"Failed to read text file: {exc}", original=exc)

        if not content.strip():
            logger.info("Empty text file: %s", file_path)
            return ParsedDocument(
                title=os.path.basename(file_path),
                format="text",
                metadata={"file_size_bytes": file_size, "empty": True},
            )

        if self._is_markdown(content):
            return self._parse_markdown(file_path, content, file_size)
        return self._parse_plain_text(file_path, content, file_size)

    def _is_markdown(self, content: str) -> bool:
        """Heuristic: detect Markdown vs plain text."""
        # Check for common Markdown patterns
        if HEADING_RE.search(content):
            return True
        if re.search(r"\[.+\]\(.+\)", content):  # links
            return True
        if re.search(r"\*\*.+?\*\*", content):  # bold
            return True
        if re.search(r"(?<!\*)\*(?!\*)(.+?)(?<!\*)\*(?!\*)", content):  # italic
            return True
        return False

    def _parse_markdown(self, file_path: str, content: str, file_size: int) -> ParsedDocument:
        """Parse Markdown content with full structure detection."""
        filename = os.path.basename(file_path)

        # Extract code blocks first (remove them to avoid false matches)
        code_blocks: list[str] = []

        def _extract_code(m: re.Match) -> str:
            code_blocks.append(m.group(2))
            return f"\n```codeblock-{len(code_blocks) - 1}```\n"

        content_no_code = CODE_BLOCK_RE.sub(_extract_code, content)

        # Split into sections by headings
        sections: list[Section] = []
        current_section = Section(heading="__root__", level=0)
        heading_matches = list(HEADING_RE.finditer(content_no_code))

        if not heading_matches:
            # No headings found — treat entire document as one section
            current_section.paragraphs = [
                p.strip() for p in content_no_code.split("\n\n") if p.strip()
            ]
            sections.append(current_section)
        else:
            prev_end = 0
            for i, match in enumerate(heading_matches):
                level = len(match.group(1))
                heading_text = match.group(2).strip()

                # Collect text between previous heading and this one
                if i > 0:
                    body = content_no_code[prev_end : match.start()].strip()
                    self._parse_body_into_section(current_section, body)

                if current_section.heading != "__root__":
                    sections.append(current_section)

                current_section = Section(heading=heading_text, level=level)
                prev_end = match.end()

            # Last section body
            body = content_no_code[prev_end:].strip()
            if body:
                self._parse_body_into_section(current_section, body)
            if current_section.heading != "__root__":
                sections.append(current_section)

        # Attach code blocks to nearest sections
        cb_idx = 0
        for section in sections:
            while cb_idx < len(code_blocks) and not self._section_has_code(
                content_no_code, section.heading
            ):
                pass
            # Simple: distribute code blocks across sections
        # Distribute code blocks evenly
        if code_blocks and sections:
            per_section = max(1, len(code_blocks) // len(sections))
            for i, section in enumerate(sections):
                start = i * per_section
                end = start + per_section
                section.code_blocks = code_blocks[start:end]

        title = self._extract_title(filename, sections)
        section_count = len(sections)

        logger.debug(
            "Parsed %s: %d sections, %d code blocks",
            filename,
            section_count,
            len(code_blocks),
        )

        return ParsedDocument(
            title=title,
            format="markdown",
            sections=sections,
            metadata={
                "file_size_bytes": file_size,
                "section_count": section_count,
                "code_block_count": len(code_blocks),
            },
        )

    def _extract_title(self, filename: str, sections: list[Section]) -> str:
        """Extract the document title from sections or filename."""
        if sections:
            # First H1 heading is the title
            for s in sections:
                if s.level == 1 and s.heading != "__root__":
                    return s.heading
            # First section heading (any level)
            name = sections[0].heading
            if name and name != "__root__":
                return name
        # Fallback to filename without extension
        name = os.path.splitext(filename)[0]
        return name.replace("_", " ").replace("-", " ").title()

    def _parse_plain_text(self, file_path: str, content: str, file_size: int) -> ParsedDocument:
        """Parse plain text as a single-section document."""
        paragraphs = [p.strip() for p in content.split("\n\n") if p.strip()]
        filename = os.path.basename(file_path)

        section = Section(
            heading="",
            level=0,
            paragraphs=paragraphs,
        )

        logger.debug(
            "Parsed plain text %s: %d paragraphs",
            filename,
            len(paragraphs),
        )

        return ParsedDocument(
            title=os.path.splitext(filename)[0].replace("_", " ").replace("-", " ").title(),
            format="text",
            sections=[section],
            metadata={"file_size_bytes": file_size, "paragraph_count": len(paragraphs)},
        )

    def _parse_body_into_section(self, section: Section, body: str) -> None:
        """Parse the body text between headings into paragraphs, lists, tables."""
        if not body:
            return

        lines = body.split("\n")
        i = 0
        while i < len(lines):
            line = lines[i]

            # Check for table row
            if TABLE_ROW_RE.match(line):
                table, consumed = self._parse_table(lines, i)
                if table:
                    section.tables.append(table)
                    i += consumed
                    continue

            # Check for list item
            bullet_match = BULLET_LIST_RE.match(line)
            numbered_match = NUMBERED_LIST_RE.match(line)

            if bullet_match or numbered_match:
                items, consumed = self._parse_list(lines, i)
                section.lists.extend(items)
                i += consumed
                continue

            # Regular paragraph (non-empty, non-separator)
            stripped = line.strip()
            if stripped and not TABLE_SEP_RE.match(line):
                section.paragraphs.append(stripped)

            i += 1

    def _parse_table(self, lines: list[str], start: int) -> tuple[Table | None, int]:
        """Parse a Markdown table starting at line index start."""
        table = Table()
        consumed = 0
        header_done = False

        for idx in range(start, len(lines)):
            line = lines[idx]
            if not TABLE_ROW_RE.match(line):
                break

            cells_str = line.strip().strip("|")
            cells = [c.strip() for c in cells_str.split("|")]

            # Skip separator row
            if TABLE_SEP_RE.match(line):
                header_done = True
                consumed += 1
                continue

            row = TableRow(
                cells=[TableCell(text=c) for c in cells],
                is_header=not header_done and idx == start,
            )
            table.rows.append(row)
            consumed += 1

        return table if table.rows else None, consumed

    def _parse_list(self, lines: list[str], start: int) -> tuple[list[ListItem], int]:
        """Parse a continuous list starting at line index start."""
        items: list[ListItem] = []
        consumed = 0

        for idx in range(start, len(lines)):
            line = lines[idx]

            bullet_match = BULLET_LIST_RE.match(line)
            numbered_match = NUMBERED_LIST_RE.match(line)

            if not (bullet_match or numbered_match):
                if line.strip() == "":
                    consumed += 1
                    break
                # Continuation of last item
                if items:
                    items[-1].text += " " + line.strip()
                    consumed += 1
                    continue
                break

            if bullet_match:
                indent = len(bullet_match.group(1))
                text = bullet_match.group(2).strip()
                items.append(ListItem(text=text, level=indent // 2, ordered=False))
            else:
                indent = len(numbered_match.group(1))
                text = numbered_match.group(2).strip()
                items.append(
                    ListItem(
                        text=text,
                        level=indent // 2,
                        ordered=True,
                        number=idx - start + 1,
                    )
                )
            consumed += 1

        return items, consumed

    def _section_has_code(self, content: str, heading: str) -> bool:
        """Check if a code block is near a section heading."""
        if not heading:
            return False
        pattern = rf"#+\s+{re.escape(heading)}.*?\n```"
        return bool(re.search(pattern, content, re.DOTALL))
