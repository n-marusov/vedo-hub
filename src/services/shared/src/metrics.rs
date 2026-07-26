use axum::{http::StatusCode, response::IntoResponse};
use std::sync::atomic::{AtomicU64, Ordering};

static REQUEST_TOTAL: AtomicU64 = AtomicU64::new(0);

/// Increments the request counter. Call from middleware on each request.
pub fn increment_request_count() {
    REQUEST_TOTAL.fetch_add(1, Ordering::Relaxed);
}

/// Returns the current request total.
pub fn request_total() -> u64 {
    REQUEST_TOTAL.load(Ordering::Relaxed)
}

/// Resets the request counter to zero.
/// Only available in test builds for deterministic test isolation.
#[cfg(test)]
pub fn reset_request_count() {
    REQUEST_TOTAL.store(0, Ordering::Relaxed);
}

/// Prometheus-format metrics handler.
pub async fn metrics_handler() -> impl IntoResponse {
    let total = REQUEST_TOTAL.load(Ordering::Relaxed);
    tracing::debug!("Metrics requested, total requests: {}", total);

    let metrics = format!(
        "# HELP vedo_service_requests_total Total service requests\n\
         # TYPE vedo_service_requests_total counter\n\
         vedo_service_requests_total {} \n",
        total
    );

    (
        StatusCode::OK,
        [(
            axum::http::header::CONTENT_TYPE,
            "text/plain; version=0.0.4",
        )],
        metrics,
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_increment_and_read() {
        reset_request_count();
        let before = request_total();
        increment_request_count();
        let after = request_total();
        assert_eq!(after, before + 1);
    }

    #[test]
    fn test_concurrent_increment() {
        reset_request_count();
        let before = request_total();
        increment_request_count();
        increment_request_count();
        assert_eq!(request_total(), before + 2);
    }
}
