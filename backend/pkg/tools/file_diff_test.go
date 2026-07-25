package tools

import (
	"strings"
	"testing"
)

// TestApplyUnifiedDiff covers ApplyUnifiedDiff end to end (parse + patch +
// apply) with both positive (diff applies, exact content produced) and
// negative (diff rejected, or doesn't match, with content left untouched)
// cases, using table-driven test cases as sub-tests.
func TestApplyUnifiedDiff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		diff    string
		want    string // expected new content, when wantErr is false
		wantErr string // substring expected in the error, when non-empty
	}{
		{
			name:    "replace a single line",
			content: "line1\nline2\nline3\n",
			diff: "@@ -1,2 +1,2 @@\n" +
				" line1\n" +
				"-line2\n" +
				"+line2 changed\n",
			want: "line1\nline2 changed\nline3\n",
		},
		{
			name:    "insert lines with pure addition",
			content: "line1\nline2\nline3\n",
			diff: "@@ -1,2 +1,3 @@\n" +
				" line1\n" +
				"+inserted\n" +
				" line2\n",
			want: "line1\ninserted\nline2\nline3\n",
		},
		{
			name:    "delete a line",
			content: "line1\nline2\nline3\n",
			diff: "@@ -1,3 +1,2 @@\n" +
				" line1\n" +
				"-line2\n" +
				" line3\n",
			want: "line1\nline3\n",
		},
		{
			name:    "multiple non-overlapping hunks applied together",
			content: "a\nb\nc\nd\ne\nf\ng\n",
			diff: "@@ -1,2 +1,2 @@\n" +
				" a\n" +
				"-b\n" +
				"+B\n" +
				"@@ -6,2 +6,2 @@\n" +
				" f\n" +
				"-g\n" +
				"+G\n",
			want: "a\nB\nc\nd\ne\nf\nG\n",
		},
		{
			name:    "tolerates minor line-number drift when content matches",
			content: "line1\nline2\nline3\nline4\n",
			// Header claims the hunk starts at line 2; its own context lines
			// show it's really at line 1 (as if a line had been
			// inserted/removed elsewhere since the diff was written) -
			// PatchApply's fuzzy search must find the real position from
			// the content, not the (slightly wrong) header line number.
			diff: "@@ -2,3 +2,3 @@\n" +
				" line1\n" +
				"-line2\n" +
				"+line2 changed\n" +
				" line3\n",
			want: "line1\nline2 changed\nline3\nline4\n",
		},
		{
			name:    "tolerates '--- a/file' / '+++ b/file' headers",
			content: "line1\nline2\n",
			diff: "--- a/file.txt\n" +
				"+++ b/file.txt\n" +
				"@@ -2,1 +2,1 @@\n" +
				"-line2\n" +
				"+line2 changed\n",
			want: "line1\nline2 changed\n",
		},
		{
			name:    "tolerates a fully blank line as an empty context line",
			content: "line1\n\nline3\n",
			diff: "@@ -1,3 +1,3 @@\n" +
				" line1\n" +
				"\n" +
				"-line3\n" +
				"+line3 changed\n",
			want: "line1\n\nline3 changed\n",
		},
		{
			name:    "preserves special characters requiring URL-escaping",
			content: "a = 1 + 2 % 3 & done\n",
			diff: "@@ -1,1 +1,1 @@\n" +
				"-a = 1 + 2 % 3 & done\n" +
				"+a = 4 + 5 % 6 & done\n",
			want: "a = 4 + 5 % 6 & done\n",
		},
		{
			name:    "multi-hunk: middle hunk has no context on either side",
			content: "a\nb\nc\nd\ne\nf\ng\n",
			// Only the first and last hunk need synthesized context (see
			// ensureContextBoundaries); the middle hunk works fine without it.
			diff: "@@ -1,2 +1,2 @@\n" +
				" a\n" +
				"-b\n" +
				"+B\n" +
				"@@ -4,1 +4,1 @@\n" +
				"-d\n" +
				"+D\n" +
				"@@ -6,2 +6,2 @@\n" +
				"-f\n" +
				"+F\n" +
				" g\n",
			want: "a\nB\nc\nD\ne\nF\ng\n",
		},
		{
			name:    "hunk header without explicit line counts (implicit 1)",
			content: "line1\nline2\nline3\n",
			diff: "@@ -2 +2 @@\n" +
				"-line2\n" +
				"+line2 changed\n",
			want: "line1\nline2 changed\nline3\n",
		},
		{
			name:    "negative: empty diff is rejected",
			content: "line1\n",
			diff:    "",
			wantErr: "diff is empty",
		},
		{
			name:    "negative: whitespace-only diff is rejected",
			content: "line1\n",
			diff:    "   \n\n  ",
			wantErr: "no hunks",
		},
		{
			name:    "negative: no hunk header at all",
			content: "line1\n",
			diff:    "just some text\nwith no diff markers\n",
			wantErr: "expected a hunk header",
		},
		{
			// Observed from a real model (gpt-oss:120b via Ollama): it
			// sometimes emits a bare "@@" with no position info at all when
			// unsure of exact line numbers. diffmatchpatch's content-based
			// fuzzy search must still locate the hunk from context alone -
			// which requires the target line to be textually distinct from
			// the file's other lines (unlike near-duplicates such as
			// "line1"/"line2"/"line3", fuzzy search without any position
			// hint can genuinely prefer a closer-but-wrong near-duplicate).
			name:    "tolerates a hunk header with no position info at all (bare '@@')",
			content: "Status: draft\nOwner: alice\nPriority: low\n",
			diff:    "@@\n-Priority: low\n+Priority: high\n",
			want:    "Status: draft\nOwner: alice\nPriority: high\n",
		},
		{
			name:    "tolerates unparseable text after a bare '@@' header",
			content: "line1\n",
			diff:    "@@ not a real header @@\n-line1\n+line2\n",
			want:    "line2\n",
		},
		{
			name:    "negative: hunk body line with invalid prefix",
			content: "line1\nline2\n",
			diff:    "@@ -1,2 +1,2 @@\n line1\n*line2\n",
			wantErr: "invalid diff line",
		},
		{
			name:    "negative: hunk header with no content lines",
			content: "line1\nline2\n",
			diff:    "@@ -1,1 +1,1 @@\n@@ -2,1 +2,1 @@\n-line2\n+line2 changed\n",
			wantErr: "no content lines",
		},
		{
			name:    "negative: context does not match file content",
			content: "line1\nline2\nline3\n",
			diff: "@@ -2,1 +2,1 @@\n" +
				"-this line does not exist in the file\n" +
				"+replacement\n",
			wantErr: "could not be applied",
		},
		{
			name:    "negative: content matches nowhere near the claimed location",
			content: strings.Repeat("filler line\n", 50) + "target line\n" + strings.Repeat("filler line\n", 50),
			diff: "@@ -1,1 +1,1 @@\n" +
				"-completely different text that is nowhere in the file\n" +
				"+replacement\n",
			wantErr: "could not be applied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, hunks, err := ApplyUnifiedDiff(tt.content, tt.diff)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("ApplyUnifiedDiff() expected error containing %q, got nil (result: %q)", tt.wantErr, got)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("ApplyUnifiedDiff() error = %q, want it to contain %q", err.Error(), tt.wantErr)
				}
				if got != "" {
					t.Errorf("ApplyUnifiedDiff() on error must return empty content, got %q", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("ApplyUnifiedDiff() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ApplyUnifiedDiff() content = %q, want %q", got, tt.want)
			}
			if hunks <= 0 {
				t.Errorf("ApplyUnifiedDiff() hunks applied = %d, want > 0", hunks)
			}
		})
	}
}

