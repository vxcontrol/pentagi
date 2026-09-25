package processor

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A failure inside wrapCommand's 500 ms window arrives as a wait message ahead of its completion.
func TestModel_HandleMsg_KeepsPollingAfterAnErrorUntilTheCompletion(t *testing.T) {
	model := &processorModel{processor: &processor{}}
	state := testOperationState(t)
	state.sendCompletion(ProductStackAll, errors.New("failed before the window closed"))

	next := model.HandleMsg(ProcessorWaitMsg{
		Error: errors.New("failed before the window closed"),
		Stack: ProductStackAll, state: state,
	})
	require.NotNil(t, next, "polling stopped, so the completion already in the queue is never delivered")

	completion, ok := next().(ProcessorCompletionMsg)
	require.True(t, ok, "expected the queued completion")
	assert.EqualError(t, completion.Error, "failed before the window closed")

	assert.Nil(t, model.HandleMsg(completion), "polling must stop once the completion is delivered")
}
