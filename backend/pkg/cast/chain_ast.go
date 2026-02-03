package cast

import (
	"fmt"
	"pentagi/pkg/templates"
	"sort"
	"strings"

	"github.com/vxcontrol/langchaingo/llms"
)

// Constants for common operations in chainAST
const (
	fallbackRequestArgs     = `{}`
	FallbackResponseContent = "the call was not handled, please try again"
	SummarizationToolName   = "execute_task_and_return_summary"
	SummarizationToolArgs   = `{"question": "delegate and execute the task, then return the summary of the result"}`
	ToolCallIDTemplate      = "call_{r:24:x}"
)

// BodyPairType represents the type of body pair in the chain
type BodyPairType int

const (
	// RequestResponse represents an AI message with one or more tool calls and their responses
	RequestResponse BodyPairType = iota
	// Completion represents an AI message without tool calls
	Completion
	// Summarization represents a summarization task
	Summarization
)

// ChainAST represents a message chain as an abstract syntax tree
type ChainAST struct {
	Sections []*ChainSection
}

// ChainSection represents a section of the chain starting with a header
// and containing body pairs
type ChainSection struct {
	Header    *Header
	Body      []*BodyPair
	sizeBytes int // Total size of the section in bytes
}

// Header represents the header of a chain section
// It can contain a system message, a human message, or both
type Header struct {
	SystemMessage *llms.MessageContent
	HumanMessage  *llms.MessageContent
	sizeBytes     int // Total size of the header in bytes
}

// BodyPair represents a pair of AI and Tool messages
type BodyPair struct {
	Type         BodyPairType
	AIMessage    *llms.MessageContent
	ToolMessages []*llms.MessageContent // Can be empty for Completion type
	sizeBytes    int                    // Size of this body pair in bytes
}

// ToolCallPair tracks tool calls and responses
type ToolCallPair struct {
	ToolCall llms.ToolCall
	Response llms.ToolCallResponse
}

// ToolCallsInfo tracks tool calls and responses
type ToolCallsInfo struct {
	PendingToolCallIDs   []string
	UnmatchedToolCallIDs []string
	PendingToolCalls     map[string]*ToolCallPair
	CompletedToolCalls   map[string]*ToolCallPair
	UnmatchedToolCalls   map[string]*ToolCallPair
}

