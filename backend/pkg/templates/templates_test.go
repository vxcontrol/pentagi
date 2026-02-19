package templates_test

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"pentagi/pkg/templates"
	"pentagi/pkg/templates/validator"
)

// TestPromptTemplatesIntegrity validates all prompt templates against their declared variables
func TestPromptTemplatesIntegrity(t *testing.T) {
	defaultPrompts, err := templates.GetDefaultPrompts()
	if err != nil {
		t.Fatalf("Failed to load default prompts: %v", err)
	}

	// Use reflection to iterate over all prompts in the structure
	agents := validatePromptsStructure(t, reflect.ValueOf(defaultPrompts.AgentsPrompts), "AgentsPrompts")
	tools := validatePromptsStructure(t, reflect.ValueOf(defaultPrompts.ToolsPrompts), "ToolsPrompts")

	// According to the code, structure AgentsPrompts should have 27 prompts
	if agents > 27 {
		t.Fatalf("agents prompts amount is %d, expected 27", agents)
	}
	// According to the code, structure ToolsPrompts should have 9 prompts
	if tools > 9 {
		t.Fatalf("tools prompts amount is %d, expected 9", tools)
	}
}

// validatePromptsStructure recursively validates prompt structures using reflection
func validatePromptsStructure(t *testing.T, v reflect.Value, structName string) int {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return 0
	}

	count := 0
	vType := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := vType.Field(i)
		fieldName := fmt.Sprintf("%s.%s", structName, fieldType.Name)

		switch field.Kind() {
		case reflect.Struct:
			switch field.Type().Name() {
			case "AgentPrompt":
				// Single system prompt
				systemPrompt := field.FieldByName("System")
				if systemPrompt.IsValid() {
					count += validateSinglePrompt(t, systemPrompt, fieldName+".System")
				}
			case "AgentPrompts":
				// System and human prompts
				systemPrompt := field.FieldByName("System")
				humanPrompt := field.FieldByName("Human")
				if systemPrompt.IsValid() {
					count += validateSinglePrompt(t, systemPrompt, fieldName+".System")
				}
				if humanPrompt.IsValid() {
					count += validateSinglePrompt(t, humanPrompt, fieldName+".Human")
				}
			case "Prompt":
				// Direct prompt
				count += validateSinglePrompt(t, field, fieldName)
			default:
				// Recurse into nested structures
				count += validatePromptsStructure(t, field, fieldName)
			}
		}
	}

	return count
}

// validateSinglePrompt validates a single Prompt struct
func validateSinglePrompt(t *testing.T, promptValue reflect.Value, fieldName string) int {
	if promptValue.Kind() == reflect.Ptr {
		promptValue = promptValue.Elem()
	}

	typeField := promptValue.FieldByName("Type")
	templateField := promptValue.FieldByName("Template")
	variablesField := promptValue.FieldByName("Variables")

	if !typeField.IsValid() || !templateField.IsValid() || !variablesField.IsValid() {
		return 0
	}

	successed := 0
	promptType := typeField.Interface().(templates.PromptType)
	template := templateField.String()
	declaredVars := variablesField.Interface().([]string)

	t.Run(fmt.Sprintf("Validate_%s", promptType), func(t *testing.T) {
		// Test 1: Template should not be empty
		if strings.TrimSpace(template) == "" {
			t.Errorf("Template for %s (%s) is empty", promptType, fieldName)
			return
		}

		// Test 2: Template should parse without errors using validator package
		actualVars, err := validator.ExtractTemplateVariables(template)
		if err != nil {
			t.Errorf("Failed to parse template for %s (%s): %v", promptType, fieldName, err)
			return
		}

		// Test 3: Declared variables must match actual template usage
		expectedVars := make([]string, len(declaredVars))
		copy(expectedVars, declaredVars)
		sort.Strings(expectedVars)

		// Check for variables used in template but not declared
		var undeclared []string
		declaredSet := make(map[string]bool)
		for _, v := range declaredVars {
			declaredSet[v] = true
		}

		for _, v := range actualVars {
			if !declaredSet[v] {
				undeclared = append(undeclared, v)
			}
		}

		if len(undeclared) > 0 {
			t.Errorf("Template %s (%s) uses undeclared variables: %v", promptType, fieldName, undeclared)
			return
		}

		// Check for variables declared but not used in template
		var unused []string
		actualSet := make(map[string]bool)
		for _, v := range actualVars {
			actualSet[v] = true
		}

		for _, v := range declaredVars {
			if !actualSet[v] {
				unused = append(unused, v)
			}
		}

		if len(unused) > 0 {
			t.Errorf("Template %s (%s) declares unused variables: %v", promptType, fieldName, unused)
			return
		}

		// Test 4: Verify declared variables from promptVariables map match the prompt's Variables field
		expectedFromMap, exists := templates.PromptVariables[promptType]
		if !exists {
			t.Errorf("PromptType %s not found in promptVariables map", promptType)
			return
		}

		if !reflect.DeepEqual(expectedFromMap, declaredVars) {
			t.Errorf("Variables mismatch for %s (%s):\n  promptVariables: %v\n  prompt.Variables: %v",
				promptType, fieldName, expectedFromMap, declaredVars)
			return
		}

		successed = 1
	})

	return successed
}

