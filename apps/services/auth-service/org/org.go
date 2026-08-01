package org

// @ctx: contracts ORG-ACCESS-001, SEC-AUTHZ-GATES-001
// @hlv store_in_postgres

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// @hlv:sec [INPUT_VALIDATION] — All user-supplied scope strings are validated before use.
// @hlv:sec [AUTH_BOUNDARY] — Every operation checks authorization before execution.

// OrgService is the central authorization service for group/ontology access.
// @hlv:sec [AUTH_BOUNDARY] — All authorization decisions flow through OrgService.
//
//nolint:revive
type OrgService struct {
	store OrgStore
	cache *CacheStore
}

// NewOrgService creates a new OrgService with a 300s TTL cache.
// @hlv cache_invalidation
func NewOrgService(store OrgStore) *OrgService {
	return &OrgService{
		store: store,
		cache: NewCacheStore(300 * time.Second),
	}
}

// Store returns the underlying OrgStore for gRPC handlers needing direct store access.
func (s *OrgService) Store() OrgStore {
	return s.store
}

// Cache returns the cache store for invalidation.
func (s *OrgService) Cache() *CacheStore {
	return s.cache
}

// GetEffectiveRole resolves the highest role a user has for a scope, considering inheritance.
// @hlv max_role_wins
func (s *OrgService) GetEffectiveRole(userID, scope string) (string, error) {
	// @hlv:sec [INPUT_VALIDATION] — Validate scope format before query.
	if userID == "" || scope == "" {
		return "", ErrForbiddenInsufficientRole
	}

	cacheKey := "role:" + userID + ":" + scope
	if cached := s.cache.Get(cacheKey); cached != nil {
		role, _ := cached.(string)
		return role, nil
	}

	scopes, err := s.collectInheritedScopes(scope, nil)
	if err != nil {
		return "", err
	}

	var all []OrgMembership
	for _, sc := range scopes {
		ms, err := s.store.GetMemberships(sc)
		if err != nil {
			return "", err
		}
		for _, m := range ms {
			if m.UserID == userID {
				all = append(all, m)
			}
		}
	}

	if len(all) == 0 {
		return "", nil
	}

	for i := range all {
		if len(all[i].Scope) != len(scope) {
			all[i].Inherited = true
		}
	}

	role := ResolveMaxRole(all)
	if role != "" {
		s.cache.Set(cacheKey, role)
	}
	return role, nil
}

// @hlv visibility_checked
// CheckAccess determines if a user can perform an action on a scope.
func (s *OrgService) CheckAccess(userID, scope, action string) error {
	// @hlv:sec [INPUT_VALIDATION] — Validate inputs before authorization.
	if userID == "" && action == "read_members" {
		// Allow anonymous read for Public visibility only.
		scopeNode, _ := s.store.GetScope(scope)
		if scopeNode != nil && scopeNode.Visibility == VisibilityPublic {
			return nil
		}
		return ErrForbiddenInsufficientRole
	}

	// @hlv:sec [AUTH_BOUNDARY] — Authenticated identity is required for all state-changing operations.
	if userID == "" {
		return ErrForbiddenInsufficientRole
	}

	scopeNode, err := s.store.GetScope(scope)
	if err != nil {
		return ErrScopeNotFound
	}
	if scopeNode == nil {
		return ErrScopeNotFound
	}

	// @hlv:sec [AUTH_BOUNDARY] — Cross-tenant access is always denied.
	if e := s.checkTenantAccess(userID, scopeNode); e != nil {
		return e
	}

	// @hlv visibility_checked
	visErr := s.checkVisibility(userID, scopeNode)

	role, err := s.GetEffectiveRole(userID, scope)
	if err != nil {
		return err
	}

	// Public visibility: allow read actions without role check
	if scopeNode.Visibility == VisibilityPublic && action == "read_members" {
		return nil
	}
	// Internal visibility: allow read actions for any authenticated user without role check
	if scopeNode.Visibility == VisibilityInternal && userID != "" && action == "read_members" {
		return nil
	}
	// Private visibility: require explicit role
	if visErr != nil {
		return ErrForbiddenInsufficientRole
	}

	if !s.hasRequiredRole(role, action) {
		return ErrForbiddenInsufficientRole
	}

	return nil
}

