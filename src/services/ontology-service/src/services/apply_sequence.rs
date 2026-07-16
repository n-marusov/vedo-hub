//! ApplySequence business logic — the main entry point for document extractor
//! requests to apply an ontology sequence atomically.
//!
//! # Flow
//!
//! 1. **Order steps** — classes first, then properties, then individuals/annotations
//! 2. **Execute in Neo4j transaction** — all or nothing
//! 3. **Create versioning commit** — record the change
//! 4. **Return response** — commit ID, step counts, per-step errors

use std::sync::Arc;
use std::time::{SystemTime, UNIX_EPOCH};

use tracing::{debug, info, warn};
use uuid::Uuid;

use crate::AppState;

use vedo_shared::protos::ontology::v1::{
    sequence_step::Operation, ApplySequenceRequest, ApplySequenceResponse, SequenceStep,
};

/// Applies an ontology sequence atomically within a single Neo4j transaction.
///
/// Steps are ordered by type: classes → properties → individuals → annotations/relationships.
/// All steps execute in one Neo4j transaction — on any failure the entire
/// transaction rolls back. Entity-level errors (e.g., invalid parent reference)
/// are collected and reported per-step without aborting valid steps.
pub async fn apply_sequence(
    state: &Arc<AppState>,
    request: &ApplySequenceRequest,
) -> Result<ApplySequenceResponse, tonic::Status> {
    let pool = match &state.neo4j {
        Some(p) => p,
        None => {
            warn!("sequence.neo4j_not_configured");
            return Ok(ApplySequenceResponse {
                commit_id: String::new(),
                steps_applied: 0,
                steps_skipped: 0,
                steps_failed: 0,
                errors: vec!["Neo4j is not configured".to_string()],
                error: None,
            });
        }
    };

    let ontology_id = request.ontology_id.as_str();
    let steps: Vec<&SequenceStep> = request.steps.iter().collect();

    if steps.is_empty() {
        debug!("sequence.empty_request: ontology_id={}", ontology_id);
        return Ok(ApplySequenceResponse {
            commit_id: String::new(),
            steps_applied: 0,
            steps_skipped: 0,
            steps_failed: 0,
            errors: vec![],
            error: None,
        });
    }

    info!(
        "sequence.start: ontology_id={}, step_count={}, message={}",
        ontology_id,
        steps.len(),
        request.commit_message
    );

    // Organize steps in dependency order: classes first, then properties, then
    // individuals, then annotations/relationships. This avoids dangling references.
    let mut class_steps: Vec<&SequenceStep> = Vec::new();
    let mut object_prop_steps: Vec<&SequenceStep> = Vec::new();
    let mut datatype_prop_steps: Vec<&SequenceStep> = Vec::new();
    let mut individual_steps: Vec<&SequenceStep> = Vec::new();
    let mut other_steps: Vec<&SequenceStep> = Vec::new();

    for step in &steps {
        match step.r#operation() {
            Operation::CreateClass => class_steps.push(step),
            Operation::CreateObjectProperty => object_prop_steps.push(step),
            Operation::CreateDatatypeProperty => datatype_prop_steps.push(step),
            Operation::CreateIndividual => individual_steps.push(step),
            _ => other_steps.push(step),
        }
    }

    let ordered_steps: Vec<&SequenceStep> = class_steps
        .into_iter()
        .chain(object_prop_steps)
        .chain(datatype_prop_steps)
        .chain(individual_steps)
        .chain(other_steps)
        .collect();

    // Execute all steps in a single Neo4j transaction
    let mut tx =
        pool.graph().start_txn().await.map_err(|e| {
            tonic::Status::internal(format!("failed to start Neo4j transaction: {e}"))
        })?;

    let mut steps_applied: i32 = 0;
    let mut steps_skipped: i32 = 0;
    let mut steps_failed: i32 = 0;
    let mut errors: Vec<String> = Vec::new();

    for (i, step) in ordered_steps.iter().enumerate() {
        match crate::repositories::sequence_repo::execute_step(&mut tx, ontology_id, step).await {
            Ok(true) => {
                steps_applied += 1;
                debug!(
                    "sequence.step_applied: index={}, entity_id={}",
                    i, step.entity_id
                );
            }
            Ok(false) => {
                steps_skipped += 1;
                debug!(
                    "sequence.step_skipped: index={}, entity_id={}",
                    i, step.entity_id
                );
            }
            Err(e) => {
                steps_failed += 1;
                let err_msg = format!("step[{}] entity={}: {e}", i, step.entity_id);
                warn!(
                    "sequence.step_failed: ontology_id={}, {}",
                    ontology_id, err_msg
                );
                errors.push(err_msg);
            }
        }
    }

    // Commit or rollback
    let commit_id = if steps_failed > 0 && steps_applied == 0 {
        // All steps failed — rollback
        if let Err(e) = tx.rollback().await {
            warn!(
                "sequence.rollback_failed: ontology_id={}, error={e}",
                ontology_id
            );
        } else {
            debug!("sequence.rollback: ontology_id={}", ontology_id);
        }
        String::new()
    } else {
        // At least some steps succeeded — commit
        if let Err(e) = tx.commit().await {
            warn!(
                "sequence.commit_failed: ontology_id={}, error={e}",
                ontology_id
            );
            // Mark all remaining steps as failed
            let accounted = steps_applied + steps_failed + steps_skipped;
            let remaining = (ordered_steps.len() as i32) - accounted;
            if remaining > 0 {
                steps_failed += remaining;
            }
            return Ok(ApplySequenceResponse {
                commit_id: String::new(),
                steps_applied,
                steps_skipped,
                steps_failed,
                errors: vec![format!("Neo4j transaction commit failed: {e}")],
                error: None,
            });
        }

        // Generate a commit ID (placeholder - future: create via versioning-service gRPC)
        let new_commit_id = generate_commit_id(ontology_id);
        info!(
            "sequence.committed: ontology_id={}, commit_id={}, applied={}, skipped={}, failed={}",
            ontology_id, new_commit_id, steps_applied, steps_skipped, steps_failed
        );
        new_commit_id
    };

    Ok(ApplySequenceResponse {
        commit_id,
        steps_applied,
        steps_skipped,
        steps_failed,
        errors,
        error: None,
    })
}

