use actix_web::dev::{Service, ServiceRequest, ServiceResponse, Transform};
use actix_web::{web, App, HttpRequest, HttpResponse, HttpServer, Responder};
use opentelemetry::global;
use opentelemetry::trace::TraceContextExt;
use opentelemetry_otlp::WithExportConfig;
use prometheus::{Encoder, HistogramVec, IntCounterVec, Registry, TextEncoder};
use serde::Serialize;
use serde_json::json;
use std::future::{ready, Future, Ready};
use std::pin::Pin;
use std::sync::Arc;
use std::task::{Context, Poll};
use tracing::info;
use tracing_opentelemetry::OpenTelemetrySpanExt;
use tracing_subscriber::layer::SubscriberExt;
use tracing_subscriber::util::SubscriberInitExt;
use uuid::Uuid;

#[derive(Clone)]
struct Metrics {
    requests_total: IntCounterVec,
    request_duration_seconds: HistogramVec,
    request_errors_total: IntCounterVec,
    registry: Registry,
}

#[derive(Serialize)]
struct MetadataResponse {
    name: &'static str,
    version: &'static str,
    description: &'static str,
    stub: bool,
}

fn init_observability(otel_endpoint: &str) {
    let tracer = opentelemetry_otlp::new_pipeline()
        .tracing()
        .with_exporter(opentelemetry_otlp::new_exporter().tonic().with_endpoint(otel_endpoint))
        .install_simple()
        .unwrap();

    let telemetry = tracing_opentelemetry::layer().with_tracer(tracer);
    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::new("info"))
        .with(tracing_subscriber::fmt::layer().json())
        .with(telemetry)
        .init();
}

fn redact(input: &str) -> String {
    input
        .replace("password", "[REDACTED]")
        .replace("token", "[REDACTED]")
        .replace("secret", "[REDACTED]")
}

fn current_trace_id() -> String {
    let cx = tracing::Span::current().context();
    let id = cx.span().span_context().trace_id();
    id.to_string()
}

async fn root() -> impl Responder {
    HttpResponse::Ok().json(MetadataResponse {
        name: "service-rust-template",
        version: "0.1.0",
        description: "Rust service template with observability",
        stub: true,
    })
}

async fn health() -> impl Responder {
    HttpResponse::Ok().json(json!({ "status": "healthy" }))
}

async fn ready() -> impl Responder {
    HttpResponse::Ok().json(json!({ "status": "ready" }))
}

async fn metrics_endpoint(metrics: web::Data<Arc<Metrics>>) -> impl Responder {
    let mut buffer = Vec::new();
    let encoder = TextEncoder::new();
    let families = metrics.registry.gather();
    encoder.encode(&families, &mut buffer).unwrap();
    HttpResponse::Ok().body(String::from_utf8(buffer).unwrap())
}

pub struct LoggingMiddleware;

impl<S, B> Transform<S, ServiceRequest> for LoggingMiddleware
where
    S: Service<ServiceRequest, Response = ServiceResponse<B>, Error = actix_web::Error> + 'static,
    S::Future: 'static,
    B: 'static,
{
    type Response = ServiceResponse<B>;
    type Error = actix_web::Error;
    type InitError = ();
    type Transform = LoggingMiddlewareService<S>;
    type Future = Ready<Result<Self::Transform, Self::InitError>>;

    fn new_transform(&self, service: S) -> Self::Future {
        ready(Ok(LoggingMiddlewareService { service }))
    }
}

pub struct LoggingMiddlewareService<S> {
    service: S,
}

impl<S, B> Service<ServiceRequest> for LoggingMiddlewareService<S>
where
    S: Service<ServiceRequest, Response = ServiceResponse<B>, Error = actix_web::Error> + 'static,
    S::Future: 'static,
    B: 'static,
{
    type Response = ServiceResponse<B>;
    type Error = actix_web::Error;
    type Future = Pin<Box<dyn Future<Output = Result<Self::Response, Self::Error>>>>;

    fn poll_ready(&self, ctx: &mut Context<'_>) -> Poll<Result<(), Self::Error>> {
        self.service.poll_ready(ctx)
    }

    fn call(&self, req: ServiceRequest) -> Self::Future {
        let correlation_id = req
            .headers()
            .get("x-correlation-id")
            .and_then(|value| value.to_str().ok())
            .map(|s| s.to_string())
            .unwrap_or_else(|| Uuid::new_v4().to_string());
        let method = req.method().to_string();
        let path = req.path().to_string();
        let fut = self.service.call(req);

        Box::pin(async move {
            let response = fut.await?;
            let trace_id = current_trace_id();
            info!(
                "{}",
                json!({
                    "level": "info",
                    "message": redact("request_completed"),
                    "method": method,
                    "path": path,
                    "trace_id": trace_id,
                    "correlation_id": correlation_id
                })
            );
            Ok(response)
        })
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let endpoint = std::env::var("OTEL_EXPORTER_OTLP_ENDPOINT")
        .unwrap_or_else(|_| "http://otel-collector:4317".to_string());
    init_observability(&endpoint);

    let registry = Registry::new();
    let requests_total = IntCounterVec::new(
        prometheus::Opts::new("vedo_requests_total", "Total requests"),
        &["method", "path", "status"],
    )
    .unwrap();
    let request_duration_seconds = HistogramVec::new(
        prometheus::HistogramOpts::new("vedo_request_duration_seconds", "Request duration"),
        &["method", "path"],
    )
    .unwrap();
    let request_errors_total = IntCounterVec::new(
        prometheus::Opts::new("vedo_request_errors_total", "Request errors"),
        &["method", "path"],
    )
    .unwrap();
    registry.register(Box::new(requests_total.clone())).unwrap();
    registry
        .register(Box::new(request_duration_seconds.clone()))
        .unwrap();
    registry
        .register(Box::new(request_errors_total.clone()))
        .unwrap();

    let metrics = Arc::new(Metrics {
        requests_total,
        request_duration_seconds,
        request_errors_total,
        registry,
    });

    HttpServer::new(move || {
        App::new()
            .wrap(LoggingMiddleware)
            .app_data(web::Data::new(metrics.clone()))
            .route("/", web::get().to(root))
            .route("/health", web::get().to(health))
            .route("/ready", web::get().to(ready))
            .route("/metrics", web::get().to(metrics_endpoint))
    })
    .bind(("0.0.0.0", 8080))?
    .run()
    .await?;

    global::shutdown_tracer_provider();
    Ok(())
}
