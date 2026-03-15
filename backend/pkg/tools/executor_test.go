package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

func TestGetMessage(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{}

	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "valid message field",
			args: `{"message": "hello world", "other": "data"}`,
			want: "hello world",
		},
		{
			name: "empty message",
			args: `{"message": ""}`,
			want: "",
		},
		{
			name: "missing message field",
			args: `{"other": "data"}`,
			want: "",
		},
		{
			name: "invalid json",
			args: `{invalid}`,
			want: "",
		},
		{
			name: "empty json object",
			args: `{}`,
			want: "",
		},
		{
			name: "message with unicode",
			args: `{"message": "testing: \u0041\u0042\u0043"}`,
			want: "testing: ABC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ce.getMessage(json.RawMessage(tt.args))
			if got != tt.want {
				t.Errorf("getMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestArgsToMarkdown(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{}

	tests := []struct {
		name    string
		args    string
		wantErr bool
		check   func(t *testing.T, result string)
	}{
		{
			name: "single field",
			args: `{"query": "test search"}`,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "* query: test search") {
					t.Errorf("expected query bullet, got: %s", result)
				}
			},
		},
		{
			name: "message field skipped",
			args: `{"query": "test", "message": "should be skipped"}`,
			check: func(t *testing.T, result string) {
				if strings.Contains(result, "message") {
					t.Error("message field should be skipped")
				}
				if !strings.Contains(result, "* query: test") {
					t.Errorf("expected query bullet, got: %s", result)
				}
			},
		},
		{
			name: "only message field",
			args: `{"message": "only message"}`,
			check: func(t *testing.T, result string) {
				if result != "" {
					t.Errorf("expected empty string when only message field, got: %q", result)
				}
			},
		},
		{
			name:    "invalid json",
			args:    `{invalid}`,
			wantErr: true,
		},
		{
			name: "empty json object",
			args: `{}`,
			check: func(t *testing.T, result string) {
				if result != "" {
					t.Errorf("expected empty result for empty args, got: %q", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ce.argsToMarkdown(json.RawMessage(tt.args))
			if (err != nil) != tt.wantErr {
				t.Errorf("argsToMarkdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestIsBarrierFunction(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{
		barriers: map[string]struct{}{
			FinalyToolName:  {},
			AskUserToolName: {},
		},
	}

	tests := []struct {
		name     string
		toolName string
		want     bool
	}{
		{name: "done is barrier", toolName: FinalyToolName, want: true},
		{name: "ask is barrier", toolName: AskUserToolName, want: true},
		{name: "terminal is not barrier", toolName: TerminalToolName, want: false},
		{name: "empty string is not barrier", toolName: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ce.IsBarrierFunction(tt.toolName); got != tt.want {
				t.Errorf("IsBarrierFunction(%q) = %v, want %v", tt.toolName, got, tt.want)
			}
		})
	}
}

func TestGetBarrierToolNames(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{
		barriers: map[string]struct{}{
			FinalyToolName:  {},
			AskUserToolName: {},
		},
	}

	names := ce.GetBarrierToolNames()
	if len(names) != 2 {
		t.Fatalf("GetBarrierToolNames() returned %d names, want 2", len(names))
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	if !nameSet[FinalyToolName] {
		t.Errorf("GetBarrierToolNames() missing %q", FinalyToolName)
	}
	if !nameSet[AskUserToolName] {
		t.Errorf("GetBarrierToolNames() missing %q", AskUserToolName)
	}
}

func TestToolsReturnsDefinitions(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{
		definitions: []llms.FunctionDefinition{
			{Name: TerminalToolName, Description: "terminal"},
			{Name: FileToolName, Description: "file"},
		},
	}

	tools := ce.Tools()
	if len(tools) != 2 {
		t.Fatalf("Tools() returned %d tools, want 2", len(tools))
	}
	if tools[0].Function == nil || tools[0].Function.Name != TerminalToolName {
		t.Fatalf("Tools()[0] mismatch: %+v", tools[0].Function)
	}
	if tools[1].Function == nil || tools[1].Function.Name != FileToolName {
		t.Fatalf("Tools()[1] mismatch: %+v", tools[1].Function)
	}
}

func TestExecuteEarlyReturns(t *testing.T) {
	t.Parallel()

	t.Run("unknown tool returns helper message", func(t *testing.T) {
		t.Parallel()
		ce := &customExecutor{
			handlers: map[string]ExecutorHandler{},
		}
		result, err := ce.Execute(t.Context(), 1, "id", "unknown_tool", "", "", json.RawMessage(`{}`))
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if !strings.Contains(result, "function 'unknown_tool' not found") {
			t.Fatalf("Execute() result = %q, expected not found message", result)
		}
	})

	t.Run("invalid args json returns fix message", func(t *testing.T) {
		t.Parallel()
		ce := &customExecutor{
			handlers: map[string]ExecutorHandler{
				TerminalToolName: func(ctx context.Context, name string, args json.RawMessage) (string, error) {
					return "ok", nil
				},
			},
		}
		result, err := ce.Execute(t.Context(), 1, "id", TerminalToolName, "", "", json.RawMessage(`{invalid`))
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if !strings.Contains(result, "failed to unmarshal") || !strings.Contains(result, "fix it") {
			t.Fatalf("Execute() result = %q, expected argument-fix message", result)
		}
	})
}

func TestGetToolSchemaFallbackAndUnknown(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{
		definitions: []llms.FunctionDefinition{
			registryDefinitions[TerminalToolName],
		},
	}

	schemaObj, err := ce.GetToolSchema(TerminalToolName)
	if err != nil {
		t.Fatalf("GetToolSchema(%q) unexpected error: %v", TerminalToolName, err)
	}
	if schemaObj == nil {
		t.Fatalf("GetToolSchema(%q) returned nil schema", TerminalToolName)
	}

	// Should fallback to global registry definitions when not in ce.definitions
	schemaObj, err = ce.GetToolSchema(BrowserToolName)
	if err != nil {
		t.Fatalf("GetToolSchema(%q) fallback unexpected error: %v", BrowserToolName, err)
	}
	if schemaObj == nil {
		t.Fatalf("GetToolSchema(%q) fallback returned nil schema", BrowserToolName)
	}

	_, err = ce.GetToolSchema("unknown_tool")
	if err == nil {
		t.Fatal("GetToolSchema(unknown_tool) should return error")
	}
}

func TestGetBarrierToolsSkipsUnknownBarriers(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{
		barriers: map[string]struct{}{
			FinalyToolName: {},
			"unknown_tool": {},
		},
	}

	tools := ce.GetBarrierTools()
	if len(tools) != 1 {
		t.Fatalf("GetBarrierTools() returned %d tools, want 1", len(tools))
	}
	if tools[0].Name != FinalyToolName {
		t.Fatalf("GetBarrierTools()[0].Name = %q, want %q", tools[0].Name, FinalyToolName)
	}
	if tools[0].Schema == "" {
		t.Fatal("GetBarrierTools()[0].Schema should not be empty")
	}
}

func TestGetSummarizePromptTruncatesLongArgValues(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{
		definitions: []llms.FunctionDefinition{
			registryDefinitions[TerminalToolName],
		},
	}

	longValue := strings.Repeat("x", maxArgValueLength+50)
	args := map[string]any{
		"message": "hello",
		"query":   longValue,
	}
	argsBytes, err := json.Marshal(args)
	if err != nil {
		t.Fatalf("failed to marshal args: %v", err)
	}

	prompt, err := ce.getSummarizePrompt(TerminalToolName, string(argsBytes), "result")
	if err != nil {
		t.Fatalf("getSummarizePrompt() unexpected error: %v", err)
	}
	if !strings.Contains(prompt, "... [truncated]") {
		t.Fatalf("prompt should contain truncated marker, got: %q", prompt)
	}
}

func TestConverToJSONSchemaErrorPath(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{}
	_, err := ce.converToJSONSchema(make(chan int))
	if err == nil {
		t.Fatal("converToJSONSchema() should fail on non-marshalable type")
	}
	if !strings.Contains(err.Error(), "failed to marshal parameters") {
		t.Fatalf("unexpected error: %v", err)
	}
}
