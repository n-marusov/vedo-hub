//! Axum handler for ontology import endpoints.
//!
//! Provides:
//! - `POST /api/v1/ontologies/{ontology_id}/import` — upload Turtle or RDF/XML

use std::sync::Arc;

use axum::{
    extract::{Path, Query, State},
    http::{header, StatusCode},
    response::IntoResponse,
    Json,
};
use serde::Deserialize;
use tracing::{debug, warn};

use crate::services::import_service::{ImportError, ImportService};
use crate::AppState;

/// Query parameters for the import endpoint.
#[derive(Debug, Deserialize)]
pub struct ImportParams {
    /// Import format: "turtle", "rdf-xml", or "owl-xml".
    /// If omitted, auto-detected from Content-Type header.
    #[serde(default)]
    pub format: Option<String>,
    /// Import strategy: "replace", "merge", or "version".
    /// - replace: delete existing ontology then import
    /// - merge: add new triples, skip existing (default)
    /// - version: import to a new branch
    #[serde(default = "default_strategy")]
    pub strategy: String,
}

fn default_strategy() -> String {
    "merge".to_string()
}

/// Handles `POST /api/v1/ontologies/{ontology_id}/import`.
///
/// Accepts Turtle (`text/turtle`) or RDF/XML (`application/rdf+xml`) content
/// in the request body. The format can be specified via the `format` query
/// parameter or auto-detected from the `Content-Type` header.
pub async fn import_ontology_handler(
    State(state): State<Arc<AppState>>,
    Path(ontology_id): Path<String>,
    Query(params): Query<ImportParams>,
    headers: axum::http::HeaderMap,
    body: String,
) -> impl IntoResponse {
    debug!(ontology_id, body_bytes = body.len(), "Import requested");

    let pool = match &state.neo4j {
        Some(pool) => pool.clone(),
        None => {
            warn!("Import failed: Neo4j not configured");
            return (
                StatusCode::SERVICE_UNAVAILABLE,
                [(header::CONTENT_TYPE, "application/json")],
                Json(serde_json::json!({
                    "error": "NEO4J_NOT_CONFIGURED",
                    "detail": "Neo4j database is not configured",
                })),
            );
        }
    };

    // Determine format: query param > Content-Type header
    let format = params
        .format
        .or_else(|| {
            headers
                .get(axum::http::header::CONTENT_TYPE)
                .and_then(|v| v.to_str().ok())
                .map(|ct| {
                    if ct.contains("turtle") {
                        "turtle"
                    } else if ct.contains("rdf+xml") {
                        "rdf-xml"
                    } else if ct.contains("owl+xml") {
                        "owl-xml"
                    } else {
                        // Fall back to turtle by default
                        "turtle"
                    }
                })
                .map(String::from)
        })
        .unwrap_or_else(|| "turtle".to_string());

    let service = ImportService::new(pool);

    debug!(
        ontology_id,
        format = %format,
        strategy = %params.strategy,
        "Import with strategy"
    );

    let result = match (format.as_str(), params.strategy.as_str()) {
        ("turtle", _) => {
            service
                .import_with_strategy(&ontology_id, &body, "turtle", &params.strategy)
                .await
        }
        ("rdf-xml", _) => {
            service
                .import_with_strategy(&ontology_id, &body, "rdf-xml", &params.strategy)
                .await
        }
        ("owl-xml", _) => {
            service
                .import_with_strategy(&ontology_id, &body, "owl-xml", &params.strategy)
                .await
        }
        (other, _) => {
            warn!(format = %other, "Unsupported import format");
            return (
                StatusCode::BAD_REQUEST,
                [(header::CONTENT_TYPE, "application/json")],
                Json(serde_json::json!({
                    "error": "UNSUPPORTED_FORMAT",
                    "detail": format!(
                        "Unsupported import format: {other}. Supported values: turtle, rdf-xml, owl-xml"
                    ),
                })),
            );
        }
    };

    match result {
        Ok(report) => {
            debug!(
                ontology_id,
                created = report.entities_created,
                skipped = report.entities_skipped,
                errors = report.errors.len(),
                "Import completed successfully"
            );

            let has_errors = !report.errors.is_empty();
            let status = if has_errors {
                StatusCode::OK // Partial success
            } else {
                StatusCode::OK
            };

            // Return report as JSON
            (
                status,
                [(header::CONTENT_TYPE, "application/json")],
                Json(serde_json::json!({
                    "status": if has_errors { "partial" } else { "success" },
                    "entities_created": report.entities_created,
                    "entities_updated": report.entities_updated,
                    "entities_skipped": report.entities_skipped,
                    "triple_count": report.triple_count,
                    "warnings": report.warnings,
                    "errors": report.errors,
                })),
            )
        }
        Err(e) => {
            let (status, code) = match &e {
                ImportError::ParseError(_) => (StatusCode::BAD_REQUEST, "PARSE_ERROR"),
                ImportError::Database(_) => (StatusCode::INTERNAL_SERVER_ERROR, "DATABASE_ERROR"),
            };

            warn!(
                ontology_id,
                error = %e,
                "Import failed"
            );

            (
                status,
                [(header::CONTENT_TYPE, "application/json")],
                Json(serde_json::json!({
                    "error": code,
                    "detail": e.to_string(),
                })),
            )
        }
    }
}
