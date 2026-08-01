#!/bin/bash
# Cross-milestone Playwright smoke test — verifies all frontend pages serve correctly
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
SRC_DIR="$ROOT_DIR/src"
COMPOSE_FILE="$SRC_DIR/docker-compose.yaml"

FAILED=0
PASS=0

log() { echo "[$(date +%H:%M:%S)] $*"; }

pass() { PASS=$((PASS + 1)); echo "  PASS: $*"; }
fail() { FAILED=$((FAILED + 1)); echo "  FAIL: $*"; }

# Pages to test on main frontend (port 3000)
FRONTEND_PAGES=(
  "/"
  "/login"
  "/dashboard"
  "/ontology"
  "/tickets"
  "/members"
  "/metrics"
  "/sparql"
  "/versioning"
  "/validation"
  "/settings"
)

# Pages to test on publish-browse-ui (port 3002)
PUBLISH_PAGES=(
  "/"
  "/browse"
  "/catalog"
)

test_compose_build() {
    log "Building frontend and publish-browse-ui..."
    if (cd "$SRC_DIR" && docker compose build frontend publish-browse-ui 2>&1); then
        pass "docker compose build frontend + publish-browse-ui succeeded"
    else
        fail "docker compose build failed"
        return 1
    fi
}

test_compose_up() {
    log "Starting frontend and publish-browse-ui..."
    (cd "$SRC_DIR" && docker compose down frontend publish-browse-ui --remove-orphans 2>/dev/null || true)
    if (cd "$SRC_DIR" && docker compose up -d frontend publish-browse-ui 2>&1); then
        pass "docker compose up frontend + publish-browse-ui succeeded"
    else
        fail "docker compose up failed"
        return 1
    fi
}

test_services_healthy() {
    log "Waiting for services to become healthy (up to 60s)..."
    local max_wait=60
    local waited=0
    local interval=5

    while [ $waited -lt $max_wait ]; do
        local fe_health pub_health
        fe_health=$(cd "$SRC_DIR" && docker compose ps --format '{{.Name}}:{{.Health}}' 2>/dev/null | grep "frontend" | head -1 | cut -d: -f2)
        pub_health=$(cd "$SRC_DIR" && docker compose ps --format '{{.Name}}:{{.Health}}' 2>/dev/null | grep "publish-browse-ui" | head -1 | cut -d: -f2)

        if [ "$fe_health" = "healthy" ] && [ "$pub_health" = "healthy" ]; then
            pass "Both services healthy after ${waited}s"
            return 0
        fi

        log "  Waiting... (${waited}s/${max_wait}s) frontend=$fe_health publish=$pub_health"
        sleep $interval
        waited=$((waited + interval))
    done

    fail "Services did not become healthy within ${max_wait}s"
    return 1
}

test_frontend_pages() {
    log "Testing main frontend pages (port 3000)..."

    for page in "${FRONTEND_PAGES[@]}"; do
        local http_code body
        http_code=$(curl -s -o /tmp/fe_response.html -w "%{http_code}" "http://127.0.0.1:3000${page}" 2>/dev/null || echo "000")
        body=$(cat /tmp/fe_response.html 2>/dev/null || echo "")

        if [ "$http_code" = "200" ]; then
            if echo "$body" | grep -q '<div id="app"></div>'; then
                pass "frontend ${page} returned 200 with Vue app mount point"
            else
                fail "frontend ${page} returned 200 but missing <div id=\"app\"></div>"
            fi
        else
            fail "frontend ${page} returned HTTP ${http_code}"
        fi
    done
}

test_publish_pages() {
    log "Testing publish-browse-ui pages (port 3002)..."

    for page in "${PUBLISH_PAGES[@]}"; do
        local http_code body
        http_code=$(curl -s -o /tmp/pub_response.html -w "%{http_code}" "http://127.0.0.1:3002${page}" 2>/dev/null || echo "000")
        body=$(cat /tmp/pub_response.html 2>/dev/null || echo "")

        if [ "$http_code" = "200" ]; then
            if echo "$body" | grep -q '<div id="app"></div>'; then
                pass "publish-browse-ui ${page} returned 200 with Vue app mount point"
            else
                fail "publish-browse-ui ${page} returned 200 but missing <div id=\"app\"></div>"
            fi
        else
            fail "publish-browse-ui ${page} returned HTTP ${http_code}"
        fi
    done
}

