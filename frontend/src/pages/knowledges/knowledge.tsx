import { zodResolver } from '@hookform/resolvers/zod';
import { GripVertical, LibraryBig, Save } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { useNavigate, useParams } from 'react-router-dom';
import { z } from 'zod';

import type {
    CreateKnowledgeDocumentInput,
    KnowledgeAnswerType as KnowledgeAnswerTypeT,
    KnowledgeDocumentFragmentFragment,
    KnowledgeGuideType as KnowledgeGuideTypeT,
    UpdateKnowledgeDocumentInput,
} from '@/graphql/types';

import { MarkdownEditor } from '@/components/shared/markdown-editor';
import { Badge } from '@/components/ui/badge';
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { InputGroup, InputGroupTextareaAutosize } from '@/components/ui/input-group';
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { Spinner } from '@/components/ui/spinner';
import { KnowledgeAnswerType, KnowledgeDocType, KnowledgeGuideType, useKnowledgeDocumentQuery } from '@/graphql/types';
import { useBreakpoint } from '@/hooks/use-breakpoint';
import { useKnowledges } from '@/providers/knowledges-provider';

const docTypeValues = [KnowledgeDocType.Answer, KnowledgeDocType.Guide, KnowledgeDocType.Code] as const;
const guideTypeValues = Object.values(KnowledgeGuideType) as KnowledgeGuideTypeT[];
const answerTypeValues = Object.values(KnowledgeAnswerType) as KnowledgeAnswerTypeT[];

const formSchema = z
    .object({
        answerType: z.nativeEnum(KnowledgeAnswerType).optional(),
        codeLang: z.string().trim().optional(),
        content: z.string().trim().min(1, { message: 'Content is required' }),
        description: z.string().trim().optional(),
        docType: z.nativeEnum(KnowledgeDocType),
        guideType: z.nativeEnum(KnowledgeGuideType).optional(),
        question: z.string().trim().min(1, { message: 'Question is required' }),
    })
    .superRefine((value, ctx) => {
        if (value.docType === KnowledgeDocType.Guide && !value.guideType) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: 'Guide type is required',
                path: ['guideType'],
            });
        }

        if (value.docType === KnowledgeDocType.Answer && !value.answerType) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: 'Answer type is required',
                path: ['answerType'],
            });
        }

        if (value.docType === KnowledgeDocType.Code && (!value.codeLang || value.codeLang.length === 0)) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: 'Code language is required',
                path: ['codeLang'],
            });
        }
    });

type FormValues = z.infer<typeof formSchema>;

const newDocumentDefaults: FormValues = {
    answerType: undefined,
    codeLang: '',
    content: '',
    description: '',
    docType: KnowledgeDocType.Answer,
    guideType: undefined,
    question: '',
};

const documentToFormValues = (k: KnowledgeDocumentFragmentFragment): FormValues => ({
    answerType: k.answerType ?? undefined,
    codeLang: k.codeLang ?? '',
    content: k.content,
    description: k.description ?? '',
    docType: k.docType,
    guideType: k.guideType ?? undefined,
    question: k.question,
});

interface KnowledgeFormViewProps {
    initialValues: FormValues;
    isNew: boolean;
    knowledge?: KnowledgeDocumentFragmentFragment | null;
    knowledgeName: null | string;
    onSubmit: (values: FormValues) => Promise<void>;
}

