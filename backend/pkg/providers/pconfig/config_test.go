package pconfig

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/openai"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"gopkg.in/yaml.v3"
)

func configPtr[T any](value T) *T { return &value }

func configResolve(options []llms.CallOption) llms.CallOptions {
	var resolved llms.CallOptions
	for _, option := range options {
		option(&resolved)
	}
	return resolved
}

// configAgentFields reads the agent slots off the struct: a hardcoded list would defeat the comparison with AllAgentTypes.
func configAgentFields(t *testing.T) map[ProviderOptionsType]string {
	t.Helper()

	agentConfigType := reflect.TypeOf(&AgentConfig{})
	structType := reflect.TypeOf(ProviderConfig{})

	found := make(map[ProviderOptionsType]string)
	for i := range structType.NumField() {
		field := structType.Field(i)
		if field.Type != agentConfigType {
			continue
		}

		tag, _, _ := strings.Cut(field.Tag.Get("yaml"), ",")
		if tag == "" {
			t.Fatalf("agent field %s carries no yaml name", field.Name)
		}
		found[ProviderOptionsType(tag)] = field.Name
	}

	if len(found) == 0 {
		t.Fatal("no agent fields found on ProviderConfig")
	}
	return found
}

var configAgentCallsCarryTools = map[ProviderOptionsType]bool{
	OptionsTypeSimple:       true,
	OptionsTypeSimpleJSON:   false,
	OptionsTypePrimaryAgent: true,
	OptionsTypeAssistant:    true,
	OptionsTypeGenerator:    true,
	OptionsTypeRefiner:      true,
	OptionsTypeAdviser:      false,
	OptionsTypeReflector:    false,
	OptionsTypeSearcher:     true,
	OptionsTypeEnricher:     true,
	OptionsTypeCoder:        true,
	OptionsTypeInstaller:    true,
	OptionsTypePentester:    true,
}

// configChatServer answers every chat completion, so an error from the door is a refusal raised before sending.
func configChatServer(t *testing.T) string {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"x","object":"chat.completion","created":1,"model":"m",`+
			`"choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],`+
			`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`)
	}))
	t.Cleanup(srv.Close)

	return srv.URL
}

func configCallTheDoor(t *testing.T, url, model string, options []llms.CallOption) error {
	t.Helper()

	llm, err := openai.New(openai.WithBaseURL(url), openai.WithToken("test"), openai.WithModel(model))
	require.NoError(t, err)

	_, err = llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, options...)
	return err
}

func TestConfig_NewCallUsage_ReadsTheReportedTokensAndCostsAndIsZeroWithoutThem(t *testing.T) {
	for _, tt := range []struct {
		name string
		info map[string]any
		want CallUsage
	}{
		{name: "prompt and completion tokens as int", info: map[string]any{"PromptTokens": 100, "CompletionTokens": 50},
			want: CallUsage{Input: 100, Output: 50}},
		{name: "prompt and completion tokens as int32",
			info: map[string]any{"PromptTokens": int32(100), "CompletionTokens": int32(50)},
			want: CallUsage{Input: 100, Output: 50}},
		{
			name: "cache counts and upstream costs as int64 and float64",
			info: map[string]any{"PromptTokens": int64(1200), "CompletionTokens": float64(300),
				"CacheReadInputTokens": int64(800), "CacheCreationInputTokens": int64(100),
				"UpstreamInferencePromptCost": 0.0012, "UpstreamInferenceCompletionsCost": 0.0045},
			want: CallUsage{Input: 1200, Output: 300, CacheRead: 800, CacheWrite: 100, CostInput: 0.0012, CostOutput: 0.0045},
		},
		{name: "an empty report", info: map[string]any{}},
		{name: "no report at all"},
		{name: "tokens under other names", info: map[string]any{"InputTokens": 100, "OutputTokens": 50}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			usage := NewCallUsage(tt.info)
			assert.Equal(t, tt.want, usage)
			assert.Equal(t, tt.want == (CallUsage{}), usage.IsZero(), "IsZero")
		})
	}

	t.Run("any one count or cost alone is not zero", func(t *testing.T) {
		for key, value := range map[string]any{"PromptTokens": 1, "CompletionTokens": 1, "CacheReadInputTokens": 1,
			"CacheCreationInputTokens": 1, "UpstreamInferencePromptCost": 0.5, "UpstreamInferenceCompletionsCost": 0.5} {
			usage := NewCallUsage(map[string]any{key: value})
			assert.False(t, usage.IsZero(), "%s alone", key)
		}
	})
}

func TestConfig_AllAgentTypes_MatchesTheConfigStruct(t *testing.T) {
	fields := configAgentFields(t)

	listed := make(map[ProviderOptionsType]bool, len(AllAgentTypes))
	for _, agentType := range AllAgentTypes {
		listed[agentType] = true
	}

	for agentType, fieldName := range fields {
		if !listed[agentType] {
			t.Errorf("ProviderConfig.%s configures agent %q, which AllAgentTypes does not list", fieldName, agentType)
		}
	}
	for agentType := range listed {
		if _, ok := fields[agentType]; !ok {
			t.Errorf("AllAgentTypes lists %q, which is not a field of ProviderConfig", agentType)
		}
	}
}

func TestConfig_AgentConfigForType_ResolvesEveryAgentToItsOwnConfig(t *testing.T) {
	fields := configAgentFields(t)

	config := &ProviderConfig{}
	value := reflect.ValueOf(config).Elem()
	for agentType, fieldName := range fields {
		value.FieldByName(fieldName).Set(reflect.ValueOf(&AgentConfig{Model: string(agentType) + "-model"}))
	}

	for _, agentType := range AllAgentTypes {
		agentConfig := config.AgentConfigForType(agentType)
		if agentConfig == nil {
			t.Errorf("agent %q resolves to no config at all, so it would run unconfigured", agentType)
			continue
		}
		if want := string(agentType) + "-model"; agentConfig.Model != want {
			t.Errorf("agent %q resolved to model %q, want %q", agentType, agentConfig.Model, want)
		}
	}
}

func TestConfig_AgentConfigForType_FallsBackToTheStandInOfAnUnsetAgent(t *testing.T) {
	config := &ProviderConfig{
		Simple:       &AgentConfig{Model: "simple-model"},
		PrimaryAgent: &AgentConfig{Model: "primary-model"},
	}

	for _, tt := range []struct {
		agentType ProviderOptionsType
		want      string
	}{
		{OptionsTypeSimpleJSON, "simple-model"},
		{OptionsTypeAssistant, "primary-model"},
	} {
		agentConfig := config.AgentConfigForType(tt.agentType)
		if agentConfig == nil {
			t.Errorf("unset agent %q resolves to no config", tt.agentType)
			continue
		}
		if agentConfig.Model != tt.want {
			t.Errorf("unset agent %q resolved to %q, want %q", tt.agentType, agentConfig.Model, tt.want)
		}
	}
}

func TestConfig_UsesTools_NamesEveryAgentWhoseCallsCarryTools(t *testing.T) {
	require.Len(t, configAgentCallsCarryTools, len(AllAgentTypes))

	for _, opt := range AllAgentTypes {
		want, ok := configAgentCallsCarryTools[opt]
		require.True(t, ok, "%s has no expected answer", opt)
		assert.Equal(t, want, opt.UsesTools(), "%s", opt)
	}
}

