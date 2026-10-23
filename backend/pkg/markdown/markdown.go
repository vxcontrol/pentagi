// Package markdown holds helpers for writing Markdown around text PentAGI does not control.
package markdown

import "strings"

// CodeBlock wraps content in a fenced code block. The fence is one backtick longer than the
// longest backtick run in content, three at least, so no line of content can close the block early.
// Any run counts, not only one that alone makes up a line, so the fence can be longer than strictly needed.
func CodeBlock(lang, content string) string {
	longest, run := 0, 0
	for i := range len(content) {
		if content[i] != '`' {
			run = 0
			continue
		}
		run++
		longest = max(longest, run)
	}
	fence := strings.Repeat("`", max(3, longest+1))

	return fence + lang + "\n" + content + "\n" + fence
}
