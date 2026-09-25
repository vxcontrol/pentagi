import { useCallback, useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';

import { getLatestSearchParams } from '@/lib/url-params';

interface SearchParamTabOptions {
    fallback: string;
    /** The set of tabs grows as data arrives, so the fallback would move the reader on its own. */
    hasVolatileValues?: boolean;
    key: string;
    values: readonly string[];
}

export function useSearchParamTab({ fallback, hasVolatileValues = false, key, values }: SearchParamTabOptions) {
    const [searchParams, setSearchParams] = useSearchParams();

    const raw = searchParams.get(key);
    const named = raw && values.includes(raw) ? raw : null;

    // The fallback is a moving target while data arrives one subscription at a time, so what is
    // already on screen is held instead.
    const [held, setHeld] = useState<null | string>(null);
    const value = named ?? (held && values.includes(held) ? held : fallback);

    const setValue = useCallback(
        (next: string) => {
            setSearchParams(
                (previous) => {
                    const params = getLatestSearchParams(previous);

                    params.set(key, next);

                    return params;
                },
                { replace: true },
            );
        },
        [key, setSearchParams],
    );

    useEffect(() => {
        const remember = (next: null | string) => setHeld(next);

        remember(named ? null : value);
    });

    useEffect(() => {
        // Only when the URL says nothing: a value it does name may belong to a tab whose data has
        // not arrived yet.
        if (!hasVolatileValues || raw !== null || !value) {
            return;
        }

        setValue(value);
    }, [hasVolatileValues, raw, setValue, value]);

    return [value, setValue] as const;
}
