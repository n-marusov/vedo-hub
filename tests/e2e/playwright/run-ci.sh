#!/bin/bash
# @ctx: Playwright CI integration entrypoint for stage 3
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"

if command -v npm >/dev/null 2>&1; then
  cd "$ROOT_DIR"
  npm install
  npx playwright install --with-deps chromium firefox webkit
  npx playwright test
else
  echo "[E2E] npm is unavailable; Playwright runtime execution skipped in this environment"
fi
