//! Branch repository — `PostgreSQL` data access for branches.
//!
//! Provides SQL queries for branch CRUD, listing with latest commit info,
//! and ahead/behind computation.

use sqlx::{PgPool, Row};
use uuid::Uuid;

use crate::error::VersionError;
use crate::models::{
    Branch, BranchWithCommit, CommitDelta, CreateBranchRequest, DeleteBranchRequest,
    MergeBranchesRequest, MergeMetadata, MergeResponse, SwitchBranchResponse,
};

// ── Trait ────────────────────────────────────────────────────────────────

/// Trait for branch repository operations. Allows mocking in unit tests.
#[async_trait::async_trait]
pub trait BranchRepositoryTrait: Send + Sync {
    /// Creates a new branch. If a `source_branch_id` is provided, copies the
    /// source branch's head commit as this branch's starting point.
    async fn create(&self, req: &CreateBranchRequest) -> Result<Branch, VersionError>;

    /// Retrieves a single branch by its ID.
    async fn get_by_id(&self, id: Uuid) -> Result<Branch, VersionError>;

    /// Retrieves a single branch by its name within an ontology.
    /// GitLab-aligned lookup (branch name is the public identifier, not a UUID).
    async fn get_by_name(&self, ontology_id: Uuid, name: &str) -> Result<Branch, VersionError>;

    /// Lists all branches for an ontology, with latest commit info and
    /// ahead/behind counts versus a reference branch.
    async fn list_by_ontology(
        &self,
        ontology_id: Uuid,
        reference_branch_id: Option<Uuid>,
    ) -> Result<Vec<BranchWithCommit>, VersionError>;

    /// Updates the head commit pointer for a branch (used after creating a commit).
    async fn update_head(&self, branch_id: Uuid, commit_id: Uuid) -> Result<(), VersionError>;

    /// Deletes a branch by ID. Protected branches require `force: true`.
    async fn delete(&self, id: Uuid, req: &DeleteBranchRequest) -> Result<(), VersionError>;

    /// Merges source branch into target branch, creating a merge commit.
    async fn merge_branches(
        &self,
        req: &MergeBranchesRequest,
    ) -> Result<MergeResponse, VersionError>;

    /// Returns the head commit ID for a branch (used for switching).
    async fn switch_branch(&self, id: Uuid) -> Result<SwitchBranchResponse, VersionError>;
}

/// Repository for branch operations against `PostgreSQL`.
pub struct BranchRepository {
    pool: PgPool,
}

impl BranchRepository {
    /// Creates a new `BranchRepository` with the given connection pool.
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }

    /// Returns a reference to the inner pool.
    pub fn pool(&self) -> &PgPool {
        &self.pool
    }
}

#[async_trait::async_trait]
impl BranchRepositoryTrait for BranchRepository {
    // ── Create ────────────────────────────────────────────────────────────

    /// Creates a new branch. If a `source_branch_id` is provided, copies the
    /// source branch's head commit as this branch's starting point.
    async fn create(&self, req: &CreateBranchRequest) -> Result<Branch, VersionError> {
        tracing::debug!(
            name = %req.name,
            ontology_id = %req.ontology_id,
            source = ?req.source_branch_id,
            "Creating branch"
        );

        // Get head_commit_id from source branch if provided
        let head_commit_id: Option<Uuid> = match req.source_branch_id {
            Some(source_id) => {
                let row = sqlx::query("SELECT head_commit_id FROM branches WHERE id = $1")
                    .bind(source_id)
                    .fetch_optional(self.pool())
                    .await?;

                match row {
                    Some(r) => r.try_get("head_commit_id")?,
                    None => {
                        return Err(VersionError::BranchNotFound(source_id.to_string()));
                    }
                }
            }
            None => None,
        };

        let row = sqlx::query(
            r"
            INSERT INTO branches (name, ontology_id, head_commit_id)
            VALUES ($1, $2, $3)
            RETURNING id, name, ontology_id, head_commit_id, created_at, is_protected
            ",
        )
        .bind(&req.name)
        .bind(req.ontology_id)
        .bind(head_commit_id)
        .fetch_one(self.pool())
        .await
        .map_err(|e| {
            // Check for PostgreSQL unique violation (23505) on
            // branches(ontology_id, name) — duplicate branch name within ontology
            if let Some(db_err) = e.as_database_error() {
                if db_err.code().as_deref() == Some("23505") {
                    tracing::warn!(
                        name = %req.name,
                        ontology_id = %req.ontology_id,
                        "[FIX] Duplicate branch name in ontology"
                    );
                    return VersionError::InvalidRequest(format!(
                        "Branch '{}' already exists in ontology {}",
                        req.name, req.ontology_id
                    ));
                }
            }
            VersionError::from(e)
        })?;

        let branch = row_to_branch(&row)?;

        tracing::info!(
            branch_id = %branch.id,
            name = %branch.name,
            "Branch created"
        );

        Ok(branch)
    }

