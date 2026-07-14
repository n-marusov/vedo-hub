//! Property (ObjectProperty + DatatypeProperty) CRUD operations against Neo4j.
//!
//! Provides the domain model (`Property`), a repository layer for Neo4j Cypher
//! queries, and axum HTTP handlers for the REST API.
//!
//! # Edge direction convention
//!
//! Domain/Range edges point FROM property TO class, matching the RDF semantics
//! `rdfs:domain` / `rdfs:range` (the property *has* a domain/range Class).
//!
//! ```cypher
//! (p:Property)-[:DOMAIN]->(c:Class)   -- p's domain is c
//! (p:Property)-[:RANGE]->(c:Class)    -- p's range is c (ObjectProperty only)
//! ```
//!
//! This matches the `Prop.find_dependents` path already compiled in `classes.rs`
//! which traverses `(c)<-[:DOMAIN]-(prop:Property)`. Do NOT reverse the direction
//! without updating that query as well.

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    Json,
};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use thiserror::Error;
use tracing::{debug, error, info, warn};

use crate::neo4j::Neo4jPool;
use crate::AppState;

/// Helper: extracts a PropertyRepository from the application state or returns
/// a `Neo4jNotConfigured` error when no database pool is available.
fn repo_from_state(state: &AppState) -> Result<PropertyRepository, PropertyError> {
    match &state.neo4j {
        Some(pool) => Ok(PropertyRepository::new(pool.clone())),
        None => Err(PropertyError::Neo4jNotConfigured),
    }
}

// ── Domain Model ────────────────────────────────────────────────────────────

/// The type of a property.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum PropertyType {
    /// ObjectProperty — relates individuals to individuals (range = Class).
    Object,
    /// DatatypeProperty — relates individuals to literal values (range = XSD type).
    Datatype,
}

impl PropertyType {
    fn as_str(&self) -> &'static str {
        match self {
            PropertyType::Object => "object",
            PropertyType::Datatype => "datatype",
        }
    }
}

/// Characteristics applicable to ObjectProperties.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PropertyCharacteristics {
    #[serde(default)]
    pub functional: bool,
    #[serde(default)]
    pub inverse_functional: bool,
    #[serde(default)]
    pub transitive: bool,
    #[serde(default)]
    pub symmetric: bool,
    /// Minimum cardinality constraint (OWL minCardinality).
    #[serde(skip_serializing_if = "Option::is_none", default)]
    pub min_cardinality: Option<i32>,
    /// Maximum cardinality constraint (OWL maxCardinality).
    #[serde(skip_serializing_if = "Option::is_none", default)]
    pub max_cardinality: Option<i32>,
}

impl Default for PropertyCharacteristics {
    fn default() -> Self {
        Self {
            functional: false,
            inverse_functional: false,
            transitive: false,
            symmetric: false,
            min_cardinality: None,
            max_cardinality: None,
        }
    }
}

/// An annotation on a property (custom annotation properties beyond label/comment).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Annotation {
    /// The annotation property IRI (e.g. `"skos:definition"`, `"rdfs:seeAlso"`).
    pub property_iri: String,
    /// The annotation value.
    pub value: String,
}

/// Input for adding a single annotation.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AnnotationInput {
    pub property_iri: String,
    pub value: String,
}

/// A property in the ontology (ObjectProperty or DatatypeProperty).
#[derive(Debug, Clone, Serialize)]
pub struct Property {
    /// Unique identifier within the ontology.
    pub id: String,
    /// Human-readable label (rdfs:label).
    pub label: String,
    /// Optional comment (rdfs:comment).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub comment: Option<String>,
    /// The property type.
    pub property_type: PropertyType,
    /// Domain class IDs (classes this property applies to).
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub domains: Vec<String>,
    /// Range class IDs (ObjectProperty only — the classes linked by this property).
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub ranges: Vec<String>,
    /// XSD type for DatatypeProperty (e.g. `"string"`, `"integer"`, `"boolean"`).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub xsd_type: Option<String>,
    /// ObjectProperty characteristics (ignored for DatatypeProperty).
    #[serde(default)]
    pub characteristics: PropertyCharacteristics,
    /// Custom annotations (beyond rdfs:label / rdfs:comment).
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub annotations: Vec<Annotation>,
}

/// Valid XSD types for DatatypeProperty.
const VALID_XSD_TYPES: &[&str] = &["string", "integer", "boolean", "date", "float"];

/// Normalizes an XSD type input to the short form.
/// Accepts: `"string"`, `"xsd:string"`, `"http://www.w3.org/2001/XMLSchema#string"`.
fn normalize_xsd_type(raw: &str) -> Option<&'static str> {
    let clean = raw
        .strip_prefix("xsd:")
        .or_else(|| raw.strip_prefix("http://www.w3.org/2001/XMLSchema#"))
        .unwrap_or(raw);
    VALID_XSD_TYPES.iter().find(|&&t| t == clean).copied()
}

/// Request body for creating a property.
#[derive(Debug, Deserialize)]
pub struct CreatePropertyRequest {
    pub id: String,
    pub label: String,
    #[serde(default)]
    pub comment: Option<String>,
    pub property_type: PropertyType,
    #[serde(default)]
    pub domains: Vec<String>,
    /// Range class IDs (ObjectProperty only).
    #[serde(default)]
    pub ranges: Vec<String>,
    /// XSD type (DatatypeProperty only).
    #[serde(default)]
    pub xsd_type: Option<String>,
    /// ObjectProperty characteristics.
    #[serde(default)]
    pub characteristics: PropertyCharacteristics,
    /// Custom annotations.
    #[serde(default)]
    pub annotations: Vec<AnnotationInput>,
}

/// Request body for updating a property.
///
/// Note: property_type cannot be changed after creation. If present in the body
/// it is validated against the existing type and any mismatch is rejected.
#[derive(Debug, Deserialize)]
pub struct UpdatePropertyRequest {
    pub label: String,
    #[serde(default)]
    pub comment: Option<String>,
    #[serde(default)]
    pub domains: Vec<String>,
    /// Range class IDs (ObjectProperty only; ignored for DatatypeProperty).
    #[serde(default)]
    pub ranges: Vec<String>,
    /// XSD type (DatatypeProperty only; ignored for ObjectProperty).
    #[serde(default)]
    pub xsd_type: Option<String>,
    /// ObjectProperty characteristics (ignored for DatatypeProperty).
    #[serde(default)]
    pub characteristics: PropertyCharacteristics,
    /// Custom annotations — replaces the entire annotation list.
    #[serde(default)]
    pub annotations: Vec<AnnotationInput>,
}

