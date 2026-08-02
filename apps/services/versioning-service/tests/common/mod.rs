//! Common test utilities for versioning-service integration tests.

#![allow(dead_code)]

use versioning_service::postgres::{create_pool, run_manual_migrations, PgConfig, PgPoolWrapper};

/// Guards schema migration so it runs exactly once per test binary.
/// Even with pg_advisory_lock in `run_manual_migrations`, running migrations
/// once avoids redundant DDL churn when every test calls `connect_test_pg`.
static MIGRATED: tokio::sync::OnceCell<()> = tokio::sync::OnceCell::const_new();

/// Returns `true` if PostgreSQL integration tests are enabled.
pub fn is_integration_enabled() -> bool {
    std::env::var("PG_TEST_DATABASE_URL").is_ok()
}

/// Requires PostgreSQL to be available — panics with a clear message if not.
///
/// Call this at the beginning of every integration test that needs PostgreSQL.
/// Unlike the old `skip_if_no_pg()` which silently returned `false`, this
/// function **fails loudly** so `cargo test` never silently skips tests.
pub fn require_pg() {
    if !is_integration_enabled() {
        panic!(
            "PG_TEST_DATABASE_URL is not set.\n\
             Set it to run integration tests, e.g.:\n\
             PG_TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/vedo_versioning_test\n\
             Or run via Makefile: make test-versioning"
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
    let pool = create_pool(&config)
        .await
        .expect("create PG pool for tests");
    // Apply the service's own schema so integration tests work against a
    // fresh test database (build_test_app does not run migrations).
    // Run once per binary — concurrent migration runs race on index DDL
    // (pg_advisory_lock in run_manual_migrations is the second line of defense).
    MIGRATED
        .get_or_init(|| async {
            run_manual_migrations(pool.pool())
                .await
                .expect("run versioning migrations for tests");
        })
        .await;
    pool
}

/// Builds a test app with a real PG connection.
pub fn build_test_app(pool: PgPoolWrapper) -> axum::Router {
    let state = std::sync::Arc::new(versioning_service::AppState {
        pg: Some(pool),
        branch_repo: None,
        commit_repo: None,
    });
    versioning_service::build_app(state)
}
