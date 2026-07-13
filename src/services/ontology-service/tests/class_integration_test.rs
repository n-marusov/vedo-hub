//! Integration tests for class (TBox) CRUD operations.
//!
//! Requires running Neo4j. Set NEO4J_TEST_URI env var to enable.

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

fn get(uri: &str) -> Request<Body> {
    Request::builder()
        .method(Method::GET)
        .uri(uri)
        .body(Body::empty())
        .unwrap()
}

fn delete(uri: &str) -> Request<Body> {
    Request::builder()
        .method(Method::DELETE)
        .uri(uri)
        .body(Body::empty())
        .unwrap()
}

fn put_json(uri: &str, body: &str) -> Request<Body> {
    Request::builder()
        .method(Method::PUT)
        .uri(uri)
        .header("Content-Type", "application/json")
        .body(Body::from(body.to_string()))
        .unwrap()
}

async fn body_text(response: axum::response::Response) -> String {
    String::from_utf8_lossy(
        &axum::body::to_bytes(response.into_body(), usize::MAX)
            .await
            .unwrap(),
    )
    .to_string()
}

fn ontology_url(o: &str) -> String {
    format!("/api/v1/ontologies/{o}")
}

#[tokio::test]
async fn test_create_class_creates_node_in_neo4j() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("create_class");

    let resp = app
        .clone()
        .oneshot(post_json(
            &(ontology_url(&oid) + "/classes"),
            r#"{"id":"Person","label":"Person","comment":"A person entity"}"#,
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);

    let mut result = pool
        .graph()
        .execute(
            neo4rs::query(
                "MATCH (c:Class) WHERE c.ontology_id = $id AND c.class_id = $cid RETURN c",
            )
            .param("id", oid.clone())
            .param("cid", "Person".to_string()),
        )
        .await
        .unwrap();
    let row = result.next().await.unwrap();
    assert!(row.is_some(), "Class should exist in Neo4j");
    let node: neo4rs::Node = row.unwrap().get("c").unwrap();
    assert_eq!(node.get::<String>("label").unwrap(), "Person");

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_get_class_returns_correct_data() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("get_class");

    let _ = pool
        .graph()
        .execute(
            neo4rs::query("CREATE (c:Class {ontology_id: $id, class_id: $cid, label: $label})")
                .param("id", oid.clone())
                .param("cid", "Vehicle".to_string())
                .param("label", "Vehicle".to_string()),
        )
        .await;

    let resp = app
        .clone()
        .oneshot(get(&format!("/api/v1/ontologies/{oid}/classes/Vehicle")))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
    let body = body_text(resp).await;
    assert!(body.contains("Vehicle"), "response should contain Vehicle");

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_update_class_modifies_node_in_neo4j() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("update_class");

    let _ = pool
        .graph()
        .execute(
            neo4rs::query(
                "CREATE (c:Class {ontology_id: $id, class_id: $cid, label: $label, comment: $comment})",
            )
            .param("id", oid.clone())
            .param("cid", "Task".to_string())
            .param("label", "Task".to_string())
            .param("comment", "".to_string()),
        )
        .await;

    let resp = app
        .clone()
        .oneshot(put_json(
            &format!("/api/v1/ontologies/{oid}/classes/Task"),
            r#"{"label":"WorkItem","comment":"A unit of work"}"#,
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);

    let mut result = pool
        .graph()
        .execute(
            neo4rs::query(
                "MATCH (c:Class {ontology_id: $id, class_id: 'Task'}) RETURN c.label AS lbl, c.comment AS cmt",
            )
            .param("id", oid.clone()),
        )
        .await
        .unwrap();
    let row = result.next().await.unwrap().unwrap();
    assert_eq!(row.get::<String>("lbl").unwrap(), "WorkItem");
    assert_eq!(row.get::<String>("cmt").unwrap(), "A unit of work");

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_delete_class_removes_node_from_neo4j() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("delete_class");

    let _ = pool
        .graph()
        .execute(
            neo4rs::query("CREATE (c:Class {ontology_id: $id, class_id: $cid, label: $label})")
                .param("id", oid.clone())
                .param("cid", "Obsolete".to_string())
                .param("label", "Obsolete".to_string()),
        )
        .await;

    let resp = app
        .clone()
        .oneshot(delete(&format!(
            "/api/v1/ontologies/{oid}/classes/Obsolete"
        )))
        .await
        .unwrap();
    assert!(resp.status().is_success());

    let mut result = pool
        .graph()
        .execute(
            neo4rs::query(
                "MATCH (c:Class {ontology_id: $id, class_id: 'Obsolete'}) RETURN count(c) AS cnt",
            )
            .param("id", oid.clone()),
        )
        .await
        .unwrap();
    let row = result.next().await.unwrap().unwrap();
    assert_eq!(row.get::<i64>("cnt").unwrap(), 0);

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_list_classes_returns_data() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("list_classes");

    for cls in &["ClassA", "ClassB", "ClassC"] {
        let _ = pool
            .graph()
            .execute(
                neo4rs::query("CREATE (c:Class {ontology_id: $id, class_id: $cid, label: $label})")
                    .param("id", oid.clone())
                    .param("cid", cls.to_string())
                    .param("label", cls.to_string()),
            )
            .await;
    }

    let resp = app
        .clone()
        .oneshot(get(&format!("/api/v1/ontologies/{oid}/classes?per_page=2")))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
    let body = body_text(resp).await;
    assert!(!body.is_empty(), "response should contain class data");

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_create_duplicate_class_returns_error() {
    common::skip_if_no_neo4j();
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("duplicate_class");

    let _ = pool
        .graph()
        .execute(
            neo4rs::query("CREATE (c:Class {ontology_id: $id, class_id: $cid, label: $label})")
                .param("id", oid.clone())
                .param("cid", "Unique".to_string())
                .param("label", "UniqueClass".to_string()),
        )
        .await;

    let resp = app
        .clone()
        .oneshot(post_json(
            &format!("/api/v1/ontologies/{oid}/classes"),
            r#"{"id":"Unique","label":"UniqueClass"}"#,
        ))
        .await
        .unwrap();
    assert!(
        resp.status().is_client_error() || resp.status() == StatusCode::CONFLICT,
        "duplicate class should error"
    );

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_create_class_missing_fields_returns_error() {
    common::skip_if_no_neo4j();
    let (app, _pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("missing_fields");

    let resp = app
        .clone()
        .oneshot(post_json(
            &format!("/api/v1/ontologies/{oid}/classes"),
            r#"{}"#,
        ))
        .await
        .unwrap();
    assert!(
        resp.status().is_client_error(),
        "missing fields should return 4xx"
    );
}
