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

use std::sync::Arc;

use async_graphql::{Context, InputObject, Object, Result};
use chrono::Utc;

use super::types::{GqlDraftUpdateResult, GqlMember};

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

    // ── Organization model mutations ─────────────────────────────────

    /// Updates a member's role in an ontology scope.
    async fn update_member_role(
        &self,
        ctx: &Context<'_>,
        ontology_id: String,
        user_id: String,
        role: String,
    ) -> Result<GqlMember> {
        use crate::AppState;
        let state = ctx
            .data::<Arc<AppState>>()
            .map_err(|e| async_graphql::Error::new(format!("Failed to access app state: {e:?}")))?;

        let scope = format!("ontology/{}", ontology_id);
        let member = state
            .auth_client
            .update_member_role(&scope, &user_id, &role)
            .await
            .map_err(|e| async_graphql::Error::new(format!("Failed to update member role: {e}")))?;

        Ok(GqlMember {
            user_id: member.user_id,
            scope: member.scope,
            role: member.role,
            username: None,
            avatar_url: None,
            added_at: None,
        })
    }

    /// Removes a member from an ontology scope.
    async fn remove_member(
        &self,
        ctx: &Context<'_>,
        ontology_id: String,
        user_id: String,
    ) -> Result<bool> {
        use crate::AppState;
        let state = ctx
            .data::<Arc<AppState>>()
            .map_err(|e| async_graphql::Error::new(format!("Failed to access app state: {e:?}")))?;

        let scope = format!("ontology/{}", ontology_id);
        state
            .auth_client
            .remove_member(&scope, &user_id)
            .await
            .map_err(|e| async_graphql::Error::new(format!("Failed to remove member: {e}")))?;

        Ok(true)
    }
}
