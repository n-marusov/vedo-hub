package auth

// Validates: REQ-CON.STACK.publishing-extension

// RED-phase test (TDD): the publish-gate weight constant MUST exist and
// equal 2 (Maintainer) per ADR-DES.INFRA.publishing-extension. The publish
// action (release creation) is gated at Maintainer+ (role weight >= 2) with
// no separate publisher role; Owner (weight 3) can also publish.
//
// This test does not compile until the `RoleWeightPublish` constant is added
// (Task 20 of the plan) — a classic RED state. Once the constant exists, the
// assertion locks in the required value.

import "testing"

// TestRoleWeightPublish_ShouldEqualMaintainerWeight verifies the publish
// gate weight constant equals the Maintainer role weight (2).
func TestRoleWeightPublish_ShouldEqualMaintainerWeight(t *testing.T) {
	if RoleWeightPublish != 2 {
		t.Errorf("RoleWeightPublish = %d, want 2 (Maintainer weight per ADR-DES.INFRA.publishing-extension)", RoleWeightPublish)
	}
}

// TestRoleWeightPublish_ShouldMatchRoleWeightMap verifies the constant is
// consistent with the roleWeight map's maintainer entry. The map keys are
// lowercase (matching Keycloak realm roles); RoleMaintainer is the display
// name, so compare against the lowercase key.
func TestRoleWeightPublish_ShouldMatchRoleWeightMap(t *testing.T) {
	maintainerWeight := roleWeight["maintainer"]
	if RoleWeightPublish != maintainerWeight {
		t.Errorf("RoleWeightPublish = %d, roleWeight[maintainer] = %d — must match",
			RoleWeightPublish, maintainerWeight)
	}
}

// TestRoleWeightPublish_ShouldBeBelowOwnerWeight verifies the publish gate
// is strictly below the Owner weight so Owner can also publish.
func TestRoleWeightPublish_ShouldBeBelowOwnerWeight(t *testing.T) {
	ownerWeight := roleWeight["owner"]
	if RoleWeightPublish >= ownerWeight {
		t.Errorf("RoleWeightPublish = %d, want < Owner weight %d", RoleWeightPublish, ownerWeight)
	}
}
