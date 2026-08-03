const LAST_RELOAD_AT_KEY = 'pentagi:chunk-reload-at';
const RELOAD_DEBOUNCE_MS = 10_000;

// The error browsers throw when a dynamically imported chunk can't be loaded —
// usually a redeploy rotated the hashed filenames this document still references.
export function isChunkLoadError(error: unknown): boolean {
    const message = error instanceof Error ? error.message : String(error ?? '');

    return (
        /failed to fetch dynamically imported module/i.test(message) ||
        /error loading dynamically imported module/i.test(message) ||
        /importing a module script failed/i.test(message) ||
        // A server that answers a missing chunk with index.html (HTML, not JS)
        // trips the module loader's MIME check instead.
        /failed to load module script/i.test(message) ||
        /expected a javascript[- ]module script/i.test(message)
    );
}

// React reconciliation desync: it tried to remove/insert/replace a node whose real
// parent no longer matches the one it tracked, because something outside React (a
// browser extension, auto-translation) rewrote the DOM React owns. Transient — a
// reload rebuilds a consistent tree. This is the exact wording Chrome and Firefox use.
export function isDomDesyncError(error: unknown): boolean {
    const message = error instanceof Error ? error.message : String(error ?? '');

    return /the node (to be removed|before which the new node is to be inserted|to be replaced) is not a child of this node/i.test(
        message,
    );
}

// Hard-reload, debounced so the two recovery paths (vite:preloadError and the route
// error boundary) firing for one failure can't double-reload. Returns true if it reloaded.
export function reloadOnce(): boolean {
    if (Date.now() - lastReloadAt() < RELOAD_DEBOUNCE_MS) {
        return false;
    }

    try {
        sessionStorage.setItem(LAST_RELOAD_AT_KEY, String(Date.now()));
    } catch {
        /* best-effort: without storage the debounce can't persist */
    }

    window.location.reload();

    return true;
}

function lastReloadAt(): number {
    try {
        return Number(sessionStorage.getItem(LAST_RELOAD_AT_KEY)) || 0;
    } catch {
        return 0;
    }
}
