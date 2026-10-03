import type { JSONContent } from '@tiptap/core';

import { Editor } from '@tiptap/core';
import { beforeAll, describe, expect, it } from 'vitest';

import { GROWTH_IF_QUADRATIC, slowdownWhenInputQuadruples } from '@/test-utils/cost-growth';

import { createMarkdownExtensions } from './markdown-editor-extensions';
import { markdownCodec, roundTrip, setupEditorJsdom } from './markdown-editor-test-setup';

beforeAll(setupEditorJsdom);

const load = (content: JSONContent | string) =>
    new Editor({
        content,
        ...(typeof content === 'string' ? { contentType: 'markdown' as const } : {}),
        extensions: createMarkdownExtensions(),
    });

const textsOf = (markdown: string): null | string[] => {
    const editor = load(markdown);
    const texts = JSON.stringify(editor.getJSON()).match(/"text":"[^"]*"/g);

    editor.destroy();

    return texts;
};

// The inline nodes of every paragraph, a line break as `/`.
const linesOf = (markdown: string): string[] => {
    const editor = load(markdown);
    const paragraphs: string[] = [];

    editor.state.doc.descendants((node) => {
        if (node.type.name === 'paragraph') {
            const inline: string[] = [];

            node.forEach((child) => inline.push(child.isText ? (child.text ?? '') : '/'));
            paragraphs.push(inline.join('|'));
        }
    });
    editor.destroy();

    return paragraphs;
};

// What the document is after `edit`, and what it is saved as.
const afterEdit = (content: JSONContent | string, edit: (editor: Editor) => void = () => {}): string => {
    const editor = load(content);

    edit(editor);

    const saved = editor.getMarkdown();

    editor.destroy();

    return saved;
};

describe('inline text is saved the way it was written', () => {
    it.each([
        ['five asterisks', 'at §*****HTTP URL with Credentials*****§ now'],
        ['uneven asterisks', '| a |\n|---|\n| §**ipv4***§ |'],
        ['bold around it', '**§*CVE Number*§ Advisory**'],
        ['emphasis around it', '*a §****ipv4*****§ b*'],
        ['a backslash before it', '*a \\§****ipv4*****§ b*'],
        ['a closer left over from a second pass', '**URL**: §*****HTTP URL*****§ Name Enhanced*§.net'],
        ['an asterisk right after it', 'text §*a*§*b*'],
        ['bold earlier on its line', '**Note:** see **records**. Spec: §***HTTP URL****§ Name Enhanced*§.com/en-us/x'],
    ])('asterisks of an anonymizer placeholder, with %s', (_name, md) => {
        expect(roundTrip(md)).toBe(md);
    });

    it.each([
        ['padded with spaces', 'use `` alert(`XSS`) `` and ` a ` here'],
        ['padded because it holds a backtick', 'a `` `x` `` b'],
        ['unbalanced by a stray backtick', 'see `curl x and the `Set-Cookie` verbatim (keep `Bearer ` now'],
        ['that runs over a line', 'A PHP `Warning: failed to\nopen stream` message'],
        ['that runs over a line in an item', '- see `foo\n  bar` baz\n- next'],
        ['inside emphasis over a line', '**x `c\nd` y**'],
        ['that opens and ends with a line break', '`\nfoo\n`'],
    ])('a code span %s', (_name, md) => {
        expect(roundTrip(md)).toBe(md);
    });

    it.each([
        ['emphasis over a line', '*a\nb* and **c\nd**'],
        ['emphasis on each of two lines', '*a*\n*b*'],
        ['a link on each of two lines', '[x](/u "t")\n[x](/u "t")'],
        ['a link text over a line', '[l\nm](/u)'],
        ['asterisks that close nothing', '- **No /dev/sd*, /dev/vd*, /dev/fuse**'],
        ['a line indented four past its item', '- item\n        indented\n  more\n- next'],
        ['a numbered line that is the text of a loose item', '- 1. First thing\n\n- 2. Second thing'],
        ['a backslash before a hash inside emphasis', '**\\#hashtag** and *\\>x*'],
        ['a backslash before a hash in a link text', '[\\#1](http://a.b) text'],
        ['a backslash before a hash in struck text', 'text ~~\\#no~~'],
    ])('%s', (_name, md) => {
        expect(roundTrip(md)).toBe(md);
    });
});

