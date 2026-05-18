import type { ComponentType } from 'react';

import {
    useFlowQuery,
    useFlowTemplateQuery,
    useKnowledgeDocumentQuery,
    useSettingsProvidersQuery,
} from '@/graphql/types';

import { apolloTitle } from './apollo-title';
import { formatPromptId } from './format-prompt-id';
import { type RouteParams } from './render-title';

export interface RouteTitleHandle {
    title: TitleResolver;
}

/**
 * A `handle.title` value can be one of three forms:
 *   - `string` — fully static, known at build time.
 *   - `(params) => string` — derived synchronously from URL params.
 *   - `ComponentType<{ params }>` — reactive (e.g. subscribes to Apollo
 *     cache for resource-driven titles). Must be PascalCase-named so the
 *     runtime detector in `DocumentTitle` distinguishes it from a plain
 *     resolver function.
 */
export type TitleResolver = ((params: RouteParams) => string) | ComponentType<{ params: RouteParams }> | string;

/**
 * Single source of truth for every route's document `<title>`. `app.tsx`
 * imports nothing from Apollo for title purposes — it only spreads handles
 * from this registry onto the matching <Route>.
 */
export const routeTitles = {
    apiTokens: { title: 'API tokens' },
    dashboard: { title: 'Dashboard' },
    flow: {
        title: apolloTitle({
            select: (data, { flowId }) =>
                data?.flow?.title && flowId ? `Flow #${flowId} — ${data.flow.title}` : 'Flow',
            useQuery: useFlowQuery,
            variables: ({ flowId }) => (flowId ? { id: flowId } : null),
        }),
    },
    flowReport: { title: 'Flow report' },
    flows: { title: 'Flows' },
    knowledge: {
        title: apolloTitle({
            select: (data, { knowledgeId }) =>
                knowledgeId === 'new' ? 'New knowledge' : data?.knowledgeDocument?.question || 'Knowledge',
            useQuery: useKnowledgeDocumentQuery,
            variables: ({ knowledgeId }) => (!knowledgeId || knowledgeId === 'new' ? null : { id: knowledgeId }),
        }),
    },
    knowledges: { title: 'Knowledges' },
    login: { title: 'Login' },
    newFlow: { title: 'New flow' },
    oauth: { title: 'OAuth' },
    prompt: {
        title: (params: RouteParams) => (params.promptId ? formatPromptId(params.promptId) : 'Prompt'),
    },
    prompts: { title: 'Prompts' },

    provider: {
        title: apolloTitle({
            select: (data, { providerId }) => {
                if (providerId === 'new') {
                    return 'New provider';
                }

                const provider = data?.settingsProviders.userDefined?.find(
                    (candidate) => String(candidate.id) === providerId,
                );

                return provider?.name || 'Provider';
            },
            useQuery: useSettingsProvidersQuery,
            variables: ({ providerId }) => (providerId === 'new' ? null : {}),
        }),
    },

    providers: { title: 'Providers' },

    resources: { title: 'Resources' },

    template: {
        title: apolloTitle({
            select: (data, { templateId }) =>
                templateId === 'new' ? 'New template' : data?.flowTemplate?.title || 'Template',
            useQuery: useFlowTemplateQuery,
            variables: ({ templateId }) => (!templateId || templateId === 'new' ? null : { templateId }),
        }),
    },

    templates: { title: 'Templates' },
} as const satisfies Record<string, RouteTitleHandle>;
