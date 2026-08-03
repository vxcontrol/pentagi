/**
 * Normalize a string for case- and diacritic-insensitive substring matching.
 *
 * `NFKD` decomposes accented characters into base + combining marks, then we
 * strip the combining marks (`\p{Diacritic}` regex class) so e.g. `café`
 * matches `cafe`. Lowercasing happens last to fold case differences.
 *
 * Exported for callers that need to align their own search semantics with the
 * one this module uses (e.g. server-side prefiltering).
 */
export const normalizeForFilter = (text: string): string =>
    text
        .normalize('NFKD')
        .replace(/\p{Diacritic}/gu, '')
        .toLowerCase();

/**
 * Build a reusable text-matcher specialised for `query`. The query is
 * normalized + lowercased once at factory time — when the resulting matcher
 * is invoked N times during list filtering, that work is amortised to one
 * allocation instead of N.
 *
 * The matcher uses substring matching semantics, with case + diacritic
 * folding (`café` matches `cafe`). It backs both the DataTable global
 * filter and detail-navigation's Prev/Next subset — the same `?q=` drives
 * both, so a semantics change here shifts what matches on every list page
 * and its detail pager at once.
 *
 * - Empty query → every row passes (matcher returns `true`).
 * - `text === null | undefined` with a non-empty query → no match.
 */
export const createTextMatcher = (query: string): ((text: null | string | undefined) => boolean) => {
    if (!query.length) {
        return () => true;
    }

    const normalizedQuery = normalizeForFilter(query);

    return (text) => {
        if (text === null || text === undefined) {
            return false;
        }

        return normalizeForFilter(text).includes(normalizedQuery);
    };
};

/**
 * One-shot variant of {@link createTextMatcher} for callers that need a
 * single comparison and don't want to bother with the factory. Equivalent to
 * `createTextMatcher(query)(text)` — kept as a thin wrapper for readability
 * at call sites and to preserve the original API.
 */
export const matchesTextFilter = (text: null | string | undefined, query: string): boolean =>
    createTextMatcher(query)(text);
