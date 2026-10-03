import { Editor } from '@tiptap/core';
import { describe, expect, it } from 'vitest';

import { GROWTH_IF_QUADRATIC, slowdownWhenInputQuadruples } from '@/test-utils/cost-growth';

import { createMarkdownExtensions } from './markdown-editor-extensions';
import { escapeTablePipes } from './markdown-editor-table-pipes';
import { cellsOf, roundTrip, setupEditorJsdom, structuralCounts } from './markdown-editor-test-setup';

setupEditorJsdom();

const plainRow = (tokenLength: number) => `| a | b |\n| --- | --- |\n| ${'a'.repeat(tokenLength)} | z |`;
const phantomCellRow = (tokenLength: number) => `| a | b |\n| --- | --- |\n| ${'a'.repeat(tokenLength)} | y | z |`;
const delimiterLikeLine = (spaceRun: number) => `x|y\n${'-'.repeat(50)}${' '.repeat(spaceRun)}z\n`;

describe('escapeTablePipes — pure pre-lex pipe protection', () => {
    it('escapes a pipe inside a code span in a body row', () => {
        expect(escapeTablePipes('| Op | Meaning |\n| --- | --- |\n| `x | y` | z |')).toBe(
            '| Op | Meaning |\n| --- | --- |\n| `x \\| y` | z |',
        );
    });

    it('escapes a pipe inside a Go-template action in a body row', () => {
        expect(escapeTablePipes('| Var | Out |\n| --- | --- |\n| {{.X | upper}} | ok |')).toBe(
            '| Var | Out |\n| --- | --- |\n| {{.X \\| upper}} | ok |',
        );
    });

    it('leaves a plain-text pipe (real column delimiter) untouched', () => {
        const table = '| a | b |\n| --- | --- |\n| 1 | 2 |';

        expect(escapeTablePipes(table)).toBe(table);
    });

    it('does not double-escape an already-escaped pipe', () => {
        const table = '| a | b |\n| --- | --- |\n| `x \\| y` | z |';

        expect(escapeTablePipes(table)).toBe(table);
    });

    it('handles a multi-backtick code span', () => {
        expect(escapeTablePipes('| a | b |\n| --- | --- |\n| ``x | y`` | z |')).toBe(
            '| a | b |\n| --- | --- |\n| ``x \\| y`` | z |',
        );
    });

    it('leaves an unclosed backtick run alone (not a code span)', () => {
        const table = '| a | b |\n| --- | --- |\n| `x | y |';

        expect(escapeTablePipes(table)).toBe(table);
    });

    it('never touches a fenced code block that happens to hold a table', () => {
        const doc = '```\n| a | b |\n| --- | --- |\n| `x | y` | z |\n```';

        expect(escapeTablePipes(doc)).toBe(doc);
    });

    it('ignores a non-table line whose text contains pipes in code', () => {
        const prose = 'run `a | b` in the shell';

        expect(escapeTablePipes(prose)).toBe(prose);
    });

    it('returns the input unchanged when there is no pipe at all', () => {
        const doc = '# heading\n\nsome `code` here';

        expect(escapeTablePipes(doc)).toBe(doc);
    });
});

describe('table cell with a piped code span — content survives load and converges', () => {
    it('keeps a Go-template action with a pipe inside a cell', () => {
        const out = roundTrip('| Var | Out |\n| --- | --- |\n| {{.X | upper}} | done |');

        expect(cellsOf(out)).toEqual(['Var', 'Out', '{{.X | upper}}', 'done']);
        expect(out).toContain('{{.X | upper}}');
        expect(out).not.toContain('\\|');
        expect(roundTrip(out)).toBe(out);
    });

    it('keeps a code span with a backtick and a pipe inside a cell', () => {
        const out = roundTrip('| Op | Meaning |\n| --- | --- |\n| `` `x` | y `` | kept |');

        expect(cellsOf(out)).toEqual(['Op', 'Meaning', '`x` | y', 'kept']);
        expect(out).toContain('`` `x` | y ``');
        expect(roundTrip(out)).toBe(out);
    });
});

