package csum

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"

	"pentagi/pkg/cast"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

const (
	chainSummaryReply = "summary text"
	// chainSummaryNew stands, in an expected body, for a summary pair the summarizer wrote.
	chainSummaryNew = -1
)

// chainSummaryRecorder records what it is asked to summarize; with fail set it fails the way a provider call does,
// with numbered set the n-th call answers SUMMARY-n, n zero-padded to three digits.
type chainSummaryRecorder struct {
	fail     bool
	numbered bool
	mx       sync.Mutex
	texts    []string
}

func (r *chainSummaryRecorder) handle(ctx context.Context, text string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	r.mx.Lock()
	defer r.mx.Unlock()
	r.texts = append(r.texts, text)
	if r.fail {
		return "", errors.New("provider unavailable")
	}
	if r.numbered {
		return fmt.Sprintf("SUMMARY-%03d", len(r.texts)), nil
	}
	return chainSummaryReply, nil
}

func (r *chainSummaryRecorder) sent() []string {
	r.mx.Lock()
	defer r.mx.Unlock()
	return slices.Clone(r.texts)
}

// chainSummaryAssertSent checks that every want string is in one of prompts and no notWant string is in any.
func chainSummaryAssertSent(t *testing.T, prompts, want, notWant []string) {
	t.Helper()
	all := strings.Join(prompts, "\n")
	for _, s := range want {
		assert.Contains(t, all, s, "text sent to the summarizer")
	}
	for _, s := range notWant {
		assert.NotContains(t, all, s, "text sent to the summarizer")
	}
}

func chainSummaryMsg(role llms.ChatMessageType, text string) *llms.MessageContent {
	return &llms.MessageContent{Role: role, Parts: []llms.ContentPart{llms.TextContent{Text: text}}}
}

func chainSummarySection(human string, body ...*cast.BodyPair) *cast.ChainSection {
	return cast.NewChainSection(cast.NewHeader(nil, chainSummaryMsg(llms.ChatMessageTypeHuman, human)), body)
}

// chainSummaryQA builds a section answering human with one Completion pair per answer.
func chainSummaryQA(human string, answers ...string) *cast.ChainSection {
	body := make([]*cast.BodyPair, 0, len(answers))
	for _, answer := range answers {
		body = append(body, cast.NewBodyPairFromCompletion(answer))
	}
	return chainSummarySection(human, body...)
}

// chainSummaryFirst gives section the system message the first section of a chain carries.
func chainSummaryFirst(section *cast.ChainSection) *cast.ChainSection {
	system := chainSummaryMsg(llms.ChatMessageTypeSystem, "System message")
	return cast.NewChainSection(cast.NewHeader(system, section.Header.HumanMessage), section.Body)
}

// chainSummaryToolPair builds a RequestResponse pair: the lead parts, then a call of name that response answers.
func chainSummaryToolPair(
	id, name, response string, signature *reasoning.ContentReasoning, lead ...llms.ContentPart,
) *cast.BodyPair {
	call := llms.ToolCall{
		ID:           id,
		Type:         "function",
		FunctionCall: &llms.FunctionCall{Name: name, Arguments: `{"query": "test"}`},
		Reasoning:    signature,
	}
	ai := &llms.MessageContent{Role: llms.ChatMessageTypeAI, Parts: append(slices.Clone(lead), call)}
	tool := &llms.MessageContent{Role: llms.ChatMessageTypeTool, Parts: []llms.ContentPart{
		llms.ToolCallResponse{ToolCallID: id, Name: name, Content: response},
	}}
	return cast.NewBodyPair(ai, []*llms.MessageContent{tool})
}

func chainSummarySignature(signature string) *reasoning.ContentReasoning {
	return &reasoning.ContentReasoning{Signature: []byte(signature)}
}

// chainSummaryThought is a text part carrying reasoning, the way Kimi places it before a tool call.
func chainSummaryThought(text, thought string) llms.ContentPart {
	return llms.TextContent{Text: text, Reasoning: &reasoning.ContentReasoning{Content: thought}}
}

// chainSummarySignedThought is an Anthropic thinking block: no text, signed reasoning, placed before the tool call.
func chainSummarySignedThought(thought, signature string) llms.ContentPart {
	return llms.TextContent{Reasoning: &reasoning.ContentReasoning{Content: thought, Signature: []byte(signature)}}
}

// chainSummaryPad returns prefix padded with x to size bytes.
func chainSummaryPad(prefix string, size int) string {
	return prefix + strings.Repeat("x", size-len(prefix))
}

func chainSummaryJSON(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return string(data)
}

func chainSummaryPairsJSON(t *testing.T, body []*cast.BodyPair) []string {
	t.Helper()
	out := make([]string, 0, len(body))
	for _, pair := range body {
		out = append(out, chainSummaryJSON(t, pair))
	}
	return out
}

// chainSummaryCheckSummary asserts pair summarizes chainSummaryReply as kind, after lead, signed only when signed is set.
func chainSummaryCheckSummary(
	t *testing.T, pair *cast.BodyPair, kind cast.BodyPairType, signed bool, lead ...llms.ContentPart,
) {
	t.Helper()
	chainSummaryCheckSummaryText(t, pair, chainSummaryReply, kind, signed, lead...)
}

func chainSummaryCheckSummaryText(
	t *testing.T, pair *cast.BodyPair, text string, kind cast.BodyPairType, signed bool, lead ...llms.ContentPart,
) {
	t.Helper()
	require.Equal(t, kind, pair.Type, "summary pair type")
	if kind == cast.Completion {
		want := []llms.ContentPart{llms.TextContent{Text: "**summarized content:**\n" + text}}
		assert.Equal(t, want, pair.AIMessage.Parts)
		return
	}
	parts := pair.AIMessage.Parts
	require.Len(t, parts, len(lead)+1, "summary parts")
	if len(lead) > 0 {
		assert.Equal(t, lead, parts[:len(lead)], "parts carried before the summary call")
	}
	call, ok := parts[len(lead)].(llms.ToolCall)
	require.True(t, ok, "the summary ends with a tool call")
	assert.Equal(t, "execute_task_and_return_summary", call.FunctionCall.Name)
	var signature *reasoning.ContentReasoning
	if signed {
		signature = chainSummarySignature("skip_thought_signature_validator")
	}
	assert.Equal(t, signature, call.Reasoning, "summary call signature")
	require.Len(t, pair.ToolMessages, 1)
	response := llms.ToolCallResponse{ToolCallID: call.ID, Name: "execute_task_and_return_summary", Content: text}
	assert.Equal(t, []llms.ContentPart{response}, pair.ToolMessages[0].Parts)
}

// chainSummaryCheckBody asserts body against want: chainSummaryNew is a summary, any other entry an original pair kept as is.
func chainSummaryCheckBody(
	t *testing.T, body []*cast.BodyPair, original []string, want []int,
	kind cast.BodyPairType, signed bool, lead ...llms.ContentPart,
) {
	t.Helper()
	require.Len(t, body, len(want), "body pairs")
	for i, w := range want {
		if w == chainSummaryNew {
			chainSummaryCheckSummary(t, body[i], kind, signed, lead...)
			continue
		}
		assert.Equal(t, original[w], chainSummaryJSON(t, body[i]), "pair %d should be original pair %d", i, w)
	}
}

// chainSummaryVerifyAST asserts ast is well formed: headers, answered tool calls, sizes, and a lossless message round trip.
func chainSummaryVerifyAST(t *testing.T, ast *cast.ChainAST) {
	t.Helper()
	total := 0
	for i, section := range ast.Sections {
		if i == 0 {
			assert.False(t, section.Header.SystemMessage == nil && section.Header.HumanMessage == nil, "section 0 header is empty")
		} else {
			assert.Nil(t, section.Header.SystemMessage, "section %d carries a system message", i)
			assert.NotNil(t, section.Header.HumanMessage, "section %d has no human message", i)
		}
		if i < len(ast.Sections)-1 {
			assert.NotEmpty(t, section.Body, "section %d has no body", i)
		}
		size := section.Header.Size()
		for j, pair := range section.Body {
			size += pair.Size()
			require.NotNil(t, pair.AIMessage, "section %d pair %d has no AI message", i, j)
			calls, responses := 0, 0
			for _, part := range pair.AIMessage.Parts {
				if _, ok := part.(llms.ToolCall); ok {
					calls++
				}
			}
			for _, msg := range pair.ToolMessages {
				for _, part := range msg.Parts {
					if _, ok := part.(llms.ToolCallResponse); ok {
						responses++
					}
				}
			}
			switch pair.Type {
			case cast.RequestResponse, cast.Summarization:
				assert.Equal(t, calls, responses, "section %d pair %d: tool calls and responses differ", i, j)
			case cast.Completion:
				assert.Zero(t, calls+len(pair.ToolMessages), "section %d pair %d: completion with tool traffic", i, j)
			default:
				t.Errorf("section %d pair %d: unexpected pair type %d", i, j, pair.Type)
			}
		}
		assert.Equal(t, size, section.Size(), "section %d size", i)
		total += size
	}
	assert.Equal(t, total, ast.Size(), "chain size")

	messages := ast.Messages()
	again, err := cast.NewChainAST(messages, false)
	require.NoError(t, err)
	assert.Equal(t, chainSummaryJSON(t, messages), chainSummaryJSON(t, again.Messages()), "round trip through messages")
}

// chainSummaryIsSummary reports whether pair is a summary: a summarization call, or a completion under the summary prefix.
func chainSummaryIsSummary(pair *cast.BodyPair) bool {
	if pair.Type == cast.Summarization {
		return true
	}
	if pair.Type != cast.Completion || len(pair.AIMessage.Parts) == 0 {
		return false
	}
	text, ok := pair.AIMessage.Parts[0].(llms.TextContent)
	return ok && strings.HasPrefix(text.Text, "**summarized content:**\n")
}

// chainSummaryPromptData returns the tasks/messages portion of a summarizer prompt after its single instructions block.
func chainSummaryPromptData(t *testing.T, sent string) string {
	t.Helper()
	const sep = "</instructions>\n\n"
	assert.Equal(t, 1, strings.Count(sent, "<instructions>"), "instruction blocks")
	idx := strings.Index(sent, sep)
	require.GreaterOrEqual(t, idx, 0, "the prompt opens with an instructions block")
	return sent[idx+len(sep):]
}

