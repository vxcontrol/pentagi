import type { Locator, Page } from '@playwright/test';

import { routes } from '@/lib/routes';

import type { Cassette } from './mocks/cassette.ts';

import { apiTokensCassette } from './mocks/cassettes/api-tokens.ts';
import { dashboardCassette } from './mocks/cassettes/dashboard.ts';
import { flowsCassette } from './mocks/cassettes/flows.ts';
import { knowledgesCassette } from './mocks/cassettes/knowledges.ts';
import { resourcesCassette } from './mocks/cassettes/resources.ts';
import { settingsPromptsCassette } from './mocks/cassettes/settings-prompts.ts';
import { settingsProvidersCassette } from './mocks/cassettes/settings-providers.ts';
import { templatesCassette } from './mocks/cassettes/templates.ts';

export interface RouteManifestEntry {
    cassette: () => Cassette;
    path: string;
    /** The route counts as rendered when this locator is visible. */
    ready: (page: Page) => Locator;
    /**
     * Owning src/ areas — the substrate for changed-files → affected-routes
     * mapping. A dir rendered by several routes must be listed under every one
     * of them: ownership never falls back to the run-everything path, so a
     * single-route listing silently skips the other consumers.
     */
    sources: string[];
}

/**
 * The routes swept by NAV/visual/a11y specs and CI diff-scoping. This is a
 * subset of the app's routes — route-manifest.unit.test.ts pins the excluded
 * remainder so a new route cannot silently stay out of every sweep. Paths come
 * from the app's own routes module so a route rename breaks this file at
 * compile time.
 */
export const ROUTE_MANIFEST: RouteManifestEntry[] = [
    {
        cassette: flowsCassette,
        path: routes.flows,
        ready: (page) => page.getByRole('row', { name: /E2E Alpha/ }),
        sources: ['src/pages/flows', 'src/features/flows', 'src/providers/flows-provider.tsx'],
    },
    {
        cassette: flowsCassette,
        path: routes.flow('5'),
        // The terminal mounts after the header, and the visual spec masks it —
        // capturing before it exists compares live pixels against a masked baseline.
        ready: (page) => page.locator('.xterm').first(),
        sources: [
            'src/pages/flows',
            'src/features/flows',
            'src/providers/flow-provider.tsx',
            'src/providers/flows-provider.tsx',
            // Rendered inside the detail page's tabs alongside their owning routes.
            'src/components/shared/file-manager',
            'src/components/dashboard',
            'src/features/resources',
        ],
    },
    {
        cassette: templatesCassette,
        path: routes.templates,
        ready: (page) => page.getByRole('row', { name: /E2E Seed Template/ }),
        sources: ['src/pages/templates', 'src/providers/templates-provider.tsx'],
    },
    {
        cassette: knowledgesCassette,
        path: routes.knowledges,
        ready: (page) => page.getByRole('row', { name: /E2E Seed Question/ }),
        sources: ['src/pages/knowledges', 'src/features/knowledges', 'src/providers/knowledges-provider.tsx'],
    },
    {
        cassette: apiTokensCassette,
        path: routes.settings.apiTokens,
        ready: (page) => page.getByRole('row', { name: /E2E seed token/ }),
        sources: ['src/pages/settings/settings-api-tokens.tsx'],
    },
    {
        cassette: dashboardCassette,
        path: routes.dashboard,
        ready: (page) => page.getByRole('heading', { name: 'Flows Activity Over Time' }),
        sources: ['src/pages/dashboard', 'src/components/dashboard'],
    },
    {
        cassette: settingsPromptsCassette,
        path: routes.settings.prompts,
        ready: (page) => page.getByRole('heading', { name: 'Agent Prompts' }),
        sources: ['src/pages/settings/settings-prompts.tsx'],
    },
    {
        cassette: settingsProvidersCassette,
        path: routes.settings.providers,
        ready: (page) => page.getByText('No providers configured'),
        sources: ['src/pages/settings/settings-providers.tsx'],
    },
    {
        cassette: resourcesCassette,
        path: routes.resources,
        ready: (page) => page.getByRole('treeitem', { name: /reports/ }),
        sources: ['src/pages/resources', 'src/features/resources', 'src/components/shared/file-manager'],
    },
];
