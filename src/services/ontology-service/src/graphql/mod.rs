//! GraphQL schema module for ontology-service.
//!
//! Provides async-graphql types and resolvers that wrap the existing
//! repository layer, enabling the frontend Apollo Client to fetch
//! ontology data via GraphQL queries through the API Gateway.

pub mod query;
pub mod schema;
pub mod types;