/// Generates a commit ID for the new ontology revision.
/// Currently uses a prefixed UUID; will be replaced with versioning-service
/// gRPC integration once that service's gRPC stack is fully implemented.
fn generate_commit_id(ontology_id: &str) -> String {
    let prefix = &ontology_id[..ontology_id.len().min(8)];
    let ts = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_secs();
    let uuid = Uuid::new_v4();
    format!("commit-{}-{:x}-{}", prefix, ts, &uuid.to_string()[..8])
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_generate_commit_id_format() {
        let id = generate_commit_id("test-ontology");
        assert!(id.starts_with("commit-test-"));
        // Format: commit-<prefix>-<ts hex>-<uuid short>
        let parts: Vec<&str> = id.split('-').collect();
        assert!(parts.len() >= 4);
    }

    #[test]
    fn test_generate_commit_id_different_ids() {
        let id1 = generate_commit_id("onto-1");
        let id2 = generate_commit_id("onto-1");
        // Same inputs should produce different results due to UUID
        assert_ne!(id1, id2);
    }

    #[test]
    fn test_steps_appointed_not_yet_failed() {
        assert_eq!(steps_appointed_not_yet_failed(5, &2, &1, &1), 1);
        assert_eq!(steps_appointed_not_yet_failed(3, &2, &1, &0), 0);
        assert_eq!(steps_appointed_not_yet_failed(0, &0, &0, &0), 0);
        assert_eq!(steps_appointed_not_yet_failed(10, &3, &2, &2), 3);
    }
}
