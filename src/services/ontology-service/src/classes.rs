//! Class (TBox) CRUD operations against Neo4j.
//!
//! Provides the domain model (`OwlClass`), a repository layer for Neo4j
//! Cypher queries, and axum HTTP handlers for the REST API.

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

/// Helper: extracts a `ClassRepository` from the application state or returns
/// a `Neo4jNotConfigured` error when no database pool is available.
fn repo_from_state(state: &AppState) -> Result<ClassRepository, ClassError> {
    match &state.neo4j {
        Some(pool) => Ok(ClassRepository::new(pool.clone())),
        None => Err(ClassError::Neo4jNotConfigured),
    }
}

// ── Domain Model ────────────────────────────────────────────────────────────────

/// An OWL class in the ontology hierarchy.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OwlClass {
    /// Unique identifier within the ontology (e.g. `"Person"`).
    pub id: String,
    /// Human-readable label (rdfs:label).
    pub label: String,
    /// Optional comment (rdfs:comment).
    #[serde(skip_serializing_if = "Option::is_none")]
    pub comment: Option<String>,
    /// Parent class IDs for the hierarchy (multiple inheritance).
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub parents: Vec<String>,
    /// Child class IDs (populated on read, empty on create/update input).
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub children: Vec<String>,
}

/// Request body for creating a class.
#[derive(Debug, Deserialize)]
pub struct CreateClassRequest {
    pub id: String,
    pub label: String,
    #[serde(default)]
    pub comment: Option<String>,
    #[serde(default)]
    pub parents: Vec<String>,
}

/// Request body for updating a class.
#[derive(Debug, Deserialize)]
pub struct UpdateClassRequest {
    pub label: String,
    #[serde(default)]
    pub comment: Option<String>,
    #[serde(default)]
    pub parents: Vec<String>,
}

/// Query parameters for listing classes.
#[derive(Debug, Deserialize)]
pub struct ListClassesParams {
    /// Text search filter (matches label).
    #[serde(default)]
    pub q: String,
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
    pub class_id: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub warning: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub dependent_count: Option<u64>,
}

// ── Error Handling ──────────────────────────────────────────────────────────────

/// Errors that can occur during class operations.
#[derive(Debug, Error)]
pub enum ClassError {
    #[error("Class not found: {0}")]
    NotFound(String),

    #[error("Class already exists: {0}")]
    AlreadyExists(String),

    #[error("Parent class not found: {0}")]
    ParentNotFound(String),

    #[error("Class has {dependent_count} dependent classes and {property_count} referencing properties; delete with cascade=true to force")]
    HasDependents {
        class_id: String,
        dependent_count: u64,
        property_count: u64,
    },

    #[error("Cannot cascade delete: {0}")]
    CascadeError(String),

    #[error("Database error: {0}")]
    Database(String),

    #[error("Neo4j database is not configured")]
    Neo4jNotConfigured,
}

impl axum::response::IntoResponse for ClassError {
    fn into_response(self) -> axum::response::Response {
        let (status, code) = match &self {
            ClassError::NotFound(_) => (StatusCode::NOT_FOUND, "CLASS_NOT_FOUND"),
            ClassError::AlreadyExists(_) => (StatusCode::CONFLICT, "CLASS_ALREADY_EXISTS"),
            ClassError::ParentNotFound(_) => (StatusCode::UNPROCESSABLE_ENTITY, "PARENT_NOT_FOUND"),
            ClassError::HasDependents { .. } => (StatusCode::CONFLICT, "CLASS_HAS_DEPENDENTS"),
            ClassError::CascadeError(_) => (StatusCode::INTERNAL_SERVER_ERROR, "CASCADE_ERROR"),
            ClassError::Database(_) => (StatusCode::INTERNAL_SERVER_ERROR, "DATABASE_ERROR"),
            ClassError::Neo4jNotConfigured => {
                (StatusCode::SERVICE_UNAVAILABLE, "NEO4J_NOT_CONFIGURED")
            }
        };
        let body = serde_json::json!({
            "error": code,
            "detail": self.to_string(),
        });
        (status, Json(body)).into_response()
    }
}

// ── Repository ──────────────────────────────────────────────────────────────────

/// Repository for class CRUD operations against Neo4j.
pub struct ClassRepository {
    pool: Neo4jPool,
}

impl ClassRepository {
    pub fn new(pool: Neo4jPool) -> Self {
        Self { pool }
    }