describe('a Go template action is literal text', () => {
    // Byte identity alone passes for a mark that writes its own delimiters back, hence the mark count.
    it.each([
        ['an entity in a string', '{{ printf "&amp; &lt;x&gt;" }} done', '"italic"', 0],
        ['a quote entity in a string', '{{ printf "&quot;%s&quot;" .Name }}', '"italic"', 0],
        ['an entity in a string after text', 'text {{ printf "&lt;" }} more', '"italic"', 0],
        ['a raw string beside a code span', '{{ printf `%s` .A }} and `code` and {{ `x` }}', '"code"', 1],
        ['a raw string after text, beside a code span', 'run {{ printf `a` }} and `code`', '"code"', 1],
        ['a raw string with a pipe in a cell', '| a | b |\n|---|---|\n| {{ printf `a|b` }} | 2 |', '"code"', 0],
        ['an asterisk in a comment', '*a {{/* c */}} b*', '"italic"', 1],
        ['asterisks inside and outside', '*note {{ mul 2 3 }}* and {{/* c */}} *x*', '"italic"', 2],
        ['a URL in a string', '{{ template "x" "https://a.b/c" }} and https://a.b/d', '"link"', 1],
        ['a URL that runs into it', 'See https://example.com/{{ .Path }} now', '"link"', 0],
        ['bold around it', '**{{ .Name }}** and *{{ .X | print }}*', '"bold"', 1],
        ['a brace in a string', 'send {{ printf "{%s} &lt;" .X }} as JSON', '"link"', 0],
        ['a quote in a comment', 'x {{/* 5" &lt; */}} y', '"link"', 0],
        ['an apostrophe that opens no rune', "{{ .Team's &amp; }} done", '"link"', 0],
        ['JSON around it', '{"name": "{{ printf "&lt;%s" .Name }}", "n": {{ .N }}}', '"link"', 0],
        [
            'a brace in a string, in a code span in a cell',
            '| a |\n|---|\n| `{{ printf "{%s}" .X | print }}` |',
            '"code"',
            1,
        ],
        [
            'a brace in a string in a cell',
            '| a | b |\n|---|---|\n| {{ printf "{%s}" .X | print }} | 2 |',
            '"tableCell"',
            2,
        ],
        [
            'the end of an action in a string in a cell',
            '| a | b |\n|---|---|\n| {{ print "}}" | print }} | 2 |',
            '"tableCell"',
            2,
        ],
        [
            'a quote in a rune in a cell',
            '| a | b |\n|---|---|\n| {{ printf "%c" \'"\' | print }} | 2 |',
            '"tableCell"',
            2,
        ],
        [
            'a pipeline in a code span beside an escaped pipe',
            '| a | b |\n|---|---|\n| `x \\| {{ .A | print }}` | 2 |',
            '"tableCell"',
            2,
        ],
        [
            'a pipeline in a code span between bare pipes',
            '| a | b |\n|---|---|\n| `x | {{ .A | print }} | y` | 2 |',
            '"tableCell"',
            2,
        ],
        [
            'a raw string with a pipe beside an escaped pipe',
            '| a | b |\n|---|---|\n| {{ printf `a|b` }} | `p \\| q` |',
            '"tableCell"',
            2,
        ],
    ])('%s', (_name, md, kind, count) => {
        const editor = load(md);
        const found = JSON.stringify(editor.getJSON()).split(`"type":${kind}`).length - 1;

        editor.destroy();

        expect(found).toBe(count);
        expect(roundTrip(md)).toBe(md);
    });

    it('still decodes an entity in the text beside one', () => {
        expect(roundTrip('a &amp; b {{ .X }} c &lt; d')).toBe('a & b {{ .X }} c < d');
    });

    it('leaves emphasis between a code span that holds an opener and one that holds a closer to be read', () => {
        const editor = load('Jinja uses `{{` *expression* `}}` syntax');
        const italics = (editor.getJSON().content?.[0]?.content ?? [])
            .filter((node: JSONContent) => node.marks?.some((mark) => mark.type === 'italic'))
            .map((node: JSONContent) => node.text);

        editor.destroy();

        expect(italics).toEqual(['expression']);
    });

    it.each([
        ['one over two lines', '{{ if\n.X }}a{{ end }}', ['{{ if|/|.X }}a{{ end }}']],
        ['a pipeline over two lines', 'x {{ .Name\n| print }} y', ['x {{ .Name|/|| print }} y']],
    ])('keeps the line break inside %s as a break node', (_name, md, lines) => {
        expect(linesOf(md)).toEqual(lines);
        expect(roundTrip(md)).toBe(md);
    });

    it('is still whole after the URL before it is edited', () => {
        const saved = afterEdit('See https://example.com/{{ .Path }} now', (editor) =>
            editor.chain().setTextSelection(25).insertContent('x').run(),
        );

        expect(saved).toBe('See https://example.com/x{{ .Path }} now');
    });
});

