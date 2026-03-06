package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/graphiti"
	"pentagi/pkg/providers/embeddings"
	"pentagi/pkg/schema"

	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/vectorstores/pgvector"
)

type ExecutorHandler func(ctx context.Context, name string, args json.RawMessage) (string, error)

type SummarizeHandler func(ctx context.Context, result string) (string, error)

type Functions struct {
	Token    *string            `form:"token,omitempty" json:"token,omitempty" validate:"omitempty"`
	Disabled []DisableFunction  `form:"disabled,omitempty" json:"disabled,omitempty" validate:"omitempty,valid"`
	Function []ExternalFunction `form:"functions,omitempty" json:"functions,omitempty" validate:"omitempty,valid"`
}

func (f *Functions) Scan(input any) error {
	switch v := input.(type) {
	case string:
		return json.Unmarshal([]byte(v), f)
	case []byte:
		return json.Unmarshal(v, f)
	case json.RawMessage:
		return json.Unmarshal(v, f)
	}
	return fmt.Errorf("unsupported type of input value to scan")
}

type DisableFunction struct {
	Name    string   `form:"name" json:"name" validate:"required"`
	Context []string `form:"context,omitempty" json:"context,omitempty" validate:"omitempty,dive,oneof=agent adviser coder searcher generator memorist enricher reporter assistant,required"`
}

type ExternalFunction struct {
	Name    string        `form:"name" json:"name" validate:"required"`
	URL     string        `form:"url" json:"url" validate:"required,url" example:"https://example.com/api/v1/function"`
	Timeout *int64        `form:"timeout,omitempty" json:"timeout,omitempty" validate:"omitempty,min=1" example:"60"`
	Context []string      `form:"context,omitempty" json:"context,omitempty" validate:"omitempty,dive,oneof=agent adviser coder searcher generator memorist enricher reporter assistant,required"`
	Schema  schema.Schema `form:"schema" json:"schema" validate:"required" swaggertype:"object"`
}

type FunctionInfo struct {
	Name   string
	Schema string
}

type Tool interface {
	Handle(ctx context.Context, name string, args json.RawMessage) (string, error)
	IsAvailable() bool
}

type ScreenshotProvider interface {
	PutScreenshot(ctx context.Context, name, url string, taskID, subtaskID *int64) (int64, error)
}

type AgentLogProvider interface {
	PutLog(
		ctx context.Context,
		initiator, executor database.MsgchainType,
		task, result string,
		taskID, subtaskID *int64,
	) (int64, error)
}

type MsgLogProvider interface {
	PutMsg(
		ctx context.Context,
		msgType database.MsglogType,
		taskID, subtaskID *int64,
		streamID int64,
		thinking, msg string,
	) (int64, error)
	UpdateMsgResult(
		ctx context.Context,
		msgID, streamID int64,
		result string,
		resultFormat database.MsglogResultFormat,
	) error
}

type SearchLogProvider interface {
	PutLog(
		ctx context.Context,
		initiator database.MsgchainType,
		executor database.MsgchainType,
		engine database.SearchengineType,
		query string,
		result string,
		taskID *int64,
		subtaskID *int64,
	) (int64, error)
}

type TermLogProvider interface {
	PutMsg(
		ctx context.Context,
		msgType database.TermlogType,
		msg string,
		containerID int64,
		taskID, subtaskID *int64,
	) (int64, error)
}

type VectorStoreLogProvider interface {
	PutLog(
		ctx context.Context,
		initiator database.MsgchainType,
		executor database.MsgchainType,
		filter json.RawMessage,
		query string,
		action database.VecstoreActionType,
		result string,
		taskID *int64,
		subtaskID *int64,
	) (int64, error)
}

type flowToolsExecutor struct {
	flowID int64
	scp    ScreenshotProvider
	alp    AgentLogProvider
	mlp    MsgLogProvider
	slp    SearchLogProvider
	tlp    TermLogProvider
	vslp   VectorStoreLogProvider

	db             database.Querier
	cfg            *config.Config
	store          *pgvector.Store
	graphitiClient *graphiti.Client
	image          string
	docker         docker.DockerClient
	primaryID      int64
	primaryLID     string
	functions      *Functions

	definitions map[string]llms.FunctionDefinition
	handlers    map[string]ExecutorHandler
}

type ContextToolsExecutor interface {
	Tools() []llms.Tool
	Execute(ctx context.Context, streamID int64, id, name, thinking string, args json.RawMessage) (string, error)
	IsBarrierFunction(name string) bool
	GetBarrierToolNames() []string
	GetBarrierTools() []FunctionInfo
	GetToolSchema(name string) (*schema.Schema, error)
}

type CustomExecutorConfig struct {
	TaskID      *int64
	SubtaskID   *int64
	Builtin     []string
	Definitions []llms.FunctionDefinition
	Handlers    map[string]ExecutorHandler
	Barriers    []string
	Summarizer  SummarizeHandler
}

type AssistantExecutorConfig struct {
	UseAgents  bool
	Adviser    ExecutorHandler
	Coder      ExecutorHandler
	Installer  ExecutorHandler
	Memorist   ExecutorHandler
	Pentester  ExecutorHandler
	Searcher   ExecutorHandler
	Summarizer SummarizeHandler
}

type PrimaryExecutorConfig struct {
	TaskID     int64
	SubtaskID  int64
	Barrier    ExecutorHandler
	Adviser    ExecutorHandler
	Coder      ExecutorHandler
	Installer  ExecutorHandler
	Memorist   ExecutorHandler
	Pentester  ExecutorHandler
	Searcher   ExecutorHandler
	Summarizer SummarizeHandler
}

type InstallerExecutorConfig struct {
	TaskID            *int64
	SubtaskID         *int64
	Adviser           ExecutorHandler
	Memorist          ExecutorHandler
	Searcher          ExecutorHandler
	MaintenanceResult ExecutorHandler
	Summarizer        SummarizeHandler
}

