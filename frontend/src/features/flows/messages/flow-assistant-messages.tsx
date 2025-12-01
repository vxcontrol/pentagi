import { zodResolver } from '@hookform/resolvers/zod';
import debounce from 'lodash/debounce';
import { Check, ChevronDown, ChevronsUpDown, Loader2, Plus, Search, Trash2, X, XCircle } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import type { AssistantFragmentFragment, ProviderFragmentFragment } from '@/graphql/types';

import { ProviderIcon } from '@/components/icons/provider-icon';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Button } from '@/components/ui/button';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { StatusType } from '@/graphql/types';
import { useChatScroll } from '@/hooks/use-chat-scroll';
import { Log } from '@/lib/log';
import { cn } from '@/lib/utils';
import { isProviderValid } from '@/models/provider';
import { useFlow } from '@/providers/flow-provider';
import { useProviders } from '@/providers/providers-provider';
import { useSystemSettings } from '@/providers/system-settings-provider';

import { FlowForm, type FlowFormValues } from '../flow-form';
import FlowMessage from './flow-message';

interface AssistantDropdownProps {
    assistants: AssistantFragmentFragment[];
    isAssistantCreating: boolean;
    onAssistantCreate: () => void;
    onAssistantDelete: (assistantId: string) => void;
    onAssistantSelect: (assistantId: string) => void;
    providers: ProviderFragmentFragment[];
    selectedAssistantId: null | string;
}

