package bedrock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"pentagi/pkg/providers/pconfig"

	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/vxcontrol/langchaingo/llms"
)

type adaptiveThinkingEffortContextKey struct{}

func withAdaptiveThinkingEffort(ctx context.Context, effort string) context.Context {
	if effort == "" {
		effort = string(llms.ReasoningHigh)
	}

	return context.WithValue(ctx, adaptiveThinkingEffortContextKey{}, effort)
}

func adaptiveThinkingEffortFromContext(ctx context.Context) (string, bool) {
	effort, ok := ctx.Value(adaptiveThinkingEffortContextKey{}).(string)
	return effort, ok && effort != ""
}

func addAdaptiveThinkingMiddleware(stack *middleware.Stack) error {
	return stack.Build.Add(middleware.BuildMiddlewareFunc(
		"PentAGIBedrockAdaptiveThinking",
		func(ctx context.Context, in middleware.BuildInput, next middleware.BuildHandler) (
			middleware.BuildOutput,
			middleware.Metadata,
			error,
		) {
			effort, ok := adaptiveThinkingEffortFromContext(ctx)
			if !ok {
				return next.HandleBuild(ctx, in)
			}

			req, ok := in.Request.(*smithyhttp.Request)
			if !ok || req.GetStream() == nil {
				return next.HandleBuild(ctx, in)
			}

			body, err := io.ReadAll(req.GetStream())
			if err != nil {
				return middleware.BuildOutput{}, middleware.Metadata{}, fmt.Errorf("read Bedrock request body: %w", err)
			}

			updatedBody, err := rewriteAdaptiveThinkingBody(body, effort)
			if err != nil {
				return middleware.BuildOutput{}, middleware.Metadata{}, err
			}

			updatedReq, err := req.SetStream(bytes.NewReader(updatedBody))
			if err != nil {
				return middleware.BuildOutput{}, middleware.Metadata{}, fmt.Errorf("replace Bedrock request body: %w", err)
			}
			updatedReq.ContentLength = int64(len(updatedBody))
			updatedReq.Header.Set("Content-Length", strconv.Itoa(len(updatedBody)))
			in.Request = updatedReq

			return next.HandleBuild(ctx, in)
		},
	), middleware.After)
}

func rewriteAdaptiveThinkingBody(body []byte, effort string) ([]byte, error) {
	// langchaingo currently emits budget-based Anthropic thinking for Bedrock;
	// Claude adaptive thinking expects thinking.type=adaptive plus output_config.effort.
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode Bedrock request body: %w", err)
	}

	fields, ok := payload["additionalModelRequestFields"].(map[string]any)
	if !ok {
		return body, nil
	}

	thinking, ok := fields["thinking"].(map[string]any)
	if !ok {
		return body, nil
	}

	thinking["type"] = "adaptive"
	delete(thinking, "budget_tokens")
	fields["output_config"] = map[string]any{
		"effort": effort,
	}

	updatedBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode Bedrock request body: %w", err)
	}

	return updatedBody, nil
}

func (p *bedrockProvider) prepareCallOptions(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	options []llms.CallOption,
) (context.Context, []llms.CallOption) {
	reasoning, ok := p.reasoningConfigForType(opt)
	if !ok || reasoning.EffectiveMode() != pconfig.ReasoningModeAdaptive {
		return ctx, options
	}

	effort := string(reasoning.Effort)
	ctx = withAdaptiveThinkingEffort(ctx, effort)
	options = append(options, llms.WithReasoning(llms.ReasoningHigh, 0))

	return ctx, options
}

func (p *bedrockProvider) reasoningConfigForType(opt pconfig.ProviderOptionsType) (pconfig.ReasoningConfig, bool) {
	agentConfig := p.agentConfigForType(opt)
	if agentConfig == nil || agentConfig.Reasoning.IsZero() {
		return pconfig.ReasoningConfig{}, false
	}

	return agentConfig.Reasoning, true
}

func (p *bedrockProvider) agentConfigForType(opt pconfig.ProviderOptionsType) *pconfig.AgentConfig {
	if p == nil || p.providerConfig == nil {
		return nil
	}

	switch opt {
	case pconfig.OptionsTypeSimple:
		return p.providerConfig.Simple
	case pconfig.OptionsTypeSimpleJSON:
		if p.providerConfig.SimpleJSON != nil {
			return p.providerConfig.SimpleJSON
		}
		return p.providerConfig.Simple
	case pconfig.OptionsTypePrimaryAgent:
		return p.providerConfig.PrimaryAgent
	case pconfig.OptionsTypeAssistant:
		if p.providerConfig.Assistant != nil {
			return p.providerConfig.Assistant
		}
		return p.providerConfig.PrimaryAgent
	case pconfig.OptionsTypeGenerator:
		return p.providerConfig.Generator
	case pconfig.OptionsTypeRefiner:
		return p.providerConfig.Refiner
	case pconfig.OptionsTypeAdviser:
		return p.providerConfig.Adviser
	case pconfig.OptionsTypeReflector:
		return p.providerConfig.Reflector
	case pconfig.OptionsTypeSearcher:
		return p.providerConfig.Searcher
	case pconfig.OptionsTypeEnricher:
		return p.providerConfig.Enricher
	case pconfig.OptionsTypeCoder:
		return p.providerConfig.Coder
	case pconfig.OptionsTypeInstaller:
		return p.providerConfig.Installer
	case pconfig.OptionsTypePentester:
		return p.providerConfig.Pentester
	default:
		return nil
	}
}
