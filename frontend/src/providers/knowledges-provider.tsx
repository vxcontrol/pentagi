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
    useCreateKnowledgeDocumentMutation,
    useDeleteKnowledgeDocumentMutation,
    useKnowledgeDocumentCreatedSubscription,
    useKnowledgeDocumentDeletedSubscription,
    useKnowledgeDocumentsQuery,
    useKnowledgeDocumentUpdatedSubscription,
    useSearchKnowledgeQuery,
    useUpdateKnowledgeDocumentMutation,
} from '@/graphql/types';
import { useLatestRef } from '@/hooks/use-latest-ref';
import { Log } from '@/lib/log';
import { URL_PARAMS } from '@/lib/url-params';
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

// Cap on the server-returned semantic-search result set. Prev/Next inside
// `<DetailNavigation>` walks the same array, so the limit also bounds how
// many neighbours a user can step through after running a search. 100 is
// generous for a top-K relevance list and matches what other list pages
// expect to render without virtualization.
const SEARCH_RESULT_LIMIT = 100;

// Debounce for `?qs=` before we hit the server. Filter typing fires a
// keystroke per character; without a debounce we'd spawn an embedding
// + vector-search round-trip on each one. 400ms is the sweet spot users
// don't perceive as laggy while still collapsing burst typing into one
// network call.
const SEARCH_DEBOUNCE_MS = 400;

export function KnowledgesProvider({ children }: KnowledgesProviderProps) {
    const { authInfo, isAuthenticated } = useUser();

    const shouldFetch = Boolean(authInfo && authInfo.type !== 'guest' && isAuthenticated());

    // `?qs=` is read directly from the URL — there is no in-provider setter.
    // Pages drive it via `useSearchParams` (or the soon-to-arrive semantic
    // search input). Trimming + debouncing happens here so the rest of the
    // provider sees one canonical "is the user actively searching" flag.
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
    const { data: listData, loading: isListLoading } = useKnowledgeDocumentsQuery({
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-and-network',
        skip: !shouldFetch || inSearchMode,
        variables: { withContent: true },
    });

    // `searchKnowledge` does not honour `withContent` — the backend always
    // returns the full chunk text plus a relevance score we currently drop.
    // `filter` is wired as `null` for now; when facet filtering arrives
    // (`?f.docType=…`), the parsed `KnowledgeFilter` slots in here and into
    // the list query above without any other downstream change.
    const { data: searchData, loading: isSearchLoading } = useSearchKnowledgeQuery({
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-and-network',
        skip: !shouldFetch || !inSearchMode,
        variables: { filter: null, limit: SEARCH_RESULT_LIMIT, query: debouncedSemanticQuery },
    });

    const [createKnowledgeMutation] = useCreateKnowledgeDocumentMutation();
    const [updateKnowledgeMutation] = useUpdateKnowledgeDocumentMutation();
    const [deleteKnowledgeMutation] = useDeleteKnowledgeDocumentMutation();

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

    useKnowledgeDocumentCreatedSubscription({ onData: refetchSearchOnEvent, skip: !shouldFetch });
    useKnowledgeDocumentUpdatedSubscription({ onData: refetchSearchOnEvent, skip: !shouldFetch });
    useKnowledgeDocumentDeletedSubscription({ onData: refetchSearchOnEvent, skip: !shouldFetch });

    const knowledges = useMemo<Knowledge[]>(() => {
        if (inSearchMode) {
            // Drop `score` — every consumer (DataTable, DetailNavigation,
            // mutations) already speaks plain `KnowledgeDocumentFragment`.
            // If a future UI wants to render relevance, the score is still
            // reachable via the raw Apollo result by lifting a small helper
            // into context — out of scope for this change.
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
}

export function useKnowledges() {
    const context = useContext(KnowledgesContext);

    if (context === undefined) {
        throw new Error('useKnowledges must be used within KnowledgesProvider');
    }

    return context;
}
