use std::sync::Arc;

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;
use tracing::info;

use crate::models::SearchResult;

/// A cached published ontology entry served by the public browse API.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PublishedOntology {
    pub id: String,
    pub snapshot_id: String,
    pub name: String,
    pub description: Option<String>,
    pub class_count: i32,
    pub property_count: i32,
    pub individual_count: i32,
    pub published_at: DateTime<Utc>,
    pub format: String,
}

/// Summary of a published ontology (list view).
#[derive(Debug, Serialize)]
pub struct OntologySummary {
    pub id: String,
    pub name: String,
    pub description: Option<String>,
    pub class_count: i32,
    pub published_at: DateTime<Utc>,
}

/// Detailed published ontology with class hierarchy.
#[derive(Debug, Serialize)]
pub struct OntologyDetail {
    pub id: String,
    pub name: String,
    pub description: Option<String>,
    pub class_count: i32,
    pub property_count: i32,
    pub individual_count: i32,
    pub published_at: DateTime<Utc>,
    pub format: String,
    pub class_tree: Vec<ClassNode>,
}

/// A node in the class hierarchy tree.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClassNode {
    pub id: String,
    #[serde(rename = "label")]
    pub name: String,
    pub children: Vec<ClassNode>,
}

/// Search parameters for published entities.
#[derive(Debug, Deserialize)]
pub struct SearchQuery {
    pub q: String,
    pub max_results: Option<usize>,
}

/// SnapshotReader provides read-only access to published ontology snapshots.
///
/// For MVP, it stores published ontology metadata in-memory. In future iterations,
/// this will read from the publisher-service API or MinIO directly.
#[derive(Clone)]
pub struct SnapshotReader {
    ontologies: Arc<RwLock<Vec<PublishedOntology>>>,
}

impl SnapshotReader {
    /// Creates a new empty SnapshotReader.
    pub fn new() -> Self {
        SnapshotReader {
            ontologies: Arc::new(RwLock::new(Vec::new())),
        }
    }

    /// Lists all published ontologies.
    pub async fn list_ontologies(&self) -> Vec<PublishedOntology> {
        let guard = self.ontologies.read().await;
        guard.clone()
    }

    /// Gets a specific published ontology by ID.
    pub async fn get_ontology(&self, id: &str) -> Option<PublishedOntology> {
        let guard = self.ontologies.read().await;
        guard.iter().find(|o| o.id == id).cloned()
    }

    /// Publishes a new ontology (called when a snapshot is created).
    pub async fn add_ontology(&self, ontology: PublishedOntology) {
        info!(
            ontology_id = %ontology.id,
            name = %ontology.name,
            "Published ontology registered for public browse"
        );
        let mut guard = self.ontologies.write().await;
        guard.push(ontology);
    }

    /// Searches published entities by label/keyword.
    /// Performs simple substring matching against ontology name/description.
    pub async fn search(&self, query: &str, max_results: Option<usize>) -> Vec<SearchResult> {
        let max = max_results.unwrap_or(20);
        let guard = self.ontologies.read().await;
        let query_lower = query.to_lowercase();
        let mut results = Vec::new();

        for ontology in guard.iter() {
            // Search in ontology name
            if ontology.name.to_lowercase().contains(&query_lower) {
                results.push(SearchResult {
                    ontology_id: ontology.id.clone(),
                    entity_id: ontology.id.clone(),
                    entity_type: "ontology".to_string(),
                    label: ontology.name.clone(),
                    match_field: "name".to_string(),
                });
            }

            // Search in ontology description
            if let Some(desc) = &ontology.description {
                if desc.to_lowercase().contains(&query_lower) {
                    results.push(SearchResult {
                        ontology_id: ontology.id.clone(),
                        entity_id: ontology.id.clone(),
                        entity_type: "ontology".to_string(),
                        label: ontology.name.clone(),
                        match_field: "description".to_string(),
                    });
                }
            }

            if results.len() >= max {
                break;
            }
        }

        results.truncate(max);
        results
    }

    /// Fetches class tree for a published ontology.
    /// For MVP, returns the class tree from the snapshot metadata.
    /// In a full implementation, this materializes from the stored snapshot.
    pub async fn get_class_tree(&self, ontology_id: &str) -> Option<Vec<ClassNode>> {
        // MVP: Return a placeholder class tree with root + count hint
        let guard = self.ontologies.read().await;
        guard.iter().find(|o| o.id == ontology_id).map(|_| {
            vec![ClassNode {
                id: format!("root-{ontology_id}"),
                name: "owl:Thing".to_string(),
                children: vec![ClassNode {
                    id: format!("loading-{ontology_id}"),
                    name: format!(
                        "... ({}) classes",
                        guard
                            .iter()
                            .find(|o| o.id == ontology_id)
                            .map_or(0, |o| o.class_count)
                    ),
                    children: Vec::new(),
                }],
            }]
        })
    }
}

impl Default for SnapshotReader {
    fn default() -> Self {
        Self::new()
    }
}
