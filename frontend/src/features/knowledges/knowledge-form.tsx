import { useMutation } from '@apollo/client/react';
import { Save } from 'lucide-react';
import { type ComponentProps, useCallback, useState } from 'react';
import { type Control, type FieldPath, type SubmitHandler, useWatch } from 'react-hook-form';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { z } from 'zod';

import type {
    CreateKnowledgeDocumentInput,
    KnowledgeDocumentFragmentFragment,
    UpdateKnowledgeDocumentInput,
} from '@/graphql/types';

import { AppHeaderAction } from '@/components/layouts/app/app-header';
import { type EditorViewMode } from '@/components/shared/markdown-editor';
import { UnsavedChangesDialog, useUnsavedChangesGuard } from '@/components/shared/unsaved-changes';
import { Form } from '@/components/ui/form';
import { Spinner } from '@/components/ui/spinner';
import { AnonymizeTextDocument, KnowledgeAnswerType, KnowledgeDocType, KnowledgeGuideType } from '@/graphql/types';
import { useAppForm } from '@/hooks/use-app-form';
import { useBreakpoint } from '@/hooks/use-breakpoint';
import { Log } from '@/lib/log';
import { useUser } from '@/providers/user-provider';

import { KnowledgeFormLayoutDesktop, KnowledgeFormLayoutMobile } from './knowledge-form-layout';
import { KnowledgeHeader } from './knowledge-header';

// Length limits mirror the REST validation tags on the Go side
// (`backend/pkg/server/models/knowledge.go`). The GraphQL layer itself does
// not enforce them, so without these the user could submit a payload that
// later round-trips through REST and gets rejected.
export const KNOWLEDGE_LIMITS = {
    codeLang: 100,
    content: 65536,
    description: 1000,
    question: 2048,
} as const;

// Optional text fields are trimmed and length-checked but NOT collapsed to
// `undefined` — the partial-update logic in `formValuesToUpdateInput` needs
// to distinguish "user cleared a previously-set value" (send `""` so the
// backend clears it) from "field was empty and untouched" (don't send at
// all so the backend leaves it alone). Mapping `"" → undefined` here would
// erase that signal and break the "clear an existing description" use case.
const optionalTrimmed = (max: number, label: string) =>
    z
        .string()
        .trim()
        .max(max, { message: `${label} must be ${max} characters or fewer` })
        .optional();

