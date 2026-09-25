package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pentagi/pkg/providers/pconfig"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLitellm_ApplyModelPrefix_PrefixesOnlyWhenAPrefixIsSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, modelName, prefix, expected string
	}{
		{"with prefix", "deepseek-chat", "deepseek", "deepseek/deepseek-chat"},
		{"without prefix", "deepseek-chat", "", "deepseek-chat"},
		{"a model already carrying another prefix", "anthropic/claude-3", "openrouter", "openrouter/anthropic/claude-3"},
		{"a name with special characters", "claude-3.5-sonnet@20241022", "provider", "provider/claude-3.5-sonnet@20241022"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, ApplyModelPrefix(tc.modelName, tc.prefix))
		})
	}
}

func TestLitellm_RemoveModelPrefix_StripsOnlyItsOwnPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, modelName, prefix, expected string
	}{
		{"a model with the matching prefix", "deepseek/deepseek-chat", "deepseek", "deepseek-chat"},
		{"a model without a prefix", "deepseek-chat", "deepseek", "deepseek-chat"},
		{"an empty prefix", "deepseek/deepseek-chat", "", "deepseek/deepseek-chat"},
		{"a model with another prefix", "openrouter/deepseek-chat", "deepseek", "openrouter/deepseek-chat"},
		{"a model with nested prefixes", "openrouter/anthropic/claude-3", "openrouter", "anthropic/claude-3"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, RemoveModelPrefix(tc.modelName, tc.prefix))
		})
	}
}

// listedModels serves one answer on a keyed GET /models and loads it the way a door does.
func listedModels(t *testing.T, prefix string, status int, body string) (pconfig.ModelsConfig, error) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/models" || r.Header.Get("Authorization") != "Bearer test-key" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
	defer server.Close()

	return LoadModelsFromHTTP(server.URL, "test-key", &http.Client{Timeout: 5 * time.Second}, prefix)
}

