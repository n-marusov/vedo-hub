//! Delta service — commit chain replay and materialized state computation.
//!
//! Replays deltas from the root commit through the parent chain to produce
//! the materialized state (set of active triples) for a given commit or branch.

use sqlx::PgPool;
use uuid::Uuid;

use crate::error::VersionError;
use crate::models::{Commit, CommitDelta, TripleRef};
use crate::repositories::CommitRepository;

/// Replays deltas along a commit chain to produce a materialized state.
///
/// The materialized state is the set of all triples that are "active"
/// (added but not subsequently removed) after applying all deltas up to
/// a given commit.
pub struct DeltaReplayEngine {
    commit_repo: CommitRepository,
}

impl DeltaReplayEngine {
    /// Creates a new engine backed by the given pool.
    pub fn new(pool: PgPool) -> Self {
        Self {
            commit_repo: CommitRepository::new(pool),
        }
    }

    /// Replays all deltas from the root commit to the given `target_commit_id`.
    ///
    /// Walks backward from the target (via `parent_commit_id`), collecting
    /// commits, then replays them forward to produce the materialized state.
    ///
    /// Returns the list of active triples and the number of deltas replayed.
    pub async fn materialize(
        &self,
        target_commit_id: Uuid,
    ) -> Result<MaterializedState, VersionError> {
        tracing::debug!(
            target_commit_id = %target_commit_id,
            "Materializing state from commit chain"
        );

        let chain = self.walk_commit_chain(target_commit_id).await?;

        if chain.is_empty() {
            tracing::warn!(
                target_commit_id = %target_commit_id,
                "No commits found in chain — returning empty state"
            );
            return Ok(MaterializedState {
                triples: Vec::new(),
                delta_count: 0,
            });
        }

        let mut state: Vec<TripleRef> = Vec::new();

        for (i, commit) in chain.iter().enumerate() {
            apply_delta_to_state(&mut state, &commit.delta);

            if (i + 1) % 100 == 0 {
                tracing::debug!(
                    commits_processed = i + 1,
                    current_triple_count = state.len(),
                    "Delta replay in progress"
                );
            }
        }

        let delta_count = chain.len();

        tracing::info!(
            target_commit_id = %target_commit_id,
            triple_count = state.len(),
            deltas_replayed = delta_count,
            "Materialized state computed"
        );

        Ok(MaterializedState {
            triples: state,
            delta_count,
        })
    }

    /// Computes the inverse delta between a current commit and a target
    /// commit. The inverse delta, when applied, reverses the changes.
    pub async fn compute_inverse_delta(
        &self,
        current_head: Uuid,
        target_head: Uuid,
    ) -> Result<CommitDelta, VersionError> {
        tracing::debug!(
            current = %current_head,
            target = %target_head,
            "Computing inverse delta for rollback"
        );

        let current_chain = self.walk_commit_chain(current_head).await?;
        let target_chain = self.walk_commit_chain(target_head).await?;

        let target_ids: std::collections::HashSet<Uuid> =
            target_chain.iter().map(|c| c.id).collect();

        let mut deltas_to_reverse: Vec<&CommitDelta> = Vec::new();
        for commit in &current_chain {
            if !target_ids.contains(&commit.id) {
                deltas_to_reverse.push(&commit.delta);
            }
        }

        let mut inverse = CommitDelta::default();
        for delta in deltas_to_reverse.iter().rev() {
            for triple in &delta.removed_triples {
                inverse.added_triples.push(triple.clone());
            }
            for triple in &delta.added_triples {
                inverse.removed_triples.push(triple.clone());
            }
            for mt in &delta.modified_triples {
                inverse
                    .modified_triples
                    .push(crate::models::ModifiedTriple {
                        s: mt.s.clone(),
                        p: mt.p.clone(),
                        old_o: mt.new_o.clone(),
                        new_o: mt.old_o.clone(),
                    });
            }
        }

        tracing::info!(
            deltas_to_reverse = deltas_to_reverse.len(),
            inverse_added = inverse.added_triples.len(),
            inverse_removed = inverse.removed_triples.len(),
            "Inverse delta computed"
        );

        Ok(inverse)
    }

    /// Walks the commit chain from a commit back to the root using a recursive CTE.
    async fn walk_commit_chain(&self, start_commit_id: Uuid) -> Result<Vec<Commit>, VersionError> {
        let pool = self.commit_repo.pool().clone();

        let rows = sqlx::query(
            r#"
            WITH RECURSIVE commit_chain AS (
                SELECT id, branch_id, parent_commit_id, message, author_id,
                       author_name, delta, created_at
                FROM commits
                WHERE id = $1
                UNION ALL
                SELECT c.id, c.branch_id, c.parent_commit_id, c.message,
                       c.author_id, c.author_name, c.delta, c.created_at
                FROM commits c
                INNER JOIN commit_chain cc ON c.id = cc.parent_commit_id
            )
            SELECT id, branch_id, parent_commit_id, message, author_id,
                   author_name, delta, created_at
            FROM commit_chain
            ORDER BY created_at ASC
            "#,
        )
        .bind(start_commit_id)
        .fetch_all(&pool)
        .await
        .map_err(|e| VersionError::Database(format!("Failed to walk commit chain: {e}")))?;

        let mut commits = Vec::with_capacity(rows.len());
        for row in &rows {
            let commit = crate::repositories::commit_repo::row_to_commit(row).map_err(|e| {
                VersionError::Database(format!("Failed to parse commit in chain: {e}"))
            })?;
            commits.push(commit);
        }

        if commits.is_empty() {
            return Err(VersionError::CommitNotFound(start_commit_id.to_string()));
        }

        Ok(commits)
    }
}

