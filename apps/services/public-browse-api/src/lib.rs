//! Application library for the public browse API.
//!
//! Provides read-only endpoints for browsing published ontology snapshots
//! without authentication.

pub mod handlers;
pub mod models;
pub mod snapshot_reader;

use std::sync::Arc;

use axum::{routing::get, Json, Router};
use serde::Serialize;
use tower_http::trace::TraceLayer;

pub const SERVICE_NAME: &str = "public-browse-api";
pub const DEFAULT_PORT: &str = "8087";

/// Application state shared across handlers.
#[derive(Clone)]
pub struct AppState {
    pub reader: snapshot_reader::SnapshotReader,
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
        description: "Public browse API — read-only published ontology browser",
    })
}

/// Builds the complete axum application with all routes, middleware,
/// and shared state.
///
/// # Arguments
///
/// * `state` - The shared application state (wraps snapshot reader).
///
/// # Returns
///
/// A configured `Router` ready to serve.
pub fn build_app(state: Arc<AppState>) -> Router {
    let app = Router::new()
        .route("/", get(root_handler))
        .layer(TraceLayer::new_for_http());

    // Merge shared health, ready, and metrics routes
    let app = vedo_shared::add_health_routes(app, SERVICE_NAME);
    let app = app.route("/metrics", get(vedo_shared::metrics_handler));

    // Public browse routes
    let browse_routes = Router::new()
        .route(
            "/api/v1/ontologies",
            get(handlers::list_published_ontologies),
        )
        .route(
            "/api/v1/ontologies/:ontology_id",
            get(handlers::get_published_ontology),
        )
        .route(
            "/api/v1/ontologies/:ontology_id/class-tree",
            get(handlers::get_class_tree),
        )
        .route("/api/v1/search", get(handlers::search_published));

    let app = app.merge(browse_routes.with_state(state));

    tracing::info!(
        routes_registered = "list_ontologies,get_ontology,class_tree,search",
        "Public-browse-api routes registered"
    );

    app
}