// TestParseUnifiedDiff checks the parser in isolation, independent of
// go-diff, so a bug in hunk parsing and a bug in patch application are
// distinguishable from their respective test failures.
func TestParseUnifiedDiff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		diff      string
		wantHunks int
		wantLines int // lines in the first hunk, when wantHunks > 0
		wantErr   string
	}{
		{
			name:      "single hunk",
			diff:      "@@ -1,2 +1,2 @@\n line1\n-line2\n+line2 changed\n",
			wantHunks: 1,
			wantLines: 3,
		},
		{
			name:      "two hunks",
			diff:      "@@ -1,1 +1,1 @@\n-a\n+A\n@@ -3,1 +3,1 @@\n-c\n+C\n",
			wantHunks: 2,
		},
		{
			name:    "empty",
			diff:    "",
			wantErr: "diff is empty",
		},
		{
			name:    "garbage before first hunk",
			diff:    "not a diff at all",
			wantErr: "expected a hunk header",
		},
		{
			name:    "garbage between header lines and hunk",
			diff:    "--- a/file\nsome garbage line\n@@ -1,1 +1,1 @@\n-a\n+A\n",
			wantErr: "expected a hunk header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hunks, err := parseUnifiedDiff(tt.diff)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("parseUnifiedDiff() expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("parseUnifiedDiff() error = %q, want it to contain %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseUnifiedDiff() unexpected error: %v", err)
			}
			if len(hunks) != tt.wantHunks {
				t.Fatalf("parseUnifiedDiff() hunks = %d, want %d", len(hunks), tt.wantHunks)
			}
			if tt.wantLines > 0 && len(hunks[0].lines) != tt.wantLines {
				t.Errorf("parseUnifiedDiff() first hunk lines = %d, want %d", len(hunks[0].lines), tt.wantLines)
			}
		})
	}
}
