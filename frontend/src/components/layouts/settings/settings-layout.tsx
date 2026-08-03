import { Outlet } from 'react-router-dom';

import { SettingsSidebar } from '@/components/layouts/settings/settings-sidebar';
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar';

function SettingsLayout() {
    return (
        <SidebarProvider>
            <SettingsSidebar />
            <SidebarInset>
                <Outlet />
            </SidebarInset>
        </SidebarProvider>
    );
}

export default SettingsLayout;
