import type {
    AnyExtension,
    CommandProps,
    Editor,
    JSONContent,
    MarkdownParseHelpers,
    MarkdownParseResult,
    MarkdownRendererHelpers,
    MarkdownToken,
} from '@tiptap/core';
import type { NodeType, Node as ProseMirrorNode } from '@tiptap/pm/model';
import type { EditorState } from '@tiptap/pm/state';

import { isList, mergeAttributes } from '@tiptap/core';
import { CodeBlockLowlight } from '@tiptap/extension-code-block-lowlight';
import { Image } from '@tiptap/extension-image';
import { TaskItem, TaskList } from '@tiptap/extension-list';
import { TableCell, TableHeader, TableRow } from '@tiptap/extension-table';
import { Placeholder } from '@tiptap/extensions';
import { NodeRange } from '@tiptap/pm/model';
import { Selection, TextSelection } from '@tiptap/pm/state';
import { findWrapping, liftTarget } from '@tiptap/pm/transform';
import StarterKit from '@tiptap/starter-kit';
import { common, createLowlight } from 'lowlight';

import { HeadingAutoformat } from './markdown-editor-heading-autoformat';
import { createMarkdownLayer, TunedTable } from './markdown-editor-marked';
import { MarkdownPaste } from './markdown-editor-paste';
import { TagHighlight } from './markdown-editor-tag-highlight';
import { VariableHighlight } from './markdown-editor-variable-highlight';

const dropUnderscoreRules = (rules: { find: unknown }[]) =>
    rules.filter((rule) => !(rule.find instanceof RegExp && rule.find.source.includes('_')));

type BlockAttrs = Record<string, unknown> | undefined;
type BlockFamily = {
    isApplied: (type: NodeType, state: EditorState, attrs?: BlockAttrs) => boolean;
    toggle: (type: NodeType, props: CommandProps, delegate: DelegatedToggle, attrs?: BlockAttrs) => boolean;
};
type CommandMap = Record<string, ToggleCommand | undefined>;
type DelegatedToggle = (props?: CommandProps) => boolean;
type ToggleCommand = (...args: never[]) => (props: never) => boolean;
type TopLevelChild = { node: ProseMirrorNode; pos: number };

// Ctrl+A and a mouse drag across everything are different Selection classes but the same intent, so keying off
// `instanceof AllSelection` silently leaves the drag path on the broken code path.
const isWholeDocumentSelection = ({ doc, selection }: EditorState): boolean => {
    const { empty, from, to } = selection;

    return (
        !empty &&
        ((from === 0 && to === doc.content.size) ||
            (from <= Selection.atStart(doc).from && to >= Selection.atEnd(doc).to))
    );
};

// TrailingNode appends its empty paragraph on the selectAll transaction itself, so it is already inside the
// user's selection before any toggle runs. Counting it makes an already-wrapped document read as unwrapped.
const contentChildren = (doc: ProseMirrorNode): TopLevelChild[] => {
    const children: TopLevelChild[] = [];

    doc.forEach((node, pos) => children.push({ node, pos }));

    const last = children.at(-1);

    if (children.length > 1 && last?.node.isTextblock && last.node.content.size === 0) {
        children.pop();
    }

    return children;
};

const contentSpan = (children: TopLevelChild[]) => {
    const last = children.at(-1)!;

    return { end: last.pos + last.node.nodeSize, start: children[0]!.pos };
};

// `liftTarget` is null for EVERY range at depth 0, so `commands.lift` can never undo a wrapper spanning the
// whole document. A range over the wrapper's own content resolves at depth 1, which does have a target.
const liftRangeOf = (doc: ProseMirrorNode, pos: number, node: ProseMirrorNode) =>
    doc.resolve(pos + 1).blockRange(doc.resolve(pos + node.nodeSize - 1));

const wrapFamily: BlockFamily = {
    isApplied: (type, { doc }) => {
        const children = contentChildren(doc);

        return children.length > 0 && children.every(({ node }) => node.type === type);
    },
    toggle: (type, { dispatch, state, tr }) => {
        const children = contentChildren(state.doc);

        if (!children.length) {
            return false;
        }

        if (wrapFamily.isApplied(type, state)) {
            const targets = children.map(({ node, pos }) => {
                const range = liftRangeOf(state.doc, pos, node);

                return range ? liftTarget(range) : null;
            });

            if (targets.some((target) => target === null)) {
                return false;
            }

            if (dispatch) {
                for (const { node, pos } of [...children].reverse()) {
                    const range = liftRangeOf(tr.doc, tr.mapping.map(pos), node);
                    const target = range ? liftTarget(range) : null;

                    if (range && target !== null) {
                        tr.lift(range, target);
                    }
                }
            }

            return true;
        }

        const { end, start } = contentSpan(children);
        const range = new NodeRange(state.doc.resolve(start), state.doc.resolve(end), 0);
        const wrapping = findWrapping(range, type);

        if (!wrapping) {
            return false;
        }

        if (dispatch) {
            tr.wrap(range, wrapping);
        }

        return true;
    },
};

