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
    ID: { input: string; output: string };
    String: { input: string; output: string };
    Boolean: { input: boolean; output: boolean };
    Int: { input: number; output: number };
    Float: { input: number; output: number };
    Time: { input: any; output: any };
};

export type ApiToken = {
    createdAt: Scalars['Time']['output'];
    id: Scalars['ID']['output'];
    name?: Maybe<Scalars['String']['output']>;
    roleId: Scalars['ID']['output'];
    status: TokenStatus;
    tokenId: Scalars['String']['output'];
    ttl: Scalars['Int']['output'];
    updatedAt: Scalars['Time']['output'];
    userId: Scalars['ID']['output'];
};

export type ApiTokenWithSecret = {
    createdAt: Scalars['Time']['output'];
    id: Scalars['ID']['output'];
    name?: Maybe<Scalars['String']['output']>;
    roleId: Scalars['ID']['output'];
    status: TokenStatus;
    token: Scalars['String']['output'];
    tokenId: Scalars['String']['output'];
    ttl: Scalars['Int']['output'];
    updatedAt: Scalars['Time']['output'];
    userId: Scalars['ID']['output'];
};

export type AgentConfig = {
    frequencyPenalty?: Maybe<Scalars['Float']['output']>;
    maxLength?: Maybe<Scalars['Int']['output']>;
    maxTokens?: Maybe<Scalars['Int']['output']>;
    minLength?: Maybe<Scalars['Int']['output']>;
    model: Scalars['String']['output'];
    presencePenalty?: Maybe<Scalars['Float']['output']>;
    price?: Maybe<ModelPrice>;
    reasoning?: Maybe<ReasoningConfig>;
    repetitionPenalty?: Maybe<Scalars['Float']['output']>;
    temperature?: Maybe<Scalars['Float']['output']>;
    topK?: Maybe<Scalars['Int']['output']>;
    topP?: Maybe<Scalars['Float']['output']>;
};

export type AgentConfigInput = {
    frequencyPenalty?: InputMaybe<Scalars['Float']['input']>;
    maxLength?: InputMaybe<Scalars['Int']['input']>;
    maxTokens?: InputMaybe<Scalars['Int']['input']>;
    minLength?: InputMaybe<Scalars['Int']['input']>;
    model: Scalars['String']['input'];
    presencePenalty?: InputMaybe<Scalars['Float']['input']>;
    price?: InputMaybe<ModelPriceInput>;
    reasoning?: InputMaybe<ReasoningConfigInput>;
    repetitionPenalty?: InputMaybe<Scalars['Float']['input']>;
    temperature?: InputMaybe<Scalars['Float']['input']>;
    topK?: InputMaybe<Scalars['Int']['input']>;
    topP?: InputMaybe<Scalars['Float']['input']>;
};

export enum AgentConfigType {
    Adviser = 'adviser',
    Assistant = 'assistant',
    Coder = 'coder',
    Enricher = 'enricher',
    Generator = 'generator',
    Installer = 'installer',
    Pentester = 'pentester',
    PrimaryAgent = 'primary_agent',
    Refiner = 'refiner',
    Reflector = 'reflector',
    Searcher = 'searcher',
    Simple = 'simple',
    SimpleJson = 'simple_json',
}

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

export type AgentPrompt = {
    system: DefaultPrompt;
};

export type AgentPrompts = {
    human: DefaultPrompt;
    system: DefaultPrompt;
};

export type AgentTestResult = {
    tests: Array<TestResult>;
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
    ToolCallFixer = 'tool_call_fixer',
}

export type AgentTypeUsageStats = {
    agentType: AgentType;
    stats: UsageStats;
};

export type AgentsConfig = {
    adviser: AgentConfig;
    assistant: AgentConfig;
    coder: AgentConfig;
    enricher: AgentConfig;
    generator: AgentConfig;
    installer: AgentConfig;
    pentester: AgentConfig;
    primaryAgent: AgentConfig;
    refiner: AgentConfig;
    reflector: AgentConfig;
    searcher: AgentConfig;
    simple: AgentConfig;
    simpleJson: AgentConfig;
};

export type AgentsConfigInput = {
    adviser: AgentConfigInput;
    assistant: AgentConfigInput;
    coder: AgentConfigInput;
    enricher: AgentConfigInput;
    generator: AgentConfigInput;
    installer: AgentConfigInput;
    pentester: AgentConfigInput;
    primaryAgent: AgentConfigInput;
    refiner: AgentConfigInput;
    reflector: AgentConfigInput;
    searcher: AgentConfigInput;
    simple: AgentConfigInput;
    simpleJson: AgentConfigInput;
};

export type AgentsPrompts = {
    adviser: AgentPrompts;
    assistant: AgentPrompt;
    coder: AgentPrompts;
    enricher: AgentPrompts;
    generator: AgentPrompts;
    installer: AgentPrompts;
    memorist: AgentPrompts;
    pentester: AgentPrompts;
    primaryAgent: AgentPrompt;
    refiner: AgentPrompts;
    reflector: AgentPrompts;
    reporter: AgentPrompts;
    searcher: AgentPrompts;
    summarizer: AgentPrompt;
    toolCallFixer: AgentPrompts;
};

