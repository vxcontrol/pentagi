package provider

import (
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
)

func textMessage(role llms.ChatMessageType, body string) llms.MessageContent {
	return llms.MessageContent{
		Role:  role,
		Parts: []llms.ContentPart{llms.TextContent{Text: body}},
	}
}

func resolveMaxTokens(t *testing.T, options []llms.CallOption) (int, bool) {
	t.Helper()
	resolved := llms.CallOptions{}
	for _, option := range options {
		option(&resolved)
	}
	if resolved.MaxTokens == nil {
		return 0, false
	}
	return *resolved.MaxTokens, true
}

func TestEstimatePromptTokensCountsToolSchemas(t *testing.T) {
	messages := []llms.MessageContent{textMessage(llms.ChatMessageTypeHuman, strings.Repeat("a", 400))}

	bare := EstimatePromptTokens(messages, nil)
	withTool := EstimatePromptTokens(messages, []llms.Tool{{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "run_terminal",
			Description: strings.Repeat("d", 200),
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{"cmd": map[string]any{"type": "string"}},
			},
		},
	}})

	if withTool <= bare {
		t.Fatalf("tool schemas must add to the estimate: bare=%d withTool=%d", bare, withTool)
	}
}

func TestEstimatePromptTokensChargesUnknownParts(t *testing.T) {
	// A part type with no cheap size must never be counted as free, or the
	// clamp can be defeated by exactly the payloads most likely to be large.
	messages := []llms.MessageContent{{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.BinaryContent{MIMEType: "image/png", Data: []byte{1, 2, 3}}},
	}}

	if got := EstimatePromptTokens(messages, nil); got <= 0 {
		t.Fatalf("unknown part must carry a cost, got %d", got)
	}
}

func TestClampMaxTokensDisabledWithoutWindow(t *testing.T) {
	options := []llms.CallOption{llms.WithMaxTokens(32000)}
	messages := []llms.MessageContent{textMessage(llms.ChatMessageTypeHuman, strings.Repeat("a", 40000))}

	_, _, ok := ClampMaxTokens(0, messages, options)
	if ok {
		t.Fatal("clamping must be off when no window is configured")
	}
}

func TestClampMaxTokensLeavesFittingRequestAlone(t *testing.T) {
	options := []llms.CallOption{llms.WithMaxTokens(1024)}
	messages := []llms.MessageContent{textMessage(llms.ChatMessageTypeHuman, "short prompt")}

	_, _, ok := ClampMaxTokens(32768, messages, options)
	if ok {
		t.Fatal("a request that already fits must not be modified")
	}
}

func TestClampMaxTokensShrinksOversizedBudget(t *testing.T) {
	// The reported scenario: 32k window, a big first Generator prompt, and a
	// max_tokens that leaves no room for it.
	prompt := strings.Repeat("a", 80000) // ~20000 tokens at 4 bytes/token
	messages := []llms.MessageContent{textMessage(llms.ChatMessageTypeHuman, prompt)}
	options := []llms.CallOption{llms.WithMaxTokens(26214)}

	clamped, budget, ok := ClampMaxTokens(32768, messages, options)
	if !ok {
		t.Fatal("an overflowing request must be clamped")
	}

	resolved, present := resolveMaxTokens(t, clamped)
	if !present {
		t.Fatal("clamped options must still carry max_tokens")
	}
	if resolved != budget {
		t.Fatalf("returned budget %d does not match resolved option %d", budget, resolved)
	}
	if resolved >= 26214 {
		t.Fatalf("clamp must reduce the budget, got %d", resolved)
	}

	estimate := EstimatePromptTokens(messages, nil)
	if estimate+resolved+WindowSafetyMargin > 32768 {
		t.Fatalf(
			"clamped request still overflows: estimate=%d budget=%d margin=%d window=32768",
			estimate, resolved, WindowSafetyMargin,
		)
	}
}

func TestClampMaxTokensDoesNotProduceUselessBudget(t *testing.T) {
	// Prompt at the window: clamping could only yield a stub reply, which hides
	// the real problem. Leave it and let the backend report the overflow.
	prompt := strings.Repeat("a", 32768*BytesPerTokenEstimate)
	messages := []llms.MessageContent{textMessage(llms.ChatMessageTypeHuman, prompt)}
	options := []llms.CallOption{llms.WithMaxTokens(4096)}

	_, _, ok := ClampMaxTokens(32768, messages, options)
	if ok {
		t.Fatal("must not clamp when the prompt alone fills the window")
	}
}

func TestClampMaxTokensIgnoresUnsetBudget(t *testing.T) {
	messages := []llms.MessageContent{textMessage(llms.ChatMessageTypeHuman, strings.Repeat("a", 80000))}

	_, _, ok := ClampMaxTokens(32768, messages, nil)
	if ok {
		t.Fatal("no configured max_tokens means there is nothing of ours to clamp")
	}
}

func TestClampMaxTokensCountsToolsTowardWindow(t *testing.T) {
	// Tool schemas are part of the prompt on the wire. A request that fits
	// without them but not with them must still be clamped — this is the first
	// Generator call in the issue, which carries several tool definitions.
	messages := []llms.MessageContent{textMessage(llms.ChatMessageTypeHuman, strings.Repeat("a", 60000))}
	options := []llms.CallOption{llms.WithMaxTokens(16000)}

	tools := []llms.Tool{{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "terminal",
			Description: strings.Repeat("d", 20000),
			Parameters:  map[string]any{"type": "object"},
		},
	}}
	withTools := append(options, llms.WithTools(tools))

	// Sized so the messages alone fit and the tool schemas tip it over. If tool
	// schemas were not counted, the second call would also be left alone — and
	// the request would then be rejected by the backend.
	if _, _, ok := ClampMaxTokens(32768, messages, options); ok {
		t.Fatal("messages alone fit this window; nothing should be clamped")
	}

	_, budget, ok := ClampMaxTokens(32768, messages, withTools)
	if !ok {
		t.Fatal("tool schemas push this request over the window and must be counted")
	}
	if budget >= 16000 {
		t.Fatalf("clamp must reduce the budget below the configured value, got %d", budget)
	}
}
