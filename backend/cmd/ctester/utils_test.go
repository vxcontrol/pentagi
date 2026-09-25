package main

import (
	"testing"
	"unicode/utf8"
)

func TestUtils_TruncateString_CutsWholeCharacters(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		want      string
	}{
		{name: "short enough is untouched", input: "ok", maxLength: 10, want: "ok"},
		{name: "ascii is cut with an ellipsis", input: "abcdefghij", maxLength: 8, want: "abcde..."},
		{name: "a multibyte answer keeps whole characters", input: "провайдер ответил", maxLength: 8, want: "прова..."},
		{name: "the boundary itself is not cut", input: "провайд", maxLength: 7, want: "провайд"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := TruncateString(tc.input, tc.maxLength)

			if got != tc.want {
				t.Errorf("TruncateString(%q, %d) = %q, want %q", tc.input, tc.maxLength, got, tc.want)
			}
			if !utf8.ValidString(got) {
				t.Errorf("the report line carries a broken character: %q", got)
			}
		})
	}
}
