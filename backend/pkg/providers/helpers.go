package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"pentagi/pkg/cast"
	"pentagi/pkg/csum"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
	"github.com/vxcontrol/langchaingo/llms"
)

const (
	RepeatingToolCallThreshold   = 3
	maxQASectionsAfterRestore    = 3
	keepQASectionsAfterRestore   = 1
	lastSecBytesAfterRestore     = 16 * 1024 // 16 KB
	maxBPBytesAfterRestore       = 8 * 1024  // 8 KB
	maxQABytesAfterRestore       = 20 * 1024 // 20 KB
	msgLogResultSummarySizeLimit = 70 * 1024 // 70 KB
	msgLogResultEntrySizeLimit   = 1024      // 1 KB
)

type repeatingDetector struct {
	funcCalls []llms.FunctionCall
}

func (rd *repeatingDetector) detect(toolCall llms.ToolCall) bool {
	if toolCall.FunctionCall == nil {
		return false
	}

	funcCall := rd.clearCallArguments(toolCall.FunctionCall)

	if len(rd.funcCalls) == 0 {
		rd.funcCalls = append(rd.funcCalls, funcCall)
		return false
	}

	lastToolCall := rd.funcCalls[len(rd.funcCalls)-1]
	if lastToolCall.Name != funcCall.Name || lastToolCall.Arguments != funcCall.Arguments {
		rd.funcCalls = []llms.FunctionCall{funcCall}
		return false
	}

	rd.funcCalls = append(rd.funcCalls, funcCall)

	return len(rd.funcCalls) >= RepeatingToolCallThreshold
}

func (rd *repeatingDetector) clearCallArguments(toolCall *llms.FunctionCall) llms.FunctionCall {
	var v map[string]any
	if err := json.Unmarshal([]byte(toolCall.Arguments), &v); err != nil {
		return *toolCall
	}

	delete(v, "message")
	var keys []string
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buffer strings.Builder
	for _, k := range keys {
		buffer.WriteString(fmt.Sprintf("%s: %v\n", k, v[k]))
	}

	return llms.FunctionCall{
		Name:      toolCall.Name,
		Arguments: buffer.String(),
	}
}

func (fp *flowProvider) getTasksInfo(ctx context.Context, taskID int64) (*tasksInfo, error) {
	var (
		err  error
		info tasksInfo
	)

	info.Tasks, err = fp.db.GetFlowTasks(ctx, fp.flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow %d tasks: %w", fp.flowID, err)
	}

	for idx, t := range info.Tasks {
		if t.ID == taskID {
			info.Task = t
			info.Tasks = append(info.Tasks[:idx], info.Tasks[idx+1:]...)
			break
		}
	}

	info.Subtasks, err = fp.db.GetFlowSubtasks(ctx, fp.flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow %d subtasks: %w", fp.flowID, err)
	}

	return &info, nil
}

func (fp *flowProvider) getSubtasksInfo(taskID int64, subtasks []database.Subtask) *subtasksInfo {
	var info subtasksInfo
	for _, subtask := range subtasks {
		if subtask.TaskID != taskID && taskID != 0 {
			continue
		}

		switch subtask.Status {
		case database.SubtaskStatusCreated:
			info.Planned = append(info.Planned, subtask)
		case database.SubtaskStatusFinished, database.SubtaskStatusFailed:
			info.Completed = append(info.Completed, subtask)
		default:
			info.Subtask = &subtask
		}
	}

	return &info
}

