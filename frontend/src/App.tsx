import { ApolloProvider } from '@apollo/client';
import { Suspense } from 'react';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';

import AppLayout from '@/components/layouts/AppLayout';
import MainLayout from '@/components/layouts/MainLayout';
import SettingsLayout from '@/components/layouts/SettingsLayout';
import ProtectedRoute from '@/components/routes/ProtectedRoute';
import PublicRoute from '@/components/routes/PublicRoute';
import PageLoader from '@/components/shared/PageLoader';
import { Toaster } from '@/components/ui/sonner';
import client from '@/lib/apollo';
import Chat from '@/pages/Chat';
import Flows from '@/pages/Flows';
import Login from '@/pages/Login';
import OAuthResult from '@/pages/OAuthResult';
import Report from '@/pages/Report';
import SettingsPrompt from '@/pages/settings/SettingsPrompt';
import SettingsPrompts from '@/pages/settings/SettingsPrompts';
import SettingsProvider from '@/pages/settings/SettingsProvider';
import SettingsProviders from '@/pages/settings/SettingsProviders';
import ThemeProvider from '@/providers/ThemeProvider';
import { UserProvider } from '@/providers/UserProvider';

const App = () => {
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
                <Toaster />
                <BrowserRouter>
                    <UserProvider>
                        <Suspense fallback={<PageLoader />}>
                            <Routes>
                                {/* private routes */}
                                <Route element={renderProtectedRoute()}>
                                    {/* Main layout for chat pages */}
                                    <Route element={<MainLayout />}>
                                        <Route
                                            path="flows"
                                            element={<Flows />}
                                        />
                                        <Route
                                            path="flows/:flowId"
                                            element={<Chat />}
                                        />
                                    </Route>

                                    {/* Settings with nested routes */}
                                    <Route
                                        path="settings"
                                        element={<SettingsLayout />}
                                    >
                                        <Route
                                            index
                                            element={(
                                                <Navigate
                                                    to="providers"
                                                    replace
                                                />
                                            )}
                                        />
                                        <Route
                                            path="providers"
                                            element={<SettingsProviders />}
                                        />
                                        <Route
                                            path="providers/:providerId"
                                            element={<SettingsProvider />}
                                        />
                                        <Route
                                            path="prompts"
                                            element={<SettingsPrompts />}
                                        />
                                        <Route
                                            path="prompts/:promptId"
                                            element={<SettingsPrompt />}
                                        />
                                        {/* <Route
                                        path="mcp-servers"
                                        element={<SettingsMcpServers />}
                                    />
                                    <Route
                                        path="mcp-servers/new"
                                        element={<SettingsMcpServer />}
                                    />
                                    <Route
                                        path="mcp-servers/:mcpServerId"
                                        element={<SettingsMcpServer />}
                                    /> */}
                                        {/* Catch-all route for unknown settings paths */}
                                        <Route
                                            path="*"
                                            element={(
                                                <Navigate
                                                    to="/settings/providers"
                                                    replace
                                                />
                                            )}
                                        />
                                    </Route>
                                </Route>

                                {/* report routes */}
                                <Route
                                    path="flows/:flowId/report"
                                    element={(
                                        <ProtectedRoute>
                                            <Report />
                                        </ProtectedRoute>
                                    )}
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
                                    path="/"
                                    element={<Navigate to="/flows" />}
                                />
                                <Route
                                    path="*"
                                    element={<Navigate to="/flows" />}
                                />
                            </Routes>
                        </Suspense>
                    </UserProvider>
                </BrowserRouter>
            </ThemeProvider>
        </ApolloProvider>
    );
};

export default App;
