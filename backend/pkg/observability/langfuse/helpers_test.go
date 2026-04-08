package langfuse

import (
	"regexp"
	"testing"
	"time"

	"pentagi/pkg/observability/langfuse/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestMergeMaps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		dst      map[string]any
		src      map[string]any
		expected map[string]any
	}{
		{
			name:     "both nil",
			dst:      nil,
			src:      nil,
			expected: nil,
		},
		{
			name:     "src nil returns dst",
			dst:      map[string]any{"a": 1},
			src:      nil,
			expected: map[string]any{"a": 1},
		},
		{
			name:     "dst nil copies src",
			dst:      nil,
			src:      map[string]any{"b": 2},
			expected: map[string]any{"b": 2},
		},
		{
			name:     "disjoint keys",
			dst:      map[string]any{"a": 1},
			src:      map[string]any{"b": 2},
			expected: map[string]any{"a": 1, "b": 2},
		},
		{
			name:     "overlapping keys src overrides",
			dst:      map[string]any{"a": 1, "b": 2},
			src:      map[string]any{"b": 99, "c": 3},
			expected: map[string]any{"a": 1, "b": 99, "c": 3},
		},
		{
			name:     "empty maps",
			dst:      map[string]any{},
			src:      map[string]any{},
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := mergeMaps(tt.dst, tt.src)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMergeMaps_DoesNotMutateDst(t *testing.T) {
	t.Parallel()

	dst := map[string]any{"a": 1}
	src := map[string]any{"b": 2}
	result := mergeMaps(dst, src)

	assert.Equal(t, map[string]any{"a": 1, "b": 2}, result)
	assert.Equal(t, map[string]any{"a": 1}, dst, "original dst must not be mutated")
}

func TestObservationLevel_ToLangfuse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		level    ObservationLevel
		expected api.ObservationLevel
	}{
		{"default", ObservationLevelDefault, api.ObservationLevelDefault},
		{"debug", ObservationLevelDebug, api.ObservationLevelDebug},
		{"warning", ObservationLevelWarning, api.ObservationLevelWarning},
		{"error", ObservationLevelError, api.ObservationLevelError},
		{"unknown falls back to default", ObservationLevel(99), api.ObservationLevelDefault},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.level.ToLangfuse()
			require.NotNil(t, result)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

func TestGenerationUsageUnit_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		unit     GenerationUsageUnit
		expected string
	}{
		{"tokens", GenerationUsageUnitTokens, "TOKENS"},
		{"characters", GenerationUsageUnitCharacters, "CHARACTERS"},
		{"milliseconds", GenerationUsageUnitMilliseconds, "MILLISECONDS"},
		{"seconds", GenerationUsageUnitSeconds, "seconds"},
		{"images", GenerationUsageUnitImages, "IMAGES"},
		{"requests", GenerationUsageUnitRequests, "REQUESTS"},
		{"unknown returns empty", GenerationUsageUnit(99), ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.unit.String())
		})
	}
}

func TestGenerationUsageUnit_ToLangfuse(t *testing.T) {
	t.Parallel()

	t.Run("valid unit returns pointer", func(t *testing.T) {
		t.Parallel()
		result := GenerationUsageUnitTokens.ToLangfuse()
		require.NotNil(t, result)
		assert.Equal(t, "TOKENS", *result)
	})

	t.Run("unknown unit returns nil", func(t *testing.T) {
		t.Parallel()
		result := GenerationUsageUnit(99).ToLangfuse()
		assert.Nil(t, result)
	})
}

