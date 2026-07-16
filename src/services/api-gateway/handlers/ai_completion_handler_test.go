package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"vedo-core/src/services/api-gateway/models"
	"vedo-core/src/services/shared/llm"
)

// ---------------------------------------------------------------------------
// Helpers — AI completion
// ---------------------------------------------------------------------------

func newAiCompletionTestRouter(provider llm.Provider, renderer PromptRenderer) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAiCompletionHandler(provider, renderer, nil)
	api := r.Group("/api/v1")
	api.POST("/ontologies/:id/ai/suggest-classes", h.HandleSuggestClasses)
	api.POST("/ontologies/:id/ai/suggest-properties", h.HandleSuggestProperties)
	api.POST("/ontologies/:id/ai/suggest-relationships", h.HandleSuggestRelationships)
	return r
}

func serveAiRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---------------------------------------------------------------------------
// Suggest classes
// ---------------------------------------------------------------------------

func TestHandleSuggestClasses_Success(t *testing.T) {
	llmResponse := `{
		"suggestions": [
			{"entity_id": "Employee", "label": "Employee", "relationship_to_focus": "subClassOf", "confidence": 0.95, "rationale": "Employee is a type of Person"},
			{"entity_id": "Customer", "label": "Customer", "relationship_to_focus": "subClassOf", "confidence": 0.85, "rationale": "Customer is a role of Person"}
		]
	}`

	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			if name != "class_completion" {
				return "", nil
			}
			return "Suggest classes for Person", nil
		},
	}
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{Text: "```json\n" + llmResponse + "\n```"}, nil
		},
	}

	r := newAiCompletionTestRouter(provider, renderer)
	w := serveAiRequest(r, "POST", "/api/v1/ontologies/test-onto/ai/suggest-classes",
		models.SuggestClassesRequest{ClassID: "Person"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp models.AISuggestionsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.TotalSuggestions != 2 {
		t.Errorf("expected 2 suggestions, got %d", resp.TotalSuggestions)
	}
	if resp.OntologyID != "test-onto" {
		t.Errorf("expected ontology_id 'test-onto', got %q", resp.OntologyID)
	}
	if resp.Suggestions[0].EntityID != "Employee" {
		t.Errorf("expected 'Employee', got %q", resp.Suggestions[0].EntityID)
	}
	if resp.Suggestions[0].Confidence != 0.95 {
		t.Errorf("expected confidence 0.95, got %f", resp.Suggestions[0].Confidence)
	}
}

func TestHandleSuggestClasses_EmptyOntologyID(t *testing.T) {
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{}, errors.New("should not be called")
		},
	}
	r := newAiCompletionTestRouter(provider, nil)
	w := serveAiRequest(r, "POST", "/api/v1/ontologies//ai/suggest-classes",
		models.SuggestClassesRequest{ClassID: "Person"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-MISSING-ONTOLOGY-ID")
}

func TestHandleSuggestClasses_LLMError(t *testing.T) {
	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			return "prompt", nil
		},
	}
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{}, errMockLLMFailure
		},
	}
	r := newAiCompletionTestRouter(provider, renderer)
	w := serveAiRequest(r, "POST", "/api/v1/ontologies/test-onto/ai/suggest-classes",
		models.SuggestClassesRequest{ClassID: "Person"})

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-LLM-ERROR")
}

