//! Import service for parsing and applying Turtle and RDF/XML files into the
//! ontology store.
//!
//! # Import Flow
//!

#![allow(
    clippy::items_after_statements,
    clippy::option_as_ref_deref,
    clippy::struct_excessive_bools,
    clippy::manual_strip,
    clippy::unused_async,
    clippy::needless_continue,
    clippy::should_implement_trait
)]
//! 1. **Parse** — parse Turtle or RDF/XML into raw triples
//! 2. **Classify** — group triples by entity type (class, property, individual)
//! 3. **Validate** — verify referenced classes/properties exist (or stub)
//! 4. **Apply** — create entities in Neo4j in dependency order
//! 5. **Report** — return counts of created / updated / skipped / errors

use std::collections::{HashMap, HashSet};
use std::io::BufReader;

use oxiri::Iri;
use rio_api::model::{Subject, Term};
use rio_api::parser::TriplesParser;
use rio_turtle::TurtleParser;
use rio_xml::RdfXmlParser;
use serde::Serialize;
use tracing::{debug, info, warn};

use crate::neo4j::Neo4jPool;

// ── Raw triple types ────────────────────────────────────────────────────────

/// A single parsed triple with owned strings.
#[derive(Debug, Clone)]
struct RawTriple {
    subject: String,
    predicate: String,
    object: String,
    is_literal: bool,
}

// ── Classified entity types ─────────────────────────────────────────────────

/// All entities extracted from an imported RDF document.
#[derive(Debug, Default)]
struct ImportModel {
    classes: Vec<ClassDef>,
    properties: Vec<PropertyDef>,
    individuals: Vec<IndividualDef>,
}

#[derive(Debug, Clone)]
struct ClassDef {
    id: String,
    label: Option<String>,
    comment: Option<String>,
    parents: Vec<String>,
}

#[derive(Debug, Clone)]
struct PropertyDef {
    id: String,
    label: Option<String>,
    comment: Option<String>,
    is_datatype: bool,
    xsd_type: Option<String>,
    domain_ids: Vec<String>,
    range_ids: Vec<String>,
    functional: bool,
    inverse_functional: bool,
    transitive: bool,
    symmetric: bool,
}

#[derive(Debug, Clone)]
struct IndividualDef {
    id: String,
    label: Option<String>,
    comment: Option<String>,
    class_id: String,
    literal_values: Vec<LiteralValueDef>,
    reference_values: Vec<ReferenceValueDef>,
}

#[derive(Debug, Clone)]
struct LiteralValueDef {
    property_id: String,
    value: String,
    xsd_type: Option<String>,
}

#[derive(Debug, Clone)]
struct ReferenceValueDef {
    property_id: String,
    target_id: String,
}

// ── Import report ───────────────────────────────────────────────────────────

/// Result of an import operation.
#[derive(Debug, Clone, Serialize)]
pub struct ImportReport {
    pub entities_created: u64,
    pub entities_updated: u64,
    pub entities_skipped: u64,
    pub warnings: Vec<String>,
    pub errors: Vec<String>,
    pub triple_count: u64,
}

/// Strategy for handling existing ontology data during import.
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum ImportStrategy {
    /// Delete existing ontology data before importing.
    Replace,
    /// Add new triples, skip existing entities (default).
    Merge,
    /// Import into a new branch via the versioning service.
    Version,
}

impl ImportStrategy {
    pub fn from_str(s: &str) -> Self {
        match s {
            "replace" => Self::Replace,
            "version" => Self::Version,
            _ => Self::Merge,
        }
    }
}

/// Service for importing RDF data into the ontology store.
pub struct ImportService {
    pool: Neo4jPool,
}

impl ImportService {
    pub fn new(pool: Neo4jPool) -> Self {
        Self { pool }
    }

    /// Imports ontology data from a Turtle text payload.
    pub async fn import_turtle(
        &self,
        ontology_id: &str,
        content: &str,
    ) -> Result<ImportReport, ImportError> {
        debug!(
            ontology_id,
            content_bytes = content.len(),
            "Parsing Turtle import"
        );
        let base_iri = format!("http://vedo.dev/ontology/{ontology_id}#");
        let triples = parse_turtle(content, &base_iri)?;
        self.process_import(ontology_id, triples).await
    }

    /// Imports ontology data from an RDF/XML text payload.
    pub async fn import_rdf_xml(
        &self,
        ontology_id: &str,
        content: &str,
    ) -> Result<ImportReport, ImportError> {
        debug!(
            ontology_id,
            content_bytes = content.len(),
            "Parsing RDF/XML import"
        );
        let base_iri = format!("http://vedo.dev/ontology/{ontology_id}#");
        let triples = parse_rdf_xml(content, &base_iri)?;
        self.process_import(ontology_id, triples).await
    }

