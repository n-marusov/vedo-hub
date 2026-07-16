//! Repository for executing ontology sequence steps against Neo4j.
//!
//! Each `ApplySequence` request contains an ordered list of `SequenceStep`
//! operations (create class, create property, create individual, etc.).
//! This repository provides methods to execute each step type within a
//! single Neo4j transaction, ensuring atomicity.

use neo4rs::Txn;
use tracing::{debug, warn};

use vedo_shared::protos::ontology::v1::SequenceStep;

/// Executes a single `SequenceStep` within the given transaction.
///
/// Returns `Ok(true)` if the step was applied, `Ok(false)` if it was
/// skipped (skip_if_exists + already exists), or an error on failure.
pub async fn execute_step(
    tx: &mut Txn,
    ontology_id: &str,
    step: &SequenceStep,
) -> Result<bool, String> {
    use vedo_shared::protos::ontology::v1::sequence_step::Operation;

    let operation = step.r#operation();
    let entity_id = &step.entity_id;
    let label = &step.label;

    debug!(
        "sequence.execute_step: op={:?} entity_id={} label={} ontology_id={}",
        operation, entity_id, label, ontology_id
    );

    match operation {
        Operation::CreateClass => create_class(tx, ontology_id, step).await,
        Operation::CreateObjectProperty => create_object_property(tx, ontology_id, step).await,
        Operation::CreateDatatypeProperty => create_datatype_property(tx, ontology_id, step).await,
        Operation::CreateIndividual => create_individual(tx, ontology_id, step).await,
        Operation::AddAnnotation => add_annotation(tx, ontology_id, step).await,
        Operation::SetParent => set_parent(tx, ontology_id, step).await,
        Operation::SetDomain => set_domain(tx, step).await,
        Operation::SetRange => set_range(tx, step).await,
        Operation::Unspecified => {
            warn!("sequence.unspecified_operation: entity_id={}", entity_id);
            Ok(true) // Skip unspecified operations silently
        }
    }
}

/// Checks if an entity with the given IRI already exists in the ontology.
async fn entity_exists(tx: &mut Txn, ontology_id: &str, entity_id: &str) -> Result<bool, String> {
    let query = neo4rs::query(
        "MATCH (e) WHERE e.id = $entity_id AND e.ontology_id = $ontology_id RETURN count(e) AS cnt",
    )
    .param("entity_id", entity_id.to_string())
    .param("ontology_id", ontology_id.to_string());

    let mut result = tx
        .execute(query)
        .await
        .map_err(|e| format!("entity_exists query failed: {e}"))?;
    if let Some(row) = result
        .next(&mut *tx)
        .await
        .map_err(|e| format!("entity_exists fetch failed: {e}"))?
    {
        let cnt: i64 = row
            .get::<Option<i64>>("cnt")
            .map_err(|e| format!("entity_exists parse failed: {e}"))?
            .unwrap_or(0);
        Ok(cnt > 0)
    } else {
        Ok(false)
    }
}

/// Creates a class node with label and optional annotations.
async fn create_class(
    tx: &mut Txn,
    ontology_id: &str,
    step: &SequenceStep,
) -> Result<bool, String> {
    let entity_id = &step.entity_id;
    let label = &step.label;

    // Check for existing entity
    if step.skip_if_exists && entity_exists(tx, ontology_id, entity_id).await? {
        debug!("sequence.skip_class: entity_id={}", entity_id);
        return Ok(false);
    }

    let parent_id = if step.parent_id.is_empty() {
        "owl:Thing".to_string()
    } else {
        step.parent_id.clone()
    };

    let query = neo4rs::query(
        r"
        CREATE (c:Class {
            id: $entity_id,
            label: $label,
            ontology_id: $ontology_id,
            parent_id: $parent_id,
            created_at: timestamp(),
            updated_at: timestamp()
        })
        RETURN c.id AS id
        ",
    )
    .param("entity_id", entity_id.to_string())
    .param("label", label.to_string())
    .param("ontology_id", ontology_id.to_string())
    .param("parent_id", parent_id);

    tx.run(query)
        .await
        .map_err(|e| format!("create_class failed for {entity_id}: {e}"))?;

    // Create subClassOf relationship to parent
    if !step.parent_id.is_empty() {
        let rel_query = neo4rs::query(
            r"
            MATCH (child:Class {id: $entity_id, ontology_id: $ontology_id})
            MATCH (parent {id: $parent_id, ontology_id: $ontology_id})
            MERGE (child)-[:SUB_CLASS_OF]->(parent)
            ",
        )
        .param("entity_id", entity_id.to_string())
        .param("ontology_id", ontology_id.to_string())
        .param("parent_id", step.parent_id.clone());

        tx.run(rel_query)
            .await
            .map_err(|e| format!("create_class relationship failed for {entity_id}: {e}"))?;
    }

    // Store annotations as node properties
    for annotation in &step.annotations {
        let ann_query = neo4rs::query(
            r"
            MATCH (c:Class {id: $entity_id, ontology_id: $ontology_id})
            SET c += {annotations: coalesce(c.annotations, []) + $annotation}
            ",
        )
        .param("entity_id", entity_id.to_string())
        .param("ontology_id", ontology_id.to_string())
        .param("annotation", annotation.clone());

        tx.run(ann_query)
            .await
            .map_err(|e| format!("create_class annotation failed for {entity_id}: {e}"))?;
    }

    debug!("sequence.class_created: entity_id={}", entity_id);
    Ok(true)
}

