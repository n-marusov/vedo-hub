use std::net::SocketAddr;
use std::sync::Arc;

use public_browse_api::{build_app, AppState, DEFAULT_PORT, SERVICE_NAME};

#[tokio::main]
async fn main() {
    vedo_shared::tracing::init_tracing(SERVICE_NAME);

    let port = std::env::var("SERVICE_PORT").unwrap_or_else(|_| DEFAULT_PORT.to_string());

    let reader = public_browse_api::snapshot_reader::SnapshotReader::new();
    let state = Arc::new(AppState { reader });

    let app = build_app(state);

    let addr: SocketAddr = format!("0.0.0.0:{port}").parse().expect("invalid address");
    tracing::info!(port = %port, "Starting public-browse-api");
    let listener = tokio::net::TcpListener::bind(addr)
        .await
        .expect("bind failed");
    axum::serve(listener, app.into_make_service())
        .await
        .expect("server failed");
}
