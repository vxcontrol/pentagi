package validator_test

import (
	"sort"
	"strings"
	"testing"

	"pentagi/pkg/templates"
	"pentagi/pkg/templates/validator"
)

// TestDummyDataCompleteness verifies that createDummyTemplateData contains all variables from PromptVariables
func TestDummyDataCompleteness(t *testing.T) {
	// Extract all unique variables from PromptVariables map
	allVariables := make(map[string]bool)
	for _, variables := range templates.PromptVariables {
		for _, variable := range variables {
			allVariables[variable] = true
		}
	}

	// Get dummy data
	dummyData := validator.CreateDummyTemplateData()

	// Check that all variables from PromptVariables exist in dummy data
	var missingVars []string
	for variable := range allVariables {
		if _, exists := dummyData[variable]; !exists {
			missingVars = append(missingVars, variable)
		}
	}

	if len(missingVars) > 0 {
		sort.Strings(missingVars)
		t.Errorf("createDummyTemplateData() is missing variables declared in PromptVariables: %v", missingVars)
	}

	// Check for potentially unused variables in dummy data (optional warning)
	var unusedVars []string
	for variable := range dummyData {
		if !allVariables[variable] {
			unusedVars = append(unusedVars, variable)
		}
	}

	if len(unusedVars) > 0 {
		sort.Strings(unusedVars)
		t.Logf("WARNING: createDummyTemplateData() contains variables not declared in PromptVariables: %v", unusedVars)
	}

	t.Logf("Total variables in PromptVariables: %d", len(allVariables))
	t.Logf("Total variables in createDummyTemplateData: %d", len(dummyData))
}

// TestExtractTemplateVariables tests the AST-based variable extraction
func TestExtractTemplateVariables(t *testing.T) {
	testCases := []struct {
		name      string
		template  string
		expected  []string
		shouldErr bool
	}{
		{
			name:      "empty template",
			template:  "",
			shouldErr: true,
		},
		{
			name:     "simple variable",
			template: "Hello {{.Name}}!",
			expected: []string{"Name"},
		},
		{
			name:     "multiple variables",
			template: "User {{.Name}} has {{.Age}} years and {{.Email}} email",
			expected: []string{"Age", "Email", "Name"},
		},
		{
			name:     "nested fields",
			template: "{{.User.Name}} works at {{.Company.Name}}",
			expected: []string{"Company", "User"},
		},
		{
			name:     "range context",
			template: "{{range .Items}}Item: {{.Name}} - {{.Value}}{{end}}",
			expected: []string{"Items"},
		},
		{
			name:     "with builtin functions",
			template: "{{if .Condition}}{{.Items}}{{end}}",
			expected: []string{"Condition", "Items"},
		},
		{
			name:      "syntax error",
			template:  "{{.Name",
			shouldErr: true,
		},
		{
			name:     "complex template with conditions",
			template: `{{if .UseAgents}}Agent: {{.AgentName}}{{else}}Tool: {{.ToolName}}{{end}}`,
			expected: []string{"AgentName", "ToolName", "UseAgents"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validator.ExtractTemplateVariables(tc.template)

			if tc.shouldErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d variables, got %d: %v", len(tc.expected), len(result), result)
				return
			}

			for i, expected := range tc.expected {
				if result[i] != expected {
					t.Errorf("Variable %d: expected %s, got %s", i, expected, result[i])
				}
			}
		})
	}
}

// TestValidatePrompt tests the main validation function
func TestValidatePrompt(t *testing.T) {
	testCases := []struct {
		name        string
		promptType  templates.PromptType
		template    string
		expectedErr string
		errorType   validator.ErrorType
	}{
		{
			name:        "valid template",
			promptType:  templates.PromptTypePrimaryAgent,
			template:    "You are an AI assistant. Your name is {{.FinalyToolName}} and you can use {{.SearchToolName}} for searches.",
			expectedErr: "",
		},
		{
			name:        "empty template",
			promptType:  templates.PromptTypePrimaryAgent,
			template:    "",
			expectedErr: "Empty Template",
			errorType:   validator.ErrorTypeEmptyTemplate,
		},
		{
			name:        "unauthorized variable",
			promptType:  templates.PromptTypePrimaryAgent,
			template:    "Hello {{.UnauthorizedVar}}! You can use {{.FinalyToolName}}.",
			expectedErr: "Unauthorized Variable",
			errorType:   validator.ErrorTypeUnauthorizedVar,
		},
		{
			name:        "syntax error",
			promptType:  templates.PromptTypePrimaryAgent,
			template:    "{{.FinalyToolName",
			expectedErr: "Syntax Error",
			errorType:   validator.ErrorTypeSyntax,
		},
		{
			name:        "unknown prompt type",
			promptType:  "unknown_prompt_type",
			template:    "{{.SomeVar}}",
			expectedErr: "Unauthorized Variable",
			errorType:   validator.ErrorTypeUnauthorizedVar,
		},
		{
			name:        "multiple unauthorized variables",
			promptType:  templates.PromptTypePrimaryAgent,
			template:    "{{.UnauthorizedVar1}} and {{.UnauthorizedVar2}} with valid {{.FinalyToolName}}",
			expectedErr: "Unauthorized Variable",
			errorType:   validator.ErrorTypeUnauthorizedVar,
		},
		{
			name:       "valid complex template",
			promptType: templates.PromptTypeAssistant,
			template: `You are an assistant with the following tools:
{{if .UseAgents}}
- {{.SearchToolName}}
- {{.PentesterToolName}}
- {{.CoderToolName}}
{{end}}
Current time: {{.CurrentTime}}
Language: {{.Lang}}`,
			expectedErr: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidatePrompt(tc.promptType, tc.template)

			if tc.expectedErr == "" {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				return
			}

			if err == nil {
				t.Errorf("Expected error containing '%s', but got no error", tc.expectedErr)
				return
			}

			if !strings.Contains(err.Error(), tc.expectedErr) {
				t.Errorf("Expected error containing '%s', but got: %v", tc.expectedErr, err)
			}

			// Check error type if it's a ValidationError
			if validationErr, ok := err.(*validator.ValidationError); ok {
				if validationErr.Type != tc.errorType {
					t.Errorf("Expected error type %s, but got %s", tc.errorType, validationErr.Type)
				}
			}
		})
	}
}

