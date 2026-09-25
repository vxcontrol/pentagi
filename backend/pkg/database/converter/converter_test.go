package converter

import (
	"database/sql"
	"slices"
	"testing"
	"time"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/model"
	"pentagi/pkg/providers/anthropic"
	"pentagi/pkg/providers/bedrock"
	"pentagi/pkg/providers/deepseek"
	"pentagi/pkg/providers/glm"
	"pentagi/pkg/providers/kimi"
	"pentagi/pkg/providers/minimax"
	"pentagi/pkg/providers/openai"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func converterPtr[T any](value T) *T { return &value }

func converterByName(models []*model.ModelConfig) map[string]*model.ModelConfig {
	byName := make(map[string]*model.ModelConfig, len(models))
	for _, m := range models {
		byName[m.Name] = m
	}
	return byName
}

func converterRejectsEffortWithTools(info *model.ModelReasoningInfo) bool {
	return info != nil && info.RejectsEffortWithTools != nil && *info.RejectsEffortWithTools
}

func converterTakesNoThinkingDepth(info *model.ModelReasoningInfo) bool {
	return info != nil && info.TakesNoThinkingDepth != nil && *info.TakesNoThinkingDepth
}

func TestConverter_IsAgentTool_RecognisesAgentAndAgentResultTools(t *testing.T) {
	tests := []struct {
		name          string
		functionNames []string
		expected      bool
	}{
		{"a delegation to another agent", []string{"coder", "pentester", "maintenance", "memorist", "search", "advice"}, true},
		{
			"a tool an agent stores its result through",
			[]string{
				"code_result", "hack_result", "maintenance_result", "memorist_result", "search_result",
				"enricher_result", "report_result", "subtask_list", "subtask_patch",
			},
			true,
		},
		{
			"a tool of any other type",
			[]string{
				"terminal", "file", "browser", "google", "duckduckgo", "tavily", "sploitus", "searxng",
				"search_in_memory", "store_guide", "done", "ask",
			},
			false,
		},
		{"a name no tool carries", []string{"unknown_tool"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, functionName := range tt.functionNames {
				if result := isAgentTool(functionName); result != tt.expected {
					t.Errorf("isAgentTool(%q) = %v, want %v", functionName, result, tt.expected)
				}
			}
		})
	}
}

func TestConverter_ConvertModels_MapsTheCatalogueReasoningModeAndLevels(t *testing.T) {
	models := pconfig.ModelsConfig{
		{Name: "ao", Reasoning: &pconfig.ModelReasoningInfo{
			Mode:    pconfig.ModelReasoningAdaptiveOnly,
			Efforts: []llms.ReasoningEffort{llms.ReasoningLow, llms.ReasoningEffort("xhigh")},
		}},
		{Name: "ad", Reasoning: &pconfig.ModelReasoningInfo{Mode: pconfig.ModelReasoningAdaptive}},
		{Name: "bd", Reasoning: &pconfig.ModelReasoningInfo{Mode: pconfig.ModelReasoningBudget}},
		{Name: "none"},
		{Name: "gpt-5.5", Reasoning: &pconfig.ModelReasoningInfo{
			Efforts: []llms.ReasoningEffort{
				llms.ReasoningEffort("none"),
				llms.ReasoningLow,
				llms.ReasoningEffort("xhigh"),
			},
		}},
		{Name: "gpt-5", Reasoning: &pconfig.ModelReasoningInfo{
			Efforts: []llms.ReasoningEffort{
				llms.ReasoningEffort("minimal"),
				llms.ReasoningLow,
			},
		}},
	}

	byName := converterByName(ConvertModels(models, reasoning.ProviderUnknown))

	require.NotNil(t, byName["ao"].Reasoning)
	require.NotNil(t, byName["ao"].Reasoning.Mode)
	assert.Equal(t, model.ModelReasoningModeAdaptiveOnly, *byName["ao"].Reasoning.Mode, "hyphenated pconfig value must map to the GraphQL underscore enum")
	assert.Equal(t, []model.ReasoningEffort{model.ReasoningEffort("low"), model.ReasoningEffort("xhigh")}, byName["ao"].Reasoning.Efforts)

	require.NotNil(t, byName["ad"].Reasoning.Mode)
	assert.Equal(t, model.ModelReasoningModeAdaptive, *byName["ad"].Reasoning.Mode)
	require.NotNil(t, byName["bd"].Reasoning.Mode)
	assert.Equal(t, model.ModelReasoningModeBudget, *byName["bd"].Reasoning.Mode)

	assert.Nil(t, byName["none"].Reasoning, "model without a reasoning descriptor maps to nil")

	assert.Equal(t,
		[]model.ReasoningEffort{model.ReasoningEffortLow, model.ReasoningEffortXhigh},
		byName["gpt-5.5"].Reasoning.Efforts,
		"a catalog value outside the GraphQL enum reaches the form as an unlabeled row",
	)

	assert.Equal(t,
		[]model.ReasoningEffort{model.ReasoningEffortMinimal, model.ReasoningEffortLow},
		byName["gpt-5"].Reasoning.Efforts,
		"minimal is a level the vendor documents, so the contract has to carry it",
	)
}

func TestConverter_ConvertModels_TakesEffortLevelsFromTheLibraryWhenTheCatalogueNamesNone(t *testing.T) {
	thinking := true
	library := llms.ReasoningSupportFor("claude-opus-5", reasoning.ProviderAnthropic).Efforts
	require.NotEmpty(t, library, "the case needs a model whose levels the library knows")

	undeclared := ConvertModels(pconfig.ModelsConfig{{Name: "claude-opus-5", Thinking: &thinking}}, reasoning.ProviderAnthropic)
	require.NotNil(t, undeclared[0].Reasoning)
	want := make([]model.ReasoningEffort, 0, len(library))
	for _, effort := range library {
		want = append(want, model.ReasoningEffort(effort))
	}
	assert.Equal(t, want, undeclared[0].Reasoning.Efforts)

	declared := ConvertModels(pconfig.ModelsConfig{{Name: "claude-opus-5", Reasoning: &pconfig.ModelReasoningInfo{
		Efforts: []llms.ReasoningEffort{llms.ReasoningLow},
	}}}, reasoning.ProviderAnthropic)
	assert.Equal(t, []model.ReasoningEffort{model.ReasoningEffortLow}, declared[0].Reasoning.Efforts,
		"levels the catalogue declares win over the library's")

	unknown := ConvertModels(pconfig.ModelsConfig{{Name: "deepseek-r1:7b", Thinking: &thinking}}, reasoning.ProviderOllama)
	require.NotNil(t, unknown[0].Reasoning)
	assert.Empty(t, unknown[0].Reasoning.Efforts, "no levels where the library names none")

	gptOSS := ConvertModels(pconfig.ModelsConfig{{Name: "gpt-oss:120b", Thinking: &thinking}}, reasoning.ProviderOllama)
	require.NotNil(t, gptOSS[0].Reasoning)
	assert.Equal(t, []model.ReasoningEffort{model.ReasoningEffortLow, model.ReasoningEffortMedium, model.ReasoningEffortHigh},
		gptOSS[0].Reasoning.Efforts, "gpt-oss on the ollama door gets the levels the library names for it")
}

func TestConverter_ConvertModels_CarriesTheLimits(t *testing.T) {
	window, ceiling := 200000, 64000
	byName := converterByName(ConvertModels(pconfig.ModelsConfig{
		{Name: "with-limits", ContextWindow: &window, MaxOutputTokens: &ceiling},
		{Name: "without-limits"},
	}, reasoning.ProviderUnknown))

	require.NotNil(t, byName["with-limits"].ContextWindow, "the form cannot clamp what the contract drops")
	require.NotNil(t, byName["with-limits"].MaxOutputTokens)
	assert.Equal(t, window, *byName["with-limits"].ContextWindow)
	assert.Equal(t, ceiling, *byName["with-limits"].MaxOutputTokens)

	assert.Nil(t, byName["without-limits"].ContextWindow, "an unknown limit stays unknown rather than becoming zero")
	assert.Nil(t, byName["without-limits"].MaxOutputTokens)
}

func TestConverter_ConvertModels_OffersOffOnlyWhereItTakesEffect(t *testing.T) {
	thinks := true
	budget := &pconfig.ModelReasoningInfo{Mode: pconfig.ModelReasoningBudget}
	adaptiveOnly := &pconfig.ModelReasoningInfo{Mode: pconfig.ModelReasoningAdaptiveOnly}
	ollamaDoor := provider.ProviderOllama.ReasoningProvider()

	for _, tt := range []struct {
		name          string
		model         pconfig.ModelConfig
		door          reasoning.Provider
		cannotDisable bool
	}{
		{"gemini switches off through a zero budget", pconfig.ModelConfig{Name: "gemini-2.5-flash", Thinking: &thinks},
			reasoning.ProviderGoogleAI, false},
		{"minimax m3 switches off through the thinking object", pconfig.ModelConfig{Name: "MiniMax-M3", Thinking: &thinks},
			reasoning.ProviderOpenAI, false},
		{"the minimax m2 line takes no disable wire", pconfig.ModelConfig{Name: "MiniMax-M2.5", Thinking: &thinks},
			reasoning.ProviderOpenAI, true},
		{"a model with no disable wire and an unknown default is a silent no-op",
			pconfig.ModelConfig{Name: "glm-5.2", Reasoning: budget}, reasoning.ProviderUnknown, true},
		{"the same model on the openai door sends the thinking object",
			pconfig.ModelConfig{Name: "glm-5.2", Reasoning: budget}, reasoning.ProviderOpenAI, false},
		{"a claude that is off by default switches off by omission",
			pconfig.ModelConfig{Name: "claude-opus-4-8", Reasoning: adaptiveOnly}, reasoning.ProviderAnthropic, false},
		{"an always-on claude rejects off",
			pconfig.ModelConfig{Name: "claude-fable-5", Reasoning: adaptiveOnly}, reasoning.ProviderAnthropic, true},
		{"deepseek-r1 on the ollama door takes think false",
			pconfig.ModelConfig{Name: "deepseek-r1:7b", Reasoning: budget}, ollamaDoor, false},
		{"gpt-oss on the ollama door cannot switch its trace off",
			pconfig.ModelConfig{Name: "gpt-oss:120b", Reasoning: budget}, ollamaDoor, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			info := ConvertModels(pconfig.ModelsConfig{tt.model}, tt.door)[0].Reasoning
			require.NotNil(t, info, "a thinking-capable model must surface its reasoning capability")
			require.NotNil(t, info.CannotDisable)
			assert.Equal(t, tt.cannotDisable, *info.CannotDisable)
		})
	}

	assert.Nil(t, ConvertModels(pconfig.ModelsConfig{{Name: "gpt-4.1"}}, reasoning.ProviderOpenAI)[0].Reasoning,
		"a model that does not think offers no reasoning control")
}

