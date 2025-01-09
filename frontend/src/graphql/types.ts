import * as Apollo from '@apollo/client';
import { gql } from '@apollo/client';

export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends Record<string, unknown>> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends Record<string, unknown>, K extends keyof T> = Partial<Record<K, never>>;
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
const defaultOptions = {} as const;
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
    ID: { input: string; output: string };
    String: { input: string; output: string };
    Boolean: { input: boolean; output: boolean };
    Int: { input: number; output: number };
    Float: { input: number; output: number };
    Time: { input: any; output: any };
};

export type AgentLog = {
    createdAt: Scalars['Time']['output'];
    executor: AgentType;
    flowId: Scalars['ID']['output'];
    id: Scalars['ID']['output'];
    initiator: AgentType;
    result: Scalars['String']['output'];
    subtaskId?: Maybe<Scalars['ID']['output']>;
    task: Scalars['String']['output'];
    taskId?: Maybe<Scalars['ID']['output']>;
};

export enum AgentType {
    Adviser = 'adviser',
    Coder = 'coder',
    Enricher = 'enricher',
    Generator = 'generator',
    Installer = 'installer',
    Memorist = 'memorist',
    Pentester = 'pentester',
    PrimaryAgent = 'primary_agent',
    Refiner = 'refiner',
    Reflector = 'reflector',
    Reporter = 'reporter',
    Searcher = 'searcher',
    Summarizer = 'summarizer',
}

export type Flow = {
    createdAt: Scalars['Time']['output'];
    id: Scalars['ID']['output'];
    provider: Scalars['String']['output'];
    status: StatusType;
    terminals?: Maybe<Array<Terminal>>;
    title: Scalars['String']['output'];
    updatedAt: Scalars['Time']['output'];
};

export type MessageLog = {
    createdAt: Scalars['Time']['output'];
    flowId: Scalars['ID']['output'];
    id: Scalars['ID']['output'];
    message: Scalars['String']['output'];
    result: Scalars['String']['output'];
    resultFormat: ResultFormat;
    subtaskId?: Maybe<Scalars['ID']['output']>;
    taskId?: Maybe<Scalars['ID']['output']>;
    type: MessageLogType;
};

export enum MessageLogType {
    Advice = 'advice',
    Ask = 'ask',
    Browser = 'browser',
    Done = 'done',
    File = 'file',
    Input = 'input',
    Search = 'search',
    Terminal = 'terminal',
    Thoughts = 'thoughts',
}

export type Mutation = {
    createFlow: Flow;
    deleteFlow: ResultType;
    finishFlow: Flow;
    putUserInput: ResultType;
    resetPrompt: ResultType;
    updatePrompt: ResultType;
};

export type MutationCreateFlowArgs = {
    input: Scalars['String']['input'];
    modelProvider: Scalars['String']['input'];
};

export type MutationDeleteFlowArgs = {
    flowId: Scalars['ID']['input'];
};

export type MutationFinishFlowArgs = {
    flowId: Scalars['ID']['input'];
};

export type MutationPutUserInputArgs = {
    flowId: Scalars['ID']['input'];
    input: Scalars['String']['input'];
};

export type MutationResetPromptArgs = {
    promptType: Scalars['String']['input'];
};

export type MutationUpdatePromptArgs = {
    prompt: Scalars['String']['input'];
    promptType: Scalars['String']['input'];
};

export type Prompt = {
    prompt: Scalars['String']['output'];
    type: Scalars['String']['output'];
};

export type Query = {
    agentLogs?: Maybe<Array<AgentLog>>;
    flow: Flow;
    flows?: Maybe<Array<Flow>>;
    messageLogs?: Maybe<Array<MessageLog>>;
    prompt: Scalars['String']['output'];
    prompts: Array<Prompt>;
    providers: Array<Scalars['String']['output']>;
    screenshots?: Maybe<Array<Screenshot>>;
    searchLogs?: Maybe<Array<SearchLog>>;
    tasks?: Maybe<Array<Task>>;
    terminalLogs?: Maybe<Array<TerminalLog>>;
    vectorStoreLogs?: Maybe<Array<VectorStoreLog>>;
};

export type QueryAgentLogsArgs = {
    flowId: Scalars['ID']['input'];
};

export type QueryFlowArgs = {
    flowId: Scalars['ID']['input'];
};

export type QueryMessageLogsArgs = {
    flowId: Scalars['ID']['input'];
};

export type QueryPromptArgs = {
    promptType: Scalars['String']['input'];
};

export type QueryScreenshotsArgs = {
    flowId: Scalars['ID']['input'];
};

export type QuerySearchLogsArgs = {
    flowId: Scalars['ID']['input'];
};

export type QueryTasksArgs = {
    flowId: Scalars['ID']['input'];
};

export type QueryTerminalLogsArgs = {
    flowId: Scalars['ID']['input'];
};

export type QueryVectorStoreLogsArgs = {
    flowId: Scalars['ID']['input'];
};

