package providers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/csum"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/graphiti"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/providers/anthropic"
	"pentagi/pkg/providers/bedrock"
	"pentagi/pkg/providers/custom"
	"pentagi/pkg/providers/embeddings"
	"pentagi/pkg/providers/gemini"
	"pentagi/pkg/providers/ollama"
	"pentagi/pkg/providers/openai"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
)

const deltaCallCounter = 10000

const defaultTestParallelWorkersNumber = 16

const pentestDockerImage = "vxcontrol/kali-linux"

type ProviderController interface {
	NewFlowProvider(
		ctx context.Context,
		prvname provider.ProviderName,
		prompter templates.Prompter,
		executor tools.FlowToolsExecutor,
		flowID, userID int64,
		askUser bool,
		input string,
	) (FlowProvider, error)
	LoadFlowProvider(
		ctx context.Context,
		prvname provider.ProviderName,
		prompter templates.Prompter,
		executor tools.FlowToolsExecutor,
		flowID, userID int64,
		askUser bool,
		image, language, title, tcIDTemplate string,
	) (FlowProvider, error)
	NewAssistantProvider(
		ctx context.Context,
		prvname provider.ProviderName,
		prompter templates.Prompter,
		executor tools.FlowToolsExecutor,
		assistantID, flowID, userID int64,
		image, input string,
		streamCb StreamMessageHandler,
	) (AssistantProvider, error)
	LoadAssistantProvider(
		ctx context.Context,
		prvname provider.ProviderName,
		prompter templates.Prompter,
		executor tools.FlowToolsExecutor,
		assistantID, flowID, userID int64,
		image, language, title, tcIDTemplate string,
		streamCb StreamMessageHandler,
	) (AssistantProvider, error)

	Embedder() embeddings.Embedder
	GraphitiClient() *graphiti.Client
	DefaultProviders() provider.Providers
	DefaultProvidersConfig() provider.ProvidersConfig
	GetProvider(
		ctx context.Context,
		prvname provider.ProviderName,
		userID int64,
	) (provider.Provider, error)
	GetProviders(
		ctx context.Context,
		userID int64,
	) (provider.Providers, error)

	NewProvider(prv database.Provider) (provider.Provider, error)
	CreateProvider(
		ctx context.Context,
		userID int64,
		prvname provider.ProviderName,
		prvtype provider.ProviderType,
		config *pconfig.ProviderConfig,
	) (database.Provider, error)
	UpdateProvider(
		ctx context.Context,
		userID int64,
		prvID int64,
		prvname provider.ProviderName,
		config *pconfig.ProviderConfig,
	) (database.Provider, error)
	DeleteProvider(
		ctx context.Context,
		userID int64,
		prvID int64,
	) (database.Provider, error)

	TestAgent(
		ctx context.Context,
		prvtype provider.ProviderType,
		agentType pconfig.ProviderOptionsType,
		config *pconfig.AgentConfig,
	) (tester.AgentTestResults, error)
	TestProvider(
		ctx context.Context,
		prvtype provider.ProviderType,
		config *pconfig.ProviderConfig,
	) (tester.ProviderTestResults, error)
}

type providerController struct {
	db             database.Querier
	cfg            *config.Config
	docker         docker.DockerClient
	publicIP       string
	embedder       embeddings.Embedder
	graphitiClient *graphiti.Client

	startCallNumber *atomic.Int64

	defaultDockerImageForPentest string

	summarizerAgent     csum.Summarizer
	summarizerAssistant csum.Summarizer

	defaultConfigs provider.ProvidersConfig

	provider.Providers
}

