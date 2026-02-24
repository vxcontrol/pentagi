import { ApolloProvider } from '@apollo/client';
import { lazy, Suspense } from 'react';
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
import { FavoritesProvider } from '@/providers/favorites-provider';
import { FlowProvider } from '@/providers/flow-provider';
import { ProvidersProvider } from '@/providers/providers-provider';
import { SidebarFlowsProvider } from '@/providers/sidebar-flows-provider';
import { TemplatesProvider } from '@/providers/templates-provider';
import { ThemeProvider } from '@/providers/theme-provider';
import { UserProvider } from '@/providers/user-provider';

import { SystemSettingsProvider } from './providers/system-settings-provider';

const Dashboard = lazy(() => import('@/pages/dashboard/dashboard'));
const Flow = lazy(() => import('@/pages/flows/flow'));
const FlowReport = lazy(() => import('@/pages/flows/flow-report'));
const Flows = lazy(() => import('@/pages/flows/flows'));
const NewFlow = lazy(() => import('@/pages/flows/new-flow'));
const Login = lazy(() => import('@/pages/login'));
const Template = lazy(() => import('@/pages/templates/template'));
const Templates = lazy(() => import('@/pages/templates/templates'));
const OAuthResult = lazy(() => import('@/pages/oauth-result'));
const SettingsAPITokens = lazy(() => import('@/pages/settings/settings-api-tokens'));
const SettingsPrompt = lazy(() => import('@/pages/settings/settings-prompt'));
const SettingsPrompts = lazy(() => import('@/pages/settings/settings-prompts'));
const SettingsProvider = lazy(() => import('@/pages/settings/settings-provider'));
const SettingsProviders = lazy(() => import('@/pages/settings/settings-providers'));

const App = () => {
    const renderProtectedRoute = () => (
        <ProtectedRoute>
            <SystemSettingsProvider>
                <ProvidersProvider>
                    <SidebarFlowsProvider>
                        <AppLayout />
                    </SidebarFlowsProvider>
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
                        <FavoritesProvider>
                            <TemplatesProvider>
                                <Suspense fallback={<PageLoader />}>
                                    <Routes>
                                        {/* private routes */}
                                        <Route element={renderProtectedRoute()}>
                                            {/* Main layout for chat pages */}
                                            <Route element={<MainLayout />}>
                                                <Route
                                                    element={<Dashboard />}
                                                    path="dashboard"
                                                />

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

                                                <Route
                                                    element={<Templates />}
                                                    path="templates"
                                                />
                                                <Route
                                                    element={<Template />}
                                                    path="templates/:templateId"
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
                                                <Route
                                                    element={<SettingsAPITokens />}
                                                    path="api-tokens"
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
                                            element={<Navigate to="/dashboard" />}
                                            path="/"
                                        />
                                        <Route
                                            element={<Navigate to="/dashboard" />}
                                            path="*"
                                        />
                                    </Routes>
                                </Suspense>
                            </TemplatesProvider>
                        </FavoritesProvider>
                    </UserProvider>
                </BrowserRouter>
            </ThemeProvider>
        </ApolloProvider>
    );
};

export default App;
