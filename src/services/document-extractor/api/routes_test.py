"""Tests for the document-extractor API routes with mocked dependencies."""

from __future__ import annotations

import json
from unittest.mock import patch

# Patch config before importing app to avoid env-dependent failures
with patch("config.settings") as mock_settings:
    mock_settings.SERVICE_NAME = "document-extractor"
    mock_settings.SERVICE_VERSION = "0.1.0"
    mock_settings.SERVICE_PORT = 8092
    mock_settings.MAX_FILE_SIZE_MB = 20
    mock_settings.LLM_PROVIDER = "openai"
    mock_settings.LLM_MODEL = "gpt-4o-mini"
    mock_settings.LLM_API_KEY = "test-key"
    mock_settings.LLM_MAX_RETRIES = 1
    mock_settings.LLM_TIMEOUT_SECONDS = 30
    mock_settings.ONTOLOGY_SERVICE_URL = "localhost:9001"

    from fastapi.testclient import TestClient

    from main import app


client = TestClient(app)


class TestHealthEndpoints:
    """Test health/ready/root endpoints."""

    def test_root(self) -> None:
        """GET / returns service info."""
        response = client.get("/")
        assert response.status_code == 200
        data = response.json()
        assert data["name"] == "document-extractor"
        assert data["stub"] is False

    def test_health(self) -> None:
        """GET /health returns healthy."""
        response = client.get("/health")
        assert response.status_code == 200
        assert response.json()["status"] == "healthy"

    def test_ready(self) -> None:
        """GET /ready returns ready."""
        response = client.get("/ready")
        assert response.status_code == 200
        assert response.json()["status"] == "ready"

    def test_metrics(self) -> None:
        """GET /metrics returns Prometheus-format text."""
        response = client.get("/metrics")
        assert response.status_code == 200
        assert "vedo_service_requests_total" in response.text


class TestDocumentExtract:
    """Test POST /api/v1/documents/extract."""

    def test_no_file(self) -> None:
        """POST without file returns 422."""
        response = client.post("/api/v1/documents/extract")
        assert response.status_code == 422

    def test_unsupported_format(self) -> None:
        """Uploading unsupported format returns 400."""
        response = client.post(
            "/api/v1/documents/extract",
            files={"file": ("test.exe", b"fake content", "application/octet-stream")},
        )
        assert response.status_code == 400
        assert "Unsupported" in response.json()["detail"]

    def test_empty_file(self) -> None:
        """Uploading empty file returns 400."""
        response = client.post(
            "/api/v1/documents/extract",
            files={"file": ("empty.txt", b"", "text/plain")},
        )
        assert response.status_code == 400
        assert "Empty" in response.json()["detail"]


class TestDocumentValidate:
    """Test POST /api/v1/documents/validate."""

    def test_validate_valid_sequence(self) -> None:
        """POST valid JSON sequence to validate endpoint."""
        response = client.post(
            "/api/v1/documents/validate",
            data={
                "sequence_json": json.dumps(
                    [{"operation": "create_class", "entity_id": "test", "label": "Test"}]
                )
            },
        )
        assert response.status_code == 200
        data = response.json()
        assert data["step_count"] == 1

    def test_validate_invalid_sequence(self) -> None:
        """POST invalid JSON to validate endpoint."""
        response = client.post(
            "/api/v1/documents/validate",
            data={"sequence_json": "not json"},
        )
        assert response.status_code == 400

    def test_validate_empty_steps(self) -> None:
        """POST empty steps array."""
        response = client.post(
            "/api/v1/documents/validate",
            data={"sequence_json": json.dumps([])},
        )
        assert response.status_code == 200
        data = response.json()
        assert data["error"] is not None
