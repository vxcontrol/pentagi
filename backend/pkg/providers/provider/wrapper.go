package provider

import (
	"context"
	"fmt"
	"strings"

	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/providers/pconfig"

	"github.com/vxcontrol/langchaingo/llms"
)

type GenerateContentFunc func(
	ctx context.Context,
	messages []llms.MessageContent,
	options ...llms.CallOption,
) (*llms.ContentResponse, error)

func buildMetadata(
	provider Provider,
	opt pconfig.ProviderOptionsType,
	messages []llms.MessageContent,
	options ...llms.CallOption,
) langfuse.Metadata {
	opts := llms.CallOptions{}
	for _, option := range options {
		option(&opts)
	}

	toolNames := make([]string, 0, len(opts.Tools))
	for _, tool := range opts.Tools {
		toolNames = append(toolNames, tool.Function.Name)
	}

	var (
		totalInputSize        int
		totalOutputSize       int
		totalSystemPromptSize int
		totalToolCallsSize    int
		totalMessagesSize     int
	)
	for _, message := range messages {
		partsSize := 0
		for _, part := range message.Parts {
			switch part := part.(type) {
			case llms.TextContent:
				partsSize += len(part.Text)
			case llms.ImageURLContent:
				partsSize += len(part.Detail) + len(part.URL)
			case llms.BinaryContent:
				partsSize += len(part.MIMEType) + len(part.Data)
			case llms.ToolCall:
				if part.FunctionCall != nil {
					partsSize += len(part.FunctionCall.Name) + len(part.FunctionCall.Arguments)
				}
			case llms.ToolCallResponse:
				partsSize += len(part.Name) + len(part.Content)
			}
		}

		totalMessagesSize += partsSize

		switch message.Role {
		case llms.ChatMessageTypeHuman:
			totalInputSize += partsSize
		case llms.ChatMessageTypeAI:
			totalOutputSize += partsSize
		case llms.ChatMessageTypeSystem:
			totalSystemPromptSize += partsSize
		case llms.ChatMessageTypeTool:
			totalToolCallsSize += partsSize
		}
	}

	return langfuse.Metadata{
		"provider":              provider.Type().String(),
		"agent":                 opt,
		"tools":                 toolNames,
		"messages_len":          len(messages),
		"messages_size":         totalMessagesSize,
		"has_system_prompt":     totalSystemPromptSize != 0,
		"system_prompt_size":    totalSystemPromptSize,
		"total_input_size":      totalInputSize,
		"total_output_size":     totalOutputSize,
		"total_tool_calls_size": totalToolCallsSize,
	}
}

func WrapGenerateFromSinglePrompt(
	ctx context.Context,
	provider Provider,
	opt pconfig.ProviderOptionsType,
	llm llms.Model,
	prompt string,
	options ...llms.CallOption,
) (string, error) {
	ctx, observation := obs.Observer.NewObservation(ctx)
	model := provider.Model(opt)
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}
	generation := observation.Generation(
		langfuse.WithStartGenerationName(fmt.Sprintf("%s-generation", provider.Type().String())),
		langfuse.WithStartGenerationMetadata(buildMetadata(provider, opt, messages, options...)),
		langfuse.WithStartGenerationInput(messages),
		langfuse.WithStartGenerationModel(model),
		langfuse.WithStartGenerationModelParameters(langfuse.GetLangchainModelParameters(options)),
	)

	msg := llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
	}

	resp, err := llm.GenerateContent(ctx, []llms.MessageContent{msg}, options...)
	if err != nil {
		generation.End(
			langfuse.WithEndGenerationStatus(err.Error()),
			langfuse.WithEndGenerationLevel(langfuse.ObservationLevelError),
		)
		return "", err
	}

	choices := resp.Choices
	if len(choices) < 1 {
		err = fmt.Errorf("empty response from model")
		generation.End(
			langfuse.WithEndGenerationStatus(err.Error()),
			langfuse.WithEndGenerationLevel(langfuse.ObservationLevelError),
		)

		return "", err
	}

	if len(resp.Choices) == 1 {
		choice := resp.Choices[0]
		input, output := provider.GetUsage(choice.GenerationInfo)

		generation.End(
			langfuse.WithEndGenerationOutput(choice),
			langfuse.WithEndGenerationStatus("success"),
			langfuse.WithEndGenerationUsage(&langfuse.GenerationUsage{
				Input:  int(input),
				Output: int(output),
			}),
		)

		return choice.Content, nil
	}

	choicesOutput := make([]string, 0, len(resp.Choices))
	totalInput, totalOutput := int64(0), int64(0)
	for _, choice := range resp.Choices {
		input, output := provider.GetUsage(choice.GenerationInfo)
		if input > 0 {
			totalInput = input
		}
		if output > 0 {
			totalOutput = output
		}
		choicesOutput = append(choicesOutput, choice.Content)
	}

	respOutput := strings.Join(choicesOutput, "\n-----\n")
	generation.End(
		langfuse.WithEndGenerationOutput(resp.Choices),
		langfuse.WithEndGenerationStatus("success"),
		langfuse.WithEndGenerationUsage(&langfuse.GenerationUsage{
			Input:  int(totalInput),
			Output: int(totalOutput),
		}),
	)

	return respOutput, nil
}

func WrapGenerateContent(
	ctx context.Context,
	provider Provider,
	opt pconfig.ProviderOptionsType,
	fn GenerateContentFunc,
	messages []llms.MessageContent,
	options ...llms.CallOption,
) (*llms.ContentResponse, error) {
	ctx, observation := obs.Observer.NewObservation(ctx)
	generation := observation.Generation(
		langfuse.WithStartGenerationName(fmt.Sprintf("%s-generation-ex", provider.Type().String())),
		langfuse.WithStartGenerationMetadata(buildMetadata(provider, opt, messages, options...)),
		langfuse.WithStartGenerationInput(messages),
		langfuse.WithStartGenerationModel(provider.Model(opt)),
		langfuse.WithStartGenerationModelParameters(langfuse.GetLangchainModelParameters(options)),
	)

	resp, err := fn(ctx, messages, options...)
	if err != nil {
		generation.End(
			langfuse.WithEndGenerationStatus(err.Error()),
			langfuse.WithEndGenerationLevel(langfuse.ObservationLevelError),
		)
		return nil, err
	}

	if len(resp.Choices) == 1 {
		choice := resp.Choices[0]
		input, output := provider.GetUsage(choice.GenerationInfo)

		generation.End(
			langfuse.WithEndGenerationOutput(choice),
			langfuse.WithEndGenerationStatus("success"),
			langfuse.WithEndGenerationUsage(&langfuse.GenerationUsage{
				Input:  int(input),
				Output: int(output),
			}),
		)

		return resp, nil
	}

	totalInput, totalOutput := int64(0), int64(0)
	for _, choice := range resp.Choices {
		input, output := provider.GetUsage(choice.GenerationInfo)
		if input > 0 {
			totalInput = input
		}
		if output > 0 {
			totalOutput = output
		}
	}

	generation.End(
		langfuse.WithEndGenerationOutput(resp.Choices),
		langfuse.WithEndGenerationStatus("success"),
		langfuse.WithEndGenerationUsage(&langfuse.GenerationUsage{
			Input:  int(totalInput),
			Output: int(totalOutput),
		}),
	)

	return resp, nil
}
