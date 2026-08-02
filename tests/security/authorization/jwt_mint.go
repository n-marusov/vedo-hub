//go:build integration

// JWT minting helper for RBAC security tests.
//
// The test stack (docker-compose.test.yml) configures the API Gateway with
// JWT_DEV_PUBLIC_KEY_PEM, which takes precedence over the remote Keycloak
// JWKS. Real Keycloak tokens are therefore REJECTED in the test environment —
// the gateway only accepts RS256 tokens signed with tests/e2e/scripts/test-jwt-key.pem.
//
// These tests mint their own dev-signed JWTs (same pattern as the Playwright
// E2E global-setup) with the claims each RBAC scenario needs (roles, tenant).
// Token minting is a Keycloak concern in production; in the test stack the
// dev key is the established trust anchor.

package authorization

import (
	"crypto/rsa"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// testUserSpec maps a test user alias to the JWT claims required by the
// RBAC scenario (role ladder + tenant boundary).
type testUserSpec struct {
	userID   string
	roles    []string
	tenantID string
}

// userSpecs defines the claim shape for every alias used by the suite.
// Roles follow the project role ladder (see api-gateway auth.roleWeight):
// viewer=10, reporter=20, editor=30, maintainer=40, owner=50.
var userSpecs = map[string]testUserSpec{
	userViewerA:     {userID: "u_viewer_A", roles: []string{"viewer"}, tenantID: "tenant_A"},
	userReporterA:   {userID: "u_reporter_A", roles: []string{"reporter"}, tenantID: "tenant_A"},
	userEditorA:     {userID: "u_editor_A", roles: []string{"editor"}, tenantID: "tenant_A"},
	userMaintainerA: {userID: "u_maintainer_A", roles: []string{"maintainer"}, tenantID: "tenant_A"},
	userOwnerA:      {userID: "u_owner_A", roles: []string{"owner"}, tenantID: "tenant_A"},
	userOutsiderA:   {userID: "u_outsider_A", roles: []string{}, tenantID: "tenant_A"},
	userViewerB:     {userID: "u_viewer_B", roles: []string{"viewer"}, tenantID: "tenant_B"},
	userOwnerB:      {userID: "u_owner_B", roles: []string{"owner"}, tenantID: "tenant_B"},
}

// devPrivateKeyPath resolves the test signing key. Order:
//  1. VEDO_TEST_JWT_KEY env var (explicit override)
//  2. tests/e2e/scripts/test-jwt-key.pem relative to the repo root
//     (walked up from the current working directory)
func devPrivateKeyPath() string {
	if p := os.Getenv("VEDO_TEST_JWT_KEY"); p != "" {
		return p
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "test-jwt-key.pem"
	}
	// Walk up to the repo root looking for tests/e2e/scripts/test-jwt-key.pem.
	for dir := cwd; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, "tests", "e2e", "scripts", "test-jwt-key.pem")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "test-jwt-key.pem"
}

// loadDevPrivateKey reads and parses the RS256 test private key.
func loadDevPrivateKey() (*rsa.PrivateKey, error) {
	path := devPrivateKeyPath()
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return jwt.ParseRSAPrivateKeyFromPEM(raw)
}

// mintJWT signs an RS256 token with the dev test key.
// Returns the signed token string.
func mintJWT(spec testUserSpec, ttl time.Duration) (string, error) {
	key, err := loadDevPrivateKey()
	if err != nil {
		return "", err
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":       spec.userID,
		"user_id":   spec.userID,
		"tenant_id": spec.tenantID,
		"roles":     spec.roles,
		"realm_access": map[string]any{
			"roles": append([]string{"default-roles-vedo-core"}, spec.roles...),
		},
		"iat": now.Unix(),
		"exp": now.Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(key)
}
