#!/bin/bash
# Performance gate checks for stage 2 scripts
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
REPORT_PATH="$ROOT_DIR/../validation/gate-results/performance-runtime.json"

start_ts=$(date +%s)
bash "$ROOT_DIR/tests/gates/test_ci_and_compose.sh" > /tmp/vedo_perf_test.log
end_ts=$(date +%s)

duration=$((end_ts - start_ts))
failed_lines=$(grep -c '^FAIL:' /tmp/vedo_perf_test.log || true)
total_lines=$(grep -c '^TEST:' /tmp/vedo_perf_test.log || true)

if [ "$total_lines" -eq 0 ]; then
  error_rate=1
else
  error_rate=$(awk -v f="$failed_lines" -v t="$total_lines" 'BEGIN { printf "%.6f", f / t }')
fi

cat > "$REPORT_PATH" <<JSON
{
  "gate": "GATE-PERF-001",
  "duration_seconds": $duration,
  "failed": $failed_lines,
  "total": $total_lines,
  "error_rate": $error_rate
}
JSON

if [ "$duration" -gt 120 ]; then
  exit 1
fi

awk -v e="$error_rate" 'BEGIN { exit (e <= 0.001 ? 0 : 1) }'
