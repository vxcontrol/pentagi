import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { InlineEditInput } from './inline-edit-input';

const renderInput = (busy: boolean) => {
    const onCancel = vi.fn();
    const onSave = vi.fn();

    render(
        <InlineEditInput
            busy={busy}
            defaultValue="Alpha"
            onCancel={onCancel}
            onSave={onSave}
        />,
    );

    return { input: screen.getByRole('textbox'), onCancel, onSave };
};

describe('InlineEditInput', () => {
    it('saves on Enter and cancels on Escape while idle', () => {
        const { input, onCancel, onSave } = renderInput(false);

        fireEvent.keyDown(input, { key: 'Enter' });
        expect(onSave).toHaveBeenCalledTimes(1);
        expect(onCancel).not.toHaveBeenCalled();

        fireEvent.keyDown(input, { key: 'Escape' });
        expect(onCancel).toHaveBeenCalledTimes(1);
    });

    it('ignores Enter while busy, matching the disabled Save button', () => {
        const { input, onSave } = renderInput(true);

        expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled();

        fireEvent.keyDown(input, { key: 'Enter' });
        expect(onSave).not.toHaveBeenCalled();
    });

    it('ignores Escape while busy, matching the disabled Cancel button', () => {
        const { input, onCancel } = renderInput(true);

        expect(screen.getByRole('button', { name: 'Cancel' })).toBeDisabled();

        fireEvent.keyDown(input, { key: 'Escape' });
        expect(onCancel).not.toHaveBeenCalled();
    });

    it('keeps the input editable while busy', () => {
        const { input } = renderInput(true);

        expect(input).not.toBeDisabled();
    });
});