describe('pipe-less GFM tables (no outer pipe) — cells survive too', () => {
    it('protects a body row without a leading pipe', () => {
        expect(escapeTablePipes('A | B | C\n--- | --- | ---\n`git log | head` | notes | done')).toBe(
            'A | B | C\n--- | --- | ---\n`git log \\| head` | notes | done',
        );
    });

    it('keeps the trailing cell of a no-leading-pipe table on round-trip', () => {
        const out = roundTrip('A | B | C\n--- | --- | ---\n`git log | head` | notes | done');

        expect(cellsOf(out)).toEqual(['A', 'B', 'C', 'git log | head', 'notes', 'done']);
        expect(out).toContain('`git log | head`');
        expect(roundTrip(out)).toBe(out);
    });

    it('protects a template action in a no-leading-pipe body row', () => {
        const out = roundTrip('Var | Out\n--- | ---\n{{.Host | lower}} | done');

        expect(cellsOf(out)).toEqual(['Var', 'Out', '{{.Host | lower}}', 'done']);
        expect(out).toContain('{{.Host | lower}}');
    });

    it('protects rows whether or not each has a leading pipe (mixed)', () => {
        const out = roundTrip('| A | B |\n| --- | --- |\n| `p | q` | one |\n`r | s` | two');

        expect(cellsOf(out)).toEqual(['A', 'B', 'p | q', 'one', 'r | s', 'two']);
    });

    it('stops at a block boundary — a heading after the table is not escaped', () => {
        const src = 'A | B\n--- | ---\nr1 | r2\n# next `x | y` heading';

        expect(escapeTablePipes(src)).toBe(src);
    });
});

describe('table cell with a pipe inside a URL — content survives load and converges', () => {
    it('escapes a pipe inside a link destination', () => {
        expect(escapeTablePipes('| A | B |\n| --- | --- |\n| [x](https://h/?a=1|2) | end |')).toBe(
            '| A | B |\n| --- | --- |\n| [x](https://h/?a=1\\|2) | end |',
        );
    });

    it('escapes a pipe inside a bare autolink and an image src', () => {
        expect(escapeTablePipes('| A | B |\n| --- | --- |\n| https://h/?a=1|2 | ![p](https://c/i.png?w=1|2) |')).toBe(
            '| A | B |\n| --- | --- |\n| https://h/?a=1\\|2 | ![p](https://c/i.png?w=1\\|2) |',
        );
    });

    it('leaves a real column delimiter (spaced pipe, no scheme run) untouched', () => {
        const table = '| A | B |\n| --- | --- |\n| http://h/x | plain |';

        expect(escapeTablePipes(table)).toBe(table);
    });

    it('leaves structural pipes in a compact (spaceless) URL row untouched', () => {
        const table = '| url | desc |\n| --- | --- |\n|http://a.com|b|';

        expect(escapeTablePipes(table)).toBe(table);
    });

    it('scans a long non-URL token linearly on a row the cell-count guard skips', () => {
        const row = plainRow(120_000);

        expect(escapeTablePipes(row)).toBe(row);
        expect(slowdownWhenInputQuadruples(plainRow, escapeTablePipes, 120_000)).toBeLessThan(GROWTH_IF_QUADRATIC / 2);
    });

    it('scans a long non-URL token linearly on a row whose extra cell drives the URL pass', () => {
        const row = phantomCellRow(120_000);

        expect(escapeTablePipes(row)).toBe(row);
        expect(slowdownWhenInputQuadruples(phantomCellRow, escapeTablePipes, 120_000)).toBeLessThan(
            GROWTH_IF_QUADRATIC / 2,
        );
    });

    it('keeps the URL and the trailing cell on round-trip', () => {
        const md = '| A | B |\n| --- | --- |\n| [go](https://h/?x=1|2) | TRAILING |';

        expect(structuralCounts(md)).toMatchObject({ tableCell: 2 });
        expect(roundTrip(md)).toBe(md);
    });
});