type chainSummaryCheck func(t *testing.T, got, before *cast.ChainAST)

type chainSummaryStep struct {
	grow    func(ast *cast.ChainAST) // nil summarizes the chain as it stands
	calls   int
	sent    []string
	notSent []string
	prompts []string // when set, the data of every prompt of the step in call order
	checks  []chainSummaryCheck
}

// chainSummaryAddAnswers appends count answers of size bytes to the last section.
func chainSummaryAddAnswers(count, size int) func(*cast.ChainAST) {
	return func(ast *cast.ChainAST) {
		last := ast.Sections[len(ast.Sections)-1]
		for i := range count {
			last.AddBodyPair(cast.NewBodyPairFromCompletion(fmt.Sprintf("Response %d: %s", i, strings.Repeat("A", size))))
		}
	}
}

func chainSummaryAddPairs(pairs ...*cast.BodyPair) func(*cast.ChainAST) {
	return func(ast *cast.ChainAST) {
		last := ast.Sections[len(ast.Sections)-1]
		for _, pair := range pairs {
			last.AddBodyPair(pair)
		}
	}
}

// chainSummaryAddQA appends a section answering human with one answer of size bytes past its label.
func chainSummaryAddQA(human string, size int) func(*cast.ChainAST) {
	return func(ast *cast.ChainAST) {
		ast.AddSection(chainSummaryQA(human, "Answer to "+human+": "+strings.Repeat("B", size)))
	}
}

// chainSummaryHistory builds n sections: the older ones already summarized, the newest one answered.
func chainSummaryHistory(n int) []*cast.ChainSection {
	sections := make([]*cast.ChainSection, 0, n)
	for i := range n - 1 {
		sections = append(sections, chainSummaryQA(fmt.Sprintf("Question %02d", i),
			fmt.Sprintf("**summarized content:**\nSummary %02d", i)))
	}
	sections = append(sections, chainSummaryQA(fmt.Sprintf("Question %02d", n-1), fmt.Sprintf("Answer %02d", n-1)))
	sections[0] = chainSummaryFirst(sections[0])
	return sections
}

// chainSummaryBudgetChain builds a 65536-byte chain of three summarized sections before a newest one of lastSize bytes.
func chainSummaryBudgetChain(lastSize int) []*cast.ChainSection {
	return []*cast.ChainSection{
		chainSummaryFirst(chainSummaryQA("Question 0", chainSummaryPad("**summarized content:**\nSummary 0 ", 976))),
		chainSummaryQA("Question 1", chainSummaryPad("**summarized content:**\nSummary 1 ", 990)),
		chainSummaryQA("Question 2", chainSummaryPad("**summarized content:**\nSummary 2 ", 31758)),
		chainSummaryQA("Question 3", chainSummaryPad("Answer 3 ", lastSize-10)),
	}
}

func chainSummaryUnchanged(t *testing.T, got, before *cast.ChainAST) {
	t.Helper()
	assert.Equal(t, chainSummaryJSON(t, before.Messages()), chainSummaryJSON(t, got.Messages()), "the chain should be unchanged")
}

func chainSummarySections(n int) chainSummaryCheck {
	return func(t *testing.T, got, _ *cast.ChainAST) {
		t.Helper()
		assert.Len(t, got.Sections, n, "sections")
	}
}

// chainSummaryInputSize guards a row whose point is a byte boundary: the chain it summarizes has size n.
func chainSummaryInputSize(n int) chainSummaryCheck {
	return func(t *testing.T, _, before *cast.ChainAST) {
		t.Helper()
		require.Equal(t, n, before.Size(), "input chain size")
	}
}

// chainSummaryBodyAt checks the body of section i, counted from the end when negative, against the same section before.
func chainSummaryBodyAt(
	i int, want []int, kind cast.BodyPairType, signed bool, lead ...llms.ContentPart,
) chainSummaryCheck {
	return func(t *testing.T, got, before *cast.ChainAST) {
		t.Helper()
		g, b := i, i
		if i < 0 {
			g, b = len(got.Sections)+i, len(before.Sections)+i
		}
		require.Less(t, g, len(got.Sections), "section %d", i)
		chainSummaryCheckBody(t, got.Sections[g].Body, chainSummaryPairsJSON(t, before.Sections[b].Body), want, kind, signed, lead...)
	}
}

func chainSummaryHeaderKept(i int) chainSummaryCheck {
	return func(t *testing.T, got, before *cast.ChainAST) {
		t.Helper()
		assert.Equal(t, chainSummaryJSON(t, before.Sections[i].Header.Messages()), chainSummaryJSON(t, got.Sections[i].Header.Messages()),
			"header of section %d", i)
	}
}

// chainSummaryKeptTail checks the n newest sections are the n newest sections before, as they were.
func chainSummaryKeptTail(n int) chainSummaryCheck {
	return func(t *testing.T, got, before *cast.ChainAST) {
		t.Helper()
		require.GreaterOrEqual(t, len(got.Sections), n)
		for k := range n {
			assert.Equal(t, chainSummaryJSON(t, before.Sections[len(before.Sections)-n+k]),
				chainSummaryJSON(t, got.Sections[len(got.Sections)-n+k]), "kept section %d from the end", n-k)
		}
	}
}

// chainSummaryQAHead checks the first section is the QA summary: the system message, questions as its human parts,
// and one summary of kind reading text.
func chainSummaryQAHead(kind cast.BodyPairType, text string, questions ...string) chainSummaryCheck {
	return func(t *testing.T, got, _ *cast.ChainAST) {
		t.Helper()
		require.NotEmpty(t, got.Sections)
		head := got.Sections[0]
		assert.Equal(t, chainSummaryMsg(llms.ChatMessageTypeSystem, "System message"), head.Header.SystemMessage)
		parts := make([]llms.ContentPart, 0, len(questions))
		for _, q := range questions {
			parts = append(parts, llms.TextContent{Text: q})
		}
		require.NotNil(t, head.Header.HumanMessage)
		assert.Equal(t, parts, head.Header.HumanMessage.Parts, "questions of the QA summary")
		require.Len(t, head.Body, 1)
		chainSummaryCheckSummaryText(t, head.Body[0], text, kind, false)
	}
}

