//! Integration tests for property CRUD operations.
//!
//! Validates: REQ-FUN.API.property-crud
//!
//! Requires running Neo4j. Run via `make test-integration-rust` (auto-starts Neo4j via Docker Compose).
//! Set NEO4J_TEST_URI env var to run manually.

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

fn post_json(uri: &str, body: &str) -> Request<Body> {
    request(Method::POST, uri, Some(body))
}

fn delete_req(uri: &str) -> Request<Body> {
    request(Method::DELETE, uri, None)
}

fn oid_url(oid: &str, path: &str) -> String {
    format!("/api/v1/ontologies/{oid}{path}")
}

#[tokio::test]
async fn test_create_object_property_stores_in_neo4j() {
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("create_obj_prop");

    // Create class first: handler validates domain/range classes exist
    common::execute_query(
        &pool,
        neo4rs::query("CREATE (c:Class {ontology_id:$id, id:$cid, label:$label})")
            .param("id", oid.clone())
            .param("cid", "Person".to_string())
            .param("label", "Person".to_string()),
    )
    .await;

    let resp = app
        .clone()
        .oneshot(post_json(
            &oid_url(&oid, "/properties"),
            r#"{"id":"hasParent","label":"has parent","property_type":"object","domains":["Person"],"ranges":["Person"]}"#,
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::CREATED);

    let mut result = pool
        .graph()
        .execute(
            neo4rs::query("MATCH (p:Property) WHERE p.ontology_id=$id AND p.id=$pid RETURN p")
                .param("id", oid.clone())
                .param("pid", "hasParent".to_string()),
        )
        .await
        .unwrap();
    let row = result.next().await.unwrap();
    assert!(row.is_some());
    let node: neo4rs::Node = row.unwrap().get("p").unwrap();
    assert_eq!(node.get::<String>("property_type").unwrap(), "object");

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_create_datatype_property_stores_in_neo4j() {
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("create_dt_prop");

    // Create class first: handler validates domain class exists
    common::execute_query(
        &pool,
        neo4rs::query("CREATE (c:Class {ontology_id:$id, id:$cid, label:$label})")
            .param("id", oid.clone())
            .param("cid", "Person".to_string())
            .param("label", "Person".to_string()),
    )
    .await;

    let resp = app
        .clone()
        .oneshot(post_json(
            &oid_url(&oid, "/properties"),
            r#"{"id":"age","label":"age","property_type":"datatype","domains":["Person"],"xsd_type":"integer"}"#,
        ))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::CREATED);

    let mut result = pool
        .graph()
        .execute(
            neo4rs::query(
                "MATCH (p:Property {ontology_id:$id,id:'age'}) RETURN p.property_type AS t",
            )
            .param("id", oid.clone()),
        )
        .await
        .unwrap();
    let row = result.next().await.unwrap().unwrap();
    assert_eq!(row.get::<String>("t").unwrap(), "datatype");

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_get_property_with_domain_range() {
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("get_prop");

    common::execute_query(
        &pool,
        neo4rs::query(
            r#"CREATE (p:Property {ontology_id:$id,id:'worksFor',label:'works for',property_type:'object'})"#,
        )
        .param("id", oid.clone()),
    ).await;
    common::execute_query(
        &pool,
        neo4rs::query(r#"CREATE (:Class {ontology_id:$id,id:'Person',label:'Person'})"#)
            .param("id", oid.clone()),
    )
    .await;
    common::execute_query(
        &pool,
        neo4rs::query(
            r#"CREATE (:Class {ontology_id:$id,id:'Organization',label:'Organization'})"#,
        )
        .param("id", oid.clone()),
    )
    .await;
    common::execute_query(
        &pool,
        neo4rs::query(
            r#"MATCH (p:Property {ontology_id:$id,id:'worksFor'})
               MATCH (d:Class {ontology_id:$id,id:'Person'})
               MATCH (r:Class {ontology_id:$id,id:'Organization'})
               CREATE (p)-[:DOMAIN]->(d)
               CREATE (p)-[:RANGE]->(r)"#,
        )
        .param("id", oid.clone()),
    )
    .await;

    let resp = app
        .clone()
        .oneshot(get(&oid_url(&oid, "/properties/worksFor")))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);
    let body = String::from_utf8_lossy(
        &axum::body::to_bytes(resp.into_body(), usize::MAX)
            .await
            .unwrap(),
    )
    .to_string();
    assert!(body.contains("worksFor"));
    assert!(body.contains("Person"));
    assert!(body.contains("Organization"));

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_list_properties_returns_data() {
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("list_props");

    common::execute_query(
        &pool,
        neo4rs::query(r#"CREATE (p:Property {ontology_id:$id,id:'rel1',label:'Rel1',property_type:'object'})"#)
            .param("id", oid.clone()),
    ).await;
    common::execute_query(
        &pool,
        neo4rs::query(r#"CREATE (p:Property {ontology_id:$id,id:'dt1',label:'Dt1',property_type:'datatype'})"#)
            .param("id", oid.clone()),
    ).await;

    let resp = app
        .clone()
        .oneshot(get(&oid_url(&oid, "/properties")))
        .await
        .unwrap();
    assert_eq!(resp.status(), StatusCode::OK);

    common::clean_ontology(&pool, &oid).await;
}

#[tokio::test]
async fn test_delete_property_removes_from_neo4j() {
    let (app, pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("delete_prop");

    common::execute_query(
        &pool,
        neo4rs::query(r#"CREATE (p:Property {ontology_id:$id,id:'tempProp',label:'Temp'})"#)
            .param("id", oid.clone()),
    )
    .await;

    let resp = app
        .clone()
        .oneshot(delete_req(&oid_url(&oid, "/properties/tempProp")))
        .await
        .unwrap();
    assert!(resp.status().is_success());

    let mut result = pool
        .graph()
        .execute(
            neo4rs::query(
                "MATCH (p:Property {ontology_id:$id,id:'tempProp'}) RETURN count(p) AS cnt",
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
async fn test_create_property_missing_required_field_returns_error() {
    let (app, _pool) = common::create_test_app().await;
    let oid = common::test_ontology_id("prop_missing");

    let resp = app
        .clone()
        .oneshot(post_json(
            &oid_url(&oid, "/properties"),
            r#"{"property_type":"object"}"#,
        ))
        .await
        .unwrap();
    assert!(resp.status().is_client_error());
}
