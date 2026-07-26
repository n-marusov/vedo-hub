//! Common test utilities for ontology-service integration tests.
//!
//! Provides:
//! - `get_neo4j_config()` — reads test Neo4j configuration from environment
//! - `connect_to_neo4j()` — creates a Neo4j pool for test verification
//! - `create_test_app()` — builds an app with a real Neo4j connection
//! - `clean_ontology()` — removes test data after a test run
//! - `skip_if_no_neo4j()` — helper to skip tests when Neo4j isn't available

#![allow(dead_code)]

use std::sync::Arc;

use neo4rs;
use ontology_service::clients::auth_client::AuthClient;
use ontology_service::neo4j::{self, Neo4jPool};
use ontology_service::{build_app, AppState};

/// Returns `true` if integration tests are enabled (NEO4J_TEST_URI is set).
pub fn is_integration_enabled() -> bool {
    std::env::var("NEO4J_TEST_URI").is_ok()
}

/// Returns `false` and prints a warning if Neo4j integration is not configured.
///
/// Call this at the beginning of every integration test. When `NEO4J_TEST_URI`
/// is not set, the function prints a clear message to stderr and returns `false`
/// so the caller can skip the test gracefully (TQS B4: stderr ≠ silent exit).
///
/// Use `NEO4J_TEST_URI=bolt://localhost:7687 cargo test` (or your actual URI)
/// to run Neo4j-backed integration tests.
pub fn skip_if_no_neo4j() -> bool {
    if !is_integration_enabled() {
        eprintln!(
            "⚠️  Skipping Neo4j integration test. \
             Set NEO4J_TEST_URI to run, e.g.: \
             NEO4J_TEST_URI=bolt://localhost:7687"
        );
        return false;
    }
    true
}

/// Gets Neo4j config from environment, falling back to defaults.
/// Uses NEO4J_TEST_URI when set, otherwise NEO4J_URI.
pub fn get_neo4j_config() -> neo4j::Neo4jConfig {
    let uri = std::env::var("NEO4J_TEST_URI")
        .or_else(|_| std::env::var("NEO4J_URI"))
        .unwrap_or_else(|_| "bolt://localhost:7687".to_string());

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
/// Panics if connection fails (caller should check `is_integration_enabled` first).
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
    let result = pool
        .graph()
        .execute(
            neo4rs::query("MATCH (n) WHERE n.ontology_id = $id DETACH DELETE n")
                .param("id", ontology_id.to_string()),
        )
        .await;
    if let Err(e) = result {
        eprintln!("Warning: cleanup failed for {ontology_id}: {e}");
    }
}