/// Query parameters for listing properties.
#[derive(Debug, Deserialize)]
pub struct ListPropertiesParams {
    /// Text search filter (matches label).
    #[serde(default)]
    pub q: String,
    /// Filter by type: "object" or "datatype".
    #[serde(default)]
    pub property_type: Option<String>,
    /// Zero-based page offset (default: 0).
    #[serde(default = "default_page")]
    pub page: u64,
    /// Items per page (default: 20, max: 100).
    #[serde(default = "default_per_page")]
    pub per_page: u64,
}

fn default_page() -> u64 {
    0
}
fn default_per_page() -> u64 {
    20
}

/// Lightweight property summary used in list responses.
#[derive(Debug, Clone, Serialize)]
pub struct PropertySummary {
    pub id: String,
    pub label: String,
    pub property_type: PropertyType,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub xsd_type: Option<String>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub domains: Vec<String>,
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
    pub property_id: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub warning: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub dependent_count: Option<u64>,
}

// ── Error Handling ──────────────────────────────────────────────────────────

/// Errors that can occur during property operations.
#[derive(Debug, Error)]
pub enum PropertyError {
    #[error("Property not found: {0}")]
    NotFound(String),

    #[error("Property already exists: {0}")]
    AlreadyExists(String),

    #[error("Domain class not found: {0}")]
    DomainClassNotFound(String),

    #[error("Range class not found: {0}")]
    RangeClassNotFound(String),

    #[error("Invalid XSD type: {0} (valid values: string, integer, boolean, date, float)")]
    InvalidXsdType(String),

    #[error("ObjectProperty must have at least one range class")]
    MissingRange,

    #[error("Property must have at least one domain class")]
    MissingDomain,

    #[error("Annotation not found on property: {0}")]
    AnnotationNotFound(String),

    #[error("Property type mismatch: expected {expected}, got {actual}")]
    PropertyTypeMismatch { expected: String, actual: String },

    #[error(
        "Property has {dependent_count} dependent references; delete with cascade=true to force"
    )]
    HasDependents {
        property_id: String,
        dependent_count: u64,
    },

    #[error("Cannot cascade delete: {0}")]
    CascadeError(String),

    #[error("Database error: {0}")]
    Database(String),

    #[error("Neo4j database is not configured")]
    Neo4jNotConfigured,
}

impl axum::response::IntoResponse for PropertyError {
    fn into_response(self) -> axum::response::Response {
        let (status, code) = match &self {
            PropertyError::NotFound(_) => (StatusCode::NOT_FOUND, "ONT-PROPERTY-NOT-FOUND"),
            PropertyError::AlreadyExists(_) => {
                (StatusCode::CONFLICT, "ONT-PROPERTY-ALREADY-EXISTS")
            }
            PropertyError::DomainClassNotFound(_) => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-PROPERTY-DOMAIN-NOT-FOUND",
            ),
            PropertyError::RangeClassNotFound(_) => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-PROPERTY-RANGE-NOT-FOUND",
            ),
            PropertyError::InvalidXsdType(_) => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-PROPERTY-INVALID-XSD-TYPE",
            ),
            PropertyError::MissingRange => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-PROPERTY-MISSING-RANGE",
            ),
            PropertyError::MissingDomain => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-PROPERTY-MISSING-DOMAIN",
            ),
            PropertyError::AnnotationNotFound(_) => {
                (StatusCode::NOT_FOUND, "ONT-PROPERTY-ANNOTATION-NOT-FOUND")
            }
            PropertyError::PropertyTypeMismatch { .. } => {
                (StatusCode::CONFLICT, "ONT-PROPERTY-TYPE-MISMATCH")
            }
            PropertyError::HasDependents { .. } => {
                (StatusCode::CONFLICT, "ONT-PROPERTY-HAS-DEPENDENTS")
            }
            PropertyError::CascadeError(_) => (
                StatusCode::INTERNAL_SERVER_ERROR,
                "ONT-PROPERTY-CASCADE-ERROR",
            ),
            PropertyError::Database(_) => (StatusCode::INTERNAL_SERVER_ERROR, "ONT-DATABASE-ERROR"),
            PropertyError::Neo4jNotConfigured => {
                (StatusCode::SERVICE_UNAVAILABLE, "ONT-NEO4J-NOT-CONFIGURED")
            }
        };
        let body = serde_json::json!({
            "error": code,
            "detail": self.to_string(),
        });
        (status, Json(body)).into_response()
    }
}

// ── Repository ──────────────────────────────────────────────────────────────

/// Repository for property CRUD operations against Neo4j.
pub struct PropertyRepository {
    pool: Neo4jPool,
}

impl PropertyRepository {
    pub fn new(pool: Neo4jPool) -> Self {
        Self { pool }
    }

