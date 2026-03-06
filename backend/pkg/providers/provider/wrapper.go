package provider

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"time"

	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/providers/pconfig"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/vxcontrol/langchaingo/llms"
)

const (
	MaxTooManyRequestsRetries = 10
	TooManyRequestsRetryDelay = 5 * time.Second
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

func wrapMetadataWithStopReason(metadata langfuse.Metadata, resp *llms.ContentResponse) langfuse.Metadata {
	if resp == nil || len(resp.Choices) == 0 {
		return metadata
	}

	newMetadata := make(langfuse.Metadata, len(metadata))
	maps.Copy(newMetadata, metadata)

	for _, choice := range resp.Choices {
		if choice.StopReason != "" {
			newMetadata["stop_reason"] = choice.StopReason
		}
	}

	return newMetadata
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
	metadata := buildMetadata(provider, opt, messages, options...)
	generation := observation.Generation(
		langfuse.WithGenerationName(fmt.Sprintf("%s-generation", provider.Type().String())),
		langfuse.WithGenerationMetadata(metadata),
		langfuse.WithGenerationInput(messages),
		langfuse.WithGenerationTools(extractToolsFromOptions(options...)),
		langfuse.WithGenerationModel(model),
		langfuse.WithGenerationModelParameters(langfuse.GetLangchainModelParameters(options)),
	)

	msg := llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
	}

	var (
		err  error
		resp *llms.ContentResponse
	)

	for idx := range MaxTooManyRequestsRetries {
		resp, err = llm.GenerateContent(ctx, []llms.MessageContent{msg}, options...)
		if err != nil {
			if isTooManyRequestsError(err) {
				_, observation = generation.Observation(ctx)
				observation.Event(
					langfuse.WithEventName(fmt.Sprintf("%s-generation-error", provider.Type().String())),
					langfuse.WithEventMetadata(wrapMetadataWithStopReason(metadata, resp)),
					langfuse.WithEventInput(messages),
					langfuse.WithEventStatus("TOO_MANY_REQUESTS"),
					langfuse.WithEventOutput(err.Error()),
					langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
				)
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				case <-time.After(TooManyRequestsRetryDelay + time.Duration(idx)*time.Second):
				}
				continue
			}
		}
		break
	}

	if err != nil {
		generation.End(
			langfuse.WithGenerationMetadata(wrapMetadataWithStopReason(metadata, resp)),
			langfuse.WithGenerationStatus(err.Error()),
			langfuse.WithGenerationLevel(langfuse.ObservationLevelError),
		)
		return "", err
	}

	choices := resp.Choices
	if len(choices) < 1 {
		err = fmt.Errorf("empty response from model")
		generation.End(
			langfuse.WithGenerationMetadata(wrapMetadataWithStopReason(metadata, resp)),
			langfuse.WithGenerationStatus(err.Error()),
			langfuse.WithGenerationLevel(langfuse.ObservationLevelError),
		)

		return "", err
	}

	if len(resp.Choices) == 1 {
		choice := resp.Choices[0]
		usage := provider.GetUsage(choice.GenerationInfo)
		usage.UpdateCost(provider.GetPriceInfo(opt))

		generation.End(
			langfuse.WithGenerationMetadata(wrapMetadataWithStopReason(metadata, resp)),
			langfuse.WithGenerationOutput(choice),
			langfuse.WithGenerationStatus("success"),
			langfuse.WithGenerationUsage(&langfuse.GenerationUsage{
				Input:      int(usage.Input),
				Output:     int(usage.Output),
				InputCost:  getUsageCost(usage.CostInput),
				OutputCost: getUsageCost(usage.CostOutput),
				Unit:       langfuse.GenerationUsageUnitTokens,
			}),
		)

		return choice.Content, nil
	}

	var usage pconfig.CallUsage
	choicesOutput := make([]string, 0, len(resp.Choices))
	for _, choice := range resp.Choices {
		usage.Merge(provider.GetUsage(choice.GenerationInfo))
		choicesOutput = append(choicesOutput, choice.Content)
	}

	usage.UpdateCost(provider.GetPriceInfo(opt))

	respOutput := strings.Join(choicesOutput, "\n-----\n")
	generation.End(
		langfuse.WithGenerationMetadata(wrapMetadataWithStopReason(metadata, resp)),
		langfuse.WithGenerationOutput(resp.Choices),
		langfuse.WithGenerationStatus("success"),
		langfuse.WithGenerationUsage(&langfuse.GenerationUsage{
			Input:      int(usage.Input),
			Output:     int(usage.Output),
			InputCost:  getUsageCost(usage.CostInput),
			OutputCost: getUsageCost(usage.CostOutput),
			Unit:       langfuse.GenerationUsageUnitTokens,
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
	metadata := buildMetadata(provider, opt, messages, options...)
	generation := observation.Generation(
		langfuse.WithGenerationName(fmt.Sprintf("%s-generation-ex", provider.Type().String())),
		langfuse.WithGenerationMetadata(metadata),
		langfuse.WithGenerationInput(messages),
		langfuse.WithGenerationTools(extractToolsFromOptions(options...)),
		langfuse.WithGenerationModel(provider.Model(opt)),
		langfuse.WithGenerationModelParameters(langfuse.GetLangchainModelParameters(options)),
	)

	var (
		err  error
		resp *llms.ContentResponse
	)

	for idx := range MaxTooManyRequestsRetries {
		resp, err = fn(ctx, messages, options...)
		if err != nil {
			if isTooManyRequestsError(err) {
				_, observation = generation.Observation(ctx)
				observation.Event(
					langfuse.WithEventName(fmt.Sprintf("%s-generation-error", provider.Type().String())),
					langfuse.WithEventMetadata(wrapMetadataWithStopReason(metadata, resp)),
					langfuse.WithEventInput(messages),
					langfuse.WithEventStatus("TOO_MANY_REQUESTS"),
					langfuse.WithEventOutput(err.Error()),
					langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
				)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(TooManyRequestsRetryDelay + time.Duration(idx)*time.Second):
				}
				continue
			}
		}
		break
	}

	if err != nil {
		generation.End(
			langfuse.WithGenerationMetadata(wrapMetadataWithStopReason(metadata, resp)),
			langfuse.WithGenerationStatus(err.Error()),
			langfuse.WithGenerationLevel(langfuse.ObservationLevelError),
		)
		return nil, err
	}

	if len(resp.Choices) < 1 {
		err = fmt.Errorf("empty response from model")
		generation.End(
			langfuse.WithGenerationMetadata(wrapMetadataWithStopReason(metadata, resp)),
			langfuse.WithGenerationStatus(err.Error()),
			langfuse.WithGenerationLevel(langfuse.ObservationLevelError),
		)
		return nil, err
	}

	if len(resp.Choices) == 1 {
		choice := resp.Choices[0]
		usage := provider.GetUsage(choice.GenerationInfo)
		usage.UpdateCost(provider.GetPriceInfo(opt))

		generation.End(
			langfuse.WithGenerationMetadata(wrapMetadataWithStopReason(metadata, resp)),
			langfuse.WithGenerationOutput(choice),
			langfuse.WithGenerationStatus("success"),
			langfuse.WithGenerationUsage(&langfuse.GenerationUsage{
				Input:      int(usage.Input),
				Output:     int(usage.Output),
				InputCost:  getUsageCost(usage.CostInput),
				OutputCost: getUsageCost(usage.CostOutput),
				Unit:       langfuse.GenerationUsageUnitTokens,
			}),
		)

		return resp, nil
	}

	var usage pconfig.CallUsage
	for _, choice := range resp.Choices {
		usage.Merge(provider.GetUsage(choice.GenerationInfo))
	}

	usage.UpdateCost(provider.GetPriceInfo(opt))

	generation.End(
		langfuse.WithGenerationMetadata(wrapMetadataWithStopReason(metadata, resp)),
		langfuse.WithGenerationOutput(resp.Choices),
		langfuse.WithGenerationStatus("success"),
		langfuse.WithGenerationUsage(&langfuse.GenerationUsage{
			Input:      int(usage.Input),
			Output:     int(usage.Output),
			InputCost:  getUsageCost(usage.CostInput),
			OutputCost: getUsageCost(usage.CostOutput),
			Unit:       langfuse.GenerationUsageUnitTokens,
		}),
	)

	return resp, nil
}

func isTooManyRequestsError(err error) bool {
	if err == nil {
		return false
	}

	for errNested := err; errNested != nil; errNested = errors.Unwrap(errNested) {
		if errResp, ok := errNested.(*awshttp.ResponseError); ok {
			return errResp.Response.StatusCode == http.StatusTooManyRequests
		}
		if errThrottling, ok := errNested.(*types.ThrottlingException); ok && errThrottling.Message != nil {
			return strings.Contains(strings.ToLower(*errThrottling.Message), "too many requests")
		}
	}

	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "statuscode: 429") {
		return true
	}
	if strings.Contains(errStr, "toomanyrequests") || strings.Contains(errStr, "too many requests") {
		return true
	}

	return false
}

func getUsageCost(usage float64) *float64 {
	if usage == 0.0 {
		return nil
	}

	return &usage
}

func extractToolsFromOptions(options ...llms.CallOption) []llms.Tool {
	opts := llms.CallOptions{}
	for _, option := range options {
		option(&opts)
	}

	return opts.Tools
}
