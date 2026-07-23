//! Individual (`ABox`) CRUD operations against Neo4j.
//!
//! Provides full lifecycle for OWL individuals (instances of classes) including:
//! - Create individual with `INSTANCE_OF` relationship and property values
//! - Read individual with literal and reference property values

#![allow(
    clippy::cast_possible_wrap,
    clippy::struct_excessive_bools,
    clippy::result_large_err,
    clippy::unused_self
)]
//! - Update property values (add, remove, replace)
//! - Delete individual with reference check and cascade support
//! - List individuals by class (paginated) with property-based filtering
//! - Cross-link navigation between linked individuals

use std::sync::Arc;

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    Json,
};
use serde::{Deserialize, Serialize};
use thiserror::Error;
use tracing::{debug, info, warn};

use crate::neo4j::Neo4jPool;
use crate::AppState;

// ── Helpers ──────────────────────────────────────────────────────────────────────

fn repo_from_state(state: &AppState) -> Result<IndividualRepository, IndividualError> {
    match &state.neo4j {
        Some(pool) => Ok(IndividualRepository::new(pool.clone())),
        None => Err(IndividualError::Neo4jNotConfigured),
    }
}

fn default_page() -> u64 {
    0
}
fn default_per_page() -> u64 {
    20
}

// ── Domain Model ─────────────────────────────────────────────────────────────────

/// A literal property value assigned to an individual.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LiteralValue {
    /// The property ID this value is assigned to.
    pub property_id: String,
    /// The property label (human-readable).
    pub property_label: String,
    /// The literal value as string.
    pub value: String,
    /// The XSD type of the value (e.g. "string", "integer", "boolean").
    #[serde(skip_serializing_if = "Option::is_none")]
    pub xsd_type: Option<String>,
    /// Unique ID of the value node (for targeted removal).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub value_id: Option<String>,
}

/// A reference value — an `ObjectProperty` linking this individual to another.
#[derive(Debug, Clone, Serialize)]
pub struct ReferenceValue {
    /// The `ObjectProperty` ID.
    pub property_id: String,
    /// The property label.
    pub property_label: String,
    /// The target individual ID.
    pub target_id: String,
    /// The target individual label.
    pub target_label: String,
    /// Unique ID of the reference edge (for targeted removal).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub edge_id: Option<String>,
}

/// An OWL individual (instance of a class) in the `ABox`.
#[derive(Debug, Clone, Serialize)]
pub struct Individual {
    /// Unique identifier within the ontology.
    pub id: String,
    /// Human-readable label.
    pub label: String,
    /// Optional comment (rdfs:comment).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub comment: Option<String>,
    /// The class this individual is an instance of.
    pub class_id: String,
    /// The class label.
    pub class_label: String,
    /// Literal property values.
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub literal_values: Vec<LiteralValue>,
    /// Reference property values (links to other individuals).
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub reference_values: Vec<ReferenceValue>,
}

/// Request body for creating an individual.
#[derive(Debug, Deserialize)]
pub struct CreateIndividualRequest {
    /// Optional ID. Auto-generated as UUID if not provided.
    #[serde(default)]
    pub id: Option<String>,
    /// Human-readable label.
    pub label: String,
    /// Optional comment.
    #[serde(default)]
    pub comment: Option<String>,
    /// The class ID this individual instantiates.
    pub class_id: String,
    /// Initial literal property values.
    #[serde(default)]
    pub literal_values: Vec<LiteralValueInput>,
    /// Initial reference property values.
    #[serde(default)]
    pub reference_values: Vec<ReferenceValueInput>,
}

/// Input for a literal property value.
#[derive(Debug, Deserialize)]
pub struct LiteralValueInput {
    pub property_id: String,
    pub value: String,
    #[serde(default)]
    pub xsd_type: Option<String>,
}

/// Input for a reference property value.
#[derive(Debug, Deserialize)]
pub struct ReferenceValueInput {
    pub property_id: String,
    pub target_id: String,
}

/// Request body for updating an individual's property values.
#[derive(Debug, Deserialize)]
pub struct UpdateIndividualRequest {
    /// Updated label.
    #[serde(default)]
    pub label: Option<String>,
    /// Updated comment.
    #[serde(default)]
    pub comment: Option<String>,
    /// Literal values to add.
    #[serde(default)]
    pub add_literals: Vec<LiteralValueInput>,
    /// Reference values to add.
    #[serde(default)]
    pub add_references: Vec<ReferenceValueInput>,
    /// Value IDs of literal values to remove.
    #[serde(default)]
    pub remove_literal_ids: Vec<String>,
    /// Edge IDs of reference values to remove.
    #[serde(default)]
    pub remove_reference_ids: Vec<String>,
}

/// Summary view of an individual for list responses.
#[derive(Debug, Clone, Serialize)]
pub struct IndividualSummary {
    pub id: String,
    pub label: String,
    pub class_id: String,
    pub class_label: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub comment: Option<String>,
}

/// Response for a list operation.
#[derive(Debug, Serialize)]
pub struct PaginatedResponse<T: Serialize> {
    pub items: Vec<T>,
    pub total: u64,
    pub page: u64,
    pub per_page: u64,
}

/// Response for a delete operation.
#[derive(Debug, Serialize)]
pub struct DeleteResponse {
    pub deleted: bool,
    pub individual_id: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub warning: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub dependent_count: Option<u64>,
}

/// Query parameters for listing individuals.
#[derive(Debug, Deserialize)]
pub struct ListIndividualsParams {
    /// Filter by class ID.
    #[serde(default)]
    pub class_id: String,
    /// Fulltext search across label and ID.
    #[serde(default)]
    pub q: String,
    /// Filter by literal property value: `property_id:operator:value`.
    #[serde(default)]
    pub property_filter: String,
    /// Zero-based page offset (default: 0).
    #[serde(default = "default_page")]
    pub page: u64,
    /// Items per page (default: 20, max: 100).
    #[serde(default = "default_per_page")]
    pub per_page: u64,
}

// ── Error Handling ───────────────────────────────────────────────────────────────

/// Errors that can occur during individual operations.
#[derive(Debug, Error)]
pub enum IndividualError {
    #[error("Individual not found: {0}")]
    NotFound(String),