describe('tables inside a blockquote — prefix-stripped and protected', () => {
    it('escapes a code-span pipe in a blockquoted table row', () => {
        expect(escapeTablePipes('> | a | b |\n> | --- | --- |\n> | `x | y` | z |')).toBe(
            '> | a | b |\n> | --- | --- |\n> | `x \\| y` | z |',
        );
    });

    it('handles a nested blockquote table', () => {
        expect(escapeTablePipes('> > | a | b |\n> > | --- | --- |\n> > | `x | y` | z |')).toBe(
            '> > | a | b |\n> > | --- | --- |\n> > | `x \\| y` | z |',
        );
    });

    it('keeps the trailing cell of a blockquoted table on round-trip', () => {
        const src = '> | a | b |\n> | --- | --- |\n> | `x | y` | z |';

        expect(roundTrip(src)).toBe(src);
        expect(cellsOf(src)).toEqual(['a', 'b', 'x | y', 'z']);
    });

    it('leaves blockquote prose (no table) untouched', () => {
        const doc = '> a quote with `a | b` inline code\n> and more text';

        expect(escapeTablePipes(doc)).toBe(doc);
    });
});

describe('fence length tracking — a longer fence is not closed by a shorter inner run', () => {
    it('keeps protecting a table after a 4-backtick block that contains a ``` line', () => {
        const src = '````\n```\ninner\n````\n\n| a | b |\n| --- | --- |\n| `x | y` | z |';

        expect(escapeTablePipes(src)).toBe('````\n```\ninner\n````\n\n| a | b |\n| --- | --- |\n| `x \\| y` | z |');
    });

    it('does not escape a table sitting inside a 4-backtick block that also holds a ``` line', () => {
        const src = '````\n```\n| a | b |\n| --- | --- |\n| `x | y` | z |\n````';

        expect(escapeTablePipes(src)).toBe(src);
    });

    it('round-trips a table after a fence-demonstrating code block without losing the cell', () => {
        const src = '````\n```\ninner\n````\n\n| a | b |\n| --- | --- |\n| `x | y` | z |';

        expect(roundTrip(src)).toBe(src);
        expect(cellsOf(src)).toEqual(['a', 'b', 'x | y', 'z']);
    });
});

describe('CRLF line endings — tables still protected', () => {
    it('escapes a code-span pipe in a CRLF table row', () => {
        expect(escapeTablePipes('| a | b |\r\n| --- | --- |\r\n| `x | y` | z |\r\n')).toBe(
            '| a | b |\n| --- | --- |\n| `x \\| y` | z |\n',
        );
    });

    it('keeps the trailing cell of a CRLF table on round-trip', () => {
        const src = '| a | b |\r\n| --- | --- |\r\n| `x | y` | z |\r\n';

        expect(roundTrip(src)).toBe('| a | b |\n| --- | --- |\n| `x | y` | z |\n');
        expect(cellsOf(src)).toEqual(['a', 'b', 'x | y', 'z']);
    });

    it('leaves CRLF bytes untouched when there is no table to escape', () => {
        const doc = 'line one\r\nline `a | b` two\r\n';

        expect(escapeTablePipes(doc)).toBe(doc);
    });
});

describe('TABLE_DELIMITER_LINE is linear (ReDoS guard)', () => {
    it('scans a crafted delimiter-looking line with a long trailing space run in linear time', () => {
        const line = delimiterLikeLine(120_000);

        expect(escapeTablePipes(line)).toBe(line);
        expect(slowdownWhenInputQuadruples(delimiterLikeLine, escapeTablePipes, 120_000)).toBeLessThan(
            GROWTH_IF_QUADRATIC / 2,
        );
    });
});

