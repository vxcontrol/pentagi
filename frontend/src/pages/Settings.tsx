import { useState } from 'react';
import { useLocation, useParams } from 'react-router-dom';

import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Separator } from '@/components/ui/separator';
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar';
import ModelForm from '@/features/settings/ModelForm';
import PromptForm from '@/features/settings/PromptForm';
import SettingsSidebar from '@/features/settings/SettingsSidebar';
import { Bot, Settings as SettingsIcon, Wrench } from 'lucide-react';

const Settings = () => {
    const location = useLocation();
    const { providerId, agentId, toolId } = useParams<{
        providerId?: string;
        agentId?: string;
        toolId?: string;
    }>();

    const [selectedSection, setSelectedSection] = useState(() => {
        if (providerId) return providerId;
        if (agentId) return `agent-${agentId}`;
        if (toolId) return `tool-${toolId}`;
        return 'models';
    });

    // Определяем тип страницы на основе роута
    const getPageType = () => {
        if (location.pathname.includes('/models/')) return 'model';
        if (location.pathname.includes('/prompts/agents/')) return 'prompt-agent';
        if (location.pathname.includes('/prompts/tools/')) return 'prompt-tool';
        return 'default';
    };

    // Получаем данные для отображения
    const getDisplayData = () => {
        const pageType = getPageType();

        // Мапы для отображения имен (позже заменим на API)
        const providerNames: Record<string, string> = {
            openai: 'OpenAI',
            anthropic: 'Anthropic',
            deepseek: 'DeepSeek',
            openrouter: 'OpenRouter',
        };

        const agentNames: Record<string, string> = {
            assistant: 'Assistant',
            coder: 'Coder',
            pentest: 'Pentest',
            reviewer: 'Code Reviewer',
        };

        const toolNames: Record<string, string> = {
            browser: 'Browser',
            terminal: 'Terminal',
            search: 'Search',
            file: 'File Operations',
        };

        switch (pageType) {
            case 'model':
                return {
                    type: 'model' as const,
                    id: providerId!,
                    name: providerNames[providerId!] || providerId!,
                    icon: Bot,
                    title: `${providerNames[providerId!] || providerId!} Model Settings`,
                };
            case 'prompt-agent':
                return {
                    type: 'prompt-agent' as const,
                    id: agentId!,
                    name: agentNames[agentId!] || agentId!,
                    icon: Bot,
                    title: `${agentNames[agentId!] || agentId!} Agent Prompts`,
                };
            case 'prompt-tool':
                return {
                    type: 'prompt-tool' as const,
                    id: toolId!,
                    name: toolNames[toolId!] || toolId!,
                    icon: Wrench,
                    title: `${toolNames[toolId!] || toolId!} Tool Prompts`,
                };
            default:
                return {
                    type: 'default' as const,
                    id: '',
                    name: 'Settings',
                    icon: SettingsIcon,
                    title: 'Settings',
                };
        }
    };

    const displayData = getDisplayData();

    const handleSaveSettings = async (values: Record<string, any>) => {
        console.log('Saving settings:', values);
        console.log('Section:', selectedSection);
        console.log('Type:', displayData.type);

        // TODO: Добавить API вызовы для сохранения настроек
        // const apiEndpoint = getApiEndpointForType(displayData.type);
        // await fetch(apiEndpoint, { method: 'POST', body: JSON.stringify(values) });

        // Показываем уведомление об успешном сохранении
        alert('Settings saved successfully!');
    };

    // Рендерим соответствующую форму
    const renderForm = () => {
        const { type, id, name } = displayData;

        switch (type) {
            case 'model':
                return (
                    <ModelForm
                        providerId={id}
                        providerName={name}
                        onSave={handleSaveSettings}
                    />
                );
            case 'prompt-agent':
                return (
                    <PromptForm
                        itemId={id}
                        itemName={name}
                        itemType="agent"
                        onSave={handleSaveSettings}
                    />
                );
            case 'prompt-tool':
                return (
                    <PromptForm
                        itemId={id}
                        itemName={name}
                        itemType="tool"
                        onSave={handleSaveSettings}
                    />
                );
            default:
                return (
                    <div className="flex items-center justify-center h-64">
                        <div className="text-center">
                            <SettingsIcon className="h-16 w-16 mx-auto text-muted-foreground mb-4" />
                            <h3 className="text-lg font-medium mb-2">Welcome to Settings</h3>
                            <p className="text-muted-foreground">
                                Select a configuration option from the sidebar to get started.
                            </p>
                        </div>
                    </div>
                );
        }
    };

    return (
        <SidebarProvider>
            <SettingsSidebar
                selectedSection={selectedSection}
                onSectionChange={setSelectedSection}
            />
            <SidebarInset>
                <header className="fixed top-0 z-10 flex h-12 w-full shrink-0 items-center gap-2 border-b bg-background transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12">
                    <div className="flex items-center gap-2 px-4">
                        <SidebarTrigger className="-ml-1" />
                        <Separator
                            orientation="vertical"
                            className="mr-2 h-4"
                        />
                        <Breadcrumb>
                            <BreadcrumbList>
                                <BreadcrumbItem>
                                    <BreadcrumbPage className="flex items-center gap-2">
                                        {(() => {
                                            const IconComponent = displayData.icon;
                                            return <IconComponent className="h-4 w-4" />;
                                        })()}
                                        {displayData.title}
                                    </BreadcrumbPage>
                                </BreadcrumbItem>
                            </BreadcrumbList>
                        </Breadcrumb>
                    </div>
                </header>

                <div className="mt-12 flex h-[calc(100dvh-3rem)] w-full max-w-full flex-1 overflow-auto">
                    <div className="container mx-auto px-6 py-6">
                        <div className="mx-auto max-w-4xl">{renderForm()}</div>
                    </div>
                </div>
            </SidebarInset>
        </SidebarProvider>
    );
};

export default Settings;
