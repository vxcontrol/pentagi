import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';

import type { SubtaskFragmentFragment, TaskFragmentFragment } from '@/graphql/types';

import { TooltipProvider } from '@/components/ui/tooltip';
import { StatusType } from '@/graphql/types';

import FlowTask from './flow-task';

const subtask = (overrides: Partial<SubtaskFragmentFragment> = {}): SubtaskFragmentFragment => ({
    createdAt: '2026-09-22T10:00:00Z',
    description: 'nmap -sV',
    id: '9',
    result: '',
    status: StatusType.Finished,
    taskId: '5',
    title: 'Scan ports',
    updatedAt: '2026-09-22T10:00:00Z',
    ...overrides,
});

const task = (overrides: Partial<TaskFragmentFragment> = {}): TaskFragmentFragment => ({
    createdAt: '2026-09-22T10:00:00Z',
    flowId: '7',
    id: '5',
    input: 'scan 10.0.0.5',
    result: '',
    status: StatusType.Running,
    subtasks: [],
    title: 'Scan the host',
    updatedAt: '2026-09-22T10:00:00Z',
    ...overrides,
});

const renderTask = (value: TaskFragmentFragment) =>
    render(
        <TooltipProvider delayDuration={0}>
            <FlowTask task={value} />
        </TooltipProvider>,
    );

describe('FlowTask', () => {
    it('does not say a failed task waits for subtasks', () => {
        renderTask(task({ status: StatusType.Failed }));

        expect(screen.queryByText('Waiting for subtasks to be created...')).not.toBeInTheDocument();
        expect(screen.getByText('The task failed before any subtasks were created.')).toBeInTheDocument();
    });

    it('still says a running task waits for its subtasks', () => {
        renderTask(task({ status: StatusType.Running }));

        expect(screen.getByText('Waiting for subtasks to be created...')).toBeInTheDocument();
    });

    it('names the status icons of the task and its subtasks', () => {
        renderTask(task({ status: StatusType.Failed, subtasks: [subtask({ status: StatusType.Finished })] }));

        expect(screen.getByRole('img', { name: 'Failed' })).toBeInTheDocument();
        expect(screen.getByRole('img', { name: 'Finished' })).toBeInTheDocument();
    });

    it('shows why a failed task stopped', async () => {
        const user = userEvent.setup({ delay: null });
        renderTask(task({ result: 'failed to generate subtasks: status code 404', status: StatusType.Failed }));

        await user.click(screen.getByText('Show details'));

        expect(screen.getByText(/status code 404/)).toBeInTheDocument();
    });
});
