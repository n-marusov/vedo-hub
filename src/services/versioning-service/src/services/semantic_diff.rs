//! Semantic diff computation for commit visualization.
//!
//! Transforms low-level triple deltas into entity-level semantic changes
//! suitable for UI rendering with color-coded diff views (added=green,
//! removed=red, modified=yellow).
//!
//! # Entity Classification
//!
//! Entities are classified from `rdf:type` triples in the delta:
//! - `owl:Class` → Class
//! - `owl:ObjectProperty`, `owl:DatatypeProperty` → Property
//! - `owl:NamedIndividual` or any entity with `rdf:type` pointing to a class → Individual
//!
//! # Diff Structure
//!
//! Each entry contains `{entity_type, entity_id, change_type, before, after}`,
//! where `before`/`after` are maps of predicate → object triples describing
//! the entity's state before and after the commit.

use std::collections::{HashMap, HashSet};

use serde::Serialize;

use crate::models::{CommitDelta, ModifiedTriple, TripleRef};

// ── Diff Types ───────────────────────────────────────────────────────────────

/// The type of ontology entity.
#[derive(Debug, Clone, Serialize, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum EntityType {
    Class,
    Property,
    Individual,
}

/// The type of change at the entity level.
#[derive(Debug, Clone, Serialize, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum ChangeType {
    /// Entity was created in this commit.
    Added,
    /// Entity was deleted in this commit.
    Removed,
    /// Entity's attributes changed (one or more triples modified or mixed
    /// add/remove on the same subject).
    Modified,
}

/// A single semantic diff entry — describes how one entity changed.
#[derive(Debug, Clone, Serialize)]
pub struct SemanticDiffEntry {
    /// The entity type (class, property, individual).
    pub entity_type: EntityType,
    /// The entity's local ID or IRI.
    pub entity_id: String,
    /// The type of change.
    pub change_type: ChangeType,
    /// The entity's triples before the change (empty for Added).
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub before: Vec<TripleRef>,
    /// The entity's triples after the change (empty for Removed).
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub after: Vec<TripleRef>,
}

/// The full semantic diff result for a commit delta.
#[derive(Debug, Clone, Serialize)]
pub struct SemanticDiff {
    /// Entity-level diff entries.
    pub entries: Vec<SemanticDiffEntry>,
    /// Summary counts for convenience.
    pub summary: DiffSummary,
}

/// Summary counts of entity changes.
#[derive(Debug, Clone, Serialize)]
pub struct DiffSummary {
    pub classes_added: usize,
    pub classes_removed: usize,
    pub classes_modified: usize,
    pub properties_added: usize,
    pub properties_removed: usize,
    pub properties_modified: usize,
    pub individuals_added: usize,
    pub individuals_removed: usize,
    pub individuals_modified: usize,
}

// ── Constants ────────────────────────────────────────────────────────────────

const RDF_TYPE: &str = "rdf:type";
const OWL_CLASS: &str = "owl:Class";
const OWL_OBJECT_PROPERTY: &str = "owl:ObjectProperty";
const OWL_DATATYPE_PROPERTY: &str = "owl:DatatypeProperty";
const OWL_FUNCTIONAL_PROPERTY: &str = "owl:FunctionalProperty";
const OWL_INVERSE_FUNCTIONAL_PROPERTY: &str = "owl:InverseFunctionalProperty";
const OWL_TRANSITIVE_PROPERTY: &str = "owl:TransitiveProperty";
const OWL_SYMMETRIC_PROPERTY: &str = "owl:SymmetricProperty";

/// Additional OWL types that indicate property characteristics (sub-types).
const PROPERTY_CHARACTERISTICS: &[&str] = &[
    OWL_FUNCTIONAL_PROPERTY,
    OWL_INVERSE_FUNCTIONAL_PROPERTY,
    OWL_TRANSITIVE_PROPERTY,
    OWL_SYMMETRIC_PROPERTY,
];

