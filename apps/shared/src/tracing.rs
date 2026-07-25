use opentelemetry::trace::TracerProvider;
use opentelemetry_otlp::WithExportConfig;
use opentelemetry_sdk::trace as sdktrace;
use opentelemetry_sdk::Resource;
use tracing_subscriber::layer::SubscriberExt;
use tracing_subscriber::util::SubscriberInitExt;
use tracing_subscriber::EnvFilter;

/// Initializes structured JSON logging with OpenTelemetry OTLP export.
/// Reads log level from `LOG_LEVEL` env var (default: `info`).
/// Reads OTLP endpoint from `OTEL_EXPORTER_OTLP_ENDPOINT` env var
/// (default: `http://otel-collector:4317`).
///
/// OTLP exporter failure is non-fatal — falls back to JSON-only logging.
///
/// # Panics
/// Panics if the tracing subscriber cannot be installed (should only happen
/// if called more than once, which is a programming error).
pub fn init_tracing(service_name: &str) {
    let log_level = std::env::var("LOG_LEVEL").unwrap_or_else(|_| "info".to_string());
    let otlp_endpoint = std::env::var("OTEL_EXPORTER_OTLP_ENDPOINT")
        .unwrap_or_else(|_| "http://otel-collector:4317".to_string());

    // Build a tracer — real OTLP on success, no-op on failure
    let tracer: opentelemetry_sdk::trace::Tracer =
        match try_build_provider(service_name, &otlp_endpoint) {
            Ok(provider) => {
                tracing::info!(
                    endpoint = %otlp_endpoint,
                    service = %service_name,
                    "OpenTelemetry initialized, exporting to {otlp_endpoint}",
                );
                provider.tracer(service_name.to_string())
            }
            Err(e) => {
                tracing::warn!(
                    error = %e,
                    endpoint = %otlp_endpoint,
                    "OTLP exporter failed to connect, falling back to JSON logging only",
                );
                // Build a no-op provider — spans are dropped harmlessly
                sdktrace::TracerProvider::builder()
                    .build()
                    .tracer(service_name.to_string())
            }
        };

    let otel_layer = tracing_opentelemetry::layer().with_tracer(tracer);

    tracing_subscriber::registry()
        .with(EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new(&log_level)))
        .with(
            tracing_subscriber::fmt::layer()
                .json()
                .with_current_span(true)
                .with_span_list(true)
                .with_target(true)
                .with_timer(tracing_subscriber::fmt::time::UtcTime::rfc_3339()),
        )
        .with(otel_layer)
        .init();

    tracing::info!(
        service = %service_name,
        log_level = %log_level,
        "Tracing initialized for {service_name}",
    );
}

/// Attempts to build an OTLP tracer provider.
/// Returns an error when the collector is unreachable or endpoint is invalid.
fn try_build_provider(
    service_name: &str,
    endpoint: &str,
) -> Result<sdktrace::TracerProvider, Box<dyn std::error::Error>> {
    use opentelemetry_otlp::SpanExporter;

    let exporter = SpanExporter::builder()
        .with_tonic()
        .with_endpoint(endpoint.to_string())
        .build()?;

    let provider = sdktrace::TracerProvider::builder()
        .with_simple_exporter(exporter)
        .with_resource(Resource::new(vec![opentelemetry::KeyValue::new(
            "service.name",
            service_name.to_string(),
        )]))
        .build();

    Ok(provider)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_env_filter_parsing() {
        let filter = EnvFilter::try_new("info");
        assert!(filter.is_ok());
    }

    #[test]
    fn test_debug_env_filter() {
        let filter = EnvFilter::try_new("debug");
        assert!(filter.is_ok());
    }

    #[tokio::test]
    async fn test_otlp_init_unreachable_collector() {
        // The OTLP exporter should not panic when the collector is unreachable.
        // Build may succeed (lazy connection) or fail — either is acceptable.
        let result = std::panic::catch_unwind(|| {
            let _ = try_build_provider("test-service", "http://127.0.0.1:1");
        });
        assert!(result.is_ok(), "try_build_provider should not panic");
    }
}
