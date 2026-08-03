package worker

import (
	"strings"
	"testing"

	"pentagi/pkg/tools"
)

func TestFillStructFromMap_StringsField_FromSlice(t *testing.T) {
	action := &tools.SearchInMemoryAction{}

	err := fillStructFromMap(action, map[string]any{
		"questions": []string{"query one", "query two"},
		"message":   "test",
	})
	if err != nil {
		t.Fatalf("fillStructFromMap() unexpected error: %v", err)
	}
	if len(action.Questions) != 2 || action.Questions[0] != "query one" || action.Questions[1] != "query two" {
		t.Fatalf("Questions = %v, want [query one query two]", []string(action.Questions))
	}
	if action.Message != "test" {
		t.Fatalf("Message = %q, want %q", action.Message, "test")
	}
}

func TestFillStructFromMap_StringsField_FromSingleString(t *testing.T) {
	action := &tools.SearchInMemoryAction{}

	err := fillStructFromMap(action, map[string]any{
		"questions": "single question",
	})
	if err != nil {
		t.Fatalf("fillStructFromMap() unexpected error: %v", err)
	}
	if len(action.Questions) != 1 || action.Questions[0] != "single question" {
		t.Fatalf("Questions = %v, want [single question]", []string(action.Questions))
	}
}

func TestFillStructFromMap_TypeMismatch_ReturnsClearError_NotPanic(t *testing.T) {
	t.Run("string field given int", func(t *testing.T) {
		action := &tools.SearchInMemoryAction{}
		// The whole point of this test: a mismatched type must return an
		// error, not panic with "interface conversion: interface {} is X, not Y".
		err := fillStructFromMap(action, map[string]any{"message": 42})
		assertContainsError(t, err, "expected string")
	})

	t.Run("int field given string", func(t *testing.T) {
		// MaxResults is a non-pointer tools.Int64 field (Kind() == Int64).
		action := &tools.SploitusAction{}
		err := fillStructFromMap(action, map[string]any{"max_results": "not-a-number"})
		assertContainsError(t, err, "expected int")
	})

	t.Run("slice field given bool", func(t *testing.T) {
		action := &tools.SearchInMemoryAction{}
		err := fillStructFromMap(action, map[string]any{"questions": true})
		assertContainsError(t, err, "expected []string or string")
	})
}

func assertContainsError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error for mismatched argument type, got nil")
	}
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err.Error(), want)
	}
}
