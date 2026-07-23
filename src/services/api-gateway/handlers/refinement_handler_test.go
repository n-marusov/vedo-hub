package handlers

// Validates: REQ-FUN.API.iterative-refinement-context

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
// Helpers — refinement
// ---------------------------------------------------------------------------

func newRefinementTestRouter(provider llm.Provider, renderer PromptRenderer) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRefinementHandler(provider, renderer)
	api := r.Group("/api/v1")
	api.POST("/ontologies/:id/ai/refine", h.HandleRefine)
	return r
}

func serveRefineRequest(r *gin.Engine, ontologyID string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest("POST", "/api/v1/ontologies/"+ontologyID+"/ai/refine", &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---------------------------------------------------------------------------
// Tests — refinement
// ---------------------------------------------------------------------------

func TestHandleRefine_Success(t *testing.T) {
	llmResponse := `{
		"changes": [
			{
				"operation": "MODIFY_CLASS",
				"target_id": "Person",
				"modification": "Add hasAge property",
				"rationale": "Age is a common attribute for Person"
			}
		],
		"summary": "Added hasAge property to Person"
	}`

	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			if name != "refinement" {
				t.Errorf("expected template 'refinement', got %q", name)
			}
			return "Refine the ontology based on feedback", nil
		},
	}
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{Text: "```json\n" + llmResponse + "\n```"}, nil
		},
	}

	r := newRefinementTestRouter(provider, renderer)
	w := serveRefineRequest(r, "test-onto", models.RefinementRequest{
		SequenceID:   "seq-test-1",
		FeedbackText: "Add an age property to Person",
	})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp models.RefinementResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Iteration != 1 {
		t.Errorf("expected iteration 1, got %d", resp.Iteration)
	}
	if len(resp.Changes) != 1 {
		t.Errorf("expected 1 change, got %d", len(resp.Changes))
	}
	if resp.Changes[0].Operation != "MODIFY_CLASS" {
		t.Errorf("expected MODIFY_CLASS, got %q", resp.Changes[0].Operation)
	}
	if resp.Summary != "Added hasAge property to Person" {
		t.Errorf("unexpected summary: %q", resp.Summary)
	}
	if resp.MaxIterations != MaxRefinementRounds {
		t.Errorf("expected MaxIterations %d, got %d", MaxRefinementRounds, resp.MaxIterations)
	}
}

func TestHandleRefine_MissingOntologyID(t *testing.T) {
	// Need a mock provider for nil-guard to pass
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{}, errors.New("should not be called - missing ontology ID")
		},
	}
	r := newRefinementTestRouter(provider, nil)
	w := serveRefineRequest(r, "", models.RefinementRequest{
		SequenceID:   "seq-1",
		FeedbackText: "Make it better",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-MISSING-ONTOLOGY-ID")
}

func TestHandleRefine_InvalidRequest(t *testing.T) {
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{}, errors.New("should not be called - invalid request")
		},
	}
	r := newRefinementTestRouter(provider, nil)
	req := httptest.NewRequest("POST", "/api/v1/ontologies/test-onto/ai/refine",
		bytes.NewReader([]byte(`{"feedback_text":"test"}`))) // missing sequence_id
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-INVALID-REQUEST")
}

