package org

// @ctx: contracts ORG-ACCESS-001, SEC-AUTHZ-GATES-001
// @hlv store_in_postgres

import "time"

// ScopeType identifies whether a scope is a group, project, or legacy ontology.
// Under the 1:1 Project ↔ Ontology model, 'project' is the canonical type for
// the workspace container. 'ontology' is retained as a legacy alias during
// the migration window and removed in the follow-up cleanup task.
type ScopeType string

const (
	ScopeGroup    ScopeType = "group"
	ScopeProject  ScopeType = "project"  // canonical (GitLab-aligned)
	ScopeOntology ScopeType = "ontology" // legacy alias — accepted during migration window
)

// Ontology represents the graph content (TBox/ABox, classes, properties,
// individuals, axioms) of a Project. The 1:1 invariant is enforced at the
// schema level: one Ontology row per project scope.
// @hlv:sec [AUTH_BOUNDARY] — Ontology rows inherit access from their Project via 1:1.
type Ontology struct {
	ProjectScope string    `json:"project_scope"` // FK to scopes.id (PK = FK enforces 1:1)
	OntologyID   string    `json:"ontology_id"`   // identifier used by ontology-service
	IRI          string    `json:"iri"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OrgMembership represents a user's role assignment within a scope.
// @hlv:sec [AUTH_BOUNDARY] — Membership records control authorization decisions.
//
//nolint:revive // exported type name stutter is intentional for clarity
type OrgMembership struct {
	UserID    string `json:"user_id"`
	Scope     string `json:"scope"`
	Role      string `json:"role"`
	Inherited bool   `json:"inherited"`
}

// Visibility represents the visibility level of an ontology.
type Visibility string

const (
	VisibilityPrivate  Visibility = "Private"
	VisibilityInternal Visibility = "Internal"
	VisibilityPublic   Visibility = "Public"
)

// AttributePolicy defines an attribute-based access control policy.
// @hlv:sec [AUTH_BOUNDARY] — Policy definitions control attribute-level access rights.
type AttributePolicy struct {
	Pattern map[string]string `json:"pattern"`
	Right   string            `json:"right"`
	Scope   string            `json:"scope"`
}

// ScopeNode represents a group or ontology node in the hierarchy.
type ScopeNode struct {
	ID                string     `json:"id"`
	Type              ScopeType  `json:"type"`
	ParentID          string     `json:"parent_id,omitempty"`
	Visibility        Visibility `json:"visibility"`
	TenantID          string     `json:"tenant_id"`
	UpstreamProjectID string     `json:"upstream_project_id,omitempty"`
	Slug              string     `json:"slug,omitempty"`
	Name              string     `json:"name,omitempty"`
	Description       string     `json:"description,omitempty"`
}

// OrgStore is the storage interface for org data.
// @hlv:sec [AUTH_BOUNDARY] — Store operations must maintain authorization invariants.
//
//nolint:revive // exported type name stutter is intentional for clarity
type OrgStore interface {
	UpsertMembership(m OrgMembership) error
	DeleteMembership(scope, userID string) error
	GetMemberships(scope string) ([]OrgMembership, error)
	GetUserMemberships(userID string) ([]OrgMembership, error)
	GetEffectiveMemberships(userID, scope string) ([]OrgMembership, error)
	UpsertPolicy(p AttributePolicy) error
	DeletePolicy(scope string, policyID string) error
	GetPolicies(scope string) ([]AttributePolicy, error)
	GetAllPolicies() ([]AttributePolicy, error)
	SetVisibility(scope string, v Visibility) error
	GetVisibility(scope string) (Visibility, error)
	GetScope(id string) (*ScopeNode, error)
	UpsertScope(s ScopeNode) error
	DeleteScope(id string) error
	ListChildScopes(parentID string) ([]ScopeNode, error)
	ListAllScopes() ([]ScopeNode, error)
	// Ontology pairing (1:1 with project scopes)
	CreateOntology(ont Ontology) error
	GetOntologyByProjectScope(projectScope string) (*Ontology, error)
	InvalidateCache(scope string)
}

// AuditEvent represents an authorization audit event.
// @hlv:sec [SECRET_HANDLING] — Audit events must not contain secrets or PII beyond identifiers.
type AuditEvent struct {
	Event           string `json:"event"`
	Reason          string `json:"reason"`
	UserID          string `json:"user_id"`
	TenantRequested string `json:"tenant_id_requested,omitempty"`
	ObjectType      string `json:"object_type,omitempty"`
	ObjectID        string `json:"object_id,omitempty"`
	SourceIP        string `json:"source_ip,omitempty"`
	Timestamp       string `json:"timestamp"`
}

// OrgError represents a structured error response.
//
//nolint:revive // exported type name stutter is intentional for clarity
type OrgError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *OrgError) Error() string { return e.Code + ": " + e.Message }

var (
	ErrForbiddenInsufficientRole  = &OrgError{Code: "FORBIDDEN_INSUFFICIENT_ROLE", Message: "User lacks required role for operation"}
	ErrForbiddenAdminOnly         = &OrgError{Code: "FORBIDDEN_ADMIN_ONLY", Message: "Owner role required for membership management"}
	ErrForbiddenCrossTenantAccess = &OrgError{Code: "FORBIDDEN_CROSS_TENANT_ACCESS", Message: "Cross-tenant access denied"}
	ErrScopeNotFound              = &OrgError{Code: "SCOPE_NOT_FOUND", Message: "Group or ontology scope does not exist"}
	ErrPolicyConflict             = &OrgError{Code: "POLICY_CONFLICT", Message: "Conflicting ABAC policy patterns"}
	ErrVisibilityViolation        = &OrgError{Code: "VISIBILITY_VIOLATION", Message: "User cannot access resource at this visibility level"}
	ErrForbiddenObjectNotFound    = &OrgError{Code: "FORBIDDEN_OBJECT_NOT_FOUND_OR_ACCESS_DENIED", Message: "Object not found or access denied"}
	ErrCycleDetected              = &OrgError{Code: "CYCLE_DETECTED", Message: "Circular group membership detected and rejected"}
	ErrHierarchyDepthExceeded     = &OrgError{Code: "HIERARCHY_DEPTH_EXCEEDED", Message: "Group nesting depth exceeds maximum of 5 levels"}
	ErrLastOwnerRemovalBlocked    = &OrgError{Code: "LAST_OWNER_REMOVAL_BLOCKED", Message: "Cannot remove the last Owner from a scope"}
)

// rolePriority maps role names to their priority (higher = more privileged).
// @hlv:sec [AUTH_BOUNDARY] — Role priority defines authorization escalation boundaries.
var rolePriority = map[string]int{
	// MVP roles (user-facing)
	"Guest":      1,
	"Reporter":   2,
	"Developer":  3,
	"Maintainer": 4,
	"Owner":      9,
	// Legacy roles (backward compatibility — accepted as aliases for MVP roles)
	"Viewer": 1,
	"Editor": 3,
	// Enterprise/ops roles (outside MVP happy path)
	"SupportEngineer": 5,
	"SRE":             6,
	"SecurityLead":    7,
	"ProductOwner":    8,
}

// ResolveMaxRole returns the highest role from a list of memberships.
// @hlv max_role_wins
func ResolveMaxRole(memberships []OrgMembership) string {
	var maxRole string
	var maxPrio int
	for _, m := range memberships {
		prio, ok := rolePriority[m.Role]
		if !ok {
			continue
		}
		if prio > maxPrio {
			maxPrio = prio
			maxRole = m.Role
		}
	}
	return maxRole
}

// IsOwner checks if a user has Owner role in the given scope.
func IsOwner(role string) bool {
	return role == "Owner"
}

// IsMaintainerOrAbove checks if a role is at least Maintainer.
func IsMaintainerOrAbove(role string) bool {
	return rolePriority[role] >= rolePriority["Maintainer"]
}

// CacheEntry holds cached auth data with TTL.
type CacheEntry struct {
	Data      any
	ExpiresAt time.Time
}

// CacheStore is a simple TTL-based auth cache.
type CacheStore struct {
	entries map[string]*CacheEntry
	ttl     time.Duration
}

func NewCacheStore(ttl time.Duration) *CacheStore {
	return &CacheStore{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}
}

func (c *CacheStore) Get(key string) any {
	e, ok := c.entries[key]
	if !ok {
		return nil
	}
	if time.Now().After(e.ExpiresAt) {
		delete(c.entries, key)
		return nil
	}
	return e.Data
}

func (c *CacheStore) Set(key string, data any) {
	c.entries[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// @hlv cache_invalidation
func (c *CacheStore) Invalidate(key string) {
	delete(c.entries, key)
}

func (c *CacheStore) InvalidateScope(scope string) {
	for k := range c.entries {
		if containsScope(k, scope) {
			delete(c.entries, k)
		}
	}
}

func containsScope(key, scope string) bool {
	for i := 0; i <= len(key)-len(scope); i++ {
		if key[i:i+len(scope)] == scope {
			return true
		}
	}
	return false
}
