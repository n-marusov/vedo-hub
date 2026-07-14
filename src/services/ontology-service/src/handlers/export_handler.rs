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
    /// Export format: "turtle", "rdf-xml", or "owl-xml" (default: "turtle").
    #[serde(default = "default_format")]
    pub format: String,
    /// Optional branch ID for versioned export.
    /// When specified, exports the materialized state at that branch.
    #[serde(default)]
    pub branch_id: Option<String>,
    /// Optional commit ID for versioned export.
    /// When specified, exports the materialized state at that commit.
    /// Takes precedence over branch_id when both are provided.
    #[serde(default)]
    pub commit_id: Option<String>,
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

    // Check if version-aware export is requested
    let has_version_params = params.branch_id.is_some() || params.commit_id.is_some();

    let result = if has_version_params {
        let versioning_url = std::env::var("VERSIONING_SERVICE_URL")
            .ok()
            .unwrap_or_else(|| "http://localhost:8083".to_string());
        let commit_id = params.commit_id.as_deref();
        let branch_id = params.branch_id.as_deref();

        match params.format.as_str() {
            "turtle" => {
                service
                    .export_turtle_versioned(&ontology_id, &versioning_url, branch_id, commit_id)
                    .await
            }
            "rdf-xml" => {
                service
                    .export_rdf_xml_versioned(&ontology_id, &versioning_url, branch_id, commit_id)
                    .await
            }
            "owl-xml" => {
                service
                    .export_owl_xml_versioned(&ontology_id, &versioning_url, branch_id, commit_id)
                    .await
            }
            other => {
                warn!(format = %other, "Unsupported export format");
                return (
                    StatusCode::BAD_REQUEST,
                    [(header::CONTENT_TYPE, "application/json")],
                    Json(serde_json::json!({
                        "error": "UNSUPPORTED_FORMAT",
                        "detail": format!("Unsupported export format: {other}. Supported values: turtle, rdf-xml, owl-xml"),
                    })),
                )
                    .into_response();
            }
        }
    } else {
        match params.format.as_str() {
            "turtle" => service.export_turtle(&ontology_id).await,
            "rdf-xml" => service.export_rdf_xml(&ontology_id).await,
            "owl-xml" => service.export_owl_xml(&ontology_id).await,
            other => {
                warn!(format = %other, "Unsupported export format");
                return (
                    StatusCode::BAD_REQUEST,
                    [(header::CONTENT_TYPE, "application/json")],
                    Json(serde_json::json!({
                        "error": "UNSUPPORTED_FORMAT",
                        "detail": format!("Unsupported export format: {other}. Supported values: turtle, rdf-xml, owl-xml"),
                    })),
                )
                    .into_response();
            }
        }
    };

    match result {
        Ok(content) => {
            let content_type = match params.format.as_str() {
                "turtle" => "text/turtle",
                "rdf-xml" => "application/rdf+xml",
                "owl-xml" => "application/owl+xml",
                _ => "text/plain",
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
                // [FIX] Versioning service unavailable — fail loudly instead
                // of silently falling back to stale state.
                ExportError::VersioningUnavailable(_) => {
                    (StatusCode::BAD_GATEWAY, "VERSIONING_UNAVAILABLE")
                }
            };

            warn!(
                ontology_id,
                format = %params.format,
                error = %e,
                "Export failed"
            );

            // [FIX] Sanitize versioning errors server-side — log full message,
            // return generic error code without exposing the internal URL
            // or transport error details.
            let detail = match &e {
                ExportError::VersioningUnavailable(msg) => {
                    warn!(
                        ontology_id,
                        raw_error = %msg,
                        code = %code,
                        "[FIX] Versioning service unavailable"
                    );
                    "Versioning service unavailable. The requested snapshot could not be materialized.".to_string()
                }
                ExportError::Database(msg) => {
                    warn!(
                        ontology_id,
                        raw_error = %msg,
                        code = %code,
                        "[FIX] Database error during export"
                    );
                    "Database error during export.".to_string()
                }
                _ => e.to_string(),
            };

            (
                status,
                [(header::CONTENT_TYPE, "application/json")],
                Json(serde_json::json!({
                    "error": code,
                    "detail": detail,
                })),
            )
                .into_response()
        }
    }
}
