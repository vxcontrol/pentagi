import { useCallback, useMemo } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Spinner } from '@/components/ui/spinner';
import {
    documentToFormValues,
    type FormValues,
    formValuesToCreateInput,
    formValuesToUpdateInput,
    KnowledgeForm,
    newDocumentDefaults,
    type SubmitResult,
} from '@/features/knowledges/knowledge-form';
import { KnowledgeLayout } from '@/features/knowledges/knowledge-layout';
import { useKnowledgeDocumentQuery } from '@/graphql/types';
import { useKnowledges } from '@/providers/knowledges-provider';

const Knowledge = () => {
    const navigate = useNavigate();
    const { knowledgeId } = useParams<{ knowledgeId?: string }>();
    const { createKnowledge, updateKnowledge } = useKnowledges();

    const isNew = knowledgeId === 'new';
    const shouldFetch = Boolean(knowledgeId) && !isNew;

    const { data, loading: isLoadingKnowledge } = useKnowledgeDocumentQuery({
        skip: !shouldFetch,
        variables: shouldFetch && knowledgeId ? { id: knowledgeId } : undefined,
    });

    const knowledge = data?.knowledgeDocument ?? null;
    const knowledgeName = knowledge?.question ?? null;

    const initialValues = useMemo<FormValues>(
        () => (knowledge ? documentToFormValues(knowledge) : newDocumentDefaults),
        [knowledge],
    );

    const handleSubmit = useCallback(
        async (values: FormValues): Promise<SubmitResult> => {
            if (isNew) {
                const created = await createKnowledge(formValuesToCreateInput(values));

                return {
                    document: created ?? undefined,
                    redirectTo: created?.id ? `/knowledges/${created.id}` : '/knowledges',
                };
            }

            if (!knowledgeId) {
                return {};
            }

            const updated = await updateKnowledge(knowledgeId, formValuesToUpdateInput(values));

            return { document: updated ?? undefined };
        },
        [createKnowledge, isNew, knowledgeId, updateKnowledge],
    );

    if (!isNew && isLoadingKnowledge) {
        return (
            <KnowledgeLayout
                isNew={false}
                knowledgeName={knowledgeName}
            >
                <div className="flex flex-1 items-center justify-center">
                    <Spinner variant="circle" />
                </div>
            </KnowledgeLayout>
        );
    }

    if (!isNew && !knowledge) {
        return (
            <KnowledgeLayout
                isNew={false}
                knowledgeName={knowledgeName}
            >
                <div className="flex flex-1 items-center justify-center p-4">
                    <Card className="w-full max-w-2xl">
                        <CardContent className="flex flex-col items-center gap-4 pt-6 text-center">
                            <h2 className="text-xl font-semibold">Knowledge not found</h2>
                            <p className="text-muted-foreground">
                                The knowledge document you are looking for does not exist.
                            </p>
                            <Button onClick={() => navigate('/knowledges')}>Back to Knowledges</Button>
                        </CardContent>
                    </Card>
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
            knowledgeName={knowledgeName}
            onSubmit={handleSubmit}
        />
    );
};

export default Knowledge;
