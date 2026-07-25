//! HTTP client used by GraphQL resolvers to query the versioning-service.
//!
//! The GraphQL schema on `ontology-service` is the single facade the frontend
//! Apollo Client talks to. To expose commit and branch data through that same
//! schema without duplicating storage, the resolvers proxy REST calls to the
//! versioning-service (`/api/v1/versioning/*`).
//!
//! The base URL is resolved from the `VERSIONING_SERVICE_URL` env var and
//! falls back to `http://localhost:8083` (the canonical port allocation per
//! `ADR-IMPL.STACK.port-mapping-strategy`). All calls log with the `[FIX]`
//! prefix so failures during the GraphQL integration are easy to isolate.

use chrono::{DateTime, Utc};
use reqwest::Client;
use serde::Deserialize;
use std::fmt::Write as _;
use uuid::Uuid;

/// Default versioning-service URL (matches `ADR-IMPL.STACK.port-mapping-strategy`).
const DEFAULT_VERSIONING_URL: &str = "http://localhost:8083";

/// HTTP shape of `CommitSummary` returned by `GET /api/v1/versioning/commits`.
#[derive(Debug, Deserialize)]
pub struct RemoteCommitSummary {
    pub id: Uuid,
    pub branch_id: Uuid,
    pub parent_commit_id: Option<Uuid>,
    pub message: String,
    pub author_id: String,
    pub author_name: String,
    pub total_changes: usize,
    pub created_at: DateTime<Utc>,
}

/// HTTP shape of `PaginatedResponse<CommitSummary>`.
#[derive(Debug, Deserialize)]
pub struct RemoteCommitPage {
    pub items: Vec<RemoteCommitSummary>,
    pub total: u64,
    pub page: u64,
    pub per_page: u64,
}

/// HTTP shape of `BranchWithCommit` returned by `GET /api/v1/versioning/branches`.
#[derive(Debug, Deserialize)]
pub struct RemoteBranch {
    pub id: Uuid,
    pub name: String,
    pub ontology_id: Uuid,
    pub head_commit_id: Option<Uuid>,
    pub created_at: DateTime<Utc>,
    pub is_protected: bool,
    #[serde(default)]
    pub last_commit_message: Option<String>,
    #[serde(default)]
    pub last_commit_author: Option<String>,
    #[serde(default)]
    pub last_commit_at: Option<DateTime<Utc>>,
    #[serde(default)]
    pub ahead_count: i64,
    #[serde(default)]
    pub behind_count: i64,
}

/// HTTP shape of `PaginatedBranchesResponse<BranchWithCommit>`.
#[derive(Debug, Deserialize)]
pub struct RemoteBranchPage {
    pub items: Vec<RemoteBranch>,
    pub total: u64,
}

/// Thin HTTP client for the versioning-service REST API.
pub struct VersioningClient {
    base_url: String,
    client: Client,
}

impl VersioningClient {
    /// Resolves the versioning-service URL from `VERSIONING_SERVICE_URL`,
    /// falling back to the default port-mapping value.
    pub fn from_env() -> Self {
        let base_url = std::env::var("VERSIONING_SERVICE_URL")
            .unwrap_or_else(|_| DEFAULT_VERSIONING_URL.to_string());
        Self {
            base_url,
            client: Client::new(),
        }
    }

    fn base(&self) -> &str {
        self.base_url.trim_end_matches('/')
    }

    /// Lists commits, optionally filtered by `branch_id`.
    ///
    /// Calls `GET /api/v1/versioning/commits?branch_id=&page=&per_page=`.
    pub async fn list_commits(
        &self,
        branch_id: Option<Uuid>,
        page: u64,
        per_page: u64,
    ) -> Result<RemoteCommitPage, String> {
        let mut url = format!("{}/api/v1/versioning/commits", self.base());
        let mut sep = '?';
        if let Some(b) = branch_id {
            let _ = write!(url, "{sep}branch_id={b}");
            sep = '&';
        }
        let _ = write!(url, "{sep}page={page}&per_page={per_page}");

        tracing::debug!(url = %url, "[FIX] listing commits via versioning-service");

        self.client
            .get(&url)
            .timeout(std::time::Duration::from_secs(10))
            .send()
            .await
            .map_err(|e| {
                tracing::warn!(error = %e, url = %url, "[FIX] commit list fetch failed");
                format!("versioning-service unreachable: {e}")
            })?
            .error_for_status()
            .map_err(|e| {
                tracing::warn!(status = %e, url = %url, "[FIX] commit list non-2xx");
                format!("versioning-service returned error: {e}")
            })?
            .json::<RemoteCommitPage>()
            .await
            .map_err(|e| {
                tracing::warn!(error = %e, "[FIX] commit list body parse failed");
                format!("invalid commit payload: {e}")
            })
    }

    /// Lists branches for an ontology, optionally relative to a reference branch.
    ///
    /// Calls `GET /api/v1/versioning/branches?ontology_id=&reference_branch_id=`.
    pub async fn list_branches(
        &self,
        ontology_id: Uuid,
        reference_branch_id: Option<Uuid>,
    ) -> Result<RemoteBranchPage, String> {
        let mut url = format!(
            "{base}/api/v1/versioning/branches?ontology_id={ontology_id}",
            base = self.base(),
            ontology_id = ontology_id
        );
        if let Some(r) = reference_branch_id {
            let _ = write!(url, "&reference_branch_id={r}");
        }

        tracing::debug!(url = %url, "[FIX] listing branches via versioning-service");

        self.client
            .get(&url)
            .timeout(std::time::Duration::from_secs(10))
            .send()
            .await
            .map_err(|e| {
                tracing::warn!(error = %e, url = %url, "[FIX] branch list fetch failed");
                format!("versioning-service unreachable: {e}")
            })?
            .error_for_status()
            .map_err(|e| {
                tracing::warn!(status = %e, url = %url, "[FIX] branch list non-2xx");
                format!("versioning-service returned error: {e}")
            })?
            .json::<RemoteBranchPage>()
            .await
            .map_err(|e| {
                tracing::warn!(error = %e, "[FIX] branch list body parse failed");
                format!("invalid branch payload: {e}")
            })
    }
}