// NewChainAST creates a new ChainAST from a message chain
// If force is true, it will attempt to fix inconsistencies in the chain
func NewChainAST(chain []llms.MessageContent, force bool) (*ChainAST, error) {
	if len(chain) == 0 {
		return &ChainAST{}, nil
	}

	ast := &ChainAST{
		Sections: []*ChainSection{},
	}

	var currentSection *ChainSection
	var currentHeader *Header
	var currentBodyPair *BodyPair

	// Check if the chain starts with a valid message type
	if len(chain) > 0 && chain[0].Role != llms.ChatMessageTypeSystem && chain[0].Role != llms.ChatMessageTypeHuman {
		return nil, fmt.Errorf("unexpected chain begin: first message must be System or Human, got %s", chain[0].Role)
	}

	// Validate that there are no pending tool calls in the current section
	checkAndFixPendingToolCalls := func() error {
		if currentBodyPair == nil || currentBodyPair.Type == Completion {
			return nil
		}

		toolCallsInfo := currentBodyPair.GetToolCallsInfo()
		if len(toolCallsInfo.PendingToolCallIDs) > 0 {
			if !force {
				pendingToolCallIDs := strings.Join(toolCallsInfo.PendingToolCallIDs, ", ")
				return fmt.Errorf("tool calls with IDs [%s] have no response", pendingToolCallIDs)
			}
			for _, toolCallID := range toolCallsInfo.PendingToolCallIDs {
				toolCallPair := toolCallsInfo.PendingToolCalls[toolCallID]
				currentBodyPair.ToolMessages = append(currentBodyPair.ToolMessages, &llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{llms.ToolCallResponse{
						ToolCallID: toolCallID,
						Name:       toolCallPair.ToolCall.FunctionCall.Name,
						Content:    FallbackResponseContent,
					}},
				})
			}
		}

		return nil
	}

	checkAndFixUnmatchedToolCalls := func() error {
		if currentBodyPair == nil || currentBodyPair.Type == Completion {
			return nil
		}

		toolCallsInfo := currentBodyPair.GetToolCallsInfo()
		if len(toolCallsInfo.UnmatchedToolCallIDs) > 0 {
			if !force {
				unmatchedToolCallIDs := strings.Join(toolCallsInfo.UnmatchedToolCallIDs, ", ")
				return fmt.Errorf("tool calls with IDs [%s] have no response", unmatchedToolCallIDs)
			}
			// Try to add a fallback request for each unmatched tool call
			for _, toolCallID := range toolCallsInfo.UnmatchedToolCallIDs {
				toolCallResponse := toolCallsInfo.UnmatchedToolCalls[toolCallID].Response
				currentBodyPair.AIMessage.Parts = append(currentBodyPair.AIMessage.Parts, llms.ToolCall{
					ID: toolCallID,
					FunctionCall: &llms.FunctionCall{
						Name:      toolCallResponse.Name,
						Arguments: fallbackRequestArgs,
					},
				})
			}
		}

		return nil
	}

	for _, msg := range chain {

		switch msg.Role {
		case llms.ChatMessageTypeSystem:
			// System message should only appear at the beginning of a section
			if currentSection != nil {
				return nil, fmt.Errorf("unexpected system message in the middle of a chain")
			}

			// Start a new section with a system message
			systemMsgCopy := msg // Create a copy to avoid reference issues
			currentHeader = NewHeader(&systemMsgCopy, nil)
			currentSection = NewChainSection(currentHeader, []*BodyPair{})
			ast.AddSection(currentSection)
			currentBodyPair = nil

		case llms.ChatMessageTypeHuman:
			// Handle normal case for human messages
			humanMsgCopy := msg // Create a copy to avoid reference issues

			if currentSection != nil && currentSection.Header.HumanMessage != nil {
				// If we already have a human message in this section, start a new one or append to the existing one
				if len(currentSection.Body) == 0 {
					if !force {
						return nil, fmt.Errorf("double human messages in the middle of a chain")
					}
					// Merge parts of the human message with the existing one
					currentSection.Header.HumanMessage.Parts = append(currentSection.Header.HumanMessage.Parts, humanMsgCopy.Parts...)
					msgSize := CalculateMessageSize(&humanMsgCopy)
					currentSection.Header.sizeBytes += msgSize
					currentSection.sizeBytes += msgSize
				} else {
					currentHeader = NewHeader(nil, &humanMsgCopy)
					currentSection = NewChainSection(currentHeader, []*BodyPair{})
					ast.AddSection(currentSection)
					if err := checkAndFixPendingToolCalls(); err != nil {
						return nil, err
					}
					currentBodyPair = nil
				}
			} else if currentSection != nil && currentSection.Header.HumanMessage == nil {
				// If we already have an opening section without a human message, try to set it
				if len(currentSection.Body) != 0 && !force {
					return nil, fmt.Errorf("got human message after AI message in the middle of a chain")
				}
				currentSection.SetHeader(NewHeader(currentSection.Header.SystemMessage, &humanMsgCopy))
			} else {
				// No section set yet, add this one
				currentHeader = NewHeader(nil, &humanMsgCopy)
				currentSection = NewChainSection(currentHeader, []*BodyPair{})
				ast.AddSection(currentSection)
				if err := checkAndFixPendingToolCalls(); err != nil {
					return nil, err
				}
				currentBodyPair = nil
			}

		case llms.ChatMessageTypeAI:
			// Ensure we have a section to add this AI message to
			if currentSection == nil {
				return nil, fmt.Errorf("unexpected AI message without a preceding header")
			}

			// Ensure that there are no pending tool calls in the current section before adding the AI message
			if err := checkAndFixPendingToolCalls(); err != nil {
				return nil, err
			}

			// Prepare the AI message for the body pair
			aiMsgCopy := msg // Create a copy to avoid reference issues
			currentBodyPair = NewBodyPair(&aiMsgCopy, []*llms.MessageContent{})
			currentSection.AddBodyPair(currentBodyPair)

		case llms.ChatMessageTypeTool:
			// Ensure we have a section to add this tool message to
			if currentSection == nil {
				return nil, fmt.Errorf("unexpected tool message without a preceding header")
			}

			// Ensure we have a body pair to add this tool message to
			if currentBodyPair == nil || currentBodyPair.Type == Completion {
				if !force {
					return nil, fmt.Errorf("unexpected tool message without a preceding AI message with tool calls")
				}
				// If force is true and we don't have a proper body pair, skip this message
				continue
			}

			// Add this tool message to the current body pair
			toolMsgCopy := msg // Create a copy to avoid reference issues

			currentBodyPair.ToolMessages = append(currentBodyPair.ToolMessages, &toolMsgCopy)
			if err := checkAndFixUnmatchedToolCalls(); err != nil {
				return nil, err
			}

			// Update sizes
			toolMsgSize := CalculateMessageSize(&toolMsgCopy)
			currentBodyPair.sizeBytes += toolMsgSize
			currentSection.sizeBytes += toolMsgSize

		default:
			return nil, fmt.Errorf("unexpected message role: %s", msg.Role)
		}
	}

	// Check if there are any pending tool calls in the last section
	if err := checkAndFixPendingToolCalls(); err != nil {
		return nil, err
	}

	return ast, nil
}

