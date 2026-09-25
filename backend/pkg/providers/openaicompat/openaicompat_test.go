package openaicompat

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestOpenaicompat_GetUsage_ReadsCountsCacheAndUpstreamCostFromTheClientsReport(t *testing.T) {
	prov, err := New(Spec{Type: provider.ProviderDeepSeek, APIKey: "k"}, &http.Client{}, nil,
		provider.DefaultProviderNameDeepSeek, nil)
	require.NoError(t, err)

	usage := prov.GetUsage(map[string]any{
		"PromptTokens": 100, "CompletionTokens": 50, "TotalTokens": 150, "ReasoningTokens": 20,
		"PromptCachedTokens": 30, "CacheReadInputTokens": 30, "CacheCreationInputTokens": 10,
		"UpstreamInferencePromptCost": 0.002, "UpstreamInferenceCompletionsCost": 0.004,
	})
	assert.Equal(t, pconfig.CallUsage{Input: 100, Output: 50, CacheRead: 30, CacheWrite: 10, CostInput: 0.002, CostOutput: 0.004}, usage)
}

func TestOpenaicompat_StructuredOutputFallback_SendsJSONObjectAndTheSchemaInThePrompt(t *testing.T) {
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(body))
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"x","object":"chat.completion","created":1,"model":"m",`+
			`"choices":[{"index":0,"message":{"role":"assistant","content":"{\"answer\":\"42\"}"},"finish_reason":"stop"}],`+
			`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`)
	}))
	defer srv.Close()

	schema := llms.WithStructuredOutput(llms.StructuredOutputConfig{
		Name:   "answer",
		Schema: json.RawMessage(`{"type":"object","properties":{"answer":{"type":"string"}},"required":["answer"],"additionalProperties":false}`),
	})
	human := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "question")}

	for _, fallback := range []bool{true, false} {
		prov, err := New(Spec{
			Type: provider.ProviderDeepSeek, APIKey: "k", Model: "deepseek-flash", ServerURL: srv.URL,
			StructuredOutputFallback: fallback,
		}, &http.Client{}, nil, provider.DefaultProviderNameDeepSeek, nil)
		require.NoError(t, err)

		_, err = prov.(*Provider).llm.GenerateContent(t.Context(), human, schema)
		if !fallback {
			var unsupported *llms.ErrStructuredOutputUnsupported
			require.ErrorAs(t, err, &unsupported, "without the flag the schema is refused before the request")
			continue
		}
		require.NoError(t, err)
		require.Len(t, bodies, 1)
		assert.Contains(t, bodies[0], `"response_format":{"type":"json_object"}`)
		assert.Contains(t, bodies[0], "JSON Schema:")
	}
	assert.Len(t, bodies, 1, "the refused call never reached the server")
}
