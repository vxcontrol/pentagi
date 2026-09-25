package config

import (
	"strings"
	"testing"
)

// Rows are keyed by helper; the untenanted column holds the literal each name had before tenancy existed.
func TestTenant_ScopesNamesOnlyWhenATenantIsSet(t *testing.T) {
	plain := &Config{CookieSigningSalt: "salt"}
	acme := &Config{TenantID: "acme", CookieSigningSalt: "salt"}

	tests := []struct {
		what       string
		get        func(*Config) string
		untenanted string
		tenanted   string
	}{
		{"docker container name", func(c *Config) string { return c.ScopedName("pentagi-terminal-1") }, "pentagi-terminal-1", "acme-pentagi-terminal-1"},
		{"session cookie name", func(c *Config) string { return c.ScopedName("auth") }, "auth", "acme-auth"},
		{"oauth state cookie", func(c *Config) string { return c.TenantPrefix() + "state" }, "state", "acme-state"},
		{"graphiti group id", func(c *Config) string { return c.GroupID(1) }, "flow-1", "acme-flow-1"},
		{"large graphiti group id", func(c *Config) string { return c.GroupID(9999999999) }, "flow-9999999999", "acme-flow-9999999999"},
		{"postgres schema", (*Config).SchemaName, "public", "acme"},
		{"cookie and jwt salt", (*Config).AuthSalt, "salt", "salt|tenant|acme"},
		{"langfuse trace prefix", (*Config).TenantLabel, "", "acme "},
		{"langfuse user id", func(c *Config) string { return c.TenantUserID("admin@pentagi.com") }, "admin@pentagi.com", "acme/admin@pentagi.com"},
	}

	for _, tt := range tests {
		t.Run(tt.what, func(t *testing.T) {
			if got := tt.get(plain); got != tt.untenanted {
				t.Errorf("without a tenant = %q, want %q (backward compatibility broken)", got, tt.untenanted)
			}
			if got := tt.get(acme); got != tt.tenanted {
				t.Errorf("with tenant acme = %q, want %q", got, tt.tenanted)
			}
		})
	}

	if plain.HasTenant() || !acme.HasTenant() {
		t.Errorf("HasTenant() = %v without a tenant and %v with one", plain.HasTenant(), acme.HasTenant())
	}
	if got := plain.TenantLabels(); got != nil {
		t.Errorf("TenantLabels() = %v without a tenant, want nil so docker objects stay unlabelled", got)
	}
	if got := acme.TenantLabels()[TenantLabelKey]; got != "acme" {
		t.Errorf("TenantLabels()[%s] = %q, want %q", TenantLabelKey, got, "acme")
	}
}

// The installer force-removes volumes matching pentagi-terminal-*-data; a tenant's volumes must fall outside it.
func TestTenant_ScopedName_EscapesTheInstallerVolumeSweep(t *testing.T) {
	sweepMatches := func(name string) bool {
		return strings.HasPrefix(name, "pentagi-terminal-") && strings.HasSuffix(name, "-data")
	}

	if tenantVolume := (&Config{TenantID: "acme"}).ScopedName("pentagi-terminal-1") + "-data"; sweepMatches(tenantVolume) {
		t.Errorf("tenant volume %q matches the installer sweep; it would be destroyed by another tenant's purge", tenantVolume)
	}
	if defaultVolume := (&Config{}).ScopedName("pentagi-terminal-1") + "-data"; !sweepMatches(defaultVolume) {
		t.Errorf("default volume %q no longer matches the installer sweep; single-instance cleanup would break", defaultVolume)
	}
}

