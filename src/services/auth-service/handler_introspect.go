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

	// 1) Local auth-service tokens (`vtok_<hex>`) — used by tests and the
	//    synthetic local store. This is the original behaviour and must keep
	//    working without Keycloak.
	mu.RLock()
	session, exists := tokenStore[req.Token]
	mu.RUnlock()

	if exists {
		if time.Now().After(session.TokenExpiry) {
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{
				Error: APIError{Code: "TOKEN_EXPIRED", Message: "Token has expired"},
			})
			return
		}
		respondIntrospectActive(w, &IntrospectResponse{
			Active:   true,
			UserID:   session.UserID,
			Roles:    session.Roles,
			TenantID: session.TenantID,
			Exp:      session.TokenExpiry.Unix(),
		})
		log.Printf("Token introspected (local) user=%s active=true", redactToken(session.UserID))
		return
	}

	// 2) JWTs are treated as Keycloak-issued access tokens and forwarded to
	//    the realm's token-introspection endpoint. Synthetic tokens without a
	//    local match fall through to the 401 branch below.
	if isJWT(req.Token) {
		out, err := introspectWithKeycloak(req.Token)
		if err != nil {
			globalIdpState.RecordFailure()
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{
				Error: APIError{Code: "INVALID_TOKEN", Message: "Token is invalid or revoked"},
			})
			return
		}
		globalIdpState.RecordSuccess()
		respondIntrospectActive(w, out)
		log.Printf("Token introspected (keycloak) user=%s active=true", redactToken(out.UserID))
		return
	}

	// 3) Unknown synthetic token, no JWT shape, no local match.
	writeJSON(w, http.StatusUnauthorized, ErrorResponse{
		Error: APIError{Code: "INVALID_TOKEN", Message: "Token is invalid or revoked"},
	})
}

func respondIntrospectActive(w http.ResponseWriter, resp *IntrospectResponse) {
	w.Header().Set("X-Token-Status", "valid")
	writeJSON(w, http.StatusOK, resp)
}
