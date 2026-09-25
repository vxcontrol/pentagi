package password

import (
	"strings"
	"testing"
)

var policyCases = []struct {
	Name     string
	Password string
	Strong   bool
	Fits     bool
}{
	{Name: "an empty password", Password: "", Strong: false, Fits: true},
	{Name: "five digits", Password: "12345", Strong: false, Fits: true},
	{Name: "four cyrillic runes in eight bytes", Password: "абвг", Strong: false, Fits: true},
	{Name: "seven of every class", Password: "Aa1!aa1", Strong: false, Fits: true},
	{Name: "eight of every class", Password: "Aa1!aa1!", Strong: true, Fits: true},
	{Name: "eight without a special", Password: "Aa1aaa1a", Strong: false, Fits: true},
	{Name: "eight without a digit", Password: "Aa!aaa!a", Strong: false, Fits: true},
	{Name: "eight without an uppercase", Password: "aa1!aa1!", Strong: false, Fits: true},
	{Name: "eight without a lowercase", Password: "AA1!AA1!", Strong: false, Fits: true},
	{Name: "fifteen plain runes", Password: "abcdefghijklmno", Strong: false, Fits: true},
	{Name: "sixteen plain runes", Password: "abcdefghijklmnop", Strong: true, Fits: true},
	{Name: "sixteen cyrillic runes", Password: strings.Repeat("я", 16), Strong: true, Fits: true},
	{Name: "seventy-two bytes", Password: strings.Repeat("a", 72), Strong: true, Fits: true},
	{Name: "seventy-three bytes", Password: strings.Repeat("a", 73), Strong: true, Fits: false},
	{Name: "thirty-six cyrillic runes in seventy-two bytes", Password: strings.Repeat("я", 36), Strong: true, Fits: true},
	{Name: "thirty-seven cyrillic runes over the byte limit", Password: strings.Repeat("я", 37), Strong: true, Fits: false},
}

func TestPolicy_JudgesStrengthAndHashLimit(t *testing.T) {
	t.Parallel()

	for _, tt := range policyCases {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			if got := IsStrong(tt.Password); got != tt.Strong {
				t.Errorf("IsStrong(%q) = %v, want %v", tt.Password, got, tt.Strong)
			}
			if got := FitsHashLimit(tt.Password); got != tt.Fits {
				t.Errorf("FitsHashLimit(%q) = %v, want %v", tt.Password, got, tt.Fits)
			}
		})
	}
}
