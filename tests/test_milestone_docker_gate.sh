#!/bin/bash
# Milestone Docker Gate — verifies all service images build, compose starts,
# and all services report healthy. Intended to run at milestone completion.
#
# Usage:
#   ./test_milestone_docker_gate.sh              # full gate
#   ./test_milestone_docker_gate.sh --build-only  # skip compose up
#   ./test_milestone_docker_gate.sh --health-only # skip build
#   ./test_milestone_docker_gate.sh --skip-cleanup # leave containers running
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
COMPOSE_DIR="$ROOT_DIR/deploy"
COMPOSE_FILE="$COMPOSE_DIR/docker-compose.yml"

FAILED=0
PASS=0
SKIPPED=0

GATE="PASS"

log()    { echo "[$(date +%H:%M:%S)] $*"; }
pass()   { PASS=$((PASS + 1)); echo "  ✅ PASS: $*"; }
fail()   { FAILED=$((FAILED + 1)); GATE="FAIL"; echo "  ❌ FAIL: $*"; }
skip()   { SKIPPED=$((SKIPPED + 1)); echo "  ⏭️  SKIP: $*"; }

MODE="$@"  # --build-only | --health-only | --skip-cleanup

# ════════════════════════════════════════════════════════════════
# SECTION 1: Docker Build Verification
# ════════════════════════════════════════════════════════════════

test_compose_config_valid() {
    log "Gate: docker compose config — syntax validation"
    if (cd "$COMPOSE_DIR" && docker compose config --quiet 2>&1); then
        pass "docker compose config — valid syntax"
    else
        fail "docker compose config — INVALID syntax"
    fi
}

test_compose_services_listed() {
    log "Gate: verify expected services are defined in compose"
    local expected_services=(
        api-gateway auth-service ontology-service versioning-service
        metrics-service commenting-service publisher-service public-browse-api
        ticket-api ticket-classifier ticket-telemetry-listener ticket-notifier
        frontend publish-browse-ui document-extractor ai-orchestration-service
        neo4j postgres redis rabbitmq minio keycloak
    )
    local actual_services
    actual_services=$(cd "$COMPOSE_DIR" && docker compose config --services 2>/dev/null)

    local missing=0
    for svc in "${expected_services[@]}"; do
        if ! echo "$actual_services" | grep -qx "$svc"; then
            fail "Expected service '$svc' not found in compose"
            missing=$((missing + 1))
        fi
    done

    if [ "$missing" -eq 0 ]; then
        pass "All expected services present in docker-compose.yml"
    fi
}

test_dockerfiles_exist() {
    log "Gate: verify Dockerfile exists for each service"
    local missing=0

    while IFS= read -r svc; do
        [ -z "$svc" ] && continue
        # Skip infrastructure services that use 'image:' not 'build:'
        local df
        df=$(cd "$COMPOSE_DIR" && docker compose config 2>/dev/null \
            | grep -A2 "  ${svc}:" \
            | grep 'dockerfile:' \
            | awk '{print $2}' || true)
        if [ -z "$df" ]; then
            continue  # service uses pre-built image, skip
        fi
        # dockerfile path in compose is relative to context; resolve from project root
        local ctx
        ctx=$(cd "$COMPOSE_DIR" && docker compose config 2>/dev/null \
            | grep -A2 "  ${svc}:" \
            | grep 'context:' \
            | awk '{print $2}' || true)
        if [ -n "$ctx" ] && [ -n "$df" ]; then
            local full_path="$ROOT_DIR/$ctx/$df"
            if [ ! -f "$full_path" ]; then
                # try alternative path — context might be relative
                full_path="$ROOT_DIR/src/services/$df"
            fi
            if [ ! -f "$full_path" ]; then
                fail "Dockerfile not found for service '$svc' (looked: $full_path)"
                missing=$((missing + 1))
            fi
        fi
    done < <(cd "$COMPOSE_DIR" && docker compose config --services 2>/dev/null)

    if [ "$missing" -eq 0 ]; then
        pass "All service Dockerfiles exist"
    fi
}