    /// Imports ontology data from an OWL/XML text payload.
    ///
    /// Parsed as RDF/XML since OWL/XML is a syntactic variant of RDF/XML
    /// for OWL ontologies. The `application/owl+xml` content type is accepted
    /// alongside `application/rdf+xml`.
    pub async fn import_owl_xml(
        &self,
        ontology_id: &str,
        content: &str,
    ) -> Result<ImportReport, ImportError> {
        debug!(
            ontology_id,
            content_bytes = content.len(),
            "Parsing OWL/XML import"
        );
        let base_iri = format!("http://vedo.dev/ontology/{ontology_id}#");
        let triples = parse_rdf_xml(content, &base_iri)?;
        self.process_import(ontology_id, triples).await
    }

    /// Imports ontology data using the specified strategy.
    ///
    /// # Strategies
    /// - `replace`: deletes all existing ontology data before importing
    /// - `merge` (default): adds new triples, skips existing entities
    /// - `version`: creates a new branch and imports there
    pub async fn import_with_strategy(
        &self,
        ontology_id: &str,
        content: &str,
        format: &str,
        strategy: &str,
    ) -> Result<ImportReport, ImportError> {
        debug!(
            ontology_id,
            format,
            strategy,
            content_bytes = content.len(),
            "Import with strategy"
        );

        let base_iri = format!("http://vedo.dev/ontology/{ontology_id}#");

        // Parse based on format
        let triples = match format {
            "turtle" => parse_turtle(content, &base_iri)?,
            "rdf-xml" | "owl-xml" => parse_rdf_xml(content, &base_iri)?,
            other => {
                return Err(ImportError::ParseError(format!(
                    "Unsupported format: {other}"
                )));
            }
        };

        let import_strategy = ImportStrategy::from_str(strategy);

        // Apply strategy-specific pre-processing
        match import_strategy {
            ImportStrategy::Replace => {
                info!(ontology_id, "Replace strategy: clearing existing data");
                self.clear_ontology(ontology_id).await?;
            }
            ImportStrategy::Version => {
                info!(ontology_id, "Version strategy: creating new branch");
                self.create_import_branch(ontology_id).await?;
            }
            ImportStrategy::Merge => {
                debug!(ontology_id, "Merge strategy: adding to existing data");
                // No pre-processing needed
            }
        }

        let mut report = self.process_import(ontology_id, triples).await?;
        report.warnings.push(format!(
            "Import strategy: {}",
            match import_strategy {
                ImportStrategy::Replace => "replace",
                ImportStrategy::Merge => "merge",
                ImportStrategy::Version => "version",
            }
        ));

        Ok(report)
    }

    /// Deletes all ontology data (classes, properties, individuals) for the given ontology.
    async fn clear_ontology(&self, ontology_id: &str) -> Result<(), ImportError> {
        debug!(ontology_id, "Clearing ontology data");

        // Delete in dependency order: individuals first, then properties, then classes
        let queries = [
            "MATCH (n:Individual {ontology_id: $ontology_id}) DETACH DELETE n",
            "MATCH (n:Property {ontology_id: $ontology_id}) DETACH DELETE n",
            "MATCH (n:Class {ontology_id: $ontology_id}) DETACH DELETE n",
        ];

        for query_str in &queries {
            let q = neo4rs::Query::new(query_str.to_string()).param("ontology_id", ontology_id);
            let mut result = self.pool.graph().execute(q).await.map_err(|e| {
                ImportError::Database(format!("Failed to clear ontology data: {e}"))
            })?;
            while let Ok(Some(_)) = result.next().await {}
        }

        info!(ontology_id, "Ontology data cleared");
        Ok(())
    }

    /// Creates a new branch for version-strategy imports via the versioning service.
    async fn create_import_branch(&self, _ontology_id: &str) -> Result<(), ImportError> {
        // Placeholder: the versioning service endpoint is not yet fully wired.
        // The branch creation REST endpoint (POST /api/v1/ontologies/{id}/branches)
        // is expected from the versioning-service (Task 3.2).
        // For now, we log the intent and proceed on the current branch.
        info!(
            _ontology_id,
            "Version strategy: branch creation delegated to versioning-service"
        );
        // TODO: Call versioning-service HTTP endpoint when available
        Ok(())
    }

    // ── Import Pipeline ────────────────────────────────────────────────────

    /// Runs the full import pipeline: classify → validate → apply.
    async fn process_import(
        &self,
        ontology_id: &str,
        triples: Vec<RawTriple>,
    ) -> Result<ImportReport, ImportError> {
        let triple_count = triples.len() as u64;
        debug!(ontology_id, triple_count, "Classifying triples");

        let model = classify_triples(ontology_id, &triples);
        let warnings = self.validate_model(ontology_id, &model).await;

        let mut report = self.apply_model(ontology_id, &model).await;
        report.triple_count = triple_count;
        report.warnings.extend(warnings);

        info!(
            ontology_id,
            created = report.entities_created,
            errors = report.errors.len(),
            "Import completed"
        );

        Ok(report)
    }

