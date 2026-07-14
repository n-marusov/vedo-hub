use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

/// Represents the status of a published snapshot.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum SnapshotStatus {
    Creating,
    Published,
    Failed,
    Retired,
}

/// A published snapshot of an ontology at a specific point in time.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Snapshot {
    pub id: String,
    pub ontology_id: String,
    pub branch_id: Option<String>,
    pub commit_id: Option<String>,
    pub status: SnapshotStatus,
    pub created_at: DateTime<Utc>,
    pub size_bytes: i64,
    pub storage_path: String,
    pub format: String,
}

/// Request body for creating a new snapshot.
#[derive(Debug, Deserialize)]
pub struct PublishRequest {
    pub ontology_id: String,
    pub branch_id: Option<String>,
    pub commit_id: Option<String>,
    pub format: Option<String>,
}

/// Response for a snapshot creation request.
#[derive(Debug, Serialize)]
pub struct PublishResponse {
    pub id: String,
    pub ontology_id: String,
    pub status: SnapshotStatus,
    pub created_at: DateTime<Utc>,
    pub storage_path: String,
}

/// Response for listing snapshots.
#[derive(Debug, Serialize)]
pub struct SnapshotListResponse {
    pub snapshots: Vec<SnapshotSummary>,
    pub total: usize,
}

/// Summary view of a snapshot (without full metadata).
#[derive(Debug, Serialize)]
pub struct SnapshotSummary {
    pub id: String,
    pub ontology_id: String,
    pub branch_id: Option<String>,
    pub commit_id: Option<String>,
    pub status: SnapshotStatus,
    pub created_at: DateTime<Utc>,
    pub size_bytes: i64,
    pub format: String,
}

impl From<Snapshot> for SnapshotSummary {
    fn from(s: Snapshot) -> Self {
        SnapshotSummary {
            id: s.id,
            ontology_id: s.ontology_id,
            branch_id: s.branch_id,
            commit_id: s.commit_id,
            status: s.status,
            created_at: s.created_at,
            size_bytes: s.size_bytes,
            format: s.format,
        }
    }
}

/// Error codes used in API responses.
#[derive(Debug, Serialize)]
pub struct ApiErrorResponse {
    pub error: String,
    pub message: String,
}
