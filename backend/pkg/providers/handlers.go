package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"pentagi/pkg/cast"
	"pentagi/pkg/csum"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/schema"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
)

func wrapErrorEndSpan(ctx context.Context, span langfuse.Span, msg string, err error) error {
	logrus.WithContext(ctx).WithError(err).Error(msg)
	err = fmt.Errorf("%s: %w", msg, err)
	span.End(
		langfuse.WithEndSpanStatus(err.Error()),
		langfuse.WithEndSpanLevel(langfuse.ObservationLevelError),
	)
	return err
}

func (fp *flowProvider) GetAskAdviceHandler(ctx context.Context, taskID, subtaskID *int64) (tools.ExecutorHandler, error) {
	var (
		err        error
		ptrTask    *database.Task
		ptrSubtask *database.Subtask
	)

	if taskID != nil {
		task, err := fp.db.GetTask(ctx, *taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task: %w", err)
		}
		ptrTask = &task
	}
	if subtaskID != nil {
		subtask, err := fp.db.GetSubtask(ctx, *subtaskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get subtask: %w", err)
		}
		ptrSubtask = &subtask
	}

	executionContext, err := fp.getExecutionContext(ctx, taskID, subtaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution context: %w", err)
	}

	enricherHandler := func(ctx context.Context, ask tools.AskAdvice) (string, error) {
		enricherContext := map[string]map[string]any{
			"user": {
				"Question": ask.Question,
				"Code":     ask.Code,
				"Output":   ask.Output,
			},
			"system": {
				"EnricherToolName":        tools.EnricherResultToolName,
				"SummarizationToolName":   cast.SummarizationToolName,
				"SummarizedContentPrefix": strings.ReplaceAll(csum.SummarizedContentPrefix, "\n", "\\n"),
				"ExecutionContext":        executionContext,
				"Lang":                    fp.language,
				"CurrentTime":             getCurrentTime(),
				"ToolPlaceholder":         ToolPlaceholder,
			},
		}

		enricherCtx, observation := obs.Observer.NewObservation(ctx)
		enricherSpan := observation.Span(
			langfuse.WithStartSpanName("enricher agent"),
			langfuse.WithStartSpanInput(ask.Question),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"user_context":   enricherContext["user"],
				"system_context": enricherContext["system"],
			}),
		)
		enricherCtx, _ = enricherSpan.Observation(enricherCtx)

		userEnricherTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeQuestionEnricher, enricherContext["user"])
		if err != nil {
			return "", wrapErrorEndSpan(enricherCtx, enricherSpan, "failed to get user enricher template", err)
		}

		systemEnricherTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeEnricher, enricherContext["system"])
		if err != nil {
			return "", wrapErrorEndSpan(enricherCtx, enricherSpan, "failed to get system enricher template", err)
		}

		enriches, err := fp.performEnricher(enricherCtx, taskID, subtaskID, systemEnricherTmpl, userEnricherTmpl, ask.Question)
		if err != nil {
			return "", wrapErrorEndSpan(enricherCtx, enricherSpan, "failed to get enriches for the question", err)
		}

		enricherSpan.End(
			langfuse.WithEndSpanStatus("success"),
			langfuse.WithEndSpanOutput(enriches),
		)

		return enriches, nil
	}

	adviserHandler := func(ctx context.Context, ask tools.AskAdvice, enriches string) (string, error) {
		adviserContext := map[string]map[string]any{
			"user": {
				"Question": ask.Question,
				"Code":     ask.Code,
				"Output":   ask.Output,
				"Enriches": enriches,
			},
			"system": {
				"ExecutionContext": executionContext,
				"CurrentTime":      getCurrentTime(),
			},
		}

		adviserCtx, observation := obs.Observer.NewObservation(ctx)
		adviserSpan := observation.Span(
			langfuse.WithStartSpanName("adviser agent"),
			langfuse.WithStartSpanInput(ask.Question),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"user_context":   adviserContext["user"],
				"system_context": adviserContext["system"],
			}),
		)
		adviserCtx, _ = adviserSpan.Observation(adviserCtx)

		userAdviserTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeQuestionAdviser, adviserContext["user"])
		if err != nil {
			return "", wrapErrorEndSpan(adviserCtx, adviserSpan, "failed to get user adviser template", err)
		}

		systemAdviserTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeAdviser, adviserContext["system"])
		if err != nil {
			return "", wrapErrorEndSpan(adviserCtx, adviserSpan, "failed to get system adviser template", err)
		}

		opt := provider.OptionsTypeAdviser
		msgChainType := database.MsgchainTypeAdviser
		advice, err := fp.performSimpleChain(adviserCtx, taskID, subtaskID, opt, msgChainType, systemAdviserTmpl, userAdviserTmpl)
		if err != nil {
			return "", wrapErrorEndSpan(adviserCtx, adviserSpan, "failed to get advice", err)
		}

		adviserSpan.End(
			langfuse.WithEndSpanStatus("success"),
			langfuse.WithEndSpanOutput(advice),
		)

		return advice, nil
	}

	return func(ctx context.Context, name string, args json.RawMessage) (string, error) {
		ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.getAskAdviceHandler")
		defer span.End()

		var ask tools.AskAdvice
		if err := json.Unmarshal(args, &ask); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to unmarshal ask advice payload")
			return "", fmt.Errorf("failed to unmarshal ask advice payload: %w", err)
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		handlerSpan := observation.Span(
			langfuse.WithStartSpanName("adviser handler"),
			langfuse.WithStartSpanInput(ask),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"task":    ptrTask,
				"subtask": ptrSubtask,
				"context": executionContext,
				"lang":    fp.language,
				"tool":    name,
			}),
		)
		ctx, _ = handlerSpan.Observation(ctx)

		enriches, err := enricherHandler(ctx, ask)
		if err != nil {
			handlerSpan.End(
				langfuse.WithEndSpanStatus(err.Error()),
				langfuse.WithEndSpanLevel(langfuse.ObservationLevelError),
			)
			return "", err
		}

		advice, err := adviserHandler(ctx, ask, enriches)
		if err != nil {
			handlerSpan.End(
				langfuse.WithEndSpanStatus(err.Error()),
				langfuse.WithEndSpanLevel(langfuse.ObservationLevelError),
			)
			return "", err
		}

		handlerSpan.End(
			langfuse.WithEndSpanStatus("success"),
			langfuse.WithEndSpanOutput(advice),
		)

		return advice, nil
	}, nil
}

