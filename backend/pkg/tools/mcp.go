package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/mcp"

	"github.com/vxcontrol/langchaingo/llms"
)

// MCPClient is the part of the MCP client the tool layer relies on. It is
// an interface so tests can substitute a fake without running servers.
type MCPClient interface {
	IsEnabled() bool
	Tools(agentContext string) []mcp.Tool
	CallTool(ctx context.Context, name string, args json.RawMessage) (string, error)
}

// mcpTool dispatches every namespaced mcp_* tool call to the shared MCP
// client, which routes it to the owning server. One instance serves every
// MCP tool of an executor.
type mcpTool struct {
	client MCPClient
}

// NewMCPTool creates a tool that forwards calls to external MCP servers.
func NewMCPTool(client MCPClient) Tool {
	return &mcpTool{client: client}
}

// IsAvailable checks whether any MCP server is connected.
func (t *mcpTool) IsAvailable() bool {
	return t.client != nil && t.client.IsEnabled()
}

// Handle executes the MCP tool call and returns its output to the agent.
func (t *mcpTool) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if !t.IsAvailable() {
		return "", fmt.Errorf("MCP tool '%s' is not available: no MCP server is connected", name)
	}

	return t.client.CallTool(ctx, name, args)
}

// mcpToolDefinition converts one discovered MCP tool into the function
// definition handed to the LLM. The server-announced JSON Schema is passed
// through as the parameters object; a schema that does not decode into a
// JSON object is dropped, which providers treat as "no arguments".
func mcpToolDefinition(tool mcp.Tool) llms.FunctionDefinition {
	definition := llms.FunctionDefinition{
		Name:        tool.Name,
		Description: tool.Description,
	}

	if definition.Description == "" {
		definition.Description = fmt.Sprintf(
			"External tool '%s' exposed by the '%s' MCP server", tool.Name, tool.Server,
		)
	}

	if len(tool.InputSchema) > 0 {
		var parameters map[string]any
		if err := json.Unmarshal(tool.InputSchema, &parameters); err == nil {
			definition.Parameters = parameters
		}
	}

	return definition
}

// sharedMCPClient is a process-level singleton: MCP server connections are
// long-lived and expensive, and every flow executor talks to the same
// configured fleet. cfg.MCPServers is fixed at startup, so the first
// caller's config is representative for all callers.
var (
	sharedMCPClient     MCPClient
	sharedMCPClientOnce sync.Once
)

// SharedMCPClient connects to the configured MCP servers exactly once per
// process and returns the client, or nil when the integration is disabled.
// Connection failures are logged by the client and degrade to "disabled".
func SharedMCPClient(cfg *config.Config) MCPClient {
	sharedMCPClientOnce.Do(func() {
		if cfg == nil || len(cfg.MCPServers) == 0 {
			return
		}
		sharedMCPClient = mcp.NewClient(context.Background(), cfg)
	})
	return sharedMCPClient
}

// CloseSharedMCPClient terminates the shared MCP server connections
// (stopping stdio child processes). Safe to call when disabled.
func CloseSharedMCPClient() {
	if client, ok := sharedMCPClient.(*mcp.Client); ok {
		client.Close()
	}
}

// mcpToolsClient returns the executor-specific client when one was injected
// (tests), falling back to the shared process-level client.
func (fte *flowToolsExecutor) mcpToolsClient() MCPClient {
	if fte.mcpClient != nil {
		return fte.mcpClient
	}
	return SharedMCPClient(fte.cfg)
}

// appendMCPTools exposes the MCP tools visible to the agent context as
// first-class tools of the executor, next to the built-in ones.
func (fte *flowToolsExecutor) appendMCPTools(ce *customExecutor, agentContext database.MsgchainType) {
	client := fte.mcpToolsClient()
	if client == nil || !client.IsEnabled() {
		return
	}

	tool := NewMCPTool(client)
	for _, discovered := range client.Tools(string(agentContext)) {
		definition := mcpToolDefinition(discovered)
		ce.definitions = append(ce.definitions, definition)
		ce.handlers[definition.Name] = tool.Handle
	}
}
