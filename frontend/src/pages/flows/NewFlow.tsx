import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Card, CardContent } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { FlowForm, type FlowFormValues } from '@/features/flows/FlowForm';
import { useCreateAssistantMutation, useCreateFlowMutation } from '@/graphql/types';
import { Log } from '@/lib/log';
import { useProviders } from '@/providers/ProvidersProvider';

const NewFlow = () => {
    const navigate = useNavigate();

    const { selectedProvider } = useProviders();

    const [createFlow] = useCreateFlowMutation();
    const [createAssistant] = useCreateAssistantMutation();
    const [isLoading, setIsLoading] = useState(false);
    const [flowType, setFlowType] = useState<'assistant' | 'automation'>('automation');

    const handleSubmit = async (values: FlowFormValues) => {
        const { message, providerName, useAgents } = values;

        const input = message.trim();
        const modelProvider = providerName.trim();

        if (!input || !modelProvider || isLoading) {
            return;
        }

        try {
            setIsLoading(true);

            if (flowType === 'automation') {
                const { data } = await createFlow({
                    variables: {
                        input,
                        modelProvider,
                    },
                });

                if (data?.createFlow) {
                    // Navigate to the new flow page
                    navigate(`/flows/${data.createFlow.id}`);
                }
            } else {
                const { data } = await createAssistant({
                    variables: {
                        flowId: '0',
                        input,
                        modelProvider,
                        useAgents,
                    },
                });

                if (data?.createAssistant?.flow) {
                    // Navigate to the new flow page
                    navigate(`/flows/${data.createAssistant.flow.id}`);
                }
            }
        } catch (error) {
            const title = flowType === 'automation' ? 'Failed to create flow' : 'Failed to create assistant';
            const description = error instanceof Error ? error.message : 'An error occurred while creating the flow';
            toast.error(title, {
                description,
            });
            Log.error(`Error creating ${flowType}:`, error);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <>
            <header className="sticky top-0 z-10 flex h-12 shrink-0 items-center gap-2 border-b bg-background px-4">
                <SidebarTrigger className="-ml-1" />
                <Separator
                    className="mr-2 h-4"
                    orientation="vertical"
                />
                <Breadcrumb>
                    <BreadcrumbList>
                        <BreadcrumbItem>
                            <BreadcrumbPage>New flow</BreadcrumbPage>
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>
            </header>
            <div className="flex min-h-[calc(100dvh-3rem)] items-center justify-center p-4">
                <Card className="w-full max-w-2xl">
                    <CardContent className="flex flex-col gap-4 pt-6">
                        <div className="text-center">
                            <h1 className="text-2xl font-semibold">Create a new flow</h1>
                            <p className="mt-2 text-muted-foreground">Describe what you would like PentAGI to test</p>
                        </div>
                        <Tabs
                            onValueChange={(value) => setFlowType(value as 'assistant' | 'automation')}
                            value={flowType}
                        >
                            <TabsList className="grid w-full grid-cols-2">
                                <TabsTrigger
                                    disabled={isLoading}
                                    value="automation"
                                >
                                    Automation
                                </TabsTrigger>
                                <TabsTrigger
                                    disabled={isLoading}
                                    value="assistant"
                                >
                                    Assistant
                                </TabsTrigger>
                            </TabsList>
                        </Tabs>
                        <FlowForm
                            defaultValues={{
                                providerName: selectedProvider?.name ?? '',
                            }}
                            isSubmitting={isLoading}
                            onSubmit={handleSubmit}
                            type={flowType}
                        />
                    </CardContent>
                </Card>
            </div>
        </>
    );
};

export default NewFlow;
