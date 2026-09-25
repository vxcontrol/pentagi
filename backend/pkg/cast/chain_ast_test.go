package cast

import (
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

const (
	chainASTFallback    = "the call was not handled, please try again"
	chainASTAnthropicID = `^toolu_[0-9A-Za-z]{24}$`
)

func chainASTPairTypes(ast *ChainAST) [][]BodyPairType {
	var sections [][]BodyPairType
	for _, section := range ast.Sections {
		var types []BodyPairType
		for _, pair := range section.Body {
			types = append(types, pair.Type)
		}
		sections = append(sections, types)
	}
	return sections
}

// assertChainASTSizes checks, for every section, its size, its header's size and the size of each pair, in
// that order, and then the total.
func assertChainASTSizes(t *testing.T, want [][]int, ast *ChainAST, msgAndArgs ...any) {
	t.Helper()

	var got [][]int
	for _, section := range ast.Sections {
		sizes := []int{section.Size(), section.Header.Size()}
		for _, pair := range section.Body {
			sizes = append(sizes, pair.Size())
		}
		got = append(got, sizes)
	}
	assert.Equal(t, want, got, msgAndArgs...)

	total := 0
	for _, sizes := range want {
		total += sizes[0]
	}
	assert.Equal(t, total, ast.Size(), msgAndArgs...)
}

// chainASTResponses lists, for every pair of the chain, the call ids its tool messages answer, in order.
func chainASTResponses(ast *ChainAST) [][]string {
	var pairs [][]string
	for _, section := range ast.Sections {
		for _, pair := range section.Body {
			var ids []string
			for _, msg := range pair.ToolMessages {
				for _, part := range msg.Parts {
					if resp, ok := part.(llms.ToolCallResponse); ok {
						if resp.Content == chainASTFallback {
							ids = append(ids, resp.ToolCallID+" (fallback)")
						} else {
							ids = append(ids, resp.ToolCallID)
						}
					}
				}
			}
			pairs = append(pairs, ids)
		}
	}
	return pairs
}

func chainASTDropFallbacks(chain []llms.MessageContent) []llms.MessageContent {
	var kept []llms.MessageContent
	for _, msg := range chain {
		if resp, ok := msg.Parts[0].(llms.ToolCallResponse); ok && len(msg.Parts) == 1 && resp.Content == chainASTFallback {
			continue
		}
		kept = append(kept, msg)
	}
	return kept
}

func chainASTClone(chain []llms.MessageContent) []llms.MessageContent {
	clone := make([]llms.MessageContent, len(chain))
	for i, msg := range chain {
		parts := make([]llms.ContentPart, len(msg.Parts))
		for j, part := range msg.Parts {
			if call, ok := part.(llms.ToolCall); ok && call.FunctionCall != nil {
				function := *call.FunctionCall
				call.FunctionCall = &function
				part = call
			}
			parts[j] = part
		}
		clone[i] = llms.MessageContent{Role: msg.Role, Parts: parts}
	}
	return clone
}

// chainASTUndoRename checks that got is before with only the ids of calls changed, every response following
// its call, and returns got with the old ids put back along with the renames it found.
func chainASTUndoRename(t *testing.T, before, got []llms.MessageContent) ([]llms.MessageContent, map[string]string) {
	t.Helper()
	require.Len(t, got, len(before))

	renames := map[string]string{}
	restored := make([]llms.MessageContent, len(got))
	for i, msg := range got {
		require.Len(t, msg.Parts, len(before[i].Parts), "the parts of message %d", i)
		parts := make([]llms.ContentPart, len(msg.Parts))
		for j, part := range msg.Parts {
			switch p := part.(type) {
			case llms.ToolCall:
				old, ok := before[i].Parts[j].(llms.ToolCall)
				require.True(t, ok, "part %d of message %d", j, i)
				if p.ID != old.ID {
					renames[old.ID] = p.ID
					p.ID = old.ID
				}
				part = p
			case llms.ToolCallResponse:
				old, ok := before[i].Parts[j].(llms.ToolCallResponse)
				require.True(t, ok, "part %d of message %d", j, i)
				want := old.ToolCallID
				if id, renamed := renames[old.ToolCallID]; renamed {
					want = id
				}
				assert.Equal(t, want, p.ToolCallID, "the response to %s", old.ToolCallID)
				p.ToolCallID = old.ToolCallID
				part = p
			}
			parts[j] = part
		}
		restored[i] = llms.MessageContent{Role: msg.Role, Parts: parts}
	}
	return restored, renames
}

// chainASTNilIfEmpty lets a want literal leave out an empty set of ToolCallsInfo, nil or not: the callers only
// range over a set, take its length, join it or look a key up.
func chainASTNilIfEmpty[T ~[]string | ~map[string]*ToolCallPair](set T) T {
	if len(set) == 0 {
		return nil
	}
	return set
}

func chainASTPair(ai llms.MessageContent, tools ...llms.MessageContent) *BodyPair {
	var toolMsgs []*llms.MessageContent
	for i := range tools {
		toolMsgs = append(toolMsgs, &tools[i])
	}
	return NewBodyPair(&ai, toolMsgs)
}

func TestChainAST_NewChainAST_ParsesAWellFormedChain(t *testing.T) {
	tests := []struct {
		name      string
		chain     func() []llms.MessageContent
		wantPairs [][]BodyPairType
		wantSizes [][]int
	}{
		{
			name:  "an empty chain has no sections",
			chain: emptyChain,
		},
		{
			name:      "a system message opens a section",
			chain:     systemOnlyChain,
			wantPairs: [][]BodyPairType{nil},
			wantSizes: [][]int{{28, 28}},
		},
		{
			name:      "a human message opens a section",
			chain:     humanOnlyChain,
			wantPairs: [][]BodyPairType{nil},
			wantSizes: [][]int{{23, 23}},
		},
		{
			name:      "a human message joins the system message in the header",
			chain:     systemHumanChain,
			wantPairs: [][]BodyPairType{nil},
			wantSizes: [][]int{{47, 47}},
		},
		{
			name:      "an answer makes a completion",
			chain:     basicConversationChain,
			wantPairs: [][]BodyPairType{{Completion}},
			wantSizes: [][]int{{88, 47, 41}},
		},
		{
			name:      "a call and its response make a request-response pair",
			chain:     chainWithSingleToolResponse,
			wantPairs: [][]BodyPairType{{RequestResponse}},
			wantSizes: [][]int{{172, 52, 120}},
		},
		{
			name:      "a human message after answers opens a second section",
			chain:     chainWithMultipleSections,
			wantPairs: [][]BodyPairType{{Completion}, {RequestResponse}},
			wantSizes: [][]int{{88, 47, 41}, {144, 24, 120}},
		},
		{
			name:      "a summarization pair is followed by other pairs of its section",
			chain:     chainWithSummarizationAndOtherPairs,
			wantPairs: [][]BodyPairType{{Summarization, Completion, RequestResponse}},
			wantSizes: [][]int{{455, 81, 219, 35, 120}},
		},
		{
			name: "a call without a function after a real call is kept and needs no response",
			chain: func() []llms.MessageContent {
				return []llms.MessageContent{
					chainASTSystem("You are a helpful assistant."),
					chainASTHuman("What's the weather like?"),
					chainASTAI(weatherCall(), llms.ToolCall{ID: "tool-2", Type: "function"}),
					weatherResponse(),
				}
			},
			wantPairs: [][]BodyPairType{{RequestResponse}},
			wantSizes: [][]int{{186, 52, 134}},
		},
		{
			name:      "reasoning is not counted and the note of a tool message is",
			chain:     func() []llms.MessageContent { return providerSwitchChain(true) },
			wantPairs: [][]BodyPairType{{RequestResponse, Completion}, {Completion, RequestResponse}, {Summarization}},
			wantSizes: [][]int{{297, 43, 237, 17}, {146, 22, 20, 104}, {229, 23, 206}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, force := range []bool{false, true} {
				ast, err := NewChainAST(tt.chain(), force)
				require.NoError(t, err, "force=%v", force)

				assert.Equal(t, tt.wantPairs, chainASTPairTypes(ast), "force=%v", force)
				assert.Equal(t, tt.chain(), ast.Messages(), "force=%v", force)
				assertChainASTSizes(t, tt.wantSizes, ast, "force=%v", force)
				for _, section := range ast.Sections {
					for i, pair := range section.Body {
						assert.True(t, pair.IsValid(), "force=%v: pair %d", force, i)
					}
				}
			}
		})
	}
}

func TestChainAST_NewChainAST_RepairsAMalformedChainOnlyWhenForced(t *testing.T) {
	tests := []struct {
		name      string
		chain     []llms.MessageContent
		strictErr string
		want      []llms.MessageContent
		wantPairs [][]BodyPairType
		wantSizes [][]int // of the forced AST; nil where the repair adds a message whose bytes it does not count
	}{
		{
			name:      "a call left without a response gets a fallback response",
			chain:     chainWithTool(),
			strictErr: "tool calls with IDs [tool-1] have no response",
			want:      append(chainWithTool(), chainASTTool("tool-1", "get_weather", chainASTFallback)),
			wantPairs: [][]BodyPairType{{RequestResponse}},
		},
		{
			name:      "two calls left without a response get one fallback each",
			chain:     chainWithMultipleTools(),
			strictErr: "tool calls with IDs [tool-1, tool-2] have no response",
			want: append(chainWithMultipleTools(),
				chainASTTool("tool-1", "get_weather", chainASTFallback),
				chainASTTool("tool-2", "get_time", chainASTFallback),
			),
			wantPairs: [][]BodyPairType{{RequestResponse}},
		},
		{
			name:      "the second of two calls left without a response gets a fallback",
			chain:     chainWithMissingToolResponse(),
			strictErr: "tool calls with IDs [tool-2] have no response",
			want:      append(chainWithMissingToolResponse(), chainASTTool("tool-2", "get_time", chainASTFallback)),
			wantPairs: [][]BodyPairType{{RequestResponse}},
		},
		{
			name: "a summarization call left without a response gets a fallback",
			chain: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTHuman("question"),
				chainASTAI(chainASTCall("summary-1", "execute_task_and_return_summary")),
			},
			strictErr: "tool calls with IDs [summary-1] have no response",
			want: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTHuman("question"),
				chainASTAI(chainASTCall("summary-1", "execute_task_and_return_summary")),
				chainASTTool("summary-1", "execute_task_and_return_summary", chainASTFallback),
			},
			wantPairs: [][]BodyPairType{{Summarization}},
		},
		{
			name: "three calls made out of id order get fallbacks in id order",
			chain: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTHuman("question"),
				chainASTAI(
					chainASTCall("tool-2", "get_time"),
					chainASTCall("tool-1", "get_weather"),
					chainASTCall("tool-4", "get_news"),
					chainASTCall("tool-3", "get_maps"),
				),
				chainASTTool("tool-4", "get_news", "calm"),
			},
			strictErr: "tool calls with IDs [tool-1, tool-2, tool-3] have no response",
			want: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTHuman("question"),
				chainASTAI(
					chainASTCall("tool-2", "get_time"),
					chainASTCall("tool-1", "get_weather"),
					chainASTCall("tool-4", "get_news"),
					chainASTCall("tool-3", "get_maps"),
				),
				chainASTTool("tool-4", "get_news", "calm"),
				chainASTTool("tool-1", "get_weather", chainASTFallback),
				chainASTTool("tool-2", "get_time", chainASTFallback),
				chainASTTool("tool-3", "get_maps", chainASTFallback),
			},
			wantPairs: [][]BodyPairType{{RequestResponse}},
		},
		{
			name: "a call pending when the next human message arrives gets a fallback",
			chain: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTHuman("question"),
				chainASTAI(chainASTCall("tool-1", "get_weather")),
				chainASTHuman("follow-up"),
				chainASTAnswer("answer"),
			},
			strictErr: "tool calls with IDs [tool-1] have no response",
			want: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTHuman("question"),
				chainASTAI(chainASTCall("tool-1", "get_weather")),
				chainASTTool("tool-1", "get_weather", chainASTFallback),
				chainASTHuman("follow-up"),
				chainASTAnswer("answer"),
			},
			wantPairs: [][]BodyPairType{{RequestResponse}, {Completion}},
		},
		{
			name: "a call pending when the next AI message arrives gets a fallback",
			chain: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTHuman("question"),
				chainASTAI(chainASTCall("tool-1", "get_weather")),
				chainASTAnswer("answer"),
			},
			strictErr: "tool calls with IDs [tool-1] have no response",
			want: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTHuman("question"),
				chainASTAI(chainASTCall("tool-1", "get_weather")),
				chainASTTool("tool-1", "get_weather", chainASTFallback),
				chainASTAnswer("answer"),
			},
			wantPairs: [][]BodyPairType{{RequestResponse, Completion}},
		},
		{
			name: "consecutive human messages are merged",
			chain: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTHuman("first"),
				chainASTHuman("second"),
			},
			strictErr: "double human messages in the middle of a chain",
			want: []llms.MessageContent{
				chainASTSystem("system"),
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{
					llms.TextContent{Text: "first"},
					llms.TextContent{Text: "second"},
				}},
			},
			wantPairs: [][]BodyPairType{nil},
			wantSizes: [][]int{{17, 17}},
		},
		{
			name:      "a tool message that answers no tool call is dropped",
			chain:     chainWithUnexpectedTool(),
			strictErr: "unexpected tool message without a preceding AI message with tool calls",
			want:      basicConversationChain(),
			wantPairs: [][]BodyPairType{{Completion}},
			wantSizes: [][]int{{88, 47, 41}},
		},
		{
			name: "a human message after an answer moves into a header without one",
			chain: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTAnswer("answer"),
				chainASTHuman("question"),
			},
			strictErr: "got human message after AI message in the middle of a chain",
			want: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTHuman("question"),
				chainASTAnswer("answer"),
			},
			wantPairs: [][]BodyPairType{{Completion}},
			wantSizes: [][]int{{20, 14, 6}},
		},
		{
			name: "responses to calls the AI message never made get the calls added",
			chain: []llms.MessageContent{
				chainASTHuman("question"),
				chainASTAI(chainASTCall("tool-1", "get_weather")),
				chainASTTool("tool-1", "get_weather", "sunny"),
				{Role: llms.ChatMessageTypeTool, Parts: []llms.ContentPart{
					llms.ToolCallResponse{ToolCallID: "tool-3", Name: "get_news", Content: "quiet"},
					llms.ToolCallResponse{ToolCallID: "tool-2", Name: "get_time", Content: "noon"},
					llms.ToolCallResponse{ToolCallID: "tool-4", Name: "get_maps", Content: "far"},
				}},
			},
			strictErr: "tool calls with IDs [tool-2, tool-3, tool-4] have no response",
			want: []llms.MessageContent{
				chainASTHuman("question"),
				chainASTAI(
					chainASTCall("tool-1", "get_weather"),
					llms.ToolCall{ID: "tool-2", FunctionCall: &llms.FunctionCall{Name: "get_time", Arguments: "{}"}},
					llms.ToolCall{ID: "tool-3", FunctionCall: &llms.FunctionCall{Name: "get_news", Arguments: "{}"}},
					llms.ToolCall{ID: "tool-4", FunctionCall: &llms.FunctionCall{Name: "get_maps", Arguments: "{}"}},
				),
				chainASTTool("tool-1", "get_weather", "sunny"),
				{Role: llms.ChatMessageTypeTool, Parts: []llms.ContentPart{
					llms.ToolCallResponse{ToolCallID: "tool-3", Name: "get_news", Content: "quiet"},
					llms.ToolCallResponse{ToolCallID: "tool-2", Name: "get_time", Content: "noon"},
					llms.ToolCallResponse{ToolCallID: "tool-4", Name: "get_maps", Content: "far"},
				}},
			},
			wantPairs: [][]BodyPairType{{RequestResponse}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strict, err := NewChainAST(tt.chain, false)
			assert.EqualError(t, err, tt.strictErr)
			assert.Nil(t, strict)

			ast, err := NewChainAST(tt.chain, true)
			require.NoError(t, err)
			assert.Equal(t, tt.wantPairs, chainASTPairTypes(ast))
			assert.Equal(t, tt.want, ast.Messages())
			if tt.wantSizes != nil {
				assertChainASTSizes(t, tt.wantSizes, ast)
			}

			_, err = NewChainAST(ast.Messages(), false)
			assert.NoError(t, err, "the repaired chain must parse without force")
		})
	}
}

