package org

// @ctx: contracts ORG-ACCESS-001 — In-memory implementation of OrgStore
// @hlv store_in_postgres

import (
	"fmt"
	"sync"
)

// @hlv:sec [AUTH_BOUNDARY] — Store holds all authorization data; must maintain invariants.

// MemStore is an in-memory implementation of OrgStore for testing and stub mode.
type MemStore struct {
	mu          sync.RWMutex
	memberships map[string][]OrgMembership // scope -> memberships
	userIndex   map[string][]OrgMembership // userID -> memberships
	policies    map[string][]AttributePolicy
	visibility  map[string]Visibility
	scopes      map[string]*ScopeNode
	ontologies  map[string]*Ontology // projectScope -> Ontology (1:1)
}

func NewMemStore() *MemStore {
	return &MemStore{
		memberships: make(map[string][]OrgMembership),
		userIndex:   make(map[string][]OrgMembership),
		policies:    make(map[string][]AttributePolicy),
		visibility:  make(map[string]Visibility),
		scopes:      make(map[string]*ScopeNode),
		ontologies:  make(map[string]*Ontology),
	}
}

func (m *MemStore) UpsertMembership(mem OrgMembership) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	mems := m.memberships[mem.Scope]
	found := false
	for i, existing := range mems {
		if existing.UserID == mem.UserID {
			mems[i] = mem
			found = true
			break
		}
	}
	if !found {
		m.memberships[mem.Scope] = append(mems, mem)
	}

	// Update user index
	uidx := m.userIndex[mem.UserID]
	found = false
	for i, existing := range uidx {
		if existing.Scope == mem.Scope {
			uidx[i] = mem
			found = true
			break
		}
	}
	if !found {
		m.userIndex[mem.UserID] = append(uidx, mem)
	}
	return nil
}

func (m *MemStore) DeleteMembership(scope, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	mems := m.memberships[scope]
	for i, existing := range mems {
		if existing.UserID == userID {
			m.memberships[scope] = append(mems[:i], mems[i+1:]...)
			break
		}
	}
	uidx := m.userIndex[userID]
	for i, existing := range uidx {
		if existing.Scope == scope {
			m.userIndex[userID] = append(uidx[:i], uidx[i+1:]...)
			break
		}
	}
	return nil
}

func (m *MemStore) GetMemberships(scope string) ([]OrgMembership, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]OrgMembership, len(m.memberships[scope]))
	copy(result, m.memberships[scope])
	return result, nil
}

func (m *MemStore) GetUserMemberships(userID string) ([]OrgMembership, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]OrgMembership, len(m.userIndex[userID]))
	copy(result, m.userIndex[userID])
	return result, nil
}

func (m *MemStore) GetEffectiveMemberships(userID, scope string) ([]OrgMembership, error) {
	return m.GetMemberships(scope)
}

func (m *MemStore) UpsertPolicy(p AttributePolicy) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	pols := m.policies[p.Scope]
	for i, existing := range pols {
		if eqPatterns(existing.Pattern, p.Pattern) {
			pols[i] = p
			m.policies[p.Scope] = pols
			return nil
		}
	}
	m.policies[p.Scope] = append(pols, p)
	return nil
}

func (m *MemStore) GetPolicies(scope string) ([]AttributePolicy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]AttributePolicy, len(m.policies[scope]))
	copy(result, m.policies[scope])
	return result, nil
}

func (m *MemStore) GetAllPolicies() ([]AttributePolicy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var all []AttributePolicy
	for _, pols := range m.policies {
		all = append(all, pols...)
	}
	return all, nil
}

func (m *MemStore) SetVisibility(scope string, v Visibility) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.visibility[scope] = v
	if node, ok := m.scopes[scope]; ok {
		node.Visibility = v
	}
	return nil
}

func (m *MemStore) GetVisibility(scope string) (Visibility, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if v, ok := m.visibility[scope]; ok {
		return v, nil
	}
	return VisibilityPrivate, nil
}

func (m *MemStore) GetScope(id string) (*ScopeNode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.scopes[id]; ok {
		return s, nil
	}
	return nil, nil
}

func (m *MemStore) UpsertScope(s ScopeNode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scopes[s.ID] = &s
	if s.Visibility != "" {
		m.visibility[s.ID] = s.Visibility
	} else {
		m.visibility[s.ID] = VisibilityPrivate
	}
	return nil
}

func (m *MemStore) ListChildScopes(parentID string) ([]ScopeNode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var children []ScopeNode
	for _, s := range m.scopes {
		if s.ParentID == parentID {
			children = append(children, *s)
		}
	}
	return children, nil
}

func (m *MemStore) ListAllScopes() ([]ScopeNode, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var all []ScopeNode
	for _, s := range m.scopes {
		all = append(all, *s)
	}
	return all, nil
}

func (m *MemStore) InvalidateCache(scope string) {}

// DeleteScope removes a scope and associated data.
// Cascades to the paired ontologies row (matches the PostgreSQL FK ON DELETE CASCADE).
func (m *MemStore) DeleteScope(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.scopes, id)
	delete(m.visibility, id)
	delete(m.memberships, id)
	delete(m.policies, id)
	delete(m.ontologies, id) // cascade: remove paired ontology
	// Clean up user index entries for this scope
	for uid, mems := range m.userIndex {
		filtered := mems[:0]
		for _, mem := range mems {
			if mem.Scope != id {
				filtered = append(filtered, mem)
			}
		}
		m.userIndex[uid] = filtered
	}
	return nil
}

// CreateOntology inserts a 1:1 paired ontology row for a project scope.
// The project scope must already exist in the scopes map (FK invariant).
func (m *MemStore) CreateOntology(ont Ontology) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.scopes[ont.ProjectScope]; !ok {
		return &OrgError{Code: "SCOPE_NOT_FOUND", Message: "project scope does not exist; cannot create paired ontology"}
	}
	m.ontologies[ont.ProjectScope] = &ont
	return nil
}

// GetOntologyByProjectScope returns the paired ontology for a project scope, or nil if none.
func (m *MemStore) GetOntologyByProjectScope(projectScope string) (*Ontology, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if ont, ok := m.ontologies[projectScope]; ok {
		return ont, nil
	}
	return nil, nil
}

// DeletePolicy removes a policy from a scope by matching pattern + right.
func (m *MemStore) DeletePolicy(scope string, policyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	pols := m.policies[scope]
	for i, p := range pols {
		// Use pattern + right as policy identity
		if fmt.Sprintf("%v:%s", p.Pattern, p.Right) == policyID {
			m.policies[scope] = append(pols[:i], pols[i+1:]...)
			return nil
		}
	}
	return nil
}

func eqPatterns(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
