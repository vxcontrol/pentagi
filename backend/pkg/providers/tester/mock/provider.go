package mock

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/templates"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

// Provider implements provider.Provider for testing purposes
type Provider struct {
	providerType   provider.ProviderType
	providerName   provider.ProviderName
	modelName      string
	responses      map[string]any // key -> response mapping
	defaultResp    string
	streamingDelay time.Duration
	providerConfig *pconfig.ProviderConfig
	models         pconfig.ModelsConfig
	rawConfig      []byte

	// sequence, when set via SetSequentialResponses, makes CallWithTools
	// ignore content-based matching and return each response strictly in
	// call order instead. Needed for multi-turn tool-calling scenarios
	// where a tool call is answered and the model is called again with the
	// same TextContent (a tool response carries no text), so content-based
	// matching alone can't tell the calls apart.
	sequence      []any
	sequenceCalls atomic.Int32
}

// ResponseConfig configures mock responses
type ResponseConfig struct {
	Key      string // Request identifier (prompt/message content)
	Response any    // Response (string, *llms.ContentResponse, or error)
}

// NewProvider creates a new mock provider
func NewProvider(providerType provider.ProviderType, providerName provider.ProviderName, modelName string) *Provider {
	return &Provider{
		providerType:   providerType,
		providerName:   providerName,
		modelName:      modelName,
		responses:      make(map[string]any),
		defaultResp:    "Mock response",
		streamingDelay: time.Millisecond * 10,
	}
}

// SetResponses configures responses for specific requests
func (p *Provider) SetResponses(configs []ResponseConfig) {
	for _, config := range configs {
		p.responses[config.Key] = config.Response
	}
}

// SetDefaultResponse sets fallback response for unmatched requests
func (p *Provider) SetDefaultResponse(response string) {
	p.defaultResp = response
}

// SetSequentialResponses configures CallWithTools to return responses
// strictly in order, one per call, bypassing content-based matching
// entirely. The last response repeats for any call beyond len(responses).
// Each element is handled exactly like a ResponseConfig.Response value
// (string, *llms.ContentResponse, or error).
func (p *Provider) SetSequentialResponses(responses ...any) {
	p.sequence = responses
	p.sequenceCalls.Store(0)
}

// SetStreamingDelay configures delay between streaming chunks
func (p *Provider) SetStreamingDelay(delay time.Duration) {
	p.streamingDelay = delay
}

// SetProviderConfig configures the per-agent config GetProviderConfig
// returns, so capability gating (tester.capabilitySupported) can be exercised
// against a mock the same way it reads a real provider's loaded YAML — e.g.
// setting AgentConfig.Reasoning.Mode to off/adaptive for a given agent type.
func (p *Provider) SetProviderConfig(pc *pconfig.ProviderConfig) {
	p.providerConfig = pc
}

// SetModels configures the model catalog GetModels returns, so tests can
// simulate an adaptive-only (or otherwise reasoning-classified) model without
// depending on the SDK's real model-name tables.
func (p *Provider) SetModels(models pconfig.ModelsConfig) {
	p.models = models
}

// Type implements provider.Provider
func (p *Provider) Type() provider.ProviderType {
	return p.providerType
}

// Name implements provider.Provider
func (p *Provider) Name() provider.ProviderName {
	return p.providerName
}

// Model implements provider.Provider
func (p *Provider) Model(opt pconfig.ProviderOptionsType) string {
	return p.modelName
}

// ModelWithPrefix implements provider.Provider
func (p *Provider) ModelWithPrefix(opt pconfig.ProviderOptionsType) string {
	return p.Model(opt)
}

// GetUsage implements provider.Provider
func (p *Provider) GetUsage(info map[string]any) pconfig.CallUsage {
	return pconfig.CallUsage{Input: 100, Output: 50} // Mock token counts
}

// GetModels implements provider.Provider
func (p *Provider) GetModels() pconfig.ModelsConfig {
	if p.models != nil {
		return p.models
	}
	return pconfig.ModelsConfig{}
}

// GetToolCallIDTemplate implements provider.Provider
func (p *Provider) GetToolCallIDTemplate(ctx context.Context, prompter templates.Prompter) (string, error) {
	return "toolu_{r:24:b}", nil
}

