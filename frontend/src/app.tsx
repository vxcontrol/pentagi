import { ApolloProvider } from '@apollo/client';
import { Suspense } from 'react';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';

import AppLayout from '@/components/layouts/app-layout';
import FlowsLayout from '@/components/layouts/flows-layout';
import MainLayout from '@/components/layouts/main-layout';
import SettingsLayout from '@/components/layouts/settings-layout';
import ProtectedRoute from '@/components/routes/protected-route';
import PublicRoute from '@/components/routes/public-route';
import PageLoader from '@/components/shared/page-loader';
import { Toaster } from '@/components/ui/sonner';
import client from '@/lib/apollo';
import Flow from '@/pages/flows/flow';
import FlowReport from '@/pages/flows/flow-report';
import Flows from '@/pages/flows/Flows';
import NewFlow from '@/pages/flows/new-flow';
import Login from '@/pages/login';
import OAuthResult from '@/pages/oauth-result';
import SettingsPrompt from '@/pages/settings/settings-prompt';
import SettingsPrompts from '@/pages/settings/settings-prompts';
import SettingsProvider from '@/pages/settings/settings-provider';
import SettingsProviders from '@/pages/settings/settings-providers';
import { FlowProvider } from '@/providers/flow-provider';
import { ProvidersProvider } from '@/providers/providers-provider';
import ThemeProvider from '@/providers/theme-provider';
import { UserProvider } from '@/providers/user-provider';

import { SystemSettingsProvider } from './providers/system-settings-provider';

const App = () => {
    const renderProtectedRoute = () => (
        <ProtectedRoute>
            <SystemSettingsProvider>
                <ProvidersProvider>
                    <AppLayout />
                </ProvidersProvider>
            </SystemSettingsProvider>
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
                                        {/* Flows section with FlowsProvider */}
                                        <Route element={<FlowsLayout />}>
                                            <Route
                                                element={<Flows />}
                                                path="flows"
                                            />
                                            <Route
                                                element={<NewFlow />}
                                                path="flows/new"
                                            />
                                            <Route
                                                element={
                                                    <FlowProvider>
                                                        <Flow />
                                                    </FlowProvider>
                                                }
                                                path="flows/:flowId"
                                            />
                                        </Route>

                                        {/* Other pages can be added here without FlowsProvider */}
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
                                            <SystemSettingsProvider>
                                                <FlowReport />
                                            </SystemSettingsProvider>
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