func TestChainSummary_SummarizeChain_AppliesTheConfiguredStrategies(t *testing.T) {
	qaCount := SummarizerConfig{UseQA: true, MaxQASections: 2, MaxQABytes: 10000, MaxBPBytes: 1000}
	combined := SummarizerConfig{
		PreserveLast: true, LastSecBytes: 500, UseQA: true, MaxQASections: 2, MaxQABytes: 10000, MaxBPBytes: 1000,
	}
	flow := SummarizerConfig{
		PreserveLast: true, UseQA: true, LastSecBytes: 51200, MaxBPBytes: 16384,
		MaxQASections: 10, MaxQABytes: 65536, KeepQASections: 1,
	}
	summaryResponse := "<tool_call_response name=\"execute_task_and_return_summary\" part=\"0\">\nsummary text\n</tool_call_response>"
	terminal := func(n int) *cast.BodyPair {
		return chainSummaryToolPair(fmt.Sprintf("call_%d", n), "terminal", fmt.Sprintf("out %d ", n)+strings.Repeat("r", 640), nil)
	}
	twoTasks := "<tasks>\n<task id=\"0\">\nQuestion 1\n</task>\n<task id=\"1\">\nQuestion 2\n</task>\n</tasks>\n"

	tests := []struct {
		name     string
		cfg      SummarizerConfig
		numbered bool
		initial  []*cast.ChainSection
		steps    []chainSummaryStep
		stable   bool // after every step a second pass calls nothing and returns the same chain
	}{
		{
			name:   "an empty chain is returned as it is",
			cfg:    flow,
			steps:  []chainSummaryStep{{checks: []chainSummaryCheck{chainSummaryUnchanged}}},
			stable: true,
		},
		{
			name: "a kept section before the last is rotated too",
			cfg:  SummarizerConfig{PreserveLast: true, LastSecBytes: 500, MaxBPBytes: 1000, KeepQASections: 2},
			initial: []*cast.ChainSection{
				chainSummaryFirst(chainSummaryQA("Question 1", slices.Repeat([]string{strings.Repeat("A", 200)}, 5)...)),
				chainSummaryQA("Question 2", "Answer 2"),
			},
			steps: []chainSummaryStep{{
				calls:   1,
				sent:    []string{"Question 1", "AAA"},
				notSent: []string{"Question 2", "Answer 2"},
				checks: []chainSummaryCheck{
					chainSummarySections(2), chainSummaryHeaderKept(0),
					chainSummaryBodyAt(0, []int{chainSummaryNew, 4}, cast.Completion, false), chainSummaryKeptTail(1),
				},
			}},
			stable: true,
		},
		{
			name: "the oldest sections collapse once there are too many",
			cfg:  qaCount,
			initial: []*cast.ChainSection{
				chainSummaryFirst(chainSummaryQA("Question 1", "Answer 1")),
				chainSummaryQA("Question 2", "Answer 2"),
			},
			steps: []chainSummaryStep{
				{
					grow:    chainSummaryAddQA("Question 3", 100),
					calls:   2, // each older section becomes a summary, and a lone oldest summary is not collapsed
					sent:    []string{"Answer 1", "Answer 2"},
					notSent: []string{"Answer to Question 3"},
					checks: []chainSummaryCheck{
						chainSummarySections(3), chainSummaryBodyAt(0, []int{chainSummaryNew}, cast.Completion, false),
						chainSummaryBodyAt(1, []int{chainSummaryNew}, cast.Completion, false), chainSummaryKeptTail(1),
					},
				},
				{
					grow:    chainSummaryAddQA("Question 4", 100),
					calls:   2,
					sent:    []string{"Answer to Question 3", "**summarized content:**\nsummary text"},
					notSent: []string{"Answer to Question 4"},
					checks: []chainSummaryCheck{
						chainSummarySections(3), chainSummaryQAHead(cast.Completion, chainSummaryReply, "Question 1", "Question 2"),
						chainSummaryBodyAt(1, []int{chainSummaryNew}, cast.Completion, false), chainSummaryKeptTail(1),
					},
				},
				{
					grow:    chainSummaryAddQA("Question 5", 100),
					calls:   2,
					sent:    []string{"Answer to Question 4", "Question 1"},
					notSent: []string{"Answer to Question 5"},
					checks: []chainSummaryCheck{
						chainSummarySections(3),
						chainSummaryQAHead(cast.Completion, chainSummaryReply, "Question 1", "Question 2", "Question 3"),
						chainSummaryKeptTail(1),
					},
				},
				{
					grow:  chainSummaryAddQA("Question 6", 100),
					calls: 2,
					checks: []chainSummaryCheck{
						chainSummarySections(3),
						chainSummaryQAHead(cast.Completion, chainSummaryReply, "Question 1", "Question 2", "Question 3", "Question 4"),
						chainSummaryKeptTail(1),
					},
				},
			},
			stable: true,
		},
		{
			name:    "the last section is rotated, then the oldest sections collapse as the chain grows",
			cfg:     combined,
			initial: []*cast.ChainSection{chainSummaryFirst(chainSummaryQA("Question 1", "Answer 1"))},
			steps: []chainSummaryStep{
				{
					grow:    chainSummaryAddAnswers(5, 200),
					calls:   1,
					sent:    []string{"Question 1", "Answer 1", "Response 0", "Response 3"},
					notSent: []string{"Response 4"},
					checks: []chainSummaryCheck{
						chainSummarySections(1), chainSummaryHeaderKept(0), chainSummaryBodyAt(0, []int{chainSummaryNew, 5}, cast.Completion, false),
					},
				},
				{
					grow:    chainSummaryAddQA("Question 2", 100),
					calls:   1,
					sent:    []string{"**summarized content:**\nsummary text", "Response 4"},
					notSent: []string{"Answer to Question 2"},
					checks: []chainSummaryCheck{
						chainSummarySections(2), chainSummaryHeaderKept(0),
						chainSummaryBodyAt(0, []int{chainSummaryNew}, cast.Completion, false), chainSummaryKeptTail(1),
					},
				},
				{
					grow:    chainSummaryAddQA("Question 3", 100),
					calls:   1,
					sent:    []string{"Answer to Question 2"},
					notSent: []string{"Answer to Question 3"},
					checks: []chainSummaryCheck{
						chainSummarySections(3), chainSummaryBodyAt(1, []int{chainSummaryNew}, cast.Completion, false), chainSummaryKeptTail(1),
					},
				},
				{
					grow:    chainSummaryAddQA("Question 4", 100),
					calls:   2,
					sent:    []string{"Answer to Question 3", "Question 1"},
					notSent: []string{"Answer to Question 4"},
					checks: []chainSummaryCheck{
						chainSummarySections(3), chainSummaryQAHead(cast.Completion, chainSummaryReply, "Question 1", "Question 2"),
						chainSummaryBodyAt(1, []int{chainSummaryNew}, cast.Completion, false), chainSummaryKeptTail(1),
					},
				},
			},
			stable: true,
		},
		{
			name:    "the oldest sections collapse into one once the chain outgrows the byte budget",
			cfg:     SummarizerConfig{UseQA: true, MaxQASections: 10, MaxQABytes: 1600, MaxBPBytes: 5000},
			initial: []*cast.ChainSection{chainSummaryFirst(chainSummaryQA("Question 1", strings.Repeat("A", 1500)))},
			steps: []chainSummaryStep{
				{
					grow:    chainSummaryAddQA("Question 2", 1500),
					calls:   1,
					sent:    []string{"Question 1", "AAA"},
					notSent: []string{"Question 2"},
					checks: []chainSummaryCheck{
						chainSummarySections(2), chainSummaryHeaderKept(0),
						chainSummaryBodyAt(0, []int{chainSummaryNew}, cast.Completion, false), chainSummaryKeptTail(1),
					},
				},
				{
					grow:    chainSummaryAddQA("Question 3", 1500),
					calls:   2,
					sent:    []string{"Answer to Question 2"},
					notSent: []string{"Question 3"},
					checks: []chainSummaryCheck{
						chainSummarySections(2), chainSummaryQAHead(cast.Completion, chainSummaryReply, "Question 1", "Question 2"),
						chainSummaryKeptTail(1),
					},
				},
				{
					grow:    chainSummaryAddQA("Question 4", 1500),
					calls:   2,
					sent:    []string{"Answer to Question 3", "Question 1"},
					notSent: []string{"Question 4"},
					checks: []chainSummaryCheck{
						chainSummarySections(2),
						chainSummaryQAHead(cast.Completion, chainSummaryReply, "Question 1", "Question 2", "Question 3"),
						chainSummaryKeptTail(1),
					},
				},
			},
			stable: true,
		},
		{
			name: "with both limits exceeded the section limit decides when it is the tighter",
			cfg:  SummarizerConfig{UseQA: true, MaxQASections: 2, MaxQABytes: 3800},
			initial: []*cast.ChainSection{
				chainSummaryFirst(chainSummaryQA("Question 1", chainSummaryPad("**summarized content:**\nSummary 1 ", 600))),
				chainSummaryQA("Question 2", chainSummaryPad("**summarized content:**\nSummary 2 ", 600)),
				chainSummaryQA("Question 3", chainSummaryPad("**summarized content:**\nSummary 3 ", 600)),
				chainSummaryQA("Question 4", chainSummaryPad("**summarized content:**\nSummary 4 ", 600)),
				chainSummaryQA("Question 5", chainSummaryPad("Answer 5 ", 1500)),
			},
			steps: []chainSummaryStep{{
				calls:   1,
				sent:    []string{"Summary 1 ", "Summary 3 "},
				notSent: []string{"Summary 4 ", "Answer 5 "},
				checks: []chainSummaryCheck{
					chainSummaryInputSize(3964), chainSummarySections(3), chainSummaryKeptTail(2),
					chainSummaryQAHead(cast.Completion, chainSummaryReply, "Question 1", "Question 2", "Question 3"),
				},
			}},
			stable: true,
		},
		{
			name: "with both limits exceeded the byte budget decides when it is the tighter",
			cfg:  SummarizerConfig{UseQA: true, MaxQASections: 3, MaxQABytes: 3000},
			initial: []*cast.ChainSection{
				chainSummaryFirst(chainSummaryQA("Question 1", chainSummaryPad("**summarized content:**\nSummary 1 ", 600))),
				chainSummaryQA("Question 2", chainSummaryPad("**summarized content:**\nSummary 2 ", 600)),
				chainSummaryQA("Question 3", chainSummaryPad("**summarized content:**\nSummary 3 ", 600)),
				chainSummaryQA("Question 4", chainSummaryPad("**summarized content:**\nSummary 4 ", 600)),
				chainSummaryQA("Question 5", chainSummaryPad("Answer 5 ", 1500)),
			},
			steps: []chainSummaryStep{{
				calls:   1,
				sent:    []string{"Summary 1 ", "Summary 4 "},
				notSent: []string{"Answer 5 "},
				checks: []chainSummaryCheck{
					chainSummaryInputSize(3964), chainSummarySections(2), chainSummaryKeptTail(1),
					chainSummaryQAHead(cast.Completion, chainSummaryReply, "Question 1", "Question 2", "Question 3", "Question 4"),
				},
			}},
			stable: true,
		},
		{
			name:     "summarizing questions replaces them with their own summary",
			cfg:      SummarizerConfig{UseQA: true, SummHumanInQA: true, MaxQASections: 2, MaxQABytes: 10000},
			numbered: true,
			initial: []*cast.ChainSection{
				chainSummaryFirst(chainSummaryQA("Question 1", "**summarized content:**\nAnswer 1")),
				chainSummaryQA("Question 2", "**summarized content:**\nAnswer 2"),
				chainSummaryQA("Question 3", "**summarized content:**\nAnswer 3"),
				chainSummaryQA("Question 4", "Answer 4"),
			},
			steps: []chainSummaryStep{{
				calls: 2,
				prompts: []string{
					twoTasks,
					twoTasks + "<messages>\n" +
						"<message id=\"0\" role=\"ai\">\n<content part=\"0\">\n**summarized content:**\nAnswer 1\n</content>\n</message>\n" +
						"<message id=\"1\" role=\"ai\">\n<content part=\"0\">\n**summarized content:**\nAnswer 2\n</content>\n</message>\n" +
						"</messages>",
				},
				checks: []chainSummaryCheck{
					chainSummarySections(3), chainSummaryQAHead(cast.Completion, "SUMMARY-002", "SUMMARY-001"), chainSummaryKeptTail(2),
				},
			}},
			stable: true,
		},
		{
			name: "every section but the kept two is summarized",
			cfg:  SummarizerConfig{KeepQASections: 2, MaxBPBytes: 5000},
			initial: []*cast.ChainSection{
				chainSummaryFirst(chainSummaryQA("Question 1", "Answer 1a", "Answer 1b")),
				chainSummaryQA("Question 2", "Answer 2a", "Answer 2b"),
				chainSummaryQA("Question 3", "Answer 3a", "Answer 3b"),
				chainSummaryQA("Question 4", "Answer 4a", "Answer 4b"),
			},
			steps: []chainSummaryStep{{
				calls:   2,
				sent:    []string{"Answer 1a", "Answer 2b"},
				notSent: []string{"Answer 3a", "Answer 4b"},
				checks: []chainSummaryCheck{
					chainSummarySections(4), chainSummaryHeaderKept(0),
					chainSummaryBodyAt(0, []int{chainSummaryNew}, cast.Completion, false),
					chainSummaryBodyAt(1, []int{chainSummaryNew}, cast.Completion, false), chainSummaryKeptTail(2),
				},
			}},
			stable: true,
		},
		{
			name:    "only the pair over the pair budget is summarized",
			cfg:     SummarizerConfig{PreserveLast: true, LastSecBytes: 30 * 1024, MaxBPBytes: 1000},
			initial: []*cast.ChainSection{chainSummaryFirst(chainSummaryQA("Question requiring various size responses"))},
			steps: []chainSummaryStep{{
				grow: func(ast *cast.ChainAST) {
					last := ast.Sections[0]
					for _, text := range []string{
						"Short initial response", strings.Repeat("A", 300), strings.Repeat("B", 1000), // at the budget, not over it
						strings.Repeat("C", 2000), "Another normal response",
					} {
						last.AddBodyPair(cast.NewBodyPairFromCompletion(text))
					}
					for i := range 10 {
						last.AddBodyPair(cast.NewBodyPairFromCompletion(fmt.Sprintf("Additional message %d", i)))
					}
				},
				calls:   1,
				sent:    []string{"CCC"},
				notSent: []string{"BBB", "Another normal response"},
				checks: []chainSummaryCheck{chainSummaryBodyAt(0,
					[]int{0, 1, 2, chainSummaryNew, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}, cast.Completion, false)},
			}},
			stable: true,
		},
		{
			name: "the newest pair is never summarized however large",
			cfg:  SummarizerConfig{PreserveLast: true, LastSecBytes: 100 * 1024, MaxBPBytes: 16 * 1024, KeepQASections: 1},
			initial: []*cast.ChainSection{chainSummaryQA("Question",
				strings.Repeat("A", 20*1024)+"First", strings.Repeat("B", 20*1024)+"Second", strings.Repeat("C", 20*1024)+"Last")},
			steps: []chainSummaryStep{{
				calls:   2,
				sent:    []string{"AFirst", "BSecond"},
				notSent: []string{"CLast"},
				checks:  []chainSummaryCheck{chainSummaryBodyAt(0, []int{chainSummaryNew, chainSummaryNew, 2}, cast.Completion, false)},
			}},
			stable: true,
		},
		{
			name: "an oversized newest pair is kept until a newer pair arrives, then summarized in place",
			cfg:  SummarizerConfig{PreserveLast: true, LastSecBytes: 50 * 1024, MaxBPBytes: 16 * 1024},
			initial: []*cast.ChainSection{
				chainSummaryFirst(chainSummaryQA("Question with potentially large responses", "Initial normal response")),
			},
			steps: []chainSummaryStep{
				{
					grow: chainSummaryAddPairs(
						cast.NewBodyPairFromCompletion("Another normal response that is well within size limits"),
						cast.NewBodyPairFromCompletion("Oversized response: "+strings.Repeat("X", 17*1024)),
					),
					checks: []chainSummaryCheck{chainSummaryUnchanged},
				},
				{
					grow:    chainSummaryAddPairs(cast.NewBodyPairFromCompletion("A newer response")),
					calls:   1,
					sent:    []string{"Question with potentially large responses", "Oversized response: XXX"},
					notSent: []string{"Initial normal response", "Another normal", "A newer response"},
					checks:  []chainSummaryCheck{chainSummaryBodyAt(0, []int{0, 1, chainSummaryNew, 3}, cast.Completion, false)},
				},
			},
			stable: true,
		},
		{
			name:    "a tool exchange rotated out of the last section becomes a summarization call",
			cfg:     SummarizerConfig{PreserveLast: true, LastSecBytes: 2000, MaxBPBytes: 5000},
			initial: []*cast.ChainSection{chainSummaryFirst(chainSummaryQA("Initial question", "Initial response"))},
			steps: []chainSummaryStep{
				{
					grow: chainSummaryAddPairs(chainSummaryToolPair("search-id", "search", "Tool response for search", nil,
						llms.TextContent{Text: "Let me use a tool"})),
					checks: []chainSummaryCheck{chainSummaryUnchanged},
				},
				{
					grow:    chainSummaryAddAnswers(6, 500),
					calls:   1,
					sent:    []string{"Initial response", "Tool response for search", "Response 3"},
					notSent: []string{"Response 4", "Response 5"},
					checks:  []chainSummaryCheck{chainSummaryBodyAt(0, []int{chainSummaryNew, 6, 7}, cast.Summarization, false)},
				},
			},
			stable: true,
		},
		{
			name: "each rotation folds the previous summarization call into the next one",
			cfg:  SummarizerConfig{PreserveLast: true, LastSecBytes: 2000, MaxBPBytes: 5000},
			initial: []*cast.ChainSection{
				chainSummaryFirst(cast.NewChainSection(cast.NewHeader(nil, chainSummaryMsg(llms.ChatMessageTypeHuman, "Run the scan")), nil)),
			},
			steps: []chainSummaryStep{
				{grow: chainSummaryAddPairs(terminal(1)), checks: []chainSummaryCheck{chainSummaryUnchanged}},
				{grow: chainSummaryAddPairs(terminal(2)), checks: []chainSummaryCheck{chainSummaryUnchanged}},
				{
					grow:    chainSummaryAddPairs(terminal(3)),
					calls:   1,
					sent:    []string{"out 1 "},
					notSent: []string{"out 2 ", "out 3 "},
					checks:  []chainSummaryCheck{chainSummaryBodyAt(0, []int{chainSummaryNew, 1, 2}, cast.Summarization, false)},
				},
				{
					grow:    chainSummaryAddPairs(terminal(4)),
					calls:   1,
					sent:    []string{summaryResponse, "out 2 "},
					notSent: []string{"out 3 ", "out 4 "},
					checks:  []chainSummaryCheck{chainSummaryBodyAt(0, []int{chainSummaryNew, 2, 3}, cast.Summarization, false)},
				},
				{
					grow:    chainSummaryAddPairs(terminal(5)),
					calls:   1,
					sent:    []string{summaryResponse, "out 3 "},
					notSent: []string{"out 4 ", "out 5 "},
					checks:  []chainSummaryCheck{chainSummaryBodyAt(0, []int{chainSummaryNew, 2, 3}, cast.Summarization, false)},
				},
			},
			stable: true,
		},
		{
			// Not stable: a second pass summarizes the completion summary again, see determineLastSectionPairs.
			name: "the newest tool exchange survives the rotation of the last section",
			cfg:  SummarizerConfig{PreserveLast: true, LastSecBytes: 500, MaxBPBytes: 1000, KeepQASections: 1},
			initial: []*cast.ChainSection{chainSummarySection("Question",
				cast.NewBodyPairFromCompletion(strings.Repeat("A", 100)+"First"),
				cast.NewBodyPairFromCompletion(strings.Repeat("B", 100)+"Second"),
				chainSummaryToolPair("call_test_large", "search", strings.Repeat("Result: ", 10000), nil),
			)},
			steps: []chainSummaryStep{{
				calls:   1,
				sent:    []string{"AFirst", "BSecond"},
				notSent: []string{"Result: "},
				checks:  []chainSummaryCheck{chainSummaryBodyAt(0, []int{chainSummaryNew, 2}, cast.Completion, false)},
			}},
		},
		{
			name: "a first tool call returning 90 KB stays whole with its signed thinking block",
			cfg:  flow,
			initial: []*cast.ChainSection{chainSummaryFirst(chainSummarySection("Pentest the target",
				chainSummaryToolPair("toolu_011qigRrFEuu5dHKE78v3CuN", "terminal", strings.Repeat("Response data: ", 6000), nil,
					chainSummarySignedThought("The data isn't reflected. Let me try send.php.", "anthropic_crypto_signature")),
			))},
			steps:  []chainSummaryStep{{checks: []chainSummaryCheck{chainSummaryUnchanged}}},
			stable: true,
		},
		{
			name: "a lone summarization call signed by Gemini stays whole",
			cfg:  flow,
			initial: []*cast.ChainSection{chainSummaryFirst(chainSummarySection("Pentest the target",
				chainSummaryToolPair("fcall_test123gemini", "execute_task_and_return_summary",
					strings.Repeat("Data row: extensive output\n", 2000), chainSummarySignature("gemini_thought_signature_data"),
					llms.TextContent{Text: "Let me execute this"}),
			))},
			steps:  []chainSummaryStep{{checks: []chainSummaryCheck{chainSummaryUnchanged}}},
			stable: true,
		},
		{
			name: "a last QA section over the byte budget is kept intact",
			cfg: SummarizerConfig{
				UseQA: true, MaxQASections: 5, MaxQABytes: 64000, KeepQASections: 1, MaxBPBytes: 16 * 1024,
			},
			initial: []*cast.ChainSection{
				chainSummaryFirst(chainSummaryQA("Question 1", "Answer 1")),
				chainSummaryQA("Question 2", "Answer 2"),
				chainSummarySection("Question 3 - search for code", chainSummaryToolPair("call_search", "search_code",
					strings.Repeat("Code result line\n", 5000), nil, llms.TextContent{Text: "Let me search for that"})),
			},
			steps: []chainSummaryStep{{
				calls:   3,
				sent:    []string{"Answer 1", "Answer 2", "**summarized content:**\nsummary text"},
				notSent: []string{"Code result line"},
				checks: []chainSummaryCheck{
					chainSummarySections(2), chainSummaryQAHead(cast.Completion, chainSummaryReply, "Question 1", "Question 2"),
					chainSummaryKeptTail(1),
				},
			}},
			stable: true,
		},
		{
			name: "an unset last-section budget keeps a section of exactly 50 KB",
			cfg:  SummarizerConfig{PreserveLast: true},
			initial: []*cast.ChainSection{chainSummaryQA("Question", chainSummaryPad("Answer 0: ", 12800),
				chainSummaryPad("Answer 1: ", 12800), chainSummaryPad("Answer 2: ", 12800), chainSummaryPad("Answer 3: ", 12792))},
			steps:  []chainSummaryStep{{checks: []chainSummaryCheck{chainSummaryInputSize(51200), chainSummaryUnchanged}}},
			stable: true,
		},
		{
			name: "an unset last-section budget rotates a section one byte over 50 KB",
			cfg:  SummarizerConfig{PreserveLast: true},
			initial: []*cast.ChainSection{chainSummaryQA("Question", chainSummaryPad("Answer 0: ", 12800),
				chainSummaryPad("Answer 1: ", 12800), chainSummaryPad("Answer 2: ", 12800), chainSummaryPad("Answer 3: ", 12793))},
			steps: []chainSummaryStep{{
				calls:   1,
				sent:    []string{"Answer 0: ", "Answer 1: "},
				notSent: []string{"Answer 2: ", "Answer 3: "},
				checks: []chainSummaryCheck{
					chainSummaryInputSize(51201), chainSummaryBodyAt(0, []int{chainSummaryNew, 2, 3}, cast.Completion, false),
				},
			}},
			stable: true,
		},
		{
			name:    "an unset pair budget keeps an older pair of exactly 16 KB",
			cfg:     SummarizerConfig{PreserveLast: true, LastSecBytes: 1 << 20},
			initial: []*cast.ChainSection{chainSummaryQA("Question", chainSummaryPad("Answer 0: ", 16384), "Answer 1")},
			steps:   []chainSummaryStep{{checks: []chainSummaryCheck{chainSummaryUnchanged}}},
			stable:  true,
		},
		{
			name:    "an unset pair budget summarizes an older pair one byte over 16 KB",
			cfg:     SummarizerConfig{PreserveLast: true, LastSecBytes: 1 << 20},
			initial: []*cast.ChainSection{chainSummaryQA("Question", chainSummaryPad("Answer 0: ", 16385), "Answer 1")},
			steps: []chainSummaryStep{{
				calls:   1,
				sent:    []string{"Answer 0: "},
				notSent: []string{"Answer 1"},
				checks:  []chainSummaryCheck{chainSummaryBodyAt(0, []int{chainSummaryNew, 1}, cast.Completion, false)},
			}},
			stable: true,
		},
		{
			// Ten newest sections are kept; the one older section is a summary already.
			name:    "an unset QA section limit keeps eleven sections whose oldest is a summary",
			cfg:     SummarizerConfig{UseQA: true, MaxQABytes: 1 << 20},
			initial: chainSummaryHistory(11),
			steps:   []chainSummaryStep{{checks: []chainSummaryCheck{chainSummaryUnchanged}}},
			stable:  true,
		},
		{
			name:    "an unset QA section limit collapses what is older than the ten newest sections",
			cfg:     SummarizerConfig{UseQA: true, MaxQABytes: 1 << 20},
			initial: chainSummaryHistory(12),
			steps: []chainSummaryStep{{
				calls:   1,
				sent:    []string{"Summary 00", "Summary 01"},
				notSent: []string{"Summary 02", "Answer 11"},
				checks: []chainSummaryCheck{
					chainSummarySections(11), chainSummaryQAHead(cast.Completion, chainSummaryReply, "Question 00", "Question 01"),
					chainSummaryKeptTail(10),
				},
			}},
			stable: true,
		},
		{
			name:    "an unset QA byte budget keeps a chain of exactly 64 KB",
			cfg:     SummarizerConfig{UseQA: true, MaxQASections: 100},
			initial: chainSummaryBudgetChain(31768),
			steps:   []chainSummaryStep{{checks: []chainSummaryCheck{chainSummaryInputSize(65536), chainSummaryUnchanged}}},
			stable:  true,
		},
		{
			name:    "an unset QA byte budget collapses a chain one byte over 64 KB",
			cfg:     SummarizerConfig{UseQA: true, MaxQASections: 100},
			initial: chainSummaryBudgetChain(31769),
			steps: []chainSummaryStep{{
				calls:   1,
				sent:    []string{"Summary 0 ", "Summary 1 "},
				notSent: []string{"Summary 2 ", "Answer 3 "},
				checks: []chainSummaryCheck{
					chainSummaryInputSize(65537), chainSummarySections(3),
					chainSummaryQAHead(cast.Completion, chainSummaryReply, "Question 0", "Question 1"), chainSummaryKeptTail(2),
				},
			}},
			stable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &chainSummaryRecorder{numbered: tt.numbered}
			summarizer := NewSummarizer(tt.cfg)
			ast := &cast.ChainAST{Sections: tt.initial}
			for i, step := range tt.steps {
				ok := t.Run(fmt.Sprintf("step %d", i+1), func(t *testing.T) {
					if step.grow != nil {
						step.grow(ast)
					}
					in := ast.Messages()
					before, err := cast.NewChainAST(in, false)
					require.NoError(t, err)
					called := len(rec.sent())

					out, err := summarizer.SummarizeChain(t.Context(), rec.handle, in, cast.ToolCallIDTemplate)
					require.NoError(t, err)
					prompts := rec.sent()[called:]
					require.Len(t, prompts, step.calls, "summarizer calls")
					chainSummaryAssertSent(t, prompts, step.sent, step.notSent)
					for j, want := range step.prompts {
						assert.Equal(t, want, chainSummaryPromptData(t, prompts[j]), "prompt %d", j)
					}
					got, err := cast.NewChainAST(out, false)
					require.NoError(t, err)
					chainSummaryVerifyAST(t, got)
					for _, check := range step.checks {
						check(t, got, before)
					}

					if tt.stable {
						again, err := summarizer.SummarizeChain(t.Context(), rec.handle, out, cast.ToolCallIDTemplate)
						require.NoError(t, err)
						assert.Len(t, rec.sent(), called+step.calls, "a second pass called the summarizer")
						assert.Equal(t, chainSummaryJSON(t, out), chainSummaryJSON(t, again), "a second pass changed the chain")
					}
					ast = got
				})
				if !ok {
					return
				}
			}
		})
	}
}