// Call implements provider.Provider for simple prompt calls
func (p *Provider) Call(ctx context.Context, opt pconfig.ProviderOptionsType, prompt string) (string, error) {
	// Look for exact match
	if resp, ok := p.responses[prompt]; ok {
		return p.handleResponse(resp)
	}

	// Look for partial match
	for key, resp := range p.responses {
		if strings.Contains(prompt, key) {
			return p.handleResponse(resp)
		}
	}

	return p.defaultResp, nil
}

// CallEx implements provider.Provider for message-based calls
func (p *Provider) CallEx(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	chain []llms.MessageContent,
	streamCb streaming.Callback,
) (*llms.ContentResponse, error) {
	// Extract content for matching
	var content string
	for _, msg := range chain {
		for _, part := range msg.Parts {
			if textContent, ok := part.(llms.TextContent); ok {
				content += textContent.Text + " "
			}
		}
	}
	content = strings.TrimSpace(content)

	// Look for response
	var respInterface any
	if resp, ok := p.responses[content]; ok {
		respInterface = resp
	} else {
		// Look for partial match
		for key, resp := range p.responses {
			if strings.Contains(content, key) {
				respInterface = resp
				break
			}
		}
	}

	if respInterface == nil {
		respInterface = p.defaultResp
	}

	// Handle streaming if callback provided
	if streamCb != nil {
		return p.handleStreamingResponse(ctx, respInterface, streamCb)
	}

	return p.handleContentResponse(respInterface)
}

// CallWithTools implements provider.Provider for tool-calling
func (p *Provider) CallWithTools(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	chain []llms.MessageContent,
	tools []llms.Tool,
	streamCb streaming.Callback,
) (*llms.ContentResponse, error) {
	if p.sequence != nil {
		idx := int(p.sequenceCalls.Add(1)) - 1
		if idx >= len(p.sequence) {
			idx = len(p.sequence) - 1 // repeat the last configured response past the end
		}

		if streamCb != nil {
			return p.handleStreamingResponse(ctx, p.sequence[idx], streamCb)
		}
		return p.handleContentResponse(p.sequence[idx])
	}

	// Extract content for matching
	var content string
	for _, msg := range chain {
		for _, part := range msg.Parts {
			if textContent, ok := part.(llms.TextContent); ok {
				content += textContent.Text + " "
			}
		}
	}
	content = strings.TrimSpace(content)

	// Look for tool-specific response
	var respInterface any
	toolKey := fmt.Sprintf("tools:%s", content)
	if resp, ok := p.responses[toolKey]; ok {
		respInterface = resp
	} else if resp, ok := p.responses[content]; ok {
		respInterface = resp
	} else {
		// Create default tool call response
		if len(tools) > 0 {
			respInterface = &llms.ContentResponse{
				Choices: []*llms.ContentChoice{
					{
						Content: "",
						ToolCalls: []llms.ToolCall{
							{
								FunctionCall: &llms.FunctionCall{
									Name:      tools[0].Function.Name,
									Arguments: `{"message": "mock response"}`,
								},
							},
						},
					},
				},
			}
		} else {
			respInterface = p.defaultResp
		}
	}

	// Handle streaming if callback provided
	if streamCb != nil {
		return p.handleStreamingResponse(ctx, respInterface, streamCb)
	}

	return p.handleContentResponse(respInterface)
}

