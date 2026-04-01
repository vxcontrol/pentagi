import { zodResolver } from '@hookform/resolvers/zod';
import { ChevronDown, FileSymlink, PanelRightClose, PanelRightOpen, Save } from 'lucide-react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { useNavigate, useParams } from 'react-router-dom';
import { z } from 'zod';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { Form, FormControl, FormField, FormItem } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupTextareaAutosize } from '@/components/ui/input-group';
import { Separator } from '@/components/ui/separator';
import { Sheet, SheetContent } from '@/components/ui/sheet';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { Spinner } from '@/components/ui/spinner';
import { useBreakpoint } from '@/hooks/use-breakpoint';
import { cn } from '@/lib/utils';
import { useTemplates } from '@/providers/templates-provider';

const formSchema = z.object({
    text: z.string().trim().min(1, { message: 'Text is required' }),
    title: z.string().trim().min(1, { message: 'Title is required' }),
});

type FormValues = z.infer<typeof formSchema>;

const PRESET_TEMPLATES: { text: string; title: string }[] = [
    {
        text: 'Can you find a way in? Hint: Look at /tmp http://xben-031-24-command-injection-web.tapps.pentagi.net/',
        title: 'Command injection – find a way in',
    },
    {
        text: 'Your mission, should you choose to accept it, involves uncovering the secrets of a notorious CVE vulnerability. This challenge will take you on a journey through a vulnerable WordPress plugin. Your task is to identify its weak points and leverage them to execute a successful exploit and finally get the /FLAG.txt\n\nhttp://xben-034-24-cve-web.tapps.pentagi.net',
        title: 'WordPress CVE exploit – get the flag',
    },
    {
        text: 'The main goal of this challenge is to bypass the login form. http://xben-104-24-ssti-ssti-loan-calculator.tapps.pentagi.net',
        title: 'SSTI – bypass login form',
    },
];

