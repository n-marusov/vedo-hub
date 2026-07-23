//! Route-registration smoke tests for the ontology-service.
//!
//! Validates: REQ-FUN.API.route-registration
//!
//! These tests verify that every parameterized route registered in
//! `build_app` resolves correctly under axum 0.7's `:param` syntax. The
//! regression they guard against: a previous revision used `{param}` syntax,
//! which axum 0.7 (matchit 0.7) treats as a literal path segment — every
//! parameterized request returned 404 silently.
//!
//! The tests do NOT require Neo4j. The app is built with `neo4j: None`; a
//! successful route match returns a handler-specific status (often 503
//! Service Unavailable) — the assertion is "not 404".

mod common;

use std::sync::Arc;

use axum::{
    body::Body,
    http::{Method, Request, StatusCode},
};
use ontology_service::clients::auth_client::AuthClient;
use ontology_service::{build_app, AppState};
use tower::ServiceExt;

/// Builds the app with no Neo4j connection so route resolution can be tested
/// without a database. Each registered route should resolve to a handler
/// (returning 503 SERVICE_UNAVAILABLE for DB-backed endpoints) rather than
/// the axum default 404.
fn app_no_db() -> axum::Router {
    let state = Arc::new(AppState {
        neo4j: None,
        auth_client: AuthClient::new(None),
    });
    build_app(state)
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

async fn empty() -> Body {
    Body::empty()
}

#[tokio::test]
async fn test_class_routes_resolve() {
    let app = app_no_db();
    assert_route_resolves(
        app.clone(),
        Method::POST,
        "/api/v1/ontologies/ont-1/classes",
        Body::from(r#"{"id":"X","label":"X"}"#),
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/ontologies/ont-1/classes",
        empty().await,
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/ontologies/ont-1/classes/root",
        empty().await,
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/ontologies/ont-1/classes/search/autocomplete",
        empty().await,
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/ontologies/ont-1/classes/Person",
        empty().await,
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::PUT,
        "/api/v1/ontologies/ont-1/classes/Person",
        Body::from(r#"{"label":"New"}"#),
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::DELETE,
        "/api/v1/ontologies/ont-1/classes/Person",
        empty().await,
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/ontologies/ont-1/classes/Person/children",
        empty().await,
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/ontologies/ont-1/classes/Person/ancestors",
        empty().await,
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/ontologies/ont-1/classes/Person/descendants",
        empty().await,
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/ontologies/ont-1/classes/Person/breadcrumb",
        empty().await,
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/ontologies/ont-1/classes/Person/neighborhood",
        empty().await,
    )
    .await;
}

#[tokio::test]
async fn test_property_routes_resolve() {
    let app = app_no_db();
    assert_route_resolves(
        app.clone(),
        Method::POST,
        "/api/v1/ontologies/ont-1/properties",
        Body::from(r#"{"id":"P","label":"P"}"#),
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/ontologies/ont-1/properties",
        empty().await,
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/ontologies/ont-1/properties/hasName",
        empty().await,
    )
    .await;
}

#[tokio::test]
async fn test_individual_routes_resolve() {
    let app = app_no_db();
    assert_route_resolves(
        app.clone(),
        Method::POST,
        "/api/v1/ontologies/ont-1/individuals",
        Body::from(r#"{"id":"i","label":"i"}"#),
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/ontologies/ont-1/individuals",
        empty().await,
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/ontologies/ont-1/individuals/ind-1",
        empty().await,
    )
    .await;
}

#[tokio::test]
async fn test_export_import_routes_resolve() {
    let app = app_no_db();
    assert_route_resolves(
        app.clone(),
        Method::GET,
        "/api/v1/ontologies/ont-1/export",
        empty().await,
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::POST,
        "/api/v1/ontologies/ont-1/import",
        Body::from(r#"{"format":"turtle","content":"<a> <b> <c> ."}"#),
    )
    .await;
}

#[tokio::test]
async fn test_query_routes_resolve() {
    let app = app_no_db();
    assert_route_resolves(
        app.clone(),
        Method::POST,
        "/api/v1/sparql",
        Body::from(r#"{"query":"SELECT ?s ?p ?o WHERE { ?s ?p ?o }"}"#),
    )
    .await;
    assert_route_resolves(
        app.clone(),
        Method::POST,
        "/api/v1/cypher",
        Body::from(r#"{"query":"MATCH (n) RETURN n LIMIT 5"}"#),
    )
    .await;
}

#[tokio::test]
async fn test_graphql_route_resolves() {
    let app = app_no_db();
    assert_route_resolves(
        app.clone(),
        Method::POST,
        "/api/v1/graphql",
        Body::from(r#"{"query":"{ __typename }"}"#),
    )
    .await;
}