func TestChainAST_NewChainAST_RejectsAChainItCannotRepair(t *testing.T) {
	tests := []struct {
		name    string
		chain   []llms.MessageContent
		wantErr string
	}{
		{
			name:    "a chain opened by an AI message is rejected",
			chain:   []llms.MessageContent{chainASTAnswer("answer"), chainASTHuman("question")},
			wantErr: "unexpected chain begin: first message must be System or Human, got ai",
		},
		{
			name:    "a chain opened by a tool message is rejected",
			chain:   []llms.MessageContent{chainASTTool("tool-1", "get_weather", "sunny")},
			wantErr: "unexpected chain begin: first message must be System or Human, got tool",
		},
		{
			name: "a system message after the conversation began is rejected",
			chain: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTHuman("question"),
				chainASTAnswer("answer"),
				chainASTSystem("another system"),
			},
			wantErr: "unexpected system message in the middle of a chain",
		},
		{
			name:    "a message of a role the chain does not know is rejected",
			chain:   []llms.MessageContent{chainASTHuman("question"), chainASTText(llms.ChatMessageTypeGeneric, "note")},
			wantErr: "unexpected message role: generic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, force := range []bool{false, true} {
				ast, err := NewChainAST(tt.chain, force)
				assert.EqualError(t, err, tt.wantErr, "force=%v", force)
				assert.Nil(t, ast, "force=%v", force)
			}
		})
	}
}

