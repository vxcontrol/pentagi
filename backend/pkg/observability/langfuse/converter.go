package langfuse

import (
	"encoding/json"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

// convertInput converts various input formats to Langfuse-compatible format.
func convertInput(input any, tools []llms.Tool) any {
	switch v := input.(type) {
	case nil:
		return nil
	case []*llms.MessageContent:
		return convertChain(v, tools)
	case []llms.MessageContent:
		msgChain := make([]*llms.MessageContent, 0, len(v))
		for _, message := range v {
			msgChain = append(msgChain, &message)
		}
		return convertChain(msgChain, tools)
	default:
		return input
	}
}

func convertChain(chain []*llms.MessageContent, tools []llms.Tool) any {
	// Build mapping of tool_call_id -> function_name for tool responses
	toolCallNames := buildToolCallMapping(chain)

	msgChain := make([]any, 0, len(chain))
	for _, message := range chain {
		msgChain = append(msgChain, convertMessageWithContext(message, toolCallNames))
	}

	if len(tools) > 0 {
		return map[string]any{
			"tools":    tools,
			"messages": msgChain,
		}
	}

	return msgChain
}

// buildToolCallMapping creates a map of tool_call_id -> function_name
// by scanning all tool calls in the chain
func buildToolCallMapping(chain []*llms.MessageContent) map[string]string {
	mapping := make(map[string]string)

	for _, message := range chain {
		if message == nil {
			continue
		}

		for _, part := range message.Parts {
			switch p := part.(type) {
			case *llms.ToolCall:
				if p.FunctionCall != nil {
					mapping[p.ID] = p.FunctionCall.Name
				}
			case llms.ToolCall:
				if p.FunctionCall != nil {
					mapping[p.ID] = p.FunctionCall.Name
				}
			}
		}
	}

	return mapping
}

// convertMessageWithContext is like convertMessage but with access to tool call names
func convertMessageWithContext(message *llms.MessageContent, toolCallNames map[string]string) any {
	if message == nil {
		return nil
	}

	role := mapRole(message.Role)

	result := map[string]any{
		"role": role,
	}

	// Extract thinking content
	var thinking []any
	for _, part := range message.Parts {
		if convertedThinking := convertPartWithThinking(part); convertedThinking != nil {
			thinking = append(thinking, convertedThinking)
		}
	}

	// Handle tool role specially (tool responses)
	if role == "tool" {
		return convertToolMessageWithNames(message, result, toolCallNames)
	}

	// Separate parts by type
	var textParts []string
	var toolCalls []any
	var contentArray []any // For multimodal content (images, etc.)
	hasMultimodal := false

	for _, part := range message.Parts {
		switch p := part.(type) {
		case *llms.TextContent:
			textParts = append(textParts, p.Text)
		case llms.TextContent:
			textParts = append(textParts, p.Text)
		case *llms.ToolCall:
			if tc := convertToolCallToOpenAI(p); tc != nil {
				toolCalls = append(toolCalls, tc)
			}
		case llms.ToolCall:
			if tc := convertToolCallToOpenAI(&p); tc != nil {
				toolCalls = append(toolCalls, tc)
			}
		case *llms.ImageURLContent, llms.ImageURLContent,
			*llms.BinaryContent, llms.BinaryContent:
			hasMultimodal = true
			if converted := convertMultimodalPart(part); converted != nil {
				contentArray = append(contentArray, converted)
			}
		}
	}

	// Build content field
	if hasMultimodal {
		// For multimodal: content is array of parts
		for _, text := range textParts {
			contentArray = append([]any{map[string]any{
				"type": "text",
				"text": text,
			}}, contentArray...)
		}
		if len(contentArray) > 0 {
			result["content"] = contentArray
		}
	} else if len(textParts) > 0 {
		// For text-only: content is string
		result["content"] = joinTextParts(textParts)
	} else if len(toolCalls) > 0 {
		// Tool calls without text content
		result["content"] = ""
	}

	// Add tool_calls array if present
	if len(toolCalls) > 0 {
		result["tool_calls"] = toolCalls
	}

	// Add thinking if present
	if len(thinking) > 0 {
		result["thinking"] = thinking
	}

	return result
}

// convertMessage converts a single message to OpenAI format.
// Role mapping:
// - "human" → "user"
// - "ai" → "assistant"
// - "system" remains "system"
// - "tool" remains "tool"
// convertMessage converts a single message without tool call name context.
// Used for single message outputs where we don't have the full chain.
func convertMessage(message *llms.MessageContent) any {
	return convertMessageWithContext(message, make(map[string]string))
}

func mapRole(role llms.ChatMessageType) string {
	switch role {
	case llms.ChatMessageTypeHuman:
		return "user"
	case llms.ChatMessageTypeAI:
		return "assistant"
	case llms.ChatMessageTypeSystem:
		return "system"
	case llms.ChatMessageTypeTool:
		return "tool"
	case llms.ChatMessageTypeGeneric:
		return "assistant" // fallback to assistant
	default:
		return string(role)
	}
}

// convertToolMessageWithNames handles tool role messages (tool responses)
// with access to tool call name mapping
func convertToolMessageWithNames(message *llms.MessageContent, result map[string]any, toolCallNames map[string]string) any {
	var toolCallID string
	var content any

	for _, part := range message.Parts {
		switch p := part.(type) {
		case *llms.ToolCallResponse:
			toolCallID = p.ToolCallID
			content = parseToolContent(p.Content)
		case llms.ToolCallResponse:
			toolCallID = p.ToolCallID
			content = parseToolContent(p.Content)
		case *llms.TextContent:
			if content == nil {
				content = p.Text
			}
		case llms.TextContent:
			if content == nil {
				content = p.Text
			}
		}
	}

	result["tool_call_id"] = toolCallID

	// Add function name from mapping (makes UI more readable)
	if functionName, ok := toolCallNames[toolCallID]; ok {
		result["name"] = functionName
	}

	// Keep content as object if it's complex (for rich table rendering)
	// OpenAI format expects content as string or object
	result["content"] = content

	return result
}

// parseToolContent tries to parse JSON content to object for rich rendering.
// If parsing fails or content is simple, returns as string.
func parseToolContent(content string) any {
	if content == "" {
		return ""
	}

	// Try to parse as JSON
	var parsedContent any
	if err := json.Unmarshal([]byte(content), &parsedContent); err != nil {
		// Not JSON, return as string
		return content
	}

	// Check if it's a rich object (3+ keys or nested structure)
	if obj, ok := parsedContent.(map[string]any); ok {
		if isRichObject(obj) {
			// Return as object for table rendering
			return obj
		}
	}

	// For arrays or simple objects, keep as parsed JSON
	// (could be stringified again, but this allows Langfuse to decide)
	return parsedContent
}

// isRichObject checks if object should be rendered as table.
// Rich = 3+ keys OR nested structure (objects/arrays).
func isRichObject(obj map[string]any) bool {
	// More than 2 keys → rich
	if len(obj) > 2 {
		return true
	}

	// Check for nested structures
	for _, value := range obj {
		switch value.(type) {
		case map[string]any, []any:
			return true // Has nested structure
		}
	}

	// 1-2 keys with scalar values → simple
	return false
}

// convertToolCallToOpenAI converts ToolCall to OpenAI format:
// {id: "call_123", type: "function", function: {name: "...", arguments: "..."}}
func convertToolCallToOpenAI(toolCall *llms.ToolCall) any {
	if toolCall == nil || toolCall.FunctionCall == nil {
		return nil
	}

	// Arguments should be a JSON string in OpenAI format
	arguments := toolCall.FunctionCall.Arguments
	if arguments == "" {
		arguments = "{}"
	}

	return map[string]any{
		"id":   toolCall.ID,
		"type": "function",
		"function": map[string]any{
			"name":      toolCall.FunctionCall.Name,
			"arguments": arguments,
		},
	}
}

// convertMultimodalPart converts image/binary content for multimodal messages
func convertMultimodalPart(part llms.ContentPart) any {
	switch p := part.(type) {
	case *llms.ImageURLContent:
		imageURL := map[string]any{
			"url": p.URL,
		}
		if p.Detail != "" {
			imageURL["detail"] = p.Detail
		}
		return map[string]any{
			"type":      "image_url",
			"image_url": imageURL,
		}
	case llms.ImageURLContent:
		imageURL := map[string]any{
			"url": p.URL,
		}
		if p.Detail != "" {
			imageURL["detail"] = p.Detail
		}
		return map[string]any{
			"type":      "image_url",
			"image_url": imageURL,
		}
	case *llms.BinaryContent:
		return map[string]any{
			"type": "binary",
			"binary": map[string]any{
				"mime_type": p.MIMEType,
				"data":      p.Data,
			},
		}
	case llms.BinaryContent:
		return map[string]any{
			"type": "binary",
			"binary": map[string]any{
				"mime_type": p.MIMEType,
				"data":      p.Data,
			},
		}
	}
	return nil
}

// joinTextParts joins multiple text parts into a single string
func joinTextParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}

	// Join with space or newline as separator
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += " "
		}
		result += part
	}
	return result
}

