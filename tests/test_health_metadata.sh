#!/bin/bash
# @ctx: Contract checks for health, ready, root metadata, and endpoint conventions
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
COMPOSE_FILE="$ROOT_DIR/src/docker-compose.yaml"
SERVICES_DIR="$ROOT_DIR/src/services"

ALL_PASS=1

check() {
  local label="$1"
  shift
  if "$@"; then
    echo "PASS: $label"
  else
    echo "FAIL: $label"
    ALL_PASS=0
  fi
}

# @hlv CT-API-001
check "compose has api-gateway" grep -q '^  api-gateway:' "$COMPOSE_FILE"
check "compose maps api-gateway 8080" grep -q '"8080:8080"' "$COMPOSE_FILE"

# @hlv CT-API-002
check "go stubs expose /health" grep -R -q '"/health"' "$SERVICES_DIR" --include='main.go'
check "python stubs expose /health" grep -R -q '"/health"' "$SERVICES_DIR" --include='main.py'
check "ts stubs expose /health" grep -R -q 'location = /health' "$SERVICES_DIR" --include='nginx.conf'

# @hlv CT-API-003
check "go stubs expose /ready" grep -R -q '"/ready"' "$SERVICES_DIR" --include='main.go'
check "python stubs expose /ready" grep -R -q '"/ready"' "$SERVICES_DIR" --include='main.py'
check "ts stubs expose /ready" grep -R -q 'location = /ready' "$SERVICES_DIR" --include='nginx.conf'

# @hlv CT-API-004
check "root metadata has name" grep -R -q '"name"' "$SERVICES_DIR"
check "root metadata has version" grep -R -q '"version"' "$SERVICES_DIR"
check "root metadata has stub flag" grep -R -q '"stub"' "$SERVICES_DIR"

# @hlv CT-API-005
check "unknown path uses ENDPOINT_NOT_FOUND" grep -R -q 'ENDPOINT_NOT_FOUND' "$SERVICES_DIR"
check "method not allowed path exists" grep -R -q 'METHOD_NOT_ALLOWED' "$SERVICES_DIR"

# @hlv CT-API-006
check "compose has health checks" grep -q 'healthcheck:' "$COMPOSE_FILE"
check "compose health checks target /health" grep -q '/health' "$COMPOSE_FILE"

if [ "$ALL_PASS" -eq 0 ]; then
  echo "FAILED: one or more contract checks failed"
  exit 1
fi

echo "ALL PASSED"
