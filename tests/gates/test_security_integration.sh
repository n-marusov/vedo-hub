#!/bin/bash
# Security integration gate — runs BOLA/BFLA/RBAC full-stack authorization tests
# Requires: Docker test stack (API Gateway, Keycloak, auth-service, PostgreSQL, Neo4j)
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"

# detect Go across platforms (Linux, macOS, Windows/Git Bash, WSL)
detect_go() {
  if command -v go &>/dev/null; then
    echo "go"
    return 0
  fi
  if command -v go.exe &>/dev/null; then
    echo "go.exe"
    return 0
  fi
  for candidate in \
    "/mnt/c/Program Files/Go/bin/go.exe" \
    "/c/Program Files/Go/bin/go.exe" \
    "/mnt/c/Program Files (x86)/Go/bin/go.exe" \
    "/c/Program Files (x86)/Go/bin/go.exe" \
    "$LOCALAPPDATA/Programs/Go/bin/go.exe" \
    "$HOME/go/bin/go.exe"; do
    if [ -x "$candidate" ]; then
      echo "$candidate"
      return 0
    fi
  done
  return 1
}

GO_CMD=$(detect_go) || true
if [ -z "$GO_CMD" ]; then
  echo "=== Go not available — skipping security integration tests ==="
  exit 0
fi

echo "=== Running BOLA/BFLA/RBAC full-stack integration tests with $GO_CMD ==="

# check API Gateway availability before running tests
if ! curl -sf http://localhost:8080/api/v1/health >/dev/null 2>&1; then
  echo "ERROR: API Gateway not available at http://localhost:8080"
  echo ""
  echo "  The security integration tests require a running Docker test stack."
  echo "  Start it with:"
  echo ""
  echo "    make docker-up-test"
  echo ""
  echo "  This starts all services (API Gateway, Keycloak, auth-service,"
  echo "  PostgreSQL, Neo4j, etc.) needed for full-stack authorization tests."
  exit 1
fi

# run BOLA/BFLA fixture tests and comprehensive RBAC suite (integration tag required)
cd "$ROOT/tests/security/authorization" && "$GO_CMD" test -tags=integration ./... -count=1 -v 2>&1

echo "=== Security integration suite complete ==="
