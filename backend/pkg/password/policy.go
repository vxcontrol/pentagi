// Package password holds the password policy the product enforces.
package password

import (
	"regexp"
	"unicode/utf8"
)

// MaxBytes is the longest password bcrypt.GenerateFromPassword accepts.
const MaxBytes = 72

const (
	minComposedRunes = 8
	minPlainRunes    = 16
)

var (
	numberRegex = regexp.MustCompile("[0-9]")
	alphaLRegex = regexp.MustCompile("[a-z]")
	alphaURegex = regexp.MustCompile("[A-Z]")
	specRegex   = regexp.MustCompile("[!@#$&*]")
)

// IsStrong reports whether the password satisfies the composition rule.
func IsStrong(password string) bool {
	length := utf8.RuneCountInString(password)

	return length >= minPlainRunes || (length >= minComposedRunes &&
		numberRegex.MatchString(password) &&
		alphaLRegex.MatchString(password) &&
		alphaURegex.MatchString(password) &&
		specRegex.MatchString(password))
}

// FitsHashLimit reports whether the password is within MaxBytes.
func FitsHashLimit(password string) bool {
	return len(password) <= MaxBytes
}
