#!/bin/bash
# BOLA/BFLA unit gate — runs auth middleware unit tests only
# Integration tests moved to test_security_integration.sh
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
  echo "=== Go not available — skipping BOLA/BFLA unit tests ==="
  exit 0
fi

echo "=== Running BOLA/BFLA/RBAC unit test suite with $GO_CMD ==="

# run auth middleware tests with BOLA/BFLA focus (all CT-SEC-* tests)
cd "$ROOT/apps/services/api-gateway" && "$GO_CMD" test ./auth/... -v -count=1 -run "TestCT_SEC|TestProperty|TestInvariant" 2>&1

# run org-level access control tests (membership, policies, visibility enforcement)
cd "$ROOT/apps/services/auth-service/org" && "$GO_CMD" test ./... -count=1 2>&1

echo "=== BOLA/BFLA/RBAC unit suite complete ===""