    /// Creates a new property with domain/range relationships.
    pub async fn create(
        &self,
        ontology_id: &str,
        req: &CreatePropertyRequest,
    ) -> Result<Property, PropertyError> {
        debug!(
            ontology_id,
            property_id = %req.id,
            property_type = ?req.property_type,
            "Creating property"
        );

        // Validate basic constraints
        if req.domains.is_empty() {
            return Err(PropertyError::MissingDomain);
        }

        let (xsd_type, ranges) = match req.property_type {
            PropertyType::Object => {
                if req.ranges.is_empty() {
                    return Err(PropertyError::MissingRange);
                }
                (None, req.ranges.clone())
            }
            PropertyType::Datatype => {
                let raw = req.xsd_type.as_deref().unwrap_or("");
                if raw.is_empty() {
                    return Err(PropertyError::InvalidXsdType("(empty)".to_string()));
                }
                let normalized = normalize_xsd_type(raw)
                    .ok_or_else(|| PropertyError::InvalidXsdType(raw.to_string()))?;
                (Some(normalized.to_string()), Vec::new())
            }
        };

        // Check existence
        let exists = self.exists(ontology_id, &req.id).await?;
        if exists {
            return Err(PropertyError::AlreadyExists(req.id.clone()));
        }

        // Validate domain classes exist
        for class_id in &req.domains {
            if !self.class_exists(ontology_id, class_id).await? {
                return Err(PropertyError::DomainClassNotFound(class_id.clone()));
            }
        }

        // Validate range classes exist (ObjectProperty)
        for class_id in &ranges {
            if !self.class_exists(ontology_id, class_id).await? {
                return Err(PropertyError::RangeClassNotFound(class_id.clone()));
            }
        }

        let comments = req.comment.as_deref().unwrap_or("");
        let annotations_json = serde_json::to_string(&req.annotations).unwrap_or_default();
        let is_datatype = req.property_type == PropertyType::Datatype;
        let xsd = xsd_type.as_deref().unwrap_or("");

        // Create property node
        let query = "\
            CREATE (p:Property {\
                id: $id, ontology_id: $ontology_id,\
                label: $label, comment: $comment,\
                property_type: $property_type,\
                is_datatype: $is_datatype,\
                xsd_type: $xsd_type,\
                functional: $functional,\
                inverse_functional: $inverse_functional,\
                transitive: $transitive,\
                symmetric: $symmetric,\
                min_cardinality: $min_cardinality,\
                max_cardinality: $max_cardinality,\
                annotations: $annotations\
            })\
            WITH p\
            UNWIND $domain_ids AS domain_id\
            MATCH (c:Class {id: domain_id, ontology_id: $ontology_id})\
            CREATE (p)-[:DOMAIN]->(c)\
            WITH p\
            UNWIND $range_ids AS range_id\
            MATCH (c:Class {id: range_id, ontology_id: $ontology_id})\
            CREATE (p)-[:RANGE]->(c)\
            RETURN p\
        ";

        let q = neo4rs::Query::new(query.to_string())
            .param("ontology_id", ontology_id)
            .param("id", req.id.as_str())
            .param("label", req.label.as_str())
            .param("comment", comments)
            .param("property_type", req.property_type.as_str())
            .param("is_datatype", is_datatype)
            .param("xsd_type", xsd)
            .param("functional", req.characteristics.functional)
            .param("inverse_functional", req.characteristics.inverse_functional)
            .param("transitive", req.characteristics.transitive)
            .param("symmetric", req.characteristics.symmetric)
            .param("min_cardinality", req.characteristics.min_cardinality)
            .param("max_cardinality", req.characteristics.max_cardinality)
            .param("annotations", annotations_json.as_str())
            .param("domain_ids", req.domains.clone())
            .param("range_ids", ranges.clone());

        match self.pool.graph().execute(q).await {
            Ok(mut result) => {
                while let Ok(Some(_)) = result.next().await {}
                info!(
                    ontology_id,
                    property_id = %req.id,
                    property_type = ?req.property_type,
                    domain_count = req.domains.len(),
                    range_count = if is_datatype { 0 } else { ranges.len() },
                    "Property created successfully"
                );
                Ok(self.assemble_property(
                    &req.id,
                    &req.label,
                    &req.comment,
                    req.property_type,
                    &req.domains,
                    req.property_type == PropertyType::Object,
                    if req.property_type == PropertyType::Object {
                        &ranges
                    } else {
                        &[]
                    },
                    &xsd_type,
                    &req.characteristics,
                    &req.annotations,
                ))
            }
            Err(e) => {
                error!(error = %e, ontology_id, property_id = %req.id, "Failed to create property");
                Err(PropertyError::Database(e.to_string()))
            }
        }
    }

    /// Retrieves a single property by ID with its domain/range relationships.
    pub async fn get(
        &self,
        ontology_id: &str,
        property_id: &str,
    ) -> Result<Property, PropertyError> {
        debug!(ontology_id, %property_id, "Reading property");

        let query = "\
            MATCH (p:Property {id: $property_id, ontology_id: $ontology_id})\
            OPTIONAL MATCH (p)-[:DOMAIN]->(d:Class)\
            OPTIONAL MATCH (p)-[:RANGE]->(r:Class)\
            RETURN \
                p.id AS id, p.label AS label, p.comment AS comment, \
                p.property_type AS property_type, p.xsd_type AS xsd_type, \
                p.functional AS functional, p.inverse_functional AS inverse_functional, \
                p.transitive AS transitive, p.symmetric AS symmetric, \
                p.min_cardinality AS min_cardinality, \
                p.max_cardinality AS max_cardinality, \
                p.annotations AS annotations, \
                collect(DISTINCT d.id) AS domain_ids, \
                collect(DISTINCT r.id) AS range_ids\
        ";

        let q = neo4rs::Query::new(query.to_string())
            .param("ontology_id", ontology_id)
            .param("property_id", property_id);

        let mut result = self.pool.graph().execute(q).await.map_err(|e| {
            error!(error = %e, ontology_id, %property_id, "Failed to read property");
            PropertyError::Database(e.to_string())
        })?;

        match result
            .next()
            .await
            .map_err(|e| PropertyError::Database(e.to_string()))?
        {
            Some(row) => {
                let id: String = row
                    .get("id")
                    .map_err(|e| PropertyError::Database(e.to_string()))?;
                let label: String = row
                    .get("label")
                    .map_err(|e| PropertyError::Database(e.to_string()))?;

                if id.is_empty() {
                    return Err(PropertyError::NotFound(property_id.to_string()));
                }

                let comment_val: Option<String> =
                    row.get("comment").ok().filter(|c: &String| !c.is_empty());
                let property_type_str: String = row
                    .get("property_type")
                    .unwrap_or_else(|_| "object".to_string());
                let property_type = match property_type_str.as_str() {
                    "datatype" => PropertyType::Datatype,
                    _ => PropertyType::Object,
                };
                let xsd_type: Option<String> =
                    row.get("xsd_type").ok().filter(|s: &String| !s.is_empty());

                let characteristics = PropertyCharacteristics {
                    functional: row.get("functional").unwrap_or(false),
                    inverse_functional: row.get("inverse_functional").unwrap_or(false),
                    transitive: row.get("transitive").unwrap_or(false),
                    symmetric: row.get("symmetric").unwrap_or(false),
                    min_cardinality: row.get("min_cardinality").ok(),
                    max_cardinality: row.get("max_cardinality").ok(),
                };

                let annotations_raw: Option<String> = row.get("annotations").ok();
                let annotations: Vec<Annotation> = annotations_raw
                    .as_deref()
                    .and_then(|s| serde_json::from_str(s).ok())
                    .unwrap_or_default();

                let domain_ids: Vec<String> = row.get("domain_ids").unwrap_or_default();
                let range_ids: Vec<String> = row.get("range_ids").unwrap_or_default();

                debug!(
                    ontology_id,
                    %property_id,
                    property_type = ?property_type,
                    domain_count = domain_ids.len(),
                    range_count = range_ids.len(),
                    annotation_count = annotations.len(),
                    "Property read successfully"
                );

                Ok(Property {
                    id,
                    label,
                    comment: comment_val,
                    property_type,
                    domains: domain_ids,
                    ranges: if property_type == PropertyType::Object {
                        range_ids
                    } else {
                        Vec::new()
                    },
                    xsd_type: if property_type == PropertyType::Datatype {
                        xsd_type
                    } else {
                        None
                    },
                    characteristics,
                    annotations,
                })
            }
            None => Err(PropertyError::NotFound(property_id.to_string())),
        }
    }

