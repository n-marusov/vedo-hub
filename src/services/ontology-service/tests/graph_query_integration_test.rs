//! Integration tests for graph queries (hierarchy, neighborhood, breadcrumb).
//!
//! Requires running Neo4j. Set NEO4J_TEST_URI env var to enable.

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

async fn seed_class(pool: &ontology_service::neo4j::Neo4jPool, oid: &str, cid: &str, parent: Option<&str>) {
    if let Some(p) = parent {
        let _ = pool.graph().execute(
            neo4rs::query("MATCH (parent:Class {ontology_id:$id,id:$parent}) CREATE (c:Class {ontology_id:$id,id:$cid,label:$label}) CREATE (c)-[:CHILD_OF]->(parent)")
                .param("id", oid.to_string())
                .param("parent", p.to_string())
                .param("cid", cid.to_string())
                .param("label", cid.to_string()),
        ).await;
    } else {
        let _ = pool.graph().execute(
            neo4rs::query("CREATE (c:Class {ontology_id:$id,id:$cid,label:$label})")
                .param("id", oid.to_string())
                .param("cid", cid.to_string())
                .param("label", cid.to_string()),
        ).await;
    }
}

#[tokio::test]
async fn test_hierarchy_tree_returns_ancestors() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("hierarchy");
    seed_class(&pool, &oid, "owl:Thing", None).await;
    seed_class(&pool, &oid, "Animal", Some("owl:Thing")).await;
    seed_class(&pool, &oid, "Dog", Some("Animal")).await;

    for path in &["ancestors", "children", "descendants"] {
        let resp = app.clone().oneshot(get(
            &format!("/api/v1/ontologies/{oid}/classes/Dog/{path}"),
        )).await.unwrap();
        assert_eq!(resp.status(), StatusCode::OK, "{path} should succeed");
    }
    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_neighborhood_query() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("neighborhood");
    seed_class(&pool, &oid, "owl:Thing", None).await;
    seed_class(&pool, &oid, "Person", Some("owl:Thing")).await;
    seed_class(&pool, &oid, "Employee", Some("Person")).await;

    let resp = app.clone().oneshot(get(
        &format!("/api/v1/ontologies/{oid}/classes/Employee/neighborhood"),
    )).await.unwrap();
    assert_eq!(resp.status(), StatusCode::OK);

    let mut result = pool.graph().execute(
        neo4rs::query("MATCH (c:Class {ontology_id:$id,id:'Employee'})-[:CHILD_OF]->(p) RETURN p.id AS pid")
            .param("id", oid.clone()),
    ).await.unwrap();
    assert_eq!(result.next().await.unwrap().unwrap().get::<String>("pid").unwrap(), "Person");
    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_root_classes_endpoint() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("root");
    seed_class(&pool, &oid, "owl:Thing", None).await;

    let resp = app.clone().oneshot(get(
        &format!("/api/v1/ontologies/{oid}/classes/root"),
    )).await.unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_breadcrumb_endpoint() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("breadcrumb");
    seed_class(&pool, &oid, "owl:Thing", None).await;
    seed_class(&pool, &oid, "Animal", Some("owl:Thing")).await;
    seed_class(&pool, &oid, "Dog", Some("Animal")).await;

    let resp = app.clone().oneshot(get(
        &format!("/api/v1/ontologies/{oid}/classes/Dog/breadcrumb"),
    )).await.unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
    common::clean_ontology(&pool, &oid).await;
}
