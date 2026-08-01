//! Integration tests for commit operations with PostgreSQL.
//!
//! Validates: REQ-FUN.DATA.versioning
//!
//! Only `test_create_commit_via_api` remains here because it verifies data
//! directly in PostgreSQL (SELECT COUNT(*)). Other commit smoke tests have
//! been moved to `src/handlers/commit_handler.rs` (mock-based).
//!
//! Requires running PostgreSQL. Set PG_TEST_DATABASE_URL env var to enable.

mod common;

use axum::{
    body::Body,
    http::{Method, Request, StatusCode},
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
async fn test_create_commit_via_api() {
    common::require_pg();
    let pool = common::connect_test_pg().await;
    let app = common::build_test_app(pool.clone());

    // Commits reference branches via FK, so create the branch first.
    let ontology_id = Uuid::new_v4();
    let branch_id = create_branch(&app, &ontology_id, "feat/commit-test").await;

    // Payload matches `CreateCommitRequest` (branch_id, message, author_id,
    // author_name, optional delta) and the handler returns 201 Created.
    let resp = app
        .clone()
        .oneshot(req(
            Method::POST,
            "/api/v1/versioning/commits",
            Some(
                &serde_json::json!({
                    "branch_id": branch_id.to_string(),
                    "message": "Initial commit",
                    "author_id": "test-user",
                    "author_name": "Test User",
                    // Non-empty delta is required — the repository rejects
                    // empty deltas with VER-COMMIT-EMPTY-DELTA (400).
                    "delta": {
                        "added_triples": [
                            {"s": "Person", "p": "rdfs:label", "o": "Person"}
                        ]
                    },
                })
                .to_string(),
            ),
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::CREATED);

    // Verify in PG
    let count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM commits WHERE author_id = $1")
        .bind("test-user")
        .fetch_one(pool.pool())
        .await
        .unwrap_or(0);
    assert!(count > 0, "commit should exist in PostgreSQL");
}
