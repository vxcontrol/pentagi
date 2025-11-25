import { zodResolver } from '@hookform/resolvers/zod';
import '@xterm/xterm/css/xterm.css';
import debounce from 'lodash/debounce';
import { ChevronDown, ChevronUp, Search, X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import Terminal from '@/components/shared/Terminal';
import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { useFlow } from '@/providers/FlowProvider';

const searchFormSchema = z.object({
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

    // Filter logs based on debounced search value for better performance
    const filteredLogs = useMemo(() => {
        const search = debouncedSearchValue.toLowerCase().trim();
        const logs = terminalLogs.map((log) => log.text);

        if (!search) {
            return logs;
        }

        return logs.filter((log) => log.toLowerCase().includes(search));
    }, [terminalLogs, debouncedSearchValue]);

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

    const hasSearchValue = !!debouncedSearchValue.trim();

    return (
        <div className="flex size-full flex-col gap-4">
            <div className="sticky top-0 z-10 bg-background pr-4">
                <Form {...form}>
                    <FormField
                        control={form.control}
                        name="search"
                        render={({ field }) => (
                            <FormControl>
                                <div className="relative p-px">
                                    <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                                    <Input
                                        {...field}
                                        autoComplete="off"
                                        className="px-9 pr-24"
                                        placeholder="Search terminal logs..."
                                        type="text"
                                    />
                                    <div className="absolute right-px top-1/2 flex -translate-y-1/2">
                                        {hasSearchValue && (
                                            <>
                                                <Button
                                                    className="size-8"
                                                    onClick={handleFindPrevious}
                                                    size="icon"
                                                    title="Previous match"
                                                    type="button"
                                                    variant="ghost"
                                                >
                                                    <ChevronUp className="size-4" />
                                                </Button>
                                                <Button
                                                    className="size-8"
                                                    onClick={handleFindNext}
                                                    size="icon"
                                                    title="Next match"
                                                    type="button"
                                                    variant="ghost"
                                                >
                                                    <ChevronDown className="size-4" />
                                                </Button>
                                            </>
                                        )}
                                        {field.value && (
                                            <Button
                                                className="size-8"
                                                onClick={handleClearSearch}
                                                size="icon"
                                                title="Clear search"
                                                type="button"
                                                variant="ghost"
                                            >
                                                <X className="size-4" />
                                            </Button>
                                        )}
                                    </div>
                                </div>
                            </FormControl>
                        )}
                    />
                </Form>
            </div>
            <Terminal
                className="w-full grow"
                logs={filteredLogs}
                ref={terminalRef}
                searchValue={debouncedSearchValue}
            />
        </div>
    );
};

export default FlowTerminal;

