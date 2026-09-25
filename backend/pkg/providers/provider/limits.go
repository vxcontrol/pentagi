package provider

import (
	"encoding/json"
	"fmt"

	"pentagi/pkg/cast"
	"pentagi/pkg/providers/pconfig"

	"github.com/vxcontrol/langchaingo/llms"
)

const (
	bytesPerToken     = 4
	minimumOutputRoom = 1024
)

type ModelLimits struct {
	ContextWindow   *int
	MaxOutputTokens *int
}

func ModelLimitsFor(prv Provider, opt pconfig.ProviderOptionsType) ModelLimits {
	if prv == nil {
		return ModelLimits{}
	}

	name := prv.Model(opt)
	if name == "" {
		return ModelLimits{}
	}

	for _, model := range prv.GetModels() {
		if model.Name == name {
			return ModelLimits{
				ContextWindow:   model.ContextWindow,
				MaxOutputTokens: model.MaxOutputTokens,
			}
		}
	}

	return ModelLimits{}
}

func ChainFitsWindow(messages []llms.MessageContent, tools []llms.Tool, window int) bool {
	if window <= 0 {
		return true
	}

	return estimateInputTokens(messages)+estimateToolTokens(tools) <= window-minimumOutputRoom
}

func ChainByteBudget(window int, tools []llms.Tool) int {
	room := window - minimumOutputRoom - estimateToolTokens(tools)
	if room <= 0 {
		return 0
	}

	return room * bytesPerToken
}

func capOutputTokens(options []llms.CallOption, ceiling *int) []llms.CallOption {
	if ceiling == nil || *ceiling <= 0 {
		return options
	}

	var applied llms.CallOptions
	for _, option := range options {
		option(&applied)
	}

	if applied.MaxTokens == nil || *applied.MaxTokens <= *ceiling {
		return options
	}

	return append(options, llms.WithMaxTokens(*ceiling))
}

type ContextWindowError struct {
	Model     string
	Window    int
	Estimated int
}

func (e *ContextWindowError) Error() string {
	return fmt.Sprintf(
		"request to %s is estimated at %d input tokens and leaves no room for an answer in its %d token context window",
		e.Model, e.Estimated, e.Window,
	)
}

func estimateInputTokens(messages []llms.MessageContent) int {
	size := 0
	for idx := range messages {
		size += cast.CalculateMessageSize(&messages[idx])
	}

	return size / bytesPerToken
}

// A vendor bounds the tool schemas a call carries as input, like any message.
func estimateToolTokens(tools []llms.Tool) int {
	if len(tools) == 0 {
		return 0
	}

	schemas, err := json.Marshal(tools)
	if err != nil {
		return 0
	}

	return len(schemas) / bytesPerToken
}

// The window is read as prompt plus answer, and an ask is lowered to exactly the room
// the estimate leaves: a prompt that tokenises above the estimate still overflows it.
func fitWithinWindow(
	model string,
	messages []llms.MessageContent,
	options []llms.CallOption,
	window *int,
) ([]llms.CallOption, error) {
	if window == nil || *window <= 0 {
		return options, nil
	}

	var applied llms.CallOptions
	for _, option := range options {
		option(&applied)
	}

	estimated := estimateInputTokens(messages) + estimateToolTokens(applied.Tools)

	room := *window - estimated
	if room < minimumOutputRoom {
		return nil, &ContextWindowError{Model: model, Window: *window, Estimated: estimated}
	}

	if applied.MaxTokens == nil || *applied.MaxTokens <= room {
		return options, nil
	}

	return append(options, llms.WithMaxTokens(room)), nil
}
