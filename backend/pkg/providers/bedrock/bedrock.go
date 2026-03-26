package bedrock

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sync"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/templates"

	bconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	smithybearer "github.com/aws/smithy-go/auth/bearer"
	"github.com/invopop/jsonschema"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/bedrock"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

//go:embed config.yml models.yml
var configFS embed.FS

const BedrockAgentModel = bedrock.ModelAnthropicClaudeSonnet4

const BedrockToolCallIDTemplate = "tooluse_{r:22:x}"

func BuildProviderConfig(configData []byte) (*pconfig.ProviderConfig, error) {
	defaultOptions := []llms.CallOption{
		llms.WithModel(BedrockAgentModel),
		llms.WithTemperature(1.0),
		llms.WithN(1),
		llms.WithMaxTokens(4000),
	}

	providerConfig, err := pconfig.LoadConfigData(configData, defaultOptions)
	if err != nil {
		return nil, err
	}

	return providerConfig, nil
}

func DefaultProviderConfig() (*pconfig.ProviderConfig, error) {
	configData, err := configFS.ReadFile("config.yml")
	if err != nil {
		return nil, err
	}

	return BuildProviderConfig(configData)
}

func DefaultModels() (pconfig.ModelsConfig, error) {
	configData, err := configFS.ReadFile("models.yml")
	if err != nil {
		return nil, err
	}

	return pconfig.LoadModelsConfigData(configData)
}

type bedrockProvider struct {
	llm            *bedrock.LLM
	models         pconfig.ModelsConfig
	providerConfig *pconfig.ProviderConfig

	toolCallIDTemplate     string
	toolCallIDTemplateOnce sync.Once
	toolCallIDTemplateErr  error
}

func New(cfg *config.Config, providerConfig *pconfig.ProviderConfig) (provider.Provider, error) {
	opts := []func(*bconfig.LoadOptions) error{
		bconfig.WithRegion(cfg.BedrockRegion),
	}

	// Choose authentication strategy based on configuration
	if cfg.BedrockDefaultAuth {
		// Use default AWS SDK credential chain (environment, EC2 role, etc.)
		// Don't add any explicit credentials provider
	} else if cfg.BedrockBearerToken != "" {
		// Use bearer token authentication
		opts = append(opts, bconfig.WithBearerAuthTokenProvider(smithybearer.StaticTokenProvider{
			Token: smithybearer.Token{
				Value: cfg.BedrockBearerToken,
			},
		}))
	} else if cfg.BedrockAccessKey != "" && cfg.BedrockSecretKey != "" {
		// Use static credentials (traditional approach)
		opts = append(opts, bconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.BedrockAccessKey,
			cfg.BedrockSecretKey,
			cfg.BedrockSessionToken,
		)))
	} else {
		return nil, fmt.Errorf("no valid authentication method configured for Bedrock")
	}

	if cfg.BedrockServerURL != "" {
		opts = append(opts, bconfig.WithBaseEndpoint(cfg.BedrockServerURL))
	}

	if cfg.ProxyURL != "" {
		opts = append(opts, bconfig.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(cfg.ProxyURL)
				},
			},
		}))
	}

	bcfg, err := bconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}

	bclient := bedrockruntime.NewFromConfig(bcfg)

	models, err := DefaultModels()
	if err != nil {
		return nil, err
	}

	client, err := bedrock.New(
		bedrock.WithClient(bclient),
		bedrock.WithModel(BedrockAgentModel),
		bedrock.WithConverseAPI(),
	)
	if err != nil {
		return nil, err
	}

	return &bedrockProvider{
		llm:            client,
		models:         models,
		providerConfig: providerConfig,
	}, nil
}

func (p *bedrockProvider) Type() provider.ProviderType {
	return provider.ProviderBedrock
}

func (p *bedrockProvider) GetRawConfig() []byte {
	return p.providerConfig.GetRawConfig()
}

func (p *bedrockProvider) GetProviderConfig() *pconfig.ProviderConfig {
	return p.providerConfig
}

func (p *bedrockProvider) GetPriceInfo(opt pconfig.ProviderOptionsType) *pconfig.PriceInfo {
	return p.providerConfig.GetPriceInfoForType(opt)
}

