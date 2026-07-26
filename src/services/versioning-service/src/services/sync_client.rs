//! HTTP client for syncing materialized state to the ontology service.
//!
//! After checkout or rollback, the materialized state is pushed to the
//! ontology service's Neo4j state endpoint so the ontology graph reflects
//! the correct version.

use reqwest::Client;
use serde::Serialize;

use crate::error::VersionError;
use crate::models::TripleRef;

/// Default ontology service URL.
const DEFAULT_ONTOLOGY_URL: &str = "http://localhost:8082";

/// Client for communicating with the ontology service.
pub struct SyncClient {
    base_url: String,
    client: Client,
}

impl SyncClient {
    /// Creates a new `SyncClient` with the given base URL.
    /// Falls back to the default if the URL is empty.
    pub fn new(base_url: Option<String>) -> Self {
        Self {
            base_url: base_url.unwrap_or_else(|| DEFAULT_ONTOLOGY_URL.to_string()),
            client: Client::new(),
        }
    }

    /// Pushes a materialized state (list of triples) to the ontology service
    /// for a specific ontology.
    ///
    /// This replaces the entire Neo4j state for that ontology with the
    /// provided triples.
    pub async fn push_state(
        &self,
        ontology_id: &str,
        triples: &[TripleRef],
    ) -> Result<(), VersionError> {
        tracing::debug!(
            ontology_id = %ontology_id,
            triple_count = triples.len(),
            "Pushing materialized state to ontology service"
        );

        let payload = StatePushRequest {
            ontology_id: ontology_id.to_string(),
            triples: triples.to_vec(),
        };

        let url = format!(
            "{}/api/v1/internal/ontologies/{}/state",
            self.base_url, ontology_id
        );

        match self
            .client
            .post(&url)
            .json(&payload)
            .timeout(std::time::Duration::from_secs(30))
            .send()
            .await
        {
            Ok(resp) => {
                if resp.status().is_success() {
                    tracing::info!(
                        ontology_id = %ontology_id,
                        triple_count = triples.len(),
                        "Materialized state synced to ontology service"
                    );
                    Ok(())
                } else {
                    let status = resp.status();
                    let body = resp.text().await.unwrap_or_default();
                    tracing::warn!(
                        ontology_id = %ontology_id,
                        status = %status,
                        body = %body,
                        "Ontology service rejected state push"
                    );
                    Err(VersionError::Database(format!(
                        "Ontology service returned {status}: {body}"
                    )))
                }
            }
            Err(e) => {
                tracing::warn!(
                    ontology_id = %ontology_id,
                    error = %e,
                    "[FIX] Transport error syncing state to ontology service — propagating"
                );
                Err(VersionError::SyncFailed(e.to_string()))
            }
        }
    }

    /// Checks if the ontology service is reachable.
    pub async fn health_check(&self) -> bool {
        let url = format!("{}/health", self.base_url);
        match self
            .client
            .get(&url)
            .timeout(std::time::Duration::from_secs(5))
            .send()
            .await
        {
            Ok(resp) => resp.status().is_success(),
            Err(_) => false,
        }
    }
}

/// Request payload for pushing state to the ontology service.
#[derive(Debug, Serialize)]
struct StatePushRequest {
    ontology_id: String,
    triples: Vec<TripleRef>,
}
