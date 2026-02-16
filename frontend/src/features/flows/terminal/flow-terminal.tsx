import { zodResolver } from '@hookform/resolvers/zod';
import '@xterm/xterm/css/xterm.css';
import debounce from 'lodash/debounce';
import { ChevronDown, ChevronUp, Search, X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import MonacoTerminal from '@/components/shared/monaco-terminal';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { useFlow } from '@/providers/flow-provider';

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
            <div className="bg-background sticky top-0 z-10 pr-4">
                <Form {...form}>
                    <div className="p-px">
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
                    </div>
                </Form>
            </div>
            <MonacoTerminal
                className="w-full grow"
                logs={filteredLogs}
                ref={terminalRef}
                searchValue={debouncedSearchValue}
            />
        </div>
    );
};

export default FlowTerminal;