    /// Updates a property's fields and relationships.
    pub async fn update(
        &self,
        ontology_id: &str,
        property_id: &str,
        req: &UpdatePropertyRequest,
    ) -> Result<Property, PropertyError> {
        debug!(ontology_id, %property_id, "Updating property");

        // Fetch existing property to get its type (can't change type)
        let existing = self.get(ontology_id, property_id).await?;
        let property_type = existing.property_type;

        if property_type == PropertyType::Object && req.domains.is_empty() {
            return Err(PropertyError::MissingDomain);
        }
        if property_type == PropertyType::Object && req.ranges.is_empty() {
            return Err(PropertyError::MissingRange);
        }
        if property_type == PropertyType::Datatype {
            let raw = req.xsd_type.as_deref().unwrap_or("");
            if raw.is_empty() {
                return Err(PropertyError::InvalidXsdType("(empty)".to_string()));
            }
            let _ = normalize_xsd_type(raw)
                .ok_or_else(|| PropertyError::InvalidXsdType(raw.to_string()))?;
        }

        // Validate domain/range classes exist
        for class_id in &req.domains {
            if !self.class_exists(ontology_id, class_id).await? {
                return Err(PropertyError::DomainClassNotFound(class_id.clone()));
            }
        }
        if property_type == PropertyType::Object {
            for class_id in &req.ranges {
                if !self.class_exists(ontology_id, class_id).await? {
                    return Err(PropertyError::RangeClassNotFound(class_id.clone()));
                }
            }
        }

        let comments = req.comment.as_deref().unwrap_or("");
        let annotations_json = serde_json::to_string(&req.annotations).unwrap_or_default();

        // For DatatypeProperty, characteristics are always cleared
        let characteristics = if property_type == PropertyType::Datatype {
            PropertyCharacteristics::default()
        } else {
            req.characteristics.clone()
        };

        let xsd_normalized = if property_type == PropertyType::Datatype {
            req.xsd_type.as_deref().and_then(normalize_xsd_type)
        } else {
            None
        };
        let xsd_val = xsd_normalized.as_deref().unwrap_or("");

        // Rewrite: update node fields, delete old domain/range edges, recreate
        let query = "\
            MATCH (p:Property {id: $property_id, ontology_id: $ontology_id})\
            SET \
                p.label = $label,\
                p.comment = $comment,\
                p.xsd_type = $xsd_type,\
                p.functional = $functional,\
                p.inverse_functional = $inverse_functional,\
                p.transitive = $transitive,\
                p.symmetric = $symmetric,\
                p.min_cardinality = $min_cardinality,\
                p.max_cardinality = $max_cardinality,\
                p.annotations = $annotations\
            OPTIONAL MATCH (p)-[d:DOMAIN]->()\
            DELETE d\
            WITH p\
            UNWIND $domain_ids AS domain_id\
            MATCH (c:Class {id: domain_id, ontology_id: $ontology_id})\
            CREATE (p)-[:DOMAIN]->(c)\
            WITH p\
            OPTIONAL MATCH (p)-[r:RANGE]->()\
            DELETE r\
            WITH p\
            UNWIND $range_ids AS range_id\
            MATCH (c:Class {id: range_id, ontology_id: $ontology_id})\
            CREATE (p)-[:RANGE]->(c)\
            WITH p\
            RETURN p\
        ";

        let q = neo4rs::Query::new(query.to_string())
            .param("ontology_id", ontology_id)
            .param("property_id", property_id)
            .param("label", req.label.as_str())
            .param("comment", comments)
            .param("xsd_type", xsd_val)
            .param("functional", characteristics.functional)
            .param("inverse_functional", characteristics.inverse_functional)
            .param("transitive", characteristics.transitive)
            .param("symmetric", characteristics.symmetric)
            .param("min_cardinality", characteristics.min_cardinality)
            .param("max_cardinality", characteristics.max_cardinality)
            .param("annotations", annotations_json.as_str())
            .param("domain_ids", req.domains.clone())
            .param("range_ids", req.ranges.clone());

        let mut result = self.pool.graph().execute(q).await.map_err(|e| {
            error!(error = %e, ontology_id, %property_id, "Failed to update property");
            PropertyError::Database(e.to_string())
        })?;

        while let Ok(Some(_)) = result.next().await {}

        info!(ontology_id, %property_id, "Property updated successfully");

        Ok(Property {
            id: property_id.to_string(),
            label: req.label.clone(),
            comment: req.comment.clone(),
            property_type,
            domains: req.domains.clone(),
            ranges: if property_type == PropertyType::Object {
                req.ranges.clone()
            } else {
                Vec::new()
            },
            xsd_type: if property_type == PropertyType::Datatype {
                xsd_normalized.map(|s| s.to_string())
            } else {
                None
            },
            characteristics,
            annotations: req
                .annotations
                .iter()
                .map(|a| Annotation {
                    property_iri: a.property_iri.clone(),
                    value: a.value.clone(),
                })
                .collect(),
        })
    }

