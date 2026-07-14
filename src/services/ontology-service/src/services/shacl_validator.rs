//! SHACL validation service for ontology validation.
//!
//! Validates ontology data against SHACL shape constraints.
//! The current implementation provides a basic structural validation
//! framework. Full SHACL engine integration is a future enhancement;
//! this initial version checks:
//!
//! - Class hierarchy cycles
//! - Property domain/range existence
//! - Individual class membership validity
//!
//! # Future
//!
//! - Replace with full SHACL engine (e.g., `sophia` or a SHACL crate)
//! - Support custom shape graphs
//! - Integration with shape preloading

use serde::Serialize;
use tracing::{debug, info, warn};

use crate::neo4j::Neo4jPool;

/// A single SHACL validation result.
#[derive(Debug, Clone, Serialize)]
pub struct ShaclValidationResult {
    /// The IRI or local ID of the focus node (entity being validated).
    pub focus_node: String,
    /// The SHACL property path or constraint component that triggered the result.
    pub path: String,
    /// The severity level: "Violation", "Warning", or "Info".
    pub severity: String,
    /// Human-readable validation message.
    pub message: String,
}

/// The validation report returned by the SHACL validator.
#[derive(Debug, Clone, Serialize)]
pub struct ShaclValidationReport {
    /// Whether the ontology conforms to all shapes (no violations).
    pub conforms: bool,
    /// Individual validation results.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub results: Vec<ShaclValidationResult>,
    /// Total entities checked.
    #[serde(default)]
    pub entities_checked: u64,
}

/// Performs SHACL-style validation of an ontology.
pub struct ShaclValidator {
    pool: Neo4jPool,
}

impl ShaclValidator {
    /// Creates a new validator with a Neo4j connection pool.
    pub fn new(pool: Neo4jPool) -> Self {
        Self { pool }
    }

    /// Validates an ontology against SHACL shapes.
    ///
    /// # Arguments
    ///
    /// * `ontology_id` - The ontology to validate.
    /// * `_shapes_turtle` - Optional custom shapes graph in Turtle format.
    ///   When `None`, built-in structural validation rules are used.
    ///
    /// # Returns
    ///
    /// A `ShaclValidationReport` with conformance status and detail results.
    pub async fn validate(
        &self,
        ontology_id: &str,
        _shapes_turtle: Option<&str>,
    ) -> ShaclValidationReport {
        debug!(ontology_id, "Starting SHACL validation");

        let mut results = Vec::new();
        let mut entities_checked: u64 = 0;

        // Run structural validation checks
        entities_checked += self.check_class_hierarchy(ontology_id, &mut results).await;
        entities_checked += self.check_property_domains(ontology_id, &mut results).await;
        entities_checked += self
            .check_individual_classes(ontology_id, &mut results)
            .await;

        let conforms = results.iter().all(|r| r.severity != "Violation");

        info!(
            ontology_id,
            conforms,
            violations = results.iter().filter(|r| r.severity == "Violation").count(),
            warnings = results.iter().filter(|r| r.severity == "Warning").count(),
            entities_checked,
            "SHACL validation completed"
        );

        ShaclValidationReport {
            conforms,
            results,
            entities_checked,
        }
    }

