package bedrock

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRewriteAdaptiveThinkingBody(t *testing.T) {
	body := []byte(`{
		"additionalModelRequestFields": {
			"thinking": {
				"type": "enabled",
				"budget_tokens": 4096
			}
		},
		"inferenceConfig": {
			"maxTokens": 16384
		}
	}`)

	updatedBody, err := rewriteAdaptiveThinkingBody(body, "xhigh")
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(updatedBody, &payload))

	fields := payload["additionalModelRequestFields"].(map[string]any)
	thinking := fields["thinking"].(map[string]any)
	outputConfig := fields["output_config"].(map[string]any)

	assert.Equal(t, "adaptive", thinking["type"])
	assert.NotContains(t, thinking, "budget_tokens")
	assert.Equal(t, "xhigh", outputConfig["effort"])
	assert.Equal(t, map[string]any{"maxTokens": float64(16384)}, payload["inferenceConfig"])
}

func TestRewriteAdaptiveThinkingBodyWithoutThinking(t *testing.T) {
	body := []byte(`{"inferenceConfig":{"maxTokens":1024}}`)

	updatedBody, err := rewriteAdaptiveThinkingBody(body, "high")
	require.NoError(t, err)
	assert.JSONEq(t, string(body), string(updatedBody))
}