// chainSummaryIDs returns the numbered summaries a chain holds.
func chainSummaryIDs(chain []llms.MessageContent) []string {
	var ids []string
	for _, msg := range chain {
		for _, part := range msg.Parts {
			var text string
			switch v := part.(type) {
			case llms.TextContent:
				text = v.Text
			case llms.ToolCallResponse:
				text = v.Content
			}
			for _, field := range strings.Fields(text) {
				if strings.HasPrefix(field, "SUMMARY-") {
					ids = append(ids, field)
				}
			}
		}
	}
	return ids
}

func TestChainSummary_SummarizeChain_CarriesHistoryWithinBudgetOverALongConversation(t *testing.T) {
	type growth struct {
		opening []llms.MessageContent // closes the previous turn or opens a new one
		newest  []llms.MessageContent
	}
	exchange := func(id string, size int) []llms.MessageContent {
		output := fmt.Sprintf("output of %s: ", id) + strings.Repeat("r", size)
		return chainSummaryToolPair(id, "terminal", output, nil, llms.TextContent{Text: "running " + id}).Messages()
	}

	var agent []growth
	for i := range 96 {
		var g growth
		if i%6 == 5 {
			g.opening = []llms.MessageContent{
				*chainSummaryMsg(llms.ChatMessageTypeAI, fmt.Sprintf("Task %d is done", i/6)),
				*chainSummaryMsg(llms.ChatMessageTypeHuman, fmt.Sprintf("Next task %d", i/6+1)),
			}
		}
		size := 10000
		if i%5 == 3 {
			size = 20000
		}
		g.newest = exchange(fmt.Sprintf("call_%d", i), size)
		agent = append(agent, g)
	}

	var assistant []growth
	for turn := range 12 {
		for k := range 7 {
			var g growth
			if k == 0 {
				g.opening = []llms.MessageContent{*chainSummaryMsg(llms.ChatMessageTypeHuman, fmt.Sprintf("Question %d", turn))}
			}
			size := 16000
			if k == 2 {
				size = 20000
			}
			g.newest = exchange(fmt.Sprintf("call_%d_%d", turn, k), size)
			if k == 6 {
				g.newest = []llms.MessageContent{*chainSummaryMsg(llms.ChatMessageTypeAI, fmt.Sprintf("Answer %d", turn))}
			}
			assistant = append(assistant, g)
		}
	}

	tests := []struct {
		name  string
		cfg   SummarizerConfig
		start []llms.MessageContent
		steps []growth
	}{
		{
			name: "a flow agent under the default flow settings",
			cfg: SummarizerConfig{
				PreserveLast: true, UseQA: true, LastSecBytes: 51200, MaxBPBytes: 16384,
				MaxQASections: 10, MaxQABytes: 65536, KeepQASections: 1,
			},
			start: []llms.MessageContent{
				*chainSummaryMsg(llms.ChatMessageTypeSystem, "You are a pentester"),
				*chainSummaryMsg(llms.ChatMessageTypeHuman, "Next task 0"),
			},
			steps: agent,
		},
		{
			name: "the assistant under the default assistant settings",
			cfg: SummarizerConfig{
				PreserveLast: true, UseQA: true, LastSecBytes: 76800, MaxBPBytes: 16384,
				MaxQASections: 7, MaxQABytes: 76800, KeepQASections: 3,
			},
			start: []llms.MessageContent{*chainSummaryMsg(llms.ChatMessageTypeSystem, "You are an assistant")},
			steps: assistant,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &chainSummaryRecorder{numbered: true}
			summarizer := NewSummarizer(tt.cfg)
			chain := tt.start
			folded := 0
			for i, g := range tt.steps {
				in := slices.Concat(chain, g.opening, g.newest)
				called := len(rec.sent())

				out, err := summarizer.SummarizeChain(t.Context(), rec.handle, in, cast.ToolCallIDTemplate)
				require.NoError(t, err, "step %d", i)
				prompts := strings.Join(rec.sent()[called:], "\n")
				kept := chainSummaryIDs(out)
				for _, id := range chainSummaryIDs(in) {
					if !slices.Contains(kept, id) {
						folded++
						assert.Contains(t, prompts, id, "step %d: %s left the chain without reaching the summarizer", i, id)
					}
				}
				require.GreaterOrEqual(t, len(out), len(g.newest))
				assert.Equal(t, chainSummaryJSON(t, g.newest), chainSummaryJSON(t, out[len(out)-len(g.newest):]),
					"step %d: the newest pair", i)

				got, err := cast.NewChainAST(out, false)
				require.NoError(t, err)
				chainSummaryVerifyAST(t, got)
				assert.LessOrEqual(t, len(got.Sections), tt.cfg.MaxQASections+1, "step %d: sections", i)
				if got.Size() > tt.cfg.MaxQABytes {
					assert.LessOrEqual(t, len(got.Sections), 1+tt.cfg.KeepQASections,
						"step %d: a chain over the byte budget holds more than its summary and the kept sections", i)
				}
				// Every summary here covers tool traffic, so it must be a summarization call.
				for j, section := range got.Sections {
					if j < len(got.Sections)-tt.cfg.KeepQASections {
						require.Len(t, section.Body, 1, "step %d: older section %d", i, j)
						assert.Equal(t, cast.Summarization, section.Body[0].Type, "step %d: older section %d", i, j)
						continue
					}
					assert.LessOrEqual(t, section.Size(), tt.cfg.LastSecBytes, "step %d: kept section %d", i, j)
					for k, pair := range section.Body {
						assert.False(t, pair.Type == cast.Completion && chainSummaryIsSummary(pair),
							"step %d: kept section %d pair %d is a completion summary", i, j, k)
					}
				}

				calls := len(rec.sent())
				again, err := summarizer.SummarizeChain(t.Context(), rec.handle, out, cast.ToolCallIDTemplate)
				require.NoError(t, err)
				assert.Len(t, rec.sent(), calls, "step %d: a second pass called the summarizer", i)
				assert.Equal(t, chainSummaryJSON(t, out), chainSummaryJSON(t, again), "step %d: a second pass changed the chain", i)
				chain = out
			}
			assert.Positive(t, folded, "no summary was ever folded into a later one")
		})
	}
}