describe('a backtick in a backtick fence info string is not a fence opener', () => {
    const TABLE = ['| Op | Meaning |', '| --- | --- |', '| `x | y` | KEEP |'].join('\n');

    // marked's fence rule is /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})…/ — the no-backtick lookahead applies to
    // the BACKTICK branch only. A prose line holding an inline triple-backtick span is not a fence for marked,
    // so treating it as one desynchronises the scanner from the parser.
    it('keeps pipe protection for a table that follows an inline triple-backtick span', () => {
        expect(cellsOf(['```pnpm run dev``` starts it', '', TABLE].join('\n'))).toEqual([
            'Op',
            'Meaning',
            'x | y',
            'KEEP',
        ]);
    });

    it('protects a table that follows no fence at all, unchanged', () => {
        expect(cellsOf(TABLE)).toEqual(['Op', 'Meaning', 'x | y', 'KEEP']);
    });

    // The opposite direction: once the phantom fence closes, the scan takes the lines of a real code block for a table.
    it('leaves a real code block byte-identical, injecting no escape into its content', () => {
        const source = ['```a`b', '```', '', TABLE].join('\n');

        expect(escapeTablePipes(source)).toBe(source);
    });

    it('still treats a genuine fence as a fence', () => {
        expect(escapeTablePipes(['```js', '| a | b |', '```'].join('\n'))).toBe(
            ['```js', '| a | b |', '```'].join('\n'),
        );
    });
});

describe('tables nested inside list items keep their pipe protection', () => {
    const table = (indent: string) =>
        [`${indent}| Op | Meaning |`, `${indent}| --- | --- |`, `${indent}| \`x | y\` | KEEP |`].join('\n');

    const EXPECTED = ['Op', 'Meaning', 'x | y', 'KEEP'];

    it('protects a table under a bullet at its natural content column', () => {
        expect(cellsOf(['- item', '', table('  ')].join('\n'))).toEqual(EXPECTED);
    });

    // The ordinary way anyone writes a table under a sub-bullet: the content column is 4, which the top-level
    // scanner reads as indented code and skips.
    it('protects a table under a nested bullet', () => {
        expect(cellsOf(['- outer', '  - inner', '', table('    ')].join('\n'))).toEqual(EXPECTED);
    });

    it('protects a table under an ordered item', () => {
        expect(cellsOf(['1. item', '', table('   ')].join('\n'))).toEqual(EXPECTED);
    });

    it('protects a table under a bullet inside a blockquote', () => {
        expect(cellsOf(['> - item', '>', `> ${table('  ').split('\n').join('\n> ')}`].join('\n'))).toEqual(EXPECTED);
    });

    // Relative to the item's content column, four more spaces is still indented code — marked does not make a
    // table there, so escaping it would write a backslash into code content.
    it('leaves a table indented four columns past the item content alone', () => {
        const source = ['- item', '', table('      ')].join('\n');

        expect(escapeTablePipes(source)).toBe(source);
    });

    // The same rule at top level, which is what a blanket relaxation of the leading-space cap would break.
    it('leaves a four-space-indented table at top level alone', () => {
        const source = table('    ');

        expect(escapeTablePipes(source)).toBe(source);
    });

    it('does not swallow the next item at the same level', () => {
        const source = ['- first', '', table('  '), '', '- second | not a table'].join('\n');

        expect(escapeTablePipes(source)).toContain('- second | not a table');
    });
});