// ── Public API ───────────────────────────────────────────────────────────────

/// Computes a semantic diff from a raw commit delta.
///
/// Converts the low-level triple add/remove/modify operations into
/// entity-level semantic changes grouped by subject (entity ID).
///
/// # Arguments
///
/// * `delta` - The raw triple delta from a commit.
///
/// # Returns
///
/// A `SemanticDiff` containing entity-level entries and summary counts.
pub fn compute_semantic_diff(delta: &CommitDelta) -> SemanticDiff {
    tracing::debug!(
        added = delta.added_triples.len(),
        removed = delta.removed_triples.len(),
        modified = delta.modified_triples.len(),
        "Computing semantic diff"
    );

    // Collect all affected entity IDs from added/removed/modified triples
    let mut affected_entities: HashSet<&str> = HashSet::new();
    for t in &delta.added_triples {
        affected_entities.insert(t.s.as_str());
    }
    for t in &delta.removed_triples {
        affected_entities.insert(t.s.as_str());
    }
    for t in &delta.modified_triples {
        affected_entities.insert(t.s.as_str());
    }

    // Build lookups: subject → [triples] for added and removed
    let added_by_subject = group_by_subject(&delta.added_triples);
    let removed_by_subject = group_by_subject(&delta.removed_triples);

    // Build lookups for modified triples: subject → [&ModifiedTriple]
    // Each modified triple contributes a "before" TripleRef (old_o) and
    // an "after" TripleRef (new_o) so modified-only entities have non-empty
    // before/after vectors.
    let mut modified_by_subject: HashMap<&str, Vec<&ModifiedTriple>> = HashMap::new();
    for t in &delta.modified_triples {
        modified_by_subject.entry(t.s.as_str()).or_default().push(t);
    }

    // Build a set of "before" subjects and "after" subjects
    let _after_subjects: HashSet<&str> = added_by_subject.keys().copied().collect();
    let _before_subjects: HashSet<&str> = removed_by_subject.keys().copied().collect();

    let mut entries = Vec::new();

    for entity_id in affected_entities {
        let mut after_triples = added_by_subject.get(entity_id).cloned().unwrap_or_default();
        let mut before_triples = removed_by_subject
            .get(entity_id)
            .cloned()
            .unwrap_or_default();

        // Merge modified triples into before/after
        if let Some(modified) = modified_by_subject.get(entity_id) {
            for m in modified {
                before_triples.push(TripleRef {
                    s: m.s.clone(),
                    p: m.p.clone(),
                    o: m.old_o.clone(),
                });
                after_triples.push(TripleRef {
                    s: m.s.clone(),
                    p: m.p.clone(),
                    o: m.new_o.clone(),
                });
            }
        }

        let has_after = !after_triples.is_empty();
        let has_before = !before_triples.is_empty();

        let is_added = has_after && !has_before;
        let is_removed = has_before && !has_after;

        let change_type = if is_added {
            ChangeType::Added
        } else if is_removed {
            ChangeType::Removed
        } else {
            ChangeType::Modified
        };

        // Determine entity type from rdf:type triples
        let entity_type = classify_entity_type(&after_triples, &before_triples);

        let entry = SemanticDiffEntry {
            entity_type,
            entity_id: entity_id.to_string(),
            change_type,
            before: before_triples,
            after: after_triples,
        };

        entries.push(entry);
    }

    // Sort entries: class → property → individual, then by entity_id
    entries.sort_by(|a, b| {
        let type_ord = |et: &EntityType| -> u8 {
            match et {
                EntityType::Class => 0,
                EntityType::Property => 1,
                EntityType::Individual => 2,
            }
        };
        type_ord(&a.entity_type)
            .cmp(&type_ord(&b.entity_type))
            .then_with(|| a.entity_id.cmp(&b.entity_id))
    });

    let summary = compute_summary(&entries);

    tracing::info!(
        total_entities = entries.len(),
        classes_added = summary.classes_added,
        classes_removed = summary.classes_removed,
        classes_modified = summary.classes_modified,
        properties_added = summary.properties_added,
        properties_removed = summary.properties_removed,
        properties_modified = summary.properties_modified,
        individuals_added = summary.individuals_added,
        individuals_removed = summary.individuals_removed,
        individuals_modified = summary.individuals_modified,
        "Semantic diff computed"
    );

    SemanticDiff { entries, summary }
}