func TestConverter_ConvertModels_KeepsEachRequestRuleToTheDoorThatEnforcesIt(t *testing.T) {
	thinks, silent := true, false
	gateway := provider.ProviderCustom.ReasoningProvider()

	for _, tt := range []struct {
		name          string
		model         pconfig.ModelConfig
		door          reasoning.Provider
		controls      bool
		refusesTools  bool
		takesNoBudget bool
	}{
		{"a refusing generation behind a gateway rides the openai door",
			pconfig.ModelConfig{Name: "openai/gpt-5.6-sol", Thinking: &thinks}, gateway, true, true, false},
		{"the same model on another door has no tool rule",
			pconfig.ModelConfig{Name: "openai/gpt-5.6-sol", Thinking: &thinks}, reasoning.ProviderOllama, true, false, false},
		{"a dated snapshot of a refusing generation listed without reasoning facts",
			pconfig.ModelConfig{Name: "gpt-5.5-2026-04-23"}, gateway, true, true, false},
		{"a gateway alias of a refusing generation listed without reasoning facts",
			pconfig.ModelConfig{Name: "openai/gpt-5.6-sol-high"}, gateway, true, true, false},
		{"a model outside the rule listed without reasoning facts keeps no controls",
			pconfig.ModelConfig{Name: "gpt-4.1-mini"}, gateway, false, false, false},
		{"a listing that says the model does not think is kept",
			pconfig.ModelConfig{Name: "gpt-5.4-legacy", Thinking: &silent}, gateway, false, false, false},
		{"a model that takes no thinking budget on the openai door",
			pconfig.ModelConfig{Name: "glm-5-turbo", Thinking: &thinks}, reasoning.ProviderOpenAI, true, false, true},
		{"the same model on another door has no budget rule",
			pconfig.ModelConfig{Name: "glm-5-turbo", Thinking: &thinks}, reasoning.ProviderOllama, true, false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			info := ConvertModels(pconfig.ModelsConfig{tt.model}, tt.door)[0].Reasoning
			if !tt.controls {
				assert.Nil(t, info)
				return
			}
			require.NotNil(t, info)
			require.NotNil(t, info.RejectsEffortWithTools)
			require.NotNil(t, info.TakesNoThinkingDepth)
			assert.Equal(t, tt.refusesTools, *info.RejectsEffortWithTools, "rejects an effort with tools")
			assert.Equal(t, tt.takesNoBudget, *info.TakesNoThinkingDepth, "takes no thinking budget")
		})
	}
}