    #[error("Individual already exists: {0}")]
    AlreadyExists(String),

    #[error("Class not found: {0}")]
    ClassNotFound(String),

    #[error("Property not found: {0}")]
    PropertyNotFound(String),

    #[error("Target individual not found: {0}")]
    TargetNotFound(String),

    #[error("Individual has {0} incoming references; delete with cascade=true to force")]
    HasIncomingReferences(u64),

    #[error("Cannot cascade delete: {0}")]
    CascadeError(String),

    #[error("Database error: {0}")]
    Database(String),

    #[error("Neo4j database is not configured")]
    Neo4jNotConfigured,

    #[error("Invalid filter format: {0} (expected property_id:operator:value)")]
    InvalidFilter(String),
}

impl axum::response::IntoResponse for IndividualError {
    fn into_response(self) -> axum::response::Response {
        let (status, code) = match &self {
            IndividualError::NotFound(_) => (StatusCode::NOT_FOUND, "ONT-INDIVIDUAL-NOT-FOUND"),
            IndividualError::AlreadyExists(_) => {
                (StatusCode::CONFLICT, "ONT-INDIVIDUAL-ALREADY-EXISTS")
            }
            IndividualError::ClassNotFound(_) => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-INDIVIDUAL-CLASS-NOT-FOUND",
            ),
            IndividualError::PropertyNotFound(_) => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-INDIVIDUAL-PROPERTY-NOT-FOUND",
            ),
            IndividualError::TargetNotFound(_) => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-INDIVIDUAL-TARGET-NOT-FOUND",
            ),
            IndividualError::HasIncomingReferences(_) => {
                (StatusCode::CONFLICT, "ONT-INDIVIDUAL-HAS-REFERENCES")
            }
            IndividualError::CascadeError(_) => (
                StatusCode::INTERNAL_SERVER_ERROR,
                "ONT-INDIVIDUAL-CASCADE-ERROR",
            ),
            IndividualError::Database(_) => (
                StatusCode::INTERNAL_SERVER_ERROR,
                "ONT-INDIVIDUAL-DATABASE-ERROR",
            ),
            IndividualError::Neo4jNotConfigured => {
                (StatusCode::SERVICE_UNAVAILABLE, "NEO4J_NOT_CONFIGURED")
            }
            IndividualError::InvalidFilter(_) => {
                (StatusCode::BAD_REQUEST, "ONT-INDIVIDUAL-INVALID-FILTER")
            }
        };
        let detail = match &self {
            IndividualError::Database(msg) => {
                let trace_id = uuid::Uuid::new_v4().to_string();
                tracing::error!(
                    error = %msg,
                    trace_id = %trace_id,
                    code = %code,
                    "Database error [trace_id={trace_id}]",
                );
                format!("Internal database error (trace_id: {trace_id})")
            }
            _ => self.to_string(),
        };
        let body = serde_json::json!({
            "error": code,
            "detail": detail,
        });
        (status, Json(body)).into_response()
    }
}

// ── Repository ───────────────────────────────────────────────────────────────────

/// Repository for individual (`ABox`) CRUD operations against Neo4j.
pub struct IndividualRepository {
    pool: Neo4jPool,
}

impl IndividualRepository {
    pub fn new(pool: Neo4jPool) -> Self {
        Self { pool }
    }

    /// Creates a new individual of the given class with optional property values.
    ///
    /// Graph structure:
    /// ```text
    /// (i:Individual {id, label, comment, ontology_id})
    /// (i)-[:INSTANCE_OF]->(:Class {id: class_id})
    /// (i)-[:HAS_VALUE {value_id}]->(:LiteralValue {value, xsd_type, property_id})
    /// (i)-[:HAS_REF {edge_id}]->(:Individual {id: target_id})
    /// ```
    pub async fn create(
        &self,
        ontology_id: &str,
        req: &CreateIndividualRequest,
    ) -> Result<Individual, IndividualError> {
        debug!(
            ontology_id,
            class_id = %req.class_id, label = %req.label,
            "Creating individual"
        );

        // Verify class exists
        let class_exists = self.class_exists(ontology_id, &req.class_id).await?;
        if !class_exists {
            return Err(IndividualError::ClassNotFound(req.class_id.clone()));
        }

        // Resolve ID
        let individual_id = req
            .id
            .clone()
            .unwrap_or_else(|| uuid::Uuid::new_v4().to_string());

        // Verify no duplicate
        let exists = self.exists(ontology_id, &individual_id).await?;
        if exists {
            return Err(IndividualError::AlreadyExists(individual_id));
        }

        let comment = req.comment.as_deref().unwrap_or("");
        let create_query = "\
            MATCH (c:Class {id: $class_id, ontology_id: $ontology_id})\
            CREATE (i:Individual {id: $id, label: $label, comment: $comment, ontology_id: $ontology_id})\
            CREATE (i)-[:INSTANCE_OF]->(c)\
            RETURN i, c.id AS class_id, c.label AS class_label\
        ";
        let q = neo4rs::Query::new(create_query.to_string())
            .param("ontology_id", ontology_id)
            .param("class_id", req.class_id.as_str())
            .param("id", individual_id.as_str())
            .param("label", req.label.as_str())
            .param("comment", comment);

        let mut result = self.pool.graph().execute(q).await.map_err(|e| {
            warn!(error = %e, ontology_id, class_id = %req.class_id, "Failed to create individual");
            IndividualError::Database(e.to_string())
        })?;

        let individual = if let Ok(Some(row)) = result.next().await {
            let class_id: String = row.get("class_id").unwrap_or_default();
            let class_label: String = row.get("class_label").unwrap_or_default();
            if let Ok(node) = row.get::<neo4rs::Node>("i") {
                let mut ind = self.build_individual_from_node(&node);
                ind.class_id = class_id;
                ind.class_label = class_label;
                ind
            } else {
                return Err(IndividualError::Database(
                    "Failed to parse created individual".to_string(),
                ));
            }
        } else {
            return Err(IndividualError::Database(
                "Failed to retrieve created individual".to_string(),
            ));
        };

        // Create literal values individually
        let literal_count = req.literal_values.len();
        for lv_input in &req.literal_values {
            let val_id = uuid::Uuid::new_v4().to_string();
            let lv_query = "\
                MATCH (i:Individual {id: $id, ontology_id: $ontology_id})\
                CREATE (i)-[:HAS_VALUE {value_id: $value_id}]->\
                    (:LiteralValue {id: $value_id, property_id: $prop_id, value: $value, xsd_type: $xsd_type})\
            ";
            let lq = neo4rs::Query::new(lv_query.to_string())
                .param("id", individual_id.as_str())
                .param("ontology_id", ontology_id)
                .param("value_id", val_id.as_str())
                .param("prop_id", lv_input.property_id.as_str())
                .param("value", lv_input.value.as_str())
                .param("xsd_type", lv_input.xsd_type.as_deref().unwrap_or("string"));
            match self.pool.graph().execute(lq).await {
                Ok(mut stream) => while let Ok(Some(_)) = stream.next().await {},
                Err(e) => {
                    warn!(error = %e, "Failed to create literal value");
                    return Err(IndividualError::Database(e.to_string()));
                }
            }
        }

        // Create reference values individually
        let ref_count = req.reference_values.len();
        for rv_input in &req.reference_values {
            let edge_id = uuid::Uuid::new_v4().to_string();
            let ref_query = "\
                MATCH (i:Individual {id: $id, ontology_id: $ontology_id})\
                MATCH (target:Individual {id: $target_id, ontology_id: $ontology_id})\
                CREATE (i)-[:HAS_REF {edge_id: $edge_id, property_id: $prop_id}]->(target)\
            ";
            let rq = neo4rs::Query::new(ref_query.to_string())
                .param("id", individual_id.as_str())
                .param("ontology_id", ontology_id)
                .param("target_id", rv_input.target_id.as_str())
                .param("edge_id", edge_id.as_str())
                .param("prop_id", rv_input.property_id.as_str());
            match self.pool.graph().execute(rq).await {
                Ok(mut stream) => while let Ok(Some(_)) = stream.next().await {},
                Err(e) => {
                    warn!(error = %e, "Failed to create reference value");
                    return Err(IndividualError::Database(e.to_string()));
                }
            }
        }

        // Return the created individual with attached values
        let mut result = individual;
        result.literal_values = self.get_literal_values(ontology_id, &individual_id).await?;
        result.reference_values = self
            .get_reference_values(ontology_id, &individual_id)
            .await?;

        info!(
            ontology_id,
            individual_id = %result.id,
            class_id = %result.class_id,
            literal_count,
            ref_count,
            "Individual created"
        );

        Ok(result)
    }

