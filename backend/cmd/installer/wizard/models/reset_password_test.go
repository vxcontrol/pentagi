package models

import (
	"strings"
	"testing"
)

func TestResetPassword_ValidatePasswords_RefusesWhatTheAPIRefuses(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name, password, confirm string
		refusal                 string // empty when accepted
	}{
		{name: "an empty password", refusal: "cannot be empty"},
		{name: "seven runes of every class", password: "Aa1!aa1", confirm: "Aa1!aa1", refusal: "must be 16+ characters"},
		{name: "eight runes of every class", password: "Aa1!aa1!", confirm: "Aa1!aa1!"},
		{name: "seventy-two plain bytes", password: strings.Repeat("a", 72), confirm: strings.Repeat("a", 72)},
		{name: "seventy-three plain bytes", password: strings.Repeat("a", 73), confirm: strings.Repeat("a", 73), refusal: "must not exceed 72 bytes"},
		{name: "a confirmation that differs", password: "Aa1!aa1!", confirm: "Aa1!aa1?", refusal: "do not match"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := (&ResetPasswordModel{}).validatePasswords(tt.password, tt.confirm)

			switch {
			case tt.refusal == "" && err != nil:
				t.Errorf("validatePasswords(%q) refused a password the API accepts: %v", tt.password, err)
			case tt.refusal != "" && (err == nil || !strings.Contains(err.Error(), tt.refusal)):
				t.Errorf("validatePasswords(%q, %q) = %v, want a refusal saying %q", tt.password, tt.confirm, err, tt.refusal)
			}
		})
	}
}