func TestConverter_ConvertModelsWithPrefix_NamesTheShippedModelsEachRequestRuleRefuses(t *testing.T) {
	for _, tt := range []struct {
		name      string
		catalogue func() (pconfig.ModelsConfig, error)
		prvtype   provider.ProviderType
		prefix    string
		rule      func(*model.ModelReasoningInfo) bool
		refuse    []string
		take      []string
	}{
		{"openai models that refuse a level with tools", openai.DefaultModels, provider.ProviderOpenAI, "", converterRejectsEffortWithTools,
			[]string{"gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna", "gpt-5.5", "gpt-5.4", "gpt-5.4-mini", "gpt-5.4-nano"},
			[]string{"gpt-5.2", "gpt-5.1", "gpt-5", "o3"}},
		{"glm models that take no thinking budget", glm.DefaultModels, provider.ProviderGLM, "", converterTakesNoThinkingDepth,
			[]string{"glm-5", "glm-4.5-air", "glm-5.1"}, []string{"glm-5.2"}},
		{"glm models behind zai that take no thinking budget", glm.DefaultModels, provider.ProviderGLM, "zai", converterTakesNoThinkingDepth,
			[]string{"glm-5", "glm-4.5-air"}, []string{"glm-5.2"}},
		{"glm models behind dashscope all take a thinking budget", glm.DefaultModels, provider.ProviderGLM, "dashscope", converterTakesNoThinkingDepth,
			nil, []string{"glm-5", "glm-4.5-air", "glm-5.2"}},
		{"kimi models behind moonshot that take no thinking budget", kimi.DefaultModels, provider.ProviderKimi, "moonshot", converterTakesNoThinkingDepth,
			[]string{"kimi-k2.6", "kimi-k2.7-code", "kimi-k2.7-code-highspeed"}, []string{"kimi-k3"}},
		{"minimax models that take no thinking budget", minimax.DefaultModels, provider.ProviderMiniMax, "", converterTakesNoThinkingDepth,
			[]string{"MiniMax-M3", "MiniMax-M2.7"}, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			catalogue, err := tt.catalogue()
			require.NoError(t, err)

			byName := converterByName(ConvertModelsWithPrefix(catalogue, tt.prvtype.ReasoningProvider(), tt.prefix))
			for _, name := range tt.refuse {
				require.Contains(t, byName, name)
				assert.True(t, tt.rule(byName[name].Reasoning), name)
			}
			for _, name := range tt.take {
				require.Contains(t, byName, name)
				assert.False(t, tt.rule(byName[name].Reasoning), name)
			}
		})
	}
}

