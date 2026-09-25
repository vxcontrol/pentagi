package tools

import (
	"strings"
	"testing"
)

// A hunk lands where its lines match the file; the header's line number only breaks a tie.
func TestFileDiff_ApplyUnifiedDiff_PatchesTheLinesItsHunksLocate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		content   string
		diff      string
		want      string
		wantHunks int
	}{
		{
			name:    "a replaced line",
			content: "line1\nline2\nline3\n",
			diff: "@@ -1,2 +1,2 @@\n" +
				" line1\n" +
				"-line2\n" +
				"+line2 changed\n",
			want:      "line1\nline2 changed\nline3\n",
			wantHunks: 1,
		},
		{
			name:    "an inserted line",
			content: "line1\nline2\nline3\n",
			diff: "@@ -1,2 +1,3 @@\n" +
				" line1\n" +
				"+inserted\n" +
				" line2\n",
			want:      "line1\ninserted\nline2\nline3\n",
			wantHunks: 1,
		},
		{
			name:    "a deleted line",
			content: "line1\nline2\nline3\n",
			diff: "@@ -1,3 +1,2 @@\n" +
				" line1\n" +
				"-line2\n" +
				" line3\n",
			want:      "line1\nline3\n",
			wantHunks: 1,
		},
		{
			name:    "two hunks apart",
			content: "a\nb\nc\nd\ne\nf\ng\n",
			diff: "@@ -1,2 +1,2 @@\n" +
				" a\n" +
				"-b\n" +
				"+B\n" +
				"@@ -6,2 +6,2 @@\n" +
				" f\n" +
				"-g\n" +
				"+G\n",
			want:      "a\nB\nc\nd\ne\nf\nG\n",
			wantHunks: 2,
		},
		{
			name:    "three hunks, the middle one without context",
			content: "a\nb\nc\nd\ne\nf\ng\n",
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
			want:      "a\nB\nc\nD\ne\nF\ng\n",
			wantHunks: 3,
		},
		{
			name:    "hunks given out of file order",
			content: "a\nb\nc\nd\n",
			diff: "@@ -4,1 +4,1 @@\n" +
				"-d\n" +
				"+D\n" +
				"@@ -1,1 +1,1 @@\n" +
				"-a\n" +
				"+A\n",
			want:      "A\nb\nc\nD\n",
			wantHunks: 2,
		},
		{
			name:    "a header number off by one, the context right",
			content: "line1\nline2\nline3\nline4\n",
			diff: "@@ -2,3 +2,3 @@\n" +
				" line1\n" +
				"-line2\n" +
				"+line2 changed\n" +
				" line3\n",
			want:      "line1\nline2 changed\nline3\nline4\n",
			wantHunks: 1,
		},
		{
			name:    "a header number far from the real line",
			content: strings.Repeat("filler line\n", 60) + "target line\n",
			diff: "@@ -1,1 +1,1 @@\n" +
				"-target line\n" +
				"+changed line\n",
			want:      strings.Repeat("filler line\n", 60) + "changed line\n",
			wantHunks: 1,
		},
		{
			name:    "a header number past the end of the file",
			content: "Status: draft\nOwner: alice\nPriority: low\n",
			diff: "@@ -99,1 +99,1 @@\n" +
				"-Priority: low\n" +
				"+Priority: high\n",
			want:      "Status: draft\nOwner: alice\nPriority: high\n",
			wantHunks: 1,
		},
		{
			name:    "a header without line counts",
			content: "line1\nline2\nline3\n",
			diff: "@@ -2 +2 @@\n" +
				"-line2\n" +
				"+line2 changed\n",
			want:      "line1\nline2 changed\nline3\n",
			wantHunks: 1,
		},
		{
			name:      "a bare '@@' header",
			content:   "Status: draft\nOwner: alice\nPriority: low\n",
			diff:      "@@\n-Priority: low\n+Priority: high\n",
			want:      "Status: draft\nOwner: alice\nPriority: high\n",
			wantHunks: 1,
		},
		{
			name:      "text after a bare '@@' header",
			content:   "line1\n",
			diff:      "@@ not a real header @@\n-line1\n+line2\n",
			want:      "line2\n",
			wantHunks: 1,
		},
		{
			name:    "'--- a/file' and '+++ b/file' headers",
			content: "line1\nline2\n",
			diff: "--- a/file.txt\n" +
				"+++ b/file.txt\n" +
				"@@ -2,1 +2,1 @@\n" +
				"-line2\n" +
				"+line2 changed\n",
			want:      "line1\nline2 changed\n",
			wantHunks: 1,
		},
		{
			// The empty context line is all that tells the two "b" lines apart.
			name:      "a fully blank line standing for an empty context line",
			content:   "b\nc\n\nb\n",
			diff:      "@@\n\n-b\n+B\n",
			want:      "b\nc\n\nB\n",
			wantHunks: 1,
		},
		{
			// The shape `diff -u` gives a file whose last line has no newline.
			name:    "a '\\ No newline at end of file' marker",
			content: "a\nb",
			diff: "@@ -1,2 +1,2 @@\n" +
				" a\n" +
				"-b\n" +
				"\\ No newline at end of file\n" +
				"+B\n" +
				"\\ No newline at end of file\n",
			want:      "a\nB",
			wantHunks: 1,
		},
		{
			name:    "'%', '&' and '+' inside a line",
			content: "a = 1 + 2 % 3 & done\n",
			diff: "@@ -1,1 +1,1 @@\n" +
				"-a = 1 + 2 % 3 & done\n" +
				"+a = 4 + 5 % 6 & done\n",
			want:      "a = 4 + 5 % 6 & done\n",
			wantHunks: 1,
		},
		{
			name:    "context lines are taken from the file, not the diff",
			content: "a\nfoo   \nb\n",
			diff: "@@ -1,3 +1,3 @@\n" +
				" a\n" +
				" foo\n" +
				"-b\n" +
				"+B\n",
			want:      "a\nfoo   \nB\n",
			wantHunks: 1,
		},
		{
			name:    "a stale context line around a change the file has once",
			content: "a\nb\nc\n",
			diff: "@@ -1,2 +1,2 @@\n" +
				" z\n" +
				"-b\n" +
				"+B\n",
			want:      "a\nB\nc\n",
			wantHunks: 1,
		},
		{
			name:    "the duplicate its context points at",
			content: "dup\nx\ndup\ny\ndup\n",
			diff: "@@ -2,3 +2,2 @@\n" +
				" x\n" +
				"-dup\n" +
				" y\n",
			want:      "dup\nx\ny\ndup\n",
			wantHunks: 1,
		},
		{
			name:    "identical lines: the header number picks one",
			content: strings.Repeat("cfg = 1\n", 5),
			diff: "@@ -3,1 +3,1 @@\n" +
				"-cfg = 1\n" +
				"+cfg = 2\n",
			want:      "cfg = 1\ncfg = 1\ncfg = 2\ncfg = 1\ncfg = 1\n",
			wantHunks: 1,
		},
		{
			name:    "a file without a trailing newline",
			content: "line1\nline2",
			diff: "@@ -2,1 +2,1 @@\n" +
				"-line2\n" +
				"+line2 changed\n",
			want:      "line1\nline2 changed",
			wantHunks: 1,
		},
		{
			name:    "crlf line endings",
			content: "line1\r\nline2\r\n",
			diff: "@@ -2,1 +2,1 @@\n" +
				"-line2\n" +
				"+line2 changed\n",
			want:      "line1\r\nline2 changed\r\n",
			wantHunks: 1,
		},
		{
			name:    "a crlf file with repeated lines",
			content: "cfg = 1\r\ncfg = 1\r\ncfg = 1\r\n",
			diff: "@@ -2,1 +2,1 @@\n" +
				"-cfg = 1\n" +
				"+cfg = 2\n",
			want:      "cfg = 1\r\ncfg = 2\r\ncfg = 1\r\n",
			wantHunks: 1,
		},
		{
			name:      "an addition past the end of a crlf file",
			content:   "a\r\nb\r\n",
			diff:      "@@ -9,0 +10 @@\n+c\n",
			want:      "a\r\nb\r\nc\r\n",
			wantHunks: 1,
		},
		{
			name:      "an empty file",
			content:   "",
			diff:      "@@ -0,0 +1,2 @@\n+x\n+y\n",
			want:      "x\ny\n",
			wantHunks: 1,
		},
		{
			name:      "every line deleted",
			content:   "only\n",
			diff:      "@@ -1 +0,0 @@\n-only\n",
			want:      "",
			wantHunks: 1,
		},
		{
			name:    "multi-byte content is edited by line, not by byte",
			content: "заголовок\nстатус: черновик\nконец\n",
			diff: "@@ -1,1 +1,1 @@\n" +
				"-статус: черновик\n" +
				"+статус: готово\n",
			want:      "заголовок\nстатус: готово\nконец\n",
			wantHunks: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, hunks, err := ApplyUnifiedDiff(tt.content, tt.diff)
			if err != nil {
				t.Fatalf("ApplyUnifiedDiff() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ApplyUnifiedDiff() content = %q, want %q", got, tt.want)
			}
			if hunks != tt.wantHunks {
				t.Errorf("ApplyUnifiedDiff() hunks applied = %d, want %d", hunks, tt.wantHunks)
			}
		})
	}
}

