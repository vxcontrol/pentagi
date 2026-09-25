package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Rows are keyed by the type whose Valid they call.
func TestFlows_Valid_RefusesExactlyTheInvalidField(t *testing.T) {
	t.Parallel()

	longTraceID := "trace-id-that-is-longer-than-the-supported-seventy-character-validation-limit-123"
	untitled := modelsWith(modelsFlow(), func(f *Flow) { f.Title = "" })
	tasks := []TaskSubtasks{{Subtasks: []Subtask{modelsSubtask()}, Task: modelsTask()}}
	input, name := "new input", "new name"
	budget, tooShort, tooLong := uint64(600), uint64(29), uint64(7201)

	modelsCheckValid(t, []modelsValidCase{
		{"a created flow status", FlowStatusCreated, ""},
		{"a running flow status", FlowStatusRunning, ""},
		{"a waiting flow status", FlowStatusWaiting, ""},
		{"a finished flow status", FlowStatusFinished, ""},
		{"a failed flow status", FlowStatusFailed, ""},
		{"an empty flow status", FlowStatus(""), "invalid FlowStatus: "},
		{"an unknown flow status", FlowStatus("unknown"), "invalid FlowStatus: unknown"},

		{"a complete flow", modelsFlow(), ""},
		{"a flow whose trace id is pending", modelsWith(modelsFlow(), func(f *Flow) { f.TraceID = nil }), ""},
		{"a flow with an overlong trace id", modelsWith(modelsFlow(), func(f *Flow) { f.TraceID = &longTraceID }),
			"Flow.TraceID:max"},
		{"a flow without a title", untitled, "Flow.Title:required"},
		{"a flow in an unknown status", modelsWith(modelsFlow(), func(f *Flow) { f.Status = "invalid" }), "Flow.Status:valid"},
		{"a flow of an unknown provider type", modelsWith(modelsFlow(), func(f *Flow) { f.ModelProviderType = "invalid" }),
			"Flow.ModelProviderType:valid"},

		{"a flow request", CreateFlow{Input: "scan target", Provider: "openai"}, ""},
		{"a flow request without input", CreateFlow{Provider: "openai"}, "CreateFlow.Input:required"},
		{"a flow request without a provider", CreateFlow{Input: "scan target"}, "CreateFlow.Provider:required"},

		{"stopping a flow", PatchFlow{Action: "stop"}, ""},
		{"finishing a flow", PatchFlow{Action: "finish"}, ""},
		{"answering a flow", PatchFlow{Action: "input", Input: &input}, ""},
		{"renaming a flow", PatchFlow{Action: "rename", Name: &name}, ""},
		{"reporting on a flow", PatchFlow{Action: "report"}, ""},
		{"reporting on a flow within a budget", PatchFlow{Action: "report", Timeout: &budget}, ""},
		{"a report budget below the floor", PatchFlow{Action: "report", Timeout: &tooShort}, "PatchFlow.Timeout:min"},
		{"a report budget above the ceiling", PatchFlow{Action: "report", Timeout: &tooLong}, "PatchFlow.Timeout:max"},
		{"an action flows do not take", PatchFlow{Action: "invalid"}, "PatchFlow.Action:oneof"},
		{"no action", PatchFlow{}, "PatchFlow.Action:required"},
		{"answering a flow with nothing", PatchFlow{Action: "input"}, "PatchFlow.Input:required_if"},
		{"renaming a flow to nothing", PatchFlow{Action: "rename"}, "PatchFlow.Name:required_if"},

		{"a flow with its tasks", FlowTasksSubtasks{Tasks: tasks, Flow: modelsFlow()}, ""},
		{"an invalid flow with its tasks", FlowTasksSubtasks{Tasks: tasks, Flow: untitled}, "Flow.Title:required"},
		{"a flow with an invalid subtask", FlowTasksSubtasks{Tasks: []TaskSubtasks{{
			Subtasks: []Subtask{modelsWith(modelsSubtask(), func(s *Subtask) { s.Title = "" })},
			Task:     modelsTask(),
		}}, Flow: modelsFlow()}, "Subtask.Title:required"},

		{"a flow with its containers", FlowContainers{Containers: []Container{modelsContainer()}, Flow: modelsFlow()}, ""},
		{"an invalid flow with its containers", FlowContainers{Containers: []Container{modelsContainer()}, Flow: untitled},
			"Flow.Title:required"},
		{"a flow with an invalid container", FlowContainers{
			Containers: []Container{modelsWith(modelsContainer(), func(c *Container) { c.Name = "" })},
			Flow:       modelsFlow(),
		}, "Container.Name:required"},
	})
}

func TestFlows_String_SpellsTheStoredValue(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		value fmt.Stringer
		want  string
	}{
		{FlowStatusCreated, "created"},
		{FlowStatusRunning, "running"},
		{FlowStatusFinished, "finished"},
	} {
		assert.Equal(t, tc.want, tc.value.String())
	}
}

func TestFlows_TableName_NamesTheMappedTable(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		model modelsTabled
		want  string
	}{
		{&Flow{}, "flows"},
		{&FlowTasksSubtasks{}, "flows"},
		{&FlowContainers{}, "flows"},
	} {
		assert.Equal(t, tc.want, tc.model.TableName(), "%T", tc.model)
	}
}
