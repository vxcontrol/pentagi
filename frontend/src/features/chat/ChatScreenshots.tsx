import { zodResolver } from '@hookform/resolvers/zod';
import { Search, X } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import type { ScreenshotFragmentFragment } from '@/graphql/types';

import ChatScreenshot from './ChatScreenshot';

interface ChatScreenshotsProps {
    screenshots: ScreenshotFragmentFragment[];
}

const searchFormSchema = z.object({
    search: z.string(),
});

const ChatScreenshots = ({ screenshots }: ChatScreenshotsProps) => {
    const form = useForm<z.infer<typeof searchFormSchema>>({
        resolver: zodResolver(searchFormSchema),
        defaultValues: {
            search: '',
        },
    });

    const filteredScreenshots = screenshots?.filter((screenshot) => {
        const search = form.watch('search').toLowerCase();

        if (!search) {
            return true;
        }

        return screenshot.url.toLowerCase().includes(search);
    });

    return (
        <div className="flex flex-col">
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
                                        type="text"
                                        placeholder="Search screenshots..."
                                        className="px-9"
                                    />
                                    {field.value && (
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            className="absolute right-0 top-1/2 -translate-y-1/2"
                                            onClick={() => form.reset({ search: '' })}
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

            {filteredScreenshots?.length ? (
                <div className="space-y-4 pb-4">
                    {filteredScreenshots.map((screenshot) => (
                        <ChatScreenshot
                            key={screenshot.id}
                            screenshot={screenshot}
                        />
                    ))}
                </div>
            ) : (
                <p className="m-auto text-lg font-semibold text-muted-foreground">No screenshots available</p>
            )}
        </div>
    );
};

export default ChatScreenshots;
