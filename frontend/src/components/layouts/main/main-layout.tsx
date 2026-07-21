import { Outlet } from 'react-router-dom';

import { MainSidebar } from '@/components/layouts/main/main-sidebar';
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar';

function MainLayout() {
    return (
        <SidebarProvider>
            <MainSidebar />
            <SidebarInset>
                <Outlet />
            </SidebarInset>
        </SidebarProvider>
    );
}

export default MainLayout;
