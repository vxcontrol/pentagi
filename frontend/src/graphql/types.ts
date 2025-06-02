import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
const defaultOptions = {} as const;
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
  Assistant = 'assistant',
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
  ToolCallFixer = 'tool_call_fixer'
}

export type Assistant = {
  createdAt: Scalars['Time']['output'];
  flowId: Scalars['ID']['output'];
  id: Scalars['ID']['output'];
  provider: Scalars['String']['output'];
  status: StatusType;
  title: Scalars['String']['output'];
  updatedAt: Scalars['Time']['output'];
  useAgents: Scalars['Boolean']['output'];
};

export type AssistantLog = {
  appendPart: Scalars['Boolean']['output'];
  assistantId: Scalars['ID']['output'];
  createdAt: Scalars['Time']['output'];
  flowId: Scalars['ID']['output'];
  id: Scalars['ID']['output'];
  message: Scalars['String']['output'];
  result: Scalars['String']['output'];
  resultFormat: ResultFormat;
  thinking?: Maybe<Scalars['String']['output']>;
  type: MessageLogType;
};

export type Flow = {
  createdAt: Scalars['Time']['output'];
  id: Scalars['ID']['output'];
  provider: Scalars['String']['output'];
  status: StatusType;
  terminals?: Maybe<Array<Terminal>>;
  title: Scalars['String']['output'];
  updatedAt: Scalars['Time']['output'];
};

export type FlowAssistant = {
  assistant: Assistant;
  flow: Flow;
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
  thinking?: Maybe<Scalars['String']['output']>;
  type: MessageLogType;
};

export enum MessageLogType {
  Advice = 'advice',
  Answer = 'answer',
  Ask = 'ask',
  Browser = 'browser',
  Done = 'done',
  File = 'file',
  Input = 'input',
  Report = 'report',
  Search = 'search',
  Terminal = 'terminal',
  Thoughts = 'thoughts'
}

export type Mutation = {
  callAssistant: ResultType;
  createAssistant: FlowAssistant;
  createFlow: Flow;
  deleteAssistant: ResultType;
  deleteFlow: ResultType;
  finishFlow: ResultType;
  putUserInput: ResultType;
  resetPrompt: ResultType;
  stopAssistant: Assistant;
  stopFlow: ResultType;
  updatePrompt: ResultType;
};


export type MutationCallAssistantArgs = {
  assistantId: Scalars['ID']['input'];
  flowId: Scalars['ID']['input'];
  input: Scalars['String']['input'];
  useAgents: Scalars['Boolean']['input'];
};


export type MutationCreateAssistantArgs = {
  flowId: Scalars['ID']['input'];
  input: Scalars['String']['input'];
  modelProvider: Scalars['String']['input'];
  useAgents: Scalars['Boolean']['input'];
};


export type MutationCreateFlowArgs = {
  input: Scalars['String']['input'];
  modelProvider: Scalars['String']['input'];
};


