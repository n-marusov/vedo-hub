"""Metrics service — ontology analytics and Prometheus metrics.

Consumes domain events (commit, import, publish), computes ontology metrics
(axiom count, class depth, property density), and exposes Prometheus-compatible
metrics for dashboard queries.
"""

import asyncio
import json
import logging
import os
import time
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request
from fastapi.responses import PlainTextResponse

from analytics import MetricsComputer
from collectors import EventCollector
from prometheus_exporter import (
    generate_metrics,
    record_event_consumed,
    record_event_failed,
    record_metrics_computed,
    record_ontology_metrics,
    update_cached_count,
)

SERVICE_NAME = "metrics-service"
DEFAULT_PORT = 8084

# ─── Logging ───────────────────────────────────────────────────────────────────

logging.basicConfig(
    level=logging.INFO,
    format='{"service":"%(name)s","level":"%(levelname)s","message":"%(message)s"}',
    datefmt="%Y-%m-%dT%H:%M:%S%z",
)
logger = logging.getLogger(SERVICE_NAME)

# ─── Globals ───────────────────────────────────────────────────────────────────

computer: MetricsComputer | None = None
collector: EventCollector | None = None
request_count: int = 0
start_time: float = time.time()


async def handle_event(event: dict) -> None:
    """Handle a domain event by computing metrics."""
    global computer

    if computer is None:
        return

    event_type = event.get("type", "unknown")
    ontology_id = event.get("ontology_id", "unknown")

    record_event_consumed(event_type)
    start = time.time()

    try:
        metrics = await computer.compute_from_event(event)
        duration = time.time() - start

        record_metrics_computed(ontology_id, duration)
        record_ontology_metrics(
            ontology_id,
            {
                "axiom_count": metrics.axiom_count,
                "class_count": metrics.class_count,
                "property_count": metrics.property_count,
                "individual_count": metrics.individual_count,
                "max_class_depth": metrics.max_class_depth,
                "property_density": metrics.property_density,
            },
        )

        count = await computer.get_stored_count()
        update_cached_count(count)

        logger.info(
            "Processed %s event for ontology %s in %.3fs",
            event_type,
            ontology_id,
            duration,
        )
    except Exception as exc:
        record_event_failed(event_type)
        logger.error("Failed to process event %s: %s", event_type, exc)


# ─── Lifecycle ─────────────────────────────────────────────────────────────────


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan: initialize and clean up resources."""
    global computer, collector

    # Read configuration from environment
    redis_url = os.environ.get("REDIS_URL")
    rabbitmq_url = os.environ.get("RABBITMQ_URL")

    logger.info(
        "Starting metrics-service (redis=%s, rabbitmq=%s)",
        redis_url or "none",
        rabbitmq_url or "none",
    )

    # Initialize MetricsComputer
    computer = MetricsComputer(redis_url=redis_url)
    await computer.connect()

    # Start event collector
    collector = EventCollector(
        rabbitmq_url=rabbitmq_url,
        event_handler=handle_event,
    )

    # Start consumer in background
    consume_task = asyncio.create_task(collector.start())

    logger.info("Metrics-service started")

    yield  # Application runs here

    # Cleanup
    if collector:
        await collector.stop()
    consume_task.cancel()
    try:
        await consume_task
    except asyncio.CancelledError:
        pass
    if computer:
        await computer.close()

    logger.info("Metrics-service stopped")


# ─── FastAPI Application ───────────────────────────────────────────────────────

app = FastAPI(
    title=SERVICE_NAME,
    version="0.3.0",
    lifespan=lifespan,
)


@app.get("/")
async def root():
    """Service information endpoint."""
    return {
        "name": SERVICE_NAME,
        "version": "0.3.0",
        "description": "Metrics service — ontology analytics and Prometheus metrics",
        "stub": False,
    }


@app.get("/health")
async def health():
    """Health check endpoint."""
    redis_status = "disconnected"
    if computer and computer._redis:
        try:
            await computer._redis.ping()
            redis_status = "connected"
        except Exception:
            redis_status = "disconnected"

    return {
        "status": "healthy",
        "service": SERVICE_NAME,
        "redis": redis_status,
    }


@app.get("/ready")
async def ready():
    """Readiness check endpoint."""
    return {
        "status": "ready",
        "service": SERVICE_NAME,
    }


@app.get("/metrics")
async def metrics():
    """Prometheus metrics endpoint."""
    return PlainTextResponse(
        content=generate_metrics(),
        media_type="text/plain; version=0.0.4",
    )


@app.post("/api/v1/events/ingest")
async def ingest_event(request: Request):
    """HTTP ingestion endpoint for domain events (fallback when RabbitMQ unavailable)."""
    global collector
    try:
        event = await request.json()
    except json.JSONDecodeError:
        return {"error": "INVALID_EVENT", "message": "Request body must be valid JSON"}

    if collector is None:
        return {"error": "COLLECTOR_UNAVAILABLE", "message": "Event collector not initialized"}

    success = await collector.ingest_event(event)
    status_code = 202 if success else 500

    return {"accepted": success, "event_type": event.get("type", "unknown")}


@app.get("/api/v1/metrics/ontologies")
async def list_metrics():
    """List all cached ontology metrics."""
    global computer
    if computer is None:
        return {"metrics": [], "total": 0}

    all_metrics = await computer.get_all_metrics()
    return {
        "metrics": [
            {
                "ontology_id": m.ontology_id,
                "axiom_count": m.axiom_count,
                "class_count": m.class_count,
                "property_count": m.property_count,
                "individual_count": m.individual_count,
                "max_class_depth": m.max_class_depth,
                "avg_class_depth": m.avg_class_depth,
                "property_density": m.property_density,
            }
            for m in all_metrics
        ],
        "total": len(all_metrics),
    }


@app.get("/api/v1/metrics/ontologies/{ontology_id}")
async def get_ontology_metrics(ontology_id: str):
    """Get cached metrics for a specific ontology."""
    global computer
    if computer is None:
        return {"error": "METRICS_UNAVAILABLE", "message": "Metrics computer not initialized"}

    metrics = await computer.get_metrics(ontology_id)
    if metrics is None:
        return {"error": "NOT_FOUND", "message": f"No metrics found for ontology {ontology_id}"}

    return {
        "ontology_id": metrics.ontology_id,
        "axiom_count": metrics.axiom_count,
        "class_count": metrics.class_count,
        "property_count": metrics.property_count,
        "individual_count": metrics.individual_count,
        "max_class_depth": metrics.max_class_depth,
        "avg_class_depth": metrics.avg_class_depth,
        "property_density": metrics.property_density,
    }


# ─── Main Entry Point ──────────────────────────────────────────────────────────


def main() -> None:
    """Start the metrics service."""
    import uvicorn

    port = int(os.getenv("SERVICE_PORT", str(DEFAULT_PORT)))
    logger.info("Starting %s on port %d", SERVICE_NAME, port)
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=port,
        log_level="info",
        access_log=True,
    )


if __name__ == "__main__":
    main()