// TestPromptVariablesCompleteness ensures all PromptTypes have corresponding entries in promptVariables
func TestPromptVariablesCompleteness(t *testing.T) {
	// Get all declared PromptType constants by checking defaultPrompts structure
	defaultPrompts, err := templates.GetDefaultPrompts()
	if err != nil {
		t.Fatalf("Failed to load default prompts: %v", err)
	}

	allPromptTypes := make(map[templates.PromptType]bool)
	collectPromptTypes(reflect.ValueOf(defaultPrompts), allPromptTypes)

	// Verify each PromptType has an entry in promptVariables
	for promptType := range allPromptTypes {
		if _, exists := templates.PromptVariables[promptType]; !exists {
			t.Errorf("PromptType %s missing from promptVariables map", promptType)
		}
	}

	// Verify no extra entries in promptVariables
	for promptType := range templates.PromptVariables {
		if !allPromptTypes[promptType] {
			t.Errorf("promptVariables contains unused PromptType: %s", promptType)
		}
	}
}

// collectPromptTypes recursively collects all PromptType values from the prompts structure
func collectPromptTypes(v reflect.Value, types map[templates.PromptType]bool) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		switch field.Kind() {
		case reflect.Struct:
			// Check if this struct has a Type field of PromptType
			typeField := field.FieldByName("Type")
			if typeField.IsValid() && typeField.Type().String() == "templates.PromptType" {
				promptType := typeField.Interface().(templates.PromptType)
				types[promptType] = true
			} else {
				// Recurse into nested structures
				collectPromptTypes(field, types)
			}
		}
	}
}

// TestTemplateRenderability ensures all templates can be rendered with dummy data
func TestTemplateRenderability(t *testing.T) {
	defaultPrompts, err := templates.GetDefaultPrompts()
	if err != nil {
		t.Fatalf("Failed to load default prompts: %v", err)
	}

	// Create dummy data for all known variable names
	dummyData := validator.CreateDummyTemplateData()

	testRenderability(t, reflect.ValueOf(defaultPrompts), dummyData, "DefaultPrompts")
}

// testRenderability recursively tests if all prompts can be rendered with dummy data
func testRenderability(t *testing.T, v reflect.Value, dummyData map[string]any, structName string) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	vType := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := vType.Field(i)
		fieldName := fmt.Sprintf("%s.%s", structName, fieldType.Name)

		if field.Kind() == reflect.Struct {
			typeField := field.FieldByName("Type")
			templateField := field.FieldByName("Template")

			if typeField.IsValid() && templateField.IsValid() {
				promptType := typeField.Interface().(templates.PromptType)
				template := templateField.String()

				t.Run(fmt.Sprintf("Render_%s", promptType), func(t *testing.T) {
					_, err := templates.RenderPrompt(string(promptType), template, dummyData)
					if err != nil {
						t.Errorf("Failed to render template %s (%s): %v", promptType, fieldName, err)
					}
				})
			} else {
				// Recurse into nested structures
				testRenderability(t, field, dummyData, fieldName)
			}
		}
	}
}

