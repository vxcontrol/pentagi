import { matchPath } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import { routes } from './routes';

// Mirrors the absolute route patterns from app.tsx; keep in sync when routes change there.
const builtUrlToPattern: [string, string][] = [
    [routes.dashboard, '/dashboard'],
    [routes.flows, '/flows'],
    [routes.newFlow, '/flows/new'],
    [routes.flow('abc'), '/flows/:flowId'],
    [routes.flowReport('abc'), '/flows/:flowId/report'],
    [routes.templates, '/templates'],
    [routes.newTemplate, '/templates/:templateId'],
    [routes.template('t1'), '/templates/:templateId'],
    [routes.knowledges, '/knowledges'],
    [routes.newKnowledge, '/knowledges/:knowledgeId'],
    [routes.knowledge('k1'), '/knowledges/:knowledgeId'],
    [routes.resources, '/resources'],
    [routes.settings.root, '/settings'],
    [routes.settings.account, '/settings/account'],
    [routes.settings.providers, '/settings/providers'],
    [routes.settings.provider('p1'), '/settings/providers/:providerId'],
    [routes.settings.newProvider(), '/settings/providers/:providerId'],
    [routes.settings.prompts, '/settings/prompts'],
    [routes.settings.prompt('p1'), '/settings/prompts/:promptId'],
    [routes.settings.apiTokens, '/settings/api-tokens'],
    [routes.login(), '/login'],
    [routes.oauthResult, '/oauth/result'],
    [routes.root, '/'],
];

describe('routes registry', () => {
    it.each(builtUrlToPattern)('builds %s, matched by route pattern %s', (url, pattern) => {
        expect(matchPath(pattern, url)).not.toBeNull();
    });

    it('encodes query params for builders that accept them', () => {
        expect(routes.flow('1', { tab: 'chat' })).toBe('/flows/1?tab=chat');
        expect(routes.flow('1')).toBe('/flows/1');
        expect(routes.settings.newProvider({ id: 'a', type: 'b' })).toBe('/settings/providers/new?id=a&type=b');
        expect(routes.settings.newProvider()).toBe('/settings/providers/new');
    });

    it('builds the login url with an encoded return path, except for non-returnable origins', () => {
        expect(routes.login('/dashboard')).toBe('/login?returnUrl=%2Fdashboard');
        expect(routes.login()).toBe('/login');
        expect(routes.login('/flows/new')).toBe('/login');
        expect(routes.login('/login')).toBe('/login');
    });
});
