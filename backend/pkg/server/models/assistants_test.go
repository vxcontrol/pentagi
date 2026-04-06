package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssistantStatusValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  AssistantStatus
		wantErr bool
	}{
		{"valid created", AssistantStatusCreated, false},
		{"valid running", AssistantStatusRunning, false},
		{"valid waiting", AssistantStatusWaiting, false},
		{"valid finished", AssistantStatusFinished, false},
		{"valid failed", AssistantStatusFailed, false},
		{"invalid empty", AssistantStatus(""), true},
		{"invalid unknown", AssistantStatus("unknown"), true},
		{"invalid paused", AssistantStatus("paused"), true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.status.Valid()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid AssistantStatus")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAssistantStatusString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "created", AssistantStatusCreated.String())
	assert.Equal(t, "running", AssistantStatusRunning.String())
	assert.Equal(t, "finished", AssistantStatusFinished.String())
}

func TestAssistantValid(t *testing.T) {
	t.Parallel()

	traceID := "trace-123"
	msgchainID := uint64(1)
	validAssistant := Assistant{
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

	t.Run("valid assistant", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, validAssistant.Valid())
	})

	t.Run("invalid status", func(t *testing.T) {
		t.Parallel()
		a := validAssistant
		a.Status = AssistantStatus("invalid")
		assert.Error(t, a.Valid())
	})

	t.Run("missing title", func(t *testing.T) {
		t.Parallel()
		a := validAssistant
		a.Title = ""
		assert.Error(t, a.Valid())
	})

	t.Run("missing model", func(t *testing.T) {
		t.Parallel()
		a := validAssistant
		a.Model = ""
		assert.Error(t, a.Valid())
	})
}

func TestAssistantTableName(t *testing.T) {
	t.Parallel()
	a := &Assistant{}
	assert.Equal(t, "assistants", a.TableName())
}

func TestCreateAssistantValid(t *testing.T) {
	t.Parallel()

	t.Run("valid create assistant", func(t *testing.T) {
		t.Parallel()
		ca := CreateAssistant{Input: "hello", Provider: "openai"}
		assert.NoError(t, ca.Valid())
	})

	t.Run("missing input", func(t *testing.T) {
		t.Parallel()
		ca := CreateAssistant{Input: "", Provider: "openai"}
		assert.Error(t, ca.Valid())
	})

	t.Run("missing provider", func(t *testing.T) {
		t.Parallel()
		ca := CreateAssistant{Input: "hello", Provider: ""}
		assert.Error(t, ca.Valid())
	})
}

func TestPatchAssistantValid(t *testing.T) {
	t.Parallel()

	t.Run("valid stop action", func(t *testing.T) {
		t.Parallel()
		pa := PatchAssistant{Action: "stop"}
		assert.NoError(t, pa.Valid())
	})

	t.Run("valid input action with input", func(t *testing.T) {
		t.Parallel()
		input := "user response"
		pa := PatchAssistant{Action: "input", Input: &input}
		assert.NoError(t, pa.Valid())
	})

	t.Run("invalid action", func(t *testing.T) {
		t.Parallel()
		pa := PatchAssistant{Action: "restart"}
		assert.Error(t, pa.Valid())
	})

	t.Run("input action without input", func(t *testing.T) {
		t.Parallel()
		pa := PatchAssistant{Action: "input"}
		assert.Error(t, pa.Valid())
	})
}

func TestAssistantFlowValid(t *testing.T) {
	t.Parallel()

	traceID := "trace-123"
	mcID := uint64(1)
	validAssistant := Assistant{
		Status:             AssistantStatusCreated,
		Title:              "test",
		Model:              "gpt-4",
		ModelProviderName:  "openai",
		ModelProviderType:  ProviderType("openai"),
		Language:           "en",
		ToolCallIDTemplate: "call_{id}",
		TraceID:            &traceID,
		FlowID:             1,
		MsgchainID:         &mcID,
	}
	validFlow := Flow{
		Status:             FlowStatusCreated,
		Title:              "flow",
		Model:              "gpt-4",
		ModelProviderName:  "openai",
		ModelProviderType:  ProviderType("openai"),
		Language:           "en",
		ToolCallIDTemplate: "call_{id}",
		TraceID:            &traceID,
		UserID:             1,
	}

	t.Run("valid assistant flow", func(t *testing.T) {
		t.Parallel()
		af := AssistantFlow{
			Flow:      validFlow,
			Assistant: validAssistant,
		}
		assert.NoError(t, af.Valid())
	})

	t.Run("invalid flow", func(t *testing.T) {
		t.Parallel()
		badFlow := validFlow
		badFlow.Title = ""
		af := AssistantFlow{
			Flow:      badFlow,
			Assistant: validAssistant,
		}
		assert.Error(t, af.Valid())
	})

	t.Run("invalid assistant", func(t *testing.T) {
		t.Parallel()
		badAssistant := validAssistant
		badAssistant.Title = ""
		af := AssistantFlow{
			Flow:      validFlow,
			Assistant: badAssistant,
		}
		assert.Error(t, af.Valid())
	})
}
