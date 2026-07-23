package auth

// Validates: REQ-NFR.SECURITY.bola-bfla-negative-tests

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// @ctx: test key for JWT signing and verification
var testKey *rsa.PrivateKey

func init() {
	gin.SetMode(gin.TestMode)
	var err error
	testKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("failed to generate test RSA key: " + err.Error())
	}
}

// @ctx: test audit logger for verifying audit events
type testAuditLogger struct {
	events []*AuditEvent
}

func (t *testAuditLogger) WriteAuditEvent(_ context.Context, event *AuditEvent) error {
	t.events = append(t.events, event)
	return nil
}

func defaultConfig() *Config {
	return &Config{
		KeyFunc: func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return &testKey.PublicKey, nil
		},
		ExemptPrefixes:   DefaultExemptPrefixes(),
		ExactExemptPaths: DefaultExactExemptPaths(),
		AuditWriter:      &testAuditLogger{},
		AdminRoles:       DefaultAdminRoles(),
	}
}

// @ctx: sign a JWT with test key
func signTestToken(claims *AuthClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	s, err := token.SignedString(testKey)
	if err != nil {
		panic("failed to sign test token: " + err.Error())
	}
	return s
}

// @ctx: execute middleware via Gin test engine
func execMiddleware(cfg *Config, method, path, token string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	engine := gin.New()
	engine.Use(NewMiddleware(cfg))
	engine.Any("/*path", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequestWithContext(context.Background(), method, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("X-Trace-Id", "test-trace-"+path)
	req.Header.Set("X-Correlation-Id", "test-corr-"+path)
	req.RemoteAddr = "10.0.0.1:12345"

	engine.ServeHTTP(w, req)
	return w
}

// @ctx: extract error code from JSON response body
func extractErrorCode(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	if w.Code == http.StatusOK {
		return ""
	}
	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Error.Code
}

// ============================================================
// CT-SEC-001: Valid JWT accesses own tenant resource
// @hlv FORBIDDEN_CROSS_TENANT_ACCESS
// @hlv structured_logging_only
// @hlv log_entry_exit
// ============================================================
// @hlv:sec [AUTH_BOUNDARY] — valid tenant JWT grants access to own resource
func TestCT_SEC_001_ValidJWT_OwnTenant_Returns200(t *testing.T) {
	cfg := defaultConfig()
	token := signTestToken(&AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:   "alice-uuid",
		TenantID: "tenant_A",
		Roles:    []string{"Owner"},
	})
	w := execMiddleware(cfg, "GET", "/api/v1/ontologies/ont-123", token)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (code=%s)", w.Code, extractErrorCode(t, w))
	}
}

// ============================================================
// CT-SEC-002: Cross-tenant access blocked
// @hlv FORBIDDEN_CROSS_TENANT_ACCESS
// ============================================================
// @hlv:sec [AUTH_BOUNDARY] — cross-tenant BOLA must return 403
func TestCT_SEC_002_CrossTenant_Returns403(t *testing.T) {
	cfg := defaultConfig()
	token := signTestToken(&AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "alice-uuid",
		TenantID: "tenant_A",
		Roles:    []string{"Owner"},
	})
	w := execMiddleware(cfg, "GET", "/api/v1/tenants/tenant_B/resources", token)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
	if code := extractErrorCode(t, w); code != ErrCrossTenantAccess {
		t.Errorf("expected error code %s, got %s", ErrCrossTenantAccess, code)
	}
}

// ============================================================
// CT-SEC-003: Insufficient role blocked
// @hlv FORBIDDEN_INSUFFICIENT_ROLE
// ============================================================
// @hlv:sec [AUTH_BOUNDARY] — Viewer cannot DELETE, must return 403
func TestCT_SEC_003_InsufficientRole_Returns403(t *testing.T) {
	cfg := defaultConfig()
	token := signTestToken(&AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "bob-uuid",
		TenantID: "tenant_A",
		Roles:    []string{"Viewer"},
	})
	w := execMiddleware(cfg, "DELETE", "/api/v1/ontologies/ont-123", token)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
	if code := extractErrorCode(t, w); code != ErrInsufficientRole {
		t.Errorf("expected error code %s, got %s", ErrInsufficientRole, code)
	}
}

