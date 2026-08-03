import type { ReactNode } from 'react';

import { renderHook } from '@testing-library/react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import { useTableQueryFilterReader } from './use-table-query-filter';

const STORAGE_KEY = 'table_4_/flows';
const SHORT_DEBOUNCE_MS = 5;

beforeEach(() => {
    localStorage.clear();
});

afterEach(() => {
    localStorage.clear();
});

describe('useTableQueryFilterReader', () => {
    it('observes the URL filter without writing storage', () => {
        const Wrapper = ({ children }: { children: ReactNode }) => (
            <MemoryRouter initialEntries={['/flows/abc?q=alpha']}>{children}</MemoryRouter>
        );
        const { result } = renderHook(() => useTableQueryFilterReader({ debounceMs: SHORT_DEBOUNCE_MS }), {
            wrapper: Wrapper,
        });

        expect(result.current.filter).toBe('alpha');
        expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
    });

    it('does not replay a stored filter into a clean URL', () => {
        localStorage.setItem(STORAGE_KEY, JSON.stringify({ filter: 'alpha' }));
        const Wrapper = ({ children }: { children: ReactNode }) => (
            <MemoryRouter initialEntries={['/flows/abc']}>{children}</MemoryRouter>
        );
        const { result } = renderHook(
            () => {
                const reader = useTableQueryFilterReader({
                    debounceMs: SHORT_DEBOUNCE_MS,
                });
                const { search } = useLocation();

                return { ...reader, search };
            },
            { wrapper: Wrapper },
        );

        // A shared `/flows/abc` link explicitly clears the filter on entry —
        // detail-page subscribers must not inject the previous tab's `?q=`
        // back into the URL.
        expect(result.current.search).toBe('');
        expect(result.current.filter).toBe('');
    });
});
