import { Editor, getMarkRange } from '@tiptap/core';
import { beforeAll, describe, expect, it } from 'vitest';

import { createMarkdownExtensions } from './markdown-editor-extensions';
import { setupEditorJsdom } from './markdown-editor-test-setup';

beforeAll(setupEditorJsdom);

describe('link handle popover key is stable while typing inside a link (LINK-REMOUNT)', () => {
    it('keeps range.from fixed as range.to grows, so a from-based key does not remount', () => {
        const editor = new Editor({
            content: '[label](https://example.com)',
            contentType: 'markdown',
            extensions: createMarkdownExtensions(),
        });
        const linkType = editor.schema.marks.link!;

        editor.commands.setTextSelection(3);
        const before = getMarkRange(editor.state.selection.$from, linkType);

        editor.commands.insertContent('X');
        const after = getMarkRange(editor.state.selection.$from, linkType);

        editor.destroy();

        expect(before && after).toBeTruthy();
        expect(after!.from).toBe(before!.from);
        // range.to grew by the inserted char — the old `${from}-${to}` key would have remounted every keystroke.
        expect(after!.to).toBe(before!.to + 1);
    });
});