func TestChainSummary_SummarizeChain_ReturnsTheInputChainWhenSummarizingFails(t *testing.T) {
	tests := []struct {
		name     string
		cfg      SummarizerConfig
		sections []*cast.ChainSection
		chain    []llms.MessageContent // given as is when set, in place of sections
		calls    int
		wantErr  string
	}{
		{
			name:     "an older section fails",
			sections: []*cast.ChainSection{chainSummaryQA("Question 1", "Answer 1"), chainSummaryQA("Question 2", "Answer 2")},
			calls:    1,
			wantErr:  "failed to summarize sections: failed to summarize sections: section 0 summary generation failed",
		},
		{
			name:     "the last section fails",
			cfg:      SummarizerConfig{PreserveLast: true, LastSecBytes: 100},
			sections: []*cast.ChainSection{chainSummaryQA("Question", strings.Repeat("A", 100), strings.Repeat("B", 100))},
			calls:    1,
			wantErr:  "failed to summarize last section 0: last section summary generation failed",
		},
		{
			name:     "an oversized pair whose summary fails is reported by the rotation after it",
			cfg:      SummarizerConfig{PreserveLast: true, LastSecBytes: 1000, MaxBPBytes: 500},
			sections: []*cast.ChainSection{chainSummaryQA("Question", strings.Repeat("A", 2000), strings.Repeat("B", 100))},
			calls:    2,
			wantErr:  "failed to summarize last section 0: last section summary generation failed",
		},
		{
			name: "the QA sections fail",
			cfg:  SummarizerConfig{UseQA: true, MaxQASections: 1},
			sections: []*cast.ChainSection{
				chainSummaryQA("Question 1", "**summarized content:**\nAnswer 1"),
				chainSummaryQA("Question 2", "**summarized content:**\nAnswer 2"),
				chainSummaryQA("Question 3", "Answer 3"),
			},
			calls:   1,
			wantErr: "failed to summarize QA pairs: QA (ai) summary generation failed",
		},
		{
			name:    "a chain opening with an AI message cannot be parsed",
			chain:   []llms.MessageContent{*chainSummaryMsg(llms.ChatMessageTypeAI, "Answer")},
			wantErr: "failed to create ChainAST: unexpected chain begin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := tt.chain
			if in == nil {
				in = (&cast.ChainAST{Sections: tt.sections}).Messages()
			}
			want := chainSummaryJSON(t, in)
			rec := &chainSummaryRecorder{fail: true}

			out, err := NewSummarizer(tt.cfg).SummarizeChain(t.Context(), rec.handle, in, cast.ToolCallIDTemplate)
			require.ErrorContains(t, err, tt.wantErr)
			assert.Len(t, rec.sent(), tt.calls, "summarizer calls")
			assert.Equal(t, want, chainSummaryJSON(t, out), "the chain should come back as it was given")
		})
	}
}

