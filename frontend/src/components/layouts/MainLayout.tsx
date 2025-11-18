import { Outlet } from 'react-router-dom';

import MainSidebar from '@/components/layouts/MainSidebar';
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar';
import { useFlowCreatedSubscription, useFlowDeletedSubscription, useFlowUpdatedSubscription } from '@/graphql/types';

const MainLayout = () => {
    // Global flow subscriptions - always active regardless of selected flow
    useFlowCreatedSubscription();
    useFlowDeletedSubscription();
    useFlowUpdatedSubscription();

    return (
        <SidebarProvider>
            <MainSidebar />
            <SidebarInset>
                <Outlet />
            </SidebarInset>
        </SidebarProvider>
    );
};

export default MainLayout;
