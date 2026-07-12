package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupTest() {
	tokenStore = map[string]*StoredSession{}
	refreshStore = map[string]*StoredRefreshToken{}
	globalRateLimiter = newSlidingWindowCounter()
	globalIdpState = &IdpStateMachine{currentState: IdpPrimary, localDBTTL: 3600}
}

func doRequest(method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/token", rateLimitMiddleware(handleToken))
	mux.HandleFunc("/api/v1/auth/token/introspect", rateLimitMiddleware(handleIntrospect))
	mux.HandleFunc("/api/v1/auth/token/refresh", rateLimitMiddleware(handleRefresh))
	mux.HandleFunc("/api/v1/auth/session", rateLimitMiddleware(handleGetSession))
	mux.HandleFunc("/api/v1/auth/logout", rateLimitMiddleware(handleLogout))
	mux.HandleFunc("/api/v1/auth/status", rateLimitMiddleware(handleStatus))
	mux.ServeHTTP(w, req)
	return w
}

func readBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return result
}

// @hlv [AUTH-RBAC-001] Token issue and introspection happy path
// @hlv:sec [AUTH_BOUNDARY] Token validation at auth boundary
func TestTokenIssueAndIntrospect_HappyPath(t *testing.T) {
	setupTest()

	tokenResp := doRequest("POST", "/api/v1/auth/token", TokenRequest{
		GrantType: "authorization_code",
		Code:      "valid-code-123",
	})
	if tokenResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", tokenResp.Code)
	}

	body := readBody(t, tokenResp)
	token := body["token"].(map[string]any)
	accessToken := token["access_token"].(string)

	introResp := doRequest("POST", "/api/v1/auth/token/introspect", IntrospectRequest{
		Token: accessToken,
	})
	if introResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", introResp.Code)
	}

	introBody := readBody(t, introResp)
	if introBody["active"] != true {
		t.Fatal("expected token to be active")
	}
	if introBody["user_id"] == "" {
		t.Fatal("expected user_id in introspection")
	}
}

// @hlv [API-REST-001] Invalid token returns 401
// @hlv:sec [AUTH_BOUNDARY] Token rejection at auth boundary
func TestIntrospect_InvalidToken_Returns401(t *testing.T) {
	setupTest()

	resp := doRequest("POST", "/api/v1/auth/token/introspect", IntrospectRequest{
		Token: "non-existent-token",
	})
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.Code)
	}

	body := readBody(t, resp)
	errObj := body["error"].(map[string]any)
	if errObj["code"] != "INVALID_TOKEN" {
		t.Fatalf("expected INVALID_TOKEN, got %s", errObj["code"])
	}
}

// @hlv [API-REST-001] Expired token returns 401
func TestIntrospect_ExpiredToken_Returns401(t *testing.T) {
	setupTest()

	tokenResp := doRequest("POST", "/api/v1/auth/token", TokenRequest{
		GrantType: "authorization_code",
		Code:      "valid-code-123",
	})
	body := readBody(t, tokenResp)
	token := body["token"].(map[string]any)
	accessToken := token["access_token"].(string)

	mu.Lock()
	s, ok := tokenStore[accessToken]
	if ok {
		s.TokenExpiry = time.Now().Add(-1 * time.Hour)
	}
	mu.Unlock()

	resp := doRequest("POST", "/api/v1/auth/token/introspect", IntrospectRequest{
		Token: accessToken,
	})
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.Code)
	}

	body = readBody(t, resp)
	errObj := body["error"].(map[string]any)
	if errObj["code"] != "TOKEN_EXPIRED" {
		t.Fatalf("expected TOKEN_EXPIRED, got %s", errObj["code"])
	}
}

// @hlv [API-REST-001] Rate limiting returns 429 when exceeded
func TestRateLimit_Returns429_WhenExceeded(t *testing.T) {
	setupTest()

	for i := 0; i < 12; i++ {
		resp := doRequest("GET", "/api/v1/auth/status", nil)
		if i < 10 && resp.Code != http.StatusOK {
			t.Fatalf("expected 200 on request %d, got %d", i, resp.Code)
		}
		if i >= 10 && resp.Code != http.StatusTooManyRequests {
			t.Fatalf("expected 429 on request %d, got %d", i, resp.Code)
		}
	}
}

