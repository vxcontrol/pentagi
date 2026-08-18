package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"pentagi/pkg/config"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolName(t *testing.T) {
	tests := []struct {
		name   string
		server string
		tool   string
		want   string
	}{
		{"simple", "burp", "scan", "mcp_burp_scan"},
		{"hyphens", "my-server", "active-scan", "mcp_my_server_active_scan"},
		{"uppercase", "Nuclei", "RunTemplate", "mcp_nuclei_runtemplate"},
		{"spaces and dots", "shodan api", "host.search", "mcp_shodan_api_host_search"},
		{"server-only sanity", "burp", "", "mcp_burp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ToolName(tt.server, tt.tool))
			assert.LessOrEqual(t, len(ToolName(tt.server, tt.tool)), maxToolNameBytes)
		})
	}
}

func TestToolNameTruncation(t *testing.T) {
	server := strings.Repeat("s", 40)
	tool := strings.Repeat("t", 60)

	name := ToolName(server, tool)

	require.NotEmpty(t, name)
	assert.LessOrEqual(t, len(name), maxToolNameBytes)
}

func TestLocalToolNameInvertsToolName(t *testing.T) {
	server := "Burp-Suite"
	tool := "active_scan"

	name := ToolName(server, tool)
	require.Equal(t, "mcp_burp_suite_active_scan", name)
	assert.Equal(t, tool, localToolName(server, name))
}

func TestSanitizeName(t *testing.T) {
	assert.Equal(t, "a_b_c", sanitizeName("a b//c"))
	assert.Equal(t, "", sanitizeName("---"))
	assert.Equal(t, "abc123", sanitizeName("abc123"))
}

func TestSelector(t *testing.T) {
	tests := []struct {
		name    string
		allowed []string
		denied  []string
		server  string
		tool    string
		want    bool
	}{
		{"no filters", nil, nil, "burp", "scan", true},
		{"allow whole server", []string{"burp"}, nil, "burp", "scan", true},
		{"allow other server", []string{"nuclei"}, nil, "burp", "scan", false},
		{"allow one tool", []string{"burp/scan"}, nil, "burp", "scan", true},
		{"allow one tool, other tool", []string{"burp/scan"}, nil, "burp", "spider", false},
		{"allow star", []string{"*"}, nil, "burp", "scan", true},
		{"allow server wildcard", []string{"burp/*"}, nil, "burp", "spider", true},
		{"allow server wildcard, other server", []string{"burp/*"}, nil, "nuclei", "scan", false},
		{"deny whole server", nil, []string{"burp"}, "burp", "scan", false},
		{"deny overrides allow star", []string{"*"}, []string{"burp"}, "burp", "scan", false},
		{"deny one tool overrides server allow", []string{"burp"}, []string{"burp/spider"}, "burp", "spider", false},
		{"deny one tool keeps siblings", []string{"burp"}, []string{"burp/spider"}, "burp", "scan", true},
		{"case and separator forgiving", []string{"Burp Suite"}, nil, "burp-suite", "scan", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := newToolSelector(tt.allowed, tt.denied)
			assert.Equal(t, tt.want, selector.allows(tt.server, tt.tool))
		})
	}
}

// fakeServer starts an in-memory MCP server exposing one echo tool and
// returns the client-side transport for it.
func fakeServer(t *testing.T, name string) gomcp.Transport {
	t.Helper()

	server := gomcp.NewServer(&gomcp.Implementation{Name: name}, nil)
	server.AddTool(&gomcp.Tool{
		Name:        "echo",
		Description: "echo the message back",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"message": map[string]any{"type": "string"},
			},
			"required": []string{"message"},
		},
	}, func(ctx context.Context, req *gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		var params struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
			return nil, err
		}
		return &gomcp.CallToolResult{
			Content: []gomcp.Content{&gomcp.TextContent{Text: "echo: " + params.Message}},
		}, nil
	})

	serverTransport, clientTransport := gomcp.NewInMemoryTransports()
	go func() {
		_ = server.Run(context.Background(), serverTransport)
	}()

	// The in-memory transport pair has no explicit Close; the server side
	// unblocks when the client session closes its end of the pipe.
	return clientTransport
}

// withTransportFactory swaps the package transport factory for the test
// duration and restores it afterwards.
func withTransportFactory(t *testing.T, factory func(config.MCPServerConfig) (gomcp.Transport, error)) {
	t.Helper()

	previous := transportFactory
	transportFactory = factory
	t.Cleanup(func() { transportFactory = previous })
}

func mcpConfig(servers ...config.MCPServerConfig) *config.Config {
	return &config.Config{MCPServers: servers, MCPToolTimeout: 5}
}

