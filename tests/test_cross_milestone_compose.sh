#!/bin/bash
# @ctx: Cross-milestone docker compose build + health check scenario
# @hlv DEPLOY-OPS-001
# @hlv STUB-NATIVE-001
# @hlv SERVICE_UNHEALTHY
# @hlv PORT_CONFLICT
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SRC_DIR="$ROOT_DIR/src"
COMPOSE_FILE="$SRC_DIR/docker-compose.yaml"

FAILED=0
PASS=0

log() { echo "[$(date +%H:%M:%S)] $*"; }

pass() { PASS=$((PASS + 1)); echo "  PASS: $*"; }
fail() { FAILED=$((FAILED + 1)); echo "  FAIL: $*"; }

# @hlv BUILD_FAILED
test_compose_build() {
    log "Building all compose services..."
    if (cd "$SRC_DIR" && docker compose build 2>&1); then
        pass "docker compose build succeeded"
    else
        fail "docker compose build failed"
        return 1
    fi
}

# @hlv SERVICE_NOT_FOUND
test_compose_up() {
    log "Starting all compose services..."
    # Clean up first
    (cd "$SRC_DIR" && docker compose down --remove-orphans 2>/dev/null || true)
    if (cd "$SRC_DIR" && docker compose up -d 2>&1); then
        pass "docker compose up -d succeeded"
    else
        fail "docker compose up -d failed"
        return 1
    fi
}

# @hlv SERVICE_UNHEALTHY
test_all_services_healthy() {
    log "Waiting for all services to become healthy (up to 120s)..."
    local max_wait=120
    local waited=0
    local interval=5

    while [ $waited -lt $max_wait ]; do
        local all_healthy=true
        local unhealthy_list=""

        while IFS= read -r line; do
            local name="$line"
            [ -z "$name" ] && continue
            local health
            health=$(cd "$SRC_DIR" && docker compose ps --format '{{.Name}}:{{.Health}}' 2>/dev/null | grep "^${name}:" | head -1 | cut -d: -f2)
            if [ "$health" != "healthy" ] && [ "$health" != "running" ]; then
                all_healthy=false
                unhealthy_list="$unhealthy_list $name($health)"
            fi
        done < <(cd "$SRC_DIR" && docker compose ps --format '{{.Name}}' 2>/dev/null)

        if $all_healthy; then
            pass "All services healthy after ${waited}s"
            return 0
        fi

        log "  Waiting... (${waited}s/${max_wait}s) unhealthy:${unhealthy_list}"
        sleep $interval
        waited=$((waited + interval))
    done

    fail "Services did not become healthy within ${max_wait}s"
    # Show status of unhealthy services
    cd "$SRC_DIR" && docker compose ps 2>&1 || true
    return 1
}

# @hlv structured_logging_only
test_health_endpoints() {
    log "Checking health endpoints..."
    local services
    services=$(cd "$SRC_DIR" && docker compose ps --format '{{.Name}}' 2>/dev/null)

    while IFS= read -r name; do
        [ -z "$name" ] && continue
        local port
        port=$(cd "$SRC_DIR" && docker compose port "$name" 8080 2>/dev/null | cut -d: -f2)
        if [ -z "$port" ]; then
            port=$(cd "$SRC_DIR" && docker compose port "$name" 3000 2>/dev/null | cut -d: -f2)
        fi
        if [ -z "$port" ]; then
            log "  SKIP: $name (no mapped port for health check)"
            continue
        fi

        local http_code
        http_code=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:${port}/health" 2>/dev/null || echo "000")
        if [ "$http_code" = "200" ]; then
            pass "$name /health returned 200 (port $port)"
        else
            fail "$name /health returned $http_code (port $port)"
        fi
    done <<< "$services"
}

# @hlv STUB_MISSING
test_metadata_endpoints() {
    log "Checking metadata endpoints..."
    local services
    services=$(cd "$SRC_DIR" && docker compose ps --format '{{.Name}}' 2>/dev/null)

    while IFS= read -r name; do
        [ -z "$name" ] && continue
        local port
        port=$(cd "$SRC_DIR" && docker compose port "$name" 8080 2>/dev/null | cut -d: -f2)
        if [ -z "$port" ]; then
            port=$(cd "$SRC_DIR" && docker compose port "$name" 3000 2>/dev/null | cut -d: -f2)
        fi
        if [ -z "$port" ]; then
            continue
        fi

        local response
        response=$(curl -s "http://127.0.0.1:${port}/" 2>/dev/null || echo "")
        if echo "$response" | grep -q '"name"'; then
            pass "$name / returned valid metadata"
        else
            fail "$name / did not return valid metadata"
        fi
    done <<< "$services"
}

# Cleanup
cleanup() {
    log "Shutting down compose services..."
    (cd "$SRC_DIR" && docker compose down --remove-orphans 2>/dev/null || true)
}

run_all() {
    trap cleanup EXIT

    test_compose_build || true
    test_compose_up || true
    test_all_services_healthy || true
    test_health_endpoints || true
    test_metadata_endpoints || true

    echo ""
    echo "=== Results: $PASS passed, $FAILED failures ==="
    return $FAILED
}

run_all
