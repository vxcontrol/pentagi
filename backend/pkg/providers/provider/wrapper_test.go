package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	"pentagi/pkg/providers/pconfig"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestWrapper_RecognizesTheObservedRateLimitResponses(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		err  string
		want bool
	}{
		{"Mistral through LiteLLM", "API returned unexpected status code: 429: litellm.RateLimitError: MistralException - Rate limit exceeded", true},
		{"compact status spelling", "statuscode: 429", true},
		{"rate limit without an HTTP status", "Rate limit exceeded", true},
		{"other rejected request", "API returned unexpected status code: 400: invalid arguments", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isTooManyRequestsError(errors.New(tc.err)); got != tc.want {
				t.Errorf("isTooManyRequestsError(%q) = %t, want %t", tc.err, got, tc.want)
			}
		})
	}
}

func adaptiveProvider(t *testing.T, raw string) *catalogProvider {
	t.Helper()

	config, err := pconfig.LoadConfigData([]byte(raw), []llms.CallOption{llms.WithModel("gateway-model")})
	require.NoError(t, err)

	return &catalogProvider{model: "gateway-model", prefix: "gateway-model", config: config}
}

func resolvedReasoning(options []llms.CallOption) *llms.ReasoningConfig {
	var resolved llms.CallOptions
	for _, option := range options {
		option(&resolved)
	}
	return resolved.Reasoning
}

type capturingModel struct {
	calls   int
	options []llms.CallOption
}

func (m *capturingModel) GenerateContent(
	_ context.Context, _ []llms.MessageContent, options ...llms.CallOption,
) (*llms.ContentResponse, error) {
	m.calls++
	m.options = options
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "ok"}}}, nil
}

func (m *capturingModel) Call(context.Context, string, ...llms.CallOption) (string, error) {
	panic("the wrappers only call GenerateContent")
}

// wrapperPaths drives both wrappers with the same input, keyed by the shape of the call.
var wrapperPaths = []struct {
	name string
	call func(prv Provider, model *capturingModel, input string, options ...llms.CallOption) error
}{
	{"a single prompt", func(prv Provider, model *capturingModel, input string, options ...llms.CallOption) error {
		_, err := WrapGenerateFromSinglePrompt(
			context.Background(), prv, pconfig.OptionsTypeSimple, model, input, options...)
		return err
	}},
	{"a chain", func(prv Provider, model *capturingModel, input string, options ...llms.CallOption) error {
		_, err := WrapGenerateContent(
			context.Background(), prv, pconfig.OptionsTypeSimple, model.GenerateContent,
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, input)}, options...)
		return err
	}},
}

func TestWrapper_AdaptiveOptions_CarriesOnlyAnAdaptiveChoice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		raw    string
		models pconfig.ModelsConfig
		want   *llms.ReasoningConfig
	}{
		{
			name: "an agent that chose adaptive",
			raw:  `{"simple": {"reasoning": {"mode": "adaptive", "effort": "low"}}}`,
			want: &llms.ReasoningConfig{Adaptive: true, Effort: llms.ReasoningLow},
		},
		{
			name: "an adaptive-only model the agent set no reasoning for",
			raw:  `{"simple": {"model": "claude-opus-4-8"}}`,
			models: pconfig.ModelsConfig{{
				Name:      "claude-opus-4-8",
				Reasoning: &pconfig.ModelReasoningInfo{Mode: pconfig.ModelReasoningAdaptiveOnly},
			}},
			want: &llms.ReasoningConfig{Adaptive: true},
		},
		{name: "an agent on a thinking budget", raw: `{"simple": {"reasoning": {"mode": "budget", "max_tokens": 2048}}}`},
		{name: "a provider without a config"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prv := &catalogProvider{}
			if tc.raw != "" {
				prv = adaptiveProvider(t, tc.raw)
			}
			prv.models = tc.models

			_, options := adaptiveOptions(context.Background(), prv, pconfig.OptionsTypeSimple, nil)

			assert.Equal(t, tc.want, resolvedReasoning(options))
		})
	}
}

// TestWrapper_WrapGenerate_SendsTheCatalogueCeilingAndTheAgentsAdaptiveChoice is keyed by wrapper.
func TestWrapper_WrapGenerate_SendsTheCatalogueCeilingAndTheAgentsAdaptiveChoice(t *testing.T) {
	t.Parallel()

	for _, path := range wrapperPaths {
		t.Run(path.name, func(t *testing.T) {
			t.Parallel()

			prv := adaptiveProvider(t, `{"simple": {"reasoning": {"mode": "adaptive", "effort": "medium"}}}`)
			prv.model, prv.prefix, prv.models = "gpt-4o", "gpt-4o", catalogue()
			options := append(prv.config.GetOptionsForType(pconfig.OptionsTypeSimple),
				llms.WithMaxTokens(32768), llms.WithReasoningDisabled())

			model := &capturingModel{}
			require.NoError(t, path.call(prv, model, "hi", options...))

			assert.Equal(t, limit(16384), appliedMaxTokens(t, model.options))
			assert.Equal(t, &llms.ReasoningConfig{Adaptive: true, Effort: llms.ReasoningMedium},
				resolvedReasoning(model.options), "a reasoning extra replaced the agent's adaptive thinking")
		})
	}
}

// TestWrapper_WrapGenerate_RefusesAChainThatCannotFit is keyed by wrapper.
func TestWrapper_WrapGenerate_RefusesAChainThatCannotFit(t *testing.T) {
	t.Parallel()

	for _, path := range wrapperPaths {
		t.Run(path.name, func(t *testing.T) {
			t.Parallel()

			model := &capturingModel{}
			err := path.call(&catalogProvider{model: "gpt-4o", prefix: "gpt-4o", models: catalogue()},
				model, strings.Repeat("x", 600_000), llms.WithMaxTokens(4096))

			var windowErr *ContextWindowError
			require.ErrorAs(t, err, &windowErr)
			assert.Zero(t, model.calls, "the vendor must not be called with a chain that cannot fit")
		})
	}
}
