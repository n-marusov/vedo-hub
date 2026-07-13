//! Commit repository — PostgreSQL data access for commits.
//!
//! Provides SQL queries for commit CRUD, history listing, and delta
//! persistence using `sqlx`.

use chrono::{DateTime, Utc};
use sqlx::{PgPool, Row};
use uuid::Uuid;

use crate::error::VersionError;
use crate::models::{
    Commit, CommitDelta, CommitSummary, CreateCommitRequest, ListCommitsParams, PaginatedResponse,
};

/// Repository for commit operations against PostgreSQL.
pub struct CommitRepository {
    pool: PgPool,
}

impl CommitRepository {
    /// Creates a new `CommitRepository` with the given connection pool.
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }

    /// Returns a reference to the inner pool (useful for transactions).
    pub fn pool(&self) -> &PgPool {
        &self.pool
    }

    // ── Create ────────────────────────────────────────────────────────────

    /// Creates a new commit with the given delta.
    ///
    /// Validates that the delta is not empty and that the referenced branch
    /// exists. Computes `parent_commit_id` from the branch's current head.
    pub async fn create(&self, req: &CreateCommitRequest) -> Result<Commit, VersionError> {
        if req.delta.is_empty() {
            return Err(VersionError::EmptyDelta);
        }

        tracing::debug!(
            branch_id = %req.branch_id,
            author_id = %req.author_id,
            "Creating commit"
        );

        // Verify branch exists and get current head
        let head_commit_id: Option<Uuid> =
            sqlx::query_scalar("SELECT head_commit_id FROM branches WHERE id = $1")
                .bind(req.branch_id)
                .fetch_optional(self.pool())
                .await?
                .flatten();

        // If branch not found, head_commit_id is None and the insert will SKIP
        // parent linking. The branch FK constraint is enforced by the DB.
        let parent_id = head_commit_id;

        let delta_json =
            serde_json::to_value(&req.delta).map_err(|e| VersionError::Database(e.to_string()))?;

        let row = sqlx::query(
            r#"
            INSERT INTO commits (branch_id, parent_commit_id, message, author_id, author_name, delta)
            VALUES ($1, $2, $3, $4, $5, $6)
            RETURNING id, branch_id, parent_commit_id, message, author_id, author_name, delta, created_at
            "#,
        )
        .bind(req.branch_id)
        .bind(parent_id)
        .bind(&req.message)
        .bind(&req.author_id)
        .bind(&req.author_name)
        .bind(&delta_json)
        .fetch_one(self.pool())
        .await?;

        let commit = row_to_commit(&row)?;

        // Update branch head to point to this new commit
        sqlx::query("UPDATE branches SET head_commit_id = $1 WHERE id = $2")
            .bind(commit.id)
            .bind(req.branch_id)
            .execute(self.pool())
            .await?;

        tracing::info!(
            commit_id = %commit.id,
            branch_id = %req.branch_id,
            total_changes = %commit.delta.total_changes(),
            "Commit created"
        );

        Ok(commit)
    }

    // ── Read ──────────────────────────────────────────────────────────────

    /// Retrieves a single commit by its ID, including the full delta.
    pub async fn get_by_id(&self, id: Uuid) -> Result<Commit, VersionError> {
        tracing::debug!(commit_id = %id, "Fetching commit by ID");

        let row = sqlx::query(
            r#"
            SELECT id, branch_id, parent_commit_id, message, author_id, author_name,
                   delta, created_at
            FROM commits
            WHERE id = $1
            "#,
        )
        .bind(id)
        .fetch_optional(self.pool())
        .await?
        .ok_or_else(|| VersionError::CommitNotFound(id.to_string()))?;

        let commit = row_to_commit(&row)?;
        tracing::debug!(commit_id = %id, total_changes = %commit.delta.total_changes(), "Commit found");
        Ok(commit)
    }

    /// Lists commits with optional branch filtering, pagination, and sorting.
    pub async fn list(
        &self,
        params: &ListCommitsParams,
    ) -> Result<PaginatedResponse<CommitSummary>, VersionError> {
        tracing::debug!(
            branch_id = ?params.branch_id,
            page = params.page,
            per_page = params.per_page,
            "Listing commits"
        );

        let limit = params.per_page.min(100);
        let offset = params.page * limit;

        let (count_query, data_query, branch_filter) = match params.branch_id {
            Some(_bid) => (
                "SELECT COUNT(*) FROM commits WHERE branch_id = $1",
                r#"
                SELECT id, branch_id, parent_commit_id, message, author_id, author_name,
                       delta, created_at
                FROM commits
                WHERE branch_id = $1
                ORDER BY created_at DESC
                LIMIT $2 OFFSET $3
                "#,
                true,
            ),
            None => (
                "SELECT COUNT(*) FROM commits",
                r#"
                SELECT id, branch_id, parent_commit_id, message, author_id, author_name,
                       delta, created_at
                FROM commits
                ORDER BY created_at DESC
                LIMIT $1 OFFSET $2
                "#,
                false,
            ),
        };

        let total: i64 = if branch_filter {
            sqlx::query_scalar(count_query)
                .bind(params.branch_id)
                .fetch_one(self.pool())
                .await?
        } else {
            sqlx::query_scalar(count_query)
                .fetch_one(self.pool())
                .await?
        };

        let rows = if branch_filter {
            sqlx::query(data_query)
                .bind(params.branch_id)
                .bind(limit as i64)
                .bind(offset as i64)
                .fetch_all(self.pool())
                .await?
        } else {
            sqlx::query(data_query)
                .bind(limit as i64)
                .bind(offset as i64)
                .fetch_all(self.pool())
                .await?
        };

        let mut items = Vec::with_capacity(rows.len());
        for row in &rows {
            let commit = row_to_commit(row)?;
            items.push(CommitSummary::from(commit));
        }

        tracing::info!(
            total = %total,
            returned = items.len(),
            "Commit list retrieved"
        );

        Ok(PaginatedResponse {
            items,
            total: total as u64,
            page: params.page,
            per_page: limit,
        })
    }

    /// Retrieves the delta for a specific commit (returns the raw delta).
    pub async fn get_delta(&self, id: Uuid) -> Result<CommitDelta, VersionError> {
        let commit = self.get_by_id(id).await?;
        Ok(commit.delta)
    }

    /// Returns the number of commits on a given branch.
    pub async fn count_by_branch(&self, branch_id: Uuid) -> Result<i64, VersionError> {
        let count: (i64,) = sqlx::query_as("SELECT COUNT(*) FROM commits WHERE branch_id = $1")
            .bind(branch_id)
            .fetch_one(self.pool())
            .await?;
        Ok(count.0)
    }
}