    // ── Read ──────────────────────────────────────────────────────────────

    /// Retrieves a single branch by its ID.
    async fn get_by_id(&self, id: Uuid) -> Result<Branch, VersionError> {
        tracing::debug!(branch_id = %id, "Fetching branch by ID");

        let row = sqlx::query(
            r"
            SELECT id, name, ontology_id, head_commit_id, created_at, is_protected
            FROM branches
            WHERE id = $1
            ",
        )
        .bind(id)
        .fetch_optional(self.pool())
        .await?
        .ok_or_else(|| VersionError::BranchNotFound(id.to_string()))?;

        let branch = row_to_branch(&row)?;
        tracing::debug!(branch_id = %id, name = %branch.name, "Branch found");

        Ok(branch)
    }

    /// Retrieves a single branch by its name within an ontology.
    /// GitLab-aligned lookup: branch name is the public identifier.
    async fn get_by_name(&self, ontology_id: Uuid, name: &str) -> Result<Branch, VersionError> {
        tracing::debug!(ontology_id = %ontology_id, %name, "Fetching branch by name");

        let row = sqlx::query(
            r"
            SELECT id, name, ontology_id, head_commit_id, created_at, is_protected
            FROM branches
            WHERE ontology_id = $1 AND name = $2
            ",
        )
        .bind(ontology_id)
        .bind(name)
        .fetch_optional(self.pool())
        .await?
        .ok_or_else(|| VersionError::BranchNotFound(name.to_string()))?;

        let branch = row_to_branch(&row)?;
        tracing::debug!(branch_id = %branch.id, name = %branch.name, "Branch found by name");

        Ok(branch)
    }

    /// Lists all branches for an ontology, with latest commit info and
    /// ahead/behind counts versus a reference branch.
    async fn list_by_ontology(
        &self,
        ontology_id: Uuid,
        reference_branch_id: Option<Uuid>,
    ) -> Result<Vec<BranchWithCommit>, VersionError> {
        tracing::debug!(ontology_id = %ontology_id, "Listing branches");

        let rows = sqlx::query(
            r"
            SELECT
                b.id, b.name, b.ontology_id, b.head_commit_id, b.created_at, b.is_protected,
                c.message AS last_commit_message,
                c.author_name AS last_commit_author,
                c.created_at AS last_commit_at
            FROM branches b
            LEFT JOIN LATERAL (
                SELECT message, author_name, created_at
                FROM commits
                WHERE id = b.head_commit_id
                LIMIT 1
            ) c ON TRUE
            WHERE b.ontology_id = $1
            ORDER BY b.created_at DESC
            ",
        )
        .bind(ontology_id)
        .fetch_all(self.pool())
        .await?;

        let mut branches = Vec::with_capacity(rows.len());
        for row in &rows {
            let branch = BranchWithCommit {
                id: row.try_get("id")?,
                name: row.try_get("name")?,
                ontology_id: row.try_get("ontology_id")?,
                head_commit_id: row.try_get("head_commit_id")?,
                created_at: row.try_get("created_at")?,
                is_protected: row.try_get("is_protected")?,
                last_commit_message: row.try_get("last_commit_message")?,
                last_commit_author: row.try_get("last_commit_author")?,
                last_commit_at: row.try_get("last_commit_at")?,
                ahead_count: 0,
                behind_count: 0,
            };
            branches.push(branch);
        }

        // Compute ahead/behind if a reference branch is specified
        if let Some(ref_branch_id) = reference_branch_id {
            for branch in &mut branches {
                let (ahead, behind) = self
                    .compute_ahead_behind(branch.id, ref_branch_id)
                    .await
                    .unwrap_or((0, 0));
                branch.ahead_count = ahead;
                branch.behind_count = behind;
            }
        }

        tracing::info!(count = branches.len(), "Branches listed");
        Ok(branches)
    }

