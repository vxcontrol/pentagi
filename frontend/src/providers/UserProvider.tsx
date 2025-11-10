import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

import { axios } from '@/lib/axios';
import type { AuthInfo, AuthInfoResponse } from '@/models/Info';
import type { User } from '@/models/User';

interface UserContextType {
    user: User | null;
    authInfo: AuthInfo | null;
    expiresAt: string | null;
    setAuth: (authInfo: AuthInfo) => void;
    clearAuth: () => void;
    isAuthenticated: () => boolean;
    refreshAuth: () => Promise<{ success: boolean; authInfo?: AuthInfo }>;
}

const UserContext = createContext<UserContextType | undefined>(undefined);

const AUTH_STORAGE_KEY = 'auth';

export const UserProvider = ({ children }: { children: ReactNode }) => {
    const navigate = useNavigate();
    const location = useLocation();
    const [user, setUser] = useState<User | null>(null);
    const [authInfo, setAuthInfo] = useState<AuthInfo | null>(null);
    const [expiresAt, setExpiresAt] = useState<string | null>(null);

    // Load auth data from localStorage on mount
    useEffect(() => {
        try {
            const storedData = localStorage.getItem(AUTH_STORAGE_KEY);
            if (storedData) {
                const parsedAuthInfo: AuthInfo = JSON.parse(storedData);
                if (parsedAuthInfo) {
                    setAuthInfo(parsedAuthInfo);
                    setUser(parsedAuthInfo.user || null);
                    setExpiresAt(parsedAuthInfo.expires_at || null);
                }
            }
        } catch (error) {
            // If parsing fails, clear invalid data
            localStorage.removeItem(AUTH_STORAGE_KEY);
        }
    }, []);

    const setAuth = useCallback((newAuthInfo: AuthInfo) => {
        const expiresAtValue = newAuthInfo.expires_at || null;

        setAuthInfo(newAuthInfo);
        setUser(newAuthInfo.user || null);
        setExpiresAt(expiresAtValue);

        // Persist to localStorage
        localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(newAuthInfo));
    }, []);

    const clearAuth = useCallback(() => {
        setAuthInfo(null);
        setUser(null);
        setExpiresAt(null);
        localStorage.removeItem(AUTH_STORAGE_KEY);
    }, []);

    const isAuthenticated = useCallback(() => {
        if (!user || !expiresAt) {
            return false;
        }
        const now = new Date();
        const expirationDate = new Date(expiresAt);
        return expirationDate > now;
    }, [user, expiresAt]);

    const refreshAuth = useCallback(async (): Promise<{ success: boolean; authInfo?: AuthInfo }> => {
        try {
            const info: AuthInfoResponse = await axios.get('/info', {
                params: {
                    refresh_cookie: true,
                },
            });

            if (info?.status === 'success' && info.data) {
                setAuth(info.data);
                return { success: true, authInfo: info.data };
            }

            clearAuth();
            return { success: false };
        } catch {
            clearAuth();
            return { success: false };
        }
    }, [setAuth, clearAuth]);

    // Sync auth state on route changes
    useEffect(() => {
        const syncAuth = async () => {
            // Skip for public routes
            const publicRoutes = ['/login', '/oauth/result'];

            // Check authentication status directly using state
            if (!user || !expiresAt) {
                return;
            }

            const now = new Date();
            const expirationDate = new Date(expiresAt);
            if (expirationDate <= now) {
                return;
            }

            if (publicRoutes.includes(location.pathname)) {
                return;
            }

            try {
                const info: AuthInfoResponse = await axios.get('/info', {
                    params: {
                        refresh_cookie: true,
                    },
                });

                if (info?.status === 'success' && info.data) {
                    setAuth(info.data);
                } else {
                    clearAuth();
                    // Save current path for redirect after login
                    const currentPath = location.pathname;
                    // Only save if it's not the default route
                    const returnParam =
                        currentPath !== '/flows/new' ? `?returnUrl=${encodeURIComponent(currentPath)}` : '';
                    navigate(`/login${returnParam}`);
                }
            } catch {
                clearAuth();
                // Save current path for redirect after login
                const currentPath = location.pathname;
                // Only save if it's not the default route
                const returnParam = currentPath !== '/flows/new' ? `?returnUrl=${encodeURIComponent(currentPath)}` : '';
                navigate(`/login${returnParam}`);
            }
        };

        syncAuth();
    }, [location.pathname]);

    return (
        <UserContext.Provider
            value={{
                user,
                authInfo,
                expiresAt,
                setAuth,
                clearAuth,
                isAuthenticated,
                refreshAuth,
            }}
        >
            {children}
        </UserContext.Provider>
    );
};

export const useUser = () => {
    const context = useContext(UserContext);
    if (context === undefined) {
        throw new Error('useUser must be used within a UserProvider');
    }
    return context;
};