func TestChainAST_NewChainAST_ParsesOrRepairsAGeneratedChain(t *testing.T) {
	answered := DefaultChainConfig()
	answered.Sections = 3
	answered.BodyPairsPerSection = []int{1, 2, 2}
	answered.ToolsForBodyPairs = []bool{false, true, false, true, true}
	answered.ToolCallsPerBodyPair = []int{0, 2, 0, 3, 1}
	unanswered := answered
	unanswered.IncludeAllToolResponses = false

	tests := []struct {
		name          string
		chain         func() []llms.MessageContent
		strictErr     string
		wantPairs     [][]BodyPairType
		wantResponses [][]string
		wantSizes     [][]int // of the chain, repaired where it had to be, parsed without force
	}{
		{
			name:      "several pairs and calls per section, all answered, parse as they are",
			chain:     func() []llms.MessageContent { return GenerateChain(answered) },
			wantPairs: [][]BodyPairType{{Completion}, {RequestResponse, Completion}, {RequestResponse, RequestResponse}},
			wantResponses: [][]string{
				nil,
				{"tool-1", "tool-2"},
				nil,
				{"tool-3", "tool-4", "tool-5"},
				{"tool-6"},
			},
			wantSizes: [][]int{{60, 38, 22}, {200, 10, 168, 22}, {346, 10, 252, 84}},
		},
		{
			name:      "several pairs and calls per section, none answered, get a fallback for every call",
			chain:     func() []llms.MessageContent { return GenerateChain(unanswered) },
			strictErr: "tool calls with IDs [tool-1, tool-2] have no response",
			wantPairs: [][]BodyPairType{{Completion}, {RequestResponse, Completion}, {RequestResponse, RequestResponse}},
			wantResponses: [][]string{
				nil,
				{"tool-1 (fallback)", "tool-2 (fallback)"},
				nil,
				{"tool-3 (fallback)", "tool-4 (fallback)", "tool-5 (fallback)"},
				{"tool-6 (fallback)"},
			},
			wantSizes: [][]int{{60, 38, 22}, {246, 10, 214, 22}, {438, 10, 321, 107}},
		},
		{
			name:      "a repaired first section leaves the answered third one as it was",
			chain:     func() []llms.MessageContent { return GenerateComplexChain(3, 2, 2) },
			strictErr: "tool calls with IDs [tool-1, tool-2] have no response",
			wantPairs: [][]BodyPairType{{RequestResponse}, {Completion}, {RequestResponse}},
			wantResponses: [][]string{
				{"tool-1 (fallback)", "tool-2 (fallback)"},
				nil,
				{"tool-3", "tool-4"},
			},
			wantSizes: [][]int{{252, 38, 214}, {32, 10, 22}, {178, 10, 168}},
		},
		{
			name:      "the first seven responses of five alternating sections missing are replaced by fallbacks",
			chain:     func() []llms.MessageContent { return GenerateComplexChain(5, 3, 7) },
			strictErr: "tool calls with IDs [tool-1, tool-2, tool-3] have no response",
			wantPairs: [][]BodyPairType{{RequestResponse}, {Completion}, {RequestResponse}, {Completion}, {RequestResponse}},
			wantResponses: [][]string{
				{"tool-1 (fallback)", "tool-2 (fallback)", "tool-3 (fallback)"},
				nil,
				{"tool-4 (fallback)", "tool-5 (fallback)", "tool-6 (fallback)"},
				nil,
				{"tool-8", "tool-9", "tool-7 (fallback)"},
			},
			wantSizes: [][]int{{359, 38, 321}, {32, 10, 22}, {331, 10, 321}, {32, 10, 22}, {285, 10, 275}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strict, err := NewChainAST(tt.chain(), false)
			if tt.strictErr != "" {
				assert.EqualError(t, err, tt.strictErr)
				assert.Nil(t, strict)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.chain(), strict.Messages())
			}

			ast, err := NewChainAST(tt.chain(), true)
			require.NoError(t, err)
			assert.Equal(t, tt.wantPairs, chainASTPairTypes(ast))
			assert.Equal(t, tt.wantResponses, chainASTResponses(ast))
			assert.Equal(t, tt.chain(), chainASTDropFallbacks(ast.Messages()), "the repair only adds responses")

			reparsed, err := NewChainAST(ast.Messages(), false)
			require.NoError(t, err, "the repaired chain must parse without force")
			assertChainASTSizes(t, tt.wantSizes, reparsed)
		})
	}
}

// Header, BodyPair, ChainSection and ChainAST each return their own messages in chain order.
func TestChainAST_Messages_ReturnsHeaderThenBodyInOrder(t *testing.T) {
	chain := chainWithMultipleSections()
	ast, err := NewChainAST(chain, false)
	require.NoError(t, err)
	require.Len(t, ast.Sections, 2)

	assert.Equal(t, chain[0:2], ast.Sections[0].Header.Messages())
	assert.Equal(t, chain[2:3], ast.Sections[0].Body[0].Messages())
	assert.Equal(t, chain[0:3], ast.Sections[0].Messages())
	assert.Equal(t, chain[3:4], ast.Sections[1].Header.Messages())
	assert.Equal(t, chain[4:6], ast.Sections[1].Body[0].Messages())
	assert.Equal(t, chain[3:6], ast.Sections[1].Messages())
	assert.Equal(t, chain, ast.Messages())
}

func TestChainAST_String_DescribesTheHeaderAndEveryPair(t *testing.T) {
	ast, err := NewChainAST(append(chainWithSummarizationAndOtherPairs(), chainASTHuman("Thanks.")), false)
	require.NoError(t, err)

	assert.Equal(t, `ChainAST {
  Section 0 {
    Header {
      SystemMessage
      HumanMessage
    }
    Body {
      BodyPair 0 (Summarization) {
        AIMessage
        ToolMessages: 1
      }
      BodyPair 1 (Completion) {
        AIMessage
        ToolMessages: 0
      }
      BodyPair 2 (RequestResponse) {
        AIMessage
        ToolMessages: 1
      }
    }
  }
  Section 1 {
    Header {
      HumanMessage
    }
    Body {
    }
  }
}
`, ast.String())
}

func TestChainAST_BodyPairTypeString_NamesEachType(t *testing.T) {
	assert.Equal(t, "request-response", RequestResponse.String())
	assert.Equal(t, "completion", Completion.String())
	assert.Equal(t, "summarization", Summarization.String())
	assert.Equal(t, "unknown", BodyPairType(3).String())
}

func TestChainAST_AppendHumanMessage_OpensOrExtendsTheLastSection(t *testing.T) {
	tests := []struct {
		name      string
		chain     []llms.MessageContent
		want      []llms.MessageContent
		wantSizes [][]int
	}{
		{
			name:      "an empty chain gets its first section",
			chain:     []llms.MessageContent{},
			want:      []llms.MessageContent{chainASTHuman("hello")},
			wantSizes: [][]int{{5, 5}},
		},
		{
			name:      "a header with only a system message takes the text",
			chain:     []llms.MessageContent{chainASTSystem("system")},
			want:      []llms.MessageContent{chainASTSystem("system"), chainASTHuman("hello")},
			wantSizes: [][]int{{11, 11}},
		},
		{
			name:  "a header without answers gets the text appended to its human message",
			chain: []llms.MessageContent{chainASTSystem("system"), chainASTHuman("question")},
			want: []llms.MessageContent{
				chainASTSystem("system"),
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{
					llms.TextContent{Text: "question"},
					llms.TextContent{Text: "hello"},
				}},
			},
			wantSizes: [][]int{{19, 19}},
		},
		{
			name:      "the last of several sections, with answers, is followed by a new one",
			chain:     chainWithMultipleSections(),
			want:      append(chainWithMultipleSections(), chainASTHuman("hello")),
			wantSizes: [][]int{{88, 47, 41}, {144, 24, 120}, {5, 5}},
		},
		{
			name: "the last of several sections, without answers, gets the text appended",
			chain: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTHuman("question"),
				chainASTAnswer("answer"),
				chainASTHuman("follow-up"),
			},
			want: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTHuman("question"),
				chainASTAnswer("answer"),
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{
					llms.TextContent{Text: "follow-up"},
					llms.TextContent{Text: "hello"},
				}},
			},
			wantSizes: [][]int{{20, 14, 6}, {14, 14}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := NewChainAST(tt.chain, false)
			require.NoError(t, err)

			ast.AppendHumanMessage("hello")

			assert.Equal(t, tt.want, ast.Messages())
			assertChainASTSizes(t, tt.wantSizes, ast)
		})
	}
}

