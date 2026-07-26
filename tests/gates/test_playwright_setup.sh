#!/bin/bash
# Contract checks for Playwright E2E infrastructure
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
E2E_ROOT="$ROOT_DIR/tests/e2e/playwright"

test_playwright_config_exists() {
  test -f "$E2E_ROOT/playwright.config.ts"
  grep -q "chromium" "$E2E_ROOT/playwright.config.ts"
  grep -q "firefox" "$E2E_ROOT/playwright.config.ts"
  grep -q "webkit" "$E2E_ROOT/playwright.config.ts"
}

test_fixtures_exist() {
  test -f "$E2E_ROOT/tests/fixtures.ts"
  grep -q "seededUsers" "$E2E_ROOT/tests/fixtures.ts"
}

test_reporter_config_exists() {
  grep -q "reporter" "$E2E_ROOT/playwright.config.ts"
}

test_trace_capture_config_exists() {
  grep -q "trace: 'retain-on-failure'" "$E2E_ROOT/playwright.config.ts"
  grep -q "video: 'retain-on-failure'" "$E2E_ROOT/playwright.config.ts"
}

test_e2e_error_paths_covered() {
  grep -q "timeout" "$E2E_ROOT/playwright.config.ts"
  grep -q "webServer" "$E2E_ROOT/playwright.config.ts"
}

for t in $(declare -F | awk '{print $3}' | grep '^test_'); do
  "$t"
done