const KnowledgeFormView = ({ initialValues, isNew, knowledge, knowledgeName, onSubmit }: KnowledgeFormViewProps) => {
    const { isDesktop } = useBreakpoint();
    const [isSaving, setIsSaving] = useState(false);

    const form = useForm<FormValues>({
        defaultValues: initialValues,
        mode: 'onChange',
        resolver: zodResolver(formSchema),
    });

    const { control, formState, handleSubmit, reset, watch } = form;

    const docType = watch('docType');
    const hasUnsavedChanges = formState.isDirty;
    const canSubmit = !isSaving && formState.isValid && (isNew || hasUnsavedChanges);

    const onSubmitWithGuard = async (values: FormValues) => {
        if (isSaving) {
            return;
        }

        setIsSaving(true);

        try {
            await onSubmit(values);

            if (!isNew) {
                reset(values, { keepDefaultValues: false });
            }
        } catch {
            // Error already surfaced via toast in the provider
        } finally {
            setIsSaving(false);
        }
    };

    const introBlock = (
        <div className="flex flex-col gap-4">
            <div className="text-center">
                <h1 className="text-2xl font-semibold">
                    {isNew ? 'Create a new knowledge document' : 'Edit knowledge document'}
                </h1>
                <p className="text-muted-foreground mt-2">
                    {isNew
                        ? 'Add an entry to the vector knowledge base'
                        : 'Edits to content or metadata will trigger re-embedding'}
                </p>
            </div>

            {!isNew && knowledge ? (
                <div className="text-muted-foreground flex flex-wrap items-center gap-2 text-xs">
                    <Badge variant={knowledge.manual ? 'secondary' : 'outline'}>
                        {knowledge.manual ? 'manual' : 'agent'}
                    </Badge>
                    {knowledge.flowId ? <Badge variant="outline">flow #{knowledge.flowId}</Badge> : null}
                    {knowledge.taskId ? <Badge variant="outline">task #{knowledge.taskId}</Badge> : null}
                    {knowledge.subtaskId ? <Badge variant="outline">subtask #{knowledge.subtaskId}</Badge> : null}
                    <span>·</span>
                    <span>
                        chunk {knowledge.partSize} of {knowledge.totalSize}
                    </span>
                </div>
            ) : null}
        </div>
    );

    const metaFields = (
        <div className="flex flex-col gap-4">
            <div className="grid gap-4 sm:grid-cols-2">
                <FormField
                    control={control}
                    name="docType"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Document type</FormLabel>
                            {isNew ? (
                                <Select
                                    disabled={isSaving}
                                    onValueChange={field.onChange}
                                    value={field.value}
                                >
                                    <FormControl>
                                        <SelectTrigger>
                                            <SelectValue placeholder="Select type" />
                                        </SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                        {docTypeValues.map((value) => (
                                            <SelectItem
                                                key={value}
                                                value={value}
                                            >
                                                {value}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            ) : (
                                <div className="border-input bg-muted/30 text-muted-foreground flex h-9 items-center rounded-md border px-3 text-sm">
                                    {field.value || '—'}
                                </div>
                            )}
                            <FormMessage />
                        </FormItem>
                    )}
                />

                {docType === KnowledgeDocType.Guide ? (
                    <FormField
                        control={control}
                        name="guideType"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Guide type</FormLabel>
                                <Select
                                    disabled={isSaving}
                                    onValueChange={field.onChange}
                                    value={field.value}
                                >
                                    <FormControl>
                                        <SelectTrigger>
                                            <SelectValue placeholder="Select guide type" />
                                        </SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                        {guideTypeValues.map((value) => (
                                            <SelectItem
                                                key={value}
                                                value={value}
                                            >
                                                {value}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                ) : null}

                {docType === KnowledgeDocType.Answer ? (
                    <FormField
                        control={control}
                        name="answerType"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Answer type</FormLabel>
                                <Select
                                    disabled={isSaving}
                                    onValueChange={field.onChange}
                                    value={field.value}
                                >
                                    <FormControl>
                                        <SelectTrigger>
                                            <SelectValue placeholder="Select answer type" />
                                        </SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                        {answerTypeValues.map((value) => (
                                            <SelectItem
                                                key={value}
                                                value={value}
                                            >
                                                {value}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                ) : null}

                {docType === KnowledgeDocType.Code ? (
                    <FormField
                        control={control}
                        name="codeLang"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Code language</FormLabel>
                                <FormControl>
                                    <Input
                                        disabled={isSaving}
                                        placeholder="e.g. python, go, typescript"
                                        {...field}
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                ) : null}
            </div>

            <FormField
                control={control}
                name="question"
                render={({ field }) => (
                    <FormItem>
                        <FormLabel>Question</FormLabel>
                        <FormControl>
                            <InputGroup className="block">
                                <InputGroupTextareaAutosize
                                    {...field}
                                    autoFocus={isNew}
                                    className="min-h-0"
                                    disabled={isSaving}
                                    maxRows={6}
                                    minRows={1}
                                    placeholder="Short title or question this document answers"
                                />
                            </InputGroup>
                        </FormControl>
                        <FormMessage />
                    </FormItem>
                )}
            />

            <FormField
                control={control}
                name="description"
                render={({ field }) => (
                    <FormItem>
                        <FormLabel>Description (optional)</FormLabel>
                        <FormControl>
                            <InputGroup className="block">
                                <InputGroupTextareaAutosize
                                    {...field}
                                    className="min-h-0"
                                    disabled={isSaving}
                                    maxRows={8}
                                    minRows={1}
                                    placeholder="Optional short description"
                                />
                            </InputGroup>
                        </FormControl>
                        <FormMessage />
                    </FormItem>
                )}
            />
        </div>
    );

    const contentField = (
        <FormField
            control={control}
            name="content"
            render={({ field }) => (
                <FormItem className="flex min-h-0 flex-1 flex-col">
                    <FormControl>
                        <MarkdownEditor
                            className="min-h-0 flex-1"
                            disabled={isSaving}
                            onBlur={field.onBlur}
                            onChange={field.onChange}
                            placeholder="Knowledge content (will be embedded into the vector store)"
                            value={field.value}
                        />
                    </FormControl>
                    <FormMessage />
                </FormItem>
            )}
        />
    );

    const contentFieldStandalone = (
        <FormField
            control={control}
            name="content"
            render={({ field }) => (
                <FormItem>
                    <FormLabel>Content</FormLabel>
                    <FormControl>
                        <MarkdownEditor
                            className="min-h-[280px]"
                            contentClassName="min-h-[240px]"
                            disabled={isSaving}
                            onBlur={field.onBlur}
                            onChange={field.onChange}
                            placeholder="Knowledge content (will be embedded into the vector store)"
                            value={field.value}
                        />
                    </FormControl>
                    <FormMessage />
                </FormItem>
            )}
        />
    );

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
                        <LibraryBig className="size-4 shrink-0" />
                        <BreadcrumbPage className="max-w-[240px] truncate">
                            {isNew ? 'New knowledge' : (knowledgeName ?? 'Knowledge')}
                        </BreadcrumbPage>
                    </BreadcrumbItem>
                </BreadcrumbList>
            </Breadcrumb>
            <div className="ml-auto flex items-center gap-2">
                <Button
                    disabled={!canSubmit}
                    size="sm"
                    type="submit"
                >
                    {isSaving ? <Spinner variant="circle" /> : <Save />}
                    {isNew ? 'Create' : 'Save'}
                </Button>
            </div>
        </header>
    );

    if (isDesktop) {
        return (
            <Form {...form}>
                <form
                    className="flex h-full min-h-0 w-full flex-1 flex-col"
                    onSubmit={handleSubmit(onSubmitWithGuard)}
                >
                    {pageHeader}
                    <div className="flex h-[calc(100dvh-3rem)] min-h-0 w-full max-w-full flex-1 overflow-hidden">
                        <ResizablePanelGroup
                            className="w-full"
                            direction="horizontal"
                        >
                            <ResizablePanel
                                className="h-[calc(100dvh-3rem)] min-h-0"
                                defaultSize={45}
                                minSize={30}
                            >
                                <div className="h-full min-h-0 overflow-y-auto">
                                    <Card className="mx-auto min-h-full w-full max-w-2xl rounded-none border-0">
                                        <CardContent className="flex flex-col gap-6 py-6">
                                            {introBlock}
                                            {metaFields}
                                        </CardContent>
                                    </Card>
                                </div>
                            </ResizablePanel>
                            <ResizableHandle withHandle>
                                <GripVertical className="size-4" />
                            </ResizableHandle>
                            <ResizablePanel
                                className="h-[calc(100dvh-3rem)] min-h-0"
                                defaultSize={55}
                                minSize={30}
                            >
                                <div className="flex h-full min-h-0 flex-col overflow-hidden p-4">{contentField}</div>
                            </ResizablePanel>
                        </ResizablePanelGroup>
                    </div>
                </form>
            </Form>
        );
    }

    return (
        <Form {...form}>
            <form
                className="flex min-h-[100dvh] flex-col"
                onSubmit={handleSubmit(onSubmitWithGuard)}
            >
                {pageHeader}
                <div className="flex min-w-0 flex-1 items-start justify-center p-4">
                    <Card className="w-full max-w-3xl">
                        <CardContent className="flex flex-col gap-6 pt-6">
                            {introBlock}
                            {metaFields}
                            {contentFieldStandalone}
                        </CardContent>
                    </Card>
                </div>
            </form>
        </Form>
    );
};

const Knowledge = () => {
    const navigate = useNavigate();
    const { knowledgeId } = useParams<{ knowledgeId?: string }>();
    const { createKnowledge, updateKnowledge } = useKnowledges();

    const isNew = knowledgeId === 'new';

    const { data, loading: isLoadingKnowledge } = useKnowledgeDocumentQuery({
        skip: isNew || !knowledgeId,
        variables: knowledgeId && !isNew ? { id: knowledgeId } : undefined,
    });

    const knowledge = data?.knowledgeDocument ?? null;
    const knowledgeName = knowledge?.question ?? null;

    const initialValues = useMemo<FormValues>(
        () => (knowledge ? documentToFormValues(knowledge) : newDocumentDefaults),
        [knowledge],
    );

    const handleSubmit = async (values: FormValues) => {
        if (isNew) {
            const input: CreateKnowledgeDocumentInput = {
                answerType: values.docType === KnowledgeDocType.Answer ? values.answerType : undefined,
                codeLang: values.docType === KnowledgeDocType.Code && values.codeLang ? values.codeLang : undefined,
                content: values.content,
                description: values.description ? values.description : undefined,
                docType: values.docType,
                guideType: values.docType === KnowledgeDocType.Guide ? values.guideType : undefined,
                question: values.question,
            };

            const created = await createKnowledge(input);
            navigate(created?.id ? `/knowledges/${created.id}` : '/knowledges');

            return;
        }

        if (!knowledgeId) {
            return;
        }

        const input: UpdateKnowledgeDocumentInput = {
            answerType: values.docType === KnowledgeDocType.Answer ? values.answerType : undefined,
            codeLang: values.docType === KnowledgeDocType.Code && values.codeLang ? values.codeLang : undefined,
            content: values.content,
            description: values.description ? values.description : undefined,
            guideType: values.docType === KnowledgeDocType.Guide ? values.guideType : undefined,
            question: values.question,
        };

        await updateKnowledge(knowledgeId, input);
    };

    const placeholderHeader = (
        <header className="bg-background sticky top-0 z-10 flex h-12 shrink-0 items-center gap-2 border-b px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator
                className="mr-2 h-4"
                orientation="vertical"
            />
            <Breadcrumb>
                <BreadcrumbList>
                    <BreadcrumbItem>
                        <LibraryBig className="size-4 shrink-0" />
                        <BreadcrumbPage className="max-w-[240px] truncate">
                            {isNew ? 'New knowledge' : (knowledgeName ?? 'Knowledge')}
                        </BreadcrumbPage>
                    </BreadcrumbItem>
                </BreadcrumbList>
            </Breadcrumb>
        </header>
    );

    if (!isNew && isLoadingKnowledge) {
        return (
            <>
                {placeholderHeader}
                <div className="flex min-h-[calc(100dvh-3rem)] items-center justify-center">
                    <Spinner variant="circle" />
                </div>
            </>
        );
    }

    if (!isNew && !knowledge) {
        return (
            <>
                {placeholderHeader}
                <div className="flex min-h-[calc(100dvh-3rem)] items-center justify-center p-4">
                    <Card className="w-full max-w-2xl">
                        <CardContent className="flex flex-col items-center gap-4 pt-6 text-center">
                            <h2 className="text-xl font-semibold">Knowledge not found</h2>
                            <p className="text-muted-foreground">
                                The knowledge document you are looking for does not exist.
                            </p>
                            <Button onClick={() => navigate('/knowledges')}>Back to Knowledges</Button>
                        </CardContent>
                    </Card>
                </div>
            </>
        );
    }

    return (
        <KnowledgeFormView
            initialValues={initialValues}
            isNew={isNew}
            key={knowledgeId ?? 'new'}
            knowledge={knowledge}
            knowledgeName={knowledgeName}
            onSubmit={handleSubmit}
        />
    );
};

export default Knowledge;
