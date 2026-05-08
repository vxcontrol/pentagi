import { createContext, type ReactNode, useCallback, useContext, useMemo } from 'react';
import { toast } from 'sonner';

import type { CreateKnowledgeDocumentInput, UpdateKnowledgeDocumentInput } from '@/graphql/types';

import {
    KnowledgeAnswerType,
    KnowledgeDocType,
    KnowledgeGuideType,
    useCreateKnowledgeDocumentMutation,
    useDeleteKnowledgeDocumentMutation,
    useKnowledgeDocumentCreatedSubscription,
    useKnowledgeDocumentDeletedSubscription,
    useKnowledgeDocumentsQuery,
    useKnowledgeDocumentUpdatedSubscription,
    useUpdateKnowledgeDocumentMutation,
} from '@/graphql/types';
import { Log } from '@/lib/log';
import { useUser } from '@/providers/user-provider';

export interface Knowledge {
    answerType?: KnowledgeAnswerType | null;
    codeLang?: null | string;
    content: string;
    description?: null | string;
    docType: KnowledgeDocType;
    flowId?: null | string;
    guideType?: KnowledgeGuideType | null;
    id: string;
    manual: boolean;
    partSize: number;
    question: string;
    subtaskId?: null | string;
    taskId?: null | string;
    totalSize: number;
    userId: string;
}

interface KnowledgesContextValue {
    createKnowledge: (input: CreateKnowledgeDocumentInput) => Promise<Knowledge | undefined>;
    deleteKnowledge: (id: string) => Promise<void>;
    getKnowledge: (id: string) => Knowledge | undefined;
    isLoading: boolean;
    knowledges: Knowledge[];
    updateKnowledge: (id: string, input: UpdateKnowledgeDocumentInput) => Promise<Knowledge | undefined>;
}

interface KnowledgesProviderProps {
    children: ReactNode;
}

const KnowledgesContext = createContext<KnowledgesContextValue | undefined>(undefined);

export const KnowledgesProvider = ({ children }: KnowledgesProviderProps) => {
    const { authInfo, isAuthenticated } = useUser();

    const shouldFetch = Boolean(authInfo && authInfo.type !== 'guest' && isAuthenticated());

    const { data, loading: isLoading } = useKnowledgeDocumentsQuery({
        fetchPolicy: 'cache-and-network',
        skip: !shouldFetch,
        variables: { withContent: true },
    });

    const [createKnowledgeMutation] = useCreateKnowledgeDocumentMutation();
    const [updateKnowledgeMutation] = useUpdateKnowledgeDocumentMutation();
    const [deleteKnowledgeMutation] = useDeleteKnowledgeDocumentMutation();

    useKnowledgeDocumentCreatedSubscription({ skip: !shouldFetch });
    useKnowledgeDocumentUpdatedSubscription({ skip: !shouldFetch });
    useKnowledgeDocumentDeletedSubscription({ skip: !shouldFetch });

    const knowledges = useMemo<Knowledge[]>(() => {
        const raw = data?.knowledgeDocuments ?? [];

        return raw.map((d) => ({
            answerType: d.answerType ?? null,
            codeLang: d.codeLang ?? null,
            content: d.content,
            description: d.description ?? null,
            docType: d.docType,
            flowId: d.flowId ?? null,
            guideType: d.guideType ?? null,
            id: d.id,
            manual: d.manual,
            partSize: d.partSize,
            question: d.question,
            subtaskId: d.subtaskId ?? null,
            taskId: d.taskId ?? null,
            totalSize: d.totalSize,
            userId: d.userId,
        }));
    }, [data?.knowledgeDocuments]);

    const getKnowledge = useCallback(
        (id: string): Knowledge | undefined => knowledges.find((k) => k.id === id),
        [knowledges],
    );

    const createKnowledge = useCallback(
        async (input: CreateKnowledgeDocumentInput) => {
            try {
                const { data: result } = await createKnowledgeMutation({ variables: { input } });

                return result?.createKnowledgeDocument as Knowledge | undefined;
            } catch (error) {
                const errorMessage = error instanceof Error ? error.message : 'Failed to create knowledge document';
                toast.error('Failed to create knowledge document', { description: errorMessage });
                Log.error('Error creating knowledge document:', error);
                throw error;
            }
        },
        [createKnowledgeMutation],
    );

    const updateKnowledge = useCallback(
        async (id: string, input: UpdateKnowledgeDocumentInput) => {
            try {
                const { data: result } = await updateKnowledgeMutation({ variables: { id, input } });

                return result?.updateKnowledgeDocument as Knowledge | undefined;
            } catch (error) {
                const errorMessage = error instanceof Error ? error.message : 'Failed to update knowledge document';
                toast.error('Failed to update knowledge document', { description: errorMessage });
                Log.error('Error updating knowledge document:', error);
                throw error;
            }
        },
        [updateKnowledgeMutation],
    );

    const deleteKnowledge = useCallback(
        async (id: string) => {
            try {
                await deleteKnowledgeMutation({ variables: { id } });
            } catch (error) {
                const errorMessage = error instanceof Error ? error.message : 'Failed to delete knowledge document';
                toast.error('Failed to delete knowledge document', { description: errorMessage });
                Log.error('Error deleting knowledge document:', error);
                throw error;
            }
        },
        [deleteKnowledgeMutation],
    );

    const value = useMemo<KnowledgesContextValue>(
        () => ({
            createKnowledge,
            deleteKnowledge,
            getKnowledge,
            isLoading,
            knowledges,
            updateKnowledge,
        }),
        [createKnowledge, deleteKnowledge, getKnowledge, isLoading, knowledges, updateKnowledge],
    );

    return <KnowledgesContext.Provider value={value}>{children}</KnowledgesContext.Provider>;
};

export const useKnowledges = () => {
    const context = useContext(KnowledgesContext);

    if (context === undefined) {
        throw new Error('useKnowledges must be used within KnowledgesProvider');
    }

    return context;
};
