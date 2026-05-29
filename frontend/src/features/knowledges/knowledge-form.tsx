import { zodResolver } from '@hookform/resolvers/zod';
import { Save } from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { type FieldPath, type SubmitHandler, useForm } from 'react-hook-form';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { z } from 'zod';

import type {
    CreateKnowledgeDocumentInput,
    KnowledgeDocumentFragmentFragment,
    UpdateKnowledgeDocumentInput,
} from '@/graphql/types';

import { HeaderButton } from '@/components/shared/header-button';
import { UnsavedChangesDialog, useUnsavedChangesGuard } from '@/components/shared/unsaved-changes';
import { Form } from '@/components/ui/form';
import { Spinner } from '@/components/ui/spinner';
import { KnowledgeAnswerType, KnowledgeDocType, KnowledgeGuideType, useAnonymizeTextMutation } from '@/graphql/types';
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

// react-hook-form's `dirtyFields` is a partial map of the same shape as
// `FormValues`, with `true` for fields the user actually changed compared to
// `defaultValues`. We project it onto the (flat) FormValues keys here.
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
//   - subtype-related fields cleared by `setValue` on docType change are
//     marked dirty by the form, so they reach the backend with the right
//     "clear me" value (the backend additionally wipes mismatching subtypes
//     itself, but we mirror the user-visible state explicitly).
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
    const [anonymizeMutation] = useAnonymizeTextMutation();
    const { authInfo } = useUser();
    const canAnonymize = authInfo?.privileges?.includes('anonymize.call') ?? false;

    const form = useForm<FormValues>({
        defaultValues: initialValues,
        // `onTouched` validates a field on its first blur and on every change
        // afterwards. With `onChange` we'd run the entire Zod schema on every
        // keystroke (including every emit from the multi-kilobyte `content`
        // markdown editor) — same UX after the first interaction, no waste
        // on initial mount or untouched fields.
        mode: 'onTouched',
        resetOptions: {
            // When `values` changes (e.g. a GraphQL subscription pushes an
            // updated document after an inline rename from the header),
            // refresh the form's defaults but keep any unsaved edits the
            // user is still working on. Without this, an external update
            // would silently wipe their in-flight changes.
            keepDirtyValues: true,
        },
        resolver: zodResolver(formSchema),
        // `values` reactively syncs the form with `initialValues`. The page
        // recomputes `initialValues` from `knowledge` whenever the cache
        // refreshes (rename, refetch, etc.), and RHF reapplies the new
        // values on top of the form respecting `resetOptions` above.
        values: initialValues,
    });

    const { control, formState, handleSubmit, reset } = form;
    const { isDirty, isValid } = formState;

    const performSave = useCallback(
        async (values: FormValues): Promise<boolean> => {
            try {
                // Snapshot dirty flags from the latest formState. We read it
                // here (instead of capturing into deps) so partial-update
                // logic in the page sees the same state RHF used to decide
                // `isDirty`/`canSubmit` at submit time.
                const result = await onSubmit(values, form.formState.dirtyFields as DirtyFlags);

                // Prefer the server's view of the document — backend may have
                // trimmed/normalized fields, attached derived data, or filled
                // optional fields. Falling back to the local `values` keeps
                // the form stable when the mutation hook can't return the
                // saved fragment for some reason.
                const resetValues = result.document ? documentToFormValues(result.document) : values;

                // Reset BEFORE navigate so `isDirty` is false by the time the
                // blocker re-evaluates. We also `skipNextBlock` defensively
                // because reset's state propagation is async.
                reset(resetValues, { keepDefaultValues: false });

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
        [form, navigate, onSubmit, reset],
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
            return await performSave(parsed.data);
        } finally {
            setIsSaving(false);
        }
    }, [form, isSaving, isValid, performSave]);

    const guard = useUnsavedChangesGuard({
        isDirty,
        isFormValid: isValid,
        onSave: onSaveFromDialog,
    });

    useEffect(() => {
        skipNextBlockRef.current = guard.skipNextBlock;
    }, [guard.skipNextBlock]);

    const canSubmit = !isSaving && isValid && (isNew || isDirty);

    const saveButton = (
        <HeaderButton
            disabled={!canSubmit}
            icon={isSaving ? <Spinner variant="circle" /> : <Save aria-hidden="true" />}
            label={isNew ? 'Create' : 'Save'}
            type="submit"
        />
    );

    // Subscribe to `content` so the anonymize button toggles its disabled
    // state as the user types. `form.watch('content')` triggers a re-render
    // on every keystroke, which is what we want for snappy UX.
    const contentValue = form.watch('content');
    const isAnonymizeDisabled = isAnonymizing || isSaving || !contentValue?.trim();

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

            form.setValue('content', anonymizedContent, { shouldDirty: true, shouldValidate: true });
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
                <form
                    // Desktop: lock to the viewport so the resizable panels
                    // inside the body can fill the remaining space below the
                    // sticky header. Mobile: allow the page to grow with its
                    // content (single column, vertical scroll).
                    className={isDesktop ? 'flex h-[100dvh] min-h-0 w-full flex-col' : 'flex min-h-[100dvh] flex-col'}
                    onSubmit={handleSubmit(onSubmitWithGuard)}
                >
                    <KnowledgeHeader
                        canAnonymize={canAnonymize}
                        isAnonymizeDisabled={isAnonymizeDisabled}
                        isAnonymizing={isAnonymizing}
                        isNew={isNew}
                        knowledge={knowledge}
                        onAnonymize={handleAnonymize}
                        onBeforeNavigateAway={() => skipNextBlockRef.current()}
                        saveButton={saveButton}
                    />
                    {isDesktop ? (
                        <KnowledgeFormLayoutDesktop
                            control={control}
                            isNew={isNew}
                            isSaving={isSaving}
                            knowledge={knowledge}
                        />
                    ) : (
                        <KnowledgeFormLayoutMobile
                            control={control}
                            isNew={isNew}
                            isSaving={isSaving}
                            knowledge={knowledge}
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
