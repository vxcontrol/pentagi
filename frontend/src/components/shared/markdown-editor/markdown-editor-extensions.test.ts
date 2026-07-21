import { Editor } from '@tiptap/core';
import { beforeAll, describe, expect, it } from 'vitest';

import { createMarkdownExtensions } from './markdown-editor-extensions';
import { roundTrip, setupEditorJsdom, structuralCounts } from './markdown-editor-test-setup';
import { findVariableOccurrences } from './markdown-editor-variable-highlight';

beforeAll(setupEditorJsdom);

const words = (s: string): string[] => s.match(/[\p{L}\p{N}]+/gu) ?? [];
const sameWords = (a: string, b: string) => expect(words(b).sort()).toEqual(words(a).sort());

describe('literal <tags> survive (marked HTML-tokenizers neutralized + no entity-encode)', () => {
    it.each([
        '<container_environment>',
        '</language_policy>',
        '<specialist name="searcher">',
        // void HTML elements: marked would SWALLOW these without the neutralized html/tag tokenizers.
        '<input>',
        '<br>',
    ])('keeps %s verbatim (no &lt;, no swallow)', (tag) => {
        const out = roundTrip('lead ' + tag + ' tail');

        expect(out).toContain(tag);
        expect(out).not.toContain('&lt;');
        expect(out).not.toContain('&gt;');
    });
});

describe('Go-template variables survive verbatim', () => {
    it.each(['{{.Var}}', '{{- if .X}}', '{{range .Items}}', '{{.X | upper}}', '{{printf "%d" .N}}', '{{- .Both -}}'])(
        'keeps %s',
        (v) => {
            expect(roundTrip('lead ' + v + ' tail')).toContain(v);
        },
    );
});

describe('selective escape — no stray backslashes on literal punctuation', () => {
    it.each(['Scan ports [1-1000].', 'run nmap *.php here', 'a snake_case_name and array[i] index'])(
        'leaves %s unescaped',
        (s) => {
            const out = roundTrip(s);

            expect(out).not.toContain('\\[');
            expect(out).not.toContain('\\*');
            expect(out).not.toContain('\\_');
            sameWords(s, out);
        },
    );
});

describe('single tildes stay literal — no over-escape, no accidental strikethrough', () => {
    it.each([
        '~10% for environment setup',
        'approximately ~5 minutes',
        'a ~ b ~ c',
        'from ~5~ to ~10~',
        'price ~5~ dollars',
        'path a~b~c here',
    ])('keeps %s byte-identical', (s) => {
        expect(roundTrip(s)).toBe(s);
    });

    it('still round-trips ~~strikethrough~~ (double tilde is GFM strike)', () => {
        expect(roundTrip('use ~~deprecated~~ here')).toContain('~~deprecated~~');
    });
});

describe('underscore delimiters stay literal — dunders/identifiers survive, * emphasis still works', () => {
    it.each(['Override __init__ then __call__', 'a _emphasis_ word', 'snake_case_name plus __dunder__'])(
        'keeps %s byte-identical',
        (s) => {
            expect(roundTrip(s)).toBe(s);
        },
    );

    it('still renders *star* / **star** emphasis (the toolbar delimiter)', () => {
        const out = roundTrip('use *italic* and **bold** here');
        expect(out).toContain('*italic*');
        expect(out).toContain('**bold**');
    });
});

describe('literal backslashes survive — no doubling (regex classes, Windows paths, escape sequences)', () => {
    it.each([
        'match \\d+ digits, \\w+ words, \\s whitespace, \\b boundary',
        'path C:\\Users\\bob\\report.txt',
        'escape \\n \\t \\r sequences and a\\Zb',
    ])('keeps %s byte-identical', (s) => {
        expect(roundTrip(s)).toBe(s);
    });
});

describe('backslash before punctuation survives — escape tokenizer neutralized (regex/glob/UNC content)', () => {
    it.each([
        'regex \\d+ and \\.php extension',
        'glob \\* and \\? wildcards',
        'escaped \\[ \\] \\( \\) \\{ \\} brackets',
        'dunder path C:\\Users\\_admin\\app',
        'table pipe \\| and plus \\+ and dot \\.',
        'unc \\\\server\\share and double \\\\ backslash',
    ])('keeps %s byte-identical', (s) => {
        expect(roundTrip(s)).toBe(s);
    });

    it('a line-leading \\* stays a paragraph, never a bullet list', () => {
        expect(roundTrip('\\* not a bullet, literal star')).toBe('\\* not a bullet, literal star');
    });
});

