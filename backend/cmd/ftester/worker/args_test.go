package worker

import (
	"testing"

	"pentagi/pkg/tools"
)

// TestParseFunctionArgs_SearchInMemory_QuestionsArray verifies the CLI-args mode
// still works for the "questions" field after it moved from []string to the more
// lenient tools.Strings type: repeated "-questions" flags must still build a real
// JSON array that unmarshals cleanly (the primary/common path of Strings.UnmarshalJSON).
func TestParseFunctionArgs_SearchInMemory_QuestionsArray(t *testing.T) {
	result, err := ParseFunctionArgs(tools.SearchInMemoryToolName, []string{
		"-questions", "first query",
		"-questions", "second query",
		"-message", "test run",
	})
	if err != nil {
		t.Fatalf("ParseFunctionArgs() unexpected error: %v", err)
	}

	action, ok := result.(*tools.SearchInMemoryAction)
	if !ok {
		t.Fatalf("ParseFunctionArgs() returned %T, want *tools.SearchInMemoryAction", result)
	}

	if len(action.Questions) != 2 {
		t.Fatalf("Questions length = %d, want 2 (got: %v)", len(action.Questions), []string(action.Questions))
	}
	if action.Questions[0] != "first query" || action.Questions[1] != "second query" {
		t.Errorf("Questions = %v, want [first query second query]", []string(action.Questions))
	}
	if action.Message != "test run" {
		t.Errorf("Message = %q, want %q", action.Message, "test run")
	}
}

// TestParseFunctionArgs_SearchInMemory_SingleQuestion verifies a single -questions
// flag (one array element) still round-trips correctly.
func TestParseFunctionArgs_SearchInMemory_SingleQuestion(t *testing.T) {
	result, err := ParseFunctionArgs(tools.SearchInMemoryToolName, []string{
		"-questions", "only query",
	})
	if err != nil {
		t.Fatalf("ParseFunctionArgs() unexpected error: %v", err)
	}

	action, ok := result.(*tools.SearchInMemoryAction)
	if !ok {
		t.Fatalf("ParseFunctionArgs() returned %T, want *tools.SearchInMemoryAction", result)
	}

	if len(action.Questions) != 1 || action.Questions[0] != "only query" {
		t.Fatalf("Questions = %v, want [only query]", []string(action.Questions))
	}
}