func TestConverter_ConvertModels_OffersTheEffortsTheCatalogueDeclares(t *testing.T) {
	models, err := openai.DefaultModels()
	require.NoError(t, err)

	declared := map[string][]model.ReasoningEffort{}
	for _, m := range models {
		if m.Reasoning == nil || len(m.Reasoning.Efforts) == 0 {
			continue
		}
		efforts := make([]model.ReasoningEffort, 0, len(m.Reasoning.Efforts))
		for _, effort := range m.Reasoning.Efforts {
			if level := model.ReasoningEffort(effort); level.IsValid() {
				efforts = append(efforts, level)
			}
		}
		declared[m.Name] = efforts
	}
	require.NotEmpty(t, declared, "the openai catalogue must declare efforts for at least one model")

	var checked, differs int
	for _, converted := range ConvertModels(models, reasoning.ProviderOpenAI) {
		want, ok := declared[converted.Name]
		if !ok {
			continue
		}
		checked++
		require.NotNil(t, converted.Reasoning, "%s declares efforts but reaches the interface without a reasoning block", converted.Name)
		assert.Equal(t, want, converted.Reasoning.Efforts, "%s must offer the efforts models.yml declares, not the langchaingo hint", converted.Name)

		var hint []model.ReasoningEffort
		for _, effort := range llms.ReasoningSupportFor(converted.Name, reasoning.ProviderOpenAI).Efforts {
			hint = append(hint, model.ReasoningEffort(effort))
		}
		if !slices.Equal(hint, want) {
			differs++
		}
	}

	assert.Equal(t, len(declared), checked, "every catalogue model with efforts must reach the interface")
	assert.NotZero(t, differs, "vacuous guard: the langchaingo hint agrees with the catalogue on every model, so this test cannot tell the two sources apart")
}