export type MutationDeleteAssistantArgs = {
  assistantId: Scalars['ID']['input'];
  flowId: Scalars['ID']['input'];
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


export type MutationStopAssistantArgs = {
  assistantId: Scalars['ID']['input'];
  flowId: Scalars['ID']['input'];
};


export type MutationStopFlowArgs = {
  flowId: Scalars['ID']['input'];
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
  assistantLogs?: Maybe<Array<AssistantLog>>;
  assistants?: Maybe<Array<Assistant>>;
  flow: Flow;
  flows?: Maybe<Array<Flow>>;
  messageLogs?: Maybe<Array<MessageLog>>;
  prompt: Scalars['String']['output'];
  prompts: Array<Prompt>;
  providers: Array<Scalars['String']['output']>;
  screenshots?: Maybe<Array<Screenshot>>;
  searchLogs?: Maybe<Array<SearchLog>>;
  settings: Settings;
  tasks?: Maybe<Array<Task>>;
  terminalLogs?: Maybe<Array<TerminalLog>>;
  vectorStoreLogs?: Maybe<Array<VectorStoreLog>>;
};


export type QueryAgentLogsArgs = {
  flowId: Scalars['ID']['input'];
};


export type QueryAssistantLogsArgs = {
  assistantId: Scalars['ID']['input'];
  flowId: Scalars['ID']['input'];
};


export type QueryAssistantsArgs = {
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

export type Settings = {
  askUser: Scalars['Boolean']['output'];
  assistantUseAgents: Scalars['Boolean']['output'];
  debug: Scalars['Boolean']['output'];
  dockerInside: Scalars['Boolean']['output'];
};

export enum StatusType {
  Created = 'created',
  Failed = 'failed',
  Finished = 'finished',
  Running = 'running',
  Waiting = 'waiting'
}

export type Subscription = {
  agentLogAdded: AgentLog;
  assistantCreated: Assistant;
  assistantDeleted: Assistant;
  assistantLogAdded: AssistantLog;
  assistantLogUpdated: AssistantLog;
  assistantUpdated: Assistant;
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


export type SubscriptionAssistantCreatedArgs = {
  flowId: Scalars['ID']['input'];
};


export type SubscriptionAssistantDeletedArgs = {
  flowId: Scalars['ID']['input'];
};


export type SubscriptionAssistantLogAddedArgs = {
  flowId: Scalars['ID']['input'];
};


export type SubscriptionAssistantLogUpdatedArgs = {
  flowId: Scalars['ID']['input'];
};


export type SubscriptionAssistantUpdatedArgs = {
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

export type FlowOverviewFragmentFragment = { id: string, title: string, status: StatusType };

export type FlowsQueryVariables = Exact<{ [key: string]: never; }>;


export type FlowsQuery = { flows?: Array<FlowOverviewFragmentFragment> | null };

export type ProvidersQueryVariables = Exact<{ [key: string]: never; }>;


export type ProvidersQuery = { providers: Array<string> };

export type SettingsFragmentFragment = { debug: boolean, askUser: boolean, dockerInside: boolean, assistantUseAgents: boolean };

export type SettingsQueryVariables = Exact<{ [key: string]: never; }>;


export type SettingsQuery = { settings: SettingsFragmentFragment };

export type FlowFragmentFragment = { id: string, title: string, status: StatusType, provider: string, createdAt: any, updatedAt: any, terminals?: Array<TerminalFragmentFragment> | null };

export type TerminalFragmentFragment = { id: string, type: TerminalType, name: string, image: string, connected: boolean, createdAt: any };

export type TaskFragmentFragment = { id: string, title: string, status: StatusType, input: string, result: string, flowId: string, createdAt: any, updatedAt: any, subtasks?: Array<SubtaskFragmentFragment> | null };

export type SubtaskFragmentFragment = { id: string, status: StatusType, title: string, description: string, result: string, taskId: string, createdAt: any, updatedAt: any };

export type TerminalLogFragmentFragment = { id: string, flowId: string, type: TerminalLogType, text: string, terminal: string, createdAt: any };

export type MessageLogFragmentFragment = { id: string, type: MessageLogType, message: string, thinking?: string | null, result: string, resultFormat: ResultFormat, flowId: string, taskId?: string | null, subtaskId?: string | null, createdAt: any };

export type ScreenshotFragmentFragment = { id: string, flowId: string, name: string, url: string, createdAt: any };

export type AgentLogFragmentFragment = { id: string, flowId: string, initiator: AgentType, executor: AgentType, task: string, result: string, taskId?: string | null, subtaskId?: string | null, createdAt: any };

export type SearchLogFragmentFragment = { id: string, flowId: string, initiator: AgentType, executor: AgentType, engine: string, query: string, result: string, taskId?: string | null, subtaskId?: string | null, createdAt: any };

export type VectorStoreLogFragmentFragment = { id: string, flowId: string, initiator: AgentType, executor: AgentType, filter: string, query: string, action: VectorStoreAction, result: string, taskId?: string | null, subtaskId?: string | null, createdAt: any };

export type AssistantFragmentFragment = { id: string, title: string, status: StatusType, provider: string, flowId: string, useAgents: boolean, createdAt: any, updatedAt: any };

export type AssistantLogFragmentFragment = { id: string, type: MessageLogType, message: string, thinking?: string | null, result: string, resultFormat: ResultFormat, appendPart: boolean, flowId: string, assistantId: string, createdAt: any };

export type PromptFragmentFragment = { type: string, prompt: string };

export type FlowQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type FlowQuery = { flow: FlowFragmentFragment, tasks?: Array<TaskFragmentFragment> | null, screenshots?: Array<ScreenshotFragmentFragment> | null, terminalLogs?: Array<TerminalLogFragmentFragment> | null, messageLogs?: Array<MessageLogFragmentFragment> | null, agentLogs?: Array<AgentLogFragmentFragment> | null, searchLogs?: Array<SearchLogFragmentFragment> | null, vectorStoreLogs?: Array<VectorStoreLogFragmentFragment> | null };

export type TasksQueryVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type TasksQuery = { tasks?: Array<TaskFragmentFragment> | null };

export type PromptsQueryVariables = Exact<{ [key: string]: never; }>;


export type PromptsQuery = { prompts: Array<PromptFragmentFragment> };

export type PromptQueryVariables = Exact<{
  promptType: Scalars['String']['input'];
}>;


export type PromptQuery = { prompt: string };

export type AssistantsQueryVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type AssistantsQuery = { assistants?: Array<AssistantFragmentFragment> | null };

export type AssistantLogsQueryVariables = Exact<{
  flowId: Scalars['ID']['input'];
  assistantId: Scalars['ID']['input'];
}>;


export type AssistantLogsQuery = { assistantLogs?: Array<AssistantLogFragmentFragment> | null };

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


export type FinishFlowMutation = { finishFlow: ResultType };

export type StopFlowMutationVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type StopFlowMutation = { stopFlow: ResultType };

export type CreateAssistantMutationVariables = Exact<{
  flowId: Scalars['ID']['input'];
  modelProvider: Scalars['String']['input'];
  input: Scalars['String']['input'];
  useAgents: Scalars['Boolean']['input'];
}>;


export type CreateAssistantMutation = { createAssistant: { flow: FlowFragmentFragment, assistant: AssistantFragmentFragment } };

export type CallAssistantMutationVariables = Exact<{
  flowId: Scalars['ID']['input'];
  assistantId: Scalars['ID']['input'];
  input: Scalars['String']['input'];
  useAgents: Scalars['Boolean']['input'];
}>;


export type CallAssistantMutation = { callAssistant: ResultType };

export type StopAssistantMutationVariables = Exact<{
  flowId: Scalars['ID']['input'];
  assistantId: Scalars['ID']['input'];
}>;


export type StopAssistantMutation = { stopAssistant: AssistantFragmentFragment };

export type DeleteAssistantMutationVariables = Exact<{
  flowId: Scalars['ID']['input'];
  assistantId: Scalars['ID']['input'];
}>;


export type DeleteAssistantMutation = { deleteAssistant: ResultType };

export type UpdatePromptMutationVariables = Exact<{
  promptType: Scalars['String']['input'];
  prompt: Scalars['String']['input'];
}>;


export type UpdatePromptMutation = { updatePrompt: ResultType };

export type ResetPromptMutationVariables = Exact<{
  promptType: Scalars['String']['input'];
}>;


export type ResetPromptMutation = { resetPrompt: ResultType };

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

export type AssistantCreatedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type AssistantCreatedSubscription = { assistantCreated: AssistantFragmentFragment };

export type AssistantUpdatedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type AssistantUpdatedSubscription = { assistantUpdated: AssistantFragmentFragment };

export type AssistantDeletedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type AssistantDeletedSubscription = { assistantDeleted: AssistantFragmentFragment };

export type AssistantLogAddedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type AssistantLogAddedSubscription = { assistantLogAdded: AssistantLogFragmentFragment };

export type AssistantLogUpdatedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type AssistantLogUpdatedSubscription = { assistantLogUpdated: AssistantLogFragmentFragment };

export type FlowCreatedSubscriptionVariables = Exact<{ [key: string]: never; }>;


export type FlowCreatedSubscription = { flowCreated: { id: string, title: string, status: StatusType, provider: string, createdAt: any, updatedAt: any, terminals?: Array<TerminalFragmentFragment> | null } };

export type FlowDeletedSubscriptionVariables = Exact<{ [key: string]: never; }>;


export type FlowDeletedSubscription = { flowDeleted: { id: string, status: StatusType, updatedAt: any } };

export type FlowUpdatedSubscriptionVariables = Exact<{ [key: string]: never; }>;


export type FlowUpdatedSubscription = { flowUpdated: { id: string, title: string, status: StatusType, updatedAt: any, terminals?: Array<TerminalFragmentFragment> | null } };

export type TaskCreatedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type TaskCreatedSubscription = { taskCreated: TaskFragmentFragment };

export type TaskUpdatedSubscriptionVariables = Exact<{
  flowId: Scalars['ID']['input'];
}>;


export type TaskUpdatedSubscription = { taskUpdated: { id: string, status: StatusType, result: string, updatedAt: any, subtasks?: Array<SubtaskFragmentFragment> | null } };

export type FlowReportQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type FlowReportQuery = { flow: FlowFragmentFragment, tasks?: Array<TaskFragmentFragment> | null };

export const FlowOverviewFragmentFragmentDoc = gql`
    fragment flowOverviewFragment on Flow {
  id
  title
  status
}
    `;
export const SettingsFragmentFragmentDoc = gql`
    fragment settingsFragment on Settings {
  debug
  askUser
  dockerInside
  assistantUseAgents
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
  thinking
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
export const AssistantFragmentFragmentDoc = gql`
    fragment assistantFragment on Assistant {
  id
  title
  status
  provider
  flowId
  useAgents
  createdAt
  updatedAt
}
    `;
export const AssistantLogFragmentFragmentDoc = gql`
    fragment assistantLogFragment on AssistantLog {
  id
  type
  message
  thinking
  result
  resultFormat
  appendPart
  flowId
  assistantId
  createdAt
}
    `;
export const PromptFragmentFragmentDoc = gql`
    fragment promptFragment on Prompt {
  type
  prompt
}
    `;
export const FlowsDocument = gql`
    query flows {
  flows {
    ...flowOverviewFragment
  }
}
    ${FlowOverviewFragmentFragmentDoc}`;

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
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<FlowsQuery, FlowsQueryVariables>(FlowsDocument, options);
      }
export function useFlowsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<FlowsQuery, FlowsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<FlowsQuery, FlowsQueryVariables>(FlowsDocument, options);
        }
export function useFlowsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<FlowsQuery, FlowsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
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
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ProvidersQuery, ProvidersQueryVariables>(ProvidersDocument, options);
      }
export function useProvidersLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ProvidersQuery, ProvidersQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ProvidersQuery, ProvidersQueryVariables>(ProvidersDocument, options);
        }
export function useProvidersSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ProvidersQuery, ProvidersQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<ProvidersQuery, ProvidersQueryVariables>(ProvidersDocument, options);
        }
export type ProvidersQueryHookResult = ReturnType<typeof useProvidersQuery>;
export type ProvidersLazyQueryHookResult = ReturnType<typeof useProvidersLazyQuery>;
export type ProvidersSuspenseQueryHookResult = ReturnType<typeof useProvidersSuspenseQuery>;
export type ProvidersQueryResult = Apollo.QueryResult<ProvidersQuery, ProvidersQueryVariables>;
export const SettingsDocument = gql`
    query settings {
  settings {
    ...settingsFragment
  }
}
    ${SettingsFragmentFragmentDoc}`;

/**
 * __useSettingsQuery__
 *
 * To run a query within a React component, call `useSettingsQuery` and pass it any options that fit your needs.
 * When your component renders, `useSettingsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSettingsQuery({
 *   variables: {
 *   },
 * });
 */
export function useSettingsQuery(baseOptions?: Apollo.QueryHookOptions<SettingsQuery, SettingsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<SettingsQuery, SettingsQueryVariables>(SettingsDocument, options);
      }
export function useSettingsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<SettingsQuery, SettingsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<SettingsQuery, SettingsQueryVariables>(SettingsDocument, options);
        }
export function useSettingsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<SettingsQuery, SettingsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<SettingsQuery, SettingsQueryVariables>(SettingsDocument, options);
        }
export type SettingsQueryHookResult = ReturnType<typeof useSettingsQuery>;
export type SettingsLazyQueryHookResult = ReturnType<typeof useSettingsLazyQuery>;
export type SettingsSuspenseQueryHookResult = ReturnType<typeof useSettingsSuspenseQuery>;
export type SettingsQueryResult = Apollo.QueryResult<SettingsQuery, SettingsQueryVariables>;
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
${VectorStoreLogFragmentFragmentDoc}`;

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
export function useFlowQuery(baseOptions: Apollo.QueryHookOptions<FlowQuery, FlowQueryVariables> & ({ variables: FlowQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<FlowQuery, FlowQueryVariables>(FlowDocument, options);
      }
export function useFlowLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<FlowQuery, FlowQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<FlowQuery, FlowQueryVariables>(FlowDocument, options);
        }
export function useFlowSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<FlowQuery, FlowQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
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
${SubtaskFragmentFragmentDoc}`;

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
export function useTasksQuery(baseOptions: Apollo.QueryHookOptions<TasksQuery, TasksQueryVariables> & ({ variables: TasksQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<TasksQuery, TasksQueryVariables>(TasksDocument, options);
      }
export function useTasksLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<TasksQuery, TasksQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<TasksQuery, TasksQueryVariables>(TasksDocument, options);
        }
export function useTasksSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<TasksQuery, TasksQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<TasksQuery, TasksQueryVariables>(TasksDocument, options);
        }
export type TasksQueryHookResult = ReturnType<typeof useTasksQuery>;
export type TasksLazyQueryHookResult = ReturnType<typeof useTasksLazyQuery>;
export type TasksSuspenseQueryHookResult = ReturnType<typeof useTasksSuspenseQuery>;
export type TasksQueryResult = Apollo.QueryResult<TasksQuery, TasksQueryVariables>;
export const PromptsDocument = gql`
    query prompts {
  prompts {
    ...promptFragment
  }
}
    ${PromptFragmentFragmentDoc}`;

/**
 * __usePromptsQuery__
 *
 * To run a query within a React component, call `usePromptsQuery` and pass it any options that fit your needs.
 * When your component renders, `usePromptsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = usePromptsQuery({
 *   variables: {
 *   },
 * });
 */
export function usePromptsQuery(baseOptions?: Apollo.QueryHookOptions<PromptsQuery, PromptsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<PromptsQuery, PromptsQueryVariables>(PromptsDocument, options);
      }
export function usePromptsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<PromptsQuery, PromptsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<PromptsQuery, PromptsQueryVariables>(PromptsDocument, options);
        }
export function usePromptsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<PromptsQuery, PromptsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<PromptsQuery, PromptsQueryVariables>(PromptsDocument, options);
        }
