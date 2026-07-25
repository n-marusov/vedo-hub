"""Document extractor service — AI-assisted ontology extraction from documents.

Accepts documents (MD, TXT, PDF, DOCX, JSON, XML, CSV, XLSX), parses them into
structured content, sends to LLM for ontology extraction, validates the resulting
sequence, and applies it to the ontology-service via gRPC ApplySequence.
"""

from __future__ import annotations

import logging
import os
import time
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.responses import PlainTextResponse

from config import settings

SERVICE_NAME = settings.SERVICE_NAME
SERVICE_VERSION = settings.SERVICE_VERSION

# ─── Logging ───────────────────────────────────────────────────────────────────

logging.basicConfig(
    level=logging.DEBUG if os.getenv("LOG_LEVEL", "INFO").upper() == "DEBUG" else logging.INFO,
    format='{"service":"%(name)s","level":"%(levelname)s","message":"%(message)s"}',
    datefmt="%Y-%m-%dT%H:%M:%S%z",
)
logger = logging.getLogger(SERVICE_NAME)

# ─── Globals ───────────────────────────────────────────────────────────────────

request_count: int = 0
start_time: float = time.time()

# Lazy imports for route modules — loaded at lifespan start
# to avoid circular deps with config.
api_routes = None


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    """Application lifespan: initialize and clean up resources."""
    global api_routes

    logger.info(
        "Starting %s v%s (port=%d, max_file=%dMB)",
        SERVICE_NAME,
        SERVICE_VERSION,
        settings.SERVICE_PORT,
        settings.MAX_FILE_SIZE_MB,
    )
    logger.info(
        "LLM provider=%s model=%s ontology=%s",
        settings.LLM_PROVIDER,
        settings.LLM_MODEL,
        settings.ONTOLOGY_SERVICE_URL,
    )

    # Import and register API routes
    from api.routes import router as extract_router

    api_routes = extract_router
    app.include_router(extract_router, prefix="/api/v1")

    logger.info("%s started", SERVICE_NAME)

    yield

    logger.info("%s stopped", SERVICE_NAME)


# ─── FastAPI Application ───────────────────────────────────────────────────────

app = FastAPI(
    title=SERVICE_NAME,
    version=SERVICE_VERSION,
    lifespan=lifespan,
)


@app.get("/")
async def root() -> dict:
    """Service information endpoint."""
    return {
        "name": SERVICE_NAME,
        "version": SERVICE_VERSION,
        "description": "Document extractor — AI-assisted ontology extraction from documents",
        "stub": False,
    }


@app.get("/health")
async def health() -> dict:
    """Health check endpoint."""
    return {
        "status": "healthy",
        "service": SERVICE_NAME,
    }


@app.get("/ready")
async def ready() -> dict:
    """Readiness check endpoint."""
    return {
        "status": "ready",
        "service": SERVICE_NAME,
    }


@app.get("/metrics")
async def metrics() -> PlainTextResponse:
    """Prometheus metrics endpoint."""
    global request_count
    text = (
        f"# HELP vedo_service_requests_total Total service requests\n"
        f"# TYPE vedo_service_requests_total counter\n"
        f'vedo_service_requests_total{{service="{SERVICE_NAME}"}} {request_count}\n'
        f"# HELP vedo_service_uptime_seconds Service uptime\n"
        f"# TYPE vedo_service_uptime_seconds gauge\n"
        f'vedo_service_uptime_seconds{{service="{SERVICE_NAME}"}} {time.time() - start_time:.0f}\n'
    )
    return PlainTextResponse(
        content=text,
        media_type="text/plain; version=0.0.4",
    )


# ─── Main Entry Point ──────────────────────────────────────────────────────────


def main() -> None:
    """Start the document extractor service."""
    import uvicorn

    port = settings.SERVICE_PORT
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
