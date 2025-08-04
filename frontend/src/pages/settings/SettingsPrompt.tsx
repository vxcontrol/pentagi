import ConfirmationDialog from '@/components/ConfirmationDialog';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Form, FormControl, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { StatusCard } from '@/components/ui/status-card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import {
    PromptType,
    useCreatePromptMutation,
    useDeletePromptMutation,
    useSettingsPromptsQuery,
    useUpdatePromptMutation,
    useValidatePromptMutation,
    type AgentPrompt,
    type AgentPrompts,
    type DefaultPrompt,
} from '@/graphql/types';
import { cn } from '@/lib/utils';
import { zodResolver } from '@hookform/resolvers/zod';
import { AlertCircle, Bot, CheckCircle, Code, Loader2, RotateCcw, Save, User, Wrench, XCircle } from 'lucide-react';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useController, useForm } from 'react-hook-form';
import { useNavigate, useParams } from 'react-router-dom';
import { z } from 'zod';

// Form schemas for each tab
const systemFormSchema = z.object({
    template: z.string().min(1, 'System template is required'),
});

const humanFormSchema = z.object({
    template: z.string().min(1, 'Human template is required'),
});

type SystemFormData = z.infer<typeof systemFormSchema>;
type HumanFormData = z.infer<typeof humanFormSchema>;

// Universal field components using useController
interface ControllerProps {
    name: string;
    control: any;
    disabled?: boolean;
}

interface BaseTextareaProps {
    placeholder?: string;
    className?: string;
}

interface BaseFieldProps extends ControllerProps {
    label?: string;
}

interface FormTextareaItemProps extends BaseFieldProps, BaseTextareaProps {
    description?: string;
}

const FormTextareaItem: React.FC<FormTextareaItemProps> = ({
    name,
    control,
    disabled,
    label,
    placeholder,
    className,
    description,
}) => {
    const { field, fieldState } = useController({
        name,
        control,
        defaultValue: '',
        disabled,
    });

    return (
        <FormItem>
            {label && <FormLabel>{label}</FormLabel>}
            <FormControl>
                <Textarea
                    {...field}
                    placeholder={placeholder}
                    className={cn('!min-h-[640px] font-mono text-sm', className)}
                    disabled={disabled}
                />
            </FormControl>
            {fieldState.error && <FormMessage>{fieldState.error.message}</FormMessage>}
        </FormItem>
    );
};

// Helper function to format display name
const formatName = (key: string): string => {
    return key.replace(/([A-Z])/g, ' $1').replace(/^./, (str) => str.toUpperCase());
};

// Helper function to extract used variables from template
const getUsedVariables = (template: string | undefined): Set<string> => {
    const usedVariables = new Set<string>();
    if (!template) return usedVariables;

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
    variables: string[];
    currentTemplate: string;
    onVariableClick: (variable: string) => void;
}

