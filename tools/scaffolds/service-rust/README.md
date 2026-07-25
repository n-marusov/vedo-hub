# service-rust template

Rust template service with OpenTelemetry, structured JSON logs, RED metrics, and health endpoints.

## Features
- OTLP export via `OTEL_EXPORTER_OTLP_ENDPOINT` (default `http://otel-collector:4317`)
- JSON logs with `trace_id` and `correlation_id`
- Log redaction middleware for sensitive keywords
- Endpoints: `GET /`, `GET /health`, `GET /ready`, `GET /metrics`
- RED metrics: `vedo_requests_total`, `vedo_request_errors_total`, `vedo_request_duration_seconds`

## Run
```bash
cargo run
```

## Build
```bash
cargo build
```
