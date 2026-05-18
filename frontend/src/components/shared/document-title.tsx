import type { ComponentType } from 'react';

import { useMatches } from 'react-router-dom';

import {
    useFlowQuery,
    useFlowTemplateQuery,
    useKnowledgeDocumentQuery,
    useSettingsProvidersQuery,
} from '@/graphql/types';

const APP_NAME = 'PentAGI';

export interface RouteHandleWithTitle {
    titleComponent: TitleComponent;
}

type RouteParams = Record<string, string | undefined>;

type TitleComponent = ComponentType<{ params: RouteParams }>;

const hasTitleComponent = (handle: unknown): handle is RouteHandleWithTitle =>
    typeof handle === 'object' &&
    handle !== null &&
    'titleComponent' in handle &&
    typeof (handle as RouteHandleWithTitle).titleComponent === 'function';

const renderTitle = (label: null | string) => <title>{label ? `${label} — ${APP_NAME}` : APP_NAME}</title>;

/**
 * Renders the document `<title>` driven by react-router route handles. Walks
 * matches deepest-first; the first match exposing `handle.titleComponent`
 * wins. Lives in the app shell so it survives navigation between sibling
 * detail routes — that's what fixes the previous "Provider — PentAGI"
 * flash when DetailNavigation switches between siblings and the destination
 * page unmounts/remounts during data fetch.
 *
 * Routes without a `titleComponent` keep using `<PageTitle>` inside the page
 * component — the migration is route-by-route.
 */
export function DocumentTitle() {
    const matches = useMatches();

    for (let i = matches.length - 1; i >= 0; i--) {
        const handle = matches[i].handle;

        if (hasTitleComponent(handle)) {
            const TitleComponent = handle.titleComponent;

            return <TitleComponent params={matches[i].params} />;
        }
    }

    return null;
}

// Per-resource title components read the same Apollo cache the destination
// page is about to populate. `fetchPolicy: 'cache-only'` avoids a duplicate
// HTTP request — the page's own query fills the cache, this subscription
// reacts to it. The variables shape mirrors the page so cache lookups hit.

export function FlowTitle({ params }: { params: RouteParams }) {
    const flowId = params.flowId;
    const { data } = useFlowQuery({
        fetchPolicy: 'cache-only',
        skip: !flowId,
        variables: flowId ? { id: flowId } : undefined,
    });

    const flowTitle = data?.flow?.title;

    return renderTitle(flowTitle && flowId ? `Flow #${flowId} — ${flowTitle}` : 'Flow');
}

export function KnowledgeTitle({ params }: { params: RouteParams }) {
    const knowledgeId = params.knowledgeId;
    const isNew = knowledgeId === 'new';
    const { data } = useKnowledgeDocumentQuery({
        fetchPolicy: 'cache-only',
        skip: isNew || !knowledgeId,
        variables: !isNew && knowledgeId ? { id: knowledgeId } : undefined,
    });

    if (isNew) {
        return renderTitle('New knowledge');
    }

    return renderTitle(data?.knowledgeDocument?.question || 'Knowledge');
}

export function ProviderTitle({ params }: { params: RouteParams }) {
    const providerId = params.providerId;
    const isNew = providerId === 'new';
    const { data } = useSettingsProvidersQuery({
        fetchPolicy: 'cache-only',
        skip: isNew,
    });

    if (isNew) {
        return renderTitle('New provider');
    }

    // Provider detail isn't a separate query — `settingsProviders` returns the
    // full list and the page filters by id. We do the same here so a single
    // cache entry serves both.
    const provider = data?.settingsProviders.userDefined?.find((candidate) => candidate.id == providerId);

    return renderTitle(provider?.name || 'Provider');
}

export function TemplateTitle({ params }: { params: RouteParams }) {
    const templateId = params.templateId;
    const isNew = templateId === 'new';
    const { data } = useFlowTemplateQuery({
        fetchPolicy: 'cache-only',
        skip: isNew || !templateId,
        variables: templateId && !isNew ? { templateId } : undefined,
    });

    if (isNew) {
        return renderTitle('New template');
    }

    return renderTitle(data?.flowTemplate?.title || 'Template');
}