// Messages returns the ChainAST as a message chain (renamed from Dump)
func (ast *ChainAST) Messages() []llms.MessageContent {
	if len(ast.Sections) == 0 {
		return []llms.MessageContent{}
	}

	var result []llms.MessageContent

	for _, section := range ast.Sections {
		// Add all messages from the section
		sectionMessages := section.Messages()
		result = append(result, sectionMessages...)
	}

	return result
}

// Messages returns all messages in the section in order: header messages followed by body pairs
func (section *ChainSection) Messages() []llms.MessageContent {
	var messages []llms.MessageContent

	// Add header messages
	headerMessages := section.Header.Messages()
	messages = append(messages, headerMessages...)

	// Add body pair messages
	for _, pair := range section.Body {
		pairMessages := pair.Messages()
		messages = append(messages, pairMessages...)
	}

	return messages
}

// Messages returns all messages in the header (system and human)
func (header *Header) Messages() []llms.MessageContent {
	var messages []llms.MessageContent

	// Add system message if present
	if header.SystemMessage != nil {
		messages = append(messages, *header.SystemMessage)
	}

	// Add human message if present
	if header.HumanMessage != nil {
		messages = append(messages, *header.HumanMessage)
	}

	return messages
}

// Messages returns all messages in the body pair (AI and Tool messages)
func (pair *BodyPair) Messages() []llms.MessageContent {
	var messages []llms.MessageContent

	// Add AI message
	if pair.AIMessage != nil {
		messages = append(messages, *pair.AIMessage)
	}

	// Add all tool messages
	for _, toolMsg := range pair.ToolMessages {
		messages = append(messages, *toolMsg)
	}

	return messages
}

