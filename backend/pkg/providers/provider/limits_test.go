package provider

import (
	"strings"
	"testing"

	"pentagi/pkg/providers/pconfig"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestLimits_ModelLimitsFor_ReadsTheCatalogueEntryOfTheModelTheAgentCalls(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		model  string
		prefix string
		want   ModelLimits
	}{
		{
			name: "a prefixed call joins on the bare name", model: "gpt-4o", prefix: "openai/gpt-4o",
			want: ModelLimits{ContextWindow: limit(128000), MaxOutputTokens: limit(16384)},
		},
		{name: "a model the catalogue omits stays unlimited", model: "gpt-unknown"},
		{name: "a model without published limits stays unlimited", model: "no-limits"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ModelLimitsFor(
				&catalogProvider{model: tc.model, prefix: tc.prefix, models: catalogue()},
				pconfig.OptionsTypeSimple,
			)

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestLimits_CapOutputTokens_LowersOnlyAnAskOverTheCeiling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options []llms.CallOption
		ceiling *int
		want    *int
	}{
		{"an ask over the ceiling is lowered", []llms.CallOption{llms.WithMaxTokens(32768)}, limit(16384), limit(16384)},
		{"a fitting ask is left alone", []llms.CallOption{llms.WithMaxTokens(8192)}, limit(16384), limit(8192)},
		{"a silent catalogue leaves the ask alone", []llms.CallOption{llms.WithMaxTokens(32768)}, nil, limit(32768)},
		{"an absent ask keeps the door's own default", nil, limit(16384), nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, appliedMaxTokens(t, capOutputTokens(tc.options, tc.ceiling)))
		})
	}
}

func TestLimits_FitWithinWindow_LeavesTheSmallestUsefulAnswerOrRefuses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		model   string
		bytes   int
		tools   []llms.Tool
		ask     int
		window  *int
		want    *int
		wantErr string
	}{
		{name: "a fitting request is left alone", model: "gpt-4o", bytes: 4000, ask: 4096, window: limit(128000), want: limit(4096)},
		{name: "the ceiling is lowered to the room that is left", model: "small", bytes: 40000, ask: 8192, window: limit(12000), want: limit(2000)},
		{name: "the smallest useful answer still fits", model: "small", bytes: 43904, ask: 8192, window: limit(12000), want: limit(1024)},
		{name: "an unknown window leaves the call alone", model: "no-limits", bytes: 4_000_000, ask: 8192, want: limit(8192)},
		{
			name: "a request that leaves no room is refused", model: "small", bytes: 47000, ask: 8192, window: limit(12000),
			wantErr: "request to small is estimated at 11750 input tokens and leaves no room for an answer in its 12000 token context window",
		},
		{
			name: "one token below the smallest useful answer is refused", model: "small", bytes: 43908, ask: 8192, window: limit(12000),
			wantErr: "request to small is estimated at 10977 input tokens and leaves no room for an answer in its 12000 token context window",
		},
		{
			name: "the tool schemas take their share of the room", model: "small", bytes: 4000, tools: toolsOfBytes(t, 20000),
			ask: 8192, window: limit(12000), want: limit(5975),
		},
		{
			name: "a chain that fits alone is refused once its tool schemas leave no room", model: "small", bytes: 4000,
			tools: toolsOfBytes(t, 40000), ask: 8192, window: limit(12000),
			wantErr: "request to small is estimated at 11025 input tokens and leaves no room for an answer in its 12000 token context window",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fitted, err := fitWithinWindow(
				tc.model, chainOfBytes(t, tc.bytes),
				[]llms.CallOption{llms.WithMaxTokens(tc.ask), llms.WithTools(tc.tools)}, tc.window,
			)

			if tc.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.want, appliedMaxTokens(t, fitted))
				return
			}

			assert.Nil(t, fitted)

			var windowErr *ContextWindowError
			require.ErrorAs(t, err, &windowErr)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func TestLimits_ChainByteBudget_MeasuresTheWindowInTheBytesTheSummarizerCounts(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 43904, ChainByteBudget(12000, nil))
	assert.Equal(t, 23804, ChainByteBudget(12000, toolsOfBytes(t, 20000)), "the tool schemas take their share first")
	assert.Zero(t, ChainByteBudget(1024, nil), "a window this small cannot hold a call at all")
	assert.Zero(t, ChainByteBudget(6049, toolsOfBytes(t, 20000)), "a window the tool schemas fill leaves nothing to compact into")
	assert.Zero(t, ChainByteBudget(0, nil))
}

func TestLimits_ChainFitsWindow_LeavesRoomForTheToolSchemasAndTheSmallestUsefulAnswer(t *testing.T) {
	t.Parallel()

	schemas := toolsOfBytes(t, 20000)

	assert.True(t, ChainFitsWindow(chainOfBytes(t, 43904), nil, 12000))
	assert.False(t, ChainFitsWindow(chainOfBytes(t, 43908), nil, 12000))
	assert.True(t, ChainFitsWindow(chainOfBytes(t, 23804), schemas, 12000))
	assert.False(t, ChainFitsWindow(chainOfBytes(t, 23808), schemas, 12000), "a chain that fits alone overflows beside its tools")
	assert.True(t, ChainFitsWindow(chainOfBytes(t, 4_000_000), schemas, 0), "an unknown window never asks for compaction")
}

func chainOfBytes(t *testing.T, size int) []llms.MessageContent {
	t.Helper()

	return []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, strings.Repeat("x", size)),
	}
}

// toolsOfBytes is one tool whose schema marshals to description+100 bytes.
func toolsOfBytes(t *testing.T, description int) []llms.Tool {
	t.Helper()

	return []llms.Tool{{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "terminal",
			Description: strings.Repeat("d", description),
			Parameters:  map[string]any{"type": "object"},
		},
	}}
}
