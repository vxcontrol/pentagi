package validator

import (
	"fmt"
	"sort"
	"strings"
	"text/template"
	"text/template/parse"

	"pentagi/pkg/templates"
)

// ValidationError represents different types of validation errors
type ValidationError struct {
	Type    ErrorType
	Message string
	Line    int // line number if available
	Details string
}

func (e *ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s at line %d: %s", e.Type, e.Line, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

type ErrorType string

const (
	ErrorTypeSyntax               ErrorType = "Syntax Error"
	ErrorTypeUnauthorizedVar      ErrorType = "Unauthorized Variable"
	ErrorTypeRenderingFailed      ErrorType = "Rendering Failed"
	ErrorTypeEmptyTemplate        ErrorType = "Empty Template"
	ErrorTypeVariableTypeMismatch ErrorType = "Variable Type Mismatch"
)

// ValidatePrompt validates a user-provided prompt template against the declared variables
func ValidatePrompt(promptType templates.PromptType, prompt string) error {
	if strings.TrimSpace(prompt) == "" {
		return &ValidationError{
			Type:    ErrorTypeEmptyTemplate,
			Message: "template content cannot be empty",
		}
	}

	// Extract variables from the template
	actualVars, err := ExtractTemplateVariables(prompt)
	if err != nil {
		return &ValidationError{
			Type:    ErrorTypeSyntax,
			Message: fmt.Sprintf("failed to parse template: %v", err),
			Details: extractSyntaxDetails(err),
		}
	}

	// Get declared variables for this prompt type
	declaredVars, exists := templates.PromptVariables[promptType]
	if !exists {
		return &ValidationError{
			Type:    ErrorTypeUnauthorizedVar,
			Message: fmt.Sprintf("unknown prompt type: %s", promptType),
		}
	}

	// Check for unauthorized variables (variables not in PromptVariables)
	declaredSet := make(map[string]bool)
	for _, v := range declaredVars {
		declaredSet[v] = true
	}

	var unauthorizedVars []string
	for _, v := range actualVars {
		if !declaredSet[v] {
			unauthorizedVars = append(unauthorizedVars, v)
		}
	}

	if len(unauthorizedVars) > 0 {
		sort.Strings(unauthorizedVars)
		return &ValidationError{
			Type:    ErrorTypeUnauthorizedVar,
			Message: fmt.Sprintf("template uses unauthorized variables: %v", unauthorizedVars),
			Details: "These variables are not declared in PromptVariables for this prompt type. Backend code cannot provide these variables.",
		}
	}

	// Test template rendering with mock data
	mockData := CreateDummyTemplateData()
	if err := testTemplateRendering(prompt, mockData); err != nil {
		return &ValidationError{
			Type:    ErrorTypeRenderingFailed,
			Message: fmt.Sprintf("template rendering failed: %v", err),
			Details: extractRenderingDetails(err),
		}
	}

	return nil
}

// ExtractTemplateVariables parses a template and extracts all top-level variables
func ExtractTemplateVariables(templateContent string) ([]string, error) {
	if strings.TrimSpace(templateContent) == "" {
		return nil, fmt.Errorf("template content is empty")
	}

	// Create function map with all builtin functions as nil values for the parser
	funcMap := template.FuncMap{
		// Builtin comparison and logic functions
		"and": nil, "or": nil, "not": nil,
		"eq": nil, "ne": nil, "lt": nil, "le": nil, "gt": nil, "ge": nil,
		// Builtin utility functions
		"len": nil, "index": nil, "slice": nil, "print": nil, "printf": nil, "println": nil,
		"html": nil, "js": nil, "urlquery": nil, "call": nil,
		// Additional common functions that might be used
		"add": nil, "sub": nil, "mul": nil, "div": nil, "mod": nil,
		"upper": nil, "lower": nil, "title": nil, "trim": nil, "trimSpace": nil,
		"default": nil, "empty": nil, "contains": nil, "hasPrefix": nil, "hasSuffix": nil,
	}

	// Parse template with function map to get AST
	parsed, err := parse.Parse("validation", templateContent, "{{", "}}", funcMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	variables := make(map[string]bool)

	// Analyze each tree in the template
	for _, tree := range parsed {
		if tree != nil && tree.Root != nil {
			extractVariablesFromNode(tree.Root, variables, false)
		}
	}

	// Convert to sorted slice for consistent comparison
	var result []string
	for varName := range variables {
		result = append(result, varName)
	}
	sort.Strings(result)

	return result, nil
}

// extractVariablesFromNode recursively extracts variables from AST nodes
func extractVariablesFromNode(node parse.Node, variables map[string]bool, inRangeContext bool) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *parse.ListNode:
		if n != nil {
			for _, child := range n.Nodes {
				extractVariablesFromNode(child, variables, inRangeContext)
			}
		}

	case *parse.ActionNode:
		extractVariablesFromPipe(n.Pipe, variables, inRangeContext)

	case *parse.IfNode:
		extractVariablesFromPipe(n.Pipe, variables, inRangeContext)
		extractVariablesFromNode(n.List, variables, inRangeContext)
		extractVariablesFromNode(n.ElseList, variables, inRangeContext)

	case *parse.RangeNode:
		// Extract the range variable itself
		extractVariablesFromPipe(n.Pipe, variables, false)
		// Process contents in range context (skip field extractions)
		extractVariablesFromNode(n.List, variables, true)
		extractVariablesFromNode(n.ElseList, variables, inRangeContext)

	case *parse.WithNode:
		extractVariablesFromPipe(n.Pipe, variables, inRangeContext)
		extractVariablesFromNode(n.List, variables, inRangeContext)
		extractVariablesFromNode(n.ElseList, variables, inRangeContext)

	case *parse.TemplateNode:
		extractVariablesFromPipe(n.Pipe, variables, inRangeContext)
	}
}

