//! Unit tests for the GraphQL facade assembled by `ontology-service`.
//!
//! These tests pin down the M1 contract that was added to close the
//! `$aif-verify` gap on Task 8.1:
//!
//! 1. `build_schema()` wires `MutationRoot` (NOT `EmptyMutation`) so the
//!    frontend `UPDATE_DRAFT_MUTATION` validates against the schema.
//! 2. `VersioningClient` correctly parses the REST payload shapes returned
//!    by the versioning-service (`RemoteCommitPage`, `RemoteBranchPage`),
//!    including `Option`al commit metadata and `ahead/behind` defaults.
//! 3. The `branch_into_gql` mapping helper preserves all fields documented
//!    by the GraphQL `Branch` type, including `Option` handling.
//!
//! The tests intentionally avoid the network: the schema test executes a
//! GraphQL document against an in-memory schema, and the client/mapping
//! tests feed hand-rolled JSON through `serde_json::from_value` so they
//! remain deterministic and fast.

use chrono::{TimeZone, Utc};
use serde_json::json;
use uuid::Uuid;

use super::mutation::MutationRoot;
use super::query::branch_into_gql;
use super::schema::{build_schema, OntologySchema};
use super::types::{GqlBranch, GqlCommit};
use super::versioning_client::{
    RemoteBranch, RemoteBranchPage, RemoteCommitPage, RemoteCommitSummary,
};

// ── Schema construction ──────────────────────────────────────────────────────

/// build_schema() must return a schema whose mutation root is MutationRoot,
/// not EmptyMutation. The frontend relies on `updateDraft` existing in the
/// schema; if a future refactor accidentally swaps back to EmptyMutation this
/// test fails before the change ships.
#[tokio::test]
async fn test_build_schema_wires_mutation_root() {
    let schema: OntologySchema = build_schema();
    // Execute the updateDraft mutation against an in-memory schema. EmptyMutation
    // would return "Cannot query field \"updateDraft\" on type \"Mutation\"" —
    // we assert success instead so the contract is observable.
    use async_graphql::{Name, Variables};
    let mut vars = Variables::default();
    vars.insert(
        Name::new("ontologyId"),
        async_graphql::value!("00000000-0000-0000-0000-000000000000"),
    );
    vars.insert(
        Name::new("changes"),
        async_graphql::value!({ "changes": "{}" }),
    );

    let response = schema
        .execute(
            async_graphql::Request::new(
                r#"mutation UpdateDraft($ontologyId: ID!, $changes: DraftInput!) {
                    updateDraft(ontologyId: $ontologyId, changes: $changes) {
                        success
                        timestamp
                    }
                }"#,
            )
            .variables(vars),
        )
        .await;

    assert!(
        response.errors.is_empty(),
        "expected no GraphQL errors from updateDraft, got: {:?}",
        response.errors
    );
    let data = response.data.into_json().expect("response data is JSON");
    assert_eq!(data["updateDraft"]["success"], true);
    // Timestamp is RFC3339 — just assert it is present and non-empty.
    assert!(
        !data["updateDraft"]["timestamp"]
            .as_str()
            .unwrap_or("")
            .is_empty(),
        "expected non-empty timestamp from updateDraft"
    );
}

/// Sanity check: the MutationRoot type is constructed with a default and the
/// schema type assembles without panicking. This catches regressions where a
/// removed resolver method leaves the mutation root incompatible with
/// `Schema::build`.
#[test]
fn test_mutation_root_default_is_constructible() {
    let _ = MutationRoot::default();
    let _schema = build_schema();
}

// ── VersioningClient payload parsing ─────────────────────────────────────────

