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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_search_query_defaults() {
        let q: SearchQuery = serde_json::from_str(r#"{"q":"Person"}"#).unwrap();
        assert_eq!(q.q, "Person");
        assert_eq!(q.max_results, None);
    }

    #[test]
    fn test_search_query_with_max_results() {
        let q: SearchQuery = serde_json::from_str(r#"{"q":"Student","max_results":50}"#).unwrap();
        assert_eq!(q.q, "Student");
        assert_eq!(q.max_results, Some(50));
    }

    #[test]
    fn test_api_error_response_serialization() {
        let err = ApiErrorResponse {
            error: "NOT_FOUND".to_string(),
            message: "Entity not found".to_string(),
        };
        let json = serde_json::to_string(&err).unwrap();
        assert!(json.contains("NOT_FOUND"));
        assert!(json.contains("Entity not found"));
    }
}