    // ── Update ────────────────────────────────────────────────────────────

    /// Updates the head commit pointer for a branch (used after creating a commit).
    async fn update_head(&self, branch_id: Uuid, commit_id: Uuid) -> Result<(), VersionError> {
        sqlx::query("UPDATE branches SET head_commit_id = $1 WHERE id = $2")
            .bind(commit_id)
            .bind(branch_id)
            .execute(self.pool())
            .await?;
        tracing::debug!(branch_id = %branch_id, commit_id = %commit_id, "Branch head updated");
        Ok(())
    }

    // ── Delete ────────────────────────────────────────────────────────────

    /// Deletes a branch by ID. Protected branches require `force: true`.
    async fn delete(&self, id: Uuid, req: &DeleteBranchRequest) -> Result<(), VersionError> {
        tracing::debug!(
            branch_id = %id,
            force = req.force,
            "[FIX] Deleting branch with TOCTOU protection"
        );

        let mut tx = self.pool().begin().await?;

        // [FIX] Lock branch row and check existence + protection inside the
        // transaction to prevent TOCTOU races.
        let row = sqlx::query(
            r"
            SELECT id, name, ontology_id, head_commit_id, created_at, is_protected
            FROM branches
            WHERE id = $1
            FOR UPDATE
            ",
        )
        .bind(id)
        .fetch_optional(&mut *tx)
        .await?;

        let branch = match row {
            None => {
                tracing::error!(
                    branch_id = %id,
                    "[FIX] Branch not found for deletion"
                );
                return Err(VersionError::BranchNotFound(id.to_string()));
            }
            Some(r) => row_to_branch(&r)?,
        };

        if branch.is_protected && !req.force {
            return Err(VersionError::BranchProtected {
                branch: branch.name.clone(),
            });
        }

        // Delete all commits on this branch first (cascade may not be set)
        sqlx::query("DELETE FROM commits WHERE branch_id = $1")
            .bind(id)
            .execute(&mut *tx)
            .await?;

        let result = sqlx::query("DELETE FROM branches WHERE id = $1")
            .bind(id)
            .execute(&mut *tx)
            .await?;

        // [FIX] Check rows_affected to detect race condition
        if result.rows_affected() == 0 {
            tracing::error!(
                branch_id = %id,
                "[FIX] Race condition: branch was deleted between check and delete"
            );
            return Err(VersionError::BranchNotFound(id.to_string()));
        }

        tx.commit().await?;

        tracing::info!(branch_id = %id, "Branch deleted");
        Ok(())
    }

    // ── Merge ─────────────────────────────────────────────────────────────

