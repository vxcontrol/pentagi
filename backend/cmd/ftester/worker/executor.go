package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"pentagi/cmd/ftester/mocks"
	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/graphiti"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/embeddings"
	"pentagi/pkg/terminal"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
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

// toolExecutor holds the necessary data for creating and managing tools
type toolExecutor struct {
	flowExecutor   tools.FlowToolsExecutor
	cfg            *config.Config
	db             database.Querier
	dockerClient   docker.DockerClient
	handlers       providers.FlowProviderHandlers
	store          *pgvector.Store
	graphitiClient *graphiti.Client
	proxies        mocks.ProxyProviders
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
	flowID int64,
	taskID, subtaskID *int64,
	embedder embeddings.Embedder,
	graphitiClient *graphiti.Client,
) *toolExecutor {
	var store *pgvector.Store
	if embedder.IsAvailable() {
		s, err := pgvector.New(
			context.Background(),
			pgvector.WithConnectionURL(cfg.DatabaseURL),
			pgvector.WithEmbedder(embedder),
		)
		if err != nil {
			logrus.WithError(err).Error("failed to create pgvector store")
		} else {
			store = &s
		}
	}

	return &toolExecutor{
		flowExecutor:   flowExecutor,
		cfg:            cfg,
		db:             db,
		dockerClient:   dockerClient,
		handlers:       handlers,
		store:          store,
		graphitiClient: graphitiClient,
		proxies:        proxies,
		flowID:         flowID,
		taskID:         taskID,
		subtaskID:      subtaskID,
	}
}

// GetTool returns the appropriate tool for a given function name
func (te *toolExecutor) GetTool(ctx context.Context, funcName string) (tools.Tool, error) {
	// Get primary container for terminal/file operations
	var containerID int64
	var containerLID string
	if cnt, err := te.db.GetFlowPrimaryContainer(ctx, te.flowID); err == nil {
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
			te.dockerClient,
			te.proxies.GetTermLogProvider(),
		), nil

	case tools.FileToolName:
		// For file operations - uses the same terminal tool
		return tools.NewTerminalTool(
			te.flowID,
			te.taskID,
			te.subtaskID,
			containerID,
			containerLID,
			te.dockerClient,
			te.proxies.GetTermLogProvider(),
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

	case tools.GoogleToolName:
		return tools.NewGoogleTool(
			te.flowID,
			te.taskID,
			te.subtaskID,
			te.cfg.GoogleAPIKey,
			te.cfg.GoogleCXKey,
			te.cfg.GoogleLRKey,
			te.cfg.ProxyURL,
			te.proxies.GetSearchLogProvider(),
		), nil

	case tools.DuckDuckGoToolName:
		return tools.NewDuckDuckGoTool(
			te.flowID,
			te.taskID,
			te.subtaskID,
			te.cfg.DuckDuckGoEnabled,
			te.cfg.ProxyURL,
			"", // region (default)
			"", // safeSearch (default)
			"", // timeRange (default)
			te.proxies.GetSearchLogProvider(),
		), nil

	case tools.TavilyToolName:
		return tools.NewTavilyTool(
			te.flowID,
			te.taskID,
			te.subtaskID,
			te.cfg.TavilyAPIKey,
			te.cfg.ProxyURL,
			te.proxies.GetSearchLogProvider(),
			te.GetSummarizer(),
		), nil

	case tools.TraversaalToolName:
		return tools.NewTraversaalTool(
			te.flowID,
			te.taskID,
			te.subtaskID,
			te.cfg.TraversaalAPIKey,
			te.cfg.ProxyURL,
			te.proxies.GetSearchLogProvider(),
		), nil

	case tools.PerplexityToolName:
		return tools.NewPerplexityTool(
			te.flowID,
			te.taskID,
			te.subtaskID,
			te.cfg.PerplexityAPIKey,
			te.cfg.ProxyURL,
			te.cfg.PerplexityModel,
			te.cfg.PerplexityContextSize,
			0, // default temperature
			0, // default topP
			0, // default maxTokens
			0, // default timeout
			te.proxies.GetSearchLogProvider(),
			te.GetSummarizer(),
		), nil

	case tools.SearxngToolName:
		return tools.NewSearxngTool(
			te.flowID,
			te.taskID,
			te.subtaskID,
			te.cfg.SearxngURL,
			te.cfg.SearxngCategories,
			te.cfg.SearxngLanguage,
			te.cfg.SearxngSafeSearch,
			te.cfg.SearxngTimeRange,
			te.cfg.ProxyURL,
			0, // timeout (will use default)
			te.proxies.GetSearchLogProvider(),
			te.GetSummarizer(),
		), nil

	case tools.SploitusToolName:
		return tools.NewSploitusTool(
			te.flowID,
			te.taskID,
			te.subtaskID,
			te.cfg.SploitusEnabled,
			te.cfg.ProxyURL,
			te.proxies.GetSearchLogProvider(),
		), nil

	case tools.SearchInMemoryToolName:
		return tools.NewMemoryTool(
			te.flowID,
			te.store,
			te.proxies.GetVectorStoreLogProvider(),
		), nil

	case tools.SearchGuideToolName:
		return tools.NewGuideTool(
			te.flowID,
			te.taskID,
			te.subtaskID,
			te.store,
			te.proxies.GetVectorStoreLogProvider(),
		), nil

	case tools.SearchAnswerToolName:
		return tools.NewSearchTool(
			te.flowID,
			te.taskID,
			te.subtaskID,
			te.store,
			te.proxies.GetVectorStoreLogProvider(),
		), nil

	case tools.SearchCodeToolName:
		return tools.NewCodeTool(
			te.flowID,
			te.taskID,
			te.subtaskID,
			te.store,
			te.proxies.GetVectorStoreLogProvider(),
		), nil

	case tools.GraphitiSearchToolName:
		return tools.NewGraphitiSearchTool(
			te.flowID,
			te.taskID,
			te.subtaskID,
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
