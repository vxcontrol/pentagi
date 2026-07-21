import { describe, expect, it } from 'vitest';

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

    it('keeps a Go-template action with a pipe inside a cell', () => {
        const out = roundTrip('| Var | Out |\n| --- | --- |\n| {{.X | upper}} | done |');

        expect(out).toContain('done');
        expect(out).toContain('{{.X \\| upper}}');
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
        expect(out).toContain('{{.Host \\| lower}}');
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