    /// Retrieves a single individual with all property values.
    pub async fn get(
        &self,
        ontology_id: &str,
        individual_id: &str,
    ) -> Result<Individual, IndividualError> {
        debug!(ontology_id, %individual_id, "Getting individual");

        let query = "\
            MATCH (i:Individual {id: $id, ontology_id: $ontology_id})\
            OPTIONAL MATCH (i)-[:INSTANCE_OF]->(c:Class)\
            RETURN i, c.id AS class_id, c.label AS class_label\
        ";

        let q = neo4rs::Query::new(query.to_string())
            .param("id", individual_id.to_owned())
            .param("ontology_id", ontology_id.to_owned());

        let mut result = self.pool.graph().execute(q).await.map_err(|e| {
            warn!(error = %e, ontology_id, %individual_id, "Failed to get individual");
            IndividualError::Database(e.to_string())
        })?;

        if let Ok(Some(row)) = result.next().await {
            if let Ok(node) = row.get::<neo4rs::Node>("i") {
                let mut individual = self.build_individual_from_node(&node);

                // Override class info from OPTIONAL MATCH
                individual.class_id = row
                    .get::<Option<String>>("class_id")
                    .ok()
                    .flatten()
                    .unwrap_or_default();
                individual.class_label = row
                    .get::<Option<String>>("class_label")
                    .ok()
                    .flatten()
                    .unwrap_or_default();

                // Attach property values
                individual.literal_values =
                    self.get_literal_values(ontology_id, individual_id).await?;
                individual.reference_values = self
                    .get_reference_values(ontology_id, individual_id)
                    .await?;

                info!(
                    ontology_id,
                    %individual_id,
                    literal_count = individual.literal_values.len(),
                    ref_count = individual.reference_values.len(),
                    "Individual retrieved"
                );

                Ok(individual)
            } else {
                Err(IndividualError::Database(
                    "Failed to parse individual node".to_string(),
                ))
            }
        } else {
            Err(IndividualError::NotFound(individual_id.to_string()))
        }
    }