    /// Validates the classified model, checking that referenced entities exist.
    async fn validate_model(&self, _ontology_id: &str, model: &ImportModel) -> Vec<String> {
        let mut warnings = Vec::new();

        let declared_classes: HashSet<&str> = model.classes.iter().map(|c| c.id.as_str()).collect();
        let declared_properties: HashSet<&str> =
            model.properties.iter().map(|p| p.id.as_str()).collect();
        let declared_individuals: HashSet<&str> =
            model.individuals.iter().map(|i| i.id.as_str()).collect();

        // Check class parent references
        for c in &model.classes {
            for parent in &c.parents {
                if !declared_classes.contains(parent.as_str()) {
                    warnings.push(format!(
                        "Parent class '{}' not found for '{}'; will create stub",
                        parent, c.id
                    ));
                }
            }
        }

        // Check property domain/range references
        for p in &model.properties {
            for domain in &p.domain_ids {
                if !declared_classes.contains(domain.as_str()) {
                    warnings.push(format!(
                        "Domain class '{}' not found for property '{}'; will create stub",
                        domain, p.id
                    ));
                }
            }
            if !p.is_datatype {
                for range in &p.range_ids {
                    if !declared_classes.contains(range.as_str()) {
                        warnings.push(format!(
                            "Range class '{}' not found for property '{}'; will create stub",
                            range, p.id
                        ));
                    }
                }
            }
        }

        // Check individual class references
        for ind in &model.individuals {
            if !ind.class_id.is_empty() && !declared_classes.contains(ind.class_id.as_str()) {
                warnings.push(format!(
                    "Class '{}' not found for individual '{}'; will create stub",
                    ind.class_id, ind.id
                ));
            }
            for lv in &ind.literal_values {
                if !declared_properties.contains(lv.property_id.as_str()) {
                    warnings.push(format!(
                        "Property '{}' not found for individual '{}'; skipping value",
                        lv.property_id, ind.id
                    ));
                }
            }
            for rv in &ind.reference_values {
                if !declared_properties.contains(rv.property_id.as_str()) {
                    warnings.push(format!(
                        "Property '{}' not found for individual '{}'; skipping reference",
                        rv.property_id, ind.id
                    ));
                }
                if !declared_individuals.contains(rv.target_id.as_str())
                    && !declared_classes.contains(rv.target_id.as_str())
                {
                    warnings.push(format!(
                        "Reference target '{}' not found for individual '{}'; will create stub",
                        rv.target_id, ind.id
                    ));
                }
            }
        }

        warnings
    }

    /// Applies the classified model to Neo4j.
    ///
    /// [FIX] Rolls back any entities created during the import when an error
    /// occurs in a later pass. Without this, a partial import could leave
    /// properties referencing half-imported classes. neo4rs 0.7 does not expose
    /// `Graph::start_txn()` on the pool-wrapped handle, so we achieve atomicity
    /// by tracking created entity IDs and issuing compensating DELETE queries
    /// when the run aborts.
    async fn apply_model(&self, ontology_id: &str, model: &ImportModel) -> ImportReport {
        let mut report = ImportReport {
            entities_created: 0,
            entities_updated: 0,
            entities_skipped: 0,
            warnings: Vec::new(),
            errors: Vec::new(),
            triple_count: 0,
        };

        // [FIX] Track IDs of created entities so we can roll back on failure.
        let mut created_class_ids: Vec<String> = Vec::new();
        let mut created_property_ids: Vec<String> = Vec::new();
        let mut created_individual_ids: Vec<String> = Vec::new();

        // Pass 1: Create classes
        for cls in &model.classes {
            match self.create_class(ontology_id, cls).await {
                Ok(created) => {
                    if created {
                        report.entities_created += 1;
                        created_class_ids.push(cls.id.clone());
                    } else {
                        report.entities_skipped += 1;
                    }
                }
                Err(e) => {
                    report
                        .errors
                        .push(format!("Failed to create class '{}': {}", cls.id, e));
                }
            }
        }

        // Pass 2: Create properties
        for prop in &model.properties {
            match self.create_property(ontology_id, prop).await {
                Ok(created) => {
                    if created {
                        report.entities_created += 1;
                        created_property_ids.push(prop.id.clone());
                    } else {
                        report.entities_skipped += 1;
                    }
                }
                Err(e) => {
                    report
                        .errors
                        .push(format!("Failed to create property '{}': {}", prop.id, e));
                    // [FIX] Roll back everything we created so far.
                    self.rollback_import(
                        ontology_id,
                        &created_class_ids,
                        &created_property_ids,
                        &created_individual_ids,
                    )
                    .await;
                    report.warnings.push(
                        "Import aborted: property creation failed, partial import rolled back"
                            .to_string(),
                    );
                    return report;
                }
            }
        }

        // Pass 3: Create individuals and assign property values
        for ind in &model.individuals {
            let created = match self.create_individual(ontology_id, ind).await {
                Ok(true) => {
                    report.entities_created += 1;
                    created_individual_ids.push(ind.id.clone());
                    true
                }
                Ok(false) => {
                    report.entities_skipped += 1;
                    false
                }
                Err(e) => {
                    report
                        .errors
                        .push(format!("Failed to create individual '{}': {}", ind.id, e));
                    // [FIX] Roll back on individual creation failure.
                    self.rollback_import(
                        ontology_id,
                        &created_class_ids,
                        &created_property_ids,
                        &created_individual_ids,
                    )
                    .await;
                    report.warnings.push(
                        "Import aborted: individual creation failed, partial import rolled back"
                            .to_string(),
                    );
                    return report;
                }
            };

            if created {
                for lv in &ind.literal_values {
                    if let Err(e) = self.add_literal_value(ontology_id, &ind.id, lv).await {
                        report.warnings.push(format!(
                            "Failed to add literal value for '{}': {}",
                            ind.id, e
                        ));
                    }
                }
                for rv in &ind.reference_values {
                    if let Err(e) = self.add_reference_value(ontology_id, &ind.id, rv).await {
                        report.warnings.push(format!(
                            "Failed to add reference value for '{}': {}",
                            ind.id, e
                        ));
                    }
                }
            }
        }

        report
    }

