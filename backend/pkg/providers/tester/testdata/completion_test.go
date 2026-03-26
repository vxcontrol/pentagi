package testdata

import (
	"testing"
	"time"

	"github.com/vxcontrol/langchaingo/llms"
	"gopkg.in/yaml.v3"
)

func TestCompletionTestCase(t *testing.T) {
	testYAML := `
- id: "test_basic"
  name: "Basic Math Test"
  type: "completion"
  group: "basic"
  prompt: "What is 2+2?"
  expected: "4"
  streaming: false

- id: "test_messages"
  name: "System User Test"
  type: "completion"
  group: "basic"
  messages:
    - role: "system"
      content: "You are a math assistant"
    - role: "user"
      content: "Calculate 5 * 10"
  expected: "50"
  streaming: false
`

	var definitions []TestDefinition
	err := yaml.Unmarshal([]byte(testYAML), &definitions)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if len(definitions) != 2 {
		t.Fatalf("Expected 2 definitions, got %d", len(definitions))
	}

	// test basic completion case
	basicDef := definitions[0]
	testCase, err := newCompletionTestCase(basicDef)
	if err != nil {
		t.Fatalf("Failed to create basic test case: %v", err)
	}

	if testCase.ID() != "test_basic" {
		t.Errorf("Expected ID 'test_basic', got %s", testCase.ID())
	}
	if testCase.Type() != TestTypeCompletion {
		t.Errorf("Expected type completion, got %s", testCase.Type())
	}
	if testCase.Prompt() != "What is 2+2?" {
		t.Errorf("Expected prompt 'What is 2+2?', got %s", testCase.Prompt())
	}
	if len(testCase.Messages()) != 0 {
		t.Errorf("Expected no messages for basic test, got %d", len(testCase.Messages()))
	}

	// test execution with correct response
	result := testCase.Execute("The answer is 4", time.Millisecond*100)
	if !result.Success {
		t.Errorf("Expected success for correct response, got failure: %v", result.Error)
	}
	if result.Latency != time.Millisecond*100 {
		t.Errorf("Expected latency 100ms, got %v", result.Latency)
	}

	// test execution with incorrect response
	result = testCase.Execute("The answer is 5", time.Millisecond*50)
	if result.Success {
		t.Errorf("Expected failure for incorrect response, got success")
	}

	// test messages case
	messagesDef := definitions[1]
	testCase, err = newCompletionTestCase(messagesDef)
	if err != nil {
		t.Fatalf("Failed to create messages test case: %v", err)
	}

	if len(testCase.Messages()) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(testCase.Messages()))
	}

	// test with ContentResponse
	response := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: "The result is 50",
			},
		},
	}
	result = testCase.Execute(response, time.Millisecond*200)
	if !result.Success {
		t.Errorf("Expected success for ContentResponse, got failure: %v", result.Error)
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected string
		want     bool
	}{
		// Basic exact matches
		{"exact_match", "4", "4", true},
		{"exact_match_text", "hello world", "hello world", true},
		{"empty_response", "", "4", false},

		// Basic contains matches (no modifiers needed)
		{"contains_simple", "The answer is 4", "4", true},
		{"reverse_contains", "4", "The answer is 4", true},

		// Case normalization tests
		{"case_insensitive", "HELLO WORLD", "hello world", true},
		{"mixed_case", "Hello World", "HELLO world", true},
		{"case_in_sentence", "The Answer Is CORRECT", "answer is correct", true},

		// Whitespace removal tests
		{"whitespace_spaces", "1,2,3,4,5", "1, 2, 3, 4, 5", true},
		{"whitespace_tabs", "hello\tworld", "hello world", true},
		{"whitespace_newlines", "hello\nworld", "hello world", true},
		{"whitespace_mixed", "a\t b\n c\r d", "a b c d", true},
		{"number_sequence_normalized", "1 2 3 4 5", "1,2,3,4,5", true},

		// Markdown removal tests
		{"markdown_bold", "This is **bold** text", "This is bold text", true},
		{"markdown_italic", "This is *italic* text", "This is italic text", true},
		{"markdown_code", "Use `code` here", "Use code here", true},
		{"markdown_headers", "# Header text", "Header text", true},
		{"markdown_links", "[link text](url)", "link text url", true},
		{"markdown_blockquote", "> quoted text", "quoted text", true},
		{"markdown_list", "- item one", "item one", true},
		{"markdown_complex", "**Bold** and *italic* with `code`", "Bold and italic with code", true},

		// Punctuation removal tests
		{"punctuation_basic", "Hello, world!", "Hello world", true},
		{"punctuation_question", "Is this correct?", "Is this correct", true},
		{"punctuation_parentheses", "Text (in brackets)", "Text in brackets", true},
		{"punctuation_mixed", "Hello, world! How are you?", "Hello world How are you", true},

		// Quote removal tests
		{"quotes_double", `He said "hello"`, "He said hello", true},
		{"quotes_single", "It's a 'test'", "Its a test", true},
		{"quotes_smart", "\"Smart quotes\"", "Smart quotes", true},
		{"quotes_backticks", "`quoted text`", "quoted text", true},

		// Number normalization tests
		{"numbers_comma_spaced", "sequence: 1, 2, 3, 4, 5", "1,2,3,4,5", true},
		{"numbers_space_separated", "count 1 2 3 4 5", "1,2,3,4,5", true},
		{"numbers_dash_separated", "range: 1-2-3-4-5", "1,2,3,4,5", true},
		{"numbers_dot_separated", "version 1.2.3.4.5", "1,2,3,4,5", true},

		// Combined modifier tests (multiple modifiers working together)
		{"combined_case_whitespace", "HELLO  WORLD", "hello world", true},
		{"combined_case_punctuation", "HELLO, WORLD!", "hello world", true},
		{"combined_markdown_case", "**BOLD TEXT**", "bold text", true},
		{"combined_all_modifiers", "**HELLO,**  `world`!", "hello world", true},
		{"complex_markdown_case", "> **Important:** Use `this` method!", "Important Use this method", true},

		// Edge cases and challenging scenarios
		{"nested_markdown", "**Bold *and italic* text**", "Bold and italic text", true},
		{"multiple_spaces", "hello    world", "hello world", true},
		{"unicode_quotes", "\"Unicode quotes\"", "Unicode quotes", true},
		{"mixed_punctuation", "Hello... world!!!", "Hello world", true},
		{"code_block", "```\ncode here\n```", "code here", true},

		// Tests that should fail
		{"no_match_different_text", "completely different", "expected text", false},
		{"no_match_numbers", "1,2,3", "4,5,6", false},
		{"no_match_partial", "partial", "completely different text", false},

		// Real-world LLM response scenarios
		{"llm_response_natural", "The answer to your question is: 42", "42", true},
		{"llm_response_formatted", "**Answer:** The result is `50`", "The result is 50", true},
		{"llm_response_list", "Here are the steps:\n- Step 1\n- Step 2", "Step 1 Step 2", true},
		{"llm_response_code", "Use this function: `calculateSum()`", "calculateSum", true},
		{"llm_response_explanation", "The value (approximately 3.14) is correct", "3.14", true},

		// Bidirectional matching tests
		{"bidirectional_short_in_long", "answer", "The answer is 42", true},
		{"bidirectional_long_in_short", "The answer is 42", "answer", true},
		{"bidirectional_with_modifiers", "ANSWER", "the **answer** is correct", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsString(tt.response, tt.expected)
			if got != tt.want {
				t.Errorf("containsString(%q, %q) = %v, want %v", tt.response, tt.expected, got, tt.want)
			}
		})
	}
}

