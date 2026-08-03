package deepseek

import (
	"embed"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/openaicompat"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/system"

	"github.com/vxcontrol/langchaingo/llms"
)

//go:embed config.yml models.yml
var configFS embed.FS

const DeepSeekAgentModel = "deepseek-v4-flash"

const DeepSeekToolCallIDTemplate = "call_{r:2:d}_{r:24:b}"

func BuildProviderConfig(configData []byte) (*pconfig.ProviderConfig, error) {
	defaultOptions := []llms.CallOption{
		llms.WithModel(DeepSeekAgentModel),
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

func New(
	cfg *config.Config,
	providerName provider.ProviderName,
	providerConfig *pconfig.ProviderConfig,
) (provider.Provider, error) {
	httpClient, err := system.GetHTTPClient(cfg)
	if err != nil {
		return nil, err
	}

	models, err := DefaultModels()
	if err != nil {
		return nil, err
	}

	return openaicompat.New(openaicompat.Spec{
		Type:               provider.ProviderDeepSeek,
		Model:              DeepSeekAgentModel,
		ToolCallIDTemplate: DeepSeekToolCallIDTemplate,
		APIKey:             cfg.DeepSeekAPIKey,
		ServerURL:          cfg.DeepSeekServerURL,
		Prefix:             cfg.DeepSeekProvider,
		PreserveReasoning:  true,
	}, httpClient, models, providerName, providerConfig)
}
