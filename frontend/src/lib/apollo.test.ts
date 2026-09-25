import type { DocumentNode, FieldNode, OperationDefinitionNode, SelectionSetNode } from 'graphql';

import { ApolloClient, ApolloLink, gql, InMemoryCache, Observable } from '@apollo/client';
import { CombinedGraphQLErrors } from '@apollo/client/errors';
import { parse, print } from 'graphql';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import {
    ApiTokenCreatedDocument,
    ApiTokenDeletedDocument,
    ApiTokenUpdatedDocument,
    AssistantLogFragmentFragmentDoc,
} from '@/graphql/types';

import operationsDocument from '../../graphql-schema.graphql?raw';
import {
    APPEND_ONLY_LIST_FIELDS,
    createCache,
    createSocketPresence,
    createStreamingLink,
    createSubscriptionCacheLink,
    lastLocalWriteAt,
    queryKeyArgs,
    subscriptionToCacheFieldMap,
    updateCacheForSubscription,
    watchQueryDefaults,
} from './apollo';
import { listOrder } from './list-order';

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

const FLOW = gql`
    query Fl($id: ID!) {
        flow(flowId: $id) {
            id
            title
        }
    }
`;

const FLOW_FILES = gql`
    query FF($flowId: ID!) {
        flowFiles(flowId: $flowId) {
            id
            flowId
            name
            size
        }
    }
`;

const ASSISTANT_LOGS = gql`
    query AL($flowId: ID!, $assistantId: ID!) {
        assistantLogs(flowId: $flowId, assistantId: $assistantId) {
            id
            assistantId
        }
    }
`;

const ASSISTANT_LOGS_WITH_RESULT = gql`
    query ALR($flowId: ID!, $assistantId: ID!) {
        assistantLogs(flowId: $flowId, assistantId: $assistantId) {
            id
            message
            result
            resultFormat
        }
    }
`;

const RESOURCES = gql`
    query R($path: String, $recursive: Boolean) {
        resources(path: $path, recursive: $recursive) {
            id
            name
            path
        }
    }
`;

const MESSAGE_LOGS = gql`
    query ML($flowId: ID!) {
        messageLogs(flowId: $flowId) {
            id
            message
            result
            resultFormat
        }
    }
`;

const KNOWLEDGE = gql`
    query K($withContent: Boolean!) {
        knowledgeDocuments(withContent: $withContent) {
            id
            content
        }
    }
`;

const SETTINGS_USER = gql`
    query SU {
        settingsUser {
            id
            favoriteFlows
        }
    }
`;

const API_TOKENS = gql`
    query AT {
        apiTokens {
            id
            tokenId
            name
        }
    }
`;

