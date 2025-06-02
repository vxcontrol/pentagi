package custom

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
	"gopkg.in/yaml.v3"
)

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
	raw               map[string]any  `json:"-" yaml:"-"`
}

// ProvidersConfig represents the configuration for all agents
type ProvidersConfig struct {
	Simple         *AgentConfig      `json:"simple,omitempty" yaml:"simple,omitempty"`
	SimpleJSON     *AgentConfig      `json:"simple_json,omitempty" yaml:"simple_json,omitempty"`
	Agent          *AgentConfig      `json:"agent,omitempty" yaml:"agent,omitempty"`
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
}

func LoadConfig(configPath string, defaultOptions []llms.CallOption) (*ProvidersConfig, error) {
	if configPath == "" {
		return nil, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ProvidersConfig
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
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	// use agent config for assistant if not set (for backward compatibility)
	if config.Assistant == nil {
		config.Assistant = config.Agent
	}

	config.defaultOptions = defaultOptions

	return &config, nil
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

	return output
}

func (ac *AgentConfig) MarshalJSON() ([]byte, error) {
	if ac == nil {
		return []byte("null"), nil
	}
	return json.Marshal(ac.marshalMap())
}

func (ac *AgentConfig) MarshalYAML() (interface{}, error) {
	if ac == nil {
		return nil, nil
	}
	return ac.marshalMap(), nil
}

func (pc *ProvidersConfig) GetOptionsForType(optType provider.ProviderOptionsType) []llms.CallOption {
	if pc == nil {
		return nil
	}

	var agentConfig *AgentConfig
	switch optType {
	case provider.OptionsTypeSimple:
		agentConfig = pc.Simple
	case provider.OptionsTypeSimpleJSON:
		agentConfig = pc.SimpleJSON
	case provider.OptionsTypeAgent:
		agentConfig = pc.Agent
	case provider.OptionsTypeAssistant:
		agentConfig = pc.Assistant
	case provider.OptionsTypeGenerator:
		agentConfig = pc.Generator
	case provider.OptionsTypeRefiner:
		agentConfig = pc.Refiner
	case provider.OptionsTypeAdviser:
		agentConfig = pc.Adviser
	case provider.OptionsTypeReflector:
		agentConfig = pc.Reflector
	case provider.OptionsTypeSearcher:
		agentConfig = pc.Searcher
	case provider.OptionsTypeEnricher:
		agentConfig = pc.Enricher
	case provider.OptionsTypeCoder:
		agentConfig = pc.Coder
	case provider.OptionsTypeInstaller:
		agentConfig = pc.Installer
	case provider.OptionsTypePentester:
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
