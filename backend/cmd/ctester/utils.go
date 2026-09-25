package main

import (
	"strings"
)

// Helper functions

// TruncateString truncates a string to a specified maximum length and adds ellipsis
func TruncateString(s string, maxLength int) string {
	s = strings.Trim(s, "\n\r\t ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	runes := []rune(s)
	if len(runes) <= maxLength {
		return s
	}

	return string(runes[:maxLength-3]) + "..."
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
		result = strings.ReplaceAll(result, r.from, r.to)
	}

	return result
}
