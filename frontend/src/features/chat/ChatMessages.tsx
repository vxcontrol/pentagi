import { zodResolver } from '@hookform/resolvers/zod';
import { Search, X } from 'lucide-react';
import { memo, useCallback, useEffect, useMemo, useRef } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import type { MessageLogFragmentFragment } from '@/graphql/types';
import { cn } from '@/lib/utils';

import ChatMessage from './ChatMessage';

interface ChatMessagesProps {
    logs?: MessageLogFragmentFragment[];
    className?: string;
}

const searchFormSchema = z.object({
    search: z.string(),
});

const ChatMessages = ({ logs, className }: ChatMessagesProps) => {
    const messagesEndRef = useRef<HTMLDivElement>(null);

    // Memoize the scroll function to avoid recreating it on every render
    const scrollMessages = useCallback(() => {
        if (messagesEndRef.current) {
            messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
        }
    }, []);

    const form = useForm<z.infer<typeof searchFormSchema>>({
        resolver: zodResolver(searchFormSchema),
        defaultValues: {
            search: '',
        },
    });

    const searchValue = form.watch('search');

    // Memoize filtered logs to avoid recomputing on every render
    const filteredLogs = useMemo(() => {
        const search = searchValue.toLowerCase();

        if (!search || !logs) {
            return logs || [];
        }

        return logs.filter(
            (log) =>
                log.message.toLowerCase().includes(search) || (log.result && log.result.toLowerCase().includes(search)),
        );
    }, [logs, searchValue]);

    useEffect(() => {
        // Only scroll when new messages come in, not on every render
        const timeout = setTimeout(() => {
            scrollMessages();
        }, 100);

        return () => clearTimeout(timeout);
    }, [filteredLogs.length, scrollMessages]);

    return (
        <div className={cn('flex flex-col', className)}>
            <div className="sticky top-0 z-10 bg-background py-4">
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
                                        type="text"
                                        placeholder="Search messages..."
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
            <div className="space-y-4 pb-4">
                {filteredLogs.map((log) => (
                    <ChatMessage
                        key={log.id}
                        log={log}
                    />
                ))}
                <div ref={messagesEndRef} />
            </div>
        </div>
    );
};

// Using React.memo to prevent unnecessary rerenders
export default memo(ChatMessages);
