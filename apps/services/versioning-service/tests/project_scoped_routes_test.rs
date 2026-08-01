//! Project-scoped versioning route tests (RED phase — TDD).
//!
//! Validates: REQ-FUN.API.rest-gitlab-alignment
//!
//! Per ADR-DES.API.rest-gitlab-alignment, the versioning surface must be
//! project-scoped and GitLab-aligned:
//!
//!   `/api/v1/versioning/commits|branches`  →  `/api/v1/projects/{{pid}}/repository/commits|branches`
//!
//! Target mapping (plan §4 "Versioning"):
//!   - GET/POST  /projects/{{pid}}/repository/commits              (list/create; create = VEDO extension)
//!   - GET       /projects/{{pid}}/repository/commits/{{sha}}        (id → sha rename)
//!   - GET       /projects/{{pid}}/repository/commits/{{sha}}/diff   (delta → diff rename)
//!   - GET/POST  /projects/{{pid}}/repository/branches             (list/create)
//!   - GET/DELETE /projects/{{pid}}/repository/branches/{{name}}     (id → name rename)
//!   - checkout → NOT a REST endpoint (internal-only per F3)
//!   - switch   → REMOVED from REST
//!   - merge    → via MR workflow (501 planned stub until M10)
//!
//! RED state: these tests FAIL against the current code (only the legacy
//! `/api/v1/versioning/*` routes exist). They pass after Tasks 25-27.
//!
//! Tests use `pg: None` (no DB required): a successful route match returns
//! 503 (Service Unavailable), 404 means the route is not registered.

mod common;

use std::sync::Arc;

use axum::{
    body::Body,
    http::{Method, Request, StatusCode},
};
use tower::ServiceExt;
use versioning_service::{routes::build_routes, AppState};

fn app_no_db() -> axum::Router {
    let state = Arc::new(AppState {
        pg: None,
        branch_repo: None,
        commit_repo: None,
    });
    build_routes().with_state(state)
}

async fn request_status(app: axum::Router, method: Method, uri: &str, body: Body) -> StatusCode {
    app.oneshot(
        Request::builder()
            .method(method)
            .uri(uri)
            .header("Content-Type", "application/json")
            .body(body)
            .unwrap(),
    )
    .await
    .unwrap()
    .status()
}

// ---------------------------------------------------------------------------
// Positive contract tests: project-scoped routes MUST exist
// ---------------------------------------------------------------------------

// TestProjectCommits_List_ShouldResolve verifies
// GET /api/v1/projects/{{pid}}/repository/commits is registered.
#[tokio::test]
async fn test_project_commits_list_route_resolves() {
    let status = request_status(
        app_no_db(),
        Method::GET,
        "/api/v1/projects/p1/repository/commits",
        Body::empty(),
    )
    .await;
    assert_ne!(
        status,
        StatusCode::NOT_FOUND,
        "GET /projects/{{pid}}/repository/commits resolved to 404 — route not registered"
    );
}

