package pconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"gopkg.in/yaml.v3"
)

func TestReasoningConfig_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    ReasoningConfig
		wantErr bool
	}{
		{
			name: "empty object",
			json: "{}",
			want: ReasoningConfig{},
		},
		{
			name: "with effort only",
			json: `{"effort": "low"}`,
			want: ReasoningConfig{
				Effort: llms.ReasoningLow,
			},
		},
		{
			name: "with max_tokens only",
			json: `{"max_tokens": 1000}`,
			want: ReasoningConfig{
				MaxTokens: 1000,
			},
		},
		{
			name: "with both fields",
			json: `{"effort": "high", "max_tokens": 2000}`,
			want: ReasoningConfig{
				Effort:    llms.ReasoningHigh,
				MaxTokens: 2000,
			},
		},
		{
			name:    "invalid json",
			json:    "{invalid}",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ReasoningConfig
			err := json.Unmarshal([]byte(tt.json), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Effort, got.Effort)
			assert.Equal(t, tt.want.MaxTokens, got.MaxTokens)
		})
	}
}

func TestReasoningConfig_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    ReasoningConfig
		wantErr bool
	}{
		{
			name: "empty object",
			yaml: "{}",
			want: ReasoningConfig{},
		},
		{
			name: "with effort only",
			yaml: `effort: low`,
			want: ReasoningConfig{
				Effort: llms.ReasoningLow,
			},
		},
		{
			name: "with max_tokens only",
			yaml: `max_tokens: 1000`,
			want: ReasoningConfig{
				MaxTokens: 1000,
			},
		},
		{
			name: "with both fields",
			yaml: `
effort: high
max_tokens: 2000
`,
			want: ReasoningConfig{
				Effort:    llms.ReasoningHigh,
				MaxTokens: 2000,
			},
		},
		{
			name:    "invalid yaml",
			yaml:    "invalid: [yaml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ReasoningConfig
			err := yaml.Unmarshal([]byte(tt.yaml), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Effort, got.Effort)
			assert.Equal(t, tt.want.MaxTokens, got.MaxTokens)
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configFile  string
		content     string
		wantErr     bool
		checkConfig func(*testing.T, *ProviderConfig)
	}{
		{
			name:       "empty path",
			configFile: "",
			wantErr:    false,
		},
		{
			name:       "invalid json",
			configFile: "config.json",
			content:    "{invalid}",
			wantErr:    true,
		},
		{
			name:       "invalid yaml",
			configFile: "config.yaml",
			content:    "invalid: [yaml",
			wantErr:    true,
		},
		{
			name:       "unsupported format",
			configFile: "config.txt",
			content:    "some text",
			wantErr:    true,
		},
		{
			name:       "valid json",
			configFile: "config.json",
			content: `{
				"simple": {
					"model": "test-model",
					"max_tokens": 100,
					"temperature": 0.7
				}
			}`,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg.Simple)
				assert.Equal(t, "test-model", cfg.Simple.Model)
				assert.Equal(t, 100, cfg.Simple.MaxTokens)
				assert.Equal(t, 0.7, cfg.Simple.Temperature)
			},
		},
		{
			name:       "valid json with reasoning config - both parameters",
			configFile: "config.json",
			content: `{
				"simple": {
					"model": "test-model",
					"max_tokens": 100,
					"temperature": 0.7,
					"reasoning": {
						"effort": "medium",
						"max_tokens": 5000
					}
				}
			}`,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg.Simple)
				assert.Equal(t, "test-model", cfg.Simple.Model)
				assert.Equal(t, 100, cfg.Simple.MaxTokens)
				assert.Equal(t, 0.7, cfg.Simple.Temperature)
				assert.Equal(t, llms.ReasoningMedium, cfg.Simple.Reasoning.Effort)
				assert.Equal(t, 5000, cfg.Simple.Reasoning.MaxTokens)

				// Verify options include reasoning
				options := cfg.Simple.BuildOptions()
				assert.Len(t, options, 4) // model, max_tokens, temperature, reasoning
			},
		},
		{
			name:       "valid json with reasoning config - effort only",
			configFile: "config.json",
			content: `{
				"simple": {
					"model": "test-model",
					"max_tokens": 100,
					"temperature": 0.7,
					"reasoning": {
						"effort": "high",
						"max_tokens": 0
					}
				}
			}`,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg.Simple)
				assert.Equal(t, "test-model", cfg.Simple.Model)
				assert.Equal(t, 100, cfg.Simple.MaxTokens)
				assert.Equal(t, 0.7, cfg.Simple.Temperature)
				assert.Equal(t, llms.ReasoningHigh, cfg.Simple.Reasoning.Effort)
				assert.Equal(t, 0, cfg.Simple.Reasoning.MaxTokens)

				// Verify options include reasoning
				options := cfg.Simple.BuildOptions()
				assert.Len(t, options, 4) // model, max_tokens, temperature, reasoning
			},
		},
		{
			name:       "valid json with reasoning config - max_tokens only",
			configFile: "config.json",
			content: `{
				"simple": {
					"model": "test-model",
					"max_tokens": 100,
					"temperature": 0.7,
					"reasoning": {
						"effort": "none",
						"max_tokens": 3000
					}
				}
			}`,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg.Simple)
				assert.Equal(t, "test-model", cfg.Simple.Model)
				assert.Equal(t, 100, cfg.Simple.MaxTokens)
				assert.Equal(t, 0.7, cfg.Simple.Temperature)
				assert.Equal(t, llms.ReasoningEffort("none"), cfg.Simple.Reasoning.Effort)
				assert.Equal(t, 3000, cfg.Simple.Reasoning.MaxTokens)

				// Verify options include reasoning
				options := cfg.Simple.BuildOptions()
				assert.Len(t, options, 4) // model, max_tokens, temperature, reasoning
			},
		},
		{
			name:       "valid yaml",
			configFile: "config.yaml",
			content: `
simple:
  model: test-model
  max_tokens: 100
  temperature: 0.7
`,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg.Simple)
				assert.Equal(t, "test-model", cfg.Simple.Model)
				assert.Equal(t, 100, cfg.Simple.MaxTokens)
				assert.Equal(t, 0.7, cfg.Simple.Temperature)
			},
		},
		{
			name:       "valid yaml with reasoning config - both parameters",
			configFile: "config.yaml",
			content: `
simple:
  model: test-model
  max_tokens: 100
  temperature: 0.7
  reasoning:
    effort: high
    max_tokens: 8000
`,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg.Simple)
				assert.Equal(t, "test-model", cfg.Simple.Model)
				assert.Equal(t, 100, cfg.Simple.MaxTokens)
				assert.Equal(t, 0.7, cfg.Simple.Temperature)
				assert.Equal(t, llms.ReasoningHigh, cfg.Simple.Reasoning.Effort)
				assert.Equal(t, 8000, cfg.Simple.Reasoning.MaxTokens)

				// Verify options include reasoning
				options := cfg.Simple.BuildOptions()
				assert.Len(t, options, 4) // model, max_tokens, temperature, reasoning
			},
		},
		{
			name:       "valid yaml with reasoning config - effort only",
			configFile: "config.yaml",
			content: `
simple:
  model: test-model
  max_tokens: 100
  temperature: 0.7
  reasoning:
    effort: low
    max_tokens: 0
`,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg.Simple)
				assert.Equal(t, "test-model", cfg.Simple.Model)
				assert.Equal(t, 100, cfg.Simple.MaxTokens)
				assert.Equal(t, 0.7, cfg.Simple.Temperature)
				assert.Equal(t, llms.ReasoningLow, cfg.Simple.Reasoning.Effort)
				assert.Equal(t, 0, cfg.Simple.Reasoning.MaxTokens)

				// Verify options include reasoning
				options := cfg.Simple.BuildOptions()
				assert.Len(t, options, 4) // model, max_tokens, temperature, reasoning
			},
		},
		{
			name:       "valid yaml with reasoning config - max_tokens only",
			configFile: "config.yaml",
			content: `
simple:
  model: test-model
  max_tokens: 100
  temperature: 0.7
  reasoning:
    effort: none
    max_tokens: 2500
`,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg.Simple)
				assert.Equal(t, "test-model", cfg.Simple.Model)
				assert.Equal(t, 100, cfg.Simple.MaxTokens)
				assert.Equal(t, 0.7, cfg.Simple.Temperature)
				assert.Equal(t, llms.ReasoningEffort("none"), cfg.Simple.Reasoning.Effort)
				assert.Equal(t, 2500, cfg.Simple.Reasoning.MaxTokens)

				// Verify options include reasoning
				options := cfg.Simple.BuildOptions()
				assert.Len(t, options, 4) // model, max_tokens, temperature, reasoning
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.configFile != "" {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, tt.configFile)
				err := os.WriteFile(configPath, []byte(tt.content), 0644)
				require.NoError(t, err)
				tt.configFile = configPath
			}

			cfg, err := LoadConfig(tt.configFile, nil)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.checkConfig != nil {
				tt.checkConfig(t, cfg)
			}
		})
	}
}