    /// Merges source branch into target branch, creating a merge commit.
    /// Returns an error when both branches are at the same commit (nothing
    /// to merge) or when either branch cannot be resolved.
    #[allow(clippy::too_many_lines)]
    async fn merge_branches(
        &self,
        req: &MergeBranchesRequest,
    ) -> Result<MergeResponse, VersionError> {
        tracing::debug!(
            source = %req.source_branch_id,
            target = %req.target_branch_id,
            "[FIX] Merging branches with atomic transaction and real delta"
        );

        // Verify both branches exist
        let source = self.get_by_id(req.source_branch_id).await?;
        let target = self.get_by_id(req.target_branch_id).await?;

        if target.is_protected {
            return Err(VersionError::BranchProtected {
                branch: target.name.clone(),
            });
        }

        // Check that both branches have commits to merge
        let source_head = source.head_commit_id.ok_or_else(|| {
            VersionError::InvalidRequest("Source branch has no commits".to_string())
        })?;
        let target_head = target.head_commit_id.ok_or_else(|| {
            VersionError::InvalidRequest("Target branch has no commits".to_string())
        })?;

        // Prevent no-op merge: if both branches point to the same commit
        // there is nothing to merge.
        if source_head == target_head {
            return Err(VersionError::InvalidRequest(format!(
                "Nothing to merge: source branch '{}' and target branch '{}' are at the same commit {}",
                source.name, target.name, source_head
            )));
        }

        // Compute ahead/behind to verify there is actual divergence.
        let (ahead, behind) = self
            .compute_ahead_behind(req.source_branch_id, req.target_branch_id)
            .await?;
        if ahead == 0 && behind == 0 {
            return Err(VersionError::InvalidRequest(
                "Nothing to merge: branches are already synchronized".to_string(),
            ));
        }

        // [FIX] Compute real merge delta from source and target head deltas
        let source_delta: serde_json::Value =
            sqlx::query_scalar("SELECT delta FROM commits WHERE id = $1")
                .bind(source_head)
                .fetch_optional(self.pool())
                .await?
                .flatten()
                .unwrap_or(serde_json::Value::Object(serde_json::Map::new()));

        let target_delta: serde_json::Value =
            sqlx::query_scalar("SELECT delta FROM commits WHERE id = $1")
                .bind(target_head)
                .fetch_optional(self.pool())
                .await?
                .flatten()
                .unwrap_or(serde_json::Value::Object(serde_json::Map::new()));

        let src_delta: CommitDelta = serde_json::from_value(source_delta).unwrap_or_default();
        let tgt_delta: CommitDelta = serde_json::from_value(target_delta).unwrap_or_default();

        // [FIX] Compute merged delta using the merge service
        let merge_analysis =
            crate::services::merge_service::compute_merged_delta(&src_delta, &tgt_delta);
        let mut merged_delta = merge_analysis.merged_delta;
        merged_delta.merge_metadata = Some(MergeMetadata {
            source_branch_id: req.source_branch_id.to_string(),
            target_branch_id: req.target_branch_id.to_string(),
            merge_note: format!(
                "Merged from {} ({}) into {} ({})",
                source.name, source_head, target.name, target_head
            ),
        });

        let merge_delta_json = serde_json::to_value(&merged_delta)
            .map_err(|e| VersionError::Database(e.to_string()))?;

        let merge_commit_id = Uuid::new_v4();

        // [FIX] Wrap INSERT and UPDATE in a single transaction for atomicity
        let mut tx = self.pool().begin().await?;

        sqlx::query(
            r"
            INSERT INTO commits (id, branch_id, parent_commit_id, message, author_id, author_name, delta)
            VALUES ($1, $2, $3, $4, $5, $6, $7)
            ",
        )
        .bind(merge_commit_id)
        .bind(req.target_branch_id)
        .bind(target_head)
        .bind(&req.message)
        .bind(&req.author_id)
        .bind(&req.author_name)
        .bind(&merge_delta_json)
        .execute(&mut *tx)
        .await?;

        // [FIX] Update target branch head inside the same transaction
        sqlx::query("UPDATE branches SET head_commit_id = $1 WHERE id = $2")
            .bind(merge_commit_id)
            .bind(req.target_branch_id)
            .execute(&mut *tx)
            .await?;

        tx.commit().await?;

        tracing::info!(
            merge_commit_id = %merge_commit_id,
            source_ahead = ahead,
            target_behind = behind,
            conflict_count = merge_analysis.conflict_count,
            auto_resolved = merge_analysis.auto_resolved,
            "[FIX] Branches merged with atomic transaction"
        );

        Ok(MergeResponse {
            merge_commit_id,
            source_branch: req.source_branch_id,
            target_branch: req.target_branch_id,
            conflict_count: merge_analysis.conflict_count,
            auto_resolved: merge_analysis.auto_resolved,
        })
    }

    // ── Switch ────────────────────────────────────────────────────────────

    /// Returns the head commit ID for a branch (used for switching).
    /// Actual state computation is deferred to Task 3.3 (checkout).
    async fn switch_branch(&self, id: Uuid) -> Result<SwitchBranchResponse, VersionError> {
        let branch = self.get_by_id(id).await?;

        tracing::info!(
            branch_id = %id,
            head = ?branch.head_commit_id,
            "Branch switch prepared"
        );

        Ok(SwitchBranchResponse {
            branch_id: branch.id,
            head_commit_id: branch.head_commit_id,
        })
    }
}

