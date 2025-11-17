import { Outlet, useParams } from 'react-router-dom';

import MainSidebar from '@/components/layouts/MainSidebar';
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar';
import { useFlowCreatedSubscription, useFlowDeletedSubscription, useFlowUpdatedSubscription } from '@/graphql/types';

const MainLayout = () => {
    const { flowId } = useParams();

    // Use flowId directly from URL params
    const selectedFlowId = flowId ?? null;

    // Global flow subscriptions - always active regardless of selected flow
    useFlowCreatedSubscription();
    useFlowDeletedSubscription();
    useFlowUpdatedSubscription();

    return (
        <SidebarProvider>
            <MainSidebar />
            <SidebarInset>
                <Outlet
                    context={{
                        selectedFlowId,
                    }}
                />
            </SidebarInset>
        </SidebarProvider>
    );
};

export default MainLayout;
