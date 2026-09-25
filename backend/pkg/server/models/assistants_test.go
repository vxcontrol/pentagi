package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assistantsAssistant() Assistant {
	traceID := "trace-123"
	msgchainID := uint64(1)

	return Assistant{
		Status:             AssistantStatusCreated,
		Title:              "test assistant",
		Model:              "gpt-4",
		ModelProviderName:  "openai",
		ModelProviderType:  ProviderType("openai"),
		Language:           "en",
		ToolCallIDTemplate: "call_{id}",
		TraceID:            &traceID,
		FlowID:             1,
		MsgchainID:         &msgchainID,
	}
}

// Rows are keyed by the type whose Valid they call.
func TestAssistants_Valid_RefusesExactlyTheInvalidField(t *testing.T) {
	t.Parallel()

	input := "user response"

	modelsCheckValid(t, []modelsValidCase{
		{"a created assistant status", AssistantStatusCreated, ""},
		{"a running assistant status", AssistantStatusRunning, ""},
		{"a waiting assistant status", AssistantStatusWaiting, ""},
		{"a finished assistant status", AssistantStatusFinished, ""},
		{"a failed assistant status", AssistantStatusFailed, ""},
		{"an empty assistant status", AssistantStatus(""), "invalid AssistantStatus: "},
		{"an unknown assistant status", AssistantStatus("unknown"), "invalid AssistantStatus: unknown"},

		{"a complete assistant", assistantsAssistant(), ""},
		{"an assistant in an unknown status", modelsWith(assistantsAssistant(), func(a *Assistant) { a.Status = "invalid" }),
			"Assistant.Status:valid"},
		{"an assistant without a title", modelsWith(assistantsAssistant(), func(a *Assistant) { a.Title = "" }),
			"Assistant.Title:required"},
		{"an assistant without a model", modelsWith(assistantsAssistant(), func(a *Assistant) { a.Model = "" }),
			"Assistant.Model:required"},

		{"an assistant request", CreateAssistant{Input: "hello", Provider: "openai"}, ""},
		{"an assistant request without input", CreateAssistant{Provider: "openai"}, "CreateAssistant.Input:required"},
		{"an assistant request without a provider", CreateAssistant{Input: "hello"}, "CreateAssistant.Provider:required"},

		{"stopping an assistant", PatchAssistant{Action: "stop"}, ""},
		{"answering an assistant", PatchAssistant{Action: "input", Input: &input}, ""},
		{"an action assistants do not take", PatchAssistant{Action: "restart"}, "PatchAssistant.Action:oneof"},
		{"answering an assistant with nothing", PatchAssistant{Action: "input"}, "PatchAssistant.Input:required_if"},

		{"an assistant with its flow", AssistantFlow{Flow: modelsFlow(), Assistant: assistantsAssistant()}, ""},
		{"an assistant with an invalid flow", AssistantFlow{Flow: modelsWith(modelsFlow(), func(f *Flow) { f.Title = "" }),
			Assistant: assistantsAssistant()}, "Flow.Title:required"},
		{"an invalid assistant with its flow", AssistantFlow{Flow: modelsFlow(),
			Assistant: modelsWith(assistantsAssistant(), func(a *Assistant) { a.Title = "" })}, "Assistant.Title:required"},
	})
}

func TestAssistants_String_SpellsTheStoredValue(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		value fmt.Stringer
		want  string
	}{
		{AssistantStatusCreated, "created"},
		{AssistantStatusRunning, "running"},
		{AssistantStatusFinished, "finished"},
	} {
		assert.Equal(t, tc.want, tc.value.String())
	}
}

func TestAssistants_TableName_NamesTheMappedTable(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "assistants", (&Assistant{}).TableName())
}