// extractVariablesFromPipe extracts variables from pipe expressions
func extractVariablesFromPipe(pipe *parse.PipeNode, variables map[string]bool, inRangeContext bool) {
	if pipe == nil {
		return
	}

	for _, cmd := range pipe.Cmds {
		extractVariablesFromCommand(cmd, variables, inRangeContext)
	}
}

// extractVariablesFromCommand extracts variables from command nodes
func extractVariablesFromCommand(cmd *parse.CommandNode, variables map[string]bool, inRangeContext bool) {
	if cmd == nil {
		return
	}

	for _, arg := range cmd.Args {
		extractVariablesFromArg(arg, variables, inRangeContext)
	}
}

// extractVariablesFromArg extracts variables from argument nodes
func extractVariablesFromArg(arg parse.Node, variables map[string]bool, inRangeContext bool) {
	switch n := arg.(type) {
	case *parse.FieldNode:
		// Extract top-level variable from field access like .User.Name -> User
		if len(n.Ident) > 0 {
			topLevel := n.Ident[0]
			if topLevel != "." && !isBuiltinFunction(topLevel) {
				// In range context, skip direct field access as they refer to current item
				if !inRangeContext {
					variables[topLevel] = true
				}
			}
		}

	case *parse.VariableNode:
		// Handle variable references, skip local variables starting with $
		if len(n.Ident) > 0 {
			topLevel := n.Ident[0]
			if !strings.HasPrefix(topLevel, "$") && !isBuiltinFunction(topLevel) {
				variables[topLevel] = true
			}
		}

	case *parse.PipeNode:
		extractVariablesFromPipe(n, variables, inRangeContext)
	}
}

// isBuiltinFunction checks if a name is a Go template builtin function
func isBuiltinFunction(name string) bool {
	builtins := map[string]bool{
		// Template actions and comparison
		"and": true, "call": true, "html": true, "index": true, "slice": true,
		"js": true, "len": true, "not": true, "or": true, "print": true,
		"printf": true, "println": true, "urlquery": true, "eq": true,
		"ne": true, "lt": true, "le": true, "gt": true, "ge": true,
		"with": true, "if": true, "range": true, "template": true, "block": true,
		// Math functions
		"add": true, "sub": true, "mul": true, "div": true, "mod": true,
		// String functions
		"upper": true, "lower": true, "title": true, "trim": true, "trimSpace": true,
		// Additional common functions
		"default": true, "empty": true, "contains": true, "hasPrefix": true, "hasSuffix": true,
	}
	return builtins[name]
}

// testTemplateRendering tests if template can be rendered with mock data
func testTemplateRendering(templateContent string, data map[string]any) error {
	_, err := templates.RenderPrompt("validation", templateContent, data)
	return err
}

// extractSyntaxDetails extracts more detailed information from parsing errors
func extractSyntaxDetails(err error) string {
	errStr := err.Error()
	if strings.Contains(errStr, "unexpected") || strings.Contains(errStr, "expected") {
		return "Check for missing closing braces '}}' or incorrect template syntax"
	}
	if strings.Contains(errStr, "function") && strings.Contains(errStr, "not defined") {
		return "Unknown function or incorrect function call syntax"
	}
	if strings.Contains(errStr, "EOF") || strings.Contains(errStr, "unclosed") {
		return "Template appears to be incomplete - missing closing braces"
	}
	return "Review template syntax according to Go template documentation"
}

// extractRenderingDetails extracts more detailed information from rendering errors
func extractRenderingDetails(err error) string {
	errStr := err.Error()
	if strings.Contains(errStr, "nil pointer") || strings.Contains(errStr, "can't evaluate") {
		return "Variable type mismatch - check if template expects different data structure"
	}
	if strings.Contains(errStr, "undefined") {
		return "Referenced variable or field not found in provided data"
	}
	if strings.Contains(errStr, "index") {
		return "Array/slice index out of bounds or incorrect index type"
	}
	return "Check variable types and data structure in template"
}
