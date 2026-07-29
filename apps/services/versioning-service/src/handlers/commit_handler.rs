//! Axum HTTP handlers for commit operations.
//!
//! Provides handlers for creating commits, retrieving commit details,
//! and listing commit history with pagination.

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    Json,
};
use std::sync::Arc;
use tracing::{debug, info};
use uuid::Uuid;

use crate::error::VersionError;
use crate::models::{
    CommitDeltaPreview, CommitDetailResponse, CommitSummary, CreateCommitRequest,
    ListCommitsParams, PaginatedResponse,
};
use crate::repositories::{CommitRepository, CommitRepositoryTrait};
use crate::services::delta_service::DeltaReplayEngine;
use crate::services::semantic_diff::SemanticDiff;
use crate::services::state_service::StateService;
use crate::services::sync_client::SyncClient;
use crate::AppState;

/// Helper: extracts a `CommitRepository` from the application state or returns
/// a `PgNotConfigured` error when no database pool is available.
/// Prefers an injected mock repo when available.
fn repo_from_state(
    state: &AppState,
) -> Result<Arc<dyn CommitRepositoryTrait + Send + Sync>, VersionError> {
    // Use injected mock repo if available (for testing)
    if let Some(repo) = &state.commit_repo {
        return Ok(Arc::clone(repo));
    }
    match &state.pg {
        Some(pool) => Ok(Arc::new(CommitRepository::new(pool.pool().clone()))),
        None => Err(VersionError::PgNotConfigured),
    }
}

// ── Handlers ──────────────────────────────────────────────────────────────

/// POST /api/v1/versioning/commits — Create a new commit.
///
/// Accepts a commit request with `branch_id`, message, author info, and delta.
/// Returns the created commit with full details.
pub async fn create_commit_handler(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateCommitRequest>,
) -> Result<(StatusCode, Json<CommitDetailResponse>), VersionError> {
    debug!(
        branch_id = %req.branch_id,
        author_id = %req.author_id,
        "create_commit_handler called"
    );

    let repo = repo_from_state(&state)?;
    let commit = repo.create(&req).await?;

    // Best-effort state snapshot creation (every 50 commits)
    let pg = state.pg.as_ref().ok_or(VersionError::PgNotConfigured)?;
    let engine = DeltaReplayEngine::new(pg.pool().clone());
    engine.maybe_create_snapshot(commit.id).await;

    info!(
        commit_id = %commit.id,
        branch_id = %req.branch_id,
        message = %req.message,
        "Commit created successfully"
    );

    Ok((
        StatusCode::CREATED,
        Json(CommitDetailResponse::from(commit)),
    ))
}

