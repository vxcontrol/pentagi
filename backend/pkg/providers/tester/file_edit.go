package tester

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"pentagi/pkg/providers/tester/testdata"
	"pentagi/pkg/tools"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

// The fixture fileEditTestCase's handler simulates: read_file for
// FileEditTestPath must return FileEditTestContent, and edit_file's diff
// must turn FileEditTestOldLine into FileEditTestNewLine somewhere in it.
// Exported so callers driving this scenario end to end through a mock
// provider (e.g. cmd/ctester's own tests) can build a matching response
// sequence without duplicating the fixture.
const (
	FileEditTestPath    = "/work/report.txt"
	FileEditTestContent = "Status: draft\nOwner: alice\nPriority: low\n"
	FileEditTestOldLine = "Priority: low"
	FileEditTestNewLine = "Priority: high"

	fileEditID       = "file_read_then_edit"
	fileEditTestName = "Read a file, then edit it via unified diff"
)

// newFileEditTestCase builds a two-turn scenario exercising PentAGI's real
// file tool end to end: read a file, then change one line in it. Unlike
// every YAML-driven TestCase, this one is hand-built in Go (not tests.yml)
// because its whole point is a genuinely dynamic exchange - the harness
// doesn't know what the model will send until it sends it, so it can't be
// expressed as a fixed list of messages the way testdata.TestDefinition is.
//
// The tool declaration is not hand-copied JSON: it comes straight from
// tools.GetRegistryDefinitions(), the exact schema real PentAGI agents get,
// so this test also catches accidental schema regressions in pkg/tools. The
// read/edit "handler" below is this package's own in-memory simulation (no
// Docker container is involved), deliberately separate from - but for the
// actual diff-merge step, reusing - tools.ApplyUnifiedDiff, the same
// function terminal.EditFile calls in production.
func newFileEditTestCase() (testdata.TestCase, error) {
	def, ok := tools.GetRegistryDefinitions()[tools.FileToolName]
	if !ok {
		return nil, fmt.Errorf("tools.GetRegistryDefinitions() has no definition for %q", tools.FileToolName)
	}
	defCopy := def

	prompt := fmt.Sprintf(
		"Read the file at %q. Then change the line %q to exactly %q and save the file.",
		FileEditTestPath, FileEditTestOldLine, FileEditTestNewLine,
	)

	return &fileEditTestCase{
		messages: []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, prompt)},
		tools: []llms.Tool{{
			Type:     "function",
			Function: &defCopy,
		}},
	}, nil
}

// fileEditTestCase implements testdata.TestCase and testdata.MultiTurnTestCase.
type fileEditTestCase struct {
	mu       sync.Mutex
	messages []llms.MessageContent
	tools    []llms.Tool

	readFileSeen bool // the model has made its first (expected read_file) call
	editFileSeen bool // the model has made its second (expected edit_file) call
	editApplied  bool // edit_file's diff actually produced the requested change
	failure      string
}

func (f *fileEditTestCase) ID() string                            { return fileEditID }
func (f *fileEditTestCase) Name() string                          { return fileEditTestName }
func (f *fileEditTestCase) Type() testdata.TestType               { return testdata.TestTypeFileEdit }
func (f *fileEditTestCase) Group() testdata.TestGroup             { return testdata.TestGroupAdvanced }
func (f *fileEditTestCase) Streaming() bool                       { return false }
func (f *fileEditTestCase) Prompt() string                        { return "" }
func (f *fileEditTestCase) Tools() []llms.Tool                    { return f.tools }
func (f *fileEditTestCase) Capability() testdata.TestCapability   { return testdata.CapabilityNone }
func (f *fileEditTestCase) ExtraOptions() []llms.CallOption       { return nil }
func (f *fileEditTestCase) StreamingCallback() streaming.Callback { return nil }

func (f *fileEditTestCase) Messages() []llms.MessageContent {
	f.mu.Lock()
	defer f.mu.Unlock()

	out := make([]llms.MessageContent, len(f.messages))
	copy(out, f.messages)
	return out
}

