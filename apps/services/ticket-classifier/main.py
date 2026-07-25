import json
import os
import time
from http.server import BaseHTTPRequestHandler, HTTPServer

SERVICE_NAME = "ticket-classifier"
DEFAULT_PORT = 8089


def resolve_id(value: str | None) -> str:
    if value:
        return value
    return str(time.time_ns())


class Handler(BaseHTTPRequestHandler):
    request_total = 0

    def do_GET(self) -> None:
        trace_id = resolve_id(self.headers.get("X-Trace-Id"))
        correlation_id = resolve_id(self.headers.get("X-Correlation-Id"))
        path = self.path.split("?", 1)[0]
        Handler.request_total += 1

        if path == "/":
            self._json(
                200,
                {
                    "name": SERVICE_NAME,
                    "version": "0.2.0",
                    "description": f"{SERVICE_NAME} service",
                    "stub": False,
                },
            )
        elif path == "/health":
            self._json(200, {"status": "healthy"})
        elif path == "/ready":
            self._json(200, {"status": "ready"})
        elif path == "/metrics":
            self._text(
                200,
                f'# HELP vedo_service_requests_total Total service requests\n# TYPE vedo_service_requests_total counter\nvedo_service_requests_total{{service="{SERVICE_NAME}"}} {Handler.request_total}\n',
            )
        else:
            self._json(
                404,
                {
                    "error": "ENDPOINT_NOT_FOUND",
                    "message": f"The requested endpoint {path} does not exist",
                    "available": ["/", "/health", "/ready", "/metrics"],
                },
            )

        print(
            json.dumps(
                {
                    "service": SERVICE_NAME,
                    "path": path,
                    "status": self._last_status,
                    "trace_id": trace_id,
                    "correlation_id": correlation_id,
                }
            ),
            flush=True,
        )

    def do_POST(self) -> None:
        self._method_not_allowed()

    def do_PUT(self) -> None:
        self._method_not_allowed()

    def do_DELETE(self) -> None:
        self._method_not_allowed()

    def _method_not_allowed(self) -> None:
        path = self.path.split("?", 1)[0]
        if path in ["/", "/health", "/ready", "/metrics"]:
            self._json(
                405,
                {
                    "error": "METHOD_NOT_ALLOWED",
                    "message": f"Method {self.command} not allowed on {path}",
                    "available": ["/", "/health", "/ready", "/metrics"],
                },
            )
            return
        self._json(
            404,
            {
                "error": "ENDPOINT_NOT_FOUND",
                "message": f"The requested endpoint {path} does not exist",
                "available": ["/", "/health", "/ready", "/metrics"],
            },
        )

    def _json(self, status: int, payload: dict) -> None:
        body = json.dumps(payload).encode("utf-8")
        self._last_status = status
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _text(self, status: int, payload: str) -> None:
        body = payload.encode("utf-8")
        self._last_status = status
        self.send_response(status)
        self.send_header("Content-Type", "text/plain; version=0.0.4")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format: str, *args: object) -> None:  # noqa: ARG002
        return


if __name__ == "__main__":
    port = int(os.getenv("SERVICE_PORT", str(DEFAULT_PORT)))
    HTTPServer(("0.0.0.0", port), Handler).serve_forever()
