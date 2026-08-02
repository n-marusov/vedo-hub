//! Axum HTTP handlers for branch operations.
//!
//! Provides handlers for creating, listing, getting, deleting, merging,
//! and switching branches.
use std::sync::Arc;

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    Json,
};
use serde::Deserialize;
use tracing::{debug, info};
use uuid::Uuid;

use crate::error::VersionError;
use crate::models::{
    BranchSummary, BranchWithCommit, CreateBranchRequest, DeleteBranchRequest,
    MergeBranchesRequest, MergeResponse, PaginatedBranchesResponse, SwitchBranchResponse,
};
use crate::repositories::{BranchRepository, BranchRepositoryTrait};
use crate::AppState;

/// Helper: extracts a `BranchRepository` from the application state.
/// Prefers an injected mock repo when available, otherwise creates a real
/// repository from the `PostgreSQL` pool.
fn repo_from_state(
    state: &AppState,
) -> Result<Arc<dyn BranchRepositoryTrait + Send + Sync>, VersionError> {
    // Use injected mock repo if available (for testing)
    if let Some(repo) = &state.branch_repo {
        return Ok(Arc::clone(repo));
    }
    match &state.pg {
        Some(pool) => Ok(Arc::new(BranchRepository::new(pool.pool().clone()))),
        None => Err(VersionError::PgNotConfigured),
    }
}

/// Query parameters for listing branches.
#[derive(Debug, Deserialize)]
pub struct ListBranchesParams {
    pub ontology_id: Uuid,
    /// Optional reference branch ID for ahead/behind computation.
    pub reference_branch_id: Option<Uuid>,
}

// ── Handlers ──────────────────────────────────────────────────────────────

/// POST /api/v1/versioning/branches — Create a new branch.
pub async fn create_branch_handler(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateBranchRequest>,
) -> Result<(StatusCode, Json<BranchSummary>), VersionError> {
    debug!(
        name = %req.name,
        ontology_id = %req.ontology_id,
        "create_branch_handler called"
    );

    let repo = repo_from_state(&state)?;
    let branch = repo.create(&req).await?;

    info!(
        branch_id = %branch.id,
        name = %branch.name,
        "Branch created successfully"
    );

    Ok((StatusCode::CREATED, Json(BranchSummary::from(branch))))
}

