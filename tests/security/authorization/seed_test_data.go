//go:build integration

// RBAC test data seeding.
//
// The RBAC suite requires a pre-seeded fixture world (projects, memberships,
// visibility, group hierarchy) that the Docker test stack does NOT provision.
// This helper creates the fixtures through the real API once per test binary
// (sync.Once), using dev-signed JWTs minted with the owner claims of each
// tenant.
//
// Fixture map (see .ai-factory/qa/rbac-full-34ac4293/test-cases.md):
//   tenant_A: proj_private_A (private), proj_internal_A (internal),
//             proj_public_A (public), proj_private_A2 (private)
//   tenant_B: proj_private_B (private)
//   members:  u_viewer_A/reporter/editor/maintainer/owner → proj_private_A
//   groups:   group_parent → subgroup_child (max-role-wins / inheritance)

package authorization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// seedOnce guarantees the fixture world is created exactly once per binary.
var seedOnce sync.Once

// seedErr captures the first seeding failure so tests can report it.
var seedErr error

// seedRBACData creates the fixture projects/memberships/groups via the API.
// Must be called before any RBAC test that references the fixtures.
func seedRBACData(t interface{ Helper() }) error {
	seedOnce.Do(func() {
		seedErr = doSeedRBACData()
	})
	return seedErr
}

// doSeedRBACData runs the actual seeding sequence. It starts from a clean
// slate: the fixture world is deleted (if it exists from a previous run) and
// recreated fresh, so the auth-service auto-owner grant on group creation
// always fires and no stale top-level projects pollute inheritance.
func doSeedRBACData() error {
	// Tenant_A owner token for creating tenant_A fixtures.
	ownerA, err := mintJWT(userSpecs[userOwnerA], tokenTTL)
	if err != nil {
		return fmt.Errorf("mint owner_A token: %w", err)
	}
	// Tenant_B owner token for tenant_B fixtures.
	ownerB, err := mintJWT(userSpecs[userOwnerB], tokenTTL)
	if err != nil {
		return fmt.Errorf("mint owner_B token: %w", err)
	}

	// 0. Clean slate: delete any fixture from a previous run (projects first,
	//    then groups). Deletion is best-effort — a fresh DB has nothing.
	for _, p := range []string{projPrivateA, projInternalA, projPublicA, projPrivateA2, projPrivateB} {
		deleteByName(ownerA, "/projects", p)
	}
	for _, g := range []string{groupParent, subgroupChild, "group_internal_A", "group_public_A", "group_tenant_B"} {
		deleteByName(ownerA, "/groups", g)
	}

	// 1. Groups (tenant_A): group_parent → subgroup_child.
	parentGroupID, err := ensureGroup(ownerA, groupParent, "", "private")
	if err != nil {
		return fmt.Errorf("ensure group_parent: %w", err)
	}
	childGroupID, err := ensureGroup(ownerA, subgroupChild, parentGroupID, "private")
	if err != nil {
		return fmt.Errorf("ensure subgroup_child: %w", err)
	}
	_ = childGroupID // referenced by inheritance tests via alias

	// 2. Projects (tenant_A). Placed under groups matching or exceeding each
	//    project's visibility so the creator inherits Owner (projects do not
	//    auto-grant membership) AND the child visibility constraint
	//    (VISIBILITY_VIOLATION) is satisfied.
	projA, err := ensureProject(ownerA, projPrivateA, parentGroupID, "private")
	if err != nil {
		return fmt.Errorf("ensure proj_private_A: %w", err)
	}

	// Internal project needs a parent group with visibility >= internal.
	internalGroupID, err := ensureGroup(ownerA, "group_internal_A", "", "internal")
	if err != nil {
		return fmt.Errorf("ensure group_internal_A: %w", err)
	}
	_, err = ensureProject(ownerA, projInternalA, internalGroupID, "internal")
	if err != nil {
		return fmt.Errorf("ensure proj_internal_A: %w", err)
	}

	// Public project needs a parent group with visibility >= public.
	publicGroupID, err := ensureGroup(ownerA, "group_public_A", "", "public")
	if err != nil {
		return fmt.Errorf("ensure group_public_A: %w", err)
	}
	_, err = ensureProject(ownerA, projPublicA, publicGroupID, "public")
	if err != nil {
		return fmt.Errorf("ensure proj_public_A: %w", err)
	}

	_, err = ensureProject(ownerA, projPrivateA2, parentGroupID, "private")
	if err != nil {
		return fmt.Errorf("ensure proj_private_A2: %w", err)
	}

	// 3. Projects (tenant_B). Same pattern: under a tenant_B group so the
	//    creator inherits Owner and can manage members.
	bGroupID, err := ensureGroup(ownerB, "group_tenant_B", "", "private")
	if err != nil {
		return fmt.Errorf("ensure group_tenant_B: %w", err)
	}
	_, err = ensureProject(ownerB, projPrivateB, bGroupID, "private")
	if err != nil {
		return fmt.Errorf("ensure proj_private_B: %w", err)
	}

	// 4. Memberships on proj_private_A (tenant_A ladder). Roles are
	//    PascalCase per auth-service rolePriority (Owner/Developer/...).
	members := []struct {
		alias string
		role  string
	}{
		{userViewerA, "Viewer"},
		{userReporterA, "Reporter"},
		{userEditorA, "Editor"},
		{userMaintainerA, "Maintainer"},
	}
	for _, m := range members {
		if err := ensureMember(ownerA, projA, m.alias, m.role); err != nil {
			return fmt.Errorf("ensure member %s on %s: %w", m.alias, projA, err)
		}
	}

	return nil
}

