package mcp

import (
	"strings"
)

// ToolNamePrefix namespaces every MCP tool exposed to LLM agents, so MCP
// tools can never collide with or shadow built-in tool names.
const ToolNamePrefix = "mcp_"

// maxToolNameBytes matches the strictest function-name length limit among
// the LLM providers PentAGI integrates with (OpenAI caps function names at
// 64 characters).
const maxToolNameBytes = 64

// ToolName builds the agent-visible function name for a tool announced by a
// server: mcp_<server>_<tool>. Both parts are sanitized to [a-z0-9_]; the
// result is truncated to maxToolNameBytes. It returns an empty string when
// nothing usable remains.
func ToolName(server, tool string) string {
	var builder strings.Builder
	builder.WriteString(ToolNamePrefix)
	for _, part := range []string{server, tool} {
		sanitized := sanitizeName(part)
		if sanitized == "" {
			continue
		}
		if builder.Len() > len(ToolNamePrefix) {
			builder.WriteByte('_')
		}
		builder.WriteString(sanitized)
	}

	name := builder.String()
	if len(name) > maxToolNameBytes {
		name = name[:maxToolNameBytes]
	}
	return strings.Trim(name, "_")
}

// localToolName inverts ToolName for a registered tool: it strips the
// mcp_<server>_ prefix, returning the name the server itself announced.
// Callers must only pass names produced by ToolName with the same server.
func localToolName(server, name string) string {
	prefix := ToolNamePrefix
	if sanitized := sanitizeName(server); sanitized != "" {
		prefix += sanitized + "_"
	}
	return strings.TrimPrefix(name, prefix)
}

// sanitizeName lowercases the value and replaces every character outside
// ASCII [a-z0-9_] with an underscore, collapsing runs of underscores and
// trimming them from the ends. MCP tool names conventionally use
// [a-zA-Z0-9_-], while several LLM providers only accept [a-zA-Z0-9_] in
// function names; non-ASCII names therefore collapse to an empty string.
func sanitizeName(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))

	previousUnderscore := false
	for _, r := range value {
		switch {
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r + 'a' - 'A')
			previousUnderscore = false
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			builder.WriteRune(r)
			previousUnderscore = false
		case !previousUnderscore:
			builder.WriteByte('_')
			previousUnderscore = true
		}
	}

	return strings.Trim(builder.String(), "_")
}
