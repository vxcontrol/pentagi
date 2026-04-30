import { useEffect, useState } from 'react';

/**
 * Returns a debounced version of `value` that updates only after `delayMs` of inactivity.
 *
 * The hook owns its timer entirely: every change to `value` schedules a new update and
 * cancels the previous one. When the component unmounts, the pending timer is cleared,
 * so callers do not need to manage cancellation themselves.
 */
export const useDebouncedValue = <Value>(value: Value, delayMs: number): Value => {
    const [debouncedValue, setDebouncedValue] = useState<Value>(value);

    useEffect(() => {
        const timeoutId = window.setTimeout(() => {
            setDebouncedValue(value);
        }, delayMs);

        return () => {
            window.clearTimeout(timeoutId);
        };
    }, [value, delayMs]);

    return debouncedValue;
};
