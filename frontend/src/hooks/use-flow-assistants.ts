import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { useAssistantLogsQuery, useAssistantsQuery } from '@/graphql/types';
import { useFlow } from '@/providers/FlowProvider';

export const useFlowAssistants = () => {
    const { flowId } = useFlow();

    const assistantCreationTimeoutRef = useRef<NodeJS.Timeout | null>(null);

    const [selectedAssistantIds, setSelectedAssistantIds] = useState<Record<string, null | string>>({});

    const { data: assistantsData } = useAssistantsQuery({
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        skip: !flowId,
        variables: { flowId: flowId ?? '' },
    });

    const selectedAssistantId = useMemo(() => {
        if (!flowId) {
            return null;
        }

        if (flowId in selectedAssistantIds) {
            const explicitSelection = selectedAssistantIds[flowId];

            if (explicitSelection === null) {
                return null;
            }

            if (assistantsData?.assistants?.some((assistant) => assistant.id === explicitSelection)) {
                return explicitSelection;
            }
        }

        const firstAssistantId = assistantsData?.assistants?.[0]?.id;

        return firstAssistantId || null;
    }, [flowId, selectedAssistantIds, assistantsData?.assistants]);

    const { data: assistantLogsData, refetch: refetchAssistantLogs } = useAssistantLogsQuery({
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        skip: !flowId || !selectedAssistantId || selectedAssistantId === '',
        variables: { assistantId: selectedAssistantId ?? '', flowId: flowId ?? '' },
    });

    useEffect(() => {
        return () => {
            if (assistantCreationTimeoutRef.current) {
                clearTimeout(assistantCreationTimeoutRef.current);
                assistantCreationTimeoutRef.current = null;
            }
        };
    }, []);

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

    return {
        assistantCreationTimeoutRef,
        assistantLogs: assistantLogsData?.assistantLogs ?? [],
        assistants: assistantsData?.assistants ?? [],
        handleInitiateAssistantCreation,
        handleSelectAssistant,
        refetchAssistantLogs,
        selectedAssistantId,
    };
};
