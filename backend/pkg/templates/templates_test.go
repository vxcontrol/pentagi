package templates

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Every row also proves ValidatePattern accepts what the pattern generated.
func TestTemplates_GenerateFromPattern_ProducesWhatItsPatternDescribes(t *testing.T) {
	randomPart := regexp.MustCompile(`\{r:\d+:`)

	tests := []struct {
		name         string
		pattern      string
		functionName string
		wantRegex    string
		wantLen      int
	}{
		{"an anthropic tool id", "toolu_{r:24:b}", "", `^toolu_[0-9A-Za-z]{24}$`, 30},
		{"an anthropic tooluse id", "tooluse_{r:22:b}", "", `^tooluse_[0-9A-Za-z]{22}$`, 30},
		{"an anthropic bedrock id", "toolu_bdrk_{r:24:b}", "", `^toolu_bdrk_[0-9A-Za-z]{24}$`, 35},
		{"an openai call id", "call_{r:24:x}", "", `^call_[a-zA-Z0-9]{24}$`, 29},
		{"an openai call id with a numeric prefix", "call_{r:2:d}_{r:24:x}", "", `^call_\d{2}_[a-zA-Z0-9]{24}$`, 32},
		{"a chatgpt tool id", "chatcmpl-tool-{r:32:h}", "", `^chatcmpl-tool-[0-9a-f]{32}$`, 46},
		{"a gemini tool id", "tool_{r:20:l}_{r:15:x}", "", `^tool_[a-z]{20}_[a-zA-Z0-9]{15}$`, 41},
		{"a short random id", "{r:9:b}", "", `^[0-9A-Za-z]{9}$`, 9},
		{"only digits", "id-{r:10:d}", "", `^id-\d{10}$`, 13},
		{"only lowercase", "key_{r:16:l}", "", `^key_[a-z]{16}$`, 20},
		{"only uppercase", "KEY_{r:8:u}", "", `^KEY_[A-Z]{8}$`, 12},
		{"uppercase hex", "0x{r:16:H}", "", `^0x[0-9A-F]{16}$`, 18},
		{"an empty pattern", "", "", `^$`, 0},
		{"only a literal", "fixed_string", "", `^fixed_string$`, 12},
		{"several random parts", "{r:4:u}-{r:4:u}-{r:4:u}-{r:12:h}", "", `^[A-Z]{4}-[A-Z]{4}-[A-Z]{4}-[0-9a-f]{12}$`, 27},
		{"adjacent random parts", "{r:4:d}{r:4:l}{r:4:u}", "", `^\d{4}[a-z]{4}[A-Z]{4}$`, 12},
		{"a malformed placeholder stays literal", "{r:invalid}", "", `^\{r:invalid\}$`, 11},
		{"a function name and a digit", "{f}:{r:1:d}", "get_number", `^get_number:\d$`, 12},
		{"an empty function name falls back to function", "{f}:{r:1:d}", "", `^function:\d$`, 10},
		{"a function name and random hex", "{f}_{r:8:h}", "call_tool", `^call_tool_[0-9a-f]{8}$`, 18},
		{"only a function name", "{f}", "test_func", `^test_func$`, 9},
		{"a function name between literals", "prefix_{f}_suffix", "my_tool", `^prefix_my_tool_suffix$`, 21},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := regexp.MustCompile(tt.wantRegex)
			seen := map[string]bool{}
			samples := make([]PatternSample, 0, 10)
			for range 10 {
				got := GenerateFromPattern(tt.pattern, tt.functionName)
				assert.Len(t, got, tt.wantLen, "result %q", got)
				assert.Regexp(t, re, got)
				seen[got] = true
				samples = append(samples, PatternSample{Value: got, FunctionName: tt.functionName})
			}

			if randomPart.MatchString(tt.pattern) && tt.wantLen > 1 {
				assert.Greater(t, len(seen), 1, "every generated value is identical; the random parts are not random")
			}
			assert.NoError(t, ValidatePattern(tt.pattern, samples), "a generated value must satisfy its own pattern")
		})
	}
}

