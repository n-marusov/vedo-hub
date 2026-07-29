//! Common test utilities for ontology-service integration tests.
//!
//! Provides:
//! - `get_neo4j_config()` — reads test Neo4j configuration from environment
//! - `connect_to_neo4j()` — creates a Neo4j pool for test verification
//! - `create_test_app()` — builds an app with a real Neo4j connection
//! - `clean_ontology()` — removes test data after a test run
//! - `require_neo4j()` — ensures Neo4j is configured or panics

#![allow(dead_code)]

use std::sync::Arc;

use neo4rs;
use ontology_service::clients::auth_client::AuthClient;
use ontology_service::neo4j::{self, Neo4jPool};
use ontology_service::{build_app, AppState};

/// Executes a Cypher query and consumes the RowStream to ensure the
/// transaction is committed. `neo4rs 0.7` requires consuming the stream
/// for writes to persist — dropped streams may roll back.
pub async fn execute_query(pool: &Neo4jPool, query: neo4rs::Query) {
    let mut stream = pool
        .graph()
        .execute(query)
        .await
        .expect("Query execution failed");
    while let Ok(Some(_)) = stream.next().await {}
}

/// Initialises tracing for integration tests so `tracing::error!` calls
/// (e.g. database error details in `IntoResponse`) appear on stderr.
fn init_tracing() {
    let _ = tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "error".into()),
        )
        .with_test_writer()
        .try_init();
}

/// Returns the Neo4j test URI from the environment, or panics.
///
/// Integration tests MUST be run with `NEO4J_TEST_URI` set.
/// The Makefile target `test-integration-rust` sets this automatically
/// by starting Neo4j via Docker Compose and exporting the URI.
///
/// Call this at the beginning of every integration test that needs Neo4j.
pub fn require_neo4j() -> String {
    std::env::var("NEO4J_TEST_URI").expect(
        "NEO4J_TEST_URI is not set. Integration tests require a running Neo4j instance.\n\
         Run via: make test-integration-rust\n\
         Or set manually: NEO4J_TEST_URI=bolt://localhost:7687 cargo test\n\
         The Makefile auto-starts Neo4j via Docker Compose if not already running.",
    )
}

/// Gets Neo4j config from environment.
///
/// Requires `NEO4J_TEST_URI` or `NEO4J_URI` to be set.
/// Panics if neither is available — integration tests require a running Neo4j instance.
pub fn get_neo4j_config() -> neo4j::Neo4jConfig {
    let uri = std::env::var("NEO4J_TEST_URI")
        .or_else(|_| std::env::var("NEO4J_URI"))
        .expect(
            "NEO4J_TEST_URI or NEO4J_URI must be set. Integration tests require Neo4j.\n\
             Run via: make test-integration-rust",
        );

    let user = std::env::var("NEO4J_USER").unwrap_or_else(|_| "neo4j".to_string());
    let password = std::env::var("NEO4J_PASSWORD").unwrap_or_else(|_| "password".to_string());
    let max_connections = std::env::var("NEO4J_MAX_CONNECTIONS")
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(5);

    neo4j::Neo4jConfig {
        uri,
        user,
        password,
        max_connections,
    }
}

/// Creates a Neo4j connection pool for test verification.
/// Panics if connection fails (caller should set NEO4J_TEST_URI first).
pub async fn connect_to_neo4j() -> Neo4jPool {
    let config = get_neo4j_config();
    neo4j::create_pool(&config)
        .await
        .expect("Failed to connect to Neo4j for integration tests")
}

/// Creates an axum Router with a real Neo4j connection.
/// This allows tests to send HTTP requests to the service and verify
/// the state directly in Neo4j.
pub async fn create_test_app() -> (axum::Router, Neo4jPool) {
    init_tracing();
    let pool = connect_to_neo4j().await;
    let state = Arc::new(AppState {
        neo4j: Some(pool.clone()),
        auth_client: AuthClient::new(None),
    });
    let app = build_app(state);
    (app, pool)
}

/// Creates a unique test ontology ID for isolation.
pub fn test_ontology_id(test_name: &str) -> String {
    let ts = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap()
        .as_millis();
    format!("test-{test_name}-{ts}")
}

/// Cleans up all test data for a given ontology ID.
/// Runs a Cypher DELETE on all nodes with the test prefix.
pub async fn clean_ontology(pool: &Neo4jPool, ontology_id: &str) {
    let mut stream = pool
        .graph()
        .execute(
            neo4rs::query("MATCH (n) WHERE n.ontology_id = $id DETACH DELETE n")
                .param("id", ontology_id.to_string()),
        )
        .await
        .expect("cleanup query should succeed");
    while let Ok(Some(_)) = stream.next().await {}
}
