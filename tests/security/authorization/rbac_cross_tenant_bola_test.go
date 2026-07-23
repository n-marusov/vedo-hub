// RBAC Full Test Suite — Cross-Tenant and Cross-Object BOLA
//
// Validates: TC-001 through TC-006, TC-032 through TC-034
// Categories: A (Cross-Tenant BOLA), B (Cross-Object BOLA),
//             L (GraphQL RBAC), M (SPARQL RBAC)

package authorization

import (
	"encoding/json"
	"fmt"
	"testing"
)

// ============================================================================
// A. Cross-Tenant BOLA (Type A)
// ============================================================================

// TestRBAC_TC001_CrossTenantReadOntology_Returns403 validates that a user from
// tenant_A cannot read a private ontology in tenant_B.
//
// TC-001: Cross-tenant read of private ontology via REST.
func TestRBAC_TC001_CrossTenantReadOntology_Returns403(t *testing.T) {
	// Given: u_viewer_A is authenticated (tenant_A) and ont_B exists in tenant_B.
	jwt, ok := setupTest(t, userViewerA)
	if !ok {
		return
	}

	// When: GET /api/v1/ontologies/ont_B with u_viewer_A's JWT.
	resp := doRequest(t, "GET", apiBase+"/ontologies/"+ontB, nil, jwt, "")
	if resp == nil {
		return
	}
	defer resp.Body.Close()
	body := readBody(t, resp)

	// Then: HTTP 403 with FORBIDDEN_CROSS_TENANT_ACCESS, never 404 or 500.
	assertStatus(t, resp, 403, "TC-001: cross-tenant read ontology")
	assertNotStatus(t, resp, 404, "TC-001: must not return 404")
	assertNotStatus(t, resp, 500, "TC-001: must not return 500")
	assertBodyContains(t, body, errCrossTenantAccess, "TC-001: error code")
}

// TestRBAC_TC002_CrossTenantCreateOntology_Returns403 validates that a user
// from tenant_A cannot create an ontology scoped to tenant_B.
//
// TC-002: Cross-tenant create ontology in foreign tenant.
func TestRBAC_TC002_CrossTenantCreateOntology_Returns403(t *testing.T) {
	// Given: u_owner_A is authenticated (tenant_A) and tenant_B exists.
	jwt, ok := setupTest(t, userOwnerA)
	if !ok {
		return
	}

	// When: POST /api/v1/ontologies with tenantId=tenant_B.
	payload := fmt.Sprintf(`{"tenantId":"tenant_B","name":"stolen-ont"}`)
	resp := doRequest(t, "POST", apiBase+"/ontologies", []byte(payload), jwt, "11111111-1111-1111-1111-111111111111")
	if resp == nil {
		return
	}
	defer resp.Body.Close()
	body := readBody(t, resp)

	// Then: HTTP 403, no ontology created in tenant_B.
	assertStatus(t, resp, 403, "TC-002: cross-tenant create ontology")
	assertBodyContains(t, body, errCrossTenantAccess, "TC-002: error code")
}

// TestRBAC_TC003_CrossTenantBranchAccess_Returns403 validates that a user
// from tenant_A cannot access a branch of an ontology in tenant_B.
//
// TC-003: Cross-tenant access to branch of foreign ontology.
func TestRBAC_TC003_CrossTenantBranchAccess_Returns403(t *testing.T) {
	// Given: u_editor_A is authenticated (tenant_A), branch_feature_B exists on ont_B.
	jwt, ok := setupTest(t, userEditorA)
	if !ok {
		return
	}

	// When: GET /api/v1/branches/branch_feature_B with JWT of cross-tenant user.
	resp := doRequest(t, "GET", apiBase+"/branches/"+branchFeatureB, nil, jwt, "")
	if resp == nil {
		return
	}
	defer resp.Body.Close()
	body := readBody(t, resp)

	// Then: HTTP 403, never 404 or 500.
	assertStatus(t, resp, 403, "TC-003: cross-tenant branch access")
	assertNotStatus(t, resp, 404, "TC-003: must not return 404")
	assertNotStatus(t, resp, 500, "TC-003: must not return 500")
	assertBodyContains(t, body, errCrossTenantAccess, "TC-003: error code")
}

