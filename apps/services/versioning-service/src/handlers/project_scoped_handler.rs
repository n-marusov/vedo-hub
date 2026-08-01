//! Project-scoped, GitLab-aligned versioning handlers.
//!
//! Validates: REQ-FUN.API.rest-gitlab-alignment
//!
//! These handlers back the GitLab-aligned REST surface:
//!
//!   /api/v1/projects/{pid}/repository/commits|branches
//!
//! per ADR-DES.API.rest-gitlab-alignment §4 (Versioning). The `{pid}` path
//! parameter is accepted as the project identifier; the underlying data
//! model still uses `ontology_id` internally (Project ↔ Ontology is 1:1 per
//! ADR-DES.API.organization-rest-endpoints), so `pid` resolves to the
//! ontology scope of the repository calls.
//!
//! URL-level renames vs the legacy surface:
//!   - commit `{id}` → `{sha}`   (the value remains the commit UUID for now;
//!     a dedicated SHA column is a schema-level change tracked separately)
//!   - `delta`      → `diff`
//!   - branch `{id}` → `{name}`  (Branch.name is the public identifier)

use std::sync::Arc;

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    Json,
};
use uuid::Uuid;

use crate::error::VersionError;
use crate::handlers::branch_handler::{self, ListBranchesParams};
use crate::handlers::commit_handler::{self};
use crate::models::branch::{
    BranchSummary, BranchWithCommit, CreateBranchRequest, DeleteBranchRequest,
    PaginatedBranchesResponse,
};
use crate::models::commit::{
    CommitDetailResponse, CommitSummary, CreateCommitRequest, ListCommitsParams, PaginatedResponse,
};
use crate::repositories::branch_repo::BranchRepositoryTrait;
use crate::AppState;

/// Returns the branch repository from state.
fn branch_repo(
    state: &AppState,
) -> Result<Arc<dyn BranchRepositoryTrait + Send + Sync>, VersionError> {
    if let Some(repo) = &state.branch_repo {
        return Ok(repo.clone());
    }
    let pool = state.pg.as_ref().ok_or(VersionError::PgNotConfigured)?;
    Ok(Arc::new(
        crate::repositories::branch_repo::BranchRepository::new(pool.pool().clone()),
    ))
}

/// POST /api/v1/projects/{pid}/repository/commits — create a commit
/// (VEDO extension; GitLab creates commits only via git push).
pub async fn create_project_commit_handler(
    State(state): State<Arc<AppState>>,
    Path(_project_id): Path<String>,
    Json(req): Json<CreateCommitRequest>,
) -> Result<(StatusCode, Json<CommitDetailResponse>), VersionError> {
    commit_handler::create_commit_handler(State(state), Json(req)).await
}

/// GET /api/v1/projects/{pid}/repository/commits — list commits (paginated).
pub async fn list_project_commits_handler(
    State(state): State<Arc<AppState>>,
    Path(_project_id): Path<String>,
    Query(params): Query<ListCommitsParams>,
) -> Result<Json<PaginatedResponse<CommitSummary>>, VersionError> {
    commit_handler::list_commits_handler(State(state), Query(params)).await
}

/// GET /api/v1/projects/{pid}/repository/commits/{sha} — get a commit.
pub async fn get_project_commit_handler(
    State(state): State<Arc<AppState>>,
    Path((_project_id, sha)): Path<(String, String)>,
) -> Result<Json<CommitDetailResponse>, VersionError> {
    let id = parse_sha(&sha)?;
    commit_handler::get_commit_handler(State(state), Path(id)).await
}

/// GET /api/v1/projects/{pid}/repository/commits/{sha}/diff — semantic diff
/// (legacy `delta` renamed to `diff`).
pub async fn get_project_commit_diff_handler(
    State(state): State<Arc<AppState>>,
    Path((_project_id, sha)): Path<(String, String)>,
) -> Result<Json<crate::models::commit::CommitDeltaPreview>, VersionError> {
    let id = parse_sha(&sha)?;
    commit_handler::get_commit_delta_handler(State(state), Path(id)).await
}

