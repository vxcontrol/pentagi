import { zodResolver } from '@hookform/resolvers/zod';
import { Loader2, Send, Square } from 'lucide-react';
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

interface ChatAutomationFormInputProps {
    selectedFlowId: string | null;
    flowStatus?: StatusType;
    isCreatingFlow?: boolean;
    onSubmitMessage: (message: string) => Promise<void>;
    onStopFlow?: (flowId: string) => Promise<void>;
}

const ChatAutomationFormInput = ({
    selectedFlowId,
    flowStatus,
    isCreatingFlow = false,
    onSubmitMessage,
    onStopFlow,
}: ChatAutomationFormInputProps) => {
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [isStopping, setIsStopping] = useState(false);
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

        if (isCreatingFlow) {
            return 'Creating a new flow...';
        }

        if (selectedFlowId === 'new') {
            return 'Describe what you would like PentAGI to test...';
        }

        // Flow-specific statuses
        switch (flowStatus) {
            case StatusType.Waiting: {
                return 'Provide additional context or instructions...';
            }
            case StatusType.Running: {
                return 'PentAGI is working... Click Stop to interrupt';
            }
            case StatusType.Created: {
                return 'The flow is starting...';
            }
            case StatusType.Finished:
            case StatusType.Failed: {
                return 'This flow has ended. Create a new one to continue.';
            }
            default: {
                return 'Type your message...';
            }
        }
    };

    const handleSubmit = async (values: z.infer<typeof formSchema>) => {
        const message = values.message.trim();
        if (!message) {
            return;
        }

        try {
            setIsSubmitting(true);
            await onSubmitMessage(message);
            // Only reset form on success
            form.reset();
        } catch (error) {
            Log.error('Error submitting message:', error);
            // Don't reset form on error so user doesn't lose their message
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleStopFlow = async () => {
        if (!selectedFlowId || !onStopFlow) return;

        try {
            setIsStopping(true);
            await onStopFlow(selectedFlowId);
        } catch (error) {
            Log.error('Error stopping flow:', error);
        } finally {
            setIsStopping(false);
        }
    };

    const handleKeyDown = (event: React.KeyboardEvent) => {
        // Don't process keyboard shortcuts when flow is running or is being created
        if (flowStatus === StatusType.Running || isCreatingFlow) {
            return;
        }

        const isEnterPress = event.key === 'Enter';
        const isCtrlEnter = isEnterPress && (event.ctrlKey || event.metaKey);
        const isEnterOnly = isEnterPress && !event.shiftKey;

        if (!isEnterOnly && !isCtrlEnter) {
            return;
        }

        event.preventDefault();

        if (isInputDisabled) {
            return;
        }

        form.handleSubmit(handleSubmit)();
    };

    const isRunning = flowStatus === StatusType.Running;
    const isCreated = flowStatus === StatusType.Created;
    const isFlowTerminal = flowStatus === StatusType.Finished || flowStatus === StatusType.Failed;

    const isInputDisabled =
        !selectedFlowId ||
        isSubmitting ||
        isCreatingFlow ||
        isRunning ||
        isCreated ||
        isFlowTerminal;

    const isButtonDisabled =
        !selectedFlowId ||
        isSubmitting ||
        isCreatingFlow ||
        isStopping ||
        isCreated ||
        isFlowTerminal;

    // Auto-focus on textarea when needed
    useEffect(() => {
        if (
            !isInputDisabled &&
            (selectedFlowId === 'new' ||
                flowStatus === StatusType.Waiting ||
                !flowStatus)
        ) {
            const textarea = document.querySelector(`#${textareaId}`) as HTMLTextAreaElement;
            if (textarea) {
                const timeoutId = setTimeout(() => textarea.focus(), 0);
                return () => clearTimeout(timeoutId);
            }
        }
    }, [selectedFlowId, flowStatus, isInputDisabled]);

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
                                disabled={isInputDisabled}
                                onKeyDown={handleKeyDown}
                                className="resize-none"
                            />
                        </FormControl>
                    )}
                />
                {isRunning ? (
                    <Button
                        className="mb-px mt-auto"
                        type="button"
                        variant="destructive"
                        disabled={isButtonDisabled || isStopping}
                        onClick={handleStopFlow}
                    >
                        {isStopping ? <Loader2 className="size-4 animate-spin" /> : <Square className="size-4" />}
                        <span className="sr-only">Stop</span>
                    </Button>
                ) : (
                    <Button
                        className="mb-px mt-auto"
                        type="submit"
                        disabled={isButtonDisabled}
                    >
                        {isSubmitting || isCreatingFlow ? <Loader2 className="size-4 animate-spin" /> : <Send className="size-4" />}
                        <span className="sr-only">Send</span>
                    </Button>
                )}
            </form>
        </Form>
    );
};

export default ChatAutomationFormInput;
