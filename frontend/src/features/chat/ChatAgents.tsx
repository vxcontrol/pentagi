import { zodResolver } from '@hookform/resolvers/zod';
import { Search, X } from 'lucide-react';
import { useEffect, useRef } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import type { AgentLogFragmentFragment } from '@/graphql/types';

import ChatAgent from './ChatAgent';

interface ChatAgentsProps {
    logs?: AgentLogFragmentFragment[];
}

const searchFormSchema = z.object({
    search: z.string(),
});

const ChatAgents = ({ logs }: ChatAgentsProps) => {
    const agentsEndRef = useRef<HTMLDivElement>(null);
    const scrollAgents = () => {
        agentsEndRef.current?.scrollIntoView();
    };

    const form = useForm<z.infer<typeof searchFormSchema>>({
        resolver: zodResolver(searchFormSchema),
        defaultValues: {
            search: '',
        },
    });

    const filteredLogs = logs?.filter((log) => {
        const search = form.watch('search').toLowerCase();

        if (!search) {
            return true;
        }

        return (
            log.task.toLowerCase().includes(search) ||
            log.result?.toLowerCase().includes(search) ||
            log.executor.toLowerCase().includes(search) ||
            log.initiator.toLowerCase().includes(search)
        );
    });

    useEffect(() => {
        scrollAgents();
    }, [logs]);

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
                                        placeholder="Search agent logs..."
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
                {filteredLogs?.map((log) => (
                    <ChatAgent
                        key={log.id}
                        log={log}
                    />
                ))}
                <div ref={agentsEndRef} />
            </div>
        </div>
    );
};

export default ChatAgents;
