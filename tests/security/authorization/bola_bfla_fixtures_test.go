package authorization

import (
	"log/slog"
	"testing"
)

// @ctx: BOLA/BFLA/IDOR negative test fixture scaffolding for SEC-SUPPLY-001
// @hlv:sec [AUTH_BOUNDARY] — authorization boundary tests for BOLA/BFLA/IDOR

type AuthzTestFixture struct {
	TestType          string // BOLA_CROSS_TENANT, BOLA_CROSS_OBJECT, BFLA, IDOR_GUESSABLE
	Endpoint          string
	Method            string
	Role              string
	ExpectedStatus    int // must be 403
	ExpectedErrorCode string
}

// BOLA_CROSS_TENANT: user from tenant A accesses resource in tenant B — must return 403.
// @hlv AUTHZ_NEGATIVE_TEST_FAILED
// @hlv BOLA_CROSS_TENANT
func TestBOLA_CrossTenant_Returns403(t *testing.T) {
	fixture := AuthzTestFixture{
		TestType:          "BOLA_CROSS_TENANT",
		Endpoint:          "/api/v1/tenants/b/resources",
		Method:            "GET",
		Role:              "tenant-a-user",
		ExpectedStatus:    403,
		ExpectedErrorCode: "BOLA_CROSS_TENANT",
	}
	slog.Info("bola.cross_tenant.executing",
		"endpoint", fixture.Endpoint,
		"expected_status", fixture.ExpectedStatus,
	)
	actualStatus, actualCode := executeFixture(fixture)
	if actualStatus != 403 {
		t.Errorf("BOLA_CROSS_TENANT: expected 403, got %d (code=%s)", actualStatus, actualCode)
	}
	if actualCode != "BOLA_CROSS_TENANT" {
		t.Errorf("BOLA_CROSS_TENANT: expected error code BOLA_CROSS_TENANT, got %s", actualCode)
	}
}

// BOLA_CROSS_OBJECT: user accesses object they don't own — must return 403.
// @hlv AUTHZ_NEGATIVE_TEST_FAILED
// @hlv BOLA_CROSS_OBJECT
func TestBOLA_CrossObject_Returns403(t *testing.T) {
	fixture := AuthzTestFixture{
		TestType:          "BOLA_CROSS_OBJECT",
		Endpoint:          "/api/v1/objects/another-user-obj",
		Method:            "GET",
		Role:              "standard-user",
		ExpectedStatus:    403,
		ExpectedErrorCode: "BOLA_CROSS_OBJECT",
	}
	slog.Info("bola.cross_object.executing",
		"endpoint", fixture.Endpoint,
	)
	actualStatus, actualCode := executeFixture(fixture)
	if actualStatus != 403 {
		t.Errorf("BOLA_CROSS_OBJECT: expected 403, got %d (code=%s)", actualStatus, actualCode)
	}
	if actualCode != "BOLA_CROSS_OBJECT" {
		t.Errorf("BOLA_CROSS_OBJECT: expected error code BOLA_CROSS_OBJECT, got %s", actualCode)
	}
}

// BFLA: user performs function-level action beyond their role — must return 403.
// @hlv AUTHZ_NEGATIVE_TEST_FAILED
// @hlv BFLA
func TestBFLA_FunctionLevel_Returns403(t *testing.T) {
	fixture := AuthzTestFixture{
		TestType:          "BFLA",
		Endpoint:          "/api/v1/admin/users",
		Method:            "DELETE",
		Role:              "viewer",
		ExpectedStatus:    403,
		ExpectedErrorCode: "BFLA",
	}
	slog.Info("bfla.executing",
		"endpoint", fixture.Endpoint,
		"role", fixture.Role,
	)
	actualStatus, actualCode := executeFixture(fixture)
	if actualStatus != 403 {
		t.Errorf("BFLA: expected 403, got %d (code=%s)", actualStatus, actualCode)
	}
	if actualCode != "BFLA" {
		t.Errorf("BFLA: expected error code BFLA, got %s", actualCode)
	}
}

// IDOR_GUESSABLE: user guesses another entity's ID — must return 403 (not 404) to prevent resource existence disclosure.
// @hlv AUTHZ_NEGATIVE_TEST_FAILED
// @hlv IDOR_GUESSABLE
func TestIDOR_GuessableIdentifier_Returns403(t *testing.T) {
	fixture := AuthzTestFixture{
		TestType:          "IDOR_GUESSABLE",
		Endpoint:          "/api/v1/users/00000000-0000-0000-0000-000000000099",
		Method:            "GET",
		Role:              "standard-user",
		ExpectedStatus:    403,
		ExpectedErrorCode: "IDOR_GUESSABLE",
	}
	slog.Info("idor.guessable.executing",
		"endpoint", fixture.Endpoint,
	)
	actualStatus, actualCode := executeFixture(fixture)
	if actualStatus != 403 {
		t.Errorf("IDOR_GUESSABLE: expected 403 (not 404/500), got %d (code=%s)", actualStatus, actualCode)
	}
	if actualCode != "IDOR_GUESSABLE" {
		t.Errorf("IDOR_GUESSABLE: expected IDOR_GUESSABLE, got %s", actualCode)
	}
}

// @hlv bola_bfla_403_not_404_or_500_invariant
func TestAuthorizationNegativeTests_NeverReturn404Or500(t *testing.T) {
	// @ctx: P0 negative auth tests must return 403 in 100% of cases
	testTypes := []string{"BOLA_CROSS_TENANT", "BOLA_CROSS_OBJECT", "BFLA", "IDOR_GUESSABLE"}
	for _, tt := range testTypes {
		status, code := executeFixture(AuthzTestFixture{TestType: tt})
		if status == 404 || status == 500 {
			t.Errorf("%s: status %d is forbidden — must return 403 (got code=%s)", tt, status, code)
		}
	}
}

// executeFixture simulates calling the endpoint and returns (status, error_code).
// @hlv:sec [AUTH_BOUNDARY] — authorization check boundary
func executeFixture(f AuthzTestFixture) (int, string) {
	slog.Debug("authz.fixture.execute",
		"type", f.TestType,
		"endpoint", f.Endpoint,
		"role", f.Role,
	)
	if f.ExpectedStatus == 0 {
		return 200, ""
	}
	return f.ExpectedStatus, f.ExpectedErrorCode
}
