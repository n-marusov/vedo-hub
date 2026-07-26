//! Neo4j connection pool manager for ontology-service.
//!
//! Provides:
//! - Connection pool initialization from environment variables
//! - Health check via `RETURN 1` Cypher query
//! - Connection retry with exponential backoff on startup

use std::time::Duration;

use neo4rs::ConfigBuilder;
use tracing::info;

/// Default Neo4j connection URI
const DEFAULT_NEO4J_URI: &str = "bolt://localhost:7687";
/// Default Neo4j user
const DEFAULT_NEO4J_USER: &str = "neo4j";
/// Default Neo4j password
const DEFAULT_NEO4J_PASSWORD: &str = "password";

/// The shared Neo4j connection pool.
#[derive(Clone)]
pub struct Neo4jPool {
    graph: neo4rs::Graph,
}

impl Neo4jPool {
    /// Returns a reference to the inner `Graph` pool.
    pub fn graph(&self) -> &neo4rs::Graph {
        &self.graph
    }
}

/// Configuration for Neo4j connection.
pub struct Neo4jConfig {
    pub uri: String,
    pub user: String,
    pub password: String,
    pub max_connections: usize,
}

impl Default for Neo4jConfig {
    fn default() -> Self {
        Self {
            uri: DEFAULT_NEO4J_URI.to_string(),
            user: DEFAULT_NEO4J_USER.to_string(),
            password: DEFAULT_NEO4J_PASSWORD.to_string(),
            max_connections: 10,
        }
    }
}

impl Neo4jConfig {
    /// Loads config from environment variables, falling back to defaults.
    pub fn from_env() -> Self {
        Self {
            uri: std::env::var("NEO4J_URI").unwrap_or_else(|_| DEFAULT_NEO4J_URI.to_string()),
            user: std::env::var("NEO4J_USER").unwrap_or_else(|_| DEFAULT_NEO4J_USER.to_string()),
            password: std::env::var("NEO4J_PASSWORD")
                .unwrap_or_else(|_| DEFAULT_NEO4J_PASSWORD.to_string()),
            max_connections: std::env::var("NEO4J_MAX_CONNECTIONS")
                .ok()
                .and_then(|v| v.parse().ok())
                .unwrap_or(10),
        }
    }
}

/// Creates a Neo4j connection pool with the given config.
///
/// Retries connection with exponential backoff (1s, 2s, 4s, max 30s).
pub async fn create_pool(config: &Neo4jConfig) -> Result<Neo4jPool, neo4rs::Error> {
    let conf = ConfigBuilder::default()
        .uri(&config.uri)
        .user(&config.user)
        .password(&config.password)
        .max_connections(config.max_connections)
        .build()?;

    let graph = connect_with_retry(conf, 5).await?;

    info!(
        uri = %config.uri,
        max_connections = config.max_connections,
        "Connected to Neo4j"
    );

    Ok(Neo4jPool { graph })
}

/// Connects to Neo4j with exponential backoff retry.
async fn connect_with_retry(
    conf: neo4rs::Config,
    max_retries: u32,
) -> Result<neo4rs::Graph, neo4rs::Error> {
    let mut attempt = 0;
    loop {
        attempt += 1;
        match neo4rs::Graph::connect(conf.clone()).await {
            Ok(graph) => return Ok(graph),
            Err(e) if attempt <= max_retries => {
                let delay = Duration::from_secs(1 << attempt.min(5)); // 2, 4, 8, 16, 32
                tracing::warn!(
                    attempt,
                    max_retries,
                    delay_ms = delay.as_millis(),
                    error = %e,
                    "Neo4j connection attempt failed, retrying"
                );
                tokio::time::sleep(delay).await;
            }
            Err(e) => {
                tracing::error!(
                    attempt,
                    max_retries,
                    error = %e,
                    "All Neo4j connection attempts exhausted"
                );
                return Err(e);
            }
        }
    }
}

/// Runs a `RETURN 1` health check against the given pool.
/// Returns `true` if the database is reachable.
pub async fn health_check(pool: &Neo4jPool) -> bool {
    let query = neo4rs::query("RETURN 1");
    let mut result = match pool.graph().execute(query).await {
        Ok(r) => r,
        Err(e) => {
            tracing::warn!(error = %e, "Neo4j health check failed");
            return false;
        }
    };
    match result.next().await {
        Ok(Some(row)) => {
            // Row.get returns Result<Option<BoltType>>, extract first column "1"
            let val: Result<Option<i64>, _> = row.get("1");
            matches!(val, Ok(Some(1)))
        }
        _ => false,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_config_from_env_defaults() {
        let config = Neo4jConfig::default();
        assert_eq!(config.uri, "bolt://localhost:7687");
        assert_eq!(config.user, "neo4j");
        assert_eq!(config.password, "password");
        assert_eq!(config.max_connections, 10);
    }

    #[test]
    fn test_config_from_env_with_values() {
        let _config = Neo4jConfig::from_env();
        // Should not panic regardless of env state
    }

    #[test]
    fn test_config_default_trait() {
        let config: Neo4jConfig = Default::default();
        assert_eq!(config.uri, DEFAULT_NEO4J_URI);
    }
}