func TestFileDiff_ApplyUnifiedDiff_RefusesADiffItCannotApplyWhole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		diff    string
		wantErr string
	}{
		{
			name:    "an empty diff",
			content: "line1\n",
			diff:    "",
			wantErr: "diff is empty",
		},
		{
			name:    "a diff of blank lines",
			content: "line1\n",
			diff:    "   \n\n  ",
			wantErr: "no hunks",
		},
		{
			name:    "text with no hunk header",
			content: "line1\n",
			diff:    "just some text\nwith no diff markers\n",
			wantErr: "expected a hunk header",
		},
		{
			name:    "a stray line after the file headers",
			content: "a\n",
			diff:    "--- a/file\nsome garbage line\n@@ -1,1 +1,1 @@\n-a\n+A\n",
			wantErr: "expected a hunk header",
		},
		{
			name:    "a hunk line with an unknown prefix",
			content: "line1\nline2\n",
			diff:    "@@ -1,2 +1,2 @@\n line1\n*line2\n",
			wantErr: "invalid diff line",
		},
		{
			name:    "a hunk header with no lines under it",
			content: "line1\nline2\n",
			diff:    "@@ -1,1 +1,1 @@\n@@ -2,1 +2,1 @@\n-line2\n+line2 changed\n",
			wantErr: "no content lines",
		},
		{
			name:    "a hunk of context lines only",
			content: "a\nb\n",
			diff:    "@@ -1,2 +1,2 @@\n a\n b\n",
			wantErr: "has no '-' or '+' line, so it changes nothing",
		},
		{
			name:    "a removed line the file does not have",
			content: "line1\nline2\nline3\n",
			diff: "@@ -2,1 +2,1 @@\n" +
				"-this line does not exist in the file\n" +
				"+replacement\n",
			wantErr: `@@ -2,1 +2,1 @@ (not found in the file, looked for: "this line does not exist in the file\n")`,
		},
		{
			name:    "a one-character typo in the removed line",
			content: "Status: draft\nOwner: alice\nPriority: low\n",
			diff: "@@ -2,1 +2,1 @@\n" +
				"-Owner: alicia\n" +
				"+Owner: carol\n",
			wantErr: "could not be applied",
		},
		{
			name:    "indentation that differs from the file",
			content: "func f() {\n\treturn 1\n}\n",
			diff: "@@ -2,1 +2,1 @@\n" +
				"-    return 1\n" +
				"+    return 2\n",
			wantErr: "could not be applied",
		},
		{
			name:    "a stale context line around a change the file has twice",
			content: "a\nb\na\nb\n",
			diff: "@@ -1,2 +1,2 @@\n" +
				" z\n" +
				"-b\n" +
				"+B\n",
			wantErr: "could not be applied",
		},
		{
			name:    "one bad hunk among good ones",
			content: "a\nb\nc\n",
			diff: "@@ -1,1 +1,1 @@\n" +
				"-a\n" +
				"+A\n" +
				"@@ -2,1 +2,1 @@\n" +
				"-nope\n" +
				"+X\n",
			wantErr: "1 of 2 hunk(s)",
		},
		{
			name:    "two hunks that change the same line",
			content: "a\n",
			diff: "@@ -1 +1 @@\n" +
				"-a\n" +
				"+A\n" +
				"@@ -1 +1 @@\n" +
				"-a\n" +
				"+B\n",
			wantErr: "overlaps an earlier hunk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, hunks, err := ApplyUnifiedDiff(tt.content, tt.diff)
			if err == nil {
				t.Fatalf("ApplyUnifiedDiff() expected error containing %q, got nil (result: %q)", tt.wantErr, got)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("ApplyUnifiedDiff() error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
			if got != "" {
				t.Errorf("ApplyUnifiedDiff() on error must return empty content, got %q", got)
			}
			if hunks != 0 {
				t.Errorf("ApplyUnifiedDiff() on error reports %d hunks applied, want 0", hunks)
			}
		})
	}
}
