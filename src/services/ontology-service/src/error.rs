//! Domain-specific error types for the ontology service.
//!
//! Provides a unified `OntologyError` enum covering all error variants from
//! classes, properties, individuals, and export operations. Implements axum's
//! `IntoResponse` so errors can be returned directly from handlers.
//!
//! Error codes use the `ONT-*` domain prefix for consistency across the
//! VEDO Core platform, as specified in the architecture guidelines.

use axum::{http::StatusCode, response::IntoResponse, Json};
use thiserror::Error;
use uuid::Uuid;

/// All error variants that can occur in the ontology service.
///
/// Code prefix: `ONT-` followed by the domain (CLASS, PROPERTY, INDIVIDUAL)
/// and the specific error name.
#[derive(Debug, Error)]
pub enum OntologyError {
    // ── Class errors ────────────────────────────────────────────────────
    #[error("Class not found: {0}")]
    ClassNotFound(String),

    #[error("Class already exists: {0}")]
    ClassAlreadyExists(String),

    #[error("Parent class not found: {0}")]
    ParentNotFound(String),

    #[error("Class has {dependent_count} dependent classes and {property_count} referencing properties; delete with cascade=true to force")]
    ClassHasDependents {
        class_id: String,
        dependent_count: u64,
        property_count: u64,
    },

    #[error("Cannot cascade delete class: {0}")]
    ClassCascadeError(String),

    // ── Property errors ─────────────────────────────────────────────────
    #[error("Property not found: {0}")]
    PropertyNotFound(String),

    #[error("Property already exists: {0}")]
    PropertyAlreadyExists(String),

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
    PropertyHasDependents {
        property_id: String,
        dependent_count: u64,
    },

    #[error("Cannot cascade delete property: {0}")]
    PropertyCascadeError(String),

    // ── Individual (ABox) errors ─────────────────────────────────────────
    #[error("Individual not found: {0}")]
    IndividualNotFound(String),

    #[error("Individual already exists: {0}")]
    IndividualAlreadyExists(String),

    #[error("Class not found for individual: {0}")]
    IndividualClassNotFound(String),

    #[error("Property not found for individual: {0}")]
    IndividualPropertyNotFound(String),

    #[error("Target individual not found: {0}")]
    IndividualTargetNotFound(String),

    #[error("Individual has {0} incoming references; delete with cascade=true to force")]
    IndividualHasReferences(u64),

    #[error("Cannot cascade delete individual: {0}")]
    IndividualCascadeError(String),

    #[error("Invalid filter format: {0} (expected property_id:operator:value)")]
    InvalidFilter(String),

    // ── Shared errors ────────────────────────────────────────────────────
    #[error("Database error: {0}")]
    Database(String),

    #[error("Neo4j database is not configured")]
    Neo4jNotConfigured,
}

