import { render, waitFor } from '@testing-library/react';
import { createRef } from 'react';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { MarkdownEditorField, type MarkdownEditorFieldHandle } from './markdown-editor-field';
import { setupEditorJsdom } from './markdown-editor-test-setup';
import { nextVariableRange } from './markdown-editor-variable-syntax';

beforeAll(setupEditorJsdom);

describe('nextVariableRange', () => {
    const ranges = [
        { end: 5, start: 0 },
        { end: 15, start: 10 },
    ];

    it('returns undefined when there are no occurrences', () => {
        expect(nextVariableRange([], 0, 0)).toBeUndefined();
    });

    it('advances from the selected occurrence and wraps past the last', () => {
        expect(nextVariableRange(ranges, 0, 5)).toEqual({ end: 15, start: 10 });
        expect(nextVariableRange(ranges, 10, 15)).toEqual({ end: 5, start: 0 });
    });

    it('jumps to the first occurrence at/after the caret when none is selected', () => {
        expect(nextVariableRange(ranges, 6, 6)).toEqual({ end: 15, start: 10 });
        expect(nextVariableRange(ranges, 99, 99)).toEqual({ end: 5, start: 0 });
    });
});

describe('MarkdownEditorField raw-mode handle', () => {
    it('focuses, cycles occurrences, and inserts through the textarea', () => {
        const onChange = vi.fn();
        const ref = createRef<MarkdownEditorFieldHandle>();
        const { container } = render(
            <MarkdownEditorField
                mode="raw"
                onChange={onChange}
                ref={ref}
                value={'a {{.Foo}} b {{.Foo}} c'}
            />,
        );
        const textarea = container.querySelector('textarea')!;

        ref.current!.focus();
        expect(document.activeElement).toBe(textarea);

        expect(ref.current!.selectNextUse('Nope')).toBe(false);
        expect(ref.current!.selectNextUse('Foo')).toBe(true);
        expect(textarea.value.slice(textarea.selectionStart, textarea.selectionEnd)).toBe('{{.Foo}}');

        textarea.setSelectionRange(0, 0);
        ref.current!.insertAtCursor('{{.Bar}}');
        expect(onChange).toHaveBeenCalledWith(expect.stringContaining('{{.Bar}}'));
    });

    it('insertAtCursor is a no-op when the field is disabled (a mid-save variable click cannot dirty it)', () => {
        const onChange = vi.fn();
        const ref = createRef<MarkdownEditorFieldHandle>();
        const { container } = render(
            <MarkdownEditorField
                disabled
                mode="raw"
                onChange={onChange}
                ref={ref}
                value="seed"
            />,
        );

        container.querySelector('textarea')!.setSelectionRange(0, 0);
        ref.current!.insertAtCursor('{{.Bar}}');

        expect(onChange).not.toHaveBeenCalled();
    });
});

describe('MarkdownEditorField rich-mode handle', () => {
    it('lazy-mounts the editor and delegates handle methods to it', async () => {
        const ref = createRef<MarkdownEditorFieldHandle>();
        const { container } = render(
            <MarkdownEditorField
                mode="rich"
                onChange={vi.fn()}
                ref={ref}
                value={'x {{.Foo}} y'}
            />,
        );
        await waitFor(() => expect(container.querySelector('.ProseMirror')?.textContent).toContain('Foo'));

        expect(ref.current?.selectNextUse('Foo')).toBe(true);
        expect(ref.current?.selectNextUse('Nope')).toBe(false);
    });
});

describe('MarkdownEditorField handle contract', () => {
    it('exposes focus, selectNextUse, and insertAtCursor in both modes', async () => {
        const rawRef = createRef<MarkdownEditorFieldHandle>();
        render(
            <MarkdownEditorField
                mode="raw"
                onChange={vi.fn()}
                ref={rawRef}
                value=""
            />,
        );

        const richRef = createRef<MarkdownEditorFieldHandle>();
        const { container } = render(
            <MarkdownEditorField
                mode="rich"
                onChange={vi.fn()}
                ref={richRef}
                value=""
            />,
        );
        await waitFor(() => expect(container.querySelector('.ProseMirror')).not.toBeNull());

        for (const current of [rawRef.current, richRef.current]) {
            expect(typeof current?.focus).toBe('function');
            expect(typeof current?.selectNextUse).toBe('function');
            expect(typeof current?.insertAtCursor).toBe('function');
        }
    });
});
