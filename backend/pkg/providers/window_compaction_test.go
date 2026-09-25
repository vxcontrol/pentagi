package providers

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester/mock"

	"github.com/stretchr/testify/assert"
	"github.com/vxcontrol/langchaingo/llms"
)

func windowCompactionWindow(value int) *int { return &value }

func TestWindowCompaction_CompactChainForWindow_SummarisesOnlyAChainThatOutgrewTheWindow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		window        *int
		turns         int
		tools         []llms.Tool
		failing       bool
		wantDone      bool
		wantSummaries bool
	}{
		{name: "a chain that fits is left alone", window: windowCompactionWindow(200000), turns: 4},
		{name: "an unknown window leaves the chain alone", window: nil, turns: 40},
		{name: "a window with no room for an answer leaves the chain alone", window: windowCompactionWindow(512), turns: 40},
		{name: "a chain far above the window is summarised", window: windowCompactionWindow(8000), turns: 40, wantDone: true, wantSummaries: true},
		{name: "a failed summarisation keeps the chain", window: windowCompactionWindow(8000), turns: 40, failing: true, wantSummaries: true},
		{name: "a chain that fits when it carries no tools is left alone", window: windowCompactionWindow(8000), turns: 3},
		{
			name: "a chain that fits only without its tools is summarised", window: windowCompactionWindow(8000), turns: 3,
			tools: windowCompactionTools(20000), wantDone: true, wantSummaries: true,
		},
		{
			name: "a window the tool schemas fill leaves the chain alone", window: windowCompactionWindow(6049), turns: 3,
			tools: windowCompactionTools(20000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			prv := mock.NewProvider(provider.ProviderOpenAI, "openai", "small-model")
			prv.SetModels(pconfig.ModelsConfig{{Name: "small-model", ContextWindow: tt.window}})

			var calls atomic.Int64
			handler := func(context.Context, string) (string, error) {
				calls.Add(1)
				if tt.failing {
					return "", errors.New("summarizer is down")
				}
				return "short summary", nil
			}

			chain := windowCompactionChain(tt.turns, 4000)
			compacted, done := compactChainForWindow(
				context.Background(), prv, pconfig.OptionsTypeSimple, chain, tt.tools, handler, "call_%d",
			)

			assert.Equal(t, tt.wantDone, done)
			if tt.wantDone {
				assert.Less(t, windowCompactionBytes(compacted), windowCompactionBytes(chain))
			} else {
				assert.Equal(t, windowCompactionChain(tt.turns, 4000), compacted)
			}
			assert.Equal(t, tt.wantSummaries, calls.Load() > 0, "summarisation calls made: %d", calls.Load())
		})
	}
}