export type Assistant = {
    createdAt: Scalars['Time']['output'];
    flowId: Scalars['ID']['output'];
    id: Scalars['ID']['output'];
    provider: Provider;
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

export type CreateApiTokenInput = {
    name?: InputMaybe<Scalars['String']['input']>;
    ttl: Scalars['Int']['input'];
};

export type DailyFlowsStats = {
    date: Scalars['Time']['output'];
    stats: FlowsStats;
};

export type DailyToolcallsStats = {
    date: Scalars['Time']['output'];
    stats: ToolcallsStats;
};

export type DailyUsageStats = {
    date: Scalars['Time']['output'];
    stats: UsageStats;
};

export type DefaultPrompt = {
    template: Scalars['String']['output'];
    type: PromptType;
    variables: Array<Scalars['String']['output']>;
};

export type DefaultPrompts = {
    agents: AgentsPrompts;
    tools: ToolsPrompts;
};

export type DefaultProvidersConfig = {
    anthropic: ProviderConfig;
    bedrock?: Maybe<ProviderConfig>;
    custom?: Maybe<ProviderConfig>;
    gemini?: Maybe<ProviderConfig>;
    ollama?: Maybe<ProviderConfig>;
    openai: ProviderConfig;
};

export type Flow = {
    createdAt: Scalars['Time']['output'];
    id: Scalars['ID']['output'];
    provider: Provider;
    status: StatusType;
    terminals?: Maybe<Array<Terminal>>;
    title: Scalars['String']['output'];
    updatedAt: Scalars['Time']['output'];
};

export type FlowAssistant = {
    assistant: Assistant;
    flow: Flow;
};

export type FlowExecutionStats = {
    flowId: Scalars['ID']['output'];
    flowTitle: Scalars['String']['output'];
    tasks: Array<TaskExecutionStats>;
    totalAssistantsCount: Scalars['Int']['output'];
    totalDurationSeconds: Scalars['Float']['output'];
    totalToolcallsCount: Scalars['Int']['output'];
};

export type FlowStats = {
    totalAssistantsCount: Scalars['Int']['output'];
    totalSubtasksCount: Scalars['Int']['output'];
    totalTasksCount: Scalars['Int']['output'];
};

export type FlowsStats = {
    totalAssistantsCount: Scalars['Int']['output'];
    totalFlowsCount: Scalars['Int']['output'];
    totalSubtasksCount: Scalars['Int']['output'];
    totalTasksCount: Scalars['Int']['output'];
};

export type FunctionToolcallsStats = {
    avgDurationSeconds: Scalars['Float']['output'];
    functionName: Scalars['String']['output'];
    isAgent: Scalars['Boolean']['output'];
    totalCount: Scalars['Int']['output'];
    totalDurationSeconds: Scalars['Float']['output'];
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
    Thoughts = 'thoughts',
}

export type ModelConfig = {
    description?: Maybe<Scalars['String']['output']>;
    name: Scalars['String']['output'];
    price?: Maybe<ModelPrice>;
    releaseDate?: Maybe<Scalars['Time']['output']>;
    thinking?: Maybe<Scalars['Boolean']['output']>;
};

export type ModelPrice = {
    cacheRead: Scalars['Float']['output'];
    cacheWrite: Scalars['Float']['output'];
    input: Scalars['Float']['output'];
    output: Scalars['Float']['output'];
};

export type ModelPriceInput = {
    cacheRead: Scalars['Float']['input'];
    cacheWrite: Scalars['Float']['input'];
    input: Scalars['Float']['input'];
    output: Scalars['Float']['input'];
};

export type ModelUsageStats = {
    model: Scalars['String']['output'];
    provider: Scalars['String']['output'];
    stats: UsageStats;
};

export type Mutation = {
    addFavoriteFlow: ResultType;
    callAssistant: ResultType;
    createAPIToken: ApiTokenWithSecret;
    createAssistant: FlowAssistant;
    createFlow: Flow;
    createPrompt: UserPrompt;
    createProvider: ProviderConfig;
    deleteAPIToken: Scalars['Boolean']['output'];
    deleteAssistant: ResultType;
    deleteFavoriteFlow: ResultType;
    deleteFlow: ResultType;
    deletePrompt: ResultType;
    deleteProvider: ResultType;
    finishFlow: ResultType;
    putUserInput: ResultType;
    renameFlow: ResultType;
    stopAssistant: Assistant;
    stopFlow: ResultType;
    testAgent: AgentTestResult;
    testProvider: ProviderTestResult;
    updateAPIToken: ApiToken;
    updatePrompt: UserPrompt;
    updateProvider: ProviderConfig;
    validatePrompt: PromptValidationResult;
};

export type MutationAddFavoriteFlowArgs = {
    flowId: Scalars['ID']['input'];
};

export type MutationCallAssistantArgs = {
    assistantId: Scalars['ID']['input'];
    flowId: Scalars['ID']['input'];
    input: Scalars['String']['input'];
    useAgents: Scalars['Boolean']['input'];
};

export type MutationCreateApiTokenArgs = {
    input: CreateApiTokenInput;
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

export type MutationCreatePromptArgs = {
    template: Scalars['String']['input'];
    type: PromptType;
};

export type MutationCreateProviderArgs = {
    agents: AgentsConfigInput;
    name: Scalars['String']['input'];
    type: ProviderType;
};

export type MutationDeleteApiTokenArgs = {
    tokenId: Scalars['String']['input'];
};

export type MutationDeleteAssistantArgs = {
    assistantId: Scalars['ID']['input'];
    flowId: Scalars['ID']['input'];
};

export type MutationDeleteFavoriteFlowArgs = {
    flowId: Scalars['ID']['input'];
};

export type MutationDeleteFlowArgs = {
    flowId: Scalars['ID']['input'];
};

export type MutationDeletePromptArgs = {
    promptId: Scalars['ID']['input'];
};

export type MutationDeleteProviderArgs = {
    providerId: Scalars['ID']['input'];
};

export type MutationFinishFlowArgs = {
    flowId: Scalars['ID']['input'];
};

export type MutationPutUserInputArgs = {
    flowId: Scalars['ID']['input'];
    input: Scalars['String']['input'];
};

export type MutationRenameFlowArgs = {
    flowId: Scalars['ID']['input'];
    title: Scalars['String']['input'];
};

export type MutationStopAssistantArgs = {
    assistantId: Scalars['ID']['input'];
    flowId: Scalars['ID']['input'];
};

export type MutationStopFlowArgs = {
    flowId: Scalars['ID']['input'];
};

export type MutationTestAgentArgs = {
    agent: AgentConfigInput;
    agentType: AgentConfigType;
    type: ProviderType;
};

export type MutationTestProviderArgs = {
    agents: AgentsConfigInput;
    type: ProviderType;
};

export type MutationUpdateApiTokenArgs = {
    input: UpdateApiTokenInput;
    tokenId: Scalars['String']['input'];
};

export type MutationUpdatePromptArgs = {
    promptId: Scalars['ID']['input'];
    template: Scalars['String']['input'];
};

export type MutationUpdateProviderArgs = {
    agents: AgentsConfigInput;
    name: Scalars['String']['input'];
    providerId: Scalars['ID']['input'];
};

export type MutationValidatePromptArgs = {
    template: Scalars['String']['input'];
    type: PromptType;
};

export enum PromptType {
    Adviser = 'adviser',
    Assistant = 'assistant',
    Coder = 'coder',
    Enricher = 'enricher',
    ExecutionLogs = 'execution_logs',
    FlowDescriptor = 'flow_descriptor',
    FullExecutionContext = 'full_execution_context',
    Generator = 'generator',
    ImageChooser = 'image_chooser',
    InputToolcallFixer = 'input_toolcall_fixer',
    Installer = 'installer',
    LanguageChooser = 'language_chooser',
    Memorist = 'memorist',
    Pentester = 'pentester',
    PrimaryAgent = 'primary_agent',
    QuestionAdviser = 'question_adviser',
    QuestionCoder = 'question_coder',
    QuestionEnricher = 'question_enricher',
    QuestionInstaller = 'question_installer',
    QuestionMemorist = 'question_memorist',
    QuestionPentester = 'question_pentester',
    QuestionReflector = 'question_reflector',
    QuestionSearcher = 'question_searcher',
    Refiner = 'refiner',
    Reflector = 'reflector',
    Reporter = 'reporter',
    Searcher = 'searcher',
    ShortExecutionContext = 'short_execution_context',
    SubtasksGenerator = 'subtasks_generator',
    SubtasksRefiner = 'subtasks_refiner',
    Summarizer = 'summarizer',
    TaskDescriptor = 'task_descriptor',
    TaskReporter = 'task_reporter',
    ToolCallIdCollector = 'tool_call_id_collector',
    ToolCallIdDetector = 'tool_call_id_detector',
    ToolcallFixer = 'toolcall_fixer',
}

export enum PromptValidationErrorType {
    EmptyTemplate = 'empty_template',
    RenderingFailed = 'rendering_failed',
    SyntaxError = 'syntax_error',
    UnauthorizedVariable = 'unauthorized_variable',
    UnknownType = 'unknown_type',
    VariableTypeMismatch = 'variable_type_mismatch',
}

export type PromptValidationResult = {
    details?: Maybe<Scalars['String']['output']>;
    errorType?: Maybe<PromptValidationErrorType>;
    line?: Maybe<Scalars['Int']['output']>;
    message?: Maybe<Scalars['String']['output']>;
    result: ResultType;
};

export type PromptsConfig = {
    default: DefaultPrompts;
    userDefined?: Maybe<Array<UserPrompt>>;
};

export type Provider = {
    name: Scalars['String']['output'];
    type: ProviderType;
};

export type ProviderConfig = {
    agents: AgentsConfig;
    createdAt: Scalars['Time']['output'];
    id: Scalars['ID']['output'];
    name: Scalars['String']['output'];
    type: ProviderType;
    updatedAt: Scalars['Time']['output'];
};

export type ProviderTestResult = {
    adviser: AgentTestResult;
    assistant: AgentTestResult;
    coder: AgentTestResult;
    enricher: AgentTestResult;
    generator: AgentTestResult;
    installer: AgentTestResult;
    pentester: AgentTestResult;
    primaryAgent: AgentTestResult;
    refiner: AgentTestResult;
    reflector: AgentTestResult;
    searcher: AgentTestResult;
    simple: AgentTestResult;
    simpleJson: AgentTestResult;
};

export enum ProviderType {
    Anthropic = 'anthropic',
    Bedrock = 'bedrock',
    Custom = 'custom',
    Gemini = 'gemini',
    Ollama = 'ollama',
    Openai = 'openai',
}

export type ProviderUsageStats = {
    provider: Scalars['String']['output'];
    stats: UsageStats;
};

export type ProvidersConfig = {
    default: DefaultProvidersConfig;
    enabled: ProvidersReadinessStatus;
    models: ProvidersModelsList;
    userDefined?: Maybe<Array<ProviderConfig>>;
};

export type ProvidersModelsList = {
    anthropic: Array<ModelConfig>;
    bedrock?: Maybe<Array<ModelConfig>>;
    custom?: Maybe<Array<ModelConfig>>;
    gemini: Array<ModelConfig>;
    ollama?: Maybe<Array<ModelConfig>>;
    openai: Array<ModelConfig>;
};

export type ProvidersReadinessStatus = {
    anthropic: Scalars['Boolean']['output'];
    bedrock: Scalars['Boolean']['output'];
    custom: Scalars['Boolean']['output'];
    gemini: Scalars['Boolean']['output'];
    ollama: Scalars['Boolean']['output'];
    openai: Scalars['Boolean']['output'];
};

export type Query = {
    agentLogs?: Maybe<Array<AgentLog>>;
    apiToken?: Maybe<ApiToken>;
    apiTokens: Array<ApiToken>;
    assistantLogs?: Maybe<Array<AssistantLog>>;
    assistants?: Maybe<Array<Assistant>>;
    flow: Flow;
    flowStatsByFlow: FlowStats;
    flows?: Maybe<Array<Flow>>;
    flowsExecutionStatsByPeriod: Array<FlowExecutionStats>;
    flowsStatsByPeriod: Array<DailyFlowsStats>;
    flowsStatsTotal: FlowsStats;
    messageLogs?: Maybe<Array<MessageLog>>;
    providers: Array<Provider>;
    screenshots?: Maybe<Array<Screenshot>>;
    searchLogs?: Maybe<Array<SearchLog>>;
    settings: Settings;
    settingsPrompts: PromptsConfig;
    settingsProviders: ProvidersConfig;
    settingsUser: UserPreferences;
    tasks?: Maybe<Array<Task>>;
    terminalLogs?: Maybe<Array<TerminalLog>>;
    toolcallsStatsByFlow: ToolcallsStats;
    toolcallsStatsByFunction: Array<FunctionToolcallsStats>;
    toolcallsStatsByFunctionForFlow: Array<FunctionToolcallsStats>;
    toolcallsStatsByPeriod: Array<DailyToolcallsStats>;
    toolcallsStatsTotal: ToolcallsStats;
    usageStatsByAgentType: Array<AgentTypeUsageStats>;
    usageStatsByAgentTypeForFlow: Array<AgentTypeUsageStats>;
    usageStatsByFlow: UsageStats;
    usageStatsByModel: Array<ModelUsageStats>;
    usageStatsByPeriod: Array<DailyUsageStats>;
    usageStatsByProvider: Array<ProviderUsageStats>;
    usageStatsTotal: UsageStats;
    vectorStoreLogs?: Maybe<Array<VectorStoreLog>>;
};

export type QueryAgentLogsArgs = {
    flowId: Scalars['ID']['input'];
};

export type QueryApiTokenArgs = {
    tokenId: Scalars['String']['input'];
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

export type QueryFlowStatsByFlowArgs = {
    flowId: Scalars['ID']['input'];
};

export type QueryFlowsExecutionStatsByPeriodArgs = {
    period: UsageStatsPeriod;
};

export type QueryFlowsStatsByPeriodArgs = {
    period: UsageStatsPeriod;
};

export type QueryMessageLogsArgs = {
    flowId: Scalars['ID']['input'];
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

export type QueryToolcallsStatsByFlowArgs = {
    flowId: Scalars['ID']['input'];
};

export type QueryToolcallsStatsByFunctionForFlowArgs = {
    flowId: Scalars['ID']['input'];
};

export type QueryToolcallsStatsByPeriodArgs = {
    period: UsageStatsPeriod;
};

export type QueryUsageStatsByAgentTypeForFlowArgs = {
    flowId: Scalars['ID']['input'];
};

export type QueryUsageStatsByFlowArgs = {
    flowId: Scalars['ID']['input'];
};

export type QueryUsageStatsByPeriodArgs = {
    period: UsageStatsPeriod;
};

export type QueryVectorStoreLogsArgs = {
    flowId: Scalars['ID']['input'];
};

export type ReasoningConfig = {
    effort?: Maybe<ReasoningEffort>;
    maxTokens?: Maybe<Scalars['Int']['output']>;
};

export type ReasoningConfigInput = {
    effort?: InputMaybe<ReasoningEffort>;
    maxTokens?: InputMaybe<Scalars['Int']['input']>;
};

export enum ReasoningEffort {
    High = 'high',
    Low = 'low',
    Medium = 'medium',
}

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
    subtaskId?: Maybe<Scalars['ID']['output']>;
    taskId?: Maybe<Scalars['ID']['output']>;
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
    Waiting = 'waiting',
}

export type Subscription = {
    agentLogAdded: AgentLog;
    apiTokenCreated: ApiToken;
    apiTokenDeleted: ApiToken;
    apiTokenUpdated: ApiToken;
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
    providerCreated: ProviderConfig;
    providerDeleted: ProviderConfig;
    providerUpdated: ProviderConfig;
    screenshotAdded: Screenshot;
    searchLogAdded: SearchLog;
    settingsUserUpdated: UserPreferences;
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

export type SubtaskExecutionStats = {
    subtaskId: Scalars['ID']['output'];
    subtaskTitle: Scalars['String']['output'];
    totalDurationSeconds: Scalars['Float']['output'];
    totalToolcallsCount: Scalars['Int']['output'];
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

export type TaskExecutionStats = {
    subtasks: Array<SubtaskExecutionStats>;
    taskId: Scalars['ID']['output'];
    taskTitle: Scalars['String']['output'];
    totalDurationSeconds: Scalars['Float']['output'];
    totalToolcallsCount: Scalars['Int']['output'];
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
    subtaskId?: Maybe<Scalars['ID']['output']>;
    taskId?: Maybe<Scalars['ID']['output']>;
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

export type TestResult = {
    error?: Maybe<Scalars['String']['output']>;
    latency?: Maybe<Scalars['Int']['output']>;
    name: Scalars['String']['output'];
    reasoning: Scalars['Boolean']['output'];
    result: Scalars['Boolean']['output'];
    streaming: Scalars['Boolean']['output'];
    type: Scalars['String']['output'];
};

export enum TokenStatus {
    Active = 'active',
    Expired = 'expired',
    Revoked = 'revoked',
}

export type ToolcallsStats = {
    totalCount: Scalars['Int']['output'];
    totalDurationSeconds: Scalars['Float']['output'];
};

export type ToolsPrompts = {
    chooseDockerImage: DefaultPrompt;
    chooseUserLanguage: DefaultPrompt;
    collectToolCallID: DefaultPrompt;
    detectToolCallIDPattern: DefaultPrompt;
    getExecutionLogs: DefaultPrompt;
    getFlowDescription: DefaultPrompt;
    getFullExecutionContext: DefaultPrompt;
    getShortExecutionContext: DefaultPrompt;
    getTaskDescription: DefaultPrompt;
};

export type UpdateApiTokenInput = {
    name?: InputMaybe<Scalars['String']['input']>;
    status?: InputMaybe<TokenStatus>;
};

export type UsageStats = {
    totalUsageCacheIn: Scalars['Int']['output'];
    totalUsageCacheOut: Scalars['Int']['output'];
    totalUsageCostIn: Scalars['Float']['output'];
    totalUsageCostOut: Scalars['Float']['output'];
    totalUsageIn: Scalars['Int']['output'];
    totalUsageOut: Scalars['Int']['output'];
};

export enum UsageStatsPeriod {
    Month = 'month',
    Quarter = 'quarter',
    Week = 'week',
}

export type UserPreferences = {
    favoriteFlows: Array<Scalars['ID']['output']>;
    id: Scalars['ID']['output'];
};

export type UserPrompt = {
    createdAt: Scalars['Time']['output'];
    id: Scalars['ID']['output'];
    template: Scalars['String']['output'];
    type: PromptType;
    updatedAt: Scalars['Time']['output'];
};

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

export type SettingsFragmentFragment = {
    debug: boolean;
    askUser: boolean;
    dockerInside: boolean;
    assistantUseAgents: boolean;
};

export type FlowFragmentFragment = {
    id: string;
    title: string;
    status: StatusType;
    createdAt: any;
    updatedAt: any;
    terminals?: Array<TerminalFragmentFragment> | null;
    provider: ProviderFragmentFragment;
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
    taskId?: string | null;
    subtaskId?: string | null;
    type: TerminalLogType;
    text: string;
    terminal: string;
    createdAt: any;
};

export type MessageLogFragmentFragment = {
    id: string;
    type: MessageLogType;
    message: string;
    thinking?: string | null;
    result: string;
    resultFormat: ResultFormat;
    flowId: string;
    taskId?: string | null;
    subtaskId?: string | null;
    createdAt: any;
};

export type ScreenshotFragmentFragment = {
    id: string;
    flowId: string;
    taskId?: string | null;
    subtaskId?: string | null;
    name: string;
    url: string;
    createdAt: any;
};

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

export type AssistantFragmentFragment = {
    id: string;
    title: string;
    status: StatusType;
    flowId: string;
    useAgents: boolean;
    createdAt: any;
    updatedAt: any;
    provider: ProviderFragmentFragment;
};

export type AssistantLogFragmentFragment = {
    id: string;
    type: MessageLogType;
    message: string;
    thinking?: string | null;
    result: string;
    resultFormat: ResultFormat;
    appendPart: boolean;
    flowId: string;
    assistantId: string;
    createdAt: any;
};

export type TestResultFragmentFragment = {
    name: string;
    type: string;
    result: boolean;
    reasoning: boolean;
    streaming: boolean;
    latency?: number | null;
    error?: string | null;
};

export type AgentTestResultFragmentFragment = { tests: Array<TestResultFragmentFragment> };

export type ProviderTestResultFragmentFragment = {
    simple: AgentTestResultFragmentFragment;
    simpleJson: AgentTestResultFragmentFragment;
    primaryAgent: AgentTestResultFragmentFragment;
    assistant: AgentTestResultFragmentFragment;
    generator: AgentTestResultFragmentFragment;
    refiner: AgentTestResultFragmentFragment;
    adviser: AgentTestResultFragmentFragment;
    reflector: AgentTestResultFragmentFragment;
    searcher: AgentTestResultFragmentFragment;
    enricher: AgentTestResultFragmentFragment;
    coder: AgentTestResultFragmentFragment;
    installer: AgentTestResultFragmentFragment;
    pentester: AgentTestResultFragmentFragment;
};

export type ModelConfigFragmentFragment = {
    name: string;
    price?: { input: number; output: number; cacheRead: number; cacheWrite: number } | null;
};

export type ProviderFragmentFragment = { name: string; type: ProviderType };

export type ProviderConfigFragmentFragment = {
    id: string;
    name: string;
    type: ProviderType;
    createdAt: any;
    updatedAt: any;
    agents: AgentsConfigFragmentFragment;
};

export type AgentsConfigFragmentFragment = {
    simple: AgentConfigFragmentFragment;
    simpleJson: AgentConfigFragmentFragment;
    primaryAgent: AgentConfigFragmentFragment;
    assistant: AgentConfigFragmentFragment;
    generator: AgentConfigFragmentFragment;
    refiner: AgentConfigFragmentFragment;
    adviser: AgentConfigFragmentFragment;
    reflector: AgentConfigFragmentFragment;
    searcher: AgentConfigFragmentFragment;
    enricher: AgentConfigFragmentFragment;
    coder: AgentConfigFragmentFragment;
    installer: AgentConfigFragmentFragment;
    pentester: AgentConfigFragmentFragment;
};

export type AgentConfigFragmentFragment = {
    model: string;
    maxTokens?: number | null;
    temperature?: number | null;
    topK?: number | null;
    topP?: number | null;
    minLength?: number | null;
    maxLength?: number | null;
    repetitionPenalty?: number | null;
    frequencyPenalty?: number | null;
    presencePenalty?: number | null;
    reasoning?: { effort?: ReasoningEffort | null; maxTokens?: number | null } | null;
    price?: { input: number; output: number; cacheRead: number; cacheWrite: number } | null;
};

export type UserPromptFragmentFragment = {
    id: string;
    type: PromptType;
    template: string;
    createdAt: any;
    updatedAt: any;
};

export type DefaultPromptFragmentFragment = { type: PromptType; template: string; variables: Array<string> };

export type PromptValidationResultFragmentFragment = {
    result: ResultType;
    errorType?: PromptValidationErrorType | null;
    message?: string | null;
    line?: number | null;
    details?: string | null;
};

export type ApiTokenFragmentFragment = {
    id: string;
    tokenId: string;
    userId: string;
    roleId: string;
    name?: string | null;
    ttl: number;
    status: TokenStatus;
    createdAt: any;
    updatedAt: any;
};

export type ApiTokenWithSecretFragmentFragment = {
    id: string;
    tokenId: string;
    userId: string;
    roleId: string;
    name?: string | null;
    ttl: number;
    status: TokenStatus;
    createdAt: any;
    updatedAt: any;
    token: string;
};

export type UsageStatsFragmentFragment = {
    totalUsageIn: number;
    totalUsageOut: number;
    totalUsageCacheIn: number;
    totalUsageCacheOut: number;
    totalUsageCostIn: number;
    totalUsageCostOut: number;
};

export type DailyUsageStatsFragmentFragment = { date: any; stats: UsageStatsFragmentFragment };

export type ProviderUsageStatsFragmentFragment = { provider: string; stats: UsageStatsFragmentFragment };

export type ModelUsageStatsFragmentFragment = { model: string; provider: string; stats: UsageStatsFragmentFragment };

export type AgentTypeUsageStatsFragmentFragment = { agentType: AgentType; stats: UsageStatsFragmentFragment };

export type ToolcallsStatsFragmentFragment = { totalCount: number; totalDurationSeconds: number };

export type DailyToolcallsStatsFragmentFragment = { date: any; stats: ToolcallsStatsFragmentFragment };

export type FunctionToolcallsStatsFragmentFragment = {
    functionName: string;
    isAgent: boolean;
    totalCount: number;
    totalDurationSeconds: number;
    avgDurationSeconds: number;
};

export type FlowsStatsFragmentFragment = {
    totalFlowsCount: number;
    totalTasksCount: number;
    totalSubtasksCount: number;
    totalAssistantsCount: number;
};

export type FlowStatsFragmentFragment = {
    totalTasksCount: number;
    totalSubtasksCount: number;
    totalAssistantsCount: number;
};

export type DailyFlowsStatsFragmentFragment = { date: any; stats: FlowsStatsFragmentFragment };

export type SubtaskExecutionStatsFragmentFragment = {
    subtaskId: string;
    subtaskTitle: string;
    totalDurationSeconds: number;
    totalToolcallsCount: number;
};

export type TaskExecutionStatsFragmentFragment = {
    taskId: string;
    taskTitle: string;
    totalDurationSeconds: number;
    totalToolcallsCount: number;
    subtasks: Array<SubtaskExecutionStatsFragmentFragment>;
};

export type FlowExecutionStatsFragmentFragment = {
    flowId: string;
    flowTitle: string;
    totalDurationSeconds: number;
    totalToolcallsCount: number;
    totalAssistantsCount: number;
    tasks: Array<TaskExecutionStatsFragmentFragment>;
};

export type FlowsQueryVariables = Exact<{ [key: string]: never }>;

export type FlowsQuery = { flows?: Array<FlowFragmentFragment> | null };

export type ProvidersQueryVariables = Exact<{ [key: string]: never }>;

export type ProvidersQuery = { providers: Array<ProviderFragmentFragment> };

export type SettingsQueryVariables = Exact<{ [key: string]: never }>;

export type SettingsQuery = { settings: SettingsFragmentFragment };

export type SettingsProvidersQueryVariables = Exact<{ [key: string]: never }>;

export type SettingsProvidersQuery = {
    settingsProviders: {
        enabled: {
            openai: boolean;
            anthropic: boolean;
            gemini: boolean;
            bedrock: boolean;
            ollama: boolean;
            custom: boolean;
        };
        default: {
            openai: ProviderConfigFragmentFragment;
            anthropic: ProviderConfigFragmentFragment;
            gemini?: ProviderConfigFragmentFragment | null;
            bedrock?: ProviderConfigFragmentFragment | null;
            ollama?: ProviderConfigFragmentFragment | null;
            custom?: ProviderConfigFragmentFragment | null;
        };
        userDefined?: Array<ProviderConfigFragmentFragment> | null;
        models: {
            openai: Array<ModelConfigFragmentFragment>;
            anthropic: Array<ModelConfigFragmentFragment>;
            gemini: Array<ModelConfigFragmentFragment>;
            bedrock?: Array<ModelConfigFragmentFragment> | null;
            ollama?: Array<ModelConfigFragmentFragment> | null;
            custom?: Array<ModelConfigFragmentFragment> | null;
        };
    };
};

export type SettingsPromptsQueryVariables = Exact<{ [key: string]: never }>;

export type SettingsPromptsQuery = {
    settingsPrompts: {
        default: {
            agents: {
                primaryAgent: { system: DefaultPromptFragmentFragment };
                assistant: { system: DefaultPromptFragmentFragment };
                pentester: { system: DefaultPromptFragmentFragment; human: DefaultPromptFragmentFragment };
                coder: { system: DefaultPromptFragmentFragment; human: DefaultPromptFragmentFragment };
                installer: { system: DefaultPromptFragmentFragment; human: DefaultPromptFragmentFragment };
                searcher: { system: DefaultPromptFragmentFragment; human: DefaultPromptFragmentFragment };
                memorist: { system: DefaultPromptFragmentFragment; human: DefaultPromptFragmentFragment };
                adviser: { system: DefaultPromptFragmentFragment; human: DefaultPromptFragmentFragment };
                generator: { system: DefaultPromptFragmentFragment; human: DefaultPromptFragmentFragment };
                refiner: { system: DefaultPromptFragmentFragment; human: DefaultPromptFragmentFragment };
                reporter: { system: DefaultPromptFragmentFragment; human: DefaultPromptFragmentFragment };
                reflector: { system: DefaultPromptFragmentFragment; human: DefaultPromptFragmentFragment };
                enricher: { system: DefaultPromptFragmentFragment; human: DefaultPromptFragmentFragment };
                toolCallFixer: { system: DefaultPromptFragmentFragment; human: DefaultPromptFragmentFragment };
                summarizer: { system: DefaultPromptFragmentFragment };
            };
            tools: {
                getFlowDescription: DefaultPromptFragmentFragment;
                getTaskDescription: DefaultPromptFragmentFragment;
                getExecutionLogs: DefaultPromptFragmentFragment;
                getFullExecutionContext: DefaultPromptFragmentFragment;
                getShortExecutionContext: DefaultPromptFragmentFragment;
                chooseDockerImage: DefaultPromptFragmentFragment;
                chooseUserLanguage: DefaultPromptFragmentFragment;
                collectToolCallID: DefaultPromptFragmentFragment;
                detectToolCallIDPattern: DefaultPromptFragmentFragment;
            };
        };
        userDefined?: Array<UserPromptFragmentFragment> | null;
    };
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

export type AssistantsQueryVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type AssistantsQuery = { assistants?: Array<AssistantFragmentFragment> | null };

export type AssistantLogsQueryVariables = Exact<{
    flowId: Scalars['ID']['input'];
    assistantId: Scalars['ID']['input'];
}>;

export type AssistantLogsQuery = { assistantLogs?: Array<AssistantLogFragmentFragment> | null };

export type FlowReportQueryVariables = Exact<{
    id: Scalars['ID']['input'];
}>;

export type FlowReportQuery = { flow: FlowFragmentFragment; tasks?: Array<TaskFragmentFragment> | null };

export type UsageStatsTotalQueryVariables = Exact<{ [key: string]: never }>;

export type UsageStatsTotalQuery = { usageStatsTotal: UsageStatsFragmentFragment };

export type UsageStatsByPeriodQueryVariables = Exact<{
    period: UsageStatsPeriod;
}>;

export type UsageStatsByPeriodQuery = { usageStatsByPeriod: Array<DailyUsageStatsFragmentFragment> };

export type UsageStatsByProviderQueryVariables = Exact<{ [key: string]: never }>;

export type UsageStatsByProviderQuery = { usageStatsByProvider: Array<ProviderUsageStatsFragmentFragment> };

export type UsageStatsByModelQueryVariables = Exact<{ [key: string]: never }>;

export type UsageStatsByModelQuery = { usageStatsByModel: Array<ModelUsageStatsFragmentFragment> };

export type UsageStatsByAgentTypeQueryVariables = Exact<{ [key: string]: never }>;

export type UsageStatsByAgentTypeQuery = { usageStatsByAgentType: Array<AgentTypeUsageStatsFragmentFragment> };

export type UsageStatsByFlowQueryVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type UsageStatsByFlowQuery = { usageStatsByFlow: UsageStatsFragmentFragment };

export type UsageStatsByAgentTypeForFlowQueryVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type UsageStatsByAgentTypeForFlowQuery = {
    usageStatsByAgentTypeForFlow: Array<AgentTypeUsageStatsFragmentFragment>;
};

export type ToolcallsStatsTotalQueryVariables = Exact<{ [key: string]: never }>;

export type ToolcallsStatsTotalQuery = { toolcallsStatsTotal: ToolcallsStatsFragmentFragment };

export type ToolcallsStatsByPeriodQueryVariables = Exact<{
    period: UsageStatsPeriod;
}>;

export type ToolcallsStatsByPeriodQuery = { toolcallsStatsByPeriod: Array<DailyToolcallsStatsFragmentFragment> };

export type ToolcallsStatsByFunctionQueryVariables = Exact<{ [key: string]: never }>;

export type ToolcallsStatsByFunctionQuery = { toolcallsStatsByFunction: Array<FunctionToolcallsStatsFragmentFragment> };

export type ToolcallsStatsByFlowQueryVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type ToolcallsStatsByFlowQuery = { toolcallsStatsByFlow: ToolcallsStatsFragmentFragment };

export type ToolcallsStatsByFunctionForFlowQueryVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type ToolcallsStatsByFunctionForFlowQuery = {
    toolcallsStatsByFunctionForFlow: Array<FunctionToolcallsStatsFragmentFragment>;
};

export type FlowsStatsTotalQueryVariables = Exact<{ [key: string]: never }>;

export type FlowsStatsTotalQuery = { flowsStatsTotal: FlowsStatsFragmentFragment };

export type FlowsStatsByPeriodQueryVariables = Exact<{
    period: UsageStatsPeriod;
}>;

export type FlowsStatsByPeriodQuery = { flowsStatsByPeriod: Array<DailyFlowsStatsFragmentFragment> };

export type FlowStatsByFlowQueryVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type FlowStatsByFlowQuery = { flowStatsByFlow: FlowStatsFragmentFragment };

export type FlowsExecutionStatsByPeriodQueryVariables = Exact<{
    period: UsageStatsPeriod;
}>;

export type FlowsExecutionStatsByPeriodQuery = {
    flowsExecutionStatsByPeriod: Array<FlowExecutionStatsFragmentFragment>;
};

export type ApiTokensQueryVariables = Exact<{ [key: string]: never }>;

export type ApiTokensQuery = { apiTokens: Array<ApiTokenFragmentFragment> };

export type ApiTokenQueryVariables = Exact<{
    tokenId: Scalars['String']['input'];
}>;

export type ApiTokenQuery = { apiToken?: ApiTokenFragmentFragment | null };

export type UserPreferencesFragmentFragment = { id: string; favoriteFlows: Array<string> };

export type SettingsUserQueryVariables = Exact<{ [key: string]: never }>;

export type SettingsUserQuery = { settingsUser: UserPreferencesFragmentFragment };

export type AddFavoriteFlowMutationVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type AddFavoriteFlowMutation = { addFavoriteFlow: ResultType };

export type DeleteFavoriteFlowMutationVariables = Exact<{
    flowId: Scalars['ID']['input'];
}>;

export type DeleteFavoriteFlowMutation = { deleteFavoriteFlow: ResultType };

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

export type RenameFlowMutationVariables = Exact<{
    flowId: Scalars['ID']['input'];
    title: Scalars['String']['input'];
}>;

export type RenameFlowMutation = { renameFlow: ResultType };

export type CreateAssistantMutationVariables = Exact<{
    flowId: Scalars['ID']['input'];
    modelProvider: Scalars['String']['input'];
    input: Scalars['String']['input'];
    useAgents: Scalars['Boolean']['input'];
}>;

export type CreateAssistantMutation = {
    createAssistant: { flow: FlowFragmentFragment; assistant: AssistantFragmentFragment };
};

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

export type TestAgentMutationVariables = Exact<{
    type: ProviderType;
    agentType: AgentConfigType;
    agent: AgentConfigInput;
}>;

export type TestAgentMutation = { testAgent: AgentTestResultFragmentFragment };

export type TestProviderMutationVariables = Exact<{
    type: ProviderType;
    agents: AgentsConfigInput;
}>;

export type TestProviderMutation = { testProvider: ProviderTestResultFragmentFragment };

export type CreateProviderMutationVariables = Exact<{
    name: Scalars['String']['input'];
    type: ProviderType;
    agents: AgentsConfigInput;
}>;

export type CreateProviderMutation = { createProvider: ProviderConfigFragmentFragment };

export type UpdateProviderMutationVariables = Exact<{
    providerId: Scalars['ID']['input'];
    name: Scalars['String']['input'];
    agents: AgentsConfigInput;
}>;

export type UpdateProviderMutation = { updateProvider: ProviderConfigFragmentFragment };

export type DeleteProviderMutationVariables = Exact<{
    providerId: Scalars['ID']['input'];
}>;

export type DeleteProviderMutation = { deleteProvider: ResultType };

export type ValidatePromptMutationVariables = Exact<{
    type: PromptType;
    template: Scalars['String']['input'];
}>;

export type ValidatePromptMutation = { validatePrompt: PromptValidationResultFragmentFragment };

export type CreatePromptMutationVariables = Exact<{
    type: PromptType;
    template: Scalars['String']['input'];
}>;

export type CreatePromptMutation = { createPrompt: UserPromptFragmentFragment };

export type UpdatePromptMutationVariables = Exact<{
    promptId: Scalars['ID']['input'];
    template: Scalars['String']['input'];
}>;

export type UpdatePromptMutation = { updatePrompt: UserPromptFragmentFragment };

export type DeletePromptMutationVariables = Exact<{
    promptId: Scalars['ID']['input'];
}>;

export type DeletePromptMutation = { deletePrompt: ResultType };

export type CreateApiTokenMutationVariables = Exact<{
    input: CreateApiTokenInput;
}>;

export type CreateApiTokenMutation = { createAPIToken: ApiTokenWithSecretFragmentFragment };

export type UpdateApiTokenMutationVariables = Exact<{
    tokenId: Scalars['String']['input'];
    input: UpdateApiTokenInput;
}>;

export type UpdateApiTokenMutation = { updateAPIToken: ApiTokenFragmentFragment };

export type DeleteApiTokenMutationVariables = Exact<{
    tokenId: Scalars['String']['input'];
}>;

export type DeleteApiTokenMutation = { deleteAPIToken: boolean };

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

export type FlowCreatedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type FlowCreatedSubscription = { flowCreated: FlowFragmentFragment };

export type FlowDeletedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type FlowDeletedSubscription = { flowDeleted: FlowFragmentFragment };

export type FlowUpdatedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type FlowUpdatedSubscription = { flowUpdated: FlowFragmentFragment };

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

export type ProviderCreatedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type ProviderCreatedSubscription = { providerCreated: ProviderConfigFragmentFragment };

export type ProviderUpdatedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type ProviderUpdatedSubscription = { providerUpdated: ProviderConfigFragmentFragment };

export type ProviderDeletedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type ProviderDeletedSubscription = { providerDeleted: ProviderConfigFragmentFragment };

export type ApiTokenCreatedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type ApiTokenCreatedSubscription = { apiTokenCreated: ApiTokenFragmentFragment };

export type ApiTokenUpdatedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type ApiTokenUpdatedSubscription = { apiTokenUpdated: ApiTokenFragmentFragment };

export type ApiTokenDeletedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type ApiTokenDeletedSubscription = { apiTokenDeleted: ApiTokenFragmentFragment };

export type SettingsUserUpdatedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type SettingsUserUpdatedSubscription = { settingsUserUpdated: UserPreferencesFragmentFragment };

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
export const ProviderFragmentFragmentDoc = gql`
    fragment providerFragment on Provider {
        name
        type
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
        provider {
            ...providerFragment
        }
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
        taskId
        subtaskId
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
        taskId
        subtaskId
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
        provider {
            ...providerFragment
        }
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
export const TestResultFragmentFragmentDoc = gql`
    fragment testResultFragment on TestResult {
        name
        type
        result
        reasoning
        streaming
        latency
        error
    }
`;
export const AgentTestResultFragmentFragmentDoc = gql`
    fragment agentTestResultFragment on AgentTestResult {
        tests {
            ...testResultFragment
        }
    }
`;
export const ProviderTestResultFragmentFragmentDoc = gql`
    fragment providerTestResultFragment on ProviderTestResult {
        simple {
            ...agentTestResultFragment
        }
        simpleJson {
            ...agentTestResultFragment
        }
        primaryAgent {
            ...agentTestResultFragment
        }
        assistant {
            ...agentTestResultFragment
        }
        generator {
            ...agentTestResultFragment
        }
        refiner {
            ...agentTestResultFragment
        }
        adviser {
            ...agentTestResultFragment
        }
        reflector {
            ...agentTestResultFragment
        }
        searcher {
            ...agentTestResultFragment
        }
        enricher {
            ...agentTestResultFragment
        }
        coder {
            ...agentTestResultFragment
        }
        installer {
            ...agentTestResultFragment
        }
        pentester {
            ...agentTestResultFragment
        }
    }
`;
export const ModelConfigFragmentFragmentDoc = gql`
    fragment modelConfigFragment on ModelConfig {
        name
        price {
            input
            output
            cacheRead
            cacheWrite
        }
    }
`;
export const AgentConfigFragmentFragmentDoc = gql`
    fragment agentConfigFragment on AgentConfig {
        model
        maxTokens
        temperature
        topK
        topP
        minLength
        maxLength
        repetitionPenalty
        frequencyPenalty
        presencePenalty
        reasoning {
            effort
            maxTokens
        }
        price {
            input
            output
            cacheRead
            cacheWrite
        }
    }
`;
export const AgentsConfigFragmentFragmentDoc = gql`
    fragment agentsConfigFragment on AgentsConfig {
        simple {
            ...agentConfigFragment
        }
        simpleJson {
            ...agentConfigFragment
        }
        primaryAgent {
            ...agentConfigFragment
        }
        assistant {
            ...agentConfigFragment
        }
        generator {
            ...agentConfigFragment
        }
        refiner {
            ...agentConfigFragment
        }
        adviser {
            ...agentConfigFragment
        }
        reflector {
            ...agentConfigFragment
        }
        searcher {
            ...agentConfigFragment
        }
        enricher {
            ...agentConfigFragment
        }
        coder {
            ...agentConfigFragment
        }
        installer {
            ...agentConfigFragment
        }
        pentester {
            ...agentConfigFragment
        }
    }
`;
export const ProviderConfigFragmentFragmentDoc = gql`
    fragment providerConfigFragment on ProviderConfig {
        id
        name
        type
        agents {
            ...agentsConfigFragment
        }
        createdAt
        updatedAt
    }
`;
export const UserPromptFragmentFragmentDoc = gql`
    fragment userPromptFragment on UserPrompt {
        id
        type
        template
        createdAt
        updatedAt
    }
`;
export const DefaultPromptFragmentFragmentDoc = gql`
    fragment defaultPromptFragment on DefaultPrompt {
        type
        template
        variables
    }
`;
export const PromptValidationResultFragmentFragmentDoc = gql`
    fragment promptValidationResultFragment on PromptValidationResult {
        result
        errorType
        message
        line
        details
    }
`;
export const ApiTokenFragmentFragmentDoc = gql`
    fragment apiTokenFragment on APIToken {
        id
        tokenId
        userId
        roleId
        name
        ttl
        status
        createdAt
        updatedAt
    }
`;
export const ApiTokenWithSecretFragmentFragmentDoc = gql`
    fragment apiTokenWithSecretFragment on APITokenWithSecret {
        id
        tokenId
        userId
        roleId
        name
        ttl
        status
        createdAt
        updatedAt
        token
    }
`;
export const UsageStatsFragmentFragmentDoc = gql`
    fragment usageStatsFragment on UsageStats {
        totalUsageIn
        totalUsageOut
        totalUsageCacheIn
        totalUsageCacheOut
        totalUsageCostIn
        totalUsageCostOut
    }
`;
export const DailyUsageStatsFragmentFragmentDoc = gql`
    fragment dailyUsageStatsFragment on DailyUsageStats {
        date
        stats {
            ...usageStatsFragment
        }
    }
`;
export const ProviderUsageStatsFragmentFragmentDoc = gql`
    fragment providerUsageStatsFragment on ProviderUsageStats {
        provider
        stats {
            ...usageStatsFragment
        }
    }
`;
export const ModelUsageStatsFragmentFragmentDoc = gql`
    fragment modelUsageStatsFragment on ModelUsageStats {
        model
        provider
        stats {
            ...usageStatsFragment
        }
    }
`;
export const AgentTypeUsageStatsFragmentFragmentDoc = gql`
    fragment agentTypeUsageStatsFragment on AgentTypeUsageStats {
        agentType
        stats {
            ...usageStatsFragment
        }
    }
`;
export const ToolcallsStatsFragmentFragmentDoc = gql`
    fragment toolcallsStatsFragment on ToolcallsStats {
        totalCount
        totalDurationSeconds
    }
`;
export const DailyToolcallsStatsFragmentFragmentDoc = gql`
    fragment dailyToolcallsStatsFragment on DailyToolcallsStats {
        date
        stats {
            ...toolcallsStatsFragment
        }
    }
`;
export const FunctionToolcallsStatsFragmentFragmentDoc = gql`
    fragment functionToolcallsStatsFragment on FunctionToolcallsStats {
        functionName
        isAgent
        totalCount
        totalDurationSeconds
        avgDurationSeconds
    }
`;
export const FlowStatsFragmentFragmentDoc = gql`
    fragment flowStatsFragment on FlowStats {
        totalTasksCount
        totalSubtasksCount
        totalAssistantsCount
    }
`;
export const FlowsStatsFragmentFragmentDoc = gql`
    fragment flowsStatsFragment on FlowsStats {
        totalFlowsCount
        totalTasksCount
        totalSubtasksCount
        totalAssistantsCount
    }
`;
export const DailyFlowsStatsFragmentFragmentDoc = gql`
    fragment dailyFlowsStatsFragment on DailyFlowsStats {
        date
        stats {
            ...flowsStatsFragment
        }
    }
`;
export const SubtaskExecutionStatsFragmentFragmentDoc = gql`
    fragment subtaskExecutionStatsFragment on SubtaskExecutionStats {
        subtaskId
        subtaskTitle
        totalDurationSeconds
        totalToolcallsCount
    }
`;
export const TaskExecutionStatsFragmentFragmentDoc = gql`
    fragment taskExecutionStatsFragment on TaskExecutionStats {
        taskId
        taskTitle
        totalDurationSeconds
        totalToolcallsCount
        subtasks {
            ...subtaskExecutionStatsFragment
        }
    }
`;
export const FlowExecutionStatsFragmentFragmentDoc = gql`
    fragment flowExecutionStatsFragment on FlowExecutionStats {
        flowId
        flowTitle
        totalDurationSeconds
        totalToolcallsCount
        totalAssistantsCount
        tasks {
            ...taskExecutionStatsFragment
        }
    }
`;
export const UserPreferencesFragmentFragmentDoc = gql`
    fragment userPreferencesFragment on UserPreferences {
        id
        favoriteFlows
    }
`;
export const FlowsDocument = gql`
    query flows {
        flows {
            ...flowFragment
        }
    }
    ${FlowFragmentFragmentDoc}
    ${TerminalFragmentFragmentDoc}
    ${ProviderFragmentFragmentDoc}
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
        providers {
            ...providerFragment
        }
    }
    ${ProviderFragmentFragmentDoc}
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
export const SettingsDocument = gql`
    query settings {
        settings {
            ...settingsFragment
        }
    }
    ${SettingsFragmentFragmentDoc}
`;

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
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<SettingsQuery, SettingsQueryVariables>(SettingsDocument, options);
}
export function useSettingsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<SettingsQuery, SettingsQueryVariables>) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<SettingsQuery, SettingsQueryVariables>(SettingsDocument, options);
}
export function useSettingsSuspenseQuery(
    baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<SettingsQuery, SettingsQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<SettingsQuery, SettingsQueryVariables>(SettingsDocument, options);
}
export type SettingsQueryHookResult = ReturnType<typeof useSettingsQuery>;
export type SettingsLazyQueryHookResult = ReturnType<typeof useSettingsLazyQuery>;
export type SettingsSuspenseQueryHookResult = ReturnType<typeof useSettingsSuspenseQuery>;
export type SettingsQueryResult = Apollo.QueryResult<SettingsQuery, SettingsQueryVariables>;
export const SettingsProvidersDocument = gql`
    query settingsProviders {
        settingsProviders {
            enabled {
                openai
                anthropic
                gemini
                bedrock
                ollama
                custom
            }
            default {
                openai {
                    ...providerConfigFragment
                }
                anthropic {
                    ...providerConfigFragment
                }
                gemini {
                    ...providerConfigFragment
                }
                bedrock {
                    ...providerConfigFragment
                }
                ollama {
                    ...providerConfigFragment
                }
                custom {
                    ...providerConfigFragment
                }
            }
            userDefined {
                ...providerConfigFragment
            }
            models {
                openai {
                    ...modelConfigFragment
                }
                anthropic {
                    ...modelConfigFragment
                }
                gemini {
                    ...modelConfigFragment
                }
                bedrock {
                    ...modelConfigFragment
                }
                ollama {
                    ...modelConfigFragment
                }
                custom {
                    ...modelConfigFragment
                }
            }
        }
    }
    ${ProviderConfigFragmentFragmentDoc}
    ${AgentsConfigFragmentFragmentDoc}
    ${AgentConfigFragmentFragmentDoc}
    ${ModelConfigFragmentFragmentDoc}
`;

/**
 * __useSettingsProvidersQuery__
 *
 * To run a query within a React component, call `useSettingsProvidersQuery` and pass it any options that fit your needs.
 * When your component renders, `useSettingsProvidersQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSettingsProvidersQuery({
 *   variables: {
 *   },
 * });
 */
export function useSettingsProvidersQuery(
    baseOptions?: Apollo.QueryHookOptions<SettingsProvidersQuery, SettingsProvidersQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<SettingsProvidersQuery, SettingsProvidersQueryVariables>(SettingsProvidersDocument, options);
}
export function useSettingsProvidersLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<SettingsProvidersQuery, SettingsProvidersQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<SettingsProvidersQuery, SettingsProvidersQueryVariables>(
        SettingsProvidersDocument,
        options,
    );
}
export function useSettingsProvidersSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<SettingsProvidersQuery, SettingsProvidersQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<SettingsProvidersQuery, SettingsProvidersQueryVariables>(
        SettingsProvidersDocument,
        options,
    );
}
export type SettingsProvidersQueryHookResult = ReturnType<typeof useSettingsProvidersQuery>;
export type SettingsProvidersLazyQueryHookResult = ReturnType<typeof useSettingsProvidersLazyQuery>;
export type SettingsProvidersSuspenseQueryHookResult = ReturnType<typeof useSettingsProvidersSuspenseQuery>;
export type SettingsProvidersQueryResult = Apollo.QueryResult<SettingsProvidersQuery, SettingsProvidersQueryVariables>;
export const SettingsPromptsDocument = gql`
    query settingsPrompts {
        settingsPrompts {
            default {
                agents {
                    primaryAgent {
                        system {
                            ...defaultPromptFragment
                        }
                    }
                    assistant {
                        system {
                            ...defaultPromptFragment
                        }
                    }
                    pentester {
                        system {
                            ...defaultPromptFragment
                        }
                        human {
                            ...defaultPromptFragment
                        }
                    }
                    coder {
                        system {
                            ...defaultPromptFragment
                        }
                        human {
                            ...defaultPromptFragment
                        }
                    }
                    installer {
                        system {
                            ...defaultPromptFragment
                        }
                        human {
                            ...defaultPromptFragment
                        }
                    }
                    searcher {
                        system {
                            ...defaultPromptFragment
                        }
                        human {
                            ...defaultPromptFragment
                        }
                    }
                    memorist {
                        system {
                            ...defaultPromptFragment
                        }
                        human {
                            ...defaultPromptFragment
                        }
                    }
                    adviser {
                        system {
                            ...defaultPromptFragment
                        }
                        human {
                            ...defaultPromptFragment
                        }
                    }
                    generator {
                        system {
                            ...defaultPromptFragment
                        }
                        human {
                            ...defaultPromptFragment
                        }
                    }
                    refiner {
                        system {
                            ...defaultPromptFragment
                        }
                        human {
                            ...defaultPromptFragment
                        }
                    }
                    reporter {
                        system {
                            ...defaultPromptFragment
                        }
                        human {
                            ...defaultPromptFragment
                        }
                    }
                    reflector {
                        system {
                            ...defaultPromptFragment
                        }
                        human {
                            ...defaultPromptFragment
                        }
                    }
                    enricher {
                        system {
                            ...defaultPromptFragment
                        }
                        human {
                            ...defaultPromptFragment
                        }
                    }
                    toolCallFixer {
                        system {
                            ...defaultPromptFragment
                        }
                        human {
                            ...defaultPromptFragment
                        }
                    }
                    summarizer {
                        system {
                            ...defaultPromptFragment
                        }
                    }
                }
                tools {
                    getFlowDescription {
                        ...defaultPromptFragment
                    }
                    getTaskDescription {
                        ...defaultPromptFragment
                    }
                    getExecutionLogs {
                        ...defaultPromptFragment
                    }
                    getFullExecutionContext {
                        ...defaultPromptFragment
                    }
                    getShortExecutionContext {
                        ...defaultPromptFragment
                    }
                    chooseDockerImage {
                        ...defaultPromptFragment
                    }
                    chooseUserLanguage {
                        ...defaultPromptFragment
                    }
                    collectToolCallID {
                        ...defaultPromptFragment
                    }
                    detectToolCallIDPattern {
                        ...defaultPromptFragment
                    }
                }
            }
            userDefined {
                ...userPromptFragment
            }
        }
    }
    ${DefaultPromptFragmentFragmentDoc}
    ${UserPromptFragmentFragmentDoc}
`;

/**
 * __useSettingsPromptsQuery__
 *
 * To run a query within a React component, call `useSettingsPromptsQuery` and pass it any options that fit your needs.
 * When your component renders, `useSettingsPromptsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSettingsPromptsQuery({
 *   variables: {
 *   },
 * });
 */
export function useSettingsPromptsQuery(
    baseOptions?: Apollo.QueryHookOptions<SettingsPromptsQuery, SettingsPromptsQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<SettingsPromptsQuery, SettingsPromptsQueryVariables>(SettingsPromptsDocument, options);
}
export function useSettingsPromptsLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<SettingsPromptsQuery, SettingsPromptsQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<SettingsPromptsQuery, SettingsPromptsQueryVariables>(SettingsPromptsDocument, options);
}
export function useSettingsPromptsSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<SettingsPromptsQuery, SettingsPromptsQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<SettingsPromptsQuery, SettingsPromptsQueryVariables>(
        SettingsPromptsDocument,
        options,
    );
}
export type SettingsPromptsQueryHookResult = ReturnType<typeof useSettingsPromptsQuery>;
export type SettingsPromptsLazyQueryHookResult = ReturnType<typeof useSettingsPromptsLazyQuery>;
export type SettingsPromptsSuspenseQueryHookResult = ReturnType<typeof useSettingsPromptsSuspenseQuery>;
export type SettingsPromptsQueryResult = Apollo.QueryResult<SettingsPromptsQuery, SettingsPromptsQueryVariables>;
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
    ${ProviderFragmentFragmentDoc}
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
export const AssistantsDocument = gql`
    query assistants($flowId: ID!) {
        assistants(flowId: $flowId) {
            ...assistantFragment
        }
    }
    ${AssistantFragmentFragmentDoc}
    ${ProviderFragmentFragmentDoc}
`;

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
export function useAssistantsQuery(
    baseOptions: Apollo.QueryHookOptions<AssistantsQuery, AssistantsQueryVariables> &
        ({ variables: AssistantsQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<AssistantsQuery, AssistantsQueryVariables>(AssistantsDocument, options);
}
export function useAssistantsLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<AssistantsQuery, AssistantsQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<AssistantsQuery, AssistantsQueryVariables>(AssistantsDocument, options);
}
export function useAssistantsSuspenseQuery(
    baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AssistantsQuery, AssistantsQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
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
    ${AssistantLogFragmentFragmentDoc}
`;

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
export function useAssistantLogsQuery(
    baseOptions: Apollo.QueryHookOptions<AssistantLogsQuery, AssistantLogsQueryVariables> &
        ({ variables: AssistantLogsQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<AssistantLogsQuery, AssistantLogsQueryVariables>(AssistantLogsDocument, options);
}
export function useAssistantLogsLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<AssistantLogsQuery, AssistantLogsQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<AssistantLogsQuery, AssistantLogsQueryVariables>(AssistantLogsDocument, options);
}
export function useAssistantLogsSuspenseQuery(
    baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AssistantLogsQuery, AssistantLogsQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<AssistantLogsQuery, AssistantLogsQueryVariables>(AssistantLogsDocument, options);
}
export type AssistantLogsQueryHookResult = ReturnType<typeof useAssistantLogsQuery>;
export type AssistantLogsLazyQueryHookResult = ReturnType<typeof useAssistantLogsLazyQuery>;
export type AssistantLogsSuspenseQueryHookResult = ReturnType<typeof useAssistantLogsSuspenseQuery>;
export type AssistantLogsQueryResult = Apollo.QueryResult<AssistantLogsQuery, AssistantLogsQueryVariables>;
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
    ${ProviderFragmentFragmentDoc}
    ${TaskFragmentFragmentDoc}
    ${SubtaskFragmentFragmentDoc}
`;

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
export function useFlowReportQuery(
    baseOptions: Apollo.QueryHookOptions<FlowReportQuery, FlowReportQueryVariables> &
        ({ variables: FlowReportQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<FlowReportQuery, FlowReportQueryVariables>(FlowReportDocument, options);
}
export function useFlowReportLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<FlowReportQuery, FlowReportQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<FlowReportQuery, FlowReportQueryVariables>(FlowReportDocument, options);
}
export function useFlowReportSuspenseQuery(
    baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<FlowReportQuery, FlowReportQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<FlowReportQuery, FlowReportQueryVariables>(FlowReportDocument, options);
}
export type FlowReportQueryHookResult = ReturnType<typeof useFlowReportQuery>;
export type FlowReportLazyQueryHookResult = ReturnType<typeof useFlowReportLazyQuery>;
export type FlowReportSuspenseQueryHookResult = ReturnType<typeof useFlowReportSuspenseQuery>;
export type FlowReportQueryResult = Apollo.QueryResult<FlowReportQuery, FlowReportQueryVariables>;
export const UsageStatsTotalDocument = gql`
    query usageStatsTotal {
        usageStatsTotal {
            ...usageStatsFragment
        }
    }
    ${UsageStatsFragmentFragmentDoc}
`;

/**
 * __useUsageStatsTotalQuery__
 *
 * To run a query within a React component, call `useUsageStatsTotalQuery` and pass it any options that fit your needs.
 * When your component renders, `useUsageStatsTotalQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useUsageStatsTotalQuery({
 *   variables: {
 *   },
 * });
 */
export function useUsageStatsTotalQuery(
    baseOptions?: Apollo.QueryHookOptions<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>(UsageStatsTotalDocument, options);
}
export function useUsageStatsTotalLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>(UsageStatsTotalDocument, options);
}
export function useUsageStatsTotalSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>(
        UsageStatsTotalDocument,
        options,
    );
}
export type UsageStatsTotalQueryHookResult = ReturnType<typeof useUsageStatsTotalQuery>;
export type UsageStatsTotalLazyQueryHookResult = ReturnType<typeof useUsageStatsTotalLazyQuery>;
export type UsageStatsTotalSuspenseQueryHookResult = ReturnType<typeof useUsageStatsTotalSuspenseQuery>;
export type UsageStatsTotalQueryResult = Apollo.QueryResult<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>;
export const UsageStatsByPeriodDocument = gql`
    query usageStatsByPeriod($period: UsageStatsPeriod!) {
        usageStatsByPeriod(period: $period) {
            ...dailyUsageStatsFragment
        }
    }
    ${DailyUsageStatsFragmentFragmentDoc}
    ${UsageStatsFragmentFragmentDoc}
`;

/**
 * __useUsageStatsByPeriodQuery__
 *
 * To run a query within a React component, call `useUsageStatsByPeriodQuery` and pass it any options that fit your needs.
 * When your component renders, `useUsageStatsByPeriodQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useUsageStatsByPeriodQuery({
 *   variables: {
 *      period: // value for 'period'
 *   },
 * });
 */
export function useUsageStatsByPeriodQuery(
    baseOptions: Apollo.QueryHookOptions<UsageStatsByPeriodQuery, UsageStatsByPeriodQueryVariables> &
        ({ variables: UsageStatsByPeriodQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<UsageStatsByPeriodQuery, UsageStatsByPeriodQueryVariables>(
        UsageStatsByPeriodDocument,
        options,
    );
}
export function useUsageStatsByPeriodLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<UsageStatsByPeriodQuery, UsageStatsByPeriodQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<UsageStatsByPeriodQuery, UsageStatsByPeriodQueryVariables>(
        UsageStatsByPeriodDocument,
        options,
    );
}
export function useUsageStatsByPeriodSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<UsageStatsByPeriodQuery, UsageStatsByPeriodQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<UsageStatsByPeriodQuery, UsageStatsByPeriodQueryVariables>(
        UsageStatsByPeriodDocument,
        options,
    );
}
export type UsageStatsByPeriodQueryHookResult = ReturnType<typeof useUsageStatsByPeriodQuery>;
export type UsageStatsByPeriodLazyQueryHookResult = ReturnType<typeof useUsageStatsByPeriodLazyQuery>;
export type UsageStatsByPeriodSuspenseQueryHookResult = ReturnType<typeof useUsageStatsByPeriodSuspenseQuery>;
export type UsageStatsByPeriodQueryResult = Apollo.QueryResult<
    UsageStatsByPeriodQuery,
    UsageStatsByPeriodQueryVariables
>;
export const UsageStatsByProviderDocument = gql`
    query usageStatsByProvider {
        usageStatsByProvider {
            ...providerUsageStatsFragment
        }
    }
    ${ProviderUsageStatsFragmentFragmentDoc}
    ${UsageStatsFragmentFragmentDoc}
`;

/**
 * __useUsageStatsByProviderQuery__
 *
 * To run a query within a React component, call `useUsageStatsByProviderQuery` and pass it any options that fit your needs.
 * When your component renders, `useUsageStatsByProviderQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useUsageStatsByProviderQuery({
 *   variables: {
 *   },
 * });
 */
export function useUsageStatsByProviderQuery(
    baseOptions?: Apollo.QueryHookOptions<UsageStatsByProviderQuery, UsageStatsByProviderQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<UsageStatsByProviderQuery, UsageStatsByProviderQueryVariables>(
        UsageStatsByProviderDocument,
        options,
    );
}
export function useUsageStatsByProviderLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<UsageStatsByProviderQuery, UsageStatsByProviderQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<UsageStatsByProviderQuery, UsageStatsByProviderQueryVariables>(
        UsageStatsByProviderDocument,
        options,
    );
}
export function useUsageStatsByProviderSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<UsageStatsByProviderQuery, UsageStatsByProviderQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<UsageStatsByProviderQuery, UsageStatsByProviderQueryVariables>(
        UsageStatsByProviderDocument,
        options,
    );
}
export type UsageStatsByProviderQueryHookResult = ReturnType<typeof useUsageStatsByProviderQuery>;
export type UsageStatsByProviderLazyQueryHookResult = ReturnType<typeof useUsageStatsByProviderLazyQuery>;
export type UsageStatsByProviderSuspenseQueryHookResult = ReturnType<typeof useUsageStatsByProviderSuspenseQuery>;
export type UsageStatsByProviderQueryResult = Apollo.QueryResult<
    UsageStatsByProviderQuery,
    UsageStatsByProviderQueryVariables