func (fp *flowProvider) updateMsgChainResult(chain []llms.MessageContent, name, result string) ([]llms.MessageContent, error) {
	if len(chain) == 0 {
		return []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, result)}, nil
	}

	ast, err := cast.NewChainAST(chain, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create chain ast: %w", err)
	}

	lastSection := ast.Sections[len(ast.Sections)-1]
	if len(lastSection.Body) == 0 {
		ast.AppendHumanMessage(result)
		return ast.Messages(), nil
	}

	lastBody := lastSection.Body[len(lastSection.Body)-1]
	switch lastBody.Type {
	case cast.Completion, cast.Summarization:
		ast.AppendHumanMessage(result)
		return ast.Messages(), nil
	case cast.RequestResponse:
		for _, msg := range lastBody.ToolMessages {
			for pdx, part := range msg.Parts {
				toolCallResp, ok := part.(llms.ToolCallResponse)
				if !ok {
					continue
				}

				if toolCallResp.Name == name {
					toolCallResp.Content = result
					msg.Parts[pdx] = toolCallResp
					return ast.Messages(), nil
				}
			}
		}

		ast.AppendHumanMessage(result)
		return ast.Messages(), nil
	default:
		return nil, fmt.Errorf("unknown message type: %d", lastBody.Type)
	}
}

// Makes chain consistent by adding default responses for any pending tool calls
func (fp *flowProvider) ensureChainConsistency(chain []llms.MessageContent) ([]llms.MessageContent, error) {
	if len(chain) == 0 {
		return chain, nil
	}

	ast, err := cast.NewChainAST(chain, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create chain ast: %w", err)
	}

	return ast.Messages(), nil
}

func (fp *flowProvider) getTaskPrimaryAgentChainSummary(
	ctx context.Context,
	taskID int64,
	summarizerHandler tools.SummarizeHandler,
) (string, error) {
	msgChain, err := fp.db.GetFlowTaskTypeLastMsgChain(ctx, database.GetFlowTaskTypeLastMsgChainParams{
		FlowID: fp.flowID,
		TaskID: database.Int64ToNullInt64(&taskID),
		Type:   database.MsgchainTypePrimaryAgent,
	})
	if err != nil || isEmptyChain(msgChain.Chain) {
		return "", fmt.Errorf("failed to get task primary agent chain: %w", err)
	}

	chain := []llms.MessageContent{}
	if err := json.Unmarshal(msgChain.Chain, &chain); err != nil {
		return "", fmt.Errorf("failed to unmarshal task primary agent chain: %w", err)
	}

	ast, err := cast.NewChainAST(chain, true)
	if err != nil {
		return "", fmt.Errorf("failed to create refiner chain ast: %w", err)
	}

	var humanMessages, aiMessages []llms.MessageContent
	for _, section := range ast.Sections {
		if section.Header.HumanMessage != nil {
			humanMessages = append(humanMessages, *section.Header.HumanMessage)
		}
		for _, pair := range section.Body {
			aiMessages = append(aiMessages, pair.Messages()...)
		}
	}

	humanSummary, err := csum.GenerateSummary(ctx, summarizerHandler, humanMessages, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate human summary: %w", err)
	}

	aiSummary, err := csum.GenerateSummary(ctx, summarizerHandler, humanMessages, aiMessages)
	if err != nil {
		return "", fmt.Errorf("failed to generate ai summary: %w", err)
	}

	summary := fmt.Sprintf(`## Task Summary

### User Requirements
*Summarized input from user:*

%s

### Execution Results
*Summarized actions and outcomes:*

%s`, humanSummary, aiSummary)
	return summary, nil
}

