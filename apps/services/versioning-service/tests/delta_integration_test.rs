//! Integration tests for commit delta/diff operations with PostgreSQL.
//!
//! Validates: REQ-FUN.DATA.versioning
//!
//! Only `test_commit_delta_endpoint` remains here — it's a full end-to-end flow
//! that creates a branch and commit, then verifies the delta response structure.
//! The 404 edge-case tests have been moved to `src/handlers/commit_handler.rs`
//! (mock-based).
//!
//! Requires running PostgreSQL. Set PG_TEST_DATABASE_URL env var to enable.

mod common;

use axum::{
    body::Body,
    http::{Method, Request},
};
use tower::ServiceExt;
use uuid::Uuid;

fn req(method: Method, uri: &str, body: Option<&str>) -> Request<Body> {
    let mut b = Request::builder().method(method).uri(uri);
    if let Some(body) = body {
        b = b.header("Content-Type", "application/json");
        b.body(Body::from(body.to_string())).unwrap()
    } else {
        b.body(Body::empty()).unwrap()
    }
}

/// Helper: creates a branch and returns its ID (UUID) from the response.
async fn create_branch(app: &axum::Router, ontology_id: &Uuid, name: &str) -> Uuid {
    let body = serde_json::json!({
        "name": name,
        "ontology_id": ontology_id.to_string(),
    });
    let resp = app
        .clone()
        .oneshot(req(
            Method::POST,
            "/api/v1/versioning/branches",
            Some(&body.to_string()),
        ))
        .await
        .unwrap();
    assert!(
        resp.status().is_success(),
        "create branch should succeed, got {}",
        resp.status()
    );
    let body_bytes = axum::body::to_bytes(resp.into_body(), usize::MAX)
        .await
        .unwrap();
    let json: serde_json::Value = serde_json::from_slice(&body_bytes).unwrap();
    let id_str = json["id"].as_str().expect("branch response must have 'id'");
    Uuid::parse_str(id_str).expect("branch id must be a valid UUID")
}

#[tokio::test]
/// TC-B05: Diff between commits (delta endpoint).
/// Creates a branch and a commit, then fetches its delta to verify the response
/// contains the expected delta structure.
async fn test_commit_delta_endpoint() {
    common::require_pg();
    let pool = common::connect_test_pg().await;
    let app = common::build_test_app(pool);

    let ontology_id = Uuid::new_v4();

    // Create a branch to get a valid branch_id
    let branch_id = create_branch(&app, &ontology_id, "feat/delta-test").await;

    // Create a commit on that branch
    let body = serde_json::json!({
        "branch_id": branch_id.to_string(),
        "message": "Delta test commit",
        "author_id": "test-delta-author",
        "author_name": "Test Author",
        // Non-empty delta is required — the repository rejects empty deltas
        // with VER-COMMIT-EMPTY-DELTA (400).
        "delta": {
            "added_triples": [
                {"s": "ClassA", "p": "rdfs:label", "o": "Class A"}
            ]
        },
    });
    let resp = app
        .clone()
        .oneshot(req(
            Method::POST,
            "/api/v1/versioning/commits",
            Some(&body.to_string()),
        ))
        .await
        .unwrap();
    let status = resp.status();
    assert!(
        status.is_success(),
        "create commit should succeed, got {status}"
    );

    // Parse commit ID from the response body
    let body_bytes = axum::body::to_bytes(resp.into_body(), usize::MAX)
        .await
        .unwrap();
    let commit_json: serde_json::Value = serde_json::from_slice(&body_bytes).unwrap();
    let commit_id = commit_json["id"]
        .as_str()
        .expect("commit response must have 'id'")
        .to_string();

    // Fetch the delta for the created commit
    let resp = app
        .clone()
        .oneshot(req(
            Method::GET,
            &format!("/api/v1/versioning/commits/{commit_id}/delta"),
            None,
        ))
        .await
        .unwrap();
    assert!(
        resp.status().is_success(),
        "delta endpoint should succeed, got {}",
        resp.status()
    );

    // Parse delta response
    let body_bytes = axum::body::to_bytes(resp.into_body(), usize::MAX)
        .await
        .unwrap();
    let delta_json: serde_json::Value = serde_json::from_slice(&body_bytes).unwrap();

    // The delta response is a `CommitDeltaPreview` with *_total / *_preview
    // sections per added/removed/modified triples.
    assert_eq!(
        delta_json["added_total"], 1,
        "added_total mismatch: {delta_json}"
    );
    assert_eq!(
        delta_json["added_preview"][0]["s"], "ClassA",
        "added_preview[0].s mismatch: {delta_json}"
    );
}
