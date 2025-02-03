package custom

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"pentagi/pkg/providers/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"gopkg.in/yaml.v3"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configFile  string
		content     string
		wantErr     bool
		checkConfig func(*testing.T, *ProvidersConfig)
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
			checkConfig: func(t *testing.T, cfg *ProvidersConfig) {
				require.NotNil(t, cfg.Simple)
				assert.Equal(t, "test-model", cfg.Simple.Model)
				assert.Equal(t, 100, cfg.Simple.MaxTokens)
				assert.Equal(t, 0.7, cfg.Simple.Temperature)
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
			checkConfig: func(t *testing.T, cfg *ProvidersConfig) {
				require.NotNil(t, cfg.Simple)
				assert.Equal(t, "test-model", cfg.Simple.Model)
				assert.Equal(t, 100, cfg.Simple.MaxTokens)
				assert.Equal(t, 0.7, cfg.Simple.Temperature)
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
	config := &ProvidersConfig{
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
		config  *ProvidersConfig
		optType provider.ProviderOptionsType
		wantLen int
	}{
		{
			name:    "nil config",
			config:  nil,
			optType: provider.OptionsTypeSimple,
			wantLen: 0,
		},
		{
			name:    "existing config",
			config:  config,
			optType: provider.OptionsTypeSimple,
			wantLen: 3,
		},
		{
			name:    "non-existing config",
			config:  config,
			optType: provider.OptionsTypeAgent,
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
		checkConfig    func(*testing.T, *ProvidersConfig)
	}{
		{
			name:           "empty path with defaults",
			configFile:     "",
			defaultOptions: defaultOptions,
			checkConfig: func(t *testing.T, cfg *ProvidersConfig) {
				// when configPath is empty, we should create a new config with defaults
				cfg = &ProvidersConfig{defaultOptions: defaultOptions}
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
			checkConfig: func(t *testing.T, cfg *ProvidersConfig) {
				require.NotNil(t, cfg)
				options := cfg.GetOptionsForType(provider.OptionsTypeSimple)
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
			checkConfig: func(t *testing.T, cfg *ProvidersConfig) {
				require.NotNil(t, cfg)
				options := cfg.GetOptionsForType(provider.OptionsTypeSimple)
				assert.Len(t, options, 1) // should use agent config, not defaults
				options = cfg.GetOptionsForType(provider.OptionsTypeAgent)
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
				cfg = &ProvidersConfig{defaultOptions: tt.defaultOptions}
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

	config := &ProvidersConfig{
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
		config      *ProvidersConfig
		optType     provider.ProviderOptionsType
		wantLen     int
		useDefaults bool
	}{
		{
			name:        "nil config",
			config:      nil,
			optType:     provider.OptionsTypeSimple,
			wantLen:     0,
			useDefaults: false,
		},
		{
			name:        "existing config",
			config:      config,
			optType:     provider.OptionsTypeSimple,
			wantLen:     3,
			useDefaults: false,
		},
		{
			name:        "non-existing config with defaults",
			config:      config,
			optType:     provider.OptionsTypeAgent,
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