/// POST /api/v1/projects/{pid}/repository/branches — create a branch.
pub async fn create_project_branch_handler(
    State(state): State<Arc<AppState>>,
    Path(_project_id): Path<String>,
    Json(req): Json<CreateBranchRequest>,
) -> Result<(StatusCode, Json<BranchSummary>), VersionError> {
    branch_handler::create_branch_handler(State(state), Json(req)).await
}

/// GET /api/v1/projects/{pid}/repository/branches — list branches for the
/// project's ontology.
pub async fn list_project_branches_handler(
    State(state): State<Arc<AppState>>,
    Path(_project_id): Path<String>,
    Query(params): Query<ListBranchesParams>,
) -> Result<Json<PaginatedBranchesResponse<BranchWithCommit>>, VersionError> {
    branch_handler::list_branches_handler(State(state), Query(params)).await
}

/// GET /api/v1/projects/{pid}/repository/branches/{name} — get a branch by
/// its public name (GitLab-aligned; legacy used UUID id).
pub async fn get_project_branch_handler(
    State(state): State<Arc<AppState>>,
    Path((project_id, name)): Path<(String, String)>,
) -> Result<Json<BranchSummary>, VersionError> {
    tracing::debug!(project_id = %project_id, %name, "get_project_branch_handler called");
    let repo = branch_repo(&state)?;
    let branch = repo
        .get_by_name(uuid_from_project(&project_id)?, &name)
        .await?;
    Ok(Json(BranchSummary::from(branch)))
}

/// DELETE /api/v1/projects/{pid}/repository/branches/{name} — delete a
/// branch by its public name.
pub async fn delete_project_branch_handler(
    State(state): State<Arc<AppState>>,
    Path((project_id, name)): Path<(String, String)>,
    Json(req): Json<DeleteBranchRequest>,
) -> Result<StatusCode, VersionError> {
    tracing::debug!(project_id = %project_id, %name, "delete_project_branch_handler called");
    let repo = branch_repo(&state)?;
    let branch = repo
        .get_by_name(uuid_from_project(&project_id)?, &name)
        .await?;
    repo.delete(branch.id, &req).await?;
    Ok(StatusCode::NO_CONTENT)
}

/// Parses a `{sha}` path segment into the internal commit UUID.
///
/// The URL param is named `sha` per GitLab alignment; the value is currently
/// the commit UUID serialized as a string (a dedicated SHA column is a
/// schema-level change tracked separately).
fn parse_sha(sha: &str) -> Result<Uuid, VersionError> {
    Uuid::parse_str(sha).map_err(|_| VersionError::InvalidRequest(format!("invalid sha: {sha}")))
}

/// Resolves a `{pid}` path segment to the internal ontology UUID.
///
/// Project ↔ Ontology is 1:1 (ADR-DES.API.organization-rest-endpoints);
/// the public project id and the internal ontology id are the same UUID in
/// the current model.
fn uuid_from_project(pid: &str) -> Result<Uuid, VersionError> {
    Uuid::parse_str(pid)
        .map_err(|_| VersionError::InvalidRequest(format!("invalid project id: {pid}")))
}

// Keep the direct-repo helpers reachable for unit tests of this module.
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_sha_valid_uuid() {
        let id = Uuid::new_v4();
        assert_eq!(parse_sha(&id.to_string()).unwrap(), id);
    }

    #[test]
    fn test_parse_sha_invalid_returns_error() {
        assert!(parse_sha("not-a-sha").is_err());
    }

    #[test]
    fn test_uuid_from_project_valid() {
        let id = Uuid::new_v4();
        assert_eq!(uuid_from_project(&id.to_string()).unwrap(), id);
    }

    #[test]
    fn test_uuid_from_project_invalid_returns_error() {
        assert!(uuid_from_project("p1").is_err());
    }
}