describe('a pipe inside a code span of a table cell', () => {
    it.each([
        ['written escaped', '| a | b |\n| --- | --- |\n| `x \\| y` | z |'],
        ['written bare', '| a | b |\n| --- | --- |\n| `x | y` | z |'],
        ['written both ways', '| a | b |\n| --- | --- |\n| `x | y` | `p \\| q` |'],
        ['bare beside a pipe in a link', '| a | b |\n| --- | --- |\n| `x | y` | [l](http://h/?a=1|2) end |'],
        ['escaped inside an action, in a code span', '| a | b |\n| --- | --- |\n| `x{{364\\|add:733}}y` | z |'],
        ['escaped inside an action', '| a | b |\n| --- | --- |\n| {{ .A \\| upper }} | `p \\| q` |'],
        [
            'escaped inside one action and bare inside another',
            '| a | b |\n| --- | --- |\n| `{{ a \\| b }}` | {{ c | d }} |',
        ],
        ['bare under a header cell that holds a pipeline', '| {{ .A | print }} | b |\n| --- | --- |\n| `x | y` | z |'],
    ])('saves a pipe inside a code span %s as it was written', (_name, md) => {
        expect(roundTrip(md)).toBe(md);
    });

    it.each([
        [
            'a pipe in a code span',
            '| a | b |\n| --- | --- |\n| `x | y` | z |\n| `p | q` | r |',
            'z',
            '| a | b |\n| --- | --- |\n| `x \\| y` | z! |\n| `p | q` | r |',
        ],
        [
            'a pipe in a link',
            '| a | b |\n| --- | --- |\n| [go](https://h/?x=1|2) | z |',
            'z',
            '| a | b |\n| --- | --- |\n| [go](https://h/?x=1\\|2) | z! |',
        ],
        [
            'a pipe in an action, which is never escaped',
            '| a | b |\n| --- | --- |\n| `{{ a \\| b }}` {{ c | d }} | z |',
            'z',
            '| a | b |\n| --- | --- |\n| `{{ a | b }}` {{ c | d }} | z! |',
        ],
    ])('writes %s the way GFM asks once its row is edited', (_name, md, cell, saved) => {
        const edited = afterEdit(md, (editor) => {
            let at = 0;

            editor.state.doc.descendants((node, position) => {
                at = node.text === cell ? position + node.nodeSize : at;
            });
            editor.chain().setTextSelection(at).insertContent('!').run();
        });

        expect(edited).toBe(saved);
        expect(roundTrip(edited)).toBe(edited);
    });

    // marked counts the cells of the header row before anything protects a pipe in it.
    it('escapes a pipe typed into a code span of the header and leaves the rows under it as they were', () => {
        const saved = afterEdit('| a | b |\n|---|---|\n| `x|y` | 2 |', (editor) =>
            editor
                .chain()
                .setTextSelection(5)
                .insertContent({ marks: [{ type: 'code' }], text: 'm|n', type: 'text' })
                .run(),
        );

        expect(saved).toBe('| a`m\\|n` | b |\n|---|---|\n| `x|y` | 2 |');
        expect(roundTrip(saved)).toBe(saved);
    });

    it('escapes a pipe typed into a code span of a table made in the editor', () => {
        const cell = (type: string, content: object[]) => ({ content: [{ content, type: 'paragraph' }], type });
        const editor = new Editor({
            content: {
                content: [
                    {
                        content: [
                            { content: [cell('tableHeader', [{ text: 'a', type: 'text' }])], type: 'tableRow' },
                            {
                                content: [
                                    cell('tableCell', [{ marks: [{ type: 'code' }], text: 'x | y', type: 'text' }]),
                                ],
                                type: 'tableRow',
                            },
                        ],
                        type: 'table',
                    },
                ],
                type: 'doc',
            },
            extensions: createMarkdownExtensions(),
        });
        const saved = editor.getMarkdown();

        editor.destroy();

        expect(saved).toContain('`x \\| y`');
        expect(roundTrip(saved)).toBe(saved);
    });
});