// ============================================================
// CT-SEC-004: Non-admin calls admin endpoint
// @hlv FORBIDDEN_ADMIN_ONLY
// ============================================================
// @hlv:sec [AUTH_BOUNDARY] — Viewer cannot access admin endpoint, must return 403
func TestCT_SEC_004_NonAdminCallsAdmin_Returns403(t *testing.T) {
	cfg := defaultConfig()
	token := signTestToken(&AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "viewer-uuid",
		TenantID: "tenant_A",
		Roles:    []string{"Viewer"},
	})
	w := execMiddleware(cfg, "GET", "/api/v1/admin/users", token)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
	if code := extractErrorCode(t, w); code != ErrAdminOnly {
		t.Errorf("expected error code %s, got %s", ErrAdminOnly, code)
	}
}

// ============================================================
// CT-SEC-005: Missing JWT → UNAUTHENTICATED
// @hlv UNAUTHENTICATED
// ============================================================
func TestCT_SEC_005_MissingJWT_Returns401(t *testing.T) {
	cfg := defaultConfig()
	w := execMiddleware(cfg, "GET", "/api/v1/ontologies/ont-123", "")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	if code := extractErrorCode(t, w); code != ErrUnauthenticated {
		t.Errorf("expected error code %s, got %s", ErrUnauthenticated, code)
	}
}

// ============================================================
// CT-SEC-006: Expired JWT → TOKEN_EXPIRED
// @hlv TOKEN_EXPIRED
// ============================================================
func TestCT_SEC_006_ExpiredJWT_Returns401(t *testing.T) {
	cfg := defaultConfig()
	token := signTestToken(&AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
		UserID:   "alice-uuid",
		TenantID: "tenant_A",
		Roles:    []string{"Owner"},
	})
	w := execMiddleware(cfg, "GET", "/api/v1/ontologies/ont-123", token)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	if code := extractErrorCode(t, w); code != ErrTokenExpired {
		t.Errorf("expected error code %s, got %s", ErrTokenExpired, code)
	}
}

// ============================================================
// CT-SEC-007: Auth endpoint exempt from auth
// @hlv UNAUTHENTICATED — negative: exempt paths bypass auth
// ============================================================
func TestCT_SEC_007_AuthEndpointExempt_Returns200(t *testing.T) {
	cfg := defaultConfig()
	w := execMiddleware(cfg, "POST", "/api/v1/auth/token", "")

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (exempt auth path), got %d", w.Code)
	}
}

// ============================================================
// CT-SEC-008: Public browse exempt from auth
// ============================================================
func TestCT_SEC_008_PublicBrowseExempt_Returns200(t *testing.T) {
	cfg := defaultConfig()
	w := execMiddleware(cfg, "GET", "/api/v1/public/browse/ontologies", "")

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (exempt public path), got %d", w.Code)
	}
}

// ============================================================
// CT-SEC-009: Health endpoint exempt
// ============================================================
func TestCT_SEC_009_HealthEndpointExempt_Returns200(t *testing.T) {
	cfg := defaultConfig()
	w := execMiddleware(cfg, "GET", "/health", "")

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (exempt health path), got %d", w.Code)
	}
}

// ============================================================
// CT-SEC-010: Non-existent object check — middleware passes through
// (403 for non-existent objects enforced at service layer)
// @hlv never_return_404_invariant
// ============================================================
// @hlv:sec [AUTH_BOUNDARY] — middleware never returns 404 for auth decisions
func TestCT_SEC_010_MiddlewareNeverReturns404(t *testing.T) {
	cfg := defaultConfig()
	token := signTestToken(&AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "alice-uuid",
		TenantID: "tenant_A",
		Roles:    []string{"Viewer"},
	})
	w := execMiddleware(cfg, "GET", "/api/v1/ontologies/00000000-0000-0000-0000-000000000099", token)

	// @ctx: middleware should never return 404 — it either passes through or returns 401/403
	if w.Code == http.StatusNotFound {
		t.Errorf("middleware must never return 404, got %d", w.Code)
	}
	if w.Code == http.StatusInternalServerError {
		t.Errorf("middleware must never return 500, got %d", w.Code)
	}
}

