package validator

import (
	"database/sql"
	"encoding/json"
	"time"

	"pentagi/pkg/cast"
	"pentagi/pkg/csum"
	"pentagi/pkg/database"
	"pentagi/pkg/providers"
	"pentagi/pkg/tools"
)

// CreateDummyTemplateData creates realistic test data that matches the actual data types used in production
func CreateDummyTemplateData() map[string]any {
	// Current time for database timestamps
	currentTime := sql.NullTime{
		Time:  time.Date(2025, 7, 2, 12, 30, 45, 0, time.UTC),
		Valid: true,
	}

	// Create proper BarrierTools using the same logic as GetBarrierTools
	barrierTools := createBarrierTools()

	return map[string]any{
		// Tool names - exact values from tools package constants
		"FinalyToolName":            tools.FinalyToolName,
		"SearchToolName":            tools.SearchToolName,
		"PentesterToolName":         tools.PentesterToolName,
		"CoderToolName":             tools.CoderToolName,
		"AdviceToolName":            tools.AdviceToolName,
		"MemoristToolName":          tools.MemoristToolName,
		"MaintenanceToolName":       tools.MaintenanceToolName,
		"GraphitiSearchToolName":    tools.GraphitiSearchToolName,
		"GraphitiEnabled":           true,
		"TerminalToolName":          tools.TerminalToolName,
		"FileToolName":              tools.FileToolName,
		"BrowserToolName":           tools.BrowserToolName,
		"GoogleToolName":            tools.GoogleToolName,
		"DuckDuckGoToolName":        tools.DuckDuckGoToolName,
		"SploitusToolName":          tools.SploitusToolName,
		"TavilyToolName":            tools.TavilyToolName,
		"TraversaalToolName":        tools.TraversaalToolName,
		"PerplexityToolName":        tools.PerplexityToolName,
		"SearchInMemoryToolName":    tools.SearchInMemoryToolName,
		"SearchGuideToolName":       tools.SearchGuideToolName,
		"SearchAnswerToolName":      tools.SearchAnswerToolName,
		"SearchCodeToolName":        tools.SearchCodeToolName,
		"StoreGuideToolName":        tools.StoreGuideToolName,
		"StoreAnswerToolName":       tools.StoreAnswerToolName,
		"StoreCodeToolName":         tools.StoreCodeToolName,
		"SearchResultToolName":      tools.SearchResultToolName,
		"EnricherToolName":          tools.EnricherResultToolName,
		"MemoristResultToolName":    tools.MemoristResultToolName,
		"MaintenanceResultToolName": tools.MaintenanceResultToolName,
		"CodeResultToolName":        tools.CodeResultToolName,
		"HackResultToolName":        tools.HackResultToolName,
		"EnricherResultToolName":    tools.EnricherResultToolName,
		"ReportResultToolName":      tools.ReportResultToolName,
		"SubtaskListToolName":       tools.SubtaskListToolName,
		"SubtaskPatchToolName":      tools.SubtaskPatchToolName,
		"AskUserToolName":           tools.AskUserToolName,
		"AskUserEnabled":            true,

		// Summarization related - using constants from proper packages
		"SummarizationToolName":   cast.SummarizationToolName,
		"SummarizedContentPrefix": csum.SummarizedContentPrefix,

		// Boolean flags
		"UseAgents":            true,
		"IsDefaultDockerImage": false,

		// Docker and environment
		"DockerImage": "vxcontrol/kali-linux:latest",
		"Cwd":         "/workspace",
		"ContainerPorts": `This container has the following ports which bind to the host:
		* 0.0.0.0:8080 -> 8080/tcp (in container)
		* 0.0.0.0:8443 -> 8443/tcp (in container)
		you can listen these ports the container inside and receive connections from the internet.`,

		// Context and state
		"ExecutionContext": "Test execution context with current task and subtask information",
		"ExecutionDetails": "Test execution details",
		"ExecutionLogs":    "Test execution logs summary",
		"ExecutionState":   "Test execution state summary",

		// Language and time
		"Lang":        "English",
		"CurrentTime": "2025-07-02 12:30:45",

		// Template control - using constant from providers package
		"ToolPlaceholder": providers.ToolPlaceholder,

		// Numeric limits
		"N": providers.TasksNumberLimit,

		// Input/Output data
		"Input":    "Test input for the task",
		"Question": "Test question for processing",
		"Message":  "Test message content",
		"Code":     "print('Hello, World!')",
		"Output":   "Hello, World!",
		"Query":    "test search query",
		"Result":   "Test result content",
		"Enriches": "Test enriched information from various sources",

		// Image and model selection
		"DefaultImage":           "ubuntu:latest",
		"DefaultImageForPentest": "vxcontrol/kali-linux:latest",

		// Database entities - using proper structures with correct types and all fields
		"Task": database.Task{
			ID:        1,
			Status:    database.TaskStatusRunning,
			Title:     "Test Task",
			Input:     "Test task input",
			Result:    "Test task result",
			FlowID:    100,
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
		},

		"Tasks": []database.Task{
			{
				ID:        1,
				Status:    database.TaskStatusFinished,
				Title:     "Previous Task 1",
				Input:     "Previous task input 1",
				Result:    "Previous task result 1",
				FlowID:    100,
				CreatedAt: currentTime,
				UpdatedAt: currentTime,
			},
			{
				ID:        2,
				Status:    database.TaskStatusRunning,
				Title:     "Current Task",
				Input:     "Current task input",
				Result:    "",
				FlowID:    100,
				CreatedAt: currentTime,
				UpdatedAt: currentTime,
			},
		},

		"Subtask": &database.Subtask{
			ID:          10,
			Status:      database.SubtaskStatusRunning,
			Title:       "Current Subtask",
			Description: "Test subtask description with detailed instructions",
			Result:      "",
			TaskID:      1,
			Context:     "Test subtask context",
			CreatedAt:   currentTime,
			UpdatedAt:   currentTime,
		},

		"PlannedSubtasks": []database.Subtask{
			{
				ID:          11,
				Status:      database.SubtaskStatusCreated,
				Title:       "Planned Subtask 1",
				Description: "First planned subtask description",
				Result:      "",
				TaskID:      1,
				Context:     "",
				CreatedAt:   currentTime,
				UpdatedAt:   currentTime,
			},
			{
				ID:          12,
				Status:      database.SubtaskStatusCreated,
				Title:       "Planned Subtask 2",
				Description: "Second planned subtask description",
				Result:      "",
				TaskID:      1,
				Context:     "",
				CreatedAt:   currentTime,
				UpdatedAt:   currentTime,
			},
		},

		"CompletedSubtasks": []database.Subtask{
			{
				ID:          8,
				Status:      database.SubtaskStatusFinished,
				Title:       "Completed Subtask 1",
				Description: "First completed subtask",
				Result:      "Successfully completed with test result",
				TaskID:      1,
				Context:     "Completed subtask context",
				CreatedAt:   currentTime,
				UpdatedAt:   currentTime,
			},
			{
				ID:          9,
				Status:      database.SubtaskStatusFinished,
				Title:       "Completed Subtask 2",
				Description: "Second completed subtask",
				Result:      "Another successful completion",
				TaskID:      1,
				Context:     "Another completed context",
				CreatedAt:   currentTime,
				UpdatedAt:   currentTime,
			},
		},

		"Subtasks": []database.Subtask{
			{
				ID:          8,
				Status:      database.SubtaskStatusFinished,
				Title:       "Subtask 1",
				Description: "First subtask description",
				Result:      "First subtask result",
				TaskID:      1,
				Context:     "First subtask context",
				CreatedAt:   currentTime,
				UpdatedAt:   currentTime,
			},
			{
				ID:          9,
				Status:      database.SubtaskStatusRunning,
				Title:       "Subtask 2",
				Description: "Second subtask description",
				Result:      "",
				TaskID:      1,
				Context:     "Second subtask context",
				CreatedAt:   currentTime,
				UpdatedAt:   currentTime,
			},
		},

		"MsgLogs": []database.Msglog{
			{
				ID:           1,
				Type:         database.MsglogTypeTerminal,
				Message:      "Executed terminal command",
				Result:       "Command output result",
				FlowID:       100,
				TaskID:       sql.NullInt64{Int64: 1, Valid: true},
				SubtaskID:    sql.NullInt64{Int64: 10, Valid: true},
				CreatedAt:    currentTime,
				ResultFormat: database.MsglogResultFormatTerminal,
				Thinking:     sql.NullString{String: "Thinking about terminal execution", Valid: true},
			},
			{
				ID:           2,
				Type:         database.MsglogTypeSearch,
				Message:      "Performed web search",
				Result:       "Search results data",
				FlowID:       100,
				TaskID:       sql.NullInt64{Int64: 1, Valid: true},
				SubtaskID:    sql.NullInt64{Int64: 10, Valid: true},
				CreatedAt:    currentTime,
				ResultFormat: database.MsglogResultFormatMarkdown,
				Thinking:     sql.NullString{String: "Thinking about search strategy", Valid: true},
			},
		},

		// Barrier tools - using proper logic from tools package
		"BarrierTools":     barrierTools,
		"BarrierToolNames": []string{tools.FinalyToolName, tools.AskUserToolName},

		// Request context for reflector
		"Request": "Original user request",

		// Task and subtask IDs
		"TaskID":    int64(1),
		"SubtaskID": int64(10),

		// Additional variables found in templates
		"Name":   "Test name",
		"Schema": "Test schema",

		// Tool call fixer variables
		"ToolCallName":   "test_tool_call",
		"ToolCallArgs":   `{"param1": "value1", "param2": "value2"}`,
		"ToolCallSchema": `{"type": "object", "properties": {"param1": {"type": "string"}, "param2": {"type": "string"}}}`,
		"ToolCallError":  "Test tool call error: invalid argument format",

		// Tool call ID collector variables
		"RandomContext": "Test random context",
		"FunctionName":  "test_function",
		"Samples": []string{
			"Test sample 1",
			"Test sample 2",
			"Test sample 3",
		},
		"PreviousAttempts": []struct {
			Template string
			Error    string
		}{
			{
				Template: "Test previous attempt 1",
				Error:    "Test previous attempt error 1",
			}, {
				Template: "Test previous attempt 2",
				Error:    "Test previous attempt error 2",
			}, {
				Template: "Test previous attempt 3",
				Error:    "Test previous attempt error 3",
			},
		},
	}
}

// createBarrierTools replicates the logic from GetBarrierTools() to create proper barrier tools
func createBarrierTools() []tools.FunctionInfo {
	// Get barrier tool names from registry mapping
	toolsMapping := tools.GetToolTypeMapping()
	registryDefinitions := tools.GetRegistryDefinitions()

	var barrierTools []tools.FunctionInfo

	for toolName, toolType := range toolsMapping {
		if toolType == tools.BarrierToolType {
			if def, ok := registryDefinitions[toolName]; ok {
				// Convert parameters to JSON schema (simplified version of converToJSONSchema)
				schemaJSON, err := json.Marshal(def.Parameters)
				if err != nil {
					continue
				}

				barrierTools = append(barrierTools, tools.FunctionInfo{
					Name:   toolName,
					Schema: string(schemaJSON),
				})
			}
		}
	}

	return barrierTools
}
