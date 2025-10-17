package pconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/vxcontrol/langchaingo/llms"
	"gopkg.in/yaml.v3"
)

type ProviderOptionsType string

const (
	OptionsTypePrimaryAgent ProviderOptionsType = "primary_agent"
	OptionsTypeAssistant    ProviderOptionsType = "assistant"
	OptionsTypeSimple       ProviderOptionsType = "simple"
	OptionsTypeSimpleJSON   ProviderOptionsType = "simple_json"
	OptionsTypeAdviser      ProviderOptionsType = "adviser"
	OptionsTypeGenerator    ProviderOptionsType = "generator"
	OptionsTypeRefiner      ProviderOptionsType = "refiner"
	OptionsTypeSearcher     ProviderOptionsType = "searcher"
	OptionsTypeEnricher     ProviderOptionsType = "enricher"
	OptionsTypeCoder        ProviderOptionsType = "coder"
	OptionsTypeInstaller    ProviderOptionsType = "installer"
	OptionsTypePentester    ProviderOptionsType = "pentester"
	OptionsTypeReflector    ProviderOptionsType = "reflector"
)

var AllAgentTypes = []ProviderOptionsType{
	OptionsTypeSimple,
	OptionsTypeSimpleJSON,
	OptionsTypePrimaryAgent,
	OptionsTypeAssistant,
	OptionsTypeGenerator,
	OptionsTypeRefiner,
	OptionsTypeAdviser,
	OptionsTypeReflector,
	OptionsTypeSearcher,
	OptionsTypeEnricher,
	OptionsTypeCoder,
	OptionsTypeInstaller,
	OptionsTypePentester,
}

type ModelConfig struct {
	Name        string     `json:"name,omitempty" yaml:"name,omitempty"`
	Description *string    `json:"description,omitempty" yaml:"description,omitempty"`
	ReleaseDate *time.Time `json:"release_date,omitempty" yaml:"release_date,omitempty"`
	Thinking    *bool      `json:"thinking,omitempty" yaml:"thinking,omitempty"`
	Price       *PriceInfo `json:"price,omitempty" yaml:"price,omitempty"`
}

type ModelsConfig []ModelConfig

type PriceInfo struct {
	Input  float64 `json:"input,omitempty" yaml:"input,omitempty"`
	Output float64 `json:"output,omitempty" yaml:"output,omitempty"`
}

type ReasoningConfig struct {
	Effort    llms.ReasoningEffort `json:"effort,omitempty" yaml:"effort,omitempty"`
	MaxTokens int                  `json:"max_tokens,omitempty" yaml:"max_tokens,omitempty"`
}

// AgentConfig represents the configuration for a single agent
type AgentConfig struct {
	Model             string          `json:"model,omitempty" yaml:"model,omitempty"`
	MaxTokens         int             `json:"max_tokens,omitempty" yaml:"max_tokens,omitempty"`
	Temperature       float64         `json:"temperature,omitempty" yaml:"temperature,omitempty"`
	TopK              int             `json:"top_k,omitempty" yaml:"top_k,omitempty"`
	TopP              float64         `json:"top_p,omitempty" yaml:"top_p,omitempty"`
	N                 int             `json:"n,omitempty" yaml:"n,omitempty"`
	MinLength         int             `json:"min_length,omitempty" yaml:"min_length,omitempty"`
	MaxLength         int             `json:"max_length,omitempty" yaml:"max_length,omitempty"`
	RepetitionPenalty float64         `json:"repetition_penalty,omitempty" yaml:"repetition_penalty,omitempty"`
	FrequencyPenalty  float64         `json:"frequency_penalty,omitempty" yaml:"frequency_penalty,omitempty"`
	PresencePenalty   float64         `json:"presence_penalty,omitempty" yaml:"presence_penalty,omitempty"`
	JSON              bool            `json:"json,omitempty" yaml:"json,omitempty"`
	ResponseMIMEType  string          `json:"response_mime_type,omitempty" yaml:"response_mime_type,omitempty"`
	Reasoning         ReasoningConfig `json:"reasoning,omitempty" yaml:"reasoning,omitempty"`
	Price             *PriceInfo      `json:"price,omitempty" yaml:"price,omitempty"`
	raw               map[string]any  `json:"-" yaml:"-"`
}

