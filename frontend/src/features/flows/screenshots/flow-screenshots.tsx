import { zodResolver } from '@hookform/resolvers/zod';
import debounce from 'lodash/debounce';
import { Camera, Search, X } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { useFlow } from '@/providers/flow-provider';

import FlowScreenshot from './flow-screenshot';

const searchFormSchema = z.object({
    search: z.string(),
});

const FlowScreenshots = () => {
    const { flowData, flowId } = useFlow();

    const screenshots = useMemo(() => flowData?.screenshots ?? [], [flowData?.screenshots]);
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

    // Memoize filtered screenshots to avoid recomputing on every render
    // Use debouncedSearchValue for filtering to improve performance
    const filteredScreenshots = useMemo(() => {
        const search = debouncedSearchValue.toLowerCase().trim();

        if (!search || !screenshots) {
            return screenshots || [];
        }

        return screenshots.filter((screenshot) => {
            return screenshot.url.toLowerCase().includes(search);
        });
    }, [screenshots, debouncedSearchValue]);

    const hasScreenshots = filteredScreenshots && filteredScreenshots.length > 0;

    return (
        <div className="flex h-full flex-col">
            <div className="bg-background sticky top-0 z-10 pb-4">
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
                                            placeholder="Search screenshots..."
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
                    </div>
                </Form>
            </div>

            {hasScreenshots ? (
                <div className="flex flex-1 flex-col gap-4 overflow-auto">
                    {filteredScreenshots.map((screenshot) => (
                        <FlowScreenshot
                            key={screenshot.id}
                            screenshot={screenshot}
                        />
                    ))}
                </div>
            ) : (
                <Empty>
                    <EmptyHeader>
                        <EmptyMedia variant="icon">
                            <Camera />
                        </EmptyMedia>
                        <EmptyTitle>No screenshots available</EmptyTitle>
                        <EmptyDescription>Screenshots will appear here once the agent captures them</EmptyDescription>
                    </EmptyHeader>
                </Empty>
            )}
        </div>
    );
};

export default FlowScreenshots;