func TestConverter_ConvertModels_OffersNoReasoningControlTheBedrockDoorCannotFill(t *testing.T) {
	models, err := bedrock.DefaultModels()
	require.NoError(t, err)

	converted := converterByName(ConvertModels(models, provider.ProviderBedrock.ReasoningProvider()))

	// qwen3-32b returns no reasoningContent on Bedrock under any request shape; AWS documents no reasoning field for the GLMs.
	for _, name := range []string{"qwen.qwen3-32b-v1:0", "zai.glm-4.7", "zai.glm-4.7-flash", "zai.glm-5"} {
		mc, ok := converted[name]
		if !ok {
			t.Errorf("%s is not in the bedrock catalogue", name)
			continue
		}
		assert.Nil(t, mc.Reasoning, "%s: the door sends no reasoning for it, so no reasoning control may be offered", name)
		if assert.NotNil(t, mc.Thinking, "%s must say whether it thinks", name) {
			assert.False(t, *mc.Thinking, "%s must not be listed as a thinking model on Bedrock", name)
		}
	}
}

func TestConverter_ConvertModels_OffersDeepSeekFlashAReasoningControlWithOff(t *testing.T) {
	models, err := deepseek.DefaultModels()
	require.NoError(t, err)

	mc, ok := converterByName(ConvertModels(models, provider.ProviderDeepSeek.ReasoningProvider()))["deepseek-flash"]
	require.True(t, ok, "deepseek-flash is not in the deepseek catalogue")

	require.NotNil(t, mc.Reasoning, "deepseek-flash thinks by default, so a reasoning control must be offered")
	require.NotNil(t, mc.Reasoning.Supported)
	require.NotNil(t, mc.Reasoning.CannotDisable)
	assert.True(t, *mc.Reasoning.Supported, "deepseek-flash must be reported as a reasoning model")
	assert.False(t, *mc.Reasoning.CannotDisable,
		"the thinking object disables deepseek-flash, so Off must be offered")
}

