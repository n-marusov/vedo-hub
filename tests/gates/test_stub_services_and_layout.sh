#!/bin/bash
# Stage 5 contract checks for runnable stubs and repository layout
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
SERVICES_DIR="$ROOT_DIR/apps/services"
MAKEFILE="$ROOT_DIR/Makefile"

SERVICES="api-gateway auth-service ontology-service versioning-service metrics-service publisher-service public-browse-api commenting-service ticket-api ticket-classifier ticket-telemetry-listener ticket-notifier frontend publish-browse-ui"

test_service_directories_and_dockerfiles() {
  for svc in $SERVICES; do
    test -d "$SERVICES_DIR/$svc"
    test -f "$SERVICES_DIR/$svc/Dockerfile"
  done
}

test_make_targets_for_service_builds_exist() {
  grep -q '^docker-build-%:' "$MAKEFILE"
  grep -q 'STUB_SERVICES :=' "$MAKEFILE"
}

test_error_paths_represented() {
  grep -q 'BUILD_FAILED: unknown service' "$MAKEFILE"
}

for t in $(declare -F | awk '{print $3}' | grep '^test_'); do
  "$t"
done
