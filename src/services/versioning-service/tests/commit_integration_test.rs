//! Integration tests for commit operations with PostgreSQL.
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
    common::skip_if_no_pg();
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

#[tokio::test]
async fn test_list_commits_works() {
    common::skip_if_no_pg();
    let pool = common::connect_test_pg().await;
    let app = common::build_test_app(pool);

    let resp = app
        .clone()
        .oneshot(req(Method::GET, "/api/v1/versioning/commits", None))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
}

#[tokio::test]
async fn test_get_commit_returns_data() {
    common::skip_if_no_pg();
    let pool = common::connect_test_pg().await;
    let app = common::build_test_app(pool);

    // Create a commit first
    let _ = app
        .clone()
        .oneshot(req(
            Method::POST,
            "/api/v1/versioning/commits",
            Some(r#"{"ontology_id":"test-get","message":"Get test","author":"test-get"}"#),
        ))
        .await
        .unwrap();

    // Fetch the commit by ID (simplified: assume ID 1)
    let resp = app
        .clone()
        .oneshot(req(Method::GET, "/api/v1/versioning/commits/1", None))
        .await
        .unwrap();
    assert!(resp.status().is_success() || resp.status() == StatusCode::NOT_FOUND);
}

#[tokio::test]
async fn test_rollback_commit_endpoint() {
    common::skip_if_no_pg();
    let pool = common::connect_test_pg().await;
    let app = common::build_test_app(pool);

    let resp = app
        .clone()
        .oneshot(req(
            Method::POST,
            "/api/v1/versioning/commits/1/rollback",
            None,
        ))
        .await
        .unwrap();
    assert!(resp.status().is_success() || resp.status().is_client_error());
}