// GetToolCallsInfo returns the tool calls info for the body pair
func (pair *BodyPair) GetToolCallsInfo() ToolCallsInfo {
	pendingToolCalls := make(map[string]*ToolCallPair)
	completedToolCalls := make(map[string]*ToolCallPair)
	unmatchedToolCalls := make(map[string]*ToolCallPair)

	for _, part := range pair.AIMessage.Parts {
		if toolCall, ok := part.(llms.ToolCall); ok && toolCall.FunctionCall != nil {
			pendingToolCalls[toolCall.ID] = &ToolCallPair{
				ToolCall: toolCall,
			}
		}
	}
	for _, toolMsg := range pair.ToolMessages {
		for _, part := range toolMsg.Parts {
			if resp, ok := part.(llms.ToolCallResponse); ok {
				toolCallPair, ok := pendingToolCalls[resp.ToolCallID]
				if !ok {
					unmatchedToolCalls[resp.ToolCallID] = &ToolCallPair{
						Response: resp,
					}
				} else {
					toolCallPair.Response = resp
					delete(pendingToolCalls, resp.ToolCallID)
					completedToolCalls[resp.ToolCallID] = toolCallPair
				}
			}
		}
	}

	pendingToolCallIDs := make([]string, 0, len(pendingToolCalls))
	for toolCallID := range pendingToolCalls {
		pendingToolCallIDs = append(pendingToolCallIDs, toolCallID)
	}
	sort.Strings(pendingToolCallIDs)

	unmatchedToolCallIDs := make([]string, 0, len(unmatchedToolCalls))
	for toolCallID := range unmatchedToolCalls {
		unmatchedToolCallIDs = append(unmatchedToolCallIDs, toolCallID)
	}
	sort.Strings(unmatchedToolCallIDs)

	return ToolCallsInfo{
		PendingToolCallIDs:   pendingToolCallIDs,
		UnmatchedToolCallIDs: unmatchedToolCallIDs,
		PendingToolCalls:     pendingToolCalls,
		CompletedToolCalls:   completedToolCalls,
		UnmatchedToolCalls:   unmatchedToolCalls,
	}
}

func (pair *BodyPair) IsValid() bool {
	if pair.Type != Completion && pair.Type != RequestResponse && pair.Type != Summarization {
		return false
	}

	if pair.Type == Completion && len(pair.ToolMessages) != 0 {
		return false
	}

	if pair.Type == RequestResponse && len(pair.ToolMessages) == 0 {
		return false
	}

	if pair.Type == Summarization && len(pair.ToolMessages) != 1 {
		return false
	}

	toolCallsInfo := pair.GetToolCallsInfo()
	if len(toolCallsInfo.PendingToolCalls) != 0 || len(toolCallsInfo.UnmatchedToolCalls) != 0 {
		return false
	}

	return true
}

// NewHeader creates a new Header with automatic size calculation
func NewHeader(systemMsg *llms.MessageContent, humanMsg *llms.MessageContent) *Header {
	header := &Header{
		SystemMessage: systemMsg,
		HumanMessage:  humanMsg,
	}

	// Calculate size
	header.sizeBytes = 0
	if systemMsg != nil {
		header.sizeBytes += CalculateMessageSize(systemMsg)
	}
	if humanMsg != nil {
		header.sizeBytes += CalculateMessageSize(humanMsg)
	}

	return header
}

// NewBodyPair creates a new BodyPair from an AI message and optional tool messages
// It auto determines the type (Completion or RequestResponse or Summarization) based on content
func NewBodyPair(aiMsg *llms.MessageContent, toolMsgs []*llms.MessageContent) *BodyPair {
	// Determine the type based on whether there are tool calls in the AI message
	pairType := Completion

	if aiMsg != nil {
		partsToDelete := make([]int, 0)
		for id, part := range aiMsg.Parts {
			if toolCall, isToolCall := part.(llms.ToolCall); isToolCall {
				if toolCall.FunctionCall == nil {
					partsToDelete = append(partsToDelete, id)
					continue
				} else if toolCall.FunctionCall.Name == SummarizationToolName {
					pairType = Summarization
				} else {
					pairType = RequestResponse
				}
				break
			}
		}
		for _, id := range partsToDelete {
			aiMsg.Parts = append(aiMsg.Parts[:id], aiMsg.Parts[id+1:]...)
		}
	}

	// Create the body pair
	pair := &BodyPair{
		Type:         pairType,
		AIMessage:    aiMsg,
		ToolMessages: toolMsgs,
	}

	// Calculate size
	pair.sizeBytes = CalculateBodyPairSize(pair)

	return pair
}

