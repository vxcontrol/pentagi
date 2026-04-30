import { useApolloClient } from '@apollo/client';
import { createContext, type ReactNode, useContext, useEffect, useMemo, useState } from 'react';
import { toast } from 'sonner';

import type { UserResourceFragmentFragment } from '@/graphql/types';

import { RESOURCES_API_PATH } from '@/features/resources/resources-constants';
import { useResourcesRealtime } from '@/features/resources/use-resources-realtime';
import { ResourcesDocument, useResourcesQuery } from '@/graphql/types';
import { api, getApiErrorMessage, unwrapApiResponse } from '@/lib/axios';
import { useUser } from '@/providers/user-provider';

interface ResourcesContextValue {
    /** Recursive list of every entry in the user's library (files + directories). */
    error: Error | null | undefined;
    /** Lookup helper: returns `undefined` when the resource is unknown. */
    getResource: (id: string) => undefined | UserResourceFragmentFragment;
    isInitialLoading: boolean;
    isLoading: boolean;
    /** Force a network re-read of the resources list. */
    refetch: () => Promise<unknown>;
    resources: UserResourceFragmentFragment[];
}

/**
 * Shape of the REST `GET /resources/` response wrapper.
 * The element type is intentionally aliased to the GraphQL fragment — the backend's
 * `models.ResourceEntry` is meant to be structurally identical, and we treat them as
 * the same type so REST-loaded entries can sit in the same Apollo cache as GraphQL ones.
 */
interface ResourcesListResponse {
    items?: null | UserResourceFragmentFragment[];
    total?: number;
}

interface ResourcesProviderProps {
    children: ReactNode;
}

const ResourcesContext = createContext<ResourcesContextValue | undefined>(undefined);

const RESOURCES_ERROR_TOAST_ID = 'resources-error';

/**
 * Loads the user's full resource library and keeps it in sync via three GraphQL
 * subscriptions (added / updated / deleted).
 *
 * **Temporary detail**: the initial library is fetched via REST
 * `GET /resources/?recursive=true` rather than the GraphQL `resources` query.
 * The GraphQL resolver currently has a bug for `path == "" && recursive == true`
 * that returns only top-level entries — we side-step it by hydrating the same
 * Apollo cache slot from REST and reading it back through `useResourcesQuery`
 * with a `cache-only` fetch policy. Subscriptions continue to update the cache
 * as before, so consumers see no behavioural change.
 *
 * Once the backend resolver is fixed this hook should switch back to a plain
 * `useResourcesQuery({ variables: { recursive: true } })`.
 */
export const ResourcesProvider = ({ children }: ResourcesProviderProps) => {
    const { authInfo, isAuthenticated } = useUser();

    const shouldFetchResources = Boolean(authInfo && authInfo.type !== 'guest' && isAuthenticated());

    const apolloClient = useApolloClient();

    const [restLoading, setRestLoading] = useState(false);
    const [restError, setRestError] = useState<Error | null>(null);
    const [refreshTick, setRefreshTick] = useState(0);

    // Hydrate the Apollo cache from REST. We write to the same `resources(recursive: true)`
    // cache slot that `useResourcesQuery` reads, so subscription delta updates plumbed
    // through `lib/apollo.ts` keep working without any extra wiring.
    useEffect(() => {
        if (!shouldFetchResources) {
            return;
        }

        let isCancelled = false;

        setRestLoading(true);
        setRestError(null);

        (async () => {
            try {
                const response = await api.get<ResourcesListResponse>(RESOURCES_API_PATH, {
                    params: { recursive: true },
                });
                const data = unwrapApiResponse(response);

                if (isCancelled) {
                    return;
                }

                // `__typename` is required so Apollo normalizes each entry under
                // `UserResource:<id>` in the cache. Without it the read-back through
                // `useResourcesQuery` returns objects whose scalar fields are `undefined`
                // (the cache treats every entry as an unidentifiable structural blob).
                //
                // TODO(backend): remove `String(item.id)` / `String(item.userId)` coercion
                // once the REST `/resources/` endpoint returns `id`/`userId` as strings.
                // Currently it returns them as numbers, which breaks consumers that rely
                // on the GraphQL `ID` scalar (typed as `string` everywhere) — e.g. zod
                // validation `z.array(z.string())` in `FlowForm.resourceIds`.
                apolloClient.writeQuery({
                    data: {
                        resources: (data.items ?? []).map((item) => ({
                            ...item,
                            __typename: 'UserResource' as const,
                            id: String(item.id),
                            userId: String(item.userId),
                        })),
                    },
                    query: ResourcesDocument,
                    variables: { recursive: true },
                });
            } catch (caught) {
                if (isCancelled) {
                    return;
                }

                const message = getApiErrorMessage(caught, 'Failed to load resources');

                setRestError(new Error(message));
            } finally {
                if (!isCancelled) {
                    setRestLoading(false);
                }
            }
        })();

        return () => {
            isCancelled = true;
        };
    }, [apolloClient, refreshTick, shouldFetchResources]);

    const { data, error: graphqlError } = useResourcesQuery({
        fetchPolicy: 'cache-only',
        skip: !shouldFetchResources,
        variables: { recursive: true },
    });

    const error = restError ?? graphqlError ?? null;
    const isInitialLoading = restLoading && data === undefined;

    useResourcesRealtime({ isPaused: !shouldFetchResources || isInitialLoading });

    useEffect(() => {
        if (error) {
            toast.error('Failed to load resources', {
                description: error.message,
                id: RESOURCES_ERROR_TOAST_ID,
            });
        }
    }, [error]);

    const resources = useMemo<UserResourceFragmentFragment[]>(() => data?.resources ?? [], [data?.resources]);

    const value = useMemo<ResourcesContextValue>(
        () => ({
            error,
            getResource: (id: string) => resources.find((item) => item.id === id),
            isInitialLoading,
            isLoading: restLoading,
            refetch: () => {
                setRefreshTick((tick) => tick + 1);

                return Promise.resolve();
            },
            resources,
        }),
        [error, isInitialLoading, resources, restLoading],
    );

    return <ResourcesContext.Provider value={value}>{children}</ResourcesContext.Provider>;
};

export const useResources = () => {
    const context = useContext(ResourcesContext);

    if (context === undefined) {
        throw new Error('useResources must be used within ResourcesProvider');
    }

    return context;
};