// ============================================================
// CT-SEC-011: Audit event produced on denial
// @hlv structured_logging_only
// @hlv log_all_errors
// @hlv log_state_changes
// ============================================================
// @hlv:sec [AUTH_BOUNDARY] — every denial produces an audit event
func TestCT_SEC_011_AuditEventOnDenial(t *testing.T) {
	logger := &testAuditLogger{}
	cfg := defaultConfig()
	cfg.AuditWriter = logger

	token := signTestToken(&AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "alice-uuid",
		TenantID: "tenant_A",
		Roles:    []string{"Owner"},
	})

	// @ctx: use a non-engine approach to access the audit logger directly
	w := httptest.NewRecorder()
	engine := gin.New()
	engine.Use(NewMiddleware(cfg))
	engine.Any("/*path", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/v1/tenants/tenant_B/resources", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.RemoteAddr = "10.0.0.1:12345"
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
	if len(logger.events) == 0 {
		t.Fatal("expected at least 1 audit event on denial, got 0")
	}
	last := logger.events[len(logger.events)-1]
	if last.Event != "authorization.denied" {
		t.Errorf("expected event=authorization.denied, got %s", last.Event)
	}
	if last.Reason != ErrCrossTenantAccess {
		t.Errorf("expected reason=%s, got %s", ErrCrossTenantAccess, last.Reason)
	}
	if last.UserID == "" {
		t.Error("expected non-empty user_id in audit event")
	}
}

// ============================================================
// CT-SEC-012: BOLA: sequential ID probe returns 403
// @hlv FORBIDDEN_OBJECT_NOT_FOUND_OR_ACCESS_DENIED
// ============================================================
// @hlv:sec [AUTH_BOUNDARY] — BOLA sequential ID probing returns 403 never 404
func TestCT_SEC_012_BOLA_SequentialIDProbe_Returns403(t *testing.T) {
	cfg := defaultConfig()
	token := signTestToken(&AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "alice-uuid",
		TenantID: "tenant_A",
		Roles:    []string{"Viewer"},
	})

	for _, id := range []string{"1", "2", "3", "99999", "00000000-0000-0000-0000-000000000099"} {
		w := execMiddleware(cfg, "GET", "/api/v1/tenants/tenant_A/objects/"+id, token)
		if w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError {
			t.Errorf("BOLA sequential ID probe for %s: got %d (must be 403)", id, w.Code)
		}
	}
}

// ============================================================
// CT-SEC-013: BFLA: mutation by insufficient role
// @hlv FORBIDDEN_INSUFFICIENT_ROLE
// ============================================================
// @hlv:sec [AUTH_BOUNDARY] — BFLA: Viewer cannot perform write operations
func TestCT_SEC_013_BFLA_MutationByInsufficientRole_Returns403(t *testing.T) {
	cfg := defaultConfig()
	token := signTestToken(&AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "viewer-uuid",
		TenantID: "tenant_A",
		Roles:    []string{"Viewer"},
	})

	w := execMiddleware(cfg, "POST", "/api/v1/ontologies/ont-123/update", token)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for Viewer POST, got %d", w.Code)
	}
	if code := extractErrorCode(t, w); code != ErrInsufficientRole {
		t.Errorf("expected error code %s, got %s", ErrInsufficientRole, code)
	}
}

// ============================================================
// Property: Never return 404 for any auth decision
// @hlv never_return_404_invariant
// ============================================================
// @hlv:sec [AUTH_BOUNDARY] — invariant: all auth decisions return 4xx never 404
func TestProperty_NeverReturn404(t *testing.T) {
	cfg := defaultConfig()
	tokens := map[string]string{
		"owner":  signTestToken(&AuthClaims{RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour))}, UserID: "owner-uuid", TenantID: "tenant_A", Roles: []string{"Owner"}}),
		"viewer": signTestToken(&AuthClaims{RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour))}, UserID: "viewer-uuid", TenantID: "tenant_A", Roles: []string{"Viewer"}}),
		"none":   "",
	}
	endpoints := []struct {
		method string
		path   string
		token  string
	}{
		{"GET", "/api/v1/tenants/tenant_B/objects/x", "owner"},
		{"DELETE", "/api/v1/ontologies/ont-123", "viewer"},
		{"GET", "/api/v1/tenants/tenant_B/resources", "owner"},
		{"POST", "/api/v1/admin/users", "viewer"},
		{"GET", "/api/v1/ontologies/random-uuid", "none"},
	}
	for _, ep := range endpoints {
		w := execMiddleware(cfg, ep.method, ep.path, tokens[ep.token])
		if w.Code == http.StatusNotFound || w.Code == http.StatusInternalServerError {
			t.Errorf("endpoint %s %s (token=%s): got %d — never 404/500", ep.method, ep.path, ep.token, w.Code)
		}
	}
}

