use tracing_subscriber::EnvFilter;

/// Initializes structured JSON logging with OpenTelemetry-compatible fields.
/// Reads log level from `LOG_LEVEL` env var (default: `info`).
///
/// # Panics
/// Panics if the tracing subscriber cannot be installed (should only happen
/// if called more than once, which is a programming error).
pub fn init_tracing(service_name: &str) {
    let log_level = std::env::var("LOG_LEVEL").unwrap_or_else(|_| "info".to_string());

    tracing_subscriber::fmt()
        .json()
        .with_env_filter(
            EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new(&log_level)),
        )
        .with_current_span(true)
        .with_span_list(true)
        .with_target(true)
        .with_timer(tracing_subscriber::fmt::time::UtcTime::rfc_3339())
        .init();

    tracing::info!(
        service = %service_name,
        log_level = %log_level,
        "Tracing initialized for {}",
        service_name
    );
}

#[cfg(test)]
mod tests {
    use super::*;

    // NOTE: Tracing subscriber can only be initialized once per process.
    // Test the configuration by verifying the env filter parses correctly.
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
}
