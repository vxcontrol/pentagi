import { zodResolver } from '@hookform/resolvers/zod';
import debounce from 'lodash/debounce';
import { Bot, ListFilter, Search, X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { useFlow } from '@/providers/flow-provider';

import FlowTasksDropdown from '../flow-tasks-dropdown';
import FlowAgent from './flow-agent';

const searchFormSchema = z.object({
    filter: z
        .object({
            subtaskIds: z.array(z.string()),
            taskIds: z.array(z.string()),
        })
        .optional(),
    search: z.string(),
});

const FlowAgents = () => {
    const { flowData, flowId } = useFlow();

    const logs = useMemo(() => flowData?.agentLogs ?? [], [flowData?.agentLogs]);
    const agentsEndRef = useRef<HTMLDivElement>(null);

    // Separate state for immediate input value and debounced search value
    const [debouncedSearchValue, setDebouncedSearchValue] = useState('');

    const scrollAgents = () => {
        agentsEndRef.current?.scrollIntoView();
    };

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
            filtered = filtered.filter((log) => {
                return (
                    log.task.toLowerCase().includes(search) ||
                    log.result?.toLowerCase().includes(search) ||
                    log.executor.toLowerCase().includes(search) ||
                    log.initiator.toLowerCase().includes(search)
                );
            });
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

    const hasLogs = filteredLogs && filteredLogs.length > 0;

    useEffect(() => {
        if (hasLogs) {
            scrollAgents();
        }
    }, [logs, hasLogs]);

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

    return (
        <div className="flex h-full flex-col">
            <div className="sticky top-0 z-10 bg-background pb-4">
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
                                            placeholder="Search agent logs..."
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

            {hasLogs ? (
                <div className="flex-1 space-y-4 overflow-auto pb-4">
                    {filteredLogs.map((log) => (
                        <FlowAgent
                            key={log.id}
                            log={log}
                            searchValue={debouncedSearchValue}
                        />
                    ))}
                    <div ref={agentsEndRef} />
                </div>
            ) : hasActiveFilters ? (
                <Empty>
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <ListFilter />
                        </EmptyMedia>
                        <EmptyTitle>No agent logs found</EmptyTitle>
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
                            <Bot />
                        </EmptyMedia>
                        <EmptyTitle>No agent logs available</EmptyTitle>
                        <EmptyDescription>Agent logs will appear here when agents are working</EmptyDescription>
                    </EmptyHeader>
                </Empty>
            )}
        </div>
    );
};

export default FlowAgents;
