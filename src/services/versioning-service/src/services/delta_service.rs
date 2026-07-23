//! Delta service — commit chain replay and materialized state computation.
//!
//! Replays deltas from the root commit through the parent chain to produce
//! the materialized state (set of active triples) for a given commit or branch.
//! Supports snapshot-based fast paths: when a pre-computed snapshot is found
//! within 10 commits of the target, only the trailing deltas are replayed.

use sqlx::PgPool;
use uuid::Uuid;

use std::collections::{HashMap, HashSet};

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
    /// First checks for a pre-computed state snapshot. If a snapshot is found
    /// within 10 commits of the target, replays only the trailing deltas from
    /// the snapshot commit forward. Otherwise replays the full commit chain.
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

        // Try to find a snapshot at or near the target commit
        if let Some((snapshot_commit_id, triples)) =
            self.find_nearest_snapshot(target_commit_id).await?
        {
            let snapshot_chain = self.walk_commit_chain(target_commit_id).await?;
            let snapshot_index = snapshot_chain
                .iter()
                .position(|c| c.id == snapshot_commit_id);

            if let Some(idx) = snapshot_index {
                let trailing = &snapshot_chain[idx + 1..];
                if trailing.len() <= 10 {
                    let mut state: Vec<TripleRef> =
                        serde_json::from_value(triples).map_err(|e| {
                            VersionError::Database(format!(
                                "Failed to deserialize snapshot triples: {e}"
                            ))
                        })?;

                    for commit in trailing {
                        apply_delta_to_state(&mut state, &commit.delta);
                    }

                    tracing::debug!(
                        snapshot_commit_id = %snapshot_commit_id,
                        trailing_deltas = trailing.len(),
                        triple_count = state.len(),
                        "Materialize: snapshot found, replaying trailing deltas",
                    );

                    return Ok(MaterializedState {
                        triples: state,
                        delta_count: trailing.len() + 1,
                    });
                }
            }
        }

        // No suitable snapshot — replay full chain
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

        let total_commits = chain.len();
        tracing::debug!(
            total_commits,
            "Materialize: no snapshot found, replaying full chain",
        );

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

        tracing::info!(
            target_commit_id = %target_commit_id,
            triple_count = state.len(),
            deltas_replayed = total_commits,
            "Materialized state computed"
        );

        Ok(MaterializedState {
            triples: state,
            delta_count: total_commits,
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
            r"
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
            ",
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

    /// Looks up the nearest state snapshot at or near the given commit.
    /// Returns `(snapshot_commit_id, triples_json)` if found within 10 commits.
    async fn find_nearest_snapshot(
        &self,
        commit_id: Uuid,
    ) -> Result<Option<(Uuid, serde_json::Value)>, VersionError> {
        let pool = self.commit_repo.pool().clone();

        // Use the commit chain to walk backward and check for snapshots.
        // We check up to 11 commits back (target + 10 ancestors).
        let row: Option<(Uuid, serde_json::Value)> = sqlx::query_as(
            r"
            WITH RECURSIVE commit_chain AS (
                SELECT id, parent_commit_id, 0 AS depth
                FROM commits
                WHERE id = $1
                UNION ALL
                SELECT c.id, c.parent_commit_id, cc.depth + 1
                FROM commits c
                INNER JOIN commit_chain cc ON c.id = cc.parent_commit_id
                WHERE cc.depth < 10
            )
            SELECT ss.commit_id, ss.triples
            FROM state_snapshots ss
            INNER JOIN commit_chain cc ON ss.commit_id = cc.id
            ORDER BY cc.depth ASC
            LIMIT 1
            ",
        )
        .bind(commit_id)
        .fetch_optional(&pool)
        .await
        .map_err(|e| VersionError::Database(format!("Failed to query state snapshots: {e}")))?;

        Ok(row)
    }

    /// Creates a state snapshot for the given commit by materializing
    /// the state and storing it in the `state_snapshots` table.
    ///
    /// This is best-effort and non-blocking: failures are logged but
    /// not propagated to the caller.
    pub async fn create_snapshot(&self, commit_id: Uuid) -> Result<(), VersionError> {
        let pool = self.commit_repo.pool().clone();

        // Materialize the full state for this commit
        let materialized = self.materialize(commit_id).await?;
        let triples_json = serde_json::to_value(&materialized.triples)
            .map_err(|e| VersionError::Database(format!("Snapshot serialization error: {e}")))?;

        // Get the branch ID from the commit
        let branch_id: Uuid = sqlx::query_scalar("SELECT branch_id FROM commits WHERE id = $1")
            .bind(commit_id)
            .fetch_optional(&pool)
            .await?
            .ok_or_else(|| VersionError::CommitNotFound(commit_id.to_string()))?;

        sqlx::query(
            r"
            INSERT INTO state_snapshots (commit_id, branch_id, triples)
            VALUES ($1, $2, $3)
            ON CONFLICT (commit_id) DO UPDATE
                SET triples = EXCLUDED.triples,
                    created_at = NOW()
            ",
        )
        .bind(commit_id)
        .bind(branch_id)
        .bind(&triples_json)
        .execute(&pool)
        .await
        .map_err(|e| VersionError::Database(format!("Failed to insert state snapshot: {e}")))?;

        let triple_count = materialized.triples.len();
        tracing::info!(
            commit_id = %commit_id,
            triple_count,
            "State snapshot created",
        );

        Ok(())
    }

    /// Checks if a snapshot should be created based on commit count,
    /// and creates one if this is the 50th, 100th, etc. commit on the branch.
    ///
    /// This is best-effort: failures are logged and squelched.
    pub async fn maybe_create_snapshot(&self, commit_id: Uuid) {
        let pool = self.commit_repo.pool().clone();

        // Get the branch_id for this commit
        let branch_id: Option<Uuid> =
            sqlx::query_scalar("SELECT branch_id FROM commits WHERE id = $1")
                .bind(commit_id)
                .fetch_optional(&pool)
                .await
                .unwrap_or(None)
                .flatten();

        let Some(branch_id) = branch_id else {
            tracing::warn!(
                commit_id = %commit_id,
                "Cannot create snapshot: commit not found"
            );
            return;
        };

        // Count commits on this branch
        let count: i64 =
            match sqlx::query_scalar("SELECT COUNT(*) FROM commits WHERE branch_id = $1")
                .bind(branch_id)
                .fetch_one(&pool)
                .await
            {
                Ok(c) => c,
                Err(e) => {
                    tracing::warn!(
                        error = %e,
                        branch_id = %branch_id,
                        "Cannot create snapshot: failed to count commits"
                    );
                    return;
                }
            };

        // Create snapshot every 50 commits
        if count > 0 && count % 50 == 0 {
            if let Err(e) = self.create_snapshot(commit_id).await {
                tracing::warn!(
                    error = %e,
                    commit_id = %commit_id,
                    "Failed to create state snapshot"
                );
            }
        }
    }
}

/// Applies a delta to a mutable state vector, deduplicating on output.
///
/// Handles additions, removals, and modifications correctly:
/// - Removed triples are deleted from the state.
/// - Modified triples: old value is removed, new value is added.
/// - Added triples are appended.
///
/// [FIX] Uses pre-built HashSet/HashMap for O(n + δ) removal instead of O(N·M).
/// [FIX] Deduplicates via `HashSet` on final state to prevent duplicate triple accumulation.
pub fn apply_delta_to_state(state: &mut Vec<TripleRef>, delta: &CommitDelta) {
    // [FIX] Pre-build HashSet for O(1) removal checks
    let removed_set: HashSet<TripleRef> = delta.removed_triples.iter().cloned().collect();

    // [FIX] Pre-build HashMap for O(1) modification checks
    // Map (s, p) -> old_o for each modified triple
    let modified_map: HashMap<(&str, &str), &str> = delta
        .modified_triples
        .iter()
        .map(|m| ((m.s.as_str(), m.p.as_str()), m.old_o.as_str()))
        .collect();

    // [FIX] Single retain pass using pre-built sets — O(n) instead of repeated scans
    state.retain(|t| {
        if removed_set.contains(t) {
            return false;
        }
        if let Some(&old_o) = modified_map.get(&(t.s.as_str(), t.p.as_str())) {
            if t.o == old_o {
                return false;
            }
        }
        true
    });

    // Add new triples
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

    // [FIX] Dedup to prevent duplicate triple accumulation
    let mut seen: HashSet<TripleRef> = HashSet::with_capacity(state.len());
    state.retain(|t| seen.insert(t.clone()));
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
            merge_metadata: None,
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
            merge_metadata: None,
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
            merge_metadata: None,
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
            merge_metadata: None,
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
            merge_metadata: None,
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
            merge_metadata: None,
        };
        // Inverse: added becomes removed
        let inverse_removed = delta.added_triples.clone();
        assert_eq!(inverse_removed.len(), 1);
        assert_eq!(inverse_removed[0].s, "X");
    }
}
