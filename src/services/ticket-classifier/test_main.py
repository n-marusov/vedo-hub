"""Tests for ticket-classifier — HTTP handler and utility functions."""

from __future__ import annotations

import json
import threading
from http.server import HTTPServer
from typing import TYPE_CHECKING
from urllib.error import HTTPError
from urllib.request import Request, urlopen

import pytest

from main import SERVICE_NAME, Handler, resolve_id

if TYPE_CHECKING:
    from collections.abc import Generator


class TestResolveId:
    """resolve_id utility function tests."""

    def test_returns_value_when_provided(self) -> None:
        assert resolve_id("abc-123") == "abc-123"

    def test_returns_str_when_none(self) -> None:
        result = resolve_id(None)
        assert isinstance(result, str)
        assert len(result) > 0

    def test_returns_str_when_empty(self) -> None:
        result = resolve_id("")
        assert isinstance(result, str)
        assert len(result) > 0

    def test_different_calls_different_values(self) -> None:
        r1 = resolve_id(None)
        r2 = resolve_id(None)
        assert r1 != r2


@pytest.fixture(scope="module")
def server_url() -> Generator[str, None, None]:
    """Start the ticket-classifier server on a random port and return its URL."""
    server = HTTPServer(("127.0.0.1", 0), Handler)
    port = server.server_address[1]
    url = f"http://127.0.0.1:{port}"

    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()

    yield url

    server.shutdown()
    thread.join(timeout=2)


class TestHandlerHTTP:
    """Handler HTTP integration tests — against a real server."""

    def test_get_root(self, server_url: str) -> None:
        """GET / returns service info."""
        resp = urlopen(f"{server_url}/")
        assert resp.status == 200
        data = json.loads(resp.read().decode("utf-8"))
        assert data["name"] == SERVICE_NAME
        assert data["version"] == "0.2.0"
        assert data["stub"] is False

    def test_get_health(self, server_url: str) -> None:
        """GET /health returns healthy status."""
        resp = urlopen(f"{server_url}/health")
        assert resp.status == 200
        data = json.loads(resp.read().decode("utf-8"))
        assert data["status"] == "healthy"

    def test_get_ready(self, server_url: str) -> None:
        """GET /ready returns ready status."""
        resp = urlopen(f"{server_url}/ready")
        assert resp.status == 200
        data = json.loads(resp.read().decode("utf-8"))
        assert data["status"] == "ready"

    def test_get_metrics(self, server_url: str) -> None:
        """GET /metrics returns Prometheus-format text."""
        resp = urlopen(f"{server_url}/metrics")
        assert resp.status == 200
        body = resp.read().decode("utf-8")
        assert SERVICE_NAME in body
        assert "vedo_service_requests_total" in body

    def test_get_unknown_endpoint(self, server_url: str) -> None:
        """GET /unknown returns 404 with available endpoints."""
        req = Request(f"{server_url}/unknown")
        with pytest.raises(HTTPError) as excinfo:
            urlopen(req)
        assert excinfo.value.code == 404
        data = json.loads(excinfo.value.read().decode("utf-8"))
        assert data["error"] == "ENDPOINT_NOT_FOUND"
        assert "available" in data

    def test_post_root_returns_405(self, server_url: str) -> None:
        """POST / returns 405 Method Not Allowed."""
        req = Request(f"{server_url}/", data=b"{}", method="POST")
        with pytest.raises(HTTPError) as excinfo:
            urlopen(req)
        assert excinfo.value.code == 405
        data = json.loads(excinfo.value.read().decode("utf-8"))
        assert data["error"] == "METHOD_NOT_ALLOWED"
        assert "available" in data

    def test_put_health_returns_405(self, server_url: str) -> None:
        """PUT /health returns 405 Method Not Allowed."""
        req = Request(f"{server_url}/health", data=b"{}", method="PUT")
        with pytest.raises(HTTPError) as excinfo:
            urlopen(req)
        assert excinfo.value.code == 405

    def test_delete_unknown_returns_404_not_405(self, server_url: str) -> None:
        """DELETE /nonexistent returns 404 (path not in allowed list)."""
        req = Request(f"{server_url}/nonexistent", method="DELETE")
        with pytest.raises(HTTPError) as excinfo:
            urlopen(req)
        assert excinfo.value.code == 404

    def test_request_total_increments(self, server_url: str) -> None:
        """Each request increments the request counter reflected in /metrics."""
        counter_before = Handler.request_total
        urlopen(f"{server_url}/")
        urlopen(f"{server_url}/health")
        assert Handler.request_total == counter_before + 2

    def test_trace_and_correlation_headers_logged(self, server_url: str) -> None:
        """Handler accepts trace and correlation headers (should not crash)."""
        req = Request(
            f"{server_url}/",
            headers={
                "X-Trace-Id": "trace-001",
                "X-Correlation-Id": "corr-001",
            },
        )
        resp = urlopen(req)
        assert resp.status == 200