>;
export const UsageStatsByModelDocument = gql`
    query usageStatsByModel {
        usageStatsByModel {
            ...modelUsageStatsFragment
        }
    }
    ${ModelUsageStatsFragmentFragmentDoc}
    ${UsageStatsFragmentFragmentDoc}
`;

/**
 * __useUsageStatsByModelQuery__
 *
 * To run a query within a React component, call `useUsageStatsByModelQuery` and pass it any options that fit your needs.
 * When your component renders, `useUsageStatsByModelQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useUsageStatsByModelQuery({
 *   variables: {
 *   },
 * });
 */
export function useUsageStatsByModelQuery(
    baseOptions?: Apollo.QueryHookOptions<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>(UsageStatsByModelDocument, options);
}
export function useUsageStatsByModelLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>(
        UsageStatsByModelDocument,
        options,
    );
}
export function useUsageStatsByModelSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>(
        UsageStatsByModelDocument,
        options,
    );
}
export type UsageStatsByModelQueryHookResult = ReturnType<typeof useUsageStatsByModelQuery>;
export type UsageStatsByModelLazyQueryHookResult = ReturnType<typeof useUsageStatsByModelLazyQuery>;
export type UsageStatsByModelSuspenseQueryHookResult = ReturnType<typeof useUsageStatsByModelSuspenseQuery>;
export type UsageStatsByModelQueryResult = Apollo.QueryResult<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>;
export const UsageStatsByAgentTypeDocument = gql`
    query usageStatsByAgentType {
        usageStatsByAgentType {
            ...agentTypeUsageStatsFragment
        }
    }
    ${AgentTypeUsageStatsFragmentFragmentDoc}
    ${UsageStatsFragmentFragmentDoc}
`;

