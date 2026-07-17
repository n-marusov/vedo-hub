use std::net::SocketAddr;
use std::sync::Arc;

use ontology_service::{build_app, init_neo4j_pool, AppState, DEFAULT_PORT, SERVICE_NAME};
use tonic::transport::server::TlsConfig;
use tonic::transport::Server;
use vedo_shared::protos::ontology::v1::ontology_service_server::OntologyServiceServer;
use vedo_shared::tls;

mod grpc;

#[tokio::main]
async fn main() {
    vedo_shared::tracing::init_tracing(SERVICE_NAME);
    let port = std::env::var("SERVICE_PORT").unwrap_or_else(|_| DEFAULT_PORT.to_string());
    let grpc_port = std::env::var("GRPC_PORT").unwrap_or_else(|_| "9001".to_string());

    // Initialize Neo4j connection pool
    let neo4j_pool = init_neo4j_pool().await;
    let state = Arc::new(AppState { neo4j: neo4j_pool });

    // Clone state for gRPC server
    let grpc_state = state.clone();

    // Build the HTTP application (axum)
    let app = build_app(state);

    // Clone port strings before moving into async blocks
    let http_port_str = port.clone();
    let grpc_port_str = grpc_port.clone();

    // gRPC server task
    let grpc_addr: SocketAddr = format!("0.0.0.0:{grpc_port}")
        .parse()
        .expect("invalid gRPC address");
    let grpc_svc = OntologyServiceServer::new(grpc::OntologyGrpcServer { state: grpc_state });

    // Configure TLS for gRPC server if enabled
    let grpc_builder = Server::builder();
    let grpc_builder = if let Some(identity) = tls::load_grpc_identity().await {
        let tls_config = TlsConfig::new().identity(identity);
        grpc_builder
            .tls_config(tls_config)
            .expect("invalid gRPC TLS config")
    } else {
        grpc_builder
    };

    let grpc_task = tokio::spawn(async move {
        tracing::info!(grpc_port = %grpc_port_str, "Starting ontology-service gRPC server");
        grpc_builder
            .add_service(grpc_svc)
            .serve(grpc_addr)
            .await
            .expect("gRPC server failed");
    });

    // HTTP server task
    let http_addr: SocketAddr = format!("0.0.0.0:{port}")
        .parse()
        .expect("invalid HTTP address");
    let http_task = tokio::spawn(async move {
        tracing::info!(http_port = %http_port_str, "Starting ontology-service HTTP server");
        let listener = tokio::net::TcpListener::bind(http_addr)
            .await
            .expect("bind failed");
        axum::serve(listener, app.into_make_service())
            .await
            .expect("HTTP server failed");
    });

    tracing::info!("ontology-service running: HTTP on {port}, gRPC on {grpc_port}");

    // Wait for either server to exit (shouldn't happen)
    tokio::select! {
        result = grpc_task => {
            if let Err(e) = result {
                tracing::error!(error = %e, "gRPC server task failed");
            }
        }
        result = http_task => {
            if let Err(e) = result {
                tracing::error!(error = %e, "HTTP server task failed");
            }
        }
    }
}
