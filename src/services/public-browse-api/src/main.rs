use std::env;
use std::io::{Read, Write};
use std::net::{TcpListener, TcpStream};
use std::sync::atomic::{AtomicU64, Ordering};
use std::time::{SystemTime, UNIX_EPOCH};

const SERVICE_NAME: &str = "public-browse-api";
const DEFAULT_PORT: &str = "8087";

static REQUEST_TOTAL: AtomicU64 = AtomicU64::new(0);

fn resolve_id(value: Option<&str>) -> String {
    match value {
        Some(v) if !v.is_empty() => v.to_string(),
        _ => SystemTime::now().duration_since(UNIX_EPOCH).map_or_else(|_| "0".to_string(), |d| d.as_nanos().to_string()),
    }
}

fn response(status: &str, body: &str, content_type: &str) -> String {
    format!("HTTP/1.1 {}\r\nContent-Type: {}\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}", status, content_type, body.len(), body)
}

fn handle(mut stream: TcpStream) {
    let mut buffer = [0_u8; 8192];
    let read = match stream.read(&mut buffer) { Ok(0) | Err(_) => return, Ok(n) => n };
    let request = String::from_utf8_lossy(&buffer[..read]);
    let mut lines = request.lines();
    let first = lines.next().unwrap_or_default();
    let mut parts = first.split_whitespace();
    let method = parts.next().unwrap_or_default();
    let path = parts.next().unwrap_or("/");
    let mut trace_id = String::new();
    let mut correlation_id = String::new();
    for line in lines {
        let lower = line.to_ascii_lowercase();
        if lower.starts_with("x-trace-id:") { trace_id = line.split_once(':').map(|(_, v)| v.trim().to_string()).unwrap_or_default(); }
        if lower.starts_with("x-correlation-id:") { correlation_id = line.split_once(':').map(|(_, v)| v.trim().to_string()).unwrap_or_default(); }
        if line.is_empty() { break; }
    }
    let trace_id = resolve_id(Some(trace_id.as_str()));
    let correlation_id = resolve_id(Some(correlation_id.as_str()));
    REQUEST_TOTAL.fetch_add(1, Ordering::Relaxed);

    let known = matches!(path, "/" | "/health" | "/ready" | "/metrics");
    let (status, body, content_type) = if !known {
        ("404 Not Found", format!("{{\"error\":\"ENDPOINT_NOT_FOUND\",\"message\":\"The requested endpoint {path} does not exist\",\"available\":[\"/\",\"/health\",\"/ready\",\"/metrics\"]}}"), "application/json")
    } else if method != "GET" {
        ("405 Method Not Allowed", format!("{{\"error\":\"METHOD_NOT_ALLOWED\",\"message\":\"Method {method} not allowed on {path}\",\"available\":[\"/\",\"/health\",\"/ready\",\"/metrics\"]}}"), "application/json")
    } else {
        match path {
            "/" => ("200 OK", format!("{{\"name\":\"{SERVICE_NAME}\",\"version\":\"0.2.0\",\"description\":\"Public browse API\",\"stub\":false}}"), "application/json"),
            "/health" => ("200 OK", "{\"status\":\"healthy\"}".to_string(), "application/json"),
            "/ready" => ("200 OK", "{\"status\":\"ready\"}".to_string(), "application/json"),
            _ => ("200 OK", format!("# HELP vedo_service_requests_total Total service requests\n# TYPE vedo_service_requests_total counter\nvedo_service_requests_total{{service=\"{SERVICE_NAME}\"}} {}\n", REQUEST_TOTAL.load(Ordering::Relaxed)), "text/plain; version=0.0.4"),
        }
    };

    let _ = stream.write_all(response(status, &body, content_type).as_bytes());
    println!("{{\"service\":\"{SERVICE_NAME}\",\"path\":\"{path}\",\"status\":{},\"trace_id\":\"{trace_id}\",\"correlation_id\":\"{correlation_id}\"}}", status.split_whitespace().next().unwrap_or("0"));
}

fn main() {
    let port = env::var("SERVICE_PORT").unwrap_or_else(|_| DEFAULT_PORT.to_string());
    let listener = TcpListener::bind(format!("0.0.0.0:{port}")).expect("bind failed");
    for stream in listener.incoming().flatten() { handle(stream); }
}
