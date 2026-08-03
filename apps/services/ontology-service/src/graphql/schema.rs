//! GraphQL schema assembly.
//!
//! Combines the query root into a single executable schema.
//! GraphQL in VEDO Core is **graph-only navigation** — no Mutation root,
//! no non-graph Query resolvers. All write operations and non-graph reads
//! go through REST endpoints under `/api/v1/...`.

use async_graphql::{EmptyMutation, EmptySubscription, Schema};

use super::query::QueryRoot;
use super::types::GqlEntity;

/// The composed GraphQL schema type.
pub type OntologySchema = Schema<QueryRoot, EmptyMutation, EmptySubscription>;

/// Builds the GraphQL schema with query root only.
///
/// REST/GraphQL boundary (see ADR-DES.API.rest-graphql-mutation-boundary.md
/// and ADR-DES.API.graphql-sparql-split-strategy.md):
///   - GraphQL exposes **graph navigation queries only** (class, classes, classTree,
///     classAncestors, classDescendants, graphNeighborhood, autocompleteClasses,
///     property, properties, individual, individuals — 11 resolvers).
///   - **Mutation root is empty** — all write operations go through REST.
///   - **Non-graph Query resolvers removed** — ontology(id), commits, branch,
///     branches, groups, projects, members are now served by REST endpoints.
///   - REST enforces Idempotency-Key, auth middleware, audit log and
///     `CircuitBreakerMiddleware` `DoS` protection.
pub fn build_schema() -> OntologySchema {
    Schema::build(QueryRoot, EmptyMutation, EmptySubscription)
        // Register the Entity interface explicitly — no resolver returns it
        // directly, but its implementors (Class/Property/Individual) are used.
        .register_output_type::<GqlEntity>()
        .enable_federation()
        .finish()
}

/// Returns the canonical GraphQL SDL of the ontology-service schema.
///
/// This is the single source of truth for the committed `schema.graphql`
/// artifact: the frontend codegen (G2) and the drift test both consume it.
/// The output is deterministic — async-graphql renders registry types from a
/// `BTreeMap` and object fields from an insertion-ordered `IndexMap`.
pub fn schema_sdl() -> String {
    build_schema().sdl()
}
