package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"pentagi/cmd/ftester/mocks"
	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/graphiti"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/embeddings"
	"pentagi/pkg/terminal"
	"pentagi/pkg/tools"
	"pentagi/pkg/tools/searchers"

	"github.com/sirupsen/logrus"
	"github.com/vxcontrol/cloud/anonymizer"
	"github.com/vxcontrol/cloud/anonymizer/patterns"
	"github.com/vxcontrol/langchaingo/vectorstores/pgvector"
)

type agentTool struct {
	handler tools.ExecutorHandler
}

func (at *agentTool) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if at.handler == nil {
		return "", fmt.Errorf("handler for tool %s is not set", name)
	}
	return at.handler(ctx, name, args)
}

func (at *agentTool) IsAvailable() bool {
	return at.handler != nil
}

// searcherTool adapts a single searchers.Searcher to the tools.Tool interface so the
// ftester can exercise one engine against its real API by name (e.g. `ftester -flow N
// tavily`). Production agents never take this path — they call web_search, which
// orchestrates the engines — but per-engine access is valuable for debugging a single
// provider's live integration and response parsing in isolation.
type searcherTool struct {
	searcher searchers.Searcher
}

func (s *searcherTool) IsAvailable() bool { return s.searcher.IsAvailable() }

func (s *searcherTool) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	req, err := searchRequestFromArgs(name, args)
	if err != nil {
		return "", err
	}
	return s.searcher.Handle(ctx, req)
}

// searchRequestFromArgs builds the engine-agnostic searchers.Request from a single
// engine's LLM-facing arguments. Sploitus carries extra exploit fields; every other
// engine uses SearchAction.
func searchRequestFromArgs(name string, args json.RawMessage) (searchers.Request, error) {
	if name == tools.SploitusToolName {
		var a tools.SploitusAction
		if err := json.Unmarshal(args, &a); err != nil {
			return searchers.Request{}, fmt.Errorf("failed to unmarshal %s args: %w", name, err)
		}
		return searchers.Request{
			Query:       a.Query,
			MaxResults:  a.MaxResults.Int(),
			ExploitType: a.ExploitType.String(),
			Sort:        a.Sort.String(),
		}, nil
	}

	var a tools.SearchAction
	if err := json.Unmarshal(args, &a); err != nil {
		return searchers.Request{}, fmt.Errorf("failed to unmarshal %s args: %w", name, err)
	}
	return searchers.Request{Query: a.Query, MaxResults: a.MaxResults.Int()}, nil
}

// toolExecutor holds the necessary data for creating and managing tools
type toolExecutor struct {
	flowExecutor   tools.FlowToolsExecutor
	replacer       anonymizer.Replacer
	cfg            *config.Config
	db             database.Querier
	dockerClient   docker.DockerClient
	handlers       providers.FlowProviderHandlers
	store          *pgvector.Store
	embedder       embeddings.Embedder
	graphitiClient *graphiti.Client
	proxies        mocks.ProxyProviders
	userID         int64
	flowID         int64
	taskID         *int64
	subtaskID      *int64
}

// newToolExecutor creates a new executor with the given parameters
func newToolExecutor(
	flowExecutor tools.FlowToolsExecutor,
	cfg *config.Config,
	db database.Querier,
	dockerClient docker.DockerClient,
	handlers providers.FlowProviderHandlers,
	proxies mocks.ProxyProviders,
	userID, flowID int64,
	taskID, subtaskID *int64,
	embedder embeddings.Embedder,
	graphitiClient *graphiti.Client,
) (*toolExecutor, error) {
	var store *pgvector.Store
	if embedder.IsAvailable() {
		s, err := pgvector.New(
			context.Background(),
			pgvector.WithConnectionURL(cfg.DatabaseURL),
			pgvector.WithEmbedder(embedder),
			pgvector.WithCollectionName("langchain"),
		)
		if err != nil {
			logrus.WithError(err).Error("failed to create pgvector store")
		} else {
			store = &s
		}
	}

	allPatterns, err := patterns.LoadPatterns(patterns.PatternListTypeAll)
	if err != nil {
		return nil, fmt.Errorf("failed to load all patterns: %v", err)
	}

	// combine with config secret patterns
	allPatterns.Patterns = append(allPatterns.Patterns, cfg.GetSecretPatterns()...)

	replacer, err := anonymizer.NewReplacer(allPatterns.Regexes(), allPatterns.Names())
	if err != nil {
		return nil, fmt.Errorf("failed to create replacer: %v", err)
	}

	return &toolExecutor{
		flowExecutor:   flowExecutor,
		replacer:       replacer,
		cfg:            cfg,
		db:             db,
		dockerClient:   dockerClient,
		handlers:       handlers,
		store:          store,
		embedder:       embedder,
		graphitiClient: graphitiClient,
		proxies:        proxies,
		userID:         userID,
		flowID:         flowID,
		taskID:         taskID,
		subtaskID:      subtaskID,
	}, nil
}

