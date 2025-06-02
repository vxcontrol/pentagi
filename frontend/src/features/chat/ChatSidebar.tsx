import { Avatar, AvatarFallback } from '@radix-ui/react-avatar';
import {
    Check,
    ChevronsUpDown,
    KeyRound,
    LogOut,
    Moon,
    MoreHorizontal,
    Pause,
    Plus,
    Sun,
    Trash2,
    UserIcon,
    X,
} from 'lucide-react';
import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

import Logo from '@/components/icons/Logo';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
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
    SidebarGroupLabel,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuAction,
    SidebarMenuButton,
    SidebarMenuIndicator,
    SidebarMenuItem,
    SidebarRail,
} from '@/components/ui/sidebar';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { PasswordChangeForm } from '@/features/authentication/PasswordChangeForm';
import type { FlowOverviewFragmentFragment } from '@/graphql/types';
import { StatusType } from '@/graphql/types';
import { axios } from '@/lib/axios';
import { cn } from '@/lib/utils';
import type { User } from '@/models/User';
import { useThemeStore } from '@/store/theme-store';

interface ChatSidebarProps {
    user: User | null;
    flows: FlowOverviewFragmentFragment[];
    providers: string[];
    selectedProvider: string;
    onChangeSelectedProvider: (provider: string) => void;
    selectedFlowId: string | null;
    onChangeSelectedFlowId: (id: string) => void;
    onDeleteFlow: (id: string) => Promise<void>;
    onFinishFlow: (id: string) => Promise<void>;
}

const ChatSidebarMenuItemText = ({ text }: { text: string }) => {
    const [isTruncated, setIsTruncated] = useState(false);
    const ref = useRef<HTMLSpanElement>(null);

    useLayoutEffect(() => {
        const element = ref.current;

        if (element) {
            const shouldTruncate = element.scrollWidth > element.clientWidth;
            setIsTruncated((current) => shouldTruncate !== current ? shouldTruncate : current);
        }
    }, [text]);

    if (!isTruncated) {
        return (
            <span
                ref={ref}
                className="truncate"
            >
                {text}
            </span>
        );
    }

    return (
        <Tooltip delayDuration={1000}>
            <TooltipTrigger asChild>
                <span
                    ref={ref}
                    className="truncate"
                >
                    {text}
                </span>
            </TooltipTrigger>
            <TooltipContent side="right">{text}</TooltipContent>
        </Tooltip>
    );
};