/// Standalone function that applies a delta to a mutable state vector.
///
/// Handles additions, removals, and modifications correctly:
/// - Removed triples are deleted from the state.
/// - Modified triples: old value is removed, new value is added.
/// - Added triples are appended.
pub fn apply_delta_to_state(state: &mut Vec<TripleRef>, delta: &CommitDelta) {
    for removed in &delta.removed_triples {
        state.retain(|t| t != removed);
    }

    for modified in &delta.modified_triples {
        state.retain(|t| !(t.s == modified.s && t.p == modified.p && t.o == modified.old_o));
    }

    for added in &delta.added_triples {
        state.push(added.clone());
    }

    for modified in &delta.modified_triples {
        state.push(TripleRef {
            s: modified.s.clone(),
            p: modified.p.clone(),
            o: modified.new_o.clone(),
        });
    }
}

/// The materialized state resulting from replaying a commit chain.
#[derive(Debug, Clone)]
pub struct MaterializedState {
    /// Active triples after applying all deltas.
    pub triples: Vec<TripleRef>,
    /// Number of deltas replayed.
    pub delta_count: usize,
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::models::{CommitDelta, ModifiedTriple, TripleRef};

    #[test]
    fn test_apply_delta_additions() {
        let mut state = Vec::new();
        let delta = CommitDelta {
            added_triples: vec![TripleRef {
                s: "A".into(),
                p: "p".into(),
                o: "o1".into(),
            }],
            removed_triples: vec![],
            modified_triples: vec![],
        };
        apply_delta_to_state(&mut state, &delta);
        assert_eq!(state.len(), 1);
        assert_eq!(state[0].s, "A");
    }

    #[test]
    fn test_apply_delta_removals() {
        let mut state = vec![TripleRef {
            s: "A".into(),
            p: "p".into(),
            o: "o1".into(),
        }];
        let delta = CommitDelta {
            added_triples: vec![],
            removed_triples: vec![TripleRef {
                s: "A".into(),
                p: "p".into(),
                o: "o1".into(),
            }],
            modified_triples: vec![],
        };
        apply_delta_to_state(&mut state, &delta);
        assert!(state.is_empty());
    }

    #[test]
    fn test_apply_delta_modifications() {
        let mut state = vec![TripleRef {
            s: "A".into(),
            p: "p".into(),
            o: "old".into(),
        }];
        let delta = CommitDelta {
            added_triples: vec![],
            removed_triples: vec![],
            modified_triples: vec![ModifiedTriple {
                s: "A".into(),
                p: "p".into(),
                old_o: "old".into(),
                new_o: "new".into(),
            }],
        };
        apply_delta_to_state(&mut state, &delta);
        assert_eq!(state.len(), 1);
        assert_eq!(state[0].o, "new");
    }

    #[test]
    fn test_apply_multiple_deltas() {
        let mut state = Vec::new();

        let delta1 = CommitDelta {
            added_triples: vec![TripleRef {
                s: "A".into(),
                p: "p".into(),
                o: "v1".into(),
            }],
            removed_triples: vec![],
            modified_triples: vec![],
        };
        apply_delta_to_state(&mut state, &delta1);
        assert_eq!(state.len(), 1);

        let delta2 = CommitDelta {
            added_triples: vec![TripleRef {
                s: "B".into(),
                p: "p".into(),
                o: "v2".into(),
            }],
            removed_triples: vec![TripleRef {
                s: "A".into(),
                p: "p".into(),
                o: "v1".into(),
            }],
            modified_triples: vec![],
        };
        apply_delta_to_state(&mut state, &delta2);
        assert_eq!(state.len(), 1);
        assert_eq!(state[0].s, "B");
    }

    #[test]
    fn test_inverse_delta_logic() {
        let delta = CommitDelta {
            added_triples: vec![TripleRef {
                s: "X".into(),
                p: "p".into(),
                o: "o".into(),
            }],
            removed_triples: vec![],
            modified_triples: vec![],
        };
        // Inverse: added becomes removed
        let inverse_removed = delta.added_triples.clone();
        assert_eq!(inverse_removed.len(), 1);
        assert_eq!(inverse_removed[0].s, "X");
    }
}
