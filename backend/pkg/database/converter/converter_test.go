package converter

import (
	"testing"

	"pentagi/pkg/graph/model"
	"pentagi/pkg/providers/openai"
	"pentagi/pkg/providers/pconfig"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func TestIsAgentTool(t *testing.T) {
	tests := []struct {
		name         string
		functionName string
		expected     bool
	}{
		// Agent tools
		{"coder is agent", "coder", true},
		{"pentester is agent", "pentester", true},
		{"maintenance is agent", "maintenance", true},
		{"memorist is agent", "memorist", true},
		{"search is agent", "search", true},
		{"advice is agent", "advice", true},

		// Agent result tools (also agents)
		{"coder_result is agent", "code_result", true},
		{"hack_result is agent", "hack_result", true},
		{"maintenance_result is agent", "maintenance_result", true},
		{"memorist_result is agent", "memorist_result", true},
		{"search_result is agent", "search_result", true},
		{"enricher_result is agent", "enricher_result", true},
		{"report_result is agent", "report_result", true},
		{"subtask_list is agent", "subtask_list", true},
		{"subtask_patch is agent", "subtask_patch", true},

		// Non-agent tools
		{"terminal is not agent", "terminal", false},
		{"file is not agent", "file", false},
		{"browser is not agent", "browser", false},
		{"google is not agent", "google", false},
		{"duckduckgo is not agent", "duckduckgo", false},
		{"tavily is not agent", "tavily", false},
		{"sploitus is not agent", "sploitus", false},
		{"searxng is not agent", "searxng", false},
		{"search_in_memory is not agent", "search_in_memory", false},
		{"store_guide is not agent", "store_guide", false},
		{"done is not agent", "done", false},
		{"ask is not agent", "ask", false},

		// Unknown tool
		{"unknown tool is not agent", "unknown_tool", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAgentTool(tt.functionName)
			if result != tt.expected {
				t.Errorf("isAgentTool(%q) = %v, want %v", tt.functionName, result, tt.expected)
			}
		})
	}
}

func TestConvertModelsReasoning(t *testing.T) {
	models := pconfig.ModelsConfig{
		{Name: "ao", Reasoning: &pconfig.ModelReasoningInfo{
			Mode:    pconfig.ModelReasoningAdaptiveOnly,
			Efforts: []llms.ReasoningEffort{llms.ReasoningLow, llms.ReasoningEffort("xhigh")},
		}},
		{Name: "ad", Reasoning: &pconfig.ModelReasoningInfo{Mode: pconfig.ModelReasoningAdaptive}},
		{Name: "bd", Reasoning: &pconfig.ModelReasoningInfo{Mode: pconfig.ModelReasoningBudget}},
		{Name: "none"},
	}

	byName := make(map[string]*model.ModelConfig)
	for _, m := range ConvertModels(models, reasoning.ProviderUnknown) {
		byName[m.Name] = m
	}

	require.NotNil(t, byName["ao"].Reasoning)
	require.NotNil(t, byName["ao"].Reasoning.Mode)
	assert.Equal(t, model.ModelReasoningModeAdaptiveOnly, *byName["ao"].Reasoning.Mode, "hyphenated pconfig value must map to the GraphQL underscore enum")
	assert.Equal(t, []model.ReasoningEffort{model.ReasoningEffort("low"), model.ReasoningEffort("xhigh")}, byName["ao"].Reasoning.Efforts)

	require.NotNil(t, byName["ad"].Reasoning.Mode)
	assert.Equal(t, model.ModelReasoningModeAdaptive, *byName["ad"].Reasoning.Mode)
	require.NotNil(t, byName["bd"].Reasoning.Mode)
	assert.Equal(t, model.ModelReasoningModeBudget, *byName["bd"].Reasoning.Mode)

	assert.Nil(t, byName["none"].Reasoning, "model without a reasoning descriptor maps to nil")
}