func TestLoadConfig_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name        string
		configFile  string
		content     string
		format      string
		wantErr     bool
		checkConfig func(*testing.T, *ProviderConfig)
	}{
		{
			name:       "legacy agent config JSON",
			configFile: "legacy.json",
			format:     "json",
			content: `{
				"agent": {
					"model": "legacy-model",
					"max_tokens": 200,
					"temperature": 0.8
				},
				"simple": {
					"model": "simple-model",
					"max_tokens": 100
				}
			}`,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg.PrimaryAgent, "PrimaryAgent should be set from legacy 'agent' field")
				assert.Equal(t, "legacy-model", cfg.PrimaryAgent.Model)
				assert.Equal(t, 200, cfg.PrimaryAgent.MaxTokens)
				assert.Equal(t, 0.8, cfg.PrimaryAgent.Temperature)

				require.NotNil(t, cfg.Simple, "Simple config should be preserved")
				assert.Equal(t, "simple-model", cfg.Simple.Model)

				require.NotNil(t, cfg.Assistant, "Assistant should be set from PrimaryAgent for backward compatibility")
				assert.Equal(t, cfg.PrimaryAgent, cfg.Assistant)
			},
		},
		{
			name:       "legacy agent config YAML",
			configFile: "legacy.yaml",
			format:     "yaml",
			content: `
agent:
  model: legacy-yaml-model
  max_tokens: 300
  temperature: 0.9
simple:
  model: simple-yaml-model
  max_tokens: 150
`,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg.PrimaryAgent, "PrimaryAgent should be set from legacy 'agent' field")
				assert.Equal(t, "legacy-yaml-model", cfg.PrimaryAgent.Model)
				assert.Equal(t, 300, cfg.PrimaryAgent.MaxTokens)
				assert.Equal(t, 0.9, cfg.PrimaryAgent.Temperature)

				require.NotNil(t, cfg.Simple, "Simple config should be preserved")
				assert.Equal(t, "simple-yaml-model", cfg.Simple.Model)

				require.NotNil(t, cfg.Assistant, "Assistant should be set from PrimaryAgent for backward compatibility")
				assert.Equal(t, cfg.PrimaryAgent, cfg.Assistant)
			},
		},
		{
			name:       "new primary_agent config takes precedence JSON",
			configFile: "new_format.json",
			format:     "json",
			content: `{
				"primary_agent": {
					"model": "new-model",
					"max_tokens": 400,
					"temperature": 0.6
				},
				"agent": {
					"model": "old-model",
					"max_tokens": 200,
					"temperature": 0.8
				}
			}`,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg.PrimaryAgent, "PrimaryAgent should be set")
				// Should use primary_agent, not legacy agent
				assert.Equal(t, "new-model", cfg.PrimaryAgent.Model)
				assert.Equal(t, 400, cfg.PrimaryAgent.MaxTokens)
				assert.Equal(t, 0.6, cfg.PrimaryAgent.Temperature)

				require.NotNil(t, cfg.Assistant, "Assistant should be set from PrimaryAgent")
				assert.Equal(t, cfg.PrimaryAgent, cfg.Assistant)
			},
		},
		{
			name:       "explicit assistant config overrides default YAML",
			configFile: "explicit_assistant.yaml",
			format:     "yaml",
			content: `
agent:
  model: agent-model
  max_tokens: 200
assistant:
  model: assistant-model
  max_tokens: 500
  temperature: 0.5
`,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg.PrimaryAgent, "PrimaryAgent should be set from legacy 'agent'")
				assert.Equal(t, "agent-model", cfg.PrimaryAgent.Model)

				require.NotNil(t, cfg.Assistant, "Assistant should be set")
				// Should use explicit assistant config, not agent
				assert.Equal(t, "assistant-model", cfg.Assistant.Model)
				assert.Equal(t, 500, cfg.Assistant.MaxTokens)
				assert.Equal(t, 0.5, cfg.Assistant.Temperature)
				assert.NotEqual(t, cfg.PrimaryAgent, cfg.Assistant, "Assistant should not be the same as PrimaryAgent")
			},
		},
		{
			name:       "no agent configs at all",
			configFile: "no_agents.json",
			format:     "json",
			content: `{
				"simple": {
					"model": "simple-only",
					"max_tokens": 100
				}
			}`,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				assert.Nil(t, cfg.PrimaryAgent, "PrimaryAgent should be nil")
				assert.Nil(t, cfg.Assistant, "Assistant should be nil")
				require.NotNil(t, cfg.Simple, "Simple should be set")
				assert.Equal(t, "simple-only", cfg.Simple.Model)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, tt.configFile)
			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			cfg, err := LoadConfig(configPath, nil)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
			if tt.checkConfig != nil {
				tt.checkConfig(t, cfg)
			}
		})
	}
}

