#!/usr/bin/env bash
# Integration test runner for VEDO Core Rust services.
#
# Starts Neo4j and PostgreSQL via Docker, runs ontology-service and
# versioning-service integration tests against them, then cleans up.
#
# Usage:
#   bash src/services/scripts/run_integration_rust.sh
#
# Environment:
#   SKIP_DOCKER=1    — skip container lifecycle (use existing instances)
#   NEO4J_URI        — override Neo4j Bolt URI  (default: bolt://localhost:7687)
#   NEO4J_USER       — override Neo4j user      (default: neo4j)
#   NEO4J_PASSWORD   — override Neo4j password  (default: password)
#   PG_URL           — override PG connection   (default: postgres://postgres:password@localhost:5432/vedo_versioning)
#   RUST_LOG         — logger filter            (default: info)
#
# Requirements:
#   - Docker (or compatible, e.g. Podman with `docker` alias)
#   - bash 4+
#   - cargo (Rust toolchain)
#   - psql client (optional, for post-test verification)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
SERVICES_DIR="${PROJECT_DIR}/src/services"

# ── Configuration ──────────────────────────────────────────────────────────

NEO4J_IMAGE="${NEO4J_IMAGE:-neo4j:5.15-community}"
NEO4J_PORT="${NEO4J_PORT:-7687}"
NEO4J_HTTP_PORT="${NEO4J_HTTP_PORT:-7474}"
NEO4J_USER="${NEO4J_USER:-neo4j}"
NEO4J_PASSWORD="${NEO4J_PASSWORD:-password}"

PG_IMAGE="${PG_IMAGE:-postgres:16-alpine}"
PG_PORT="${PG_PORT:-5432}"
PG_USER="${PG_USER:-postgres}"
PG_PASSWORD="${PG_PASSWORD:-password}"
PG_DB="${PG_DB:-vedo_versioning}"

# Test environment variables injected into cargo test
export NEO4J_TEST_URI="${NEO4J_URI:-bolt://localhost:${NEO4J_PORT}}"
export NEO4J_USER="${NEO4J_USER}"
export NEO4J_PASSWORD="${NEO4J_PASSWORD}"
export PG_TEST_DATABASE_URL="${PG_URL:-postgres://${PG_USER}:${PG_PASSWORD}@localhost:${PG_PORT}/${PG_DB}}"

export RUST_LOG="${RUST_LOG:-info}"

CONTAINER_PREFIX="vedo-inttest"

NEO4J_CONTAINER="${CONTAINER_PREFIX}-neo4j"
PG_CONTAINER="${CONTAINER_PREFIX}-postgres"

# ── Docker helpers ─────────────────────────────────────────────────────────

cleanup() {
    echo "=== Cleaning up test containers ==="
    docker rm -f "${NEO4J_CONTAINER}" 2>/dev/null || true
    docker rm -f "${PG_CONTAINER}" 2>/dev/null || true
}

ensure_network() {
    # Use the default bridge network — no custom network needed for
    # service-to-database connectivity when running on localhost.
    :
}

start_neo4j() {
    echo "=== Starting Neo4j ==="
    docker run --rm -d \
        --name "${NEO4J_CONTAINER}" \
        -p "${NEO4J_PORT}:7687" \
        -p "${NEO4J_HTTP_PORT}:7474" \
        -e "NEO4J_AUTH=${NEO4J_USER}/${NEO4J_PASSWORD}" \
        "${NEO4J_IMAGE}"

    echo "Waiting for Neo4j to become ready..."
    local max_attempts=30
    local attempt=0
    until docker exec "${NEO4J_CONTAINER}" cypher-shell \
        -u "${NEO4J_USER}" -p "${NEO4J_PASSWORD}" \
        "RETURN 1" >/dev/null 2>&1; do
        attempt=$((attempt + 1))
        if [ "${attempt}" -ge "${max_attempts}" ]; then
            echo "ERROR: Neo4j did not become ready in time"
            docker logs "${NEO4J_CONTAINER}" --tail 20
            cleanup
            exit 1
        fi
        sleep 2
    done
    echo "Neo4j is ready."
}

start_postgres() {
    echo "=== Starting PostgreSQL ==="
    docker run --rm -d \
        --name "${PG_CONTAINER}" \
        -p "${PG_PORT}:5432" \
        -e "POSTGRES_USER=${PG_USER}" \
        -e "POSTGRES_PASSWORD=${PG_PASSWORD}" \
        -e "POSTGRES_DB=${PG_DB}" \
        "${PG_IMAGE}"

    echo "Waiting for PostgreSQL to become ready..."
    local max_attempts=30
    local attempt=0
    until docker exec "${PG_CONTAINER}" pg_isready -U "${PG_USER}" >/dev/null 2>&1; do
        attempt=$((attempt + 1))
        if [ "${attempt}" -ge "${max_attempts}" ]; then
            echo "ERROR: PostgreSQL did not become ready in time"
            docker logs "${PG_CONTAINER}" --tail 20
            cleanup
            exit 1
        fi
        sleep 2
    done
    echo "PostgreSQL is ready."
}

# ── Main ──────────────────────────────────────────────────────────────────

echo "========================================"
echo " VEDO Core — Rust Integration Tests"
echo "========================================"
echo "Neo4j URI:   ${NEO4J_TEST_URI}"
echo "PostgreSQL:  ${PG_TEST_DATABASE_URL}"
echo "RUST_LOG:    ${RUST_LOG}"
echo ""

if [ "${SKIP_DOCKER:-0}" != "1" ]; then
    # Register cleanup on exit (normal or error)
    trap cleanup EXIT

    cleanup 2>/dev/null || true
    start_neo4j
    start_postgres
else
    echo "SKIP_DOCKER=1 — using already-running containers"
fi

echo ""
echo "=== Running ontology-service tests ==="
cargo test \
    --manifest-path "${SERVICES_DIR}/Cargo.toml" \
    -p ontology-service \
    --lib \
    -- --nocapture \
    2>&1 | tail -20

echo ""
echo "=== Running ontology-service integration tests ==="
cargo test \
    --manifest-path "${SERVICES_DIR}/Cargo.toml" \
    -p ontology-service \
    --test query_integration_test \
    --test class_integration_test \
    --test property_integration_test \
    --test individual_integration_test \
    --test import_export_integration_test \
    --test route_registration_test \
    -- --nocapture \
    2>&1 | tail -20

echo ""
echo "=== Running versioning-service lib tests ==="
cargo test \
    --manifest-path "${SERVICES_DIR}/Cargo.toml" \
    -p versioning-service \
    --lib \
    -- --nocapture \
    2>&1 | tail -20

echo ""
echo "=== Running versioning-service integration tests ==="
cargo test \
    --manifest-path "${SERVICES_DIR}/Cargo.toml" \
    -p versioning-service \
    --test branch_integration_test \
    --test checkout_integration_test \
    --test commit_integration_test \
    --test merge_integration_test \
    --test transactional_integration_test \
    --test route_registration_test \
    -- --nocapture \
    2>&1 | tail -30

echo ""
echo "=== All Rust integration tests completed ==="