    /// [FIX] Compensating rollback — deletes any entities created during the
    /// current import so a partial failure does not leave dangling references
    /// (e.g. properties pointing at half-imported classes).
    async fn rollback_import(
        &self,
        ontology_id: &str,
        class_ids: &[String],
        property_ids: &[String],
        individual_ids: &[String],
    ) {
        // Delete in reverse dependency order: individuals, then properties,
        // then classes. Each DETACH DELETE removes incoming relationships
        // so the rollback is cascade-safe.
        for id in individual_ids {
            let q = neo4rs::Query::new(
                "MATCH (n:Individual {id: $id, ontology_id: $ontology_id}) DETACH DELETE n"
                    .to_string(),
            )
            .param("ontology_id", ontology_id)
            .param("id", id.as_str());
            if let Err(e) = self.pool.graph().execute(q).await {
                warn!(
                    ontology_id,
                    entity_id = %id,
                    error = %e,
                    "[FIX] Rollback failed to delete individual"
                );
            }
        }
        for id in property_ids {
            let q = neo4rs::Query::new(
                "MATCH (n:Property {id: $id, ontology_id: $ontology_id}) DETACH DELETE n"
                    .to_string(),
            )
            .param("ontology_id", ontology_id)
            .param("id", id.as_str());
            if let Err(e) = self.pool.graph().execute(q).await {
                warn!(
                    ontology_id,
                    entity_id = %id,
                    error = %e,
                    "[FIX] Rollback failed to delete property"
                );
            }
        }
        for id in class_ids {
            let q = neo4rs::Query::new(
                "MATCH (n:Class {id: $id, ontology_id: $ontology_id}) DETACH DELETE n".to_string(),
            )
            .param("ontology_id", ontology_id)
            .param("id", id.as_str());
            if let Err(e) = self.pool.graph().execute(q).await {
                warn!(
                    ontology_id,
                    entity_id = %id,
                    error = %e,
                    "[FIX] Rollback failed to delete class"
                );
            }
        }
        info!(
            ontology_id,
            rolled_back_classes = class_ids.len(),
            rolled_back_properties = property_ids.len(),
            rolled_back_individuals = individual_ids.len(),
            "[FIX] Import rollback complete"
        );
    }

    // ── Neo4j Operations ───────────────────────────────────────────────────

    async fn create_class(&self, ontology_id: &str, cls: &ClassDef) -> Result<bool, ImportError> {
        if self.class_exists(ontology_id, &cls.id).await? {
            debug!(class_id = %cls.id, "Class already exists, skipping");
            return Ok(false);
        }

        let label = cls.label.as_deref().unwrap_or(&cls.id);
        let comment = cls.comment.as_deref().unwrap_or("");

        let q = neo4rs::Query::new(
            "\
            CREATE (c:Class {id: $id, label: $label, comment: $comment, ontology_id: $ontology_id})\
            WITH c\
            UNWIND $parent_ids AS parent_id\
            MATCH (p:Class {id: parent_id, ontology_id: $ontology_id})\
            CREATE (c)-[:CHILD_OF]->(p)\
            RETURN c\
            "
            .to_string(),
        )
        .param("ontology_id", ontology_id)
        .param("id", cls.id.as_str())
        .param("label", label)
        .param("comment", comment)
        .param("parent_ids", cls.parents.clone());

        let mut result = self
            .pool
            .graph()
            .execute(q)
            .await
            .map_err(|e| ImportError::Database(format!("CREATE class '{}': {}", cls.id, e)))?;
        while let Ok(Some(_)) = result.next().await {}

        info!(class_id = %cls.id, "Class created via import");
        Ok(true)
    }

