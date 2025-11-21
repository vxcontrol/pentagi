import { createContext, useContext, useMemo } from 'react';
import { useParams } from 'react-router-dom';

import type { FlowQuery } from '@/graphql/types';

import { useFlowQuery } from '@/graphql/types';

interface FlowContextValue {
    flowData: FlowQuery | undefined;
    flowId: null | string;
    isLoading: boolean;
}

const FlowContext = createContext<FlowContextValue | undefined>(undefined);

interface FlowProviderProps {
    children: React.ReactNode;
}

export const FlowProvider = ({ children }: FlowProviderProps) => {
    const { flowId } = useParams();

    const { data: flowData, loading: isLoading } = useFlowQuery({
        errorPolicy: 'all',
        fetchPolicy: 'cache-and-network',
        nextFetchPolicy: 'cache-first',
        notifyOnNetworkStatusChange: false,
        skip: !flowId,
        variables: { id: flowId ?? '' },
    });

    const value = useMemo(
        () => ({
            flowData,
            flowId: flowId ?? null,
            isLoading,
        }),
        [flowId, flowData, isLoading],
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