/// GET /api/v1/versioning/branches/{id} — Get branch details.
pub async fn get_branch_handler(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> Result<Json<BranchSummary>, VersionError> {
    debug!(branch_id = %id, "get_branch_handler called");

    let repo = repo_from_state(&state)?;
    let branch = repo.get_by_id(id).await?;

    Ok(Json(BranchSummary::from(branch)))
}

/// GET /api/v1/versioning/branches — List branches for an ontology.
pub async fn list_branches_handler(
    State(state): State<Arc<AppState>>,
    Query(params): Query<ListBranchesParams>,
) -> Result<Json<PaginatedBranchesResponse<BranchWithCommit>>, VersionError> {
    debug!(
        ontology_id = %params.ontology_id,
        "list_branches_handler called"
    );

    let repo = repo_from_state(&state)?;
    let branches = repo
        .list_by_ontology(params.ontology_id, params.reference_branch_id)
        .await?;
    let total = branches.len() as u64;

    Ok(Json(PaginatedBranchesResponse {
        items: branches,
        total,
    }))
}

/// DELETE /api/v1/versioning/branches/{id} — Delete a branch.
pub async fn delete_branch_handler(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
    Json(req): Json<DeleteBranchRequest>,
) -> Result<StatusCode, VersionError> {
    debug!(branch_id = %id, "delete_branch_handler called");

    let repo = repo_from_state(&state)?;
    repo.delete(id, &req).await?;

    info!(branch_id = %id, "Branch deleted successfully");
    Ok(StatusCode::NO_CONTENT)
}

/// POST /api/v1/versioning/branches/{id}/switch — Switch to a branch.
pub async fn switch_branch_handler(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> Result<Json<SwitchBranchResponse>, VersionError> {
    debug!(branch_id = %id, "switch_branch_handler called");

    let repo = repo_from_state(&state)?;
    let response = repo.switch_branch(id).await?;

    info!(
        branch_id = %response.branch_id,
        "Branch switch prepared"
    );

    Ok(Json(response))
}

/// POST /api/v1/versioning/branches/merge — Merge two branches.
pub async fn merge_branches_handler(
    State(state): State<Arc<AppState>>,
    Json(req): Json<MergeBranchesRequest>,
) -> Result<Json<MergeResponse>, VersionError> {
    debug!(
        source = %req.source_branch_id,
        target = %req.target_branch_id,
        "merge_branches_handler called"
    );

    let repo = repo_from_state(&state)?;
    let response = repo.merge_branches(&req).await?;

    info!(
        merge_commit_id = %response.merge_commit_id,
        "Branches merged successfully"
    );

    Ok(Json(response))
}

#[cfg(test)]
mod tests {
    use super::*;
    use axum::{
        body::Body,
        http::{Method, StatusCode},
    };
    use chrono::Utc;
    use tower::ServiceExt;
    use uuid::Uuid;

    use crate::models::Branch;
    use crate::repositories::branch_repo::mock::MockBranchRepository;
    use crate::routes;

    fn test_state() -> Arc<AppState> {
        Arc::new(AppState {
            pg: None,
            branch_repo: None,
            commit_repo: None,
        })
    }

    /// Builds a test router with the given mock branch repository injected.
    fn build_app(repo: MockBranchRepository) -> axum::Router {
        let state = Arc::new(AppState {
            pg: None,
            branch_repo: Some(Arc::new(repo)),
            commit_repo: None,
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

    #[tokio::test]
    async fn test_create_branch_no_db_returns_pg_not_configured() {
        let state = test_state();
        let req = CreateBranchRequest {
            name: "test-branch".to_string(),
            ontology_id: Uuid::new_v4(),
            source_branch_id: None,
        };
        let result = create_branch_handler(State(state), Json(req)).await;
        assert!(result.is_err());
        match result {
            Err(VersionError::PgNotConfigured) => {}
            _ => panic!("Expected PgNotConfigured error"),
        }
    }

    #[tokio::test]
    async fn test_get_branch_no_db_returns_pg_not_configured() {
        let state = test_state();
        let id = Uuid::new_v4();
        let result = get_branch_handler(State(state), Path(id)).await;
        assert!(result.is_err());
        match result {
            Err(VersionError::PgNotConfigured) => {}
            _ => panic!("Expected PgNotConfigured error"),
        }
    }

    #[tokio::test]
    async fn test_list_branches_no_db_returns_pg_not_configured() {
        let state = test_state();
        let params = ListBranchesParams {
            ontology_id: Uuid::new_v4(),
            reference_branch_id: None,
        };
        let result = list_branches_handler(State(state), Query(params)).await;
        assert!(result.is_err());
        match result {
            Err(VersionError::PgNotConfigured) => {}
            _ => panic!("Expected PgNotConfigured error"),
        }
    }

    #[tokio::test]
    async fn test_delete_branch_no_db_returns_pg_not_configured() {
        let state = test_state();
        let id = Uuid::new_v4();
        let req = DeleteBranchRequest { force: false };
        let result = delete_branch_handler(State(state), Path(id), Json(req)).await;
        assert!(result.is_err());
        match result {
            Err(VersionError::PgNotConfigured) => {}
            _ => panic!("Expected PgNotConfigured error"),
        }
    }

    #[tokio::test]
    async fn test_switch_branch_no_db_returns_pg_not_configured() {
        let state = test_state();
        let id = Uuid::new_v4();
        let result = switch_branch_handler(State(state), Path(id)).await;
        assert!(result.is_err());
        match result {
            Err(VersionError::PgNotConfigured) => {}
            _ => panic!("Expected PgNotConfigured error"),
        }
    }

    #[tokio::test]
    async fn test_merge_branches_no_db_returns_pg_not_configured() {
        let state = test_state();
        let req = MergeBranchesRequest {
            source_branch_id: Uuid::new_v4(),
            target_branch_id: Uuid::new_v4(),
            message: "merge".to_string(),
            author_id: "user".to_string(),
            author_name: "User".to_string(),
        };
        let result = merge_branches_handler(State(state), Json(req)).await;
        assert!(result.is_err());
        match result {
            Err(VersionError::PgNotConfigured) => {}
            _ => panic!("Expected PgNotConfigured error"),
        }
    }

    // ── Mock-based smoke tests ─────────────────────────────────────────���

    #[tokio::test]
    async fn test_create_branch_mock_returns_created() {
        let mut mock = MockBranchRepository::new();
        mock.create_result = std::sync::Mutex::new(Some(Ok(Branch {
            id: Uuid::new_v4(),
            name: "feature/test-branch".to_string(),
            ontology_id: Uuid::new_v4(),
            head_commit_id: None,
            created_at: Utc::now(),
            is_protected: false,
        })));

        let app = build_app(mock);
        let resp = app
            .oneshot(req(
                Method::POST,
                "/api/v1/versioning/branches",
                Some(r#"{"ontology_id":"00000000-0000-0000-0000-000000000001","name":"feature/test-branch"}"#),
            ))
            .await
            .unwrap();
        assert_eq!(resp.status(), StatusCode::CREATED);
    }

    #[tokio::test]
    async fn test_get_branch_mock_returns_ok() {
        let mut mock = MockBranchRepository::new();
        mock.get_by_id_result = std::sync::Mutex::new(Some(Ok(Branch {
            id: Uuid::new_v4(),
            name: "main".to_string(),
            ontology_id: Uuid::new_v4(),
            head_commit_id: None,
            created_at: Utc::now(),
            is_protected: true,
        })));

        let app = build_app(mock);
        let resp = app
            .oneshot(req(
                Method::GET,
                "/api/v1/versioning/branches/00000000-0000-0000-0000-000000000001",
                None,
            ))
            .await
            .unwrap();
        assert_eq!(resp.status(), StatusCode::OK);
    }

    #[tokio::test]
    async fn test_list_branches_mock_returns_ok() {
        let mut mock = MockBranchRepository::new();
        mock.list_by_ontology_result = std::sync::Mutex::new(Some(Ok(vec![])));

        let app = build_app(mock);
        let resp = app
            .oneshot(req(
                Method::GET,
                "/api/v1/versioning/branches?ontology_id=00000000-0000-0000-0000-000000000001",
                None,
            ))
            .await
            .unwrap();
        assert_eq!(resp.status(), StatusCode::OK);
    }

    #[tokio::test]
    async fn test_delete_branch_mock_returns_no_content() {
        let mut mock = MockBranchRepository::new();
        mock.delete_result = std::sync::Mutex::new(Some(Ok(())));

        let app = build_app(mock);
        let resp = app
            .oneshot(req(
                Method::DELETE,
                "/api/v1/versioning/branches/00000000-0000-0000-0000-000000000001",
                Some(r#"{"force":false}"#),
            ))
            .await
            .unwrap();
        assert_eq!(resp.status(), StatusCode::NO_CONTENT);
    }

    #[tokio::test]
    async fn test_switch_branch_mock_returns_ok() {
        let mut mock = MockBranchRepository::new();
        mock.switch_branch_result = std::sync::Mutex::new(Some(Ok(SwitchBranchResponse {
            branch_id: Uuid::new_v4(),
            head_commit_id: None,
        })));

        let app = build_app(mock);
        let resp = app
            .oneshot(req(
                Method::POST,
                "/api/v1/versioning/branches/00000000-0000-0000-0000-000000000001/switch",
                None,
            ))
            .await
            .unwrap();
        assert_eq!(resp.status(), StatusCode::OK);
    }

    #[tokio::test]
    async fn test_merge_branches_mock_returns_ok_or_error() {
        let mut mock = MockBranchRepository::new();
        mock.merge_branches_result = std::sync::Mutex::new(Some(Ok(MergeResponse {
            merge_commit_id: Uuid::new_v4(),
            source_branch: Uuid::new_v4(),
            target_branch: Uuid::new_v4(),
            conflict_count: 0,
            auto_resolved: true,
        })));

        let app = build_app(mock);
        let resp = app
            .oneshot(req(
                Method::POST,
                "/api/v1/versioning/branches/merge",
                Some(r#"{"source_branch_id":"00000000-0000-0000-0000-000000000001","target_branch_id":"00000000-0000-0000-0000-000000000002","message":"merge","author_id":"user","author_name":"User"}"#),
            ))
            .await
            .unwrap();
        assert!(resp.status().is_success() || resp.status().is_client_error());
    }

    #[tokio::test]
    async fn test_get_branch_mock_returns_not_found() {
        let mut mock = MockBranchRepository::new();
        mock.get_by_id_result = std::sync::Mutex::new(Some(Err(VersionError::BranchNotFound(
            "00000000-0000-0000-0000-000000000001".to_string(),
        ))));

        let app = build_app(mock);
        let resp = app
            .oneshot(req(
                Method::GET,
                "/api/v1/versioning/branches/00000000-0000-0000-0000-000000000001",
                None,
            ))
            .await
            .unwrap();
        assert_eq!(resp.status(), StatusCode::NOT_FOUND);
    }
}