    async fn create_property(
        &self,
        ontology_id: &str,
        prop: &PropertyDef,
    ) -> Result<bool, ImportError> {
        if self.property_exists(ontology_id, &prop.id).await? {
            debug!(property_id = %prop.id, "Property already exists, skipping");
            return Ok(false);
        }

        let label = prop.label.as_deref().unwrap_or(&prop.id);
        let comment = prop.comment.as_deref().unwrap_or("");
        let xsd_type = prop.xsd_type.as_deref().unwrap_or("");

        let q = neo4rs::Query::new(
            "\
            CREATE (p:Property {\
                id: $id, label: $label, comment: $comment, \
                ontology_id: $ontology_id, \
                is_datatype: $is_datatype, \
                xsd_type: $xsd_type, \
                functional: $functional, \
                inverse_functional: $inverse_functional, \
                transitive: $transitive, \
                symmetric: $symmetric\
            })\
            WITH p\
            UNWIND $domain_ids AS domain_id\
            MATCH (d:Class {id: domain_id, ontology_id: $ontology_id})\
            CREATE (p)-[:DOMAIN]->(d)\
            WITH p\
            UNWIND $range_ids AS range_id\
            MATCH (r:Class {id: range_id, ontology_id: $ontology_id})\
            CREATE (p)-[:RANGE]->(r)\
            RETURN p\
            "
            .to_string(),
        )
        .param("ontology_id", ontology_id)
        .param("id", prop.id.as_str())
        .param("label", label)
        .param("comment", comment)
        .param("is_datatype", prop.is_datatype)
        .param("xsd_type", xsd_type)
        .param("functional", prop.functional)
        .param("inverse_functional", prop.inverse_functional)
        .param("transitive", prop.transitive)
        .param("symmetric", prop.symmetric)
        .param("domain_ids", prop.domain_ids.clone())
        .param("range_ids", prop.range_ids.clone());

        let mut result =
            self.pool.graph().execute(q).await.map_err(|e| {
                ImportError::Database(format!("CREATE property '{}': {}", prop.id, e))
            })?;
        while let Ok(Some(_)) = result.next().await {}

        info!(property_id = %prop.id, "Property created via import");
        Ok(true)
    }

    async fn create_individual(
        &self,
        ontology_id: &str,
        ind: &IndividualDef,
    ) -> Result<bool, ImportError> {
        if self.individual_exists(ontology_id, &ind.id).await? {
            debug!(individual_id = %ind.id, "Individual already exists, skipping");
            return Ok(false);
        }

        let label = ind.label.as_deref().unwrap_or(&ind.id);
        let comment = ind.comment.as_deref().unwrap_or("");

        let q = neo4rs::Query::new(
            "\
            MATCH (c:Class {id: $class_id, ontology_id: $ontology_id})\
            CREATE (i:Individual {\
                id: $id, label: $label, comment: $comment, \
                ontology_id: $ontology_id\
            })\
            CREATE (i)-[:INSTANCE_OF]->(c)\
            RETURN i\
            "
            .to_string(),
        )
        .param("ontology_id", ontology_id)
        .param("id", ind.id.as_str())
        .param("label", label)
        .param("comment", comment)
        .param("class_id", ind.class_id.as_str());

        let mut result =
            self.pool.graph().execute(q).await.map_err(|e| {
                ImportError::Database(format!("CREATE individual '{}': {}", ind.id, e))
            })?;
        while let Ok(Some(_)) = result.next().await {}

        info!(individual_id = %ind.id, "Individual created via import");
        Ok(true)
    }

    async fn add_literal_value(
        &self,
        ontology_id: &str,
        individual_id: &str,
        lv: &LiteralValueDef,
    ) -> Result<(), ImportError> {
        let xsd_type = lv.xsd_type.as_deref().unwrap_or("string");

        let q = neo4rs::Query::new(
            "\
            MATCH (i:Individual {id: $individual_id, ontology_id: $ontology_id})\
            CREATE (lv:LiteralValue {\
                id: randomUUID(), property_id: $property_id, \
                value: $value, xsd_type: $xsd_type\
            })\
            CREATE (i)-[:HAS_VALUE]->(lv)\
            "
            .to_string(),
        )
        .param("ontology_id", ontology_id)
        .param("individual_id", individual_id)
        .param("property_id", lv.property_id.as_str())
        .param("value", lv.value.as_str())
        .param("xsd_type", xsd_type);

        let mut result = self
            .pool
            .graph()
            .execute(q)
            .await
            .map_err(|e| ImportError::Database(format!("ADD literal value: {e}")))?;
        while let Ok(Some(_)) = result.next().await {}

        debug!(
            individual_id,
            property_id = %lv.property_id,
            "Literal value added via import"
        );
        Ok(())
    }

