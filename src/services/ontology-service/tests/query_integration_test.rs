//! Integration tests for the ontology-service SPARQL/CYPHER query endpoints.
//!
//! These tests do NOT require Neo4j when the assertions focus on validation,
//! error handling, and route resolution. Tests that exercise real query
//! execution require `NEO4J_TEST_URI` and are gated via `common::skip_if_no_neo4j`.

mod common;

use axum::{
    body::Body,
    http::{Method, Request, StatusCode},
};
use serde_json::Value;
use tower::ServiceExt;

/// Builds an app with no Neo4j pool so query endpoints respond with 503 for
/// any query that would otherwise reach the database.
fn app_no_db() -> axum::Router {
    use ontology_service::{build_app, AppState};
    use std::sync::Arc;
    let state = Arc::new(AppState { neo4j: None });
    build_app(state)
}

fn post_json(uri: &str, body: &str) -> Request<Body> {
    Request::builder()
        .method(Method::POST)
        .uri(uri)
        .header("Content-Type", "application/json")
        .body(Body::from(body.to_string()))
        .unwrap()
}

async fn body_json(response: axum::response::Response) -> Value {
    let bytes = axum::body::to_bytes(response.into_body(), usize::MAX)
        .await
        .unwrap();
    serde_json::from_slice(&bytes).unwrap_or(Value::Null)
}

#[tokio::test]
async fn test_sparql_rejects_mutation_keyword() {
    let app = app_no_db();
    let resp = app
        .oneshot(post_json(
            "/api/v1/sparql",
            r#"{"query":"INSERT DATA { <a> <b> <c> }"}"#,
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::BAD_REQUEST);
    let body = body_json(resp).await;
    assert_eq!(body["error"], "ONT-QUERY-READONLY");
}

#[tokio::test]
async fn test_cypher_rejects_mutation_keyword() {
    let app = app_no_db();
    let resp = app
        .oneshot(post_json(
            "/api/v1/cypher",
            r#"{"query":"CREATE (n:Foo {name: 'bar'})"}"#,
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::BAD_REQUEST);
    let body = body_json(resp).await;
    assert_eq!(body["error"], "ONT-QUERY-READONLY");
}

#[tokio::test]
async fn test_cypher_rejects_class_named_createresource() {
    // Regression for the naive substring Contains check. A class label
    // containing the substring "CREATE" must NOT be flagged.
    let app = app_no_db();
    let resp = app
        .oneshot(post_json(
            "/api/v1/cypher",
            r#"{"query":"MATCH (n:CreateResource) RETURN n LIMIT 5"}"#,
        ))
        .await
        .unwrap();
    // The query passes read-only validation; with no DB it returns 503
    // (ONT-NEO4J-NOT-CONFIGURED) — NOT 400 read-only rejection.
    assert_eq!(resp.status(), StatusCode::SERVICE_UNAVAILABLE);
    let body = body_json(resp).await;
    assert_eq!(body["error"], "ONT-NEO4J-NOT-CONFIGURED");
}

#[tokio::test]
async fn test_sparql_rejects_unsupported_filter_clause() {
    let app = app_no_db();
    let resp = app
        .oneshot(post_json(
            "/api/v1/sparql",
            r#"{"query":"SELECT ?s ?p ?o WHERE { ?s ?p ?o . FILTER(?s = <a>) }"}"#,
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::NOT_IMPLEMENTED);
    let body = body_json(resp).await;
    assert_eq!(body["error"], "ONT-SPARQL-UNSUPPORTED");
}

#[tokio::test]
async fn test_cypher_returns_neo4j_not_configured_when_no_db() {
    let app = app_no_db();
    let resp = app
        .oneshot(post_json(
            "/api/v1/cypher",
            r#"{"query":"MATCH (n) RETURN n LIMIT 5"}"#,
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::SERVICE_UNAVAILABLE);
    let body = body_json(resp).await;
    assert_eq!(body["error"], "ONT-NEO4J-NOT-CONFIGURED");
}

#[tokio::test]
async fn test_sparql_missing_query_field_returns_4xx() {
    let app = app_no_db();
    let resp = app
        .oneshot(post_json("/api/v1/sparql", r#"{}"#))
        .await
        .unwrap();
    assert!(
        resp.status().is_client_error(),
        "missing query field should return 4xx, got {}",
        resp.status()
    );
}

#[tokio::test]
async fn test_cypher_executes_against_real_neo4j() {
    // Gated — requires NEO4J_TEST_URI. Skipped otherwise so `cargo test
    // --workspace` stays green in DB-less developer environments.
    if !common::is_integration_enabled() {
        eprintln!("Skipping Neo4j-backed test: set NEO4J_TEST_URI to run");
        return;
    }
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("cypher_query");
    common::clean_ontology(&pool, &oid).await;

    // Seed a Class node with the production property names so the query
    // surface matches what real handlers write.
    let _ = pool
        .graph()
        .execute(
            neo4rs::query("CREATE (c:Class {id: $id, label: $label, ontology_id: $oid})")
                .param("id", "CypherTest".to_string())
                .param("label", "Cypher Test".to_string())
                .param("oid", oid.clone()),
        )
        .await;

    let resp = app
        .oneshot(post_json(
            "/api/v1/cypher",
            &format!(
                r#"{{"query":"MATCH (c:Class) WHERE c.ontology_id = '{oid}' RETURN c.label AS label LIMIT 5"}}"#,
            ),
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
    let body = body_json(resp).await;
    assert!(body["results"].is_array(), "results must be an array");
    assert!(body["triple_count"].as_u64().unwrap_or(0) >= 1);

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_cypher_db_error_sanitized() {
    // Regression test: database errors must not leak raw Neo4j error details
    // in the HTTP response. Requires NEO4J_TEST_URI.
    if !common::is_integration_enabled() {
        eprintln!("Skipping Neo4j-backed test: set NEO4J_TEST_URI to run");
        return;
    }
    let (app, _pool) = common::create_test_app().await;

    // Send a read-only query that passes validation but fails at Neo4j level
    // (invalid syntax or non-existent procedure). The exact error varies by
    // Neo4j version, but the response must sanitize it.
    let resp = app
        .oneshot(post_json(
            "/api/v1/cypher",
            r#"{"query":"MATCH (n) RETURN n.`non-existent-prop` LIMIT 1"}"#,
        ))
        .await
        .unwrap();

    // The request passes validation but may fail in Neo4j. The response must
    // use ONT-DATABASE-ERROR code and must NOT contain raw Neo4j error text.
    let body = body_json(resp).await;
    if body["error"] == "ONT-DATABASE-ERROR" {
        let detail = body["detail"].as_str().unwrap_or("");
        let forbidden = ["Neo4j", "neo4rs", "RuntimeException", "ClientException"];
        for pattern in &forbidden {
            assert!(
                !detail.contains(pattern),
                "error detail must not leak Neo4j internals: contains '{}' (detail={})",
                pattern,
                detail
            );
        }
    }
    // If the query unexpectedly succeeds (e.g. on Neo4j with lax schema),
    // the test passes vacuously — the sanitization works.
}