describe('named HTML entities decode outside code; numeric refs, bare & and raw <tags> survive', () => {
    it.each([
        ['encode &lt;script&gt; as text', 'encode <script> as text'],
        ['ampersand AT&amp;T and a &amp; b', 'ampersand AT&T and a & b'],
        ['quote &quot;value&quot; and &gt; alone and &lt; too', 'quote "value" and > alone and < too'],
    ])('decodes %s to its characters', (src, expected) => {
        expect(roundTrip(src)).toBe(expected);
    });

    it.each(['numeric &#123; and &#40; and &#x25; stay', 'bare AT&T and a & b and 2>&1 survive'])(
        'keeps %s byte-identical',
        (s) => {
            expect(roundTrip(s)).toBe(s);
        },
    );

    it('entities inside inline code are preserved (code content is not decoded)', () => {
        expect(roundTrip('`&lt;script&gt;`')).toBe('`&lt;script&gt;`');
    });

    it('literal <tags> (raw, no entity) still stay literal, never entity-encoded', () => {
        expect(roundTrip('use <input> and <container_environment> here')).toBe(
            'use <input> and <container_environment> here',
        );
    });

    // INTENTIONAL (product decision): a multi-encoded entity in prose loses ONE encoding level per open+save
    // cycle until fully decoded — outside code, `&lt;` means `<`. Do NOT "fix" by re-encoding `&` on save:
    // that would freeze the single-level ingestion artifacts the decode exists to clean.
    it('multi-encoded prose entity decodes one level per cycle, then stabilizes', () => {
        const p1 = roundTrip('literal &amp;lt; here');
        const p2 = roundTrip(p1);
        const p3 = roundTrip(p2);

        expect(p1).toBe('literal &lt; here');
        expect(p2).toBe('literal < here');
        expect(p3).toBe(p2);
    });

    it('multi-encoded entity inside inline code is byte-stable across cycles', () => {
        const p1 = roundTrip('code `&amp;lt;` here');

        expect(p1).toBe('code `&amp;lt;` here');
        expect(roundTrip(p1)).toBe(p1);
    });
});

describe('Underline disabled — ++ never parses, C++/++flags prose survives', () => {
    it.each(['C++ then C++ again', 'a ++ b ++ c', 'compile ++flags++ here'])(
        'keeps %s byte-identical (no ++ collapse)',
        (s) => {
            expect(roundTrip(s)).toBe(s);
        },
    );
});

describe('bare URLs / emails / <autolinks> become links and converge; explicit [links] still work', () => {
    it.each([
        ['see https://example.com/path now', 'see [https://example.com/path](https://example.com/path) now'],
        ['contact me@example.com today', 'contact [me@example.com](mailto:me@example.com) today'],
        ['a <https://example.com> ref', 'a [https://example.com](https://example.com) ref'],
    ])('links %s then stays stable', (src, expected) => {
        const out = roundTrip(src);
        expect(out).toBe(expected);
        expect(roundTrip(out)).toBe(out);
    });

    it('still round-trips an explicit [link](url)', () => {
        expect(roundTrip('[docs](https://example.com)')).toContain('[docs](https://example.com)');
    });
});

describe('inline marks round-trip', () => {
    it.each([
        '**bold**',
        '*italic*',
        '`code span`',
        '~~strike~~',
        '[a link](https://example.com)',
        '**bold `code` end**',
    ])('preserves %s', (s) => {
        expect(roundTrip(s)).toContain(s);
    });
});

describe('TunedTable — cell pipes escaped + alignment preserved, idempotent', () => {
    it.each([
        [
            'pipe in a code-span cell',
            '| Op | Meaning |\n| --- | --- |\n| `\\|` | pipe op |\n| `\\|\\|` | or op |',
            'pipe op',
        ],
        [
            'pipes inside inline code',
            '| Name | Payload |\n| --- | --- |\n| chain | `echo \\| base64 \\| sh` |',
            'base64',
        ],
    ])('cell content survives two saves: %s', (_l, src, marker) => {
        const save1 = roundTrip(src);
        const save2 = roundTrip(save1);

        expect(save2).toBe(save1);
        expect(save2).toContain(marker);
    });

    it('preserves per-column alignment (left :--- / center :---: / right ---:)', () => {
        const save1 = roundTrip('| L | C | R |\n| :-- | :-: | --: |\n| a | b | c |');

        expect(save1).toContain('| :--- | :---: | ---: |');
        expect(roundTrip(save1)).toBe(save1);
    });
});

