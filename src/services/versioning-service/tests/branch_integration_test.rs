//! Integration tests for branch operations with PostgreSQL.
//!
//! Validates: REQ-FUN.DATA.versioning
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
async fn test_create_branch() {
    if !common::skip_if_no_pg() {
        return;
    }
    let pool = common::connect_test_pg().await;
    let app = common::build_test_app(pool);

    let resp = app.clone().oneshot(req(
        Method::POST,
        "/api/v1/versioning/branches",
        Some(r#"{"ontology_id":"test-onto","name":"feature/test-branch","source_branch":"main"}"#),
    )).await.unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
}

#[tokio::test]
async fn test_list_branches() {
    if !common::skip_if_no_pg() {
        return;
    }
    let pool = common::connect_test_pg().await;
    let app = common::build_test_app(pool);

    let resp = app
        .clone()
        .oneshot(req(Method::GET, "/api/v1/versioning/branches", None))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
}

#[tokio::test]
async fn test_get_branch() {
    if !common::skip_if_no_pg() {
        return;
    }
    let pool = common::connect_test_pg().await;
    let app = common::build_test_app(pool);

    let resp = app
        .clone()
        .oneshot(req(Method::GET, "/api/v1/versioning/branches/main", None))
        .await
        .unwrap();
    assert!(resp.status().is_success() || resp.status() == StatusCode::NOT_FOUND);
}

#[tokio::test]
async fn test_delete_branch() {
    if !common::skip_if_no_pg() {
        return;
    }
    let pool = common::connect_test_pg().await;
    let app = common::build_test_app(pool);

    let resp = app
        .clone()
        .oneshot(req(
            Method::DELETE,
            "/api/v1/versioning/branches/temp",
            None,
        ))
        .await
        .unwrap();
    assert!(resp.status().is_success() || resp.status() == StatusCode::NOT_FOUND);
}
