// @hlv:artifact auth-middleware implements SEC-AUTHZ-GATES-001
// @hlv:artifact auth-middleware implements API-REST-001
// @hlv:artifact tests-auth-middleware verifies SEC-AUTHZ-GATES-001
package auth

import (
	"context"
	"crypto/rsa"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// @ctx: error codes from SEC-AUTHZ-GATES-001 contract
const (
	ErrCrossTenantAccess      = "FORBIDDEN_CROSS_TENANT_ACCESS"
	ErrInsufficientRole       = "FORBIDDEN_INSUFFICIENT_ROLE"
	ErrAdminOnly              = "FORBIDDEN_ADMIN_ONLY"
	ErrObjectNotFoundOrDenied = "FORBIDDEN_OBJECT_NOT_FOUND_OR_ACCESS_DENIED"
	ErrUnauthenticated        = "UNAUTHENTICATED"
	ErrTokenExpired           = "TOKEN_EXPIRED"
)

// @ctx: role types from glossary AuthRole
type AuthRole string

const (
	RoleViewer       AuthRole = "Viewer"
	RoleEditor       AuthRole = "Editor"
	RoleMaintainer   AuthRole = "Maintainer"
	RoleOwner        AuthRole = "Owner"
	RoleSupportEng   AuthRole = "SupportEngineer"
	RoleSRE          AuthRole = "SRE"
	RoleSecurityLead AuthRole = "SecurityLead"
	RoleProductOwner AuthRole = "ProductOwner"
)

// @ctx: role weight for hierarchy comparison (BFLA enforcement)
// Keys are lowercase (matching Keycloak realm roles) — see vedo-core-realm.json.
// resolveEffectiveRole normalizes input roles via strings.ToLower before lookup.
var roleWeight = map[AuthRole]int{
	"viewer":          0,
	"editor":          1,
	"reviewer":        1, // writes trigger review_required instead of direct allow
	"maintainer":      2,
	"owner":           3,
	"admin":           3, // full admin privileges
	"service":         3, // system-to-system; same permissions as admin
	"supportengineer": 2,
	"sre":             2,
	"securitylead":    3,
	"productowner":    2,
}

// @hlv:sec [AUTH_BOUNDARY] — JWT claims parsed from auth header
// Keycloak OIDC places roles inside realm_access.roles (standard OIDC shape)
// rather than a top-level `roles` claim. The UserID field falls back to the
// standard `sub` claim when `user_id` is absent.
type AuthClaims struct {
	jwt.RegisteredClaims
	UserID      string   `json:"user_id"`
	TenantID    string   `json:"tenant_id"`
	Roles       []string `json:"roles"`
	RealmAccess struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
}

// @ctx: middleware configuration
type Config struct {
	KeyFunc           jwt.Keyfunc
	ExemptPrefixes    []string
	ExactExemptPaths  []string
	AuditWriter       AuditLogger
	AdminRoles        []AuthRole
	RequiredRoleLevel map[string]int
}

// @hlv:sec [AUTH_BOUNDARY] — audit event structure for authorization decisions
type AuditEvent struct {
	Event             string `json:"event"`
	Reason            string `json:"reason"`
	UserID            string `json:"user_id"`
	TenantIDRequested string `json:"tenant_id_requested"`
	ObjectType        string `json:"object_type"`
	ObjectID          string `json:"object_id"`
	SourceIP          string `json:"source_ip"`
	Timestamp         string `json:"timestamp"`
	TraceID           string `json:"trace_id,omitempty"`
}

// @hlv:sec [AUTH_BOUNDARY] — audit logger interface for WORM storage
type AuditLogger interface {
	WriteAuditEvent(ctx context.Context, event *AuditEvent) error
}

// @hlv:log_all_errors — slog-based audit writer
type SlogAuditWriter struct{}

func (s *SlogAuditWriter) WriteAuditEvent(ctx context.Context, event *AuditEvent) error {
	slog.WarnContext(ctx, "auth.audit",
		"event", event.Event,
		"reason", event.Reason,
		"user_id", event.UserID,
		"tenant_id_requested", event.TenantIDRequested,
		"object_type", event.ObjectType,
		"object_id", event.ObjectID,
		"source_ip", event.SourceIP,
		"timestamp", event.Timestamp,
		"trace_id", event.TraceID,
	)
	return nil
}

// @ctx: context keys for downstream handlers
type contextKey string

const (
	CtxKeyUserID     contextKey = "auth_user_id"
	CtxKeyTenantID   contextKey = "auth_tenant_id"
	CtxKeyRoles      contextKey = "auth_roles"
	CtxKeyAuthorized contextKey = "auth_authorized"
	CtxKeyTraceID    contextKey = "auth_trace_id"
	CtxKeyJWTToken   contextKey = "auth_jwt_token"
)

// @hlv:structured_logging_only
// @hlv:request_correlation
// @hlv:log_entry_exit
// @hlv:log_all_errors
// @hlv:sec [INPUT_VALIDATION] — JWT extraction and validation from HTTP header
// @hlv:sec [AUTH_BOUNDARY] — authorization boundary for every gateway request
func NewMiddleware(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		traceID := c.GetHeader("X-Trace-Id")
		if traceID == "" {
			traceID = c.GetHeader("X-Correlation-Id")
		}
		sourceIP := c.ClientIP()

		slog.InfoContext(c.Request.Context(), "auth.middleware.enter",
			"path", path,
			"method", method,
			"trace_id", traceID,
			"source_ip", sourceIP,
		)

		// @ctx: exemption check — auth endpoints, public browse, health/ready
		if isExempt(path, cfg.ExemptPrefixes, cfg.ExactExemptPaths) {
			slog.DebugContext(c.Request.Context(), "auth.exempt_path",
				"path", path,
				"trace_id", traceID,
			)
			c.Set(string(CtxKeyAuthorized), true)
			c.Next()
			slog.InfoContext(c.Request.Context(), "auth.middleware.exit",
				"path", path,
				"status", c.Writer.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
				"trace_id", traceID,
			)
			return
		}

		// @hlv:sec [INPUT_VALIDATION] — extract Bearer token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			writeAuthError(c, http.StatusUnauthorized, ErrUnauthenticated, "Missing or invalid Authorization header")
			logDenied(c, path, traceID, ErrUnauthenticated, start)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// @hlv:sec [SECRET_HANDLING] — token value must not leak to logs
		tokenHash := redactToken(tokenString)

		// @hlv:sec [INPUT_VALIDATION] — JWT parsing and validation
		claims := &AuthClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, cfg.KeyFunc)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				writeAuthError(c, http.StatusUnauthorized, ErrTokenExpired, "Token has expired")
				logDenied(c, path, traceID, ErrTokenExpired, start)
			} else {
				slog.WarnContext(c.Request.Context(), "auth.jwt_error", "jwt_error", err.Error(), "trace_id", traceID)
				writeAuthError(c, http.StatusUnauthorized, ErrUnauthenticated, "Invalid or expired token")
				logDenied(c, path, traceID, ErrUnauthenticated, start)
			}
			return
		}
		if !token.Valid {
			writeAuthError(c, http.StatusUnauthorized, ErrUnauthenticated, "Token validation failed")
			logDenied(c, path, traceID, ErrUnauthenticated, start)
			return
		}

		userID := claims.UserID
		if userID == "" {
			// Keycloak uses the standard `sub` claim; fall back when
			// `user_id` is missing (e.g., no custom protocol mapper).
			if sub, err := token.Claims.GetSubject(); err == nil {
				userID = sub
			}
		}
		tenantID := claims.TenantID
		roles := claims.Roles
		if roles == nil {
			roles = []string{}
		}
		// Keycloak places realm roles in realm_access.roles per OIDC.
		// Merge with any top-level roles already present.
		if len(claims.RealmAccess.Roles) > 0 {
			seen := make(map[string]bool, len(roles)+len(claims.RealmAccess.Roles))
			for _, r := range roles {
				seen[r] = true
			}
			for _, r := range claims.RealmAccess.Roles {
				if !seen[r] {
					roles = append(roles, r)
					seen[r] = true
				}
			}
		}
		effectiveRole := resolveEffectiveRole(roles)
		_ = tokenHash

		// @hlv:sec [AUTH_BOUNDARY] — tenant context validation (BOLA enforcement)
		requestedTenantID := extractTenantFromPath(path)
		if reason := validateTenant(requestedTenantID, tenantID); reason != "" {
			writeAuthError(c, http.StatusForbidden, reason, "Access to resource denied")
			writeAudit(cfg, &AuditEvent{
				Event:             "authorization.denied",
				Reason:            reason,
				UserID:            userID,
				TenantIDRequested: requestedTenantID,
				ObjectType:        extractObjectType(path),
				ObjectID:          extractObjectID(path),
				SourceIP:          sourceIP,
				Timestamp:         time.Now().UTC().Format(time.RFC3339),
				TraceID:           traceID,
			})
			logDenied(c, path, traceID, reason, start)
			return
		}

		// @hlv:sec [AUTH_BOUNDARY] — BFLA: function-level role sufficiency check
		requiredLevel := methodRequiredLevel(method, path, cfg.RequiredRoleLevel)
		if effectiveRole < requiredLevel {
			if isAdminEndpoint(path) {
				writeAuthError(c, http.StatusForbidden, ErrAdminOnly, "Admin access required")
				writeAudit(cfg, &AuditEvent{
					Event:             "authorization.denied",
					Reason:            ErrAdminOnly,
					UserID:            userID,
					TenantIDRequested: requestedTenantID,
					ObjectType:        extractObjectType(path),
					ObjectID:          extractObjectID(path),
					SourceIP:          sourceIP,
					Timestamp:         time.Now().UTC().Format(time.RFC3339),
					TraceID:           traceID,
				})
				logDenied(c, path, traceID, ErrAdminOnly, start)
			} else {
				writeAudit(cfg, &AuditEvent{
					Event:             "authorization.denied",
					Reason:            ErrInsufficientRole,
					UserID:            userID,
					TenantIDRequested: requestedTenantID,
					ObjectType:        extractObjectType(path),
					ObjectID:          extractObjectID(path),
					SourceIP:          sourceIP,
					Timestamp:         time.Now().UTC().Format(time.RFC3339),
					TraceID:           traceID,
				})
				writeAuthError(c, http.StatusForbidden, ErrInsufficientRole, "Insufficient role for this operation")
				logDenied(c, path, traceID, ErrInsufficientRole, start)
			}
			return
		}

		// @hlv:log_state_changes — auth granted
		c.Set(string(CtxKeyUserID), userID)
		c.Set(string(CtxKeyTenantID), tenantID)
		c.Set(string(CtxKeyRoles), roles)
		c.Set(string(CtxKeyAuthorized), true)
		c.Set(string(CtxKeyTraceID), traceID)

		// @hlv:sec [AUTH_BOUNDARY] — inject identity headers for downstream
		// proxy forwarding. The proxy's propagateHeaders reads from r.Header,
		// so we must set them here (not just in Gin context).
		c.Request.Header.Set("X-User-Id", userID)
		c.Request.Header.Set("X-User-Roles", strings.Join(roles, ","))

		// Store the raw JWT token in the request context for gRPC auth propagation.
		// The gRPC client interceptor reads this to add authorization metadata
		// to downstream gRPC calls.
		ctx := context.WithValue(c.Request.Context(), CtxKeyJWTToken, tokenString)
		c.Request = c.Request.WithContext(ctx)

		slog.InfoContext(c.Request.Context(), "auth.granted",
			"user_id", userID,
			"tenant_id", tenantID,
			"roles", roles,
			"path", path,
			"method", method,
			"trace_id", traceID,
		)

		c.Next()

		slog.InfoContext(c.Request.Context(), "auth.middleware.exit",
			"path", path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"trace_id", traceID,
		)
	}
}

