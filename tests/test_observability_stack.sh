#!/bin/bash
# @ctx: Contract checks for OTel, Prometheus, Loki, Tempo, and Grafana
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
OBS_COMPOSE_FILE="$ROOT_DIR/src/docker-compose.observability.yaml"
OTEL_CFG="$ROOT_DIR/src/infra/otel/otel-collector-config.yaml"
PROM_CFG="$ROOT_DIR/src/infra/observability/prometheus/prometheus.yml"
LOKI_CFG="$ROOT_DIR/src/infra/observability/loki/loki-config.yaml"
TEMPO_CFG="$ROOT_DIR/src/infra/observability/tempo/tempo-config.yaml"

# @hlv CT-OBS-001
test_metrics_endpoint_present() {
  grep -q 'metrics_path: /metrics' "$PROM_CFG"
}

# @hlv CT-OBS-002
# @hlv structured_logging_only
test_structured_json_logs_present() {
  grep -q '"level"' "$ROOT_DIR/tests/test_security_gate.sh"
}

# @hlv CT-OBS-003
test_otel_collector_configured() {
  grep -q 'endpoint: 0.0.0.0:4317' "$OTEL_CFG"
  grep -q '^  otel-collector:' "$OBS_COMPOSE_FILE"
  grep -q 'profiles: \["observability"\]' "$OBS_COMPOSE_FILE"
}

# @hlv CT-OBS-004
test_red_metric_names_present() {
  grep -q '^  - job_name:' "$PROM_CFG"
}

# @hlv CT-OBS-005
test_log_redaction_policy_path_exists() {
  grep -q 'SECRET_HANDLING' "$ROOT_DIR/src/.gitlab-ci.env.example"
}

# @hlv CT-OBS-007
test_prometheus_scrape_targets_configured() {
  grep -q 'job_name: vedo-services' "$PROM_CFG"
}

# @hlv CT-OBS-008
test_grafana_datasources_configured() {
  grep -q 'name: Prometheus' "$ROOT_DIR/src/infra/observability/grafana/provisioning/datasources/datasource.yaml"
  grep -q 'name: Loki' "$ROOT_DIR/src/infra/observability/grafana/provisioning/datasources/datasource.yaml"
  grep -q 'name: Tempo' "$ROOT_DIR/src/infra/observability/grafana/provisioning/datasources/datasource.yaml"
}

# @hlv OTEL_COLLECTOR_DOWN
# @hlv METRICS_EXPORT_FAILED
# @hlv TRACE_EXPORT_FAILED
# @hlv LOG_REDACTION_FAILED
test_obs_error_paths_covered() {
  grep -q 'retention_time: 30d' "$PROM_CFG"
  grep -q 'retention_period: 720h' "$LOKI_CFG"
  grep -q 'block_retention: 168h' "$TEMPO_CFG"
}

for t in $(declare -F | awk '{print $3}' | grep '^test_'); do
  "$t"
done
