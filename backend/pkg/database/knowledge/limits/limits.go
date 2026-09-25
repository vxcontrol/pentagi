package limits

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// The metadata column is plain JSON with no constraint of its own, so every door
// that writes a knowledge document applies these itself.
const (
	MaxCodeLangLen    = 100
	MaxContentLen     = 65536
	MaxDescriptionLen = 1000
	MaxQuestionLen    = 2048
	MaxSearchLimit    = 100
)

func validateLen(field, value string, limit int) error {
	if utf8.RuneCountInString(value) > limit {
		return fmt.Errorf("%s must not exceed %d characters", field, limit)
	}
	return nil
}

func ValidateQuestionLen(question string) error {
	return validateLen("question", question, MaxQuestionLen)
}

func ValidateQuestion(question string) error {
	if strings.TrimSpace(question) == "" {
		return fmt.Errorf("question is required")
	}
	return ValidateQuestionLen(question)
}

func ValidateContentLen(content string) error {
	return validateLen("content", content, MaxContentLen)
}

func ValidateDescriptionLen(description string) error {
	return validateLen("description", description, MaxDescriptionLen)
}

func ValidateCodeLangLen(codeLang string) error {
	return validateLen("code language", codeLang, MaxCodeLangLen)
}

func TruncateToLimit(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return string(runes[:limit])
}

func CanonicalCodeLang(codeLang string) string {
	canonical := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("+-#._/ ", r) {
			return r
		}
		return -1
	}, codeLang)

	return TruncateToLimit(strings.TrimSpace(canonical), MaxCodeLangLen)
}

// FillBlankQuestion supplies a title when question is blank. It does not bound
// length: the caller truncates the stored value, since anonymization can push a
// derived title past the limit.
func FillBlankQuestion(question, fallback, defaultName string) string {
	if q := strings.TrimSpace(question); q != "" {
		return q
	}
	for line := range strings.SplitSeq(fallback, "\n") {
		if title := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "#>*-` ")); title != "" {
			return title
		}
	}
	return defaultName
}
