import { useMutation, useQuery, useSubscription } from '@apollo/client/react';
import { createContext, type ReactNode, useCallback, useContext, useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';
import { useDebounce } from 'use-debounce';

import type {
    CreateKnowledgeDocumentInput,
    KnowledgeDocumentFragmentFragment,
    UpdateKnowledgeDocumentInput,
} from '@/graphql/types';

import {
    CreateKnowledgeDocumentDocument,
    DeleteKnowledgeDocumentDocument,
    KnowledgeDocumentCreatedDocument,
    KnowledgeDocumentDeletedDocument,
    KnowledgeDocumentsDocument,
    KnowledgeDocumentUpdatedDocument,
    RenameKnowledgeDocumentDocument,
    SearchKnowledgeDocument,
    UpdateKnowledgeDocumentDocument,
} from '@/graphql/types';
import { useLatestRef } from '@/hooks/use-latest-ref';
import { Log } from '@/lib/log';
import { URL_PARAMS } from '@/lib/url-params';
import { useUser } from '@/providers/user-provider';

export type Knowledge = KnowledgeDocumentFragmentFragment;

interface KnowledgesContextValue {
    createKnowledge: (input: CreateKnowledgeDocumentInput) => Promise<Knowledge | undefined>;
    deleteKnowledge: (id: string) => Promise<void>;
    getKnowledge: (id: string) => Knowledge | undefined;
    isLoading: boolean;
    knowledges: Knowledge[];
    renameKnowledge: (id: string, question: string) => Promise<Knowledge | undefined>;
    updateKnowledge: (id: string, input: UpdateKnowledgeDocumentInput) => Promise<Knowledge | undefined>;
}

interface KnowledgesProviderProps {
    children: ReactNode;
}

const KnowledgesContext = createContext<KnowledgesContextValue | undefined>(undefined);

// Also bounds how many neighbours Prev/Next inside `<DetailNavigation>` can step
// through, since it walks this same result array.
const SEARCH_RESULT_LIMIT = 100;

// Debounce `?qs=`: each keystroke otherwise spawns an embedding + vector-search round-trip.
const SEARCH_DEBOUNCE_MS = 400;

export function KnowledgesProvider({ children }: KnowledgesProviderProps) {
    const { authInfo, isAuthenticated } = useUser();

    const shouldFetch = Boolean(authInfo && authInfo.type !== 'guest' && isAuthenticated());

    const [searchParams] = useSearchParams();
    const rawSemanticQuery = searchParams.get(URL_PARAMS.SEARCH) ?? '';
    const [debouncedSemanticQueryRaw] = useDebounce(rawSemanticQuery, SEARCH_DEBOUNCE_MS);
    const debouncedSemanticQuery = debouncedSemanticQueryRaw.trim();
    const inSearchMode = debouncedSemanticQuery.length > 0;

    // Override the client's default `nextFetchPolicy: 'cache-first'`: since
    // subscriptions are scoped to this provider, the cache can drift while the
    // user is on other pages (AI agents write documents during flow runs).
    // Re-mounting the provider on return to /knowledges should refresh.
    //
    // Both queries are always mounted, with `skip` flipping the active one.
    // That keeps Apollo's cache warm for the inactive branch — when the user
    // toggles `?qs=` on/off, the previous result is shown immediately while
    // the network refetches in the background.
    const { data: listData, loading: isListLoading } = useQuery(KnowledgeDocumentsDocument, {
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-and-network',
        skip: !shouldFetch || inSearchMode,
        variables: { withContent: false },
    });

    // `searchKnowledge` ignores `withContent` — the backend always returns the full
    // chunk text plus a relevance score we currently drop.
    const { data: searchData, loading: isSearchLoading } = useQuery(SearchKnowledgeDocument, {
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-and-network',
        skip: !shouldFetch || !inSearchMode,
        variables: { filter: null, limit: SEARCH_RESULT_LIMIT, query: debouncedSemanticQuery },
    });

    const [createKnowledgeMutation] = useMutation(CreateKnowledgeDocumentDocument);
    const [updateKnowledgeMutation] = useMutation(UpdateKnowledgeDocumentDocument);
    const [renameKnowledgeMutation] = useMutation(RenameKnowledgeDocumentDocument);
    const [deleteKnowledgeMutation] = useMutation(DeleteKnowledgeDocumentDocument);

    // The Apollo subscription link (`createSubscriptionCacheLink` in
    // `lib/apollo.ts`) auto-merges each event into the `knowledgeDocuments`
    // root field — so list-mode UI stays in sync without help here. The
    // `searchKnowledge` field has no such mapping (relevance ordering can't
    // be reconstructed from a single mutated document), so in search mode we
    // refetch the active query whenever a CRUD event arrives. Read
    // `inSearchMode` through a ref because the subscription `onData`
    // callback is registered once and fires asynchronously from WebSocket
    // messages — capturing the boolean directly would freeze it at mount.
    const inSearchModeRef = useLatestRef(inSearchMode);
    const refetchSearchOnEvent = useCallback(
        ({ client }: { client: { refetchQueries: (options: { include: string[] }) => unknown } }) => {
            if (!inSearchModeRef.current) {
                return;
            }

            client.refetchQueries({ include: ['searchKnowledge'] });
        },
        [inSearchModeRef],
    );

    useSubscription(KnowledgeDocumentCreatedDocument, { onData: refetchSearchOnEvent, skip: !shouldFetch });
    useSubscription(KnowledgeDocumentUpdatedDocument, { onData: refetchSearchOnEvent, skip: !shouldFetch });
    useSubscription(KnowledgeDocumentDeletedDocument, { onData: refetchSearchOnEvent, skip: !shouldFetch });

    const knowledges = useMemo<Knowledge[]>(() => {
        if (inSearchMode) {
            return searchData?.searchKnowledge.map((entry) => entry.document) ?? [];
        }

        return listData?.knowledgeDocuments ?? [];
    }, [inSearchMode, listData?.knowledgeDocuments, searchData?.searchKnowledge]);

    const isLoading = inSearchMode ? isSearchLoading : isListLoading;

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

    const renameKnowledge = useCallback(
        async (id: string, question: string) => {
            try {
                const { data: result } = await renameKnowledgeMutation({ variables: { id, question } });

                return result?.renameKnowledgeDocument;
            } catch (error) {
                const errorMessage = error instanceof Error ? error.message : 'Failed to rename knowledge document';
                toast.error('Failed to rename knowledge document', { description: errorMessage });
                Log.error('Error renaming knowledge document:', error);
                throw error;
            }
        },
        [renameKnowledgeMutation],
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
            renameKnowledge,
            updateKnowledge,
        }),
        [createKnowledge, deleteKnowledge, getKnowledge, isLoading, knowledges, renameKnowledge, updateKnowledge],
    );

    return <KnowledgesContext.Provider value={value}>{children}</KnowledgesContext.Provider>;
}

export function useKnowledges() {
    const context = useContext(KnowledgesContext);

    if (context === undefined) {
        throw new Error('useKnowledges must be used within KnowledgesProvider');
    }

    return context;
}