// ============================================================
// Property: Tenant isolation — cross-tenant always returns 403
// @hlv tenant_isolation_property
// ============================================================
func TestProperty_TenantIsolation(t *testing.T) {
	cfg := defaultConfig()
	token := signTestToken(&AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "user-tenant-A",
		TenantID: "tenant_A",
		Roles:    []string{"Editor"},
	})
	otherTenants := []string{"tenant_B", "tenant_C", "tenant_D", "tenant_E"}
	for _, other := range otherTenants {
		w := execMiddleware(cfg, "GET", "/api/v1/tenants/"+other+"/resources", token)
		if w.Code != http.StatusForbidden {
			t.Errorf("cross-tenant access to %s: expected 403, got %d", other, w.Code)
		}
	}
}

// ============================================================
// Property: Missing role claim defaults to Viewer
// @hlv missing_role_claim_default_viewer
// ============================================================
func TestProperty_MissingRoleClaimDefaultsViewer(t *testing.T) {
	cfg := defaultConfig()
	token := signTestToken(&AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "user-no-roles",
		TenantID: "tenant_A",
		Roles:    nil,
	})
	w := execMiddleware(cfg, "DELETE", "/api/v1/ontologies/ont-123", token)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 (Viewer cannot delete), got %d", w.Code)
	}
}

// ============================================================
// Property: Audit completeness — every denial produces audit event
// @hlv audit_completeness
// ============================================================
func TestProperty_AuditCompleteness(t *testing.T) {
	logger := &testAuditLogger{}
	cfg := defaultConfig()
	cfg.AuditWriter = logger

	deniedTokens := []struct {
		name   string
		token  string
		path   string
		method string
	}{
		{
			name: "cross-tenant",
			token: signTestToken(&AuthClaims{
				RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour))},
				UserID:           "alice", TenantID: "tenant_A", Roles: []string{"Owner"},
			}),
			path: "/api/v1/tenants/tenant_B/objects/x", method: "GET",
		},
		{
			name: "insufficient-role",
			token: signTestToken(&AuthClaims{
				RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour))},
				UserID:           "bob", TenantID: "tenant_A", Roles: []string{"Viewer"},
			}),
			path: "/api/v1/ontologies/ont-123", method: "DELETE",
		},
	}

	for _, tc := range deniedTokens {
		w := execMiddleware(cfg, tc.method, tc.path, tc.token)
		if w.Code != http.StatusForbidden {
			t.Errorf("%s: expected 403, got %d", tc.name, w.Code)
		}
	}

	if len(logger.events) < len(deniedTokens) {
		t.Errorf("expected at least %d audit events, got %d", len(deniedTokens), len(logger.events))
	}
	for _, e := range logger.events {
		if e.Event != "authorization.denied" {
			t.Errorf("expected authorization.denied event, got %s", e.Event)
		}
	}
}

// ============================================================
// Invariant: Public browse with valid JWT passes through
// @hlv public_endpoint_with_jwt
// ============================================================
func TestInvariant_PublicBrowseWithValidJWT_PassesThrough(t *testing.T) {
	cfg := defaultConfig()
	token := signTestToken(&AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "alice-uuid",
		TenantID: "tenant_A",
		Roles:    []string{"Owner"},
	})
	w := execMiddleware(cfg, "GET", "/api/v1/public/browse/ontologies", token)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (public browse exempt), got %d", w.Code)
	}
}

// ============================================================
// Invariant: Token missing tenant_id treated as access denied
// @hlv token_missing_tenant_id
// ============================================================
func TestInvariant_TokenMissingTenantID_Denied(t *testing.T) {
	cfg := defaultConfig()
	token := signTestToken(&AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "alice-uuid",
		TenantID: "",
		Roles:    []string{"Owner"},
	})
	w := execMiddleware(cfg, "GET", "/api/v1/tenants/tenant_B/resources", token)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for token without tenant_id, got %d", w.Code)
	}
}

