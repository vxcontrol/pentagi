package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"pentagi/pkg/cast"
	"pentagi/pkg/csum"
	"pentagi/pkg/database"
	"pentagi/pkg/graphiti"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/providers/embeddings"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

const ToolPlaceholder = "Always use your function calling functionality, instead of returning a text result."

const TasksNumberLimit = 15

const (
	msgGeneratorSizeLimit = 150 * 1024 // 150 KB
	msgRefinerSizeLimit   = 100 * 1024 // 100 KB
	msgReporterSizeLimit  = 100 * 1024 // 100 KB
	msgSummarizerLimit    = 16 * 1024  // 16 KB
)

const textTruncateMessage = "\n\n[...truncated]"

type PerformResult int

const (
	PerformResultError PerformResult = iota
	PerformResultWaiting
	PerformResultDone
)

type StreamMessageChunkType streaming.ChunkType

const (
	StreamMessageChunkTypeThinking StreamMessageChunkType = "thinking"
	StreamMessageChunkTypeContent  StreamMessageChunkType = "content"
	StreamMessageChunkTypeResult   StreamMessageChunkType = "result"
	StreamMessageChunkTypeFlush    StreamMessageChunkType = "flush"
	StreamMessageChunkTypeUpdate   StreamMessageChunkType = "update"
)

type StreamMessageChunk struct {
	Type         StreamMessageChunkType
	MsgType      database.MsglogType
	Content      string
	Thinking     *reasoning.ContentReasoning
	Result       string
	ResultFormat database.MsglogResultFormat
	StreamID     int64
}

type StreamMessageHandler func(ctx context.Context, chunk *StreamMessageChunk) error

type FlowProvider interface {
	ID() int64
	DB() database.Querier
	Type() provider.ProviderType
	Model(opt pconfig.ProviderOptionsType) string
	Image() string
	Title() string
	Language() string
	ToolCallIDTemplate() string
	Embedder() embeddings.Embedder
	Executor() tools.FlowToolsExecutor
	Prompter() templates.Prompter

	SetTitle(title string)
	SetAgentLogProvider(agentLog tools.AgentLogProvider)
	SetMsgLogProvider(msgLog tools.MsgLogProvider)

	GetTaskTitle(ctx context.Context, input string) (string, error)
	GenerateSubtasks(ctx context.Context, taskID int64) ([]tools.SubtaskInfo, error)
	RefineSubtasks(ctx context.Context, taskID int64) ([]tools.SubtaskInfo, error)
	GetTaskResult(ctx context.Context, taskID int64) (*tools.TaskResult, error)

	PrepareAgentChain(ctx context.Context, taskID, subtaskID int64) (int64, error)
	PerformAgentChain(ctx context.Context, taskID, subtaskID, msgChainID int64) (PerformResult, error)
	PutInputToAgentChain(ctx context.Context, msgChainID int64, input string) error
	EnsureChainConsistency(ctx context.Context, msgChainID int64) error

	FlowProviderHandlers
}

type FlowProviderHandlers interface {
	GetAskAdviceHandler(ctx context.Context, taskID, subtaskID *int64) (tools.ExecutorHandler, error)
	GetCoderHandler(ctx context.Context, taskID, subtaskID *int64) (tools.ExecutorHandler, error)
	GetInstallerHandler(ctx context.Context, taskID, subtaskID *int64) (tools.ExecutorHandler, error)
	GetMemoristHandler(ctx context.Context, taskID, subtaskID *int64) (tools.ExecutorHandler, error)
	GetPentesterHandler(ctx context.Context, taskID, subtaskID *int64) (tools.ExecutorHandler, error)
	GetSubtaskSearcherHandler(ctx context.Context, taskID, subtaskID *int64) (tools.ExecutorHandler, error)
	GetTaskSearcherHandler(ctx context.Context, taskID int64) (tools.ExecutorHandler, error)
	GetSummarizeResultHandler(taskID, subtaskID *int64) tools.SummarizeHandler
}

type tasksInfo struct {
	Task     database.Task
	Tasks    []database.Task
	Subtasks []database.Subtask
}

type subtasksInfo struct {
	Subtask   *database.Subtask
	Planned   []database.Subtask
	Completed []database.Subtask
}