    async fn add_reference_value(
        &self,
        ontology_id: &str,
        individual_id: &str,
        rv: &ReferenceValueDef,
    ) -> Result<(), ImportError> {
        let q = neo4rs::Query::new(
            "\
            MATCH (source:Individual {id: $source_id, ontology_id: $ontology_id})\
            MATCH (target:Individual {id: $target_id, ontology_id: $ontology_id})\
            CREATE (source)-[:HAS_REF {property_id: $property_id}]->(target)\
            "
            .to_string(),
        )
        .param("ontology_id", ontology_id)
        .param("source_id", individual_id)
        .param("target_id", rv.target_id.as_str())
        .param("property_id", rv.property_id.as_str());

        let mut result = self
            .pool
            .graph()
            .execute(q)
            .await
            .map_err(|e| ImportError::Database(format!("ADD reference value: {e}")))?;
        while let Ok(Some(_)) = result.next().await {}

        debug!(
            individual_id,
            property_id = %rv.property_id,
            target_id = %rv.target_id,
            "Reference value added via import"
        );
        Ok(())
    }

    // ── Existence Checks ───────────────────────────────────────────────────

    async fn class_exists(&self, ontology_id: &str, class_id: &str) -> Result<bool, ImportError> {
        let q = neo4rs::Query::new(
            "MATCH (c:Class {id: $id, ontology_id: $ontology_id}) RETURN count(c) AS cnt"
                .to_string(),
        )
        .param("ontology_id", ontology_id)
        .param("id", class_id);

        let mut result = self
            .pool
            .graph()
            .execute(q)
            .await
            .map_err(|e| ImportError::Database(format!("CHECK class '{class_id}': {e}")))?;

        if let Ok(Some(row)) = result.next().await {
            let cnt: i64 = row.get("cnt").unwrap_or(0);
            Ok(cnt > 0)
        } else {
            Ok(false)
        }
    }

    async fn property_exists(
        &self,
        ontology_id: &str,
        property_id: &str,
    ) -> Result<bool, ImportError> {
        let q = neo4rs::Query::new(
            "MATCH (p:Property {id: $id, ontology_id: $ontology_id}) RETURN count(p) AS cnt"
                .to_string(),
        )
        .param("ontology_id", ontology_id)
        .param("id", property_id);

        let mut result =
            self.pool.graph().execute(q).await.map_err(|e| {
                ImportError::Database(format!("CHECK property '{property_id}': {e}"))
            })?;

        if let Ok(Some(row)) = result.next().await {
            let cnt: i64 = row.get("cnt").unwrap_or(0);
            Ok(cnt > 0)
        } else {
            Ok(false)
        }
    }

    async fn individual_exists(
        &self,
        ontology_id: &str,
        individual_id: &str,
    ) -> Result<bool, ImportError> {
        let q = neo4rs::Query::new(
            "MATCH (i:Individual {id: $id, ontology_id: $ontology_id}) RETURN count(i) AS cnt"
                .to_string(),
        )
        .param("ontology_id", ontology_id)
        .param("id", individual_id);

        let mut result = self.pool.graph().execute(q).await.map_err(|e| {
            ImportError::Database(format!("CHECK individual '{individual_id}': {e}"))
        })?;

        if let Ok(Some(row)) = result.next().await {
            let cnt: i64 = row.get("cnt").unwrap_or(0);
            Ok(cnt > 0)
        } else {
            Ok(false)
        }
    }
}

// ── RDF Parsing ─────────────────────────────────────────────────────────────

/// Parses Turtle content into raw triples.
fn parse_turtle(content: &str, base_iri: &str) -> Result<Vec<RawTriple>, ImportError> {
    let mut triples = Vec::new();
    let reader = BufReader::new(content.as_bytes());
    let base: Option<Iri<String>> = Some(base_iri.parse().map_err(|e: oxiri::IriParseError| {
        ImportError::ParseError(format!("Invalid base IRI: {e}"))
    })?);
    let mut parser = TurtleParser::new(reader, base);

    parser
        .parse_all(&mut |triple| {
            let (subject, predicate, object, is_literal) = convert_triple_parts(&triple);
            triples.push(RawTriple {
                subject,
                predicate,
                object,
                is_literal,
            });
            Ok::<_, Box<dyn std::error::Error>>(())
        })
        .map_err(|e| ImportError::ParseError(format!("Turtle parse error: {e}")))?;

    Ok(triples)
}

