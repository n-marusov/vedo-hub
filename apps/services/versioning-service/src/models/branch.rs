//! Branch domain model and request/response types.
//!
//! A branch represents an independent line of development within an ontology.
//! Each branch has its own commit history and a head pointer.

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

// ── Domain Model ───────────────────────────────────────────────────────────

/// A branch in the versioning system.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Branch {
    pub id: Uuid,
    pub name: String,
    pub ontology_id: Uuid,
    pub head_commit_id: Option<Uuid>,
    pub created_at: DateTime<Utc>,
    pub is_protected: bool,
}

/// A branch with its latest commit information (for list displays).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BranchWithCommit {
    pub id: Uuid,
    pub name: String,
    pub ontology_id: Uuid,
    pub head_commit_id: Option<Uuid>,
    pub created_at: DateTime<Utc>,
    pub is_protected: bool,
    /// Latest commit message (if a commit exists).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub last_commit_message: Option<String>,
    /// Latest commit author name.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub last_commit_author: Option<String>,
    /// Latest commit timestamp.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub last_commit_at: Option<DateTime<Utc>>,
    /// Number of commits the branch is ahead of the default branch.
    #[serde(default)]
    pub ahead_count: i64,
    /// Number of commits the branch is behind the default branch.
    #[serde(default)]
    pub behind_count: i64,
}

/// Summary view of a branch (lightweight, no commit info).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BranchSummary {
    pub id: Uuid,
    pub name: String,
    pub ontology_id: Uuid,
    pub head_commit_id: Option<Uuid>,
    pub created_at: DateTime<Utc>,
    pub is_protected: bool,
}

impl From<Branch> for BranchSummary {
    fn from(b: Branch) -> Self {
        Self {
            id: b.id,
            name: b.name,
            ontology_id: b.ontology_id,
            head_commit_id: b.head_commit_id,
            created_at: b.created_at,
            is_protected: b.is_protected,
        }
    }
}

// ── Request / Response Types ───────────────────────────────────────────────

/// Request body for creating a new branch.
#[derive(Debug, Deserialize)]
pub struct CreateBranchRequest {
    pub name: String,
    pub ontology_id: Uuid,
    /// Source branch ID to fork from. If omitted, creates an orphan branch.
    pub source_branch_id: Option<Uuid>,
}

/// Request body for deleting a branch.
#[derive(Debug, Deserialize)]
pub struct DeleteBranchRequest {
    /// Set to `true` to force-delete a protected branch.
    #[serde(default)]
    pub force: bool,
}

/// Request body for merging two branches.
#[derive(Debug, Deserialize)]
pub struct MergeBranchesRequest {
    pub source_branch_id: Uuid,
    pub target_branch_id: Uuid,
    pub message: String,
    pub author_id: String,
    pub author_name: String,
}

/// Response for a branch merge operation.
#[derive(Debug, Serialize)]
pub struct MergeResponse {
    pub merge_commit_id: Uuid,
    pub source_branch: Uuid,
    pub target_branch: Uuid,
    pub conflict_count: usize,
    pub auto_resolved: bool,
}

/// Response for switching a branch.
#[derive(Debug, Serialize)]
pub struct SwitchBranchResponse {
    pub branch_id: Uuid,
    pub head_commit_id: Option<Uuid>,
}

/// Paginated response wrapper (re-exported for convenience).
#[derive(Debug, Serialize)]
pub struct PaginatedBranchesResponse<T: Serialize> {
    pub items: Vec<T>,
    pub total: u64,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_branch_summary_from_branch() {
        let branch = Branch {
            id: Uuid::new_v4(),
            name: "main".to_string(),
            ontology_id: Uuid::new_v4(),
            head_commit_id: None,
            created_at: Utc::now(),
            is_protected: true,
        };
        let summary: BranchSummary = branch.into();
        assert_eq!(summary.name, "main");
        assert!(summary.is_protected);
    }

    #[test]
    fn test_create_branch_request_serde() {
        let json = serde_json::json!({
            "name": "feature/test",
            "ontology_id": Uuid::new_v4().to_string(),
            "source_branch_id": Uuid::new_v4().to_string(),
        });
        let req: CreateBranchRequest = serde_json::from_value(json).unwrap();
        assert_eq!(req.name, "feature/test");
        assert!(req.source_branch_id.is_some());
    }

    #[test]
    fn test_create_branch_request_no_source() {
        let json = serde_json::json!({
            "name": "orphan-branch",
            "ontology_id": Uuid::new_v4().to_string(),
        });
        let req: CreateBranchRequest = serde_json::from_value(json).unwrap();
        assert!(req.source_branch_id.is_none());
    }

    #[test]
    fn test_delete_branch_request_default_force_false() {
        let req: DeleteBranchRequest = serde_json::from_str(r#"{}"#).unwrap();
        assert!(!req.force);
    }

    #[test]
    fn test_delete_branch_request_force_true() {
        let req: DeleteBranchRequest = serde_json::from_str(r#"{"force": true}"#).unwrap();
        assert!(req.force);
    }

    #[test]
    fn test_merge_branches_request_serde() {
        let json = serde_json::json!({
            "source_branch_id": Uuid::new_v4().to_string(),
            "target_branch_id": Uuid::new_v4().to_string(),
            "message": "Merge feature",
            "author_id": "user1",
            "author_name": "User One",
        });
        let req: MergeBranchesRequest = serde_json::from_value(json).unwrap();
        assert_eq!(req.message, "Merge feature");
    }

    #[test]
    fn test_merge_response_serde() {
        let resp = MergeResponse {
            merge_commit_id: Uuid::new_v4(),
            source_branch: Uuid::new_v4(),
            target_branch: Uuid::new_v4(),
            conflict_count: 2,
            auto_resolved: true,
        };
        let json = serde_json::to_value(&resp).unwrap();
        assert_eq!(json["conflict_count"], 2);
        assert_eq!(json["auto_resolved"], true);
    }
}
