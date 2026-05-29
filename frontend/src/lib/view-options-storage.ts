import { z } from 'zod';

import { getStorageItem, removeStorageItem, setStorageItem } from './local-storage';

const viewOptionsSchema = z.record(z.string(), z.boolean());

export type ViewOptionsRecord = Record<string, boolean>;

/** Read a per-page `Record<string, boolean>` view-options bag. Returns `{}` when missing/invalid. */
export const loadViewOptions = (key: string): ViewOptionsRecord => getStorageItem(key, viewOptionsSchema) ?? {};

/** Persist the bag. Empty objects clear the slot so storage stays tidy. */
export const saveViewOptions = (key: string, value: ViewOptionsRecord): void => {
    if (Object.keys(value).length === 0) {
        removeStorageItem(key);

        return;
    }

    setStorageItem(key, value);
};

/**
 * One-shot migration of pre-unification `column_4_<path>` view-options storage
 * to the new `viewOptions_4_<path>` slot. Reads the legacy key, writes the
 * payload under the new key (if any), removes the legacy entry, and returns
 * the resulting record.
 *
 * Idempotent: once the legacy key is gone, subsequent invocations short-circuit
 * to a plain `loadViewOptions`. Safe to call on every page mount.
 */
export const migrateLegacyViewOptions = (path: string, unifiedKey: string): ViewOptionsRecord => {
    const legacyKey = `column_4_${path}`;
    const legacyRaw = localStorage.getItem(legacyKey);

    if (legacyRaw === null) {
        return loadViewOptions(unifiedKey);
    }

    const legacyValue = getStorageItem(legacyKey, viewOptionsSchema);

    if (legacyValue !== null) {
        saveViewOptions(unifiedKey, { ...loadViewOptions(unifiedKey), ...legacyValue });
    }

    removeStorageItem(legacyKey);

    return loadViewOptions(unifiedKey);
};
