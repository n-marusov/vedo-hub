# service-go template

Go template service with OpenTelemetry tracing, JSON logging, Prometheus RED metrics, and management endpoints.

## Features
- OTEL tracer bootstrap with OTLP gRPC exporter
- JSON logs with `trace_id` and `correlation_id`
- Log redaction for sensitive keywords
- Endpoints: `GET /`, `GET /health`, `GET /ready`, `GET /metrics`
- RED metrics exposed via Prometheus handler

## Run
```bash
go run .
```

## Build
```bash
go build .
```