func NewProviderController(
	cfg *config.Config,
	db database.Querier,
	docker docker.DockerClient,
) (ProviderController, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	embedder, err := embeddings.New(cfg)
	if err != nil {
		logrus.WithError(err).Errorf("failed to create embedder '%s'", cfg.EmbeddingProvider)
	}

	providers := make(provider.Providers)
	defaultConfigs := make(provider.ProvidersConfig)

	if config, err := openai.DefaultProviderConfig(); err != nil {
		return nil, fmt.Errorf("failed to create openai provider config: %w", err)
	} else {
		defaultConfigs[provider.ProviderOpenAI] = config
	}

	if config, err := anthropic.DefaultProviderConfig(); err != nil {
		return nil, fmt.Errorf("failed to create anthropic provider config: %w", err)
	} else {
		defaultConfigs[provider.ProviderAnthropic] = config
	}

	if config, err := gemini.DefaultProviderConfig(); err != nil {
		return nil, fmt.Errorf("failed to create gemini provider config: %w", err)
	} else {
		defaultConfigs[provider.ProviderGemini] = config
	}

	if config, err := bedrock.DefaultProviderConfig(); err != nil {
		return nil, fmt.Errorf("failed to create bedrock provider config: %w", err)
	} else {
		defaultConfigs[provider.ProviderBedrock] = config
	}

	if config, err := ollama.DefaultProviderConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to create ollama provider config: %w", err)
	} else {
		defaultConfigs[provider.ProviderOllama] = config
	}

	if config, err := custom.DefaultProviderConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to create custom provider config: %w", err)
	} else {
		defaultConfigs[provider.ProviderCustom] = config
	}

	if cfg.OpenAIKey != "" {
		p, err := openai.New(cfg, defaultConfigs[provider.ProviderOpenAI])
		if err != nil {
			return nil, fmt.Errorf("failed to create openai provider: %w", err)
		}

		providers[provider.DefaultProviderNameOpenAI] = p
	}

	if cfg.AnthropicAPIKey != "" {
		p, err := anthropic.New(cfg, defaultConfigs[provider.ProviderAnthropic])
		if err != nil {
			return nil, fmt.Errorf("failed to create anthropic provider: %w", err)
		}

		providers[provider.DefaultProviderNameAnthropic] = p
	}

	if cfg.GeminiAPIKey != "" {
		p, err := gemini.New(cfg, defaultConfigs[provider.ProviderGemini])
		if err != nil {
			return nil, fmt.Errorf("failed to create gemini provider: %w", err)
		}

		providers[provider.DefaultProviderNameGemini] = p
	}

	if cfg.BedrockAccessKey != "" && cfg.BedrockSecretKey != "" {
		p, err := bedrock.New(cfg, defaultConfigs[provider.ProviderBedrock])
		if err != nil {
			return nil, fmt.Errorf("failed to create bedrock provider: %w", err)
		}
		providers[provider.DefaultProviderNameBedrock] = p
	}

	if cfg.OllamaServerURL != "" {
		p, err := ollama.New(cfg, defaultConfigs[provider.ProviderOllama])
		if err != nil {
			return nil, fmt.Errorf("failed to create ollama provider: %w", err)
		}
		providers[provider.DefaultProviderNameOllama] = p
	}

	if cfg.LLMServerURL != "" && (cfg.LLMServerModel != "" || cfg.LLMServerConfig != "") {
		p, err := custom.New(cfg, defaultConfigs[provider.ProviderCustom])
		if err != nil {
			return nil, fmt.Errorf("failed to create custom provider: %w", err)
		}

		providers[provider.DefaultProviderNameCustom] = p
	}

	summarizerAgent := csum.NewSummarizer(csum.SummarizerConfig{
		PreserveLast:   cfg.SummarizerPreserveLast,
		UseQA:          cfg.SummarizerUseQA,
		SummHumanInQA:  cfg.SummarizerSumHumanInQA,
		LastSecBytes:   cfg.SummarizerLastSecBytes,
		MaxBPBytes:     cfg.SummarizerMaxBPBytes,
		MaxQASections:  cfg.SummarizerMaxQASections,
		MaxQABytes:     cfg.SummarizerMaxQABytes,
		KeepQASections: cfg.SummarizerKeepQASections,
	})

	summarizerAssistant := csum.NewSummarizer(csum.SummarizerConfig{
		PreserveLast:   cfg.AssistantSummarizerPreserveLast,
		UseQA:          true,
		SummHumanInQA:  false,
		LastSecBytes:   cfg.AssistantSummarizerLastSecBytes,
		MaxBPBytes:     cfg.AssistantSummarizerMaxBPBytes,
		MaxQASections:  cfg.AssistantSummarizerMaxQASections,
		MaxQABytes:     cfg.AssistantSummarizerMaxQABytes,
		KeepQASections: cfg.AssistantSummarizerKeepQASections,
	})

	graphitiClient, err := graphiti.NewClient(
		cfg.GraphitiURL,
		time.Duration(cfg.GraphitiTimeout)*time.Second,
		cfg.GraphitiEnabled && cfg.GraphitiURL != "",
	)
	if err != nil {
		logrus.WithError(err).Warn("failed to initialize graphiti client, continuing without it")
		graphitiClient = &graphiti.Client{}
	}

	return &providerController{
		db:             db,
		cfg:            cfg,
		docker:         docker,
		publicIP:       cfg.DockerPublicIP,
		embedder:       embedder,
		graphitiClient: graphitiClient,

		startCallNumber: newAtomicInt64(0), // 0 means to make it random

		defaultDockerImageForPentest: cfg.DockerDefaultImageForPentest,

		summarizerAgent:     summarizerAgent,
		summarizerAssistant: summarizerAssistant,

		defaultConfigs: defaultConfigs,

		Providers: providers,
	}, nil
}