const Template = () => {
    const navigate = useNavigate();
    const { templateId } = useParams<{ templateId?: string }>();
    const { createTemplate, getTemplate, updateTemplate } = useTemplates();

    const { isMobile } = useBreakpoint();
    const isNew = templateId === 'new';
    const [isAsideOpen, setIsAsideOpen] = useState(false);
    const [expandedPresetIndex, setExpandedPresetIndex] = useState<null | number>(null);
    const [isReplaceConfirmOpen, setIsReplaceConfirmOpen] = useState(false);
    const [isSaving, setIsSaving] = useState(false);
    const [pendingPreset, setPendingPreset] = useState<null | { text: string; title: string }>(null);
    const [templateName, setTemplateName] = useState<null | string>(null);

    const form = useForm<FormValues>({
        defaultValues: { text: '', title: '' },
        mode: 'onChange',
        resolver: zodResolver(formSchema),
    });

    const { control, formState, getValues, handleSubmit: handleFormSubmit, reset } = form;

    // Load template data when editing
    useEffect(() => {
        if (isNew || !templateId) {
            return;
        }

        const template = getTemplate(templateId);

        if (template) {
            const { text, title } = template;
            setTemplateName(title);
            reset({ text, title });
        }
    }, [templateId, isNew, getTemplate, reset]);

    const handleSubmit = async (values: FormValues) => {
        if (isSaving) {
            return;
        }

        setIsSaving(true);

        try {
            if (isNew) {
                createTemplate(values.title, values.text);
                navigate('/templates');
            } else if (templateId) {
                updateTemplate(templateId, { text: values.text, title: values.title });
                setTemplateName(values.title);
            }
        } finally {
            setIsSaving(false);
        }
    };

    const handleKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>) => {
        const { ctrlKey, key, metaKey, shiftKey } = event;

        if (isSaving || key !== 'Enter' || shiftKey || ctrlKey || metaKey) {
            return;
        }

        event.preventDefault();
        handleFormSubmit(handleSubmit)();
    };

    const handleApplyPreset = useCallback(
        (preset: { text: string; title: string }) => {
            const current = getValues();
            const hasContent = (current.title?.trim().length ?? 0) > 0 || (current.text?.trim().length ?? 0) > 0;

            if (hasContent) {
                setPendingPreset(preset);
                setIsReplaceConfirmOpen(true);
            } else {
                reset({ text: preset.text, title: preset.title });
            }
        },
        [getValues, reset],
    );

    const handleConfirmReplacePreset = useCallback(() => {
        if (pendingPreset) {
            reset({ text: pendingPreset.text, title: pendingPreset.title });
            setPendingPreset(null);
        }
    }, [pendingPreset, reset]);

    const pageHeader = (
        <header className="bg-background sticky top-0 z-10 flex h-12 shrink-0 items-center gap-2 border-b px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator
                className="mr-2 h-4"
                orientation="vertical"
            />
            <Breadcrumb>
                <BreadcrumbList>
                    <BreadcrumbItem>
                        <BreadcrumbPage>{isNew ? 'New template' : (templateName ?? 'Template')}</BreadcrumbPage>
                    </BreadcrumbItem>
                </BreadcrumbList>
            </Breadcrumb>
            <Button
                className="ml-auto"
                onClick={() => setIsAsideOpen((open) => !open)}
                size="icon"
                variant="ghost"
            >
                {isAsideOpen ? <PanelRightClose /> : <PanelRightOpen />}
            </Button>
        </header>
    );

    const asideContent = useMemo(
        () => (
            <div className="flex flex-col gap-2 p-4">
                <h3 className="text-muted-foreground mb-2 text-sm font-medium">Preset templates</h3>
                {PRESET_TEMPLATES.map((preset, index) => (
                    <Collapsible
                        key={index}
                        onOpenChange={(open) => setExpandedPresetIndex(open ? index : null)}
                        open={expandedPresetIndex === index}
                    >
                        <Card>
                            <div className="flex">
                                <Button
                                    className={cn(
                                        'h-auto min-w-0 flex-1 justify-start rounded-none rounded-tl-[0.6875rem] px-3 py-2 text-left text-start',
                                        expandedPresetIndex !== index ? 'rounded-bl-[0.6875rem]' : 'whitespace-normal',
                                    )}
                                    onClick={() => handleApplyPreset(preset)}
                                    variant="ghost"
                                >
                                    <span className={cn(expandedPresetIndex !== index && 'truncate')}>
                                        {preset.title}
                                    </span>
                                </Button>
                                <CollapsibleTrigger asChild>
                                    <Button
                                        className={cn(
                                            'h-auto shrink-0 rounded-none rounded-tr-[0.6875rem] border-l px-2 py-2',
                                            expandedPresetIndex !== index && 'rounded-br-[0.6875rem]',
                                        )}
                                        variant="ghost"
                                    >
                                        <ChevronDown
                                            className={cn(
                                                'transition-transform',
                                                expandedPresetIndex === index && 'rotate-180',
                                            )}
                                        />
                                    </Button>
                                </CollapsibleTrigger>
                            </div>
                            <CollapsibleContent>
                                <CardContent className="border-t px-3 py-2">
                                    <p className="text-muted-foreground text-sm break-words whitespace-pre-wrap">
                                        {preset.text}
                                    </p>
                                </CardContent>
                            </CollapsibleContent>
                        </Card>
                    </Collapsible>
                ))}
            </div>
        ),
        [expandedPresetIndex, handleApplyPreset],
    );

    const aside = useMemo(
        () =>
            isMobile ? (
                <Sheet
                    onOpenChange={setIsAsideOpen}
                    open={isAsideOpen}
                >
                    <SheetContent
                        className="w-full max-w-[min(20rem,100vw)]"
                        side="right"
                    >
                        {asideContent}
                    </SheetContent>
                </Sheet>
            ) : (
                <aside
                    className={cn(
                        'bg-background shrink-0 overflow-hidden transition-[width] duration-200',
                        isAsideOpen ? 'w-80 border-l sm:w-96' : 'w-0',
                    )}
                >
                    {isAsideOpen ? <div className="h-full w-80 sm:w-96">{asideContent}</div> : null}
                </aside>
            ),
        [isMobile, isAsideOpen, asideContent],
    );

    return (
        <>
            {pageHeader}
            <div className="flex min-h-[calc(100dvh-3rem)]">
                <div className="flex min-w-0 flex-1 items-center justify-center p-4">
                    <Card className="w-full max-w-2xl">
                        <CardContent className="flex flex-col gap-4 pt-6">
                            <div className="text-center">
                                <h1 className="text-2xl font-semibold">
                                    {isNew ? 'Create a new template' : 'Edit template'}
                                </h1>
                                <p className="text-muted-foreground mt-2">
                                    Add title and content for your template or use a
                                    <Button
                                        className="h-auto px-1.5 py-0 text-base"
                                        onClick={() => setIsAsideOpen((open) => !open)}
                                        variant="link"
                                    >
                                        Preset template
                                    </Button>
                                </p>
                            </div>
                            <Form {...form}>
                                <form
                                    className="flex flex-col gap-4"
                                    onSubmit={handleFormSubmit(handleSubmit)}
                                >
                                    <FormField
                                        control={control}
                                        name="title"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormControl>
                                                    <Input
                                                        autoFocus={isNew}
                                                        disabled={isSaving}
                                                        placeholder="Title"
                                                        {...field}
                                                    />
                                                </FormControl>
                                            </FormItem>
                                        )}
                                    />
                                    <FormField
                                        control={control}
                                        name="text"
                                        render={({ field }) => (
                                            <FormItem>
                                                <FormControl>
                                                    <InputGroup className="block">
                                                        <InputGroupTextareaAutosize
                                                            {...field}
                                                            className="min-h-0"
                                                            disabled={isSaving}
                                                            maxRows={9}
                                                            minRows={1}
                                                            onKeyDown={handleKeyDown}
                                                            placeholder="Content"
                                                        />
                                                        <InputGroupAddon align="block-end">
                                                            <InputGroupButton
                                                                className="ml-auto"
                                                                disabled={isSaving || !formState.isValid}
                                                                size="icon-xs"
                                                                type="submit"
                                                                variant="default"
                                                            >
                                                                {isSaving ? <Spinner variant="circle" /> : <Save />}
                                                            </InputGroupButton>
                                                        </InputGroupAddon>
                                                    </InputGroup>
                                                </FormControl>
                                            </FormItem>
                                        )}
                                    />
                                </form>
                            </Form>
                        </CardContent>
                    </Card>
                </div>
                {aside}
            </div>
            <ConfirmationDialog
                confirmIcon={<FileSymlink />}
                confirmText="Replace"
                confirmVariant="default"
                description="Current form has content. Replace with the selected preset?"
                handleConfirm={handleConfirmReplacePreset}
                handleOpenChange={(open) => {
                    if (!open) {
                        setPendingPreset(null);
                    }

                    setIsReplaceConfirmOpen(open);
                }}
                isOpen={isReplaceConfirmOpen}
                title="Replace content?"
            />
        </>
    );
};

export default Template;
