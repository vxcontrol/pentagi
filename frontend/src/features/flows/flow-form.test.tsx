import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { setSelectedProvider, uploadParams } = vi.hoisted(() => ({
    setSelectedProvider: vi.fn(),
    uploadParams: { current: null } as { current: null | { onSuccess: (uploaded: { items: unknown[] }) => void } },
}));

vi.mock('@/providers/providers-provider', () => ({
    useProviders: () => ({
        providers: [{ name: 'openai', type: 'openai' }],
        setSelectedProvider,
    }),
}));

vi.mock('@/providers/templates-provider', () => ({
    useTemplates: () => ({ templates: [] }),
}));

vi.mock('@/providers/resources-provider', () => ({
    useResources: () => ({
        resources: [
            {
                createdAt: '',
                id: '7',
                isDir: false,
                name: 'scope.txt',
                path: 'scope.txt',
                size: 12,
                updatedAt: '',
                userId: '1',
            },
        ],
    }),
}));

vi.mock('@/features/resources/use-resources-upload', () => ({
    useResourcesUpload: (params: { onSuccess: (uploaded: { items: unknown[] }) => void }) => {
        uploadParams.current = params;

        return { isUploading: false, uploadFiles: vi.fn() };
    },
}));

const { FlowForm } = await import('./flow-form');

const attachResource = async () => {
    await waitFor(() => expect(uploadParams.current).not.toBeNull());
    uploadParams.current!.onSuccess({ items: [{ id: 7 }] });
    await screen.findByRole('button', { name: 'Remove scope.txt' });
};

beforeEach(() => {
    uploadParams.current = null;
    setSelectedProvider.mockClear();
});

describe('FlowForm attachments', () => {
    it('clears the attached resources after a submit', async () => {
        const user = userEvent.setup({ delay: null });
        const onSubmit = vi.fn();

        render(
            <FlowForm
                defaultValues={{ providerName: 'openai' }}
                onSubmit={onSubmit}
                type="assistant"
            />,
        );

        await attachResource();

        await user.type(screen.getByRole('textbox'), 'scan it');
        await user.click(screen.getByRole('button', { name: 'Submit' }));

        await waitFor(() => expect(onSubmit).toHaveBeenCalledOnce());
        expect(onSubmit.mock.calls[0]?.[0]).toMatchObject({ message: 'scan it', resourceIds: ['7'] });

        await waitFor(() => expect(screen.queryByRole('button', { name: 'Remove scope.txt' })).not.toBeInTheDocument());
    });

    it('does not resend the previous attachments on the next submit', async () => {
        const user = userEvent.setup({ delay: null });
        const onSubmit = vi.fn();

        render(
            <FlowForm
                defaultValues={{ providerName: 'openai' }}
                onSubmit={onSubmit}
                type="assistant"
            />,
        );

        await attachResource();

        await user.type(screen.getByRole('textbox'), 'first');
        await user.click(screen.getByRole('button', { name: 'Submit' }));
        await waitFor(() => expect(onSubmit).toHaveBeenCalledOnce());

        await user.type(screen.getByRole('textbox'), 'second');
        await user.click(screen.getByRole('button', { name: 'Submit' }));

        await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(2));
        expect(onSubmit.mock.calls[1]?.[0]).toMatchObject({ message: 'second', resourceIds: [] });
    });

    it('adopts resourceIds that arrive through defaultValues after mount', async () => {
        const { rerender } = render(
            <FlowForm
                defaultValues={{ providerName: 'openai' }}
                onSubmit={vi.fn()}
                type="assistant"
            />,
        );

        rerender(
            <FlowForm
                defaultValues={{ providerName: 'openai', resourceIds: ['7'] }}
                onSubmit={vi.fn()}
                type="assistant"
            />,
        );

        expect(await screen.findByRole('button', { name: 'Remove scope.txt' })).toBeInTheDocument();
    });
});
