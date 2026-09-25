package cast

import (
	"fmt"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

// The builders and chains below return fresh Parts on every call: parsing shares them with the AST, and
// several units write into the parts of a parsed chain.

func chainASTText(role llms.ChatMessageType, text string) llms.MessageContent {
	return llms.MessageContent{Role: role, Parts: []llms.ContentPart{llms.TextContent{Text: text}}}
}

func chainASTSystem(text string) llms.MessageContent {
	return chainASTText(llms.ChatMessageTypeSystem, text)
}

func chainASTHuman(text string) llms.MessageContent {
	return chainASTText(llms.ChatMessageTypeHuman, text)
}

func chainASTAnswer(text string) llms.MessageContent {
	return chainASTText(llms.ChatMessageTypeAI, text)
}

func chainASTAI(parts ...llms.ContentPart) llms.MessageContent {
	return llms.MessageContent{Role: llms.ChatMessageTypeAI, Parts: parts}
}

func chainASTCallWith(id, name, args string) llms.ToolCall {
	return llms.ToolCall{ID: id, Type: "function", FunctionCall: &llms.FunctionCall{Name: name, Arguments: args}}
}

func chainASTCall(id, name string) llms.ToolCall {
	return chainASTCallWith(id, name, "{}")
}

func chainASTTool(id, name, content string) llms.MessageContent {
	return llms.MessageContent{
		Role:  llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{llms.ToolCallResponse{ToolCallID: id, Name: name, Content: content}},
	}
}

func emptyChain() []llms.MessageContent {
	return []llms.MessageContent{}
}

func systemOnlyChain() []llms.MessageContent {
	return []llms.MessageContent{chainASTSystem("You are a helpful assistant.")}
}

func humanOnlyChain() []llms.MessageContent {
	return []llms.MessageContent{chainASTHuman("Hello, can you help me?")}
}

func systemHumanChain() []llms.MessageContent {
	return []llms.MessageContent{
		chainASTSystem("You are a helpful assistant."),
		chainASTHuman("Hello, how are you?"),
	}
}

func basicConversationChain() []llms.MessageContent {
	return append(systemHumanChain(), chainASTAnswer("I'm doing well! How can I help you today?"))
}

func weatherCall() llms.ToolCall {
	return chainASTCallWith("tool-1", "get_weather", `{"location": "New York"}`)
}

func weatherResponse() llms.MessageContent {
	return chainASTTool("tool-1", "get_weather", "The weather in New York is sunny with a high of 75°F.")
}

// chainWithTool ends on a call that has no response.
func chainWithTool() []llms.MessageContent {
	return []llms.MessageContent{
		chainASTSystem("You are a helpful assistant."),
		chainASTHuman("What's the weather like?"),
		chainASTAI(weatherCall()),
	}
}

func chainWithSingleToolResponse() []llms.MessageContent {
	return append(chainWithTool(), weatherResponse())
}

// chainWithMultipleTools ends on two calls that have no response.
func chainWithMultipleTools() []llms.MessageContent {
	return []llms.MessageContent{
		chainASTSystem("You are a helpful assistant."),
		chainASTHuman("What's the weather and time in New York?"),
		chainASTAI(weatherCall(), chainASTCallWith("tool-2", "get_time", `{"location": "New York"}`)),
	}
}

func chainWithMultipleSections() []llms.MessageContent {
	return append(basicConversationChain(),
		chainASTHuman("What's the weather like?"),
		chainASTAI(weatherCall()),
		weatherResponse(),
	)
}

// chainWithMissingToolResponse answers the first of its two calls only.
func chainWithMissingToolResponse() []llms.MessageContent {
	return append(chainWithMultipleTools(), weatherResponse())
}

// chainWithUnexpectedTool answers a call that the completion before it never made.
func chainWithUnexpectedTool() []llms.MessageContent {
	return append(basicConversationChain(), weatherResponse())
}

func summarizationCall() llms.ToolCall {
	return chainASTCallWith("summary-1", "execute_task_and_return_summary",
		`{"question": "delegate and execute the task, then return the summary of the result"}`)
}

func chainWithSummarizationAndOtherPairs() []llms.MessageContent {
	return []llms.MessageContent{
		chainASTSystem("You are a helpful assistant."),
		chainASTHuman("Can you summarize and then tell me about the weather?"),
		chainASTAI(summarizationCall()),
		chainASTTool("summary-1", "execute_task_and_return_summary", "This is a summary of the previous conversation."),
		chainASTAnswer("Now I'll check the weather for you."),
		chainASTAI(weatherCall()),
		weatherResponse(),
	}
}

// providerSwitchChain is a scan in three sections: a text preamble beside two calls, one of them already in
// the Anthropic id format; a completion before a call whose tool message carries a note; a summarization.
// withReasoning adds the reasoning and signatures a provider attaches to text, calls and the note.
func providerSwitchChain(withReasoning bool) []llms.MessageContent {
	thought := func(content string) *reasoning.ContentReasoning {
		if !withReasoning {
			return nil
		}
		return &reasoning.ContentReasoning{Content: content, Signature: []byte("sig-" + content)}
	}
	signed := func(call llms.ToolCall, content string) llms.ToolCall {
		call.Reasoning = thought(content)
		return call
	}

	return []llms.MessageContent{
		chainASTSystem("You are a penetration tester."),
		chainASTHuman("Scan 10.0.0.5."),
		chainASTAI(
			llms.TextContent{Text: "Scanning the ports and saving the report.", Reasoning: thought("plan")},
			signed(chainASTCallWith("call_nmap01", "terminal", `{"input": "nmap 10.0.0.5"}`), "scan"),
			chainASTCallWith("toolu_0123456789abcdefghijklmn", "file", `{"path": "/work/report.txt"}`),
		),
		chainASTTool("call_nmap01", "terminal", "22/tcp open ssh"),
		chainASTTool("toolu_0123456789abcdefghijklmn", "file", "saved"),
		chainASTAI(llms.TextContent{Text: "Only SSH is open.", Reasoning: thought("result")}),
		chainASTHuman("Check the SSH version."),
		chainASTAnswer("Grabbing the banner."),
		chainASTAI(signed(chainASTCallWith("call_ssh02", "terminal", `{"input": "nc 10.0.0.5 22"}`), "banner")),
		{Role: llms.ChatMessageTypeTool, Parts: []llms.ContentPart{
			llms.TextContent{Text: "banner follows", Reasoning: thought("note")},
			llms.ToolCallResponse{ToolCallID: "call_ssh02", Name: "terminal", Content: "SSH-2.0-OpenSSH_8.9"},
		}},
		chainASTHuman("Summarize the findings."),
		chainASTAI(signed(chainASTCallWith("call_sum03", "execute_task_and_return_summary",
			`{"question": "delegate and execute the task, then return the summary of the result"}`), "summary")),
		chainASTTool("call_sum03", "execute_task_and_return_summary", "OpenSSH 8.9 is the only service."),
	}
}

// ChainConfig describes a chain for GenerateChain. ToolsForBodyPairs and ToolCallsPerBodyPair are indexed by
// the position of the pair in the whole chain: a pair past the end of ToolsForBodyPairs is a completion, a
// tool pair past the end of ToolCallsPerBodyPair makes one call, and a section past the end of
// BodyPairsPerSection holds one pair.
type ChainConfig struct {
	IncludeSystem           bool
	Sections                int
	BodyPairsPerSection     []int
	ToolsForBodyPairs       []bool
	ToolCallsPerBodyPair    []int
	IncludeAllToolResponses bool
}

// DefaultChainConfig is a system prompt, one question and its completion.
func DefaultChainConfig() ChainConfig {
	return ChainConfig{
		IncludeSystem:           true,
		Sections:                1,
		BodyPairsPerSection:     []int{1},
		ToolsForBodyPairs:       []bool{false},
		ToolCallsPerBodyPair:    []int{0},
		IncludeAllToolResponses: true,
	}
}

// GenerateChain numbers the calls tool-1, tool-2, ... across the chain; the n-th call of a pair is
// get_data_n with the arguments {"query": "Test query n"}, and its response, one tool message per call in
// the order of the calls, says "Response for tool-N".
func GenerateChain(config ChainConfig) []llms.MessageContent {
	var chain []llms.MessageContent
	if config.IncludeSystem {
		chain = append(chain, chainASTSystem("You are a helpful assistant."))
	}

	pair, callID := 0, 0
	for section := 1; section <= config.Sections; section++ {
		chain = append(chain, chainASTHuman(fmt.Sprintf("Question %d", section)))

		pairs := 1
		if section <= len(config.BodyPairsPerSection) {
			pairs = config.BodyPairsPerSection[section-1]
		}
		for ; pairs > 0; pairs, pair = pairs-1, pair+1 {
			if pair >= len(config.ToolsForBodyPairs) || !config.ToolsForBodyPairs[pair] {
				chain = append(chain, chainASTAnswer(fmt.Sprintf("Response to question %d", section)))
				continue
			}

			calls := 1
			if pair < len(config.ToolCallsPerBodyPair) {
				calls = config.ToolCallsPerBodyPair[pair]
			}
			var parts []llms.ContentPart
			var responses []llms.MessageContent
			for n := 1; n <= calls; n++ {
				callID++
				id, name := fmt.Sprintf("tool-%d", callID), fmt.Sprintf("get_data_%d", n)
				parts = append(parts, chainASTCallWith(id, name, fmt.Sprintf(`{"query": "Test query %d"}`, n)))
				responses = append(responses, chainASTTool(id, name, "Response for "+id))
			}
			chain = append(chain, chainASTAI(parts...))
			if config.IncludeAllToolResponses {
				chain = append(chain, responses...)
			}
		}
	}

	return chain
}

// GenerateComplexChain gives every section one pair, a tool pair of numToolCalls calls in the odd sections
// (first, third, ...) and a completion in the even ones, then drops the first numMissingResponses tool
// messages of the chain.
func GenerateComplexChain(numSections, numToolCalls, numMissingResponses int) []llms.MessageContent {
	config := ChainConfig{IncludeSystem: true, Sections: numSections, IncludeAllToolResponses: true}
	for section := 0; section < numSections; section++ {
		config.BodyPairsPerSection = append(config.BodyPairsPerSection, 1)
		config.ToolsForBodyPairs = append(config.ToolsForBodyPairs, section%2 == 0)
		config.ToolCallsPerBodyPair = append(config.ToolCallsPerBodyPair, numToolCalls)
	}

	var chain []llms.MessageContent
	for _, msg := range GenerateChain(config) {
		if msg.Role == llms.ChatMessageTypeTool && numMissingResponses > 0 {
			numMissingResponses--
			continue
		}
		chain = append(chain, msg)
	}
	return chain
}