func TestConfig_Validate_RefusesAValueNoProviderAccepts(t *testing.T) {
	for i, tt := range []struct {
		name  string
		agent string
		want  string
	}{
		{"every bounded value at the top of its range", `{"temperature":2,"top_p":1,"min_length":100,"max_length":100,` +
			`"repetition_penalty":2,"frequency_penalty":2,"presence_penalty":2,"reasoning":{"max_tokens":32768}}`, ""},
		{"every value at the bottom of its range", `{"temperature":0,"top_p":0,"top_k":0,"max_tokens":0,` +
			`"min_length":0,"max_length":0,"repetition_penalty":0,"frequency_penalty":-2,"presence_penalty":-2,` +
			`"reasoning":{"max_tokens":0},"price":{"input":0,"output":0,"cache_read":0,"cache_write":0}}`, ""},
		{"every value inside its range", `{"temperature":0.7,"top_p":0.9,"top_k":40,"max_tokens":8192,"min_length":1,` +
			`"max_length":100,"repetition_penalty":1.1,"frequency_penalty":-0.5,"presence_penalty":0.5,` +
			`"reasoning":{"max_tokens":4096},"price":{"input":5,"output":25,"cache_read":1,"cache_write":1}}`, ""},
		{"a budget mode with its budget", `{"reasoning":{"mode":"budget","max_tokens":4096}}`, ""},
		{"an off mode without a budget", `{"reasoning":{"mode":"off"}}`, ""},
		{"an adaptive mode without a budget", `{"reasoning":{"mode":"adaptive"}}`, ""},
		{"a named model", `{"model":"gpt-4o"}`, ""},
		{"an empty model", `{"model":""}`, ""},
		{"a temperature above two", `{"temperature":5}`, "temperature 5 out of range [0, 2]"},
		{"a negative temperature", `{"temperature":-1}`, "temperature -1 out of range [0, 2]"},
		{"a top_p above one", `{"top_p":1.5}`, "top_p 1.5 out of range [0, 1]"},
		{"a negative top_k", `{"top_k":-1}`, "top_k -1 must be >= 0"},
		{"a negative max_tokens", `{"max_tokens":-100}`, "max_tokens -100 must be >= 0"},
		{"an inverted length window", `{"min_length":5000,"max_length":10}`, "min_length 5000 must not exceed max_length 10"},
		{"a frequency penalty out of range", `{"frequency_penalty":9}`, "frequency_penalty 9 out of range [-2, 2]"},
		{"a presence penalty out of range", `{"presence_penalty":-3}`, "presence_penalty -3 out of range [-2, 2]"},
		{"a repetition penalty out of range", `{"repetition_penalty":3}`, "repetition_penalty 3 out of range [0, 2]"},
		{"a reasoning budget over the cap", `{"reasoning":{"max_tokens":40000}}`, "reasoning.max_tokens 40000 out of range [0, 32768]"},
		{"a negative reasoning budget", `{"reasoning":{"max_tokens":-1}}`, "reasoning.max_tokens -1 out of range [0, 32768]"},
		{"a budget mode without a budget", `{"reasoning":{"mode":"budget"}}`,
			`reasoning.max_tokens is required when reasoning.mode is "budget"`},
		{"a budget mode with an effort but no budget", `{"reasoning":{"mode":"budget","effort":"low"}}`,
			`reasoning.max_tokens is required when reasoning.mode is "budget"`},
		{"a negative price", `{"price":{"input":-1}}`, "price values must be >= 0"},
		{"a negative cache read price", `{"price":{"cache_read":-0.1}}`, "price values must be >= 0"},
		{"a model of spaces", `{"model":"   "}`, "model must not consist of whitespace only"},
		{"a model of a tab", `{"model":"\t"}`, "model must not consist of whitespace only"},
		{"a model of a newline", `{"model":"\n"}`, "model must not consist of whitespace only"},
	} {
		// Rows rotate through the agents so the name each agent is reported under is checked.
		agent := AllAgentTypes[i%len(AllAgentTypes)]
		t.Run(tt.name, func(t *testing.T) {
			var pc ProviderConfig
			require.NoError(t, json.Unmarshal([]byte(fmt.Sprintf(`{%q:%s}`, agent, tt.agent)), &pc))

			err := pc.Validate()
			if tt.want == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, string(agent)+": "+tt.want)
		})
	}
}

func TestConfig_EffectiveMode_DerivesTheModeTheBlockAsksFor(t *testing.T) {
	tests := []struct {
		name string
		rc   ReasoningConfig
		want ReasoningMode
	}{
		{"an empty block defers to the model", ReasoningConfig{}, ReasoningModeDefault},
		{"an adaptive mode is kept", ReasoningConfig{Mode: ReasoningModeAdaptive}, ReasoningModeAdaptive},
		{"a budget mode is kept", ReasoningConfig{Mode: ReasoningModeBudget}, ReasoningModeBudget},
		{"an off mode is kept", ReasoningConfig{Mode: ReasoningModeOff}, ReasoningModeOff},
		{"an off mode wins over a budget", ReasoningConfig{Mode: ReasoningModeOff, MaxTokens: 5000}, ReasoningModeOff},
		{"an effort alone defers to the model", ReasoningConfig{Effort: llms.ReasoningHigh}, ReasoningModeDefault},
		{"a budget alone asks for budget mode", ReasoningConfig{MaxTokens: 5000}, ReasoningModeBudget},
		{"a budget under the none effort asks for budget mode", ReasoningConfig{Effort: llms.ReasoningEffort("none"), MaxTokens: 5000}, ReasoningModeBudget},
		{"a budget beside another effort defers to the model", ReasoningConfig{Effort: llms.ReasoningHigh, MaxTokens: 5000}, ReasoningModeDefault},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.rc.EffectiveMode())
		})
	}
}

func TestConfig_IsZero_HoldsOnlyForAnEmptyBlock(t *testing.T) {
	assert.True(t, ReasoningConfig{}.IsZero())
	assert.False(t, ReasoningConfig{Mode: ReasoningModeAdaptive}.IsZero())
	assert.False(t, ReasoningConfig{Mode: ReasoningModeBudget}.IsZero())
	assert.False(t, ReasoningConfig{Mode: ReasoningModeOff}.IsZero())
	assert.False(t, ReasoningConfig{Effort: llms.ReasoningLow}.IsZero())
	assert.False(t, ReasoningConfig{MaxTokens: 1}.IsZero())
}

func TestConfig_AgentConfig_DecodesEitherFormatAndStoresWhatItRead(t *testing.T) {
	for _, tt := range []struct {
		name   string
		json   string
		yaml   string
		want   AgentConfig
		stored string // the document the decoded config marshals back to
		err    string
	}{
		{name: "an empty block", json: `{}`, yaml: `{}`, stored: `{}`},
		{
			name:   "zero values are remembered as set",
			json:   `{"model":"","max_tokens":0,"temperature":0,"top_k":0,"top_p":0}`,
			yaml:   "model: \"\"\nmax_tokens: 0\ntemperature: 0\ntop_k: 0\ntop_p: 0\n",
			stored: `{"model":"","max_tokens":0,"temperature":0,"top_k":0,"top_p":0}`,
		},
		{
			name:   "values",
			json:   `{"model":"test-model","max_tokens":100,"temperature":0.7}`,
			yaml:   "model: test-model\nmax_tokens: 100\ntemperature: 0.7\n",
			want:   AgentConfig{Model: "test-model", MaxTokens: 100, Temperature: 0.7},
			stored: `{"model":"test-model","max_tokens":100,"temperature":0.7}`,
		},
		{
			name:   "a reasoning block",
			json:   `{"reasoning":{"mode":"adaptive","effort":"xhigh","max_tokens":2000}}`,
			yaml:   "reasoning:\n  mode: adaptive\n  effort: xhigh\n  max_tokens: 2000\n",
			want:   AgentConfig{Reasoning: ReasoningConfig{Mode: ReasoningModeAdaptive, Effort: llms.ReasoningXHigh, MaxTokens: 2000}},
			stored: `{"reasoning":{"mode":"adaptive","effort":"xhigh","max_tokens":2000}}`,
		},
		{
			// YAML 1.1 would read an unquoted off as false; yaml.v3 keeps the string.
			name:   "an unquoted off stays the mode",
			json:   `{"reasoning":{"mode":"off"}}`,
			yaml:   "reasoning:\n  mode: off\n",
			want:   AgentConfig{Reasoning: ReasoningConfig{Mode: ReasoningModeOff}},
			stored: `{"reasoning":{"mode":"off"}}`,
		},
		{
			name:   "a model padded at both ends is trimmed",
			json:   `{"model":"  gpt-5  "}`,
			yaml:   "model: \"  gpt-5  \"\n",
			want:   AgentConfig{Model: "gpt-5"},
			stored: `{"model":"gpt-5"}`,
		},
		{
			name:   "a model with a trailing newline is trimmed",
			json:   `{"model":"gpt-5\n"}`,
			yaml:   "model: \"gpt-5\\n\"\n",
			want:   AgentConfig{Model: "gpt-5"},
			stored: `{"model":"gpt-5"}`,
		},
		{
			name:   "a model of whitespace only is kept for Validate to refuse",
			json:   `{"model":"   "}`,
			yaml:   "model: \"   \"\n",
			want:   AgentConfig{Model: "   "},
			stored: `{"model":"   "}`,
		},
		{name: "a malformed json document", json: `{invalid}`, err: "invalid character 'i'"},
		{name: "a malformed yaml document", yaml: "invalid: [yaml", err: "did not find expected ',' or ']'"},
	} {
		for _, codec := range []struct {
			format string
			doc    string
			decode func([]byte, any) error
		}{
			{"json", tt.json, json.Unmarshal},
			{"yaml", tt.yaml, yaml.Unmarshal},
		} {
			if codec.doc == "" {
				continue
			}
			t.Run(tt.name+" in "+codec.format, func(t *testing.T) {
				var got AgentConfig
				err := codec.decode([]byte(codec.doc), &got)
				if tt.err != "" {
					assert.ErrorContains(t, err, tt.err)
					return
				}
				require.NoError(t, err)

				stored, err := json.Marshal(&got)
				require.NoError(t, err)
				assert.JSONEq(t, tt.stored, string(stored))

				got.ClearRaw()
				assert.Equal(t, tt.want, got)
			})
		}
	}
}

