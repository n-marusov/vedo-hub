"""Tests for prometheus_exporter — Prometheus metric recording and generation.

Validates: REQ-FUN.INTEGRATION.collaboration-quality-metrics
"""

from __future__ import annotations

from prometheus_exporter import (
    generate_metrics,
    record_event_consumed,
    record_event_failed,
    record_metrics_computed,
    record_ontology_metrics,
    update_cached_count,
)


class TestPrometheusExporter:
    """Prometheus exporter tests — verifies metrics are recorded and rendered."""

    def test_generate_metrics_returns_valid_output(self) -> None:
        """generate_metrics returns Prometheus-format text."""
        output = generate_metrics()
        assert isinstance(output, str)
        assert len(output) > 0
        assert "# HELP" in output
        assert "# TYPE" in output

    def test_record_ontology_metrics_sets_gauges(self) -> None:
        """record_ontology_metrics sets all ontology gauges for an ontology_id."""
        record_ontology_metrics(
            "test-onto-1",
            {
                "axiom_count": 42,
                "class_count": 10,
                "property_count": 5,
                "individual_count": 20,
                "max_class_depth": 3,
                "property_density": 0.5,
            },
        )
        output = generate_metrics()
        assert "# HELP vedo_ontology_axiom_count" in output
        assert 'ontology_id="test-onto-1"' in output
        assert "42.0" in output  # axiom_count value

    def test_record_ontology_metrics_partial(self) -> None:
        """Missing keys in metrics dict default to 0."""
        record_ontology_metrics(
            "test-onto-partial",
            {
                "axiom_count": 99,
            },
        )
        output = generate_metrics()
        assert 'ontology_id="test-onto-partial"' in output

    def test_record_event_consumed_increments_counter(self) -> None:
        """record_event_consumed increments the event counter."""
        record_event_consumed("commit")
        output = generate_metrics()
        assert "vedo_metrics_events_consumed_total" in output

    def test_record_event_failed_increments_counter(self) -> None:
        """record_event_failed increments the failure counter."""
        record_event_failed("import")
        output = generate_metrics()
        assert "vedo_metrics_events_failed_total" in output

    def test_record_metrics_computed_increments(self) -> None:
        """record_metrics_computed increments computation counter and records duration."""
        record_metrics_computed("onto-compute-test", 0.05)
        output = generate_metrics()
        assert "vedo_metrics_computed_total" in output
        assert "vedo_metrics_computation_duration_seconds" in output

    def test_update_cached_count(self) -> None:
        """update_cached_count sets the cached metrics gauge."""
        update_cached_count(42)
        output = generate_metrics()
        assert "vedo_metrics_cached_count" in output

    def test_record_multiple_ontologies(self) -> None:
        """Metrics for multiple ontology_ids all appear in generated output."""
        record_ontology_metrics("multi-a", {"axiom_count": 10})
        record_ontology_metrics("multi-b", {"axiom_count": 20})
        output = generate_metrics()
        assert 'ontology_id="multi-a"' in output
        assert 'ontology_id="multi-b"' in output