func TestClientToolsAndCall(t *testing.T) {
	withTransportFactory(t, func(cfg config.MCPServerConfig) (gomcp.Transport, error) {
		return fakeServer(t, cfg.Name), nil
	})

	client := NewClient(context.Background(), mcpConfig(
		config.MCPServerConfig{Name: "fake", Transport: "http", URL: "http://unused"},
	))
	t.Cleanup(client.Close)

	require.True(t, client.IsEnabled())

	tools := client.Tools("agent")
	require.Len(t, tools, 1)
	assert.Equal(t, "mcp_fake_echo", tools[0].Name)
	assert.Equal(t, "echo the message back", tools[0].Description)
	assert.Equal(t, "fake", tools[0].Server)
	require.NotNil(t, tools[0].InputSchema)
	assert.Contains(t, string(tools[0].InputSchema), `"message"`)

	result, err := client.CallTool(context.Background(), "mcp_fake_echo", json.RawMessage(`{"message":"hi"}`))
	require.NoError(t, err)
	assert.Equal(t, "echo: hi", result)
}

func TestClientFiltersByAgentContext(t *testing.T) {
	withTransportFactory(t, func(cfg config.MCPServerConfig) (gomcp.Transport, error) {
		return fakeServer(t, cfg.Name), nil
	})

	client := NewClient(context.Background(), mcpConfig(
		config.MCPServerConfig{Name: "fake", Transport: "http", URL: "http://unused", Contexts: []string{"coder"}},
	))
	t.Cleanup(client.Close)

	assert.Empty(t, client.Tools("agent"))
	assert.Len(t, client.Tools("coder"), 1)
}

func TestClientAppliesAllowDenyLists(t *testing.T) {
	withTransportFactory(t, func(cfg config.MCPServerConfig) (gomcp.Transport, error) {
		return fakeServer(t, cfg.Name), nil
	})

	cfg := mcpConfig(
		config.MCPServerConfig{Name: "burp", Transport: "http", URL: "http://unused"},
		config.MCPServerConfig{Name: "nuclei", Transport: "http", URL: "http://unused"},
	)
	cfg.MCPAllowedTools = []string{"*"}
	cfg.MCPDeniedTools = []string{"nuclei"}

	client := NewClient(context.Background(), cfg)
	t.Cleanup(client.Close)

	var names []string
	for _, tool := range client.Tools("agent") {
		names = append(names, tool.Name)
	}
	assert.Equal(t, []string{"mcp_burp_echo"}, names)

	_, err := client.CallTool(context.Background(), "mcp_nuclei_echo", json.RawMessage(`{}`))
	require.Error(t, err)
}

func TestClientSkipsUnreachableServer(t *testing.T) {
	// A nonexistent stdio command: Connect must fail, the client must
	// survive with the other server still working.
	broken := config.MCPServerConfig{Name: "broken", Transport: "stdio", Command: "definitely-not-a-real-binary-xyz"}
	healthy := config.MCPServerConfig{Name: "healthy", Transport: "http", URL: "http://unused"}

	withTransportFactory(t, func(cfg config.MCPServerConfig) (gomcp.Transport, error) {
		if cfg.Name == "broken" {
			return transportFor(broken)
		}
		return fakeServer(t, cfg.Name), nil
	})

	client := NewClient(context.Background(), mcpConfig(broken, healthy))
	t.Cleanup(client.Close)

	require.True(t, client.IsEnabled())
	assert.Len(t, client.Tools("agent"), 1)

	_, err := client.CallTool(context.Background(), "mcp_broken_echo", json.RawMessage(`{}`))
	require.Error(t, err)
}

func TestClientDisabledByDefault(t *testing.T) {
	client := NewClient(context.Background(), mcpConfig())
	assert.False(t, client.IsEnabled())
	assert.Empty(t, client.Tools("agent"))

	var nilClient *Client
	assert.False(t, nilClient.IsEnabled())
	assert.Nil(t, nilClient.Tools("agent"))

	_, err := nilClient.CallTool(context.Background(), "mcp_any_tool", nil)
	require.Error(t, err)
}

func TestClientToolErrorSurfaced(t *testing.T) {
	server := gomcp.NewServer(&gomcp.Implementation{Name: "failing"}, nil)
	server.AddTool(&gomcp.Tool{
		Name:        "fail",
		Description: "always fails",
		InputSchema: map[string]any{"type": "object"},
	}, func(ctx context.Context, req *gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		return &gomcp.CallToolResult{
			IsError: true,
			Content: []gomcp.Content{&gomcp.TextContent{Text: "target unreachable"}},
		}, nil
	})

	serverTransport, clientTransport := gomcp.NewInMemoryTransports()
	go func() { _ = server.Run(context.Background(), serverTransport) }()

	withTransportFactory(t, func(config.MCPServerConfig) (gomcp.Transport, error) {
		return clientTransport, nil
	})

	client := NewClient(context.Background(), mcpConfig(
		config.MCPServerConfig{Name: "failing", Transport: "http", URL: "http://unused"},
	))
	t.Cleanup(client.Close)

	_, err := client.CallTool(context.Background(), "mcp_failing_fail", json.RawMessage(`{}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "target unreachable")
}
