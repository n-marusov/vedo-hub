//! GraphQL schema assembly.
//!
//! Combines the query and mutation roots into a single executable schema.

use async_graphql::{EmptySubscription, Schema};

use super::mutation::MutationRoot;
use super::query::QueryRoot;

/// The composed GraphQL schema type.
pub type OntologySchema = Schema<QueryRoot, MutationRoot, EmptySubscription>;

/// Builds the GraphQL schema with query root only.
///
/// REST/GraphQL boundary (see ADR-DES.API.rest-graphql-mutation-boundary.md):
///   - GraphQL exposes **navigation/read queries only** (Query root).
///   - **Any GraphQL mutation is forbidden.** All write operations, including
///     CRUD of classes/properties/individuals, draft-state coordination,
///     membership management, validation, comments and versioning, MUST be
///     performed through REST endpoints under `/api/v1/...` proxied by the
///     API Gateway. REST enforces Idempotency-Key, auth middleware, audit log
///     and CircuitBreakerMiddleware DoS protection — none of which apply to a
///     GraphQL mutation path.
///   - `sparqlQuery` is forbidden in GraphQL: SPARQL execution MUST go through
///     `POST /api/v1/sparql` so the CircuitBreakerMiddleware DoS guard cannot
///     be bypassed.
///
/// NOTE: `MutationRoot` below still holds `updateDraft`, `updateMemberRole`
/// and `removeMember` as deprecated placeholders pending REST migration.
/// They are NOT part of the public contract — see the ADR for the plan to
/// remove `MutationRoot` entirely by switching to `EmptyMutation`.
pub fn build_schema() -> OntologySchema {
    Schema::build(
        QueryRoot::default(),
        MutationRoot::default(),
        EmptySubscription,
    )
    .enable_federation()
    .finish()
}
