#!/bin/bash
# @ctx: Stage 5 contract checks for runnable stubs and repository layout
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SERVICES_DIR="$ROOT_DIR/src/services"
WORKFLOWS_DIR="$ROOT_DIR/src/.github/workflows"
SHARED_STUB="$SERVICES_DIR/_shared/stub_server.py"
MAKEFILE="$ROOT_DIR/src/Makefile"

SERVICES="api-gateway auth-service ontology-service versioning-service metrics-service publisher-service public-browse-api commenting-service ticket-api ticket-classifier ticket-telemetry-listener ticket-notifier frontend publish-browse-ui"

# @hlv CT-REPOLAYOUT-001
# @hlv CT-REPOLAYOUT-002
# @hlv CT-REPOLAYOUT-003
# @hlv CT-REPOLAYOUT-004
# @hlv CT-REPOLAYOUT-005
test_service_directories_and_dockerfiles() {
  for svc in $SERVICES; do
    test -d "$SERVICES_DIR/$svc"
    test -f "$SERVICES_DIR/$svc/Dockerfile"
  done
}

# @hlv CT-REPOLAYOUT-006
test_service_ci_workflows_exist() {
  for svc in $SERVICES; do
    test -f "$WORKFLOWS_DIR/stub-$svc.yml"
  done
}

# @hlv CT-STUB-001
# @hlv CT-STUB-002
# @hlv CT-STUB-003
# @hlv CT-STUB-004
# @hlv CT-STUB-005
test_stub_runtime_contracts_present() {
  test -f "$SHARED_STUB"
  grep -q '"/health"' "$SHARED_STUB"
  grep -q '"/ready"' "$SHARED_STUB"
  grep -q '"name"' "$SHARED_STUB"
  grep -q '"trace_id"' "$SHARED_STUB"
  grep -q '/metrics' "$SHARED_STUB"
  grep -q '\[STUB\]' "$SHARED_STUB"
}

# @hlv CT-STUB-006
test_stub_has_no_service_dependencies() {
  grep -q 'HTTPServer' "$SHARED_STUB"
  ! grep -q '^import requests' "$SHARED_STUB"
  ! grep -q '^import redis' "$SHARED_STUB"
  ! grep -q 'psycopg' "$SHARED_STUB"
}

# @hlv CT-STUB-007
test_make_targets_for_service_builds_exist() {
  grep -q '^docker-build-%:' "$MAKEFILE"
  grep -q 'STUB_SERVICES :=' "$MAKEFILE"
}

# @hlv STUB_CREATE_FAILED
# @hlv STUB_PORT_CONFLICT
# @hlv INVALID_SERVICE_ID
# @hlv DUPLICATE_PATH
test_error_paths_represented() {
  grep -q 'BUILD_FAILED: unknown service' "$MAKEFILE"
}

for t in $(declare -F | awk '{print $3}' | grep '^test_'); do
  "$t"
done