// The interface offers a thinking control only for a model whose converted entry carries reasoning information.
func TestConverter_ConvertModels_ReportsEveryThinkingClaudeAsReasoning(t *testing.T) {
	models, err := anthropic.DefaultModels()
	require.NoError(t, err)

	converted := ConvertModels(models, reasoning.ProviderAnthropic)
	require.NotEmpty(t, converted, "the catalogue converted to nothing")

	seen := make(map[string]bool, len(converted))
	for _, m := range converted {
		seen[m.Name] = true

		if !reasoning.ClaudeSupportsThinking(m.Name) {
			continue
		}
		if m.Reasoning == nil {
			t.Errorf("%s reaches the interface with no reasoning information, so no thinking control is offered for it", m.Name)
			continue
		}
		if m.Reasoning.Supported == nil || !*m.Reasoning.Supported {
			t.Errorf("%s reaches the interface reported as not supporting reasoning", m.Name)
		}
	}

	// Without naming the model the guard was written for, a catalogue that lost it would still pass.
	assert.True(t, seen["claude-haiku-4-5"], "claude-haiku-4-5 is not in the anthropic catalogue any more")
}

func TestConverter_OffEffective_ClassifiesEveryOffWire(t *testing.T) {
	classified := map[reasoning.OffWire]bool{
		reasoning.OffOmit:                  false,
		reasoning.OffDisableClaude:         true,
		reasoning.OffZeroBudget:            true,
		reasoning.OffMinimalLevel:          true,
		reasoning.OffEffortNone:            true,
		reasoning.OffDisableDashScope:      true,
		reasoning.OffDisableThinkingObject: true,
		reasoning.OffDisableThinkBool:      true,
		reasoning.OffUnsupported:           false,
	}

	require.Len(t, classified, int(reasoning.OffUnsupported)+1,
		"the fork declares an OffWire this list does not: classify it in offEffective, "+
			"otherwise it falls to the default branch and hides Off on models that can disable")

	for wire, want := range classified {
		assert.Equal(t, want, offEffective(wire, nil), "wire %d with an unknown default", int(wire))
	}

	assert.True(t, offEffective(reasoning.OffOmit, converterPtr(false)), "omitting the wire switches off a model that is off by default")
	assert.False(t, offEffective(reasoning.OffOmit, converterPtr(true)), "omitting the wire leaves a default-on model thinking")
}