test_playwright_smoke() {
    log "Running Playwright smoke test against both frontends..."

    local test_dir="$ROOT_DIR/tests/playwright-smoke"
    mkdir -p "$test_dir"

    cat > "$test_dir/frontend-smoke.spec.ts" << 'PWEOF'
import { test, expect } from '@playwright/test';

const FRONTEND_URL = process.env.FRONTEND_URL || 'http://127.0.0.1:3000';
const PUBLISH_URL = process.env.PUBLISH_URL || 'http://127.0.0.1:3002';

const frontendPages = [
  { path: '/', title: 'VEDO Core' },
  { path: '/login', title: 'VEDO Core' },
  { path: '/dashboard', title: 'VEDO Core' },
  { path: '/ontology', title: 'VEDO Core' },
  { path: '/tickets', title: 'VEDO Core' },
  { path: '/members', title: 'VEDO Core' },
  { path: '/metrics', title: 'VEDO Core' },
  { path: '/sparql', title: 'VEDO Core' },
  { path: '/versioning', title: 'VEDO Core' },
  { path: '/validation', title: 'VEDO Core' },
  { path: '/settings', title: 'VEDO Core' },
];

const publishPages = [
  { path: '/', title: 'VEDO Core' },
  { path: '/browse', title: 'VEDO Core' },
  { path: '/catalog', title: 'VEDO Core' },
];

test.describe('Main Frontend SPA pages', () => {
  for (const page of frontendPages) {
    test(`should load ${page.path}`, async ({ page: pwPage }) => {
      await pwPage.goto(`${FRONTEND_URL}${page.path}`);
      await expect(pwPage.locator('#app')).toBeVisible();
      await expect(pwPage).toHaveTitle(new RegExp(page.title));
    });
  }
});

test.describe('Publish Browse UI SPA pages', () => {
  for (const page of publishPages) {
    test(`should load ${page.path}`, async ({ page: pwPage }) => {
      await pwPage.goto(`${PUBLISH_URL}${page.path}`);
      await expect(pwPage.locator('#app')).toBeVisible();
      await expect(pwPage).toHaveTitle(new RegExp(page.title));
    });
  }
});
PWEOF

    cat > "$test_dir/playwright.config.ts" << 'PWCFG'
import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: '.',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL: process.env.FRONTEND_URL || 'http://127.0.0.1:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
});
PWCFG

    cat > "$test_dir/package.json" << 'PKGJSON'
{
  "name": "playwright-smoke",
  "version": "1.0.0",
  "private": true,
  "scripts": {
    "test": "playwright test"
  },
  "devDependencies": {
    "@playwright/test": "^1.49.0"
  }
}
PKGJSON

    if [ ! -d "$test_dir/node_modules" ]; then
        log "Installing Playwright dependencies..."
        (cd "$test_dir" && npm install 2>&1) || {
            fail "npm install for playwright-smoke failed"
            return 1
        }
    fi

    log "Running Playwright tests..."
    if (cd "$test_dir" && FRONTEND_URL="http://127.0.0.1:3000" PUBLISH_URL="http://127.0.0.1:3002" npx playwright test 2>&1); then
        pass "Playwright smoke tests passed"
    else
        fail "Playwright smoke tests failed"
        return 1
    fi
}

# Cleanup
cleanup() {
    log "Shutting down compose services..."
    (cd "$SRC_DIR" && docker compose down frontend publish-browse-ui --remove-orphans 2>/dev/null || true)
    rm -f /tmp/fe_response.html /tmp/pub_response.html
}

run_all() {
    trap cleanup EXIT

    test_compose_build || true
    test_compose_up || true
    test_services_healthy || true
    test_frontend_pages || true
    test_publish_pages || true
    test_playwright_smoke || true

    echo ""
    echo "=== Results: $PASS passed, $FAILED failures ==="
    return $FAILED
}

run_all
