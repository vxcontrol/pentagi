import { gql, InMemoryCache } from '@apollo/client';
import { beforeEach, describe, expect, it } from 'vitest';

import { updateCacheForSubscription } from './apollo';

const TERMINAL = gql`
    query T($flowId: ID!) {
        terminalLogs(flowId: $flowId) {
            id
            text
        }
    }
`;

const TASKS = gql`
    query Ts($flowId: ID!) {
        tasks(flowId: $flowId) {
            id
            title
            status
        }
    }
`;

const FLOWS = gql`
    query F {
        flows {
            id
            title
        }
    }
`;

// Mirrors the flow-scoped list fields in the production cache: keyed by flowId,
// query writes replace wholesale (the subscription link mutates via cache.modify).
const makeCache = () =>
    new InMemoryCache({
        typePolicies: {
            Query: {
                fields: {
                    flows: { merge: (_existing, incoming) => incoming },
                    tasks: { keyArgs: ['flowId'], merge: (_existing, incoming) => incoming },
                    terminalLogs: { keyArgs: ['flowId'], merge: (_existing, incoming) => incoming },
                },
            },
        },
    });

// The subscription payload carries `__typename` + fields; updateCacheForSubscription's
// param is the minimal `{ id }` contract, so route literals through this to keep
// TypeScript's excess-property check off fresh object literals.
const frame = <T extends { id: number | string }>(payload: T): T => payload;

const termIds = (cache: InMemoryCache, flowId: string) =>
    cache
        .readQuery<{ terminalLogs: { id: string }[] }>({ query: TERMINAL, variables: { flowId } })
        ?.terminalLogs.map((log) => String(log.id));

const taskList = (cache: InMemoryCache, flowId: string) =>
    cache.readQuery<{ tasks: { id: string; status: string }[] }>({ query: TASKS, variables: { flowId } })?.tasks;

describe('subscription cache merge-link (updateCacheForSubscription)', () => {
    let cache: InMemoryCache;

    beforeEach(() => {
        cache = makeCache();
    });

    it('appends an added log and de-dups a repeat of the same id', () => {
        cache.writeQuery({
            data: { terminalLogs: [{ __typename: 'TerminalLog', id: '1', text: 'first' }] },
            query: TERMINAL,
            variables: { flowId: '1' },
        });

        updateCacheForSubscription(
            cache,
            'terminalLogAdded',
            'terminalLogs',
            frame({ __typename: 'TerminalLog', id: '2', text: 'second' }),
            { flowId: '1' },
        );
        expect(termIds(cache, '1')).toEqual(['1', '2']);

        updateCacheForSubscription(
            cache,
            'terminalLogAdded',
            'terminalLogs',
            frame({ __typename: 'TerminalLog', id: '2', text: 'second-again' }),
            { flowId: '1' },
        );
        expect(termIds(cache, '1')).toEqual(['1', '2']);
    });

    it('prepends a created task (newest first), not sorted by id', () => {
        cache.writeQuery({
            data: { tasks: [{ __typename: 'Task', id: '5', status: 'finished', title: 'old' }] },
            query: TASKS,
            variables: { flowId: '1' },
        });

        updateCacheForSubscription(
            cache,
            'taskCreated',
            'tasks',
            frame({ __typename: 'Task', id: '2', status: 'running', title: 'new' }),
            { flowId: '1' },
        );

        expect(taskList(cache, '1')?.map((task) => task.id)).toEqual(['2', '5']);
    });

    it('removes a deleted flow by id', () => {
        cache.writeQuery({
            data: {
                flows: [
                    { __typename: 'Flow', id: '1', title: 'a' },
                    { __typename: 'Flow', id: '2', title: 'b' },
                ],
            },
            query: FLOWS,
        });

        updateCacheForSubscription(cache, 'flowDeleted', 'flows', frame({ __typename: 'Flow', id: '1' }));

        expect(cache.readQuery<{ flows: { id: string }[] }>({ query: FLOWS })?.flows.map((flow) => flow.id)).toEqual([
            '2',
        ]);
    });

    it('merges fields of an updated task in place without reordering', () => {
        cache.writeQuery({
            data: {
                tasks: [
                    { __typename: 'Task', id: '1', status: 'running', title: 'a' },
                    { __typename: 'Task', id: '2', status: 'running', title: 'b' },
                ],
            },
            query: TASKS,
            variables: { flowId: '1' },
        });

        updateCacheForSubscription(
            cache,
            'taskUpdated',
            'tasks',
            frame({ __typename: 'Task', id: '1', status: 'finished', title: 'a' }),
            { flowId: '1' },
        );

        const tasks = taskList(cache, '1');
        expect(tasks?.map((task) => task.id)).toEqual(['1', '2']);
        expect(tasks?.find((task) => task.id === '1')?.status).toBe('finished');
    });

    it('appends an updated task that is not yet in the cache', () => {
        cache.writeQuery({
            data: { tasks: [{ __typename: 'Task', id: '1', status: 'running', title: 'a' }] },
            query: TASKS,
            variables: { flowId: '1' },
        });

        updateCacheForSubscription(
            cache,
            'taskUpdated',
            'tasks',
            frame({ __typename: 'Task', id: '9', status: 'running', title: 'z' }),
            { flowId: '1' },
        );

        expect(taskList(cache, '1')?.map((task) => task.id)).toEqual(['1', '9']);
    });

    it('isolates the update to the matching flowId variant', () => {
        cache.writeQuery({
            data: { terminalLogs: [{ __typename: 'TerminalLog', id: '1', text: 'f1' }] },
            query: TERMINAL,
            variables: { flowId: '1' },
        });
        cache.writeQuery({
            data: { terminalLogs: [{ __typename: 'TerminalLog', id: '10', text: 'f2' }] },
            query: TERMINAL,
            variables: { flowId: '2' },
        });

        updateCacheForSubscription(
            cache,
            'terminalLogAdded',
            'terminalLogs',
            frame({ __typename: 'TerminalLog', id: '11', text: 'f1-new' }),
            { flowId: '1' },
        );

        expect(termIds(cache, '1')).toEqual(['1', '11']);
        expect(termIds(cache, '2')).toEqual(['10']);
    });

    it('de-dups across string vs number ids (REST hydration writes strings)', () => {
        cache.writeQuery({
            data: { terminalLogs: [{ __typename: 'TerminalLog', id: '5', text: 'x' }] },
            query: TERMINAL,
            variables: { flowId: '1' },
        });

        updateCacheForSubscription(
            cache,
            'terminalLogAdded',
            'terminalLogs',
            frame({ __typename: 'TerminalLog', id: 5, text: 'x-again' }),
            { flowId: '1' },
        );

        expect(termIds(cache, '1')).toEqual(['5']);
    });
});