func (fp *flowProvider) getTaskMsgLogsSummary(
	ctx context.Context,
	taskID int64,
	summarizerHandler tools.SummarizeHandler,
) (string, error) {
	msgLogs, err := fp.db.GetTaskMsgLogs(ctx, database.Int64ToNullInt64(&taskID))
	if err != nil {
		return "", fmt.Errorf("failed to get task msg logs: %w", err)
	}

	if len(msgLogs) == 0 {
		return "no msg logs", nil
	}

	// truncate msg logs result to cut down the size the message to summarize
	for _, msgLog := range msgLogs {
		if len(msgLog.Result) > msgLogResultEntrySizeLimit {
			msgLog.Result = msgLog.Result[:msgLogResultEntrySizeLimit] + textTruncateMessage
		}
	}

	message, err := fp.prompter.RenderTemplate(templates.PromptTypeExecutionLogs, map[string]any{
		"MsgLogs": msgLogs,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render task msg logs template: %w", err)
	}

	for l := len(msgLogs) / 2; l > 2; l /= 2 {
		if len(message) < msgLogResultSummarySizeLimit {
			break
		}

		msgLogs = msgLogs[len(msgLogs)-l:]
		message, err = fp.prompter.RenderTemplate(templates.PromptTypeExecutionLogs, map[string]any{
			"MsgLogs": msgLogs,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render task msg logs template: %w", err)
		}
	}

	summary, err := summarizerHandler(ctx, message)
	if err != nil {
		return "", fmt.Errorf("failed to summarize task msg logs: %w", err)
	}

	return summary, nil
}

func (fp *flowProvider) restoreChain(
	ctx context.Context,
	taskID, subtaskID *int64,
	optAgentType pconfig.ProviderOptionsType,
	msgChainType database.MsgchainType,
	systemPrompt, humanPrompt string,
) (int64, []llms.MessageContent, error) {
	var chain []llms.MessageContent
	fallback := func() {
		chain = []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		}
		if humanPrompt != "" {
			chain = append(chain, llms.TextParts(llms.ChatMessageTypeHuman, humanPrompt))
		}
	}

	msgChain, err := fp.db.GetFlowTaskTypeLastMsgChain(ctx, database.GetFlowTaskTypeLastMsgChainParams{
		FlowID: fp.flowID,
		TaskID: database.Int64ToNullInt64(taskID),
		Type:   msgChainType,
	})
	if err != nil || isEmptyChain(msgChain.Chain) {
		fallback()
	} else {
		err = func() error {
			err = json.Unmarshal(msgChain.Chain, &chain)
			if err != nil {
				return fmt.Errorf("failed to unmarshal msg chain: %w", err)
			}

			ast, err := cast.NewChainAST(chain, true)
			if err != nil {
				return fmt.Errorf("failed to create refiner chain ast: %w", err)
			}

			if len(ast.Sections) == 0 {
				return fmt.Errorf("failed to get sections from refiner chain ast")
			}

			systemMessage := llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt)
			ast.Sections[0].Header.SystemMessage = &systemMessage
			if humanPrompt != "" {
				lastSection := ast.Sections[len(ast.Sections)-1]
				if len(lastSection.Body) == 0 {
					// do not add a new human message if the previous human message is not yet completed
					lastSection.Header.HumanMessage = nil
				} else {
					lastBody := lastSection.Body[len(lastSection.Body)-1]
					if lastBody.Type == cast.RequestResponse && len(lastBody.ToolMessages) == 0 {
						// prevent using incomplete chain without tool call response
						lastSection.Body = lastSection.Body[:len(lastSection.Body)-1]
					}
				}
				ast.AppendHumanMessage(humanPrompt)
			}

			summarizeHandler := fp.GetSummarizeResultHandler(taskID, subtaskID)
			summarizer := csum.NewSummarizer(csum.SummarizerConfig{
				PreserveLast:   true,
				UseQA:          true,
				SummHumanInQA:  true,
				LastSecBytes:   lastSecBytesAfterRestore,
				MaxBPBytes:     maxBPBytesAfterRestore,
				MaxQASections:  maxQASectionsAfterRestore,
				MaxQABytes:     maxQABytesAfterRestore,
				KeepQASections: keepQASectionsAfterRestore,
			})

			chain, err = summarizer.SummarizeChain(ctx, summarizeHandler, ast.Messages())
			if err != nil {
				chain = ast.Messages()
			}

			return nil
		}()
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Warn("failed to restore chain")
			fallback()
		}
	}

	chainBlob, err := json.Marshal(chain)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to marshal msg chain: %w", err)
	}

	msgChain, err = fp.db.CreateMsgChain(ctx, database.CreateMsgChainParams{
		Type:          msgChainType,
		Model:         fp.Model(optAgentType),
		ModelProvider: string(fp.Type()),
		Chain:         chainBlob,
		FlowID:        fp.flowID,
		TaskID:        database.Int64ToNullInt64(taskID),
	})
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create msg chain: %w", err)
	}

	return msgChain.ID, chain, nil
}