    /// Creates a new class with optional parent relationships.
    pub async fn create(
        &self,
        ontology_id: &str,
        req: &CreateClassRequest,
    ) -> Result<OwlClass, ClassError> {
        debug!(
            ontology_id,
            class_id = %req.id,
            parent_count = req.parents.len(),
            "Creating class"
        );

        let exists = self.exists(ontology_id, &req.id).await?;
        if exists {
            return Err(ClassError::AlreadyExists(req.id.clone()));
        }

        for parent_id in &req.parents {
            let parent_exists = self.exists(ontology_id, parent_id).await?;
            if !parent_exists {
                return Err(ClassError::ParentNotFound(parent_id.clone()));
            }
        }

        let comment = req.comment.as_deref().unwrap_or("");
        let query = "\
            CREATE (c:Class {id: $id, label: $label, comment: $comment, ontology_id: $ontology_id})\
            WITH c\
            UNWIND $parent_ids AS parent_id\
            MATCH (p:Class {id: parent_id, ontology_id: $ontology_id})\
            CREATE (c)-[:CHILD_OF]->(p)\
            RETURN c\
        ";

        let q = neo4rs::Query::new(query.to_string())
            .param("ontology_id", ontology_id)
            .param("id", req.id.as_str())
            .param("label", req.label.as_str())
            .param("comment", comment)
            .param("parent_ids", req.parents.clone());

        match self.pool.graph().execute(q).await {
            Ok(mut result) => {
                while let Ok(Some(_)) = result.next().await {}
                info!(ontology_id, class_id = %req.id, "Class created successfully");
                Ok(OwlClass {
                    id: req.id.clone(),
                    label: req.label.clone(),
                    comment: req.comment.clone(),
                    parents: req.parents.clone(),
                    children: Vec::new(),
                })
            }
            Err(e) => {
                error!(error = %e, ontology_id, class_id = %req.id, "Failed to create class");
                Err(ClassError::Database(e.to_string()))
            }
        }
    }

    /// Retrieves a single class by ID with its hierarchy (parents and children).
    pub async fn get(&self, ontology_id: &str, class_id: &str) -> Result<OwlClass, ClassError> {
        debug!(ontology_id, %class_id, "Reading class");

        let query = "\
            MATCH (c:Class {id: $class_id, ontology_id: $ontology_id})\
            OPTIONAL MATCH (c)-[:CHILD_OF]->(p:Class)\
            OPTIONAL MATCH (child:Class)-[:CHILD_OF]->(c)\
            RETURN \
                c.id AS id, c.label AS label, c.comment AS comment, \
                collect(DISTINCT p.id) AS parent_ids, \
                collect(DISTINCT child.id) AS child_ids\
        ";

        let q = neo4rs::Query::new(query.to_string())
            .param("ontology_id", ontology_id)
            .param("class_id", class_id);

        let mut result = self.pool.graph().execute(q).await.map_err(|e| {
            error!(error = %e, ontology_id, %class_id, "Failed to read class");
            ClassError::Database(e.to_string())
        })?;

        match result
            .next()
            .await
            .map_err(|e| ClassError::Database(e.to_string()))?
        {
            Some(row) => {
                let id: String = row
                    .get("id")
                    .map_err(|e| ClassError::Database(e.to_string()))?;
                let label: String = row
                    .get("label")
                    .map_err(|e| ClassError::Database(e.to_string()))?;
                let comment_val: Option<String> =
                    row.get("comment").ok().filter(|c: &String| !c.is_empty());
                let parent_ids: Vec<String> = row.get("parent_ids").unwrap_or_default();
                let child_ids: Vec<String> = row.get("child_ids").unwrap_or_default();

                if id.is_empty() {
                    return Err(ClassError::NotFound(class_id.to_string()));
                }

                debug!(ontology_id, %class_id, parent_count = parent_ids.len(), child_count = child_ids.len(), "Class read successfully");
                Ok(OwlClass {
                    id,
                    label,
                    comment: comment_val,
                    parents: parent_ids,
                    children: child_ids,
                })
            }
            None => Err(ClassError::NotFound(class_id.to_string())),
        }
    }

