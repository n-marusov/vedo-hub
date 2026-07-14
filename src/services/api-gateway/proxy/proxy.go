package proxy

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// hop-by-hop headers per RFC 7230 §6.1. These must never be forwarded
// from client to upstream. Go's httputil.ReverseProxy already removes
// most of them at the transport level; we also strip them explicitly in
// the Director so the upstream never sees them even when the transport
// behaviour changes across Go versions.
var hopByHopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
	"Proxy-Connection",
}

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

	// Director explicitly preserves RawQuery and strips hop-by-hop headers.
	// Go's default Director already copies r.URL (including RawQuery), but
	// we set it explicitly for clarity and defense-in-depth.
	rp.Director = func(r *http.Request) {
		r.URL.Scheme = u.Scheme
		r.URL.Host = u.Host
		r.Host = u.Host
		r.URL.Path = singleJoiningSlash(u.Path, r.URL.Path)
		if r.URL.RawPath != "" {
			r.URL.RawPath = singleJoiningSlash(u.RawPath, r.URL.RawPath)
		}
		if r.URL.RawQuery == "" {
			r.URL.RawQuery = u.RawQuery
		}
		// Strip hop-by-hop headers
		for _, h := range hopByHopHeaders {
			r.Header.Del(h)
		}
		// Also strip any header listed in the Connection header value
		if connHeader := r.Header.Get("Connection"); connHeader != "" {
			for _, part := range strings.Split(connHeader, ",") {
				r.Header.Del(strings.TrimSpace(part))
			}
		}
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
	//
	// IMPORTANT: do NOT call r.Body.Close() here. The httputil.ReverseProxy
	// owns the request lifecycle and may still read from r.Body after this
	// handler returns. Closing it prematurely causes data loss for POST/PUT
	// endpoints — the upstream receives an empty body. Instead we swap in a
	// fresh NopCloser wrapping the buffered bytes, leaving the original body
	// for GC to reclaim once the proxy is done.
	if r.Body != nil && r.ContentLength != 0 {
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Warn("proxy.body_read_failed",
				"service", p.serviceName,
				"path", r.URL.Path,
				"content_length", r.ContentLength,
				"error", err,
				"trace_id", r.Header.Get("X-Trace-Id"),
			)
		} else {
			slog.Debug("proxy.body_buffered",
				"service", p.serviceName,
				"path", r.URL.Path,
				"content_length", r.ContentLength,
				"buffered_bytes", len(buf),
				"trace_id", r.Header.Get("X-Trace-Id"),
			)
			r.Body = io.NopCloser(bytes.NewReader(buf))
			r.ContentLength = int64(len(buf))
		}
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
		"traceparent",
		"tracestate",
	}

	for _, h := range headers {
		if v := r.Header.Get(h); v != "" {
			r.Header.Set(h, v)
		}
	}

	// Debug log for trace context propagation
	if tp := r.Header.Get("traceparent"); tp != "" {
		slog.Debug("proxy.trace_propagation",
			"service", p.serviceName,
			"traceparent", tp,
			"tracestate", r.Header.Get("tracestate"),
			"trace_id", r.Header.Get("X-Trace-Id"),
		)
	}
}

// singleJoiningSlash joins a and b with a single slash (copied from
// net/http/httputil since it is unexported there).
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
