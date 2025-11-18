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
import Flow from '@/pages/flows/Flow';
import Flows from '@/pages/flows/Flows';
import NewFlow from '@/pages/flows/NewFlow';
import Login from '@/pages/Login';
import OAuthResult from '@/pages/OAuthResult';
import Report from '@/pages/Report';
import SettingsPrompt from '@/pages/settings/SettingsPrompt';
import SettingsPrompts from '@/pages/settings/SettingsPrompts';
import SettingsProvider from '@/pages/settings/SettingsProvider';
import SettingsProviders from '@/pages/settings/SettingsProviders';
import { ProvidersProvider } from '@/providers/ProvidersProvider';
import ThemeProvider from '@/providers/ThemeProvider';
import { UserProvider } from '@/providers/UserProvider';

const App = () => {
    const renderProtectedRoute = () => (
        <ProtectedRoute>
            <ProvidersProvider>
                <AppLayout />
            </ProvidersProvider>
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
                                            element={<Flows />}
                                            path="flows"
                                        />
                                        <Route
                                            element={<NewFlow />}
                                            path="flows/new"
                                        />
                                        <Route
                                            element={<Flow />}
                                            path="flows/:flowId"
                                        />
                                    </Route>

                                    {/* Settings with nested routes */}
                                    <Route
                                        element={<SettingsLayout />}
                                        path="settings"
                                    >
                                        <Route
                                            element={
                                                <Navigate
                                                    replace
                                                    to="providers"
                                                />
                                            }
                                            index
                                        />
                                        <Route
                                            element={<SettingsProviders />}
                                            path="providers"
                                        />
                                        <Route
                                            element={<SettingsProvider />}
                                            path="providers/:providerId"
                                        />
                                        <Route
                                            element={<SettingsPrompts />}
                                            path="prompts"
                                        />
                                        <Route
                                            element={<SettingsPrompt />}
                                            path="prompts/:promptId"
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
                                            element={
                                                <Navigate
                                                    replace
                                                    to="/settings/providers"
                                                />
                                            }
                                            path="*"
                                        />
                                    </Route>
                                </Route>

                                {/* report routes */}
                                <Route
                                    element={
                                        <ProtectedRoute>
                                            <Report />
                                        </ProtectedRoute>
                                    }
                                    path="flows/:flowId/report"
                                />

                                {/* public routes */}
                                <Route
                                    element={renderPublicRoute()}
                                    path="login"
                                />

                                <Route
                                    element={<OAuthResult />}
                                    path="oauth/result"
                                />

                                {/* other routes */}
                                <Route
                                    element={<Navigate to="/flows" />}
                                    path="/"
                                />
                                <Route
                                    element={<Navigate to="/flows" />}
                                    path="*"
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
