use std::sync::Arc;

use versioning_service::{build_app, init_pg_pool, AppState, DEFAULT_PORT, SERVICE_NAME};

#[tokio::main]
async fn main() {
    let port = std::env::var("SERVICE_PORT").unwrap_or_else(|_| DEFAULT_PORT.to_string());

    vedo_shared::tracing::init_tracing(SERVICE_NAME);

    // Initialize PostgreSQL connection pool
    let pg_pool = init_pg_pool().await;
    let state = Arc::new(AppState { pg: pg_pool });
    let app = build_app(state);

    let addr: std::net::SocketAddr = format!("0.0.0.0:{port}").parse().expect("invalid address");
    tracing::info!(port = %port, "Starting versioning-service");
    let listener = tokio::net::TcpListener::bind(addr)
        .await
        .expect("bind failed");
    axum::serve(listener, app.into_make_service())
        .await
        .expect("server failed");
}
