import { Editor } from '@tiptap/core';
import { OrderedList, TaskList } from '@tiptap/extension-list';
import { TextSelection } from '@tiptap/pm/state';
import { beforeAll, describe, expect, it } from 'vitest';

import { createMarkdownExtensions, isBlockApplied } from './markdown-editor-extensions';
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

    // A doc-side cell text (typed/pasted in rich mode) can hold a backslash run right before a pipe — a
    // sequence a markdown LOAD can never produce. Serialize it through the editor and reload to prove the
    // escaped pipe survives as cell content instead of becoming a column delimiter that truncates the row.
    const saveTableCell = (cellText: string): string => {
        const cell = (type: string, text: string) => ({
            content: [{ content: [{ text, type: 'text' }], type: 'paragraph' }],
            type,
        });
        const editor = new Editor({
            content: {
                content: [
                    {
                        content: [
                            { content: [cell('tableHeader', 'op'), cell('tableHeader', 'note')], type: 'tableRow' },
                            { content: [cell('tableCell', cellText), cell('tableCell', 'tail')], type: 'tableRow' },
                        ],
                        type: 'table',
                    },
                ],
                type: 'doc',
            },
            extensions: createMarkdownExtensions(),
        });
        const out = editor.getMarkdown();

        editor.destroy();

        return out;
    };

    it.each([
        ['single backslash before pipe', 'a\\|b'],
        ['triple backslash before pipe', 'a\\\\\\|b'],
        ['lone backslash-pipe cell', '\\|'],
    ])('typed cell with %s keeps the row and its trailing cell across saves', (_l, cellText) => {
        const save1 = saveTableCell(cellText);
        const save2 = roundTrip(save1);

        expect(save2, `row truncated:\n${save1}\n-->\n${save2}`).toBe(save1);
        expect(save2).toContain('tail');
        expect(structuralCounts(save2)).toEqual({ table: 1, tableCell: 2, tableHeader: 2, tableRow: 2 });
    });

    it('round-trips an even backslash run before a pipe byte-identical', () => {
        const save1 = saveTableCell('a\\\\|b');

        expect(save1).toContain('a\\\\\\|b');
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

type MarkdownTokenizerShape = {
    start?: (src: string) => number;
    tokenize: (src: string, tokens: unknown[], lexer: unknown) => unknown;
};

describe('list tokenizers are guarded against the upstream per-block full-source split', () => {
    type Layer = {
        config: { markdownTokenizer?: MarkdownTokenizerShape };
        parent?: Layer;
    };

    // `.extend()` MERGES config into the child, so every wrapper layer reports a `markdownTokenizer` and
    // walking to "the first layer that has one" lands on a wrapper, not on upstream. The root of the chain is
    // the only unambiguous handle on the untouched implementation.
    const rootOf = (layer: Layer): Layer => {
        let current = layer;

        while (current.parent) {
            current = current.parent;
        }

        return current;
    };

    const extensionOf = (name: string) => {
        const editor = new Editor({ content: '', extensions: createMarkdownExtensions() });
        const extension = editor.extensionManager.extensions.find((e) => e.name === name) as Layer | undefined;

        editor.destroy();

        return extension;
    };

    const tokenizerOf = (name: string) => extensionOf(name)?.config.markdownTokenizer;

    const upstreamTokenizerOf = (name: string) => {
        const extension = extensionOf(name);

        return extension ? rootOf(extension).config.markdownTokenizer : undefined;
    };

    it.each([
        ['orderedList', OrderedList],
        ['taskList', TaskList],
    ])('%s tokenize is wrapped, not the upstream function', (name, upstream) => {
        const ours = tokenizerOf(name);
        const theirs = (upstream.config as { markdownTokenizer?: { tokenize: unknown } }).markdownTokenizer;

        expect(theirs?.tokenize).toBeTypeOf('function');
        expect(ours?.tokenize).toBeTypeOf('function');
        // An unwrapped tokenizer splits the whole remaining document at every block boundary — O(n²) loads.
        expect(ours?.tokenize).not.toBe(theirs?.tokenize);
    });

    // The guard must stay at least as wide as ORDERED_LIST_MARKER_PATTERN
    // (`\d+|[ivxlcdmIVXLCDM]+|[a-zA-Z]{1,2}`); a marker form it rejects is silently demoted to a paragraph.
    it.each([
        ['single digit', '1. alpha\n2. beta'],
        ['multi digit', '10. alpha\n11. beta'],
        ['lower roman', 'i. alpha\nii. beta'],
        ['upper roman', 'IX. alpha\nX. beta'],
        ['single alpha', 'a. alpha\nb. beta'],
        ['two-letter alpha', 'aa. alpha\nab. beta'],
        ['two-letter upper alpha', 'AB. alpha\nAC. beta'],
    ])('ordered list with a %s marker survives the guard', (_name, md) => {
        // Byte equality alone is blind here: a demoted paragraph holding the same two lines re-serialises
        // to identical bytes, so the node type is the only assertion that can see the marker being rejected.
        expect(structuralCounts(roundTrip(md)).orderedList).toBe(1);
        expect(roundTrip(md)).toBe(md);
    });

    // Upstream normalises `)` to `.` and drops a leading indent (verified identical without the guard), so
    // these two assert only what the guard governs: the block is still a list, not demoted to a paragraph.
    it.each([
        ['paren delimiter', 'aa) alpha\nab) beta'],
        ['indented', '  1. alpha\n  2. beta'],
    ])('ordered list with a %s marker is not demoted to a paragraph', (_name, md) => {
        expect(structuralCounts(roundTrip(md)).orderedList).toBe(1);
    });

    // Identity alone would also hold for a pass-through wrapper, which would silently restore the O(n^2)
    // load. This is the assertion that the short-circuit itself is wired.
    it.each(['orderedList', 'taskList'])(
        '%s never reaches upstream tokenize when the source does not open a list',
        (name) => {
            const upstream = upstreamTokenizerOf(name);
            const wrapped = tokenizerOf(name);

            expect(upstream?.tokenize).toBeTypeOf('function');

            const real = upstream!.tokenize;
            let calls = 0;

            upstream!.tokenize = (...args: Parameters<typeof real>) => {
                calls += 1;

                return real(...args);
            };

            try {
                expect(wrapped?.tokenize('plain paragraph line\n'.repeat(500), [], undefined)).toBeUndefined();
                expect(calls).toBe(0);
            } finally {
                upstream!.tokenize = real;
            }
        },
    );

    it.each([
        ['dash unchecked', '- [ ] open'],
        ['dash checked', '- [x] done'],
        ['asterisk', '* [ ] open'],
        ['plus', '+ [x] done'],
        ['upper X', '- [X] done'],
    ])('task list with a %s marker survives the guard', (_name, md) => {
        expect(structuralCounts(roundTrip(md)).taskList).toBe(1);
    });
});

describe('block toggles are reversible under a whole-document selection', () => {
    const CONTROLS = [
        ['blockquote', 'toggleBlockquote'],
        ['bulletList', 'toggleBulletList'],
        ['orderedList', 'toggleOrderedList'],
        ['taskList', 'toggleTaskList'],
        ['codeBlock', 'toggleCodeBlock'],
    ] as const;

    const apply = (editor: Editor, command: string) =>
        (editor.chain().focus() as unknown as Record<string, () => { run: () => boolean }>)[command]!().run();

    const editorWith = (markdown: string) =>
        new Editor({ content: markdown, contentType: 'markdown', extensions: createMarkdownExtensions() });

    it.each(CONTROLS)('%s applied over Ctrl+A is removed by a second press', (node, command) => {
        const editor = editorWith('sample text');

        editor.commands.selectAll();
        apply(editor, command);

        expect(structuralCounts(editor.getMarkdown())[node]).toBe(1);

        editor.commands.selectAll();
        apply(editor, command);

        const out = editor.getMarkdown();

        editor.destroy();

        // Node counts, not bytes: a second wrap nests a quote / adds an empty item, and several of those
        // re-serialise to text that still "contains" the original words.
        expect(structuralCounts(out)[node]).toBeUndefined();
        expect(out.trim()).toBe('sample text');
    });

    it.each(CONTROLS)('%s stays reversible over a multi-block Ctrl+A', (node, command) => {
        const editor = editorWith('one\n\ntwo\n\nthree');

        editor.commands.selectAll();
        apply(editor, command);
        editor.commands.selectAll();
        apply(editor, command);

        const out = editor.getMarkdown();

        editor.destroy();

        expect(structuralCounts(out)[node]).toBeUndefined();
        expect(words(out).sort()).toEqual(['one', 'three', 'two']);
    });

    // The keyboard shortcuts (Mod-Shift-b / Mod-Alt-c / Mod-Shift-7..9) go through `editor.commands.x()`,
    // not through a chain, so the retarget has to hold on that path too or only the toolbar is fixed.
    it.each(CONTROLS)('%s is reversible through editor.commands, the path the shortcuts use', (node, command) => {
        const editor = editorWith('sample text');
        const run = () => (editor.commands as unknown as Record<string, () => boolean>)[command]!();

        editor.commands.selectAll();
        run();

        expect(structuralCounts(editor.getMarkdown())[node]).toBe(1);

        editor.commands.selectAll();
        run();

        const out = editor.getMarkdown();

        editor.destroy();

        expect(structuralCounts(out)[node]).toBeUndefined();
        expect(out.trim()).toBe('sample text');
    });

    it.each(CONTROLS)('%s is unaffected for a caret selection', (node, command) => {
        const editor = editorWith('sample text');

        editor.commands.setTextSelection(4);
        apply(editor, command);

        expect(structuralCounts(editor.getMarkdown())[node]).toBe(1);

        apply(editor, command);

        const out = editor.getMarkdown();

        editor.destroy();

        expect(structuralCounts(out)[node]).toBeUndefined();
        expect(out.trim()).toBe('sample text');
    });
});

describe('a task checkbox is announced with its own text, not its subtree', () => {
    const labelsFor = (markdown: string) => {
        const element = document.createElement('div');

        document.body.appendChild(element);

        const editor = new Editor({
            content: markdown,
            contentType: 'markdown',
            element,
            extensions: createMarkdownExtensions(),
        });
        const labels = [...element.querySelectorAll('input[type=checkbox]')].map(
            (checkbox) => (checkbox as HTMLInputElement).ariaLabel,
        );

        editor.destroy();
        element.remove();

        return labels;
    };

    it('a parent task does not absorb the text of its nested tasks', () => {
        expect(labelsFor('- [ ] task open\n- [x] task done\n    - [ ] nested open\n    - [x] nested done\n')).toEqual([
            'Task item checkbox for task open',
            'Task item checkbox for task done',
            'Task item checkbox for nested open',
            'Task item checkbox for nested done',
        ]);
    });

    it('an empty task still gets a name', () => {
        expect(labelsFor('- [ ] \n')).toEqual(['Task item checkbox for empty task item']);
    });
});

describe('whole-document block toggles across document shapes', () => {
    const TOGGLES = [
        ['blockquote', 'toggleBlockquote'],
        ['bulletList', 'toggleBulletList'],
        ['orderedList', 'toggleOrderedList'],
        ['taskList', 'toggleTaskList'],
        ['codeBlock', 'toggleCodeBlock'],
    ] as const;

    // The shapes whose first child is not a textblock are the ones a hardcoded start position gets wrong; the
    // plain ones pass either way, so trimming the list to them takes the coverage away.
    const SHAPES = [
        ['single paragraph', 'sample text'],
        ['three paragraphs', 'one\n\ntwo\n\nthree'],
        ['table only', '| A | B |\n| --- | --- |\n| 1 | 2 |'],
        ['table then paragraph', '| A | B |\n| --- | --- |\n| 1 | 2 |\n\ntail'],
        ['list only', '- a\n- b\n- c'],
        ['nested list', '- a\n  - b\n- c'],
        ['leading rule', '---\n\nbody text'],
        ['leading image', '![alt](https://x/y.png)\n\nbody text'],
        ['leading heading', '# Title\n\nbody'],
        ['fenced code only', '```js\nconst a = 1;\n```'],
        ['already quoted', '> quoted\n\nplain'],
    ] as const;

    const editorWith = (markdown: string) =>
        new Editor({ content: markdown, contentType: 'markdown', extensions: createMarkdownExtensions() });

    const press = (editor: Editor, command: string) =>
        (editor.commands as unknown as Record<string, () => boolean>)[command]!();

    const topLevelTypes = (editor: Editor) => {
        const types: string[] = [];

        editor.state.doc.forEach((node) => types.push(node.type.name));

        return types;
    };

    // Conversions that drop an attribute the target block cannot carry (a listItem holds no heading level, a
    // paragraph no fence info-string). Measured irreversible with a plain caret too, so making these byte-exact
    // is an upstream change, not a bug in this path.
    const LOSSY = new Set([
        'fenced code only|toggleBulletList',
        'fenced code only|toggleCodeBlock',
        'fenced code only|toggleOrderedList',
        'fenced code only|toggleTaskList',
        'leading heading|toggleBulletList',
        'leading heading|toggleCodeBlock',
        'leading heading|toggleOrderedList',
        'leading heading|toggleTaskList',
        'list only|toggleOrderedList',
        'list only|toggleTaskList',
        'nested list|toggleOrderedList',
    ]);

    const cases = SHAPES.flatMap(([shape, source]) =>
        TOGGLES.map(([, command]) => ({ command, isLossy: LOSSY.has(`${shape}|${command}`), shape, source })),
    );

    // structuralCounts is blind to the destruction this guards: a rule left OUTSIDE the quote and one inside it
    // produce identical counts, and a table rebuilt around itself still counts one table.
    it.each(cases)(
        '$shape + $command either applies reversibly or refuses harmlessly',
        ({ command, isLossy, source }) => {
            const editor = editorWith(source);

            // Baseline AFTER selectAll: TrailingNode materialises its empty paragraph on that transaction, so a
            // baseline taken before it would blame the toggle for a block the selection itself added.
            editor.commands.selectAll();

            const original = editor.getMarkdown();
            const originalText = editor.state.doc.textContent;
            const originalTypes = topLevelTypes(editor);
            const applied = press(editor, command);

            if (!applied) {
                const { selection } = editor.state;

                expect(editor.getMarkdown()).toBe(original);
                expect(topLevelTypes(editor)).toEqual(originalTypes);
                expect(selection.empty).toBe(false);
                editor.destroy();

                return;
            }

            expect(editor.getMarkdown()).not.toBe(original);

            editor.commands.selectAll();
            press(editor, command);

            const out = editor.getMarkdown();
            const outText = editor.state.doc.textContent;

            editor.destroy();

            if (isLossy) {
                // Text content, not markdown: the declared loss is an attribute (heading level, fence info-string),
                // and comparing source words would count `js` from ```js as content.
                sameWords(originalText, outText);

                return;
            }

            expect(out.trimEnd()).toBe(original.trimEnd());
        },
    );

    // Selecting everything with the mouse is a TextSelection, so a guard keyed on AllSelection misses it entirely
    // and leaves this path on the unfixed code.
    it.each(TOGGLES)('%s is reversible under a drag-style whole-document selection', (_node, command) => {
        const editor = editorWith('one\n\ntwo');

        const selectByDrag = () => {
            const { doc } = editor.state;

            editor.view.dispatch(
                editor.state.tr.setSelection(TextSelection.between(doc.resolve(1), doc.resolve(doc.content.size - 1))),
            );
        };

        selectByDrag();

        expect(editor.state.selection).toBeInstanceOf(TextSelection);
        expect(press(editor, command)).toBe(true);

        selectByDrag();
        press(editor, command);

        const out = editor.getMarkdown();

        editor.destroy();

        expect(out.trimEnd()).toBe('one\n\ntwo');
    });

    it.each(TOGGLES)('%s reports the same state to the toolbar as its own click performs', (node, command) => {
        const editor = editorWith('sample text');

        editor.commands.selectAll();

        expect(isBlockApplied(editor, node)).toBe(false);

        press(editor, command);
        editor.commands.selectAll();

        expect(isBlockApplied(editor, node)).toBe(true);

        press(editor, command);
        editor.commands.selectAll();

        expect(isBlockApplied(editor, node)).toBe(false);
        editor.destroy();
    });

    it('reports an already-quoted document as quoted under a whole-document selection', () => {
        const editor = editorWith('> quoted');

        editor.commands.selectAll();

        expect(isBlockApplied(editor, 'blockquote')).toBe(true);
        editor.destroy();
    });

    it.each(TOGGLES)('%s leaves the document untouched when only probed through can()', (_node, command) => {
        const editor = editorWith('| A | B |\n| --- | --- |\n| 1 | 2 |');

        editor.commands.selectAll();

        const original = editor.getMarkdown();

        (editor.can() as unknown as Record<string, () => boolean>)[command]!();

        const out = editor.getMarkdown();

        editor.destroy();

        expect(out).toBe(original);
    });

    // Only these commands may drive a disabled state in the toolbar. `can()` runs with dispatch undefined, and
    // `toggleList`'s `clearNodes()` fallback is a no-op there, so the three list kinds, toggleCodeBlock and
    // toggleHeading report false for conversions that work. A tiptap upgrade that makes one faithful may add it.
    it.each([
        [
            'toggleBlockquote',
            (editor: Editor) => editor.can().toggleBlockquote(),
            (editor: Editor) => editor.commands.toggleBlockquote(),
        ],
        ['toggleBold', (editor: Editor) => editor.can().toggleBold(), (editor: Editor) => editor.commands.toggleBold()],
        ['toggleCode', (editor: Editor) => editor.can().toggleCode(), (editor: Editor) => editor.commands.toggleCode()],
        [
            'toggleItalic',
            (editor: Editor) => editor.can().toggleItalic(),
            (editor: Editor) => editor.commands.toggleItalic(),
        ],
        [
            'toggleStrike',
            (editor: Editor) => editor.can().toggleStrike(),
            (editor: Editor) => editor.commands.toggleStrike(),
        ],
        [
            'setLink',
            (editor: Editor) => editor.can().setLink({ href: 'https://example.com' }),
            (editor: Editor) => editor.commands.setLink({ href: 'https://example.com' }),
        ],
    ] as const)(
        '%s answers can() exactly as it performs, on every shape and both selection modes',
        (_name, ask, perform) => {
            const disagreements: string[] = [];

            for (const [shape, source] of SHAPES) {
                for (const mode of ['caret', 'selectAll'] as const) {
                    const probe = editorWith(source);

                    if (mode === 'caret') {
                        probe.commands.setTextSelection(3);
                    } else {
                        probe.commands.selectAll();
                    }

                    const canSays = ask(probe);
                    const commandSays = perform(probe);

                    probe.destroy();

                    if (canSays !== commandSays) {
                        disagreements.push(`${shape}/${mode}: can=${canSays} command=${commandSays}`);
                    }
                }
            }

            expect(disagreements).toEqual([]);
        },
    );

    it('builds exactly one lowlight plugin', () => {
        const editor = new Editor({ content: '', extensions: createMarkdownExtensions() });
        const lowlightPlugins = editor.state.plugins.filter((plugin) =>
            String((plugin as { key?: string }).key).startsWith('lowlight'),
        );

        editor.destroy();

        expect(lowlightPlugins).toHaveLength(1);
    });
});

describe('line-leading block markers are escaped to marked grammar, inline backslashes are not touched', () => {
    // marked's heading is /^ {0,3}(#{1,6})(?=\s|$)…/ and its blockquote is /^( {0,3}> ?…)+/ — the space after
    // the marker is a lookahead, not a literal. A serializer that escapes only `# `/`> ` leaves these live, and
    // the block type flips on the next load.
    // Starts from a PARAGRAPH node, not from markdown: `>foo` as stored markdown is legitimately a quote. The
    // defect is that a paragraph the user typed serializes to bytes that read back as a different block.
    const paragraphSurvives = (text: string) => {
        const editor = new Editor({
            content: { content: [{ content: [{ text, type: 'text' }], type: 'paragraph' }], type: 'doc' },
            extensions: createMarkdownExtensions(),
        });
        const saved = editor.getMarkdown();

        editor.destroy();

        const reloaded = new Editor({
            content: saved,
            contentType: 'markdown',
            extensions: createMarkdownExtensions(),
        });
        const out = { counts: structuralCounts(saved), text: reloaded.state.doc.textContent };

        reloaded.destroy();

        return out;
    };

    it.each([
        ['>foo', 'a quote marker with no space'],
        ['#', 'a lone hash'],
        ['###', 'three hashes alone'],
        ['#\tfoo', 'a hash followed by a tab'],
        ['>', 'a lone quote marker'],
    ])('a typed paragraph of %s stays a paragraph (%s)', (text) => {
        expect(paragraphSurvives(text)).toEqual({ counts: {}, text });
    });

    it('leaves a real heading and a real quote alone', () => {
        expect(structuralCounts(roundTrip('# real heading'))).toEqual({ heading: 1 });
        expect(structuralCounts(roundTrip('> real quote'))).toEqual({ blockquote: 1 });
    });

    it('leaves seven hashes as a paragraph, exactly as marked reads them', () => {
        expect(roundTrip('####### seven')).toBe('####### seven');
    });

    // The load side used to unescape `\#`/`\>` at ANY inline position while the serializer only re-escaped them
    // line-leading, so a backslash in the middle of a line was eaten. `\<root\>` is a GNU-grep word boundary;
    // dropping the second backslash changes what the command matches.
    it.each([
        'use \\# for comments and \\> for redirect',
        "grep '\\<root\\>' /etc/passwd",
        'sed -e "s/\\>/gt/" file',
        'shell: cat file \\> out.txt',
    ])('keeps an inline backslash escape byte-identical: %s', (source) => {
        expect(roundTrip(source)).toBe(source);
    });
});

describe('a setext underline after a hard break does not turn the paragraph into a heading', () => {
    const hardBreakParagraph = (first: string, second: string) => {
        const editor = new Editor({
            content: {
                content: [
                    {
                        content: [{ text: first, type: 'text' }, { type: 'hardBreak' }, { text: second, type: 'text' }],
                        type: 'paragraph',
                    },
                ],
                type: 'doc',
            },
            extensions: createMarkdownExtensions(),
        });
        const saved = editor.getMarkdown();

        editor.destroy();

        const reloaded = new Editor({
            content: saved,
            contentType: 'markdown',
            extensions: createMarkdownExtensions(),
        });
        const out = { counts: structuralCounts(saved), saved, text: reloaded.state.doc.textContent };

        reloaded.destroy();

        return out;
    };

    // marked's lheading run is `(=+|-+) *` — no three-character minimum, and trailing spaces are allowed. A
    // single `-` on the line after a hard break was enough to swallow the paragraph into an H2.
    it.each(['===', '---', '=', '-', '--', '===   '])('a continuation line of %s stays text', (underline) => {
        const result = hardBreakParagraph('foo', underline);

        expect(result.counts).toEqual({});
        expect(result.text).toBe(`foo${underline}`);
    });

    it('leaves a real setext heading in stored markdown alone', () => {
        expect(structuralCounts(roundTrip('Title\n===')).heading).toBe(1);
    });

    // The escape must not leak a literal backslash into the document text, and it must not eat a backslash the
    // author wrote inside a character class.
    it.each(['char class [a-z\\-_] here', 'sed s/a\\=b/x/', 'a \\- b and c \\= d'])(
        'keeps an inline backslash before = or - byte-identical: %s',
        (source) => {
            expect(roundTrip(source)).toBe(source);
        },
    );
});

describe('brackets in a link label or an image alt do not break the node on reload', () => {
    const reload = (markdown: string) => {
        const editor = new Editor({
            content: markdown,
            contentType: 'markdown',
            extensions: createMarkdownExtensions(),
        });
        const out = {
            counts: structuralCounts(markdown),
            markdown: editor.getMarkdown(),
            text: editor.state.doc.textContent,
        };

        editor.destroy();

        return out;
    };

    const linkDoc = (label: string) => ({
        content: [
            {
                content: [
                    { marks: [{ attrs: { href: 'https://example.com/a' }, type: 'link' }], text: label, type: 'text' },
                ],
                type: 'paragraph',
            },
        ],
        type: 'doc',
    });

    const imageDoc = (alt: string) => ({
        content: [{ attrs: { alt, src: 'https://example.com/a.png' }, type: 'image' }],
        type: 'doc',
    });

    const saveOf = (content: object) => {
        const editor = new Editor({ content, extensions: createMarkdownExtensions() });
        const saved = editor.getMarkdown();

        editor.destroy();

        return saved;
    };

    // A bracket-free label round-trips; the bracket case is handled where the link is inserted, because the
    // label text is serialized by the text path and @tiptap/markdown hands a mark's renderMarkdown a
    // placeholder rather than the text, so the escape cannot live there.
    it.each(['a\\b', 'plain label', 'http://[::1]:8080/x'])('a link label of %s survives a round trip', (label) => {
        const saved = saveOf(linkDoc(label));
        const first = reload(saved);

        expect(first.text).toBe(label);
        expect(reload(first.markdown).markdown).toBe(first.markdown);
    });

    // An unbalanced bracket in alt broke `![alt](src)` outright: the reload parsed ZERO image nodes.
    it.each(['a]b', 'a[b'])('an image alt of %s keeps the image node', (alt) => {
        const saved = saveOf(imageDoc(alt));

        expect(structuralCounts(saved).image).toBe(1);
        expect(reload(saved).markdown).toBe(saved);
    });
});

describe('a URL inserted as its own link label settles instead of growing', () => {
    // The Link popover inserts the URL as the visible text when there is no selection. A `]` in that text
    // closed the markdown label early — `[https://example.com/a]b](…)` — and the serialized form grew on
    // every load+save cycle (measured 50 → 103 → 156 → 260). The form keeps brackets out of the label; this
    // pins the property that matters, that the document converges.
    it.each(['https://example.com/a]b', 'http://[::1]:8080/x', 'https://example.com/plain'])(
        'inserting %s converges',
        (url) => {
            const editor = new Editor({ content: '', extensions: createMarkdownExtensions() });

            editor
                .chain()
                .insertContent({
                    marks: [{ attrs: { href: url }, type: 'link' }],
                    text: url.replace(/[[\]]/g, ''),
                    type: 'text',
                })
                .run();

            const first = editor.getMarkdown();

            editor.destroy();

            expect(roundTrip(first)).toBe(first);
        },
    );
});

describe('a heading applied over the whole document', () => {
    const apply = (source: string, level: number) => {
        const editor = new Editor({ content: source, contentType: 'markdown', extensions: createMarkdownExtensions() });

        editor.commands.selectAll();
        editor.commands.toggleHeading({ level: level as 1 | 2 | 3 | 4 | 5 | 6 });

        // selectAll materialises TrailingNode's empty paragraph, which serialises as a trailing blank line.
        const out = editor.getMarkdown().trimEnd();

        editor.destroy();

        return out;
    };

    // toggleHeading had no whole-document wrapper, so Ctrl+A + "Heading 2" converted every table cell paragraph
    // to a heading; saved as `| ## h | ## i |`, it reloads as literal cell text and escalates on each save.
    it('leaves table cells alone', () => {
        const source = ['lead', '', '| h | i |', '| --- | --- |', '| x | y |'].join('\n');
        const out = apply(source, 2);

        expect(out).not.toContain('## h');
        expect(out).toContain('| h');
        expect(out).toContain('## lead');
    });

    // The level the user picked has to reach setBlockType — a family that drops the attrs stamps H1 for every
    // choice, and one that compares only the node type reads an all-H1 document as "already Heading 2".
    it('applies the level the user chose, not the schema default', () => {
        expect(apply('intro\n\nsecond', 2)).toBe('## intro\n\n## second');
    });

    it('promotes an all-H1 document to the chosen level instead of demoting it', () => {
        expect(apply('# one\n\n# two', 2)).toBe('## one\n\n## two');
    });

    it('demotes to paragraphs when the chosen level is already applied', () => {
        expect(apply('## one\n\n## two', 2)).toBe('one\n\ntwo');
    });
});