type flowProvider struct {
	db database.Querier
	mx *sync.RWMutex

	embedder       embeddings.Embedder
	graphitiClient *graphiti.Client

	flowID   int64
	publicIP string

	callCounter *atomic.Int64

	image    string
	title    string
	language string
	askUser  bool

	tcIDTemplate string

	prompter templates.Prompter
	executor tools.FlowToolsExecutor
	agentLog tools.AgentLogProvider
	msgLog   tools.MsgLogProvider
	streamCb StreamMessageHandler

	summarizer csum.Summarizer

	provider.Provider
}

func (fp *flowProvider) SetAgentLogProvider(agentLog tools.AgentLogProvider) {
	fp.mx.Lock()
	defer fp.mx.Unlock()

	fp.agentLog = agentLog
}

func (fp *flowProvider) SetMsgLogProvider(msgLog tools.MsgLogProvider) {
	fp.mx.Lock()
	defer fp.mx.Unlock()

	fp.msgLog = msgLog
}

func (fp *flowProvider) ID() int64 {
	fp.mx.RLock()
	defer fp.mx.RUnlock()

	return fp.flowID
}

func (fp *flowProvider) DB() database.Querier {
	fp.mx.RLock()
	defer fp.mx.RUnlock()

	return fp.db
}

func (fp *flowProvider) Image() string {
	fp.mx.RLock()
	defer fp.mx.RUnlock()

	return fp.image
}

func (fp *flowProvider) Title() string {
	fp.mx.RLock()
	defer fp.mx.RUnlock()

	return fp.title
}

func (fp *flowProvider) SetTitle(title string) {
	fp.mx.Lock()
	defer fp.mx.Unlock()

	fp.title = title
}

func (fp *flowProvider) Language() string {
	fp.mx.RLock()
	defer fp.mx.RUnlock()

	return fp.language
}

func (fp *flowProvider) ToolCallIDTemplate() string {
	fp.mx.RLock()
	defer fp.mx.RUnlock()

	return fp.tcIDTemplate
}

func (fp *flowProvider) Embedder() embeddings.Embedder {
	fp.mx.RLock()
	defer fp.mx.RUnlock()

	return fp.embedder
}

func (fp *flowProvider) Executor() tools.FlowToolsExecutor {
	fp.mx.RLock()
	defer fp.mx.RUnlock()

	return fp.executor
}

func (fp *flowProvider) Prompter() templates.Prompter {
	fp.mx.RLock()
	defer fp.mx.RUnlock()

	return fp.prompter
}

func (fp *flowProvider) GetTaskTitle(ctx context.Context, input string) (string, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.GetTaskTitle")
	defer span.End()

	ctx, observation := obs.Observer.NewObservation(ctx)
	getterEvaluator := observation.Evaluator(
		langfuse.WithEvaluatorName("get task title"),
		langfuse.WithEvaluatorInput(input),
		langfuse.WithEvaluatorMetadata(langfuse.Metadata{
			"lang": fp.language,
		}),
	)
	ctx, _ = getterEvaluator.Observation(ctx)

	titleTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeTaskDescriptor, map[string]any{
		"Input":       input,
		"Lang":        fp.language,
		"CurrentTime": getCurrentTime(),
		"N":           150,
	})
	if err != nil {
		return "", wrapErrorEndEvaluatorSpan(ctx, getterEvaluator, "failed to get flow title template", err)
	}

	title, err := fp.Call(ctx, pconfig.OptionsTypeSimple, titleTmpl)
	if err != nil {
		return "", wrapErrorEndEvaluatorSpan(ctx, getterEvaluator, "failed to get flow title", err)
	}

	getterEvaluator.End(
		langfuse.WithEvaluatorStatus("success"),
		langfuse.WithEvaluatorOutput(title),
	)

	return title, nil
}

