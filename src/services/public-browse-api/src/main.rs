use std::net::SocketAddr;
use std::sync::Arc;
use tonic::transport::server::ServerTlsConfig;
use tonic::transport::Server;
use tracing::info;
use vedo_shared::protos::public_browse::v1::public_browse_service_server::PublicBrowseServiceServer;
use vedo_shared::tls;

mod grpc;

const DEFAULT_GRPC_PORT: &str = "9011";

#[tokio::main]
async fn main() {
    vedo_shared::tracing::init_tracing(public_browse_api::SERVICE_NAME);

    let http_port = std::env::var("SERVICE_PORT")
        .unwrap_or_else(|_| public_browse_api::DEFAULT_PORT.to_string());
    let grpc_port = std::env::var("GRPC_PORT").unwrap_or_else(|_| DEFAULT_GRPC_PORT.to_string());

    // Initialize snapshot reader and state
    let reader = public_browse_api::snapshot_reader::SnapshotReader::new();
    let state = Arc::new(public_browse_api::AppState { reader });
    let http_app = public_browse_api::build_app(state);

    // gRPC server
    let grpc_svc = PublicBrowseServiceServer::new(grpc::PublicBrowseGrpcServer);
    let grpc_port_clone = grpc_port.clone();
    let http_port_clone = http_port.clone();
    let grpc_addr: SocketAddr = format!("0.0.0.0:{grpc_port}")
        .parse()
        .expect("invalid gRPC address");

    // Configure TLS for gRPC server if enabled
    let grpc_builder = Server::builder();
    let mut grpc_builder = if let Some(identity) = tls::load_grpc_identity().await {
        let tls_config = ServerTlsConfig::new().identity(identity);
        grpc_builder
            .tls_config(tls_config)
            .expect("invalid gRPC TLS config")
    } else {
        grpc_builder
    };

    let grpc_task = tokio::spawn(async move {
        info!(grpc_port = %grpc_port_clone, "Starting public-browse-api gRPC server");
        grpc_builder
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
        info!(http_port = %http_port_clone, "Starting public-browse-api HTTP server");
        let listener = tokio::net::TcpListener::bind(http_addr)
            .await
            .expect("bind failed");
        axum::serve(listener, http_app.into_make_service())
            .await
            .expect("HTTP server failed");
    });

    info!("public-browse-api running: HTTP on {http_port}, gRPC on {grpc_port}");

    tokio::select! {
        r = grpc_task => if let Err(e) = r { tracing::error!(error = %e, "gRPC task failed"); },
        r = http_task => if let Err(e) = r { tracing::error!(error = %e, "HTTP task failed"); },
    }
}