type CoderExecutorConfig struct {
	TaskID     *int64
	SubtaskID  *int64
	Adviser    ExecutorHandler
	Installer  ExecutorHandler
	Memorist   ExecutorHandler
	Searcher   ExecutorHandler
	CodeResult ExecutorHandler
	Summarizer SummarizeHandler
}

type PentesterExecutorConfig struct {
	TaskID     *int64
	SubtaskID  *int64
	Adviser    ExecutorHandler
	Coder      ExecutorHandler
	Installer  ExecutorHandler
	Memorist   ExecutorHandler
	Searcher   ExecutorHandler
	HackResult ExecutorHandler
	Summarizer SummarizeHandler
}

type SearcherExecutorConfig struct {
	TaskID       *int64
	SubtaskID    *int64
	Memorist     ExecutorHandler
	SearchResult ExecutorHandler
	Summarizer   SummarizeHandler
}

type GeneratorExecutorConfig struct {
	TaskID      int64
	Memorist    ExecutorHandler
	Searcher    ExecutorHandler
	SubtaskList ExecutorHandler
}

type RefinerExecutorConfig struct {
	TaskID       int64
	Memorist     ExecutorHandler
	Searcher     ExecutorHandler
	SubtaskPatch ExecutorHandler
}

type MemoristExecutorConfig struct {
	TaskID       *int64
	SubtaskID    *int64
	SearchResult ExecutorHandler
	Summarizer   SummarizeHandler
}

type EnricherExecutorConfig struct {
	TaskID         *int64
	SubtaskID      *int64
	Memorist       ExecutorHandler
	Searcher       ExecutorHandler
	EnricherResult ExecutorHandler
}

type ReporterExecutorConfig struct {
	TaskID       *int64
	SubtaskID    *int64
	ReportResult ExecutorHandler
}

type FlowToolsExecutor interface {
	SetFlowID(flowID int64)
	SetImage(image string)
	SetEmbedder(embedder embeddings.Embedder)
	SetFunctions(functions *Functions)
	SetScreenshotProvider(sp ScreenshotProvider)
	SetAgentLogProvider(alp AgentLogProvider)
	SetMsgLogProvider(mlp MsgLogProvider)
	SetSearchLogProvider(slp SearchLogProvider)
	SetTermLogProvider(tlp TermLogProvider)
	SetVectorStoreLogProvider(vslp VectorStoreLogProvider)
	SetGraphitiClient(client *graphiti.Client)

	Prepare(ctx context.Context) error
	Release(ctx context.Context) error
	GetCustomExecutor(cfg CustomExecutorConfig) (ContextToolsExecutor, error)
	GetAssistantExecutor(cfg AssistantExecutorConfig) (ContextToolsExecutor, error)
	GetPrimaryExecutor(cfg PrimaryExecutorConfig) (ContextToolsExecutor, error)
	GetInstallerExecutor(cfg InstallerExecutorConfig) (ContextToolsExecutor, error)
	GetCoderExecutor(cfg CoderExecutorConfig) (ContextToolsExecutor, error)
	GetPentesterExecutor(cfg PentesterExecutorConfig) (ContextToolsExecutor, error)
	GetSearcherExecutor(cfg SearcherExecutorConfig) (ContextToolsExecutor, error)
	GetGeneratorExecutor(cfg GeneratorExecutorConfig) (ContextToolsExecutor, error)
	GetRefinerExecutor(cfg RefinerExecutorConfig) (ContextToolsExecutor, error)
	GetMemoristExecutor(cfg MemoristExecutorConfig) (ContextToolsExecutor, error)
	GetEnricherExecutor(cfg EnricherExecutorConfig) (ContextToolsExecutor, error)
	GetReporterExecutor(cfg ReporterExecutorConfig) (ContextToolsExecutor, error)
}

func NewFlowToolsExecutor(
	db database.Querier,
	cfg *config.Config,
	docker docker.DockerClient,
	functions *Functions,
	flowID int64,
) (FlowToolsExecutor, error) {
	return &flowToolsExecutor{
		db:          db,
		docker:      docker,
		functions:   functions,
		cfg:         cfg,
		flowID:      flowID,
		definitions: make(map[string]llms.FunctionDefinition),
		handlers:    make(map[string]ExecutorHandler),
	}, nil
}

func (fte *flowToolsExecutor) SetFlowID(flowID int64) {
	fte.flowID = flowID
}

func (fte *flowToolsExecutor) SetImage(image string) {
	fte.image = image
}

func (fte *flowToolsExecutor) SetEmbedder(embedder embeddings.Embedder) {
	if !embedder.IsAvailable() {
		return
	}

	if fte.store != nil {
		fte.store.Close()
	}

	store, err := pgvector.New(
		context.Background(),
		pgvector.WithConnectionURL(fte.cfg.DatabaseURL),
		pgvector.WithEmbedder(embedder),
	)
	if err == nil {
		fte.store = &store
	}
}

func (fte *flowToolsExecutor) SetFunctions(functions *Functions) {
	fte.functions = functions
}

func (fte *flowToolsExecutor) SetScreenshotProvider(scp ScreenshotProvider) {
	fte.scp = scp
}

func (fte *flowToolsExecutor) SetAgentLogProvider(alp AgentLogProvider) {
	fte.alp = alp
}

func (fte *flowToolsExecutor) SetMsgLogProvider(mlp MsgLogProvider) {
	fte.mlp = mlp
}

func (fte *flowToolsExecutor) SetSearchLogProvider(slp SearchLogProvider) {
	fte.slp = slp
}

func (fte *flowToolsExecutor) SetTermLogProvider(tlp TermLogProvider) {
	fte.tlp = tlp
}

func (fte *flowToolsExecutor) SetVectorStoreLogProvider(vslp VectorStoreLogProvider) {
	fte.vslp = vslp
}

func (fte *flowToolsExecutor) SetGraphitiClient(client *graphiti.Client) {
	fte.graphitiClient = client
}

