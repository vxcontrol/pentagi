# ChainAST Documentation

## Table of Contents

- [ChainAST Documentation](#chainast-documentation)
  - [Table of Contents](#table-of-contents)
  - [Introduction](#introduction)
  - [Structure Overview](#structure-overview)
  - [Constants and Default Values](#constants-and-default-values)
  - [Size Tracking Features](#size-tracking-features)
  - [Creating and Using ChainAST](#creating-and-using-chainast)
    - [Basic Creation](#basic-creation)
    - [Using Constructors](#using-constructors)
    - [Getting Messages](#getting-messages)
    - [Body Pair Validation](#body-pair-validation)
    - [Common Validation Rules](#common-validation-rules)
  - [Modifying Message Chains](#modifying-message-chains)
    - [Adding Elements](#adding-elements)
    - [Adding Human Messages](#adding-human-messages)
    - [Working with Tool Calls](#working-with-tool-calls)
  - [Testing Utilities](#testing-utilities)
    - [Predefined Test Chains](#predefined-test-chains)
    - [Generating Custom Test Chains](#generating-custom-test-chains)
  - [Message Chain Structure in LLM Providers](#message-chain-structure-in-llm-providers)
    - [Message Roles](#message-roles)
    - [Message Content](#message-content)
  - [Provider-Specific Requirements](#provider-specific-requirements)
    - [Reasoning Signatures](#reasoning-signatures)
      - [Gemini (Google AI)](#gemini-google-ai)
      - [Anthropic (Claude)](#anthropic-claude)
    - [Helper Functions](#helper-functions)
  - [Best Practices](#best-practices)
  - [Common Use Cases](#common-use-cases)
    - [1. Chain Validation and Repair](#1-chain-validation-and-repair)
    - [2. Chain Summarization](#2-chain-summarization)
    - [3. Adding Tool Responses](#3-adding-tool-responses)
    - [4. Building a Conversation](#4-building-a-conversation)
    - [5. Using Summarization in Conversation](#5-using-summarization-in-conversation)
  - [Example Usage](#example-usage)

## Introduction

ChainAST is a structured representation of message chains used in Large Language Model (LLM) conversations. It organizes conversations into a logical hierarchy, making it easier to analyze, modify, and validate message sequences, especially when they involve tool calls and their responses.

The structure helps address common issues in LLM conversations such as:
- Validating proper conversation flow
- Managing tool calls and their responses
- Handling conversation sections and state changes
- Ensuring consistent conversation structure
- Efficient size tracking for summarization and context management

## Structure Overview

ChainAST represents a message chain as an abstract syntax tree with the following components:

```
ChainAST
├── Sections[] (ChainSection)
    ├── Header
    │   ├── SystemMessage (optional)
    │   ├── HumanMessage (optional)
    │   └── sizeBytes (total header size in bytes)
    ├── sizeBytes (total section size in bytes)
    └── Body[] (BodyPair)
        ├── Type (RequestResponse, Completion, or Summarization)
        ├── AIMessage
        ├── ToolMessages[] (for RequestResponse and Summarization types)
        └── sizeBytes (total body pair size in bytes)
```

Components:
- **ChainAST**: The root structure containing an array of sections
- **ChainSection**: A logical unit of conversation, starting with a header and containing multiple body pairs
  - Includes `sizeBytes` tracking total section size in bytes
- **Header**: Contains system and/or human messages that initiate a section
  - Includes `sizeBytes` tracking total header size in bytes
- **BodyPair**: Represents an AI response, which may include tool calls and their responses
  - Includes `sizeBytes` tracking total body pair size in bytes
- **RequestResponse**: A type of body pair where the AI message contains tool calls requiring responses
- **Completion**: A simple AI message without tool calls
- **Summarization**: A special type of body pair containing a tool call to the summarization tool

## Constants and Default Values

ChainAST provides several important constants:
- `fallbackRequestArgs`: Default arguments (`{}`) for tool calls without specified arguments
- `FallbackResponseContent`: Default response content ("the call was not handled, please try again") for missing tool responses when using force=true
- `SummarizationToolName`: Name of the special summarization tool ("execute_task_and_return_summary")
- `SummarizationToolArgs`: Default arguments for the summarization tool (`{"question": "delegate and execute the task, then return the summary of the result"}`)

## Size Tracking Features

ChainAST includes built-in size tracking to support efficient summarization algorithms and context management:

```go
// Get the size of a section in bytes
sizeInBytes := section.Size()

// Get the size of a body pair in bytes
sizeInBytes := bodyPair.Size()

// Get the size of a header in bytes
sizeInBytes := header.Size()

// Get the total size of the entire ChainAST
totalSize := ast.Size()
```

Size calculation considers all content types including:
- Text content (string length)
- Image URLs (URL string length)
- Binary data (byte count)
- Tool calls (ID, type, name, and arguments length)
- Tool call responses (ID, name, and content length)

The `sizeBytes` values are automatically maintained when:
- Creating a new ChainAST from a message chain
- Appending human messages
- Adding tool responses
- Creating elements with constructors

## Creating and Using ChainAST

### Basic Creation

```go
// Create from an existing message chain
ast, err := NewChainAST(messageChain, false)
if err != nil {
    // Handle validation error
}

// Get messages (flattened chain)
flatChain := ast.Messages()
```

The `force` parameter in `NewChainAST` determines how the function handles inconsistencies:
- `force=false`: Strict validation, returns errors for any inconsistency
- `force=true`: Attempts to repair problems by:
  - Merging consecutive human messages into a single message with multiple content parts
  - Adding missing tool responses with placeholder content ("the call was not handled, please try again")
  - Skipping invalid messages like unexpected tool messages without preceding AI messages

During creation, the size of all components is calculated automatically.

### Using Constructors

ChainAST provides constructors to create elements with automatic size calculation:

```go
// Create a header
header := NewHeader(systemMsg, humanMsg)

// Create a body pair (automatically determines type based on content)
bodyPair := NewBodyPair(aiMsg, toolMsgs)

// Create a body pair from a slice of messages
bodyPair, err := NewBodyPairFromMessages(messages)

// Create a chain section
section := NewChainSection(header, bodyPairs)

// Create a completion body pair with text
completionPair := NewBodyPairFromCompletion("This is a response")

// Create a summarization body pair with text
// The third parameter (addFakeSignature) should be true if the original content
// contained reasoning signatures (required for providers like Gemini)
summarizationPair := NewBodyPairFromSummarization("This is a summary of the conversation", tcIDTemplate, false)

// Create a summarization body pair with fake reasoning signature
// This is necessary when summarizing content that originally had reasoning signatures
// to satisfy provider requirements (e.g., Gemini's thought_signature)
summarizationWithSignature := NewBodyPairFromSummarization("Summary with signature", tcIDTemplate, true)

// Check if messages contain reasoning signatures in ToolCall parts
// This is useful for determining if summarized content should include fake signatures
// Only checks ToolCall.Reasoning (not TextContent.Reasoning)
hasReasoning := ContainsToolCallReasoning(messages)

// Check if a message contains tool calls
hasCalls := HasToolCalls(aiMessage)
```

### Getting Messages

Each component provides a method to get its messages in the correct order:

```go
// Get all messages from a header (system first, then human)
headerMsgs := header.Messages()

// Get all messages from a body pair (AI first, then tools)
bodyPairMsgs := bodyPair.Messages()

// Get all messages from a section
sectionMsgs := section.Messages()

// Get all messages from the ChainAST
allMsgs := ast.Messages()
```

### Body Pair Validation

The `IsValid()` method checks if a BodyPair follows the structure rules:

```go
// Check if a body pair is valid
isValid := bodyPair.IsValid()
```

Validation rules depend on the body pair type:
- For **Completion**: No tool messages allowed
- For **RequestResponse**: Must have at least one tool message
- For **Summarization**: Must have exactly one tool message
- For all types: All tool calls must have matching responses and vice versa

The `GetToolCallsInfo()` method returns detailed information about tool calls:

```go
// Get information about tool calls and responses
toolCallsInfo := bodyPair.GetToolCallsInfo()

// Check for pending or unmatched tool calls
if len(toolCallsInfo.PendingToolCallIDs) > 0 {
    // There are tool calls without responses
}

if len(toolCallsInfo.UnmatchedToolCallIDs) > 0 {
    // There are tool responses without matching tool calls
}

// Access completed tool calls
for id, pair := range toolCallsInfo.CompletedToolCalls {
    // Use tool call and response information
    toolCall := pair.ToolCall
    response := pair.Response
}
```

### Common Validation Rules

When `force=false`, NewChainAST enforces these rules:
1. First message must be System or Human
2. No consecutive Human messages
3. Tool calls must have matching responses
4. Tool responses must reference valid tool calls
5. System messages can't appear in the middle of a chain
6. AI messages with tool calls must have responses before another AI message
7. Summarization body pairs must have exactly one tool message

## Modifying Message Chains

### Adding Elements

```go
// Add a section to the ChainAST
ast.AddSection(section)

// Add a body pair to a section
section.AddBodyPair(bodyPair)
```

### Adding Human Messages

```go
// Append a new human message
ast.AppendHumanMessage("Tell me more about this topic")
```

The function follows these rules:
1. If chain is empty: Creates a new section with this message as HumanMessage
2. If the last section has body pairs (AI responses): Creates a new section with this message
3. If the last section has no body pairs and no HumanMessage: Adds this message to that section
4. If the last section has no body pairs but has HumanMessage: Appends content to the existing message

Section and header sizes are automatically updated when human messages are added or modified.

### Working with Tool Calls

```go
// Add a response to a tool call
err := ast.AddToolResponse("tool-call-id", "tool-name", "Response content")
if err != nil {
    // Handle error (tool call not found)
}

// Find all responses for a specific tool call
responses := ast.FindToolCallResponses("tool-call-id")
```

The `AddToolResponse` function:
- Searches for the specified tool call ID in AI messages
- If the tool call is found and already has a response, updates the existing response content
- If the tool call is found but doesn't have a response, adds a new response
- If the tool call is not found, returns an error

Body pair and section sizes are automatically updated when tool responses are added or modified.

## Testing Utilities

ChainAST comes with utilities for generating test message chains to validate your code.

### Predefined Test Chains

Several test chains are available in the package for common scenarios:

```go
// Basic chains
emptyChain              // Empty message chain
systemOnlyChain         // Only a system message
humanOnlyChain          // Only a human message
systemHumanChain        // System + human messages
basicConversationChain  // System + human + AI response

// Tool-related chains
chainWithTool                  // Chain with a tool call, no response
chainWithSingleToolResponse    // Chain with a tool call and response
chainWithMultipleTools         // Chain with multiple tool calls
chainWithMultipleToolResponses // Chain with multiple tool calls and responses

// Complex chains
chainWithMultipleSections      // Multiple conversation turns
chainWithConsecutiveHumans     // Chain with error: consecutive human messages
chainWithMissingToolResponse   // Chain with error: missing tool response
chainWithUnexpectedTool        // Chain with error: unexpected tool message

// Summarization chains
chainWithSummarization         // Chain with summarization as the only body pair
chainWithSummarizationAndOtherPairs // Chain with summarization followed by other body pairs
```

### Generating Custom Test Chains

For more complex testing, use the chain generators:

```go
// Simple configuration
config := DefaultChainConfig()  // Creates a simple chain with system + human + AI

// Custom configuration
config := ChainConfig{
    IncludeSystem:           true,
    Sections:                3,                 // 3 conversation turns
    BodyPairsPerSection:     []int{1, 2, 1},    // Number of AI responses per section
    ToolsForBodyPairs:       []bool{false, true, false}, // Which responses have tool calls
    ToolCallsPerBodyPair:    []int{0, 2, 0},    // How many tool calls per response
    IncludeAllToolResponses: true,              // Whether to include responses for all tools
}

// Generate chain based on config
chain := GenerateChain(config)

// For more complex scenarios with missing responses
complexChain := GenerateComplexChain(
    5,  // Number of sections
    3,  // Number of tool calls per tool-using response
    7   // Number of missing tool responses
)
```

The `ChainConfig` struct allows fine-grained control over generated test chains:
- `IncludeSystem`: Whether to add a system message at the start
- `Sections`: Number of conversation turns (each with a human message)
- `BodyPairsPerSection`: Number of AI responses per section
- `ToolsForBodyPairs`: Which AI responses should include tool calls
- `ToolCallsPerBodyPair`: Number of tool calls to include in each tool-using response
- `IncludeAllToolResponses`: Whether to add responses for all tool calls

## Message Chain Structure in LLM Providers

ChainAST is designed to work with message chains that follow common conventions in LLM providers:

### Message Roles

- **System**: Provides context or instructions to the model
- **Human/User**: User input messages
- **AI/Assistant**: Model responses
- **Tool**: Results of tool calls executed by the system

### Message Content

Messages can contain different types of content:
- **TextContent**: Simple text messages
- **ToolCall**: Function call requests from the model
- **ToolCallResponse**: Results returned from executing tools

## Provider-Specific Requirements

### Reasoning Signatures

Different LLM providers have specific requirements for reasoning content in function calls:

#### Gemini (Google AI)

Gemini requires **thought signatures** (`thought_signature`) for function calls, especially in multi-turn conversations with tool use. These signatures:

- Are cryptographic representations of the model's internal reasoning process
- Are strictly validated only for the **current turn** (defined as all messages after the last user message with text content)
- Must be preserved when summarizing content that contains them
- Can use fake signatures when creating summarized content: `"skip_thought_signature_validator"`

**Example:**
```go
// Check if original content had reasoning
hasReasoning := ContainsReasoning(originalMessages)

// Create summarized content with fake signature if needed
summaryPair := NewBodyPairFromSummarization(summaryText, tcIDTemplate, hasReasoning)
```

#### Anthropic (Claude)

Anthropic uses **extended thinking** with cryptographic signatures that:

- Are automatically removed from previous turns (not counted in context window)
- Are only required for the current tool use loop

**Critical Rule:** Never summarize the last body pair in a section, as this preserves reasoning signatures required by both Gemini and Anthropic.

### Helper Functions

```go
// Check if messages contain reasoning signatures in ToolCall parts
// Returns true if any message contains Reasoning in ToolCall (NOT TextContent)
// This is specific to function calling scenarios which require thought_signature
hasReasoning := ContainsToolCallReasoning(messages)

// Create summarization with conditional fake signature
addFakeSignature := ContainsToolCallReasoning(originalMessages)
summaryPair := NewBodyPairFromSummarization(summaryText, tcIDTemplate, addFakeSignature)
```

## Best Practices

1. **Validation First**: Use `NewChainAST` with `force=false` to validate chains before processing
2. **Defensive Programming**: Always check for errors from ChainAST functions
3. **Complete Tool Calls**: Ensure all tool calls have corresponding responses before sending to an LLM
4. **Section Management**: Use sections to organize conversation turns logically
5. **Testing**: Use the provided generators to test code that manipulates message chains
6. **Size Management**: Leverage size tracking to maintain efficient context windows
7. **Reasoning Preservation**: Use `ContainsToolCallReasoning()` to check if fake signatures are needed when summarizing (only checks ToolCall.Reasoning, not TextContent.Reasoning)
8. **Last Pair Protection**: Never summarize the last (most recent) body pair in a section to preserve reasoning signatures

## Common Use Cases

### 1. Chain Validation and Repair

```go
// Try to parse with strict validation
ast, err := NewChainAST(chain, false)
if err != nil {
    // If validation fails, try with repair enabled
    ast, err = NewChainAST(chain, true)
    if err != nil {
        // Handle severe structural errors
    }
    // Log that the chain was repaired
}
```

### 2. Chain Summarization

```go
// Create AST from chain
ast, _ := NewChainAST(chain, true)

// Analyze total size and section sizes
totalSize := ast.Size()
if totalSize > maxContextSize {
    // Select sections to summarize
    oldestSections := ast.Sections[:len(ast.Sections)-1] // Keep last section

    // Summarize sections
    summaryText := generateSummary(oldestSections)

    // Create a new AST with the summary
    newAST := &ChainAST{Sections: []*ChainSection{}}

    // Copy system message if exists
    var systemMsg *llms.MessageContent
    if len(ast.Sections) > 0 && ast.Sections[0].Header.SystemMessage != nil {
        systemMsgCopy := *ast.Sections[0].Header.SystemMessage
        systemMsg = &systemMsgCopy
    }

    // Create header and section
    header := NewHeader(systemMsg, nil)
    section := NewChainSection(header, []*BodyPair{})
    newAST.AddSection(section)

    // Add summarization body pair
    summaryPair := NewBodyPairFromSummarization(summaryText)
    section.AddBodyPair(summaryPair)

    // Copy the most recent section
    lastSection := ast.Sections[len(ast.Sections)-1]
    // Add appropriate logic to copy the last section

    // Get the summarized chain
    summarizedChain := newAST.Messages()
}
```

### 3. Adding Tool Responses

```go
// Parse a chain with tool calls
ast, _ := NewChainAST(chain, false)

// Find unresponded tool calls and add responses
for _, section := range ast.Sections {
    for _, bodyPair := range section.Body {
        if bodyPair.Type == RequestResponse {
            for _, part := range bodyPair.AIMessage.Parts {
                if toolCall, ok := part.(llms.ToolCall); ok {
                    // Execute the tool
                    result := executeToolCall(toolCall)

                    // Add the response
                    ast.AddToolResponse(toolCall.ID, toolCall.FunctionCall.Name, result)
                }
            }
        }
    }
}

// Get the updated chain
updatedChain := ast.Messages()
```

### 4. Building a Conversation

```go
// Create an empty AST
ast := &ChainAST{Sections: []*ChainSection{}}

// Add system message
sysMsg := &llms.MessageContent{
    Role: llms.ChatMessageTypeSystem,
    Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant"}},
}
header := NewHeader(sysMsg, nil)
section := NewChainSection(header, []*BodyPair{})
ast.AddSection(section)

// Add a human message
ast.AppendHumanMessage("Hello, how can you help me?")

// Add an AI response
aiMsg := &llms.MessageContent{
    Role: llms.ChatMessageTypeAI,
    Parts: []llms.ContentPart{llms.TextContent{Text: "I can answer questions, help with tasks, and more."}},
}
bodyPair := NewBodyPair(aiMsg, nil)
section.AddBodyPair(bodyPair)

// Continue the conversation
ast.AppendHumanMessage("Can you help me find information?")

// Get the message chain
chain := ast.Messages()
```

### 5. Using Summarization in Conversation

```go
// Create an empty AST
ast := &ChainAST{Sections: []*ChainSection{}}

// Create a new header with a system message
sysMsg := &llms.MessageContent{
    Role:  llms.ChatMessageTypeSystem,
    Parts: []llms.ContentPart{llms.TextContent{Text: "You are a helpful assistant."}},
}
header := NewHeader(sysMsg, nil)

// Create a new section with the header
section := NewChainSection(header, []*BodyPair{})
ast.AddSection(section)

// Add a human message requesting a summary
ast.AppendHumanMessage("Can you summarize our discussion?")

// Create a summarization body pair
summaryPair := NewBodyPairFromSummarization("This is a summary of our previous conversation about weather and travel plans.")
section.AddBodyPair(summaryPair)

// Get the message chain
chain := ast.Messages()
```

## Example Usage

```go
// Parse a conversation chain with summarization
ast, err := NewChainAST(conversationChain, true)
if err != nil {
    log.Fatalf("Failed to parse chain: %v", err)
}

// Check if any body pairs are summarization pairs
for _, section := range ast.Sections {
    for _, bodyPair := range section.Body {
        if bodyPair.Type == Summarization {
            fmt.Println("Found a summarization body pair")

            // Extract the summary text from the tool response
            for _, toolMsg := range bodyPair.ToolMessages {
                for _, part := range toolMsg.Parts {
                    if resp, ok := part.(llms.ToolCallResponse); ok &&
                       resp.Name == SummarizationToolName {
                        fmt.Printf("Summary content: %s\n", resp.Content)
                    }
                }
            }
        }
    }
}
```
