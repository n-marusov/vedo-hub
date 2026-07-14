//! Application library for the ontology service.
//!
//! Provides the `build_app` function that assembles all routes, middleware,
//! and shared state into a single axum `Router`.

pub mod classes;
pub mod error;
pub mod graphql;
pub mod handlers;
pub mod individuals;
pub mod neo4j;
pub mod properties;
pub mod services;

use std::sync::Arc;

use axum::{
    extract::State,
    routing::{get, post},
    Json, Router,
};
use graphql::schema::build_schema;
use serde::Serialize;
use tower_http::trace::TraceLayer;

pub const SERVICE_NAME: &str = "ontology-service";
pub const DEFAULT_PORT: &str = "8082";

/// Application state shared across handlers.
#[derive(Clone)]
pub struct AppState {
    pub neo4j: Option<neo4j::Neo4jPool>,
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
        description: "Ontology service — Neo4j CRUD, graph operations",
    })
}

async fn neo4j_health_handler(State(state): State<Arc<AppState>>) -> Json<serde_json::Value> {
    match &state.neo4j {
        Some(pool) => {
            let healthy = neo4j::health_check(pool).await;
            Json(serde_json::json!({
                "status": if healthy { "healthy" } else { "degraded" },
                "database": "neo4j",
                "connected": healthy,
            }))
        }
        None => Json(serde_json::json!({
            "status": "disabled",
            "database": "neo4j",
            "connected": false,
            "message": "Neo4j not configured",
        })),
    }
}

/// Builds the complete axum application with all routes, middleware,
/// and shared state.
///
/// # Arguments
///
/// * `state` - The shared application state (wraps Neo4j pool).
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

    // Merge DB health route with state
    let app = app.merge(
        Router::new()
            .route("/health/db", get(neo4j_health_handler))
            .with_state(state.clone()),
    );

    // Class CRUD routes — careful ordering: static paths before dynamic.
    // NOTE: axum 0.7 (matchit 0.7) uses the `:param` syntax; the `{param}` form
    // is treated as a literal path segment and returns 404 for any value.
    let class_routes = Router::new()
        .route(
            "/api/v1/ontologies/:ontology_id/classes",
            post(classes::create_class_handler).get(classes::list_classes_handler),
        )
        .route(
            "/api/v1/ontologies/:ontology_id/classes/root",
            get(classes::list_root_classes_handler),
        )
        .route(
            "/api/v1/ontologies/:ontology_id/classes/search/autocomplete",
            get(classes::autocomplete_search_handler),
        )
        .route(
            "/api/v1/ontologies/:ontology_id/classes/:class_id",
            get(classes::get_class_handler)
                .put(classes::update_class_handler)
                .delete(classes::delete_class_handler),
        )
        .route(
            "/api/v1/ontologies/:ontology_id/classes/:class_id/children",
            get(classes::get_class_children_handler),
        )
        .route(
            "/api/v1/ontologies/:ontology_id/classes/:class_id/ancestors",
            get(classes::get_ancestors_handler),
        )
        .route(
            "/api/v1/ontologies/:ontology_id/classes/:class_id/descendants",
            get(classes::get_descendants_handler),
        )
        .route(
            "/api/v1/ontologies/:ontology_id/classes/:class_id/breadcrumb",
            get(classes::get_breadcrumb_handler),
        )
        .route(
            "/api/v1/ontologies/:ontology_id/classes/:class_id/neighborhood",
            get(classes::get_neighborhood_handler),
        );

    // Merge class routes with the app state
    let app = app.merge(class_routes.with_state(state.clone()));

    // Property CRUD routes
    let property_routes = Router::new()
        .route(
            "/api/v1/ontologies/:ontology_id/properties",
            post(properties::create_property_handler).get(properties::list_properties_handler),
        )
        .route(
            "/api/v1/ontologies/:ontology_id/properties/:property_id",
            get(properties::get_property_handler)
                .put(properties::update_property_handler)
                .delete(properties::delete_property_handler),
        );

    // Merge property routes with the app state
    let app = app.merge(property_routes.with_state(state.clone()));

    // Individual (ABox) CRUD routes
    let individual_routes = Router::new()
        .route(
            "/api/v1/ontologies/:ontology_id/individuals",
            post(individuals::create_individual_handler).get(individuals::list_individuals_handler),
        )
        .route(
            "/api/v1/ontologies/:ontology_id/individuals/:individual_id",
            get(individuals::get_individual_handler)
                .put(individuals::update_individual_handler)
                .delete(individuals::delete_individual_handler),
        );

    // Merge individual routes with the app state
    let app = app.merge(individual_routes.with_state(state.clone()));

    // Export routes
    let export_routes = Router::new().route(
        "/api/v1/ontologies/:ontology_id/export",
        get(handlers::export_handler::export_ontology_handler),
    );
    let app = app.merge(export_routes.with_state(state.clone()));

    // Import routes
    let import_routes = Router::new().route(
        "/api/v1/ontologies/:ontology_id/import",
        post(handlers::import_handler::import_ontology_handler),
    );
    let app = app.merge(import_routes.with_state(state.clone()));

    // GraphQL endpoint — serves one schema for the whole ontology-service
    let graphql_handler = |State(state): State<Arc<AppState>>,
                           Json(gql_req): Json<async_graphql::Request>| async move {
        let schema = build_schema();
        let response = schema.execute(gql_req.data(state)).await;
        Json(response)
    };
    let graphql_routes = Router::new().route("/api/v1/graphql", post(graphql_handler));
    let app = app.merge(graphql_routes.with_state(state.clone()));

    // Query endpoints — SPARQL/CYPHER read-only facade proxied by the API
    // gateway. The gateway enforces read-only validation and rate limiting
    // before forwarding; this handler applies defense-in-depth validation
    // server-side so internal callers cannot bypass the gate.
    let query_routes = Router::new()
        .route(
            "/api/v1/sparql",
            post(handlers::query_handler::sparql_handler),
        )
        .route(
            "/api/v1/cypher",
            post(handlers::query_handler::cypher_handler),
        );
    let app = app.merge(query_routes.with_state(state.clone()));

    tracing::info!(
        routes_registered = "sparql,cypher,graphql,classes,properties,individuals,export,import",
        "Ontology-service routes registered"
    );

    app
}

/// Initializes the Neo4j connection pool from environment config.
pub async fn init_neo4j_pool() -> Option<neo4j::Neo4jPool> {
    let neo4j_config = neo4j::Neo4jConfig::from_env();
    match neo4j::create_pool(&neo4j_config).await {
        Ok(pool) => {
            tracing::info!("Neo4j pool initialized successfully");
            Some(pool)
        }
        Err(e) => {
            tracing::warn!(error = %e, "Failed to connect to Neo4j, running without database");
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
        Arc::new(AppState { neo4j: None })
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
    async fn test_neo4j_health_degraded_when_no_db() {
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
    async fn test_root_rejects_unknown_routes() {
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
