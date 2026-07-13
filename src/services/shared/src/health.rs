use axum::{
    extract::Extension, http::StatusCode, response::IntoResponse, routing::get, Json, Router,
};
use serde::Serialize;

/// Service name injected via Extension layer.
#[derive(Clone)]
pub struct ServiceName(pub String);

#[derive(Serialize)]
struct HealthResponse {
    status: String,
    service: String,
}

async fn health_handler(Extension(name): Extension<ServiceName>) -> impl IntoResponse {
    tracing::debug!("Health check requested");
    (
        StatusCode::OK,
        Json(HealthResponse {
            status: "healthy".to_string(),
            service: name.0,
        }),
    )
}

async fn ready_handler(Extension(name): Extension<ServiceName>) -> impl IntoResponse {
    tracing::debug!("Readiness check requested");
    (
        StatusCode::OK,
        Json(HealthResponse {
            status: "ready".to_string(),
            service: name.0,
        }),
    )
}

/// Adds standard `/health` and `/ready` routes to a router.
/// Injects the given `service_name` via Extension for handler use.
/// Returns `Router<()>` so callers can continue chaining.
pub fn add_health_routes(router: Router, service_name: &str) -> Router {
    router
        .route("/health", get(health_handler))
        .route("/ready", get(ready_handler))
        .layer(Extension(ServiceName(service_name.to_string())))
}

#[cfg(test)]
mod tests {
    use super::*;
    use axum::{body::Body, http::Request};
    use tower::ServiceExt;

    #[tokio::test]
    async fn test_health_returns_200() {
        let app = add_health_routes(Router::new(), "test-service");
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
    async fn test_ready_returns_200() {
        let app = add_health_routes(Router::new(), "test-service");
        let response = app
            .oneshot(
                Request::builder()
                    .uri("/ready")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(response.status(), StatusCode::OK);
    }

    #[tokio::test]
    async fn test_health_response_body_contains_service_name() {
        let app = add_health_routes(Router::new(), "my-service");
        let response = app
            .oneshot(
                Request::builder()
                    .uri("/health")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        let body: serde_json::Value = serde_json::from_slice(
            &axum::body::to_bytes(response.into_body(), usize::MAX)
                .await
                .unwrap(),
        )
        .unwrap();
        assert_eq!(body["status"], "healthy");
        assert_eq!(body["service"], "my-service");
    }
}