// GetTool returns the appropriate tool for a given function name
func (te *toolExecutor) GetTool(ctx context.Context, funcName string) (tools.Tool, error) {
	// Get primary container for terminal/file operations (only when needed)
	var containerID int64
	var containerLID string

	requiresContainer := funcName == tools.TerminalToolName || funcName == tools.FileToolName
	if requiresContainer {
		cnt, err := te.db.GetFlowPrimaryContainer(ctx, te.flowID)
		if err != nil {
			return nil, fmt.Errorf("failed to get primary container for flow %d: %w", te.flowID, err)
		}
		containerID = cnt.ID
		containerLID = cnt.LocalID.String
	}

	// Check which tool to create based on function name
	switch funcName {
	case tools.TerminalToolName:
		return tools.NewTerminalTool(
			te.flowID,
			te.taskID,
			te.subtaskID,
			containerID,
			containerLID,
			te.cfg.TenantPrefix(),
			te.dockerClient,
			te.proxies.GetTermLogProvider(),
			time.Duration(te.cfg.TerminalToolTimeout)*time.Second,
		), nil

	case tools.FileToolName:
		// For file operations - uses the same terminal tool
		return tools.NewTerminalTool(
			te.flowID,
			te.taskID,
			te.subtaskID,
			containerID,
			containerLID,
			te.cfg.TenantPrefix(),
			te.dockerClient,
			te.proxies.GetTermLogProvider(),
			time.Duration(te.cfg.TerminalToolTimeout)*time.Second,
		), nil

	case tools.BrowserToolName:
		return tools.NewBrowserTool(
			te.flowID,
			te.taskID,
			te.subtaskID,
			te.cfg.DataDir,
			te.cfg.ScraperPrivateURL,
			te.cfg.ScraperPublicURL,
			te.proxies.GetScreenshotProvider(),
		), nil

	case tools.WebSearchToolName:
		// The unified web_search orchestrator, as production agents use it. A nil page
		// fetcher disables the optional internal analytics engine, which is out of
		// scope for the function tester.
		return tools.NewWebSearchTool(
			te.cfg,
			te.flowID,
			te.taskID,
			te.subtaskID,
			te.proxies.GetSearchLogProvider(),
			te.GetSummarizer(),
			nil,
		), nil

	// Individual engines: not used by agents (they call web_search), but exposed here
	// so a developer can test one provider's real API in isolation. Each searcher is
	// wrapped in searcherTool to bridge the searchers.Searcher → tools.Tool interface.
	case tools.GoogleToolName:
		return &searcherTool{searcher: searchers.NewGoogle(te.cfg)}, nil

	case tools.DuckDuckGoToolName:
		return &searcherTool{searcher: searchers.NewDuckDuckGo(te.cfg)}, nil

	case tools.TavilyToolName:
		return &searcherTool{searcher: searchers.NewTavily(te.cfg, searchers.SummarizeHandler(te.GetSummarizer()))}, nil

	case tools.FirecrawlToolName:
		return &searcherTool{searcher: searchers.NewFirecrawl(te.cfg, searchers.SummarizeHandler(te.GetSummarizer()))}, nil

	case tools.TraversaalToolName:
		return &searcherTool{searcher: searchers.NewTraversaal(te.cfg)}, nil

	case tools.PerplexityToolName:
		return &searcherTool{searcher: searchers.NewPerplexity(te.cfg, searchers.SummarizeHandler(te.GetSummarizer()))}, nil

	case tools.SearxngToolName:
		return &searcherTool{searcher: searchers.NewSearxng(te.cfg, searchers.SummarizeHandler(te.GetSummarizer()))}, nil

	case tools.SploitusToolName:
		return &searcherTool{searcher: searchers.NewSploitus(te.cfg)}, nil

	case "internal":
		// The internal browser-analytics engine — the opt-in alternative to Tavily.
		// Not an agent-facing tool (agents call web_search); exposed here so it can be
		// exercised in isolation. Requires WEB_SEARCH_INTERNAL_ENABLED=true, a configured
		// scraper (SCRAPER_PUBLIC_URL / SCRAPER_PRIVATE_URL), and at least one available
		// link engine (e.g. google/duckduckgo) — otherwise IsAvailable() is false.
		fetcher := tools.NewBrowserPageFetcher(te.flowID, te.taskID, te.subtaskID, te.cfg, te.proxies.GetScreenshotProvider())
		return &searcherTool{searcher: tools.NewInternalEngine(te.cfg, te.GetSummarizer(), fetcher)}, nil

	case tools.SearchInMemoryToolName:
		return tools.NewMemoryTool(
			te.flowID,
			te.store,
			te.proxies.GetVectorStoreLogProvider(),
		), nil

	case tools.SearchGuideToolName:
		return tools.NewGuideTool(
			te.userID,
			te.flowID,
			te.taskID,
			te.subtaskID,
			te.replacer,
			te.store,
			te.embedder,
			te.db,
			te.cfg.EmbeddingMaxTextBytes,
			te.proxies.GetVectorStoreLogProvider(),
			te.proxies.GetKnowledgeProvider(),
		), nil

	case tools.SearchAnswerToolName:
		return tools.NewSearchTool(
			te.userID,
			te.flowID,
			te.taskID,
			te.subtaskID,
			te.replacer,
			te.store,
			te.embedder,
			te.db,
			te.cfg.EmbeddingMaxTextBytes,
			te.proxies.GetVectorStoreLogProvider(),
			te.proxies.GetKnowledgeProvider(),
		), nil

	case tools.SearchCodeToolName:
		return tools.NewCodeTool(
			te.userID,
			te.flowID,
			te.taskID,
			te.subtaskID,
			te.replacer,
			te.store,
			te.embedder,
			te.db,
			te.cfg.EmbeddingMaxTextBytes,
			te.proxies.GetVectorStoreLogProvider(),
			te.proxies.GetKnowledgeProvider(),
		), nil

	case tools.GraphitiSearchToolName:
		return tools.NewGraphitiSearchTool(
			te.flowID,
			te.taskID,
			te.subtaskID,
			te.cfg.GroupID(te.flowID),
			te.graphitiClient,
		), nil

	// AI Agent tools
	case tools.AdviceToolName:
		var handler tools.ExecutorHandler
		if te.handlers != nil {
			if te.taskID != nil && te.subtaskID != nil {
				var err error
				handler, err = te.handlers.GetAskAdviceHandler(ctx, te.taskID, te.subtaskID)
				if err != nil {
					terminal.PrintWarning("Failed to get advice handler: %v", err)
				}
			} else {
				terminal.PrintWarning("No task or subtask ID provided for advice tool")
			}
		}
		return &agentTool{handler: handler}, nil

	case tools.CoderToolName:
		var handler tools.ExecutorHandler
		if te.handlers != nil {
			if te.taskID != nil && te.subtaskID != nil {
				var err error
				handler, err = te.handlers.GetCoderHandler(ctx, te.taskID, te.subtaskID)
				if err != nil {
					terminal.PrintWarning("Failed to get coder handler: %v", err)
				}
			} else {
				terminal.PrintWarning("No task or subtask ID provided for coder tool")
			}
		}
		return &agentTool{handler: handler}, nil

	case tools.MaintenanceToolName:
		var handler tools.ExecutorHandler
		if te.handlers != nil {
			if te.taskID != nil && te.subtaskID != nil {
				var err error
				handler, err = te.handlers.GetInstallerHandler(ctx, te.taskID, te.subtaskID)
				if err != nil {
					terminal.PrintWarning("Failed to get installer handler: %v", err)
				}
			} else {
				terminal.PrintWarning("No task or subtask ID provided for installer tool")
			}
		}
		return &agentTool{handler: handler}, nil

	case tools.MemoristToolName:
		var handler tools.ExecutorHandler
		if te.handlers != nil {
			if te.taskID != nil {
				var err error
				handler, err = te.handlers.GetMemoristHandler(ctx, te.taskID, te.subtaskID)
				if err != nil {
					terminal.PrintWarning("Failed to get memorist handler: %v", err)
				}
			} else {
				terminal.PrintWarning("No task ID provided for memorist tool")
			}
		}
		return &agentTool{handler: handler}, nil

	case tools.PentesterToolName:
		var handler tools.ExecutorHandler
		if te.handlers != nil {
			if te.taskID != nil && te.subtaskID != nil {
				var err error
				handler, err = te.handlers.GetPentesterHandler(ctx, te.taskID, te.subtaskID)
				if err != nil {
					terminal.PrintWarning("Failed to get pentester handler: %v", err)
				}
			} else {
				terminal.PrintWarning("No task or subtask ID provided for pentester tool")
			}
		}
		return &agentTool{handler: handler}, nil

	case tools.SearchToolName:
		var handler tools.ExecutorHandler
		if te.handlers != nil {
			var err error
			if te.taskID != nil && te.subtaskID != nil {
				// Use subtask specific searcher if both task and subtask IDs are available
				handler, err = te.handlers.GetSubtaskSearcherHandler(ctx, te.taskID, te.subtaskID)
			} else if te.taskID != nil {
				// Use task specific searcher if only task ID is available
				handler, err = te.handlers.GetTaskSearcherHandler(ctx, *te.taskID)
			} else {
				terminal.PrintWarning("No task or subtask ID provided for search tool")
			}
			if err != nil {
				terminal.PrintWarning("Failed to get search handler: %v", err)
			}
		}
		return &agentTool{handler: handler}, nil

	// For the rest of the functions, return TODO error for now
	default:
		return nil, fmt.Errorf("TODO: tool for function %s is not implemented yet", funcName)
	}
}