const TOOL_CALL_LOGS = gql`
    query TCL($flowId: ID!) {
        toolCallLogs(flowId: $flowId) {
            id
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

const assistantLogIds = (cache: InMemoryCache, flowId: string, assistantId: string) =>
    cache
        .readQuery<{ assistantLogs: { id: string }[] }>({
            query: ASSISTANT_LOGS,
            variables: { assistantId, flowId },
        })
        ?.assistantLogs.map((log) => String(log.id));

const resourceIds = (cache: InMemoryCache, path: string, recursive: boolean) =>
    cache
        .readQuery<{ resources: { id: string }[] }>({ query: RESOURCES, variables: { path, recursive } })
        ?.resources.map((resource) => String(resource.id));

const toolCallIds = (cache: InMemoryCache, flowId: string) =>
    cache
        .readQuery<{ toolCallLogs: { id: string }[] }>({ query: TOOL_CALL_LOGS, variables: { flowId } })
        ?.toolCallLogs.map((log) => String(log.id));

const apiTokenIds = (cache: InMemoryCache) =>
    cache.readQuery<{ apiTokens: { id: string }[] }>({ query: API_TOKENS })?.apiTokens.map((token) => String(token.id));

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

    it('puts a created task where a refetch would, not where its name suggests', () => {
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

        expect(taskList(cache, '1')?.map((task) => task.id)).toEqual(['5', '2']);
    });

    it('puts a created flow where the newest-first flows query puts it', () => {
        cache.writeQuery({
            data: { flows: [{ __typename: 'Flow', id: '5', title: 'old' }] },
            query: FLOWS,
        });

        updateCacheForSubscription(cache, 'flowCreated', 'flows', frame({ __typename: 'Flow', id: '2', title: 'new' }));

        expect(cache.readQuery<{ flows: { id: string }[] }>({ query: FLOWS })?.flows.map((flow) => flow.id)).toEqual([
            '2',
            '5',
        ]);
    });

    it('declares an order for every collection a subscription writes to, and no others', () => {
        const written = new Set(Object.values(subscriptionToCacheFieldMap));

        written.delete('settingsUser');

        expect([...written].filter((field) => !Object.hasOwn(listOrder, field)).sort()).toEqual([]);
        expect(
            Object.keys(listOrder)
                .filter((field) => !written.has(field))
                .sort(),
        ).toEqual([]);
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

    it('stops serving a deleted flow to the page that still has it open', () => {
        cache.writeQuery({
            data: { flow: { __typename: 'Flow', id: '1', title: 'a' } },
            query: FLOW,
            variables: { id: '1' },
        });

        updateCacheForSubscription(cache, 'flowDeleted', 'flows', frame({ __typename: 'Flow', id: '1' }));

        expect(cache.readQuery({ query: FLOW, variables: { id: '1' } })).toBeNull();
    });

    it('keeps two flows apart when they hold a file at the same relative path', () => {
        const id = 'md5-of-report-txt';
        const inFlow = (flowId: string, size: number) => ({
            __typename: 'FlowFile',
            flowId,
            id,
            name: 'report.txt',
            size,
        });

        cache.writeQuery({ data: { flowFiles: [inFlow('1', 10)] }, query: FLOW_FILES, variables: { flowId: '1' } });
        cache.writeQuery({ data: { flowFiles: [inFlow('2', 999)] }, query: FLOW_FILES, variables: { flowId: '2' } });

        const flowOne = cache.readQuery<{ flowFiles: { size: number }[] }>({
            query: FLOW_FILES,
            variables: { flowId: '1' },
        });

        expect(flowOne?.flowFiles[0]?.size).toBe(10);
    });

    it('still adds and removes a file, which a non-normalized FlowFile would not', () => {
        const file = { __typename: 'FlowFile', flowId: '1', id: 'a', name: 'report.txt', size: 1 };

        cache.writeQuery({ data: { flowFiles: [] }, query: FLOW_FILES, variables: { flowId: '1' } });

        updateCacheForSubscription(cache, 'flowFileAdded', 'flowFiles', frame(file), { flowId: '1' });
        expect(cache.readQuery({ query: FLOW_FILES, variables: { flowId: '1' } })).toEqual({ flowFiles: [file] });

        updateCacheForSubscription(cache, 'flowFileDeleted', 'flowFiles', frame(file), { flowId: '1' });
        expect(cache.readQuery({ query: FLOW_FILES, variables: { flowId: '1' } })).toEqual({ flowFiles: [] });
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

    it('isolates the update on a field whose variants are keyed without a keyArgs policy', () => {
        cache.writeQuery({
            data: { toolCallLogs: [{ __typename: 'ToolCallLog', id: '1' }] },
            query: TOOL_CALL_LOGS,
            variables: { flowId: '1' },
        });
        cache.writeQuery({
            data: { toolCallLogs: [{ __typename: 'ToolCallLog', id: '9' }] },
            query: TOOL_CALL_LOGS,
            variables: { flowId: '2' },
        });

        updateCacheForSubscription(
            cache,
            'toolCallLogAdded',
            'toolCallLogs',
            frame({ __typename: 'ToolCallLog', id: '2' }),
            { flowId: '1' },
        );

        expect(toolCallIds(cache, '1')).toEqual(['1', '2']);
        expect(toolCallIds(cache, '2'), 'flow 2 must not receive flow 1 delta').toEqual(['9']);
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

    it('keeps assistantId in the assistant-log fragment', () => {
        expect(print(AssistantLogFragmentFragmentDoc)).toContain('assistantId');
    });

    it('routes an assistant log to its own assistantId variant', () => {
        for (const assistantId of ['A', 'B']) {
            cache.writeQuery({
                data: { assistantLogs: [{ __typename: 'AssistantLog', assistantId, id: `seed-${assistantId}` }] },
                query: ASSISTANT_LOGS,
                variables: { assistantId, flowId: '1' },
            });
        }

        updateCacheForSubscription(
            cache,
            'assistantLogAdded',
            'assistantLogs',
            frame({ __typename: 'AssistantLog', assistantId: 'B', id: 'new' }),
            { flowId: '1' },
        );

        expect(assistantLogIds(cache, '1', 'B')).toEqual(['seed-B', 'new']);
        expect(assistantLogIds(cache, '1', 'A')).toEqual(['seed-A']);
    });

    it('adds and removes an api token without a refetch', () => {
        cache.writeQuery({
            data: { apiTokens: [{ __typename: 'APIToken', id: '1', name: 'first', tokenId: 't1' }] },
            query: API_TOKENS,
        });

        updateCacheForSubscription(
            cache,
            'apiTokenCreated',
            'apiTokens',
            frame({ __typename: 'APIToken', id: '2', name: 'second', tokenId: 't2' }),
            {},
        );
        expect(apiTokenIds(cache)).toEqual(['2', '1']);

        updateCacheForSubscription(
            cache,
            'apiTokenDeleted',
            'apiTokens',
            frame({ __typename: 'APIToken', id: '1', name: 'first', tokenId: 't1' }),
            {},
        );
        expect(apiTokenIds(cache)).toEqual(['2']);
    });

    it('keeps a nested resource in the parent path variant', () => {
        cache.writeQuery({
            data: { resources: [{ __typename: 'UserResource', id: '1', name: 'a.txt', path: '/a.txt' }] },
            query: RESOURCES,
            variables: { path: '/', recursive: true },
        });

        updateCacheForSubscription(
            cache,
            'resourceAdded',
            'resources',
            frame({ __typename: 'UserResource', id: '2', name: 'b.txt', path: '/docs/b.txt' }),
            {},
        );

        expect(resourceIds(cache, '/', true)).toEqual(['1', '2']);
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
        const link = createStreamingLink(createCache());
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

    it('takes the closing frame as the whole message, not as another delta', () => {
        const { emitted, pushLog } = drive();

        pushLog({ appendPart: true, id: 'a', message: 'lo ', result: null, thinking: null });
        now = 1010;
        pushLog({ appendPart: true, id: 'a', message: 'Wor', result: null, thinking: null }); // throttled

        pushLog({ appendPart: false, id: 'a', message: 'Hello World', result: null, thinking: null });
        expect(emitted.at(-1)?.assistantLogUpdated?.message).toBe('Hello World');

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

    const openStream = () => {
        const link = createStreamingLink(createCache());
        const emitted: EmittedData[] = [];
        let source: undefined | { next: (value: unknown) => void };

        const forward = () =>
            new Observable((observer) => {
                source = observer;

                return () => {};
            });

        const open = () =>
            link.request({} as never, forward as never)!.subscribe({
                next: (result) => emitted.push(result.data as EmittedData),
            });

        return {
            open,
            push: (message: string) =>
                source?.next({ data: { assistantLogUpdated: { appendPart: true, id: 'a', message } } }),
            resumed: () => emitted.at(-1)?.assistantLogUpdated,
        };
    };

    it('drops what it accumulated when its subscription goes away', () => {
        const { open, push, resumed } = openStream();
        const first = open();

        push('ABCDE');
        first.unsubscribe();
        open();
        now = 9000;
        push('LMN');

        expect(resumed()?.message).toBe('LMN');
    });

    it('keeps a live stream accumulating when an unrelated operation on the same link ends', () => {
        const link = createStreamingLink(createCache());
        const emitted: EmittedData[] = [];
        let stream: undefined | { next: (value: unknown) => void };

        const forwardStream = () =>
            new Observable((observer) => {
                stream = observer;

                return () => {};
            });

        const forwardQuery = () =>
            new Observable<{ data: unknown }>((observer) => {
                observer.next({ data: { flows: [] } });
                observer.complete();

                return () => {};
            });

        link.request({} as never, forwardStream as never)!.subscribe({
            next: (result) => emitted.push(result.data as EmittedData),
        });

        stream?.next({ data: { assistantLogUpdated: { appendPart: true, id: 'a', message: 'Hello' } } });

        link.request({} as never, forwardQuery as never)!.subscribe({ next: () => {} });

        now = 9000;
        stream?.next({ data: { assistantLogUpdated: { appendPart: true, id: 'a', message: ' world' } } });

        expect(emitted.at(-1)?.assistantLogUpdated?.message).toBe('Hello world');
    });

    it('resumes a reconnected stream from the cached row instead of publishing only the tail', () => {
        const cache = createCache();
        const link = createStreamingLink(cache);
        const emitted: EmittedData[] = [];
        let source: undefined | { next: (value: unknown) => void };

        link.request(
            {} as never,
            (() =>
                new Observable((observer) => {
                    source = observer;

                    return () => {};
                })) as never,
        )!.subscribe({
            next: (result) => emitted.push(result.data as EmittedData),
        });

        const push = (message: string) =>
            source?.next({ data: { assistantLogUpdated: { appendPart: true, id: 'a', message } } });

        push('ABCDE');

        // The link above this one lands every published frame in the cache.
        cache.writeFragment({
            data: { __typename: 'AssistantLog', id: 'a', message: 'ABCDE', result: '', thinking: null },
            fragment: gql`
                fragment Resumed on AssistantLog {
                    message
                    result
                    thinking
                }
            `,
            id: 'AssistantLog:a',
        });

        window.dispatchEvent(new Event('ws:reconnected'));
        now = 9000;
        push('LMN');

        expect(emitted.at(-1)?.assistantLogUpdated?.message).toBe('ABCDELMN');
    });

    it('drops what it accumulated when the socket reconnects under a live subscription', () => {
        const { open, push, resumed } = openStream();

        open();
        push('ABCDE');
        window.dispatchEvent(new Event('ws:reconnected'));
        now = 9000;
        push('LMN');

        expect(resumed()?.message).toBe('LMN');
    });
});

describe('streamed assistant text in the cache', () => {
    const STREAMED = gql`
        fragment Streamed on AssistantLog {
            message
            result
            thinking
        }
    `;

    const write = (cache: InMemoryCache, row: Record<string, unknown>) =>
        cache.writeFragment({
            data: { __typename: 'AssistantLog', id: '7', ...row },
            fragment: STREAMED,
            id: 'AssistantLog:7',
        });

    const read = (cache: InMemoryCache) =>
        cache.readFragment<{ message: string; result: string; thinking: string }>({
            fragment: STREAMED,
            id: 'AssistantLog:7',
        });

    it('keeps what the stream already delivered when a stale server row lands on top', () => {
        const cache = createCache();

        write(cache, { message: 'Hello world, still streaming', result: '', thinking: 'thinking it over' });
        write(cache, { message: 'Hel', result: '', thinking: 'thi' });

        expect(read(cache)?.message).toBe('Hello world, still streaming');
        expect(read(cache)?.thinking).toBe('thinking it over');
    });

    it('takes a revision that is not a prefix of what it holds', () => {
        const cache = createCache();

        write(cache, { message: 'Hello world, still streaming', result: '', thinking: '' });
        write(cache, { message: 'A shorter, different answer', result: '', thinking: '' });

        expect(read(cache)?.message).toBe('A shorter, different answer');
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

describe('subscription cache link (createSubscriptionCacheLink)', () => {
    let cache: InMemoryCache;

    beforeEach(() => {
        cache = makeCache();
    });

    const subscriptionField = (document: DocumentNode): string => {
        const [definition] = document.definitions as [OperationDefinitionNode];
        const [selection] = definition.selectionSet.selections as [FieldNode];

        return selection.name.value;
    };

    const drive = () => {
        const link = createSubscriptionCacheLink(cache);
        let source: undefined | { next: (value: unknown) => void };
        const forward = () =>
            new Observable((observer) => {
                source = observer;

                return () => {};
            });
        const emitted: unknown[] = [];

        link.request({ variables: {} } as never, forward as never)!.subscribe({
            next: (result) => emitted.push(result.data),
        });

        return {
            emitted,
            push: (document: DocumentNode, item: Record<string, unknown>) =>
                source?.next({ data: { [subscriptionField(document)]: item } }),
        };
    };

    const token = (id: string, name: string) => ({ __typename: 'APIToken', id, name, tokenId: `t${id}` });

    const seed = (...tokens: ReturnType<typeof token>[]) =>
        cache.writeQuery({ data: { apiTokens: tokens }, query: API_TOKENS });

    it('routes a created token to the front of the list', () => {
        seed(token('1', 'first'));

        const { push } = drive();

        push(ApiTokenCreatedDocument, token('2', 'second'));

        expect(apiTokenIds(cache)).toEqual(['2', '1']);
    });

    it('routes a deleted token out of the list', () => {
        seed(token('1', 'first'), token('2', 'second'));

        const { push } = drive();

        push(ApiTokenDeletedDocument, token('1', 'first'));

        expect(apiTokenIds(cache)).toEqual(['2']);
    });

    it('routes an updated token without moving it', () => {
        seed(token('1', 'first'), token('2', 'second'));

        const { push } = drive();

        push(ApiTokenUpdatedDocument, token('1', 'renamed'));

        const read = cache.readQuery<{ apiTokens: { id: string; name: string }[] }>({ query: API_TOKENS });

        expect(read?.apiTokens.map((entry) => entry.id)).toEqual(['1', '2']);
        expect(read?.apiTokens[0]?.name).toBe('renamed');
    });

    it('forgets the oldest local write once the tracker is full', () => {
        const { push } = drive();

        push(ApiTokenCreatedDocument, token('oldest', 'oldest'));

        expect(lastLocalWriteAt('apiTokens', 'oldest')).toBeGreaterThan(0);

        for (let index = 0; index < 1000; index += 1) {
            push(ApiTokenCreatedDocument, token(`later-${index}`, 'later'));
        }

        expect(lastLocalWriteAt('apiTokens', 'later-999')).toBeGreaterThan(0);
        expect(lastLocalWriteAt('apiTokens', 'oldest')).toBe(0);
    });

    it('ignores a redelivered create instead of duplicating the row', () => {
        seed(token('1', 'first'));

        const { push } = drive();

        push(ApiTokenCreatedDocument, token('2', 'second'));
        push(ApiTokenCreatedDocument, token('2', 'second'));

        expect(apiTokenIds(cache)).toEqual(['2', '1']);
    });

    it('treats a numeric id as the string one already in the list', () => {
        cache.writeQuery({
            data: { apiTokens: [{ __typename: 'APIToken', id: 2, name: 'second', tokenId: 't2' }] },
            query: API_TOKENS,
        });

        const { push } = drive();

        push(ApiTokenCreatedDocument, token('2', 'second'));

        expect(apiTokenIds(cache)).toEqual(['2']);
    });

    it('forwards a payload it cannot key on instead of swallowing the frame', () => {
        seed(token('1', 'first'));

        const { emitted, push } = drive();
        const before = cache.extract();
        const payload = { __typename: 'APIToken', name: 'no id', tokenId: 't9' };

        push(ApiTokenCreatedDocument, payload);

        expect(cache.extract()).toEqual(before);
        expect(emitted).toEqual([{ [subscriptionField(ApiTokenCreatedDocument)]: payload }]);
    });
});

describe('query key arguments', () => {
    const argumentsByField = (): Map<string, Set<string>> => {
        const byField = new Map<string, Set<string>>();

        for (const definition of parse(operationsDocument).definitions) {
            if (definition.kind !== 'OperationDefinition' || definition.operation !== 'query') {
                continue;
            }

            for (const selection of definition.selectionSet.selections) {
                if (selection.kind !== 'Field') {
                    continue;
                }

                const names = byField.get(selection.name.value) ?? new Set<string>();

                for (const argument of selection.arguments ?? []) {
                    names.add(argument.name.value);
                }

                byField.set(selection.name.value, names);
            }
        }

        return byField;
    };

    it('keys exactly the collections it is meant to', () => {
        expect(Object.keys(queryKeyArgs).sort()).toEqual([
            'agentLogs',
            'assistantLogs',
            'assistants',
            'flowFiles',
            'knowledgeDocuments',
            'messageLogs',
            'resources',
            'screenshots',
            'searchLogs',
            'tasks',
            'terminalLogs',
            'vectorStoreLogs',
        ]);
    });

    it('merges exactly the lists nothing can remove a row from', () => {
        expect([...APPEND_ONLY_LIST_FIELDS].sort()).toEqual([
            'agentLogs',
            'messageLogs',
            'screenshots',
            'searchLogs',
            'tasks',
            'terminalLogs',
            'vectorStoreLogs',
        ]);
    });

    it.each(Object.entries(queryKeyArgs))('%s keys on arguments the queries really pass', (field, keyArgs) => {
        const sent = argumentsByField().get(field);

        expect(sent, `no query selects ${field}; the policy keys a field nothing fetches`).toBeDefined();
        expect([...keyArgs].filter((name) => !sent?.has(name))).toEqual([]);
    });
});

describe('cache configuration the app depends on', () => {
    it('pins every watch-query default, not only the one with a behavioural test', () => {
        expect(watchQueryDefaults).toEqual({
            fetchPolicy: 'cache-and-network',
            nextFetchPolicy: 'cache-first',
            notifyOnNetworkStatusChange: true,
            refetchWritePolicy: 'merge',
        });
    });

    it('keeps a knowledge body the detail query loaded when the list writes it blank', () => {
        const cache = createCache();

        cache.writeQuery({
            data: { knowledgeDocuments: [{ __typename: 'KnowledgeDocument', content: 'BODY', id: '1' }] },
            query: KNOWLEDGE,
            variables: { withContent: true },
        });
        cache.writeQuery({
            data: { knowledgeDocuments: [{ __typename: 'KnowledgeDocument', content: '', id: '1' }] },
            query: KNOWLEDGE,
            variables: { withContent: false },
        });

        const read = cache.readQuery<{ knowledgeDocuments: { content: string }[] }>({
            query: KNOWLEDGE,
            variables: { withContent: true },
        });

        expect(read?.knowledgeDocuments[0]?.content).toBe('BODY');
    });

    it('replaces the whole user-preferences object a settings frame carries', () => {
        const cache = createCache();

        cache.writeQuery({
            data: { settingsUser: { __typename: 'UserPreferences', favoriteFlows: ['1'], id: 'u1' } },
            query: SETTINGS_USER,
        });

        updateCacheForSubscription(
            cache,
            'settingsUserUpdated',
            'settingsUser',
            frame({ __typename: 'UserPreferences', favoriteFlows: ['1', '2'], id: 'u1' }),
        );

        expect(
            cache.readQuery<{ settingsUser: { favoriteFlows: string[] } }>({ query: SETTINGS_USER })?.settingsUser
                .favoriteFlows,
        ).toEqual(['1', '2']);
    });
});

describe('a subscription writes an entity the list query can still read', () => {
    const document = parse(operationsDocument);

    const fragments = new Map(
        document.definitions
            .filter((definition) => definition.kind === 'FragmentDefinition')
            .map((definition) => [definition.name.value, definition]),
    );

    const leaves = (selectionSet: SelectionSetNode | undefined, prefix = '', seen = new Set<string>()): Set<string> => {
        const paths = new Set<string>();

        for (const selection of selectionSet?.selections ?? []) {
            if (selection.kind === 'Field') {
                const path = prefix + selection.name.value;

                paths.add(path);

                for (const nested of leaves(selection.selectionSet, `${path}.`, seen)) {
                    paths.add(nested);
                }
            } else if (selection.kind === 'FragmentSpread' && !seen.has(prefix + selection.name.value)) {
                seen.add(prefix + selection.name.value);

                for (const nested of leaves(fragments.get(selection.name.value)?.selectionSet, prefix, seen)) {
                    paths.add(nested);
                }
            }
        }

        return paths;
    };

    it('walks into nested selections, so narrowing one is not read as parity', () => {
        const { selectionSet } = parse('{ id subtasks { id title } }').definitions[0] as OperationDefinitionNode;

        expect([...leaves(selectionSet)].sort()).toEqual(['id', 'subtasks', 'subtasks.id', 'subtasks.title']);
    });

    const rootFields = (operation: 'query' | 'subscription') =>
        document.definitions
            .filter(
                (definition): definition is OperationDefinitionNode =>
                    definition.kind === 'OperationDefinition' && definition.operation === operation,
            )
            .flatMap((definition) => definition.selectionSet.selections)
            .filter((selection): selection is FieldNode => selection.kind === 'Field');

    const selectedByQuery = new Map<string, Set<string>>();

    for (const field of rootFields('query')) {
        const known = selectedByQuery.get(field.name.value) ?? new Set<string>();

        for (const name of leaves(field.selectionSet)) {
            known.add(name);
        }

        selectedByQuery.set(field.name.value, known);
    }

    const writesNoList = ['providerCreated', 'providerDeleted', 'providerUpdated'];

    it('routes every subscription the document declares, or says why it routes none', () => {
        const subscribed = rootFields('subscription').map(({ name }) => name.value);
        const routed = Object.keys(subscriptionToCacheFieldMap);

        expect(
            subscribed.filter((name) => !routed.includes(name) && !writesNoList.includes(name)).sort(),
            'a subscription that reaches neither list silently stops updating its collection',
        ).toEqual([]);
        expect(routed.filter((name) => !subscribed.includes(name)).sort()).toEqual([]);
        expect(writesNoList.filter((name) => routed.includes(name)).sort()).toEqual([]);
    });

    const membershipWrites = rootFields('subscription')
        .filter(({ name }) => !name.value.endsWith('Deleted'))
        .map(({ name, selectionSet }) => ({
            cacheField: subscriptionToCacheFieldMap[name.value],
            selected: leaves(selectionSet),
            subscription: name.value,
        }))
        .filter(
            (write): write is { cacheField: string; selected: Set<string>; subscription: string } =>
                write.cacheField !== undefined && selectedByQuery.has(write.cacheField),
        );

    it.each(membershipWrites)('$subscription carries every field the $cacheField query reads', (write) => {
        const queried = selectedByQuery.get(write.cacheField) ?? new Set<string>();

        expect([...queried].filter((name) => !write.selected.has(name)).sort()).toEqual([]);
    });
});

const GUARDED_LOGS = [
    { document: MESSAGE_LOGS, field: 'messageLogs', typename: 'MessageLog' },
    { document: ASSISTANT_LOGS_WITH_RESULT, field: 'assistantLogs', typename: 'AssistantLog' },
] as const;

describe.each(GUARDED_LOGS)('$typename result guard', ({ document, field, typename }) => {
    let cache: InMemoryCache;

    const variables = { assistantId: '3', flowId: '7' };

    const write = (result: string, resultFormat: string) =>
        cache.writeQuery({
            data: {
                [field]: [{ __typename: typename, id: '1', message: 'run hostname', result, resultFormat }],
            },
            query: document,
            variables,
        });

    const readRow = () =>
        cache.readQuery<Record<string, { result: string; resultFormat: string }[]>>({
            query: document,
            variables,
        })?.[field]?.[0];

    beforeEach(() => {
        cache = makeCache();
    });

    it('keeps a delivered result when a later response carries the column defaults', () => {
        write('', 'plain');
        write('FULL RESULT', 'markdown');

        expect(readRow()?.result).toBe('FULL RESULT');

        write('', 'plain');

        expect(readRow()?.result).toBe('FULL RESULT');
        expect(readRow()?.resultFormat).toBe('markdown');
    });

    it('still accepts a result the server actually replaced', () => {
        write('FIRST', 'plain');
        write('SECOND', 'terminal');

        expect(readRow()?.result).toBe('SECOND');
        expect(readRow()?.resultFormat).toBe('terminal');
    });
});

describe('refetch write policy', () => {
    const MESSAGE_LOGS_REFETCH = gql`
        query MLR($flowId: ID!) {
            messageLogs(flowId: $flowId) {
                id
                result
                resultFormat
            }
        }
    `;

    const row = (result: string, resultFormat: string) => ({
        __typename: 'MessageLog',
        id: '1',
        result,
        resultFormat,
    });

    it('a refetch does not blank a result the subscription already delivered', async () => {
        const responses = [{ messageLogs: [row('FULL RESULT', 'markdown')] }, { messageLogs: [row('', 'plain')] }];

        const link = new ApolloLink(
            () =>
                new Observable((observer) => {
                    observer.next({ data: responses.shift() });
                    observer.complete();
                }),
        );

        const client = new ApolloClient({
            cache: createCache(),
            defaultOptions: { watchQuery: watchQueryDefaults },
            link,
        });

        const observable = client.watchQuery({ query: MESSAGE_LOGS_REFETCH, variables: { flowId: '7' } });
        const seen: unknown[] = [];
        const subscription = observable.subscribe((result) => {
            if (result.data) {
                seen.push(result.data);
            }
        });

        while (seen.length === 0) {
            await new Promise((resolve) => setTimeout(resolve, 5));
        }

        await observable.refetch();
        subscription.unsubscribe();

        const read = client.cache.readQuery<{ messageLogs: { result: string; resultFormat: string }[] }>({
            query: MESSAGE_LOGS_REFETCH,
            variables: { flowId: '7' },
        });

        expect(read?.messageLogs[0]?.result).toBe('FULL RESULT');
        expect(read?.messageLogs[0]?.resultFormat).toBe('markdown');
    });
});

describe('a snapshot in flight against a live stream', () => {
    const log = (id: string) => ({ __typename: 'TerminalLog', id, text: `line ${id}` });

    const clientWithHeldSnapshot = (payload: Record<string, unknown>) => {
        let release = () => {};

        const held = new Promise<void>((resolve) => {
            release = resolve;
        });

        const link = new ApolloLink(
            () =>
                new Observable((observer) => {
                    void held.then(() => {
                        observer.next({ data: payload });
                        observer.complete();
                    });
                }),
        );

        return {
            client: new ApolloClient({
                cache: createCache(),
                defaultOptions: { watchQuery: watchQueryDefaults },
                link,
            }),
            release: () => release(),
        };
    };

    it('keeps a row the stream appended while a network-only query was on the wire', async () => {
        const { client, release } = clientWithHeldSnapshot({ terminalLogs: [log('1')] });

        client.cache.writeQuery({ data: { terminalLogs: [log('1')] }, query: TERMINAL, variables: { flowId: '7' } });

        const inFlight = client.query({ fetchPolicy: 'network-only', query: TERMINAL, variables: { flowId: '7' } });

        updateCacheForSubscription(client.cache as InMemoryCache, 'terminalLogAdded', 'terminalLogs', log('2'), {
            flowId: '7',
        });

        release();
        await inFlight;

        expect(termIds(client.cache as InMemoryCache, '7')).toEqual(['1', '2']);
    });

    it('keeps it across the reconnect sweep, which refetches every watched query', async () => {
        let release = () => {};

        const held = new Promise<void>((resolve) => {
            release = resolve;
        });
        let isFirst = true;

        const link = new ApolloLink(
            () =>
                new Observable((observer) => {
                    const deliver = () => {
                        observer.next({ data: { terminalLogs: [log('1')] } });
                        observer.complete();
                    };

                    if (isFirst) {
                        isFirst = false;
                        deliver();

                        return;
                    }

                    void held.then(deliver);
                }),
        );

        const client = new ApolloClient({
            cache: createCache(),
            defaultOptions: { watchQuery: watchQueryDefaults },
            link,
        });

        const observable = client.watchQuery({ query: TERMINAL, variables: { flowId: '7' } });
        const subscription = observable.subscribe(() => {});

        while (!termIds(client.cache as InMemoryCache, '7')) {
            await new Promise((resolve) => setTimeout(resolve, 5));
        }

        const swept = client.refetchObservableQueries();

        updateCacheForSubscription(client.cache as InMemoryCache, 'terminalLogAdded', 'terminalLogs', log('2'), {
            flowId: '7',
        });

        release();
        await swept;
        subscription.unsubscribe();

        expect(termIds(client.cache as InMemoryCache, '7')).toEqual(['1', '2']);
    });

    it('still lets a list whose rows the server can delete shrink', () => {
        const cache = makeCache();

        cache.writeQuery({
            data: { assistantLogs: [{ __typename: 'AssistantLog', assistantId: '3', id: '1' }] },
            query: ASSISTANT_LOGS,
            variables: { assistantId: '3', flowId: '7' },
        });
        cache.writeQuery({
            data: { assistantLogs: [] },
            query: ASSISTANT_LOGS,
            variables: { assistantId: '3', flowId: '7' },
        });

        expect(assistantLogIds(cache, '7', '3')).toEqual([]);
    });

    it('keeps the pane when a per-field error nulls the list', () => {
        const cache = makeCache();

        cache.writeQuery({ data: { terminalLogs: [log('1')] }, query: TERMINAL, variables: { flowId: '7' } });
        cache.writeQuery({ data: { terminalLogs: null }, query: TERMINAL, variables: { flowId: '7' } });

        expect(termIds(cache, '7')).toEqual(['1']);
    });

    it('records the null a per-field error delivers to a cold cache', () => {
        const cache = makeCache();

        cache.writeQuery({ data: { terminalLogs: null }, query: TERMINAL, variables: { flowId: '7' } });

        expect(cache.diff({ optimistic: false, query: TERMINAL, variables: { flowId: '7' } }).complete).toBe(true);
    });

    it('takes a delta for a list a per-field error nulled on a cold cache', () => {
        const cache = makeCache();

        cache.writeQuery({ data: { terminalLogs: null }, query: TERMINAL, variables: { flowId: '7' } });
        updateCacheForSubscription(cache, 'terminalLogAdded', 'terminalLogs', log('1'), { flowId: '7' });

        expect(termIds(cache, '7')).toEqual(['1']);
    });
});

describe('socket presence', () => {
    const openSubscription = (presence: ReturnType<typeof createSocketPresence>) => {
        const forward = () => new Observable(() => () => undefined);

        return presence.link.request({} as never, forward as never)!.subscribe({ next: () => undefined });
    };

    const record = () => {
        const seen: string[] = [];
        const down = () => seen.push('down');
        const up = () => seen.push('up');

        window.addEventListener('ws:disconnected', down);
        window.addEventListener('ws:connected', up);

        return {
            seen,
            stop: () => {
                window.removeEventListener('ws:disconnected', down);
                window.removeEventListener('ws:connected', up);
            },
        };
    };

    it('stays quiet when the socket closes with nothing subscribed', () => {
        const presence = createSocketPresence();
        const { seen, stop } = record();

        presence.closed();

        expect(seen, 'the lazy self-close after the last unsubscribe is not a lost connection').toEqual([]);
        stop();
    });

    it('reports a close that happened under a live subscription, and its recovery', () => {
        const presence = createSocketPresence();
        const { seen, stop } = record();

        openSubscription(presence);
        presence.closed();
        presence.connected();

        expect(seen).toEqual(['down', 'up']);
        stop();
    });

    it('reports being down only once however many times the socket retries', () => {
        const presence = createSocketPresence();
        const { seen, stop } = record();

        openSubscription(presence);
        presence.closed();
        presence.closed();
        presence.closed();

        expect(seen).toEqual(['down']);
        stop();
    });

    it('takes the report back when the last subscription leaves while the socket is down', () => {
        const presence = createSocketPresence();
        const { seen, stop } = record();

        const subscription = openSubscription(presence);

        presence.closed();
        subscription.unsubscribe();

        expect(seen, 'with nothing subscribed the socket stops retrying, so "reconnecting" is a lie').toEqual([
            'down',
            'up',
        ]);
        stop();
    });

    it('forgets a subscription that unsubscribed before the close', () => {
        const presence = createSocketPresence();
        const { seen, stop } = record();

        openSubscription(presence).unsubscribe();
        presence.closed();

        expect(seen).toEqual([]);
        stop();
    });
});

describe('GraphQL error message', () => {
    it('states a reason shared by every failed field once', () => {
        const errors = Array.from({ length: 8 }, (_, index) => ({
            message: 'a required service is unavailable',
            path: [`field${index}`],
        }));

        expect(new CombinedGraphQLErrors({ errors }).message).toBe('a required service is unavailable');
    });

    it('keeps distinct reasons in the order they arrived', () => {
        const errors = [{ message: 'flow not found' }, { message: 'forbidden' }, { message: 'flow not found' }];

        expect(new CombinedGraphQLErrors({ errors }).message).toBe('flow not found\nforbidden');
    });
});
