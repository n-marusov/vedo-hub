//! GraphQL schema module for ontology-service.
//!
//! Provides async-graphql types and resolvers that wrap the existing
//! repository layer, enabling the frontend Apollo Client to fetch
//! ontology data via GraphQL queries through the API Gateway.
//!
//! Mutations are intentionally minimal (`updateDraft`) — write operations
//! are served via REST endpoints proxied through the API Gateway.

pub mod mutation;
pub mod query;
pub mod schema;
pub mod types;
pub mod versioning_client;

#[cfg(test)]
mod tests;