// Eliminates code duplication by abstracting database operations on message chains
func (fp *flowProvider) processChain(
	ctx context.Context,
	msgChainID int64,
	logger *logrus.Entry,
	transform func([]llms.MessageContent) ([]llms.MessageContent, error),
) error {
	msgChain, err := fp.db.GetMsgChain(ctx, msgChainID)
	if err != nil {
		logger.WithError(err).Error("failed to get message chain")
		return fmt.Errorf("failed to get message chain %d: %w", msgChainID, err)
	}

	var chain []llms.MessageContent
	if err := json.Unmarshal(msgChain.Chain, &chain); err != nil {
		logger.WithError(err).Error("failed to unmarshal message chain")
		return fmt.Errorf("failed to unmarshal message chain %d: %w", msgChainID, err)
	}

	updatedChain, err := transform(chain)
	if err != nil {
		logger.WithError(err).Error("failed to transform chain")
		return fmt.Errorf("failed to transform chain: %w", err)
	}

	chainBlob, err := json.Marshal(updatedChain)
	if err != nil {
		logger.WithError(err).Error("failed to marshal updated chain")
		return fmt.Errorf("failed to marshal updated chain %d: %w", msgChainID, err)
	}

	_, err = fp.db.UpdateMsgChain(ctx, database.UpdateMsgChainParams{
		Chain: chainBlob,
		ID:    msgChainID,
	})
	if err != nil {
		logger.WithError(err).Error("failed to update message chain")
		return fmt.Errorf("failed to update message chain %d: %w", msgChainID, err)
	}

	return nil
}