    /// Updates an individual's label, comment, and/or property values.
    pub async fn update(
        &self,
        ontology_id: &str,
        individual_id: &str,
        req: &UpdateIndividualRequest,
    ) -> Result<Individual, IndividualError> {
        debug!(ontology_id, %individual_id, "Updating individual");

        // Verify individual exists
        let _ = self.get(ontology_id, individual_id).await?;

        // Update label/comment if provided
        if req.label.is_some() || req.comment.is_some() {
            let mut set_clauses = Vec::new();
            if req.label.is_some() {
                set_clauses.push("i.label = $label");
            }
            if req.comment.is_some() {
                set_clauses.push("i.comment = $comment");
            }

            let update_query_str = format!(
                "MATCH (i:Individual {{id: $id, ontology_id: $ontology_id}}) SET {}",
                set_clauses.join(", ")
            );

            let mut q = neo4rs::Query::new(update_query_str)
                .param("id", individual_id.to_owned())
                .param("ontology_id", ontology_id.to_owned());
            if let Some(label) = &req.label {
                q = q.param("label", label.as_str());
            }
            if let Some(comment) = &req.comment {
                q = q.param("comment", comment.as_str());
            }

            match self.pool.graph().execute(q).await {
                Ok(mut stream) => while let Ok(Some(_)) = stream.next().await {},
                Err(e) => {
                    warn!(error = %e, ontology_id, %individual_id, "Failed to update individual");
                    return Err(IndividualError::Database(e.to_string()));
                }
            }
        }

        // Remove literal values by value_id
        for vid in &req.remove_literal_ids {
            let del_query = "\
                MATCH (i:Individual {id: $id, ontology_id: $ontology_id})\
                MATCH (i)-[r:HAS_VALUE]->(lv:LiteralValue {id: $value_id})\
                DELETE r, lv\
            ";
            let q = neo4rs::Query::new(del_query.to_string())
                .param("id", individual_id.to_owned())
                .param("ontology_id", ontology_id.to_owned())
                .param("value_id", vid.as_str());

            match self.pool.graph().execute(q).await {
                Ok(mut stream) => while let Ok(Some(_)) = stream.next().await {},
                Err(e) => {
                    warn!(error = %e, ontology_id, %individual_id, "Failed to remove literal value");
                    return Err(IndividualError::Database(e.to_string()));
                }
            }
        }

        // Remove reference values by edge_id
        for eid in &req.remove_reference_ids {
            let del_query = "\
                MATCH (i:Individual {id: $id, ontology_id: $ontology_id})\
                MATCH (i)-[r:HAS_REF {edge_id: $edge_id}]->()\
                DELETE r\
            ";
            let q = neo4rs::Query::new(del_query.to_string())
                .param("id", individual_id.to_owned())
                .param("ontology_id", ontology_id.to_owned())
                .param("edge_id", eid.as_str());

            match self.pool.graph().execute(q).await {
                Ok(mut stream) => while let Ok(Some(_)) = stream.next().await {},
                Err(e) => {
                    warn!(error = %e, ontology_id, %individual_id, "Failed to remove reference");
                    return Err(IndividualError::Database(e.to_string()));
                }
            }
        }

        // Add literal values
        for lv_input in &req.add_literals {
            let val_id = uuid::Uuid::new_v4().to_string();
            let add_query = "\
                MATCH (i:Individual {id: $id, ontology_id: $ontology_id})\
                CREATE (i)-[:HAS_VALUE {value_id: $value_id}]->\
                    (:LiteralValue {id: $value_id, property_id: $prop_id, value: $value, xsd_type: $xsd_type})\
            ";
            let q = neo4rs::Query::new(add_query.to_string())
                .param("id", individual_id.to_owned())
                .param("ontology_id", ontology_id.to_owned())
                .param("value_id", val_id.as_str())
                .param("prop_id", lv_input.property_id.as_str())
                .param("value", lv_input.value.as_str())
                .param("xsd_type", lv_input.xsd_type.as_deref().unwrap_or("string"));

            match self.pool.graph().execute(q).await {
                Ok(mut stream) => while let Ok(Some(_)) = stream.next().await {},
                Err(e) => {
                    warn!(error = %e, ontology_id, %individual_id, "Failed to add literal value");
                    return Err(IndividualError::Database(e.to_string()));
                }
            }
        }

        // Add reference values
        for rv_input in &req.add_references {
            // Verify target exists
            let target_exists = self.exists(ontology_id, &rv_input.target_id).await?;
            if !target_exists {
                return Err(IndividualError::TargetNotFound(rv_input.target_id.clone()));
            }

            let edge_id = uuid::Uuid::new_v4().to_string();
            let add_query = "\
                MATCH (i:Individual {id: $id, ontology_id: $ontology_id})\
                MATCH (target:Individual {id: $target_id, ontology_id: $ontology_id})\
                CREATE (i)-[:HAS_REF {edge_id: $edge_id, property_id: $prop_id}]->(target)\
            ";
            let q = neo4rs::Query::new(add_query.to_string())
                .param("id", individual_id.to_owned())
                .param("ontology_id", ontology_id.to_owned())
                .param("target_id", rv_input.target_id.as_str())
                .param("edge_id", edge_id.as_str())
                .param("prop_id", rv_input.property_id.as_str());

            match self.pool.graph().execute(q).await {
                Ok(mut stream) => while let Ok(Some(_)) = stream.next().await {},
                Err(e) => {
                    warn!(error = %e, ontology_id, %individual_id, "Failed to add reference");
                    return Err(IndividualError::Database(e.to_string()));
                }
            }
        }

        info!(ontology_id, %individual_id, "Individual updated");

        // Return full individual
        self.get(ontology_id, individual_id).await
    }

    /// Deletes an individual. Checks for incoming references before deletion.
    pub async fn delete(
        &self,
        ontology_id: &str,
        individual_id: &str,
        cascade: bool,
    ) -> Result<DeleteResponse, IndividualError> {
        debug!(ontology_id, %individual_id, cascade, "Deleting individual");

        // Verify individual exists
        let exists = self.exists(ontology_id, individual_id).await?;
        if !exists {
            return Err(IndividualError::NotFound(individual_id.to_string()));
        }

        // Check incoming references
        let incoming_count = self
            .count_incoming_references(ontology_id, individual_id)
            .await?;

        if incoming_count > 0 && !cascade {
            return Err(IndividualError::HasIncomingReferences(incoming_count));
        }

        if cascade && incoming_count > 0 {
            warn!(
                ontology_id,
                %individual_id,
                incoming_count,
                "Cascade deleting individual with incoming references"
            );
            // Remove all incoming HAS_REF edges
            let cascade_query = "\
                MATCH (source:Individual)-[r:HAS_REF]->(target:Individual {id: $id, ontology_id: $ontology_id})\
                DELETE r\
            ";
            let q = neo4rs::Query::new(cascade_query.to_string())
                .param("id", individual_id.to_owned())
                .param("ontology_id", ontology_id.to_owned());

            match self.pool.graph().execute(q).await {
                Ok(mut stream) => while let Ok(Some(_)) = stream.next().await {},
                Err(e) => {
                    warn!(error = %e, ontology_id, %individual_id, "Cascade delete failed");
                    return Err(IndividualError::Database(e.to_string()));
                }
            }
        }

        // Delete individual node and all its incident edges and literal value nodes
        let delete_query = "\
            MATCH (i:Individual {id: $id, ontology_id: $ontology_id})\
            OPTIONAL MATCH (i)-[r:HAS_VALUE]->(lv:LiteralValue)\
            OPTIONAL MATCH (i)-[ref:HAS_REF]->()\
            OPTIONAL MATCH ()-[incoming:HAS_REF]->(i)\
            DELETE r, ref, incoming, lv, i\
        ";
        let q = neo4rs::Query::new(delete_query.to_string())
            .param("id", individual_id.to_owned())
            .param("ontology_id", ontology_id.to_owned());

        match self.pool.graph().execute(q).await {
            Ok(mut stream) => while let Ok(Some(_)) = stream.next().await {},
            Err(e) => {
                warn!(error = %e, ontology_id, %individual_id, "Failed to delete individual");
                return Err(IndividualError::Database(e.to_string()));
            }
        }

        let response = if incoming_count > 0 {
            DeleteResponse {
                deleted: true,
                individual_id: individual_id.to_string(),
                warning: Some(format!(
                    "Deleted with cascade — {incoming_count} incoming references removed"
                )),
                dependent_count: Some(incoming_count),
            }
        } else {
            DeleteResponse {
                deleted: true,
                individual_id: individual_id.to_string(),
                warning: None,
                dependent_count: None,
            }
        };

        info!(
            ontology_id,
            %individual_id,
            deleted = response.deleted,
            "Individual deleted"
        );

        Ok(response)
    }

