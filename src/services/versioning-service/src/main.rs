use std::net::SocketAddr;
use tonic::transport::server::TlsConfig;
use tonic::transport::Server;
use tracing::info;
use vedo_shared::protos::versioning::v1::versioning_service_server::VersioningServiceServer;
use vedo_shared::tls;

mod grpc;

const DEFAULT_GRPC_PORT: &str = "9002";

#[tokio::main]
async fn main() {
    vedo_shared::tracing::init_tracing(versioning_service::SERVICE_NAME);

    let http_port = std::env::var("SERVICE_PORT")
        .unwrap_or_else(|_| versioning_service::DEFAULT_PORT.to_string());
    let grpc_port = std::env::var("GRPC_PORT").unwrap_or_else(|_| DEFAULT_GRPC_PORT.to_string());

    // Initialize PostgreSQL connection pool
    let pg_pool = versioning_service::init_pg_pool().await;
    let state = std::sync::Arc::new(versioning_service::AppState { pg: pg_pool });
    let http_app = versioning_service::build_app(state);

    // gRPC server
    let grpc_svc = VersioningServiceServer::new(grpc::VersioningGrpcServer);
    let grpc_port_clone = grpc_port.clone();
    let http_port_clone = http_port.clone();
    let grpc_addr: SocketAddr = format!("0.0.0.0:{grpc_port}")
        .parse()
        .expect("invalid gRPC address");

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
        info!(grpc_port = %grpc_port_clone, "Starting versioning-service gRPC server");
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
        info!(http_port = %http_port_clone, "Starting versioning-service HTTP server");
        let listener = tokio::net::TcpListener::bind(http_addr)
            .await
            .expect("bind failed");
        axum::serve(listener, http_app.into_make_service())
            .await
            .expect("HTTP server failed");
    });

    info!("versioning-service running: HTTP on {http_port}, gRPC on {grpc_port}");

    tokio::select! {
        r = grpc_task => if let Err(e) = r { tracing::error!(error = %e, "gRPC task failed"); },
        r = http_task => if let Err(e) = r { tracing::error!(error = %e, "HTTP task failed"); },
    }
}