export type PromptsQueryHookResult = ReturnType<typeof usePromptsQuery>;
export type PromptsLazyQueryHookResult = ReturnType<typeof usePromptsLazyQuery>;
export type PromptsSuspenseQueryHookResult = ReturnType<typeof usePromptsSuspenseQuery>;
export type PromptsQueryResult = Apollo.QueryResult<PromptsQuery, PromptsQueryVariables>;
export const PromptDocument = gql`
    query prompt($promptType: String!) {
  prompt(promptType: $promptType)
}
    `;

/**
 * __usePromptQuery__
 *
 * To run a query within a React component, call `usePromptQuery` and pass it any options that fit your needs.
 * When your component renders, `usePromptQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = usePromptQuery({
 *   variables: {
 *      promptType: // value for 'promptType'
 *   },
 * });
 */
export function usePromptQuery(baseOptions: Apollo.QueryHookOptions<PromptQuery, PromptQueryVariables> & ({ variables: PromptQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<PromptQuery, PromptQueryVariables>(PromptDocument, options);
      }
export function usePromptLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<PromptQuery, PromptQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<PromptQuery, PromptQueryVariables>(PromptDocument, options);
        }
export function usePromptSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<PromptQuery, PromptQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<PromptQuery, PromptQueryVariables>(PromptDocument, options);
        }