    /// Lists individuals, optionally filtered by class, text search, and property values.
    pub async fn list(
        &self,
        ontology_id: &str,
        params: &ListIndividualsParams,
    ) -> Result<PaginatedResponse<IndividualSummary>, IndividualError> {
        debug!(
            ontology_id,
            class_id = %params.class_id,
            q = %params.q,
            page = params.page,
            per_page = params.per_page,
            "Listing individuals"
        );

        let offset = params.page * params.per_page;
        let limit = params.per_page.min(100);

        // Build dynamic WHERE clauses
        let mut conditions = vec!["i.ontology_id = $ontology_id".to_string()];

        if !params.class_id.is_empty() {
            conditions.push("c.id = $class_id".to_string());
        }

        if !params.q.is_empty() {
            conditions.push(
                "(toLower(i.label) CONTAINS toLower($q) OR toLower(i.id) CONTAINS toLower($q))"
                    .to_string(),
            );
        }

        let has_property_filter = !params.property_filter.is_empty();
        if has_property_filter {
            let parts: Vec<&str> = params.property_filter.splitn(3, ':').collect();
            if parts.len() == 3 {
                conditions.push(
                    "EXISTS { MATCH (i)-[:HAS_VALUE]->(lv:LiteralValue) \
                     WHERE lv.property_id = $f_prop AND lv.value = $f_val }"
                        .to_string(),
                );
            } else {
                return Err(IndividualError::InvalidFilter(
                    params.property_filter.clone(),
                ));
            }
        }

        let where_clause = conditions.join(" AND ");

        // Count query
        let count_query_str = format!(
            "MATCH (i:Individual)-[:INSTANCE_OF]->(c:Class) WHERE {where_clause} RETURN count(i) AS total"
        );
        let mut cq =
            neo4rs::Query::new(count_query_str).param("ontology_id", ontology_id.to_owned());
        if !params.class_id.is_empty() {
            cq = cq.param("class_id", params.class_id.as_str());
        }
        if !params.q.is_empty() {
            cq = cq.param("q", params.q.as_str());
        }
        if has_property_filter {
            let parts: Vec<&str> = params.property_filter.splitn(3, ':').collect();
            cq = cq.param("f_prop", parts[0]);
            cq = cq.param("f_val", parts[2]);
        }

        let mut count_result = self.pool.graph().execute(cq).await.map_err(|e| {
            warn!(error = %e, ontology_id, "Failed to count individuals");
            IndividualError::Database(e.to_string())
        })?;

        let total: u64 = if let Ok(Some(row)) = count_result.next().await {
            row.get("total").unwrap_or(0)
        } else {
            0
        };

        // Fetch query
        let fetch_query_str = format!(
            "MATCH (i:Individual)-[:INSTANCE_OF]->(c:Class) WHERE {where_clause} \
             RETURN i.id AS id, i.label AS label, i.comment AS comment, \
             c.id AS class_id, c.label AS class_label \
             ORDER BY i.label SKIP $skip LIMIT $limit"
        );
        let mut fq = neo4rs::Query::new(fetch_query_str)
            .param("ontology_id", ontology_id.to_owned())
            .param("skip", offset as i64)
            .param("limit", limit as i64);
        if !params.class_id.is_empty() {
            fq = fq.param("class_id", params.class_id.as_str());
        }
        if !params.q.is_empty() {
            fq = fq.param("q", params.q.as_str());
        }
        if has_property_filter {
            let parts: Vec<&str> = params.property_filter.splitn(3, ':').collect();
            fq = fq.param("f_prop", parts[0]);
            fq = fq.param("f_val", parts[2]);
        }

        let mut result = self.pool.graph().execute(fq).await.map_err(|e| {
            warn!(error = %e, ontology_id, "Failed to list individuals");
            IndividualError::Database(e.to_string())
        })?;

        let mut items = Vec::with_capacity(limit as usize);
        while let Ok(Some(row)) = result.next().await {
            if let Ok(id) = row.get::<String>("id") {
                items.push(IndividualSummary {
                    id,
                    label: row.get("label").unwrap_or_default(),
                    comment: row
                        .get::<Option<String>>("comment")
                        .ok()
                        .flatten()
                        .filter(|c| !c.is_empty()),
                    class_id: row.get("class_id").unwrap_or_default(),
                    class_label: row.get("class_label").unwrap_or_default(),
                });
            }
        }

        info!(
            ontology_id,
            total,
            returned = items.len(),
            "List individuals completed"
        );

        Ok(PaginatedResponse {
            items,
            total,
            page: params.page,
            per_page: limit,
        })
    }