func (fp *flowProvider) GenerateSubtasks(ctx context.Context, taskID int64) ([]tools.SubtaskInfo, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.GenerateSubtasks")
	defer span.End()

	logger := logrus.WithContext(ctx).WithField("task_id", taskID)

	tasksInfo, err := fp.getTasksInfo(ctx, taskID)
	if err != nil {
		logger.WithError(err).Error("failed to get tasks info")
		return nil, fmt.Errorf("failed to get tasks info: %w", err)
	}

	generatorContext := map[string]map[string]any{
		"user": {
			"Task":     tasksInfo.Task,
			"Tasks":    tasksInfo.Tasks,
			"Subtasks": tasksInfo.Subtasks,
		},
		"system": {
			"SubtaskListToolName":     tools.SubtaskListToolName,
			"SearchToolName":          tools.SearchToolName,
			"TerminalToolName":        tools.TerminalToolName,
			"FileToolName":            tools.FileToolName,
			"BrowserToolName":         tools.BrowserToolName,
			"SummarizationToolName":   cast.SummarizationToolName,
			"SummarizedContentPrefix": strings.ReplaceAll(csum.SummarizedContentPrefix, "\n", "\\n"),
			"DockerImage":             fp.image,
			"Lang":                    fp.language,
			"CurrentTime":             getCurrentTime(),
			"N":                       TasksNumberLimit,
			"ToolPlaceholder":         ToolPlaceholder,
		},
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	generatorEvaluator := observation.Evaluator(
		langfuse.WithEvaluatorName("subtasks generator"),
		langfuse.WithEvaluatorInput(tasksInfo),
		langfuse.WithEvaluatorMetadata(langfuse.Metadata{
			"user_context":   generatorContext["user"],
			"system_context": generatorContext["system"],
		}),
	)
	ctx, _ = generatorEvaluator.Observation(ctx)

	generatorTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeSubtasksGenerator, generatorContext["user"])
	if err != nil {
		return nil, wrapErrorEndEvaluatorSpan(ctx, generatorEvaluator, "failed to get task generator template", err)
	}

	subtasksLen := len(tasksInfo.Subtasks)
	for l := subtasksLen; l > 2; l /= 2 {
		if len(generatorTmpl) < msgGeneratorSizeLimit {
			break
		}

		generatorContext["user"]["Subtasks"] = tasksInfo.Subtasks[(subtasksLen - l):]
		generatorTmpl, err = fp.prompter.RenderTemplate(templates.PromptTypeSubtasksGenerator, generatorContext["user"])
		if err != nil {
			return nil, wrapErrorEndEvaluatorSpan(ctx, generatorEvaluator, "failed to get task generator template", err)
		}
	}

	systemGeneratorTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeGenerator, generatorContext["system"])
	if err != nil {
		return nil, wrapErrorEndEvaluatorSpan(ctx, generatorEvaluator, "failed to get task system generator template", err)
	}

	subtasks, err := fp.performSubtasksGenerator(ctx, taskID, systemGeneratorTmpl, generatorTmpl, tasksInfo.Task.Input)
	if err != nil {
		return nil, wrapErrorEndEvaluatorSpan(ctx, generatorEvaluator, "failed to perform subtasks generator", err)
	}

	generatorEvaluator.End(
		langfuse.WithEvaluatorStatus("success"),
		langfuse.WithEvaluatorOutput(subtasks),
	)

	return subtasks, nil
}