func TestGenerationUsage_ToLangfuse(t *testing.T) {
	t.Parallel()

	t.Run("nil receiver returns nil", func(t *testing.T) {
		t.Parallel()
		var u *GenerationUsage
		assert.Nil(t, u.ToLangfuse())
	})

	t.Run("basic usage no costs", func(t *testing.T) {
		t.Parallel()
		u := &GenerationUsage{Input: 100, Output: 50, Unit: GenerationUsageUnitTokens}
		result := u.ToLangfuse()
		require.NotNil(t, result)
		require.NotNil(t, result.Usage)
		assert.Equal(t, 100, result.Usage.Input)
		assert.Equal(t, 50, result.Usage.Output)
		assert.Equal(t, 150, result.Usage.Total)
		assert.Nil(t, result.Usage.TotalCost)
	})

	t.Run("input cost only", func(t *testing.T) {
		t.Parallel()
		inputCost := 0.01
		u := &GenerationUsage{Input: 100, InputCost: &inputCost}
		result := u.ToLangfuse()
		require.NotNil(t, result)
		require.NotNil(t, result.Usage.TotalCost)
		assert.Equal(t, 0.01, *result.Usage.TotalCost)
	})

	t.Run("output cost only", func(t *testing.T) {
		t.Parallel()
		outputCost := 0.02
		u := &GenerationUsage{Output: 50, OutputCost: &outputCost}
		result := u.ToLangfuse()
		require.NotNil(t, result)
		require.NotNil(t, result.Usage.TotalCost)
		assert.Equal(t, 0.02, *result.Usage.TotalCost)
	})

	t.Run("both costs summed", func(t *testing.T) {
		t.Parallel()
		inputCost := 0.01
		outputCost := 0.02
		u := &GenerationUsage{
			Input:      100,
			Output:     50,
			InputCost:  &inputCost,
			OutputCost: &outputCost,
			Unit:       GenerationUsageUnitTokens,
		}
		result := u.ToLangfuse()
		require.NotNil(t, result)
		require.NotNil(t, result.Usage.TotalCost)
		assert.InDelta(t, 0.03, *result.Usage.TotalCost, 1e-9)
		require.NotNil(t, result.Usage.InputCost)
		assert.Equal(t, 0.01, *result.Usage.InputCost)
		require.NotNil(t, result.Usage.OutputCost)
		assert.Equal(t, 0.02, *result.Usage.OutputCost)
	})
}

