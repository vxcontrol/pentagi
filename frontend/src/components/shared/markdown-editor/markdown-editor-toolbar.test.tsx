import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Editor } from '@tiptap/core';
import { beforeAll, describe, expect, it } from 'vitest';

import { createMarkdownExtensions } from './markdown-editor-extensions';
import { returnFocusToEditor } from './markdown-editor-focus';
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

describe('returnFocusToEditor', () => {
    const editorWithElement = () => {
        const element = document.createElement('div');

        document.body.append(element);

        return {
            cleanup: () => element.remove(),
            editor: new Editor({ content: 'text', element, extensions: createMarkdownExtensions() }),
        };
    };

    // jsdom never reports real DOM focus for the contenteditable, so the observable proof that the call reached
    // the editor is the transaction `focus()` dispatches.
    it('reaches the editor while it is alive', () => {
        const { cleanup, editor } = editorWithElement();
        let transactions = 0;

        editor.on('transaction', () => {
            transactions += 1;
        });

        returnFocusToEditor(editor)(new Event('closeAutoFocus'));

        expect(transactions).toBeGreaterThan(0);
        editor.destroy();
        cleanup();
    });

    // Radix closes a still-open layer asynchronously, so unmounting with a menu open lands this callback on a
    // destroyed editor. Without the guard `editor.commands` throws outside any test's stack: every test still
    // reads green and `vitest run` exits 1.
    it('does nothing once the editor is destroyed, instead of throwing off-stack', () => {
        const { cleanup, editor } = editorWithElement();

        editor.destroy();

        expect(() => returnFocusToEditor(editor)(new Event('closeAutoFocus'))).not.toThrow();
        cleanup();
    });
});

describe('block controls that a table cell cannot hold are disabled inside one', () => {
    const TABLE_DOC = ['| A | B |', '| --- | --- |', '| alpha | beta |'].join('\n');

    const mountWithCaretInCell = () => {
        const element = document.createElement('div');

        document.body.append(element);

        const editor = new Editor({
            content: TABLE_DOC,
            contentType: 'markdown',
            element,
            extensions: createMarkdownExtensions(),
        });
        let caret = 0;

        editor.state.doc.descendants((node, pos) => {
            if (!caret && node.type.name === 'tableCell') {
                caret = pos + 2;
            }

            return true;
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

    // A cell serialises inline, so one click silently degrades the document: a code block in a cell saves as
    // `| ``` alpha ``` |` and reloads as INLINE code, and a horizontal rule applied mid-word splits the word.
    it.each(['Code block', 'Horizontal rule'])('%s is disabled with the caret inside a cell', async (label) => {
        const mounted = mountWithCaretInCell();

        await waitFor(() => expect(screen.getByLabelText(label)).toBeDisabled());
        mounted.cleanup();
    });

    it('disables every list kind inside a cell', async () => {
        const mounted = mountWithCaretInCell();
        const user = userEvent.setup();

        await user.click(screen.getByLabelText(/^List:/));

        for (const label of ['Bullet list', 'Ordered list', 'Task list']) {
            await waitFor(() =>
                expect(screen.getByRole('menuitemradio', { name: new RegExp(label) })).toHaveAttribute(
                    'aria-disabled',
                    'true',
                ),
            );
        }

        mounted.cleanup();
    });

    it('disables every heading level inside a cell', async () => {
        const mounted = mountWithCaretInCell();
        const user = userEvent.setup();

        await user.click(screen.getByLabelText(/^Text style:/));

        await waitFor(() =>
            expect(screen.getByRole('menuitemradio', { name: /Heading 2/ })).toHaveAttribute('aria-disabled', 'true'),
        );
        mounted.cleanup();
    });
});
