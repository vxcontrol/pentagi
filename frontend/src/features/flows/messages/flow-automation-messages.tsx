import { zodResolver } from '@hookform/resolvers/zod';
import debounce from 'lodash/debounce';
import { ChevronDown, Inbox, Search, X } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { StatusType } from '@/graphql/types';
import { useChatScroll } from '@/hooks/use-chat-scroll';
import { cn } from '@/lib/utils';
import { useFlow } from '@/providers/flow-provider';

import { FlowForm, type FlowFormValues } from '../flow-form';
import FlowMessage from './flow-message';

interface FlowAutomationMessagesProps {
    className?: string;
}

const searchFormSchema = z.object({
    search: z.string(),
});

const FlowAutomationMessages = ({ className }: FlowAutomationMessagesProps) => {
    const { flowData, flowId, flowStatus, stopAutomation, submitAutomationMessage } = useFlow();

    const logs = useMemo(() => flowData?.messageLogs ?? [], [flowData?.messageLogs]);

    // Separate state for immediate input value and debounced search value
    const [debouncedSearchValue, setDebouncedSearchValue] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [isCanceling, setIsCanceling] = useState(false);

    const { containerRef, endRef, hasNewMessages, isScrolledToBottom, scrollToEnd } = useChatScroll(logs, flowId);

    const form = useForm<z.infer<typeof searchFormSchema>>({
        defaultValues: {
            search: '',
        },
        resolver: zodResolver(searchFormSchema),
    });

    const searchValue = form.watch('search');

    const debouncedUpdateSearch = useMemo(
        () =>
            debounce((value: string) => {
                setDebouncedSearchValue(value);
            }, 500),
        [],
    );

    useEffect(() => {
        debouncedUpdateSearch(searchValue);

        return () => {
            debouncedUpdateSearch.cancel();
        };
    }, [searchValue, debouncedUpdateSearch]);

    // Cleanup debounced function on unmount
    useEffect(() => {
        return () => {
            debouncedUpdateSearch.cancel();
        };
    }, [debouncedUpdateSearch]);

    // Clear search when flow changes to prevent stale search state
    useEffect(() => {
        form.reset({ search: '' });
        setDebouncedSearchValue('');
        debouncedUpdateSearch.cancel();
    }, [flowId, form, debouncedUpdateSearch]);

    // Memoize filtered logs to avoid recomputing on every render
    // Use debouncedSearchValue for filtering to improve performance
    const filteredLogs = useMemo(() => {
        const search = debouncedSearchValue.toLowerCase().trim();

        if (!search || !logs) {
            return logs || [];
        }

        return logs.filter(
            (log) =>
                log.message.toLowerCase().includes(search) ||
                (log.result && log.result.toLowerCase().includes(search)) ||
                (log.thinking && log.thinking.toLowerCase().includes(search)),
        );
    }, [logs, debouncedSearchValue]);

    // Get placeholder text based on flow status
    const placeholder = useMemo(() => {
        if (!flowId) {
            return 'Select a flow...';
        }

        // Flow-specific statuses
        switch (flowStatus) {
            case StatusType.Created: {
                return 'The flow is starting...';
            }

            case StatusType.Failed:
            case StatusType.Finished: {
                return 'This flow has ended. Create a new one to continue.';
            }

            case StatusType.Running: {
                return 'PentAGI is working... Click Stop to interrupt';
            }

            case StatusType.Waiting: {
                return 'Provide additional context or instructions...';
            }

            default: {
                return 'Type your message...';
            }
        }
    }, [flowId, flowStatus]);

    // Message submission handler
    const handleSubmitMessage = async (values: FlowFormValues) => {
        setIsSubmitting(true);

        try {
            await submitAutomationMessage(values);
        } finally {
            setIsSubmitting(false);
        }
    };

    // Stop automation handler
    const handleStopAutomation = async () => {
        setIsCanceling(true);

        try {
            await stopAutomation();
        } finally {
            setIsCanceling(false);
        }
    };

    const isFormDisabled = flowStatus === StatusType.Finished || flowStatus === StatusType.Failed;
    const isFormLoading = flowStatus === StatusType.Created || flowStatus === StatusType.Running;

    return (
        <div className={cn('flex h-full flex-col', className)}>
            <div className="sticky top-0 z-10 bg-background pb-4">
                <Form {...form}>
                    <FormField
                        control={form.control}
                        name="search"
                        render={({ field }) => (
                            <FormControl>
                                <div className="relative p-px">
                                    <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                                    <Input
                                        {...field}
                                        autoComplete="off"
                                        className="px-9"
                                        placeholder="Search messages..."
                                        type="text"
                                    />
                                    {field.value && (
                                        <Button
                                            className="absolute right-0 top-1/2 -translate-y-1/2"
                                            onClick={() => {
                                                form.reset({ search: '' });
                                                setDebouncedSearchValue('');
                                                debouncedUpdateSearch.cancel();
                                            }}
                                            size="icon"
                                            type="button"
                                            variant="ghost"
                                        >
                                            <X />
                                        </Button>
                                    )}
                                </div>
                            </FormControl>
                        )}
                    />
                </Form>
            </div>

            {filteredLogs.length > 0 ? (
                <div className="relative h-full overflow-y-hidden pb-4">
                    <div
                        className="h-full space-y-4 overflow-y-auto"
                        ref={containerRef}
                    >
                        {filteredLogs.map((log) => (
                            <FlowMessage
                                key={log.id}
                                log={log}
                                searchValue={debouncedSearchValue}
                            />
                        ))}
                        <div ref={endRef} />
                    </div>

                    {hasNewMessages && !isScrolledToBottom && (
                        <Button
                            className="absolute bottom-4 right-4 z-10 size-9 rounded-full shadow-md hover:shadow-lg"
                            onClick={() => scrollToEnd()}
                            size="icon"
                            type="button"
                        >
                            <ChevronDown />
                        </Button>
                    )}
                </div>
            ) : (
                <Empty>
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <Inbox />
                        </EmptyMedia>
                        <EmptyTitle>No Active Tasks</EmptyTitle>
                        <EmptyDescription>
                            Starting a new task may take some time as the PentAGI agent downloads the required Docker
                            image
                        </EmptyDescription>
                    </EmptyHeader>
                </Empty>
            )}

            <div className="sticky bottom-0 border-t bg-background p-px pt-4">
                <FlowForm
                    defaultValues={{
                        providerName: flowData?.flow?.provider?.name ?? '',
                    }}
                    isCanceling={isCanceling}
                    isDisabled={isFormDisabled}
                    isLoading={isFormLoading}
                    isProviderDisabled={true}
                    isSubmitting={isSubmitting}
                    onCancel={handleStopAutomation}
                    onSubmit={handleSubmitMessage}
                    placeholder={placeholder}
                    type={'automation'}
                />
            </div>
        </div>
    );
};

export default FlowAutomationMessages;