func (fp *flowProvider) RefineSubtasks(ctx context.Context, taskID int64) ([]tools.SubtaskInfo, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.RefineSubtasks")
	defer span.End()

	logger := logrus.WithContext(ctx).WithField("task_id", taskID)

	tasksInfo, err := fp.getTasksInfo(ctx, taskID)
	if err != nil {
		logger.WithError(err).Error("failed to get tasks info")
		return nil, fmt.Errorf("failed to get tasks info: %w", err)
	}

	subtasksInfo := fp.getSubtasksInfo(taskID, tasksInfo.Subtasks)

	logger.WithFields(logrus.Fields{
		"planned_count":   len(subtasksInfo.Planned),
		"completed_count": len(subtasksInfo.Completed),
	}).Debug("retrieved subtasks info for refinement")

	refinerContext := map[string]map[string]any{
		"user": {
			"Task":              tasksInfo.Task,
			"Tasks":             tasksInfo.Tasks,
			"PlannedSubtasks":   subtasksInfo.Planned,
			"CompletedSubtasks": subtasksInfo.Completed,
		},
		"system": {
			"SubtaskPatchToolName":    tools.SubtaskPatchToolName,
			"SubtaskListToolName":     tools.SubtaskListToolName,
			"SearchToolName":          tools.SearchToolName,
			"TerminalToolName":        tools.TerminalToolName,
			"FileToolName":            tools.FileToolName,
			"BrowserToolName":         tools.BrowserToolName,
			"SummarizationToolName":   cast.SummarizationToolName,
			"SummarizedContentPrefix": strings.ReplaceAll(csum.SummarizedContentPrefix, "\n", "\\n"),
			"DockerImage":             fp.image,
			"Lang":                    fp.language,
			"CurrentTime":             getCurrentTime(),
			"N":                       max(TasksNumberLimit-len(subtasksInfo.Completed), 0),
			"ToolPlaceholder":         ToolPlaceholder,
		},
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	refinerEvaluator := observation.Evaluator(
		langfuse.WithEvaluatorName("subtasks refiner"),
		langfuse.WithEvaluatorInput(refinerContext),
		langfuse.WithEvaluatorMetadata(langfuse.Metadata{
			"user_context":   refinerContext["user"],
			"system_context": refinerContext["system"],
		}),
	)
	ctx, _ = refinerEvaluator.Observation(ctx)

	refinerTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeSubtasksRefiner, refinerContext["user"])
	if err != nil {
		return nil, wrapErrorEndEvaluatorSpan(ctx, refinerEvaluator, "failed to get task subtasks refiner template (1)", err)
	}

	// TODO: here need to store it in the database and use it as a cache for next runs
	if len(refinerTmpl) < msgRefinerSizeLimit {
		summarizerHandler := fp.GetSummarizeResultHandler(&taskID, nil)
		executionState, err := fp.getTaskPrimaryAgentChainSummary(ctx, taskID, summarizerHandler)
		if err != nil {
			return nil, wrapErrorEndEvaluatorSpan(ctx, refinerEvaluator, "failed to prepare execution state", err)
		}

		refinerContext["user"]["ExecutionState"] = executionState
		refinerTmpl, err = fp.prompter.RenderTemplate(templates.PromptTypeSubtasksRefiner, refinerContext["user"])
		if err != nil {
			return nil, wrapErrorEndEvaluatorSpan(ctx, refinerEvaluator, "failed to get task subtasks refiner template (2)", err)
		}

		if len(refinerTmpl) < msgRefinerSizeLimit {
			msgLogsSummary, err := fp.getTaskMsgLogsSummary(ctx, taskID, summarizerHandler)
			if err != nil {
				return nil, wrapErrorEndEvaluatorSpan(ctx, refinerEvaluator, "failed to get task msg logs summary", err)
			}

			refinerContext["user"]["ExecutionLogs"] = msgLogsSummary
			refinerTmpl, err = fp.prompter.RenderTemplate(templates.PromptTypeSubtasksRefiner, refinerContext["user"])
			if err != nil {
				return nil, wrapErrorEndEvaluatorSpan(ctx, refinerEvaluator, "failed to get task subtasks refiner template (3)", err)
			}
		}
	}

	systemRefinerTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeRefiner, refinerContext["system"])
	if err != nil {
		return nil, wrapErrorEndEvaluatorSpan(ctx, refinerEvaluator, "failed to get task system refiner template", err)
	}

	subtasks, err := fp.performSubtasksRefiner(ctx, taskID, subtasksInfo.Planned, systemRefinerTmpl, refinerTmpl, tasksInfo.Task.Input)
	if err != nil {
		return nil, wrapErrorEndEvaluatorSpan(ctx, refinerEvaluator, "failed to perform subtasks refiner", err)
	}

	refinerEvaluator.End(
		langfuse.WithEvaluatorStatus("success"),
		langfuse.WithEvaluatorOutput(subtasks),
	)

	return subtasks, nil
}