/// GET /api/v1/versioning/commits/{id} — Get commit details.
///
/// Returns full commit details including a delta preview (first 10 triples
/// of each section).
pub async fn get_commit_handler(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> Result<Json<CommitDetailResponse>, VersionError> {
    debug!(commit_id = %id, "get_commit_handler called");

    let repo = repo_from_state(&state)?;
    let commit = repo.get_by_id(id).await?;

    info!(commit_id = %commit.id, "Commit details retrieved");
    Ok(Json(CommitDetailResponse::from(commit)))
}

/// GET /api/v1/versioning/commits — List commits with pagination.
///
/// Supports optional `branch_id` filter. Returns paginated commit summaries
/// (excludes full delta for performance).
pub async fn list_commits_handler(
    State(state): State<Arc<AppState>>,
    Query(params): Query<ListCommitsParams>,
) -> Result<Json<PaginatedResponse<CommitSummary>>, VersionError> {
    debug!(
        branch_id = ?params.branch_id,
        page = params.page,
        per_page = params.per_page,
        "list_commits_handler called"
    );

    let repo = repo_from_state(&state)?;
    let result = repo.list(&params).await?;

    info!(
        total = result.total,
        returned = result.items.len(),
        "Commits listed"
    );

    Ok(Json(result))
}

/// GET /api/v1/versioning/commits/{id}/delta — Get raw delta for a commit.
pub async fn get_commit_delta_handler(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> Result<Json<CommitDeltaPreview>, VersionError> {
    debug!(commit_id = %id, "get_commit_delta_handler called");

    let repo = repo_from_state(&state)?;
    let commit = repo.get_by_id(id).await?;

    // Return the full delta as a preview (no truncation needed for delta endpoint)
    let preview = CommitDeltaPreview {
        added_total: commit.delta.added_triples.len(),
        added_preview: commit.delta.added_triples,
        removed_total: commit.delta.removed_triples.len(),
        removed_preview: commit.delta.removed_triples,
        modified_total: commit.delta.modified_triples.len(),
        modified_preview: commit.delta.modified_triples,
    };

    info!(commit_id = %id, "Commit delta retrieved");
    Ok(Json(preview))
}

/// GET /api/v1/versioning/commits/{id}/semantic-diff — Get semantic diff for a commit.
///
/// Transforms the raw triple delta into entity-level diff entries
/// categorized by entity type (class, property, individual) and change
/// type (added, removed, modified). Useful for commit visualization.
pub async fn get_commit_semantic_diff_handler(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> Result<Json<SemanticDiff>, VersionError> {
    debug!(commit_id = %id, "get_commit_semantic_diff_handler called");

    let repo = repo_from_state(&state)?;
    let commit = repo.get_by_id(id).await?;

    let diff = crate::services::semantic_diff::compute_semantic_diff(&commit.delta);

    info!(
        commit_id = %id,
        total_entities = diff.entries.len(),
        "Semantic diff computed"
    );

    Ok(Json(diff))
}

/// Request body for checkout.
#[derive(serde::Deserialize)]
pub struct CheckoutRequest {
    pub branch_id: Uuid,
}

/// Request body for rollback.
#[derive(serde::Deserialize)]
pub struct RollbackRequest {
    pub branch_id: Uuid,
    pub author_id: String,
    pub author_name: String,
}

/// POST /api/v1/versioning/commits/{id}/checkout — Check out a commit.
///
/// Replays all deltas from root to the specified commit, producing a
/// materialized state. Optionally syncs to the ontology service.
pub async fn checkout_commit_handler(
    State(state): State<Arc<AppState>>,
    Path(commit_id): Path<Uuid>,
    Json(req): Json<CheckoutRequest>,
) -> Result<Json<serde_json::Value>, VersionError> {
    debug!(
        commit_id = %commit_id,
        branch_id = %req.branch_id,
        "checkout_commit_handler called"
    );

    let pg = state.pg.as_ref().ok_or(VersionError::PgNotConfigured)?;
    let ontology_url = std::env::var("ONTOLOGY_SERVICE_URL").ok();
    let sync_client = SyncClient::new(ontology_url);
    let state_service = StateService::new(pg.pool().clone(), Some(sync_client));

    let result = state_service
        .checkout(req.branch_id, Some(commit_id))
        .await?;

    info!(
        branch_id = %result.branch_id,
        commit_id = %result.commit_id,
        triple_count = result.triple_count,
        delta_count = result.delta_count,
        synced = result.synced,
        "Checkout completed"
    );

    Ok(Json(serde_json::json!({
        "status": "ok",
        "branch_id": result.branch_id,
        "commit_id": result.commit_id,
        "triple_count": result.triple_count,
        "delta_count": result.delta_count,
        "synced": result.synced,
    })))
}

/// POST /api/v1/versioning/commits/{id}/rollback — Roll back to a previous commit.
///
/// Creates an inverse delta commit that reverses all changes between the
/// current head and the specified target commit.
pub async fn rollback_commit_handler(
    State(state): State<Arc<AppState>>,
    Path(target_commit_id): Path<Uuid>,
    Json(req): Json<RollbackRequest>,
) -> Result<(StatusCode, Json<CommitDetailResponse>), VersionError> {
    debug!(
        target_commit_id = %target_commit_id,
        branch_id = %req.branch_id,
        "rollback_commit_handler called"
    );

    let pg = state.pg.as_ref().ok_or(VersionError::PgNotConfigured)?;
    let ontology_url = std::env::var("ONTOLOGY_SERVICE_URL").ok();
    let sync_client = SyncClient::new(ontology_url);
    let state_service = StateService::new(pg.pool().clone(), Some(sync_client));

    let rollback_commit = state_service
        .rollback(
            req.branch_id,
            target_commit_id,
            &req.author_id,
            &req.author_name,
        )
        .await?;

    info!(
        rollback_commit_id = %rollback_commit.id,
        branch_id = %req.branch_id,
        "Rollback completed"
    );

    // Best-effort state snapshot creation (every 50 commits)
    let engine = DeltaReplayEngine::new(pg.pool().clone());
    engine.maybe_create_snapshot(rollback_commit.id).await;

    Ok((
        StatusCode::CREATED,
        Json(CommitDetailResponse::from(rollback_commit)),
    ))
}

#[cfg(test)]
mod tests {
    use super::*;
    use axum::{
        body::Body,
        extract::Path,
        http::{Method, StatusCode},
    };
    use chrono::Utc;
    use tower::ServiceExt;
    use uuid::Uuid;

    use crate::models::{Commit, CommitDelta, CommitSummary, TripleRef};
    use crate::repositories::commit_repo::mock::MockCommitRepository;
    use crate::routes;

    fn test_state() -> Arc<AppState> {
        Arc::new(AppState {
            pg: None,
            branch_repo: None,
            commit_repo: None,
        })
    }

    /// Builds a test router with the given mock commit repository injected.
    fn build_app(repo: MockCommitRepository) -> axum::Router {
        let state = Arc::new(AppState {
            pg: None,
            branch_repo: None,
            commit_repo: Some(Arc::new(repo)),
        });
        routes::build_routes().with_state(state)
    }

    fn req(method: Method, uri: &str, body: Option<&str>) -> axum::http::Request<Body> {
        let mut b = axum::http::Request::builder().method(method).uri(uri);
        if let Some(body) = body {
            b = b.header("Content-Type", "application/json");
            b.body(Body::from(body.to_string())).unwrap()
        } else {
            b.body(Body::empty()).unwrap()
        }
    }

    fn sample_commit() -> Commit {
        Commit {
            id: Uuid::new_v4(),
            branch_id: Uuid::new_v4(),
            parent_commit_id: None,
            message: "test".to_string(),
            author_id: "user".to_string(),
            author_name: "User".to_string(),
            delta: CommitDelta {
                added_triples: vec![TripleRef {
                    s: "A".to_string(),
                    p: "B".to_string(),
                    o: "C".to_string(),
                }],
                removed_triples: vec![],
                modified_triples: vec![],
                merge_metadata: None,
            },
            created_at: Utc::now(),
        }
    }

    #[tokio::test]
    async fn test_get_commit_no_db_returns_pg_not_configured() {
        let state = test_state();
        let id = Uuid::new_v4();
        let result = get_commit_handler(State(state), Path(id)).await;
        assert!(result.is_err());
        match result {
            Err(VersionError::PgNotConfigured) => {} // expected
            _ => panic!("Expected PgNotConfigured error"),
        }
    }

    #[tokio::test]
    async fn test_get_commit_delta_no_db_returns_pg_not_configured() {
        use crate::handlers::get_commit_delta_handler;
        let state = test_state();
        let id = Uuid::new_v4();
        let result = get_commit_delta_handler(State(state), Path(id)).await;
        assert!(result.is_err());
        match result {
            Err(VersionError::PgNotConfigured) => {} // expected
            _ => panic!("Expected PgNotConfigured error"),
        }
    }

    #[tokio::test]
    async fn test_list_commits_no_db_returns_pg_not_configured() {
        let state = test_state();
        let params = ListCommitsParams {
            branch_id: None,
            page: 0,
            per_page: 20,
        };
        let result = list_commits_handler(State(state), Query(params)).await;
        assert!(result.is_err());
        match result {
            Err(VersionError::PgNotConfigured) => {} // expected
            _ => panic!("Expected PgNotConfigured error"),
        }
    }

    #[tokio::test]
    async fn test_create_commit_no_db_returns_pg_not_configured() {
        let state = test_state();
        let req = CreateCommitRequest {
            branch_id: Uuid::new_v4(),
            message: "test".to_string(),
            author_id: "user".to_string(),
            author_name: "User".to_string(),
            delta: CommitDelta {
                added_triples: vec![TripleRef {
                    s: "A".to_string(),
                    p: "B".to_string(),
                    o: "C".to_string(),
                }],
                removed_triples: vec![],
                modified_triples: vec![],
                ..Default::default()
            },
        };
        let result = create_commit_handler(State(state), Json(req)).await;
        assert!(result.is_err());
        match result {
            Err(VersionError::PgNotConfigured) => {} // expected
            _ => panic!("Expected PgNotConfigured error"),
        }
    }

    // ── Mock-based smoke tests ───────────────────────────────────────────

    #[tokio::test]
    async fn test_get_commit_mock_returns_ok() {
        let mut mock = MockCommitRepository::new();
        let commit = sample_commit();
        let commit_id = commit.id;
        mock.get_by_id_result = std::sync::Mutex::new(Some(Ok(commit)));

        let app = build_app(mock);
        let resp = app
            .oneshot(req(
                Method::GET,
                &format!("/api/v1/versioning/commits/{commit_id}"),
                None,
            ))
            .await
            .unwrap();
        assert_eq!(resp.status(), StatusCode::OK);
    }

    #[tokio::test]
    async fn test_list_commits_mock_returns_ok() {
        let mut mock = MockCommitRepository::new();
        mock.list_result = std::sync::Mutex::new(Some(Ok(crate::models::PaginatedResponse {
            items: vec![CommitSummary::from(sample_commit())],
            total: 1,
            page: 0,
            per_page: 20,
        })));

        let app = build_app(mock);
        let resp = app
            .oneshot(req(Method::GET, "/api/v1/versioning/commits", None))
            .await
            .unwrap();
        assert_eq!(resp.status(), StatusCode::OK);
    }

    #[tokio::test]
    async fn test_get_delta_nonexistent_mock_returns_404() {
        let mut mock = MockCommitRepository::new();
        mock.get_by_id_result = std::sync::Mutex::new(Some(Err(VersionError::CommitNotFound(
            "00000000-0000-0000-0000-000000000000".to_string(),
        ))));

        let app = build_app(mock);
        let resp = app
            .oneshot(req(
                Method::GET,
                "/api/v1/versioning/commits/00000000-0000-0000-0000-000000000000/delta",
                None,
            ))
            .await
            .unwrap();
        assert_eq!(resp.status(), StatusCode::NOT_FOUND);
    }

    #[tokio::test]
    async fn test_get_semantic_diff_nonexistent_mock_returns_404() {
        let mut mock = MockCommitRepository::new();
        mock.get_by_id_result = std::sync::Mutex::new(Some(Err(VersionError::CommitNotFound(
            "00000000-0000-0000-0000-000000000000".to_string(),
        ))));

        let app = build_app(mock);
        let resp = app
            .oneshot(req(
                Method::GET,
                "/api/v1/versioning/commits/00000000-0000-0000-0000-000000000000/semantic-diff",
                None,
            ))
            .await
            .unwrap();
        assert_eq!(resp.status(), StatusCode::NOT_FOUND);
    }
}
