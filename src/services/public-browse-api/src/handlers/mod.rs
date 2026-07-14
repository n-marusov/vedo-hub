use std::sync::Arc;

use axum::{
    extract::{Path, Query, State},
    Json,
};
use tracing::info;

use crate::models::{ApiErrorResponse, OntologyDetail, OntologySummary, SearchQuery, SearchResult};
use crate::AppState;

/// GET /api/v1/ontologies
///
/// Lists all published ontologies (public, no auth required).
pub async fn list_published_ontologies(
    State(state): State<Arc<AppState>>,
) -> Json<Vec<OntologySummary>> {
    info!("Listing published ontologies");

    let ontologies = state.reader.list_ontologies().await;
    let summaries: Vec<OntologySummary> = ontologies
        .into_iter()
        .map(|o| OntologySummary {
            id: o.id,
            name: o.name,
            description: o.description,
            class_count: o.class_count,
            published_at: o.published_at,
        })
        .collect();

    Json(summaries)
}

/// GET /api/v1/ontologies/{ontology_id}
///
/// Returns full detail for a published ontology including class tree.
pub async fn get_published_ontology(
    State(state): State<Arc<AppState>>,
    Path(ontology_id): Path<String>,
) -> Result<Json<OntologyDetail>, (axum::http::StatusCode, Json<ApiErrorResponse>)> {
    info!(ontology_id = %ontology_id, "Getting published ontology detail");

    let ontology = state
        .reader
        .get_ontology(&ontology_id)
        .await
        .ok_or_else(|| {
            (
                axum::http::StatusCode::NOT_FOUND,
                Json(ApiErrorResponse {
                    error: "ONTOLOGY_NOT_FOUND".to_string(),
                    message: format!("Published ontology not found: {ontology_id}"),
                }),
            )
        })?;

    let class_tree = state
        .reader
        .get_class_tree(&ontology_id)
        .await
        .unwrap_or_default();

    Ok(Json(OntologyDetail {
        id: ontology.id,
        name: ontology.name,
        description: ontology.description,
        class_count: ontology.class_count,
        property_count: ontology.property_count,
        individual_count: ontology.individual_count,
        published_at: ontology.published_at,
        format: ontology.format,
        class_tree,
    }))
}

/// GET /api/v1/ontologies/{ontology_id}/class-tree
///
/// Returns the class hierarchy tree for a published ontology.
pub async fn get_class_tree(
    State(state): State<Arc<AppState>>,
    Path(ontology_id): Path<String>,
) -> Result<
    Json<Vec<crate::snapshot_reader::ClassNode>>,
    (axum::http::StatusCode, Json<ApiErrorResponse>),
> {
    info!(ontology_id = %ontology_id, "Getting class tree");

    let tree = state
        .reader
        .get_class_tree(&ontology_id)
        .await
        .ok_or_else(|| {
            (
                axum::http::StatusCode::NOT_FOUND,
                Json(ApiErrorResponse {
                    error: "ONTOLOGY_NOT_FOUND".to_string(),
                    message: format!("Published ontology not found: {ontology_id}"),
                }),
            )
        })?;

    Ok(Json(tree))
}

/// GET /api/v1/search
///
/// Searches published entities by keyword (public, no auth required).
/// Query param: q (search query), max_results (optional, default 20)
pub async fn search_published(
    State(state): State<Arc<AppState>>,
    Query(query): Query<SearchQuery>,
) -> Json<Vec<SearchResult>> {
    info!(query = %query.q, "Searching published ontologies");

    let results = state.reader.search(&query.q, query.max_results).await;

    Json(results)
}
