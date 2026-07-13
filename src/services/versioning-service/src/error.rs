//! Domain-specific error types for the versioning service.
//!
//! Uses `thiserror` for ergonomic error definitions and implements axum's
//! `IntoResponse` so errors can be returned directly from handlers.

use axum::{http::StatusCode, response::IntoResponse, Json};
use thiserror::Error;

/// Errors that can occur during versioning operations.
#[derive(Debug, Error)]
pub enum VersionError {
    #[error("Commit not found: {0}")]
    CommitNotFound(String),

    #[error("Branch not found: {0}")]
    BranchNotFound(String),

    #[error("Cannot create commit with empty delta")]
    EmptyDelta,

    #[error("Commit conflict: {0}")]
    CommitConflict(String),

    #[error("Database error: {0}")]
    Database(String),

    #[error("PostgreSQL database is not configured")]
    PgNotConfigured,

    #[error("Invalid request: {0}")]
    InvalidRequest(String),

    #[error("Branch {branch} is protected")]
    BranchProtected { branch: String },

    #[error("Merge conflict: {details}")]
    MergeConflict { details: String },
}

impl IntoResponse for VersionError {
    fn into_response(self) -> axum::response::Response {
        let (status, code) = match &self {
            VersionError::CommitNotFound(_) => (StatusCode::NOT_FOUND, "VER-COMMIT-NOT-FOUND"),
            VersionError::BranchNotFound(_) => (StatusCode::NOT_FOUND, "VER-BRANCH-NOT-FOUND"),
            VersionError::EmptyDelta => (StatusCode::BAD_REQUEST, "VER-COMMIT-EMPTY-DELTA"),
            VersionError::CommitConflict(_) => (StatusCode::CONFLICT, "VER-COMMIT-CONFLICT"),
            VersionError::Database(_) => (StatusCode::INTERNAL_SERVER_ERROR, "VER-DATABASE-ERROR"),
            VersionError::PgNotConfigured => {
                (StatusCode::SERVICE_UNAVAILABLE, "VER-PG-NOT-CONFIGURED")
            }
            VersionError::InvalidRequest(_) => (StatusCode::BAD_REQUEST, "VER-INVALID-REQUEST"),
            VersionError::BranchProtected { .. } => (StatusCode::FORBIDDEN, "VER-BRANCH-PROTECTED"),
            VersionError::MergeConflict { .. } => (StatusCode::CONFLICT, "VER-MERGE-CONFLICT"),
        };
        let body = serde_json::json!({
            "error": code,
            "detail": self.to_string(),
        });
        (status, Json(body)).into_response()
    }
}

/// Helper to map `sqlx::Error` into our domain error type.
impl From<sqlx::Error> for VersionError {
    fn from(err: sqlx::Error) -> Self {
        tracing::error!(error = %err, "SQLx error");
        VersionError::Database(err.to_string())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use axum::http::StatusCode;

    #[tokio::test]
    async fn test_commit_not_found_response() {
        let err = VersionError::CommitNotFound("abc-123".to_string());
        let resp = err.into_response();
        assert_eq!(resp.status(), StatusCode::NOT_FOUND);
    }

    #[tokio::test]
    async fn test_empty_delta_response() {
        let err = VersionError::EmptyDelta;
        let resp = err.into_response();
        assert_eq!(resp.status(), StatusCode::BAD_REQUEST);
    }

    #[tokio::test]
    async fn test_pg_not_configured_response() {
        let err = VersionError::PgNotConfigured;
        let resp = err.into_response();
        assert_eq!(resp.status(), StatusCode::SERVICE_UNAVAILABLE);
    }

    #[test]
    fn test_error_messages() {
        assert_eq!(
            VersionError::CommitNotFound("id".to_string()).to_string(),
            "Commit not found: id"
        );
        assert_eq!(
            VersionError::EmptyDelta.to_string(),
            "Cannot create commit with empty delta"
        );
        assert_eq!(
            VersionError::BranchProtected {
                branch: "main".to_string()
            }
            .to_string(),
            "Branch main is protected"
        );
    }
}
