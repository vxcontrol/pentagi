import { zodResolver } from '@hookform/resolvers/zod';
import debounce from 'lodash/debounce';
import { ChevronDown, Inbox, ListFilter, Search, X } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { StatusType } from '@/graphql/types';
import { useChatScroll } from '@/hooks/use-chat-scroll';
import { cn } from '@/lib/utils';
import { useFlow } from '@/providers/flow-provider';

import { FlowForm, type FlowFormValues } from '../flow-form';
import FlowTasksDropdown from '../flow-tasks-dropdown';
import FlowMessage from './flow-message';

interface FlowAutomationMessagesProps {
    className?: string;
}

const searchFormSchema = z.object({
    filter: z
        .object({
            subtaskIds: z.array(z.string()),
            taskIds: z.array(z.string()),
        })
        .optional(),
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
            filter: {
                subtaskIds: [],
                taskIds: [],
            },
            search: '',
        },
        resolver: zodResolver(searchFormSchema),
    });

    const searchValue = form.watch('search');
    const filter = form.watch('filter');

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
        form.reset({
            filter: {
                subtaskIds: [],
                taskIds: [],
            },
            search: '',
        });
        setDebouncedSearchValue('');
        debouncedUpdateSearch.cancel();
    }, [flowId, form, debouncedUpdateSearch]);

    // Check if any filters are active
    const hasActiveFilters = useMemo(() => {
        const hasSearch = !!searchValue.trim();
        const hasTaskFilters = !!(filter?.taskIds?.length || filter?.subtaskIds?.length);

        return hasSearch || hasTaskFilters;
    }, [searchValue, filter]);

    // Memoize filtered logs to avoid recomputing on every render
    // Use debouncedSearchValue for filtering to improve performance
    const filteredLogs = useMemo(() => {
        const search = debouncedSearchValue.toLowerCase().trim();

        let filtered = logs || [];

        // Filter by search
        if (search) {
            filtered = filtered.filter(
                (log) =>
                    log.message.toLowerCase().includes(search) ||
                    (log.result && log.result.toLowerCase().includes(search)) ||
                    (log.thinking && log.thinking.toLowerCase().includes(search)),
            );
        }

        // Filter by selected tasks and subtasks
        if (filter?.taskIds?.length || filter?.subtaskIds?.length) {
            const selectedTaskIds = new Set(filter.taskIds ?? []);
            const selectedSubtaskIds = new Set(filter.subtaskIds ?? []);

            filtered = filtered.filter((log) => {
                if (log.taskId && selectedTaskIds.has(log.taskId)) {
                    return true;
                }

                if (log.subtaskId && selectedSubtaskIds.has(log.subtaskId)) {
                    return true;
                }

                return false;
            });
        }

        return filtered;
    }, [logs, debouncedSearchValue, filter]);

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

    // Reset filters handler
    const handleResetFilters = () => {
        form.reset({
            filter: {
                subtaskIds: [],
                taskIds: [],
            },
            search: '',
        });
        setDebouncedSearchValue('');
        debouncedUpdateSearch.cancel();
    };

    const isFormDisabled = flowStatus === StatusType.Finished || flowStatus === StatusType.Failed;
    const isFormLoading = flowStatus === StatusType.Created || flowStatus === StatusType.Running;

    return (
        <div className={cn('flex h-full flex-col', className)}>
            <div className="bg-background sticky top-0 z-10 pb-4">
                <Form {...form}>
                    <div className="flex gap-2 p-px">
                        <FormField
                            control={form.control}
                            name="search"
                            render={({ field }) => (
                                <FormControl>
                                    <InputGroup className="flex-1">
                                        <InputGroupAddon>
                                            <Search />
                                        </InputGroupAddon>
                                        <InputGroupInput
                                            {...field}
                                            autoComplete="off"
                                            placeholder="Search messages..."
                                            type="text"
                                        />
                                        {field.value && (
                                            <InputGroupAddon align="inline-end">
                                                <InputGroupButton
                                                    onClick={() => {
                                                        form.reset({ search: '' });
                                                        setDebouncedSearchValue('');
                                                        debouncedUpdateSearch.cancel();
                                                    }}
                                                    type="button"
                                                >
                                                    <X />
                                                </InputGroupButton>
                                            </InputGroupAddon>
                                        )}
                                    </InputGroup>
                                </FormControl>
                            )}
                        />
                        <FormField
                            control={form.control}
                            name="filter"
                            render={({ field }) => (
                                <FormControl>
                                    <FlowTasksDropdown
                                        onChange={field.onChange}
                                        value={field.value}
                                    />
                                </FormControl>
                            )}
                        />
                    </div>
                </Form>
            </div>

            {filteredLogs.length > 0 ? (
                <div className="relative h-full overflow-y-hidden">
                    <div
                        className="flex h-full flex-col gap-4 overflow-y-auto"
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
                            className="absolute right-4 bottom-4 z-10 size-9 rounded-full shadow-md hover:shadow-lg"
                            onClick={() => scrollToEnd()}
                            size="icon"
                            type="button"
                        >
                            <ChevronDown />
                        </Button>
                    )}
                </div>
            ) : hasActiveFilters ? (
                <Empty>
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <ListFilter />
                        </EmptyMedia>
                        <EmptyTitle>No messages found</EmptyTitle>
                        <EmptyDescription>Try adjusting your search or filter parameters</EmptyDescription>
                    </EmptyHeader>
                    <EmptyContent>
                        <Button
                            onClick={handleResetFilters}
                            variant="outline"
                        >
                            <X />
                            Reset filters
                        </Button>
                    </EmptyContent>
                </Empty>
            ) : (
                <Empty>
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <Inbox />
                        </EmptyMedia>
                        <EmptyTitle>No active tasks</EmptyTitle>
                        <EmptyDescription>
                            Starting a new task may take some time as the PentAGI agent downloads the required Docker
                            image
                        </EmptyDescription>
                    </EmptyHeader>
                </Empty>
            )}

            <div className="bg-background sticky bottom-0 p-px">
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