// tokenTTL keeps seeded tokens valid for the whole suite run.
const tokenTTL = 2 * 3600 * 1e9 // 2h in ns — see mintJWT signature (time.Duration)

// ensureGroup creates a group if it does not exist. Returns the group id.
// A 409 (already exists) is treated as success.
func ensureGroup(token, name, parentID, visibility string) (string, error) {
	if id := findGroupID(token, name); id != "" {
		return id, nil
	}

	payload := map[string]string{
		"name":       name,
		"visibility": visibility,
	}
	if parentID != "" {
		payload["parent_id"] = parentID
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, apiBase+"/groups", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Idempotency-Key", idemKey("seed-group-"+name+"-"+randomSuffix()))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusCreated {
		var parsed struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if jsonErr := json.Unmarshal(raw, &parsed); jsonErr == nil && parsed.Data.ID != "" {
			return parsed.Data.ID, nil
		}
	}
	if resp.StatusCode == http.StatusConflict {
		if id := findGroupID(token, name); id != "" {
			return id, nil
		}
	}
	return "", fmt.Errorf("create group %s: HTTP %d: %s", name, resp.StatusCode, truncate(string(raw)))
}

// findGroupID resolves a group's real ID by name via the list endpoint.
func findGroupID(token, name string) string {
	req, _ := http.NewRequest(http.MethodGet, apiBase+"/groups?search="+name, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var parsed struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if jsonErr := json.Unmarshal(raw, &parsed); jsonErr != nil {
		return ""
	}
	for _, g := range parsed.Data {
		if g.Name == name {
			return g.ID
		}
	}
	return ""
}

// ensureProject creates a project (auto-creates its paired ontology) if it
// does not already exist. Existence is checked by name first (idempotent
// across runs without idempotency-key collisions). Returns the project id.
func ensureProject(token, name, groupID, visibility string) (string, error) {
	if id := findProjectID(token, name); id != "" {
		return id, nil
	}

	payload := map[string]string{
		"name":       name,
		"visibility": visibility,
	}
	if groupID != "" {
		payload["group_id"] = groupID
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, apiBase+"/projects", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	// Fresh key per attempt — idempotency store persists across runs, so a
	// deterministic key would collide with an older payload.
	req.Header.Set("Idempotency-Key", idemKey("seed-project-"+name+"-"+randomSuffix()))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusCreated {
		var parsed struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if jsonErr := json.Unmarshal(raw, &parsed); jsonErr == nil && parsed.Data.ID != "" {
			return parsed.Data.ID, nil
		}
	}
	if resp.StatusCode == http.StatusConflict {
		// Race: another attempt created it meanwhile — resolve by name.
		if id := findProjectID(token, name); id != "" {
			return id, nil
		}
	}
	return "", fmt.Errorf("create project %s: HTTP %d: %s", name, resp.StatusCode, truncate(string(raw)))
}

// findProjectID resolves a project's real ID by name via the list endpoint.
func findProjectID(token, name string) string {
	req, _ := http.NewRequest(http.MethodGet, apiBase+"/projects?search="+name, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var parsed struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if jsonErr := json.Unmarshal(raw, &parsed); jsonErr != nil {
		return ""
	}
	for _, p := range parsed.Data {
		if p.Name == name {
			return p.ID
		}
	}
	return ""
}

// deleteByName deletes a scope by name via the list + delete endpoints.
// Best-effort: missing objects are fine (fresh DB).
func deleteByName(token, resource, name string) {
	listReq, _ := http.NewRequest(http.MethodGet, apiBase+resource+"?search="+name, nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listResp, err := http.DefaultClient.Do(listReq)
	if err != nil {
		return
	}
	defer listResp.Body.Close()
	raw, _ := io.ReadAll(listResp.Body)
	var parsed struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if jsonErr := json.Unmarshal(raw, &parsed); jsonErr != nil {
		return
	}
	for _, item := range parsed.Data {
		if item.Name != name {
			continue
		}
		delReq, _ := http.NewRequest(http.MethodDelete, apiBase+resource+"/"+item.ID, nil)
		delReq.Header.Set("Authorization", "Bearer "+token)
		delReq.Header.Set("Idempotency-Key", idemKey("seed-del-"+item.ID+"-"+randomSuffix()))
		resp, err := http.DefaultClient.Do(delReq)
		if err == nil {
			resp.Body.Close()
		}
	}
}

// ensureMember adds a member with a role to a project. Idempotent: if the
// membership already exists, the upsert returns 200/409 which we accept.
func ensureMember(token, projectID, userID, role string) error {
	payload := map[string]string{"user_id": userID, "role": role}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, apiBase+"/projects/"+projectID+"/members", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Idempotency-Key", idemKey("seed-member-"+projectID+"-"+userID+"-"+randomSuffix()))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusCreated, http.StatusConflict, http.StatusOK:
		return nil
	default:
		return fmt.Errorf("add member %s to %s: HTTP %d: %s", userID, projectID, resp.StatusCode, truncate(string(raw)))
	}
}

// randomSuffix returns a short random hex suffix for idempotency keys so
// repeated seeding attempts never collide with older payloads.
func randomSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// idemKey sanitizes an idempotency key to the validator's accepted charset
// (alphanumeric + hyphens, max 128 chars). Underscores are replaced by hyphens.
func idemKey(s string) string {
	if len(s) > 128 {
		s = s[:128]
	}
	return strings.ReplaceAll(s, "_", "-")
}

// truncate limits error bodies to a readable size.
func truncate(s string) string {
	if len(s) > 300 {
		return s[:300] + "…"
	}
	return s
}