// TestProjectCommits_Create_ShouldResolve verifies
// POST /api/v1/projects/{{pid}}/repository/commits is registered (VEDO extension).
#[tokio::test]
async fn test_project_commits_create_route_resolves() {
    let status = request_status(
        app_no_db(),
        Method::POST,
        "/api/v1/projects/p1/repository/commits",
        Body::from(r#"{"message":"m","parent_id":null,"author_id":"a"}"#),
    )
    .await;
    assert_ne!(
        status,
        StatusCode::NOT_FOUND,
        "POST /projects/{{pid}}/repository/commits resolved to 404 — route not registered"
    );
}

// TestProjectCommit_GetBySha_ShouldResolve verifies
// GET /api/v1/projects/{{pid}}/repository/commits/{{sha}} is registered (id → sha).
#[tokio::test]
async fn test_project_commit_get_by_sha_route_resolves() {
    let status = request_status(
        app_no_db(),
        Method::GET,
        "/api/v1/projects/p1/repository/commits/abc123",
        Body::empty(),
    )
    .await;
    assert_ne!(
        status,
        StatusCode::NOT_FOUND,
        "GET /projects/{{pid}}/repository/commits/{{sha}} resolved to 404 — route not registered"
    );
}

// TestProjectCommit_Diff_ShouldResolve verifies
// GET /api/v1/projects/{{pid}}/repository/commits/{{sha}}/diff is registered
// (delta → diff rename).
#[tokio::test]
async fn test_project_commit_diff_route_resolves() {
    let status = request_status(
        app_no_db(),
        Method::GET,
        "/api/v1/projects/p1/repository/commits/abc123/diff",
        Body::empty(),
    )
    .await;
    assert_ne!(
        status,
        StatusCode::NOT_FOUND,
        "GET /projects/{{pid}}/repository/commits/{{sha}}/diff resolved to 404 — route not registered"
    );
}

// TestProjectBranches_List_ShouldResolve verifies
// GET /api/v1/projects/{{pid}}/repository/branches is registered.
#[tokio::test]
async fn test_project_branches_list_route_resolves() {
    let status = request_status(
        app_no_db(),
        Method::GET,
        "/api/v1/projects/p1/repository/branches",
        Body::empty(),
    )
    .await;
    assert_ne!(
        status,
        StatusCode::NOT_FOUND,
        "GET /projects/{{pid}}/repository/branches resolved to 404 — route not registered"
    );
}

// TestProjectBranches_Create_ShouldResolve verifies
// POST /api/v1/projects/{{pid}}/repository/branches is registered.
#[tokio::test]
async fn test_project_branches_create_route_resolves() {
    let status = request_status(
        app_no_db(),
        Method::POST,
        "/api/v1/projects/p1/repository/branches",
        Body::from(r#"{"name":"feat","from_commit":"abc"}"#),
    )
    .await;
    assert_ne!(
        status,
        StatusCode::NOT_FOUND,
        "POST /projects/{{pid}}/repository/branches resolved to 404 — route not registered"
    );
}

// TestProjectBranch_GetByName_ShouldResolve verifies
// GET /api/v1/projects/{{pid}}/repository/branches/{{name}} is registered
// (id → name rename).
#[tokio::test]
async fn test_project_branch_get_by_name_route_resolves() {
    let status = request_status(
        app_no_db(),
        Method::GET,
        "/api/v1/projects/p1/repository/branches/main",
        Body::empty(),
    )
    .await;
    assert_ne!(
        status,
        StatusCode::NOT_FOUND,
        "GET /projects/{{pid}}/repository/branches/{{name}} resolved to 404 — route not registered"
    );
}

// TestProjectBranch_DeleteByName_ShouldResolve verifies
// DELETE /api/v1/projects/{{pid}}/repository/branches/{{name}} is registered.
#[tokio::test]
async fn test_project_branch_delete_by_name_route_resolves() {
    let status = request_status(
        app_no_db(),
        Method::DELETE,
        "/api/v1/projects/p1/repository/branches/old-feat",
        Body::empty(),
    )
    .await;
    assert_ne!(
        status,
        StatusCode::NOT_FOUND,
        "DELETE /projects/{{pid}}/repository/branches/{{name}} resolved to 404 — route not registered"
    );
}

// ---------------------------------------------------------------------------
// Negative contract tests: checkout/switch MUST NOT be REST-accessible
// ---------------------------------------------------------------------------

// TestProjectCommit_Checkout_ShouldNotBeREST verifies checkout is NOT
// exposed via the project-scoped REST surface (internal-only per F3).
#[tokio::test]
async fn test_project_commit_checkout_not_rest_accessible() {
    let status = request_status(
        app_no_db(),
        Method::POST,
        "/api/v1/projects/p1/repository/commits/abc123/checkout",
        Body::empty(),
    )
    .await;
    assert_eq!(
        status,
        StatusCode::NOT_FOUND,
        "checkout must not be a REST endpoint (internal-only per F3)"
    );
}

// TestProjectBranch_Switch_ShouldNotBeREST verifies switch is removed from
// the REST surface.
#[tokio::test]
async fn test_project_branch_switch_not_rest_accessible() {
    let status = request_status(
        app_no_db(),
        Method::POST,
        "/api/v1/projects/p1/repository/branches/main/switch",
        Body::empty(),
    )
    .await;
    assert_eq!(
        status,
        StatusCode::NOT_FOUND,
        "switch must be removed from REST"
    );
}

// ---------------------------------------------------------------------------
// Legacy-route contract: old paths stay active with deprecation until
// consumers migrate (M10). They must still resolve (503, not 404).
// ---------------------------------------------------------------------------

// TestLegacyVersioningRoutes_StillResolve verifies the legacy
// /api/v1/versioning/* surface remains registered during migration.
#[tokio::test]
async fn test_legacy_versioning_routes_still_resolve() {
    for (method, uri, body) in [
        (Method::GET, "/api/v1/versioning/commits", Body::empty()),
        (Method::GET, "/api/v1/versioning/branches", Body::empty()),
    ] {
        let status = request_status(app_no_db(), method, uri, body).await;
        assert_ne!(
            status,
            StatusCode::NOT_FOUND,
            "legacy {uri} resolved to 404 — legacy routes must stay active during migration"
        );
    }
}

// Suppress unused-import warning when the `common` module's helpers are not
// directly referenced.
#[allow(dead_code)]
fn _silence_common_module() {
    let _ = common::is_integration_enabled();
}
