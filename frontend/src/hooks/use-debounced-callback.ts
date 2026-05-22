import { useEffect, useState } from 'react';

import { useLatestRef } from './use-latest-ref';

/**
 * A function that defers invoking `fn` until `delayMs` have elapsed since the
 * last call. The returned function is stable across renders and exposes
 * imperative escape hatches for callers that need to override the schedule.
 */
export interface DebouncedCallback<Args extends unknown[]> {
    (...args: Args): void;
    /**
     * Drop any pending call. After this, the next invocation starts a fresh
     * timer. Idempotent — safe to call from `handleClear`, route transitions,
     * external-sync effects, or anywhere else that needs to override the
     * schedule.
     */
    cancel: () => void;
    /**
     * Whether a call is currently scheduled but not yet dispatched. Useful
     * for tests; rarely needed in product code.
     */
    isPending: () => boolean;
}

/**
 * Same idea as `lodash.debounce` / `use-debounce`, but small and
 * dependency-free: returns a stable callback that defers invoking `fn` until
 * `delayMs` have elapsed since the most recent call. New calls reset the
 * timer; the most recent arguments are the ones eventually dispatched.
 *
 * Why a hook (instead of inlining `setTimeout`):
 * - `cancel()` is a single, synchronous call site — usable from `handleClear`,
 *   external-sync effects, or anywhere else that needs to override the
 *   pending dispatch. There is no opportunity for a stale timer to fire
 *   after the cancel, which was the entire class of races in our previous
 *   filter implementation.
 * - The returned function identity is stable across renders, so it can be
 *   passed to memoised children and listed in effect deps without churn.
 * - `fn` is captured in a latest-ref, so the dispatched call always sees the
 *   most recent closure (props/state at fire time, not at schedule time).
 *   This is the standard React idiom for stashing an event-like handler.
 * - Unmount cleanup is automatic — `useEffect` cancels any pending timer
 *   so dispatched functions can't land in a torn-down tree.
 *
 * Tradeoff vs `useDeferredValue`: `useDeferredValue` is React's preferred
 * tool for deferring *rendering* work (a memoised list re-renders later
 * while the input stays snappy). It does not defer side-effects. Reach for
 * `useDebouncedCallback` whenever you need to throttle imperative work that
 * leaves React's render tree — URL writes, network requests, persistence —
 * and `useDeferredValue` for the render pipeline itself.
 *
 * Implementation note: the entire API object is built once inside a
 * `useState` lazy initialiser, with the timer handle living in a *closure
 * variable* (`timerId`) rather than a React `useRef`. This keeps the
 * returned `debounced` reference stable forever and side-steps the React
 * Compiler rule that flags reading refs during render — there are no refs
 * to read, only a private variable closed over by the three methods.
 */
export function useDebouncedCallback<Args extends unknown[]>(
    fn: (...args: Args) => void,
    delayMs: number,
): DebouncedCallback<Args> {
    const fnReference = useLatestRef(fn);
    const delayReference = useLatestRef(delayMs);

    const [debounced] = useState<DebouncedCallback<Args>>(() => {
        let timerId: null | number = null;

        const cancel = () => {
            if (timerId !== null) {
                window.clearTimeout(timerId);
                timerId = null;
            }
        };

        const isPending = () => timerId !== null;

        const invoker = (...args: Args) => {
            cancel();
            timerId = window.setTimeout(() => {
                timerId = null;
                fnReference.current(...args);
            }, delayReference.current);
        };

        return Object.assign(invoker, { cancel, isPending });
    });

    // Cancel any pending dispatch when the host component unmounts so it
    // can't fire into a torn-down tree.
    useEffect(() => debounced.cancel, [debounced]);

    return debounced;
}