func TestChainSummary_SummarizeSections_CollapsesEachOlderSectionIntoOneSummary(t *testing.T) {
	failing := make([]*cast.ChainSection, 0, 5)
	for i := range 5 {
		failing = append(failing, chainSummaryQA(fmt.Sprintf("Question %d", i),
			fmt.Sprintf("Answer %d", i), fmt.Sprintf("Answer %d continued", i)))
	}

	tests := []struct {
		name       string
		sections   []*cast.ChainSection
		keep       int
		fail       bool
		summarized map[int]cast.BodyPairType // sections replaced by a summary of that type; the rest stay untouched
		calls      int
		sent       []string
		notSent    []string
		wantErr    []string
	}{
		{name: "an empty chain is left alone", keep: 1},
		{
			name: "sections already holding one summary are skipped",
			sections: []*cast.ChainSection{
				chainSummaryQA("Question 1", "**summarized content:**\nAnswer 1"),
				chainSummarySection("Question 2", cast.NewBodyPairFromSummarization("Answer 2", cast.ToolCallIDTemplate, false, nil)),
				chainSummaryQA("Question 3", "Answer 3", "Answer 3 continued"),
			},
			keep: 1,
		},
		{
			name: "older sections are summarized, a tool exchange into a summarization call",
			sections: []*cast.ChainSection{
				chainSummaryFirst(chainSummaryQA("Question 1", "Answer 1a", "Answer 1b")),
				chainSummarySection("Question 2",
					chainSummaryToolPair("search-tool-1", "search", "Search results", nil, llms.TextContent{Text: "Let me search"}),
					cast.NewBodyPairFromCompletion("Based on the search, here's my answer")),
				chainSummaryQA("Follow-up question", "Final answer"),
			},
			keep:       1,
			summarized: map[int]cast.BodyPairType{0: cast.Completion, 1: cast.Summarization},
			calls:      2,
			sent:       []string{"Question 1", "Answer 1a", "Answer 1b", "Question 2", "Search results"},
			notSent:    []string{"Final answer"},
		},
		{
			name: "a signed tool exchange of a previous turn is summarized without a signature",
			sections: []*cast.ChainSection{
				chainSummarySection("Question 1",
					chainSummaryToolPair("call_reasoning_1", "search", "Result 1", chainSummarySignature("gemini_signature_abc123"))),
				chainSummaryQA("Question 2", "Answer 2"),
			},
			keep:       1,
			summarized: map[int]cast.BodyPairType{0: cast.Summarization},
			calls:      1,
		},
		{
			name: "keeping as many sections as there are changes nothing",
			sections: []*cast.ChainSection{
				chainSummaryQA("Question 1", "Answer 1a", "Answer 1b"),
				chainSummaryQA("Question 2", "Answer 2"),
				chainSummaryQA("Question 3", "Answer 3"),
			},
			keep: 3,
		},
		{
			name: "keeping more sections than there are changes nothing",
			sections: []*cast.ChainSection{
				chainSummaryQA("Question 1", "Answer 1"),
				chainSummaryQA("Question 2", "Answer 2a", "Answer 2b"),
			},
			keep: 5,
		},
		{
			name:     "a failure names every section that failed",
			sections: failing,
			keep:     1,
			fail:     true,
			calls:    4,
			wantErr:  []string{"summary generation failed", "section 0", "section 1", "section 2", "section 3"},
		},
		{
			name: "a completion with no parts is summarized rather than mistaken for a summary",
			sections: []*cast.ChainSection{
				chainSummarySection("Question 1", cast.NewBodyPair(&llms.MessageContent{Role: llms.ChatMessageTypeAI}, nil)),
				chainSummaryQA("Question 2", "Answer 2"),
			},
			keep:       1,
			summarized: map[int]cast.BodyPairType{0: cast.Completion},
			calls:      1,
		},
		{
			name: "a completion whose first part is not text is summarized rather than mistaken for a summary",
			sections: []*cast.ChainSection{
				chainSummarySection("Question 1", cast.NewBodyPair(&llms.MessageContent{
					Role:  llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{llms.ImageURLContent{URL: "http://img"}},
				}, nil)),
				chainSummaryQA("Question 2", "Answer 2"),
			},
			keep:       1,
			summarized: map[int]cast.BodyPairType{0: cast.Completion},
			calls:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast := &cast.ChainAST{Sections: slices.Clone(tt.sections)}
			before := make([]string, 0, len(tt.sections))
			for _, section := range tt.sections {
				before = append(before, chainSummaryJSON(t, section))
			}
			rec := &chainSummaryRecorder{fail: tt.fail}

			err := summarizeSections(t.Context(), ast, rec.handle, tt.keep, cast.ToolCallIDTemplate)
			assert.Len(t, rec.sent(), tt.calls, "summarizer calls")
			chainSummaryAssertSent(t, rec.sent(), tt.sent, tt.notSent)
			if tt.wantErr != nil {
				require.Error(t, err)
				for _, want := range tt.wantErr {
					assert.ErrorContains(t, err, want)
				}
				return
			}
			require.NoError(t, err)

			require.Len(t, ast.Sections, len(before))
			for i, section := range ast.Sections {
				if kind, ok := tt.summarized[i]; ok {
					require.Len(t, section.Body, 1, "section %d", i)
					chainSummaryCheckSummary(t, section.Body[0], kind, false)
					continue
				}
				assert.Equal(t, before[i], chainSummaryJSON(t, section), "section %d should be untouched", i)
			}
			chainSummaryVerifyAST(t, ast)
		})
	}
}

