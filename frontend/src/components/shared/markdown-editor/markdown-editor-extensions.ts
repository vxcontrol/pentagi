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
import {
    blankLineOf,
    indentItemBlock,
    isRule,
    isWrittenEmpty,
    itemMarker,
    joinBlocks,
    lineBreaksAtEnd,
    MarkdownLayout,
    renderDocument,
    renderList,
    withinItem,
    withLayoutFacts,
} from './markdown-editor-layout';
import {
    createMarkdownLayer,
    readsAsHeading,
    readsAsLinkedImage,
    renderInline,
    TunedTable,
    withMarkedBreaks,
    withoutMarkdownTokenizer,
} from './markdown-editor-marked';
import { MarkdownPaste } from './markdown-editor-paste';
import { TEMPLATE_ACTION } from './markdown-editor-table-pipes';
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
// document, and @tiptap/extension-list's task tokenizer opens with `src.split('\n')` before bailing on a
// first line that isn't a task item — so each boundary materialises all remaining lines, O(n²).
//
// The gate must accept every line upstream's item pattern accepts: a task line it rejects falls through to
// marked's own list rule and loads as a plain bullet list.
const OPENS_TASK_LIST = /^[ \t]*[-+*][ \t]+\[[ xX]\][ \t]/;

const guardBlockTokenizer = <T extends AnyExtension>(extension: T, opens: RegExp): T => {
    const original = (extension.config as { markdownTokenizer?: MarkdownTokenizer }).markdownTokenizer;

    if (!original) {
        return extension;
    }

    return extension.extend({
        markdownTokenizer: {
            ...original,
            tokenize: (src: string, tokens: unknown[], lexer: unknown) =>
                opens.test(src) ? original.tokenize(src, tokens, lexer) : undefined,
        },
    }) as T;
};

// @tiptap/extension-list tokenizes numbered lists with a line scanner of its own. It does not know a fence, so
// a code block under a step is re-read as list lines; it strips one column too few from an item's continuation
// lines, so that code gains a space of indentation on every load; and it takes a letter or roman marker, so
// `Mr. Smith` is a list item and is saved as `1. Smith`. marked's own list rule is CommonMark's: without the
// custom tokenizer a numbered list reaches parseMarkdown as the same native `list` token a bullet list does,
// and `a.` or `IV.` at the start of a line is the text it was written as.
// parseMarkdown is replaced with it: upstream reads `token.start || 1`, renumbering a list that opens at 0, and
// builds the items itself, past listItem's parseMarkdown. `type` goes too — nothing loaded sets it, and a
// lettered list pasted as HTML would show `a.` while renderListItem saves `1.`.
const withNativeListTokens = <T extends AnyExtension>(extension: T): T =>
    withoutMarkdownTokenizer(extension).extend({
        addAttributes() {
            const inherited = (this as unknown as { parent?: () => Record<string, object> }).parent?.() ?? {};
            const { type: _markerType, ...attributes } = inherited;

            return attributes;
        },
        // Upstream's plugin turns pasted lines that open with numbers, letters or roman numerals into a list
        // by itself. MarkdownPaste already parses a paste the way a load is parsed, and declines one inside
        // code — where this plugin would split the code block to insert its list.
        addProseMirrorPlugins: () => [],
        parseMarkdown(token: MarkdownToken, helpers: MarkdownParseHelpers): MarkdownParseResult {
            if (token.type !== 'list' || !token.ordered) {
                return [];
            }

            const start = typeof token.start === 'number' ? token.start : 1;

            return {
                ...(start === 1 ? {} : { attrs: { start } }),
                content: helpers.parseChildren(token.items ?? []),
                type: 'orderedList',
            };
        },
    }) as T;

type ParseMarkdown = (token: MarkdownToken, helpers: MarkdownParseHelpers) => MarkdownParseResult;

