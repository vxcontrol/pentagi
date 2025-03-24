package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Helper functions

// TruncateString truncates a string to a specified maximum length and adds ellipsis
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength-3] + "..."
}

// ContainsString checks if a string contains or equals another string
func ContainsString(s, substr string) bool {
	return len(s) > 0 && (s == substr ||
		Contains(s, substr) ||
		Contains(substr, s))
}

// Contains performs a more specific contains check
func Contains(s, substr string) bool {
	return s != "" && substr != "" && len(s) >= len(substr) &&
		(s == substr || (len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}

// SchemaForType creates a JSON schema from a Go struct
func SchemaForType(obj interface{}) map[string]interface{} {
	// Simplified implementation for building a JSON schema from a Go structure
	data, _ := json.Marshal(obj)
	var schema map[string]interface{}
	json.Unmarshal(data, &schema)

	// Add object type
	result := map[string]interface{}{
		"type":       "object",
		"properties": schema,
		"required":   []string{},
	}

	// Define required fields (consider all fields required)
	required := make([]string, 0)
	for key := range schema {
		required = append(required, key)
	}
	result["required"] = required

	return result
}

// StringInSlice checks if a string is in a slice
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// FilterAgentTypes filters the agent types based on a list of requested types
func FilterAgentTypes(allTypes []string, requested []string) []string {
	if len(requested) == 0 || (len(requested) == 1 && requested[0] == "all") {
		return allTypes
	}

	result := make([]string, 0)
	for _, agent := range requested {
		if StringInSlice(agent, allTypes) {
			result = append(result, agent)
		}
	}

	return result
}

// FormatTestName creates a standardized test name
func FormatTestName(testType, prompt string, length int) string {
	return fmt.Sprintf("%s: %s", testType, TruncateString(prompt, length))
}

// EscapeMarkdown escapes special characters in markdown
func EscapeMarkdown(text string) string {
	if text == "" {
		return ""
	}

	replacements := []struct {
		from string
		to   string
	}{
		{"|", "\\|"},
		{"*", "\\*"},
		{"_", "\\_"},
		{"`", "\\`"},
		{"#", "\\#"},
		{"-", "\\-"},
		{".", "\\."},
		{"!", "\\!"},
		{"(", "\\("},
		{")", "\\)"},
		{"[", "\\["},
		{"]", "\\]"},
		{"{", "\\{"},
		{"}", "\\}"},
	}

	result := text
	for _, r := range replacements {
		result = strings.Replace(result, r.from, r.to, -1)
	}

	return result
}