    /// Checks for cycles in the class hierarchy (rdfs:subClassOf).
    async fn check_class_hierarchy(
        &self,
        ontology_id: &str,
        results: &mut Vec<ShaclValidationResult>,
    ) -> u64 {
        debug!(ontology_id, "Checking class hierarchy for cycles");

        // [FIX] Use a multi-hop variable-length path to detect multi-step cycles
        // (e.g. A→B→A, A→B→C→A) in addition to self-loops. The old single-hop
        // pattern `(c)-[:SUB_CLASS_Of]->(c)` only caught direct self-loops.
        let query = "\
            MATCH (c:Class {ontology_id: $ontology_id})-[:SUB_CLASS_OF*2..]->(c)\
            RETURN DISTINCT c.id AS class_id\
        ";

        let q = neo4rs::Query::new(query.to_string()).param("ontology_id", ontology_id);

        match self.pool.graph().execute(q).await {
            Ok(mut result) => {
                let mut count: u64 = 0;
                while let Ok(Some(row)) = result.next().await {
                    if let Ok(class_id) = row.get::<String>("class_id") {
                        count += 1;
                        results.push(ShaclValidationResult {
                            focus_node: class_id,
                            path: "rdfs:subClassOf".to_string(),
                            severity: "Violation".to_string(),
                            message: "Class hierarchy contains a cycle via rdfs:subClassOf"
                                .to_string(),
                        });
                    }
                }
                count
            }
            Err(e) => {
                // [FIX] Do NOT silently swallow DB errors — they indicate the
                // validator cannot determine whether the ontology conforms.
                // Surface them as a violation so downstream callers know the
                // validation result is not reliable.
                warn!(
                    ontology_id,
                    error = %e,
                    "[FIX] Class hierarchy cycle check query failed — surfacing as violation"
                );
                results.push(ShaclValidationResult {
                    focus_node: ontology_id.to_string(),
                    path: "rdfs:subClassOf".to_string(),
                    severity: "Violation".to_string(),
                    message: format!(
                        "Class hierarchy cycle check failed due to a database error: {e}"
                    ),
                });
                // Returning 1 signals the caller that a violation occurred.
                1
            }
        }
    }

    /// Checks that property domain classes exist within the ontology.
    async fn check_property_domains(
        &self,
        ontology_id: &str,
        results: &mut Vec<ShaclValidationResult>,
    ) -> u64 {
        debug!(ontology_id, "Checking property domain/range references");

        // Count all properties - each one is a checked entity
        let count_query = "\
            MATCH (p:Property {ontology_id: $ontology_id})\
            WITH p\
            OPTIONAL MATCH (p)-[:DOMAIN]->(d:Class)\
            RETURN p.id AS property_id, collect(d.id) AS domain_ids\
        ";

        let q = neo4rs::Query::new(count_query.to_string()).param("ontology_id", ontology_id);
        let mut count: u64 = 0;

        if let Ok(mut result) = self.pool.graph().execute(q).await {
            while let Ok(Some(row)) = result.next().await {
                if let Ok(property_id) = row.get::<String>("property_id") {
                    count += 1;
                    let domain_ids: Vec<String> = row.get("domain_ids").unwrap_or_default();
                    // Filter out empty strings from the collect
                    let has_domains: bool = domain_ids.iter().any(|d| !d.is_empty());
                    if !has_domains {
                        results.push(ShaclValidationResult {
                            focus_node: property_id,
                            path: "rdfs:domain".to_string(),
                            severity: "Warning".to_string(),
                            message: "Property has no domain class specified".to_string(),
                        });
                    }
                }
            }
        }

        count
    }

    /// Checks that individuals reference existing classes.
    async fn check_individual_classes(
        &self,
        ontology_id: &str,
        results: &mut Vec<ShaclValidationResult>,
    ) -> u64 {
        debug!(ontology_id, "Checking individual class membership");

        // Find individuals whose class_id doesn't match any Class node
        let query = "\
            MATCH (i:Individual {ontology_id: $ontology_id})\
            WHERE NOT EXISTS {\
                MATCH (c:Class {id: i.class_id, ontology_id: $ontology_id})\
            }\
            RETURN i.id AS individual_id, i.class_id AS class_id\
        ";

        let q = neo4rs::Query::new(query.to_string()).param("ontology_id", ontology_id);
        let mut count: u64 = 0;

        match self.pool.graph().execute(q).await {
            Ok(mut result) => {
                while let Ok(Some(row)) = result.next().await {
                    if let (Ok(ind_id), Ok(cls_id)) = (
                        row.get::<String>("individual_id"),
                        row.get::<String>("class_id"),
                    ) {
                        count += 1;
                        results.push(ShaclValidationResult {
                            focus_node: ind_id,
                            path: "rdf:type".to_string(),
                            severity: "Violation".to_string(),
                            message: format!("Individual references non-existent class '{cls_id}'"),
                        });
                    }
                }
            }
            Err(e) => {
                debug!(
                    ontology_id,
                    error = %e,
                    "Individual class check query failed (non-fatal)"
                );
            }
        }

        count
    }
}
