import { useCallback } from 'react';
import { toast } from 'sonner';

import type { FlowFormValues } from '@/features/flows/FlowForm';
import type { Provider } from '@/models/Provider';

import {
    StatusType,
    useCallAssistantMutation,
    useCreateAssistantMutation,
    useDeleteAssistantMutation,
    usePutUserInputMutation,
    useStopAssistantMutation,
    useStopFlowMutation,
} from '@/graphql/types';
import { Log } from '@/lib/log';
import { useFlow } from '@/providers/FlowProvider';

interface UseFlowMutationsOptions {
    handleSelectAssistant: (assistantId: null | string) => void;
    refetchAssistantLogs?: () => void;
    selectedAssistantId: null | string;
    selectedProvider: null | Provider;
}

export const useFlowMutations = ({
    handleSelectAssistant,
    refetchAssistantLogs,
    selectedAssistantId,
    selectedProvider,
}: UseFlowMutationsOptions) => {
    const { flowData, flowId } = useFlow();

    const [putUserInput] = usePutUserInputMutation();
    const [stopFlow] = useStopFlowMutation();
    const [createAssistant] = useCreateAssistantMutation();
    const [callAssistant] = useCallAssistantMutation();
    const [stopAssistant] = useStopAssistantMutation();
    const [deleteAssistant] = useDeleteAssistantMutation();

    const handleSubmitAutomationMessage = useCallback(
        async (values: FlowFormValues) => {
            if (!flowId || flowData?.flow?.status === StatusType.Finished) {
                return;
            }

            const { message } = values;

            try {
                await putUserInput({
                    variables: {
                        flowId: flowId ?? '',
                        input: message,
                    },
                });
            } catch (error) {
                const description =
                    error instanceof Error ? error.message : 'An error occurred while submitting message';
                toast.error('Failed to submit message', {
                    description,
                });
                Log.error('Error submitting message:', error);
            }
        },
        [flowId, flowData?.flow?.status, putUserInput],
    );

    const handleStopAutomationFlow = useCallback(
        async (flowIdToStop: string) => {
            try {
                await stopFlow({
                    variables: {
                        flowId: flowIdToStop,
                    },
                });
            } catch (error) {
                const description = error instanceof Error ? error.message : 'An error occurred while stopping flow';
                toast.error('Failed to stop flow', {
                    description,
                });
                Log.error('Error stopping flow:', error);
            }
        },
        [stopFlow],
    );

    const handleCreateAssistant = useCallback(
        async (message: string, useAgents: boolean) => {
            if (!message.trim() || !selectedProvider || !flowId) {
                return;
            }

            try {
                const { data } = await createAssistant({
                    variables: {
                        flowId,
                        input: message.trim(),
                        modelProvider: selectedProvider.name,
                        useAgents,
                    },
                });

                if (data?.createAssistant) {
                    const { assistant } = data.createAssistant;

                    if (assistant?.id) {
                        handleSelectAssistant(assistant.id);
                    }
                }
            } catch (error) {
                const description =
                    error instanceof Error ? error.message : 'An error occurred while creating assistant';
                toast.error('Failed to create assistant', {
                    description,
                });
                Log.error('Error creating assistant:', error);
            }
        },
        [flowId, selectedProvider, createAssistant, handleSelectAssistant],
    );

    const handleCallAssistant = useCallback(
        async (assistantId: string, message: string, useAgents: boolean) => {
            if (!flowId || !assistantId || !message.trim()) {
                return;
            }

            try {
                await callAssistant({
                    variables: {
                        assistantId,
                        flowId,
                        input: message.trim(),
                        useAgents,
                    },
                });

                if (selectedAssistantId && refetchAssistantLogs) {
                    refetchAssistantLogs();
                }
            } catch (error) {
                const description =
                    error instanceof Error ? error.message : 'An error occurred while calling assistant';
                toast.error('Failed to call assistant', {
                    description,
                });
                Log.error('Error calling assistant:', error);
            }
        },
        [flowId, callAssistant, selectedAssistantId, refetchAssistantLogs],
    );

    const handleStopAssistant = useCallback(
        async (assistantId: string) => {
            if (!flowId || !assistantId) {
                return;
            }

            try {
                await stopAssistant({
                    variables: {
                        assistantId,
                        flowId,
                    },
                });

                if (selectedAssistantId && refetchAssistantLogs) {
                    refetchAssistantLogs();
                }
            } catch (error) {
                const description =
                    error instanceof Error ? error.message : 'An error occurred while stopping assistant';
                toast.error('Failed to stop assistant', {
                    description,
                });
                Log.error('Error stopping assistant:', error);
            }
        },
        [flowId, stopAssistant, selectedAssistantId, refetchAssistantLogs],
    );

    const handleDeleteAssistant = useCallback(
        async (assistantId: string) => {
            if (!flowId || !assistantId) {
                return;
            }

            try {
                const wasSelected = selectedAssistantId === assistantId;

                await deleteAssistant({
                    update: (cache) => {
                        cache.evict({ id: `Assistant:${assistantId}` });
                        cache.gc();
                    },
                    variables: {
                        assistantId,
                        flowId,
                    },
                });

                if (wasSelected) {
                    handleSelectAssistant(null);
                }
            } catch (error) {
                const description =
                    error instanceof Error ? error.message : 'An error occurred while deleting assistant';
                toast.error('Failed to delete assistant', {
                    description,
                });
                Log.error('Error deleting assistant:', error);
            }
        },
        [flowId, selectedAssistantId, deleteAssistant, handleSelectAssistant],
    );

    return {
        handleCallAssistant,
        handleCreateAssistant,
        handleDeleteAssistant,
        handleStopAssistant,
        handleStopAutomationFlow,
        handleSubmitAutomationMessage,
    };
};
