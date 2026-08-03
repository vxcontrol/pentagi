package mock

import (
	"testing"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
)

// TestProvider_SetSequentialResponses covers the sequential-response mode
// CallWithTools needs for multi-turn scenarios (see tester.fileEditTestCase):
// each call gets the next configured response regardless of content, and the
// last configured response repeats for any call beyond the configured count.
func TestProvider_SetSequentialResponses(t *testing.T) {
	t.Parallel()

	p := NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	p.SetSequentialResponses("first", "second")

	chain := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "same prompt every time")}
	tool := []llms.Tool{{Type: "function", Function: &llms.FunctionDefinition{Name: "noop"}}}

	tests := []struct {
		name string
		want string
	}{
		{name: "first call returns the first response", want: "first"},
		{name: "second call returns the second response", want: "second"},
		{name: "third call repeats the last response", want: "second"},
		{name: "fourth call still repeats the last response", want: "second"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := p.CallWithTools(t.Context(), pconfig.OptionsTypePrimaryAgent, chain, tool, nil)
			if err != nil {
				t.Fatalf("CallWithTools() error = %v", err)
			}
			if len(resp.Choices) == 0 || resp.Choices[0].Content != tt.want {
				t.Fatalf("CallWithTools() content = %+v, want %q", resp.Choices, tt.want)
			}
		})
	}
}

// TestProvider_SetSequentialResponses_IgnoresContentMatching checks that,
// once configured, sequential mode takes priority over the normal
// content-keyed response lookup (SetResponses) - the two are mutually
// exclusive per test, by design.
func TestProvider_SetSequentialResponses_IgnoresContentMatching(t *testing.T) {
	t.Parallel()

	p := NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	p.SetResponses([]ResponseConfig{{Key: "same prompt", Response: "should not be used"}})
	p.SetSequentialResponses("sequential response")

	chain := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "same prompt")}
	tool := []llms.Tool{{Type: "function", Function: &llms.FunctionDefinition{Name: "noop"}}}

	resp, err := p.CallWithTools(t.Context(), pconfig.OptionsTypePrimaryAgent, chain, tool, nil)
	if err != nil {
		t.Fatalf("CallWithTools() error = %v", err)
	}
	if resp.Choices[0].Content != "sequential response" {
		t.Fatalf("CallWithTools() content = %q, want %q (content-matching must not take priority)", resp.Choices[0].Content, "sequential response")
	}
}
