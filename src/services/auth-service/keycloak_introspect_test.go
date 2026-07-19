package main

// Validates: REQ-NFR.SECURITY.organization-access-model

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// @hlv [AUTH-RBAC-001] JWT issued by Keycloak introspects via IdP
// @hlv:sec [AUTH_BOUNDARY] Token introspection honouring external IdP
func TestIntrospect_KeycloakJWT_Active(t *testing.T) {
	setupTest()

	// Mock Keycloak token introspection endpoint: returns active=true for a
	// fixed JWT, inactive otherwise. Detects hint=access_token and client creds.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.FormValue("client_id") != "vedo-core-api" {
			t.Errorf("expected client_id=vedo-core-api, got %s", r.FormValue("client_id"))
		}
		if r.FormValue("token_type_hint") != "access_token" {
			t.Errorf("expected token_type_hint=access_token, got %s", r.FormValue("token_type_hint"))
		}
		switch r.FormValue("token") {
		case "good.jwt.token":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"active": true,
				"sub":    "kc-user-42",
				"exp":    int64(1899999999),
				"realm_access": map[string]any{
					"roles": []string{"viewer", "editor"},
				},
			})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{"active": false})
		}
	}))
	defer srv.Close()

	prev := withEnv("KEYCLOAK_URL", srv.URL)
	defer prev()
	prevSec := withEnv("KEYCLOAK_CLIENT_ID", "vedo-core-api")
	defer prevSec()
	prevSecret := withEnv("KEYCLOAK_CLIENT_SECRET", "vedo-core-api-secret")
	defer prevSecret()

	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(IntrospectRequest{Token: "good.jwt.token"})
	req := httptest.NewRequest("POST", "/api/v1/auth/token/introspect", &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/token/introspect", handleIntrospect)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var resp IntrospectResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.Active || resp.UserID != "kc-user-42" || resp.Exp != 1899999999 {
		t.Fatalf("unexpected introspect result: %+v", resp)
	}
	if len(resp.Roles) != 2 || resp.Roles[0] != "viewer" {
		t.Fatalf("unexpected roles: %+v", resp.Roles)
	}
}

// @hlv:sec [AUTH_BOUNDARY] Keycloak calling inactive tokens rejected
func TestIntrospect_KeycloakJWT_Inactive_Returns401(t *testing.T) {
	setupTest()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"active": false})
	}))
	defer srv.Close()

	withEnv("KEYCLOAK_URL", srv.URL)()
	withEnv("KEYCLOAK_CLIENT_ID", "vedo-core-api")
	withEnv("KEYCLOAK_CLIENT_SECRET", "vedo-core-api-secret")

	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(IntrospectRequest{Token: "stale.jwt.token"})
	req := httptest.NewRequest("POST", "/api/v1/auth/token/introspect", &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/token/introspect", handleIntrospect)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	var errResp ErrorResponse
	_ = json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Error.Code != "INVALID_TOKEN" {
		t.Fatalf("expected INVALID_TOKEN, got %s", errResp.Error.Code)
	}
}

// @hlv:sec [AUTH_BOUNDARY] Keycloak unreachable degrades IdP state and 401s
func TestIntrospect_KeycloakUnreachable_DegradesIdP(t *testing.T) {
	setupTest()
	// Point Keycloak at an unreachable port to force a network error.
	restore := withEnv("KEYCLOAK_URL", "http://127.0.0.1:1")
	defer restore()
	withEnv("KEYCLOAK_CLIENT_ID", "vedo-core-api")
	withEnv("KEYCLOAK_CLIENT_SECRET", "vedo-core-api-secret")

	// IdpDegradeThreshold (3) must be crossed before CurrentState flips, so
	// repeat the introspect until the IdP recorder counters stack up.
	var lastCode int
	for i := 0; i < 5; i++ {
		var buf bytes.Buffer
		_ = json.NewEncoder(&buf).Encode(IntrospectRequest{Token: "any.jwt.token"})
		req := httptest.NewRequest("POST", "/api/v1/auth/token/introspect", &buf)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux := http.NewServeMux()
		mux.HandleFunc("/api/v1/auth/token/introspect", handleIntrospect)
		mux.ServeHTTP(w, req)
		lastCode = w.Code
	}
	if lastCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", lastCode)
	}
	if globalIdpState.CurrentState() == IdpPrimary {
		t.Fatal("expected IdP to degrade after Keycloak failure")
	}
}

// @hlv [AUTH-RBAC-001] Synthetic `vtok_` tokens stay local when Keycloak env unset
func TestIntrospect_SyntheticToken_NoKeycloakEnv_LocalLookup(t *testing.T) {
	setupTest()
	// KEYCLOAK_URL unset — synthetic token must use local store only.
	if keycloakIntrospectURL() != "" {
		t.Fatalf("KEYCLOAK_URL should be empty when env unset")
	}
	if isJWT("vtok_deadbeef") {
		t.Fatal("vtok_ must not be classified as JWT")
	}
	// Sanity: a local token still introspects without touching Keycloak.
	tokenResp := doRequest("POST", "/api/v1/auth/token", TokenRequest{
		GrantType: "authorization_code",
		Code:      "valid-code-123",
	})
	body := readBody(t, tokenResp)
	tok := body["token"].(map[string]any)
	accessToken := tok["access_token"].(string)

	if !strings.HasPrefix(accessToken, "vtok_") {
		t.Fatalf("expected synthetic token prefix, got %s", accessToken)
	}
	introResp := doRequest("POST", "/api/v1/auth/token/introspect", IntrospectRequest{Token: accessToken})
	if introResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", introResp.Code)
	}
}

// withEnv sets os.Setenv and returns a restore closure.
func withEnv(key, value string) func() {
	prev, had := os.LookupEnv(key)
	_ = os.Setenv(key, value)
	return func() {
		if had {
			_ = os.Setenv(key, prev)
		} else {
			_ = os.Unsetenv(key)
		}
	}
}
