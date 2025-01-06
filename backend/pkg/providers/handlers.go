package providers

import (
	"context"
	"encoding/json"
	"fmt"

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

func (fp *flowProvider) getAskAdviceHandler(ctx context.Context, taskID, subtaskID int64) (tools.ExecutorHandler, error) {
	tasksInfo, err := fp.getTasksInfo(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks info: %w", err)
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

	enricherHandler := func(ctx context.Context, ask tools.AskAdvice) (string, error) {
		enricherContext := map[string]any{
			"EnricherToolName":  tools.EnricherResultToolName,
			"Question":          ask.Question,
			"Code":              ask.Code,
			"Output":            ask.Output,
			"Task":              tasksInfo.Task,
			"Subtask":           subtasksInfo.Subtask,
			"CompletedSubtasks": subtasksInfo.Completed,
			"Lang":              fp.language,
			"ToolPlaceholder":   ToolPlaceholder,
		}

		enricherCtx, observation := obs.Observer.NewObservation(ctx)
		enricherSpan := observation.Span(
			langfuse.WithStartSpanName("enricher agent"),
			langfuse.WithStartSpanInput(ask.Question),
			langfuse.WithStartSpanMetadata(enricherContext),
		)
		enricherCtx, _ = enricherSpan.Observation(enricherCtx)

		enricherTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeEnricher, enricherContext)
		if err != nil {
			return "", wrapErrorEndSpan(enricherCtx, enricherSpan, "failed to get enricher template", err)
		}

		enriches, err := fp.performEnricher(enricherCtx, taskID, subtaskID, enricherTmpl, ask.Question)
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
		adviserContext := map[string]any{
			"Question":          ask.Question,
			"Enriches":          enriches,
			"Code":              ask.Code,
			"Output":            ask.Output,
			"Task":              tasksInfo.Task,
			"Tasks":             tasksInfo.Tasks,
			"Subtask":           subtasksInfo.Subtask,
			"CompletedSubtasks": subtasksInfo.Completed,
			"PlannedSubtasks":   subtasksInfo.Planned,
		}

		adviserCtx, observation := obs.Observer.NewObservation(ctx)
		adviserSpan := observation.Span(
			langfuse.WithStartSpanName("adviser agent"),
			langfuse.WithStartSpanInput(ask.Question),
			langfuse.WithStartSpanMetadata(adviserContext),
		)
		adviserCtx, _ = adviserSpan.Observation(adviserCtx)

		adviserTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeAdviser, adviserContext)
		if err != nil {
			return "", wrapErrorEndSpan(adviserCtx, adviserSpan, "failed to get adviser template", err)
		}

		opt := provider.OptionsTypeAdviser
		msgChainType := database.MsgchainTypeAdviser
		advice, err := fp.performSimpleChain(adviserCtx, &taskID, &subtaskID, opt, msgChainType, adviserTmpl)
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
				"task":    tasksInfo.Task,
				"subtask": subtasksInfo.Subtask,
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

func (fp *flowProvider) getCoderHandler(ctx context.Context, taskID, subtaskID int64) (tools.ExecutorHandler, error) {
	task, err := fp.db.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	completed, err := fp.db.GetTaskCompletedSubtasks(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed subtasks: %w", err)
	}

	subtask, err := fp.db.GetSubtask(ctx, subtaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subtask: %w", err)
	}

	coderHandler := func(ctx context.Context, action tools.CoderAction) (string, error) {
		coderContext := map[string]any{
			"CodeResultToolName": tools.CodeResultToolName,
			"SearchCodeToolName": tools.SearchCodeToolName,
			"StoreCodeToolName":  tools.StoreCodeToolName,
			"DockerImage":        fp.image,
			"Cwd":                docker.WorkFolderPathInContainer,
			"ContainerPorts":     fp.getContainerPortsDescription(),
			"Task":               task,
			"Subtask":            subtask,
			"CompletedSubtasks":  completed,
			"Lang":               fp.language,
			"ToolPlaceholder":    ToolPlaceholder,
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		coderSpan := observation.Span(
			langfuse.WithStartSpanName("coder agent"),
			langfuse.WithStartSpanInput(action.Question),
			langfuse.WithStartSpanMetadata(coderContext),
		)
		ctx, _ = coderSpan.Observation(ctx)

		coderTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeCoder, coderContext)
		if err != nil {
			return "", wrapErrorEndSpan(ctx, coderSpan, "failed to get coder template", err)
		}

		code, err := fp.performCoder(ctx, taskID, subtaskID, coderTmpl, action.Question)
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
				"task":    task,
				"subtask": subtask,
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

func (fp *flowProvider) getInstallerHandler(ctx context.Context, taskID, subtaskID int64) (tools.ExecutorHandler, error) {
	task, err := fp.db.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	completed, err := fp.db.GetTaskCompletedSubtasks(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed subtasks: %w", err)
	}

	subtask, err := fp.db.GetSubtask(ctx, subtaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subtask: %w", err)
	}

	installerHandler := func(ctx context.Context, action tools.MaintenanceAction) (string, error) {
		installerContext := map[string]any{
			"MaintenanceResultToolName": tools.MaintenanceResultToolName,
			"SearchGuideToolName":       tools.SearchGuideToolName,
			"StoreGuideToolName":        tools.StoreGuideToolName,
			"DockerImage":               fp.image,
			"Cwd":                       docker.WorkFolderPathInContainer,
			"ContainerPorts":            fp.getContainerPortsDescription(),
			"Task":                      task,
			"Subtask":                   subtask,
			"CompletedSubtasks":         completed,
			"Lang":                      fp.language,
			"ToolPlaceholder":           ToolPlaceholder,
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		installerSpan := observation.Span(
			langfuse.WithStartSpanName("installer agent"),
			langfuse.WithStartSpanInput(action.Question),
			langfuse.WithStartSpanMetadata(installerContext),
		)
		ctx, _ = installerSpan.Observation(ctx)

		installerTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeInstaller, installerContext)
		if err != nil {
			return "", wrapErrorEndSpan(ctx, installerSpan, "failed to get installer template", err)
		}

		installerResult, err := fp.performInstaller(ctx, taskID, subtaskID, installerTmpl, action.Question)
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
				"task":    task,
				"subtask": subtask,
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

func (fp *flowProvider) getMemoristHandler(ctx context.Context, taskID int64, subtaskID *int64) (tools.ExecutorHandler, error) {
	tasksInfo, err := fp.getTasksInfo(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks info: %w", err)
	}

	subtasksInfo := fp.getSubtasksInfo(taskID, tasksInfo.Subtasks)
	if subtasksInfo.Subtask == nil {
		for i, subtask := range subtasksInfo.Planned {
			if subtaskID != nil && subtask.ID == *subtaskID {
				subtasksInfo.Subtask = &subtask
				subtasksInfo.Planned = append(subtasksInfo.Planned[:i], subtasksInfo.Planned[i+1:]...)
				break
			}
		}
	}

	memoristContext := map[string]any{
		"MemoristResultToolName": tools.MemoristResultToolName,
		"TerminalToolName":       tools.TerminalToolName,
		"FileToolName":           tools.FileToolName,
		"DockerImage":            fp.image,
		"Cwd":                    docker.WorkFolderPathInContainer,
		"ContainerPorts":         fp.getContainerPortsDescription(),
		"Task":                   tasksInfo.Task,
		"Tasks":                  tasksInfo.Tasks,
		"CompletedSubtasks":      subtasksInfo.Completed,
		"Lang":                   fp.language,
		"ToolPlaceholder":        ToolPlaceholder,
	}

	if subtasksInfo.Subtask != nil {
		memoristContext["Subtask"] = subtasksInfo.Subtask
	}

	memoristHandler := func(ctx context.Context, action tools.MemoristAction) (string, error) {
		question := action.Question
		if action.TaskID != nil {
			question += fmt.Sprintf("\nQuestion related to Task ID: %d", *action.TaskID)
		}
		if action.SubtaskID != nil {
			question += fmt.Sprintf("\nQuestion related to Subtask ID: %d", *action.SubtaskID)
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		memoristSpan := observation.Span(
			langfuse.WithStartSpanName("memorist agent"),
			langfuse.WithStartSpanInput(question),
			langfuse.WithStartSpanMetadata(memoristContext),
		)
		ctx, _ = memoristSpan.Observation(ctx)

		memoristTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeMemorist, memoristContext)
		if err != nil {
			return "", wrapErrorEndSpan(ctx, memoristSpan, "failed to get memorist template", err)
		}

		memoristResult, err := fp.performMemorist(ctx, &taskID, subtaskID, memoristTmpl, question)
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
				"task":    tasksInfo.Task,
				"subtask": subtasksInfo.Subtask,
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

func (fp *flowProvider) getPentesterHandler(ctx context.Context, taskID, subtaskID int64) (tools.ExecutorHandler, error) {
	task, err := fp.db.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	completed, err := fp.db.GetTaskCompletedSubtasks(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed subtasks: %w", err)
	}

	subtask, err := fp.db.GetSubtask(ctx, subtaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subtask: %w", err)
	}

	pentesterHandler := func(ctx context.Context, action tools.PentesterAction) (string, error) {
		pentesterContext := map[string]any{
			"HackResultToolName":  tools.HackResultToolName,
			"SearchGuideToolName": tools.SearchGuideToolName,
			"StoreGuideToolName":  tools.StoreGuideToolName,
			"DockerImage":         fp.image,
			"Cwd":                 docker.WorkFolderPathInContainer,
			"ContainerPorts":      fp.getContainerPortsDescription(),
			"Task":                task,
			"Subtask":             subtask,
			"CompletedSubtasks":   completed,
			"Lang":                fp.language,
			"ToolPlaceholder":     ToolPlaceholder,
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		pentesterSpan := observation.Span(
			langfuse.WithStartSpanName("pentester agent"),
			langfuse.WithStartSpanInput(action.Question),
			langfuse.WithStartSpanMetadata(pentesterContext),
		)
		ctx, _ = pentesterSpan.Observation(ctx)

		pentesterTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypePentester, pentesterContext)
		if err != nil {
			return "", wrapErrorEndSpan(ctx, pentesterSpan, "failed to get pentester template", err)
		}

		pentesterResult, err := fp.performPentester(ctx, taskID, subtaskID, pentesterTmpl, action.Question)
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
			langfuse.WithStartSpanInput(action),
			langfuse.WithStartSpanMetadata(langfuse.Metadata{
				"task":    task,
				"subtask": subtask,
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

func (fp *flowProvider) getSubtaskSearcherHandler(ctx context.Context, taskID, subtaskID int64) (tools.ExecutorHandler, error) {
	task, err := fp.db.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	completed, err := fp.db.GetTaskCompletedSubtasks(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed subtasks: %w", err)
	}

	subtask, err := fp.db.GetSubtask(ctx, subtaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subtask: %w", err)
	}

	searcherHandler := func(ctx context.Context, search tools.ComplexSearch) (string, error) {
		searcherContext := map[string]any{
			"SearchResultToolName": tools.SearchResultToolName,
			"SearchAnswerToolName": tools.SearchAnswerToolName,
			"StoreAnswerToolName":  tools.StoreAnswerToolName,
			"Task":                 task,
			"Subtask":              subtask,
			"CompletedSubtasks":    completed,
			"Lang":                 fp.language,
			"ToolPlaceholder":      ToolPlaceholder,
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		searcherSpan := observation.Span(
			langfuse.WithStartSpanName("searcher agent"),
			langfuse.WithStartSpanInput(search.Question),
			langfuse.WithStartSpanMetadata(searcherContext),
		)
		ctx, _ = searcherSpan.Observation(ctx)

		searcherTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeSearcher, searcherContext)
		if err != nil {
			return "", wrapErrorEndSpan(ctx, searcherSpan, "failed to get searcher template", err)
		}

		searcherResult, err := fp.performSearcher(ctx, &taskID, &subtaskID, searcherTmpl, search.Question)
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
				"task":    task,
				"subtask": subtask,
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

func (fp *flowProvider) getTaskSearcherHandler(ctx context.Context, taskID int64) (tools.ExecutorHandler, error) {
	task, err := fp.db.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	completed, err := fp.db.GetTaskCompletedSubtasks(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed subtasks: %w", err)
	}

	searcherHandler := func(ctx context.Context, search tools.ComplexSearch) (string, error) {
		searcherContext := map[string]any{
			"SearchResultToolName": tools.SearchResultToolName,
			"SearchAnswerToolName": tools.SearchAnswerToolName,
			"StoreAnswerToolName":  tools.StoreAnswerToolName,
			"Task":                 task,
			"CompletedSubtasks":    completed,
			"Lang":                 fp.language,
			"ToolPlaceholder":      ToolPlaceholder,
		}

		ctx, observation := obs.Observer.NewObservation(ctx)
		searcherSpan := observation.Span(
			langfuse.WithStartSpanName("searcher agent"),
			langfuse.WithStartSpanInput(search.Question),
			langfuse.WithStartSpanMetadata(searcherContext),
		)
		ctx, _ = searcherSpan.Observation(ctx)

		searcherTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeSearcher, searcherContext)
		if err != nil {
			return "", wrapErrorEndSpan(ctx, searcherSpan, "failed to get searcher template", err)
		}

		searcherResult, err := fp.performSearcher(ctx, &taskID, nil, searcherTmpl, search.Question)
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
				"task": task,
				"lang": fp.language,
				"tool": name,
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

func (fp *flowProvider) getSummarizeResultHandler(taskID, subtaskID *int64) tools.SummarizeHandler {
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

		summarizerTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeSummarizer, map[string]any{
			"Result": result,
		})
		if err != nil {
			return "", wrapErrorEndSpan(ctx, summarizerSpan, "failed to get summarizer template", err)
		}

		// TODO: here need to summarize result by chunks in iterations
		if len(summarizerTmpl) > 2*summarizeLimit {
			summarizerTmpl = summarizerTmpl[:summarizeLimit] +
				"\n\n{TRUNCATED}...\n\n" +
				summarizerTmpl[len(summarizerTmpl)-summarizeLimit:]
		}

		opt := provider.OptionsTypeSimple
		msgChainType := database.MsgchainTypeSummarizer
		summary, err := fp.performSimpleChain(ctx, taskID, subtaskID, opt, msgChainType, summarizerTmpl)
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

	toolCallFixerTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeToolCallFixer, map[string]any{
		"ToolCallName":   funcName,
		"ToolCallArgs":   string(funcArgs),
		"ToolCallSchema": string(funcJsonSchema),
		"ToolCallError":  funcExecErr.Error(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get tool call fixer template: %w", err)
	}

	opt := provider.OptionsTypeSimpleJSON
	msgChainType := database.MsgchainTypeToolCallFixer
	toolCallFixerResult, err := fp.performSimpleChain(ctx, nil, nil, opt, msgChainType, toolCallFixerTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool call fixer result: %w", err)
	}

	toolCallFixerSpan.End(
		langfuse.WithEndSpanStatus("success"),
		langfuse.WithEndSpanOutput(toolCallFixerResult),
	)

	return json.RawMessage(toolCallFixerResult), nil
}