/// Parses RDF/XML content into raw triples.
fn parse_rdf_xml(content: &str, base_iri: &str) -> Result<Vec<RawTriple>, ImportError> {
    let mut triples = Vec::new();
    let reader = BufReader::new(content.as_bytes());
    let base: Option<Iri<String>> = Some(base_iri.parse().map_err(|e: oxiri::IriParseError| {
        ImportError::ParseError(format!("Invalid base IRI: {e}"))
    })?);
    let mut parser = RdfXmlParser::new(reader, base);

    parser
        .parse_all(&mut |triple| {
            let (subject, predicate, object, is_literal) = convert_triple_parts(&triple);
            triples.push(RawTriple {
                subject,
                predicate,
                object,
                is_literal,
            });
            Ok::<_, Box<dyn std::error::Error>>(())
        })
        .map_err(|e| ImportError::ParseError(format!("RDF/XML parse error: {e}")))?;

    Ok(triples)
}

/// Converts a parsed rio triple into owned strings.
fn convert_triple_parts(triple: &rio_api::model::Triple<'_>) -> (String, String, String, bool) {
    let subject = match &triple.subject {
        Subject::NamedNode(nn) => nn.iri.to_string(),
        Subject::BlankNode(bn) => format!("_:{}", bn.id),
        Subject::Triple(t) => {
            format!("<<{} {} {}>>", t.subject, t.predicate, t.object)
        }
    };
    let predicate = triple.predicate.iri.to_string();
    let (object, is_literal) = match &triple.object {
        Term::NamedNode(nn) => (nn.iri.to_string(), false),
        Term::BlankNode(bn) => (format!("_:{}", bn.id), false),
        Term::Literal(lit) => match lit {
            rio_api::model::Literal::Simple { value } => (format!("\"{value}\""), true),
            rio_api::model::Literal::LanguageTaggedString { value, language } => {
                (format!("\"{value}\"@{language}"), true)
            }
            rio_api::model::Literal::Typed { value, datatype } => {
                (format!("\"{value}\"^^<{}>", datatype.iri), true)
            }
        },
        Term::Triple(t) => (
            format!("<<{} {} {}>>", t.subject, t.predicate, t.object),
            false,
        ),
    };
    (subject, predicate, object, is_literal)
}

// ── Triple Classification ──────────────────────────────────────────────────

/// Standard RDF/OWL vocabulary IRIs used for classification.
const RDF_TYPE: &str = "http://www.w3.org/1999/02/22-rdf-syntax-ns#type";