describe('a Go template pipeline in a table cell keeps its own pipe', () => {
    // `{{.Host | urlquery}}` is a Go text/template pipeline. Escaping its pipe makes text/template reject the
    // whole file with `unexpected "\" in operand`, so the prompt cannot be saved at all.
    it('does not escape a pipeline in a body cell on save', () => {
        expect(roundTrip('| host | value |\n| --- | --- |\n| a | {{.Host | urlquery}} |')).not.toContain('\\|');
    });

    it('keeps every cell of a table whose body holds a pipeline', () => {
        expect(cellsOf('| host | value |\n| --- | --- |\n| a | {{.Host | urlquery}} |')).toEqual([
            'host',
            'value',
            'a',
            '{{.Host | urlquery}}',
        ]);
    });

    // The header row is where dropping the escape without an action-aware cell count destroys the table: the
    // raw pipe makes the header count 3 against the delimiter's 2, detection bails, and marked degrades the
    // whole table to a paragraph.
    it('keeps every cell of a table whose HEADER holds a pipeline', () => {
        expect(cellsOf('| {{.A | urlquery}} | note |\n| --- | --- |\n| x | y |')).toEqual([
            '{{.A | urlquery}}',
            'note',
            'x',
            'y',
        ]);
    });

    it('round-trips a pipeline-bearing table byte-identically on the second save', () => {
        const source = '| host | value |\n| --- | --- |\n| a | {{.Host | urlquery}} |';
        const once = roundTrip(source);

        expect(roundTrip(once)).toBe(once);
    });
});
describe('the scan takes for a row only what marked takes for one', () => {
    it.each([
        ['a standard HTML block tag under the table', '| a |\n|---|\n<p>`x | y`</p>'],
        ['a closing block tag under the table', '| a |\n|---|\n| 1 |\n</details>\n`x | y`'],
        [
            'a template control line under the table',
            '| a | b |\n|---|---|\n{{ range .Rows }}| `x | y` | z |\n{{ end }}',
        ],
    ])('leaves the pipes of %s alone', (_name, md) => {
        expect(escapeTablePipes(md)).toBe(md);
    });

    it.each([
        ['a tag that is no HTML block', '| a |\n|---|\n<tool>`x | y`</tool>', '| a |\n|---|\n<tool>`x \\| y`</tool>'],
        ['an action that is no control line', '| a |\n|---|\n{{ .X }} `x | y`', '| a |\n|---|\n{{ .X }} `x \\| y`'],
        [
            'an action with a brace in a string',
            '| a |\n|---|\n| {{ printf "{%s}" .X | print }} |',
            '| a |\n|---|\n| {{ printf "{%s}" .X \\| print }} |',
        ],
    ])('still protects the row of %s', (_name, md, escaped) => {
        expect(escapeTablePipes(md)).toBe(escaped);
    });

    it.each([
        [
            'a heading over a row of dashes',
            '## T | N\n|---|---|\n| `a | b` | {{ .A | print }} |',
            '| `a | b` | {{ .A | print }} |',
        ],
        ['a line indented like code over a row of dashes', '    a | b\n|---|---|\n| `x | y` | c |', '| `x | y` | c |'],
    ])('leaves no trace in %s, which is no table', (_name, md, row) => {
        const saved = roundTrip(md);

        expect(structuralCounts(md).table).toBeUndefined();
        expect(saved.split('\n').at(-1)).toBe(row);
        expect(roundTrip(saved)).toBe(saved);
    });

    it.each([
        ['a row with a cell too many', '| a | b |\n|---|---|\n| 1 | 2 | `x | y` |'],
        ['a line of nothing but actions', '| a | b |\n|---|---|\n{{ $n := len .Rows }} {{/* `x | y` */}}'],
        ['the rows under the line that ended the table', '| a | b |\n|---|---|\n{{ range .R }}\n| `x | y` | 2 |'],
    ])('leaves the pipes of %s alone', (_name, md) => {
        expect(escapeTablePipes(md)).toBe(md);
    });

    it('finds the table under an item that opens its fence on the marker line', () => {
        const list = '1. Run:\n   ```sh\n   nmap\n   ```\n2. ```sh\n   curl\n   ```\n\n';

        expect(escapeTablePipes(`${list}| a | b |\n|---|---|\n| \`x | y\` | 2 |`)).toBe(
            `${list}| a | b |\n|---|---|\n| \`x \\| y\` | 2 |`,
        );
    });

    it('takes a string that holds the end of an action for part of the action', () => {
        expect(escapeTablePipes('| a | b |\n|---|---|\n| {{ print "}}" | print }} | 2 |')).toBe(
            '| a | b |\n|---|---|\n| {{ print "}}" \\| print }} | 2 |',
        );
    });

    it('finds actions that hold strings in time that grows with the row, not with the square of it', () => {
        const row = (size: number) => `| a |\n|---|\n| ${'{{ "x '.repeat(size / 7)} |`;

        expect(slowdownWhenInputQuadruples(row, escapeTablePipes, 140_000)).toBeLessThan(GROWTH_IF_QUADRATIC / 2);
    });

    it('keeps a row that opens with an action whose name only begins with a control word', () => {
        const md = '| a | b |\n|---|---|\n{{ endpoint }} | 2\n{{ iffy .X }} | 3';

        expect(structuralCounts(md).tableRow).toBe(3);
        expect(escapeTablePipes(`${md} \`x | y\``)).toBe(`${md} \`x \\| y\``);
    });

    it('protects with a backslash in a document that holds the stand-in itself', () => {
        const saved = roundTrip('x \uE002 y\n\n| a | b |\n|---|---|\n| `p | q` | 2 |');

        expect(saved).toBe('x \uE002 y\n\n| a | b |\n|---|---|\n| `p \\| q` | 2 |');
        expect(roundTrip(saved)).toBe(saved);
    });

    it('writes the pipes it protects as the stand-in it is asked for', () => {
        expect(escapeTablePipes('| a | b |\n|---|---|\n| `x | y` | {{ .A | print }} |', '\uE002')).toBe(
            '| a | b |\n|---|---|\n| `x \uE002 y` | {{ .A \uE002 print }} |',
        );
    });

    it('finds actions in time that grows with the row, not with the square of it', () => {
        const row = (size: number) => `| a |\n|---|\n| ${'{{ x { '.repeat(size / 7)} |`;

        expect(slowdownWhenInputQuadruples(row, escapeTablePipes, 140_000)).toBeLessThan(GROWTH_IF_QUADRATIC / 2);
    });
});