// ExecuteFunctionWrapper executes a function, choosing between mock or real execution
func (te *toolExecutor) ExecuteFunctionWrapper(ctx context.Context, funcName string, args json.RawMessage) (string, error) {
	// If flowID = 0, use mock responses
	if te.flowID == 0 {
		terminal.PrintInfo("Using MOCK mode (flowID=0)")
		return mocks.MockResponse(funcName, args)
	}

	// If flowID > 0, perform real function execution
	terminal.PrintInfo("Using REAL mode (flowID>0)")
	return te.ExecuteRealFunction(ctx, funcName, args)
}

// ExecuteRealFunction performs the real function using the executor
func (te *toolExecutor) ExecuteRealFunction(ctx context.Context, funcName string, args json.RawMessage) (string, error) {
	// Execute the function
	terminal.PrintInfo("Executing real function: %s", funcName)

	// Get the appropriate tool for this function
	tool, err := te.GetTool(ctx, funcName)
	if err != nil {
		return "", fmt.Errorf("error getting tool for function %s: %w", funcName, err)
	}

	// Check if the tool is available
	if !tool.IsAvailable() {
		return "", fmt.Errorf("tool for function %s is not available", funcName)
	}

	// Handle the function with the tool
	return tool.Handle(ctx, funcName, args)
}

