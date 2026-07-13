//! GraphQL schema assembly.
//!
//! Combines the query root into a single executable schema.
//! Mutations can be added here when needed.

use async_graphql::{EmptyMutation, EmptySubscription, Schema};

use super::query::QueryRoot;

/// The composed GraphQL schema type.
pub type OntologySchema = Schema<QueryRoot, EmptyMutation, EmptySubscription>;

/// Builds the GraphQL schema with query root only.
/// Mutations are handled via REST endpoints for now.
pub fn build_schema() -> OntologySchema {
    Schema::build(QueryRoot::default(), EmptyMutation, EmptySubscription)
        .enable_federation()
        .finish()
}
