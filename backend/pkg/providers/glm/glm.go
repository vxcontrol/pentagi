package glm

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

const GLMAgentModel = "glm-4.7-flashx"

const GLMToolCallIDTemplate = "call_-{r:19:d}"

func BuildProviderConfig(configData []byte) (*pconfig.ProviderConfig, error) {
	defaultOptions := []llms.CallOption{
		llms.WithModel(GLMAgentModel),
		llms.WithN(1),
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
		Type:               provider.ProviderGLM,
		Model:              GLMAgentModel,
		ToolCallIDTemplate: GLMToolCallIDTemplate,
		APIKey:             cfg.GLMAPIKey,
		ServerURL:          cfg.GLMServerURL,
		Prefix:             cfg.GLMProvider,
		PreserveReasoning:  true,
	}, httpClient, models, providerName, providerConfig)
}
