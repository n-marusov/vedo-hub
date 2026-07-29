//! Integration tests for import/export operations.
//!
//! Validates: REQ-FUN.DATA.ontology-import-export
//!
//! Requires running Neo4j. Run via `make test-integration-rust` (auto-starts Neo4j via Docker Compose).
//! Set NEO4J_TEST_URI env var to run manually.

mod common;

use axum::{
    body::Body,
    http::{Method, Request, StatusCode},
};
use tower::ServiceExt;

fn get(uri: &str) -> Request<Body> {
    Request::builder()
        .method(Method::GET)
        .uri(uri)
        .body(Body::empty())
        .unwrap()
}

fn post_body(uri: &str, content_type: &str, body: &str) -> Request<Body> {
    Request::builder()
        .method(Method::POST)
        .uri(uri)
        .header("Content-Type", content_type)
        .body(Body::from(body.to_string()))
        .unwrap()
}

#[tokio::test]
async fn test_export_turtle_returns_content() {
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("export_ttl");

    common::execute_query(
        &pool,
        neo4rs::query("CREATE (c:Class {ontology_id:$id,id:'Person',label:'Person',iri:$iri})")
            .param("id", oid.clone())
            .param("iri", format!("http://example.org/{oid}#Person")),
    )
    .await;

    let resp = app
        .clone()
        .oneshot(get(&format!(
            "/api/v1/ontologies/{oid}/export?format=turtle"
        )))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
    let text = String::from_utf8_lossy(
        &axum::body::to_bytes(resp.into_body(), usize::MAX)
            .await
            .unwrap(),
    )
    .to_string();
    assert!(!text.is_empty());
    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_export_rdfxml_returns_content() {
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("export_rdf");

    common::execute_query(
        &pool,
        neo4rs::query("CREATE (c:Class {ontology_id:$id,id:'Test',label:'Test Class'})")
            .param("id", oid.clone()),
    )
    .await;

    let resp = app
        .clone()
        .oneshot(get(&format!(
            "/api/v1/ontologies/{oid}/export?format=rdfxml"
        )))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
    let text = String::from_utf8_lossy(
        &axum::body::to_bytes(resp.into_body(), usize::MAX)
            .await
            .unwrap(),
    )
    .to_string();
    assert!(!text.is_empty());
    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_import_invalid_format_returns_error() {
    let (app, _pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("import_bad");

    let resp = app
        .clone()
        .oneshot(post_body(
            &format!("/api/v1/ontologies/{oid}/import?format=turtle"),
            "text/turtle",
            "@@@invalid@@@",
        ))
        .await
        .unwrap();
    assert!(resp.status().is_client_error());
}
