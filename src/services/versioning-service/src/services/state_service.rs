//! State service — checkout, rollback, and materialized state management.
//!
//! Orchestrates the full workflow for switching ontology versions:
//! 1. Replay deltas to materialize state
//! 2. Create inverse deltas for rollback
//! 3. Sync materialized state to Neo4j via the sync client

use sqlx::PgPool;
use uuid::Uuid;

use crate::error::VersionError;
use crate::models::{Commit, CreateCommitRequest};
use crate::repositories::{BranchRepository, CommitRepository};
use crate::services::delta_service::{DeltaReplayEngine, MaterializedState};
use crate::services::sync_client::SyncClient;

/// Service for managing ontology state across branch checkouts and rollbacks.
pub struct StateService {
    delta_engine: DeltaReplayEngine,
    commit_repo: CommitRepository,
    branch_repo: BranchRepository,
    sync_client: Option<SyncClient>,
}

impl StateService {
    /// Creates a new `StateService` with the given pool and optional sync client.
    pub fn new(pool: PgPool, sync_client: Option<SyncClient>) -> Self {
        Self {
            delta_engine: DeltaReplayEngine::new(pool.clone()),
            commit_repo: CommitRepository::new(pool.clone()),
            branch_repo: BranchRepository::new(pool.clone()),
            sync_client,
        }
    }

    /// Performs a checkout to a specific commit on a branch.
    ///
    /// Replays all deltas from root to the target commit, producing the
    /// materialized state. If a sync client is configured, pushes the state
    /// to Neo4j.
    pub async fn checkout(
        &self,
        branch_id: Uuid,
        target_commit_id: Option<Uuid>,
    ) -> Result<CheckoutResult, VersionError> {
        tracing::debug!(
            branch_id = %branch_id,
            target_commit_id = ?target_commit_id,
            "Starting checkout"
        );

        let branch = self.branch_repo.get_by_id(branch_id).await?;

        // Determine the commit to check out to
        let commit_id = target_commit_id.unwrap_or(branch.head_commit_id.ok_or_else(|| {
            VersionError::InvalidRequest(
                "Branch has no commits; use a target commit ID".to_string(),
            )
        })?);

        // Verify the target commit is reachable from the branch's head.
        // This prevents checkout to an unrelated commit from a different
        // branch or ontology.
        if let Some(head_id) = branch.head_commit_id {
            if !self.commit_repo.is_ancestor_of(head_id, commit_id).await? {
                return Err(VersionError::InvalidRequest(format!(
                    "Commit {} is not reachable from branch {} (head commit {})",
                    commit_id, branch_id, head_id
                )));
            }
        }

        // Materialize state from the commit chain
        let materialized = self.delta_engine.materialize(commit_id).await?;

        tracing::info!(
            branch_id = %branch_id,
            commit_id = %commit_id,
            triple_count = materialized.triples.len(),
            delta_count = materialized.delta_count,
            "Checkout materialized state"
        );

        // Sync to ontology service if client is configured
        let sync_succeeded = if let Some(client) = &self.sync_client {
            let ontology_id = branch.ontology_id.to_string();
            match client.push_state(&ontology_id, &materialized.triples).await {
                Ok(()) => true,
                Err(e) => {
                    tracing::warn!(
                        error = %e,
                        "[FIX] State sync to ontology service failed — continuing"
                    );
                    false
                }
            }
        } else {
            false
        };

        Ok(CheckoutResult {
            branch_id,
            commit_id,
            triple_count: materialized.triples.len() as u64,
            delta_count: materialized.delta_count as u64,
            synced: self.sync_client.is_some() && sync_succeeded,
        })
    }

    /// Performs a rollback by creating an inverse delta between the current
    /// head and the target commit, then creating a new commit.
    pub async fn rollback(
        &self,
        branch_id: Uuid,
        target_commit_id: Uuid,
        author_id: &str,
        author_name: &str,
    ) -> Result<Commit, VersionError> {
        tracing::debug!(
            branch_id = %branch_id,
            target_commit_id = %target_commit_id,
            "Starting rollback"
        );

        let branch = self.branch_repo.get_by_id(branch_id).await?;
        let current_head = branch.head_commit_id.ok_or_else(|| {
            VersionError::InvalidRequest("Branch has no commits to rollback".to_string())
        })?;

        // Verify the target commit is reachable from the current head.
        // This prevents rollback to an unrelated commit from a different
        // branch or ontology.
        if !self
            .commit_repo
            .is_ancestor_of(current_head, target_commit_id)
            .await?
        {
            return Err(VersionError::InvalidRequest(format!(
                "Target commit {} is not reachable from branch {} (head commit {})",
                target_commit_id, branch_id, current_head
            )));
        }

        // Compute inverse delta
        let inverse_delta = self
            .delta_engine
            .compute_inverse_delta(current_head, target_commit_id)
            .await?;

        if inverse_delta.is_empty() {
            return Err(VersionError::EmptyDelta);
        }

        // Create a rollback commit with the inverse delta
        let req = CreateCommitRequest {
            branch_id,
            message: format!("Rollback to commit {}", target_commit_id.to_string()),
            author_id: author_id.to_string(),
            author_name: author_name.to_string(),
            delta: inverse_delta,
        };

        let rollback_commit = self.commit_repo.create(&req).await?;

        tracing::info!(
            rollback_commit_id = %rollback_commit.id,
            branch_id = %branch_id,
            target_commit_id = %target_commit_id,
            total_changes = rollback_commit.delta.total_changes(),
            "Rollback completed"
        );

        // Optionally sync the rollback state to Neo4j
        let _sync_succeeded = if let Some(client) = &self.sync_client {
            let materialized = self.delta_engine.materialize(rollback_commit.id).await?;
            let ontology_id = branch.ontology_id.to_string();
            match client.push_state(&ontology_id, &materialized.triples).await {
                Ok(()) => true,
                Err(e) => {
                    tracing::warn!(
                        error = %e,
                        "[FIX] Post-rollback state sync failed — continuing"
                    );
                    false
                }
            }
        } else {
            false
        };

        // Best-effort state snapshot creation (every 50 commits)
        self.delta_engine
            .maybe_create_snapshot(rollback_commit.id)
            .await;

        Ok(rollback_commit)
    }

    /// Returns the materialized state for a commit without performing
    /// a checkout (read-only).
    pub async fn preview_state(&self, commit_id: Uuid) -> Result<MaterializedState, VersionError> {
        self.delta_engine.materialize(commit_id).await
    }
}

/// Result of a checkout operation.
#[derive(Debug, Clone)]
pub struct CheckoutResult {
    pub branch_id: Uuid,
    pub commit_id: Uuid,
    pub triple_count: u64,
    pub delta_count: u64,
    pub synced: bool,
}
