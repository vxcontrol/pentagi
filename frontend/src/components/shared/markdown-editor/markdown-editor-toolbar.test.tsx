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

    it('disables the list kinds a code block cannot become, on open', async () => {
        const mounted = mount(CODE_BLOCK_CARET);
        const user = userEvent.setup();

        await user.click(screen.getByLabelText(/^List:/));

        await waitFor(() =>
            expect(screen.getByRole('menuitemradio', { name: /Bullet list/ })).toHaveAttribute('aria-disabled', 'true'),
        );
        expect(screen.getByRole('menuitemradio', { name: /Ordered list/ })).toHaveAttribute('aria-disabled', 'true');
        mounted.cleanup();
    });

    it('leaves the list kinds enabled in a paragraph', async () => {
        const mounted = mount(PARAGRAPH_CARET);
        const user = userEvent.setup();

        await user.click(screen.getByLabelText(/^List:/));

        await waitFor(() =>
            expect(screen.getByRole('menuitemradio', { name: /Bullet list/ })).not.toHaveAttribute(
                'aria-disabled',
                'true',
            ),
        );
        mounted.cleanup();
    });

    // The menu asks `can()` when Radix mounts its content, so a caret move while it was closed must still be
    // reflected — a value cached at the toolbar's last render would be stale here.
    it('reflects a caret move that happened while the menu was closed', async () => {
        const mounted = mount(PARAGRAPH_CARET);
        const user = userEvent.setup();

        mounted.editor.commands.setTextSelection(CODE_BLOCK_CARET);
        await user.click(screen.getByLabelText(/^List:/));

        await waitFor(() =>
            expect(screen.getByRole('menuitemradio', { name: /Bullet list/ })).toHaveAttribute('aria-disabled', 'true'),
        );
        mounted.cleanup();
    });
});