func (p *bedrockProvider) GetModels() pconfig.ModelsConfig {
	return p.models
}

func (p *bedrockProvider) Model(opt pconfig.ProviderOptionsType) string {
	model := BedrockAgentModel
	opts := llms.CallOptions{Model: &model}
	for _, option := range p.providerConfig.GetOptionsForType(opt) {
		option(&opts)
	}

	return opts.GetModel()
}

func (p *bedrockProvider) ModelWithPrefix(opt pconfig.ProviderOptionsType) string {
	// Bedrock provider doesn't need prefix support (passthrough mode in LiteLLM)
	return p.Model(opt)
}

func (p *bedrockProvider) Call(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	prompt string,
) (string, error) {
	return provider.WrapGenerateFromSinglePrompt(
		ctx, p, opt, p.llm, prompt,
		p.providerConfig.GetOptionsForType(opt)...,
	)
}

func (p *bedrockProvider) CallEx(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	chain []llms.MessageContent,
	streamCb streaming.Callback,
) (*llms.ContentResponse, error) {
	// The AWS Bedrock Converse API requires toolConfig to be defined whenever the
	// conversation history contains toolUse or toolResult content blocks — even when
	// no new tools are being offered in the current turn.  Without it the API returns:
	//   ValidationException: The toolConfig field must be defined when using
	//   toolUse and toolResult content blocks.
	// We reconstruct minimal tool definitions from the tool-call names already
	// present in the chain so that the library sets toolConfig automatically.
	configOptions := p.providerConfig.GetOptionsForType(opt)

	// Extract and restore tools
	tools := extractToolsFromOptions(configOptions)
	tools = restoreMissedToolsFromChain(chain, tools)

	// Clean tools from $schema field
	tools = cleanToolSchemas(tools)

	// Build final options: streaming + config + cleaned tools LAST (to override any dirty tools from config)
	options := []llms.CallOption{llms.WithStreamingFunc(streamCb)}
	options = append(options, configOptions...)
	options = append(options, llms.WithTools(tools))

	return provider.WrapGenerateContent(ctx, p, opt, p.llm.GenerateContent, chain, options...)
}

func (p *bedrockProvider) CallWithTools(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	chain []llms.MessageContent,
	tools []llms.Tool,
	streamCb streaming.Callback,
) (*llms.ContentResponse, error) {
	// Same Bedrock Converse API requirement as in CallEx: if no tools were
	// explicitly provided for this turn but the chain already carries toolUse /
	// toolResult blocks, reconstruct minimal definitions so that the library
	// includes toolConfig in the request.
	tools = restoreMissedToolsFromChain(chain, tools)

	// Clean tools from $schema field
	tools = cleanToolSchemas(tools)

	configOptions := p.providerConfig.GetOptionsForType(opt)

	// Build final options: config + streaming + cleaned tools LAST (to override any dirty tools from config)
	options := append(configOptions, llms.WithStreamingFunc(streamCb), llms.WithTools(tools))

	return provider.WrapGenerateContent(ctx, p, opt, p.llm.GenerateContent, chain, options...)
}

func (p *bedrockProvider) GetUsage(info map[string]any) pconfig.CallUsage {
	return pconfig.NewCallUsage(info)
}

func (p *bedrockProvider) GetToolCallIDTemplate(ctx context.Context, prompter templates.Prompter) (string, error) {
	return provider.DetermineToolCallIDTemplate(ctx, p, pconfig.OptionsTypeSimple, prompter, BedrockToolCallIDTemplate)
}

func extractToolsFromOptions(options []llms.CallOption) []llms.Tool {
	var opts llms.CallOptions

	for _, option := range options {
		option(&opts)
	}

	return opts.Tools
}

