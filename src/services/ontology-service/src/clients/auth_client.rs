//! HTTP client for auth-service organizational model REST API.
//!
//! Provides methods to list groups, projects, and members via the auth-service
//! REST endpoints. Returns domain types that GraphQL resolvers can use.

use std::time::Duration;

use reqwest::{Client, StatusCode};
use serde::{Deserialize, Serialize};
use thiserror::Error;

/// Errors that can occur when communicating with the auth-service.
#[derive(Error, Debug)]
pub enum AuthClientError {
    #[error("HTTP request failed: {0}")]
    Http(#[from] reqwest::Error),

    #[error("Auth service returned {status}: {body}")]
    Service { status: StatusCode, body: String },

    #[error("Configuration error: {0}")]
    Config(String),
}

/// A group node in the organization hierarchy.
#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Group {
    pub id: String,
    #[serde(rename = "type")]
    pub type_: String,
    pub name: Option<String>,
    pub description: Option<String>,
    pub parent_id: Option<String>,
    pub visibility: Option<String>,
    pub tenant_id: Option<String>,
}

/// A project (ontology) node in the organization.
#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Project {
    pub id: String,
    #[serde(rename = "type")]
    pub type_: String,
    pub name: Option<String>,
    pub description: Option<String>,
    pub visibility: Option<String>,
    pub tenant_id: Option<String>,
}

/// A member with role assignment.
#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Member {
    pub user_id: String,
    pub scope: String,
    pub role: String,
    pub inherited: Option<bool>,
}

/// Client for the auth-service REST API.
#[derive(Debug, Clone)]
pub struct AuthClient {
    client: Client,
    base_url: String,
    timeout_read: Duration,
    timeout_write: Duration,
}

impl AuthClient {
    /// Creates a new `AuthClient` with the given base URL.
    ///
    /// If `base_url` is empty, uses the `AUTH_SERVICE_URL` env var or
    /// defaults to `http://auth-service:8081`.
    pub fn new(base_url: Option<String>) -> Self {
        let url = base_url
            .filter(|u| !u.is_empty())
            .or_else(|| std::env::var("AUTH_SERVICE_URL").ok())
            .unwrap_or_else(|| "http://auth-service:8081".to_string());

        Self {
            client: Client::new(),
            base_url: url,
            timeout_read: Duration::from_secs(10),
            timeout_write: Duration::from_secs(15),
        }
    }

    /// Returns the base URL of the auth-service.
    pub fn base_url(&self) -> &str {
        &self.base_url
    }

    // ── Groups ────────────────────────────────────────────────────────────

    /// Lists all groups from the auth-service.
    pub async fn list_groups(&self) -> Result<Vec<Group>, AuthClientError> {
        let resp = self
            .client
            .get(format!("{}/api/v1/groups", self.base_url))
            .timeout(self.timeout_read)
            .send()
            .await?;

        Self::check_status(&resp)?;

        #[derive(Deserialize)]
        struct GroupsResponse {
            data: Vec<Group>,
        }

        let body = resp.json::<GroupsResponse>().await?;
        Ok(body.data)
    }

    /// Lists all projects from the auth-service.
    pub async fn list_projects(
        &self,
        page: Option<i32>,
        per_page: Option<i32>,
    ) -> Result<Vec<Project>, AuthClientError> {
        let mut url = format!("{}/api/v1/projects", self.base_url);
        let mut params = vec![];
        if let Some(p) = page {
            params.push(format!("page={}", p));
        }
        if let Some(pp) = per_page {
            params.push(format!("perPage={}", pp));
        }
        if !params.is_empty() {
            url.push('?');
            url.push_str(&params.join("&"));
        }

        let resp = self
            .client
            .get(&url)
            .timeout(self.timeout_read)
            .send()
            .await?;

        Self::check_status(&resp)?;

        #[derive(Deserialize)]
        struct ProjectsResponse {
            data: Vec<Project>,
        }

        let body = resp.json::<ProjectsResponse>().await?;
        Ok(body.data)
    }

    /// Lists members for a given scope (ontology or group).
    pub async fn list_members(&self, scope: &str) -> Result<Vec<Member>, AuthClientError> {
        // Convert "ontology/ont-123" to "/api/v1/ontologies/ont-123/members"
        let api_path = scope_to_api_path(scope);
        let url = format!("{}{}", self.base_url, api_path);

        let resp = self
            .client
            .get(&url)
            .timeout(self.timeout_read)
            .send()
            .await?;

        Self::check_status(&resp)?;

        #[derive(Deserialize)]
        struct MembersResponse {
            data: Vec<Member>,
        }

        let body = resp.json::<MembersResponse>().await?;
        Ok(body.data)
    }

    /// Updates a member's role.
    pub async fn update_member_role(
        &self,
        scope: &str,
        user_id: &str,
        role: &str,
    ) -> Result<Member, AuthClientError> {
        let api_path = scope_to_api_path(scope);
        let url = format!("{}{}/{}", self.base_url, api_path, user_id);

        #[derive(Serialize)]
        struct RoleUpdate {
            role: String,
        }

        let resp = self
            .client
            .put(&url)
            .json(&RoleUpdate {
                role: role.to_string(),
            })
            .timeout(self.timeout_write)
            .send()
            .await?;

        Self::check_status(&resp)?;
        let member = resp.json::<Member>().await?;
        Ok(member)
    }

    /// Removes a member from a scope.
    pub async fn remove_member(&self, scope: &str, user_id: &str) -> Result<(), AuthClientError> {
        let api_path = scope_to_api_path(scope);
        let url = format!("{}{}/{}", self.base_url, api_path, user_id);

        let resp = self
            .client
            .delete(&url)
            .timeout(self.timeout_write)
            .send()
            .await?;

        Self::check_status(&resp)?;
        Ok(())
    }

    // ── Helpers ───────────────────────────────────────────────────────────

    fn check_status(resp: &reqwest::Response) -> Result<(), AuthClientError> {
        let status = resp.status();
        if status.is_success() {
            return Ok(());
        }
        Err(AuthClientError::Service {
            status,
            body: "upstream error".to_string(),
        })
    }
}

/// Converts a scope string like "ontology/ont-123" to an API path
/// like "/api/v1/ontologies/ont-123/members".
fn scope_to_api_path(scope: &str) -> String {
    if let Some(rest) = scope.strip_prefix("ontology/") {
        format!("/api/v1/ontologies/{}/members", rest)
    } else if let Some(rest) = scope.strip_prefix("group/") {
        format!("/api/v1/groups/{}/members", rest)
    } else {
        // Fallback: assume it's an ontology ID
        format!("/api/v1/ontologies/{}/members", scope)
    }
}
