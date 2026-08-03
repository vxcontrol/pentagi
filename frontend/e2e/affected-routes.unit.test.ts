import { describe, expect, it } from 'vitest';

import type { RouteSource } from './affected-routes.ts';

import { affectedRoutes } from './affected-routes.ts';
import { ROUTE_MANIFEST } from './routes.ts';

const MANIFEST: RouteSource[] = [
    {
        path: '/flows',
        sources: ['src/pages/flows/flows.tsx', 'src/features/flows', 'src/providers/flows-provider.tsx'],
    },
    { path: '/templates', sources: ['src/pages/templates', 'src/providers/templates-provider.tsx'] },
    { path: '/settings/api-tokens', sources: ['src/pages/settings/settings-api-tokens.tsx'] },
];

describe('affectedRoutes', () => {
    it('scopes to the routes owning the changed files', () => {
        const result = affectedRoutes(['frontend/src/features/flows/flow-message.tsx'], MANIFEST);

        expect(result.map((route) => route.path)).toEqual(['/flows']);
    });

    it('matches an exact-file source, not just a directory', () => {
        const result = affectedRoutes(['frontend/src/pages/settings/settings-api-tokens.tsx'], MANIFEST);

        expect(result.map((route) => route.path)).toEqual(['/settings/api-tokens']);
    });

    it('falls back to every route for a page no manifest route renders', () => {
        const result = affectedRoutes(['frontend/src/pages/flows/flow-report.tsx'], MANIFEST);

        expect(result).toHaveLength(MANIFEST.length);
    });

    it('returns every route when shared infra changes', () => {
        const result = affectedRoutes(['frontend/src/lib/apollo.ts'], MANIFEST);

        expect(result).toHaveLength(MANIFEST.length);
    });

    it('returns every route when the e2e engine itself changes', () => {
        const result = affectedRoutes(['frontend/e2e/mocks/world.ts'], MANIFEST);

        expect(result).toHaveLength(MANIFEST.length);
    });

    it('ignores non-frontend changes', () => {
        const result = affectedRoutes(['backend/pkg/server/router.go'], MANIFEST);

        expect(result).toEqual([]);
    });

    it('collapses multiple changes in one route to a single entry', () => {
        const result = affectedRoutes(
            ['frontend/src/pages/flows/flows.tsx', 'frontend/src/features/flows/flow-form.tsx'],
            MANIFEST,
        );

        expect(result.map((route) => route.path)).toEqual(['/flows']);
    });

    // The fake manifest above cannot catch ownership drift in the real one: a shared provider claimed
    // by a single route silently scopes a run away from the other routes that mount its consumers.
    it('scopes a shared provider to every route that mounts a consumer', () => {
        const result = affectedRoutes(['frontend/src/providers/templates-provider.tsx'], ROUTE_MANIFEST);

        expect(result.map((route) => route.path).sort()).toEqual(['/flows/5', '/templates']);
    });
});