describe('a table with no header row keeps its row count across a save', () => {
    const rowsOf = (markdown: string) => {
        const editor = new Editor({
            content: markdown,
            contentType: 'markdown',
            extensions: createMarkdownExtensions(),
        });
        let rows = 0;

        editor.state.doc.descendants((node) => {
            if (node.type.name === 'tableRow') {
                rows += 1;
            }

            return true;
        });

        const saved = editor.getMarkdown();

        editor.destroy();

        return { rows, saved };
    };

    it('does not grow when the header row is toggled off', () => {
        const editor = new Editor({
            content: ['| a | b |', '| --- | --- |', '| 1 | 2 |'].join('\n'),
            contentType: 'markdown',
            extensions: createMarkdownExtensions(),
        });

        editor.commands.setTextSelection(3);
        editor.commands.toggleHeaderRow();

        const saved = editor.getMarkdown();

        editor.destroy();

        const first = rowsOf(saved);

        expect(first.rows).toBe(2);
        expect(saved).toContain('a');
        expect(saved).toContain('1');
        expect(rowsOf(first.saved).rows).toBe(2);
    });
});

describe('adjacent tables do not accumulate blank paragraphs', () => {
    const kindsOf = (markdown: string) => {
        const editor = new Editor({
            content: markdown,
            contentType: 'markdown',
            extensions: createMarkdownExtensions(),
        });
        const kinds: string[] = [];

        editor.state.doc.forEach((node) => kinds.push(node.type.name));
        editor.destroy();

        return kinds.join(',');
    };

    it('keeps two tables adjacent across repeated saves', () => {
        const source = ['| a | b |', '| --- | --- |', '| 1 | 2 |', '', '| c | d |', '| --- | --- |', '| 3 | 4 |'].join(
            '\n',
        );
        const once = roundTrip(source);

        expect(kindsOf(source)).toBe('table,table');
        expect(kindsOf(once)).toBe('table,table');
        expect(roundTrip(once)).toBe(once);
    });
});
