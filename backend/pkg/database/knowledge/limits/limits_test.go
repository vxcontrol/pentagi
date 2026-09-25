package limits

import (
	"strings"
	"testing"
)

func TestLimits_ValidateQuestion_RequiresTextWithinTheLimit(t *testing.T) {
	tests := []struct {
		name     string
		question string
		wantErr  string
	}{
		{name: "ordinary question", question: "how do I check the kernel version?"},
		{name: "empty", question: "", wantErr: "question is required"},
		{name: "spaces only", question: "   ", wantErr: "question is required"},
		{name: "tab and newline only", question: "\t\n", wantErr: "question is required"},
		{name: "one over the length limit", question: strings.Repeat("я", 2049), wantErr: "question must not exceed 2048 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQuestion(tt.question)

			if tt.wantErr == "" && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.wantErr != "" && (err == nil || err.Error() != tt.wantErr) {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

// Rows are keyed by validator; each "я" is two bytes, so a byte count would refuse the rows at the limit.
func TestLimits_ValidateLen_CountsRunesAgainstEachFieldsLimit(t *testing.T) {
	tests := []struct {
		name    string
		check   func(string) error
		runes   int
		wantErr string
	}{
		{name: "a blank question passes the length check", check: ValidateQuestionLen, runes: 0},
		{name: "question at the limit", check: ValidateQuestionLen, runes: 2048},
		{name: "question one over", check: ValidateQuestionLen, runes: 2049, wantErr: "question must not exceed 2048 characters"},
		{name: "content at the limit", check: ValidateContentLen, runes: 65536},
		{name: "content one over", check: ValidateContentLen, runes: 65537, wantErr: "content must not exceed 65536 characters"},
		{name: "description at the limit", check: ValidateDescriptionLen, runes: 1000},
		{name: "description one over", check: ValidateDescriptionLen, runes: 1001, wantErr: "description must not exceed 1000 characters"},
		{name: "code language at the limit", check: ValidateCodeLangLen, runes: 100},
		{name: "code language one over", check: ValidateCodeLangLen, runes: 101, wantErr: "code language must not exceed 100 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.check(strings.Repeat("я", tt.runes))

			if tt.wantErr == "" && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.wantErr != "" && (err == nil || err.Error() != tt.wantErr) {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestLimits_FillBlankQuestion_FallsBackToTheFirstContentLineThenTheDefault(t *testing.T) {
	tests := []struct {
		name     string
		question string
		fallback string
		want     string
	}{
		{name: "keeps a question that carries text", question: "  real question  ", want: "real question"},
		{name: "blank falls back to the first content line", question: "  ", fallback: "First line\nsecond", want: "First line"},
		{name: "strips a markdown heading marker", question: "", fallback: "# A Title\nbody", want: "A Title"},
		{name: "skips blank leading lines", question: "", fallback: "\n\n  \n## Real\nx", want: "Real"},
		{name: "blank content falls back to the default", question: "\t", fallback: "   \n\n", want: "Untitled"},
		{name: "no content at all falls back to the default", question: "", fallback: "", want: "Untitled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FillBlankQuestion(tt.question, tt.fallback, "Untitled"); got != tt.want {
				t.Fatalf("FillBlankQuestion = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLimits_TruncateToLimit_CutsToTheLimitInRunes(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		limit int
		want  string
	}{
		{name: "below the limit is unchanged", in: "абвгд", limit: 10, want: "абвгд"},
		{name: "at the limit is unchanged", in: "абвгдежзий", limit: 10, want: "абвгдежзий"},
		{name: "one over keeps the first runes", in: "абвгдежзийк", limit: 10, want: "абвгдежзий"},
		{name: "cuts by runes not bytes", in: strings.Repeat("я", 100), limit: 10, want: strings.Repeat("я", 10)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TruncateToLimit(tt.in, tt.limit); got != tt.want {
				t.Fatalf("TruncateToLimit(%q, %d) = %q, want %q", tt.in, tt.limit, got, tt.want)
			}
		})
	}
}

func TestLimits_CanonicalCodeLang_KeepsNamesDropsQuotesAndBoundsLength(t *testing.T) {
	tests := []struct {
		name string
		lang string
		want string
	}{
		{name: "plain name survives", lang: "python", want: "python"},
		{name: "punctuation of real names survives", lang: "c++ / f# node.js objective-c", want: "c++ / f# node.js objective-c"},
		{name: "surrounding blanks are trimmed", lang: "  golang\n", want: "golang"},
		{name: "non-latin name survives", lang: "питон", want: "питон"},
		{name: "single quote is dropped", lang: "python' OR '1'='1", want: "python OR 11"},
		{name: "double quote and backslash are dropped", lang: `py"th\on`, want: "python"},
		{name: "statement punctuation is dropped", lang: "python); DROP TABLE docs;--", want: "python DROP TABLE docs--"},
		{name: "a name at the length limit is kept whole", lang: strings.Repeat("я", 100), want: strings.Repeat("я", 100)},
		{name: "a name past the length limit is cut to it", lang: strings.Repeat("я", 101), want: strings.Repeat("я", 100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanonicalCodeLang(tt.lang); got != tt.want {
				t.Fatalf("CanonicalCodeLang(%q) = %q, want %q", tt.lang, got, tt.want)
			}
		})
	}
}
