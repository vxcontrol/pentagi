package providers

import (
	"context"

	"pentagi/pkg/csum"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/tools"

	"github.com/vxcontrol/langchaingo/llms"
)

func compactChainForWindow(
	ctx context.Context,
	prv provider.Provider,
	opt pconfig.ProviderOptionsType,
	chain []llms.MessageContent,
	schemas []llms.Tool,
	handler tools.SummarizeHandler,
	tcIDTemplate string,
) ([]llms.MessageContent, bool) {
	window := provider.ModelLimitsFor(prv, opt).ContextWindow
	if window == nil || provider.ChainFitsWindow(chain, schemas, *window) {
		return chain, false
	}

	budget := provider.ChainByteBudget(*window, schemas)
	if budget <= 0 {
		return chain, false
	}

	compacted, err := csum.NewSummarizer(csum.SummarizerConfig{
		PreserveLast:   true,
		UseQA:          true,
		LastSecBytes:   budget / 2,
		MaxBPBytes:     budget / 4,
		MaxQABytes:     budget,
		KeepQASections: 1,
	}).SummarizeChain(ctx, handler, chain, tcIDTemplate)
	if err != nil {
		return chain, false
	}

	return compacted, true
}