func (fp *flowProvider) GetCoderHandler(ctx context.Context, taskID, subtaskID *int64) (tools.ExecutorHandler, error) {
	var (
		err        error
		ptrTask    *database.Task
		ptrSubtask *database.Subtask
	)

	if taskID != nil {
		task, err := fp.db.GetTask(ctx, *taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task: %w", err)
		}
		ptrTask = &task
	}

	if subtaskID != nil {
		subtask, err := fp.db.GetSubtask(ctx, *subtaskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get subtask: %w", err)
		}
		ptrSubtask = &subtask
	}

	executionContext, err := fp.getExecutionContext(ctx, taskID, subtaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution context: %w", err)
	}

	coderHandler := func(ctx context.Context, action tools.CoderAction) (string, error) {
		coderContext := map[string]map[string]any{
			"user": {
				"Question": action.Question,
			},
			"system": {
				"CodeResultToolName":      tools.CodeResultToolName,
				"SearchCodeToolName":      tools.SearchCodeToolName,
				"StoreCodeToolName":       tools.StoreCodeToolName,
				"SearchToolName":          tools.SearchToolName,
				"AdviceToolName":          tools.AdviceToolName,
				"MemoristToolName":        tools.MemoristToolName,
				"MaintenanceToolName":     tools.MaintenanceToolName,
				"SummarizationToolName":   cast.SummarizationToolName,
				"SummarizedContentPrefix": strings.ReplaceAll(csum.SummarizedContentPrefix, "\n", "\\n"),
				"DockerImage":             fp.image,
				"Cwd":                     docker.WorkFolderPathInContainer,
				"ContainerPorts":          fp.getContainerPortsDescription(),
				"ExecutionContext":        executionContext,
				"Lang":                    fp.language,
				"CurrentTime":             getCurrentTime(),
				"ToolPlaceholder":         ToolPlaceholder,
			},
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		coderSpan := observation.Span(
			langfuse.WithStartSpanName("coder agent"),
			langfuse.WithStartSpanInput(action.Question),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"user_context":   coderContext["user"],
				"system_context": coderContext["system"],
			}),
		)
		ctx, _ = coderSpan.Observation(ctx)

		userCoderTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeQuestionCoder, coderContext["user"])
		if err != nil {
			return "", wrapErrorEndSpan(ctx, coderSpan, "failed to get user coder template", err)
		}

		systemCoderTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeCoder, coderContext["system"])
		if err != nil {
			return "", wrapErrorEndSpan(ctx, coderSpan, "failed to get system coder template", err)
		}

		code, err := fp.performCoder(ctx, taskID, subtaskID, systemCoderTmpl, userCoderTmpl, action.Question)
		if err != nil {
			return "", wrapErrorEndSpan(ctx, coderSpan, "failed to get coder result", err)
		}

		coderSpan.End(
			langfuse.WithEndSpanStatus("success"),
			langfuse.WithEndSpanOutput(code),
		)

		return code, nil
	}

	return func(ctx context.Context, name string, args json.RawMessage) (string, error) {
		ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.getCoderHandler")
		defer span.End()

		var action tools.CoderAction
		if err := json.Unmarshal(args, &action); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to unmarshal code payload")
			return "", fmt.Errorf("failed to unmarshal code payload: %w", err)
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		handlerSpan := observation.Span(
			langfuse.WithStartSpanName("coder handler"),
			langfuse.WithStartSpanInput(action),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"task":    ptrTask,
				"subtask": ptrSubtask,
				"context": executionContext,
				"lang":    fp.language,
				"tool":    name,
			}),
		)
		ctx, _ = handlerSpan.Observation(ctx)

		coderResult, err := coderHandler(ctx, action)
		if err != nil {
			handlerSpan.End(
				langfuse.WithEndSpanStatus(err.Error()),
				langfuse.WithEndSpanLevel(langfuse.ObservationLevelError),
			)
			return "", err
		}

		handlerSpan.End(
			langfuse.WithEndSpanStatus("success"),
			langfuse.WithEndSpanOutput(coderResult),
		)

		return coderResult, nil
	}, nil
}

