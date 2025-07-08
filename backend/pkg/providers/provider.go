package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"

	"pentagi/pkg/cast"
	"pentagi/pkg/csum"
	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/providers/embeddings"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
	"github.com/vxcontrol/langchaingo/llms"
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
	Thinking     string
	Result       string
	ResultFormat database.MsglogResultFormat
	StreamID     int64
}

type StreamMessageHandler func(ctx context.Context, chunk *StreamMessageChunk) error

type FlowProvider interface {
	Type() provider.ProviderType
	Model(opt pconfig.ProviderOptionsType) string
	Image() string
	Title() string
	Language() string
	Embedder() embeddings.Embedder

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

	embedder embeddings.Embedder

	flowID   int64
	publicIP string

	callCounter *atomic.Int64

	image    string
	title    string
	language string

	prompter templates.Prompter
	executor tools.FlowToolsExecutor
	agentLog tools.AgentLogProvider
	msgLog   tools.MsgLogProvider
	streamCb StreamMessageHandler

	summarizer csum.Summarizer

	provider.Provider
}

func (fp *flowProvider) SetAgentLogProvider(agentLog tools.AgentLogProvider) {
	fp.agentLog = agentLog
}

func (fp *flowProvider) SetMsgLogProvider(msgLog tools.MsgLogProvider) {
	fp.msgLog = msgLog
}

func (fp *flowProvider) Image() string {
	return fp.image
}

func (fp *flowProvider) Title() string {
	return fp.title
}

func (fp *flowProvider) Language() string {
	return fp.language
}

func (fp *flowProvider) Embedder() embeddings.Embedder {
	return fp.embedder
}