// TestGenerateFromPattern tests random string generation from pattern templates
func TestGenerateFromPattern(t *testing.T) {
	testCases := []struct {
		name          string
		pattern       string
		functionName  string
		expectedRegex string
		expectedLen   int
	}{
		{
			name:          "anthropic_tool_id",
			pattern:       "toolu_{r:24:b}",
			functionName:  "",
			expectedRegex: `^toolu_[0-9A-Za-z]{24}$`,
			expectedLen:   30,
		},
		{
			name:          "anthropic_tooluse_id",
			pattern:       "tooluse_{r:22:b}",
			functionName:  "",
			expectedRegex: `^tooluse_[0-9A-Za-z]{22}$`,
			expectedLen:   30,
		},
		{
			name:          "anthropic_bedrock_id",
			pattern:       "toolu_bdrk_{r:24:b}",
			functionName:  "",
			expectedRegex: `^toolu_bdrk_[0-9A-Za-z]{24}$`,
			expectedLen:   35,
		},
		{
			name:          "openai_call_id",
			pattern:       "call_{r:24:x}",
			functionName:  "",
			expectedRegex: `^call_[a-zA-Z0-9]{24}$`,
			expectedLen:   29,
		},
		{
			name:          "openai_call_id_with_prefix",
			pattern:       "call_{r:2:d}_{r:24:x}",
			functionName:  "",
			expectedRegex: `^call_\d{2}_[a-zA-Z0-9]{24}$`,
			expectedLen:   32,
		},
		{
			name:          "chatgpt_tool_id",
			pattern:       "chatcmpl-tool-{r:32:h}",
			functionName:  "",
			expectedRegex: `^chatcmpl-tool-[0-9a-f]{32}$`,
			expectedLen:   46,
		},
		{
			name:          "gemini_tool_id",
			pattern:       "tool_{r:20:l}_{r:15:x}",
			functionName:  "",
			expectedRegex: `^tool_[a-z]{20}_[a-zA-Z0-9]{15}$`,
			expectedLen:   41,
		},
		{
			name:          "short_random_id",
			pattern:       "{r:9:b}",
			functionName:  "",
			expectedRegex: `^[0-9A-Za-z]{9}$`,
			expectedLen:   9,
		},
		{
			name:          "only_digits",
			pattern:       "id-{r:10:d}",
			functionName:  "",
			expectedRegex: `^id-\d{10}$`,
			expectedLen:   13,
		},
		{
			name:          "only_lowercase",
			pattern:       "key_{r:16:l}",
			functionName:  "",
			expectedRegex: `^key_[a-z]{16}$`,
			expectedLen:   20,
		},
		{
			name:          "only_uppercase",
			pattern:       "KEY_{r:8:u}",
			functionName:  "",
			expectedRegex: `^KEY_[A-Z]{8}$`,
			expectedLen:   12,
		},
		{
			name:          "hex_uppercase",
			pattern:       "0x{r:16:H}",
			functionName:  "",
			expectedRegex: `^0x[0-9A-F]{16}$`,
			expectedLen:   18,
		},
		{
			name:          "empty_pattern",
			pattern:       "",
			functionName:  "",
			expectedRegex: `^$`,
			expectedLen:   0,
		},
		{
			name:          "only_literal",
			pattern:       "fixed_string",
			functionName:  "",
			expectedRegex: `^fixed_string$`,
			expectedLen:   12,
		},
		{
			name:          "multiple_random_parts",
			pattern:       "{r:4:u}-{r:4:u}-{r:4:u}-{r:12:h}",
			functionName:  "",
			expectedRegex: `^[A-Z]{4}-[A-Z]{4}-[A-Z]{4}-[0-9a-f]{12}$`,
			expectedLen:   27, // 4 + 1 + 4 + 1 + 4 + 1 + 12 = 27
		},
		{
			name:          "function_with_digit",
			pattern:       "{f}:{r:1:d}",
			functionName:  "get_number",
			expectedRegex: `^get_number:\d$`,
			expectedLen:   12,
		},
		{
			name:          "function_with_random",
			pattern:       "{f}_{r:8:h}",
			functionName:  "call_tool",
			expectedRegex: `^call_tool_[0-9a-f]{8}$`,
			expectedLen:   18,
		},
		{
			name:          "function_only",
			pattern:       "{f}",
			functionName:  "test_func",
			expectedRegex: `^test_func$`,
			expectedLen:   9,
		},
		{
			name:          "function_with_prefix_suffix",
			pattern:       "prefix_{f}_suffix",
			functionName:  "my_tool",
			expectedRegex: `^prefix_my_tool_suffix$`,
			expectedLen:   21,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate multiple times to ensure randomness
			for i := 0; i < 5; i++ {
				result := templates.GenerateFromPattern(tc.pattern, tc.functionName)

				// Check length
				if len(result) != tc.expectedLen {
					t.Errorf("Expected length %d, got %d for result '%s'", tc.expectedLen, len(result), result)
				}

				// Check pattern match
				re := regexp.MustCompile(tc.expectedRegex)
				if !re.MatchString(result) {
					t.Errorf("Result '%s' doesn't match expected regex '%s'", result, tc.expectedRegex)
				}
			}

			// Check that multiple generations produce different results (for non-empty random parts)
			if strings.Contains(tc.pattern, "{r:") && tc.expectedLen > 0 {
				results := make(map[string]bool)
				for i := 0; i < 10; i++ {
					results[templates.GenerateFromPattern(tc.pattern, tc.functionName)] = true
				}
				// At least some variance expected (not all identical)
				if len(results) == 1 && tc.expectedLen > 1 {
					t.Error("All generated values are identical - randomness may be broken")
				}
			}
		})
	}
}