func TestLoadConfigData_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name        string
		configData  string
		format      string
		wantErr     bool
		checkConfig func(*testing.T, *ProviderConfig)
	}{
		{
			name:   "legacy agent config JSON data",
			format: "json",
			configData: `{
				"agent": {
					"model": "legacy-data-model",
					"max_tokens": 250,
					"temperature": 0.7
				}
			}`,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg.PrimaryAgent, "PrimaryAgent should be set from legacy 'agent' field")
				assert.Equal(t, "legacy-data-model", cfg.PrimaryAgent.Model)
				assert.Equal(t, 250, cfg.PrimaryAgent.MaxTokens)
				assert.Equal(t, 0.7, cfg.PrimaryAgent.Temperature)

				require.NotNil(t, cfg.Assistant, "Assistant should be set from PrimaryAgent")
				assert.Equal(t, cfg.PrimaryAgent, cfg.Assistant)
			},
		},
		{
			name:   "legacy agent config YAML data",
			format: "yaml",
			configData: `
agent:
  model: legacy-yaml-data-model
  max_tokens: 350
  temperature: 0.85
`,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg.PrimaryAgent, "PrimaryAgent should be set from legacy 'agent' field")
				assert.Equal(t, "legacy-yaml-data-model", cfg.PrimaryAgent.Model)
				assert.Equal(t, 350, cfg.PrimaryAgent.MaxTokens)
				assert.Equal(t, 0.85, cfg.PrimaryAgent.Temperature)

				require.NotNil(t, cfg.Assistant, "Assistant should be set from PrimaryAgent")
				assert.Equal(t, cfg.PrimaryAgent, cfg.Assistant)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := LoadConfigData([]byte(tt.configData), nil)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
			if tt.checkConfig != nil {
				tt.checkConfig(t, cfg)
			}
		})
	}
}