/**
 * __useUsageStatsByAgentTypeQuery__
 *
 * To run a query within a React component, call `useUsageStatsByAgentTypeQuery` and pass it any options that fit your needs.
 * When your component renders, `useUsageStatsByAgentTypeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useUsageStatsByAgentTypeQuery({
 *   variables: {
 *   },
 * });
 */
export function useUsageStatsByAgentTypeQuery(
    baseOptions?: Apollo.QueryHookOptions<UsageStatsByAgentTypeQuery, UsageStatsByAgentTypeQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<UsageStatsByAgentTypeQuery, UsageStatsByAgentTypeQueryVariables>(
        UsageStatsByAgentTypeDocument,
        options,
    );
}
export function useUsageStatsByAgentTypeLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<UsageStatsByAgentTypeQuery, UsageStatsByAgentTypeQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<UsageStatsByAgentTypeQuery, UsageStatsByAgentTypeQueryVariables>(
        UsageStatsByAgentTypeDocument,
        options,
    );
}
export function useUsageStatsByAgentTypeSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<UsageStatsByAgentTypeQuery, UsageStatsByAgentTypeQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<UsageStatsByAgentTypeQuery, UsageStatsByAgentTypeQueryVariables>(
        UsageStatsByAgentTypeDocument,
        options,
    );
}
export type UsageStatsByAgentTypeQueryHookResult = ReturnType<typeof useUsageStatsByAgentTypeQuery>;
export type UsageStatsByAgentTypeLazyQueryHookResult = ReturnType<typeof useUsageStatsByAgentTypeLazyQuery>;
export type UsageStatsByAgentTypeSuspenseQueryHookResult = ReturnType<typeof useUsageStatsByAgentTypeSuspenseQuery>;
export type UsageStatsByAgentTypeQueryResult = Apollo.QueryResult<
    UsageStatsByAgentTypeQuery,
    UsageStatsByAgentTypeQueryVariables