test_build_all_images() {
    log "Gate: docker compose build — all services"
    log "  This may take several minutes (especially Rust services)..."
    local start
    start=$(date +%s)

    if (cd "$COMPOSE_DIR" && docker compose build 2>&1); then
        local elapsed=$(( $(date +%s) - start ))
        pass "docker compose build — all images built (${elapsed}s)"
    else
        fail "docker compose build — FAILED"
    fi
}

test_build_images_exist() {
    log "Gate: verify built images exist in local registry"
    local missing=0

    while IFS= read -r svc; do
        [ -z "$svc" ] && continue

        # Skip services that use pre-built images (no 'build:' section)
        local svc_block
        svc_block=$(cd "$COMPOSE_DIR" && docker compose config 2>/dev/null \
            | grep -A5 "^  ${svc}:" || true)
        if ! echo "$svc_block" | grep -q "build:"; then
            continue  # service uses pre-built image, no local build to verify
        fi

        local img
        img=$(cd "$COMPOSE_DIR" && docker compose config 2>/dev/null \
            | grep -A2 "^  ${svc}:" \
            | grep 'image:' \
            | awk '{print $2}' || true)
        if [ -z "$img" ]; then
            # Service uses 'build:' without explicit 'image:' — infer from project
            img="vedo-core-${svc}"
        fi
        if ! docker image inspect "$img" &>/dev/null; then
            fail "Image not found for service '$svc' (tried: $img)"
            missing=$((missing + 1))
        fi
    done < <(cd "$COMPOSE_DIR" && docker compose config --services 2>/dev/null)

    if [ "$missing" -eq 0 ]; then
        pass "All service images exist locally"
    fi
}

# ════════════════════════════════════════════════════════════════
# SECTION 2: Docker Compose Up & Health Verification
# ════════════════════════════════════════════════════════════════

test_compose_up() {
    log "Gate: docker compose up -d"
    # Clean up first
    (cd "$COMPOSE_DIR" && docker compose down --remove-orphans 2>/dev/null || true)

    if (cd "$COMPOSE_DIR" && docker compose up -d 2>&1); then
        pass "docker compose up -d — containers started"
    else
        fail "docker compose up -d — FAILED"
    fi
}

test_all_services_healthy() {
    log "Gate: wait for all services to become healthy (up to 180s)..."
    local max_wait=180
    local waited=0
    local interval=5

    while [ $waited -lt $max_wait ]; do
        local all_healthy=true
        local unhealthy_list=""
        local service_count=0

        while IFS=: read -r name health; do
            [ -z "$name" ] && continue
            service_count=$((service_count + 1))
            if [ "$health" != "healthy" ] && [ "$health" != "running" ]; then
                all_healthy=false
                unhealthy_list="$unhealthy_list $name($health)"
            fi
        done < <(cd "$COMPOSE_DIR" && docker compose ps --format '{{.Name}}:{{.Health}}' 2>/dev/null)

        if $all_healthy; then
            pass "All ${service_count} services healthy after ${waited}s"
            return 0
        fi

        if [ $((waited % 15)) -eq 0 ]; then
            log "  Waiting... (${waited}s/${max_wait}s) unhealthy:${unhealthy_list}"
        fi
        sleep $interval
        waited=$((waited + interval))
    done

    fail "Services did not become healthy within ${max_wait}s"
    log "  --- Container status dump ---"
    cd "$COMPOSE_DIR" && docker compose ps 2>&1 || true
    log "  --- Unhealthy container logs (last 10 lines each) ---"
    for svc in $(cd "$COMPOSE_DIR" && docker compose ps --format '{{.Name}}' 2>/dev/null); do
        local health
        health=$(cd "$COMPOSE_DIR" && docker compose ps --format '{{.Health}}' "$svc" 2>/dev/null)
        if [ "$health" != "healthy" ]; then
            log "  --- $svc ($health) ---"
            cd "$COMPOSE_DIR" && docker compose logs --tail=10 "$svc" 2>/dev/null || true
        fi
    done
    return 1
}

