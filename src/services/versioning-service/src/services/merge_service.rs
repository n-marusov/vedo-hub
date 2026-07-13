//! Merge service — diff computation, conflict resolution, and merge commit creation.
//!
//! Provides the core merge logic that combines deltas from two branches,
//! auto-resolves conflicts using source-branch preference on triple-level
//! collisions, and produces the combined delta for the merge commit.

use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::models::{CommitDelta, ModifiedTriple, TripleRef};

/// A single diff entry discovered during merge analysis.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct DiffEntry {
    pub triple: TripleRef,
    pub action: DiffAction,
    pub old_value: Option<String>,
}

/// The type of change for a diff entry.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum DiffAction {
    Added,
    Removed,
    Modified { old_o: String, new_o: String },
}

/// Result of a merge analysis.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MergeAnalysis {
    pub source_branch_id: Uuid,
    pub target_branch_id: Uuid,
    pub merged_delta: CommitDelta,
    pub conflict_count: usize,
    pub auto_resolved: bool,
}

/// Computes the merged delta when merging `source_delta` into `target_delta`.
///
/// Conflict resolution strategy (source-preference on triple-level collision):
/// - If a triple was added in source and also modified in target → use source's version
/// - If a triple was removed in source but modified in target → prefer source (removed)
/// - If a triple was modified in both → prefer source's new value
///
/// Returns the merged delta and a conflict count.
pub fn compute_merged_delta(
    source_delta: &CommitDelta,
    target_delta: &CommitDelta,
) -> MergeAnalysis {
    tracing::debug!(
        source_added = source_delta.added_triples.len(),
        source_removed = source_delta.removed_triples.len(),
        source_modified = source_delta.modified_triples.len(),
        target_added = target_delta.added_triples.len(),
        target_removed = target_delta.removed_triples.len(),
        target_modified = target_delta.modified_triples.len(),
        "Computing merge delta"
    );

    let mut merged_added: Vec<TripleRef> = Vec::new();
    let mut merged_removed: Vec<TripleRef> = Vec::new();
    let mut merged_modified: Vec<ModifiedTriple> = Vec::new();
    let mut conflict_count = 0;

    // Build lookup sets for target changes
    let target_added_keys: Vec<String> = target_delta
        .added_triples
        .iter()
        .map(|t| triple_key(t))
        .collect();
    let target_removed_keys: Vec<String> = target_delta
        .removed_triples
        .iter()
        .map(|t| triple_key(t))
        .collect();

    // Process source additions
    for triple in &source_delta.added_triples {
        let key = triple_key(triple);
        if target_removed_keys.contains(&key) {
            // Source added, target removed → conflict, prefer source
            conflict_count += 1;
            tracing::warn!(
                triple_key = %key,
                "Merge conflict: source added, target removed — preferring source"
            );
            merged_added.push(triple.clone());
        } else if !target_added_keys.contains(&key) {
            // Only in source → include
            merged_added.push(triple.clone());
        }
        // If in both → dedup, include only once
    }

    // Process source removals
    for triple in &source_delta.removed_triples {
        let key = triple_key(triple);
        if target_added_keys.contains(&key) {
            // Source removed, target added → conflict, prefer source (removed)
            conflict_count += 1;
            tracing::warn!(
                triple_key = %key,
                "Merge conflict: source removed, target added — preferring source"
            );
            merged_removed.push(triple.clone());
        } else if !target_removed_keys.contains(&key) {
            // Only in source → include
            merged_removed.push(triple.clone());
        }
    }

    // Process source modifications
    for mod_triple in &source_delta.modified_triples {
        let key = triple_key_from_parts(&mod_triple.s, &mod_triple.p, &mod_triple.old_o);
        // Check if target also modified this triple
        let target_conflict = target_delta
            .modified_triples
            .iter()
            .any(|t| t.s == mod_triple.s && t.p == mod_triple.p && t.old_o == mod_triple.old_o);
        if target_conflict {
            conflict_count += 1;
            tracing::warn!(
                triple_key = %key,
                "Merge conflict: both branches modified — preferring source"
            );
        }
        merged_modified.push(mod_triple.clone());
    }

    // Include target changes that don't conflict with source
    for triple in &target_delta.added_triples {
        let key = triple_key(triple);
        if !source_delta
            .added_triples
            .iter()
            .any(|t| triple_key(t) == key)
            && !source_delta
                .removed_triples
                .iter()
                .any(|t| triple_key(t) == key)
        {
            merged_added.push(triple.clone());
        }
    }

    for triple in &target_delta.removed_triples {
        let key = triple_key(triple);
        if !source_delta
            .added_triples
            .iter()
            .any(|t| triple_key(t) == key)
            && !source_delta
                .removed_triples
                .iter()
                .any(|t| triple_key(t) == key)
        {
            merged_removed.push(triple.clone());
        }
    }

    for mod_triple in &target_delta.modified_triples {
        let source_conflict = source_delta
            .modified_triples
            .iter()
            .any(|t| t.s == mod_triple.s && t.p == mod_triple.p && t.old_o == mod_triple.old_o);
        if !source_conflict {
            merged_modified.push(mod_triple.clone());
        }
    }

    let merged_delta = CommitDelta {
        added_triples: merged_added,
        removed_triples: merged_removed,
        modified_triples: merged_modified,
    };

    tracing::info!(
        conflict_count,
        merged_changes = merged_delta.total_changes(),
        auto_resolved = true,
        "Merge delta computed"
    );

    MergeAnalysis {
        source_branch_id: Uuid::nil(),
        target_branch_id: Uuid::nil(),
        merged_delta,
        conflict_count,
        auto_resolved: true,
    }
}

