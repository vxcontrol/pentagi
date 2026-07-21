import { zodResolver } from '@hookform/resolvers/zod';
import { Camera, ChevronDown, Search, X } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { useDebouncedCallback } from 'use-debounce';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { useAutoScroll } from '@/hooks/use-auto-scroll';
import { useFlow } from '@/providers/flow-provider';

import FlowScreenshot from './flow-screenshot';

const searchFormSchema = z.object({
    search: z.string(),
});

function FlowScreenshots() {
    const { flowData, flowId } = useFlow();

    const screenshots = useMemo(() => flowData?.screenshots ?? [], [flowData?.screenshots]);
    const [debouncedSearchValue, setDebouncedSearchValue] = useState('');

    const { containerRef, endRef, hasNewMessages, isScrolledToBottom, scrollToEnd } = useAutoScroll(
        screenshots,
        flowId,
    );

    const form = useForm<z.infer<typeof searchFormSchema>>({
        defaultValues: {
            search: '',
        },
        resolver: zodResolver(searchFormSchema),
    });

    const searchValue = form.watch('search');

    const debouncedUpdateSearch = useDebouncedCallback((value: string) => {
        setDebouncedSearchValue(value);
    }, 500);

    useEffect(() => {
        debouncedUpdateSearch(searchValue);

        return () => {
            debouncedUpdateSearch.cancel();
        };
    }, [searchValue, debouncedUpdateSearch]);

    useEffect(() => {
        return () => {
            debouncedUpdateSearch.cancel();
        };
    }, [debouncedUpdateSearch]);

    useEffect(() => {
        form.reset({ search: '' });
        setDebouncedSearchValue('');
        debouncedUpdateSearch.cancel();
    }, [flowId, form, debouncedUpdateSearch]);

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
                <div className="relative flex-1 overflow-y-hidden">
                    <div
                        className="flex h-full flex-col gap-4 overflow-y-auto"
                        ref={containerRef}
                    >
                        {filteredScreenshots.map((screenshot) => (
                            <FlowScreenshot
                                key={screenshot.id}
                                screenshot={screenshot}
                            />
                        ))}
                        <div ref={endRef} />
                    </div>

                    {!isScrolledToBottom && (
                        <Button
                            className="absolute right-4 bottom-4 z-10 shadow-md hover:shadow-lg"
                            onClick={() => scrollToEnd()}
                            size="icon-sm"
                            type="button"
                            variant="outline"
                        >
                            <ChevronDown />
                            {hasNewMessages && (
                                <span className="bg-primary absolute -top-1.5 -right-1.5 size-3 rounded-full" />
                            )}
                        </Button>
                    )}
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
}

export default FlowScreenshots;
