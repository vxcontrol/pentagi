package cases

import (
	"testing"
	"time"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

func TestCompletion_Execute_MatchesTheExpectedTextInEitherResponseShape(t *testing.T) {
	definitions := testerDefinitions(t, `
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
`)

	testCase, err := newCompletionTestCase(definitions[0])
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

	result := testCase.Execute("The answer is 4", time.Millisecond*100)
	if !result.Success {
		t.Errorf("Expected success for correct response, got failure: %v", result.Error)
	}
	if result.Latency != time.Millisecond*100 {
		t.Errorf("Expected latency 100ms, got %v", result.Latency)
	}

	result = testCase.Execute("The answer is 5", time.Millisecond*50)
	if result.Success {
		t.Errorf("Expected failure for incorrect response, got success")
	}

	testCase, err = newCompletionTestCase(definitions[1])
	if err != nil {
		t.Fatalf("Failed to create messages test case: %v", err)
	}
	if len(testCase.Messages()) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(testCase.Messages()))
	}

	result = testCase.Execute(&llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "The result is 50"}}}, time.Millisecond*200)
	if !result.Success {
		t.Errorf("Expected success for ContentResponse, got failure: %v", result.Error)
	}
}

func TestCompletion_ContainsString_ToleratesFormattingButNotADifferentAnswer(t *testing.T) {
	tests := []struct {
		name     string
		response string
		expected string
		want     bool
	}{
		{"exact match", "4", "4", true},
		{"exact match of text", "hello world", "hello world", true},
		{"empty response", "", "4", false},

		{"contains the expected", "The answer is 4", "4", true},
		{"contained in the expected", "4", "The answer is 4", true},

		{"answer is one dot", ".", "The answer is 4", false},
		{"answer is one dot against a number", ".", "4", false},
		{"answer is punctuation only", "...!?", "The answer is 4", false},
		{"answer is whitespace only", "   ", "The answer is 4", false},

		{"upper case", "HELLO WORLD", "hello world", true},
		{"mixed case", "Hello World", "HELLO world", true},
		{"case in a sentence", "The Answer Is CORRECT", "answer is correct", true},

		{"spaces between numbers", "1,2,3,4,5", "1, 2, 3, 4, 5", true},
		{"a tab", "hello\tworld", "hello world", true},
		{"a newline", "hello\nworld", "hello world", true},
		{"mixed whitespace", "a\t b\n c\r d", "a b c d", true},
		{"a space separated number sequence", "1 2 3 4 5", "1,2,3,4,5", true},

		{"markdown bold", "This is **bold** text", "This is bold text", true},
		{"markdown italic", "This is *italic* text", "This is italic text", true},
		{"markdown code", "Use `code` here", "Use code here", true},
		{"markdown header", "# Header text", "Header text", true},
		{"markdown link", "[link text](url)", "link text url", true},
		{"markdown blockquote", "> quoted text", "quoted text", true},
		{"markdown list", "- item one", "item one", true},
		{"markdown mixed", "**Bold** and *italic* with `code`", "Bold and italic with code", true},

		{"punctuation", "Hello, world!", "Hello world", true},
		{"a question mark", "Is this correct?", "Is this correct", true},
		{"parentheses", "Text (in brackets)", "Text in brackets", true},
		{"mixed punctuation in a sentence", "Hello, world! How are you?", "Hello world How are you", true},

		{"double quotes", `He said "hello"`, "He said hello", true},
		{"single quotes", "It's a 'test'", "Its a test", true},
		{"a quoted phrase", "\"Smart quotes\"", "Smart quotes", true},
		{"backticks", "`quoted text`", "quoted text", true},

		{"a comma spaced number sequence", "sequence: 1, 2, 3, 4, 5", "1,2,3,4,5", true},
		{"a space separated sequence after a word", "count 1 2 3 4 5", "1,2,3,4,5", true},
		{"a dash separated sequence", "range: 1-2-3-4-5", "1,2,3,4,5", true},
		{"a dot separated sequence", "version 1.2.3.4.5", "1,2,3,4,5", true},

		{"case and whitespace together", "HELLO  WORLD", "hello world", true},
		{"case and punctuation together", "HELLO, WORLD!", "hello world", true},
		{"markdown and case together", "**BOLD TEXT**", "bold text", true},
		{"every modifier together", "**HELLO,**  `world`!", "hello world", true},
		{"a markdown quote with case", "> **Important:** Use `this` method!", "Important Use this method", true},

		{"nested markdown", "**Bold *and italic* text**", "Bold and italic text", true},
		{"repeated spaces", "hello    world", "hello world", true},
		{"repeated punctuation", "Hello... world!!!", "Hello world", true},
		{"a code block", "```\ncode here\n```", "code here", true},

		{"different text", "completely different", "expected text", false},
		{"different numbers", "1,2,3", "4,5,6", false},
		{"a word the expected lacks", "partial", "completely different text", false},

		{"a number at the end of a sentence", "The answer to your question is: 42", "42", true},
		{"a formatted answer", "**Answer:** The result is `50`", "The result is 50", true},
		{"a list of steps", "Here are the steps:\n- Step 1\n- Step 2", "Step 1 Step 2", true},
		{"a function name in code", "Use this function: `calculateSum()`", "calculateSum", true},
		{"a number in parentheses", "The value (approximately 3.14) is correct", "3.14", true},

		{"a short answer inside the expected", "answer", "The answer is 42", true},
		{"the expected inside a long answer", "The answer is 42", "answer", true},
		{"a short answer inside a formatted expected", "ANSWER", "the **answer** is correct", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsString(tt.response, tt.expected); got != tt.want {
				t.Errorf("containsString(%q, %q) = %v, want %v", tt.response, tt.expected, got, tt.want)
			}
		})
	}
}

