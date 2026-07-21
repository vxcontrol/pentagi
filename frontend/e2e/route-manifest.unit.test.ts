import { describe, expect, it } from 'vitest';

import { routes } from '@/lib/routes';

import { ROUTE_MANIFEST } from './routes.ts';

/**
 * Routes deliberately outside the manifest sweeps (nav, visual, a11y,
 * diff-scoping), each with the reason. Adding a route to lib/routes forces a
 * decision here: give it a manifest entry or list it with a reason — it cannot
 * silently stay out of every sweep. Dynamic builders (`routes.flow(id)`, …)
 * are functions and out of this static check's scope.
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

const staticPaths = (node: unknown): string[] => {
    if (typeof node === 'string') {
        return [node];
    }

    if (node && typeof node === 'object') {
        return Object.values(node).flatMap(staticPaths);
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
});
