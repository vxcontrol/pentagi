import type { JSONContent, MarkdownRendererHelpers } from '@tiptap/core';

import { Extension } from '@tiptap/core';
import { renderTableToMarkdown, Table } from '@tiptap/extension-table';
import { Markdown } from '@tiptap/markdown';
import { Marked } from 'marked';

import { escapeTablePipes, TEMPLATE_ACTION } from './markdown-editor-table-pipes';

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
            //                (CommonMark unescape), silently corrupting regex/paths on the first load. The
            //                `\#`/`\>` counterpart to the paragraph serializer is NOT here: an inline tokenizer
            //                fires at any position, so it also ate the backslash mid-line (`grep '\<root\>'`).
            //                It lives in the Lexer's inlineTokens override below, which sees line starts.
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
            escape: () => undefined,
            html: () => undefined,
            tag: () => undefined,
        },
    });

    // Load-side counterpart to TunedTable's save-side pipe escape: protect a `|` inside a code span / Go
    // action in a table cell BEFORE marked's table tokenizer splits the row (it splits raw `|` before inline
    // tokenization, dropping trailing cells). @tiptap/markdown's manager builds its block lexer via
    // `new markedInstance.Lexer(...)` — including for the construction-time initial parse — so subclassing
    // this PRIVATE instance's Lexer keeps the transform instance-scoped (no shared-class mutation) and still
    // catches every load.
    const BaseLexer = instance.Lexer;

    instance.Lexer = class extends BaseLexer {
        // Undo the paragraph serializer's line-leading `\#`/`\>` here rather than in an inline tokenizer or in
        // `lex`. An inline tokenizer has no notion of position and ate the backslash mid-line; `lex` runs
        // BEFORE block tokenization, so unescaping there hands marked a live `#`/`>` and the paragraph the
        // escape exists to protect becomes a heading or a quote again.
        override inlineTokens(src: string, tokens = []) {
            const unescaped = src
                .replace(/(^|\n)\\([#>])/g, '$1$2')
                // Only a backslash before a run that FILLS the line, matching what the serializer escapes —
                // an author's `\-` inside `[a-z\-_]` keeps its backslash.
                .replace(/(\n {0,3})\\((?:=+|-+) *)(?=\n|$)/g, '$1$2');

            return super.inlineTokens(unescaped, tokens);
        }

        override lex(src: string) {
            return super.lex(escapeTablePipes(src));
        }
    } as typeof BaseLexer;

    // A PRIVATE instance — mutating the shared global `marked` would also affect report-pdf.
    return instance;
};

type ManagerWithEncode = { encodeTextForMarkdown: (text: string, node?: SerializedNode, parent?: unknown) => string };
type SerializedNode = { marks?: (string | { type?: string })[]; text?: string };

const hasCodeMark = (node?: SerializedNode): boolean =>
    (node?.marks ?? []).some((mark) => (typeof mark === 'string' ? mark : mark.type) === 'code');

// Serialize a literal string as a CommonMark code span (spec 0.31.2 §6.1). @tiptap/markdown's mark path
// can't do this: it derives the code fence by rendering the mark against a placeholder, so the fence is a
// fixed single backtick blind to the content — a span whose text holds a backtick (`` `x` ``, common in
// pentest write-ups) serialises to invalid `` ``x`` `` and degrades further on every save. Choose a fence
// one longer than the longest backtick run inside, and pad with a space when the content starts or ends
// with a backtick (or is all-but-not-only spaces) so the reader strips exactly that pad back off. Mirrors
// prosemirror-markdown's backticksFor; the code-mark renderMarkdown override in markdown-editor-extensions.ts
// zeroes the placeholder fence so this is the ONLY fence emitted.
export const serializeCodeSpan = (content: string): string => {
    const longestRun = (content.match(/`+/g) ?? []).reduce((max, run) => Math.max(max, run.length), 0);
    const fence = '`'.repeat(longestRun + 1);
    const needsPad =
        content.startsWith('`') ||
        content.endsWith('`') ||
        (content.startsWith(' ') && content.endsWith(' ') && /\S/.test(content));

    return `${fence}${needsPad ? ` ${content} ` : content}${fence}`;
};

const TunedMarkdownText = Extension.create({
    name: 'tunedMarkdownText',
    onBeforeCreate() {
        // editor.markdown is always assigned by the Markdown extension (priority ordering below); the
        // non-optional cast makes a future regression throw here instead of silently reverting save to the
        // lossy default encoder.
        const manager = this.editor.markdown as unknown as ManagerWithEncode;

        // Replace the manager's encoder. Upstream HTML-entity-encodes (`<`→`&lt;`) and backslash-escapes
        // ``` ` * _ [ ] ~ \ ```; both are wrong here — the load side keeps `\`+punct literal and decodes named
        // entities, so re-applying either would double backslashes or re-freeze `&lt;`. Prose is therefore
        // emitted verbatim. The one exception is an inline code span: it needs a real CommonMark fence sized to
        // its content (see serializeCodeSpan). The code MARK is the discriminator, NOT the manager's broader
        // isInsideCode — a code BLOCK's body arrives here with empty marks (parent is codeBlock) and must stay
        // verbatim, its fence coming from TunedCodeBlock.
        manager.encodeTextForMarkdown = (text, node) => (hasCodeMark(node) ? serializeCodeSpan(text) : text);
    },
    // Lower priority than the Markdown extension so this runs AFTER its onBeforeCreate has created the
    // manager and assigned editor.markdown. onBeforeCreate (not onCreate) because it is synchronous —
    // a headless editor's onCreate fires after construction, too late for the first getMarkdown().
    priority: 50,
});

// marked's cell splitter honors `\|` only after an ODD run of backslashes (it counts them), never collapses
// `\\`, and truncates a row that splits into extra columns — so cell text holding an odd run + pipe (`a\|b`)
// has NO exact GFM encoding: escaping the pipe alone yields `\\|`, a live delimiter that drops the trailing
// cells on the next load. Pad an odd run by one backslash — the cell gains a `\`, the table keeps its cells,
// and the padded form is byte-stable from the first save.
// A pipe inside a Go action is the template's pipeline operator. Escaping it makes text/template reject the
// whole prompt (`unexpected "\" in operand`), so the file cannot be saved at all — and every save re-added the
// backslash, leaving no way out from rich mode. The loader counts cells with action pipes masked, so leaving
// them raw no longer skews the header/delimiter comparison.
export const escapeCellPipes = (text: string): string => {
    const escapeOutsideActions = (segment: string) =>
        segment.replace(/(\\*)\|/g, (_, run: string) => `${run}${run.length % 2 ? '\\\\|' : '\\|'}`);
    let result = '';
    let cursor = 0;

    for (const action of text.matchAll(TEMPLATE_ACTION)) {
        result += escapeOutsideActions(text.slice(cursor, action.index)) + action[0];
        cursor = action.index + action[0].length;
    }

    return result + escapeOutsideActions(text.slice(cursor));
};

// @tiptap/extension-table's renderTableToMarkdown is alignment-aware but never escapes pipes, so a literal
// `|` a cell emits (even from inside inline code) would re-parse as a column delimiter on the next SAVE and
// drop cells (tiptap PR #7884). renderTableToMarkdown emits cell content only via h.renderChildren, so wrapping
// that one call to escape pipes fixes the save side, on the official alignment-aware renderer.
// The LOAD side of the same class is covered by the tuned Lexer subclass in createTunedMarked above.
// GFM has no headerless table, and renderTableToMarkdown answers that by emitting an EMPTY header row above
// the demoted rows — so every header-off + save + reload cycle grew the table by one blank row and the header
// switch silently flipped back on. Promote the first row instead: the row count and every cell survive, and a
// second save is a no-op because the promoted table already has a header.
const withPromotedHeaderRow = (node: JSONContent): JSONContent => {
    const rows = node.content ?? [];
    const firstRow = rows[0];

    if (!firstRow?.content?.length || firstRow.content.some((cell) => cell.type === 'tableHeader')) {
        return node;
    }

    return {
        ...node,
        content: [
            { ...firstRow, content: firstRow.content.map((cell) => ({ ...cell, type: 'tableHeader' })) },
            ...rows.slice(1),
        ],
    };
};

export const TunedTable = Table.extend({
    renderMarkdown(node: JSONContent, helpers: MarkdownRendererHelpers) {
        const pipeEscaping: MarkdownRendererHelpers = {
            ...helpers,
            renderChildren: (nodes, separator) => escapeCellPipes(helpers.renderChildren(nodes, separator)),
        };

        // renderTableToMarkdown joins a multi-block cell's children (e.g. two paragraphs from Enter-in-cell) with
        // a U+001F separator, outside renderChildren. A GFM cell is single-line, so collapse it to a space rather
        // than persist a raw control byte that survives into the saved .tmpl/knowledge and round-trips unchanged.
        // Trim the renderer's own surrounding blank lines. It emits a leading and trailing newline on top of the
        // block separator the manager already adds, and between two ADJACENT tables that extra line reloads as an
        // empty paragraph — which serialises to another blank line on the next save. Measured before the trim:
        // two tables separated by one blank line gained one empty paragraph per cycle and never converged.
        return renderTableToMarkdown(withPromotedHeaderRow(node), pipeEscaping)
            .replaceAll(String.fromCharCode(0x1f), ' ')
            .replace(/^\n+/, '')
            .replace(/\n+$/, '');
    },
});

export const createMarkdownLayer = () => [
    Markdown.configure({ marked: createTunedMarked() as unknown as typeof import('marked').marked }),
    TunedMarkdownText,
];
