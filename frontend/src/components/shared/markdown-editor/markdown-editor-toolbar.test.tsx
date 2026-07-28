import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Editor } from '@tiptap/core';
import { beforeAll, describe, expect, it } from 'vitest';

import { createMarkdownExtensions } from './markdown-editor-extensions';
import { setupEditorJsdom } from './markdown-editor-test-setup';
import { MarkdownEditorToolbar } from './markdown-editor-toolbar';

beforeAll(setupEditorJsdom);

const DOC = ['paragraph text', '', '```js', 'const a = 1;', '```'].join('\n');
const CODE_BLOCK_CARET = 20;
const PARAGRAPH_CARET = 3;

const mount = (caret: number) => {
    const element = document.createElement('div');

    document.body.append(element);

    const editor = new Editor({
        content: DOC,
        contentType: 'markdown',
        element,
        extensions: createMarkdownExtensions(),
    });

    editor.commands.setTextSelection(caret);

    const view = render(<MarkdownEditorToolbar editor={editor} />);

    return {
        cleanup: () => {
            view.unmount();
            editor.destroy();
            element.remove();
        },
        editor,
    };
};

describe('toolbar disables controls whose command reports unavailable', () => {
    it.each(['Bold', 'Italic', 'Strikethrough', 'Inline code', 'Link'])(
        '%s is disabled inside a code block and enabled in a paragraph',
        async (label) => {
            const inCode = mount(CODE_BLOCK_CARET);

            await waitFor(() => expect(screen.getByLabelText(label)).toBeDisabled());
            inCode.cleanup();

            const inParagraph = mount(PARAGRAPH_CARET);

            await waitFor(() => expect(screen.getByLabelText(label)).toBeEnabled());
            inParagraph.cleanup();
        },
    );

    it('keeps Blockquote enabled inside a code block, where it does apply', async () => {
        const mounted = mount(CODE_BLOCK_CARET);

        await waitFor(() => expect(screen.getByLabelText('Blockquote')).toBeEnabled());
        mounted.cleanup();
    });

    it('keeps Code block enabled inside a code block, where a second press unwraps it', async () => {
        const mounted = mount(CODE_BLOCK_CARET);

        await waitFor(() => expect(screen.getByLabelText('Code block')).toBeEnabled());
        mounted.cleanup();
    });

    // Turning a code block into a list WORKS, but `can().toggleBulletList()` reports false for it, because the
    // `clearNodes()` fallback inside toggleList does nothing when dispatch is undefined. Disabling on that
    // answer would take a working conversion away from the user, so these items stay enabled.
    it.each(['Bullet list', 'Ordered list', 'Task list'])(
        '%s stays enabled inside a code block, where can() under-reports it',
        async (label) => {
            const mounted = mount(CODE_BLOCK_CARET);
            const user = userEvent.setup();

            expect(mounted.editor.can().toggleBulletList()).toBe(false);

            await user.click(screen.getByLabelText(/^List:/));

            await waitFor(() =>
                expect(screen.getByRole('menuitemradio', { name: new RegExp(label) })).not.toHaveAttribute(
                    'aria-disabled',
                    'true',
                ),
            );
            mounted.cleanup();
        },
    );
});
