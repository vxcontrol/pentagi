import gql from 'graphql-tag';
import * as Urql from 'urql';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
export type Omit<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>;
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  Time: { input: any; output: any; }
};

export type AgentLog = {
  __typename?: 'AgentLog';
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
  Summarizer = 'summarizer'
}

export type Flow = {
  __typename?: 'Flow';
  createdAt: Scalars['Time']['output'];
  id: Scalars['ID']['output'];
  provider: Scalars['String']['output'];
  status: StatusType;
  terminals?: Maybe<Array<Terminal>>;
  title: Scalars['String']['output'];
  updatedAt: Scalars['Time']['output'];
};

export type MessageLog = {
  __typename?: 'MessageLog';
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
  Thoughts = 'thoughts'
}

export type Mutation = {
  __typename?: 'Mutation';
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
  __typename?: 'Prompt';
  prompt: Scalars['String']['output'];
  type: Scalars['String']['output'];
};

export type Query = {
  __typename?: 'Query';
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
  Terminal = 'terminal'
}

export enum ResultType {
  Error = 'error',
  Success = 'success'
}

export type Screenshot = {
  __typename?: 'Screenshot';
  createdAt: Scalars['Time']['output'];
  flowId: Scalars['ID']['output'];
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  url: Scalars['String']['output'];
};

export type SearchLog = {
  __typename?: 'SearchLog';
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
  Waiting = 'waiting'
}

export type Subscription = {
  __typename?: 'Subscription';
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
  __typename?: 'Subtask';
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
  __typename?: 'Task';
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
  __typename?: 'Terminal';
  connected: Scalars['Boolean']['output'];
  createdAt: Scalars['Time']['output'];
  id: Scalars['ID']['output'];
  image: Scalars['String']['output'];
  name: Scalars['String']['output'];
  type: TerminalType;
};

export type TerminalLog = {
  __typename?: 'TerminalLog';
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
  Stdout = 'stdout'
}

export enum TerminalType {
  Primary = 'primary',
  Secondary = 'secondary'
}

export enum VectorStoreAction {
  Retrieve = 'retrieve',
  Store = 'store'
}

