import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { InlineEditInput } from './inline-edit-input';

const renderInput = (busy: boolean, defaultValue = 'Alpha') => {
    const onCancel = vi.fn();
    const onSave = vi.fn();

    render(
        <InlineEditInput
            busy={busy}
            defaultValue={defaultValue}
            onCancel={onCancel}
            onSave={onSave}
        />,
    );

    return { input: screen.getByRole('textbox'), onCancel, onSave };
};

const saveButton = () => screen.getByRole('button', { name: 'Save' });

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

    it('refuses to save a value that is only whitespace', () => {
        const { input, onSave } = renderInput(false);

        fireEvent.change(input, { target: { value: '   ' } });

        expect(saveButton()).toBeDisabled();

        fireEvent.keyDown(input, { key: 'Enter' });
        expect(onSave).not.toHaveBeenCalled();
    });

    it('offers save again once the value carries text', () => {
        const { input, onSave } = renderInput(false);

        fireEvent.change(input, { target: { value: '   ' } });
        fireEvent.change(input, { target: { value: '  Beta  ' } });

        expect(saveButton()).not.toBeDisabled();

        fireEvent.keyDown(input, { key: 'Enter' });
        expect(onSave).toHaveBeenCalledTimes(1);
    });

    it('starts with save unavailable when mounted empty', () => {
        renderInput(false, '');

        expect(saveButton()).toBeDisabled();
    });

    it('still cancels a whitespace value on Escape', () => {
        const { input, onCancel } = renderInput(false);

        fireEvent.change(input, { target: { value: '   ' } });
        fireEvent.keyDown(input, { key: 'Escape' });

        expect(onCancel).toHaveBeenCalledTimes(1);
    });
});