describe('a line under a table that is not a row of it', () => {
    it.each([
        [
            "a template's range around the rows",
            '| a | b |\n|---|---|\n{{ range .Rows }}\n| {{ .Name }} | {{ .Val }} |\n{{ end }}',
            1,
        ],
        [
            "a range that opens on the first row's line",
            '| a | b |\n|---|---|\n{{ range .Rows }}| {{ .Name | print }} | {{ .Val }} |\n{{ end }}',
            1,
        ],
        [
            'a range around a table and its header',
            '{{ range .Tables }}\n| a | b |\n|---|---|\n| {{ .A }} | {{ .B }} |\n{{- end }}',
            2,
        ],
        ['a control action written with a tab', '| a | b |\n|---|---|\n| 1 | 2 |\n{{\tend }}', 2],
        ['an assignment on a line of its own', '| a | b |\n|---|---|\n{{- $n := len .Rows }}\n| 1 | 2 |', 1],
        ['a comment on a line of its own', '| a | b |\n|---|---|\n| 1 | 2 |\n{{/* rows */}}', 2],
        ['a pipeline on a line of its own', '| a |\n|---|\n{{ .X | print }}', 1],
        [
            'an action after a row has closed',
            '| a | b |\n|---|---|\n| static | row |{{ range .Items }}\n| {{ .Name }} | {{ .Value }} |{{ end }}',
            1,
        ],
        ['a row with a cell too many', '| a | b |\n|---|---|\n| 1 | 2 |\n| 3 | 4 | 5 |\n| 6 | 7 |', 2],
        ['a row with a cell too many in a list item', '- t:\n  | a | b |\n  |---|---|\n  | 1 | 2 | 3 |\n- next', 1],
        [
            'a control action on the line of a row that has no cell too many',
            '| a | b | c |\n|---|---|---|\n{{ if .A }}| x | y |\n{{ end }}',
            1,
        ],
        ['the same, with a tab after the braces', '| a | b | c |\n|---|---|---|\n{{\tif .A }}| x | y |\n{{ end }}', 1],
    ])('%s', (_name, md, rows) => {
        const editor = load(md);
        const found = JSON.stringify(editor.getJSON()).split('"type":"tableRow"').length - 1;

        editor.destroy();

        expect(found).toBe(rows);
        expect(roundTrip(md)).toBe(md);
    });

    it.each([
        [
            'an action whose name begins with a control word',
            '| a | b |\n|---|---|\n{{ endpoint }} | 2\n{{ iffy .X }} | 3',
            3,
        ],
        ['a row whose last cell is an action', '| a | b |\n|---|---|\n| 1 | {{ .B }}\n| 3 | 4 |', 3],
        ['a row with a cell too few', '| a | b |\n|---|---|\n| 1 |\n| 3 | 4 |', 3],
    ])('keeps %s a row', (_name, md, rows) => {
        const editor = load(md);
        const found = JSON.stringify(editor.getJSON()).split('"type":"tableRow"').length - 1;

        editor.destroy();

        expect(found).toBe(rows);
    });

    it('keeps the cells of a table in a task item of a mixed list, which is lexed a second time', () => {
        const saved = roundTrip('- normal\n- [ ] task\n  | a | b |\n  |---|---|\n  | `x | y` | {{ .A | print }} |');

        expect(saved).toContain('| `x \\| y` | {{ .A | print }} |');
        expect(roundTrip(saved)).toBe(saved);
    });
});