func (fp *flowProvider) GetInstallerHandler(ctx context.Context, taskID, subtaskID *int64) (tools.ExecutorHandler, error) {
	var (
		err        error
		ptrTask    *database.Task
		ptrSubtask *database.Subtask
	)

	if taskID != nil {
		task, err := fp.db.GetTask(ctx, *taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task: %w", err)
		}
		ptrTask = &task
	}

	if subtaskID != nil {
		subtask, err := fp.db.GetSubtask(ctx, *subtaskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get subtask: %w", err)
		}
		ptrSubtask = &subtask
	}

	executionContext, err := fp.getExecutionContext(ctx, taskID, subtaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution context: %w", err)
	}

	installerHandler := func(ctx context.Context, action tools.MaintenanceAction) (string, error) {
		installerContext := map[string]map[string]any{
			"user": {
				"Question": action.Question,
			},
			"system": {
				"MaintenanceResultToolName": tools.MaintenanceResultToolName,
				"SearchGuideToolName":       tools.SearchGuideToolName,
				"StoreGuideToolName":        tools.StoreGuideToolName,
				"SearchToolName":            tools.SearchToolName,
				"AdviceToolName":            tools.AdviceToolName,
				"MemoristToolName":          tools.MemoristToolName,
				"SummarizationToolName":     cast.SummarizationToolName,
				"SummarizedContentPrefix":   strings.ReplaceAll(csum.SummarizedContentPrefix, "\n", "\\n"),
				"DockerImage":               fp.image,
				"Cwd":                       docker.WorkFolderPathInContainer,
				"ContainerPorts":            fp.getContainerPortsDescription(),
				"ExecutionContext":          executionContext,
				"Lang":                      fp.language,
				"CurrentTime":               getCurrentTime(),
				"ToolPlaceholder":           ToolPlaceholder,
			},
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		installerSpan := observation.Span(
			langfuse.WithStartSpanName("installer agent"),
			langfuse.WithStartSpanInput(action.Question),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"user_context":   installerContext["user"],
				"system_context": installerContext["system"],
			}),
		)
		ctx, _ = installerSpan.Observation(ctx)

		userInstallerTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeQuestionInstaller, installerContext["user"])
		if err != nil {
			return "", wrapErrorEndSpan(ctx, installerSpan, "failed to get user installer template", err)
		}

		systemInstallerTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeInstaller, installerContext["system"])
		if err != nil {
			return "", wrapErrorEndSpan(ctx, installerSpan, "failed to get system installer template", err)
		}

		installerResult, err := fp.performInstaller(ctx, taskID, subtaskID, systemInstallerTmpl, userInstallerTmpl, action.Question)
		if err != nil {
			return "", wrapErrorEndSpan(ctx, installerSpan, "failed to get installer result", err)
		}

		installerSpan.End(
			langfuse.WithEndSpanStatus("success"),
			langfuse.WithEndSpanOutput(installerResult),
		)

		return installerResult, nil
	}

	return func(ctx context.Context, name string, args json.RawMessage) (string, error) {
		ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.getInstallerHandler")
		defer span.End()

		var action tools.MaintenanceAction
		if err := json.Unmarshal(args, &action); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to unmarshal installer payload")
			return "", fmt.Errorf("failed to unmarshal installer payload: %w", err)
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		handlerSpan := observation.Span(
			langfuse.WithStartSpanName("installer handler"),
			langfuse.WithStartSpanInput(action),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"task":    ptrTask,
				"subtask": ptrSubtask,
				"context": executionContext,
				"lang":    fp.language,
				"tool":    name,
			}),
		)
		ctx, _ = handlerSpan.Observation(ctx)

		installerResult, err := installerHandler(ctx, action)
		if err != nil {
			handlerSpan.End(
				langfuse.WithEndSpanStatus(err.Error()),
				langfuse.WithEndSpanLevel(langfuse.ObservationLevelError),
			)
			return "", err
		}

		handlerSpan.End(
			langfuse.WithEndSpanStatus("success"),
			langfuse.WithEndSpanOutput(installerResult),
		)

		return installerResult, nil
	}, nil
}

