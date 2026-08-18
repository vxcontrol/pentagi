package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"pentagi/pkg/database"
	"pentagi/pkg/mcp"

	"github.com/vxcontrol/langchaingo/llms"
)

type fakeMCPClient struct {
	enabled bool
	tools   []mcp.Tool
	result  string
	err     error

	calledName string
	calledArgs json.RawMessage
}

func (f *fakeMCPClient) IsEnabled() bool {
	return f.enabled
}

func (f *fakeMCPClient) Tools(agentContext string) []mcp.Tool {
	visible := make([]mcp.Tool, 0, len(f.tools))
	for _, tool := range f.tools {
		if len(tool.Contexts) == 0 {
			visible = append(visible, tool)
			continue
		}
		for _, ctx := range tool.Contexts {
			if ctx == agentContext {
				visible = append(visible, tool)
				break
			}
		}
	}
	return visible
}

func (f *fakeMCPClient) CallTool(_ context.Context, name string, args json.RawMessage) (string, error) {
	f.calledName = name
	f.calledArgs = args
	return f.result, f.err
}

func TestMCPToolDefinition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		tool          mcp.Tool
		wantName      string
		wantDesc      string
		wantParams    bool
		paramsHasProp string
	}{
		{
			name: "full definition",
			tool: mcp.Tool{
				Name:        "mcp_burp_scan",
				Description: "launch an active scan",
				InputSchema: json.RawMessage(`{"type":"object","properties":{"url":{"type":"string"}},"required":["url"]}`),
				Server:      "burp",
			},
			wantName:      "mcp_burp_scan",
			wantDesc:      "launch an active scan",
			wantParams:    true,
			paramsHasProp: "url",
		},
		{
			name: "missing description falls back to server hint",
			tool: mcp.Tool{
				Name:   "mcp_burp_spider",
				Server: "burp",
			},
			wantName:   "mcp_burp_spider",
			wantDesc:   "External tool 'mcp_burp_spider' exposed by the 'burp' MCP server",
			wantParams: false,
		},
		{
			name: "non-object schema is dropped",
			tool: mcp.Tool{
				Name:        "mcp_burp_broken",
				Description: "schema is not an object",
				InputSchema: json.RawMessage(`["not","an","object"]`),
				Server:      "burp",
			},
			wantName:   "mcp_burp_broken",
			wantDesc:   "schema is not an object",
			wantParams: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			definition := mcpToolDefinition(tt.tool)
			if definition.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", definition.Name, tt.wantName)
			}
			if definition.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", definition.Description, tt.wantDesc)
			}

			parameters, ok := definition.Parameters.(map[string]any)
			if ok != tt.wantParams {
				t.Fatalf("Parameters present = %v, want %v", ok, tt.wantParams)
			}
			if tt.wantParams {
				if parameters["type"] != "object" {
					t.Errorf("Parameters type = %v, want object", parameters["type"])
				}
				props, ok := parameters["properties"].(map[string]any)
				if !ok {
					t.Fatalf("Parameters properties = %T, want map", parameters["properties"])
				}
				if _, ok := props[tt.paramsHasProp]; !ok {
					t.Errorf("Parameters properties missing %q", tt.paramsHasProp)
				}
			}
		})
	}
}

func TestMCPToolAvailability(t *testing.T) {
	t.Parallel()

	if NewMCPTool(nil).IsAvailable() {
		t.Error("tool with nil client must not be available")
	}

	client := &fakeMCPClient{enabled: false}
	tool := NewMCPTool(client)
	if tool.IsAvailable() {
		t.Error("tool with disabled client must not be available")
	}

	if _, err := tool.Handle(context.Background(), "mcp_burp_scan", nil); err == nil {
		t.Error("Handle must fail when the client is not enabled")
	}
}