describe('the asterisks of an anonymizer placeholder are not emphasis', () => {
    it.each([
        ['emphasis around it', '*a §*x*§ b*', ['a §*x*§ b']],
        ['a backslash before it', '*a \\§*x*§ b*', ['a \\§*x*§ b']],
        ['more asterisks than the emphasis has', '*a §**ipv4***§ b*', ['a §**ipv4***§ b']],
        ['emphasis right after it', 'text §*a*§*b*', ['b']],
        ['a backslash before uneven asterisks', '*a \\§**ipv4***§ b*', ['a \\§**ipv4***§ b']],
        ['a closer left over from a second pass', 'at §*HTTP URL*§ Name Enhanced*§.net', []],
        ['none around it', 'see §*Name*§ now', []],
    ])('with %s', (_name, md, italics) => {
        const editor = load(md);
        const nodes: JSONContent[] = editor.getJSON().content?.[0]?.content ?? [];
        const marked = nodes
            .filter((node) => node.marks?.some((mark) => mark.type === 'italic'))
            .map((node) => node.text);

        editor.destroy();

        expect(marked).toEqual(italics);
    });
});

describe('a code span whose later line could be read as a block is written on one line', () => {
    it.each([
        ['a fence of three backticks', 'x ```a`b``c\nd``` tail\n\nnext', 'x ```a`b``c d``` tail\n\nnext'],
        ['a bullet in an item', '- a `foo\n      - bar` b', '- a `foo     - bar` b'],
        ['a numbered line', 'a `foo\n2. bar` b', 'a `foo 2. bar` b'],
        ['a line indented like code', 'a `foo\n     bar` b', 'a `foo      bar` b'],
    ])('%s', (_name, md, saved) => {
        expect(roundTrip(md)).toBe(saved);
        expect(roundTrip(saved)).toBe(saved);
    });
});

describe('a line break is not a character of the text', () => {
    it('is a break node, in emphasis too, so reading the paragraph back from the DOM keeps it', () => {
        const editor = load('one\n*two\nthree* `four\nfive`');
        const html = editor.getHTML();
        const texts = JSON.stringify(editor.getJSON()).match(/"text":"[^"]*"/g);

        editor.destroy();

        expect(texts).toEqual([
            '"text":"one"',
            '"text":"two"',
            '"text":"three"',
            '"text":" "',
            '"text":"four"',
            '"text":"five"',
        ]);
        expect(html).toBe(
            '<p>one<br data-break=""><em>two<br data-break="">three</em> <code>four<br data-break="">five</code></p>',
        );
    });
});

