package langfuse

import (
	"maps"
	"testing"
	"time"

	"pentagi/pkg/observability/langfuse/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestHelpers_MergeMaps_LetsSrcWinWithoutTouchingDst(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		dst      map[string]any
		src      map[string]any
		expected map[string]any
	}{
		{name: "both nil"},
		{name: "src nil returns dst", dst: map[string]any{"a": 1}, expected: map[string]any{"a": 1}},
		{name: "dst nil copies src", src: map[string]any{"b": 2}, expected: map[string]any{"b": 2}},
		{name: "disjoint keys", dst: map[string]any{"a": 1}, src: map[string]any{"b": 2}, expected: map[string]any{"a": 1, "b": 2}},
		{
			name:     "overlapping keys src overrides",
			dst:      map[string]any{"a": 1, "b": 2},
			src:      map[string]any{"b": 99, "c": 3},
			expected: map[string]any{"a": 1, "b": 99, "c": 3},
		},
		{name: "empty maps", dst: map[string]any{}, src: map[string]any{}, expected: map[string]any{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dstBefore := maps.Clone(tt.dst)
			assert.Equal(t, tt.expected, mergeMaps(tt.dst, tt.src))
			assert.Equal(t, dstBefore, tt.dst, "dst must not be mutated")
		})
	}
}

func TestHelpers_ObservationLevel_MapsUnknownToDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		level    ObservationLevel
		expected api.ObservationLevel
	}{
		{"the default level maps to DEFAULT", ObservationLevelDefault, "DEFAULT"},
		{"the debug level maps to DEBUG", ObservationLevelDebug, "DEBUG"},
		{"the warning level maps to WARNING", ObservationLevelWarning, "WARNING"},
		{"the error level maps to ERROR", ObservationLevelError, "ERROR"},
		{"an unknown level falls back to DEFAULT", ObservationLevel(99), "DEFAULT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.level.ToLangfuse()
			require.NotNil(t, result)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// Each row checks String and ToLangfuse, which is nil exactly when String is empty.
func TestHelpers_GenerationUsageUnit_NamesEveryUnitAndNothingElse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		unit     GenerationUsageUnit
		expected string
	}{
		{"the tokens unit is named TOKENS", GenerationUsageUnitTokens, "TOKENS"},
		{"the characters unit is named CHARACTERS", GenerationUsageUnitCharacters, "CHARACTERS"},
		{"the milliseconds unit is named MILLISECONDS", GenerationUsageUnitMilliseconds, "MILLISECONDS"},
		{"the seconds unit is named SECONDS", GenerationUsageUnitSeconds, "SECONDS"},
		{"the images unit is named IMAGES", GenerationUsageUnitImages, "IMAGES"},
		{"the requests unit is named REQUESTS", GenerationUsageUnitRequests, "REQUESTS"},
		{"an unknown unit has no name", GenerationUsageUnit(99), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.unit.String())
			if tt.expected == "" {
				assert.Nil(t, tt.unit.ToLangfuse())
				return
			}
			require.NotNil(t, tt.unit.ToLangfuse())
			assert.Equal(t, tt.expected, *tt.unit.ToLangfuse())
		})
	}
}

func TestHelpers_GenerationUsage_SumsTotalsAndCosts(t *testing.T) {
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
		result := (&GenerationUsage{Input: 100, InputCost: &inputCost}).ToLangfuse()
		require.NotNil(t, result.Usage.TotalCost)
		assert.InDelta(t, 0.01, *result.Usage.TotalCost, 1e-9)
	})

	t.Run("output cost only", func(t *testing.T) {
		t.Parallel()
		outputCost := 0.02
		result := (&GenerationUsage{Output: 50, OutputCost: &outputCost}).ToLangfuse()
		require.NotNil(t, result.Usage.TotalCost)
		assert.InDelta(t, 0.02, *result.Usage.TotalCost, 1e-9)
	})

	t.Run("both costs summed", func(t *testing.T) {
		t.Parallel()
		inputCost, outputCost := 0.01, 0.02
		u := &GenerationUsage{Input: 100, Output: 50, InputCost: &inputCost, OutputCost: &outputCost, Unit: GenerationUsageUnitTokens}
		result := u.ToLangfuse()
		require.NotNil(t, result.Usage.TotalCost)
		assert.InDelta(t, 0.03, *result.Usage.TotalCost, 1e-9)
		require.NotNil(t, result.Usage.InputCost)
		assert.Equal(t, 0.01, *result.Usage.InputCost)
		require.NotNil(t, result.Usage.OutputCost)
		assert.Equal(t, 0.02, *result.Usage.OutputCost)
	})
}

