import { Avatar, AvatarFallback } from '@radix-ui/react-avatar';
import {
    ChevronsUpDown,
    Clock,
    GitFork,
    KeyRound,
    LogOut,
    Monitor,
    Moon,
    Plus,
    Settings,
    Settings2,
    Star,
    Sun,
    UserIcon,
} from 'lucide-react';
import { useMemo, useState } from 'react';
import { Link, useLocation, useMatch, useParams } from 'react-router-dom';

import type { Theme } from '@/providers/theme-provider';

import Logo from '@/components/icons/logo';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarGroup,
    SidebarGroupContent,
    SidebarGroupLabel,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuAction,
    SidebarMenuButton,
    SidebarMenuItem,
    SidebarRail,
} from '@/components/ui/sidebar';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { PasswordChangeForm } from '@/features/authentication/password-change-form';
import { useTheme } from '@/hooks/use-theme';
import { useFavorites } from '@/providers/favorites-provider';
import { useSidebarFlows } from '@/providers/sidebar-flows-provider';
import { useUser } from '@/providers/user-provider';

const MainSidebar = () => {
    const [isPasswordModalOpen, setIsPasswordModalOpen] = useState(false);
    const [clickedButtons, setClickedButtons] = useState<Set<string>>(new Set());

    const isSettingsActive = useMatch('/settings/*');
    const { flowId: flowIdParam } = useParams<{ flowId: string }>();
    const location = useLocation();

    // Flows button is active only on /flows list and /flows/new, not on specific flow pages
    const isFlowsActive = useMemo(() => {
        return location.pathname === '/flows' || location.pathname === '/flows/new';
    }, [location.pathname]);

    const { authInfo, logout } = useUser();
    const user = authInfo?.user;
    const { setTheme, theme } = useTheme();
    const { addFavoriteFlow, favoriteFlowIds, removeFavoriteFlow } = useFavorites();
    const { flows } = useSidebarFlows();

    // Convert flowId to number for comparison
    const flowId = useMemo(() => {
        return flowIdParam ? +flowIdParam : null;
    }, [flowIdParam]);

    // Check if we're on a specific flow page (not /flows/new)
    const isOnFlowPage = useMemo(() => {
        return location.pathname.startsWith('/flows/') && flowIdParam && flowIdParam !== 'new';
    }, [location.pathname, flowIdParam]);

    // Get favorite flows (full objects)
    const favoriteFlows = useMemo(() => {
        const filtered = flows
            .filter((flow) => {
                const numericFlowId = typeof flow.id === 'string' ? +flow.id : flow.id;

                return favoriteFlowIds.includes(numericFlowId);
            })
            .sort((a, b) => +b.id - +a.id);

        return filtered;
    }, [flows, favoriteFlowIds]);

    // Get recent flows (5 latest non-favorites, sorted by createdAt desc)
    const recentFlows = useMemo(() => {
        const nonFavoriteFlows = flows.filter((flow) => {
            const numericFlowId = typeof flow.id === 'string' ? +flow.id : flow.id;

            return !favoriteFlowIds.includes(numericFlowId);
        });
        const sortedByDate = [...nonFavoriteFlows].sort(
            (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
        );

        return sortedByDate.slice(0, 5);
    }, [flows, favoriteFlowIds]);

    // Get current flow (if on flow page and not in recent/favorites)
    const currentFlow = useMemo(() => {
        if (!isOnFlowPage || !flowId) {
            return null;
        }

        const isInRecent = recentFlows.some((flow) => +flow.id === flowId);
        const isInFavorites = favoriteFlows.some((flow) => +flow.id === flowId);

        if (isInRecent || isInFavorites) {
            return null;
        }

        const found = flows.find((flow) => +flow.id === flowId) || null;

        return found;
    }, [isOnFlowPage, flowId, flows, recentFlows, favoriteFlows]);

    const handlePasswordChangeSuccess = () => {
        setIsPasswordModalOpen(false);
    };

    return (
        <Sidebar collapsible="icon">
            <SidebarHeader>
                <SidebarMenu>
                    <SidebarMenuItem className="flex items-center gap-2">
                        <div className="flex aspect-square size-8 items-center justify-center">
                            <Logo className="hover:animate-logo-spin size-6" />
                        </div>
                        <div className="grid flex-1 text-left leading-tight">
                            <span className="truncate font-semibold">PentAGI</span>
                        </div>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarHeader>
            <SidebarContent>
                <SidebarGroup className="bg-sidebar sticky top-0 z-10">
                    <SidebarGroupContent>
                        <SidebarMenu>
                            <SidebarMenuItem className="group-data-[state=expanded]:hidden">
                                <SidebarMenuButton asChild>
                                    <Link to="/flows/new">
                                        <Plus />
                                        New Flow
                                    </Link>
                                </SidebarMenuButton>
                            </SidebarMenuItem>
                            <SidebarMenuItem>
                                <SidebarMenuButton
                                    asChild
                                    isActive={!!isFlowsActive}
                                >
                                    <Link to="/flows">
                                        <GitFork />
                                        Flows
                                    </Link>
                                </SidebarMenuButton>
                                <SidebarMenuAction
                                    asChild
                                    className="data-[state=open]:bg-accent rounded-sm"
                                    showOnHover
                                >
                                    <Link to="/flows/new">
                                        <Plus />
                                    </Link>
                                </SidebarMenuAction>
                            </SidebarMenuItem>
                        </SidebarMenu>
                    </SidebarGroupContent>
                </SidebarGroup>

                {currentFlow && (
                    <SidebarGroup>
                        <SidebarGroupContent>
                            <SidebarMenu>
                                <SidebarMenuItem
                                    onMouseLeave={(e) => {
                                        const menuItem = e.currentTarget;
                                        menuItem.querySelectorAll('button, a').forEach((el) => {
                                            if (el instanceof HTMLElement) {
                                                el.blur();
                                            }
                                        });

                                        const key = `current-${currentFlow.id}`;
                                        setClickedButtons((prev) => {
                                            const next = new Set(prev);
                                            next.delete(key);

                                            return next;
                                        });
                                    }}
                                >
                                    <SidebarMenuButton
                                        asChild
                                        isActive={true}
                                    >
                                        <Link to={`/flows/${currentFlow.id}`}>
                                            <span className="-mx-2 w-8 shrink-0 text-center text-xs group-data-[state=expanded]:hidden">
                                                {currentFlow.id}
                                            </span>
                                            <span className="text-muted-foreground bg-background dark:bg-muted -my-0.5 -ml-0.5 h-5 min-w-5 shrink-0 rounded-md px-px py-0.5 text-center text-xs group-data-[state=collapsed]:hidden">
                                                {currentFlow.id}
                                            </span>
                                            <span className="truncate">{currentFlow.title}</span>
                                        </Link>
                                    </SidebarMenuButton>
                                    <SidebarMenuAction
                                        className={`data-[state=open]:bg-accent rounded-sm ${clickedButtons.has(`current-${currentFlow.id}`) ? 'pointer-events-none! opacity-0!' : ''}`}
                                        onClick={(e) => {
                                            e.preventDefault();
                                            e.stopPropagation();

                                            const button = e.currentTarget;
                                            button.blur();

                                            const key = `current-${currentFlow.id}`;
                                            setClickedButtons((prev) => new Set(prev).add(key));
                                            addFavoriteFlow(currentFlow.id);

                                            setTimeout(() => {
                                                setClickedButtons((prev) => {
                                                    const next = new Set(prev);
                                                    next.delete(key);

                                                    return next;
                                                });
                                            }, 600);
                                        }}
                                        showOnHover
                                    >
                                        <Star />
                                    </SidebarMenuAction>
                                </SidebarMenuItem>
                            </SidebarMenu>
                        </SidebarGroupContent>
                    </SidebarGroup>
                )}

                {recentFlows.length > 0 && (
                    <SidebarGroup>
                        <SidebarGroupLabel className="flex items-center gap-2">
                            <Clock />
                            Recent Flows
                        </SidebarGroupLabel>
                        <SidebarGroupContent>
                            <SidebarMenu>
                                {recentFlows.map((flow) => (
                                    <SidebarMenuItem
                                        key={flow.id}
                                        onMouseLeave={(e) => {
                                            const menuItem = e.currentTarget;
                                            menuItem.querySelectorAll('button, a').forEach((el) => {
                                                if (el instanceof HTMLElement) {
                                                    el.blur();
                                                }
                                            });

                                            const key = `recent-${flow.id}`;
                                            setClickedButtons((prev) => {
                                                const next = new Set(prev);
                                                next.delete(key);

                                                return next;
                                            });
                                        }}
                                    >
                                        <SidebarMenuButton
                                            asChild
                                            isActive={flowId === +flow.id}
                                        >
                                            <Link to={`/flows/${flow.id}`}>
                                                <span className="-mx-2 w-8 shrink-0 text-center text-xs group-data-[state=expanded]:hidden">
                                                    {flow.id}
                                                </span>
                                                <span className="text-muted-foreground bg-background dark:bg-muted -my-0.5 -ml-0.5 h-5 min-w-5 shrink-0 rounded-md px-px py-0.5 text-center text-xs group-data-[state=collapsed]:hidden">
                                                    {flow.id}
                                                </span>
                                                <span className="truncate">{flow.title}</span>
                                            </Link>
                                        </SidebarMenuButton>
                                        <SidebarMenuAction
                                            className={`data-[state=open]:bg-accent rounded-sm ${clickedButtons.has(`recent-${flow.id}`) ? 'pointer-events-none! opacity-0!' : ''}`}
                                            onClick={(e) => {
                                                e.preventDefault();
                                                e.stopPropagation();
                                                const button = e.currentTarget;
                                                button.blur();

                                                const key = `recent-${flow.id}`;
                                                setClickedButtons((prev) => new Set(prev).add(key));
                                                addFavoriteFlow(flow.id);

                                                setTimeout(() => {
                                                    setClickedButtons((prev) => {
                                                        const next = new Set(prev);
                                                        next.delete(key);

                                                        return next;
                                                    });
                                                }, 600);
                                            }}
                                            showOnHover
                                        >
                                            <Star />
                                        </SidebarMenuAction>
                                    </SidebarMenuItem>
                                ))}
                            </SidebarMenu>
                        </SidebarGroupContent>
                    </SidebarGroup>
                )}

                {favoriteFlows.length > 0 && (
                    <SidebarGroup>
                        <SidebarGroupLabel className="flex items-center gap-2">
                            <Star />
                            Favorite Flows
                        </SidebarGroupLabel>
                        <SidebarGroupContent>
                            <SidebarMenu>
                                {favoriteFlows.map((flow) => (
                                    <SidebarMenuItem
                                        key={flow.id}
                                        onMouseLeave={(e) => {
                                            const menuItem = e.currentTarget;
                                            menuItem.querySelectorAll('button, a').forEach((el) => {
                                                if (el instanceof HTMLElement) {
                                                    el.blur();
                                                }
                                            });

                                            const key = `favorite-${flow.id}`;
                                            setClickedButtons((prev) => {
                                                const next = new Set(prev);
                                                next.delete(key);

                                                return next;
                                            });
                                        }}
                                    >
                                        <SidebarMenuButton
                                            asChild
                                            isActive={flowId === +flow.id}
                                        >
                                            <Link to={`/flows/${flow.id}`}>
                                                <span className="-mx-2 w-8 shrink-0 text-center text-xs group-data-[state=expanded]:hidden">
                                                    {flow.id}
                                                </span>
                                                <span className="text-muted-foreground bg-background dark:bg-muted -my-0.5 -ml-0.5 h-5 min-w-5 shrink-0 rounded-md px-px py-0.5 text-center text-xs group-data-[state=collapsed]:hidden">
                                                    {flow.id}
                                                </span>
                                                <span className="truncate">{flow.title}</span>
                                            </Link>
                                        </SidebarMenuButton>
                                        <SidebarMenuAction
                                            className={`data-[state=open]:bg-accent rounded-sm ${clickedButtons.has(`favorite-${flow.id}`) ? 'pointer-events-none! opacity-0!' : ''}`}
                                            onClick={(e) => {
                                                e.preventDefault();
                                                e.stopPropagation();
                                                const button = e.currentTarget;
                                                button.blur();

                                                const key = `favorite-${flow.id}`;
                                                setClickedButtons((prev) => new Set(prev).add(key));
                                                removeFavoriteFlow(flow.id);

                                                setTimeout(() => {
                                                    setClickedButtons((prev) => {
                                                        const next = new Set(prev);
                                                        next.delete(key);

                                                        return next;
                                                    });
                                                }, 600);
                                            }}
                                            showOnHover
                                        >
                                            <Star className="fill-yellow-500 stroke-yellow-500" />
                                        </SidebarMenuAction>
                                    </SidebarMenuItem>
                                ))}
                            </SidebarMenu>
                        </SidebarGroupContent>
                    </SidebarGroup>
                )}
            </SidebarContent>
            <SidebarFooter>
                <SidebarMenu>
                    <SidebarMenuItem>
                        <SidebarMenuButton
                            asChild
                            isActive={!!isSettingsActive}
                        >
                            <Link to="/settings">
                                <Settings />
                                Settings
                            </Link>
                        </SidebarMenuButton>
                    </SidebarMenuItem>
                    <SidebarMenuItem>
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <SidebarMenuButton
                                    className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                                    size="lg"
                                >
                                    <Avatar className="bg-background dark:bg-muted size-8 rounded-lg">
                                        {/* <AvatarImage
                                            alt={user.name}
                                            src={user.avatar}
                                        /> */}
                                        <AvatarFallback className="flex size-8 items-center justify-center">
                                            <UserIcon className="size-4" />
                                        </AvatarFallback>
                                    </Avatar>
                                    <div className="grid flex-1 text-left text-sm leading-tight">
                                        <span className="truncate font-semibold">{user?.name}</span>
                                        <span className="truncate text-xs">{user?.mail}</span>
                                    </div>
                                    <ChevronsUpDown className="ml-auto size-4" />
                                </SidebarMenuButton>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                                align="end"
                                className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
                                side="bottom"
                                sideOffset={4}
                            >
                                <DropdownMenuLabel className="p-0 font-normal">
                                    <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                                        <Avatar className="bg-muted flex size-8 items-center justify-center rounded-lg">
                                            <AvatarFallback className="flex items-center justify-center rounded-lg">
                                                <UserIcon className="size-4" />
                                            </AvatarFallback>
                                        </Avatar>
                                        <div className="grid flex-1 text-left text-sm leading-tight">
                                            <span className="truncate font-semibold">{user?.name}</span>
                                            <span className="truncate text-xs">{user?.mail}</span>
                                            <span className="text-muted-foreground truncate text-xs">
                                                {user?.type === 'local' ? 'local' : 'oauth'}
                                            </span>
                                        </div>
                                    </div>
                                </DropdownMenuLabel>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem
                                    className="cursor-default hover:bg-transparent focus:bg-transparent"
                                    onSelect={(event) => event.preventDefault()}
                                >
                                    <Settings2 />
                                    Theme
                                    <Tabs
                                        className="-my-1.5 -mr-2 ml-auto"
                                        onValueChange={(value) => setTheme(value as Theme)}
                                        value={theme || 'system'}
                                    >
                                        <TabsList className="h-7 p-0.5">
                                            <TabsTrigger
                                                className="h-6 px-2"
                                                value="system"
                                            >
                                                <Monitor className="size-4" />
                                            </TabsTrigger>
                                            <TabsTrigger
                                                className="h-6 px-2"
                                                value="light"
                                            >
                                                <Sun className="size-4" />
                                            </TabsTrigger>
                                            <TabsTrigger
                                                className="h-6 px-2"
                                                value="dark"
                                            >
                                                <Moon className="size-4" />
                                            </TabsTrigger>
                                        </TabsList>
                                    </Tabs>
                                </DropdownMenuItem>
                                {user?.type === 'local' && (
                                    <>
                                        <DropdownMenuSeparator />
                                        <DropdownMenuItem onClick={() => setIsPasswordModalOpen(true)}>
                                            <KeyRound className="mr-2 size-4" />
                                            Change Password
                                        </DropdownMenuItem>
                                    </>
                                )}
                                <DropdownMenuSeparator />
                                <DropdownMenuItem onClick={() => logout()}>
                                    <LogOut className="mr-2 size-4" />
                                    Log out
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarFooter>
            <SidebarRail />

            <Dialog
                onOpenChange={(open) => setIsPasswordModalOpen(open)}
                open={isPasswordModalOpen}
            >
                <DialogContent className="sm:max-w-[425px]">
                    <DialogHeader>
                        <DialogTitle>Change Password</DialogTitle>
                    </DialogHeader>
                    <PasswordChangeForm
                        onCancel={() => setIsPasswordModalOpen(false)}
                        onSuccess={handlePasswordChangeSuccess}
                    />
                </DialogContent>
            </Dialog>
        </Sidebar>
    );
};

export default MainSidebar;