func (fp *flowProvider) GetTaskTitle(ctx context.Context, input string) (string, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.GetTaskTitle")
	defer span.End()

	ctx, observation := obs.Observer.NewObservation(ctx)
	getterSpan := observation.Span(
		langfuse.WithStartSpanName("get task title"),
		langfuse.WithStartSpanInput(input),
		langfuse.WithStartSpanMetadata(langfuse.Metadata{
			"lang": fp.language,
		}),
	)
	ctx, _ = getterSpan.Observation(ctx)

	titleTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeTaskDescriptor, map[string]any{
		"Input":       input,
		"Lang":        fp.language,
		"CurrentTime": getCurrentTime(),
		"N":           150,
	})
	if err != nil {
		return "", wrapErrorEndSpan(ctx, getterSpan, "failed to get flow title template", err)
	}

	title, err := fp.Call(ctx, pconfig.OptionsTypeSimple, titleTmpl)
	if err != nil {
		return "", wrapErrorEndSpan(ctx, getterSpan, "failed to get flow title", err)
	}

	getterSpan.End(
		langfuse.WithEndSpanStatus("success"),
		langfuse.WithEndSpanOutput(title),
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
	generatorSpan := observation.Span(
		langfuse.WithStartSpanName("generator agent"),
		langfuse.WithStartSpanInput(tasksInfo),
		langfuse.WithStartSpanMetadata(langfuse.Metadata{
			"user_context":   generatorContext["user"],
			"system_context": generatorContext["system"],
		}),
	)
	ctx, _ = generatorSpan.Observation(ctx)

	generatorTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeSubtasksGenerator, generatorContext["user"])
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, generatorSpan, "failed to get task generator template", err)
	}

	subtasksLen := len(tasksInfo.Subtasks)
	for l := subtasksLen; l > 2; l /= 2 {
		if len(generatorTmpl) < msgGeneratorSizeLimit {
			break
		}

		generatorContext["user"]["Subtasks"] = tasksInfo.Subtasks[(subtasksLen - l):]
		generatorTmpl, err = fp.prompter.RenderTemplate(templates.PromptTypeSubtasksGenerator, generatorContext["user"])
		if err != nil {
			return nil, wrapErrorEndSpan(ctx, generatorSpan, "failed to get task generator template", err)
		}
	}

	systemGeneratorTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeGenerator, generatorContext["system"])
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, generatorSpan, "failed to get task system generator template", err)
	}

	subtasks, err := fp.performSubtasksGenerator(ctx, taskID, systemGeneratorTmpl, generatorTmpl, tasksInfo.Task.Input)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, generatorSpan, "failed to perform subtasks generator", err)
	}

	generatorSpan.End(
		langfuse.WithEndSpanStatus("success"),
		langfuse.WithEndSpanOutput(subtasks),
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

	refinerContext := map[string]map[string]any{
		"user": {
			"Task":              tasksInfo.Task,
			"Tasks":             tasksInfo.Tasks,
			"PlannedSubtasks":   subtasksInfo.Planned,
			"CompletedSubtasks": subtasksInfo.Completed,
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
			"N":                       max(TasksNumberLimit-len(subtasksInfo.Completed), 0),
			"ToolPlaceholder":         ToolPlaceholder,
		},
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	refinerSpan := observation.Span(
		langfuse.WithStartSpanName("refiner agent"),
		langfuse.WithStartSpanInput(refinerContext),
		langfuse.WithStartSpanMetadata(langfuse.Metadata{
			"user_context":   refinerContext["user"],
			"system_context": refinerContext["system"],
		}),
	)
	ctx, _ = refinerSpan.Observation(ctx)

	refinerTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeSubtasksRefiner, refinerContext["user"])
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, refinerSpan, "failed to get task subtasks refiner template (1)", err)
	}

	// TODO: here need to store it in the database and use it as a cache for next runs
	if len(refinerTmpl) < msgRefinerSizeLimit {
		summarizerHandler := fp.GetSummarizeResultHandler(&taskID, nil)
		executionState, err := fp.getTaskPrimaryAgentChainSummary(ctx, taskID, summarizerHandler)
		if err != nil {
			return nil, wrapErrorEndSpan(ctx, refinerSpan, "failed to prepare execution state", err)
		}

		refinerContext["user"]["ExecutionState"] = executionState
		refinerTmpl, err = fp.prompter.RenderTemplate(templates.PromptTypeSubtasksRefiner, refinerContext["user"])
		if err != nil {
			return nil, wrapErrorEndSpan(ctx, refinerSpan, "failed to get task subtasks refiner template (2)", err)
		}

		if len(refinerTmpl) < msgRefinerSizeLimit {
			msgLogsSummary, err := fp.getTaskMsgLogsSummary(ctx, taskID, summarizerHandler)
			if err != nil {
				return nil, wrapErrorEndSpan(ctx, refinerSpan, "failed to get task msg logs summary", err)
			}

			refinerContext["user"]["ExecutionLogs"] = msgLogsSummary
			refinerTmpl, err = fp.prompter.RenderTemplate(templates.PromptTypeSubtasksRefiner, refinerContext["user"])
			if err != nil {
				return nil, wrapErrorEndSpan(ctx, refinerSpan, "failed to get task subtasks refiner template (3)", err)
			}
		}
	}

	systemRefinerTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeRefiner, refinerContext["system"])
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, refinerSpan, "failed to get task system refiner template", err)
	}

	subtasks, err := fp.performSubtasksRefiner(ctx, taskID, systemRefinerTmpl, refinerTmpl, tasksInfo.Task.Input)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, refinerSpan, "failed to perform subtasks refiner", err)
	}

	refinerSpan.End(
		langfuse.WithEndSpanStatus("success"),
		langfuse.WithEndSpanOutput(subtasks),
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
	reporterSpan := observation.Span(
		langfuse.WithStartSpanName("reporter agent"),
		langfuse.WithStartSpanInput(reporterContext),
		langfuse.WithStartSpanMetadata(langfuse.Metadata{
			"user_context":   reporterContext["user"],
			"system_context": reporterContext["system"],
		}),
	)
	ctx, _ = reporterSpan.Observation(ctx)

	reporterTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeTaskReporter, reporterContext["user"])
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, reporterSpan, "failed to get task reporter template (1)", err)
	}

	if len(reporterTmpl) < msgReporterSizeLimit {
		summarizerHandler := fp.GetSummarizeResultHandler(&taskID, nil)
		executionState, err := fp.getTaskPrimaryAgentChainSummary(ctx, taskID, summarizerHandler)
		if err != nil {
			return nil, wrapErrorEndSpan(ctx, reporterSpan, "failed to prepare execution state", err)
		}

		reporterContext["user"]["ExecutionState"] = executionState
		reporterTmpl, err = fp.prompter.RenderTemplate(templates.PromptTypeTaskReporter, reporterContext["user"])
		if err != nil {
			return nil, wrapErrorEndSpan(ctx, reporterSpan, "failed to get task reporter template (2)", err)
		}

		if len(reporterTmpl) < msgReporterSizeLimit {
			msgLogsSummary, err := fp.getTaskMsgLogsSummary(ctx, taskID, summarizerHandler)
			if err != nil {
				return nil, wrapErrorEndSpan(ctx, reporterSpan, "failed to get task msg logs summary", err)
			}

			reporterContext["user"]["ExecutionLogs"] = msgLogsSummary
			reporterTmpl, err = fp.prompter.RenderTemplate(templates.PromptTypeTaskReporter, reporterContext["user"])
			if err != nil {
				return nil, wrapErrorEndSpan(ctx, reporterSpan, "failed to get task reporter template (3)", err)
			}
		}
	}

	systemReporterTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeReporter, reporterContext["system"])
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, reporterSpan, "failed to get task system reporter template", err)
	}

	result, err := fp.performTaskResultReporter(ctx, &taskID, nil, systemReporterTmpl, reporterTmpl, tasksInfo.Task.Input)
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, reporterSpan, "failed to perform task result reporter", err)
	}

	reporterSpan.End(
		langfuse.WithEndSpanStatus("success"),
		langfuse.WithEndSpanOutput(result),
	)

	return result, nil
}