// The schema wants a paragraph first in every item, and tiptap builds the document without checking it.
// `1. ```sh` puts a code block first, so the item gets the empty paragraph the schema implies; renderListItem
// writes the pair back on the marker's line.
const withLeadingParagraph = (parsed: MarkdownParseResult): MarkdownParseResult => {
    if (
        Array.isArray(parsed) ||
        'mark' in parsed ||
        parsed.type !== 'listItem' ||
        parsed.content?.[0]?.type === 'paragraph'
    ) {
        return parsed;
    }

    return { ...parsed, content: [{ content: [], type: 'paragraph' }, ...(parsed.content ?? [])] };
};

// marked lifts a task marker out of an item's text into a `checkbox` token: a block token ahead of the text in
// a tight list, the first inline token of the paragraph in a loose one. A task list is tokenized upstream and
// never gets here, so in these items the marker is part of what was written: dropped, `1. [ ] step` is saved
// as `1. step`.
const withLiteralCheckbox = (token: MarkdownToken): MarkdownToken => {
    const [first, second, ...rest] = token.tokens ?? [];
    const asText = (checkbox: MarkdownToken): MarkdownToken => ({
        raw: checkbox.raw,
        text: checkbox.raw,
        type: 'text',
    });

    if (first?.type === 'checkbox' && second) {
        const literal = asText(first);
        const merged = {
            ...second,
            raw: literal.raw + (second.raw ?? ''),
            text: literal.text + (second.text ?? ''),
            tokens: [literal, ...(second.tokens ?? [])],
        };

        return { ...token, tokens: [merged, ...rest] };
    }

    const [lead, ...inline] = first?.tokens ?? [];

    if (first && lead?.type === 'checkbox') {
        return {
            ...token,
            tokens: [{ ...first, tokens: [asText(lead), ...inline] }, ...(second ? [second] : []), ...rest],
        };
    }

    return token;
};

// marked types a tight item's text as bare `text` tokens, and upstream wraps only the first one in a paragraph:
// text that follows another block reaches the document as a text node outside any paragraph, with its inline
// markup unread. Typed as paragraphs they all take the path a loose item's text takes.
const withParagraphs = (token: MarkdownToken): MarkdownToken => ({
    ...token,
    tokens: token.tokens?.map((block) => (block.type === 'text' ? { ...block, type: 'paragraph' } : block)),
});

const renderWith = (helpers: MarkdownRendererHelpers) => (child: JSONContent, index: number) =>
    helpers.renderChild?.(child, index) ?? helpers.renderChildren([child]);

const renderListNode = (node: JSONContent, helpers: MarkdownRendererHelpers): string =>
    renderList(node, renderWith(helpers));

// Upstream writes an item's first block after the marker and indents only the blocks that follow it, so a
// first block of several lines (a fence opened on the marker's line) loses its indentation and swallows the
// rest of the list on the next load. Here every line after the marker's is aligned to the content column — the
// marker's own width, not a fixed two spaces — which is what makes a block under `10.` a child for every other
// markdown reader, and keeps a wrapped item's hanging indent as it was written.
const renderListItem = (node: JSONContent, helpers: MarkdownRendererHelpers): string => {
    const marker = itemMarker(node);
    const pad = ' '.repeat(marker.length);
    const children = node.content ?? [];
    // An empty paragraph ahead of another block is the one withLeadingParagraph added: the block shares the
    // marker's line again, as it was written. A rule there is drawn with the character the bullet is not —
    // `- ---` and `* ***` are themselves rules, not items.
    const firstIndex = isWrittenEmpty(children[0]) && children.length > 1 && children[1]!.type !== 'paragraph' ? 1 : 0;
    const isRuleFirst = firstIndex && children[1]!.type === 'horizontalRule';
    const { lazyDepth, lazyFrom } = node.attrs ?? {};
    const isLazy = typeof lazyFrom === 'number' && !firstIndex && children[0]?.type === 'paragraph';
    const render = renderWith(helpers);

    return withinItem(() =>
        joinBlocks(
            children.slice(firstIndex),
            (child, at) => {
                const written = render(child, firstIndex + at);

                return isRuleFirst && !at && written[0] === marker[0] ? (marker[0] === '*' ? '---' : '***') : written;
            },
            true,
            (written, at, child) => {
                const blank = blankLineOf(child, pad);

                if (at) {
                    return indentItemBlock(written, pad, 0, Infinity, 1, blank);
                }

                const depth = Math.min(Math.max(Math.trunc(Number(lazyDepth)) || 1, 1), 64);
                const block = indentItemBlock(written, pad, 1, isLazy ? lazyFrom : Infinity, depth, blank);

                // A marker with nothing after it on its line is written bare; text keeps even the spaces its
                // line ends with.
                return (!block || block.startsWith('\n') ? marker.trimEnd() : marker) + block;
            },
        ),
    );
};

