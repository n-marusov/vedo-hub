package main

import (
	"log"
	"net/http"
	"time"
)

func handleGetSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
			Error: APIError{Code: "METHOD_NOT_ALLOWED", Message: "Only GET allowed"},
		})
		return
	}

	tokenStr := extractBearerToken(r)
	if tokenStr == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{
			Error: APIError{Code: "INVALID_TOKEN", Message: "Missing authorization header"},
		})
		return
	}

	mu.RLock()
	session, exists := tokenStore[tokenStr]
	mu.RUnlock()

	if !exists || time.Now().After(session.TokenExpiry) {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error: APIError{Code: "SESSION_NOT_FOUND", Message: "No active session for given token"},
		})
		return
	}

	resp := SessionInfo{
		Active:    true,
		UserID:    session.UserID,
		Roles:     session.Roles,
		TenantID:  session.TenantID,
		ExpiresAt: session.TokenExpiry.Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, resp)
	log.Printf("Session retrieved user=%s", redactToken(session.UserID))
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
			Error: APIError{Code: "METHOD_NOT_ALLOWED", Message: "Only POST allowed"},
		})
		return
	}

	tokenStr := extractBearerToken(r)
	if tokenStr == "" {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{
			Error: APIError{Code: "INVALID_TOKEN", Message: "Missing authorization header"},
		})
		return
	}

	mu.Lock()
	session, exists := tokenStore[tokenStr]
	if exists {
		delete(tokenStore, tokenStr)
		if session.RefreshToken != "" {
			delete(refreshStore, session.RefreshToken)
		}
	}
	mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{"status": "logged_out"})
	log.Printf("Session terminated token=%s", redactToken(tokenStr))
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}
