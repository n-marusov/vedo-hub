"""Security placeholder tests for document-extractor P0 requirements.

Validates: REQ-NFR.SECURITY.doc-extract-security
Validates: REQ-NFR.SECURITY.excel-import-security

Remove @pytest.mark.skip and implement real assertions when the
corresponding security features are wired.
"""

from __future__ import annotations

import pytest


class TestDocExtractSecurity:
    """Document extraction security (REQ-NFR.SECURITY.doc-extract-security)."""

    @pytest.mark.skip(reason="REQ-NFR.SECURITY.doc-extract-security: requires document upload sanitization middleware")
    async def test_doc_extract_rejects_malicious_content(self) -> None:
        """Uploading a document with embedded scripts must be sanitized."""
        raise NotImplementedError

    @pytest.mark.skip(reason="REQ-NFR.SECURITY.doc-extract-security: requires file type validation")
    async def test_doc_extract_rejects_unsupported_formats(self) -> None:
        """Uploading an unsupported file type must return a clear error."""
        raise NotImplementedError


class TestExcelImportSecurity:
    """Excel import security (REQ-NFR.SECURITY.excel-import-security)."""

    @pytest.mark.skip(reason="REQ-NFR.SECURITY.excel-import-security: requires Excel macro/sanitization checks")
    async def test_excel_import_rejects_macros(self) -> None:
        """XLSX files with macros must be rejected during import."""
        raise NotImplementedError

    @pytest.mark.skip(reason="REQ-NFR.SECURITY.excel-import-security: requires formula injection protection")
    async def test_excel_import_sanitizes_formulas(self) -> None:
        """Cells starting with =, +, -, @ must be escaped or rejected."""
        raise NotImplementedError