// An ATX heading is one line. A setext heading may run over several, and it is what a paragraph with `---`
// straight under it parses as — a front-matter block, a separator typed without a blank line. Written as ATX
// its later lines fall out of the heading on the next load; from level 3 on there is no other form, so there
// a line break is written as the space it reads as. So is one an edit has left where setext cannot hold it:
// at either end of the heading, or above a line that would be read as a block of its own.
const renderHeading = (node: JSONContent, helpers: MarkdownRendererHelpers): string => {
    const level = Number(node.attrs?.level) || 1;
    const written = (node.content ? renderInline(node.content, helpers) : '').split('\n');
    // Each line without what a hard break is written as at its end.
    const lines = written.map((line, at) =>
        at === written.length - 1 ? line : line.endsWith('\\') ? line.slice(0, -1) : line.trimEnd(),
    );
    const from = Math.max(
        0,
        lines.findIndex((line) => line.trim()),
    );
    const to = lines.findLastIndex((line) => line.trim()) + 1;
    // The underline the heading was written with, while the heading is still of its level.
    const row = level === 1 ? '=' : '-';
    const underline: unknown = node.attrs?.underline;
    const isUnderlined =
        typeof underline === 'string' && underline !== '' && underline === row.repeat(underline.length);

    if (level <= 2 && to > from && (to - from > 1 || isUnderlined)) {
        const setext = [...written.slice(from, to - 1), lines[to - 1], isUnderlined ? underline : row.repeat(3)].join(
            '\n',
        );

        if (readsAsHeading(setext, level)) {
            return setext;
        }
    }

    const line = lines
        .slice(from, to)
        .map((part, at) => (at ? part.trimStart() : part))
        .filter(Boolean)
        .join(' ');

    // Not after text that ends with white space of its own: read back, the two would be one trail.
    const trail: unknown = line === line.trimEnd() ? node.attrs?.trail : '';

    return line ? `${'#'.repeat(level)} ${line}${typeof trail === 'string' && /^ +$/.test(trail) ? trail : ''}` : '';
};

const longestRun = (text: string, runs: RegExp): number =>
    (text.match(runs) ?? []).reduce((max, run) => Math.max(max, run.length), 0);