// ── Internal helpers ─────────────────────────────────────────────────────────

/// Groups triples by their subject.
fn group_by_subject(triples: &[TripleRef]) -> HashMap<&str, Vec<TripleRef>> {
    let mut map: HashMap<&str, Vec<TripleRef>> = HashMap::new();
    for t in triples {
        map.entry(t.s.as_str()).or_default().push(t.clone());
    }
    map
}

/// Classifies an entity by examining its rdf:type triples.
fn classify_entity_type(after: &[TripleRef], before: &[TripleRef]) -> EntityType {
    // Check "after" triples first (most current)
    for t in after.iter().chain(before.iter()) {
        if t.p != RDF_TYPE {
            continue;
        }
        match t.o.as_str() {
            OWL_CLASS => return EntityType::Class,
            OWL_OBJECT_PROPERTY | OWL_DATATYPE_PROPERTY => return EntityType::Property,
            _ if PROPERTY_CHARACTERISTICS.contains(&t.o.as_str()) => return EntityType::Property,
            _ => {
                // If the object is a class ID (not a built-in OWL type),
                // it's an Individual with rdf:type pointing to its class
                if !t.o.starts_with("owl:") {
                    return EntityType::Individual;
                }
            }
        }
    }
    // Fallback: if we can't determine from rdf:type, look at naming patterns
    // or default to Individual
    EntityType::Individual
}

