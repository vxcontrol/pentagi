package providers

import (
	"context"
	"encoding/json"
	"fmt"

	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
)

const ToolPlaceholder = "Always use your function calling functionality, instead of returning a text result."

const TasksNumberLimit = 15

const summarizeLimit = 16 * 1024 // 16 KB

type PerformResult int

const (
	PerformResultError PerformResult = iota
	PerformResultWaiting
	PerformResultDone
)

type FlowProvider interface {
	Type() provider.ProviderType
	Model(opt provider.ProviderOptionsType) string
	Image() string
	Title() string
	Language() string
	Embedder() *embeddings.EmbedderImpl

	SetAgentLogProvider(agentLog tools.AgentLogProvider)

	GetTaskTitle(ctx context.Context, input string) (string, error)
	GenerateSubtasks(ctx context.Context, taskID int64) ([]tools.SubtaskInfo, error)
	RefineSubtasks(ctx context.Context, taskID int64) ([]tools.SubtaskInfo, error)
	GetTaskResult(ctx context.Context, taskID int64) (*tools.TaskResult, error)

	PrepareAgentChain(ctx context.Context, taskID, subtaskID int64) (int64, error)
	PerformAgentChain(ctx context.Context, taskID, subtaskID, msgChainID int64) (PerformResult, error)
	PutInputToAgentChain(ctx context.Context, msgChainID int64, input string) error

	FlowProviderHandlers
}

type FlowProviderHandlers interface {
	GetAskAdviceHandler(ctx context.Context, taskID, subtaskID int64) (tools.ExecutorHandler, error)
	GetCoderHandler(ctx context.Context, taskID, subtaskID int64) (tools.ExecutorHandler, error)
	GetInstallerHandler(ctx context.Context, taskID, subtaskID int64) (tools.ExecutorHandler, error)
	GetMemoristHandler(ctx context.Context, taskID int64, subtaskID *int64) (tools.ExecutorHandler, error)
	GetPentesterHandler(ctx context.Context, taskID, subtaskID int64) (tools.ExecutorHandler, error)
	GetSubtaskSearcherHandler(ctx context.Context, taskID, subtaskID int64) (tools.ExecutorHandler, error)
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

	flowID   int64
	publicIP string

	image    string
	title    string
	language string

	prompter templates.Prompter
	executor tools.FlowToolsExecutor
	agentLog tools.AgentLogProvider

	provider.Provider
}

func (fp *flowProvider) SetAgentLogProvider(agentLog tools.AgentLogProvider) {
	fp.agentLog = agentLog
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
		"Input": input,
		"Lang":  fp.language,
		"N":     150,
	})
	if err != nil {
		return "", wrapErrorEndSpan(ctx, getterSpan, "failed to get flow title template", err)
	}

	title, err := fp.Call(ctx, provider.OptionsTypeSimple, titleTmpl)
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
			"SubtaskListToolName": tools.SubtaskListToolName,
			"SearchToolName":      tools.SearchToolName,
			"DockerImage":         fp.image,
			"Lang":                fp.language,
			"N":                   TasksNumberLimit,
			"ToolPlaceholder":     ToolPlaceholder,
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
			"SubtaskListToolName": tools.SubtaskListToolName,
			"SearchToolName":      tools.SearchToolName,
			"DockerImage":         fp.image,
			"Lang":                fp.language,
			"N":                   max(TasksNumberLimit-len(subtasksInfo.Completed), 0),
			"ToolPlaceholder":     ToolPlaceholder,
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
		return nil, wrapErrorEndSpan(ctx, refinerSpan, "failed to get task subtasks refiner template", err)
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

	msgLogs, err := fp.db.GetTaskMsgLogs(ctx, database.Int64ToNullInt64(&taskID))
	if err != nil {
		logger.WithError(err).Error("failed to get task msg logs")
		return nil, fmt.Errorf("failed to get task %d msg logs: %w", taskID, err)
	}

	reporterContext := map[string]map[string]any{
		"user": {
			"Task":              tasksInfo.Task,
			"Tasks":             tasksInfo.Tasks,
			"CompletedSubtasks": subtasksInfo.Completed,
			"PlannedSubtasks":   subtasksInfo.Planned,
			"MsgLogs":           msgLogs,
		},
		"system": {
			"ReportResultToolName": tools.ReportResultToolName,
			"Lang":                 fp.language,
			"N":                    2000,
			"ToolPlaceholder":      ToolPlaceholder,
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
		return nil, wrapErrorEndSpan(ctx, reporterSpan, "failed to get task reporter template", err)
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
		"task_id":    taskID,
		"subtask_id": subtaskID,
	})

	tasksInfo, err := fp.getTasksInfo(ctx, taskID)
	if err != nil {
		logger.WithError(err).Error("failed to get tasks info")
		return 0, fmt.Errorf("failed to get tasks info: %w", err)
	}

	subtasksInfo := fp.getSubtasksInfo(taskID, tasksInfo.Subtasks)
	if subtasksInfo.Subtask == nil {
		for i, subtask := range subtasksInfo.Planned {
			if subtask.ID == subtaskID {
				subtasksInfo.Subtask = &subtask
				subtasksInfo.Planned = append(subtasksInfo.Planned[:i], subtasksInfo.Planned[i+1:]...)
				break
			}
		}
	}

	systemAgentTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypePrimaryAgent, map[string]any{
		"FinalyToolName":    tools.FinalyToolName,
		"Task":              tasksInfo.Task,
		"Tasks":             tasksInfo.Tasks,
		"Subtask":           subtasksInfo.Subtask,
		"PlannedSubtasks":   subtasksInfo.Planned,
		"CompletedSubtasks": subtasksInfo.Completed,
		"Lang":              fp.language,
		"DockerImage":       fp.image,
		"ToolPlaceholder":   ToolPlaceholder,
	})
	if err != nil {
		logger.WithError(err).Error("failed to get system prompt for primary agent template")
		return 0, fmt.Errorf("failed to get system prompt for primary agent template: %w", err)
	}

	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemAgentTmpl),
		llms.TextParts(llms.ChatMessageTypeHuman, subtasksInfo.Subtask.Description),
	}
	chainBlob, err := json.Marshal(chain)
	if err != nil {
		logger.WithError(err).Error("failed to marshal primary agent msg chain")
		return 0, fmt.Errorf("failed to marshal primary agent msg chain: %w", err)
	}

	msgChain, err := fp.db.CreateMsgChain(ctx, database.CreateMsgChainParams{
		Type:          database.MsgchainTypePrimaryAgent,
		Model:         fp.Model(provider.OptionsTypeAgent),
		ModelProvider: string(fp.Type()),
		Chain:         chainBlob,
		FlowID:        fp.flowID,
		TaskID:        database.Int64ToNullInt64(&taskID),
		SubtaskID:     database.Int64ToNullInt64(&subtaskID),
	})
	if err != nil {
		logger.WithError(err).Error("failed to create primary agent msg chain")
		return 0, fmt.Errorf("failed to create primary agent msg chain: %w", err)
	}

	return msgChain.ID, nil
}

