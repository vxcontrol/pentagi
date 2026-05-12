import { useCallback, useState } from 'react';

import { useLatestRef } from '@/hooks/use-latest-ref';

/**
 * Radix-style controllable state. When `controlled` is `undefined` the hook
 * owns the state; otherwise the parent does and we forward updates via
 * `onChange`. `onChange` always fires so fully-controlled consumers can
 * observe every set (e.g. for logging).
 *
 * Mirrors `@radix-ui/react-use-controllable-state` without pulling in the
 * dependency for a couple of state slots.
 */
export const useControllable = <T>(controlled: T | undefined, defaultValue: T, onChange?: (value: T) => void) => {
    const [internal, setInternal] = useState<T>(defaultValue);
    const isControlled = controlled !== undefined;
    const value = isControlled ? (controlled as T) : internal;

    const onChangeRef = useLatestRef(onChange);

    const set = useCallback(
        (next: T) => {
            if (!isControlled) {
                setInternal(next);
            }

            onChangeRef.current?.(next);
        },
        [isControlled, onChangeRef],
    );

    return [value, set] as const;
};