    /// Deletes a property, optionally cascading (detaching dependencies).
    pub async fn delete(
        &self,
        ontology_id: &str,
        property_id: &str,
        cascade: bool,
    ) -> Result<DeleteResponse, PropertyError> {
        debug!(ontology_id, %property_id, cascade, "Deleting property");

        let inverse_count = self
            .count_inverse_references(ontology_id, property_id)
            .await?;

        if !cascade && inverse_count > 0 {
            warn!(ontology_id, %property_id, inverse_count, "Delete blocked: property has inverse references");
            return Err(PropertyError::HasDependents {
                property_id: property_id.to_string(),
                dependent_count: inverse_count,
            });
        }

        // Verify existence first
        if !self.exists(ontology_id, property_id).await? {
            return Err(PropertyError::NotFound(property_id.to_string()));
        }

        let q = neo4rs::Query::new(
            "MATCH (p:Property {id: $property_id, ontology_id: $ontology_id}) DETACH DELETE p RETURN count(p) AS deleted"
                .to_string(),
        )
        .param("ontology_id", ontology_id)
        .param("property_id", property_id);

        let mut result = self.pool.graph().execute(q).await.map_err(|e| {
            error!(error = %e, ontology_id, %property_id, "Failed to delete property");
            PropertyError::Database(e.to_string())
        })?;

        let nodes_deleted: i64 = match result
            .next()
            .await
            .map_err(|e| PropertyError::Database(e.to_string()))?
        {
            Some(row) => row
                .get("deleted")
                .map_err(|e| PropertyError::Database(e.to_string()))?,
            None => 0,
        };

        if nodes_deleted == 0 {
            return Err(PropertyError::NotFound(property_id.to_string()));
        }

        let warning = if cascade && inverse_count > 0 {
            Some(format!(
                "Deleted with cascade: {} inverse property references detached",
                inverse_count,
            ))
        } else {
            None
        };

        info!(ontology_id, %property_id, nodes_deleted, "Property deleted successfully");
        Ok(DeleteResponse {
            deleted: true,
            property_id: property_id.to_string(),
            warning,
            dependent_count: if cascade { Some(inverse_count) } else { None },
        })
    }

    /// Lists properties with optional type filter, text search, and pagination.
    pub async fn list(
        &self,
        ontology_id: &str,
        params: &ListPropertiesParams,
    ) -> Result<PaginatedResponse<PropertySummary>, PropertyError> {
        debug!(
            ontology_id,
            search = %params.q,
            property_type = ?params.property_type,
            page = params.page,
            per_page = params.per_page,
            "Listing properties"
        );

        let skip = params.page * params.per_page;
        let limit = params.per_page.min(100);

        let type_filter = |ptype: &Option<String>| -> String {
            match ptype.as_deref() {
                Some("datatype") => "AND p.property_type = 'datatype'".to_string(),
                Some("object") => "AND p.property_type = 'object'".to_string(),
                _ => String::new(),
            }
        };

        let type_pred = type_filter(&params.property_type);

        let query = format!(
            "\
            MATCH (p:Property {{ontology_id: $ontology_id}})\
            WHERE ($search = '' OR toLower(p.label) CONTAINS toLower($search)) {type_pred}\
            OPTIONAL MATCH (p)-[:DOMAIN]->(d:Class)\
            WITH p, collect(DISTINCT d.id) AS domain_ids\
            RETURN p.id AS id, p.label AS label, \
                   p.property_type AS property_type, p.xsd_type AS xsd_type, domain_ids\
            ORDER BY p.label SKIP $skip LIMIT $limit\
        "
        );

        let q = neo4rs::Query::new(query.to_string())
            .param("ontology_id", ontology_id)
            .param("search", params.q.as_str())
            .param("skip", skip as i64)
            .param("limit", limit as i64);

        let mut result = self.pool.graph().execute(q).await.map_err(|e| {
            error!(error = %e, ontology_id, "Failed to list properties");
            PropertyError::Database(e.to_string())
        })?;

        let mut items = Vec::new();
        while let Ok(Some(row)) = result.next().await {
            if let Ok(id) = row.get::<String>("id") {
                let label: String = row.get("label").unwrap_or_default();
                let ptype_str: String = row
                    .get("property_type")
                    .unwrap_or_else(|_| "object".to_string());
                let ptype = match ptype_str.as_str() {
                    "datatype" => PropertyType::Datatype,
                    _ => PropertyType::Object,
                };
                let xsd: Option<String> =
                    row.get("xsd_type").ok().filter(|s: &String| !s.is_empty());
                let domain_ids: Vec<String> = row.get("domain_ids").unwrap_or_default();
                items.push(PropertySummary {
                    id,
                    label,
                    property_type: ptype,
                    xsd_type: xsd,
                    domains: domain_ids,
                });
            }
        }

        let total = self
            .count(ontology_id, &params.q, &params.property_type)
            .await?;
        debug!(
            ontology_id,
            returned = items.len(),
            total,
            "Properties listed"
        );
        Ok(PaginatedResponse {
            items,
            total,
            page: params.page,
            per_page: limit,
        })
    }

    // ── Internal helpers ────────────────────────────────────────────────────

    fn assemble_property(
        &self,
        id: &str,
        label: &str,
        comment: &Option<String>,
        property_type: PropertyType,
        domains: &[String],
        _is_object: bool,
        ranges: &[String],
        xsd_type: &Option<String>,
        characteristics: &PropertyCharacteristics,
        annotations: &[AnnotationInput],
    ) -> Property {
        Property {
            id: id.to_string(),
            label: label.to_string(),
            comment: comment.clone(),
            property_type,
            domains: domains.to_vec(),
            ranges: ranges.to_vec(),
            xsd_type: xsd_type.clone(),
            characteristics: characteristics.clone(),
            annotations: annotations
                .iter()
                .map(|a| Annotation {
                    property_iri: a.property_iri.clone(),
                    value: a.value.clone(),
                })
                .collect(),
        }
    }

    async fn exists(&self, ontology_id: &str, property_id: &str) -> Result<bool, PropertyError> {
        let query =
            "MATCH (p:Property {id: $property_id, ontology_id: $ontology_id}) RETURN count(p) AS cnt";
        let mut result = self
            .pool
            .graph()
            .execute(
                neo4rs::Query::new(query.to_string())
                    .param("ontology_id", ontology_id)
                    .param("property_id", property_id),
            )
            .await
            .map_err(|e| PropertyError::Database(e.to_string()))?;

        match result
            .next()
            .await
            .map_err(|e| PropertyError::Database(e.to_string()))?
        {
            Some(row) => {
                let cnt: i64 = row
                    .get("cnt")
                    .map_err(|e| PropertyError::Database(e.to_string()))?;
                Ok(cnt > 0)
            }
            None => Ok(false),
        }
    }

