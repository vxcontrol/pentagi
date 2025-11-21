import { createContext, useCallback, useContext, useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';

import type { AssistantFragmentFragment, AssistantLogFragmentFragment, FlowQuery } from '@/graphql/types';

import { useAssistantLogsQuery, useAssistantsQuery, useFlowQuery } from '@/graphql/types';

interface FlowContextValue {
    assistantLogs: Array<AssistantLogFragmentFragment>;
    assistants: Array<AssistantFragmentFragment>;
    flowData: FlowQuery | undefined;
    flowId: null | string;
    handleInitiateAssistantCreation: () => void;
    handleSelectAssistant: (assistantId: null | string) => void;
    isAssistantsLoading: boolean;
    isLoading: boolean;
    refetchAssistantLogs: () => void;
    selectedAssistantId: null | string;
}

const FlowContext = createContext<FlowContextValue | undefined>(undefined);

interface FlowProviderProps {
    children: React.ReactNode;
}

export const FlowProvider = ({ children }: FlowProviderProps) => {
    const { flowId } = useParams();

    const [selectedAssistantIds, setSelectedAssistantIds] = useState<Record<string, null | string>>({});

    const { data: flowData, loading: isLoading } = useFlowQuery({
        errorPolicy: 'all',
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        notifyOnNetworkStatusChange: false,
        skip: !flowId,
        variables: { id: flowId ?? '' },
    });

    const { data: assistantsData, loading: isAssistantsLoading } = useAssistantsQuery({
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        skip: !flowId,
        variables: { flowId: flowId ?? '' },
    });

    const assistants = useMemo(() => assistantsData?.assistants ?? [], [assistantsData?.assistants]);

    const selectedAssistantId = useMemo(() => {
        if (!flowId) {
            return null;
        }

        const explicitSelection = selectedAssistantIds[flowId];

        // If there's an explicit selection (including null for "no selection")
        if (explicitSelection !== undefined) {
            // If explicitly set to null, return null
            if (explicitSelection === null) {
                return null;
            }

            // If the selected assistant still exists in the list, return it
            if (assistants.some((assistant) => assistant.id === explicitSelection)) {
                return explicitSelection;
            }
        }

        // Otherwise, auto-select the first assistant
        return assistants?.[0]?.id ?? null;
    }, [flowId, selectedAssistantIds, assistants]);

    const { data: assistantLogsData, refetch: refetchAssistantLogs } = useAssistantLogsQuery({
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        skip: !flowId || !selectedAssistantId || selectedAssistantId === '',
        variables: { assistantId: selectedAssistantId ?? '', flowId: flowId ?? '' },
    });

    const handleSelectAssistant = useCallback(
        (assistantId: null | string) => {
            if (!flowId) {
                return;
            }

            setSelectedAssistantIds((prev) => ({
                ...prev,
                [flowId]: assistantId,
            }));
        },
        [flowId],
    );

    const handleInitiateAssistantCreation = useCallback(() => {
        if (!flowId) {
            return;
        }

        handleSelectAssistant(null);
    }, [flowId, handleSelectAssistant]);

    const value = useMemo(
        () => ({
            assistantLogs: assistantLogsData?.assistantLogs ?? [],
            assistants,
            flowData,
            flowId: flowId ?? null,
            handleInitiateAssistantCreation,
            handleSelectAssistant,
            isAssistantsLoading,
            isLoading,
            refetchAssistantLogs,
            selectedAssistantId,
        }),
        [
            assistantLogsData?.assistantLogs,
            assistants,
            flowData,
            flowId,
            handleInitiateAssistantCreation,
            handleSelectAssistant,
            isAssistantsLoading,
            isLoading,
            refetchAssistantLogs,
            selectedAssistantId,
        ],
    );

    return <FlowContext.Provider value={value}>{children}</FlowContext.Provider>;
};

export const useFlow = () => {
    const context = useContext(FlowContext);

    if (context === undefined) {
        throw new Error('useFlow must be used within FlowProvider');
    }

    return context;
};
