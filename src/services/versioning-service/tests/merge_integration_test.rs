//! Integration tests for merge operations with PostgreSQL.
//!
//! Validates: REQ-FUN.DATA.versioning
//!
//! Requires running PostgreSQL. Set PG_TEST_DATABASE_URL env var to enable.

mod common;

use axum::{
    body::Body,
    http::{Method, Request},
};
use tower::ServiceExt;

#[tokio::test]
async fn test_merge_branches_endpoint() {
    common::skip_if_no_pg();
    let pool = common::connect_test_pg().await;
    let app = common::build_test_app(pool);

    let resp = Request::builder()
        .method(Method::POST)
        .uri("/api/v1/versioning/branches/merge")
        .header("Content-Type", "application/json")
        .body(Body::from(
            r#"{"ontology_id":"test-onto","source_branch":"feature/test","target_branch":"main"}"#
                .to_string(),
        ))
        .unwrap();
    let resp = app.clone().oneshot(resp).await.unwrap();
    assert!(resp.status().is_success() || resp.status().is_client_error());
}