func TestTenant_ParseGroupID_InvertsGroupID(t *testing.T) {
	for _, tenantID := range []string{"", "acme"} {
		c := &Config{TenantID: tenantID}
		for _, flowID := range []int64{0, 1, 42, 9999999999} {
			gid := c.GroupID(flowID)
			got, err := c.ParseGroupID(gid)
			if err != nil {
				t.Errorf("tenant %q: ParseGroupID(%q) returned error: %v", tenantID, gid, err)
				continue
			}
			if got != flowID {
				t.Errorf("tenant %q: ParseGroupID(%q) = %d, want %d", tenantID, gid, got, flowID)
			}
		}
	}
}

func TestTenant_ParseGroupID_RejectsForeignAndMalformedIDs(t *testing.T) {
	tests := []struct {
		tenantID string
		groupIDs []string
		wantErr  string
	}{
		{tenantID: "acme", groupIDs: []string{"flow-1", "other-flow-1", "acme2-flow-1", ""}, wantErr: `does not belong to tenant "acme"`},
		{tenantID: "", groupIDs: []string{"flow-", "flow-abc", "flow-1x", "notflow-1", "1", ""}, wantErr: "invalid groupId format"},
	}

	for _, tt := range tests {
		c := &Config{TenantID: tt.tenantID}
		for _, gid := range tt.groupIDs {
			if _, err := c.ParseGroupID(gid); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("tenant %q: ParseGroupID(%q) error = %v, want one containing %q", tt.tenantID, gid, err, tt.wantErr)
			}
		}
	}
}

func TestTenant_ValidateTenantID_AcceptsOnlyShortLowercaseIdentifiers(t *testing.T) {
	valid := []string{
		"",  // single-instance mode
		"a", // minimum
		"acme",
		"acme2",
		"a_b_c",
		"tenant_01",
		strings.Repeat("a", 32), // maximum length
	}
	for _, id := range valid {
		c := &Config{TenantID: id}
		if err := c.ValidateTenantID(); err != nil {
			t.Errorf("ValidateTenantID(%q) = %v, want nil", id, err)
		}
	}

	invalid := []string{
		"Acme",                  // uppercase breaks unquoted postgres identifiers
		"1acme",                 // must start with a letter (docker + postgres)
		"_acme",                 // must start with a letter
		"acme-prod",             // hyphen is the group-id / docker-name separator
		"acme.prod",             // dot is not valid in a postgres identifier
		"acme prod",             // whitespace
		"../etc",                // path traversal
		"acme/prod",             // path separator
		"acme;DROP SCHEMA",      // sql metacharacters
		strings.Repeat("a", 33), // over the length limit
	}
	for _, id := range invalid {
		c := &Config{TenantID: id}
		if err := c.ValidateTenantID(); err == nil || !strings.Contains(err.Error(), "invalid TENANT_ID") {
			t.Errorf("ValidateTenantID(%q) = %v, want an invalid TENANT_ID error", id, err)
		}
	}
}

// Two instances with different tenant ids must not produce one colliding name for the same flow id.
func TestTenant_IsolationIsMutual(t *testing.T) {
	alpha := &Config{TenantID: "alpha", CookieSigningSalt: "shared"}
	beta := &Config{TenantID: "beta", CookieSigningSalt: "shared"}
	plain := &Config{CookieSigningSalt: "shared"}

	const flowID = 1
	for _, probe := range []struct {
		what string
		fn   func(*Config) string
	}{
		{"container name", func(c *Config) string { return c.ScopedName("pentagi-terminal-1") }},
		{"group id", func(c *Config) string { return c.GroupID(flowID) }},
		{"schema", func(c *Config) string { return c.SchemaName() }},
		{"cookie name", func(c *Config) string { return c.ScopedName("auth") }},
		{"auth salt", func(c *Config) string { return c.AuthSalt() }},
	} {
		a, b, p := probe.fn(alpha), probe.fn(beta), probe.fn(plain)
		if a == b {
			t.Errorf("%s collides between tenants: both %q", probe.what, a)
		}
		if a == p || b == p {
			t.Errorf("%s collides with the untenanted instance: %q / %q / %q", probe.what, a, b, p)
		}
	}
}
