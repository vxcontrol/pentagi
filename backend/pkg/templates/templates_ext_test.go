// External package: these tests use pkg/templates/validator, which imports pkg/templates.
package templates_test

import (
	"reflect"
	"strings"
	"testing"

	"pentagi/pkg/templates"
	"pentagi/pkg/templates/validator"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// templatesShippedPrompts returns every Prompt of GetDefaultPrompts keyed by its field path.
func templatesShippedPrompts(t *testing.T) map[string]templates.Prompt {
	t.Helper()

	defaultPrompts, err := templates.GetDefaultPrompts()
	require.NoError(t, err)

	prompts := map[string]templates.Prompt{}
	var walk func(v reflect.Value, path string)
	walk = func(v reflect.Value, path string) {
		for i := range v.NumField() {
			field, name := v.Field(i), path+"."+v.Type().Field(i).Name
			if prompt, ok := field.Interface().(templates.Prompt); ok {
				prompts[name] = prompt
			} else if field.Kind() == reflect.Struct {
				walk(field, name)
			}
		}
	}
	walk(reflect.ValueOf(*defaultPrompts), "DefaultPrompts")

	return prompts
}

// templatesCaseName turns a prompt type into a subtest sentence.
func templatesCaseName(promptType templates.PromptType) string {
	return "the " + strings.ReplaceAll(string(promptType), "_", " ") + " prompt"
}

func TestTemplates_GetDefaultPrompts_DeclaresExactlyTheVariablesEachPromptUses(t *testing.T) {
	prompts := templatesShippedPrompts(t)

	agents, tools := 0, 0
	shipped := map[templates.PromptType]bool{}
	for name, prompt := range prompts {
		if strings.HasPrefix(name, "DefaultPrompts.AgentsPrompts.") {
			agents++
		} else {
			tools++
		}
		shipped[prompt.Type] = true

		t.Run(templatesCaseName(prompt.Type), func(t *testing.T) {
			require.NotEmpty(t, strings.TrimSpace(prompt.Template), "%s has no template", name)

			used, err := validator.ExtractTemplateVariables(prompt.Template)
			require.NoError(t, err, name)
			assert.ElementsMatch(t, prompt.Variables, used, "%s: declared variables against the ones the template uses", name)

			_, declared := templates.PromptVariables[prompt.Type]
			assert.True(t, declared, "%s is missing from PromptVariables", prompt.Type)
		})
	}

	assert.Equal(t, 27, agents, "agent prompts")
	assert.Equal(t, 12, tools, "tool prompts")
	for promptType := range templates.PromptVariables {
		assert.True(t, shipped[promptType], "PromptVariables holds %s, which no shipped prompt has", promptType)
	}
}

func TestTemplates_GetDefaultPrompts_EveryPromptRendersWithDummyData(t *testing.T) {
	dummyData := validator.CreateDummyTemplateData()

	for name, prompt := range templatesShippedPrompts(t) {
		t.Run(templatesCaseName(prompt.Type), func(t *testing.T) {
			_, err := templates.RenderPrompt(string(prompt.Type), prompt.Template, dummyData)
			assert.NoError(t, err, name)
		})
	}
}

func TestTemplates_GetDefaultPrompts_PromptsCarryTheirKeyPhrases(t *testing.T) {
	defaultPrompts, err := templates.GetDefaultPrompts()
	require.NoError(t, err)

	dummyData := validator.CreateDummyTemplateData()
	str := func(key string) string { return dummyData[key].(string) }
	recent := dummyData["RecentMessages"].([]map[string]string)[0]
	executed := dummyData["ExecutedToolCalls"].([]map[string]string)[0]

	tests := []struct {
		name       string
		promptType templates.PromptType
		template   string
		phrases    []string
		forbidden  []string
	}{
		{
			name:       "the execution monitor question",
			promptType: templates.PromptTypeQuestionExecutionMonitor,
			template:   defaultPrompts.ToolsPrompts.QuestionExecutionMonitor.Template,
			phrases: []string{
				str("SubtaskDescription"), str("AgentType"), str("AgentPrompt"),
				str("LastToolName"), str("LastToolArgs"), str("LastToolResult"),
				recent["name"], recent["msg"], executed["name"], executed["result"],
				"my_current_assignment", "my_role_and_capabilities", "recent_conversation_history",
				"all_tool_calls_i_executed", "my_most_recent_action",
				"making real, measurable progress", "repeating the same actions", "stuck in a loop",
				"completely different strategy", "impossible to complete", "critical and actionable next steps",
			},
		},
		{
			name:       "the task planner question",
			promptType: templates.PromptTypeQuestionTaskPlanner,
			template:   defaultPrompts.ToolsPrompts.QuestionTaskPlanner.Template,
			phrases: []string{
				str("AgentType"), str("TaskQuestion"),
				"my_task", "structured execution plan", "concise checklist", "actionable steps",
				"specific, actionable steps", "check or verify", "potential pitfalls",
				"stay focused only on this current task", "avoid redundant work", "efficient task completion",
				"numbered checklist", "1. [First critical action",
			},
		},
		{
			name:       "the pentester's tool-agnostic cli argument guidance",
			promptType: templates.PromptTypePentester,
			template:   defaultPrompts.AgentsPrompts.Pentester.System.Template,
			phrases: []string{
				"cli_argument_protocol", "Hallucinated flags", "--help", "Cross-tool flag assumptions",
				"Never copy a flag from one tool to another", "shell redirection", "Machine-readable output", "Argument quoting",
			},
			forbidden: []string{"XSStrike", "xsstrike"},
		},
		{
			name:       "the task assignment wrapper",
			promptType: templates.PromptTypeTaskAssignmentWrapper,
			template:   defaultPrompts.ToolsPrompts.TaskAssignmentWrapper.Template,
			phrases: []string{
				str("OriginalRequest"), str("ExecutionPlan"),
				"task_assignment", "original_request", "execution_plan", "hint",
				"primary objective", "prepared by analyzing the broader context", "decomposing the task", "suggested steps",
				"Use this plan as guidance", "adapt your actions", "staying aligned with the objective",
				"</task_assignment>", "</original_request>", "</execution_plan>", "</hint>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered, err := templates.RenderPrompt(string(tt.promptType), tt.template, dummyData)
			require.NoError(t, err)

			for _, phrase := range tt.phrases {
				assert.Contains(t, rendered, phrase)
			}
			for _, phrase := range tt.forbidden {
				assert.NotContains(t, rendered, phrase, "the guidance must not single out one tool")
			}
		})
	}
}