func (fp *flowProvider) GetTaskResult(ctx context.Context, taskID int64) (*tools.TaskResult, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.GetTaskResult")
	defer span.End()

	logger := logrus.WithContext(ctx).WithField("task_id", taskID)

	tasksInfo, err := fp.getTasksInfo(ctx, taskID)
	if err != nil {
		logger.WithError(err).Error("failed to get tasks info")
		return nil, fmt.Errorf("failed to get tasks info: %w", err)
	}

	subtasksInfo := fp.getSubtasksInfo(taskID, tasksInfo.Subtasks)
	reporterContext := map[string]map[string]any{
		"user": {
			"Task":              tasksInfo.Task,
			"Tasks":             tasksInfo.Tasks,
			"CompletedSubtasks": subtasksInfo.Completed,
			"PlannedSubtasks":   subtasksInfo.Planned,
		},
		"system": {
			"ReportResultToolName":    tools.ReportResultToolName,
			"SummarizationToolName":   cast.SummarizationToolName,
			"SummarizedContentPrefix": strings.ReplaceAll(csum.SummarizedContentPrefix, "\n", "\\n"),
			"Lang":                    fp.language,
			"N":                       4000,
			"ToolPlaceholder":         ToolPlaceholder,
		},
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	reporterEvaluator := observation.Evaluator(
		langfuse.WithEvaluatorName("reporter agent"),
		langfuse.WithEvaluatorInput(reporterContext),
		langfuse.WithEvaluatorMetadata(langfuse.Metadata{
			"user_context":   reporterContext["user"],
			"system_context": reporterContext["system"],
		}),
	)
	ctx, _ = reporterEvaluator.Observation(ctx)

	reporterTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeTaskReporter, reporterContext["user"])
	if err != nil {
		return nil, wrapErrorEndEvaluatorSpan(ctx, reporterEvaluator, "failed to get task reporter template (1)", err)
	}

	if len(reporterTmpl) < msgReporterSizeLimit {
		summarizerHandler := fp.GetSummarizeResultHandler(&taskID, nil)
		executionState, err := fp.getTaskPrimaryAgentChainSummary(ctx, taskID, summarizerHandler)
		if err != nil {
			return nil, wrapErrorEndEvaluatorSpan(ctx, reporterEvaluator, "failed to prepare execution state", err)
		}

		reporterContext["user"]["ExecutionState"] = executionState
		reporterTmpl, err = fp.prompter.RenderTemplate(templates.PromptTypeTaskReporter, reporterContext["user"])
		if err != nil {
			return nil, wrapErrorEndEvaluatorSpan(ctx, reporterEvaluator, "failed to get task reporter template (2)", err)
		}

		if len(reporterTmpl) < msgReporterSizeLimit {
			msgLogsSummary, err := fp.getTaskMsgLogsSummary(ctx, taskID, summarizerHandler)
			if err != nil {
				return nil, wrapErrorEndEvaluatorSpan(ctx, reporterEvaluator, "failed to get task msg logs summary", err)
			}

			reporterContext["user"]["ExecutionLogs"] = msgLogsSummary
			reporterTmpl, err = fp.prompter.RenderTemplate(templates.PromptTypeTaskReporter, reporterContext["user"])
			if err != nil {
				return nil, wrapErrorEndEvaluatorSpan(ctx, reporterEvaluator, "failed to get task reporter template (3)", err)
			}
		}
	}

	systemReporterTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeReporter, reporterContext["system"])
	if err != nil {
		return nil, wrapErrorEndEvaluatorSpan(ctx, reporterEvaluator, "failed to get task system reporter template", err)
	}

	result, err := fp.performTaskResultReporter(ctx, &taskID, nil, systemReporterTmpl, reporterTmpl, tasksInfo.Task.Input)
	if err != nil {
		return nil, wrapErrorEndEvaluatorSpan(ctx, reporterEvaluator, "failed to perform task result reporter", err)
	}

	reporterEvaluator.End(
		langfuse.WithEvaluatorStatus("success"),
		langfuse.WithEvaluatorOutput(result),
	)

	return result, nil
}

