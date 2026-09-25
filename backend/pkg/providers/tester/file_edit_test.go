package tester

import (
	"strings"
	"testing"
	"time"

	"pentagi/pkg/providers/tester/cases"
	"pentagi/pkg/tools"

	"github.com/vxcontrol/langchaingo/llms"
)

func TestFileEdit_NewFileEditTestCase_DeclaresPentAGIsOwnFileTool(t *testing.T) {
	t.Parallel()

	tc, err := newFileEditTestCase()
	if err != nil {
		t.Fatalf("newFileEditTestCase() error = %v", err)
	}

	if tc.ID() == "" || tc.Name() == "" {
		t.Error("expected non-empty ID and Name")
	}
	if tc.Group() != cases.TestGroupAdvanced {
		t.Errorf("Group() = %q, want %q", tc.Group(), cases.TestGroupAdvanced)
	}
	if tc.Type() != cases.TestTypeFileEdit {
		t.Errorf("Type() = %q, want %q", tc.Type(), cases.TestTypeFileEdit)
	}
	if tc.Capability() != cases.CapabilityNone {
		t.Errorf("Capability() = %q, want CapabilityNone", tc.Capability())
	}

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

	if _, ok := tc.(cases.MultiTurnTestCase); !ok {
		t.Fatal("fileEditTestCase must implement cases.MultiTurnTestCase")
	}
}

func TestFileEdit_HandleToolResponse_JudgesTheReadThenEditExchange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		firstResp   *llms.ContentResponse
		secondResp  *llms.ContentResponse // nil: the exchange must end after firstResp
		wantSuccess bool
		wantErr     string
	}{
		{
			name:        "read_file then edit_file with a correct diff",
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
			mt := tc.(cases.MultiTurnTestCase)

			finalResp := tt.firstResp
			hasMore := mt.HandleToolResponse(tt.firstResp)

			if tt.secondResp != nil {
				if !hasMore {
					t.Fatal("expected HandleToolResponse to request a second round after the first response")
				}
				// the initial prompt, the assistant's tool call and the tool result
				if got := len(tc.Messages()); got != 3 {
					t.Fatalf("expected 3 messages after round 1, got %d", got)
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
				if got := len(tc.Messages()); got != 5 {
					t.Fatalf("expected 5 messages after round 2, got %d", got)
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