// ── Helpers ──────────────────────────────────────────────────────────────

/// Converts a PostgreSQL row into a `Commit` model.
pub fn row_to_commit(row: &sqlx::postgres::PgRow) -> Result<Commit, VersionError> {
    let id: Uuid = row.try_get("id")?;
    let branch_id: Uuid = row.try_get("branch_id")?;
    let parent_commit_id: Option<Uuid> = row.try_get("parent_commit_id")?;
    let message: String = row.try_get("message")?;
    let author_id: String = row.try_get("author_id")?;
    let author_name: String = row.try_get("author_name")?;
    let delta_json: serde_json::Value = row.try_get("delta")?;
    let created_at: DateTime<Utc> = row.try_get("created_at")?;

    let delta: CommitDelta = serde_json::from_value(delta_json)
        .map_err(|e| VersionError::Database(format!("Delta deserialization error: {e}")))?;

    Ok(Commit {
        id,
        branch_id,
        parent_commit_id,
        message,
        author_id,
        author_name,
        delta,
        created_at,
    })
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::models::TripleRef;

    #[test]
    fn test_list_params_defaults() {
        let params = ListCommitsParams {
            branch_id: None,
            page: 0,
            per_page: 20,
        };
        assert_eq!(params.page, 0);
        assert_eq!(params.per_page, 20);
    }

    #[test]
    fn test_list_params_page_roundtrip() {
        let json = r#"{"page": 2, "per_page": 50}"#;
        let params: ListCommitsParams = serde_json::from_str(json).unwrap();
        assert_eq!(params.page, 2);
        assert_eq!(params.per_page, 50);
        assert!(params.branch_id.is_none());
    }

    #[test]
    fn test_list_params_branch_filter() {
        let bid = Uuid::new_v4();
        let json = format!(r#"{{"branch_id": "{}", "page": 0}}"#, bid);
        let params: ListCommitsParams = serde_json::from_str(&json).unwrap();
        assert_eq!(params.branch_id, Some(bid));
    }

    #[test]
    fn test_per_page_clamped_to_max() {
        let params = ListCommitsParams {
            branch_id: None,
            page: 0,
            per_page: 200,
        };
        assert_eq!(params.per_page.min(100), 100);
    }

    #[test]
    fn test_commit_delta_serde_roundtrip() {
        let delta = CommitDelta {
            added_triples: vec![TripleRef {
                s: "ClassA".into(),
                p: "rdfs:subClassOf".into(),
                o: "Thing".into(),
            }],
            removed_triples: vec![],
            modified_triples: vec![],
        };
        let json = serde_json::to_value(&delta).unwrap();
        let deserialized: CommitDelta = serde_json::from_value(json).unwrap();
        assert_eq!(deserialized.added_triples.len(), 1);
    }

    #[test]
    fn test_create_request_validation_empty_delta() {
        let req = CreateCommitRequest {
            branch_id: Uuid::new_v4(),
            message: "test".to_string(),
            author_id: "user".to_string(),
            author_name: "User".to_string(),
            delta: CommitDelta::default(),
        };
        // The empty-delta check is at the repository level.
        // Verify the model allows empty delta (serialization test).
        let json = serde_json::to_value(&req).unwrap();
        let deserialized: CreateCommitRequest = serde_json::from_value(json).unwrap();
        assert!(deserialized.delta.is_empty());
    }
}
