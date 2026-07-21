import type {
    JSONContent,
    MarkdownParseHelpers,
    MarkdownParseResult,
    MarkdownRendererHelpers,
    MarkdownToken,
} from '@tiptap/core';

import { Image } from '@tiptap/extension-image';
import { TaskItem, TaskList } from '@tiptap/extension-list';
import { TableCell, TableHeader, TableRow } from '@tiptap/extension-table';
import { Placeholder } from '@tiptap/extensions';
import StarterKit from '@tiptap/starter-kit';

import { HeadingAutoformat } from './markdown-editor-heading-autoformat';
import { createMarkdownLayer, TunedTable } from './markdown-editor-marked';
import { MarkdownPaste } from './markdown-editor-paste';
import { TagHighlight } from './markdown-editor-tag-highlight';
import { VariableHighlight } from './markdown-editor-variable-highlight';

const dropUnderscoreRules = (rules: { find: unknown }[]) =>
    rules.filter((rule) => !(rule.find instanceof RegExp && rule.find.source.includes('_')));

const longestBacktickRun = (text: string): number =>
    (text.match(/`+/g) ?? []).reduce((max, run) => Math.max(max, run.length), 0);

// @tiptap/extension-code-block's renderMarkdown always emits a 3-backtick fence, so a code block whose content
// contains a ``` line (a doc demonstrating fenced markdown — common in knowledge/prompt examples) re-parses as
// TWO blocks on the next load: the inner fence closes the outer one. CommonMark requires the fence to be longer
// than any backtick run inside — widen it. Otherwise identical to upstream.
const renderTunedCodeBlock = (node: JSONContent, helpers: MarkdownRendererHelpers): string => {
    const language = node.attrs?.language || '';

    if (!node.content) {
        return `\`\`\`${language}\n\n\`\`\``;
    }

    const content = helpers.renderChildren(node.content);
    const fence = '`'.repeat(Math.max(3, longestBacktickRun(content) + 1));

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
// so typing matches load. codeBlock gets the indented-fence fix above. (Underline off below.)
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

            if (extension.name === 'codeBlock') {
                return extension.extend({
                    parseMarkdown: parseTunedCodeBlock,
                    renderMarkdown: renderTunedCodeBlock,
                });
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
        codeBlock: { HTMLAttributes: { class: 'hljs' } },
        link: { autolink: true, linkOnPaste: true, openOnClick: false },
        underline: false,
    }),
    HeadingAutoformat,
    TunedTable.configure({ resizable: true }),
    TableRow,
    TableHeader,
    TableCell,
    TaskList,
    TaskItem.configure({ nested: true }),
    Image,
    VariableHighlight,
    TagHighlight,
    Placeholder.configure({ emptyEditorClass: 'is-editor-empty', placeholder }),
    ...createMarkdownLayer(),
    MarkdownPaste,
];