/// `RemoteCommitPage` must deserialize from the response shape produced by
/// `GET /api/v1/versioning/commits`. The versioning-service encodes UUIDs as
/// strings, optional parent commit IDs as null, and `total_changes` as a number.
#[test]
fn test_remote_commit_page_parses_versioning_service_shape() {
    let commit_id = Uuid::new_v4();
    let branch_id = Uuid::new_v4();
    let created_at = Utc.with_ymd_and_hms(2026, 7, 14, 12, 0, 0).unwrap();

    let payload = json!({
        "items": [{
            "id": commit_id,
            "branch_id": branch_id,
            "parent_commit_id": null,
            "message": "Add Person class",
            "author_id": "user-1",
            "author_name": "Alice",
            "total_changes": 3,
            "created_at": created_at.to_rfc3339(),
        }],
        "total": 1,
        "page": 0,
        "per_page": 20,
    });

    let page: RemoteCommitPage = serde_json::from_value(payload).expect("parse commit page");
    assert_eq!(page.total, 1);
    assert_eq!(page.page, 0);
    assert_eq!(page.per_page, 20);
    assert_eq!(page.items.len(), 1);
    let c: &RemoteCommitSummary = &page.items[0];
    assert_eq!(c.id, commit_id);
    assert_eq!(c.branch_id, branch_id);
    assert!(
        c.parent_commit_id.is_none(),
        "parent commit must be Option::None"
    );
    assert_eq!(c.message, "Add Person class");
    assert_eq!(c.author_id, "user-1");
    assert_eq!(c.author_name, "Alice");
    assert_eq!(c.total_changes, 3);
    assert_eq!(c.created_at, created_at);
}

/// `RemoteBranchPage` must deserialize from the versioning-service branch
/// response, with `Option`al commit metadata and defaulted `ahead/behind`.
#[test]
fn test_remote_branch_page_parses_with_optionals_and_defaults() {
    let branch_id = Uuid::new_v4();
    let ontology_id = Uuid::new_v4();
    let created_at = Utc.with_ymd_and_hms(2026, 7, 14, 10, 30, 0).unwrap();

    // Payload omits last_commit_*. The #[serde(default)] attributes must fill
    // them with None / 0 so resolvers do not need to special-case missing keys.
    let payload = json!({
        "items": [{
            "id": branch_id,
            "name": "main",
            "ontology_id": ontology_id,
            "head_commit_id": null,
            "created_at": created_at.to_rfc3339(),
            "is_protected": true,
        }],
        "total": 1,
    });

    let page: RemoteBranchPage = serde_json::from_value(payload).expect("parse branch page");
    assert_eq!(page.total, 1);
    assert_eq!(page.items.len(), 1);
    let b: &RemoteBranch = &page.items[0];
    assert_eq!(b.id, branch_id);
    assert_eq!(b.name, "main");
    assert_eq!(b.ontology_id, ontology_id);
    assert!(b.head_commit_id.is_none());
    assert_eq!(b.created_at, created_at);
    assert!(b.is_protected);
    assert!(
        b.last_commit_message.is_none()
            && b.last_commit_author.is_none()
            && b.last_commit_at.is_none(),
        "optional commit metadata must default to None when the upstream omits it"
    );
    assert_eq!(b.ahead_count, 0, "ahead_count must default to 0");
    assert_eq!(b.behind_count, 0, "behind_count must default to 0");
}

/// When the upstream supplies commit metadata, it is preserved verbatim. This
/// guards against accidental renaming of the wire field names.
#[test]
fn test_remote_branch_preserves_commit_metadata_when_present() {
    let branch_id = Uuid::new_v4();
    let ontology_id = Uuid::new_v4();
    let head_commit_id = Uuid::new_v4();
    let created_at = Utc.with_ymd_and_hms(2026, 7, 14, 9, 0, 0).unwrap();
    let last_commit_at = Utc.with_ymd_and_hms(2026, 7, 14, 9, 5, 0).unwrap();

    let payload = json!({
        "items": [{
            "id": branch_id,
            "name": "feature/x",
            "ontology_id": ontology_id,
            "head_commit_id": head_commit_id,
            "created_at": created_at.to_rfc3339(),
            "is_protected": false,
            "last_commit_message": "feat: add x",
            "last_commit_author": "Bob",
            "last_commit_at": last_commit_at.to_rfc3339(),
            "ahead_count": 2,
            "behind_count": 5,
        }],
        "total": 1,
    });

    let page: RemoteBranchPage = serde_json::from_value(payload).expect("parse branch page");
    let b: &RemoteBranch = &page.items[0];
    assert_eq!(b.head_commit_id, Some(head_commit_id));
    assert_eq!(b.last_commit_message.as_deref(), Some("feat: add x"));
    assert_eq!(b.last_commit_author.as_deref(), Some("Bob"));
    assert_eq!(b.last_commit_at, Some(last_commit_at));
    assert_eq!(b.ahead_count, 2);
    assert_eq!(b.behind_count, 5);
}

