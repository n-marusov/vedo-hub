#!/usr/bin/env bash
# Run all Rust integration tests that require a live database.
#
# Usage:
#   tests/run_integration_rust.sh                 # uses default env vars
#   NEO4J_TEST_URI=bolt://... PG_TEST_DATABASE_URL=... tests/run_integration_rust.sh
#
# These tests are normally SKIPPED by `cargo test --workspace` (the test
# binaries exit 0 when the relevant env var is unset). This script sets them
# up to actually run against a developer database.
#
# This script is a thin wrapper that delegates to the canonical script at
# src/services/scripts/run_integration_rust.sh. Keep the canonical version
# in sync with the CI pipeline.
set -euo pipefail

# Delegate to the canonical script under src/services/ scripts/
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CANONICAL="${SCRIPT_DIR}/../src/services/scripts/run_integration_rust.sh"

if [ -f "$CANONICAL" ]; then
    exec "$CANONICAL" "$@"
fi

# Fallback: run integration tests directly with required env vars.
# These defaults MUST be overridden in CI via environment variables.
cd "${SCRIPT_DIR}/../src/services"

: "${NEO4J_TEST_URI:=bolt://localhost:7687}"
: "${NEO4J_TEST_USER:=neo4j}"
: "${NEO4J_TEST_PASSWORD:=CHANGE_ME}"
: "${PG_TEST_DATABASE_URL:=postgres://vedo:CHANGE_ME@localhost:5432/vedo_versioning_test}"

export NEO4J_URI="$NEO4J_TEST_URI"
export NEO4J_USER="$NEO4J_TEST_USER"
export NEO4J_PASSWORD="$NEO4J_TEST_PASSWORD"
export DATABASE_URL="$PG_TEST_DATABASE_URL"

echo "=== Running ontology-service integration tests ==="
cargo test -p ontology-service --test '*' -- --test-threads=1 --nocapture

echo "=== Running versioning-service integration tests ==="
cargo test -p versioning-service --test '*' -- --test-threads=1 --nocapture

echo "=== Rust integration tests complete ==="
