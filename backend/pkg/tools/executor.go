package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"text/template"
	"time"

	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/schema"

	"github.com/vxcontrol/langchaingo/documentloaders"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/textsplitter"
	"github.com/vxcontrol/langchaingo/vectorstores/pgvector"
)

const DefaultResultSizeLimit = 16 * 1024 // 16 KB

const maxArgValueLength = 1024 // 1 KB limit for argument values

type dummyMessage struct {
	Message string `json:"message"`
}

// observationWrapper wraps different observation types with unified interface
type observationWrapper interface {
	ctx() context.Context
	end(result string, err error, durationSeconds float64)
}

// toolObservationWrapper wraps TOOL observation
type toolObservationWrapper struct {
	context     context.Context
	observation langfuse.Tool
}

func (w *toolObservationWrapper) ctx() context.Context {
	return w.context
}

func (w *toolObservationWrapper) end(result string, err error, durationSeconds float64) {
	opts := []langfuse.ToolOption{
		langfuse.WithToolOutput(result),
	}
	if err != nil {
		opts = append(opts,
			langfuse.WithToolStatus(err.Error()),
			langfuse.WithToolLevel(langfuse.ObservationLevelError),
		)
	} else {
		opts = append(opts,
			langfuse.WithToolStatus("success"),
		)
	}
	w.observation.End(opts...)
}

// agentObservationWrapper wraps AGENT observation
type agentObservationWrapper struct {
	context     context.Context
	observation langfuse.Agent
}

func (w *agentObservationWrapper) ctx() context.Context {
	return w.context
}

func (w *agentObservationWrapper) end(result string, err error, durationSeconds float64) {
	opts := []langfuse.AgentOption{
		langfuse.WithAgentOutput(result),
	}
	if err != nil {
		opts = append(opts,
			langfuse.WithAgentStatus(err.Error()),
			langfuse.WithAgentLevel(langfuse.ObservationLevelError),
		)
	} else {
		opts = append(opts,
			langfuse.WithAgentStatus("success"),
		)
	}
	w.observation.End(opts...)
}

// spanObservationWrapper wraps SPAN observation (used for barrier tools)
type spanObservationWrapper struct {
	context     context.Context
	observation langfuse.Span
}

func (w *spanObservationWrapper) ctx() context.Context {
	return w.context
}

func (w *spanObservationWrapper) end(result string, err error, durationSeconds float64) {
	opts := []langfuse.SpanOption{
		langfuse.WithSpanOutput(result),
	}
	if err != nil {
		opts = append(opts,
			langfuse.WithSpanStatus(err.Error()),
			langfuse.WithSpanLevel(langfuse.ObservationLevelError),
		)
	} else {
		opts = append(opts,
			langfuse.WithSpanStatus("success"),
		)
	}
	w.observation.End(opts...)
}

// noopObservationWrapper is a no-op wrapper for tools that create observations internally
type noopObservationWrapper struct {
	context context.Context
}

func (w *noopObservationWrapper) ctx() context.Context {
	return w.context
}