func (fte *flowToolsExecutor) Prepare(ctx context.Context) error {
	if cnt, err := fte.db.GetFlowPrimaryContainer(ctx, fte.flowID); err == nil {
		switch cnt.Status {
		case database.ContainerStatusRunning:
			fte.primaryID = cnt.ID
			fte.primaryLID = cnt.LocalID.String
			return nil
		default:
			fte.docker.DeleteContainer(ctx, cnt.LocalID.String, cnt.ID)
		}
	}

	capAdd := []string{"NET_RAW"}
	if fte.cfg.DockerNetAdmin {
		capAdd = append(capAdd, "NET_ADMIN")
	}

	containerName := PrimaryTerminalName(fte.flowID)
	cnt, err := fte.docker.SpawnContainer(
		ctx,
		containerName,
		database.ContainerTypePrimary,
		fte.flowID,
		&container.Config{
			Image:      fte.image,
			Entrypoint: []string{"tail", "-f", "/dev/null"},
		},
		&container.HostConfig{
			CapAdd: capAdd,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to spawn container '%s': %w", containerName, err)
	}

	fte.primaryID = cnt.ID
	fte.primaryLID = cnt.LocalID.String

	return nil
}

func (fte *flowToolsExecutor) Release(ctx context.Context) error {
	if fte.store != nil {
		fte.store.Close()
	}

	// TODO: here better to get flow containers list and delete all of them
	if err := fte.docker.DeleteContainer(ctx, fte.primaryLID, fte.primaryID); err != nil {
		containerName := PrimaryTerminalName(fte.flowID)
		return fmt.Errorf("failed to delete container '%s': %w", containerName, err)
	}

	return nil
}

func (fte *flowToolsExecutor) GetCustomExecutor(cfg CustomExecutorConfig) (ContextToolsExecutor, error) {
	if len(cfg.Definitions) != len(cfg.Handlers) {
		return nil, fmt.Errorf("definitions and handlers must have the same length")
	}

	for _, def := range cfg.Definitions {
		if _, ok := cfg.Handlers[def.Name]; !ok {
			return nil, fmt.Errorf("handler for function %s not found", def.Name)
		}
	}

	for _, builtin := range cfg.Builtin {
		if def, ok := fte.definitions[builtin]; !ok {
			return nil, fmt.Errorf("builtin function %s not found", builtin)
		} else {
			cfg.Definitions = append(cfg.Definitions, def)
			cfg.Handlers[builtin] = fte.handlers[builtin]
		}
	}

	barriers := make(map[string]struct{})
	for _, barrier := range cfg.Barriers {
		if _, ok := fte.handlers[barrier]; !ok {
			return nil, fmt.Errorf("barrier function %s not found", barrier)
		}
		barriers[barrier] = struct{}{}
	}

	return &customExecutor{
		flowID:      fte.flowID,
		taskID:      cfg.TaskID,
		subtaskID:   cfg.SubtaskID,
		mlp:         fte.mlp,
		vslp:        fte.vslp,
		db:          fte.db,
		store:       fte.store,
		definitions: cfg.Definitions,
		handlers:    cfg.Handlers,
		barriers:    barriers,
		summarizer:  cfg.Summarizer,
	}, nil
}

func (fte *flowToolsExecutor) GetAssistantExecutor(cfg AssistantExecutorConfig) (ContextToolsExecutor, error) {
	if cfg.Adviser == nil {
		return nil, fmt.Errorf("adviser handler is required")
	}

	if cfg.Coder == nil {
		return nil, fmt.Errorf("coder handler is required")
	}

	if cfg.Installer == nil {
		return nil, fmt.Errorf("installer handler is required")
	}

	if cfg.Memorist == nil {
		return nil, fmt.Errorf("memorist handler is required")
	}

	if cfg.Pentester == nil {
		return nil, fmt.Errorf("pentester handler is required")
	}

	if cfg.Searcher == nil {
		return nil, fmt.Errorf("searcher handler is required")
	}

	container, err := fte.db.GetFlowPrimaryContainer(context.Background(), fte.flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container %d: %w", fte.flowID, err)
	}

	term := &terminal{
		flowID:       fte.flowID,
		containerID:  container.ID,
		containerLID: container.LocalID.String,
		dockerClient: fte.docker,
		tlp:          fte.tlp,
	}

	definitions := []llms.FunctionDefinition{
		registryDefinitions[TerminalToolName],
		registryDefinitions[FileToolName],
	}
	handlers := map[string]ExecutorHandler{
		TerminalToolName: term.Handle,
		FileToolName:     term.Handle,
	}

	browser := &browser{
		flowID:   fte.flowID,
		dataDir:  fte.cfg.DataDir,
		scPrvURL: fte.cfg.ScraperPrivateURL,
		scPubURL: fte.cfg.ScraperPublicURL,
		scp:      fte.scp,
	}
	if browser.IsAvailable() {
		definitions = append(definitions, registryDefinitions[BrowserToolName])
		handlers[BrowserToolName] = browser.Handle
	}

	if cfg.UseAgents {
		definitions = append(definitions,
			registryDefinitions[AdviceToolName],
			registryDefinitions[CoderToolName],
			registryDefinitions[MaintenanceToolName],
			registryDefinitions[MemoristToolName],
			registryDefinitions[PentesterToolName],
			registryDefinitions[SearchToolName],
		)
		handlers[AdviceToolName] = cfg.Adviser
		handlers[CoderToolName] = cfg.Coder
		handlers[MaintenanceToolName] = cfg.Installer
		handlers[MemoristToolName] = cfg.Memorist
		handlers[PentesterToolName] = cfg.Pentester
		handlers[SearchToolName] = cfg.Searcher
	} else {
		memory := &memory{
			flowID: fte.flowID,
			store:  fte.store,
			vslp:   fte.vslp,
		}
		if memory.IsAvailable() {
			definitions = append(definitions, registryDefinitions[SearchInMemoryToolName])
			handlers[SearchInMemoryToolName] = memory.Handle
		}

		guide := &guide{
			flowID: fte.flowID,
			store:  fte.store,
			vslp:   fte.vslp,
		}
		if guide.IsAvailable() {
			definitions = append(definitions, registryDefinitions[SearchGuideToolName])
			handlers[SearchGuideToolName] = guide.Handle
		}

		search := &search{
			flowID: fte.flowID,
			store:  fte.store,
			vslp:   fte.vslp,
		}
		if search.IsAvailable() {
			definitions = append(definitions, registryDefinitions[SearchAnswerToolName])
			handlers[SearchAnswerToolName] = search.Handle
		}

		code := &code{
			flowID: fte.flowID,
			store:  fte.store,
			vslp:   fte.vslp,
		}
		if code.IsAvailable() {
			definitions = append(definitions, registryDefinitions[SearchCodeToolName])
			handlers[SearchCodeToolName] = code.Handle
		}

		google := &google{
			flowID:   fte.flowID,
			apiKey:   fte.cfg.GoogleAPIKey,
			cxKey:    fte.cfg.GoogleCXKey,
			lrKey:    fte.cfg.GoogleLRKey,
			proxyURL: fte.cfg.ProxyURL,
			slp:      fte.slp,
		}
		if google.IsAvailable() {
			definitions = append(definitions, registryDefinitions[GoogleToolName])
			handlers[GoogleToolName] = google.Handle
		}

		duckduckgo := &duckduckgo{
			flowID:   fte.flowID,
			enabled:  fte.cfg.DuckDuckGoEnabled,
			proxyURL: fte.cfg.ProxyURL,
			slp:      fte.slp,
		}
		if duckduckgo.IsAvailable() {
			definitions = append(definitions, registryDefinitions[DuckDuckGoToolName])
			handlers[DuckDuckGoToolName] = duckduckgo.Handle
		}

		tavily := &tavily{
			flowID:     fte.flowID,
			apiKey:     fte.cfg.TavilyAPIKey,
			proxyURL:   fte.cfg.ProxyURL,
			slp:        fte.slp,
			summarizer: cfg.Summarizer,
		}
		if tavily.IsAvailable() {
			definitions = append(definitions, registryDefinitions[TavilyToolName])
			handlers[TavilyToolName] = tavily.Handle
		}

		traversaal := &traversaal{
			flowID:   fte.flowID,
			apiKey:   fte.cfg.TraversaalAPIKey,
			proxyURL: fte.cfg.ProxyURL,
			slp:      fte.slp,
		}
		if traversaal.IsAvailable() {
			definitions = append(definitions, registryDefinitions[TraversaalToolName])
			handlers[TraversaalToolName] = traversaal.Handle
		}

		perplexity := &perplexity{
			flowID:      fte.flowID,
			apiKey:      fte.cfg.PerplexityAPIKey,
			proxyURL:    fte.cfg.ProxyURL,
			model:       fte.cfg.PerplexityModel,
			contextSize: fte.cfg.PerplexityContextSize,
			temperature: perplexityTemperature,
			topP:        perplexityTopP,
			maxTokens:   perplexityMaxTokens,
			timeout:     perplexityTimeout,
			slp:         fte.slp,
			summarizer:  cfg.Summarizer,
		}
		if perplexity.IsAvailable() {
			definitions = append(definitions, registryDefinitions[PerplexityToolName])
			handlers[PerplexityToolName] = perplexity.Handle
		}

		searxng := NewSearxngTool(
			fte.flowID,
			nil, // taskID
			nil, // subtaskID
			fte.cfg.SearxngURL,
			fte.cfg.SearxngCategories,
			fte.cfg.SearxngLanguage,
			fte.cfg.SearxngSafeSearch,
			fte.cfg.SearxngTimeRange,
			fte.cfg.ProxyURL,
			0, // timeout (will use default)
			fte.slp,
			cfg.Summarizer,
		)
		if searxng.IsAvailable() {
			definitions = append(definitions, registryDefinitions[SearxngToolName])
			handlers[SearxngToolName] = searxng.Handle
		}

		sploitus := NewSploitusTool(
			fte.flowID,
			nil, // taskID
			nil, // subtaskID
			fte.cfg.SploitusEnabled,
			fte.cfg.ProxyURL,
			fte.slp,
		)
		if sploitus.IsAvailable() {
			definitions = append(definitions, registryDefinitions[SploitusToolName])
			handlers[SploitusToolName] = sploitus.Handle
		}
	}

	ce := &customExecutor{
		flowID:      fte.flowID,
		mlp:         fte.mlp,
		vslp:        fte.vslp,
		db:          fte.db,
		store:       fte.store,
		definitions: definitions,
		handlers:    handlers,
		barriers:    map[string]struct{}{},
		summarizer:  cfg.Summarizer,
	}

	return ce, nil
}

func (fte *flowToolsExecutor) GetPrimaryExecutor(cfg PrimaryExecutorConfig) (ContextToolsExecutor, error) {
	if cfg.Barrier == nil {
		return nil, fmt.Errorf("barrier (done) handler is required")
	}

	if cfg.Adviser == nil {
		return nil, fmt.Errorf("adviser handler is required")
	}

	if cfg.Coder == nil {
		return nil, fmt.Errorf("coder handler is required")
	}

	if cfg.Installer == nil {
		return nil, fmt.Errorf("installer handler is required")
	}

	if cfg.Memorist == nil {
		return nil, fmt.Errorf("memorist handler is required")
	}

	if cfg.Pentester == nil {
		return nil, fmt.Errorf("pentester handler is required")
	}

	if cfg.Searcher == nil {
		return nil, fmt.Errorf("searcher handler is required")
	}

	ce := &customExecutor{
		flowID:    fte.flowID,
		taskID:    &cfg.TaskID,
		subtaskID: &cfg.SubtaskID,
		mlp:       fte.mlp,
		vslp:      fte.vslp,
		db:        fte.db,
		store:     fte.store,
		definitions: []llms.FunctionDefinition{
			registryDefinitions[FinalyToolName],
			registryDefinitions[AdviceToolName],
			registryDefinitions[CoderToolName],
			registryDefinitions[MaintenanceToolName],
			registryDefinitions[MemoristToolName],
			registryDefinitions[PentesterToolName],
			registryDefinitions[SearchToolName],
		},
		handlers: map[string]ExecutorHandler{
			FinalyToolName:      cfg.Barrier,
			AdviceToolName:      cfg.Adviser,
			CoderToolName:       cfg.Coder,
			MaintenanceToolName: cfg.Installer,
			MemoristToolName:    cfg.Memorist,
			PentesterToolName:   cfg.Pentester,
			SearchToolName:      cfg.Searcher,
		},
		barriers: map[string]struct{}{
			FinalyToolName: {},
		},
		summarizer: cfg.Summarizer,
	}

	if fte.cfg.AskUser {
		ce.definitions = append(ce.definitions, registryDefinitions[AskUserToolName])
		ce.handlers[AskUserToolName] = cfg.Barrier
		ce.barriers[AskUserToolName] = struct{}{}
	}

	return ce, nil
}

func (fte *flowToolsExecutor) GetInstallerExecutor(cfg InstallerExecutorConfig) (ContextToolsExecutor, error) {
	if cfg.MaintenanceResult == nil {
		return nil, fmt.Errorf("maintenance result handler is required")
	}

	if cfg.Adviser == nil {
		return nil, fmt.Errorf("adviser handler is required")
	}

	if cfg.Memorist == nil {
		return nil, fmt.Errorf("memorist handler is required")
	}

	if cfg.Searcher == nil {
		return nil, fmt.Errorf("searcher handler is required")
	}

	container, err := fte.db.GetFlowPrimaryContainer(context.Background(), fte.flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container %d: %w", fte.flowID, err)
	}

	term := &terminal{
		flowID:       fte.flowID,
		taskID:       cfg.TaskID,
		subtaskID:    cfg.SubtaskID,
		containerID:  container.ID,
		containerLID: container.LocalID.String,
		dockerClient: fte.docker,
		tlp:          fte.tlp,
	}

	ce := &customExecutor{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		mlp:       fte.mlp,
		vslp:      fte.vslp,
		db:        fte.db,
		store:     fte.store,
		definitions: []llms.FunctionDefinition{
			registryDefinitions[MaintenanceResultToolName],
			registryDefinitions[AdviceToolName],
			registryDefinitions[MemoristToolName],
			registryDefinitions[SearchToolName],
			registryDefinitions[TerminalToolName],
			registryDefinitions[FileToolName],
		},
		handlers: map[string]ExecutorHandler{
			MaintenanceResultToolName: cfg.MaintenanceResult,
			AdviceToolName:            cfg.Adviser,
			MemoristToolName:          cfg.Memorist,
			SearchToolName:            cfg.Searcher,
			TerminalToolName:          term.Handle,
			FileToolName:              term.Handle,
		},
		barriers: map[string]struct{}{
			MaintenanceResultToolName: {},
		},
		summarizer: cfg.Summarizer,
	}

	browser := &browser{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		dataDir:   fte.cfg.DataDir,
		scPrvURL:  fte.cfg.ScraperPrivateURL,
		scPubURL:  fte.cfg.ScraperPublicURL,
		scp:       fte.scp,
	}
	if browser.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[BrowserToolName])
		ce.handlers[BrowserToolName] = browser.Handle
	}

	guide := &guide{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		store:     fte.store,
		vslp:      fte.vslp,
	}
	if guide.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[StoreGuideToolName])
		ce.definitions = append(ce.definitions, registryDefinitions[SearchGuideToolName])
		ce.handlers[StoreGuideToolName] = guide.Handle
		ce.handlers[SearchGuideToolName] = guide.Handle
	}

	return ce, nil
}