func TestTemplates_ValidatePattern_NamesWhereASampleDiverges(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		samples []PatternSample
		wantErr []string
	}{
		{
			name:    "valid anthropic ids",
			pattern: "toolu_{r:24:b}",
			samples: []PatternSample{{Value: "toolu_013wc5CxNCjWGN2rsAR82rJK"}, {Value: "toolu_9ZxY8WvU7tS6rQ5pO4nM3lK2"}},
		},
		{
			name:    "valid openai ids",
			pattern: "call_{r:24:x}",
			samples: []PatternSample{{Value: "call_Z8ofZnYOCeOnpu0h2auwOgeR"}, {Value: "call_aBc123XyZ456MnO789PqR012"}},
		},
		{
			name:    "a valid hex id",
			pattern: "chatcmpl-tool-{r:32:h}",
			samples: []PatternSample{{Value: "chatcmpl-tool-23c5c0da71854f9bbd8774f7d0113a69"}},
		},
		{
			name:    "valid mixed parts",
			pattern: "prefix_{r:4:d}_{r:8:l}_suffix",
			samples: []PatternSample{{Value: "prefix_1234_abcdefgh_suffix"}, {Value: "prefix_9876_zyxwvuts_suffix"}},
		},
		{name: "no samples", pattern: "toolu_{r:24:b}", samples: []PatternSample{}},
		{
			name:    "a sample too short",
			pattern: "toolu_{r:24:b}",
			samples: []PatternSample{{Value: "toolu_123"}},
			wantErr: []string{"incorrect length: expected 30, got 9"},
		},
		{
			name:    "a sample too long",
			pattern: "call_{r:24:x}",
			samples: []PatternSample{{Value: "call_Z8ofZnYOCeOnpu0h2auwOgeRXXXXX"}},
			wantErr: []string{"incorrect length"},
		},
		{
			name:    "a wrong prefix",
			pattern: "toolu_{r:24:b}",
			samples: []PatternSample{{Value: "wrong_013wc5CxNCjWGN2rsAR82rJK"}},
			wantErr: []string{"position 0", "expected 'toolu_'", "pattern mismatch"},
		},
		{
			name:    "a letter among digits",
			pattern: "id_{r:10:d}",
			samples: []PatternSample{{Value: "id_123abc7890"}},
			wantErr: []string{"position 6", "[0-9]", "got 'a'"},
		},
		{
			name:    "uppercase in lowercase hex",
			pattern: "hex_{r:8:h}",
			samples: []PatternSample{{Value: "hex_ABCD1234"}},
			wantErr: []string{"pattern mismatch"},
		},
		{
			name:    "lowercase among uppercase",
			pattern: "KEY_{r:8:u}",
			samples: []PatternSample{{Value: "KEY_ABCDefgh"}},
			wantErr: []string{"pattern mismatch"},
		},
		{
			name:    "one bad sample among good ones",
			pattern: "toolu_{r:24:b}",
			samples: []PatternSample{{Value: "toolu_013wc5CxNCjWGN2rsAR82rJK"}, {Value: "invalid_string"}},
			wantErr: []string{"incorrect length"},
		},
		{
			name:    "a literal-only pattern matched",
			pattern: "fixed_string",
			samples: []PatternSample{{Value: "fixed_string"}, {Value: "fixed_string"}},
		},
		{
			name:    "a literal-only pattern missed",
			pattern: "fixed_string",
			samples: []PatternSample{{Value: "wrong_string"}},
			wantErr: []string{"pattern mismatch"},
		},
		{name: "a zero-length random part", pattern: "prefix_{r:0:b}_suffix", samples: []PatternSample{{Value: "prefix__suffix"}}},
		{
			name:    "a valid multi-part id",
			pattern: "{r:4:u}-{r:4:u}-{r:4:u}-{r:12:h}",
			samples: []PatternSample{{Value: "ABCD-EFGH-IJKL-0123456789ab"}},
		},
		{
			name:    "a bad section in a multi-part id",
			pattern: "{r:4:u}-{r:4:u}-{r:4:u}-{r:12:h}",
			samples: []PatternSample{{Value: "ABCD-EfGH-IJKL-0123456789ab"}},
			wantErr: []string{"pattern mismatch"},
		},
		{
			name:    "each sample brings its own function name",
			pattern: "{f}:{r:1:d}",
			samples: []PatternSample{
				{Value: "get_number:0", FunctionName: "get_number"},
				{Value: "submit_pattern:5", FunctionName: "submit_pattern"},
			},
		},
		{
			name:    "a function name and hex",
			pattern: "{f}_{r:8:h}",
			samples: []PatternSample{{Value: "call_tool_abc12345", FunctionName: "call_tool"}},
		},
		{
			name:    "a sample carrying another function name",
			pattern: "{f}:{r:1:d}",
			samples: []PatternSample{{Value: "wrong_name:0", FunctionName: "get_number"}},
			wantErr: []string{"pattern mismatch"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePattern(tt.pattern, tt.samples)
			if len(tt.wantErr) == 0 {
				assert.NoError(t, err)
				return
			}
			for _, part := range tt.wantErr {
				assert.ErrorContains(t, err, part)
			}
		})
	}
}
