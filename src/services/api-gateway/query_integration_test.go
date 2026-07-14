package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestQuery_SPARQLValidSelectProxied verifies that a SELECT query reaches the
// ontology-service mock with the sanitized query string (LIMIT injected if
// omitted) and produces a 200 response.
func TestQuery_SPARQLValidSelectProxied(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	var receivedQuery string
	env.ontologyServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		receivedQuery = body.Query
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(gin.H{
			"results": []gin.H{{"s": "x", "p": "y", "o": "z"}},
		})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sparql",
		jsonBody(t, gin.H{"query": "SELECT ?s WHERE { ?s ?p ?o }"}))
	req.Header.Set("Authorization", env.bearer(t, "sparql-user", []string{"Viewer"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for SELECT, got %d (body=%s)", w.Code, w.Body.String())
	}
	if receivedQuery == "" {
		t.Fatalf("expected upstream to receive the query payload")
	}
	// Validate that LIMIT was injected upstream (_QUERY_MAX_LIMIT defaults to 1000).
	if !contains(receivedQuery, "LIMIT 1000") {
		t.Errorf("expected sanitized query to include 'LIMIT 1000', got %q", receivedQuery)
	}
}

// TestQuery_SPARQLMutationRejected verifies that mutation keywords cause a
// 400 GATEWAY-QUERY-READONLY response, never reaching the upstream.
func TestQuery_SPARQLMutationRejected(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	upstreamCalled := false
	env.ontologyServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sparql",
		jsonBody(t, gin.H{"query": "INSERT DATA { <a> <b> <c> }"}))
	req.Header.Set("Authorization", env.bearer(t, "sparql-mut", []string{"Editor"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for mutation SPARQL, got %d (body=%s)", w.Code, w.Body.String())
	}
	if upstreamCalled {
		t.Fatalf("upstream must not be reached when rejecting mutations")
	}
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	errObj, _ := body["error"].(map[string]any)
	if errObj["code"] != "GATEWAY-QUERY-READONLY" {
		t.Errorf("expected GATEWAY-QUERY-READONLY, got %v", errObj["code"])
	}
}

// TestQuery_CYPHERValidMatchProxied verifies CYPHER SELECT/MATCH passes
// through and is forwarded to the upstream.
func TestQuery_CYPHERValidMatchProxied(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	env.ontologyServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(gin.H{"results": []gin.H{}})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cypher",
		jsonBody(t, gin.H{"query": "MATCH (n) RETURN n LIMIT 5"}))
	req.Header.Set("Authorization", env.bearer(t, "cypher-user", []string{"Viewer"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for CYPHER MATCH, got %d (body=%s)", w.Code, w.Body.String())
	}
}

// TestQuery_CYPHERCreateMutationRejected verifies CREATE keywords are rejected.
func TestQuery_CYPHERCreateMutationRejected(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	upstreamCalled := false
	env.ontologyServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cypher",
		jsonBody(t, gin.H{"query": "CREATE (n:Thing {id: '1'})"}))
	req.Header.Set("Authorization", env.bearer(t, "cy-mut", []string{"Editor"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for CYPHER CREATE, got %d (body=%s)", w.Code, w.Body.String())
	}
	if upstreamCalled {
		t.Fatalf("upstream must not be reached when rejecting mutations")
	}
}

// contains is a minimal case-sensitive substring helper used in tests.
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
