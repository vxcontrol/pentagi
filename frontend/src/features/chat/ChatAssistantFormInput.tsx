import { zodResolver } from '@hookform/resolvers/zod';
import { Loader2, Send, Square, Users } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField } from '@/components/ui/form';
import { Textarea } from '@/components/ui/textarea';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { StatusType } from '@/graphql/types';
import { Log } from '@/lib/log';
import { useFlow } from '@/providers/FlowProvider';

const formSchema = z.object({
    message: z.string().min(1, { message: 'Message cannot be empty' }),
    useAgents: z.boolean().default(false),
});

interface ChatAssistantFormInputProps {
    assistantStatus?: StatusType;
    isCreatingAssistant?: boolean;
    isProviderAvailable?: boolean;
    isUseAgentsDefault?: boolean;
    onStopFlow?: (flowId: string) => Promise<void>;
    onSubmitMessage: (message: string, useAgents: boolean) => Promise<void>;
}

const ChatAssistantFormInput = ({
    assistantStatus,
    isCreatingAssistant = false,
    isProviderAvailable = true,
    isUseAgentsDefault = false,
    onStopFlow,
    onSubmitMessage,
}: ChatAssistantFormInputProps) => {
    const { flowId } = useFlow();
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [isStopping, setIsStopping] = useState(false);
    const [useAgents, setUseAgents] = useState(isUseAgentsDefault || false);
    const textareaId = 'chat-textarea';

    // Input is disabled in these scenarios:
    // 1. No flow selected
    // 2. Currently submitting or creating an assistant
    // 3. Assistant is running (not waiting)
    // 4. Provider is unavailable
    // 5. Assistant is in a terminal state (finished/failed)
    const isRunning = assistantStatus === StatusType.Running;
    const isCreated = assistantStatus === StatusType.Created;
    const isAssistantTerminal = assistantStatus === StatusType.Finished || assistantStatus === StatusType.Failed;

    const isInputDisabled =
        !flowId ||
        isSubmitting ||
        isCreatingAssistant ||
        isRunning ||
        isCreated ||
        !isProviderAvailable ||
        isAssistantTerminal;

    const isButtonDisabled =
        !flowId ||
        isSubmitting ||
        isCreatingAssistant ||
        isStopping ||
        isCreated ||
        !isProviderAvailable ||
        isAssistantTerminal;

    const form = useForm<z.infer<typeof formSchema>>({
        defaultValues: {
            message: '',
            useAgents,
        },
        resolver: zodResolver(formSchema),
    });

    // Reset form when flow ID changes
    useEffect(() => {
        form.reset({
            message: '',
            useAgents: isUseAgentsDefault,
        });
    }, [flowId, form, isUseAgentsDefault]);

    // Update local useAgents state when isUseAgentsDefault changes
    useEffect(() => {
        setUseAgents(isUseAgentsDefault || false);
    }, [isUseAgentsDefault]);

    // Update the form value when useAgents state changes
    useEffect(() => {
        form.setValue('useAgents', useAgents);
    }, [useAgents, form]);

    const getPlaceholderText = () => {
        if (!flowId) {
            return 'Select a flow...';
        }

        if (flowId === 'new') {
            return 'What would you like me to help you with?';
        }

        // Show creating assistant message while in creation mode
        if (isCreatingAssistant) {
            return 'Creating assistant...';
        }

        // Provider unavailable has highest priority message
        if (!isProviderAvailable) {
            return 'The selected provider is unavailable...';
        }

        // No assistant selected - prompt to create one
        if (!assistantStatus) {
            return 'Type a message to create a new assistant...';
        }

        // Assistant-specific statuses
        switch (assistantStatus) {
            case StatusType.Created: {
                return 'Assistant is starting...';
            }

            case StatusType.Failed:

            // eslint-disable-next-line no-fallthrough
            case StatusType.Finished: {
                return 'This assistant session has ended. Create a new one to continue.';
            }

            case StatusType.Running: {
                return 'Assistant is running... Click Stop to interrupt';
            }

            case StatusType.Waiting: {
                return 'Continue the conversation...';
            }

            default: {
                return 'Type your message...';
            }
        }
    };

    const handleSubmit = async (values: z.infer<typeof formSchema>) => {
        // Make sure we have a non-empty message
        const message = values.message.trim();

        if (!message) {
            return;
        }

        try {
            setIsSubmitting(true);
            await onSubmitMessage(message, values.useAgents);
            // Only reset the form on successful submission
            form.reset();
        } catch (error) {
            Log.error('Error submitting message:', error);
            // Don't reset the form on error so user doesn't lose their message
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
            // TODO: Add error notification UI
        } finally {
            setIsStopping(false);
        }
    };

    const handleKeyDown = (event: React.KeyboardEvent) => {
        // Don't process keyboard shortcuts when assistant is running
        if (assistantStatus === StatusType.Running || isCreatingAssistant) {
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

    // Auto-focus on textarea when needed
    useEffect(() => {
        if (!isInputDisabled && (flowId === 'new' || assistantStatus === StatusType.Waiting || !assistantStatus)) {
            const textarea = document.querySelector(`#${textareaId}`) as HTMLTextAreaElement;

            if (textarea) {
                const timeoutId = setTimeout(() => textarea.focus(), 100);

                return () => clearTimeout(timeoutId);
            }
        }
    }, [flowId, assistantStatus, isInputDisabled]);

    return (
        <Form {...form}>
            <form
                className="flex w-full items-center space-x-2"
                onSubmit={form.handleSubmit(handleSubmit)}
            >
                <TooltipProvider>
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button
                                className="mb-px mt-auto"
                                disabled={isInputDisabled}
                                onClick={() => setUseAgents(!useAgents)}
                                size="icon"
                                type="button"
                                variant={useAgents ? 'default' : 'outline'}
                            >
                                <Users className="size-4" />
                                <span className="sr-only">{useAgents ? 'Disable agents' : 'Enable agents'}</span>
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent>{useAgents ? 'Disable agents' : 'Enable agents'}</TooltipContent>
                    </Tooltip>
                </TooltipProvider>
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
                        {isSubmitting || isCreatingAssistant ? (
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

export default ChatAssistantFormInput;
