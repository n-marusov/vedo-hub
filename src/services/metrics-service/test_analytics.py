"""Tests for MetricsComputer — core ontology metric computation.

Validates: REQ-FUN.INTEGRATION.collaboration-quality-metrics
"""

from __future__ import annotations

import pytest

from analytics import MetricsComputer, OntologyMetrics


@pytest.fixture
def computer() -> MetricsComputer:
    """MetricsComputer without Redis (local-cache only)."""
    return MetricsComputer(redis_url=None)


class TestMetricsComputer:
    """MetricsComputer unit tests — local cache only, no Redis."""

    async def test_connect_without_redis(self, computer: MetricsComputer) -> None:
        """No Redis URL → connect succeeds gracefully, no error."""
        await computer.connect()
        assert computer._redis is None

    async def test_compute_from_event_basic(self, computer: MetricsComputer) -> None:
        """Compute metrics from a minimal event payload."""
        event = {
            "type": "commit",
            "ontology_id": "onto-001",
            "axiom_count": 100,
            "class_count": 10,
            "property_count": 5,
            "individual_count": 20,
        }
        metrics = await computer.compute_from_event(event)
        assert isinstance(metrics, OntologyMetrics)
        assert metrics.ontology_id == "onto-001"
        assert metrics.axiom_count == 100
        assert metrics.class_count == 10
        assert metrics.property_count == 5
        assert metrics.individual_count == 20
        assert metrics.max_class_depth == 0  # no depth_scores in event

    async def test_compute_from_event_with_depth(self, computer: MetricsComputer) -> None:
        """Compute metrics with class depth scores."""
        event = {
            "type": "import",
            "ontology_id": "onto-002",
            "axiom_count": 500,
            "class_count": 25,
            "property_count": 12,
            "individual_count": 100,
            "class_depth_scores": {
                "Thing": 0,
                "Person": 1,
                "Student": 2,
                "Employee": 2,
                "Professor": 3,
            },
        }
        metrics = await computer.compute_from_event(event)
        assert metrics.max_class_depth == 3
        assert metrics.avg_class_depth == 1.6  # (0+1+2+2+3) / 5
        assert metrics.property_density == 0.48  # 12 / 25

    async def test_compute_from_event_zero_class_count(self, computer: MetricsComputer) -> None:
        """Property density must be 0.0 when class_count is 0."""
        event = {
            "type": "publish",
            "ontology_id": "onto-empty",
            "axiom_count": 0,
            "class_count": 0,
            "property_count": 5,
            "individual_count": 0,
        }
        metrics = await computer.compute_from_event(event)
        assert metrics.property_density == 0.0
        assert metrics.max_class_depth == 0
        assert metrics.avg_class_depth == 0.0

    async def test_get_metrics_from_local_cache(self, computer: MetricsComputer) -> None:
        """get_metrics returns locally cached metrics."""
        event = {
            "type": "commit",
            "ontology_id": "onto-cache",
            "axiom_count": 200,
            "class_count": 15,
            "property_count": 8,
            "individual_count": 30,
        }
        await computer.compute_from_event(event)
        retrieved = await computer.get_metrics("onto-cache")
        assert retrieved is not None
        assert retrieved.class_count == 15
        assert retrieved.property_count == 8

    async def test_get_metrics_nonexistent(self, computer: MetricsComputer) -> None:
        """get_metrics returns None for unknown ontology."""
        result = await computer.get_metrics("nonexistent-onto")
        assert result is None

    async def test_get_all_metrics(self, computer: MetricsComputer) -> None:
        """get_all_metrics returns all locally cached metrics."""
        event1 = {"type": "commit", "ontology_id": "onto-a", "class_count": 5}
        event2 = {"type": "commit", "ontology_id": "onto-b", "class_count": 10}
        await computer.compute_from_event(event1)
        await computer.compute_from_event(event2)
        all_metrics = await computer.get_all_metrics()
        assert len(all_metrics) == 2
        ids = {m.ontology_id for m in all_metrics}
        assert ids == {"onto-a", "onto-b"}

    async def test_get_stored_count(self, computer: MetricsComputer) -> None:
        """get_stored_count returns local-cache length when no Redis."""
        event = {"type": "commit", "ontology_id": "onto-count"}
        await computer.compute_from_event(event)
        count = await computer.get_stored_count()
        assert count == 1

    async def test_compute_from_event_multiple_updates(self, computer: MetricsComputer) -> None:
        """Computing for the same ontology_id updates the cached metrics."""
        event1 = {"type": "commit", "ontology_id": "onto-update", "axiom_count": 100}
        event2 = {"type": "commit", "ontology_id": "onto-update", "axiom_count": 300}
        await computer.compute_from_event(event1)
        await computer.compute_from_event(event2)
        retrieved = await computer.get_metrics("onto-update")
        assert retrieved is not None
        assert retrieved.axiom_count == 300  # last write wins

    async def test_close_without_redis(self, computer: MetricsComputer) -> None:
        """close is a no-op when Redis is not configured."""
        await computer.connect()
        await computer.close()  # should not raise

    async def test_unknown_event_type_defaults(self, computer: MetricsComputer) -> None:
        """Events with unexpected fields use default values safely."""
        event = {"type": "unknown"}
        metrics = await computer.compute_from_event(event)
        assert metrics.ontology_id == "unknown"
        assert metrics.axiom_count == 0
        assert metrics.class_count == 0
