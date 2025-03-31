import { zodResolver } from '@hookform/resolvers/zod';
import { Loader2, Send } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { Textarea } from '@/components/ui/textarea';
import { StatusType } from '@/graphql/types';
import { Log } from '@/lib/log';

const formSchema = z.object({
    message: z.string().min(1, { message: 'Message cannot be empty' }),
});

interface ChatFormProps {
    selectedFlowId: string | null;
    flowStatus?: StatusType;
    onSubmit: (message: string) => Promise<void>;
}

const ChatForm = ({ selectedFlowId, flowStatus, onSubmit }: ChatFormProps) => {
    const [isSubmitting, setIsSubmitting] = useState(false);
    const textareaId = 'chat-textarea';

    const form = useForm<z.infer<typeof formSchema>>({
        resolver: zodResolver(formSchema),
        defaultValues: {
            message: '',
        },
    });

    // Reset form when flow ID changes
    useEffect(() => {
        form.reset();
    }, [selectedFlowId, form]);

    const getPlaceholderText = () => {
        if (!selectedFlowId) {
            return 'Select a flow...';
        }

        if (selectedFlowId === 'new') {
            return 'What would you like me to help you with?';
        }

        if (flowStatus === StatusType.Waiting) {
            return 'Continue the conversation...';
        }

        if (flowStatus === StatusType.Finished) {
            return 'The flow is finished';
        }

        return 'Type your message...';
    };

    const handleSubmit = async (values: z.infer<typeof formSchema>) => {
        try {
            setIsSubmitting(true);
            await onSubmit(values.message);
            form.reset();
        } catch (error) {
            Log.error('Error submitting message:', error);
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleKeyDown = (event: React.KeyboardEvent) => {
        const isEnterPress = event.key === 'Enter';
        const isCtrlEnter = isEnterPress && (event.ctrlKey || event.metaKey);
        const isEnterOnly = isEnterPress && !event.shiftKey;

        if (!isEnterOnly && !isCtrlEnter) {
            return;
        }

        event.preventDefault();

        if (isDisabled) {
            return;
        }

        form.handleSubmit(handleSubmit)();
    };

    const isDisabled =
        !selectedFlowId ||
        isSubmitting ||
        (selectedFlowId !== 'new' && (!flowStatus || ![StatusType.Waiting].includes(flowStatus)));

    // Auto-focus on textarea when needed
    useEffect(() => {
        if (
            !isDisabled &&
            (selectedFlowId === 'new' || flowStatus === StatusType.Waiting)
        ) {
            const textarea = document.querySelector(`#${textareaId}`) as HTMLTextAreaElement;
            if (textarea) {
                const timeoutId = setTimeout(() => textarea.focus(), 0);
                return () => clearTimeout(timeoutId);
            }
        }
    }, [selectedFlowId, flowStatus, isDisabled]);

    return (
        <Form {...form}>
            <form
                onSubmit={form.handleSubmit(handleSubmit)}
                className="flex w-full items-center space-x-2"
            >
                <FormField
                    control={form.control}
                    name="message"
                    render={({ field }) => (
                        <FormControl>
                            <Textarea
                                {...field}
                                id={textareaId}
                                placeholder={getPlaceholderText()}
                                disabled={isDisabled}
                                onKeyDown={handleKeyDown}
                                className="resize-none"
                            />
                        </FormControl>
                    )}
                />
                <Button
                    className="mb-px mt-auto"
                    type="submit"
                    disabled={isDisabled}
                >
                    {isSubmitting ? <Loader2 className="size-4 animate-spin" /> : <Send className="size-4" />}
                    <span className="sr-only">Send</span>
                </Button>
            </form>
        </Form>
    );
};

export default ChatForm;