func TestChainAST_AddToolResponse_ReplacesOrAddsTheResponseWhereverTheCallIs(t *testing.T) {
	ast, err := NewChainAST([]llms.MessageContent{
		chainASTSystem("system"),
		chainASTHuman("question"),
		chainASTAI(chainASTCall("tool-1", "get_weather"), chainASTCall("tool-2", "get_time")),
		chainASTTool("tool-1", "get_weather", "initial"),
		chainASTTool("tool-2", "get_time", "initial"),
		chainASTAI(chainASTCall("tool-3", "get_news")),
		chainASTTool("tool-3", "get_news", "initial"),
		chainASTHuman("follow-up"),
		chainASTAnswer("checking the map"),
	}, false)
	require.NoError(t, err)
	awaiting := chainASTAI(chainASTCall("tool-4", "get_maps"), chainASTCall("tool-5", "get_route"))
	ast.Sections[1].AddBodyPair(NewBodyPair(&awaiting, nil))
	assertChainASTSizes(t, [][]int{{155, 14, 96, 45}, {74, 9, 16, 49}}, ast)

	steps := []struct {
		name      string
		id        string
		toolName  string
		content   string
		wantErr   string
		wantSizes [][]int
	}{
		{
			name:      "the response in the second tool message of a pair is replaced",
			id:        "tool-2",
			toolName:  "get_time",
			content:   "noon",
			wantSizes: [][]int{{152, 14, 93, 45}, {74, 9, 16, 49}},
		},
		{
			name:      "the response in a later pair of the same section is replaced",
			id:        "tool-3",
			toolName:  "get_news",
			content:   "quiet",
			wantSizes: [][]int{{150, 14, 93, 43}, {74, 9, 16, 49}},
		},
		{
			name:      "a pair without responses gets a tool message",
			id:        "tool-4",
			toolName:  "get_maps",
			content:   "far",
			wantSizes: [][]int{{150, 14, 93, 43}, {91, 9, 16, 66}},
		},
		{
			name:      "the next response joins the last tool message of the pair",
			id:        "tool-5",
			toolName:  "get_route",
			content:   "north",
			wantSizes: [][]int{{150, 14, 93, 43}, {111, 9, 16, 86}},
		},
		{
			name:      "a response that shares its tool message gets the new content and keeps its name",
			id:        "tool-5",
			toolName:  "get_directions",
			content:   "south-west",
			wantSizes: [][]int{{150, 14, 93, 43}, {116, 9, 16, 91}},
		},
		{
			name:      "an unknown call is rejected",
			id:        "missing",
			toolName:  "get_news",
			content:   "late",
			wantErr:   "tool call with ID missing not found",
			wantSizes: [][]int{{150, 14, 93, 43}, {116, 9, 16, 91}},
		},
	}

	for _, step := range steps {
		err := ast.AddToolResponse(step.id, step.toolName, step.content)
		if step.wantErr != "" {
			assert.EqualError(t, err, step.wantErr, step.name)
		} else {
			assert.NoError(t, err, step.name)
		}
		assertChainASTSizes(t, step.wantSizes, ast, step.name)
	}

	assert.Equal(t, []llms.MessageContent{
		chainASTSystem("system"),
		chainASTHuman("question"),
		chainASTAI(chainASTCall("tool-1", "get_weather"), chainASTCall("tool-2", "get_time")),
		chainASTTool("tool-1", "get_weather", "initial"),
		chainASTTool("tool-2", "get_time", "noon"),
		chainASTAI(chainASTCall("tool-3", "get_news")),
		chainASTTool("tool-3", "get_news", "quiet"),
		chainASTHuman("follow-up"),
		chainASTAnswer("checking the map"),
		chainASTAI(chainASTCall("tool-4", "get_maps"), chainASTCall("tool-5", "get_route")),
		{Role: llms.ChatMessageTypeTool, Parts: []llms.ContentPart{
			llms.ToolCallResponse{ToolCallID: "tool-4", Name: "get_maps", Content: "far"},
			llms.ToolCallResponse{ToolCallID: "tool-5", Name: "get_route", Content: "south-west"},
		}},
	}, ast.Messages())

	reparsed, err := NewChainAST(ast.Messages(), false)
	require.NoError(t, err)
	assertChainASTSizes(t, [][]int{{150, 14, 93, 43}, {116, 9, 16, 91}}, reparsed)
}

// The sizes of the forced AST are left unasserted: the repair does not count the fallbacks it adds.
func TestChainAST_AddToolResponse_ReplacesTheFallbacksOfAForcedRepair(t *testing.T) {
	ast, err := NewChainAST(chainWithMultipleTools(), true)
	require.NoError(t, err)

	steps := []struct {
		name     string
		id       string
		toolName string
		content  string
		want     []llms.MessageContent
	}{
		{
			name:     "the first fallback is replaced",
			id:       "tool-1",
			toolName: "get_weather",
			content:  "sunny",
			want: append(chainWithMultipleTools(),
				chainASTTool("tool-1", "get_weather", "sunny"),
				chainASTTool("tool-2", "get_time", chainASTFallback),
			),
		},
		{
			name:     "the second fallback is replaced",
			id:       "tool-2",
			toolName: "get_time",
			content:  "noon",
			want: append(chainWithMultipleTools(),
				chainASTTool("tool-1", "get_weather", "sunny"),
				chainASTTool("tool-2", "get_time", "noon"),
			),
		},
	}

	for _, step := range steps {
		require.NoError(t, ast.AddToolResponse(step.id, step.toolName, step.content), step.name)
		assert.Equal(t, step.want, ast.Messages(), step.name)
		assert.Equal(t, []llms.ToolCallResponse{{ToolCallID: step.id, Name: step.toolName, Content: step.content}},
			ast.FindToolCallResponses(step.id), step.name)
	}

	reparsed, err := NewChainAST(ast.Messages(), false)
	require.NoError(t, err)
	assertChainASTSizes(t, [][]int{{203, 68, 135}}, reparsed)
}

func TestChainAST_Size_StaysExactWhileAConversationIsBuilt(t *testing.T) {
	system, newSystem := chainASTSystem("system"), chainASTSystem("new system")
	question, summarize := chainASTHuman("question"), chainASTHuman("summarize")
	call, response := chainASTAI(chainASTCall("tool-1", "get_weather")), chainASTTool("tool-1", "get_weather", "sunny")
	ast := &ChainAST{}

	steps := []struct {
		name      string
		edit      func()
		wantSizes [][]int
	}{
		{
			name:      "a system message opens a section",
			edit:      func() { ast.AddSection(NewChainSection(NewHeader(&system, nil), nil)) },
			wantSizes: [][]int{{6, 6}},
		},
		{
			name:      "the question joins its header",
			edit:      func() { ast.AppendHumanMessage("question") },
			wantSizes: [][]int{{14, 14}},
		},
		{
			name:      "an answer is added",
			edit:      func() { ast.Sections[0].AddBodyPair(NewBodyPairFromCompletion("answer")) },
			wantSizes: [][]int{{20, 14, 6}},
		},
		{
			name:      "a follow-up opens a section",
			edit:      func() { ast.AppendHumanMessage("follow-up") },
			wantSizes: [][]int{{20, 14, 6}, {9, 9}},
		},
		{
			name:      "a tool exchange is added",
			edit:      func() { ast.Sections[1].AddBodyPair(NewBodyPair(&call, []*llms.MessageContent{&response})) },
			wantSizes: [][]int{{20, 14, 6}, {58, 9, 49}},
		},
		{
			name:      "the first header is replaced",
			edit:      func() { ast.Sections[0].SetHeader(NewHeader(&newSystem, &question)) },
			wantSizes: [][]int{{24, 18, 6}, {58, 9, 49}},
		},
		{
			name: "a section holding a summary is added",
			edit: func() {
				summary := NewBodyPairFromSummarization("summary", "call_{r:24:x}", false, nil)
				ast.AddSection(NewChainSection(NewHeader(nil, &summarize), []*BodyPair{summary}))
			},
			wantSizes: [][]int{{24, 18, 6}, {58, 9, 49}, {228, 9, 219}},
		},
	}

	for _, step := range steps {
		step.edit()
		assertChainASTSizes(t, step.wantSizes, ast, step.name)
	}

	summaryCall, ok := ast.Sections[2].Body[0].AIMessage.Parts[0].(llms.ToolCall)
	require.True(t, ok)
	assert.Regexp(t, `^call_[0-9A-Za-z]{24}$`, summaryCall.ID)
	assert.Equal(t, []llms.MessageContent{
		chainASTSystem("new system"),
		chainASTHuman("question"),
		chainASTAnswer("answer"),
		chainASTHuman("follow-up"),
		chainASTAI(chainASTCall("tool-1", "get_weather")),
		chainASTTool("tool-1", "get_weather", "sunny"),
		chainASTHuman("summarize"),
		chainASTAI(chainASTCallWith(summaryCall.ID, "execute_task_and_return_summary",
			`{"question": "delegate and execute the task, then return the summary of the result"}`)),
		chainASTTool(summaryCall.ID, "execute_task_and_return_summary", "summary"),
	}, ast.Messages())

	reparsed, err := NewChainAST(ast.Messages(), false)
	require.NoError(t, err)
	assertChainASTSizes(t, [][]int{{24, 18, 6}, {58, 9, 49}, {228, 9, 219}}, reparsed)
}

func TestChainAST_FindToolCallResponses_SearchesOnlyRequestResponsePairs(t *testing.T) {
	ast, err := NewChainAST(append(chainWithSummarizationAndOtherPairs(),
		chainASTHuman("And tomorrow?"),
		chainASTAI(chainASTCallWith("tool-1", "get_weather", `{"location": "New York", "day": "tomorrow"}`)),
		chainASTTool("tool-1", "get_weather", "Rain tomorrow."),
	), false)
	require.NoError(t, err)

	tests := []struct {
		name string
		id   string
		want []llms.ToolCallResponse
	}{
		{
			name: "an id reused by a later section finds every response in chain order",
			id:   "tool-1",
			want: []llms.ToolCallResponse{
				{ToolCallID: "tool-1", Name: "get_weather", Content: "The weather in New York is sunny with a high of 75°F."},
				{ToolCallID: "tool-1", Name: "get_weather", Content: "Rain tomorrow."},
			},
		},
		{
			name: "the summarization call is not searched",
			id:   "summary-1",
		},
		{
			name: "an unknown id finds nothing",
			id:   "missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ast.FindToolCallResponses(tt.id))
		})
	}
}

