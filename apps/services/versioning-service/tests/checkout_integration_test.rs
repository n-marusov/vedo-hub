//! Integration tests for checkout/switch operations with PostgreSQL.
//!
//! Validates: REQ-FUN.DATA.versioning
//!
//! The switch_branch smoke test has been moved to `src/handlers/branch_handler.rs`
//! (mock-based). The checkout_commit test below still needs PostgreSQL because
//! `checkout_commit_handler` creates a `StateService` directly from the PG pool.
//!
//! Requires running PostgreSQL. Set PG_TEST_DATABASE_URL env var to enable.

mod common;

use axum::{
    body::Body,
    http::{Method, Request, StatusCode},
};
use tower::ServiceExt;

fn post_json(uri: &str, body: &str) -> Request<Body> {
    Request::builder()
        .method(Method::POST)
        .uri(uri)
        .header("Content-Type", "application/json")
        .body(Body::from(body.to_string()))
        .unwrap()
}

#[tokio::test]
async fn test_checkout_commit_endpoint() {
    common::require_pg();
    let pool = common::connect_test_pg().await;
    let app = common::build_test_app(pool);

    // Use nil UUIDs since no actual commit/branch exists in test DB.
    // The endpoint should return 404 (commit not found).
    let commit_id = uuid::Uuid::nil();
    let branch_id = uuid::Uuid::nil();

    let resp = app
        .clone()
        .oneshot(post_json(
            &format!("/api/v1/versioning/commits/{commit_id}/checkout"),
            &format!(r#"{{"branch_id":"{branch_id}"}}"#),
        ))
        .await
        .unwrap();
    assert!(
        resp.status().is_success() || resp.status() == StatusCode::NOT_FOUND,
        "expected success or 404, got {}",
        resp.status()
    );
}
