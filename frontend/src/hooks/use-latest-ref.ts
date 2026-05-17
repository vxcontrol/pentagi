import { type RefObject, useEffect, useRef } from 'react';

/**
 * Stash a value in a ref that's kept in sync with the latest render.
 *
 * Useful for the "stable callback" pattern: a memoized child accepts a stable
 * handler from the parent, but the handler internally needs to read the most
 * recent prop / state value without invalidating downstream memos. Reading
 * `ref.current` inside the stable handler sees the value committed in the
 * most recent successful render.
 *
 * **Timing caveat — one-commit lag for synchronous reads.** The ref is
 * synced inside a passive `useEffect`. This means handlers that fire
 * *during* a commit (e.g. Radix's `onCloseAutoFocus`, which it dispatches
 * from its own layout-effect in the same flush) read the *previous* render's
 * value, because the sync effect from the current render hasn't run yet.
 * For those cases mutate a `useRef` directly in render — write
 * `const ref = useRef(value); ref.current = value;` — instead of using this
 * hook. Render-time mutation is safe so long as nothing *reads* the ref
 * during render (only handlers and effects do).
 *
 * For asynchronous reads — pointer/key events, timers, network callbacks —
 * the lag never matters because at least one commit cycle has finished by
 * the time the handler fires. That's the common case this hook is built for.
 *
 * Don't read `ref.current` during render — use the value prop directly instead.
 * Reading the ref during render breaks React's snapshot guarantee (the ref may
 * be updated by a concurrent render before the current one commits).
 */
export function useLatestRef<T>(value: T): RefObject<T> {
    const ref = useRef(value);

    useEffect(() => {
        ref.current = value;
    }, [value]);

    return ref;
}
