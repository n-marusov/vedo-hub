#!/usr/bin/env bash
# Run all Rust integration tests that require a live database.
#
# Usage (infrastructure managed by Makefile — just set env vars):
#   NEO4J_TEST_URI=bolt://localhost:7687 tests/run_integration_rust.sh
#
# Infrastructure is expected to be available (started via `make test-integration-rust`
# or provided externally in CI). Tests will FAIL if Neo4j is not reachable.
#
# This script is a thin wrapper that delegates to the canonical script at
# apps/services/scripts/run_integration_rust.sh. Keep the canonical version
# in sync with the CI pipeline.
set -euo pipefail

# Delegate to the canonical script under apps/services/scripts/
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CANONICAL="${SCRIPT_DIR}/../apps/services/scripts/run_integration_rust.sh"

if [ -f "$CANONICAL" ]; then
    exec "$CANONICAL" "$@"
fi

# Ensure NEO4J_TEST_URI is set — fail early if not.
: "${NEO4J_TEST_URI:?NEO4J_TEST_URI is required — set it or run via 'make test-integration-rust'}"
: "${NEO4J_TEST_USER:=neo4j}"
: "${NEO4J_TEST_PASSWORD:=password}"

export NEO4J_URI="$NEO4J_TEST_URI"
export NEO4J_USER="$NEO4J_TEST_USER"
export NEO4J_PASSWORD="$NEO4J_TEST_PASSWORD"

echo "=== Running ontology-service integration tests ==="
cd "${SCRIPT_DIR}/../apps/services/ontology-service"
cargo test --test '*' -- --test-threads=1 --nocapture

echo "=== Running versioning-service integration tests ==="
cd "${SCRIPT_DIR}/../apps/services/versioning-service"
cargo test --test '*' -- --test-threads=1 --nocapture

echo "=== Rust integration tests complete ==="
