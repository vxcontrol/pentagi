package tools

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// maxDiffHunkPreviewBytes bounds how much of a failed hunk's "old content"
// is ever echoed back to the LLM in an error message.
const maxDiffHunkPreviewBytes = 200

// unifiedDiffHunkHeaderRe matches a unified-diff hunk header. The strict
// form is "@@ -12,3 +12,4 @@"; the old/new line counts are optional (default
// to 1, per the unified diff spec) and, like the line numbers, are treated
// only as hints - see applyUnifiedDiff. The entire "-old +new" position
// clause is ALSO optional, tolerating a bare "@@" some models emit when
// they're unsure of exact line numbers: see hasPosition on diffHunk for how
// that case is handled downstream.
var unifiedDiffHunkHeaderRe = regexp.MustCompile(`^@@(?:\s+-(\d+)(?:,(\d+))?\s+\+(\d+)(?:,(\d+))?)?\s*(?:@@)?`)

// diffHunkLine is one line of a hunk body: sign is ' ' (context), '-'
// (removed), or '+' (added); text excludes the sign and any line terminator.
type diffHunkLine struct {
	sign byte
	text string
}

// diffHunk is one parsed "@@ ... @@" section of a unified diff.
type diffHunk struct {
	header string // original header line, kept only for error messages
	// oldStart is the 1-based line number in the current file where the
	// hunk begins - defaults to 1 when the header omitted it (hasPosition
	// false), in which case it is NOT a real hint and callers must not
	// derive anything positional (e.g. "the line before/after this hunk")
	// from it - only diffmatchpatch's own content-based fuzzy search can
	// locate such a hunk.
	oldStart    int
	hasPosition bool
	lines       []diffHunkLine
}

