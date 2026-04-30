import { zodResolver } from '@hookform/resolvers/zod';
import {
    ArrowUp,
    Check,
    ChevronDown,
    FileSymlink,
    FileText,
    Folder,
    Loader2,
    Paperclip,
    Plus,
    Square,
    X,
} from 'lucide-react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useRef } from 'react';
import { useForm, useWatch } from 'react-hook-form';
import { z } from 'zod';

import type { UserResourceFragmentFragment } from '@/graphql/types';

import { ProviderIcon } from '@/components/icons/provider-icon';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuGroup,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Form, FormControl, FormField, FormLabel } from '@/components/ui/form';
import {
    InputGroup,
    InputGroupAddon,
    InputGroupButton,
    InputGroupInput,
    InputGroupTextareaAutosize,
} from '@/components/ui/input-group';
import { Spinner } from '@/components/ui/spinner';
import { Switch } from '@/components/ui/switch';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { useResourcesUpload } from '@/features/resources/use-resources-upload';
import { getProviderDisplayName } from '@/models/provider';
import { useProviders } from '@/providers/providers-provider';
import { useResources } from '@/providers/resources-provider';
import { type Template, useTemplates } from '@/providers/templates-provider';

const formSchema = z.object({
    message: z.string().trim().min(1, { message: 'Message cannot be empty' }),
    providerName: z.string().trim().min(1, { message: 'Provider must be selected' }),
    resourceIds: z.array(z.string()),
    useAgents: z.boolean(),
});

export interface FlowFormProps {
    defaultValues?: Partial<FlowFormValues>;
    isCanceling?: boolean;
    isDisabled?: boolean;
    isLoading?: boolean;
    isProviderDisabled?: boolean;
    isSubmitting?: boolean;
    onCancel?: () => Promise<void> | void;
    onSubmit: (values: FlowFormValues) => Promise<void> | void;
    placeholder?: string;
    type: 'assistant' | 'automation';
}

export type FlowFormValues = z.infer<typeof formSchema>;