func TestLitellm_LoadModelsFromHTTP_ListsEachUsableModelOnceWithWhatTheGatewayStates(t *testing.T) {
	t.Parallel()

	yes, no := true, false
	text := func(value string) *string { return &value }
	released := time.Date(2023, 6, 12, 16, 54, 56, 0, time.UTC)
	stated := `"max_input_tokens": 128000, "max_output_tokens": 16384`
	tighter := `"max_input_tokens": 64000, "max_output_tokens": 8192`
	limited := func(window, output int) pconfig.ModelsConfig {
		return pconfig.ModelsConfig{{Name: "gpt-4o", Thinking: &no, ContextWindow: limit(window), MaxOutputTokens: limit(output)}}
	}
	unlimited := pconfig.ModelsConfig{{Name: "gpt-4o", Thinking: &no}}
	thinking := func(value bool) pconfig.ModelsConfig {
		return pconfig.ModelsConfig{{Name: "claude-sonnet-5", Thinking: &value}}
	}
	windowOnly := func(name string, window int) pconfig.ModelsConfig {
		return pconfig.ModelsConfig{{Name: name, ContextWindow: limit(window)}}
	}

	tests := []struct {
		name    string
		prefix  string
		entries string
		want    pconfig.ModelsConfig
	}{
		{
			name: "every field the listing states",
			entries: `{"id": "model-a", "description": "Model A description", "supported_parameters": ["tools", "max_tokens"]},
				{"id": "model-b", "created": 1686588896, "description": "Model B description",
				 "supported_parameters": ["reasoning", "tools"], "pricing": {"prompt": "0.0001", "completion": "0.0005"}}`,
			want: pconfig.ModelsConfig{
				{Name: "model-a", Description: text("Model A description"), Thinking: &no},
				{
					Name: "model-b", Description: text("Model B description"), ReleaseDate: &released, Thinking: &yes,
					Price: &pconfig.PriceInfo{Input: 100, Output: 500},
				},
			},
		},
		{
			name:   "a prefix keeps only its own models, stripped, with per-million prices as stated",
			prefix: "deepseek",
			entries: `{"id": "deepseek/deepseek-chat", "description": "DeepSeek chat model",
				 "supported_parameters": ["tools", "max_tokens"], "pricing": {"prompt": "0.28", "completion": "0.42"}},
				{"id": "deepseek/deepseek-reasoner", "description": "DeepSeek reasoning model",
				 "supported_parameters": ["reasoning", "tools"], "pricing": {"prompt": "0.28", "completion": "0.42"}},
				{"id": "openai/gpt-4", "supported_parameters": ["tools"], "pricing": {"prompt": "30.0", "completion": "60.0"}},
				{"id": "anthropic/claude-3-opus", "supported_parameters": ["tools"]}`,
			want: pconfig.ModelsConfig{
				{
					Name: "deepseek-chat", Description: text("DeepSeek chat model"), Thinking: &no,
					Price: &pconfig.PriceInfo{Input: 0.28, Output: 0.42},
				},
				{
					Name: "deepseek-reasoner", Description: text("DeepSeek reasoning model"), Thinking: &yes,
					Price: &pconfig.PriceInfo{Input: 0.28, Output: 0.42},
				},
			},
		},
		{
			name: "a model that can neither call tools nor answer in a schema is skipped",
			entries: `{"id": "model-with-tools", "supported_parameters": ["tools", "max_tokens"]},
				{"id": "model-without-tools", "supported_parameters": ["max_tokens", "temperature"]},
				{"id": "model-with-structured-outputs", "supported_parameters": ["structured_outputs"]}`,
			want: pconfig.ModelsConfig{
				{Name: "model-with-tools", Thinking: &no},
				{Name: "model-with-structured-outputs", Thinking: &no},
			},
		},
		{
			name: "repeated names are listed once in first-seen order",
			entries: `{"id": "openai/gpt-5.6-luna", "supported_parameters": ["tools"]},
				{"id": "openai/gpt-5.6-luna", "supported_parameters": ["tools"]},
				{"id": "anthropic/claude-sonnet-5", "supported_parameters": ["tools"]},
				{"id": "openai/gpt-5.6-luna", "supported_parameters": ["tools"]}`,
			want: pconfig.ModelsConfig{
				{Name: "openai/gpt-5.6-luna", Thinking: &no},
				{Name: "anthropic/claude-sonnet-5", Thinking: &no},
			},
		},
		{
			name: "a listing only its ids can be read from falls back to the ids, once each",
			entries: `{"id": "m1", "created": "not-a-number", "pricing": {"prompt": "0.001", "completion": "0.002"}},
				{"id": "m1", "created": "not-a-number"},
				{"id": "m2", "created": "not-a-number"},
				{"id": "m1", "created": "not-a-number"}`,
			want: pconfig.ModelsConfig{{Name: "m1"}, {Name: "m2"}},
		},

		{
			name:    "reasoning stated by the first entry of a name",
			entries: `{"id": "claude-sonnet-5", "supported_parameters": ["tools", "reasoning"]}, {"id": "claude-sonnet-5", "supported_parameters": ["tools"]}`,
			want:    thinking(true),
		},
		{
			name:    "reasoning stated by the second entry of a name",
			entries: `{"id": "claude-sonnet-5", "supported_parameters": ["tools"]}, {"id": "claude-sonnet-5", "supported_parameters": ["tools", "reasoning"]}`,
			want:    thinking(true),
		},
		{
			name:    "reasoning stated only by an unusable entry listed first",
			entries: `{"id": "claude-sonnet-5", "supported_parameters": ["reasoning"]}, {"id": "claude-sonnet-5", "supported_parameters": ["tools"]}`,
			want:    thinking(false),
		},
		{
			name:    "reasoning stated only by an unusable entry listed second",
			entries: `{"id": "claude-sonnet-5", "supported_parameters": ["tools"]}, {"id": "claude-sonnet-5", "supported_parameters": ["reasoning"]}`,
			want:    thinking(false),
		},
		{
			name:    "an entry stating no parameters listed first",
			entries: `{"id": "claude-sonnet-5"}, {"id": "claude-sonnet-5", "supported_parameters": ["tools"]}`,
			want:    thinking(false),
		},
		{
			name:    "an entry stating no parameters listed second",
			entries: `{"id": "claude-sonnet-5", "supported_parameters": ["tools"]}, {"id": "claude-sonnet-5"}`,
			want:    thinking(false),
		},
		{
			name:    "an entry stating no parameters before a reasoning one",
			entries: `{"id": "claude-sonnet-5"}, {"id": "claude-sonnet-5", "supported_parameters": ["tools", "reasoning"]}`,
			want:    thinking(true),
		},
		{
			name: "structured outputs make an entry usable for reasoning",
			entries: `{"id": "claude-sonnet-5", "supported_parameters": ["structured_outputs"]},
				{"id": "claude-sonnet-5", "supported_parameters": ["structured_outputs", "reasoning"]}`,
			want: thinking(true),
		},

		{
			name:    "limits the gateway states",
			entries: `{"id": "gpt-4o", "supported_parameters": ["tools"], ` + stated + `}`,
			want:    limited(128000, 16384),
		},
		{name: "limits absent", entries: `{"id": "gpt-4o", "supported_parameters": ["tools"]}`, want: unlimited},
		{
			name:    "limits null",
			entries: `{"id": "gpt-4o", "supported_parameters": ["tools"], "max_input_tokens": null, "max_output_tokens": null}`,
			want:    unlimited,
		},
		{
			name:    "limits zero",
			entries: `{"id": "gpt-4o", "supported_parameters": ["tools"], "max_input_tokens": 0, "max_output_tokens": 0}`,
			want:    unlimited,
		},
		{
			name:    "limits negative",
			entries: `{"id": "gpt-4o", "supported_parameters": ["tools"], "max_input_tokens": -1, "max_output_tokens": -1}`,
			want:    unlimited,
		},
		{
			name: "a silent copy after a stated one keeps the stated limits",
			entries: `{"id": "gpt-4o", "supported_parameters": ["tools"], ` + stated + `},
				{"id": "gpt-4o", "supported_parameters": ["tools"]}`,
			want: limited(128000, 16384),
		},
		{
			name: "a silent copy before a stated one takes the stated limits",
			entries: `{"id": "gpt-4o", "supported_parameters": ["tools"]},
				{"id": "gpt-4o", "supported_parameters": ["tools"], ` + stated + `}`,
			want: limited(128000, 16384),
		},
		{
			name: "a tighter copy second wins",
			entries: `{"id": "gpt-4o", "supported_parameters": ["tools"], ` + stated + `},
				{"id": "gpt-4o", "supported_parameters": ["tools"], ` + tighter + `}`,
			want: limited(64000, 8192),
		},
		{
			name: "a tighter copy first wins",
			entries: `{"id": "gpt-4o", "supported_parameters": ["tools"], ` + tighter + `},
				{"id": "gpt-4o", "supported_parameters": ["tools"], ` + stated + `}`,
			want: limited(64000, 8192),
		},
		{
			name: "a copy stating max_model_len tighter than another's input limit wins",
			entries: `{"id": "gpt-4o", "supported_parameters": ["tools"], "max_input_tokens": 128000},
				{"id": "gpt-4o", "supported_parameters": ["tools"], "max_model_len": 32768}`,
			want: pconfig.ModelsConfig{{Name: "gpt-4o", Thinking: &no, ContextWindow: limit(32768)}},
		},

		{
			name:    "the window vLLM and SGLang state as max_model_len",
			entries: `{"id": "qwen3-32b", "object": "model", "max_model_len": 32768}`,
			want:    windowOnly("qwen3-32b", 32768),
		},
		{
			name:    "the window Groq states as context_window",
			entries: `{"id": "qwen3-32b", "object": "model", "context_window": 32768}`,
			want:    windowOnly("qwen3-32b", 32768),
		},
		{
			name:    "the window OpenRouter states as context_length",
			entries: `{"id": "qwen3-32b", "object": "model", "context_length": 32768}`,
			want:    windowOnly("qwen3-32b", 32768),
		},
		{
			name:    "an input limit wins over the total context stated beside it",
			entries: `{"id": "gpt-4o", "max_input_tokens": 128000, "context_length": 200000}`,
			want:    windowOnly("gpt-4o", 128000),
		},
		{
			name: "a limit that is not a number reads as unstated and keeps the rest of the entry",
			entries: `{"id": "qwen3-32b", "supported_parameters": ["tools"], "pricing": {"prompt": "0.000001", "completion": "0.000002"},
				"context_length": "32768", "max_model_len": 32768.0, "max_output_tokens": {"value": 8192}}`,
			want: pconfig.ModelsConfig{{
				Name: "qwen3-32b", Thinking: &no, Price: &pconfig.PriceInfo{Input: 1, Output: 2}, ContextWindow: limit(32768),
			}},
		},
		{
			name:    "a limit no real window reaches reads as unstated",
			entries: `{"id": "gpt-4o", "supported_parameters": ["tools"], "max_input_tokens": 1e10, "max_output_tokens": 2147483648}`,
			want:    unlimited,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			models, err := listedModels(t, tc.prefix, http.StatusOK, `{"data": [`+tc.entries+`]}`)
			require.NoError(t, err)
			assert.Equal(t, tc.want, models)
		})
	}
}

func TestLitellm_LoadModelsFromHTTP_ReportsWhyTheListingIsUnusable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  int
		body    string
		wantErr string
	}{
		{"an error status", http.StatusInternalServerError, `{"error": "internal error"}`, "unexpected status code: 500"},
		{"a body that is not JSON", http.StatusOK, `{invalid json}`, "failed to parse models response"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := listedModels(t, "", tc.status, tc.body)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