func convertPartWithThinking(thinking llms.ContentPart) any {
	switch p := thinking.(type) {
	case *llms.TextContent:
		return convertThinking(p.Reasoning)
	case llms.TextContent:
		return convertThinking(p.Reasoning)
	case *llms.ToolCall:
		return convertThinking(p.Reasoning)
	case llms.ToolCall:
		return convertThinking(p.Reasoning)
	default:
		return nil
	}
}

func convertThinking(thinking *reasoning.ContentReasoning) any {
	if thinking.IsEmpty() || thinking.Content == "" {
		return nil
	}

	return map[string]any{
		"type":    "thinking",
		"content": thinking.Content,
	}
}

// convertOutput converts various output formats to Langfuse-compatible format.
func convertOutput(output any) any {
	switch v := output.(type) {
	case nil:
		return nil
	case *llms.MessageContent:
		return convertMessage(v)
	case llms.MessageContent:
		return convertMessage(&v)
	case []*llms.MessageContent:
		return convertInput(v, nil)
	case []llms.MessageContent:
		return convertInput(v, nil)
	case *llms.ContentChoice:
		return convertChoice(v)
	case llms.ContentChoice:
		return convertChoice(&v)
	case []*llms.ContentChoice:
		switch len(v) {
		case 0:
			return nil
		case 1:
			return convertChoice(v[0])
		default:
			choices := make([]any, 0, len(v))
			for _, choice := range v {
				choices = append(choices, convertChoice(choice))
			}
			return choices
		}
	case []llms.ContentChoice:
		switch len(v) {
		case 0:
			return nil
		case 1:
			return convertChoice(&v[0])
		default:
			choices := make([]any, 0, len(v))
			for _, choice := range v {
				choices = append(choices, convertChoice(&choice))
			}
			return choices
		}
	default:
		return output
	}
}

