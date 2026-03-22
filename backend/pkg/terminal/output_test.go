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

func TestIsMarkdownContent_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"single bracket", "[", false},
		{"incomplete link", "[text", false},
		{"star without pair", "this has * one star", false},
		{"backtick without triple", "single ` backtick", false},
		{"hyphen without list", "text - not a list", false},
		{"complete link", "[link](url)", true},
		{"double star pair", "text **bold** text", true},
		{"triple backticks", "```code```", true},
		{"proper list", "item\n- list item", true},
		{"multiple markdown features", "# Title\n\n**bold** and [link](url)", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsMarkdownContent(tt.input))
		})
	}
}

func TestInteractivePromptContext_ReadsInput(t *testing.T) {
	reader := strings.NewReader("hello world\n")

	result, err := InteractivePromptContext(t.Context(), "Enter", reader)
	require.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestInteractivePromptContext_TrimsWhitespace(t *testing.T) {
	reader := strings.NewReader("  trimmed  \n")

	result, err := InteractivePromptContext(t.Context(), "Enter", reader)
	require.NoError(t, err)
	assert.Equal(t, "trimmed", result)
}

func TestInteractivePromptContext_CancelledContext(t *testing.T) {
	pr, pw := io.Pipe()
	defer pw.Close()

	ctx, cancel := context.WithCancel(t.Context())
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
			result, err := GetYesNoInputContext(t.Context(), "Confirm?", reader)
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
			result, err := GetYesNoInputContext(t.Context(), "Confirm?", reader)
			require.NoError(t, err)
			assert.False(t, result)
		})
	}
}

func TestGetYesNoInputContext_CancelledContext(t *testing.T) {
	pr, pw := io.Pipe()
	defer pw.Close()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := GetYesNoInputContext(ctx, "Confirm?", pr)
	require.ErrorIs(t, err, context.Canceled)
}

func TestGetYesNoInputContext_InvalidInput(t *testing.T) {
	// Test with invalid input followed by EOF
	reader := strings.NewReader("invalid\n")

	_, err := GetYesNoInputContext(t.Context(), "Confirm?", reader)
	require.Error(t, err)
	assert.ErrorIs(t, err, io.EOF)
}

func TestGetYesNoInputContext_EOFError(t *testing.T) {
	reader := strings.NewReader("") // EOF immediately

	_, err := GetYesNoInputContext(t.Context(), "Confirm?", reader)
	require.Error(t, err)
	assert.ErrorIs(t, err, io.EOF)
}

func TestInteractivePromptContext_EOFError(t *testing.T) {
	reader := strings.NewReader("") // EOF immediately

	_, err := InteractivePromptContext(t.Context(), "Enter", reader)
	require.Error(t, err)
	assert.ErrorIs(t, err, io.EOF)
}

func TestInteractivePromptContext_EmptyInput(t *testing.T) {
	reader := strings.NewReader("\n") // Just newline

	result, err := InteractivePromptContext(t.Context(), "Enter", reader)
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestPrintJSON_ValidData(t *testing.T) {
	data := map[string]string{"key": "value"}
	assert.NotPanics(t, func() {
		PrintJSON(data)
	})
}

func TestPrintJSON_InvalidData(t *testing.T) {
	assert.NotPanics(t, func() {
		PrintJSON(make(chan int))
	})
}

func TestPrintJSON_ComplexData(t *testing.T) {
	data := map[string]interface{}{
		"string": "value",
		"number": 42,
		"bool":   true,
		"array":  []string{"a", "b", "c"},
		"nested": map[string]int{"x": 1, "y": 2},
	}

	assert.NotPanics(t, func() {
		PrintJSON(data)
	})
}

func TestPrintJSON_NilData(t *testing.T) {
	assert.NotPanics(t, func() {
		PrintJSON(nil)
	})
}

func TestRenderMarkdown_Empty(t *testing.T) {
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

func TestPrintResultWithKey_PlainText(t *testing.T) {
	// PrintResultWithKey uses colored output for key, which goes to stderr
	assert.NotPanics(t, func() {
		PrintResultWithKey("Result", "plain text output")
	})
}

func TestPrintResultWithKey_MarkdownContent(t *testing.T) {
	// PrintResultWithKey uses colored output for key, which goes to stderr
	assert.NotPanics(t, func() {
		PrintResultWithKey("Analysis", "# Findings\n\n- **Critical**: Issue found")
	})
}

func TestColoredOutputFunctions_DoNotPanic(t *testing.T) {
	// Color output functions use fatih/color which writes to a custom output
	// that may behave differently in test environments. We verify they don't panic.
	tests := []struct {
		name string
		fn   func(string, ...interface{})
	}{
		{"Info", Info},
		{"Success", Success},
		{"Error", Error},
		{"Warning", Warning},
		{"PrintInfo", PrintInfo},
		{"PrintSuccess", PrintSuccess},
		{"PrintError", PrintError},
		{"PrintWarning", PrintWarning},
		{"PrintMock", PrintMock},
		{"PrintHeader", func(s string, _ ...interface{}) { PrintHeader(s) }},
		{"PrintValueFormat", PrintValueFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				tt.fn("test message")
			})
		})
	}
}

func TestPrintKeyValue_DoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		PrintKeyValue("Name", "PentAGI")
	})
}

func TestPrintKeyValueFormat_DoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		PrintKeyValueFormat("Score", "%d%%", 95)
	})
}

func TestPrintSeparators_DoNotPanic(t *testing.T) {
	t.Run("thin separator", func(t *testing.T) {
		assert.NotPanics(t, func() {
			PrintThinSeparator()
		})
	})

	t.Run("thick separator", func(t *testing.T) {
		assert.NotPanics(t, func() {
			PrintThickSeparator()
		})
	})
}