func (fp *flowProvider) GetMemoristHandler(ctx context.Context, taskID, subtaskID *int64) (tools.ExecutorHandler, error) {
	var (
		err        error
		ptrTask    *database.Task
		ptrSubtask *database.Subtask
	)

	if taskID != nil {
		task, err := fp.db.GetTask(ctx, *taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task: %w", err)
		}
		ptrTask = &task
	}

	if subtaskID != nil {
		subtask, err := fp.db.GetSubtask(ctx, *subtaskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get subtask: %w", err)
		}
		ptrSubtask = &subtask
	}

	executionContext, err := fp.getExecutionContext(ctx, taskID, subtaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution context: %w", err)
	}

	memoristHandler := func(ctx context.Context, action tools.MemoristAction) (string, error) {
		executionDetails := ""

		var requestedTask *database.Task
		if action.TaskID != nil && taskID != nil && action.TaskID.Int64() == *taskID {
			executionDetails += fmt.Sprintf("user requested current task '%d'\n", *taskID)
		} else if action.TaskID != nil {
			taskID := action.TaskID.Int64()
			t, err := fp.db.GetFlowTask(ctx, database.GetFlowTaskParams{
				ID:     taskID,
				FlowID: fp.flowID,
			})
			if err != nil {
				executionDetails += fmt.Sprintf("failed to get requested task '%d': %s\n", taskID, err)
			}
			requestedTask = &t
		} else {
			executionDetails += fmt.Sprintf("user no specified task, using current task '%d'\n", taskID)
		}

		var requestedSubtask *database.Subtask
		if action.SubtaskID != nil && subtaskID != nil && action.SubtaskID.Int64() == *subtaskID {
			executionDetails += fmt.Sprintf("user requested current subtask '%d'\n", *subtaskID)
		} else if action.SubtaskID != nil {
			subtaskID := action.SubtaskID.Int64()
			st, err := fp.db.GetFlowSubtask(ctx, database.GetFlowSubtaskParams{
				ID:     subtaskID,
				FlowID: fp.flowID,
			})
			if err != nil {
				executionDetails += fmt.Sprintf("failed to get requested subtask '%d': %s\n", subtaskID, err)
			}
			requestedSubtask = &st
		} else if subtaskID != nil {
			executionDetails += fmt.Sprintf("user no specified subtask, using current subtask '%d'\n", *subtaskID)
		} else {
			executionDetails += "user no specified subtask, using all subtasks related to the task\n"
		}

		memoristContext := map[string]map[string]any{
			"user": {
				"Question":         action.Question,
				"Task":             requestedTask,
				"Subtask":          requestedSubtask,
				"ExecutionDetails": executionDetails,
			},
			"system": {
				"MemoristResultToolName":  tools.MemoristResultToolName,
				"TerminalToolName":        tools.TerminalToolName,
				"FileToolName":            tools.FileToolName,
				"SummarizationToolName":   cast.SummarizationToolName,
				"SummarizedContentPrefix": strings.ReplaceAll(csum.SummarizedContentPrefix, "\n", "\\n"),
				"DockerImage":             fp.image,
				"Cwd":                     docker.WorkFolderPathInContainer,
				"ContainerPorts":          fp.getContainerPortsDescription(),
				"ExecutionContext":        executionContext,
				"Lang":                    fp.language,
				"CurrentTime":             getCurrentTime(),
				"ToolPlaceholder":         ToolPlaceholder,
			},
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		memoristSpan := observation.Span(
			langfuse.WithStartSpanName("memorist agent"),
			langfuse.WithStartSpanInput(action.Question),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"user_context":      memoristContext["user"],
				"system_context":    memoristContext["system"],
				"requested_task":    requestedTask,
				"requested_subtask": requestedSubtask,
				"execution_details": executionDetails,
			}),
		)
		ctx, _ = memoristSpan.Observation(ctx)

		userMemoristTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeQuestionMemorist, memoristContext["user"])
		if err != nil {
			return "", wrapErrorEndSpan(ctx, memoristSpan, "failed to get user memorist template", err)
		}

		systemMemoristTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeMemorist, memoristContext["system"])
		if err != nil {
			return "", wrapErrorEndSpan(ctx, memoristSpan, "failed to get system memorist template", err)
		}

		memoristResult, err := fp.performMemorist(ctx, taskID, subtaskID, systemMemoristTmpl, userMemoristTmpl, action.Question)
		if err != nil {
			return "", wrapErrorEndSpan(ctx, memoristSpan, "failed to get memorist result", err)
		}

		memoristSpan.End(
			langfuse.WithEndSpanStatus("success"),
			langfuse.WithEndSpanOutput(memoristResult),
		)

		return memoristResult, nil
	}

	return func(ctx context.Context, name string, args json.RawMessage) (string, error) {
		ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.getMemoristHandler")
		defer span.End()

		var action tools.MemoristAction
		if err := json.Unmarshal(args, &action); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to unmarshal memorist payload")
			return "", fmt.Errorf("failed to unmarshal memorist payload: %w", err)
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		handlerSpan := observation.Span(
			langfuse.WithStartSpanName("memorist handler"),
			langfuse.WithStartSpanInput(action),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"task":    ptrTask,
				"subtask": ptrSubtask,
				"context": executionContext,
				"lang":    fp.language,
				"tool":    name,
			}),
		)
		ctx, _ = handlerSpan.Observation(ctx)

		memoristResult, err := memoristHandler(ctx, action)
		if err != nil {
			handlerSpan.End(
				langfuse.WithEndSpanStatus(err.Error()),
				langfuse.WithEndSpanLevel(langfuse.ObservationLevelError),
			)
			return "", err
		}

		handlerSpan.End(
			langfuse.WithEndSpanStatus("success"),
			langfuse.WithEndSpanOutput(memoristResult),
		)

		return memoristResult, nil
	}, nil
}

