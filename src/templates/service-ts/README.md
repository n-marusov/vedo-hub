# service-ts template

TypeScript Express template with OpenTelemetry, structured JSON logs, RED metrics, and health endpoints.

## Features
- OTLP tracing with `@opentelemetry/sdk-node`
- Structured logs with `trace_id` and `correlation_id`
- Log redaction middleware for sensitive values
- Endpoints: `GET /`, `GET /health`, `GET /ready`, `GET /metrics`
- Prometheus RED metrics via `prom-client`

## Run
```bash
npm install
npm run dev
```

## Build
```bash
npm run build
npm start
```
