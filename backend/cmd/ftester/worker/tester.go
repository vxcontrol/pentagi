package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"pentagi/cmd/ftester/mocks"
	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/templates"
	"pentagi/pkg/terminal"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
)

type Tester interface {
	Execute(args []string) error
}

// tester represents the main testing utility for tools functions
type tester struct {
	db           database.Querier
	cfg          *config.Config
	ctx          context.Context
	docker       docker.DockerClient
	providers    providers.ProviderController
	providerName provider.ProviderName
	providerType provider.ProviderType
	userID       int64
	flowID       int64
	taskID       *int64
	subtaskID    *int64
	provider     provider.Provider
	toolExecutor *toolExecutor
	flowExecutor tools.FlowToolsExecutor
	flowProvider providers.FlowProvider
	proxies      mocks.ProxyProviders
	functions    *tools.Functions
}

// NewTester creates a new instance of the tester with all necessary components
func NewTester(
	db database.Querier,
	cfg *config.Config,
	ctx context.Context,
	dockerClient docker.DockerClient,
	providerController providers.ProviderController,
	flowID, userID int64,
	taskID, subtaskID *int64,
	prvname provider.ProviderName,
) (Tester, error) {
	// New provider by user
	prv, err := providerController.GetProvider(ctx, prvname, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Create empty functions definition
	functions := &tools.Functions{}

	// Initialize tools flowExecutor
	flowExecutor, err := tools.NewFlowToolsExecutor(db, cfg, dockerClient, functions, flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to create flow tools executor: %w", err)
	}

	// Initialize proxy providers
	proxies := mocks.NewProxyProviders()

	// Set proxy providers to the executor
	flowExecutor.SetScreenshotProvider(proxies.GetScreenshotProvider())
	flowExecutor.SetAgentLogProvider(proxies.GetAgentLogProvider())
	flowExecutor.SetMsgLogProvider(proxies.GetMsgLogProvider())
	flowExecutor.SetSearchLogProvider(proxies.GetSearchLogProvider())
	flowExecutor.SetTermLogProvider(proxies.GetTermLogProvider())
	flowExecutor.SetVectorStoreLogProvider(proxies.GetVectorStoreLogProvider())
	flowExecutor.SetGraphitiClient(providerController.GraphitiClient())

	// Initialize tool executor
	toolExecutor := newToolExecutor(
		flowExecutor, cfg, db, dockerClient, nil, proxies,
		flowID, taskID, subtaskID, providerController.Embedder(),
		providerController.GraphitiClient(),
	)

	t := &tester{
		db:           db,
		cfg:          cfg,
		ctx:          ctx,
		docker:       dockerClient,
		providers:    providerController,
		providerName: prvname,
		providerType: prv.Type(),
		userID:       userID,
		flowID:       flowID,
		taskID:       taskID,
		subtaskID:    subtaskID,
		provider:     prv,
		toolExecutor: toolExecutor,
		flowExecutor: flowExecutor,
		proxies:      proxies,
		functions:    functions,
	}
	if err := t.initFlowProviderController(); err != nil {
		return nil, fmt.Errorf("failed to initialize flow provider controller: %w", err)
	}

	return t, nil
}

// initFlowProviderController initializes the flow provider when flowID is set
func (t *tester) initFlowProviderController() error {
	// When flowID=0, we're in mock mode and don't need real container or provider
	// This allows testing tools functions without a running flow
	if t.flowID == 0 {
		return nil
	}

	flow, err := t.db.GetFlow(t.ctx, t.flowID)
	if err != nil {
		return fmt.Errorf("failed to get flow: %w", err)
	}

	container, err := t.db.GetFlowPrimaryContainer(t.ctx, flow.ID)
	if err != nil {
		return fmt.Errorf("failed to get flow primary container: %w", err)
	}

	user, err := t.db.GetUser(t.ctx, flow.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user %d: %w", flow.UserID, err)
	}

	// Setup Langfuse observability to track the execution lifecycle
	// This is critical for debugging and monitoring flow performance
	// We use trace context to connect this execution with earlier/later runs
	ctx, observation := obs.Observer.NewObservation(t.ctx,
		langfuse.WithObservationTraceID(flow.TraceID.String),
		langfuse.WithObservationTraceContext(
			langfuse.WithTraceName(fmt.Sprintf("%d flow worker", flow.ID)),
			langfuse.WithTraceUserId(user.Mail),
			langfuse.WithTraceTags([]string{"controller"}),
			langfuse.WithTraceSessionId(fmt.Sprintf("flow-%d", flow.ID)),
			langfuse.WithTraceMetadata(langfuse.Metadata{
				"flow_id":       flow.ID,
				"user_id":       flow.UserID,
				"user_email":    user.Mail,
				"user_name":     user.Name,
				"user_hash":     user.Hash,
				"user_role":     user.RoleName,
				"provider_name": flow.ModelProviderName,
				"provider_type": flow.ModelProviderType,
			}),
		),
	)

	// Create a span for tracking the entire worker lifecycle
	flowSpan := observation.Span(langfuse.WithStartSpanName("run tester flow worker"))
	t.ctx, _ = flowSpan.Observation(ctx)

	// Each flow has its own JSON configuration of allowed functions
	// These determine what tools the AI can access during execution
	functions := &tools.Functions{}
	if err := json.Unmarshal(flow.Functions, functions); err != nil {
		return wrapErrorEndSpan(t.ctx, flowSpan, "failed to unmarshal functions", err)
	}
	t.flowExecutor.SetFunctions(functions)

	// Create a prompter for communicating with the AI model
	// TODO: This will eventually be customized per user/flow
	prompter := templates.NewDefaultPrompter() // TODO: change to flow prompter by userID from DB

	// The flow provider is the bridge between the AI model and the tools executor
	// It determines which AI service (OpenAI, Claude, etc) will be used and how
	// the instructions are formatted and interpreted
	flowProvider, err := t.providers.LoadFlowProvider(
		t.ctx,
		t.providerName,
		prompter,
		t.flowExecutor,
		t.flowID,
		t.userID,
		container.Image,
		flow.Language,
		flow.Title,
	)
	if err != nil {
		return wrapErrorEndSpan(t.ctx, flowSpan, "failed to load flow provider", err)
	}

	// Connect the provider's image and embedding model to the executor
	// This ensures we use the right container and vector DB configuration
	t.flowExecutor.SetImage(flowProvider.Image())
	t.flowExecutor.SetEmbedder(flowProvider.Embedder())

	// Setup log capturing for later inspection and debugging
	flowProvider.SetAgentLogProvider(t.proxies.GetAgentLogProvider())
	flowProvider.SetMsgLogProvider(t.proxies.GetMsgLogProvider())

	// Store references to complete the initialization chain
	t.flowProvider = flowProvider
	t.toolExecutor.handlers = flowProvider

	return nil
}

// Execute processes command line arguments and runs the appropriate function
func (t *tester) Execute(args []string) error {
	// If no args or first arg is '-help' or no args after flags processing, show general help
	if len(args) == 0 || args[0] == "-help" || args[0] == "--help" {
		return t.showGeneralHelp()
	}

	funcName := args[0]

	if len(args) > 1 && (args[1] == "-help" || args[1] == "--help") {
		// Show function-specific help
		return t.showFunctionHelp(funcName)
	}

	var funcArgs any
	var err error

	// Handle the describe function
	if funcName == "describe" {
		// If no arguments are provided, use interactive mode
		if len(args) == 1 {
			terminal.PrintInfo("No arguments provided, using interactive mode")
			funcArgs, err = InteractiveFillArgs(t.ctx, funcName, t.taskID, t.subtaskID)
		} else {
			// Parse describe function arguments
			funcArgs, err = ParseFunctionArgs(funcName, args[1:])
		}

		if err != nil {
			return fmt.Errorf("error parsing arguments: %w", err)
		}

		// Call the describe function
		return t.executeDescribe(t.ctx, funcArgs.(*DescribeParams))
	}

	// Check if arguments are provided
	if len(args) == 1 {
		terminal.PrintInfo("No arguments provided, using interactive mode")
		funcArgs, err = InteractiveFillArgs(t.ctx, funcName, t.taskID, t.subtaskID)
	} else {
		// Parse function arguments
		funcArgs, err = ParseFunctionArgs(funcName, args[1:])
	}

	if err != nil {
		return fmt.Errorf("error parsing arguments: %w", err)
	}

	// If flowID > 0 and the function requires terminal preparation, prepare it
	if t.flowID > 0 && t.needsTeminalPrepare(funcName) {
		terminal.PrintInfo("Preparing container for terminal operations...")
		if err := t.flowExecutor.Prepare(t.ctx); err != nil {
			return fmt.Errorf("failed to prepare executor: %w", err)
		}
		defer func() {
			if err := t.flowExecutor.Release(t.ctx); err != nil {
				terminal.PrintWarning("Failed to release executor: %v", err)
			}
		}()
	}

	// Execute the function with appropriate mode based on flowID
	return t.toolExecutor.ExecuteFunctionWithMode(t.ctx, funcName, funcArgs)
}

// executeDescribe shows information about tasks and subtasks for the current flow
func (t *tester) executeDescribe(ctx context.Context, params *DescribeParams) error {
	// If flowID is 0, show list of all flows
	if t.flowID == 0 {
		return t.executeDescribeFlows(ctx, params)
	}

	// If subtask_id is specified, only show that specific subtask
	if t.subtaskID != nil {
		return t.executeDescribeSubtask(ctx, params)
	}

	// If task_id is specified, show only that task and its subtasks
	if t.taskID != nil {
		return t.executeDescribeTask(ctx, params)
	}

	// Show flow info and all tasks and subtasks for this flow
	return t.executeDescribeFlowTasks(ctx, params)
}

// executeDescribeFlows shows list of all flows in the system
func (t *tester) executeDescribeFlows(ctx context.Context, params *DescribeParams) error {
	// Get all flows
	flows, err := t.db.GetFlows(ctx)
	if err != nil {
		return fmt.Errorf("failed to get flows: %w", err)
	}

	if len(flows) == 0 {
		terminal.PrintInfo("No flows found")
		return nil
	}

	terminal.PrintHeader("Available Flows:")
	terminal.PrintThickSeparator()
	for _, flow := range flows {
		// Always display basic info
		terminal.PrintKeyValue("Flow ID", fmt.Sprintf("%d", flow.ID))
		terminal.PrintKeyValue("Title", flow.Title)
		terminal.PrintKeyValue("Status", string(flow.Status))
		if flow.CreatedAt.Valid {
			terminal.PrintKeyValue("Created At", flow.CreatedAt.Time.Format("2006-01-02 15:04:05"))
		}

		// Display additional info if verbose mode is enabled
		if params.Verbose {
			terminal.PrintKeyValue("Model", flow.Model)
			terminal.PrintKeyValue("ProviderName", flow.ModelProviderName)
			terminal.PrintKeyValue("ProviderType", string(flow.ModelProviderType))
			terminal.PrintKeyValue("Language", flow.Language)

			// Get user info who created this flow
			if user, err := t.db.GetUser(ctx, flow.UserID); err == nil {
				terminal.PrintKeyValue("User", fmt.Sprintf("%s (%s)", user.Name, user.Mail))
				terminal.PrintKeyValue("User Role", user.RoleName)
			}
		}
		terminal.PrintThickSeparator()
	}

	return nil
}

// executeDescribeSubtask shows information about a specific subtask
func (t *tester) executeDescribeSubtask(ctx context.Context, params *DescribeParams) error {
	subtask, err := t.db.GetSubtask(ctx, *t.subtaskID)
	if err != nil {
		return fmt.Errorf("failed to get subtask: %w", err)
	}

	task, err := t.db.GetTask(ctx, subtask.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get parent task: %w", err)
	}

	if task.FlowID != t.flowID {
		return fmt.Errorf("subtask %d does not belong to flow %d", *t.subtaskID, t.flowID)
	}

	// Get flow information
	flow, err := t.db.GetFlow(ctx, t.flowID)
	if err != nil {
		return fmt.Errorf("failed to get flow information: %w", err)
	}

	// Display flow info
	terminal.PrintHeader("Flow Information")
	terminal.PrintKeyValue("Flow ID", fmt.Sprintf("%d", flow.ID))
	terminal.PrintKeyValue("Title", flow.Title)
	terminal.PrintKeyValue("Status", string(flow.Status))
	fmt.Println()

	// Display task info
	terminal.PrintHeader("Task Information")
	terminal.PrintKeyValue("Task ID", fmt.Sprintf("%d", task.ID))
	terminal.PrintKeyValue("Task Title", task.Title)
	terminal.PrintKeyValue("Task Status", string(task.Status))
	if params.Verbose {
		terminal.PrintThinSeparator()
		terminal.PrintHeader("Task Input")
		terminal.RenderMarkdown(task.Input)
		terminal.PrintThinSeparator()
		terminal.PrintHeader("Task Result")
		terminal.RenderMarkdown(task.Result)
	}
	fmt.Println()

	// Print subtask details
	terminal.PrintHeader("Subtask Information")
	terminal.PrintKeyValue("Subtask ID", fmt.Sprintf("%d", subtask.ID))
	terminal.PrintKeyValue("Subtask Title", subtask.Title)
	terminal.PrintKeyValue("Subtask Status", string(subtask.Status))
	if params.Verbose {
		terminal.PrintThinSeparator()
		terminal.PrintHeader("Subtask Description")
		terminal.RenderMarkdown(subtask.Description)
		terminal.PrintThinSeparator()
		terminal.PrintHeader("Subtask Result")
		terminal.RenderMarkdown(subtask.Result)
	}
	return nil
}

// executeDescribeTask shows information about a specific task and its subtasks
func (t *tester) executeDescribeTask(ctx context.Context, params *DescribeParams) error {
	task, err := t.db.GetFlowTask(ctx, database.GetFlowTaskParams{
		ID:     *t.taskID,
		FlowID: t.flowID,
	})
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get flow information
	flow, err := t.db.GetFlow(ctx, t.flowID)
	if err != nil {
		return fmt.Errorf("failed to get flow information: %w", err)
	}

	// Display flow info
	terminal.PrintHeader("Flow Information")
	terminal.PrintKeyValue("Flow ID", fmt.Sprintf("%d", flow.ID))
	terminal.PrintKeyValue("Title", flow.Title)
	terminal.PrintKeyValue("Status", string(flow.Status))
	fmt.Println()

	// Display task info
	terminal.PrintHeader("Task Information")
	terminal.PrintKeyValue("Task ID", fmt.Sprintf("%d", task.ID))
	terminal.PrintKeyValue("Task Title", task.Title)
	terminal.PrintKeyValue("Task Status", string(task.Status))
	if params.Verbose {
		terminal.PrintThinSeparator()
		terminal.PrintHeader("Task Input")
		terminal.RenderMarkdown(task.Input)
		terminal.PrintThinSeparator()
		terminal.PrintHeader("Task Result")
		terminal.RenderMarkdown(task.Result)
	}
	fmt.Println()

	// Get subtasks for this task
	subtasks, err := t.db.GetFlowTaskSubtasks(ctx, database.GetFlowTaskSubtasksParams{
		FlowID: t.flowID,
		TaskID: *t.taskID,
	})
	if err != nil {
		return fmt.Errorf("failed to get subtasks: %w", err)
	}

	if len(subtasks) == 0 {
		terminal.PrintInfo("No subtasks found for this task")
		return nil
	}

	terminal.PrintHeader(fmt.Sprintf("Subtasks for Task %d:", task.ID))
	terminal.PrintThinSeparator()
	for _, subtask := range subtasks {
		terminal.PrintKeyValue("Subtask ID", fmt.Sprintf("%d", subtask.ID))
		terminal.PrintKeyValue("Subtask Title", subtask.Title)
		terminal.PrintKeyValue("Subtask Status", string(subtask.Status))
		if params.Verbose {
			terminal.PrintThinSeparator()
			terminal.PrintHeader("Subtask Description")
			terminal.RenderMarkdown(subtask.Description)
			terminal.PrintThinSeparator()
			terminal.PrintHeader("Subtask Result")
			terminal.RenderMarkdown(subtask.Result)
		}
		terminal.PrintThinSeparator()
	}
	return nil
}

// executeDescribeFlowTasks shows information about a flow and all its tasks and subtasks
func (t *tester) executeDescribeFlowTasks(ctx context.Context, params *DescribeParams) error {
	// Get flow information
	flow, err := t.db.GetFlow(ctx, t.flowID)
	if err != nil {
		return fmt.Errorf("failed to get flow information: %w", err)
	}

	terminal.PrintHeader("Flow Information")
	terminal.PrintKeyValue("Flow ID", fmt.Sprintf("%d", flow.ID))
	terminal.PrintKeyValue("Title", flow.Title)
	terminal.PrintKeyValue("Status", string(flow.Status))
	terminal.PrintKeyValue("Language", flow.Language)
	terminal.PrintKeyValue("Model", fmt.Sprintf("%s (%s)", flow.Model, flow.ModelProviderName))
	if flow.CreatedAt.Valid {
		terminal.PrintKeyValue("Created At", flow.CreatedAt.Time.Format("2006-01-02 15:04:05"))
	}
	fmt.Println()

	// Show all tasks and subtasks for this flow
	tasks, err := t.db.GetFlowTasks(ctx, t.flowID)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	if len(tasks) == 0 {
		terminal.PrintInfo("No tasks found for this flow")
		return nil
	}

	terminal.PrintHeader(fmt.Sprintf("Tasks for Flow %d:", t.flowID))
	terminal.PrintThickSeparator()
	for _, task := range tasks {
		terminal.PrintKeyValue("Task ID", fmt.Sprintf("%d", task.ID))
		terminal.PrintKeyValue("Task Title", task.Title)
		terminal.PrintKeyValue("Task Status", string(task.Status))
		if params.Verbose {
			terminal.PrintThinSeparator()
			terminal.PrintHeader("Task Input")
			terminal.RenderMarkdown(task.Input)
			terminal.PrintThinSeparator()
			terminal.PrintHeader("Task Result")
			terminal.RenderMarkdown(task.Result)
		}
		fmt.Println()

		// Get subtasks for this task
		subtasks, err := t.db.GetTaskSubtasks(ctx, task.ID)
		if err != nil {
			return fmt.Errorf("failed to get subtasks for task %d: %w", task.ID, err)
		}

		if len(subtasks) > 0 {
			terminal.PrintHeader(fmt.Sprintf("Subtasks for Task %d:", task.ID))
			terminal.PrintThinSeparator()
			for _, subtask := range subtasks {
				terminal.PrintKeyValue("Subtask ID", fmt.Sprintf("%d", subtask.ID))
				terminal.PrintKeyValue("Subtask Title", subtask.Title)
				terminal.PrintKeyValue("Subtask Status", string(subtask.Status))
				if params.Verbose {
					terminal.PrintThinSeparator()
					terminal.PrintHeader("Subtask Description")
					terminal.RenderMarkdown(subtask.Description)
					terminal.PrintThinSeparator()
					terminal.PrintHeader("Subtask Result")
					terminal.RenderMarkdown(subtask.Result)
				}
				terminal.PrintThinSeparator()
			}
		} else {
			terminal.PrintInfo(fmt.Sprintf("No subtasks found for Task %d", task.ID))
		}
		terminal.PrintThickSeparator()
	}

	return nil
}

// showGeneralHelp displays the general help message with a list of available functions
func (t *tester) showGeneralHelp() error {
	functions := GetAvailableFunctions()
	toolsByType := tools.GetToolsByType()

	terminal.PrintHeader("Usage: ftester FUNCTION [ARGUMENTS]")
	fmt.Println()
	terminal.PrintHeader("Built-in functions:")
	terminal.PrintValueFormat("  %-20s", "describe")
	fmt.Printf(" - %s\n", describeFuncInfo.Description)

	// Define type names for better readability
	typeNames := map[tools.ToolType]string{
		tools.EnvironmentToolType:    "Work with terminal and files (work with environment)",
		tools.SearchNetworkToolType:  "Search in the internet",
		tools.SearchVectorDbToolType: "Search in the Vector DB",
		tools.AgentToolType:          "Agents",
	}

	// Process each type in the order we want to display them
	for _, toolType := range []tools.ToolType{
		tools.SearchNetworkToolType,
		tools.EnvironmentToolType,
		tools.SearchVectorDbToolType,
		tools.AgentToolType,
	} {
		// Get type name
		typeName, ok := typeNames[toolType]
		if !ok {
			continue
		}

		// Get tools for this type
		toolsOfType := toolsByType[toolType]
		if len(toolsOfType) == 0 {
			continue
		}

		// Print section header
		fmt.Println()
		terminal.PrintHeader(typeName + ":")

		// Print each function in this group
		for _, tool := range toolsOfType {
			// Skip functions that are not available for user invocation
			if !isToolAvailableForCall(tool) {
				continue
			}

			// Find function info
			var description string
			for _, fn := range functions {
				if fn.Name == tool {
					description = fn.Description
					break
				}
			}

			terminal.PrintValueFormat("  %-20s", tool)
			fmt.Printf(" - %s\n", description)
		}
	}

	fmt.Println()
	terminal.PrintInfo("For help on a specific function, use: ftester FUNCTION -help")
	terminal.PrintKeyValue("Current mode", t.getModeDescription())

	return nil
}

// getModeDescription returns a description of the current mode based on flowID
func (t *tester) getModeDescription() string {
	if t.flowID == 0 {
		return "MOCK (flowID=0)"
	}
	return fmt.Sprintf("REAL (flowID=%d)", t.flowID)
}

// showFunctionHelp displays help for a specific function, including its arguments
func (t *tester) showFunctionHelp(funcName string) error {
	// Get function info
	fnInfo, err := GetFunctionInfo(funcName)
	if err != nil {
		return err
	}

	terminal.PrintHeader(fmt.Sprintf("Function: %s", fnInfo.Name))
	terminal.PrintKeyValue("Description", fnInfo.Description)
	fmt.Println()

	terminal.PrintHeader("Arguments:")

	for _, arg := range fnInfo.Arguments {
		requiredStr := ""
		if arg.Required {
			requiredStr = " (required)"
		}
		terminal.PrintValueFormat("  -%-20s", arg.Name)
		fmt.Printf(" %s%s\n", arg.Description, requiredStr)
	}

	return nil
}

// needsTeminalPrepare determines if a function needs terminal preparation
func (t *tester) needsTeminalPrepare(funcName string) bool {
	// These functions require terminal preparation
	terminalFunctions := map[string]bool{
		tools.TerminalToolName: true,
		tools.FileToolName:     true,
	}

	// For all other functions, no preparation is needed instead of terminal or agents functions
	return terminalFunctions[funcName] || tools.GetToolTypeMapping()[funcName] == tools.AgentToolType
}

// wrapErrorEndSpan wraps an error with an end span in langfuse
func wrapErrorEndSpan(ctx context.Context, span langfuse.Span, msg string, err error) error {
	logrus.WithContext(ctx).WithError(err).Error(msg)
	err = fmt.Errorf("%s: %w", msg, err)
	span.End(
		langfuse.WithEndSpanStatus(err.Error()),
		langfuse.WithEndSpanLevel(langfuse.ObservationLevelError),
	)
	return err
}
