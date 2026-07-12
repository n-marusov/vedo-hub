package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func handleIntrospect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
			Error: APIError{Code: "METHOD_NOT_ALLOWED", Message: "Only POST allowed"},
		})
		return
	}

	var req IntrospectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error: APIError{Code: "INVALID_REQUEST", Message: "Invalid request body"},
		})
		return
	}

	mu.RLock()
	session, exists := tokenStore[req.Token]
	mu.RUnlock()

	if !exists {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{
			Error: APIError{Code: "INVALID_TOKEN", Message: "Token is invalid or revoked"},
		})
		return
	}

	if time.Now().After(session.TokenExpiry) {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{
			Error: APIError{Code: "TOKEN_EXPIRED", Message: "Token has expired"},
		})
		return
	}

	resp := IntrospectResponse{
		Active:   true,
		UserID:   session.UserID,
		Roles:    session.Roles,
		TenantID: session.TenantID,
		Exp:      session.TokenExpiry.Unix(),
	}

	w.Header().Set("X-Token-Status", "valid")
	writeJSON(w, http.StatusOK, resp)
	log.Printf("Token introspected user=%s active=true", redactToken(session.UserID))
}