describe('nesting & sequencing — content preserved and converges (≤2 saves)', () => {
    it.each([
        ['bullet + code', '- item one\n- item two:\n\n  ```\n  code alpha\n  ```\n- item three'],
        ['ordered + code', '1. first\n2. second:\n\n   ```python\n   code beta\n   ```\n3. third'],
        ['nested bullets', '- outer aaa\n  - inner bbb\n    - deepest ccc'],
        ['blockquote + list', '> - quoted alpha\n> - quoted beta'],
        ['blockquote + code', '> intro\n>\n> ```\n> quoted code zeta\n> ```'],
        ['task list nested', '- [ ] task alpha\n  - [x] subtask beta\n- [x] task delta'],
        ['heading > list > code', '## Title\n\n- item alpha\n- item beta\n\n```\nstandalone epsilon\n```'],
        ['table > list > quote', '| A | B |\n| --- | --- |\n| 1 | 2 |\n\n- item one\n\n> quote alpha'],
    ])('%s', (_l, src) => {
        const save1 = roundTrip(src);
        const save2 = roundTrip(save1);

        expect(save2).toBe(save1);
        sameWords(src, save2);
    });
});

// A code block nested in a list (ordered > bullet > code) has an indented opening fence — the same root
// cause as a top-level indented fence, covered by parseTunedCodeBlock (see markdown-editor-extensions.ts).
describe('ordered > bullet > code — indented nested code survives', () => {
    it('keeps the deeply-nested code block', () => {
        const out = roundTrip('1. lvl1\n   - lvl2\n\n     ```\n     deepcode\n     ```');

        expect(out).toContain('deepcode');
    });
});

describe('findVariableOccurrences — doc spans for the Available-variables cycle', () => {
    const docOf = (md: string) => {
        const editor = new Editor({ content: md, contentType: 'markdown', extensions: createMarkdownExtensions() });
        const { doc } = editor.state;
        editor.destroy();

        return doc;
    };

    it('locates every {{.Var}} and each span maps to the literal token', () => {
        const doc = docOf('Use {{.Foo}} then {{.Foo}} and {{.Bar}}.');
        const foo = findVariableOccurrences(doc, 'Foo');

        expect(foo).toHaveLength(2);

        for (const hit of foo) {
            expect(doc.textBetween(hit.from, hit.to)).toBe('{{.Foo}}');
        }

        expect(findVariableOccurrences(doc, 'Bar')).toHaveLength(1);
    });

    it('returns none for an unused variable (caller then inserts instead of cycling)', () => {
        expect(findVariableOccurrences(docOf('Use {{.Foo}} only.'), 'Missing')).toHaveLength(0);
    });

    it('matches inside control-flow actions but honors word boundaries', () => {
        const doc = docOf('{{if .Enabled}}on{{end}} but not {{.EnabledExtra}}');

        expect(findVariableOccurrences(doc, 'Enabled')).toHaveLength(1);
    });

    it('finds a {{.Var}} split across text nodes by a mark (mark-boundary fix)', () => {
        const doc = docOf('**{{**.Foo}} tail');

        // sanity — the variable really is split: bolding a brace gives the textblock >1 inline child
        expect(doc.firstChild?.childCount ?? 0).toBeGreaterThan(1);

        const foo = findVariableOccurrences(doc, 'Foo');

        expect(foo).toHaveLength(1);

        for (const hit of foo) {
            expect(doc.textBetween(hit.from, hit.to)).toBe('{{.Foo}}');
        }
    });

    it('spans the full token when a hard break sits inside it (non-text-node off-by-N)', () => {
        const doc = docOf('{{.Foo\\\n}}');

        // sanity — the hard break really is inside the token (text, hardBreak, text)
        expect(doc.firstChild?.childCount ?? 0).toBeGreaterThan(1);

        const foo = findVariableOccurrences(doc, 'Foo');

        expect(foo).toHaveLength(1);

        for (const hit of foo) {
            expect(doc.textBetween(hit.from, hit.to)).toBe('{{.Foo}}');
        }
    });
});

