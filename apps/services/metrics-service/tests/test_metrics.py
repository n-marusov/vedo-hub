"""First test suite for metrics-service.

Covers:
1. Health/ready endpoints return the expected payloads (direct async calls —
   no TestClient/httpx dependency).
2. The Prometheus registration guard (`main._metric`) is idempotent: a second
   registration of the same metric name reuses the existing collector instead
   of raising ValueError (regression for the "Duplicated timeseries in
   CollectorRegistry" startup crash).
3. The /metrics payload is non-empty and contains our metric families.
"""

from __future__ import annotations

import sys
from pathlib import Path

# Ensure the service root (main.py, config.py) is importable when pytest runs
# from the tests/ directory.
sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

import pytest
from prometheus_client import Counter, generate_latest

import main


@pytest.mark.asyncio
async def test_health_endpoint_returns_healthy() -> None:
    """[HealthChecked]_[ReturnsHealthyStatus]_[WhenCalled]."""
    result = await main.health()
    assert result == {"status": "healthy"}


@pytest.mark.asyncio
async def test_ready_endpoint_returns_ok_http() -> None:
    """[ReadyChecked]_[IncludesHttpOk]_[Always]."""
    result = await main.ready()
    assert "checks" in result
    assert any("http:ok" in c for c in result["checks"])


@pytest.mark.asyncio
async def test_root_endpoint_reports_service_identity() -> None:
    """[RootChecked]_[ReportsServiceNameAndVersion]_[WhenCalled]."""
    result = await main.root()
    assert result["name"] == main.SERVICE_NAME
    assert result["version"] == main.SERVICE_VERSION
    assert "uptime_seconds" in result
    assert result["uptime_seconds"] >= 0


def test_metric_registration_is_idempotent() -> None:
    """[MetricRegistered]_[ReusesExistingCollector]_[OnDuplicateName]."""
    # Registering the same metric name again must NOT raise ValueError and
    # must return the existing collector.
    counter = main._metric(
        Counter,
        "vedo_requests_total",
        "Total requests",
        ["method", "path", "status"],
    )
    assert counter is main.REQUESTS_TOTAL


def test_metrics_payload_contains_metric_families() -> None:
    """[MetricsExported]_[ContainsCoreFamilies]_[Always]."""
    payload = generate_latest().decode("utf-8")
    assert "vedo_requests_total" in payload
    assert "vedo_events_consumed_total" in payload
    assert "vedo_active_connections" in payload