func (fp *flowProvider) GetPentesterHandler(ctx context.Context, taskID, subtaskID *int64) (tools.ExecutorHandler, error) {
	var (
		err        error
		ptrTask    *database.Task
		ptrSubtask *database.Subtask
	)

	if taskID != nil {
		task, err := fp.db.GetTask(ctx, *taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task: %w", err)
		}
		ptrTask = &task
	}

	if subtaskID != nil {
		subtask, err := fp.db.GetSubtask(ctx, *subtaskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get subtask: %w", err)
		}
		ptrSubtask = &subtask
	}

	executionContext, err := fp.getExecutionContext(ctx, taskID, subtaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution context: %w", err)
	}

	pentesterHandler := func(ctx context.Context, action tools.PentesterAction) (string, error) {
		pentesterContext := map[string]map[string]any{
			"user": {
				"Question": action.Question,
			},
			"system": {
				"HackResultToolName":      tools.HackResultToolName,
				"SearchGuideToolName":     tools.SearchGuideToolName,
				"StoreGuideToolName":      tools.StoreGuideToolName,
				"SearchToolName":          tools.SearchToolName,
				"CoderToolName":           tools.CoderToolName,
				"AdviceToolName":          tools.AdviceToolName,
				"MemoristToolName":        tools.MemoristToolName,
				"MaintenanceToolName":     tools.MaintenanceToolName,
				"SummarizationToolName":   cast.SummarizationToolName,
				"SummarizedContentPrefix": strings.ReplaceAll(csum.SummarizedContentPrefix, "\n", "\\n"),
				"IsDefaultDockerImage":    strings.HasPrefix(strings.ToLower(fp.image), pentestDockerImage),
				"DockerImage":             fp.image,
				"Cwd":                     docker.WorkFolderPathInContainer,
				"ContainerPorts":          fp.getContainerPortsDescription(),
				"ExecutionContext":        executionContext,
				"Lang":                    fp.language,
				"CurrentTime":             getCurrentTime(),
				"ToolPlaceholder":         ToolPlaceholder,
			},
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		pentesterSpan := observation.Span(
			langfuse.WithStartSpanName("pentester agent"),
			langfuse.WithStartSpanInput(action.Question),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"user_context":   pentesterContext["user"],
				"system_context": pentesterContext["system"],
			}),
		)
		ctx, _ = pentesterSpan.Observation(ctx)

		userPentesterTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeQuestionPentester, pentesterContext["user"])
		if err != nil {
			return "", wrapErrorEndSpan(ctx, pentesterSpan, "failed to get user pentester template", err)
		}

		systemPentesterTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypePentester, pentesterContext["system"])
		if err != nil {
			return "", wrapErrorEndSpan(ctx, pentesterSpan, "failed to get system pentester template", err)
		}

		pentesterResult, err := fp.performPentester(ctx, taskID, subtaskID, systemPentesterTmpl, userPentesterTmpl, action.Question)
		if err != nil {
			return "", wrapErrorEndSpan(ctx, pentesterSpan, "failed to get pentester result", err)
		}

		pentesterSpan.End(
			langfuse.WithEndSpanStatus("success"),
			langfuse.WithEndSpanOutput(pentesterResult),
		)

		return pentesterResult, nil
	}

	return func(ctx context.Context, name string, args json.RawMessage) (string, error) {
		ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.getPentesterHandler")
		defer span.End()

		var action tools.PentesterAction
		if err := json.Unmarshal(args, &action); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to unmarshal pentester payload")
			return "", fmt.Errorf("failed to unmarshal pentester payload: %w", err)
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		handlerSpan := observation.Span(
			langfuse.WithStartSpanName("pentester handler"),
			langfuse.WithStartSpanInput(action.Question),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"task":    ptrTask,
				"subtask": ptrSubtask,
				"context": executionContext,
				"lang":    fp.language,
				"tool":    name,
			}),
		)
		ctx, _ = handlerSpan.Observation(ctx)

		pentesterResult, err := pentesterHandler(ctx, action)
		if err != nil {
			handlerSpan.End(
				langfuse.WithEndSpanStatus(err.Error()),
				langfuse.WithEndSpanLevel(langfuse.ObservationLevelError),
			)
			return "", err
		}

		handlerSpan.End(
			langfuse.WithEndSpanStatus("success"),
			langfuse.WithEndSpanOutput(pentesterResult),
		)

		return pentesterResult, nil
	}, nil
}