/// Computes summary counts from a list of semantic diff entries.
fn compute_summary(entries: &[SemanticDiffEntry]) -> DiffSummary {
    let mut summary = DiffSummary {
        classes_added: 0,
        classes_removed: 0,
        classes_modified: 0,
        properties_added: 0,
        properties_removed: 0,
        properties_modified: 0,
        individuals_added: 0,
        individuals_removed: 0,
        individuals_modified: 0,
    };

    for entry in entries {
        match (&entry.entity_type, &entry.change_type) {
            (EntityType::Class, ChangeType::Added) => summary.classes_added += 1,
            (EntityType::Class, ChangeType::Removed) => summary.classes_removed += 1,
            (EntityType::Class, ChangeType::Modified) => summary.classes_modified += 1,
            (EntityType::Property, ChangeType::Added) => summary.properties_added += 1,
            (EntityType::Property, ChangeType::Removed) => summary.properties_removed += 1,
            (EntityType::Property, ChangeType::Modified) => summary.properties_modified += 1,
            (EntityType::Individual, ChangeType::Added) => summary.individuals_added += 1,
            (EntityType::Individual, ChangeType::Removed) => summary.individuals_removed += 1,
            (EntityType::Individual, ChangeType::Modified) => summary.individuals_modified += 1,
        }
    }

    summary
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

    #[test]
    fn test_class_added() {
        let delta = CommitDelta {
            added_triples: vec![
                make_triple("Person", "rdf:type", "owl:Class"),
                make_triple("Person", "rdfs:label", "\"Person\"^^xsd:string"),
            ],
            removed_triples: vec![],
            modified_triples: vec![],
            merge_metadata: None,
        };

        let diff = compute_semantic_diff(&delta);
        assert_eq!(diff.entries.len(), 1);
        assert_eq!(diff.entries[0].entity_type, EntityType::Class);
        assert_eq!(diff.entries[0].entity_id, "Person");
        assert_eq!(diff.entries[0].change_type, ChangeType::Added);
        assert!(diff.entries[0].before.is_empty());
        assert_eq!(diff.entries[0].after.len(), 2);
        assert_eq!(diff.summary.classes_added, 1);
    }

    #[test]
    fn test_property_removed() {
        let delta = CommitDelta {
            added_triples: vec![],
            removed_triples: vec![
                make_triple("hasName", "rdf:type", "owl:DatatypeProperty"),
                make_triple("hasName", "rdfs:label", "\"hasName\"^^xsd:string"),
            ],
            modified_triples: vec![],
            merge_metadata: None,
        };

        let diff = compute_semantic_diff(&delta);
        assert_eq!(diff.entries.len(), 1);
        assert_eq!(diff.entries[0].entity_type, EntityType::Property);
        assert_eq!(diff.entries[0].change_type, ChangeType::Removed);
        assert!(diff.entries[0].after.is_empty());
        assert_eq!(diff.summary.properties_removed, 1);
    }

    #[test]
    fn test_individual_modified() {
        let delta = CommitDelta {
            added_triples: vec![make_triple(
                "alice",
                "rdfs:label",
                "\"Alice Smith\"^^xsd:string",
            )],
            removed_triples: vec![make_triple("alice", "rdfs:label", "\"Alice\"^^xsd:string")],
            modified_triples: vec![],
            merge_metadata: None,
        };

        let diff = compute_semantic_diff(&delta);
        assert_eq!(diff.entries.len(), 1);
        assert_eq!(diff.entries[0].entity_type, EntityType::Individual);
        assert_eq!(diff.entries[0].change_type, ChangeType::Modified);
        assert_eq!(diff.summary.individuals_modified, 1);
    }

    #[test]
    fn test_multiple_entities() {
        let delta = CommitDelta {
            added_triples: vec![
                make_triple("Person", "rdf:type", "owl:Class"),
                make_triple("hasAge", "rdf:type", "owl:DatatypeProperty"),
            ],
            removed_triples: vec![make_triple("OldClass", "rdf:type", "owl:Class")],
            modified_triples: vec![],
            merge_metadata: None,
        };

        let diff = compute_semantic_diff(&delta);
        assert_eq!(diff.entries.len(), 3);
        // Sorted: classes first
        assert_eq!(diff.entries[0].entity_type, EntityType::Class);
        assert_eq!(diff.entries[1].entity_type, EntityType::Class);
        assert_eq!(diff.entries[2].entity_type, EntityType::Property);
        assert_eq!(diff.summary.classes_added, 1);
        assert_eq!(diff.summary.classes_removed, 1);
        assert_eq!(diff.summary.properties_added, 1);
    }

    #[test]
    fn test_empty_delta() {
        let delta = CommitDelta::default();
        let diff = compute_semantic_diff(&delta);
        assert!(diff.entries.is_empty());
        assert_eq!(diff.summary.classes_added, 0);
    }

    #[test]
    fn test_modified_triples_combined_with_add_remove() {
        let delta = CommitDelta {
            added_triples: vec![make_triple(
                "Person",
                "rdfs:comment",
                "\"Updated\"^^xsd:string",
            )],
            removed_triples: vec![make_triple("Person", "rdfs:comment", "\"Old\"^^xsd:string")],
            modified_triples: vec![],
            merge_metadata: None,
        };

        let diff = compute_semantic_diff(&delta);
        assert_eq!(diff.entries.len(), 1);
        assert_eq!(diff.entries[0].change_type, ChangeType::Modified);
    }

    #[test]
    fn test_property_with_characteristics() {
        let delta = CommitDelta {
            added_triples: vec![
                make_triple("isParentOf", "rdf:type", "owl:ObjectProperty"),
                make_triple("isParentOf", "rdf:type", "owl:TransitiveProperty"),
            ],
            removed_triples: vec![],
            modified_triples: vec![],
            merge_metadata: None,
        };

        let diff = compute_semantic_diff(&delta);
        assert_eq!(diff.entries.len(), 1);
        assert_eq!(diff.entries[0].entity_type, EntityType::Property);
        assert_eq!(diff.entries[0].change_type, ChangeType::Added);
    }
}
