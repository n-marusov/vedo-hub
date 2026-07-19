//! Auth service HTTP client for organizational model queries.
//!
//! Provides methods to query the auth-service REST API for groups, projects,
//! and members data. Used by GraphQL resolvers.
//!
//! Configuration via `AUTH_SERVICE_URL` env var (default: `http://auth-service:8081`).

pub mod auth_client;
pub use auth_client::*;
