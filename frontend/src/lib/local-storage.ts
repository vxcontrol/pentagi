import type { z } from 'zod';

/**
 * Read a value from `localStorage` and validate it with a Zod schema. Returns
 * `null` when the key is missing, JSON parsing fails, or the payload doesn't
 * match the schema — callers don't need to distinguish between those cases.
 *
 * `localStorage` access is wrapped in try/catch so this works in environments
 * where storage is disabled (private mode, SSR, sandboxed iframes).
 */
export function getStorageItem<T>(key: string, schema: z.ZodType<T>): null | T {
    try {
        const raw = localStorage.getItem(key);

        if (raw === null) {
            return null;
        }

        const result = schema.safeParse(JSON.parse(raw));

        return result.success ? result.data : null;
    } catch {
        return null;
    }
}

/** Remove `key` from `localStorage`. Silently no-ops on failure. */
export function removeStorageItem(key: string): void {
    try {
        localStorage.removeItem(key);
    } catch {
        /* localStorage may be unavailable */
    }
}

/** Serialize `value` to JSON and store it in `localStorage`. Silently no-ops on failure. */
export function setStorageItem(key: string, value: unknown): void {
    try {
        localStorage.setItem(key, JSON.stringify(value));
    } catch {
        /* localStorage may be unavailable */
    }
}
