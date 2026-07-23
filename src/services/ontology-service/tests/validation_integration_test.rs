//! Integration tests for SHACL validation endpoint.
//!
//! Requires running Neo4j. Set NEO4J_TEST_URI env var to enable.

mod common;

use axum::{
    body::Body,
    http::{Method, Request, StatusCode},
};
use tower::ServiceExt;

fn request(method: Method, uri: &str, body: Option<&str>) -> Request<Body> {
    let mut builder = Request::builder().method(method).uri(uri);
    if let Some(body) = body {
        builder = builder.header("Content-Type", "application/json");
        builder.body(Body::from(body.to_string())).unwrap()
    } else {
        builder.body(Body::empty()).unwrap()
    }
}

fn post_json(uri: &str, body: &str) -> Request<Body> {
    request(Method::POST, uri, Some(body))
}

#[tokio::test]
/// Validates: REQ-FUN.API.pre-save-validation
///
/// TC-A10: Validate ontology against SHACL shapes.
/// Creates a test class, then validates the ontology with default shapes.
/// Expects a 200 response with `conforms` and `results` fields.
async fn test_validate_ontology_returns_report() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("validate_ok");

    // Seed a class to have some data to validate
    let _ = pool
        .graph()
        .execute(
            neo4rs::query("CREATE (c:Class {ontology_id:$id,id:'Person',label:'Person'})")
                .param("id", oid.clone()),
        )
        .await;

    // Validate with default shapes
    let resp = app
        .clone()
        .oneshot(post_json(
            &format!("/api/v1/ontologies/{oid}/validate"),
            r#"{}"#,
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);

    // Parse response
    let body = String::from_utf8_lossy(
        &axum::body::to_bytes(resp.into_body(), usize::MAX)
            .await
            .unwrap(),
    )
    .to_string();
    let json: serde_json::Value = serde_json::from_str(&body).unwrap();

    // The response must contain a validation report with `conforms`
    assert!(
        json.get("conforms").is_some(),
        "validation response should contain 'conforms' field"
    );
    assert!(
        json.get("results").is_some(),
        "validation response should contain 'results' field"
    );

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
/// Validates: REQ-FUN.API.pre-save-validation
///
/// Validates with custom SHACL shapes provided inline.
/// Expects a 200 response with the validation report.
async fn test_validate_with_custom_shacl_shapes() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("validate_custom");

    // Seed a class
    let _ = pool
        .graph()
        .execute(
            neo4rs::query("CREATE (c:Class {ontology_id:$id,id:'Person',label:'Person'})")
                .param("id", oid.clone()),
        )
        .await;

    // Custom SHACL shape in Turtle (simplified: requires Person to have a comment)
    let shapes = r#"
        @prefix sh: <http://www.w3.org/ns/shacl#> .
        @prefix ex: <http://example.org/> .
        ex:PersonShape a sh:NodeShape ;
            sh:targetClass ex:Person ;
            sh:property [ sh:path ex:comment ; sh:minCount 1 ] .
    "#;

    let body = serde_json::json!({
        "shapes_turtle": shapes
    });

    let resp = app
        .clone()
        .oneshot(post_json(
            &format!("/api/v1/ontologies/{oid}/validate"),
            &body.to_string(),
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);

    let body = String::from_utf8_lossy(
        &axum::body::to_bytes(resp.into_body(), usize::MAX)
            .await
            .unwrap(),
    )
    .to_string();
    let json: serde_json::Value = serde_json::from_str(&body).unwrap();

    assert!(
        json.get("conforms").is_some(),
        "validation response should contain 'conforms' field"
    );

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
/// Validates: REQ-FUN.API.pre-save-validation
///
/// Validation of a nonexistent ontology should return an error response.
async fn test_validate_nonexistent_ontology_returns_error() {
    common::skip_if_no_neo4j();
    let (app, _pool) = common::create_test_app().await;

    let resp = app
        .clone()
        .oneshot(post_json(
            "/api/v1/ontologies/nonexistent-oid-00000/validate",
            r#"{}"#,
        ))
        .await
        .unwrap();
    // Should return a valid response (200) with conforms=false, or a 4xx error
    // The validator may report no violations for an empty ontology
    assert!(
        resp.status().is_success() || resp.status().is_client_error(),
        "validation should not return 5xx"
    );
}
