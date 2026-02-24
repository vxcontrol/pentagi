package converter

import (
	"encoding/json"

	"pentagi/pkg/database"
	"pentagi/pkg/graph/model"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/tester"
	"pentagi/pkg/providers/tester/testdata"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/vxcontrol/langchaingo/llms"
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
	provider := &model.Provider{
		Name: flow.ModelProviderName,
		Type: model.ProviderType(flow.ModelProviderType),
	}
	return &model.Flow{
		ID:        flow.ID,
		Title:     flow.Title,
		Status:    model.StatusType(flow.Status),
		Terminals: ConvertContainers(containers),
		Provider:  provider,
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
	provider := &model.Provider{
		Name: assistant.ModelProviderName,
		Type: model.ProviderType(assistant.ModelProviderType),
	}
	return &model.Assistant{
		ID:        assistant.ID,
		Title:     assistant.Title,
		Status:    model.StatusType(assistant.Status),
		Provider:  provider,
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
		TaskID:    database.NullInt64ToInt64(screenshot.TaskID),
		SubtaskID: database.NullInt64ToInt64(screenshot.SubtaskID),
		Name:      screenshot.Name,
		URL:       screenshot.Url,
		CreatedAt: screenshot.CreatedAt.Time,
	}
}

func ConvertTerminalLogs(logs []database.Termlog) []*model.TerminalLog {
	glogs := make([]*model.TerminalLog, 0, len(logs))
	for _, log := range logs {
		glogs = append(glogs, ConvertTerminalLog(log))
	}

	return glogs
}

func ConvertTerminalLog(log database.Termlog) *model.TerminalLog {
	return &model.TerminalLog{
		ID:        log.ID,
		FlowID:    log.FlowID,
		TaskID:    database.NullInt64ToInt64(log.TaskID),
		SubtaskID: database.NullInt64ToInt64(log.SubtaskID),
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

func ConvertPrompts(prompts []database.Prompt) []*model.UserPrompt {
	gprompts := make([]*model.UserPrompt, 0, len(prompts))
	for _, prompt := range prompts {
		gprompts = append(gprompts, &model.UserPrompt{
			ID:        prompt.ID,
			Type:      model.PromptType(prompt.Type),
			Template:  prompt.Prompt,
			CreatedAt: prompt.CreatedAt.Time,
			UpdatedAt: prompt.UpdatedAt.Time,
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

func ConvertDefaultPrompt(prompt *templates.Prompt) *model.DefaultPrompt {
	if prompt == nil {
		return nil
	}

	return &model.DefaultPrompt{
		Type:      model.PromptType(prompt.Type),
		Template:  prompt.Template,
		Variables: prompt.Variables,
	}
}

func ConvertAgentPrompt(prompt *templates.AgentPrompt) *model.AgentPrompt {
	if prompt == nil {
		return nil
	}

	return &model.AgentPrompt{
		System: ConvertDefaultPrompt(&prompt.System),
	}
}

func ConvertAgentPrompts(prompts *templates.AgentPrompts) *model.AgentPrompts {
	if prompts == nil {
		return nil
	}

	return &model.AgentPrompts{
		System: ConvertDefaultPrompt(&prompts.System),
		Human:  ConvertDefaultPrompt(&prompts.Human),
	}
}

func ConvertDefaultPrompts(prompts *templates.DefaultPrompts) *model.DefaultPrompts {
	return &model.DefaultPrompts{
		Agents: &model.AgentsPrompts{
			PrimaryAgent:  ConvertAgentPrompt(&prompts.AgentsPrompts.PrimaryAgent),
			Assistant:     ConvertAgentPrompt(&prompts.AgentsPrompts.Assistant),
			Pentester:     ConvertAgentPrompts(&prompts.AgentsPrompts.Pentester),
			Coder:         ConvertAgentPrompts(&prompts.AgentsPrompts.Coder),
			Installer:     ConvertAgentPrompts(&prompts.AgentsPrompts.Installer),
			Searcher:      ConvertAgentPrompts(&prompts.AgentsPrompts.Searcher),
			Memorist:      ConvertAgentPrompts(&prompts.AgentsPrompts.Memorist),
			Adviser:       ConvertAgentPrompts(&prompts.AgentsPrompts.Adviser),
			Generator:     ConvertAgentPrompts(&prompts.AgentsPrompts.Generator),
			Refiner:       ConvertAgentPrompts(&prompts.AgentsPrompts.Refiner),
			Reporter:      ConvertAgentPrompts(&prompts.AgentsPrompts.Reporter),
			Reflector:     ConvertAgentPrompts(&prompts.AgentsPrompts.Reflector),
			Enricher:      ConvertAgentPrompts(&prompts.AgentsPrompts.Enricher),
			ToolCallFixer: ConvertAgentPrompts(&prompts.AgentsPrompts.ToolCallFixer),
			Summarizer:    ConvertAgentPrompt(&prompts.AgentsPrompts.Summarizer),
		},
		Tools: &model.ToolsPrompts{
			GetFlowDescription:       ConvertDefaultPrompt(&prompts.ToolsPrompts.GetFlowDescription),
			GetTaskDescription:       ConvertDefaultPrompt(&prompts.ToolsPrompts.GetTaskDescription),
			GetExecutionLogs:         ConvertDefaultPrompt(&prompts.ToolsPrompts.GetExecutionLogs),
			GetFullExecutionContext:  ConvertDefaultPrompt(&prompts.ToolsPrompts.GetFullExecutionContext),
			GetShortExecutionContext: ConvertDefaultPrompt(&prompts.ToolsPrompts.GetShortExecutionContext),
			ChooseDockerImage:        ConvertDefaultPrompt(&prompts.ToolsPrompts.ChooseDockerImage),
			ChooseUserLanguage:       ConvertDefaultPrompt(&prompts.ToolsPrompts.ChooseUserLanguage),
			CollectToolCallID:        ConvertDefaultPrompt(&prompts.ToolsPrompts.CollectToolCallID),
			DetectToolCallIDPattern:  ConvertDefaultPrompt(&prompts.ToolsPrompts.DetectToolCallIDPattern),
		},
	}
}

func ConvertPrompt(prompt database.Prompt) *model.UserPrompt {
	return &model.UserPrompt{
		ID:        prompt.ID,
		Type:      model.PromptType(prompt.Type),
		Template:  prompt.Prompt,
		CreatedAt: prompt.CreatedAt.Time,
		UpdatedAt: prompt.UpdatedAt.Time,
	}
}

func ConvertUserPreferences(pref database.UserPreference) *model.UserPreferences {
	var data struct {
		FavoriteFlows []int64 `json:"favoriteFlows"`
	}
	if err := json.Unmarshal(pref.Preferences, &data); err != nil {
		return &model.UserPreferences{
			ID:            pref.UserID,
			FavoriteFlows: []int64{},
		}
	}

	// requires by schema validation
	if data.FavoriteFlows == nil {
		data.FavoriteFlows = []int64{}
	}

	return &model.UserPreferences{
		ID:            pref.UserID,
		FavoriteFlows: data.FavoriteFlows,
	}
}

func ConvertAPIToken(token database.ApiToken) *model.APIToken {
	var name *string
	if token.Name.Valid {
		name = &token.Name.String
	}

	return &model.APIToken{
		ID:        token.ID,
		TokenID:   token.TokenID,
		UserID:    token.UserID,
		RoleID:    token.RoleID,
		Name:      name,
		TTL:       int(token.Ttl),
		Status:    model.TokenStatus(token.Status),
		CreatedAt: token.CreatedAt.Time,
		UpdatedAt: token.UpdatedAt.Time,
	}
}

func ConvertAPITokenRemoveSecret(token database.APITokenWithSecret) *model.APIToken {
	var name *string
	if token.Name.Valid {
		name = &token.Name.String
	}

	return &model.APIToken{
		ID:        token.ID,
		TokenID:   token.TokenID,
		UserID:    token.UserID,
		RoleID:    token.RoleID,
		Name:      name,
		TTL:       int(token.Ttl),
		Status:    model.TokenStatus(token.Status),
		CreatedAt: token.CreatedAt.Time,
		UpdatedAt: token.UpdatedAt.Time,
	}
}

func ConvertAPITokenWithSecret(token database.APITokenWithSecret) *model.APITokenWithSecret {
	var name *string
	if token.Name.Valid {
		name = &token.Name.String
	}

	return &model.APITokenWithSecret{
		ID:        token.ID,
		TokenID:   token.TokenID,
		UserID:    token.UserID,
		RoleID:    token.RoleID,
		Name:      name,
		TTL:       int(token.Ttl),
		Status:    model.TokenStatus(token.Status),
		CreatedAt: token.CreatedAt.Time,
		UpdatedAt: token.UpdatedAt.Time,
		Token:     token.Token,
	}
}

func ConvertAPITokens(tokens []database.ApiToken) []*model.APIToken {
	result := make([]*model.APIToken, 0, len(tokens))
	for _, token := range tokens {
		result = append(result, ConvertAPIToken(token))
	}
	return result
}

func ConvertModels(models pconfig.ModelsConfig) []*model.ModelConfig {
	gmodels := make([]*model.ModelConfig, 0, len(models))
	for _, m := range models {
		modelConfig := &model.ModelConfig{
			Name: m.Name,
		}

		if m.Price != nil {
			modelConfig.Price = &model.ModelPrice{
				Input:      m.Price.Input,
				Output:     m.Price.Output,
				CacheRead:  m.Price.CacheRead,
				CacheWrite: m.Price.CacheWrite,
			}
		}

		if m.Description != nil {
			modelConfig.Description = m.Description
		}
		if m.ReleaseDate != nil {
			modelConfig.ReleaseDate = m.ReleaseDate
		}
		if m.Thinking != nil {
			modelConfig.Thinking = m.Thinking
		}

		gmodels = append(gmodels, modelConfig)
	}

	return gmodels
}

func ConvertProvider(prv database.Provider, cfg *pconfig.ProviderConfig) *model.ProviderConfig {
	return &model.ProviderConfig{
		ID:        prv.ID,
		Name:      prv.Name,
		Type:      model.ProviderType(prv.Type),
		Agents:    ConvertProviderConfigToGqlModel(cfg),
		CreatedAt: prv.CreatedAt.Time,
		UpdatedAt: prv.UpdatedAt.Time,
	}
}

func ConvertProviderConfigToGqlModel(cfg *pconfig.ProviderConfig) *model.AgentsConfig {
	if cfg == nil {
		return nil
	}

	return &model.AgentsConfig{
		Simple:       ConvertAgentConfigToGqlModel(cfg.Simple),
		SimpleJSON:   ConvertAgentConfigToGqlModel(cfg.SimpleJSON),
		PrimaryAgent: ConvertAgentConfigToGqlModel(cfg.PrimaryAgent),
		Assistant:    ConvertAgentConfigToGqlModel(cfg.Assistant),
		Generator:    ConvertAgentConfigToGqlModel(cfg.Generator),
		Refiner:      ConvertAgentConfigToGqlModel(cfg.Refiner),
		Adviser:      ConvertAgentConfigToGqlModel(cfg.Adviser),
		Reflector:    ConvertAgentConfigToGqlModel(cfg.Reflector),
		Searcher:     ConvertAgentConfigToGqlModel(cfg.Searcher),
		Enricher:     ConvertAgentConfigToGqlModel(cfg.Enricher),
		Coder:        ConvertAgentConfigToGqlModel(cfg.Coder),
		Installer:    ConvertAgentConfigToGqlModel(cfg.Installer),
		Pentester:    ConvertAgentConfigToGqlModel(cfg.Pentester),
	}
}

func ConvertAgentConfigToGqlModel(ac *pconfig.AgentConfig) *model.AgentConfig {
	if ac == nil {
		return nil
	}

	result := &model.AgentConfig{
		Model: ac.Model,
	}

	if ac.MaxTokens != 0 {
		result.MaxTokens = &ac.MaxTokens
	}
	if ac.Temperature != 0 {
		result.Temperature = &ac.Temperature
	}
	if ac.TopK != 0 {
		result.TopK = &ac.TopK
	}
	if ac.TopP != 0 {
		result.TopP = &ac.TopP
	}
	if ac.MinLength != 0 {
		result.MinLength = &ac.MinLength
	}
	if ac.MaxLength != 0 {
		result.MaxLength = &ac.MaxLength
	}
	if ac.RepetitionPenalty != 0 {
		result.RepetitionPenalty = &ac.RepetitionPenalty
	}
	if ac.FrequencyPenalty != 0 {
		result.FrequencyPenalty = &ac.FrequencyPenalty
	}
	if ac.PresencePenalty != 0 {
		result.PresencePenalty = &ac.PresencePenalty
	}

	if ac.Reasoning.Effort != llms.ReasoningNone || ac.Reasoning.MaxTokens != 0 {
		reasoning := &model.ReasoningConfig{}

		if ac.Reasoning.Effort != llms.ReasoningNone {
			effort := model.ReasoningEffort(ac.Reasoning.Effort)
			reasoning.Effort = &effort
		}
		if ac.Reasoning.MaxTokens != 0 {
			reasoning.MaxTokens = &ac.Reasoning.MaxTokens
		}

		result.Reasoning = reasoning
	}

	if ac.Price != nil {
		result.Price = &model.ModelPrice{
			Input:      ac.Price.Input,
			Output:     ac.Price.Output,
			CacheRead:  ac.Price.CacheRead,
			CacheWrite: ac.Price.CacheWrite,
		}
	}

	return result
}

func ConvertAgentsConfigFromGqlModel(cfg *model.AgentsConfig) *pconfig.ProviderConfig {
	if cfg == nil {
		return nil
	}

	pc := &pconfig.ProviderConfig{
		Simple:       ConvertAgentConfigFromGqlModel(cfg.Simple),
		SimpleJSON:   ConvertAgentConfigFromGqlModel(cfg.SimpleJSON),
		PrimaryAgent: ConvertAgentConfigFromGqlModel(cfg.PrimaryAgent),
		Assistant:    ConvertAgentConfigFromGqlModel(cfg.Assistant),
		Generator:    ConvertAgentConfigFromGqlModel(cfg.Generator),
		Refiner:      ConvertAgentConfigFromGqlModel(cfg.Refiner),
		Adviser:      ConvertAgentConfigFromGqlModel(cfg.Adviser),
		Reflector:    ConvertAgentConfigFromGqlModel(cfg.Reflector),
		Searcher:     ConvertAgentConfigFromGqlModel(cfg.Searcher),
		Enricher:     ConvertAgentConfigFromGqlModel(cfg.Enricher),
		Coder:        ConvertAgentConfigFromGqlModel(cfg.Coder),
		Installer:    ConvertAgentConfigFromGqlModel(cfg.Installer),
		Pentester:    ConvertAgentConfigFromGqlModel(cfg.Pentester),
	}

	rawConfig, err := json.Marshal(pc)
	if err != nil {
		return nil
	}

	pc.SetRawConfig(rawConfig)

	return pc
}

func ConvertAgentConfigFromGqlModel(ac *model.AgentConfig) *pconfig.AgentConfig {
	if ac == nil {
		return nil
	}

	rawConfig := make(map[string]any)
	rawConfig["model"] = ac.Model

	if ac.MaxTokens != nil {
		rawConfig["max_tokens"] = *ac.MaxTokens
	}
	if ac.Temperature != nil {
		rawConfig["temperature"] = *ac.Temperature
	}
	if ac.TopK != nil {
		rawConfig["top_k"] = *ac.TopK
	}
	if ac.TopP != nil {
		rawConfig["top_p"] = *ac.TopP
	}
	if ac.MinLength != nil {
		rawConfig["min_length"] = *ac.MinLength
	}
	if ac.MaxLength != nil {
		rawConfig["max_length"] = *ac.MaxLength
	}
	if ac.RepetitionPenalty != nil {
		rawConfig["repetition_penalty"] = *ac.RepetitionPenalty
	}
	if ac.FrequencyPenalty != nil {
		rawConfig["frequency_penalty"] = *ac.FrequencyPenalty
	}
	if ac.PresencePenalty != nil {
		rawConfig["presence_penalty"] = *ac.PresencePenalty
	}

	if ac.Reasoning != nil {
		reasoning := map[string]any{}
		if ac.Reasoning.Effort != nil {
			reasoning["effort"] = llms.ReasoningEffort(*ac.Reasoning.Effort)
		}
		if ac.Reasoning.MaxTokens != nil {
			reasoning["max_tokens"] = *ac.Reasoning.MaxTokens
		}
		rawConfig["reasoning"] = reasoning
	}

	if ac.Price != nil {
		rawConfig["price"] = map[string]any{
			"input":       ac.Price.Input,
			"output":      ac.Price.Output,
			"cache_read":  ac.Price.CacheRead,
			"cache_write": ac.Price.CacheWrite,
		}
	}

	jsonConfig, err := json.Marshal(rawConfig)
	if err != nil {
		return nil
	}

	var result pconfig.AgentConfig
	err = json.Unmarshal(jsonConfig, &result)
	if err != nil {
		return nil
	}

	return &result
}

func ConvertTestResult(result testdata.TestResult) *model.TestResult {
	var (
		errString *string
		latency   *int
	)

	if result.Error != nil {
		err := result.Error.Error()
		errString = &err
	}

	if result.Latency != 0 {
		latencyMs := int(result.Latency.Milliseconds())
		latency = &latencyMs
	}

	return &model.TestResult{
		Name:      result.Name,
		Type:      string(result.Type),
		Result:    result.Success,
		Error:     errString,
		Streaming: result.Streaming,
		Reasoning: result.Reasoning,
		Latency:   latency,
	}
}

func ConvertTestResults(results tester.AgentTestResults) *model.AgentTestResult {
	gresults := make([]*model.TestResult, 0, len(results))
	for _, result := range results {
		gresults = append(gresults, ConvertTestResult(result))
	}

	return &model.AgentTestResult{
		Tests: gresults,
	}
}

func ConvertProviderTestResults(results tester.ProviderTestResults) *model.ProviderTestResult {
	return &model.ProviderTestResult{
		Simple:       ConvertTestResults(results.Simple),
		SimpleJSON:   ConvertTestResults(results.SimpleJSON),
		PrimaryAgent: ConvertTestResults(results.PrimaryAgent),
		Assistant:    ConvertTestResults(results.Assistant),
		Generator:    ConvertTestResults(results.Generator),
		Refiner:      ConvertTestResults(results.Refiner),
		Adviser:      ConvertTestResults(results.Adviser),
		Reflector:    ConvertTestResults(results.Reflector),
		Searcher:     ConvertTestResults(results.Searcher),
		Enricher:     ConvertTestResults(results.Enricher),
		Coder:        ConvertTestResults(results.Coder),
		Installer:    ConvertTestResults(results.Installer),
		Pentester:    ConvertTestResults(results.Pentester),
	}
}

// UsageStatsRow constraint for generic conversion
type UsageStatsRow interface {
	database.GetFlowUsageStatsRow |
		database.GetTaskUsageStatsRow |
		database.GetSubtaskUsageStatsRow |
		database.GetUserTotalUsageStatsRow |
		database.GetUsageStatsByDayLastWeekRow |
		database.GetUsageStatsByDayLastMonthRow |
		database.GetUsageStatsByDayLast3MonthsRow |
		database.GetUsageStatsByProviderRow |
		database.GetUsageStatsByModelRow |
		database.GetUsageStatsByTypeRow |
		database.GetUsageStatsByTypeForFlowRow
}

// ToolcallsStatsRow constraint for generic conversion
type ToolcallsStatsRow interface {
	database.GetFlowToolcallsStatsRow |
		database.GetTaskToolcallsStatsRow |
		database.GetSubtaskToolcallsStatsRow |
		database.GetUserTotalToolcallsStatsRow
}

// FlowsStatsRow constraint for generic conversion
type FlowsStatsRow interface {
	database.GetUserTotalFlowsStatsRow
}

// FlowStatsRow constraint for conversion of single flow stats
type FlowStatsRow interface {
	database.GetFlowStatsRow
}

// ConvertUsageStats converts database usage stats to GraphQL model using generics
func ConvertUsageStats[T UsageStatsRow](stats T) *model.UsageStats {
	var in, out, cacheIn, cacheOut int64
	var costIn, costOut float64

	// Extract fields based on type
	switch v := any(stats).(type) {
	case database.GetFlowUsageStatsRow:
		in, out = v.TotalUsageIn, v.TotalUsageOut
		cacheIn, cacheOut = v.TotalUsageCacheIn, v.TotalUsageCacheOut
		costIn, costOut = v.TotalUsageCostIn, v.TotalUsageCostOut
	case database.GetTaskUsageStatsRow:
		in, out = v.TotalUsageIn, v.TotalUsageOut
		cacheIn, cacheOut = v.TotalUsageCacheIn, v.TotalUsageCacheOut
		costIn, costOut = v.TotalUsageCostIn, v.TotalUsageCostOut
	case database.GetSubtaskUsageStatsRow:
		in, out = v.TotalUsageIn, v.TotalUsageOut
		cacheIn, cacheOut = v.TotalUsageCacheIn, v.TotalUsageCacheOut
		costIn, costOut = v.TotalUsageCostIn, v.TotalUsageCostOut
	case database.GetUserTotalUsageStatsRow:
		in, out = v.TotalUsageIn, v.TotalUsageOut
		cacheIn, cacheOut = v.TotalUsageCacheIn, v.TotalUsageCacheOut
		costIn, costOut = v.TotalUsageCostIn, v.TotalUsageCostOut
	case database.GetUsageStatsByDayLastWeekRow:
		in, out = v.TotalUsageIn, v.TotalUsageOut
		cacheIn, cacheOut = v.TotalUsageCacheIn, v.TotalUsageCacheOut
		costIn, costOut = v.TotalUsageCostIn, v.TotalUsageCostOut
	case database.GetUsageStatsByDayLastMonthRow:
		in, out = v.TotalUsageIn, v.TotalUsageOut
		cacheIn, cacheOut = v.TotalUsageCacheIn, v.TotalUsageCacheOut
		costIn, costOut = v.TotalUsageCostIn, v.TotalUsageCostOut
	case database.GetUsageStatsByDayLast3MonthsRow:
		in, out = v.TotalUsageIn, v.TotalUsageOut
		cacheIn, cacheOut = v.TotalUsageCacheIn, v.TotalUsageCacheOut
		costIn, costOut = v.TotalUsageCostIn, v.TotalUsageCostOut
	case database.GetUsageStatsByProviderRow:
		in, out = v.TotalUsageIn, v.TotalUsageOut
		cacheIn, cacheOut = v.TotalUsageCacheIn, v.TotalUsageCacheOut
		costIn, costOut = v.TotalUsageCostIn, v.TotalUsageCostOut
	case database.GetUsageStatsByModelRow:
		in, out = v.TotalUsageIn, v.TotalUsageOut
		cacheIn, cacheOut = v.TotalUsageCacheIn, v.TotalUsageCacheOut
		costIn, costOut = v.TotalUsageCostIn, v.TotalUsageCostOut
	case database.GetUsageStatsByTypeRow:
		in, out = v.TotalUsageIn, v.TotalUsageOut
		cacheIn, cacheOut = v.TotalUsageCacheIn, v.TotalUsageCacheOut
		costIn, costOut = v.TotalUsageCostIn, v.TotalUsageCostOut
	case database.GetUsageStatsByTypeForFlowRow:
		in, out = v.TotalUsageIn, v.TotalUsageOut
		cacheIn, cacheOut = v.TotalUsageCacheIn, v.TotalUsageCacheOut
		costIn, costOut = v.TotalUsageCostIn, v.TotalUsageCostOut
	}

	return &model.UsageStats{
		TotalUsageIn:       int(in),
		TotalUsageOut:      int(out),
		TotalUsageCacheIn:  int(cacheIn),
		TotalUsageCacheOut: int(cacheOut),
		TotalUsageCostIn:   costIn,
		TotalUsageCostOut:  costOut,
	}
}

// ConvertDailyUsageStats converts daily usage stats to GraphQL model
func ConvertDailyUsageStats(stats []database.GetUsageStatsByDayLastWeekRow) []*model.DailyUsageStats {
	result := make([]*model.DailyUsageStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, &model.DailyUsageStats{
			Date:  stat.Date,
			Stats: ConvertUsageStats(stat),
		})
	}
	return result
}

// ConvertDailyUsageStatsMonth converts monthly usage stats to GraphQL model
func ConvertDailyUsageStatsMonth(stats []database.GetUsageStatsByDayLastMonthRow) []*model.DailyUsageStats {
	result := make([]*model.DailyUsageStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, &model.DailyUsageStats{
			Date:  stat.Date,
			Stats: ConvertUsageStats(stat),
		})
	}
	return result
}

// ConvertDailyUsageStatsQuarter converts quarterly usage stats to GraphQL model
func ConvertDailyUsageStatsQuarter(stats []database.GetUsageStatsByDayLast3MonthsRow) []*model.DailyUsageStats {
	result := make([]*model.DailyUsageStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, &model.DailyUsageStats{
			Date:  stat.Date,
			Stats: ConvertUsageStats(stat),
		})
	}
	return result
}

// ConvertProviderUsageStats converts provider usage stats to GraphQL model
func ConvertProviderUsageStats(stats []database.GetUsageStatsByProviderRow) []*model.ProviderUsageStats {
	result := make([]*model.ProviderUsageStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, &model.ProviderUsageStats{
			Provider: stat.ModelProvider,
			Stats:    ConvertUsageStats(stat),
		})
	}
	return result
}

// ConvertModelUsageStats converts model usage stats to GraphQL model
func ConvertModelUsageStats(stats []database.GetUsageStatsByModelRow) []*model.ModelUsageStats {
	result := make([]*model.ModelUsageStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, &model.ModelUsageStats{
			Model:    stat.Model,
			Provider: stat.ModelProvider,
			Stats:    ConvertUsageStats(stat),
		})
	}
	return result
}

// ConvertAgentTypeUsageStats converts agent type usage stats to GraphQL model
func ConvertAgentTypeUsageStats(stats []database.GetUsageStatsByTypeRow) []*model.AgentTypeUsageStats {
	result := make([]*model.AgentTypeUsageStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, &model.AgentTypeUsageStats{
			AgentType: model.AgentType(stat.Type),
			Stats:     ConvertUsageStats(stat),
		})
	}
	return result
}

// ConvertAgentTypeUsageStatsForFlow converts agent type usage stats for flow to GraphQL model
func ConvertAgentTypeUsageStatsForFlow(stats []database.GetUsageStatsByTypeForFlowRow) []*model.AgentTypeUsageStats {
	result := make([]*model.AgentTypeUsageStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, &model.AgentTypeUsageStats{
			AgentType: model.AgentType(stat.Type),
			Stats:     ConvertUsageStats(stat),
		})
	}
	return result
}

