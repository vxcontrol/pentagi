import { ArrowLeft, Bot, Wrench } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

import Logo from '@/components/icons/Logo';
import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarGroup,
    SidebarGroupLabel,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
    SidebarRail,
} from '@/components/ui/sidebar';
import { cn } from '@/lib/utils';

interface SettingsSidebarProps {
    selectedSection: string;
    onSectionChange: (section: string) => void;
}

// Типы для данных
interface Provider {
    id: string;
    name: string;
    displayName: string;
}

interface Agent {
    id: string;
    name: string;
    displayName: string;
}

interface Tool {
    id: string;
    name: string;
    displayName: string;
}

// Тестовые данные - позже заменим на API вызовы
const mockProviders: Provider[] = [
    { id: 'openai', name: 'openai', displayName: 'OpenAI' },
    { id: 'anthropic', name: 'anthropic', displayName: 'Anthropic' },
    { id: 'deepseek', name: 'deepseek', displayName: 'DeepSeek' },
    { id: 'openrouter', name: 'openrouter', displayName: 'OpenRouter' },
];

const mockAgents: Agent[] = [
    { id: 'assistant', name: 'assistant', displayName: 'Assistant' },
    { id: 'coder', name: 'coder', displayName: 'Coder' },
    { id: 'pentest', name: 'pentest', displayName: 'Pentest' },
    { id: 'reviewer', name: 'reviewer', displayName: 'Code Reviewer' },
];

const mockTools: Tool[] = [
    { id: 'browser', name: 'browser', displayName: 'Browser' },
    { id: 'terminal', name: 'terminal', displayName: 'Terminal' },
    { id: 'search', name: 'search', displayName: 'Search' },
    { id: 'file', name: 'file', displayName: 'File Operations' },
];

const SettingsSidebar = ({ selectedSection, onSectionChange }: SettingsSidebarProps) => {
    const navigate = useNavigate();
    const location = useLocation();
    const [providers, setProviders] = useState<Provider[]>([]);
    const [agents, setAgents] = useState<Agent[]>([]);
    const [tools, setTools] = useState<Tool[]>([]);

    // Загружаем данные (пока тестовые)
    useEffect(() => {
        // TODO: Заменить на реальные API вызовы
        setProviders(mockProviders);
        setAgents(mockAgents);
        setTools(mockTools);
    }, []);

    // Автоматически перенаправляем на первый провайдер если мы на главной странице настроек
    useEffect(() => {
        if (location.pathname === '/settings' && providers.length > 0) {
            const firstProvider = providers[0];
            if (firstProvider) {
                navigate(`/settings/models/${firstProvider.id}`);
                onSectionChange(firstProvider.id);
            }
        }
    }, [location.pathname, providers, navigate, onSectionChange]);

    // Создаем структуру меню динамически
    const settingsMenuItems = [
        {
            id: 'models',
            label: 'Models',
            icon: Bot,
            items: providers.map((provider) => ({
                id: provider.id,
                label: provider.displayName,
                path: `/settings/models/${provider.id}`,
            })),
        },
        {
            id: 'prompts',
            label: 'Prompts',
            icon: Wrench,
            items: [
                // Agents группа
                ...agents.map((agent) => ({
                    id: `agent-${agent.id}`,
                    label: agent.displayName,
                    path: `/settings/prompts/agents/${agent.id}`,
                    group: 'Agents',
                })),
                // Tools группа
                ...tools.map((tool) => ({
                    id: `tool-${tool.id}`,
                    label: tool.displayName,
                    path: `/settings/prompts/tools/${tool.id}`,
                    group: 'Tools',
                })),
            ],
        },
    ];

    const handleNavigate = (path: string, sectionId: string) => {
        navigate(path);
        onSectionChange(sectionId);
    };

    // Группируем элементы Prompts по группам
    const groupPromptsItems = (items: any[]) => {
        const grouped = items.reduce((acc: any, item) => {
            if (item.group) {
                if (!acc[item.group]) acc[item.group] = [];
                acc[item.group].push(item);
            }
            return acc;
        }, {});

        return grouped;
    };

    return (
        <Sidebar>
            <SidebarHeader>
                <SidebarMenu>
                    <SidebarMenuItem className="flex items-center gap-2">
                        <div className="flex items-center gap-2">
                            <Logo className="size-6 hover:animate-logo-spin" />
                            <span className="font-semibold">Settings</span>
                        </div>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarHeader>

            <SidebarContent>
                {settingsMenuItems.map((group) => (
                    <SidebarGroup
                        key={group.id}
                        className="group-data-[collapsible=icon]:hidden"
                    >
                        <SidebarGroupLabel className="flex items-center gap-2">
                            <group.icon className="size-4" />
                            {group.label}
                        </SidebarGroupLabel>
                        <SidebarMenu>
                            {/* Для Models - обычный список */}
                            {group.id === 'models' &&
                                group.items.map((item: any) => {
                                    const isSelected = location.pathname === item.path;
                                    return (
                                        <SidebarMenuItem key={item.id}>
                                            <SidebarMenuButton
                                                asChild
                                                className={cn(
                                                    'relative cursor-pointer overflow-hidden hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                                                    {
                                                        'bg-sidebar-accent text-sidebar-accent-foreground font-medium before:absolute before:left-0 before:top-0 before:bottom-0 before:w-1 before:bg-primary dark:before:bg-primary-foreground':
                                                            isSelected,
                                                    },
                                                )}
                                                onClick={() => handleNavigate(item.path, item.id)}
                                            >
                                                <div className="flex w-full items-center gap-2">
                                                    <span className="truncate">{item.label}</span>
                                                </div>
                                            </SidebarMenuButton>
                                        </SidebarMenuItem>
                                    );
                                })}

                            {/* Для Prompts - группированный список */}
                            {group.id === 'prompts' &&
                                Object.entries(groupPromptsItems(group.items)).map(
                                    ([groupName, groupItems]: [string, any]) => (
                                        <div key={groupName}>
                                            <div className="px-2 py-1.5 text-xs font-medium text-muted-foreground">
                                                {groupName}
                                            </div>
                                            {groupItems.map((item: any) => {
                                                const isSelected = location.pathname === item.path;
                                                return (
                                                    <SidebarMenuItem key={item.id}>
                                                        <SidebarMenuButton
                                                            asChild
                                                            className={cn(
                                                                'relative cursor-pointer overflow-hidden hover:bg-sidebar-accent hover:text-sidebar-accent-foreground ml-2',
                                                                {
                                                                    'bg-sidebar-accent text-sidebar-accent-foreground font-medium before:absolute before:left-0 before:top-0 before:bottom-0 before:w-1 before:bg-primary dark:before:bg-primary-foreground':
                                                                        isSelected,
                                                                },
                                                            )}
                                                            onClick={() => handleNavigate(item.path, item.id)}
                                                        >
                                                            <div className="flex w-full items-center gap-2">
                                                                <span className="truncate">{item.label}</span>
                                                            </div>
                                                        </SidebarMenuButton>
                                                    </SidebarMenuItem>
                                                );
                                            })}
                                        </div>
                                    ),
                                )}
                        </SidebarMenu>
                    </SidebarGroup>
                ))}
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
                                        location.pathname.startsWith('/chat'),
                                },
                            )}
                            onClick={() => navigate('/chat/new')}
                        >
                            <div className="flex w-full items-center gap-2">
                                <ArrowLeft className="size-4" />
                                <span>Back to Chat</span>
                            </div>
                        </SidebarMenuButton>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarFooter>

            <SidebarRail />
        </Sidebar>
    );
};

export default SettingsSidebar;