export enum ResultFormat {
    Markdown = 'markdown',
    Plain = 'plain',
    Terminal = 'terminal',
}

export enum ResultType {
    Error = 'error',
    Success = 'success',
}

export type Screenshot = {
    createdAt: Scalars['Time']['output'];
    flowId: Scalars['ID']['output'];
    id: Scalars['ID']['output'];
    name: Scalars['String']['output'];
    url: Scalars['String']['output'];
};

export type SearchLog = {
    createdAt: Scalars['Time']['output'];
    engine: Scalars['String']['output'];
    executor: AgentType;
    flowId: Scalars['ID']['output'];
    id: Scalars['ID']['output'];
    initiator: AgentType;
    query: Scalars['String']['output'];
    result: Scalars['String']['output'];
    subtaskId?: Maybe<Scalars['ID']['output']>;
    taskId?: Maybe<Scalars['ID']['output']>;
};

export enum StatusType {
    Failed = 'failed',
    Finished = 'finished',
    Running = 'running',
    Starting = 'starting',
    Waiting = 'waiting',
}

export type Subscription = {
    agentLogAdded: AgentLog;
    flowCreated: Flow;
    flowDeleted: Flow;
    flowUpdated: Flow;
    messageLogAdded: MessageLog;
    messageLogUpdated: MessageLog;
    screenshotAdded: Screenshot;
    searchLogAdded: SearchLog;
    taskCreated: Task;
    taskUpdated: Task;
    terminalLogAdded: TerminalLog;
    vectorStoreLogAdded: VectorStoreLog;
};

export type SubscriptionAgentLogAddedArgs = {
    flowId: Scalars['ID']['input'];
};

export type SubscriptionFlowUpdatedArgs = {
    flowId: Scalars['ID']['input'];
};

export type SubscriptionMessageLogAddedArgs = {
    flowId: Scalars['ID']['input'];
};

export type SubscriptionMessageLogUpdatedArgs = {
    flowId: Scalars['ID']['input'];
};

export type SubscriptionScreenshotAddedArgs = {
    flowId: Scalars['ID']['input'];
};

export type SubscriptionSearchLogAddedArgs = {
    flowId: Scalars['ID']['input'];
};

export type SubscriptionTaskCreatedArgs = {
    flowId: Scalars['ID']['input'];
};

export type SubscriptionTaskUpdatedArgs = {
    flowId: Scalars['ID']['input'];
};

export type SubscriptionTerminalLogAddedArgs = {
    flowId: Scalars['ID']['input'];
};

export type SubscriptionVectorStoreLogAddedArgs = {
    flowId: Scalars['ID']['input'];
};

export type Subtask = {
    createdAt: Scalars['Time']['output'];
    description: Scalars['String']['output'];
    id: Scalars['ID']['output'];
    result: Scalars['String']['output'];
    status: StatusType;
    taskId: Scalars['ID']['output'];
    title: Scalars['String']['output'];
    updatedAt: Scalars['Time']['output'];
};

export type Task = {
    createdAt: Scalars['Time']['output'];
    flowId: Scalars['ID']['output'];
    id: Scalars['ID']['output'];
    input: Scalars['String']['output'];
    result: Scalars['String']['output'];
    status: StatusType;
    subtasks?: Maybe<Array<Subtask>>;
    title: Scalars['String']['output'];
    updatedAt: Scalars['Time']['output'];
};

export type Terminal = {
    connected: Scalars['Boolean']['output'];
    createdAt: Scalars['Time']['output'];
    id: Scalars['ID']['output'];
    image: Scalars['String']['output'];
    name: Scalars['String']['output'];
    type: TerminalType;
};

export type TerminalLog = {
    createdAt: Scalars['Time']['output'];
    flowId: Scalars['ID']['output'];
    id: Scalars['ID']['output'];
    terminal: Scalars['ID']['output'];
    text: Scalars['String']['output'];
    type: TerminalLogType;
};

export enum TerminalLogType {
    Stderr = 'stderr',
    Stdin = 'stdin',
    Stdout = 'stdout',
}

export enum TerminalType {
    Primary = 'primary',
    Secondary = 'secondary',
}

export enum VectorStoreAction {
    Retrieve = 'retrieve',
    Store = 'store',
}

export type VectorStoreLog = {
    action: VectorStoreAction;
    createdAt: Scalars['Time']['output'];
    executor: AgentType;
    filter: Scalars['String']['output'];
    flowId: Scalars['ID']['output'];
    id: Scalars['ID']['output'];
    initiator: AgentType;
    query: Scalars['String']['output'];
    result: Scalars['String']['output'];
    subtaskId?: Maybe<Scalars['ID']['output']>;
    taskId?: Maybe<Scalars['ID']['output']>;
};

export type FlowOverviewFragmentFragment = { id: string; title: string; status: StatusType };

export type FlowsQueryVariables = Exact<Record<string, never>>;