const ChatSidebar = ({
    user,
    flows,
    providers,
    selectedProvider,
    onChangeSelectedProvider,
    selectedFlowId,
    onChangeSelectedFlowId,
    onDeleteFlow,
    onFinishFlow,
}: ChatSidebarProps) => {
    const navigate = useNavigate();
    const location = useLocation();
    const [isPasswordModalOpen, setIsPasswordModalOpen] = useState(false);
    const prevPathRef = useRef(location.pathname);
    const [openMenuFlowId, setOpenMenuFlowId] = useState<string | null>(null);
    const [hoveredFlowId, setHoveredFlowId] = useState<string | null>(null);

    const theme = useThemeStore((store) => store.theme);
    const toggleTheme = useThemeStore((store) => store.setTheme);

    useEffect(() => {
        prevPathRef.current = location.pathname;
    }, [location.pathname]);

    useEffect(() => {
        const handleGlobalClick = () => {
            setHoveredFlowId(null);
        };

        document.addEventListener('click', handleGlobalClick);
        return () => {
            document.removeEventListener('click', handleGlobalClick);
        };
    }, []);

    const logout = async () => {
        try {
            await axios.get('/auth/logout');
        } finally {
            // Save current location for redirect after login
            const currentPath = location.pathname;
            // Don't save if it's the default new chat page
            const returnUrl = currentPath === '/chat/new' ? '' : `?returnUrl=${encodeURIComponent(currentPath)}`;

            localStorage.removeItem('auth');
            navigate(`/login${returnUrl}`);
        }
    };

    const handlePasswordChangeSuccess = () => {
        setIsPasswordModalOpen(false);
    };

    return (
        <Sidebar>
            <SidebarHeader>
                <SidebarMenu>
                    <SidebarMenuItem className="flex items-center gap-2">
                        <div className="flex items-center gap-2">
                            <Logo className="size-6 hover:animate-logo-spin" />
                            <span className="font-semibold">PentAGI</span>
                        </div>
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="ml-auto h-8 gap-1 focus:outline-none focus-visible:outline-none focus-visible:ring-0 data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                                >
                                    <span>{selectedProvider}</span>
                                    <ChevronsUpDown className="size-4" />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                                align="end"
                                className="w-[--radix-dropdown-menu-trigger-width]"
                                onCloseAutoFocus={(e) => {
                                    e.preventDefault();
                                }}
                            >
                                {providers.map((provider) => (
                                    <DropdownMenuItem
                                        key={provider}
                                        onSelect={(e) => {
                                            e.preventDefault();
                                            onChangeSelectedProvider(provider);
                                        }}
                                        className="focus:outline-none focus-visible:outline-none focus-visible:ring-0"
                                    >
                                        {provider} {provider === selectedProvider && <Check className="ml-auto" />}
                                    </DropdownMenuItem>
                                ))}
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarHeader>
            <SidebarContent>
                <SidebarGroup className="group-data-[collapsible=icon]:hidden">
                    <div className="flex items-center justify-between">
                        <SidebarGroupLabel>Flows</SidebarGroupLabel>
                        <Button
                            variant="ghost"
                            size="icon"
                            className={cn(
                                'relative size-8',
                                (selectedFlowId === 'new' || window.location.pathname === '/chat/new') && 'text-primary after:absolute after:left-1/2 after:top-full after:size-1.5 after:-translate-x-1/2 after:rounded-full after:bg-primary dark:text-primary-foreground dark:after:bg-primary-foreground',
                            )}
                            onClick={() => onChangeSelectedFlowId('new')}
                        >
                            <Plus className="size-4" />
                            <span className="sr-only">Create new flow</span>
                        </Button>
                    </div>
                    <SidebarMenu>
                        {flows.map((flow) => {
                            const isSelected = selectedFlowId === flow.id || location.pathname === `/chat/${flow.id}`;
                            const isRunning = flow.status === StatusType.Running;
                            const shouldShowIndicator = isRunning && !isSelected;
                            const isMenuOpen = openMenuFlowId === flow.id;
                            const shouldShowAction = isSelected || isMenuOpen || hoveredFlowId === flow.id;

                            return (
                                <SidebarMenuItem
                                    key={flow.id}
                                    data-has-action="true"
                                    onMouseEnter={() => setHoveredFlowId(flow.id)}
                                    onMouseLeave={() => {
                                        if (!isMenuOpen) {
                                            setHoveredFlowId(null);
                                        }
                                    }}
                                >
                                    <SidebarMenuButton
                                        asChild
                                        className={cn(
                                            'relative cursor-pointer overflow-hidden hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                                            {
                                                'bg-sidebar-accent text-sidebar-accent-foreground font-medium before:absolute before:left-0 before:top-0 before:bottom-0 before:w-1 before:bg-primary dark:before:bg-primary-foreground':
                                                    isSelected,
                                                'text-muted-foreground': [StatusType.Finished, StatusType.Failed].includes(
                                                    flow.status,
                                                ),
                                            },
                                        )}
                                        onClick={() => onChangeSelectedFlowId(flow.id)}
                                    >
                                        <div className="flex w-full items-center gap-2">
                                            <ChatSidebarMenuItemText text={flow.title} />
                                            {flow.status === StatusType.Finished && <Check className="opacity-25" />}
                                            {flow.status === StatusType.Failed && <X className="opacity-25" />}
                                        </div>
                                    </SidebarMenuButton>

                                    {shouldShowIndicator && !isMenuOpen && !shouldShowAction && (
                                        <SidebarMenuIndicator />
                                    )}

                                    {shouldShowAction && (
                                        <DropdownMenu
                                            open={isMenuOpen}
                                            onOpenChange={(open) => {
                                                setOpenMenuFlowId(open ? flow.id : null);
                                                if (!open) {
                                                    setHoveredFlowId(null);

                                                    if (document.activeElement instanceof HTMLElement) {
                                                        document.activeElement.blur();
                                                    }
                                                }
                                            }}
                                        >
                                            <DropdownMenuTrigger asChild>
                                                <SidebarMenuAction
                                                    className="focus:outline-none focus-visible:outline-none focus-visible:ring-0"
                                                >
                                                    <MoreHorizontal />
                                                </SidebarMenuAction>
                                            </DropdownMenuTrigger>
                                            <DropdownMenuContent
                                                className="w-48 rounded-lg"
                                                side="bottom"
                                                align="end"
                                            >
                                                <DropdownMenuItem
                                                    onClick={() => onFinishFlow(flow.id)}
                                                    className="cursor-pointer"
                                                >
                                                    <Pause className="mr-2 size-4" />
                                                    <span>Finish Flow</span>
                                                </DropdownMenuItem>
                                                <DropdownMenuSeparator />
                                                <DropdownMenuItem
                                                    onClick={() => onDeleteFlow(flow.id)}
                                                    className="cursor-pointer text-destructive hover:bg-destructive hover:text-destructive-foreground"
                                                >
                                                    <Trash2 className="mr-2 size-4" />
                                                    <span>Delete Flow</span>
                                                </DropdownMenuItem>
                                            </DropdownMenuContent>
                                        </DropdownMenu>
                                    )}
                                </SidebarMenuItem>
                            );
                        })}
                    </SidebarMenu>
                </SidebarGroup>
            </SidebarContent>
            <SidebarFooter>
                <SidebarMenu>
                    <SidebarMenuItem>
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <SidebarMenuButton
                                    size="lg"
                                    className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                                >
                                    <Avatar className="flex size-8 items-center justify-center rounded-lg bg-muted">
                                        <AvatarFallback className="flex items-center justify-center rounded-lg">
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
                                className="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg"
                                side="bottom"
                                align="end"
                                sideOffset={4}
                            >
                                <DropdownMenuLabel className="p-0 font-normal">
                                    <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                                        <Avatar className="flex size-8 items-center justify-center rounded-lg bg-muted">
                                            <AvatarFallback className="flex items-center justify-center rounded-lg">
                                                <UserIcon className="size-4" />
                                            </AvatarFallback>
                                        </Avatar>
                                        <div className="grid flex-1 text-left text-sm leading-tight">
                                            <span className="truncate font-semibold">{user?.name}</span>
                                            <span className="truncate text-xs">{user?.mail}</span>
                                            <span className="truncate text-xs text-muted-foreground">
                                                {user?.type === 'local' ? 'local' : 'oauth'}
                                            </span>
                                        </div>
                                    </div>
                                </DropdownMenuLabel>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem onClick={toggleTheme}>
                                    {theme === 'light' ? (
                                        <Moon className="mr-2 size-4" />
                                    ) : (
                                        <Sun className="mr-2 size-4" />
                                    )}
                                    <span>{theme === 'light' ? 'Dark mode' : 'Light mode'}</span>
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
                                <DropdownMenuItem onClick={logout}>
                                    <LogOut className="mr-2 size-4" />
                                    Log out
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarFooter>
            <SidebarRail />

            <Dialog open={isPasswordModalOpen} onOpenChange={(open) => setIsPasswordModalOpen(open)}>
                <DialogContent className="sm:max-w-[425px]">
                    <DialogHeader>
                        <DialogTitle>Change Password</DialogTitle>
                    </DialogHeader>
                    <PasswordChangeForm
                        onSuccess={handlePasswordChangeSuccess}
                        onCancel={() => setIsPasswordModalOpen(false)}
                    />
                </DialogContent>
            </Dialog>
        </Sidebar>
    );
};

export default ChatSidebar;