// TestRBAC_TC004_CrossTenantWebSocketJoin_ReturnsError validates that a user
// from tenant_A cannot join a WebSocket room for an ontology in tenant_B.
//
// TC-004: Cross-tenant WebSocket JoinRoom for foreign ontology.
func TestRBAC_TC004_CrossTenantWebSocketJoin_ReturnsError(t *testing.T) {
	// WebSocket integration requires a WS client and a running collaboration service.
	// This test is skipped until WebSocket test infrastructure is wired.
	//
	// When implementing:
	//   1. Open WS connection to ws://localhost:8080/ws/collaboration
	//   2. Send JoinRoom: {"type":"join","ontologyId":"ont_B"}
	//   3. Expect error frame: {"type":"error","code":"FORBIDDEN_CROSS_TENANT_ACCESS"}
	//   4. Assert connection NOT joined to ont_B room
	//   5. Assert audit_events contains denied-join record
	t.Skip("TC-004: WebSocket test requires WS client infrastructure — skipped until wired")
}

// ============================================================================
// B. Cross-Object BOLA (Type B — same tenant, no role)
// ============================================================================

// TestRBAC_TC005_OutsiderReadsRestrictedProject_Returns403 validates that an
// authenticated user without a role on a project cannot read it.
//
// TC-005: Authenticated user without role reads restricted project.
func TestRBAC_TC005_OutsiderReadsRestrictedProject_Returns403(t *testing.T) {
	// Given: u_outsider_A is authenticated but NOT a member of proj_private_A.
	jwt, ok := setupTest(t, userOutsiderA)
	if !ok {
		return
	}

	// When: GET /api/v1/projects/proj_private_A with outsiderts JWT.
	resp := doRequest(t, "GET", apiBase+"/projects/"+projPrivateA, nil, jwt, "")
	if resp == nil {
		return
	}
	defer resp.Body.Close()
	body := readBody(t, resp)

	// Then: HTTP 403 with FORBIDDEN_INSUFFICIENT_ROLE.
	assertStatus(t, resp, 403, "TC-005: outsider reads restricted project")
	assertBodyContains(t, body, errInsufficientRole, "TC-005: error code")
}

// TestRBAC_TC006_MemberCannotReadOtherProject_Returns403 validates that a
// member of one project cannot read another project in the same tenant.
//
// TC-006: Member of one project cannot read another project in same tenant.
func TestRBAC_TC006_MemberCannotReadOtherProject_Returns403(t *testing.T) {
	// Given: u_viewer_A is a member of proj_private_A but NOT of proj_private_A2.
	jwt, ok := setupTest(t, userViewerA)
	if !ok {
		return
	}

	// When: GET /api/v1/projects/proj_private_A2 with u_viewer_A's JWT.
	resp := doRequest(t, "GET", apiBase+"/projects/"+projPrivateA2, nil, jwt, "")
	if resp == nil {
		return
	}
	defer resp.Body.Close()
	body := readBody(t, resp)

	// Then: HTTP 403 with FORBIDDEN_INSUFFICIENT_ROLE.
	assertStatus(t, resp, 403, "TC-006: member reads other project in same tenant")
	assertBodyContains(t, body, errInsufficientRole, "TC-006: error code")
}

// ============================================================================
// L. GraphQL RBAC
// ============================================================================

