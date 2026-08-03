import { Editor } from '@tiptap/core';

import { createMarkdownExtensions } from './markdown-editor-extensions';

// jsdom has no layout engine, so a mounted tiptap/ProseMirror view throws when it reads coordinates. Stub the
// three surfaces it touches (elementFromPoint + Range rects). Call from a test's beforeAll.
export const setupEditorJsdom = (): void => {
    document.elementFromPoint = () => null;
    const rect = { bottom: 0, height: 0, left: 0, right: 0, toJSON: () => ({}), top: 0, width: 0, x: 0, y: 0 };
    Range.prototype.getBoundingClientRect = () => rect as DOMRect;
    Range.prototype.getClientRects = () =>
        ({ item: () => null, length: 0, [Symbol.iterator]: [][Symbol.iterator] }) as unknown as DOMRectList;
};

export const roundTrip = (markdown: string): string => {
    const editor = new Editor({ content: markdown, contentType: 'markdown', extensions: createMarkdownExtensions() });
    const out = editor.getMarkdown();
    editor.destroy();

    return out;
};

// Structural node types whose counts must survive a round-trip. paragraph/text/hardBreak are excluded: benign
// reflow (a paragraph splitting, whitespace) changes those without losing content, whereas a change in a
// structural count is a real downgrade — a dropped code block, list item, table cell, heading. Complements the
// word-multiset oracle (which a structure-only downgrade can pass while every word survives).
const STRUCTURAL_TYPES = new Set([
    'blockquote',
    'bulletList',
    'codeBlock',
    'heading',
    'horizontalRule',
    'image',
    'listItem',
    'orderedList',
    'table',
    'tableCell',
    'tableHeader',
    'tableRow',
    'taskItem',
    'taskList',
]);

export const structuralCounts = (markdown: string): Record<string, number> => {
    const editor = new Editor({ content: markdown, contentType: 'markdown', extensions: createMarkdownExtensions() });
    const counts: Record<string, number> = {};

    editor.state.doc.descendants((node) => {
        if (STRUCTURAL_TYPES.has(node.type.name)) {
            counts[node.type.name] = (counts[node.type.name] ?? 0) + 1;
        }
    });
    editor.destroy();

    return counts;
};
