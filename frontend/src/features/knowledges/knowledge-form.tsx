import { zodResolver } from '@hookform/resolvers/zod';
import { Save } from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { type FieldPath, type SubmitHandler, useForm } from 'react-hook-form';
import { useNavigate } from 'react-router-dom';
import { z } from 'zod';

import type {
    CreateKnowledgeDocumentInput,
    KnowledgeDocumentFragmentFragment,
    UpdateKnowledgeDocumentInput,
} from '@/graphql/types';

import { UnsavedChangesDialog } from '@/components/shared/unsaved-changes-dialog';
import { Button } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { Spinner } from '@/components/ui/spinner';
import { KnowledgeAnswerType, KnowledgeDocType, KnowledgeGuideType } from '@/graphql/types';
import { useBreakpoint } from '@/hooks/use-breakpoint';
import { useUnsavedChangesGuard } from '@/hooks/use-unsaved-changes-guard';
import { Log } from '@/lib/log';

import { KnowledgeFormLayoutDesktop, KnowledgeFormLayoutMobile } from './knowledge-form-layout';
import { KnowledgeHeader } from './knowledge-header';

// ---------------------------------------------------------------------------
// Schema, types, pure helpers
// ---------------------------------------------------------------------------

export const formSchema = z
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

// `description` and `codeLang` are optional and we want empty strings to map
// to `undefined` so the backend treats them as "absent" instead of "set to ''".
const trimmedOrUndefined = (value: null | string | undefined): string | undefined => {
    if (!value) {
        return undefined;
    }

    const trimmed = value.trim();

    return trimmed.length === 0 ? undefined : trimmed;
};

// Shared field projection used by both create and update payloads. The only
// difference between the two GraphQL inputs is that `docType` is required on
// create and immutable on update — see `formValuesTo{Create,Update}Input`
// below.
const formValuesToBasePayload = (values: FormValues) => ({
    answerType: values.docType === KnowledgeDocType.Answer ? values.answerType : undefined,
    codeLang: values.docType === KnowledgeDocType.Code ? trimmedOrUndefined(values.codeLang) : undefined,
    content: values.content,
    description: trimmedOrUndefined(values.description),
    guideType: values.docType === KnowledgeDocType.Guide ? values.guideType : undefined,
    question: values.question,
});

export const formValuesToCreateInput = (values: FormValues): CreateKnowledgeDocumentInput => ({
    ...formValuesToBasePayload(values),
    docType: values.docType,
});

export const formValuesToUpdateInput = (values: FormValues): UpdateKnowledgeDocumentInput =>
    formValuesToBasePayload(values);

// ---------------------------------------------------------------------------
// Form component
// ---------------------------------------------------------------------------

export interface SubmitResult {
    document?: KnowledgeDocumentFragmentFragment;
    redirectTo?: string;
}

interface KnowledgeFormProps {
    initialValues: FormValues;
    isNew: boolean;
    knowledge?: KnowledgeDocumentFragmentFragment | null;
    knowledgeName: null | string;
    onSubmit: (values: FormValues) => Promise<SubmitResult>;
}

export const KnowledgeForm = ({ initialValues, isNew, knowledge, knowledgeName, onSubmit }: KnowledgeFormProps) => {
    const navigate = useNavigate();
    const { isDesktop } = useBreakpoint();
    const [isSaving, setIsSaving] = useState(false);

    const form = useForm<FormValues>({
        defaultValues: initialValues,
        // `onTouched` validates a field on its first blur and on every change
        // afterwards. With `onChange` we'd run the entire Zod schema on every
        // keystroke (including every emit from the multi-kilobyte `content`
        // markdown editor) — same UX after the first interaction, no waste
        // on initial mount or untouched fields.
        mode: 'onTouched',
        resolver: zodResolver(formSchema),
    });

    const { control, formState, handleSubmit, reset } = form;
    const { isDirty, isValid } = formState;

    const performSave = useCallback(
        async (values: FormValues): Promise<boolean> => {
            try {
                const result = await onSubmit(values);

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
            {isSaving ? <Spinner variant="circle" /> : <Save aria-hidden="true" />}
            {isNew ? 'Create' : 'Save'}
        </Button>
    );

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
                        isNew={isNew}
                        knowledgeName={knowledgeName}
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
};
