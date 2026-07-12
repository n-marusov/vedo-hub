#!/bin/bash
# Contract gate runner — native stub and milestone 003 contract checks
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# detect Go across platforms (Linux, macOS, Windows/Git Bash, WSL)
detect_go() {
  if command -v go &>/dev/null; then echo "go"; return 0; fi
  if command -v go.exe &>/dev/null; then echo "go.exe"; return 0; fi
  for candidate in \
    "/mnt/c/Program Files/Go/bin/go.exe" \
    "/c/Program Files/Go/bin/go.exe" \
    "$LOCALAPPDATA/Programs/Go/bin/go.exe" \
    "$HOME/go/bin/go.exe"; do
    if [ -x "$candidate" ]; then echo "$candidate"; return 0; fi
  done
  return 1
}

GO_CMD=$(detect_go) || true
if [ -n "$GO_CMD" ]; then
  echo "=== Running Go contract tests with $GO_CMD ==="
  # each Go module has its own go.mod — run from module root
  for mod in "$ROOT/src/services/api-gateway" "$ROOT/src/cli" "$ROOT/src/services/auth-service" "$ROOT/src/services/auth-service/org" "$ROOT/src/services/ticket-api" "$ROOT/src/services/ticket-telemetry-listener" "$ROOT/src/services/ticket-notifier" "$ROOT/src/services/ticket-sync"; do
    echo "  testing: $mod"
    (cd "$mod" && "$GO_CMD" test ./... -count=1)
  done
  # ticket-api integration tests (separate Go module)
  if [ -d "$ROOT/tests/ticket-api" ]; then
    echo "  testing: $ROOT/tests/ticket-api"
    (cd "$ROOT/tests/ticket-api" && "$GO_CMD" test ./... -count=1)
  fi
else
  echo "=== Go not available — skipping Go contract tests ==="
fi

# existing milestone 001/002 contract tests
bash "$ROOT/tests/test_build_commands.sh"
bash "$ROOT/tests/test_docker_build.sh"
bash "$ROOT/tests/test_ci_and_compose.sh"
bash "$ROOT/tests/test_health_metadata.sh"
bash "$ROOT/tests/test_native_stubs.sh"