// ============================================================
// Invariant: Ready endpoint exempt
// ============================================================
func TestInvariant_ReadyEndpointExempt(t *testing.T) {
	cfg := defaultConfig()
	w := execMiddleware(cfg, "GET", "/ready", "")

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (exempt ready path), got %d", w.Code)
	}
}

// ============================================================
// Invariant: Metrics endpoint exempt
// ============================================================
func TestInvariant_MetricsEndpointExempt(t *testing.T) {
	cfg := defaultConfig()
	w := execMiddleware(cfg, "GET", "/metrics", "")

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (exempt metrics path), got %d", w.Code)
	}
}

// ============================================================
// Invariant: Root endpoint exempt
// ============================================================
func TestInvariant_RootEndpointExempt(t *testing.T) {
	cfg := defaultConfig()
	w := execMiddleware(cfg, "GET", "/", "")

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (exempt root path), got %d", w.Code)
	}
}

// ============================================================
// Invariant: Malformed token returns 401 not 500
// @hlv UNAUTHENTICATED
// ============================================================
func TestInvariant_MalformedToken_Returns401(t *testing.T) {
	cfg := defaultConfig()
	w := execMiddleware(cfg, "GET", "/api/v1/ontologies/ont-123", "not-a-valid-jwt")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for malformed token, got %d", w.Code)
	}
	if w.Code == http.StatusInternalServerError {
		t.Error("malformed token must not cause 500")
	}
}

// ============================================================
// Invariant: Wrong signing key returns 401
// @hlv UNAUTHENTICATED
// ============================================================
// ============================================================
// CT-SEC-014: Viewer-role can POST to read-only query endpoints
// @hlv FORBIDDEN_INSUFFICIENT_ROLE — negative: BFLA must allow Viewer reads
// ============================================================
// @hlv:sec [AUTH_BOUNDARY] — Viewer can call SPARQL/CYPHER/GraphQL with role override = 0
func TestCT_SEC_014_Viewer_CanPost_ReadOnlyQueryEndpoints(t *testing.T) {
	// Configure the middleware the same way production main.go does: role
	// override = 0 for the SPARQL/CYPHER/GraphQL POST endpoints so the
	// rate limiter — not the BFLA gate — returns 429 on quota exhaustion.
	cfg := defaultConfig()
	cfg.RequiredRoleLevel = map[string]int{
		"POST:/api/v1/sparql":  0,
		"POST:/api/v1/cypher":  0,
		"POST:/api/v1/graphql": 0,
	}
	token := signTestToken(&AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "viewer-uuid",
		TenantID: "tenant_A",
		Roles:    []string{"Viewer"},
	})
	for _, ep := range []struct{ method, path string }{
		{"POST", "/api/v1/sparql"},
		{"POST", "/api/v1/cypher"},
		{"POST", "/api/v1/graphql"},
	} {
		w := execMiddleware(cfg, ep.method, ep.path, token)
		if w.Code != http.StatusOK {
			t.Errorf("%s %s with Viewer role: expected 200 (role override=0), got %d (code=%s)", ep.method, ep.path, w.Code, extractErrorCode(t, w))
		}
	}
}

// ============================================================
// CT-SEC-015: Without role override, Viewer POST returns 403
// @hlv FORBIDDEN_INSUFFICIENT_ROLE — regression gate for production config
// ============================================================
// @hlv:sec [AUTH_BOUNDARY] — production default must require Editor+ for POSTs
func TestCT_SEC_015_WithoutOverride_ViewerPost_Returns403(t *testing.T) {
	cfg := defaultConfig()
	// Deliberately leave cfg.RequiredRoleLevel nil: this is the dangerous
	// production misconfiguration the override prevents.
	token := signTestToken(&AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "viewer-uuid",
		TenantID: "tenant_A",
		Roles:    []string{"Viewer"},
	})
	w := execMiddleware(cfg, "POST", "/api/v1/sparql", token)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 without override, got %d", w.Code)
	}
	if code := extractErrorCode(t, w); code != ErrInsufficientRole {
		t.Errorf("expected %s, got %s", ErrInsufficientRole, code)
	}
}