// ==================== Toolcalls Statistics Converters ====================

// ConvertToolcallsStats converts database toolcalls stats to GraphQL model using generics
func ConvertToolcallsStats[T ToolcallsStatsRow](stats T) *model.ToolcallsStats {
	var count int64
	var duration float64

	// Extract fields based on type
	switch v := any(stats).(type) {
	case database.GetFlowToolcallsStatsRow:
		count, duration = v.TotalCount, v.TotalDurationSeconds
	case database.GetTaskToolcallsStatsRow:
		count, duration = v.TotalCount, v.TotalDurationSeconds
	case database.GetSubtaskToolcallsStatsRow:
		count, duration = v.TotalCount, v.TotalDurationSeconds
	case database.GetUserTotalToolcallsStatsRow:
		count, duration = v.TotalCount, v.TotalDurationSeconds
	}

	return &model.ToolcallsStats{
		TotalCount:           int(count),
		TotalDurationSeconds: duration,
	}
}

// ConvertDailyToolcallsStats converts daily toolcalls stats to GraphQL model
func ConvertDailyToolcallsStatsWeek(stats []database.GetToolcallsStatsByDayLastWeekRow) []*model.DailyToolcallsStats {
	result := make([]*model.DailyToolcallsStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, &model.DailyToolcallsStats{
			Date: stat.Date,
			Stats: &model.ToolcallsStats{
				TotalCount:           int(stat.TotalCount),
				TotalDurationSeconds: stat.TotalDurationSeconds,
			},
		})
	}
	return result
}

