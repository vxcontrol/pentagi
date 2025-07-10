import { zodResolver } from '@hookform/resolvers/zod';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Textarea } from '@/components/ui/textarea';

// Интерфейсы
interface PromptFormProps {
    itemId: string;
    itemName: string;
    itemType: 'agent' | 'tool';
    onSave: (values: any) => Promise<void>;
    isLoading?: boolean;
}

interface Variable {
    name: string;
    description: string;
    example?: string;
}

// Тестовые данные переменных - позже заменим на API
const getVariablesForType = (type: 'agent' | 'tool', itemId: string): Variable[] => {
    const commonVariables: Variable[] = [
        { name: '{user_input}', description: "User's input text", example: 'Write a function to calculate fibonacci' },
        { name: '{context}', description: 'Current conversation context', example: 'Previous messages and context' },
        { name: '{datetime}', description: 'Current date and time', example: '2024-01-15 14:30:00' },
    ];

    const agentVariables: Variable[] = [
        ...commonVariables,
        { name: '{flow_id}', description: 'Current flow identifier', example: 'flow_123' },
        {
            name: '{task_description}',
            description: 'Description of the current task',
            example: 'Perform security analysis',
        },
        { name: '{assistant_name}', description: 'Name of the current assistant', example: 'Pentest Assistant' },
    ];

    const toolVariables: Variable[] = [
        ...commonVariables,
        { name: '{command}', description: 'Command to execute', example: 'ls -la' },
        { name: '{file_path}', description: 'Target file path', example: '/home/user/document.txt' },
        { name: '{search_query}', description: 'Search query text', example: 'vulnerability scan' },
    ];

    // Специфичные переменные для разных тулзов
    const toolSpecificVariables: Record<string, Variable[]> = {
        browser: [
            { name: '{url}', description: 'Target URL', example: 'https://example.com' },
            { name: '{action}', description: 'Browser action', example: 'navigate, click, scroll' },
        ],
        terminal: [
            { name: '{shell}', description: 'Shell type', example: 'bash, zsh, fish' },
            { name: '{working_dir}', description: 'Current working directory', example: '/home/user/project' },
        ],
        search: [
            { name: '{search_type}', description: 'Type of search', example: 'web, file, code' },
            { name: '{filters}', description: 'Search filters', example: 'filetype:pdf, site:github.com' },
        ],
        file: [
            { name: '{operation}', description: 'File operation', example: 'read, write, create, delete' },
            { name: '{encoding}', description: 'File encoding', example: 'utf-8, ascii' },
        ],
    };

    if (type === 'agent') {
        return agentVariables;
    } else {
        const baseToolVars = toolVariables;
        const specificVars = toolSpecificVariables[itemId] || [];
        return [...baseToolVars, ...specificVars];
    }
};

// Схема валидации
const promptFormSchema = z.object({
    systemPrompt: z.string().min(1, 'System prompt is required'),
    userPrompt: z.string().min(1, 'User prompt is required'),
});

type PromptFormValues = z.infer<typeof promptFormSchema>;

