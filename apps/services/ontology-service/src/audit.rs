//! Structured audit trail for query execution (SPARQL/CYPHER).
//!
//! Implements vision F3.878: every read-only query execution produces an
//! audit record with trace_id, user, query, LIMIT, and execution time. The
//! record is emitted as a structured JSON log line via `tracing::info` with
//! the `query.audit` event name — consumers (Loki, M6 MCP audit) can filter
//! on it without a dedicated table.

use axum::http::HeaderMap;
use tracing::info;

/// Identity headers forwarded by the API Gateway. The gateway injects the
/// authenticated user (from the JWT) and the trace id; when absent (direct
/// calls) the fields fall back to empty strings.
const HDR_TRACE_ID: &str = "x-trace-id";
const HDR_USER_ID: &str = "x-user-id";
const HDR_FORWARDED_USER: &str = "x-forwarded-user";

/// Outcome of a query audit.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum AuditStatus {
    /// Query executed successfully.
    Succeeded,
    /// Query rejected before execution (read-only violation, unsupported form,
    /// DB error).
    Rejected,
}

impl AuditStatus {
    pub fn as_str(self) -> &'static str {
        match self {
            AuditStatus::Succeeded => "succeeded",
            AuditStatus::Rejected => "rejected",
        }
    }
}

/// One audit record for a query execution.
pub struct AuditEntry {
    pub trace_id: String,
    pub user_id: String,
    pub dialect: &'static str,
    pub query: String,
    pub limit: Option<usize>,
    pub execution_time_ms: u64,
    pub status: AuditStatus,
}

impl AuditEntry {
    /// Emits the audit record as a structured JSON log line.
    pub fn emit(&self) {
        info!(
            event = "query.audit",
            trace_id = %self.trace_id,
            user_id = %self.user_id,
            dialect = self.dialect,
            query = %self.query,
            limit = self.limit.map(|l| l.to_string()).unwrap_or_default(),
            execution_time_ms = self.execution_time_ms,
            status = self.status.as_str(),
            "query audit record",
        );
    }
}

/// Extracts the trace id from the request headers, if present.
pub fn trace_id_from_headers(headers: &HeaderMap) -> String {
    headers
        .get(HDR_TRACE_ID)
        .and_then(|v| v.to_str().ok())
        .unwrap_or_default()
        .to_string()
}

/// Extracts the authenticated user id from the request headers, if present.
/// The gateway forwards the JWT subject as `x-user-id` (fallback:
/// `x-forwarded-user`).
pub fn user_id_from_headers(headers: &HeaderMap) -> String {
    for name in [HDR_USER_ID, HDR_FORWARDED_USER] {
        if let Some(v) = headers.get(name).and_then(|v| v.to_str().ok()) {
            if !v.is_empty() {
                return v.to_string();
            }
        }
    }
    String::new()
}

#[cfg(test)]
mod tests {
    use super::*;

    fn headers_with(values: &[(&str, &str)]) -> HeaderMap {
        let mut map = HeaderMap::new();
        for (k, v) in values {
            map.append(
                axum::http::header::HeaderName::from_bytes(k.as_bytes()).unwrap(),
                v.parse().unwrap(),
            );
        }
        map
    }

    #[test]
    fn test_trace_id_from_headers_returns_value() {
        let headers = headers_with(&[("x-trace-id", "trace-abc")]);
        assert_eq!(trace_id_from_headers(&headers), "trace-abc");
    }

    #[test]
    fn test_trace_id_from_headers_missing_returns_empty() {
        let headers = HeaderMap::new();
        assert_eq!(trace_id_from_headers(&headers), "");
    }

    #[test]
    fn test_user_id_from_headers_prefers_x_user_id() {
        let headers = headers_with(&[("x-user-id", "user-42"), ("x-forwarded-user", "other")]);
        assert_eq!(user_id_from_headers(&headers), "user-42");
    }

    #[test]
    fn test_user_id_from_headers_falls_back_to_forwarded() {
        let headers = headers_with(&[("x-forwarded-user", "user-7")]);
        assert_eq!(user_id_from_headers(&headers), "user-7");
    }

    #[test]
    fn test_user_id_from_headers_missing_returns_empty() {
        let headers = HeaderMap::new();
        assert_eq!(user_id_from_headers(&headers), "");
    }

    #[test]
    fn test_audit_status_str() {
        assert_eq!(AuditStatus::Succeeded.as_str(), "succeeded");
        assert_eq!(AuditStatus::Rejected.as_str(), "rejected");
    }
}
