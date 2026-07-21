import { Editor } from '@tiptap/core';
import { beforeAll, describe, expect, it } from 'vitest';

import { createMarkdownExtensions } from './markdown-editor-extensions';
import { setupEditorJsdom } from './markdown-editor-test-setup';

beforeAll(setupEditorJsdom);

const load = (md: string) => {
    const editor = new Editor({ content: md, contentType: 'markdown', extensions: createMarkdownExtensions() });
    const json = JSON.stringify(editor.getJSON());
    const out = editor.getMarkdown();
    editor.destroy();

    return { json, out };
};

// CommonMark allows a fenced code block's opening fence to be indented up to 3 spaces. Upstream
// @tiptap/extension-code-block gated its parseMarkdown on `token.raw.startsWith('```')`, so an indented
// fence (whose raw begins with that whitespace) was dropped on load.
describe('indented fenced code blocks survive load', () => {
    it('keeps a 3-space-indented fence and its content', () => {
        const { json, out } = load('intro paragraph\n\n   ```\ntail content KEEP\n```');

        expect(out).toContain('tail content KEEP');
        expect(json).toContain('codeBlock');
        expect(out).toContain('intro paragraph');
    });

    it('keeps a 1-space-indented fence', () => {
        const { out } = load(' ```\none space BODY\n```');

        expect(out).toContain('one space BODY');
    });

    it('does not drop content following an indented fence (cascade guard)', () => {
        const { out } = load('para one\n\n   ```\ncode ALPHA\n```\n\n## Heading BETA\n\ntrailing GAMMA');

        expect(out).toContain('code ALPHA');
        expect(out).toContain('Heading BETA');
        expect(out).toContain('trailing GAMMA');
    });
});
