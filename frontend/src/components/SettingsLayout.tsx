import { ArrowLeft, FileText, Plug, Settings as SettingsIcon } from 'lucide-react';
import { useMemo } from 'react';
import { NavLink, Outlet, useMatch } from 'react-router-dom';

import { Separator } from '@/components/ui/separator';
import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarGroup,
    SidebarGroupContent,
    SidebarGroupLabel,
    SidebarHeader,
    SidebarInset,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
    SidebarProvider,
    SidebarTrigger,
} from '@/components/ui/sidebar';

// Types
export interface MenuItem {
    id: string;
    title: string;
    path: string;
    icon?: React.ReactNode;
    isActive?: boolean;
}

interface SettingsSidebarMenuItemProps {
    item: MenuItem;
}

// Settings menu items definition
const menuItems: readonly MenuItem[] = [
    {
        id: 'providers',
        title: 'Providers',
        path: '/settings/providers',
        icon: <Plug className="size-4" />,
    },
    {
        id: 'prompts',
        title: 'Prompts',
        path: '/settings/prompts',
        icon: <FileText className="size-4" />,
    },
] as const;

// Individual menu item component to properly use hooks
const SettingsSidebarMenuItem = ({ item }: SettingsSidebarMenuItemProps) => {
    const match = useMatch(item.path);
    const isActive = !!match;

    return (
        <SidebarMenuItem>
            <SidebarMenuButton
                asChild
                isActive={isActive}
            >
                <NavLink to={item.path}>
                    {item.icon}
                    {item.title}
                </NavLink>
            </SidebarMenuButton>
        </SidebarMenuItem>
    );
};

// Settings header component
const SettingsHeader = () => {
    // Call useMatch for each menu item at the top level
    const items = menuItems.map((item) => ({
        ...item,
        isActive: !!useMatch(item.path),
    }));

    // Memoize title calculation for better performance
    const title = useMemo(() => {
        const item = items.find((item) => item.isActive) ?? menuItems[0];
        return item?.title ?? 'Settings';
    }, [items]);

    return (
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator
                orientation="vertical"
                className="mr-2 h-4"
            />
            <h1 className="text-lg font-semibold">{title}</h1>
        </header>
    );
};

// Settings sidebar component
const SettingsSidebar = () => {
    return (
        <Sidebar>
            <SidebarHeader>
                <div className="flex items-center gap-2">
                    <SettingsIcon className="size-6" />
                    <span className="font-semibold">Settings</span>
                </div>
            </SidebarHeader>
            <SidebarContent>
                <SidebarGroup>
                    <SidebarGroupLabel>Settings</SidebarGroupLabel>
                    <SidebarGroupContent>
                        <SidebarMenu>
                            {menuItems.map((item) => (
                                <SettingsSidebarMenuItem
                                    key={item.id}
                                    item={item}
                                />
                            ))}
                        </SidebarMenu>
                    </SidebarGroupContent>
                </SidebarGroup>
            </SidebarContent>
            <SidebarFooter>
                <SidebarMenuButton asChild>
                    <NavLink to="/chat/new">
                        <ArrowLeft className="size-4" />
                        Back to Chat
                    </NavLink>
                </SidebarMenuButton>
            </SidebarFooter>
        </Sidebar>
    );
};

// Settings layout component
const SettingsLayout = () => {
    return (
        <SidebarProvider>
            <div className="flex h-screen w-full">
                <SettingsSidebar />
                <SidebarInset className="flex flex-1 flex-col">
                    <SettingsHeader />
                    {/* Content area for nested routes */}
                    <main className="flex-1 overflow-auto p-4">
                        <Outlet />
                    </main>
                </SidebarInset>
            </div>
        </SidebarProvider>
    );
};

export default SettingsLayout;