export type FlowsQuery = { flows?: Array<FlowOverviewFragmentFragment> | null };

export type ProvidersQueryVariables = Exact<Record<string, never>>;

export type ProvidersQuery = { providers: Array<string> };

export type FlowFragmentFragment = {
    id: string;
    title: string;
    status: StatusType;
    provider: string;
    createdAt: any;
    updatedAt: any;
    terminals?: Array<TerminalFragmentFragment> | null;
};

export type TerminalFragmentFragment = {
    id: string;
    type: TerminalType;
    name: string;
    image: string;
    connected: boolean;
    createdAt: any;
};

export type TaskFragmentFragment = {
    id: string;
    title: string;
    status: StatusType;
    input: string;
    result: string;
    flowId: string;
    createdAt: any;
    updatedAt: any;
    subtasks?: Array<SubtaskFragmentFragment> | null;
};

export type SubtaskFragmentFragment = {
    id: string;
    status: StatusType;
    title: string;
    description: string;
    result: string;
    taskId: string;
    createdAt: any;
    updatedAt: any;
};

export type TerminalLogFragmentFragment = {
    id: string;
    flowId: string;
    type: TerminalLogType;
    text: string;
    terminal: string;
    createdAt: any;
};

export type MessageLogFragmentFragment = {
    id: string;
    type: MessageLogType;
    message: string;
    result: string;
    resultFormat: ResultFormat;
    flowId: string;
    taskId?: string | null;
    subtaskId?: string | null;
    createdAt: any;
};

export type ScreenshotFragmentFragment = { id: string; flowId: string; name: string; url: string; createdAt: any };

export type AgentLogFragmentFragment = {
    id: string;
    flowId: string;
    initiator: AgentType;
    executor: AgentType;
    task: string;
    result: string;
    taskId?: string | null;
    subtaskId?: string | null;
    createdAt: any;
};

export type SearchLogFragmentFragment = {
    id: string;
    flowId: string;
    initiator: AgentType;
    executor: AgentType;
    engine: string;
    query: string;
    result: string;
    taskId?: string | null;
    subtaskId?: string | null;
    createdAt: any;
};

export type VectorStoreLogFragmentFragment = {
    id: string;
    flowId: string;
    initiator: AgentType;
    executor: AgentType;
    filter: string;
    query: string;
    action: VectorStoreAction;
    result: string;
    taskId?: string | null;
    subtaskId?: string | null;
    createdAt: any;
};

export type FlowQueryVariables = Exact<{
    id: Scalars['ID']['input'];
}>;

export type FlowQuery = {
    flow: FlowFragmentFragment;
    tasks?: Array<TaskFragmentFragment> | null;
    screenshots?: Array<ScreenshotFragmentFragment> | null;
    terminalLogs?: Array<TerminalLogFragmentFragment> | null;
    messageLogs?: Array<MessageLogFragmentFragment> | null;
    agentLogs?: Array<AgentLogFragmentFragment> | null;
    searchLogs?: Array<SearchLogFragmentFragment> | null;
    vectorStoreLogs?: Array<VectorStoreLogFragmentFragment> | null;
};

export type TasksQueryVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type TasksQuery = { tasks?: Array<TaskFragmentFragment> | null };

export type CreateFlowMutationVariables = Exact<{
    modelProvider: Scalars['String']['input'];
    input: Scalars['String']['input'];
}>;

export type CreateFlowMutation = { createFlow: FlowFragmentFragment };

export type DeleteFlowMutationVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type DeleteFlowMutation = { deleteFlow: ResultType };

export type PutUserInputMutationVariables = Exact<{
    flowId: Scalars['ID']['input'];
    input: Scalars['String']['input'];
}>;

export type PutUserInputMutation = { putUserInput: ResultType };

export type FinishFlowMutationVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type FinishFlowMutation = { finishFlow: FlowFragmentFragment };

export type TerminalLogAddedSubscriptionVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type TerminalLogAddedSubscription = { terminalLogAdded: TerminalLogFragmentFragment };

export type MessageLogAddedSubscriptionVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type MessageLogAddedSubscription = { messageLogAdded: MessageLogFragmentFragment };

export type MessageLogUpdatedSubscriptionVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type MessageLogUpdatedSubscription = { messageLogUpdated: MessageLogFragmentFragment };

export type ScreenshotAddedSubscriptionVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type ScreenshotAddedSubscription = { screenshotAdded: ScreenshotFragmentFragment };

export type AgentLogAddedSubscriptionVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type AgentLogAddedSubscription = { agentLogAdded: AgentLogFragmentFragment };

export type SearchLogAddedSubscriptionVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type SearchLogAddedSubscription = { searchLogAdded: SearchLogFragmentFragment };

export type VectorStoreLogAddedSubscriptionVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type VectorStoreLogAddedSubscription = { vectorStoreLogAdded: VectorStoreLogFragmentFragment };

export type FlowUpdatedSubscriptionVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type FlowUpdatedSubscription = {
    flowUpdated: { id: string; status: StatusType; updatedAt: any; terminals?: Array<TerminalFragmentFragment> | null };
};

export type TaskCreatedSubscriptionVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type TaskCreatedSubscription = { taskCreated: TaskFragmentFragment };

export type TaskUpdatedSubscriptionVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type TaskUpdatedSubscription = {
    taskUpdated: {
        id: string;
        status: StatusType;
        result: string;
        updatedAt: any;
        subtasks?: Array<SubtaskFragmentFragment> | null;
    };
};

export const FlowOverviewFragmentFragmentDoc = gql`
    fragment flowOverviewFragment on Flow {
        id
        title
        status
    }
`;
export const TerminalFragmentFragmentDoc = gql`
    fragment terminalFragment on Terminal {
        id
        type
        name
        image
        connected
        createdAt
    }
`;
export const FlowFragmentFragmentDoc = gql`
    fragment flowFragment on Flow {
        id
        title
        status
        terminals {
            ...terminalFragment
        }
        provider
        createdAt
        updatedAt
    }
`;
export const SubtaskFragmentFragmentDoc = gql`
    fragment subtaskFragment on Subtask {
        id
        status
        title
        description
        result
        taskId
        createdAt
        updatedAt
    }
`;
export const TaskFragmentFragmentDoc = gql`
    fragment taskFragment on Task {
        id
        title
        status
        input
        result
        flowId
        subtasks {
            ...subtaskFragment
        }
        createdAt
        updatedAt
    }
`;
export const TerminalLogFragmentFragmentDoc = gql`
    fragment terminalLogFragment on TerminalLog {
        id
        flowId
        type
        text
        terminal
        createdAt
    }
`;
export const MessageLogFragmentFragmentDoc = gql`
    fragment messageLogFragment on MessageLog {
        id
        type
        message
        result
        resultFormat
        flowId
        taskId
        subtaskId
        createdAt
    }
`;
export const ScreenshotFragmentFragmentDoc = gql`
    fragment screenshotFragment on Screenshot {
        id
        flowId
        name
        url
        createdAt
    }
`;
export const AgentLogFragmentFragmentDoc = gql`
    fragment agentLogFragment on AgentLog {
        id
        flowId
        initiator
        executor
        task
        result
        taskId
        subtaskId
        createdAt
    }
`;
export const SearchLogFragmentFragmentDoc = gql`
    fragment searchLogFragment on SearchLog {
        id
        flowId
        initiator
        executor
        engine
        query
        result
        taskId
        subtaskId
        createdAt
    }
`;
export const VectorStoreLogFragmentFragmentDoc = gql`
    fragment vectorStoreLogFragment on VectorStoreLog {
        id
        flowId
        initiator
        executor
        filter
        query
        action
        result
        taskId
        subtaskId
        createdAt
    }
`;
export const FlowsDocument = gql`
    query flows {
        flows {
            ...flowOverviewFragment
        }
    }
    ${FlowOverviewFragmentFragmentDoc}
`;

/**
 * __useFlowsQuery__
 *
 * To run a query within a React component, call `useFlowsQuery` and pass it any options that fit your needs.
 * When your component renders, `useFlowsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowsQuery({
 *   variables: {
 *   },
 * });
 */
export function useFlowsQuery(baseOptions?: Apollo.QueryHookOptions<FlowsQuery, FlowsQueryVariables>) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<FlowsQuery, FlowsQueryVariables>(FlowsDocument, options);
}
export function useFlowsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<FlowsQuery, FlowsQueryVariables>) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<FlowsQuery, FlowsQueryVariables>(FlowsDocument, options);
}
export function useFlowsSuspenseQuery(
    baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<FlowsQuery, FlowsQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<FlowsQuery, FlowsQueryVariables>(FlowsDocument, options);
}
export type FlowsQueryHookResult = ReturnType<typeof useFlowsQuery>;
export type FlowsLazyQueryHookResult = ReturnType<typeof useFlowsLazyQuery>;
export type FlowsSuspenseQueryHookResult = ReturnType<typeof useFlowsSuspenseQuery>;
export type FlowsQueryResult = Apollo.QueryResult<FlowsQuery, FlowsQueryVariables>;
export const ProvidersDocument = gql`
    query providers {
        providers
    }
`;

/**
 * __useProvidersQuery__
 *
 * To run a query within a React component, call `useProvidersQuery` and pass it any options that fit your needs.
 * When your component renders, `useProvidersQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useProvidersQuery({
 *   variables: {
 *   },
 * });
 */