// ConvertDailyToolcallsStatsMonth converts monthly toolcalls stats to GraphQL model
func ConvertDailyToolcallsStatsMonth(stats []database.GetToolcallsStatsByDayLastMonthRow) []*model.DailyToolcallsStats {
	result := make([]*model.DailyToolcallsStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, &model.DailyToolcallsStats{
			Date: stat.Date,
			Stats: &model.ToolcallsStats{
				TotalCount:           int(stat.TotalCount),
				TotalDurationSeconds: stat.TotalDurationSeconds,
			},
		})
	}
	return result
}

// ConvertDailyToolcallsStatsQuarter converts quarterly toolcalls stats to GraphQL model
func ConvertDailyToolcallsStatsQuarter(stats []database.GetToolcallsStatsByDayLast3MonthsRow) []*model.DailyToolcallsStats {
	result := make([]*model.DailyToolcallsStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, &model.DailyToolcallsStats{
			Date: stat.Date,
			Stats: &model.ToolcallsStats{
				TotalCount:           int(stat.TotalCount),
				TotalDurationSeconds: stat.TotalDurationSeconds,
			},
		})
	}
	return result
}

// isAgentTool checks if a function name represents an agent tool
func isAgentTool(functionName string) bool {
	toolTypeMapping := tools.GetToolTypeMapping()
	toolType, exists := toolTypeMapping[functionName]
	if !exists {
		return false
	}
	// Agent tools include AgentToolType and StoreAgentResultToolType
	return toolType == tools.AgentToolType || toolType == tools.StoreAgentResultToolType
}