func (pc *providerController) NewFlowProvider(
	ctx context.Context,
	prvname provider.ProviderName,
	prompter templates.Prompter,
	executor tools.FlowToolsExecutor,
	flowID, userID int64,
	askUser bool,
	input string,
) (FlowProvider, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.NewFlowProvider")
	defer span.End()

	prv, err := pc.GetProvider(ctx, prvname, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	imageTmpl, err := prompter.RenderTemplate(templates.PromptTypeImageChooser, map[string]any{
		"DefaultImage":           pc.docker.GetDefaultImage(),
		"DefaultImageForPentest": pc.defaultDockerImageForPentest,
		"Input":                  input,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get primary docker image template: %w", err)
	}

	image, err := prv.Call(ctx, pconfig.OptionsTypeSimple, imageTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to get primary docker image: %w", err)
	}
	image = strings.ToLower(strings.TrimSpace(image))

	languageTmpl, err := prompter.RenderTemplate(templates.PromptTypeLanguageChooser, map[string]any{
		"Input": input,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get language template: %w", err)
	}

	language, err := prv.Call(ctx, pconfig.OptionsTypeSimple, languageTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to get language: %w", err)
	}
	language = strings.TrimSpace(language)

	titleTmpl, err := prompter.RenderTemplate(templates.PromptTypeFlowDescriptor, map[string]any{
		"Input":       input,
		"Lang":        language,
		"CurrentTime": getCurrentTime(),
		"N":           20,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get flow title template: %w", err)
	}

	title, err := prv.Call(ctx, pconfig.OptionsTypeSimple, titleTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow title: %w", err)
	}
	title = strings.TrimSpace(title)

	tcIDTemplate, err := prv.GetToolCallIDTemplate(ctx, prompter)
	if err != nil {
		return nil, fmt.Errorf("failed to determine tool call ID template: %w", err)
	}

	fp := &flowProvider{
		db:             pc.db,
		mx:             &sync.RWMutex{},
		embedder:       pc.embedder,
		graphitiClient: pc.graphitiClient,
		flowID:         flowID,
		publicIP:       pc.publicIP,
		callCounter:    newAtomicInt64(pc.startCallNumber.Add(deltaCallCounter)),
		image:          image,
		title:          title,
		language:       language,
		askUser:        askUser,
		tcIDTemplate:   tcIDTemplate,
		prompter:       prompter,
		executor:       executor,
		summarizer:     pc.summarizerAgent,
		Provider:       prv,
	}

	return fp, nil
}

func (pc *providerController) LoadFlowProvider(
	ctx context.Context,
	prvname provider.ProviderName,
	prompter templates.Prompter,
	executor tools.FlowToolsExecutor,
	flowID, userID int64,
	askUser bool,
	image, language, title, tcIDTemplate string,
) (FlowProvider, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.LoadFlowProvider")
	defer span.End()

	prv, err := pc.GetProvider(ctx, prvname, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	fp := &flowProvider{
		db:             pc.db,
		mx:             &sync.RWMutex{},
		embedder:       pc.embedder,
		graphitiClient: pc.graphitiClient,
		flowID:         flowID,
		publicIP:       pc.publicIP,
		callCounter:    newAtomicInt64(pc.startCallNumber.Add(deltaCallCounter)),
		image:          image,
		title:          title,
		language:       language,
		askUser:        askUser,
		tcIDTemplate:   tcIDTemplate,
		prompter:       prompter,
		executor:       executor,
		summarizer:     pc.summarizerAgent,
		Provider:       prv,
	}

	return fp, nil
}

func (pc *providerController) Embedder() embeddings.Embedder {
	return pc.embedder
}

func (pc *providerController) GraphitiClient() *graphiti.Client {
	return pc.graphitiClient
}

func (pc *providerController) NewAssistantProvider(
	ctx context.Context,
	prvname provider.ProviderName,
	prompter templates.Prompter,
	executor tools.FlowToolsExecutor,
	assistantID, flowID, userID int64,
	image, input string,
	streamCb StreamMessageHandler,
) (AssistantProvider, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.NewAssistantProvider")
	defer span.End()

	prv, err := pc.GetProvider(ctx, prvname, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	languageTmpl, err := prompter.RenderTemplate(templates.PromptTypeLanguageChooser, map[string]any{
		"Input": input,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get language template: %w", err)
	}

	language, err := prv.Call(ctx, pconfig.OptionsTypeSimple, languageTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to get language: %w", err)
	}
	language = strings.TrimSpace(language)

	titleTmpl, err := prompter.RenderTemplate(templates.PromptTypeFlowDescriptor, map[string]any{
		"Input":       input,
		"Lang":        language,
		"CurrentTime": getCurrentTime(),
		"N":           20,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get flow title template: %w", err)
	}

	title, err := prv.Call(ctx, pconfig.OptionsTypeSimple, titleTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow title: %w", err)
	}
	title = strings.TrimSpace(title)

	tcIDTemplate, err := prv.GetToolCallIDTemplate(ctx, prompter)
	if err != nil {
		return nil, fmt.Errorf("failed to determine tool call ID template: %w", err)
	}

	ap := &assistantProvider{
		id:         assistantID,
		summarizer: pc.summarizerAssistant,
		fp: flowProvider{
			db:             pc.db,
			mx:             &sync.RWMutex{},
			embedder:       pc.embedder,
			graphitiClient: pc.graphitiClient,
			flowID:         flowID,
			publicIP:       pc.publicIP,
			callCounter:    newAtomicInt64(pc.startCallNumber.Add(deltaCallCounter)),
			image:          image,
			title:          title,
			language:       language,
			tcIDTemplate:   tcIDTemplate,
			prompter:       prompter,
			executor:       executor,
			streamCb:       streamCb,
			summarizer:     pc.summarizerAgent,
			Provider:       prv,
		},
	}

	return ap, nil
}

func (pc *providerController) LoadAssistantProvider(
	ctx context.Context,
	prvname provider.ProviderName,
	prompter templates.Prompter,
	executor tools.FlowToolsExecutor,
	assistantID, flowID, userID int64,
	image, language, title, tcIDTemplate string,
	streamCb StreamMessageHandler,
) (AssistantProvider, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.LoadAssistantProvider")
	defer span.End()

	prv, err := pc.GetProvider(ctx, prvname, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	ap := &assistantProvider{
		id:         assistantID,
		summarizer: pc.summarizerAssistant,
		fp: flowProvider{
			db:             pc.db,
			mx:             &sync.RWMutex{},
			embedder:       pc.embedder,
			graphitiClient: pc.graphitiClient,
			flowID:         flowID,
			publicIP:       pc.publicIP,
			callCounter:    newAtomicInt64(pc.startCallNumber.Add(deltaCallCounter)),
			image:          image,
			title:          title,
			language:       language,
			tcIDTemplate:   tcIDTemplate,
			prompter:       prompter,
			executor:       executor,
			streamCb:       streamCb,
			summarizer:     pc.summarizerAgent,
			Provider:       prv,
		},
	}

	return ap, nil
}

func (pc *providerController) DefaultProviders() provider.Providers {
	return pc.Providers
}

func (pc *providerController) DefaultProvidersConfig() provider.ProvidersConfig {
	return pc.defaultConfigs
}

func (pc *providerController) GetProvider(
	ctx context.Context,
	prvname provider.ProviderName,
	userID int64,
) (provider.Provider, error) {
	// Lookup default providers first
	switch prvname {
	case provider.DefaultProviderNameOpenAI:
		return pc.Providers.Get(provider.DefaultProviderNameOpenAI)
	case provider.DefaultProviderNameAnthropic:
		return pc.Providers.Get(provider.DefaultProviderNameAnthropic)
	case provider.DefaultProviderNameGemini:
		return pc.Providers.Get(provider.DefaultProviderNameGemini)
	case provider.DefaultProviderNameBedrock:
		return pc.Providers.Get(provider.DefaultProviderNameBedrock)
	case provider.DefaultProviderNameOllama:
		return pc.Providers.Get(provider.DefaultProviderNameOllama)
	case provider.DefaultProviderNameCustom:
		return pc.Providers.Get(provider.DefaultProviderNameCustom)
	}

	// Lookup user defined providers by name and build it
	prv, err := pc.db.GetUserProviderByName(ctx, database.GetUserProviderByNameParams{
		Name:   string(prvname),
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get provider '%s' from database: %w", prvname, err)
	}

	return pc.NewProvider(prv)
}

func (pc *providerController) GetProviders(
	ctx context.Context,
	userID int64,
) (provider.Providers, error) {
	providersMap := make(provider.Providers, len(pc.Providers))

	// Copy default providers
	for prvname, prv := range pc.Providers {
		providersMap[prvname] = prv
	}

	// Copy user providers
	providers, err := pc.db.GetUserProviders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user providers: %w", err)
	}

	for _, prv := range providers {
		p, err := pc.NewProvider(prv)
		if err != nil {
			return nil, fmt.Errorf("failed to build provider: %w", err)
		}
		providersMap[provider.ProviderName(prv.Name)] = p
	}

	return providersMap, nil
}

func (pc *providerController) NewProvider(prv database.Provider) (provider.Provider, error) {
	if len(prv.Config) == 0 {
		prv.Config = []byte(pconfig.EmptyProviderConfigRaw)
	}

	// Check if the provider type is available via check default one
	providerType := provider.ProviderType(prv.Type)
	if !pc.ListTypes().Contains(providerType) {
		return nil, fmt.Errorf("provider type '%s' is not available", prv.Type)
	}

	switch providerType {
	case provider.ProviderOpenAI:
		openaiConfig, err := openai.BuildProviderConfig(prv.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to build openai provider config: %w", err)
		}
		return openai.New(pc.cfg, openaiConfig)
	case provider.ProviderAnthropic:
		anthropicConfig, err := anthropic.BuildProviderConfig(prv.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to build anthropic provider config: %w", err)
		}
		return anthropic.New(pc.cfg, anthropicConfig)
	case provider.ProviderGemini:
		geminiConfig, err := gemini.BuildProviderConfig(prv.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to build gemini provider config: %w", err)
		}
		return gemini.New(pc.cfg, geminiConfig)
	case provider.ProviderBedrock:
		bedrockConfig, err := bedrock.BuildProviderConfig(prv.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to build bedrock provider config: %w", err)
		}
		return bedrock.New(pc.cfg, bedrockConfig)
	case provider.ProviderOllama:
		ollamaConfig, err := ollama.BuildProviderConfig(pc.cfg, prv.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to build ollama provider config: %w", err)
		}
		return ollama.New(pc.cfg, ollamaConfig)
	case provider.ProviderCustom:
		customConfig, err := custom.BuildProviderConfig(pc.cfg, prv.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to build custom provider config: %w", err)
		}
		return custom.New(pc.cfg, customConfig)
	default:
		return nil, fmt.Errorf("unknown provider type: %s", prv.Type)
	}
}

func (pc *providerController) CreateProvider(
	ctx context.Context,
	userID int64,
	prvname provider.ProviderName,
	prvtype provider.ProviderType,
	config *pconfig.ProviderConfig,
) (database.Provider, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.CreateProvider")
	defer span.End()

	var (
		err    error
		result database.Provider
	)

	if config, err = pc.patchProviderConfig(prvtype, config); err != nil {
		return result, fmt.Errorf("failed to patch provider config: %w", err)
	}

	rawConfig, err := json.Marshal(config)
	if err != nil {
		return result, fmt.Errorf("failed to marshal provider config: %w", err)
	}

	result, err = pc.db.CreateProvider(ctx, database.CreateProviderParams{
		UserID: userID,
		Type:   database.ProviderType(prvtype),
		Name:   string(prvname),
		Config: rawConfig,
	})
	if err != nil {
		return result, fmt.Errorf("failed to create provider: %w", err)
	}

	return result, nil
}

func (pc *providerController) UpdateProvider(
	ctx context.Context,
	userID int64,
	prvID int64,
	prvname provider.ProviderName,
	config *pconfig.ProviderConfig,
) (database.Provider, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.UpdateProvider")
	defer span.End()

	var (
		err    error
		result database.Provider
	)

	prv, err := pc.db.GetUserProvider(ctx, database.GetUserProviderParams{
		ID:     prvID,
		UserID: userID,
	})
	if err != nil {
		return result, fmt.Errorf("failed to get provider: %w", err)
	}
	prvtype := provider.ProviderType(prv.Type)

	if config, err = pc.patchProviderConfig(prvtype, config); err != nil {
		return result, fmt.Errorf("failed to patch provider config: %w", err)
	}

	rawConfig, err := json.Marshal(config)
	if err != nil {
		return result, fmt.Errorf("failed to marshal provider config: %w", err)
	}

	result, err = pc.db.UpdateUserProvider(ctx, database.UpdateUserProviderParams{
		ID:     prvID,
		UserID: userID,
		Name:   string(prvname),
		Config: rawConfig,
	})
	if err != nil {
		return result, fmt.Errorf("failed to update provider: %w", err)
	}

	return result, nil
}

func (pc *providerController) DeleteProvider(
	ctx context.Context,
	userID int64,
	prvID int64,
) (database.Provider, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.DeleteProvider")
	defer span.End()

	result, err := pc.db.DeleteUserProvider(ctx, database.DeleteUserProviderParams{
		ID:     prvID,
		UserID: userID,
	})
	if err != nil {
		return result, fmt.Errorf("failed to delete provider: %w", err)
	}

	return result, nil
}

func (pc *providerController) TestAgent(
	ctx context.Context,
	prvtype provider.ProviderType,
	agentType pconfig.ProviderOptionsType,
	config *pconfig.AgentConfig,
) (tester.AgentTestResults, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.TestAgent")
	defer span.End()

	var result tester.AgentTestResults

	// Create provider config with single agent configuration
	testConfig := &pconfig.ProviderConfig{}

	// Set the agent config to the appropriate field based on agent type
	switch agentType {
	case pconfig.OptionsTypeSimple:
		testConfig.Simple = config
	case pconfig.OptionsTypeSimpleJSON:
		testConfig.SimpleJSON = config
	case pconfig.OptionsTypePrimaryAgent:
		testConfig.PrimaryAgent = config
	case pconfig.OptionsTypeAssistant:
		testConfig.Assistant = config
	case pconfig.OptionsTypeGenerator:
		testConfig.Generator = config
	case pconfig.OptionsTypeRefiner:
		testConfig.Refiner = config
	case pconfig.OptionsTypeAdviser:
		testConfig.Adviser = config
	case pconfig.OptionsTypeReflector:
		testConfig.Reflector = config
	case pconfig.OptionsTypeSearcher:
		testConfig.Searcher = config
	case pconfig.OptionsTypeEnricher:
		testConfig.Enricher = config
	case pconfig.OptionsTypeCoder:
		testConfig.Coder = config
	case pconfig.OptionsTypeInstaller:
		testConfig.Installer = config
	case pconfig.OptionsTypePentester:
		testConfig.Pentester = config
	default:
		return result, fmt.Errorf("unsupported agent type: %s", agentType)
	}

	// Patch with defaults
	patchedConfig, err := pc.patchProviderConfig(prvtype, testConfig)
	if err != nil {
		return result, fmt.Errorf("failed to patch provider config: %w", err)
	}

	// Create temporary provider for testing using existing provider logic
	tempProvider, err := pc.buildProviderFromConfig(prvtype, patchedConfig)
	if err != nil {
		return result, fmt.Errorf("failed to create provider for testing: %w", err)
	}

	// Run tests for specific agent type only
	results, err := tester.TestProvider(
		ctx,
		tempProvider,
		tester.WithAgentTypes(agentType),
		tester.WithVerbose(false),
		tester.WithParallelWorkers(defaultTestParallelWorkersNumber),
	)
	if err != nil {
		return result, fmt.Errorf("failed to test agent: %w", err)
	}

	// Extract results for the specific agent type
	switch agentType {
	case pconfig.OptionsTypeSimple:
		result = results.Simple
	case pconfig.OptionsTypeSimpleJSON:
		result = results.SimpleJSON
	case pconfig.OptionsTypePrimaryAgent:
		result = results.PrimaryAgent
	case pconfig.OptionsTypeAssistant:
		result = results.Assistant
	case pconfig.OptionsTypeGenerator:
		result = results.Generator
	case pconfig.OptionsTypeRefiner:
		result = results.Refiner
	case pconfig.OptionsTypeAdviser:
		result = results.Adviser
	case pconfig.OptionsTypeReflector:
		result = results.Reflector
	case pconfig.OptionsTypeSearcher:
		result = results.Searcher
	case pconfig.OptionsTypeEnricher:
		result = results.Enricher
	case pconfig.OptionsTypeCoder:
		result = results.Coder
	case pconfig.OptionsTypeInstaller:
		result = results.Installer
	case pconfig.OptionsTypePentester:
		result = results.Pentester
	default:
		return result, fmt.Errorf("unexpected agent type: %s", agentType)
	}

	return result, nil
}

func (pc *providerController) TestProvider(
	ctx context.Context,
	prvtype provider.ProviderType,
	config *pconfig.ProviderConfig,
) (tester.ProviderTestResults, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.TestProvider")
	defer span.End()

	var results tester.ProviderTestResults

	// Patch config with defaults
	patchedConfig, err := pc.patchProviderConfig(prvtype, config)
	if err != nil {
		return results, fmt.Errorf("failed to patch provider config: %w", err)
	}

	// Create provider for testing
	testProvider, err := pc.buildProviderFromConfig(prvtype, patchedConfig)
	if err != nil {
		return results, fmt.Errorf("failed to create provider for testing: %w", err)
	}

	// Run full provider testing
	results, err = tester.TestProvider(
		ctx,
		testProvider,
		tester.WithVerbose(false),
		tester.WithParallelWorkers(defaultTestParallelWorkersNumber),
	)
	if err != nil {
		return results, fmt.Errorf("failed to test provider: %w", err)
	}

	return results, nil
}

func (pc *providerController) patchProviderConfig(
	prvtype provider.ProviderType,
	config *pconfig.ProviderConfig,
) (*pconfig.ProviderConfig, error) {
	var (
		defaultCfg *pconfig.ProviderConfig
		ok         bool
	)

	if defaultCfg, ok = pc.defaultConfigs[prvtype]; !ok {
		return nil, fmt.Errorf("default provider config not found for type: %s", prvtype.String())
	}

	if config == nil {
		return defaultCfg, nil
	}

	if config.Simple == nil {
		config.Simple = defaultCfg.Simple
	}
	if config.SimpleJSON == nil {
		config.SimpleJSON = defaultCfg.SimpleJSON
	}
	if config.PrimaryAgent == nil {
		config.PrimaryAgent = defaultCfg.PrimaryAgent
	}
	if config.Assistant == nil {
		config.Assistant = defaultCfg.Assistant
	}
	if config.Generator == nil {
		config.Generator = defaultCfg.Generator
	}
	if config.Refiner == nil {
		config.Refiner = defaultCfg.Refiner
	}
	if config.Adviser == nil {
		config.Adviser = defaultCfg.Adviser
	}
	if config.Reflector == nil {
		config.Reflector = defaultCfg.Reflector
	}
	if config.Searcher == nil {
		config.Searcher = defaultCfg.Searcher
	}
	if config.Enricher == nil {
		config.Enricher = defaultCfg.Enricher
	}
	if config.Coder == nil {
		config.Coder = defaultCfg.Coder
	}
	if config.Installer == nil {
		config.Installer = defaultCfg.Installer
	}
	if config.Pentester == nil {
		config.Pentester = defaultCfg.Pentester
	}

	config.SetDefaultOptions(defaultCfg.GetDefaultOptions())

	return config, nil
}

func (pc *providerController) buildProviderFromConfig(
	prvtype provider.ProviderType,
	config *pconfig.ProviderConfig,
) (provider.Provider, error) {
	switch prvtype {
	case provider.ProviderOpenAI:
		return openai.New(pc.cfg, config)
	case provider.ProviderAnthropic:
		return anthropic.New(pc.cfg, config)
	case provider.ProviderCustom:
		return custom.New(pc.cfg, config)
	case provider.ProviderGemini:
		return gemini.New(pc.cfg, config)
	case provider.ProviderBedrock:
		return bedrock.New(pc.cfg, config)
	case provider.ProviderOllama:
		return ollama.New(pc.cfg, config)
	default:
		return nil, fmt.Errorf("unknown provider type: %s", prvtype)
	}
}

func newAtomicInt64(seed int64) *atomic.Int64 {
	var number atomic.Int64

	if seed == 0 {
		bigID, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			return &number
		}
		seed = bigID.Int64()
	}

	number.Store(seed)
	return &number
}
