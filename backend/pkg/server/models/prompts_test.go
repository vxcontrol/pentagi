package models

import (
	"testing"

	"pentagi/pkg/templates"

	"github.com/stretchr/testify/assert"
)

func TestPromptTypeValid(t *testing.T) {
	t.Parallel()

	validTypes := []struct {
		name string
		pt   PromptType
	}{
		{"primary_agent", PromptType(templates.PromptTypePrimaryAgent)},
		{"assistant", PromptType(templates.PromptTypeAssistant)},
		{"pentester", PromptType(templates.PromptTypePentester)},
		{"coder", PromptType(templates.PromptTypeCoder)},
		{"searcher", PromptType(templates.PromptTypeSearcher)},
		{"summarizer", PromptType(templates.PromptTypeSummarizer)},
		{"reporter", PromptType(templates.PromptTypeReporter)},
		{"tool_call_id_detector", PromptType(templates.PromptTypeToolCallIDDetector)},
	}

	for _, tt := range validTypes {
		tt := tt
		t.Run("valid_"+tt.name, func(t *testing.T) {
			t.Parallel()
			assert.NoError(t, tt.pt.Valid())
		})
	}

	invalidTypes := []struct {
		name string
		pt   PromptType
	}{
		{"empty", PromptType("")},
		{"unknown", PromptType("unknown")},
		{"typo", PromptType("primarry_agent")},
	}

	for _, tt := range invalidTypes {
		tt := tt
		t.Run("invalid_"+tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.pt.Valid()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid PromptType")
		})
	}
}

func TestPromptTypeString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "primary_agent", PromptType(templates.PromptTypePrimaryAgent).String())
	assert.Equal(t, "assistant", PromptType(templates.PromptTypeAssistant).String())
}

func TestPromptValid(t *testing.T) {
	t.Parallel()

	validPrompt := Prompt{
		Type:   PromptType(templates.PromptTypePrimaryAgent),
		Prompt: "You are a security assistant.",
	}

	t.Run("valid prompt", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, validPrompt.Valid())
	})

	t.Run("invalid type", func(t *testing.T) {
		t.Parallel()
		p := validPrompt
		p.Type = PromptType("invalid")
		assert.Error(t, p.Valid())
	})

	t.Run("missing prompt text", func(t *testing.T) {
		t.Parallel()
		p := validPrompt
		p.Prompt = ""
		assert.Error(t, p.Valid())
	})
}

func TestPromptTableName(t *testing.T) {
	t.Parallel()
	p := &Prompt{}
	assert.Equal(t, "prompts", p.TableName())
}

func TestPatchPromptValid(t *testing.T) {
	t.Parallel()

	t.Run("valid patch", func(t *testing.T) {
		t.Parallel()
		pp := PatchPrompt{Prompt: "updated prompt text"}
		assert.NoError(t, pp.Valid())
	})

	t.Run("missing prompt", func(t *testing.T) {
		t.Parallel()
		pp := PatchPrompt{Prompt: ""}
		assert.Error(t, pp.Valid())
	})
}
