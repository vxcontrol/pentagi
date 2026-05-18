import {
    useFlowQuery,
    useFlowTemplateQuery,
    useKnowledgeDocumentQuery,
    useSettingsProvidersQuery,
} from '@/graphql/types';

const APP_NAME = 'PentAGI';

type RouteParams = Record<string, string | undefined>;

const renderTitle = (label: null | string) => <title>{label ? `${label} — ${APP_NAME}` : APP_NAME}</title>;

// Reactive title components used by detail routes via `handle.titleComponent`.
// Each uses `fetchPolicy: 'cache-only'` so it subscribes to the Apollo cache
// without issuing a duplicate HTTP request — the destination page's own
// query fills the cache. The variables shape mirrors the page so cache
// lookups hit.

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

    const provider = data?.settingsProviders.userDefined?.find((candidate) => String(candidate.id) === providerId);

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