func TestChainSummary_SummarizeLastSection_RotatesOldPairsIntoASummary(t *testing.T) {
	fourAnswers := []string{
		strings.Repeat("A", 100) + "Response 1", strings.Repeat("B", 100) + "Response 2",
		strings.Repeat("C", 100) + "Response 3", strings.Repeat("D", 100) + "Response 4",
	}
	thinking := chainSummarySignedThought("The data isn't reflected. Let me try examining send.php more carefully.",
		"anthropic_crypto_signature_base64")

	tests := []struct {
		name      string
		sections  []*cast.ChainSection
		maxBytes  int
		reserve   int
		hasIndex  bool
		index     int
		body      []int // nil: the chain is left untouched
		kind      cast.BodyPairType
		signed    bool
		lead      []llms.ContentPart
		calls     int
		sent      []string
		notSent   []string
		sentOrder []string // substrings that must appear in this order in what reached the summarizer
	}{
		{name: "an index outside the chain is ignored", maxBytes: 1000, reserve: 25},
		{
			name:     "an index equal to the number of sections is ignored",
			sections: []*cast.ChainSection{chainSummaryQA("Test question", "Test response")},
			maxBytes: 10,
			reserve:  25,
			hasIndex: true,
			index:    1,
		},
		{
			name: "a pair whose running total exactly meets the threshold is kept, oldest summarized first",
			sections: []*cast.ChainSection{chainSummaryQA("Q",
				"FIRST"+strings.Repeat("a", 45), "SECOND"+strings.Repeat("b", 44),
				"THIRD"+strings.Repeat("c", 95), "LAST"+strings.Repeat("d", 96))},
			maxBytes:  201,
			reserve:   0,
			body:      []int{chainSummaryNew, 2, 3},
			kind:      cast.Completion,
			calls:     1,
			sent:      []string{"FIRST", "SECOND"},
			notSent:   []string{"THIRD", "LAST"},
			sentOrder: []string{"FIRST", "SECOND"},
		},
		{
			name: "a rotated signed exchange keeps its reasoning and a fake signature",
			sections: []*cast.ChainSection{chainSummarySection("Analyze this data",
				chainSummaryToolPair("call_current_turn", "analyze", strings.Repeat("Z", 300),
					chainSummarySignature("gemini_current_turn_signature_xyz"),
					chainSummaryThought("Analyzing the data...", "This appears to be a privilege escalation issue.")),
				cast.NewBodyPairFromCompletion("Final response"),
			)},
			maxBytes: 200,
			reserve:  25,
			body:     []int{chainSummaryNew, 1},
			kind:     cast.Summarization,
			signed:   true,
			lead:     []llms.ContentPart{chainSummaryThought("Analyzing the data...", "This appears to be a privilege escalation issue.")},
			calls:    1,
			notSent:  []string{"Final response"},
		},
		{
			name: "a rotated exchange keeps its signed thinking block before an unsigned summary call",
			sections: []*cast.ChainSection{chainSummarySection("Probe the contact form",
				chainSummaryToolPair("toolu_011qigRrFEuu5dHKE78v3CuN", "terminal", strings.Repeat("Response data: ", 30), nil, thinking),
				cast.NewBodyPairFromCompletion("Final response"),
			)},
			maxBytes: 200,
			reserve:  25,
			body:     []int{chainSummaryNew, 1},
			kind:     cast.Summarization,
			lead:     []llms.ContentPart{thinking},
			calls:    1,
			sent:     []string{"Response data: "},
			notSent:  []string{"Final response"},
		},
		{
			name: "a header over the budget with no body is left alone",
			sections: []*cast.ChainSection{cast.NewChainSection(cast.NewHeader(
				chainSummaryMsg(llms.ChatMessageTypeSystem, strings.Repeat("S", 150)),
				chainSummaryMsg(llms.ChatMessageTypeHuman, strings.Repeat("H", 150)),
			), []*cast.BodyPair{})},
			maxBytes: 200,
			reserve:  25,
		},
		{
			name: "an oversized pair is summarized in place before rotation is weighed",
			sections: []*cast.ChainSection{chainSummaryQA("Question with large response",
				"Normal size answer", strings.Repeat("X", 20*1024), "Another normal size answer")},
			maxBytes: 20 * 1024, // over the budget only while the oversized pair is whole
			reserve:  25,
			body:     []int{0, chainSummaryNew, 2},
			kind:     cast.Completion,
			calls:    1,
			sent:     []string{"Question with large response", "XXX"},
			notSent:  []string{"Normal size answer", "Another normal"},
		},
		{
			name:     "no reserve keeps every newest pair that fits",
			sections: []*cast.ChainSection{chainSummaryQA("Test question", fourAnswers...)},
			maxBytes: 300,
			reserve:  0,
			body:     []int{chainSummaryNew, 2, 3},
			kind:     cast.Completion,
			calls:    1,
			sent:     []string{"Response 1", "Response 2"},
			notSent:  []string{"Response 3", "Response 4"},
		},
		{
			name:     "a large reserve keeps only the newest pair",
			sections: []*cast.ChainSection{chainSummaryQA("Test question", fourAnswers...)},
			maxBytes: 300,
			reserve:  50,
			body:     []int{chainSummaryNew, 3},
			kind:     cast.Completion,
			calls:    1,
			sent:     []string{"Response 1", "Response 2", "Response 3"},
			notSent:  []string{"Response 4"},
		},
		{
			name: "a lone summary over the budget is not summarized again",
			sections: []*cast.ChainSection{chainSummarySection("Test question",
				cast.NewBodyPairFromSummarization(strings.Repeat("S", 20*1024), cast.ToolCallIDTemplate, false, nil),
				cast.NewBodyPairFromCompletion("Normal response"),
			)},
			maxBytes: 10 * 1024,
			reserve:  25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast := &cast.ChainAST{Sections: tt.sections}
			before := chainSummaryJSON(t, ast)
			last := len(ast.Sections) - 1
			idx := last
			if tt.hasIndex {
				idx = tt.index
			}
			var original []string
			if last >= 0 {
				original = chainSummaryPairsJSON(t, ast.Sections[last].Body)
			}
			rec := &chainSummaryRecorder{}

			err := summarizeLastSection(t.Context(), ast, rec.handle, idx, tt.maxBytes, 16*1024, tt.reserve, cast.ToolCallIDTemplate)
			require.NoError(t, err)
			assert.Len(t, rec.sent(), tt.calls, "summarizer calls")
			chainSummaryAssertSent(t, rec.sent(), tt.sent, tt.notSent)
			if len(tt.sentOrder) > 1 {
				all := strings.Join(rec.sent(), "\n")
				for i := 1; i < len(tt.sentOrder); i++ {
					assert.Less(t, strings.Index(all, tt.sentOrder[i-1]), strings.Index(all, tt.sentOrder[i]),
						"the summarizer sees the older pair before the newer one")
				}
			}
			if tt.body == nil {
				assert.Equal(t, before, chainSummaryJSON(t, ast), "the chain should be untouched")
				return
			}
			chainSummaryCheckBody(t, ast.Sections[last].Body, original, tt.body, tt.kind, tt.signed, tt.lead...)
			chainSummaryVerifyAST(t, ast)
		})
	}
}

func TestChainSummary_SummarizeOversizedBodyPairs_CarriesTheReasoningOfASummarizedToolCall(t *testing.T) {
	oversized := strings.Repeat("X", 20*1024)
	thought := chainSummaryThought("Let me analyze the wp-abilities plugin...",
		"The wp-abilities plugin seems to be the main target here. Need to find vulnerabilities.")
	signed := chainSummarySignature("original_gemini_signature_12345")

	tests := []struct {
		name   string
		body   []*cast.BodyPair
		signed bool
		lead   []llms.ContentPart
	}{
		{
			name: "a signed tool call is summarized with a fake signature",
			body: []*cast.BodyPair{
				chainSummaryToolPair("call_test123", "get_data", oversized, signed),
				cast.NewBodyPairFromCompletion("This is a normal response"),
			},
			signed: true,
		},
		{
			name: "an unsigned tool call is summarized without a signature",
			body: []*cast.BodyPair{
				chainSummaryToolPair("call_test456", "get_info", oversized, nil),
				cast.NewBodyPairFromCompletion("This is a normal response"),
			},
		},
		{
			name: "reasoning text is carried before the summary call",
			body: []*cast.BodyPair{
				chainSummaryToolPair("call_kimi_test", "search", oversized, nil, thought),
				cast.NewBodyPairFromCompletion("This is the final response"),
			},
			lead: []llms.ContentPart{thought},
		},
		{
			name: "reasoning text and a signed call keep both",
			body: []*cast.BodyPair{
				chainSummaryToolPair("call_kimi_full", "exploit", oversized, signed, thought),
				cast.NewBodyPairFromCompletion("Final response"),
			},
			signed: true,
			lead:   []llms.ContentPart{thought},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			section := chainSummarySection("Test question", tt.body...)
			original := chainSummaryPairsJSON(t, section.Body)
			rec := &chainSummaryRecorder{}

			err := summarizeOversizedBodyPairs(t.Context(), section, rec.handle, 16*1024, cast.ToolCallIDTemplate)
			require.NoError(t, err)
			assert.Len(t, rec.sent(), 1, "summarizer calls")
			chainSummaryCheckBody(t, section.Body, original, []int{chainSummaryNew, 1}, cast.Summarization, tt.signed, tt.lead...)
			chainSummaryVerifyAST(t, &cast.ChainAST{Sections: []*cast.ChainSection{section}})
		})
	}
}

func TestChainSummary_SummarizeQAPairs_CollapsesTheOldestSectionsIntoOne(t *testing.T) {
	prior := func() *cast.BodyPair {
		return cast.NewBodyPairFromSummarization("Prior summary", cast.ToolCallIDTemplate, false, nil)
	}

	tests := []struct {
		name        string
		sections    []*cast.ChainSection
		maxSections int
		maxBytes    int
		human       bool
		fail        bool
		calls       int
		sent        []string
		notSent     []string
		questions   []string // the human parts of the summary section; nil: the chain is left untouched
		kind        cast.BodyPairType
		wantErr     string
	}{
		{name: "an empty chain is left alone", maxSections: 5, maxBytes: 2000},
		{
			name: "a chain exactly at the section limit and the byte budget is left alone",
			sections: []*cast.ChainSection{
				chainSummaryFirst(chainSummaryQA("Question 1", "Answer 1")),
				chainSummaryQA("Question 2", "Answer 2"),
			},
			maxSections: 2,
			maxBytes:    50,
		},
		{
			name: "tool exchanges collapse into a summarization call",
			sections: []*cast.ChainSection{
				chainSummaryFirst(chainSummarySection("Question 1", chainSummaryToolPair("call_1", "search", "Result 1", nil))),
				chainSummarySection("Question 2", chainSummaryToolPair("call_2", "search", "Result 2", nil)),
				chainSummaryQA("Question 3", "Answer 3"),
			},
			maxSections: 1,
			maxBytes:    10000,
			calls:       1,
			sent:        []string{"Result 1", "Result 2"},
			notSent:     []string{"Answer 3"},
			questions:   []string{"Question 1", "Question 2"},
			kind:        cast.Summarization,
		},
		{
			name: "a lone older section is summarized when it is not yet a summary",
			sections: []*cast.ChainSection{
				chainSummaryFirst(chainSummaryQA("Question 1", "Answer 1")),
				chainSummaryQA("Question 2", "Answer 2"),
			},
			maxSections: 1,
			maxBytes:    10000,
			calls:       1,
			sent:        []string{"Answer 1"},
			notSent:     []string{"Answer 2"},
			questions:   []string{"Question 1"},
			kind:        cast.Completion,
		},
		{
			name: "an oldest summary collapses together with a newer section",
			sections: []*cast.ChainSection{
				chainSummaryFirst(chainSummarySection("Question 1", prior())),
				chainSummaryQA("Question 2", "Answer 2"),
				chainSummaryQA("Question 3", "Answer 3"),
			},
			maxSections: 1,
			maxBytes:    10000,
			calls:       1,
			sent:        []string{"Prior summary", "Answer 2"},
			notSent:     []string{"Answer 3"},
			questions:   []string{"Question 1", "Question 2"},
			kind:        cast.Summarization,
		},
		{
			name: "a lone older section opening with a summary is collapsed again",
			sections: []*cast.ChainSection{
				chainSummaryFirst(chainSummarySection("Question 1", prior(), cast.NewBodyPairFromCompletion("Answer 1b"))),
				chainSummaryQA("Question 2", "Answer 2"),
			},
			maxSections: 1,
			maxBytes:    10000,
			calls:       1,
			sent:        []string{"Prior summary", "Answer 1b"},
			notSent:     []string{"Answer 2"},
			questions:   []string{"Question 1"},
			kind:        cast.Summarization,
		},
		{
			name: "a failed question summary is reported",
			sections: []*cast.ChainSection{
				chainSummaryFirst(chainSummaryQA("Question 1", "Answer 1")),
				chainSummaryQA("Question 2", "Answer 2"),
				chainSummaryQA("Question 3", "Answer 3"),
			},
			maxSections: 1,
			maxBytes:    10000,
			human:       true,
			fail:        true,
			calls:       1,
			wantErr:     "QA (human) summary generation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast := &cast.ChainAST{Sections: tt.sections}
			before := chainSummaryJSON(t, ast)
			var newest string
			if len(tt.sections) > 0 {
				newest = chainSummaryJSON(t, tt.sections[len(tt.sections)-1])
			}
			rec := &chainSummaryRecorder{fail: tt.fail, numbered: true}

			err := summarizeQAPairs(t.Context(), ast, rec.handle, 1, tt.maxSections, tt.maxBytes, tt.human, cast.ToolCallIDTemplate)
			require.Len(t, rec.sent(), tt.calls, "summarizer calls")
			chainSummaryAssertSent(t, rec.sent(), tt.sent, tt.notSent)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Equal(t, before, chainSummaryJSON(t, ast), "a failed summary should leave the chain untouched")
				return
			}
			require.NoError(t, err)
			if tt.questions == nil {
				assert.Equal(t, before, chainSummaryJSON(t, ast), "the chain should be untouched")
				return
			}

			require.Len(t, ast.Sections, 2)
			chainSummaryQAHead(tt.kind, "SUMMARY-001", tt.questions...)(t, ast, nil)
			assert.Equal(t, newest, chainSummaryJSON(t, ast.Sections[1]), "the newest section is kept as it was")
			chainSummaryVerifyAST(t, ast)
		})
	}
}

