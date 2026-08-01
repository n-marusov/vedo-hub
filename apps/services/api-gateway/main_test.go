package main

// Validates: REQ-NFR.SECURITY.bola-bfla-negative-tests

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"vedo-core/src/services/api-gateway/auth"
)

// clearAuthEnv unsets every JWT verification env var so LoadKeyFuncFromEnv
// behaves deterministically. Each variable is restored automatically by
// t.Setenv when the test cleans up.
func clearAuthEnv(t *testing.T) {
	t.Helper()
	t.Setenv(auth.EnvJWTPublicKeyPath, "")
	t.Setenv(auth.EnvKeycloakJWKSURL, "")
	t.Setenv(auth.EnvJWTDevPublicKeyPEM, "")
}

// writeTempPEMFile writes the RSA public key as a PEM file and returns its
// absolute path. The file is removed when the test finishes.
func writeTempPEMFile(t *testing.T, pub *rsa.PublicKey) string {
	t.Helper()
	dir := t.TempDir()
	asn1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatalf("marshal PKIX public key: %v", err)
	}
	pemBlock := &pem.Block{Type: "PUBLIC KEY", Bytes: asn1}
	path := filepath.Join(dir, "test-jwt.pem")
	if err := os.WriteFile(path, pem.EncodeToMemory(pemBlock), 0o600); err != nil {
		t.Fatalf("write pem: %v", err)
	}
	return path
}

// jwksFor returns the JWKS document (in the order Go's map yields) for the
// given public key. kid is fixed so the verify path can resolve it.
func jwksFor(t *testing.T, pub *rsa.PublicKey, kid string) []byte {
	t.Helper()
	doc := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"kid": kid,
				"use": "sig",
				"alg": "RS256",
				"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
			},
		},
	}
	b, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal jwks: %v", err)
	}
	return b
}

func TestLoadKeyFuncFromEnv_NoConfiguration_ReturnsError(t *testing.T) {
	clearAuthEnv(t)
	_, err := auth.LoadKeyFuncFromEnv()
	if err == nil {
		t.Fatal("expected error when JWT verification is unconfigured")
	}
	if !strings.Contains(err.Error(), "JWT verification not configured") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestLoadKeyFuncFromEnv_PEMFile_Success(t *testing.T) {
	clearAuthEnv(t)
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	path := writeTempPEMFile(t, &priv.PublicKey)
	t.Setenv(auth.EnvJWTPublicKeyPath, path)

	// Ensure no other source shadows the PEM file.
	t.Setenv(auth.EnvKeycloakJWKSURL, "")
	t.Setenv(auth.EnvJWTDevPublicKeyPEM, "")

	keyFunc, err := auth.LoadKeyFuncFromEnv()
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if keyFunc == nil {
		t.Fatal("expected non-nil keyfunc")
	}

	// Verify the keyfunc round-trips a signed token.
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		Issuer:    "test",
		ExpiresAt: nil,
	})
	signed, err := tok.SignedString(priv)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	parsed, err := jwt.Parse(signed, keyFunc)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if !parsed.Valid {
		t.Fatal("expected token to validate against PEM-based keyfunc")
	}
}

func TestLoadKeyFuncFromEnv_JWKS_Success(t *testing.T) {
	clearAuthEnv(t)
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	const kid = "test-kid-1"

	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwksFor(t, &priv.PublicKey, kid))
	}))
	t.Cleanup(jwksServer.Close)

	t.Setenv(auth.EnvKeycloakJWKSURL, jwksServer.URL)
	t.Setenv(auth.EnvJWTPublicKeyPath, "")

	keyFunc, err := auth.LoadKeyFuncFromEnv()
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if keyFunc == nil {
		t.Fatal("expected non-nil keyfunc")
	}

	// Token with the matching kid must validate; an unknown kid must reject.
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{Issuer: "test"})
	tok.Header["kid"] = kid
	signed, err := tok.SignedString(priv)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	if _, err := jwt.Parse(signed, keyFunc); err != nil {
		t.Fatalf("verify token with known kid: %v", err)
	}

	tok.Header["kid"] = "unknown-kid"
	bad, err := tok.SignedString(priv)
	if err != nil {
		t.Fatalf("sign bad token: %v", err)
	}
	if _, err := jwt.Parse(bad, keyFunc); err == nil {
		t.Fatal("expected verification to fail for unknown kid")
	}
}

func TestLoadKeyFuncFromEnv_NonRSAToken_Rejected(t *testing.T) {
	clearAuthEnv(t)
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	path := writeTempPEMFile(t, &priv.PublicKey)
	t.Setenv(auth.EnvJWTPublicKeyPath, path)

	keyFunc, err := auth.LoadKeyFuncFromEnv()
	if err != nil {
		t.Fatalf("load keyfunc: %v", err)
	}

	// HMAC token with any key must be rejected — the gateway only trusts RS256.
	hmacTok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Issuer: "attacker"})
	signed, err := hmacTok.SignedString([]byte("any-secret"))
	if err != nil {
		t.Fatalf("sign HMAC token: %v", err)
	}
	if _, err := jwt.Parse(signed, keyFunc); err == nil {
		t.Fatal("expected HMAC-signed token to be rejected with RS256 keyfunc")
	}
}

func TestOverrideSummary_Empty_ReturnsNone(t *testing.T) {
	if got := overrideSummary(map[string]int{}); got != "<none>" {
		t.Fatalf("expected <none>, got %q", got)
	}
}

func TestEmbeddedOpenAPISpec_ValidJSON(t *testing.T) {
	var doc map[string]any
	if err := json.Unmarshal(openAPISpec, &doc); err != nil {
		t.Fatalf("embedded docs/openapi.json is not valid JSON: %v", err)
	}

	required := []string{"openapi", "paths", "components"}
	for _, key := range required {
		if _, ok := doc[key]; !ok {
			t.Errorf("embedded openapi.json missing required key: %q", key)
		}
	}

	paths, ok := doc["paths"].(map[string]any)
	if !ok {
		t.Fatal("embedded openapi.json 'paths' is not an object")
	}
	if len(paths) == 0 {
		t.Error("embedded openapi.json 'paths' is empty — expected at least one API path entry")
	}
}

func TestOverrideSummary_DeterministicOrder(t *testing.T) {
	overrides := map[string]int{
		"POST:/api/v1/sparql":  0,
		"POST:/api/v1/cypher":  0,
		"POST:/api/v1/graphql": 0,
	}
	// Map iteration order is non-deterministic in Go; assert set equality,
	// not element-by-element ordering.
	gotEntries := strings.Split(overrideSummary(overrides), ",")
	if len(gotEntries) != len(overrides) {
		t.Fatalf("expected %d entries, got %d (%q)", len(overrides), len(gotEntries), gotEntries)
	}
	wantSet := make(map[string]struct{}, len(overrides))
	for k, v := range overrides {
		wantSet[fmt.Sprintf("%s=%d", k, v)] = struct{}{}
	}
	for _, entry := range gotEntries {
		if _, ok := wantSet[entry]; !ok {
			t.Errorf("unexpected entry %q not in %v", entry, wantSet)
		}
	}
}
