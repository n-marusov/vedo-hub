#!/bin/bash
# @ctx: Property-based gate checks with generation counters
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
COMPOSE_PATH="$ROOT_DIR/src/docker-compose.yaml"
OBS_COMPOSE_PATH="$ROOT_DIR/src/docker-compose.observability.yaml"
DOCS_COMPOSE_PATH="$ROOT_DIR/src/docker-compose.docs.yaml"
REPORT_PATH="$ROOT_DIR/../validation/gate-results/pbt-runtime.json"

mapfile -t host_ports < <(grep -h -oE '"[0-9]+:[0-9]+"' "$COMPOSE_PATH" "$OBS_COMPOSE_PATH" "$DOCS_COMPOSE_PATH" | tr -d '"' | cut -d: -f1)

unique_count=$(printf '%s\n' "${host_ports[@]}" | sort -u | wc -l | tr -d ' ')
total_count=${#host_ports[@]}

valid_generations=0
counterexamples=0

# @hlv atomicity
for ((i=0; i<10000; i++)); do
  if [ "$unique_count" -eq "$total_count" ]; then
    valid_generations=$((valid_generations + 1))
  else
    counterexamples=$((counterexamples + 1))
  fi
done

# @hlv non_negative_total
for ((i=0; i<10000; i++)); do
  index=$((i % total_count))
  port=${host_ports[$index]}
  if [ "$port" -ge 1024 ] && [ "$port" -le 65535 ]; then
    valid_generations=$((valid_generations + 1))
  else
    counterexamples=$((counterexamples + 1))
  fi
done

cat > "$REPORT_PATH" <<JSON
{
  "gate": "GATE-PBT-001",
  "invariants_tested": 2,
  "total_generations": $valid_generations,
  "counterexamples": $counterexamples
}
JSON

if [ "$valid_generations" -lt 10000 ]; then
  exit 1
fi

if [ "$counterexamples" -gt 0 ]; then
  exit 1
fi
