import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { AppHeader, AppHeaderContent, AppHeaderTitle } from '@/components/layouts/app/app-header';
import { Card, CardContent } from '@/components/ui/card';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { FlowForm, type FlowFormValues } from '@/features/flows/flow-form';
import { useI18n } from '@/hooks/use-i18n';
import { routes } from '@/lib/routes';
import { useFlows } from '@/providers/flows-provider';
import { useProviders } from '@/providers/providers-provider';
import { useSystemSettings } from '@/providers/system-settings-provider';

function NewFlow() {
    const navigate = useNavigate();
    const { t } = useI18n();

    const { selectedProvider } = useProviders();
    const { createFlow, createFlowWithAssistant } = useFlows();
    const { settings } = useSystemSettings();

    const [isLoading, setIsLoading] = useState(false);
    const [flowType, setFlowType] = useState<'assistant' | 'automation'>('automation');

    const shouldUseAgents = useMemo(() => {
        return settings?.assistantUseAgents ?? false;
    }, [settings?.assistantUseAgents]);

    const handleSubmit = async (values: FlowFormValues) => {
        if (isLoading) {
            return;
        }

        setIsLoading(true);

        try {
            const flowId = flowType === 'automation' ? await createFlow(values) : await createFlowWithAssistant(values);

            if (flowId) {
                navigate(routes.flow(flowId, { tab: flowType }));
            }
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <>
            <AppHeader>
                <AppHeaderContent>
                    <AppHeaderTitle>{t('flow.newFlow')}</AppHeaderTitle>
                </AppHeaderContent>
            </AppHeader>
            <div className="flex min-h-[calc(100dvh-3rem)] items-center justify-center p-4">
                <Card className="w-full max-w-2xl">
                    <CardContent className="flex flex-col gap-4 pt-6">
                        <div className="flex flex-col gap-2 text-center">
                            <h2 className="text-2xl font-semibold">{t('flow.createFlow')}</h2>
                            <p className="text-muted-foreground">{t('flow.describeFlow')}</p>
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
                                    {t('flow.automation')}
                                </TabsTrigger>
                                <TabsTrigger
                                    disabled={isLoading}
                                    value="assistant"
                                >
                                    {t('flow.assistant')}
                                </TabsTrigger>
                            </TabsList>
                        </Tabs>
                        <FlowForm
                            defaultValues={{
                                providerName: selectedProvider?.name ?? '',
                                useAgents: shouldUseAgents,
                            }}
                            isSubmitting={isLoading}
                            onSubmit={handleSubmit}
                            placeholder={
                                !isLoading
                                    ? flowType === 'automation'
                                        ? t('flow.describeAutomation')
                                        : t('flow.describeAssistant')
                                    : t('flow.creatingFlow')
                            }
                            type={flowType}
                        />
                    </CardContent>
                </Card>
            </div>
        </>
    );
}

export default NewFlow;
