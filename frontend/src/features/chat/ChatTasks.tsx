import { zodResolver } from '@hookform/resolvers/zod';
import debounce from 'lodash/debounce';
import { Copy, Download, ExternalLink, NotepadText, Search, X } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { Log } from '@/lib/log';
import { copyToClipboard, downloadTextFile, generateFileName, generateReport } from '@/lib/report';
import { useFlow } from '@/providers/FlowProvider';

import ChatTask from './ChatTask';

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

const ChatTasks = () => {
    const { flowData, flowId } = useFlow();

    const flow = flowData?.flow;
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

    // Check if flow is available for report generation
    const isReportDisabled = !flow || !flowId;

    // Report export handlers
    const handleCopyToClipboard = async () => {
        if (isReportDisabled) {
            return;
        }

        const reportContent = generateReport(tasks || [], flow);
        const success = await copyToClipboard(reportContent);

        if (!success) {
            Log.error('Failed to copy report to clipboard');
        }
    };

    const handleDownloadMD = () => {
        if (isReportDisabled || !flow) {
            return;
        }

        try {
            // Generate report content
            const reportContent = generateReport(tasks || [], flow);

            // Generate file name
            const baseFileName = generateFileName(flow);
            const fileName = `${baseFileName}.md`;

            // Download file
            downloadTextFile(reportContent, fileName, 'text/markdown; charset=UTF-8');
        } catch (error) {
            Log.error('Failed to download markdown report:', error);
        }
    };

    const handleDownloadPDF = () => {
        if (isReportDisabled || !flow || !flowId) {
            return;
        }

        // Open new tab (not popup) with report page and download flag
        const url = `/flows/${flowId}/report?download=true&silent=true`;
        window.open(url, '_blank');
    };

    const handleOpenWebView = () => {
        if (isReportDisabled || !flowId) {
            return;
        }

        // Open new tab with report page for web viewing
        const url = `/flows/${flowId}/report`;
        window.open(url, '_blank');
    };

    return (
        <div className="flex h-full flex-col">
            <div className="sticky top-0 z-10 bg-background pb-4">
                <div className="flex gap-2 p-px">
                    {/* Report Export Dropdown */}
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <div>
                                <Tooltip>
                                    <TooltipTrigger asChild>
                                        <Button
                                            aria-label="Export Report"
                                            className="shrink-0"
                                            disabled={isReportDisabled}
                                            size="icon"
                                            variant="outline"
                                        >
                                            <NotepadText className="size-4" />
                                        </Button>
                                    </TooltipTrigger>
                                    <TooltipContent>
                                        {isReportDisabled ? 'Select a flow to export report' : 'Report'}
                                    </TooltipContent>
                                </Tooltip>
                            </div>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="start">
                            <DropdownMenuItem
                                className="flex items-center gap-2"
                                disabled={isReportDisabled}
                                onClick={handleOpenWebView}
                            >
                                <ExternalLink className="size-4" />
                                Open web view
                            </DropdownMenuItem>
                            <DropdownMenuItem
                                className="flex items-center gap-2"
                                disabled={isReportDisabled}
                                onClick={handleCopyToClipboard}
                            >
                                <Copy className="size-4" />
                                Copy to clipboard
                            </DropdownMenuItem>
                            <DropdownMenuItem
                                className="flex items-center gap-2"
                                disabled={isReportDisabled}
                                onClick={handleDownloadMD}
                            >
                                <Download className="size-4" />
                                Download MD
                            </DropdownMenuItem>
                            <DropdownMenuItem
                                className="flex items-center gap-2"
                                disabled={isReportDisabled}
                                onClick={handleDownloadPDF}
                            >
                                <Download className="size-4" />
                                Download PDF
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>

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
            </div>

            {hasTasks ? (
                <div className="flex-1 space-y-4 overflow-auto pb-4">
                    {sortedTasks.map((task) => (
                        <ChatTask
                            key={task.id}
                            searchValue={debouncedSearchValue}
                            task={task}
                        />
                    ))}
                </div>
            ) : (
                <div className="flex flex-1 items-center justify-center">
                    <div className="flex flex-col items-center gap-2 text-muted-foreground">
                        <p>No tasks found for this flow</p>
                        <p className="text-xs">Tasks will appear here once the agent starts working</p>
                    </div>
                </div>
            )}
        </div>
    );
};

export default ChatTasks;