>;
export const UsageStatsByFlowDocument = gql`
    query usageStatsByFlow($flowId: ID!) {
        usageStatsByFlow(flowId: $flowId) {
            ...usageStatsFragment
        }
    }
    ${UsageStatsFragmentFragmentDoc}
`;

/**
 * __useUsageStatsByFlowQuery__
 *
 * To run a query within a React component, call `useUsageStatsByFlowQuery` and pass it any options that fit your needs.
 * When your component renders, `useUsageStatsByFlowQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useUsageStatsByFlowQuery({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useUsageStatsByFlowQuery(
    baseOptions: Apollo.QueryHookOptions<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables> &
        ({ variables: UsageStatsByFlowQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables>(UsageStatsByFlowDocument, options);
}
export function useUsageStatsByFlowLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables>(
        UsageStatsByFlowDocument,
        options,
    );
}
export function useUsageStatsByFlowSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables>(
        UsageStatsByFlowDocument,
        options,
    );
}
export type UsageStatsByFlowQueryHookResult = ReturnType<typeof useUsageStatsByFlowQuery>;
export type UsageStatsByFlowLazyQueryHookResult = ReturnType<typeof useUsageStatsByFlowLazyQuery>;
export type UsageStatsByFlowSuspenseQueryHookResult = ReturnType<typeof useUsageStatsByFlowSuspenseQuery>;
export type UsageStatsByFlowQueryResult = Apollo.QueryResult<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables>;
export const UsageStatsByAgentTypeForFlowDocument = gql`
    query usageStatsByAgentTypeForFlow($flowId: ID!) {
        usageStatsByAgentTypeForFlow(flowId: $flowId) {
            ...agentTypeUsageStatsFragment
        }
    }
    ${AgentTypeUsageStatsFragmentFragmentDoc}
    ${UsageStatsFragmentFragmentDoc}
`;

/**
 * __useUsageStatsByAgentTypeForFlowQuery__
 *
 * To run a query within a React component, call `useUsageStatsByAgentTypeForFlowQuery` and pass it any options that fit your needs.
 * When your component renders, `useUsageStatsByAgentTypeForFlowQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useUsageStatsByAgentTypeForFlowQuery({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useUsageStatsByAgentTypeForFlowQuery(
    baseOptions: Apollo.QueryHookOptions<
        UsageStatsByAgentTypeForFlowQuery,
        UsageStatsByAgentTypeForFlowQueryVariables
    > &
        ({ variables: UsageStatsByAgentTypeForFlowQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<UsageStatsByAgentTypeForFlowQuery, UsageStatsByAgentTypeForFlowQueryVariables>(
        UsageStatsByAgentTypeForFlowDocument,
        options,
    );
}
export function useUsageStatsByAgentTypeForFlowLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<
        UsageStatsByAgentTypeForFlowQuery,
        UsageStatsByAgentTypeForFlowQueryVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<UsageStatsByAgentTypeForFlowQuery, UsageStatsByAgentTypeForFlowQueryVariables>(
        UsageStatsByAgentTypeForFlowDocument,
        options,
    );
}
export function useUsageStatsByAgentTypeForFlowSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<
              UsageStatsByAgentTypeForFlowQuery,
              UsageStatsByAgentTypeForFlowQueryVariables
          >,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<UsageStatsByAgentTypeForFlowQuery, UsageStatsByAgentTypeForFlowQueryVariables>(
        UsageStatsByAgentTypeForFlowDocument,
        options,
    );
}
export type UsageStatsByAgentTypeForFlowQueryHookResult = ReturnType<typeof useUsageStatsByAgentTypeForFlowQuery>;
export type UsageStatsByAgentTypeForFlowLazyQueryHookResult = ReturnType<
    typeof useUsageStatsByAgentTypeForFlowLazyQuery
>;
export type UsageStatsByAgentTypeForFlowSuspenseQueryHookResult = ReturnType<
    typeof useUsageStatsByAgentTypeForFlowSuspenseQuery
>;
export type UsageStatsByAgentTypeForFlowQueryResult = Apollo.QueryResult<
    UsageStatsByAgentTypeForFlowQuery,
    UsageStatsByAgentTypeForFlowQueryVariables
>;
export const ToolcallsStatsTotalDocument = gql`
    query toolcallsStatsTotal {
        toolcallsStatsTotal {
            ...toolcallsStatsFragment
        }
    }
    ${ToolcallsStatsFragmentFragmentDoc}
`;

/**
 * __useToolcallsStatsTotalQuery__
 *
 * To run a query within a React component, call `useToolcallsStatsTotalQuery` and pass it any options that fit your needs.
 * When your component renders, `useToolcallsStatsTotalQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useToolcallsStatsTotalQuery({
 *   variables: {
 *   },
 * });
 */
export function useToolcallsStatsTotalQuery(
    baseOptions?: Apollo.QueryHookOptions<ToolcallsStatsTotalQuery, ToolcallsStatsTotalQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<ToolcallsStatsTotalQuery, ToolcallsStatsTotalQueryVariables>(
        ToolcallsStatsTotalDocument,
        options,
    );
}
export function useToolcallsStatsTotalLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<ToolcallsStatsTotalQuery, ToolcallsStatsTotalQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<ToolcallsStatsTotalQuery, ToolcallsStatsTotalQueryVariables>(
        ToolcallsStatsTotalDocument,
        options,
    );
}
export function useToolcallsStatsTotalSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<ToolcallsStatsTotalQuery, ToolcallsStatsTotalQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<ToolcallsStatsTotalQuery, ToolcallsStatsTotalQueryVariables>(
        ToolcallsStatsTotalDocument,
        options,
    );
}
export type ToolcallsStatsTotalQueryHookResult = ReturnType<typeof useToolcallsStatsTotalQuery>;
export type ToolcallsStatsTotalLazyQueryHookResult = ReturnType<typeof useToolcallsStatsTotalLazyQuery>;
export type ToolcallsStatsTotalSuspenseQueryHookResult = ReturnType<typeof useToolcallsStatsTotalSuspenseQuery>;
export type ToolcallsStatsTotalQueryResult = Apollo.QueryResult<
    ToolcallsStatsTotalQuery,
    ToolcallsStatsTotalQueryVariables
>;
export const ToolcallsStatsByPeriodDocument = gql`
    query toolcallsStatsByPeriod($period: UsageStatsPeriod!) {
        toolcallsStatsByPeriod(period: $period) {
            ...dailyToolcallsStatsFragment
        }
    }
    ${DailyToolcallsStatsFragmentFragmentDoc}
    ${ToolcallsStatsFragmentFragmentDoc}
`;

/**
 * __useToolcallsStatsByPeriodQuery__
 *
 * To run a query within a React component, call `useToolcallsStatsByPeriodQuery` and pass it any options that fit your needs.
 * When your component renders, `useToolcallsStatsByPeriodQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useToolcallsStatsByPeriodQuery({
 *   variables: {
 *      period: // value for 'period'
 *   },
 * });
 */
export function useToolcallsStatsByPeriodQuery(
    baseOptions: Apollo.QueryHookOptions<ToolcallsStatsByPeriodQuery, ToolcallsStatsByPeriodQueryVariables> &
        ({ variables: ToolcallsStatsByPeriodQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<ToolcallsStatsByPeriodQuery, ToolcallsStatsByPeriodQueryVariables>(
        ToolcallsStatsByPeriodDocument,
        options,
    );
}
export function useToolcallsStatsByPeriodLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<ToolcallsStatsByPeriodQuery, ToolcallsStatsByPeriodQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<ToolcallsStatsByPeriodQuery, ToolcallsStatsByPeriodQueryVariables>(
        ToolcallsStatsByPeriodDocument,
        options,
    );
}
export function useToolcallsStatsByPeriodSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<ToolcallsStatsByPeriodQuery, ToolcallsStatsByPeriodQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<ToolcallsStatsByPeriodQuery, ToolcallsStatsByPeriodQueryVariables>(
        ToolcallsStatsByPeriodDocument,
        options,
    );
}
export type ToolcallsStatsByPeriodQueryHookResult = ReturnType<typeof useToolcallsStatsByPeriodQuery>;
export type ToolcallsStatsByPeriodLazyQueryHookResult = ReturnType<typeof useToolcallsStatsByPeriodLazyQuery>;
export type ToolcallsStatsByPeriodSuspenseQueryHookResult = ReturnType<typeof useToolcallsStatsByPeriodSuspenseQuery>;
export type ToolcallsStatsByPeriodQueryResult = Apollo.QueryResult<
    ToolcallsStatsByPeriodQuery,
    ToolcallsStatsByPeriodQueryVariables
>;
export const ToolcallsStatsByFunctionDocument = gql`
    query toolcallsStatsByFunction {
        toolcallsStatsByFunction {
            ...functionToolcallsStatsFragment
        }
    }
    ${FunctionToolcallsStatsFragmentFragmentDoc}
`;

/**
 * __useToolcallsStatsByFunctionQuery__
 *
 * To run a query within a React component, call `useToolcallsStatsByFunctionQuery` and pass it any options that fit your needs.
 * When your component renders, `useToolcallsStatsByFunctionQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useToolcallsStatsByFunctionQuery({
 *   variables: {
 *   },
 * });
 */
export function useToolcallsStatsByFunctionQuery(
    baseOptions?: Apollo.QueryHookOptions<ToolcallsStatsByFunctionQuery, ToolcallsStatsByFunctionQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<ToolcallsStatsByFunctionQuery, ToolcallsStatsByFunctionQueryVariables>(
        ToolcallsStatsByFunctionDocument,
        options,
    );
}
export function useToolcallsStatsByFunctionLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<ToolcallsStatsByFunctionQuery, ToolcallsStatsByFunctionQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<ToolcallsStatsByFunctionQuery, ToolcallsStatsByFunctionQueryVariables>(
        ToolcallsStatsByFunctionDocument,
        options,
    );
}
export function useToolcallsStatsByFunctionSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<ToolcallsStatsByFunctionQuery, ToolcallsStatsByFunctionQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<ToolcallsStatsByFunctionQuery, ToolcallsStatsByFunctionQueryVariables>(
        ToolcallsStatsByFunctionDocument,
        options,
    );
}
export type ToolcallsStatsByFunctionQueryHookResult = ReturnType<typeof useToolcallsStatsByFunctionQuery>;
export type ToolcallsStatsByFunctionLazyQueryHookResult = ReturnType<typeof useToolcallsStatsByFunctionLazyQuery>;
export type ToolcallsStatsByFunctionSuspenseQueryHookResult = ReturnType<
    typeof useToolcallsStatsByFunctionSuspenseQuery
>;
export type ToolcallsStatsByFunctionQueryResult = Apollo.QueryResult<
    ToolcallsStatsByFunctionQuery,
    ToolcallsStatsByFunctionQueryVariables
>;
export const ToolcallsStatsByFlowDocument = gql`
    query toolcallsStatsByFlow($flowId: ID!) {
        toolcallsStatsByFlow(flowId: $flowId) {
            ...toolcallsStatsFragment
        }
    }
    ${ToolcallsStatsFragmentFragmentDoc}
`;

/**
 * __useToolcallsStatsByFlowQuery__
 *
 * To run a query within a React component, call `useToolcallsStatsByFlowQuery` and pass it any options that fit your needs.
 * When your component renders, `useToolcallsStatsByFlowQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useToolcallsStatsByFlowQuery({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useToolcallsStatsByFlowQuery(
    baseOptions: Apollo.QueryHookOptions<ToolcallsStatsByFlowQuery, ToolcallsStatsByFlowQueryVariables> &
        ({ variables: ToolcallsStatsByFlowQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<ToolcallsStatsByFlowQuery, ToolcallsStatsByFlowQueryVariables>(
        ToolcallsStatsByFlowDocument,
        options,
    );
}
export function useToolcallsStatsByFlowLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<ToolcallsStatsByFlowQuery, ToolcallsStatsByFlowQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<ToolcallsStatsByFlowQuery, ToolcallsStatsByFlowQueryVariables>(
        ToolcallsStatsByFlowDocument,
        options,
    );
}
export function useToolcallsStatsByFlowSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<ToolcallsStatsByFlowQuery, ToolcallsStatsByFlowQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<ToolcallsStatsByFlowQuery, ToolcallsStatsByFlowQueryVariables>(
        ToolcallsStatsByFlowDocument,
        options,
    );
}
export type ToolcallsStatsByFlowQueryHookResult = ReturnType<typeof useToolcallsStatsByFlowQuery>;
export type ToolcallsStatsByFlowLazyQueryHookResult = ReturnType<typeof useToolcallsStatsByFlowLazyQuery>;
export type ToolcallsStatsByFlowSuspenseQueryHookResult = ReturnType<typeof useToolcallsStatsByFlowSuspenseQuery>;
export type ToolcallsStatsByFlowQueryResult = Apollo.QueryResult<
    ToolcallsStatsByFlowQuery,
    ToolcallsStatsByFlowQueryVariables
>;
export const ToolcallsStatsByFunctionForFlowDocument = gql`
    query toolcallsStatsByFunctionForFlow($flowId: ID!) {
        toolcallsStatsByFunctionForFlow(flowId: $flowId) {
            ...functionToolcallsStatsFragment
        }
    }
    ${FunctionToolcallsStatsFragmentFragmentDoc}
`;

/**
 * __useToolcallsStatsByFunctionForFlowQuery__
 *
 * To run a query within a React component, call `useToolcallsStatsByFunctionForFlowQuery` and pass it any options that fit your needs.
 * When your component renders, `useToolcallsStatsByFunctionForFlowQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useToolcallsStatsByFunctionForFlowQuery({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useToolcallsStatsByFunctionForFlowQuery(
    baseOptions: Apollo.QueryHookOptions<
        ToolcallsStatsByFunctionForFlowQuery,
        ToolcallsStatsByFunctionForFlowQueryVariables
    > &
        ({ variables: ToolcallsStatsByFunctionForFlowQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<ToolcallsStatsByFunctionForFlowQuery, ToolcallsStatsByFunctionForFlowQueryVariables>(
        ToolcallsStatsByFunctionForFlowDocument,
        options,
    );
}
export function useToolcallsStatsByFunctionForFlowLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<
        ToolcallsStatsByFunctionForFlowQuery,
        ToolcallsStatsByFunctionForFlowQueryVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<ToolcallsStatsByFunctionForFlowQuery, ToolcallsStatsByFunctionForFlowQueryVariables>(
        ToolcallsStatsByFunctionForFlowDocument,
        options,
    );
}
export function useToolcallsStatsByFunctionForFlowSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<
              ToolcallsStatsByFunctionForFlowQuery,
              ToolcallsStatsByFunctionForFlowQueryVariables
          >,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<ToolcallsStatsByFunctionForFlowQuery, ToolcallsStatsByFunctionForFlowQueryVariables>(
        ToolcallsStatsByFunctionForFlowDocument,
        options,
    );
}
export type ToolcallsStatsByFunctionForFlowQueryHookResult = ReturnType<typeof useToolcallsStatsByFunctionForFlowQuery>;
export type ToolcallsStatsByFunctionForFlowLazyQueryHookResult = ReturnType<
    typeof useToolcallsStatsByFunctionForFlowLazyQuery
>;
export type ToolcallsStatsByFunctionForFlowSuspenseQueryHookResult = ReturnType<
    typeof useToolcallsStatsByFunctionForFlowSuspenseQuery
>;
export type ToolcallsStatsByFunctionForFlowQueryResult = Apollo.QueryResult<
    ToolcallsStatsByFunctionForFlowQuery,
    ToolcallsStatsByFunctionForFlowQueryVariables
>;
export const FlowsStatsTotalDocument = gql`
    query flowsStatsTotal {
        flowsStatsTotal {
            ...flowsStatsFragment
        }
    }
    ${FlowsStatsFragmentFragmentDoc}
`;

/**
 * __useFlowsStatsTotalQuery__
 *
 * To run a query within a React component, call `useFlowsStatsTotalQuery` and pass it any options that fit your needs.
 * When your component renders, `useFlowsStatsTotalQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowsStatsTotalQuery({
 *   variables: {
 *   },
 * });
 */
