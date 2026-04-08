const STORAGE_KEY_SEPARATOR = '_4_';

export type LocalStorageKeyType = 'column' | 'page' | 'sorting';

export function getColumnStorageKey(urlPath?: string): string {
    return getStorageKey('column', urlPath);
}

export function getPageStorageKey(urlPath?: string): string {
    return getStorageKey('page', urlPath);
}

export function getSortingStorageKey(urlPath?: string): string {
    return getStorageKey('sorting', urlPath);
}

/**
 * Builds a storage key from type and current page path.
 * Format: `${type}_4_${urlPath}`
 * If urlPath is not passed, uses window.location.pathname (client only).
 */
export function getStorageKey(type: LocalStorageKeyType, urlPath?: string): string {
    const path = urlPath ?? location?.pathname ?? '';

    return `${type}${STORAGE_KEY_SEPARATOR}${path}`;
}
