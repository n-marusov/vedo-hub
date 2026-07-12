package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

const serviceName = "commenting-service"
const defaultPort = "8085"

var requestTotal atomic.Uint64

func resolveID(headerValue string) string {
	if headerValue != "" {
		return headerValue
	}
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func logRequest(path string, status int, traceID string, correlationID string) {
	entry := map[string]any{"service": serviceName, "path": path, "status": status, "trace_id": traceID, "correlation_id": correlationID}
	body, _ := json.Marshal(entry)
	log.Print(string(body))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeMetrics(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "# HELP vedo_service_requests_total Total service requests\n# TYPE vedo_service_requests_total counter\nvedo_service_requests_total{service=\"%s\"} %d\n", serviceName, requestTotal.Load())
}

func main() {
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = defaultPort
	}

	http.HandleFunc("/", handler)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	traceID := resolveID(r.Header.Get("X-Trace-Id"))
	correlationID := resolveID(r.Header.Get("X-Correlation-Id"))
	requestTotal.Add(1)
	knownPath := r.URL.Path == "/" || r.URL.Path == "/health" || r.URL.Path == "/ready" || r.URL.Path == "/metrics"
	if !knownPath {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "ENDPOINT_NOT_FOUND", "message": "The requested endpoint " + r.URL.Path + " does not exist", "available": []string{"/", "/health", "/ready", "/metrics"}})
		logRequest(r.URL.Path, http.StatusNotFound, traceID, correlationID)
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "METHOD_NOT_ALLOWED", "message": "Method " + r.Method + " not allowed on " + r.URL.Path, "available": []string{"/", "/health", "/ready", "/metrics"}})
		logRequest(r.URL.Path, http.StatusMethodNotAllowed, traceID, correlationID)
		return
	}
	switch r.URL.Path {
	case "/":
		writeJSON(w, http.StatusOK, map[string]any{"name": serviceName, "version": "0.2.0", "description": "Commenting service", "stub": false})
	case "/health":
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	case "/ready":
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	case "/metrics":
		writeMetrics(w)
	}
	logRequest(r.URL.Path, http.StatusOK, traceID, correlationID)
}