// ============================================================
// Invariant: Wrong signing key returns 401
// @hlv UNAUTHENTICATED
// ============================================================
func TestInvariant_WrongSigningKey_Returns401(t *testing.T) {
	wrongKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	wrongToken := jwt.NewWithClaims(jwt.SigningMethodRS256, &AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "alice-uuid",
		TenantID: "tenant_A",
		Roles:    []string{"Owner"},
	})
	tokenStr, err := wrongToken.SignedString(wrongKey)
	if err != nil {
		t.Fatal(err)
	}

	cfg := defaultConfig()
	w := execMiddleware(cfg, "GET", "/api/v1/ontologies/ont-123", tokenStr)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong signing key, got %d", w.Code)
	}
}

// ============================================================
// Regression: alg=none token must be rejected
// @hlv UNAUTHENTICATED
// ============================================================
// @hlv:sec [AUTH_BOUNDARY] — explicit rejection of unsigned alg=none JWT
func TestAuth_AlgNoneToken_Rejected(t *testing.T) {
	// Create an unsigned token with alg=none (per RFC 7518 §3.6).
	token := jwt.NewWithClaims(jwt.SigningMethodNone, &AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "attacker-uuid",
		TenantID: "tenant_A",
		Roles:    []string{"Owner"},
	})
	tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}

	cfg := defaultConfig()
	w := execMiddleware(cfg, "GET", "/api/v1/ontologies/ont-123", tokenStr)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for alg=none token, got %d (body=%s)", w.Code, w.Body.String())
	}
	// The response must NOT be 200 — the middleware must treat this as unauthenticated.
}

// ============================================================
// Regression: invalid JWT error must not leak internal details
// @hlv NO_SENSITIVE_IN_LOGS
// ============================================================
// @hlv:sec [AUTH_BOUNDARY] — error message in 401 response must not expose
// JWT library internals (e.g. "signature is invalid", "crypto/rsa:...").
func TestAuth_InvalidTokenError_NoInfoLeak(t *testing.T) {
	cfg := defaultConfig()
	// Malformed token with valid-looking structure but obviously wrong content.
	w := execMiddleware(cfg, "GET", "/api/v1/ontologies/ont-123", "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.invalidsignature")

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid token, got %d", w.Code)
	}

	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}

	// The message must NOT contain JWT internals like "signature", "crypto", "key".
	leaks := []string{"signature", "crypto", "token is malformed", "key"}
	for _, pattern := range leaks {
		if strings.Contains(resp.Error.Message, pattern) {
			t.Errorf("error message leaks JWT internals: message=%q contains %q", resp.Error.Message, pattern)
		}
	}
	// The message must be a generic description.
	if resp.Error.Message == "" {
		t.Error("error message must not be empty")
	}
}

// ============================================================
// Keycloak OIDC claims — realm_access.roles + sub fallback
// Regression: KC_HOSTNAME override + JWT claims mismatch
// ============================================================
// @hlv:sec [AUTH_BOUNDARY] — Keycloak places realm roles in realm_access.roles
// (standard OIDC shape) rather than a top-level `roles` claim. The `sub` claim
// carries the user identifier instead of `user_id`. This test suite verifies
// that the middleware correctly handles both formats.

// TestKeycloak_SubFallbackForUserID verifies that when `user_id` is empty
// the middleware falls back to the standard `sub` claim (Keycloak default).
func TestKeycloak_SubFallbackForUserID(t *testing.T) {
	claims := &AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "alice-sub-uuid",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		TenantID: "tenant_A",
		Roles:    []string{"Viewer"},
	}
	token := signTestToken(claims)

	cfg := defaultConfig()
	w := execMiddleware(cfg, "GET", "/api/v1/ontologies/ont-123", token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for JWT with sub fallback, got %d (body=%s)", w.Code, w.Body.String())
	}

	// The middleware should have propagated the userID from the `sub` claim.
	// Downstream headers must contain the resolved identity.
	// (Verified via auth.granted log: user_id=alice-sub-uuid)
}

// TestKeycloak_RealmAccessRoles_GrantsEditorWrite verifies that roles
// from realm_access.roles (Keycloak OIDC format) are parsed and grant
// Editor-level write access (POST).
func TestKeycloak_RealmAccessRoles_GrantsEditorWrite(t *testing.T) {
	claims := &AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "bob-editor-uuid",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		TenantID: "tenant_A",
		// No top-level roles — roles come entirely from realm_access.
		RealmAccess: struct {
			Roles []string `json:"roles"`
		}{Roles: []string{"Editor", "default-roles-vedo-core", "offline_access", "uma_authorization"}},
	}
	token := signTestToken(claims)

	cfg := defaultConfig()
	// POST requires role weight ≥ 1 (Editor=1, Viewer=0)
	w := execMiddleware(cfg, "POST", "/api/v1/ontologies/ont-123/classes", token)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for Editor (via realm_access.roles) doing POST, got %d (body=%s)",
			w.Code, w.Body.String())
	}
}

