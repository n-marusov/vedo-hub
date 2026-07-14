//! GraphQL mutation root.
//!
//! Currently exposes only the `updateDraft` mutation, which marks the
//! workspace as dirty. The mutation is intentionally thin: it returns the
//! current server timestamp so the frontend can correlate draft-state
//! changes across the toolbar (`useVersionContext`).
//!
//! Other write operations (class/property/individual CRUD) are served via
//! the existing REST endpoints and proxied through the API Gateway. They
//! can be mirrored to GraphQL mutations later without breaking the schema.

use async_graphql::{Context, InputObject, Object, Result};
use chrono::Utc;

use super::types::GqlDraftUpdateResult;

/// Input for the `updateDraft` mutation.
///
/// The `changes` field is a free-form JSON blob describing the staged
/// mutations the user intends to commit later. The backend does not interpret
/// the contents — it only records the fact that the workspace is now dirty.
#[derive(InputObject)]
pub struct DraftInput {
    /// Free-form description of pending changes (JSON-encoded by the client).
    pub changes: String,
}

/// Mutation root for ontology-service GraphQL schema.
#[derive(Default)]
pub struct MutationRoot;

#[Object]
impl MutationRoot {
    /// Marks the workspace as dirty and returns the timestamp of the update.
    ///
    /// The ontology dirty state is tracked by the frontend (`useVersionContext`)
    /// and persisted in browser storage. The backend only echoes the event so
    /// multiple collaborating clients can converge on a shared clock.
    async fn update_draft(
        &self,
        _ctx: &Context<'_>,
        _ontology_id: String,
        _changes: DraftInput,
    ) -> Result<GqlDraftUpdateResult> {
        tracing::debug!(
            ontology_id = %_ontology_id,
            changes_len = _changes.changes.len(),
            "[FIX] update_draft mutation invoked"
        );

        let timestamp = Utc::now().to_rfc3339();
        tracing::info!(
            ontology_id = %_ontology_id,
            timestamp = %timestamp,
            "Draft marked dirty"
        );

        Ok(GqlDraftUpdateResult {
            success: true,
            timestamp,
        })
    }
}
