import { zodResolver } from '@hookform/resolvers/zod';
import debounce from 'lodash/debounce';
import { ListTodo, Search, X } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { useFlow } from '@/providers/flow-provider';

import FlowTask from './flow-task';

const searchFormSchema = z.object({
    search: z.string(),
});

// Helper function to check if text contains search value (case-insensitive)
const containsSearchValue = (text: null | string | undefined, searchValue: string): boolean => {
    if (!text || !searchValue.trim()) {
        return false;
    }

    return text.toLowerCase().includes(searchValue.toLowerCase().trim());
};

const FlowTasks = () => {
    const { flowData, flowId } = useFlow();

    const tasks = useMemo(() => flowData?.tasks ?? [], [flowData?.tasks]);
    // Separate state for immediate input value and debounced search value
    const [debouncedSearchValue, setDebouncedSearchValue] = useState('');

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

    // Memoize filtered tasks to avoid recomputing on every render
    // Use debouncedSearchValue for filtering to improve performance
    const filteredTasks = useMemo(() => {
        const search = debouncedSearchValue.toLowerCase().trim();

        if (!search || !tasks) {
            return tasks || [];
        }

        return tasks.filter((task) => {
            const taskMatches = containsSearchValue(task.title, search) || containsSearchValue(task.result, search);

            const subtaskMatches =
                task.subtasks?.some(
                    (subtask) =>
                        containsSearchValue(subtask.title, search) ||
                        containsSearchValue(subtask.description, search) ||
                        containsSearchValue(subtask.result, search),
                ) || false;

            return taskMatches || subtaskMatches;
        });
    }, [tasks, debouncedSearchValue]);

    const sortedTasks = [...(filteredTasks || [])].sort((a, b) => +a.id - +b.id);
    const hasTasks = filteredTasks && filteredTasks.length > 0;

    return (
        <div className="flex h-full flex-col">
            <div className="sticky top-0 z-10 bg-background pb-4">
                {/* Search Input */}
                <div className="p-px">
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
                                            placeholder="Search tasks and subtasks..."
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
            </div>

            {hasTasks ? (
                <div className="flex-1 space-y-4 overflow-auto pb-4">
                    {sortedTasks.map((task) => (
                        <FlowTask
                            key={task.id}
                            searchValue={debouncedSearchValue}
                            task={task}
                        />
                    ))}
                </div>
            ) : (
                <Empty>
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <ListTodo />
                        </EmptyMedia>
                        <EmptyTitle>No tasks found for this flow</EmptyTitle>
                        <EmptyDescription>Tasks will appear here once the agent starts working</EmptyDescription>
                    </EmptyHeader>
                </Empty>
            )}
        </div>
    );
};

export default FlowTasks;