// NewBodyPairFromMessages creates a new BodyPair from a slice of messages
// The first message should be an AI message, followed by optional tool messages
func NewBodyPairFromMessages(messages []llms.MessageContent) (*BodyPair, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("cannot create body pair from empty message slice")
	}

	// The first message must be an AI message
	if messages[0].Role != llms.ChatMessageTypeAI {
		return nil, fmt.Errorf("first message in body pair must be an AI message")
	}

	aiMsg := &messages[0]
	var toolMsgs []*llms.MessageContent

	// Remaining messages should be tool messages
	for i := 1; i < len(messages); i++ {
		if messages[i].Role != llms.ChatMessageTypeTool {
			return nil, fmt.Errorf("non-tool message found in body pair at position %d", i)
		}

		msg := messages[i] // Create a copy to avoid reference issues
		toolMsgs = append(toolMsgs, &msg)
	}

	return NewBodyPair(aiMsg, toolMsgs), nil
}

// NewBodyPairFromSummarization creates a new BodyPair from a summarization tool call
func NewBodyPairFromSummarization(text string, tcIDTemplate string) *BodyPair {
	toolCallID := templates.GenerateFromPattern(tcIDTemplate)
	return NewBodyPair(
		&llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.ToolCall{
					ID:   toolCallID,
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      SummarizationToolName,
						Arguments: SummarizationToolArgs,
					},
				},
			},
		},
		[]*llms.MessageContent{
			{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: toolCallID,
						Name:       SummarizationToolName,
						Content:    text,
					},
				},
			},
		},
	)
}

// NewBodyPairFromCompletion creates a new Completion body pair with the given text
func NewBodyPairFromCompletion(text string) *BodyPair {
	return NewBodyPair(
		&llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: text},
			},
		},
		nil,
	)
}

// NewChainSection creates a new ChainSection with automatic size calculation
func NewChainSection(header *Header, bodyPairs []*BodyPair) *ChainSection {
	section := &ChainSection{
		Header: header,
		Body:   bodyPairs,
	}

	// Calculate section size
	section.sizeBytes = header.Size()
	for _, pair := range bodyPairs {
		section.sizeBytes += pair.Size()
	}

	return section
}

func (bpt BodyPairType) String() string {
	switch bpt {
	case Completion:
		return "completion"
	case RequestResponse:
		return "request-response"
	case Summarization:
		return "summarization"
	default:
		return "unknown"
	}
}

// Size returns the size of the header in bytes
func (header *Header) Size() int {
	return header.sizeBytes
}

// SetHeader sets the header of the section
func (section *ChainSection) SetHeader(header *Header) {
	section.sizeBytes -= section.Header.Size()
	section.Header = header
	section.sizeBytes += header.Size()
}

// AddBodyPair adds a body pair to a section and updates the section size
func (section *ChainSection) AddBodyPair(pair *BodyPair) {
	section.Body = append(section.Body, pair)
	section.sizeBytes += pair.Size()
}

// AddSection adds a section to the ChainAST
func (ast *ChainAST) AddSection(section *ChainSection) {
	ast.Sections = append(ast.Sections, section)
}

// HasToolCalls checks if an AI message contains tool calls
func HasToolCalls(msg *llms.MessageContent) bool {
	if msg == nil {
		return false
	}

	for _, part := range msg.Parts {
		if _, isToolCall := part.(llms.ToolCall); isToolCall {
			return true
		}
	}

	return false
}