    /// Checks whether a Class exists (used by property domain/range validation).
    async fn class_exists(&self, ontology_id: &str, class_id: &str) -> Result<bool, PropertyError> {
        let query =
            "MATCH (c:Class {id: $class_id, ontology_id: $ontology_id}) RETURN count(c) AS cnt";
        let mut result = self
            .pool
            .graph()
            .execute(
                neo4rs::Query::new(query.to_string())
                    .param("ontology_id", ontology_id)
                    .param("class_id", class_id),
            )
            .await
            .map_err(|e| PropertyError::Database(e.to_string()))?;

        match result
            .next()
            .await
            .map_err(|e| PropertyError::Database(e.to_string()))?
        {
            Some(row) => {
                let cnt: i64 = row
                    .get("cnt")
                    .map_err(|e| PropertyError::Database(e.to_string()))?;
                Ok(cnt > 0)
            }
            None => Ok(false),
        }
    }

    /// Counts inverse references to this property (from `:INVERSE_OF` edges).
    async fn count_inverse_references(
        &self,
        ontology_id: &str,
        property_id: &str,
    ) -> Result<u64, PropertyError> {
        let query = "\
            MATCH (p:Property {id: $property_id, ontology_id: $ontology_id})\
            OPTIONAL MATCH (p)<-[:INVERSE_OF]-(inv:Property)\
            RETURN count(DISTINCT inv) AS cnt\
        ";
        let mut result = self
            .pool
            .graph()
            .execute(
                neo4rs::Query::new(query.to_string())
                    .param("ontology_id", ontology_id)
                    .param("property_id", property_id),
            )
            .await
            .map_err(|e| PropertyError::Database(e.to_string()))?;

        match result
            .next()
            .await
            .map_err(|e| PropertyError::Database(e.to_string()))?
        {
            Some(row) => {
                let cnt: i64 = row
                    .get("cnt")
                    .map_err(|e| PropertyError::Database(e.to_string()))?;
                Ok(cnt as u64)
            }
            None => Ok(0),
        }
    }

    async fn count(
        &self,
        ontology_id: &str,
        search: &str,
        property_type: &Option<String>,
    ) -> Result<u64, PropertyError> {
        let type_pred = match property_type.as_deref() {
            Some("datatype") => "AND p.property_type = 'datatype'".to_string(),
            Some("object") => "AND p.property_type = 'object'".to_string(),
            _ => String::new(),
        };
        let query = format!(
            "\
            MATCH (p:Property {{ontology_id: $ontology_id}})\
            WHERE ($search = '' OR toLower(p.label) CONTAINS toLower($search)) {type_pred}\
            RETURN count(p) AS total\
        "
        );
        let mut result = self
            .pool
            .graph()
            .execute(
                neo4rs::Query::new(query.to_string())
                    .param("ontology_id", ontology_id)
                    .param("search", search),
            )
            .await
            .map_err(|e| PropertyError::Database(e.to_string()))?;

        match result
            .next()
            .await
            .map_err(|e| PropertyError::Database(e.to_string()))?
        {
            Some(row) => {
                let total: i64 = row
                    .get("total")
                    .map_err(|e| PropertyError::Database(e.to_string()))?;
                Ok(total as u64)
            }
            None => Ok(0),
        }
    }
}

// ── Axum Handlers ───────────────────────────────────────────────────────────

/// POST `/api/v1/ontologies/{ontology_id}/properties`
pub async fn create_property_handler(
    State(state): State<Arc<AppState>>,
    Path(ontology_id): Path<String>,
    Json(req): Json<CreatePropertyRequest>,
) -> Result<(StatusCode, Json<Property>), PropertyError> {
    debug!(ontology_id, property_id = %req.id, "POST create property handler invoked");
    let repo = repo_from_state(&state)?;
    let property = repo.create(&ontology_id, &req).await?;
    info!(ontology_id, property_id = %property.id, "POST create property handler completed");
    Ok((StatusCode::CREATED, Json(property)))
}

/// GET `/api/v1/ontologies/{ontology_id}/properties`
pub async fn list_properties_handler(
    State(state): State<Arc<AppState>>,
    Path(ontology_id): Path<String>,
    Query(params): Query<ListPropertiesParams>,
) -> Result<Json<PaginatedResponse<PropertySummary>>, PropertyError> {
    debug!(ontology_id, "GET list properties handler invoked");
    let repo = repo_from_state(&state)?;
    let result = repo.list(&ontology_id, &params).await?;
    Ok(Json(result))
}

/// GET `/api/v1/ontologies/{ontology_id}/properties/{property_id}`
pub async fn get_property_handler(
    State(state): State<Arc<AppState>>,
    Path((ontology_id, property_id)): Path<(String, String)>,
) -> Result<Json<Property>, PropertyError> {
    debug!(ontology_id, %property_id, "GET property handler invoked");
    let repo = repo_from_state(&state)?;
    let property = repo.get(&ontology_id, &property_id).await?;
    Ok(Json(property))
}

/// PUT `/api/v1/ontologies/{ontology_id}/properties/{property_id}`
pub async fn update_property_handler(
    State(state): State<Arc<AppState>>,
    Path((ontology_id, property_id)): Path<(String, String)>,
    Json(req): Json<UpdatePropertyRequest>,
) -> Result<Json<Property>, PropertyError> {
    debug!(ontology_id, %property_id, "PUT update property handler invoked");
    let repo = repo_from_state(&state)?;
    let property = repo.update(&ontology_id, &property_id, &req).await?;
    info!(ontology_id, property_id = %property.id, "PUT update property handler completed");
    Ok(Json(property))
}

/// DELETE `/api/v1/ontologies/{ontology_id}/properties/{property_id}`
pub async fn delete_property_handler(
    State(state): State<Arc<AppState>>,
    Path((ontology_id, property_id)): Path<(String, String)>,
    Query(params): Query<DeletePropertyParams>,
) -> Result<Json<DeleteResponse>, PropertyError> {
    debug!(ontology_id, %property_id, cascade = params.cascade, "DELETE property handler invoked");
    let repo = repo_from_state(&state)?;
    let result = repo
        .delete(&ontology_id, &property_id, params.cascade)
        .await?;
    info!(ontology_id, property_id = %result.property_id, "DELETE property handler completed");
    Ok(Json(result))
}

/// Query parameters for property deletion.
#[derive(Debug, Deserialize)]
pub struct DeletePropertyParams {
    #[serde(default)]
    pub cascade: bool,
}

