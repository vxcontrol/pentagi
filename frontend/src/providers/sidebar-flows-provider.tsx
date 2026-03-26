import { createContext, type ReactNode, useContext, useMemo } from 'react';

import type { FlowFragmentFragment } from '@/graphql/types';

import { useFlowsQuery } from '@/graphql/types';

export type Flow = FlowFragmentFragment;

interface SidebarFlowsContextValue {
    flows: Array<Flow>;
}

const SidebarFlowsContext = createContext<SidebarFlowsContextValue | undefined>(undefined);

interface SidebarFlowsProviderProps {
    children: ReactNode;
}

export const SidebarFlowsProvider = ({ children }: SidebarFlowsProviderProps) => {
    // Single query for sidebar flows with cache-first policy
    // Subscriptions are handled by FlowsProvider in FlowsLayout
    const { data: flowsData } = useFlowsQuery({
        fetchPolicy: 'cache-first',
        nextFetchPolicy: 'cache-first',
    });

    const flows = useMemo(() => flowsData?.flows ?? [], [flowsData?.flows]);

    const value = useMemo(
        () => ({
            flows,
        }),
        [flows],
    );

    return <SidebarFlowsContext.Provider value={value}>{children}</SidebarFlowsContext.Provider>;
};

export const useSidebarFlows = () => {
    const context = useContext(SidebarFlowsContext);

    if (context === undefined) {
        throw new Error('useSidebarFlows must be used within SidebarFlowsProvider');
    }

    return context;
};
