import '@testing-library/jest-dom/vitest';

import { cleanup, configure } from '@testing-library/react';
import { afterEach } from 'vitest';

// Product markup carries `data-slot`, the shadcn anatomy attribute, and never `data-testid`.
// Pointing the query at it means a stub that reaches for `data-testid` stops being found.
configure({ testIdAttribute: 'data-slot' });

// jsdom doesn't implement `Element.prototype.scrollIntoView` — components
// that call it from effects (e.g. roving-focus + scroll into view) crash in
// tests without this no-op polyfill.
if (typeof Element !== 'undefined' && !Element.prototype.scrollIntoView) {
    Element.prototype.scrollIntoView = function scrollIntoView() {
        /* no-op for jsdom */
    };
}

// Radix ScrollArea (used inside the navigation sheet) calls
// `new ResizeObserver(...)` from a layout effect; jsdom doesn't ship one.
// The component degrades gracefully — a no-op observer never invokes the
// callback, which is fine for tests that don't assert on overflow state.
if (typeof globalThis.ResizeObserver === 'undefined') {
    class NoopResizeObserver implements ResizeObserver {
        disconnect() {
            /* no-op */
        }
        observe() {
            /* no-op */
        }
        unobserve() {
            /* no-op */
        }
    }
    globalThis.ResizeObserver = NoopResizeObserver as unknown as typeof ResizeObserver;
}

// Radix Select gates pointer events behind `hasPointerCapture` / `setPointerCapture`,
// which jsdom doesn't implement. Without these no-ops, `userEvent.click` on a
// SelectTrigger throws "target.hasPointerCapture is not a function" mid-render.
if (typeof Element !== 'undefined' && !Element.prototype.hasPointerCapture) {
    Element.prototype.hasPointerCapture = function hasPointerCapture() {
        return false;
    };
    Element.prototype.setPointerCapture = function setPointerCapture() {
        /* no-op for jsdom */
    };
    Element.prototype.releasePointerCapture = function releasePointerCapture() {
        /* no-op for jsdom */
    };
}

// React Testing Library leaves rendered nodes attached to `document.body`
// after each test. Without this, tests would leak DOM state into each other
// — e.g. two `render(<X />)` calls would both end up on screen at once.
afterEach(() => {
    cleanup();
});
