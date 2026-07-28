import type {
    AnyExtension,
    ChainedCommands,
    JSONContent,
    MarkdownParseHelpers,
    MarkdownParseResult,
    MarkdownRendererHelpers,
    MarkdownToken,
} from '@tiptap/core';

import { mergeAttributes } from '@tiptap/core';
import { CodeBlockLowlight } from '@tiptap/extension-code-block-lowlight';
import { Image } from '@tiptap/extension-image';
import { TaskItem, TaskList } from '@tiptap/extension-list';
import { TableCell, TableHeader, TableRow } from '@tiptap/extension-table';
import { Placeholder } from '@tiptap/extensions';
import { AllSelection, TextSelection } from '@tiptap/pm/state';
import StarterKit from '@tiptap/starter-kit';
import { common, createLowlight } from 'lowlight';

import { HeadingAutoformat } from './markdown-editor-heading-autoformat';
import { createMarkdownLayer, TunedTable } from './markdown-editor-marked';
import { MarkdownPaste } from './markdown-editor-paste';
import { TagHighlight } from './markdown-editor-tag-highlight';
import { VariableHighlight } from './markdown-editor-variable-highlight';

const dropUnderscoreRules = (rules: { find: unknown }[]) =>
    rules.filter((rule) => !(rule.find instanceof RegExp && rule.find.source.includes('_')));

type CommandMap = Record<string, ToggleCommand | undefined>;
type ToggleCommand = (...args: never[]) => (props: never) => boolean;

// Under Ctrl+A a block toggle wraps a second time instead of unwrapping, for two independent reasons:
// TrailingNode's empty paragraph sits inside the selection but outside the wrapper, so `isNodeActive` — which
// demands the type cover the WHOLE selection — reads false; and AllSelection anchors at depth 0, so
// `blockRange` resolves to the doc and `liftTarget` is null, making the unwrap a silent no-op.
//
// Excluding the trailing paragraph is what fixes both. Trimming to `1 … size - 1` is NOT enough: the
// paragraph stays inside the range and `isNodeActive` still reads false.
// Upstream https://github.com/ueberdosis/tiptap/issues/7398 covers only the two list cases — drop this once
// it lands, after checking quote/fence/task and the lift no-op are covered too.
const withSelectAllRetargeted = <T extends AnyExtension>(extension: T, commandName: string): T =>
    extension.extend({
        addCommands() {
            const parent = (this.parent?.() ?? {}) as CommandMap;
            const original = parent[commandName];

            if (!original) {
                return parent;
            }

            return {
                ...parent,
                [commandName]:
                    (...args: never[]) =>
                    ({ chain }: { chain: () => ChainedCommands }) =>
                        chain()
                            .command(({ state, tr }) => {
                                if (!(state.selection instanceof AllSelection)) {
                                    return true;
                                }

                                const last = tr.doc.lastChild;
                                const trailing =
                                    last && last.type.name === 'paragraph' && last.content.size === 0
                                        ? last.nodeSize
                                        : 0;
                                const to = Math.max(1, tr.doc.content.size - trailing - 1);

                                tr.setSelection(TextSelection.create(tr.doc, 1, to));

                                return true;
                            })
                            .command((props) => original(...args)(props as never))
                            .run(),
            };
        },
    }) as T;

type MarkdownTokenizer = {
    start?: (src: string) => number;
    tokenize: (src: string, tokens: unknown[], lexer: unknown) => unknown;
};

// marked runs every block extension's `tokenize` at EVERY block boundary, handing it the whole remaining
// document, and @tiptap/extension-list's ordered/task tokenizers open with `src.split('\n')` before bailing
// on a first line that isn't a list item — so each boundary materialises all remaining lines, O(n²).
//
// The gate is the tokenizer's OWN `start`, never a copy of its marker syntax: `start` is anchored and cheap,
// and reusing it means the guard cannot drift narrower than the grammar upstream accepts. A non-zero result
// says the block does not begin here, which is exactly when `tokenize` would have returned undefined anyway.
const guardBlockTokenizer = <T extends AnyExtension>(extension: T): T => {
    const original = (extension.config as { markdownTokenizer?: MarkdownTokenizer }).markdownTokenizer;
    const start = original?.start;

    if (!original || !start) {
        return extension;
    }

    return extension.extend({
        markdownTokenizer: {
            ...original,
            tokenize: (src: string, tokens: unknown[], lexer: unknown) =>
                start(src) === 0 ? original.tokenize(src, tokens, lexer) : undefined,
        },
    }) as T;
};

