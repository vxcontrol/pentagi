package terminal

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsMarkdownContent_Headers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"h1 prefix", "# Title", true},
		{"h2 prefix", "## Subtitle", true},
		{"h3 in body", "some text\n# Header", true},
		{"plain text", "just some regular text", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsMarkdownContent(tt.input))
		})
	}
}

func TestIsMarkdownContent_CodeBlocks(t *testing.T) {
	assert.True(t, IsMarkdownContent("```go\nfmt.Println()\n```"))
}

func TestIsMarkdownContent_Bold(t *testing.T) {
	assert.True(t, IsMarkdownContent("this is **bold** text"))
}

func TestIsMarkdownContent_Links(t *testing.T) {
	assert.True(t, IsMarkdownContent("click [here](https://example.com)"))
}

func TestIsMarkdownContent_Lists(t *testing.T) {
	assert.True(t, IsMarkdownContent("items:\n- first\n- second"))
}

func TestIsMarkdownContent_PlainText(t *testing.T) {
	assert.False(t, IsMarkdownContent("no markdown here at all"))
	assert.False(t, IsMarkdownContent("single line"))
}

func TestInteractivePromptContext_ReadsInput(t *testing.T) {
	reader := strings.NewReader("hello world\n")
	ctx := context.Background()

	result, err := InteractivePromptContext(ctx, "Enter", reader)
	require.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestInteractivePromptContext_TrimsWhitespace(t *testing.T) {
	reader := strings.NewReader("  trimmed  \n")
	ctx := context.Background()

	result, err := InteractivePromptContext(ctx, "Enter", reader)
	require.NoError(t, err)
	assert.Equal(t, "trimmed", result)
}

func TestInteractivePromptContext_CancelledContext(t *testing.T) {
	pr, pw := io.Pipe()
	defer pw.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := InteractivePromptContext(ctx, "Enter", pr)
	require.ErrorIs(t, err, context.Canceled)
}

func TestGetYesNoInputContext_Yes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"lowercase y", "y\n"},
		{"lowercase yes", "yes\n"},
		{"uppercase Y", "Y\n"},
		{"uppercase YES", "YES\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result, err := GetYesNoInputContext(context.Background(), "Confirm?", reader)
			require.NoError(t, err)
			assert.True(t, result)
		})
	}
}

func TestGetYesNoInputContext_No(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"lowercase n", "n\n"},
		{"lowercase no", "no\n"},
		{"uppercase N", "N\n"},
		{"uppercase NO", "NO\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result, err := GetYesNoInputContext(context.Background(), "Confirm?", reader)
			require.NoError(t, err)
			assert.False(t, result)
		})
	}
}

func TestGetYesNoInputContext_CancelledContext(t *testing.T) {
	pr, pw := io.Pipe()
	defer pw.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := GetYesNoInputContext(ctx, "Confirm?", pr)
	require.ErrorIs(t, err, context.Canceled)
}

func TestPrintJSON_ValidData(t *testing.T) {
	// PrintJSON writes to stdout; verify it doesn't panic
	data := map[string]string{"key": "value"}
	assert.NotPanics(t, func() {
		PrintJSON(data)
	})
}

func TestPrintJSON_InvalidData(t *testing.T) {
	// Channels cannot be marshaled to JSON
	assert.NotPanics(t, func() {
		PrintJSON(make(chan int))
	})
}

func TestRenderMarkdown_Empty(t *testing.T) {
	// Should return without error on empty input
	assert.NotPanics(t, func() {
		RenderMarkdown("")
	})
}

func TestRenderMarkdown_ValidContent(t *testing.T) {
	assert.NotPanics(t, func() {
		RenderMarkdown("# Hello\n\nThis is **bold**")
	})
}

func TestPrintResult_PlainText(t *testing.T) {
	assert.NotPanics(t, func() {
		PrintResult("plain text output")
	})
}

func TestPrintResult_MarkdownContent(t *testing.T) {
	assert.NotPanics(t, func() {
		PrintResult("# Header\n\nSome **bold** text")
	})
}

func TestPrintResultWithKey(t *testing.T) {
	assert.NotPanics(t, func() {
		PrintResultWithKey("Result", "plain text output")
	})
}
