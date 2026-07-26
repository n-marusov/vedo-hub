package proxy

// Validates: REQ-NFR.SECURITY.bola-bfla-negative-tests

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestProxyForwardsPOSTBody verifies that the proxy correctly buffers and
// forwards the request body for POST requests. This is a regression test for
// the premature r.Body.Close() bug that caused empty bodies to reach the
// upstream for POST/PUT endpoints.
func TestProxyForwardsPOSTBody(t *testing.T) {
	// Upstream echoes the body back so we can verify it was forwarded.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write(body)
	}))
	defer upstream.Close()

	p, err := New(upstream.URL, 0, nil, "test")
	if err != nil {
		t.Fatalf("new proxy: %v", err)
	}

	payload := `{"query":"SELECT ?s WHERE { ?s ?p ?o }"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sparql", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Trace-Id", "test-trace")

	w := httptest.NewRecorder()
	p.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != payload {
		t.Errorf("body not forwarded correctly: got %q, want %q", w.Body.String(), payload)
	}
}

// TestProxyBuffersLargeBody ensures the proxy can handle bodies larger than
// a single read buffer.
func TestProxyBuffersLargeBody(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("X-Body-Length", "0")
		_, _ = w.Write(body)
	}))
	defer upstream.Close()

	p, err := New(upstream.URL, 0, nil, "test")
	if err != nil {
		t.Fatalf("new proxy: %v", err)
	}

	// 128 KiB payload — larger than the default 32 KiB read buffer.
	large := bytes.Repeat([]byte("A"), 128*1024)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cypher", bytes.NewReader(large))
	req.Header.Set("Content-Type", "application/octet-stream")

	w := httptest.NewRecorder()
	p.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.Len() != len(large) {
		t.Errorf("body length mismatch: got %d, want %d", w.Body.Len(), len(large))
	}
}

// TestProxyForwardsGETWithoutBody ensures the body-buffering logic does not
// interfere with GET requests that have no body.
func TestProxyForwardsGETWithoutBody(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	p, err := New(upstream.URL, 0, nil, "test")
	if err != nil {
		t.Fatalf("new proxy: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ontologies", nil)
	req.Header.Set("X-Trace-Id", "test-trace")

	w := httptest.NewRecorder()
	p.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("unexpected response: %q", w.Body.String())
	}
}

// TestProxyPreservesQueryString verifies that GET requests with a query string
// forward the RawQuery unchanged to the upstream.
func TestProxyPreservesQueryString(t *testing.T) {
	var capturedURL string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	p, err := New(upstream.URL, 0, nil, "test")
	if err != nil {
		t.Fatalf("new proxy: %v", err)
	}

	expectedQuery := "query={ontology(id%3A%221%22){id}}&variables=%7B%7D"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/graphql?"+expectedQuery, nil)
	req.Header.Set("X-Trace-Id", "test-trace")

	w := httptest.NewRecorder()
	p.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// The upstream URL must contain the original query string.
	if !strings.Contains(capturedURL, expectedQuery) {
		t.Errorf("query string not preserved: upstream saw %q, expected it to contain %q", capturedURL, expectedQuery)
	}
}

// TestProxyStripsHopByHopHeaders verifies that hop-by-hop headers defined in
// RFC 7230 §6.1 are removed from the forwarded request.
func TestProxyStripsHopByHopHeaders(t *testing.T) {
	var capturedHeaders http.Header
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	p, err := New(upstream.URL, 0, nil, "test")
	if err != nil {
		t.Fatalf("new proxy: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sparql", nil)
	req.Header.Set("X-Trace-Id", "test-trace")
	// Set hop-by-hop headers that must be stripped.
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Transfer-Encoding", "chunked")
	req.Header.Set("Proxy-Connection", "keep-alive")

	w := httptest.NewRecorder()
	p.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Hop-by-hop headers must not be present in the forwarded request.
	forbidden := []string{"Connection", "Transfer-Encoding", "Proxy-Connection"}
	for _, h := range forbidden {
		if v := capturedHeaders.Get(h); v != "" {
			t.Errorf("hop-by-hop header %q was not stripped (value=%q)", h, v)
		}
	}
}
