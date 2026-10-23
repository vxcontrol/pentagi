package markdown

import "testing"

func TestMarkdown_CodeBlock_SizesTheFenceToTheContent(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		lang    string
		content string
		want    string
	}{
		{name: "content without backticks", content: "id", want: "```\nid\n```"},
		{name: "empty content", want: "```\n\n```"},
		{name: "a language after the opening fence", lang: "python", content: "print(1)", want: "```python\nprint(1)\n```"},
		{name: "a single backtick", content: "echo `id`", want: "```\necho `id`\n```"},
		{name: "a run of two", content: "a``b", want: "```\na``b\n```"},
		{name: "a run of three", content: "a\n```\nb", want: "````\na\n```\nb\n````"},
		{name: "a run of five", content: "a\n`````\nb", want: "``````\na\n`````\nb\n``````"},
		{name: "a run at the very start", content: "```sh\nid", want: "````\n```sh\nid\n````"},
		{name: "a run at the very end", content: "id\n```", want: "````\nid\n```\n````"},
		{name: "a run inside a line", content: "see ``` here", want: "````\nsee ``` here\n````"},
		{name: "the longest of several runs", content: "``` and ````` and `", want: "``````\n``` and ````` and `\n``````"},
		{name: "a run of three with a language", lang: "md", content: "```\nx\n```", want: "````md\n```\nx\n```\n````"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := CodeBlock(tt.lang, tt.content); got != tt.want {
				t.Errorf("CodeBlock(%q, %q) = %q, want %q", tt.lang, tt.content, got, tt.want)
			}
		})
	}
}
