package tester

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/providers/tester/mock"
	"pentagi/pkg/providers/tester/testdata"
	"pentagi/pkg/tools"

	"github.com/vxcontrol/langchaingo/llms"
)

// fileToolCallResponse builds a *llms.ContentResponse whose single choice
// calls PentAGI's file tool with the given arguments (empty strings are
// omitted, mirroring how a real model would only send fields it needs).
func fileToolCallResponse(id string, args map[string]string) *llms.ContentResponse {
	argMap := make(map[string]any, len(args))
	for k, v := range args {
		if v != "" {
			argMap[k] = v
		}
	}
	argsJSON, _ := json.Marshal(argMap)

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{
			ToolCalls: []llms.ToolCall{{
				ID:   id,
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      tools.FileToolName,
					Arguments: string(argsJSON),
				},
			}},
		}},
	}
}

// correctFileEditDiff is a unified diff that turns FileEditTestOldLine into
// FileEditTestNewLine in FileEditTestContent (Priority is its 3rd line).
const correctFileEditDiff = "@@ -3,1 +3,1 @@\n-" + FileEditTestOldLine + "\n+" + FileEditTestNewLine + "\n"

func TestNewFileEditTestCase(t *testing.T) {
	t.Parallel()

	tc, err := newFileEditTestCase()
	if err != nil {
		t.Fatalf("newFileEditTestCase() error = %v", err)
	}

	if tc.ID() == "" || tc.Name() == "" {
		t.Error("expected non-empty ID and Name")
	}
	if tc.Group() != testdata.TestGroupAdvanced {
		t.Errorf("Group() = %q, want %q", tc.Group(), testdata.TestGroupAdvanced)
	}
	if tc.Type() != testdata.TestTypeFileEdit {
		t.Errorf("Type() = %q, want %q", tc.Type(), testdata.TestTypeFileEdit)
	}
	if tc.Capability() != testdata.CapabilityNone {
		t.Errorf("Capability() = %q, want CapabilityNone", tc.Capability())
	}

	// the tool declaration must be PentAGI's real one, not a hand-copied schema
	toolList := tc.Tools()
	if len(toolList) != 1 || toolList[0].Function == nil || toolList[0].Function.Name != tools.FileToolName {
		t.Fatalf("expected exactly the %q tool declaration, got %+v", tools.FileToolName, toolList)
	}
	realDef := tools.GetRegistryDefinitions()[tools.FileToolName]
	if toolList[0].Function.Description != realDef.Description {
		t.Error("tool description does not match tools.GetRegistryDefinitions() - declarations must come from pkg/tools")
	}

	if len(tc.Messages()) != 1 {
		t.Fatalf("expected exactly 1 initial message, got %d", len(tc.Messages()))
	}

	if _, ok := tc.(testdata.MultiTurnTestCase); !ok {
		t.Fatal("fileEditTestCase must implement testdata.MultiTurnTestCase")
	}
}

