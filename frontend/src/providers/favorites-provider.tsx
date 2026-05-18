import { createContext, type ReactNode, startTransition, useCallback, useContext, useEffect, useMemo, useOptimistic } from 'react';
import { toast } from 'sonner';

import {
    useAddFavoriteFlowMutation,
    useDeleteFavoriteFlowMutation,
    useSettingsUserQuery,
    useSettingsUserUpdatedSubscription,
} from '@/graphql/types';
import { Log } from '@/lib/log';
import { useUser } from '@/providers/user-provider';

interface FavoritesContextValue {
    addFavoriteFlow: (flowId: number | string) => Promise<void>;
    favoriteFlowIds: number[];
    isFavoriteFlow: (flowId: number | string) => boolean;
    isLoading: boolean;
    removeFavoriteFlow: (flowId: number | string) => Promise<void>;
    toggleFavoriteFlow: (flowId: number | string) => Promise<void>;
}

interface FavoritesProviderProps {
    children: ReactNode;
}

const FavoritesContext = createContext<FavoritesContextValue | undefined>(undefined);

const FAVORITES_STORAGE_KEY = 'favorites';

export function FavoritesProvider({ children }: FavoritesProviderProps) {
    const { authInfo, isAuthenticated } = useUser();

    // Only fetch user preferences if user is authenticated and not a guest
    // authInfo must exist and type must be 'user' or 'api' (not 'guest' and not null/undefined)
    // Also check isAuthenticated() to ensure session is valid
    const shouldFetchPreferences = Boolean(authInfo && authInfo.type !== 'guest' && isAuthenticated());

    // GraphQL query for user preferences
    const { data: userPreferencesData, loading: isLoadingPreferences } = useSettingsUserQuery({
        fetchPolicy: 'cache-and-network',
        skip: !shouldFetchPreferences,
    });

    // GraphQL mutations
    const [addFavoriteFlowMutation] = useAddFavoriteFlowMutation();
    const [deleteFavoriteFlowMutation] = useDeleteFavoriteFlowMutation();

    // GraphQL subscription (only for authenticated users)
    useSettingsUserUpdatedSubscription({
        skip: !shouldFetchPreferences,
    });

    // Get favorite flow IDs from GraphQL as numbers
    const actualFavoriteFlowIds = useMemo(() => {
        const ids = userPreferencesData?.settingsUser?.favoriteFlows ?? [];

        return ids.map((id) => +id);
    }, [userPreferencesData?.settingsUser?.favoriteFlows]);

    // Surface the optimistic set so star-clicks flip in the UI before the
    // mutation + subscription round-trip lands. React rolls back to
    // `actualFavoriteFlowIds` automatically if the transition's action throws.
    const [favoriteFlowIds, applyOptimisticFavorite] = useOptimistic(
        actualFavoriteFlowIds,
        (current: number[], action: { id: number; type: 'add' | 'remove' }) => {
            if (action.type === 'add') {
                return current.includes(action.id) ? current : [...current, action.id];
            }

            return current.filter((id) => id !== action.id);
        },
    );

    // Migration: sync localStorage favorites to backend on first load
    useEffect(() => {
        const migrateLocalStorageFavorites = async () => {
            try {
                const stored = localStorage.getItem(FAVORITES_STORAGE_KEY);

                if (!stored) {
                    return;
                }

                const parsed = JSON.parse(stored);

                if (typeof parsed !== 'object' || !parsed) {
                    return;
                }

                // Get current user's favorites from localStorage
                const userIds = Object.keys(parsed);

                if (userIds.length === 0) {
                    localStorage.removeItem(FAVORITES_STORAGE_KEY);

                    return;
                }

                const userId = userIds[0];

                if (!userId) {
                    localStorage.removeItem(FAVORITES_STORAGE_KEY);

                    return;
                }

                const localFavorites = parsed[userId]?.flows ?? [];

                if (localFavorites.length === 0) {
                    // No local favorites to migrate
                    localStorage.removeItem(FAVORITES_STORAGE_KEY);

                    return;
                }

                // Migrate each favorite to backend
                for (const flow of localFavorites) {
                    // Check if already in backend
                    if (!favoriteFlowIds.includes(flow.id)) {
                        try {
                            await addFavoriteFlowMutation({
                                variables: { flowId: flow.id },
                            });
                        } catch (error) {
                            Log.error('Error migrating favorite flow:', error);
                        }
                    }
                }

                // Clear localStorage after successful migration
                localStorage.removeItem(FAVORITES_STORAGE_KEY);
                Log.info('Successfully migrated favorites from localStorage to backend');
            } catch (error) {
                Log.error('Error during favorites migration:', error);
            }
        };

        // Only run migration if we have loaded preferences and localStorage data exists
        // and user is authenticated (not a guest)
        if (!isLoadingPreferences && userPreferencesData && shouldFetchPreferences) {
            migrateLocalStorageFavorites();
        }
    }, [isLoadingPreferences, userPreferencesData, favoriteFlowIds, addFavoriteFlowMutation, shouldFetchPreferences]);

    const addFavoriteFlow = useCallback(
        async (flowId: number | string) => {
            const id = typeof flowId === 'string' ? flowId : flowId.toString();
            const numericId = typeof flowId === 'string' ? +flowId : flowId;

            startTransition(async () => {
                applyOptimisticFavorite({ id: numericId, type: 'add' });

                try {
                    await addFavoriteFlowMutation({
                        variables: { flowId: id },
                    });
                } catch (error) {
                    const errorMessage = error instanceof Error ? error.message : 'Failed to add favorite';
                    toast.error('Failed to add to favorites', {
                        description: errorMessage,
                    });
                    Log.error('Error adding favorite flow:', error);
                }
            });
        },
        [addFavoriteFlowMutation, applyOptimisticFavorite],
    );

    const removeFavoriteFlow = useCallback(
        async (flowId: number | string) => {
            const id = typeof flowId === 'string' ? flowId : flowId.toString();
            const numericId = typeof flowId === 'string' ? +flowId : flowId;

            startTransition(async () => {
                applyOptimisticFavorite({ id: numericId, type: 'remove' });

                try {
                    await deleteFavoriteFlowMutation({
                        variables: { flowId: id },
                    });
                } catch (error) {
                    const errorMessage = error instanceof Error ? error.message : 'Failed to remove favorite';
                    toast.error('Failed to remove from favorites', {
                        description: errorMessage,
                    });
                    Log.error('Error removing favorite flow:', error);
                }
            });
        },
        [applyOptimisticFavorite, deleteFavoriteFlowMutation],
    );

    const toggleFavoriteFlow = useCallback(
        async (flowId: number | string) => {
            const numId = typeof flowId === 'string' ? +flowId : flowId;
            const isFavorite = favoriteFlowIds.includes(numId);

            if (isFavorite) {
                await removeFavoriteFlow(flowId);
            } else {
                await addFavoriteFlow(flowId);
            }
        },
        [favoriteFlowIds, addFavoriteFlow, removeFavoriteFlow],
    );

    const isFavoriteFlow = useCallback(
        (flowId: number | string) => {
            const numId = typeof flowId === 'string' ? +flowId : flowId;

            return favoriteFlowIds.includes(numId);
        },
        [favoriteFlowIds],
    );

    const value = useMemo(
        () => ({
            addFavoriteFlow,
            favoriteFlowIds,
            isFavoriteFlow,
            isLoading: isLoadingPreferences,
            removeFavoriteFlow,
            toggleFavoriteFlow,
        }),
        [
            addFavoriteFlow,
            favoriteFlowIds,
            isFavoriteFlow,
            isLoadingPreferences,
            removeFavoriteFlow,
            toggleFavoriteFlow,
        ],
    );

    return <FavoritesContext.Provider value={value}>{children}</FavoritesContext.Provider>;
}

export function useFavorites() {
    const context = useContext(FavoritesContext);

    if (context === undefined) {
        throw new Error('useFavorites must be used within FavoritesProvider');
    }

    return context;
}