const AssistantDropdown = ({
    assistants,
    isAssistantCreating,
    onAssistantCreate,
    onAssistantDelete,
    onAssistantSelect,
    providers,
    selectedAssistantId,
}: AssistantDropdownProps) => {
    const [open, setOpen] = useState(false);
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [assistantToDelete, setAssistantToDelete] = useState<AssistantFragmentFragment | null>(null);

    // Get the current selected assistant
    const selectedAssistant = useMemo(() => {
        if (!selectedAssistantId) {
            return null;
        }

        return assistants.find((assistant) => assistant.id === selectedAssistantId) || null;
    }, [assistants, selectedAssistantId]);

    // Get the current selected assistant index (1-based, reversed)
    const selectedAssistantIndex = useMemo(() => {
        if (!selectedAssistantId) {
            return null;
        }

        const index = assistants.findIndex((assistant) => assistant.id === selectedAssistantId);

        return index !== -1 ? assistants.length - index : null;
    }, [assistants, selectedAssistantId]);

    // Group assistants by status
    const groupedAssistants = useMemo(() => {
        const active: AssistantFragmentFragment[] = [];
        const finished: AssistantFragmentFragment[] = [];
        const failed: AssistantFragmentFragment[] = [];

        assistants.forEach((assistant) => {
            if (assistant.status === StatusType.Running || assistant.status === StatusType.Waiting) {
                active.push(assistant);
            } else if (assistant.status === StatusType.Finished) {
                finished.push(assistant);
            } else if (assistant.status === StatusType.Failed) {
                failed.push(assistant);
            }
        });

        return { active, failed, finished };
    }, [assistants]);

    // Get status indicator
    const getStatusIndicator = (status: StatusType) => {
        switch (status) {
            case StatusType.Failed: {
                return <XCircle className="size-3 text-red-500" />;
            }

            case StatusType.Finished: {
                return <Check className="size-3 text-gray-500" />;
            }

            case StatusType.Running: {
                return <Loader2 className="size-3 animate-spin text-blue-500" />;
            }

            case StatusType.Waiting: {
                return <div className="size-2 rounded-full bg-green-500" />;
            }

            default: {
                return null;
            }
        }
    };

    // Handle assistant selection
    const handleAssistantSelect = (assistantId: string) => {
        onAssistantSelect(assistantId);
        setOpen(false);
    };

    // Handle delete click
    const handleDeleteClick = (assistant: AssistantFragmentFragment, event: React.MouseEvent) => {
        event.stopPropagation();
        setAssistantToDelete(assistant);
        setDeleteDialogOpen(true);
    };

    // Confirm delete
    const handleConfirmDelete = () => {
        if (assistantToDelete) {
            onAssistantDelete(assistantToDelete.id);
            setAssistantToDelete(null);
        }
    };

    // Render assistant item
    const renderAssistantItem = (assistant: AssistantFragmentFragment, index: number) => {
        const isSelected = selectedAssistantId === assistant.id;
        const isValid = isProviderValid(assistant.provider, providers);
        const displayNumber = assistants.length - index;

        return (
            <CommandItem
                className={cn('group flex items-center justify-between gap-2', !isValid && 'opacity-50')}
                key={assistant.id}
                onSelect={() => handleAssistantSelect(assistant.id)}
                value={`${assistant.id}-${assistant.title}`}
            >
                <div className="flex flex-1 items-center gap-2 overflow-hidden">
                    {getStatusIndicator(assistant.status)}
                    <span className="flex size-5 shrink-0 items-center justify-center rounded bg-muted text-xs font-medium text-muted-foreground">
                        {displayNumber}
                    </span>
                    <ProviderIcon
                        className="size-4 shrink-0"
                        provider={assistant.provider}
                    />
                    <div className="flex flex-1 flex-col overflow-hidden">
                        <div className="flex items-center gap-2">
                            <span className="truncate text-sm">{assistant.title}</span>
                            {!isValid && <span className="shrink-0 text-xs text-destructive">(unavailable)</span>}
                        </div>
                    </div>
                    {isSelected && <Check className="ml-auto size-4 shrink-0 text-primary" />}
                </div>
                <Button
                    className="size-6 shrink-0 opacity-0 transition-opacity group-hover:opacity-100"
                    onClick={(e) => handleDeleteClick(assistant, e)}
                    size="icon"
                    variant="ghost"
                >
                    <Trash2 className="size-3.5 text-muted-foreground hover:text-destructive" />
                </Button>
            </CommandItem>
        );
    };

    return (
        <>
            <Popover
                onOpenChange={setOpen}
                open={open}
            >
                <PopoverTrigger asChild>
                    <Button
                        aria-expanded={open}
                        aria-label="Select Assistant"
                        className="w-full justify-between"
                        disabled={isAssistantCreating}
                        role="combobox"
                        variant="outline"
                    >
                        <div className="flex min-w-0 flex-1 items-center gap-2">
                            {isAssistantCreating ? (
                                <>
                                    <Loader2 className="size-4 shrink-0 animate-spin" />
                                    <span className="truncate">Creating assistant...</span>
                                </>
                            ) : selectedAssistant ? (
                                <>
                                    {getStatusIndicator(selectedAssistant.status)}
                                    <span className="flex size-5 shrink-0 items-center justify-center rounded bg-muted text-xs font-medium text-muted-foreground">
                                        {selectedAssistantIndex}
                                    </span>
                                    <ProviderIcon
                                        className="size-4 shrink-0"
                                        provider={selectedAssistant.provider}
                                    />
                                    <span className="truncate">{selectedAssistant.title}</span>
                                </>
                            ) : (
                                <span className="text-muted-foreground">Select assistant...</span>
                            )}
                        </div>
                        <ChevronsUpDown className="ml-2 size-4 shrink-0 opacity-50" />
                    </Button>
                </PopoverTrigger>
                <PopoverContent
                    align="start"
                    className="w-[400px] p-0"
                >
                    <Command>
                        <CommandInput placeholder="Search assistants..." />
                        <CommandList>
                            <CommandEmpty>No assistants found.</CommandEmpty>

                            <CommandGroup>
                                <CommandItem
                                    className="flex items-center gap-2 font-medium"
                                    onSelect={() => {
                                        onAssistantCreate();
                                        setOpen(false);
                                    }}
                                    value="create-new-assistant"
                                >
                                    <Plus className="size-4" />
                                    Create new assistant
                                </CommandItem>
                            </CommandGroup>

                            {groupedAssistants.active.length > 0 && (
                                <CommandGroup heading={`Active (${groupedAssistants.active.length})`}>
                                    {groupedAssistants.active.map((assistant) => {
                                        const globalIndex = assistants.findIndex((a) => a.id === assistant.id);

                                        return renderAssistantItem(assistant, globalIndex);
                                    })}
                                </CommandGroup>
                            )}

                            {groupedAssistants.finished.length > 0 && (
                                <CommandGroup heading={`Finished (${groupedAssistants.finished.length})`}>
                                    {groupedAssistants.finished.map((assistant) => {
                                        const globalIndex = assistants.findIndex((a) => a.id === assistant.id);

                                        return renderAssistantItem(assistant, globalIndex);
                                    })}
                                </CommandGroup>
                            )}

                            {groupedAssistants.failed.length > 0 && (
                                <CommandGroup heading={`Failed (${groupedAssistants.failed.length})`}>
                                    {groupedAssistants.failed.map((assistant) => {
                                        const globalIndex = assistants.findIndex((a) => a.id === assistant.id);

                                        return renderAssistantItem(assistant, globalIndex);
                                    })}
                                </CommandGroup>
                            )}
                        </CommandList>
                    </Command>
                </PopoverContent>
            </Popover>

            <ConfirmationDialog
                cancelText="Cancel"
                confirmText="Delete"
                handleConfirm={handleConfirmDelete}
                handleOpenChange={setDeleteDialogOpen}
                isOpen={deleteDialogOpen}
                itemName={assistantToDelete?.title}
                itemType="assistant"
                title="Delete Assistant"
            />
        </>
    );
};

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
        createAssistant,
        deleteAssistant,
        flowId,
        flowStatus,
        initiateAssistantCreation,
        selectAssistant,
        selectedAssistantId,
        stopAssistant,
        submitAssistantMessage,
    } = useFlow();

    const [isAssistantCreating, setIsAssistantCreating] = useState(false);

    // Separate state for immediate input value and debounced search value
    const [debouncedSearchValue, setDebouncedSearchValue] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [isCanceling, setIsCanceling] = useState(false);

    const { containerRef, endRef, hasNewMessages, isScrolledToBottom, scrollToEnd } = useChatScroll(
        logs ?? [],
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

    // Calculate default useAgents value
    const shouldUseAgents = useMemo(() => {
        // If creating a new assistant, use system setting
        if (isAssistantCreating || !selectedAssistant) {
            return settings?.assistantUseAgents ?? false;
        }

        // If assistant is selected and not creating new, use its useAgents setting
        return selectedAssistant.useAgents;
    }, [selectedAssistant, settings?.assistantUseAgents, isAssistantCreating]);

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
    const handleAssistantDelete = (assistantId: string) => {
        if (deleteAssistant) {
            deleteAssistant(assistantId);
        }
    };

    // Message submission handler
    const handleSubmitMessage = async (values: FlowFormValues) => {
        if (!values.message.trim()) {
            return;
        }

        setIsSubmitting(true);

        try {
            if (!selectedAssistantId) {
                // If no assistant is selected, create a new one
                setIsAssistantCreating(true);

                if (createAssistant) {
                    await createAssistant(values);
                }
            } else if (submitAssistantMessage) {
                // Otherwise call the existing assistant
                await submitAssistantMessage(selectedAssistantId, values);
            }
        } catch (error) {
            Log.error('Error submitting message:', error);
            throw error;
        } finally {
            setIsSubmitting(false);
            setIsAssistantCreating(false);
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
    const handleAssistantCreate = () => {
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
        if (isAssistantCreating) {
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
    }, [flowId, isAssistantCreating, selectedAssistant?.status]);

    const assistantStatus = selectedAssistant?.status;
    const isFormDisabled =
        flowStatus === StatusType.Finished ||
        flowStatus === StatusType.Failed ||
        assistantStatus === StatusType.Finished ||
        assistantStatus === StatusType.Failed;
    const isFormLoading = assistantStatus === StatusType.Created || assistantStatus === StatusType.Running;

    return (
        <div className={cn('flex h-full flex-col', className)}>
            <div className="sticky top-0 z-10 bg-background pb-4">
                <div className="flex flex-col gap-2 p-px">
                    {/* Assistant Dropdown */}
                    {flowId && (
                        <AssistantDropdown
                            assistants={assistants}
                            isAssistantCreating={isAssistantCreating}
                            onAssistantCreate={handleAssistantCreate}
                            onAssistantDelete={handleAssistantDelete}
                            onAssistantSelect={(assistantId) => selectAssistant?.(assistantId)}
                            providers={providers}
                            selectedAssistantId={selectedAssistantId}
                        />
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
                                                disabled={isAssistantCreating}
                                                placeholder="Search messages..."
                                                type="text"
                                            />
                                            {field.value && (
                                                <Button
                                                    className="absolute right-0 top-1/2 -translate-y-1/2"
                                                    disabled={isAssistantCreating}
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

            {isAssistantCreating ? (
                <Empty>
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <Loader2 className="animate-spin" />
                        </EmptyMedia>
                        <EmptyTitle>Creating assistant...</EmptyTitle>
                        <EmptyDescription>Please wait while we set up your new assistant</EmptyDescription>
                    </EmptyHeader>
                </Empty>
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
                <Empty>
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <Plus />
                        </EmptyMedia>
                        <EmptyTitle>New assistant</EmptyTitle>
                        <EmptyDescription>Type a message below to create a new assistant...</EmptyDescription>
                    </EmptyHeader>
                </Empty>
            )}

            <div className="sticky bottom-0 border-t bg-background p-px pt-4">
                <FlowForm
                    defaultValues={{
                        providerName: selectedAssistant?.provider?.name ?? '',
                        useAgents: shouldUseAgents,
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