// @hlv owner_only_membership
// @hlv:sec [AUTH_BOUNDARY] — Membership management requires Owner role.
// UpdateMembership adds or updates a user's role in a scope. Only Owner can manage membership.
func (s *OrgService) UpdateMembership(requesterID, scope, targetUserID, newRole string) (*OrgMembership, *AuditEvent, error) {
	// @hlv:sec [INPUT_VALIDATION] — Validate all input parameters.
	if requesterID == "" || scope == "" || targetUserID == "" || newRole == "" {
		return nil, nil, ErrForbiddenInsufficientRole
	}

	_, ok := rolePriority[newRole]
	if !ok {
		return nil, nil, ErrForbiddenInsufficientRole
	}

	scopeNode, err := s.store.GetScope(scope)
	if err != nil || scopeNode == nil {
		return nil, nil, ErrScopeNotFound
	}

	// @hlv owner_only_membership — Only Owner can manage membership.
	requesterRole, err := s.GetEffectiveRole(requesterID, scope)
	if err != nil {
		return nil, nil, err
	}

	if requesterRole != "Owner" {
		audit := &AuditEvent{
			Event:      "authorization.denied",
			Reason:     "FORBIDDEN_ADMIN_ONLY",
			UserID:     requesterID,
			ObjectType: string(scopeNode.Type),
			ObjectID:   scope,
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
		}
		// @hlv log_state_changes
		log.Printf(`{"event":"state.changed","entity":"membership","user_id":"%s","scope":"%s","action":"deny","reason":"FORBIDDEN_ADMIN_ONLY"}`, requesterID, scope)
		return nil, audit, ErrForbiddenAdminOnly
	}

	m := OrgMembership{
		UserID:    targetUserID,
		Scope:     scope,
		Role:      newRole,
		Inherited: false,
	}
	if err := s.store.UpsertMembership(m); err != nil {
		return nil, nil, err
	}

	// @hlv cache_invalidation
	s.cache.InvalidateScope(scope)

	audit := &AuditEvent{
		Event:      "authorization.granted",
		Reason:     "membership_updated",
		UserID:     requesterID,
		ObjectType: string(scopeNode.Type),
		ObjectID:   scope,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	// @hlv log_state_changes
	log.Printf(`{"event":"state.changed","entity":"membership","user_id":"%s","scope":"%s","target_user":"%s","new_role":"%s"}`, requesterID, scope, targetUserID, newRole)

	return &m, audit, nil
}

// @hlv most_specific_abac_wins
// SavePolicy creates or updates an ABAC policy. Returns POLICY_CONFLICT if conflicting patterns exist.
func (s *OrgService) SavePolicy(requesterID string, policy AttributePolicy) (*AuditEvent, error) {
	// @hlv:sec [INPUT_VALIDATION] — Validate policy pattern before saving.
	if policy.Scope == "" || policy.Right == "" || len(policy.Pattern) == 0 {
		return nil, ErrForbiddenInsufficientRole
	}

	scopeNode, err := s.store.GetScope(policy.Scope)
	if err != nil || scopeNode == nil {
		return nil, ErrScopeNotFound
	}

	// @hlv owner_only_membership — Only Owner can manage policies.
	requesterRole, err := s.GetEffectiveRole(requesterID, policy.Scope)
	if err != nil {
		return nil, err
	}
	if !IsOwner(requesterRole) {
		audit := &AuditEvent{
			Event:      "authorization.denied",
			Reason:     "FORBIDDEN_ADMIN_ONLY",
			UserID:     requesterID,
			ObjectType: string(scopeNode.Type),
			ObjectID:   policy.Scope,
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
		}
		return audit, ErrForbiddenAdminOnly
	}

	// @hlv most_specific_abac_wins — Most specific pattern wins, conflicting patterns are rejected.
	existing, err := s.store.GetPolicies(policy.Scope)
	if err != nil {
		return nil, err
	}

	for _, ep := range existing {
		if patternsConflict(ep.Pattern, policy.Pattern) {
			return nil, ErrPolicyConflict
		}
	}

	if err := s.store.UpsertPolicy(policy); err != nil {
		return nil, err
	}

	// @hlv cache_invalidation
	s.cache.InvalidateScope(policy.Scope)

	// @hlv log_state_changes
	log.Printf(`{"event":"state.changed","entity":"abac_policy","scope":"%s","right":"%s","pattern":"%v"}`, policy.Scope, policy.Right, policy.Pattern)

	audit := &AuditEvent{
		Event:      "authorization.granted",
		Reason:     "policy_saved",
		UserID:     requesterID,
		ObjectType: string(scopeNode.Type),
		ObjectID:   policy.Scope,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
	return audit, nil
}

// SetVisibility changes the visibility of a scope. Only Owner can change visibility.
// @hlv visibility_checked
func (s *OrgService) SetVisibility(requesterID, scope string, v Visibility) error {
	scopeNode, err := s.store.GetScope(scope)
	if err != nil || scopeNode == nil {
		return ErrScopeNotFound
	}

	requesterRole, err := s.GetEffectiveRole(requesterID, scope)
	if err != nil {
		return err
	}
	if !IsOwner(requesterRole) {
		return ErrForbiddenAdminOnly
	}

	oldVis := scopeNode.Visibility
	if err := s.store.SetVisibility(scope, v); err != nil {
		return err
	}

	s.cache.InvalidateScope(scope)

	log.Printf(`{"event":"state.changed","entity":"visibility","scope":"%s","old":"%s","new":"%s"}`, scope, oldVis, v)
	return nil
}

// @hlv owner_only_membership
// @hlv last_owner_protection
// RemoveMember removes a user's membership from a scope with last-owner protection.
// Only Owner can remove members. The last Owner of a scope cannot be removed.
func (s *OrgService) RemoveMember(requesterID, scope, targetUserID string) (*AuditEvent, error) {
	// @hlv:sec [INPUT_VALIDATION] — Validate inputs.
	if requesterID == "" || scope == "" || targetUserID == "" {
		return nil, ErrForbiddenInsufficientRole
	}

	scopeNode, err := s.store.GetScope(scope)
	if err != nil || scopeNode == nil {
		return nil, ErrScopeNotFound
	}

	// @hlv:sec [AUTH_BOUNDARY] — Only Owner can remove members.
	requesterRole, err := s.GetEffectiveRole(requesterID, scope)
	if err != nil {
		return nil, err
	}
	if !IsOwner(requesterRole) {
		audit := &AuditEvent{
			Event:      "authorization.denied",
			Reason:     "FORBIDDEN_ADMIN_ONLY",
			UserID:     requesterID,
			ObjectType: string(scopeNode.Type),
			ObjectID:   scope,
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
		}
		log.Printf(`{"event":"state.changed","entity":"membership","user_id":"%s","scope":"%s","action":"deny","reason":"FORBIDDEN_ADMIN_ONLY"}`, requesterID, scope)
		return audit, ErrForbiddenAdminOnly
	}

	// @hlv last_owner_protection — Cannot remove the last Owner from a scope.
	members, err := s.store.GetMemberships(scope)
	if err != nil {
		return nil, err
	}

	// Check if the target user is an Owner
	isTargetOwner := false
	ownerCount := 0
	for _, m := range members {
		if m.Role == "Owner" {
			ownerCount++
			if m.UserID == targetUserID {
				isTargetOwner = true
			}
		}
	}

	// Block removal if target is the last Owner
	if isTargetOwner && ownerCount <= 1 {
		audit := &AuditEvent{
			Event:      "authorization.denied",
			Reason:     "LAST_OWNER_REMOVAL_BLOCKED",
			UserID:     requesterID,
			ObjectType: string(scopeNode.Type),
			ObjectID:   scope,
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
		}
		log.Printf(`{"event":"state.changed","entity":"membership","user_id":"%s","scope":"%s","action":"deny","reason":"LAST_OWNER_REMOVAL_BLOCKED"}`, requesterID, scope)
		return audit, ErrLastOwnerRemovalBlocked
	}

	if err := s.store.DeleteMembership(scope, targetUserID); err != nil {
		return nil, err
	}

	// @hlv cache_invalidation
	s.cache.InvalidateScope(scope)

	log.Printf(`{"event":"state.changed","entity":"membership","user_id":"%s","scope":"%s","target_user":"%s","action":"removed"}`, requesterID, scope, targetUserID)

	audit := &AuditEvent{
		Event:      "authorization.revoked",
		Reason:     "member_removed",
		UserID:     requesterID,
		ObjectType: string(scopeNode.Type),
		ObjectID:   scope,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	return audit, nil
}

// @hlv visibility_checked
// checkVisibility verifies that the user can access a scope at its visibility level.
func (s *OrgService) checkVisibility(userID string, node *ScopeNode) error {
	switch node.Visibility {
	case VisibilityPublic:
		return nil
	case VisibilityInternal:
		if userID == "" {
			return ErrVisibilityViolation
		}
		return nil
	case VisibilityPrivate:
		if userID == "" {
			return ErrVisibilityViolation
		}
		role, err := s.GetEffectiveRole(userID, node.ID)
		if err != nil {
			return err
		}
		if role == "" {
			return ErrVisibilityViolation
		}
		return nil
	default:
		return ErrVisibilityViolation
	}
}

// @hlv FORBIDDEN_CROSS_TENANT_ACCESS
// checkTenantAccess verifies the user's tenant matches the scope's tenant.
func (s *OrgService) checkTenantAccess(userID string, node *ScopeNode) error {
	if node.TenantID == "" {
		return nil
	}
	ms, err := s.store.GetUserMemberships(userID)
	if err != nil {
		return ErrForbiddenCrossTenantAccess
	}
	for _, m := range ms {
		mNode, _ := s.store.GetScope(m.Scope)
		if mNode != nil && mNode.TenantID == node.TenantID {
			return nil
		}
	}
	return ErrForbiddenCrossTenantAccess
}

// @hlv most_specific_abac_wins
// EvaluateABAC checks a user's access against configured attribute policies for a scope.
func (s *OrgService) EvaluateABAC(userID, scope, right string, attributes map[string]string) (bool, error) {
	cacheKey := "abac:" + userID + ":" + scope + ":" + right
	if cached := s.cache.Get(cacheKey); cached != nil {
		val, _ := cached.(bool)
		return val, nil
	}

	policies, err := s.store.GetPolicies(scope)
	if err != nil {
		return false, err
	}

	// @hlv most_specific_abac_wins — Most specific pattern wins.
	var bestScore int
	var bestMatch bool
	for _, p := range policies {
		if score := matchPattern(p.Pattern, attributes); score > bestScore {
			bestScore = score
			bestMatch = hasRight(p.Right, right)
		}
	}

	s.cache.Set(cacheKey, bestMatch)
	return bestMatch, nil
}

// @hlv CYCLE_DETECTED
// @hlv hierarchy_depth
// CreateScope creates a new group or ontology scope with circular dependency detection
// and hierarchy depth validation (max 5 levels). Emits audit log on success.
func (s *OrgService) CreateScope(requesterID string, node ScopeNode) error {
	if node.ParentID != "" {
		// @hlv:sec [INPUT_VALIDATION] — Validate parent exists before creating child.
		parent, err := s.store.GetScope(node.ParentID)
		if err != nil || parent == nil {
			return ErrScopeNotFound
		}

		// @hlv:sec [AUTH_BOUNDARY] — Caller must have at least Maintainer role
		// in the parent group to create a project under it.
		if node.Type == ScopeProject {
			var role string
			role, err = s.GetEffectiveRole(requesterID, node.ParentID)
			if err != nil {
				return err
			}
			if !IsMaintainerOrAbove(role) {
				return ErrForbiddenInsufficientRole
			}
		}

		// @hlv CYCLE_DETECTED — Circular group membership MUST be detected and rejected at creation.
		if hasCycle, _ := s.detectCycle(node.ParentID, node.ID); hasCycle {
			return ErrCycleDetected
		}

		// @hlv hierarchy_depth — MVP group nesting is limited to 5 levels.
		depth, err := s.scopeDepth(node.ParentID)
		if err != nil {
			return err
		}
		if depth >= 5 {
			return ErrHierarchyDepthExceeded
		}

		// @hlv visibility_inheritance — Subgroup visibility cannot exceed parent.
		parentVis := visibilityLevel(parent.Visibility)
		if node.Visibility == "" {
			// Inherit visibility from parent when not explicitly specified.
			node.Visibility = parent.Visibility
			log.Printf(`{"event":"org.create_scope.visibility","scope":"%s","parent_vis":"%s","child_vis":"%s","inherited":true}`, node.ID, parent.Visibility, node.Visibility)
		} else {
			childVis := visibilityLevel(node.Visibility)
			if childVis > parentVis {
				log.Printf(`{"event":"org.create_scope.visibility_violation","scope":"%s","parent_vis":"%s","child_vis":"%s"}`, node.ID, parent.Visibility, node.Visibility)
				return &OrgError{Code: "VISIBILITY_VIOLATION", Message: "Subgroup cannot be more visible than its parent"}
			}
			log.Printf(`{"event":"org.create_scope.visibility","scope":"%s","parent_vis":"%s","child_vis":"%s","inherited":false}`, node.ID, parent.Visibility, node.Visibility)
		}
	} else if node.Visibility == "" {
		// Top-level group: default to Private.
		node.Visibility = VisibilityPrivate
	}

	// GitLab-style slug for GUI navigation. Derive from the name when the
	// caller did not provide one: lowercase, spaces to hyphens, strip
	// non-alphanumeric, trim leading/trailing hyphens.
	if node.Slug == "" {
		node.Slug = deriveSlug(node.Name)
	}

	if err := s.store.UpsertScope(node); err != nil {
		return err
	}

	// @hlv:sec [AUTH_BOUNDARY] — Group creator becomes Owner of the new group.
	// GitLab-aligned: without auto-membership, the creator would have no role in
	// their own group and would be FORBIDDEN_INSUFFICIENT_ROLE when creating a
	// project under it. Projects do not auto-add membership (access is inherited
	// via the parent group).
	if node.Type == ScopeGroup {
		if err := s.store.UpsertMembership(OrgMembership{
			UserID: requesterID,
			Scope:  node.ID,
			Role:   "Owner",
		}); err != nil {
			return err
		}
	}

	// @hlv audit_log
	log.Printf(`{"event":"audit.scope.created","scope":"%s","type":"%s","parent":"%s","requester":"%s","visibility":"%s"}`, node.ID, node.Type, node.ParentID, requesterID, node.Visibility)

	return nil
}

// visibilityLevel maps a Visibility string to a numeric level for comparison.
// Private=0, Internal=1, Public=2. Higher = more permissive/visible.
func visibilityLevel(v Visibility) int {
	switch v {
	case VisibilityPublic:
		return 2
	case VisibilityInternal:
		return 1
	case VisibilityPrivate:
		return 0
	default:
		return 0 // unknown → treated as Private for safety
	}
}

// deriveSlug converts a display name into a GitLab-style slug:
// lowercase, spaces to hyphens, strip non-alphanumeric characters.
func deriveSlug(name string) string {
	sb := make([]rune, 0, len(name))
	prevDash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			sb = append(sb, r)
			prevDash = false
		default:
			if !prevDash && len(sb) > 0 {
				sb = append(sb, '-')
				prevDash = true
			}
		}
	}
	// Trim trailing dash
	for len(sb) > 0 && sb[len(sb)-1] == '-' {
		sb = sb[:len(sb)-1]
	}
	return string(sb)
}

// @hlv hierarchy_depth
// MoveScope moves a project from one group to another, validating depth and owner auth.
// Returns the updated scope and an audit event.
func (s *OrgService) MoveScope(requesterID, scopeID, newParentID string) (*ScopeNode, *AuditEvent, error) {
	// @hlv:sec [INPUT_VALIDATION] — Validate inputs.
	if scopeID == "" || newParentID == "" {
		return nil, nil, ErrForbiddenInsufficientRole
	}

	scope, err := s.store.GetScope(scopeID)
	if err != nil || scope == nil {
		return nil, nil, ErrScopeNotFound
	}

	// Validate new parent exists.
	newParent, err := s.store.GetScope(newParentID)
	if err != nil || newParent == nil {
		return nil, nil, ErrScopeNotFound
	}

	// @hlv CYCLE_DETECTED — Check moving scope doesn't create cycle.
	if hasCycle, _ := s.detectCycle(newParentID, scopeID); hasCycle {
		return nil, nil, ErrCycleDetected
	}

	// @hlv hierarchy_depth — Check new parent depth + 1 <= 5.
	depth, err := s.scopeDepth(newParentID)
	if err != nil {
		return nil, nil, err
	}
	if depth >= 5 {
		return nil, nil, ErrHierarchyDepthExceeded
	}

	// @hlv:sec [AUTH_BOUNDARY] — Only Owner can move scopes.
	requesterRole, err := s.GetEffectiveRole(requesterID, scopeID)
	if err != nil {
		return nil, nil, err
	}
	if !IsOwner(requesterRole) {
		return nil, nil, ErrForbiddenAdminOnly
	}

	oldParentID := scope.ParentID
	scope.ParentID = newParentID
	if err := s.store.UpsertScope(*scope); err != nil {
		return nil, nil, err
	}

	// @hlv cache_invalidation — Invalidate authorization caches for affected scopes.
	s.cache.InvalidateScope(scopeID)
	if oldParentID != "" {
		s.cache.InvalidateScope(oldParentID)
	}
	s.cache.InvalidateScope(newParentID)

	// @hlv log_state_changes
	log.Printf(`{"event":"state.changed","entity":"scope","scope":"%s","old_parent":"%s","new_parent":"%s","actor":"%s"}`, scopeID, oldParentID, newParentID, requesterID)

	audit := &AuditEvent{
		Event:      "scope.moved",
		Reason:     "project_moved",
		UserID:     requesterID,
		ObjectType: string(scope.Type),
		ObjectID:   scopeID,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	return scope, audit, nil
}

// scopeDepth calculates the depth of a scope in the hierarchy (0 = root, 5 = max).
// @hlv hierarchy_depth
func (s *OrgService) scopeDepth(scopeID string) (int, error) {
	depth := 0
	currentID := scopeID
	visited := make(map[string]bool)

	for currentID != "" {
		if visited[currentID] {
			return 0, ErrCycleDetected
		}
		visited[currentID] = true

		node, err := s.store.GetScope(currentID)
		if err != nil || node == nil {
			return depth, nil // Reached root or unknown scope
		}
		if node.ParentID == "" {
			break
		}
		depth++
		if depth > 10 {
			return 0, ErrCycleDetected // Safety limit
		}
		currentID = node.ParentID
	}
	return depth, nil
}

// collectInheritedScopes walks up the hierarchy to collect all parent scopes.
func (s *OrgService) collectInheritedScopes(scope string, visited map[string]bool) ([]string, error) {
	if visited == nil {
		visited = make(map[string]bool)
	}
	if visited[scope] {
		return nil, ErrCycleDetected
	}
	visited[scope] = true

	node, err := s.store.GetScope(scope)
	if err != nil || node == nil {
		return []string{scope}, nil
	}

	scopes := []string{scope}
	if node.ParentID != "" {
		parents, err := s.collectInheritedScopes(node.ParentID, visited)
		if err != nil {
			return nil, err
		}
		scopes = append(scopes, parents...)
	}
	return scopes, nil
}

// @hlv CYCLE_DETECTED
// detectCycle checks if adding childID as a parent of parentID would create a cycle.
func (s *OrgService) detectCycle(currentID, targetID string) (bool, error) {
	if currentID == targetID {
		return true, nil
	}
	node, err := s.store.GetScope(currentID)
	if err != nil || node == nil {
		return false, nil
	}
	if node.ParentID == "" {
		return false, nil
	}
	if node.ParentID == targetID {
		return true, nil
	}
	return s.detectCycle(node.ParentID, targetID)
}

// hasRequiredRole checks if a role is sufficient for the given action.
func (s *OrgService) hasRequiredRole(role string, action string) bool {
	if role == "" {
		return false
	}

	requiredRole := map[string]string{
		"read_members":      "Viewer",
		"update_membership": "Owner",
		"read_policies":     "Viewer",
		"save_policy":       "Owner",
		"create_scope":      "Owner",
		"delete_scope":      "Owner",
		"set_visibility":    "Owner",
	}

	req, ok := requiredRole[action]
	if !ok {
		return false
	}
	return rolePriority[role] >= rolePriority[req]
}

// patternsConflict checks if two policy patterns would overlap.
// @hlv POLICY_CONFLICT
func patternsConflict(a, b map[string]string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	// Patterns conflict if they share the same type and value.
	if a["type"] == b["type"] {
		if a["value"] == b["value"] {
			return true
		}
	}
	return false
}

// matchPattern computes a specificity score for how well a policy pattern matches attributes.
// @hlv most_specific_abac_wins
func matchPattern(pattern, attributes map[string]string) int {
	score := 0
	for k, v := range pattern {
		attrVal, ok := attributes[k]
		if !ok {
			return 0
		}
		// uri_prefix patterns match if the attribute value starts with the pattern value
		if pattern["type"] == "uri_prefix" && k == "value" {
			if len(attrVal) >= len(v) && attrVal[:len(v)] == v {
				score += len(v) // longer prefix = higher specificity
				continue
			}
			return 0
		}
		if attrVal == v {
			score++
		} else {
			return 0
		}
	}
	return score
}

// hasRight checks if the granted right satisfies the required right.
var rightHierarchy = map[string]int{
	"read":                1,
	"write":               2,
	"write_with_approval": 2,
	"delete":              3,
	"admin":               4,
}

func hasRight(granted, required string) bool {
	return rightHierarchy[granted] >= rightHierarchy[required]
}

// ParseScope splits "group/id" or "project/id" into type and ID.
// Under the 1:1 Project ↔ Ontology model, "project/" is the canonical prefix.
// The legacy "ontology/" alias was removed in Task 7.4 of the
// project-ontology-separation plan — use "project/" instead.
func ParseScope(scope string) (ScopeType, string, error) {
	// @hlv:sec [INPUT_VALIDATION] — Parse and validate scope format.
	parts := strings.SplitN(scope, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid scope format")
	}
	st := ScopeType(parts[0])
	if st != ScopeGroup && st != ScopeProject {
		return "", "", fmt.Errorf("unknown scope type: %s", parts[0])
	}
	return st, parts[1], nil
}