// Компонент для отображения переменных
const VariablesBadges = ({
    variables,
    text,
    onVariableClick,
}: {
    variables: Variable[];
    text: string;
    onVariableClick: (variable: string) => void;
}) => {
    const usedVariables = variables.filter((variable) => text.includes(variable.name));

    const availableVariables = variables.filter((variable) => !text.includes(variable.name));

    return (
        <div className="space-y-2">
            {usedVariables.length > 0 && (
                <div>
                    <div className="text-xs font-medium text-green-600 dark:text-green-400 mb-1">Used Variables:</div>
                    <div className="flex flex-wrap gap-1">
                        {usedVariables.map((variable) => (
                            <Badge
                                key={variable.name}
                                variant="default"
                                className="cursor-pointer bg-green-100 text-green-800 hover:bg-green-200 dark:bg-green-900 dark:text-green-100"
                                onClick={() => onVariableClick(variable.name)}
                                title={variable.description}
                            >
                                {variable.name}
                            </Badge>
                        ))}
                    </div>
                </div>
            )}

            {availableVariables.length > 0 && (
                <div>
                    <div className="text-xs font-medium text-muted-foreground mb-1">Available Variables:</div>
                    <div className="flex flex-wrap gap-1">
                        {availableVariables.map((variable) => (
                            <Badge
                                key={variable.name}
                                variant="outline"
                                className="cursor-pointer hover:bg-accent"
                                onClick={() => onVariableClick(variable.name)}
                                title={`${variable.description}${variable.example ? `\nExample: ${variable.example}` : ''}`}
                            >
                                {variable.name}
                            </Badge>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};

const PromptForm = ({ itemId, itemName, itemType, onSave, isLoading = false }: PromptFormProps) => {
    const [variables, setVariables] = useState<Variable[]>([]);

    const form = useForm<PromptFormValues>({
        resolver: zodResolver(promptFormSchema),
        defaultValues: {
            systemPrompt: '',
            userPrompt: '',
        },
    });

    const systemPromptValue = form.watch('systemPrompt');
    const userPromptValue = form.watch('userPrompt');

    useEffect(() => {
        // TODO: Заменить на API вызов
        setVariables(getVariablesForType(itemType, itemId));
    }, [itemType, itemId]);

    const handleSubmit = async (values: PromptFormValues) => {
        try {
            await onSave(values);
        } catch (error) {
            console.error('Error saving prompt settings:', error);
        }
    };

    const insertVariable = (variable: string, fieldName: 'systemPrompt' | 'userPrompt') => {
        const currentValue = form.getValues(fieldName);
        const newValue = currentValue + (currentValue.endsWith(' ') ? '' : ' ') + variable + ' ';
        form.setValue(fieldName, newValue);
    };

    return (
        <div className="space-y-6">
            <Form {...form}>
                <form
                    onSubmit={form.handleSubmit(handleSubmit)}
                    className="space-y-6"
                >
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                {itemName} {itemType === 'agent' ? 'Agent' : 'Tool'} Configuration
                            </CardTitle>
                            <CardDescription>
                                Configure the prompts for {itemName} {itemType}
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-6">
                            {/* System Prompt */}
                            <FormField
                                control={form.control}
                                name="systemPrompt"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>System Prompt</FormLabel>
                                        <FormControl>
                                            <Textarea
                                                placeholder={`Enter the system prompt for ${itemName}...`}
                                                disabled={isLoading}
                                                {...field}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                        <VariablesBadges
                                            variables={variables}
                                            text={systemPromptValue}
                                            onVariableClick={(variable) => insertVariable(variable, 'systemPrompt')}
                                        />
                                    </FormItem>
                                )}
                            />

                            {/* User Prompt */}
                            <FormField
                                control={form.control}
                                name="userPrompt"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>User Prompt</FormLabel>
                                        <FormControl>
                                            <Textarea
                                                placeholder={`Enter the user prompt template for ${itemName}...`}
                                                disabled={isLoading}
                                                {...field}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                        <VariablesBadges
                                            variables={variables}
                                            text={userPromptValue}
                                            onVariableClick={(variable) => insertVariable(variable, 'userPrompt')}
                                        />
                                    </FormItem>
                                )}
                            />
                        </CardContent>
                    </Card>

                    {/* Variables Reference */}
                    <Card>
                        <CardHeader>
                            <CardTitle>Variable Reference</CardTitle>
                            <CardDescription>Available variables and their descriptions</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="space-y-3">
                                {variables.map((variable) => (
                                    <div
                                        key={variable.name}
                                        className="border-l-2 border-muted pl-3"
                                    >
                                        <div className="flex items-center gap-2">
                                            <Badge
                                                variant="outline"
                                                className="font-mono"
                                            >
                                                {variable.name}
                                            </Badge>
                                            <span className="text-sm text-muted-foreground">
                                                {variable.description}
                                            </span>
                                        </div>
                                        {variable.example && (
                                            <div className="mt-1 text-xs text-muted-foreground font-mono bg-muted px-2 py-1 rounded">
                                                Example: {variable.example}
                                            </div>
                                        )}
                                    </div>
                                ))}
                            </div>
                        </CardContent>
                    </Card>

                    {/* Save Button */}
                    <div className="flex justify-end">
                        <Button
                            type="submit"
                            disabled={isLoading}
                        >
                            {isLoading ? 'Saving...' : 'Save Prompts'}
                        </Button>
                    </div>
                </form>
            </Form>
        </div>
    );
};

export default PromptForm;
