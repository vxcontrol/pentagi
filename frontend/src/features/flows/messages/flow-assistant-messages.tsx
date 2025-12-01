import { zodResolver } from '@hookform/resolvers/zod';
import debounce from 'lodash/debounce';
import { Check, ChevronDown, ListFilter, Loader2, Plus, Search, Trash2, X } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import type { AssistantFragmentFragment, ProviderFragmentFragment } from '@/graphql/types';

import { FlowStatusIcon } from '@/components/icons/flow-status-icon';
import { ProviderIcon } from '@/components/icons/provider-icon';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Button } from '@/components/ui/button';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { StatusType } from '@/graphql/types';
import { useChatScroll } from '@/hooks/use-chat-scroll';
import { Log } from '@/lib/log';
import { cn } from '@/lib/utils';
import { formatName } from '@/lib/utils/format';
import { isProviderValid } from '@/models/provider';
import { useFlow } from '@/providers/flow-provider';
import { useProviders } from '@/providers/providers-provider';
import { useSystemSettings } from '@/providers/system-settings-provider';

import { FlowForm, type FlowFormValues } from '../flow-form';
import FlowMessage from './flow-message';

interface AssistantsDropdownProps {
    assistants: AssistantFragmentFragment[];
    isAssistantCreating: boolean;
    isDisabled: boolean;
    onAssistantCreate: () => void;
    onAssistantDelete: (assistantId: string) => void;
    onAssistantSelect: (assistantId: string) => void;
    providers: ProviderFragmentFragment[];
    selectedAssistantId: null | string;
}