const canRetype = (doc: ProseMirrorNode, pos: number, type: NodeType) => {
    const $pos = doc.resolve(pos);

    return $pos.parent.canReplaceWith($pos.index(), $pos.index() + 1, type);
};

const blockTypeCandidates = (state: EditorState, type: NodeType): TopLevelChild[] => {
    const children = contentChildren(state.doc);

    if (!children.length) {
        return [];
    }

    const { end, start } = contentSpan(children);
    const candidates: TopLevelChild[] = [];

    state.doc.nodesBetween(start, end, (node, pos) => {
        // A cell serialises inline, so a fence placed inside one comes back from the next load as literal
        // backticks. Descending into a table trades a reversible refusal for silent corruption.
        if (node.type.spec.tableRole === 'table') {
            return false;
        }

        if (node.isTextblock && (node.type === type || canRetype(state.doc, pos, type))) {
            candidates.push({ node, pos });
        }

        return true;
    });

    return candidates;
};

// A heading carries its level in attrs, so "already applied" and the value written both have to account for
// them: comparing node type alone reads an all-H1 document as already being Heading 2 and demotes it, and
// dropping the attrs on the way in stamps the schema default level for every choice the user makes.
const hasAttrs = (node: ProseMirrorNode, attrs: BlockAttrs) =>
    !attrs || Object.entries(attrs).every(([key, value]) => node.attrs[key] === value);

const blockTypeFamily: BlockFamily = {
    isApplied: (type, state, attrs) => {
        const candidates = blockTypeCandidates(state, type);

        return candidates.length > 0 && candidates.every(({ node }) => node.type === type && hasAttrs(node, attrs));
    },
    toggle: (type, { dispatch, state, tr }, _delegate, attrs) => {
        const candidates = blockTypeCandidates(state, type);

        if (!candidates.length) {
            return false;
        }

        const isApplied = blockTypeFamily.isApplied(type, state, attrs);
        const target = isApplied ? state.schema.nodes.paragraph! : type;
        const targetAttrs = isApplied ? null : ((attrs ?? null) as null | Record<string, unknown>);

        if (
            !candidates.some(
                ({ node, pos }) =>
                    (node.type !== target || !hasAttrs(node, attrs)) && canRetype(state.doc, pos, target),
            )
        ) {
            return false;
        }

        if (dispatch) {
            for (const { node, pos } of [...candidates].reverse()) {
                tr.setBlockType(pos, pos + node.nodeSize, target, targetAttrs);
            }
        }

        return true;
    },
};

const listFamily: BlockFamily = {
    isApplied: (type, state) => {
        const children = contentChildren(state.doc);

        return children.length > 0 && children.every(({ node }) => node.type === type);
    },
    toggle: (type, { chain, editor, state }, delegate) => {
        const { extensions } = editor.extensionManager;
        const children = contentChildren(state.doc);
        // Derived from `children` rather than recomputed, so the index lookups below compare the same objects.
        const candidates = children.filter(({ node }) => node.isTextblock || isList(node.type.name, extensions));
        const first = candidates[0];
        const last = candidates.at(-1);

        if (!first || !last) {
            return false;
        }

        const isApplied = children.every(({ node }) => node.type === type);
        const hasGap = children.indexOf(last) - children.indexOf(first) + 1 !== candidates.length;

        if (!isApplied && hasGap) {
            return false;
        }

        const { end, start } = contentSpan(candidates);
        const seated = isApplied
            ? TextSelection.between(state.doc.resolve(start + 1), state.doc.resolve(end - 1))
            : TextSelection.between(state.doc.resolve(start), state.doc.resolve(end));
        const restored = state.selection;
        let hasToggled = false;

        return (
            // Nothing here may branch on `dispatch`. Under `can()` the transaction is discarded anyway, so
            // seating costs nothing — but skipping it there hands the delegate the whole-document selection it
            // cannot use, and the control reports unavailable for a click that would have worked.
            chain()
                .command(({ tr }) => {
                    tr.setSelection(seated);

                    return true;
                })
                .command((chained) => {
                    hasToggled = delegate(chained);

                    return hasToggled;
                })
                // A chain dispatches what its callbacks already wrote even when one returns false, so a delegate
                // that half-applied its `clearNodes` fallback would commit a flattened document without this.
                .command(({ tr }) => {
                    if (hasToggled) {
                        tr.setSelection(restored.map(tr.doc, tr.mapping));
                    } else {
                        tr.setMeta('preventDispatch', true);
                    }

                    return true;
                })
                .run()
        );
    },
};

const BLOCK_FAMILIES: Record<string, BlockFamily | undefined> = {
    blockquote: wrapFamily,
    bulletList: listFamily,
    codeBlock: blockTypeFamily,
    orderedList: listFamily,
    taskList: listFamily,
};

/**
 * The toolbar MUST read its pressed state from here rather than from `editor.isActive`: tiptap's predicate
 * demands the type cover the whole selection, so an already-quoted document under Ctrl+A reads as not quoted
 * and the button would contradict what its own click does.
 */
