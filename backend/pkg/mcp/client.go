// Package mcp provides a Model Context Protocol client that connects
// PentAGI to external MCP servers and exposes their tools to LLM agents as
// first-class tools alongside the built-in ones.
//
// The client is configured through the MCP_SERVERS environment variable
// (a JSON array, see config.MCPServerConfig). Servers are contacted once at
// client creation; a server that cannot be reached is logged and skipped so
// a broken integration degrades to "tool not available" instead of taking
// the whole application down.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"pentagi/pkg/config"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// clientImplementation identifies PentAGI to remote MCP servers.
var clientImplementation = &gomcp.Implementation{
	Name:    "pentagi",
	Version: "0.1.0",
}

// Tool is the provider-agnostic description of one MCP tool exposed to
// agents. InputSchema is the raw JSON Schema announced by the server.
type Tool struct {
	Name        string
	Description string
	InputSchema json.RawMessage
	// Server is the configured server name the tool belongs to.
	Server string
	// Contexts restricts the agent contexts the tool is offered to; empty
	// means "every tool-capable agent".
	Contexts []string
}

// toolDescriptor caches everything needed to call a tool after discovery,
// so serving a call never re-queries the server.
type toolDescriptor struct {
	localName   string
	description string
	inputSchema json.RawMessage
}

// serverConn couples one connected MCP session with its configuration and
// discovered tool descriptors keyed by namespaced tool name.
type serverConn struct {
	config  config.MCPServerConfig
	tools   map[string]*toolDescriptor
	session *gomcp.ClientSession
}

// Client manages all configured MCP server connections and dispatches tool
// calls by namespaced tool name. It is safe for concurrent use.
type Client struct {
	mu      sync.RWMutex
	timeout time.Duration
	// servers maps configured server name -> live connection.
	servers map[string]*serverConn
	// tools maps namespaced tool name -> owning server name.
	tools map[string]string
}

// NewClient connects to every configured server and collects its tools.
//
// Connection and discovery failures are logged as warnings and the affected
// server is skipped: an unreachable MCP server must never prevent PentAGI
// from starting. The returned client reports IsEnabled()==false when no
// tool could be registered at all.
func NewClient(ctx context.Context, cfg *config.Config) *Client {
	timeout := time.Duration(cfg.MCPToolTimeout) * time.Second
	if cfg.MCPToolTimeout < 0 {
		timeout = 0
	}

	client := &Client{
		timeout: timeout,
		servers: make(map[string]*serverConn, len(cfg.MCPServers)),
		tools:   make(map[string]string),
	}

	for _, serverCfg := range cfg.MCPServers {
		client.connectServer(ctx, serverCfg)
	}

	client.filterTools(newToolSelector(cfg.MCPAllowedTools, cfg.MCPDeniedTools))

	return client
}

// connectServer establishes one server connection, lists its tools and
// registers them under namespaced names. Any error is logged and the server
// is left out — the rest of the fleet keeps working.
func (c *Client) connectServer(ctx context.Context, serverCfg config.MCPServerConfig) {
	log := logrus.WithField("mcp_server", serverCfg.Name)

	transport, err := transportFactory(serverCfg)
	if err != nil {
		log.WithError(err).Warn("failed to configure MCP server transport, skipping")
		return
	}

	client := gomcp.NewClient(clientImplementation, nil)
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.WithError(err).Warn("failed to connect to MCP server, skipping")
		return
	}

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		log.WithError(err).Warn("failed to list MCP server tools, skipping")
		session.Close()
		return
	}

	conn := &serverConn{
		config:  serverCfg,
		tools:   make(map[string]*toolDescriptor, len(result.Tools)),
		session: session,
	}

	for _, tool := range result.Tools {
		if tool == nil {
			continue
		}
		name := ToolName(serverCfg.Name, tool.Name)
		if name == "" {
			log.WithField("mcp_tool", tool.Name).Warn("MCP tool name collapses to an empty function name, skipping")
			continue
		}
		if owner, dup := c.tools[name]; dup {
			logrus.WithFields(logrus.Fields{
				"mcp_server": serverCfg.Name,
				"mcp_tool":   tool.Name,
				"owner":      owner,
			}).Warn("namespaced MCP tool name is already taken, skipping duplicate")
			continue
		}

		conn.tools[name] = &toolDescriptor{
			localName:   tool.Name,
			description: tool.Description,
			inputSchema: marshalSchema(tool.InputSchema),
		}
		c.tools[name] = serverCfg.Name
	}

	if len(conn.tools) == 0 {
		log.Warn("MCP server exposes no tools, disconnecting")
		session.Close()
		return
	}

	c.servers[serverCfg.Name] = conn

	log.WithFields(logrus.Fields{
		"transport":   serverCfg.Transport,
		"tools_total": len(result.Tools),
		"tools_added": len(conn.tools),
	}).Info("connected to MCP server")
}

// filterTools removes tools that the allow/deny selectors reject and drops
// servers left without a single tool — they would hold a connection for
// nothing.
func (c *Client) filterTools(selector *toolSelector) {
	if selector.disabled() {
		return
	}

	for name, server := range c.tools {
		if !selector.allows(server, localToolName(server, name)) {
			logrus.WithFields(logrus.Fields{
				"mcp_server": server,
				"mcp_tool":   name,
			}).Info("MCP tool filtered out by allow/deny lists")
			delete(c.tools, name)
		}
	}

	for serverName, conn := range c.servers {
		if !c.serverHasTools(serverName) {
			logrus.WithField("mcp_server", serverName).Info("MCP server has no tools left after filtering, disconnecting")
			conn.session.Close()
			delete(c.servers, serverName)
		}
	}
}

func (c *Client) serverHasTools(serverName string) bool {
	for _, owner := range c.tools {
		if owner == serverName {
			return true
		}
	}
	return false
}

