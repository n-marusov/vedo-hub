package proxy

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// Proxy is a generic reverse proxy to an upstream service with header
// propagation, streaming response, error sanitization, and circuit breaker.
type Proxy struct {
	upstreamURL    *url.URL
	reverseProxy   *httputil.ReverseProxy
	timeout        time.Duration
	circuitBreaker *CircuitBreaker
	serviceName    string
}

// New creates a new Proxy that forwards requests to the given upstream.
//   - upstream: the target service URL (e.g. "http://ontology-service:8082")
//   - timeout:  upstream request timeout (default: 30s)
//   - cb:       circuit breaker instance (shared across proxies for the same upstream)
//   - name:     human-readable service name for logging
func New(upstream string, timeout time.Duration, cb *CircuitBreaker, name string) (*Proxy, error) {
	u, err := url.Parse(upstream)
	if err != nil {
		return nil, err
	}

	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	if cb == nil {
		cb = NewCircuitBreaker(5, 30*time.Second, 60*time.Second)
	}

	rp := httputil.NewSingleHostReverseProxy(u)
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		slog.Error("proxy.upstream_error",
			"service", name,
			"upstream", upstream,
			"path", r.URL.Path,
			"error", err,
			"trace_id", r.Header.Get("X-Trace-Id"),
		)
		cb.RecordFailure()
		http.Error(w, `{"error":{"code":"GATEWAY-UPSTREAM-ERROR","message":"Upstream service unavailable"}}`, http.StatusServiceUnavailable)
	}

	rp.ModifyResponse = func(r *http.Response) error {
		if r.StatusCode >= 500 {
			cb.RecordFailure()
		} else {
			cb.RecordSuccess()
		}
		return nil
	}

	return &Proxy{
		upstreamURL:    u,
		reverseProxy:   rp,
		timeout:        timeout,
		circuitBreaker: cb,
		serviceName:    name,
	}, nil
}

// ServeHTTP implements http.Handler. It checks the circuit breaker, propagates
// headers, and delegates to the reverse proxy.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p.circuitBreaker.IsOpen() {
		slog.Warn("proxy.circuit_open",
			"service", p.serviceName,
			"path", r.URL.Path,
			"trace_id", r.Header.Get("X-Trace-Id"),
		)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, `{"error":{"code":"GATEWAY-CIRCUIT-OPEN","message":"Upstream circuit breaker is open; retry later"}}`)
		return
	}

	// Propagate headers from auth middleware
	p.propagateHeaders(r)

	// Buffer the request body so the upstream handler reading response writer
	// before consuming r.Body does not cause the keep-alive transport to return
	// EOF to the caller. This keeps mock upstreams and any handler that ignores
	// body bytes behaving predictably alongside persistent connections.
	if r.Body != nil && r.ContentLength > 0 {
		buf, err := io.ReadAll(r.Body)
		if err == nil {
			r.Body = io.NopCloser(bytes.NewReader(buf))
		}
		_ = r.Body.Close()
	}

	// Remove hop-by-hop headers
	r.RequestURI = ""

	slog.Debug("proxy.forwarding",
		"service", p.serviceName,
		"method", r.Method,
		"path", r.URL.Path,
		"trace_id", r.Header.Get("X-Trace-Id"),
		"user_id", r.Header.Get("X-User-Id"),
	)

	p.reverseProxy.ServeHTTP(w, r)
}

// propagateHeaders copies auth and tracing headers from the incoming request
// to the proxied request.
func (p *Proxy) propagateHeaders(r *http.Request) {
	headers := []string{
		"X-Trace-Id",
		"X-Correlation-Id",
		"X-User-Id",
		"X-User-Roles",
		"X-User-Tenant-Id",
		"Authorization",
	}

	for _, h := range headers {
		if v := r.Header.Get(h); v != "" {
			r.Header.Set(h, v)
		}
	}
}
