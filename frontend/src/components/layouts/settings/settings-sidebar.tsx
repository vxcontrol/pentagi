import type { ReactNode } from 'react';

import { ArrowLeft, FileText, Key, Plug, Settings as SettingsIcon, User, Users } from 'lucide-react';
import { useState } from 'react';
import { NavLink, useLocation } from 'react-router-dom';

import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarGroup,
    SidebarGroupContent,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
} from '@/components/ui/sidebar';
import { routes } from '@/lib/routes';
import { getSafeReturnUrl } from '@/lib/utils/auth';
import { useUser } from '@/providers/user-provider';

interface MenuItem {
    icon?: ReactNode;
    id: string;
    path: string;
    // Privilege the account must hold for the item to be shown at all.
    privilege?: string;
    title: string;
}

interface SettingsSidebarMenuItemProps {
    item: MenuItem;
}

const menuItems: readonly MenuItem[] = [
    {
        icon: <User className="size-4" />,
        id: 'account',
        path: routes.settings.account,
        title: 'Account',
    },
    {
        icon: <Plug className="size-4" />,
        id: 'providers',
        path: routes.settings.providers,
        title: 'Providers',
    },
    {
        icon: <FileText className="size-4" />,
        id: 'prompts',
        path: routes.settings.prompts,
        title: 'Prompts',
    },
    {
        icon: <Key className="size-4" />,
        id: 'api-tokens',
        path: routes.settings.apiTokens,
        title: 'API Tokens',
    },
    {
        icon: <Users className="size-4" />,
        id: 'users',
        path: routes.settings.users,
        privilege: 'users.view',
        title: 'Users',
    },
] as const;

export function SettingsSidebar() {
    const location = useLocation();
    const { authInfo } = useUser();
    const privileges = authInfo?.privileges ?? [];
    const visibleItems = menuItems.filter((item) => !item.privilege || privileges.includes(item.privilege));
    const [returnUrl] = useState(() =>
        getSafeReturnUrl((location.state as null | { from?: string })?.from ?? null, routes.flows),
    );

    return (
        <Sidebar collapsible="icon">
            <SidebarHeader>
                <SidebarMenu>
                    <SidebarMenuItem className="flex items-center gap-2">
                        <div className="flex aspect-square size-8 items-center justify-center">
                            <SettingsIcon className="size-6" />
                        </div>
                        <div className="grid flex-1 text-left leading-tight">
                            <span className="truncate font-semibold">Settings</span>
                        </div>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarHeader>
            <SidebarContent>
                <SidebarGroup>
                    <SidebarGroupContent>
                        <SidebarMenu>
                            {visibleItems.map((item) => (
                                <SettingsSidebarMenuItem
                                    item={item}
                                    key={item.id}
                                />
                            ))}
                        </SidebarMenu>
                    </SidebarGroupContent>
                </SidebarGroup>
            </SidebarContent>
            <SidebarFooter>
                <SidebarMenuButton asChild>
                    <NavLink to={returnUrl}>
                        <ArrowLeft />
                        Back to App
                    </NavLink>
                </SidebarMenuButton>
            </SidebarFooter>
        </Sidebar>
    );
}

function SettingsSidebarMenuItem({ item }: SettingsSidebarMenuItemProps) {
    const location = useLocation();
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
}
