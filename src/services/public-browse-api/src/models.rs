use serde::{Deserialize, Serialize};

use crate::snapshot_reader::ClassNode;

/// Summary of a published ontology (list view).
#[derive(Debug, Serialize)]
pub struct OntologySummary {
    pub id: String,
    pub name: String,
    pub description: Option<String>,
    pub class_count: i32,
    pub published_at: chrono::DateTime<chrono::Utc>,
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
    pub published_at: chrono::DateTime<chrono::Utc>,
    pub format: String,
    pub class_tree: Vec<ClassNode>,
}

/// Search parameters for published entities.
#[derive(Debug, Deserialize)]
pub struct SearchQuery {
    pub q: String,
    #[serde(default)]
    pub max_results: Option<usize>,
}

/// Search result entry.
#[derive(Debug, Serialize)]
pub struct SearchResult {
    pub ontology_id: String,
    pub entity_id: String,
    pub entity_type: String,
    pub label: String,
    pub match_field: String,
}

/// Error response for API errors.
#[derive(Debug, Serialize)]
pub struct ApiErrorResponse {
    pub error: String,
    pub message: String,
}