func restoreMissedToolsFromChain(chain []llms.MessageContent, tools []llms.Tool) []llms.Tool {
	// Build index of already declared tools to avoid overwriting them
	declaredTools := make(map[string]llms.Tool)
	for _, tool := range tools {
		if tool.Function != nil && tool.Function.Name != "" {
			declaredTools[tool.Function.Name] = tool
		}
	}

	// Collect tool usage from chain with their arguments for schema inference
	toolUsage := collectToolUsageFromChain(chain)
	if len(toolUsage) == 0 {
		return tools
	}

	// Build enhanced tool definitions only for tools not already declared
	result := make([]llms.Tool, len(tools))
	copy(result, tools)

	for name, args := range toolUsage {
		if _, exists := declaredTools[name]; exists {
			// Trust the existing declaration - don't update it
			continue
		}

		// Infer schema from arguments found in the chain
		schema := inferSchemaFromArguments(args)
		result = append(result, llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        name,
				Description: fmt.Sprintf("Tool: %s", name),
				Parameters:  schema,
			},
		})
	}

	return result
}

// collectToolUsageFromChain scans the message chain and collects all unique
// tool names along with sample arguments from their invocations. This allows
// us to infer parameter schemas using reflection on actual usage.
func collectToolUsageFromChain(chain []llms.MessageContent) map[string][]string {
	usage := make(map[string][]string)

	for _, msg := range chain {
		for _, part := range msg.Parts {
			switch p := part.(type) {
			case llms.ToolCall:
				if p.FunctionCall != nil && p.FunctionCall.Name != "" {
					usage[p.FunctionCall.Name] = append(usage[p.FunctionCall.Name], p.FunctionCall.Arguments)
				}
			case llms.ToolCallResponse:
				if p.Name != "" {
					// ToolCallResponse doesn't have arguments, but we record the tool name
					if _, exists := usage[p.Name]; !exists {
						usage[p.Name] = []string{}
					}
				}
			}
		}
	}

	return usage
}

// inferSchemaFromArguments attempts to infer a JSON schema for a tool's parameters
// by analyzing actual argument samples from the chain. It uses reflection to determine
// top-level property types (simple types, arrays, or objects) without descending deeper.
func inferSchemaFromArguments(argumentSamples []string) map[string]any {
	schema := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}

	if len(argumentSamples) == 0 {
		return schema
	}

	// Aggregate properties from all samples to build a complete schema
	properties := make(map[string]any)

	for _, argJSON := range argumentSamples {
		if argJSON == "" {
			continue
		}

		var args map[string]any
		if err := json.Unmarshal([]byte(argJSON), &args); err != nil {
			// Invalid JSON - skip this sample
			continue
		}

		for key, value := range args {
			if _, exists := properties[key]; exists {
				// Already inferred from a previous sample - trust first occurrence
				continue
			}

			propType := inferPropertyType(value)
			properties[key] = map[string]any{"type": propType}
		}
	}

	schema["properties"] = properties
	return schema
}

// inferPropertyType determines the JSON schema type for a property value.
// It only classifies top-level types: string, number, boolean, array, object, or null.
func inferPropertyType(value any) string {
	if value == nil {
		return "null"
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	default:
		return "object"
	}
}

// cleanToolSchemas removes the $schema field from tool parameters to ensure
// compatibility with AWS Bedrock Converse API, which rejects schemas containing
// the $schema metadata field (returns ValidationException).
func cleanToolSchemas(tools []llms.Tool) []llms.Tool {
	if len(tools) == 0 {
		return tools
	}

	cleaned := make([]llms.Tool, len(tools))
	for i, tool := range tools {
		cleaned[i] = tool
		if tool.Function != nil && tool.Function.Parameters != nil {
			cleanedParams := cleanParameters(tool.Function.Parameters)
			if cleanedParams != nil {
				cleanedFunc := *tool.Function
				cleanedFunc.Parameters = cleanedParams
				cleaned[i].Function = &cleanedFunc
			}
		}
	}

	return cleaned
}

// cleanParameters removes $schema field from parameters of any type
func cleanParameters(params any) any {
	// Case 1: *jsonschema.Schema - convert to map[string]any without $schema
	if schema, ok := params.(*jsonschema.Schema); ok {
		data, err := schema.MarshalJSON()
		if err != nil {
			return params
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			return params
		}

		delete(result, "$schema")
		return result
	}

	// Case 2: map[string]any - just remove $schema
	if paramsMap, ok := params.(map[string]any); ok {
		cleanedParams := make(map[string]any, len(paramsMap))
		for key, value := range paramsMap {
			if key != "$schema" {
				cleanedParams[key] = value
			}
		}
		return cleanedParams
	}

	// Case 3: other types - return as is
	return params
}
