//! Common test utilities for versioning-service integration tests.

#![allow(dead_code)]

use versioning_service::postgres::{create_pool, PgConfig, PgPoolWrapper};

/// Returns `true` if PostgreSQL integration tests are enabled.
pub fn is_integration_enabled() -> bool {
    std::env::var("PG_TEST_DATABASE_URL").is_ok()
}

/// Panics with a clear message if PostgreSQL integration is not configured.
///
/// Call this at the beginning of every integration test. When `PG_TEST_DATABASE_URL`
/// is not set, the test panics so the developer knows the test was skipped,
/// rather than silently exiting the binary (which violates B4 of the TQS).
///
/// Use `PG_TEST_DATABASE_URL=postgres://... cargo test` to run PostgreSQL-backed
/// integration tests.
pub fn skip_if_no_pg() {
    if !is_integration_enabled() {
        panic!(
            "PostgreSQL integration tests require PG_TEST_DATABASE_URL environment variable. \
             Set it to run these tests, e.g.: \
             PG_TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/vedo_test cargo test --workspace"
        );
    }
}

/// Creates a PG pool from the test env var or panics.
pub async fn connect_test_pg() -> PgPoolWrapper {
    let url = std::env::var("PG_TEST_DATABASE_URL")
        .or_else(|_| std::env::var("DATABASE_URL"))
        .expect("PG_TEST_DATABASE_URL or DATABASE_URL must be set");
    let config = PgConfig {
        database_url: url,
        max_connections: 5,
    };
    create_pool(&config)
        .await
        .expect("create PG pool for tests")
}

/// Builds a test app with a real PG connection.
pub fn build_test_app(pool: PgPoolWrapper) -> axum::Router {
    let state = std::sync::Arc::new(versioning_service::AppState { pg: Some(pool) });
    versioning_service::build_app(state)
}
