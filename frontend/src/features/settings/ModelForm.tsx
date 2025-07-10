import { zodResolver } from '@hookform/resolvers/zod';
import { Plus, Trash2 } from 'lucide-react';
import { useFieldArray, useForm } from 'react-hook-form';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';

// Интерфейсы
interface ModelOption {
    value: string;
    label: string;
}

interface CustomField {
    key: string;
    value: string;
}

interface ModelFormProps {
    providerId: string;
    providerName: string;
    onSave: (values: any) => Promise<void>;
    isLoading?: boolean;
}

// Тестовые данные для моделей - позже заменим на API
const getModelsForProvider = (providerId: string): ModelOption[] => {
    const modelsMap: Record<string, ModelOption[]> = {
        openai: [
            { value: 'gpt-4o', label: 'GPT-4o' },
            { value: 'gpt-4o-mini', label: 'GPT-4o Mini' },
            { value: 'gpt-4-turbo', label: 'GPT-4 Turbo' },
            { value: 'gpt-3.5-turbo', label: 'GPT-3.5 Turbo' },
        ],
        anthropic: [
            { value: 'claude-3-5-sonnet-20241022', label: 'Claude 3.5 Sonnet' },
            { value: 'claude-3-haiku-20240307', label: 'Claude 3 Haiku' },
            { value: 'claude-3-opus-20240229', label: 'Claude 3 Opus' },
        ],
        deepseek: [
            { value: 'deepseek-chat', label: 'DeepSeek Chat' },
            { value: 'deepseek-coder', label: 'DeepSeek Coder' },
        ],
        openrouter: [
            { value: 'meta-llama/llama-3.1-405b-instruct', label: 'Llama 3.1 405B' },
            { value: 'anthropic/claude-3.5-sonnet', label: 'Claude 3.5 Sonnet' },
            { value: 'openai/gpt-4o', label: 'GPT-4o' },
        ],
    };

    return modelsMap[providerId] || [];
};

// Схема валидации
const modelFormSchema = z.object({
    model: z.string().min(1, 'Model is required'),
    temperature: z.number().min(0).max(2),
    topP: z.number().min(0).max(1),
    customFields: z.array(
        z.object({
            key: z.string().min(1, 'Key is required'),
            value: z.string().min(1, 'Value is required'),
        }),
    ),
});

type ModelFormValues = z.infer<typeof modelFormSchema>;

const ModelForm = ({ providerId, providerName, onSave, isLoading = false }: ModelFormProps) => {
    const availableModels = getModelsForProvider(providerId);

    const form = useForm<ModelFormValues>({
        resolver: zodResolver(modelFormSchema),
        defaultValues: {
            model: '',
            temperature: 0.7,
            topP: 1.0,
            customFields: [],
        },
    });

    const { fields, append, remove } = useFieldArray({
        control: form.control,
        name: 'customFields',
    });

    const handleSubmit = async (values: ModelFormValues) => {
        try {
            await onSave(values);
        } catch (error) {
            console.error('Error saving model settings:', error);
        }
    };

    const addCustomField = () => {
        append({ key: '', value: '' });
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
                                {providerName} Model Configuration
                            </CardTitle>
                            <CardDescription>Configure the AI model and parameters for {providerName}</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            {/* Model Selection */}
                            <FormField
                                control={form.control}
                                name="model"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Model</FormLabel>
                                        <FormControl>
                                            <Select
                                                onValueChange={field.onChange}
                                                value={field.value}
                                                disabled={isLoading}
                                                placeholder="Select a model"
                                            >
                                                {availableModels.map((model) => (
                                                    <option
                                                        key={model.value}
                                                        value={model.value}
                                                    >
                                                        {model.label}
                                                    </option>
                                                ))}
                                            </Select>
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            {/* Temperature */}
                            <FormField
                                control={form.control}
                                name="temperature"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Temperature</FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                step="0.1"
                                                min="0"
                                                max="2"
                                                placeholder="0.7"
                                                disabled={isLoading}
                                                {...field}
                                                onChange={(e) => field.onChange(Number(e.target.value))}
                                            />
                                        </FormControl>
                                        <div className="text-sm text-muted-foreground">
                                            Controls randomness (0 = deterministic, 2 = very random)
                                        </div>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            {/* Top P */}
                            <FormField
                                control={form.control}
                                name="topP"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Top P</FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                step="0.1"
                                                min="0"
                                                max="1"
                                                placeholder="1.0"
                                                disabled={isLoading}
                                                {...field}
                                                onChange={(e) => field.onChange(Number(e.target.value))}
                                            />
                                        </FormControl>
                                        <div className="text-sm text-muted-foreground">
                                            Controls diversity via nucleus sampling (0.1 = narrow, 1.0 = wide)
                                        </div>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </CardContent>
                    </Card>

                    {/* Custom Fields */}
                    <Card>
                        <CardHeader>
                            <CardTitle>Custom Parameters</CardTitle>
                            <CardDescription>Add additional parameters specific to this provider</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            {fields.map((field, index) => (
                                <div
                                    key={field.id}
                                    className="flex gap-2 items-end"
                                >
                                    <FormField
                                        control={form.control}
                                        name={`customFields.${index}.key`}
                                        render={({ field }) => (
                                            <FormItem className="flex-1">
                                                <FormLabel>Parameter Name</FormLabel>
                                                <FormControl>
                                                    <Input
                                                        placeholder="max_tokens"
                                                        disabled={isLoading}
                                                        {...field}
                                                    />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                    <FormField
                                        control={form.control}
                                        name={`customFields.${index}.value`}
                                        render={({ field }) => (
                                            <FormItem className="flex-1">
                                                <FormLabel>Value</FormLabel>
                                                <FormControl>
                                                    <Input
                                                        placeholder="4096"
                                                        disabled={isLoading}
                                                        {...field}
                                                    />
                                                </FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )}
                                    />
                                    <Button
                                        type="button"
                                        variant="outline"
                                        size="icon"
                                        onClick={() => remove(index)}
                                        disabled={isLoading}
                                    >
                                        <Trash2 className="h-4 w-4" />
                                    </Button>
                                </div>
                            ))}

                            <Button
                                type="button"
                                variant="outline"
                                onClick={addCustomField}
                                disabled={isLoading}
                                className="w-full"
                            >
                                <Plus className="h-4 w-4 mr-2" />
                                Add Custom Parameter
                            </Button>
                        </CardContent>
                    </Card>

                    {/* Save Button */}
                    <div className="flex justify-end">
                        <Button
                            type="submit"
                            disabled={isLoading}
                        >
                            {isLoading ? 'Saving...' : 'Save Configuration'}
                        </Button>
                    </div>
                </form>
            </Form>
        </div>
    );
};

export default ModelForm;
