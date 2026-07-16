package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"vedo-core/src/services/api-gateway/models"
	"vedo-core/src/services/shared/llm"
)

// ---------------------------------------------------------------------------
// Helpers — template handler
// ---------------------------------------------------------------------------

func newTemplateTestRouter(provider llm.Provider, renderer PromptRenderer) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewTemplateHandler(provider, renderer)
	api := r.Group("/api/v1")
	api.GET("/templates/ontologies", h.HandleListTemplates)
	api.POST("/ontologies/:id/apply-template", h.HandleApplyTemplate)
	return r
}

func serveGetTemplates(r *gin.Engine) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "/api/v1/templates/ontologies", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func serveApplyTemplate(r *gin.Engine, ontologyID string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest("POST", "/api/v1/ontologies/"+ontologyID+"/apply-template", &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---------------------------------------------------------------------------
// Tests — list templates
// ---------------------------------------------------------------------------

func TestHandleListTemplates_Success(t *testing.T) {
	r := newTemplateTestRouter(nil, nil)
	w := serveGetTemplates(r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp []models.TemplateSummary
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Should have at least 4 templates
	if len(resp) < 4 {
		t.Errorf("expected at least 4 templates, got %d", len(resp))
	}

	// Check required template IDs exist
	templateIDs := make(map[string]bool)
	for _, tpl := range resp {
		templateIDs[tpl.ID] = true
		if tpl.Name == "" {
			t.Error("template name must not be empty")
		}
		if tpl.Domain == "" {
			t.Error("template domain must not be empty")
		}
	}

	required := []string{"person", "product", "software", "medical"}
	for _, id := range required {
		if !templateIDs[id] {
			t.Errorf("required template %q not found", id)
		}
	}
}

// ---------------------------------------------------------------------------
// Tests — apply template
// ---------------------------------------------------------------------------

func TestHandleApplyTemplate_Success(t *testing.T) {
	// We need a provider for the template handler even though
	// apply-template only uses gRPC (mocked via template load).
	r := newTemplateTestRouter(nil, nil)
	w := serveApplyTemplate(r, "test-onto", models.ApplyTemplateRequest{
		TemplateID: "person",
	})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp models.ApplyTemplateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.TemplateID != "person" {
		t.Errorf("expected template_id 'person', got %q", resp.TemplateID)
	}
	if resp.OntologyID != "test-onto" {
		t.Errorf("expected ontology_id 'test-onto', got %q", resp.OntologyID)
	}
}

func TestHandleApplyTemplate_MissingOntologyID(t *testing.T) {
	r := newTemplateTestRouter(nil, nil)
	w := serveApplyTemplate(r, "", models.ApplyTemplateRequest{
		TemplateID: "person",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-MISSING-ONTOLOGY-ID")
}

func TestHandleApplyTemplate_InvalidRequest(t *testing.T) {
	r := newTemplateTestRouter(nil, nil)

	// Missing template_id in body
	req := httptest.NewRequest("POST", "/api/v1/ontologies/test-onto/apply-template",
		bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-INVALID-REQUEST")
}

func TestHandleApplyTemplate_UnknownTemplate(t *testing.T) {
	r := newTemplateTestRouter(nil, nil)
	w := serveApplyTemplate(r, "test-onto", models.ApplyTemplateRequest{
		TemplateID: "nonexistent-template",
	})
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d; body: %s", w.Code, w.Body.String())
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-TEMPLATE-NOT-FOUND")
}
