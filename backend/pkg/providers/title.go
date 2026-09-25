package providers

import (
	"regexp"
	"strings"
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