// CallWithExtraOptions implements provider.Provider for completeness. The
// capability test pipeline no longer routes through it for adaptive
// thinking/reasoning off (those are gated to configs that already produce the
// behavior via the plain CallWithTools/CallEx path below, so tests should set
// up the mock's canned response accordingly rather than relying on this
// method to synthesize it); it is reachable by the CapabilityStructuredOutput
// path. Uses partial-match lookup like CallEx (CallWithTools only matches
// exact content, too strict for multi-sentence prompts).
func (p *Provider) CallWithExtraOptions(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	chain []llms.MessageContent,
	tools []llms.Tool,
	streamCb streaming.Callback,
	extra ...llms.CallOption,
) (*llms.ContentResponse, error) {
	var applied llms.CallOptions
	for _, o := range extra {
		o(&applied)
	}

	var content string
	for _, msg := range chain {
		for _, part := range msg.Parts {
			if textContent, ok := part.(llms.TextContent); ok {
				content += textContent.Text + " "
			}
		}
	}
	content = strings.TrimSpace(content)

	var respInterface any
	if resp, ok := p.responses[content]; ok {
		respInterface = resp
	} else {
		for key, resp := range p.responses {
			if strings.Contains(content, key) {
				respInterface = resp
				break
			}
		}
	}
	if respInterface == nil {
		respInterface = p.defaultResp
	}

	resp, err := p.handleContentResponse(respInterface)
	if err != nil {
		return nil, err
	}

	switch {
	case applied.Reasoning != nil && applied.Reasoning.IsDisabled():
		for _, choice := range resp.Choices {
			choice.Reasoning = nil
		}
	case applied.Reasoning != nil && applied.Reasoning.Adaptive:
		for _, choice := range resp.Choices {
			if choice.Reasoning.IsEmpty() {
				choice.Reasoning = &reasoning.ContentReasoning{Content: "mock adaptive reasoning trace"}
			}
		}
	}

	if streamCb == nil {
		return resp, nil
	}

	return p.handleStreamingResponse(ctx, resp, streamCb)
}

// GetRawConfig implements provider.Provider
func (p *Provider) GetRawConfig() []byte {
	if p.rawConfig != nil {
		return p.rawConfig
	}
	return []byte(`{"mock": true}`)
}

// SetRawConfig overrides what GetRawConfig returns, so a test can build two
// providers that share a name but not a configuration — exactly what a user
// provider named like a built-in one produces once it is deleted or renamed.
func (p *Provider) SetRawConfig(raw []byte) {
	p.rawConfig = raw
}

// GetProviderConfig implements provider.Provider
func (p *Provider) GetProviderConfig() *pconfig.ProviderConfig {
	if p.providerConfig != nil {
		return p.providerConfig
	}
	return &pconfig.ProviderConfig{}
}

// GetPriceInfo implements provider.Provider
func (p *Provider) GetPriceInfo(opt pconfig.ProviderOptionsType) *pconfig.PriceInfo {
	return &pconfig.PriceInfo{
		Input:  0.01,
		Output: 0.02,
	}
}

// handleResponse processes different response types for Call method
func (p *Provider) handleResponse(resp any) (string, error) {
	switch r := resp.(type) {
	case string:
		return r, nil
	case error:
		return "", r
	case *llms.ContentResponse:
		if len(r.Choices) > 0 {
			return r.Choices[0].Content, nil
		}
		return p.defaultResp, nil
	default:
		return fmt.Sprintf("%v", resp), nil
	}
}

// handleContentResponse processes responses for CallEx/CallWithTools
func (p *Provider) handleContentResponse(resp any) (*llms.ContentResponse, error) {
	switch r := resp.(type) {
	case error:
		return nil, r
	case *llms.ContentResponse:
		return r, nil
	case string:
		return &llms.ContentResponse{
			Choices: []*llms.ContentChoice{
				{
					Content: r,
				},
			},
		}, nil
	default:
		return &llms.ContentResponse{
			Choices: []*llms.ContentChoice{
				{
					Content: fmt.Sprintf("%v", resp),
				},
			},
		}, nil
	}
}

// handleStreamingResponse simulates streaming behavior
func (p *Provider) handleStreamingResponse(
	ctx context.Context,
	resp any,
	streamCb streaming.Callback,
) (*llms.ContentResponse, error) {
	contentResp, err := p.handleContentResponse(resp)
	if err != nil {
		return nil, err
	}

	if len(contentResp.Choices) == 0 {
		return contentResp, nil
	}

	choice := contentResp.Choices[0]

	// Simulate streaming by sending content in chunks
	content := choice.Content
	thinking := choice.Reasoning
	chunkSize := 5

	for i := 0; i < len(content); i += chunkSize {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}

		chunk := streaming.Chunk{
			Content: content[i:end],
		}

		// Add reasoning content to first chunk
		if i == 0 && !thinking.IsEmpty() {
			chunk.Reasoning = &reasoning.ContentReasoning{
				Content:   thinking.Content,
				Signature: thinking.Signature,
			}
		}

		if err := streamCb(ctx, chunk); err != nil {
			return nil, err
		}

		time.Sleep(p.streamingDelay)
	}

	return contentResp, nil
}
