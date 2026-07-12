package membership

import "log"

// @hlv max_role_wins
// @ctx: contract ORG-ACCESS-001 — effective role resolution for authorization

// rolePriority maps role names to their priority (higher = more privileged).
// @hlv:sec config — Role priority defines authorization escalation boundaries.
var rolePriority = map[string]int{
	"Viewer":          1,
	"Editor":          2,
	"Maintainer":      3,
	"SupportEngineer": 4,
	"SRE":             5,
	"SecurityLead":    6,
	"ProductOwner":    7,
	"Owner":           8,
}

// ResolveEffectiveRole returns the highest role from a list of memberships using priority-based resolution.
func ResolveEffectiveRole(memberships []Membership) string {
	var maxRole string
	var maxPrio int
	for _, m := range memberships {
		prio, ok := rolePriority[m.Role]
		if !ok {
			log.Printf("tracing: unknown role %q, skipping", m.Role)
			continue
		}
		if prio > maxPrio {
			maxPrio = prio
			maxRole = m.Role
		}
	}
	return maxRole
}