// ProviderConfig represents the configuration for all agents
type ProviderConfig struct {
	Simple         *AgentConfig      `json:"simple,omitempty" yaml:"simple,omitempty"`
	SimpleJSON     *AgentConfig      `json:"simple_json,omitempty" yaml:"simple_json,omitempty"`
	PrimaryAgent   *AgentConfig      `json:"primary_agent,omitempty" yaml:"primary_agent,omitempty"`
	Assistant      *AgentConfig      `json:"assistant,omitempty" yaml:"assistant,omitempty"`
	Generator      *AgentConfig      `json:"generator,omitempty" yaml:"generator,omitempty"`
	Refiner        *AgentConfig      `json:"refiner,omitempty" yaml:"refiner,omitempty"`
	Adviser        *AgentConfig      `json:"adviser,omitempty" yaml:"adviser,omitempty"`
	Reflector      *AgentConfig      `json:"reflector,omitempty" yaml:"reflector,omitempty"`
	Searcher       *AgentConfig      `json:"searcher,omitempty" yaml:"searcher,omitempty"`
	Enricher       *AgentConfig      `json:"enricher,omitempty" yaml:"enricher,omitempty"`
	Coder          *AgentConfig      `json:"coder,omitempty" yaml:"coder,omitempty"`
	Installer      *AgentConfig      `json:"installer,omitempty" yaml:"installer,omitempty"`
	Pentester      *AgentConfig      `json:"pentester,omitempty" yaml:"pentester,omitempty"`
	defaultOptions []llms.CallOption `json:"-" yaml:"-"`
	rawConfig      []byte            `json:"-" yaml:"-"`
}

const EmptyProviderConfigRaw = `{
  "simple": {},
  "simple_json": {},
  "primary_agent": {},
  "assistant": {},
  "generator": {},
  "refiner": {},
  "adviser": {},
  "reflector": {},
  "searcher": {},
  "enricher": {},
  "coder": {},
  "installer": {},
  "pentester": {}
}`

func LoadConfig(configPath string, defaultOptions []llms.CallOption) (*ProviderConfig, error) {
	if configPath == "" {
		return nil, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ProviderConfig
	ext := filepath.Ext(configPath)
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file extension: %s", ext)
	}

	// handle backward compatibility with legacy config format
	handleLegacyConfig(&config, data)

	config.defaultOptions = defaultOptions
	config.rawConfig = data

	return &config, nil
}

func LoadConfigData(configData []byte, defaultOptions []llms.CallOption) (*ProviderConfig, error) {
	var config ProviderConfig

	if err := yaml.Unmarshal(configData, &config); err != nil {
		if err := json.Unmarshal(configData, &config); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// handle backward compatibility with legacy config format
	handleLegacyConfig(&config, configData)

	config.defaultOptions = defaultOptions
	config.rawConfig = configData

	return &config, nil
}

func LoadModelsConfigData(configData []byte) (ModelsConfig, error) {
	var modelsConfig ModelsConfig

	if err := yaml.Unmarshal(configData, &modelsConfig); err != nil {
		return nil, fmt.Errorf("failed to parse models config: %w", err)
	}

	return modelsConfig, nil
}

// UnmarshalJSON implements custom JSON unmarshaling for ModelConfig
func (mc *ModelConfig) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Parse each field manually
	if name, ok := raw["name"].(string); ok {
		mc.Name = name
	}

	if desc, ok := raw["description"].(string); ok {
		mc.Description = &desc
	}

	if thinking, ok := raw["thinking"].(bool); ok {
		mc.Thinking = &thinking
	}

	if dateStr, ok := raw["release_date"].(string); ok && dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Errorf("invalid release_date format, expected YYYY-MM-DD: %w", err)
		}
		mc.ReleaseDate = &parsedDate
	}

	if priceData, ok := raw["price"]; ok && priceData != nil {
		priceBytes, err := json.Marshal(priceData)
		if err != nil {
			return err
		}
		var price PriceInfo
		if err := json.Unmarshal(priceBytes, &price); err != nil {
			return err
		}
		mc.Price = &price
	}

	return nil
}