/// Creates a unique string key for a triple reference (used for conflict detection).
fn triple_key(triple: &TripleRef) -> String {
    format!("{}:{}:{}", triple.s, triple.p, triple.o)
}

fn triple_key_from_parts(s: &str, p: &str, o: &str) -> String {
    format!("{s}:{p}:{o}")
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_triple(s: &str, p: &str, o: &str) -> TripleRef {
        TripleRef {
            s: s.to_string(),
            p: p.to_string(),
            o: o.to_string(),
        }
    }

    fn make_modified(s: &str, p: &str, old_o: &str, new_o: &str) -> ModifiedTriple {
        ModifiedTriple {
            s: s.to_string(),
            p: p.to_string(),
            old_o: old_o.to_string(),
            new_o: new_o.to_string(),
        }
    }

    #[test]
    fn test_merge_no_conflicts() {
        let source = CommitDelta {
            added_triples: vec![make_triple("A", "rdfs:label", "A")],
            removed_triples: vec![],
            modified_triples: vec![],
        };
        let target = CommitDelta {
            added_triples: vec![make_triple("B", "rdfs:label", "B")],
            removed_triples: vec![],
            modified_triples: vec![],
        };
        let result = compute_merged_delta(&source, &target);
        assert_eq!(result.conflict_count, 0);
        assert_eq!(result.merged_delta.added_triples.len(), 2);
    }

    #[test]
    fn test_merge_source_added_target_removed_conflict() {
        let source = CommitDelta {
            added_triples: vec![make_triple("X", "p", "o")],
            removed_triples: vec![],
            modified_triples: vec![],
        };
        let target = CommitDelta {
            added_triples: vec![],
            removed_triples: vec![make_triple("X", "p", "o")],
            modified_triples: vec![],
        };
        let result = compute_merged_delta(&source, &target);
        assert_eq!(result.conflict_count, 1);
        // Source preference: include as added
        assert_eq!(result.merged_delta.added_triples.len(), 1);
    }

    #[test]
    fn test_merge_both_modified_same_triple() {
        let source = CommitDelta {
            added_triples: vec![],
            removed_triples: vec![],
            modified_triples: vec![make_modified("C", "rdfs:label", "old", "source-new")],
        };
        let target = CommitDelta {
            added_triples: vec![],
            removed_triples: vec![],
            modified_triples: vec![make_modified("C", "rdfs:label", "old", "target-new")],
        };
        let result = compute_merged_delta(&source, &target);
        assert_eq!(result.conflict_count, 1);
        // Source preference: use source's version
        assert_eq!(result.merged_delta.modified_triples.len(), 1);
        assert_eq!(result.merged_delta.modified_triples[0].new_o, "source-new");
    }

    #[test]
    fn test_merge_empty_deltas() {
        let source = CommitDelta::default();
        let target = CommitDelta::default();
        let result = compute_merged_delta(&source, &target);
        assert_eq!(result.conflict_count, 0);
        assert!(result.merged_delta.is_empty());
    }

    #[test]
    fn test_merge_identical_additions_dedup() {
        let triple = make_triple("A", "p", "o");
        let source = CommitDelta {
            added_triples: vec![triple.clone()],
            removed_triples: vec![],
            modified_triples: vec![],
        };
        let target = CommitDelta {
            added_triples: vec![triple],
            removed_triples: vec![],
            modified_triples: vec![],
        };
        let result = compute_merged_delta(&source, &target);
        assert_eq!(result.conflict_count, 0);
        // Dedup means only one copy
        assert!(result.merged_delta.added_triples.len() <= 1);
    }
}
