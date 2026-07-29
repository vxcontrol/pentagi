package config

import (
	"strings"
	"testing"
)

// TestTenantHelpersAreNoOpWhenEmpty is the backward-compatibility contract: with
// TENANT_ID unset every helper must return exactly the value the codebase used
// before tenancy existed.
func TestTenantHelpersAreNoOpWhenEmpty(t *testing.T) {
	c := &Config{CookieSigningSalt: "salt"}

	if c.HasTenant() {
		t.Error("HasTenant() = true, want false for empty tenant id")
	}
	if got := c.TenantPrefix(); got != "" {
		t.Errorf("TenantPrefix() = %q, want %q", got, "")
	}
	if got := c.ScopedName("pentagi-terminal-1"); got != "pentagi-terminal-1" {
		t.Errorf("ScopedName() = %q, want %q", got, "pentagi-terminal-1")
	}
	if got := c.ScopedName("auth"); got != "auth" {
		t.Errorf("ScopedName(auth) = %q, want %q", got, "auth")
	}
	if got := c.GroupID(42); got != "flow-42" {
		t.Errorf("GroupID(42) = %q, want %q", got, "flow-42")
	}
	if got := c.TenantLabel(); got != "" {
		t.Errorf("TenantLabel() = %q, want %q", got, "")
	}
	if got := c.SchemaName(); got != "public" {
		t.Errorf("SchemaName() = %q, want %q", got, "public")
	}
	if got := c.AuthSalt(); got != "salt" {
		t.Errorf("AuthSalt() = %q, want %q (must equal CookieSigningSalt verbatim)", got, "salt")
	}
	if got := c.TenantLabels(); got != nil {
		t.Errorf("TenantLabels() = %v, want nil so docker objects stay unlabelled", got)
	}
}

func TestTenantHelpersWithTenant(t *testing.T) {
	c := &Config{TenantID: "acme", CookieSigningSalt: "salt"}

	if !c.HasTenant() {
		t.Error("HasTenant() = false, want true")
	}
	if got, want := c.TenantPrefix(), "acme-"; got != want {
		t.Errorf("TenantPrefix() = %q, want %q", got, want)
	}
	if got, want := c.ScopedName("pentagi-terminal-1"), "acme-pentagi-terminal-1"; got != want {
		t.Errorf("ScopedName() = %q, want %q", got, want)
	}
	if got, want := c.GroupID(42), "acme-flow-42"; got != want {
		t.Errorf("GroupID(42) = %q, want %q", got, want)
	}
	if got, want := c.TenantLabel(), "acme "; got != want {
		t.Errorf("TenantLabel() = %q, want %q", got, want)
	}
	if got, want := c.SchemaName(), "acme"; got != want {
		t.Errorf("SchemaName() = %q, want %q", got, want)
	}
	if got := c.AuthSalt(); got == "salt" {
		t.Error("AuthSalt() must differ from the raw salt when a tenant is configured")
	}
	if got := c.TenantLabels(); got[TenantLabelKey] != "acme" {
		t.Errorf("TenantLabels()[%s] = %q, want %q", TenantLabelKey, got[TenantLabelKey], "acme")
	}
}

// TestScopedNameEscapesInstallerVolumeSweep pins the property that motivated
// putting the tenant prefix in front rather than in the middle: the installer's
// garbage collector force-removes volumes matching
// HasPrefix("pentagi-terminal-") && HasSuffix("-data"), and a tenant's volumes
// must fall outside that filter.
func TestScopedNameEscapesInstallerVolumeSweep(t *testing.T) {
	tenant := &Config{TenantID: "acme"}
	def := &Config{}

	tenantVolume := tenant.ScopedName("pentagi-terminal-1") + "-data"
	defaultVolume := def.ScopedName("pentagi-terminal-1") + "-data"

	sweepMatches := func(name string) bool {
		return strings.HasPrefix(name, "pentagi-terminal-") && strings.HasSuffix(name, "-data")
	}

	if sweepMatches(tenantVolume) {
		t.Errorf("tenant volume %q matches the installer sweep; it would be destroyed by another tenant's purge", tenantVolume)
	}
	if !sweepMatches(defaultVolume) {
		t.Errorf("default volume %q no longer matches the installer sweep; single-instance cleanup would break", defaultVolume)
	}
}

