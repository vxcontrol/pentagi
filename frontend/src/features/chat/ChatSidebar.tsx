import { Avatar, AvatarFallback } from '@radix-ui/react-avatar';
import {
    Check,
    ChevronsUpDown,
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
import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import Logo from '@/components/icons/Logo';
import { Button } from '@/components/ui/button';
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
    SidebarMenuItem,
    SidebarRail,
} from '@/components/ui/sidebar';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
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

    useEffect(() => {
        const element = ref.current;

        if (element) {
            setIsTruncated(element.scrollWidth > element.clientWidth);
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
        <Tooltip>
            <TooltipTrigger asChild>
                <span
                    ref={ref}
                    className="truncate"
                >
                    {text}
                </span>
            </TooltipTrigger>
            <TooltipContent>{text}</TooltipContent>
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

    const theme = useThemeStore((store) => store.theme);
    const toggleTheme = useThemeStore((store) => store.setTheme);

    const logout = async () => {
        try {
            await axios.get('/auth/logout');
        } finally {
            localStorage.removeItem('auth');
            navigate('/login');
        }
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
                                    className="ml-auto h-8 gap-1 data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                                >
                                    <span>{selectedProvider}</span>
                                    <ChevronsUpDown className="size-4" />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                                align="end"
                                className="w-[--radix-dropdown-menu-trigger-width]"
                            >
                                {providers.map((provider) => (
                                    <DropdownMenuItem
                                        key={provider}
                                        onSelect={() => onChangeSelectedProvider(provider)}
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
                            className="size-8"
                            onClick={() => onChangeSelectedFlowId('new')}
                        >
                            <Plus className="size-4" />
                            <span className="sr-only">Create new flow</span>
                        </Button>
                    </div>
                    <SidebarMenu>
                        {flows.map((flow) => (
                            <SidebarMenuItem key={flow.id}>
                                <SidebarMenuButton
                                    asChild
                                    className={cn(
                                        'cursor-pointer hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                                        {
                                            'bg-sidebar-accent text-sidebar-accent-foreground':
                                                selectedFlowId === flow.id,
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
                                <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                        <SidebarMenuAction showOnHover>
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
                                            <span>Stop Flow</span>
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
                            </SidebarMenuItem>
                        ))}
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
                                    {theme === 'light' ? 'Dark mode' : 'Light mode'}
                                </DropdownMenuItem>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem onClick={logout}>
                                    <LogOut />
                                    Log out
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarFooter>
            <SidebarRail />
        </Sidebar>
    );
};

export default ChatSidebar;