describe('a line break is written only between two lines of text', () => {
    const paragraph = (...content: JSONContent[]): JSONContent => ({
        content: [{ content, type: 'paragraph' }],
        type: 'doc',
    });
    const text = (value: string): JSONContent => ({ text: value, type: 'text' });
    const soft: JSONContent = { attrs: { marker: '' }, type: 'hardBreak' };
    const hard: JSONContent = { type: 'hardBreak' };

    it.each([
        [
            'Enter at the end of a line',
            'line one\nline two\nline three',
            (editor: Editor) => editor.chain().setTextSelection(9).splitBlock().run(),
            'line one\n\nline two\nline three',
        ],
        [
            'Enter at the start of a line',
            'line one\nline two\nline three',
            (editor: Editor) => editor.chain().setTextSelection(10).splitBlock().run(),
            'line one\n\nline two\nline three',
        ],
        [
            'Enter at the end of a line of an item',
            '- item one\n  item two\n- next',
            (editor: Editor) => editor.chain().setTextSelection(11).splitListItem('listItem').run(),
            '- item one\n- item two\n- next',
        ],
    ])('%s splits the text there and leaves no empty line', (_name, md, edit, saved) => {
        expect(afterEdit(md, edit)).toBe(saved);
        expect(roundTrip(saved)).toBe(saved);
    });

    it.each([
        ['two breaks in a row', paragraph(text('a'), soft, soft, text('b')), 'a\nb'],
        ['a break under a hard break', paragraph(text('a'), hard, soft, text('b')), 'a  \nb'],
        ['a break above a hard break', paragraph(text('a'), soft, hard, text('b')), 'a  \nb'],
        ['a break at the start', paragraph(soft, text('a')), 'a'],
        ['a break at the end', paragraph(text('a'), soft), 'a'],
        ['nothing but a break', paragraph(soft), ''],
    ])('%s is one line break at most', (_name, content, saved) => {
        expect(afterEdit(content)).toBe(saved);
    });

    it('is one break where the text of a middle line is deleted', () => {
        expect(
            afterEdit('line one\nline two\nline three', (editor) => editor.commands.deleteRange({ from: 10, to: 18 })),
        ).toBe('line one\nline three');
    });

    // marked merges a definition or an indented line into the text above it and leaves a line break behind.
    it.each([
        [
            'a definition above a row of dashes',
            'para\n[r]: u\n|---|---|',
            'para\n[r]: u\n\n|---|---|',
            ['para', '[r]: u', '|---|---|'],
        ],
        ['a definition between quoted lines', '> q\n[r]: u\n> q', '> q\n> [r]: u\n> q', ['q', '[r]: u', 'q']],
        [
            'an indented line above a rule',
            'Some text\n    indented\n---',
            'Some text\nindented\n\n---',
            ['Some text', 'indented'],
        ],
        [
            'a tab-indented line in an indented item',
            '   - item\n\tmore text',
            '- item\n  more text',
            ['item', 'more text'],
        ],
        [
            'a definition whose target is the row under it',
            'Columns\n[x]:\n|:--|--:|',
            'Columns\n[x]:\n|:--|--:|',
            ['Columns', '[x]:', '|:--|--:|'],
        ],
    ])('%s is read without an empty line', (_name, md, saved, texts) => {
        expect(linesOf(md).join('|/|')).toBe(texts.join('|/|'));
        expect(roundTrip(md)).toBe(saved);
        expect(roundTrip(saved)).toBe(saved);
    });
});

describe('a reference definition stays the text it was written as', () => {
    it.each([
        ['a references section', '## References\n\n[1]: https://example.com/advisory\n[2]: https://example.com/poc'],
        ['a cited reference', 'See [the advisory][1] for details.\n\n[1]: https://example.com/advisory'],
        ['a footnote-style definition', 'Text with a note[^1].\n\n[^1]: https://example.com/src'],
        ['a definition with a title', '[spec]: https://example.com/spec "The spec"\n\ntext'],
        ['a key and its value', '[target]: 10.0.0.5'],
        ['a definition as an item', '- [docs]: https://example.com/docs\n- next'],
        [
            'a second run of definitions',
            '[1]: https://a.b/one\n\ntext between\n\n[2]: https://a.b/two\n[3]: https://a.b/c',
        ],
        ['definitions right under a heading', '## References\n[1]: https://a.b/c\n[2]: https://a.b/d'],
        ['definitions a blank line apart', '[1]: https://a.b/one\n\n[2]: https://a.b/two'],
    ])('%s', (_name, md) => {
        expect(roundTrip(md)).toBe(md);
    });

    // marked merges the lines under a paragraph into it, and a run of definitions is not one it queued.
    it.each([
        ['a quoted definition above quoted text', '> [1]: http://x.test\n> text', ['[1]: http://x.test', 'text']],
        ['a definition above an indented line', '[1]: http://x.test\n    indented', ['[1]: http://x.test', 'indented']],
        [
            'a definition between two paragraphs',
            'Intro paragraph.\n\n[1]: http://x.test\n\n    indented',
            ['Intro paragraph.', '[1]: http://x.test', 'indented'],
        ],
    ])('%s loads with every text in place', (_name, md, texts) => {
        const editor = new Editor({ content: md, contentType: 'markdown', extensions: createMarkdownExtensions() });
        const loaded = JSON.stringify(editor.getJSON()).match(/"text":"[^"]*"/g);

        editor.destroy();

        expect(loaded).toEqual(texts.map((text) => `"text":"${text}"`));
    });

    it.each([
        ['a row of dashes', '[Summary]:\n---\n\ntext', '[Summary]:\n\\---\n\ntext', '---'],
        ['a redirect', '[x]:\n>out.txt', '[x]:\n\\>out.txt', '>out.txt'],
    ])(
        'reads a target on its own line that is %s without the backslash it is written with',
        (_name, md, saved, target) => {
            expect(roundTrip(md)).toBe(saved);
            expect(roundTrip(saved)).toBe(saved);
            expect(textsOf(saved)).toContain(`"text":"${target}"`);
        },
    );

    it('keeps the lines of a run apart as break nodes, like the lines of any paragraph', () => {
        const editor = load('[1]: https://a.b/one\n[2]: https://a.b/two\n[3]: https://a.b/three');
        const [run] = editor.getJSON().content ?? [];

        editor.destroy();

        expect(run?.content?.map((node: JSONContent) => node.text ?? node.attrs?.marker)).toEqual([
            '[1]: https://a.b/one',
            '',
            '[2]: https://a.b/two',
            '',
            '[3]: https://a.b/three',
        ]);
    });
});