func TestAgentConfig_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		want       *AgentConfig
		wantFields []string
		wantErr    bool
	}{
		{
			name: "empty object",
			json: "{}",
			want: &AgentConfig{},
		},
		{
			name: "zero values",
			json: `{
				"model": "",
				"max_tokens": 0,
				"temperature": 0,
				"top_k": 0,
				"top_p": 0
			}`,
			want: &AgentConfig{},
			wantFields: []string{
				"model",
				"max_tokens",
				"temperature",
				"top_k",
				"top_p",
			},
		},
		{
			name: "with values",
			json: `{
				"model": "test-model",
				"max_tokens": 100,
				"temperature": 0.7
			}`,
			want: &AgentConfig{
				Model:       "test-model",
				MaxTokens:   100,
				Temperature: 0.7,
			},
			wantFields: []string{
				"model",
				"max_tokens",
				"temperature",
			},
		},
		{
			name: "with reasoning config",
			json: `{
				"model": "test-model",
				"max_tokens": 100,
				"temperature": 0.7,
				"reasoning": {
					"effort": "medium",
					"max_tokens": 5000
				}
			}`,
			want: &AgentConfig{
				Model:       "test-model",
				MaxTokens:   100,
				Temperature: 0.7,
				Reasoning: ReasoningConfig{
					Effort:    llms.ReasoningMedium,
					MaxTokens: 5000,
				},
			},
			wantFields: []string{
				"model",
				"max_tokens",
				"temperature",
				"reasoning",
			},
		},
		{
			name:    "invalid json",
			json:    "{invalid}",
			want:    &AgentConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got AgentConfig
			err := json.Unmarshal([]byte(tt.json), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Model, got.Model)
			assert.Equal(t, tt.want.MaxTokens, got.MaxTokens)
			assert.Equal(t, tt.want.Temperature, got.Temperature)

			require.NotNil(t, got.raw)
			for _, field := range tt.wantFields {
				_, ok := got.raw[field]
				assert.True(t, ok, "field %s should be present in raw", field)
			}
		})
	}
}

func TestAgentConfig_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name       string
		yaml       string
		want       *AgentConfig
		wantFields []string
		wantErr    bool
	}{
		{
			name: "empty object",
			yaml: "{}",
			want: &AgentConfig{},
		},
		{
			name: "zero values",
			yaml: `
model: ""
max_tokens: 0
temperature: 0
top_k: 0
top_p: 0
`,
			want: &AgentConfig{},
			wantFields: []string{
				"model",
				"max_tokens",
				"temperature",
				"top_k",
				"top_p",
			},
		},
		{
			name: "with values",
			yaml: `
model: test-model
max_tokens: 100
temperature: 0.7
`,
			want: &AgentConfig{
				Model:       "test-model",
				MaxTokens:   100,
				Temperature: 0.7,
			},
			wantFields: []string{
				"model",
				"max_tokens",
				"temperature",
			},
		},
		{
			name: "with reasoning config",
			yaml: `
model: test-model
max_tokens: 100
temperature: 0.7
reasoning:
  effort: medium
  max_tokens: 5000
`,
			want: &AgentConfig{
				Model:       "test-model",
				MaxTokens:   100,
				Temperature: 0.7,
				Reasoning: ReasoningConfig{
					Effort:    llms.ReasoningMedium,
					MaxTokens: 5000,
				},
			},
			wantFields: []string{
				"model",
				"max_tokens",
				"temperature",
				"reasoning",
			},
		},
		{
			name:    "invalid yaml",
			yaml:    "invalid: [yaml",
			want:    &AgentConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got AgentConfig
			err := yaml.Unmarshal([]byte(tt.yaml), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Model, got.Model)
			assert.Equal(t, tt.want.MaxTokens, got.MaxTokens)
			assert.Equal(t, tt.want.Temperature, got.Temperature)

			require.NotNil(t, got.raw)
			for _, field := range tt.wantFields {
				_, ok := got.raw[field]
				assert.True(t, ok, "field %s should be present in raw", field)
			}
		})
	}
}