func (fp *flowProvider) PrepareAgentChain(ctx context.Context, taskID, subtaskID int64) (int64, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.PrepareAgentChain")
	defer span.End()

	optAgentType := pconfig.OptionsTypePrimaryAgent
	msgChainType := database.MsgchainTypePrimaryAgent

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"provider":   fp.Type(),
		"agent":      optAgentType,
		"flow_id":    fp.flowID,
		"task_id":    taskID,
		"subtask_id": subtaskID,
	})

	subtask, err := fp.db.GetSubtask(ctx, subtaskID)
	if err != nil {
		logger.WithError(err).Error("failed to get subtask")
		return 0, fmt.Errorf("failed to get subtask: %w", err)
	}

	executionContext, err := fp.prepareExecutionContext(ctx, taskID, subtaskID)
	if err != nil {
		logger.WithError(err).Error("failed to prepare execution context")
		return 0, fmt.Errorf("failed to prepare execution context: %w", err)
	}

	subtask, err = fp.db.UpdateSubtaskContext(ctx, database.UpdateSubtaskContextParams{
		Context: executionContext,
		ID:      subtaskID,
	})
	if err != nil {
		logger.WithError(err).Error("failed to update subtask context")
		return 0, fmt.Errorf("failed to update subtask context: %w", err)
	}

	systemAgentTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypePrimaryAgent, map[string]any{
		"FinalyToolName":          tools.FinalyToolName,
		"SearchToolName":          tools.SearchToolName,
		"PentesterToolName":       tools.PentesterToolName,
		"CoderToolName":           tools.CoderToolName,
		"AdviceToolName":          tools.AdviceToolName,
		"MemoristToolName":        tools.MemoristToolName,
		"MaintenanceToolName":     tools.MaintenanceToolName,
		"SummarizationToolName":   cast.SummarizationToolName,
		"SummarizedContentPrefix": strings.ReplaceAll(csum.SummarizedContentPrefix, "\n", "\\n"),
		"AskUserToolName":         tools.AskUserToolName,
		"AskUserEnabled":          fp.askUser,
		"ExecutionContext":        executionContext,
		"Lang":                    fp.language,
		"DockerImage":             fp.image,
		"CurrentTime":             getCurrentTime(),
		"ToolPlaceholder":         ToolPlaceholder,
	})
	if err != nil {
		logger.WithError(err).Error("failed to get system prompt for primary agent template")
		return 0, fmt.Errorf("failed to get system prompt for primary agent template: %w", err)
	}

	msgChainID, _, err := fp.restoreChain(
		ctx, &taskID, &subtaskID, optAgentType, msgChainType, systemAgentTmpl, subtask.Description,
	)
	if err != nil {
		logger.WithError(err).Error("failed to restore primary agent msg chain")
		return 0, fmt.Errorf("failed to restore primary agent msg chain: %w", err)
	}

	return msgChainID, nil
}

