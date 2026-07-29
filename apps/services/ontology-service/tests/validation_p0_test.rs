/// Placeholder integration tests for not-yet-implemented endpoints (M12).
/// Kept #[ignore] until endpoints are wired; remove #[ignore] when implemented.
///
///! Validates: REQ-FUN.API.validation-import
///! Validates: REQ-FUN.API.validation-pr
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

// Import Validation (REQ-FUN.API.validation-import)

#[ignore = "REQ-FUN.API.validation-import: requires import validation endpoint with SHACL — implement in M12"]
#[tokio::test]
async fn test_import_validation_rejects_invalid_turtle() {
    let (app, _pool) = common::create_test_app().await;

    let resp = app
        .clone()
        .oneshot(post_json(
            "/api/v1/ontologies/test-onto/validate-import",
            r#"{"turtle": "@prefix : <http://example.org/> . :Invalid"}"#,
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::UNPROCESSABLE_ENTITY);
}

#[ignore = "REQ-FUN.API.validation-import: requires import validation endpoint"]
#[tokio::test]
async fn test_import_validation_accepts_valid_turtle() {
    let (app, _pool) = common::create_test_app().await;

    let resp = app
        .clone()
        .oneshot(post_json(
            "/api/v1/ontologies/test-onto/validate-import",
            r#"{"turtle": "@prefix : <http://example.org/> . :Person a owl:Class ."}"#,
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
}

// PR Validation (REQ-FUN.API.validation-pr)

#[ignore = "REQ-FUN.API.validation-pr: requires merge request validation endpoint — implement in M12"]
#[tokio::test]
async fn test_pr_validation_detects_cyclic_hierarchy() {
    let (app, _pool) = common::create_test_app().await;

    let resp = app
        .clone()
        .oneshot(post_json(
            "/api/v1/ontologies/test-onto/validate-pr",
            r#"{"from_commit": "abc", "to_commit": "def"}"#,
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
    // TODO: Verify response contains cycle detection result
}

#[ignore = "REQ-FUN.API.validation-pr: requires merge request validation endpoint"]
#[tokio::test]
async fn test_pr_validation_reports_broken_references() {
    let (app, _pool) = common::create_test_app().await;

    let resp = app
        .clone()
        .oneshot(post_json(
            "/api/v1/ontologies/test-onto/validate-pr",
            r#"{"from_commit": "abc", "to_commit": "def"}"#,
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
    // TODO: Verify response contains broken reference report
}