func (fp *flowProvider) GetSubtaskSearcherHandler(ctx context.Context, taskID, subtaskID *int64) (tools.ExecutorHandler, error) {
	var (
		err        error
		ptrTask    *database.Task
		ptrSubtask *database.Subtask
	)

	if taskID != nil {
		task, err := fp.db.GetTask(ctx, *taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task: %w", err)
		}
		ptrTask = &task
	}

	if subtaskID != nil {
		subtask, err := fp.db.GetSubtask(ctx, *subtaskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get subtask: %w", err)
		}
		ptrSubtask = &subtask
	}

	executionContext, err := fp.getExecutionContext(ctx, taskID, subtaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution context: %w", err)
	}

	searcherHandler := func(ctx context.Context, search tools.ComplexSearch) (string, error) {
		searcherContext := map[string]map[string]any{
			"user": {
				"Question": search.Question,
				"Task":     ptrTask,
				"Subtask":  ptrSubtask,
			},
			"system": {
				"SearchResultToolName":    tools.SearchResultToolName,
				"SearchAnswerToolName":    tools.SearchAnswerToolName,
				"StoreAnswerToolName":     tools.StoreAnswerToolName,
				"SummarizationToolName":   cast.SummarizationToolName,
				"SummarizedContentPrefix": strings.ReplaceAll(csum.SummarizedContentPrefix, "\n", "\\n"),
				"ExecutionContext":        executionContext,
				"Lang":                    fp.language,
				"CurrentTime":             getCurrentTime(),
				"ToolPlaceholder":         ToolPlaceholder,
			},
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		searcherSpan := observation.Span(
			langfuse.WithStartSpanName("searcher agent"),
			langfuse.WithStartSpanInput(search.Question),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"user_context":   searcherContext["user"],
				"system_context": searcherContext["system"],
			}),
		)
		ctx, _ = searcherSpan.Observation(ctx)

		userSearcherTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeQuestionSearcher, searcherContext["user"])
		if err != nil {
			return "", wrapErrorEndSpan(ctx, searcherSpan, "failed to get user searcher template", err)
		}

		systemSearcherTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeSearcher, searcherContext["system"])
		if err != nil {
			return "", wrapErrorEndSpan(ctx, searcherSpan, "failed to get system searcher template", err)
		}

		searcherResult, err := fp.performSearcher(ctx, taskID, subtaskID, systemSearcherTmpl, userSearcherTmpl, search.Question)
		if err != nil {
			return "", wrapErrorEndSpan(ctx, searcherSpan, "failed to get searcher result", err)
		}

		searcherSpan.End(
			langfuse.WithEndSpanStatus("success"),
			langfuse.WithEndSpanOutput(searcherResult),
		)

		return searcherResult, nil
	}

	return func(ctx context.Context, name string, args json.RawMessage) (string, error) {
		ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.getSubtaskSearcherHandler")
		defer span.End()

		var search tools.ComplexSearch
		if err := json.Unmarshal(args, &search); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to unmarshal search payload")
			return "", fmt.Errorf("failed to unmarshal search payload: %w", err)
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		handlerSpan := observation.Span(
			langfuse.WithStartSpanName("searcher handler"),
			langfuse.WithStartSpanInput(search),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"task":    ptrTask,
				"subtask": ptrSubtask,
				"context": executionContext,
				"lang":    fp.language,
				"tool":    name,
			}),
		)
		ctx, _ = handlerSpan.Observation(ctx)

		searcherResult, err := searcherHandler(ctx, search)
		if err != nil {
			handlerSpan.End(
				langfuse.WithEndSpanStatus(err.Error()),
				langfuse.WithEndSpanLevel(langfuse.ObservationLevelError),
			)
			return "", err
		}

		handlerSpan.End(
			langfuse.WithEndSpanStatus("success"),
			langfuse.WithEndSpanOutput(searcherResult),
		)

		return searcherResult, nil
	}, nil
}