// Test individual modifiers
func TestStringModifiers(t *testing.T) {
	t.Run("normalizeCase", func(t *testing.T) {
		result := normalizeCase("HELLO World")
		expected := "hello world"
		if result != expected {
			t.Errorf("normalizeCase() = %q, want %q", result, expected)
		}
	})

	t.Run("removeWhitespace", func(t *testing.T) {
		result := removeWhitespace("hello \t\n\r world")
		expected := "helloworld"
		if result != expected {
			t.Errorf("removeWhitespace() = %q, want %q", result, expected)
		}
	})

	t.Run("removeMarkdown", func(t *testing.T) {
		result := removeMarkdown("**bold** and *italic* with `code`")
		expected := "bold and italic with code"
		if result != expected {
			t.Errorf("removeMarkdown() = %q, want %q", result, expected)
		}
	})

	t.Run("removePunctuation", func(t *testing.T) {
		result := removePunctuation("Hello, world!")
		expected := "Hello world"
		if result != expected {
			t.Errorf("removePunctuation() = %q, want %q", result, expected)
		}
	})

	t.Run("removeQuotes", func(t *testing.T) {
		result := removeQuotes(`"Hello" and 'world'`)
		expected := "Hello and world"
		if result != expected {
			t.Errorf("removeQuotes() = %q, want %q", result, expected)
		}
	})

	t.Run("normalizeNumbers", func(t *testing.T) {
		result := normalizeNumbers("sequence: 1, 2, 3, 4, 5")
		expected := "sequence: 1,2,3,4,5"
		if result != expected {
			t.Errorf("normalizeNumbers() = %q, want %q", result, expected)
		}
	})
}

// Test modifier combinations
func TestModifierCombinations(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		modifiers []stringModifier
		expected  string
	}{
		{
			name:      "case_and_whitespace",
			input:     "HELLO  WORLD",
			modifiers: []stringModifier{normalizeCase, removeWhitespace},
			expected:  "helloworld",
		},
		{
			name:      "markdown_and_case",
			input:     "**BOLD TEXT**",
			modifiers: []stringModifier{removeMarkdown, normalizeCase},
			expected:  "bold text",
		},
		{
			name:      "all_modifiers",
			input:     `**"HELLO, WORLD!"** with 1, 2, 3`,
			modifiers: availableModifiers,
			expected:  "helloworldwith123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyModifiers(tt.input, tt.modifiers)
			if result != tt.expected {
				t.Errorf("applyModifiers() = %q, want %q", result, tt.expected)
			}
		})
	}
}
