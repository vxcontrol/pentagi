package mcp

import (
	"sort"
	"strings"
)

// toolSelector implements the MCP_ALLOWED_TOOLS / MCP_DENIED_TOOLS env
// filters. A selector matches tools by "server" (the whole server) or
// "server/tool" (one tool); the special entry "*" matches everything.
//
// Semantics: when the allow-list is non-empty, only matching tools pass.
// The deny-list is applied afterwards and always wins, so
// allow "*" + deny "burp" exposes everything except the burp server.
type toolSelector struct {
	entries  map[string]struct{}
	denied   map[string]struct{}
	allowAll bool
}

func newToolSelector(allowed, denied []string) *toolSelector {
	selector := &toolSelector{
		entries: parseEntries(allowed),
		denied:  parseEntries(denied),
	}
	_, selector.allowAll = selector.entries["*"]
	return selector
}

func parseEntries(values []string) map[string]struct{} {
	entries := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		entry := sanitizeSelectorEntry(value)
		if entry == "" {
			continue
		}
		entries[entry] = struct{}{}
	}
	return entries
}

// sanitizeSelectorEntry normalizes a selector to the sanitized
// server/tool vocabulary produced by ToolName, so operators can write the
// names exactly as the MCP server announces them (case and separators are
// forgiving). The wildcards "*" (everything) and "server/*" (every tool of
// one server) are preserved verbatim.
func sanitizeSelectorEntry(value string) string {
	if value == "*" {
		return "*"
	}

	parts := strings.SplitN(value, "/", 2)
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "*" && i == len(parts)-1 {
			// trailing wildcard: "server/*"
			parts[i] = "*"
			continue
		}
		parts[i] = sanitizeName(part)
	}

	return strings.Join(parts, "/")
}

// disabled reports that no filter is configured, allowing a fast path.
func (s *toolSelector) disabled() bool {
	return s == nil || (len(s.entries) == 0 && len(s.denied) == 0)
}

// allows reports whether the tool passes the allow/deny filters.
func (s *toolSelector) allows(server, tool string) bool {
	if s.disabled() {
		return true
	}

	if matchSelector(s.denied, server, tool) {
		return false
	}

	if len(s.entries) == 0 || s.allowAll {
		return true
	}

	return matchSelector(s.entries, server, tool)
}

func matchSelector(entries map[string]struct{}, server, tool string) bool {
	if _, ok := entries["*"]; ok {
		return true
	}
	sanitizedServer := sanitizeName(server)
	if _, ok := entries[sanitizedServer]; ok {
		return true
	}
	if _, ok := entries[sanitizedServer+"/*"]; ok {
		return true
	}
	_, ok := entries[sanitizedServer+"/"+sanitizeName(tool)]
	return ok
}

// SortedEntries returns the configured entries in a stable order; used by
// tests and diagnostics.
func (s *toolSelector) SortedEntries() []string {
	if s == nil {
		return nil
	}
	entries := make([]string, 0, len(s.entries)+len(s.denied))
	for entry := range s.entries {
		entries = append(entries, entry)
	}
	for entry := range s.denied {
		entries = append(entries, "-"+entry)
	}
	sort.Strings(entries)
	return entries
}