export const formSchema = z
    .object({
        answerType: z.nativeEnum(KnowledgeAnswerType).optional(),
        codeLang: optionalTrimmed(KNOWLEDGE_LIMITS.codeLang, 'Code language'),
        content: z
            .string()
            .trim()
            .min(1, { message: 'Content is required' })
            .max(KNOWLEDGE_LIMITS.content, {
                message: `Content must be ${KNOWLEDGE_LIMITS.content} characters or fewer`,
            }),
        description: optionalTrimmed(KNOWLEDGE_LIMITS.description, 'Description'),
        docType: z.nativeEnum(KnowledgeDocType),
        guideType: z.nativeEnum(KnowledgeGuideType).optional(),
        question: z
            .string()
            .trim()
            .min(1, { message: 'Question is required' })
            .max(KNOWLEDGE_LIMITS.question, {
                message: `Question must be ${KNOWLEDGE_LIMITS.question} characters or fewer`,
            }),
    })
    .superRefine((value, ctx) => {
        const requiredByDocType: Partial<Record<KnowledgeDocType, { field: FieldPath<FormValues>; message: string }>> =
            {
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

export type FormValues = z.infer<typeof formSchema>;

export const newDocumentDefaults: FormValues = {
    answerType: undefined,
    codeLang: '',
    content: '',
    description: '',
    docType: KnowledgeDocType.Answer,
    guideType: undefined,
    question: '',
};

export const documentToFormValues = (k: KnowledgeDocumentFragmentFragment): FormValues => ({
    answerType: k.answerType ?? undefined,
    codeLang: k.codeLang ?? '',
    content: k.content,
    description: k.description ?? '',
    docType: k.docType,
    guideType: k.guideType ?? undefined,
    question: k.question,
});

export type DirtyFlags = Partial<Record<keyof FormValues, boolean>>;

// CREATE: send all required fields and only non-empty optional fields. There
// is no prior document to "clear", so an empty `description`/`codeLang` just
// means "don't store anything in cmetadata for this field".
export const formValuesToCreateInput = (values: FormValues): CreateKnowledgeDocumentInput => ({
    answerType: values.answerType,
    codeLang: values.codeLang ? values.codeLang : undefined,
    content: values.content,
    description: values.description ? values.description : undefined,
    docType: values.docType,
    guideType: values.guideType,
    question: values.question,
});

// UPDATE: send only fields the user actually edited. `content` is GraphQL-
// required (the backend always re-embeds), so it goes through unconditionally;
// every other field is gated by `dirty`. This way:
//   - untouched fields stay `undefined` and the backend keeps the existing value;
//   - cleared fields go out as `""` so the backend wipes them;
//   - subtype-related fields cleared by `setValue` on docType change are marked
//     dirty by the form, so they reach the backend with the right "clear me" value.
export const formValuesToUpdateInput = (values: FormValues, dirty: DirtyFlags): UpdateKnowledgeDocumentInput => {
    const input: UpdateKnowledgeDocumentInput = { content: values.content };

    if (dirty.docType) {
        input.docType = values.docType;
    }

    if (dirty.question) {
        input.question = values.question;
    }

    if (dirty.description) {
        input.description = values.description ?? '';
    }

    if (dirty.guideType) {
        input.guideType = values.guideType;
    }

    if (dirty.answerType) {
        input.answerType = values.answerType;
    }

    if (dirty.codeLang) {
        input.codeLang = values.codeLang ?? '';
    }

    return input;
};

export interface SubmitResult {
    document?: KnowledgeDocumentFragmentFragment;
    redirectTo?: string;
}

interface KnowledgeFormHeaderProps extends Omit<ComponentProps<typeof KnowledgeHeader>, 'isAnonymizeDisabled'> {
    control: Control<FormValues>;
    isSaving?: boolean;
}

interface KnowledgeFormProps {
    initialValues: FormValues;
    isNew: boolean;
    knowledge?: KnowledgeDocumentFragmentFragment | null;
    onSubmit: (values: FormValues, dirtyFields: DirtyFlags) => Promise<SubmitResult>;
}

export function KnowledgeForm({ initialValues, isNew, knowledge, onSubmit }: KnowledgeFormProps) {
    const navigate = useNavigate();
    const { isDesktop } = useBreakpoint();
    const [isSaving, setIsSaving] = useState(false);
    const [isAnonymizing, setIsAnonymizing] = useState(false);
    const [viewMode, setViewMode] = useState<EditorViewMode>('rich');
    const [anonymizeMutation] = useMutation(AnonymizeTextDocument);
    const { authInfo } = useUser();
    const canAnonymize = authInfo?.privileges?.includes('anonymize.call') ?? false;

    const form = useAppForm<FormValues>({
        defaultValues: initialValues,
        resetOptions: {
            // When `values` changes (e.g. a GraphQL subscription pushes an
            // updated document after an inline rename from the header),
            // refresh the form's defaults but keep any unsaved edits the
            // user is still working on. Without this, an external update
            // would silently wipe their in-flight changes.
            keepDirtyValues: true,
        },
        schema: formSchema,
        // `values` reactively syncs the form with `initialValues`. The page
        // recomputes `initialValues` from `knowledge` whenever the cache
        // refreshes (rename, refetch, etc.), and RHF reapplies the new
        // values on top of the form respecting `resetOptions` above.
        values: initialValues,
    });

    const { control, formState, handleSubmit, reset } = form;
    const { isDirty, isValid } = formState;

    // Navigation is the caller's job — it reads `redirectTo` off the returned result.
    // Pulling it in here would make onSaveFromDialog depend on guard.skipNextBlock and
    // form a real cycle (guard ← onSaveFromDialog ← performSave ← guard.skipNextBlock).
    const performSave = useCallback(
        async (values: FormValues): Promise<null | SubmitResult> => {
            try {
                // Snapshot dirty flags from the latest formState. We read it
                // here (instead of capturing into deps) so partial-update
                // logic in the page sees the same state RHF used to decide
                // `isDirty`/`canSubmit` at submit time.
                const result = await onSubmit(values, form.formState.dirtyFields as DirtyFlags);

                // The backend may trim/normalize fields, so reset to its returned document when present.
                const resetValues = result.document ? documentToFormValues(result.document) : values;

                // Reset BEFORE the caller navigates so `isDirty` is false by the
                // time the blocker re-evaluates (the caller also calls
                // `skipNextBlock` to cover reset's async state propagation).
                reset(resetValues, { keepDefaultValues: false });

                return result;
            } catch (error) {
                Log.error('Failed to save knowledge document', error);

                return null;
            }
        },
        [form, onSubmit, reset],
    );

    const onSaveFromDialog = useCallback(async (): Promise<boolean> => {
        if (isSaving || !isValid) {
            return false;
        }

        // `form.getValues()` returns raw field state (no zod transforms applied),
        // so we run it through the schema explicitly. This way the dialog path
        // produces the same trimmed/normalized values as the form-button path
        // (which gets parsed values directly from `handleSubmit`'s callback).
        const parsed = formSchema.safeParse(form.getValues());

        if (!parsed.success) {
            return false;
        }

        setIsSaving(true);

        try {
            // The guard proceeds the originally-blocked navigation on success;
            // a CREATE's `redirectTo` is intentionally not honored here ("Save
            // and leave" leaves to where the user was going, not the new doc).
            const result = await performSave(parsed.data);

            return result !== null;
        } finally {
            setIsSaving(false);
        }
    }, [form, isSaving, isValid, performSave]);

    const guard = useUnsavedChangesGuard({
        isDirty,
        isFormValid: isValid,
        onSave: onSaveFromDialog,
    });
    const { skipNextBlock } = guard;

    // Defined after `guard` so it can own the post-save redirect (CREATE only)
    // via the stable `skipNextBlock` — the form-button path navigates to the new
    // document; the dialog path above deliberately does not.
    const onSubmitWithGuard: SubmitHandler<FormValues> = useCallback(
        async (values) => {
            if (isSaving) {
                return;
            }

            setIsSaving(true);

            try {
                const result = await performSave(values);

                if (result?.redirectTo) {
                    skipNextBlock();
                    navigate(result.redirectTo);
                }
            } finally {
                setIsSaving(false);
            }
        },
        [isSaving, navigate, performSave, skipNextBlock],
    );

    // Not gated on `isValid`: the button must be clickable to trigger the first submit, which is what turns
    // validation on (mode:'onSubmit'). Clicking an invalid form surfaces the errors instead of silently doing
    // nothing. `isDirty` still gates edits so an unchanged existing doc can't be re-saved.
    const canSubmit = !isSaving && (isNew || isDirty);

    const saveButton = (
        <AppHeaderAction
            disabled={!canSubmit}
            icon={isSaving ? <Spinner variant="circle" /> : <Save aria-hidden="true" />}
            label={isNew ? 'Create' : 'Save'}
            type="submit"
        />
    );

    const handleAnonymize = useCallback(async () => {
        const currentContent = form.getValues('content');

        if (!currentContent?.trim()) {
            return;
        }

        setIsAnonymizing(true);

        try {
            const { data } = await anonymizeMutation({ variables: { text: currentContent } });
            const anonymizedContent = data?.anonymizeText;

            if (anonymizedContent == null) {
                toast.error('Anonymizer returned no result');

                return;
            }

            if (anonymizedContent === currentContent) {
                toast.info('No sensitive data detected');

                return;
            }

            form.setValue('content', anonymizedContent, { shouldDirty: true });
            toast.success('Content anonymized');
        } catch (error) {
            Log.error('Failed to anonymize content', error);
            toast.error(error instanceof Error ? error.message : 'Failed to anonymize content');
        } finally {
            setIsAnonymizing(false);
        }
    }, [anonymizeMutation, form]);

    return (
        <>
            <Form {...form}>
                <form // Desktop: lock to the viewport so the resizable panels
                    // inside the body can fill the remaining space below the
                    // sticky header. Mobile: allow the page to grow with its
                    // content (single column, vertical scroll).
                    className={isDesktop ? 'flex h-[100dvh] min-h-0 w-full flex-col' : 'flex min-h-[100dvh] flex-col'}
                    noValidate
                    onSubmit={handleSubmit(onSubmitWithGuard)}
                >
                    <KnowledgeFormHeader
                        canAnonymize={canAnonymize}
                        control={control}
                        isAnonymizing={isAnonymizing}
                        isNew={isNew}
                        isSaving={isSaving}
                        knowledge={knowledge}
                        onAnonymize={handleAnonymize}
                        onBeforeNavigateAway={skipNextBlock}
                        onModeChange={setViewMode}
                        saveButton={saveButton}
                        viewMode={viewMode}
                    />
                    {isDesktop ? (
                        <KnowledgeFormLayoutDesktop
                            control={control}
                            isNew={isNew}
                            isSaving={isSaving}
                            knowledge={knowledge}
                            viewMode={viewMode}
                        />
                    ) : (
                        <KnowledgeFormLayoutMobile
                            control={control}
                            isNew={isNew}
                            isSaving={isSaving}
                            knowledge={knowledge}
                            viewMode={viewMode}
                        />
                    )}
                </form>
            </Form>
            <UnsavedChangesDialog
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
}

// Don't hoist this useWatch to the parent — it would re-render the whole form per keystroke.
function KnowledgeFormHeader({ control, isAnonymizing, isSaving = false, ...rest }: KnowledgeFormHeaderProps) {
    const content = useWatch({ control, name: 'content' });
    const isAnonymizeDisabled = isAnonymizing || isSaving || !content?.trim();

    return (
        <KnowledgeHeader
            {...rest}
            isAnonymizeDisabled={isAnonymizeDisabled}
            isAnonymizing={isAnonymizing}
        />
    );
}
