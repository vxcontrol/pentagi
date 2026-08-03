import { gql, InMemoryCache, Observable } from '@apollo/client';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { createCache, createStreamingLink, updateCacheForSubscription } from './apollo';

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

const makeCache = createCache;

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

type EmittedData = {
    [key: string]: unknown;
    assistantLogUpdated?: { appendPart?: boolean; message?: null | string };
};

describe('streaming assistant-log link (createStreamingLink)', () => {
    let now = 0;

    beforeEach(() => {
        now = 1000;
        vi.spyOn(Date, 'now').mockImplementation(() => now);
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    const drive = () => {
        const link = createStreamingLink();
        let source: undefined | { next: (value: unknown) => void };
        const forward = () =>
            new Observable((observer) => {
                source = observer;

                return () => {};
            });
        const emitted: EmittedData[] = [];

        link.request({} as never, forward as never)!.subscribe({
            next: (result) => emitted.push(result.data as EmittedData),
        });

        return {
            emitted,
            pushLog: (log: Record<string, unknown>) => source?.next({ data: { assistantLogUpdated: log } }),
            pushRaw: (data: unknown) => source?.next({ data }),
        };
    };

    it('coalesces rapid append parts but accumulates the full running message', () => {
        const { emitted, pushLog } = drive();

        pushLog({ appendPart: true, id: 'a', message: 'Hello', result: null, thinking: null });
        expect(emitted).toHaveLength(1);
        expect(emitted[0]?.assistantLogUpdated?.message).toBe('Hello');
        expect(emitted[0]?.assistantLogUpdated?.appendPart).toBe(false);

        // within the 50ms throttle window -> accumulated internally, not emitted
        pushLog({ appendPart: true, id: 'a', message: ' World', result: null, thinking: null });
        expect(emitted).toHaveLength(1);

        // past the window -> emits the full running total, not just the latest delta
        now = 1060;
        pushLog({ appendPart: true, id: 'a', message: '!', result: null, thinking: null });
        expect(emitted).toHaveLength(2);
        expect(emitted[1]?.assistantLogUpdated?.message).toBe('Hello World!');
    });

    it('flushes the accumulation on the final part, then starts fresh for the same id', () => {
        const { emitted, pushLog } = drive();

        pushLog({ appendPart: true, id: 'a', message: 'Hello', result: null, thinking: null });
        now = 1010;
        pushLog({ appendPart: true, id: 'a', message: ' World', result: null, thinking: null }); // throttled

        // the final (non-append) part concatenates the cached total + its own delta and always emits
        pushLog({ appendPart: false, id: 'a', message: '!', result: null, thinking: null });
        expect(emitted.at(-1)?.assistantLogUpdated?.message).toBe('Hello World!');

        // its cache entry was deleted -> a fresh append for the same id starts empty
        now = 2000;
        pushLog({ appendPart: true, id: 'a', message: 'Next', result: null, thinking: null });
        expect(emitted.at(-1)?.assistantLogUpdated?.message).toBe('Next');
    });

    it('passes a non-assistant-log result straight through', () => {
        const { emitted, pushRaw } = drive();

        pushRaw({ somethingElse: { id: '1' } });

        expect(emitted).toEqual([{ somethingElse: { id: '1' } }]);
    });
});

describe('ProviderConfig cache identity', () => {
    const PROVIDERS = gql`
        query P {
            settingsProviders {
                default {
                    anthropic {
                        id
                        name
                        type
                    }
                    openai {
                        id
                        name
                        type
                    }
                }
            }
        }
    `;

    // The resolver builds default providers without an ID, so every one of them arrives as id 0.
    // Normalising on that shared id would collapse them onto a single cache entry and the last
    // write would win — the create form then seeds one provider type from another's defaults.
    it('keeps two defaults apart even though both carry id 0', () => {
        const cache = createCache();

        cache.writeQuery({
            data: {
                settingsProviders: {
                    __typename: 'ProvidersConfig',
                    default: {
                        __typename: 'DefaultProvidersConfig',
                        anthropic: { __typename: 'ProviderConfig', id: 0, name: 'anthropic', type: 'anthropic' },
                        openai: { __typename: 'ProviderConfig', id: 0, name: 'openai', type: 'openai' },
                    },
                },
            },
            query: PROVIDERS,
        });

        const read = cache.readQuery<{
            settingsProviders: { default: { anthropic: { name: string }; openai: { name: string } } };
        }>({ query: PROVIDERS });

        expect(read?.settingsProviders.default.anthropic.name).toBe('anthropic');
        expect(read?.settingsProviders.default.openai.name).toBe('openai');
    });
});
