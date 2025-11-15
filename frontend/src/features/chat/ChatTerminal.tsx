import '@xterm/xterm/css/xterm.css';

import { zodResolver } from '@hookform/resolvers/zod';
import debounce from 'lodash/debounce';
import { ChevronDown, ChevronUp, Search, X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import Terminal from '@/components/shared/Terminal';
import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import type { TerminalLogFragmentFragment } from '@/graphql/types';

const searchFormSchema = z.object({
    search: z.string(),
});

interface ChatTerminalProps {
    logs: TerminalLogFragmentFragment[];
    selectedFlowId?: string | null;
}

const ChatTerminal = ({ logs: terminalLog, selectedFlowId }: ChatTerminalProps) => {
    // Separate state for immediate input value and debounced search value
    const [debouncedSearchValue, setDebouncedSearchValue] = useState('');
    const terminalRef = useRef<{ findNext: () => void; findPrevious: () => void } | null>(null);

    const form = useForm<z.infer<typeof searchFormSchema>>({
        resolver: zodResolver(searchFormSchema),
        defaultValues: {
            search: '',
        },
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
    }, [selectedFlowId, form, debouncedUpdateSearch]);

    // Filter logs based on debounced search value for better performance
    const filteredLogs = useMemo(() => {
        const search = debouncedSearchValue.toLowerCase().trim();
        const logs = terminalLog.map((log) => log.text);

        if (!search) {
            return logs;
        }

        return logs.filter((log) => log.toLowerCase().includes(search));
    }, [terminalLog, debouncedSearchValue]);

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
                                        type="text"
                                        placeholder="Search terminal logs..."
                                        className="px-9 pr-24"
                                        autoComplete="off"
                                    />
                                    <div className="absolute right-px top-1/2 flex -translate-y-1/2">
                                        {hasSearchValue && (
                                            <>
                                                <Button
                                                    type="button"
                                                    variant="ghost"
                                                    size="icon"
                                                    className="size-8"
                                                    onClick={handleFindPrevious}
                                                    title="Previous match"
                                                >
                                                    <ChevronUp className="size-4" />
                                                </Button>
                                                <Button
                                                    type="button"
                                                    variant="ghost"
                                                    size="icon"
                                                    className="size-8"
                                                    onClick={handleFindNext}
                                                    title="Next match"
                                                >
                                                    <ChevronDown className="size-4" />
                                                </Button>
                                            </>
                                        )}
                                        {field.value && (
                                            <Button
                                                type="button"
                                                variant="ghost"
                                                size="icon"
                                                className="size-8"
                                                onClick={handleClearSearch}
                                                title="Clear search"
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
                ref={terminalRef}
                logs={filteredLogs}
                searchValue={debouncedSearchValue}
                className="w-full grow"
            />
        </div>
    );
};

export default ChatTerminal;
