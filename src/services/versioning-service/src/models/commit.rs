//! Commit domain model, delta format, and request/response types.
//!
//! A commit represents a snapshot of ontology changes at a point in time.
//! The delta captures the diff between the previous state and the new state.

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

// ── Delta Types ────────────────────────────────────────────────────────────

/// A single RDF triple reference used in delta entries.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct TripleRef {
    /// Subject IRI or local ID.
    pub s: String,
    /// Predicate IRI or property ID.
    pub p: String,
    /// Object IRI, local ID, or literal value.
    pub o: String,
}

/// A triple modification entry (before → after).
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct ModifiedTriple {
    /// Subject IRI or local ID.
    pub s: String,
    /// Predicate IRI or property ID.
    pub p: String,
    /// Old object value before modification.
    pub old_o: String,
    /// New object value after modification.
    pub new_o: String,
}

/// The delta payload stored as JSONB in PostgreSQL.
///
/// ```json
/// {
///   "added_triples": [{"s": "Person", "p": "rdfs:label", "o": "Person"}],
///   "removed_triples": [],
///   "modified_triples": []
/// }
/// ```
#[derive(Debug, Clone, Default, Serialize, Deserialize, PartialEq)]
pub struct CommitDelta {
    /// Triples added in this commit.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub added_triples: Vec<TripleRef>,

    /// Triples removed in this commit.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub removed_triples: Vec<TripleRef>,

    /// Triples whose object value changed.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub modified_triples: Vec<ModifiedTriple>,
}

impl CommitDelta {
    /// Returns `true` when the delta contains no changes.
    pub fn is_empty(&self) -> bool {
        self.added_triples.is_empty()
            && self.removed_triples.is_empty()
            && self.modified_triples.is_empty()
    }

    /// Returns the total number of triple changes in this delta.
    pub fn total_changes(&self) -> usize {
        self.added_triples.len() + self.removed_triples.len() + self.modified_triples.len()
    }
}

// ── Domain Model ───────────────────────────────────────────────────────────

/// A commit in the version history — represents a point-in-time snapshot
/// of ontology changes stored as a delta.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Commit {
    /// Unique commit identifier.
    pub id: Uuid,
    /// The branch this commit belongs to.
    pub branch_id: Uuid,
    /// Optional parent commit ID (null for root commits).
    pub parent_commit_id: Option<Uuid>,
    /// Commit message describing the changes.
    pub message: String,
    /// External user/author identifier.
    pub author_id: String,
    /// Display name of the author.
    pub author_name: String,
    /// The delta payload — added, removed, modified triples.
    pub delta: CommitDelta,
    /// Timestamp when the commit was created.
    pub created_at: DateTime<Utc>,
}

/// Summary view of a commit for list endpoints (excludes full delta).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CommitSummary {
    pub id: Uuid,
    pub branch_id: Uuid,
    pub parent_commit_id: Option<Uuid>,
    pub message: String,
    pub author_id: String,
    pub author_name: String,
    pub total_changes: usize,
    pub created_at: DateTime<Utc>,
}

impl From<Commit> for CommitSummary {
    fn from(c: Commit) -> Self {
        let total_changes = c.delta.total_changes();
        Self {
            id: c.id,
            branch_id: c.branch_id,
            parent_commit_id: c.parent_commit_id,
            message: c.message,
            author_id: c.author_id,
            author_name: c.author_name,
            total_changes,
            created_at: c.created_at,
        }
    }
}

// ── Request / Response Types ───────────────────────────────────────────────

/// Request body for creating a new commit.
#[derive(Debug, Serialize, Deserialize)]
pub struct CreateCommitRequest {
    pub branch_id: Uuid,
    pub message: String,
    pub author_id: String,
    pub author_name: String,
    #[serde(default)]
    pub delta: CommitDelta,
}

/// Query parameters for listing commits.
#[derive(Debug, Deserialize)]
pub struct ListCommitsParams {
    /// Filter by branch ID (optional).
    pub branch_id: Option<Uuid>,
    /// Zero-based page offset (default: 0).
    #[serde(default = "default_page")]
    pub page: u64,
    /// Items per page (default: 20, max: 100).
    #[serde(default = "default_per_page")]
    pub per_page: u64,
}

fn default_page() -> u64 {
    0
}
fn default_per_page() -> u64 {
    20
}

/// Paginated response wrapper.
#[derive(Debug, Serialize)]
pub struct PaginatedResponse<T: Serialize> {
    pub items: Vec<T>,
    pub total: u64,
    pub page: u64,
    pub per_page: u64,
}

