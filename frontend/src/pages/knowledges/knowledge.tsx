import { zodResolver } from '@hookform/resolvers/zod';
import { GripVertical, LibraryBig, Save } from 'lucide-react';
import { type ReactNode, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { type Control, type FieldPath, type SubmitHandler, useForm, useWatch } from 'react-hook-form';
import { type BlockerFunction, useBlocker, useNavigate, useParams } from 'react-router-dom';
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
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
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
import { Log } from '@/lib/log';
import { useKnowledges } from '@/providers/knowledges-provider';

// ---------------------------------------------------------------------------
// Schema, types, pure helpers
// ---------------------------------------------------------------------------

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
        const requiredByDocType: Partial<Record<KnowledgeDocType, { field: FieldPath<FormValues>; message: string }>> = {
            [KnowledgeDocType.Answer]: { field: 'answerType', message: 'Answer type is required' },
            [KnowledgeDocType.Code]: { field: 'codeLang', message: 'Code language is required' },
            [KnowledgeDocType.Guide]: { field: 'guideType', message: 'Guide type is required' },
        };

        const rule = requiredByDocType[value.docType];

        if (!rule) {
            return;
        }

        const fieldValue = value[rule.field];
        const isMissing = fieldValue === undefined || fieldValue === null || fieldValue === '';

        if (isMissing) {
            ctx.addIssue({
                code: z.ZodIssueCode.custom,
                message: rule.message,
                path: [rule.field],
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

// `description` and `codeLang` are optional and we want empty strings to map
// to `undefined` so the backend treats them as "absent" instead of "set to ''".
const trimmedOrUndefined = (value: null | string | undefined): string | undefined => {
    if (!value) {
        return undefined;
    }

    const trimmed = value.trim();

    return trimmed.length === 0 ? undefined : trimmed;
};

const formValuesToCreateInput = (v: FormValues): CreateKnowledgeDocumentInput => ({
    answerType: v.docType === KnowledgeDocType.Answer ? v.answerType : undefined,
    codeLang: v.docType === KnowledgeDocType.Code ? trimmedOrUndefined(v.codeLang) : undefined,
    content: v.content,
    description: trimmedOrUndefined(v.description),
    docType: v.docType,
    guideType: v.docType === KnowledgeDocType.Guide ? v.guideType : undefined,
    question: v.question,
});

const formValuesToUpdateInput = (v: FormValues): UpdateKnowledgeDocumentInput => ({
    answerType: v.docType === KnowledgeDocType.Answer ? v.answerType : undefined,
    codeLang: v.docType === KnowledgeDocType.Code ? trimmedOrUndefined(v.codeLang) : undefined,
    content: v.content,
    description: trimmedOrUndefined(v.description),
    guideType: v.docType === KnowledgeDocType.Guide ? v.guideType : undefined,
    question: v.question,
});

// ---------------------------------------------------------------------------
// Unsaved-changes guard hook
// ---------------------------------------------------------------------------

interface UnsavedChangesGuard {
    handleCancel: () => void;
    handleDiscard: () => void;
    handleOpenChange: (open: boolean) => void;
    handleSaveAndLeave: () => Promise<void>;
    isOpen: boolean;
    isSavingFromDialog: boolean;
    /**
     * Allows the next router navigation to bypass the blocker. Use this after
     * a successful save when the form (or its parent) needs to navigate to
     * a fresh URL (e.g. the new document page) without showing the dialog.
     */
    skipNextBlock: () => void;
}

interface UseUnsavedChangesGuardArgs {
    isDirty: boolean;
    isFormValid: boolean;
    onSave: () => Promise<boolean>;
}

const useUnsavedChangesGuard = ({
    isDirty,
    isFormValid,
    onSave,
}: UseUnsavedChangesGuardArgs): UnsavedChangesGuard => {
    const allowNextRef = useRef(false);
    const isDirtyRef = useRef(isDirty);

    useEffect(() => {
        isDirtyRef.current = isDirty;
    }, [isDirty]);

    const blockerFn = useCallback<BlockerFunction>(({ currentLocation, nextLocation }) => {
        if (allowNextRef.current) {
            allowNextRef.current = false;

            return false;
        }

        if (currentLocation.pathname === nextLocation.pathname && currentLocation.search === nextLocation.search) {
            return false;
        }

        return isDirtyRef.current;
    }, []);

    const blocker = useBlocker(blockerFn);
    const [isSavingFromDialog, setIsSavingFromDialog] = useState(false);

    useEffect(() => {
        if (!isDirty) {
            return;
        }

        const handleBeforeUnload = (event: BeforeUnloadEvent) => {
            event.preventDefault();
            event.returnValue = '';
        };

        window.addEventListener('beforeunload', handleBeforeUnload);

        return () => {
            window.removeEventListener('beforeunload', handleBeforeUnload);
        };
    }, [isDirty]);

    const isOpen = blocker.state === 'blocked';

    const handleCancel = useCallback(() => {
        if (blocker.state === 'blocked') {
            blocker.reset();
        }
    }, [blocker]);

    const handleDiscard = useCallback(() => {
        if (blocker.state === 'blocked') {
            blocker.proceed();
        }
    }, [blocker]);

    const handleSaveAndLeave = useCallback(async () => {
        if (isSavingFromDialog || !isFormValid) {
            return;
        }

        setIsSavingFromDialog(true);

        try {
            const success = await onSave();

            if (success && blocker.state === 'blocked') {
                blocker.proceed();
            }
        } finally {
            setIsSavingFromDialog(false);
        }
    }, [blocker, isFormValid, isSavingFromDialog, onSave]);

    const handleOpenChange = useCallback(
        (open: boolean) => {
            if (isSavingFromDialog) {
                return;
            }

            if (!open && blocker.state === 'blocked') {
                blocker.reset();
            }
        },
        [blocker, isSavingFromDialog],
    );

    const skipNextBlock = useCallback(() => {
        allowNextRef.current = true;
    }, []);

    return {
        handleCancel,
        handleDiscard,
        handleOpenChange,
        handleSaveAndLeave,
        isOpen,
        isSavingFromDialog,
        skipNextBlock,
    };
};

// ---------------------------------------------------------------------------
// Sub-components
// ---------------------------------------------------------------------------

interface KnowledgePageHeaderProps {
    isNew: boolean;
    knowledgeName: null | string;
    saveButton?: ReactNode;
}

const KnowledgePageHeader = ({ isNew, knowledgeName, saveButton }: KnowledgePageHeaderProps) => (
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
        {saveButton ? <div className="ml-auto flex items-center gap-2">{saveButton}</div> : null}
    </header>
);

interface KnowledgeIntroBlockProps {
    isNew: boolean;
    knowledge?: KnowledgeDocumentFragmentFragment | null;
}

const KnowledgeIntroBlock = ({ isNew, knowledge }: KnowledgeIntroBlockProps) => (
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

interface KnowledgeMetaFieldsProps {
    control: Control<FormValues>;
    isNew: boolean;
    isSaving: boolean;
}

const KnowledgeMetaFields = ({ control, isNew, isSaving }: KnowledgeMetaFieldsProps) => {
    // Targeted subscription: only this component re-renders when docType changes,
    // not the whole form. The full-form `useWatch` from the original code
    // re-rendered on every keystroke in the markdown editor.
    const docType = useWatch({ control, name: 'docType' });

    return (
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
};

interface KnowledgeContentFieldProps {
    control: Control<FormValues>;
    /** When `true`, the editor stretches to fill its parent (desktop split view). */
    fillParent?: boolean;
    isSaving: boolean;
    showLabel?: boolean;
}

const KnowledgeContentField = ({ control, fillParent = false, isSaving, showLabel = false }: KnowledgeContentFieldProps) => (
    <FormField
        control={control}
        name="content"
        render={({ field }) => (
            <FormItem className={fillParent ? 'flex min-h-0 flex-1 flex-col' : undefined}>
                {showLabel ? <FormLabel>Content</FormLabel> : null}
                <FormControl>
                    <MarkdownEditor
                        className={fillParent ? 'min-h-0 flex-1' : 'min-h-[280px]'}
                        contentClassName={fillParent ? undefined : 'min-h-[240px]'}
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

interface KnowledgeBodyProps {
    control: Control<FormValues>;
    isNew: boolean;
    isSaving: boolean;
    knowledge?: KnowledgeDocumentFragmentFragment | null;
}

const KnowledgeBodyDesktop = ({ control, isNew, isSaving, knowledge }: KnowledgeBodyProps) => (
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
                            <KnowledgeIntroBlock
                                isNew={isNew}
                                knowledge={knowledge}
                            />
                            <KnowledgeMetaFields
                                control={control}
                                isNew={isNew}
                                isSaving={isSaving}
                            />
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
                <div className="flex h-full min-h-0 flex-col overflow-hidden p-4">
                    <KnowledgeContentField
                        control={control}
                        fillParent
                        isSaving={isSaving}
                    />
                </div>
            </ResizablePanel>
        </ResizablePanelGroup>
    </div>
);

const KnowledgeBodyMobile = ({ control, isNew, isSaving, knowledge }: KnowledgeBodyProps) => (
    <div className="flex min-w-0 flex-1 items-start justify-center p-4">
        <Card className="w-full max-w-3xl">
            <CardContent className="flex flex-col gap-6 pt-6">
                <KnowledgeIntroBlock
                    isNew={isNew}
                    knowledge={knowledge}
                />
                <KnowledgeMetaFields
                    control={control}
                    isNew={isNew}
                    isSaving={isSaving}
                />
                <KnowledgeContentField
                    control={control}
                    isSaving={isSaving}
                    showLabel
                />
            </CardContent>
        </Card>
    </div>
);

interface KnowledgeLeaveDialogProps extends Pick<UnsavedChangesGuard, 'handleCancel' | 'handleDiscard' | 'handleOpenChange' | 'handleSaveAndLeave' | 'isOpen' | 'isSavingFromDialog'> {
    canSave: boolean;
}

const KnowledgeLeaveDialog = ({
    canSave,
    handleCancel,
    handleDiscard,
    handleOpenChange,
    handleSaveAndLeave,
    isOpen,
    isSavingFromDialog,
}: KnowledgeLeaveDialogProps) => (
    <Dialog
        onOpenChange={handleOpenChange}
        open={isOpen}
    >
        <DialogContent
            className="sm:max-w-md"
            onEscapeKeyDown={(event) => {
                if (isSavingFromDialog) {
                    event.preventDefault();
                }
            }}
            onInteractOutside={(event) => {
                if (isSavingFromDialog) {
                    event.preventDefault();
                }
            }}
        >
            <DialogHeader>
                <DialogTitle>Unsaved changes</DialogTitle>
                <DialogDescription>
                    You have unsaved changes on this page. Would you like to save them before leaving?
                </DialogDescription>
            </DialogHeader>
            <DialogFooter className="flex-col-reverse gap-2 sm:flex-row sm:justify-end">
                <Button
                    disabled={isSavingFromDialog}
                    onClick={handleCancel}
                    variant="outline"
                >
                    Cancel
                </Button>
                <Button
                    disabled={isSavingFromDialog}
                    onClick={handleDiscard}
                    variant="destructive"
                >
                    Discard
                </Button>
                <Button
                    disabled={isSavingFromDialog || !canSave}
                    onClick={handleSaveAndLeave}
                    variant="default"
                >
                    {isSavingFromDialog ? <Spinner variant="circle" /> : <Save />}
                    Save
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
);

// ---------------------------------------------------------------------------
// Form view (uses the form, owns navigation guard)
// ---------------------------------------------------------------------------

interface KnowledgeFormViewProps {
    initialValues: FormValues;
    isNew: boolean;
    knowledge?: KnowledgeDocumentFragmentFragment | null;
    knowledgeName: null | string;
    onSubmit: (values: FormValues) => Promise<SubmitResult>;
}

interface SubmitResult {
    redirectTo?: string;
}

const KnowledgeFormView = ({ initialValues, isNew, knowledge, knowledgeName, onSubmit }: KnowledgeFormViewProps) => {
    const navigate = useNavigate();
    const { isDesktop } = useBreakpoint();
    const [isSaving, setIsSaving] = useState(false);

    const form = useForm<FormValues>({
        defaultValues: initialValues,
        mode: 'onChange',
        resolver: zodResolver(formSchema),
    });

    const { control, formState, handleSubmit, reset } = form;
    const { isDirty, isValid } = formState;

    const performSave = useCallback(
        async (values: FormValues): Promise<boolean> => {
            try {
                const result = await onSubmit(values);

                // Reset BEFORE navigate so `isDirty` is false by the time the
                // blocker re-evaluates. We also `skipNextBlock` defensively
                // because reset's state propagation is async.
                reset(values, { keepDefaultValues: false });

                if (result.redirectTo) {
                    skipNextBlockRef.current();
                    navigate(result.redirectTo);
                }

                return true;
            } catch (error) {
                Log.error('Failed to save knowledge document', error);

                return false;
            }
        },
        [navigate, onSubmit, reset],
    );

    // The ref below breaks an otherwise circular hook dependency:
    //
    //   performSave           → skipNextBlockRef.current()        (ref filled by effect below)
    //   onSaveFromDialog      → performSave
    //   useUnsavedChangesGuard({ onSave: onSaveFromDialog }) → exposes skipNextBlock
    //   useEffect             → wires the exposed skipNextBlock back into the ref
    //
    // Replacing the ref with a plain dep would force `performSave` to depend
    // on `guard.skipNextBlock`, which is produced by a hook (`guard`) whose
    // own input (`onSave`) closes over `performSave` — a real cycle that
    // can't be expressed in deps without `useRef`.
    const skipNextBlockRef = useRef<() => void>(() => {});

    const onSubmitWithGuard: SubmitHandler<FormValues> = useCallback(
        async (values) => {
            if (isSaving) {
                return;
            }

            setIsSaving(true);

            try {
                await performSave(values);
            } finally {
                setIsSaving(false);
            }
        },
        [isSaving, performSave],
    );

    const onSaveFromDialog = useCallback(async (): Promise<boolean> => {
        if (isSaving || !isValid) {
            return false;
        }

        setIsSaving(true);

        try {
            return await performSave(form.getValues());
        } finally {
            setIsSaving(false);
        }
    }, [form, isSaving, isValid, performSave]);

    const guard = useUnsavedChangesGuard({
        isDirty,
        isFormValid: isValid,
        onSave: onSaveFromDialog,
    });

    // Wire the ref so `performSave` can call into the guard's `skipNextBlock`.
    useEffect(() => {
        skipNextBlockRef.current = guard.skipNextBlock;
    }, [guard.skipNextBlock]);

    const canSubmit = !isSaving && isValid && (isNew || isDirty);

    const saveButton = (
        <Button
            disabled={!canSubmit}
            size="sm"
            type="submit"
        >
            {isSaving ? <Spinner variant="circle" /> : <Save />}
            {isNew ? 'Create' : 'Save'}
        </Button>
    );

    return (
        <>
            <Form {...form}>
                <form
                    className={isDesktop ? 'flex h-full min-h-0 w-full flex-1 flex-col' : 'flex min-h-[100dvh] flex-col'}
                    onSubmit={handleSubmit(onSubmitWithGuard)}
                >
                    <KnowledgePageHeader
                        isNew={isNew}
                        knowledgeName={knowledgeName}
                        saveButton={saveButton}
                    />
                    {isDesktop ? (
                        <KnowledgeBodyDesktop
                            control={control}
                            isNew={isNew}
                            isSaving={isSaving}
                            knowledge={knowledge}
                        />
                    ) : (
                        <KnowledgeBodyMobile
                            control={control}
                            isNew={isNew}
                            isSaving={isSaving}
                            knowledge={knowledge}
                        />
                    )}
                </form>
            </Form>
            <KnowledgeLeaveDialog
                canSave={isValid}
                handleCancel={guard.handleCancel}
                handleDiscard={guard.handleDiscard}
                handleOpenChange={guard.handleOpenChange}
                handleSaveAndLeave={guard.handleSaveAndLeave}
                isOpen={guard.isOpen}
                isSavingFromDialog={guard.isSavingFromDialog}
            />
        </>
    );
};

// ---------------------------------------------------------------------------
// Page container
// ---------------------------------------------------------------------------

const Knowledge = () => {
    const navigate = useNavigate();
    const { knowledgeId } = useParams<{ knowledgeId?: string }>();
    const { createKnowledge, updateKnowledge } = useKnowledges();

    const isNew = knowledgeId === 'new';
    const shouldFetch = Boolean(knowledgeId) && !isNew;

    const { data, loading: isLoadingKnowledge } = useKnowledgeDocumentQuery({
        skip: !shouldFetch,
        variables: shouldFetch && knowledgeId ? { id: knowledgeId } : undefined,
    });

    const knowledge = data?.knowledgeDocument ?? null;
    const knowledgeName = knowledge?.question ?? null;

    const initialValues = useMemo<FormValues>(
        () => (knowledge ? documentToFormValues(knowledge) : newDocumentDefaults),
        [knowledge],
    );

    const handleSubmit = useCallback(
        async (values: FormValues): Promise<SubmitResult> => {
            if (isNew) {
                const created = await createKnowledge(formValuesToCreateInput(values));

                return {
                    redirectTo: created?.id ? `/knowledges/${created.id}` : '/knowledges',
                };
            }

            if (!knowledgeId) {
                return {};
            }

            await updateKnowledge(knowledgeId, formValuesToUpdateInput(values));

            return {};
        },
        [createKnowledge, isNew, knowledgeId, updateKnowledge],
    );

    if (!isNew && isLoadingKnowledge) {
        return (
            <>
                <KnowledgePageHeader
                    isNew={false}
                    knowledgeName={knowledgeName}
                />
                <div className="flex min-h-[calc(100dvh-3rem)] items-center justify-center">
                    <Spinner variant="circle" />
                </div>
            </>
        );
    }

    if (!isNew && !knowledge) {
        return (
            <>
                <KnowledgePageHeader
                    isNew={false}
                    knowledgeName={knowledgeName}
                />
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
