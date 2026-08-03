import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { type Control, Controller } from 'react-hook-form';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { KnowledgeDocumentFragmentFragment } from '@/graphql/types';

import { KnowledgeAnswerType, KnowledgeDocType } from '@/graphql/types';

import type { FormValues, SubmitResult } from './knowledge-form';

const { navigate } = vi.hoisted(() => ({ navigate: vi.fn() }));

// `useBlocker` throws outside a data router; the guard only needs a stable
// inert blocker for these save/navigation tests.
vi.mock('react-router-dom', async (importOriginal) => {
    const actual = await importOriginal<typeof import('react-router-dom')>();

    return {
        ...actual,
        useBlocker: () => ({ proceed: undefined, reset: undefined, state: 'unblocked' }),
        useNavigate: () => navigate,
    };
});

vi.mock('@/hooks/use-breakpoint', () => ({
    useBreakpoint: () => ({ isDesktop: true, isMobile: false }),
}));

vi.mock('@apollo/client/react', () => ({ useMutation: () => [vi.fn(), {}] }));

vi.mock('@/providers/user-provider', () => ({
    useUser: () => ({ authInfo: { privileges: ['anonymize.call'] } }),
}));

vi.mock('sonner', () => ({ toast: { error: vi.fn(), info: vi.fn(), success: vi.fn() } }));

vi.mock('./knowledge-form-layout', () => {
    const Layout = ({ control }: { control: Control<FormValues> }) => (
        <>
            <Controller
                control={control}
                name="content"
                render={({ field }) => (
                    <textarea
                        aria-label="content"
                        {...field}
                    />
                )}
            />
            <Controller
                control={control}
                name="question"
                render={({ field }) => (
                    <input
                        aria-label="question"
                        {...field}
                    />
                )}
            />
        </>
    );

    return { KnowledgeFormLayoutDesktop: Layout, KnowledgeFormLayoutMobile: Layout };
});

vi.mock('./knowledge-header', () => ({
    KnowledgeHeader: ({
        isAnonymizeDisabled,
        saveButton,
    }: {
        isAnonymizeDisabled?: boolean;
        saveButton?: React.ReactNode;
    }) => (
        <div>
            {saveButton}
            <button
                disabled={isAnonymizeDisabled}
                type="button"
            >
                Anonymize
            </button>
        </div>
    ),
}));

import { KnowledgeForm } from './knowledge-form';

const newValues: FormValues = {
    answerType: KnowledgeAnswerType.Other,
    codeLang: '',
    content: '',
    description: '',
    docType: KnowledgeDocType.Answer,
    guideType: undefined,
    question: '',
};

const editValues: FormValues = {
    answerType: KnowledgeAnswerType.Other,
    codeLang: '',
    content: 'existing content',
    description: '',
    docType: KnowledgeDocType.Answer,
    guideType: undefined,
    question: 'existing question',
};

beforeEach(() => {
    navigate.mockClear();
});

describe('KnowledgeForm — create', () => {
    it('submits values and navigates to the redirect target', async () => {
        const user = userEvent.setup();
        const onSubmit = vi.fn<(values: FormValues, dirty: unknown) => Promise<SubmitResult>>().mockResolvedValue({
            redirectTo: '/knowledges/123',
        });

        render(
            <KnowledgeForm
                initialValues={newValues}
                isNew
                onSubmit={onSubmit}
            />,
        );

        await user.type(screen.getByLabelText('content'), 'hello world');
        await user.type(screen.getByLabelText('question'), 'why');
        await user.click(screen.getByRole('button', { name: 'Create' }));

        await waitFor(() => expect(onSubmit).toHaveBeenCalledOnce());
        const [values, dirty] = onSubmit.mock.lastCall!;
        expect(values.content).toBe('hello world');
        expect(values.question).toBe('why');
        expect(dirty).toMatchObject({ content: true, question: true });

        await waitFor(() => expect(navigate).toHaveBeenCalledWith('/knowledges/123'));
    });
});

