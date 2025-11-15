import { ArrowLeft, FileText, Plug, Settings as SettingsIcon } from 'lucide-react';
import { useMemo } from 'react';
import { NavLink, Outlet, useLocation, useParams } from 'react-router-dom';

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
    // {
    //     id: 'mcp-servers',
    //     title: 'MCP Servers',
    //     path: '/settings/mcp-servers',
    //     icon: <Server className="size-4" />,
    // },
] as const;

// Individual menu item component to properly use hooks
const SettingsSidebarMenuItem = ({ item }: SettingsSidebarMenuItemProps) => {
    const location = useLocation();
    // Check if current path starts with item path (for nested routes)
    const isActive = location.pathname.startsWith(item.path);

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
    const location = useLocation();
    const params = useParams();

    // Memoize title calculation for better performance
    const title = useMemo(() => {
        const path = location.pathname;

        // Check for specific nested routes
        if (path === '/settings/providers/new') {
            return 'Create Provider';
        }

        if (path.startsWith('/settings/providers/') && params.providerId && params.providerId !== 'new') {
            return 'Edit Provider';
        }

        if (path === '/settings/mcp-servers/new') {
            return 'Create MCP Server';
        }

        if (path.startsWith('/settings/mcp-servers/')) {
            return 'Edit MCP Server';
        }

        if (path === '/settings/prompts/new') {
            return 'Create Prompt';
        }

        if (path.startsWith('/settings/prompts/') && params.promptId && params.promptId !== 'new') {
            return 'Edit Prompt';
        }

        // Find matching main section
        const activeItem = menuItems.find((item) => path.startsWith(item.path));

        return activeItem?.title ?? 'Settings';
    }, [location.pathname, params]);

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
                    <NavLink to="/flows">
                        <ArrowLeft className="size-4" />
                        Back to App
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
            <div className="flex h-screen w-full overflow-hidden">
                <SettingsSidebar />
                <SidebarInset className="flex flex-1 flex-col">
                    <SettingsHeader />
                    {/* Content area for nested routes */}
                    <main className="min-h-0 flex-1 overflow-auto p-4">
                        <Outlet />
                    </main>
                </SidebarInset>
            </div>
        </SidebarProvider>
    );
};

export default SettingsLayout;
