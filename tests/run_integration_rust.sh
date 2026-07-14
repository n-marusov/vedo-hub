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
set -euo pipefail

cd "$(dirname "$0")/../src/services"

echo "=== Running ontology-service integration tests ==="
cargo test -p ontology-service --test '*' -- --test-threads=1 --nocapture

echo "=== Running versioning-service integration tests ==="
cargo test -p versioning-service --test '*' -- --test-threads=1 --nocapture

echo "=== Rust integration tests complete ==="
