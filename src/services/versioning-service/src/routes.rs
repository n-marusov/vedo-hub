//! Route definitions for the versioning service API.
//!
//! All versioning-specific endpoints are grouped under `/api/v1/versioning/`.

use std::sync::Arc;

use axum::{
    routing::{get, post},
    Router,
};

use crate::handlers::{
    checkout_commit_handler, create_branch_handler, create_commit_handler, delete_branch_handler,
    get_branch_handler, get_commit_delta_handler, get_commit_handler, list_branches_handler,
    list_commits_handler, merge_branches_handler, rollback_commit_handler, switch_branch_handler,
};
use crate::AppState;

/// Builds the API routes for the versioning service.
///
/// # Endpoints
///
/// | Method | Path | Description |
/// |--------|------|-------------|
/// | POST   | `/api/v1/versioning/commits` | Create a new commit |
/// | GET    | `/api/v1/versioning/commits` | List commits (paginated) |
/// | GET    | `/api/v1/versioning/commits/:id` | Get commit details |
/// | GET    | `/api/v1/versioning/commits/:id/delta` | Get commit delta |
/// | POST   | `/api/v1/versioning/commits/:id/checkout` | Checkout a commit |
/// | POST   | `/api/v1/versioning/commits/:id/rollback` | Rollback to a commit |
/// | POST   | `/api/v1/versioning/branches` | Create a new branch |
/// | GET    | `/api/v1/versioning/branches` | List branches (paginated) |
/// | GET    | `/api/v1/versioning/branches/:id` | Get branch details |
/// | DELETE | `/api/v1/versioning/branches/:id` | Delete a branch |
/// | POST   | `/api/v1/versioning/branches/:id/switch` | Switch to a branch |
/// | POST   | `/api/v1/versioning/branches/merge` | Merge branches |
///
/// NOTE: axum 0.7 (matchit 0.7) uses the `:param` syntax; the `{param}` form
/// is treated as a literal path segment and returns 404 for any value.
pub fn build_routes() -> Router<Arc<AppState>> {
    let commits_routes = Router::new()
        .route("/", post(create_commit_handler).get(list_commits_handler))
        .route("/:id", get(get_commit_handler))
        .route("/:id/delta", get(get_commit_delta_handler))
        .route("/:id/checkout", post(checkout_commit_handler))
        .route("/:id/rollback", post(rollback_commit_handler));

    let branches_routes = Router::new()
        .route("/", post(create_branch_handler).get(list_branches_handler))
        .route(
            "/:id",
            get(get_branch_handler).delete(delete_branch_handler),
        )
        .route("/:id/switch", post(switch_branch_handler))
        .route("/merge", post(merge_branches_handler));

    Router::new()
        .nest("/api/v1/versioning/commits", commits_routes)
        .nest("/api/v1/versioning/branches", branches_routes)
}

#[cfg(test)]
mod tests {
    use super::*;
    use axum::{
        body::Body,
        http::{Request, StatusCode},
    };
    use tower::ServiceExt;

    fn test_state() -> Arc<AppState> {
        Arc::new(AppState { pg: None })
    }

    #[tokio::test]
    async fn test_unknown_route_returns_404() {
        let app = build_routes().with_state(test_state());
        let response = app
            .oneshot(
                Request::builder()
                    .uri("/api/v1/versioning/nonexistent")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(response.status(), StatusCode::NOT_FOUND);
    }

    #[tokio::test]
    async fn test_commits_list_route_registered() {
        let app = build_routes().with_state(test_state());
        let response = app
            .oneshot(
                Request::builder()
                    .uri("/api/v1/versioning/commits")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        // No DB - expect SERVICE_UNAVAILABLE (route exists)
        assert_eq!(response.status(), StatusCode::SERVICE_UNAVAILABLE);
    }
}
