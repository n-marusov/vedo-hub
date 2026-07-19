package main

// Validates: REQ-NFR.SECURITY.bola-bfla-negative-tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestProxy_OntologyRoutesForwardedAndHeadersPropagated verifies that
// `/api/v1/ontologies/*` reaches the ontology-service mock and the
// X-Trace-Id and X-User-Id headers are propagated downstream.
func TestProxy_OntologyRoutesForwardedAndHeadersPropagated(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	// Capture the upstream request for header assertions.
	var capturedHeaders = capturedUpstreamHeaders{}

	env.ontologyServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders.trace = r.Header.Get("X-Trace-Id")
		capturedHeaders.user = r.Header.Get("X-User-Id")
		capturedHeaders.roles = r.Header.Get("X-User-Roles")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(gin.H{
			"path": r.URL.Path,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ontologies/test-ont/classes", nil)
	req.Header.Set("Authorization", env.bearer(t, "user-prop", []string{"Editor"}))
	req.Header.Set("X-Trace-Id", "trace-proxy-1")
	req.Header.Set("X-User-Id", "user-prop")
	req.Header.Set("X-User-Roles", "Editor")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from ontology proxy, got %d (body=%s)", w.Code, w.Body.String())
	}
	if capturedHeaders.trace != "trace-proxy-1" {
		t.Errorf("expected X-Trace-Id propagated, got %q", capturedHeaders.trace)
	}
	if capturedHeaders.user != "user-prop" {
		t.Errorf("expected X-User-Id propagated, got %q", capturedHeaders.user)
	}
	if capturedHeaders.roles != "Editor" {
		t.Errorf("expected X-User-Roles propagated, got %q", capturedHeaders.roles)
	}
}

// TestProxy_VersioningRoutesForwarded verifies `/api/v1/versioning/*` is
// served by the versioning-service mock (distinct upstream).
func TestProxy_VersioningRoutesForwarded(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/versioning/branches?ontology_id=00000000-0000-0000-0000-000000000000", nil)
	req.Header.Set("Authorization", env.bearer(t, "user-v", []string{"Viewer"}))
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from versioning proxy, got %d (body=%s)", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["versioning"] != true {
		t.Errorf("expected response from versioning upstream, got %v", body)
	}
}

// TestProxy_GraphQLEndpointProxied verifies that `/api/v1/graphql` forwards
// to the ontology service. GraphQL write mutations are still served via REST,
// but the storefront Apollo Client routes everything through this endpoint.
func TestProxy_GraphQLEndpointProxied(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	env.ontologyServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/api/v1/graphql" {
			t.Errorf("expected upstream path /api/v1/graphql, got %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(gin.H{"data": gin.H{"ontology": gin.H{"id": "1"}}})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/graphql",
		jsonBody(t, gin.H{"query": "{ ontology(id: \"1\") { id } }"}))
	req.Header.Set("Authorization", env.bearer(t, "gql-user", []string{"Viewer"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from graphql proxy, got %d (body=%s)", w.Code, w.Body.String())
	}
}

// capturedUpstreamHeaders stores the headers seen by the mock upstream for
// later assertion in tests.
type capturedUpstreamHeaders struct {
	trace string
	user  string
	roles string
}

// TestProxy_IdentityHeadersFromAuth verifies that the auth middleware's
// resolved user identity (from JWT claims) is propagated to the upstream
// as X-User-Id and X-User-Roles even when the client request does not
// include those headers.
func TestProxy_IdentityHeadersFromAuth(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	var capturedHeaders = capturedUpstreamHeaders{}

	env.ontologyServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders.trace = r.Header.Get("X-Trace-Id")
		capturedHeaders.user = r.Header.Get("X-User-Id")
		capturedHeaders.roles = r.Header.Get("X-User-Roles")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Only send Authorization + Trace — no X-User-Id or X-User-Roles.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ontologies/test-ont/classes", nil)
	req.Header.Set("Authorization", env.bearer(t, "user-auth-flow", []string{"Editor"}))
	req.Header.Set("X-Trace-Id", "trace-auth-flow-1")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%s)", w.Code, w.Body.String())
	}
	if capturedHeaders.user == "" {
		t.Error("X-User-Id header was not set on upstream request — auth middleware must inject it")
	}
	if capturedHeaders.roles == "" {
		t.Error("X-User-Roles header was not set on upstream request — auth middleware must inject it")
	}
	if capturedHeaders.user != "user-auth-flow" {
		t.Errorf("expected X-User-Id 'user-auth-flow', got %q", capturedHeaders.user)
	}
}