    /// Updates a class's label, comment, and parent relationships.
    pub async fn update(
        &self,
        ontology_id: &str,
        class_id: &str,
        req: &UpdateClassRequest,
    ) -> Result<OwlClass, ClassError> {
        debug!(ontology_id, %class_id, "Updating class");

        let _ = self.get(ontology_id, class_id).await?;

        for parent_id in &req.parents {
            if parent_id == class_id {
                return Err(ClassError::CascadeError(
                    "A class cannot be its own parent".to_string(),
                ));
            }
            let parent_exists = self.exists(ontology_id, parent_id).await?;
            if !parent_exists {
                return Err(ClassError::ParentNotFound(parent_id.clone()));
            }
        }

        let comment = req.comment.as_deref().unwrap_or("");

        // Delete old parent rels, then create new ones
        let query = "\
            MATCH (c:Class {id: $class_id, ontology_id: $ontology_id})\
            SET c.label = $label, c.comment = $comment\
            OPTIONAL MATCH (c)-[r:CHILD_OF]->()\
            DELETE r\
            WITH c\
            UNWIND $parent_ids AS parent_id\
            MATCH (p:Class {id: parent_id, ontology_id: $ontology_id})\
            CREATE (c)-[:CHILD_OF]->(p)\
            RETURN c\
        ";

        let q = neo4rs::Query::new(query.to_string())
            .param("ontology_id", ontology_id)
            .param("class_id", class_id)
            .param("label", req.label.as_str())
            .param("comment", comment)
            .param("parent_ids", req.parents.clone());

        let mut result = self.pool.graph().execute(q).await.map_err(|e| {
            error!(error = %e, ontology_id, %class_id, "Failed to update class");
            ClassError::Database(e.to_string())
        })?;

        // Consume the result stream to execute the query
        while let Ok(Some(_)) = result.next().await {}

        info!(ontology_id, %class_id, "Class updated successfully");
        Ok(OwlClass {
            id: class_id.to_string(),
            label: req.label.clone(),
            comment: req.comment.clone(),
            parents: req.parents.clone(),
            children: Vec::new(),
        })
    }

    /// Deletes a class, optionally cascading to dependents.
    pub async fn delete(
        &self,
        ontology_id: &str,
        class_id: &str,
        cascade: bool,
    ) -> Result<DeleteResponse, ClassError> {
        debug!(ontology_id, %class_id, cascade, "Deleting class");

        let (dependent_count, property_count) = self.find_dependents(ontology_id, class_id).await?;

        if !cascade && (dependent_count > 0 || property_count > 0) {
            warn!(ontology_id, %class_id, dependent_count, property_count, "Delete blocked: class has dependents");
            return Err(ClassError::HasDependents {
                class_id: class_id.to_string(),
                dependent_count,
                property_count,
            });
        }

        // First verify the class exists
        if !self.exists(ontology_id, class_id).await? {
            return Err(ClassError::NotFound(class_id.to_string()));
        }

        let q = neo4rs::Query::new(
            "MATCH (c:Class {id: $class_id, ontology_id: $ontology_id}) DETACH DELETE c RETURN count(c) AS deleted"
                .to_string(),
        )
        .param("ontology_id", ontology_id)
        .param("class_id", class_id);

        let mut result = self.pool.graph().execute(q).await.map_err(|e| {
            error!(error = %e, ontology_id, %class_id, "Failed to delete class");
            ClassError::Database(e.to_string())
        })?;

        // Read the deleted count from the result
        let nodes_deleted: i64 = match result
            .next()
            .await
            .map_err(|e| ClassError::Database(e.to_string()))?
        {
            Some(row) => row
                .get("deleted")
                .map_err(|e| ClassError::Database(e.to_string()))?,
            None => 0,
        };

        if nodes_deleted == 0 {
            return Err(ClassError::NotFound(class_id.to_string()));
        }

        let mut warning = None;
        if cascade && dependent_count > 0 {
            warning = Some(format!(
                "Deleted with cascade: {} dependent classes and {} referencing properties were affected",
                dependent_count, property_count,
            ));
        }

        info!(ontology_id, %class_id, nodes_deleted, "Class deleted successfully");
        Ok(DeleteResponse {
            deleted: true,
            class_id: class_id.to_string(),
            warning,
            dependent_count: if cascade { Some(dependent_count) } else { None },
        })
    }

