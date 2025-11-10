import { Avatar, AvatarFallback } from '@radix-ui/react-avatar';
import { Check, ChevronsUpDown, GitFork, KeyRound, LogOut, Moon, Settings, Sun, UserIcon } from 'lucide-react';
import { useState } from 'react';
import { useMatch, useNavigate } from 'react-router-dom';

import Logo from '@/components/icons/Logo';
import { Button } from '@/components/ui/button';
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
    SidebarHeader,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
    SidebarRail,
} from '@/components/ui/sidebar';
import { PasswordChangeForm } from '@/features/authentication/PasswordChangeForm';
import { cn } from '@/lib/utils';
import { getProviderDisplayName, getProviderIcon, type Provider } from '@/models/Provider';
import { useTheme } from '@/providers/ThemeProvider';
import { useUser } from '@/providers/UserProvider';

interface ChatSidebarProps {
    providers: Provider[];
    selectedProvider: Provider | null;
    onChangeSelectedProvider: (provider: Provider) => void;
}

const ChatSidebar = ({ providers, selectedProvider, onChangeSelectedProvider }: ChatSidebarProps) => {
    const navigate = useNavigate();
    const [isPasswordModalOpen, setIsPasswordModalOpen] = useState(false);

    const isFlowsActive = useMatch('/flows/*');
    const isSettingsActive = useMatch('/settings/*');

    const { authInfo, logout } = useUser();
    const user = authInfo?.user;
    const { theme, setTheme } = useTheme();

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
                                    <span className="truncate max-w-[90px]">
                                        {selectedProvider
                                            ? getProviderDisplayName(selectedProvider)
                                            : 'Select Provider'}
                                    </span>
                                    <ChevronsUpDown className="size-4 shrink-0" />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                                align="end"
                                className="min-w-[150px] max-w-[280px] w-fit"
                                onCloseAutoFocus={(e) => {
                                    e.preventDefault();
                                }}
                            >
                                {providers.map((provider) => (
                                    <DropdownMenuItem
                                        key={provider.name}
                                        onSelect={(e) => {
                                            e.preventDefault();
                                            onChangeSelectedProvider(provider);
                                        }}
                                        className="focus:outline-none focus-visible:outline-none focus-visible:ring-0"
                                    >
                                        <div className="flex items-center gap-2 w-full min-w-0">
                                            <div className="shrink-0">
                                                {getProviderIcon(provider, 'h-4 w-4 shrink-0')}
                                            </div>
                                            <span className="flex-1 truncate max-w-[180px]">
                                                {getProviderDisplayName(provider)}
                                            </span>
                                            {selectedProvider?.name === provider.name && (
                                                <div className="shrink-0">
                                                    <Check className="h-4 w-4 shrink-0" />
                                                </div>
                                            )}
                                        </div>
                                    </DropdownMenuItem>
                                ))}
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarHeader>
            <SidebarContent>
                <SidebarGroup>
                    <SidebarGroupContent>
                        <SidebarMenu>
                            <SidebarMenuItem>
                                <SidebarMenuButton
                                    asChild
                                    className={cn(
                                        'relative cursor-pointer overflow-hidden hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                                        {
                                            'bg-sidebar-accent text-sidebar-accent-foreground font-medium before:absolute before:left-0 before:top-0 before:bottom-0 before:w-1 before:bg-primary dark:before:bg-primary-foreground':
                                                isFlowsActive,
                                        },
                                    )}
                                    onClick={() => navigate('/flows')}
                                >
                                    <div className="flex w-full items-center gap-2">
                                        <GitFork className="size-4" />
                                        <span>Flows</span>
                                    </div>
                                </SidebarMenuButton>
                            </SidebarMenuItem>
                        </SidebarMenu>
                    </SidebarGroupContent>
                </SidebarGroup>
            </SidebarContent>
            <SidebarFooter>
                <SidebarMenu>
                    <SidebarMenuItem>
                        <SidebarMenuButton
                            asChild
                            className={cn(
                                'relative cursor-pointer overflow-hidden hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                                {
                                    'bg-sidebar-accent text-sidebar-accent-foreground font-medium before:absolute before:left-0 before:top-0 before:bottom-0 before:w-1 before:bg-primary dark:before:bg-primary-foreground':
                                        isSettingsActive,
                                },
                            )}
                            onClick={() => navigate('/settings')}
                        >
                            <div className="flex w-full items-center gap-2">
                                <Settings className="size-4" />
                                <span>Settings</span>
                            </div>
                        </SidebarMenuButton>
                    </SidebarMenuItem>
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
                                <DropdownMenuItem onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}>
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
                open={isPasswordModalOpen}
                onOpenChange={(open) => setIsPasswordModalOpen(open)}
            >
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
