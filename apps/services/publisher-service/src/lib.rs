//! Application library for the publisher service.
//!
//! Provides the `build_app` function that assembles all routes, middleware,
//! and shared state into a single axum `Router`.

pub mod handlers;
pub mod models;
pub mod storage;

use std::sync::Arc;

use axum::{
    routing::{delete, get, post},
    Json, Router,
};
use serde::Serialize;
use tower_http::trace::TraceLayer;

pub const SERVICE_NAME: &str = "publisher-service";
pub const DEFAULT_PORT: &str = "8086";

/// Application state shared across handlers.
#[derive(Clone)]
pub struct AppState {
    pub store: storage::SnapshotStore,
    pub ontology_service_url: String,
    pub http_client: reqwest::Client,
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
        description: "Publisher service — ontology snapshot publishing",
    })
}

/// Builds the complete axum application with all routes, middleware,
/// and shared state.
///
/// # Arguments
///
/// * `state` - The shared application state (wraps snapshot store, HTTP client).
///
/// # Returns
///
/// A configured `Router` ready to serve.
pub fn build_app(state: Arc<AppState>) -> Router {
    // Build stateless routes
    let app = Router::new()
        .route("/", get(root_handler))
        .layer(TraceLayer::new_for_http());

    // Merge shared health, ready, and metrics routes
    let app = vedo_shared::add_health_routes(app, SERVICE_NAME);
    let app = app.route("/metrics", get(vedo_shared::metrics_handler));

    // Publishing routes
    let publish_routes = Router::new()
        .route(
            "/api/v1/ontologies/:ontology_id/publish",
            post(handlers::publish_snapshot_handler),
        )
        .route(
            "/api/v1/ontologies/:ontology_id/snapshots",
            get(handlers::list_snapshots_handler),
        )
        .route(
            "/api/v1/snapshots/:snapshot_id",
            get(handlers::get_snapshot_handler),
        )
        .route(
            "/api/v1/snapshots/:snapshot_id",
            delete(handlers::retire_snapshot_handler),
        );

    let app = app.merge(publish_routes.with_state(state));

    tracing::info!(
        routes_registered = "publish,list_snapshots,get_snapshot,retire_snapshot",
        "Publisher-service routes registered"
    );

    app
}
