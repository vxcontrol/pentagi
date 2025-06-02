package converter

import (
	"pentagi/pkg/database"
	"pentagi/pkg/graph/model"
)

func ConvertFlows(flows []database.Flow, containers []database.Container) []*model.Flow {
	containersMap := map[int64][]database.Container{}
	for _, container := range containers {
		containersMap[container.FlowID] = append(containersMap[container.FlowID], container)
	}

	gflows := make([]*model.Flow, 0, len(flows))
	for _, flow := range flows {
		gflows = append(gflows, ConvertFlow(flow, containersMap[flow.ID]))
	}

	return gflows
}

func ConvertFlow(flow database.Flow, containers []database.Container) *model.Flow {
	return &model.Flow{
		ID:        flow.ID,
		Title:     flow.Title,
		Status:    model.StatusType(flow.Status),
		Terminals: ConvertContainers(containers),
		Provider:  flow.ModelProvider,
		CreatedAt: flow.CreatedAt.Time,
		UpdatedAt: flow.UpdatedAt.Time,
	}
}

func ConvertContainers(containers []database.Container) []*model.Terminal {
	gcontainers := make([]*model.Terminal, 0, len(containers))
	for _, container := range containers {
		gcontainers = append(gcontainers, ConvertContainer(container))
	}

	return gcontainers
}

func ConvertContainer(container database.Container) *model.Terminal {
	return &model.Terminal{
		ID:        container.ID,
		Type:      model.TerminalType(container.Type),
		Name:      container.Name,
		Image:     container.Image,
		Connected: container.Status == database.ContainerStatusRunning,
		CreatedAt: container.CreatedAt.Time,
	}
}

func ConvertTasks(tasks []database.Task, subtasks []database.Subtask) []*model.Task {
	subtasksMap := map[int64][]database.Subtask{}
	for _, subtask := range subtasks {
		subtasksMap[subtask.TaskID] = append(subtasksMap[subtask.TaskID], subtask)
	}

	gtasks := make([]*model.Task, 0, len(tasks))
	for _, task := range tasks {
		gtasks = append(gtasks, ConvertTask(task, subtasksMap[task.ID]))
	}

	return gtasks
}

func ConvertSubtasks(subtasks []database.Subtask) []*model.Subtask {
	gsubtasks := make([]*model.Subtask, 0, len(subtasks))
	for _, subtask := range subtasks {
		gsubtasks = append(gsubtasks, ConvertSubtask(subtask))
	}

	return gsubtasks
}

func ConvertTask(task database.Task, subtasks []database.Subtask) *model.Task {
	return &model.Task{
		ID:        task.ID,
		Title:     task.Title,
		Status:    model.StatusType(task.Status),
		Input:     task.Input,
		Result:    task.Result,
		FlowID:    task.FlowID,
		Subtasks:  ConvertSubtasks(subtasks),
		CreatedAt: task.CreatedAt.Time,
		UpdatedAt: task.UpdatedAt.Time,
	}
}

func ConvertSubtask(subtask database.Subtask) *model.Subtask {
	return &model.Subtask{
		ID:          subtask.ID,
		Status:      model.StatusType(subtask.Status),
		Title:       subtask.Title,
		Description: subtask.Description,
		Result:      subtask.Result,
		TaskID:      subtask.TaskID,
		CreatedAt:   subtask.CreatedAt.Time,
		UpdatedAt:   subtask.UpdatedAt.Time,
	}
}

func ConvertFlowAssistant(flow database.Flow, containers []database.Container, assistant database.Assistant) *model.FlowAssistant {
	return &model.FlowAssistant{
		Flow:      ConvertFlow(flow, containers),
		Assistant: ConvertAssistant(assistant),
	}
}

func ConvertAssistants(assistants []database.Assistant) []*model.Assistant {
	gassistants := make([]*model.Assistant, 0, len(assistants))
	for _, assistant := range assistants {
		gassistants = append(gassistants, ConvertAssistant(assistant))
	}

	return gassistants
}

func ConvertAssistant(assistant database.Assistant) *model.Assistant {
	return &model.Assistant{
		ID:        assistant.ID,
		Title:     assistant.Title,
		Status:    model.StatusType(assistant.Status),
		Provider:  assistant.ModelProvider,
		FlowID:    assistant.FlowID,
		UseAgents: assistant.UseAgents,
		CreatedAt: assistant.CreatedAt.Time,
		UpdatedAt: assistant.UpdatedAt.Time,
	}
}

func ConvertScreenshots(screenshots []database.Screenshot) []*model.Screenshot {
	gscreenshots := make([]*model.Screenshot, 0, len(screenshots))
	for _, screenshot := range screenshots {
		gscreenshots = append(gscreenshots, ConvertScreenshot(screenshot))
	}

	return gscreenshots
}

func ConvertScreenshot(screenshot database.Screenshot) *model.Screenshot {
	return &model.Screenshot{
		ID:        screenshot.ID,
		FlowID:    screenshot.FlowID,
		Name:      screenshot.Name,
		URL:       screenshot.Url,
		CreatedAt: screenshot.CreatedAt.Time,
	}
}

