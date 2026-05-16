import { createContext, type ReactNode, useCallback, useContext, useMemo } from 'react';
import { toast } from 'sonner';

import type {
    CreateKnowledgeDocumentInput,
    KnowledgeDocumentFragmentFragment,
    UpdateKnowledgeDocumentInput,
} from '@/graphql/types';

import {
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

// The provider operates directly on the GraphQL fragment. Previously we kept a
// hand-rolled `Knowledge` shape that mirrored the fragment field-by-field; that
// duplication forced a manual mapping step and drifted from the schema. The
// alias keeps the public surface (`Knowledge`) for callers while making it
// obvious there is no extra translation layer.
export type Knowledge = KnowledgeDocumentFragmentFragment;

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

    // Override the client's default `nextFetchPolicy: 'cache-first'`: since
    // subscriptions are scoped to this provider, the cache can drift while the
    // user is on other pages (AI agents write documents during flow runs).
    // Re-mounting the provider on return to /knowledges should refresh.
    const { data, loading: isLoading } = useKnowledgeDocumentsQuery({
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-and-network',
        skip: !shouldFetch,
        variables: { withContent: true },
    });

    const [createKnowledgeMutation] = useCreateKnowledgeDocumentMutation();
    const [updateKnowledgeMutation] = useUpdateKnowledgeDocumentMutation();
    const [deleteKnowledgeMutation] = useDeleteKnowledgeDocumentMutation();

    useKnowledgeDocumentCreatedSubscription({ skip: !shouldFetch });
    useKnowledgeDocumentUpdatedSubscription({ skip: !shouldFetch });
    useKnowledgeDocumentDeletedSubscription({ skip: !shouldFetch });

    const knowledges = useMemo<Knowledge[]>(() => data?.knowledgeDocuments ?? [], [data?.knowledgeDocuments]);

    const getKnowledge = useCallback(
        (id: string): Knowledge | undefined => knowledges.find((k) => k.id === id),
        [knowledges],
    );

    const createKnowledge = useCallback(
        async (input: CreateKnowledgeDocumentInput) => {
            try {
                const { data: result } = await createKnowledgeMutation({ variables: { input } });

                return result?.createKnowledgeDocument;
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

                return result?.updateKnowledgeDocument;
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