// helpersPlainValues unwraps each MapValue into the one Go value it holds, so "1024" and 1024 differ.
func helpersPlainValues(t *testing.T, values map[string]*api.MapValue) map[string]any {
	t.Helper()

	plain := make(map[string]any, len(values))
	for key, v := range values {
		switch {
		case v.GetStringOptional() != nil:
			plain[key] = *v.GetStringOptional()
		case v.GetIntegerOptional() != nil:
			plain[key] = *v.GetIntegerOptional()
		case v.GetBooleanOptional() != nil:
			plain[key] = *v.GetBooleanOptional()
		case v.GetStringListOptional() != nil:
			plain[key] = v.GetStringListOptional()
		default:
			t.Fatalf("%s holds no value", key)
		}
	}
	return plain
}

func TestHelpers_ModelParameters_ConvertsEachFieldToItsLangfuseType(t *testing.T) {
	t.Parallel()

	t.Run("nil receiver returns nil", func(t *testing.T) {
		t.Parallel()
		var m *ModelParameters
		assert.Nil(t, m.ToLangfuse())
	})

	t.Run("an empty struct reports max_tokens as inf", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, map[string]any{"max_tokens": "inf"}, helpersPlainValues(t, (&ModelParameters{}).ToLangfuse()))
	})

	t.Run("floats become one-decimal strings and integers stay integers", func(t *testing.T) {
		t.Parallel()
		temp, topP, minP := 0.5, 0.9, 0.1
		repPenalty, freqPenalty, presPenalty := 1.1, 0.5, 0.6
		topK, seed, maxTokens, candidateCount, minLen, maxLen, n := 40, 42, 2048, 3, 10, 500, 2
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
			StopWords:         []string{"END", "STOP"},
		}

		assert.Equal(t, map[string]any{
			"temperature":        "0.5",
			"top_p":              "0.9",
			"min_p":              "0.1",
			"top_k":              40,
			"seed":               42,
			"max_tokens":         2048,
			"candidate_count":    3,
			"min_length":         10,
			"max_length":         500,
			"n":                  2,
			"repetition_penalty": "1.1",
			"frequency_penalty":  "0.5",
			"presence_penalty":   "0.6",
			"json":               true,
			"stop_words":         []string{"END", "STOP"},
		}, helpersPlainValues(t, m.ToLangfuse()))
	})
}

func TestHelpers_GetLangchainModelParameters_ReadsTheCallOptions(t *testing.T) {
	t.Parallel()

	assert.Nil(t, GetLangchainModelParameters(nil))
	assert.Nil(t, GetLangchainModelParameters([]llms.CallOption{}))

	result := GetLangchainModelParameters([]llms.CallOption{llms.WithTemperature(0.7), llms.WithMaxTokens(512)})
	require.NotNil(t, result)
	require.NotNil(t, result.Temperature)
	assert.InDelta(t, 0.7, *result.Temperature, 1e-9)
	require.NotNil(t, result.MaxTokens)
	assert.Equal(t, 512, *result.MaxTokens)
}

func TestHelpers_NewTraceID_IsUniqueW3CHex(t *testing.T) {
	t.Parallel()

	ids := make(map[string]bool)
	for range 100 {
		id := newTraceID()
		assert.Regexp(t, `^[0-9a-f]{32}$`, id)
		assert.False(t, ids[id], "duplicate trace ID generated")
		ids[id] = true
	}
}

func TestHelpers_NewSpanID_IsUniqueW3CHex(t *testing.T) {
	t.Parallel()

	ids := make(map[string]bool)
	for range 100 {
		id := newSpanID()
		assert.Regexp(t, `^[0-9a-f]{16}$`, id)
		assert.False(t, ids[id], "duplicate span ID generated")
		ids[id] = true
	}
}

func TestHelpers_GetCurrentTimeRef_PointsAtTheCurrentUTCTime(t *testing.T) {
	t.Parallel()

	ref := getCurrentTimeRef()
	require.NotNil(t, ref)
	assert.Equal(t, time.UTC, ref.Location(), "must be UTC")
	assert.WithinDuration(t, time.Now().UTC(), *ref, time.Minute)
}

func TestHelpers_GetTimeRefString_FormatsTheGivenOrCurrentTime(t *testing.T) {
	t.Parallel()

	fixed := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	assert.Equal(t, "2026-01-15T10:30:00.000000Z", getTimeRefString(&fixed))

	current, err := time.Parse(timeFormat8601, getTimeRefString(nil))
	require.NoError(t, err, "a nil time formats the current time")
	assert.WithinDuration(t, time.Now().UTC(), current, time.Minute)
}

// Subtests are keyed by unit.
func TestHelpers_RefHelpersPointAtTheirArgument(t *testing.T) {
	t.Parallel()

	t.Run("an empty string becomes nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, getStringRef(""))
		require.NotNil(t, getStringRef("hello"))
		assert.Equal(t, "hello", *getStringRef("hello"))
	})

	t.Run("an int keeps its value", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, getIntRef(42))
		assert.Equal(t, 42, *getIntRef(42))
	})

	t.Run("false stays false rather than nil", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, getBoolRef(false))
		assert.False(t, *getBoolRef(false))
		assert.True(t, *getBoolRef(true))
	})
}