// ExecuteFunctionWithMode handles the general function call and displays the result
func (te *toolExecutor) ExecuteFunctionWithMode(ctx context.Context, funcName string, args any) error {
	// Marshal arguments to JSON
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("error marshaling arguments: %w", err)
	}

	// Nicely print function information
	terminal.PrintHeader("Executing function: " + funcName)
	terminal.PrintHeader("Arguments:")
	terminal.PrintJSON(args)

	// Execute the function (either in mock mode or real)
	result, err := te.ExecuteFunctionWrapper(ctx, funcName, argsJSON)
	if err != nil {
		return fmt.Errorf("error executing function: %w", err)
	}

	// Nicely print the result
	terminal.PrintHeader("\nResult:")
	var resultObj any
	if err := json.Unmarshal([]byte(result), &resultObj); err != nil {
		// If the result is not JSON, check if it's markdown and render appropriately
		terminal.PrintResult(result)
	} else {
		terminal.PrintJSON(resultObj)
	}

	terminal.PrintSuccess("\nExecution completed successfully.")
	return nil
}

func (te *toolExecutor) GetSummarizer() tools.SummarizeHandler {
	if te.handlers == nil {
		return nil
	}

	return te.handlers.GetSummarizeResultHandler(te.taskID, te.subtaskID)
}
