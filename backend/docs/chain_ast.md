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
    - [Message Builders](#message-builders)
    - [Predefined Test Chains](#predefined-test-chains)
    - [Generating Custom Test Chains](#generating-custom-test-chains)
  - [Message Chain Structure in LLM Providers](#message-chain-structure-in-llm-providers)
    - [Message Roles](#message-roles)
    - [Message Content](#message-content)
  - [Provider-Specific Requirements](#provider-specific-requirements)
    - [Reasoning Signatures](#reasoning-signatures)
      - [Gemini (Google AI)](#gemini-google-ai)
      - [Anthropic (Claude)](#anthropic-claude)
      - [Kimi/Moonshot (OpenAI-compatible)](#kimimoonshot-openai-compatible)
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
// contained ToolCall reasoning signatures (required for providers like Gemini)
// The fourth parameter (reasoningMsg) preserves reasoning TextContent before ToolCall
// (required for providers like Kimi/Moonshot)
summarizationPair := NewBodyPairFromSummarization("This is a summary of the conversation", tcIDTemplate, false, nil)

// Create a summarization body pair with fake reasoning signature (Gemini)
// This is necessary when summarizing content that originally had ToolCall reasoning
// to satisfy provider requirements (e.g., Gemini's thought_signature)
summarizationWithSignature := NewBodyPairFromSummarization("Summary with signature", tcIDTemplate, true, nil)

// Extract reasoning message for Kimi/Moonshot compatibility
// Returns the first AI message with TextContent containing reasoning (or nil)
reasoningMsg := ExtractReasoningMessage(messages)

// Create summarization with preserved reasoning message (Kimi/Moonshot)
summarizationWithReasoning := NewBodyPairFromSummarization("Summary", tcIDTemplate, false, reasoningMsg)

// Create summarization with BOTH fake signature AND reasoning message
// Required when original had both ToolCall.Reasoning and TextContent.Reasoning
summarizationFull := NewBodyPairFromSummarization("Summary", tcIDTemplate, true, reasoningMsg)

// Check if messages contain reasoning signatures in ToolCall parts
// This is useful for determining if summarized content should include fake signatures
// Only checks ToolCall.Reasoning (not TextContent.Reasoning)
hasToolCallReasoning := ContainsToolCallReasoning(messages)

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

The package tests build their message chains from `pkg/cast/fixtures_test.go`. Every helper there is a function that returns a fresh chain on each call: parsing shares the `Parts` of the input messages with the AST, and `AddToolResponse`, `NormalizeToolCallIDs`, `ClearReasoning` and `SanitizeToolCallArguments` write into those parts, so a chain kept in a shared variable would change under the next test. The file is a `_test.go` file of package `cast`, so only the tests of this package can use it.

### Message Builders

```go
chainASTSystem(text)             // a system message with one text part
chainASTHuman(text)              // a human message with one text part
chainASTAnswer(text)             // an AI message with one text part: a completion
chainASTText(role, text)         // a message of any role with one text part
chainASTAI(parts...)             // an AI message with the given parts
chainASTCall(id, name)           // a "function" tool call with the arguments {}
chainASTCallWith(id, name, args) // the same call with the given arguments
chainASTTool(id, name, content)  // a tool message holding one response
```

### Predefined Test Chains

```go
// Basic chains
emptyChain()             // no messages
systemOnlyChain()        // only a system message
humanOnlyChain()         // only a human message
systemHumanChain()       // system + human messages
basicConversationChain() // system + human + AI completion

// Tool-related chains
chainWithTool()               // a call to get_weather (tool-1) with no response
chainWithSingleToolResponse() // the same call and its response
chainWithMultipleTools()      // calls to get_weather (tool-1) and get_time (tool-2), neither answered

// Complex chains
chainWithMultipleSections()    // a completion, then a second section with a call and its response
chainWithMissingToolResponse() // the two calls of chainWithMultipleTools with only tool-1 answered
chainWithUnexpectedTool()      // a completion followed by a tool message that answers no call

// Summarization chains
chainWithSummarizationAndOtherPairs() // a summarization pair, a completion, then a call and its response

// Provider switching
providerSwitchChain(withReasoning) // a scan in three sections, described below
```

`providerSwitchChain` is the chain a provider switch is tested against: `NormalizeToolCallIDs` and `ClearReasoning` run on it each on its own and, clearing first, one after the other; `NormalizeToolCallIDs`, `SanitizeToolCallArguments` and `ClearReasoning` run in the order `pkg/providers` runs them on a variant of it that needs a forced repair. Its first section holds a text preamble beside two parallel calls, one with an id already in the Anthropic format (`toolu_…`) and one with an OpenAI-style id, their two responses and a completion; the second section opens with a completion before a call whose tool message carries a text note beside the response; the third holds a summarization pair. With `withReasoning` set, the preamble, the first completion, the note and every call except the one already in the Anthropic format carry reasoning with a signature. Without it the chain is otherwise identical, which is what `ClearReasoning` must turn the first form into; both forms have the same sizes, because reasoning is not counted.

### Generating Custom Test Chains

For chains with more sections, pairs and calls than the predefined ones, use the generators:

```go
// A system prompt, one question and its completion
config := DefaultChainConfig()

// Custom configuration
config := ChainConfig{
    IncludeSystem:           true,
    Sections:                3,                                      // 3 conversation turns
    BodyPairsPerSection:     []int{1, 2, 2},                         // AI responses per section
    ToolsForBodyPairs:       []bool{false, true, false, true, true}, // which responses, counted across the chain, make calls
    ToolCallsPerBodyPair:    []int{0, 2, 0, 3, 1},                   // how many calls each of those responses makes
    IncludeAllToolResponses: true,                                   // whether every call gets its response
}

// Generate chain based on config
chain := GenerateChain(config)

// Five sections, a response with three calls in the first, third and fifth and a completion in the
// others, with the first seven tool responses of the chain dropped
complexChain := GenerateComplexChain(
    5, // number of sections
    3, // number of calls per tool-using response
    7, // number of tool responses to drop, counted from the start of the chain
)
```

The `ChainConfig` struct controls the generated chain:
- `IncludeSystem`: whether the chain opens with the system message "You are a helpful assistant."
- `Sections`: the number of conversation turns; turn n opens with the human message "Question n"
- `BodyPairsPerSection`: the number of AI responses in each section; a section past the end of the slice has one
- `ToolsForBodyPairs`: which AI responses make tool calls, indexed by the position of the response in the whole chain; a response past the end of the slice is the completion "Response to question n"
- `ToolCallsPerBodyPair`: how many calls each tool-using response makes, indexed the same way; a response past the end of the slice makes one
- `IncludeAllToolResponses`: whether every call is followed by its response; without it no call is answered

Calls are numbered `tool-1`, `tool-2`, … across the whole chain. The k-th call of a response is `get_data_k` with the arguments `{"query": "Test query k"}`, and its response, one tool message per call in the order of the calls, reads "Response for tool-N".

`GenerateComplexChain(sections, calls, missing)` gives every section one AI response, a response with `calls` calls in the first, third, fifth… section and a completion in the others, answers every call, and then drops the first `missing` tool messages of the chain.

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

#### Kimi/Moonshot (OpenAI-compatible)

Kimi reasoning models require **reasoning_content in TextContent** before ToolCall:

- Reasoning must be present in a TextContent part before any ToolCall when thinking is enabled
- Error: "thinking is enabled but reasoning_content is missing in assistant tool call message"
- Use `ExtractReasoningMessage()` to preserve reasoning TextContent when summarizing
- Combine with fake ToolCall signatures for full multi-provider compatibility

**Example structure:**
```go
AIMessage.Parts = [
    TextContent{Text: "...", Reasoning: {Content: "..."}},  // Required by Kimi
    ToolCall{..., Reasoning: {Signature: []byte("...")}},  // Required by Gemini
]
```

**Critical Rule:** Never summarize the last body pair in a section, as this preserves reasoning signatures required by Gemini, Anthropic, and Kimi.

### Helper Functions

```go
// Check if messages contain reasoning signatures in ToolCall parts
// Returns true if any message contains Reasoning in ToolCall (NOT TextContent)
// This is specific to function calling scenarios which require thought_signature
hasToolCallReasoning := ContainsToolCallReasoning(messages)

// Extract reasoning message from AI messages
// Returns the first AI message with TextContent containing reasoning (or nil)
// Useful for preserving reasoning content for providers like Kimi (Moonshot)
reasoningMsg := ExtractReasoningMessage(messages)

// Create summarization with conditional fake signature and reasoning message
addFakeSignature := ContainsToolCallReasoning(originalMessages)
reasoningMsg := ExtractReasoningMessage(originalMessages)
summaryPair := NewBodyPairFromSummarization(summaryText, tcIDTemplate, addFakeSignature, reasoningMsg)
```

## Best Practices

1. **Validation First**: Use `NewChainAST` with `force=false` to validate chains before processing
2. **Defensive Programming**: Always check for errors from ChainAST functions
3. **Complete Tool Calls**: Ensure all tool calls have corresponding responses before sending to an LLM
4. **Section Management**: Use sections to organize conversation turns logically
5. **Testing**: Use the predefined chains and the generators of `fixtures_test.go` to test the code of this package that manipulates message chains
6. **Size Management**: Leverage size tracking to maintain efficient context windows
7. **Reasoning Preservation**: 
   - Use `ContainsToolCallReasoning()` to check if fake signatures are needed (checks only ToolCall.Reasoning)
   - Use `ExtractReasoningMessage()` to preserve reasoning TextContent for Kimi/Moonshot
8. **Last Pair Protection**: Never summarize the last (most recent) body pair in a section to preserve reasoning signatures
9. **Multi-Provider Support**: When summarizing for current turn, preserve both ToolCall and TextContent reasoning for maximum compatibility

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