func TestHandleSuggestClasses_EmptyBody(t *testing.T) {
	// Empty body should be accepted with defaults
	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			return "prompt", nil
		},
	}
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{Text: `{"suggestions":[]}`}, nil
		},
	}
	r := newAiCompletionTestRouter(provider, renderer)

	// Send empty body
	req := httptest.NewRequest("POST", "/api/v1/ontologies/test-onto/ai/suggest-classes", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleSuggestClasses_ExcessiveSuggestions(t *testing.T) {
	// More than MaxSuggestionCount suggestions should be limited
	suggestions := make([]map[string]interface{}, 30)
	for i := range suggestions {
		suggestions[i] = map[string]interface{}{
			"entity_id":             "Class" + string(rune('A'+i)),
			"label":                 "Class" + string(rune('A'+i)),
			"relationship_to_focus": "subClassOf",
			"confidence":            0.5,
		}
	}

	response := map[string]interface{}{"suggestions": suggestions}
	respBytes, _ := json.Marshal(response)

	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			return "prompt", nil
		},
	}
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{Text: string(respBytes)}, nil
		},
	}
	r := newAiCompletionTestRouter(provider, renderer)
	w := serveAiRequest(r, "POST", "/api/v1/ontologies/test-onto/ai/suggest-classes",
		models.SuggestClassesRequest{ClassID: "Person"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result models.AISuggestionsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result.TotalSuggestions > 25 {
		t.Errorf("expected at most 25 suggestions, got %d", result.TotalSuggestions)
	}
}

// ---------------------------------------------------------------------------
// Suggest properties
// ---------------------------------------------------------------------------

func TestHandleSuggestProperties_Success(t *testing.T) {
	llmResponse := `{
		"properties": [
			{"entity_id": "hasName", "label": "has Name", "domain": "Person", "range": "string", "property_type": "dataProperty", "confidence": 0.98},
			{"entity_id": "hasBirthDate", "label": "has Birth Date", "domain": "Person", "range": "date", "property_type": "dataProperty", "confidence": 0.92}
		]
	}`

	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			return "Suggest properties for Person", nil
		},
	}
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{Text: "```json\n" + llmResponse + "\n```"}, nil
		},
	}

	r := newAiCompletionTestRouter(provider, renderer)
	w := serveAiRequest(r, "POST", "/api/v1/ontologies/test-onto/ai/suggest-properties",
		models.SuggestPropertiesRequest{ClassID: "Person"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp models.AISuggestionsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.TotalSuggestions != 2 {
		t.Errorf("expected 2 suggestions, got %d", resp.TotalSuggestions)
	}
}

func TestHandleSuggestProperties_MissingClassID(t *testing.T) {
	r := newAiCompletionTestRouter(nil, nil)
	w := serveAiRequest(r, "POST", "/api/v1/ontologies/test-onto/ai/suggest-properties",
		models.SuggestPropertiesRequest{})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-INVALID-REQUEST")
}

// ---------------------------------------------------------------------------
// Suggest relationships
// ---------------------------------------------------------------------------

func TestHandleSuggestRelationships_Success(t *testing.T) {
	llmResponse := `{
		"suggestions": [
			{"entity_id": "worksFor", "label": "works For", "domain": "Person", "range": "Organization", "property_type": "objectProperty", "confidence": 0.94, "rationale": "A person works for an organization"}
		]
	}`

	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			return "Suggest relationships", nil
		},
	}
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{Text: "```json\n" + llmResponse + "\n```"}, nil
		},
	}

	r := newAiCompletionTestRouter(provider, renderer)
	w := serveAiRequest(r, "POST", "/api/v1/ontologies/test-onto/ai/suggest-relationships",
		models.SuggestRelationshipsRequest{SourceClassID: "Person", TargetClassID: "Organization"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp models.AISuggestionsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.TotalSuggestions != 1 {
		t.Errorf("expected 1 suggestion, got %d", resp.TotalSuggestions)
	}
}

func TestHandleSuggestRelationships_MissingSourceClassID(t *testing.T) {
	r := newAiCompletionTestRouter(nil, nil)
	w := serveAiRequest(r, "POST", "/api/v1/ontologies/test-onto/ai/suggest-relationships",
		models.SuggestRelationshipsRequest{})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-INVALID-REQUEST")
}

func TestHandleSuggestRelationships_EmptyOntologyID(t *testing.T) {
	r := newAiCompletionTestRouter(nil, nil)
	w := serveAiRequest(r, "POST", "/api/v1/ontologies//ai/suggest-relationships",
		models.SuggestRelationshipsRequest{SourceClassID: "Person"})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-MISSING-ONTOLOGY-ID")
}