func TestChainAST_GetToolCallsInfo_SortsPendingAndUnmatchedCallsAndPairsTheRest(t *testing.T) {
	response := func(id, name, content string) llms.ToolCallResponse {
		return llms.ToolCallResponse{ToolCallID: id, Name: name, Content: content}
	}

	tests := []struct {
		name string
		pair *BodyPair
		want ToolCallsInfo
	}{
		{
			name: "answered calls are paired with their responses",
			pair: chainASTPair(
				chainASTAI(llms.TextContent{Text: "checking"}, chainASTCall("tool-1", "get_weather"), chainASTCall("tool-2", "get_time")),
				chainASTTool("tool-2", "get_time", "noon"),
				chainASTTool("tool-1", "get_weather", "sunny"),
			),
			want: ToolCallsInfo{
				CompletedToolCalls: map[string]*ToolCallPair{
					"tool-1": {ToolCall: chainASTCall("tool-1", "get_weather"), Response: response("tool-1", "get_weather", "sunny")},
					"tool-2": {ToolCall: chainASTCall("tool-2", "get_time"), Response: response("tool-2", "get_time", "noon")},
				},
			},
		},
		{
			name: "calls without a response are pending in id order and a call without a function is not one",
			pair: chainASTPair(
				chainASTAI(
					chainASTCall("tool-2", "get_time"),
					chainASTCall("tool-4", "get_news"),
					chainASTCall("tool-3", "get_maps"),
					chainASTCall("tool-1", "get_weather"),
					llms.ToolCall{ID: "tool-5", Type: "function"},
				),
				chainASTTool("tool-1", "get_weather", "sunny"),
			),
			want: ToolCallsInfo{
				PendingToolCallIDs: []string{"tool-2", "tool-3", "tool-4"},
				PendingToolCalls: map[string]*ToolCallPair{
					"tool-2": {ToolCall: chainASTCall("tool-2", "get_time")},
					"tool-3": {ToolCall: chainASTCall("tool-3", "get_maps")},
					"tool-4": {ToolCall: chainASTCall("tool-4", "get_news")},
				},
				CompletedToolCalls: map[string]*ToolCallPair{
					"tool-1": {ToolCall: chainASTCall("tool-1", "get_weather"), Response: response("tool-1", "get_weather", "sunny")},
				},
			},
		},
		{
			name: "responses to calls never made are unmatched in id order",
			pair: chainASTPair(
				chainASTAI(chainASTCall("tool-1", "get_weather")),
				llms.MessageContent{Role: llms.ChatMessageTypeTool, Parts: []llms.ContentPart{
					llms.TextContent{Text: "output follows"},
					response("tool-3", "get_news", "quiet"),
					response("tool-1", "get_weather", "sunny"),
					response("tool-2", "get_time", "noon"),
					response("tool-4", "get_maps", "far"),
				}},
			),
			want: ToolCallsInfo{
				UnmatchedToolCallIDs: []string{"tool-2", "tool-3", "tool-4"},
				CompletedToolCalls: map[string]*ToolCallPair{
					"tool-1": {ToolCall: chainASTCall("tool-1", "get_weather"), Response: response("tool-1", "get_weather", "sunny")},
				},
				UnmatchedToolCalls: map[string]*ToolCallPair{
					"tool-2": {Response: response("tool-2", "get_time", "noon")},
					"tool-3": {Response: response("tool-3", "get_news", "quiet")},
					"tool-4": {Response: response("tool-4", "get_maps", "far")},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pair.GetToolCallsInfo()
			got.PendingToolCallIDs = chainASTNilIfEmpty(got.PendingToolCallIDs)
			got.UnmatchedToolCallIDs = chainASTNilIfEmpty(got.UnmatchedToolCallIDs)
			got.PendingToolCalls = chainASTNilIfEmpty(got.PendingToolCalls)
			got.CompletedToolCalls = chainASTNilIfEmpty(got.CompletedToolCalls)
			got.UnmatchedToolCalls = chainASTNilIfEmpty(got.UnmatchedToolCalls)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestChainAST_IsValid_RequiresTheResponsesThePairTypeCallsFor(t *testing.T) {
	tests := []struct {
		name string
		pair *BodyPair
		want bool
	}{
		{
			name: "a completion is valid",
			pair: chainASTPair(chainASTAnswer("done")),
			want: true,
		},
		{
			name: "a request-response pair with every call answered is valid",
			pair: chainASTPair(
				chainASTAI(chainASTCall("tool-1", "get_weather"), chainASTCall("tool-2", "get_time")),
				chainASTTool("tool-1", "get_weather", "sunny"),
				chainASTTool("tool-2", "get_time", "noon"),
			),
			want: true,
		},
		{
			name: "a request-response pair whose responses have not arrived is not",
			pair: chainASTPair(chainASTAI(chainASTCall("tool-1", "get_weather"))),
		},
		{
			name: "a request-response pair with one of its two calls answered is not",
			pair: chainASTPair(
				chainASTAI(chainASTCall("tool-1", "get_weather"), chainASTCall("tool-2", "get_time")),
				chainASTTool("tool-1", "get_weather", "sunny"),
			),
		},
		{
			name: "a request-response pair answering a call it never made is not",
			pair: chainASTPair(
				chainASTAI(chainASTCall("tool-1", "get_weather")),
				chainASTTool("tool-1", "get_weather", "sunny"),
				chainASTTool("tool-2", "get_time", "noon"),
			),
		},
		{
			name: "a summarization pair with its response is valid",
			pair: chainASTPair(
				chainASTAI(chainASTCall("summary-1", "execute_task_and_return_summary")),
				chainASTTool("summary-1", "execute_task_and_return_summary", "summary"),
			),
			want: true,
		},
		{
			name: "a summarization pair whose response has not arrived is not",
			pair: chainASTPair(chainASTAI(chainASTCall("summary-1", "execute_task_and_return_summary"))),
		},
		{
			name: "a pair of none of the three types is not",
			pair: &BodyPair{Type: BodyPairType(3), AIMessage: &llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{llms.TextContent{Text: "done"}},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.pair.IsValid())
		})
	}
}

func TestChainAST_NewHeader_SumsTheSizesOfItsMessages(t *testing.T) {
	system, human := chainASTSystem("system"), chainASTHuman("question")

	tests := []struct {
		name     string
		system   *llms.MessageContent
		human    *llms.MessageContent
		wantSize int
	}{
		{name: "a system and a human message are both counted", system: &system, human: &human, wantSize: 14},
		{name: "a human message alone is counted", human: &human, wantSize: 8},
		{name: "a system message alone is counted", system: &system, wantSize: 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := NewHeader(tt.system, tt.human)
			assert.Same(t, tt.system, header.SystemMessage)
			assert.Same(t, tt.human, header.HumanMessage)
			assert.Equal(t, tt.wantSize, header.Size())
		})
	}
}

func TestChainAST_NewBodyPair_ClassifiesByItsToolCalls(t *testing.T) {
	tests := []struct {
		name      string
		parts     []llms.ContentPart
		tools     []llms.MessageContent
		wantType  BodyPairType
		wantParts []llms.ContentPart // nil: the parts are kept as given
		wantSize  int
		invalid   bool
	}{
		{
			name:     "text parts make a completion",
			parts:    []llms.ContentPart{llms.TextContent{Text: "first"}, llms.TextContent{Text: "second"}},
			wantType: Completion,
			wantSize: 11,
		},
		{
			name:     "a tool message beside text parts leaves it an invalid completion",
			parts:    []llms.ContentPart{llms.TextContent{Text: "answer"}},
			tools:    []llms.MessageContent{chainASTText(llms.ChatMessageTypeTool, "sunny")},
			wantType: Completion,
			wantSize: 11,
			invalid:  true,
		},
		{
			name:     "a tool call makes a request-response pair",
			parts:    []llms.ContentPart{chainASTCall("tool-1", "get_weather")},
			tools:    []llms.MessageContent{chainASTTool("tool-1", "get_weather", "sunny")},
			wantType: RequestResponse,
			wantSize: 49,
		},
		{
			name:     "the summarization call makes a summarization pair",
			parts:    []llms.ContentPart{chainASTCall("summary-1", "execute_task_and_return_summary")},
			tools:    []llms.MessageContent{chainASTTool("summary-1", "execute_task_and_return_summary", "summary")},
			wantType: Summarization,
			wantSize: 97,
		},
		{
			name: "tool calls without a function are dropped",
			parts: []llms.ContentPart{
				llms.ToolCall{ID: "a", Type: "function"},
				llms.ToolCall{ID: "b", Type: "function"},
				llms.TextContent{Text: "keep me"},
			},
			wantType:  Completion,
			wantParts: []llms.ContentPart{llms.TextContent{Text: "keep me"}},
			wantSize:  7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ai := chainASTAI(append([]llms.ContentPart(nil), tt.parts...)...)
			var tools []*llms.MessageContent
			for i := range tt.tools {
				tools = append(tools, &tt.tools[i])
			}
			wantParts := tt.wantParts
			if wantParts == nil {
				wantParts = tt.parts
			}

			pair := NewBodyPair(&ai, tools)

			assert.Equal(t, tt.wantType, pair.Type)
			assert.Same(t, &ai, pair.AIMessage)
			assert.Equal(t, wantParts, ai.Parts)
			assert.Equal(t, tools, pair.ToolMessages)
			assert.Equal(t, !tt.invalid, pair.IsValid())
			assert.Equal(t, tt.wantSize, pair.Size())
		})
	}
}

func TestChainAST_NewBodyPairFromMessages_TakesAnAIMessageAndItsToolResponses(t *testing.T) {
	tests := []struct {
		name     string
		messages func() []llms.MessageContent
		wantType BodyPairType
		wantErr  string
	}{
		{
			name: "an AI message and the responses to its calls make a request-response pair",
			messages: func() []llms.MessageContent {
				return []llms.MessageContent{
					chainASTAI(chainASTCall("tool-1", "get_weather"), chainASTCall("tool-2", "get_time")),
					chainASTTool("tool-1", "get_weather", "sunny"),
					chainASTTool("tool-2", "get_time", "noon"),
				}
			},
			wantType: RequestResponse,
		},
		{
			name:     "a lone AI message is a completion",
			messages: func() []llms.MessageContent { return []llms.MessageContent{chainASTAnswer("answer")} },
			wantType: Completion,
		},
		{
			name:     "an empty slice is rejected",
			messages: func() []llms.MessageContent { return []llms.MessageContent{} },
			wantErr:  "cannot create body pair from empty message slice",
		},
		{
			name:     "a first message not from the AI is rejected",
			messages: func() []llms.MessageContent { return []llms.MessageContent{chainASTHuman("question")} },
			wantErr:  "first message in body pair must be an AI message",
		},
		{
			name: "a non-tool message after the AI message is rejected",
			messages: func() []llms.MessageContent {
				return []llms.MessageContent{chainASTAnswer("answer"), chainASTHuman("question")}
			},
			wantErr: "non-tool message found in body pair at position 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pair, err := NewBodyPairFromMessages(tt.messages())
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				assert.Nil(t, pair)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantType, pair.Type)
			assert.Equal(t, tt.messages(), pair.Messages())
			assert.True(t, pair.IsValid())
		})
	}
}