// ConvertFunctionToolcallsStats converts function toolcalls stats to GraphQL model
func ConvertFunctionToolcallsStats(stats []database.GetToolcallsStatsByFunctionRow) []*model.FunctionToolcallsStats {
	result := make([]*model.FunctionToolcallsStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, &model.FunctionToolcallsStats{
			FunctionName:         stat.FunctionName,
			IsAgent:              isAgentTool(stat.FunctionName),
			TotalCount:           int(stat.TotalCount),
			TotalDurationSeconds: stat.TotalDurationSeconds,
			AvgDurationSeconds:   stat.AvgDurationSeconds,
		})
	}
	return result
}

// ConvertFunctionToolcallsStatsForFlow converts function toolcalls stats for flow to GraphQL model
func ConvertFunctionToolcallsStatsForFlow(stats []database.GetToolcallsStatsByFunctionForFlowRow) []*model.FunctionToolcallsStats {
	result := make([]*model.FunctionToolcallsStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, &model.FunctionToolcallsStats{
			FunctionName:         stat.FunctionName,
			IsAgent:              isAgentTool(stat.FunctionName),
			TotalCount:           int(stat.TotalCount),
			TotalDurationSeconds: stat.TotalDurationSeconds,
			AvgDurationSeconds:   stat.AvgDurationSeconds,
		})
	}
	return result
}