// transportFactory builds the transport for a server configuration. It is
// a package-level variable so tests can substitute in-memory transports.
var transportFactory = transportFor

// transportFor builds the SDK transport matching the server configuration.
func transportFor(serverCfg config.MCPServerConfig) (gomcp.Transport, error) {
	switch serverCfg.Transport {
	case "stdio":
		cmd := exec.Command(serverCfg.Command, serverCfg.Args...)
		cmd.Env = os.Environ()
		for key, value := range serverCfg.Env {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
		return &gomcp.CommandTransport{Command: cmd}, nil
	case "http":
		return &gomcp.StreamableClientTransport{Endpoint: serverCfg.URL}, nil
	case "sse":
		return &gomcp.SSEClientTransport{Endpoint: serverCfg.URL}, nil
	default:
		return nil, fmt.Errorf("unsupported transport %q", serverCfg.Transport)
	}
}

// Close terminates every server connection. Safe to call multiple times.
func (c *Client) Close() {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for name, conn := range c.servers {
		if err := conn.session.Close(); err != nil {
			logrus.WithError(err).WithField("mcp_server", name).Warn("failed to close MCP server connection")
		}
		delete(c.servers, name)
	}
	c.tools = make(map[string]string)
}

// IsEnabled reports whether at least one MCP tool is available.
func (c *Client) IsEnabled() bool {
	return c != nil && len(c.tools) > 0
}

// Tools returns the MCP tools visible to the given agent context (for the
// context vocabulary see config.ValidMCPAgentContexts). The result is sorted
// by tool name for deterministic tool listings.
func (c *Client) Tools(agentContext string) []Tool {
	if c == nil {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	tools := make([]Tool, 0, len(c.tools))
	for name, serverName := range c.tools {
		conn := c.servers[serverName]
		if conn == nil {
			continue
		}
		if len(conn.config.Contexts) > 0 && !containsString(conn.config.Contexts, agentContext) {
			continue
		}
		descriptor := conn.tools[name]
		if descriptor == nil {
			continue
		}
		tools = append(tools, Tool{
			Name:        name,
			Description: descriptor.description,
			InputSchema: descriptor.inputSchema,
			Server:      serverName,
			Contexts:    conn.config.Contexts,
		})
	}

	sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })
	return tools
}

// CallTool invokes an MCP tool by its namespaced name. Arguments are passed
// through as a raw JSON object; the result is rendered as the tool output
// string the agent sees.
func (c *Client) CallTool(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if c == nil {
		return "", fmt.Errorf("MCP integration is not enabled")
	}

	c.mu.RLock()
	serverName, registered := c.tools[name]
	var conn *serverConn
	var descriptor *toolDescriptor
	if registered {
		conn = c.servers[serverName]
		if conn != nil {
			descriptor = conn.tools[name]
		}
	}
	c.mu.RUnlock()

	if conn == nil || descriptor == nil {
		return "", fmt.Errorf("MCP tool '%s' is not available", name)
	}

	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	var arguments map[string]any
	if len(args) > 0 {
		if err := json.Unmarshal(args, &arguments); err != nil {
			return "", fmt.Errorf("failed to unmarshal '%s' tool arguments: %w", name, err)
		}
	}

	result, err := conn.session.CallTool(ctx, &gomcp.CallToolParams{
		Name:      descriptor.localName,
		Arguments: arguments,
	})
	if err != nil {
		return "", fmt.Errorf("MCP tool '%s' call failed: %w", name, err)
	}

	output := renderResult(result)
	if result.IsError {
		// A tool-level error means the server rejected the invocation
		// (bad arguments, out-of-scope guard, upstream refused). That is
		// data for the model, not an infrastructure failure: return the
		// payload as the tool observation so the agent reads it and
		// corrects course. Surfacing it as a Go error makes the executor
		// retry the identical call and then abort the whole chain.
		return fmt.Sprintf("TOOL ERROR: %s", truncateForError(output)), nil
	}
	return output, nil
}

// renderResult converts a CallToolResult into the string returned to the
// agent. Text content is concatenated; other content blocks are rendered as
// JSON so no information is silently dropped.
func renderResult(result *gomcp.CallToolResult) string {
	var builder strings.Builder
	for _, content := range result.Content {
		if text, ok := content.(*gomcp.TextContent); ok {
			if builder.Len() > 0 {
				builder.WriteByte('\n')
			}
			builder.WriteString(text.Text)
			continue
		}
		encoded, err := json.Marshal(content)
		if err == nil {
			if builder.Len() > 0 {
				builder.WriteByte('\n')
			}
			builder.Write(encoded)
		}
	}

	if builder.Len() == 0 && result.StructuredContent != nil {
		encoded, err := json.Marshal(result.StructuredContent)
		if err == nil {
			return string(encoded)
		}
	}
	return builder.String()
}

// maxToolErrorOutputBytes caps how much of a failed tool payload is inlined
// into the error string; customExecutor persists the full result, but the
// error also feeds prompts that must stay bounded.
const maxToolErrorOutputBytes = 1024

func truncateForError(output string) string {
	if len(output) <= maxToolErrorOutputBytes {
		return output
	}
	return output[:maxToolErrorOutputBytes] + "... [truncated]"
}

// marshalSchema converts the server-announced input schema into raw JSON.
// The SDK decodes inputSchema into a generic map on the client side, so
// re-marshalling restores the schema document verbatim.
func marshalSchema(inputSchema any) json.RawMessage {
	if inputSchema == nil {
		return nil
	}
	if raw, ok := inputSchema.(json.RawMessage); ok {
		return raw
	}
	encoded, err := json.Marshal(inputSchema)
	if err != nil {
		return nil
	}
	return encoded
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