func TestModelParameters_ToLangfuse(t *testing.T) {
	t.Parallel()

	t.Run("nil receiver returns nil", func(t *testing.T) {
		t.Parallel()
		var m *ModelParameters
		assert.Nil(t, m.ToLangfuse())
	})

	t.Run("empty struct sets max_tokens to inf", func(t *testing.T) {
		t.Parallel()
		m := &ModelParameters{}
		result := m.ToLangfuse()
		require.NotNil(t, result)
		v, ok := result["max_tokens"]
		require.True(t, ok, "max_tokens key must exist")
		require.NotNil(t, v)
		require.NotNil(t, v.GetStringOptional())
		assert.Equal(t, "inf", *v.GetStringOptional())
	})

	t.Run("temperature and top_p values", func(t *testing.T) {
		t.Parallel()
		temp := 0.7
		topP := 0.9
		m := &ModelParameters{Temperature: &temp, TopP: &topP}
		result := m.ToLangfuse()
		require.NotNil(t, result)
		require.NotNil(t, result["temperature"])
		require.NotNil(t, result["temperature"].GetStringOptional())
		assert.Equal(t, "0.7", *result["temperature"].GetStringOptional())
		require.NotNil(t, result["top_p"])
		require.NotNil(t, result["top_p"].GetStringOptional())
		assert.Equal(t, "0.9", *result["top_p"].GetStringOptional())
	})

	t.Run("max_tokens explicit integer value", func(t *testing.T) {
		t.Parallel()
		maxTokens := 1024
		m := &ModelParameters{MaxTokens: &maxTokens}
		result := m.ToLangfuse()
		require.NotNil(t, result)
		require.NotNil(t, result["max_tokens"])
		require.NotNil(t, result["max_tokens"].GetIntegerOptional())
		assert.Equal(t, 1024, *result["max_tokens"].GetIntegerOptional())
	})

	t.Run("json mode boolean value", func(t *testing.T) {
		t.Parallel()
		m := &ModelParameters{JSONMode: true}
		result := m.ToLangfuse()
		require.NotNil(t, result)
		require.NotNil(t, result["json"])
		require.NotNil(t, result["json"].GetBooleanOptional())
		assert.True(t, *result["json"].GetBooleanOptional())
	})

	t.Run("stop words list value", func(t *testing.T) {
		t.Parallel()
		m := &ModelParameters{StopWords: []string{"END", "STOP"}}
		result := m.ToLangfuse()
		require.NotNil(t, result)
		require.NotNil(t, result["stop_words"])
		assert.Equal(t, []string{"END", "STOP"}, result["stop_words"].GetStringListOptional())
	})

	t.Run("integer fields serialized correctly", func(t *testing.T) {
		t.Parallel()
		topK := 40
		seed := 42
		candidateCount := 3
		n := 2
		m := &ModelParameters{
			TopK:           &topK,
			Seed:           &seed,
			CandidateCount: &candidateCount,
			N:              &n,
		}
		result := m.ToLangfuse()
		require.NotNil(t, result)
		require.NotNil(t, result["top_k"].GetIntegerOptional())
		assert.Equal(t, 40, *result["top_k"].GetIntegerOptional())
		require.NotNil(t, result["seed"].GetIntegerOptional())
		assert.Equal(t, 42, *result["seed"].GetIntegerOptional())
		require.NotNil(t, result["candidate_count"].GetIntegerOptional())
		assert.Equal(t, 3, *result["candidate_count"].GetIntegerOptional())
		require.NotNil(t, result["n"].GetIntegerOptional())
		assert.Equal(t, 2, *result["n"].GetIntegerOptional())
	})

	t.Run("float fields formatted as strings", func(t *testing.T) {
		t.Parallel()
		minP := 0.1
		repPenalty := 1.1
		freqPenalty := 0.5
		presPenalty := 0.6
		m := &ModelParameters{
			MinP:              &minP,
			RepetitionPenalty: &repPenalty,
			FrequencyPenalty:  &freqPenalty,
			PresencePenalty:   &presPenalty,
		}
		result := m.ToLangfuse()
		require.NotNil(t, result)
		assert.Equal(t, "0.1", *result["min_p"].GetStringOptional())
		assert.Equal(t, "1.1", *result["repetition_penalty"].GetStringOptional())
		assert.Equal(t, "0.5", *result["frequency_penalty"].GetStringOptional())
		assert.Equal(t, "0.6", *result["presence_penalty"].GetStringOptional())
	})

	t.Run("all optional fields present", func(t *testing.T) {
		t.Parallel()
		temp := 0.5
		topP := 0.9
		minP := 0.1
		topK := 40
		seed := 42
		maxTokens := 2048
		candidateCount := 1
		minLen := 10
		maxLen := 500
		n := 1
		repPenalty := 1.1
		freqPenalty := 0.5
		presPenalty := 0.6
		m := &ModelParameters{
			Temperature:       &temp,
			TopP:              &topP,
			MinP:              &minP,
			TopK:              &topK,
			Seed:              &seed,
			MaxTokens:         &maxTokens,
			CandidateCount:    &candidateCount,
			MinLength:         &minLen,
			MaxLength:         &maxLen,
			N:                 &n,
			RepetitionPenalty: &repPenalty,
			FrequencyPenalty:  &freqPenalty,
			PresencePenalty:   &presPenalty,
			JSONMode:          true,
			StopWords:         []string{"END"},
		}
		result := m.ToLangfuse()
		require.NotNil(t, result)
		assert.Len(t, result, 15, "all 15 parameter keys must be present")
	})
}

func TestGetLangchainModelParameters(t *testing.T) {
	t.Parallel()

	t.Run("nil options returns nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, GetLangchainModelParameters(nil))
	})

	t.Run("empty options returns nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, GetLangchainModelParameters([]llms.CallOption{}))
	})

	t.Run("with temperature option", func(t *testing.T) {
		t.Parallel()
		opts := []llms.CallOption{
			llms.WithTemperature(0.7),
		}
		result := GetLangchainModelParameters(opts)
		require.NotNil(t, result)
		require.NotNil(t, result.Temperature)
		assert.InDelta(t, 0.7, *result.Temperature, 1e-9)
	})

	t.Run("with max tokens option", func(t *testing.T) {
		t.Parallel()
		opts := []llms.CallOption{
			llms.WithMaxTokens(512),
		}
		result := GetLangchainModelParameters(opts)
		require.NotNil(t, result)
		require.NotNil(t, result.MaxTokens)
		assert.Equal(t, 512, *result.MaxTokens)
	})
}