// TestFileEditTestCase_HandleToolResponse covers the read_file -> edit_file
// exchange with both a positive (correct sequence and diff) and negative
// (wrong tool, wrong action, wrong path, wrong order, bad diff, diff that
// applies but doesn't produce the requested change) cases.
func TestFileEditTestCase_HandleToolResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		firstResp   *llms.ContentResponse
		secondResp  *llms.ContentResponse // nil: the exchange must end after firstResp
		wantSuccess bool
		wantErr     string
	}{
		{
			name:        "happy path: read_file then edit_file with a correct diff",
			firstResp:   fileToolCallResponse("c1", map[string]string{"action": "read_file", "path": FileEditTestPath}),
			secondResp:  fileToolCallResponse("c2", map[string]string{"action": "edit_file", "path": FileEditTestPath, "diff": correctFileEditDiff}),
			wantSuccess: true,
		},
		{
			name:        "action omitted on first call still infers read_file",
			firstResp:   fileToolCallResponse("c1", map[string]string{"path": FileEditTestPath}),
			secondResp:  fileToolCallResponse("c2", map[string]string{"action": "edit_file", "path": FileEditTestPath, "diff": correctFileEditDiff}),
			wantSuccess: true,
		},
		{
			name:      "no tool call at all",
			firstResp: &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "I can't help with that."}}},
			wantErr:   `did not call the "file" tool`,
		},
		{
			name:      "calls a different tool entirely",
			firstResp: &llms.ContentResponse{Choices: []*llms.ContentChoice{{ToolCalls: []llms.ToolCall{{FunctionCall: &llms.FunctionCall{Name: "terminal", Arguments: `{}`}}}}}},
			wantErr:   `did not call the "file" tool`,
		},
		{
			name:      "first call is write_file instead of read_file",
			firstResp: fileToolCallResponse("c1", map[string]string{"action": "write_file", "path": FileEditTestPath, "content": "whatever"}),
			wantErr:   "expected the first call to be read_file",
		},
		{
			name:      "read_file targets the wrong path",
			firstResp: fileToolCallResponse("c1", map[string]string{"action": "read_file", "path": "/etc/passwd"}),
			wantErr:   "expected read_file to target",
		},
		{
			name:       "second call is write_file instead of edit_file",
			firstResp:  fileToolCallResponse("c1", map[string]string{"action": "read_file", "path": FileEditTestPath}),
			secondResp: fileToolCallResponse("c2", map[string]string{"action": "write_file", "path": FileEditTestPath, "content": "whatever"}),
			wantErr:    "expected the second call to be edit_file",
		},
		{
			name:       "edit_file targets the wrong path",
			firstResp:  fileToolCallResponse("c1", map[string]string{"action": "read_file", "path": FileEditTestPath}),
			secondResp: fileToolCallResponse("c2", map[string]string{"action": "edit_file", "path": "/etc/passwd", "diff": correctFileEditDiff}),
			wantErr:    "expected edit_file to target",
		},
		{
			name:       "edit_file diff doesn't match the file content",
			firstResp:  fileToolCallResponse("c1", map[string]string{"action": "read_file", "path": FileEditTestPath}),
			secondResp: fileToolCallResponse("c2", map[string]string{"action": "edit_file", "path": FileEditTestPath, "diff": "@@ -1,1 +1,1 @@\n-this line does not exist\n+replacement\n"}),
			wantErr:    "did not apply",
		},
		{
			name:       "edit_file diff applies but produces the wrong content",
			firstResp:  fileToolCallResponse("c1", map[string]string{"action": "read_file", "path": FileEditTestPath}),
			secondResp: fileToolCallResponse("c2", map[string]string{"action": "edit_file", "path": FileEditTestPath, "diff": "@@ -1,1 +1,1 @@\n-Status: draft\n+Status: final\n"}),
			wantErr:    "did not produce",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tc, err := newFileEditTestCase()
			if err != nil {
				t.Fatalf("newFileEditTestCase() error = %v", err)
			}
			mt := tc.(testdata.MultiTurnTestCase)

			finalResp := tt.firstResp
			hasMore := mt.HandleToolResponse(tt.firstResp)

			if tt.secondResp != nil {
				if !hasMore {
					t.Fatal("expected HandleToolResponse to request a second round after the first response")
				}
				finalResp = tt.secondResp
				hasMore = mt.HandleToolResponse(tt.secondResp)
			}
			if hasMore {
				t.Fatal("expected the exchange to end, but HandleToolResponse asked for another round")
			}

			result := tc.Execute(finalResp, 5*time.Millisecond)

			if tt.wantSuccess {
				if !result.Success {
					t.Fatalf("expected success, got error: %v", result.Error)
				}
				return
			}

			if result.Success {
				t.Fatal("expected failure, got success")
			}
			if result.Error == nil || !strings.Contains(result.Error.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want it to contain %q", result.Error, tt.wantErr)
			}
		})
	}
}

// TestFileEditTestCase_MessagesAccumulate checks that each answered tool
// call is appended to the conversation, so the model's next Messages() call
// actually sees the growing history (assistant tool-call + tool result per turn).
func TestFileEditTestCase_MessagesAccumulate(t *testing.T) {
	t.Parallel()

	tc, err := newFileEditTestCase()
	if err != nil {
		t.Fatalf("newFileEditTestCase() error = %v", err)
	}
	mt := tc.(testdata.MultiTurnTestCase)

	if got := len(tc.Messages()); got != 1 {
		t.Fatalf("expected 1 message before any turn, got %d", got)
	}

	mt.HandleToolResponse(fileToolCallResponse("c1", map[string]string{"action": "read_file", "path": FileEditTestPath}))
	if got := len(tc.Messages()); got != 3 {
		t.Fatalf("expected 3 messages after round 1 (initial + assistant tool-call + tool result), got %d", got)
	}

	mt.HandleToolResponse(fileToolCallResponse("c2", map[string]string{"action": "edit_file", "path": FileEditTestPath, "diff": correctFileEditDiff}))
	if got := len(tc.Messages()); got != 5 {
		t.Fatalf("expected 5 messages after round 2, got %d", got)
	}
}

