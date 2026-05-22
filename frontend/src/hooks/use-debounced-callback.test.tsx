import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useDebouncedCallback } from './use-debounced-callback';

describe('useDebouncedCallback', () => {
    beforeEach(() => {
        vi.useFakeTimers();
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('defers invocation until delayMs of silence', () => {
        const fn = vi.fn();
        const { result } = renderHook(() => useDebouncedCallback(fn, 100));

        act(() => {
            result.current('a');
            result.current('b');
            result.current('c');
        });

        expect(fn).not.toHaveBeenCalled();

        act(() => {
            vi.advanceTimersByTime(100);
        });

        // Only the latest call lands; intermediate ones are dropped.
        expect(fn).toHaveBeenCalledExactlyOnceWith('c');
    });

    it('cancel() drops the pending dispatch', () => {
        const fn = vi.fn();
        const { result } = renderHook(() => useDebouncedCallback(fn, 100));

        act(() => {
            result.current('x');
            result.current.cancel();
            vi.advanceTimersByTime(500);
        });

        expect(fn).not.toHaveBeenCalled();
    });

    it('isPending() flips with the timer', () => {
        const fn = vi.fn();
        const { result } = renderHook(() => useDebouncedCallback(fn, 100));

        expect(result.current.isPending()).toBe(false);

        act(() => {
            result.current('a');
        });

        expect(result.current.isPending()).toBe(true);

        act(() => {
            vi.advanceTimersByTime(100);
        });

        expect(result.current.isPending()).toBe(false);
    });

    it('reads the latest `fn` closure at fire time, not at schedule time', () => {
        // Mirrors the production use case: a stale fn would close over an
        // old prop. The debounced callback identity stays stable across
        // renders, but the dispatched fn must be the freshest one.
        const calls: string[] = [];
        const { rerender, result } = renderHook(({ tag }: { tag: string }) => useDebouncedCallback(() => calls.push(tag), 100), {
            initialProps: { tag: 'old' },
        });

        const initialDebounced = result.current;

        act(() => {
            result.current();
        });

        rerender({ tag: 'new' });

        // Same callback identity across renders — caller can pass it to
        // memoised children or list it in effect deps without churn.
        expect(result.current).toBe(initialDebounced);

        act(() => {
            vi.advanceTimersByTime(100);
        });

        expect(calls).toEqual(['new']);
    });

    it('cancels pending dispatch on unmount so dead components never fire', () => {
        const fn = vi.fn();
        const { result, unmount } = renderHook(() => useDebouncedCallback(fn, 100));

        act(() => {
            result.current('a');
        });

        unmount();

        act(() => {
            vi.advanceTimersByTime(500);
        });

        expect(fn).not.toHaveBeenCalled();
    });
});
