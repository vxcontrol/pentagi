import type { JSONContent, MarkdownRendererHelpers } from '@tiptap/core';

import { Extension } from '@tiptap/core';
import { renderTableToMarkdown, Table } from '@tiptap/extension-table';
import { Markdown } from '@tiptap/markdown';
import { Marked } from 'marked';

import { escapeTablePipes } from './markdown-editor-table-pipes';

// @tiptap/markdown parses with `marked`, which — unlike markdown-it's `html: false` — always tries to
// interpret `<...>` as HTML. Our content uses literal XML-ish tags (`<container_environment>`, `<input>`)
// that must survive verbatim; marked silently swallows the ones whose names match real HTML elements
// (`<input>`, `<br>`, …). Neutralising marked's block (`html`) and inline (`tag`) HTML tokenizers makes
// every `<...>` fall through to plain text, recreating markdown-it's `html: false`.
const createTunedMarked = () => {
    const instance = new Marked();

    instance.use({
        tokenizer: {
            // These marked tokenizers auto-convert literal text into markup, mangling Go-template / pentest
            // prose on round-trip. Returning `undefined` forces the char to stay literal text; `false` defers
            // to marked's default. Neutralise the lossy cases, keep what the toolbar emits:
            //   • del      — keep GFM `~~strike~~`, drop a lone `~…~` (else `~5~` / ranges become <del>)
            //   • emStrong — keep `*`/`**`, drop `_`-delimited emphasis (else `__init__`/`_word_` become em/strong)
            //   • escape   — keep `\`+punct literal (`\.` `\*` `\|` `\\`); marked's default DROPS the backslash
            //                (CommonMark unescape), silently corrupting regex/paths on the first load. EXCEPT
            //                `\#`/`\>`: unescape those, symmetric with the paragraph serializer escaping a
            //                line-leading `# `/`> ` so body text doesn't re-parse as a heading/quote
            //   • html/tag — keep `<xml-like>` tags literal (marked swallows real-HTML-element names)
            // NB: autolink/url are intentionally NOT neutralised — a bare `https://…`, `<url>` or email is
            // meant to become a link (see markdown-editor-extensions.ts link config, kept symmetric with typing).
            // NB: named HTML entities (`&lt; &gt; &amp; &quot;`) are decoded downstream — @tiptap/markdown's
            // token parsing runs @tiptap/core's decodeHtmlEntities, so a bare-prose `&lt;` becomes `<` (fixes
            // HTML-encoding artifacts from ingestion). Numeric refs (`&#123;`) and anything inside code are
            // untouched; a bare `&` survives as `&`. Don't re-add a literalAmpersand token to "preserve"
            // `&lt;` — that re-freezes the artifacts.
            del: (src: string) => (/^~~(?!~)/.test(src) ? false : undefined),
            emStrong: (src: string) => (/^_/.test(src) ? undefined : false),
            escape: (src: string) => {
                const marker = /^\\([#>])/.exec(src);

                return marker ? { raw: marker[0], text: marker[1]!, type: 'escape' as const } : undefined;
            },
            html: () => undefined,
            tag: () => undefined,
        },
    });

    // Load-side counterpart to TunedTable's save-side pipe escape: protect a `|` inside a code span / Go
    // action in a table cell BEFORE marked's table tokenizer splits the row (it splits raw `|` before inline
    // tokenization, dropping trailing cells). @tiptap/markdown's manager builds its block lexer via
    // `new markedInstance.Lexer(...)` — including for the construction-time initial parse — so subclassing
    // this PRIVATE instance's Lexer keeps the transform instance-scoped (no shared-class mutation) and still
    // catches every load. `inlineTokens` is inherited unchanged, so inline fragments are not touched.
    const BaseLexer = instance.Lexer;

    instance.Lexer = class extends BaseLexer {
        override lex(src: string) {
            return super.lex(escapeTablePipes(src));
        }
    } as typeof BaseLexer;

    // A PRIVATE instance — mutating the shared global `marked` would also affect report-pdf.
    return instance;
};

// The serialize-side counterpart to createTunedMarked. @tiptap/markdown's MarkdownManager
// .encodeTextForMarkdown HTML-entity-encodes text (`<` → `&lt;`) and backslash-escapes ``` ` * _ [ ] ~ \ ```.
// Both are wrong here: the load side keeps `\`+punct literal (escape tokenizer off), so re-escaping on save
// would double every backslash; and named HTML entities are DECODED on load (`&lt;`→`<`), so re-encoding
// would freeze them back into `&lt;`. Serialization is hard-coded in the manager (no per-extension hook), so
// replace that one method with identity — save emits exactly the text the doc holds.
type ManagerWithEncode = { encodeTextForMarkdown: (text: string) => string };

const TunedMarkdownText = Extension.create({
    name: 'tunedMarkdownText',
    onBeforeCreate() {
        // editor.markdown is always assigned by the Markdown extension (priority ordering below); the
        // non-optional cast makes a future regression throw here instead of silently reverting save to the
        // lossy default encoder.
        const manager = this.editor.markdown as unknown as ManagerWithEncode;

        manager.encodeTextForMarkdown = (text) => text;
    },
    // Lower priority than the Markdown extension so this runs AFTER its onBeforeCreate has created the
    // manager and assigned editor.markdown. onBeforeCreate (not onCreate) because it is synchronous —
    // a headless editor's onCreate fires after construction, too late for the first getMarkdown().
    priority: 50,
});

// @tiptap/extension-table's renderTableToMarkdown is alignment-aware but never escapes pipes, so a literal
// `|` a cell emits (even from inside inline code) would re-parse as a column delimiter on the next SAVE and
// drop cells (tiptap PR #7884). renderTableToMarkdown emits cell content only via h.renderChildren, so wrapping
// that one call to escape pipes fixes the save side, on the official alignment-aware renderer.
// The LOAD side of the same class is covered by the tuned Lexer subclass in createTunedMarked above.
export const TunedTable = Table.extend({
    renderMarkdown(node: JSONContent, helpers: MarkdownRendererHelpers) {
        const pipeEscaping: MarkdownRendererHelpers = {
            ...helpers,
            renderChildren: (nodes, separator) => helpers.renderChildren(nodes, separator).replace(/\|/g, '\\|'),
        };

        // renderTableToMarkdown joins a multi-block cell's children (e.g. two paragraphs from Enter-in-cell) with
        // a U+001F separator, outside renderChildren. A GFM cell is single-line, so collapse it to a space rather
        // than persist a raw control byte that survives into the saved .tmpl/knowledge and round-trips unchanged.
        return renderTableToMarkdown(node, pipeEscaping).replaceAll(String.fromCharCode(0x1f), ' ');
    },
});

export const createMarkdownLayer = () => [
    Markdown.configure({ marked: createTunedMarked() as unknown as typeof import('marked').marked }),
    TunedMarkdownText,
];