export const isBlockApplied = (editor: Editor, nodeName: string): boolean => {
    const { state } = editor;
    const family = BLOCK_FAMILIES[nodeName];
    const type = state.schema.nodes[nodeName];

    if (!family || !type || !isWholeDocumentSelection(state)) {
        return editor.isActive(nodeName);
    }

    return family.isApplied(type, state);
};

const withWholeDocumentToggle = <T extends AnyExtension>(extension: T, commandName: string, family: BlockFamily): T =>
    extension.extend({
        addCommands() {
            const parent = (this.parent?.() ?? {}) as CommandMap;
            const original = parent[commandName];
            const { name } = this;

            if (!original) {
                return parent;
            }

            return {
                ...parent,
                [commandName]:
                    (...args: never[]) =>
                    (props: never) => {
                        const commandProps = props as CommandProps;
                        const delegate = (chained: CommandProps = commandProps) => original(...args)(chained as never);

                        if (!isWholeDocumentSelection(commandProps.state)) {
                            return delegate();
                        }

                        return family.toggle(
                            commandProps.state.schema.nodes[name]!,
                            commandProps,
                            delegate,
                            args[0] as BlockAttrs,
                        );
                    },
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
    // CodeBlockLowlight builds its plugins as `[...this.parent?.(), LowlightPlugin(...)]` and `.extend()` copies
    // that into every layer, so each layer highlights the whole document again. Measured: this pass-through
    // collapses them to one only from INSIDE this literal — as an outer `.extend()` layer it leaves three.
    addProseMirrorPlugins() {
        return this.parent?.() ?? [];
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
        /(^|\n)( {0,3})(#{1,6}(?=[ \t\n]|$)|>)/g,
        (_match, lineStart: string, indent: string, marker: string) => `${lineStart}${indent}\\${marker}`,
    );

// marked reads a line of only `=` or `-` sitting directly under text as a setext underline (`(=+|-+) *` — no
// three-character minimum), so a hard break followed by such a line turns the whole paragraph into a heading
// and eats the run. Only continuation lines need it: a paragraph's first line is preceded by a blank line,
// which stops lheading. The load side unescapes exactly this shape, so an author's `\-` inside a character
// class is left alone.
const escapeSetextUnderlines = (markdown: string): string =>
    markdown.replace(
        /(\n {0,3})((?:=+|-+) *)(?=\n|$)/g,
        (_match, lineStart: string, run: string) => `${lineStart}\\${run}`,
    );

// A `]` inside an image alt closes the markdown label early and the image loses its node entirely on reload.
// marked unescapes `\[`/`\]` inside a label through its own outputLink pass, independent of the neutralised
// inline escape tokenizer, so escaping just those two closes the round trip. A lone `\` is NOT escaped —
// marked would not unescape it back.
const escapeBracketsInLabel = (text: string): string => text.replace(/[[\]]/g, (bracket) => `\\${bracket}`);

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
                return withWholeDocumentToggle(guardBlockTokenizer(extension), 'toggleOrderedList', listFamily);
            }

            if (extension.name === 'bulletList') {
                return withWholeDocumentToggle(extension, 'toggleBulletList', listFamily);
            }

            if (extension.name === 'blockquote') {
                return withWholeDocumentToggle(extension, 'toggleBlockquote', wrapFamily);
            }

            if (extension.name === 'heading') {
                return withWholeDocumentToggle(extension, 'toggleHeading', blockTypeFamily);
            }

            if (extension.name === 'paragraph') {
                return extension.extend({
                    renderMarkdown(node: JSONContent, helpers: MarkdownRendererHelpers) {
                        return escapeSetextUnderlines(
                            escapeLineLeadingBlockMarkers(helpers.renderChildren(node.content ?? [])),
                        );
                    },
                });
            }

            // Emit the code mark's content with NO backtick fence: the real content-sized fence comes from
            // serializeCodeSpan in the text encoder (markdown-editor-marked.ts). The two are paired — restoring a
            // fence here double-wraps every code span.
            if (extension.name === 'code') {
                return extension.extend({
                    renderMarkdown(node: JSONContent, helpers: MarkdownRendererHelpers) {
                        return node.content ? helpers.renderChildren(node.content) : '';
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
    withWholeDocumentToggle(TunedCodeBlock, 'toggleCodeBlock', blockTypeFamily),
    HeadingAutoformat,
    TunedTable.configure({ resizable: true }),
    TableRow,
    TableHeader,
    TableCell,
    withWholeDocumentToggle(guardBlockTokenizer(TaskList), 'toggleTaskList', listFamily),
    TaskItem.configure({
        a11y: {
            checkboxLabel: (node) => `Task item checkbox for ${node.firstChild?.textContent || 'empty task item'}`,
        },
        nested: true,
    }),
    Image.extend({
        renderMarkdown(node: JSONContent) {
            const { alt = '', src = '', title = '' } = node.attrs ?? {};
            const label = escapeBracketsInLabel(String(alt));

            return title ? `![${label}](${src} "${title}")` : `![${label}](${src})`;
        },
    }),
    VariableHighlight,
    TagHighlight,
    Placeholder.configure({ emptyEditorClass: 'is-editor-empty', placeholder }),
    ...createMarkdownLayer(),
    MarkdownPaste,
];
