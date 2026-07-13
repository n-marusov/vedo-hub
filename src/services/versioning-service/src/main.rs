mod postgres;

use std::env;
use std::net::SocketAddr;
use std::sync::Arc;

use axum::{extract::State, routing::get, Json, Router};
use serde::Serialize;
use tower_http::trace::TraceLayer;

const SERVICE_NAME: &str = "versioning-service";
const DEFAULT_PORT: &str = "8083";

/// Application state shared across handlers.
#[derive(Clone)]
struct AppState {
    pg: Option<postgres::PgPoolWrapper>,
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

#[tokio::main]
async fn main() {
    vedo_shared::tracing::init_tracing(SERVICE_NAME);
    let port = env::var("SERVICE_PORT").unwrap_or_else(|_| DEFAULT_PORT.to_string());

    // Initialize PostgreSQL connection pool
    let pg_config = postgres::PgConfig::from_env();
    let pg_pool = match postgres::create_pool(&pg_config).await {
        Ok(pool) => {
            // Run manual migrations
            match postgres::run_manual_migrations(pool.pool()).await {
                Ok(_) => tracing::info!("PostgreSQL migrations applied successfully"),
                Err(e) => tracing::warn!(error = %e, "Migration warning, continuing"),
            }
            tracing::info!("PostgreSQL pool initialized successfully");
            Some(pool)
        }
        Err(e) => {
            tracing::warn!(error = %e, "Failed to connect to PostgreSQL, running without database");
            None
        }
    };

    let state = Arc::new(AppState { pg: pg_pool });

    // Build stateless routes
    let app = Router::new()
        .route("/", get(root_handler))
        .layer(TraceLayer::new_for_http());

    // Build stateful routes separately and merge after providing state
    let app = app.merge(
        Router::new()
            .route("/health/db", get(pg_health_handler))
            .with_state(state),
    );

    // Add shared health, ready, and metrics routes
    let app = vedo_shared::add_health_routes(app, SERVICE_NAME);
    let app = app.route("/metrics", get(vedo_shared::metrics_handler));

    let addr: SocketAddr = format!("0.0.0.0:{port}").parse().expect("invalid address");
    tracing::info!(port = %port, "Starting versioning-service");
    let listener = tokio::net::TcpListener::bind(addr)
        .await
        .expect("bind failed");
    axum::serve(listener, app.into_make_service())
        .await
        .expect("server failed");
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
    async fn test_root_returns_service_info() {
        let app = Router::new().route("/", get(root_handler)).merge(
            Router::new()
                .route("/health/db", get(pg_health_handler))
                .with_state(test_state()),
        );
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
    async fn test_pg_health_degraded_when_no_db() {
        let app = Router::new().route("/", get(root_handler)).merge(
            Router::new()
                .route("/health/db", get(pg_health_handler))
                .with_state(test_state()),
        );
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
        let app = Router::new().route("/", get(root_handler)).merge(
            Router::new()
                .route("/health/db", get(pg_health_handler))
                .with_state(test_state()),
        );
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
