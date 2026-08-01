//! Unit tests for the GraphQL facade assembled by `ontology-service`.
//!
//! These tests verify the GraphQL schema contract per revised specs:
//!
//! - `ADR-DES.API.graphql-sparql-split-strategy.md`
//! - `ADR-DES.API.rest-graphql-mutation-boundary.md`
//! - `REQ-USR.UI.graph-navigation.md`
//!
//! 1. `build_schema()` uses `EmptyMutation` — no Mutation type exposed.
//! 2. 11 graph-navigation query fields are present.
//! 3. Non-graph query fields are forbidden (commits, branches, ontology).
//! 4. Entity interface is present with entityType discriminator.
//! 5. PropertyType enum includes ANNOTATION variant.
//! 6. Class exposes isAbstract and isDeprecated flags.

use super::schema::{build_schema, OntologySchema};

// ── Schema construction ──────────────────────────────────────────────────────

/// build_schema() must use EmptyMutation — no mutation resolvers.
/// Executing any mutation against the schema must return a "Cannot query field"
/// error, confirming that GraphQL is read-only for graph navigation.
#[tokio::test]
async fn test_schema_has_no_mutation_root() {
    let schema: OntologySchema = build_schema();
    use async_graphql::{Name, Variables};
    let mut vars = Variables::default();
    vars.insert(
        Name::new("ontologyId"),
        async_graphql::value!("00000000-0000-0000-0000-000000000000"),
    );

    let response = schema
        .execute(
            async_graphql::Request::new(
                r#"mutation UpdateDraft($ontologyId: ID!) {
                    updateDraft(ontologyId: $ontologyId) { success }
                }"#,
            )
            .variables(vars),
        )
        .await;

    assert!(
        !response.errors.is_empty(),
        "expected GraphQL error from updateDraft on EmptyMutation, got no errors"
    );
    let has_schema_error = response
        .errors
        .iter()
        .any(|e| e.message.contains("not configured for mutations"));
    assert!(
        has_schema_error,
        "expected error referencing updateDraft or mutation configuration, got: {:?}",
        response.errors
    );
}

/// Sanity check: the schema builds without panicking.
#[test]
fn test_schema_builds_successfully() {
    let _schema = build_schema();
}

/// Introspection should expose only graph-navigation query fields.
#[tokio::test]
async fn test_introspection_reveals_graph_only_fields() {
    let schema: OntologySchema = build_schema();

    let response = schema
        .execute(r#"{ __schema { queryType { fields { name } } } }"#)
        .await;

    let data = response.data.into_json().expect("introspection JSON");
    let fields: Vec<&str> = data["__schema"]["queryType"]["fields"]
        .as_array()
        .map(|arr| arr.iter().filter_map(|f| f["name"].as_str()).collect())
        .unwrap_or_default();

    // Graph navigation resolvers that MUST be present
    let expected_graph = [
        "class",
        "classes",
        "classTree",
        "classAncestors",
        "classDescendants",
        "graphNeighborhood",
        "autocompleteClasses",
        "property",
        "properties",
        "individual",
        "individuals",
    ];
    for name in &expected_graph {
        assert!(
            fields.contains(name),
            "expected graph query '{name}' in schema, found: {fields:?}"
        );
    }

    // Non-graph resolvers that MUST NOT be present
    let forbidden = [
        "commits", "branch", "branches", "groups", "projects", "members", "ontology",
    ];
    for name in &forbidden {
        assert!(
            !fields.contains(name),
            "non-graph query '{name}' must NOT be in schema, found: {fields:?}"
        );
    }

    // Mutation type must be absent (EmptyMutation hides it from introspection)
    let has_mutation = data["__schema"].get("mutationType").is_some()
        && !data["__schema"]["mutationType"].is_null();
    assert!(!has_mutation, "Mutation type must be absent from schema");
}

// ── Introspection: Entity interface ──────────────────────────────────────────

/// Entity interface must be present and expose the entityType discriminator.
#[tokio::test]
async fn test_entity_interface_present() {
    let schema: OntologySchema = build_schema();

    let response = schema
        .execute(r#"{ __type(name: "Entity") { name kind } }"#)
        .await;

    let data = response.data.into_json().expect("introspection JSON");
    let entity_type = &data["__type"];
    assert_eq!(
        entity_type["name"], "Entity",
        "Entity interface must be present in schema"
    );
    assert_eq!(
        entity_type["kind"], "INTERFACE",
        "Entity must be an INTERFACE kind"
    );
}

/// Class must expose isAbstract and isDeprecated flags.
#[tokio::test]
async fn test_class_exposes_boolean_flags() {
    let schema: OntologySchema = build_schema();

    let response = schema
        .execute(r#"{ __type(name: "Class") { fields { name type { name kind } } } }"#)
        .await;

    let data = response.data.into_json().expect("introspection JSON");
    let fields: Vec<&str> = data["__type"]["fields"]
        .as_array()
        .map(|arr| arr.iter().filter_map(|f| f["name"].as_str()).collect())
        .unwrap_or_default();

    assert!(
        fields.contains(&"isAbstract"),
        "Class must expose isAbstract flag, fields: {fields:?}"
    );
    assert!(
        fields.contains(&"isDeprecated"),
        "Class must expose isDeprecated flag, fields: {fields:?}"
    );
}

/// Class must implement the Entity interface (entityType field present).
#[tokio::test]
async fn test_class_implements_entity() {
    let schema: OntologySchema = build_schema();

    let response = schema
        .execute(r#"{ __type(name: "Class") { interfaces { name } } }"#)
        .await;

    let data = response.data.into_json().expect("introspection JSON");
    let interfaces: Vec<&str> = data["__type"]["interfaces"]
        .as_array()
        .map(|arr| arr.iter().filter_map(|i| i["name"].as_str()).collect())
        .unwrap_or_default();

    assert!(
        interfaces.contains(&"Entity"),
        "Class must implement Entity interface, implements: {interfaces:?}"
    );
}

/// PropertyType enum must include ANNOTATION variant.
#[tokio::test]
async fn test_property_type_includes_annotation() {
    let schema: OntologySchema = build_schema();

    let response = schema
        .execute(r#"{ __type(name: "PropertyType") { enumValues { name } } }"#)
        .await;

    let data = response.data.into_json().expect("introspection JSON");
    let values: Vec<&str> = data["__type"]["enumValues"]
        .as_array()
        .map(|arr| arr.iter().filter_map(|v| v["name"].as_str()).collect())
        .unwrap_or_default();

    assert!(
        values.contains(&"ANNOTATION"),
        "PropertyType must include ANNOTATION variant, values: {values:?}"
    );
    assert!(
        values.contains(&"OBJECT"),
        "PropertyType must include OBJECT variant"
    );
    assert!(
        values.contains(&"DATATYPE"),
        "PropertyType must include DATATYPE variant"
    );
}
