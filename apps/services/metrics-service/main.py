"""Metrics service — event-driven analytics and monitoring.

Consumes telemetry events from RabbitMQ, aggregates metrics into Redis,
and exposes them via a REST API for the API Gateway and Grafana.
"""

from __future__ import annotations

import json
import logging
import os
import signal
import sys
import time
from contextlib import asynccontextmanager
from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from collections.abc import AsyncIterator

import aio_pika
import redis.asyncio as aioredis
from fastapi import FastAPI, Request
from fastapi.responses import PlainTextResponse
from prometheus_client import CONTENT_TYPE_LATEST, Counter, Gauge, generate_latest

from config import settings

if TYPE_CHECKING:
    from aio_pika.abc import AbstractIncomingMessage

SERVICE_NAME = settings.SERVICE_NAME
SERVICE_VERSION = settings.SERVICE_VERSION

# ─── Logging ───────────────────────────────────────────────────────────────────

logging.basicConfig(
    level=logging.DEBUG if os.getenv("LOG_LEVEL", "INFO").upper() == "DEBUG" else logging.INFO,
    format='{"service":"%(name)s","level":"%(levelname)s","message":"%(message)s"}',
    datefmt="%Y-%m-%dT%H:%M:%S%z",
)
logger = logging.getLogger(SERVICE_NAME)

# ─── Prometheus Metrics ────────────────────────────────────────────────────────

REQUESTS_TOTAL = Counter("vedo_requests_total", "Total requests", ["method", "path", "status"])
EVENTS_CONSUMED = Counter(
    "vedo_events_consumed_total", "Total events consumed from RabbitMQ", ["source"]
)
EVENTS_FAILED = Counter(
    "vedo_events_failed_total", "Total events that failed processing", ["source"]
)
ACTIVE_CONNECTIONS = Gauge("vedo_active_connections", "Currently active connections", ["type"])

# ─── Globals ───────────────────────────────────────────────────────────────────

request_count: int = 0
start_time: float = time.time()
redis_client: aioredis.Redis | None = None
rabbit_connection: aio_pika.Connection | None = None
rabbit_channel: aio_pika.Channel | None = None


# ─── RabbitMQ Consumer ─────────────────────────────────────────────────────────


async def process_event(message: AbstractIncomingMessage) -> None:
    """Process a single telemetry event from RabbitMQ."""
    async with message.process(requeue=True):
        try:
            body = json.loads(message.body.decode())
            source = body.get("source", "unknown")
            event_type = body.get("type", "unknown")

            # Store raw event in Redis with TTL
            if redis_client:
                key = f"event:{source}:{event_type}:{int(time.time())}"
                await redis_client.setex(key, 3600, message.body)

                # Increper aggregated counters in Redis
                await redis_client.hincrby(f"stats:{source}", event_type, 1)
                await redis_client.hincrby("stats:total", event_type, 1)

            EVENTS_CONSUMED.labels(source=source).inc()
            logger.info("event_processed", extra={"source": source, "type": event_type})

        except json.JSONDecodeError:
            EVENTS_FAILED.labels(source="parse_error").inc()
            logger.warning("event_parse_error", extra={"body": message.body[:200]})
        except Exception as exc:
            EVENTS_FAILED.labels(source="unknown").inc()
            logger.error("event_processing_error", extra={"error": str(exc)})


async def consume_events() -> None:
    """Connect to RabbitMQ and start consuming telemetry events."""
    global rabbit_connection, rabbit_channel  # noqa: PLW0603
    try:
        rabbit_connection = await aio_pika.connect_robust(settings.RABBITMQ_URL)
        rabbit_channel = await rabbit_connection.channel()
        await rabbit_channel.set_qos(prefetch_count=10)

        queue = await rabbit_channel.declare_queue("telemetry.events", durable=True)
        await queue.consume(process_event)

        ACTIVE_CONNECTIONS.labels(type="rabbitmq").set(1)
        logger.info("rabbitmq_consumer_started", extra={"queue": "telemetry.events"})
    except Exception as exc:
        logger.error("rabbitmq_connection_failed", extra={"error": str(exc)})