func (fte *flowToolsExecutor) GetCoderExecutor(cfg CoderExecutorConfig) (ContextToolsExecutor, error) {
	if cfg.CodeResult == nil {
		return nil, fmt.Errorf("code result handler is required")
	}

	if cfg.Adviser == nil {
		return nil, fmt.Errorf("adviser handler is required")
	}

	if cfg.Installer == nil {
		return nil, fmt.Errorf("installer handler is required")
	}

	if cfg.Memorist == nil {
		return nil, fmt.Errorf("memorist handler is required")
	}

	if cfg.Searcher == nil {
		return nil, fmt.Errorf("searcher handler is required")
	}

	ce := &customExecutor{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		mlp:       fte.mlp,
		vslp:      fte.vslp,
		db:        fte.db,
		store:     fte.store,
		definitions: []llms.FunctionDefinition{
			registryDefinitions[CodeResultToolName],
			registryDefinitions[AdviceToolName],
			registryDefinitions[MaintenanceToolName],
			registryDefinitions[MemoristToolName],
			registryDefinitions[SearchToolName],
		},
		handlers: map[string]ExecutorHandler{
			CodeResultToolName:  cfg.CodeResult,
			AdviceToolName:      cfg.Adviser,
			MaintenanceToolName: cfg.Installer,
			MemoristToolName:    cfg.Memorist,
			SearchToolName:      cfg.Searcher,
		},
		barriers: map[string]struct{}{
			CodeResultToolName: {},
		},
		summarizer: cfg.Summarizer,
	}

	browser := &browser{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		dataDir:   fte.cfg.DataDir,
		scPrvURL:  fte.cfg.ScraperPrivateURL,
		scPubURL:  fte.cfg.ScraperPublicURL,
		scp:       fte.scp,
	}
	if browser.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[BrowserToolName])
		ce.handlers[BrowserToolName] = browser.Handle
	}

	code := &code{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		store:     fte.store,
		vslp:      fte.vslp,
	}
	if code.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[SearchCodeToolName])
		ce.definitions = append(ce.definitions, registryDefinitions[StoreCodeToolName])
		ce.handlers[SearchCodeToolName] = code.Handle
		ce.handlers[StoreCodeToolName] = code.Handle
	}

	graphitiSearch := NewGraphitiSearchTool(
		fte.flowID,
		cfg.TaskID,
		cfg.SubtaskID,
		fte.graphitiClient,
	)
	if graphitiSearch.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[GraphitiSearchToolName])
		ce.handlers[GraphitiSearchToolName] = graphitiSearch.Handle
	}

	return ce, nil
}

