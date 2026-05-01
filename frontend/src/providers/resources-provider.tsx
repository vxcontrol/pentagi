import { useApolloClient } from '@apollo/client';
import { createContext, type ReactNode, useContext, useEffect, useMemo, useState } from 'react';
import { toast } from 'sonner';

import type { UserResourceFragmentFragment } from '@/graphql/types';

import { RESOURCES_API_PATH } from '@/features/resources/resources-constants';
import { restResourceEntryToFragment, type RestResourceList } from '@/features/resources/resources-rest';
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

interface ResourcesProviderProps {
    children: ReactNode;
}

const ResourcesContext = createContext<ResourcesContextValue | undefined>(undefined);

const RESOURCES_ERROR_TOAST_ID = 'resources-error';

/**
 * Loads the user's full resource library and keeps it in sync via three GraphQL
 * subscriptions (added / updated / deleted).
 *
 * The initial library is fetched via REST `GET /resources/?recursive=true` and
 * the result is written into the same Apollo cache slot that `useResourcesQuery`
 * reads (via `cache-only` fetch policy). Subscriptions continue to update the
 * cache through `lib/apollo.ts`, so consumers see one unified store.
 *
 * NOTE: the equivalent GraphQL `resources(recursive: true)` query now works
 * correctly for the root path. Switching this provider to a plain
 * `useResourcesQuery({ variables: { recursive: true } })` is a viable follow-up
 * once we no longer need the snake_case → camelCase conversion that REST
 * requires (see `restResourceEntryToFragment`).
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
                const response = await api.get<RestResourceList>(RESOURCES_API_PATH, {
                    params: { recursive: true },
                });
                const data = unwrapApiResponse(response);

                if (isCancelled) {
                    return;
                }

                // The REST endpoint returns `models.ResourceEntry` with snake_case
                // JSON tags and numeric `id`/`user_id`. We convert each entry into the
                // camelCase GraphQL fragment shape so it can sit in the same Apollo
                // cache slot that subscriptions and the GraphQL `resources` query write
                // to. `__typename` is added explicitly because Apollo refuses to
                // normalise an object as `UserResource:<id>` without it.
                apolloClient.writeQuery({
                    data: {
                        resources: (data.items ?? []).map((item) => ({
                            ...restResourceEntryToFragment(item),
                            __typename: 'UserResource' as const,
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
            // `item.id` is canonically numeric (see `resources-rest.ts`) but typed
            // as `string` by codegen — coerce both sides for the lookup.
            getResource: (id: string) => resources.find((item) => String(item.id) === id),
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
