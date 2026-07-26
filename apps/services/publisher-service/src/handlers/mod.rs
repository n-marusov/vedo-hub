use std::sync::Arc;

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    Json,
};
use serde::Deserialize;
use tracing::{info, warn};

use crate::models::{
    ApiErrorResponse, PublishRequest, PublishResponse, SnapshotListResponse, SnapshotSummary,
};
use crate::AppState;

// [FIX] Allowed formats for published snapshots. Any other value must be
// rejected with 400 to prevent format strings leaking into filesystem paths
// (e.g. `format = "../../etc/passwd"`).
const ALLOWED_FORMATS: &[&str] = &["turtle", "rdf/xml", "owl", "n-triples", "jsonld"];

/// [FIX] Validates `ontology_id` against `^[A-Za-z0-9_-]+$` and returns a
/// `(StatusCode, Json<ApiErrorResponse>)` tuple on mismatch. This blocks path
/// traversal attempts like `../../etc/passwd` from reaching the storage layer
/// (which builds a filesystem path from the `ontology_id`) or the upstream
/// ontology-service URL path.
fn validate_ontology_id(ontology_id: &str) -> Result<(), (StatusCode, Json<ApiErrorResponse>)> {
    let valid = !ontology_id.is_empty()
        && ontology_id
            .chars()
            .all(|c| c.is_ascii_alphanumeric() || c == '_' || c == '-');
    if !valid {
        warn!(
            ontology_id = %ontology_id,
            "[FIX] Rejecting ontology_id with invalid characters"
        );
        return Err((
            StatusCode::BAD_REQUEST,
            Json(ApiErrorResponse {
                error: "INVALID_ONTOLOGY_ID".to_string(),
                message: "ontology_id must match ^[A-Za-z0-9_-]+$".to_string(),
            }),
        ));
    }
    Ok(())
}

/// [FIX] Validates the requested export format against a known allowlist.
/// Returns `(StatusCode, Json<ApiErrorResponse>)` on mismatch.
fn validate_format(format: &str) -> Result<(), (StatusCode, Json<ApiErrorResponse>)> {
    if !ALLOWED_FORMATS.contains(&format) {
        warn!(
            format = %format,
            "[FIX] Rejecting unknown export format"
        );
        return Err((
            StatusCode::BAD_REQUEST,
            Json(ApiErrorResponse {
                error: "UNSUPPORTED_FORMAT".to_string(),
                message: format!("format must be one of: {}", ALLOWED_FORMATS.join(", ")),
            }),
        ));
    }
    Ok(())
}

/// Query parameters for listing snapshots.
#[derive(Debug, Deserialize)]
pub struct ListSnapshotsQuery {
    pub branch_id: Option<String>,
    pub commit_id: Option<String>,
}

/// `POST /api/v1/ontologies/{ontology_id}/publish`
///
/// Creates a published snapshot of an ontology. The payload is the serialized
/// ontology in the requested format (default: turtle).
pub async fn publish_snapshot_handler(
    State(state): State<Arc<AppState>>,
    Path(ontology_id): Path<String>,
    Json(req): Json<PublishRequest>,
) -> Result<Json<PublishResponse>, (StatusCode, Json<ApiErrorResponse>)> {
    info!(
        ontology_id = %ontology_id,
        branch = ?req.branch_id,
        commit = ?req.commit_id,
        "Publish snapshot requested"
    );

    // [FIX] Validate ontology_id against path-traversal patterns before the
    // storage layer constructs a filesystem path from it.
    validate_ontology_id(&ontology_id)?;

    // Fetch ontology data from the ontology-service
    let format = req.format.unwrap_or_else(|| "turtle".to_string());
    // [FIX] Whitelist the format so arbitrary strings cannot leak into the
    // snapshot storage path extension.
    validate_format(&format)?;
    let ontology_payload = fetch_ontology_data(
        &state,
        &ontology_id,
        req.branch_id.as_deref(),
        req.commit_id.as_deref(),
        &format,
    )
    .await
    .map_err(|e| {
        warn!(error = %e, "Failed to fetch ontology data");
        (
            StatusCode::BAD_GATEWAY,
            Json(ApiErrorResponse {
                error: "ONTOLOGY_FETCH_FAILED".to_string(),
                message: format!("Failed to fetch ontology data from ontology-service: {e}"),
            }),
        )
    })?;

    // Create the snapshot record
    let snapshot = state
        .store
        .create_snapshot(
            &ontology_id,
            req.branch_id.as_deref(),
            req.commit_id.as_deref(),
            &ontology_payload,
            &format,
        )
        .await;

    info!(
        snapshot_id = %snapshot.id,
        ontology_id = %ontology_id,
        size_bytes = snapshot.size_bytes,
        "Snapshot published successfully"
    );

    Ok(Json(PublishResponse {
        id: snapshot.id.clone(),
        ontology_id: snapshot.ontology_id.clone(),
        status: snapshot.status,
        created_at: snapshot.created_at,
        storage_path: snapshot.storage_path.clone(),
    }))
}

