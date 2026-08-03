import type { Node as PMNode } from '@tiptap/pm/model';

import { Extension } from '@tiptap/core';
import { Plugin, PluginKey } from '@tiptap/pm/state';
import { Decoration, DecorationSet } from '@tiptap/pm/view';

import { collectInlineMatches } from './markdown-editor-inline-scan';

// Highlights xml-like tags (<container_environment>, </language_policy>, <specialist name="x">)
// as VIEW-ONLY decorations — same byte-neutral approach as editor-variable-highlight. A node would
// have to tell a structural tag from a tag NAME quoted inline as documentation (the prompts do both,
// in prose AND code), which can't be done reliably and risks corrupting the template on save; a
// decoration colours any tag-shaped text wherever it appears and never touches the document.
const tagHighlightKey = new PluginKey('tagHighlight');

// `[^<>]` for the attribute span keeps the scan linear and stops at the next angle bracket.
export const TAG_RE = /<\/?[a-zA-Z][\w-]*(?:\s[^<>]*)?\/?>/g;

const buildDecorations = (doc: PMNode): DecorationSet =>
    DecorationSet.create(
        doc,
        collectInlineMatches(doc, TAG_RE).map(({ from, to }) => Decoration.inline(from, to, { class: 'template-tag' })),
    );

export const TagHighlight = Extension.create({
    addProseMirrorPlugins() {
        return [
            new Plugin({
                key: tagHighlightKey,
                props: {
                    decorations(state) {
                        return tagHighlightKey.getState(state);
                    },
                },
                state: {
                    apply: (tr, old: DecorationSet) => (tr.docChanged ? buildDecorations(tr.doc) : old),
                    init: (_config, { doc }) => buildDecorations(doc),
                },
            }),
        ];
    },
    name: 'tagHighlight',
});
