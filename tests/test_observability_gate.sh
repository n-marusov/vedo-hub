#!/bin/bash
# @ctx: Observability gate checks for marker and capability coverage
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
REPORT_PATH="$ROOT_DIR/../validation/gate-results/observability-runtime.json"
PROM_CFG="$ROOT_DIR/src/infra/observability/prometheus/prometheus.yml"
COMPOSE_FILE="$ROOT_DIR/src/docker-compose.yaml"
OBS_COMPOSE_FILE="$ROOT_DIR/src/docker-compose.observability.yaml"

check_marker() {
  local marker="$1"
  grep -R -q "@hlv ${marker}" "$ROOT_DIR/src" "$ROOT_DIR/tests"
}

# @hlv log_entry_exit
check_marker "log_entry_exit"
# @hlv log_all_errors
check_marker "log_all_errors"
# @hlv log_state_changes
check_marker "log_state_changes"
# @hlv log_external_calls
check_marker "log_external_calls"
# @hlv request_correlation
check_marker "request_correlation"
# @hlv log_levels_correct
check_marker "log_levels_correct"

# @hlv structured_logging_only
grep -R -q '"level"' "$ROOT_DIR/tests"

# @hlv CT-OBS-009
# @hlv CT-OBS-010
! grep -q '^storage:' "$PROM_CFG"
grep -q '^  prometheus:' "$OBS_COMPOSE_FILE"
grep -q 'profiles: \["observability"\]' "$OBS_COMPOSE_FILE"
grep -q -- '--storage.tsdb.retention.time=30d' "$OBS_COMPOSE_FILE"

cat > "$REPORT_PATH" <<JSON
{
  "gate": "GATE-OBS-001",
  "required_capabilities": ["metrics", "trace_spans", "structured_logs"],
  "status": "passed"
}
JSON
