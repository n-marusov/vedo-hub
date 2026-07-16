//! Integration tests for checkout/switch operations with PostgreSQL.
//!
//! Requires running PostgreSQL. Set PG_TEST_DATABASE_URL env var to enable.

mod common;

use axum::{
    body::Body,
    http::{Method, Request, StatusCode},
};
use tower::ServiceExt;

fn post_empty(uri: &str) -> Request<Body> {
    Request::builder()
        .method(Method::POST)
        .uri(uri)
        .body(Body::empty())
        .unwrap()
}

#[tokio::test]
async fn test_checkout_commit_endpoint() {
    common::skip_if_no_pg();
    let pool = common::connect_test_pg().await;
    let app = common::build_test_app(pool);

    let resp = app
        .clone()
        .oneshot(post_empty("/api/v1/versioning/commits/1/checkout"))
        .await
        .unwrap();
    assert!(resp.status().is_success() || resp.status() == StatusCode::NOT_FOUND);
}

#[tokio::test]
async fn test_switch_branch_endpoint() {
    common::skip_if_no_pg();
    let pool = common::connect_test_pg().await;
    let app = common::build_test_app(pool);

    let resp = app
        .clone()
        .oneshot(post_empty("/api/v1/versioning/branches/main/switch"))
        .await
        .unwrap();
    assert!(resp.status().is_success() || resp.status() == StatusCode::NOT_FOUND);
}