func TestChainAST_NewBodyPairFromSummarization_AnswersItsOwnSummarizationCall(t *testing.T) {
	thought := llms.TextContent{Text: "thinking", Reasoning: &reasoning.ContentReasoning{Content: "analysis"}}
	signature := &reasoning.ContentReasoning{Signature: []byte("skip_thought_signature_validator")}

	tests := []struct {
		name          string
		fakeSignature bool
		reasoningMsg  *llms.MessageContent
		wantBefore    []llms.ContentPart
		wantReasoning *reasoning.ContentReasoning
	}{
		{
			name: "the call carries no signature unless asked to",
		},
		{
			name:          "a fake signature is put on the call",
			fakeSignature: true,
			wantReasoning: signature,
		},
		{
			name:          "the parts of a reasoning message go before the call",
			fakeSignature: true,
			reasoningMsg:  &llms.MessageContent{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{thought}},
			wantBefore:    []llms.ContentPart{thought},
			wantReasoning: signature,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pair := NewBodyPairFromSummarization("summary", "call_{r:24:x}", tt.fakeSignature, tt.reasoningMsg)

			require.NotEmpty(t, pair.AIMessage.Parts)
			call, ok := pair.AIMessage.Parts[len(pair.AIMessage.Parts)-1].(llms.ToolCall)
			require.True(t, ok, "the last part must be the call")
			assert.Regexp(t, `^call_[0-9A-Za-z]{24}$`, call.ID)

			wantCall := llms.ToolCall{
				ID:   call.ID,
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      "execute_task_and_return_summary",
					Arguments: `{"question": "delegate and execute the task, then return the summary of the result"}`,
				},
				Reasoning: tt.wantReasoning,
			}
			response := chainASTTool(call.ID, "execute_task_and_return_summary", "summary")
			assert.Equal(t, Summarization, pair.Type)
			assert.Equal(t, chainASTAI(append(tt.wantBefore, wantCall)...), *pair.AIMessage)
			assert.Equal(t, []*llms.MessageContent{&response}, pair.ToolMessages)
			assert.True(t, pair.IsValid())
		})
	}
}

func TestChainAST_NewBodyPairFromCompletion_WrapsTheTextInACompletion(t *testing.T) {
	pair := NewBodyPairFromCompletion("done")

	assert.Equal(t, Completion, pair.Type)
	assert.Equal(t, chainASTAnswer("done"), *pair.AIMessage)
	assert.Nil(t, pair.ToolMessages)
	assert.Equal(t, 4, pair.Size())
}

func TestChainAST_NewChainSection_SumsHeaderAndPairSizes(t *testing.T) {
	system, human := chainASTSystem("system"), chainASTHuman("question")
	call, response := chainASTAI(chainASTCall("tool-1", "get_weather")), chainASTTool("tool-1", "get_weather", "sunny")
	header := NewHeader(&system, &human)
	completion := NewBodyPairFromCompletion("answer")
	exchange := NewBodyPair(&call, []*llms.MessageContent{&response})

	section := NewChainSection(header, []*BodyPair{completion, exchange})

	assert.Same(t, header, section.Header)
	assert.Equal(t, []*BodyPair{completion, exchange}, section.Body)
	assert.Equal(t, 69, section.Size())
}

func TestChainAST_HasToolCalls_DetectsAToolCallPart(t *testing.T) {
	withCall := chainASTAI(llms.TextContent{Text: "checking"}, chainASTCall("tool-1", "get_weather"))
	textOnly := chainASTAnswer("answer")

	assert.True(t, HasToolCalls(&withCall))
	assert.False(t, HasToolCalls(&textOnly))
	assert.False(t, HasToolCalls(nil))
}

func TestChainAST_CalculateMessageSize_CountsEveryPartType(t *testing.T) {
	tests := []struct {
		name string
		part llms.ContentPart
		want int
	}{
		{
			name: "text counts its text",
			part: llms.TextContent{Text: "Hello world"},
			want: 11,
		},
		{
			name: "an image counts its URL",
			part: llms.ImageURLContent{URL: "https://example.com/image.jpg"},
			want: 29,
		},
		{
			name: "binary data counts its bytes",
			part: llms.BinaryContent{MIMEType: "image/png", Data: []byte{0x89, 'P', 'N', 'G'}},
			want: 4,
		},
		{
			name: "a tool call counts its id, type, name and arguments",
			part: llms.ToolCall{ID: "call-1", Type: "function", FunctionCall: &llms.FunctionCall{
				Name:      "test_function",
				Arguments: `{"param1": "value1"}`,
			}},
			want: 47,
		},
		{
			name: "a tool response counts its call id, name and content",
			part: llms.ToolCallResponse{ToolCallID: "call-1", Name: "test_function", Content: "Response content"},
			want: 35,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := llms.MessageContent{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{tt.part}}
			assert.Equal(t, tt.want, CalculateMessageSize(&msg))
		})
	}
}

func TestChainAST_NormalizeToolCallIDs_ReplacesIDsTheTemplateRejects(t *testing.T) {
	const (
		openAIID   = `^call_[0-9A-Za-z]{24}$`
		validToolu = "toolu_0123456789abcdefghijklmn"
		validCall  = "call_0123456789abcdefghijklmn"
	)

	tests := []struct {
		name     string
		template string
		pattern  string // every new id must match it
		chain    []llms.MessageContent
		renamed  []string // the ids that must be replaced; every other one stays
	}{
		{
			name:     "an id with the right prefix and the wrong length is replaced",
			template: "call_{r:24:x}",
			pattern:  openAIID,
			chain: []llms.MessageContent{
				chainASTHuman("question"),
				chainASTAI(chainASTCall("call_abc", "get_weather")),
				chainASTTool("call_abc", "get_weather", "sunny"),
			},
			renamed: []string{"call_abc"},
		},
		{
			name:     "an id the template accepts is kept",
			template: "call_{r:24:x}",
			pattern:  openAIID,
			chain: []llms.MessageContent{
				chainASTHuman("question"),
				chainASTAI(chainASTCall(validCall, "get_weather")),
				chainASTTool(validCall, "get_weather", "sunny"),
			},
		},
		{
			name:     "responses that share one tool message follow their own calls",
			template: "toolu_{r:24:b}",
			pattern:  chainASTAnthropicID,
			chain: []llms.MessageContent{
				chainASTHuman("question"),
				chainASTAI(chainASTCall(validToolu, "search_weather"), chainASTCall("call_invalid", "search_news")),
				{Role: llms.ChatMessageTypeTool, Parts: []llms.ContentPart{
					llms.ToolCallResponse{ToolCallID: "call_invalid", Name: "search_news", Content: "quiet"},
					llms.ToolCallResponse{ToolCallID: validToolu, Name: "search_weather", Content: "sunny"},
				}},
			},
			renamed: []string{"call_invalid"},
		},
		{
			name:     "an id a provider reuses in every turn gets a new id in each",
			template: "toolu_{r:24:b}",
			pattern:  chainASTAnthropicID,
			chain: []llms.MessageContent{
				chainASTHuman("question"),
				chainASTAI(chainASTCall("0", "get_weather")),
				chainASTTool("0", "get_weather", "sunny"),
				chainASTHuman("follow-up"),
				chainASTAI(chainASTCall("0", "get_time")),
				chainASTTool("0", "get_time", "noon"),
			},
			renamed: []string{"0"},
		},
		{
			name:     "a call without a function after a real call keeps its id",
			template: "toolu_{r:24:b}",
			pattern:  chainASTAnthropicID,
			chain: []llms.MessageContent{
				chainASTHuman("question"),
				chainASTAI(chainASTCall("call_abc", "get_weather"), llms.ToolCall{ID: "call_partial", Type: "function"}),
				chainASTTool("call_abc", "get_weather", "sunny"),
			},
			renamed: []string{"call_abc"},
		},
		{
			name:     "every rejected id across three sections is replaced",
			template: "toolu_{r:24:b}",
			pattern:  chainASTAnthropicID,
			chain:    providerSwitchChain(true),
			renamed:  []string{"call_nmap01", "call_ssh02", "call_sum03"},
		},
		{
			name:     "an empty chain is left empty",
			template: "toolu_{r:24:b}",
			pattern:  chainASTAnthropicID,
			chain:    emptyChain(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := chainASTClone(tt.chain)
			ast, err := NewChainAST(tt.chain, false)
			require.NoError(t, err)

			require.NoError(t, ast.NormalizeToolCallIDs(tt.template))

			restored, renames := chainASTUndoRename(t, before, ast.Messages())
			assert.Equal(t, before, restored, "only the ids of calls and of their responses may change")
			assert.ElementsMatch(t, tt.renamed, slices.Collect(maps.Keys(renames)))
			for _, id := range renames {
				assert.Regexp(t, tt.pattern, id)
			}

			ids := map[string]bool{}
			calls := 0
			for _, msg := range ast.Messages() {
				for _, part := range msg.Parts {
					if call, ok := part.(llms.ToolCall); ok {
						calls++
						ids[call.ID] = true
					}
				}
			}
			assert.Len(t, ids, calls, "every call must keep an id of its own")

			_, err = NewChainAST(ast.Messages(), false)
			assert.NoError(t, err, "the normalized chain must parse without force")
		})
	}
}

