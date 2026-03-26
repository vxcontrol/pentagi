import { zodResolver } from '@hookform/resolvers/zod';
import {
    AlertCircle,
    Bot,
    CheckCircle,
    Code,
    FileDiff,
    Loader2,
    RotateCcw,
    Save,
    User,
    Wrench,
    XCircle,
} from 'lucide-react';
import * as React from 'react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import ReactDiffViewer from 'react-diff-viewer-continued';
import { useController, useForm, useFormState } from 'react-hook-form';
import { useNavigate, useParams } from 'react-router-dom';
import { z } from 'zod';

import type { AgentPrompt, AgentPrompts, DefaultPrompt, PromptType } from '@/graphql/types';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Form, FormControl, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { StatusCard } from '@/components/ui/status-card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import {
    useCreatePromptMutation,
    useDeletePromptMutation,
    useSettingsPromptsQuery,
    useUpdatePromptMutation,
    useValidatePromptMutation,
} from '@/graphql/types';
import { cn } from '@/lib/utils';

// Form schemas for each tab
const systemFormSchema = z.object({
    template: z.string().min(1, 'System template is required'),
});

const humanFormSchema = z.object({
    template: z.string().min(1, 'Human template is required'),
});

interface BaseFieldProps extends ControllerProps {
    label?: string;
}
interface BaseTextareaProps {
    className?: string;
    placeholder?: string;
}

// Universal field components using useController
interface ControllerProps {
    control: any;
    disabled?: boolean;
    name: string;
}

interface FormTextareaItemProps extends BaseFieldProps, BaseTextareaProps {
    description?: string;
}

type HumanFormData = z.infer<typeof humanFormSchema>;

type SystemFormData = z.infer<typeof systemFormSchema>;

const FormTextareaItem: React.FC<FormTextareaItemProps> = ({
    className,
    control,
    disabled,
    label,
    name,
    placeholder,
}) => {
    const { field, fieldState } = useController({
        control,
        defaultValue: '',
        disabled,
        name,
    });

    return (
        <FormItem>
            {label && <FormLabel>{label}</FormLabel>}
            <FormControl>
                <Textarea
                    {...field}
                    className={cn('min-h-[640px]! font-mono text-sm', className)}
                    disabled={disabled}
                    placeholder={placeholder}
                />
            </FormControl>
            {fieldState.error && <FormMessage>{fieldState.error.message}</FormMessage>}
        </FormItem>
    );
};

// Helper function to format display name
const formatName = (key: string): string => {
    return key.replaceAll(/([A-Z])/g, ' $1').replace(/^./, (str) => str.toUpperCase());
};

// Helper function to extract used variables from template
const getUsedVariables = (template: string | undefined): Set<string> => {
    const usedVariables = new Set<string>();

    if (!template) {
        return usedVariables;
    }

    const variableRegex = /\{\{\.(\w+)\}\}/g;
    let match;

    while ((match = variableRegex.exec(template)) !== null) {
        const variable = match[1];

        if (variable) {
            usedVariables.add(variable);
        }
    }

    return usedVariables;
};

// Variables Component
interface VariablesProps {
    currentTemplate: string;
    onVariableClick: (variable: string) => void;
    variables: string[];
}

const Variables: React.FC<VariablesProps> = ({ currentTemplate, onVariableClick, variables }) => {
    if (variables.length === 0) {
        return null;
    }

    const usedVariables = getUsedVariables(currentTemplate);

    return (
        <div className="bg-muted/50 mb-4 rounded-md border p-3">
            <h4 className="text-muted-foreground mb-2 text-sm font-medium">Available Variables:</h4>
            <div className="flex flex-wrap gap-1">
                {variables.map((variable) => {
                    const isUsed = usedVariables.has(variable);

                    return (
                        <code
                            className={`cursor-pointer rounded border px-2 py-1 font-mono text-xs transition-colors ${
                                isUsed
                                    ? 'border-green-300 bg-green-100 text-green-800 hover:bg-green-200'
                                    : 'bg-background text-foreground hover:bg-accent'
                            }`}
                            key={variable}
                            onClick={() => onVariableClick(variable)}
                        >
                            {`{{.${variable}}}`}
                        </code>
                    );
                })}
            </div>
        </div>
    );
};