export type VectorStoreLog = {
  __typename?: 'VectorStoreLog';
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
    ${TerminalFragmentFragmentDoc}`;
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
    ${SubtaskFragmentFragmentDoc}`;
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
    ${FlowOverviewFragmentFragmentDoc}`;

export function useFlowsQuery(options?: Omit<Urql.UseQueryArgs<FlowsQueryVariables>, 'query'>) {
  return Urql.useQuery<FlowsQuery, FlowsQueryVariables>({ query: FlowsDocument, ...options });
};
export const ProvidersDocument = gql`
    query providers {
  providers
}
    `;

export function useProvidersQuery(options?: Omit<Urql.UseQueryArgs<ProvidersQueryVariables>, 'query'>) {
  return Urql.useQuery<ProvidersQuery, ProvidersQueryVariables>({ query: ProvidersDocument, ...options });
};
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
${TaskFragmentFragmentDoc}
${ScreenshotFragmentFragmentDoc}
${TerminalLogFragmentFragmentDoc}
${MessageLogFragmentFragmentDoc}
${AgentLogFragmentFragmentDoc}
${SearchLogFragmentFragmentDoc}
${VectorStoreLogFragmentFragmentDoc}`;

export function useFlowQuery(options: Omit<Urql.UseQueryArgs<FlowQueryVariables>, 'query'>) {
  return Urql.useQuery<FlowQuery, FlowQueryVariables>({ query: FlowDocument, ...options });
};
export const TasksDocument = gql`
    query tasks($flowId: ID!) {
  tasks(flowId: $flowId) {
    ...taskFragment
  }
}
    ${TaskFragmentFragmentDoc}`;

export function useTasksQuery(options: Omit<Urql.UseQueryArgs<TasksQueryVariables>, 'query'>) {
  return Urql.useQuery<TasksQuery, TasksQueryVariables>({ query: TasksDocument, ...options });
};
export const CreateFlowDocument = gql`
    mutation createFlow($modelProvider: String!, $input: String!) {
  createFlow(modelProvider: $modelProvider, input: $input) {
    ...flowFragment
  }
}
    ${FlowFragmentFragmentDoc}`;

export function useCreateFlowMutation() {
  return Urql.useMutation<CreateFlowMutation, CreateFlowMutationVariables>(CreateFlowDocument);
};
export const PutUserInputDocument = gql`
    mutation putUserInput($flowId: ID!, $input: String!) {
  putUserInput(flowId: $flowId, input: $input)
}
    `;

export function usePutUserInputMutation() {
  return Urql.useMutation<PutUserInputMutation, PutUserInputMutationVariables>(PutUserInputDocument);
};
export const FinishFlowDocument = gql`
    mutation finishFlow($flowId: ID!) {
  finishFlow(flowId: $flowId) {
    ...flowFragment
  }
}
    ${FlowFragmentFragmentDoc}`;

export function useFinishFlowMutation() {
  return Urql.useMutation<FinishFlowMutation, FinishFlowMutationVariables>(FinishFlowDocument);
};
export const TerminalLogAddedDocument = gql`
    subscription terminalLogAdded($flowId: ID!) {
  terminalLogAdded(flowId: $flowId) {
    ...terminalLogFragment
  }
}
    ${TerminalLogFragmentFragmentDoc}`;

export function useTerminalLogAddedSubscription<TData = TerminalLogAddedSubscription>(options: Omit<Urql.UseSubscriptionArgs<TerminalLogAddedSubscriptionVariables>, 'query'>, handler?: Urql.SubscriptionHandler<TerminalLogAddedSubscription, TData>) {
  return Urql.useSubscription<TerminalLogAddedSubscription, TData, TerminalLogAddedSubscriptionVariables>({ query: TerminalLogAddedDocument, ...options }, handler);
};
export const MessageLogAddedDocument = gql`
    subscription messageLogAdded($flowId: ID!) {
  messageLogAdded(flowId: $flowId) {
    ...messageLogFragment
  }
}
    ${MessageLogFragmentFragmentDoc}`;

export function useMessageLogAddedSubscription<TData = MessageLogAddedSubscription>(options: Omit<Urql.UseSubscriptionArgs<MessageLogAddedSubscriptionVariables>, 'query'>, handler?: Urql.SubscriptionHandler<MessageLogAddedSubscription, TData>) {
  return Urql.useSubscription<MessageLogAddedSubscription, TData, MessageLogAddedSubscriptionVariables>({ query: MessageLogAddedDocument, ...options }, handler);
};
export const MessageLogUpdatedDocument = gql`
    subscription messageLogUpdated($flowId: ID!) {
  messageLogUpdated(flowId: $flowId) {
    ...messageLogFragment
  }
}
    ${MessageLogFragmentFragmentDoc}`;

export function useMessageLogUpdatedSubscription<TData = MessageLogUpdatedSubscription>(options: Omit<Urql.UseSubscriptionArgs<MessageLogUpdatedSubscriptionVariables>, 'query'>, handler?: Urql.SubscriptionHandler<MessageLogUpdatedSubscription, TData>) {
  return Urql.useSubscription<MessageLogUpdatedSubscription, TData, MessageLogUpdatedSubscriptionVariables>({ query: MessageLogUpdatedDocument, ...options }, handler);
};
export const ScreenshotAddedDocument = gql`
    subscription screenshotAdded($flowId: ID!) {
  screenshotAdded(flowId: $flowId) {
    ...screenshotFragment
  }
}
    ${ScreenshotFragmentFragmentDoc}`;

export function useScreenshotAddedSubscription<TData = ScreenshotAddedSubscription>(options: Omit<Urql.UseSubscriptionArgs<ScreenshotAddedSubscriptionVariables>, 'query'>, handler?: Urql.SubscriptionHandler<ScreenshotAddedSubscription, TData>) {
  return Urql.useSubscription<ScreenshotAddedSubscription, TData, ScreenshotAddedSubscriptionVariables>({ query: ScreenshotAddedDocument, ...options }, handler);
};
export const AgentLogAddedDocument = gql`
    subscription agentLogAdded($flowId: ID!) {
  agentLogAdded(flowId: $flowId) {
    ...agentLogFragment
  }
}
    ${AgentLogFragmentFragmentDoc}`;

export function useAgentLogAddedSubscription<TData = AgentLogAddedSubscription>(options: Omit<Urql.UseSubscriptionArgs<AgentLogAddedSubscriptionVariables>, 'query'>, handler?: Urql.SubscriptionHandler<AgentLogAddedSubscription, TData>) {
  return Urql.useSubscription<AgentLogAddedSubscription, TData, AgentLogAddedSubscriptionVariables>({ query: AgentLogAddedDocument, ...options }, handler);
};
export const SearchLogAddedDocument = gql`
    subscription searchLogAdded($flowId: ID!) {
  searchLogAdded(flowId: $flowId) {
    ...searchLogFragment
  }
}
    ${SearchLogFragmentFragmentDoc}`;

export function useSearchLogAddedSubscription<TData = SearchLogAddedSubscription>(options: Omit<Urql.UseSubscriptionArgs<SearchLogAddedSubscriptionVariables>, 'query'>, handler?: Urql.SubscriptionHandler<SearchLogAddedSubscription, TData>) {
  return Urql.useSubscription<SearchLogAddedSubscription, TData, SearchLogAddedSubscriptionVariables>({ query: SearchLogAddedDocument, ...options }, handler);
};
export const VectorStoreLogAddedDocument = gql`
    subscription vectorStoreLogAdded($flowId: ID!) {
  vectorStoreLogAdded(flowId: $flowId) {
    ...vectorStoreLogFragment
  }
}
    ${VectorStoreLogFragmentFragmentDoc}`;

export function useVectorStoreLogAddedSubscription<TData = VectorStoreLogAddedSubscription>(options: Omit<Urql.UseSubscriptionArgs<VectorStoreLogAddedSubscriptionVariables>, 'query'>, handler?: Urql.SubscriptionHandler<VectorStoreLogAddedSubscription, TData>) {
  return Urql.useSubscription<VectorStoreLogAddedSubscription, TData, VectorStoreLogAddedSubscriptionVariables>({ query: VectorStoreLogAddedDocument, ...options }, handler);
};
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
    ${TerminalFragmentFragmentDoc}`;

export function useFlowUpdatedSubscription<TData = FlowUpdatedSubscription>(options: Omit<Urql.UseSubscriptionArgs<FlowUpdatedSubscriptionVariables>, 'query'>, handler?: Urql.SubscriptionHandler<FlowUpdatedSubscription, TData>) {
  return Urql.useSubscription<FlowUpdatedSubscription, TData, FlowUpdatedSubscriptionVariables>({ query: FlowUpdatedDocument, ...options }, handler);
};
export const TaskCreatedDocument = gql`
    subscription taskCreated($flowId: ID!) {
  taskCreated(flowId: $flowId) {
    ...taskFragment
  }
}
    ${TaskFragmentFragmentDoc}`;

export function useTaskCreatedSubscription<TData = TaskCreatedSubscription>(options: Omit<Urql.UseSubscriptionArgs<TaskCreatedSubscriptionVariables>, 'query'>, handler?: Urql.SubscriptionHandler<TaskCreatedSubscription, TData>) {
  return Urql.useSubscription<TaskCreatedSubscription, TData, TaskCreatedSubscriptionVariables>({ query: TaskCreatedDocument, ...options }, handler);
};
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
    ${SubtaskFragmentFragmentDoc}`;

export function useTaskUpdatedSubscription<TData = TaskUpdatedSubscription>(options: Omit<Urql.UseSubscriptionArgs<TaskUpdatedSubscriptionVariables>, 'query'>, handler?: Urql.SubscriptionHandler<TaskUpdatedSubscription, TData>) {
  return Urql.useSubscription<TaskUpdatedSubscription, TData, TaskUpdatedSubscriptionVariables>({ query: TaskUpdatedDocument, ...options }, handler);
};
export type FlowOverviewFragmentFragment = { __typename?: 'Flow', id: string, title: string, status: StatusType };

export type FlowsQueryVariables = Exact<{ [key: string]: never; }>;


export type FlowsQuery = { __typename?: 'Query', flows?: Array<{ __typename?: 'Flow', id: string, title: string, status: StatusType }> | null };

export type ProvidersQueryVariables = Exact<{ [key: string]: never; }>;


export type ProvidersQuery = { __typename?: 'Query', providers: Array<string> };

export type FlowFragmentFragment = { __typename?: 'Flow', id: string, title: string, status: StatusType, provider: string, createdAt: any, updatedAt: any, terminals?: Array<{ __typename?: 'Terminal', id: string, type: TerminalType, name: string, image: string, connected: boolean, createdAt: any }> | null };

export type TerminalFragmentFragment = { __typename?: 'Terminal', id: string, type: TerminalType, name: string, image: string, connected: boolean, createdAt: any };

export type TaskFragmentFragment = { __typename?: 'Task', id: string, title: string, status: StatusType, input: string, result: string, flowId: string, createdAt: any, updatedAt: any, subtasks?: Array<{ __typename?: 'Subtask', id: string, status: StatusType, title: string, description: string, result: string, taskId: string, createdAt: any, updatedAt: any }> | null };

export type SubtaskFragmentFragment = { __typename?: 'Subtask', id: string, status: StatusType, title: string, description: string, result: string, taskId: string, createdAt: any, updatedAt: any };

export type TerminalLogFragmentFragment = { __typename?: 'TerminalLog', id: string, flowId: string, type: TerminalLogType, text: string, terminal: string, createdAt: any };

export type MessageLogFragmentFragment = { __typename?: 'MessageLog', id: string, type: MessageLogType, message: string, result: string, resultFormat: ResultFormat, flowId: string, taskId?: string | null, subtaskId?: string | null, createdAt: any };

export type ScreenshotFragmentFragment = { __typename?: 'Screenshot', id: string, flowId: string, name: string, url: string, createdAt: any };

export type AgentLogFragmentFragment = { __typename?: 'AgentLog', id: string, flowId: string, initiator: AgentType, executor: AgentType, task: string, result: string, taskId?: string | null, subtaskId?: string | null, createdAt: any };

export type SearchLogFragmentFragment = { __typename?: 'SearchLog', id: string, flowId: string, initiator: AgentType, executor: AgentType, engine: string, query: string, result: string, taskId?: string | null, subtaskId?: string | null, createdAt: any };

export type VectorStoreLogFragmentFragment = { __typename?: 'VectorStoreLog', id: string, flowId: string, initiator: AgentType, executor: AgentType, filter: string, query: string, action: VectorStoreAction, result: string, taskId?: string | null, subtaskId?: string | null, createdAt: any };

export type FlowQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type FlowQuery = { __typename?: 'Query', flow: { __typename?: 'Flow', id: string, title: string, status: StatusType, provider: string, createdAt: any, updatedAt: any, terminals?: Array<{ __typename?: 'Terminal', id: string, type: TerminalType, name: string, image: string, connected: boolean, createdAt: any }> | null }, tasks?: Array<{ __typename?: 'Task', id: string, title: string, status: StatusType, input: string, result: string, flowId: string, createdAt: any, updatedAt: any, subtasks?: Array<{ __typename?: 'Subtask', id: string, status: StatusType, title: string, description: string, result: string, taskId: string, createdAt: any, updatedAt: any }> | null }> | null, screenshots?: Array<{ __typename?: 'Screenshot', id: string, flowId: string, name: string, url: string, createdAt: any }> | null, terminalLogs?: Array<{ __typename?: 'TerminalLog', id: string, flowId: string, type: TerminalLogType, text: string, terminal: string, createdAt: any }> | null, messageLogs?: Array<{ __typename?: 'MessageLog', id: string, type: MessageLogType, message: string, result: string, resultFormat: ResultFormat, flowId: string, taskId?: string | null, subtaskId?: string | null, createdAt: any }> | null, agentLogs?: Array<{ __typename?: 'AgentLog', id: string, flowId: string, initiator: AgentType, executor: AgentType, task: string, result: string, taskId?: string | null, subtaskId?: string | null, createdAt: any }> | null, searchLogs?: Array<{ __typename?: 'SearchLog', id: string, flowId: string, initiator: AgentType, executor: AgentType, engine: string, query: string, result: string, taskId?: string | null, subtaskId?: string | null, createdAt: any }> | null, vectorStoreLogs?: Array<{ __typename?: 'VectorStoreLog', id: string, flowId: string, initiator: AgentType, executor: AgentType, filter: string, query: string, action: VectorStoreAction, result: string, taskId?: string | null, subtaskId?: string | null, createdAt: any }> | null };

export type TasksQueryVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type TasksQuery = { __typename?: 'Query', tasks?: Array<{ __typename?: 'Task', id: string, title: string, status: StatusType, input: string, result: string, flowId: string, createdAt: any, updatedAt: any, subtasks?: Array<{ __typename?: 'Subtask', id: string, status: StatusType, title: string, description: string, result: string, taskId: string, createdAt: any, updatedAt: any }> | null }> | null };

export type CreateFlowMutationVariables = Exact<{
  modelProvider: Scalars['String']['input'];
  input: Scalars['String']['input'];
}>;


export type CreateFlowMutation = { __typename?: 'Mutation', createFlow: { __typename?: 'Flow', id: string, title: string, status: StatusType, provider: string, createdAt: any, updatedAt: any, terminals?: Array<{ __typename?: 'Terminal', id: string, type: TerminalType, name: string, image: string, connected: boolean, createdAt: any }> | null } };

export type PutUserInputMutationVariables = Exact<{
  flowId: Scalars['ID']['input'];
  input: Scalars['String']['input'];
}>;


export type PutUserInputMutation = { __typename?: 'Mutation', putUserInput: ResultType };

export type FinishFlowMutationVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type FinishFlowMutation = { __typename?: 'Mutation', finishFlow: { __typename?: 'Flow', id: string, title: string, status: StatusType, provider: string, createdAt: any, updatedAt: any, terminals?: Array<{ __typename?: 'Terminal', id: string, type: TerminalType, name: string, image: string, connected: boolean, createdAt: any }> | null } };

export type TerminalLogAddedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type TerminalLogAddedSubscription = { __typename?: 'Subscription', terminalLogAdded: { __typename?: 'TerminalLog', id: string, flowId: string, type: TerminalLogType, text: string, terminal: string, createdAt: any } };

export type MessageLogAddedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type MessageLogAddedSubscription = { __typename?: 'Subscription', messageLogAdded: { __typename?: 'MessageLog', id: string, type: MessageLogType, message: string, result: string, resultFormat: ResultFormat, flowId: string, taskId?: string | null, subtaskId?: string | null, createdAt: any } };

export type MessageLogUpdatedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type MessageLogUpdatedSubscription = { __typename?: 'Subscription', messageLogUpdated: { __typename?: 'MessageLog', id: string, type: MessageLogType, message: string, result: string, resultFormat: ResultFormat, flowId: string, taskId?: string | null, subtaskId?: string | null, createdAt: any } };

export type ScreenshotAddedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type ScreenshotAddedSubscription = { __typename?: 'Subscription', screenshotAdded: { __typename?: 'Screenshot', id: string, flowId: string, name: string, url: string, createdAt: any } };

export type AgentLogAddedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type AgentLogAddedSubscription = { __typename?: 'Subscription', agentLogAdded: { __typename?: 'AgentLog', id: string, flowId: string, initiator: AgentType, executor: AgentType, task: string, result: string, taskId?: string | null, subtaskId?: string | null, createdAt: any } };

export type SearchLogAddedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type SearchLogAddedSubscription = { __typename?: 'Subscription', searchLogAdded: { __typename?: 'SearchLog', id: string, flowId: string, initiator: AgentType, executor: AgentType, engine: string, query: string, result: string, taskId?: string | null, subtaskId?: string | null, createdAt: any } };

export type VectorStoreLogAddedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type VectorStoreLogAddedSubscription = { __typename?: 'Subscription', vectorStoreLogAdded: { __typename?: 'VectorStoreLog', id: string, flowId: string, initiator: AgentType, executor: AgentType, filter: string, query: string, action: VectorStoreAction, result: string, taskId?: string | null, subtaskId?: string | null, createdAt: any } };

export type FlowUpdatedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type FlowUpdatedSubscription = { __typename?: 'Subscription', flowUpdated: { __typename?: 'Flow', id: string, status: StatusType, updatedAt: any, terminals?: Array<{ __typename?: 'Terminal', id: string, type: TerminalType, name: string, image: string, connected: boolean, createdAt: any }> | null } };

export type TaskCreatedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type TaskCreatedSubscription = { __typename?: 'Subscription', taskCreated: { __typename?: 'Task', id: string, title: string, status: StatusType, input: string, result: string, flowId: string, createdAt: any, updatedAt: any, subtasks?: Array<{ __typename?: 'Subtask', id: string, status: StatusType, title: string, description: string, result: string, taskId: string, createdAt: any, updatedAt: any }> | null } };

export type TaskUpdatedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type TaskUpdatedSubscription = { __typename?: 'Subscription', taskUpdated: { __typename?: 'Task', id: string, status: StatusType, result: string, updatedAt: any, subtasks?: Array<{ __typename?: 'Subtask', id: string, status: StatusType, title: string, description: string, result: string, taskId: string, createdAt: any, updatedAt: any }> | null } };

import { IntrospectionQuery } from 'graphql';
export default {
  "__schema": {
    "queryType": {
      "name": "Query"
    },
    "mutationType": {
      "name": "Mutation"
    },
    "subscriptionType": {
      "name": "Subscription"
    },
    "types": [
      {
        "kind": "OBJECT",
        "name": "AgentLog",
        "fields": [
          {
            "name": "createdAt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "executor",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "flowId",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "id",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "initiator",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "result",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "subtaskId",
            "type": {
              "kind": "SCALAR",
              "name": "Any"
            },
            "args": []
          },
          {
            "name": "task",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "taskId",
            "type": {
              "kind": "SCALAR",
              "name": "Any"
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "Flow",
        "fields": [
          {
            "name": "createdAt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "id",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "provider",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "status",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "terminals",
            "type": {
              "kind": "LIST",
              "ofType": {
                "kind": "NON_NULL",
                "ofType": {
                  "kind": "OBJECT",
                  "name": "Terminal",
                  "ofType": null
                }
              }
            },
            "args": []
          },
          {
            "name": "title",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "updatedAt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "MessageLog",
        "fields": [
          {
            "name": "createdAt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "flowId",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "id",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "message",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "result",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "resultFormat",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "subtaskId",
            "type": {
              "kind": "SCALAR",
              "name": "Any"
            },
            "args": []
          },
          {
            "name": "taskId",
            "type": {
              "kind": "SCALAR",
              "name": "Any"
            },
            "args": []
          },
          {
            "name": "type",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "Mutation",
        "fields": [
          {
            "name": "createFlow",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "Flow",
                "ofType": null
              }
            },
            "args": [
              {
                "name": "input",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              },
              {
                "name": "modelProvider",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "deleteFlow",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "finishFlow",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "Flow",
                "ofType": null
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "putUserInput",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              },
              {
                "name": "input",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "resetPrompt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": [
              {
                "name": "promptType",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "updatePrompt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": [
              {
                "name": "prompt",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              },
              {
                "name": "promptType",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "Prompt",
        "fields": [
          {
            "name": "prompt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "type",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "Query",
        "fields": [
          {
            "name": "agentLogs",
            "type": {
              "kind": "LIST",
              "ofType": {
                "kind": "NON_NULL",
                "ofType": {
                  "kind": "OBJECT",
                  "name": "AgentLog",
                  "ofType": null
                }
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "flow",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "Flow",
                "ofType": null
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "flows",
            "type": {
              "kind": "LIST",
              "ofType": {
                "kind": "NON_NULL",
                "ofType": {
                  "kind": "OBJECT",
                  "name": "Flow",
                  "ofType": null
                }
              }
            },
            "args": []
          },
          {
            "name": "messageLogs",
            "type": {
              "kind": "LIST",
              "ofType": {
                "kind": "NON_NULL",
                "ofType": {
                  "kind": "OBJECT",
                  "name": "MessageLog",
                  "ofType": null
                }
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "prompt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": [
              {
                "name": "promptType",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "prompts",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "LIST",
                "ofType": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "OBJECT",
                    "name": "Prompt",
                    "ofType": null
                  }
                }
              }
            },
            "args": []
          },
          {
            "name": "providers",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "LIST",
                "ofType": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            },
            "args": []
          },
          {
            "name": "screenshots",
            "type": {
              "kind": "LIST",
              "ofType": {
                "kind": "NON_NULL",
                "ofType": {
                  "kind": "OBJECT",
                  "name": "Screenshot",
                  "ofType": null
                }
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "searchLogs",
            "type": {
              "kind": "LIST",
              "ofType": {
                "kind": "NON_NULL",
                "ofType": {
                  "kind": "OBJECT",
                  "name": "SearchLog",
                  "ofType": null
                }
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "tasks",
            "type": {
              "kind": "LIST",
              "ofType": {
                "kind": "NON_NULL",
                "ofType": {
                  "kind": "OBJECT",
                  "name": "Task",
                  "ofType": null
                }
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "terminalLogs",
            "type": {
              "kind": "LIST",
              "ofType": {
                "kind": "NON_NULL",
                "ofType": {
                  "kind": "OBJECT",
                  "name": "TerminalLog",
                  "ofType": null
                }
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "vectorStoreLogs",
            "type": {
              "kind": "LIST",
              "ofType": {
                "kind": "NON_NULL",
                "ofType": {
                  "kind": "OBJECT",
                  "name": "VectorStoreLog",
                  "ofType": null
                }
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "Screenshot",
        "fields": [
          {
            "name": "createdAt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "flowId",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "id",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "name",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "url",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "SearchLog",
        "fields": [
          {
            "name": "createdAt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "engine",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "executor",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "flowId",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "id",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "initiator",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "query",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "result",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "subtaskId",
            "type": {
              "kind": "SCALAR",
              "name": "Any"
            },
            "args": []
          },
          {
            "name": "taskId",
            "type": {
              "kind": "SCALAR",
              "name": "Any"
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "Subscription",
        "fields": [
          {
            "name": "agentLogAdded",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "AgentLog",
                "ofType": null
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "flowCreated",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "Flow",
                "ofType": null
              }
            },
            "args": []
          },
          {
            "name": "flowDeleted",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "Flow",
                "ofType": null
              }
            },
            "args": []
          },
          {
            "name": "flowUpdated",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "Flow",
                "ofType": null
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "messageLogAdded",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "MessageLog",
                "ofType": null
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "messageLogUpdated",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "MessageLog",
                "ofType": null
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "screenshotAdded",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "Screenshot",
                "ofType": null
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "searchLogAdded",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "SearchLog",
                "ofType": null
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "taskCreated",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "Task",
                "ofType": null
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "taskUpdated",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "Task",
                "ofType": null
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "terminalLogAdded",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "TerminalLog",
                "ofType": null
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          },
          {
            "name": "vectorStoreLogAdded",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "OBJECT",
                "name": "VectorStoreLog",
                "ofType": null
              }
            },
            "args": [
              {
                "name": "flowId",
                "type": {
                  "kind": "NON_NULL",
                  "ofType": {
                    "kind": "SCALAR",
                    "name": "Any"
                  }
                }
              }
            ]
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "Subtask",
        "fields": [
          {
            "name": "createdAt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "description",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "id",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "result",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "status",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "taskId",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "title",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "updatedAt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "Task",
        "fields": [
          {
            "name": "createdAt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "flowId",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "id",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "input",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "result",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "status",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "subtasks",
            "type": {
              "kind": "LIST",
              "ofType": {
                "kind": "NON_NULL",
                "ofType": {
                  "kind": "OBJECT",
                  "name": "Subtask",
                  "ofType": null
                }
              }
            },
            "args": []
          },
          {
            "name": "title",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "updatedAt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "Terminal",
        "fields": [
          {
            "name": "connected",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "createdAt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "id",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "image",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "name",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "type",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "TerminalLog",
        "fields": [
          {
            "name": "createdAt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "flowId",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "id",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "terminal",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "text",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "type",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "OBJECT",
        "name": "VectorStoreLog",
        "fields": [
          {
            "name": "action",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "createdAt",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "executor",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "filter",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "flowId",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "id",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "initiator",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "query",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "result",
            "type": {
              "kind": "NON_NULL",
              "ofType": {
                "kind": "SCALAR",
                "name": "Any"
              }
            },
            "args": []
          },
          {
            "name": "subtaskId",
            "type": {
              "kind": "SCALAR",
              "name": "Any"
            },
            "args": []
          },
          {
            "name": "taskId",
            "type": {
              "kind": "SCALAR",
              "name": "Any"
            },
            "args": []
          }
        ],
        "interfaces": []
      },
      {
        "kind": "SCALAR",
        "name": "Any"
      }
    ],
    "directives": []
  }
} as unknown as IntrospectionQuery;