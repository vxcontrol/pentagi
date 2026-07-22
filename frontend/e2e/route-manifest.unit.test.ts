import { describe, expect, it } from 'vitest';

import { routes } from '@/lib/routes';

import { ROUTE_MANIFEST } from './routes.ts';

/**
 * Routes deliberately outside the manifest sweeps (nav, visual, a11y,
 * diff-scoping), each with the reason. Adding a route to lib/routes forces a
 * decision here: give it a manifest entry or list it with a reason — it cannot
 * silently stay out of every sweep.
 */
const EXCLUDED: Record<string, string> = {
    '/': 'redirects to /dashboard',
    '/flows/new': 'create form; no stable list/detail ready-state to sweep',
    '/knowledges/new': 'create-mode variant of the knowledge detail page',
    '/oauth/result': 'OAuth popup landing; only meaningful mid-OAuth-roundtrip',
    '/settings': 'redirects to /settings/account',
    '/settings/account': 'needs an account cassette + visual baseline before joining the sweep',
    '/templates/new': 'create-mode variant of the template detail page',
};

/** Route builders are functions, so the static walk below cannot see them. */
const DYNAMIC_ROUTES: Record<string, string> = {
    flow: 'manifest entry (routes.flow("5"))',
    flowReport: 'not swept: needs a finished-flow report cassette',
    knowledge: 'specs/crud/knowledges.spec.ts — detail page after create',
    login: 'specs/smoke.spec.ts + the /login a11y scan',
    'settings.newProvider': 'specs/settings/providers.spec.ts — opened from the empty state',
    'settings.prompt': 'specs/settings/prompt-detail.spec.ts',
    'settings.provider': 'not swept: provider form; unit-covered by settings-provider.test.tsx',
    template: 'specs/crud/template-detail.spec.ts',
};

const staticPaths = (node: unknown): string[] => {
    if (typeof node === 'string') {
        return [node];
    }

    if (node && typeof node === 'object') {
        return Object.values(node).flatMap(staticPaths);
    }

    return [];
};

const dynamicRouteKeys = (node: unknown, prefix = ''): string[] => {
    if (typeof node === 'function') {
        return [prefix];
    }

    if (node && typeof node === 'object') {
        return Object.entries(node).flatMap(([key, value]) =>
            dynamicRouteKeys(value, prefix ? `${prefix}.${key}` : key),
        );
    }

    return [];
};

describe('ROUTE_MANIFEST completeness', () => {
    const manifestPaths = new Set(ROUTE_MANIFEST.map((entry) => entry.path));

    it('covers or explicitly excludes every static app route', () => {
        const uncovered = staticPaths(routes).filter((path) => !manifestPaths.has(path) && !(path in EXCLUDED));

        expect(uncovered).toEqual([]);
    });

    it('keeps the exclusion list free of routes the manifest already covers', () => {
        expect(Object.keys(EXCLUDED).filter((path) => manifestPaths.has(path))).toEqual([]);
    });

    it('forces a coverage decision for every dynamic route builder', () => {
        expect(dynamicRouteKeys(routes).sort()).toEqual(Object.keys(DYNAMIC_ROUTES).sort());
    });
});
