// @hlv:artifact auth-keyfunc implements SEC-AUTHZ-GATES-001
// @hlv:artifact auth-keyfunc-source verifies API-AUTH-JWT
package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT verification configuration sources, resolved in order: a local
// PEM-encoded RSA public key file takes precedence over a remote JWKS URL.
// The gateway refuses to start (see main.go) when neither is configured so
// the auth boundary can never silently degrade to "accept all tokens".
const (
	EnvJWTPublicKeyPath = "JWT_PUBLIC_KEY_PATH"
	EnvKeycloakJWKSURL  = "KEYCLOAK_JWKS_URL"
	// EnvJWTDevPublicKeyPEM is an escape hatch used by some startup tests
	// that inject the keypair via env instead of a file path. It is never
	// documented in .env.example and MUST NOT be set in production.
	EnvJWTDevPublicKeyPEM = "JWT_DEV_PUBLIC_KEY_PEM"

	jwksDefaultTTL = 15 * time.Minute
	jwksFetchLimit = 1 << 20 // 1 MiB
)

// LoadKeyFuncFromEnv constructs the production jwt.Keyfunc used by the
// authentication middleware. Resolution order:
//
//  1. JWT_PUBLIC_KEY_PATH  — PEM-encoded RS256 public key file on disk.
//  2. KEYCLOAK_JWKS_URL    — JSON Web Key Set URL (refreshed on rotation).
//  3. JWT_DEV_PUBLIC_KEY_PEM — raw PEM bytes from env (tests only).
//
// Returns an error when no source is configured so the caller can refuse to
// boot the gateway with disabled auth. The returned keyfunc verifies that
// every token uses the RS256 signing method and that the kid header maps to a
// known key in the JWKS set (when the JWKS source is used).
func LoadKeyFuncFromEnv() (jwt.Keyfunc, error) {
	if pemBytes := os.Getenv(EnvJWTDevPublicKeyPEM); pemBytes != "" {
		slog.Warn("auth.keyfunc.dev_pem",
			"reason",
			"JWT_DEV_PUBLIC_KEY_PEM is set; this must never be enabled in production",
		)
		pub, err := parseRSAPublicKeyPEM([]byte(pemBytes))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", EnvJWTDevPublicKeyPEM, err)
		}
		slog.Info("auth.keyfunc.loaded", "source", "dev_pem")
		return DefaultKeyFunc(pub), nil
	}

	if path := os.Getenv(EnvJWTPublicKeyPath); path != "" {
		pub, err := loadRSAPublicKeyPEM(path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", EnvJWTPublicKeyPath, err)
		}
		slog.Info("auth.keyfunc.loaded", "source", "pem_file", "path", path)
		return DefaultKeyFunc(pub), nil
	}

	if jwksURL := os.Getenv(EnvKeycloakJWKSURL); jwksURL != "" {
		kf, err := newJWKSKeyFunc(jwksURL, jwksDefaultTTL)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", EnvKeycloakJWKSURL, err)
		}
		slog.Info("auth.keyfunc.loaded", "source", "jwks", "url", jwksURL)
		return kf, nil
	}

	return nil, errors.New(
		"JWT verification not configured: set " + EnvJWTPublicKeyPath +
			" or " + EnvKeycloakJWKSURL,
	)
}

// loadRSAPublicKeyPEM reads a PEM-encoded RSA public key from disk and returns
// the parsed *rsa.PublicKey. Accepts both "PUBLIC KEY" (SubjectPublicKeyInfo)
// and PKCS#1 "RSA PUBLIC KEY" PEM blocks.
func loadRSAPublicKeyPEM(path string) (*rsa.PublicKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseRSAPublicKeyPEM(raw)
}

// parseRSAPublicKeyPEM parses a PEM-encoded RSA public key from a byte slice.
func parseRSAPublicKeyPEM(raw []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("no PEM block found in public key")
	}

	switch block.Type {
	case "RSA PUBLIC KEY":
		pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse PKCS#1 public key: %w", err)
		}
		return pub, nil
	case "PUBLIC KEY":
		pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse PKIX public key: %w", err)
		}
		pub, ok := pubAny.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("public key is not RSA")
		}
		return pub, nil
	default:
		return nil, fmt.Errorf("unexpected PEM block type %q (want PUBLIC KEY or RSA PUBLIC KEY)", block.Type)
	}
}

// jwksKeyFunc resolves tokens against a cached JWKS document fetched from a
// Keycloak (or any RFC 7517 compliant) endpoint. The cache is refreshed on a
// fixed TTL or when a token's kid is missing from the current cache.
type jwksKeyFunc struct {
	url    string
	ttl    time.Duration
	client *http.Client

	mu            sync.RWMutex
	keysByKid     map[string]*rsa.PublicKey
	keysByEmptyID []*rsa.PublicKey // keys without a kid (legacy issuers)
	lastFetch     time.Time
}