test_containers_running() {
    log "Gate: verify all services have running containers"
    local expected_count
    expected_count=$(cd "$COMPOSE_DIR" && docker compose config --services 2>/dev/null | wc -l)

    # Retry loop — containers may briefly lag during startup
    local retries=3
    local actual_count=0
    for i in $(seq 1 $retries); do
        actual_count=$(cd "$COMPOSE_DIR" && docker compose ps --format '{{.Name}}' 2>/dev/null | wc -l)
        if [ "$actual_count" -ge "$expected_count" ]; then
            break
        fi
        sleep 2
    done

    if [ "$actual_count" -ge "$expected_count" ]; then
        pass "All ${expected_count} services have running containers (${actual_count} total)"
    else
        fail "Expected ${expected_count} containers, got ${actual_count}"
    fi
}

test_no_restarting_containers() {
    log "Gate: check for restarting/crashed containers"
    local restarting
    restarting=$(cd "$COMPOSE_DIR" && docker compose ps --format '{{.Name}}:{{.Status}}' 2>/dev/null \
        | grep -i 'restarting\|exited' || true)
    if [ -z "$restarting" ]; then
        pass "No restarting or exited containers"
    else
        fail "Containers with issues: $restarting"
    fi
}

# ════════════════════════════════════════════════════════════════
# CLEANUP
# ════════════════════════════════════════════════════════════════

cleanup() {
    if [[ "$MODE" != *"--skip-cleanup"* ]]; then
        log "Cleanup: shutting down compose services..."
        (cd "$COMPOSE_DIR" && docker compose down --remove-orphans 2>/dev/null || true)
        log "Cleanup: complete"
    fi
}

# ════════════════════════════════════════════════════════════════
# RUN ALL
# ════════════════════════════════════════════════════════════════

run_all() {
    trap cleanup EXIT

    echo ""
    echo "═══════════════════════════════════════════════════════════════════"
    echo "  Milestone Docker Gate — Build + Health Verification"
    echo "  Compose: $COMPOSE_FILE"
    echo "  Mode: ${MODE:-full}"
    echo "═══════════════════════════════════════════════════════════════════"
    echo ""

    # ── Section 1: Build ────────────────────────────────────────────
    echo "─── Section 1: Build Verification ───"

    test_compose_config_valid || true
    test_compose_services_listed || true
    test_dockerfiles_exist || true

    if [[ "$MODE" != *"--health-only"* ]]; then
        test_build_all_images || true
        test_build_images_exist || true
    else
        skip "Build tests skipped (--health-only mode)"
    fi

    # ── Section 2: Health ────────────────────────────────────────────
    echo ""
    echo "─── Section 2: Health Verification ───"

    if [[ "$MODE" != *"--build-only"* ]]; then
        test_compose_up || true
        test_containers_running || true
        test_all_services_healthy || true
        test_no_restarting_containers || true
    else
        skip "Health tests skipped (--build-only mode)"
    fi

    # ── Summary ──────────────────────────────────────────────────────
    echo ""
    echo "═══════════════════════════════════════════════════════════════════"
    if [ "$GATE" = "PASS" ]; then
        echo "  ✅ GATE PASSED — ${PASS} passed, ${FAILED} failed, ${SKIPPED} skipped"
    else
        echo "  ❌ GATE FAILED — ${PASS} passed, ${FAILED} failed, ${SKIPPED} skipped"
    fi
    echo "═══════════════════════════════════════════════════════════════════"
    echo "<!-- milestone-docker-gate: ${GATE} -->"
    echo "<!-- passed: ${PASS} | failed: ${FAILED} | skipped: ${SKIPPED} -->"

    # Return failure if any tests failed
    [ "$GATE" = "PASS" ]
}

run_all