const longestRun = (text: string, runs: RegExp): number =>
    (text.match(runs) ?? []).reduce((max, run) => Math.max(max, run.length), 0);

// @tiptap/extension-code-block's renderMarkdown always emits a 3-backtick fence, so a code block whose content
// contains a ``` line (a doc demonstrating fenced markdown — common in knowledge/prompt examples) re-parses as
// TWO blocks on the next load: the inner fence closes the outer one. CommonMark requires the fence to be longer
// than any backtick run inside — widen it. CommonMark also forbids a backtick in a BACKTICK fence's info
// string while allowing one in a tilde fence's, so a language holding a backtick must ride a `~~~` fence.
// Otherwise identical to upstream.
const renderTunedCodeBlock = (node: JSONContent, helpers: MarkdownRendererHelpers): string => {
    const language: string = node.attrs?.language || '';
    const [marker, runs] = language.includes('`') ? (['~', /~+/g] as const) : (['`', /`+/g] as const);
    const content = node.content ? helpers.renderChildren(node.content) : '';
    const fence = marker.repeat(Math.max(3, longestRun(content, runs) + 1));

    return [`${fence}${language}`, content, fence].join('\n');
};

// @tiptap/extension-code-block's own parseMarkdown gates on `token.raw.startsWith('```')`, but CommonMark
// lets a fenced code block's opening fence be indented up to 3 spaces — marked then emits a valid `code`
// token whose `raw` starts with that whitespace, the gate rejects it, and the block is dropped on load. When
// a document mixes fences at different indents the mis-detection cascades and everything after the first
// dropped fence vanishes too. Trim the leading indent before the gate; otherwise identical to upstream.
const parseTunedCodeBlock = (token: MarkdownToken, helpers: MarkdownParseHelpers): MarkdownParseResult => {
    const fence = token.raw?.trimStart() ?? '';

    if (!fence.startsWith('```') && !fence.startsWith('~~~') && token.codeBlockStyle !== 'indented') {
        return [];
    }

    return helpers.createNode(
        'codeBlock',
        { language: token.lang || null },
        token.text ? [helpers.createTextNode(token.text)] : [],
    );
};

const lowlight = createLowlight(common);

// CodeBlockLowlight extends the default codeBlock, so the same byte-fidelity
// parse/render tuning applies verbatim; the highlighting it adds is a view-only
// ProseMirror decoration and never touches the serialized markdown. renderHTML
// stamps the fence language onto the `<pre>` as data-language so a CSS caption
// (index.css) can name the block — the label lives in the DOM, not the document.
const TunedCodeBlock = CodeBlockLowlight.extend({
    // `defaultLanguage` below feeds the highlight plugin, which reads `attrs.language || defaultLanguage`.
    // As an ATTRIBUTE default it would also stamp `plaintext` onto every block created in the editor, and
    // renderMarkdown would write that into the fence — a language the author never typed.
    addAttributes() {
        const attributes = this.parent?.() as undefined | { language?: Record<string, unknown> };

        return { ...attributes, language: { ...attributes?.language, default: null } };
    },
    parseMarkdown: parseTunedCodeBlock,
    renderHTML({ HTMLAttributes, node }) {
        const language = (node.attrs.language as null | string) || null;

        return [
            'pre',
            mergeAttributes(this.options.HTMLAttributes, HTMLAttributes, language ? { 'data-language': language } : {}),
            ['code', { class: language ? `${this.options.languageClassPrefix}${language}` : null }, 0],
        ];
    },
    renderMarkdown: renderTunedCodeBlock,
    // Without defaultLanguage an info-string-less fence falls to highlightAuto, which re-scans all 37 `common`
    // grammars over every code block in the document on each keystroke inside one.
}).configure({ defaultLanguage: 'plaintext', HTMLAttributes: { class: 'hljs' }, lowlight });

// A paragraph line that literally starts with `# ` or `> ` re-parses as a heading / blockquote on the next load
// (an ATX heading interrupts a paragraph; `>` opens a quote), silently changing the block TYPE of body text —
// reachable by Shift+Enter then a `# ` line. Escape those markers at line start; createTunedMarked's escape
// tokenizer unescapes `\#`/`\>` on load, so the round-trip stays faithful. Only `#`/`>` are handled: `-`/`*`/`+`/
// `1.`/fences overlap with literal regex/glob/backref escapes (`\*`, `\1`, `\|`) the editor must preserve.
const escapeLineLeadingBlockMarkers = (markdown: string): string =>
    markdown.replace(
        /(^|\n)( {0,3})(#{1,6} |> )/g,
        (_match, lineStart: string, indent: string, marker: string) => `${lineStart}${indent}\\${marker}`,
    );

// StarterKit's Bold/Italic register BOTH `**`/`*` and `__`/`_` input+paste rules. The marked layer keeps
// `_`-emphasis literal on load/paste, so leaving the underscore TYPING rules on would diverge — typed
// `__init__`/`_word_` would emphasize (→ `**init**`/`*word*`) while the same text loaded stays literal,
// breaking identifiers. Drop only the underscore rules (their `find` regex mentions `_`; the `*` rules stay)
// so typing matches load. (codeBlock is replaced by TunedCodeBlock below; underline off below.)
const TunedStarterKit = StarterKit.extend({
    addExtensions() {
        return (this.parent?.() ?? []).map((extension) => {
            if (extension.name === 'bold' || extension.name === 'italic') {
                return extension.extend({
                    addInputRules() {
                        return dropUnderscoreRules(this.parent?.() ?? []);
                    },
                    addPasteRules() {
                        return dropUnderscoreRules(this.parent?.() ?? []);
                    },
                });
            }

            if (extension.name === 'orderedList') {
                return withSelectAllRetargeted(guardBlockTokenizer(extension), 'toggleOrderedList');
            }

            if (extension.name === 'bulletList') {
                return withSelectAllRetargeted(extension, 'toggleBulletList');
            }

            if (extension.name === 'blockquote') {
                return withSelectAllRetargeted(extension, 'toggleBlockquote');
            }

            if (extension.name === 'paragraph') {
                return extension.extend({
                    renderMarkdown(node: JSONContent, helpers: MarkdownRendererHelpers) {
                        return escapeLineLeadingBlockMarkers(helpers.renderChildren(node.content ?? []));
                    },
                });
            }

            return extension;
        });
    },
});

// Single source of truth for the editor's extension stack — shared by markdown-editor.tsx AND the
// round-trip tests so they can never drift. createMarkdownLayer is the official @tiptap/markdown layer
// tuned for our content (see markdown-editor-marked.ts); VariableHighlight/TagHighlight are view-only decorations
// ({{vars}} / <tags>) that don't affect serialization.
//   • underline: false — its `++text++` markdown corrupts `C++ … C++` prose on load and Ctrl+U emits `++`.
//   • link autolink/linkOnPaste: true — a bare URL/email becomes a link on load, paste, AND typing, kept
//     symmetric with the marked layer (which no longer neutralises autolink/url). Do NOT set false: it
//     diverges typing from load and re-freezes bare URLs as text.
//   • link openOnClick: false — a click seats the caret in the link instead of navigating away, so LinkHandle
//     (markdown-editor-link-handle.tsx) can show the edit popover; opening still works via that popover's button.
export const createMarkdownExtensions = (placeholder?: string) => [
    TunedStarterKit.configure({
        codeBlock: false,
        link: { autolink: true, linkOnPaste: true, openOnClick: false },
        underline: false,
    }),
    withSelectAllRetargeted(TunedCodeBlock, 'toggleCodeBlock'),
    HeadingAutoformat,
    TunedTable.configure({ resizable: true }),
    TableRow,
    TableHeader,
    TableCell,
    withSelectAllRetargeted(guardBlockTokenizer(TaskList), 'toggleTaskList'),
    TaskItem.configure({ nested: true }),
    Image,
    VariableHighlight,
    TagHighlight,
    Placeholder.configure({ emptyEditorClass: 'is-editor-empty', placeholder }),
    ...createMarkdownLayer(),
    MarkdownPaste,
];
