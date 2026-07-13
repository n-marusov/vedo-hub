//! Axum HTTP handlers for branch operations.
//!
//! Provides handlers for creating, listing, getting, deleting, merging,
//! and switching branches.

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    Json,
};
use serde::Deserialize;
use std::sync::Arc;
use tracing::{debug, info};
use uuid::Uuid;

use crate::error::VersionError;
use crate::models::{
    BranchSummary, BranchWithCommit, CreateBranchRequest, DeleteBranchRequest,
    MergeBranchesRequest, MergeResponse, PaginatedBranchesResponse, SwitchBranchResponse,
};
use crate::repositories::BranchRepository;
use crate::AppState;

/// Helper: extracts a `BranchRepository` from the application state.
fn repo_from_state(state: &AppState) -> Result<BranchRepository, VersionError> {
    match &state.pg {
        Some(pool) => Ok(BranchRepository::new(pool.pool().clone())),
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
    use uuid::Uuid;

    fn test_state() -> Arc<AppState> {
        Arc::new(AppState { pg: None })
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
}
