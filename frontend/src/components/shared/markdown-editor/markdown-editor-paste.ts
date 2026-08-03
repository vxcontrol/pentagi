import { Extension } from '@tiptap/core';
import { Plugin, PluginKey } from '@tiptap/pm/state';

// @tiptap/markdown parses markdown for load/insertContent but never for the clipboard, so a paste of block
// markdown (# ## - 1. | >) would land as literal text (only StarterKit's inline mark paste-rules fire).
// Route markdown-looking plain-text pastes through the same markdown layer as load; defer to ProseMirror's
// own path for rich sources — an in-editor copy (carries `data-pm-slice`), web/Office HTML (block tags), or
// inline-formatted HTML whose text carries no markdown — so their fidelity survives, and keep pastes inside
// code literal.
const RICH_HTML_BLOCK = /<(?:h[1-6]|ul|ol|li|table|thead|tbody|tr|td|th|blockquote|pre|img|hr)\b/i;

// Every cue must be linear-time: they run synchronously in handlePaste on every paste. The link cue bounds
// its two spans with `[^\]]+`/`[^)]+` (not `.+`) — a double-`.+` catastrophically backtracks (ReDoS) on a
// large bracket-heavy non-link paste and freezes the tab. Keep any new cue single-quantifier / negated-class.
const MARKDOWN_CUES = [
    /^#{1,6}\s/m, // heading
    /\*\*[^*]+\*\*/, // bold
    /\[[^\]]+\]\([^)]+\)/, // link
    /^[-*+]\s/m, // bullet item
    /^\d+\.\s/m, // ordered item
    /^ {0,3}(?:```|~~~)/m, // fence
    /^\|/m, // table row
    /^>\s/m, // blockquote
];

const isMarkdownLike = (text: string): boolean => MARKDOWN_CUES.some((cue) => cue.test(text));

export const shouldParseMarkdownOnPaste = (text: string, html: string, isCodeContext: boolean): boolean => {
    if (!text.trim() || isCodeContext) {
        return false;
    }

    if (html.includes('data-pm-slice') || RICH_HTML_BLOCK.test(html)) {
        return false;
    }

    // Inline-only rich HTML (a Google Docs/Word/mail single paragraph: <b>, <a href>, styled spans) whose
    // text/plain carries no markdown: defer to ProseMirror's native HTML parse — it runs the clipboard HTML
    // through the schema, so bold/italic/links survive as marks. Markdown-looking text still wins the
    // markdown parse: a VS Code copy of markdown source arrives wrapped in syntax-color spans, and parsing
    // its text/plain (not the span noise) is the point of this plugin.
    if (html && !isMarkdownLike(text)) {
        return false;
    }

    return true;
};

export const MarkdownPaste = Extension.create({
    addProseMirrorPlugins() {
        const { editor } = this;

        return [
            new Plugin({
                key: new PluginKey('markdownPaste'),
                props: {
                    handlePaste(view, event) {
                        const text = event.clipboardData?.getData('text/plain') ?? '';
                        const html = event.clipboardData?.getData('text/html') ?? '';
                        const isCodeContext =
                            view.state.selection.$from.parent.type.spec.code || editor.isActive('code');

                        if (!shouldParseMarkdownOnPaste(text, html, isCodeContext)) {
                            return false;
                        }

                        return editor.commands.insertContent(text, { contentType: 'markdown' });
                    },
                },
            }),
        ];
    },
    name: 'markdownPaste',
});