// TestValidatePattern tests pattern validation functionality
func TestValidatePattern(t *testing.T) {
	testCases := []struct {
		name        string
		pattern     string
		samples     []templates.PatternSample
		expectError bool
		errorSubstr string
	}{
		{
			name:    "valid_anthropic_ids",
			pattern: "toolu_{r:24:b}",
			samples: []templates.PatternSample{
				{Value: "toolu_013wc5CxNCjWGN2rsAR82rJK"},
				{Value: "toolu_9ZxY8WvU7tS6rQ5pO4nM3lK2"},
			},
			expectError: false,
		},
		{
			name:    "valid_openai_ids",
			pattern: "call_{r:24:x}",
			samples: []templates.PatternSample{
				{Value: "call_Z8ofZnYOCeOnpu0h2auwOgeR"},
				{Value: "call_aBc123XyZ456MnO789PqR012"},
			},
			expectError: false,
		},
		{
			name:    "valid_hex_ids",
			pattern: "chatcmpl-tool-{r:32:h}",
			samples: []templates.PatternSample{
				{Value: "chatcmpl-tool-23c5c0da71854f9bbd8774f7d0113a69"},
			},
			expectError: false,
		},
		{
			name:    "valid_mixed_pattern",
			pattern: "prefix_{r:4:d}_{r:8:l}_suffix",
			samples: []templates.PatternSample{
				{Value: "prefix_1234_abcdefgh_suffix"},
				{Value: "prefix_9876_zyxwvuts_suffix"},
			},
			expectError: false,
		},
		{
			name:        "empty_values",
			pattern:     "toolu_{r:24:b}",
			samples:     []templates.PatternSample{},
			expectError: false,
		},
		{
			name:    "invalid_length_too_short",
			pattern: "toolu_{r:24:b}",
			samples: []templates.PatternSample{
				{Value: "toolu_123"},
			},
			expectError: true,
			errorSubstr: "incorrect length",
		},
		{
			name:    "invalid_length_too_long",
			pattern: "call_{r:24:x}",
			samples: []templates.PatternSample{
				{Value: "call_Z8ofZnYOCeOnpu0h2auwOgeRXXXXX"},
			},
			expectError: true,
			errorSubstr: "incorrect length",
		},
		{
			name:    "invalid_prefix",
			pattern: "toolu_{r:24:b}",
			samples: []templates.PatternSample{
				{Value: "wrong_013wc5CxNCjWGN2rsAR82rJK"},
			},
			expectError: true,
			errorSubstr: "pattern mismatch",
		},
		{
			name:    "invalid_charset_has_special_chars",
			pattern: "id_{r:10:d}",
			samples: []templates.PatternSample{
				{Value: "id_123abc7890"},
			},
			expectError: true,
			errorSubstr: "pattern mismatch",
		},
		{
			name:    "invalid_hex_has_uppercase",
			pattern: "hex_{r:8:h}",
			samples: []templates.PatternSample{
				{Value: "hex_ABCD1234"},
			},
			expectError: true,
			errorSubstr: "pattern mismatch",
		},
		{
			name:    "invalid_uppercase_has_lowercase",
			pattern: "KEY_{r:8:u}",
			samples: []templates.PatternSample{
				{Value: "KEY_ABCDefgh"},
			},
			expectError: true,
			errorSubstr: "pattern mismatch",
		},
		{
			name:    "multiple_values_one_invalid",
			pattern: "toolu_{r:24:b}",
			samples: []templates.PatternSample{
				{Value: "toolu_013wc5CxNCjWGN2rsAR82rJK"},
				{Value: "invalid_string"},
			},
			expectError: true,
			errorSubstr: "incorrect length",
		},
		{
			name:    "literal_only_pattern_valid",
			pattern: "fixed_string",
			samples: []templates.PatternSample{
				{Value: "fixed_string"},
				{Value: "fixed_string"},
			},
			expectError: false,
		},
		{
			name:    "literal_only_pattern_invalid",
			pattern: "fixed_string",
			samples: []templates.PatternSample{
				{Value: "wrong_string"},
			},
			expectError: true,
			errorSubstr: "pattern mismatch",
		},
		{
			name:    "edge_case_zero_length_random",
			pattern: "prefix_{r:0:b}_suffix",
			samples: []templates.PatternSample{
				{Value: "prefix__suffix"},
			},
			expectError: false,
		},
		{
			name:    "complex_multi_part_valid",
			pattern: "{r:4:u}-{r:4:u}-{r:4:u}-{r:12:h}",
			samples: []templates.PatternSample{
				{Value: "ABCD-EFGH-IJKL-0123456789ab"},
			},
			expectError: false,
		},
		{
			name:    "complex_multi_part_invalid_section",
			pattern: "{r:4:u}-{r:4:u}-{r:4:u}-{r:12:h}",
			samples: []templates.PatternSample{
				{Value: "ABCD-EfGH-IJKL-0123456789ab"},
			},
			expectError: true,
			errorSubstr: "pattern mismatch",
		},
		{
			name:    "function_placeholder_different_names",
			pattern: "{f}:{r:1:d}",
			samples: []templates.PatternSample{
				{Value: "get_number:0", FunctionName: "get_number"},
				{Value: "submit_pattern:5", FunctionName: "submit_pattern"},
			},
			expectError: false,
		},
		{
			name:    "function_placeholder_valid",
			pattern: "{f}_{r:8:h}",
			samples: []templates.PatternSample{
				{Value: "call_tool_abc12345", FunctionName: "call_tool"},
			},
			expectError: false,
		},
		{
			name:    "function_placeholder_mismatch",
			pattern: "{f}:{r:1:d}",
			samples: []templates.PatternSample{
				{Value: "wrong_name:0", FunctionName: "get_number"},
			},
			expectError: true,
			errorSubstr: "pattern mismatch",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := templates.ValidatePattern(tc.pattern, tc.samples)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tc.errorSubstr != "" && !strings.Contains(err.Error(), tc.errorSubstr) {
					t.Errorf("Expected error to contain '%s', got: %v", tc.errorSubstr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestGenerateAndValidateRoundTrip tests that generated values validate correctly
func TestGenerateAndValidateRoundTrip(t *testing.T) {
	testCases := []struct {
		pattern      string
		functionName string
	}{
		{"toolu_{r:24:b}", ""},
		{"call_{r:24:x}", ""},
		{"chatcmpl-tool-{r:32:h}", ""},
		{"prefix_{r:4:d}_{r:8:l}_suffix", ""},
		{"{r:9:b}", ""},
		{"KEY_{r:8:u}", ""},
		{"{r:4:u}-{r:4:u}-{r:4:u}-{r:12:h}", ""},
		{"{f}:{r:1:d}", "test_function"},
		{"{f}_{r:8:h}", "my_tool"},
		{"prefix_{f}_suffix", "tool_name"},
	}

	for _, tc := range testCases {
		t.Run(tc.pattern, func(t *testing.T) {
			// Generate multiple samples
			samples := make([]templates.PatternSample, 10)
			for i := 0; i < 10; i++ {
				samples[i] = templates.PatternSample{
					Value:        templates.GenerateFromPattern(tc.pattern, tc.functionName),
					FunctionName: tc.functionName,
				}
			}

			// Validate all generated samples
			err := templates.ValidatePattern(tc.pattern, samples)
			if err != nil {
				t.Errorf("Generated values failed validation: %v\nSamples: %v", err, samples)
			}
		})
	}
}

// TestValidatePatternErrorDetails tests detailed error reporting
func TestValidatePatternErrorDetails(t *testing.T) {
	testCases := []struct {
		name            string
		pattern         string
		sample          templates.PatternSample
		expectedPos     int
		expectedInError []string
	}{
		{
			name:            "wrong_prefix",
			pattern:         "toolu_{r:10:b}",
			sample:          templates.PatternSample{Value: "wrong_0123456789"},
			expectedPos:     0,
			expectedInError: []string{"position 0", "'toolu_'"},
		},
		{
			name:            "invalid_char_in_random",
			pattern:         "id_{r:5:d}",
			sample:          templates.PatternSample{Value: "id_12a45"},
			expectedPos:     5,
			expectedInError: []string{"position 5", "0-9"},
		},
		{
			name:            "length_mismatch",
			pattern:         "key_{r:10:b}",
			sample:          templates.PatternSample{Value: "key_123"},
			expectedPos:     -1,
			expectedInError: []string{"incorrect length", "expected 14", "got 7"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := templates.ValidatePattern(tc.pattern, []templates.PatternSample{tc.sample})
			if err == nil {
				t.Fatal("Expected error but got nil")
			}

			errMsg := err.Error()
			for _, expected := range tc.expectedInError {
				if !strings.Contains(errMsg, expected) {
					t.Errorf("Error message should contain '%s', got: %s", expected, errMsg)
				}
			}
		})
	}
}

// TestPatternEdgeCases tests boundary and edge cases
func TestPatternEdgeCases(t *testing.T) {
	testCases := []struct {
		name    string
		pattern string
		test    func(t *testing.T, pattern string)
	}{
		{
			name:    "empty_pattern_generates_empty",
			pattern: "",
			test: func(t *testing.T, pattern string) {
				result := templates.GenerateFromPattern(pattern, "")
				if result != "" {
					t.Errorf("Expected empty string, got '%s'", result)
				}
			},
		},
		{
			name:    "only_literals_no_random",
			pattern: "completely_fixed_string",
			test: func(t *testing.T, pattern string) {
				result1 := templates.GenerateFromPattern(pattern, "")
				result2 := templates.GenerateFromPattern(pattern, "")
				if result1 != result2 {
					t.Error("Literal-only pattern should always produce same result")
				}
				if result1 != "completely_fixed_string" {
					t.Errorf("Expected 'completely_fixed_string', got '%s'", result1)
				}
			},
		},
		{
			name:    "consecutive_random_parts",
			pattern: "{r:4:d}{r:4:l}{r:4:u}",
			test: func(t *testing.T, pattern string) {
				result := templates.GenerateFromPattern(pattern, "")
				if len(result) != 12 {
					t.Errorf("Expected length 12, got %d", len(result))
				}
				// First 4 should be digits
				for i := 0; i < 4; i++ {
					if result[i] < '0' || result[i] > '9' {
						t.Errorf("Position %d should be digit, got '%c'", i, result[i])
					}
				}
				// Next 4 should be lowercase
				for i := 4; i < 8; i++ {
					if result[i] < 'a' || result[i] > 'z' {
						t.Errorf("Position %d should be lowercase, got '%c'", i, result[i])
					}
				}
				// Last 4 should be uppercase
				for i := 8; i < 12; i++ {
					if result[i] < 'A' || result[i] > 'Z' {
						t.Errorf("Position %d should be uppercase, got '%c'", i, result[i])
					}
				}
			},
		},
		{
			name:    "malformed_pattern_is_treated_as_literal",
			pattern: "{r:invalid}",
			test: func(t *testing.T, pattern string) {
				result := templates.GenerateFromPattern(pattern, "")
				// Malformed pattern should be treated as literal
				if result != "{r:invalid}" {
					t.Errorf("Expected literal '{r:invalid}', got '%s'", result)
				}
			},
		},
		{
			name:    "function_placeholder_with_empty_name",
			pattern: "{f}:{r:1:d}",
			test: func(t *testing.T, pattern string) {
				result := templates.GenerateFromPattern(pattern, "")
				// Empty function name should use "function" as fallback
				if !strings.HasPrefix(result, "function:") {
					t.Errorf("Expected prefix 'function:', got '%s'", result)
				}
			},
		},
		{
			name:    "function_placeholder_only",
			pattern: "{f}",
			test: func(t *testing.T, pattern string) {
				result := templates.GenerateFromPattern(pattern, "my_func")
				if result != "my_func" {
					t.Errorf("Expected 'my_func', got '%s'", result)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t, tc.pattern)
		})
	}
}
