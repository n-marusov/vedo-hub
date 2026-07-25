//! Route-registration smoke tests for the versioning-service.
//!
//! Validates: REQ-FUN.API.route-registration
//!
//! Verifies that every parameterized route registered in `build_routes`
//! resolves correctly under axum 0.7's `:param` syntax. The regression they
//! guard against: the previous revision used `{param}` syntax, which axum 0.7
//! (matchit 0.7) treats as a literal path segment — every parameterized
//! request returned 404 silently.
//!
//! These tests do NOT require PostgreSQL. The app is built with `pg: None`;
//! a successful route match returns 503 (Service Unavailable) — the
//! assertion is "not 404".

mod common;

use std::sync::Arc;

use axum::{
    body::Body,
    http::{Method, Request, StatusCode},
};
use tower::ServiceExt;
use versioning_service::{routes::build_routes, AppState};

fn app_no_db() -> axum::Router {
    let state = Arc::new(AppState { pg: None });
    build_routes().with_state(state)
}

async fn assert_route_resolves(app: axum::Router, method: Method, uri: &str, body: Body) {
    let method_for_msg = method.clone();
    let resp = app
        .oneshot(
            Request::builder()
                .method(method)
                .uri(uri)
                .header("Content-Type", "application/json")
                .body(body)
                .unwrap(),
        )
        .await
        .unwrap();
    assert_ne!(
        resp.status(),
        StatusCode::NOT_FOUND,
        "route {method_for_msg} {uri} resolved to 404 — check axum :param syntax"
    );
}

#[tokio::test]
async fn test_commit_routes_resolve() {
    let app = app_no_db();
    assert_route_resolves(
        app.clone(),
        Method::POST,
        "/api/v1/versioning/commits",
        Body::from(r#"{"message":"m","parent_id":null,"author_id":"a"}"#),
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/versioning/commits",
        Body::empty(),
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/versioning/commits/abc-123",
        Body::empty(),
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/versioning/commits/abc-123/delta",
        Body::empty(),
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::POST,
        "/api/v1/versioning/commits/abc-123/checkout",
        Body::empty(),
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::POST,
        "/api/v1/versioning/commits/abc-123/rollback",
        Body::empty(),
    )
    .await;
}

#[tokio::test]
async fn test_branch_routes_resolve() {
    let app = app_no_db();
    assert_route_resolves(
        app.clone(),
        Method::POST,
        "/api/v1/versioning/branches",
        Body::from(r#"{"name":"feat","from_commit":"abc"}"#),
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/versioning/branches",
        Body::empty(),
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/versioning/branches/feat-1",
        Body::empty(),
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::DELETE,
        "/api/v1/versioning/branches/feat-1",
        Body::empty(),
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::POST,
        "/api/v1/versioning/branches/feat-1/switch",
        Body::empty(),
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::POST,
        "/api/v1/versioning/branches/merge",
        Body::from(r#"{"source":"a","target":"b"}"#),
    )
    .await;
}

// Suppress unused-import warning when the `common` module's helpers are not
// directly referenced.
#[allow(dead_code)]
fn _silence_common_module() {
    let _ = common::is_integration_enabled();
}