const SettingsPrompt = () => {
    const { promptId } = useParams<{ promptId: string }>();
    const navigate = useNavigate();

    // GraphQL queries and mutations
    const { data, error, loading } = useSettingsPromptsQuery();
    const [createPrompt, { error: createError, loading: isCreateLoading }] = useCreatePromptMutation();
    const [updatePrompt, { error: updateError, loading: isUpdateLoading }] = useUpdatePromptMutation();
    const [deletePrompt, { error: deleteError, loading: isDeleteLoading }] = useDeletePromptMutation();
    const [validatePrompt, { error: validateError, loading: isValidateLoading }] = useValidatePromptMutation();

    // Local state management
    const [submitError, setSubmitError] = useState<null | string>(null);
    const [activeTab, setActiveTab] = useState<'human' | 'system'>('system');
    const [resetDialogOpen, setResetDialogOpen] = useState(false);
    const [validationResult, setValidationResult] = useState<any>(null);
    const [validationDialogOpen, setValidationDialogOpen] = useState(false);
    const [isDiffDialogOpen, setIsDiffDialogOpen] = useState(false);
    const [isLeaveDialogOpen, setIsLeaveDialogOpen] = useState(false);
    const [pendingBrowserBack, setPendingBrowserBack] = useState(false);
    const allowBrowserLeaveRef = useRef(false);
    const hasPushedBlockerStateRef = useRef(false);

    const isLoading = isCreateLoading || isUpdateLoading || isDeleteLoading || isValidateLoading;

    // Helper function to handle variable insertion/selection
    const handleVariableClick = (variable: string, field: any, formId: string) => {
        const textarea = document.querySelector(`#${formId} textarea`) as HTMLTextAreaElement;

        if (textarea) {
            const currentValue = field.value || '';
            const variablePattern = `{{.${variable}}}`;

            // Check if variable is already used
            const variableIndex = currentValue.indexOf(variablePattern);

            if (variableIndex !== -1) {
                // Variable exists - select it and scroll to it
                textarea.focus();
                textarea.setSelectionRange(variableIndex, variableIndex + variablePattern.length);

                // Scroll to center the selection
                const lineHeight = 20; // Approximate line height
                const textBeforeSelection = currentValue.slice(0, Math.max(0, variableIndex));
                const linesBeforeSelection = textBeforeSelection.split('\n').length - 1;
                const selectionTop = linesBeforeSelection * lineHeight;
                const textareaHeight = textarea.clientHeight;
                const scrollTop = Math.max(0, selectionTop - textareaHeight / 2);

                textarea.scrollTop = scrollTop;
            } else {
                // Variable doesn't exist - insert it at cursor position (no scrolling)
                const start = textarea.selectionStart;
                const end = textarea.selectionEnd;
                const newValue =
                    currentValue.slice(0, Math.max(0, start)) + variablePattern + currentValue.slice(Math.max(0, end));
                field.onChange(newValue);

                // Focus and set cursor position after the inserted variable (no scrolling)
                setTimeout(() => {
                    textarea.focus({ preventScroll: true });
                    textarea.setSelectionRange(start + variablePattern.length, start + variablePattern.length);
                }, 0);
            }
        }
    };

    // Handle reset to default prompt
    const handleReset = () => {
        setResetDialogOpen(true);
    };

    const handleConfirmReset = async () => {
        if (!promptInfo) {
            return;
        }

        try {
            setSubmitError(null);

            if (activeTab === 'system' && promptInfo.userSystemPrompt) {
                await deletePrompt({
                    refetchQueries: ['settingsPrompts'],
                    variables: { promptId: promptInfo.userSystemPrompt.id },
                });
                // Reset form to default value
                systemForm.setValue('template', promptInfo.defaultSystemTemplate);
            } else if (activeTab === 'human' && promptInfo.userHumanPrompt) {
                await deletePrompt({
                    refetchQueries: ['settingsPrompts'],
                    variables: { promptId: promptInfo.userHumanPrompt.id },
                });
                // Reset form to default value
                humanForm.setValue('template', promptInfo.defaultHumanTemplate);
            }

            setResetDialogOpen(false);
        } catch (error) {
            console.error('Reset error:', error);
            setSubmitError(error instanceof Error ? error.message : 'An error occurred while resetting');
            setResetDialogOpen(false);
        }
    };

    // Handle validate prompt
    const handleValidate = async () => {
        if (!promptInfo) {
            return;
        }

        try {
            setSubmitError(null);
            setValidationResult(null);

            let promptType: PromptType;
            let currentTemplate: string;

            if (activeTab === 'system') {
                if (promptInfo.type === 'agent') {
                    const agentData = promptInfo.data as AgentPrompt | AgentPrompts;
                    promptType = agentData.system.type;
                } else {
                    const toolData = promptInfo.data as DefaultPrompt;
                    promptType = toolData.type;
                }

                currentTemplate = systemTemplate;
            } else {
                const agentData = promptInfo.data as AgentPrompts;
                promptType = agentData.human!.type;
                currentTemplate = humanTemplate;
            }

            const result = await validatePrompt({
                variables: {
                    template: currentTemplate,
                    type: promptType,
                },
            });

            setValidationResult(result.data?.validatePrompt);
            setValidationDialogOpen(true);
        } catch (error) {
            console.error('Validation error:', error);
            setSubmitError(error instanceof Error ? error.message : 'An error occurred while validating');
        }
    };

    // Form instances for each tab
    const systemForm = useForm<SystemFormData>({
        defaultValues: {
            template: '',
        },
        resolver: zodResolver(systemFormSchema),
    });

    const humanForm = useForm<HumanFormData>({
        defaultValues: {
            template: '',
        },
        resolver: zodResolver(humanFormSchema),
    });

    // Reactive dirty state across both forms
    const { isDirty: isSystemDirty } = useFormState({ control: systemForm.control });
    const { isDirty: isHumanDirty } = useFormState({ control: humanForm.control });
    const isDirty = isSystemDirty || isHumanDirty;

    // Watch form values to detect used variables
    const systemTemplate = systemForm.watch('template');
    const humanTemplate = humanForm.watch('template');

    // Determine prompt type and get prompt data
    const promptInfo = useMemo(() => {
        if (!promptId || !data?.settingsPrompts) {
            return null;
        }

        const { default: defaultPrompts, userDefined } = data.settingsPrompts;

        if (!defaultPrompts) {
            return null;
        }

        const { agents, tools } = defaultPrompts;

        // Check if it's an agent prompt
        const agentData = agents?.[promptId as keyof typeof agents] as AgentPrompt | AgentPrompts | undefined;

        if (agentData) {
            // Check if we have user-defined system or human prompts
            const userSystemPrompt = userDefined?.find((p) => p.type === agentData.system.type);
            const userHumanPrompt = userDefined?.find((p) => p.type === (agentData as AgentPrompts)?.human?.type);

            return {
                data: agentData,
                defaultHumanTemplate: (agentData as AgentPrompts)?.human?.template || '',
                defaultSystemTemplate: agentData?.system?.template || '',
                displayName: formatName(promptId),
                hasHuman: !!(agentData as AgentPrompts)?.human,
                humanTemplate: userHumanPrompt?.template || (agentData as AgentPrompts)?.human?.template || '',
                systemTemplate: userSystemPrompt?.template || agentData?.system?.template || '',
                type: 'agent' as const,
                userHumanPrompt,
                userSystemPrompt,
            };
        }

        // Check if it's a tool prompt
        const toolData = tools?.[promptId as keyof typeof tools] as DefaultPrompt | undefined;

        if (toolData) {
            const userToolPrompt = userDefined?.find((p) => p.type === toolData.type);

            return {
                data: toolData,
                defaultHumanTemplate: '',
                defaultSystemTemplate: toolData?.template || '',
                displayName: formatName(promptId),
                hasHuman: false,
                humanTemplate: '',
                systemTemplate: userToolPrompt?.template || toolData?.template || '',
                type: 'tool' as const,
                userHumanPrompt: null,
                userSystemPrompt: userToolPrompt,
            };
        }

        return null;
    }, [promptId, data?.settingsPrompts]);

    // Compute variables data based on active tab and prompt info
    const variablesData = useMemo(() => {
        if (!promptInfo) {
            return null;
        }

        let variables: string[] = [];
        let formId = '';
        let currentTemplate = '';

        if (activeTab === 'system') {
            variables =
                promptInfo.type === 'agent'
                    ? (promptInfo.data as AgentPrompt | AgentPrompts)?.system?.variables || []
                    : (promptInfo.data as DefaultPrompt)?.variables || [];
            formId = 'system-prompt-form';
            currentTemplate = systemTemplate;
        } else if (activeTab === 'human' && promptInfo.type === 'agent' && promptInfo.hasHuman) {
            variables = (promptInfo.data as AgentPrompts)?.human?.variables || [];
            formId = 'human-prompt-form';
            currentTemplate = humanTemplate;
        }

        return { currentTemplate, formId, variables };
    }, [promptInfo, activeTab, systemTemplate, humanTemplate]);

    // Handle variable click with useCallback for better performance
    const handleVariableClickCallback = useCallback(
        (variable: string) => {
            if (!variablesData) {
                return;
            }

            const field =
                activeTab === 'system'
                    ? {
                          onChange: (value: string) => systemForm.setValue('template', value),
                          value: systemTemplate,
                      }
                    : {
                          onChange: (value: string) => humanForm.setValue('template', value),
                          value: humanTemplate,
                      };
            handleVariableClick(variable, field, variablesData.formId);
        },
        [activeTab, systemTemplate, humanTemplate, variablesData, systemForm, humanForm],
    );

    // Fill forms with current prompt data when available
    useEffect(() => {
        if (promptInfo) {
            systemForm.reset({
                template: promptInfo.systemTemplate,
            });
            humanForm.reset({
                template: promptInfo.humanTemplate,
            });
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [promptInfo]);

    // Push a blocker entry when form is dirty to manage browser back
    useEffect(() => {
        if (isDirty && !hasPushedBlockerStateRef.current) {
            window.history.pushState({ __pentagiBlock__: true }, '');
            hasPushedBlockerStateRef.current = true;
        }
    }, [isDirty]);

    // Intercept browser back to show confirmation dialog
    useEffect(() => {
        const handlePopState = () => {
            if (!isDirty) {
                return;
            }

            if (allowBrowserLeaveRef.current) {
                allowBrowserLeaveRef.current = false;

                return;
            }

            setPendingBrowserBack(true);
            setIsLeaveDialogOpen(true);
            window.history.forward();
        };

        window.addEventListener('popstate', handlePopState, { capture: true });

        return () => {
            window.removeEventListener('popstate', handlePopState, { capture: true } as any);
        };
    }, [isDirty]);

    const handleBack = () => {
        if (isDirty) {
            setIsLeaveDialogOpen(true);

            return;
        }

        navigate('/settings/prompts');
    };

    const handleConfirmLeave = () => {
        if (pendingBrowserBack) {
            allowBrowserLeaveRef.current = true;
            setPendingBrowserBack(false);
            window.history.go(-2);

            return;
        }

        navigate('/settings/prompts');
    };

    const handleLeaveDialogOpenChange = (open: boolean) => {
        if (!open && pendingBrowserBack) {
            setPendingBrowserBack(false);
        }

        setIsLeaveDialogOpen(open);
    };

    // Form submission handlers
    const handleSystemSubmit = async (formData: SystemFormData) => {
        if (!promptInfo) {
            return;
        }

        const isUpdate = !!promptInfo.userSystemPrompt;

        // For creation, check if the template is identical to the default
        if (!isUpdate && formData.template === promptInfo.defaultSystemTemplate) {

            return;
        }

        try {
            setSubmitError(null);

            // Get the real type from data
            let promptType: PromptType;

            if (promptInfo.type === 'agent') {
                const agentData = promptInfo.data as AgentPrompt | AgentPrompts;
                promptType = agentData.system.type;
            } else {
                const toolData = promptInfo.data as DefaultPrompt;
                promptType = toolData.type;
            }

            if (isUpdate) {
                // Update existing user-defined prompt
                await updatePrompt({
                    refetchQueries: ['settingsPrompts'],
                    variables: {
                        promptId: promptInfo.userSystemPrompt!.id,
                        template: formData.template,
                    },
                });
            } else {
                // Create new user-defined prompt
                await createPrompt({
                    refetchQueries: ['settingsPrompts'],
                    variables: {
                        template: formData.template,
                        type: promptType,
                    },
                });
            }
        } catch (error) {
            console.error('Submit error:', error);
            setSubmitError(error instanceof Error ? error.message : 'An error occurred while saving');
        }
    };

    const handleHumanSubmit = async (formData: HumanFormData) => {
        if (!promptInfo) {
            return;
        }

        const isUpdate = !!promptInfo.userHumanPrompt;

        // For creation, check if the template is identical to the default
        if (!isUpdate && formData.template === promptInfo.defaultHumanTemplate) {

            return;
        }

        try {
            setSubmitError(null);

            // Get the real human prompt type from data
            const agentData = promptInfo.data as AgentPrompts;
            const humanPromptType = agentData.human?.type;

            if (!humanPromptType) {
                setSubmitError('Human prompt type not found');

                return;
            }

            if (isUpdate) {
                // Update existing user-defined prompt
                await updatePrompt({
                    refetchQueries: ['settingsPrompts'],
                    variables: {
                        promptId: promptInfo.userHumanPrompt!.id,
                        template: formData.template,
                    },
                });
            } else {
                // Create new user-defined prompt
                await createPrompt({
                    refetchQueries: ['settingsPrompts'],
                    variables: {
                        template: formData.template,
                        type: humanPromptType,
                    },
                });
            }
        } catch (error) {
            console.error('Submit error:', error);
            setSubmitError(error instanceof Error ? error.message : 'An error occurred while saving');
        }
    };

    // Loading state
    if (loading) {
        return (
            <StatusCard
                description="Please wait while we fetch prompt information"
                icon={<Loader2 className="text-muted-foreground size-16 animate-spin" />}
                title="Loading prompt data..."
            />
        );
    }

    // Error state
    if (error) {
        return (
            <Alert variant="destructive">
                <AlertCircle className="size-4" />
                <AlertTitle>Error loading prompt data</AlertTitle>
                <AlertDescription>{error.message}</AlertDescription>
            </Alert>
        );
    }

    // Prompt not found state
    if (!promptInfo) {
        return (
            <Alert variant="destructive">
                <AlertCircle className="size-4" />
                <AlertTitle>Prompt not found</AlertTitle>
                <AlertDescription>
                    The prompt "{promptId}" could not be found or is not supported for editing.
                </AlertDescription>
            </Alert>
        );
    }

    // Templates for diff based on active tab
    const currentTemplate = activeTab === 'system' ? systemTemplate : humanTemplate;
    const defaultTemplate = activeTab === 'system' ? promptInfo.defaultSystemTemplate : promptInfo.defaultHumanTemplate;

    // Styles for ReactDiffViewer aligned with shadcn (Tailwind CSS vars)
    const diffStyles = {
        content: {
            fontFamily:
                'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
            fontSize: '0.875rem',
            width: '50%',
        },
        diffContainer: {
            border: '1px solid var(--border)',
            borderRadius: '0.5rem',
        },
        gutter: {
            borderRight: '1px solid var(--border)',
        },
        line: {
            borderBottom: '1px solid oklch(from var(--border) l c h / 0.50)',
        },
        lineNumber: {
            color: 'var(--muted-foreground)',
        },
        splitView: {
            gap: '0',
        },
        variables: {
            dark: {
                addedBackground: 'hsl(142 70% 45% / 0.50)',
                addedColor: 'var(--foreground)',
                addedGutterBackground: 'hsl(142 70% 45% / 0.40)',
                addedGutterColor: 'var(--muted-foreground)',
                codeFoldBackground: 'var(--muted)',
                codeFoldContentColor: 'var(--muted-foreground)',
                codeFoldGutterBackground: 'var(--muted)',
                diffViewerBackground: 'var(--background)',
                diffViewerColor: 'var(--foreground)',
                diffViewerTitleBackground: 'var(--card)',
                diffViewerTitleBorderColor: 'var(--border)',
                diffViewerTitleColor: 'var(--card-foreground)',
                emptyLineBackground: 'var(--background)',
                gutterBackground: 'var(--muted)',
                gutterBackgroundDark: 'var(--muted)',
                gutterColor: 'var(--muted-foreground)',
                highlightBackground: 'oklch(from var(--primary) l c h / 0.20)',
                highlightGutterBackground: 'oklch(from var(--primary) l c h / 0.30)',
                removedBackground: 'oklch(from var(--destructive) l c h / 0.50)',
                removedColor: 'var(--foreground)',
                removedGutterBackground: 'oklch(from var(--destructive) l c h / 0.40)',
                removedGutterColor: 'var(--muted-foreground)',
                wordAddedBackground: 'hsl(142 70% 45% / 0.70)',
                wordRemovedBackground: 'oklch(from var(--destructive) l c h / 0.70)',
            },
            light: {
                addedBackground: 'hsl(142 70% 45% / 0.50)',
                addedColor: 'var(--foreground)',
                addedGutterBackground: 'hsl(142 70% 45% / 0.40)',
                addedGutterColor: 'var(--muted-foreground)',
                codeFoldBackground: 'var(--muted)',
                codeFoldContentColor: 'var(--muted-foreground)',
                codeFoldGutterBackground: 'var(--muted)',
                diffViewerBackground: 'var(--background)',
                diffViewerColor: 'var(--foreground)',
                diffViewerTitleBackground: 'var(--card)',
                diffViewerTitleBorderColor: 'var(--border)',
                diffViewerTitleColor: 'var(--card-foreground)',
                emptyLineBackground: 'var(--background)',
                gutterBackground: 'var(--muted)',
                gutterBackgroundDark: 'var(--muted)',
                gutterColor: 'var(--muted-foreground)',
                highlightBackground: 'oklch(from var(--primary) l c h / 0.20)',
                highlightGutterBackground: 'oklch(from var(--primary) l c h / 0.30)',
                removedBackground: 'oklch(from var(--destructive) l c h / 0.50)',
                removedColor: 'var(--foreground)',
                removedGutterBackground: 'oklch(from var(--destructive) l c h / 0.40)',
                removedGutterColor: 'var(--muted-foreground)',
                wordAddedBackground: 'hsl(142 70% 45% / 0.70)',
                wordRemovedBackground: 'oklch(from var(--destructive) l c h / 0.70)',
            },
        },
    };

    const mutationError = createError || updateError || deleteError || validateError || submitError;

    return (
        <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-2">
                <h2 className="flex items-center gap-2 text-lg font-semibold">
                    {promptInfo.type === 'agent' ? (
                        <Bot className="text-muted-foreground size-5" />
                    ) : (
                        <Wrench className="text-muted-foreground size-5" />
                    )}
                    {promptInfo.displayName}
                </h2>

                <div className="text-muted-foreground">
                    {promptInfo.type === 'agent'
                        ? 'Configure prompts for this AI agent'
                        : 'Configure the prompt for this tool'}
                </div>
            </div>

            <Tabs
                className="w-full"
                defaultValue="system"
                onValueChange={(value) => setActiveTab(value as 'human' | 'system')}
            >
                <TabsList>
                    <TabsTrigger value="system">
                        <div className="flex items-center gap-2">
                            <Code className="size-4" />
                            System Prompt
                        </div>
                    </TabsTrigger>
                    {promptInfo.type === 'agent' && promptInfo.hasHuman && (
                        <TabsTrigger value="human">
                            <div className="flex items-center gap-2">
                                <User className="size-4" />
                                Human Prompt
                            </div>
                        </TabsTrigger>
                    )}
                </TabsList>

                <TabsContent
                    className="mt-4"
                    value="system"
                >
                    <Form {...systemForm}>
                        <form
                            className="flex flex-col gap-6"
                            id="system-prompt-form"
                            onSubmit={systemForm.handleSubmit(handleSystemSubmit)}
                        >
                            {/* Error Alert */}
                            {mutationError && (
                                <Alert variant="destructive">
                                    <AlertCircle className="size-4" />
                                    <AlertTitle>Error</AlertTitle>
                                    <AlertDescription>
                                        {mutationError instanceof Error ? (
                                            mutationError.message
                                        ) : (
                                            <div className="whitespace-pre-line">{mutationError}</div>
                                        )}
                                    </AlertDescription>
                                </Alert>
                            )}

                            {/* System Template Field */}
                            <FormTextareaItem
                                control={systemForm.control}
                                disabled={isLoading}
                                name="template"
                                placeholder={
                                    promptInfo.type === 'tool'
                                        ? 'Enter the tool template...'
                                        : 'Enter the system prompt template...'
                                }
                            />
                        </form>
                    </Form>
                </TabsContent>

                {promptInfo.type === 'agent' && promptInfo.hasHuman && (
                    <TabsContent
                        className="mt-6"
                        value="human"
                    >
                        <Form {...humanForm}>
                            <form
                                className="flex flex-col gap-6"
                                id="human-prompt-form"
                                onSubmit={humanForm.handleSubmit(handleHumanSubmit)}
                            >
                                {/* Error Alert */}
                                {mutationError && (
                                    <Alert variant="destructive">
                                        <AlertCircle className="size-4" />
                                        <AlertTitle>Error</AlertTitle>
                                        <AlertDescription>
                                            {mutationError instanceof Error ? (
                                                mutationError.message
                                            ) : (
                                                <div className="whitespace-pre-line">{mutationError}</div>
                                            )}
                                        </AlertDescription>
                                    </Alert>
                                )}

                                {/* Human Template Field */}
                                <FormTextareaItem
                                    control={humanForm.control}
                                    disabled={isLoading}
                                    name="template"
                                    placeholder="Enter the human prompt template..."
                                />
                            </form>
                        </Form>
                    </TabsContent>
                )}
            </Tabs>

            {/* Sticky footer with variables and buttons */}
            <div className="bg-background sticky -bottom-4 -mx-4 mt-4 -mb-4 border-t p-4 shadow-lg">
                {/* Variables */}
                {variablesData && (
                    <Variables
                        currentTemplate={variablesData.currentTemplate}
                        onVariableClick={handleVariableClickCallback}
                        variables={variablesData.variables}
                    />
                )}

                {/* Action buttons */}
                <div className="flex items-center">
                    <div className="flex gap-2">
                        {/* Reset button - only show when user has custom prompt */}
                        {((activeTab === 'system' && promptInfo?.userSystemPrompt) ||
                            (activeTab === 'human' && promptInfo?.userHumanPrompt)) && (
                            <>
                                <Button
                                    disabled={isLoading}
                                    onClick={handleReset}
                                    type="button"
                                    variant="destructive"
                                >
                                    {isDeleteLoading ? <Loader2 className="size-4 animate-spin" /> : <RotateCcw />}
                                    {isDeleteLoading ? 'Resetting...' : 'Reset'}
                                </Button>

                                <Button
                                    disabled={isLoading}
                                    onClick={() => setIsDiffDialogOpen(true)}
                                    type="button"
                                    variant="outline"
                                >
                                    <FileDiff className="size-4" />
                                    Diff
                                </Button>
                            </>
                        )}
                        <Button
                            disabled={isLoading}
                            onClick={handleValidate}
                            type="button"
                            variant="outline"
                        >
                            {isValidateLoading ? (
                                <Loader2 className="size-4 animate-spin" />
                            ) : (
                                <CheckCircle className="size-4" />
                            )}
                            {isValidateLoading ? 'Validating...' : 'Validate'}
                        </Button>
                    </div>

                    <div className="ml-auto flex gap-2">
                        <Button
                            disabled={isLoading}
                            onClick={handleBack}
                            type="button"
                            variant="outline"
                        >
                            Cancel
                        </Button>
                        {activeTab === 'system' && (
                            <Button
                                disabled={isLoading}
                                form="system-prompt-form"
                                type="submit"
                                variant="secondary"
                            >
                                {isLoading ? <Loader2 className="size-4 animate-spin" /> : <Save className="size-4" />}
                                {isLoading ? 'Saving...' : 'Save Changes'}
                            </Button>
                        )}
                        {activeTab === 'human' && promptInfo?.type === 'agent' && promptInfo?.hasHuman && (
                            <Button
                                disabled={isLoading}
                                form="human-prompt-form"
                                type="submit"
                                variant="secondary"
                            >
                                {isLoading ? <Loader2 className="size-4 animate-spin" /> : <Save className="size-4" />}
                                {isLoading ? 'Saving...' : 'Save Changes'}
                            </Button>
                        )}
                    </div>
                </div>
            </div>

            {/* Reset Confirmation Dialog */}
            <ConfirmationDialog
                cancelText="Cancel"
                cancelVariant="outline"
                confirmIcon={<RotateCcw />}
                confirmText="Reset"
                confirmVariant="destructive"
                description="Are you sure you want to reset this prompt to its default value? This action cannot be undone."
                handleConfirm={handleConfirmReset}
                handleOpenChange={setResetDialogOpen}
                isOpen={resetDialogOpen}
                itemName={`${activeTab} prompt`}
                itemType="template"
                title="Reset Prompt"
            />

            {/* Leave Confirmation Dialog */}
            <ConfirmationDialog
                cancelText="Stay"
                confirmIcon={undefined}
                confirmText="Leave"
                confirmVariant="destructive"
                description="You have unsaved changes. Are you sure you want to leave without saving?"
                handleConfirm={handleConfirmLeave}
                handleOpenChange={handleLeaveDialogOpenChange}
                isOpen={isLeaveDialogOpen}
                title="Discard changes?"
            />

            {/* Validation Results Dialog */}
            <Dialog
                onOpenChange={setValidationDialogOpen}
                open={validationDialogOpen}
            >
                <DialogContent className="max-w-2xl">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <AlertCircle className="size-5" />
                            Validation Results
                        </DialogTitle>
                        <DialogDescription>
                            The validation result for the {activeTab} prompt template.
                        </DialogDescription>
                    </DialogHeader>

                    {validationResult && (
                        <div className="flex flex-col gap-4">
                            <Alert variant={validationResult.result ? 'default' : 'destructive'}>
                                {validationResult.result === 'success' ? (
                                    <CheckCircle className="size-4 text-green-500!" />
                                ) : (
                                    <XCircle className="size-4 text-red-500!" />
                                )}
                                <AlertTitle>
                                    {validationResult.result === 'success' ? 'Valid Template' : 'Validation Error'}
                                </AlertTitle>
                                <AlertDescription>
                                    <div className="whitespace-pre-line">
                                        {validationResult.message}
                                        {validationResult.details && (
                                            <div className="mt-2">
                                                <strong>Details:</strong> {validationResult.details}
                                            </div>
                                        )}
                                        {validationResult.line && (
                                            <div className="mt-1">
                                                <strong>Line:</strong> {validationResult.line}
                                            </div>
                                        )}
                                    </div>
                                </AlertDescription>
                            </Alert>

                            <div className="flex justify-end">
                                <Button onClick={() => setValidationDialogOpen(false)}>Close</Button>
                            </div>
                        </div>
                    )}
                </DialogContent>
            </Dialog>

            {/* Diff Dialog */}
            <Dialog
                onOpenChange={setIsDiffDialogOpen}
                open={isDiffDialogOpen}
            >
                <DialogContent className="max-w-7xl">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <FileDiff className="size-5" />
                            Diff
                        </DialogTitle>
                        <DialogDescription>Changes between current value and default template.</DialogDescription>
                    </DialogHeader>
                    <div className="max-h-[70vh] overflow-auto">
                        <ReactDiffViewer
                            newValue={currentTemplate}
                            oldValue={defaultTemplate}
                            splitView
                            styles={diffStyles}
                            useDarkTheme
                        />
                    </div>
                </DialogContent>
            </Dialog>
        </div>
    );
};

export default SettingsPrompt;