// TestRBAC_TC032_GraphQLMutation_CrossTenantBOLA_ReturnsError validates that
// a GraphQL mutation from tenant_A cannot modify an ontology in tenant_B.
//
// TC-032: GraphQL mutation — cross-tenant BOLA on ontology mutation.
func TestRBAC_TC032_GraphQLMutation_CrossTenantBOLA_ReturnsError(t *testing.T) {
	// Given: u_editor_A is authenticated (tenant_A), ont_B exists in tenant_B.
	jwt, ok := setupTest(t, userEditorA)
	if !ok {
		return
	}

	// When: GraphQL mutation updating ont_B's name.
	mutation := map[string]string{
		"query": fmt.Sprintf(`mutation { updateOntology(id: "%s", name: "hacked-name") { id } }`, ontB),
	}
	payload, _ := json.Marshal(mutation)

	resp := doRequest(t, "POST", apiBase+"/graphql", payload, jwt, "")
	if resp == nil {
		return
	}
	defer resp.Body.Close()
	body := readBody(t, resp)

	// Then: GraphQL returns errors with FORBIDDEN_CROSS_TENANT_ACCESS.
	if resp.StatusCode == 200 {
		assertBodyContains(t, body, errCrossTenantAccess, "TC-032: GraphQL error code")
		assertBodyContains(t, body, `"errors"`, "TC-032: GraphQL errors array")
	} else {
		assertStatus(t, resp, 403, "TC-032: GraphQL mutation cross-tenant")
	}
}

// TestRBAC_TC033_GraphQLQuery_RestrictedOntology_ReturnsError validates that
// a Viewer cannot query a restricted ontology via GraphQL.
//
// TC-033: GraphQL query — Viewer cannot query restricted ontology.
func TestRBAC_TC033_GraphQLQuery_RestrictedOntology_ReturnsError(t *testing.T) {
	// Given: u_viewer_A is authenticated, ont_B exists in tenant_B (cross-tenant).
	jwt, ok := setupTest(t, userViewerA)
	if !ok {
		return
	}

	// When: GraphQL query requesting ont_B's id, name, and classes.
	query := map[string]string{
		"query": fmt.Sprintf(`query { ontology(id: "%s") { id name classes { name } } }`, ontB),
	}
	payload, _ := json.Marshal(query)

	resp := doRequest(t, "POST", apiBase+"/graphql", payload, jwt, "")
	if resp == nil {
		return
	}
	defer resp.Body.Close()
	body := readBody(t, resp)

	// Then: GraphQL returns errors, no ontology data leaked.
	if resp.StatusCode == 200 {
		assertBodyContains(t, body, errCrossTenantAccess, "TC-033: GraphQL error code")
		assertBodyContains(t, body, `"errors"`, "TC-033: GraphQL errors array")
		assertBodyContains(t, body, `"data":null`, "TC-033: data should be null on error")
	} else {
		assertStatus(t, resp, 403, "TC-033: GraphQL query cross-tenant")
	}
}

// ============================================================================
// M. SPARQL RBAC
// ============================================================================

// TestRBAC_TC034_SPARQL_CrossTenant_Returns403 validates that a SPARQL query
// scoped to a foreign tenant's ontology returns 403.
//
// TC-034: Cross-tenant SPARQL query returns 403.
func TestRBAC_TC034_SPARQL_CrossTenant_Returns403(t *testing.T) {
	// Given: u_viewer_A is authenticated, ont_B exists in tenant_B.
	jwt, ok := setupTest(t, userViewerA)
	if !ok {
		return
	}

	// When: POST /api/v1/sparql with query default-graph-uri=ont_B.
	sparqlBody := map[string]string{
		"query":             "SELECT ?s ?p ?o WHERE { ?s ?p ?o }",
		"default-graph-uri": ontB,
	}
	payload, _ := json.Marshal(sparqlBody)

	resp := doRequest(t, "POST", apiBase+"/sparql", payload, jwt, "")
	if resp == nil {
		return
	}
	defer resp.Body.Close()

	// Then: HTTP 403 — cross-tenant SPARQL access denied.
	assertStatus(t, resp, 403, "TC-034: cross-tenant SPARQL query")
}