const Variables: React.FC<VariablesProps> = ({ variables, currentTemplate, onVariableClick }) => {
    if (variables.length === 0) return null;

    const usedVariables = getUsedVariables(currentTemplate);

    return (
        <div className="mb-4 p-3 bg-muted/50 rounded-md border">
            <h4 className="text-sm font-medium text-muted-foreground mb-2">Available Variables:</h4>
            <div className="flex flex-wrap gap-1">
                {variables.map((variable) => {
                    const isUsed = usedVariables.has(variable);
                    return (
                        <code
                            key={variable}
                            className={`px-2 py-1 border rounded text-xs font-mono cursor-pointer transition-colors ${
                                isUsed
                                    ? 'bg-green-100 border-green-300 text-green-800 hover:bg-green-200'
                                    : 'bg-background text-foreground hover:bg-accent'
                            }`}
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
    const { data, loading, error } = useSettingsPromptsQuery();
    const [createPrompt, { loading: isCreateLoading, error: createError }] = useCreatePromptMutation();
    const [updatePrompt, { loading: isUpdateLoading, error: updateError }] = useUpdatePromptMutation();
    const [deletePrompt, { loading: isDeleteLoading, error: deleteError }] = useDeletePromptMutation();
    const [validatePrompt, { loading: isValidateLoading, error: validateError }] = useValidatePromptMutation();

    // Local state management
    const [submitError, setSubmitError] = useState<string | null>(null);
    const [activeTab, setActiveTab] = useState<'system' | 'human'>('system');
    const [resetDialogOpen, setResetDialogOpen] = useState(false);
    const [validationResult, setValidationResult] = useState<any>(null);
    const [validationDialogOpen, setValidationDialogOpen] = useState(false);

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
                const textBeforeSelection = currentValue.substring(0, variableIndex);
                const linesBeforeSelection = textBeforeSelection.split('\n').length - 1;
                const selectionTop = linesBeforeSelection * lineHeight;
                const textareaHeight = textarea.clientHeight;
                const scrollTop = Math.max(0, selectionTop - textareaHeight / 2);

                textarea.scrollTop = scrollTop;
            } else {
                // Variable doesn't exist - insert it at cursor position (no scrolling)
                const start = textarea.selectionStart;
                const end = textarea.selectionEnd;
                const newValue = currentValue.substring(0, start) + variablePattern + currentValue.substring(end);
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
        if (!promptInfo) return;

        try {
            setSubmitError(null);

            if (activeTab === 'system' && promptInfo.userSystemPrompt) {
                await deletePrompt({
                    variables: { promptId: promptInfo.userSystemPrompt.id },
                    refetchQueries: ['settingsPrompts'],
                });
                // Reset form to default value
                systemForm.setValue('template', promptInfo.defaultSystemTemplate);
            } else if (activeTab === 'human' && promptInfo.userHumanPrompt) {
                await deletePrompt({
                    variables: { promptId: promptInfo.userHumanPrompt.id },
                    refetchQueries: ['settingsPrompts'],
                });
                // Reset form to default value
                humanForm.setValue('template', promptInfo.defaultHumanTemplate);
            }

            setResetDialogOpen(false);
            console.log('Prompt reset to default successfully');
        } catch (error) {
            console.error('Reset error:', error);
            setSubmitError(error instanceof Error ? error.message : 'An error occurred while resetting');
            setResetDialogOpen(false);
        }
    };

    // Handle validate prompt
    const handleValidate = async () => {
        if (!promptInfo) return;

        try {
            setSubmitError(null);
            setValidationResult(null);

            let promptType: PromptType;
            let currentTemplate: string;

            if (activeTab === 'system') {
                if (promptInfo.type === 'agent') {
                    const agentData = promptInfo.data as AgentPrompts | AgentPrompt;
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
                    type: promptType,
                    template: currentTemplate,
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
        resolver: zodResolver(systemFormSchema),
        defaultValues: {
            template: '',
        },
    });

    const humanForm = useForm<HumanFormData>({
        resolver: zodResolver(humanFormSchema),
        defaultValues: {
            template: '',
        },
    });

    // Watch form values to detect used variables
    const systemTemplate = systemForm.watch('template');
    const humanTemplate = humanForm.watch('template');

    // Determine prompt type and get prompt data
    const promptInfo = useMemo(() => {
        if (!promptId || !data?.settingsPrompts) {
            return null;
        }

        const { default: defaultPrompts, userDefined } = data.settingsPrompts;
        if (!defaultPrompts) return null;

        const { agents, tools } = defaultPrompts;

        // First check if there's a user-defined prompt
        const userPrompt = userDefined?.find((prompt) => {
            // For agents, check if this prompt matches system or human type
            const agentData = agents?.[promptId as keyof typeof agents] as AgentPrompts | AgentPrompt | undefined;
            if (agentData) {
                return (
                    prompt.type === agentData.system.type || prompt.type === (agentData as AgentPrompts)?.human?.type
                );
            }

            // For tools, check if this prompt matches tool type
            const toolData = tools?.[promptId as keyof typeof tools] as DefaultPrompt | undefined;
            if (toolData) {
                return prompt.type === toolData.type;
            }

            return false;
        });

        // Check if it's an agent prompt
        const agentData = agents?.[promptId as keyof typeof agents] as AgentPrompts | AgentPrompt | undefined;
        if (agentData) {
            // Check if we have user-defined system or human prompts
            const userSystemPrompt = userDefined?.find((p) => p.type === agentData.system.type);
            const userHumanPrompt = userDefined?.find((p) => p.type === (agentData as AgentPrompts)?.human?.type);

            return {
                type: 'agent' as const,
                displayName: formatName(promptId),
                data: agentData,
                hasHuman: !!(agentData as AgentPrompts)?.human,
                systemTemplate: userSystemPrompt?.template || agentData?.system?.template || '',
                humanTemplate: userHumanPrompt?.template || (agentData as AgentPrompts)?.human?.template || '',
                defaultSystemTemplate: agentData?.system?.template || '',
                defaultHumanTemplate: (agentData as AgentPrompts)?.human?.template || '',
                userSystemPrompt,
                userHumanPrompt,
            };
        }

        // Check if it's a tool prompt
        const toolData = tools?.[promptId as keyof typeof tools] as DefaultPrompt | undefined;
        if (toolData) {
            const userToolPrompt = userDefined?.find((p) => p.type === toolData.type);

            return {
                type: 'tool' as const,
                displayName: formatName(promptId),
                data: toolData,
                hasHuman: false,
                systemTemplate: userToolPrompt?.template || toolData?.template || '',
                humanTemplate: '',
                defaultSystemTemplate: toolData?.template || '',
                defaultHumanTemplate: '',
                userSystemPrompt: userToolPrompt,
                userHumanPrompt: null,
            };
        }

        return null;
    }, [promptId, data]);

    // Compute variables data based on active tab and prompt info
    const variablesData = useMemo(() => {
        if (!promptInfo) return null;

        let variables: string[] = [];
        let formId = '';
        let currentTemplate = '';

        if (activeTab === 'system') {
            variables =
                promptInfo.type === 'agent'
                    ? (promptInfo.data as AgentPrompts | AgentPrompt)?.system?.variables || []
                    : (promptInfo.data as DefaultPrompt)?.variables || [];
            formId = 'system-prompt-form';
            currentTemplate = systemTemplate;
        } else if (activeTab === 'human' && promptInfo.type === 'agent' && promptInfo.hasHuman) {
            variables = (promptInfo.data as AgentPrompts)?.human?.variables || [];
            formId = 'human-prompt-form';
            currentTemplate = humanTemplate;
        }

        return { variables, formId, currentTemplate };
    }, [promptInfo, activeTab, systemTemplate, humanTemplate]);

    // Handle variable click with useCallback for better performance
    const handleVariableClickCallback = useCallback(
        (variable: string) => {
            if (!variablesData) return;

            const field =
                activeTab === 'system'
                    ? {
                          value: systemTemplate,
                          onChange: (value: string) => systemForm.setValue('template', value),
                      }
                    : {
                          value: humanTemplate,
                          onChange: (value: string) => humanForm.setValue('template', value),
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
    }, [promptInfo, systemForm, humanForm]);

    // Form submission handlers
    const handleSystemSubmit = async (formData: SystemFormData) => {
        if (!promptInfo) return;

        const isUpdate = !!promptInfo.userSystemPrompt;

        // For creation, check if the template is identical to the default
        if (!isUpdate && formData.template === promptInfo.defaultSystemTemplate) {
            console.log('Template is identical to default, skipping save');
            return;
        }

        try {
            setSubmitError(null);

            // Get the real type from data
            let promptType: PromptType;

            if (promptInfo.type === 'agent') {
                const agentData = promptInfo.data as AgentPrompts | AgentPrompt;
                promptType = agentData.system.type;
            } else {
                const toolData = promptInfo.data as DefaultPrompt;
                promptType = toolData.type;
            }

            if (isUpdate) {
                // Update existing user-defined prompt
                await updatePrompt({
                    variables: {
                        promptId: promptInfo.userSystemPrompt!.id,
                        template: formData.template,
                    },
                    refetchQueries: ['settingsPrompts'],
                });
                console.log('System prompt updated successfully');
            } else {
                // Create new user-defined prompt
                await createPrompt({
                    variables: {
                        type: promptType,
                        template: formData.template,
                    },
                    refetchQueries: ['settingsPrompts'],
                });
                console.log('System prompt created successfully');
            }
        } catch (error) {
            console.error('Submit error:', error);
            setSubmitError(error instanceof Error ? error.message : 'An error occurred while saving');
        }
    };

    const handleHumanSubmit = async (formData: HumanFormData) => {
        if (!promptInfo) return;

        const isUpdate = !!promptInfo.userHumanPrompt;

        // For creation, check if the template is identical to the default
        if (!isUpdate && formData.template === promptInfo.defaultHumanTemplate) {
            console.log('Human template is identical to default, skipping save');
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
                    variables: {
                        promptId: promptInfo.userHumanPrompt!.id,
                        template: formData.template,
                    },
                    refetchQueries: ['settingsPrompts'],
                });
                console.log('Human prompt updated successfully');
            } else {
                // Create new user-defined prompt
                await createPrompt({
                    variables: {
                        type: humanPromptType,
                        template: formData.template,
                    },
                    refetchQueries: ['settingsPrompts'],
                });
                console.log('Human prompt created successfully');
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
                icon={<Loader2 className="w-16 h-16 animate-spin text-muted-foreground" />}
                title="Loading prompt data..."
                description="Please wait while we fetch prompt information"
            />
        );
    }

    // Error state
    if (error) {
        return (
            <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>Error loading prompt data</AlertTitle>
                <AlertDescription>{error.message}</AlertDescription>
            </Alert>
        );
    }

    // Prompt not found state
    if (!promptInfo) {
        return (
            <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>Prompt not found</AlertTitle>
                <AlertDescription>
                    The prompt "{promptId}" could not be found or is not supported for editing.
                </AlertDescription>
            </Alert>
        );
    }

    const mutationError = createError || updateError || deleteError || validateError || submitError;

    return (
        <>
            <Card>
                <CardHeader>
                    <div className="flex items-center gap-2">
                        {promptInfo.type === 'agent' ? (
                            <Bot className="h-5 w-5 text-muted-foreground" />
                        ) : (
                            <Wrench className="h-5 w-5 text-muted-foreground" />
                        )}
                        <CardTitle>{promptInfo.displayName}</CardTitle>
                    </div>
                    <CardDescription>
                        {promptInfo.type === 'agent'
                            ? 'Configure prompts for this AI agent'
                            : 'Configure the prompt for this tool'}
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <Tabs
                        defaultValue="system"
                        className="w-full"
                        onValueChange={(value) => setActiveTab(value as 'system' | 'human')}
                    >
                        <TabsList>
                            <TabsTrigger value="system">
                                <div className="flex items-center gap-2">
                                    <Code className="h-4 w-4" />
                                    System Prompt
                                </div>
                            </TabsTrigger>
                            {promptInfo.type === 'agent' && promptInfo.hasHuman && (
                                <TabsTrigger value="human">
                                    <div className="flex items-center gap-2">
                                        <User className="h-4 w-4" />
                                        Human Prompt
                                    </div>
                                </TabsTrigger>
                            )}
                        </TabsList>

                        <TabsContent
                            value="system"
                            className="mt-6"
                        >
                            <Form {...systemForm}>
                                <form
                                    id="system-prompt-form"
                                    onSubmit={systemForm.handleSubmit(handleSystemSubmit)}
                                    className="space-y-6"
                                >
                                    {/* Error Alert */}
                                    {mutationError && (
                                        <Alert variant="destructive">
                                            <AlertCircle className="h-4 w-4" />
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
                                        name="template"
                                        control={systemForm.control}
                                        disabled={isLoading}
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
                                value="human"
                                className="mt-6"
                            >
                                <Form {...humanForm}>
                                    <form
                                        id="human-prompt-form"
                                        onSubmit={humanForm.handleSubmit(handleHumanSubmit)}
                                        className="space-y-6"
                                    >
                                        {/* Error Alert */}
                                        {mutationError && (
                                            <Alert variant="destructive">
                                                <AlertCircle className="h-4 w-4" />
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
                                            name="template"
                                            control={humanForm.control}
                                            disabled={isLoading}
                                            placeholder="Enter the human prompt template..."
                                        />
                                    </form>
                                </Form>
                            </TabsContent>
                        )}
                    </Tabs>
                </CardContent>
            </Card>

            {/* Sticky footer with variables and buttons */}
            <div className="sticky -bottom-4 bg-background border-t mt-4 -mx-4 -mb-4 p-4 shadow-lg">
                {/* Variables */}
                {variablesData && (
                    <Variables
                        variables={variablesData.variables}
                        currentTemplate={variablesData.currentTemplate}
                        onVariableClick={handleVariableClickCallback}
                    />
                )}

                {/* Action buttons */}
                <div className="flex items-center">
                    <div className="flex space-x-2">
                        {/* Reset button - only show when user has custom prompt */}
                        {((activeTab === 'system' && promptInfo?.userSystemPrompt) ||
                            (activeTab === 'human' && promptInfo?.userHumanPrompt)) && (
                            <Button
                                type="button"
                                variant="destructive"
                                onClick={handleReset}
                                disabled={isLoading}
                            >
                                {isDeleteLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <RotateCcw />}
                                {isDeleteLoading ? 'Resetting...' : 'Reset'}
                            </Button>
                        )}
                        <Button
                            type="button"
                            variant="outline"
                            onClick={handleValidate}
                            disabled={isLoading}
                        >
                            {isValidateLoading ? (
                                <Loader2 className="h-4 w-4 animate-spin" />
                            ) : (
                                <CheckCircle className="h-4 w-4" />
                            )}
                            {isValidateLoading ? 'Validating...' : 'Validate'}
                        </Button>
                    </div>

                    <div className="flex space-x-2 ml-auto">
                        <Button
                            type="button"
                            variant="outline"
                            onClick={() => navigate('/settings/prompts')}
                            disabled={isLoading}
                        >
                            Cancel
                        </Button>
                        {activeTab === 'system' && (
                            <Button
                                form="system-prompt-form"
                                variant="secondary"
                                type="submit"
                                disabled={isLoading}
                            >
                                {isLoading ? (
                                    <Loader2 className="h-4 w-4 animate-spin" />
                                ) : (
                                    <Save className="h-4 w-4" />
                                )}
                                {isLoading ? 'Saving...' : 'Save Changes'}
                            </Button>
                        )}
                        {activeTab === 'human' && promptInfo?.type === 'agent' && promptInfo?.hasHuman && (
                            <Button
                                form="human-prompt-form"
                                variant="secondary"
                                type="submit"
                                disabled={isLoading}
                            >
                                {isLoading ? (
                                    <Loader2 className="h-4 w-4 animate-spin" />
                                ) : (
                                    <Save className="h-4 w-4" />
                                )}
                                {isLoading ? 'Saving...' : 'Save Changes'}
                            </Button>
                        )}
                    </div>
                </div>
            </div>

            {/* Reset Confirmation Dialog */}
            <ConfirmationDialog
                isOpen={resetDialogOpen}
                handleOpenChange={setResetDialogOpen}
                handleConfirm={handleConfirmReset}
                title="Reset Prompt"
                description="Are you sure you want to reset this prompt to its default value? This action cannot be undone."
                itemName={`${activeTab} prompt`}
                itemType="template"
                confirmText="Reset"
                cancelText="Cancel"
                confirmVariant="destructive"
                cancelVariant="outline"
                confirmIcon={<RotateCcw />}
            />

            {/* Validation Results Dialog */}
            <Dialog
                open={validationDialogOpen}
                onOpenChange={setValidationDialogOpen}
            >
                <DialogContent className="max-w-2xl">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <AlertCircle className="h-5 w-5" />
                            Validation Results
                        </DialogTitle>
                        <DialogDescription>
                            The validation result for the {activeTab} prompt template.
                        </DialogDescription>
                    </DialogHeader>

                    {validationResult && (
                        <div className="space-y-4">
                            <Alert variant={validationResult.result ? 'default' : 'destructive'}>
                                {validationResult.result === 'success' ? (
                                    <CheckCircle className="h-4 w-4 !text-green-500" />
                                ) : (
                                    <XCircle className="h-4 w-4 !text-red-500" />
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
        </>
    );
};

export default SettingsPrompt;