// String returns a string representation of the ChainAST for debugging
func (ast *ChainAST) String() string {
	var b strings.Builder
	b.WriteString("ChainAST {\n")

	for i, section := range ast.Sections {
		b.WriteString(fmt.Sprintf("  Section %d {\n", i))
		b.WriteString("    Header {\n")
		if section.Header.SystemMessage != nil {
			b.WriteString("      SystemMessage\n")
		}
		if section.Header.HumanMessage != nil {
			b.WriteString("      HumanMessage\n")
		}
		b.WriteString("    }\n")

		b.WriteString("    Body {\n")
		for j, bodyPair := range section.Body {
			switch bodyPair.Type {
			case RequestResponse:
				b.WriteString(fmt.Sprintf("      BodyPair %d (RequestResponse) {\n", j))
			case Completion:
				b.WriteString(fmt.Sprintf("      BodyPair %d (Completion) {\n", j))
			case Summarization:
				b.WriteString(fmt.Sprintf("      BodyPair %d (Summarization) {\n", j))
			}
			b.WriteString("        AIMessage\n")
			b.WriteString(fmt.Sprintf("        ToolMessages: %d\n", len(bodyPair.ToolMessages)))
			b.WriteString("      }\n")
		}
		b.WriteString("    }\n")
		b.WriteString("  }\n")
	}

	b.WriteString("}\n")
	return b.String()
}

// FindToolCallResponses finds all tool call responses for a given tool call ID
func (ast *ChainAST) FindToolCallResponses(toolCallID string) []llms.ToolCallResponse {
	var responses []llms.ToolCallResponse

	for _, section := range ast.Sections {
		for _, bodyPair := range section.Body {
			if bodyPair.Type != RequestResponse {
				continue
			}
			for _, toolMsg := range bodyPair.ToolMessages {
				for _, part := range toolMsg.Parts {
					resp, ok := part.(llms.ToolCallResponse)
					if ok && resp.ToolCallID == toolCallID {
						responses = append(responses, resp)
					}
				}
			}
		}
	}

	return responses
}

// CalculateMessageSize calculates the size of a message in bytes
func CalculateMessageSize(msg *llms.MessageContent) int {
	size := 0
	for _, part := range msg.Parts {
		switch p := part.(type) {
		case llms.TextContent:
			size += len(p.Text)
		case llms.ImageURLContent:
			size += len(p.URL)
		case llms.BinaryContent:
			size += len(p.Data)
		case llms.ToolCall:
			size += len(p.ID) + len(p.Type)
			if p.FunctionCall != nil {
				size += len(p.FunctionCall.Name) + len(p.FunctionCall.Arguments)
			}
		case llms.ToolCallResponse:
			size += len(p.ToolCallID) + len(p.Name) + len(p.Content)
		}
	}
	return size
}

// CalculateBodyPairSize calculates the size of a body pair in bytes
func CalculateBodyPairSize(pair *BodyPair) int {
	size := 0
	if pair.AIMessage != nil {
		size += CalculateMessageSize(pair.AIMessage)
	}

	for _, toolMsg := range pair.ToolMessages {
		size += CalculateMessageSize(toolMsg)
	}
	return size
}

// AppendHumanMessage adds a human message to the chain following these rules:
// 1. If chain is empty, creates a new section with this message as HumanMessage
// 2. If the last section has body pairs (AI responses), creates a new section with this message
// 3. If the last section has no body pairs and no HumanMessage, adds this message to that section
// 4. If the last section has no body pairs but has HumanMessage, appends content to existing message
func (ast *ChainAST) AppendHumanMessage(content string) {
	newTextPart := llms.TextContent{Text: content}

	// Case 1: Chain is empty - create a new section
	if len(ast.Sections) == 0 {
		humanMsg := &llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{newTextPart},
		}

		// Create new header and section with calculated sizes
		header := NewHeader(nil, humanMsg)
		section := NewChainSection(header, []*BodyPair{})

		ast.Sections = append(ast.Sections, section)
		return
	}

	// Get the last section
	lastSection := ast.Sections[len(ast.Sections)-1]

	// Case 2: Last section has body pairs - create a new section
	if len(lastSection.Body) > 0 {
		humanMsg := &llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{newTextPart},
		}

		// Create new header and section with calculated sizes
		header := NewHeader(nil, humanMsg)
		section := NewChainSection(header, []*BodyPair{})

		ast.Sections = append(ast.Sections, section)
		return
	}

	// Case 3: Last section has no HumanMessage - add to this section
	// This includes the case where there's only a SystemMessage
	if lastSection.Header.HumanMessage == nil {
		lastSection.SetHeader(NewHeader(lastSection.Header.SystemMessage, &llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{newTextPart},
		}))
		return
	}

	// Case 4: Last section has HumanMessage - append to existing message
	lastSection.Header.HumanMessage.Parts = append(lastSection.Header.HumanMessage.Parts, newTextPart)
	lastSection.SetHeader(NewHeader(lastSection.Header.SystemMessage, lastSection.Header.HumanMessage))
}