func (fte *flowToolsExecutor) GetPentesterExecutor(cfg PentesterExecutorConfig) (ContextToolsExecutor, error) {
	if cfg.HackResult == nil {
		return nil, fmt.Errorf("hack result handler is required")
	}

	if cfg.Adviser == nil {
		return nil, fmt.Errorf("adviser handler is required")
	}

	if cfg.Coder == nil {
		return nil, fmt.Errorf("coder handler is required")
	}

	if cfg.Installer == nil {
		return nil, fmt.Errorf("installer handler is required")
	}

	if cfg.Memorist == nil {
		return nil, fmt.Errorf("memorist handler is required")
	}

	if cfg.Searcher == nil {
		return nil, fmt.Errorf("searcher handler is required")
	}

	container, err := fte.db.GetFlowPrimaryContainer(context.Background(), fte.flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container %d: %w", fte.flowID, err)
	}

	term := &terminal{
		flowID:       fte.flowID,
		taskID:       cfg.TaskID,
		subtaskID:    cfg.SubtaskID,
		containerID:  container.ID,
		containerLID: container.LocalID.String,
		dockerClient: fte.docker,
		tlp:          fte.tlp,
	}

	ce := &customExecutor{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		mlp:       fte.mlp,
		vslp:      fte.vslp,
		db:        fte.db,
		store:     fte.store,
		definitions: []llms.FunctionDefinition{
			registryDefinitions[HackResultToolName],
			registryDefinitions[AdviceToolName],
			registryDefinitions[CoderToolName],
			registryDefinitions[MaintenanceToolName],
			registryDefinitions[MemoristToolName],
			registryDefinitions[SearchToolName],
			registryDefinitions[TerminalToolName],
			registryDefinitions[FileToolName],
		},
		handlers: map[string]ExecutorHandler{
			HackResultToolName:  cfg.HackResult,
			AdviceToolName:      cfg.Adviser,
			CoderToolName:       cfg.Coder,
			MaintenanceToolName: cfg.Installer,
			MemoristToolName:    cfg.Memorist,
			SearchToolName:      cfg.Searcher,
			TerminalToolName:    term.Handle,
			FileToolName:        term.Handle,
		},
		barriers: map[string]struct{}{
			HackResultToolName: {},
		},
		summarizer: cfg.Summarizer,
	}

	browser := &browser{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		dataDir:   fte.cfg.DataDir,
		scPrvURL:  fte.cfg.ScraperPrivateURL,
		scPubURL:  fte.cfg.ScraperPublicURL,
		scp:       fte.scp,
	}
	if browser.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[BrowserToolName])
		ce.handlers[BrowserToolName] = browser.Handle
	}

	guide := &guide{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		store:     fte.store,
		vslp:      fte.vslp,
	}
	if guide.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[StoreGuideToolName])
		ce.definitions = append(ce.definitions, registryDefinitions[SearchGuideToolName])
		ce.handlers[StoreGuideToolName] = guide.Handle
		ce.handlers[SearchGuideToolName] = guide.Handle
	}

	graphitiSearch := NewGraphitiSearchTool(
		fte.flowID,
		cfg.TaskID,
		cfg.SubtaskID,
		fte.graphitiClient,
	)
	if graphitiSearch.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[GraphitiSearchToolName])
		ce.handlers[GraphitiSearchToolName] = graphitiSearch.Handle
	}

	sploitus := NewSploitusTool(
		fte.flowID,
		cfg.TaskID,
		cfg.SubtaskID,
		fte.cfg.SploitusEnabled,
		fte.cfg.ProxyURL,
		fte.slp,
	)
	if sploitus.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[SploitusToolName])
		ce.handlers[SploitusToolName] = sploitus.Handle
	}

	return ce, nil
}

