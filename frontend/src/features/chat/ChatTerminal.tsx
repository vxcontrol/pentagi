import '@xterm/xterm/css/xterm.css';

import { zodResolver } from '@hookform/resolvers/zod';
import { Search, X } from 'lucide-react';
import { useForm, useWatch } from 'react-hook-form';
import { z } from 'zod';

import Terminal from '@/components/Terminal';
import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import type { TerminalLogFragmentFragment } from '@/graphql/types';

const searchFormSchema = z.object({
    search: z.string(),
});

interface ChatTerminalProps {
    logs: TerminalLogFragmentFragment[];
}

const ChatTerminal = ({ logs: terminalLog }: ChatTerminalProps) => {
    const form = useForm<z.infer<typeof searchFormSchema>>({
        resolver: zodResolver(searchFormSchema),
        defaultValues: {
            search: '',
        },
    });

    const search = useWatch({
        control: form.control,
        name: 'search',
    });

    const logs = terminalLog.map((log) => log.text);
    const filteredLogs = search ? logs?.filter((log) => log.toLowerCase().includes(search.toLowerCase())) : logs;

    return (
        <div className="flex size-full flex-col gap-4">
            <div className="sticky top-0 z-10 bg-background pr-6">
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
                                        className="px-9"
                                    />
                                    {field.value && (
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            className="absolute right-px top-1/2 -translate-y-1/2"
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
            <Terminal
                logs={filteredLogs}
                className="w-full grow"
            />
        </div>
    );
};

export default ChatTerminal;