// @tiptap/extension-code-block's renderMarkdown always emits a 3-backtick fence, so a code block whose content
// contains a ``` line (a doc demonstrating fenced markdown — common in knowledge/prompt examples) re-parses as
// TWO blocks on the next load: the inner fence closes the outer one. Only a line that is nothing but a fence
// run closes a block — ` ```bash ` inside one does not — so the fence is one longer than the longest such
// line, and never shorter than the one the block was written with. CommonMark also forbids a backtick in a
// BACKTICK fence's info string while allowing one in a tilde fence's, so a language holding a backtick rides
// a `~~~` fence, as does a block that was written with one.
const renderTunedCodeBlock = (node: JSONContent, helpers: MarkdownRendererHelpers): string => {
    const language: string = node.attrs?.language || '';
    const written = typeof node.attrs?.fence === 'string' ? node.attrs.fence : '';
    const marker = language.includes('`') || written.startsWith('~') ? '~' : '`';
    const content = node.content ? helpers.renderChildren(node.content) : '';
    // Each run is of one character, so a line of a thousand backticks is matched once, not from every length.
    const closers =
        content.match(marker === '~' ? /^ {0,3}~{3,}(?:`[~`]*)? *$/gm : /^ {0,3}`{3,}(?:~[~`]*)? *$/gm) ?? [];
    const longest = closers.reduce((max, line) => Math.max(max, longestRun(line, marker === '~' ? /~+/g : /`+/g)), 0);
    const fence = marker.repeat(Math.max(3, longest + 1, written.startsWith(marker) ? written.length : 0));

    return [`${fence}${language.startsWith('~') ? ' ' : ''}${language}`, content, fence].join('\n');
};

// @tiptap/extension-code-block's own parseMarkdown gates on `token.raw.startsWith('```')`, but CommonMark
// lets a fenced code block's opening fence be indented up to 3 spaces — marked then emits a valid `code`
// token whose `raw` starts with that whitespace, the gate rejects it, and the block is dropped on load. When
// a document mixes fences at different indents the mis-detection cascades and everything after the first
// dropped fence vanishes too. Trim the leading indent before the gate.
const parseTunedCodeBlock = (token: MarkdownToken, helpers: MarkdownParseHelpers): MarkdownParseResult => {
    const fence = token.raw?.trimStart() ?? '';

    if (!fence.startsWith('```') && !fence.startsWith('~~~') && token.codeBlockStyle !== 'indented') {
        return [];
    }

    // marked leaves the last line break in the text of an indented block that no blank line follows, which
    // a fence would then hold as an empty line of code.
    const code = token.text ?? '';
    const text = token.codeBlockStyle === 'indented' ? code.slice(0, code.length - lineBreaksAtEnd(code).length) : code;

    return helpers.createNode(
        'codeBlock',
        { language: token.lang || null },
        text ? [helpers.createTextNode(text)] : [],
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
    // exitOnArrowUp adds an empty paragraph above a code block that opens the document: a navigation key that
    // edits the document and dirties the form, on exactly the documents agents store as code.
}).configure({ defaultLanguage: 'plaintext', exitOnArrowUp: false, HTMLAttributes: { class: 'hljs' }, lowlight });

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
// A bracket inside a Go action is the template's — escaped, `{{ index .M "[k]" }}` no longer parses.
const LABEL_BRACKET = new RegExp(`${TEMPLATE_ACTION.source}|[[\\]]`, 'g');

const escapeBracketsInLabel = (text: string): string =>
    text.replace(LABEL_BRACKET, (match) => (match.length > 1 ? match : `\\${match}`));

// StarterKit's Bold/Italic register BOTH `**`/`*` and `__`/`_` input+paste rules. The marked layer keeps
// `_`-emphasis literal on load/paste, so leaving the underscore TYPING rules on would diverge — typed
// `__init__`/`_word_` would emphasize (→ `**init**`/`*word*`) while the same text loaded stays literal,
// breaking identifiers. Drop only the underscore rules (their `find` regex mentions `_`; the `*` rules stay)
// so typing matches load. (codeBlock is replaced by TunedCodeBlock below; underline off below.)
// A Go action is literal text, so a delimiter inside one is not a delimiter: a backtick typed after
// {{ printf `%s` .A }} must not pair with the one that closes the raw string and take it out of the action.
// A match that holds a whole action — `**{{ .Name }}**` — cuts nothing.
const crossesAction = (match: { 0: string; index?: number; input?: string }): boolean => {
    const from = match.index ?? 0;
    const to = from + match[0].length;

    return [...(match.input ?? '').matchAll(TEMPLATE_ACTION)].some((action) => {
        const end = action.index + action[0].length;

        return action.index < to && end > from && !(from <= action.index && to >= end);
    });
};

type MatchHandler = (props: { match: RegExpMatchArray }) => unknown;

const outsideActions = <Rule extends object>(rules: Rule[]): Rule[] =>
    rules.map((rule) => {
        const matched = rule as { handler: MatchHandler };
        const handle = matched.handler;

        matched.handler = (props) => (crossesAction(props.match) ? null : handle(props));

        return rule;
    });

type WithParentRules = { parent?: () => object[] };

const tuneStarterKitExtension = (extension: AnyExtension): AnyExtension => {
    if (extension.name === 'bold' || extension.name === 'italic') {
        return extension.extend({
            addInputRules() {
                return outsideActions(dropUnderscoreRules(this.parent?.() ?? []));
            },
            addPasteRules() {
                return outsideActions(dropUnderscoreRules(this.parent?.() ?? []));
            },
        });
    }

    if (extension.name === 'strike') {
        return extension.extend({
            addInputRules(this: WithParentRules) {
                return outsideActions(this.parent?.() ?? []) as never;
            },
            addPasteRules(this: WithParentRules) {
                return outsideActions(this.parent?.() ?? []) as never;
            },
        });
    }

    if (extension.name === 'orderedList') {
        return withWholeDocumentToggle(
            withNativeListTokens(extension).extend({ renderMarkdown: renderListNode }),
            'toggleOrderedList',
            listFamily,
        );
    }

    // ListKeymap's Tab pulls the block under a list into its last item. Tab is how the keyboard leaves
    // the editor, so from that one caret position it would edit the document instead.
    if (extension.name === 'listKeymap') {
        return extension.extend({
            addKeyboardShortcuts() {
                const { Tab: _sinkIntoList, ...shortcuts } = this.parent?.() ?? {};

                return shortcuts;
            },
        });
    }

    if (extension.name === 'listItem') {
        const parseItem = (extension.config as { parseMarkdown?: ParseMarkdown }).parseMarkdown;

        return extension.extend({
            parseMarkdown: (token: MarkdownToken, helpers: MarkdownParseHelpers) =>
                withLeadingParagraph(parseItem?.(withParagraphs(withLiteralCheckbox(token)), helpers) ?? []),
            renderMarkdown: renderListItem,
        });
    }

    if (extension.name === 'doc') {
        return extension.extend({
            renderMarkdown: (node: JSONContent, helpers: MarkdownRendererHelpers) =>
                renderDocument(node, renderWith(helpers)),
        });
    }

    if (extension.name === 'bulletList') {
        return withWholeDocumentToggle(
            extension.extend({ renderMarkdown: renderListNode }),
            'toggleBulletList',
            listFamily,
        );
    }

    // `>` with nothing after it is a quote that holds nothing, which the schema does not allow and upstream
    // writes as nothing: it holds an empty paragraph, as a quote made in the editor does.
    if (extension.name === 'blockquote') {
        const parse = (extension.config as { parseMarkdown?: (...args: unknown[]) => JSONContent }).parseMarkdown;

        return withWholeDocumentToggle(
            extension.extend({
                parseMarkdown(...args: unknown[]) {
                    const quote = parse?.apply(this, args) ?? { type: 'blockquote' };

                    return quote.content?.length ? quote : { ...quote, content: [{ type: 'paragraph' }] };
                },
                // Upstream puts an empty quoted line between every two blocks; the blocks of a quote are
                // joined the way the blocks of the document are, a list right under its lead-in included.
                renderMarkdown: (node: JSONContent, helpers: MarkdownRendererHelpers) =>
                    joinBlocks(node.content ?? [], renderWith(helpers), false)
                        .split('\n')
                        .map((line) => (line.trim() ? `> ${line}` : '>'))
                        .join('\n'),
            }),
            'toggleBlockquote',
            wrapFamily,
        );
    }

    if (extension.name === 'heading') {
        return withWholeDocumentToggle(
            extension.extend({ renderMarkdown: renderHeading }),
            'toggleHeading',
            blockTypeFamily,
        );
    }

    // tiptap's input-rule runner skips a match whose text differs from `textBetween` over the same span.
    // A hard break is "\n" to the matcher but "" to textBetween unless its spec has leafText, so a mark
    // typed right after or across Shift+Enter stays literal while the same text loads formatted.
    // extendNodeSchema runs for EVERY node of the schema, hence the name check: without it horizontalRule
    // and image would report the text too.
    if (extension.name === 'hardBreak') {
        return extension.extend({
            // How the break was written: nothing but the line's end, two or more spaces, or a backslash. It
            // has to be in the DOM too — ProseMirror reads an edited paragraph back from there.
            addAttributes() {
                return {
                    marker: {
                        default: null,
                        keepOnSplit: false,
                        parseHTML: (element: HTMLElement) => element.getAttribute('data-break'),
                        renderHTML: ({ marker }: { marker?: unknown }) =>
                            typeof marker === 'string' ? { 'data-break': marker } : {},
                    },
                };
            },
            extendNodeSchema(node: { name: string }) {
                return node.name === 'hardBreak' ? { leafText: () => '\n' } : {};
            },
            parseMarkdown: (token: MarkdownToken) => {
                const marker = (token.raw ?? '').replace(/\n$/, '');

                return { type: 'hardBreak', ...(marker === '  ' ? {} : { attrs: { marker } }) };
            },
            renderMarkdown: (node: JSONContent) => {
                const marker: unknown = node.attrs?.marker;

                return `${typeof marker === 'string' && /^(?: {2,}|\\)?$/.test(marker) ? marker : '  '}\n`;
            },
        });
    }

    // The form a link was written in, for renderInline to write it in again.
    if (extension.name === 'link') {
        return extension.extend({
            // linkify ends a URL inside the action it runs into, and the link then cuts the action in two:
            // `https://host/{{ .Path }}` typed or pasted stays text, as it does when it is loaded.
            addOptions(this: { parent?: () => unknown }) {
                const options = this.parent?.() as { shouldAutoLink: (url: string) => boolean };

                return {
                    ...options,
                    shouldAutoLink: (url: string) => !/\{\{|\}\}/.test(url) && options.shouldAutoLink(url),
                } as never;
            },
            parseMarkdown: (token: MarkdownToken, helpers: MarkdownParseHelpers) => {
                const opener = (token.raw ?? '[')[0];
                const form = opener === '[' ? null : opener === '<' ? 'angle' : 'bare';

                const target = { href: token.href, title: token.title || null };
                // An image takes no mark: the link around one is kept on the image, and written from there.
                const content = helpers
                    .parseInline(token.tokens ?? [])
                    .map((node) =>
                        node.type === 'image' ? { ...node, attrs: { ...node.attrs, link: target } } : node,
                    );

                return helpers.applyMark('link', content, { ...target, ...(form ? { form } : {}) });
            },
        });
    }

    // `***`, `___` and `- - -` are rules as much as `---` is, and are written back as they were.
    if (extension.name === 'horizontalRule') {
        return extension.extend({
            renderMarkdown: (node: JSONContent) => {
                const rule: unknown = node.attrs?.rule;

                return typeof rule === 'string' && isRule(rule) ? rule : '---';
            },
        });
    }

    if (extension.name === 'paragraph') {
        return extension.extend({
            // Four columns of space before the first word make the line code, here and after an item's
            // marker. marked drops fewer than that, so text typed after them is written without them.
            renderMarkdown(node: JSONContent, helpers: MarkdownRendererHelpers) {
                const text = renderInline(node.content ?? [], helpers);

                return escapeSetextUnderlines(
                    escapeLineLeadingBlockMarkers(/^(?: {4}| {0,3}\t)/.test(text) ? text.trimStart() : text),
                );
            },
        });
    }

    // Emit the code mark's content with NO backtick fence: the real content-sized fence comes from
    // serializeCodeSpan in the text encoder (markdown-editor-marked.ts). The two are paired — restoring a
    // fence here double-wraps every code span.
    // A span's line breaks are break nodes like a paragraph's (withLineBreaks).
    if (extension.name === 'code') {
        return extension.extend({
            addInputRules(this: WithParentRules) {
                return outsideActions(this.parent?.() ?? []) as never;
            },
            addPasteRules(this: WithParentRules) {
                return outsideActions(this.parent?.() ?? []) as never;
            },
            parseMarkdown: (token: MarkdownToken, helpers: MarkdownParseHelpers) => {
                const text = token.text ?? '';
                // The fence the span was written with, where it is not the one the writer would choose.
                const fence = /^`+/.exec(token.raw ?? '')?.[0].length ?? 0;
                const longest = (text.match(/`+/g) ?? []).reduce((max, run) => Math.max(max, run.length), 0);

                return helpers.applyMark(
                    'code',
                    text
                        .split('\n')
                        .flatMap((line, at) => [
                            ...(at ? [{ attrs: { marker: '' }, type: 'hardBreak' }] : []),
                            ...(line ? [helpers.createTextNode(line)] : []),
                        ]),
                    fence && fence !== longest + 1 ? { fence } : undefined,
                );
            },
            renderMarkdown(node: JSONContent, helpers: MarkdownRendererHelpers) {
                return node.content ? helpers.renderChildren(node.content) : '';
            },
        });
    }

    return extension;
};