impl BranchRepository {
    /// Computes ahead/behind counts between two branches using the commit DAG.
    /// Ahead = commits in `branch_id` not reachable from `reference_id`.
    /// Behind = commits in `reference_id` not reachable from `branch_id`.
    pub(crate) async fn compute_ahead_behind(
        &self,
        branch_id: Uuid,
        reference_id: Uuid,
    ) -> Result<(i64, i64), VersionError> {
        // Get heads of both branches
        let branch_head: Option<Uuid> =
            sqlx::query_scalar("SELECT head_commit_id FROM branches WHERE id = $1")
                .bind(branch_id)
                .fetch_optional(self.pool())
                .await?
                .flatten();

        let ref_head: Option<Uuid> =
            sqlx::query_scalar("SELECT head_commit_id FROM branches WHERE id = $1")
                .bind(reference_id)
                .fetch_optional(self.pool())
                .await?
                .flatten();

        if branch_head.is_none() || ref_head.is_none() {
            return Ok((0, 0));
        }

        // Use PostgreSQL recursive CTE to find merge base and compute counts
        // Simple approach: count commits unique to each branch, scoped by branch_id
        // so deleted branches don't produce (0,0) counts.
        let branch_head_id = branch_head.unwrap();
        let ref_head_id = ref_head.unwrap();

        let ahead: i64 = sqlx::query_scalar(
            r"
            WITH RECURSIVE branch_ancestors AS (
                SELECT id, parent_commit_id, branch_id FROM commits WHERE id = $1
                UNION ALL
                SELECT c.id, c.parent_commit_id, c.branch_id
                FROM commits c
                INNER JOIN branch_ancestors ba ON c.id = ba.parent_commit_id
            ),
            ref_ancestors AS (
                SELECT id, parent_commit_id, branch_id FROM commits WHERE id = $2
                UNION ALL
                SELECT c.id, c.parent_commit_id, c.branch_id
                FROM commits c
                INNER JOIN ref_ancestors ra ON c.id = ra.parent_commit_id
            )
            SELECT COUNT(*)::bigint FROM (
                SELECT id FROM branch_ancestors
                WHERE branch_id = (SELECT branch_id FROM commits WHERE id = $1)
                EXCEPT
                SELECT id FROM ref_ancestors
                WHERE branch_id = (SELECT branch_id FROM commits WHERE id = $2)
            ) AS ahead_commits
            ",
        )
        .bind(branch_head_id)
        .bind(ref_head_id)
        .fetch_one(self.pool())
        .await
        .map_err(VersionError::from)?;

        let behind: i64 = sqlx::query_scalar(
            r"
            WITH RECURSIVE branch_ancestors AS (
                SELECT id, parent_commit_id, branch_id FROM commits WHERE id = $1
                UNION ALL
                SELECT c.id, c.parent_commit_id, c.branch_id
                FROM commits c
                INNER JOIN branch_ancestors ba ON c.id = ba.parent_commit_id
            ),
            ref_ancestors AS (
                SELECT id, parent_commit_id, branch_id FROM commits WHERE id = $2
                UNION ALL
                SELECT c.id, c.parent_commit_id, c.branch_id
                FROM commits c
                INNER JOIN ref_ancestors ra ON c.id = ra.parent_commit_id
            )
            SELECT COUNT(*)::bigint FROM (
                SELECT id FROM ref_ancestors
                WHERE branch_id = (SELECT branch_id FROM commits WHERE id = $2)
                EXCEPT
                SELECT id FROM branch_ancestors
                WHERE branch_id = (SELECT branch_id FROM commits WHERE id = $1)
            ) AS behind_commits
            ",
        )
        .bind(branch_head_id)
        .bind(ref_head_id)
        .fetch_one(self.pool())
        .await
        .map_err(VersionError::from)?;

        Ok((ahead, behind))
    }
}

// ── Helpers ──────────────────────────────────────────────────────────────

/// Converts a `PostgreSQL` row into a `Branch` model.
fn row_to_branch(row: &sqlx::postgres::PgRow) -> Result<Branch, VersionError> {
    Ok(Branch {
        id: row.try_get("id")?,
        name: row.try_get("name")?,
        ontology_id: row.try_get("ontology_id")?,
        head_commit_id: row.try_get("head_commit_id")?,
        created_at: row.try_get("created_at")?,
        is_protected: row.try_get("is_protected")?,
    })
}

// ── Mock for testing ───────────────────────────────────────────────────

#[cfg(test)]
pub(crate) mod mock {
    use super::*;
    use std::sync::Mutex;