// TestValidationErrorTypes tests that validation errors provide helpful information
func TestValidationErrorTypes(t *testing.T) {
	testCases := []struct {
		name         string
		promptType   templates.PromptType
		template     string
		checkDetails func(t *testing.T, err error)
	}{
		{
			name:       "syntax error with details",
			promptType: templates.PromptTypePrimaryAgent,
			template:   "{{.FinalyToolName",
			checkDetails: func(t *testing.T, err error) {
				validationErr, ok := err.(*validator.ValidationError)
				if !ok {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if validationErr.Details == "" {
					t.Error("Expected syntax error details, but got empty string")
				}
				if !strings.Contains(validationErr.Details, "brace") {
					t.Errorf("Expected syntax details to mention braces, got: %s", validationErr.Details)
				}
			},
		},
		{
			name:       "unauthorized variable with explanation",
			promptType: templates.PromptTypePrimaryAgent,
			template:   "{{.FinalyToolName}} and {{.NonExistentVar}}",
			checkDetails: func(t *testing.T, err error) {
				validationErr, ok := err.(*validator.ValidationError)
				if !ok {
					t.Errorf("Expected ValidationError, got %T", err)
					return
				}
				if !strings.Contains(validationErr.Message, "NonExistentVar") {
					t.Errorf("Expected error message to mention NonExistentVar, got: %s", validationErr.Message)
				}
				if !strings.Contains(validationErr.Details, "Backend code") {
					t.Errorf("Expected details to explain backend limitation, got: %s", validationErr.Details)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidatePrompt(tc.promptType, tc.template)
			if err == nil {
				t.Fatal("Expected error, but got none")
			}
			tc.checkDetails(t, err)
		})
	}
}

// TestValidatePromptWithRealTemplates tests validation using actual templates from the system
func TestValidatePromptWithRealTemplates(t *testing.T) {
	defaultPrompts, err := templates.GetDefaultPrompts()
	if err != nil {
		t.Fatalf("Failed to load default prompts: %v", err)
	}

	// Test a few key templates to ensure they validate correctly
	testCases := []struct {
		name        string
		promptType  templates.PromptType
		getTemplate func() string
	}{
		{
			name:        "primary agent template",
			promptType:  templates.PromptTypePrimaryAgent,
			getTemplate: func() string { return defaultPrompts.AgentsPrompts.PrimaryAgent.System.Template },
		},
		{
			name:        "assistant template",
			promptType:  templates.PromptTypeAssistant,
			getTemplate: func() string { return defaultPrompts.AgentsPrompts.Assistant.System.Template },
		},
		{
			name:        "pentester template",
			promptType:  templates.PromptTypePentester,
			getTemplate: func() string { return defaultPrompts.AgentsPrompts.Pentester.System.Template },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			template := tc.getTemplate()
			err := validator.ValidatePrompt(tc.promptType, template)
			if err != nil {
				t.Errorf("Real template failed validation: %v", err)
			}
		})
	}
}

// TestVariableExtractionEdgeCases tests edge cases in variable extraction
func TestVariableExtractionEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		template string
		expected []string
	}{
		{
			name:     "local variables ignored",
			template: "{{range .Items}}{{$item := .}}{{$item.Name}}{{end}}",
			expected: []string{"Items"},
		},
		{
			name:     "builtin functions ignored",
			template: "{{.Items}} {{range .Items}}{{end}} {{if .Condition}}{{end}}",
			expected: []string{"Condition", "Items"},
		},
		{
			name:     "nested range contexts",
			template: "{{range .Categories}}{{range .Items}}{{.Name}}{{end}}{{end}}",
			expected: []string{"Categories", "Items"},
		},
		{
			name:     "complex conditions",
			template: "{{if .A}}{{.C}}{{else}}{{if .D}}{{.E}}{{end}}{{end}}",
			expected: []string{"A", "C", "D", "E"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validator.ExtractTemplateVariables(tc.template)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
				return
			}

			for i, expected := range tc.expected {
				if result[i] != expected {
					t.Errorf("Variable %d: expected %s, got %s", i, expected, result[i])
				}
			}
		})
	}
}
