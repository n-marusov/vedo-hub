"""Abstract base parser for all document formats."""

from __future__ import annotations

import logging
from abc import ABC, abstractmethod

from parsers.models import ParsedDocument

logger = logging.getLogger("document-extractor.parsers.base")

# ─── Magic bytes signatures ────────────────────────────────────────────────────
# First bytes used to verify file identity, mapped by extension.

MAGIC_BYTES: dict[str, list[bytes]] = {
    ".pdf": [b"%PDF"],
    ".docx": [b"PK\x03\x04"],  # ZIP-based (OOXML)
    ".docm": [b"PK\x03\x04"],  # ZIP-based (OOXML with macros)
    ".xlsx": [b"PK\x03\x04"],  # ZIP-based (OOXML)
    ".xlsm": [b"PK\x03\x04"],  # ZIP-based (OOXML with macros)
    ".xml": [b"<?xml", b"\xff\xfe<"],  # text XML or UTF-16 BOM + <
    ".json": [b"{", b"["],  # object or array start
    ".csv": [],  # no reliable magic bytes — validated by parser
    ".tsv": [],
    ".txt": [],
    ".md": [],
    ".markdown": [],
}


def validate_file_signature(file_path: str, expected_ext: str | None = None) -> bool:
    """Check magic bytes of a file against known signatures.

    Reads the first 16 bytes and compares against expected signatures
    for the file extension. Formats without reliable magic bytes (text,
    CSV, MD) skip validation and return True.

    Args:
        file_path: Path to the file to validate.
        expected_ext: Expected extension (e.g. ``.pdf``). If None,
                      inferred from file path.

    Returns:
        True if signature matches or format has no magic bytes.

    Raises:
        ValueError: If signature doesn't match expected format.
    """
    import os

    if expected_ext is None:
        _, ext = os.path.splitext(file_path)
        expected_ext = ext.lower()

    signatures = MAGIC_BYTES.get(expected_ext, [])
    if not signatures:
        # No magic bytes defined for this format — skip check
        return True

    try:
        with open(file_path, "rb") as f:
            header = f.read(16)
    except OSError:
        raise ValueError(f"Cannot read file for magic byte check: {file_path}")

    for sig in signatures:
        if header.startswith(sig):
            return True

    # Provide a human-readable hex preview of the header for diagnostics
    hex_preview = " ".join(f"{b:02x}" for b in header[:8])
    sig_desc = " | ".join(s.hex() if isinstance(s, bytes) else str(s) for s in signatures)
    raise ValueError(
        f"File {file_path!r} has extension '{expected_ext}' but its "
        f"magic bytes ({hex_preview}...) do not match expected signatures "
        f"({sig_desc}). The file may be corrupted or misnamed."
    )


class BaseParser(ABC):
    """Base class for all document parsers.

    Each subclass handles a specific file format and returns a uniform
    ``ParsedDocument`` with sections preserving hierarchy.
    """

    SUPPORTED_EXTENSIONS: list[str] = []
    """File extensions this parser handles (e.g., ``[".md", ".txt"]``)."""

    @abstractmethod
    async def parse(self, file_path: str) -> ParsedDocument:
        """Parse a file and return a structured document.

        Args:
            file_path: Absolute or relative path to the file.

        Returns:
            ParsedDocument with extracted sections, tables, and metadata.

        Raises:
            FileNotFoundError: If the file does not exist.
            ValueError: If the file format is invalid or unsupported.
            ParseError: On unrecoverable parse failures.
        """
        ...

    def validate_extension(self, file_path: str) -> None:
        """Check that the file has a supported extension.

        Args:
            file_path: Path to validate.

        Raises:
            ValueError: If the extension is not supported.
        """
        if not self.SUPPORTED_EXTENSIONS:
            return
        ext = file_path.lower()
        if not any(ext.endswith(e) for e in self.SUPPORTED_EXTENSIONS):
            raise ValueError(
                f"Unsupported extension for {type(self).__name__}: "
                f"{file_path!r}. Supported: {self.SUPPORTED_EXTENSIONS}"
            )

    def validate_file(self, file_path: str) -> None:
        """Run all file-level validations: extension check and magic bytes.

        Args:
            file_path: Path to the file.

        Raises:
            ValueError: On any validation failure.
        """
        self.validate_extension(file_path)
        import os

        _, ext = os.path.splitext(file_path)
        ext = ext.lower()

        # Skip magic byte check for formats without defined signatures
        if ext in (".txt", ".md", ".markdown", ".csv", ".tsv"):
            logger.debug("Skipping magic byte check for %s (no signature defined)", ext)
            return

        try:
            validate_file_signature(file_path, ext)
            logger.debug("Magic bytes validated for %s", file_path)
        except ValueError as exc:
            logger.warning("Magic byte validation failed: %s", exc)
            raise


class ParseError(Exception):
    """Raised when parsing fails unrecoverably."""

    def __init__(self, message: str, original: Exception | None = None) -> None:
        super().__init__(message)
        self.original = original
