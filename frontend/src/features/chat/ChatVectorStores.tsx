import { zodResolver } from '@hookform/resolvers/zod';
import debounce from 'lodash/debounce';
import { Search, X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import type { VectorStoreLogFragmentFragment } from '@/graphql/types';

import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { Input } from '@/components/ui/input';

import ChatVectorStore from './ChatVectorStore';

interface ChatVectorStoresProps {
    logs?: VectorStoreLogFragmentFragment[];
    selectedFlowId?: null | string;
}

const searchFormSchema = z.object({
    search: z.string(),
});

const ChatVectorStores = ({ logs, selectedFlowId }: ChatVectorStoresProps) => {
    const vectorStoresEndRef = useRef<HTMLDivElement>(null);

    // Separate state for immediate input value and debounced search value
    const [debouncedSearchValue, setDebouncedSearchValue] = useState('');

    const scrollVectorStores = () => {
        vectorStoresEndRef.current?.scrollIntoView();
    };

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
    }, [selectedFlowId, form, debouncedUpdateSearch]);

    // Memoize filtered logs to avoid recomputing on every render
    // Use debouncedSearchValue for filtering to improve performance
    const filteredLogs = useMemo(() => {
        const search = debouncedSearchValue.toLowerCase().trim();

        if (!search || !logs) {
            return logs || [];
        }

        return logs.filter((log) => {
            return (
                log.query.toLowerCase().includes(search) ||
                log.result?.toLowerCase().includes(search) ||
                log.filter.toLowerCase().includes(search) ||
                log.action.toLowerCase().includes(search) ||
                log.executor.toLowerCase().includes(search) ||
                log.initiator.toLowerCase().includes(search)
            );
        });
    }, [logs, debouncedSearchValue]);

    const hasLogs = filteredLogs && filteredLogs.length > 0;

    // Only scroll when logs data changes, not when filtering changes
    useEffect(() => {
        if (logs && logs.length > 0) {
            scrollVectorStores();
        }
    }, [logs]);

    return (
        <div className="flex h-full flex-col">
            <div className="sticky top-0 z-10 bg-background pb-4">
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
                                        className="px-9"
                                        placeholder="Search vector store logs..."
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
            {hasLogs ? (
                <div className="flex-1 space-y-4 overflow-auto pb-4">
                    {filteredLogs.map((log) => (
                        <ChatVectorStore
                            key={log.id}
                            log={log}
                            searchValue={debouncedSearchValue}
                        />
                    ))}
                    <div ref={vectorStoresEndRef} />
                </div>
            ) : (
                <div className="flex flex-1 items-center justify-center">
                    <div className="flex flex-col items-center gap-2 text-muted-foreground">
                        <p>No vector store logs available</p>
                        <p className="text-xs">
                            Vector store logs will appear here when the agent uses knowledge database
                        </p>
                    </div>
                </div>
            )}
        </div>
    );
};

export default ChatVectorStores;