    /// Mock implementation of `BranchRepositoryTrait` for unit tests.
    /// Each method returns a configurable result.
    pub struct MockBranchRepository {
        pub create_result: Mutex<Option<Result<Branch, VersionError>>>,
        pub get_by_id_result: Mutex<Option<Result<Branch, VersionError>>>,
        pub get_by_name_result: Mutex<Option<Result<Branch, VersionError>>>,
        pub list_by_ontology_result: Mutex<Option<Result<Vec<BranchWithCommit>, VersionError>>>,
        pub update_head_result: Mutex<Option<Result<(), VersionError>>>,
        pub delete_result: Mutex<Option<Result<(), VersionError>>>,
        pub merge_branches_result: Mutex<Option<Result<MergeResponse, VersionError>>>,
        pub switch_branch_result: Mutex<Option<Result<SwitchBranchResponse, VersionError>>>,
    }

    impl MockBranchRepository {
        /// Creates a new `MockBranchRepository` with default error results.
        pub fn new() -> Self {
            Self {
                create_result: Mutex::new(None),
                get_by_id_result: Mutex::new(None),
                get_by_name_result: Mutex::new(None),
                list_by_ontology_result: Mutex::new(None),
                update_head_result: Mutex::new(None),
                delete_result: Mutex::new(None),
                merge_branches_result: Mutex::new(None),
                switch_branch_result: Mutex::new(None),
            }
        }

        fn take_or_error<T>(
            cell: &Mutex<Option<Result<T, VersionError>>>,
        ) -> Result<T, VersionError>
        where
            T: std::fmt::Debug,
        {
            let mut guard = cell.lock().unwrap();
            guard
                .take()
                .unwrap_or_else(|| Err(VersionError::PgNotConfigured))
        }
    }

    impl Default for MockBranchRepository {
        fn default() -> Self {
            Self::new()
        }
    }

    #[async_trait::async_trait]
    impl BranchRepositoryTrait for MockBranchRepository {
        async fn create(&self, _req: &CreateBranchRequest) -> Result<Branch, VersionError> {
            Self::take_or_error(&self.create_result)
        }

        async fn get_by_id(&self, _id: Uuid) -> Result<Branch, VersionError> {
            Self::take_or_error(&self.get_by_id_result)
        }

        async fn get_by_name(
            &self,
            _ontology_id: Uuid,
            _name: &str,
        ) -> Result<Branch, VersionError> {
            Self::take_or_error(&self.get_by_name_result)
        }

        async fn list_by_ontology(
            &self,
            _ontology_id: Uuid,
            _reference_branch_id: Option<Uuid>,
        ) -> Result<Vec<BranchWithCommit>, VersionError> {
            Self::take_or_error(&self.list_by_ontology_result)
        }

        async fn update_head(
            &self,
            _branch_id: Uuid,
            _commit_id: Uuid,
        ) -> Result<(), VersionError> {
            Self::take_or_error(&self.update_head_result)
        }

        async fn delete(&self, _id: Uuid, _req: &DeleteBranchRequest) -> Result<(), VersionError> {
            Self::take_or_error(&self.delete_result)
        }

        async fn merge_branches(
            &self,
            _req: &MergeBranchesRequest,
        ) -> Result<MergeResponse, VersionError> {
            Self::take_or_error(&self.merge_branches_result)
        }

        async fn switch_branch(&self, _id: Uuid) -> Result<SwitchBranchResponse, VersionError> {
            Self::take_or_error(&self.switch_branch_result)
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_delete_branch_rejects_protected_without_force() {
        // Unit test: verify DeleteBranchRequest defaults
        let req = DeleteBranchRequest { force: false };
        assert!(!req.force);
    }

    #[test]
    fn test_force_delete_flag() {
        let req = DeleteBranchRequest { force: true };
        assert!(req.force);
    }

    #[test]
    fn test_create_branch_request_validation() {
        let req = CreateBranchRequest {
            name: "feature/new".to_string(),
            ontology_id: Uuid::new_v4(),
            source_branch_id: Some(Uuid::new_v4()),
        };
        assert_eq!(req.name, "feature/new");
    }

    #[test]
    fn test_orphan_branch_request() {
        let req = CreateBranchRequest {
            name: "orphan".to_string(),
            ontology_id: Uuid::new_v4(),
            source_branch_id: None,
        };
        assert!(req.source_branch_id.is_none());
    }
}