export const FlowForm = ({
    defaultValues,
    isCanceling,
    isDisabled,
    isLoading,
    isProviderDisabled,
    isSubmitting,
    onCancel,
    onSubmit,
    placeholder = 'Describe what you would like PentAGI to test...',
    type,
}: FlowFormProps) => {
    const { providers, setSelectedProvider } = useProviders();
    const { templates } = useTemplates();
    const { resources } = useResources();
    const [isReplaceConfirmOpen, setIsReplaceConfirmOpen] = useState(false);
    const [pendingTemplate, setPendingTemplate] = useState<null | Template>(null);
    const [providerSearch, setProviderSearch] = useState('');
    const [templateSearch, setTemplateSearch] = useState('');
    const [resourceSearch, setResourceSearch] = useState('');

    const fileInputRef = useRef<HTMLInputElement>(null);

    const sortedResources = useMemo(
        () => [...resources].sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()),
        [resources],
    );

    const filteredResources = useMemo(() => {
        const queryValue = resourceSearch.trim().toLowerCase();

        if (!queryValue) {
            return sortedResources;
        }

        return sortedResources.filter(
            (resource) =>
                resource.name.toLowerCase().includes(queryValue) || resource.path.toLowerCase().includes(queryValue),
        );
    }, [sortedResources, resourceSearch]);

    const filteredTemplates = useMemo(() => {
        if (!templateSearch.trim()) {
            return templates;
        }

        const searchLower = templateSearch.toLowerCase();

        return templates.filter(
            (template) =>
                template.title.toLowerCase().includes(searchLower) || template.text.toLowerCase().includes(searchLower),
        );
    }, [templates, templateSearch]);

    const filteredProviders = useMemo(() => {
        if (!providerSearch.trim()) {
            return providers;
        }

        const searchLower = providerSearch.toLowerCase();

        return providers.filter((provider) => {
            const displayName = getProviderDisplayName(provider).toLowerCase();

            return displayName.includes(searchLower) || provider.name.toLowerCase().includes(searchLower);
        });
    }, [providers, providerSearch]);

    const form = useForm<FlowFormValues>({
        defaultValues: {
            message: defaultValues?.message ?? '',
            providerName: defaultValues?.providerName ?? '',
            resourceIds: defaultValues?.resourceIds ?? [],
            useAgents: defaultValues?.useAgents ?? false,
        },
        mode: 'onChange',
        resolver: zodResolver(formSchema),
    });

    const {
        control,
        formState: { dirtyFields, isValid },
        getValues,
        handleSubmit: handleFormSubmit,
        resetField,
        setValue,
    } = form;

    const resourceIds = useWatch({ control, name: 'resourceIds' });

    const updateResourceIds = useCallback(
        (updater: ((current: string[]) => string[]) | string[]) => {
            const current = getValues('resourceIds') ?? [];
            const raw = typeof updater === 'function' ? updater(current) : updater;
            // TODO(backend): drop the `String(id)` coercion once the REST `/resources/`
            // endpoints return `id` as a string (matching the GraphQL `ID` scalar). Until
            // then this is a safety net for stray numeric ids that would fail zod's
            // `z.array(z.string())` and silently block submission.
            const next = raw.map((id) => String(id));
            setValue('resourceIds', next, { shouldDirty: true, shouldValidate: true });
        },
        [getValues, setValue],
    );

    const flowResources = useMemo<UserResourceFragmentFragment[]>(() => {
        const byId = new Map(resources.map((item) => [item.id, item]));

        return resourceIds
            .map((id) => byId.get(id))
            .filter((item): item is UserResourceFragmentFragment => Boolean(item));
    }, [resourceIds, resources]);

    const upload = useResourcesUpload({
        onSuccess: (uploaded) => {
            const ids = uploaded.items.map((item) => item.id);

            if (ids.length === 0) {
                return;
            }

            updateResourceIds((current) => {
                const merged = new Set(current);
                ids.forEach((id) => merged.add(id));

                return Array.from(merged);
            });
        },
    });

    // Update form values from defaultValues if user hasn't manually changed them
    useEffect(() => {
        if (!defaultValues) {
            return;
        }

        const currentValues = getValues();

        // Update only fields that user hasn't manually changed and that differ from current values.
        // Arrays are compared shallowly so a new-but-identical `resourceIds` reference doesn't
        // trigger an unnecessary setValue (and the re-render it causes).
        Object.entries(defaultValues)
            .filter(([fieldName, defaultValue]) => {
                const typedFieldName = fieldName as keyof FlowFormValues;

                if (defaultValue === undefined || dirtyFields[typedFieldName]) {
                    return false;
                }

                const currentValue = currentValues[typedFieldName];

                if (Array.isArray(defaultValue) && Array.isArray(currentValue)) {
                    return (
                        currentValue.length !== defaultValue.length ||
                        currentValue.some((item, index) => item !== defaultValue[index])
                    );
                }

                return currentValue !== defaultValue;
            })
            .forEach(([fieldName, defaultValue]) => {
                const typedFieldName = fieldName as keyof FlowFormValues;
                setValue(typedFieldName, defaultValue as never, { shouldDirty: false });
            });
    }, [defaultValues, dirtyFields, setValue, getValues]);

    const isFormDisabled = isDisabled || isLoading || isSubmitting || isCanceling;

    const textareaRef = useRef<HTMLTextAreaElement>(null);
    const previousFormDisabledRef = useRef(isFormDisabled);

    useEffect(() => {
        const wasDisabled = previousFormDisabledRef.current;
        previousFormDisabledRef.current = isFormDisabled;

        if (wasDisabled && !isFormDisabled) {
            textareaRef.current?.focus();
        }
    }, [isFormDisabled]);

    const handleSubmit = async (values: FlowFormValues) => {
        await onSubmit(values);
        resetField('message');
        resetField('resourceIds');
    };

    const handleAttachClick = () => {
        fileInputRef.current?.click();
    };

    const handleFileInputChange = useCallback(
        async (event: React.ChangeEvent<HTMLInputElement>) => {
            const selectedFiles = Array.from(event.target.files ?? []);

            // Reset native input so re-uploading the same file fires `change` again.
            event.target.value = '';

            if (selectedFiles.length === 0) {
                return;
            }

            await upload.uploadFiles(selectedFiles);
        },
        [upload],
    );

    const handleRemoveAttachment = (id: string) => {
        updateResourceIds((current) => current.filter((fileId) => fileId !== id));
    };

    const handleToggleAttachment = (id: string) => {
        updateResourceIds((current) =>
            current.includes(id) ? current.filter((fileId) => fileId !== id) : [...current, id],
        );
    };

    const handleKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
        const { ctrlKey, key, metaKey, shiftKey } = event;

        if (isFormDisabled || key !== 'Enter' || shiftKey || ctrlKey || metaKey) {
            return;
        }

        event.preventDefault();
        handleFormSubmit(handleSubmit)();
    };

    const handleApplyTemplate = useCallback(
        (template: Template) => {
            const currentMessage = getValues('message')?.trim() ?? '';

            if (currentMessage.length > 0) {
                setPendingTemplate(template);
                setIsReplaceConfirmOpen(true);
            } else {
                setValue('message', template.text, { shouldValidate: true });
                setTemplateSearch('');
            }
        },
        [getValues, setValue],
    );

    const handleConfirmReplaceTemplate = useCallback(() => {
        if (pendingTemplate) {
            setValue('message', pendingTemplate.text, { shouldValidate: true });
            setTemplateSearch('');
            setPendingTemplate(null);
        }
    }, [pendingTemplate, setValue]);

    return (
        <Form {...form}>
            <form onSubmit={handleFormSubmit(handleSubmit)}>
                <FormField
                    control={control}
                    name="message"
                    render={({ field }) => (
                        <FormControl>
                            <InputGroup className="block">
                                {flowResources.length > 0 && (
                                    <InputGroupAddon
                                        align="block-start"
                                        className="flex-wrap gap-1.5"
                                    >
                                        {flowResources.map((resource) => {
                                            const Icon = resource.isDir ? Folder : FileText;

                                            return (
                                                <div
                                                    className="bg-muted/50 flex max-w-full items-center gap-1.5 rounded-md border px-2 py-1 text-xs"
                                                    key={resource.id}
                                                    title={resource.path}
                                                >
                                                    <Icon className="text-muted-foreground size-3.5 shrink-0" />
                                                    <span className="text-foreground max-w-40 truncate">
                                                        {resource.name}
                                                    </span>
                                                    <button
                                                        aria-label={`Remove ${resource.name}`}
                                                        className="text-muted-foreground hover:text-destructive ml-0.5 flex shrink-0 items-center justify-center"
                                                        disabled={isFormDisabled}
                                                        onClick={() => handleRemoveAttachment(resource.id)}
                                                        type="button"
                                                    >
                                                        <X className="size-3.5" />
                                                    </button>
                                                </div>
                                            );
                                        })}
                                    </InputGroupAddon>
                                )}
                                <InputGroupTextareaAutosize
                                    {...field}
                                    autoFocus
                                    className="min-h-0"
                                    disabled={isFormDisabled}
                                    maxRows={9}
                                    minRows={1}
                                    onKeyDown={handleKeyDown}
                                    placeholder={placeholder}
                                    ref={(element) => {
                                        field.ref(element);
                                        textareaRef.current = element;
                                    }}
                                />
                                <InputGroupAddon align="block-end">
                                    <FormField
                                        control={control}
                                        name="providerName"
                                        render={({ field: providerField }) => {
                                            const currentProvider = providers.find(
                                                (p) => p.name === providerField.value,
                                            );

                                            return (
                                                <DropdownMenu>
                                                    <DropdownMenuTrigger asChild>
                                                        <InputGroupButton
                                                            disabled={isFormDisabled || isProviderDisabled}
                                                            variant="ghost"
                                                        >
                                                            {currentProvider && (
                                                                <ProviderIcon provider={currentProvider} />
                                                            )}
                                                            <span className="max-w-40 truncate">
                                                                {currentProvider
                                                                    ? getProviderDisplayName(currentProvider)
                                                                    : 'Select Provider'}
                                                            </span>
                                                            <ChevronDown />
                                                        </InputGroupButton>
                                                    </DropdownMenuTrigger>
                                                    <DropdownMenuContent
                                                        align="start"
                                                        side="top"
                                                    >
                                                        <DropdownMenuGroup className="-m-1 rounded-none p-0">
                                                            <InputGroup className="-mb-1 rounded-none border-0 shadow-none [&:has([data-slot=input-group-control]:focus-visible)]:border-0 [&:has([data-slot=input-group-control]:focus-visible)]:ring-0">
                                                                <InputGroupInput
                                                                    onChange={(event) =>
                                                                        setProviderSearch(event.target.value)
                                                                    }
                                                                    onClick={(event) => event.stopPropagation()}
                                                                    onKeyDown={(event) => event.stopPropagation()}
                                                                    placeholder="Search..."
                                                                    value={providerSearch}
                                                                />
                                                                {providerSearch && (
                                                                    <InputGroupAddon align="inline-end">
                                                                        <InputGroupButton
                                                                            onClick={(event) => {
                                                                                event.stopPropagation();
                                                                                setProviderSearch('');
                                                                            }}
                                                                        >
                                                                            <X />
                                                                        </InputGroupButton>
                                                                    </InputGroupAddon>
                                                                )}
                                                            </InputGroup>
                                                            <DropdownMenuSeparator />
                                                        </DropdownMenuGroup>
                                                        <DropdownMenuGroup className="max-h-64 overflow-y-auto">
                                                            {!filteredProviders.length ? (
                                                                <DropdownMenuItem
                                                                    className="min-h-16 justify-center"
                                                                    disabled
                                                                >
                                                                    {providerSearch
                                                                        ? 'No results found'
                                                                        : 'No available providers'}
                                                                </DropdownMenuItem>
                                                            ) : (
                                                                filteredProviders.map((provider) => (
                                                                    <DropdownMenuItem
                                                                        key={provider.name}
                                                                        onSelect={() => {
                                                                            if (isFormDisabled || isProviderDisabled) {
                                                                                return;
                                                                            }

                                                                            providerField.onChange(provider.name);
                                                                            setSelectedProvider(provider);
                                                                            setProviderSearch('');
                                                                        }}
                                                                    >
                                                                        <div className="flex w-full min-w-0 items-center gap-2">
                                                                            <ProviderIcon
                                                                                className="size-4 shrink-0"
                                                                                provider={provider}
                                                                            />

                                                                            <span className="flex-1 truncate">
                                                                                {getProviderDisplayName(provider)}
                                                                            </span>
                                                                            {providerField.value === provider.name && (
                                                                                <Check className="ml-auto size-4 shrink-0" />
                                                                            )}
                                                                        </div>
                                                                    </DropdownMenuItem>
                                                                ))
                                                            )}
                                                        </DropdownMenuGroup>
                                                    </DropdownMenuContent>
                                                </DropdownMenu>
                                            );
                                        }}
                                    />

                                    {type === 'assistant' && (
                                        <FormField
                                            control={control}
                                            name="useAgents"
                                            render={({ field: useAgentsField }) => (
                                                <TooltipProvider>
                                                    <Tooltip>
                                                        <TooltipTrigger asChild>
                                                            <div className="flex items-center">
                                                                <FormControl>
                                                                    <Switch
                                                                        checked={useAgentsField.value}
                                                                        disabled={isFormDisabled}
                                                                        onCheckedChange={useAgentsField.onChange}
                                                                    />
                                                                </FormControl>
                                                                <FormLabel
                                                                    className="flex cursor-pointer pl-2 text-xs font-normal"
                                                                    onClick={() =>
                                                                        useAgentsField.onChange(!useAgentsField.value)
                                                                    }
                                                                >
                                                                    Use Agents
                                                                </FormLabel>
                                                            </div>
                                                        </TooltipTrigger>
                                                        <TooltipContent>
                                                            <p className="max-w-48">
                                                                Enable multi-agent collaboration for complex tasks
                                                            </p>
                                                        </TooltipContent>
                                                    </Tooltip>
                                                </TooltipProvider>
                                            )}
                                        />
                                    )}

                                    <DropdownMenu>
                                        <DropdownMenuTrigger asChild>
                                            <InputGroupButton
                                                disabled={isFormDisabled}
                                                variant="ghost"
                                            >
                                                <FileText className="shrink-0" />
                                                <ChevronDown />
                                            </InputGroupButton>
                                        </DropdownMenuTrigger>
                                        <DropdownMenuContent
                                            align="start"
                                            side="top"
                                        >
                                            <DropdownMenuGroup className="-m-1 rounded-none p-0">
                                                <InputGroup className="-mb-1 rounded-none border-0 shadow-none [&:has([data-slot=input-group-control]:focus-visible)]:border-0 [&:has([data-slot=input-group-control]:focus-visible)]:ring-0">
                                                    <InputGroupInput
                                                        onChange={(event) => setTemplateSearch(event.target.value)}
                                                        onClick={(event) => event.stopPropagation()}
                                                        onKeyDown={(event) => event.stopPropagation()}
                                                        placeholder="Search..."
                                                        value={templateSearch}
                                                    />
                                                    {templateSearch && (
                                                        <InputGroupAddon align="inline-end">
                                                            <InputGroupButton
                                                                onClick={(event) => {
                                                                    event.stopPropagation();
                                                                    setTemplateSearch('');
                                                                }}
                                                            >
                                                                <X />
                                                            </InputGroupButton>
                                                        </InputGroupAddon>
                                                    )}
                                                </InputGroup>
                                                <DropdownMenuSeparator />
                                            </DropdownMenuGroup>
                                            <DropdownMenuGroup className="max-h-64 overflow-y-auto">
                                                {!filteredTemplates.length ? (
                                                    <DropdownMenuItem
                                                        className="min-h-16 justify-center"
                                                        disabled
                                                    >
                                                        {templateSearch ? 'No results found' : 'No available templates'}
                                                    </DropdownMenuItem>
                                                ) : (
                                                    filteredTemplates.map((template) => (
                                                        <DropdownMenuItem
                                                            key={template.id}
                                                            onSelect={() => {
                                                                if (isFormDisabled) {
                                                                    return;
                                                                }

                                                                handleApplyTemplate(template);
                                                            }}
                                                        >
                                                            <span className="max-w-80 flex-1 truncate">
                                                                {template.title}
                                                            </span>
                                                        </DropdownMenuItem>
                                                    ))
                                                )}
                                            </DropdownMenuGroup>
                                        </DropdownMenuContent>
                                    </DropdownMenu>

                                    <DropdownMenu
                                        onOpenChange={(open) => {
                                            if (!open) {
                                                setResourceSearch('');
                                            }
                                        }}
                                    >
                                        <DropdownMenuTrigger asChild>
                                            <InputGroupButton
                                                disabled={isFormDisabled}
                                                variant="ghost"
                                            >
                                                <Paperclip className="shrink-0" />
                                                {flowResources.length > 0 && (
                                                    <span className="bg-muted text-muted-foreground -mx-0.5 flex h-4 min-w-4 items-center justify-center rounded px-1 text-xs font-medium tabular-nums">
                                                        {flowResources.length}
                                                    </span>
                                                )}
                                                <ChevronDown />
                                            </InputGroupButton>
                                        </DropdownMenuTrigger>
                                        <DropdownMenuContent
                                            align="start"
                                            side="top"
                                        >
                                            <DropdownMenuGroup className="-m-1 rounded-none p-0">
                                                <InputGroup className="-mb-1 rounded-none border-0 shadow-none [&:has([data-slot=input-group-control]:focus-visible)]:border-0 [&:has([data-slot=input-group-control]:focus-visible)]:ring-0">
                                                    <InputGroupInput
                                                        onChange={(event) => setResourceSearch(event.target.value)}
                                                        onClick={(event) => event.stopPropagation()}
                                                        onKeyDown={(event) => event.stopPropagation()}
                                                        placeholder="Search..."
                                                        value={resourceSearch}
                                                    />
                                                    {resourceSearch && (
                                                        <InputGroupAddon align="inline-end">
                                                            <InputGroupButton
                                                                onClick={(event) => {
                                                                    event.stopPropagation();
                                                                    setResourceSearch('');
                                                                }}
                                                            >
                                                                <X />
                                                            </InputGroupButton>
                                                        </InputGroupAddon>
                                                    )}
                                                </InputGroup>
                                                <DropdownMenuSeparator />
                                            </DropdownMenuGroup>
                                            <DropdownMenuGroup className="max-h-64 overflow-y-auto">
                                                {!filteredResources.length ? (
                                                    <DropdownMenuItem
                                                        className="min-h-16 justify-center"
                                                        disabled
                                                    >
                                                        {resourceSearch ? 'No results found' : 'No available resources'}
                                                    </DropdownMenuItem>
                                                ) : (
                                                    filteredResources.map((resource) => {
                                                        const isSelected = resourceIds.includes(resource.id);
                                                        const Icon = resource.isDir ? Folder : FileText;

                                                        return (
                                                            <DropdownMenuItem
                                                                key={resource.id}
                                                                onSelect={(event) => {
                                                                    event.preventDefault();

                                                                    if (isFormDisabled) {
                                                                        return;
                                                                    }

                                                                    handleToggleAttachment(resource.id);
                                                                }}
                                                            >
                                                                <div className="flex w-full min-w-0 items-center gap-2">
                                                                    <Icon className="text-muted-foreground size-4 shrink-0" />
                                                                    <div className="flex min-w-0 flex-1 flex-col">
                                                                        <span className="truncate">
                                                                            {resource.name}
                                                                        </span>
                                                                        {resource.path !== resource.name && (
                                                                            <span className="text-muted-foreground truncate text-xs">
                                                                                {resource.path}
                                                                            </span>
                                                                        )}
                                                                    </div>
                                                                    {isSelected && (
                                                                        <Check className="ml-auto size-4 shrink-0" />
                                                                    )}
                                                                </div>
                                                            </DropdownMenuItem>
                                                        );
                                                    })
                                                )}
                                            </DropdownMenuGroup>
                                            <DropdownMenuSeparator />
                                            <DropdownMenuItem
                                                disabled={upload.isUploading}
                                                onSelect={(event) => {
                                                    event.preventDefault();

                                                    if (isFormDisabled) {
                                                        return;
                                                    }

                                                    handleAttachClick();
                                                }}
                                            >
                                                {upload.isUploading ? <Loader2 className="animate-spin" /> : <Plus />}
                                                {upload.isUploading ? 'Uploading…' : 'Upload file'}
                                            </DropdownMenuItem>
                                        </DropdownMenuContent>
                                    </DropdownMenu>

                                    {!isLoading || isSubmitting ? (
                                        <InputGroupButton
                                            className="ml-auto"
                                            disabled={isSubmitting || !isValid || upload.isUploading}
                                            size="icon-xs"
                                            type="submit"
                                            variant="default"
                                        >
                                            {isSubmitting ? <Spinner variant="circle" /> : <ArrowUp />}
                                        </InputGroupButton>
                                    ) : (
                                        <InputGroupButton
                                            className="ml-auto"
                                            disabled={isCanceling || !onCancel}
                                            onClick={() => onCancel?.()}
                                            size="icon-xs"
                                            type="button"
                                            variant="destructive"
                                        >
                                            {isCanceling ? <Spinner variant="circle" /> : <Square />}
                                        </InputGroupButton>
                                    )}
                                </InputGroupAddon>
                            </InputGroup>
                        </FormControl>
                    )}
                />
            </form>
            <input
                className="hidden"
                multiple
                onChange={handleFileInputChange}
                ref={fileInputRef}
                type="file"
            />
            <ConfirmationDialog
                confirmIcon={<FileSymlink />}
                confirmText="Replace"
                confirmVariant="default"
                description="Current message has content. Replace with the selected template?"
                handleConfirm={handleConfirmReplaceTemplate}
                handleOpenChange={(open) => {
                    if (!open) {
                        setPendingTemplate(null);
                    }

                    setIsReplaceConfirmOpen(open);
                }}
                isOpen={isReplaceConfirmOpen}
                title="Replace content?"
            />
        </Form>
    );
};
