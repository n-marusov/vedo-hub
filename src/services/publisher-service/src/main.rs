use std::net::SocketAddr;
use std::sync::Arc;

use publisher_service::{build_app, AppState, DEFAULT_PORT, SERVICE_NAME};

#[tokio::main]
async fn main() {
    vedo_shared::tracing::init_tracing(SERVICE_NAME);

    let port = std::env::var("SERVICE_PORT").unwrap_or_else(|_| DEFAULT_PORT.to_string());
    let ontology_service_url = std::env::var("ONTOLOGY_SERVICE_URL")
        .unwrap_or_else(|_| "http://ontology-service:8082".to_string());

    let http_client = reqwest::Client::new();
    let store = publisher_service::storage::SnapshotStore::new();

    let state = Arc::new(AppState {
        store,
        ontology_service_url,
        http_client,
    });

    let app = build_app(state);

    let addr: SocketAddr = format!("0.0.0.0:{port}").parse().expect("invalid address");
    tracing::info!(port = %port, "Starting publisher-service");
    let listener = tokio::net::TcpListener::bind(addr)
        .await
        .expect("bind failed");
    axum::serve(listener, app.into_make_service())
        .await
        .expect("server failed");
}