const AssistantsDropdown = ({
    assistants,
    isAssistantCreating,
    isDisabled,
    onAssistantCreate,
    onAssistantDelete,
    onAssistantSelect,
    providers,
    selectedAssistantId,
}: AssistantsDropdownProps) => {
    const [isOpen, setIsOpen] = useState(false);
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [currentAssistant, setCurrentAssistant] = useState<AssistantFragmentFragment | null>(null);

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

    // Group assistants by status with pre-calculated indices
    const assistantsGroup = useMemo(() => {
        type AssistantItem = { assistant: AssistantFragmentFragment; index: number };

        return assistants.reduce<{
            active: AssistantItem[];
            failed: AssistantItem[];
            finished: AssistantItem[];
        }>(
            (accumulator, assistant, originalIndex) => {
                const index = assistants.length - originalIndex;
                const item = { assistant, index };

                return {
                    ...accumulator,
                    active:
                        assistant.status === StatusType.Running || assistant.status === StatusType.Waiting
                            ? [...accumulator.active, item]
                            : accumulator.active,
                    failed: assistant.status === StatusType.Failed ? [...accumulator.failed, item] : accumulator.failed,
                    finished:
                        assistant.status === StatusType.Finished
                            ? [...accumulator.finished, item]
                            : accumulator.finished,
                };
            },
            { active: [], failed: [], finished: [] },
        );
    }, [assistants]);

    // Handle assistant selection
    const handleAssistantSelect = (assistantId: string) => {
        onAssistantSelect(assistantId);
        setIsOpen(false);
    };

    // Handle delete click
    const handleDeleteClick = (assistant: AssistantFragmentFragment) => {
        if (isDisabled) {
            return;
        }

        setCurrentAssistant(assistant);
        setDeleteDialogOpen(true);
    };

    // Confirm delete
    const handleConfirmDelete = () => {
        if (currentAssistant) {
            onAssistantDelete(currentAssistant.id);
            setCurrentAssistant(null);
        }
    };

    // Render assistant item
    const renderAssistantItem = (assistant: AssistantFragmentFragment, index: number) => {
        const isSelected = selectedAssistantId === assistant.id;
        const isValid = isProviderValid(assistant.provider, providers);

        return (
            <CommandItem
                className={cn('group', !isValid && 'opacity-50')}
                key={assistant.id}
                onSelect={() => handleAssistantSelect(assistant.id)}
                value={`${assistant.id}-${assistant.title}`}
            >
                <FlowStatusIcon
                    status={assistant.status}
                    tooltip={formatName(assistant.status)}
                />

                <ProviderIcon
                    className="shrink-0"
                    provider={assistant.provider}
                />

                <span className="flex size-5 shrink-0 items-center justify-center rounded bg-muted text-xs font-medium text-muted-foreground">
                    {index}
                </span>

                <div className="flex flex-1 items-center gap-2 overflow-hidden">
                    <span className="truncate text-sm">{assistant.title}</span>
                    {!isValid && <span className="shrink-0 text-xs text-destructive">(unavailable)</span>}
                </div>

                <Check
                    className={cn(
                        'ml-auto size-4 shrink-0 text-primary transition-opacity group-hover:opacity-0',
                        isSelected ? 'opacity-100' : 'opacity-0',
                    )}
                />

                {!isDisabled && (
                    <Button
                        className="absolute right-0.5 top-1/2 shrink-0 -translate-y-1/2 text-muted-foreground opacity-0 transition-opacity hover:text-destructive group-hover:opacity-100"
                        onClick={(event) => {
                            event.stopPropagation();
                            handleDeleteClick(assistant);
                        }}
                        size="icon-xs"
                        variant="ghost"
                    >
                        <Trash2 />
                    </Button>
                )}
            </CommandItem>
        );
    };

    return (
        <>
            <Popover
                onOpenChange={setIsOpen}
                open={isOpen}
            >
                <PopoverTrigger asChild>
                    <Button
                        className="px-2"
                        disabled={isAssistantCreating}
                        variant="outline"
                    >
                        {selectedAssistant ? (
                            <>
                                <FlowStatusIcon
                                    status={selectedAssistant.status}
                                    tooltip={formatName(selectedAssistant.status)}
                                />
                                <ProviderIcon provider={selectedAssistant.provider} />
                                <span className="flex size-5 shrink-0 items-center justify-center rounded bg-muted text-xs font-medium text-muted-foreground">
                                    {selectedAssistantIndex}
                                </span>
                            </>
                        ) : (
                            <span className="flex h-5 shrink-0 items-center justify-center rounded bg-muted px-1 text-xs font-medium text-muted-foreground">
                                New
                            </span>
                        )}
                        <ChevronDown className="opacity-50" />
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

                            {!isDisabled && (
                                <CommandGroup>
                                    <CommandItem
                                        className="font-medium"
                                        onSelect={() => {
                                            onAssistantCreate();
                                            setIsOpen(false);
                                        }}
                                        value="create-new-assistant"
                                    >
                                        <Plus />
                                        Create new assistant
                                    </CommandItem>
                                </CommandGroup>
                            )}

                            {assistantsGroup.active.length > 0 && (
                                <CommandGroup heading={`Active (${assistantsGroup.active.length})`}>
                                    {assistantsGroup.active.map(({ assistant, index }) =>
                                        renderAssistantItem(assistant, index),
                                    )}
                                </CommandGroup>
                            )}

                            {assistantsGroup.finished.length > 0 && (
                                <CommandGroup heading={`Finished (${assistantsGroup.finished.length})`}>
                                    {assistantsGroup.finished.map(({ assistant, index }) =>
                                        renderAssistantItem(assistant, index),
                                    )}
                                </CommandGroup>
                            )}

                            {assistantsGroup.failed.length > 0 && (
                                <CommandGroup heading={`Failed (${assistantsGroup.failed.length})`}>
                                    {assistantsGroup.failed.map(({ assistant, index }) =>
                                        renderAssistantItem(assistant, index),
                                    )}
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
                itemName={currentAssistant?.title}
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

    // Check if search filter is active
    const hasActiveFilters = useMemo(() => {
        return !!searchValue.trim();
    }, [searchValue]);

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

    // Reset filters handler
    const handleResetFilters = () => {
        form.reset({ search: '' });
        setDebouncedSearchValue('');
        debouncedUpdateSearch.cancel();
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
                <div className="flex gap-2 p-px">
                    {/* Assistant Dropdown */}
                    {flowId && (
                        <AssistantsDropdown
                            assistants={assistants}
                            isAssistantCreating={isAssistantCreating}
                            isDisabled={isFormDisabled}
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
                                        <InputGroup>
                                            <InputGroupAddon>
                                                <Search />
                                            </InputGroupAddon>
                                            <InputGroupInput
                                                {...field}
                                                autoComplete="off"
                                                disabled={isAssistantCreating}
                                                placeholder="Search messages..."
                                                type="text"
                                            />
                                            {field.value && (
                                                <InputGroupAddon align="inline-end">
                                                    <InputGroupButton
                                                        disabled={isAssistantCreating}
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
                filteredLogs.length > 0 ? (
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
                                <Plus />
                            </EmptyMedia>
                            <EmptyTitle>No messages</EmptyTitle>
                            <EmptyDescription>No messages found for this assistant</EmptyDescription>
                        </EmptyHeader>
                    </Empty>
                )
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
