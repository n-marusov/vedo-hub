//! Axum handler for ontology export endpoints.
//!
//! Provides:
//! - `GET /api/v1/ontologies/{ontology_id}/export?format=turtle|rdf-xml`

use std::sync::Arc;

use axum::{
    extract::{Path, Query, State},
    http::{header, StatusCode},
    response::{IntoResponse, Response},
    Json,
};
use serde::Deserialize;
use tracing::{debug, warn};

use crate::services::export_service::{ExportError, ExportService};
use crate::AppState;

/// Query parameters for the export endpoint.
#[derive(Debug, Deserialize)]
pub struct ExportParams {
    /// Export format: "turtle" or "rdf-xml" (default: "turtle").
    #[serde(default = "default_format")]
    pub format: String,
}

fn default_format() -> String {
    "turtle".to_string()
}

/// Handles `GET /api/v1/ontologies/{ontology_id}/export?format=turtle|rdf-xml`.
pub async fn export_ontology_handler(
    State(state): State<Arc<AppState>>,
    Path(ontology_id): Path<String>,
    Query(params): Query<ExportParams>,
) -> Response {
    debug!(ontology_id, format = %params.format, "Export requested");

    let pool = match &state.neo4j {
        Some(pool) => pool.clone(),
        None => {
            warn!("Export failed: Neo4j not configured");
            return (
                StatusCode::SERVICE_UNAVAILABLE,
                [(header::CONTENT_TYPE, "application/json")],
                Json(serde_json::json!({
                    "error": "NEO4J_NOT_CONFIGURED",
                    "detail": "Neo4j database is not configured",
                })),
            )
                .into_response();
        }
    };

    let service = ExportService::new(pool);

    let result = match params.format.as_str() {
        "turtle" => service.export_turtle(&ontology_id).await,
        "rdf-xml" => service.export_rdf_xml(&ontology_id).await,
        other => {
            warn!(format = %other, "Unsupported export format");
            return (
                StatusCode::BAD_REQUEST,
                [(header::CONTENT_TYPE, "application/json")],
                Json(serde_json::json!({
                    "error": "UNSUPPORTED_FORMAT",
                    "detail": format!("Unsupported export format: {other}. Supported values: turtle, rdf-xml"),
                })),
            )
                .into_response();
        }
    };

    match result {
        Ok(content) => {
            let content_type = if params.format == "turtle" {
                "text/turtle"
            } else {
                "application/rdf+xml"
            };

            debug!(
                ontology_id,
                format = %params.format,
                bytes = content.len(),
                "Export successful"
            );

            (
                StatusCode::OK,
                [(header::CONTENT_TYPE, content_type)],
                content,
            )
                .into_response()
        }
        Err(e) => {
            let (status, code) = match &e {
                ExportError::Database(_) => (StatusCode::INTERNAL_SERVER_ERROR, "DATABASE_ERROR"),
                ExportError::Serialization(_) => {
                    (StatusCode::INTERNAL_SERVER_ERROR, "SERIALIZATION_ERROR")
                }
                ExportError::Neo4jNotConfigured => {
                    (StatusCode::SERVICE_UNAVAILABLE, "NEO4J_NOT_CONFIGURED")
                }
                ExportError::UnsupportedFormat(_) => {
                    (StatusCode::BAD_REQUEST, "UNSUPPORTED_FORMAT")
                }
            };

            warn!(
                ontology_id,
                format = %params.format,
                error = %e,
                "Export failed"
            );

            (
                status,
                [(header::CONTENT_TYPE, "application/json")],
                Json(serde_json::json!({
                    "error": code,
                    "detail": e.to_string(),
                })),
            )
                .into_response()
        }
    }
}
