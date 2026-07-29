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

fn req(method: Method, uri: &str, body: Option<&str>) -> Request<Body> {
    let mut b = Request::builder().method(method).uri(uri);
    if let Some(body) = body {
        b = b.header("Content-Type", "application/json");
        b.body(Body::from(body.to_string())).unwrap()
    } else {
        b.body(Body::empty()).unwrap()
    }
}

#[tokio::test]
async fn test_create_commit_via_api() {
    common::require_pg();
    let pool = common::connect_test_pg().await;
    let app = common::build_test_app(pool.clone());

    let resp = app
        .clone()
        .oneshot(req(
            Method::POST,
            "/api/v1/versioning/commits",
            Some(r#"{"ontology_id":"test-onto","message":"Initial commit","author":"test-user"}"#),
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);

    // Verify in PG
    let count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM commits WHERE author_id = $1")
        .bind("test-user")
        .fetch_one(pool.pool())
        .await
        .unwrap_or(0);
    assert!(count > 0, "commit should exist in PostgreSQL");
}