describe('typing matches load — underscore emphasis literal, bare URL autolinks like load', () => {
    const typeString = (input: string): { html: string; md: string } => {
        const editor = new Editor({ content: '', contentType: 'markdown', extensions: createMarkdownExtensions() });
        const { view } = editor;

        for (const ch of input) {
            const { from } = view.state.selection;
            const handled = view.someProp('handleTextInput', (handler) =>
                handler(view, from, from, ch, () => view.state.tr),
            );

            if (!handled) {
                view.dispatch(view.state.tr.insertText(ch));
            }
        }

        const result = { html: editor.getHTML(), md: editor.getMarkdown() };
        editor.destroy();

        return result;
    };

    it('typed __dunder__ / _word_ stay literal (identifiers survive), matching load', () => {
        const { html, md } = typeString('call __init__ and _word_ done');

        expect(html).not.toContain('<strong>');
        expect(html).not.toContain('<em>');
        expect(md).toContain('__init__');
        expect(md).toContain('_word_');
    });

    it('typed bare URL autolinks — matches load', () => {
        const { html, md } = typeString('see https://evil.example.com/x done');

        expect(html).toContain('<a ');
        expect(md).toContain('[https://evil.example.com/x](https://evil.example.com/x)');
    });

    it('typed *italic* / **bold** / ~~strike~~ still convert (star + double-tilde kept)', () => {
        expect(typeString('a **bold** b').html).toContain('<strong>');
        expect(typeString('a *ital* b').html).toContain('<em>');
        expect(typeString('a ~~del~~ b').html).toContain('<s>');
    });
});

describe('line-leading # / > in a paragraph stay body text on round-trip (ENCODER-LEADING)', () => {
    const paraDoc = (...content: unknown[]) => ({ content: [{ content, type: 'paragraph' }], type: 'doc' });

    const save = (doc: unknown) => {
        const editor = new Editor({ content: doc as string, extensions: createMarkdownExtensions() });
        const md = editor.getMarkdown();

        editor.destroy();

        return md;
    };

    it('a Shift+Enter line starting with "# " is escaped and does not become a heading', () => {
        const md = save(paraDoc({ text: 'foo', type: 'text' }, { type: 'hardBreak' }, { text: '# bar', type: 'text' }));

        expect(md).toContain('\\# bar');
        expect(structuralCounts(md).heading ?? 0).toBe(0);
        expect(roundTrip(md)).toBe(md);
    });

    it('a Shift+Enter line starting with "> " does not become a blockquote', () => {
        const md = save(
            paraDoc({ text: 'foo', type: 'text' }, { type: 'hardBreak' }, { text: '> quote', type: 'text' }),
        );

        expect(md).toContain('\\> quote');
        expect(structuralCounts(md).blockquote ?? 0).toBe(0);
        expect(roundTrip(md)).toBe(md);
    });

    it('an escaped \\# at line start loads as clean body text and converges', () => {
        expect(structuralCounts(roundTrip('\\# not a heading')).heading ?? 0).toBe(0);
        expect(roundTrip(roundTrip('\\# x'))).toBe(roundTrip('\\# x'));
    });

    it('a REAL heading / blockquote is untouched', () => {
        expect(structuralCounts(roundTrip('# foo')).heading).toBe(1);
        expect(structuralCounts(roundTrip('> foo')).blockquote).toBe(1);
    });

    it('only # and > are escaped — regex/glob literals and a mid-line # survive byte-identical', () => {
        for (const s of ['glob \\* and \\? here', 'regex \\d+ and \\.php', 'see #123 issue', 'a # b mid line']) {
            expect(roundTrip(s)).toBe(s);
        }
    });
});

describe('multi-paragraph table cell does not persist a raw control byte on save', () => {
    const header = (text: string) => ({
        content: [{ content: [{ text, type: 'text' }], type: 'paragraph' }],
        type: 'tableHeader',
    });
    const cell = (...paragraphs: string[]) => ({
        content: paragraphs.map((text) => ({ content: [{ text, type: 'text' }], type: 'paragraph' })),
        type: 'tableCell',
    });

    it('collapses two cell paragraphs to a space instead of joining with U+001F', () => {
        const doc = {
            content: [
                {
                    content: [
                        { content: [header('h'), header('i')], type: 'tableRow' },
                        { content: [cell('one', 'two'), cell('z')], type: 'tableRow' },
                    ],
                    type: 'table',
                },
            ],
            type: 'doc',
        };
        const editor = new Editor({ content: doc, extensions: createMarkdownExtensions() });
        const md = editor.getMarkdown();

        editor.destroy();

        expect(md).not.toContain('\u001f');
        expect(md).toContain('one two');
        expect(md).toContain('z');
    });
});