    /// Retrieves literal property values for an individual.
    async fn get_literal_values(
        &self,
        ontology_id: &str,
        individual_id: &str,
    ) -> Result<Vec<LiteralValue>, IndividualError> {
        let query = "\
            MATCH (i:Individual {id: $id, ontology_id: $ontology_id})\
            MATCH (i)-[r:HAS_VALUE]->(lv:LiteralValue)\
            OPTIONAL MATCH (p:Property {id: lv.property_id})\
            RETURN lv.id AS value_id, lv.property_id AS property_id, \
                   p.label AS property_label, lv.value AS value, lv.xsd_type AS xsd_type\
        ";
        let q = neo4rs::Query::new(query.to_string())
            .param("id", individual_id.to_owned())
            .param("ontology_id", ontology_id.to_owned());

        let mut result = self.pool.graph().execute(q).await.map_err(|e| {
            warn!(error = %e, ontology_id, %individual_id, "Failed to get literal values");
            IndividualError::Database(e.to_string())
        })?;

        let mut values = Vec::new();
        while let Ok(Some(row)) = result.next().await {
            values.push(LiteralValue {
                property_id: row.get("property_id").unwrap_or_default(),
                property_label: row
                    .get::<Option<String>>("property_label")
                    .ok()
                    .flatten()
                    .unwrap_or_else(|| "unknown".to_string()),
                value: row.get("value").unwrap_or_default(),
                xsd_type: row.get::<Option<String>>("xsd_type").ok().flatten(),
                value_id: row.get::<Option<String>>("value_id").ok().flatten(),
            });
        }

        Ok(values)
    }

    /// Retrieves reference property values for an individual.
    async fn get_reference_values(
        &self,
        ontology_id: &str,
        individual_id: &str,
    ) -> Result<Vec<ReferenceValue>, IndividualError> {
        let query = "\
            MATCH (i:Individual {id: $id, ontology_id: $ontology_id})\
            MATCH (i)-[r:HAS_REF]->(target:Individual)\
            OPTIONAL MATCH (p:Property {id: r.property_id})\
            RETURN r.edge_id AS edge_id, r.property_id AS property_id, \
                   p.label AS property_label, target.id AS target_id, target.label AS target_label\
        ";
        let q = neo4rs::Query::new(query.to_string())
            .param("id", individual_id.to_owned())
            .param("ontology_id", ontology_id.to_owned());

        let mut result = self.pool.graph().execute(q).await.map_err(|e| {
            warn!(error = %e, ontology_id, %individual_id, "Failed to get reference values");
            IndividualError::Database(e.to_string())
        })?;

        let mut values = Vec::new();
        while let Ok(Some(row)) = result.next().await {
            values.push(ReferenceValue {
                property_id: row.get("property_id").unwrap_or_default(),
                property_label: row
                    .get::<Option<String>>("property_label")
                    .ok()
                    .flatten()
                    .unwrap_or_else(|| "unknown".to_string()),
                target_id: row.get("target_id").unwrap_or_default(),
                target_label: row.get("target_label").unwrap_or_default(),
                edge_id: row.get::<Option<String>>("edge_id").ok().flatten(),
            });
        }

        Ok(values)
    }

    /// Counts incoming `HAS_REF` references to an individual.
    async fn count_incoming_references(
        &self,
        ontology_id: &str,
        individual_id: &str,
    ) -> Result<u64, IndividualError> {
        let query = "\
            MATCH ()-[r:HAS_REF]->(target:Individual {id: $id, ontology_id: $ontology_id})\
            RETURN count(r) AS count\
        ";
        let q = neo4rs::Query::new(query.to_string())
            .param("id", individual_id.to_owned())
            .param("ontology_id", ontology_id.to_owned());

        let mut result = self.pool.graph().execute(q).await.map_err(|e| {
            warn!(error = %e, ontology_id, %individual_id, "Failed to count incoming references");
            IndividualError::Database(e.to_string())
        })?;

        if let Ok(Some(row)) = result.next().await {
            Ok(row.get("count").unwrap_or(0))
        } else {
            Ok(0)
        }
    }

    /// Checks if an individual with the given ID exists in the ontology.
    async fn exists(
        &self,
        ontology_id: &str,
        individual_id: &str,
    ) -> Result<bool, IndividualError> {
        let query = "\
            MATCH (i:Individual {id: $id, ontology_id: $ontology_id}) RETURN i\
        ";
        let q = neo4rs::Query::new(query.to_string())
            .param("id", individual_id.to_owned())
            .param("ontology_id", ontology_id.to_owned());

        let mut result = self
            .pool
            .graph()
            .execute(q)
            .await
            .map_err(|e| IndividualError::Database(e.to_string()))?;

        Ok(result
            .next()
            .await
            .map_err(|e| IndividualError::Database(e.to_string()))?
            .is_some())
    }

    /// Checks if a class exists in the ontology.
    async fn class_exists(
        &self,
        ontology_id: &str,
        class_id: &str,
    ) -> Result<bool, IndividualError> {
        let query = "\
            MATCH (c:Class {id: $id, ontology_id: $ontology_id}) RETURN c\
        ";
        let q = neo4rs::Query::new(query.to_string())
            .param("id", class_id.to_owned())
            .param("ontology_id", ontology_id.to_owned());

        let mut result = self
            .pool
            .graph()
            .execute(q)
            .await
            .map_err(|e| IndividualError::Database(e.to_string()))?;

        Ok(result
            .next()
            .await
            .map_err(|e| IndividualError::Database(e.to_string()))?
            .is_some())
    }

    /// Builds an Individual struct from a Neo4j node.
    fn build_individual_from_node(&self, node: &neo4rs::Node) -> Individual {
        Individual {
            id: node.get::<String>("id").unwrap_or_default(),
            label: node.get::<String>("label").unwrap_or_default(),
            comment: node.get::<Option<String>>("comment").ok().flatten(),
            class_id: String::new(),    // populated by caller
            class_label: String::new(), // populated by caller
            literal_values: Vec::new(),
            reference_values: Vec::new(),
        }
    }
}

// ── Handlers ─────────────────────────────────────────────────────────────────────

/// POST `/api/v1/ontologies/{ontology_id}/individuals`
pub async fn create_individual_handler(
    State(state): State<Arc<AppState>>,
    Path(ontology_id): Path<String>,
    Json(req): Json<CreateIndividualRequest>,
) -> Result<(StatusCode, Json<Individual>), IndividualError> {
    debug!(ontology_id, label = %req.label, class_id = %req.class_id, "POST create individual handler invoked");
    let repo = repo_from_state(&state)?;
    let individual = repo.create(&ontology_id, &req).await?;
    info!(ontology_id, individual_id = %individual.id, "POST create individual handler completed");
    Ok((StatusCode::CREATED, Json(individual)))
}