func (fp *flowProvider) GetTaskSearcherHandler(ctx context.Context, taskID int64) (tools.ExecutorHandler, error) {
	task, err := fp.db.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	executionContext, err := fp.getExecutionContext(ctx, &taskID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution context: %w", err)
	}

	searcherHandler := func(ctx context.Context, search tools.ComplexSearch) (string, error) {
		searcherContext := map[string]map[string]any{
			"user": {
				"Question": search.Question,
				"Task":     task,
			},
			"system": {
				"SearchResultToolName":    tools.SearchResultToolName,
				"SearchAnswerToolName":    tools.SearchAnswerToolName,
				"StoreAnswerToolName":     tools.StoreAnswerToolName,
				"SummarizationToolName":   cast.SummarizationToolName,
				"SummarizedContentPrefix": strings.ReplaceAll(csum.SummarizedContentPrefix, "\n", "\\n"),
				"ExecutionContext":        executionContext,
				"Lang":                    fp.language,
				"CurrentTime":             getCurrentTime(),
				"ToolPlaceholder":         ToolPlaceholder,
			},
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		searcherSpan := observation.Span(
			langfuse.WithStartSpanName("searcher agent"),
			langfuse.WithStartSpanInput(search.Question),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"user_context":   searcherContext["user"],
				"system_context": searcherContext["system"],
			}),
		)
		ctx, _ = searcherSpan.Observation(ctx)

		userSearcherTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeQuestionSearcher, searcherContext["user"])
		if err != nil {
			return "", wrapErrorEndSpan(ctx, searcherSpan, "failed to get user searcher template", err)
		}

		systemSearcherTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeSearcher, searcherContext["system"])
		if err != nil {
			return "", wrapErrorEndSpan(ctx, searcherSpan, "failed to get system searcher template", err)
		}

		searcherResult, err := fp.performSearcher(ctx, &taskID, nil, systemSearcherTmpl, userSearcherTmpl, search.Question)
		if err != nil {
			return "", wrapErrorEndSpan(ctx, searcherSpan, "failed to get searcher result", err)
		}

		searcherSpan.End(
			langfuse.WithEndSpanStatus("success"),
			langfuse.WithEndSpanOutput(searcherResult),
		)

		return searcherResult, nil
	}

	return func(ctx context.Context, name string, args json.RawMessage) (string, error) {
		ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.getTaskSearcherHandler")
		defer span.End()

		var search tools.ComplexSearch
		if err := json.Unmarshal(args, &search); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to unmarshal search payload")
			return "", fmt.Errorf("failed to unmarshal search payload: %w", err)
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		handlerSpan := observation.Span(
			langfuse.WithStartSpanName("searcher handler"),
			langfuse.WithStartSpanInput(search),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"task":    task,
				"context": executionContext,
				"lang":    fp.language,
				"tool":    name,
			}),
		)
		ctx, _ = handlerSpan.Observation(ctx)

		searcherResult, err := searcherHandler(ctx, search)
		if err != nil {
			handlerSpan.End(
				langfuse.WithEndSpanStatus(err.Error()),
				langfuse.WithEndSpanLevel(langfuse.ObservationLevelError),
			)
			return "", err
		}

		handlerSpan.End(
			langfuse.WithEndSpanStatus("success"),
			langfuse.WithEndSpanOutput(searcherResult),
		)

		return searcherResult, nil
	}, nil
}