// ── Tests ───────────────────────────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;
    use axum::{http::StatusCode, response::IntoResponse};

    // ── Model / Serialization Tests ─────────────────────────────────────────

    #[test]
    fn test_property_type_serde() {
        let json = serde_json::json!("object");
        let pt: PropertyType = serde_json::from_value(json).unwrap();
        assert_eq!(pt, PropertyType::Object);

        let json = serde_json::json!("datatype");
        let pt: PropertyType = serde_json::from_value(json).unwrap();
        assert_eq!(pt, PropertyType::Datatype);
    }

    #[test]
    fn test_property_type_as_str() {
        assert_eq!(PropertyType::Object.as_str(), "object");
        assert_eq!(PropertyType::Datatype.as_str(), "datatype");
    }

    #[test]
    fn test_property_characteristics_default() {
        let c = PropertyCharacteristics::default();
        assert!(!c.functional);
        assert!(!c.transitive);
        assert!(!c.symmetric);
        assert!(!c.inverse_functional);
    }

    #[test]
    fn test_property_characteristics_deserialize() {
        let json = serde_json::json!({
            "functional": true,
            "transitive": false,
            "symmetric": true,
            "inverse_functional": false
        });
        let c: PropertyCharacteristics = serde_json::from_value(json).unwrap();
        assert!(c.functional);
        assert!(!c.transitive);
        assert!(c.symmetric);
        assert!(!c.inverse_functional);
    }

    #[test]
    fn test_property_characteristics_partial_deserialize() {
        let json = serde_json::json!({
            "functional": true
        });
        let c: PropertyCharacteristics = serde_json::from_value(json).unwrap();
        assert!(c.functional);
        assert!(!c.transitive);
        assert!(!c.symmetric);
        assert!(!c.inverse_functional);
    }

    #[test]
    fn test_property_serialization_object() {
        let prop = Property {
            id: "hasChild".to_string(),
            label: "has child".to_string(),
            comment: Some("Parent-child relation".to_string()),
            property_type: PropertyType::Object,
            domains: vec!["Person".to_string()],
            ranges: vec!["Person".to_string()],
            xsd_type: None,
            characteristics: PropertyCharacteristics {
                transitive: true,
                ..Default::default()
            },
            annotations: vec![Annotation {
                property_iri: "skos:definition".to_string(),
                value: "A relationship between a parent and a child".to_string(),
            }],
        };
        let json = serde_json::to_value(&prop).unwrap();
        assert_eq!(json["id"], "hasChild");
        assert_eq!(json["label"], "has child");
        assert_eq!(json["comment"], "Parent-child relation");
        assert_eq!(json["property_type"], "object");
        assert_eq!(json["domains"].as_array().unwrap().len(), 1);
        assert_eq!(json["ranges"].as_array().unwrap().len(), 1);
        assert!(json.get("xsd_type").is_none());
        assert_eq!(json["characteristics"]["transitive"], true);
        assert_eq!(json["annotations"].as_array().unwrap().len(), 1);
    }

    #[test]
    fn test_property_serialization_datatype() {
        let prop = Property {
            id: "age".to_string(),
            label: "age".to_string(),
            comment: None,
            property_type: PropertyType::Datatype,
            domains: vec!["Person".to_string()],
            ranges: vec![],
            xsd_type: Some("integer".to_string()),
            characteristics: PropertyCharacteristics::default(),
            annotations: vec![],
        };
        let json = serde_json::to_value(&prop).unwrap();
        assert_eq!(json["id"], "age");
        assert_eq!(json["property_type"], "datatype");
        assert!(json.get("ranges").is_none()); // empty → skipped
        assert_eq!(json["xsd_type"], "integer");
        assert!(
            json.get("characteristics").is_none() || json["characteristics"]["functional"] == false
        );
    }

    #[test]
    fn test_create_property_request_minimal_object() {
        let json = serde_json::json!({
            "id": "hasChild",
            "label": "has child",
            "property_type": "object",
            "domains": ["Person"],
            "ranges": ["Person"]
        });
        let req: CreatePropertyRequest = serde_json::from_value(json).unwrap();
        assert_eq!(req.id, "hasChild");
        assert_eq!(req.property_type as PropertyType, PropertyType::Object);
        assert_eq!(req.domains, vec!["Person"]);
        assert_eq!(req.ranges, vec!["Person"]);
        assert!(req.comment.is_none());
        assert!(req.xsd_type.is_none());
    }

    #[test]
    fn test_create_property_request_minimal_datatype() {
        let json = serde_json::json!({
            "id": "age",
            "label": "age",
            "property_type": "datatype",
            "domains": ["Person"],
            "xsd_type": "integer"
        });
        let req: CreatePropertyRequest = serde_json::from_value(json).unwrap();
        assert_eq!(req.id, "age");
        assert_eq!(req.property_type as PropertyType, PropertyType::Datatype);
        assert_eq!(req.xsd_type, Some("integer".to_string()));
    }

    #[test]
    fn test_create_property_request_with_annotations() {
        let json = serde_json::json!({
            "id": "hasChild",
            "label": "has child",
            "property_type": "object",
            "domains": ["Person"],
            "ranges": ["Person"],
            "annotations": [
                {"property_iri": "skos:definition", "value": "def"},
                {"property_iri": "rdfs:seeAlso", "value": "see"}
            ]
        });
        let req: CreatePropertyRequest = serde_json::from_value(json).unwrap();
        assert_eq!(req.annotations.len(), 2);
        assert_eq!(req.annotations[0].property_iri, "skos:definition");
    }

    // ── XSD Type Normalization ──────────────────────────────────────────────

    #[test]
    fn test_normalize_xsd_type_short() {
        assert_eq!(normalize_xsd_type("string"), Some("string"));
        assert_eq!(normalize_xsd_type("integer"), Some("integer"));
        assert_eq!(normalize_xsd_type("boolean"), Some("boolean"));
        assert_eq!(normalize_xsd_type("date"), Some("date"));
        assert_eq!(normalize_xsd_type("float"), Some("float"));
    }

    #[test]
    fn test_normalize_xsd_type_prefixed() {
        assert_eq!(normalize_xsd_type("xsd:string"), Some("string"));
        assert_eq!(normalize_xsd_type("xsd:integer"), Some("integer"));
    }

    #[test]
    fn test_normalize_xsd_type_url() {
        assert_eq!(
            normalize_xsd_type("http://www.w3.org/2001/XMLSchema#string"),
            Some("string")
        );
    }

    #[test]
    fn test_normalize_xsd_type_invalid() {
        assert_eq!(normalize_xsd_type("decimal"), None);
        assert_eq!(normalize_xsd_type(""), None);
        assert_eq!(normalize_xsd_type("xsd:decimal"), None);
    }

    // ── Paginated Response Tests ────────────────────────────────────────────

    #[test]
    fn test_paginated_response_serialization() {
        let resp = PaginatedResponse::<PropertySummary> {
            items: vec![PropertySummary {
                id: "hasChild".to_string(),
                label: "has child".to_string(),
                property_type: PropertyType::Object,
                xsd_type: None,
                domains: vec![],
            }],
            total: 1,
            page: 0,
            per_page: 20,
        };
        let json = serde_json::to_value(&resp).unwrap();
        assert_eq!(json["total"], 1);
        assert_eq!(json["page"], 0);
        assert_eq!(json["per_page"], 20);
        assert_eq!(json["items"].as_array().unwrap().len(), 1);
    }

    // ── Delete Response Tests ───────────────────────────────────────────────

    #[test]
    fn test_delete_response_serialization() {
        let resp = DeleteResponse {
            deleted: true,
            property_id: "age".to_string(),
            warning: Some("Deleted with cascade: 2 inverse refs".to_string()),
            dependent_count: Some(2),
        };
        let json = serde_json::to_value(&resp).unwrap();
        assert_eq!(json["deleted"], true);
        assert_eq!(json["property_id"], "age");
        assert_eq!(json["warning"], "Deleted with cascade: 2 inverse refs");
    }

    #[test]
    fn test_delete_response_no_warning() {
        let resp = DeleteResponse {
            deleted: true,
            property_id: "orphan".to_string(),
            warning: None,
            dependent_count: None,
        };
        let json = serde_json::to_value(&resp).unwrap();
        assert_eq!(json["deleted"], true);
        assert!(json.get("warning").is_none());
        assert!(json.get("dependent_count").is_none());
    }

    // ── Error Response Tests ────────────────────────────────────────────────

    #[tokio::test]
    async fn test_property_error_not_found_response() {
        let response = PropertyError::NotFound("age".to_string()).into_response();
        assert_eq!(response.status(), StatusCode::NOT_FOUND);
    }

    #[tokio::test]
    async fn test_property_error_already_exists_response() {
        let response = PropertyError::AlreadyExists("age".to_string()).into_response();
        assert_eq!(response.status(), StatusCode::CONFLICT);
    }

    #[tokio::test]
    async fn test_property_error_domain_class_not_found() {
        let response = PropertyError::DomainClassNotFound("Missing".to_string()).into_response();
        assert_eq!(response.status(), StatusCode::UNPROCESSABLE_ENTITY);
    }

    #[tokio::test]
    async fn test_property_error_neo4j_not_configured() {
        let response = PropertyError::Neo4jNotConfigured.into_response();
        assert_eq!(response.status(), StatusCode::SERVICE_UNAVAILABLE);
    }

    #[tokio::test]
    async fn test_property_error_missing_domain() {
        let response = PropertyError::MissingDomain.into_response();
        assert_eq!(response.status(), StatusCode::UNPROCESSABLE_ENTITY);
    }

    #[tokio::test]
    async fn test_property_error_missing_range() {
        let response = PropertyError::MissingRange.into_response();
        assert_eq!(response.status(), StatusCode::UNPROCESSABLE_ENTITY);
    }

    #[tokio::test]
    async fn test_property_error_has_dependents_response() {
        let err = PropertyError::HasDependents {
            property_id: "age".to_string(),
            dependent_count: 3,
        };
        let response = err.into_response();
        assert_eq!(response.status(), StatusCode::CONFLICT);
    }

    #[tokio::test]
    async fn test_property_error_invalid_xsd_type() {
        let response = PropertyError::InvalidXsdType("decimal".to_string()).into_response();
        assert_eq!(response.status(), StatusCode::UNPROCESSABLE_ENTITY);
    }

    // ── Handler Logic Tests (direct, no router) ────────────────────────────

    #[test]
    fn test_repo_from_state_none_returns_error() {
        let state = AppState { neo4j: None };
        let result = repo_from_state(&state);
        assert!(result.is_err());
    }

    #[test]
    fn test_list_properties_params_defaults() {
        let params: ListPropertiesParams = serde_json::from_value(serde_json::json!({})).unwrap();
        assert_eq!(params.q, "");
        assert_eq!(params.page, 0);
        assert_eq!(params.per_page, 20);
        assert!(params.property_type.is_none());
    }

    #[test]
    fn test_list_properties_params_custom() {
        let params: ListPropertiesParams = serde_json::from_value(
            serde_json::json!({"q": "age", "page": 2, "per_page": 10, "property_type": "datatype"}),
        )
        .unwrap();
        assert_eq!(params.q, "age");
        assert_eq!(params.page, 2);
        assert_eq!(params.per_page, 10);
        assert_eq!(params.property_type, Some("datatype".to_string()));
    }

    #[test]
    fn test_delete_property_params_default() {
        let params: DeletePropertyParams = serde_json::from_value(serde_json::json!({})).unwrap();
        assert!(!params.cascade);
    }

    #[test]
    fn test_delete_property_params_cascade_true() {
        let params: DeletePropertyParams =
            serde_json::from_value(serde_json::json!({"cascade": true})).unwrap();
        assert!(params.cascade);
    }

    #[test]
    fn test_annotation_serde_roundtrip() {
        let ann = Annotation {
            property_iri: "skos:definition".to_string(),
            value: "A definition".to_string(),
        };
        let json = serde_json::to_value(&ann).unwrap();
        let deserialized: Annotation = serde_json::from_value(json).unwrap();
        assert_eq!(deserialized.property_iri, "skos:definition");
        assert_eq!(deserialized.value, "A definition");
    }

    #[test]
    fn test_annotation_input_serde() {
        let json = serde_json::json!({
            "property_iri": "skos:definition",
            "value": "test value"
        });
        let input: AnnotationInput = serde_json::from_value(json).unwrap();
        assert_eq!(input.property_iri, "skos:definition");
        assert_eq!(input.value, "test value");
    }
}
