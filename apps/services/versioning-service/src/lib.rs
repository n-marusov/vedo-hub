//! Application library for the versioning service.
//!
//! Provides the `build_app` function that assembles all routes, middleware,
//! and shared state into a single axum `Router`.

pub mod error;
pub mod handlers;
pub mod models;
pub mod postgres;
pub mod repositories;
pub mod routes;
pub mod services;

use std::sync::Arc;

use axum::{extract::State, routing::get, Json, Router};
use serde::Serialize;
use tower_http::trace::TraceLayer;

pub const SERVICE_NAME: &str = "versioning-service";
pub const DEFAULT_PORT: &str = "8083";

/// Application state shared across handlers.
#[derive(Clone)]
pub struct AppState {
    pub pg: Option<postgres::PgPoolWrapper>,
}

#[derive(Serialize)]
struct RootResponse {
    name: &'static str,
    version: &'static str,
    description: &'static str,
}

async fn root_handler() -> Json<RootResponse> {
    tracing::debug!("Root endpoint requested");
    Json(RootResponse {
        name: SERVICE_NAME,
        version: "0.2.0",
        description: "Versioning service — Git-like commits, branches, diffs",
    })
}

async fn pg_health_handler(State(state): State<Arc<AppState>>) -> Json<serde_json::Value> {
    match &state.pg {
        Some(pool) => {
            let healthy = postgres::health_check(pool).await;
            Json(serde_json::json!({
                "status": if healthy { "healthy" } else { "degraded" },
                "database": "postgresql",
                "connected": healthy,
            }))
        }
        None => Json(serde_json::json!({
            "status": "disabled",
            "database": "postgresql",
            "connected": false,
            "message": "PostgreSQL not configured",
        })),
    }
}

/// Builds the complete axum application with all routes, middleware,
/// and shared state.
///
/// # Arguments
///
/// * `state` - The shared application state (wraps `PostgreSQL` pool).
///
/// # Returns
///
/// A configured `Router` ready to serve.
pub fn build_app(state: Arc<AppState>) -> Router {
    // Build stateless routes (root, shared health)
    let app = Router::new()
        .route("/", get(root_handler))
        .layer(TraceLayer::new_for_http());

    // Merge shared health, ready, and metrics routes
    let app = vedo_shared::add_health_routes(app, SERVICE_NAME);
    let app = app.route("/metrics", get(vedo_shared::metrics_handler));

    // Merge DB health route with state
    let app = app.merge(
        Router::new()
            .route("/health/db", get(pg_health_handler))
            .with_state(state.clone()),
    );

    // Merge versioning API routes with state

    app.merge(routes::build_routes().with_state(state))
}

/// Initializes the `PostgreSQL` connection pool from environment config.
pub async fn init_pg_pool() -> Option<postgres::PgPoolWrapper> {
    let pg_config = postgres::PgConfig::from_env();
    match postgres::create_pool(&pg_config).await {
        Ok(pool) => {
            match postgres::run_manual_migrations(pool.pool()).await {
                Ok(()) => tracing::info!("PostgreSQL migrations applied successfully"),
                Err(e) => tracing::warn!(error = %e, "Migration warning, continuing"),
            }
            tracing::info!("PostgreSQL pool initialized successfully");
            Some(pool)
        }
        Err(e) => {
            tracing::warn!(error = %e, "Failed to connect to PostgreSQL, running without database");
            None
        }
    }
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
    async fn test_health_no_db() {
        let app = build_app(test_state());
        let response = app
            .oneshot(
                Request::builder()
                    .uri("/health")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(response.status(), StatusCode::OK);
    }

    #[tokio::test]
    async fn test_db_health_degraded_when_no_db() {
        let app = build_app(test_state());
        let response = app
            .oneshot(
                Request::builder()
                    .uri("/health/db")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(response.status(), StatusCode::OK);
        let body: serde_json::Value = serde_json::from_slice(
            &axum::body::to_bytes(response.into_body(), usize::MAX)
                .await
                .unwrap(),
        )
        .unwrap();
        assert_eq!(body["status"], "disabled");
    }

    #[tokio::test]
    async fn test_root_returns_service_info() {
        let app = build_app(test_state());
        let response = app
            .oneshot(Request::builder().uri("/").body(Body::empty()).unwrap())
            .await
            .unwrap();
        assert_eq!(response.status(), StatusCode::OK);
        let body: serde_json::Value = serde_json::from_slice(
            &axum::body::to_bytes(response.into_body(), usize::MAX)
                .await
                .unwrap(),
        )
        .unwrap();
        assert_eq!(body["name"], SERVICE_NAME);
    }

    #[tokio::test]
    async fn test_unknown_route_returns_404() {
        let app = build_app(test_state());
        let response = app
            .oneshot(
                Request::builder()
                    .uri("/unknown")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(response.status(), StatusCode::NOT_FOUND);
    }
}
