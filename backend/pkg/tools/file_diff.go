package tools

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// maxDiffHunkPreviewBytes bounds how much of a failed hunk's "old content"
// is ever echoed back to the LLM in an error message.
const maxDiffHunkPreviewBytes = 200

// unifiedDiffHunkHeaderRe matches a unified-diff hunk header. The strict
// form is "@@ -12,3 +12,4 @@"; the old/new line counts are optional (default
// to 1, per the unified diff spec) and, like the line numbers, are treated
// only as hints - see ApplyUnifiedDiff. The entire "-old +new" position
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
	// from it - such a hunk is located by the text of its lines alone.
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

func splitFileLines(content string) ([]string, bool) {
	if content == "" {
		return nil, true
	}
	hasEOL := strings.HasSuffix(content, "\n")
	if hasEOL {
		content = content[:len(content)-1]
	}
	return strings.Split(content, "\n"), hasEOL
}

func joinFileLines(lines []string, hasEOL bool) string {
	if len(lines) == 0 {
		return ""
	}
	out := strings.Join(lines, "\n")
	if hasEOL {
		out += "\n"
	}
	return out
}

func changeSpan(h diffHunk) (int, int, bool) {
	first, last := -1, -1
	for i, l := range h.lines {
		if l.sign != ' ' {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	return first, last, first >= 0
}

func hunkWindow(h diffHunk, fuzz int) []diffHunkLine {
	first, last, ok := changeSpan(h)
	if !ok {
		return nil
	}
	return h.lines[max(first-fuzz, 0):min(last+1+fuzz, len(h.lines))]
}

func windowOldSide(win []diffHunkLine) []string {
	out := make([]string, 0, len(win))
	for _, l := range win {
		if l.sign != '+' {
			out = append(out, l.text)
		}
	}
	return out
}

const relaxedTrimSet = " \t\r"

func linesEqual(fileLine, diffLine string, relaxed bool) bool {
	if strings.TrimSuffix(fileLine, "\r") == strings.TrimSuffix(diffLine, "\r") {
		return true
	}
	return relaxed && strings.TrimRight(fileLine, relaxedTrimSet) == strings.TrimRight(diffLine, relaxedTrimSet)
}

func findMatches(lines, want []string, relaxed bool) []int {
	if len(want) == 0 || len(want) > len(lines) {
		return nil
	}
	var out []int
	for i := 0; i+len(want) <= len(lines); i++ {
		ok := true
		for j := range want {
			if !linesEqual(lines[i+j], want[j], relaxed) {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, i)
		}
	}
	return out
}

func absDiff(a, b int) int { return max(a-b, b-a) }

func pickMatch(cands []int, h diffHunk, searchFrom int) int {
	inOrder := make([]int, 0, len(cands))
	for _, c := range cands {
		if c >= searchFrom {
			inOrder = append(inOrder, c)
		}
	}
	if len(inOrder) == 0 {
		inOrder = cands
	}
	if !h.hasPosition {
		return inOrder[0]
	}
	hint := h.oldStart - 1
	best := inOrder[0]
	for _, c := range inOrder[1:] {
		if absDiff(c, hint) < absDiff(best, hint) {
			best = c
		}
	}
	return best
}

type locatedHunk struct {
	hunk   diffHunk
	win    []diffHunkLine
	start  int
	oldLen int
}

func locateHunk(lines []string, h diffHunk, searchFrom int) (locatedHunk, bool) {
	first, last, hasChange := changeSpan(h)
	if !hasChange {
		return locatedHunk{}, false
	}
	maxFuzz := max(first, len(h.lines)-1-last)

	full := hunkWindow(h, maxFuzz)
	if want := windowOldSide(full); len(want) > 0 {
		if cands := findMatches(lines, want, false); len(cands) > 0 {
			return locatedHunk{hunk: h, win: full, start: pickMatch(cands, h, searchFrom), oldLen: len(want)}, true
		}
		if core := windowOldSide(hunkWindow(h, 0)); len(core) > 0 && len(findMatches(lines, core, true)) == 0 {
			return locatedHunk{}, false
		}
	}

	for fuzz := maxFuzz; fuzz >= 0; fuzz-- {
		win := hunkWindow(h, fuzz)
		want := windowOldSide(win)
		if len(want) == 0 {
			continue
		}
		for _, relaxed := range []bool{false, true} {
			cands := findMatches(lines, want, relaxed)
			if len(cands) != 1 {
				continue
			}
			return locatedHunk{hunk: h, win: win, start: cands[0], oldLen: len(want)}, true
		}
	}

	if len(windowOldSide(h.lines)) == 0 && (h.hasPosition || len(lines) == 0) {
		start := min(max(h.oldStart-1, 0), len(lines))
		return locatedHunk{hunk: h, win: h.lines, start: start, oldLen: 0}, true
	}
	return locatedHunk{}, false
}

func hunkEOL(lines []string, start int) string {
	idx := start
	if idx >= len(lines) {
		idx = len(lines) - 1
	}
	if idx >= 0 && strings.HasSuffix(lines[idx], "\r") {
		return "\r"
	}
	return ""
}

func renderNewSide(l locatedHunk, lines []string) []string {
	eol := hunkEOL(lines, l.start)
	out := make([]string, 0, len(l.win))
	cursor := l.start
	for _, hl := range l.win {
		switch hl.sign {
		case ' ':
			out = append(out, lines[cursor])
			cursor++
		case '-':
			cursor++
		case '+':
			out = append(out, strings.TrimSuffix(hl.text, "\r")+eol)
		}
	}
	return out
}

func applyLocatedHunks(lines []string, located []locatedHunk) ([]string, error) {
	ordered := make([]locatedHunk, len(located))
	copy(ordered, located)
	slices.SortStableFunc(ordered, func(a, b locatedHunk) int { return a.start - b.start })

	out := make([]string, 0, len(lines))
	cursor := 0
	for _, l := range ordered {
		if l.start < cursor {
			return nil, fmt.Errorf("hunk %q overlaps an earlier hunk in the same diff", l.hunk.header)
		}
		out = append(out, lines[cursor:l.start]...)
		out = append(out, renderNewSide(l, lines)...)
		cursor = l.start + l.oldLen
	}
	return append(out, lines[cursor:]...), nil
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

// ApplyUnifiedDiff applies a unified diff to content entirely in memory. A
// hunk is located by matching its context and removed lines against the file
// line by line; the header's line number never decides where a hunk lands and
// only breaks ties between equally good matches, so a stale number applies
// correctly while a removed line the file does not carry is refused. It
// returns the patched content
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

	lines, hasEOL := splitFileLines(content)

	located := make([]locatedHunk, 0, len(hunks))
	var failed []string
	searchFrom := 0
	for _, h := range hunks {
		if _, _, ok := changeSpan(h); !ok {
			failed = append(failed, fmt.Sprintf("%s (has no '-' or '+' line, so it changes nothing)", h.header))
			continue
		}
		l, ok := locateHunk(lines, h, searchFrom)
		if !ok {
			failed = append(failed, fmt.Sprintf("%s (not found in the file, looked for: %q)", h.header, hunkOldPreview(h)))
			continue
		}
		located = append(located, l)
		searchFrom = l.start + l.oldLen
	}
	if len(failed) > 0 {
		return "", 0, fmt.Errorf(
			"%d of %d hunk(s) could not be applied - read the file again and retry with context that matches its current content exactly:\n%s",
			len(failed), len(hunks), strings.Join(failed, "\n"),
		)
	}

	newLines, err := applyLocatedHunks(lines, located)
	if err != nil {
		return "", 0, err
	}
	return joinFileLines(newLines, hasEOL), len(hunks), nil
}