func TestMCPToolHandleDispatches(t *testing.T) {
	t.Parallel()

	client := &fakeMCPClient{
		enabled: true,
		result:  "scan finished",
	}
	tool := NewMCPTool(client)

	args := json.RawMessage(`{"url":"https://example.com"}`)
	result, err := tool.Handle(context.Background(), "mcp_burp_scan", args)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}
	if result != "scan finished" {
		t.Errorf("result = %q, want %q", result, "scan finished")
	}
	if client.calledName != "mcp_burp_scan" {
		t.Errorf("client called with name %q, want mcp_burp_scan", client.calledName)
	}
	if string(client.calledArgs) != string(args) {
		t.Errorf("client called with args %s, want %s", client.calledArgs, args)
	}

	client.err = errors.New("server unreachable")
	if _, err := tool.Handle(context.Background(), "mcp_burp_scan", args); err == nil {
		t.Error("Handle must propagate client errors")
	}
}

func TestAppendMCPTools(t *testing.T) {
	t.Parallel()

	discovered := []mcp.Tool{
		{
			Name:        "mcp_burp_scan",
			Description: "launch an active scan",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"url":{"type":"string"}}}`),
			Server:      "burp",
		},
		{
			Name:     "mcp_burp_spider",
			Server:   "burp",
			Contexts: []string{string(database.MsgchainTypeCoder)},
		},
	}

	fte := &flowToolsExecutor{}
	fte.SetMCPClient(&fakeMCPClient{enabled: true, tools: discovered})

	ce := &customExecutor{
		definitions: []llms.FunctionDefinition{{Name: TerminalToolName}},
		handlers:    map[string]ExecutorHandler{},
	}

	fte.appendMCPTools(ce, database.MsgchainTypeCoder)

	if len(ce.definitions) != 3 {
		t.Fatalf("definitions = %d, want 3 (built-in + 2 MCP)", len(ce.definitions))
	}
	if ce.definitions[1].Name != "mcp_burp_scan" || ce.definitions[2].Name != "mcp_burp_spider" {
		t.Errorf("unexpected MCP definitions: %q, %q", ce.definitions[1].Name, ce.definitions[2].Name)
	}

	for _, name := range []string{"mcp_burp_scan", "mcp_burp_spider"} {
		if _, ok := ce.handlers[name]; !ok {
			t.Errorf("handler for %q not registered", name)
		}
	}

	if GetToolType("mcp_burp_scan") != EnvironmentToolType {
		t.Error("MCP tool must map to EnvironmentToolType")
	}
}

func TestAppendMCPToolsFiltersByContext(t *testing.T) {
	t.Parallel()

	discovered := []mcp.Tool{
		{
			Name:     "mcp_burp_scan",
			Server:   "burp",
			Contexts: []string{string(database.MsgchainTypeCoder)},
		},
	}

	fte := &flowToolsExecutor{}
	fte.SetMCPClient(&fakeMCPClient{enabled: true, tools: discovered})

	ce := &customExecutor{
		definitions: []llms.FunctionDefinition{},
		handlers:    map[string]ExecutorHandler{},
	}

	fte.appendMCPTools(ce, database.MsgchainTypeSearcher)

	if len(ce.definitions) != 0 {
		t.Errorf("definitions = %d, want 0 for non-matching context", len(ce.definitions))
	}
}

func TestAppendMCPToolsSkipsDisabledClient(t *testing.T) {
	t.Parallel()

	fte := &flowToolsExecutor{}
	fte.SetMCPClient(&fakeMCPClient{enabled: false, tools: []mcp.Tool{{Name: "mcp_burp_scan", Server: "burp"}}})

	ce := &customExecutor{
		definitions: []llms.FunctionDefinition{},
		handlers:    map[string]ExecutorHandler{},
	}

	fte.appendMCPTools(ce, database.MsgchainTypeCoder)

	if len(ce.definitions) != 0 {
		t.Errorf("definitions = %d, want 0 for disabled client", len(ce.definitions))
	}
	if len(ce.handlers) != 0 {
		t.Errorf("handlers = %d, want 0 for disabled client", len(ce.handlers))
	}
}