func (fp *flowProvider) GetSummarizeResultHandler(taskID, subtaskID *int64) tools.SummarizeHandler {
	return func(ctx context.Context, result string) (string, error) {
		ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.getSummarizeResultHandler")
		defer span.End()

		ctx, observation := obs.Observer.NewObservation(ctx)
		summarizerSpan := observation.Span(
			langfuse.WithStartSpanName("summarizer agent"),
			langfuse.WithStartSpanInput(result),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"task_id":    taskID,
				"subtask_id": subtaskID,
				"lang":       fp.language,
			}),
		)
		ctx, _ = summarizerSpan.Observation(ctx)

		systemSummarizerTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeSummarizer, map[string]any{
			"TaskID":                  taskID,
			"SubtaskID":               subtaskID,
			"CurrentTime":             getCurrentTime(),
			"SummarizedContentPrefix": strings.ReplaceAll(csum.SummarizedContentPrefix, "\n", "\\n"),
		})
		if err != nil {
			return "", wrapErrorEndSpan(ctx, summarizerSpan, "failed to get summarizer template", err)
		}

		// TODO: here need to summarize result by chunks in iterations
		if len(result) > 2*msgSummarizerLimit {
			result = result[:msgSummarizerLimit] +
				"\n\n{TRUNCATED}...\n\n" +
				result[len(result)-msgSummarizerLimit:]
		}

		opt := provider.OptionsTypeSimple
		msgChainType := database.MsgchainTypeSummarizer
		summary, err := fp.performSimpleChain(ctx, taskID, subtaskID, opt, msgChainType, systemSummarizerTmpl, result)
		if err != nil {
			return "", wrapErrorEndSpan(ctx, summarizerSpan, "failed to get summary", err)
		}

		summarizerSpan.End(
			langfuse.WithEndSpanStatus("success"),
			langfuse.WithEndSpanOutput(summary),
		)

		return summary, nil
	}
}

func (fp *flowProvider) fixToolCallArgs(
	ctx context.Context,
	funcName string,
	funcArgs json.RawMessage,
	funcSchema *schema.Schema,
	funcExecErr error,
) (json.RawMessage, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.fixToolCallArgsHandler")
	defer span.End()

	funcJsonSchema, err := json.Marshal(funcSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool call schema: %w", err)
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	toolCallFixerSpan := observation.Span(
		langfuse.WithStartSpanName("tool call fixer agent"),
		langfuse.WithStartSpanInput(string(funcArgs)),
		langfuse.WithStartSpanMetadata(langfuse.Metadata{
			"func_name":     funcName,
			"func_schema":   string(funcJsonSchema),
			"func_exec_err": funcExecErr.Error(),
		}),
	)
	ctx, _ = toolCallFixerSpan.Observation(ctx)

	userToolCallFixerTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeInputToolCallFixer, map[string]any{
		"ToolCallName":   funcName,
		"ToolCallArgs":   string(funcArgs),
		"ToolCallSchema": string(funcJsonSchema),
		"ToolCallError":  funcExecErr.Error(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user tool call fixer template: %w", err)
	}

	systemToolCallFixerTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeToolCallFixer, map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("failed to get system tool call fixer template: %w", err)
	}

	opt := provider.OptionsTypeSimpleJSON
	msgChainType := database.MsgchainTypeToolCallFixer
	toolCallFixerResult, err := fp.performSimpleChain(ctx, nil, nil, opt, msgChainType, systemToolCallFixerTmpl, userToolCallFixerTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool call fixer result: %w", err)
	}

	toolCallFixerSpan.End(
		langfuse.WithEndSpanStatus("success"),
		langfuse.WithEndSpanOutput(toolCallFixerResult),
	)

	return json.RawMessage(toolCallFixerResult), nil
}
