package models

import (
	"testing"

	"pentagi/pkg/templates"

	"github.com/stretchr/testify/assert"
)

// Rows are keyed by the type whose Valid they call.
func TestPrompts_Valid_RefusesExactlyTheInvalidField(t *testing.T) {
	t.Parallel()

	prompt := Prompt{Type: "primary_agent", Prompt: "You are a security assistant."}
	cases := []modelsValidCase{
		{"the primary agent prompt type", PromptType("primary_agent"), ""},
		{"an empty prompt type", PromptType(""), "invalid PromptType: "},
		{"a misspelt prompt type", PromptType("primarry_agent"), "invalid PromptType: primarry_agent"},

		{"a complete prompt", prompt, ""},
		{"a prompt of an unknown type", modelsWith(prompt, func(p *Prompt) { p.Type = "invalid" }), "Prompt.Type:valid"},
		{"a prompt without text", modelsWith(prompt, func(p *Prompt) { p.Prompt = "" }), "Prompt.Prompt:required"},

		{"a prompt edit", PatchPrompt{Prompt: "updated prompt text"}, ""},
		{"a prompt edit without text", PatchPrompt{}, "PatchPrompt.Prompt:required"},
	}
	for declared := range templates.PromptVariables {
		cases = append(cases, modelsValidCase{"the declared prompt type " + string(declared), PromptType(declared), ""})
	}

	modelsCheckValid(t, cases)
}

func TestPrompts_String_SpellsTheStoredValue(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "primary_agent", PromptType(templates.PromptTypePrimaryAgent).String())
	assert.Equal(t, "assistant", PromptType(templates.PromptTypeAssistant).String())
}

func TestPrompts_TableName_NamesTheMappedTable(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "prompts", (&Prompt{}).TableName())
}
