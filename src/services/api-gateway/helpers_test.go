// Package gateway_test provides executable integration tests for the API Gateway.
//
// These tests do not require Keycloak or running upstream services. They stand
// up a full Gin engine identical to the production wiring via RegisterRoutes,
// backed by httptest upstreams and a self-signed RSA keypair for the auth
// middleware. Each test file covers one M1 contract area:
//
//   - auth_integration_test.go    → JWT enforcement (401/403 paths)
//   - proxy_integration_test.go   → header propagation and upstream routing
//   - query_integration_test.go    → SPARQL/CYPHER read-only enforcement + LIMIT
//   - rate_limit_integration_test.go → sliding-window 429 behaviour
//
// Tests run as part of `go test ./...` (no build tag required). They use a
// throwaway RSA keypair and in-memory httptest upstreams, so they stay fast
// and do not touch the network.
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"vedo-core/src/services/api-gateway/auth"
	corsmw "vedo-core/src/services/api-gateway/middleware"
)

// testEnv is the configuration shared across all integration tests in this
// package. Each test receives its own gin.Engine plus mock upstreams so the
// gateway can be exercised end-to-end without external dependencies.
type testEnv struct {
	router         *gin.Engine
	ontologyServer *httptest.Server
	versioningSrv  *httptest.Server
	privateKey     *rsa.PrivateKey
	cleanup        func()
}

// newTestEnv returns a fresh gateway env. Callers can mutate the upstream
// handlers before issuing requests. `t.Cleanup(env.cleanup)` is mandatory.
func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	// Generate a throwaway RSA keypair; sign JWTs with the private key and
	// pass the public key to the auth middleware KeyFunc.
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}

	ontologyUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Drain the request body so the keep-alive transport does not return EOF
		// to the proxy when the handler ignores body bytes (e.g. POST /ontologies).
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		// Default OK response — handlers may swap the handler in setupUpstream.
		_ = json.NewEncoder(w).Encode(gin.H{
			"method":   r.Method,
			"path":     r.URL.Path,
			"trace_id": r.Header.Get("X-Trace-Id"),
			"user_id":  r.Header.Get("X-User-Id"),
		})
	}))

	versioningUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(gin.H{
			"method":     r.Method,
			"path":       r.URL.Path,
			"versioning": true,
			"trace_id":   r.Header.Get("X-Trace-Id"),
			"user_id":    r.Header.Get("X-User-Id"),
			"user_roles": r.Header.Get("X-User-Roles"),
		})
	}))

	// Point the gateway at the mock upstreams.
	_ = os.Setenv("ONTOLOGY_SERVICE_URL", ontologyUpstream.URL)
	_ = os.Setenv("VERSIONING_SERVICE_URL", versioningUpstream.URL)
	_ = os.Setenv("UPSTREAM_TIMEOUT", "2")

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsmw.CORS())

	// Auth middleware with the test RSA public key.
	authCfg := &auth.Config{
		KeyFunc:          testKeyFunc(privKey.Public().(*rsa.PublicKey)),
		ExemptPrefixes:   append(auth.DefaultExemptPrefixes(), "/api/v1/public/"),
		ExactExemptPaths: auth.DefaultExactExemptPaths(),
		AuditWriter:      &auth.SlogAuditWriter{},
		AdminRoles:       auth.DefaultAdminRoles(),
		RequiredRoleLevel: map[string]int{
			// Read-only SPARQL/CYPHER endpoints must let Viewer-role / anonymous
			// clients authenticate normally (level 0) so the rate limiter — not
			// the BFLA gate — returns 429 when the tier quota is exceeded.
			"POST:/api/v1/sparql":  0,
			"POST:/api/v1/cypher":  0,
			"POST:/api/v1/graphql": 0,
		},
	}
	r.Use(auth.NewMiddleware(authCfg))
	r.Use(corsmw.Timeout(2 * time.Second))

	// RegisterRoutes uses the env vars set above to wire proxies.
	RegisterRoutes(r)

	cleanup := func() {
		ontologyUpstream.Close()
		versioningUpstream.Close()
	}

	return &testEnv{
		router:         r,
		ontologyServer: ontologyUpstream,
		versioningSrv:  versioningUpstream,
		privateKey:     privKey,
		cleanup:        cleanup,
	}
}

// testKeyFunc returns an auth.KeyFunc that validates RS256 tokens against the
// provided public key. This mirrors the production behaviour but without a
// keycloak dependency.
func testKeyFunc(pub *rsa.PublicKey) jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return pub, nil
	}
}

// signTestJWT returns a bearer token signed with the test private key. The
// token carries the provided roles and user ID (and an empty tenant so the
// `validateTenant` BOLA check passes for paths without a tenant segment).
func (e *testEnv) signTestJWT(t *testing.T, userID string, roles []string, expiresAt time.Time) string {
	t.Helper()
	claims := &auth.AuthClaims{
		UserID:   userID,
		TenantID: "",
		Roles:    roles,
	}
	claims.ExpiresAt = jwt.NewNumericDate(expiresAt)
	claims.IssuedAt = jwt.NewNumericDate(time.Now())
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(e.privateKey)
	if err != nil {
		t.Fatalf("sign test JWT: %v", err)
	}
	return signed
}

// bearer wraps signTestJWT as a ready-to-use Authorization header value.
func (e *testEnv) bearer(t *testing.T, userID string, roles []string) string {
	return "Bearer " + e.signTestJWT(t, userID, roles, time.Now().Add(1*time.Hour))
}

// writeJSON encodes the payload as JSON and writes it with the given status.
// Helper used by upstream handlers that need to swap in custom responses.
func writeJSON(w http.ResponseWriter, status int, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(payload)
}
