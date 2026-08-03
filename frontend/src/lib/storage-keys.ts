/**
 * Separator joining the namespace tag and the URL path inside a storage key.
 * Reads as "for" — e.g. `table_4_/flows` is "table for /flows". Exported so
 * the legacy-key migration helpers can reuse the exact same delimiter rather
 * than re-stating it (a desync between the two would silently break
 * migrations).
 */
export const STORAGE_KEY_SEPARATOR = '_4_';

/**
 * Discrete storage namespaces. Every list/detail page persists at most three
 * slots — `table` (filter + sorting + columnVisibility + pageSize as one
 * JSON object), `period` (dashboard-only time window), and `viewOptions`
 * (FileManager-style screens that aren't backed by `DataTable`). Add new
 * namespaces here rather than passing raw strings to {@link getStorageKey}.
 */
export type LocalStorageKeyType = 'period' | 'table' | 'viewOptions';

/** Dashboard analytics time window. Lives outside the unified table slot. */
export function getPeriodStorageKey(urlPath: string): string {
    return getStorageKey('period', urlPath);
}

/**
 * Build a storage key from a `type` tag and the URL path that owns the slot.
 *
 * Format: `${type}_4_${urlPath}` (where `_4_` reads as "for", linking the
 * type to the path it scopes).
 *
 * `urlPath` is required: callers must source it from `useLocation()` (or
 * pass an override for nested routes), not from a global `location` read.
 * That keeps this module pure and reactive — when react-router navigates,
 * a hook that depends on `pathname` re-evaluates the key automatically.
 */
export function getStorageKey(type: LocalStorageKeyType, urlPath: string): string {
    return `${type}${STORAGE_KEY_SEPARATOR}${urlPath}`;
}

/**
 * Unified per-page table state slot — one JSON object holds filter, sorting,
 * column visibility, and page size. Replaces the older four-key fan-out
 * (`column_4_`, `sorting_4_`, `filter_4_`, `page_4_`); see
 * `migrateLegacyTableState` in `table-state.ts` for the one-shot reader of
 * those legacy slots.
 */
export function getTableStorageKey(urlPath: string): string {
    return getStorageKey('table', urlPath);
}

/**
 * Returns the first non-empty segment of a path with a leading `/`.
 *
 * Used by detail pages to derive the parent list's path without hardcoding
 * it. Example:
 *   `/flows`              → `/flows`
 *   `/flows/abc-123`      → `/flows`
 *   `/knowledges/abc/foo` → `/knowledges`
 *   `/`                   → `''`
 *
 * Limitation: this only walks one level deep, so it will misidentify the
 * parent for nested lists like `/admin/flows/:id` (returns `/admin` instead
 * of `/admin/flows`). All current list pages live at the top level, but if
 * a nested list ships, the corresponding detail page must hardcode its
 * parent path explicitly instead of using this helper.
 */
export function getTopLevelPath(pathname: string): string {
    const firstSegment = pathname.split('/').filter(Boolean)[0];

    return firstSegment ? `/${firstSegment}` : '';
}

/**
 * View options for `FileManager`-style screens (currently `/resources`).
 * Lives outside the unified `table` slot.
 */
export function getViewOptionsStorageKey(urlPath: string): string {
    return getStorageKey('viewOptions', urlPath);
}
