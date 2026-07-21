import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { isChunkLoadError, isDomDesyncError, reloadOnce } from './chunk-reload';

describe('isChunkLoadError', () => {
    it('matches the dynamic-import failures browsers throw for a stale chunk', () => {
        expect(
            isChunkLoadError(new Error('Failed to fetch dynamically imported module: https://x/assets/login-AbC1.js')),
        ).toBe(true);
        expect(isChunkLoadError(new Error('error loading dynamically imported module'))).toBe(true);
        // Safari's wording differs from Chromium's.
        expect(isChunkLoadError(new Error('Importing a module script failed.'))).toBe(true);
        // A server that serves index.html for a missing chunk trips the MIME check.
        expect(
            isChunkLoadError(
                new Error(
                    'Failed to load module script: Expected a JavaScript module script but the server responded with a MIME type of "text/html".',
                ),
            ),
        ).toBe(true);
    });

    it('does not match unrelated runtime errors', () => {
        expect(isChunkLoadError(new Error("Failed to execute 'removeChild' on 'Node'"))).toBe(false);
        expect(isChunkLoadError(new TypeError('x is not a function'))).toBe(false);
        expect(isChunkLoadError('a bare string')).toBe(false);
        expect(isChunkLoadError(null)).toBe(false);
        expect(isChunkLoadError(undefined)).toBe(false);
    });
});

describe('isDomDesyncError', () => {
    it('matches the removeChild/insertBefore "not a child" reconciliation crash', () => {
        expect(
            isDomDesyncError(
                new Error(
                    "Failed to execute 'removeChild' on 'Node': The node to be removed is not a child of this node.",
                ),
            ),
        ).toBe(true);
        // React surfaces it as a DOMException (not an `instanceof Error`) — match via toString.
        expect(
            isDomDesyncError(
                new DOMException(
                    "Failed to execute 'insertBefore' on 'Node': The node before which the new node is to be inserted is not a child of this node.",
                    'NotFoundError',
                ),
            ),
        ).toBe(true);
        expect(
            isDomDesyncError(
                new DOMException(
                    "Failed to execute 'replaceChild' on 'Node': The node to be replaced is not a child of this node.",
                    'NotFoundError',
                ),
            ),
        ).toBe(true);
    });

    it('does not match chunk-load or generic errors', () => {
        expect(isDomDesyncError(new Error('Failed to fetch dynamically imported module: /assets/x.js'))).toBe(false);
        expect(isDomDesyncError(new TypeError('Cannot read properties of undefined (reading length)'))).toBe(false);
        expect(isDomDesyncError(null)).toBe(false);
    });
});

describe('reload guard', () => {
    let reloadSpy: ReturnType<typeof vi.fn>;
    const originalLocation = window.location;
    let now = 1_000_000;

    beforeEach(() => {
        sessionStorage.clear();
        now = 1_000_000;
        vi.spyOn(Date, 'now').mockImplementation(() => now);
        reloadSpy = vi.fn();
        // jsdom's `location.reload` is non-configurable, so swap the whole
        // `location` object — the standard jsdom workaround for asserting reloads.
        Object.defineProperty(window, 'location', {
            configurable: true,
            value: { ...originalLocation, reload: reloadSpy },
            writable: true,
        });
    });

    afterEach(() => {
        vi.restoreAllMocks();
        Object.defineProperty(window, 'location', {
            configurable: true,
            value: originalLocation,
            writable: true,
        });
        sessionStorage.clear();
    });

    it('reloads, then debounces a rapid second attempt so it cannot loop', () => {
        expect(reloadOnce()).toBe(true);
        now += 500;
        expect(reloadOnce()).toBe(false);
        expect(reloadSpy).toHaveBeenCalledTimes(1);
    });

    it('reloads again once the debounce window has passed (a later deploy)', () => {
        expect(reloadOnce()).toBe(true);
        now += 11_000;
        expect(reloadOnce()).toBe(true);
        expect(reloadSpy).toHaveBeenCalledTimes(2);
    });
});