// UnmarshalYAML implements custom YAML unmarshaling for ModelConfig
func (mc *ModelConfig) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]any
	if err := value.Decode(&raw); err != nil {
		return err
	}

	// Parse each field manually
	if name, ok := raw["name"].(string); ok {
		mc.Name = name
	}

	if desc, ok := raw["description"].(string); ok {
		mc.Description = &desc
	}

	if thinking, ok := raw["thinking"].(bool); ok {
		mc.Thinking = &thinking
	}

	// Handle release_date - YAML can parse it as string or time.Time
	if dateValue, ok := raw["release_date"]; ok && dateValue != nil {
		switch v := dateValue.(type) {
		case string:
			if v != "" {
				parsedDate, err := time.Parse("2006-01-02", v)
				if err != nil {
					return fmt.Errorf("invalid release_date format, expected YYYY-MM-DD: %w", err)
				}
				mc.ReleaseDate = &parsedDate
			}
		case time.Time:
			// YAML automatically parsed it as time.Time
			mc.ReleaseDate = &v
		}
	}

	if priceData, ok := raw["price"]; ok && priceData != nil {
		priceBytes, err := yaml.Marshal(priceData)
		if err != nil {
			return err
		}
		var price PriceInfo
		if err := yaml.Unmarshal(priceBytes, &price); err != nil {
			return err
		}
		mc.Price = &price
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling for ModelConfig
func (mc ModelConfig) MarshalJSON() ([]byte, error) {
	aux := map[string]any{}

	if mc.Name != "" {
		aux["name"] = mc.Name
	}
	if mc.Description != nil {
		aux["description"] = *mc.Description
	}
	if mc.Thinking != nil {
		aux["thinking"] = *mc.Thinking
	}
	if mc.ReleaseDate != nil {
		aux["release_date"] = mc.ReleaseDate.Format("2006-01-02")
	}
	if mc.Price != nil {
		aux["price"] = mc.Price
	}

	return json.Marshal(aux)
}

// MarshalYAML implements custom YAML marshaling for ModelConfig
func (mc ModelConfig) MarshalYAML() (any, error) {
	aux := map[string]any{}

	if mc.Name != "" {
		aux["name"] = mc.Name
	}
	if mc.Description != nil {
		aux["description"] = *mc.Description
	}
	if mc.Thinking != nil {
		aux["thinking"] = *mc.Thinking
	}
	if mc.ReleaseDate != nil {
		aux["release_date"] = mc.ReleaseDate.Format("2006-01-02")
	}
	if mc.Price != nil {
		aux["price"] = mc.Price
	}

	return aux, nil
}

// handleLegacyConfig provides backward compatibility for old config format
// where "agent" was used instead of "primary_agent"
func handleLegacyConfig(config *ProviderConfig, data []byte) {
	// only process if PrimaryAgent is not set
	if config.PrimaryAgent != nil {
		// still handle assistant backward compatibility
		if config.Assistant == nil {
			config.Assistant = config.PrimaryAgent
		}
		return
	}

	// define legacy config structure with old "agent" field
	type LegacyProviderConfig struct {
		Agent *AgentConfig `json:"agent,omitempty" yaml:"agent,omitempty"`
	}

	var legacyConfig LegacyProviderConfig

	if err := yaml.Unmarshal(data, &legacyConfig); err != nil {
		if err := json.Unmarshal(data, &legacyConfig); err != nil {
			return
		}
	}

	if legacyConfig.Agent != nil {
		config.PrimaryAgent = legacyConfig.Agent
	}

	if config.Assistant == nil {
		config.Assistant = config.PrimaryAgent
	}
}

func (ac *AgentConfig) UnmarshalJSON(data []byte) error {
	type embed AgentConfig
	var unmarshaler embed
	if err := json.Unmarshal(data, &unmarshaler); err != nil {
		return err
	}
	*ac = AgentConfig(unmarshaler)

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	ac.raw = raw
	return nil
}

func (ac *AgentConfig) UnmarshalYAML(value *yaml.Node) error {
	type embed AgentConfig
	var unmarshaler embed
	if err := value.Decode(&unmarshaler); err != nil {
		return err
	}
	*ac = AgentConfig(unmarshaler)

	var raw map[string]any
	if err := value.Decode(&raw); err != nil {
		return err
	}
	ac.raw = raw
	return nil
}

func (ac *AgentConfig) BuildOptions() []llms.CallOption {
	if ac == nil || ac.raw == nil {
		return nil
	}

	var options []llms.CallOption

	if _, ok := ac.raw["model"]; ok && ac.Model != "" {
		options = append(options, llms.WithModel(ac.Model))
	}
	if _, ok := ac.raw["max_tokens"]; ok {
		options = append(options, llms.WithMaxTokens(ac.MaxTokens))
	}
	if _, ok := ac.raw["temperature"]; ok {
		options = append(options, llms.WithTemperature(ac.Temperature))
	}
	if _, ok := ac.raw["top_k"]; ok {
		options = append(options, llms.WithTopK(ac.TopK))
	}
	if _, ok := ac.raw["top_p"]; ok {
		options = append(options, llms.WithTopP(ac.TopP))
	}
	if _, ok := ac.raw["n"]; ok {
		options = append(options, llms.WithN(ac.N))
	}
	if _, ok := ac.raw["min_length"]; ok {
		options = append(options, llms.WithMinLength(ac.MinLength))
	}
	if _, ok := ac.raw["max_length"]; ok {
		options = append(options, llms.WithMaxLength(ac.MaxLength))
	}
	if _, ok := ac.raw["repetition_penalty"]; ok {
		options = append(options, llms.WithRepetitionPenalty(ac.RepetitionPenalty))
	}
	if _, ok := ac.raw["frequency_penalty"]; ok {
		options = append(options, llms.WithFrequencyPenalty(ac.FrequencyPenalty))
	}
	if _, ok := ac.raw["presence_penalty"]; ok {
		options = append(options, llms.WithPresencePenalty(ac.PresencePenalty))
	}
	if _, ok := ac.raw["json"]; ok {
		options = append(options, llms.WithJSONMode())
	}
	if _, ok := ac.raw["response_mime_type"]; ok && ac.ResponseMIMEType != "" {
		options = append(options, llms.WithResponseMIMEType(ac.ResponseMIMEType))
	}
	if _, ok := ac.raw["reasoning"]; ok && (ac.Reasoning.Effort != llms.ReasoningNone || ac.Reasoning.MaxTokens != 0) {
		switch ac.Reasoning.Effort {
		case llms.ReasoningLow, llms.ReasoningMedium, llms.ReasoningHigh:
			options = append(options, llms.WithReasoning(ac.Reasoning.Effort, 0))
		default:
			if ac.Reasoning.MaxTokens > 0 && ac.Reasoning.MaxTokens <= 32000 {
				options = append(options, llms.WithReasoning(llms.ReasoningNone, ac.Reasoning.MaxTokens))
			}
		}
	}

	return options
}

func (ac *AgentConfig) marshalMap() map[string]any {
	if ac == nil {
		return nil
	}

	// use raw map if available, otherwise create a new one
	if ac.raw != nil {
		return ac.raw
	}

	// add non-zero values
	output := make(map[string]any)
	if ac.Model != "" {
		output["model"] = ac.Model
	}
	if ac.MaxTokens != 0 {
		output["max_tokens"] = ac.MaxTokens
	}
	if ac.Temperature != 0 {
		output["temperature"] = ac.Temperature
	}
	if ac.TopK != 0 {
		output["top_k"] = ac.TopK
	}
	if ac.TopP != 0 {
		output["top_p"] = ac.TopP
	}
	if ac.N != 0 {
		output["n"] = ac.N
	}
	if ac.MinLength != 0 {
		output["min_length"] = ac.MinLength
	}
	if ac.MaxLength != 0 {
		output["max_length"] = ac.MaxLength
	}
	if ac.RepetitionPenalty != 0 {
		output["repetition_penalty"] = ac.RepetitionPenalty
	}
	if ac.FrequencyPenalty != 0 {
		output["frequency_penalty"] = ac.FrequencyPenalty
	}
	if ac.PresencePenalty != 0 {
		output["presence_penalty"] = ac.PresencePenalty
	}
	if ac.JSON {
		output["json"] = ac.JSON
	}
	if ac.ResponseMIMEType != "" {
		output["response_mime_type"] = ac.ResponseMIMEType
	}
	if ac.Reasoning.Effort != llms.ReasoningNone || ac.Reasoning.MaxTokens != 0 {
		output["reasoning"] = ac.Reasoning
	}
	if ac.Price != nil {
		output["price"] = ac.Price
	}

	return output
}

func (ac *AgentConfig) MarshalJSON() ([]byte, error) {
	if ac == nil {
		return []byte("null"), nil
	}
	return json.Marshal(ac.marshalMap())
}

func (ac *AgentConfig) MarshalYAML() (any, error) {
	if ac == nil {
		return nil, nil
	}
	return ac.marshalMap(), nil
}

func (pc *ProviderConfig) SetDefaultOptions(defaultOptions []llms.CallOption) {
	if pc == nil {
		return
	}
	pc.defaultOptions = defaultOptions
}

func (pc *ProviderConfig) GetDefaultOptions() []llms.CallOption {
	if pc == nil {
		return nil
	}
	return pc.defaultOptions
}

func (pc *ProviderConfig) SetRawConfig(rawConfig []byte) {
	if pc == nil {
		return
	}
	pc.rawConfig = rawConfig
}

func (pc *ProviderConfig) GetRawConfig() []byte {
	if pc == nil {
		return nil
	}
	return pc.rawConfig
}

func (pc *ProviderConfig) GetOptionsForType(optType ProviderOptionsType) []llms.CallOption {
	if pc == nil {
		return nil
	}

	var agentConfig *AgentConfig
	switch optType {
	case OptionsTypeSimple:
		agentConfig = pc.Simple
	case OptionsTypeSimpleJSON:
		return pc.buildSimpleJSONOptions()
	case OptionsTypePrimaryAgent:
		agentConfig = pc.PrimaryAgent
	case OptionsTypeAssistant:
		return pc.buildAssistantOptions()
	case OptionsTypeGenerator:
		agentConfig = pc.Generator
	case OptionsTypeRefiner:
		agentConfig = pc.Refiner
	case OptionsTypeAdviser:
		agentConfig = pc.Adviser
	case OptionsTypeReflector:
		agentConfig = pc.Reflector
	case OptionsTypeSearcher:
		agentConfig = pc.Searcher
	case OptionsTypeEnricher:
		agentConfig = pc.Enricher
	case OptionsTypeCoder:
		agentConfig = pc.Coder
	case OptionsTypeInstaller:
		agentConfig = pc.Installer
	case OptionsTypePentester:
		agentConfig = pc.Pentester
	default:
		return nil
	}

	if agentConfig != nil {
		if options := agentConfig.BuildOptions(); options != nil {
			return options
		}
	}

	return pc.defaultOptions
}

func (pc *ProviderConfig) GetPriceInfoForType(optType ProviderOptionsType) *PriceInfo {
	if pc == nil {
		return nil
	}

	var agentConfig *AgentConfig
	switch optType {
	case OptionsTypeSimple:
		agentConfig = pc.Simple
	case OptionsTypeSimpleJSON:
		if pc.SimpleJSON != nil {
			agentConfig = pc.SimpleJSON
		} else {
			agentConfig = pc.Simple
		}
	case OptionsTypePrimaryAgent:
		agentConfig = pc.PrimaryAgent
	case OptionsTypeAssistant:
		if pc.Assistant != nil {
			agentConfig = pc.Assistant
		} else {
			agentConfig = pc.PrimaryAgent
		}
	case OptionsTypeGenerator:
		agentConfig = pc.Generator
	case OptionsTypeRefiner:
		agentConfig = pc.Refiner
	case OptionsTypeAdviser:
		agentConfig = pc.Adviser
	case OptionsTypeReflector:
		agentConfig = pc.Reflector
	case OptionsTypeSearcher:
		agentConfig = pc.Searcher
	case OptionsTypeEnricher:
		agentConfig = pc.Enricher
	case OptionsTypeCoder:
		agentConfig = pc.Coder
	case OptionsTypeInstaller:
		agentConfig = pc.Installer
	case OptionsTypePentester:
		agentConfig = pc.Pentester
	default:
		return nil
	}

	if agentConfig != nil && agentConfig.Price != nil {
		return agentConfig.Price
	}

	return nil
}

func (pc *ProviderConfig) BuildOptionsMap() map[ProviderOptionsType][]llms.CallOption {
	if pc == nil {
		return nil
	}

	options := map[ProviderOptionsType][]llms.CallOption{
		OptionsTypeSimple:       pc.GetOptionsForType(OptionsTypeSimple),
		OptionsTypeSimpleJSON:   pc.GetOptionsForType(OptionsTypeSimpleJSON),
		OptionsTypePrimaryAgent: pc.GetOptionsForType(OptionsTypePrimaryAgent),
		OptionsTypeAssistant:    pc.GetOptionsForType(OptionsTypeAssistant),
		OptionsTypeGenerator:    pc.GetOptionsForType(OptionsTypeGenerator),
		OptionsTypeRefiner:      pc.GetOptionsForType(OptionsTypeRefiner),
		OptionsTypeAdviser:      pc.GetOptionsForType(OptionsTypeAdviser),
		OptionsTypeReflector:    pc.GetOptionsForType(OptionsTypeReflector),
		OptionsTypeSearcher:     pc.GetOptionsForType(OptionsTypeSearcher),
		OptionsTypeEnricher:     pc.GetOptionsForType(OptionsTypeEnricher),
		OptionsTypeCoder:        pc.GetOptionsForType(OptionsTypeCoder),
		OptionsTypeInstaller:    pc.GetOptionsForType(OptionsTypeInstaller),
		OptionsTypePentester:    pc.GetOptionsForType(OptionsTypePentester),
	}

	return options
}

func (pc *ProviderConfig) buildSimpleJSONOptions() []llms.CallOption {
	if pc == nil {
		return nil
	}

	if pc.SimpleJSON != nil {
		options := pc.SimpleJSON.BuildOptions()
		if options != nil {
			return options
		}
	}

	if pc.Simple != nil {
		options := pc.Simple.BuildOptions()
		if options != nil {
			return append(options, llms.WithJSONMode())
		}
	}

	if pc.defaultOptions != nil {
		return append(pc.defaultOptions, llms.WithJSONMode())
	}

	return nil
}

func (pc *ProviderConfig) buildAssistantOptions() []llms.CallOption {
	if pc == nil {
		return nil
	}

	if pc.Assistant != nil {
		options := pc.Assistant.BuildOptions()
		if options != nil {
			return options
		}
	}

	if pc.PrimaryAgent != nil {
		options := pc.PrimaryAgent.BuildOptions()
		if options != nil {
			return options
		}
	}

	return pc.defaultOptions
}