// ==================== Flows Statistics Converters ====================

// ConvertFlowsStats converts database flows stats to GraphQL model using generics
func ConvertFlowsStats[T FlowsStatsRow](stats T) *model.FlowsStats {
	var flowsCount, tasksCount, subtasksCount, assistantsCount int64

	// Extract fields based on type
	switch v := any(stats).(type) {
	case database.GetUserTotalFlowsStatsRow:
		flowsCount, tasksCount, subtasksCount, assistantsCount = v.TotalFlowsCount, v.TotalTasksCount, v.TotalSubtasksCount, v.TotalAssistantsCount
	}

	return &model.FlowsStats{
		TotalFlowsCount:      int(flowsCount),
		TotalTasksCount:      int(tasksCount),
		TotalSubtasksCount:   int(subtasksCount),
		TotalAssistantsCount: int(assistantsCount),
	}
}

// ConvertFlowStats converts database single flow stats to GraphQL model using generics
func ConvertFlowStats[T FlowStatsRow](stats T) *model.FlowStats {
	var tasksCount, subtasksCount, assistantsCount int64

	// Extract fields based on type
	switch v := any(stats).(type) {
	case database.GetFlowStatsRow:
		tasksCount, subtasksCount, assistantsCount = v.TotalTasksCount, v.TotalSubtasksCount, v.TotalAssistantsCount
	}

	return &model.FlowStats{
		TotalTasksCount:      int(tasksCount),
		TotalSubtasksCount:   int(subtasksCount),
		TotalAssistantsCount: int(assistantsCount),
	}
}

