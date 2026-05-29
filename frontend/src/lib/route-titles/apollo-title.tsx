import type { QueryHookOptions, QueryResult } from '@apollo/client/react';
import type { ComponentType } from 'react';

import { renderTitle, type RouteParams } from './render-title';

// Apollo codegen hook signature — `useXxxQuery({ variables, skip? })` or
// `useXxxQuery({ skip: true })`. We narrow to `cache-only` so the title
// component never issues an HTTP request — the destination page's own query
// fills the cache.
type ApolloQueryHook<TData, TVars extends Record<string, unknown>> = (
    options: QueryHookOptions<TData, TVars> &
        ({ skip: boolean; variables?: TVars } | { skip?: boolean; variables: TVars }),
) => QueryResult<TData, TVars>;

interface ApolloTitleOpts<TData, TVars extends Record<string, unknown>> {
    /**
     * Compute the label from the cached data and route params. Receives
     * `undefined` when the query is skipped or the cache is empty.
     */
    select: (data: null | TData | undefined, params: RouteParams) => string;
    /**
     * Apollo codegen hook for the target query.
     */
    useQuery: ApolloQueryHook<TData, TVars>;
    /**
     * Return the variables for the query, or `null` to skip the query
     * entirely (used for `:id === 'new'` routes and missing params).
     */
    variables: (params: RouteParams) => null | TVars;
}

// Terser drops the `function ApolloTitle()` name in production builds, so
// `DocumentTitle` cannot rely on `fn.name` to distinguish a reactive title
// component from a plain `(params) => string` resolver. This explicit marker
// survives minification because Vite preserves property names by default,
// and Terser leaves Symbol-keyed properties alone even under aggressive
// `mangle.properties`. `Symbol.for` is used so an accidental double-bundle
// (worker, SSR boundary) still resolves to the same key.
const APOLLO_TITLE_MARKER = Symbol.for('pentagi.apolloTitle');

export type ApolloTitleComponent = ComponentType<{ params: RouteParams }> & {
    readonly [APOLLO_TITLE_MARKER]: true;
};

/**
 * Type guard for components produced by {@link apolloTitle}. `DocumentTitle`
 * uses this to decide whether to render `<TitleComp params={...} />` or call
 * the value as a `(params) => string` resolver. Keep the marker private to
 * this module — callers should only depend on this guard.
 */
export const isApolloTitle = (value: unknown): value is ApolloTitleComponent =>
    typeof value === 'function' && (value as { [APOLLO_TITLE_MARKER]?: unknown })[APOLLO_TITLE_MARKER] === true;

/**
 * Factory for Apollo-cache-driven `<title>` components used by route handles.
 *
 * Tags the returned component with an internal marker so `DocumentTitle`
 * (via {@link isApolloTitle}) can identify it as a component and render via
 * JSX rather than treating it as a `(params) => string` resolver.
 */
export function apolloTitle<TData, TVars extends Record<string, unknown>>(
    opts: ApolloTitleOpts<TData, TVars>,
): ApolloTitleComponent {
    function ApolloTitle({ params }: { params: RouteParams }) {
        const vars = opts.variables(params);
        const { data } = opts.useQuery(
            vars === null
                ? { fetchPolicy: 'cache-only', skip: true }
                : { fetchPolicy: 'cache-only', skip: false, variables: vars },
        );

        return renderTitle(opts.select(data, params));
    }

    return Object.assign(ApolloTitle, { [APOLLO_TITLE_MARKER]: true as const });
}