func convertChoice(choice *llms.ContentChoice) any {
	if choice == nil {
		return nil
	}

	result := map[string]any{
		"role": "assistant",
	}

	// Add thinking if present
	var thinking []any
	if convertedThinking := convertThinking(choice.Reasoning); convertedThinking != nil {
		thinking = append(thinking, convertedThinking)
	}

	// Add content
	if choice.Content != "" {
		result["content"] = choice.Content
	} else if len(choice.ToolCalls) > 0 {
		// Tool calls without content
		result["content"] = ""
	}

	// Convert tool calls to OpenAI format
	var toolCalls []any
	for _, toolCall := range choice.ToolCalls {
		if tc := convertToolCallToOpenAI(&toolCall); tc != nil {
			toolCalls = append(toolCalls, tc)
		}
	}

	// Handle legacy FuncCall (convert to tool call format if ToolCalls is empty)
	if choice.FuncCall != nil && len(choice.ToolCalls) == 0 {
		arguments := choice.FuncCall.Arguments
		if arguments == "" {
			arguments = "{}"
		}

		toolCalls = append(toolCalls, map[string]any{
			"id":   "legacy_func_call",
			"type": "function",
			"function": map[string]any{
				"name":      choice.FuncCall.Name,
				"arguments": arguments,
			},
		})
	}

	// Add tool_calls array if present
	if len(toolCalls) > 0 {
		result["tool_calls"] = toolCalls
	}

	// Add thinking if present
	if len(thinking) > 0 {
		result["thinking"] = thinking
	}

	return result
}
