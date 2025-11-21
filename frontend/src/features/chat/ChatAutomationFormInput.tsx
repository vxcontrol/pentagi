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
import { useFlow } from '@/providers/FlowProvider';

const formSchema = z.object({
    message: z.string().min(1, { message: 'Message cannot be empty' }),
});

interface ChatAutomationFormInputProps {
    flowStatus?: StatusType;
    isCreatingFlow?: boolean;
    onStopFlow?: (flowId: string) => Promise<void>;
    onSubmitMessage: (message: string) => Promise<void>;
}

const ChatAutomationFormInput = ({
    flowStatus,
    isCreatingFlow = false,
    onStopFlow,
    onSubmitMessage,
}: ChatAutomationFormInputProps) => {
    const { flowId } = useFlow();
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [isStopping, setIsStopping] = useState(false);
    const textareaId = 'chat-textarea';

    const form = useForm<z.infer<typeof formSchema>>({
        defaultValues: {
            message: '',
        },
        resolver: zodResolver(formSchema),
    });

    // Reset form when flow ID changes
    useEffect(() => {
        form.reset();
    }, [flowId, form]);

    const getPlaceholderText = () => {
        if (!flowId) {
            return 'Select a flow...';
        }

        if (isCreatingFlow) {
            return 'Creating a new flow...';
        }

        if (flowId === 'new') {
            return 'Describe what you would like PentAGI to test...';
        }

        // Flow-specific statuses
        switch (flowStatus) {
            case StatusType.Created: {
                return 'The flow is starting...';
            }

            case StatusType.Failed:

            // eslint-disable-next-line no-fallthrough
            case StatusType.Finished: {
                return 'This flow has ended. Create a new one to continue.';
            }

            case StatusType.Running: {
                return 'PentAGI is working... Click Stop to interrupt';
            }

            case StatusType.Waiting: {
                return 'Provide additional context or instructions...';
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
        if (!flowId || !onStopFlow) {
            return;
        }

        try {
            setIsStopping(true);
            await onStopFlow(flowId);
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

    const isInputDisabled = !flowId || isSubmitting || isCreatingFlow || isRunning || isCreated || isFlowTerminal;

    const isButtonDisabled = !flowId || isSubmitting || isCreatingFlow || isStopping || isCreated || isFlowTerminal;

    // Auto-focus on textarea when needed
    useEffect(() => {
        if (!isInputDisabled && (flowId === 'new' || flowStatus === StatusType.Waiting || !flowStatus)) {
            const textarea = document.querySelector(`#${textareaId}`) as HTMLTextAreaElement;

            if (textarea) {
                const timeoutId = setTimeout(() => textarea.focus(), 0);

                return () => clearTimeout(timeoutId);
            }
        }
    }, [flowId, flowStatus, isInputDisabled]);

    return (
        <Form {...form}>
            <form
                className="flex w-full items-center space-x-2"
                onSubmit={form.handleSubmit(handleSubmit)}
            >
                <FormField
                    control={form.control}
                    name="message"
                    render={({ field }) => (
                        <FormControl>
                            <Textarea
                                {...field}
                                className="resize-none"
                                disabled={isInputDisabled}
                                id={textareaId}
                                onKeyDown={handleKeyDown}
                                placeholder={getPlaceholderText()}
                            />
                        </FormControl>
                    )}
                />
                {isRunning ? (
                    <Button
                        className="mb-px mt-auto"
                        disabled={isButtonDisabled || isStopping}
                        onClick={handleStopFlow}
                        type="button"
                        variant="destructive"
                    >
                        {isStopping ? <Loader2 className="size-4 animate-spin" /> : <Square className="size-4" />}
                        <span className="sr-only">Stop</span>
                    </Button>
                ) : (
                    <Button
                        className="mb-px mt-auto"
                        disabled={isButtonDisabled}
                        type="submit"
                    >
                        {isSubmitting || isCreatingFlow ? (
                            <Loader2 className="size-4 animate-spin" />
                        ) : (
                            <Send className="size-4" />
                        )}
                        <span className="sr-only">Send</span>
                    </Button>
                )}
            </form>
        </Form>
    );
};

export default ChatAutomationFormInput;