// ── branch_into_gql mapping helper ────────────────────────────────────────────

/// `branch_into_gql` must convert a `RemoteBranch` into a `GqlBranch` while
/// stringifying UUIDs and preserving `Option`al commit metadata.
#[test]
fn test_branch_into_gql_maps_all_fields() {
    let branch_id = Uuid::new_v4();
    let ontology_id = Uuid::new_v4();
    let head_commit_id = Uuid::new_v4();
    let created_at = Utc.with_ymd_and_hms(2026, 7, 14, 8, 0, 0).unwrap();

    let remote = RemoteBranch {
        id: branch_id,
        name: "main".to_string(),
        ontology_id,
        head_commit_id: Some(head_commit_id),
        created_at,
        is_protected: false,
        last_commit_message: Some("msg".to_string()),
        last_commit_author: Some("Alice".to_string()),
        last_commit_at: None,
        ahead_count: 7,
        behind_count: 1,
    };

    let gql: GqlBranch = branch_into_gql(remote);

    assert_eq!(gql.id, branch_id.to_string());
    assert_eq!(gql.name, "main");
    assert_eq!(gql.ontology_id, ontology_id.to_string());
    assert_eq!(
        gql.head_commit_id.as_deref(),
        Some(head_commit_id.to_string()).as_deref()
    );
    assert_eq!(gql.created_at, created_at.to_rfc3339());
    assert!(!gql.is_protected);
    assert_eq!(gql.last_commit_message.as_deref(), Some("msg"));
    assert_eq!(gql.last_commit_author.as_deref(), Some("Alice"));
    assert_eq!(gql.ahead_count, 7);
    assert_eq!(gql.behind_count, 1);
}

/// `branch_into_gql` must serialise UUIDs as lowercase strings even when the
/// source branch has no head commit. Guards against accidental `.unwrap()`
/// regressions in the `Option` projection.
#[test]
fn test_branch_into_gql_handles_branch_without_head_commit() {
    let branch_id = Uuid::new_v4();
    let ontology_id = Uuid::new_v4();
    let created_at = Utc.with_ymd_and_hms(2026, 7, 14, 8, 0, 0).unwrap();

    let remote = RemoteBranch {
        id: branch_id,
        name: "orphan".to_string(),
        ontology_id,
        head_commit_id: None,
        created_at,
        is_protected: false,
        last_commit_message: None,
        last_commit_author: None,
        last_commit_at: None,
        ahead_count: 0,
        behind_count: 0,
    };

    let gql: GqlBranch = branch_into_gql(remote);
    assert!(gql.head_commit_id.is_none());
    assert!(gql.last_commit_message.is_none());
    assert!(gql.last_commit_author.is_none());
}

// ── GraphQL type smoke tests ──────────────────────────────────────────────────

/// `GqlCommit` must serialize to JSON without the field renames (camelCase).
/// async-graphql derives the GraphQL field names automatically; this test just
/// confirms the struct serializes so rustc does not drop unused fields.
#[test]
fn test_gql_commit_serializes() {
    let commit = GqlCommit {
        id: "c1".to_string(),
        branch_id: "b1".to_string(),
        parent_commit_id: None,
        message: "msg".to_string(),
        author_id: "u1".to_string(),
        author_name: "Author".to_string(),
        total_changes: 4,
        created_at: "2026-07-14T10:00:00+00:00".to_string(),
    };
    let v = serde_json::to_value(&commit).expect("serialize GqlCommit");
    assert_eq!(v["id"], "c1");
    assert_eq!(v["branch_id"], "b1");
    assert_eq!(v["total_changes"], 4);
}

#[test]
fn test_gql_branch_serializes() {
    let branch = GqlBranch {
        id: "b1".to_string(),
        name: "main".to_string(),
        ontology_id: "o1".to_string(),
        head_commit_id: None,
        created_at: "2026-07-14T10:00:00+00:00".to_string(),
        is_protected: true,
        last_commit_message: None,
        last_commit_author: None,
        ahead_count: 0,
        behind_count: 0,
    };
    let v = serde_json::to_value(&branch).expect("serialize GqlBranch");
    assert_eq!(v["name"], "main");
    assert_eq!(v["is_protected"], true);
    assert!(v.get("head_commit_id").unwrap().is_null());
}
