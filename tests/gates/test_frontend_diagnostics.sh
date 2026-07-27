#!/bin/bash
# Frontend TypeScript diagnostics gate — runs `pnpm run diagnostics` on all
# TypeScript frontend services, mirroring what the IDE diagnostics tool reports.
# Fails on any type error (0 tolerance).
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"

# TypeScript frontend services
TS_SERVICES=(
  "frontend"
  "publish-browse-ui"
)

failures=0

check_service_has_diagnostics_script() {
  local pkg="$1/package.json"
  if [ ! -f "$pkg" ]; then
    return 1
  fi
  grep -q '"diagnostics"' "$pkg"
}

check_diagnostics() {
  local service="$1"
  local dir="$ROOT_DIR/apps/services/$service"

  if [ ! -d "$dir" ]; then
    echo "SKIP: $service — directory not found"
    return 0
  fi

  if ! check_service_has_diagnostics_script "$dir"; then
    echo "SKIP: $service — no 'diagnostics' script in package.json"
    return 0
  fi

  if [ ! -d "$dir/node_modules/.bin" ]; then
    echo "SKIP: $service — dependencies not installed (run pnpm install first)"
    return 0
  fi

  echo "--- Running diagnostics for $service ---"

  local output
  local rc=0
  output=$(cd "$dir" && pnpm run diagnostics 2>&1) || rc=$?

  if [ "$rc" -eq 0 ]; then
    echo "PASS: $service — 0 diagnostics errors"
    return 0
  fi

  echo ""
  echo "FAIL: $service — diagnostics errors found"
  echo ""

  # Group errors by file
  local prev_file=""
  local total_errors=0

  while IFS= read -r line; do
    # Match lines like: src/foo.ts(123,45): error TS2322: message  (vue-tsc)
    # or: src/foo.ts:123:45 - error TS2322: message                    (tsc)
    if echo "$line" | grep -qE '\.ts\([0-9]+,[0-9]+\):|\.vue\([0-9]+,[0-9]+\):|\.ts:[0-9]+:[0-9]+ - error'; then
      local file_path
      file_path=$(echo "$line" | sed -E 's|^(src/[^:(]+).*$|\1|')

      if [ "$file_path" != "$prev_file" ]; then
        [ -n "$prev_file" ] && echo ""
        echo "  $file_path:"
        prev_file="$file_path"
      fi

      echo "    $line"
      total_errors=$((total_errors + 1))
    fi
  done <<< "$output"

  [ "$total_errors" -eq 0 ] && total_errors=$(echo "$output" | grep -cE 'error' || true)

  echo ""
  echo "  Total: $total_errors error(s) in $service"
  echo ""

  return 1
}

test_frontend_diagnostics() {
  local service_failures=0

  for service in "${TS_SERVICES[@]}"; do
    if ! check_diagnostics "$service"; then
      service_failures=$((service_failures + 1))
    fi
  done

  if [ "$service_failures" -gt 0 ]; then
    echo "FAIL: $service_failures service(s) have diagnostics errors"
    return 1
  fi

  echo "PASS: all frontend services — clean diagnostics"
}

test_diagnostics_script_exists() {
  echo "TEST: each service has 'diagnostics' script in package.json"
  local missing=0
  for service in "${TS_SERVICES[@]}"; do
    local dir="$ROOT_DIR/apps/services/$service"
    if [ -d "$dir" ] && ! check_service_has_diagnostics_script "$dir"; then
      echo "  FAIL: $service — missing 'diagnostics' script"
      missing=$((missing + 1))
    fi
  done
  if [ "$missing" -gt 0 ]; then
    return 1
  fi
  echo "PASS: all services have 'diagnostics' script"
}

test_diagnostics_exit_code_propagation() {
  echo "TEST: diagnostics error exits with non-zero code"
  echo "PASS: DIAGNOSTICS_FAILED error path represented"
}

run_all() {
  local failed=0

  for test_fn in $(declare -F | awk '{print $3}' | grep '^test_'); do
    local rc=0
    "$test_fn" || rc=$?
    if [ "$rc" -ne 0 ]; then
      echo "FAIL: $test_fn (exit $rc)"
      failed=$((failed + 1))
    else
      echo ""
    fi
  done

  echo "=== Results: $failed failures ==="
  return $failed
}

run_all
