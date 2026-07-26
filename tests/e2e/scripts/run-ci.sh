#!/bin/bash
# E2E CI gate — runs API tests first (fast feedback), then GUI tests.
# Stops at first phase failure. GUI tests halt on first test failure.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"

fail() { echo "[E2E] $1"; exit 1; }

# ── Prerequisites ────────────────────────────────────────────────────────────
if ! command -v pnpm >/dev/null 2>&1; then
  echo "[E2E] pnpm unavailable — skipping Playwright run"
  exit 0
fi

cd "$ROOT_DIR"
pnpm install --frozen-lockfile
pnpm exec playwright install chromium

# ── Phase 1: API tests (fast) ──────────────────────────────────────────────
echo ""
echo "=== Phase 1/2: API tests ==="
pnpm exec playwright test --config=playwright.api.config.ts || fail "API tests failed"

# ── Phase 2: GUI tests (slow, stop on first failure) ────────────────────────
echo ""
echo "=== Phase 2/2: GUI tests ==="
pnpm exec playwright test --config=playwright.gui.config.ts || fail "GUI tests failed"

echo ""
echo "=== All E2E tests passed ==="