/// GET `/api/v1/ontologies/{ontology_id}/individuals`
pub async fn list_individuals_handler(
    State(state): State<Arc<AppState>>,
    Path(ontology_id): Path<String>,
    Query(params): Query<ListIndividualsParams>,
) -> Result<Json<PaginatedResponse<IndividualSummary>>, IndividualError> {
    debug!(ontology_id, "GET list individuals handler invoked");
    let repo = repo_from_state(&state)?;
    let result = repo.list(&ontology_id, &params).await?;
    Ok(Json(result))
}

/// GET `/api/v1/ontologies/{ontology_id}/individuals/{individual_id}`
pub async fn get_individual_handler(
    State(state): State<Arc<AppState>>,
    Path((ontology_id, individual_id)): Path<(String, String)>,
) -> Result<Json<Individual>, IndividualError> {
    debug!(ontology_id, %individual_id, "GET individual handler invoked");
    let repo = repo_from_state(&state)?;
    let individual = repo.get(&ontology_id, &individual_id).await?;
    Ok(Json(individual))
}

/// PUT `/api/v1/ontologies/{ontology_id}/individuals/{individual_id}`
pub async fn update_individual_handler(
    State(state): State<Arc<AppState>>,
    Path((ontology_id, individual_id)): Path<(String, String)>,
    Json(req): Json<UpdateIndividualRequest>,
) -> Result<Json<Individual>, IndividualError> {
    debug!(ontology_id, %individual_id, "PUT update individual handler invoked");
    let repo = repo_from_state(&state)?;
    let individual = repo.update(&ontology_id, &individual_id, &req).await?;
    info!(ontology_id, individual_id = %individual.id, "PUT update individual handler completed");
    Ok(Json(individual))
}

/// DELETE `/api/v1/ontologies/{ontology_id}/individuals/{individual_id}`
pub async fn delete_individual_handler(
    State(state): State<Arc<AppState>>,
    Path((ontology_id, individual_id)): Path<(String, String)>,
    Query(params): Query<DeleteIndividualParams>,
) -> Result<Json<DeleteResponse>, IndividualError> {
    debug!(ontology_id, %individual_id, cascade = params.cascade, "DELETE individual handler invoked");
    let repo = repo_from_state(&state)?;
    let result = repo
        .delete(&ontology_id, &individual_id, params.cascade)
        .await?;
    info!(ontology_id, individual_id = %result.individual_id, "DELETE individual handler completed");
    Ok(Json(result))
}

/// Query parameters for individual deletion.
#[derive(Debug, Deserialize)]
pub struct DeleteIndividualParams {
    #[serde(default)]
    pub cascade: bool,
}