export type PromptQueryHookResult = ReturnType<typeof usePromptQuery>;
export type PromptLazyQueryHookResult = ReturnType<typeof usePromptLazyQuery>;
export type PromptSuspenseQueryHookResult = ReturnType<typeof usePromptSuspenseQuery>;
export type PromptQueryResult = Apollo.QueryResult<PromptQuery, PromptQueryVariables>;
export const AssistantsDocument = gql`
    query assistants($flowId: ID!) {
  assistants(flowId: $flowId) {
    ...assistantFragment
  }
}
    ${AssistantFragmentFragmentDoc}`;

/**
 * __useAssistantsQuery__
 *
 * To run a query within a React component, call `useAssistantsQuery` and pass it any options that fit your needs.
 * When your component renders, `useAssistantsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAssistantsQuery({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useAssistantsQuery(baseOptions: Apollo.QueryHookOptions<AssistantsQuery, AssistantsQueryVariables> & ({ variables: AssistantsQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AssistantsQuery, AssistantsQueryVariables>(AssistantsDocument, options);
      }
export function useAssistantsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AssistantsQuery, AssistantsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AssistantsQuery, AssistantsQueryVariables>(AssistantsDocument, options);
        }
export function useAssistantsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AssistantsQuery, AssistantsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<AssistantsQuery, AssistantsQueryVariables>(AssistantsDocument, options);
        }
export type AssistantsQueryHookResult = ReturnType<typeof useAssistantsQuery>;
export type AssistantsLazyQueryHookResult = ReturnType<typeof useAssistantsLazyQuery>;
export type AssistantsSuspenseQueryHookResult = ReturnType<typeof useAssistantsSuspenseQuery>;
export type AssistantsQueryResult = Apollo.QueryResult<AssistantsQuery, AssistantsQueryVariables>;
export const AssistantLogsDocument = gql`
    query assistantLogs($flowId: ID!, $assistantId: ID!) {
  assistantLogs(flowId: $flowId, assistantId: $assistantId) {
    ...assistantLogFragment
  }
}
    ${AssistantLogFragmentFragmentDoc}`;

/**
 * __useAssistantLogsQuery__
 *
 * To run a query within a React component, call `useAssistantLogsQuery` and pass it any options that fit your needs.
 * When your component renders, `useAssistantLogsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAssistantLogsQuery({
 *   variables: {
 *      flowId: // value for 'flowId'
 *      assistantId: // value for 'assistantId'
 *   },
 * });
 */
export function useAssistantLogsQuery(baseOptions: Apollo.QueryHookOptions<AssistantLogsQuery, AssistantLogsQueryVariables> & ({ variables: AssistantLogsQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AssistantLogsQuery, AssistantLogsQueryVariables>(AssistantLogsDocument, options);
      }
export function useAssistantLogsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AssistantLogsQuery, AssistantLogsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AssistantLogsQuery, AssistantLogsQueryVariables>(AssistantLogsDocument, options);
        }
export function useAssistantLogsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AssistantLogsQuery, AssistantLogsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<AssistantLogsQuery, AssistantLogsQueryVariables>(AssistantLogsDocument, options);
        }