func TestHandleRefine_EmptyFeedback(t *testing.T) {
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{}, errors.New("should not be called - empty feedback")
		},
	}
	r := newRefinementTestRouter(provider, nil)
	w := serveRefineRequest(r, "test-onto", models.RefinementRequest{
		SequenceID:   "seq-1",
		FeedbackText: "   ",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-EMPTY-FEEDBACK")
}

func TestHandleRefine_MaxRoundsLimit(t *testing.T) {
	// We'll manually set up a history that's already at the max
	seqID := "seq-max-rounds"

	// Create a history entry with MaxRefinementRounds iterations already done
	sequenceHistoryMu.Lock()
	sequenceHistory[seqID] = &models.SequenceStepHistory{
		SequenceID:  seqID,
		OntologyID:  "test-onto",
		Iteration:   MaxRefinementRounds,
		FeedbackLog: []models.FeedbackEntry{},
	}
	sequenceHistoryMu.Unlock()

	// Clean up after test
	defer func() {
		sequenceHistoryMu.Lock()
		delete(sequenceHistory, seqID)
		sequenceHistoryMu.Unlock()
	}()

	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			return "prompt", nil
		},
	}
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{}, nil
		},
	}

	r := newRefinementTestRouter(provider, renderer)
	w := serveRefineRequest(r, "test-onto", models.RefinementRequest{
		SequenceID:   seqID,
		FeedbackText: "Make more changes",
	})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-REFINEMENT-LIMIT")
}

func TestHandleRefine_LLMError(t *testing.T) {
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

	r := newRefinementTestRouter(provider, renderer)
	w := serveRefineRequest(r, "test-onto", models.RefinementRequest{
		SequenceID:   "seq-llm-error",
		FeedbackText: "Fix the structure",
	})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-LLM-ERROR")
}

func TestHandleRefine_TemplateError(t *testing.T) {
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{}, errors.New("should not be called - template error")
		},
	}
	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			return "", errMockLLMFailure
		},
	}
	r := newRefinementTestRouter(provider, renderer)
	w := serveRefineRequest(r, "test-onto", models.RefinementRequest{
		SequenceID:   "seq-template-error",
		FeedbackText: "Fix the structure",
	})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-TEMPLATE-ERROR")
}

func TestHandleRefine_ParseError(t *testing.T) {
	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			return "prompt", nil
		},
	}
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			// Response missing required JSON structure
			return llm.Completion{Text: "I'm not sure how to refine this."}, nil
		},
	}

	r := newRefinementTestRouter(provider, renderer)
	w := serveRefineRequest(r, "test-onto", models.RefinementRequest{
		SequenceID:   "seq-parse-error",
		FeedbackText: "Fix the structure",
	})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d; body: %s", w.Code, w.Body.String())
	}
	verifyErrorResponse(t, w.Body.Bytes(), "GATEWAY-REFINEMENT-PARSE-ERROR")
}

func TestHandleRefine_IterationCounter(t *testing.T) {
	seqID := "seq-iteration-test"

	renderer := &mockPromptRenderer{
		renderFunc: func(name string, data interface{}) (string, error) {
			return "prompt", nil
		},
	}
	provider := &mockLLMProvider{
		completeFunc: func(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
			return llm.Completion{
				Text: `{"changes":[{"operation":"MODIFY_CLASS","target_id":"X","modification":"Fix X","rationale":"Needed"}],"summary":"Fixed X"}`,
			}, nil
		},
	}

	r := newRefinementTestRouter(provider, renderer)

	// First refinement
	w1 := serveRefineRequest(r, "test-onto", models.RefinementRequest{
		SequenceID:   seqID,
		FeedbackText: "Fix this",
	})
	if w1.Code != http.StatusOK {
		t.Errorf("expected 200 on first refine, got %d", w1.Code)
	}
	var resp1 models.RefinementResponse
	json.Unmarshal(w1.Body.Bytes(), &resp1)
	if resp1.Iteration != 1 {
		t.Errorf("expected iteration 1 on first refine, got %d", resp1.Iteration)
	}

	// Second refinement
	w2 := serveRefineRequest(r, "test-onto", models.RefinementRequest{
		SequenceID:   seqID,
		FeedbackText: "Fix more",
	})
	var resp2 models.RefinementResponse
	json.Unmarshal(w2.Body.Bytes(), &resp2)
	if resp2.Iteration != 2 {
		t.Errorf("expected iteration 2 on second refine, got %d", resp2.Iteration)
	}

	// Clean up
	sequenceHistoryMu.Lock()
	delete(sequenceHistory, seqID)
	sequenceHistoryMu.Unlock()
}