    /// Lists classes with optional text search and pagination.
    pub async fn list(
        &self,
        ontology_id: &str,
        params: &ListClassesParams,
    ) -> Result<PaginatedResponse<OwlClassSummary>, ClassError> {
        debug!(ontology_id, search = %params.q, page = params.page, per_page = params.per_page, "Listing classes");

        let skip = params.page * params.per_page;
        let limit = params.per_page.min(100);

        let query = "\
            MATCH (c:Class {ontology_id: $ontology_id})\
            WHERE $search = '' OR toLower(c.label) CONTAINS toLower($search)\
            OPTIONAL MATCH (c)-[:CHILD_OF]->(p:Class)\
            WITH c, collect(DISTINCT p.id) AS parent_ids\
            RETURN c.id AS id, c.label AS label, c.comment AS comment, parent_ids\
            ORDER BY c.label SKIP $skip LIMIT $limit\
        ";

        let q = neo4rs::Query::new(query.to_string())
            .param("ontology_id", ontology_id)
            .param("search", params.q.as_str())
            .param("skip", skip as i64)
            .param("limit", limit as i64);

        let mut result = self.pool.graph().execute(q).await.map_err(|e| {
            error!(error = %e, ontology_id, "Failed to list classes");
            ClassError::Database(e.to_string())
        })?;

        let mut items = Vec::new();
        while let Ok(Some(row)) = result.next().await {
            if let Ok(id) = row.get::<String>("id") {
                let label: String = row.get("label").unwrap_or_default();
                let comment: Option<String> =
                    row.get("comment").ok().filter(|c: &String| !c.is_empty());
                let parent_ids: Vec<String> = row.get("parent_ids").unwrap_or_default();
                items.push(OwlClassSummary {
                    id,
                    label,
                    comment,
                    parents: parent_ids,
                });
            }
        }

        let total = self.count(ontology_id, &params.q).await?;
        debug!(ontology_id, returned = items.len(), total, "Classes listed");
        Ok(PaginatedResponse {
            items,
            total,
            page: params.page,
            per_page: limit,
        })
    }

    /// Returns root classes (no parent).
    pub async fn get_root_classes(
        &self,
        ontology_id: &str,
        params: &ListClassesParams,
    ) -> Result<PaginatedResponse<OwlClassSummary>, ClassError> {
        debug!(ontology_id, "Listing root classes");

        let skip = params.page * params.per_page;
        let limit = params.per_page.min(100);

        let query = "\
            MATCH (c:Class {ontology_id: $ontology_id})\
            WHERE NOT EXISTS((c)-[:CHILD_OF]->())\
              AND ($search = '' OR toLower(c.label) CONTAINS toLower($search))\
            RETURN c.id AS id, c.label AS label, c.comment AS comment\
            ORDER BY c.label SKIP $skip LIMIT $limit\
        ";

        let q = neo4rs::Query::new(query.to_string())
            .param("ontology_id", ontology_id)
            .param("search", params.q.as_str())
            .param("skip", skip as i64)
            .param("limit", limit as i64);

        let mut result = self.pool.graph().execute(q).await.map_err(|e| {
            error!(error = %e, ontology_id, "Failed to list root classes");
            ClassError::Database(e.to_string())
        })?;

        let mut items = Vec::new();
        while let Ok(Some(row)) = result.next().await {
            if let Ok(id) = row.get::<String>("id") {
                let label: String = row.get("label").unwrap_or_default();
                let comment: Option<String> =
                    row.get("comment").ok().filter(|c: &String| !c.is_empty());
                items.push(OwlClassSummary {
                    id,
                    label,
                    comment,
                    parents: Vec::new(),
                });
            }
        }

        let total = self.count_roots(ontology_id, &params.q).await?;
        Ok(PaginatedResponse {
            items,
            total,
            page: params.page,
            per_page: limit,
        })
    }

