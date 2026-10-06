package smallmodel

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vxcontrol/langchaingo/llms"
)

func msg(text string) llms.MessageContent {
	return llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextContent{Text: text}},
	}
}

func TestBudget_TokenLimit_IsWindowTimesPercent(t *testing.T) {
	b := Budget{CtxWindow: 32000, Percent: 40}
	assert.Equal(t, 12800, b.TokenLimit())
}

func TestBudget_Exceeded_UsesByteEstimateWhenNoTokenizer(t *testing.T) {
	b := Budget{CtxWindow: 1000, Percent: 40, BytesPerToken: 3}
	limitBytes := b.TokenLimit() * b.BytesPerToken // 400 tokens -> 1200 bytes

	under := []llms.MessageContent{msg(strings.Repeat("a", limitBytes-30))}
	over := []llms.MessageContent{msg(strings.Repeat("a", limitBytes+30))}

	assert.False(t, b.Exceeded(under))
	assert.True(t, b.Exceeded(over))
}

func TestBudget_Tokens_PrefersTokenizerOverEstimate(t *testing.T) {
	b := Budget{
		CtxWindow:     1000,
		Percent:       40,
		BytesPerToken: 3,
		CountTokens: func([]llms.MessageContent) (int, error) {
			return 7, nil
		},
	}
	assert.Equal(t, 7, b.Tokens([]llms.MessageContent{msg(strings.Repeat("a", 9000))}))
}

func TestBudget_Tokens_FallsBackWhenTokenizerErrors(t *testing.T) {
	b := Budget{
		CtxWindow:     1000,
		Percent:       40,
		BytesPerToken: 3,
		CountTokens: func([]llms.MessageContent) (int, error) {
			return 0, errors.New("tokenizer unavailable")
		},
	}
	assert.Equal(t, 300, b.Tokens([]llms.MessageContent{msg(strings.Repeat("a", 900))}))
}

func TestBudget_Exceeded_DisabledWhenWindowNonPositive(t *testing.T) {
	b := Budget{CtxWindow: 0, Percent: 40, BytesPerToken: 3}
	assert.False(t, b.Exceeded([]llms.MessageContent{msg(strings.Repeat("a", 1_000_000))}))
}