/// `GET /api/v1/ontologies/{ontology_id}/snapshots`
///
/// Lists all snapshots for a given ontology.
pub async fn list_snapshots_handler(
    State(state): State<Arc<AppState>>,
    Path(ontology_id): Path<String>,
    Query(_query): Query<ListSnapshotsQuery>,
) -> Result<Json<SnapshotListResponse>, (StatusCode, Json<ApiErrorResponse>)> {
    info!(ontology_id = %ontology_id, "Listing snapshots");
    // [FIX] Validate ontology_id on this handler too — it routes to
    // store.list_snapshots which builds a per-ontology lookup.
    validate_ontology_id(&ontology_id)?;

    let snapshots = state.store.list_snapshots(&ontology_id).await;
    let total = snapshots.len();

    let summaries: Vec<SnapshotSummary> = snapshots.into_iter().map(Into::into).collect();

    Ok(Json(SnapshotListResponse {
        snapshots: summaries,
        total,
    }))
}

/// `GET /api/v1/snapshots/{snapshot_id}`
///
/// Retrieves details of a specific snapshot.
pub async fn get_snapshot_handler(
    State(state): State<Arc<AppState>>,
    Path(snapshot_id): Path<String>,
) -> Result<Json<SnapshotSummary>, (StatusCode, Json<ApiErrorResponse>)> {
    info!(snapshot_id = %snapshot_id, "Getting snapshot details");

    match state.store.get_snapshot(&snapshot_id).await {
        Some(snapshot) => Ok(Json(snapshot.into())),
        None => Err((
            StatusCode::NOT_FOUND,
            Json(ApiErrorResponse {
                error: "SNAPSHOT_NOT_FOUND".to_string(),
                message: format!("Snapshot not found: {snapshot_id}"),
            }),
        )),
    }
}

/// `DELETE /api/v1/snapshots/{snapshot_id}`
///
/// Retires (soft-deletes) a snapshot.
pub async fn retire_snapshot_handler(
    State(state): State<Arc<AppState>>,
    Path(snapshot_id): Path<String>,
) -> Result<StatusCode, (StatusCode, Json<ApiErrorResponse>)> {
    info!(snapshot_id = %snapshot_id, "Retiring snapshot");

    if state.store.retire_snapshot(&snapshot_id).await {
        Ok(StatusCode::NO_CONTENT)
    } else {
        Err((
            StatusCode::NOT_FOUND,
            Json(ApiErrorResponse {
                error: "SNAPSHOT_NOT_FOUND".to_string(),
                message: format!("Snapshot not found: {snapshot_id}"),
            }),
        ))
    }
}

/// Calls the ontology-service to fetch materialized ontology data.
async fn fetch_ontology_data(
    state: &AppState,
    ontology_id: &str,
    branch_id: Option<&str>,
    commit_id: Option<&str>,
    format: &str,
) -> Result<Vec<u8>, String> {
    let mut url = format!(
        "{}/api/v1/ontologies/{}/export",
        state.ontology_service_url, ontology_id
    );

    let mut params = Vec::new();
    params.push(format!("format={}", urlencoding(format)));
    if let Some(branch) = branch_id {
        params.push(format!("branch_id={}", urlencoding(branch)));
    }
    if let Some(commit) = commit_id {
        params.push(format!("commit_id={}", urlencoding(commit)));
    }
    if !params.is_empty() {
        url.push('?');
        url.push_str(&params.join("&"));
    }

    info!(
        url = %url,
        ontology_id = %ontology_id,
        format = %format,
        "Fetching ontology data from ontology-service"
    );

    let response = state
        .http_client
        .get(&url)
        .header("X-Trace-Id", format!("publish-{ontology_id}"))
        .send()
        .await
        .map_err(|e| format!("HTTP request failed: {e}"))?;

    if !response.status().is_success() {
        let status = response.status();
        let body = response.text().await.unwrap_or_default();
        return Err(format!("ontology-service returned {status}: {body}"));
    }

    let bytes = response
        .bytes()
        .await
        .map_err(|e| format!("Failed to read response body: {e}"))?;
    Ok(bytes.to_vec())
}

/// Simple URL encoding (replaces spaces with %20 and other reserved chars).
fn urlencoding(s: &str) -> String {
    s.chars()
        .map(|c| match c {
            ' ' => "%20".to_string(),
            '/' => "%2F".to_string(),
            '?' => "%3F".to_string(),
            '&' => "%26".to_string(),
            '=' => "%3D".to_string(),
            '#' => "%23".to_string(),
            '%' => "%25".to_string(),
            _ => c.to_string(),
        })
        .collect()
}