func TestConvertModels_OffCapability(t *testing.T) {
	tr := func(b bool) *bool { return &b }
	models := pconfig.ModelsConfig{
		// Gemini: thinking:true, no reasoning block -> capability still surfaced,
		// Off effective via thinkingBudget:0.
		{Name: "gemini-2.5-flash", Thinking: tr(true)},
		// Reasoning model with no provider-specific disable wire (OffOmit) and
		// unknown default-on -> Off is a silent no-op, so cannotDisable is true.
		{Name: "glm-5.2", Reasoning: &pconfig.ModelReasoningInfo{Mode: pconfig.ModelReasoningBudget}},
		// Adaptive-only but off by default: Off works by omission -> disablable.
		{Name: "claude-opus-4-8", Reasoning: &pconfig.ModelReasoningInfo{Mode: pconfig.ModelReasoningAdaptiveOnly}},
		// Always-on Claude: Off is rejected (OffUnsupported) -> not disablable.
		{Name: "claude-fable-5", Reasoning: &pconfig.ModelReasoningInfo{Mode: pconfig.ModelReasoningAdaptiveOnly}},
		// Plain non-thinking model: no reasoning capability at all.
		{Name: "gpt-4.1"},
	}

	byProvider := map[string]reasoning.Provider{
		"gemini-2.5-flash": reasoning.ProviderGoogleAI,
		// ProviderUnknown: ResolveOff returns OffOmit (no clean disable signal).
		// ProviderOpenAI would instead return OffEffortNone (optimistic).
		"glm-5.2":         reasoning.ProviderUnknown,
		"claude-opus-4-8": reasoning.ProviderAnthropic,
		"claude-fable-5":  reasoning.ProviderAnthropic,
		"gpt-4.1":         reasoning.ProviderOpenAI,
	}

	get := func(name string) *model.ModelReasoningInfo {
		for _, mc := range ConvertModels(pconfig.ModelsConfig{lookup(models, name)}, byProvider[name]) {
			if mc.Name == name {
				return mc.Reasoning
			}
		}
		return nil
	}

	// Gemini: capability present, Off effective.
	gem := get("gemini-2.5-flash")
	require.NotNil(t, gem, "thinking-capable model must surface reasoning capability")
	require.NotNil(t, gem.CannotDisable)
	assert.False(t, *gem.CannotDisable, "Gemini Off works via thinkingBudget:0")

	// glm-5.2: OffOmit + unknown default-on -> no-op, hide Off.
	glm := get("glm-5.2")
	require.NotNil(t, glm)
	require.NotNil(t, glm.CannotDisable)
	assert.True(t, *glm.CannotDisable, "OffOmit with unknown default-on is a no-op, must hide Off")

	// Same model on OpenAI: ResolveOff is optimistic (OffEffortNone) -> Off offered.
	glmOA := ConvertModels(
		pconfig.ModelsConfig{lookup(models, "glm-5.2")},
		reasoning.ProviderOpenAI,
	)[0].Reasoning
	require.NotNil(t, glmOA)
	require.NotNil(t, glmOA.CannotDisable)
	assert.False(t, *glmOA.CannotDisable, "OpenAI ResolveOff attempts effort none")

	// opus-4-8: off by default -> disablable.
	opus := get("claude-opus-4-8")
	require.NotNil(t, opus.CannotDisable)
	assert.False(t, *opus.CannotDisable, "off-by-default model: Off works by omission")

	// fable-5: always on -> not disablable.
	fable := get("claude-fable-5")
	require.NotNil(t, fable.CannotDisable)
	assert.True(t, *fable.CannotDisable, "always-on model rejects Off")

	// non-thinking model: no capability.
	assert.Nil(t, get("gpt-4.1"), "non-thinking model has no reasoning capability")
}

func lookup(ms pconfig.ModelsConfig, name string) pconfig.ModelConfig {
	for _, m := range ms {
		if m.Name == name {
			return m
		}
	}
	return pconfig.ModelConfig{}
}

func TestOffEffective(t *testing.T) {
	tr := func(b bool) *bool { return &b }
	assert.True(t, offEffective(reasoning.OffDisableClaude, nil))
	assert.True(t, offEffective(reasoning.OffZeroBudget, nil))
	assert.True(t, offEffective(reasoning.OffEffortNone, nil))
	assert.True(t, offEffective(reasoning.OffOmit, tr(false)), "off-by-default: omit disables")
	assert.False(t, offEffective(reasoning.OffOmit, tr(true)), "default-on + omit = no-op")
	assert.False(t, offEffective(reasoning.OffOmit, nil), "unknown default + omit = assume no-op")
	assert.False(t, offEffective(reasoning.OffUnsupported, nil), "rejected")
}

