import { zodResolver } from '@hookform/resolvers/zod';
import '@xterm/xterm/css/xterm.css';
import debounce from 'lodash/debounce';
import { ChevronDown, ChevronUp, ListFilter, Search, X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import Terminal from '@/components/shared/terminal';
import { Button } from '@/components/ui/button';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { cn } from '@/lib/utils';
import { useFlow } from '@/providers/flow-provider';

import FlowTasksDropdown from '../flow-tasks-dropdown';

const searchFormSchema = z.object({
    filter: z
        .object({
            subtaskIds: z.array(z.string()),
            taskIds: z.array(z.string()),
        })
        .optional(),
    search: z.string(),
});

const FlowTerminal = () => {
    const { flowData, flowId } = useFlow();

    const terminalLogs = useMemo(() => flowData?.terminalLogs ?? [], [flowData?.terminalLogs]);
    // Separate state for immediate input value and debounced search value
    const [debouncedSearchValue, setDebouncedSearchValue] = useState('');
    const terminalRef = useRef<null | { findNext: () => void; findPrevious: () => void }>(null);

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

    const hasActiveFilters = useMemo(() => {
        const hasSearch = !!searchValue.trim();
        const hasTaskFilters = !!(filter?.taskIds?.length || filter?.subtaskIds?.length);

        return hasSearch || hasTaskFilters;
    }, [searchValue, filter]);

    const filteredLogs = useMemo(() => {
        const search = debouncedSearchValue.toLowerCase().trim();

        let filtered = terminalLogs;

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

        const texts = filtered.map((log) => log.text);

        if (!search) {
            return texts;
        }

        return texts.filter((text) => text.toLowerCase().includes(search));
    }, [terminalLogs, debouncedSearchValue, filter]);

    const handleFindNext = () => {
        if (terminalRef.current && debouncedSearchValue.trim()) {
            terminalRef.current.findNext();
        }
    };

    const handleFindPrevious = () => {
        if (terminalRef.current && debouncedSearchValue.trim()) {
            terminalRef.current.findPrevious();
        }
    };

    const handleClearSearch = () => {
        form.reset({ search: '' });
        setDebouncedSearchValue('');
        debouncedUpdateSearch.cancel();
    };

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

    const hasSearchValue = !!debouncedSearchValue.trim();
    const hasLogs = filteredLogs.length > 0;

    return (
        <div className="flex size-full flex-col gap-4">
            <div className="bg-background sticky top-0 z-10 pr-4">
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
                                            placeholder="Search terminal logs..."
                                            type="text"
                                        />
                                        <InputGroupAddon align="inline-end">
                                            {hasSearchValue && (
                                                <>
                                                    <InputGroupButton
                                                        onClick={handleFindPrevious}
                                                        size="icon-xs"
                                                        title="Previous match"
                                                        type="button"
                                                    >
                                                        <ChevronUp className="size-4" />
                                                    </InputGroupButton>
                                                    <InputGroupButton
                                                        onClick={handleFindNext}
                                                        size="icon-xs"
                                                        title="Next match"
                                                        type="button"
                                                    >
                                                        <ChevronDown className="size-4" />
                                                    </InputGroupButton>
                                                </>
                                            )}
                                            {field.value && (
                                                <InputGroupButton
                                                    onClick={handleClearSearch}
                                                    size="icon-xs"
                                                    title="Clear search"
                                                    type="button"
                                                >
                                                    <X className="size-4" />
                                                </InputGroupButton>
                                            )}
                                        </InputGroupAddon>
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
            <Terminal
                className={cn('w-full grow', hasActiveFilters && !hasLogs && 'hidden')}
                logs={filteredLogs}
                ref={terminalRef}
                searchValue={debouncedSearchValue}
            />
            {hasActiveFilters && !hasLogs && (
                <Empty>
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <ListFilter />
                        </EmptyMedia>
                        <EmptyTitle>No terminal logs found</EmptyTitle>
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
            )}
        </div>
    );
};

export default FlowTerminal;
