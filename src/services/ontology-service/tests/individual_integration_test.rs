//! Integration tests for individual (ABox) CRUD operations.
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
    if body.is_some() {
        builder = builder.header("Content-Type", "application/json");
    }
    builder
        .body(
            body.map(|b| Body::from(b.to_string()))
                .unwrap_or(Body::empty()),
        )
        .unwrap()
}

fn get(uri: &str) -> Request<Body> {
    request(Method::GET, uri, None)
}
fn post_json(uri: &str, b: &str) -> Request<Body> {
    request(Method::POST, uri, Some(b))
}
fn put_json(uri: &str, b: &str) -> Request<Body> {
    request(Method::PUT, uri, Some(b))
}
fn delete_req(uri: &str) -> Request<Body> {
    request(Method::DELETE, uri, None)
}

async fn seed_class(pool: &ontology_service::neo4j::Neo4jPool, oid: &str, cid: &str) {
    let _ = pool
        .graph()
        .execute(
            neo4rs::query("CREATE (c:Class {ontology_id:$id,id:$cid,label:$label})")
                .param("id", oid.to_string())
                .param("cid", cid.to_string())
                .param("label", cid.to_string()),
        )
        .await;
}

#[tokio::test]
async fn test_create_individual_stores_in_neo4j() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("create_indiv");
    seed_class(&pool, &oid, "Person").await;

    let resp = app
        .clone()
        .oneshot(post_json(
            &format!("/api/v1/ontologies/{oid}/individuals"),
            r#"{"id":"john_doe","label":"John Doe","class_id":"Person"}"#,
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);

    let mut result = pool
        .graph()
        .execute(
            neo4rs::query("MATCH (i:Individual) WHERE i.ontology_id=$id AND i.id=$iid RETURN i")
                .param("id", oid.clone())
                .param("iid", "john_doe".to_string()),
        )
        .await
        .unwrap();
    assert!(result.next().await.unwrap().is_some());
    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_get_individual_returns_data() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("get_indiv");
    seed_class(&pool, &oid, "Person").await;
    let _ = pool
        .graph()
        .execute(
            neo4rs::query("CREATE (i:Individual {ontology_id:$id,id:'jane',label:'Jane'})")
                .param("id", oid.clone()),
        )
        .await;

    let resp = app
        .clone()
        .oneshot(get(&format!("/api/v1/ontologies/{oid}/individuals/jane")))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_update_individual_changes_label() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("update_indiv");
    seed_class(&pool, &oid, "Person").await;
    let _ = pool
        .graph()
        .execute(
            neo4rs::query("CREATE (i:Individual {ontology_id:$id,id:'bob',label:'Bob'})")
                .param("id", oid.clone()),
        )
        .await;

    let resp = app
        .clone()
        .oneshot(put_json(
            &format!("/api/v1/ontologies/{oid}/individuals/bob"),
            r#"{"label":"Robert"}"#,
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);

    let mut result = pool
        .graph()
        .execute(
            neo4rs::query("MATCH (i:Individual {ontology_id:$id,id:'bob'}) RETURN i.label AS lbl")
                .param("id", oid.clone()),
        )
        .await
        .unwrap();
    assert_eq!(
        result
            .next()
            .await
            .unwrap()
            .unwrap()
            .get::<String>("lbl")
            .unwrap(),
        "Robert"
    );
    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_list_individuals_returns_data() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("list_indiv");
    seed_class(&pool, &oid, "Person").await;
    for name in &["Alice", "Bob"] {
        let _ = pool
            .graph()
            .execute(
                neo4rs::query("CREATE (i:Individual {ontology_id:$id,id:$iid,label:$label})")
                    .param("id", oid.clone())
                    .param("iid", name.to_lowercase())
                    .param("label", name.to_string()),
            )
            .await;
    }

    let resp = app
        .clone()
        .oneshot(get(&format!("/api/v1/ontologies/{oid}/individuals")))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_delete_individual_removes_from_neo4j() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("delete_indiv");
    seed_class(&pool, &oid, "Person").await;
    let _ = pool
        .graph()
        .execute(
            neo4rs::query("CREATE (i:Individual {ontology_id:$id,id:'temp',label:'Temp'})")
                .param("id", oid.clone()),
        )
        .await;

    let resp = app
        .clone()
        .oneshot(delete_req(&format!(
            "/api/v1/ontologies/{oid}/individuals/temp"
        )))
        .await
        .unwrap();
    assert!(resp.status().is_success());

    let mut result = pool
        .graph()
        .execute(
            neo4rs::query(
                "MATCH (i:Individual {ontology_id:$id,id:'temp'}) RETURN count(i) AS cnt",
            )
            .param("id", oid.clone()),
        )
        .await
        .unwrap();
    assert_eq!(
        result
            .next()
            .await
            .unwrap()
            .unwrap()
            .get::<i64>("cnt")
            .unwrap(),
        0
    );
    common::clean_ontology(&pool, &oid).await;
}
