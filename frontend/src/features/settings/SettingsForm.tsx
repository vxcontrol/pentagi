import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Switch } from '@/components/ui/switch';
import { Textarea } from '@/components/ui/textarea';

// Типы полей формы
export type SettingsFieldType = 'switch' | 'input' | 'select' | 'textarea' | 'number';

// Интерфейс для конфигурации поля
export interface SettingsField {
    id: string;
    label: string;
    description?: string;
    type: SettingsFieldType;
    defaultValue?: any;
    options?: { value: string; label: string }[]; // Для select
    placeholder?: string;
    required?: boolean;
    disabled?: boolean;
    min?: number; // Для number
    max?: number; // Для number
    rows?: number; // Для textarea
}

// Интерфейс для секции настроек
export interface SettingsSection {
    id: string;
    title: string;
    description?: string;
    fields: SettingsField[];
}

// Интерфейс для пропсов компонента
export interface SettingsFormProps {
    sections: SettingsSection[];
    onSave: (values: Record<string, any>) => Promise<void>;
    onCancel?: () => void;
    saveButtonText?: string;
    isLoading?: boolean;
}

const SettingsForm = ({
    sections,
    onSave,
    onCancel,
    saveButtonText = 'Save Changes',
    isLoading = false,
}: SettingsFormProps) => {
    // Создаем схему валидации динамически
    const createValidationSchema = () => {
        const schemaFields: Record<string, z.ZodTypeAny> = {};

        sections.forEach((section) => {
            section.fields.forEach((field) => {
                let fieldSchema: z.ZodTypeAny;

                switch (field.type) {
                    case 'switch':
                        fieldSchema = z.boolean();
                        break;
                    case 'number':
                        let numberSchema = z.number();
                        if (field.min !== undefined) numberSchema = numberSchema.min(field.min);
                        if (field.max !== undefined) numberSchema = numberSchema.max(field.max);
                        fieldSchema = numberSchema;
                        break;
                    case 'input':
                    case 'textarea':
                    case 'select':
                    default:
                        fieldSchema = z.string();
                        break;
                }

                if (!field.required) {
                    fieldSchema = fieldSchema.optional();
                }

                schemaFields[field.id] = fieldSchema;
            });
        });

        return z.object(schemaFields);
    };

    // Создаем значения по умолчанию
    const createDefaultValues = () => {
        const defaultValues: Record<string, any> = {};

        sections.forEach((section) => {
            section.fields.forEach((field) => {
                defaultValues[field.id] =
                    field.defaultValue ?? (field.type === 'switch' ? false : field.type === 'number' ? 0 : '');
            });
        });

        return defaultValues;
    };

    const form = useForm({
        resolver: zodResolver(createValidationSchema()),
        defaultValues: createDefaultValues(),
    });

    const handleSubmit = async (values: Record<string, any>) => {
        try {
            await onSave(values);
        } catch (error) {
            console.error('Error saving settings:', error);
        }
    };

    // Рендерим поле в зависимости от типа
    const renderField = (field: SettingsField) => {
        switch (field.type) {
            case 'switch':
                return (
                    <FormField
                        key={field.id}
                        control={form.control}
                        name={field.id}
                        render={({ field: formField }) => (
                            <FormItem>
                                <div className="flex items-center justify-between">
                                    <div className="space-y-0.5">
                                        <FormLabel className="text-base">{field.label}</FormLabel>
                                        {field.description && (
                                            <div className="text-sm text-muted-foreground">{field.description}</div>
                                        )}
                                    </div>
                                    <FormControl>
                                        <Switch
                                            checked={formField.value}
                                            onCheckedChange={formField.onChange}
                                            disabled={field.disabled || isLoading}
                                        />
                                    </FormControl>
                                </div>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                );

            case 'input':
                return (
                    <FormField
                        key={field.id}
                        control={form.control}
                        name={field.id}
                        render={({ field: formField }) => (
                            <FormItem>
                                <FormLabel>{field.label}</FormLabel>
                                {field.description && (
                                    <div className="text-sm text-muted-foreground">{field.description}</div>
                                )}
                                <FormControl>
                                    <Input
                                        placeholder={field.placeholder}
                                        disabled={field.disabled || isLoading}
                                        {...formField}
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                );

            case 'textarea':
                return (
                    <FormField
                        key={field.id}
                        control={form.control}
                        name={field.id}
                        render={({ field: formField }) => (
                            <FormItem>
                                <FormLabel>{field.label}</FormLabel>
                                {field.description && (
                                    <div className="text-sm text-muted-foreground">{field.description}</div>
                                )}
                                <FormControl>
                                    <Textarea
                                        placeholder={field.placeholder}
                                        disabled={field.disabled || isLoading}
                                        rows={field.rows || 3}
                                        {...formField}
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                );

            case 'select':
                return (
                    <FormField
                        key={field.id}
                        control={form.control}
                        name={field.id}
                        render={({ field: formField }) => (
                            <FormItem>
                                <FormLabel>{field.label}</FormLabel>
                                {field.description && (
                                    <div className="text-sm text-muted-foreground">{field.description}</div>
                                )}
                                <FormControl>
                                    <Select
                                        onValueChange={formField.onChange}
                                        value={formField.value}
                                        disabled={field.disabled || isLoading}
                                        placeholder={field.placeholder}
                                    >
                                        {field.options?.map((option) => (
                                            <option
                                                key={option.value}
                                                value={option.value}
                                            >
                                                {option.label}
                                            </option>
                                        ))}
                                    </Select>
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                );

            case 'number':
                return (
                    <FormField
                        key={field.id}
                        control={form.control}
                        name={field.id}
                        render={({ field: formField }) => (
                            <FormItem>
                                <FormLabel>{field.label}</FormLabel>
                                {field.description && (
                                    <div className="text-sm text-muted-foreground">{field.description}</div>
                                )}
                                <FormControl>
                                    <Input
                                        type="number"
                                        placeholder={field.placeholder}
                                        disabled={field.disabled || isLoading}
                                        min={field.min}
                                        max={field.max}
                                        {...formField}
                                        onChange={(e) => formField.onChange(Number(e.target.value))}
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                );

            default:
                return null;
        }
    };

    return (
        <div className="space-y-6">
            <Form {...form}>
                <form
                    onSubmit={form.handleSubmit(handleSubmit)}
                    className="space-y-6"
                >
                    {sections.map((section, sectionIndex) => (
                        <Card key={section.id}>
                            <CardHeader>
                                <CardTitle>{section.title}</CardTitle>
                                {section.description && <CardDescription>{section.description}</CardDescription>}
                            </CardHeader>
                            <CardContent className="space-y-4">
                                {section.fields.map((field, fieldIndex) => (
                                    <div key={field.id}>
                                        {renderField(field)}
                                        {fieldIndex < section.fields.length - 1 && <Separator />}
                                    </div>
                                ))}
                            </CardContent>
                        </Card>
                    ))}

                    {/* Кнопки действий */}
                    <div className="flex justify-end gap-2">
                        {onCancel && (
                            <Button
                                type="button"
                                variant="outline"
                                onClick={onCancel}
                                disabled={isLoading}
                            >
                                Cancel
                            </Button>
                        )}
                        <Button
                            type="submit"
                            disabled={isLoading}
                        >
                            {isLoading ? 'Saving...' : saveButtonText}
                        </Button>
                    </div>
                </form>
            </Form>
        </div>
    );
};

export default SettingsForm;