export function useProvidersQuery(baseOptions?: Apollo.QueryHookOptions<ProvidersQuery, ProvidersQueryVariables>) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<ProvidersQuery, ProvidersQueryVariables>(ProvidersDocument, options);
}
export function useProvidersLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<ProvidersQuery, ProvidersQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<ProvidersQuery, ProvidersQueryVariables>(ProvidersDocument, options);
}
export function useProvidersSuspenseQuery(
    baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ProvidersQuery, ProvidersQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<ProvidersQuery, ProvidersQueryVariables>(ProvidersDocument, options);
}
export type ProvidersQueryHookResult = ReturnType<typeof useProvidersQuery>;
export type ProvidersLazyQueryHookResult = ReturnType<typeof useProvidersLazyQuery>;
export type ProvidersSuspenseQueryHookResult = ReturnType<typeof useProvidersSuspenseQuery>;
export type ProvidersQueryResult = Apollo.QueryResult<ProvidersQuery, ProvidersQueryVariables>;
export const FlowDocument = gql`
    query flow($id: ID!) {
        flow(flowId: $id) {
            ...flowFragment
        }
        tasks(flowId: $id) {
            ...taskFragment
        }
        screenshots(flowId: $id) {
            ...screenshotFragment
        }
        terminalLogs(flowId: $id) {
            ...terminalLogFragment
        }
        messageLogs(flowId: $id) {
            ...messageLogFragment
        }
        agentLogs(flowId: $id) {
            ...agentLogFragment
        }
        searchLogs(flowId: $id) {
            ...searchLogFragment
        }
        vectorStoreLogs(flowId: $id) {
            ...vectorStoreLogFragment
        }
    }
    ${FlowFragmentFragmentDoc}
    ${TerminalFragmentFragmentDoc}
    ${TaskFragmentFragmentDoc}
    ${SubtaskFragmentFragmentDoc}
    ${ScreenshotFragmentFragmentDoc}
    ${TerminalLogFragmentFragmentDoc}
    ${MessageLogFragmentFragmentDoc}
    ${AgentLogFragmentFragmentDoc}
    ${SearchLogFragmentFragmentDoc}
    ${VectorStoreLogFragmentFragmentDoc}
`;

/**
 * __useFlowQuery__
 *
 * To run a query within a React component, call `useFlowQuery` and pass it any options that fit your needs.
 * When your component renders, `useFlowQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useFlowQuery(
    baseOptions: Apollo.QueryHookOptions<FlowQuery, FlowQueryVariables> &
        ({ variables: FlowQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<FlowQuery, FlowQueryVariables>(FlowDocument, options);
}
export function useFlowLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<FlowQuery, FlowQueryVariables>) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<FlowQuery, FlowQueryVariables>(FlowDocument, options);
}
export function useFlowSuspenseQuery(
    baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<FlowQuery, FlowQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<FlowQuery, FlowQueryVariables>(FlowDocument, options);
}
export type FlowQueryHookResult = ReturnType<typeof useFlowQuery>;
export type FlowLazyQueryHookResult = ReturnType<typeof useFlowLazyQuery>;
export type FlowSuspenseQueryHookResult = ReturnType<typeof useFlowSuspenseQuery>;
export type FlowQueryResult = Apollo.QueryResult<FlowQuery, FlowQueryVariables>;
export const TasksDocument = gql`
    query tasks($flowId: ID!) {
        tasks(flowId: $flowId) {
            ...taskFragment
        }
    }
    ${TaskFragmentFragmentDoc}
    ${SubtaskFragmentFragmentDoc}
`;

/**
 * __useTasksQuery__
 *
 * To run a query within a React component, call `useTasksQuery` and pass it any options that fit your needs.
 * When your component renders, `useTasksQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useTasksQuery({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useTasksQuery(
    baseOptions: Apollo.QueryHookOptions<TasksQuery, TasksQueryVariables> &
        ({ variables: TasksQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<TasksQuery, TasksQueryVariables>(TasksDocument, options);
}
export function useTasksLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<TasksQuery, TasksQueryVariables>) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<TasksQuery, TasksQueryVariables>(TasksDocument, options);
}
export function useTasksSuspenseQuery(
    baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<TasksQuery, TasksQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<TasksQuery, TasksQueryVariables>(TasksDocument, options);
}
export type TasksQueryHookResult = ReturnType<typeof useTasksQuery>;
export type TasksLazyQueryHookResult = ReturnType<typeof useTasksLazyQuery>;
export type TasksSuspenseQueryHookResult = ReturnType<typeof useTasksSuspenseQuery>;
export type TasksQueryResult = Apollo.QueryResult<TasksQuery, TasksQueryVariables>;
export const CreateFlowDocument = gql`
    mutation createFlow($modelProvider: String!, $input: String!) {
        createFlow(modelProvider: $modelProvider, input: $input) {
            ...flowFragment
        }
    }
    ${FlowFragmentFragmentDoc}
    ${TerminalFragmentFragmentDoc}
`;
export type CreateFlowMutationFn = Apollo.MutationFunction<CreateFlowMutation, CreateFlowMutationVariables>;

/**
 * __useCreateFlowMutation__
 *
 * To run a mutation, you first call `useCreateFlowMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateFlowMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createFlowMutation, { data, loading, error }] = useCreateFlowMutation({
 *   variables: {
 *      modelProvider: // value for 'modelProvider'
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateFlowMutation(
    baseOptions?: Apollo.MutationHookOptions<CreateFlowMutation, CreateFlowMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<CreateFlowMutation, CreateFlowMutationVariables>(CreateFlowDocument, options);
}
export type CreateFlowMutationHookResult = ReturnType<typeof useCreateFlowMutation>;
export type CreateFlowMutationResult = Apollo.MutationResult<CreateFlowMutation>;
export type CreateFlowMutationOptions = Apollo.BaseMutationOptions<CreateFlowMutation, CreateFlowMutationVariables>;
export const DeleteFlowDocument = gql`
    mutation deleteFlow($flowId: ID!) {
        deleteFlow(flowId: $flowId)
    }
`;
export type DeleteFlowMutationFn = Apollo.MutationFunction<DeleteFlowMutation, DeleteFlowMutationVariables>;

/**
 * __useDeleteFlowMutation__
 *
 * To run a mutation, you first call `useDeleteFlowMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteFlowMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteFlowMutation, { data, loading, error }] = useDeleteFlowMutation({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useDeleteFlowMutation(
    baseOptions?: Apollo.MutationHookOptions<DeleteFlowMutation, DeleteFlowMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<DeleteFlowMutation, DeleteFlowMutationVariables>(DeleteFlowDocument, options);
}
export type DeleteFlowMutationHookResult = ReturnType<typeof useDeleteFlowMutation>;
export type DeleteFlowMutationResult = Apollo.MutationResult<DeleteFlowMutation>;
export type DeleteFlowMutationOptions = Apollo.BaseMutationOptions<DeleteFlowMutation, DeleteFlowMutationVariables>;
export const PutUserInputDocument = gql`
    mutation putUserInput($flowId: ID!, $input: String!) {
        putUserInput(flowId: $flowId, input: $input)
    }
`;
export type PutUserInputMutationFn = Apollo.MutationFunction<PutUserInputMutation, PutUserInputMutationVariables>;

/**
 * __usePutUserInputMutation__
 *
 * To run a mutation, you first call `usePutUserInputMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `usePutUserInputMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [putUserInputMutation, { data, loading, error }] = usePutUserInputMutation({
 *   variables: {
 *      flowId: // value for 'flowId'
 *      input: // value for 'input'
 *   },
 * });
 */
