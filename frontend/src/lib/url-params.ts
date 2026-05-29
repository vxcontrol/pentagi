/**
 * URL query parameter keys shared across list/detail pages.
 *
 * Keep all conventional names here so the filter hook, the pagination
 * handlers, and any future feature that wants to coordinate with them agree
 * on a single source of truth — renaming a key shouldn't require grepping
 * the codebase for string literals.
 */
export const URL_PARAMS = {
    /** `?page=` — 1-based page number for list-style pagination. */
    PAGE: 'page',
    /** `?q=` — free-text filter applied by `useTableQueryFilter`. */
    QUERY: 'q',
    /**
     * `?qs=` — server-side semantic search query. Triggers a vector-store
     * lookup (e.g. `searchKnowledge`) instead of the regular list query.
     * Orthogonal to `?q=`: when both are present, the client text filter
     * further narrows the server-returned subset on the page.
     */
    SEARCH: 'qs',
} as const;

/**
 * Synthetic base for `new URL(path, base)`. `URL` rejects a path-only first
 * argument unless a base is provided; we strip the origin from the output
 * below by reading only `pathname` / `search` / `hash`. The constant is
 * intentionally bland — anything resolvable as a URL works.
 */
const URL_PARSE_BASE = 'http://_';

/**
 * Merge a relative href with an iterable of "current" query params: every key
 * that the iterable supplies is added to the href, unless the href already
 * specifies that key (then the href wins).
 *
 * Built around `URL` rather than string `split('?')` so the hash fragment
 * survives untouched and the contract holds even if `base` happens to grow
 * an `#anchor` someday.
 *
 * Examples:
 *   merge('/flows/1', new URLSearchParams('q=foo'))            // '/flows/1?q=foo'
 *   merge('/flows/1?bar=1', new URLSearchParams('bar=2&q=x'))  // '/flows/1?bar=1&q=x'
 *   merge('/flows/1#tab=logs', new URLSearchParams('q=foo'))   // '/flows/1?q=foo#tab=logs'
 */
export const mergeHrefWithSearchParams = (
    base: string,
    incoming: Iterable<[string, string]> | URLSearchParams,
): string => {
    const url = new URL(base, URL_PARSE_BASE);

    const source = incoming instanceof URLSearchParams ? incoming.entries() : incoming;

    for (const [key, value] of source) {
        if (!url.searchParams.has(key)) {
            url.searchParams.set(key, value);
        }
    }

    return `${url.pathname}${url.search}${url.hash}`;
};
