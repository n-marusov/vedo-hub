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

// ── SDL contract (committed schema.graphql) ────────────────────────────────

/// The committed `schema.graphql` artifact MUST match the schema generated
/// from code. This is the canonical contract consumed by frontend codegen.
///
/// Regenerate the artifact when the schema changes:
///
/// ```text
/// EXPORT_SDL=1 cargo test --lib test_committed_sdl_matches_generated
/// ```
#[test]
fn test_committed_sdl_matches_generated() {
    let generated = super::schema::schema_sdl();
    assert!(!generated.trim().is_empty(), "schema SDL must not be empty");

    let committed_path = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("schema.graphql");

    // Opt-in regeneration: write the canonical SDL from code so the artifact
    // can be refreshed deterministically (see doc comment above).
    if std::env::var_os("EXPORT_SDL").is_some() {
        std::fs::write(&committed_path, &generated)
            .expect("failed to write schema.graphql from generated SDL");
    }

    let committed = std::fs::read_to_string(&committed_path).unwrap_or_else(|e| {
        panic!(
            "schema.graphql missing at {} — regenerate with EXPORT_SDL=1 cargo test --lib test_committed_sdl_matches_generated: {e}",
            committed_path.display()
        )
    });

    assert_eq!(
        committed, generated,
        "schema.graphql is out of date — regenerate with EXPORT_SDL=1 cargo test --lib test_committed_sdl_matches_generated"
    );
}

// ── Introspection JSON contract (committed introspection.json) ──────────────

/// The committed `introspection.json` artifact is a machine-readable snapshot
/// of the schema (used by breaking-change diff tools and codegen tooling that
/// prefer JSON over SDL). It MUST match a fresh introspection of the schema.
///
/// Regenerate the artifact when the schema changes:
///
/// ```text
/// EXPORT_SDL=1 cargo test --lib test_committed_introspection_matches_generated
/// ```
#[tokio::test]
async fn test_committed_introspection_matches_generated() {
    let schema = super::schema::build_schema();
    let response = schema
        .execute(
            r#"{
            __schema {
              queryType { name }
              mutationType { name }
              subscriptionType { name }
              types {
                kind
                name
                description
                fields(includeDeprecated: true) {
                  name
                  description
                  isDeprecated
                  deprecationReason
                  type { kind name ofType { kind name ofType { kind name ofType { kind name } } } }
                  args { name description type { kind name ofType { kind name ofType { kind name } } } defaultValue }
                }
                inputFields { name description type { kind name ofType { kind name ofType { kind name } } } defaultValue }
                interfaces { kind name }
                enumValues(includeDeprecated: true) { name description isDeprecated deprecationReason }
                possibleTypes { kind name }
              }
              directives { name description locations args { name description type { kind name ofType { kind name ofType { kind name } } } defaultValue } isRepeatable }
            }
          }"#,
        )
        .await;

    assert!(
        response.errors.is_empty(),
        "introspection query returned errors: {:?}",
        response.errors
    );
    let mut data = response
        .data
        .into_json()
        .expect("introspection result must be JSON");
    // Sort type names so the snapshot is deterministic regardless of registry order.
    let data_obj = data.as_object_mut().expect("data must be an object");
    let schema_obj = data_obj["__schema"]
        .as_object_mut()
        .expect("__schema must be an object");
    if let Some(types) = schema_obj.get_mut("types").and_then(|t| t.as_array_mut()) {
        types.sort_by(|a, b| {
            a.as_object()
                .and_then(|o| o.get("name"))
                .and_then(|n| n.as_str())
                .unwrap_or_default()
                .cmp(
                    b.as_object()
                        .and_then(|o| o.get("name"))
                        .and_then(|n| n.as_str())
                        .unwrap_or_default(),
                )
        });
    }
    let generated =
        serde_json::to_string_pretty(&data).expect("introspection must serialize to JSON");

    let committed_path =
        std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("introspection.json");

    // Opt-in regeneration (same pattern as the SDL drift test).
    if std::env::var_os("EXPORT_SDL").is_some() {
        std::fs::write(&committed_path, format!("{generated}\n"))
            .expect("failed to write introspection.json from generated data");
    }

    let committed = std::fs::read_to_string(&committed_path).unwrap_or_else(|e| {
        panic!(
            "introspection.json missing at {} — regenerate with EXPORT_SDL=1 cargo test --lib test_committed_introspection_matches_generated: {e}",
            committed_path.display()
        )
    });

    assert_eq!(
        committed,
        format!("{generated}\n"),
        "introspection.json is out of date — regenerate with EXPORT_SDL=1 cargo test --lib test_committed_introspection_matches_generated"
    );
}