func (w *noopObservationWrapper) end(result string, err error, durationSeconds float64) {
	// no-op
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

func (ce *customExecutor) createToolObservation(ctx context.Context, name string, args json.RawMessage) observationWrapper {
	ctx, observation := obs.Observer.NewObservation(ctx)
	metadata := langfuse.Metadata{
		"tool_name":     name,
		"tool_category": GetToolType(name).String(),
		"flow_id":       ce.flowID,
	}
	if ce.taskID != nil {
		metadata["task_id"] = *ce.taskID
	}
	if ce.subtaskID != nil {
		metadata["subtask_id"] = *ce.subtaskID
	}

	tool := observation.Tool(
		langfuse.WithToolName(name),
		langfuse.WithToolInput(args),
		langfuse.WithToolMetadata(metadata),
	)
	ctx, _ = tool.Observation(ctx)

	return &toolObservationWrapper{
		context:     ctx,
		observation: tool,
	}
}

func (ce *customExecutor) createAgentObservation(ctx context.Context, name string, args json.RawMessage) observationWrapper {
	ctx, observation := obs.Observer.NewObservation(ctx)
	metadata := langfuse.Metadata{
		"agent_name":    name,
		"tool_category": GetToolType(name).String(),
		"flow_id":       ce.flowID,
	}
	if ce.taskID != nil {
		metadata["task_id"] = *ce.taskID
	}
	if ce.subtaskID != nil {
		metadata["subtask_id"] = *ce.subtaskID
	}

	agent := observation.Agent(
		langfuse.WithAgentName(name),
		langfuse.WithAgentInput(args),
		langfuse.WithAgentMetadata(metadata),
	)
	ctx, _ = agent.Observation(ctx)

	return &agentObservationWrapper{
		context:     ctx,
		observation: agent,
	}
}

func (ce *customExecutor) createSpanObservation(ctx context.Context, name string, args json.RawMessage) observationWrapper {
	ctx, observation := obs.Observer.NewObservation(ctx)
	metadata := langfuse.Metadata{
		"barrier_name":  name,
		"tool_category": GetToolType(name).String(),
		"flow_id":       ce.flowID,
	}
	if ce.taskID != nil {
		metadata["task_id"] = *ce.taskID
	}
	if ce.subtaskID != nil {
		metadata["subtask_id"] = *ce.subtaskID
	}

	span := observation.Span(
		langfuse.WithSpanName(name),
		langfuse.WithSpanInput(args),
		langfuse.WithSpanMetadata(metadata),
	)
	ctx, _ = span.Observation(ctx)

	return &spanObservationWrapper{
		context:     ctx,
		observation: span,
	}
}

func (ce *customExecutor) Execute(
	ctx context.Context,
	streamID int64,
	id, name, thinking string,
	args json.RawMessage,
) (string, error) {
	startTime := time.Now()

	handler, ok := ce.handlers[name]
	if !ok {
		return fmt.Sprintf("function '%s' not found in available tools list", name), nil
	}

	var raw any
	if err := json.Unmarshal(args, &raw); err != nil {
		return fmt.Sprintf("failed to unmarshal '%s' tool call arguments: %v: fix it", name, err), nil
	}

	// Create observation based on tool type
	toolType := GetToolType(name)
	var obsWrapper observationWrapper

	switch toolType {
	case EnvironmentToolType, SearchNetworkToolType, StoreAgentResultToolType, StoreVectorDbToolType:
		obsWrapper = ce.createToolObservation(ctx, name, args)
	case AgentToolType:
		obsWrapper = ce.createAgentObservation(ctx, name, args)
	case BarrierToolType:
		obsWrapper = ce.createSpanObservation(ctx, name, args)
	case SearchVectorDbToolType:
		// Skip - handlers create RETRIEVER internally
		obsWrapper = &noopObservationWrapper{context: ctx}
	default:
		// Unknown type - use no-op wrapper
		obsWrapper = &noopObservationWrapper{context: ctx}
	}

	// Use context from observation wrapper
	ctx = obsWrapper.ctx()

	var err error
	msgID, msg := int64(0), ce.getMessage(args)
	if msg != "" {
		msgType := getMessageType(name)
		msgID, err = ce.mlp.PutMsg(ctx, msgType, ce.taskID, ce.subtaskID, streamID, thinking, msg)
		if err != nil {
			return "", err
		}
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
		obsWrapper.end("", err, time.Since(startTime).Seconds())
		return "", fmt.Errorf("failed to create toolcall: %w", err)
	}

	wrapHandler := func(ctx context.Context, name string, args json.RawMessage) (string, database.MsglogResultFormat, error) {
		resultFormat := getMessageResultFormat(name)
		result, err := handler(ctx, name, args)
		if err != nil {
			durationDelta := time.Since(startTime).Seconds()
			_, _ = ce.db.UpdateToolcallFailedResult(ctx, database.UpdateToolcallFailedResultParams{
				Result:          fmt.Sprintf("failed to execute handler: %s", err.Error()),
				DurationSeconds: durationDelta,
				ID:              tc.ID,
			})
			return "", resultFormat, fmt.Errorf("failed to execute handler: %w", err)
		}

		result = database.SanitizeUTF8(result)
		allowSummarize := slices.Contains(allowedSummarizingToolsResult, name)
		if ce.summarizer != nil && allowSummarize && len(result) > DefaultResultSizeLimit {
			summarizePrompt, err := ce.getSummarizePrompt(name, string(args), result)
			if err != nil {
				return "", resultFormat, fmt.Errorf("failed to get summarize prompt: %w", err)
			}
			result, err = ce.summarizer(ctx, summarizePrompt)
			if err != nil {
				durationDelta := time.Since(startTime).Seconds()
				_, _ = ce.db.UpdateToolcallFailedResult(ctx, database.UpdateToolcallFailedResultParams{
					Result:          fmt.Sprintf("failed to summarize result: %s", err.Error()),
					DurationSeconds: durationDelta,
					ID:              tc.ID,
				})
				return "", resultFormat, fmt.Errorf("failed to summarize result: %w", err)
			}
			resultFormat = database.MsglogResultFormatMarkdown
		}

		durationDelta := time.Since(startTime).Seconds()
		_, err = ce.db.UpdateToolcallFinishedResult(ctx, database.UpdateToolcallFinishedResultParams{
			Result:          result,
			DurationSeconds: durationDelta,
			ID:              tc.ID,
		})
		if err != nil {
			return "", resultFormat, fmt.Errorf("failed to update toolcall result: %w", err)
		}

		return result, resultFormat, nil
	}

	if msg == "" { // no arg message to log and execute handler immediately
		result, _, err := wrapHandler(ctx, name, args)
		obsWrapper.end(result, err, time.Since(startTime).Seconds())
		return result, err
	}

	result, resultFormat, err := wrapHandler(ctx, name, args)
	if err != nil {
		obsWrapper.end(result, err, time.Since(startTime).Seconds())
		return "", err
	}

	if err := ce.storeToolResult(ctx, name, result, args); err != nil {
		obsWrapper.end(result, err, time.Since(startTime).Seconds())
		return "", fmt.Errorf("failed to store tool result in long-term memory: %w", err)
	}

	if err := ce.mlp.UpdateMsgResult(ctx, msgID, streamID, result, resultFormat); err != nil {
		obsWrapper.end(result, err, time.Since(startTime).Seconds())
		return "", err
	}

	obsWrapper.end(result, nil, time.Since(startTime).Seconds())

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

func (ce *customExecutor) GetBarrierTools() []FunctionInfo {
	tools := make([]FunctionInfo, 0, len(ce.barriers))
	for name := range ce.barriers {
		schema, err := ce.GetToolSchema(name)
		if err != nil {
			continue
		}
		schemaJSON, err := json.Marshal(schema)
		if err != nil {
			continue
		}
		tools = append(tools, FunctionInfo{Name: name, Schema: string(schemaJSON)})
	}
	return tools
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

func (ce *customExecutor) getSummarizePrompt(funcName, funcArgs, result string) (string, error) {
	templateText := `<instructions>
TASK: Summarize the execution result from '{{.FuncName}}' function call

DATA:
- <function> contains structured information about the function call
- <arguments> contains the parameters passed to the function
- <schema> contains the JSON schema of the function parameters
- <result> contains the raw output that NEEDS summarization

REQUIREMENTS:
1. Create a focused summary (max {{.MaxLength}} chars) that preserves critical information
2. Keep all actionable insights, technical details, and information relevant to the function's purpose
3. Preserve exact error messages, file paths, URLs, commands, and technical terminology
4. Structure information logically with appropriate formatting (headings, bullet points)
5. Begin with what the function accomplished or attempted

The summary must provide the same practical value as the original while being concise.
</instructions>

<function name="{{.FuncName}}">
<arguments>
{{.FormattedArgs}}
</arguments>
<schema>
{{.SchemaJSON}}
</schema>
</function>

<result>
{{.Result}}
</result>`

	var argsMap map[string]interface{}
	if err := json.Unmarshal([]byte(funcArgs), &argsMap); err != nil {
		return "", fmt.Errorf("failed to parse function arguments: %w", err)
	}

	var formattedArgs strings.Builder
	for key, value := range argsMap {
		strValue := fmt.Sprintf("%v", value)
		if len(strValue) > maxArgValueLength {
			strValue = strValue[:maxArgValueLength] + "... [truncated]"
		}
		formattedArgs.WriteString(fmt.Sprintf("%s: %s\n", key, strValue))
	}

	var schemaJSON string
	schemaObj, err := ce.GetToolSchema(funcName)
	if err == nil && schemaObj != nil {
		schemaBytes, err := json.MarshalIndent(schemaObj, "", "  ")
		if err == nil {
			schemaJSON = string(schemaBytes)
		}
	}

	templateContext := map[string]interface{}{
		"FuncName":      funcName,
		"FormattedArgs": formattedArgs.String(),
		"SchemaJSON":    schemaJSON,
		"Result":        result,
		"MaxLength":     DefaultResultSizeLimit / 2,
	}

	tmpl, err := template.New("summarize").Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("error creating template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateContext); err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}

	return buf.String(), nil
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
	buffer.WriteString(fmt.Sprintf("### Incoming arguments\n\n```json\n%s\n```\n\n", args))
	buffer.WriteString(fmt.Sprintf("#### Tool result\n\n%s\n\n", result))
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
		data := map[string]any{
			"doc_type":  memoryVectorStoreDefaultType,
			"tool_name": name,
			"flow_id":   ce.flowID,
		}
		if ce.taskID != nil {
			data["task_id"] = *ce.taskID
		}
		if ce.subtaskID != nil {
			data["subtask_id"] = *ce.subtaskID
		}
		filtersData, err := json.Marshal(data)
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