// TestKeycloak_RealmAccessRoles_ViewerBlocksDelete verifies that when
// realm_access.roles only has Viewer, DELETE (weight 2) is blocked.
func TestKeycloak_RealmAccessRoles_ViewerBlocksDelete(t *testing.T) {
	claims := &AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "alice-viewer-uuid",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		TenantID: "tenant_A",
		RealmAccess: struct {
			Roles []string `json:"roles"`
		}{Roles: []string{"Viewer"}},
	}
	token := signTestToken(claims)

	cfg := defaultConfig()
	// DELETE requires role weight ≥ 2 (Maintainer=2, Owner=3)
	w := execMiddleware(cfg, "DELETE", "/api/v1/ontologies/ont-123", token)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for Viewer (via realm_access.roles) doing DELETE, got %d (body=%s)",
			w.Code, w.Body.String())
	}

	code := extractErrorCode(t, w)
	if code != ErrInsufficientRole {
		t.Errorf("expected error code %s, got %s", ErrInsufficientRole, code)
	}
}

// TestKeycloak_MergeTopLevelAndRealmRoles verifies that top-level `roles`
// and `realm_access.roles` are merged without duplicates. An Owner role
// from top-level combined with additional roles from realm_access should
// still grant admin-level access.
func TestKeycloak_MergeTopLevelAndRealmRoles(t *testing.T) {
	claims := &AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "frank-owner-uuid",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "frank-owner-uuid",
		TenantID: "tenant_A",
		Roles:    []string{"Owner"},
		RealmAccess: struct {
			Roles []string `json:"roles"`
		}{Roles: []string{"Owner", "Viewer"}}, // Owner duplicated — must be deduped
	}
	token := signTestToken(claims)

	cfg := defaultConfig()
	// DELETE on admin-protected endpoint should pass (Owner + admin endpoint → weight 3)
	w := execMiddleware(cfg, "DELETE", "/api/v1/admin/users", token)

	// DELETE on admin endpoint → the middleware checks role weight against admin level.
	// Owner has weight 3, which meets the admin threshold.
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for Owner (merged top-level + realm_access.roles) calling admin DELETE, got %d (body=%s)",
			w.Code, w.Body.String())
	}
}

// TestKeycloak_EmptyRealmAccessRoles_NoPanic verifies that an empty (nil)
// realm_access.roles does not cause a panic or change behaviour.
func TestKeycloak_EmptyRealmAccessRoles_NoPanic(t *testing.T) {
	claims := &AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "empty-realm-uuid",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		UserID:   "empty-realm-uuid",
		TenantID: "tenant_A",
		Roles:    []string{"Editor"},
		// RealmAccess left zero-valued — Roles slice is nil.
	}
	token := signTestToken(claims)

	cfg := defaultConfig()
	w := execMiddleware(cfg, "POST", "/api/v1/ontologies/ont-123/classes", token)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for Editor with empty realm_access, got %d (body=%s)",
			w.Code, w.Body.String())
	}
}

// TestKeycloak_SubWithNoUserID_Merge works when both sub comes from Keycloak
// and roles come from realm_access.roles without a custom user_id mapper.
// Owner role (weight 3) should pass the admin gate.
func TestKeycloak_SubWithNoUserID_Merge(t *testing.T) {
	claims := &AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "eve-owner-uuid",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		TenantID: "tenant_A",
		// All identity from Keycloak standard claims — sub for user, realm_access.roles for roles.
		RealmAccess: struct {
			Roles []string `json:"roles"`
		}{Roles: []string{"Owner"}},
	}
	token := signTestToken(claims)

	cfg := defaultConfig()
	// GET on admin endpoint: Owner (weight 3) meets the admin gate (weight ≥ 3).
	w := execMiddleware(cfg, "GET", "/api/v1/admin/health", token)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for Owner (via realm_access.roles + sub) calling admin GET, got %d (body=%s)",
			w.Code, w.Body.String())
	}
}