func TestNewTraceID(t *testing.T) {
	t.Parallel()

	t.Run("length is 32 hex chars", func(t *testing.T) {
		t.Parallel()
		id := newTraceID()
		assert.Len(t, id, 32)
		assert.Regexp(t, regexp.MustCompile(`^[0-9a-f]{32}$`), id)
	})

	t.Run("unique across calls", func(t *testing.T) {
		t.Parallel()
		ids := make(map[string]bool)
		for i := 0; i < 100; i++ {
			id := newTraceID()
			assert.False(t, ids[id], "duplicate trace ID generated")
			ids[id] = true
		}
	})
}

func TestNewSpanID(t *testing.T) {
	t.Parallel()

	t.Run("length is 16 hex chars", func(t *testing.T) {
		t.Parallel()
		id := newSpanID()
		assert.Len(t, id, 16)
		assert.Regexp(t, regexp.MustCompile(`^[0-9a-f]{16}$`), id)
	})

	t.Run("unique across calls", func(t *testing.T) {
		t.Parallel()
		ids := make(map[string]bool)
		for i := 0; i < 100; i++ {
			id := newSpanID()
			assert.False(t, ids[id], "duplicate span ID generated")
			ids[id] = true
		}
	})
}

func TestGetCurrentTime(t *testing.T) {
	t.Parallel()

	now := getCurrentTime()
	assert.Equal(t, time.UTC, now.Location(), "must be UTC")
	assert.WithinDuration(t, time.Now().UTC(), now, 2*time.Second)
}

func TestGetCurrentTimeString(t *testing.T) {
	t.Parallel()

	ts := getCurrentTimeString()
	_, err := time.Parse(timeFormat8601, ts)
	assert.NoError(t, err, "must parse with timeFormat8601")
}

func TestGetCurrentTimeRef(t *testing.T) {
	t.Parallel()

	ref := getCurrentTimeRef()
	require.NotNil(t, ref)
	assert.Equal(t, time.UTC, ref.Location(), "must be UTC")
	assert.WithinDuration(t, time.Now().UTC(), *ref, 2*time.Second)
}

func TestGetTimeRef(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	ref := getTimeRef(now)
	require.NotNil(t, ref)
	assert.Equal(t, now, *ref)
}

func TestGetTimeRefString(t *testing.T) {
	t.Parallel()

	t.Run("nil time returns current time string", func(t *testing.T) {
		t.Parallel()
		result := getTimeRefString(nil)
		_, err := time.Parse(timeFormat8601, result)
		assert.NoError(t, err)
	})

	t.Run("non-nil time formats correctly", func(t *testing.T) {
		t.Parallel()
		fixed := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
		result := getTimeRefString(&fixed)
		assert.Equal(t, "2026-01-15T10:30:00.000000Z", result)
	})
}

func TestGetStringRef(t *testing.T) {
	t.Parallel()

	t.Run("empty string returns nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, getStringRef(""))
	})

	t.Run("non-empty string returns pointer", func(t *testing.T) {
		t.Parallel()
		result := getStringRef("hello")
		require.NotNil(t, result)
		assert.Equal(t, "hello", *result)
	})
}

func TestGetIntRef(t *testing.T) {
	t.Parallel()

	result := getIntRef(42)
	require.NotNil(t, result)
	assert.Equal(t, 42, *result)
}

func TestGetBoolRef(t *testing.T) {
	t.Parallel()

	t.Run("true", func(t *testing.T) {
		t.Parallel()
		result := getBoolRef(true)
		require.NotNil(t, result)
		assert.True(t, *result)
	})

	t.Run("false", func(t *testing.T) {
		t.Parallel()
		result := getBoolRef(false)
		require.NotNil(t, result)
		assert.False(t, *result)
	})
}