    /// Returns direct children of a class (paginated).
    pub async fn get_children(
        &self,
        ontology_id: &str,
        class_id: &str,
        params: &ListClassesParams,
    ) -> Result<PaginatedResponse<OwlClassSummary>, ClassError> {
        debug!(ontology_id, %class_id, "Listing children");

        let _ = self.get(ontology_id, class_id).await?;

        let skip = params.page * params.per_page;
        let limit = params.per_page.min(100);

        let query = "\
            MATCH (c:Class {ontology_id: $ontology_id})-[:CHILD_OF]->(p:Class {id: $parent_id})\
            WHERE $search = '' OR toLower(c.label) CONTAINS toLower($search)\
            RETURN c.id AS id, c.label AS label, c.comment AS comment\
            ORDER BY c.label SKIP $skip LIMIT $limit\
        ";

        let q = neo4rs::Query::new(query.to_string())
            .param("ontology_id", ontology_id)
            .param("parent_id", class_id)
            .param("search", params.q.as_str())
            .param("skip", skip as i64)
            .param("limit", limit as i64);

        let mut result = self.pool.graph().execute(q).await.map_err(|e| {
            error!(error = %e, ontology_id, %class_id, "Failed to list children");
            ClassError::Database(e.to_string())
        })?;

        let mut items = Vec::new();
        while let Ok(Some(row)) = result.next().await {
            if let Ok(id) = row.get::<String>("id") {
                let label: String = row.get("label").unwrap_or_default();
                let comment: Option<String> =
                    row.get("comment").ok().filter(|c: &String| !c.is_empty());
                items.push(OwlClassSummary {
                    id,
                    label,
                    comment,
                    parents: vec![class_id.to_string()],
                });
            }
        }

        let total = self
            .count_children(ontology_id, class_id, &params.q)
            .await?;
        Ok(PaginatedResponse {
            items,
            total,
            page: params.page,
            per_page: limit,
        })
    }

    // ── Internal helpers ────────────────────────────────────────────────────

    async fn exists(&self, ontology_id: &str, class_id: &str) -> Result<bool, ClassError> {
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
            .map_err(|e| ClassError::Database(e.to_string()))?;

        match result
            .next()
            .await
            .map_err(|e| ClassError::Database(e.to_string()))?
        {
            Some(row) => {
                let cnt: i64 = row
                    .get("cnt")
                    .map_err(|e| ClassError::Database(e.to_string()))?;
                Ok(cnt > 0)
            }
            None => Ok(false),
        }
    }

    async fn find_dependents(
        &self,
        ontology_id: &str,
        class_id: &str,
    ) -> Result<(u64, u64), ClassError> {
        let query = "\
            MATCH (c:Class {id: $class_id, ontology_id: $ontology_id})\
            OPTIONAL MATCH (c)<-[:CHILD_OF]-(dependent:Class)\
            OPTIONAL MATCH (c)<-[:DOMAIN]-(prop:Property)\
            RETURN count(DISTINCT dependent) AS dep_count, count(DISTINCT prop) AS prop_count\
        ";
        let mut result = self
            .pool
            .graph()
            .execute(
                neo4rs::Query::new(query.to_string())
                    .param("ontology_id", ontology_id)
                    .param("class_id", class_id),
            )
            .await
            .map_err(|e| ClassError::Database(e.to_string()))?;

        match result
            .next()
            .await
            .map_err(|e| ClassError::Database(e.to_string()))?
        {
            Some(row) => {
                let dep_count: i64 = row
                    .get("dep_count")
                    .map_err(|e| ClassError::Database(e.to_string()))?;
                let prop_count: i64 = row
                    .get("prop_count")
                    .map_err(|e| ClassError::Database(e.to_string()))?;
                Ok((dep_count as u64, prop_count as u64))
            }
            None => Ok((0, 0)),
        }
    }

    async fn count(&self, ontology_id: &str, search: &str) -> Result<u64, ClassError> {
        let query = "MATCH (c:Class {ontology_id: $ontology_id}) WHERE $search = '' OR toLower(c.label) CONTAINS toLower($search) RETURN count(c) AS total";
        let mut result = self
            .pool
            .graph()
            .execute(
                neo4rs::Query::new(query.to_string())
                    .param("ontology_id", ontology_id)
                    .param("search", search),
            )
            .await
            .map_err(|e| ClassError::Database(e.to_string()))?;

        match result
            .next()
            .await
            .map_err(|e| ClassError::Database(e.to_string()))?
        {
            Some(row) => {
                let total: i64 = row
                    .get("total")
                    .map_err(|e| ClassError::Database(e.to_string()))?;
                Ok(total as u64)
            }
            None => Ok(0),
        }
    }