func TestChainSummary_SummarizeQAPairs_KeepsRecentSectionsWithinTheByteBudget(t *testing.T) {
	sized := func(human string, size int) *cast.ChainSection {
		return chainSummaryQA(human, strings.Repeat("x", size-len(human)))
	}

	tests := []struct {
		name         string
		sizes        []int // section sizes oldest first, the last one is the newest kept section
		maxSections  int
		maxBytes     int
		wantSections int
		keptTail     int
	}{
		{
			name:         "equal sections are kept until the byte budget is met",
			sizes:        []int{2000, 2000, 2000, 2000, 2000},
			maxSections:  100,
			maxBytes:     7000,
			wantSections: 4,
			keptTail:     3,
		},
		{
			name:         "a section too large to keep does not strand the older ones",
			sizes:        []int{500, 500, 4000, 500, 500},
			maxSections:  100,
			maxBytes:     5000,
			wantSections: 3,
			keptTail:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections := make([]*cast.ChainSection, len(tt.sizes))
			for i, size := range tt.sizes {
				sections[i] = sized(fmt.Sprintf("Q%d", i), size)
			}
			ast := &cast.ChainAST{Sections: sections}
			kept := make([]string, tt.keptTail)
			for i := range tt.keptTail {
				kept[i] = chainSummaryJSON(t, sections[len(sections)-tt.keptTail+i])
			}
			rec := &chainSummaryRecorder{}

			err := summarizeQAPairs(t.Context(), ast, rec.handle, 1, tt.maxSections, tt.maxBytes, false, cast.ToolCallIDTemplate)
			require.NoError(t, err)
			assert.Len(t, rec.sent(), 1, "summarizer calls")
			require.Len(t, ast.Sections, tt.wantSections)
			assert.True(t, chainSummaryIsSummary(ast.Sections[0].Body[0]), "the oldest sections collapse into a summary")
			for i := range tt.keptTail {
				assert.Equal(t, kept[i], chainSummaryJSON(t, ast.Sections[tt.wantSections-tt.keptTail+i]), "kept tail section %d", i)
			}
			chainSummaryVerifyAST(t, ast)
		})
	}
}

func TestChainSummary_GenerateSummary_BuildsTheExactPrompt(t *testing.T) {
	image := llms.ImageURLContent{URL: "http://img", Detail: "a cat"}
	binary := llms.BinaryContent{MIMEType: "application/pdf", Data: []byte{0xde, 0xad, 0xbe, 0xef}}
	human := llms.MessageContent{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{
		llms.TextContent{Text: "task text"}, image, binary,
	}}
	aiTurn := llms.MessageContent{Role: llms.ChatMessageTypeAI, Parts: []llms.ContentPart{
		llms.TextContent{Text: "ai text"},
		llms.ToolCall{ID: "call_x", Type: "function", FunctionCall: &llms.FunctionCall{Name: "terminal", Arguments: `{"cmd":"ls"}`}},
		image, binary,
	}}
	toolTurn := llms.MessageContent{Role: llms.ChatMessageTypeTool, Parts: []llms.ContentPart{
		llms.ToolCallResponse{ToolCallID: "call_x", Name: "terminal", Content: "file listing"},
	}}

	tasks := "<tasks>\n<task id=\"0\">\ntask text\n" +
		"<image url=\"http://img\">\na cat\n</image>\n" +
		"<binary mime=\"application/pdf\">\nfirst 100 bytes in hex: deadbeef\n</binary>\n" +
		"</task>\n</tasks>\n"
	messages := "<messages>\n" +
		"<message id=\"0\" role=\"ai\">\n" +
		"<content part=\"0\">\nai text\n</content>\n" +
		"<tool_call name=\"terminal\" part=\"1\">\n{\"cmd\":\"ls\"}\n</tool_call>\n" +
		"<image url=\"http://img\" part=\"2\">\na cat\n</image>\n" +
		"<binary mime=\"application/pdf\" part=\"3\">\nfirst 100 bytes in hex: deadbeef\n</binary>\n" +
		"</message>\n" +
		"<message id=\"1\" role=\"tool\">\n" +
		"<tool_call_response name=\"terminal\" part=\"0\">\nfile listing\n</tool_call_response>\n" +
		"</message>\n" +
		"</messages>"
	withContext := "SUMMARIZATION TASK: Create a concise summary of AI responses while preserving essential information from the conversation context."
	standalone := "SUMMARIZATION TASK: Distill standalone AI responses into a comprehensive yet concise summary."
	queries := "SUMMARIZATION TASK: Extract key requirements and context from user queries."
	priorSummaries := "1. A message contains <tool_call name=\"execute_task_and_return_summary\">"

	tests := []struct {
		name         string
		human        []llms.MessageContent
		ai           []llms.MessageContent
		instructions []string // the first line of the instructions, then lines they must also hold
		wantData     string
		wantErr      string
	}{
		{
			name:         "human context around ai messages",
			human:        []llms.MessageContent{human},
			ai:           []llms.MessageContent{aiTurn, toolTurn},
			instructions: []string{withContext, priorSummaries},
			wantData:     tasks + messages,
		},
		{
			name:         "ai messages alone",
			ai:           []llms.MessageContent{aiTurn, toolTurn},
			instructions: []string{standalone, priorSummaries},
			wantData:     messages,
		},
		{name: "human messages alone", human: []llms.MessageContent{human}, instructions: []string{queries}, wantData: tasks},
		{
			name: "a non-human message in the task list is skipped",
			human: []llms.MessageContent{
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "first"}}},
				{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextContent{Text: "sys"}}},
				{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "second"}}},
			},
			instructions: []string{queries},
			wantData:     "<tasks>\n<task id=\"0\">\nfirst\n</task>\n<task id=\"2\">\nsecond\n</task>\n</tasks>\n",
		},
		{name: "no messages at all is refused", wantErr: "cannot summarize empty message list"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &chainSummaryRecorder{}
			got, err := GenerateSummary(t.Context(), rec.handle, tt.human, tt.ai)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Empty(t, rec.sent(), "a refused summary should not reach the summarizer")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, chainSummaryReply, got)
			require.Len(t, rec.sent(), 1)
			prompt := rec.sent()[0]
			assert.True(t, strings.HasPrefix(prompt, "<instructions>\n"+tt.instructions[0]+"\n"), "instructions open with %q", tt.instructions[0])
			instructions, _, _ := strings.Cut(prompt, "</instructions>")
			for _, line := range tt.instructions[1:] {
				assert.Contains(t, instructions, line)
			}
			assert.Equal(t, tt.wantData, chainSummaryPromptData(t, prompt))
		})
	}

	t.Run("a nil handler is refused", func(t *testing.T) {
		_, err := GenerateSummary(t.Context(), nil, []llms.MessageContent{human}, nil)
		require.ErrorContains(t, err, "summarizer handler cannot be nil")
	})
}

func TestChainSummary_SummarizeChain_RepairsADanglingToolCall(t *testing.T) {
	dangling := cast.NewBodyPair(&llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{llms.ToolCall{
			ID: "call_unanswered", Type: "function",
			FunctionCall: &llms.FunctionCall{Name: "terminal", Arguments: `{"cmd":"ls"}`},
		}},
	}, nil)
	in := (&cast.ChainAST{Sections: []*cast.ChainSection{chainSummarySection("Question", dangling)}}).Messages()
	rec := &chainSummaryRecorder{}

	out, err := NewSummarizer(SummarizerConfig{}).SummarizeChain(t.Context(), rec.handle, in, cast.ToolCallIDTemplate)
	require.NoError(t, err)
	assert.Empty(t, rec.sent(), "summarizer calls")
	require.Len(t, out, 3)
	assert.Equal(t, chainSummaryJSON(t, in), chainSummaryJSON(t, out[:2]), "the question and the call are kept")
	want := llms.MessageContent{Role: llms.ChatMessageTypeTool, Parts: []llms.ContentPart{llms.ToolCallResponse{
		ToolCallID: "call_unanswered", Name: "terminal", Content: "the call was not handled, please try again",
	}}}
	assert.Equal(t, want, out[2], "the unanswered tool call gains a fallback response")
}
