package tools

import (
	"fmt"
	"strings"
	"testing"
)

func TestPrimaryTerminalName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		flowID int64
		want   string
	}{
		{1, "pentagi-terminal-1"},
		{0, "pentagi-terminal-0"},
		{12345, "pentagi-terminal-12345"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("flowID=%d", tt.flowID), func(t *testing.T) {
			t.Parallel()

			if got := PrimaryTerminalName(tt.flowID); got != tt.want {
				t.Errorf("PrimaryTerminalName(%d) = %q, want %q", tt.flowID, got, tt.want)
			}
		})
	}
}

func TestFormatTerminalInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cwd  string
		text string
	}{
		{name: "basic command", cwd: "/home/user", text: "ls -la"},
		{name: "empty cwd", cwd: "", text: "pwd"},
		{name: "complex command", cwd: "/tmp", text: "find . -name '*.go'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FormatTerminalInput(tt.cwd, tt.text)
			if tt.cwd != "" && !strings.Contains(result, tt.cwd) {
				t.Errorf("result should contain cwd %q", tt.cwd)
			}
			if tt.cwd == "" && !strings.HasPrefix(result, " $ ") {
				t.Errorf("empty cwd should produce prompt prefix ' $ ', got %q", result)
			}
			if !strings.Contains(result, tt.text) {
				t.Errorf("result should contain text %q", tt.text)
			}
			if !strings.HasSuffix(result, "\r\n") {
				t.Error("result should end with CRLF")
			}
			// Should contain ANSI yellow escape code
			if !strings.Contains(result, "\033[33m") {
				t.Error("result should contain yellow ANSI code")
			}
			if !strings.Contains(result, "\033[0m") {
				t.Error("result should contain reset ANSI code")
			}
		})
	}
}

func TestFormatTerminalSystemOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
	}{
		{name: "simple output", text: "file written successfully"},
		{name: "empty output", text: ""},
		{name: "multiline output", text: "line 1\nline 2\nline 3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := FormatTerminalSystemOutput(tt.text)
			if tt.text != "" && !strings.Contains(result, tt.text) {
				t.Errorf("result should contain text %q", tt.text)
			}
			if !strings.HasSuffix(result, "\r\n") {
				t.Error("result should end with CRLF")
			}
			// Should contain ANSI blue escape code
			if !strings.Contains(result, "\033[34m") {
				t.Error("result should contain blue ANSI code")
			}
			if !strings.Contains(result, "\033[0m") {
				t.Error("result should contain reset ANSI code")
			}
			if tt.text == "" {
				expected := "\033[34m\033[0m\r\n"
				if result != expected {
					t.Errorf("empty output formatting mismatch: got %q, want %q", result, expected)
				}
			}
		})
	}
}