func TestChainAST_NormalizeToolCallIDs_SkipsAPairAndAToolMessageLeftNil(t *testing.T) {
	ast, err := NewChainAST([]llms.MessageContent{
		chainASTHuman("question"),
		chainASTAI(chainASTCall("call_abc", "get_weather")),
		chainASTTool("call_abc", "get_weather", "sunny"),
	}, false)
	require.NoError(t, err)
	pair := ast.Sections[0].Body[0]
	pair.ToolMessages = append([]*llms.MessageContent{nil}, pair.ToolMessages...)
	ast.Sections[0].Body = []*BodyPair{{Type: RequestResponse}, pair}

	require.NoError(t, ast.NormalizeToolCallIDs("toolu_{r:24:b}"))

	call, ok := pair.AIMessage.Parts[0].(llms.ToolCall)
	require.True(t, ok)
	assert.Regexp(t, chainASTAnthropicID, call.ID)
	assert.Equal(t, chainASTTool(call.ID, "get_weather", "sunny"), *pair.ToolMessages[1])
}

func TestChainAST_NormalizeToolCallIDs_FollowsAForcedRepairThroughAProviderSwitch(t *testing.T) {
	const banner, stray = 8, 10 // in providerSwitchChain: the AI message of the banner call, and the slot after its tool message
	chain := func(withReasoning bool) []llms.MessageContent {
		return slices.Insert(providerSwitchChain(withReasoning), stray, chainASTTool("call_orphan04", "terminal", "late output"))
	}
	repaired := func(withReasoning bool, args string) []llms.MessageContent {
		repaired := chain(withReasoning)
		repaired[banner].Parts = append(repaired[banner].Parts,
			llms.ToolCall{ID: "call_orphan04", FunctionCall: &llms.FunctionCall{Name: "terminal", Arguments: "{}"}})
		return append(repaired,
			chainASTAI(chainASTCallWith("call_ls05", "terminal", args)),
			chainASTTool("call_ls05", "terminal", chainASTFallback),
		)
	}
	renamed := []string{"call_nmap01", "call_ssh02", "call_orphan04", "call_sum03", "call_ls05"}

	ast, err := NewChainAST(append(chain(true), chainASTAI(chainASTCallWith("call_ls05", "terminal", "{"))), true)
	require.NoError(t, err)
	assert.Equal(t, repaired(true, "{"), ast.Messages())

	steps := []struct {
		name    string
		apply   func() error
		want    []llms.MessageContent
		renamed []string
	}{
		{
			name:  "a continuation opens a section",
			apply: func() error { ast.AppendHumanMessage("continue"); return nil },
			want:  append(repaired(true, "{"), chainASTHuman("continue")),
		},
		{
			name:    "the added call and the fallback are renamed with the rest",
			apply:   func() error { return ast.NormalizeToolCallIDs("toolu_{r:24:b}") },
			want:    append(repaired(true, "{"), chainASTHuman("continue")),
			renamed: renamed,
		},
		{
			name:    "the truncated arguments become an empty object",
			apply:   func() error { ast.SanitizeToolCallArguments(); return nil },
			want:    append(repaired(true, "{}"), chainASTHuman("continue")),
			renamed: renamed,
		},
		{
			name:    "the reasoning is cleared",
			apply:   ast.ClearReasoning,
			want:    append(repaired(false, "{}"), chainASTHuman("continue")),
			renamed: renamed,
		},
	}

	var ids map[string]string
	for _, step := range steps {
		require.NoError(t, step.apply(), step.name)
		restored, renames := chainASTUndoRename(t, step.want, ast.Messages())
		assert.Equal(t, step.want, restored, step.name)
		assert.ElementsMatch(t, step.renamed, slices.Collect(maps.Keys(renames)), step.name)
		if ids != nil {
			assert.Equal(t, ids, renames, "%s: the new ids must stay", step.name)
		} else if len(renames) > 0 {
			ids = renames
		}
	}
	for _, id := range ids {
		assert.Regexp(t, chainASTAnthropicID, id)
	}

	reparsed, err := NewChainAST(ast.Messages(), false)
	require.NoError(t, err, "the chain must parse without force")
	assertChainASTSizes(t, [][]int{{335, 43, 275, 17}, {275, 22, 20, 233}, {397, 23, 246, 128}, {8, 8}}, reparsed)
}

func TestChainAST_ClearReasoning_StripsReasoningAndKeepsContent(t *testing.T) {
	thought := func() *reasoning.ContentReasoning {
		return &reasoning.ContentReasoning{Content: "thinking", Signature: []byte("signature")}
	}
	call := chainASTCall("tool-1", "get_weather")
	call.Reasoning = thought()

	tests := []struct {
		name  string
		chain []llms.MessageContent
		want  []llms.MessageContent
	}{
		{
			name: "reasoning is stripped from the header, the text and the call",
			chain: []llms.MessageContent{
				{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextContent{Text: "system", Reasoning: thought()}}},
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "question", Reasoning: thought()}}},
				chainASTAI(llms.TextContent{Text: "answer", Reasoning: thought()}, call),
				chainASTTool("tool-1", "get_weather", "sunny"),
			},
			want: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTHuman("question"),
				chainASTAI(llms.TextContent{Text: "answer"}, chainASTCall("tool-1", "get_weather")),
				chainASTTool("tool-1", "get_weather", "sunny"),
			},
		},
		{
			name:  "reasoning is stripped across three sections, the note of a tool message included",
			chain: providerSwitchChain(true),
			want:  providerSwitchChain(false),
		},
		{
			name:  "an empty chain is left empty",
			chain: emptyChain(),
			want:  []llms.MessageContent{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := NewChainAST(tt.chain, false)
			require.NoError(t, err)

			require.NoError(t, ast.ClearReasoning())

			assert.Equal(t, tt.want, ast.Messages())
			_, err = NewChainAST(ast.Messages(), false)
			assert.NoError(t, err, "the cleared chain must parse without force")
		})
	}
}

func TestChainAST_ClearReasoning_SkipsASectionLeftWithoutAHeader(t *testing.T) {
	ast, err := NewChainAST([]llms.MessageContent{
		chainASTHuman("question"),
		chainASTAI(llms.TextContent{Text: "answer", Reasoning: &reasoning.ContentReasoning{Content: "thinking"}}),
	}, false)
	require.NoError(t, err)
	ast.Sections = append([]*ChainSection{{}}, ast.Sections...)

	require.NoError(t, ast.ClearReasoning())

	assert.Equal(t, []llms.MessageContent{chainASTHuman("question"), chainASTAnswer("answer")}, ast.Sections[1].Messages())
}

func TestChainAST_ClearReasoning_LeavesNormalizeToolCallIDsTheSameIDsToReplace(t *testing.T) {
	ast, err := NewChainAST(providerSwitchChain(true), false)
	require.NoError(t, err)

	require.NoError(t, ast.ClearReasoning())
	require.NoError(t, ast.NormalizeToolCallIDs("toolu_{r:24:b}"))

	restored, renames := chainASTUndoRename(t, providerSwitchChain(false), ast.Messages())
	assert.Equal(t, providerSwitchChain(false), restored)
	assert.ElementsMatch(t, []string{"call_nmap01", "call_ssh02", "call_sum03"}, slices.Collect(maps.Keys(renames)))
	for _, id := range renames {
		assert.Regexp(t, chainASTAnthropicID, id)
	}
	_, err = NewChainAST(ast.Messages(), false)
	assert.NoError(t, err, "the chain must parse without force")
}

