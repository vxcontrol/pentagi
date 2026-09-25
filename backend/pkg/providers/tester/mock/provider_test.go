package mock

import (
	"errors"
	"testing"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func TestProvider_SetSequentialResponses_AnswersInCallOrderAheadOfContentKeys(t *testing.T) {
	t.Parallel()

	chain := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "same prompt")}
	tool := []llms.Tool{{Type: "function", Function: &llms.FunctionDefinition{Name: "noop"}}}

	for _, tt := range []struct {
		name string
		call func(t *testing.T, p *Provider) (*llms.ContentResponse, error)
	}{
		{"a tool call", func(t *testing.T, p *Provider) (*llms.ContentResponse, error) {
			return p.CallWithTools(t.Context(), pconfig.OptionsTypePrimaryAgent, chain, tool, nil)
		}},
		{"a call carrying extra options", func(t *testing.T, p *Provider) (*llms.ContentResponse, error) {
			return p.CallWithExtraOptions(t.Context(), pconfig.OptionsTypeSimple, chain, nil, nil, llms.WithTemperature(0.25))
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
			p.SetResponses([]ResponseConfig{{Key: "same prompt", Response: "content key"}})
			p.SetSequentialResponses("first", "second")

			// the last configured response repeats past the end of the sequence
			for i, want := range []string{"first", "second", "second", "second"} {
				resp, err := tt.call(t, p)
				if err != nil {
					t.Fatalf("call %d: %v", i+1, err)
				}
				if len(resp.Choices) == 0 || resp.Choices[0].Content != want {
					t.Fatalf("call %d answered %+v, want %q", i+1, resp.Choices, want)
				}
			}
		})
	}
}

func TestProvider_CallWithExtraOptions_AnswersAToolRequestWithACall(t *testing.T) {
	t.Parallel()

	p := NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	tools := []llms.Tool{{Type: "function", Function: &llms.FunctionDefinition{Name: "noop"}}}
	chain := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "nothing canned matches this")}

	resp, err := p.CallWithExtraOptions(t.Context(), pconfig.OptionsTypeSimple, chain, tools, nil,
		llms.WithToolChoice("required"))
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if len(resp.Choices) == 0 || len(resp.Choices[0].ToolCalls) == 0 {
		t.Fatalf("a tool request answered with no tool call: any case carrying options routes here; resp=%+v", resp)
	}
	if got := resp.Choices[0].ToolCalls[0].FunctionCall.Name; got != "noop" {
		t.Errorf("tool call names %q, want the tool the request carried", got)
	}
}

func TestProvider_CallWithExtraOptions_LeavesTheCannedResponseAlone(t *testing.T) {
	t.Parallel()

	canned := &llms.ContentResponse{Choices: []*llms.ContentChoice{{
		Content:   "ok",
		Reasoning: &reasoning.ContentReasoning{Content: "canned trace"},
	}}}
	p := NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	p.SetResponses([]ResponseConfig{{Key: "hi", Response: canned}})
	chain := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}

	off, err := p.CallWithExtraOptions(t.Context(), pconfig.OptionsTypeSimple, chain, nil, nil,
		llms.WithReasoningDisabled())
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if off.Choices[0].Reasoning != nil {
		t.Errorf("reasoning survived a disabled request: %+v", off.Choices[0].Reasoning)
	}

	if canned.Choices[0].Reasoning.IsEmpty() {
		t.Fatal("the canned response lost its reasoning: a later case reusing this mock would see none")
	}

	again, err := p.CallWithExtraOptions(t.Context(), pconfig.OptionsTypeSimple, chain, nil, nil,
		llms.WithTemperature(0.25))
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if again.Choices[0].Reasoning.IsEmpty() {
		t.Error("a later plain request saw no reasoning: the earlier disabled one consumed it")
	}
}

func TestProvider_CallWithExtraOptions_MatchesAKeyInsideALongerPrompt(t *testing.T) {
	t.Parallel()

	p := NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	p.SetResponses([]ResponseConfig{{Key: "the key", Response: "canned"}})
	p.SetDefaultResponse("default")
	chain := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman,
		"a long request that mentions the key somewhere in the middle")}

	resp, err := p.CallWithExtraOptions(t.Context(), pconfig.OptionsTypeSimple, chain, nil, nil,
		llms.WithTemperature(0.25))
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if got := resp.Choices[0].Content; got != "canned" {
		t.Errorf("content = %q, want the canned answer: a case carrying options is rarely one sentence, so this path matches a key inside the prompt", got)
	}
}

func TestProvider_Call_AnswersWithTheResponseOfAKeyInsideThePrompt(t *testing.T) {
	t.Parallel()

	errRefused := errors.New("canned refusal")
	p := NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	p.SetResponses([]ResponseConfig{{Key: "the answer", Response: "canned"}, {Key: "the refusal", Response: errRefused}})
	p.SetDefaultResponse("default")

	for _, tc := range []struct {
		name    string
		prompt  string
		want    string
		wantErr error
	}{
		{"a canned answer", "a long prompt asking for the answer somewhere", "canned", nil},
		{"a canned error", "a long prompt asking for the refusal somewhere", "", errRefused},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := p.Call(t.Context(), pconfig.OptionsTypeSimple, tc.prompt)
			if got != tc.want || !errors.Is(err, tc.wantErr) {
				t.Errorf("Call = %q, %v; want %q, %v", got, err, tc.want, tc.wantErr)
			}
		})
	}
}
