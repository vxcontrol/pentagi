import { Editor } from '@tiptap/core';
import { beforeAll, describe, expect, it } from 'vitest';

import { createMarkdownExtensions } from './markdown-editor-extensions';
import { hasHeaderRow } from './markdown-editor-table-commands';
import { setupEditorJsdom } from './markdown-editor-test-setup';

beforeAll(setupEditorJsdom);

const makeEditor = (content: string) =>
    new Editor({ content, contentType: 'markdown', extensions: createMarkdownExtensions() });

const firstNodePos = (editor: Editor, typeName: string): number => {
    let found = -1;

    editor.state.doc.descendants((node, pos) => {
        if (found === -1 && node.type.name === typeName) {
            found = pos;
        }
    });

    return found;
};

describe('hasHeaderRow', () => {
    it('detects a header row from a cell position inside the table', () => {
        const editor = makeEditor('| h | i |\n| --- | --- |\n| a | b |');
        const headerCell = firstNodePos(editor, 'tableHeader');

        expect(hasHeaderRow(editor, headerCell + 1)).toBe(true);

        editor.destroy();
    });

    it('returns false for a position outside any table', () => {
        const editor = makeEditor('just a paragraph');

        expect(hasHeaderRow(editor, 1)).toBe(false);

        editor.destroy();
    });

    it('returns false (never throws) for a stale position past the doc — a grip that outlived a shrinking edit', () => {
        const editor = makeEditor('| h | i |\n| --- | --- |\n| a | b |');
        const stalePos = editor.state.doc.content.size + 100;

        expect(() => hasHeaderRow(editor, stalePos)).not.toThrow();
        expect(hasHeaderRow(editor, stalePos)).toBe(false);

        editor.destroy();
    });
});
