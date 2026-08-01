package main

// Validates: REQ-FUN.API.graphql-sparql (REST write boundary)
// Validates: ADR-DES.API.rest-graphql-mutation-boundary
// Validates: REQ-FUN.API.rest-gitlab-alignment
//
// These tests assert that the entity write endpoints (POST/PUT/DELETE on
// /api/v1/ontologies/{id}/classes, /properties, /individuals) are routed
// through the REST API Gateway — i.e. they reach the ontology upstream when
// the caller is authorized, and return 401 when the Authorization header is
// missing. Per ADR-DES.API.rest-graphql-mutation-boundary, GraphQL must not
// expose these writes; the REST contract is the only legitimate path.

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestREST_EntityCRUD_RequiresAuth verifies that every entity write endpoint
// returns 401 UNAUTHENTICATED when no Authorization header is supplied.
// This guards against accidental removal of auth middleware on write paths.
func TestREST_EntityCRUD_RequiresAuth(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/ontologies"},
		{http.MethodPut, "/api/v1/ontologies/ont-1"},
		{http.MethodDelete, "/api/v1/ontologies/ont-1"},

		{http.MethodPost, "/api/v1/ontologies/ont-1/classes"},
		{http.MethodPut, "/api/v1/ontologies/ont-1/classes/cls-1"},
		{http.MethodDelete, "/api/v1/ontologies/ont-1/classes/cls-1"},

		{http.MethodPost, "/api/v1/ontologies/ont-1/properties"},
		{http.MethodPut, "/api/v1/ontologies/ont-1/properties/prop-1"},
		{http.MethodDelete, "/api/v1/ontologies/ont-1/properties/prop-1"},

		{http.MethodPost, "/api/v1/ontologies/ont-1/individuals"},
		{http.MethodPut, "/api/v1/ontologies/ont-1/individuals/ind-1"},
		{http.MethodDelete, "/api/v1/ontologies/ont-1/individuals/ind-1"},

		{http.MethodPost, "/api/v1/ontologies/ont-1/import"},
		{http.MethodGet, "/api/v1/ontologies/ont-1/export"},
	}

	for _, c := range cases {
		t.Run(c.method+"_"+c.path, func(t *testing.T) {
			var body io.Reader
			if c.method == http.MethodPost || c.method == http.MethodPut {
				body = jsonBody(t, map[string]any{"label": "x"})
			}
			req := httptest.NewRequest(c.method, c.path, body)
			if body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()
			env.router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401 for %s %s without auth, got %d (body=%s)",
					c.method, c.path, w.Code, w.Body.String())
			}
			var resp map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("decode error body for %s %s: %v", c.method, c.path, err)
			}
			errObj, ok := resp["error"].(map[string]any)
			if !ok {
				t.Fatalf("expected error object for %s %s, got %v", c.method, c.path, resp)
			}
			if errObj["code"] != "UNAUTHENTICATED" {
				t.Errorf("expected UNAUTHENTICATED for %s %s, got %v", c.method, c.path, errObj["code"])
			}
		})
	}
}

