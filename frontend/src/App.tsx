import { ApolloProvider } from '@apollo/client';
import { Suspense } from 'react';
import type { Location, NavigateFunction } from 'react-router-dom';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';

import AppLayout from '@/components/AppLayout';
import PageLoader from '@/components/PageLoader';
import ProtectedRoute from '@/components/ProtectedRoute';
import PublicRoute from '@/components/PublicRoute';
import RouteChangeTracker from '@/components/RouteChangeTracker';
import client from '@/lib/apollo';
import { isAuthenticated } from '@/lib/auth';
import { axios } from '@/lib/axios';
import Chat from '@/pages/Chat';
import Login from '@/pages/Login';
import OAuthResult from '@/pages/OAuthResult';
import Report from '@/pages/Report';
import ThemeProvider from '@/providers/ThemeProvider';

import type { AuthInfoResponse } from './models/Info';

const App = () => {
    const handleRouteChange = async ({ navigate, location }: { navigate: NavigateFunction; location: Location }) => {
        if (!isAuthenticated() || location.pathname === '/login') {
            return;
        }

        try {
            const info: AuthInfoResponse = await axios.get('/info', {
                params: {
                    refresh_cookie: true,
                },
            });

            if (info?.status !== 'success') {
                localStorage.removeItem('auth');
                // Save current path for redirect after login
                const currentPath = location.pathname;
                // Only save if it's not the default route
                const returnParam = currentPath !== '/chat/new' ? `?returnUrl=${encodeURIComponent(currentPath)}` : '';
                navigate(`/login${returnParam}`);
                return;
            }

            localStorage.setItem('auth', JSON.stringify(info.data));
        } catch {
            localStorage.removeItem('auth');
            // Save current path for redirect after login
            const currentPath = location.pathname;
            // Only save if it's not the default route
            const returnParam = currentPath !== '/chat/new' ? `?returnUrl=${encodeURIComponent(currentPath)}` : '';
            navigate(`/login${returnParam}`);
        }
    };

    const renderProtectedRoute = () => (
        <ProtectedRoute>
            <AppLayout />
        </ProtectedRoute>
    );

    const renderPublicRoute = () => (
        <PublicRoute>
            <Login />
        </PublicRoute>
    );

    return (
        <ApolloProvider client={client}>
            <ThemeProvider>
                <BrowserRouter>
                    <RouteChangeTracker onRouteChange={handleRouteChange} />
                    <Suspense fallback={<PageLoader />}>
                        <Routes>
                            {/* private routes */}
                            <Route element={renderProtectedRoute()}>
                                <Route
                                    path="chat"
                                    element={<Chat />}
                                />

                                <Route
                                    path="chat/:flowId"
                                    element={<Chat />}
                                />
                            </Route>

                            {/* report routes */}
                            <Route
                                path="chat/:flowId/report"
                                element={
                                    <ProtectedRoute>
                                        <Report />
                                    </ProtectedRoute>
                                }
                            />

                            {/* public routes */}
                            <Route
                                path="login"
                                element={renderPublicRoute()}
                            />

                            <Route
                                path="oauth/result"
                                element={<OAuthResult />}
                            />

                            {/* other routes */}
                            <Route
                                path="*"
                                element={<Navigate to="/chat/new" />}
                            />
                            <Route
                                path="/"
                                element={<Navigate to="/chat/new" />}
                            />
                        </Routes>
                    </Suspense>
                </BrowserRouter>
            </ThemeProvider>
        </ApolloProvider>
    );
};

export default App;