func TestCompletion_ApplyModifiers_AppliesEachModifierInOrder(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		modifiers []stringModifier
		expected  string
	}{
		{"lower case", "HELLO World", []stringModifier{normalizeCase}, "hello world"},
		{"no whitespace", "hello \t\n\r world", []stringModifier{removeWhitespace}, "helloworld"},
		{"no markdown", "**bold** and *italic* with `code`", []stringModifier{removeMarkdown}, "bold and italic with code"},
		{"no punctuation", "Hello, world!", []stringModifier{removePunctuation}, "Hello world"},
		{"no quotes", `"Hello" and 'world'`, []stringModifier{removeQuotes}, "Hello and world"},
		{"a joined number sequence", "sequence: 1, 2, 3, 4, 5", []stringModifier{normalizeNumbers}, "sequence: 1,2,3,4,5"},
		{"case then whitespace", "HELLO  WORLD", []stringModifier{normalizeCase, removeWhitespace}, "helloworld"},
		{"markdown then case", "**BOLD TEXT**", []stringModifier{removeMarkdown, normalizeCase}, "bold text"},
		{"numbers joined before the whitespace goes", "1 2 3 4 5", []stringModifier{normalizeNumbers, removeWhitespace}, "1,2,3,4,5"},
		{"every modifier", `**"HELLO, WORLD!"** with 1, 2, 3`, availableModifiers, "helloworldwith123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := applyModifiers(tt.input, tt.modifiers); result != tt.expected {
				t.Errorf("applyModifiers() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCompletion_Execute_FailsReasoningOffOnTextOrACountAboveOneToken(t *testing.T) {
	for _, tc := range []struct {
		label    string
		choice   llms.ContentChoice
		streamed string
		wantPass bool
	}{
		{
			label:    "a reasoning token counted with no text, as Moonshot reports thinking disabled",
			choice:   llms.ContentChoice{Content: "Paris", GenerationInfo: map[string]any{"ReasoningTokens": 1}},
			wantPass: true,
		},
		{
			label:  "two reasoning tokens counted with no text",
			choice: llms.ContentChoice{Content: "Paris", GenerationInfo: map[string]any{"ReasoningTokens": 2}},
		},
		{
			label:  "reasoning counted with no text, as OpenAI reports a reasoning pass on Chat Completions",
			choice: llms.ContentChoice{Content: "Paris", GenerationInfo: map[string]any{"ReasoningTokens": 6}},
		},
		{
			label:    "no reasoning at all",
			choice:   llms.ContentChoice{Content: "Paris", GenerationInfo: map[string]any{"ReasoningTokens": 0}},
			wantPass: true,
		},
		{
			label: "reasoning text returned",
			choice: llms.ContentChoice{
				Content:        "Paris",
				Reasoning:      &reasoning.ContentReasoning{Content: "The capital of France is Paris."},
				GenerationInfo: map[string]any{"ReasoningTokens": 44},
			},
		},
		{
			label:    "reasoning text that arrived while streaming",
			choice:   llms.ContentChoice{Content: "Paris"},
			streamed: "thinking it over",
		},
	} {
		t.Run(tc.label, func(t *testing.T) {
			testCase, err := newCompletionTestCase(builtinDefinition(t, "reasoning_off_suppresses_reasoning"))
			if err != nil {
				t.Fatalf("case: %v", err)
			}

			if tc.streamed != "" {
				completion := testCase.(*testCaseCompletion)
				completion.def.Streaming = true
				if err := completion.StreamingCallback()(t.Context(), streaming.Chunk{
					Reasoning: &reasoning.ContentReasoning{Content: tc.streamed},
				}); err != nil {
					t.Fatalf("stream: %v", err)
				}
			}

			choice := tc.choice
			result := testCase.Execute(&llms.ContentResponse{Choices: []*llms.ContentChoice{&choice}}, time.Second)
			if result.Success != tc.wantPass {
				t.Errorf("success = %v, want %v (error: %v)", result.Success, tc.wantPass, result.Error)
			}
		})
	}
}