func (fte *flowToolsExecutor) GetSearcherExecutor(cfg SearcherExecutorConfig) (ContextToolsExecutor, error) {
	if cfg.SearchResult == nil {
		return nil, fmt.Errorf("search result handler is required")
	}

	if cfg.Memorist == nil {
		return nil, fmt.Errorf("memorist handler is required")
	}

	ce := &customExecutor{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		mlp:       fte.mlp,
		vslp:      fte.vslp,
		db:        fte.db,
		store:     fte.store,
		definitions: []llms.FunctionDefinition{
			registryDefinitions[SearchResultToolName],
			registryDefinitions[MemoristToolName],
		},
		handlers: map[string]ExecutorHandler{
			SearchResultToolName: cfg.SearchResult,
			MemoristToolName:     cfg.Memorist,
		},
		barriers: map[string]struct{}{
			SearchResultToolName: {},
		},
		summarizer: cfg.Summarizer,
	}

	browser := &browser{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		dataDir:   fte.cfg.DataDir,
		scPrvURL:  fte.cfg.ScraperPrivateURL,
		scPubURL:  fte.cfg.ScraperPublicURL,
		scp:       fte.scp,
	}
	if browser.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[BrowserToolName])
		ce.handlers[BrowserToolName] = browser.Handle
	}

	google := &google{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		apiKey:    fte.cfg.GoogleAPIKey,
		cxKey:     fte.cfg.GoogleCXKey,
		lrKey:     fte.cfg.GoogleLRKey,
		proxyURL:  fte.cfg.ProxyURL,
		slp:       fte.slp,
	}
	if google.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[GoogleToolName])
		ce.handlers[GoogleToolName] = google.Handle
	}

	duckduckgo := &duckduckgo{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		enabled:   fte.cfg.DuckDuckGoEnabled,
		proxyURL:  fte.cfg.ProxyURL,
		slp:       fte.slp,
	}
	if duckduckgo.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[DuckDuckGoToolName])
		ce.handlers[DuckDuckGoToolName] = duckduckgo.Handle
	}

	tavily := &tavily{
		flowID:     fte.flowID,
		taskID:     cfg.TaskID,
		subtaskID:  cfg.SubtaskID,
		apiKey:     fte.cfg.TavilyAPIKey,
		proxyURL:   fte.cfg.ProxyURL,
		slp:        fte.slp,
		summarizer: cfg.Summarizer,
	}
	if tavily.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[TavilyToolName])
		ce.handlers[TavilyToolName] = tavily.Handle
	}

	traversaal := &traversaal{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		apiKey:    fte.cfg.TraversaalAPIKey,
		proxyURL:  fte.cfg.ProxyURL,
		slp:       fte.slp,
	}
	if traversaal.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[TraversaalToolName])
		ce.handlers[TraversaalToolName] = traversaal.Handle
	}

	perplexity := &perplexity{
		flowID:      fte.flowID,
		taskID:      cfg.TaskID,
		subtaskID:   cfg.SubtaskID,
		apiKey:      fte.cfg.PerplexityAPIKey,
		proxyURL:    fte.cfg.ProxyURL,
		model:       fte.cfg.PerplexityModel,
		contextSize: fte.cfg.PerplexityContextSize,
		temperature: perplexityTemperature,
		topP:        perplexityTopP,
		maxTokens:   perplexityMaxTokens,
		timeout:     perplexityTimeout,
		slp:         fte.slp,
		summarizer:  cfg.Summarizer,
	}
	if perplexity.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[PerplexityToolName])
		ce.handlers[PerplexityToolName] = perplexity.Handle
	}

	searxng := NewSearxngTool(
		fte.flowID,
		cfg.TaskID,
		cfg.SubtaskID,
		fte.cfg.SearxngURL,
		fte.cfg.SearxngCategories,
		fte.cfg.SearxngLanguage,
		fte.cfg.SearxngSafeSearch,
		fte.cfg.SearxngTimeRange,
		fte.cfg.ProxyURL,
		0, // timeout (will use default)
		fte.slp,
		cfg.Summarizer,
	)
	if searxng.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[SearxngToolName])
		ce.handlers[SearxngToolName] = searxng.Handle
	}

	sploitus := NewSploitusTool(
		fte.flowID,
		cfg.TaskID,
		cfg.SubtaskID,
		fte.cfg.SploitusEnabled,
		fte.cfg.ProxyURL,
		fte.slp,
	)
	if sploitus.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[SploitusToolName])
		ce.handlers[SploitusToolName] = sploitus.Handle
	}

	search := &search{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		store:     fte.store,
		vslp:      fte.vslp,
	}
	if search.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[SearchAnswerToolName])
		ce.definitions = append(ce.definitions, registryDefinitions[StoreAnswerToolName])
		ce.handlers[SearchAnswerToolName] = search.Handle
		ce.handlers[StoreAnswerToolName] = search.Handle
	}

	return ce, nil
}