export function useFlowsStatsTotalQuery(
    baseOptions?: Apollo.QueryHookOptions<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>(FlowsStatsTotalDocument, options);
}
export function useFlowsStatsTotalLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>(FlowsStatsTotalDocument, options);
}
export function useFlowsStatsTotalSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>(
        FlowsStatsTotalDocument,
        options,
    );
}
export type FlowsStatsTotalQueryHookResult = ReturnType<typeof useFlowsStatsTotalQuery>;
export type FlowsStatsTotalLazyQueryHookResult = ReturnType<typeof useFlowsStatsTotalLazyQuery>;
export type FlowsStatsTotalSuspenseQueryHookResult = ReturnType<typeof useFlowsStatsTotalSuspenseQuery>;
export type FlowsStatsTotalQueryResult = Apollo.QueryResult<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>;
export const FlowsStatsByPeriodDocument = gql`
    query flowsStatsByPeriod($period: UsageStatsPeriod!) {
        flowsStatsByPeriod(period: $period) {
            ...dailyFlowsStatsFragment
        }
    }
    ${DailyFlowsStatsFragmentFragmentDoc}
    ${FlowsStatsFragmentFragmentDoc}
`;

/**
 * __useFlowsStatsByPeriodQuery__
 *
 * To run a query within a React component, call `useFlowsStatsByPeriodQuery` and pass it any options that fit your needs.
 * When your component renders, `useFlowsStatsByPeriodQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowsStatsByPeriodQuery({
 *   variables: {
 *      period: // value for 'period'
 *   },
 * });
 */
export function useFlowsStatsByPeriodQuery(
    baseOptions: Apollo.QueryHookOptions<FlowsStatsByPeriodQuery, FlowsStatsByPeriodQueryVariables> &
        ({ variables: FlowsStatsByPeriodQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<FlowsStatsByPeriodQuery, FlowsStatsByPeriodQueryVariables>(
        FlowsStatsByPeriodDocument,
        options,
    );
}
export function useFlowsStatsByPeriodLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<FlowsStatsByPeriodQuery, FlowsStatsByPeriodQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<FlowsStatsByPeriodQuery, FlowsStatsByPeriodQueryVariables>(
        FlowsStatsByPeriodDocument,
        options,
    );
}
export function useFlowsStatsByPeriodSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<FlowsStatsByPeriodQuery, FlowsStatsByPeriodQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<FlowsStatsByPeriodQuery, FlowsStatsByPeriodQueryVariables>(
        FlowsStatsByPeriodDocument,
        options,
    );
}
export type FlowsStatsByPeriodQueryHookResult = ReturnType<typeof useFlowsStatsByPeriodQuery>;
export type FlowsStatsByPeriodLazyQueryHookResult = ReturnType<typeof useFlowsStatsByPeriodLazyQuery>;
export type FlowsStatsByPeriodSuspenseQueryHookResult = ReturnType<typeof useFlowsStatsByPeriodSuspenseQuery>;
export type FlowsStatsByPeriodQueryResult = Apollo.QueryResult<
    FlowsStatsByPeriodQuery,
    FlowsStatsByPeriodQueryVariables
>;
export const FlowStatsByFlowDocument = gql`
    query flowStatsByFlow($flowId: ID!) {
        flowStatsByFlow(flowId: $flowId) {
            ...flowStatsFragment
        }
    }
    ${FlowStatsFragmentFragmentDoc}
`;

/**
 * __useFlowStatsByFlowQuery__
 *
 * To run a query within a React component, call `useFlowStatsByFlowQuery` and pass it any options that fit your needs.
 * When your component renders, `useFlowStatsByFlowQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowStatsByFlowQuery({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useFlowStatsByFlowQuery(
    baseOptions: Apollo.QueryHookOptions<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables> &
        ({ variables: FlowStatsByFlowQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables>(FlowStatsByFlowDocument, options);
}
export function useFlowStatsByFlowLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables>(FlowStatsByFlowDocument, options);
}
export function useFlowStatsByFlowSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables>(
        FlowStatsByFlowDocument,
        options,
    );
}
export type FlowStatsByFlowQueryHookResult = ReturnType<typeof useFlowStatsByFlowQuery>;
export type FlowStatsByFlowLazyQueryHookResult = ReturnType<typeof useFlowStatsByFlowLazyQuery>;
export type FlowStatsByFlowSuspenseQueryHookResult = ReturnType<typeof useFlowStatsByFlowSuspenseQuery>;
export type FlowStatsByFlowQueryResult = Apollo.QueryResult<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables>;
export const FlowsExecutionStatsByPeriodDocument = gql`
    query flowsExecutionStatsByPeriod($period: UsageStatsPeriod!) {
        flowsExecutionStatsByPeriod(period: $period) {
            ...flowExecutionStatsFragment
        }
    }
    ${FlowExecutionStatsFragmentFragmentDoc}
    ${TaskExecutionStatsFragmentFragmentDoc}
    ${SubtaskExecutionStatsFragmentFragmentDoc}
`;

/**
 * __useFlowsExecutionStatsByPeriodQuery__
 *
 * To run a query within a React component, call `useFlowsExecutionStatsByPeriodQuery` and pass it any options that fit your needs.
 * When your component renders, `useFlowsExecutionStatsByPeriodQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowsExecutionStatsByPeriodQuery({
 *   variables: {
 *      period: // value for 'period'
 *   },
 * });
 */
export function useFlowsExecutionStatsByPeriodQuery(
    baseOptions: Apollo.QueryHookOptions<FlowsExecutionStatsByPeriodQuery, FlowsExecutionStatsByPeriodQueryVariables> &
        ({ variables: FlowsExecutionStatsByPeriodQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<FlowsExecutionStatsByPeriodQuery, FlowsExecutionStatsByPeriodQueryVariables>(
        FlowsExecutionStatsByPeriodDocument,
        options,
    );
}
export function useFlowsExecutionStatsByPeriodLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<
        FlowsExecutionStatsByPeriodQuery,
        FlowsExecutionStatsByPeriodQueryVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<FlowsExecutionStatsByPeriodQuery, FlowsExecutionStatsByPeriodQueryVariables>(
        FlowsExecutionStatsByPeriodDocument,
        options,
    );
}
export function useFlowsExecutionStatsByPeriodSuspenseQuery(
    baseOptions?:
        | Apollo.SkipToken
        | Apollo.SuspenseQueryHookOptions<FlowsExecutionStatsByPeriodQuery, FlowsExecutionStatsByPeriodQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<FlowsExecutionStatsByPeriodQuery, FlowsExecutionStatsByPeriodQueryVariables>(
        FlowsExecutionStatsByPeriodDocument,
        options,
    );
}
export type FlowsExecutionStatsByPeriodQueryHookResult = ReturnType<typeof useFlowsExecutionStatsByPeriodQuery>;
export type FlowsExecutionStatsByPeriodLazyQueryHookResult = ReturnType<typeof useFlowsExecutionStatsByPeriodLazyQuery>;
export type FlowsExecutionStatsByPeriodSuspenseQueryHookResult = ReturnType<
    typeof useFlowsExecutionStatsByPeriodSuspenseQuery
>;
export type FlowsExecutionStatsByPeriodQueryResult = Apollo.QueryResult<
    FlowsExecutionStatsByPeriodQuery,
    FlowsExecutionStatsByPeriodQueryVariables
>;
export const ApiTokensDocument = gql`
    query apiTokens {
        apiTokens {
            ...apiTokenFragment
        }
    }
    ${ApiTokenFragmentFragmentDoc}
`;

/**
 * __useApiTokensQuery__
 *
 * To run a query within a React component, call `useApiTokensQuery` and pass it any options that fit your needs.
 * When your component renders, `useApiTokensQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useApiTokensQuery({
 *   variables: {
 *   },
 * });
 */
export function useApiTokensQuery(baseOptions?: Apollo.QueryHookOptions<ApiTokensQuery, ApiTokensQueryVariables>) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<ApiTokensQuery, ApiTokensQueryVariables>(ApiTokensDocument, options);
}
export function useApiTokensLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<ApiTokensQuery, ApiTokensQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<ApiTokensQuery, ApiTokensQueryVariables>(ApiTokensDocument, options);
}
export function useApiTokensSuspenseQuery(
    baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ApiTokensQuery, ApiTokensQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<ApiTokensQuery, ApiTokensQueryVariables>(ApiTokensDocument, options);
}
export type ApiTokensQueryHookResult = ReturnType<typeof useApiTokensQuery>;
export type ApiTokensLazyQueryHookResult = ReturnType<typeof useApiTokensLazyQuery>;
export type ApiTokensSuspenseQueryHookResult = ReturnType<typeof useApiTokensSuspenseQuery>;
export type ApiTokensQueryResult = Apollo.QueryResult<ApiTokensQuery, ApiTokensQueryVariables>;
export const ApiTokenDocument = gql`
    query apiToken($tokenId: String!) {
        apiToken(tokenId: $tokenId) {
            ...apiTokenFragment
        }
    }
    ${ApiTokenFragmentFragmentDoc}
`;

/**
 * __useApiTokenQuery__
 *
 * To run a query within a React component, call `useApiTokenQuery` and pass it any options that fit your needs.
 * When your component renders, `useApiTokenQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useApiTokenQuery({
 *   variables: {
 *      tokenId: // value for 'tokenId'
 *   },
 * });
 */
export function useApiTokenQuery(
    baseOptions: Apollo.QueryHookOptions<ApiTokenQuery, ApiTokenQueryVariables> &
        ({ variables: ApiTokenQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<ApiTokenQuery, ApiTokenQueryVariables>(ApiTokenDocument, options);
}
export function useApiTokenLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ApiTokenQuery, ApiTokenQueryVariables>) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<ApiTokenQuery, ApiTokenQueryVariables>(ApiTokenDocument, options);
}
export function useApiTokenSuspenseQuery(
    baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ApiTokenQuery, ApiTokenQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<ApiTokenQuery, ApiTokenQueryVariables>(ApiTokenDocument, options);
}
export type ApiTokenQueryHookResult = ReturnType<typeof useApiTokenQuery>;
export type ApiTokenLazyQueryHookResult = ReturnType<typeof useApiTokenLazyQuery>;
export type ApiTokenSuspenseQueryHookResult = ReturnType<typeof useApiTokenSuspenseQuery>;
export type ApiTokenQueryResult = Apollo.QueryResult<ApiTokenQuery, ApiTokenQueryVariables>;
export const SettingsUserDocument = gql`
    query settingsUser {
        settingsUser {
            ...userPreferencesFragment
        }
    }
    ${UserPreferencesFragmentFragmentDoc}
`;

/**
 * __useSettingsUserQuery__
 *
 * To run a query within a React component, call `useSettingsUserQuery` and pass it any options that fit your needs.
 * When your component renders, `useSettingsUserQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSettingsUserQuery({
 *   variables: {
 *   },
 * });
 */
export function useSettingsUserQuery(
    baseOptions?: Apollo.QueryHookOptions<SettingsUserQuery, SettingsUserQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useQuery<SettingsUserQuery, SettingsUserQueryVariables>(SettingsUserDocument, options);
}
export function useSettingsUserLazyQuery(
    baseOptions?: Apollo.LazyQueryHookOptions<SettingsUserQuery, SettingsUserQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useLazyQuery<SettingsUserQuery, SettingsUserQueryVariables>(SettingsUserDocument, options);
}
export function useSettingsUserSuspenseQuery(
    baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<SettingsUserQuery, SettingsUserQueryVariables>,
) {
    const options = baseOptions === Apollo.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return Apollo.useSuspenseQuery<SettingsUserQuery, SettingsUserQueryVariables>(SettingsUserDocument, options);
}
export type SettingsUserQueryHookResult = ReturnType<typeof useSettingsUserQuery>;
export type SettingsUserLazyQueryHookResult = ReturnType<typeof useSettingsUserLazyQuery>;
export type SettingsUserSuspenseQueryHookResult = ReturnType<typeof useSettingsUserSuspenseQuery>;
export type SettingsUserQueryResult = Apollo.QueryResult<SettingsUserQuery, SettingsUserQueryVariables>;
export const AddFavoriteFlowDocument = gql`
    mutation addFavoriteFlow($flowId: ID!) {
        addFavoriteFlow(flowId: $flowId)
    }
`;
export type AddFavoriteFlowMutationFn = Apollo.MutationFunction<
    AddFavoriteFlowMutation,
    AddFavoriteFlowMutationVariables
>;

/**
 * __useAddFavoriteFlowMutation__
 *
 * To run a mutation, you first call `useAddFavoriteFlowMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAddFavoriteFlowMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [addFavoriteFlowMutation, { data, loading, error }] = useAddFavoriteFlowMutation({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useAddFavoriteFlowMutation(
    baseOptions?: Apollo.MutationHookOptions<AddFavoriteFlowMutation, AddFavoriteFlowMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<AddFavoriteFlowMutation, AddFavoriteFlowMutationVariables>(
        AddFavoriteFlowDocument,
        options,
    );
}
export type AddFavoriteFlowMutationHookResult = ReturnType<typeof useAddFavoriteFlowMutation>;
export type AddFavoriteFlowMutationResult = Apollo.MutationResult<AddFavoriteFlowMutation>;
export type AddFavoriteFlowMutationOptions = Apollo.BaseMutationOptions<
    AddFavoriteFlowMutation,
    AddFavoriteFlowMutationVariables
>;
export const DeleteFavoriteFlowDocument = gql`
    mutation deleteFavoriteFlow($flowId: ID!) {
        deleteFavoriteFlow(flowId: $flowId)
    }
`;
export type DeleteFavoriteFlowMutationFn = Apollo.MutationFunction<
    DeleteFavoriteFlowMutation,
    DeleteFavoriteFlowMutationVariables
>;

/**
 * __useDeleteFavoriteFlowMutation__
 *
 * To run a mutation, you first call `useDeleteFavoriteFlowMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteFavoriteFlowMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteFavoriteFlowMutation, { data, loading, error }] = useDeleteFavoriteFlowMutation({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useDeleteFavoriteFlowMutation(
    baseOptions?: Apollo.MutationHookOptions<DeleteFavoriteFlowMutation, DeleteFavoriteFlowMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<DeleteFavoriteFlowMutation, DeleteFavoriteFlowMutationVariables>(
        DeleteFavoriteFlowDocument,
        options,
    );
}
export type DeleteFavoriteFlowMutationHookResult = ReturnType<typeof useDeleteFavoriteFlowMutation>;
export type DeleteFavoriteFlowMutationResult = Apollo.MutationResult<DeleteFavoriteFlowMutation>;
export type DeleteFavoriteFlowMutationOptions = Apollo.BaseMutationOptions<
    DeleteFavoriteFlowMutation,
    DeleteFavoriteFlowMutationVariables
>;
export const CreateFlowDocument = gql`
    mutation createFlow($modelProvider: String!, $input: String!) {
        createFlow(modelProvider: $modelProvider, input: $input) {
            ...flowFragment
        }
    }
    ${FlowFragmentFragmentDoc}
    ${TerminalFragmentFragmentDoc}
    ${ProviderFragmentFragmentDoc}
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
export function useFinishFlowMutation(
    baseOptions?: Apollo.MutationHookOptions<FinishFlowMutation, FinishFlowMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
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
export function useStopFlowMutation(
    baseOptions?: Apollo.MutationHookOptions<StopFlowMutation, StopFlowMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<StopFlowMutation, StopFlowMutationVariables>(StopFlowDocument, options);
}
export type StopFlowMutationHookResult = ReturnType<typeof useStopFlowMutation>;
export type StopFlowMutationResult = Apollo.MutationResult<StopFlowMutation>;
export type StopFlowMutationOptions = Apollo.BaseMutationOptions<StopFlowMutation, StopFlowMutationVariables>;
export const RenameFlowDocument = gql`
    mutation renameFlow($flowId: ID!, $title: String!) {
        renameFlow(flowId: $flowId, title: $title)
    }
`;
export type RenameFlowMutationFn = Apollo.MutationFunction<RenameFlowMutation, RenameFlowMutationVariables>;

/**
 * __useRenameFlowMutation__
 *
 * To run a mutation, you first call `useRenameFlowMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRenameFlowMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [renameFlowMutation, { data, loading, error }] = useRenameFlowMutation({
 *   variables: {
 *      flowId: // value for 'flowId'
 *      title: // value for 'title'
 *   },
 * });
 */