// @ctx: check if path is exempt from authentication
func isExempt(path string, prefixes []string, exactPaths []string) bool {
	for _, e := range exactPaths {
		if path == e {
			return true
		}
	}
	for _, p := range prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// @ctx: validate tenant context — returns denial reason or empty string
func validateTenant(requested, actual string) string {
	if requested == "" {
		return ""
	}
	if actual == "" {
		return ErrObjectNotFoundOrDenied
	}
	if requested != actual {
		return ErrCrossTenantAccess
	}
	return ""
}

// @ctx: determine if path is an admin endpoint
func isAdminEndpoint(path string) bool {
	if strings.HasPrefix(path, "/api/v1/admin/") {
		return true
	}
	if path == "/api/v1/users" && strings.Contains(path, "users") {
		// @ctx: user management endpoints are admin-only per CT-SEC-004
		return true
	}
	if strings.Contains(path, "/membership") {
		return true
	}
	// NOTE: org management read endpoints (GET /groups, GET /projects,
	// GET /policies, etc.) are NOT admin-only — they are accessible to any
	// authenticated user (Viewer+). The auth-service enforces fine-grained
	// authorization at the gRPC level. Mutation operations on groups/projects
	// use the default method-level gate:
	// POST→Editor, PUT→Maintainer, DELETE→Maintainer.
	return false
}

// @ctx: resolve max role weight from user's role list (max-wins)
// Normalizes roles to lowercase because Keycloak realm roles are lowercase
// (e.g. "owner", "admin") while the roleWeight map uses all-lowercase keys.
func resolveEffectiveRole(roles []string) int {
	maxWeight := -1
	for _, r := range roles {
		if w, ok := roleWeight[AuthRole(strings.ToLower(r))]; ok && w > maxWeight {
			maxWeight = w
		}
	}
	if maxWeight < 0 {
		return 0
	}
	return maxWeight
}

// @hlv:sec [AUTH_BOUNDARY] — determine required role level for method on path
func methodRequiredLevel(method string, path string, overrides map[string]int) int {
	if isAdminEndpoint(path) {
		return 3
	}
	if override, ok := overrides[method+":"+path]; ok {
		return override
	}
	switch method {
	case http.MethodGet, http.MethodOptions, http.MethodHead:
		return 0
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return 1
	case http.MethodDelete:
		return 2
	default:
		return 0
	}
}

// @ctx: extract tenant ID from URL path patterns like /api/v1/tenants/{tenant_id}/...
func extractTenantFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, p := range parts {
		if p == "tenants" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// @ctx: extract object type from path
func extractObjectType(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, p := range parts {
		if p == "v1" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

// @ctx: extract object ID from path (last non-empty segment)
func extractObjectID(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" && !strings.HasPrefix(parts[i], "api") && !isPathVerb(parts[i]) {
			return parts[i]
		}
	}
	return ""
}

func isPathVerb(s string) bool {
	verbs := map[string]bool{
		"token": true, "introspect": true, "refresh": true,
		"revoke": true, "session": true, "logout": true,
		"status": true, "health": true, "ready": true,
	}
	return verbs[s]
}

// @hlv:log_all_errors
// @hlv:no_sensitive_in_logs
func writeAuthError(c *gin.Context, status int, code string, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}

// @hlv:log_all_errors
func logDenied(c *gin.Context, path string, traceID string, code string, start time.Time) {
	slog.ErrorContext(c.Request.Context(), "auth.denied",
		"path", path,
		"error_code", code,
		"trace_id", traceID,
		"duration_ms", time.Since(start).Milliseconds(),
		"source_ip", c.ClientIP(),
	)
}

// @hlv:log_external_calls
// @hlv:no_sensitive_in_logs
func writeAudit(cfg *Config, event *AuditEvent) {
	if cfg.AuditWriter != nil {
		ctx := context.Background()
		if err := cfg.AuditWriter.WriteAuditEvent(ctx, event); err != nil {
			slog.ErrorContext(ctx, "auth.audit_write_failed",
				"reason", event.Reason,
				"error", err,
			)
		}
	}
}

// @hlv:no_sensitive_in_logs
func redactToken(s string) string {
	if len(s) <= 12 {
		return "***"
	}
	return s[:6] + "..." + s[len(s)-6:]
}

// @hlv:sec [AUTH_BOUNDARY] — default admin roles
func DefaultAdminRoles() []AuthRole {
	return []AuthRole{RoleOwner, RoleSecurityLead}
}

// @ctx: default exempt paths
func DefaultExemptPrefixes() []string {
	return []string{
		"/api/v1/auth/",
		"/api/v1/public/browse/",
	}
}

// @ctx: default exact exempt paths
func DefaultExactExemptPaths() []string {
	return []string{
		"/health",
		"/ready",
		"/metrics",
		"/",
	}
}

// @hlv:sec [AUTH_BOUNDARY] — default keyfunc for RS256 JWT verification
func DefaultKeyFunc(publicKey *rsa.PublicKey) jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		if alg, _ := token.Header["alg"].(string); strings.EqualFold(alg, "none") {
			slog.Warn("auth.keyfunc.alg_none_rejected", "kid", token.Header["kid"])
			return nil, jwt.ErrSignatureInvalid
		}
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return publicKey, nil
	}
}
