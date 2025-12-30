import { createContext, type ReactNode, useCallback, useContext, useEffect, useMemo, useState } from 'react';

import { Log } from '@/lib/log';
import { useUser } from '@/providers/user-provider';

interface FavoriteFlow {
    id: string;
    name: string;
}

interface FavoritesContextValue {
    addFavoriteFlow: (flowId: string, flowName: string) => void;
    favoriteFlows: FavoriteFlow[];
    isFavoriteFlow: (flowId: string) => boolean;
    removeFavoriteFlow: (flowId: string) => void;
    toggleFavoriteFlow: (flowId: string, flowName: string) => void;
}

interface FavoritesProviderProps {
    children: ReactNode;
}

interface FavoritesStorage {
    [userId: string]: UserFavorites;
}

interface UserFavorites {
    flows?: FavoriteFlow[];
}

const FavoritesContext = createContext<FavoritesContextValue | undefined>(undefined);

const FAVORITES_STORAGE_KEY = 'favorites';

const loadFavorites = (): FavoritesStorage => {
    try {
        const stored = localStorage.getItem(FAVORITES_STORAGE_KEY);

        if (stored) {
            const parsed = JSON.parse(stored);

            return typeof parsed === 'object' && parsed !== null ? parsed : {};
        }
    } catch (error) {
        Log.error('Error loading favorites from storage:', error);
    }

    return {};
};

const saveFavorites = (favorites: FavoritesStorage): void => {
    try {
        localStorage.setItem(FAVORITES_STORAGE_KEY, JSON.stringify(favorites));
    } catch (error) {
        Log.error('Error saving favorites to storage:', error);
    }
};

export const FavoritesProvider = ({ children }: FavoritesProviderProps) => {
    const { authInfo } = useUser();
    const userId = authInfo?.user?.id?.toString() ?? 'guest';

    const [favoritesStorage, setFavoritesStorage] = useState<FavoritesStorage>(() => loadFavorites());

    // Get current user's favorite flows sorted by id (descending)
    const favoriteFlows = useMemo(() => {
        const flows = favoritesStorage[userId]?.flows ?? [];

        return [...flows].sort((a, b) => +b.id - +a.id);
    }, [favoritesStorage, userId]);

    // Save to localStorage whenever favorites change
    useEffect(() => {
        saveFavorites(favoritesStorage);
    }, [favoritesStorage]);

    const addFavoriteFlow = useCallback(
        (flowId: string, flowName: string) => {
            setFavoritesStorage((previousStorage) => {
                const userFavorites = previousStorage[userId] ?? {};
                const currentFlows = userFavorites.flows ?? [];

                // Check if already exists
                if (currentFlows.some((flow) => flow.id === flowId)) {
                    return previousStorage;
                }

                return {
                    ...previousStorage,
                    [userId]: {
                        ...userFavorites,
                        flows: [...currentFlows, { id: flowId, name: flowName }],
                    },
                };
            });
        },
        [userId],
    );

    const removeFavoriteFlow = useCallback(
        (flowId: string) => {
            setFavoritesStorage((previousStorage) => {
                const userFavorites = previousStorage[userId];

                if (!userFavorites?.flows) {
                    return previousStorage;
                }

                const updatedFlows = userFavorites.flows.filter((flow) => flow.id !== flowId);

                return {
                    ...previousStorage,
                    [userId]: {
                        ...userFavorites,
                        flows: updatedFlows,
                    },
                };
            });
        },
        [userId],
    );

    const toggleFavoriteFlow = useCallback(
        (flowId: string, flowName: string) => {
            const isFavorite = favoriteFlows.some((flow) => flow.id === flowId);

            if (isFavorite) {
                removeFavoriteFlow(flowId);
            } else {
                addFavoriteFlow(flowId, flowName);
            }
        },
        [favoriteFlows, addFavoriteFlow, removeFavoriteFlow],
    );

    const isFavoriteFlow = useCallback(
        (flowId: string) => {
            return favoriteFlows.some((flow) => flow.id === flowId);
        },
        [favoriteFlows],
    );

    const value = useMemo(
        () => ({
            addFavoriteFlow,
            favoriteFlows,
            isFavoriteFlow,
            removeFavoriteFlow,
            toggleFavoriteFlow,
        }),
        [addFavoriteFlow, favoriteFlows, isFavoriteFlow, removeFavoriteFlow, toggleFavoriteFlow],
    );

    return <FavoritesContext.Provider value={value}>{children}</FavoritesContext.Provider>;
};

export const useFavorites = () => {
    const context = useContext(FavoritesContext);

    if (context === undefined) {
        throw new Error('useFavorites must be used within FavoritesProvider');
    }

    return context;
};