export type AssistantLogsQueryHookResult = ReturnType<typeof useAssistantLogsQuery>;
export type AssistantLogsLazyQueryHookResult = ReturnType<typeof useAssistantLogsLazyQuery>;
export type AssistantLogsSuspenseQueryHookResult = ReturnType<typeof useAssistantLogsSuspenseQuery>;
export type AssistantLogsQueryResult = Apollo.QueryResult<AssistantLogsQuery, AssistantLogsQueryVariables>;
export const CreateFlowDocument = gql`
    mutation createFlow($modelProvider: String!, $input: String!) {
  createFlow(modelProvider: $modelProvider, input: $input) {
    ...flowFragment
  }
}
    ${FlowFragmentFragmentDoc}
${TerminalFragmentFragmentDoc}`;
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
export function useCreateFlowMutation(baseOptions?: Apollo.MutationHookOptions<CreateFlowMutation, CreateFlowMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
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
export function useDeleteFlowMutation(baseOptions?: Apollo.MutationHookOptions<DeleteFlowMutation, DeleteFlowMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
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
export function usePutUserInputMutation(baseOptions?: Apollo.MutationHookOptions<PutUserInputMutation, PutUserInputMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<PutUserInputMutation, PutUserInputMutationVariables>(PutUserInputDocument, options);
      }
export type PutUserInputMutationHookResult = ReturnType<typeof usePutUserInputMutation>;
export type PutUserInputMutationResult = Apollo.MutationResult<PutUserInputMutation>;
export type PutUserInputMutationOptions = Apollo.BaseMutationOptions<PutUserInputMutation, PutUserInputMutationVariables>;
export const FinishFlowDocument = gql`
    mutation finishFlow($flowId: ID!) {
  finishFlow(flowId: $flowId)
}
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
export function useFinishFlowMutation(baseOptions?: Apollo.MutationHookOptions<FinishFlowMutation, FinishFlowMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<FinishFlowMutation, FinishFlowMutationVariables>(FinishFlowDocument, options);
      }
export type FinishFlowMutationHookResult = ReturnType<typeof useFinishFlowMutation>;
export type FinishFlowMutationResult = Apollo.MutationResult<FinishFlowMutation>;
export type FinishFlowMutationOptions = Apollo.BaseMutationOptions<FinishFlowMutation, FinishFlowMutationVariables>;
export const StopFlowDocument = gql`
    mutation stopFlow($flowId: ID!) {
  stopFlow(flowId: $flowId)
}
    `;
export type StopFlowMutationFn = Apollo.MutationFunction<StopFlowMutation, StopFlowMutationVariables>;

/**
 * __useStopFlowMutation__
 *
 * To run a mutation, you first call `useStopFlowMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useStopFlowMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [stopFlowMutation, { data, loading, error }] = useStopFlowMutation({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useStopFlowMutation(baseOptions?: Apollo.MutationHookOptions<StopFlowMutation, StopFlowMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<StopFlowMutation, StopFlowMutationVariables>(StopFlowDocument, options);
      }
export type StopFlowMutationHookResult = ReturnType<typeof useStopFlowMutation>;
export type StopFlowMutationResult = Apollo.MutationResult<StopFlowMutation>;
export type StopFlowMutationOptions = Apollo.BaseMutationOptions<StopFlowMutation, StopFlowMutationVariables>;
export const CreateAssistantDocument = gql`
    mutation createAssistant($flowId: ID!, $modelProvider: String!, $input: String!, $useAgents: Boolean!) {
  createAssistant(
    flowId: $flowId
    modelProvider: $modelProvider
    input: $input
    useAgents: $useAgents
  ) {
    flow {
      ...flowFragment
    }
    assistant {
      ...assistantFragment
    }
  }
}
    ${FlowFragmentFragmentDoc}
${TerminalFragmentFragmentDoc}
${AssistantFragmentFragmentDoc}`;
export type CreateAssistantMutationFn = Apollo.MutationFunction<CreateAssistantMutation, CreateAssistantMutationVariables>;

/**
 * __useCreateAssistantMutation__
 *
 * To run a mutation, you first call `useCreateAssistantMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateAssistantMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createAssistantMutation, { data, loading, error }] = useCreateAssistantMutation({
 *   variables: {
 *      flowId: // value for 'flowId'
 *      modelProvider: // value for 'modelProvider'
 *      input: // value for 'input'
 *      useAgents: // value for 'useAgents'
 *   },
 * });
 */
export function useCreateAssistantMutation(baseOptions?: Apollo.MutationHookOptions<CreateAssistantMutation, CreateAssistantMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateAssistantMutation, CreateAssistantMutationVariables>(CreateAssistantDocument, options);
      }
export type CreateAssistantMutationHookResult = ReturnType<typeof useCreateAssistantMutation>;
export type CreateAssistantMutationResult = Apollo.MutationResult<CreateAssistantMutation>;
export type CreateAssistantMutationOptions = Apollo.BaseMutationOptions<CreateAssistantMutation, CreateAssistantMutationVariables>;
export const CallAssistantDocument = gql`
    mutation callAssistant($flowId: ID!, $assistantId: ID!, $input: String!, $useAgents: Boolean!) {
  callAssistant(
    flowId: $flowId
    assistantId: $assistantId
    input: $input
    useAgents: $useAgents
  )
}
    `;
export type CallAssistantMutationFn = Apollo.MutationFunction<CallAssistantMutation, CallAssistantMutationVariables>;

/**
 * __useCallAssistantMutation__
 *
 * To run a mutation, you first call `useCallAssistantMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCallAssistantMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [callAssistantMutation, { data, loading, error }] = useCallAssistantMutation({
 *   variables: {
 *      flowId: // value for 'flowId'
 *      assistantId: // value for 'assistantId'
 *      input: // value for 'input'
 *      useAgents: // value for 'useAgents'
 *   },
 * });
 */
export function useCallAssistantMutation(baseOptions?: Apollo.MutationHookOptions<CallAssistantMutation, CallAssistantMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CallAssistantMutation, CallAssistantMutationVariables>(CallAssistantDocument, options);
      }
export type CallAssistantMutationHookResult = ReturnType<typeof useCallAssistantMutation>;
export type CallAssistantMutationResult = Apollo.MutationResult<CallAssistantMutation>;
export type CallAssistantMutationOptions = Apollo.BaseMutationOptions<CallAssistantMutation, CallAssistantMutationVariables>;
export const StopAssistantDocument = gql`
    mutation stopAssistant($flowId: ID!, $assistantId: ID!) {
  stopAssistant(flowId: $flowId, assistantId: $assistantId) {
    ...assistantFragment
  }
}
    ${AssistantFragmentFragmentDoc}`;
export type StopAssistantMutationFn = Apollo.MutationFunction<StopAssistantMutation, StopAssistantMutationVariables>;

/**
 * __useStopAssistantMutation__
 *
 * To run a mutation, you first call `useStopAssistantMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useStopAssistantMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [stopAssistantMutation, { data, loading, error }] = useStopAssistantMutation({
 *   variables: {
 *      flowId: // value for 'flowId'
 *      assistantId: // value for 'assistantId'
 *   },
 * });
 */
export function useStopAssistantMutation(baseOptions?: Apollo.MutationHookOptions<StopAssistantMutation, StopAssistantMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<StopAssistantMutation, StopAssistantMutationVariables>(StopAssistantDocument, options);
      }
export type StopAssistantMutationHookResult = ReturnType<typeof useStopAssistantMutation>;
export type StopAssistantMutationResult = Apollo.MutationResult<StopAssistantMutation>;
export type StopAssistantMutationOptions = Apollo.BaseMutationOptions<StopAssistantMutation, StopAssistantMutationVariables>;
export const DeleteAssistantDocument = gql`
    mutation deleteAssistant($flowId: ID!, $assistantId: ID!) {
  deleteAssistant(flowId: $flowId, assistantId: $assistantId)
}
    `;
export type DeleteAssistantMutationFn = Apollo.MutationFunction<DeleteAssistantMutation, DeleteAssistantMutationVariables>;

/**
 * __useDeleteAssistantMutation__
 *
 * To run a mutation, you first call `useDeleteAssistantMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteAssistantMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteAssistantMutation, { data, loading, error }] = useDeleteAssistantMutation({
 *   variables: {
 *      flowId: // value for 'flowId'
 *      assistantId: // value for 'assistantId'
 *   },
 * });
 */
export function useDeleteAssistantMutation(baseOptions?: Apollo.MutationHookOptions<DeleteAssistantMutation, DeleteAssistantMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteAssistantMutation, DeleteAssistantMutationVariables>(DeleteAssistantDocument, options);
      }
export type DeleteAssistantMutationHookResult = ReturnType<typeof useDeleteAssistantMutation>;
export type DeleteAssistantMutationResult = Apollo.MutationResult<DeleteAssistantMutation>;
export type DeleteAssistantMutationOptions = Apollo.BaseMutationOptions<DeleteAssistantMutation, DeleteAssistantMutationVariables>;
export const UpdatePromptDocument = gql`
    mutation updatePrompt($promptType: String!, $prompt: String!) {
  updatePrompt(promptType: $promptType, prompt: $prompt)
}
    `;
export type UpdatePromptMutationFn = Apollo.MutationFunction<UpdatePromptMutation, UpdatePromptMutationVariables>;

/**
 * __useUpdatePromptMutation__
 *
 * To run a mutation, you first call `useUpdatePromptMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdatePromptMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updatePromptMutation, { data, loading, error }] = useUpdatePromptMutation({
 *   variables: {
 *      promptType: // value for 'promptType'
 *      prompt: // value for 'prompt'
 *   },
 * });
 */
export function useUpdatePromptMutation(baseOptions?: Apollo.MutationHookOptions<UpdatePromptMutation, UpdatePromptMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdatePromptMutation, UpdatePromptMutationVariables>(UpdatePromptDocument, options);
      }
export type UpdatePromptMutationHookResult = ReturnType<typeof useUpdatePromptMutation>;
export type UpdatePromptMutationResult = Apollo.MutationResult<UpdatePromptMutation>;
export type UpdatePromptMutationOptions = Apollo.BaseMutationOptions<UpdatePromptMutation, UpdatePromptMutationVariables>;
export const ResetPromptDocument = gql`
    mutation resetPrompt($promptType: String!) {
  resetPrompt(promptType: $promptType)
}
    `;
export type ResetPromptMutationFn = Apollo.MutationFunction<ResetPromptMutation, ResetPromptMutationVariables>;

/**
 * __useResetPromptMutation__
 *
 * To run a mutation, you first call `useResetPromptMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useResetPromptMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [resetPromptMutation, { data, loading, error }] = useResetPromptMutation({
 *   variables: {
 *      promptType: // value for 'promptType'
 *   },
 * });
 */
export function useResetPromptMutation(baseOptions?: Apollo.MutationHookOptions<ResetPromptMutation, ResetPromptMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ResetPromptMutation, ResetPromptMutationVariables>(ResetPromptDocument, options);
      }
export type ResetPromptMutationHookResult = ReturnType<typeof useResetPromptMutation>;
export type ResetPromptMutationResult = Apollo.MutationResult<ResetPromptMutation>;
export type ResetPromptMutationOptions = Apollo.BaseMutationOptions<ResetPromptMutation, ResetPromptMutationVariables>;
export const TerminalLogAddedDocument = gql`
    subscription terminalLogAdded($flowId: ID!) {
  terminalLogAdded(flowId: $flowId) {
    ...terminalLogFragment
  }
}
    ${TerminalLogFragmentFragmentDoc}`;

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
export function useTerminalLogAddedSubscription(baseOptions: Apollo.SubscriptionHookOptions<TerminalLogAddedSubscription, TerminalLogAddedSubscriptionVariables> & ({ variables: TerminalLogAddedSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<TerminalLogAddedSubscription, TerminalLogAddedSubscriptionVariables>(TerminalLogAddedDocument, options);
      }
export type TerminalLogAddedSubscriptionHookResult = ReturnType<typeof useTerminalLogAddedSubscription>;
export type TerminalLogAddedSubscriptionResult = Apollo.SubscriptionResult<TerminalLogAddedSubscription>;
export const MessageLogAddedDocument = gql`
    subscription messageLogAdded($flowId: ID!) {
  messageLogAdded(flowId: $flowId) {
    ...messageLogFragment
  }
}
    ${MessageLogFragmentFragmentDoc}`;

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
export function useMessageLogAddedSubscription(baseOptions: Apollo.SubscriptionHookOptions<MessageLogAddedSubscription, MessageLogAddedSubscriptionVariables> & ({ variables: MessageLogAddedSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<MessageLogAddedSubscription, MessageLogAddedSubscriptionVariables>(MessageLogAddedDocument, options);
      }
export type MessageLogAddedSubscriptionHookResult = ReturnType<typeof useMessageLogAddedSubscription>;
export type MessageLogAddedSubscriptionResult = Apollo.SubscriptionResult<MessageLogAddedSubscription>;
export const MessageLogUpdatedDocument = gql`
    subscription messageLogUpdated($flowId: ID!) {
  messageLogUpdated(flowId: $flowId) {
    ...messageLogFragment
  }
}
    ${MessageLogFragmentFragmentDoc}`;

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
export function useMessageLogUpdatedSubscription(baseOptions: Apollo.SubscriptionHookOptions<MessageLogUpdatedSubscription, MessageLogUpdatedSubscriptionVariables> & ({ variables: MessageLogUpdatedSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<MessageLogUpdatedSubscription, MessageLogUpdatedSubscriptionVariables>(MessageLogUpdatedDocument, options);
      }
export type MessageLogUpdatedSubscriptionHookResult = ReturnType<typeof useMessageLogUpdatedSubscription>;
export type MessageLogUpdatedSubscriptionResult = Apollo.SubscriptionResult<MessageLogUpdatedSubscription>;
export const ScreenshotAddedDocument = gql`
    subscription screenshotAdded($flowId: ID!) {
  screenshotAdded(flowId: $flowId) {
    ...screenshotFragment
  }
}
    ${ScreenshotFragmentFragmentDoc}`;

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
export function useScreenshotAddedSubscription(baseOptions: Apollo.SubscriptionHookOptions<ScreenshotAddedSubscription, ScreenshotAddedSubscriptionVariables> & ({ variables: ScreenshotAddedSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<ScreenshotAddedSubscription, ScreenshotAddedSubscriptionVariables>(ScreenshotAddedDocument, options);
      }
export type ScreenshotAddedSubscriptionHookResult = ReturnType<typeof useScreenshotAddedSubscription>;
export type ScreenshotAddedSubscriptionResult = Apollo.SubscriptionResult<ScreenshotAddedSubscription>;
export const AgentLogAddedDocument = gql`
    subscription agentLogAdded($flowId: ID!) {
  agentLogAdded(flowId: $flowId) {
    ...agentLogFragment
  }
}
    ${AgentLogFragmentFragmentDoc}`;

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
export function useAgentLogAddedSubscription(baseOptions: Apollo.SubscriptionHookOptions<AgentLogAddedSubscription, AgentLogAddedSubscriptionVariables> & ({ variables: AgentLogAddedSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<AgentLogAddedSubscription, AgentLogAddedSubscriptionVariables>(AgentLogAddedDocument, options);
      }
export type AgentLogAddedSubscriptionHookResult = ReturnType<typeof useAgentLogAddedSubscription>;
export type AgentLogAddedSubscriptionResult = Apollo.SubscriptionResult<AgentLogAddedSubscription>;
export const SearchLogAddedDocument = gql`
    subscription searchLogAdded($flowId: ID!) {
  searchLogAdded(flowId: $flowId) {
    ...searchLogFragment
  }
}
    ${SearchLogFragmentFragmentDoc}`;

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
export function useSearchLogAddedSubscription(baseOptions: Apollo.SubscriptionHookOptions<SearchLogAddedSubscription, SearchLogAddedSubscriptionVariables> & ({ variables: SearchLogAddedSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<SearchLogAddedSubscription, SearchLogAddedSubscriptionVariables>(SearchLogAddedDocument, options);
      }
export type SearchLogAddedSubscriptionHookResult = ReturnType<typeof useSearchLogAddedSubscription>;
export type SearchLogAddedSubscriptionResult = Apollo.SubscriptionResult<SearchLogAddedSubscription>;
export const VectorStoreLogAddedDocument = gql`
    subscription vectorStoreLogAdded($flowId: ID!) {
  vectorStoreLogAdded(flowId: $flowId) {
    ...vectorStoreLogFragment
  }
}
    ${VectorStoreLogFragmentFragmentDoc}`;

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
export function useVectorStoreLogAddedSubscription(baseOptions: Apollo.SubscriptionHookOptions<VectorStoreLogAddedSubscription, VectorStoreLogAddedSubscriptionVariables> & ({ variables: VectorStoreLogAddedSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<VectorStoreLogAddedSubscription, VectorStoreLogAddedSubscriptionVariables>(VectorStoreLogAddedDocument, options);
      }
export type VectorStoreLogAddedSubscriptionHookResult = ReturnType<typeof useVectorStoreLogAddedSubscription>;
export type VectorStoreLogAddedSubscriptionResult = Apollo.SubscriptionResult<VectorStoreLogAddedSubscription>;
export const AssistantCreatedDocument = gql`
    subscription assistantCreated($flowId: ID!) {
  assistantCreated(flowId: $flowId) {
    ...assistantFragment
  }
}
    ${AssistantFragmentFragmentDoc}`;

/**
 * __useAssistantCreatedSubscription__
 *
 * To run a query within a React component, call `useAssistantCreatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useAssistantCreatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAssistantCreatedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useAssistantCreatedSubscription(baseOptions: Apollo.SubscriptionHookOptions<AssistantCreatedSubscription, AssistantCreatedSubscriptionVariables> & ({ variables: AssistantCreatedSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<AssistantCreatedSubscription, AssistantCreatedSubscriptionVariables>(AssistantCreatedDocument, options);
      }
export type AssistantCreatedSubscriptionHookResult = ReturnType<typeof useAssistantCreatedSubscription>;
export type AssistantCreatedSubscriptionResult = Apollo.SubscriptionResult<AssistantCreatedSubscription>;
export const AssistantUpdatedDocument = gql`
    subscription assistantUpdated($flowId: ID!) {
  assistantUpdated(flowId: $flowId) {
    ...assistantFragment
  }
}
    ${AssistantFragmentFragmentDoc}`;

/**
 * __useAssistantUpdatedSubscription__
 *
 * To run a query within a React component, call `useAssistantUpdatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useAssistantUpdatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAssistantUpdatedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useAssistantUpdatedSubscription(baseOptions: Apollo.SubscriptionHookOptions<AssistantUpdatedSubscription, AssistantUpdatedSubscriptionVariables> & ({ variables: AssistantUpdatedSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<AssistantUpdatedSubscription, AssistantUpdatedSubscriptionVariables>(AssistantUpdatedDocument, options);
      }
export type AssistantUpdatedSubscriptionHookResult = ReturnType<typeof useAssistantUpdatedSubscription>;
export type AssistantUpdatedSubscriptionResult = Apollo.SubscriptionResult<AssistantUpdatedSubscription>;
export const AssistantDeletedDocument = gql`
    subscription assistantDeleted($flowId: ID!) {
  assistantDeleted(flowId: $flowId) {
    ...assistantFragment
  }
}
    ${AssistantFragmentFragmentDoc}`;

/**
 * __useAssistantDeletedSubscription__
 *
 * To run a query within a React component, call `useAssistantDeletedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useAssistantDeletedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAssistantDeletedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useAssistantDeletedSubscription(baseOptions: Apollo.SubscriptionHookOptions<AssistantDeletedSubscription, AssistantDeletedSubscriptionVariables> & ({ variables: AssistantDeletedSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<AssistantDeletedSubscription, AssistantDeletedSubscriptionVariables>(AssistantDeletedDocument, options);
      }
export type AssistantDeletedSubscriptionHookResult = ReturnType<typeof useAssistantDeletedSubscription>;
export type AssistantDeletedSubscriptionResult = Apollo.SubscriptionResult<AssistantDeletedSubscription>;
export const AssistantLogAddedDocument = gql`
    subscription assistantLogAdded($flowId: ID!) {
  assistantLogAdded(flowId: $flowId) {
    ...assistantLogFragment
  }
}
    ${AssistantLogFragmentFragmentDoc}`;

/**
 * __useAssistantLogAddedSubscription__
 *
 * To run a query within a React component, call `useAssistantLogAddedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useAssistantLogAddedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAssistantLogAddedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useAssistantLogAddedSubscription(baseOptions: Apollo.SubscriptionHookOptions<AssistantLogAddedSubscription, AssistantLogAddedSubscriptionVariables> & ({ variables: AssistantLogAddedSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<AssistantLogAddedSubscription, AssistantLogAddedSubscriptionVariables>(AssistantLogAddedDocument, options);
      }
export type AssistantLogAddedSubscriptionHookResult = ReturnType<typeof useAssistantLogAddedSubscription>;
export type AssistantLogAddedSubscriptionResult = Apollo.SubscriptionResult<AssistantLogAddedSubscription>;
export const AssistantLogUpdatedDocument = gql`
    subscription assistantLogUpdated($flowId: ID!) {
  assistantLogUpdated(flowId: $flowId) {
    ...assistantLogFragment
  }
}
    ${AssistantLogFragmentFragmentDoc}`;

/**
 * __useAssistantLogUpdatedSubscription__
 *
 * To run a query within a React component, call `useAssistantLogUpdatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useAssistantLogUpdatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAssistantLogUpdatedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useAssistantLogUpdatedSubscription(baseOptions: Apollo.SubscriptionHookOptions<AssistantLogUpdatedSubscription, AssistantLogUpdatedSubscriptionVariables> & ({ variables: AssistantLogUpdatedSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<AssistantLogUpdatedSubscription, AssistantLogUpdatedSubscriptionVariables>(AssistantLogUpdatedDocument, options);
      }
export type AssistantLogUpdatedSubscriptionHookResult = ReturnType<typeof useAssistantLogUpdatedSubscription>;
export type AssistantLogUpdatedSubscriptionResult = Apollo.SubscriptionResult<AssistantLogUpdatedSubscription>;
export const FlowCreatedDocument = gql`
    subscription flowCreated {
  flowCreated {
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
}
    ${TerminalFragmentFragmentDoc}`;

/**
 * __useFlowCreatedSubscription__
 *
 * To run a query within a React component, call `useFlowCreatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useFlowCreatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowCreatedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useFlowCreatedSubscription(baseOptions?: Apollo.SubscriptionHookOptions<FlowCreatedSubscription, FlowCreatedSubscriptionVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<FlowCreatedSubscription, FlowCreatedSubscriptionVariables>(FlowCreatedDocument, options);
      }
export type FlowCreatedSubscriptionHookResult = ReturnType<typeof useFlowCreatedSubscription>;
export type FlowCreatedSubscriptionResult = Apollo.SubscriptionResult<FlowCreatedSubscription>;
export const FlowDeletedDocument = gql`
    subscription flowDeleted {
  flowDeleted {
    id
    status
    updatedAt
  }
}
    `;

/**
 * __useFlowDeletedSubscription__
 *
 * To run a query within a React component, call `useFlowDeletedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useFlowDeletedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowDeletedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useFlowDeletedSubscription(baseOptions?: Apollo.SubscriptionHookOptions<FlowDeletedSubscription, FlowDeletedSubscriptionVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<FlowDeletedSubscription, FlowDeletedSubscriptionVariables>(FlowDeletedDocument, options);
      }
export type FlowDeletedSubscriptionHookResult = ReturnType<typeof useFlowDeletedSubscription>;
export type FlowDeletedSubscriptionResult = Apollo.SubscriptionResult<FlowDeletedSubscription>;
export const FlowUpdatedDocument = gql`
    subscription flowUpdated {
  flowUpdated {
    id
    title
    status
    terminals {
      ...terminalFragment
    }
    updatedAt
  }
}
    ${TerminalFragmentFragmentDoc}`;

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
 *   },
 * });
 */
export function useFlowUpdatedSubscription(baseOptions?: Apollo.SubscriptionHookOptions<FlowUpdatedSubscription, FlowUpdatedSubscriptionVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<FlowUpdatedSubscription, FlowUpdatedSubscriptionVariables>(FlowUpdatedDocument, options);
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
${SubtaskFragmentFragmentDoc}`;

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
export function useTaskCreatedSubscription(baseOptions: Apollo.SubscriptionHookOptions<TaskCreatedSubscription, TaskCreatedSubscriptionVariables> & ({ variables: TaskCreatedSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<TaskCreatedSubscription, TaskCreatedSubscriptionVariables>(TaskCreatedDocument, options);
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
    ${SubtaskFragmentFragmentDoc}`;

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
export function useTaskUpdatedSubscription(baseOptions: Apollo.SubscriptionHookOptions<TaskUpdatedSubscription, TaskUpdatedSubscriptionVariables> & ({ variables: TaskUpdatedSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<TaskUpdatedSubscription, TaskUpdatedSubscriptionVariables>(TaskUpdatedDocument, options);
      }
export type TaskUpdatedSubscriptionHookResult = ReturnType<typeof useTaskUpdatedSubscription>;
export type TaskUpdatedSubscriptionResult = Apollo.SubscriptionResult<TaskUpdatedSubscription>;
export const FlowReportDocument = gql`
    query flowReport($id: ID!) {
  flow(flowId: $id) {
    ...flowFragment
  }
  tasks(flowId: $id) {
    ...taskFragment
  }
}
    ${FlowFragmentFragmentDoc}
${TerminalFragmentFragmentDoc}
${TaskFragmentFragmentDoc}
${SubtaskFragmentFragmentDoc}`;

/**
 * __useFlowReportQuery__
 *
 * To run a query within a React component, call `useFlowReportQuery` and pass it any options that fit your needs.
 * When your component renders, `useFlowReportQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowReportQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useFlowReportQuery(baseOptions: Apollo.QueryHookOptions<FlowReportQuery, FlowReportQueryVariables> & ({ variables: FlowReportQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<FlowReportQuery, FlowReportQueryVariables>(FlowReportDocument, options);
      }
export function useFlowReportLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<FlowReportQuery, FlowReportQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<FlowReportQuery, FlowReportQueryVariables>(FlowReportDocument, options);
        }
export function useFlowReportSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<FlowReportQuery, FlowReportQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<FlowReportQuery, FlowReportQueryVariables>(FlowReportDocument, options);
        }
export type FlowReportQueryHookResult = ReturnType<typeof useFlowReportQuery>;
export type FlowReportLazyQueryHookResult = ReturnType<typeof useFlowReportLazyQuery>;
export type FlowReportSuspenseQueryHookResult = ReturnType<typeof useFlowReportSuspenseQuery>;
export type FlowReportQueryResult = Apollo.QueryResult<FlowReportQuery, FlowReportQueryVariables>;