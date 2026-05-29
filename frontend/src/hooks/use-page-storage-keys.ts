import { useMemo } from 'react';
import { useLocation } from 'react-router-dom';

import { getPeriodStorageKey, getTableStorageKey, getTopLevelPath, getViewOptionsStorageKey } from '@/lib/storage-keys';

interface PageStorageKeys {
    /** Dashboard analytics time window — single string value, no schema merge. */
    period: string;
    /** Unified per-page table state (filter + sorting + columnVisibility + pageSize). */
    table: string;
    /** FileManager-style view options (folder-first toggle, expanded dirs, etc.). */
    viewOptions: string;
}

interface UsePageStorageKeysOptions {
    /**
     * Optional override for the path used to build the keys. Pass when the
     * caller wants to share a storage bucket with a different route — e.g.
     * a detail page (`/flows/:id`) keying into the parent list's slot
     * (`/flows`).
     *
     * When omitted, the live `useLocation().pathname` is used as-is.
     */
    pathname?: string;
    /**
     * If `true`, the effective path is the top-level segment of `pathname`
     * (e.g. `/flows/abc-123` → `/flows`). Convenience for detail pages that
     * want to inherit their list's storage slot without spelling it out.
     */
    useTopLevel?: boolean;
}

/**
 * Resolve the per-route localStorage keys (table, period, viewOptions) for
 * the current location, reactively.
 *
 * Replaces the older pattern of calling individual `get*StorageKey()`
 * functions without an argument, which read the global `location.pathname`.
 * The global read is not reactive to react-router navigation and is unsafe
 * to call during module load (e.g. tests, SSR). Routing through
 * `useLocation` makes the result re-evaluate whenever the active route
 * changes.
 *
 * Each list page now stores at most **one** key — `table_4_<path>` — that
 * bundles filter + sorting + columnVisibility + pageSize. Old `column_4_`,
 * `sorting_4_`, `filter_4_`, `page_4_` slots are migrated into the unified
 * key on first mount (see `migrateLegacyTableState`).
 */
export function usePageStorageKeys(options: UsePageStorageKeysOptions = {}): PageStorageKeys {
    const { pathname: livePathname } = useLocation();
    const { pathname: override, useTopLevel = false } = options;

    return useMemo(() => {
        const source = override ?? livePathname;
        const effective = useTopLevel ? getTopLevelPath(source) : source;

        return {
            period: getPeriodStorageKey(effective),
            table: getTableStorageKey(effective),
            viewOptions: getViewOptionsStorageKey(effective),
        };
    }, [livePathname, override, useTopLevel]);
}
