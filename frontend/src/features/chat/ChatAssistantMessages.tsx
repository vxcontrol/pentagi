import { zodResolver } from '@hookform/resolvers/zod';
import { ChevronDown, Loader2, Plus, Search, Trash2, X } from 'lucide-react';
import debounce from 'lodash/debounce';
import { memo, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { BreadcrumbProvider } from '@/components/ui/breadcrumb';
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
import ChatAssistantFormInput from '@/features/chat/ChatAssistantFormInput';
import type { AssistantFragmentFragment, AssistantLogFragmentFragment } from '@/graphql/types';
import { useSettingsQuery } from '@/graphql/types';
import { Log } from '@/lib/log';
import { cn } from '@/lib/utils';
import { isProviderValid, type Provider } from '@/models/Provider';

import ChatMessage from './ChatMessage';

interface ChatAssistantMessagesProps {
    selectedFlowId: string | null;
    logs?: AssistantLogFragmentFragment[];
    className?: string;
    assistants: AssistantFragmentFragment[];
    selectedAssistantId?: string | null;
    selectedProvider: Provider | null;
    providers: Provider[];
    onSelectAssistant?: (assistantId: string | null) => void;
    onCreateAssistant?: () => void;
    onDeleteAssistant?: (assistantId: string) => void;
    onSubmitMessage?: (assistantId: string, message: string, useAgents: boolean) => Promise<void>;
    onCreateNewAssistant?: (message: string, useAgents: boolean) => Promise<void>;
    onStopAssistant?: (assistantId: string) => Promise<void>;
}

const searchFormSchema = z.object({
    search: z.string(),
});

const ChatAssistantMessages = ({
    selectedFlowId,
    logs,
    className,
    assistants,
    selectedAssistantId,
    selectedProvider,
    providers,
    onSelectAssistant,
    onCreateAssistant,
    onDeleteAssistant,
    onSubmitMessage,
    onCreateNewAssistant,
    onStopAssistant,
}: ChatAssistantMessagesProps) => {
    const messagesEndRef = useRef<HTMLDivElement>(null);
    const [isCreatingAssistant, setIsCreatingAssistant] = useState(false);

    // Separate state for immediate input value and debounced search value
    const [debouncedSearchValue, setDebouncedSearchValue] = useState('');

    const scrollableContainerRef = useRef<HTMLDivElement>(null);
    const [isScrolledToBottom, setIsScrolledToBottom] = useState(true);
    const isScrolledToBottomRef = useRef(true);
    const needsScrollRef = useRef(false);
    const animationFrameRef = useRef<number | null>(null);
    const prevScrollTopRef = useRef<number>(0);
    const lastScrollTimeRef = useRef<number>(0);
    const scrollTimeoutRef = useRef<NodeJS.Timeout | null>(null);
    const scrollThrottleDelay = 500; // 500ms throttle
    const textareaId = 'chat-textarea';

    // Get system settings
    const { data: settingsData } = useSettingsQuery();

    // Update ref when state changes
    useEffect(() => {
        isScrolledToBottomRef.current = isScrolledToBottom;
    }, [isScrolledToBottom]);

    const scrollMessages = useCallback(() => {
        const now = Date.now();
        const timeSinceLastScroll = now - lastScrollTimeRef.current;

        // Only scroll if enough time has passed since the last scroll
        if (timeSinceLastScroll >= scrollThrottleDelay && messagesEndRef.current) {
            messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
            lastScrollTimeRef.current = now;
        }
    }, [scrollThrottleDelay]);

    // Throttled scroll function that respects the delay
    const throttledScrollMessages = useCallback(() => {
        const now = Date.now();
        const timeSinceLastScroll = now - lastScrollTimeRef.current;

        // Clear any existing timeout
        if (scrollTimeoutRef.current) {
            clearTimeout(scrollTimeoutRef.current);
            scrollTimeoutRef.current = null;
        }

        if (timeSinceLastScroll >= scrollThrottleDelay) {
            // Enough time has passed, scroll immediately
            scrollMessages();
        } else {
            // Not enough time has passed, schedule a scroll for later
            const remainingTime = scrollThrottleDelay - timeSinceLastScroll;
            scrollTimeoutRef.current = setTimeout(() => {
                scrollMessages();
                scrollTimeoutRef.current = null;
            }, remainingTime);
        }
    }, [scrollMessages, scrollThrottleDelay]);

    useEffect(() => {
        if (scrollableContainerRef.current) {
            prevScrollTopRef.current = scrollableContainerRef.current.scrollTop;
        }
    }, []);

    const handleScroll = useCallback(() => {
        const el = scrollableContainerRef.current;
        if (!el) return;

        const { scrollTop, scrollHeight, clientHeight } = el;
        const effectivelyAtBottom = scrollHeight - scrollTop <= clientHeight + 2;

        if (effectivelyAtBottom) {
            if (!isScrolledToBottomRef.current) {
                setIsScrolledToBottom(true);
                needsScrollRef.current = false;
            }
        } else {
            const scrollUpThreshold = 10;
            if (isScrolledToBottomRef.current && scrollTop < prevScrollTopRef.current - scrollUpThreshold) {
                setIsScrolledToBottom(false);
            }
        }
        prevScrollTopRef.current = scrollTop;
    }, []);

    useEffect(() => {
        const scrollableElement = scrollableContainerRef.current;
        if (!scrollableElement) return;

        scrollableElement.addEventListener('scroll', handleScroll);
        return () => {
            scrollableElement.removeEventListener('scroll', handleScroll);
        };
    }, [handleScroll]);

    const form = useForm<z.infer<typeof searchFormSchema>>({
        resolver: zodResolver(searchFormSchema),
        defaultValues: {
            search: '',
        },
    });

    const searchValue = form.watch('search');

    // Create debounced function to update search value
    const debouncedUpdateSearch = useMemo(
        () => debounce((value: string) => {
            setDebouncedSearchValue(value);
        }, 500),
        []
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
    }, [selectedFlowId, form, debouncedUpdateSearch]);

    // Get the current selected assistant
    const selectedAssistant = useMemo(() => {
        if (!selectedAssistantId || !assistants) return null;
        return assistants.find((assistant) => assistant.id === selectedAssistantId) || null;
    }, [assistants, selectedAssistantId]);

    // Get the current selected assistant index (1-based, reversed)
    const selectedAssistantIndex = useMemo(() => {
        if (!selectedAssistantId || !assistants) return null;
        const index = assistants.findIndex((assistant) => assistant.id === selectedAssistantId);
        return index !== -1 ? assistants.length - index : null;
    }, [assistants, selectedAssistantId]);

    // Check if selected provider is available
    const isProviderAvailable = useMemo(() => {
        if (!selectedProvider) return false;
        return isProviderValid(selectedProvider, providers);
    }, [providers, selectedProvider]);

    // Check if the assistant's provider is available
    const isAssistantProviderAvailable = useMemo(() => {
        if (!selectedAssistant) return true; // If no assistant is selected, consider provider available
        return isProviderValid(selectedAssistant.provider, providers);
    }, [providers, selectedAssistant]);

    // Calculate default useAgents value
    const isUseAgentsDefault = useMemo(() => {
        // If creating a new assistant, use system setting
        if (isCreatingAssistant || !selectedAssistant) {
            return settingsData?.settings?.assistantUseAgents ?? false;
        }
        // If assistant is selected and not creating new, use its useAgents setting
        return selectedAssistant.useAgents;
    }, [selectedAssistant, settingsData?.settings?.assistantUseAgents, isCreatingAssistant]);

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
        if (onDeleteAssistant) {
            onDeleteAssistant(assistantId);
        }
    };

    const handleSubmitMessage = async (message: string, useAgents: boolean) => {
        if (!message.trim()) {
            return;
        }

        try {
            if (!selectedAssistantId) {
                // If no assistant is selected, create a new one
                setIsCreatingAssistant(true);
                if (onCreateNewAssistant) {
                    await onCreateNewAssistant(message, useAgents);
                }
            } else if (onSubmitMessage) {
                // Otherwise call the existing assistant
                await onSubmitMessage(selectedAssistantId, message, useAgents);
            }
        } catch (error) {
            Log.error('Error submitting message:', error);
            throw error;
        } finally {
            setIsCreatingAssistant(false);
        }
    };

    const handleStopAssistant = async () => {
        if (selectedAssistantId && onStopAssistant) {
            try {
                await onStopAssistant(selectedAssistantId);
            } catch (error) {
                Log.error('Error stopping assistant:', error);
                throw error;
            }
        }
    };

    // Handle click on Create Assistant option in dropdown
    const handleCreateAssistantClick = () => {
        if (onCreateAssistant) {
            onCreateAssistant();
        }
    };

    const focusTextarea = useCallback(() => {
        const textarea = document.querySelector(`#${textareaId}`) as HTMLTextAreaElement;
        if (textarea && !textarea.disabled) {
            textarea.focus();
        }
    }, [textareaId]);

    // Single useEffect for all cleanup and scroll logic
    useEffect(() => {
        if (isScrolledToBottom) {
            needsScrollRef.current = true;
            if (!animationFrameRef.current) {
                animationFrameRef.current = requestAnimationFrame(() => {
                    if (needsScrollRef.current) {
                        throttledScrollMessages();
                        needsScrollRef.current = false;
                    }
                    animationFrameRef.current = null;
                });
            }
        } else {
            needsScrollRef.current = false;
            if (animationFrameRef.current) {
                cancelAnimationFrame(animationFrameRef.current);
                animationFrameRef.current = null;
            }
        }

        return () => {
            if (animationFrameRef.current) {
                cancelAnimationFrame(animationFrameRef.current);
                animationFrameRef.current = null;
            }
            if (scrollTimeoutRef.current) {
                clearTimeout(scrollTimeoutRef.current);
                scrollTimeoutRef.current = null;
            }
        };
    }, [filteredLogs, throttledScrollMessages, isScrolledToBottom]);

    return (
        <div className={cn('flex h-full flex-col', className)}>
            <div className="sticky top-0 z-10 bg-background pb-4">
                <div className="flex gap-2 p-px">
                    {/* Assistant Selector Dropdown */}
                    {selectedFlowId && selectedFlowId !== 'new' && (
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    variant="outline"
                                    size="icon"
                                    className="relative w-12 shrink-0"
                                    aria-label="Select Assistant"
                                    disabled={isCreatingAssistant}
                                    tabIndex={-1}
                                    onFocus={(e) => e.preventDefault()}
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
                                onCloseAutoFocus={(e) => {
                                    e.preventDefault();
                                    focusTextarea();
                                }}
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
                                        key={assistant.id}
                                        className={cn(
                                            'flex items-center justify-between gap-2',
                                            selectedAssistantId === assistant.id && 'bg-accent font-medium',
                                            !isProviderValid(assistant.provider, providers) && 'opacity-50',
                                        )}
                                        tabIndex={-1}
                                    >
                                        <div
                                            className="flex grow cursor-pointer items-center gap-2"
                                            onClick={() => {
                                                onSelectAssistant?.(assistant.id);
                                            }}
                                        >
                                            <span className="flex size-5 shrink-0 items-center justify-center rounded bg-muted text-xs font-medium text-muted-foreground">
                                                {assistants.length - index}
                                            </span>
                                            <BreadcrumbProvider provider={assistant.provider} />
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
                                            variant="ghost"
                                            size="icon"
                                            className="ml-auto size-6"
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                handleDeleteAssistant(assistant.id);
                                            }}
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
                                                type="text"
                                                placeholder="Search messages..."
                                                className="px-9"
                                                autoComplete="off"
                                                disabled={isCreatingAssistant}
                                            />
                                            {field.value && (
                                                <Button
                                                    type="button"
                                                    variant="ghost"
                                                    size="icon"
                                                    className="absolute right-0 top-1/2 -translate-y-1/2"
                                                    onClick={() => {
                                                        form.reset({ search: '' });
                                                        setDebouncedSearchValue('');
                                                        debouncedUpdateSearch.cancel();
                                                    }}
                                                    disabled={isCreatingAssistant}
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

            {isCreatingAssistant
                ? (
                    <div className="flex flex-1 items-center justify-center">
                        <div className="flex flex-col items-center gap-4 text-muted-foreground">
                            <Loader2 className="size-12 animate-spin" />
                            <p>Creating assistant...</p>
                        </div>
                    </div>
                )
                : selectedAssistantId
                    ? (
                        // Show messages for selected assistant
                        <div ref={scrollableContainerRef} className="flex-1 space-y-4 overflow-y-auto pb-4">
                            {filteredLogs.map((log) => (
                                <ChatMessage
                                    key={log.id}
                                    log={log}
                                    searchValue={debouncedSearchValue}
                                />
                            ))}
                            <div ref={messagesEndRef} />
                        </div>
                    )
                    : selectedFlowId === 'new'
                        ? (
                            // Show placeholder for new flow
                            <div className="flex flex-1 items-center justify-center">
                                <div className="flex flex-col items-center gap-2 text-muted-foreground">
                                    <p>No Active Assistants</p>
                                    <p className="text-xs">Start by typing a message to create a new assistant</p>
                                </div>
                            </div>
                        )
                        : (
                            // Show placeholder when no assistant is selected in existing flow
                            <div className="flex flex-1 items-center justify-center">
                                <div className="flex flex-col items-center gap-2 text-muted-foreground">
                                    <p>No Assistant Selected</p>
                                    <p className="text-xs">Select an assistant from the dropdown or create a new one</p>
                                </div>
                            </div>
                        )}

            <div className="sticky bottom-0 border-t bg-background p-px pt-4">
                <ChatAssistantFormInput
                    selectedFlowId={selectedFlowId}
                    assistantStatus={selectedAssistant?.status}
                    isUseAgentsDefault={isUseAgentsDefault}
                    isProviderAvailable={isProviderAvailable && isAssistantProviderAvailable}
                    onSubmitMessage={handleSubmitMessage}
                    onStopFlow={handleStopAssistant}
                    isCreatingAssistant={isCreatingAssistant}
                />
            </div>
        </div>
    );
};

export default memo(ChatAssistantMessages);
