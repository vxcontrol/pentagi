import { type RefObject, useEffect, useRef } from 'react';

/**
 * Stash a value in a ref that's kept in sync with the latest render.
 *
 * Useful for the "stable callback" pattern: a memoized child accepts a stable
 * handler from the parent, but the handler internally needs to read the most
 * recent prop / state value without invalidating downstream memos. Reading
 * `ref.current` inside the stable handler always sees the latest value
 * because this hook re-syncs the ref after every commit.
 *
 * Don't read `ref.current` during render — use the value prop directly instead.
 * Reading the ref during render breaks React's snapshot guarantee (the ref may
 * be updated by a concurrent render before the current one commits).
 */
export const useLatestRef = <T>(value: T): RefObject<T> => {
    const ref = useRef(value);

    useEffect(() => {
        ref.current = value;
    }, [value]);

    return ref;
};
