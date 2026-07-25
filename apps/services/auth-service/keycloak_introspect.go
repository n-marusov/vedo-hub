package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// keycloakIntrospectClient wraps RFC 7662 token introspection against a Keycloak
// realm endpoint. It is used as a secondary path in handleIntrospect for JWTs
// issued by Keycloak (frontend OIDC flow); tokens minted by auth-service itself
// (the `vtok_` prefix) stay on the local in-memory store and never reach here.
//
// All env vars default to empty; when KEYCLOAK_URL is empty the caller treats
// introspection as unavailable and returns 401 to the client (the service keeps
// relying on the local synthetic-token store for tests/dev without Keycloak).

const keycloakIntrospectTimeout = 5 * time.Second

type keycloakIntrospectResponse struct {
	Active      bool   `json:"active"`
	Username    string `json:"username"`
	Subject     string `json:"sub"`
	Exp         int64  `json:"exp"`
	RealmAccess struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	// tenant_id is not a standard Keycloak claim; we expose it for parity with
	// the local StoredSession shape when present (custom mapper in realm).
	TenantID string `json:"tenant_id,omitempty"`
}

func keycloakIntrospectURL() string {
	base := strings.TrimRight(os.Getenv("KEYCLOAK_URL"), "/")
	realm := os.Getenv("KEYCLOAK_REALM")
	if realm == "" {
		realm = "vedo-core"
	}
	if base == "" {
		return ""
	}
	return base + "/realms/" + realm + "/protocol/openid-connect/token/introspect"
}

// isJWT distinguishes Keycloak-issued access tokens from auth-service's own
// `vtok_` synthetic tokens. Keycloak access tokens are compact JWS (JWE is
// never used by Keycloak OIDC) — three base64url segments separated by `.`.
// `vtok_<hex>` has no `.` separators.
func isJWT(token string) bool {
	parts := strings.Split(token, ".")
	return len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != ""
}

// introspectWithKeycloak returns a server-facing IntrospectResponse when the
// token is active according to Keycloak. If the IdP is unreachable, the token
// is expired, or the response marks it inactive, returns nil + error so the
// handler can fall back to the 401 INVALID_TOKEN path and record IdP failures.
func introspectWithKeycloak(token string) (*IntrospectResponse, error) {
	endpoint := keycloakIntrospectURL()
	if endpoint == "" {
		return nil, errKeycloakUnavailable
	}

	clientID := os.Getenv("KEYCLOAK_CLIENT_ID")
	clientSecret := os.Getenv("KEYCLOAK_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return nil, errKeycloakUnavailable
	}

	form := url.Values{}
	form.Set("token", token)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("token_type_hint", "access_token")

	ctx, cancel := context.WithTimeout(context.Background(), keycloakIntrospectTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := keycloakHTTPClient.Do(req)
	if err != nil {
		return nil, errKeycloakUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errKeycloakUnavailable
	}

	var kc keycloakIntrospectResponse
	if err := json.NewDecoder(resp.Body).Decode(&kc); err != nil {
		return nil, err
	}
	if !kc.Active {
		return nil, errTokenInactive
	}

	roles := kc.RealmAccess.Roles
	if roles == nil {
		roles = []string{}
	}

	out := &IntrospectResponse{
		Active:   true,
		UserID:   kc.Subject,
		Roles:    roles,
		TenantID: kc.TenantID,
		Exp:      kc.Exp,
	}
	if out.UserID == "" {
		out.UserID = kc.Username
	}
	if out.TenantID == "" {
		out.TenantID = "default"
	}
	return out, nil
}

var errKeycloakUnavailable = &introspectError{msg: "Keycloak introspection endpoint unavailable"}
var errTokenInactive = &introspectError{msg: "token inactive"}

type introspectError struct{ msg string }

func (e *introspectError) Error() string { return e.msg }

// keycloakHTTPClient wires a per-process HTTP client with a sane timeout so
// unit tests can substitute httptest.Server URLs via KEYCLOAK_URL without
// leaking sockets.
var keycloakHTTPClient = &http.Client{Timeout: keycloakIntrospectTimeout}