// @hlv [API-REST-001] Session active/inactive
func TestSession_ActiveAndInactive(t *testing.T) {
	setupTest()

	tokenResp := doRequest("POST", "/api/v1/auth/token", TokenRequest{
		GrantType: "authorization_code",
		Code:      "valid-code-123",
	})
	body := readBody(t, tokenResp)
	token := body["token"].(map[string]any)
	accessToken := token["access_token"].(string)

	req := httptest.NewRequest("GET", "/api/v1/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/session", rateLimitMiddleware(handleGetSession))
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var session SessionInfo
	if err := json.NewDecoder(w.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	if !session.Active {
		t.Fatal("expected session to be active")
	}

	logoutReq := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+accessToken)
	lw := httptest.NewRecorder()
	mux2 := http.NewServeMux()
	mux2.HandleFunc("/api/v1/auth/logout", rateLimitMiddleware(handleLogout))
	mux2.ServeHTTP(lw, logoutReq)

	if lw.Code != http.StatusOK {
		t.Fatalf("expected 200 on logout, got %d", lw.Code)
	}

	afterReq := httptest.NewRequest("GET", "/api/v1/auth/session", nil)
	afterReq.Header.Set("Authorization", "Bearer "+accessToken)
	aw := httptest.NewRecorder()
	mux3 := http.NewServeMux()
	mux3.HandleFunc("/api/v1/auth/session", rateLimitMiddleware(handleGetSession))
	mux3.ServeHTTP(aw, afterReq)

	if aw.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after logout, got %d", aw.Code)
	}
}

// @hlv [API-REST-001] Token refresh with rotation
func TestTokenRefresh_Rotation(t *testing.T) {
	setupTest()

	tokenResp := doRequest("POST", "/api/v1/auth/token", TokenRequest{
		GrantType: "authorization_code",
		Code:      "valid-code-123",
	})
	body := readBody(t, tokenResp)
	token := body["token"].(map[string]any)
	oldRefresh := token["refresh_token"].(string)

	refreshResp := doRequest("POST", "/api/v1/auth/token/refresh", TokenRequest{
		RefreshToken: oldRefresh,
	})
	if refreshResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", refreshResp.Code)
	}

	reuseResp := doRequest("POST", "/api/v1/auth/token/refresh", TokenRequest{
		RefreshToken: oldRefresh,
	})
	if reuseResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on reused refresh token, got %d", reuseResp.Code)
	}

	body = readBody(t, reuseResp)
	errObj := body["error"].(map[string]any)
	if errObj["code"] != "REFRESH_FAILED" {
		t.Fatalf("expected REFRESH_FAILED, got %s", errObj["code"])
	}
}

// @hlv [API-REST-001] Invalid authorization code returns 401
func TestToken_InvalidCode_Returns401(t *testing.T) {
	setupTest()

	resp := doRequest("POST", "/api/v1/auth/token", TokenRequest{
		GrantType: "authorization_code",
		Code:      "invalid-code",
	})
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.Code)
	}

	body := readBody(t, resp)
	errObj := body["error"].(map[string]any)
	if errObj["code"] != "INVALID_TOKEN" {
		t.Fatalf("expected INVALID_TOKEN, got %s", errObj["code"])
	}
}

// @hlv [API-REST-001] Client credentials grant works
func TestToken_ClientCredentials(t *testing.T) {
	setupTest()

	resp := doRequest("POST", "/api/v1/auth/token", TokenRequest{
		GrantType:    "client_credentials",
		ClientID:     "vedo-spa",
		ClientSecret: "secret",
	})
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	body := readBody(t, resp)
	token := body["token"].(map[string]any)
	if token["token_type"] != "Bearer" {
		t.Fatalf("expected Bearer, got %s", token["token_type"])
	}
}

// @hlv:sec [AUTH_BOUNDARY] Unauthenticated session returns 404
func TestSession_NoAuth_Returns401(t *testing.T) {
	setupTest()

	resp := doRequest("GET", "/api/v1/auth/session", nil)
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.Code)
	}
}

// @hlv [API-REST-001] Status endpoint returns IdP state
func TestStatus_ReturnsIdPState(t *testing.T) {
	setupTest()

	resp := doRequest("GET", "/api/v1/auth/status", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	body := readBody(t, resp)
	if body["idp_state"] != "primary" {
		t.Fatalf("expected primary, got %s", body["idp_state"])
	}
}

// @hlv:sec [AUTH_BOUNDARY] Degraded IdP state escalates fallback chain
func TestIdp_FallbackChainDegrade(t *testing.T) {
	setupTest()

	state := &IdpStateMachine{currentState: IdpPrimary, localDBTTL: 3600}

	if state.IsDegraded() {
		t.Fatal("expected not degraded initially")
	}

	for range 3 {
		state.RecordFailure()
	}
	if state.CurrentState() != IdpFallbackLocalDB {
		t.Fatalf("expected fallback_local_db, got %s", state.CurrentState())
	}

	for range 3 {
		state.RecordFailure()
	}
	if state.CurrentState() != IdpFallbackReadonly {
		t.Fatalf("expected fallback_readonly_tokens, got %s", state.CurrentState())
	}

	for range 3 {
		state.RecordFailure()
	}
	if state.CurrentState() != IdpEmergencyBypass {
		t.Fatalf("expected emergency_bypass, got %s", state.CurrentState())
	}

	state.RecordSuccess()
	if state.CurrentState() != IdpPrimary {
		t.Fatalf("expected primary after recovery, got %s", state.CurrentState())
	}
}
