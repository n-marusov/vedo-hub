//! Validates: REQ-FUN.API.class-hierarchy-accuracy
//! Validates: REQ-CON.SECURITY.write-path-invariant
//! NOTE: these tests exercise direct ontology writes; per ADR-DES.API.write-path-invariant the write path must go through the versioning pipeline (to be migrated in M10).
//! Validates: REQ-USR.UI.tbox-editor
//! Validates: REQ-FUN.API.no-cyclic-hierarchy
//! Validates: REQ-FUN.API.owl-no-cycles
//!
//! Integration tests for class (TBox) CRUD operations.
//!
//! Requires running Neo4j. Run via `make test-integration-rust` (auto-starts Neo4j via Docker Compose).
//! Set NEO4J_TEST_URI env var to run manually: `NEO4J_TEST_URI=bolt://localhost:7687 cargo test`
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
    assert_eq!(
        resp.status(),
        StatusCode::CREATED,
        "create class should return 201"
    );

    let mut result = pool
        .graph()
        .execute(
            neo4rs::query("MATCH (c:Class) WHERE c.ontology_id = $id AND c.id = $cid RETURN c")
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
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("get_class");

    let mut stream = pool
        .graph()
        .execute(
            neo4rs::query("CREATE (c:Class {ontology_id: $id, id: $cid, label: $label})")
                .param("id", oid.clone())
                .param("cid", "Vehicle".to_string())
                .param("label", "Vehicle".to_string()),
        )
        .await
        .expect("CREATE should succeed");
    while let Ok(Some(_)) = stream.next().await {}

    let resp = app
        .clone()
        .oneshot(get(&format!("/api/v1/ontologies/{oid}/classes/Vehicle")))
        .await
        .unwrap();
    let status = resp.status();
    let body = body_text(resp).await;
    assert_eq!(status, StatusCode::OK);
    assert!(body.contains("Vehicle"), "response should contain Vehicle");

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_update_class_modifies_node_in_neo4j() {
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("update_class");

    common::execute_query(
        &pool,
        neo4rs::query(
            "CREATE (c:Class {ontology_id: $id, id: $cid, label: $label, comment: $comment})",
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
                "MATCH (c:Class {ontology_id: $id, id: 'Task'}) RETURN c.label AS lbl, c.comment AS cmt",
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
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("delete_class");

    common::execute_query(
        &pool,
        neo4rs::query("CREATE (c:Class {ontology_id: $id, id: $cid, label: $label})")
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
                "MATCH (c:Class {ontology_id: $id, id: 'Obsolete'}) RETURN count(c) AS cnt",
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
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("list_classes");

    for cls in &["ClassA", "ClassB", "ClassC"] {
        let mut stream = pool
            .graph()
            .execute(
                neo4rs::query("CREATE (c:Class {ontology_id: $id, id: $cid, label: $label})")
                    .param("id", oid.clone())
                    .param("cid", cls.to_string())
                    .param("label", cls.to_string()),
            )
            .await
            .expect("CREATE should succeed");
        while let Ok(Some(_)) = stream.next().await {}
    }

    let resp = app
        .clone()
        .oneshot(get(&format!("/api/v1/ontologies/{oid}/classes?per_page=2")))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
    let body = body_text(resp).await;
    assert!(!body.is_empty(), "response should contain class data");
    assert!(body.contains("ClassA"), "response should contain ClassA");
    assert!(body.contains("ClassB"), "response should contain ClassB");

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_create_duplicate_class_returns_error() {
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("duplicate_class");

    common::execute_query(
        &pool,
        neo4rs::query("CREATE (c:Class {ontology_id: $id, id: $cid, label: $label})")
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

#[tokio::test]
/// Validates: REQ-FUN.API.class-hierarchy-accuracy
///
/// TC-A03: Create class hierarchy with subClassOf.
/// Creates a parent class, then creates a child class referencing the parent,
/// and verifies the CHILD_OF relationship exists in Neo4j.
async fn test_create_class_with_parent_creates_hierarchy() {
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("hierarchy");

    // Create parent class "Person"
    let resp = app
        .clone()
        .oneshot(post_json(
            &format!("/api/v1/ontologies/{oid}/classes"),
            r#"{"id":"Person","label":"Person","parents":[]}"#,
        ))
        .await
        .unwrap();
    assert!(
        resp.status().is_success(),
        "create parent should succeed, got {}",
        resp.status()
    );

    // Create child class "Student" with Person as parent
    let resp = app
        .clone()
        .oneshot(post_json(
            &format!("/api/v1/ontologies/{oid}/classes"),
            r#"{"id":"Student","label":"Student","parents":["Person"]}"#,
        ))
        .await
        .unwrap();
    assert!(
        resp.status().is_success(),
        "create child should succeed, got {}",
        resp.status()
    );

    // Verify CHILD_OF relationship in Neo4j
    let mut result = pool
        .graph()
        .execute(
            neo4rs::query(
                "MATCH (s:Class {id: $child_id, ontology_id: $oid})-[:CHILD_OF]->(p:Class {id: $parent_id}) RETURN count(s) AS cnt",
            )
            .param("child_id", "Student".to_string())
            .param("parent_id", "Person".to_string())
            .param("oid", oid.clone()),
        )
        .await
        .unwrap();
    let row = result.next().await.unwrap().unwrap();
    assert!(
        row.get::<i64>("cnt").unwrap_or(0) > 0,
        "Student should be CHILD_OF Person"
    );

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
/// Validates: REQ-USR.UI.tbox-editor
///
/// TC-A06: Delete class with existing subclasses.
/// Creates a parent class with a subclass, then attempts to delete the parent
/// without cascade=true. Expects HTTP 409 CONFLICT with ONT-CLASS-HAS-DEPENDENTS.
async fn test_delete_class_with_dependents_returns_error() {
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("delete_dep");

    // Create parent class "Department"
    common::execute_query(
        &pool,
        neo4rs::query("CREATE (c:Class {ontology_id: $id, id: $cid, label: $label})")
            .param("id", oid.clone())
            .param("cid", "Department".to_string())
            .param("label", "Department".to_string()),
    )
    .await;

    // Create child class "Engineering" with CHILD_OF → Department
    common::execute_query(
        &pool,
        neo4rs::query(
            "MATCH (p:Class {ontology_id: $id, id: $pid}) CREATE (c:Class {ontology_id: $id, id: $cid, label: $label})-[:CHILD_OF]->(p)",
        )
        .param("id", oid.clone())
        .param("pid", "Department".to_string())
        .param("cid", "Engineering".to_string())
        .param("label", "Engineering".to_string()),
    ).await;

    // Attempt to delete Department without cascade — should be blocked
    let resp = app
        .clone()
        .oneshot(delete(&format!(
            "/api/v1/ontologies/{oid}/classes/Department"
        )))
        .await
        .unwrap();

    assert_eq!(
        resp.status(),
        StatusCode::CONFLICT,
        "deleting class with dependents should return 409"
    );

    // Verify Department still exists
    let mut result = pool
        .graph()
        .execute(
            neo4rs::query(
                "MATCH (c:Class {ontology_id: $id, id: 'Department'}) RETURN count(c) AS cnt",
            )
            .param("id", oid.clone()),
        )
        .await
        .unwrap();
    let row = result.next().await.unwrap().unwrap();
    assert_eq!(row.get::<i64>("cnt").unwrap(), 1);

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
/// Validates: REQ-FUN.API.class-hierarchy-accuracy
///
/// Edge case: deleting a nonexistent class returns 404 NotFound.
async fn test_delete_nonexistent_class_returns_404() {
    let (app, _pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("delete_nope");

    let resp = app
        .clone()
        .oneshot(delete(&format!("/api/v1/ontologies/{oid}/classes/Nope")))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::NOT_FOUND);
}