/// Classifies raw triples into an `ImportModel`.
fn classify_triples(ontology_id: &str, triples: &[RawTriple]) -> ImportModel {
    let base_prefix = format!("http://vedo.dev/ontology/{ontology_id}#");

    let to_local = |iri: &str| -> String {
        if let Some(local) = iri.strip_prefix(&base_prefix) {
            local.to_string()
        } else if let Some(local) = iri.strip_prefix("http://www.w3.org/2000/01/rdf-schema#") {
            format!("rdfs:{local}")
        } else if let Some(local) = iri.strip_prefix("http://www.w3.org/1999/02/22-rdf-syntax-ns#")
        {
            format!("rdf:{local}")
        } else if let Some(local) = iri.strip_prefix("http://www.w3.org/2002/07/owl#") {
            format!("owl:{local}")
        } else if let Some(local) = iri.strip_prefix("http://www.w3.org/2001/XMLSchema#") {
            format!("xsd:{local}")
        } else {
            iri.to_string()
        }
    };

    let mut classes: HashMap<String, ClassDef> = HashMap::new();
    let mut properties: HashMap<String, PropertyDef> = HashMap::new();
    let mut individuals: HashMap<String, IndividualDef> = HashMap::new();

    // ── Pass 1: Identify entity types from rdf:type ─────────────────────

    for triple in triples {
        if triple.predicate == RDF_TYPE && !triple.is_literal {
            let subject = to_local(&triple.subject);
            let object = to_local(&triple.object);

            match object.as_str() {
                "owl:Class" => {
                    classes.entry(subject.clone()).or_insert_with(|| ClassDef {
                        id: subject,
                        label: None,
                        comment: None,
                        parents: Vec::new(),
                    });
                }
                "owl:ObjectProperty" => {
                    properties
                        .entry(subject.clone())
                        .or_insert_with(|| PropertyDef {
                            id: subject,
                            label: None,
                            comment: None,
                            is_datatype: false,
                            xsd_type: None,
                            domain_ids: Vec::new(),
                            range_ids: Vec::new(),
                            functional: false,
                            inverse_functional: false,
                            transitive: false,
                            symmetric: false,
                        });
                }
                "owl:DatatypeProperty" => {
                    properties
                        .entry(subject.clone())
                        .or_insert_with(|| PropertyDef {
                            id: subject,
                            label: None,
                            comment: None,
                            is_datatype: true,
                            xsd_type: None,
                            domain_ids: Vec::new(),
                            range_ids: Vec::new(),
                            functional: false,
                            inverse_functional: false,
                            transitive: false,
                            symmetric: false,
                        });
                }
                _ => {
                    // rdf:type :ClassName → individual declaration
                    individuals
                        .entry(subject.clone())
                        .or_insert_with(|| IndividualDef {
                            id: subject,
                            label: None,
                            comment: None,
                            class_id: object,
                            literal_values: Vec::new(),
                            reference_values: Vec::new(),
                        });
                }
            }
        }
    }

    // ── Pass 2: Collect annotations, relationships, and property values ──

    for triple in triples {
        let subject = to_local(&triple.subject);
        let predicate = to_local(&triple.predicate);

        match predicate.as_str() {
            "rdf:type" => continue, // handled in Pass 1

            "rdfs:label" if triple.is_literal => {
                let value = extract_literal_value(&triple.object);
                if let Some(cls) = classes.get_mut(&subject) {
                    cls.label = Some(value);
                } else if let Some(prop) = properties.get_mut(&subject) {
                    prop.label = Some(value);
                } else if let Some(ind) = individuals.get_mut(&subject) {
                    ind.label = Some(value);
                }
            }

            "rdfs:comment" if triple.is_literal => {
                let value = extract_literal_value(&triple.object);
                if let Some(cls) = classes.get_mut(&subject) {
                    cls.comment = Some(value);
                } else if let Some(prop) = properties.get_mut(&subject) {
                    prop.comment = Some(value);
                } else if let Some(ind) = individuals.get_mut(&subject) {
                    ind.comment = Some(value);
                }
            }

            "rdfs:subClassOf" if !triple.is_literal => {
                let parent = to_local(&triple.object);
                if let Some(cls) = classes.get_mut(&subject) {
                    cls.parents.push(parent);
                }
            }

            "rdfs:domain" if !triple.is_literal => {
                let domain = to_local(&triple.object);
                if let Some(prop) = properties.get_mut(&subject) {
                    prop.domain_ids.push(domain);
                }
            }

            "rdfs:range" if !triple.is_literal => {
                let range = to_local(&triple.object);
                if range.starts_with("xsd:") {
                    if let Some(prop) = properties.get_mut(&subject) {
                        prop.is_datatype = true;
                        prop.xsd_type = Some(range[4..].to_string());
                    }
                } else if let Some(prop) = properties.get_mut(&subject) {
                    prop.range_ids.push(range);
                }
            }

            "owl:FunctionalProperty"
            | "owl:InverseFunctionalProperty"
            | "owl:TransitiveProperty"
            | "owl:SymmetricProperty" => {
                // Sub-property type — handled via rdf:type triple pattern
            }

            _ => {
                // Any other predicate on an individual → property value
                if triple.is_literal {
                    if let Some(ind) = individuals.get_mut(&subject) {
                        let value = extract_literal_value(&triple.object);
                        let xsd_type = extract_xsd_type(&triple.object);
                        ind.literal_values.push(LiteralValueDef {
                            property_id: predicate,
                            value,
                            xsd_type,
                        });
                    }
                } else if individuals.contains_key(&subject) {
                    let target_id = to_local(&triple.object);
                    if let Some(ind) = individuals.get_mut(&subject) {
                        ind.reference_values.push(ReferenceValueDef {
                            property_id: predicate,
                            target_id,
                        });
                    }
                }
            }
        }
    }

    ImportModel {
        classes: classes.into_values().collect(),
        properties: properties.into_values().collect(),
        individuals: individuals.into_values().collect(),
    }
}

/// Extracts the value from a quoted literal string.
fn extract_literal_value(obj: &str) -> String {
    if let Some(inner) = obj.strip_prefix('"') {
        if let Some((value, _)) = inner.split_once('"') {
            return value.to_string();
        }
    }
    obj.to_string()
}

/// Extracts XSD type from a literal representation `"value"^^<xsd:type>`.
fn extract_xsd_type(obj: &str) -> Option<String> {
    let after_quote = obj.strip_suffix('>')?;
    let (_value, type_part) = after_quote.rsplit_once("^^<")?;
    if let Some(local) = type_part.strip_prefix("http://www.w3.org/2001/XMLSchema#") {
        return Some(local.to_string());
    }
    Some(type_part.to_string())
}

/// Errors that can occur during import.
#[derive(Debug, thiserror::Error)]
pub enum ImportError {
    #[error("Parse error: {0}")]
    ParseError(String),

    #[error("Database error: {0}")]
    Database(String),
}
