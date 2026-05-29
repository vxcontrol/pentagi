import { ApolloProvider } from '@apollo/client/react';
import { lazy, Suspense } from 'react';
import {
    createBrowserRouter,
    createRoutesFromElements,
    Navigate,
    Outlet,
    Route,
    RouterProvider,
} from 'react-router-dom';

import AppLayout from '@/components/layouts/app-layout';
import FlowsLayout from '@/components/layouts/flows-layout';
import MainLayout from '@/components/layouts/main-layout';
import SettingsLayout from '@/components/layouts/settings-layout';
import ProtectedRoute from '@/components/routes/protected-route';
import PublicRoute from '@/components/routes/public-route';
import { DocumentTitle } from '@/components/shared/document-title';
import PageLoader from '@/components/shared/page-loader';
import { Toaster } from '@/components/ui/sonner';
import client from '@/lib/apollo';
import { routeTitles } from '@/lib/route-titles';
import { FavoritesProvider } from '@/providers/favorites-provider';
import { FlowProvider } from '@/providers/flow-provider';
import { KnowledgesProvider } from '@/providers/knowledges-provider';
import { ProvidersProvider } from '@/providers/providers-provider';
import { ResourcesProvider } from '@/providers/resources-provider';
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
const Knowledge = lazy(() => import('@/pages/knowledges/knowledge'));
const Knowledges = lazy(() => import('@/pages/knowledges/knowledges'));
const Resources = lazy(() => import('@/pages/resources/resources'));
const Template = lazy(() => import('@/pages/templates/template'));
const Templates = lazy(() => import('@/pages/templates/templates'));
const OAuthResult = lazy(() => import('@/pages/oauth-result'));
const SettingsAPITokens = lazy(() => import('@/pages/settings/settings-api-tokens'));
const SettingsPrompt = lazy(() => import('@/pages/settings/settings-prompt'));
const SettingsPrompts = lazy(() => import('@/pages/settings/settings-prompts'));
const SettingsProvider = lazy(() => import('@/pages/settings/settings-provider'));
const SettingsProviders = lazy(() => import('@/pages/settings/settings-providers'));

function FlowWithProvider() {
    return (
        <FlowProvider>
            <Flow />
        </FlowProvider>
    );
}

function KnowledgesLayout() {
    return (
        <KnowledgesProvider>
            <Outlet />
        </KnowledgesProvider>
    );
}

function ProtectedAppLayout() {
    return (
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
}

function ProtectedReportLayout() {
    return (
        <ProtectedRoute>
            <SystemSettingsProvider>
                <FlowReport />
            </SystemSettingsProvider>
        </ProtectedRoute>
    );
}

function PublicLoginLayout() {
    return (
        <PublicRoute>
            <Login />
        </PublicRoute>
    );
}

// Root layout for the data router. Everything that previously sat between
// `<BrowserRouter>` and `<Routes>` (providers, Suspense) lives here so it has
// access to router hooks (`useNavigate`, `useLocation`, ...) while still being
// rendered under the data router. This is what enables `useBlocker` and other
// data-router-only features inside our pages.
function RootLayout() {
    return (
        <UserProvider>
            <FavoritesProvider>
                <TemplatesProvider>
                    <ResourcesProvider>
                        {/* Document <title> driven by route handles — lives in the
                            shell so it survives navigation between sibling detail
                            routes (templates/:id → templates/:id') without
                            flashing a generic fallback during data fetch. */}
                        <DocumentTitle />
                        <Suspense fallback={<PageLoader />}>
                            <Outlet />
                        </Suspense>
                    </ResourcesProvider>
                </TemplatesProvider>
            </FavoritesProvider>
        </UserProvider>
    );
}

const router = createBrowserRouter(
    createRoutesFromElements(
        <Route element={<RootLayout />}>
            {/* private routes */}
            <Route element={<ProtectedAppLayout />}>
                {/* Main layout for chat pages */}
                <Route element={<MainLayout />}>
                    <Route
                        element={<Dashboard />}
                        handle={routeTitles.dashboard}
                        path="dashboard"
                    />

                    {/* Flows section with FlowsProvider */}
                    <Route element={<FlowsLayout />}>
                        <Route
                            element={<Flows />}
                            handle={routeTitles.flows}
                            path="flows"
                        />
                        <Route
                            element={<NewFlow />}
                            handle={routeTitles.newFlow}
                            path="flows/new"
                        />
                        <Route
                            element={<FlowWithProvider />}
                            handle={routeTitles.flow}
                            path="flows/:flowId"
                        />
                    </Route>

                    <Route
                        element={<Templates />}
                        handle={routeTitles.templates}
                        path="templates"
                    />
                    <Route
                        element={<Template />}
                        handle={routeTitles.template}
                        path="templates/:templateId"
                    />

                    <Route element={<KnowledgesLayout />}>
                        <Route
                            element={<Knowledges />}
                            handle={routeTitles.knowledges}
                            path="knowledges"
                        />
                        <Route
                            element={<Knowledge />}
                            handle={routeTitles.knowledge}
                            path="knowledges/:knowledgeId"
                        />
                    </Route>

                    <Route
                        element={<Resources />}
                        handle={routeTitles.resources}
                        path="resources"
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
                        handle={routeTitles.providers}
                        path="providers"
                    />
                    <Route
                        element={<SettingsProvider />}
                        handle={routeTitles.provider}
                        path="providers/:providerId"
                    />
                    <Route
                        element={<SettingsPrompts />}
                        handle={routeTitles.prompts}
                        path="prompts"
                    />
                    <Route
                        element={<SettingsPrompt />}
                        handle={routeTitles.prompt}
                        path="prompts/:promptId"
                    />
                    <Route
                        element={<SettingsAPITokens />}
                        handle={routeTitles.apiTokens}
                        path="api-tokens"
                    />
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
                element={<ProtectedReportLayout />}
                handle={routeTitles.flowReport}
                path="flows/:flowId/report"
            />

            {/* public routes */}
            <Route
                element={<PublicLoginLayout />}
                handle={routeTitles.login}
                path="login"
            />

            <Route
                element={<OAuthResult />}
                handle={routeTitles.oauth}
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
        </Route>,
    ),
);

function App() {
    return (
        <ApolloProvider client={client}>
            <ThemeProvider>
                <Toaster />
                <RouterProvider router={router} />
            </ThemeProvider>
        </ApolloProvider>
    );
}

export default App;
