//! Axum handler for SHACL validation endpoints.
//!
//! Provides:
//! - `POST /api/v1/ontologies/{ontology_id}/validate` — validate ontology against SHACL shapes

use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::{header, StatusCode},
    response::{IntoResponse, Response},
    Json,
};
use serde::Deserialize;
use tracing::{debug, warn};

use crate::services::shacl_validator::ShaclValidator;
use crate::AppState;

/// Request body for SHACL validation.
#[derive(Debug, Deserialize)]
pub struct ValidateRequest {
    /// SHACL shapes graph in Turtle format.
    /// If omitted, built-in default shapes are used.
    #[serde(default)]
    pub shapes_turtle: Option<String>,
}

/// Handles `POST /api/v1/ontologies/{ontology_id}/validate`.
///
/// Validates the ontology against SHACL shapes. Accepts optional shapes
/// in Turtle format, or uses built-in default shapes when not provided.
pub async fn validate_ontology_handler(
    State(state): State<Arc<AppState>>,
    Path(ontology_id): Path<String>,
    Json(req): Json<ValidateRequest>,
) -> Response {
    debug!(ontology_id, "SHACL validation requested");

    let pool = match &state.neo4j {
        Some(pool) => pool.clone(),
        None => {
            warn!("Validation failed: Neo4j not configured");
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

    let validator = ShaclValidator::new(pool);

    let report = validator
        .validate(&ontology_id, req.shapes_turtle.as_deref())
        .await;

    let status = if report.conforms {
        StatusCode::OK
    } else {
        StatusCode::OK // Still 200 OK with validation results
    };

    debug!(
        ontology_id,
        conforms = report.conforms,
        result_count = report.results.len(),
        "Validation completed"
    );

    (
        status,
        [(header::CONTENT_TYPE, "application/json")],
        Json(report),
    )
        .into_response()
}
