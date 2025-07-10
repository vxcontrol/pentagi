package mock

import (
	"context"
	"fmt"
	"strings"
	"time"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

// Provider implements provider.Provider for testing purposes
type Provider struct {
	providerType   provider.ProviderType
	modelName      string
	responses      map[string]interface{} // key -> response mapping
	defaultResp    string
	streamingDelay time.Duration
}

// ResponseConfig configures mock responses
type ResponseConfig struct {
	Key      string      // Request identifier (prompt/message content)
	Response interface{} // Response (string, *llms.ContentResponse, or error)
}

// NewProvider creates a new mock provider
func NewProvider(providerType provider.ProviderType, modelName string) *Provider {
	return &Provider{
		providerType:   providerType,
		modelName:      modelName,
		responses:      make(map[string]interface{}),
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

// SetStreamingDelay configures delay between streaming chunks
func (p *Provider) SetStreamingDelay(delay time.Duration) {
	p.streamingDelay = delay
}

// Type implements provider.Provider
func (p *Provider) Type() provider.ProviderType {
	return p.providerType
}

// Model implements provider.Provider
func (p *Provider) Model(opt pconfig.ProviderOptionsType) string {
	return p.modelName
}

// GetUsage implements provider.Provider
func (p *Provider) GetUsage(info map[string]any) (int64, int64) {
	return 100, 50 // Mock token counts
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
	var respInterface interface{}
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
	var respInterface interface{}
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

// GetRawConfig implements provider.Provider
func (p *Provider) GetRawConfig() []byte {
	return []byte(`{"mock": true}`)
}

// GetProviderConfig implements provider.Provider
func (p *Provider) GetProviderConfig() *pconfig.ProviderConfig {
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
func (p *Provider) handleResponse(resp interface{}) (string, error) {
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
func (p *Provider) handleContentResponse(resp interface{}) (*llms.ContentResponse, error) {
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
	resp interface{},
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
	reasoning := choice.ReasoningContent
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
		if i == 0 && reasoning != "" {
			chunk.ReasoningContent = reasoning
		}

		if err := streamCb(ctx, chunk); err != nil {
			return nil, err
		}

		time.Sleep(p.streamingDelay)
	}

	return contentResp, nil
}