const TunedStarterKit = StarterKit.extend({
    addExtensions() {
        return (this.parent?.() ?? []).map((extension) =>
            withMarkedBreaks(withLayoutFacts(tuneStarterKitExtension(extension))),
        );
    },
});

// Single source of truth for the editor's extension stack — shared by markdown-editor.tsx AND the
// round-trip tests so they can never drift. createMarkdownLayer is the official @tiptap/markdown layer
// tuned for our content (see markdown-editor-marked.ts); VariableHighlight/TagHighlight are view-only decorations
// ({{vars}} / <tags>) that don't affect serialization.
//   • underline: false — its `++text++` markdown corrupts `C++ … C++` prose on load and Ctrl+U emits `++`.
//   • link autolink/linkOnPaste: true — a bare URL/email becomes a link on load, paste, AND typing, kept
//     symmetric with the marked layer (which reads a web address or an e-mail as a link). Do NOT set false: it
//     diverges typing from load and re-freezes bare URLs as text.
//   • link openOnClick: false — a click seats the caret in the link instead of navigating away, so LinkHandle
//     (markdown-editor-link-handle.tsx) can show the edit popover; opening still works via that popover's button.
export const createMarkdownExtensions = (placeholder?: string) => [
    TunedStarterKit.configure({
        codeBlock: false,
        link: { autolink: true, linkOnPaste: true, openOnClick: false },
        underline: false,
    }),
    withWholeDocumentToggle(withLayoutFacts(TunedCodeBlock), 'toggleCodeBlock', blockTypeFamily),
    HeadingAutoformat,
    withLayoutFacts(TunedTable).configure({ resizable: true }),
    TableRow,
    TableHeader,
    TableCell,
    withWholeDocumentToggle(
        withLayoutFacts(guardBlockTokenizer(TaskList, OPENS_TASK_LIST)),
        'toggleTaskList',
        listFamily,
    ),
    TaskItem.configure({
        a11y: {
            checkboxLabel: (node) =>
                `Task item checkbox for ${node.firstChild?.textContent.trim() || 'empty task item'}`,
        },
        nested: true,
    }),
    Image.extend({
        renderMarkdown(node: JSONContent) {
            const { alt = '', link, src = '', title = '' } = node.attrs ?? {};
            const label = escapeBracketsInLabel(String(alt));
            const image = title ? `![${label}](${src} "${title}")` : `![${label}](${src})`;
            const { href, title: linkTitle } = (link ?? {}) as { href?: unknown; title?: unknown };

            if (typeof href !== 'string' || !href) {
                return image;
            }

            // A target with a space or a parenthesis stands between angle brackets and a quote in the title
            // behind a backslash; where even that is read as another link, the image is written without one.
            const caption = typeof linkTitle === 'string' && linkTitle ? linkTitle : null;
            const target = /[\s()]/.test(href) ? `<${href}>` : href;
            const linked = `[${image}](${target}${caption === null ? '' : ` "${caption.replaceAll('"', '\\"')}"`})`;

            return readsAsLinkedImage(linked, href, caption) ? linked : image;
        },
    }),
    VariableHighlight,
    TagHighlight,
    Placeholder.configure({ emptyEditorClass: 'is-editor-empty', placeholder }),
    ...createMarkdownLayer(),
    MarkdownLayout,
    MarkdownPaste,
];
