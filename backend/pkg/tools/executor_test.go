package tools

import (
	"encoding/json"
	"strings"
	"testing"
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
