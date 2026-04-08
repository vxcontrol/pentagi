import { Check, ChevronRight, ListFilter, X } from 'lucide-react';
import { useMemo, useState } from 'react';

import { Button } from '@/components/ui/button';
import {
    Command,
    CommandEmpty,
    CommandGroup,
    CommandInput,
    CommandItem,
    CommandList,
    CommandSeparator,
} from '@/components/ui/command';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { cn } from '@/lib/utils';
import { useFlow } from '@/providers/flow-provider';

export interface FlowTasksDropdownValue {
    subtaskIds: string[];
    taskIds: string[];
}

interface FlowTasksDropdownProps {
    disabled?: boolean;
    onChange?: (value: FlowTasksDropdownValue) => void;
    value?: FlowTasksDropdownValue;
}

const FlowTasksDropdown = ({ disabled, onChange, value }: FlowTasksDropdownProps) => {
    const { flowData } = useFlow();
    const tasks = useMemo(() => flowData?.tasks ?? [], [flowData?.tasks]);
    const [isOpen, setIsOpen] = useState(false);
    const [expandedTaskIds, setExpandedTaskIds] = useState<Set<string>>(new Set());

    const taskIds = useMemo(() => new Set(value?.taskIds ?? []), [value?.taskIds]);
    const subtaskIds = useMemo(() => new Set(value?.subtaskIds ?? []), [value?.subtaskIds]);

    // Check if any filter is active
    const hasActiveFilters = taskIds.size > 0 || subtaskIds.size > 0;

    // Toggle task expansion
    const toggleTaskExpansion = (taskId: string) => {
        setExpandedTaskIds((prev) => {
            const newSet = new Set(prev);

            if (newSet.has(taskId)) {
                newSet.delete(taskId);
            } else {
                newSet.add(taskId);
            }

            return newSet;
        });
    };

    // Toggle task selection
    const toggleTaskSelection = (taskId: string) => {
        if (!onChange) {
            return;
        }

        const task = tasks.find((t) => t.id === taskId);
        const taskSubtaskIds = task?.subtasks?.map((st) => st.id) ?? [];
        const isSelected = taskIds.has(taskId);

        onChange({
            subtaskIds: isSelected
                ? Array.from(subtaskIds).filter((id) => !taskSubtaskIds.includes(id))
                : [...new Set([...subtaskIds, ...taskSubtaskIds])],
            taskIds: isSelected ? Array.from(taskIds).filter((id) => id !== taskId) : [...taskIds, taskId],
        });
    };

    // Toggle subtask selection
    const toggleSubtaskSelection = (subtaskId: string) => {
        if (!onChange) {
            return;
        }

        const task = tasks.find((t) => t.subtasks?.some((st) => st.id === subtaskId));
        const isSelected = subtaskIds.has(subtaskId);

        const selectedSubtaskIds = isSelected
            ? Array.from(subtaskIds).filter((id) => id !== subtaskId)
            : [...subtaskIds, subtaskId];

        const isSubtasksSelected = !!task?.subtasks?.every((st) => st.id === subtaskId || subtaskIds.has(st.id));
        const isTaskSelected = task && taskIds.has(task.id);

        const selectedTaskIds =
            isSelected && isTaskSelected
                ? Array.from(taskIds).filter((id) => id !== task.id)
                : !isSelected && isSubtasksSelected
                  ? [...taskIds, task!.id]
                  : Array.from(taskIds);

        onChange({
            subtaskIds: selectedSubtaskIds,
            taskIds: selectedTaskIds,
        });
    };

    // Clear all filters
    const clearFilters = () => {
        if (!onChange) {
            return;
        }

        onChange({
            subtaskIds: [],
            taskIds: [],
        });
    };

    return (
        <Popover
            onOpenChange={setIsOpen}
            open={isOpen}
        >
            <PopoverTrigger asChild>
                <Button
                    disabled={disabled}
                    size="icon"
                    variant="outline"
                >
                    <ListFilter className={cn(hasActiveFilters ? 'text-foreground' : 'text-muted-foreground')} />
                </Button>
            </PopoverTrigger>
            <PopoverContent
                align="end"
                className="w-[360px] p-0"
            >
                <Command>
                    <CommandInput placeholder="Search tasks..." />
                    <CommandList>
                        <CommandEmpty>Tasks not found</CommandEmpty>
                        {tasks?.length ? (
                            tasks.map((task) => (
                                <CommandGroup key={task.id}>
                                    <CommandItem
                                        onSelect={() => {
                                            toggleTaskSelection(task.id);
                                        }}
                                    >
                                        <div
                                            className={cn(
                                                'size-4 shrink-0',
                                                !!task?.subtasks?.length && 'cursor-pointer',
                                            )}
                                            onClick={(event) => {
                                                event.preventDefault();
                                                event.stopPropagation();

                                                if (task?.subtasks?.length) {
                                                    toggleTaskExpansion(task.id);
                                                }
                                            }}
                                        >
                                            {!!task?.subtasks?.length && (
                                                <ChevronRight
                                                    className={cn(
                                                        'transition-transform',
                                                        expandedTaskIds.has(task.id) && 'rotate-90',
                                                    )}
                                                />
                                            )}
                                        </div>
                                        <div className="flex-1 truncate">{task.title}</div>
                                        <Check
                                            className={cn(
                                                'ml-auto size-4 shrink-0',
                                                taskIds.has(task.id) ? 'opacity-100' : 'opacity-0',
                                            )}
                                        />
                                    </CommandItem>
                                    {!!task?.subtasks?.length &&
                                        expandedTaskIds.has(task.id) &&
                                        task.subtasks.map((subtask) => (
                                            <CommandItem
                                                className="ml-8 flex items-center gap-2"
                                                key={subtask.id}
                                                onSelect={() => {
                                                    toggleSubtaskSelection(subtask.id);
                                                }}
                                            >
                                                <div className="text-muted-foreground flex-1 truncate text-sm">
                                                    {subtask.title}
                                                </div>
                                                <Check
                                                    className={cn(
                                                        'ml-auto size-4 shrink-0',
                                                        subtaskIds.has(subtask.id) ? 'opacity-100' : 'opacity-0',
                                                    )}
                                                />
                                            </CommandItem>
                                        ))}
                                </CommandGroup>
                            ))
                        ) : (
                            <CommandItem
                                className="text-muted-foreground justify-center py-6 text-center"
                                disabled
                            >
                                No available tasks
                            </CommandItem>
                        )}
                    </CommandList>
                    {hasActiveFilters && (
                        <>
                            <CommandSeparator />
                            <CommandGroup>
                                <CommandItem
                                    onSelect={() => {
                                        clearFilters();
                                        setIsOpen(false);
                                    }}
                                >
                                    <X />
                                    Clear filter
                                </CommandItem>
                            </CommandGroup>
                        </>
                    )}
                </Command>
            </PopoverContent>
        </Popover>
    );
};

export default FlowTasksDropdown;
