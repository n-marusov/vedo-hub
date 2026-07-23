"""Tests for metrics-service main — FastAPI endpoints and event handling.

Validates: REQ-FUN.INTEGRATION.collaboration-quality-metrics
"""

from __future__ import annotations

from unittest.mock import AsyncMock

import pytest
from fastapi.testclient import TestClient

import main as _main_mod
from analytics import OntologyMetrics
from main import app, computer, handle_event


@pytest.fixture
def client() -> TestClient:
    """FastAPI TestClient for endpoint tests."""
    return TestClient(app)


class TestEndpoints:
    """FastAPI endpoint tests for metrics-service."""

    def test_root_returns_service_info(self, client: TestClient) -> None:
        resp = client.get("/")
        assert resp.status_code == 200
        data = resp.json()
        assert data["name"] == "metrics-service"
        assert data["version"] == "0.3.0"
        assert data["stub"] is False

    def test_health_without_redis(self, client: TestClient) -> None:
        """health returns 'healthy' with Redis 'disconnected' when no Redis configured."""
        resp = client.get("/health")
        assert resp.status_code == 200
        data = resp.json()
        assert data["status"] == "healthy"
        assert data["service"] == "metrics-service"
        assert "redis" in data

    def test_ready_returns_ready(self, client: TestClient) -> None:
        resp = client.get("/ready")
        assert resp.status_code == 200
        data = resp.json()
        assert data["status"] == "ready"
        assert data["service"] == "metrics-service"

    def test_metrics_endpoint_returns_prometheus_text(self, client: TestClient) -> None:
        resp = client.get("/metrics")
        assert resp.status_code == 200
        assert "text/plain" in resp.headers["content-type"]
        assert "# HELP" in resp.text

    def test_ingest_event_without_collector(self, client: TestClient) -> None:
        """POST /api/v1/events/ingest with collector=None returns error."""
        resp = client.post("/api/v1/events/ingest", json={"type": "commit"})
        assert resp.status_code == 200
        data = resp.json()
        assert "error" in data  # collector is None in test mode

    def test_ingest_event_invalid_json(self, client: TestClient) -> None:
        """Invalid JSON body returns INVALID_EVENT error."""
        resp = client.post(
            "/api/v1/events/ingest",
            content=b"not json",
            headers={"Content-Type": "application/json"},
        )
        assert resp.status_code == 200
        data = resp.json()
        assert "INVALID_EVENT" in str(data)

    def test_list_metrics_without_computer(self, client: TestClient) -> None:
        """list_metrics returns empty list when computer is None."""
        resp = client.get("/api/v1/metrics/ontologies")
        assert resp.status_code == 200
        data = resp.json()
        assert data["metrics"] == []
        assert data["total"] == 0

    def test_get_metrics_without_computer(self, client: TestClient) -> None:
        """get_ontology_metrics returns error when computer is None."""
        resp = client.get("/api/v1/metrics/ontologies/onto-nonexistent")
        assert resp.status_code == 200
        data = resp.json()
        assert "METRICS_UNAVAILABLE" in str(data)


class TestHandleEvent:
    """handle_event function tests (bypasses FastAPI, calls directly)."""

    async def test_handle_event_without_computer(self) -> None:
        """handle_event returns early when computer is None."""
        # Temporarily unset computer
        saved = computer
        try:
            _main_mod.computer = None
            await handle_event({"type": "commit"})  # should not raise
        finally:
            _main_mod.computer = saved

    async def test_handle_event_with_mock_computer(self) -> None:
        """handle_event processes an event through the mock computer."""
        saved = computer
        try:
            mock = AsyncMock()
            mock.compute_from_event = AsyncMock(
                return_value=OntologyMetrics(
                    ontology_id="mock-onto",
                    axiom_count=42,
                    class_count=10,
                    property_count=5,
                    individual_count=20,
                    max_class_depth=2,
                    property_density=0.5,
                )
            )
            mock.get_stored_count = AsyncMock(return_value=1)
            _main_mod.computer = mock

            await handle_event({"type": "commit", "ontology_id": "mock-onto"})
            mock.compute_from_event.assert_awaited_once()
            mock.get_stored_count.assert_awaited_once()
        finally:
            _main_mod.computer = saved
