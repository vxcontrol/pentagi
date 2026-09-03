package aimlapi

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

// AIMLAPIAgentModel is the fallback model used when an agent role resolves to no
// model of its own. Model ids are namespaced by upstream vendor and must be sent
// verbatim; the gateway also accepts short aliases, but an alias can resolve to a
// different model than its name suggests, so the canonical id is used everywhere.
const AIMLAPIAgentModel = "deepseek/deepseek-v4-flash"

// AIMLAPIToolCallIDTemplate is deliberately empty. AI/ML API multiplexes many
// upstream vendors behind one endpoint, so the tool-call id format follows the
// model actually serving the request rather than the gateway. Leaving it empty
// makes the shared detector derive the template per model instead of asserting a
// single format for the whole provider.
const AIMLAPIToolCallIDTemplate = ""

func BuildProviderConfig(configData []byte) (*pconfig.ProviderConfig, error) {
	defaultOptions := []llms.CallOption{
		llms.WithModel(AIMLAPIAgentModel),
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
		Type:               provider.ProviderAIMLAPI,
		Model:              AIMLAPIAgentModel,
		ToolCallIDTemplate: AIMLAPIToolCallIDTemplate,
		APIKey:             cfg.AIMLAPIKey,
		ServerURL:          cfg.AIMLAPIServerURL,
		Prefix:             cfg.AIMLAPIProvider,
		PreserveReasoning:  true,
	}, withAttribution(httpClient, cfg.AIMLAPIServerURL), models, providerName, providerConfig)
}