func ConvertTerminalLogs(logs []database.Termlog, flowID int64) []*model.TerminalLog {
	glogs := make([]*model.TerminalLog, 0, len(logs))
	for _, log := range logs {
		glogs = append(glogs, ConvertTerminalLog(log, flowID))
	}

	return glogs
}

func ConvertTerminalLog(log database.Termlog, flowID int64) *model.TerminalLog {
	return &model.TerminalLog{
		ID:        log.ID,
		FlowID:    flowID,
		Type:      model.TerminalLogType(log.Type),
		Text:      log.Text,
		Terminal:  log.ContainerID,
		CreatedAt: log.CreatedAt.Time,
	}
}

func ConvertMessageLogs(logs []database.Msglog) []*model.MessageLog {
	glogs := make([]*model.MessageLog, 0, len(logs))
	for _, log := range logs {
		glogs = append(glogs, ConvertMessageLog(log))
	}

	return glogs
}

func ConvertMessageLog(log database.Msglog) *model.MessageLog {
	return &model.MessageLog{
		ID:           log.ID,
		Type:         model.MessageLogType(log.Type),
		Message:      log.Message,
		Thinking:     database.NullStringToPtrString(log.Thinking),
		Result:       log.Result,
		ResultFormat: model.ResultFormat(log.ResultFormat),
		FlowID:       log.FlowID,
		TaskID:       database.NullInt64ToInt64(log.TaskID),
		SubtaskID:    database.NullInt64ToInt64(log.SubtaskID),
		CreatedAt:    log.CreatedAt.Time,
	}
}

func ConvertPrompts(prompts []database.Prompt) []*model.Prompt {
	gprompts := make([]*model.Prompt, 0, len(prompts))
	for _, prompt := range prompts {
		gprompts = append(gprompts, &model.Prompt{
			Type:   prompt.Type,
			Prompt: prompt.Prompt,
		})
	}

	return gprompts
}

func ConvertAgentLogs(logs []database.Agentlog) []*model.AgentLog {
	glogs := make([]*model.AgentLog, 0, len(logs))
	for _, log := range logs {
		glogs = append(glogs, ConvertAgentLog(log))
	}

	return glogs
}

func ConvertAgentLog(log database.Agentlog) *model.AgentLog {
	return &model.AgentLog{
		ID:        log.ID,
		Initiator: model.AgentType(log.Initiator),
		Executor:  model.AgentType(log.Executor),
		Task:      log.Task,
		Result:    log.Result,
		FlowID:    log.FlowID,
		TaskID:    database.NullInt64ToInt64(log.TaskID),
		SubtaskID: database.NullInt64ToInt64(log.SubtaskID),
		CreatedAt: log.CreatedAt.Time,
	}
}

func ConvertSearchLogs(logs []database.Searchlog) []*model.SearchLog {
	glogs := make([]*model.SearchLog, 0, len(logs))
	for _, log := range logs {
		glogs = append(glogs, ConvertSearchLog(log))
	}

	return glogs
}

func ConvertSearchLog(log database.Searchlog) *model.SearchLog {
	return &model.SearchLog{
		ID:        log.ID,
		Initiator: model.AgentType(log.Initiator),
		Executor:  model.AgentType(log.Executor),
		Engine:    string(log.Engine),
		Query:     log.Query,
		Result:    log.Result,
		FlowID:    log.FlowID,
		TaskID:    database.NullInt64ToInt64(log.TaskID),
		SubtaskID: database.NullInt64ToInt64(log.SubtaskID),
		CreatedAt: log.CreatedAt.Time,
	}
}

func ConvertVectorStoreLogs(logs []database.Vecstorelog) []*model.VectorStoreLog {
	glogs := make([]*model.VectorStoreLog, 0, len(logs))
	for _, log := range logs {
		glogs = append(glogs, ConvertVectorStoreLog(log))
	}

	return glogs
}

func ConvertVectorStoreLog(log database.Vecstorelog) *model.VectorStoreLog {
	return &model.VectorStoreLog{
		ID:        log.ID,
		Initiator: model.AgentType(log.Initiator),
		Executor:  model.AgentType(log.Executor),
		Filter:    string(log.Filter),
		Query:     log.Query,
		Action:    model.VectorStoreAction(log.Action),
		Result:    log.Result,
		FlowID:    log.FlowID,
		TaskID:    database.NullInt64ToInt64(log.TaskID),
		SubtaskID: database.NullInt64ToInt64(log.SubtaskID),
		CreatedAt: log.CreatedAt.Time,
	}
}

func ConvertAssistantLogs(logs []database.Assistantlog) []*model.AssistantLog {
	glogs := make([]*model.AssistantLog, 0, len(logs))
	for _, log := range logs {
		glogs = append(glogs, ConvertAssistantLog(log, false))
	}

	return glogs
}

func ConvertAssistantLog(log database.Assistantlog, appendPart bool) *model.AssistantLog {
	return &model.AssistantLog{
		ID:           log.ID,
		Type:         model.MessageLogType(log.Type),
		Message:      log.Message,
		Thinking:     database.NullStringToPtrString(log.Thinking),
		Result:       log.Result,
		ResultFormat: model.ResultFormat(log.ResultFormat),
		AppendPart:   appendPart,
		FlowID:       log.FlowID,
		AssistantID:  log.AssistantID,
		CreatedAt:    log.CreatedAt.Time,
	}
}