func (fp *flowProvider) PerformAgentChain(ctx context.Context, taskID, subtaskID, msgChainID int64) (PerformResult, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.PerformAgentChain")
	defer span.End()

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
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

	adviser, err := fp.GetAskAdviceHandler(ctx, taskID, subtaskID)
	if err != nil {
		logger.WithError(err).Error("failed to get ask advice handler")
		return PerformResultError, fmt.Errorf("failed to get ask advice handler: %w", err)
	}

	coder, err := fp.GetCoderHandler(ctx, taskID, subtaskID)
	if err != nil {
		logger.WithError(err).Error("failed to get coder handler")
		return PerformResultError, fmt.Errorf("failed to get coder handler: %w", err)
	}

	installer, err := fp.GetInstallerHandler(ctx, taskID, subtaskID)
	if err != nil {
		logger.WithError(err).Error("failed to get installer handler")
		return PerformResultError, fmt.Errorf("failed to get installer handler: %w", err)
	}

	memorist, err := fp.GetMemoristHandler(ctx, taskID, &subtaskID)
	if err != nil {
		logger.WithError(err).Error("failed to get memorist handler")
		return PerformResultError, fmt.Errorf("failed to get memorist handler: %w", err)
	}

	pentester, err := fp.GetPentesterHandler(ctx, taskID, subtaskID)
	if err != nil {
		logger.WithError(err).Error("failed to get pentester handler")
		return PerformResultError, fmt.Errorf("failed to get pentester handler: %w", err)
	}

	searcher, err := fp.GetSubtaskSearcherHandler(ctx, taskID, subtaskID)
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
				_, err = fp.db.UpdateSubtaskResult(ctx, database.UpdateSubtaskResultParams{
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
	chain, err = fp.performAgentChain(ctx, provider.OptionsTypeAgent, msgChain.ID, &taskID, &subtaskID, chain, executor)
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
		"msg_chain_id": msgChainID,
		"input":        input[:min(len(input), 1000)],
	})

	msgChain, err := fp.db.GetMsgChain(ctx, msgChainID)
	if err != nil {
		logger.WithError(err).Error("failed to get primary agent msg chain")
		return fmt.Errorf("failed to get primary agent msg chain %d: %w", msgChainID, err)
	}

	var chain []llms.MessageContent
	if err := json.Unmarshal(msgChain.Chain, &chain); err != nil {
		logger.WithError(err).Error("failed to unmarshal primary agent msg chain")
		return fmt.Errorf("failed to unmarshal primary agent msg chain %d: %w", msgChainID, err)
	}

	// replace user's dummy input in the last function call response if it's possible
	// otherwise just append it to the end of the chain as a new human message
	if err = fp.updateMsgChainResult(chain, tools.AskUserToolName, input); err != nil {
		chain = append(chain, llms.TextParts(llms.ChatMessageTypeHuman, input))
	}

	chainBlob, err := json.Marshal(chain)
	if err != nil {
		logger.WithError(err).Error("failed to marshal primary agent msg chain")
		return fmt.Errorf("failed to marshal primary agent msg chain %d: %w", msgChainID, err)
	}

	_, err = fp.db.UpdateMsgChain(ctx, database.UpdateMsgChainParams{
		Chain: chainBlob,
		ID:    msgChainID,
	})
	if err != nil {
		logger.WithError(err).Error("failed to update primary agent msg chain")
		return fmt.Errorf("failed to update primary agent msg chain %d: %w", msgChainID, err)
	}

	return nil
}