// jwk represents the subset of a JSON Web Key needed for RS256 verification.
type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwksDocument struct {
	Keys []jwk `json:"keys"`
}

func newJWKSKeyFunc(jwksURL string, ttl time.Duration) (jwt.Keyfunc, error) {
	kf := &jwksKeyFunc{
		url:    jwksURL,
		ttl:    ttl,
		client: &http.Client{Timeout: 10 * time.Second},
	}
	if err := kf.refresh(); err != nil {
		return nil, err
	}
	return kf.verify, nil
}

// verify is the jwt.Keyfunc closure handed to jwt.ParseWithClaims. It enforces
// the RS256 signing method and resolves the kid header against the JWKS cache.
func (j *jwksKeyFunc) verify(token *jwt.Token) (any, error) {
	if alg, _ := token.Header["alg"].(string); strings.EqualFold(alg, "none") {
		slog.Warn("auth.keyfunc.jwks_alg_none_rejected", "kid", token.Header["kid"])
		return nil, jwt.ErrSignatureInvalid
	}
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, jwt.ErrSignatureInvalid
	}

	kid, _ := token.Header["kid"].(string)

	j.mu.RLock()
	pub, ok := j.keysByKid[kid]
	if !ok && kid == "" {
		// Legacy issuer without kid; accept if exactly one RSA key exists.
		if len(j.keysByEmptyID) == 1 {
			pub = j.keysByEmptyID[0]
			ok = true
		}
	}
	j.mu.RUnlock()
	if ok {
		return pub, nil
	}

	// Key rotation: refresh once, then retry. If still missing the token
	// is treated as untrusted.
	if err := j.refresh(); err != nil {
		slog.Warn("auth.keyfunc.jwks_refresh_failed", "url", j.url, "error", err)
		return nil, jwt.ErrSignatureInvalid
	}

	j.mu.RLock()
	defer j.mu.RUnlock()
	if pub, ok := j.keysByKid[kid]; ok {
		return pub, nil
	}
	if kid == "" && len(j.keysByEmptyID) == 1 {
		return j.keysByEmptyID[0], nil
	}

	slog.Warn("auth.keyfunc.kid_not_found", "kid", kid, "url", j.url)
	return nil, jwt.ErrSignatureInvalid
}

// refresh fetches the JWKS document and rebuilds the in-memory key index.
func (j *jwksKeyFunc) refresh() error {
	req, err := http.NewRequest(http.MethodGet, j.url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := j.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, jwksFetchLimit))
	if err != nil {
		return err
	}

	var doc jwksDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return fmt.Errorf("parse JWKS JSON: %w", err)
	}

	if len(doc.Keys) == 0 {
		return errors.New("JWKS document contains no keys")
	}

	byKid := make(map[string]*rsa.PublicKey, len(doc.Keys))
	var noKid []*rsa.PublicKey
	for _, k := range doc.Keys {
		if k.Kty != "RSA" {
			continue
		}
		if k.Use != "" && k.Use != "sig" {
			continue
		}
		if k.Alg != "" && k.Alg != "RS256" {
			continue
		}
		pub, err := jwkToRSAPublicKey(k)
		if err != nil {
			slog.Warn("auth.keyfunc.jwk_skipped", "kid", k.Kid, "error", err)
			continue
		}
		if k.Kid == "" {
			noKid = append(noKid, pub)
			continue
		}
		byKid[k.Kid] = pub
	}

	if len(byKid) == 0 && len(noKid) == 0 {
		return errors.New("JWKS document contains no usable RS256 signatures")
	}

	j.mu.Lock()
	j.keysByKid = byKid
	j.keysByEmptyID = noKid
	j.lastFetch = time.Now()
	j.mu.Unlock()

	slog.Info("auth.keyfunc.jwks_refreshed",
		"url", j.url,
		"kid_count", len(byKid),
		"anon_count", len(noKid),
	)
	return nil
}

// jwkToRSAPublicKey converts a JWK to a Go *rsa.PublicKey. n and e are
// base64url-encoded integers per RFC 7518 §6.3.
func jwkToRSAPublicKey(k jwk) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("decode n: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("decode e: %w", err)
	}
	eInt := new(big.Int).SetBytes(eBytes).Int64()
	if eInt < 3 || eInt%2 == 0 {
		return nil, fmt.Errorf("invalid public exponent %d", eInt)
	}
	n := new(big.Int).SetBytes(nBytes)
	// Guard against malformed (negative) modulus bytes.
	if n.Sign() <= 0 {
		return nil, errors.New("invalid modulus")
	}
	return &rsa.PublicKey{N: n, E: int(eInt)}, nil
}