func (fte *flowToolsExecutor) GetGeneratorExecutor(cfg GeneratorExecutorConfig) (ContextToolsExecutor, error) {
	if cfg.SubtaskList == nil {
		return nil, fmt.Errorf("subtask list handler is required")
	}

	if cfg.Memorist == nil {
		return nil, fmt.Errorf("memorist handler is required")
	}

	container, err := fte.db.GetFlowPrimaryContainer(context.Background(), fte.flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container %d: %w", fte.flowID, err)
	}

	term := &terminal{
		flowID:       fte.flowID,
		taskID:       &cfg.TaskID,
		containerID:  container.ID,
		containerLID: container.LocalID.String,
		dockerClient: fte.docker,
		tlp:          fte.tlp,
	}

	ce := &customExecutor{
		flowID: fte.flowID,
		taskID: &cfg.TaskID,
		mlp:    fte.mlp,
		vslp:   fte.vslp,
		db:     fte.db,
		store:  fte.store,
		definitions: []llms.FunctionDefinition{
			registryDefinitions[MemoristToolName],
			registryDefinitions[SearchToolName],
			registryDefinitions[SubtaskListToolName],
			registryDefinitions[TerminalToolName],
			registryDefinitions[FileToolName],
		},
		handlers: map[string]ExecutorHandler{
			MemoristToolName:    cfg.Memorist,
			SearchToolName:      cfg.Searcher,
			SubtaskListToolName: cfg.SubtaskList,
			TerminalToolName:    term.Handle,
			FileToolName:        term.Handle,
		},
		barriers: map[string]struct{}{SubtaskListToolName: {}},
	}

	browser := &browser{
		flowID:   fte.flowID,
		taskID:   &cfg.TaskID,
		dataDir:  fte.cfg.DataDir,
		scPrvURL: fte.cfg.ScraperPrivateURL,
		scPubURL: fte.cfg.ScraperPublicURL,
		scp:      fte.scp,
	}
	if browser.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[BrowserToolName])
		ce.handlers[BrowserToolName] = browser.Handle
	}

	return ce, nil
}

