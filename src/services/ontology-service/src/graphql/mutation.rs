//! GraphQL mutation root — **DEPRECATED, pending REST migration**.
//!
//! ADR-DES.API.rest-graphql-mutation-boundary.md устанавливает, что GraphQL
//! в VEDO Core не содержит никаких mutations: все write-операции выполняются
//! через REST `/api/v1/...` под API Gateway (Idempotency-Key, auth middleware,
//! audit log, CircuitBreakerMiddleware DoS protection).
//!
//! Резолверы ниже (`update_draft`, `update_member_role`, `remove_member`)
//! оставлены как **deprecated placeholders** на время миграции фронтенда на
//! REST. После завершения миграции `MutationRoot` будет заменён на
//! `EmptyMutation` в `build_schema` (см. `schema.rs`).

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
#[allow(deprecated)] // async-graphql Object macro references deprecated fields in generated introspection code.
impl MutationRoot {
    /// Marks the workspace as dirty and returns the timestamp of the update.
    ///
    /// **DEPRECATED.** Use `PUT /api/v1/ontologies/{id}/draft` instead.
    /// Mutation root будет удалён (см. ADR-DES.API.rest-graphql-mutation-boundary.md).
    ///
    /// The ontology dirty state is tracked by the frontend (`useVersionContext`)
    /// and persisted in browser storage. The backend only echoes the event so
    /// multiple collaborating clients can converge on a shared clock.
    #[deprecated(
        note = "GraphQL mutations are forbidden; use PUT /api/v1/ontologies/{id}/draft. MutationRoot will be removed."
    )]
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

    // ── Organization model mutations — DEPRECATED, use REST ──────────

    /// Updates a member's role in an ontology scope.
    ///
    /// **DEPRECATED.** Use `PUT /api/v1/ontologies/{id}/members/{userId}` instead.
    #[deprecated(
        note = "GraphQL mutations are forbidden; use PUT /api/v1/ontologies/{id}/members/{userId}. MutationRoot will be removed."
    )]
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
    ///
    /// **DEPRECATED.** Use `DELETE /api/v1/ontologies/{id}/members/{userId}` instead.
    #[deprecated(
        note = "GraphQL mutations are forbidden; use DELETE /api/v1/ontologies/{id}/members/{userId}. MutationRoot will be removed."
    )]
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
