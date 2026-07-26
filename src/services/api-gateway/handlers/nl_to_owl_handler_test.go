package handlers

// Validates: REQ-FUN.API.class-hierarchy-accuracy
// Validates: REQ-FUN.API.properties-accuracy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"vedo-core/src/services/api-gateway/models"
	"vedo-core/src/services/shared/llm"
)

// Shared mock error for AI handler tests.
var errMockLLMFailure = errors.New("mock LLM failure")

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

// mockLLMProvider is a test double for llm.Provider.
type mockLLMProvider struct {
	completeFunc func(ctx llm.Context, p llm.Prompt) (llm.Completion, error)
}

func (m *mockLLMProvider) Complete(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
	if m.completeFunc != nil {
		return m.completeFunc(ctx, p)
	}
	return llm.Completion{}, errors.New("mock not configured")
}

func (m *mockLLMProvider) StreamComplete(ctx llm.Context, p llm.Prompt) (<-chan llm.Token, error) {
	return nil, errors.New("stream not implemented in mock")
}

// mockPromptRenderer implements PromptRenderer for testing.
type mockPromptRenderer struct {
	renderFunc func(name string, data interface{}) (string, error)
}

func (m *mockPromptRenderer) Render(name string, data interface{}) (string, error) {
	if m.renderFunc != nil {
		return m.renderFunc(name, data)
	}
	return "", errors.New("mock renderer not configured")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newNlToOwlTestRouter creates a Gin engine with the NL→OWL handler wired up.
func newNlToOwlTestRouter(provider llm.Provider, renderer PromptRenderer) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewNlToOwlHandler(provider, renderer)
	api := r.Group("/api/v1")
	api.POST("/ontologies/:id/generate-from-text", h.HandleGenerateFromText)
	return r
}

