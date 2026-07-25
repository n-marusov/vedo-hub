# service-python template

Python FastAPI template with OTLP tracing, JSON logs, log redaction, RED metrics, and management endpoints.

## Features
- OpenTelemetry SDK and FastAPI instrumentation
- JSON logging with `trace_id` and `correlation_id`
- Redaction middleware for password/token/secret patterns
- Endpoints: `GET /`, `GET /health`, `GET /ready`, `GET /metrics`
- RED metrics with `prometheus_client`

## Run
```bash
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
python main.py
```
