import { zodResolver } from '@hookform/resolvers/zod';
import debounce from 'lodash/debounce';
import { ChevronDown, Loader2, Plus, Search, Trash2, X } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { ProviderIcon } from '@/components/icons/ProviderIcon';
import { Button } from '@/components/ui/button';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { StatusType } from '@/graphql/types';
import { useChatScroll } from '@/hooks/use-chat-scroll';
import { Log } from '@/lib/log';
import { cn } from '@/lib/utils';
import { isProviderValid } from '@/models/Provider';
import { useFlow } from '@/providers/FlowProvider';
import { useProviders } from '@/providers/ProvidersProvider';
import { useSystemSettings } from '@/providers/SystemSettingsProvider';

import { FlowForm, type FlowFormValues } from '../FlowForm';
import FlowMessage from './FlowMessage';

interface FlowAssistantMessagesProps {
    className?: string;
}

const searchFormSchema = z.object({
    search: z.string(),
});

const FlowAssistantMessages = ({ className }: FlowAssistantMessagesProps) => {
    const { providers } = useProviders();

    const {
        assistantLogs: logs,
        assistants,
        callAssistant,
        createAssistant,
        deleteAssistant,
        flowId,
        initiateAssistantCreation,
        selectAssistant,
        selectedAssistantId,
        stopAssistant,
    } = useFlow();

    const [isCreatingAssistant, setIsCreatingAssistant] = useState(false);

    // Separate state for immediate input value and debounced search value
    const [debouncedSearchValue, setDebouncedSearchValue] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [isCanceling, setIsCanceling] = useState(false);

    const { containerRef, endRef, hasNewMessages, isScrolledToBottom, scrollToEnd } = useChatScroll(
        useMemo(() => (logs ? [...(logs || [])] : []), [logs]),
        selectedAssistantId ?? null,
    );

    // Get system settings
    const { settings } = useSystemSettings();

    const form = useForm<z.infer<typeof searchFormSchema>>({
        defaultValues: {
            search: '',
        },
        resolver: zodResolver(searchFormSchema),
    });

    const searchValue = form.watch('search');

    // Create debounced function to update search value
    const debouncedUpdateSearch = useMemo(
        () =>
            debounce((value: string) => {
                setDebouncedSearchValue(value);
            }, 500),
        [],
    );

    // Update debounced search value when input value changes
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

    // Get the current selected assistant
    const selectedAssistant = useMemo(() => {
        if (!selectedAssistantId || !assistants) {
            return null;
        }

        return assistants.find((assistant) => assistant.id === selectedAssistantId) || null;
    }, [assistants, selectedAssistantId]);

    // Get the current selected assistant index (1-based, reversed)
    const selectedAssistantIndex = useMemo(() => {
        if (!selectedAssistantId || !assistants) {
            return null;
        }

        const index = assistants.findIndex((assistant) => assistant.id === selectedAssistantId);

        return index !== -1 ? assistants.length - index : null;
    }, [assistants, selectedAssistantId]);

    // Calculate default useAgents value
    const isUseAgentsDefault = useMemo(() => {
        // If creating a new assistant, use system setting
        if (isCreatingAssistant || !selectedAssistant) {
            return settings?.assistantUseAgents ?? false;
        }

        // If assistant is selected and not creating new, use its useAgents setting
        return selectedAssistant.useAgents;
    }, [selectedAssistant, settings?.assistantUseAgents, isCreatingAssistant]);

    // Memoize filtered logs to avoid recomputing on every render
    // Use debouncedSearchValue for filtering to improve performance
    const filteredLogs = useMemo(() => {
        if (!logs) {
            return [];
        }

        // First filter by selected assistant
        let assistantFilteredLogs = logs;

        if (selectedAssistantId) {
            assistantFilteredLogs = logs.filter((log) => log.assistantId === selectedAssistantId);
        } else {
            // If no assistant is selected, show no logs
            assistantFilteredLogs = [];
        }

        // Then filter by search query if present
        const search = debouncedSearchValue.toLowerCase().trim();

        if (!search) {
            return assistantFilteredLogs;
        }

        return assistantFilteredLogs.filter(
            (log) =>
                log.message.toLowerCase().includes(search) ||
                (log.result && log.result.toLowerCase().includes(search)) ||
                (log.thinking && log.thinking.toLowerCase().includes(search)),
        );
    }, [logs, debouncedSearchValue, selectedAssistantId]);

    // Handlers for interacting with assistant
    const handleDeleteAssistant = (assistantId: string) => {
        if (deleteAssistant) {
            deleteAssistant(assistantId);
        }
    };

    // Message submission handler
    const handleSubmitMessage = async (values: FlowFormValues) => {
        const { message, useAgents } = values;

        if (!message.trim()) {
            return;
        }

        setIsSubmitting(true);

        try {
            if (!selectedAssistantId) {
                // If no assistant is selected, create a new one
                setIsCreatingAssistant(true);

                if (createAssistant) {
                    await createAssistant(message, useAgents);
                }
            } else if (callAssistant) {
                // Otherwise call the existing assistant
                await callAssistant(selectedAssistantId, message, useAgents);
            }
        } catch (error) {
            Log.error('Error submitting message:', error);
            throw error;
        } finally {
            setIsSubmitting(false);
            setIsCreatingAssistant(false);
        }
    };

    // Stop assistant handler
    const handleStopAssistant = async () => {
        if (!selectedAssistantId || !stopAssistant) {
            return;
        }

        setIsCanceling(true);

        try {
            await stopAssistant(selectedAssistantId);
        } catch (error) {
            Log.error('Error stopping assistant:', error);
            throw error;
        } finally {
            setIsCanceling(false);
        }
    };

    // Handle click on Create Assistant option in dropdown
    const handleCreateAssistantClick = () => {
        if (initiateAssistantCreation) {
            initiateAssistantCreation();
        }
    };

    // Get placeholder text based on assistant status
    const placeholder = useMemo(() => {
        if (!flowId) {
            return 'Select a flow...';
        }

        // Show creating assistant message while in creation mode
        if (isCreatingAssistant) {
            return 'Creating assistant...';
        }

        // No assistant selected - prompt to create one
        if (!selectedAssistant?.status) {
            return 'Type a message to create a new assistant...';
        }

        // Assistant-specific statuses
        switch (selectedAssistant.status) {
            case StatusType.Created: {
                return 'Assistant is starting...';
            }

            case StatusType.Failed:
            case StatusType.Finished: {
                return 'This assistant session has ended. Create a new one to continue.';
            }

            case StatusType.Running: {
                return 'Assistant is running... Click Stop to interrupt';
            }

            case StatusType.Waiting: {
                return 'Continue the conversation...';
            }

            default: {
                return 'Type your message...';
            }
        }
    }, [flowId, isCreatingAssistant, selectedAssistant?.status]);

    const assistantStatus = selectedAssistant?.status;
    const isFormDisabled = assistantStatus === StatusType.Finished || assistantStatus === StatusType.Failed;
    const isFormLoading = assistantStatus === StatusType.Created || assistantStatus === StatusType.Running;

    return (
        <div className={cn('flex h-full flex-col', className)}>
            <div className="sticky top-0 z-10 bg-background pb-4">
                <div className="flex gap-2 p-px">
                    {/* Assistant Selector Dropdown */}
                    {flowId && (
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    aria-label="Select Assistant"
                                    className="relative w-12 shrink-0"
                                    disabled={isCreatingAssistant}
                                    onFocus={(e) => e.preventDefault()}
                                    size="icon"
                                    tabIndex={-1}
                                    variant="outline"
                                >
                                    {isCreatingAssistant ? (
                                        <Loader2 className="size-4 animate-spin" />
                                    ) : (
                                        <div className="flex items-center gap-1">
                                            {selectedAssistantIndex && (
                                                <span className="flex size-4 items-center justify-center rounded bg-muted text-[10px] font-medium text-muted-foreground">
                                                    {selectedAssistantIndex}
                                                </span>
                                            )}
                                            <ChevronDown className="size-4" />
                                        </div>
                                    )}
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                                align="start"
                                className="max-h-[300px] w-[100] overflow-y-auto"
                            >
                                <DropdownMenuItem
                                    className="flex items-center gap-2 font-medium"
                                    onClick={() => handleCreateAssistantClick()}
                                    tabIndex={-1}
                                >
                                    <Plus className="size-4" />
                                    Create new assistant
                                </DropdownMenuItem>
                                {assistants.length > 0 && <DropdownMenuSeparator />}
                                {assistants.map((assistant, index) => (
                                    <DropdownMenuItem
                                        className={cn(
                                            'flex items-center justify-between gap-2',
                                            selectedAssistantId === assistant.id && 'bg-accent font-medium',
                                            !isProviderValid(assistant.provider, providers) && 'opacity-50',
                                        )}
                                        key={assistant.id}
                                        tabIndex={-1}
                                    >
                                        <div
                                            className="flex grow cursor-pointer items-center gap-2"
                                            onClick={() => {
                                                selectAssistant?.(assistant.id);
                                            }}
                                        >
                                            <span className="flex size-5 shrink-0 items-center justify-center rounded bg-muted text-xs font-medium text-muted-foreground">
                                                {assistants.length - index}
                                            </span>
                                            <ProviderIcon provider={assistant.provider} />
                                            <div className="flex-1">
                                                <div className="flex items-center gap-2">
                                                    <span className="text-sm">{assistant.title}</span>
                                                    {!isProviderValid(assistant.provider, providers) && (
                                                        <span className="text-xs text-destructive">(unavailable)</span>
                                                    )}
                                                </div>
                                            </div>
                                        </div>
                                        <Button
                                            className="ml-auto size-6"
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                handleDeleteAssistant(assistant.id);
                                            }}
                                            size="icon"
                                            variant="ghost"
                                        >
                                            <Trash2 className="size-3.5 text-muted-foreground hover:text-destructive" />
                                        </Button>
                                    </DropdownMenuItem>
                                ))}
                            </DropdownMenuContent>
                        </DropdownMenu>
                    )}
                    {/* Search Input */}
                    <div className="flex-1">
                        <Form {...form}>
                            <FormField
                                control={form.control}
                                name="search"
                                render={({ field }) => (
                                    <FormControl>
                                        <div className="relative">
                                            <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                                            <Input
                                                {...field}
                                                autoComplete="off"
                                                className="px-9"
                                                disabled={isCreatingAssistant}
                                                placeholder="Search messages..."
                                                type="text"
                                            />
                                            {field.value && (
                                                <Button
                                                    className="absolute right-0 top-1/2 -translate-y-1/2"
                                                    disabled={isCreatingAssistant}
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
                </div>
            </div>

            {isCreatingAssistant ? (
                <div className="flex flex-1 items-center justify-center">
                    <div className="flex flex-col items-center gap-4 text-muted-foreground">
                        <Loader2 className="size-12 animate-spin" />
                        <p>Creating assistant...</p>
                    </div>
                </div>
            ) : selectedAssistantId ? (
                // Show messages for selected assistant
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
                // Show placeholder when no assistant is selected
                <div className="flex flex-1 items-center justify-center">
                    <div className="flex flex-col items-center gap-2 text-muted-foreground">
                        <p>No Assistant Selected</p>
                        <p className="text-xs">Select an assistant from the dropdown or create a new one</p>
                    </div>
                </div>
            )}

            <div className="sticky bottom-0 border-t bg-background p-px pt-4">
                <FlowForm
                    defaultValues={{
                        providerName: selectedAssistant?.provider?.name ?? '',
                        useAgents: isUseAgentsDefault,
                    }}
                    isCanceling={isCanceling}
                    isDisabled={isFormDisabled}
                    isLoading={isFormLoading}
                    isProviderDisabled={!!selectedAssistant}
                    isSubmitting={isSubmitting}
                    onCancel={handleStopAssistant}
                    onSubmit={handleSubmitMessage}
                    placeholder={placeholder}
                    type={'assistant'}
                />
            </div>
        </div>
    );
};

export default FlowAssistantMessages;
