package templates_test

import (
	"fmt"
	"reflect"
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
	// According to the code, structure ToolsPrompts should have 7 prompts
	if tools > 7 {
		t.Fatalf("tools prompts amount is %d, expected 7", tools)
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
