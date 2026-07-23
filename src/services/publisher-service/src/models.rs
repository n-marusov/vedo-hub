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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_snapshot_status_serde() {
        let status: SnapshotStatus = serde_json::from_str("\"creating\"").unwrap();
        assert_eq!(status, SnapshotStatus::Creating);

        let published: SnapshotStatus = serde_json::from_str("\"published\"").unwrap();
        assert_eq!(published, SnapshotStatus::Published);

        let failed: SnapshotStatus = serde_json::from_str("\"failed\"").unwrap();
        assert_eq!(failed, SnapshotStatus::Failed);

        let retired: SnapshotStatus = serde_json::from_str("\"retired\"").unwrap();
        assert_eq!(retired, SnapshotStatus::Retired);
    }

    #[test]
    fn test_publish_request_minimal() {
        let req: PublishRequest = serde_json::from_str(r#"{"ontology_id":"onto-1"}"#).unwrap();
        assert_eq!(req.ontology_id, "onto-1");
        assert_eq!(req.branch_id, None);
        assert_eq!(req.commit_id, None);
        assert_eq!(req.format, None);
    }

    #[test]
    fn test_publish_request_full() {
        let req: PublishRequest = serde_json::from_str(
            r#"{"ontology_id":"onto-1","branch_id":"main","commit_id":"abc","format":"turtle"}"#,
        )
        .unwrap();
        assert_eq!(req.ontology_id, "onto-1");
        assert_eq!(req.branch_id, Some("main".to_string()));
        assert_eq!(req.format, Some("turtle".to_string()));
    }

    #[test]
    fn test_snapshot_to_summary_conversion() {
        let snap = Snapshot {
            id: "snap-1".to_string(),
            ontology_id: "onto-1".to_string(),
            branch_id: Some("main".to_string()),
            commit_id: Some("abc".to_string()),
            status: SnapshotStatus::Published,
            created_at: DateTime::from_timestamp_nanos(0),
            size_bytes: 1024,
            storage_path: "/data/snap-1.ttl".to_string(),
            format: "turtle".to_string(),
        };
        let summary: SnapshotSummary = snap.into();
        assert_eq!(summary.id, "snap-1");
        assert_eq!(summary.ontology_id, "onto-1");
        assert_eq!(summary.status, SnapshotStatus::Published);
        assert_eq!(summary.size_bytes, 1024);
    }

    #[test]
    fn test_api_error_response_serialization() {
        let err = ApiErrorResponse {
            error: "SNAPSHOT_NOT_FOUND".to_string(),
            message: "Snapshot not found".to_string(),
        };
        let json = serde_json::to_string(&err).unwrap();
        assert!(json.contains("SNAPSHOT_NOT_FOUND"));
        assert!(json.contains("Snapshot not found"));
    }
}