impl IntoResponse for OntologyError {
    fn into_response(self) -> axum::response::Response {
        let (status, code): (StatusCode, &str) = match &self {
            // Class errors
            OntologyError::ClassNotFound(_) => (StatusCode::NOT_FOUND, "ONT-CLASS-NOT-FOUND"),
            OntologyError::ClassAlreadyExists(_) => {
                (StatusCode::CONFLICT, "ONT-CLASS-ALREADY-EXISTS")
            }
            OntologyError::ParentNotFound(_) => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-CLASS-PARENT-NOT-FOUND",
            ),
            OntologyError::ClassHasDependents { .. } => {
                (StatusCode::CONFLICT, "ONT-CLASS-HAS-DEPENDENTS")
            }
            OntologyError::ClassCascadeError(_) => {
                (StatusCode::INTERNAL_SERVER_ERROR, "ONT-CLASS-CASCADE-ERROR")
            }
            // Property errors
            OntologyError::PropertyNotFound(_) => (StatusCode::NOT_FOUND, "ONT-PROPERTY-NOT-FOUND"),
            OntologyError::PropertyAlreadyExists(_) => {
                (StatusCode::CONFLICT, "ONT-PROPERTY-ALREADY-EXISTS")
            }
            OntologyError::DomainClassNotFound(_) => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-PROPERTY-DOMAIN-NOT-FOUND",
            ),
            OntologyError::RangeClassNotFound(_) => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-PROPERTY-RANGE-NOT-FOUND",
            ),
            OntologyError::InvalidXsdType(_) => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-PROPERTY-INVALID-XSD-TYPE",
            ),
            OntologyError::MissingRange => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-PROPERTY-MISSING-RANGE",
            ),
            OntologyError::MissingDomain => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-PROPERTY-MISSING-DOMAIN",
            ),
            OntologyError::AnnotationNotFound(_) => {
                (StatusCode::NOT_FOUND, "ONT-PROPERTY-ANNOTATION-NOT-FOUND")
            }
            OntologyError::PropertyTypeMismatch { .. } => {
                (StatusCode::CONFLICT, "ONT-PROPERTY-TYPE-MISMATCH")
            }
            OntologyError::PropertyHasDependents { .. } => {
                (StatusCode::CONFLICT, "ONT-PROPERTY-HAS-DEPENDENTS")
            }
            OntologyError::PropertyCascadeError(_) => (
                StatusCode::INTERNAL_SERVER_ERROR,
                "ONT-PROPERTY-CASCADE-ERROR",
            ),
            // Individual errors
            OntologyError::IndividualNotFound(_) => {
                (StatusCode::NOT_FOUND, "ONT-INDIVIDUAL-NOT-FOUND")
            }
            OntologyError::IndividualAlreadyExists(_) => {
                (StatusCode::CONFLICT, "ONT-INDIVIDUAL-ALREADY-EXISTS")
            }
            OntologyError::IndividualClassNotFound(_) => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-INDIVIDUAL-CLASS-NOT-FOUND",
            ),
            OntologyError::IndividualPropertyNotFound(_) => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-INDIVIDUAL-PROPERTY-NOT-FOUND",
            ),
            OntologyError::IndividualTargetNotFound(_) => (
                StatusCode::UNPROCESSABLE_ENTITY,
                "ONT-INDIVIDUAL-TARGET-NOT-FOUND",
            ),
            OntologyError::IndividualHasReferences(_) => {
                (StatusCode::CONFLICT, "ONT-INDIVIDUAL-HAS-REFERENCES")
            }
            OntologyError::IndividualCascadeError(_) => (
                StatusCode::INTERNAL_SERVER_ERROR,
                "ONT-INDIVIDUAL-CASCADE-ERROR",
            ),
            OntologyError::InvalidFilter(_) => {
                (StatusCode::BAD_REQUEST, "ONT-INDIVIDUAL-INVALID-FILTER")
            }
            // Shared errors
            OntologyError::Database(_) => (StatusCode::INTERNAL_SERVER_ERROR, "ONT-DATABASE-ERROR"),
            OntologyError::Neo4jNotConfigured => {
                (StatusCode::SERVICE_UNAVAILABLE, "ONT-NEO4J-NOT-CONFIGURED")
            }
        };
        let detail = match &self {
            OntologyError::Database(msg) => {
                let trace_id = Uuid::new_v4().to_string();
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

#[cfg(test)]
mod tests {
    use super::*;
    use axum::http::StatusCode;

    #[tokio::test]
    async fn test_class_not_found_response() {
        let err = OntologyError::ClassNotFound("Person".to_string());
        let resp = err.into_response();
        assert_eq!(resp.status(), StatusCode::NOT_FOUND);
    }

    #[tokio::test]
    async fn test_class_already_exists_response() {
        let err = OntologyError::ClassAlreadyExists("Person".to_string());
        let resp = err.into_response();
        assert_eq!(resp.status(), StatusCode::CONFLICT);
    }

    #[tokio::test]
    async fn test_property_not_found_response() {
        let err = OntologyError::PropertyNotFound("hasAge".to_string());
        let resp = err.into_response();
        assert_eq!(resp.status(), StatusCode::NOT_FOUND);
    }

    #[tokio::test]
    async fn test_individual_not_found_response() {
        let err = OntologyError::IndividualNotFound("i-42".to_string());
        let resp = err.into_response();
        assert_eq!(resp.status(), StatusCode::NOT_FOUND);
    }

    #[tokio::test]
    async fn test_neo4j_not_configured_response() {
        let err = OntologyError::Neo4jNotConfigured;
        let resp = err.into_response();
        assert_eq!(resp.status(), StatusCode::SERVICE_UNAVAILABLE);
    }

    #[tokio::test]
    async fn test_database_error_response() {
        let err = OntologyError::Database("connection refused".to_string());
        let resp = err.into_response();
        assert_eq!(resp.status(), StatusCode::INTERNAL_SERVER_ERROR);
    }

    #[tokio::test]
    async fn test_database_error_sanitized() {
        // The raw error message must NOT leak into the HTTP response.
        let err = OntologyError::Database("connection refused".to_string());
        let resp = err.into_response();
        let body = axum::body::to_bytes(resp.into_body(), usize::MAX)
            .await
            .unwrap();
        let json: serde_json::Value = serde_json::from_slice(&body).unwrap();

        assert_eq!(json["error"], "ONT-DATABASE-ERROR");
        let detail = json["detail"].as_str().unwrap();
        assert!(
            detail.starts_with("Internal database error (trace_id:"),
            "detail should be generic, got: {detail}",
        );
        assert!(
            !detail.contains("connection refused"),
            "raw error message leaked: {detail}",
        );
    }

    #[test]
    fn test_error_messages() {
        assert_eq!(
            OntologyError::ClassNotFound("Thing".to_string()).to_string(),
            "Class not found: Thing"
        );
        assert_eq!(
            OntologyError::InvalidXsdType("hex".to_string()).to_string(),
            "Invalid XSD type: hex (valid values: string, integer, boolean, date, float)"
        );
        assert_eq!(
            OntologyError::IndividualHasReferences(3).to_string(),
            "Individual has 3 incoming references; delete with cascade=true to force"
        );
    }
}