func TestConfig_AgentConfig_EncodesTheStoredKeysOrTheSetFields(t *testing.T) {
	for _, tt := range []struct {
		name   string
		config *AgentConfig
		want   string
	}{
		{"no config", nil, `null`},
		{"a config never decoded and left empty", &AgentConfig{}, `{}`},
		{
			"a config never decoded writes its set fields",
			&AgentConfig{Model: "test-model", MaxTokens: 100, Temperature: 0.7},
			`{"max_tokens":100,"model":"test-model","temperature":0.7}`,
		},
		{
			"a decoded config writes back the keys it read",
			&AgentConfig{
				Model:       "test-model",
				MaxTokens:   100,
				Temperature: 0.7,
				Reasoning:   ReasoningConfig{Effort: llms.ReasoningMedium, MaxTokens: 5000},
				raw: map[string]any{
					"model":        "test-model",
					"max_tokens":   100,
					"temperature":  0.7,
					"custom_field": "custom_value",
					"reasoning":    map[string]any{"effort": "medium", "max_tokens": 5000},
				},
			},
			`{"custom_field":"custom_value","max_tokens":100,"model":"test-model",` +
				`"reasoning":{"effort":"medium","max_tokens":5000},"temperature":0.7}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			asJSON, err := json.Marshal(tt.config)
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, string(asJSON), "json")

			asYAML, err := yaml.Marshal(tt.config)
			require.NoError(t, err)
			var doc any
			require.NoError(t, yaml.Unmarshal(asYAML, &doc))
			sameDoc, err := json.Marshal(doc)
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, string(sameDoc), "yaml")
		})
	}
}

func TestConfig_BuildOptions_EmitsOnlyTheKeysTheBlockCarries(t *testing.T) {
	every := llms.CallOptions{
		Model: configPtr("test-model"), MaxTokens: configPtr(100), Temperature: configPtr(0.7),
		TopK: configPtr(10), TopP: configPtr(0.9), MinP: configPtr(0.05), N: configPtr(2),
		MinLength: configPtr(10), MaxLength: configPtr(100), RepetitionPenalty: configPtr(1.1),
		FrequencyPenalty: configPtr(1.2), PresencePenalty: configPtr(1.3), JSONMode: true,
		ResponseMIMEType: configPtr("application/json"),
	}
	modelAnd := func(reasoning *llms.ReasoningConfig) llms.CallOptions {
		return llms.CallOptions{Model: configPtr("test-model"), Temperature: configPtr(0.7), Reasoning: reasoning}
	}
	withReasoning := func(reasoning string) string {
		return `{"model":"test-model","temperature":0.7,"reasoning":` + reasoning + `}`
	}

	for _, tt := range []struct {
		name  string
		json  string
		yaml  string // decoded instead of json when set
		count int
		want  llms.CallOptions
	}{
		{name: "no block at all"},
		{name: "an empty block", json: `{}`},
		{
			name:  "zero values are sent because their keys are set, an empty model is not",
			json:  `{"model":"","max_tokens":0,"temperature":0,"top_k":0,"top_p":0}`,
			count: 4,
			want:  llms.CallOptions{MaxTokens: configPtr(0), Temperature: configPtr(0.0), TopK: configPtr(0), TopP: configPtr(0.0)},
		},
		{
			name: "every sampling key in json",
			json: `{"model":"test-model","max_tokens":100,"temperature":0.7,"top_k":10,"top_p":0.9,"min_p":0.05,"n":2,` +
				`"min_length":10,"max_length":100,"repetition_penalty":1.1,"frequency_penalty":1.2,"presence_penalty":1.3,` +
				`"json":true,"response_mime_type":"application/json"}`,
			count: 14,
			want:  every,
		},
		{
			name: "every sampling key in yaml",
			yaml: "model: test-model\nmax_tokens: 100\ntemperature: 0.7\ntop_k: 10\ntop_p: 0.9\nmin_p: 0.05\nn: 2\n" +
				"min_length: 10\nmax_length: 100\nrepetition_penalty: 1.1\nfrequency_penalty: 1.2\npresence_penalty: 1.3\n" +
				"json: true\nresponse_mime_type: application/json\n",
			count: 14,
			want:  every,
		},
		{
			name:  "an extra body is passed through",
			json:  `{"model":"test-model","extra_body":{"chat_template_kwargs":{"enable_thinking":false}}}`,
			count: 2,
			want: llms.CallOptions{Model: configPtr("test-model"), ExtraBody: map[string]any{
				"chat_template_kwargs": map[string]any{"enable_thinking": false},
			}},
		},
		{
			name:  "an effort with a budget beside it sends the effort alone",
			json:  withReasoning(`{"effort":"low","max_tokens":1000}`),
			count: 3,
			want:  modelAnd(&llms.ReasoningConfig{Effort: llms.ReasoningLow}),
		},
		{name: "a minimal effort", json: withReasoning(`{"effort":"minimal"}`), count: 3,
			want: modelAnd(&llms.ReasoningConfig{Effort: llms.ReasoningMinimal})},
		{name: "a medium effort", json: withReasoning(`{"effort":"medium","max_tokens":0}`), count: 3,
			want: modelAnd(&llms.ReasoningConfig{Effort: llms.ReasoningMedium})},
		{name: "a high effort", json: withReasoning(`{"effort":"high","max_tokens":0}`), count: 3,
			want: modelAnd(&llms.ReasoningConfig{Effort: llms.ReasoningHigh})},
		{name: "an xhigh effort", json: withReasoning(`{"effort":"xhigh"}`), count: 3,
			want: modelAnd(&llms.ReasoningConfig{Effort: llms.ReasoningXHigh})},
		{name: "a max effort", json: withReasoning(`{"effort":"max"}`), count: 3,
			want: modelAnd(&llms.ReasoningConfig{Effort: llms.ReasoningMax})},
		{name: "a budget under a none effort", json: withReasoning(`{"effort":"none","max_tokens":5000}`), count: 3,
			want: modelAnd(&llms.ReasoningConfig{Tokens: 5000})},
		{name: "reasoning off sends the disable", json: withReasoning(`{"mode":"off"}`), count: 3,
			want: modelAnd(&llms.ReasoningConfig{Mode: llms.ReasoningOff})},
		{name: "an adaptive mode is left to the call", json: withReasoning(`{"mode":"adaptive","effort":"xhigh"}`), count: 2,
			want: modelAnd(nil)},
		{name: "a budget over the cap is not sent", json: withReasoning(`{"effort":"none","max_tokens":50000}`), count: 2,
			want: modelAnd(nil)},
		{name: "a negative budget is not sent", json: withReasoning(`{"effort":"none","max_tokens":-100}`), count: 2,
			want: modelAnd(nil)},
		{name: "a none effort without a budget sends nothing", json: withReasoning(`{"effort":"none","max_tokens":0}`), count: 2,
			want: modelAnd(nil)},
		{name: "only the keys given", json: `{"model":"test-model","temperature":0.7}`, count: 2, want: modelAnd(nil)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var ac *AgentConfig
			switch {
			case tt.yaml != "":
				ac = &AgentConfig{}
				require.NoError(t, yaml.Unmarshal([]byte(tt.yaml), ac))
			case tt.json != "":
				ac = &AgentConfig{}
				require.NoError(t, json.Unmarshal([]byte(tt.json), ac))
			}

			options := ac.BuildOptions()
			assert.Len(t, options, tt.count)
			if tt.count == 0 {
				assert.Nil(t, options, "the simple_json and assistant fallbacks read nil as an agent that sets nothing")
			}
			assert.Equal(t, tt.want, configResolve(options))
		})
	}
}

func TestConfig_GetOptionsForType_LayersTheAgentBlockOverTheSharedDefaults(t *testing.T) {
	shared := []llms.CallOption{llms.WithModel("shared-model"), llms.WithMaxTokens(4096), llms.WithTemperature(0.3)}
	sharedAnd := func(temperature float64, reasoning *llms.ReasoningConfig) llms.CallOptions {
		return llms.CallOptions{Model: configPtr("shared-model"), MaxTokens: configPtr(4096),
			Temperature: configPtr(temperature), Reasoning: reasoning}
	}
	const simpleBlock = `{"simple":{"model":"test-model","max_tokens":100,"temperature":0.7}}`
	off := &llms.ReasoningConfig{Mode: llms.ReasoningOff}

	for _, tt := range []struct {
		name     string
		config   string // empty for no config at all
		defaults []llms.CallOption
		agent    ProviderOptionsType
		count    int
		want     llms.CallOptions
	}{
		{name: "no config at all", agent: OptionsTypeSimple},
		{name: "an agent type no config knows", config: simpleBlock, defaults: shared, agent: "invalid"},
		{name: "a block without shared defaults", config: simpleBlock, agent: OptionsTypeSimple, count: 3,
			want: llms.CallOptions{Model: configPtr("test-model"), MaxTokens: configPtr(100), Temperature: configPtr(0.7)}},
		{name: "an unset agent without shared defaults", config: simpleBlock, agent: OptionsTypePrimaryAgent},
		{name: "a block over the shared defaults", config: simpleBlock, defaults: shared, agent: OptionsTypeSimple, count: 6,
			want: llms.CallOptions{Model: configPtr("test-model"), MaxTokens: configPtr(100), Temperature: configPtr(0.7)}},
		{name: "a value the block leaves out takes the default or stays off the wire without one",
			config:   `{"simple":{"model":"test-model"}}`,
			defaults: []llms.CallOption{llms.WithModel("shared-model"), llms.WithMaxTokens(4096)}, agent: OptionsTypeSimple,
			count: 3, want: llms.CallOptions{Model: configPtr("test-model"), MaxTokens: configPtr(4096)}},
		{name: "an unset agent takes the shared defaults", config: simpleBlock, defaults: shared,
			agent: OptionsTypePrimaryAgent, count: 3, want: sharedAnd(0.3, nil)},
		{
			name:     "a block naming some values keeps the shared rest",
			config:   `{"primary_agent":{"model":"agent-model","temperature":0.9,"reasoning":{"mode":"off"}}}`,
			defaults: shared, agent: OptionsTypePrimaryAgent, count: 6,
			want: llms.CallOptions{Model: configPtr("agent-model"), MaxTokens: configPtr(4096), Temperature: configPtr(0.9), Reasoning: off},
		},
		{name: "a block naming only reasoning off keeps the shared settings",
			config:   `{"primary_agent":{"reasoning":{"mode":"off"}}}`,
			defaults: shared, agent: OptionsTypePrimaryAgent, count: 4, want: sharedAnd(0.3, off)},
		{name: "a block naming only a budget keeps the shared settings",
			config:   `{"primary_agent":{"reasoning":{"mode":"budget","max_tokens":2048}}}`,
			defaults: shared, agent: OptionsTypePrimaryAgent, count: 4, want: sharedAnd(0.3, &llms.ReasoningConfig{Tokens: 2048})},
		{name: "a block naming only an effort keeps the shared settings",
			config:   `{"primary_agent":{"reasoning":{"mode":"effort","effort":"low"}}}`,
			defaults: shared, agent: OptionsTypePrimaryAgent, count: 4, want: sharedAnd(0.3, &llms.ReasoningConfig{Effort: llms.ReasoningLow})},
		{name: "a block naming only adaptive keeps the shared settings",
			config:   `{"primary_agent":{"reasoning":{"mode":"adaptive"}}}`,
			defaults: shared, agent: OptionsTypePrimaryAgent, count: 3, want: sharedAnd(0.3, nil)},
		{name: "simple_json falls back to simple and asks for json", config: `{"simple":{"reasoning":{"mode":"off"}}}`,
			defaults: shared, agent: OptionsTypeSimpleJSON, count: 5,
			want: llms.CallOptions{Model: configPtr("shared-model"), MaxTokens: configPtr(4096), Temperature: configPtr(0.3),
				Reasoning: off, JSONMode: true}},
		{name: "an assistant block that sets nothing falls back to the primary agent",
			config:   `{"assistant":{},"primary_agent":{"reasoning":{"mode":"off"}}}`,
			defaults: shared, agent: OptionsTypeAssistant, count: 4, want: sharedAnd(0.3, off)},
		{name: "a legacy agent block configures the primary agent", config: `{"agent":{"temperature":0.7}}`,
			defaults: shared, agent: OptionsTypePrimaryAgent, count: 4, want: sharedAnd(0.7, nil)},
		{name: "a legacy agent block configures the assistant", config: `{"agent":{"temperature":0.7}}`,
			defaults: shared, agent: OptionsTypeAssistant, count: 4, want: sharedAnd(0.7, nil)},
		{name: "an assistant block of its own", config: `{"assistant":{"temperature":0.7}}`,
			defaults: shared, agent: OptionsTypeAssistant, count: 4, want: sharedAnd(0.7, nil)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var pc *ProviderConfig
			if tt.config != "" {
				var err error
				pc, err = LoadConfigData([]byte(tt.config), tt.defaults)
				require.NoError(t, err)
			}

			options := pc.GetOptionsForType(tt.agent)
			assert.Len(t, options, tt.count)
			assert.Equal(t, tt.want, configResolve(options))
		})
	}
}

func TestConfig_LoadConfig_ReadsAJSONOrYAMLFile(t *testing.T) {
	defaults := []llms.CallOption{llms.WithMaxTokens(1000)}

	for _, tt := range []struct {
		name    string
		file    string // empty passes no path; a file without content is never written
		content string
		want    *AgentConfig
		err     string
	}{
		{name: "no path at all"},
		{name: "a file that is not there", file: "absent.json", err: "failed to read config file"},
		{name: "malformed json", file: "config.json", content: "{invalid}", err: "failed to parse JSON config"},
		{name: "malformed yaml", file: "config.yaml", content: "invalid: [yaml", err: "failed to parse YAML config"},
		{name: "an extension it does not read", file: "config.txt", content: "some text", err: "unsupported config file extension: .txt"},
		{
			name:    "json",
			file:    "config.json",
			content: `{"simple":{"model":"test-model","max_tokens":100,"temperature":0.7,"reasoning":{"effort":"medium","max_tokens":5000}}}`,
			want: &AgentConfig{Model: "test-model", MaxTokens: 100, Temperature: 0.7,
				Reasoning: ReasoningConfig{Effort: llms.ReasoningMedium, MaxTokens: 5000}},
		},
		{
			name:    "yaml",
			file:    "config.yaml",
			content: "simple:\n  model: test-model\n  max_tokens: 100\n  temperature: 0.7\n  reasoning:\n    effort: high\n    max_tokens: 8000\n",
			want: &AgentConfig{Model: "test-model", MaxTokens: 100, Temperature: 0.7,
				Reasoning: ReasoningConfig{Effort: llms.ReasoningHigh, MaxTokens: 8000}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.file != "" {
				path = filepath.Join(t.TempDir(), tt.file)
				if tt.content != "" {
					require.NoError(t, os.WriteFile(path, []byte(tt.content), 0o600))
				}
			}

			cfg, err := LoadConfig(path, defaults)
			if tt.err != "" {
				assert.ErrorContains(t, err, tt.err)
				return
			}
			require.NoError(t, err)
			if tt.want == nil {
				assert.Nil(t, cfg)
				return
			}

			require.NotNil(t, cfg.Simple)
			simple := *cfg.Simple
			simple.ClearRaw()
			assert.Equal(t, *tt.want, simple)
			assert.Equal(t, llms.CallOptions{MaxTokens: configPtr(1000)}, configResolve(cfg.GetOptionsForType(OptionsTypeCoder)),
				"an agent the file leaves unset takes the defaults the loader was given")
			assert.Equal(t, tt.content, string(cfg.GetRawConfig()))
		})
	}
}

func TestConfig_HandleLegacyConfig_ReadsTheLegacyAgentBlock(t *testing.T) {
	for _, tt := range []struct {
		name          string
		file          string
		content       string
		primary       *AgentConfig
		assistant     *AgentConfig // compared only when the assistant is not the primary agent itself
		simple        *AgentConfig
		sameAssistant bool
	}{
		{
			name: "a legacy agent block in json",
			file: "legacy.json",
			content: `{"agent":{"model":"legacy-model","max_tokens":200,"temperature":0.8},` +
				`"simple":{"model":"simple-model","max_tokens":100}}`,
			primary:       &AgentConfig{Model: "legacy-model", MaxTokens: 200, Temperature: 0.8},
			simple:        &AgentConfig{Model: "simple-model", MaxTokens: 100},
			sameAssistant: true,
		},
		{
			name: "a legacy agent block in yaml",
			file: "legacy.yaml",
			content: "agent:\n  model: legacy-yaml-model\n  max_tokens: 300\n  temperature: 0.9\n" +
				"simple:\n  model: simple-yaml-model\n  max_tokens: 150\n",
			primary:       &AgentConfig{Model: "legacy-yaml-model", MaxTokens: 300, Temperature: 0.9},
			simple:        &AgentConfig{Model: "simple-yaml-model", MaxTokens: 150},
			sameAssistant: true,
		},
		{
			name: "a primary_agent block wins over the legacy one",
			file: "new_format.json",
			content: `{"primary_agent":{"model":"new-model","max_tokens":400,"temperature":0.6},` +
				`"agent":{"model":"old-model","max_tokens":200,"temperature":0.8}}`,
			primary:       &AgentConfig{Model: "new-model", MaxTokens: 400, Temperature: 0.6},
			sameAssistant: true,
		},
		{
			name: "an assistant block of its own is kept",
			file: "explicit_assistant.yaml",
			content: "agent:\n  model: agent-model\n  max_tokens: 200\n" +
				"assistant:\n  model: assistant-model\n  max_tokens: 500\n  temperature: 0.5\n",
			primary:   &AgentConfig{Model: "agent-model", MaxTokens: 200},
			assistant: &AgentConfig{Model: "assistant-model", MaxTokens: 500, Temperature: 0.5},
		},
		{
			name:    "no agent block at all",
			file:    "no_agents.json",
			content: `{"simple":{"model":"simple-only","max_tokens":100}}`,
			simple:  &AgentConfig{Model: "simple-only", MaxTokens: 100},
		},
	} {
		path := filepath.Join(t.TempDir(), tt.file)
		require.NoError(t, os.WriteFile(path, []byte(tt.content), 0o600))

		for _, loader := range []struct {
			name string
			load func() (*ProviderConfig, error)
		}{
			{"from a file", func() (*ProviderConfig, error) { return LoadConfig(path, nil) }},
			{"from data", func() (*ProviderConfig, error) { return LoadConfigData([]byte(tt.content), nil) }},
		} {
			t.Run(tt.name+" "+loader.name, func(t *testing.T) {
				cfg, err := loader.load()
				require.NoError(t, err)
				require.NotNil(t, cfg)

				configAssertAgent(t, "primary agent", tt.primary, cfg.PrimaryAgent)
				configAssertAgent(t, "simple", tt.simple, cfg.Simple)
				if tt.sameAssistant {
					assert.Same(t, cfg.PrimaryAgent, cfg.Assistant, "the assistant runs on the primary agent's block")
				} else {
					configAssertAgent(t, "assistant", tt.assistant, cfg.Assistant)
				}
			})
		}
	}
}

func configAssertAgent(t *testing.T, slot string, want, got *AgentConfig) {
	t.Helper()

	if want == nil {
		assert.Nil(t, got, slot)
		return
	}
	require.NotNil(t, got, slot)
	decoded := *got
	decoded.ClearRaw()
	assert.Equal(t, *want, decoded, slot)
}

func TestConfig_UsesAdaptiveThinking_FollowsTheAgentChoiceAndTheModelCapability(t *testing.T) {
	adaptiveOnly := ModelsConfig{{Name: "opus-4-8", Reasoning: &ModelReasoningInfo{Mode: ModelReasoningAdaptiveOnly}}}
	adaptiveCapable := ModelsConfig{{Name: "opus-4-6", Reasoning: &ModelReasoningInfo{Mode: ModelReasoningAdaptive}}}
	budgetOnly := ModelsConfig{{Name: "sonnet-4-5", Reasoning: &ModelReasoningInfo{Mode: ModelReasoningBudget}}}

	tests := []struct {
		name   string
		models ModelsConfig
		agent  *AgentConfig
		want   bool
	}{
		{"adaptive-only model forces adaptive with no agent reasoning",
			adaptiveOnly, &AgentConfig{Model: "opus-4-8"}, true},
		{"adaptive-only model overrides an agent budget choice",
			adaptiveOnly, &AgentConfig{Model: "opus-4-8", Reasoning: ReasoningConfig{Mode: ReasoningModeBudget, MaxTokens: 4096}}, true},
		{"explicit off wins over the adaptive-only auto-adaptive",
			adaptiveOnly, &AgentConfig{Model: "opus-4-8", Reasoning: ReasoningConfig{Mode: ReasoningModeOff}}, false},
		{"agent selects adaptive on an adaptive-capable model",
			adaptiveCapable, &AgentConfig{Model: "opus-4-6", Reasoning: ReasoningConfig{Mode: ReasoningModeAdaptive}}, true},
		{"agent selects budget on an adaptive-capable model",
			adaptiveCapable, &AgentConfig{Model: "opus-4-6", Reasoning: ReasoningConfig{Mode: ReasoningModeBudget, MaxTokens: 4096}}, false},
		{"unknown model still honors an explicit adaptive agent choice",
			ModelsConfig{}, &AgentConfig{Model: "unknown", Reasoning: ReasoningConfig{Mode: ReasoningModeAdaptive}}, true},
		{"budget-only model with a budget agent",
			budgetOnly, &AgentConfig{Model: "sonnet-4-5", Reasoning: ReasoningConfig{Mode: ReasoningModeBudget, MaxTokens: 4096}}, false},
		{"no reasoning anywhere",
			budgetOnly, &AgentConfig{Model: "sonnet-4-5"}, false},
		{"empty model cannot match adaptive-only catalog entry",
			adaptiveOnly, &AgentConfig{Model: ""}, false},
		{"nil agent is not adaptive",
			adaptiveOnly, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProviderConfig{Simple: tt.agent}
			assert.Equal(t, tt.want, pc.UsesAdaptiveThinking(tt.models, OptionsTypeSimple))
		})
	}
}

func TestConfig_UsesAdaptiveThinking_FollowsTheDoorDefaultModelWhenTheBlockNamesNone(t *testing.T) {
	adaptiveOnly := ModelsConfig{{Name: "opus-4-8", Reasoning: &ModelReasoningInfo{Mode: ModelReasoningAdaptiveOnly}}}

	pc, err := LoadConfigData([]byte("simple:\n  temperature: 0.5\nagent:\n  temperature: 0.5\n"),
		[]llms.CallOption{llms.WithModel("opus-4-8")})
	require.NoError(t, err)

	for _, opt := range []ProviderOptionsType{
		OptionsTypeSimple, OptionsTypeSimpleJSON, OptionsTypePrimaryAgent, OptionsTypeAssistant, OptionsTypeCoder,
	} {
		assert.True(t, pc.UsesAdaptiveThinking(adaptiveOnly, opt), opt)
	}
}

func TestConfig_PrepareAdaptiveCallOptions_AppendsTheAdaptiveOptionOnlyWhenAdaptive(t *testing.T) {
	adaptiveOnly := ModelsConfig{{Name: "opus-4-8", Reasoning: &ModelReasoningInfo{Mode: ModelReasoningAdaptiveOnly}}}
	budgetOnly := ModelsConfig{{Name: "sonnet-4-5", Reasoning: &ModelReasoningInfo{Mode: ModelReasoningBudget}}}

	for _, tt := range []struct {
		name     string
		models   ModelsConfig
		agent    *AgentConfig
		existing []llms.CallOption
		count    int
		want     llms.CallOptions
	}{
		{
			name:     "appends after the existing options and forwards the effort",
			models:   adaptiveOnly,
			agent:    &AgentConfig{Model: "opus-4-8", Reasoning: ReasoningConfig{Effort: llms.ReasoningHigh}},
			existing: []llms.CallOption{llms.WithTemperature(0.5)},
			count:    2,
			want: llms.CallOptions{Temperature: configPtr(0.5),
				Reasoning: &llms.ReasoningConfig{Effort: llms.ReasoningHigh, Adaptive: true}},
		},
		{
			name:   "sends no effort when the agent has no reasoning block",
			models: adaptiveOnly,
			agent:  &AgentConfig{Model: "opus-4-8"},
			count:  1,
			want:   llms.CallOptions{Reasoning: &llms.ReasoningConfig{Adaptive: true}},
		},
		{
			name:     "appends nothing when the agent turned reasoning off",
			models:   adaptiveOnly,
			agent:    &AgentConfig{Model: "opus-4-8", Reasoning: ReasoningConfig{Mode: ReasoningModeOff}},
			existing: []llms.CallOption{llms.WithTemperature(1)},
			count:    1,
			want:     llms.CallOptions{Temperature: configPtr(1.0)},
		},
		{
			name:     "appends nothing to a call that is not adaptive",
			models:   budgetOnly,
			agent:    &AgentConfig{Model: "sonnet-4-5", Reasoning: ReasoningConfig{Mode: ReasoningModeBudget, MaxTokens: 4096}},
			existing: []llms.CallOption{llms.WithTemperature(1)},
			count:    1,
			want:     llms.CallOptions{Temperature: configPtr(1.0)},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProviderConfig{Simple: tt.agent}
			_, options := pc.PrepareAdaptiveCallOptions(context.Background(), tt.models, OptionsTypeSimple, tt.existing)
			assert.Len(t, options, tt.count)
			assert.Equal(t, tt.want, configResolve(options))
		})
	}
}

func TestConfig_SetDefaultOptions_KeepsOneAgentsEffortOutOfAnother(t *testing.T) {
	raw := []byte(`{
		"simple": {"reasoning": {"mode": "adaptive", "effort": "low"}},
		"pentester": {"reasoning": {"mode": "adaptive", "effort": "xhigh"}}
	}`)

	defaults := []llms.CallOption{
		llms.WithTemperature(1.0),
		llms.WithTopP(1.0),
		llms.WithN(1),
		llms.WithMaxTokens(16384),
	}
	defaults = append(defaults, llms.WithModel("gateway-model"))
	require.Greater(t, cap(defaults), len(defaults),
		"this test is only meaningful while the defaults carry spare capacity to append into")

	pc, err := LoadConfigData(raw, defaults)
	require.NoError(t, err)
	require.Len(t, pc.GetOptionsForType(OptionsTypeSimple), len(defaults),
		"a reasoning-only agent block must fall back to the shared defaults")

	_, simple := pc.PrepareAdaptiveCallOptions(
		context.Background(), nil, OptionsTypeSimple, pc.GetOptionsForType(OptionsTypeSimple))
	_, pentester := pc.PrepareAdaptiveCallOptions(
		context.Background(), nil, OptionsTypePentester, pc.GetOptionsForType(OptionsTypePentester))

	require.NotNil(t, configResolve(pentester).Reasoning)
	require.NotNil(t, configResolve(simple).Reasoning)
	assert.Equal(t, llms.ReasoningXHigh, configResolve(pentester).Reasoning.Effort)
	assert.Equal(t, llms.ReasoningLow, configResolve(simple).Reasoning.Effort,
		"preparing another agent must not rewrite an already-prepared option slice")
}

func TestConfig_LoadModelsConfigData_ReadsTheCatalogueFormat(t *testing.T) {
	models, err := LoadModelsConfigData([]byte(`
- name: gpt-4o
  description: Fast, intelligent, flexible GPT model
  thinking: false
  release_date: 2024-05-13
  price:
    input: 2.5
    output: 10.0
- name: o3-mini
  thinking: true
  release_date: 2025-01-31
  reasoning:
    mode: adaptive-only
    efforts: [low, high, xhigh]
  context_window: 200000
- name: gemma-3-27b-it
  thinking: false
- name: free-model
  price:
    input: 0.0
    output: 0.0
`))
	require.NoError(t, err)

	description, thinks, silent := "Fast, intelligent, flexible GPT model", true, false
	gpt4oRelease, o3Release := time.Date(2024, 5, 13, 0, 0, 0, 0, time.UTC), time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	window := 200000
	assert.Equal(t, ModelsConfig{
		{Name: "gpt-4o", Description: &description, Thinking: &silent, ReleaseDate: &gpt4oRelease,
			Price: &PriceInfo{Input: 2.5, Output: 10}},
		{Name: "o3-mini", Thinking: &thinks, ReleaseDate: &o3Release, ContextWindow: &window,
			Reasoning: &ModelReasoningInfo{Mode: ModelReasoningAdaptiveOnly,
				Efforts: []llms.ReasoningEffort{llms.ReasoningLow, llms.ReasoningHigh, llms.ReasoningXHigh}}},
		{Name: "gemma-3-27b-it", Thinking: &silent},
		{Name: "free-model", Price: &PriceInfo{}},
	}, models)

	empty, err := LoadModelsConfigData(nil)
	require.NoError(t, err)
	assert.Empty(t, empty)
}

// A user-defined provider is stored serialised, so a field either codec loses disappears on save.
func TestConfig_ModelConfig_SurvivesBothCodecs(t *testing.T) {
	description, thinking := "Fast model", false
	released := time.Date(2024, 5, 13, 0, 0, 0, 0, time.UTC)
	window, ceiling := 200000, 64000

	for _, tt := range []struct {
		name  string
		model ModelConfig
		json  string
		yaml  string
	}{
		{
			name: "every field",
			model: ModelConfig{
				Name: "gpt-4o", Description: &description, Thinking: &thinking, ReleaseDate: &released,
				Reasoning: &ModelReasoningInfo{Mode: ModelReasoningAdaptiveOnly,
					Efforts: []llms.ReasoningEffort{llms.ReasoningLow, llms.ReasoningHigh, llms.ReasoningXHigh}},
				Price:         &PriceInfo{Input: 2.5, Output: 10, CacheRead: 1.25, CacheWrite: 3},
				ContextWindow: &window, MaxOutputTokens: &ceiling,
			},
			json: `{"name":"gpt-4o","description":"Fast model","thinking":false,"release_date":"2024-05-13",` +
				`"reasoning":{"mode":"adaptive-only","efforts":["low","high","xhigh"]},` +
				`"price":{"input":2.5,"output":10,"cache_read":1.25,"cache_write":3},` +
				`"context_window":200000,"max_output_tokens":64000}`,
			yaml: "name: gpt-4o\ndescription: Fast model\nthinking: false\nrelease_date: \"2024-05-13\"\n" +
				"reasoning: {mode: adaptive-only, efforts: [low, high, xhigh]}\n" +
				"price: {input: 2.5, output: 10, cache_read: 1.25, cache_write: 3}\n" +
				"context_window: 200000\nmax_output_tokens: 64000\n",
		},
		{name: "every field left unset stays absent rather than zero", model: ModelConfig{}, json: `{}`, yaml: "{}\n"},
	} {
		t.Run(tt.name+" in json", func(t *testing.T) {
			data, err := json.Marshal(tt.model)
			require.NoError(t, err)
			assert.JSONEq(t, tt.json, string(data))

			var back ModelConfig
			require.NoError(t, json.Unmarshal(data, &back))
			assert.Equal(t, tt.model, back)
		})
		t.Run(tt.name+" in yaml", func(t *testing.T) {
			data, err := yaml.Marshal(tt.model)
			require.NoError(t, err)
			assert.YAMLEq(t, tt.yaml, string(data))

			var back ModelConfig
			require.NoError(t, yaml.Unmarshal(data, &back))
			assert.Equal(t, tt.model, back)
		})
	}
}

func TestConfig_ModelConfig_RefusesAMalformedEntry(t *testing.T) {
	for _, tt := range []struct {
		name string
		json string
		yaml string // a catalogue read through LoadModelsConfigData
		err  string
	}{
		{name: "a json release date not in YYYY-MM-DD", json: `{"name":"test","release_date":"invalid-date"}`,
			err: "invalid release_date format, expected YYYY-MM-DD"},
		{name: "a malformed json entry", json: `{invalid}`, err: "invalid character 'i'"},
		{name: "a yaml release date not in YYYY-MM-DD", yaml: "- name: test\n  release_date: \"invalid-date\"\n",
			err: "failed to parse models config: invalid release_date format, expected YYYY-MM-DD"},
		{name: "a malformed yaml catalogue", yaml: "invalid: [yaml", err: "failed to parse models config: yaml:"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.json != "" {
				var model ModelConfig
				err = json.Unmarshal([]byte(tt.json), &model)
			} else {
				_, err = LoadModelsConfigData([]byte(tt.yaml))
			}
			assert.ErrorContains(t, err, tt.err)
		})
	}
}

func TestConfig_ValidateToolReasoning_RefusesWhatTheDoorRefuses(t *testing.T) {
	url := configChatServer(t)

	reasonings := []string{
		`{}`,
		`{"effort":"minimal"}`,
		`{"effort":"low"}`,
		`{"effort":"xhigh"}`,
		`{"effort":"max"}`,
		`{"max_tokens":2048}`,
		`{"mode":"budget","max_tokens":2048}`,
		`{"mode":"adaptive"}`,
		`{"mode":"adaptive","effort":"medium"}`,
		`{"mode":"off"}`,
		`{"mode":"off","effort":"high","max_tokens":2048}`,
		`{"effort":"high","max_tokens":2048}`,
	}
	tool := llms.Tool{Type: "function", Function: &llms.FunctionDefinition{
		Name: "get_weather", Parameters: map[string]any{"type": "object"},
	}}

	refusals := 0
	for _, opt := range AllAgentTypes {
		for _, model := range []string{"gpt-5.6-sol", "gpt-5.5", "gpt-5.4-nano", "gpt-5.2", "gpt-4o"} {
			for _, r := range reasonings {
				t.Run(fmt.Sprintf("%s on %s with reasoning %s", opt, model, r), func(t *testing.T) {
					cfg, err := LoadConfigData([]byte(fmt.Sprintf(`{%q:{"model":%q,"reasoning":%s}}`, opt, model, r)), nil)
					require.NoError(t, err)

					options := cfg.GetOptionsForType(opt)
					if configAgentCallsCarryTools[opt] {
						options = append(options, llms.WithTools([]llms.Tool{tool}))
					}
					_, options = cfg.PrepareAdaptiveCallOptions(context.Background(), nil, opt, options)

					var refusal *reasoning.ErrEffortWithTools
					refused := errors.As(configCallTheDoor(t, url, model, options), &refusal)
					if refused {
						refusals++
					}

					err = cfg.ValidateToolReasoning(reasoning.ProviderOpenAI)
					assert.Equal(t, refused, err != nil, "door refuses: %v, validation: %v", refused, err)
				})
			}
		}
	}
	require.Positive(t, refusals, "the table must reach the refusal it checks")
}

func TestConfig_ValidateToolReasoning_NamesTheAgentModelAndLevelToClear(t *testing.T) {
	const tail = " on requests with function tools; clear it or turn reasoning off"

	for _, tt := range []struct {
		name     string
		config   string
		defaults []llms.CallOption
		door     reasoning.Provider
		want     string
	}{
		{name: "an effort", config: `{"simple":{"model":"gpt-5.6-sol","reasoning":{"effort":"medium"}}}`,
			door: reasoning.ProviderOpenAI, want: `simple: model "gpt-5.6-sol" refuses reasoning effort "medium"` + tail},
		{name: "an adaptive effort", config: `{"simple":{"model":"gpt-5.6-sol","reasoning":{"mode":"adaptive","effort":"high"}}}`,
			door: reasoning.ProviderOpenAI, want: `simple: model "gpt-5.6-sol" refuses reasoning effort "high"` + tail},
		{name: "a budget without a mode", config: `{"simple":{"model":"gpt-5.6-sol","reasoning":{"max_tokens":2048}}}`,
			door: reasoning.ProviderOpenAI, want: `simple: model "gpt-5.6-sol" refuses a reasoning budget of 2048 tokens` + tail},
		{name: "a budget mode with a leftover effort",
			config: `{"simple":{"model":"gpt-5.6-sol","reasoning":{"mode":"budget","max_tokens":4096,"effort":"low"}}}`,
			door:   reasoning.ProviderOpenAI, want: `simple: model "gpt-5.6-sol" refuses a reasoning budget of 4096 tokens` + tail},
		{name: "the door default model when the block names none", config: `{"coder":{"reasoning":{"effort":"high"}}}`,
			defaults: []llms.CallOption{llms.WithModel("gpt-5.5")},
			door:     reasoning.ProviderOpenAI, want: `coder: model "gpt-5.5" refuses reasoning effort "high"` + tail},
		{name: "another door does not refuse", config: `{"coder":{"model":"gpt-5.6-sol","reasoning":{"effort":"high"}}}`,
			door: reasoning.ProviderOllama},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := LoadConfigData([]byte(tt.config), tt.defaults)
			require.NoError(t, err)

			err = cfg.ValidateToolReasoning(tt.door)
			if tt.want == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, tt.want)
		})
	}
}

func TestConfig_ValidateReasoningOff_RefusesWhatTheDoorRefuses(t *testing.T) {
	url := configChatServer(t)

	refusals := 0
	for _, model := range []string{"o3", "o4-mini", "gpt-oss-120b", "deepseek-reasoner", "gpt-5.2", "gpt-4o", "glm-5-turbo"} {
		for _, r := range []string{`{}`, `{"mode":"off"}`, `{"effort":"high"}`, `{"mode":"off","max_tokens":2048}`} {
			t.Run(fmt.Sprintf("%s with reasoning %s", model, r), func(t *testing.T) {
				cfg, err := LoadConfigData([]byte(fmt.Sprintf(`{"adviser":{"model":%q,"reasoning":%s}}`, model, r)), nil)
				require.NoError(t, err)

				var refusal *reasoning.ErrReasoningOffUnsupported
				refused := errors.As(configCallTheDoor(t, url, model, cfg.GetOptionsForType(OptionsTypeAdviser)), &refusal)
				if refused {
					refusals++
				}

				err = cfg.ValidateReasoningOff(reasoning.ProviderOpenAI)
				assert.Equal(t, refused, err != nil, "door refuses: %v, validation: %v", refused, err)
			})
		}
	}
	require.Positive(t, refusals, "the table must reach the refusal it checks")
}

func TestConfig_ValidateReasoningOff_NamesTheAgentAndModelThatCannotTurnOff(t *testing.T) {
	for _, tt := range []struct {
		name     string
		model    string
		defaults []llms.CallOption
		door     reasoning.Provider
		want     string
	}{
		{name: "a model the ollama door cannot turn off", model: "gpt-oss:120b", door: reasoning.ProviderOllama,
			want: `coder: model "gpt-oss:120b" cannot turn reasoning off; clear Reasoning Mode Off`},
		{name: "a model the ollama door turns off and the openai door cannot", model: "deepseek-r1:7b", door: reasoning.ProviderOllama},
		{name: "the door default model when the block names none", defaults: []llms.CallOption{llms.WithModel("o3")},
			door: reasoning.ProviderOpenAI, want: `coder: model "o3" cannot turn reasoning off; clear Reasoning Mode Off`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			block := `{"coder":{"reasoning":{"mode":"off"}}}`
			if tt.model != "" {
				block = fmt.Sprintf(`{"coder":{"model":%q,"reasoning":{"mode":"off"}}}`, tt.model)
			}
			cfg, err := LoadConfigData([]byte(block), tt.defaults)
			require.NoError(t, err)

			err = cfg.ValidateReasoningOff(tt.door)
			if tt.want == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, tt.want)
		})
	}
}

func TestConfig_ValidateThinkingBudget_RefusesWhatTheDoorRefuses(t *testing.T) {
	url := configChatServer(t)

	refusals := 0
	for _, route := range []struct{ prefix, model string }{
		{"", "glm-5-turbo"}, {"zai", "glm-5-turbo"}, {"dashscope", "glm-5-turbo"}, {"zai", "glm-5.2"},
		{"moonshot", "kimi-k2.6"}, {"minimax", "MiniMax-M3"}, {"", "o3"},
	} {
		wire := route.model
		if route.prefix != "" {
			wire = route.prefix + "/" + route.model
		}
		for _, r := range []string{`{}`, `{"max_tokens":2048}`, `{"mode":"budget","max_tokens":2048}`, `{"mode":"off"}`, `{"effort":"high"}`} {
			t.Run(fmt.Sprintf("%s with reasoning %s", wire, r), func(t *testing.T) {
				cfg, err := LoadConfigData([]byte(fmt.Sprintf(`{"adviser":{"model":%q,"reasoning":%s}}`, route.model, r)), nil)
				require.NoError(t, err)

				options := append(cfg.GetOptionsForType(OptionsTypeAdviser), llms.WithModel(wire))
				var refusal *reasoning.ErrThinkingBudgetUnsupported
				refused := errors.As(configCallTheDoor(t, url, wire, options), &refusal)
				if refused {
					refusals++
				}

				err = cfg.ValidateThinkingBudget(reasoning.ProviderOpenAI, route.prefix)
				assert.Equal(t, refused, err != nil, "door refuses: %v, validation: %v", refused, err)
			})
		}
	}
	require.Positive(t, refusals, "the table must reach the refusal it checks")
}

func TestConfig_ValidateThinkingBudget_KeepsToTheOpenAIDoor(t *testing.T) {
	cfg, err := LoadConfigData([]byte(`{"coder":{"model":"glm-5-turbo","reasoning":{"max_tokens":2048}}}`), nil)
	require.NoError(t, err)

	assert.EqualError(t, cfg.ValidateThinkingBudget(reasoning.ProviderOpenAI, "zai"),
		`coder: model "zai/glm-5-turbo" takes no thinking budget; clear Reasoning Max Tokens`)
	assert.NoError(t, cfg.ValidateThinkingBudget(reasoning.ProviderBedrock, ""))
}

func TestConfig_ValidateThinkingAgreement_RefusesAnExtraBodyToggleThatContradictsReasoning(t *testing.T) {
	const (
		enablesAgainstOff = `adviser: extra_body enables thinking while reasoning.mode is "off"; ` +
			`drop the extra_body thinking toggle or clear reasoning.mode`
		disablesAgainstOn = `adviser: extra_body disables thinking while the reasoning block requests it; ` +
			`drop the extra_body thinking toggle or turn reasoning off`
	)

	for _, tc := range []struct {
		name      string
		agentJSON string
		want      string
	}{
		{name: "off with no extra_body thinking", agentJSON: `{"model":"m","reasoning":{"mode":"off"}}`},
		{name: "off agreeing with extra_body thinking disabled",
			agentJSON: `{"model":"m","reasoning":{"mode":"off"},"extra_body":{"thinking":{"type":"disabled"}}}`},
		{name: "off with a leftover effort agreeing with extra_body thinking disabled",
			agentJSON: `{"model":"m","reasoning":{"mode":"off","effort":"high"},"extra_body":{"thinking":{"type":"disabled"}}}`},
		{name: "off with a leftover effort agreeing with enable_thinking false",
			agentJSON: `{"model":"m","reasoning":{"mode":"off","effort":"high"},"extra_body":{"enable_thinking":false}}`},
		{name: "off beside a vendor key in the extra_body thinking object",
			agentJSON: `{"model":"m","reasoning":{"mode":"off"},"extra_body":{"thinking":{"clear_thinking":false}}}`},
		{name: "effort with a non-toggle extra_body thinking object",
			agentJSON: `{"model":"m","reasoning":{"effort":"high"},"extra_body":{"thinking":{"clear_thinking":false}}}`},
		{name: "default reasoning leaves the extra_body toggle alone",
			agentJSON: `{"model":"m","extra_body":{"enable_thinking":true}}`},
		{name: "off agreeing with reasoning_effort none",
			agentJSON: `{"model":"m","reasoning":{"mode":"off"},"extra_body":{"reasoning_effort":"none"}}`},
		{name: "effort beside another reasoning_effort level",
			agentJSON: `{"model":"m","reasoning":{"effort":"high"},"extra_body":{"reasoning_effort":"low"}}`},
		{name: "off contradicted by reasoning_effort high",
			agentJSON: `{"model":"m","reasoning":{"mode":"off"},"extra_body":{"reasoning_effort":"high"}}`, want: enablesAgainstOff},
		{name: "off contradicted by reasoning.effort high",
			agentJSON: `{"model":"m","reasoning":{"mode":"off"},"extra_body":{"reasoning":{"effort":"high"}}}`, want: enablesAgainstOff},
		{name: "effort contradicted by reasoning_effort none",
			agentJSON: `{"model":"m","reasoning":{"effort":"high"},"extra_body":{"reasoning_effort":"none"}}`, want: disablesAgainstOn},
		{name: "budget contradicted by reasoning.effort none",
			agentJSON: `{"model":"m","reasoning":{"mode":"budget","max_tokens":2048},"extra_body":{"reasoning":{"effort":"none"}}}`,
			want:      disablesAgainstOn},
		{name: "off contradicted by enable_thinking true",
			agentJSON: `{"model":"m","reasoning":{"mode":"off"},"extra_body":{"enable_thinking":true}}`, want: enablesAgainstOff},
		{name: "off contradicted by thinking type enabled",
			agentJSON: `{"model":"m","reasoning":{"mode":"off"},"extra_body":{"thinking":{"type":"enabled"}}}`, want: enablesAgainstOff},
		{name: "off contradicted by thinking type adaptive",
			agentJSON: `{"model":"m","reasoning":{"mode":"off"},"extra_body":{"thinking":{"type":"adaptive"}}}`, want: enablesAgainstOff},
		{name: "effort contradicted by enable_thinking false",
			agentJSON: `{"model":"m","reasoning":{"effort":"high"},"extra_body":{"enable_thinking":false}}`, want: disablesAgainstOn},
		{name: "effort contradicted by thinking type disabled",
			agentJSON: `{"model":"m","reasoning":{"effort":"high"},"extra_body":{"thinking":{"type":"disabled"}}}`, want: disablesAgainstOn},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var pc ProviderConfig
			require.NoError(t, json.Unmarshal([]byte(`{"adviser":`+tc.agentJSON+`}`), &pc))

			err := pc.ValidateThinkingAgreement(reasoning.ProviderOpenAI)
			if tc.want == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, tc.want)
		})
	}
}
