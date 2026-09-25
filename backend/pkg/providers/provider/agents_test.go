package provider

import (
	"testing"

	"pentagi/pkg/templates"

	"github.com/stretchr/testify/assert"
)

func TestAgents_FallbackHeuristicDetection_BuildsThePatternTheSamplesShare(t *testing.T) {
	t.Parallel()

	samples := func(values ...string) []templates.PatternSample {
		out := make([]templates.PatternSample, 0, len(values))
		for _, value := range values {
			out = append(out, templates.PatternSample{Value: value})
		}
		return out
	}

	tests := []struct {
		name     string
		samples  []templates.PatternSample
		expected string
	}{
		{"anthropic tool ids", samples(
			"toolu_013wc5CxNCjWGN2rsAR82rJK", "toolu_9ZxY8WvU7tS6rQ5pO4nM3lK2", "toolu_aBcDeFgHiJkLmNoPqRsTuVwX",
		), "toolu_{r:24:b}"},
		{"openai call ids", samples(
			"call_Z8ofZnYOCeOnpu0h2auwOgeR", "call_aBc123XyZ456MnO789PqR012", "call_XyZ9AbC8dEf7GhI6jKl5MnO4",
		), "call_{r:24:b}"},
		{"hex ids", samples(
			"chatcmpl-tool-23c5c0da71854f9bbd8774f7d0113a69",
			"chatcmpl-tool-456789abcdef0123456789abcdef0123",
			"chatcmpl-tool-fedcba9876543210fedcba9876543210",
		), "chatcmpl-tool-{r:32:h}"},
		{"literals between random runs", samples(
			"prefix_1234_abcdefgh_suffix", "prefix_5678_zyxwvuts_suffix", "prefix_9012_qponmlkj_suffix",
		), "prefix_{r:4:d}_{r:8:l}_suffix"},
		{"short ids with no literal part", samples("qGGHVb8Pm", "c9nzLUf4t", "XyZ9AbC8d"), "{r:9:b}"},
		{"digits only", samples("id_1234567890", "id_9876043210", "id_5551235555"), "id_{r:10:d}"},
		{"upper case only", samples("KEY_ABCDEFGH", "KEY_ZYXWVUTS", "KEY_QPONMLKJ"), "KEY_{r:8:u}"},
		{"no samples", samples(), ""},
		{"a single sample is all literal", samples("test_123abc"), "test_123abc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, fallbackHeuristicDetection(tc.samples))
		})
	}
}

func TestAgents_DetermineMinimalCharset_PicksTheNarrowestClassOfTheCharacters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		chars    string
		expected string
	}{
		{"digits", "12345", "d"},
		{"lower case", "abcde", "l"},
		{"upper case", "ABCDE", "u"},
		{"mixed case letters", "aBcDe", "a"},
		{"lower case hex", "01abf", "h"},
		{"upper case hex", "01ABF", "H"},
		{"digits and both cases", "09azAZ", "b"},
		{"digits and lower case beyond hex", "05az", "x"},
		{"digits and upper case beyond hex", "05AZ", "x"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, determineMinimalCharset([]byte(tc.chars)))
		})
	}
}

func TestAgents_DetermineCommonCharset_CoversEveryPosition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		chars    []string
		expected string
	}{
		{"digits at every position", []string{"123", "456", "789"}, "d"},
		{"lower case hex across positions", []string{"abc", "def", "012"}, "h"},
		{"both cases and digits across positions", []string{"aBc", "DeF", "012"}, "b"},
		{"lower case beyond hex", []string{"abc", "xyz"}, "l"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			positions := make([][]byte, 0, len(tc.chars))
			for _, chars := range tc.chars {
				positions = append(positions, []byte(chars))
			}

			assert.Equal(t, tc.expected, determineCommonCharset(positions))
		})
	}
}
