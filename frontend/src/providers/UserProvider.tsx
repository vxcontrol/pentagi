import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

import { axios } from '@/lib/axios';
import { getReturnUrlParam } from '@/lib/utils/auth';
import { baseUrl } from '@/models/Api';
import type { AuthInfo, AuthInfoResponse } from '@/models/Info';

export interface LoginCredentials {
    mail: string;
    password: string;
}

export interface LoginResult {
    success: boolean;
    error?: string;
    passwordChangeRequired?: boolean;
}

export type OAuthProvider = 'google' | 'github';

interface UserContextType {
    authInfo: AuthInfo | null;
    isLoading: boolean;
    setAuth: (authInfo: AuthInfo) => void;
    clearAuth: () => void;
    isAuthenticated: () => boolean;
    logout: (returnUrl?: string) => Promise<void>;
    login: (credentials: LoginCredentials) => Promise<LoginResult>;
    loginWithOAuth: (provider: OAuthProvider) => Promise<LoginResult>;
}

const UserContext = createContext<UserContextType | undefined>(undefined);

export const AUTH_STORAGE_KEY = 'auth';

export const UserProvider = ({ children }: { children: ReactNode }) => {
    const navigate = useNavigate();
    const location = useLocation();
    const [authInfo, setAuthInfo] = useState<AuthInfo | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    // Load auth data from localStorage on mount, then load from API if needed
    useEffect(() => {
        const initializeAuth = async () => {
            try {
                const storedData = localStorage.getItem(AUTH_STORAGE_KEY);
                if (storedData) {
                    const parsedAuthInfo: AuthInfo = JSON.parse(storedData);
                    if (parsedAuthInfo) {
                        setAuthInfo(parsedAuthInfo);
                        setIsLoading(false);
                        return;
                    }
                }
            } catch (error) {
                // If parsing fails, clear invalid data
                localStorage.removeItem(AUTH_STORAGE_KEY);
            }

            // If no auth data in localStorage, load from API (for guest with providers list)
            try {
                const info: AuthInfoResponse = await axios.get('/info');
                if (info?.status === 'success' && info.data) {
                    // Set authInfo even for guest (contains providers list)
                    setAuthInfo(info.data);
                }
            } catch {
                // ignore
            } finally {
                setIsLoading(false);
            }
        };

        initializeAuth();
    }, []);

    const setAuth = useCallback((newAuthInfo: AuthInfo) => {
        setAuthInfo(newAuthInfo);
        // Persist to localStorage
        localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(newAuthInfo));
    }, []);

    const clearAuth = useCallback(() => {
        setAuthInfo(null);
        localStorage.removeItem(AUTH_STORAGE_KEY);
    }, []);

    const isAuthenticated = useCallback(() => {
        if (!authInfo?.user || !authInfo?.expires_at) {
            return false;
        }
        const now = new Date();
        const expirationDate = new Date(authInfo.expires_at);
        return expirationDate > now;
    }, [authInfo]);

    const logout = useCallback(async (returnUrl?: string) => {
        const currentPath = location.pathname;
        const finalReturnUrl = returnUrl || getReturnUrlParam(currentPath);

        try {
            await axios.get('/auth/logout');
        } finally {
            clearAuth();
            window.location.href = `/login${finalReturnUrl}`;
        }
    }, []);

    const login = useCallback(async (credentials: LoginCredentials): Promise<LoginResult> => {
        try {
            const loginResponse = await axios.post<unknown, { status: string; data?: unknown; error?: string }>(
                '/auth/login',
                credentials,
            );

            if (loginResponse?.status !== 'success') {
                return { success: false, error: 'Invalid login or password' };
            }

            // After login, backend set cookie, so we need to get fresh auth info
            const infoResponse: AuthInfoResponse = await axios.get('/info');

            if (infoResponse?.status !== 'success' || !infoResponse.data) {
                return { success: false, error: 'Failed to load user information' };
            }

            // Save auth info
            setAuth(infoResponse.data);

            // Check if password change is required for local users
            if (infoResponse.data.user?.type === 'local' && infoResponse.data.user.password_change_required) {
                return { success: true, passwordChangeRequired: true };
            }

            return { success: true };
        } catch {
            return { success: false, error: 'Login failed. Please try again.' };
        }
    }, []);

    const loginWithOAuth = useCallback(async (provider: OAuthProvider): Promise<LoginResult> => {
        const returnOAuthUri = '/oauth/result';
        const width = 500;
        const height = 600;
        const left = window.screenX + (window.outerWidth - width) / 2;
        const top = window.screenY + (window.outerHeight - height) / 2;

        const popup = window.open(
            `${baseUrl}/auth/authorize?provider=${provider}&return_uri=${returnOAuthUri}`,
            `${provider} Sign In`,
            `width=${width},height=${height},left=${left},top=${top}`,
        );

        if (!popup) {
            return {
                success: false,
                error: 'Popup blocked. Please allow popups for this site.',
            };
        }

        return new Promise<LoginResult>((resolve) => {
            const messageHandler = async (event: MessageEvent) => {
                if (event.origin !== window.location.origin || event.data?.type !== 'oauth-result') {
                    return;
                }

                window.removeEventListener('message', messageHandler);

                const cleanup = () => {
                    if (popup && !popup.closed) {
                        popup.close();
                    }
                };

                if (event.data.status === 'success') {
                    try {
                        const info: AuthInfoResponse = await axios.get('/info');

                        if (info?.status === 'success' && info.data?.type === 'user') {
                            setAuth(info.data);
                            cleanup();
                            resolve({ success: true });
                            return;
                        }
                    } catch (error) {
                        // In case of error, fall through to common handling below
                        console.error('Error during OAuth result handling:', error);
                    }
                }

                cleanup();
                resolve({
                    success: false,
                    error: event.data.error || 'Authentication failed',
                });
            };

            window.addEventListener('message', messageHandler);

            const popupCheckInterval = 500;
            const popupTimeout = 300000;

            const popupCheck = setInterval(() => {
                if (popup?.closed) {
                    clearInterval(popupCheck);
                    window.removeEventListener('message', messageHandler);
                    resolve({
                        success: false,
                        error: 'Authentication cancelled',
                    });
                }
            }, popupCheckInterval);

            setTimeout(() => {
                clearInterval(popupCheck);
                window.removeEventListener('message', messageHandler);

                if (popup && !popup.closed) {
                    popup.close();
                }

                resolve({
                    success: false,
                    error: 'Authentication timeout',
                });
            }, popupTimeout);
        });
    }, []);

    // Update auth state on route changes
    useEffect(() => {
        const updateAuth = async () => {
            // Skip for public routes
            const publicRoutes = ['/login', '/oauth/result'];

            // Check if user is authenticated
            if (!isAuthenticated()) {
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
                    const returnParam = getReturnUrlParam(location.pathname);
                    navigate(`/login${returnParam}`);
                }
            } catch {
                clearAuth();
                const returnParam = getReturnUrlParam(location.pathname);
                navigate(`/login${returnParam}`);
            }
        };

        updateAuth();
    }, [location.pathname, authInfo]);

    return (
        <UserContext.Provider
            value={{
                authInfo,
                isLoading,
                setAuth,
                clearAuth,
                isAuthenticated,
                logout,
                login,
                loginWithOAuth,
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
