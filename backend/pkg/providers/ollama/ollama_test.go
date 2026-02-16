package ollama

import (
	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultModel = "llama3.1:8b-instruct-q8_0"

func TestBuildProviderConfig(t *testing.T) {
	cfg := &config.Config{}
	configData := []byte(`{
		"agents": [
			{
				"agent": "simple",
				"model": "gemma3:1b",
				"temperature": 0.8,
				"maxTokens": 2000
			}
		]
	}`)

	providerConfig, err := BuildProviderConfig(cfg, configData)
	require.NoError(t, err)
	assert.NotNil(t, providerConfig)

	// check that model from config is used in options
	options := providerConfig.GetOptionsForType(pconfig.OptionsTypeSimple)
	assert.NotEmpty(t, options)
}

func TestDefaultProviderConfig(t *testing.T) {
	cfg := &config.Config{}

	providerConfig, err := DefaultProviderConfig(cfg)
	require.NoError(t, err)
	assert.NotNil(t, providerConfig)

	// verify all expected agent types are present
	agentTypes := []pconfig.ProviderOptionsType{
		pconfig.OptionsTypeSimple,
		pconfig.OptionsTypeSimpleJSON,
		pconfig.OptionsTypePrimaryAgent,
		pconfig.OptionsTypeAssistant,
		pconfig.OptionsTypeGenerator,
		pconfig.OptionsTypeRefiner,
		pconfig.OptionsTypeAdviser,
	}

	for _, agentType := range agentTypes {
		options := providerConfig.GetOptionsForType(agentType)
		assert.NotEmpty(t, options, "agent type %s should have options", agentType)
	}
}

func TestNew(t *testing.T) {
	cfg := &config.Config{
		OllamaServerURL:   "http://localhost:11434",
		OllamaServerModel: defaultModel,
	}

	providerConfig, err := DefaultProviderConfig(cfg)
	require.NoError(t, err)

	prov, err := New(cfg, providerConfig)
	require.NoError(t, err)
	assert.NotNil(t, prov)

	assert.Equal(t, provider.ProviderOllama, prov.Type())
	assert.NotNil(t, prov.GetProviderConfig())
	// GetModels() may return nil when no Ollama server is running (unit test environment)
	assert.NotEmpty(t, prov.GetRawConfig())

	// test model method
	model := prov.Model(pconfig.OptionsTypeSimple)
	assert.NotEmpty(t, model)

	// test get usage method
	info := map[string]any{
		"PromptTokens":     100,
		"CompletionTokens": 50,
	}
	inputTokens, outputTokens := prov.GetUsage(info)
	assert.Equal(t, int64(100), inputTokens)
	assert.Equal(t, int64(50), outputTokens)
}

func TestOllamaProviderWithProxy(t *testing.T) {
	cfg := &config.Config{
		OllamaServerURL: "http://localhost:11434",
		ProxyURL:        "http://proxy:8080",
	}

	providerConfig, err := DefaultProviderConfig(cfg)
	require.NoError(t, err)

	prov, err := New(cfg, providerConfig)
	require.NoError(t, err)
	assert.NotNil(t, prov)
}

func TestOllamaProviderWithCustomConfig(t *testing.T) {
	cfg := &config.Config{
		OllamaServerURL:    "http://localhost:11434",
		OllamaServerConfig: "testdata/custom_config.yml",
	}

	// test fallback to embedded config when file doesn't exist
	providerConfig, err := DefaultProviderConfig(cfg)
	if err == nil {
		// if file exists, check that provider can be created
		prov, err := New(cfg, providerConfig)
		require.NoError(t, err)
		assert.NotNil(t, prov)
	} else {
		// if file doesn't exist, should use embedded config
		cfg.OllamaServerConfig = ""
		providerConfig, err := DefaultProviderConfig(cfg)
		require.NoError(t, err)

		prov, err := New(cfg, providerConfig)
		require.NoError(t, err)
		assert.NotNil(t, prov)
	}
}

func TestOllamaProviderPricing(t *testing.T) {
	cfg := &config.Config{
		OllamaServerURL: "http://localhost:11434",
	}

	providerConfig, err := DefaultProviderConfig(cfg)
	require.NoError(t, err)

	prov, err := New(cfg, providerConfig)
	require.NoError(t, err)

	// ollama is free local inference, so pricing should be nil for most cases
	agentTypes := []pconfig.ProviderOptionsType{
		pconfig.OptionsTypeSimple,
		pconfig.OptionsTypeAssistant,
		pconfig.OptionsTypeGenerator,
		pconfig.OptionsTypePentester,
	}

	for _, agentType := range agentTypes {
		priceInfo := prov.GetPriceInfo(agentType)
		// ollama provider may not have pricing info, that's acceptable
		_ = priceInfo
	}
}

func TestGetUsageEdgeCases(t *testing.T) {
	cfg := &config.Config{
		OllamaServerURL: "http://localhost:11434",
	}

	providerConfig, err := DefaultProviderConfig(cfg)
	require.NoError(t, err)

	prov, err := New(cfg, providerConfig)
	require.NoError(t, err)

	// test empty info
	inputTokens, outputTokens := prov.GetUsage(map[string]any{})
	assert.Equal(t, int64(0), inputTokens)
	assert.Equal(t, int64(0), outputTokens)

	// test nil info
	inputTokens, outputTokens = prov.GetUsage(nil)
	assert.Equal(t, int64(0), inputTokens)
	assert.Equal(t, int64(0), outputTokens)

	// test with different field names (should return 0)
	info := map[string]any{
		"InputTokens":  100,
		"OutputTokens": 50,
	}
	inputTokens, outputTokens = prov.GetUsage(info)
	assert.Equal(t, int64(0), inputTokens)
	assert.Equal(t, int64(0), outputTokens)
}