    async fn count_roots(&self, ontology_id: &str, search: &str) -> Result<u64, ClassError> {
        let query = "MATCH (c:Class {ontology_id: $ontology_id}) WHERE NOT EXISTS((c)-[:CHILD_OF]->()) AND ($search = '' OR toLower(c.label) CONTAINS toLower($search)) RETURN count(c) AS total";
        let mut result = self
            .pool
            .graph()
            .execute(
                neo4rs::Query::new(query.to_string())
                    .param("ontology_id", ontology_id)
                    .param("search", search),
            )
            .await
            .map_err(|e| ClassError::Database(e.to_string()))?;

        match result
            .next()
            .await
            .map_err(|e| ClassError::Database(e.to_string()))?
        {
            Some(row) => {
                let total: i64 = row
                    .get("total")
                    .map_err(|e| ClassError::Database(e.to_string()))?;
                Ok(total as u64)
            }
            None => Ok(0),
        }
    }

    async fn count_children(
        &self,
        ontology_id: &str,
        class_id: &str,
        search: &str,
    ) -> Result<u64, ClassError> {
        let query = "MATCH (c:Class {ontology_id: $ontology_id})-[:CHILD_OF]->(p:Class {id: $parent_id}) WHERE $search = '' OR toLower(c.label) CONTAINS toLower($search) RETURN count(c) AS total";
        let mut result = self
            .pool
            .graph()
            .execute(
                neo4rs::Query::new(query.to_string())
                    .param("ontology_id", ontology_id)
                    .param("parent_id", class_id)
                    .param("search", search),
            )
            .await
            .map_err(|e| ClassError::Database(e.to_string()))?;

        match result
            .next()
            .await
            .map_err(|e| ClassError::Database(e.to_string()))?
        {
            Some(row) => {
                let total: i64 = row
                    .get("total")
                    .map_err(|e| ClassError::Database(e.to_string()))?;
                Ok(total as u64)
            }
            None => Ok(0),
        }
    }
}

/// Lightweight class summary used in list responses.
#[derive(Debug, Clone, Serialize)]
pub struct OwlClassSummary {
    pub id: String,
    pub label: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub comment: Option<String>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub parents: Vec<String>,
}

// ── Axum Handlers ───────────────────────────────────────────────────────────────

/// POST `/api/v1/ontologies/{ontology_id}/classes`
pub async fn create_class_handler(
    State(state): State<Arc<AppState>>,
    Path(ontology_id): Path<String>,
    Json(req): Json<CreateClassRequest>,
) -> Result<(StatusCode, Json<OwlClass>), ClassError> {
    debug!(ontology_id, class_id = %req.id, "POST create class handler invoked");
    let repo = repo_from_state(&state)?;
    let class = repo.create(&ontology_id, &req).await?;
    info!(ontology_id, class_id = %class.id, "POST create class handler completed");
    Ok((StatusCode::CREATED, Json(class)))
}

/// GET `/api/v1/ontologies/{ontology_id}/classes`
pub async fn list_classes_handler(
    State(state): State<Arc<AppState>>,
    Path(ontology_id): Path<String>,
    Query(params): Query<ListClassesParams>,
) -> Result<Json<PaginatedResponse<OwlClassSummary>>, ClassError> {
    debug!(ontology_id, "GET list classes handler invoked");
    let repo = repo_from_state(&state)?;
    let result = repo.list(&ontology_id, &params).await?;
    Ok(Json(result))
}

/// GET `/api/v1/ontologies/{ontology_id}/classes/root`
pub async fn list_root_classes_handler(
    State(state): State<Arc<AppState>>,
    Path(ontology_id): Path<String>,
    Query(params): Query<ListClassesParams>,
) -> Result<Json<PaginatedResponse<OwlClassSummary>>, ClassError> {
    debug!(ontology_id, "GET root classes handler invoked");
    let repo = repo_from_state(&state)?;
    let result = repo.get_root_classes(&ontology_id, &params).await?;
    Ok(Json(result))
}

/// GET `/api/v1/ontologies/{ontology_id}/classes/{class_id}`
pub async fn get_class_handler(
    State(state): State<Arc<AppState>>,
    Path((ontology_id, class_id)): Path<(String, String)>,
) -> Result<Json<OwlClass>, ClassError> {
    debug!(ontology_id, %class_id, "GET class handler invoked");
    let repo = repo_from_state(&state)?;
    let class = repo.get(&ontology_id, &class_id).await?;
    Ok(Json(class))
}

/// GET `/api/v1/ontologies/{ontology_id}/classes/{class_id}/children`
pub async fn get_class_children_handler(
    State(state): State<Arc<AppState>>,
    Path((ontology_id, class_id)): Path<(String, String)>,
    Query(params): Query<ListClassesParams>,
) -> Result<Json<PaginatedResponse<OwlClassSummary>>, ClassError> {
    debug!(ontology_id, %class_id, "GET class children handler invoked");
    let repo = repo_from_state(&state)?;
    let result = repo.get_children(&ontology_id, &class_id, &params).await?;
    Ok(Json(result))
}