func TestConvertAgentConfigExtraBodyRoundTrip(t *testing.T) {
	extraBody := map[string]any{
		"chat_template_kwargs": map[string]any{"enable_thinking": false},
		"provider":             "vllm",
	}

	from := ConvertAgentConfigFromGqlModel(&model.AgentConfig{Model: "m", ExtraBody: extraBody})
	require.NotNil(t, from)
	assert.Equal(t, extraBody, from.ExtraBody, "extra_body must survive the GraphQL→pconfig converter, not get dropped")

	back := ConvertAgentConfigToGqlModel(from)
	require.NotNil(t, back)
	assert.Equal(t, extraBody, back.ExtraBody, "extra_body must round-trip back to the GraphQL model")
}

func TestConvertAgentConfigExtraBodyNilStaysNil(t *testing.T) {
	from := ConvertAgentConfigFromGqlModel(&model.AgentConfig{Model: "m"})
	require.NotNil(t, from)
	assert.Nil(t, from.ExtraBody, "absent extra_body must stay nil, not become an empty map")

	assert.Nil(t, ConvertAgentConfigToGqlModel(from).ExtraBody)
}

func ptr[T any](value T) *T { return &value }

func TestConvertAgentConfigCallOptionFieldsRoundTrip(t *testing.T) {
	from := ConvertAgentConfigFromGqlModel(&model.AgentConfig{
		Model:            "m",
		MinP:             ptr(0.05),
		N:                ptr(3),
		JSON:             ptr(true),
		ResponseMimeType: ptr("application/json"),
	})
	require.NotNil(t, from)
	assert.Equal(t, 0.05, from.MinP)
	assert.Equal(t, 3, from.N)
	assert.True(t, from.JSON)
	assert.Equal(t, "application/json", from.ResponseMIMEType)

	back := ConvertAgentConfigToGqlModel(from)
	require.NotNil(t, back)
	require.NotNil(t, back.MinP)
	require.NotNil(t, back.N)
	require.NotNil(t, back.JSON)
	require.NotNil(t, back.ResponseMimeType)
	assert.Equal(t, 0.05, *back.MinP)
	assert.Equal(t, 3, *back.N)
	assert.True(t, *back.JSON)
	assert.Equal(t, "application/json", *back.ResponseMimeType)
}

// pconfig gates each emitter on the raw key being PRESENT, not on its value, so a zero written
// through would switch the option on: `json: false` must not reach the LLM as WithJSONMode().
func TestConvertAgentConfigZeroCallOptionFieldsStayAbsent(t *testing.T) {
	from := ConvertAgentConfigFromGqlModel(&model.AgentConfig{
		Model:            "m",
		MinP:             ptr(0.0),
		N:                ptr(0),
		JSON:             ptr(false),
		ResponseMimeType: ptr(""),
	})
	require.NotNil(t, from)

	options := from.BuildOptions()
	call := llms.CallOptions{}
	for _, option := range options {
		option(&call)
	}

	assert.False(t, call.JSONMode, "json:false must not enable JSON mode")
	assert.Nil(t, call.N, "n:0 must not reach the request as an explicit parameter")
	assert.Empty(t, call.ResponseMIMEType)
}

// The shipped simple_json default carries `json: true`; saving a provider through the UI used to
// strip it, so the reloaded config called the LLM without JSON mode while the built-in one did not.
func TestShippedSimpleJSONKeepsJSONModeThroughTheGraphQLRoundTrip(t *testing.T) {
	pc, err := openai.DefaultProviderConfig()
	require.NoError(t, err)
	require.NotNil(t, pc.SimpleJSON)
	require.True(t, pc.SimpleJSON.JSON, "the shipped default is what makes this test meaningful")

	gql := ConvertProviderConfigToGqlModel(pc)
	require.NotNil(t, gql)

	restored := ConvertAgentsConfigFromGqlModel(gql)
	require.NotNil(t, restored)
	require.NotNil(t, restored.SimpleJSON)

	call := llms.CallOptions{}
	for _, option := range restored.SimpleJSON.BuildOptions() {
		option(&call)
	}

	assert.True(t, call.JSONMode, "the simple_json agent must still ask for JSON mode after a save")
}