// TestFileEditTestCase_Integration_ViaExecuteTest drives the scenario
// through tester.executeTest itself (not just the TestCase's own methods),
// using a mock provider configured to answer in sequence - proving the
// runner's multi-turn loop (see executeTest's MultiTurnTestCase branch)
// actually re-calls the provider with the growing conversation and stops at
// the right point.
func TestFileEditTestCase_Integration_ViaExecuteTest(t *testing.T) {
	tc, err := newFileEditTestCase()
	if err != nil {
		t.Fatalf("newFileEditTestCase() error = %v", err)
	}

	mockProvider := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	mockProvider.SetSequentialResponses(
		fileToolCallResponse("call_1", map[string]string{"action": "read_file", "path": FileEditTestPath}),
		fileToolCallResponse("call_2", map[string]string{"action": "edit_file", "path": FileEditTestPath, "diff": correctFileEditDiff}),
	)

	result, err := executeTest(t.Context(), testRequest{
		agentType: pconfig.OptionsTypePrimaryAgent,
		testCase:  tc,
		provider:  mockProvider,
	})
	if err != nil {
		t.Fatalf("executeTest() error = %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got error: %v", result.Error)
	}
	if result.Type != testdata.TestTypeFileEdit {
		t.Errorf("result.Type = %q, want %q", result.Type, testdata.TestTypeFileEdit)
	}
}

// TestCollectTestRequests_FileEditTest_GetsIndependentInstancePerAgent is a
// regression test for a real bug found by running ctester live against
// Ollama: collectTestRequests used to build ONE fileEditTestCase per group
// and reuse that same pointer for every agentType's testRequest. Every other
// TestCase is immutable once built, so sharing is harmless there, but
// fileEditTestCase mutates itself turn by turn (see HandleToolResponse) -
// sharing it meant one agent's conversation, and even its pass/fail outcome
// (f.failure persists until overwritten), leaked into the next agent's
// "independent" run. Each agentType must get its own instance.
func TestCollectTestRequests_FileEditTest_GetsIndependentInstancePerAgent(t *testing.T) {
	emptyRegistry, err := testdata.LoadRegistryFromYAML([]byte("[]"))
	if err != nil {
		t.Fatalf("LoadRegistryFromYAML() error = %v", err)
	}

	config := &testConfig{
		agentTypes:     []pconfig.ProviderOptionsType{pconfig.OptionsTypePrimaryAgent, pconfig.OptionsTypeCoder, pconfig.OptionsTypePentester},
		groups:         []testdata.TestGroup{testdata.TestGroupAdvanced},
		customRegistry: emptyRegistry,
	}

	requests := collectTestRequests(emptyRegistry, mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model"), config)

	if len(requests) != len(config.agentTypes) {
		t.Fatalf("expected %d file_edit requests (one per agent type), got %d", len(config.agentTypes), len(requests))
	}

	seen := make(map[testdata.TestCase]pconfig.ProviderOptionsType, len(requests))
	for _, req := range requests {
		if req.testCase.Type() != testdata.TestTypeFileEdit {
			t.Fatalf("expected only file_edit requests, got type %q", req.testCase.Type())
		}
		if owner, dup := seen[req.testCase]; dup {
			t.Fatalf("agent types %q and %q were given the SAME fileEditTestCase instance - state from one run would leak into the other", owner, req.agentType)
		}
		seen[req.testCase] = req.agentType

		if _, ok := req.testCase.(testdata.MultiTurnTestCase); !ok {
			t.Fatalf("file_edit test case for %q does not implement MultiTurnTestCase", req.agentType)
		}
	}
}

// TestFileEditTestCase_Integration_StopsAfterOneRoundOnFailure checks that a
// bad first call doesn't spend a second provider request: the mock's
// sequential responses only ever get consumed once here.
func TestFileEditTestCase_Integration_StopsAfterOneRoundOnFailure(t *testing.T) {
	tc, err := newFileEditTestCase()
	if err != nil {
		t.Fatalf("newFileEditTestCase() error = %v", err)
	}

	mockProvider := mock.NewProvider(provider.ProviderCustom, provider.DefaultProviderNameCustom, "test-model")
	mockProvider.SetSequentialResponses(
		fileToolCallResponse("call_1", map[string]string{"action": "write_file", "path": FileEditTestPath, "content": "oops"}),
		fileToolCallResponse("call_2", map[string]string{"action": "edit_file", "path": FileEditTestPath, "diff": correctFileEditDiff}),
	)

	result, err := executeTest(t.Context(), testRequest{
		agentType: pconfig.OptionsTypePrimaryAgent,
		testCase:  tc,
		provider:  mockProvider,
	})
	if err != nil {
		t.Fatalf("executeTest() error = %v", err)
	}
	if result.Success {
		t.Fatal("expected failure since the model called write_file instead of read_file first")
	}
}
