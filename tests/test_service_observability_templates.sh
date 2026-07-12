#!/bin/bash
# Stage 4 checks for service observability templates
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
TEMPLATES_DIR="$ROOT_DIR/src/templates"

test_all_templates_have_metrics_endpoint() {
  grep -q '"/metrics"' "$TEMPLATES_DIR/service-rust/src/main.rs"
  grep -q '"/metrics"' "$TEMPLATES_DIR/service-go/main.go"
  grep -q '"/metrics"' "$TEMPLATES_DIR/service-python/main.py"
  grep -q '"/metrics"' "$TEMPLATES_DIR/service-ts/src/index.ts"
}

test_all_templates_emit_structured_logs() {
  grep -q 'trace_id' "$TEMPLATES_DIR/service-rust/src/main.rs"
  grep -q 'trace_id' "$TEMPLATES_DIR/service-go/main.go"
  grep -q 'trace_id' "$TEMPLATES_DIR/service-python/main.py"
  grep -q 'trace_id' "$TEMPLATES_DIR/service-ts/src/index.ts"
  grep -q 'correlation_id' "$TEMPLATES_DIR/service-rust/src/main.rs"
  grep -q 'correlation_id' "$TEMPLATES_DIR/service-go/main.go"
  grep -q 'correlation_id' "$TEMPLATES_DIR/service-python/main.py"
  grep -q 'correlation_id' "$TEMPLATES_DIR/service-ts/src/index.ts"
}

test_all_templates_bootstrap_otel_sdk() {
  grep -q 'OTEL_EXPORTER_OTLP_ENDPOINT' "$TEMPLATES_DIR/service-rust/src/main.rs"
  grep -q 'OTEL_EXPORTER_OTLP_ENDPOINT' "$TEMPLATES_DIR/service-go/main.go"
  grep -q 'OTEL_EXPORTER_OTLP_ENDPOINT' "$TEMPLATES_DIR/service-python/main.py"
  grep -q 'OTEL_EXPORTER_OTLP_ENDPOINT' "$TEMPLATES_DIR/service-ts/src/index.ts"
}

test_all_templates_define_red_metrics() {
  grep -q 'vedo_requests_total' "$TEMPLATES_DIR/service-rust/src/main.rs"
  grep -q 'vedo_request_errors_total' "$TEMPLATES_DIR/service-rust/src/main.rs"
  grep -q 'vedo_request_duration_seconds' "$TEMPLATES_DIR/service-rust/src/main.rs"
  grep -q 'vedo_requests_total' "$TEMPLATES_DIR/service-go/main.go"
  grep -q 'vedo_request_errors_total' "$TEMPLATES_DIR/service-go/main.go"
  grep -q 'vedo_request_duration_seconds' "$TEMPLATES_DIR/service-go/main.go"
  grep -q 'vedo_requests_total' "$TEMPLATES_DIR/service-python/main.py"
  grep -q 'vedo_request_errors_total' "$TEMPLATES_DIR/service-python/main.py"
  grep -q 'vedo_request_duration_seconds' "$TEMPLATES_DIR/service-python/main.py"
  grep -q 'vedo_requests_total' "$TEMPLATES_DIR/service-ts/src/index.ts"
  grep -q 'vedo_request_errors_total' "$TEMPLATES_DIR/service-ts/src/index.ts"
  grep -q 'vedo_request_duration_seconds' "$TEMPLATES_DIR/service-ts/src/index.ts"
}

test_all_templates_include_log_redaction() {
  grep -q 'REDACTED' "$TEMPLATES_DIR/service-rust/src/main.rs"
  grep -q 'REDACTED' "$TEMPLATES_DIR/service-go/main.go"
  grep -q 'REDACTED' "$TEMPLATES_DIR/service-python/main.py"
  grep -q 'REDACTED' "$TEMPLATES_DIR/service-ts/src/index.ts"
}

test_all_templates_have_metadata_and_health_handlers() {
  grep -q '"/"' "$TEMPLATES_DIR/service-rust/src/main.rs"
  grep -q '"/health"' "$TEMPLATES_DIR/service-rust/src/main.rs"
  grep -q '"/ready"' "$TEMPLATES_DIR/service-rust/src/main.rs"
  grep -q '"name"' "$TEMPLATES_DIR/service-go/main.go"
  grep -q '"/health"' "$TEMPLATES_DIR/service-go/main.go"
  grep -q '"/ready"' "$TEMPLATES_DIR/service-go/main.go"
  grep -q '@app.get("/")' "$TEMPLATES_DIR/service-python/main.py"
  grep -q '@app.get("/health")' "$TEMPLATES_DIR/service-python/main.py"
  grep -q '@app.get("/ready")' "$TEMPLATES_DIR/service-python/main.py"
  grep -q 'app.get("/",' "$TEMPLATES_DIR/service-ts/src/index.ts"
  grep -q 'app.get("/health",' "$TEMPLATES_DIR/service-ts/src/index.ts"
  grep -q 'app.get("/ready",' "$TEMPLATES_DIR/service-ts/src/index.ts"
}

# @hlv METRICS_EXPORT_FAILED
# @hlv TRACE_EXPORT_FAILED
# @hlv LOG_REDACTION_FAILED
test_error_paths_are_represented() {
  grep -q 'OTEL_EXPORTER_OTLP_ENDPOINT' "$TEMPLATES_DIR/service-rust/src/main.rs"
  grep -q 'request_errors_total' "$TEMPLATES_DIR/service-go/main.go"
  grep -q 'BatchSpanProcessor' "$TEMPLATES_DIR/service-python/main.py"
  grep -q 'redact' "$TEMPLATES_DIR/service-ts/src/index.ts"
}

test_template_docs_exist() {
  test -f "$TEMPLATES_DIR/service-rust/README.md"
  test -f "$TEMPLATES_DIR/service-go/README.md"
  test -f "$TEMPLATES_DIR/service-python/README.md"
  test -f "$TEMPLATES_DIR/service-ts/README.md"
}

for t in $(declare -F | awk '{print $3}' | grep '^test_'); do
  "$t"
done
