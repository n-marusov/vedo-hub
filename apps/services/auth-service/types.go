package main

import "time"

type IdpState string

const (
	IdpPrimary          IdpState = "primary"
	IdpFallbackLocalDB  IdpState = "fallback_local_db"
	IdpFallbackReadonly IdpState = "fallback_readonly_tokens"
	IdpEmergencyBypass  IdpState = "emergency_bypass"
)

type ClientTier string

const (
	TierAnonymous      ClientTier = "anonymous"
	TierFree           ClientTier = "free"
	TierPro            ClientTier = "pro"
	TierEnterprise     ClientTier = "enterprise"
	TierServiceAccount ClientTier = "service_account"
)

type TokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	GrantType    string `json:"grant_type"`
	Code         string `json:"code,omitempty"`
	RedirectURI  string `json:"redirect_uri,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

type IntrospectRequest struct {
	Token string `json:"token"`
}

type IntrospectResponse struct {
	Active   bool     `json:"active"`
	UserID   string   `json:"user_id,omitempty"`
	Roles    []string `json:"roles,omitempty"`
	TenantID string   `json:"tenant_id,omitempty"`
	Exp      int64    `json:"exp,omitempty"`
}

type SessionInfo struct {
	Active    bool     `json:"active"`
	UserID    string   `json:"user_id"`
	Roles     []string `json:"roles"`
	TenantID  string   `json:"tenant_id"`
	ExpiresAt string   `json:"expires_at"`
}

type AuthStatus struct {
	IDPState           IdpState `json:"idp_state"`
	RateLimitRemaining int      `json:"rate_limit_remaining"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}

type StoredSession struct {
	UserID       string
	Roles        []string
	TenantID     string
	TokenExpiry  time.Time
	RefreshToken string
}

type StoredRefreshToken struct {
	UserID    string
	ExpiresAt time.Time
}
