import { skipToken, useQuery } from '@apollo/client/react';
import { useCallback, useEffect, useMemo } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';

import { ErrorState } from '@/components/shared/error-state';
import { Spinner } from '@/components/ui/spinner';
import {
    type DirtyFlags,
    documentToFormValues,
    type FormValues,
    formValuesToCreateInput,
    formValuesToUpdateInput,
    KnowledgeForm,
    newDocumentDefaults,
    type SubmitResult,
} from '@/features/knowledges/knowledge-form';
import { KnowledgeLayout } from '@/features/knowledges/knowledge-layout';
import { KnowledgeDocumentDocument } from '@/graphql/types';
import { isNotFoundError } from '@/lib/errors';
import { routes } from '@/lib/routes';
import { useKnowledges } from '@/providers/knowledges-provider';

function Knowledge() {
    const navigate = useNavigate();
    const { knowledgeId } = useParams<{ knowledgeId?: string }>();
    const { createKnowledge, updateKnowledge } = useKnowledges();

    const isNew = knowledgeId === 'new';
    const shouldFetch = Boolean(knowledgeId) && !isNew;

    const {
        data,
        error,
        loading: isLoadingKnowledge,
        refetch,
    } = useQuery(
        KnowledgeDocumentDocument,
        shouldFetch && knowledgeId ? { variables: { id: knowledgeId } } : skipToken,
    );

    const knowledge = data?.knowledgeDocument ?? null;
    // A real load failure that left nothing to show, as opposed to a genuine not-found: the page
    // renders it as an in-page ErrorState + Retry instead of bouncing to the list. Mirrors flow.
    const loadError = error && !knowledge && !isNotFoundError(error) ? error : undefined;

    useEffect(() => {
        // Redirect only when the document is genuinely gone (query settled with no document, or a
        // not-found error) — never on a transient load failure, which Retry can recover.
        if (isNew || isLoadingKnowledge || loadError) {
            return;
        }

        if (!knowledge) {
            toast.error('Knowledge document not found');
            navigate(routes.knowledges, { replace: true });
        }
    }, [isNew, isLoadingKnowledge, knowledge, loadError, navigate]);

    const initialValues = useMemo<FormValues>(
        () => (knowledge ? documentToFormValues(knowledge) : newDocumentDefaults),
        [knowledge],
    );

    const handleSubmit = useCallback(
        // `values` are the zod-parsed form output (trimmed, length-validated).
        // CREATE sends a full payload; UPDATE sends only fields the user
        // actually changed (`dirtyFields`) so untouched optional fields stay
        // untouched on the backend and explicit clears (e.g. wiping an existing
        // description) reach it as `""`.
        async (values: FormValues, dirtyFields: DirtyFlags): Promise<SubmitResult> => {
            if (isNew) {
                const created = await createKnowledge(formValuesToCreateInput(values));

                return {
                    document: created ?? undefined,
                    redirectTo: created?.id ? routes.knowledge(created.id) : routes.knowledges,
                };
            }

            if (!knowledgeId) {
                return {};
            }

            const updated = await updateKnowledge(knowledgeId, formValuesToUpdateInput(values, dirtyFields));

            return { document: updated ?? undefined };
        },
        [createKnowledge, isNew, knowledgeId, updateKnowledge],
    );

    if (loadError) {
        return (
            <KnowledgeLayout
                isNew={false}
                knowledge={null}
            >
                <div className="flex flex-1 flex-col gap-4 p-4">
                    <ErrorState
                        message={loadError.message}
                        onRetry={() => refetch()}
                        title="Error loading knowledge document"
                    />
                </div>
            </KnowledgeLayout>
        );
    }

    if (!isNew && !knowledge) {
        return (
            <KnowledgeLayout
                isNew={false}
                knowledge={knowledge}
            >
                <div className="flex flex-1 items-center justify-center">
                    <Spinner variant="circle" />
                </div>
            </KnowledgeLayout>
        );
    }

    return (
        <KnowledgeForm
            initialValues={initialValues}
            isNew={isNew}
            key={knowledgeId ?? 'new'}
            knowledge={knowledge}
            onSubmit={handleSubmit}
        />
    );
}

export default Knowledge;