# ─── Lifespan ──────────────────────────────────────────────────────────────────


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:  # noqa: ARG001
    """Startup and shutdown lifecycle."""
    global redis_client  # noqa: PLW0603

    # Connect to Redis
    try:
        redis_client = aioredis.from_url(settings.REDIS_URL, decode_responses=True)
        await redis_client.ping()
        ACTIVE_CONNECTIONS.labels(type="redis").set(1)
        logger.info("redis_connected")
    except Exception as exc:
        logger.warning("redis_connection_failed", extra={"error": str(exc)})

    # Start RabbitMQ consumer in background
    task = None
    try:
        task = await consume_events()
    except Exception as exc:
        logger.warning("rabbitmq_consumer_setup_failed", extra={"error": str(exc)})

    yield

    # Shutdown
    if rabbit_channel:
        await rabbit_channel.close()
    if rabbit_connection:
        await rabbit_connection.close()
    if redis_client:
        await redis_client.close()
    if task:
        task.cancel()
    logger.info("shutdown_complete")


# ─── FastAPI App ───────────────────────────────────────────────────────────────

app = FastAPI(title=SERVICE_NAME, version=SERVICE_VERSION, lifespan=lifespan)


@app.middleware("http")
async def count_requests(request: Request, call_next: Any) -> Any:
    """Increment request counter and track metrics."""
    global request_count  # noqa: PLW0603
    request_count += 1
    response = await call_next(request)
    REQUESTS_TOTAL.labels(
        method=request.method, path=request.url.path, status=response.status_code
    ).inc()
    return response


# ─── Endpoints ─────────────────────────────────────────────────────────────────


@app.get("/")
async def root() -> dict[str, Any]:
    uptime_seconds = int(time.time() - start_time)
    return {
        "name": SERVICE_NAME,
        "version": SERVICE_VERSION,
        "description": "Event-driven metrics and analytics service",
        "uptime_seconds": uptime_seconds,
        "requests_served": request_count,
    }


@app.get("/health")
async def health() -> dict[str, str]:
    return {"status": "healthy"}


@app.get("/ready")
async def ready() -> dict[str, str]:
    checks: list[str] = []
    if redis_client:
        try:
            await redis_client.ping()
            checks.append("redis:ok")
        except Exception:
            checks.append("redis:unreachable")
    checks.append("http:ok")
    return {"status": "ready" if all(":ok" in c for c in checks) else "degraded", "checks": checks}


@app.get("/metrics")
async def metrics() -> PlainTextResponse:
    return PlainTextResponse(generate_latest().decode("utf-8"), media_type=CONTENT_TYPE_LATEST)


@app.get("/api/v1/stats/summary")
async def stats_summary() -> dict[str, Any]:
    """Return aggregated stats from Redis."""
    if not redis_client:
        return {"status": "redis_not_available", "stats": {}}
    try:
        total = await redis_client.hgetall("stats:total")
        return {"status": "ok", "stats": {k: int(v) for k, v in total.items()}}
    except Exception as exc:
        logger.error("stats_read_failed", extra={"error": str(exc)})
        return {"status": "error", "stats": {}}


# ─── Main ──────────────────────────────────────────────────────────────────────


def handle_sigterm(signum: int, frame: object) -> None:  # noqa: ARG001
    logger.info("sigterm_received")
    sys.exit(0)


if __name__ == "__main__":
    import uvicorn

    signal.signal(signal.SIGTERM, handle_sigterm)

    host = os.getenv("HOST", "0.0.0.0")
    port = int(os.getenv("SERVICE_PORT", "8084"))
    logger.info("starting", extra={"host": host, "port": port})
    uvicorn.run("main:app", host=host, port=port, log_config=None)