/// Creates an object property node with domain and range.
async fn create_object_property(
    tx: &mut Txn,
    ontology_id: &str,
    step: &SequenceStep,
) -> Result<bool, String> {
    let entity_id = &step.entity_id;
    let label = &step.label;

    if step.skip_if_exists && entity_exists(tx, ontology_id, entity_id).await? {
        debug!("sequence.skip_property: entity_id={}", entity_id);
        return Ok(false);
    }

    let query = neo4rs::query(
        r"
        CREATE (p:ObjectProperty {
            id: $entity_id,
            label: $label,
            ontology_id: $ontology_id,
            domain_id: $domain_id,
            range_id: $range_id,
            created_at: timestamp(),
            updated_at: timestamp()
        })
        RETURN p.id AS id
        ",
    )
    .param("entity_id", entity_id.to_string())
    .param("label", label.to_string())
    .param("ontology_id", ontology_id.to_string())
    .param("domain_id", step.domain_id.clone())
    .param("range_id", step.range_id.clone());

    tx.run(query)
        .await
        .map_err(|e| format!("create_object_property failed for {entity_id}: {e}"))?;

    debug!("sequence.object_property_created: entity_id={}", entity_id);
    Ok(true)
}

/// Creates a datatype property node with domain and xsd range.
async fn create_datatype_property(
    tx: &mut Txn,
    ontology_id: &str,
    step: &SequenceStep,
) -> Result<bool, String> {
    let entity_id = &step.entity_id;
    let label = &step.label;

    if step.skip_if_exists && entity_exists(tx, ontology_id, entity_id).await? {
        debug!("sequence.skip_datatype_property: entity_id={}", entity_id);
        return Ok(false);
    }

    let query = neo4rs::query(
        r"
        CREATE (p:DatatypeProperty {
            id: $entity_id,
            label: $label,
            ontology_id: $ontology_id,
            domain_id: $domain_id,
            range_id: $range_id,
            created_at: timestamp(),
            updated_at: timestamp()
        })
        RETURN p.id AS id
        ",
    )
    .param("entity_id", entity_id.to_string())
    .param("label", label.to_string())
    .param("ontology_id", ontology_id.to_string())
    .param("domain_id", step.domain_id.clone())
    .param("range_id", step.range_id.clone());

    tx.run(query)
        .await
        .map_err(|e| format!("create_datatype_property failed for {entity_id}: {e}"))?;

    debug!(
        "sequence.datatype_property_created: entity_id={}",
        entity_id
    );
    Ok(true)
}

/// Creates an individual (instance) of a class type.
async fn create_individual(
    tx: &mut Txn,
    ontology_id: &str,
    step: &SequenceStep,
) -> Result<bool, String> {
    let entity_id = &step.entity_id;
    let label = &step.label;

    if step.skip_if_exists && entity_exists(tx, ontology_id, entity_id).await? {
        debug!("sequence.skip_individual: entity_id={}", entity_id);
        return Ok(false);
    }

    let query = neo4rs::query(
        r"
        CREATE (i:Individual {
            id: $entity_id,
            label: $label,
            ontology_id: $ontology_id,
            type_id: $type_id,
            created_at: timestamp(),
            updated_at: timestamp()
        })
        WITH i
        MATCH (c {id: $type_id, ontology_id: $ontology_id})
        MERGE (i)-[:INSTANCE_OF]->(c)
        RETURN i.id AS id
        ",
    )
    .param("entity_id", entity_id.to_string())
    .param("label", label.to_string())
    .param("ontology_id", ontology_id.to_string())
    .param("type_id", step.parent_id.clone());

    tx.run(query)
        .await
        .map_err(|e| format!("create_individual failed for {entity_id}: {e}"))?;

    debug!("sequence.individual_created: entity_id={}", entity_id);
    Ok(true)
}

