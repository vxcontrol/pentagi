package config

import (
	"fmt"
	"regexp"
	"strings"
)

// tenantIDRegex is the intersection of every constraint TENANT_ID must satisfy
// simultaneously:
//
//   - Docker object name component: [a-zA-Z0-9][a-zA-Z0-9_.-]*
//   - PostgreSQL identifier: <= 63 bytes, unquoted-safe as [a-z_][a-z0-9_]*
//   - Filesystem path segment: no separators, no "." / ".."
//   - Graphiti / Neo4j group id: parsed on a hyphen boundary, so no hyphens
//
// The hyphen is deliberately excluded: it is the separator used by group ids and
// docker object names, and admitting it would make those strings ambiguous to
// parse. 32 characters leaves ample headroom under PostgreSQL's 63-byte limit
// even once prefixes are applied.
var tenantIDRegex = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)

// ValidateTenantID reports whether the configured TENANT_ID is usable. An empty
// tenant id is always valid and selects single-instance mode. Anything else must
// match tenantIDRegex; callers are expected to treat a non-nil result as fatal,
// because silently normalising an invalid value could collapse two distinct
// tenants onto one namespace — exactly the collision tenancy exists to prevent.
func (c *Config) ValidateTenantID() error {
	if c == nil || c.TenantID == "" {
		return nil
	}

	if !tenantIDRegex.MatchString(c.TenantID) {
		return fmt.Errorf(
			"invalid TENANT_ID %q: must match %s (lowercase letter first, then lowercase "+
				"letters, digits or underscores, max 32 characters)",
			c.TenantID, tenantIDRegex.String(),
		)
	}

	return nil
}

// HasTenant reports whether this instance runs in multi-tenant mode.
func (c *Config) HasTenant() bool { return c != nil && c.TenantID != "" }

// TenantPrefix returns "<tenant>-", or "" when no tenant is configured. Prefer
// ScopedName when building a complete name; this exists for the call sites that
// must hand a prefix to another package.
func (c *Config) TenantPrefix() string {
	if c == nil || c.TenantID == "" {
		return ""
	}
	return c.TenantID + "-"
}

// ScopedName prefixes a docker object name, cookie name, or any other
// hyphen-separated identifier with this instance's tenant.
//
//	ScopedName("pentagi-terminal-1") -> "pentagi-terminal-1"      (no tenant)
//	ScopedName("pentagi-terminal-1") -> "acme-pentagi-terminal-1" (tenant "acme")
//
// The prefix goes in front rather than in the middle so that tenant-owned
// objects fall outside existing "pentagi-*" prefix sweeps instead of inside
// them — see the installer's volume garbage collector.
func (c *Config) ScopedName(base string) string {
	if c == nil || c.TenantID == "" {
		return base
	}
	return c.TenantID + "-" + base
}

// GroupID builds the Graphiti / Neo4j group identifier for a flow.
//
//	GroupID(42) -> "flow-42"      (no tenant)
//	GroupID(42) -> "acme-flow-42" (tenant "acme")
//
// ParseGroupID is the exact inverse; the two must always change together.
func (c *Config) GroupID(flowID int64) string {
	return c.TenantPrefix() + fmt.Sprintf("flow-%d", flowID)
}

// ParseGroupID is the inverse of GroupID. It strips this instance's tenant
// prefix and returns the flow id, rejecting group ids that belong to a different
// tenant so that one instance can never resolve another's knowledge-graph data.
//
// This must only be applied to strings that came back from Graphiti or Neo4j —
// never to a value supplied by an API client, which would let the client choose
// its own namespace.
func (c *Config) ParseGroupID(groupID string) (int64, error) {
	rest := groupID
	if c != nil && c.TenantID != "" {
		prefix := c.TenantID + "-"
		if !strings.HasPrefix(groupID, prefix) {
			return 0, fmt.Errorf("group id %q does not belong to tenant %q", groupID, c.TenantID)
		}
		rest = strings.TrimPrefix(groupID, prefix)
	}

	var flowID int64
	// Sscanf alone would accept trailing garbage ("flow-1x"), so verify the
	// round-trip instead of trusting the scan.
	if _, err := fmt.Sscanf(rest, "flow-%d", &flowID); err != nil {
		return 0, fmt.Errorf("invalid groupId format: %q (expected %q)", groupID, c.GroupID(0))
	}
	if c.GroupID(flowID) != groupID {
		return 0, fmt.Errorf("invalid groupId format: %q (expected %q)", groupID, c.GroupID(flowID))
	}

	return flowID, nil
}

// TenantUserID namespaces a telemetry "user" identity. Instances share the
// seeded admin@pentagi.com account, so without this every tenant's traces would
// collapse onto one Langfuse user. Returns the address unchanged in
// single-instance mode.
func (c *Config) TenantUserID(mail string) string {
	if c == nil || c.TenantID == "" {
		return mail
	}
	return c.TenantID + "/" + mail
}

// TenantTags appends a "tenant:<id>" tag so traces from several instances
// sharing one Langfuse project stay filterable. Returns exactly the base tags in
// single-instance mode, so the emitted payload is unchanged.
func (c *Config) TenantTags(base ...string) []string {
	if c == nil || c.TenantID == "" {
		return base
	}
	return append(append(make([]string, 0, len(base)+1), base...), "tenant:"+c.TenantID)
}

// TenantLabel returns a display prefix for human-facing telemetry names
// ("<tenant> " or ""), used for Langfuse trace names.
func (c *Config) TenantLabel() string {
	if c == nil || c.TenantID == "" {
		return ""
	}
	return c.TenantID + " "
}

// SchemaName returns the PostgreSQL schema this instance owns. Single-instance
// deployments keep using "public" so nothing about their layout changes.
func (c *Config) SchemaName() string {
	if c == nil || c.TenantID == "" {
		return "public"
	}
	return c.TenantID
}

// ExtensionSchema returns the schema every tenant's search_path must include for
// shared extensions to resolve. See DATABASE_EXTENSIONS_SCHEMA in
// backend/docs/config.md for details; defaults to "public".
func (c *Config) ExtensionSchema() string {
	if c == nil || c.DatabaseExtensionsSchema == "" {
		return "public"
	}
	return c.DatabaseExtensionsSchema
}

// AuthSalt returns the effective salt for cookie and JWT key derivation. Mixing
// the tenant in makes one instance's session cookies and API tokens
// cryptographically invalid on another even when COOKIE_SIGNING_SALT is shared
// across instances — without it, identical salts plus identical user id
// sequences would let a session from one tenant authenticate against another.
func (c *Config) AuthSalt() string {
	if c == nil {
		return ""
	}
	if c.TenantID == "" {
		return c.CookieSigningSalt
	}
	return c.CookieSigningSalt + "|tenant|" + c.TenantID
}

// TenantLabels returns the docker object labels identifying this instance's
// resources. Returns nil when no tenant is configured, so labels stay absent and
// created objects are byte-identical to today.
func (c *Config) TenantLabels() map[string]string {
	if c == nil || c.TenantID == "" {
		return nil
	}
	return map[string]string{TenantLabelKey: c.TenantID}
}

// TenantLabelKey is the docker label carrying the owning tenant id. Sweeps over
// the daemon should filter on this rather than matching name prefixes.
const TenantLabelKey = "pentagi.tenant"
