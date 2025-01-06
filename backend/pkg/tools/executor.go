package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"pentagi/pkg/database"
	"pentagi/pkg/schema"

	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

const DefaultResultSizeLimit = 8 * 1024 // 8 KB

type dummyMessage struct {
	Message string `json:"message"`
}

type customExecutor struct {
	flowID    int64
	taskID    *int64
	subtaskID *int64

	db    database.Querier
	mlp   MsgLogProvider
	store *pgvector.Store
	vslp  VectorStoreLogProvider

	definitions []llms.FunctionDefinition
	handlers    map[string]ExecutorHandler
	barriers    map[string]struct{}
	summarizer  SummarizeHandler
}

func (ce *customExecutor) Tools() []llms.Tool {
	tools := make([]llms.Tool, 0, len(ce.definitions))
	for idx := range ce.definitions {
		tools = append(tools, llms.Tool{
			Type:     "function",
			Function: &ce.definitions[idx],
		})
	}

	return tools
}

func (ce *customExecutor) Execute(ctx context.Context, id, name string, args json.RawMessage) (string, error) {
	handler, ok := ce.handlers[name]
	if !ok {
		return fmt.Sprintf("function '%s' not found in available tools list", name), nil
	}

	var raw any
	if err := json.Unmarshal(args, &raw); err != nil {
		return fmt.Sprintf("failed to unmarshal '%s' tool call arguments: %v: fix it", name, err), nil
	}

	tc, err := ce.db.CreateToolcall(ctx, database.CreateToolcallParams{
		CallID:    id,
		Status:    database.ToolcallStatusRunning,
		Name:      name,
		Args:      args,
		FlowID:    ce.flowID,
		TaskID:    database.Int64ToNullInt64(ce.taskID),
		SubtaskID: database.Int64ToNullInt64(ce.subtaskID),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create toolcall: %w", err)
	}

	wrapHandler := func(ctx context.Context, name string, args json.RawMessage) (string, database.MsglogResultFormat, error) {
		resultFormat := getMessageResultFormat(name)
		result, err := handler(ctx, name, args)
		if err != nil {
			_, _ = ce.db.UpdateToolcallFailedResult(ctx, database.UpdateToolcallFailedResultParams{
				Result: fmt.Sprintf("failed to execute handler: %s", err.Error()),
				ID:     tc.ID,
			})
			return "", resultFormat, fmt.Errorf("failed to execute handler: %w", err)
		}

		allowSummarize := slices.Contains(allowedSummarizingToolsResult, name)
		if ce.summarizer != nil && allowSummarize && len(result) > DefaultResultSizeLimit {
			result, err = ce.summarizer(ctx, result)
			if err != nil {
				_, _ = ce.db.UpdateToolcallFailedResult(ctx, database.UpdateToolcallFailedResultParams{
					Result: fmt.Sprintf("failed to summarize result: %s", err.Error()),
					ID:     tc.ID,
				})
				return "", resultFormat, fmt.Errorf("failed to summarize result: %w", err)
			}
			resultFormat = database.MsglogResultFormatMarkdown
		}

		_, err = ce.db.UpdateToolcallFinishedResult(ctx, database.UpdateToolcallFinishedResultParams{
			Result: result,
			ID:     tc.ID,
		})
		if err != nil {
			return "", resultFormat, fmt.Errorf("failed to update toolcall result: %w", err)
		}

		return result, resultFormat, nil
	}

	msg := ce.getMessage(args)
	if msg == "" { // no arg message to log and execute handler immediately
		result, _, err := wrapHandler(ctx, name, args)
		return result, err
	}

	msgID, err := ce.mlp.PutMsg(ctx, getMessageType(name), ce.taskID, ce.subtaskID, msg)
	if err != nil {
		return "", err
	}

	result, resultFormat, err := wrapHandler(ctx, name, args)
	if err != nil {
		return "", err
	}

	if err := ce.storeToolResult(ctx, name, result, args); err != nil {
		return "", fmt.Errorf("failed to store tool result in long-term memory: %w", err)
	}

	if err := ce.mlp.UpdateMsgResult(ctx, msgID, result, resultFormat); err != nil {
		return "", err
	}

	return result, nil
}

func (ce *customExecutor) IsBarrierFunction(name string) bool {
	_, ok := ce.barriers[name]
	return ok
}

func (ce *customExecutor) GetBarrierToolNames() []string {
	names := make([]string, 0, len(ce.barriers))
	for name := range ce.barriers {
		names = append(names, name)
	}

	return names
}

func (ce *customExecutor) GetToolSchema(name string) (*schema.Schema, error) {
	for _, def := range ce.definitions {
		if def.Name == name {
			return ce.converToJSONSchema(def.Parameters)
		}
	}

	if def, ok := registryDefinitions[name]; ok {
		return ce.converToJSONSchema(def.Parameters)
	}

	return nil, fmt.Errorf("tool %s not found", name)
}

func (ce *customExecutor) converToJSONSchema(params any) (*schema.Schema, error) {
	jsonSchema, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	var schema schema.Schema
	if err := json.Unmarshal(jsonSchema, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return &schema, nil
}

func (ce *customExecutor) getMessage(args json.RawMessage) string {
	var msg dummyMessage
	if err := json.Unmarshal(args, &msg); err != nil {
		return ""
	}

	return msg.Message
}

func (ce *customExecutor) storeToolResult(ctx context.Context, name, result string, args json.RawMessage) error {
	if ce.store == nil {
		return nil
	}

	if !slices.Contains(allowedStoringInMemoryTools, name) {
		return nil
	}

	var buffer strings.Builder
	buffer.WriteString(fmt.Sprintf("## Incomming arguments %s\n\n", args))
	buffer.WriteString(fmt.Sprintf("### Tool result\n\n%s\n\n", result))
	text := buffer.String()

	split := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(2000),
		textsplitter.WithChunkOverlap(100),
		textsplitter.WithCodeBlocks(true),
		textsplitter.WithHeadingHierarchy(true),
	)
	docs, err := documentloaders.NewText(strings.NewReader(text)).LoadAndSplit(ctx, split)
	if err != nil {
		return fmt.Errorf("failed to split tool result: %w", err)
	}

	for _, doc := range docs {
		if doc.Metadata == nil {
			doc.Metadata = map[string]any{}
		}
		if ce.taskID != nil {
			doc.Metadata["task_id"] = *ce.taskID
		}
		if ce.subtaskID != nil {
			doc.Metadata["subtask_id"] = *ce.subtaskID
		}
		doc.Metadata["flow_id"] = ce.flowID
		doc.Metadata["tool_name"] = name
		if def, ok := registryDefinitions[name]; ok {
			doc.Metadata["tool_description"] = def.Description
		}
		doc.Metadata["doc_type"] = memoryVectorStoreDefaultType
		doc.Metadata["part_size"] = len(doc.PageContent)
		doc.Metadata["total_size"] = len(text)
	}

	if _, err := ce.store.AddDocuments(ctx, docs); err != nil {
		return fmt.Errorf("failed to store tool result: %w", err)
	}

	if agentCtx, ok := GetAgentContext(ctx); ok {
		filtersData, err := json.Marshal(map[string]any{
			"doc_type":   memoryVectorStoreDefaultType,
			"tool_name":  name,
			"task_id":    ce.taskID,
			"subtask_id": ce.subtaskID,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal filters: %w", err)
		}
		query, err := ce.argsToMarkdown(args)
		if err != nil {
			return fmt.Errorf("failed to convert arguments to markdown: %w", err)
		}
		_, _ = ce.vslp.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			filtersData,
			query,
			database.VecstoreActionTypeStore,
			result,
			ce.taskID,
			ce.subtaskID,
		)
	}

	return nil
}

func (ce *customExecutor) argsToMarkdown(args json.RawMessage) (string, error) {
	var argsMap map[string]any
	if err := json.Unmarshal(args, &argsMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	var buffer strings.Builder
	for key, value := range argsMap {
		if key == "message" {
			continue
		}
		buffer.WriteString(fmt.Sprintf("* %s: %v\n", key, value))
	}

	return buffer.String(), nil
}
