import { act, renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { useInlineEdit } from './use-inline-edit';

describe('useInlineEdit', () => {
    it('starts in non-editing state', () => {
        const { result } = renderHook(() => useInlineEdit({ resetKey: 'a' }));
        expect(result.current.isEditing).toBe(false);
    });

    it('startEdit flips to editing, stopEdit flips back', () => {
        const { result } = renderHook(() => useInlineEdit({ resetKey: 'a' }));

        act(() => result.current.startEdit());
        expect(result.current.isEditing).toBe(true);

        act(() => result.current.stopEdit());
        expect(result.current.isEditing).toBe(false);
    });

    it('exits edit mode when resetKey changes', () => {
        const { rerender, result } = renderHook(({ resetKey }) => useInlineEdit({ resetKey }), {
            initialProps: { resetKey: 'a' as null | string | undefined },
        });

        act(() => result.current.startEdit());
        expect(result.current.isEditing).toBe(true);

        rerender({ resetKey: 'b' });
        expect(result.current.isEditing).toBe(false);
    });

    it('keeps edit mode across rerenders with the same resetKey', () => {
        const { rerender, result } = renderHook(({ resetKey }) => useInlineEdit({ resetKey }), {
            initialProps: { resetKey: 'a' as null | string | undefined },
        });

        act(() => result.current.startEdit());
        rerender({ resetKey: 'a' });
        expect(result.current.isEditing).toBe(true);
    });

    it('treats `null` and `undefined` as distinct keys (state reset on transition)', () => {
        const { rerender, result } = renderHook(({ resetKey }) => useInlineEdit({ resetKey }), {
            initialProps: { resetKey: null as null | string | undefined },
        });

        act(() => result.current.startEdit());

        // Transition null -> undefined trips the reset because they aren't
        // strictly equal. Documenting this so callers know to keep the
        // resetKey type stable.
        rerender({ resetKey: undefined });
        expect(result.current.isEditing).toBe(false);
    });

    it('returns a ref object whose .current starts at null', () => {
        const { result } = renderHook(() => useInlineEdit({ resetKey: 'a' }));
        expect(result.current.inputRef.current).toBeNull();
    });

    it('handleDropdownCloseAutoFocus prevents default while editing', () => {
        const { result } = renderHook(() => useInlineEdit({ resetKey: 'a' }));

        act(() => result.current.startEdit());

        // `Event` is cancelable by default but `defaultPrevented` only flips
        // once `preventDefault` is called — exactly what the hook does.
        const event = new Event('autofocus', { cancelable: true });
        result.current.handleDropdownCloseAutoFocus(event);

        expect(event.defaultPrevented).toBe(true);
    });

    it('handleDropdownCloseAutoFocus is a no-op when not editing', () => {
        const { result } = renderHook(() => useInlineEdit({ resetKey: 'a' }));

        const event = new Event('autofocus', { cancelable: true });
        result.current.handleDropdownCloseAutoFocus(event);

        expect(event.defaultPrevented).toBe(false);
    });

    it('focuses and selects the input on the next animation frame after startEdit', async () => {
        const { result } = renderHook(() => useInlineEdit({ resetKey: 'a' }));

        const input = document.createElement('input');
        input.value = 'hello';
        document.body.appendChild(input);

        // Attach the ref manually — `useRef` exposes `.current` as writable
        // even though TypeScript narrows it to readonly. Casting through
        // an interface keeps the intent visible in the test.
        (result.current.inputRef as unknown as { current: HTMLInputElement }).current = input;

        act(() => result.current.startEdit());

        // requestAnimationFrame in jsdom resolves in a microtask — wait for it.
        await act(async () => {
            await new Promise((resolve) => requestAnimationFrame(() => resolve(null)));
        });

        expect(document.activeElement).toBe(input);
        expect(input.selectionStart).toBe(0);
        expect(input.selectionEnd).toBe(input.value.length);

        document.body.removeChild(input);
    });

    it('cancels the pending focus when isEditing flips off before the frame fires', async () => {
        const { result } = renderHook(() => useInlineEdit({ resetKey: 'a' }));

        const input = document.createElement('input');
        document.body.appendChild(input);
        (result.current.inputRef as unknown as { current: HTMLInputElement }).current = input;

        const otherInput = document.createElement('input');
        document.body.appendChild(otherInput);
        otherInput.focus();

        act(() => result.current.startEdit());
        act(() => result.current.stopEdit());

        await act(async () => {
            await new Promise((resolve) => requestAnimationFrame(() => resolve(null)));
        });

        // Focus must not have moved to the inline input — the effect cleanup
        // cancelled the rAF before it ran.
        expect(document.activeElement).toBe(otherInput);

        document.body.removeChild(input);
        document.body.removeChild(otherInput);
    });
});
