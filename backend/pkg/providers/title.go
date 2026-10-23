package providers

import (
	"regexp"
	"strings"
)

const (
	// taskTitleLimit is the length the task title prompt asks for and the fallback is cut to.
	taskTitleLimit = 150
	untitledTask   = "Untitled task"
)

const titleSpan = `[\p{L}\p{N}](?:[^*]*[^*\s\p{L}\p{N}]\*+)*(?:[^*]*[^*\s])?`

var (
	titleWrappers = []struct{ wrapper, closer *regexp.Regexp }{
		{
			regexp.MustCompile(`^\*\*\*([^\s*](?:.*\S)?)\*\*\*$`),
			regexp.MustCompile(`[^\s*]\*\*\*(?:[\s,.;:!?]|$)`),
		},
		{
			regexp.MustCompile(`^\*\*([^\s*](?:.*\S)?)\*\*$`),
			regexp.MustCompile(`[^\s*]\*\*(?:[\s,.;:!?]|$)`),
		},
		{
			regexp.MustCompile(`^\*([\p{L}\p{N}](?:.*\S)?)\*$`),
			regexp.MustCompile(`[\p{L}\p{N}]\*(?:[\s,.;:!?]|$)`),
		},
	}
	titleEmphasis = regexp.MustCompile(
		`(^|\s)(?:\*\*\*(` + titleSpan + `)\*\*\*|\*\*(` + titleSpan + `)\*\*|\*(` + titleSpan + `)\*)`,
	)
	titleControlToken = regexp.MustCompile(`<\|[^|<>\s]{1,64}\|>`)
)

func normalizeTitle(answer string) string {
	title := strings.Join(strings.Fields(titleControlToken.ReplaceAllString(answer, " ")), " ")
	for _, w := range titleWrappers {
		if m := w.wrapper.FindStringSubmatch(title); m != nil {
			if !w.closer.MatchString(m[1]) {
				title = m[1]
			}
			break
		}
	}
	return titleEmphasis.ReplaceAllString(title, "$1$2$3$4")
}

// titleFromInput derives a title from the first non-blank line of input, cut to limit runes.
func titleFromInput(input string, limit int, placeholder string) string {
	for line := range strings.Lines(input) {
		title := []rune(strings.Join(strings.Fields(line), " "))
		if len(title) == 0 {
			continue
		}
		if len(title) > limit {
			title = append(title[:limit-1], '…')
		}
		return string(title)
	}
	return placeholder
}