export function usePutUserInputMutation(
    baseOptions?: Apollo.MutationHookOptions<PutUserInputMutation, PutUserInputMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<PutUserInputMutation, PutUserInputMutationVariables>(PutUserInputDocument, options);
}
export type PutUserInputMutationHookResult = ReturnType<typeof usePutUserInputMutation>;
export type PutUserInputMutationResult = Apollo.MutationResult<PutUserInputMutation>;
export type PutUserInputMutationOptions = Apollo.BaseMutationOptions<
    PutUserInputMutation,
    PutUserInputMutationVariables
>;
export const FinishFlowDocument = gql`
    mutation finishFlow($flowId: ID!) {
        finishFlow(flowId: $flowId) {
            ...flowFragment
        }
    }
    ${FlowFragmentFragmentDoc}
    ${TerminalFragmentFragmentDoc}
`;
export type FinishFlowMutationFn = Apollo.MutationFunction<FinishFlowMutation, FinishFlowMutationVariables>;

/**
 * __useFinishFlowMutation__
 *
 * To run a mutation, you first call `useFinishFlowMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useFinishFlowMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [finishFlowMutation, { data, loading, error }] = useFinishFlowMutation({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useFinishFlowMutation(
    baseOptions?: Apollo.MutationHookOptions<FinishFlowMutation, FinishFlowMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<FinishFlowMutation, FinishFlowMutationVariables>(FinishFlowDocument, options);
}
export type FinishFlowMutationHookResult = ReturnType<typeof useFinishFlowMutation>;
export type FinishFlowMutationResult = Apollo.MutationResult<FinishFlowMutation>;
export type FinishFlowMutationOptions = Apollo.BaseMutationOptions<FinishFlowMutation, FinishFlowMutationVariables>;
export const TerminalLogAddedDocument = gql`
    subscription terminalLogAdded($flowId: ID!) {
        terminalLogAdded(flowId: $flowId) {
            ...terminalLogFragment
        }
    }
    ${TerminalLogFragmentFragmentDoc}
`;

/**
 * __useTerminalLogAddedSubscription__
 *
 * To run a query within a React component, call `useTerminalLogAddedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useTerminalLogAddedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useTerminalLogAddedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useTerminalLogAddedSubscription(
    baseOptions: Apollo.SubscriptionHookOptions<TerminalLogAddedSubscription, TerminalLogAddedSubscriptionVariables> &
        ({ variables: TerminalLogAddedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<TerminalLogAddedSubscription, TerminalLogAddedSubscriptionVariables>(
        TerminalLogAddedDocument,
        options,
    );
}
export type TerminalLogAddedSubscriptionHookResult = ReturnType<typeof useTerminalLogAddedSubscription>;
export type TerminalLogAddedSubscriptionResult = Apollo.SubscriptionResult<TerminalLogAddedSubscription>;
export const MessageLogAddedDocument = gql`
    subscription messageLogAdded($flowId: ID!) {
        messageLogAdded(flowId: $flowId) {
            ...messageLogFragment
        }
    }
    ${MessageLogFragmentFragmentDoc}
`;

/**
 * __useMessageLogAddedSubscription__
 *
 * To run a query within a React component, call `useMessageLogAddedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useMessageLogAddedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useMessageLogAddedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useMessageLogAddedSubscription(
    baseOptions: Apollo.SubscriptionHookOptions<MessageLogAddedSubscription, MessageLogAddedSubscriptionVariables> &
        ({ variables: MessageLogAddedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<MessageLogAddedSubscription, MessageLogAddedSubscriptionVariables>(
        MessageLogAddedDocument,
        options,
    );
}
export type MessageLogAddedSubscriptionHookResult = ReturnType<typeof useMessageLogAddedSubscription>;
export type MessageLogAddedSubscriptionResult = Apollo.SubscriptionResult<MessageLogAddedSubscription>;
export const MessageLogUpdatedDocument = gql`
    subscription messageLogUpdated($flowId: ID!) {
        messageLogUpdated(flowId: $flowId) {
            ...messageLogFragment
        }
    }
    ${MessageLogFragmentFragmentDoc}
`;

/**
 * __useMessageLogUpdatedSubscription__
 *
 * To run a query within a React component, call `useMessageLogUpdatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useMessageLogUpdatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useMessageLogUpdatedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useMessageLogUpdatedSubscription(
    baseOptions: Apollo.SubscriptionHookOptions<MessageLogUpdatedSubscription, MessageLogUpdatedSubscriptionVariables> &
        ({ variables: MessageLogUpdatedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<MessageLogUpdatedSubscription, MessageLogUpdatedSubscriptionVariables>(
        MessageLogUpdatedDocument,
        options,
    );
}
export type MessageLogUpdatedSubscriptionHookResult = ReturnType<typeof useMessageLogUpdatedSubscription>;
export type MessageLogUpdatedSubscriptionResult = Apollo.SubscriptionResult<MessageLogUpdatedSubscription>;
export const ScreenshotAddedDocument = gql`
    subscription screenshotAdded($flowId: ID!) {
        screenshotAdded(flowId: $flowId) {
            ...screenshotFragment
        }
    }
    ${ScreenshotFragmentFragmentDoc}
`;

/**
 * __useScreenshotAddedSubscription__
 *
 * To run a query within a React component, call `useScreenshotAddedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useScreenshotAddedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useScreenshotAddedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useScreenshotAddedSubscription(
    baseOptions: Apollo.SubscriptionHookOptions<ScreenshotAddedSubscription, ScreenshotAddedSubscriptionVariables> &
        ({ variables: ScreenshotAddedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<ScreenshotAddedSubscription, ScreenshotAddedSubscriptionVariables>(
        ScreenshotAddedDocument,
        options,
    );
}
export type ScreenshotAddedSubscriptionHookResult = ReturnType<typeof useScreenshotAddedSubscription>;
export type ScreenshotAddedSubscriptionResult = Apollo.SubscriptionResult<ScreenshotAddedSubscription>;
export const AgentLogAddedDocument = gql`
    subscription agentLogAdded($flowId: ID!) {
        agentLogAdded(flowId: $flowId) {
            ...agentLogFragment
        }
    }
    ${AgentLogFragmentFragmentDoc}
`;

/**
 * __useAgentLogAddedSubscription__
 *
 * To run a query within a React component, call `useAgentLogAddedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useAgentLogAddedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAgentLogAddedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useAgentLogAddedSubscription(
    baseOptions: Apollo.SubscriptionHookOptions<AgentLogAddedSubscription, AgentLogAddedSubscriptionVariables> &
        ({ variables: AgentLogAddedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<AgentLogAddedSubscription, AgentLogAddedSubscriptionVariables>(
        AgentLogAddedDocument,
        options,
    );
}
export type AgentLogAddedSubscriptionHookResult = ReturnType<typeof useAgentLogAddedSubscription>;
export type AgentLogAddedSubscriptionResult = Apollo.SubscriptionResult<AgentLogAddedSubscription>;
export const SearchLogAddedDocument = gql`
    subscription searchLogAdded($flowId: ID!) {
        searchLogAdded(flowId: $flowId) {
            ...searchLogFragment
        }
    }
    ${SearchLogFragmentFragmentDoc}
`;

/**
 * __useSearchLogAddedSubscription__
 *
 * To run a query within a React component, call `useSearchLogAddedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useSearchLogAddedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSearchLogAddedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useSearchLogAddedSubscription(
    baseOptions: Apollo.SubscriptionHookOptions<SearchLogAddedSubscription, SearchLogAddedSubscriptionVariables> &
        ({ variables: SearchLogAddedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<SearchLogAddedSubscription, SearchLogAddedSubscriptionVariables>(
        SearchLogAddedDocument,
        options,
    );
}
export type SearchLogAddedSubscriptionHookResult = ReturnType<typeof useSearchLogAddedSubscription>;
export type SearchLogAddedSubscriptionResult = Apollo.SubscriptionResult<SearchLogAddedSubscription>;
export const VectorStoreLogAddedDocument = gql`
    subscription vectorStoreLogAdded($flowId: ID!) {
        vectorStoreLogAdded(flowId: $flowId) {
            ...vectorStoreLogFragment
        }
    }
    ${VectorStoreLogFragmentFragmentDoc}
`;

/**
 * __useVectorStoreLogAddedSubscription__
 *
 * To run a query within a React component, call `useVectorStoreLogAddedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useVectorStoreLogAddedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useVectorStoreLogAddedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useVectorStoreLogAddedSubscription(
    baseOptions: Apollo.SubscriptionHookOptions<
        VectorStoreLogAddedSubscription,
        VectorStoreLogAddedSubscriptionVariables
    > &
        ({ variables: VectorStoreLogAddedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<VectorStoreLogAddedSubscription, VectorStoreLogAddedSubscriptionVariables>(
        VectorStoreLogAddedDocument,
        options,
    );
}
export type VectorStoreLogAddedSubscriptionHookResult = ReturnType<typeof useVectorStoreLogAddedSubscription>;
export type VectorStoreLogAddedSubscriptionResult = Apollo.SubscriptionResult<VectorStoreLogAddedSubscription>;
export const FlowUpdatedDocument = gql`
    subscription flowUpdated($flowId: ID!) {
        flowUpdated(flowId: $flowId) {
            id
            status
            terminals {
                ...terminalFragment
            }
            updatedAt
        }
    }
    ${TerminalFragmentFragmentDoc}
`;

/**
 * __useFlowUpdatedSubscription__
 *
 * To run a query within a React component, call `useFlowUpdatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useFlowUpdatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowUpdatedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useFlowUpdatedSubscription(
    baseOptions: Apollo.SubscriptionHookOptions<FlowUpdatedSubscription, FlowUpdatedSubscriptionVariables> &
        ({ variables: FlowUpdatedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<FlowUpdatedSubscription, FlowUpdatedSubscriptionVariables>(
        FlowUpdatedDocument,
        options,
    );
}
export type FlowUpdatedSubscriptionHookResult = ReturnType<typeof useFlowUpdatedSubscription>;
export type FlowUpdatedSubscriptionResult = Apollo.SubscriptionResult<FlowUpdatedSubscription>;
export const TaskCreatedDocument = gql`
    subscription taskCreated($flowId: ID!) {
        taskCreated(flowId: $flowId) {
            ...taskFragment
        }
    }
    ${TaskFragmentFragmentDoc}
    ${SubtaskFragmentFragmentDoc}
`;

/**
 * __useTaskCreatedSubscription__
 *
 * To run a query within a React component, call `useTaskCreatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useTaskCreatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useTaskCreatedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useTaskCreatedSubscription(
    baseOptions: Apollo.SubscriptionHookOptions<TaskCreatedSubscription, TaskCreatedSubscriptionVariables> &
        ({ variables: TaskCreatedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<TaskCreatedSubscription, TaskCreatedSubscriptionVariables>(
        TaskCreatedDocument,
        options,
    );
}
export type TaskCreatedSubscriptionHookResult = ReturnType<typeof useTaskCreatedSubscription>;
export type TaskCreatedSubscriptionResult = Apollo.SubscriptionResult<TaskCreatedSubscription>;
export const TaskUpdatedDocument = gql`
    subscription taskUpdated($flowId: ID!) {
        taskUpdated(flowId: $flowId) {
            id
            status
            result
            subtasks {
                ...subtaskFragment
            }
            updatedAt
        }
    }
    ${SubtaskFragmentFragmentDoc}
`;

/**
 * __useTaskUpdatedSubscription__
 *
 * To run a query within a React component, call `useTaskUpdatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useTaskUpdatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useTaskUpdatedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useTaskUpdatedSubscription(
    baseOptions: Apollo.SubscriptionHookOptions<TaskUpdatedSubscription, TaskUpdatedSubscriptionVariables> &
        ({ variables: TaskUpdatedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<TaskUpdatedSubscription, TaskUpdatedSubscriptionVariables>(
        TaskUpdatedDocument,
        options,
    );
}
export type TaskUpdatedSubscriptionHookResult = ReturnType<typeof useTaskUpdatedSubscription>;
export type TaskUpdatedSubscriptionResult = Apollo.SubscriptionResult<TaskUpdatedSubscription>;
