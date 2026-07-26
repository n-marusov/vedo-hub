//! GraphQL schema module for ontology-service.
//!
//! Provides async-graphql types and resolvers that wrap the existing
//! repository layer, enabling the frontend Apollo Client to fetch
//! ontology data via GraphQL queries through the API Gateway.
//!
//! GraphQL in VEDO Core is **graph-only navigation**. All write operations
//! and non-graph reads are served via REST endpoints proxied through the
//! API Gateway.

pub mod query;
pub mod schema;
pub mod types;
pub mod versioning_client;

#[cfg(test)]
mod tests;
