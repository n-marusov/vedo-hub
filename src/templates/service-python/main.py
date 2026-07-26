import logging
import os
import re
import time
from collections.abc import Callable

from fastapi import FastAPI, Request
from fastapi.responses import PlainTextResponse
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from prometheus_client import CONTENT_TYPE_LATEST, Counter, Histogram, generate_latest
from pythonjsonlogger import jsonlogger

REQUEST_TOTAL = Counter("vedo_requests_total", "Total requests", ["method", "path", "status"])
REQUEST_ERRORS = Counter("vedo_request_errors_total", "Total failed requests", ["method", "path"])
REQUEST_DURATION = Histogram(
    "vedo_request_duration_seconds", "Request duration", ["method", "path"]
)

REDACTION_PATTERN = re.compile(r"(password|token|secret)", re.IGNORECASE)

HTTP_ERROR_THRESHOLD = 400


app = FastAPI(title="service-python-template")


def init_observability() -> None:
    endpoint = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4317")
    resource = Resource.create({"service.name": "service-python-template"})
    provider = TracerProvider(resource=resource)
    provider.add_span_processor(
        BatchSpanProcessor(OTLPSpanExporter(endpoint=endpoint, insecure=True))
    )
    trace.set_tracer_provider(provider)


def redact(value: str) -> str:
    return REDACTION_PATTERN.sub("[REDACTED]", value)


def setup_logging() -> logging.Logger:
    handler = logging.StreamHandler()
    formatter = jsonlogger.JsonFormatter(
        "%(asctime)s %(levelname)s %(message)s %(trace_id)s %(correlation_id)s"
    )
    handler.setFormatter(formatter)
    logger = logging.getLogger("service-python-template")
    logger.setLevel(logging.INFO)
    logger.handlers = [handler]
    return logger


logger = setup_logging()
init_observability()
FastAPIInstrumentor.instrument_app(app)


@app.middleware("http")
async def metrics_and_logging(request: Request, call_next: Callable):
    start = time.time()
    correlation_id = request.headers.get("x-correlation-id", "generated-correlation-id")
    response = await call_next(request)

    method = request.method
    path = request.url.path
    status = response.status_code
    REQUEST_TOTAL.labels(method=method, path=path, status=str(status)).inc()
    REQUEST_DURATION.labels(method=method, path=path).observe(time.time() - start)
    if status >= HTTP_ERROR_THRESHOLD:
        REQUEST_ERRORS.labels(method=method, path=path).inc()

    span = trace.get_current_span()
    trace_id = "00000000000000000000000000000000"
    if span.get_span_context().is_valid:
        trace_id = f"{span.get_span_context().trace_id:032x}"

    logger.info(
        redact("request_complete"),
        extra={
            "trace_id": trace_id,
            "correlation_id": correlation_id,
            "path": path,
            "method": method,
            "status": status,
        },
    )
    return response


@app.get("/")
async def root():
    return {
        "name": "service-python-template",
        "version": "0.1.0",
        "description": "Python service template with observability",
        "stub": True,
    }


@app.get("/health")
async def health():
    return {"status": "healthy"}


@app.get("/ready")
async def ready():
    return {"status": "ready"}


@app.get("/metrics")
async def metrics():
    return PlainTextResponse(generate_latest().decode("utf-8"), media_type=CONTENT_TYPE_LATEST)


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("main:app", host="0.0.0.0", port=8080)
