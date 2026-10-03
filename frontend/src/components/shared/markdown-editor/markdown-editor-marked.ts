import type { AnyExtension, JSONContent, MarkdownParseHelpers, MarkdownRendererHelpers } from '@tiptap/core';

import { Extension } from '@tiptap/core';
import { renderTableToMarkdown, Table } from '@tiptap/extension-table';
import { Markdown } from '@tiptap/markdown';
import { Marked, type Token } from 'marked';

import {
    annotateLayout,
    lineBreaksAtEnd,
    renderCompactTable,
    withLayoutFacts,
    withWrittenRows,
} from './markdown-editor-layout';
import {
    countRowCells,
    escapeTablePipes,
    findCodeSpanCloser,
    isNotARow,
    isTableBodyEnd,
    TABLE_DELIMITER_LINE,
    TABLE_PIPE,
    TEMPLATE_ACTION,
} from './markdown-editor-table-pipes';

// @tiptap/markdown parses with `marked`, which — unlike markdown-it's `html: false` — always tries to
// interpret `<...>` as HTML. Our content uses literal XML-ish tags (`<container_environment>`, `<input>`)
// that must survive verbatim; marked silently swallows the ones whose names match real HTML elements
// (`<input>`, `<br>`, …). Neutralising marked's block (`html`) and inline (`tag`) HTML tokenizers makes
// every `<...>` fall through to plain text, recreating markdown-it's `html: false`.
const createTunedMarked = () => {
    const instance = new Marked();
    const BaseTokenizer = instance.Tokenizer;
    const cellTokens = new WeakSet<object>();
    let isInsideTable = false;
    let definitionRuns = 0;
    // marked asks where the next literal span starts before every piece of text, with all that is left of the
    // paragraph. Each answer is kept as a distance from the end until the text has passed it — and for good
    // once there is none — or a paragraph with one placeholder and a thousand actions is searched to its end
    // a thousand times.
    let spansAhead: Record<string, null | number> = {};

    const nextSpan = (src: string, opener: string): number => {
        const known = spansAhead[opener];

        if (typeof known === 'number' && (known < 0 || src.length >= known)) {
            return known < 0 ? -1 : src.length - known;
        }

        const at = src.indexOf(opener);

        spansAhead[opener] = at < 0 ? -1 : src.length - at;

        return at;
    };

    let inlineDepth = 0;
    let inlineSource: null | string = null;

    instance.use({
        tokenizer: {
            // These marked tokenizers auto-convert literal text into markup, mangling Go-template / pentest
            // prose on round-trip. Returning `undefined` forces the char to stay literal text; `false` defers
            // to marked's default. Neutralise the lossy cases, keep what the toolbar emits:
            //   • del      — keep GFM `~~strike~~`, drop a lone `~…~` (else `~5~` / ranges become <del>)
            //   • emStrong — keep `*`/`**`, drop `_`-delimited emphasis (else `__init__`/`_word_` become em/strong),
            //                and emphasis nested in its own kind: a text node holds a mark once, so
            //                `**No /dev/sd*, /dev/vd*, /dev/fuse**` — italic inside italic to CommonMark — is
            //                saved one pair of asterisks short. Left literal, its opening asterisk is text
            //                and what follows pairs up into marks the document can hold.
            //   • escape   — keep `\`+punct literal (`\.` `\*` `\|` `\\`); marked's default DROPS the backslash
            //                (CommonMark unescape), silently corrupting regex/paths on the first load. The
            //                `\#`/`\>` counterpart to the paragraph serializer is NOT here: an inline tokenizer
            //                fires at any position, so it would also eat the backslash mid-line (`grep '\<root\>'`).
            //                It lives in the Lexer's inlineTokens override below, which sees line starts.
            //   • html/tag — keep `<xml-like>` tags literal (marked swallows real-HTML-element names)
            //   • def      — keep `[label]: target` lines as the text they are. marked lifts a definition out
            //                of the document into a link table and nothing writes it back, so a References
            //                section loses its lines and `[text][1]` is rewritten as an inline link. The
            //                run of definitions is one paragraph of literal text (LinkDefinitions below):
            //                tokenized inline, its target would be a link of its own.
            //   • autolink — `<scheme:…>` is a link for any scheme, so a Flask route `<int:order_id>` is one,
            //                shown and saved without its brackets. Only a web address or an e-mail is a
            //                link; anything else is the text it was written as, brackets and all — taken
            //                whole, or the bare-URL rule would find a link inside `<git+https://…>`.
            // NB: beyond the `autolink` case above and the `url` override below, autolink/url are intentionally
            // NOT neutralised — a bare `https://…`, `<url>` or email is meant to become a link (see
            // markdown-editor-extensions.ts link config, kept symmetric with typing).
            // NB: named HTML entities (`&lt; &gt; &amp; &quot;`) are decoded downstream — @tiptap/markdown's
            // token parsing runs @tiptap/core's decodeHtmlEntities, so a bare-prose `&lt;` becomes `<` (fixes
            // HTML-encoding artifacts from ingestion). Numeric refs (`&#123;`) and anything inside code are
            // untouched; a bare `&` survives as `&`. Don't re-add a literalAmpersand token to "preserve"
            // `&lt;` — that re-freezes the artifacts.
            autolink(src: string) {
                const link = BaseTokenizer.prototype.autolink.call(this, src);
                const isAddress = link && /^(?:https?:|ftp:|mailto:)/i.test(link.href);

                return link && !isAddress ? ({ raw: link.raw, text: link.raw, type: 'escape' } as never) : link;
            },
            codespan(src: string) {
                const span = BaseTokenizer.prototype.codespan.call(this, src);

                return span && { ...span, text: readCodeSpan(span.raw) };
            },
            def(src: string) {
                const definition = this.rules.block.def;
                let raw = '';

                // The rule takes the blank lines under a definition with it; the run ends at the first.
                for (let line = definition.exec(src); line; line = definition.exec(src.slice(raw.length))) {
                    raw += line[0];

                    if (line[0].endsWith('\n\n')) {
                        break;
                    }
                }

                // Not a `paragraph`: marked appends the lines under one to it and edits the inline source it
                // queued for it, and this run queues none. And a tag of its own: marked keeps a definition
                // only the first time it sees its tag.
                return raw
                    ? ({ raw, tag: `\0${definitionRuns++}`, text: raw.trimEnd(), type: 'linkDefinitions' } as never)
                    : undefined;
            },
            del: (src: string) => (/^~~(?!~)/.test(src) ? false : undefined),
            emStrong(src: string, maskedSrc: string, prevChar?: string) {
                if (src.startsWith('_')) {
                    return undefined;
                }

                const emphasis = BaseTokenizer.prototype.emStrong.call(this, src, maskedSrc, prevChar);

                return emphasis && nestsItsOwnKind(emphasis) ? undefined : emphasis;
            },
            escape: () => undefined,
            html: () => undefined,
            // marked hands each cell's text to `lexer.inline` from inside this tokenizer — the one place a
            // cell can be told from a paragraph line (see the Lexer's `inline` below).
            table(src: string) {
                isInsideTable = true;

                try {
                    return BaseTokenizer.prototype.table.call(this, withoutTemplateRows(src));
                } finally {
                    isInsideTable = false;
                }
            },
            tag: () => undefined,
            // A URL that runs into a Go action — `https://host/{{ .Path }}` — would be a link that ends inside
            // the action: edited, it is written as `[…{{](…{{) .Path }}`, which no longer parses.
            url(src: string) {
                const link = BaseTokenizer.prototype.url.call(this, src);

                return link?.raw.includes('{{') ? undefined : link;
            },
        },
    });

    // Two things are literal text whatever markdown would read in them. The anonymizer replaces a secret with
    // `§*Name*§`, and a second pass over its own output yields `§***Name****§`: its asterisks are not
    // emphasis, and read as emphasis they pair with asterisks outside the placeholder and come back fewer. A
    // Go template action is the template's: a raw string between backticks is not a code span, a URL in a
    // string is not a link, and `&lt;` in one is not `<` — decoded, it is another string, and `&quot;` a
    // parse error. Each is taken whole here — as an `escape` token, the one kind @tiptap/markdown copies
    // without decoding entities — and masked from the scan that looks for the other end of an emphasis run.
    // That scan measures positions by length, so the mask is as long as what it hides; and marked has already
    // blotted out `\§` in the string it hands the hook, so the spans are found in the source the Lexer saw.
    instance.use({
        extensions: [
            {
                level: 'inline',
                name: 'literalSpan',
                start(src: string) {
                    const placeholder = nextSpan(src, '§*');
                    const action = nextSpan(src, '{{');

                    return placeholder < 0 || action < 0
                        ? Math.max(placeholder, action)
                        : Math.min(placeholder, action);
                },
                tokenizer(src: string) {
                    const [raw] = LITERAL_SPAN_AT_START.exec(src) ?? [];

                    return raw ? { raw, text: raw, type: 'escape' } : undefined;
                },
            },
        ],
        hooks: {
            emStrongMask(masked: string) {
                const source = inlineSource?.length === masked.length ? inlineSource : masked;
                const hidden: string[] = [];
                let at = 0;

                inlineSource = null;

                for (const { 0: span, index } of source.matchAll(LITERAL_SPAN)) {
                    // What marked has blotted out already — `{{` inside a code span — opens no span. A `+` is
                    // its mark for an escaped character, which the span still begins with.
                    if (masked[index] !== span[0] && masked[index] !== '+') {
                        continue;
                    }

                    hidden.push(masked.slice(at, index), span[0]!, 'a'.repeat(span.length - 2), span.at(-1)!);
                    at = index + span.length;
                }

                return at ? hidden.join('') + masked.slice(at) : masked;
            },
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
        // Set for the content of a task item, which @tiptap/markdown lexes a second time and renders itself:
        // it gets the pipe protection and none of the layout its renderer would not honour.
        isNested = false;

        override inline(src: string, tokens: Token[] = []) {
            if (isInsideTable) {
                cellTokens.add(tokens);
            }

            return super.inline(src, tokens);
        }

        // Undo the paragraph serializer's line-leading `\#`/`\>` here rather than in an inline tokenizer or in
        // `lex`. An inline tokenizer has no notion of position and would eat the backslash mid-line; `lex` runs
        // BEFORE block tokenization, so unescaping there hands marked a live `#`/`>` and the paragraph the
        // escape exists to protect becomes a heading or a quote again.
        // A table cell is left alone: it cannot open a heading or a quote, cellText never escapes one, and a
        // backslash found there is the author's — `\>` in a column of regex tokens. So is the content of
        // emphasis, a link or a strike-through, which marked lexes with this method as well: its first
        // character is not the start of a line, and the paragraph's own call has already seen every line of
        // it.
        override inlineTokens(src: string, tokens: Token[] = []) {
            const isParagraph = !inlineDepth && !cellTokens.has(tokens);
            const source = isParagraph ? withoutLineEscapes(src) : src;

            const outer = spansAhead;

            spansAhead = {};
            inlineSource = source;
            inlineDepth += 1;

            try {
                super.inlineTokens(source, tokens);
            } finally {
                inlineDepth -= 1;
                spansAhead = outer;
            }

            // marked fills the array it was handed, so the line breaks are split out in place — one by one: a
            // paragraph of fifty thousand lines is more arguments than a call takes. A literal span and the
            // plain text beside it go back as one token: @tiptap/markdown merges the text nodes it makes of
            // neighbours by splicing one out of the array at a time, the square of a paragraph of actions.
            const split = withLineBreaks(tokens);

            tokens.length = 0;

            for (const token of split) {
                const last = tokens.at(-1);

                if (
                    last &&
                    isLiteral(last) &&
                    isLiteral(token) &&
                    (last.type === 'escape' || token.type === 'escape')
                ) {
                    const raw = last.raw + token.raw;

                    tokens[tokens.length - 1] = { raw, text: raw, type: 'escape' } as Token;
                } else {
                    tokens.push(token);
                }
            }

            return tokens;
        }

        override lex(src: string) {
            // A document that holds the stand-in itself gets the backslash, which marked honours as well.
            const standIn = src.includes(TABLE_PIPE) ? undefined : TABLE_PIPE;
            const escaped = escapeTablePipes(src, standIn);
            const shielded = shieldFenceTabs(escaped);
            const tokens = super.lex(shielded);

            if (shielded !== escaped) {
                restoreShielded(tokens, FENCE_TAB, '\t');
            }

            if (!this.isNested) {
                this.restoreItemLines(tokens);
                annotateLayout(tokens, src);
            }

            // After the layout is read: the style of a table is picked from rows in which a pipe of a code span
            // or of an action is still a stand-in, and no separator.
            if (standIn && escaped !== src) {
                restoreShielded(tokens, standIn, '|');
            }

            return tokens;
        }

        // Inside a list item marked reads a line indented four past the content column as indented code and
        // then merges it into the text above: without its indentation, and with the code's own line break,
        // so the paragraph gains an empty line and is saved as two. The lines as they were written are still
        // in `raw` — except a line that opens with a tab, which marked reads as four spaces the next time
        // and so would not come back the same.
        restoreItemLines(tokens: Token[], isInItem = false) {
            for (const token of tokens as (Token & { items?: Token[]; text?: string; tokens?: Token[] })[]) {
                const isText = isInItem && /^(?:text|paragraph)$/.test(token.type);
                const written = isText ? token.raw.slice(0, token.raw.length - lineBreaksAtEnd(token.raw).length) : '';

                if (isText && token.text !== written && !written.includes('\n\t')) {
                    token.text = written;
                    token.tokens = this.inlineTokens(written);
                } else if (token.type === 'list' || token.type === 'blockquote') {
                    this.restoreItemLines(token.items ?? token.tokens ?? []);
                } else if (token.type === 'list_item') {
                    this.restoreItemLines(token.tokens ?? [], true);
                }
            }
        }
    } as typeof BaseLexer;

    // @tiptap/markdown lexes the content of a task item again through `instance.lexer`, which is the base
    // class's: without the pre-pass a pipe in a code span of a table nested there splits its cell.
    instance.lexer = ((src: string, options?: typeof instance.defaults) => {
        const lexer = new instance.Lexer(options ?? instance.defaults) as InstanceType<typeof instance.Lexer> & {
            isNested: boolean;
        };

        lexer.isNested = true;

        return lexer.lex(src);
    }) as typeof instance.lexer;

    // A PRIVATE instance — mutating the shared global `marked` would also affect report-pdf.
    return instance;
};

// What the paragraph serializer puts a backslash before, without it: `#` and `>` at the start of a line, and a
// run of `=` or `-` that FILLS a line under another — an author's `\-` inside `[a-z\-_]` keeps its backslash.
const withoutLineEscapes = (text: string): string =>
    text.replace(/(^|\n)( {0,3})\\([#>])/g, '$1$2$3').replace(/(\n {0,3})\\((?:=+|-+) *)(?=\n|$)/g, '$1$2');

// An anonymizer placeholder, or a Go template action.
const LITERAL_SPAN = new RegExp(`§\\*[^§\\n]*\\*§|${TEMPLATE_ACTION.source}`, 'g');
const LITERAL_SPAN_AT_START = new RegExp(`^(?:${LITERAL_SPAN.source})`);

// What @tiptap/markdown copies into a text node as it is: an `escape` token, and text that holds no entity.
const isLiteral = (token: Token): boolean =>
    token.type === 'escape' ||
    (token.type === 'text' &&
        !('tokens' in token && token.tokens) &&
        token.text === token.raw &&
        !token.raw.includes('&'));

// The rows of a table, up to the first line that is not one (isNotARow). What is tested first is the delimiter
// row: without one marked reads no table, and the lines below are left unread — this runs at the start of
// every block, and reading on to the next blank line each time is quadratic.
const withoutTemplateRows = (src: string): string => {
    const header = src.indexOf('\n');
    const delimiter = src.indexOf('\n', header + 1);

    const row = src.slice(header + 1, delimiter);

    if (header < 0 || delimiter < 0 || !TABLE_DELIMITER_LINE.test(row)) {
        return src;
    }

    const columns = countRowCells(row);

    for (let at = delimiter + 1; at < src.length; ) {
        const end = src.indexOf('\n', at);
        const line = src.slice(at, end < 0 ? src.length : end);

        if (isTableBodyEnd(line)) {
            return src;
        }

        if (isNotARow(line, columns)) {
            return src.slice(0, at);
        }

        at = end < 0 ? src.length : end + 1;
    }

    return src;
};

// A line break inside a paragraph is a character of its text to marked. Left in a text node it does not
// survive an edit: ProseMirror reads the paragraph back from the DOM on every keystroke and turns each line
// break it finds in text into a hard break, so typing one letter gives every line of the paragraph two
// trailing spaces. As a token of its own it becomes a break node that remembers it was written as nothing.
// A paragraph has no empty line and does not end in a break; marked's merge of a definition or an indented
// line into the text above leaves both behind, and they are dropped.
const isLineBreak = (token?: Token): boolean => token?.type === 'br' && token.raw === '\n';

const withLineBreaks = (tokens: Token[]): Token[] => {
    const split: Token[] = [];

    for (const token of tokens) {
        const isText = token.type === 'text' || (token.type === 'escape' && token.raw.startsWith('{{'));

        if (!isText || ('tokens' in token && token.tokens) || !token.raw.includes('\n')) {
            split.push(token);
            continue;
        }

        token.raw.split('\n').forEach((line, at) => {
            if (at && !isLineBreak(split.at(-1))) {
                split.push({ raw: '\n', type: 'br' } as Token);
            }

            if (line) {
                split.push({ ...token, raw: line, text: line } as Token);
            }
        });
    }

    while (isLineBreak(split.at(-1))) {
        split.pop();
    }

    return split;
};

const nestsItsOwnKind = (emphasis: Token): boolean => {
    const holds = (tokens: Token[] = []): boolean =>
        tokens.some((token) => token.type === emphasis.type || holds((token as { tokens?: Token[] }).tokens));

    return holds((emphasis as { tokens?: Token[] }).tokens);
};

// marked's list rule replaces every tab on an item's continuation lines with four spaces before it dedents
// them, fenced code included, so a Makefile recipe or gofmt-indented Go under a list item loses its tabs. A
// tab past the fence's own indentation carries no structure, so it crosses the lexer as a private-use
// character and comes back in the tokens. A fence that never closes is left alone: marked ends it with the
// item, and tabs shielded past that point would stop counting as indentation.
const FENCE_TAB = '\uE000';
const FENCE_RUN = /^([ \t>]*)(`{3,}(?=[^`\n]*$)|~{3,})/;

const shieldFenceTabs = (markdown: string): string => {
    if (!markdown.includes('\t') || markdown.includes(FENCE_TAB)) {
        return markdown;
    }

    const lines = markdown.split('\n');
    let open: null | { at: number; indent: number; run: string } = null;

    lines.forEach((line, index) => {
        const fence = FENCE_RUN.exec(line);

        if (!open) {
            open = fence ? { at: index, indent: fence[1]!.length, run: fence[2]! } : null;

            return;
        }

        const closes =
            fence &&
            fence[2]![0] === open.run[0] &&
            fence[2]!.length >= open.run.length &&
            !line.slice(fence[0].length).trim();

        if (!closes) {
            return;
        }

        for (let body = open.at + 1; body < index; body++) {
            const text = lines[body]!;
            const structural = Math.min(open.indent, /^[ \t>]*/.exec(text)![0].length);

            lines[body] = text.slice(0, structural) + text.slice(structural).replaceAll('\t', FENCE_TAB);
        }

        open = null;
    });

    return lines.join('\n');
};

const restoreShielded = (value: unknown, shield: string, written: string): void => {
    if (!value || typeof value !== 'object') {
        return;
    }

    const fields = value as Record<string, unknown>;

    for (const key of Object.keys(fields)) {
        const field = fields[key];

        if (typeof field === 'string') {
            fields[key] = field.replaceAll(shield, written);
        } else {
            restoreShielded(field, shield, written);
        }
    }
};

type ManagerWithEncode = { encodeTextForMarkdown: (text: string, node?: SerializedNode, parent?: unknown) => string };
type SerializedNode = { marks?: (string | { attrs?: Record<string, unknown>; type?: string })[]; text?: string };

const codeMarkOf = (node?: SerializedNode) =>
    (node?.marks ?? []).find((mark) => (typeof mark === 'string' ? mark : mark.type) === 'code');

// A code span is written between fences one backtick longer than the longest run inside it, or between the
// ones it was written with (`written`) while no run inside is exactly that long — one backtick around
// `` a``b `` closes only at one backtick. @tiptap/markdown's mark path cannot do either: it derives the fence
// by rendering the mark against a placeholder, so a span that holds a backtick (`` `x` ``, common in pentest
// write-ups) serialises to invalid `` ``x`` `` and degrades on every save. The code-mark renderMarkdown
// override in markdown-editor-extensions.ts zeroes the placeholder fence so this is the ONLY fence emitted.
// A span whose content begins or ends with a backtick needs a space between it and the fence. That space is
// the only padding the load side takes back off (readCodeSpan), so the spaces of `` ` a ` `` are content and
// are written as they stand.
const touchesBacktick = (content: string): boolean => /^`|`$/.test(content.trim());

// A span may run over a line, and is written that way while its later lines cannot be read as anything else:
// a fence of three backticks at a line's start opens a code block, and a line that begins like a list item, a
// heading, a quote, a table row or a rule — or is indented like code — starts that block inside a list
// item. Otherwise its line breaks are written as the spaces they read as.
const CONTINUES_SPAN = /^ {0,3}(?![-+*>#=|~`[]|\d{1,9}[.)]|[_*-](?: *[_*-]){2,} *$)\S/;

const keepsItsLines = (content: string, fence: string): boolean =>
    fence.length < 3 &&
    content
        .split('\n')
        .slice(1)
        .every((line, at, lines) => CONTINUES_SPAN.test(line) || (!line && at === lines.length - 1));

export const serializeCodeSpan = (text: string, written?: unknown): string => {
    const runs = (text.match(/`+/g) ?? []).map((run) => run.length);
    const own = '`'.repeat(Math.max(0, ...runs) + 1);
    const length = typeof written === 'number' && Number.isInteger(written) ? written : 0;
    const kept = length > 0 && length <= 64 && !runs.includes(length) ? '`'.repeat(length) : own;
    // A fence that cannot hold the span's lines gives way to one that can.
    const fence = !text.includes('\n') || keepsItsLines(text, kept) ? kept : own;
    const content = keepsItsLines(text, fence) ? text : text.replaceAll('\n', ' ');

    return `${fence}${touchesBacktick(content) ? ` ${content} ` : content}${fence}`;
};

// marked strips one space from each end of every padded span, so `` `` alert(`XSS`) `` `` is saved without
// its spaces, and an agent's unbalanced `` ` capture the ` `` loses both and glues the words around it
// together; and it turns the span's line breaks into spaces.
const readCodeSpan = (raw: string): string => {
    const fence = /^`+/.exec(raw)![0].length;
    const content = raw.slice(fence, -fence);
    const isPadded = content.length > 2 && content.startsWith(' ') && content.endsWith(' ');

    return isPadded && touchesBacktick(content) ? content.slice(1, -1) : content;
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
        manager.encodeTextForMarkdown = (text, node) => {
            const code = codeMarkOf(node);

            return code ? serializeCodeSpan(text, typeof code === 'string' ? undefined : code.attrs?.fence) : text;
        };
    },
    // Lower priority than the Markdown extension so this runs AFTER its onBeforeCreate has created the
    // manager and assigned editor.markdown. onBeforeCreate (not onCreate) because it is synchronous —
    // a headless editor's onCreate fires after construction, too late for the first getMarkdown().
    priority: 50,
});

type InlineToken = Token & { href?: string; text?: string; tokens?: InlineToken[] };

let reader: ReturnType<typeof createTunedMarked> | undefined;

const textOf = (tokens: InlineToken[]): string =>
    tokens
        .map((token) => (token.tokens ? textOf(token.tokens) : token.type === 'br' ? '\n' : (token.text ?? '')))
        .join('');

// A setext heading is what its lines happen to read as: one that opens a block, or is an underline itself,
// ends the heading above it.
export const readsAsHeading = (markdown: string, level: number): boolean => {
    const { defaults, Lexer } = (reader ??= createTunedMarked());
    const [heading, ...rest] = new Lexer(defaults).lex(markdown) as Array<Token & { depth?: number }>;

    return heading?.type === 'heading' && heading.depth === level && !rest.length;
};

// Whether a line of inline markdown reads as one link to `href`, under `title`, around nothing but one image.
export const readsAsLinkedImage = (markdown: string, href: string, title: null | string): boolean => {
    const { defaults, Lexer } = (reader ??= createTunedMarked());
    const [link] = new Lexer(defaults).inlineTokens(markdown) as (InlineToken & { title?: null | string })[];

    return (
        link?.type === 'link' &&
        link.href === href &&
        (link.title || null) === title &&
        link.tokens?.length === 1 &&
        link.tokens[0]!.type === 'image'
    );
};

// Every link a line of inline markdown reads as: its target and its text.
const readLinks = (markdown: string): string => {
    const { defaults, Lexer } = (reader ??= createTunedMarked());
    // A link around an image is the image's own (its `link` attribute), not a link of the text.
    const collect = (tokens: InlineToken[]): string[] =>
        tokens.flatMap((token) => {
            if (token.type !== 'link') {
                return collect(token.tokens ?? []);
            }

            const text = textOf((token.tokens ?? []).filter((child) => child.type !== 'image'));

            return text ? [`${token.href}\0${text}`] : [];
        });

    return collect(new Lexer(defaults).inlineTokens(markdown)).join('\0\0');
};

const linkOf = (node?: JSONContent) =>
    node?.type === 'text' ? node.marks?.find((mark) => mark.type === 'link') : undefined;

const markTypesOf = (node: JSONContent): string =>
    (node.marks ?? [])
        .map((mark) => mark.type)
        .sort()
        .join();

const stylesOf = (node?: JSONContent): string =>
    (node?.marks ?? [])
        .map((mark) => mark.type)
        .filter((type) => type !== 'link')
        .sort()
        .join();

const isSoftBreak = (node?: JSONContent): boolean => node?.type === 'hardBreak' && node.attrs?.marker === '';
const marksKey = (node?: JSONContent): string => JSON.stringify(node?.marks ?? []);

const codeOf = (node: JSONContent) => node.marks?.find((mark) => mark.type === 'code');

// Code that an edit has put beside a span written with a fence of its own is one span with it and is fenced
// once: two fences back to back are read as one longer fence. The fence a span was written with is kept only
// where every piece of the span carries it.
const withOneFence = (content: JSONContent[]): JSONContent[] => {
    const result = [...content];

    for (let from = 0, to = 1; from < result.length; from = to++) {
        if (!codeOf(result[from]!)) {
            continue;
        }

        while (to < result.length && codeOf(result[to]!)) {
            to++;
        }

        const fences = new Set(result.slice(from, to).map((node) => codeOf(node)?.attrs?.fence ?? null));

        for (let at = from; fences.size > 1 && at < to; at++) {
            result[at] = {
                ...result[at],
                marks: result[at]!.marks?.map((mark) => (mark.type === 'code' ? { type: 'code' } : mark)),
            };
        }
    }

    return result;
};

// The line breaks of a paragraph as characters of its text again, which is how they are written. A break
// keeps the marks it was read under (withMarkedBreaks), and text under the same marks is one node: emphasis
// that runs over a line is not closed at its end and opened on the next, and a code span is fenced once.
// A break is written only between two lines of text. Enter at the end of a wrapped line leaves one at the
// edge of a block, and deleting a line leaves two in a row or one beside a hard break: written, each would
// be an empty line, which ends the paragraph. Inside a code span a break is content, wherever it stands.
const withLineBreakText = (nodes: JSONContent[]): JSONContent[] => {
    const content = withOneFence(nodes);
    const merged: JSONContent[] = [];
    let line: JSONContent | undefined;
    let above: JSONContent | undefined;

    for (const node of content) {
        if (isSoftBreak(node) && !markTypesOf(node).includes('code')) {
            line ??= node;
            continue;
        }

        const written = isSoftBreak(node) ? { marks: node.marks, text: '\n', type: 'text' } : node;
        const nodes =
            line && above && above.type !== 'hardBreak' && node.type !== 'hardBreak'
                ? [{ text: '\n', type: 'text', ...(line.marks ? { marks: line.marks } : {}) }, written]
                : [written];

        for (const text of nodes) {
            const last = merged.at(-1);

            if (text.type === 'text' && last?.type === 'text' && marksKey(last) === marksKey(text)) {
                merged[merged.length - 1] = { ...last, text: `${last.text}${text.text}` };
            } else {
                merged.push(text);
            }
        }

        line = undefined;
        above = node;
    }

    return merged;
};

type MarkResult = { attrs?: Record<string, unknown>; content: JSONContent[]; mark: string };
type ParseMark = (this: unknown, token: unknown, helpers: unknown) => MarkResult | undefined;

// @tiptap/markdown puts a mark on the text nodes it covers and on nothing else, so a break inside `*a\nb*`
// would be the same node as the one between `*a*` and `*b*`. Each mark is added to the breaks it covers
// here, which is also where ProseMirror puts it when it reads the paragraph back from the DOM.
export const withMarkedBreaks = <T extends AnyExtension>(extension: T): T => {
    const parse = (extension.config as { parseMarkdown?: ParseMark }).parseMarkdown;

    if (extension.type !== 'mark' || !parse) {
        return extension;
    }

    return extension.extend({
        parseMarkdown(token: unknown, helpers: unknown) {
            const parsed = parse.call(this, token, helpers);

            if (!parsed?.mark) {
                return parsed;
            }

            const { attrs, mark: type } = parsed;
            const mark = { type, ...(attrs && Object.keys(attrs).length ? { attrs } : {}) };

            return {
                ...parsed,
                content: parsed.content.map((node) =>
                    isSoftBreak(node) ? { ...node, marks: [...(node.marks ?? []), mark] } : node,
                ),
            };
        },
    }) as T;
};

// A URL or an address written bare, or between angle brackets, is a link by itself. Saved as `[url](url)` it
// is not what was written, and for some URLs not the same link either. But where a bare URL ends is decided
// by the text around it, and an edit can move that: so the line is written with those links in their own
// form, and that is kept only if it reads as the very links the document holds. A link whose text is not
// its target — an entity decoded in it, a word typed into it — is not one of them, and does not cost the
// other links of the line their form.
export const renderInline = (nodes: JSONContent[], helpers: MarkdownRendererHelpers): string => {
    const content = withLineBreakText(nodes);
    const links: string[] = [];
    const written: JSONContent[] = [];
    let hasOwnForm = false;

    for (let at = 0, end = 1; at < content.length; at = end++) {
        const link = linkOf(content[at]);

        while (link && linkOf(content[end])?.attrs?.href === link.attrs?.href) {
            end++;
        }

        const run = content.slice(at, end);
        const text = run.map((node) => node.text ?? '').join('');
        const { form, href, title } = link?.attrs ?? {};
        const isOwnTarget = href === text || href === `mailto:${text}` || href === `http://${text}`;
        const isPlain = run.every(
            (node) => markTypesOf(node) === markTypesOf(run[0]!) && !markTypesOf(node).includes('code'),
        );
        // Bold or emphasis that begins or ends at the link has its delimiter between the link and the text
        // beside it, where only a space lets marked read it as one.
        const [before, after] = [content[at - 1], content[end]];
        const isAgainstMark =
            (before !== undefined && stylesOf(before) !== stylesOf(run[0]) && /\S/.test(before.text?.at(-1) ?? ' ')) ||
            (after !== undefined && stylesOf(after) !== stylesOf(run[0]) && /\S/.test(after.text?.[0] ?? ' '));

        if (link) {
            links.push(`${href}\0${text}`);
        }

        if ((form === 'angle' || form === 'bare') && !title && isOwnTarget && isPlain && !isAgainstMark) {
            hasOwnForm = true;
            written.push({
                ...run[0],
                marks: run[0]!.marks?.filter((mark) => mark.type !== 'link'),
                text: form === 'angle' ? `<${text}>` : text,
            });
        } else {
            written.push(...run);
        }
    }

    const candidate = hasOwnForm ? helpers.renderChildren(written) : '';

    return hasOwnForm && readLinks(candidate) === links.join('\0\0') ? candidate : helpers.renderChildren(content);
};

// marked's cell splitter honors `\|` only after an ODD run of backslashes (it counts them), never collapses
// `\\`, and truncates a row that splits into extra columns — so cell text holding an odd run + pipe (`a\|b`)
// has NO exact GFM encoding: escaping the pipe alone yields `\\|`, a live delimiter that drops the trailing
// cells on the next load. Pad an odd run by one backslash — the cell gains a `\`, the table keeps its cells,
// and the padded form is byte-stable from the first save.
// A pipe inside a Go action is the template's pipeline operator. Escaping it makes text/template reject the
// whole prompt (`unexpected "\" in operand`), so the file cannot be saved at all — and every save would re-add the
// backslash, leaving no way out from rich mode. The loader counts cells with action pipes masked, so leaving
// them raw does not skew the header/delimiter comparison.
// A pipe inside a code span is escaped as well, as GFM asks. Whichever opens first is what marked reads
// there: an action that holds a raw string is an action, a code span that holds an action is code, and the
// action inside it still keeps its pipes. This is how a row is written once it has been edited; one that
// has not is saved as the line it was (withWrittenRows), with its pipes the way its author put them.
export const escapeCellPipes = (text: string): string => {
    const escapePipes = (segment: string) =>
        segment.replace(/(\\*)\|/g, (_, run: string) => `${run}${run.length % 2 ? '\\\\|' : '\\|'}`);

    const escapeOutsideActions = (segment: string): string => {
        let result = '';
        let cursor = 0;

        for (const action of segment.matchAll(TEMPLATE_ACTION)) {
            result += escapePipes(segment.slice(cursor, action.index)) + action[0];
            cursor = action.index + action[0].length;
        }

        return result + escapePipes(segment.slice(cursor));
    };

    const actions = new RegExp(TEMPLATE_ACTION.source, 'g');
    let action = actions.exec(text);
    let result = '';
    let cursor = 0;

    for (let at = 0; at < text.length; ) {
        if (action && action.index < at) {
            actions.lastIndex = at;
            action = actions.exec(text);
        }

        const tick = text.indexOf('`', at);

        if (!action && tick < 0) {
            break;
        }

        if (action && (tick < 0 || action.index < tick)) {
            result += escapePipes(text.slice(cursor, action.index)) + action[0];
            cursor = at = action.index + action[0].length;
            continue;
        }

        let runEnd = tick;

        while (text[runEnd] === '`') {
            runEnd++;
        }

        const closer = findCodeSpanCloser(text, runEnd, runEnd - tick);

        // An unclosed backtick run is literal content, not a code span.
        if (closer < 0) {
            at = runEnd;
            continue;
        }

        const code = text.slice(runEnd, closer);

        result += escapePipes(text.slice(cursor, runEnd)) + escapeOutsideActions(code);
        cursor = closer;
        at = closer + runEnd - tick;
    }

    return result + escapePipes(text.slice(cursor));
};

// GFM has no headerless table, and renderTableToMarkdown answers that by emitting an EMPTY header row above
// the demoted rows — so every header-off + save + reload cycle would grow the table by one blank row and the header
// switch would silently flip back on. Promote the first row instead: the row count and every cell survive, and a
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

// A GFM cell is one line. renderTableToMarkdown writes a break inside one as `<br>`, which this editor — where
// a tag is literal text — reloads as four characters. So each cell is rendered here, with a break collapsed to
// a space, and handed over as plain text. A paragraph's inline content is rendered directly, past the paragraph
// serializer: a cell cannot open a heading or a quote, and its line-leading escape would save `| # | Name |`
// as `| \# | Name |`. escapeCellPipes is the save side of the pipe protection the tuned Lexer gives the load.
const cellText = (cell: JSONContent, helpers: MarkdownRendererHelpers): string =>
    escapeCellPipes(
        (cell.content ?? [])
            .map((block) =>
                block.type === 'paragraph'
                    ? renderInline(block.content ?? [], helpers)
                    : helpers.renderChildren([block]),
            )
            .join(' ')
            .replace(/[ \t]*\r?\n[ \t]*/g, ' '),
    );

const withSingleLineCells = (node: JSONContent, helpers: MarkdownRendererHelpers): JSONContent => ({
    ...node,
    content: (node.content ?? []).map((row) => ({
        ...row,
        content: (row.content ?? []).map((cell) => ({
            ...cell,
            content: [{ text: cellText(cell, helpers), type: 'text' }],
        })),
    })),
});

// `extend({ markdownTokenizer: undefined })` reads as "not overridden" and falls through to the parent's
// tokenizer, so the absence has to be a value.
export const withoutMarkdownTokenizer = <T extends AnyExtension>(extension: T): T =>
    extension.extend({ markdownTokenizer: null as unknown as undefined }) as T;

// @tiptap/extension-table tokenizes tables itself to protect a pipe inside a code span. The tuned Lexer above
// already does that, for Go actions too, and the two together cost content: a row-like block whose delimiter
// row has the wrong cell count is shredded into single characters, and a list item right after a table loses
// its text when it holds a pipe. marked's own table rule, fed by our pre-pass, has neither problem.
export const TunedTable = withoutMarkdownTokenizer(Table).extend({
    renderMarkdown(node: JSONContent, helpers: MarkdownRendererHelpers) {
        const verbatim: MarkdownRendererHelpers = {
            ...helpers,
            renderChildren: (nodes) => [nodes].flat().reduce((text, child) => text + (child.text ?? ''), ''),
        };

        const table = withSingleLineCells(withPromotedHeaderRow(node), helpers);
        // Trim the renderer's own surrounding blank lines. It emits a leading and trailing newline on top of the
        // block separator the manager already adds, and between two ADJACENT tables that extra line reloads as an
        // empty paragraph — which serialises to another blank line on the next save, so two tables separated by
        // one blank line gain an empty paragraph per cycle and never converge.
        const rendered =
            typeof node.attrs?.delimiterRow === 'string'
                ? renderCompactTable(table, node.attrs.delimiterRow, node.attrs.emptyCell)
                : renderTableToMarkdown(table, verbatim).replace(/^\n+/, '').replace(/\n+$/, '');

        return withWrittenRows(rendered.split('\n'), table, node.attrs?.written);
    },
});

// The token the tuned `def` tokenizer makes of a run of `[label]: target` lines: a paragraph of literal text,
// its lines apart the way a paragraph's are (withLineBreaks), and written with a paragraph's escapes — a
// target on a line of its own may be `---`.
const LinkDefinitions = Extension.create({
    name: 'linkDefinitions',
    parseMarkdown: (token: { text?: string }, helpers: MarkdownParseHelpers) => ({
        content: withoutLineEscapes(token.text ?? '')
            .split('\n')
            .flatMap((line, at) => [
                ...(at ? [{ attrs: { marker: '' }, type: 'hardBreak' }] : []),
                ...(line ? [helpers.createTextNode(line)] : []),
            ]),
        type: 'paragraph',
    }),
});

export const createMarkdownLayer = () => [
    Markdown.configure({ marked: createTunedMarked() as unknown as typeof import('marked').marked }),
    TunedMarkdownText,
    withLayoutFacts(LinkDefinitions),
];
