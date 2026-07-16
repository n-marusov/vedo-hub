//! GraphQL schema assembly.
//!
//! Combines the query and mutation roots into a single executable schema.

use async_graphql::{EmptySubscription, Schema};

use super::mutation::MutationRoot;
use super::query::QueryRoot;

/// The composed GraphQL schema type.
pub type OntologySchema = Schema<QueryRoot, MutationRoot, EmptySubscription>;

/// Builds the GraphQL schema with query and mutation roots.
/// Write operations (class/property/individual CRUD) are served via REST
/// endpoints proxied through the API Gateway; GraphQL mutations are
/// intentionally limited to draft-state coordination for now.
pub fn build_schema() -> OntologySchema {
    Schema::build(
        QueryRoot::default(),
        MutationRoot::default(),
        EmptySubscription,
    )
    .enable_federation()
    .finish()
}