// ConvertDailyFlowsStatsWeek converts daily flows stats to GraphQL model
func ConvertDailyFlowsStatsWeek(stats []database.GetFlowsStatsByDayLastWeekRow) []*model.DailyFlowsStats {
	result := make([]*model.DailyFlowsStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, &model.DailyFlowsStats{
			Date: stat.Date,
			Stats: &model.FlowsStats{
				TotalFlowsCount:      int(stat.TotalFlowsCount),
				TotalTasksCount:      int(stat.TotalTasksCount),
				TotalSubtasksCount:   int(stat.TotalSubtasksCount),
				TotalAssistantsCount: int(stat.TotalAssistantsCount),
			},
		})
	}
	return result
}

// ConvertDailyFlowsStatsMonth converts monthly flows stats to GraphQL model
func ConvertDailyFlowsStatsMonth(stats []database.GetFlowsStatsByDayLastMonthRow) []*model.DailyFlowsStats {
	result := make([]*model.DailyFlowsStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, &model.DailyFlowsStats{
			Date: stat.Date,
			Stats: &model.FlowsStats{
				TotalFlowsCount:      int(stat.TotalFlowsCount),
				TotalTasksCount:      int(stat.TotalTasksCount),
				TotalSubtasksCount:   int(stat.TotalSubtasksCount),
				TotalAssistantsCount: int(stat.TotalAssistantsCount),
			},
		})
	}
	return result
}

// ConvertDailyFlowsStatsQuarter converts quarterly flows stats to GraphQL model
func ConvertDailyFlowsStatsQuarter(stats []database.GetFlowsStatsByDayLast3MonthsRow) []*model.DailyFlowsStats {
	result := make([]*model.DailyFlowsStats, 0, len(stats))
	for _, stat := range stats {
		result = append(result, &model.DailyFlowsStats{
			Date: stat.Date,
			Stats: &model.FlowsStats{
				TotalFlowsCount:      int(stat.TotalFlowsCount),
				TotalTasksCount:      int(stat.TotalTasksCount),
				TotalSubtasksCount:   int(stat.TotalSubtasksCount),
				TotalAssistantsCount: int(stat.TotalAssistantsCount),
			},
		})
	}
	return result
}

// ==================== Flows/Tasks/Subtasks Execution Time Converters ====================
