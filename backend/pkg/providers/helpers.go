package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/tools"

	"github.com/tmc/langchaingo/llms"
)

const RepeatingToolCallThreshold = 3

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

func (fp *flowProvider) updateMsgChainResult(chain []llms.MessageContent, name, result string) error {
	if len(chain) == 0 {
		return fmt.Errorf("msgchain is empty")
	}

	for idx := len(chain) - 1; idx > 0; idx-- {
		msgContent := chain[idx]
		if msgContent.Role != llms.ChatMessageTypeTool {
			continue
		}

		for jdx, part := range msgContent.Parts {
			toolCallResp, ok := part.(llms.ToolCallResponse)
			if !ok {
				continue
			}

			if toolCallResp.Name == name {
				toolCallResp.Content = result
				chain[idx].Parts[jdx] = toolCallResp
				return nil
			}
		}
	}

	return fmt.Errorf("message part with tool call result is not found in msgchain")
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