func (fp *flowProvider) PerformAgentChain(ctx context.Context, taskID, subtaskID, msgChainID int64) (PerformResult, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.PerformAgentChain")
	defer span.End()

	optAgentType := pconfig.OptionsTypePrimaryAgent
	msgChainType := database.MsgchainTypePrimaryAgent

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"provider":     fp.Type(),
		"agent":        optAgentType,
		"flow_id":      fp.flowID,
		"task_id":      taskID,
		"subtask_id":   subtaskID,
		"msg_chain_id": msgChainID,
	})

	msgChain, err := fp.db.GetMsgChain(ctx, msgChainID)
	if err != nil {
		logger.WithError(err).Error("failed to get primary agent msg chain")
		return PerformResultError, fmt.Errorf("failed to get primary agent msg chain %d: %w", msgChainID, err)
	}

	var chain []llms.MessageContent
	if err := json.Unmarshal(msgChain.Chain, &chain); err != nil {
		logger.WithError(err).Error("failed to unmarshal primary agent msg chain")
		return PerformResultError, fmt.Errorf("failed to unmarshal primary agent msg chain %d: %w", msgChainID, err)
	}

	adviser, err := fp.GetAskAdviceHandler(ctx, &taskID, &subtaskID)
	if err != nil {
		logger.WithError(err).Error("failed to get ask advice handler")
		return PerformResultError, fmt.Errorf("failed to get ask advice handler: %w", err)
	}

	coder, err := fp.GetCoderHandler(ctx, &taskID, &subtaskID)
	if err != nil {
		logger.WithError(err).Error("failed to get coder handler")
		return PerformResultError, fmt.Errorf("failed to get coder handler: %w", err)
	}

	installer, err := fp.GetInstallerHandler(ctx, &taskID, &subtaskID)
	if err != nil {
		logger.WithError(err).Error("failed to get installer handler")
		return PerformResultError, fmt.Errorf("failed to get installer handler: %w", err)
	}

	memorist, err := fp.GetMemoristHandler(ctx, &taskID, &subtaskID)
	if err != nil {
		logger.WithError(err).Error("failed to get memorist handler")
		return PerformResultError, fmt.Errorf("failed to get memorist handler: %w", err)
	}

	pentester, err := fp.GetPentesterHandler(ctx, &taskID, &subtaskID)
	if err != nil {
		logger.WithError(err).Error("failed to get pentester handler")
		return PerformResultError, fmt.Errorf("failed to get pentester handler: %w", err)
	}

	searcher, err := fp.GetSubtaskSearcherHandler(ctx, &taskID, &subtaskID)
	if err != nil {
		logger.WithError(err).Error("failed to get searcher handler")
		return PerformResultError, fmt.Errorf("failed to get searcher handler: %w", err)
	}

	subtask, err := fp.db.GetSubtask(ctx, subtaskID)
	if err != nil {
		logger.WithError(err).Error("failed to get subtask")
		return PerformResultError, fmt.Errorf("failed to get subtask: %w", err)
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	executorAgent := observation.Agent(
		langfuse.WithAgentName(fmt.Sprintf("primary agent for subtask %d: %s", subtaskID, subtask.Title)),
		langfuse.WithAgentInput(chain),
		langfuse.WithAgentMetadata(langfuse.Metadata{
			"flow_id":      fp.flowID,
			"task_id":      taskID,
			"subtask_id":   subtaskID,
			"msg_chain_id": msgChainID,
			"provider":     fp.Type(),
			"image":        fp.image,
			"lang":         fp.language,
			"description":  subtask.Description,
		}),
	)
	ctx, _ = executorAgent.Observation(ctx)

	performResult := PerformResultError
	cfg := tools.PrimaryExecutorConfig{
		TaskID:    taskID,
		SubtaskID: subtaskID,
		Adviser:   adviser,
		Coder:     coder,
		Installer: installer,
		Memorist:  memorist,
		Pentester: pentester,
		Searcher:  searcher,
		Barrier: func(ctx context.Context, name string, args json.RawMessage) (string, error) {
			loggerFunc := logger.WithContext(ctx).WithFields(logrus.Fields{
				"name": name,
				"args": string(args),
			})

			switch name {
			case tools.FinalyToolName:
				var done tools.Done
				if err := json.Unmarshal(args, &done); err != nil {
					loggerFunc.WithError(err).Error("failed to unmarshal done result")
					return "", fmt.Errorf("failed to unmarshal done result: %w", err)
				}

				loggerFunc = loggerFunc.WithFields(logrus.Fields{
					"status": done.Success,
					"result": done.Result[:min(len(done.Result), 1000)],
				})

				opts := []langfuse.AgentOption{
					langfuse.WithAgentOutput(done.Result),
				}
				defer func() {
					executorAgent.End(opts...)
				}()

				if !done.Success {
					performResult = PerformResultError
					opts = append(opts,
						langfuse.WithAgentStatus("done handler: failed"),
						langfuse.WithAgentLevel(langfuse.ObservationLevelWarning),
					)
				} else {
					performResult = PerformResultDone
					opts = append(opts,
						langfuse.WithAgentStatus("done handler: success"),
					)
				}

				// TODO: here need to call SetResult from SubtaskWorker interface
				subtask, err = fp.db.UpdateSubtaskResult(ctx, database.UpdateSubtaskResultParams{
					Result: done.Result,
					ID:     subtaskID,
				})
				if err != nil {
					opts = append(opts,
						langfuse.WithAgentStatus(err.Error()),
						langfuse.WithAgentLevel(langfuse.ObservationLevelError),
					)
					loggerFunc.WithError(err).Error("failed to update subtask result")
					return "", fmt.Errorf("failed to update subtask %d result: %w", subtaskID, err)
				}

				// report result to msg log as a final message for the subtask execution
				reportMsgID, err := fp.putMsgLog(
					ctx,
					database.MsglogTypeReport,
					&taskID, &subtaskID, 0,
					"", subtask.Description,
				)
				if err != nil {
					opts = append(opts,
						langfuse.WithAgentStatus(err.Error()),
						langfuse.WithAgentLevel(langfuse.ObservationLevelError),
					)
					loggerFunc.WithError(err).Error("failed to put report msg")
					return "", fmt.Errorf("failed to put report msg: %w", err)
				}

				err = fp.updateMsgLogResult(
					ctx,
					reportMsgID, 0,
					done.Result, database.MsglogResultFormatMarkdown,
				)
				if err != nil {
					opts = append(opts,
						langfuse.WithAgentStatus(err.Error()),
						langfuse.WithAgentLevel(langfuse.ObservationLevelError),
					)
					loggerFunc.WithError(err).Error("failed to update report msg result")
					return "", fmt.Errorf("failed to update report msg result: %w", err)
				}

			case tools.AskUserToolName:
				performResult = PerformResultWaiting

				var askUser tools.AskUser
				if err := json.Unmarshal(args, &askUser); err != nil {
					loggerFunc.WithError(err).Error("failed to unmarshal ask user result")
					return "", fmt.Errorf("failed to unmarshal ask user result: %w", err)
				}

				executorAgent.End(
					langfuse.WithAgentOutput(askUser.Message),
					langfuse.WithAgentStatus("ask user handler"),
				)
			}

			return fmt.Sprintf("function %s successfully processed arguments", name), nil
		},
		Summarizer: fp.GetSummarizeResultHandler(&taskID, &subtaskID),
	}

	executor, err := fp.executor.GetPrimaryExecutor(cfg)
	if err != nil {
		return PerformResultError, wrapErrorEndAgentSpan(ctx, executorAgent, "failed to get primary executor", err)
	}

	ctx = tools.PutAgentContext(ctx, msgChainType)
	err = fp.performAgentChain(
		ctx, optAgentType, msgChain.ID, &taskID, &subtaskID, chain, executor, fp.summarizer,
	)
	if err != nil {
		return PerformResultError, wrapErrorEndAgentSpan(ctx, executorAgent, "failed to perform primary agent chain", err)
	}

	executorAgent.End()

	return performResult, nil
}