// parseUnifiedDiff parses a unified diff into hunks. It tolerates the parts
// of the format that vary across generators without affecting correctness:
// optional "--- a/file" / "+++ b/file" headers (the path is already a
// separate argument), "\ No newline at end of file" markers, and a fully
// blank line standing in for a one-space (empty) context line.
func parseUnifiedDiff(diffText string) ([]diffHunk, error) {
	normalized := strings.TrimSuffix(strings.ReplaceAll(diffText, "\r\n", "\n"), "\n")
	if normalized == "" {
		return nil, fmt.Errorf("diff is empty")
	}
	lines := strings.Split(normalized, "\n")

	i := 0
	for i < len(lines) && !strings.HasPrefix(lines[i], "@@") {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed != "" && !strings.HasPrefix(trimmed, "---") && !strings.HasPrefix(trimmed, "+++") {
			return nil, fmt.Errorf("expected a hunk header (\"@@ -old +new @@\") but found: %q", lines[i])
		}
		i++
	}
	if i >= len(lines) {
		return nil, fmt.Errorf(`diff contains no hunks (no "@@ ... @@" header found)`)
	}

	var hunks []diffHunk
	for i < len(lines) {
		header := lines[i]
		m := unifiedDiffHunkHeaderRe.FindStringSubmatch(header)
		if m == nil {
			return nil, fmt.Errorf("invalid hunk header: %q", header)
		}
		hunk := diffHunk{header: header, oldStart: 1}
		if m[1] != "" {
			oldStart, err := strconv.Atoi(m[1])
			if err != nil {
				return nil, fmt.Errorf("invalid hunk header %q: %w", header, err)
			}
			hunk.oldStart = oldStart
			hunk.hasPosition = true
		}
		i++

		for i < len(lines) && !strings.HasPrefix(lines[i], "@@") {
			line := lines[i]
			i++

			if strings.HasPrefix(line, `\`) {
				// e.g. "\ No newline at end of file" - not a content line.
				continue
			}
			if line == "" {
				hunk.lines = append(hunk.lines, diffHunkLine{sign: ' ', text: ""})
				continue
			}

			sign := line[0]
			if sign != ' ' && sign != '-' && sign != '+' {
				return nil, fmt.Errorf("invalid diff line (must start with ' ', '-', or '+'): %q", line)
			}
			hunk.lines = append(hunk.lines, diffHunkLine{sign: sign, text: line[1:]})
		}

		if len(hunk.lines) == 0 {
			return nil, fmt.Errorf("hunk %q has no content lines", header)
		}
		hunks = append(hunks, hunk)
	}

	return hunks, nil
}

// buildLineOffsets returns, for content, the byte offset at which each
// 1-based line begins: offsets[0] is the (always 0) offset of line 1,
// offsets[1] of line 2, and so on.
func buildLineOffsets(content string) []int {
	offsets := make([]int, 1, strings.Count(content, "\n")+1)
	offsets[0] = 0
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			offsets = append(offsets, i+1)
		}
	}
	return offsets
}

// lineOffset returns the byte offset where 1-based lineNum begins, clamping
// to the start/end of content for out-of-range line numbers rather than
// panicking - an inaccurate hunk header should surface as a clear "hunk
// didn't apply" error, not a crash.
func lineOffset(offsets []int, content string, lineNum int) int {
	if lineNum < 1 {
		return 0
	}
	if lineNum-1 < len(offsets) {
		return offsets[lineNum-1]
	}
	return len(content)
}

// encodePatchTextLine renders one hunk line in the wire format
// diffmatchpatch.PatchFromText expects: a sign byte followed by the line
// text URL-escaped, with the escaper's space-as-'+' swapped back to a
// literal space (mirroring Patch.String in the go-diff source exactly, so
// PatchFromText's decoder - which reverses precisely that transform -
// reconstructs the original text, including any literal '+' or '%').
func encodePatchTextLine(sign byte, text string) string {
	escaped := strings.ReplaceAll(url.QueryEscape(text), "+", " ")
	return string(sign) + escaped
}

// ensureContextBoundaries returns hunks where the first hunk (if missing
// leading context) and/or the last hunk (if missing trailing context) gets
// one synthesized from the file's actual adjacent line, taken verbatim.
// Hunks in between are left exactly as parsed - they don't need this (see
// below) and, unlike the first/last hunk, a wrong hint there could inject
// misleading context instead of merely being redundant.
//
// diffmatchpatch's PatchApply pads every patch list (see PatchAddPadding in
// the go-diff source) by: (a) treating the *first* patch as if it started at
// position 0 when its first diff isn't DiffEqual, and (b) appending its own
// filler bytes as a fake trailing DiffEqual onto the *last* patch when its
// last diff isn't one either. Both corrupt that hunk's search pattern
// unless it is genuinely at the very start/end of the file - the common
// case is a single-line change with no context on one side, not actually at
// either file boundary. A real leading/trailing context line - which
// unified diffs otherwise tend to include anyway - sidesteps both special
// cases rather than relying on them being correct for hand-built patches.
func ensureContextBoundaries(hunks []diffHunk, content string) []diffHunk {
	if len(hunks) == 0 {
		return hunks
	}

	var fileLines []string // split lazily; most diffs don't need it
	linesOf := func() []string {
		if fileLines == nil {
			fileLines = strings.Split(content, "\n")
		}
		return fileLines
	}

	out := make([]diffHunk, len(hunks))
	copy(out, hunks)

	// Both fixes below derive "the file line right before/after this hunk"
	// from oldStart - meaningless, and actively misleading, for a hunk whose
	// header omitted its position entirely (hasPosition false, oldStart just
	// defaults to 1). Such a hunk relies purely on diffmatchpatch's own
	// content-based fuzzy search, which already handles the Start1==0 case
	// (the only one this function exists to work around) correctly on its
	// own - nothing to fix up here.
	if h := out[0]; h.hasPosition && len(h.lines) > 0 && h.lines[0].sign != ' ' && h.oldStart > 1 {
		lines := linesOf()
		precedingIdx := h.oldStart - 2 // 0-based index of file line (oldStart-1)
		if precedingIdx >= 0 && precedingIdx < len(lines) {
			h.oldStart--
			h.lines = append(
				[]diffHunkLine{{sign: ' ', text: lines[precedingIdx]}},
				h.lines...,
			)
			out[0] = h
		}
	}

	lastIdx := len(out) - 1
	if h := out[lastIdx]; h.hasPosition && len(h.lines) > 0 && h.lines[len(h.lines)-1].sign != ' ' {
		lines := linesOf()
		oldLineCount := 0
		for _, l := range h.lines {
			if l.sign != '+' {
				oldLineCount++
			}
		}
		followingIdx := h.oldStart - 1 + oldLineCount // 0-based index of the line right after the hunk
		// The last split element past a trailing "\n" is a Split artifact
		// (nothing actually follows it), not a real line to match against.
		isEOFArtifact := followingIdx == len(lines)-1 && strings.HasSuffix(content, "\n")
		if followingIdx >= 0 && followingIdx < len(lines) && !isEOFArtifact {
			h.lines = append(h.lines, diffHunkLine{sign: ' ', text: lines[followingIdx]})
			out[lastIdx] = h
		}
	}

	return out
}

// buildGoDiffPatchText renders parsed hunks into diffmatchpatch's patch text
// format. The "@@ -start,len +start,len @@" header uses byte offsets
// computed against the real, current file content - not the line numbers
// the LLM supplied, which only ever serve as a locate-nearby hint once the
// patch is applied (diffmatchpatch.PatchApply falls back to fuzzy matching
// keyed on the header's position when the exact offset misses).
func buildGoDiffPatchText(hunks []diffHunk, content string) string {
	offsets := buildLineOffsets(content)

	var b strings.Builder
	for _, h := range hunks {
		start := lineOffset(offsets, content, h.oldStart)

		var oldLen, newLen int
		for _, l := range h.lines {
			switch l.sign {
			case ' ':
				oldLen += len(l.text) + 1
				newLen += len(l.text) + 1
			case '-':
				oldLen += len(l.text) + 1
			case '+':
				newLen += len(l.text) + 1
			}
		}

		fmt.Fprintf(&b, "@@ -%d,%d +%d,%d @@\n", start+1, oldLen, start+1, newLen)
		for _, l := range h.lines {
			b.WriteString(encodePatchTextLine(l.sign, l.text+"\n"))
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// hunkOldPreview renders the pre-patch text a hunk searched for (context and
// removed lines only), for use in a "hunk didn't apply" error message.
func hunkOldPreview(h diffHunk) string {
	var b strings.Builder
	for _, l := range h.lines {
		if l.sign != '+' {
			b.WriteString(l.text)
			b.WriteByte('\n')
		}
	}
	return truncateString(b.String(), maxDiffHunkPreviewBytes)
}

// ApplyUnifiedDiff applies a unified diff to content entirely in memory,
// using diffmatchpatch.PatchApply for the actual merge: an exact match at
// the hunk's expected position is tried first, falling back to a
// fuzzy/context-based search nearby (tolerating minor line-number drift)
// before a hunk is considered unappliable. It returns the patched content
// and the number of hunks applied, or a descriptive error naming every hunk
// that failed to apply and a preview of the content it looked for -
// content is returned unchanged (empty) on error, so a partial/bad diff
// never corrupts the file. Exported so other packages (e.g. the provider
// tester) can exercise the exact production diff-merge semantics without
// going through EditFile's Docker-backed read/write.
func ApplyUnifiedDiff(content, diffText string) (string, int, error) {
	hunks, err := parseUnifiedDiff(diffText)
	if err != nil {
		return "", 0, err
	}
	hunks = ensureContextBoundaries(hunks, content)

	patchText := buildGoDiffPatchText(hunks, content)

	dmp := diffmatchpatch.New()
	patches, err := dmp.PatchFromText(patchText)
	if err != nil {
		return "", 0, fmt.Errorf("internal error building patch: %w", err)
	}

	newContent, applied := dmp.PatchApply(patches, content)

	var failed []string
	for i, ok := range applied {
		if !ok && i < len(hunks) {
			failed = append(failed, fmt.Sprintf("%s (not found in the file, looked for: %q)", hunks[i].header, hunkOldPreview(hunks[i])))
		}
	}
	if len(failed) > 0 {
		return "", 0, fmt.Errorf(
			"%d of %d hunk(s) could not be applied - read the file again and retry with context that matches its current content exactly:\n%s",
			len(failed), len(hunks), strings.Join(failed, "\n"),
		)
	}

	return newContent, len(hunks), nil
}