func (fte *flowToolsExecutor) GetRefinerExecutor(cfg RefinerExecutorConfig) (ContextToolsExecutor, error) {
	if cfg.SubtaskPatch == nil {
		return nil, fmt.Errorf("subtask patch handler is required")
	}

	if cfg.Memorist == nil {
		return nil, fmt.Errorf("memorist handler is required")
	}

	container, err := fte.db.GetFlowPrimaryContainer(context.Background(), fte.flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container %d: %w", fte.flowID, err)
	}

	term := &terminal{
		flowID:       fte.flowID,
		taskID:       &cfg.TaskID,
		containerID:  container.ID,
		containerLID: container.LocalID.String,
		dockerClient: fte.docker,
		tlp:          fte.tlp,
	}

	ce := &customExecutor{
		flowID: fte.flowID,
		taskID: &cfg.TaskID,
		mlp:    fte.mlp,
		vslp:   fte.vslp,
		db:     fte.db,
		store:  fte.store,
		definitions: []llms.FunctionDefinition{
			registryDefinitions[MemoristToolName],
			registryDefinitions[SearchToolName],
			registryDefinitions[SubtaskPatchToolName],
			registryDefinitions[TerminalToolName],
			registryDefinitions[FileToolName],
		},
		handlers: map[string]ExecutorHandler{
			MemoristToolName:     cfg.Memorist,
			SearchToolName:       cfg.Searcher,
			SubtaskPatchToolName: cfg.SubtaskPatch,
			TerminalToolName:     term.Handle,
			FileToolName:         term.Handle,
		},
		barriers: map[string]struct{}{SubtaskPatchToolName: {}},
	}

	browser := &browser{
		flowID:   fte.flowID,
		taskID:   &cfg.TaskID,
		dataDir:  fte.cfg.DataDir,
		scPrvURL: fte.cfg.ScraperPrivateURL,
		scPubURL: fte.cfg.ScraperPublicURL,
		scp:      fte.scp,
	}
	if browser.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[BrowserToolName])
		ce.handlers[BrowserToolName] = browser.Handle
	}

	return ce, nil
}

func (fte *flowToolsExecutor) GetMemoristExecutor(cfg MemoristExecutorConfig) (ContextToolsExecutor, error) {
	if cfg.SearchResult == nil {
		return nil, fmt.Errorf("search result handler is required")
	}

	container, err := fte.db.GetFlowPrimaryContainer(context.Background(), fte.flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container %d: %w", fte.flowID, err)
	}

	term := &terminal{
		flowID:       fte.flowID,
		taskID:       cfg.TaskID,
		subtaskID:    cfg.SubtaskID,
		containerID:  container.ID,
		containerLID: container.LocalID.String,
		dockerClient: fte.docker,
		tlp:          fte.tlp,
	}

	ce := &customExecutor{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		mlp:       fte.mlp,
		vslp:      fte.vslp,
		db:        fte.db,
		store:     fte.store,
		definitions: []llms.FunctionDefinition{
			registryDefinitions[MemoristResultToolName],
			registryDefinitions[TerminalToolName],
			registryDefinitions[FileToolName],
		},
		handlers: map[string]ExecutorHandler{
			MemoristResultToolName: cfg.SearchResult,
			TerminalToolName:       term.Handle,
			FileToolName:           term.Handle,
		},
		barriers: map[string]struct{}{
			MemoristResultToolName: {},
		},
		summarizer: cfg.Summarizer,
	}

	memory := &memory{
		flowID: fte.flowID,
		store:  fte.store,
		vslp:   fte.vslp,
	}
	if memory.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[SearchInMemoryToolName])
		ce.handlers[SearchInMemoryToolName] = memory.Handle
	}

	graphitiSearch := NewGraphitiSearchTool(
		fte.flowID,
		cfg.TaskID,
		cfg.SubtaskID,
		fte.graphitiClient,
	)
	if graphitiSearch.IsAvailable() {
		ce.definitions = append(ce.definitions, registryDefinitions[GraphitiSearchToolName])
		ce.handlers[GraphitiSearchToolName] = graphitiSearch.Handle
	}

	return ce, nil
}

func (fte *flowToolsExecutor) GetEnricherExecutor(cfg EnricherExecutorConfig) (ContextToolsExecutor, error) {
	if cfg.EnricherResult == nil {
		return nil, fmt.Errorf("enricher result handler is required")
	}

	if cfg.Memorist == nil {
		return nil, fmt.Errorf("memorist handler is required")
	}

	if cfg.Searcher == nil {
		return nil, fmt.Errorf("searcher handler is required")
	}

	return &customExecutor{
		flowID:    fte.flowID,
		taskID:    cfg.TaskID,
		subtaskID: cfg.SubtaskID,
		mlp:       fte.mlp,
		vslp:      fte.vslp,
		db:        fte.db,
		store:     fte.store,
		definitions: []llms.FunctionDefinition{
			registryDefinitions[MemoristToolName],
			registryDefinitions[SearchToolName],
			registryDefinitions[EnricherResultToolName],
		},
		handlers: map[string]ExecutorHandler{
			MemoristToolName:       cfg.Memorist,
			SearchToolName:         cfg.Searcher,
			EnricherResultToolName: cfg.EnricherResult,
		},
		barriers: map[string]struct{}{EnricherResultToolName: {}},
	}, nil
}

func (fte *flowToolsExecutor) GetReporterExecutor(cfg ReporterExecutorConfig) (ContextToolsExecutor, error) {
	if cfg.ReportResult == nil {
		return nil, fmt.Errorf("report result handler is required")
	}

	return &customExecutor{
		flowID:      fte.flowID,
		taskID:      cfg.TaskID,
		subtaskID:   cfg.SubtaskID,
		mlp:         fte.mlp,
		vslp:        fte.vslp,
		db:          fte.db,
		store:       fte.store,
		definitions: []llms.FunctionDefinition{registryDefinitions[ReportResultToolName]},
		handlers:    map[string]ExecutorHandler{ReportResultToolName: cfg.ReportResult},
		barriers:    map[string]struct{}{ReportResultToolName: {}},
	}, nil
}

func enrichLogrusFields(flowID int64, taskID, subtaskID *int64, fields logrus.Fields) logrus.Fields {
	if fields == nil {
		fields = make(logrus.Fields)
	}

	fields["flow_id"] = flowID
	if taskID != nil {
		fields["task_id"] = *taskID
	}
	if subtaskID != nil {
		fields["subtask_id"] = *subtaskID
	}

	return fields
}
