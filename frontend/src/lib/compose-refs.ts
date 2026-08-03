import type { Ref, RefCallback } from 'react';

/**
 * Merge several refs (callback or object) into one callback ref, so a single element can satisfy every owner.
 * Wrap in `useMemo(() => composeRefs(a, b), [a, b])` at the call site to keep the composed ref stable.
 */
export function composeRefs<T>(...refs: (Ref<T> | undefined)[]): RefCallback<T> {
    return (node) => {
        for (const ref of refs) {
            assignRef(ref, node);
        }
    };
}

function assignRef<T>(ref: Ref<T> | undefined, value: null | T): void {
    if (typeof ref === 'function') {
        ref(value);
    } else if (ref) {
        (ref as { current: null | T }).current = value;
    }
}
