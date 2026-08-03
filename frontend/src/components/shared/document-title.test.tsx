import { ApolloClient, ApolloLink, gql, InMemoryCache, type TypedDocumentNode } from '@apollo/client';
import { ApolloProvider } from '@apollo/client/react';
import { render, waitFor } from '@testing-library/react';
import { createMemoryRouter, Outlet, RouterProvider } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import { apolloTitle } from '@/lib/route-titles/apollo-title';

import { DocumentTitle } from './document-title';

const renderAt = (initialPath: string, routes: Parameters<typeof createMemoryRouter>[0]) => {
    const router = createMemoryRouter(routes, { initialEntries: [initialPath] });

    return render(<RouterProvider router={router} />);
};

describe('DocumentTitle', () => {
    it('renders APP_NAME when no matched route exposes a title handle', async () => {
        renderAt('/anywhere', [
            {
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

    it('renders a title component produced by apolloTitle()', async () => {
        // Cache is pre-populated so the `cache-only` read stays hermetic (no network).
        const customQuery = gql`
            query Custom {
                name
            }
        ` as TypedDocumentNode<{ name: string }, Record<string, never>>;
        const client = new ApolloClient({ cache: new InMemoryCache(), link: ApolloLink.empty() });
        client.writeQuery({ data: { name: 'thing' }, query: customQuery });

        const CustomTitle = apolloTitle<{ name: string }, Record<string, never>>({
            document: customQuery,
            select: (data, { id }) => `Custom #${id} ${data?.name ?? ''}`.trim(),
            variables: () => ({}),
        });

        const router = createMemoryRouter(
            [
                {
                    children: [{ element: <span>page</span>, handle: { title: CustomTitle }, path: 'items/:id' }],
                    element: (
                        <>
                            <DocumentTitle />
                            <Outlet />
                        </>
                    ),
                    path: '/',
                },
            ],
            { initialEntries: ['/items/42'] },
        );
        render(
            <ApolloProvider client={client}>
                <RouterProvider router={router} />
            </ApolloProvider>,
        );

        await waitFor(() => expect(document.title).toBe('Custom #42 thing — PentAGI'));
    });

    it('treats an unmarked function as a plain resolver, not a component', async () => {
        const resolveTitle = (params: Record<string, string | undefined>) => `Item ${params.id}`;

        renderAt('/items/7', [
            {
                children: [{ element: <span>page</span>, handle: { title: resolveTitle }, path: 'items/:id' }],
                element: (
                    <>
                        <DocumentTitle />
                        <Outlet />
                    </>
                ),
                path: '/',
            },
        ]);

        await waitFor(() => expect(document.title).toBe('Item 7 — PentAGI'));
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

        // Route-level convention: pages that don't want a title prefix return ''
        // intentionally, rather than omitting the handle.
        await waitFor(() => expect(document.title).toBe('PentAGI'));
    });
});
