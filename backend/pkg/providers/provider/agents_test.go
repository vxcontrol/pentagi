package provider

import (
	"testing"

	"pentagi/pkg/templates"
)

func TestFallbackHeuristicDetection(t *testing.T) {
	testCases := []struct {
		name     string
		samples  []templates.PatternSample
		expected string
	}{
		{
			name: "anthropic_tool_ids",
			samples: []templates.PatternSample{
				{Value: "toolu_013wc5CxNCjWGN2rsAR82rJK"},
				{Value: "toolu_9ZxY8WvU7tS6rQ5pO4nM3lK2"},
				{Value: "toolu_aBcDeFgHiJkLmNoPqRsTuVwX"},
			},
			expected: "toolu_{r:24:b}",
		},
		{
			name: "openai_call_ids",
			samples: []templates.PatternSample{
				{Value: "call_Z8ofZnYOCeOnpu0h2auwOgeR"},
				{Value: "call_aBc123XyZ456MnO789PqR012"},
				{Value: "call_XyZ9AbC8dEf7GhI6jKl5MnO4"},
			},
			expected: "call_{r:24:b}", // Contains all: digits, lower, upper = base62
		},
		{
			name: "hex_ids",
			samples: []templates.PatternSample{
				{Value: "chatcmpl-tool-23c5c0da71854f9bbd8774f7d0113a69"},
				{Value: "chatcmpl-tool-456789abcdef0123456789abcdef0123"},
				{Value: "chatcmpl-tool-fedcba9876543210fedcba9876543210"},
			},
			expected: "chatcmpl-tool-{r:32:h}",
		},
		{
			name: "mixed_pattern",
			samples: []templates.PatternSample{
				{Value: "prefix_1234_abcdefgh_suffix"},
				{Value: "prefix_5678_zyxwvuts_suffix"},
				{Value: "prefix_9012_qponmlkj_suffix"},
			},
			expected: "prefix_{r:4:d}_{r:8:l}_suffix",
		},
		{
			name: "short_ids",
			samples: []templates.PatternSample{
				{Value: "qGGHVb8Pm"},
				{Value: "c9nzLUf4t"},
				{Value: "XyZ9AbC8d"},
			},
			expected: "{r:9:b}",
		},
		{
			name: "only_digits",
			samples: []templates.PatternSample{
				{Value: "id_1234567890"},
				{Value: "id_9876043210"},
				{Value: "id_5551235555"},
			},
			expected: "id_{r:10:d}",
		},
		{
			name: "uppercase_only",
			samples: []templates.PatternSample{
				{Value: "KEY_ABCDEFGH"},
				{Value: "KEY_ZYXWVUTS"},
				{Value: "KEY_QPONMLKJ"},
			},
			expected: "KEY_{r:8:u}",
		},
		{
			name:     "empty_samples",
			samples:  []templates.PatternSample{},
			expected: "",
		},
		{
			name: "single_sample",
			samples: []templates.PatternSample{
				{Value: "test_123abc"},
			},
			expected: "test_123abc", // All literal when single sample
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := fallbackHeuristicDetection(tc.samples)
			if result != tc.expected {
				t.Errorf("Expected pattern '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestDetermineMinimalCharset(t *testing.T) {
	testCases := []struct {
		name     string
		chars    []byte
		expected string
	}{
		{
			name:     "only_digits",
			chars:    []byte{'1', '2', '3', '4', '5'},
			expected: "d",
		},
		{
			name:     "only_lowercase",
			chars:    []byte{'a', 'b', 'c', 'd', 'e'},
			expected: "l",
		},
		{
			name:     "only_uppercase",
			chars:    []byte{'A', 'B', 'C', 'D', 'E'},
			expected: "u",
		},
		{
			name:     "alpha_mixed",
			chars:    []byte{'a', 'B', 'c', 'D', 'e'},
			expected: "a",
		},
		{
			name:     "hex_lowercase",
			chars:    []byte{'0', '1', 'a', 'b', 'f'},
			expected: "h",
		},
		{
			name:     "hex_uppercase",
			chars:    []byte{'0', '1', 'A', 'B', 'F'},
			expected: "H",
		},
		{
			name:     "base62",
			chars:    []byte{'0', '9', 'a', 'z', 'A', 'Z'},
			expected: "b",
		},
		{
			name:     "alnum_with_all_types",
			chars:    []byte{'0', 'a', 'Z'},
			expected: "b", // has all three: digit, lower, upper = base62
		},
		{
			name:     "alnum_digit_lower_only",
			chars:    []byte{'0', '5', 'a', 'z'},
			expected: "x", // digit + lower but no upper = alnum
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := determineMinimalCharset(tc.chars)
			if result != tc.expected {
				t.Errorf("Expected charset '%s', got '%s'", tc.expected, result)
			}
		})
	}
}