// AddToolResponse adds a response to a tool call
// If the tool call is not found, it returns an error
// If the tool call already has a response, it updates the response
func (ast *ChainAST) AddToolResponse(toolCallID, toolName, content string) error {
	for _, section := range ast.Sections {
		for _, bodyPair := range section.Body {
			if bodyPair.Type == RequestResponse {
				// First check if this body pair contains the tool call we're looking for
				toolCallFound := false
				for _, part := range bodyPair.AIMessage.Parts {
					if toolCall, ok := part.(llms.ToolCall); ok &&
						toolCall.FunctionCall != nil &&
						toolCall.ID == toolCallID {
						toolCallFound = true
						break
					}
				}

				if !toolCallFound {
					continue // This body pair doesn't contain our tool call
				}

				// Check if there's already a response for this tool call
				responseUpdated := false
				for _, toolMsg := range bodyPair.ToolMessages {
					oldToolMsgSize := CalculateMessageSize(toolMsg)

					for i, part := range toolMsg.Parts {
						if resp, ok := part.(llms.ToolCallResponse); ok && resp.ToolCallID == toolCallID {
							// Update existing response
							resp.Content = content
							toolMsg.Parts[i] = resp
							responseUpdated = true

							// Recalculate tool message size and update size differences
							newToolMsgSize := CalculateMessageSize(toolMsg)
							sizeDiff := newToolMsgSize - oldToolMsgSize
							bodyPair.sizeBytes += sizeDiff
							section.sizeBytes += sizeDiff

							return nil
						}
					}
				}

				// If no existing response was found, add a new one
				if !responseUpdated {
					resp := llms.ToolCallResponse{
						ToolCallID: toolCallID,
						Name:       toolName,
						Content:    content,
					}

					// Add response to existing tool message or create a new one
					if len(bodyPair.ToolMessages) > 0 {
						oldToolMsgSize := CalculateMessageSize(bodyPair.ToolMessages[len(bodyPair.ToolMessages)-1])

						lastToolMsg := bodyPair.ToolMessages[len(bodyPair.ToolMessages)-1]
						lastToolMsg.Parts = append(lastToolMsg.Parts, resp)

						// Recalculate tool message size and update size differences
						newToolMsgSize := CalculateMessageSize(lastToolMsg)
						sizeDiff := newToolMsgSize - oldToolMsgSize
						bodyPair.sizeBytes += sizeDiff
						section.sizeBytes += sizeDiff
					} else {
						toolMsg := &llms.MessageContent{
							Role:  llms.ChatMessageTypeTool,
							Parts: []llms.ContentPart{resp},
						}
						bodyPair.ToolMessages = append(bodyPair.ToolMessages, toolMsg)

						// Calculate new tool message size and add to totals
						toolMsgSize := CalculateMessageSize(toolMsg)
						bodyPair.sizeBytes += toolMsgSize
						section.sizeBytes += toolMsgSize
					}
					return nil
				}
			}
		}
	}

	return fmt.Errorf("tool call with ID %s not found", toolCallID)
}

// Size returns the size of a section in bytes
func (section *ChainSection) Size() int {
	return section.sizeBytes
}

// Size returns the size of a body pair in bytes
func (pair *BodyPair) Size() int {
	return pair.sizeBytes
}

// Size returns the total size of the ChainAST in bytes
func (ast *ChainAST) Size() int {
	totalSize := 0
	for _, section := range ast.Sections {
		totalSize += section.sizeBytes
	}
	return totalSize
}

