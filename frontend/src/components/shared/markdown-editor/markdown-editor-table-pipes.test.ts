import { Editor } from '@tiptap/core';
import { describe, expect, it } from 'vitest';

import { createMarkdownExtensions } from './markdown-editor-extensions';
import { escapeTablePipes } from './markdown-editor-table-pipes';
import { roundTrip, setupEditorJsdom } from './markdown-editor-test-setup';

setupEditorJsdom();

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

describe('table cell with a piped code span — content survives load and converges (H1)', () => {
    it('keeps the trailing cell and preserves the code content', () => {
        const out = roundTrip('| Op | Meaning |\n| --- | --- |\n| `x | y` | z |');

        expect(out).toContain('z');
        expect(out).toContain('`x \\| y`');
        expect(roundTrip(out)).toBe(out);
    });

    // This used to pin `{{.X \| upper}}` — the escaped form, which Go text/template rejects outright
    // (`unexpected "\" in operand`), so the prompt could not be saved at all. The pipe is the template's
    // pipeline operator and stays raw; the cell count survives because the loader masks action pipes.
    it('keeps a Go-template action with a pipe inside a cell', () => {
        const out = roundTrip('| Var | Out |\n| --- | --- |\n| {{.X | upper}} | done |');

        expect(out).toContain('done');
        expect(out).toContain('{{.X | upper}}');
        expect(out).not.toContain('\\|');
        expect(roundTrip(out)).toBe(out);
    });

    it('keeps a code span with a backtick and a pipe inside a cell', () => {
        const out = roundTrip('| Op | Meaning |\n| --- | --- |\n| `` `x` | y `` | kept |');

        expect(out).toContain('kept');
        expect(out).toContain('`` `x` \\| y ``');
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

        expect(out).toContain('done');
        expect(out).toContain('`git log \\| head`');
        expect(roundTrip(out)).toBe(out);
    });

    it('protects a template action in a no-leading-pipe body row', () => {
        const out = roundTrip('Var | Out\n--- | ---\n{{.Host | lower}} | done');

        expect(out).toContain('done');
        expect(out).toContain('{{.Host | lower}}');
    });

    it('protects rows whether or not each has a leading pipe (mixed)', () => {
        const out = roundTrip('| A | B |\n| --- | --- |\n| `p | q` | one |\n`r | s` | two');

        expect(out).toContain('one');
        expect(out).toContain('two');
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

    it('scans a table row with a long non-URL token in linear time (URL escaping ReDoS guard)', () => {
        const evil = `| a | b |\n| --- | --- |\n| ${'a'.repeat(120000)} | z |`;
        const started = performance.now();

        escapeTablePipes(evil);

        expect(performance.now() - started).toBeLessThan(100);
    });

    it('keeps the URL and the trailing cell on round-trip, and converges', () => {
        const out = roundTrip('| A | B |\n| --- | --- |\n| [go](https://h/?x=1|2) | TRAILING |');

        expect(out).toContain('TRAILING');
        expect(out).toContain('x=1\\|2');
        expect(roundTrip(out)).toBe(out);
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
        const out = roundTrip('> | a | b |\n> | --- | --- |\n> | `x | y` | z |');

        expect(out).toContain('z');
        expect(out).toContain('`x \\| y`');
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
        const out = roundTrip('````\n```\ninner\n````\n\n| a | b |\n| --- | --- |\n| `x | y` | z |');

        expect(out).toContain('z');
        expect(out).toContain('`x \\| y`');
    });
});

describe('CRLF line endings — tables still protected', () => {
    it('escapes a code-span pipe in a CRLF table row', () => {
        expect(escapeTablePipes('| a | b |\r\n| --- | --- |\r\n| `x | y` | z |\r\n')).toBe(
            '| a | b |\n| --- | --- |\n| `x \\| y` | z |\n',
        );
    });

    it('keeps the trailing cell of a CRLF table on round-trip', () => {
        const out = roundTrip('| a | b |\r\n| --- | --- |\r\n| `x | y` | z |\r\n');

        expect(out).toContain('z');
        expect(out).toContain('`x \\| y`');
    });

    it('leaves CRLF bytes untouched when there is no table to escape', () => {
        const doc = 'line one\r\nline `a | b` two\r\n';

        expect(escapeTablePipes(doc)).toBe(doc);
    });
});

describe('TABLE_DELIMITER_LINE is linear (ReDoS guard)', () => {
    it('scans a crafted delimiter-looking line with a long trailing space run in linear time', () => {
        const evil = `x|y\n${'-'.repeat(50)}${' '.repeat(60000)}z\n`;
        const started = performance.now();

        escapeTablePipes(evil);

        expect(performance.now() - started).toBeLessThan(100);
    });
});

describe('a backtick in a backtick fence info string is not a fence opener', () => {
    const cellsOf = (markdown: string) => {
        const editor = new Editor({
            content: markdown,
            contentType: 'markdown',
            extensions: createMarkdownExtensions(),
        });
        const cells: string[] = [];

        editor.state.doc.descendants((node) => {
            if (node.type.name === 'tableCell' || node.type.name === 'tableHeader') {
                cells.push(node.textContent);
            }

            return true;
        });
        editor.destroy();

        return cells;
    };

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

    // The opposite direction of the same desynchronisation: once the phantom fence closes, the scanner's parity
    // is inverted and it escapes pipes INSIDE a real code block, injecting a backslash into code content.
    it('leaves a real code block byte-identical, injecting no escape into its content', () => {
        const source = ['```a`b', '```', '', TABLE].join('\n');
        const editor = new Editor({ content: source, contentType: 'markdown', extensions: createMarkdownExtensions() });
        const codeBlocks: string[] = [];

        editor.state.doc.descendants((node) => {
            if (node.type.name === 'codeBlock') {
                codeBlocks.push(node.textContent);
            }

            return true;
        });
        editor.destroy();

        expect(codeBlocks.join('')).not.toContain('\\|');
    });

    it('still treats a genuine fence as a fence', () => {
        expect(escapeTablePipes(['```js', '| a | b |', '```'].join('\n'))).toBe(
            ['```js', '| a | b |', '```'].join('\n'),
        );
    });
});

describe('tables nested inside list items keep their pipe protection', () => {
    const cellsOf = (markdown: string) => {
        const editor = new Editor({
            content: markdown,
            contentType: 'markdown',
            extensions: createMarkdownExtensions(),
        });
        const cells: string[] = [];

        editor.state.doc.descendants((node) => {
            if (node.type.name === 'tableCell' || node.type.name === 'tableHeader') {
                cells.push(node.textContent);
            }

            return true;
        });
        editor.destroy();

        return cells;
    };

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
    const cellsOf = (markdown: string) => {
        const editor = new Editor({
            content: markdown,
            contentType: 'markdown',
            extensions: createMarkdownExtensions(),
        });
        const cells: string[] = [];

        editor.state.doc.descendants((node) => {
            if (node.type.name === 'tableCell' || node.type.name === 'tableHeader') {
                cells.push(node.textContent);
            }

            return true;
        });
        editor.destroy();

        return cells;
    };

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

    // A structural pipe still has to be escaped when it is real content, not a template operator.
    it('still escapes a pipe inside a code span', () => {
        expect(roundTrip('| a | b |\n| --- | --- |\n| `x | y` | z |')).toContain('\\|');
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

    // GFM has no headerless table, so the serializer used to emit an EMPTY header row above the demoted rows:
    // every header-off + save + reload cycle grew the table by one blank row (2 → 3 → 4) and the switch
    // silently flipped back on. Promoting the first row keeps the row count and every cell.
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

    // The table renderer emits a newline of its own on top of the block separator, so between two adjacent
    // tables that extra line reloaded as an empty paragraph — which serialised to another blank line on the
    // next save. Measured before the fix: 68 → 88 → 90 → 92 → 94 bytes, one paragraph per cycle, no fixed point.
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