// TestREST_EntityCRUD_ReachableWithEditorToken verifies that the same write
// endpoints forward to the ontology upstream when an authorized bearer token
// is supplied. POST/PUT require Editor role (BFLA level 1); DELETE requires
// Maintainer role (BFLA level 2). The mock upstream returns 200 and echoes the
// request method; we assert the response status and that the request actually
// hit the upstream (path echoed back).
func TestREST_EntityCRUD_ReachableWithEditorToken(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	editorBearer := env.bearer(t, "editor-1", []string{"Editor"})
	maintainerBearer := env.bearer(t, "maint-1", []string{"Maintainer"})

	cases := []struct {
		method     string
		path       string
		wantUpPath string // path the upstream should receive (may differ due to proxy normalization)
		bearer     string // which role bearer to use
	}{
		{http.MethodPost, "/api/v1/ontologies/ont-1/classes", "/api/v1/ontologies/ont-1/classes", editorBearer},
		{http.MethodPost, "/api/v1/ontologies/ont-1/properties", "/api/v1/ontologies/ont-1/properties", editorBearer},
		{http.MethodPost, "/api/v1/ontologies/ont-1/individuals", "/api/v1/ontologies/ont-1/individuals", editorBearer},

		{http.MethodPut, "/api/v1/ontologies/ont-1/classes/cls-1", "/api/v1/ontologies/ont-1/classes/cls-1", editorBearer},
		{http.MethodDelete, "/api/v1/ontologies/ont-1/classes/cls-1", "/api/v1/ontologies/ont-1/classes/cls-1", maintainerBearer},

		{http.MethodPut, "/api/v1/ontologies/ont-1/properties/prop-1", "/api/v1/ontologies/ont-1/properties/prop-1", editorBearer},
		{http.MethodDelete, "/api/v1/ontologies/ont-1/properties/prop-1", "/api/v1/ontologies/ont-1/properties/prop-1", maintainerBearer},

		{http.MethodPut, "/api/v1/ontologies/ont-1/individuals/ind-1", "/api/v1/ontologies/ont-1/individuals/ind-1", editorBearer},
		{http.MethodDelete, "/api/v1/ontologies/ont-1/individuals/ind-1", "/api/v1/ontologies/ont-1/individuals/ind-1", maintainerBearer},

		{http.MethodGet, "/api/v1/ontologies/ont-1/export", "/api/v1/ontologies/ont-1/export", editorBearer},
		{http.MethodPost, "/api/v1/ontologies/ont-1/import", "/api/v1/ontologies/ont-1/import", editorBearer},
	}

	for _, c := range cases {
		t.Run(c.method+"_"+c.path, func(t *testing.T) {
			var body io.Reader
			if c.method == http.MethodPost || c.method == http.MethodPut {
				body = jsonBody(t, map[string]any{"label": "x"})
			}
			req := httptest.NewRequest(c.method, c.path, body)
			if body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			req.Header.Set("Authorization", c.bearer)
			req.Header.Set("X-Trace-Id", "trace-rest-crud")
			w := httptest.NewRecorder()
			env.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200 for %s %s with authorized token, got %d (body=%s)",
					c.method, c.path, w.Code, w.Body.String())
			}
			var resp map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("decode upstream echo for %s %s: %v (body=%s)",
					c.method, c.path, err, w.Body.String())
			}
			if resp["method"] != c.method {
				t.Errorf("upstream received method %v, want %s", resp["method"], c.method)
			}
			if resp["path"] != c.wantUpPath {
				t.Errorf("upstream received path %v, want %s", resp["path"], c.wantUpPath)
			}
		})
	}
}

// TestREST_EntityCRUD_ViewerCannotWrite verifies BFLA gating: a Viewer-role
// token (weight 0) cannot POST/PUT/DELETE entity endpoints (level 1+).
func TestREST_EntityCRUD_ViewerCannotWrite(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/ontologies/ont-1/classes"},
		{http.MethodPost, "/api/v1/ontologies/ont-1/properties"},
		{http.MethodPost, "/api/v1/ontologies/ont-1/individuals"},
		{http.MethodDelete, "/api/v1/ontologies/ont-1/classes/cls-1"},
		{http.MethodDelete, "/api/v1/ontologies/ont-1/individuals/ind-1"},
		{http.MethodPut, "/api/v1/ontologies/ont-1/properties/prop-1"},
	}

	bearer := env.bearer(t, "viewer-1", []string{"Viewer"})

	for _, c := range cases {
		t.Run(c.method+"_"+c.path, func(t *testing.T) {
			var body io.Reader
			if c.method == http.MethodPost || c.method == http.MethodPut {
				body = jsonBody(t, map[string]any{"label": "x"})
			}
			req := httptest.NewRequest(c.method, c.path, body)
			if body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			req.Header.Set("Authorization", bearer)
			w := httptest.NewRecorder()
			env.router.ServeHTTP(w, req)

			if w.Code != http.StatusForbidden {
				t.Fatalf("expected 403 for Viewer %s %s, got %d (body=%s)",
					c.method, c.path, w.Code, w.Body.String())
			}
			var resp map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			errObj, _ := resp["error"].(map[string]any)
			if errObj["code"] != "FORBIDDEN_INSUFFICIENT_ROLE" {
				t.Errorf("expected FORBIDDEN_INSUFFICIENT_ROLE for %s %s, got %v",
					c.method, c.path, errObj["code"])
			}
		})
	}
}