func TestConverter_ConvertAgentConfigFromGqlModel_StoresWhatTheCardSetsAndReadsItBack(t *testing.T) {
	extraBody := map[string]any{
		"chat_template_kwargs": map[string]any{"enable_thinking": false},
		"provider":             "vllm",
	}
	callOptions := model.AgentConfig{
		Model:            "m",
		MinP:             converterPtr(0.05),
		N:                converterPtr(3),
		JSON:             converterPtr(true),
		ResponseMimeType: converterPtr("application/json"),
	}
	chosenZeros := model.AgentConfig{Model: "m", MaxTokens: converterPtr(0), Temperature: converterPtr(0.0), TopP: converterPtr(0.5)}

	for _, tt := range []struct {
		name string
		card model.AgentConfig
		wire llms.CallOptions
		back model.AgentConfig
	}{
		{
			name: "an extra body survives both ways",
			card: model.AgentConfig{Model: "m", ExtraBody: extraBody},
			wire: llms.CallOptions{Model: converterPtr("m"), ExtraBody: extraBody},
			back: model.AgentConfig{Model: "m", ExtraBody: extraBody},
		},
		{
			name: "no extra body stays absent rather than empty",
			card: model.AgentConfig{Model: "m"},
			wire: llms.CallOptions{Model: converterPtr("m")},
			back: model.AgentConfig{Model: "m"},
		},
		{
			name: "call option fields survive both ways",
			card: callOptions,
			wire: llms.CallOptions{Model: converterPtr("m"), MinP: converterPtr(0.05), N: converterPtr(3), JSONMode: true,
				ResponseMIMEType: converterPtr("application/json")},
			back: callOptions,
		},
		{
			// the wire gates these on the key being present, so a stored zero would switch json mode on
			name: "a zero call option field is not stored",
			card: model.AgentConfig{Model: "m", MinP: converterPtr(0.0), N: converterPtr(0), JSON: converterPtr(false),
				ResponseMimeType: converterPtr("")},
			wire: llms.CallOptions{Model: converterPtr("m")},
			back: model.AgentConfig{Model: "m"},
		},
		{
			name: "a zero the user chose for a sampling field is kept",
			card: chosenZeros,
			wire: llms.CallOptions{Model: converterPtr("m"), MaxTokens: converterPtr(0), Temperature: converterPtr(0.0),
				TopP: converterPtr(0.5)},
			back: chosenZeros,
		},
		{
			name: "a field the user never set stays absent",
			card: model.AgentConfig{Model: "m", TopP: converterPtr(0.5)},
			wire: llms.CallOptions{Model: converterPtr("m"), TopP: converterPtr(0.5)},
			back: model.AgentConfig{Model: "m", TopP: converterPtr(0.5)},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			card := tt.card
			stored := ConvertAgentConfigFromGqlModel(&card)
			require.NotNil(t, stored)

			var wire llms.CallOptions
			for _, option := range stored.BuildOptions() {
				option(&wire)
			}
			assert.Equal(t, tt.wire, wire, "what the stored config sends")
			assert.Equal(t, &tt.back, ConvertAgentConfigToGqlModel(stored), "what the card reads back")
		})
	}
}

func TestConverter_ConvertAgentConfigToGqlModel_ExposesOnlyTheModelOfAConfigNeverDecoded(t *testing.T) {
	back := ConvertAgentConfigToGqlModel(&pconfig.AgentConfig{Model: "m", Temperature: 0.7})
	require.NotNil(t, back)

	assert.Equal(t, "m", back.Model)
	assert.Nil(t, back.Temperature, "a config that was never decoded remembers no keys")
}

// The shipped simple_json default carries `json: true`, which a save through the UI must keep.
func TestConverter_ConvertAgentsConfigFromGqlModel_KeepsTheShippedSimpleJSONMode(t *testing.T) {
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

func TestConverter_APITokenStatus_ReportsAnOutlivedActiveTokenAsExpired(t *testing.T) {
	tests := []struct {
		name   string
		status database.TokenStatus
		ttl    int64
		age    time.Duration
		want   model.TokenStatus
	}{
		{name: "young active token", status: database.TokenStatusActive, ttl: 3600, age: time.Minute, want: model.TokenStatusActive},
		{name: "outlived active token", status: database.TokenStatusActive, ttl: 60, age: 2 * time.Minute, want: model.TokenStatusExpired},
		{name: "a second past the ttl", status: database.TokenStatusActive, ttl: 60, age: 61 * time.Second, want: model.TokenStatusExpired},
		{name: "revoked stays revoked", status: database.TokenStatusRevoked, ttl: 60, age: 2 * time.Minute, want: model.TokenStatusRevoked},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := database.ApiToken{
				Status:    tt.status,
				Ttl:       tt.ttl,
				CreatedAt: sql.NullTime{Time: time.Now().Add(-tt.age), Valid: true},
			}
			withSecret := database.APITokenWithSecret{ApiToken: token, Token: "secret"}

			assert.Equal(t, tt.want, ConvertAPIToken(token).Status, "ConvertAPIToken")
			assert.Equal(t, tt.want, ConvertAPITokenRemoveSecret(withSecret).Status, "ConvertAPITokenRemoveSecret")
			assert.Equal(t, tt.want, ConvertAPITokenWithSecret(withSecret).Status, "ConvertAPITokenWithSecret")
		})
	}
}