// serveNlToOwlRequest is a convenience to POST to the NL→OWL endpoint.
func serveNlToOwlRequest(r *gin.Engine, ontologyID string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest("POST", "/api/v1/ontologies/"+ontologyID+"/generate-from-text", &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestNlToOwlHandler_EmptyOntologyID(t *testing.T) {
	// Need a non-nil provider for the nil-guard to pass
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{}, errors.New("should not be called")
		},
	}
	r := newNlToOwlTestRouter(provider, nil)
	w := serveNlToOwlRequest(r, "", models.NlToOwlRequest{Text: "Create a person ontology"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-MISSING-ONTOLOGY-ID")
}

func TestNlToOwlHandler_MissingText(t *testing.T) {
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{}, errors.New("should not be called")
		},
	}
	r := newNlToOwlTestRouter(provider, nil)

	// Send body with missing text field
	req := httptest.NewRequest("POST", "/api/v1/ontologies/test-onto/generate-from-text",
		bytes.NewReader([]byte(`{"model":"gpt-4"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	// Empty text after binding should produce GATEWAY-EMPTY-TEXT
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-EMPTY-TEXT")
}

func TestNlToOwlHandler_EmptyText(t *testing.T) {
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{}, errors.New("should not be called")
		},
	}
	r := newNlToOwlTestRouter(provider, nil)
	w := serveNlToOwlRequest(r, "test-onto", models.NlToOwlRequest{Text: "   "})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-EMPTY-TEXT")
}

func TestNlToOwlHandler_TemplateError(t *testing.T) {
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{}, errors.New("should not be called")
		},
	}
	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			return "", errors.New("template not found")
		},
	}
	r := newNlToOwlTestRouter(provider, renderer)
	w := serveNlToOwlRequest(r, "test-onto", models.NlToOwlRequest{Text: "Create a person ontology"})

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-TEMPLATE-ERROR")
}

func TestNlToOwlHandler_LLMError(t *testing.T) {
	// Renderer returns successfully
	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			return "Generate ontology from: test", nil
		},
	}
	// Provider returns error
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{}, errors.New("LLM API unavailable")
		},
	}

	r := newNlToOwlTestRouter(provider, renderer)
	w := serveNlToOwlRequest(r, "test-onto", models.NlToOwlRequest{Text: "Create a person ontology"})

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-LLM-ERROR")
}

func TestNlToOwlHandler_ParseError(t *testing.T) {
	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			return "Generate ontology from: test", nil
		},
	}
	// Provider returns unparseable response
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{
				Text:       "I don't understand, here's some plain text without JSON at all",
				TokensUsed: 15,
			}, nil
		},
	}

	r := newNlToOwlTestRouter(provider, renderer)
	w := serveNlToOwlRequest(r, "test-onto", models.NlToOwlRequest{Text: "Create something"})

	// Unparseable response should return 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-LLM-PARSE-ERROR")
}

func TestNlToOwlHandler_Success(t *testing.T) {
	validLLMResponse := `{
		"steps": [
			{
				"operation": "CREATE_CLASS",
				"entity_id": "Person",
				"label": "Person",
				"parent_id": "Thing",
				"annotations": {"rdfs:comment": "A person entity"}
			},
			{
				"operation": "CREATE_CLASS",
				"entity_id": "Organization",
				"label": "Organization",
				"parent_id": "Thing",
				"annotations": {"rdfs:comment": "An organization entity"}
			}
		]
	}`

	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			m := data.(map[string]interface{})
			if m["Description"] != "Create a person ontology" {
				return "", fmt.Errorf("unexpected description: %v", m["Description"])
			}
			return "Generate ontology from: test", nil
		},
	}
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{
				Text:         "```json\n" + validLLMResponse + "\n```",
				TokensUsed:   42,
				PromptTokens: 120,
			}, nil
		},
	}

	r := newNlToOwlTestRouter(provider, renderer)
	w := serveNlToOwlRequest(r, "test-onto", models.NlToOwlRequest{
		Text:  "Create a person ontology",
		Model: "gpt-4o",
	})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp models.SequencePreview
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.OntologyID != "test-onto" {
		t.Errorf("expected ontology_id 'test-onto', got %q", resp.OntologyID)
	}
	if resp.TotalSteps != 2 {
		t.Errorf("expected 2 steps, got %d", resp.TotalSteps)
	}
	if len(resp.Steps) != 2 {
		t.Errorf("expected 2 steps array, got %d", len(resp.Steps))
	}
	if resp.Steps[0].EntityID != "Person" {
		t.Errorf("expected first step EntityID 'Person', got %q", resp.Steps[0].EntityID)
	}
	if resp.Steps[0].Operation != "CREATE_CLASS" {
		t.Errorf("expected first step Operation 'CREATE_CLASS', got %q", resp.Steps[0].Operation)
	}
}

func TestNlToOwlHandler_SuccessWithEmptySteps(t *testing.T) {
	emptyResponse := `{"steps": []}`

	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			return "Generate ontology", nil
		},
	}
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{Text: emptyResponse, TokensUsed: 5}, nil
		},
	}

	r := newNlToOwlTestRouter(provider, renderer)
	w := serveNlToOwlRequest(r, "test-onto", models.NlToOwlRequest{Text: "Something vague"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp models.SequencePreview
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.TotalSteps != 0 {
		t.Errorf("expected 0 steps, got %d", resp.TotalSteps)
	}
	// Zero steps should produce a warning
	if len(resp.Warnings) == 0 {
		t.Error("expected a warning when no steps are generated")
	}
}

func TestNlToOwlHandler_SuccessWithMarkdownBlock(t *testing.T) {
	// Test that LLM response in markdown code block is properly parsed
	markdownResponse := "Here's the ontology:\n\n```json\n{\"steps\":[{\"operation\":\"CREATE_CLASS\",\"entity_id\":\"Animal\",\"label\":\"Animal\",\"parent_id\":\"Thing\"}]}\n```"

	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			return "prompt", nil
		},
	}
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{Text: markdownResponse, TokensUsed: 10}, nil
		},
	}

	r := newNlToOwlTestRouter(provider, renderer)
	w := serveNlToOwlRequest(r, "test-onto", models.NlToOwlRequest{Text: "Create animal ontology"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp models.SequencePreview
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.TotalSteps != 1 {
		t.Errorf("expected 1 step, got %d", resp.TotalSteps)
	}
	if resp.Steps[0].EntityID != "Animal" {
		t.Errorf("expected 'Animal', got %q", resp.Steps[0].EntityID)
	}
}

func TestNlToOwlHandler_ModelOverride(t *testing.T) {
	var capturedModel string
	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			return "prompt", nil
		},
	}
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			capturedModel = p.Model
			return llm.Completion{
				Text: `{"steps":[{"operation":"CREATE_CLASS","entity_id":"Test","label":"Test","parent_id":"Thing"}]}`,
			}, nil
		},
	}

	r := newNlToOwlTestRouter(provider, renderer)
	_ = serveNlToOwlRequest(r, "test-onto", models.NlToOwlRequest{
		Text:  "Create test",
		Model: "claude-3-opus",
	})

	if capturedModel != "claude-3-opus" {
		t.Errorf("expected model override 'claude-3-opus', got %q", capturedModel)
	}
}
