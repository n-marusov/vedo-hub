package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func handleRefresh(w http.ResponseWriter, r *http.Request) {
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

	if req.RefreshToken == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: APIError{Code: "INVALID_REQUEST", Message: "refresh_token is required"},
		})
		return
	}

	mu.Lock()
	stored, exists := refreshStore[req.RefreshToken]
	if !exists {
		mu.Unlock()
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{
			Error: APIError{Code: "REFRESH_FAILED", Message: "Refresh token is invalid or expired"},
		})
		return
	}

	if time.Now().After(stored.ExpiresAt) {
		delete(refreshStore, req.RefreshToken)
		mu.Unlock()
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{
			Error: APIError{Code: "REFRESH_FAILED", Message: "Refresh token has expired"},
		})
		return
	}

	delete(refreshStore, req.RefreshToken)

	newAccessToken := generateToken()
	newRefreshToken := generateToken()
	now := time.Now()

	refreshStore[newRefreshToken] = &StoredRefreshToken{
		UserID:    stored.UserID,
		ExpiresAt: now.Add(time.Duration(refreshTokenTTL) * time.Second),
	}
	tokenStore[newAccessToken] = &StoredSession{
		UserID:       stored.UserID,
		Roles:        rolesForUser(stored.UserID),
		TenantID:     tenantForUser(stored.UserID),
		TokenExpiry:  now.Add(time.Duration(accessTokenTTL) * time.Second),
		RefreshToken: newRefreshToken,
	}
	mu.Unlock()

	resp := map[string]any{
		"status": "authenticated",
		"token": TokenResponse{
			AccessToken:  newAccessToken,
			ExpiresIn:    accessTokenTTL,
			RefreshToken: newRefreshToken,
			TokenType:    "Bearer",
		},
		"idp_state": globalIdpState.CurrentState(),
	}

	writeJSON(w, http.StatusOK, resp)
	log.Printf("Token rotated user=%s", redactToken(stored.UserID))
}