// HandleToolResponse implements testdata.MultiTurnTestCase. It expects,
// across up to two calls, first a read_file call for FileEditTestPath
// (answered with FileEditTestContent) and then an edit_file call whose diff is applied
// in-memory via tools.ApplyUnifiedDiff; any other shape ends the exchange
// immediately with a recorded failure reason for Execute to report.
func (f *fileEditTestCase) HandleToolResponse(resp *llms.ContentResponse) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	call, args, ok := firstFileToolCall(resp)
	if !ok {
		f.failure = fmt.Sprintf("model did not call the %q tool", tools.FileToolName)
		return false
	}

	action, _ := args["action"].(string)
	path, _ := args["path"].(string)

	if !f.readFileSeen {
		f.readFileSeen = true

		if action != "" && action != string(tools.ReadFile) {
			f.failure = fmt.Sprintf("expected the first call to be read_file, got action=%q", action)
			return false
		}
		if path != FileEditTestPath {
			f.failure = fmt.Sprintf("expected read_file to target %q, got %q", FileEditTestPath, path)
			return false
		}

		f.appendToolExchange(call, FileEditTestContent)
		return true
	}

	f.editFileSeen = true

	if action != string(tools.EditFile) {
		f.failure = fmt.Sprintf("expected the second call to be edit_file, got action=%q", action)
		f.appendToolExchange(call, "unexpected action for this scenario; ending the test")
		return false
	}
	if path != FileEditTestPath {
		f.failure = fmt.Sprintf("expected edit_file to target %q, got %q", FileEditTestPath, path)
		return false
	}

	diff, _ := args["diff"].(string)
	newContent, _, err := tools.ApplyUnifiedDiff(FileEditTestContent, diff)
	if err != nil {
		f.failure = fmt.Sprintf("edit_file's diff did not apply: %v", err)
		f.appendToolExchange(call, fmt.Sprintf("failed to apply diff: %v", err))
		return false
	}

	f.editApplied = strings.Contains(newContent, FileEditTestNewLine) && !strings.Contains(newContent, FileEditTestOldLine)
	if !f.editApplied {
		f.failure = fmt.Sprintf("edit_file's diff applied but did not produce %q (result: %q)", FileEditTestNewLine, newContent)
	}
	f.appendToolExchange(call, fmt.Sprintf("Applied 1 diff hunk(s) to %s", FileEditTestPath))
	return false
}

// appendToolExchange records the assistant's tool call and a synthesized
// tool result, so the next Messages() call reflects them for the model's
// next turn. Caller must hold f.mu.
func (f *fileEditTestCase) appendToolExchange(call llms.ToolCall, result string) {
	f.messages = append(f.messages,
		llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{call},
		},
		llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: call.ID,
					Name:       call.FunctionCall.Name,
					Content:    result,
				},
			},
		},
	)
}

// Execute implements testdata.TestCase. By the time the runner calls this,
// HandleToolResponse has already driven the exchange to completion (or to
// the point where it gave up); Execute only needs to report that outcome.
func (f *fileEditTestCase) Execute(response any, latency time.Duration) testdata.TestResult {
	f.mu.Lock()
	defer f.mu.Unlock()

	result := testdata.TestResult{
		ID:      f.ID(),
		Name:    f.Name(),
		Type:    f.Type(),
		Group:   f.Group(),
		Latency: latency,
	}

	if _, ok := response.(*llms.ContentResponse); !ok {
		result.Error = fmt.Errorf("expected *llms.ContentResponse, got %T", response)
		return result
	}

	switch {
	case f.failure != "":
		result.Error = fmt.Errorf("%s", f.failure)
	case !f.readFileSeen:
		result.Error = fmt.Errorf("model never called the %q tool", tools.FileToolName)
	case !f.editFileSeen:
		result.Error = fmt.Errorf("model called read_file but never followed up with edit_file")
	case !f.editApplied:
		result.Error = fmt.Errorf("edit_file was called but did not produce the requested change")
	default:
		result.Success = true
	}

	return result
}

// firstFileToolCall returns the first tool call in resp targeting PentAGI's
// file tool (tools.FileToolName) along with its decoded arguments.
func firstFileToolCall(resp *llms.ContentResponse) (llms.ToolCall, map[string]any, bool) {
	for _, choice := range resp.Choices {
		for _, call := range choice.ToolCalls {
			if call.FunctionCall == nil || call.FunctionCall.Name != tools.FileToolName {
				continue
			}

			var args map[string]any
			if err := json.Unmarshal([]byte(call.FunctionCall.Arguments), &args); err != nil {
				continue
			}

			return call, args, true
		}
	}
	return llms.ToolCall{}, nil, false
}