// NormalizeToolCallIDs validates and replaces tool call IDs that don't match the new template.
// This is useful when switching between different LLM providers that use different ID formats.
// For example, switching from Gemini to Anthropic requires converting IDs from one format to another.
//
// The function:
// 1. Validates each tool call ID against the new template using ValidatePattern
// 2. If validation fails, generates a new ID and creates a mapping
// 3. Updates all tool call responses to use the new IDs
// 4. Preserves IDs that already match the template
func (ast *ChainAST) NormalizeToolCallIDs(newTemplate string) error {
	// Mapping from old tool call IDs to new ones
	idMapping := make(map[string]string)

	for _, section := range ast.Sections {
		for _, bodyPair := range section.Body {
			// Only process RequestResponse and Summarization types
			if bodyPair.Type != RequestResponse && bodyPair.Type != Summarization {
				continue
			}

			if bodyPair.AIMessage == nil {
				continue
			}

			// Check and replace tool call IDs in AI message
			for pdx, part := range bodyPair.AIMessage.Parts {
				toolCall, ok := part.(llms.ToolCall)
				if !ok || toolCall.FunctionCall == nil {
					continue
				}

				// Validate existing ID against new template
				err := templates.ValidatePattern(newTemplate, toolCall.ID)
				if err != nil {
					// ID doesn't match the new pattern - generate a new one
					newID := templates.GenerateFromPattern(newTemplate)
					idMapping[toolCall.ID] = newID
					toolCall.ID = newID
					bodyPair.AIMessage.Parts[pdx] = toolCall
				}
				// If err == nil, ID is already valid for the new template
			}

			// Update corresponding tool call responses with new IDs
			for _, toolMsg := range bodyPair.ToolMessages {
				if toolMsg == nil {
					continue
				}

				for pdx, part := range toolMsg.Parts {
					resp, ok := part.(llms.ToolCallResponse)
					if !ok {
						continue
					}

					// Check if this response ID needs to be updated
					if newID, exists := idMapping[resp.ToolCallID]; exists {
						resp.ToolCallID = newID
						toolMsg.Parts[pdx] = resp
					}
				}
			}
		}
	}

	return nil
}

// ClearReasoning removes all reasoning data from the chain.
// This is necessary when switching between providers because reasoning content
// (especially cryptographic signatures) is provider-specific and will cause
// API errors if sent to a different provider.
//
// The function clears reasoning from:
// - TextContent parts (may contain extended thinking signatures)
// - ToolCall parts (may contain per-tool reasoning)
func (ast *ChainAST) ClearReasoning() error {
	for _, section := range ast.Sections {
		if section.Header == nil {
			continue
		}

		// Clear reasoning from header messages
		if section.Header.SystemMessage != nil {
			clearMessageReasoning(section.Header.SystemMessage)
		}
		if section.Header.HumanMessage != nil {
			clearMessageReasoning(section.Header.HumanMessage)
		}

		// Clear reasoning from body pairs
		for _, bodyPair := range section.Body {
			if bodyPair.AIMessage != nil {
				clearMessageReasoning(bodyPair.AIMessage)
			}

			// Tool messages don't typically have reasoning, but clear them anyway
			for _, toolMsg := range bodyPair.ToolMessages {
				if toolMsg != nil {
					clearMessageReasoning(toolMsg)
				}
			}
		}
	}

	return nil
}

// clearMessageReasoning clears reasoning from all parts of a message
func clearMessageReasoning(msg *llms.MessageContent) {
	if msg == nil {
		return
	}

	for idx, part := range msg.Parts {
		switch p := part.(type) {
		case llms.TextContent:
			if p.Reasoning != nil {
				p.Reasoning = nil
				msg.Parts[idx] = p
			}
		case llms.ToolCall:
			if p.Reasoning != nil {
				p.Reasoning = nil
				msg.Parts[idx] = p
			}
		}
	}
}