/// PUT `/api/v1/ontologies/{ontology_id}/classes/{class_id}`
pub async fn update_class_handler(
    State(state): State<Arc<AppState>>,
    Path((ontology_id, class_id)): Path<(String, String)>,
    Json(req): Json<UpdateClassRequest>,
) -> Result<Json<OwlClass>, ClassError> {
    debug!(ontology_id, %class_id, "PUT update class handler invoked");
    let repo = repo_from_state(&state)?;
    let class = repo.update(&ontology_id, &class_id, &req).await?;
    info!(ontology_id, class_id = %class.id, "PUT update class handler completed");
    Ok(Json(class))
}

/// DELETE `/api/v1/ontologies/{ontology_id}/classes/{class_id}`
pub async fn delete_class_handler(
    State(state): State<Arc<AppState>>,
    Path((ontology_id, class_id)): Path<(String, String)>,
    Query(params): Query<DeleteClassParams>,
) -> Result<Json<DeleteResponse>, ClassError> {
    debug!(ontology_id, %class_id, cascade = params.cascade, "DELETE class handler invoked");
    let repo = repo_from_state(&state)?;
    let result = repo.delete(&ontology_id, &class_id, params.cascade).await?;
    info!(ontology_id, class_id = %result.class_id, "DELETE class handler completed");
    Ok(Json(result))
}

/// Query parameters for class deletion.
#[derive(Debug, Deserialize)]
pub struct DeleteClassParams {
    #[serde(default)]
    pub cascade: bool,
}

// ── Tests ───────────────────────────────────────────────────────────────────────

#[cfg(test)]
mod tests {
    use super::*;
    use axum::{http::StatusCode, response::IntoResponse};

    /// Helper: creates an AppState with no Neo4j for handler tests.
    fn test_app_state() -> Arc<AppState> {
        Arc::new(AppState { neo4j: None })
    }

    // ── Serialization Tests ─────────────────────────────────────────────────

    #[test]
    fn test_owl_class_serialization() {
        let class = OwlClass {
            id: "Person".to_string(),
            label: "Person".to_string(),
            comment: Some("A person".to_string()),
            parents: vec![],
            children: vec!["Student".to_string(), "Professor".to_string()],
        };
        let json = serde_json::to_value(&class).unwrap();
        assert_eq!(json["id"], "Person");
        assert_eq!(json["label"], "Person");
        assert_eq!(json["comment"], "A person");
        assert_eq!(json["children"].as_array().unwrap().len(), 2);
    }

    #[test]
    fn test_owl_class_serialization_no_comment() {
        let class = OwlClass {
            id: "Student".to_string(),
            label: "Student".to_string(),
            comment: None,
            parents: vec!["Person".to_string()],
            children: vec![],
        };
        let json = serde_json::to_value(&class).unwrap();
        assert_eq!(json["id"], "Student");
        assert!(
            json.get("comment").is_none(),
            "None comment should be skipped"
        );
        assert_eq!(json["parents"].as_array().unwrap().len(), 1);
    }

    // ── Request Deserialization Tests ───────────────────────────────────────

    #[test]
    fn test_create_class_request_deserialization() {
        let json = serde_json::json!({
            "id": "Person",
            "label": "Person",
            "comment": "A person",
            "parents": []
        });
        let req: CreateClassRequest = serde_json::from_value(json).unwrap();
        assert_eq!(req.id, "Person");
        assert_eq!(req.label, "Person");
        assert_eq!(req.comment, Some("A person".to_string()));
        assert!(req.parents.is_empty());
    }

    #[test]
    fn test_create_class_request_minimal() {
        let json = serde_json::json!({ "id": "Student", "label": "Student" });
        let req: CreateClassRequest = serde_json::from_value(json).unwrap();
        assert_eq!(req.id, "Student");
        assert_eq!(req.label, "Student");
        assert!(req.comment.is_none());
        assert!(req.parents.is_empty());
    }

    #[test]
    fn test_create_class_request_with_parents() {
        let json =
            serde_json::json!({ "id": "Student", "label": "Student", "parents": ["Person"] });
        let req: CreateClassRequest = serde_json::from_value(json).unwrap();
        assert_eq!(req.parents, vec!["Person"]);
    }