func (fp *flowProvider) PrepareAgentChain(ctx context.Context, taskID, subtaskID int64) (int64, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.PrepareAgentChain")
	defer span.End()

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
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

	optAgentType := pconfig.OptionsTypePrimaryAgent
	msgChainType := database.MsgchainTypePrimaryAgent
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

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
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
	executorSpan := observation.Span(
		langfuse.WithStartSpanName(fmt.Sprintf("primary agent for subtask %d: %s", subtaskID, subtask.Title)),
		langfuse.WithStartSpanInput(chain),
		langfuse.WithStartSpanMetadata(langfuse.Metadata{
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
	ctx, _ = executorSpan.Observation(ctx)

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

				opts := []langfuse.SpanEndOption{
					langfuse.WithEndSpanOutput(done.Result),
				}
				defer executorSpan.End(opts...)

				if !done.Success {
					performResult = PerformResultError
					opts = append(opts,
						langfuse.WithEndSpanStatus("done handler: failed"),
						langfuse.WithEndSpanLevel(langfuse.ObservationLevelWarning),
					)
				} else {
					performResult = PerformResultDone
					opts = append(opts,
						langfuse.WithEndSpanStatus("done handler: success"),
					)
				}

				// TODO: here need to call SetResult from SubtaskWorker interface
				subtask, err = fp.db.UpdateSubtaskResult(ctx, database.UpdateSubtaskResultParams{
					Result: done.Result,
					ID:     subtaskID,
				})
				if err != nil {
					opts = append(opts,
						langfuse.WithEndSpanStatus(err.Error()),
						langfuse.WithEndSpanLevel(langfuse.ObservationLevelError),
					)
					loggerFunc.WithError(err).Error("failed to update subtask result")
					return "", fmt.Errorf("failed to update subtask %d result: %w", subtaskID, err)
				}

				// report result to msg log as a final message for the subtask execution
				if fp.msgLog != nil {
					reportMsgID, err := fp.msgLog.PutMsg(
						ctx,
						database.MsglogTypeReport,
						&taskID, &subtaskID, 0,
						"", subtask.Description,
					)
					if err != nil {
						opts = append(opts,
							langfuse.WithEndSpanStatus(err.Error()),
							langfuse.WithEndSpanLevel(langfuse.ObservationLevelError),
						)
						loggerFunc.WithError(err).Error("failed to put report msg")
						return "", fmt.Errorf("failed to put report msg: %w", err)
					}

					err = fp.msgLog.UpdateMsgResult(
						ctx,
						reportMsgID, 0,
						done.Result, database.MsglogResultFormatMarkdown,
					)
					if err != nil {
						opts = append(opts,
							langfuse.WithEndSpanStatus(err.Error()),
							langfuse.WithEndSpanLevel(langfuse.ObservationLevelError),
						)
						loggerFunc.WithError(err).Error("failed to update report msg result")
						return "", fmt.Errorf("failed to update report msg result: %w", err)
					}
				}

			case tools.AskUserToolName:
				performResult = PerformResultWaiting

				var askUser tools.AskUser
				if err := json.Unmarshal(args, &askUser); err != nil {
					loggerFunc.WithError(err).Error("failed to unmarshal ask user result")
					return "", fmt.Errorf("failed to unmarshal ask user result: %w", err)
				}

				executorSpan.End(
					langfuse.WithEndSpanOutput(askUser.Message),
					langfuse.WithEndSpanStatus("ask user handler"),
				)
			}

			return fmt.Sprintf("function %s successfully processed arguments", name), nil
		},
		Summarizer: fp.GetSummarizeResultHandler(&taskID, &subtaskID),
	}

	executor, err := fp.executor.GetPrimaryExecutor(cfg)
	if err != nil {
		return PerformResultError, wrapErrorEndSpan(ctx, executorSpan, "failed to get primary executor", err)
	}

	ctx = tools.PutAgentContext(ctx, database.MsgchainTypePrimaryAgent)
	err = fp.performAgentChain(
		ctx, pconfig.OptionsTypePrimaryAgent, msgChain.ID, &taskID, &subtaskID, chain, executor, fp.summarizer,
	)
	if err != nil {
		return PerformResultError, wrapErrorEndSpan(ctx, executorSpan, "failed to perform primary agent chain", err)
	}

	executorSpan.End()

	return performResult, nil
}

func (fp *flowProvider) PutInputToAgentChain(ctx context.Context, msgChainID int64, input string) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.PutInputToAgentChain")
	defer span.End()

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
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
		"flow_id":      fp.flowID,
		"msg_chain_id": msgChainID,
	})

	return fp.processChain(ctx, msgChainID, logger, func(chain []llms.MessageContent) ([]llms.MessageContent, error) {
		return fp.ensureChainConsistency(chain)
	})
}
