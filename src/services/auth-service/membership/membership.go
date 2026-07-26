package membership

// @hlv seed_membership_completeness
// @ctx: contracts AUTH-RBAC-001, ORG-ACCESS-001

// Membership represents a user's role assignment within a scope (group or ontology).
// @hlv:sec config — Membership records are security-sensitive; source of truth for authorization decisions.
type Membership struct {
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
// @hlv:sec config — Policy definitions control attribute-level access rights.
type AttributePolicy struct {
	Pattern map[string]string `json:"pattern"`
	Right   string            `json:"right"`
	Scope   string            `json:"scope"`
}