describe('the cost of a paragraph grows with its text, not with the square of it', () => {
    const lines = (count: number, line: (at: number) => string): string =>
        Array.from({ length: count }, (_, at) => line(at)).join('\n');

    it.each<[string, 'read' | 'saved', (size: number) => string, number]>([
        ['lines that end in a hard break', 'saved', (size) => lines(size, (at) => `x${at}  `), 8000],
        ['emphasis after emphasis', 'saved', (size) => lines(size, () => '*em* **strong**'), 4000],
        ['lines of text', 'read', (size) => lines(size, (at) => `line ${at} of output with a *word*`), 8000],
        [
            'lines without a placeholder',
            'read',
            (size) => lines(size, (at) => `line ${at} open port on host  `),
            32_000,
        ],
        ['an action on every line', 'read', (size) => lines(size, (at) => `Host {{ .Host${at} }} is up  `), 32_000],
        [
            'a placeholder and emphasis on every line',
            'read',
            (size) => lines(size, (at) => `Host §*Host${at}*§ is *up*`),
            8000,
        ],
        [
            'one placeholder and an action on every word',
            'read',
            (size) => `§*Secret*§ ${lines(size, (at) => `w {{ .F${at} }}`)}`,
            64_000,
        ],
        [
            'a heading over text, block after block',
            'read',
            (size) => lines(size, (at) => `## Step ${at}\nRun the scan against host ${at}.`),
            4000,
        ],
        ['setext headings one after another', 'read', (size) => lines(size, (at) => `Title ${at}\n---`), 4000],
        [
            'headings underlined with equals signs, one after another',
            'read',
            (size) => lines(size, (at) => `Title ${at}\n===`),
            4000,
        ],
        [
            'tables with a blank line between them',
            'read',
            (size) => lines(size, () => '| a | b |\n|---|---|\n| 1 | 2 |\n'),
            4000,
        ],
    ])('%s, %s', (_name, side, build, size) => {
        const { destroy, parse, serialize } = markdownCodec();
        const growth =
            side === 'read'
                ? slowdownWhenInputQuadruples(build, parse, size)
                : slowdownWhenInputQuadruples((length) => parse(build(length)), serialize, size);

        destroy();

        expect(growth).toBeLessThan(GROWTH_IF_QUADRATIC / 2);
    });

    it('reads a paragraph of sixty thousand lines', () => {
        const { destroy, parse } = markdownCodec();
        const [paragraph] = parse('a\n'.repeat(60_000)).content ?? [];

        destroy();

        expect(paragraph?.content).toHaveLength(119_999);
    });
});
