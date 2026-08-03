import type { Node as PMNode } from '@tiptap/pm/model';

import { Extension } from '@tiptap/core';
import { Plugin, PluginKey } from '@tiptap/pm/state';
import { Decoration, DecorationSet } from '@tiptap/pm/view';

import { collectInlineMatches } from './markdown-editor-inline-scan';
import { VARIABLE_RE, variableUseRegex } from './markdown-editor-variable-syntax';

// View-only decorations, never a node/mark: a mark would alter the document (corrupting save) and defeat the
// variables side-panel, which inserts {{.X}} as plain text and finds its uses by scanning, not node identity.
const variableHighlightKey = new PluginKey('variableHighlight');

const buildDecorations = (doc: PMNode): DecorationSet =>
    DecorationSet.create(
        doc,
        collectInlineMatches(doc, VARIABLE_RE).map(({ from, to }) =>
            Decoration.inline(from, to, { class: 'template-variable' }),
        ),
    );

export const findVariableOccurrences = (doc: PMNode, variable: string): { from: number; to: number }[] => {
    const probe = variableUseRegex(variable);

    return collectInlineMatches(doc, VARIABLE_RE)
        .filter(({ text }) => probe.test(text))
        .map(({ from, to }) => ({ from, to }));
};

export const VariableHighlight = Extension.create({
    addProseMirrorPlugins() {
        return [
            new Plugin({
                key: variableHighlightKey,
                props: {
                    decorations(state) {
                        return variableHighlightKey.getState(state);
                    },
                },
                state: {
                    apply: (tr, old: DecorationSet) => (tr.docChanged ? buildDecorations(tr.doc) : old),
                    init: (_config, { doc }) => buildDecorations(doc),
                },
            }),
        ];
    },
    name: 'variableHighlight',
});
