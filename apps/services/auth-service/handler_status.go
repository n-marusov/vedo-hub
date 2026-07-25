package main

import (
	"net/http"
)

func handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
			Error: APIError{Code: "METHOD_NOT_ALLOWED", Message: "Only GET allowed"},
		})
		return
	}

	tier := resolveClientTier(r)
	_, remaining := globalRateLimiter.allow(r.RemoteAddr+":status", tier)

	resp := AuthStatus{
		IDPState:           globalIdpState.CurrentState(),
		RateLimitRemaining: remaining,
	}

	writeJSON(w, http.StatusOK, resp)
}
