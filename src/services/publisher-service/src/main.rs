use std::net::SocketAddr;
use std::sync::Arc;
use tonic::transport::Server;
use tracing::info;
use vedo_shared::protos::publisher::publisher_service_server::PublisherServiceServer;

mod grpc;

const DEFAULT_GRPC_PORT: &str = "9005";

#[tokio::main]
async fn main() {
    vedo_shared::tracing::init_tracing(publisher_service::SERVICE_NAME);

    let http_port = std::env::var("SERVICE_PORT")
        .unwrap_or_else(|_| publisher_service::DEFAULT_PORT.to_string());
    let grpc_port = std::env::var("GRPC_PORT").unwrap_or_else(|_| DEFAULT_GRPC_PORT.to_string());

    // Initialize state
    let ontology_service_url = std::env::var("ONTOLOGY_SERVICE_URL")
        .unwrap_or_else(|_| "http://ontology-service:8082".to_string());
    let http_client = reqwest::Client::new();
    let store = publisher_service::storage::SnapshotStore::new();
    let state = Arc::new(publisher_service::AppState {
        store,
        ontology_service_url,
        http_client,
    });
    let http_app = publisher_service::build_app(state);

    // gRPC server
    let grpc_svc = PublisherServiceServer::new(grpc::PublisherGrpcServer);
    let grpc_addr: SocketAddr = format!("0.0.0.0:{grpc_port}")
        .parse()
        .expect("invalid gRPC address");
    let grpc_task = tokio::spawn(async move {
        info!(grpc_port = %grpc_port, "Starting publisher-service gRPC server");
        Server::builder()
            .add_service(grpc_svc)
            .serve(grpc_addr)
            .await
            .expect("gRPC server failed");
    });

    // HTTP server
    let http_addr: SocketAddr = format!("0.0.0.0:{http_port}")
        .parse()
        .expect("invalid HTTP address");
    let http_task = tokio::spawn(async move {
        info!(http_port = %http_port, "Starting publisher-service HTTP server");
        let listener = tokio::net::TcpListener::bind(http_addr)
            .await
            .expect("bind failed");
        axum::serve(listener, http_app.into_make_service())
            .await
            .expect("HTTP server failed");
    });

    info!("publisher-service running: HTTP on {http_port}, gRPC on {grpc_port}");

    tokio::select! {
        r = grpc_task => if let Err(e) = r { tracing::error!(error = %e, "gRPC task failed"); },
        r = http_task => if let Err(e) = r { tracing::error!(error = %e, "HTTP task failed"); },
    }
}
