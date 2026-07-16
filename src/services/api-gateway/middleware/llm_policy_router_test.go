package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupPolicyTest creates a Gin engine with the LLM Policy Router middleware
// and an AI-related test route.
func setupPolicyTest() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LLMPolicyRouter())
	r.POST("/api/v1/ontologies/:id/generate-from-text", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return r
}

func TestLLMPolicyRouter_PublicOntology_Allowed(t *testing.T) {
	r := setupPolicyTest()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ontologies/test-onto/generate-from-text", nil)
	req.Header.Set("X-Ontology-Visibility", "public")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for public ontology, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestLLMPolicyRouter_PrivateExternalSaaS_Blocked(t *testing.T) {
	t.Setenv("DEPLOYMENT_MODE", "saas")
	_ = os.Setenv("LLM_PROVIDER", "openai")
	defer func() {
		_ = os.Unsetenv("LLM_PROVIDER")
	}()

	r := setupPolicyTest()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ontologies/test-onto/generate-from-text", nil)
	req.Header.Set("X-Ontology-Visibility", "private")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for private ontology + SaaS + external, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestLLMPolicyRouter_PrivateExternalSaaS_AdminOverride(t *testing.T) {
	t.Setenv("DEPLOYMENT_MODE", "saas")
	_ = os.Setenv("LLM_PROVIDER", "openai")
	defer func() {
		_ = os.Unsetenv("LLM_PROVIDER")
	}()

	r := setupPolicyTest()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ontologies/test-onto/generate-from-text", nil)
	req.Header.Set("X-Ontology-Visibility", "private")
	req.Header.Set("X-Admin-Override", "true")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with admin override, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestLLMPolicyRouter_OnPremise_AnyLLM_Allowed(t *testing.T) {
	t.Setenv("DEPLOYMENT_MODE", "on-premise")
	defer os.Unsetenv("DEPLOYMENT_MODE")

	r := setupPolicyTest()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ontologies/test-onto/generate-from-text", nil)
	req.Header.Set("X-Ontology-Visibility", "private")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for on-premise + private, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestLLMPolicyRouter_AirGappedExternal_Blocked(t *testing.T) {
	t.Setenv("DEPLOYMENT_MODE", "air-gapped")
	_ = os.Setenv("LLM_PROVIDER", "openai")
	defer func() {
		_ = os.Unsetenv("DEPLOYMENT_MODE")
		_ = os.Unsetenv("LLM_PROVIDER")
	}()

	r := setupPolicyTest()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ontologies/test-onto/generate-from-text", nil)
	req.Header.Set("X-Ontology-Visibility", "public")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for air-gapped + external, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestLLMPolicyRouter_AirGappedLocal_Allowed(t *testing.T) {
	t.Setenv("DEPLOYMENT_MODE", "air-gapped")
	_ = os.Setenv("LLM_PROVIDER", "ollama")
	defer func() {
		_ = os.Unsetenv("DEPLOYMENT_MODE")
		_ = os.Unsetenv("LLM_PROVIDER")
	}()

	r := setupPolicyTest()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ontologies/test-onto/generate-from-text", nil)
	req.Header.Set("X-Ontology-Visibility", "private")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for air-gapped + local LLM, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestLLMPolicyRouter_InternalLocalSaaS_Allowed(t *testing.T) {
	t.Setenv("DEPLOYMENT_MODE", "saas")
	defer os.Unsetenv("DEPLOYMENT_MODE")

	r := setupPolicyTest()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ontologies/test-onto/generate-from-text", nil)
	req.Header.Set("X-Ontology-Visibility", "internal")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for internal + local LLM + SaaS, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestLLMPolicyRouter_NonAIRoute_PassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(LLMPolicyRouter())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for non-AI route, got %d", w.Code)
	}
}

func TestLLMPolicyRouter_BlockedResponseCode(t *testing.T) {
	t.Setenv("DEPLOYMENT_MODE", "saas")
	_ = os.Setenv("LLM_PROVIDER", "openai")
	defer func() {
		_ = os.Unsetenv("DEPLOYMENT_MODE")
		_ = os.Unsetenv("LLM_PROVIDER")
	}()

	r := setupPolicyTest()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ontologies/test-onto/generate-from-text", nil)
	req.Header.Set("X-Ontology-Visibility", "private")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	errObj, ok := body["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if code, ok := errObj["code"].(string); !ok || code != "LLM-POLICY-BLOCKED" {
		t.Errorf("expected error code LLM-POLICY-BLOCKED, got %v", errObj["code"])
	}
	if msg, ok := errObj["message"].(string); !ok || msg == "" {
		t.Errorf("expected non-empty error message, got %v", errObj["message"])
	}
}

func TestIsAIRoute(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/api/v1/ontologies/test/generate-from-text", true},
		{"/api/v1/ontologies/test/ai/complete", true},
		{"/api/v1/ontologies", true},
		{"/health", false},
		{"/api/v1/ontologies/test/classes", true},
		{"/metrics", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isAIRoute(tt.path); got != tt.want {
				t.Errorf("isAIRoute(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
