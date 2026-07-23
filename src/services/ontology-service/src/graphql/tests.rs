//! Unit tests for the GraphQL facade assembled by `ontology-service`.
//!
//! These tests verify the tightened GraphQL schema contract after
//! migration to EmptyMutation and removal of non-graph resolvers:
//!
//! 1. `build_schema()` uses `EmptyMutation` — no Mutation type exposed.
//! 2. GraphQL introspection does not expose non-graph query fields
//!    (commits, branch, branches, groups, projects, members, ontology).
//! 3. Graph navigation resolvers (class, classTree, etc.) are present.
//! 4. `VersioningClient` correctly parses the REST payload shapes returned
//!    by the versioning-service (`RemoteCommitPage`, `RemoteBranchPage`).

use chrono::{TimeZone, Utc};
use serde_json::json;
use uuid::Uuid;

use super::schema::{build_schema, OntologySchema};
use super::versioning_client::{
    RemoteBranch, RemoteBranchPage, RemoteCommitPage, RemoteCommitSummary,
};

// ── Schema construction ──────────────────────────────────────────────────────

/// build_schema() must use EmptyMutation — no mutation resolvers.
/// Executing any mutation against the schema must return a "Cannot query field"
/// error, confirming that GraphQL is read-only for graph navigation.
#[tokio::test]
async fn test_schema_has_no_mutation_root() {
    let schema: OntologySchema = build_schema();
    use async_graphql::{Name, Variables};
    let mut vars = Variables::default();
    vars.insert(
        Name::new("ontologyId"),
        async_graphql::value!("00000000-0000-0000-0000-000000000000"),
    );

    let response = schema
        .execute(
            async_graphql::Request::new(
                r#"mutation UpdateDraft($ontologyId: ID!) {
                    updateDraft(ontologyId: $ontologyId) { success }
                }"#,
            )
            .variables(vars),
        )
        .await;

    assert!(
        !response.errors.is_empty(),
        "expected GraphQL error from updateDraft on EmptyMutation, got no errors"
    );
    let has_schema_error = response
        .errors
        .iter()
        .any(|e| e.message.contains("not configured for mutations"));
    assert!(
        has_schema_error,
        "expected error referencing updateDraft or mutation configuration, got: {:?}",
        response.errors
    );
}

/// Sanity check: the schema builds without panicking.
#[test]
fn test_schema_builds_successfully() {
    let _schema = build_schema();
}

/// Introspection should expose only graph-navigation query fields.
#[tokio::test]
async fn test_introspection_reveals_graph_only_fields() {
    let schema: OntologySchema = build_schema();

    let response = schema
        .execute(r#"{ __schema { queryType { fields { name } } } }"#)
        .await;

    let data = response.data.into_json().expect("introspection JSON");
    let fields: Vec<&str> = data["__schema"]["queryType"]["fields"]
        .as_array()
        .map(|arr| arr.iter().filter_map(|f| f["name"].as_str()).collect())
        .unwrap_or_default();

    // Graph navigation resolvers that MUST be present
    let expected_graph = [
        "class",
        "classes",
        "classTree",
        "classAncestors",
        "classDescendants",
        "graphNeighborhood",
        "autocompleteClasses",
        "property",
        "properties",
        "individual",
        "individuals",
    ];
    for name in &expected_graph {
        assert!(
            fields.contains(name),
            "expected graph query '{name}' in schema, found: {fields:?}"
        );
    }

    // Non-graph resolvers that MUST NOT be present
    let forbidden = [
        "commits", "branch", "branches", "groups", "projects", "members", "ontology",
    ];
    for name in &forbidden {
        assert!(
            !fields.contains(name),
            "non-graph query '{name}' must NOT be in schema, found: {fields:?}"
        );
    }

    // Mutation type must be absent (EmptyMutation hides it from introspection)
    let has_mutation = data["__schema"].get("mutationType").is_some()
        && !data["__schema"]["mutationType"].is_null();
    assert!(!has_mutation, "Mutation type must be absent from schema");
}

// ── VersioningClient payload parsing ─────────────────────────────────────────

/// `RemoteCommitPage` must deserialize from the response shape produced by
/// `GET /api/v1/versioning/commits`.
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
    assert!(c.parent_commit_id.is_none());
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

/// When the upstream supplies commit metadata, it is preserved verbatim.
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