func (fp *flowProvider) prepareExecutionContext(ctx context.Context, taskID, subtaskID int64) (string, error) {
	tasksInfo, err := fp.getTasksInfo(ctx, taskID)
	if err != nil {
		return "", fmt.Errorf("failed to get tasks info: %w", err)
	}

	subtasksInfo := fp.getSubtasksInfo(taskID, tasksInfo.Subtasks)
	if subtasksInfo.Subtask == nil {
		subtasks := make([]database.Subtask, 0, len(subtasksInfo.Planned)+len(subtasksInfo.Completed))
		subtasks = append(subtasks, subtasksInfo.Planned...)
		subtasks = append(subtasks, subtasksInfo.Completed...)
		slices.SortFunc(subtasks, func(a, b database.Subtask) int {
			return int(a.ID - b.ID)
		})

		for i, subtask := range subtasks {
			if subtask.ID == subtaskID {
				subtasksInfo.Subtask = &subtask
				subtasksInfo.Planned = subtasks[i+1:]
				subtasksInfo.Completed = subtasks[:i]
				break
			}
		}
	}

	executionContextRaw, err := fp.prompter.RenderTemplate(templates.PromptTypeFullExecutionContext, map[string]any{
		"Task":              tasksInfo.Task,
		"Tasks":             tasksInfo.Tasks,
		"CompletedSubtasks": subtasksInfo.Completed,
		"Subtask":           subtasksInfo.Subtask,
		"PlannedSubtasks":   subtasksInfo.Planned,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render execution context: %w", err)
	}

	summarizeHandler := fp.GetSummarizeResultHandler(&taskID, &subtaskID)
	executionContext, err := summarizeHandler(ctx, executionContextRaw)
	if err != nil {
		return "", fmt.Errorf("failed to summarize execution context: %w", err)
	}

	return executionContext, nil
}

func (fp *flowProvider) getExecutionContext(ctx context.Context, taskID, subtaskID *int64) (string, error) {
	if taskID != nil && subtaskID != nil {
		return fp.getExecutionContextBySubtask(ctx, *taskID, *subtaskID)
	}

	if taskID != nil {
		return fp.getExecutionContextByTask(ctx, *taskID)
	}

	return fp.getExecutionContextByFlow(ctx)
}

func (fp *flowProvider) getExecutionContextBySubtask(ctx context.Context, taskID, subtaskID int64) (string, error) {
	subtask, err := fp.db.GetSubtask(ctx, subtaskID)
	if err == nil && subtask.TaskID == taskID && subtask.Context != "" {
		return subtask.Context, nil
	}

	return fp.getExecutionContextByTask(ctx, taskID)
}

func (fp *flowProvider) getExecutionContextByTask(ctx context.Context, taskID int64) (string, error) {
	tasksInfo, err := fp.getTasksInfo(ctx, taskID)
	if err != nil {
		return fp.getExecutionContextByFlow(ctx)
	}

	subtasksInfo := fp.getSubtasksInfo(taskID, tasksInfo.Subtasks)
	executionContext, err := fp.prompter.RenderTemplate(templates.PromptTypeShortExecutionContext, map[string]any{
		"Task":              tasksInfo.Task,
		"Tasks":             tasksInfo.Tasks,
		"CompletedSubtasks": subtasksInfo.Completed,
		"Subtask":           subtasksInfo.Subtask,
		"PlannedSubtasks":   subtasksInfo.Planned,
	})
	if err != nil {
		return fp.getExecutionContextByFlow(ctx)
	}

	return executionContext, nil
}

func (fp *flowProvider) getExecutionContextByFlow(ctx context.Context) (string, error) {
	tasks, err := fp.db.GetFlowTasks(ctx, fp.flowID)
	if err != nil {
		return "", fmt.Errorf("failed to get flow tasks: %w", err)
	}

	if len(tasks) == 0 {
		return "flow has no tasks, it's using in assistant mode", nil
	}

	subtasks, err := fp.db.GetFlowSubtasks(ctx, fp.flowID)
	if err != nil {
		return "", fmt.Errorf("failed to get flow subtasks: %w", err)
	}

	for tid := len(tasks) - 1; tid >= 0; tid-- {
		taskID := tasks[tid].ID

		subtasksInfo := fp.getSubtasksInfo(taskID, subtasks)
		executionContext, err := fp.prompter.RenderTemplate(templates.PromptTypeShortExecutionContext, map[string]any{
			"Task":              tasks[tid],
			"Tasks":             tasks,
			"CompletedSubtasks": subtasksInfo.Completed,
			"Subtask":           subtasksInfo.Subtask,
			"PlannedSubtasks":   subtasksInfo.Planned,
		})
		if err != nil {
			continue
		}

		return executionContext, nil
	}

	subtasksInfo := fp.getSubtasksInfo(0, subtasks)
	executionContext, err := fp.prompter.RenderTemplate(templates.PromptTypeShortExecutionContext, map[string]any{
		"Tasks":             tasks,
		"CompletedSubtasks": subtasksInfo.Completed,
		"Subtask":           subtasksInfo.Subtask,
		"PlannedSubtasks":   subtasksInfo.Planned,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render execution context: %w", err)
	}

	return executionContext, nil
}

func (fp *flowProvider) subtasksToMarkdown(subtasks []tools.SubtaskInfo) string {
	var buffer strings.Builder
	for sid, subtask := range subtasks {
		buffer.WriteString(fmt.Sprintf("# Subtask %d\n\n", sid+1))
		buffer.WriteString(fmt.Sprintf("## %s\n\n%s\n\n", subtask.Title, subtask.Description))
	}

	return buffer.String()
}

func (fp *flowProvider) getContainerPortsDescription() string {
	ports := docker.GetPrimaryContainerPorts(fp.flowID)
	var buffer strings.Builder
	buffer.WriteString("This container has the following ports which bind to the host:\n")
	for _, port := range ports {
		buffer.WriteString(fmt.Sprintf("* %s:%d -> %d/tcp (in container)\n", fp.publicIP, port, port))
	}
	if fp.publicIP == "0.0.0.0" {
		buffer.WriteString("you need to discover the public IP yourself via the following command:\n")
		buffer.WriteString("`curl -s https://api.ipify.org` or `curl -s ipinfo.io/ip` or `curl -s ifconfig.me`\n")
	}
	buffer.WriteString("you can listen these ports the container inside and receive connections from the internet.")
	return buffer.String()
}

func getCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func isEmptyChain(msgChain json.RawMessage) bool {
	var msgList []llms.MessageContent

	if err := json.Unmarshal(msgChain, &msgList); err != nil {
		return true
	}

	return len(msgList) == 0
}