export function useRenameFlowMutation(
    baseOptions?: Apollo.MutationHookOptions<RenameFlowMutation, RenameFlowMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<RenameFlowMutation, RenameFlowMutationVariables>(RenameFlowDocument, options);
}
export type RenameFlowMutationHookResult = ReturnType<typeof useRenameFlowMutation>;
export type RenameFlowMutationResult = Apollo.MutationResult<RenameFlowMutation>;
export type RenameFlowMutationOptions = Apollo.BaseMutationOptions<RenameFlowMutation, RenameFlowMutationVariables>;
export const CreateAssistantDocument = gql`
    mutation createAssistant($flowId: ID!, $modelProvider: String!, $input: String!, $useAgents: Boolean!) {
        createAssistant(flowId: $flowId, modelProvider: $modelProvider, input: $input, useAgents: $useAgents) {
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
    ${ProviderFragmentFragmentDoc}
    ${AssistantFragmentFragmentDoc}
`;
export type CreateAssistantMutationFn = Apollo.MutationFunction<
    CreateAssistantMutation,
    CreateAssistantMutationVariables
>;

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
export function useCreateAssistantMutation(
    baseOptions?: Apollo.MutationHookOptions<CreateAssistantMutation, CreateAssistantMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<CreateAssistantMutation, CreateAssistantMutationVariables>(
        CreateAssistantDocument,
        options,
    );
}
export type CreateAssistantMutationHookResult = ReturnType<typeof useCreateAssistantMutation>;
export type CreateAssistantMutationResult = Apollo.MutationResult<CreateAssistantMutation>;
export type CreateAssistantMutationOptions = Apollo.BaseMutationOptions<
    CreateAssistantMutation,
    CreateAssistantMutationVariables
>;
export const CallAssistantDocument = gql`
    mutation callAssistant($flowId: ID!, $assistantId: ID!, $input: String!, $useAgents: Boolean!) {
        callAssistant(flowId: $flowId, assistantId: $assistantId, input: $input, useAgents: $useAgents)
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
export function useCallAssistantMutation(
    baseOptions?: Apollo.MutationHookOptions<CallAssistantMutation, CallAssistantMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<CallAssistantMutation, CallAssistantMutationVariables>(CallAssistantDocument, options);
}
export type CallAssistantMutationHookResult = ReturnType<typeof useCallAssistantMutation>;
export type CallAssistantMutationResult = Apollo.MutationResult<CallAssistantMutation>;
export type CallAssistantMutationOptions = Apollo.BaseMutationOptions<
    CallAssistantMutation,
    CallAssistantMutationVariables
>;
export const StopAssistantDocument = gql`
    mutation stopAssistant($flowId: ID!, $assistantId: ID!) {
        stopAssistant(flowId: $flowId, assistantId: $assistantId) {
            ...assistantFragment
        }
    }
    ${AssistantFragmentFragmentDoc}
    ${ProviderFragmentFragmentDoc}
`;
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
export function useStopAssistantMutation(
    baseOptions?: Apollo.MutationHookOptions<StopAssistantMutation, StopAssistantMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<StopAssistantMutation, StopAssistantMutationVariables>(StopAssistantDocument, options);
}
export type StopAssistantMutationHookResult = ReturnType<typeof useStopAssistantMutation>;
export type StopAssistantMutationResult = Apollo.MutationResult<StopAssistantMutation>;
export type StopAssistantMutationOptions = Apollo.BaseMutationOptions<
    StopAssistantMutation,
    StopAssistantMutationVariables
>;
export const DeleteAssistantDocument = gql`
    mutation deleteAssistant($flowId: ID!, $assistantId: ID!) {
        deleteAssistant(flowId: $flowId, assistantId: $assistantId)
    }
`;
export type DeleteAssistantMutationFn = Apollo.MutationFunction<
    DeleteAssistantMutation,
    DeleteAssistantMutationVariables
>;

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
export function useDeleteAssistantMutation(
    baseOptions?: Apollo.MutationHookOptions<DeleteAssistantMutation, DeleteAssistantMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<DeleteAssistantMutation, DeleteAssistantMutationVariables>(
        DeleteAssistantDocument,
        options,
    );
}
export type DeleteAssistantMutationHookResult = ReturnType<typeof useDeleteAssistantMutation>;
export type DeleteAssistantMutationResult = Apollo.MutationResult<DeleteAssistantMutation>;
export type DeleteAssistantMutationOptions = Apollo.BaseMutationOptions<
    DeleteAssistantMutation,
    DeleteAssistantMutationVariables
>;
export const TestAgentDocument = gql`
    mutation testAgent($type: ProviderType!, $agentType: AgentConfigType!, $agent: AgentConfigInput!) {
        testAgent(type: $type, agentType: $agentType, agent: $agent) {
            ...agentTestResultFragment
        }
    }
    ${AgentTestResultFragmentFragmentDoc}
    ${TestResultFragmentFragmentDoc}
`;
export type TestAgentMutationFn = Apollo.MutationFunction<TestAgentMutation, TestAgentMutationVariables>;

/**
 * __useTestAgentMutation__
 *
 * To run a mutation, you first call `useTestAgentMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useTestAgentMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [testAgentMutation, { data, loading, error }] = useTestAgentMutation({
 *   variables: {
 *      type: // value for 'type'
 *      agentType: // value for 'agentType'
 *      agent: // value for 'agent'
 *   },
 * });
 */
export function useTestAgentMutation(
    baseOptions?: Apollo.MutationHookOptions<TestAgentMutation, TestAgentMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<TestAgentMutation, TestAgentMutationVariables>(TestAgentDocument, options);
}
export type TestAgentMutationHookResult = ReturnType<typeof useTestAgentMutation>;
export type TestAgentMutationResult = Apollo.MutationResult<TestAgentMutation>;
export type TestAgentMutationOptions = Apollo.BaseMutationOptions<TestAgentMutation, TestAgentMutationVariables>;
export const TestProviderDocument = gql`
    mutation testProvider($type: ProviderType!, $agents: AgentsConfigInput!) {
        testProvider(type: $type, agents: $agents) {
            ...providerTestResultFragment
        }
    }
    ${ProviderTestResultFragmentFragmentDoc}
    ${AgentTestResultFragmentFragmentDoc}
    ${TestResultFragmentFragmentDoc}
`;
export type TestProviderMutationFn = Apollo.MutationFunction<TestProviderMutation, TestProviderMutationVariables>;

/**
 * __useTestProviderMutation__
 *
 * To run a mutation, you first call `useTestProviderMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useTestProviderMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [testProviderMutation, { data, loading, error }] = useTestProviderMutation({
 *   variables: {
 *      type: // value for 'type'
 *      agents: // value for 'agents'
 *   },
 * });
 */
export function useTestProviderMutation(
    baseOptions?: Apollo.MutationHookOptions<TestProviderMutation, TestProviderMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<TestProviderMutation, TestProviderMutationVariables>(TestProviderDocument, options);
}
export type TestProviderMutationHookResult = ReturnType<typeof useTestProviderMutation>;
export type TestProviderMutationResult = Apollo.MutationResult<TestProviderMutation>;
export type TestProviderMutationOptions = Apollo.BaseMutationOptions<
    TestProviderMutation,
    TestProviderMutationVariables
>;
export const CreateProviderDocument = gql`
    mutation createProvider($name: String!, $type: ProviderType!, $agents: AgentsConfigInput!) {
        createProvider(name: $name, type: $type, agents: $agents) {
            ...providerConfigFragment
        }
    }
    ${ProviderConfigFragmentFragmentDoc}
    ${AgentsConfigFragmentFragmentDoc}
    ${AgentConfigFragmentFragmentDoc}
`;
export type CreateProviderMutationFn = Apollo.MutationFunction<CreateProviderMutation, CreateProviderMutationVariables>;

/**
 * __useCreateProviderMutation__
 *
 * To run a mutation, you first call `useCreateProviderMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateProviderMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createProviderMutation, { data, loading, error }] = useCreateProviderMutation({
 *   variables: {
 *      name: // value for 'name'
 *      type: // value for 'type'
 *      agents: // value for 'agents'
 *   },
 * });
 */
export function useCreateProviderMutation(
    baseOptions?: Apollo.MutationHookOptions<CreateProviderMutation, CreateProviderMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<CreateProviderMutation, CreateProviderMutationVariables>(CreateProviderDocument, options);
}
export type CreateProviderMutationHookResult = ReturnType<typeof useCreateProviderMutation>;
export type CreateProviderMutationResult = Apollo.MutationResult<CreateProviderMutation>;
export type CreateProviderMutationOptions = Apollo.BaseMutationOptions<
    CreateProviderMutation,
    CreateProviderMutationVariables
>;
export const UpdateProviderDocument = gql`
    mutation updateProvider($providerId: ID!, $name: String!, $agents: AgentsConfigInput!) {
        updateProvider(providerId: $providerId, name: $name, agents: $agents) {
            ...providerConfigFragment
        }
    }
    ${ProviderConfigFragmentFragmentDoc}
    ${AgentsConfigFragmentFragmentDoc}
    ${AgentConfigFragmentFragmentDoc}
`;
export type UpdateProviderMutationFn = Apollo.MutationFunction<UpdateProviderMutation, UpdateProviderMutationVariables>;

/**
 * __useUpdateProviderMutation__
 *
 * To run a mutation, you first call `useUpdateProviderMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateProviderMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateProviderMutation, { data, loading, error }] = useUpdateProviderMutation({
 *   variables: {
 *      providerId: // value for 'providerId'
 *      name: // value for 'name'
 *      agents: // value for 'agents'
 *   },
 * });
 */
export function useUpdateProviderMutation(
    baseOptions?: Apollo.MutationHookOptions<UpdateProviderMutation, UpdateProviderMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<UpdateProviderMutation, UpdateProviderMutationVariables>(UpdateProviderDocument, options);
}
export type UpdateProviderMutationHookResult = ReturnType<typeof useUpdateProviderMutation>;
export type UpdateProviderMutationResult = Apollo.MutationResult<UpdateProviderMutation>;
export type UpdateProviderMutationOptions = Apollo.BaseMutationOptions<
    UpdateProviderMutation,
    UpdateProviderMutationVariables
>;
export const DeleteProviderDocument = gql`
    mutation deleteProvider($providerId: ID!) {
        deleteProvider(providerId: $providerId)
    }
`;
export type DeleteProviderMutationFn = Apollo.MutationFunction<DeleteProviderMutation, DeleteProviderMutationVariables>;

/**
 * __useDeleteProviderMutation__
 *
 * To run a mutation, you first call `useDeleteProviderMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteProviderMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteProviderMutation, { data, loading, error }] = useDeleteProviderMutation({
 *   variables: {
 *      providerId: // value for 'providerId'
 *   },
 * });
 */
export function useDeleteProviderMutation(
    baseOptions?: Apollo.MutationHookOptions<DeleteProviderMutation, DeleteProviderMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<DeleteProviderMutation, DeleteProviderMutationVariables>(DeleteProviderDocument, options);
}
export type DeleteProviderMutationHookResult = ReturnType<typeof useDeleteProviderMutation>;
export type DeleteProviderMutationResult = Apollo.MutationResult<DeleteProviderMutation>;
export type DeleteProviderMutationOptions = Apollo.BaseMutationOptions<
    DeleteProviderMutation,
    DeleteProviderMutationVariables
>;
export const ValidatePromptDocument = gql`
    mutation validatePrompt($type: PromptType!, $template: String!) {
        validatePrompt(type: $type, template: $template) {
            ...promptValidationResultFragment
        }
    }
    ${PromptValidationResultFragmentFragmentDoc}
`;
export type ValidatePromptMutationFn = Apollo.MutationFunction<ValidatePromptMutation, ValidatePromptMutationVariables>;

/**
 * __useValidatePromptMutation__
 *
 * To run a mutation, you first call `useValidatePromptMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useValidatePromptMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [validatePromptMutation, { data, loading, error }] = useValidatePromptMutation({
 *   variables: {
 *      type: // value for 'type'
 *      template: // value for 'template'
 *   },
 * });
 */
export function useValidatePromptMutation(
    baseOptions?: Apollo.MutationHookOptions<ValidatePromptMutation, ValidatePromptMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<ValidatePromptMutation, ValidatePromptMutationVariables>(ValidatePromptDocument, options);
}
export type ValidatePromptMutationHookResult = ReturnType<typeof useValidatePromptMutation>;
export type ValidatePromptMutationResult = Apollo.MutationResult<ValidatePromptMutation>;
export type ValidatePromptMutationOptions = Apollo.BaseMutationOptions<
    ValidatePromptMutation,
    ValidatePromptMutationVariables
>;
export const CreatePromptDocument = gql`
    mutation createPrompt($type: PromptType!, $template: String!) {
        createPrompt(type: $type, template: $template) {
            ...userPromptFragment
        }
    }
    ${UserPromptFragmentFragmentDoc}
`;
export type CreatePromptMutationFn = Apollo.MutationFunction<CreatePromptMutation, CreatePromptMutationVariables>;

/**
 * __useCreatePromptMutation__
 *
 * To run a mutation, you first call `useCreatePromptMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreatePromptMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createPromptMutation, { data, loading, error }] = useCreatePromptMutation({
 *   variables: {
 *      type: // value for 'type'
 *      template: // value for 'template'
 *   },
 * });
 */
export function useCreatePromptMutation(
    baseOptions?: Apollo.MutationHookOptions<CreatePromptMutation, CreatePromptMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<CreatePromptMutation, CreatePromptMutationVariables>(CreatePromptDocument, options);
}
export type CreatePromptMutationHookResult = ReturnType<typeof useCreatePromptMutation>;
export type CreatePromptMutationResult = Apollo.MutationResult<CreatePromptMutation>;
export type CreatePromptMutationOptions = Apollo.BaseMutationOptions<
    CreatePromptMutation,
    CreatePromptMutationVariables
>;
export const UpdatePromptDocument = gql`
    mutation updatePrompt($promptId: ID!, $template: String!) {
        updatePrompt(promptId: $promptId, template: $template) {
            ...userPromptFragment
        }
    }
    ${UserPromptFragmentFragmentDoc}
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
 *      promptId: // value for 'promptId'
 *      template: // value for 'template'
 *   },
 * });
 */
export function useUpdatePromptMutation(
    baseOptions?: Apollo.MutationHookOptions<UpdatePromptMutation, UpdatePromptMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<UpdatePromptMutation, UpdatePromptMutationVariables>(UpdatePromptDocument, options);
}
export type UpdatePromptMutationHookResult = ReturnType<typeof useUpdatePromptMutation>;
export type UpdatePromptMutationResult = Apollo.MutationResult<UpdatePromptMutation>;
export type UpdatePromptMutationOptions = Apollo.BaseMutationOptions<
    UpdatePromptMutation,
    UpdatePromptMutationVariables
>;
export const DeletePromptDocument = gql`
    mutation deletePrompt($promptId: ID!) {
        deletePrompt(promptId: $promptId)
    }
`;
export type DeletePromptMutationFn = Apollo.MutationFunction<DeletePromptMutation, DeletePromptMutationVariables>;

/**
 * __useDeletePromptMutation__
 *
 * To run a mutation, you first call `useDeletePromptMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeletePromptMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deletePromptMutation, { data, loading, error }] = useDeletePromptMutation({
 *   variables: {
 *      promptId: // value for 'promptId'
 *   },
 * });
 */
export function useDeletePromptMutation(
    baseOptions?: Apollo.MutationHookOptions<DeletePromptMutation, DeletePromptMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<DeletePromptMutation, DeletePromptMutationVariables>(DeletePromptDocument, options);
}
export type DeletePromptMutationHookResult = ReturnType<typeof useDeletePromptMutation>;
export type DeletePromptMutationResult = Apollo.MutationResult<DeletePromptMutation>;
export type DeletePromptMutationOptions = Apollo.BaseMutationOptions<
    DeletePromptMutation,
    DeletePromptMutationVariables
>;
export const CreateApiTokenDocument = gql`
    mutation createAPIToken($input: CreateAPITokenInput!) {
        createAPIToken(input: $input) {
            ...apiTokenWithSecretFragment
        }
    }
    ${ApiTokenWithSecretFragmentFragmentDoc}
`;
export type CreateApiTokenMutationFn = Apollo.MutationFunction<CreateApiTokenMutation, CreateApiTokenMutationVariables>;

