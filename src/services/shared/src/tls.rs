//! TLS configuration helpers for gRPC servers.
//!
//! Reads TLS certificate and key from environment-configured paths
//! and returns a tonic Identity for server-side TLS.
//!
//! Feature flag: `GRPC_TLS_ENABLED=true` enables TLS.
//! Cert/key paths: `GRPC_TLS_CERT_FILE`, `GRPC_TLS_KEY_FILE`
//! (default: `/etc/vedo/tls/grpc-server.crt`, `/etc/vedo/tls/grpc-server.key`)

/// Load gRPC TLS identity from file paths configured via environment variables.
///
/// Returns `Some(Identity)` when `GRPC_TLS_ENABLED=true` and both cert/key files
/// can be read. Returns `None` when TLS is disabled or files cannot be loaded
/// (warnings are logged on failure).
pub async fn load_grpc_identity() -> Option<tonic::transport::Identity> {
    if std::env::var("GRPC_TLS_ENABLED").as_deref() != Ok("true") {
        tracing::debug!("grpc.tls.disabled");
        return None;
    }

    let cert_path = std::env::var("GRPC_TLS_CERT_FILE")
        .unwrap_or_else(|_| "/etc/vedo/tls/grpc-server.crt".to_string());
    let key_path = std::env::var("GRPC_TLS_KEY_FILE")
        .unwrap_or_else(|_| "/etc/vedo/tls/grpc-server.key".to_string());

    tracing::info!(
        "grpc.tls.loading_identity cert={} key={}",
        cert_path,
        key_path
    );

    let cert = match tokio::fs::read(&cert_path).await {
        Ok(data) => data,
        Err(e) => {
            tracing::warn!("grpc.tls.cert_read_failed path={} error={}", cert_path, e);
            return None;
        }
    };
    let key = match tokio::fs::read(&key_path).await {
        Ok(data) => data,
        Err(e) => {
            tracing::warn!("grpc.tls.key_read_failed path={} error={}", key_path, e);
            return None;
        }
    };

    tracing::info!("grpc.tls.server_enabled");
    Some(tonic::transport::Identity::from_pem(cert, key))
}
