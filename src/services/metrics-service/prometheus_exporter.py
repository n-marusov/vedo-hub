"""Prometheus metrics exporter for the metrics service.

Exposes ontology-level Prometheus metrics for monitoring and dashboards.
"""

import prometheus_client
from prometheus_client import Gauge, Counter, Histogram

# ─── Metric Definitions ────────────────────────────────────────────────────────

# Ontology-level metrics (per ontology)
ontology_axiom_count = Gauge(
    "vedo_ontology_axiom_count",
    "Number of axioms in an ontology",
    ["ontology_id"],
)

ontology_class_count = Gauge(
    "vedo_ontology_class_count",
    "Number of classes in an ontology",
    ["ontology_id"],
)

ontology_property_count = Gauge(
    "vedo_ontology_property_count",
    "Number of properties in an ontology",
    ["ontology_id"],
)

ontology_individual_count = Gauge(
    "vedo_ontology_individual_count",
    "Number of individuals in an ontology",
    ["ontology_id"],
)

ontology_max_class_depth = Gauge(
    "vedo_ontology_max_class_depth",
    "Maximum class hierarchy depth in an ontology",
    ["ontology_id"],
)

ontology_property_density = Gauge(
    "vedo_ontology_property_density",
    "Property-to-class ratio in an ontology",
    ["ontology_id"],
)

# Service-level metrics
events_consumed_total = Counter(
    "vedo_metrics_events_consumed_total",
    "Total number of domain events consumed",
    ["event_type"],
)

events_failed_total = Counter(
    "vedo_metrics_events_failed_total",
    "Total number of failed event processing attempts",
    ["event_type"],
)

metrics_computed_total = Counter(
    "vedo_metrics_computed_total",
    "Total number of metric computations performed",
    ["ontology_id"],
)

computation_duration_seconds = Histogram(
    "vedo_metrics_computation_duration_seconds",
    "Duration of metric computation in seconds",
    buckets=[0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0],
)

# Cached metrics count
cached_metrics_count = Gauge(
    "vedo_metrics_cached_count",
    "Number of ontology metrics currently cached",
)


def record_ontology_metrics(ontology_id: str, metrics: dict) -> None:
    """Record ontology-level Prometheus metrics from a metrics dict."""
    ontology_axiom_count.labels(ontology_id=ontology_id).set(metrics.get("axiom_count", 0))
    ontology_class_count.labels(ontology_id=ontology_id).set(metrics.get("class_count", 0))
    ontology_property_count.labels(ontology_id=ontology_id).set(metrics.get("property_count", 0))
    ontology_individual_count.labels(ontology_id=ontology_id).set(metrics.get("individual_count", 0))
    ontology_max_class_depth.labels(ontology_id=ontology_id).set(metrics.get("max_class_depth", 0))
    ontology_property_density.labels(ontology_id=ontology_id).set(metrics.get("property_density", 0.0))


def record_event_consumed(event_type: str) -> None:
    """Record a consumed event."""
    events_consumed_total.labels(event_type=event_type).inc()


def record_event_failed(event_type: str) -> None:
    """Record a failed event."""
    events_failed_total.labels(event_type=event_type).inc()


def record_metrics_computed(ontology_id: str, duration: float) -> None:
    """Record a metric computation."""
    metrics_computed_total.labels(ontology_id=ontology_id).inc()
    computation_duration_seconds.observe(duration)


def update_cached_count(count: int) -> None:
    """Update the cached metrics count."""
    cached_metrics_count.set(count)


def generate_metrics() -> str:
    """Generate Prometheus-format metrics output."""
    return prometheus_client.generate_latest().decode("utf-8")