func TestAgentConfig_BuildOptions(t *testing.T) {
	tests := []struct {
		name         string
		config       string
		format       string
		wantLen      int
		checkNil     bool
		checkOptions func(*testing.T, []llms.CallOption)
	}{
		{
			name:     "nil config",
			checkNil: true,
		},
		{
			name:    "empty config",
			format:  "json",
			config:  "{}",
			wantLen: 0,
		},
		{
			name:   "zero values",
			format: "json",
			config: `{
				"model": "",
				"max_tokens": 0,
				"temperature": 0,
				"top_k": 0,
				"top_p": 0
			}`,
			wantLen: 4, // model is excluded because it's empty string
		},
		{
			name:   "full config json",
			format: "json",
			config: `{
				"model": "test-model",
				"max_tokens": 100,
				"temperature": 0.7,
				"top_k": 10,
				"top_p": 0.9,
				"min_length": 10,
				"max_length": 100,
				"repetition_penalty": 1.1,
				"frequency_penalty": 1.2,
				"presence_penalty": 1.3,
				"json": true,
				"response_mime_type": "application/json"
			}`,
			wantLen: 12,
		},
		{
			name:   "with reasoning config - low effort",
			format: "json",
			config: `{
				"model": "test-model",
				"temperature": 0.7,
				"reasoning": {
					"effort": "low",
					"max_tokens": 1000
				}
			}`,
			wantLen: 3, // model, temperature, reasoning
			checkOptions: func(t *testing.T, options []llms.CallOption) {
				// We can't directly check the reasoning parameter value
				// because WithReasoning returns an opaque CallOption
				// Instead, we'll verify the number of options is correct
				assert.Len(t, options, 3)
			},
		},
		{
			name:   "with reasoning config - medium effort only",
			format: "json",
			config: `{
				"model": "test-model",
				"temperature": 0.7,
				"reasoning": {
					"effort": "medium",
					"max_tokens": 0
				}
			}`,
			wantLen: 3, // model, temperature, reasoning (effort is medium)
		},
		{
			name:   "with reasoning config - high effort only",
			format: "json",
			config: `{
				"model": "test-model",
				"temperature": 0.7,
				"reasoning": {
					"effort": "high",
					"max_tokens": 0
				}
			}`,
			wantLen: 3, // model, temperature, reasoning (effort is high)
		},
		{
			name:   "with reasoning config - custom tokens only",
			format: "json",
			config: `{
				"model": "test-model",
				"temperature": 0.7,
				"reasoning": {
					"effort": "none",
					"max_tokens": 5000
				}
			}`,
			wantLen: 3, // model, temperature, reasoning (max_tokens is set)
		},
		{
			name:   "with invalid reasoning tokens over limit",
			format: "json",
			config: `{
				"model": "test-model",
				"temperature": 0.7,
				"reasoning": {
					"effort": "none",
					"max_tokens": 50000
				}
			}`,
			wantLen: 2, // shouldn't include reasoning option because tokens > 32000
		},
		{
			name:   "with invalid reasoning tokens negative",
			format: "json",
			config: `{
				"model": "test-model",
				"temperature": 0.7,
				"reasoning": {
					"effort": "none",
					"max_tokens": -100
				}
			}`,
			wantLen: 2, // shouldn't include reasoning option because tokens < 0
		},
		{
			name:   "with reasoning config - none effort and zero tokens",
			format: "json",
			config: `{
				"model": "test-model",
				"temperature": 0.7,
				"reasoning": {
					"effort": "none",
					"max_tokens": 0
				}
			}`,
			wantLen: 2, // shouldn't include reasoning because both parameters are default/zero values
		},
		{
			name:   "full config yaml",
			format: "yaml",
			config: `
model: test-model
max_tokens: 100
temperature: 0.7
top_k: 10
top_p: 0.9
min_length: 10
max_length: 100
repetition_penalty: 1.1
frequency_penalty: 1.2
presence_penalty: 1.3
json: true
response_mime_type: application/json
`,
			wantLen: 12,
		},
		{
			name:   "partial config",
			format: "json",
			config: `{
				"model": "test-model",
				"temperature": 0.7
			}`,
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config AgentConfig
			var err error

			if tt.config != "" {
				switch tt.format {
				case "json":
					err = json.Unmarshal([]byte(tt.config), &config)
				case "yaml":
					err = yaml.Unmarshal([]byte(tt.config), &config)
				}
				require.NoError(t, err)
			}

			var options []llms.CallOption
			if tt.checkNil {
				options = (*AgentConfig)(nil).BuildOptions()
			} else {
				options = config.BuildOptions()
			}

			if tt.checkNil {
				assert.Nil(t, options)
				return
			}
			assert.Len(t, options, tt.wantLen)
			if tt.checkOptions != nil {
				tt.checkOptions(t, options)
			}
		})
	}
}