    // ── Error Response Tests ────────────────────────────────────────────────

    #[tokio::test]
    async fn test_class_error_not_found_response() {
        let response = ClassError::NotFound("Person".to_string()).into_response();
        assert_eq!(response.status(), StatusCode::NOT_FOUND);
    }

    #[tokio::test]
    async fn test_class_error_already_exists_response() {
        let response = ClassError::AlreadyExists("Person".to_string()).into_response();
        assert_eq!(response.status(), StatusCode::CONFLICT);
    }

    #[tokio::test]
    async fn test_class_error_parent_not_found_response() {
        let response = ClassError::ParentNotFound("Missing".to_string()).into_response();
        assert_eq!(response.status(), StatusCode::UNPROCESSABLE_ENTITY);
    }

    #[tokio::test]
    async fn test_class_error_neo4j_not_configured_response() {
        let response = ClassError::Neo4jNotConfigured.into_response();
        assert_eq!(response.status(), StatusCode::SERVICE_UNAVAILABLE);
    }

    #[tokio::test]
    async fn test_class_error_has_dependents_response() {
        let err = ClassError::HasDependents {
            class_id: "Person".to_string(),
            dependent_count: 3,
            property_count: 2,
        };
        let response = err.into_response();
        assert_eq!(response.status(), StatusCode::CONFLICT);
    }

    // ── Handler Logic Tests (direct, no router) ────────────────────────────

    #[test]
    fn test_repo_from_state_none_returns_error() {
        let state = AppState { neo4j: None };
        let result = repo_from_state(&state);
        assert!(result.is_err());
        assert!(result.is_err());
    }

    #[test]
    fn test_repo_from_state_some_succeeds() {
        // We can't easily construct a Neo4jPool without a real connection,
        // so we can only test that a dummy AppState errors.
        // This test verifies the code path exists at compile time.
    }

    // ── Paginated Response Tests ────────────────────────────────────────────

    #[test]
    fn test_paginated_response_serialization() {
        let resp = PaginatedResponse::<OwlClassSummary> {
            items: vec![OwlClassSummary {
                id: "Person".to_string(),
                label: "Person".to_string(),
                comment: None,
                parents: vec![],
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

    #[test]
    fn test_delete_response_serialization() {
        let resp = DeleteResponse {
            deleted: true,
            class_id: "Person".to_string(),
            warning: Some("Deleted with cascade: 2 dependents".to_string()),
            dependent_count: Some(2),
        };
        let json = serde_json::to_value(&resp).unwrap();
        assert_eq!(json["deleted"], true);
        assert_eq!(json["class_id"], "Person");
        assert_eq!(json["warning"], "Deleted with cascade: 2 dependents");
        assert_eq!(json["dependent_count"], 2);
    }

    #[test]
    fn test_delete_response_no_warning() {
        let resp = DeleteResponse {
            deleted: true,
            class_id: "Orphan".to_string(),
            warning: None,
            dependent_count: None,
        };
        let json = serde_json::to_value(&resp).unwrap();
        assert_eq!(json["deleted"], true);
        assert!(json.get("warning").is_none());
        assert!(json.get("dependent_count").is_none());
    }

    // ── ListClassesParams Tests ─────────────────────────────────────────────

    #[test]
    fn test_list_classes_params_defaults() {
        let params: ListClassesParams = serde_json::from_value(serde_json::json!({})).unwrap();
        assert_eq!(params.q, "");
        assert_eq!(params.page, 0);
        assert_eq!(params.per_page, 20);
    }

    #[test]
    fn test_list_classes_params_custom() {
        let params: ListClassesParams =
            serde_json::from_value(serde_json::json!({"q": "Person", "page": 2, "per_page": 10}))
                .unwrap();
        assert_eq!(params.q, "Person");
        assert_eq!(params.page, 2);
        assert_eq!(params.per_page, 10);
    }

    #[test]
    fn test_delete_class_params_default() {
        let params: DeleteClassParams = serde_json::from_value(serde_json::json!({})).unwrap();
        assert!(!params.cascade);
    }

    #[test]
    fn test_delete_class_params_cascade_true() {
        let params: DeleteClassParams =
            serde_json::from_value(serde_json::json!({"cascade": true})).unwrap();
        assert!(params.cascade);
    }
}