/// Response for commit detail — includes full delta with preview limit.
#[derive(Debug, Serialize)]
pub struct CommitDetailResponse {
    pub id: Uuid,
    pub branch_id: Uuid,
    pub parent_commit_id: Option<Uuid>,
    pub message: String,
    pub author_id: String,
    pub author_name: String,
    pub delta: CommitDeltaPreview,
    pub created_at: DateTime<Utc>,
}

/// A preview of the delta (first 10 triples of each section).
#[derive(Debug, Serialize)]
pub struct CommitDeltaPreview {
    /// Total added triples in the full delta.
    pub added_total: usize,
    /// Preview of added triples (first 10).
    pub added_preview: Vec<TripleRef>,
    /// Total removed triples in the full delta.
    pub removed_total: usize,
    /// Preview of removed triples (first 10).
    pub removed_preview: Vec<TripleRef>,
    /// Total modified triples in the full delta.
    pub modified_total: usize,
    /// Preview of modified triples (first 10).
    pub modified_preview: Vec<ModifiedTriple>,
}

const DELTA_PREVIEW_LIMIT: usize = 10;

impl From<Commit> for CommitDetailResponse {
    fn from(c: Commit) -> Self {
        let delta_preview = CommitDeltaPreview {
            added_total: c.delta.added_triples.len(),
            added_preview: c
                .delta
                .added_triples
                .into_iter()
                .take(DELTA_PREVIEW_LIMIT)
                .collect(),
            removed_total: c.delta.removed_triples.len(),
            removed_preview: c
                .delta
                .removed_triples
                .into_iter()
                .take(DELTA_PREVIEW_LIMIT)
                .collect(),
            modified_total: c.delta.modified_triples.len(),
            modified_preview: c
                .delta
                .modified_triples
                .into_iter()
                .take(DELTA_PREVIEW_LIMIT)
                .collect(),
        };
        Self {
            id: c.id,
            branch_id: c.branch_id,
            parent_commit_id: c.parent_commit_id,
            message: c.message,
            author_id: c.author_id,
            author_name: c.author_name,
            delta: delta_preview,
            created_at: c.created_at,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_empty_delta_is_empty() {
        let delta = CommitDelta::default();
        assert!(delta.is_empty());
        assert_eq!(delta.total_changes(), 0);
    }

    #[test]
    fn test_non_empty_delta() {
        let delta = CommitDelta {
            added_triples: vec![TripleRef {
                s: "Person".to_string(),
                p: "rdfs:label".to_string(),
                o: "Person".to_string(),
            }],
            removed_triples: vec![],
            modified_triples: vec![],
        };
        assert!(!delta.is_empty());
        assert_eq!(delta.total_changes(), 1);
    }

    #[test]
    fn test_commit_summary_from_commit() {
        let commit = Commit {
            id: Uuid::new_v4(),
            branch_id: Uuid::new_v4(),
            parent_commit_id: None,
            message: "Initial commit".to_string(),
            author_id: "user-1".to_string(),
            author_name: "Alice".to_string(),
            delta: CommitDelta {
                added_triples: vec![
                    TripleRef {
                        s: "A".into(),
                        p: "B".into(),
                        o: "C".into(),
                    },
                    TripleRef {
                        s: "D".into(),
                        p: "E".into(),
                        o: "F".into(),
                    },
                ],
                removed_triples: vec![],
                modified_triples: vec![],
            },
            created_at: Utc::now(),
        };
        let summary: CommitSummary = commit.into();
        assert_eq!(summary.total_changes, 2);
        assert_eq!(summary.message, "Initial commit");
    }

    #[test]
    fn test_commit_detail_response_preview_limit() {
        let added: Vec<TripleRef> = (0..15)
            .map(|i| TripleRef {
                s: format!("s{i}"),
                p: "p".to_string(),
                o: format!("o{i}"),
            })
            .collect();
        let commit = Commit {
            id: Uuid::new_v4(),
            branch_id: Uuid::new_v4(),
            parent_commit_id: None,
            message: "Large commit".to_string(),
            author_id: "user-1".to_string(),
            author_name: "Bob".to_string(),
            delta: CommitDelta {
                added_triples: added,
                removed_triples: vec![],
                modified_triples: vec![],
            },
            created_at: Utc::now(),
        };
        let detail: CommitDetailResponse = commit.into();
        assert_eq!(detail.delta.added_total, 15);
        assert_eq!(detail.delta.added_preview.len(), 10);
    }

    #[test]
    fn test_commit_delta_serde_roundtrip() {
        let delta = CommitDelta {
            added_triples: vec![TripleRef {
                s: "Person".into(),
                p: "rdfs:label".into(),
                o: "Person".into(),
            }],
            removed_triples: vec![],
            modified_triples: vec![],
        };
        let json = serde_json::to_string(&delta).unwrap();
        let deserialized: CommitDelta = serde_json::from_str(&json).unwrap();
        assert_eq!(delta, deserialized);
    }
}