func (fp *flowProvider) PutInputToAgentChain(ctx context.Context, msgChainID int64, input string) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.PutInputToAgentChain")
	defer span.End()

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"provider":     fp.Type(),
		"flow_id":      fp.flowID,
		"msg_chain_id": msgChainID,
		"input":        input[:min(len(input), 1000)],
	})

	return fp.processChain(ctx, msgChainID, logger, func(chain []llms.MessageContent) ([]llms.MessageContent, error) {
		return fp.updateMsgChainResult(chain, tools.AskUserToolName, input)
	})
}

// EnsureChainConsistency ensures a message chain is in a consistent state by adding
// default responses to any unresponded tool calls.
func (fp *flowProvider) EnsureChainConsistency(ctx context.Context, msgChainID int64) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.EnsureChainConsistency")
	defer span.End()

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"provider":     fp.Type(),
		"flow_id":      fp.flowID,
		"msg_chain_id": msgChainID,
	})

	return fp.processChain(ctx, msgChainID, logger, func(chain []llms.MessageContent) ([]llms.MessageContent, error) {
		return fp.ensureChainConsistency(chain)
	})
}

func (fp *flowProvider) putMsgLog(
	ctx context.Context,
	msgType database.MsglogType,
	taskID, subtaskID *int64,
	streamID int64,
	thinking, msg string,
) (int64, error) {
	fp.mx.RLock()
	msgLog := fp.msgLog
	fp.mx.RUnlock()

	if msgLog == nil {
		return 0, nil
	}

	return msgLog.PutMsg(ctx, msgType, taskID, subtaskID, streamID, thinking, msg)
}

func (fp *flowProvider) updateMsgLogResult(
	ctx context.Context,
	msgID, streamID int64,
	result string,
	resultFormat database.MsglogResultFormat,
) error {
	fp.mx.RLock()
	msgLog := fp.msgLog
	fp.mx.RUnlock()

	if msgLog == nil || msgID <= 0 {
		return nil
	}

	return msgLog.UpdateMsgResult(ctx, msgID, streamID, result, resultFormat)
}

func (fp *flowProvider) putAgentLog(
	ctx context.Context,
	initiator, executor database.MsgchainType,
	task, result string,
	taskID, subtaskID *int64,
) (int64, error) {
	fp.mx.RLock()
	agentLog := fp.agentLog
	fp.mx.RUnlock()

	if agentLog == nil {
		return 0, nil
	}

	return agentLog.PutLog(ctx, initiator, executor, task, result, taskID, subtaskID)
}