describe('KnowledgeForm — update', () => {
    it('submits the edited field and does NOT navigate without a redirect', async () => {
        const user = userEvent.setup();
        const onSubmit = vi
            .fn<(values: FormValues, dirty: unknown) => Promise<SubmitResult>>()
            .mockResolvedValue({ document: undefined });

        render(
            <KnowledgeForm
                initialValues={editValues}
                isNew={false}
                onSubmit={onSubmit}
            />,
        );

        const question = screen.getByLabelText('question');
        await user.clear(question);
        await user.type(question, 'updated question');
        await user.click(screen.getByRole('button', { name: 'Save' }));

        await waitFor(() => expect(onSubmit).toHaveBeenCalledOnce());
        const [values, dirty] = onSubmit.mock.lastCall!;
        expect(values.question).toBe('updated question');
        expect(dirty).toMatchObject({ question: true });

        // Save re-disables once the post-save reset clears isDirty — a deterministic
        // anchor for the negative navigate assertion below.
        await waitFor(() => expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled());
        expect(navigate).not.toHaveBeenCalled();
    });

    it('resets untouched fields to the server document returned by onSubmit', async () => {
        const user = userEvent.setup();
        // Server normalizes `content` (a field the user never edits here). The form's
        // resetOptions keep DIRTY fields at their typed value, so asserting the reset
        // applied the server document means asserting on an untouched field.
        const serverDocument = {
            answerType: KnowledgeAnswerType.Other,
            codeLang: null,
            content: 'server normalized content',
            description: null,
            docType: KnowledgeDocType.Answer,
            flowId: null,
            guideType: null,
            id: 'doc-1',
            manual: false,
            partSize: 0,
            question: 'updated question',
            subtaskId: null,
            taskId: null,
            totalSize: 0,
            userId: 'u-1',
        } satisfies KnowledgeDocumentFragmentFragment;

        const onSubmit = vi
            .fn<(values: FormValues, dirty: unknown) => Promise<SubmitResult>>()
            .mockResolvedValue({ document: serverDocument });

        render(
            <KnowledgeForm
                initialValues={editValues}
                isNew={false}
                onSubmit={onSubmit}
            />,
        );

        const question = screen.getByLabelText('question');
        await user.clear(question);
        await user.type(question, 'updated question');
        await user.click(screen.getByRole('button', { name: 'Save' }));

        await waitFor(() => expect(onSubmit).toHaveBeenCalledOnce());

        // editValues.content is 'existing content'; reaching the server value proves the
        // form took `documentToFormValues(result.document)`, not the local `values` fallback.
        await waitFor(() => expect(screen.getByLabelText('content')).toHaveValue('server normalized content'));
        expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled();
    });

    it('keeps Save disabled until the form is dirty', async () => {
        const user = userEvent.setup();
        const onSubmit = vi.fn().mockResolvedValue({});

        render(
            <KnowledgeForm
                initialValues={editValues}
                isNew={false}
                onSubmit={onSubmit}
            />,
        );

        expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled();

        await user.type(screen.getByLabelText('question'), '!');
        await waitFor(() => expect(screen.getByRole('button', { name: 'Save' })).toBeEnabled());
    });
});

describe('KnowledgeForm — anonymize wiring', () => {
    it('disables the anonymize control until content is non-empty', async () => {
        const user = userEvent.setup();

        render(
            <KnowledgeForm
                initialValues={newValues}
                isNew
                onSubmit={vi.fn().mockResolvedValue({})}
            />,
        );

        expect(screen.getByRole('button', { name: 'Anonymize' })).toBeDisabled();

        await user.type(screen.getByLabelText('content'), 'secret data');
        await waitFor(() => expect(screen.getByRole('button', { name: 'Anonymize' })).toBeEnabled());
    });
});
