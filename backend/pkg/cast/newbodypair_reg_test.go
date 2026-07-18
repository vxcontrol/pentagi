package cast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestNewBodyPair_DropsMultipleNilFunctionCallToolCalls(t *testing.T) {
	aiMsg := &llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{
			llms.ToolCall{ID: "a", Type: "function", FunctionCall: nil},
			llms.ToolCall{ID: "b", Type: "function", FunctionCall: nil},
			llms.TextContent{Text: "keep me"},
		},
	}

	NewBodyPair(aiMsg, nil)

	assert.Len(t, aiMsg.Parts, 1, "both nil-FunctionCall tool calls should be removed, text kept")
	if assert.Len(t, aiMsg.Parts, 1) {
		txt, ok := aiMsg.Parts[0].(llms.TextContent)
		assert.True(t, ok, "surviving part should be the TextContent, got %T", aiMsg.Parts[0])
		assert.Equal(t, "keep me", txt.Text)
	}
}
