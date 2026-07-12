package keycloak

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type RealmConfig struct {
	Realm             string         `json:"realm"`
	Enabled           bool           `json:"enabled"`
	Roles             RolesConfig    `json:"roles"`
	Clients           []ClientConfig `json:"clients"`
	IdentityProviders []Provider     `json:"identityProviders"`
	Groups            []any          `json:"groups"`
	DefaultGroups     []string       `json:"defaultGroups"`
}

type RolesConfig struct {
	Realm []RoleConfig `json:"realm"`
}

type RoleConfig struct {
	Name       string          `json:"name"`
	Composite  bool            `json:"composite"`
	Composites *RoleComposites `json:"composites,omitempty"`
}

type RoleComposites struct {
	Realm []string `json:"realm"`
}

type ClientConfig struct {
	ClientID string `json:"clientId"`
	Name     string `json:"name"`
	Enabled  bool   `json:"enabled"`
	Public   bool   `json:"publicClient"`
	Protocol string `json:"protocol"`
}

type Provider struct {
	Alias      string `json:"alias"`
	Enabled    bool   `json:"enabled"`
	ProviderID string `json:"providerId"`
}

var expectedRoles = []string{
	"Viewer", "Editor", "Maintainer", "Owner",
	"SupportEngineer", "SRE", "SecurityLead", "ProductOwner",
}

func fixturePath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "realm-import.json")
}

func loadRealm(t *testing.T) RealmConfig {
	t.Helper()
	data, err := os.ReadFile(fixturePath())
	if err != nil {
		t.Fatalf("failed to read realm-import.json: %v", err)
	}
	var cfg RealmConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse realm-import.json: %v", err)
	}
	return cfg
}

func TestRealmConfig_FileExists(t *testing.T) {
	if _, err := os.Stat(fixturePath()); err != nil {
		t.Fatalf("realm-import.json not found: %v", err)
	}
}

func TestRealmConfig_AllRolesPresent(t *testing.T) {
	// @hlv role_completeness
	// @ctx: contract AUTH-RBAC-001 requires all 8 seeded roles
	cfg := loadRealm(t)
	actual := make(map[string]bool)
	for _, r := range cfg.Roles.Realm {
		actual[r.Name] = true
	}
	for _, expected := range expectedRoles {
		if !actual[expected] {
			t.Errorf("missing required role: %s", expected)
		}
	}
}

func TestRealmConfig_OwnerComposite(t *testing.T) {
	// @hlv composite_role_hierarchy
	// @ctx: Owner must composite Viewer + Editor + Maintainer
	cfg := loadRealm(t)
	for _, r := range cfg.Roles.Realm {
		if r.Name == "Owner" {
			if !r.Composite {
				t.Error("Owner role must be composite")
			}
			if r.Composites == nil {
				t.Fatal("Owner role must have composites")
			}
			expected := map[string]bool{"Viewer": true, "Editor": true, "Maintainer": true}
			for _, name := range r.Composites.Realm {
				delete(expected, name)
			}
			if len(expected) > 0 {
				t.Errorf("Owner composite missing roles: %v", expected)
			}
			return
		}
	}
	t.Error("Owner role not found")
}

func TestRealmConfig_ProductOwnerComposite(t *testing.T) {
	// @hlv composite_role_hierarchy
	// @ctx: ProductOwner must composite Viewer + Editor
	cfg := loadRealm(t)
	for _, r := range cfg.Roles.Realm {
		if r.Name == "ProductOwner" {
			if !r.Composite {
				t.Error("ProductOwner role must be composite")
			}
			if r.Composites == nil {
				t.Fatal("ProductOwner role must have composites")
			}
			expected := map[string]bool{"Viewer": true, "Editor": true}
			for _, name := range r.Composites.Realm {
				delete(expected, name)
			}
			if len(expected) > 0 {
				t.Errorf("ProductOwner composite missing roles: %v", expected)
			}
			return
		}
	}
	t.Error("ProductOwner role not found")
}

func TestRealmConfig_ClientVedoSpaExists(t *testing.T) {
	// @hlv client_config_complete
	// @ctx: contract AUTH-RBAC-001 requires vedo-spa client
	cfg := loadRealm(t)
	for _, c := range cfg.Clients {
		if c.ClientID == "vedo-spa" {
			if !c.Enabled {
				t.Error("vedo-spa client must be enabled")
			}
			if !c.Public {
				t.Error("vedo-spa client must be public")
			}
			if c.Protocol != "openid-connect" {
				t.Errorf("vedo-spa protocol must be openid-connect, got %s", c.Protocol)
			}
			return
		}
	}
	t.Error("client vedo-spa not found")
}

func TestRealmConfig_CliVedoCliExists(t *testing.T) {
	// @hlv client_config_complete
	// @ctx: contract AUTH-RBAC-001 requires vedo-cli client with service accounts
	cfg := loadRealm(t)
	for _, c := range cfg.Clients {
		if c.ClientID == "vedo-cli" {
			if !c.Enabled {
				t.Error("vedo-cli client must be enabled")
			}
			if c.Public {
				t.Error("vedo-cli client must be confidential (not public)")
			}
			return
		}
	}
	t.Error("client vedo-cli not found")
}

func TestRealmConfig_OAuthProvidersEnabled(t *testing.T) {
	// @hlv oauth_providers_configured
	// @ctx: resolved Q3 — Yandex (primary) + VK (secondary) for MVP
	cfg := loadRealm(t)
	providers := make(map[string]bool)
	for _, p := range cfg.IdentityProviders {
		if p.Enabled {
			providers[p.Alias] = true
		}
	}
	if !providers["yandex"] {
		t.Error("Yandex OAuth provider must be enabled (primary)")
	}
	if !providers["vk"] {
		t.Error("VK OAuth provider must be enabled (secondary)")
	}
}

func TestRealmConfig_RealmName(t *testing.T) {
	// @hlv realm_name_vedo_core
	// @ctx: contract AUTH-RBAC-001 requires realm name "vedo-core"
	cfg := loadRealm(t)
	if cfg.Realm != "vedo-core" {
		t.Errorf("realm name must be 'vedo-core', got '%s'", cfg.Realm)
	}
}

func TestRealmConfig_DefaultGroupsIncludeViewer(t *testing.T) {
	// @hlv default_role_assignment
	// @ctx: new users default to Viewer role
	cfg := loadRealm(t)
	found := false
	for _, g := range cfg.DefaultGroups {
		if g == "Viewer" {
			found = true
			break
		}
	}
	if !found {
		t.Error("default groups must include 'Viewer'")
	}
}
