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

/**
 * Factory for Apollo-cache-driven `<title>` components used by route handles.
 *
 * Returns a named `ApolloTitle` component so `DocumentTitle` can detect it as
 * a component (PascalCase convention) versus a plain `(params) => string`
 * resolver.
 */
export function apolloTitle<TData, TVars extends Record<string, unknown>>(
    opts: ApolloTitleOpts<TData, TVars>,
): ComponentType<{ params: RouteParams }> {
    return function ApolloTitle({ params }: { params: RouteParams }) {
        const vars = opts.variables(params);
        const { data } = opts.useQuery(
            vars === null
                ? { fetchPolicy: 'cache-only', skip: true }
                : { fetchPolicy: 'cache-only', skip: false, variables: vars },
        );

        return renderTitle(opts.select(data, params));
    };
}
