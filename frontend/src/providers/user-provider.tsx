import type { ReactNode } from 'react';

import { createContext, use, useCallback, useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

import type { AuthInfo, AuthInfoResponse } from '@/models/info';

import { axios } from '@/lib/axios';
import { getReturnUrlParam } from '@/lib/utils/auth';
import { baseUrl } from '@/models/api';

export interface LoginCredentials {
    mail: string;
    password: string;
}

export interface LoginResult {
    error?: string;
    passwordChangeRequired?: boolean;
    success: boolean;
}

export type OAuthProvider = 'github' | 'google';

interface UserContextType {
    authInfo: AuthInfo | null;
    clearAuth: () => void;
    isAuthenticated: () => boolean;
    isLoading: boolean;
    login: (credentials: LoginCredentials) => Promise<LoginResult>;
    loginWithOAuth: (provider: OAuthProvider) => Promise<LoginResult>;
    logout: (returnUrl?: string) => Promise<void>;
    setAuth: (authInfo: AuthInfo) => void;
}

const UserContext = createContext<undefined | UserContextType>(undefined);

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
            } catch {
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

    const logout = useCallback(
        async (returnUrl?: string) => {
            const currentPath = location.pathname;
            const finalReturnUrl = returnUrl || getReturnUrlParam(currentPath);

            try {
                await axios.get('/auth/logout');
                toast.success('Successfully logged out');
            } catch {
                toast.error('Logout failed, but clearing local session');
            } finally {
                clearAuth();
                window.location.href = `/login${finalReturnUrl}`;
            }
        },
        [clearAuth, location.pathname],
    );

    const login = useCallback(
        async (credentials: LoginCredentials): Promise<LoginResult> => {
            try {
                const loginResponse = await axios.post<unknown, { data?: unknown; error?: string; status: string }>(
                    '/auth/login',
                    credentials,
                );

                if (loginResponse?.status !== 'success') {
                    const errorMessage = 'Invalid login or password';
                    toast.error(errorMessage);

                    return { error: errorMessage, success: false };
                }

                // After login, backend set cookie, so we need to get fresh auth info
                const infoResponse: AuthInfoResponse = await axios.get('/info');

                if (infoResponse?.status !== 'success' || !infoResponse.data) {
                    const errorMessage = 'Failed to load user information';
                    toast.error(errorMessage);

                    return { error: errorMessage, success: false };
                }

                // Save auth info
                setAuth(infoResponse.data);

                // Check if password change is required for local users
                if (infoResponse.data.user?.type === 'local' && infoResponse.data.user.password_change_required) {
                    toast.warning('Password change required');

                    return { passwordChangeRequired: true, success: true };
                }

                // toast.success('Successfully logged in');
                return { success: true };
            } catch {
                const errorMessage = 'Login failed. Please try again.';
                toast.error(errorMessage);

                return { error: errorMessage, success: false };
            }
        },
        [setAuth],
    );

    const loginWithOAuth = useCallback(
        async (provider: OAuthProvider): Promise<LoginResult> => {
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
                const errorMessage = 'Popup blocked. Please allow popups for this site.';
                toast.error(errorMessage);

                return {
                    error: errorMessage,
                    success: false,
                };
            }

            return new Promise<LoginResult>((resolve) => {
                const popupCheckInterval = 500;
                const popupTimeout = 300000;
                let isResolved = false;

                const popupCheck = setInterval(() => {
                    if (popup?.closed && !isResolved) {
                        isResolved = true;
                        clearInterval(popupCheck);
                        clearTimeout(timeoutId);
                        window.removeEventListener('message', messageHandler);
                        const errorMessage = 'Authentication cancelled';
                        toast.info(errorMessage);
                        resolve({
                            error: errorMessage,
                            success: false,
                        });
                    }
                }, popupCheckInterval);

                const timeoutId = setTimeout(() => {
                    if (!isResolved) {
                        isResolved = true;
                        clearInterval(popupCheck);
                        window.removeEventListener('message', messageHandler);

                        if (popup && !popup.closed) {
                            popup.close();
                        }

                        const errorMessage = 'Authentication timeout';
                        toast.error(errorMessage);
                        resolve({
                            error: errorMessage,
                            success: false,
                        });
                    }
                }, popupTimeout);

                const messageHandler = async (event: MessageEvent) => {
                    if (event.origin !== window.location.origin || event.data?.type !== 'oauth-result') {
                        return;
                    }

                    if (isResolved) {
                        return;
                    }

                    isResolved = true;
                    clearInterval(popupCheck);
                    clearTimeout(timeoutId);
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
                                // toast.success('Successfully logged in');
                                resolve({ success: true });

                                return;
                            }
                        } catch (error) {
                            // In case of error, fall through to common handling below
                            console.error('Error during OAuth result handling:', error);
                        }
                    }

                    cleanup();
                    const errorMessage = event.data.error || 'Authentication failed';
                    toast.error(errorMessage);
                    resolve({
                        error: errorMessage,
                        success: false,
                    });
                };

                window.addEventListener('message', messageHandler);
            });
        },
        [setAuth],
    );

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
                    toast.error('Session expired. Please login again.');
                    const returnParam = getReturnUrlParam(location.pathname);
                    navigate(`/login${returnParam}`);
                }
            } catch {
                clearAuth();
                toast.error('Session expired. Please login again.');
                const returnParam = getReturnUrlParam(location.pathname);
                navigate(`/login${returnParam}`);
            }
        };

        updateAuth();
    }, [location.pathname]);

    return (
        <UserContext
            value={{
                authInfo,
                clearAuth,
                isAuthenticated,
                isLoading,
                login,
                loginWithOAuth,
                logout,
                setAuth,
            }}
        >
            {children}
        </UserContext>
    );
};

export const useUser = () => {
    const context = use(UserContext);

    if (context === undefined) {
        throw new Error('useUser must be used within a UserProvider');
    }

    return context;
};
