package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	tokenStore      = map[string]*StoredSession{}
	refreshStore    = map[string]*StoredRefreshToken{}
	mu              sync.RWMutex
	accessTokenTTL  = 3600
	refreshTokenTTL = 86400
)

func generateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return "vtok_" + hex.EncodeToString(b)
}

func handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
			Error: APIError{Code: "METHOD_NOT_ALLOWED", Message: "Only POST allowed"},
		})
		return
	}

	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: APIError{Code: "INVALID_REQUEST", Message: "Invalid request body"},
		})
		return
	}

	if req.GrantType != "authorization_code" && req.GrantType != "client_credentials" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: APIError{Code: "UNSUPPORTED_GRANT_TYPE", Message: "grant_type must be authorization_code or client_credentials"},
		})
		return
	}

	if req.GrantType == "authorization_code" && req.Code == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: APIError{Code: "INVALID_GRANT", Message: "authorization_code grant requires a code"},
		})
		return
	}

	if req.GrantType == "authorization_code" && req.Code == "invalid-code" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{
			Error: APIError{Code: "INVALID_TOKEN", Message: "Authorization code is invalid or expired"},
		})
		return
	}

	userID := resolveUserID(&req)
	roles := rolesForUser(userID)
	tenantID := tenantForUser(userID)

	accessToken := generateToken()
	refreshToken := generateToken()

	now := time.Now()

	mu.Lock()
	tokenStore[accessToken] = &StoredSession{
		UserID:       userID,
		Roles:        roles,
		TenantID:     tenantID,
		TokenExpiry:  now.Add(time.Duration(accessTokenTTL) * time.Second),
		RefreshToken: refreshToken,
	}
	refreshStore[refreshToken] = &StoredRefreshToken{
		UserID:    userID,
		ExpiresAt: now.Add(time.Duration(refreshTokenTTL) * time.Second),
	}
	mu.Unlock()

	remaining, _ := globalRateLimiter.allow(r.RemoteAddr+":token", resolveClientTier(r))

	resp := map[string]any{
		"status": "authenticated",
		"token": TokenResponse{
			AccessToken:  accessToken,
			ExpiresIn:    accessTokenTTL,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
		},
		"idp_state":            globalIdpState.CurrentState(),
		"rate_limit_remaining": remaining,
	}

	writeJSON(w, http.StatusOK, resp)
	log.Printf("Token issued user=%s grant_type=%s", redactToken(userID), req.GrantType)
}

func resolveUserID(req *TokenRequest) string {
	if req.GrantType == "client_credentials" {
		return req.ClientID + "-service"
	}
	return "user-" + req.Code
}

func rolesForUser(userID string) []string {
	return []string{"Viewer", "Editor"}
}

func tenantForUser(userID string) string {
	return "tenant_default"
}

func redactToken(s string) string {
	if len(s) <= 8 {
		return s
	}
	return s[:4] + "***" + s[len(s)-4:]
}
