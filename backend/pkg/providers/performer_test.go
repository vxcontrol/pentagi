package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestAgentChainIterationCap(t *testing.T) {
	t.Run("maxAgentChainIterations is set", func(t *testing.T) {
		assert.Equal(t, 100, maxAgentChainIterations,
			"maxAgentChainIterations should be 100 to prevent infinite loops")
	})
}

func TestRepeatingDetectorShouldError(t *testing.T) {
	makeToolCall := func(name, args string) llms.ToolCall {
		return llms.ToolCall{
			FunctionCall: &llms.FunctionCall{
				Name:      name,
				Arguments: args,
			},
		}
	}

	t.Run("no error before threshold", func(t *testing.T) {
		rd := &repeatingDetector{}
		tc := makeToolCall("test_tool", `{"key":"value"}`)

		// First RepeatingToolCallThreshold calls build up to first detection
		for i := 0; i < RepeatingToolCallThreshold; i++ {
			rd.detect(tc)
		}
		assert.Equal(t, 1, rd.consecutiveDetections)
		assert.False(t, rd.shouldError(),
			"should not error after only 1 consecutive detection")
	})

	t.Run("error after max consecutive detections", func(t *testing.T) {
		rd := &repeatingDetector{}
		tc := makeToolCall("test_tool", `{"key":"value"}`)

		// Each detection requires RepeatingToolCallThreshold identical calls.
		// But after the first detection the counter keeps growing, so every
		// subsequent identical call is also a detection.
		// We need enough calls to reach MaxConsecutiveRepeatingDetections.
		totalCallsNeeded := RepeatingToolCallThreshold + MaxConsecutiveRepeatingDetections - 1
		for i := 0; i < totalCallsNeeded; i++ {
			rd.detect(tc)
		}

		require.Equal(t, MaxConsecutiveRepeatingDetections, rd.consecutiveDetections)
		assert.True(t, rd.shouldError(),
			"should error after %d consecutive detections", MaxConsecutiveRepeatingDetections)
	})

	t.Run("different tool call resets detection", func(t *testing.T) {
		rd := &repeatingDetector{}
		tc1 := makeToolCall("tool_a", `{"key":"value"}`)
		tc2 := makeToolCall("tool_b", `{"key":"value"}`)

		// Build up detections for tool_a
		for i := 0; i < RepeatingToolCallThreshold; i++ {
			rd.detect(tc1)
		}
		assert.Equal(t, 1, rd.consecutiveDetections)

		// Switch to a different tool - should reset funcCalls
		rd.detect(tc2)
		assert.Equal(t, 1, rd.consecutiveDetections,
			"consecutive detections should not increase after a different tool call")
	})

	t.Run("nil function call does not trigger", func(t *testing.T) {
		rd := &repeatingDetector{}
		tc := llms.ToolCall{FunctionCall: nil}

		detected := rd.detect(tc)
		assert.False(t, detected)
		assert.False(t, rd.shouldError())
	})
}