func TestParseGroupIDRoundTrip(t *testing.T) {
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

func TestParseGroupIDRejectsForeignTenant(t *testing.T) {
	c := &Config{TenantID: "acme"}

	for _, gid := range []string{"flow-1", "other-flow-1", "acme2-flow-1", ""} {
		if _, err := c.ParseGroupID(gid); err == nil {
			t.Errorf("ParseGroupID(%q) succeeded; want rejection for tenant %q", gid, c.TenantID)
		}
	}
}

func TestParseGroupIDRejectsMalformed(t *testing.T) {
	c := &Config{}

	for _, gid := range []string{"flow-", "flow-abc", "flow-1x", "notflow-1", "1", ""} {
		if _, err := c.ParseGroupID(gid); err == nil {
			t.Errorf("ParseGroupID(%q) succeeded; want rejection", gid)
		}
	}
}

func TestValidateTenantID(t *testing.T) {
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
		if err := c.ValidateTenantID(); err == nil {
			t.Errorf("ValidateTenantID(%q) = nil, want error", id)
		}
	}
}

// TestTenantIDFitsPostgresIdentifierLimit guards the derivation in SchemaName
// against PostgreSQL's 63-byte identifier cap.
func TestTenantIDFitsPostgresIdentifierLimit(t *testing.T) {
	c := &Config{TenantID: strings.Repeat("a", 32)}
	if err := c.ValidateTenantID(); err != nil {
		t.Fatalf("max-length tenant id rejected: %v", err)
	}
	if got := len(c.SchemaName()); got > 63 {
		t.Errorf("SchemaName() is %d bytes, exceeds the PostgreSQL 63-byte identifier limit", got)
	}
}

// TestBackwardCompatibilityContract is the guard for the single hard requirement
// of the tenancy work: with TENANT_ID unset, every externally-visible name,
// path, identifier and key must be exactly what it was before tenancy existed.
//
// The literals on the right-hand side are the historical values, written out
// explicitly rather than computed, so that a future refactor of the helpers
// cannot silently move them.
func TestBackwardCompatibilityContract(t *testing.T) {
	c := &Config{
		DataDir:           "./data",
		DockerNetwork:     "pentagi-network",
		CookieSigningSalt: "salt",
		InstallationID:    "3f2504e0-4f89-11d3-9a0c-0305e82c3301",
	}
	if err := c.ValidateTenantID(); err != nil {
		t.Fatalf("empty TENANT_ID must be valid, got %v", err)
	}

	cases := []struct {
		what string
		got  string
		want string
	}{
		{"docker container name", c.ScopedName("pentagi-terminal-1"), "pentagi-terminal-1"},
		{"docker volume name", c.ScopedName("pentagi-terminal-1") + "-data", "pentagi-terminal-1-data"},
		{"docker network", c.DockerNetwork, "pentagi-network"},
		{"session cookie name", c.ScopedName("auth"), "auth"},
		{"oauth state cookie", c.TenantPrefix() + "state", "state"},
		{"oauth nonce cookie", c.TenantPrefix() + "nonce", "nonce"},
		{"graphiti group id", c.GroupID(1), "flow-1"},
		{"graphiti group id (large)", c.GroupID(9999999999), "flow-9999999999"},
		{"postgres schema", c.SchemaName(), "public"},
		{"cookie/jwt salt", c.AuthSalt(), "salt"},
		// Left verbatim, not even filepath.Clean'ed: applyTenantScoping returns
		// before touching it, so an upgraded deployment sees the identical string.
		{"data dir", c.DataDir, "./data"},
		{"langfuse trace prefix", c.TenantLabel(), ""},
		{"installation id", c.InstallationID, "3f2504e0-4f89-11d3-9a0c-0305e82c3301"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s = %q, want %q (backward compatibility broken)", tc.what, tc.got, tc.want)
		}
	}

	if c.TenantLabels() != nil {
		t.Error("docker objects must carry no labels when TENANT_ID is unset")
	}
	if c.HasTenant() {
		t.Error("HasTenant() must be false when TENANT_ID is unset")
	}
}

// TestTenantIsolationIsMutual checks the property that actually matters at
// runtime: two instances with different tenant ids must not produce a single
// colliding name for the same flow id.
func TestTenantIsolationIsMutual(t *testing.T) {
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