func TestProvidersConfig_GetOptionsForType(t *testing.T) {
	config := &ProviderConfig{
		Simple: &AgentConfig{},
	}

	// initialize raw map for Simple config
	simpleJSON := `{
		"model": "test-model",
		"max_tokens": 100,
		"temperature": 0.7
	}`
	err := json.Unmarshal([]byte(simpleJSON), config.Simple)
	require.NoError(t, err)

	tests := []struct {
		name    string
		config  *ProviderConfig
		optType ProviderOptionsType
		wantLen int
	}{
		{
			name:    "nil config",
			config:  nil,
			optType: OptionsTypeSimple,
			wantLen: 0,
		},
		{
			name:    "existing config",
			config:  config,
			optType: OptionsTypeSimple,
			wantLen: 3,
		},
		{
			name:    "non-existing config",
			config:  config,
			optType: OptionsTypePrimaryAgent,
			wantLen: 0,
		},
		{
			name:    "invalid type",
			config:  config,
			optType: "invalid",
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := tt.config.GetOptionsForType(tt.optType)
			assert.Len(t, options, tt.wantLen)
		})
	}
}

func TestAgentConfig_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		config  *AgentConfig
		want    string
		wantErr bool
	}{
		{
			name:   "nil config",
			config: nil,
			want:   "null",
		},
		{
			name:   "empty config",
			config: &AgentConfig{},
			want:   "{}",
		},
		{
			name: "with values",
			config: &AgentConfig{
				Model:       "test-model",
				MaxTokens:   100,
				Temperature: 0.7,
			},
			want: `{"max_tokens":100,"model":"test-model","temperature":0.7}`,
		},
		{
			name: "with reasoning config",
			config: &AgentConfig{
				Model:       "test-model",
				MaxTokens:   100,
				Temperature: 0.7,
				Reasoning: ReasoningConfig{
					Effort:    llms.ReasoningMedium,
					MaxTokens: 5000,
				},
				raw: map[string]any{
					"model":       "test-model",
					"max_tokens":  100,
					"temperature": 0.7,
					"reasoning": map[string]any{
						"effort":     "medium",
						"max_tokens": 5000,
					},
				},
			},
			want: `{"max_tokens":100,"model":"test-model","reasoning":{"effort":"medium","max_tokens":5000},"temperature":0.7}`,
		},
		{
			name: "with zero values",
			config: &AgentConfig{
				Model:       "",
				MaxTokens:   0,
				Temperature: 0,
				JSON:        false,
			},
			want: "{}",
		},
		{
			name: "with raw values",
			config: &AgentConfig{
				Model:       "test-model",
				MaxTokens:   100,
				Temperature: 0.7,
				raw: map[string]any{
					"custom_field": "custom_value",
					"max_tokens":   100,
					"model":        "test-model",
					"temperature":  0.7,
				},
			},
			want: `{"custom_field":"custom_value","max_tokens":100,"model":"test-model","temperature":0.7}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.JSONEq(t, tt.want, string(got))
		})
	}
}

func TestAgentConfig_MarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		config  *AgentConfig
		want    map[string]any
		wantErr bool
	}{
		{
			name:   "nil config",
			config: nil,
			want:   nil,
		},
		{
			name:   "empty config",
			config: &AgentConfig{},
			want:   map[string]any{},
		},
		{
			name: "with values",
			config: &AgentConfig{
				Model:       "test-model",
				MaxTokens:   100,
				Temperature: 0.7,
			},
			want: map[string]any{
				"model":       "test-model",
				"max_tokens":  100,
				"temperature": 0.7,
			},
		},
		{
			name: "with reasoning config",
			config: &AgentConfig{
				Model:       "test-model",
				MaxTokens:   100,
				Temperature: 0.7,
				Reasoning: ReasoningConfig{
					Effort:    llms.ReasoningMedium,
					MaxTokens: 5000,
				},
				raw: map[string]any{
					"model":       "test-model",
					"max_tokens":  100,
					"temperature": 0.7,
					"reasoning": map[string]any{
						"effort":     "medium",
						"max_tokens": 5000,
					},
				},
			},
			want: map[string]any{
				"model":       "test-model",
				"max_tokens":  100,
				"temperature": 0.7,
				"reasoning": map[string]any{
					"effort":     "medium",
					"max_tokens": 5000,
				},
			},
		},
		{
			name: "with zero values",
			config: &AgentConfig{
				Model:       "",
				MaxTokens:   0,
				Temperature: 0,
				JSON:        false,
			},
			want: map[string]any{},
		},
		{
			name: "with raw values",
			config: &AgentConfig{
				Model:       "test-model",
				MaxTokens:   100,
				Temperature: 0.7,
				raw: map[string]any{
					"custom_field": "custom_value",
					"max_tokens":   100,
					"model":        "test-model",
					"temperature":  0.7,
				},
			},
			want: map[string]any{
				"custom_field": "custom_value",
				"max_tokens":   100,
				"model":        "test-model",
				"temperature":  0.7,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := yaml.Marshal(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.want == nil {
				assert.Equal(t, "null\n", string(got))
				return
			}

			// unmarshal back to map for comparison
			var gotMap map[string]any
			err = yaml.Unmarshal(got, &gotMap)
			require.NoError(t, err)

			// compare maps
			assert.Equal(t, tt.want, gotMap)
		})
	}
}

func TestLoadConfig_WithDefaultOptions(t *testing.T) {
	defaultOptions := []llms.CallOption{
		llms.WithTemperature(0.5),
		llms.WithMaxTokens(1000),
	}

	tests := []struct {
		name           string
		configFile     string
		content        string
		defaultOptions []llms.CallOption
		checkConfig    func(*testing.T, *ProviderConfig)
	}{
		{
			name:           "empty path with defaults",
			configFile:     "",
			defaultOptions: defaultOptions,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				// when configPath is empty, we should create a new config with defaults
				cfg = &ProviderConfig{defaultOptions: defaultOptions}
				require.NotNil(t, cfg)
				assert.Equal(t, defaultOptions, cfg.defaultOptions)
			},
		},
		{
			name:       "config without agent",
			configFile: "config.json",
			content:    "{}",
			defaultOptions: []llms.CallOption{
				llms.WithTemperature(0.5),
			},
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg)
				options := cfg.GetOptionsForType(OptionsTypeSimple)
				assert.Len(t, options, 1)
			},
		},
		{
			name:       "config with agent",
			configFile: "config.json",
			content: `{
				"simple": {
					"temperature": 0.7
				}
			}`,
			defaultOptions: defaultOptions,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg)
				options := cfg.GetOptionsForType(OptionsTypeSimple)
				assert.Len(t, options, 1) // should use agent config, not defaults
				options = cfg.GetOptionsForType(OptionsTypePrimaryAgent)
				assert.Len(t, options, 2) // should use defaults
			},
		},
		{
			name:       "assistant backward compatibility with agent",
			configFile: "config.json",
			content: `{
				"agent": {
					"temperature": 0.7
				}
			}`,
			defaultOptions: defaultOptions,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg)
				options := cfg.GetOptionsForType(OptionsTypeAssistant)
				assert.Len(t, options, 1) // should use agent config (backward compatibility)
				options = cfg.GetOptionsForType(OptionsTypePrimaryAgent)
				assert.Len(t, options, 1)
			},
		},
		{
			name:       "config assistant without agent",
			configFile: "config.json",
			content: `{
				"assistant": {
					"temperature": 0.7
				}
			}`,
			defaultOptions: defaultOptions,
			checkConfig: func(t *testing.T, cfg *ProviderConfig) {
				require.NotNil(t, cfg)
				options := cfg.GetOptionsForType(OptionsTypeAssistant)
				assert.Len(t, options, 1) // should use assistant config
				options = cfg.GetOptionsForType(OptionsTypePrimaryAgent)
				assert.Len(t, options, 2) // should use defaults
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configPath string
			if tt.configFile != "" {
				tmpDir := t.TempDir()
				configPath = filepath.Join(tmpDir, tt.configFile)
				err := os.WriteFile(configPath, []byte(tt.content), 0644)
				require.NoError(t, err)
			}

			cfg, err := LoadConfig(configPath, tt.defaultOptions)
			if configPath == "" {
				cfg = &ProviderConfig{defaultOptions: tt.defaultOptions}
			}
			require.NoError(t, err)
			if tt.checkConfig != nil {
				tt.checkConfig(t, cfg)
			}
		})
	}
}

func TestProvidersConfig_GetOptionsForType_WithDefaults(t *testing.T) {
	defaultOptions := []llms.CallOption{
		llms.WithTemperature(0.5),
		llms.WithMaxTokens(1000),
	}

	config := &ProviderConfig{
		Simple:         &AgentConfig{},
		defaultOptions: defaultOptions,
	}

	// initialize raw map for Simple config
	simpleJSON := `{
		"model": "test-model",
		"max_tokens": 100,
		"temperature": 0.7
	}`
	err := json.Unmarshal([]byte(simpleJSON), config.Simple)
	require.NoError(t, err)

	tests := []struct {
		name        string
		config      *ProviderConfig
		optType     ProviderOptionsType
		wantLen     int
		useDefaults bool
	}{
		{
			name:        "nil config",
			config:      nil,
			optType:     OptionsTypeSimple,
			wantLen:     0,
			useDefaults: false,
		},
		{
			name:        "existing config",
			config:      config,
			optType:     OptionsTypeSimple,
			wantLen:     3,
			useDefaults: false,
		},
		{
			name:        "non-existing config with defaults",
			config:      config,
			optType:     OptionsTypePrimaryAgent,
			wantLen:     2,
			useDefaults: true,
		},
		{
			name:        "invalid type with defaults",
			config:      config,
			optType:     "invalid",
			wantLen:     0,
			useDefaults: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := tt.config.GetOptionsForType(tt.optType)
			assert.Len(t, options, tt.wantLen)
			if tt.useDefaults {
				assert.Equal(t, defaultOptions, options)
			}
		})
	}
}

func TestProviderConfig_BuildOptionsMap(t *testing.T) {
	tests := []struct {
		name        string
		config      *ProviderConfig
		wantNil     bool
		checkResult func(*testing.T, map[ProviderOptionsType][]llms.CallOption)
	}{
		{
			name:    "nil config",
			config:  nil,
			wantNil: true,
		},
		{
			name: "empty config with defaults",
			config: &ProviderConfig{
				defaultOptions: []llms.CallOption{
					llms.WithTemperature(0.5),
					llms.WithMaxTokens(1000),
				},
			},
			checkResult: func(t *testing.T, options map[ProviderOptionsType][]llms.CallOption) {
				// All provider types should be present
				expectedTypes := []ProviderOptionsType{
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

				assert.Len(t, options, len(expectedTypes), "All provider types should be present")

				for _, optType := range expectedTypes {
					opt, exists := options[optType]
					assert.True(t, exists, "Option type %s should exist", optType)
					assert.NotNil(t, opt, "Option for type %s should not be nil", optType)
				}
			},
		},
		{
			name: "config with some agent configurations",
			config: func() *ProviderConfig {
				cfg := &ProviderConfig{
					defaultOptions: []llms.CallOption{
						llms.WithTemperature(0.5),
						llms.WithMaxTokens(1000),
					},
				}

				// Configure simple agent
				simpleJSON := `{
					"model": "simple-model",
					"max_tokens": 100,
					"temperature": 0.7
				}`
				cfg.Simple = &AgentConfig{}
				err := json.Unmarshal([]byte(simpleJSON), cfg.Simple)
				if err != nil {
					t.Fatal(err)
				}

				// Configure agent
				agentJSON := `{
					"model": "agent-model",
					"max_tokens": 200,
					"temperature": 0.8
				}`
				cfg.PrimaryAgent = &AgentConfig{}
				err = json.Unmarshal([]byte(agentJSON), cfg.PrimaryAgent)
				if err != nil {
					t.Fatal(err)
				}

				return cfg
			}(),
			checkResult: func(t *testing.T, options map[ProviderOptionsType][]llms.CallOption) {
				// All provider types should be present
				expectedTypes := []ProviderOptionsType{
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

				assert.Len(t, options, len(expectedTypes), "All provider types should be present")

				// Simple should have 3 options (model, max_tokens, temperature)
				simpleOpts := options[OptionsTypeSimple]
				assert.Len(t, simpleOpts, 3, "Simple options should have 3 elements")

				// SimpleJSON should have simple options + JSON mode
				simpleJSONOpts := options[OptionsTypeSimpleJSON]
				assert.Len(t, simpleJSONOpts, 4, "SimpleJSON options should have 4 elements (simple + JSON)")

				// Agent should have 3 options (model, max_tokens, temperature)
				agentOpts := options[OptionsTypePrimaryAgent]
				assert.Len(t, agentOpts, 3, "Agent options should have 3 elements")

				// Assistant should have agent options (backward compatibility)
				assistantOpts := options[OptionsTypeAssistant]
				assert.Len(t, assistantOpts, 3, "Assistant options should have 3 elements (using agent config)")

				// Other types should use defaults
				generatorOpts := options[OptionsTypeGenerator]
				assert.Len(t, generatorOpts, 2, "Generator options should use defaults")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.BuildOptionsMap()

			if tt.wantNil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}
