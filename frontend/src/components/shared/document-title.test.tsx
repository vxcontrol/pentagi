import { render, waitFor } from '@testing-library/react';
import { createMemoryRouter, Outlet, RouterProvider } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import { DocumentTitle } from './document-title';

const renderAt = (initialPath: string, routes: Parameters<typeof createMemoryRouter>[0]) => {
    const router = createMemoryRouter(routes, { initialEntries: [initialPath] });

    return render(<RouterProvider router={router} />);
};

describe('DocumentTitle', () => {
    it('renders APP_NAME when no matched route exposes a title handle', async () => {
        renderAt('/anywhere', [
            {
                // No child route handle — DocumentTitle should fall back to APP_NAME only.
                children: [{ element: <span>page</span>, path: 'anywhere' }],
                element: (
                    <>
                        <DocumentTitle />
                        <Outlet />
                    </>
                ),
                path: '/',
            },
        ]);

        await waitFor(() => expect(document.title).toBe('PentAGI'));
    });

    it('renders a static title from handle', async () => {
        renderAt('/dashboard', [
            {
                children: [{ element: <span>page</span>, handle: { title: 'Dashboard' }, path: 'dashboard' }],
                element: (
                    <>
                        <DocumentTitle />
                        <Outlet />
                    </>
                ),
                path: '/',
            },
        ]);

        await waitFor(() => expect(document.title).toBe('Dashboard — PentAGI'));
    });

    it('renders a derived title from a handle.title function reading params', async () => {
        renderAt('/prompts/agentSelector', [
            {
                children: [
                    {
                        element: <span>page</span>,
                        handle: {
                            title: ({ promptId }: Record<string, string | undefined>) =>
                                promptId
                                    ? promptId.replaceAll(/([A-Z])/g, ' $1').replace(/^./, (s) => s.toUpperCase())
                                    : 'Prompt',
                        },
                        path: 'prompts/:promptId',
                    },
                ],
                element: (
                    <>
                        <DocumentTitle />
                        <Outlet />
                    </>
                ),
                path: '/',
            },
        ]);

        await waitFor(() => expect(document.title).toBe('Agent Selector — PentAGI'));
    });

    it('renders the deepest matching handle (child wins over parent)', async () => {
        renderAt('/settings/api-tokens', [
            {
                children: [
                    {
                        children: [
                            { element: <span>tokens</span>, handle: { title: 'API tokens' }, path: 'api-tokens' },
                        ],
                        element: <Outlet />,
                        handle: { title: 'Settings' },
                        path: 'settings',
                    },
                ],
                element: (
                    <>
                        <DocumentTitle />
                        <Outlet />
                    </>
                ),
                path: '/',
            },
        ]);

        await waitFor(() => expect(document.title).toBe('API tokens — PentAGI'));
    });

    it('renders a title from handle.titleComponent', async () => {
        function CustomTitle({ params }: { params: Record<string, string | undefined> }) {
            // React 19 only hoists <title> when the child is a single string
            // (not text + interpolation + text), so compute the string ahead.
            return <title>{`Custom #${params.id} — PentAGI`}</title>;
        }

        renderAt('/items/42', [
            {
                children: [{ element: <span>page</span>, handle: { titleComponent: CustomTitle }, path: 'items/:id' }],
                element: (
                    <>
                        <DocumentTitle />
                        <Outlet />
                    </>
                ),
                path: '/',
            },
        ]);

        await waitFor(() => expect(document.title).toBe('Custom #42 — PentAGI'));
    });

    it('prefers titleComponent over static title when both are present on the same handle', async () => {
        function CustomTitle() {
            return <title>From component — PentAGI</title>;
        }

        renderAt('/x', [
            {
                children: [
                    {
                        element: <span>page</span>,
                        handle: { title: 'From static', titleComponent: CustomTitle },
                        path: 'x',
                    },
                ],
                element: (
                    <>
                        <DocumentTitle />
                        <Outlet />
                    </>
                ),
                path: '/',
            },
        ]);

        await waitFor(() => expect(document.title).toBe('From component — PentAGI'));
    });

    it('falls back to APP_NAME when handle.title returns an empty string', async () => {
        renderAt('/x', [
            {
                children: [{ element: <span>page</span>, handle: { title: '' }, path: 'x' }],
                element: (
                    <>
                        <DocumentTitle />
                        <Outlet />
                    </>
                ),
                path: '/',
            },
        ]);

        // An empty string from the resolver is treated as "no title" — fall back
        // to APP_NAME alone. This guards the route-level convention: pages that
        // do not want a prefix can return '' instead of omitting the handle.
        await waitFor(() => expect(document.title).toBe('PentAGI'));
    });
});