/**
 * __useCreateApiTokenMutation__
 *
 * To run a mutation, you first call `useCreateApiTokenMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateApiTokenMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createApiTokenMutation, { data, loading, error }] = useCreateApiTokenMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateApiTokenMutation(
    baseOptions?: Apollo.MutationHookOptions<CreateApiTokenMutation, CreateApiTokenMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<CreateApiTokenMutation, CreateApiTokenMutationVariables>(CreateApiTokenDocument, options);
}
export type CreateApiTokenMutationHookResult = ReturnType<typeof useCreateApiTokenMutation>;
export type CreateApiTokenMutationResult = Apollo.MutationResult<CreateApiTokenMutation>;
export type CreateApiTokenMutationOptions = Apollo.BaseMutationOptions<
    CreateApiTokenMutation,
    CreateApiTokenMutationVariables
>;
export const UpdateApiTokenDocument = gql`
    mutation updateAPIToken($tokenId: String!, $input: UpdateAPITokenInput!) {
        updateAPIToken(tokenId: $tokenId, input: $input) {
            ...apiTokenFragment
        }
    }
    ${ApiTokenFragmentFragmentDoc}
`;
export type UpdateApiTokenMutationFn = Apollo.MutationFunction<UpdateApiTokenMutation, UpdateApiTokenMutationVariables>;

/**
 * __useUpdateApiTokenMutation__
 *
 * To run a mutation, you first call `useUpdateApiTokenMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateApiTokenMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateApiTokenMutation, { data, loading, error }] = useUpdateApiTokenMutation({
 *   variables: {
 *      tokenId: // value for 'tokenId'
 *      input: // value for 'input'
 *   },
 * });
 */
export function useUpdateApiTokenMutation(
    baseOptions?: Apollo.MutationHookOptions<UpdateApiTokenMutation, UpdateApiTokenMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<UpdateApiTokenMutation, UpdateApiTokenMutationVariables>(UpdateApiTokenDocument, options);
}
export type UpdateApiTokenMutationHookResult = ReturnType<typeof useUpdateApiTokenMutation>;
export type UpdateApiTokenMutationResult = Apollo.MutationResult<UpdateApiTokenMutation>;
export type UpdateApiTokenMutationOptions = Apollo.BaseMutationOptions<
    UpdateApiTokenMutation,
    UpdateApiTokenMutationVariables
>;
export const DeleteApiTokenDocument = gql`
    mutation deleteAPIToken($tokenId: String!) {
        deleteAPIToken(tokenId: $tokenId)
    }
`;
export type DeleteApiTokenMutationFn = Apollo.MutationFunction<DeleteApiTokenMutation, DeleteApiTokenMutationVariables>;

/**
 * __useDeleteApiTokenMutation__
 *
 * To run a mutation, you first call `useDeleteApiTokenMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteApiTokenMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteApiTokenMutation, { data, loading, error }] = useDeleteApiTokenMutation({
 *   variables: {
 *      tokenId: // value for 'tokenId'
 *   },
 * });
 */
export function useDeleteApiTokenMutation(
    baseOptions?: Apollo.MutationHookOptions<DeleteApiTokenMutation, DeleteApiTokenMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useMutation<DeleteApiTokenMutation, DeleteApiTokenMutationVariables>(DeleteApiTokenDocument, options);
}
export type DeleteApiTokenMutationHookResult = ReturnType<typeof useDeleteApiTokenMutation>;
export type DeleteApiTokenMutationResult = Apollo.MutationResult<DeleteApiTokenMutation>;
export type DeleteApiTokenMutationOptions = Apollo.BaseMutationOptions<
    DeleteApiTokenMutation,
    DeleteApiTokenMutationVariables
>;
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
export const AssistantCreatedDocument = gql`
    subscription assistantCreated($flowId: ID!) {
        assistantCreated(flowId: $flowId) {
            ...assistantFragment
        }
    }
    ${AssistantFragmentFragmentDoc}
    ${ProviderFragmentFragmentDoc}
`;

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
export function useAssistantCreatedSubscription(
    baseOptions: Apollo.SubscriptionHookOptions<AssistantCreatedSubscription, AssistantCreatedSubscriptionVariables> &
        ({ variables: AssistantCreatedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<AssistantCreatedSubscription, AssistantCreatedSubscriptionVariables>(
        AssistantCreatedDocument,
        options,
    );
}
export type AssistantCreatedSubscriptionHookResult = ReturnType<typeof useAssistantCreatedSubscription>;
export type AssistantCreatedSubscriptionResult = Apollo.SubscriptionResult<AssistantCreatedSubscription>;
export const AssistantUpdatedDocument = gql`
    subscription assistantUpdated($flowId: ID!) {
        assistantUpdated(flowId: $flowId) {
            ...assistantFragment
        }
    }
    ${AssistantFragmentFragmentDoc}
    ${ProviderFragmentFragmentDoc}
`;

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
export function useAssistantUpdatedSubscription(
    baseOptions: Apollo.SubscriptionHookOptions<AssistantUpdatedSubscription, AssistantUpdatedSubscriptionVariables> &
        ({ variables: AssistantUpdatedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<AssistantUpdatedSubscription, AssistantUpdatedSubscriptionVariables>(
        AssistantUpdatedDocument,
        options,
    );
}
export type AssistantUpdatedSubscriptionHookResult = ReturnType<typeof useAssistantUpdatedSubscription>;
export type AssistantUpdatedSubscriptionResult = Apollo.SubscriptionResult<AssistantUpdatedSubscription>;
export const AssistantDeletedDocument = gql`
    subscription assistantDeleted($flowId: ID!) {
        assistantDeleted(flowId: $flowId) {
            ...assistantFragment
        }
    }
    ${AssistantFragmentFragmentDoc}
    ${ProviderFragmentFragmentDoc}
`;

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
export function useAssistantDeletedSubscription(
    baseOptions: Apollo.SubscriptionHookOptions<AssistantDeletedSubscription, AssistantDeletedSubscriptionVariables> &
        ({ variables: AssistantDeletedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<AssistantDeletedSubscription, AssistantDeletedSubscriptionVariables>(
        AssistantDeletedDocument,
        options,
    );
}
export type AssistantDeletedSubscriptionHookResult = ReturnType<typeof useAssistantDeletedSubscription>;
export type AssistantDeletedSubscriptionResult = Apollo.SubscriptionResult<AssistantDeletedSubscription>;
export const AssistantLogAddedDocument = gql`
    subscription assistantLogAdded($flowId: ID!) {
        assistantLogAdded(flowId: $flowId) {
            ...assistantLogFragment
        }
    }
    ${AssistantLogFragmentFragmentDoc}
`;

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
export function useAssistantLogAddedSubscription(
    baseOptions: Apollo.SubscriptionHookOptions<AssistantLogAddedSubscription, AssistantLogAddedSubscriptionVariables> &
        ({ variables: AssistantLogAddedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<AssistantLogAddedSubscription, AssistantLogAddedSubscriptionVariables>(
        AssistantLogAddedDocument,
        options,
    );
}
export type AssistantLogAddedSubscriptionHookResult = ReturnType<typeof useAssistantLogAddedSubscription>;
export type AssistantLogAddedSubscriptionResult = Apollo.SubscriptionResult<AssistantLogAddedSubscription>;
export const AssistantLogUpdatedDocument = gql`
    subscription assistantLogUpdated($flowId: ID!) {
        assistantLogUpdated(flowId: $flowId) {
            ...assistantLogFragment
        }
    }
    ${AssistantLogFragmentFragmentDoc}
`;

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
export function useAssistantLogUpdatedSubscription(
    baseOptions: Apollo.SubscriptionHookOptions<
        AssistantLogUpdatedSubscription,
        AssistantLogUpdatedSubscriptionVariables
    > &
        ({ variables: AssistantLogUpdatedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<AssistantLogUpdatedSubscription, AssistantLogUpdatedSubscriptionVariables>(
        AssistantLogUpdatedDocument,
        options,
    );
}
export type AssistantLogUpdatedSubscriptionHookResult = ReturnType<typeof useAssistantLogUpdatedSubscription>;
export type AssistantLogUpdatedSubscriptionResult = Apollo.SubscriptionResult<AssistantLogUpdatedSubscription>;
export const FlowCreatedDocument = gql`
    subscription flowCreated {
        flowCreated {
            ...flowFragment
        }
    }
    ${FlowFragmentFragmentDoc}
    ${TerminalFragmentFragmentDoc}
    ${ProviderFragmentFragmentDoc}
`;

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
export function useFlowCreatedSubscription(
    baseOptions?: Apollo.SubscriptionHookOptions<FlowCreatedSubscription, FlowCreatedSubscriptionVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<FlowCreatedSubscription, FlowCreatedSubscriptionVariables>(
        FlowCreatedDocument,
        options,
    );
}
export type FlowCreatedSubscriptionHookResult = ReturnType<typeof useFlowCreatedSubscription>;
export type FlowCreatedSubscriptionResult = Apollo.SubscriptionResult<FlowCreatedSubscription>;
export const FlowDeletedDocument = gql`
    subscription flowDeleted {
        flowDeleted {
            ...flowFragment
        }
    }
    ${FlowFragmentFragmentDoc}
    ${TerminalFragmentFragmentDoc}
    ${ProviderFragmentFragmentDoc}
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
export function useFlowDeletedSubscription(
    baseOptions?: Apollo.SubscriptionHookOptions<FlowDeletedSubscription, FlowDeletedSubscriptionVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<FlowDeletedSubscription, FlowDeletedSubscriptionVariables>(
        FlowDeletedDocument,
        options,
    );
}
export type FlowDeletedSubscriptionHookResult = ReturnType<typeof useFlowDeletedSubscription>;
export type FlowDeletedSubscriptionResult = Apollo.SubscriptionResult<FlowDeletedSubscription>;
export const FlowUpdatedDocument = gql`
    subscription flowUpdated {
        flowUpdated {
            ...flowFragment
        }
    }
    ${FlowFragmentFragmentDoc}
    ${TerminalFragmentFragmentDoc}
    ${ProviderFragmentFragmentDoc}
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
 *   },
 * });
 */
export function useFlowUpdatedSubscription(
    baseOptions?: Apollo.SubscriptionHookOptions<FlowUpdatedSubscription, FlowUpdatedSubscriptionVariables>,
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
export const ProviderCreatedDocument = gql`
    subscription providerCreated {
        providerCreated {
            ...providerConfigFragment
        }
    }
    ${ProviderConfigFragmentFragmentDoc}
    ${AgentsConfigFragmentFragmentDoc}
    ${AgentConfigFragmentFragmentDoc}
`;

/**
 * __useProviderCreatedSubscription__
 *
 * To run a query within a React component, call `useProviderCreatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useProviderCreatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useProviderCreatedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useProviderCreatedSubscription(
    baseOptions?: Apollo.SubscriptionHookOptions<ProviderCreatedSubscription, ProviderCreatedSubscriptionVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<ProviderCreatedSubscription, ProviderCreatedSubscriptionVariables>(
        ProviderCreatedDocument,
        options,
    );
}
export type ProviderCreatedSubscriptionHookResult = ReturnType<typeof useProviderCreatedSubscription>;
export type ProviderCreatedSubscriptionResult = Apollo.SubscriptionResult<ProviderCreatedSubscription>;
export const ProviderUpdatedDocument = gql`
    subscription providerUpdated {
        providerUpdated {
            ...providerConfigFragment
        }
    }
    ${ProviderConfigFragmentFragmentDoc}
    ${AgentsConfigFragmentFragmentDoc}
    ${AgentConfigFragmentFragmentDoc}
`;

/**
 * __useProviderUpdatedSubscription__
 *
 * To run a query within a React component, call `useProviderUpdatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useProviderUpdatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useProviderUpdatedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useProviderUpdatedSubscription(
    baseOptions?: Apollo.SubscriptionHookOptions<ProviderUpdatedSubscription, ProviderUpdatedSubscriptionVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<ProviderUpdatedSubscription, ProviderUpdatedSubscriptionVariables>(
        ProviderUpdatedDocument,
        options,
    );
}
export type ProviderUpdatedSubscriptionHookResult = ReturnType<typeof useProviderUpdatedSubscription>;
export type ProviderUpdatedSubscriptionResult = Apollo.SubscriptionResult<ProviderUpdatedSubscription>;
export const ProviderDeletedDocument = gql`
    subscription providerDeleted {
        providerDeleted {
            ...providerConfigFragment
        }
    }
    ${ProviderConfigFragmentFragmentDoc}
    ${AgentsConfigFragmentFragmentDoc}
    ${AgentConfigFragmentFragmentDoc}
`;

/**
 * __useProviderDeletedSubscription__
 *
 * To run a query within a React component, call `useProviderDeletedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useProviderDeletedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useProviderDeletedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useProviderDeletedSubscription(
    baseOptions?: Apollo.SubscriptionHookOptions<ProviderDeletedSubscription, ProviderDeletedSubscriptionVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<ProviderDeletedSubscription, ProviderDeletedSubscriptionVariables>(
        ProviderDeletedDocument,
        options,
    );
}
export type ProviderDeletedSubscriptionHookResult = ReturnType<typeof useProviderDeletedSubscription>;
export type ProviderDeletedSubscriptionResult = Apollo.SubscriptionResult<ProviderDeletedSubscription>;
export const ApiTokenCreatedDocument = gql`
    subscription apiTokenCreated {
        apiTokenCreated {
            ...apiTokenFragment
        }
    }
    ${ApiTokenFragmentFragmentDoc}
`;

/**
 * __useApiTokenCreatedSubscription__
 *
 * To run a query within a React component, call `useApiTokenCreatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useApiTokenCreatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useApiTokenCreatedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useApiTokenCreatedSubscription(
    baseOptions?: Apollo.SubscriptionHookOptions<ApiTokenCreatedSubscription, ApiTokenCreatedSubscriptionVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<ApiTokenCreatedSubscription, ApiTokenCreatedSubscriptionVariables>(
        ApiTokenCreatedDocument,
        options,
    );
}
export type ApiTokenCreatedSubscriptionHookResult = ReturnType<typeof useApiTokenCreatedSubscription>;
export type ApiTokenCreatedSubscriptionResult = Apollo.SubscriptionResult<ApiTokenCreatedSubscription>;
export const ApiTokenUpdatedDocument = gql`
    subscription apiTokenUpdated {
        apiTokenUpdated {
            ...apiTokenFragment
        }
    }
    ${ApiTokenFragmentFragmentDoc}
`;

/**
 * __useApiTokenUpdatedSubscription__
 *
 * To run a query within a React component, call `useApiTokenUpdatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useApiTokenUpdatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useApiTokenUpdatedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useApiTokenUpdatedSubscription(
    baseOptions?: Apollo.SubscriptionHookOptions<ApiTokenUpdatedSubscription, ApiTokenUpdatedSubscriptionVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<ApiTokenUpdatedSubscription, ApiTokenUpdatedSubscriptionVariables>(
        ApiTokenUpdatedDocument,
        options,
    );
}
export type ApiTokenUpdatedSubscriptionHookResult = ReturnType<typeof useApiTokenUpdatedSubscription>;
export type ApiTokenUpdatedSubscriptionResult = Apollo.SubscriptionResult<ApiTokenUpdatedSubscription>;
export const ApiTokenDeletedDocument = gql`
    subscription apiTokenDeleted {
        apiTokenDeleted {
            ...apiTokenFragment
        }
    }
    ${ApiTokenFragmentFragmentDoc}
`;

/**
 * __useApiTokenDeletedSubscription__
 *
 * To run a query within a React component, call `useApiTokenDeletedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useApiTokenDeletedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useApiTokenDeletedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useApiTokenDeletedSubscription(
    baseOptions?: Apollo.SubscriptionHookOptions<ApiTokenDeletedSubscription, ApiTokenDeletedSubscriptionVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<ApiTokenDeletedSubscription, ApiTokenDeletedSubscriptionVariables>(
        ApiTokenDeletedDocument,
        options,
    );
}
export type ApiTokenDeletedSubscriptionHookResult = ReturnType<typeof useApiTokenDeletedSubscription>;
export type ApiTokenDeletedSubscriptionResult = Apollo.SubscriptionResult<ApiTokenDeletedSubscription>;
export const SettingsUserUpdatedDocument = gql`
    subscription settingsUserUpdated {
        settingsUserUpdated {
            ...userPreferencesFragment
        }
    }
    ${UserPreferencesFragmentFragmentDoc}
`;

/**
 * __useSettingsUserUpdatedSubscription__
 *
 * To run a query within a React component, call `useSettingsUserUpdatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useSettingsUserUpdatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSettingsUserUpdatedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useSettingsUserUpdatedSubscription(
    baseOptions?: Apollo.SubscriptionHookOptions<
        SettingsUserUpdatedSubscription,
        SettingsUserUpdatedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return Apollo.useSubscription<SettingsUserUpdatedSubscription, SettingsUserUpdatedSubscriptionVariables>(
        SettingsUserUpdatedDocument,
        options,
    );
}
export type SettingsUserUpdatedSubscriptionHookResult = ReturnType<typeof useSettingsUserUpdatedSubscription>;
export type SettingsUserUpdatedSubscriptionResult = Apollo.SubscriptionResult<SettingsUserUpdatedSubscription>;
