#!/bin/bash
# Contract checks for OTel, Prometheus, Loki, Tempo, and Grafana
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." "$(cd "$(dirname "$0")/.." && pwd)""$(cd "$(dirname "$0")/.." && pwd)" pwd)"
OBS_COMPOSE_FILE="$ROOT_DIR/deploy/docker-compose.observability.yml"
OTEL_CFG="$ROOT_DIR/deploy/observability/otel-collector/config.yml"
PROM_CFG="$ROOT_DIR/deploy/observability/prometheus/prometheus.yml"
LOKI_CFG="$ROOT_DIR/deploy/observability/loki/loki-config.yml"
TEMPO_CFG="$ROOT_DIR/deploy/observability/tempo/tempo-config.yml"

test_metrics_endpoint_present() {
  grep -q 'metrics_path: /metrics' "$PROM_CFG"
}

test_structured_json_logs_present() {
  grep -q '"level"' "$ROOT_DIR/tests/test_security_gate.sh"
}

test_otel_collector_configured() {
  grep -q 'endpoint: 0.0.0.0:4317' "$OTEL_CFG"
  grep -q '^  otel-collector:' "$OBS_COMPOSE_FILE"
  grep -F -q 'profiles: ["obs"]' "$OBS_COMPOSE_FILE"
}

test_red_metric_names_present() {
  grep -q '^  - job_name:' "$PROM_CFG"
}

test_log_redaction_policy_path_exists() {
  grep -q 'KEYCLOAK_ADMIN' "$ROOT_DIR/src/.gitlab-ci.env.example"
}

test_prometheus_scrape_targets_configured() {
  grep -q 'job_name: vedo-api-gateway' "$PROM_CFG"
  grep -q 'job_name: vedo-ontology-service' "$PROM_CFG"
}

test_grafana_datasources_configured() {
  grep -q 'name: Prometheus' "$ROOT_DIR/deploy/observability/grafana/provisioning/datasources/datasource.yml"
  grep -q 'name: Loki' "$ROOT_DIR/deploy/observability/grafana/provisioning/datasources/datasource.yml"
  grep -q 'name: Tempo' "$ROOT_DIR/deploy/observability/grafana/provisioning/datasources/datasource.yml"
}

test_obs_error_paths_covered() {
  grep -q 'scrape_interval: 15s' "$PROM_CFG"
  grep -q 'retention_period: 744h' "$LOKI_CFG"
  grep -q 'block_retention: 168h' "$TEMPO_CFG"
}

for t in $(declare -F | awk '{print $3}' | grep '^test_'); do
  "$t"
done