func TestChainAST_ContainsToolCallReasoning_LooksOnlyAtToolCalls(t *testing.T) {
	signed := &reasoning.ContentReasoning{Signature: []byte("signature")}
	signedCall := chainASTCall("tool-1", "get_weather")
	signedCall.Reasoning = signed

	tests := []struct {
		name     string
		messages []llms.MessageContent
		want     bool
	}{
		{
			name:     "an empty slice has none",
			messages: []llms.MessageContent{},
		},
		{
			name: "tool calls without reasoning have none",
			messages: []llms.MessageContent{
				chainASTHuman("question"),
				chainASTAI(llms.TextContent{Text: "answer"}, chainASTCall("tool-1", "get_weather")),
			},
		},
		{
			name: "reasoning only on text does not count",
			messages: []llms.MessageContent{
				chainASTAI(llms.TextContent{Text: "thinking", Reasoning: signed}, llms.TextContent{Text: "answer"}),
			},
		},
		{
			name: "reasoning on a call in a later message is found",
			messages: []llms.MessageContent{
				chainASTHuman("question"),
				chainASTAnswer("answer"),
				chainASTHuman("follow-up"),
				chainASTAI(signedCall),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ContainsToolCallReasoning(tt.messages))
		})
	}
}

func TestChainAST_ExtractReasoningMessage_ReturnsTheFirstReasoningPart(t *testing.T) {
	first := llms.TextContent{Text: "first", Reasoning: &reasoning.ContentReasoning{Content: "first analysis"}}
	second := llms.TextContent{Text: "second", Reasoning: &reasoning.ContentReasoning{Content: "second analysis"}}
	signedOnly := llms.TextContent{Text: "answer", Reasoning: &reasoning.ContentReasoning{Signature: []byte("signature")}}

	tests := []struct {
		name     string
		messages []llms.MessageContent
		want     *llms.MessageContent
	}{
		{
			name:     "an empty slice has none",
			messages: []llms.MessageContent{},
		},
		{
			name: "a chain without reasoning has none",
			messages: []llms.MessageContent{
				chainASTHuman("question"),
				chainASTAI(llms.TextContent{Text: "answer"}, chainASTCall("tool-1", "get_weather")),
			},
		},
		{
			name: "reasoning left empty does not count",
			messages: []llms.MessageContent{
				chainASTAI(llms.TextContent{Text: "answer", Reasoning: &reasoning.ContentReasoning{}}),
			},
		},
		{
			name:     "a signature without content counts as reasoning",
			messages: []llms.MessageContent{chainASTAI(signedOnly)},
			want:     &llms.MessageContent{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{signedOnly}},
		},
		{
			name: "the first reasoning part of an AI message is returned alone",
			messages: []llms.MessageContent{
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{
					llms.TextContent{Text: "question", Reasoning: &reasoning.ContentReasoning{Content: "human analysis"}},
				}},
				chainASTAI(llms.TextContent{Text: "plain"}, first, llms.TextContent{Text: "trailing"}),
				chainASTAI(second),
			},
			want: &llms.MessageContent{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{first}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ExtractReasoningMessage(tt.messages))
		})
	}
}

func TestChainAST_SanitizeJSONControlChars_EscapesControlCharsInsideStrings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid JSON with an escaped newline is returned unchanged",
			input: `{"input": "line1\nline2"}`,
			want:  `{"input": "line1\nline2"}`,
		},
		{
			name:  "a literal carriage return inside a string is escaped",
			input: "{\"cmd\": \"echo\rtest\"}",
			want:  `{"cmd": "echo\rtest"}`,
		},
		{
			name:  "a literal tab inside a string is escaped",
			input: "{\"v\": \"a\tb\"}",
			want:  `{"v": "a\tb"}`,
		},
		{
			name:  "a literal backspace and form feed inside a string are escaped",
			input: "{\"v\": \"a\x08\x0Cb\"}",
			want:  `{"v": "a\b\fb"}`,
		},
		{
			name:  "another control character gets a unicode escape",
			input: "{\"v\": \"a\x01b\"}",
			want:  `{"v": "a\u0001b"}`,
		},
		{
			name:  "a control character between tokens is left alone",
			input: "{\n\"k\": \"a\nb\"}",
			want:  "{\n\"k\": \"a\\nb\"}",
		},
		{
			name:  "a string at the top level is escaped too",
			input: "\"ls -la\n\"",
			want:  `"ls -la\n"`,
		},
		{
			name:  "an escaped quote does not end the string",
			input: "{\"v\": \"a \\\"quote\nb\"}",
			want:  `{"v": "a \"quote\nb"}`,
		},
		{
			name:  "an escaped backslash before a quote ends the string",
			input: "{\"v\": \"a\\\\\", \"w\": \"x\ny\"}",
			want:  `{"v": "a\\", "w": "x\ny"}`,
		},
		{
			name:  "a newline in the first of two fields, as vLLM received it, is escaped",
			input: "{\"input\": \"index.html] =\ncurl -s http://example.com\", \"cwd\": \"/work\"}",
			want:  `{"input": "index.html] =\ncurl -s http://example.com", "cwd": "/work"}`,
		},
		{
			name:  "invalid JSON without control characters is returned unchanged",
			input: `{broken`,
			want:  `{broken`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SanitizeJSONControlChars(tt.input))
		})
	}
}

func TestChainAST_SanitizeToolCallArguments_LeavesEveryCallWithValidJSON(t *testing.T) {
	terminal := func(id, args string) llms.ToolCall { return chainASTCallWith(id, "terminal", args) }
	exchange := func(parts ...llms.ContentPart) []llms.MessageContent {
		chain := []llms.MessageContent{chainASTHuman("run it"), chainASTAI(parts...)}
		for _, part := range parts {
			if call, ok := part.(llms.ToolCall); ok && call.FunctionCall != nil {
				chain = append(chain, chainASTTool(call.ID, "terminal", "ok"))
			}
		}
		return chain
	}

	tests := []struct {
		name  string
		chain []llms.MessageContent
		want  []llms.MessageContent
	}{
		{
			name:  "a valid object is left unchanged",
			chain: exchange(terminal("call_1", `{"input": "ls -la", "cwd": "/work"}`)),
			want:  exchange(terminal("call_1", `{"input": "ls -la", "cwd": "/work"}`)),
		},
		{
			name:  "literal control characters are escaped",
			chain: exchange(terminal("call_1", "{\"input\": \"a\nb\rc\", \"cwd\": \"/work\"}")),
			want:  exchange(terminal("call_1", `{"input": "a\nb\rc", "cwd": "/work"}`)),
		},
		{
			name:  "a truncated object is replaced with an empty one",
			chain: exchange(terminal("call_1", `{`)),
			want:  exchange(terminal("call_1", `{}`)),
		},
		{
			name: "a call without a function is skipped and the broken call after it repaired",
			chain: exchange(
				terminal("call_1", `{"input": "ls"}`),
				llms.ToolCall{ID: "call_nil", Type: "function"},
				terminal("call_2", "{\"input\": \"a\nb\"}"),
			),
			want: exchange(
				terminal("call_1", `{"input": "ls"}`),
				llms.ToolCall{ID: "call_nil", Type: "function"},
				terminal("call_2", `{"input": "a\nb"}`),
			),
		},
		{
			name: "a call after a completion and a summarization call in a later section are repaired",
			chain: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTHuman("scan"),
				chainASTAnswer("starting"),
				chainASTAI(terminal("call_1", "{\"input\": \"a\nb\"}")),
				chainASTTool("call_1", "terminal", "ok"),
				chainASTHuman("summarize"),
				chainASTAI(chainASTCallWith("call_2", "execute_task_and_return_summary", `{"question": "trunc`)),
				chainASTTool("call_2", "execute_task_and_return_summary", "summary"),
			},
			want: []llms.MessageContent{
				chainASTSystem("system"),
				chainASTHuman("scan"),
				chainASTAnswer("starting"),
				chainASTAI(terminal("call_1", `{"input": "a\nb"}`)),
				chainASTTool("call_1", "terminal", "ok"),
				chainASTHuman("summarize"),
				chainASTAI(chainASTCallWith("call_2", "execute_task_and_return_summary", `{}`)),
				chainASTTool("call_2", "execute_task_and_return_summary", "summary"),
			},
		},
		{
			name:  "an empty chain is left empty",
			chain: emptyChain(),
			want:  []llms.MessageContent{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := NewChainAST(tt.chain, false)
			require.NoError(t, err)

			ast.SanitizeToolCallArguments()

			assert.Equal(t, tt.want, ast.Messages())
		})
	}
}

func TestChainAST_SanitizeToolCallArguments_SkipsAPairLeftWithoutItsAIMessage(t *testing.T) {
	ast, err := NewChainAST([]llms.MessageContent{
		chainASTHuman("run it"),
		chainASTAI(chainASTCallWith("call_1", "terminal", "{")),
		chainASTTool("call_1", "terminal", "ok"),
	}, false)
	require.NoError(t, err)
	ast.Sections[0].Body = append([]*BodyPair{{Type: RequestResponse}}, ast.Sections[0].Body...)

	ast.SanitizeToolCallArguments()

	assert.Equal(t, []llms.MessageContent{
		chainASTHuman("run it"),
		chainASTAI(chainASTCallWith("call_1", "terminal", "{}")),
		chainASTTool("call_1", "terminal", "ok"),
	}, ast.Messages())
}