/// Adds an annotation to an existing entity.
async fn add_annotation(
    tx: &mut Txn,
    ontology_id: &str,
    step: &SequenceStep,
) -> Result<bool, String> {
    let entity_id = &step.entity_id;

    for annotation in &step.annotations {
        let query = neo4rs::query(
            r"
            MATCH (e {id: $entity_id, ontology_id: $ontology_id})
            SET e.annotations = coalesce(e.annotations, []) + $annotation
            ",
        )
        .param("entity_id", entity_id.to_string())
        .param("ontology_id", ontology_id.to_string())
        .param("annotation", annotation.clone());

        tx.run(query)
            .await
            .map_err(|e| format!("add_annotation failed for {entity_id}: {e}"))?;
    }

    debug!(
        "sequence.annotations_added: entity_id={}, count={}",
        entity_id,
        step.annotations.len()
    );
    Ok(true)
}

/// Sets the parent of an existing class.
async fn set_parent(tx: &mut Txn, ontology_id: &str, step: &SequenceStep) -> Result<bool, String> {
    let entity_id = &step.entity_id;
    let parent_id = &step.parent_id;

    // Remove old parent relationship
    let remove_query = neo4rs::query(
        r"
        MATCH (child {id: $entity_id, ontology_id: $ontology_id})-[r:SUB_CLASS_OF]->()
        DELETE r
        ",
    )
    .param("entity_id", entity_id.to_string())
    .param("ontology_id", ontology_id.to_string());
    let _ = tx.run(remove_query).await; // May succeed even without existing parent

    // Create new parent relationship
    if !parent_id.is_empty() {
        let set_query = neo4rs::query(
            r"
            MATCH (child {id: $entity_id, ontology_id: $ontology_id})
            MATCH (parent {id: $parent_id, ontology_id: $ontology_id})
            MERGE (child)-[:SUB_CLASS_OF]->(parent)
            ",
        )
        .param("entity_id", entity_id.to_string())
        .param("ontology_id", ontology_id.to_string())
        .param("parent_id", parent_id.clone());

        tx.run(set_query)
            .await
            .map_err(|e| format!("set_parent failed for {entity_id}: {e}"))?;

        // Update parent_id property
        let update_query = neo4rs::query(
            r"
            MATCH (e {id: $entity_id, ontology_id: $ontology_id})
            SET e.parent_id = $parent_id
            ",
        )
        .param("entity_id", entity_id.to_string())
        .param("ontology_id", ontology_id.to_string())
        .param("parent_id", parent_id.clone());
        tx.run(update_query)
            .await
            .map_err(|e| format!("set_parent update failed for {entity_id}: {e}"))?;
    }

    debug!(
        "sequence.parent_set: entity_id={}, parent_id={}",
        entity_id, parent_id
    );
    Ok(true)
}

/// Sets the domain of an existing property.
async fn set_domain(tx: &mut Txn, step: &SequenceStep) -> Result<bool, String> {
    let entity_id = &step.entity_id;
    let domain_id = &step.domain_id;

    let query = neo4rs::query(
        r"
        MATCH (p {id: $entity_id})
        SET p.domain_id = $domain_id
        ",
    )
    .param("entity_id", entity_id.to_string())
    .param("domain_id", domain_id.clone());

    tx.run(query)
        .await
        .map_err(|e| format!("set_domain failed for {entity_id}: {e}"))?;

    debug!(
        "sequence.domain_set: entity_id={}, domain_id={}",
        entity_id, domain_id
    );
    Ok(true)
}

/// Sets the range of an existing property.
async fn set_range(tx: &mut Txn, step: &SequenceStep) -> Result<bool, String> {
    let entity_id = &step.entity_id;
    let range_id = &step.range_id;

    let query = neo4rs::query(
        r"
        MATCH (p {id: $entity_id})
        SET p.range_id = $range_id
        ",
    )
    .param("entity_id", entity_id.to_string())
    .param("range_id", range_id.clone());

    tx.run(query)
        .await
        .map_err(|e| format!("set_range failed for {entity_id}: {e}"))?;

    debug!(
        "sequence.range_set: entity_id={}, range_id={}",
        entity_id, range_id
    );
    Ok(true)
}