// ── Tests ────────────────────────────────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;
    use crate::clients::auth_client::AuthClient;
    use crate::AppState;
    use axum::{http::StatusCode, response::IntoResponse};

    fn test_app_state() -> Arc<AppState> {
        Arc::new(AppState {
            neo4j: None,
            auth_client: AuthClient::new(None),
        })
    }

    // ── Domain Model Serialization ──────────────────────────────────────────

    #[test]
    fn test_literal_value_serialization() {
        let lv = LiteralValue {
            property_id: "hasAge".to_string(),
            property_label: "has age".to_string(),
            value: "42".to_string(),
            xsd_type: Some("integer".to_string()),
            value_id: Some("lv_001".to_string()),
        };
        let json = serde_json::to_value(&lv).unwrap();
        assert_eq!(json["property_id"], "hasAge");
        assert_eq!(json["value"], "42");
        assert_eq!(json["xsd_type"], "integer");
    }

    #[test]
    fn test_reference_value_serialization() {
        let rv = ReferenceValue {
            property_id: "knows".to_string(),
            property_label: "knows".to_string(),
            target_id: "person_2".to_string(),
            target_label: "Person 2".to_string(),
            edge_id: Some("ref_001".to_string()),
        };
        let json = serde_json::to_value(&rv).unwrap();
        assert_eq!(json["target_id"], "person_2");
        assert_eq!(json["property_label"], "knows");
    }

    #[test]
    fn test_individual_serialization() {
        let individual = Individual {
            id: "person_1".to_string(),
            label: "Person 1".to_string(),
            comment: Some("A test person".to_string()),
            class_id: "Person".to_string(),
            class_label: "Person".to_string(),
            literal_values: vec![LiteralValue {
                property_id: "hasAge".to_string(),
                property_label: "has age".to_string(),
                value: "42".to_string(),
                xsd_type: Some("integer".to_string()),
                value_id: None,
            }],
            reference_values: vec![ReferenceValue {
                property_id: "knows".to_string(),
                property_label: "knows".to_string(),
                target_id: "person_2".to_string(),
                target_label: "Person 2".to_string(),
                edge_id: None,
            }],
        };
        let json = serde_json::to_value(&individual).unwrap();
        assert_eq!(json["id"], "person_1");
        assert_eq!(json["class_id"], "Person");
        assert_eq!(json["literal_values"][0]["value"], "42");
        assert_eq!(json["reference_values"][0]["target_id"], "person_2");
    }

    #[test]
    fn test_individual_no_comment() {
        let individual = Individual {
            id: "person_1".to_string(),
            label: "Person 1".to_string(),
            comment: None,
            class_id: "Person".to_string(),
            class_label: "Person".to_string(),
            literal_values: vec![],
            reference_values: vec![],
        };
        let json = serde_json::to_value(&individual).unwrap();
        assert!(json.get("comment").is_none());
    }

    #[test]
    fn test_create_individual_request_deserialization() {
        let body = serde_json::json!({
            "id": "person_1",
            "label": "Person 1",
            "class_id": "Person",
            "literal_values": [
                {"property_id": "hasAge", "value": "42", "xsd_type": "integer"}
            ],
            "reference_values": [
                {"property_id": "knows", "target_id": "person_2"}
            ]
        });
        let req: CreateIndividualRequest = serde_json::from_value(body).unwrap();
        assert_eq!(req.id, Some("person_1".to_string()));
        assert_eq!(req.label, "Person 1");
        assert_eq!(req.class_id, "Person");
        assert_eq!(req.literal_values.len(), 1);
        assert_eq!(req.reference_values.len(), 1);
    }

    #[test]
    fn test_create_individual_request_minimal() {
        let body = serde_json::json!({
            "label": "Person 1",
            "class_id": "Person"
        });
        let req: CreateIndividualRequest = serde_json::from_value(body).unwrap();
        assert!(req.id.is_none());
        assert_eq!(req.label, "Person 1");
        assert!(req.literal_values.is_empty());
    }

    #[test]
    fn test_update_individual_request_deserialization() {
        let body = serde_json::json!({
            "label": "Updated Person",
            "add_literals": [
                {"property_id": "hasName", "value": "Alice"}
            ],
            "remove_literal_ids": ["lv_001"]
        });
        let req: UpdateIndividualRequest = serde_json::from_value(body).unwrap();
        assert_eq!(req.label, Some("Updated Person".to_string()));
        assert_eq!(req.add_literals.len(), 1);
        assert_eq!(req.remove_literal_ids.len(), 1);
    }

    #[test]
    fn test_individual_summary_serialization() {
        let summary = IndividualSummary {
            id: "person_1".to_string(),
            label: "Person 1".to_string(),
            class_id: "Person".to_string(),
            class_label: "Person".to_string(),
            comment: None,
        };
        let json = serde_json::to_value(&summary).unwrap();
        assert_eq!(json["id"], "person_1");
        assert_eq!(json["class_id"], "Person");
    }

    #[test]
    fn test_list_individuals_params_defaults() {
        let params: ListIndividualsParams = serde_json::from_value(serde_json::json!({})).unwrap();
        assert_eq!(params.page, 0);
        assert_eq!(params.per_page, 20);
        assert!(params.class_id.is_empty());
        assert!(params.q.is_empty());
    }

    #[test]
    fn test_list_individuals_params_custom() {
        let params: ListIndividualsParams = serde_json::from_value(serde_json::json!({
            "class_id": "Person",
            "q": "Alice",
            "page": 1,
            "per_page": 10
        }))
        .unwrap();
        assert_eq!(params.class_id, "Person");
        assert_eq!(params.q, "Alice");
        assert_eq!(params.page, 1);
        assert_eq!(params.per_page, 10);
    }

    #[test]
    fn test_delete_individual_params_default() {
        let params: DeleteIndividualParams = serde_json::from_value(serde_json::json!({})).unwrap();
        assert!(!params.cascade);
    }

    #[test]
    fn test_delete_individual_params_cascade_true() {
        let params: DeleteIndividualParams =
            serde_json::from_value(serde_json::json!({"cascade": true})).unwrap();
        assert!(params.cascade);
    }

    #[test]
    fn test_paginated_response_serialization() {
        let items = vec![
            IndividualSummary {
                id: "p1".to_string(),
                label: "Person 1".to_string(),
                class_id: "Person".to_string(),
                class_label: "Person".to_string(),
                comment: None,
            },
            IndividualSummary {
                id: "p2".to_string(),
                label: "Person 2".to_string(),
                class_id: "Person".to_string(),
                class_label: "Person".to_string(),
                comment: None,
            },
        ];
        let resp = PaginatedResponse {
            items,
            total: 10,
            page: 0,
            per_page: 20,
        };
        let json = serde_json::to_value(&resp).unwrap();
        assert_eq!(json["total"], 10);
        assert_eq!(json["items"].as_array().unwrap().len(), 2);
    }

    #[test]
    fn test_delete_response_serialization() {
        let resp = DeleteResponse {
            deleted: true,
            individual_id: "person_1".to_string(),
            warning: Some("Has 3 incoming references".to_string()),
            dependent_count: Some(3),
        };
        let json = serde_json::to_value(&resp).unwrap();
        assert_eq!(json["deleted"], true);
        assert_eq!(json["individual_id"], "person_1");
        assert_eq!(json["warning"], "Has 3 incoming references");
    }

    #[test]
    fn test_delete_response_no_warning() {
        let resp = DeleteResponse {
            deleted: true,
            individual_id: "person_1".to_string(),
            warning: None,
            dependent_count: None,
        };
        let json = serde_json::to_value(&resp).unwrap();
        assert!(json.get("warning").is_none());
        assert!(json.get("dependent_count").is_none());
    }

    #[test]
    fn test_individual_error_http_statuses() {
        let not_found = IndividualError::NotFound("test".to_string());
        let response = not_found.into_response();
        assert_eq!(response.status(), StatusCode::NOT_FOUND);

        let already_exists = IndividualError::AlreadyExists("test".to_string());
        let response = already_exists.into_response();
        assert_eq!(response.status(), StatusCode::CONFLICT);

        let class_not_found = IndividualError::ClassNotFound("test".to_string());
        let response = class_not_found.into_response();
        assert_eq!(response.status(), StatusCode::UNPROCESSABLE_ENTITY);

        let incoming_refs = IndividualError::HasIncomingReferences(3);
        let response = incoming_refs.into_response();
        assert_eq!(response.status(), StatusCode::CONFLICT);

        let neo4j_not_cfg = IndividualError::Neo4jNotConfigured;
        let response = neo4j_not_cfg.into_response();
        assert_eq!(response.status(), StatusCode::SERVICE_UNAVAILABLE);

        let invalid_filter = IndividualError::InvalidFilter("bad".to_string());
        let response = invalid_filter.into_response();
        assert_eq!(response.status(), StatusCode::BAD_REQUEST);
    }

    #[test]
    fn test_individual_error_messages() {
        let err = IndividualError::NotFound("person_1".to_string());
        assert_eq!(err.to_string(), "Individual not found: person_1");

        let err = IndividualError::HasIncomingReferences(3);
        assert_eq!(
            err.to_string(),
            "Individual has 3 incoming references; delete with cascade=true to force"
        );

        let err = IndividualError::InvalidFilter("bad".to_string());
        assert_eq!(
            err.to_string(),
            "Invalid filter format: bad (expected property_id:operator:value)"
        );
    }

    #[test]
    fn test_repo_from_state_none_returns_error() {
        let state = test_app_state();
        let result = repo_from_state(&state);
        assert!(result.is_err());
        assert!(matches!(result, Err(IndividualError::Neo4jNotConfigured)));
    }
}
