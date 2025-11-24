import { Outlet } from 'react-router-dom';

import MainSidebar from '@/components/layouts/MainSidebar';
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar';

const MainLayout = () => {
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
