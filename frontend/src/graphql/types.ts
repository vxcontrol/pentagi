/** Internal type. DO NOT USE DIRECTLY. */
type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
/** Internal type. DO NOT USE DIRECTLY. */
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
import { gql } from '@apollo/client';
import type * as ApolloReactCommon from '@apollo/client/react';
import * as ApolloReactHooks from '@apollo/client/react';
const defaultOptions = {} as const;
export type AgentConfigInput = {
    frequencyPenalty?: number | null | undefined;
    maxLength?: number | null | undefined;
    maxTokens?: number | null | undefined;
    minLength?: number | null | undefined;
    model: string;
    presencePenalty?: number | null | undefined;
    price?: ModelPriceInput | null | undefined;
    reasoning?: ReasoningConfigInput | null | undefined;
    repetitionPenalty?: number | null | undefined;
    temperature?: number | null | undefined;
    topK?: number | null | undefined;
    topP?: number | null | undefined;
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

export type CreateApiTokenInput = {
    name?: string | null | undefined;
    ttl: number;
};

export type CreateFlowTemplateInput = {
    text: string;
    title: string;
};

export type CreateKnowledgeDocumentInput = {
    answerType?: KnowledgeAnswerType | null | undefined;
    codeLang?: string | null | undefined;
    content: string;
    description?: string | null | undefined;
    docType: KnowledgeDocType;
    guideType?: KnowledgeGuideType | null | undefined;
    question: string;
};

export enum KnowledgeAnswerType {
    Code = 'code',
    Guide = 'guide',
    Other = 'other',
    Tool = 'tool',
    Vulnerability = 'vulnerability',
}

export enum KnowledgeDocType {
    Answer = 'answer',
    Code = 'code',
    Guide = 'guide',
}

export type KnowledgeFilter = {
    answerTypes?: Array<KnowledgeAnswerType> | null | undefined;
    codeLangs?: Array<string> | null | undefined;
    docTypes?: Array<KnowledgeDocType> | null | undefined;
    flowId?: string | number | null | undefined;
    guideTypes?: Array<KnowledgeGuideType> | null | undefined;
    manual?: boolean | null | undefined;
};

export enum KnowledgeGuideType {
    Configure = 'configure',
    Development = 'development',
    Install = 'install',
    Other = 'other',
    Pentest = 'pentest',
    Use = 'use',
}

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

export type ModelPriceInput = {
    cacheRead: number;
    cacheWrite: number;
    input: number;
    output: number;
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
    QuestionExecutionMonitor = 'question_execution_monitor',
    QuestionInstaller = 'question_installer',
    QuestionMemorist = 'question_memorist',
    QuestionPentester = 'question_pentester',
    QuestionReflector = 'question_reflector',
    QuestionSearcher = 'question_searcher',
    QuestionTaskPlanner = 'question_task_planner',
    Refiner = 'refiner',
    Reflector = 'reflector',
    Reporter = 'reporter',
    Searcher = 'searcher',
    ShortExecutionContext = 'short_execution_context',
    SubtasksGenerator = 'subtasks_generator',
    SubtasksRefiner = 'subtasks_refiner',
    Summarizer = 'summarizer',
    TaskAssignmentWrapper = 'task_assignment_wrapper',
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

export enum ProviderType {
    Anthropic = 'anthropic',
    Bedrock = 'bedrock',
    Custom = 'custom',
    Deepseek = 'deepseek',
    Gemini = 'gemini',
    Glm = 'glm',
    Kimi = 'kimi',
    Ollama = 'ollama',
    Openai = 'openai',
    Qwen = 'qwen',
}

export type ReasoningConfigInput = {
    effort?: ReasoningEffort | null | undefined;
    maxTokens?: number | null | undefined;
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

export enum StatusType {
    Created = 'created',
    Failed = 'failed',
    Finished = 'finished',
    Running = 'running',
    Waiting = 'waiting',
}

export enum TerminalLogType {
    Stderr = 'stderr',
    Stdin = 'stdin',
    Stdout = 'stdout',
}

export enum TerminalType {
    Primary = 'primary',
    Secondary = 'secondary',
}

export enum TokenStatus {
    Active = 'active',
    Expired = 'expired',
    Revoked = 'revoked',
}

export type UpdateApiTokenInput = {
    name?: string | null | undefined;
    status?: TokenStatus | null | undefined;
};

export type UpdateFlowTemplateInput = {
    text: string;
    title: string;
};

export type UpdateKnowledgeDocumentInput = {
    answerType?: KnowledgeAnswerType | null | undefined;
    codeLang?: string | null | undefined;
    content: string;
    description?: string | null | undefined;
    docType?: KnowledgeDocType | null | undefined;
    guideType?: KnowledgeGuideType | null | undefined;
    question?: string | null | undefined;
};

export enum UsageStatsPeriod {
    Month = 'month',
    Quarter = 'quarter',
    Week = 'week',
}

export enum VectorStoreAction {
    Retrieve = 'retrieve',
    Store = 'store',
}

export type SettingsFragmentFragment = {
    debug: boolean;
    askUser: boolean;
    version: string;
    dockerInside: boolean;
    isDevelopMode: boolean;
    assistantUseAgents: boolean;
};

export type FlowFragmentFragment = {
    id: string;
    title: string;
    status: StatusType;
    createdAt: unknown;
    updatedAt: unknown;
    terminals: Array<TerminalFragmentFragment> | null;
    provider: ProviderFragmentFragment;
};

export type TerminalFragmentFragment = {
    id: string;
    type: TerminalType;
    name: string;
    image: string;
    connected: boolean;
    createdAt: unknown;
};

export type TaskFragmentFragment = {
    id: string;
    title: string;
    status: StatusType;
    input: string;
    result: string;
    flowId: string;
    createdAt: unknown;
    updatedAt: unknown;
    subtasks: Array<SubtaskFragmentFragment> | null;
};

export type SubtaskFragmentFragment = {
    id: string;
    status: StatusType;
    title: string;
    description: string;
    result: string;
    taskId: string;
    createdAt: unknown;
    updatedAt: unknown;
};

export type TerminalLogFragmentFragment = {
    id: string;
    flowId: string;
    taskId: string | null;
    subtaskId: string | null;
    type: TerminalLogType;
    text: string;
    terminal: string;
    createdAt: unknown;
};

export type MessageLogFragmentFragment = {
    id: string;
    type: MessageLogType;
    message: string;
    thinking: string | null;
    result: string;
    resultFormat: ResultFormat;
    flowId: string;
    taskId: string | null;
    subtaskId: string | null;
    createdAt: unknown;
};

export type ScreenshotFragmentFragment = {
    id: string;
    flowId: string;
    taskId: string | null;
    subtaskId: string | null;
    name: string;
    url: string;
    createdAt: unknown;
};

export type FlowFileFragmentFragment = {
    id: string;
    name: string;
    path: string;
    size: number;
    isDir: boolean;
    modifiedAt: unknown;
};

export type UserResourceFragmentFragment = {
    id: string;
    userId: string;
    name: string;
    path: string;
    size: number;
    isDir: boolean;
    createdAt: unknown;
    updatedAt: unknown;
};

export type AgentLogFragmentFragment = {
    id: string;
    flowId: string;
    initiator: AgentType;
    executor: AgentType;
    task: string;
    result: string;
    taskId: string | null;
    subtaskId: string | null;
    createdAt: unknown;
};

export type SearchLogFragmentFragment = {
    id: string;
    flowId: string;
    initiator: AgentType;
    executor: AgentType;
    engine: string;
    query: string;
    result: string;
    taskId: string | null;
    subtaskId: string | null;
    createdAt: unknown;
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
    taskId: string | null;
    subtaskId: string | null;
    createdAt: unknown;
};

export type AssistantFragmentFragment = {
    id: string;
    title: string;
    status: StatusType;
    flowId: string;
    useAgents: boolean;
    createdAt: unknown;
    updatedAt: unknown;
    provider: ProviderFragmentFragment;
};

export type AssistantLogFragmentFragment = {
    id: string;
    type: MessageLogType;
    message: string;
    thinking: string | null;
    result: string;
    resultFormat: ResultFormat;
    appendPart: boolean;
    flowId: string;
    assistantId: string;
    createdAt: unknown;
};

export type TestResultFragmentFragment = {
    name: string;
    type: string;
    result: boolean;
    reasoning: boolean;
    streaming: boolean;
    latency: number | null;
    error: string | null;
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
    price: { input: number; output: number; cacheRead: number; cacheWrite: number } | null;
};

export type ProviderFragmentFragment = { name: string; type: ProviderType };

export type ProviderConfigFragmentFragment = {
    id: string;
    name: string;
    type: ProviderType;
    createdAt: unknown;
    updatedAt: unknown;
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
    maxTokens: number | null;
    temperature: number | null;
    topK: number | null;
    topP: number | null;
    minLength: number | null;
    maxLength: number | null;
    repetitionPenalty: number | null;
    frequencyPenalty: number | null;
    presencePenalty: number | null;
    reasoning: { effort: ReasoningEffort | null; maxTokens: number | null } | null;
    price: { input: number; output: number; cacheRead: number; cacheWrite: number } | null;
};

export type UserPromptFragmentFragment = {
    id: string;
    type: PromptType;
    template: string;
    createdAt: unknown;
    updatedAt: unknown;
};

export type DefaultPromptFragmentFragment = { type: PromptType; template: string; variables: Array<string> };

export type PromptValidationResultFragmentFragment = {
    result: ResultType;
    errorType: PromptValidationErrorType | null;
    message: string | null;
    line: number | null;
    details: string | null;
};

export type ApiTokenFragmentFragment = {
    id: string;
    tokenId: string;
    userId: string;
    roleId: string;
    name: string | null;
    ttl: number;
    status: TokenStatus;
    createdAt: unknown;
    updatedAt: unknown;
};

export type ApiTokenWithSecretFragmentFragment = {
    id: string;
    tokenId: string;
    userId: string;
    roleId: string;
    name: string | null;
    ttl: number;
    status: TokenStatus;
    createdAt: unknown;
    updatedAt: unknown;
    token: string;
};

export type FlowTemplateFragmentFragment = {
    id: string;
    userId: string;
    title: string;
    text: string;
    createdAt: unknown;
    updatedAt: unknown;
};

export type UsageStatsFragmentFragment = {
    totalUsageIn: number;
    totalUsageOut: number;
    totalUsageCacheIn: number;
    totalUsageCacheOut: number;
    totalUsageCostIn: number;
    totalUsageCostOut: number;
};

export type DailyUsageStatsFragmentFragment = { date: unknown; stats: UsageStatsFragmentFragment };

export type ProviderUsageStatsFragmentFragment = { provider: string; stats: UsageStatsFragmentFragment };

export type ModelUsageStatsFragmentFragment = { model: string; provider: string; stats: UsageStatsFragmentFragment };

export type AgentTypeUsageStatsFragmentFragment = { agentType: AgentType; stats: UsageStatsFragmentFragment };

export type ModelAgentsUsageStatsFragmentFragment = {
    model: string;
    provider: string;
    agentTypes: Array<AgentType>;
    stats: UsageStatsFragmentFragment;
};

export type ToolcallsStatsFragmentFragment = { totalCount: number; totalDurationSeconds: number };

export type DailyToolcallsStatsFragmentFragment = { date: unknown; stats: ToolcallsStatsFragmentFragment };

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

export type DailyFlowsStatsFragmentFragment = { date: unknown; stats: FlowsStatsFragmentFragment };

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

export type KnowledgeDocumentFragmentFragment = {
    id: string;
    docType: KnowledgeDocType;
    content: string;
    question: string;
    description: string | null;
    userId: string;
    flowId: string | null;
    taskId: string | null;
    subtaskId: string | null;
    guideType: KnowledgeGuideType | null;
    answerType: KnowledgeAnswerType | null;
    codeLang: string | null;
    partSize: number;
    totalSize: number;
    manual: boolean;
};

export type KnowledgeDocumentWithScoreFragmentFragment = { score: number; document: KnowledgeDocumentFragmentFragment };

export type FlowsQueryVariables = Exact<{ [key: string]: never }>;

export type FlowsQuery = { flows: Array<FlowFragmentFragment> | null };

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
            deepseek: boolean;
            glm: boolean;
            kimi: boolean;
            qwen: boolean;
        };
        default: {
            openai: ProviderConfigFragmentFragment;
            anthropic: ProviderConfigFragmentFragment;
            gemini: ProviderConfigFragmentFragment | null;
            bedrock: ProviderConfigFragmentFragment | null;
            ollama: ProviderConfigFragmentFragment | null;
            custom: ProviderConfigFragmentFragment | null;
            deepseek: ProviderConfigFragmentFragment | null;
            glm: ProviderConfigFragmentFragment | null;
            kimi: ProviderConfigFragmentFragment | null;
            qwen: ProviderConfigFragmentFragment | null;
        };
        userDefined: Array<ProviderConfigFragmentFragment> | null;
        models: {
            openai: Array<ModelConfigFragmentFragment>;
            anthropic: Array<ModelConfigFragmentFragment>;
            gemini: Array<ModelConfigFragmentFragment>;
            bedrock: Array<ModelConfigFragmentFragment> | null;
            ollama: Array<ModelConfigFragmentFragment> | null;
            custom: Array<ModelConfigFragmentFragment> | null;
            deepseek: Array<ModelConfigFragmentFragment> | null;
            glm: Array<ModelConfigFragmentFragment> | null;
            kimi: Array<ModelConfigFragmentFragment> | null;
            qwen: Array<ModelConfigFragmentFragment> | null;
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
                collectToolCallId: DefaultPromptFragmentFragment;
                detectToolCallIdPattern: DefaultPromptFragmentFragment;
                monitorAgentExecution: DefaultPromptFragmentFragment;
                planAgentTask: DefaultPromptFragmentFragment;
                wrapAgentTask: DefaultPromptFragmentFragment;
            };
        };
        userDefined: Array<UserPromptFragmentFragment> | null;
    };
};

export type FlowQueryVariables = Exact<{
    id: string | number;
}>;

export type FlowQuery = {
    flow: FlowFragmentFragment;
    tasks: Array<TaskFragmentFragment> | null;
    screenshots: Array<ScreenshotFragmentFragment> | null;
    terminalLogs: Array<TerminalLogFragmentFragment> | null;
    messageLogs: Array<MessageLogFragmentFragment> | null;
    agentLogs: Array<AgentLogFragmentFragment> | null;
    searchLogs: Array<SearchLogFragmentFragment> | null;
    vectorStoreLogs: Array<VectorStoreLogFragmentFragment> | null;
};

export type TasksQueryVariables = Exact<{
    flowId: string | number;
}>;

export type TasksQuery = { tasks: Array<TaskFragmentFragment> | null };

export type FlowFilesQueryVariables = Exact<{
    flowId: string | number;
}>;

export type FlowFilesQuery = { flowFiles: Array<FlowFileFragmentFragment> };

export type ResourcesQueryVariables = Exact<{
    path?: string | null | undefined;
    recursive?: boolean | null | undefined;
}>;

export type ResourcesQuery = { resources: Array<UserResourceFragmentFragment> };

export type AssistantsQueryVariables = Exact<{
    flowId: string | number;
}>;

export type AssistantsQuery = { assistants: Array<AssistantFragmentFragment> | null };

export type AssistantLogsQueryVariables = Exact<{
    flowId: string | number;
    assistantId: string | number;
}>;

export type AssistantLogsQuery = { assistantLogs: Array<AssistantLogFragmentFragment> | null };

export type FlowReportQueryVariables = Exact<{
    id: string | number;
}>;

export type FlowReportQuery = { flow: FlowFragmentFragment; tasks: Array<TaskFragmentFragment> | null };

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
    flowId: string | number;
}>;

export type UsageStatsByFlowQuery = { usageStatsByFlow: UsageStatsFragmentFragment };

export type UsageStatsByAgentTypeForFlowQueryVariables = Exact<{
    flowId: string | number;
}>;

export type UsageStatsByAgentTypeForFlowQuery = {
    usageStatsByAgentTypeForFlow: Array<AgentTypeUsageStatsFragmentFragment>;
};

export type UsageStatsByModelAgentsForFlowQueryVariables = Exact<{
    flowId: string | number;
}>;

export type UsageStatsByModelAgentsForFlowQuery = {
    usageStatsByModelAgentsForFlow: Array<ModelAgentsUsageStatsFragmentFragment>;
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
    flowId: string | number;
}>;

export type ToolcallsStatsByFlowQuery = { toolcallsStatsByFlow: ToolcallsStatsFragmentFragment };

export type ToolcallsStatsByFunctionForFlowQueryVariables = Exact<{
    flowId: string | number;
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
    flowId: string | number;
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
    tokenId: string;
}>;

export type ApiTokenQuery = { apiToken: ApiTokenFragmentFragment | null };

export type KnowledgeDocumentsQueryVariables = Exact<{
    filter?: KnowledgeFilter | null | undefined;
    withContent: boolean;
}>;

export type KnowledgeDocumentsQuery = { knowledgeDocuments: Array<KnowledgeDocumentFragmentFragment> };

export type KnowledgeDocumentQueryVariables = Exact<{
    id: string;
}>;

export type KnowledgeDocumentQuery = { knowledgeDocument: KnowledgeDocumentFragmentFragment };

export type SearchKnowledgeQueryVariables = Exact<{
    query: string;
    filter?: KnowledgeFilter | null | undefined;
    limit?: number | null | undefined;
}>;

export type SearchKnowledgeQuery = { searchKnowledge: Array<KnowledgeDocumentWithScoreFragmentFragment> };

export type UserPreferencesFragmentFragment = { id: string; favoriteFlows: Array<string> };

export type SettingsUserQueryVariables = Exact<{ [key: string]: never }>;

export type SettingsUserQuery = { settingsUser: UserPreferencesFragmentFragment };

export type AddFavoriteFlowMutationVariables = Exact<{
    flowId: string | number;
}>;

export type AddFavoriteFlowMutation = { addFavoriteFlow: ResultType };

export type DeleteFavoriteFlowMutationVariables = Exact<{
    flowId: string | number;
}>;

export type DeleteFavoriteFlowMutation = { deleteFavoriteFlow: ResultType };

export type AnonymizeTextMutationVariables = Exact<{
    text: string;
}>;

export type AnonymizeTextMutation = { anonymizeText: string };

export type FlowTemplatesQueryVariables = Exact<{ [key: string]: never }>;

export type FlowTemplatesQuery = { flowTemplates: Array<FlowTemplateFragmentFragment> };

export type FlowTemplateQueryVariables = Exact<{
    templateId: string | number;
}>;

export type FlowTemplateQuery = { flowTemplate: FlowTemplateFragmentFragment | null };

export type CreateFlowTemplateMutationVariables = Exact<{
    input: CreateFlowTemplateInput;
}>;

export type CreateFlowTemplateMutation = { createFlowTemplate: FlowTemplateFragmentFragment };

export type UpdateFlowTemplateMutationVariables = Exact<{
    templateId: string | number;
    input: UpdateFlowTemplateInput;
}>;

export type UpdateFlowTemplateMutation = { updateFlowTemplate: FlowTemplateFragmentFragment };

export type DeleteFlowTemplateMutationVariables = Exact<{
    templateId: string | number;
}>;

export type DeleteFlowTemplateMutation = { deleteFlowTemplate: ResultType };

export type CreateFlowMutationVariables = Exact<{
    modelProvider: string;
    input: string;
    resourceIds?: Array<string | number> | string | number | null | undefined;
}>;

export type CreateFlowMutation = { createFlow: FlowFragmentFragment };

export type DeleteFlowMutationVariables = Exact<{
    flowId: string | number;
}>;

export type DeleteFlowMutation = { deleteFlow: ResultType };

export type PutUserInputMutationVariables = Exact<{
    flowId: string | number;
    input: string;
    modelProvider?: string | null | undefined;
    resourceIds?: Array<string | number> | string | number | null | undefined;
}>;

export type PutUserInputMutation = { putUserInput: ResultType };

export type FinishFlowMutationVariables = Exact<{
    flowId: string | number;
}>;

export type FinishFlowMutation = { finishFlow: ResultType };

export type StopFlowMutationVariables = Exact<{
    flowId: string | number;
}>;

export type StopFlowMutation = { stopFlow: ResultType };

export type RenameFlowMutationVariables = Exact<{
    flowId: string | number;
    title: string;
}>;

export type RenameFlowMutation = { renameFlow: ResultType };

export type CreateAssistantMutationVariables = Exact<{
    flowId: string | number;
    modelProvider: string;
    input: string;
    useAgents: boolean;
    resourceIds?: Array<string | number> | string | number | null | undefined;
}>;

export type CreateAssistantMutation = {
    createAssistant: { flow: FlowFragmentFragment; assistant: AssistantFragmentFragment };
};

export type CallAssistantMutationVariables = Exact<{
    flowId: string | number;
    assistantId: string | number;
    input: string;
    useAgents: boolean;
    resourceIds?: Array<string | number> | string | number | null | undefined;
}>;

export type CallAssistantMutation = { callAssistant: ResultType };

export type StopAssistantMutationVariables = Exact<{
    flowId: string | number;
    assistantId: string | number;
}>;

export type StopAssistantMutation = { stopAssistant: AssistantFragmentFragment };

export type DeleteAssistantMutationVariables = Exact<{
    flowId: string | number;
    assistantId: string | number;
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
    name: string;
    type: ProviderType;
    agents: AgentsConfigInput;
}>;

export type CreateProviderMutation = { createProvider: ProviderConfigFragmentFragment };

export type UpdateProviderMutationVariables = Exact<{
    providerId: string | number;
    name: string;
    agents: AgentsConfigInput;
}>;

export type UpdateProviderMutation = { updateProvider: ProviderConfigFragmentFragment };

export type DeleteProviderMutationVariables = Exact<{
    providerId: string | number;
}>;

export type DeleteProviderMutation = { deleteProvider: ResultType };

export type ValidatePromptMutationVariables = Exact<{
    type: PromptType;
    template: string;
}>;

export type ValidatePromptMutation = { validatePrompt: PromptValidationResultFragmentFragment };

export type CreatePromptMutationVariables = Exact<{
    type: PromptType;
    template: string;
}>;

export type CreatePromptMutation = { createPrompt: UserPromptFragmentFragment };

export type UpdatePromptMutationVariables = Exact<{
    promptId: string | number;
    template: string;
}>;

export type UpdatePromptMutation = { updatePrompt: UserPromptFragmentFragment };

export type DeletePromptMutationVariables = Exact<{
    promptId: string | number;
}>;

export type DeletePromptMutation = { deletePrompt: ResultType };

export type CreateApiTokenMutationVariables = Exact<{
    input: CreateApiTokenInput;
}>;

export type CreateApiTokenMutation = { createAPIToken: ApiTokenWithSecretFragmentFragment };

export type UpdateApiTokenMutationVariables = Exact<{
    tokenId: string;
    input: UpdateApiTokenInput;
}>;

export type UpdateApiTokenMutation = { updateAPIToken: ApiTokenFragmentFragment };

export type DeleteApiTokenMutationVariables = Exact<{
    tokenId: string;
}>;

export type DeleteApiTokenMutation = { deleteAPIToken: boolean };

export type CreateKnowledgeDocumentMutationVariables = Exact<{
    input: CreateKnowledgeDocumentInput;
}>;

export type CreateKnowledgeDocumentMutation = { createKnowledgeDocument: KnowledgeDocumentFragmentFragment };

export type UpdateKnowledgeDocumentMutationVariables = Exact<{
    id: string;
    input: UpdateKnowledgeDocumentInput;
}>;

export type UpdateKnowledgeDocumentMutation = { updateKnowledgeDocument: KnowledgeDocumentFragmentFragment };

export type DeleteKnowledgeDocumentMutationVariables = Exact<{
    id: string;
}>;

export type DeleteKnowledgeDocumentMutation = { deleteKnowledgeDocument: ResultType };

export type TerminalLogAddedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type TerminalLogAddedSubscription = { terminalLogAdded: TerminalLogFragmentFragment };

export type MessageLogAddedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type MessageLogAddedSubscription = { messageLogAdded: MessageLogFragmentFragment };

export type MessageLogUpdatedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type MessageLogUpdatedSubscription = { messageLogUpdated: MessageLogFragmentFragment };

export type ScreenshotAddedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type ScreenshotAddedSubscription = { screenshotAdded: ScreenshotFragmentFragment };

export type AgentLogAddedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type AgentLogAddedSubscription = { agentLogAdded: AgentLogFragmentFragment };

export type SearchLogAddedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type SearchLogAddedSubscription = { searchLogAdded: SearchLogFragmentFragment };

export type VectorStoreLogAddedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type VectorStoreLogAddedSubscription = { vectorStoreLogAdded: VectorStoreLogFragmentFragment };

export type AssistantCreatedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type AssistantCreatedSubscription = { assistantCreated: AssistantFragmentFragment };

export type AssistantUpdatedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type AssistantUpdatedSubscription = { assistantUpdated: AssistantFragmentFragment };

export type AssistantDeletedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type AssistantDeletedSubscription = { assistantDeleted: AssistantFragmentFragment };

export type FlowFileAddedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type FlowFileAddedSubscription = { flowFileAdded: FlowFileFragmentFragment };

export type FlowFileUpdatedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type FlowFileUpdatedSubscription = { flowFileUpdated: FlowFileFragmentFragment };

export type FlowFileDeletedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type FlowFileDeletedSubscription = { flowFileDeleted: FlowFileFragmentFragment };

export type AssistantLogAddedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type AssistantLogAddedSubscription = { assistantLogAdded: AssistantLogFragmentFragment };

export type AssistantLogUpdatedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type AssistantLogUpdatedSubscription = { assistantLogUpdated: AssistantLogFragmentFragment };

export type FlowCreatedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type FlowCreatedSubscription = { flowCreated: FlowFragmentFragment };

export type FlowDeletedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type FlowDeletedSubscription = { flowDeleted: FlowFragmentFragment };

export type FlowUpdatedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type FlowUpdatedSubscription = { flowUpdated: FlowFragmentFragment };

export type TaskCreatedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type TaskCreatedSubscription = { taskCreated: TaskFragmentFragment };

export type TaskUpdatedSubscriptionVariables = Exact<{
    flowId: string | number;
}>;

export type TaskUpdatedSubscription = {
    taskUpdated: {
        id: string;
        status: StatusType;
        result: string;
        updatedAt: unknown;
        subtasks: Array<SubtaskFragmentFragment> | null;
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

export type FlowTemplateCreatedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type FlowTemplateCreatedSubscription = { flowTemplateCreated: FlowTemplateFragmentFragment };

export type FlowTemplateUpdatedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type FlowTemplateUpdatedSubscription = { flowTemplateUpdated: FlowTemplateFragmentFragment };

export type FlowTemplateDeletedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type FlowTemplateDeletedSubscription = { flowTemplateDeleted: FlowTemplateFragmentFragment };

export type ResourceAddedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type ResourceAddedSubscription = { resourceAdded: UserResourceFragmentFragment };

export type ResourceUpdatedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type ResourceUpdatedSubscription = { resourceUpdated: UserResourceFragmentFragment };

export type ResourceDeletedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type ResourceDeletedSubscription = { resourceDeleted: UserResourceFragmentFragment };

export type KnowledgeDocumentCreatedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type KnowledgeDocumentCreatedSubscription = { knowledgeDocumentCreated: KnowledgeDocumentFragmentFragment };

export type KnowledgeDocumentUpdatedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type KnowledgeDocumentUpdatedSubscription = { knowledgeDocumentUpdated: KnowledgeDocumentFragmentFragment };

export type KnowledgeDocumentDeletedSubscriptionVariables = Exact<{ [key: string]: never }>;

export type KnowledgeDocumentDeletedSubscription = { knowledgeDocumentDeleted: KnowledgeDocumentFragmentFragment };

export const SettingsFragmentFragmentDoc = gql`
    fragment settingsFragment on Settings {
        debug
        askUser
        version
        dockerInside
        isDevelopMode
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
    ${TerminalFragmentFragmentDoc}
    ${ProviderFragmentFragmentDoc}
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
    ${SubtaskFragmentFragmentDoc}
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
export const FlowFileFragmentFragmentDoc = gql`
    fragment flowFileFragment on FlowFile {
        id
        name
        path
        size
        isDir
        modifiedAt
    }
`;
export const UserResourceFragmentFragmentDoc = gql`
    fragment userResourceFragment on UserResource {
        id
        userId
        name
        path
        size
        isDir
        createdAt
        updatedAt
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
    ${ProviderFragmentFragmentDoc}
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
    ${TestResultFragmentFragmentDoc}
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
    ${AgentTestResultFragmentFragmentDoc}
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
    ${AgentConfigFragmentFragmentDoc}
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
    ${AgentsConfigFragmentFragmentDoc}
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
export const FlowTemplateFragmentFragmentDoc = gql`
    fragment flowTemplateFragment on FlowTemplate {
        id
        userId
        title
        text
        createdAt
        updatedAt
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
    ${UsageStatsFragmentFragmentDoc}
`;
export const ProviderUsageStatsFragmentFragmentDoc = gql`
    fragment providerUsageStatsFragment on ProviderUsageStats {
        provider
        stats {
            ...usageStatsFragment
        }
    }
    ${UsageStatsFragmentFragmentDoc}
`;
export const ModelUsageStatsFragmentFragmentDoc = gql`
    fragment modelUsageStatsFragment on ModelUsageStats {
        model
        provider
        stats {
            ...usageStatsFragment
        }
    }
    ${UsageStatsFragmentFragmentDoc}
`;
export const AgentTypeUsageStatsFragmentFragmentDoc = gql`
    fragment agentTypeUsageStatsFragment on AgentTypeUsageStats {
        agentType
        stats {
            ...usageStatsFragment
        }
    }
    ${UsageStatsFragmentFragmentDoc}
`;
export const ModelAgentsUsageStatsFragmentFragmentDoc = gql`
    fragment modelAgentsUsageStatsFragment on ModelAgentsUsageStats {
        model
        provider
        agentTypes
        stats {
            ...usageStatsFragment
        }
    }
    ${UsageStatsFragmentFragmentDoc}
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
    ${ToolcallsStatsFragmentFragmentDoc}
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
    ${FlowsStatsFragmentFragmentDoc}
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
    ${SubtaskExecutionStatsFragmentFragmentDoc}
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
    ${TaskExecutionStatsFragmentFragmentDoc}
`;
export const KnowledgeDocumentFragmentFragmentDoc = gql`
    fragment knowledgeDocumentFragment on KnowledgeDocument {
        id
        docType
        content
        question
        description
        userId
        flowId
        taskId
        subtaskId
        guideType
        answerType
        codeLang
        partSize
        totalSize
        manual
    }
`;
export const KnowledgeDocumentWithScoreFragmentFragmentDoc = gql`
    fragment knowledgeDocumentWithScoreFragment on KnowledgeDocumentWithScore {
        score
        document {
            ...knowledgeDocumentFragment
        }
    }
    ${KnowledgeDocumentFragmentFragmentDoc}
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
export function useFlowsQuery(baseOptions?: ApolloReactHooks.QueryHookOptions<FlowsQuery, FlowsQueryVariables>) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<FlowsQuery, FlowsQueryVariables>(FlowsDocument, options);
}
export function useFlowsLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<FlowsQuery, FlowsQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<FlowsQuery, FlowsQueryVariables>(FlowsDocument, options);
}
// @ts-ignore
export function useFlowsSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<FlowsQuery, FlowsQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowsQuery, FlowsQueryVariables>;
export function useFlowsSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<FlowsQuery, FlowsQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowsQuery | undefined, FlowsQueryVariables>;
export function useFlowsSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<FlowsQuery, FlowsQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<FlowsQuery, FlowsQueryVariables>(FlowsDocument, options);
}
export type FlowsQueryHookResult = ReturnType<typeof useFlowsQuery>;
export type FlowsLazyQueryHookResult = ReturnType<typeof useFlowsLazyQuery>;
export type FlowsSuspenseQueryHookResult = ReturnType<typeof useFlowsSuspenseQuery>;
export type FlowsQueryResult = ApolloReactCommon.QueryResult<FlowsQuery, FlowsQueryVariables>;
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
export function useProvidersQuery(
    baseOptions?: ApolloReactHooks.QueryHookOptions<ProvidersQuery, ProvidersQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<ProvidersQuery, ProvidersQueryVariables>(ProvidersDocument, options);
}
export function useProvidersLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<ProvidersQuery, ProvidersQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<ProvidersQuery, ProvidersQueryVariables>(ProvidersDocument, options);
}
// @ts-ignore
export function useProvidersSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<ProvidersQuery, ProvidersQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<ProvidersQuery, ProvidersQueryVariables>;
export function useProvidersSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<ProvidersQuery, ProvidersQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<ProvidersQuery | undefined, ProvidersQueryVariables>;
export function useProvidersSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<ProvidersQuery, ProvidersQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<ProvidersQuery, ProvidersQueryVariables>(ProvidersDocument, options);
}
export type ProvidersQueryHookResult = ReturnType<typeof useProvidersQuery>;
export type ProvidersLazyQueryHookResult = ReturnType<typeof useProvidersLazyQuery>;
export type ProvidersSuspenseQueryHookResult = ReturnType<typeof useProvidersSuspenseQuery>;
export type ProvidersQueryResult = ApolloReactCommon.QueryResult<ProvidersQuery, ProvidersQueryVariables>;
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
export function useSettingsQuery(
    baseOptions?: ApolloReactHooks.QueryHookOptions<SettingsQuery, SettingsQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<SettingsQuery, SettingsQueryVariables>(SettingsDocument, options);
}
export function useSettingsLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<SettingsQuery, SettingsQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<SettingsQuery, SettingsQueryVariables>(SettingsDocument, options);
}
// @ts-ignore
export function useSettingsSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<SettingsQuery, SettingsQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<SettingsQuery, SettingsQueryVariables>;
export function useSettingsSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<SettingsQuery, SettingsQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<SettingsQuery | undefined, SettingsQueryVariables>;
export function useSettingsSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<SettingsQuery, SettingsQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<SettingsQuery, SettingsQueryVariables>(SettingsDocument, options);
}
export type SettingsQueryHookResult = ReturnType<typeof useSettingsQuery>;
export type SettingsLazyQueryHookResult = ReturnType<typeof useSettingsLazyQuery>;
export type SettingsSuspenseQueryHookResult = ReturnType<typeof useSettingsSuspenseQuery>;
export type SettingsQueryResult = ApolloReactCommon.QueryResult<SettingsQuery, SettingsQueryVariables>;
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
                deepseek
                glm
                kimi
                qwen
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
                deepseek {
                    ...providerConfigFragment
                }
                glm {
                    ...providerConfigFragment
                }
                kimi {
                    ...providerConfigFragment
                }
                qwen {
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
                deepseek {
                    ...modelConfigFragment
                }
                glm {
                    ...modelConfigFragment
                }
                kimi {
                    ...modelConfigFragment
                }
                qwen {
                    ...modelConfigFragment
                }
            }
        }
    }
    ${ProviderConfigFragmentFragmentDoc}
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
    baseOptions?: ApolloReactHooks.QueryHookOptions<SettingsProvidersQuery, SettingsProvidersQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<SettingsProvidersQuery, SettingsProvidersQueryVariables>(
        SettingsProvidersDocument,
        options,
    );
}
export function useSettingsProvidersLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<SettingsProvidersQuery, SettingsProvidersQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<SettingsProvidersQuery, SettingsProvidersQueryVariables>(
        SettingsProvidersDocument,
        options,
    );
}
// @ts-ignore
export function useSettingsProvidersSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<SettingsProvidersQuery, SettingsProvidersQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<SettingsProvidersQuery, SettingsProvidersQueryVariables>;
export function useSettingsProvidersSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<SettingsProvidersQuery, SettingsProvidersQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<SettingsProvidersQuery | undefined, SettingsProvidersQueryVariables>;
export function useSettingsProvidersSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<SettingsProvidersQuery, SettingsProvidersQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<SettingsProvidersQuery, SettingsProvidersQueryVariables>(
        SettingsProvidersDocument,
        options,
    );
}
export type SettingsProvidersQueryHookResult = ReturnType<typeof useSettingsProvidersQuery>;
export type SettingsProvidersLazyQueryHookResult = ReturnType<typeof useSettingsProvidersLazyQuery>;
export type SettingsProvidersSuspenseQueryHookResult = ReturnType<typeof useSettingsProvidersSuspenseQuery>;
export type SettingsProvidersQueryResult = ApolloReactCommon.QueryResult<
    SettingsProvidersQuery,
    SettingsProvidersQueryVariables
>;
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
                    collectToolCallId {
                        ...defaultPromptFragment
                    }
                    detectToolCallIdPattern {
                        ...defaultPromptFragment
                    }
                    monitorAgentExecution {
                        ...defaultPromptFragment
                    }
                    planAgentTask {
                        ...defaultPromptFragment
                    }
                    wrapAgentTask {
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
    baseOptions?: ApolloReactHooks.QueryHookOptions<SettingsPromptsQuery, SettingsPromptsQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<SettingsPromptsQuery, SettingsPromptsQueryVariables>(
        SettingsPromptsDocument,
        options,
    );
}
export function useSettingsPromptsLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<SettingsPromptsQuery, SettingsPromptsQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<SettingsPromptsQuery, SettingsPromptsQueryVariables>(
        SettingsPromptsDocument,
        options,
    );
}
// @ts-ignore
export function useSettingsPromptsSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<SettingsPromptsQuery, SettingsPromptsQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<SettingsPromptsQuery, SettingsPromptsQueryVariables>;
export function useSettingsPromptsSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<SettingsPromptsQuery, SettingsPromptsQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<SettingsPromptsQuery | undefined, SettingsPromptsQueryVariables>;
export function useSettingsPromptsSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<SettingsPromptsQuery, SettingsPromptsQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<SettingsPromptsQuery, SettingsPromptsQueryVariables>(
        SettingsPromptsDocument,
        options,
    );
}
export type SettingsPromptsQueryHookResult = ReturnType<typeof useSettingsPromptsQuery>;
export type SettingsPromptsLazyQueryHookResult = ReturnType<typeof useSettingsPromptsLazyQuery>;
export type SettingsPromptsSuspenseQueryHookResult = ReturnType<typeof useSettingsPromptsSuspenseQuery>;
export type SettingsPromptsQueryResult = ApolloReactCommon.QueryResult<
    SettingsPromptsQuery,
    SettingsPromptsQueryVariables
>;
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
    baseOptions: ApolloReactHooks.QueryHookOptions<FlowQuery, FlowQueryVariables> &
        ({ variables: FlowQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<FlowQuery, FlowQueryVariables>(FlowDocument, options);
}
export function useFlowLazyQuery(baseOptions?: ApolloReactHooks.LazyQueryHookOptions<FlowQuery, FlowQueryVariables>) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<FlowQuery, FlowQueryVariables>(FlowDocument, options);
}
// @ts-ignore
export function useFlowSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<FlowQuery, FlowQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowQuery, FlowQueryVariables>;
export function useFlowSuspenseQuery(
    baseOptions?: ApolloReactHooks.SkipToken | ApolloReactHooks.SuspenseQueryHookOptions<FlowQuery, FlowQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowQuery | undefined, FlowQueryVariables>;
export function useFlowSuspenseQuery(
    baseOptions?: ApolloReactHooks.SkipToken | ApolloReactHooks.SuspenseQueryHookOptions<FlowQuery, FlowQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<FlowQuery, FlowQueryVariables>(FlowDocument, options);
}
export type FlowQueryHookResult = ReturnType<typeof useFlowQuery>;
export type FlowLazyQueryHookResult = ReturnType<typeof useFlowLazyQuery>;
export type FlowSuspenseQueryHookResult = ReturnType<typeof useFlowSuspenseQuery>;
export type FlowQueryResult = ApolloReactCommon.QueryResult<FlowQuery, FlowQueryVariables>;
export const TasksDocument = gql`
    query tasks($flowId: ID!) {
        tasks(flowId: $flowId) {
            ...taskFragment
        }
    }
    ${TaskFragmentFragmentDoc}
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
    baseOptions: ApolloReactHooks.QueryHookOptions<TasksQuery, TasksQueryVariables> &
        ({ variables: TasksQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<TasksQuery, TasksQueryVariables>(TasksDocument, options);
}
export function useTasksLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<TasksQuery, TasksQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<TasksQuery, TasksQueryVariables>(TasksDocument, options);
}
// @ts-ignore
export function useTasksSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<TasksQuery, TasksQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<TasksQuery, TasksQueryVariables>;
export function useTasksSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<TasksQuery, TasksQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<TasksQuery | undefined, TasksQueryVariables>;
export function useTasksSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<TasksQuery, TasksQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<TasksQuery, TasksQueryVariables>(TasksDocument, options);
}
export type TasksQueryHookResult = ReturnType<typeof useTasksQuery>;
export type TasksLazyQueryHookResult = ReturnType<typeof useTasksLazyQuery>;
export type TasksSuspenseQueryHookResult = ReturnType<typeof useTasksSuspenseQuery>;
export type TasksQueryResult = ApolloReactCommon.QueryResult<TasksQuery, TasksQueryVariables>;
export const FlowFilesDocument = gql`
    query flowFiles($flowId: ID!) {
        flowFiles(flowId: $flowId) {
            ...flowFileFragment
        }
    }
    ${FlowFileFragmentFragmentDoc}
`;

/**
 * __useFlowFilesQuery__
 *
 * To run a query within a React component, call `useFlowFilesQuery` and pass it any options that fit your needs.
 * When your component renders, `useFlowFilesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowFilesQuery({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useFlowFilesQuery(
    baseOptions: ApolloReactHooks.QueryHookOptions<FlowFilesQuery, FlowFilesQueryVariables> &
        ({ variables: FlowFilesQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<FlowFilesQuery, FlowFilesQueryVariables>(FlowFilesDocument, options);
}
export function useFlowFilesLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<FlowFilesQuery, FlowFilesQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<FlowFilesQuery, FlowFilesQueryVariables>(FlowFilesDocument, options);
}
// @ts-ignore
export function useFlowFilesSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<FlowFilesQuery, FlowFilesQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowFilesQuery, FlowFilesQueryVariables>;
export function useFlowFilesSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<FlowFilesQuery, FlowFilesQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowFilesQuery | undefined, FlowFilesQueryVariables>;
export function useFlowFilesSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<FlowFilesQuery, FlowFilesQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<FlowFilesQuery, FlowFilesQueryVariables>(FlowFilesDocument, options);
}
export type FlowFilesQueryHookResult = ReturnType<typeof useFlowFilesQuery>;
export type FlowFilesLazyQueryHookResult = ReturnType<typeof useFlowFilesLazyQuery>;
export type FlowFilesSuspenseQueryHookResult = ReturnType<typeof useFlowFilesSuspenseQuery>;
export type FlowFilesQueryResult = ApolloReactCommon.QueryResult<FlowFilesQuery, FlowFilesQueryVariables>;
export const ResourcesDocument = gql`
    query resources($path: String, $recursive: Boolean) {
        resources(path: $path, recursive: $recursive) {
            ...userResourceFragment
        }
    }
    ${UserResourceFragmentFragmentDoc}
`;

/**
 * __useResourcesQuery__
 *
 * To run a query within a React component, call `useResourcesQuery` and pass it any options that fit your needs.
 * When your component renders, `useResourcesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useResourcesQuery({
 *   variables: {
 *      path: // value for 'path'
 *      recursive: // value for 'recursive'
 *   },
 * });
 */
export function useResourcesQuery(
    baseOptions?: ApolloReactHooks.QueryHookOptions<ResourcesQuery, ResourcesQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<ResourcesQuery, ResourcesQueryVariables>(ResourcesDocument, options);
}
export function useResourcesLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<ResourcesQuery, ResourcesQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<ResourcesQuery, ResourcesQueryVariables>(ResourcesDocument, options);
}
// @ts-ignore
export function useResourcesSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<ResourcesQuery, ResourcesQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<ResourcesQuery, ResourcesQueryVariables>;
export function useResourcesSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<ResourcesQuery, ResourcesQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<ResourcesQuery | undefined, ResourcesQueryVariables>;
export function useResourcesSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<ResourcesQuery, ResourcesQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<ResourcesQuery, ResourcesQueryVariables>(ResourcesDocument, options);
}
export type ResourcesQueryHookResult = ReturnType<typeof useResourcesQuery>;
export type ResourcesLazyQueryHookResult = ReturnType<typeof useResourcesLazyQuery>;
export type ResourcesSuspenseQueryHookResult = ReturnType<typeof useResourcesSuspenseQuery>;
export type ResourcesQueryResult = ApolloReactCommon.QueryResult<ResourcesQuery, ResourcesQueryVariables>;
export const AssistantsDocument = gql`
    query assistants($flowId: ID!) {
        assistants(flowId: $flowId) {
            ...assistantFragment
        }
    }
    ${AssistantFragmentFragmentDoc}
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
    baseOptions: ApolloReactHooks.QueryHookOptions<AssistantsQuery, AssistantsQueryVariables> &
        ({ variables: AssistantsQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<AssistantsQuery, AssistantsQueryVariables>(AssistantsDocument, options);
}
export function useAssistantsLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<AssistantsQuery, AssistantsQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<AssistantsQuery, AssistantsQueryVariables>(AssistantsDocument, options);
}
// @ts-ignore
export function useAssistantsSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<AssistantsQuery, AssistantsQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<AssistantsQuery, AssistantsQueryVariables>;
export function useAssistantsSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<AssistantsQuery, AssistantsQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<AssistantsQuery | undefined, AssistantsQueryVariables>;
export function useAssistantsSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<AssistantsQuery, AssistantsQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<AssistantsQuery, AssistantsQueryVariables>(AssistantsDocument, options);
}
export type AssistantsQueryHookResult = ReturnType<typeof useAssistantsQuery>;
export type AssistantsLazyQueryHookResult = ReturnType<typeof useAssistantsLazyQuery>;
export type AssistantsSuspenseQueryHookResult = ReturnType<typeof useAssistantsSuspenseQuery>;
export type AssistantsQueryResult = ApolloReactCommon.QueryResult<AssistantsQuery, AssistantsQueryVariables>;
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
    baseOptions: ApolloReactHooks.QueryHookOptions<AssistantLogsQuery, AssistantLogsQueryVariables> &
        ({ variables: AssistantLogsQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<AssistantLogsQuery, AssistantLogsQueryVariables>(AssistantLogsDocument, options);
}
export function useAssistantLogsLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<AssistantLogsQuery, AssistantLogsQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<AssistantLogsQuery, AssistantLogsQueryVariables>(
        AssistantLogsDocument,
        options,
    );
}
// @ts-ignore
export function useAssistantLogsSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<AssistantLogsQuery, AssistantLogsQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<AssistantLogsQuery, AssistantLogsQueryVariables>;
export function useAssistantLogsSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<AssistantLogsQuery, AssistantLogsQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<AssistantLogsQuery | undefined, AssistantLogsQueryVariables>;
export function useAssistantLogsSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<AssistantLogsQuery, AssistantLogsQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<AssistantLogsQuery, AssistantLogsQueryVariables>(
        AssistantLogsDocument,
        options,
    );
}
export type AssistantLogsQueryHookResult = ReturnType<typeof useAssistantLogsQuery>;
export type AssistantLogsLazyQueryHookResult = ReturnType<typeof useAssistantLogsLazyQuery>;
export type AssistantLogsSuspenseQueryHookResult = ReturnType<typeof useAssistantLogsSuspenseQuery>;
export type AssistantLogsQueryResult = ApolloReactCommon.QueryResult<AssistantLogsQuery, AssistantLogsQueryVariables>;
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
    ${TaskFragmentFragmentDoc}
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
    baseOptions: ApolloReactHooks.QueryHookOptions<FlowReportQuery, FlowReportQueryVariables> &
        ({ variables: FlowReportQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<FlowReportQuery, FlowReportQueryVariables>(FlowReportDocument, options);
}
export function useFlowReportLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<FlowReportQuery, FlowReportQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<FlowReportQuery, FlowReportQueryVariables>(FlowReportDocument, options);
}
// @ts-ignore
export function useFlowReportSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<FlowReportQuery, FlowReportQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowReportQuery, FlowReportQueryVariables>;
export function useFlowReportSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<FlowReportQuery, FlowReportQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowReportQuery | undefined, FlowReportQueryVariables>;
export function useFlowReportSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<FlowReportQuery, FlowReportQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<FlowReportQuery, FlowReportQueryVariables>(FlowReportDocument, options);
}
export type FlowReportQueryHookResult = ReturnType<typeof useFlowReportQuery>;
export type FlowReportLazyQueryHookResult = ReturnType<typeof useFlowReportLazyQuery>;
export type FlowReportSuspenseQueryHookResult = ReturnType<typeof useFlowReportSuspenseQuery>;
export type FlowReportQueryResult = ApolloReactCommon.QueryResult<FlowReportQuery, FlowReportQueryVariables>;
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
    baseOptions?: ApolloReactHooks.QueryHookOptions<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>(
        UsageStatsTotalDocument,
        options,
    );
}
export function useUsageStatsTotalLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>(
        UsageStatsTotalDocument,
        options,
    );
}
// @ts-ignore
export function useUsageStatsTotalSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>;
export function useUsageStatsTotalSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<UsageStatsTotalQuery | undefined, UsageStatsTotalQueryVariables>;
export function useUsageStatsTotalSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>(
        UsageStatsTotalDocument,
        options,
    );
}
export type UsageStatsTotalQueryHookResult = ReturnType<typeof useUsageStatsTotalQuery>;
export type UsageStatsTotalLazyQueryHookResult = ReturnType<typeof useUsageStatsTotalLazyQuery>;
export type UsageStatsTotalSuspenseQueryHookResult = ReturnType<typeof useUsageStatsTotalSuspenseQuery>;
export type UsageStatsTotalQueryResult = ApolloReactCommon.QueryResult<
    UsageStatsTotalQuery,
    UsageStatsTotalQueryVariables
>;
export const UsageStatsByPeriodDocument = gql`
    query usageStatsByPeriod($period: UsageStatsPeriod!) {
        usageStatsByPeriod(period: $period) {
            ...dailyUsageStatsFragment
        }
    }
    ${DailyUsageStatsFragmentFragmentDoc}
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
    baseOptions: ApolloReactHooks.QueryHookOptions<UsageStatsByPeriodQuery, UsageStatsByPeriodQueryVariables> &
        ({ variables: UsageStatsByPeriodQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<UsageStatsByPeriodQuery, UsageStatsByPeriodQueryVariables>(
        UsageStatsByPeriodDocument,
        options,
    );
}
export function useUsageStatsByPeriodLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<UsageStatsByPeriodQuery, UsageStatsByPeriodQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<UsageStatsByPeriodQuery, UsageStatsByPeriodQueryVariables>(
        UsageStatsByPeriodDocument,
        options,
    );
}
// @ts-ignore
export function useUsageStatsByPeriodSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<UsageStatsByPeriodQuery, UsageStatsByPeriodQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<UsageStatsByPeriodQuery, UsageStatsByPeriodQueryVariables>;
export function useUsageStatsByPeriodSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<UsageStatsByPeriodQuery, UsageStatsByPeriodQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<UsageStatsByPeriodQuery | undefined, UsageStatsByPeriodQueryVariables>;
export function useUsageStatsByPeriodSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<UsageStatsByPeriodQuery, UsageStatsByPeriodQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<UsageStatsByPeriodQuery, UsageStatsByPeriodQueryVariables>(
        UsageStatsByPeriodDocument,
        options,
    );
}
export type UsageStatsByPeriodQueryHookResult = ReturnType<typeof useUsageStatsByPeriodQuery>;
export type UsageStatsByPeriodLazyQueryHookResult = ReturnType<typeof useUsageStatsByPeriodLazyQuery>;
export type UsageStatsByPeriodSuspenseQueryHookResult = ReturnType<typeof useUsageStatsByPeriodSuspenseQuery>;
export type UsageStatsByPeriodQueryResult = ApolloReactCommon.QueryResult<
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
    baseOptions?: ApolloReactHooks.QueryHookOptions<UsageStatsByProviderQuery, UsageStatsByProviderQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<UsageStatsByProviderQuery, UsageStatsByProviderQueryVariables>(
        UsageStatsByProviderDocument,
        options,
    );
}
export function useUsageStatsByProviderLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<UsageStatsByProviderQuery, UsageStatsByProviderQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<UsageStatsByProviderQuery, UsageStatsByProviderQueryVariables>(
        UsageStatsByProviderDocument,
        options,
    );
}
// @ts-ignore
export function useUsageStatsByProviderSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<
        UsageStatsByProviderQuery,
        UsageStatsByProviderQueryVariables
    >,
): ApolloReactHooks.UseSuspenseQueryResult<UsageStatsByProviderQuery, UsageStatsByProviderQueryVariables>;
export function useUsageStatsByProviderSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<UsageStatsByProviderQuery, UsageStatsByProviderQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<UsageStatsByProviderQuery | undefined, UsageStatsByProviderQueryVariables>;
export function useUsageStatsByProviderSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<UsageStatsByProviderQuery, UsageStatsByProviderQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<UsageStatsByProviderQuery, UsageStatsByProviderQueryVariables>(
        UsageStatsByProviderDocument,
        options,
    );
}
export type UsageStatsByProviderQueryHookResult = ReturnType<typeof useUsageStatsByProviderQuery>;
export type UsageStatsByProviderLazyQueryHookResult = ReturnType<typeof useUsageStatsByProviderLazyQuery>;
export type UsageStatsByProviderSuspenseQueryHookResult = ReturnType<typeof useUsageStatsByProviderSuspenseQuery>;
export type UsageStatsByProviderQueryResult = ApolloReactCommon.QueryResult<
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
    baseOptions?: ApolloReactHooks.QueryHookOptions<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>(
        UsageStatsByModelDocument,
        options,
    );
}
export function useUsageStatsByModelLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>(
        UsageStatsByModelDocument,
        options,
    );
}
// @ts-ignore
export function useUsageStatsByModelSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>;
export function useUsageStatsByModelSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<UsageStatsByModelQuery | undefined, UsageStatsByModelQueryVariables>;
export function useUsageStatsByModelSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>(
        UsageStatsByModelDocument,
        options,
    );
}
export type UsageStatsByModelQueryHookResult = ReturnType<typeof useUsageStatsByModelQuery>;
export type UsageStatsByModelLazyQueryHookResult = ReturnType<typeof useUsageStatsByModelLazyQuery>;
export type UsageStatsByModelSuspenseQueryHookResult = ReturnType<typeof useUsageStatsByModelSuspenseQuery>;
export type UsageStatsByModelQueryResult = ApolloReactCommon.QueryResult<
    UsageStatsByModelQuery,
    UsageStatsByModelQueryVariables
>;
export const UsageStatsByAgentTypeDocument = gql`
    query usageStatsByAgentType {
        usageStatsByAgentType {
            ...agentTypeUsageStatsFragment
        }
    }
    ${AgentTypeUsageStatsFragmentFragmentDoc}
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
    baseOptions?: ApolloReactHooks.QueryHookOptions<UsageStatsByAgentTypeQuery, UsageStatsByAgentTypeQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<UsageStatsByAgentTypeQuery, UsageStatsByAgentTypeQueryVariables>(
        UsageStatsByAgentTypeDocument,
        options,
    );
}
export function useUsageStatsByAgentTypeLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<
        UsageStatsByAgentTypeQuery,
        UsageStatsByAgentTypeQueryVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<UsageStatsByAgentTypeQuery, UsageStatsByAgentTypeQueryVariables>(
        UsageStatsByAgentTypeDocument,
        options,
    );
}
// @ts-ignore
export function useUsageStatsByAgentTypeSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<
        UsageStatsByAgentTypeQuery,
        UsageStatsByAgentTypeQueryVariables
    >,
): ApolloReactHooks.UseSuspenseQueryResult<UsageStatsByAgentTypeQuery, UsageStatsByAgentTypeQueryVariables>;
export function useUsageStatsByAgentTypeSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<UsageStatsByAgentTypeQuery, UsageStatsByAgentTypeQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<UsageStatsByAgentTypeQuery | undefined, UsageStatsByAgentTypeQueryVariables>;
export function useUsageStatsByAgentTypeSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<UsageStatsByAgentTypeQuery, UsageStatsByAgentTypeQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<UsageStatsByAgentTypeQuery, UsageStatsByAgentTypeQueryVariables>(
        UsageStatsByAgentTypeDocument,
        options,
    );
}
export type UsageStatsByAgentTypeQueryHookResult = ReturnType<typeof useUsageStatsByAgentTypeQuery>;
export type UsageStatsByAgentTypeLazyQueryHookResult = ReturnType<typeof useUsageStatsByAgentTypeLazyQuery>;
export type UsageStatsByAgentTypeSuspenseQueryHookResult = ReturnType<typeof useUsageStatsByAgentTypeSuspenseQuery>;
export type UsageStatsByAgentTypeQueryResult = ApolloReactCommon.QueryResult<
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
    baseOptions: ApolloReactHooks.QueryHookOptions<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables> &
        ({ variables: UsageStatsByFlowQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables>(
        UsageStatsByFlowDocument,
        options,
    );
}
export function useUsageStatsByFlowLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables>(
        UsageStatsByFlowDocument,
        options,
    );
}
// @ts-ignore
export function useUsageStatsByFlowSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables>;
export function useUsageStatsByFlowSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<UsageStatsByFlowQuery | undefined, UsageStatsByFlowQueryVariables>;
export function useUsageStatsByFlowSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables>(
        UsageStatsByFlowDocument,
        options,
    );
}
export type UsageStatsByFlowQueryHookResult = ReturnType<typeof useUsageStatsByFlowQuery>;
export type UsageStatsByFlowLazyQueryHookResult = ReturnType<typeof useUsageStatsByFlowLazyQuery>;
export type UsageStatsByFlowSuspenseQueryHookResult = ReturnType<typeof useUsageStatsByFlowSuspenseQuery>;
export type UsageStatsByFlowQueryResult = ApolloReactCommon.QueryResult<
    UsageStatsByFlowQuery,
    UsageStatsByFlowQueryVariables
>;
export const UsageStatsByAgentTypeForFlowDocument = gql`
    query usageStatsByAgentTypeForFlow($flowId: ID!) {
        usageStatsByAgentTypeForFlow(flowId: $flowId) {
            ...agentTypeUsageStatsFragment
        }
    }
    ${AgentTypeUsageStatsFragmentFragmentDoc}
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
    baseOptions: ApolloReactHooks.QueryHookOptions<
        UsageStatsByAgentTypeForFlowQuery,
        UsageStatsByAgentTypeForFlowQueryVariables
    > &
        ({ variables: UsageStatsByAgentTypeForFlowQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<UsageStatsByAgentTypeForFlowQuery, UsageStatsByAgentTypeForFlowQueryVariables>(
        UsageStatsByAgentTypeForFlowDocument,
        options,
    );
}
export function useUsageStatsByAgentTypeForFlowLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<
        UsageStatsByAgentTypeForFlowQuery,
        UsageStatsByAgentTypeForFlowQueryVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<UsageStatsByAgentTypeForFlowQuery, UsageStatsByAgentTypeForFlowQueryVariables>(
        UsageStatsByAgentTypeForFlowDocument,
        options,
    );
}
// @ts-ignore
export function useUsageStatsByAgentTypeForFlowSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<
        UsageStatsByAgentTypeForFlowQuery,
        UsageStatsByAgentTypeForFlowQueryVariables
    >,
): ApolloReactHooks.UseSuspenseQueryResult<
    UsageStatsByAgentTypeForFlowQuery,
    UsageStatsByAgentTypeForFlowQueryVariables
>;
export function useUsageStatsByAgentTypeForFlowSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<
              UsageStatsByAgentTypeForFlowQuery,
              UsageStatsByAgentTypeForFlowQueryVariables
          >,
): ApolloReactHooks.UseSuspenseQueryResult<
    UsageStatsByAgentTypeForFlowQuery | undefined,
    UsageStatsByAgentTypeForFlowQueryVariables
>;
export function useUsageStatsByAgentTypeForFlowSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<
              UsageStatsByAgentTypeForFlowQuery,
              UsageStatsByAgentTypeForFlowQueryVariables
          >,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<
        UsageStatsByAgentTypeForFlowQuery,
        UsageStatsByAgentTypeForFlowQueryVariables
    >(UsageStatsByAgentTypeForFlowDocument, options);
}
export type UsageStatsByAgentTypeForFlowQueryHookResult = ReturnType<typeof useUsageStatsByAgentTypeForFlowQuery>;
export type UsageStatsByAgentTypeForFlowLazyQueryHookResult = ReturnType<
    typeof useUsageStatsByAgentTypeForFlowLazyQuery
>;
export type UsageStatsByAgentTypeForFlowSuspenseQueryHookResult = ReturnType<
    typeof useUsageStatsByAgentTypeForFlowSuspenseQuery
>;
export type UsageStatsByAgentTypeForFlowQueryResult = ApolloReactCommon.QueryResult<
    UsageStatsByAgentTypeForFlowQuery,
    UsageStatsByAgentTypeForFlowQueryVariables
>;
export const UsageStatsByModelAgentsForFlowDocument = gql`
    query usageStatsByModelAgentsForFlow($flowId: ID!) {
        usageStatsByModelAgentsForFlow(flowId: $flowId) {
            ...modelAgentsUsageStatsFragment
        }
    }
    ${ModelAgentsUsageStatsFragmentFragmentDoc}
`;

/**
 * __useUsageStatsByModelAgentsForFlowQuery__
 *
 * To run a query within a React component, call `useUsageStatsByModelAgentsForFlowQuery` and pass it any options that fit your needs.
 * When your component renders, `useUsageStatsByModelAgentsForFlowQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useUsageStatsByModelAgentsForFlowQuery({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useUsageStatsByModelAgentsForFlowQuery(
    baseOptions: ApolloReactHooks.QueryHookOptions<
        UsageStatsByModelAgentsForFlowQuery,
        UsageStatsByModelAgentsForFlowQueryVariables
    > &
        ({ variables: UsageStatsByModelAgentsForFlowQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<UsageStatsByModelAgentsForFlowQuery, UsageStatsByModelAgentsForFlowQueryVariables>(
        UsageStatsByModelAgentsForFlowDocument,
        options,
    );
}
export function useUsageStatsByModelAgentsForFlowLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<
        UsageStatsByModelAgentsForFlowQuery,
        UsageStatsByModelAgentsForFlowQueryVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<
        UsageStatsByModelAgentsForFlowQuery,
        UsageStatsByModelAgentsForFlowQueryVariables
    >(UsageStatsByModelAgentsForFlowDocument, options);
}
// @ts-ignore
export function useUsageStatsByModelAgentsForFlowSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<
        UsageStatsByModelAgentsForFlowQuery,
        UsageStatsByModelAgentsForFlowQueryVariables
    >,
): ApolloReactHooks.UseSuspenseQueryResult<
    UsageStatsByModelAgentsForFlowQuery,
    UsageStatsByModelAgentsForFlowQueryVariables
>;
export function useUsageStatsByModelAgentsForFlowSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<
              UsageStatsByModelAgentsForFlowQuery,
              UsageStatsByModelAgentsForFlowQueryVariables
          >,
): ApolloReactHooks.UseSuspenseQueryResult<
    UsageStatsByModelAgentsForFlowQuery | undefined,
    UsageStatsByModelAgentsForFlowQueryVariables
>;
export function useUsageStatsByModelAgentsForFlowSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<
              UsageStatsByModelAgentsForFlowQuery,
              UsageStatsByModelAgentsForFlowQueryVariables
          >,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<
        UsageStatsByModelAgentsForFlowQuery,
        UsageStatsByModelAgentsForFlowQueryVariables
    >(UsageStatsByModelAgentsForFlowDocument, options);
}
export type UsageStatsByModelAgentsForFlowQueryHookResult = ReturnType<typeof useUsageStatsByModelAgentsForFlowQuery>;
export type UsageStatsByModelAgentsForFlowLazyQueryHookResult = ReturnType<
    typeof useUsageStatsByModelAgentsForFlowLazyQuery
>;
export type UsageStatsByModelAgentsForFlowSuspenseQueryHookResult = ReturnType<
    typeof useUsageStatsByModelAgentsForFlowSuspenseQuery
>;
export type UsageStatsByModelAgentsForFlowQueryResult = ApolloReactCommon.QueryResult<
    UsageStatsByModelAgentsForFlowQuery,
    UsageStatsByModelAgentsForFlowQueryVariables
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
    baseOptions?: ApolloReactHooks.QueryHookOptions<ToolcallsStatsTotalQuery, ToolcallsStatsTotalQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<ToolcallsStatsTotalQuery, ToolcallsStatsTotalQueryVariables>(
        ToolcallsStatsTotalDocument,
        options,
    );
}
export function useToolcallsStatsTotalLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<ToolcallsStatsTotalQuery, ToolcallsStatsTotalQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<ToolcallsStatsTotalQuery, ToolcallsStatsTotalQueryVariables>(
        ToolcallsStatsTotalDocument,
        options,
    );
}
// @ts-ignore
export function useToolcallsStatsTotalSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<
        ToolcallsStatsTotalQuery,
        ToolcallsStatsTotalQueryVariables
    >,
): ApolloReactHooks.UseSuspenseQueryResult<ToolcallsStatsTotalQuery, ToolcallsStatsTotalQueryVariables>;
export function useToolcallsStatsTotalSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<ToolcallsStatsTotalQuery, ToolcallsStatsTotalQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<ToolcallsStatsTotalQuery | undefined, ToolcallsStatsTotalQueryVariables>;
export function useToolcallsStatsTotalSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<ToolcallsStatsTotalQuery, ToolcallsStatsTotalQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<ToolcallsStatsTotalQuery, ToolcallsStatsTotalQueryVariables>(
        ToolcallsStatsTotalDocument,
        options,
    );
}
export type ToolcallsStatsTotalQueryHookResult = ReturnType<typeof useToolcallsStatsTotalQuery>;
export type ToolcallsStatsTotalLazyQueryHookResult = ReturnType<typeof useToolcallsStatsTotalLazyQuery>;
export type ToolcallsStatsTotalSuspenseQueryHookResult = ReturnType<typeof useToolcallsStatsTotalSuspenseQuery>;
export type ToolcallsStatsTotalQueryResult = ApolloReactCommon.QueryResult<
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
    baseOptions: ApolloReactHooks.QueryHookOptions<ToolcallsStatsByPeriodQuery, ToolcallsStatsByPeriodQueryVariables> &
        ({ variables: ToolcallsStatsByPeriodQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<ToolcallsStatsByPeriodQuery, ToolcallsStatsByPeriodQueryVariables>(
        ToolcallsStatsByPeriodDocument,
        options,
    );
}
export function useToolcallsStatsByPeriodLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<
        ToolcallsStatsByPeriodQuery,
        ToolcallsStatsByPeriodQueryVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<ToolcallsStatsByPeriodQuery, ToolcallsStatsByPeriodQueryVariables>(
        ToolcallsStatsByPeriodDocument,
        options,
    );
}
// @ts-ignore
export function useToolcallsStatsByPeriodSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<
        ToolcallsStatsByPeriodQuery,
        ToolcallsStatsByPeriodQueryVariables
    >,
): ApolloReactHooks.UseSuspenseQueryResult<ToolcallsStatsByPeriodQuery, ToolcallsStatsByPeriodQueryVariables>;
export function useToolcallsStatsByPeriodSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<ToolcallsStatsByPeriodQuery, ToolcallsStatsByPeriodQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<
    ToolcallsStatsByPeriodQuery | undefined,
    ToolcallsStatsByPeriodQueryVariables
>;
export function useToolcallsStatsByPeriodSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<ToolcallsStatsByPeriodQuery, ToolcallsStatsByPeriodQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<ToolcallsStatsByPeriodQuery, ToolcallsStatsByPeriodQueryVariables>(
        ToolcallsStatsByPeriodDocument,
        options,
    );
}
export type ToolcallsStatsByPeriodQueryHookResult = ReturnType<typeof useToolcallsStatsByPeriodQuery>;
export type ToolcallsStatsByPeriodLazyQueryHookResult = ReturnType<typeof useToolcallsStatsByPeriodLazyQuery>;
export type ToolcallsStatsByPeriodSuspenseQueryHookResult = ReturnType<typeof useToolcallsStatsByPeriodSuspenseQuery>;
export type ToolcallsStatsByPeriodQueryResult = ApolloReactCommon.QueryResult<
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
    baseOptions?: ApolloReactHooks.QueryHookOptions<
        ToolcallsStatsByFunctionQuery,
        ToolcallsStatsByFunctionQueryVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<ToolcallsStatsByFunctionQuery, ToolcallsStatsByFunctionQueryVariables>(
        ToolcallsStatsByFunctionDocument,
        options,
    );
}
export function useToolcallsStatsByFunctionLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<
        ToolcallsStatsByFunctionQuery,
        ToolcallsStatsByFunctionQueryVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<ToolcallsStatsByFunctionQuery, ToolcallsStatsByFunctionQueryVariables>(
        ToolcallsStatsByFunctionDocument,
        options,
    );
}
// @ts-ignore
export function useToolcallsStatsByFunctionSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<
        ToolcallsStatsByFunctionQuery,
        ToolcallsStatsByFunctionQueryVariables
    >,
): ApolloReactHooks.UseSuspenseQueryResult<ToolcallsStatsByFunctionQuery, ToolcallsStatsByFunctionQueryVariables>;
export function useToolcallsStatsByFunctionSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<
              ToolcallsStatsByFunctionQuery,
              ToolcallsStatsByFunctionQueryVariables
          >,
): ApolloReactHooks.UseSuspenseQueryResult<
    ToolcallsStatsByFunctionQuery | undefined,
    ToolcallsStatsByFunctionQueryVariables
>;
export function useToolcallsStatsByFunctionSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<
              ToolcallsStatsByFunctionQuery,
              ToolcallsStatsByFunctionQueryVariables
          >,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<ToolcallsStatsByFunctionQuery, ToolcallsStatsByFunctionQueryVariables>(
        ToolcallsStatsByFunctionDocument,
        options,
    );
}
export type ToolcallsStatsByFunctionQueryHookResult = ReturnType<typeof useToolcallsStatsByFunctionQuery>;
export type ToolcallsStatsByFunctionLazyQueryHookResult = ReturnType<typeof useToolcallsStatsByFunctionLazyQuery>;
export type ToolcallsStatsByFunctionSuspenseQueryHookResult = ReturnType<
    typeof useToolcallsStatsByFunctionSuspenseQuery
>;
export type ToolcallsStatsByFunctionQueryResult = ApolloReactCommon.QueryResult<
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
    baseOptions: ApolloReactHooks.QueryHookOptions<ToolcallsStatsByFlowQuery, ToolcallsStatsByFlowQueryVariables> &
        ({ variables: ToolcallsStatsByFlowQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<ToolcallsStatsByFlowQuery, ToolcallsStatsByFlowQueryVariables>(
        ToolcallsStatsByFlowDocument,
        options,
    );
}
export function useToolcallsStatsByFlowLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<ToolcallsStatsByFlowQuery, ToolcallsStatsByFlowQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<ToolcallsStatsByFlowQuery, ToolcallsStatsByFlowQueryVariables>(
        ToolcallsStatsByFlowDocument,
        options,
    );
}
// @ts-ignore
export function useToolcallsStatsByFlowSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<
        ToolcallsStatsByFlowQuery,
        ToolcallsStatsByFlowQueryVariables
    >,
): ApolloReactHooks.UseSuspenseQueryResult<ToolcallsStatsByFlowQuery, ToolcallsStatsByFlowQueryVariables>;
export function useToolcallsStatsByFlowSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<ToolcallsStatsByFlowQuery, ToolcallsStatsByFlowQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<ToolcallsStatsByFlowQuery | undefined, ToolcallsStatsByFlowQueryVariables>;
export function useToolcallsStatsByFlowSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<ToolcallsStatsByFlowQuery, ToolcallsStatsByFlowQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<ToolcallsStatsByFlowQuery, ToolcallsStatsByFlowQueryVariables>(
        ToolcallsStatsByFlowDocument,
        options,
    );
}
export type ToolcallsStatsByFlowQueryHookResult = ReturnType<typeof useToolcallsStatsByFlowQuery>;
export type ToolcallsStatsByFlowLazyQueryHookResult = ReturnType<typeof useToolcallsStatsByFlowLazyQuery>;
export type ToolcallsStatsByFlowSuspenseQueryHookResult = ReturnType<typeof useToolcallsStatsByFlowSuspenseQuery>;
export type ToolcallsStatsByFlowQueryResult = ApolloReactCommon.QueryResult<
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
    baseOptions: ApolloReactHooks.QueryHookOptions<
        ToolcallsStatsByFunctionForFlowQuery,
        ToolcallsStatsByFunctionForFlowQueryVariables
    > &
        ({ variables: ToolcallsStatsByFunctionForFlowQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<
        ToolcallsStatsByFunctionForFlowQuery,
        ToolcallsStatsByFunctionForFlowQueryVariables
    >(ToolcallsStatsByFunctionForFlowDocument, options);
}
export function useToolcallsStatsByFunctionForFlowLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<
        ToolcallsStatsByFunctionForFlowQuery,
        ToolcallsStatsByFunctionForFlowQueryVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<
        ToolcallsStatsByFunctionForFlowQuery,
        ToolcallsStatsByFunctionForFlowQueryVariables
    >(ToolcallsStatsByFunctionForFlowDocument, options);
}
// @ts-ignore
export function useToolcallsStatsByFunctionForFlowSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<
        ToolcallsStatsByFunctionForFlowQuery,
        ToolcallsStatsByFunctionForFlowQueryVariables
    >,
): ApolloReactHooks.UseSuspenseQueryResult<
    ToolcallsStatsByFunctionForFlowQuery,
    ToolcallsStatsByFunctionForFlowQueryVariables
>;
export function useToolcallsStatsByFunctionForFlowSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<
              ToolcallsStatsByFunctionForFlowQuery,
              ToolcallsStatsByFunctionForFlowQueryVariables
          >,
): ApolloReactHooks.UseSuspenseQueryResult<
    ToolcallsStatsByFunctionForFlowQuery | undefined,
    ToolcallsStatsByFunctionForFlowQueryVariables
>;
export function useToolcallsStatsByFunctionForFlowSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<
              ToolcallsStatsByFunctionForFlowQuery,
              ToolcallsStatsByFunctionForFlowQueryVariables
          >,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<
        ToolcallsStatsByFunctionForFlowQuery,
        ToolcallsStatsByFunctionForFlowQueryVariables
    >(ToolcallsStatsByFunctionForFlowDocument, options);
}
export type ToolcallsStatsByFunctionForFlowQueryHookResult = ReturnType<typeof useToolcallsStatsByFunctionForFlowQuery>;
export type ToolcallsStatsByFunctionForFlowLazyQueryHookResult = ReturnType<
    typeof useToolcallsStatsByFunctionForFlowLazyQuery
>;
export type ToolcallsStatsByFunctionForFlowSuspenseQueryHookResult = ReturnType<
    typeof useToolcallsStatsByFunctionForFlowSuspenseQuery
>;
export type ToolcallsStatsByFunctionForFlowQueryResult = ApolloReactCommon.QueryResult<
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
    baseOptions?: ApolloReactHooks.QueryHookOptions<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>(
        FlowsStatsTotalDocument,
        options,
    );
}
export function useFlowsStatsTotalLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>(
        FlowsStatsTotalDocument,
        options,
    );
}
// @ts-ignore
export function useFlowsStatsTotalSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>;
export function useFlowsStatsTotalSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowsStatsTotalQuery | undefined, FlowsStatsTotalQueryVariables>;
export function useFlowsStatsTotalSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>(
        FlowsStatsTotalDocument,
        options,
    );
}
export type FlowsStatsTotalQueryHookResult = ReturnType<typeof useFlowsStatsTotalQuery>;
export type FlowsStatsTotalLazyQueryHookResult = ReturnType<typeof useFlowsStatsTotalLazyQuery>;
export type FlowsStatsTotalSuspenseQueryHookResult = ReturnType<typeof useFlowsStatsTotalSuspenseQuery>;
export type FlowsStatsTotalQueryResult = ApolloReactCommon.QueryResult<
    FlowsStatsTotalQuery,
    FlowsStatsTotalQueryVariables
>;
export const FlowsStatsByPeriodDocument = gql`
    query flowsStatsByPeriod($period: UsageStatsPeriod!) {
        flowsStatsByPeriod(period: $period) {
            ...dailyFlowsStatsFragment
        }
    }
    ${DailyFlowsStatsFragmentFragmentDoc}
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
    baseOptions: ApolloReactHooks.QueryHookOptions<FlowsStatsByPeriodQuery, FlowsStatsByPeriodQueryVariables> &
        ({ variables: FlowsStatsByPeriodQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<FlowsStatsByPeriodQuery, FlowsStatsByPeriodQueryVariables>(
        FlowsStatsByPeriodDocument,
        options,
    );
}
export function useFlowsStatsByPeriodLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<FlowsStatsByPeriodQuery, FlowsStatsByPeriodQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<FlowsStatsByPeriodQuery, FlowsStatsByPeriodQueryVariables>(
        FlowsStatsByPeriodDocument,
        options,
    );
}
// @ts-ignore
export function useFlowsStatsByPeriodSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<FlowsStatsByPeriodQuery, FlowsStatsByPeriodQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowsStatsByPeriodQuery, FlowsStatsByPeriodQueryVariables>;
export function useFlowsStatsByPeriodSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<FlowsStatsByPeriodQuery, FlowsStatsByPeriodQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowsStatsByPeriodQuery | undefined, FlowsStatsByPeriodQueryVariables>;
export function useFlowsStatsByPeriodSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<FlowsStatsByPeriodQuery, FlowsStatsByPeriodQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<FlowsStatsByPeriodQuery, FlowsStatsByPeriodQueryVariables>(
        FlowsStatsByPeriodDocument,
        options,
    );
}
export type FlowsStatsByPeriodQueryHookResult = ReturnType<typeof useFlowsStatsByPeriodQuery>;
export type FlowsStatsByPeriodLazyQueryHookResult = ReturnType<typeof useFlowsStatsByPeriodLazyQuery>;
export type FlowsStatsByPeriodSuspenseQueryHookResult = ReturnType<typeof useFlowsStatsByPeriodSuspenseQuery>;
export type FlowsStatsByPeriodQueryResult = ApolloReactCommon.QueryResult<
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
    baseOptions: ApolloReactHooks.QueryHookOptions<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables> &
        ({ variables: FlowStatsByFlowQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables>(
        FlowStatsByFlowDocument,
        options,
    );
}
export function useFlowStatsByFlowLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables>(
        FlowStatsByFlowDocument,
        options,
    );
}
// @ts-ignore
export function useFlowStatsByFlowSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables>;
export function useFlowStatsByFlowSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowStatsByFlowQuery | undefined, FlowStatsByFlowQueryVariables>;
export function useFlowStatsByFlowSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables>(
        FlowStatsByFlowDocument,
        options,
    );
}
export type FlowStatsByFlowQueryHookResult = ReturnType<typeof useFlowStatsByFlowQuery>;
export type FlowStatsByFlowLazyQueryHookResult = ReturnType<typeof useFlowStatsByFlowLazyQuery>;
export type FlowStatsByFlowSuspenseQueryHookResult = ReturnType<typeof useFlowStatsByFlowSuspenseQuery>;
export type FlowStatsByFlowQueryResult = ApolloReactCommon.QueryResult<
    FlowStatsByFlowQuery,
    FlowStatsByFlowQueryVariables
>;
export const FlowsExecutionStatsByPeriodDocument = gql`
    query flowsExecutionStatsByPeriod($period: UsageStatsPeriod!) {
        flowsExecutionStatsByPeriod(period: $period) {
            ...flowExecutionStatsFragment
        }
    }
    ${FlowExecutionStatsFragmentFragmentDoc}
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
    baseOptions: ApolloReactHooks.QueryHookOptions<
        FlowsExecutionStatsByPeriodQuery,
        FlowsExecutionStatsByPeriodQueryVariables
    > &
        ({ variables: FlowsExecutionStatsByPeriodQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<FlowsExecutionStatsByPeriodQuery, FlowsExecutionStatsByPeriodQueryVariables>(
        FlowsExecutionStatsByPeriodDocument,
        options,
    );
}
export function useFlowsExecutionStatsByPeriodLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<
        FlowsExecutionStatsByPeriodQuery,
        FlowsExecutionStatsByPeriodQueryVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<FlowsExecutionStatsByPeriodQuery, FlowsExecutionStatsByPeriodQueryVariables>(
        FlowsExecutionStatsByPeriodDocument,
        options,
    );
}
// @ts-ignore
export function useFlowsExecutionStatsByPeriodSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<
        FlowsExecutionStatsByPeriodQuery,
        FlowsExecutionStatsByPeriodQueryVariables
    >,
): ApolloReactHooks.UseSuspenseQueryResult<FlowsExecutionStatsByPeriodQuery, FlowsExecutionStatsByPeriodQueryVariables>;
export function useFlowsExecutionStatsByPeriodSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<
              FlowsExecutionStatsByPeriodQuery,
              FlowsExecutionStatsByPeriodQueryVariables
          >,
): ApolloReactHooks.UseSuspenseQueryResult<
    FlowsExecutionStatsByPeriodQuery | undefined,
    FlowsExecutionStatsByPeriodQueryVariables
>;
export function useFlowsExecutionStatsByPeriodSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<
              FlowsExecutionStatsByPeriodQuery,
              FlowsExecutionStatsByPeriodQueryVariables
          >,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<
        FlowsExecutionStatsByPeriodQuery,
        FlowsExecutionStatsByPeriodQueryVariables
    >(FlowsExecutionStatsByPeriodDocument, options);
}
export type FlowsExecutionStatsByPeriodQueryHookResult = ReturnType<typeof useFlowsExecutionStatsByPeriodQuery>;
export type FlowsExecutionStatsByPeriodLazyQueryHookResult = ReturnType<typeof useFlowsExecutionStatsByPeriodLazyQuery>;
export type FlowsExecutionStatsByPeriodSuspenseQueryHookResult = ReturnType<
    typeof useFlowsExecutionStatsByPeriodSuspenseQuery
>;
export type FlowsExecutionStatsByPeriodQueryResult = ApolloReactCommon.QueryResult<
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
export function useApiTokensQuery(
    baseOptions?: ApolloReactHooks.QueryHookOptions<ApiTokensQuery, ApiTokensQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<ApiTokensQuery, ApiTokensQueryVariables>(ApiTokensDocument, options);
}
export function useApiTokensLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<ApiTokensQuery, ApiTokensQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<ApiTokensQuery, ApiTokensQueryVariables>(ApiTokensDocument, options);
}
// @ts-ignore
export function useApiTokensSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<ApiTokensQuery, ApiTokensQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<ApiTokensQuery, ApiTokensQueryVariables>;
export function useApiTokensSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<ApiTokensQuery, ApiTokensQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<ApiTokensQuery | undefined, ApiTokensQueryVariables>;
export function useApiTokensSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<ApiTokensQuery, ApiTokensQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<ApiTokensQuery, ApiTokensQueryVariables>(ApiTokensDocument, options);
}
export type ApiTokensQueryHookResult = ReturnType<typeof useApiTokensQuery>;
export type ApiTokensLazyQueryHookResult = ReturnType<typeof useApiTokensLazyQuery>;
export type ApiTokensSuspenseQueryHookResult = ReturnType<typeof useApiTokensSuspenseQuery>;
export type ApiTokensQueryResult = ApolloReactCommon.QueryResult<ApiTokensQuery, ApiTokensQueryVariables>;
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
    baseOptions: ApolloReactHooks.QueryHookOptions<ApiTokenQuery, ApiTokenQueryVariables> &
        ({ variables: ApiTokenQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<ApiTokenQuery, ApiTokenQueryVariables>(ApiTokenDocument, options);
}
export function useApiTokenLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<ApiTokenQuery, ApiTokenQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<ApiTokenQuery, ApiTokenQueryVariables>(ApiTokenDocument, options);
}
// @ts-ignore
export function useApiTokenSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<ApiTokenQuery, ApiTokenQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<ApiTokenQuery, ApiTokenQueryVariables>;
export function useApiTokenSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<ApiTokenQuery, ApiTokenQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<ApiTokenQuery | undefined, ApiTokenQueryVariables>;
export function useApiTokenSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<ApiTokenQuery, ApiTokenQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<ApiTokenQuery, ApiTokenQueryVariables>(ApiTokenDocument, options);
}
export type ApiTokenQueryHookResult = ReturnType<typeof useApiTokenQuery>;
export type ApiTokenLazyQueryHookResult = ReturnType<typeof useApiTokenLazyQuery>;
export type ApiTokenSuspenseQueryHookResult = ReturnType<typeof useApiTokenSuspenseQuery>;
export type ApiTokenQueryResult = ApolloReactCommon.QueryResult<ApiTokenQuery, ApiTokenQueryVariables>;
export const KnowledgeDocumentsDocument = gql`
    query knowledgeDocuments($filter: KnowledgeFilter, $withContent: Boolean!) {
        knowledgeDocuments(filter: $filter, withContent: $withContent) {
            ...knowledgeDocumentFragment
        }
    }
    ${KnowledgeDocumentFragmentFragmentDoc}
`;

/**
 * __useKnowledgeDocumentsQuery__
 *
 * To run a query within a React component, call `useKnowledgeDocumentsQuery` and pass it any options that fit your needs.
 * When your component renders, `useKnowledgeDocumentsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useKnowledgeDocumentsQuery({
 *   variables: {
 *      filter: // value for 'filter'
 *      withContent: // value for 'withContent'
 *   },
 * });
 */
export function useKnowledgeDocumentsQuery(
    baseOptions: ApolloReactHooks.QueryHookOptions<KnowledgeDocumentsQuery, KnowledgeDocumentsQueryVariables> &
        ({ variables: KnowledgeDocumentsQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<KnowledgeDocumentsQuery, KnowledgeDocumentsQueryVariables>(
        KnowledgeDocumentsDocument,
        options,
    );
}
export function useKnowledgeDocumentsLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<KnowledgeDocumentsQuery, KnowledgeDocumentsQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<KnowledgeDocumentsQuery, KnowledgeDocumentsQueryVariables>(
        KnowledgeDocumentsDocument,
        options,
    );
}
// @ts-ignore
export function useKnowledgeDocumentsSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<KnowledgeDocumentsQuery, KnowledgeDocumentsQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<KnowledgeDocumentsQuery, KnowledgeDocumentsQueryVariables>;
export function useKnowledgeDocumentsSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<KnowledgeDocumentsQuery, KnowledgeDocumentsQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<KnowledgeDocumentsQuery | undefined, KnowledgeDocumentsQueryVariables>;
export function useKnowledgeDocumentsSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<KnowledgeDocumentsQuery, KnowledgeDocumentsQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<KnowledgeDocumentsQuery, KnowledgeDocumentsQueryVariables>(
        KnowledgeDocumentsDocument,
        options,
    );
}
export type KnowledgeDocumentsQueryHookResult = ReturnType<typeof useKnowledgeDocumentsQuery>;
export type KnowledgeDocumentsLazyQueryHookResult = ReturnType<typeof useKnowledgeDocumentsLazyQuery>;
export type KnowledgeDocumentsSuspenseQueryHookResult = ReturnType<typeof useKnowledgeDocumentsSuspenseQuery>;
export type KnowledgeDocumentsQueryResult = ApolloReactCommon.QueryResult<
    KnowledgeDocumentsQuery,
    KnowledgeDocumentsQueryVariables
>;
export const KnowledgeDocumentDocument = gql`
    query knowledgeDocument($id: String!) {
        knowledgeDocument(id: $id) {
            ...knowledgeDocumentFragment
        }
    }
    ${KnowledgeDocumentFragmentFragmentDoc}
`;

/**
 * __useKnowledgeDocumentQuery__
 *
 * To run a query within a React component, call `useKnowledgeDocumentQuery` and pass it any options that fit your needs.
 * When your component renders, `useKnowledgeDocumentQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useKnowledgeDocumentQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useKnowledgeDocumentQuery(
    baseOptions: ApolloReactHooks.QueryHookOptions<KnowledgeDocumentQuery, KnowledgeDocumentQueryVariables> &
        ({ variables: KnowledgeDocumentQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<KnowledgeDocumentQuery, KnowledgeDocumentQueryVariables>(
        KnowledgeDocumentDocument,
        options,
    );
}
export function useKnowledgeDocumentLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<KnowledgeDocumentQuery, KnowledgeDocumentQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<KnowledgeDocumentQuery, KnowledgeDocumentQueryVariables>(
        KnowledgeDocumentDocument,
        options,
    );
}
// @ts-ignore
export function useKnowledgeDocumentSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<KnowledgeDocumentQuery, KnowledgeDocumentQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<KnowledgeDocumentQuery, KnowledgeDocumentQueryVariables>;
export function useKnowledgeDocumentSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<KnowledgeDocumentQuery, KnowledgeDocumentQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<KnowledgeDocumentQuery | undefined, KnowledgeDocumentQueryVariables>;
export function useKnowledgeDocumentSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<KnowledgeDocumentQuery, KnowledgeDocumentQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<KnowledgeDocumentQuery, KnowledgeDocumentQueryVariables>(
        KnowledgeDocumentDocument,
        options,
    );
}
export type KnowledgeDocumentQueryHookResult = ReturnType<typeof useKnowledgeDocumentQuery>;
export type KnowledgeDocumentLazyQueryHookResult = ReturnType<typeof useKnowledgeDocumentLazyQuery>;
export type KnowledgeDocumentSuspenseQueryHookResult = ReturnType<typeof useKnowledgeDocumentSuspenseQuery>;
export type KnowledgeDocumentQueryResult = ApolloReactCommon.QueryResult<
    KnowledgeDocumentQuery,
    KnowledgeDocumentQueryVariables
>;
export const SearchKnowledgeDocument = gql`
    query searchKnowledge($query: String!, $filter: KnowledgeFilter, $limit: Int) {
        searchKnowledge(query: $query, filter: $filter, limit: $limit) {
            ...knowledgeDocumentWithScoreFragment
        }
    }
    ${KnowledgeDocumentWithScoreFragmentFragmentDoc}
`;

/**
 * __useSearchKnowledgeQuery__
 *
 * To run a query within a React component, call `useSearchKnowledgeQuery` and pass it any options that fit your needs.
 * When your component renders, `useSearchKnowledgeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSearchKnowledgeQuery({
 *   variables: {
 *      query: // value for 'query'
 *      filter: // value for 'filter'
 *      limit: // value for 'limit'
 *   },
 * });
 */
export function useSearchKnowledgeQuery(
    baseOptions: ApolloReactHooks.QueryHookOptions<SearchKnowledgeQuery, SearchKnowledgeQueryVariables> &
        ({ variables: SearchKnowledgeQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<SearchKnowledgeQuery, SearchKnowledgeQueryVariables>(
        SearchKnowledgeDocument,
        options,
    );
}
export function useSearchKnowledgeLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<SearchKnowledgeQuery, SearchKnowledgeQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<SearchKnowledgeQuery, SearchKnowledgeQueryVariables>(
        SearchKnowledgeDocument,
        options,
    );
}
// @ts-ignore
export function useSearchKnowledgeSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<SearchKnowledgeQuery, SearchKnowledgeQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<SearchKnowledgeQuery, SearchKnowledgeQueryVariables>;
export function useSearchKnowledgeSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<SearchKnowledgeQuery, SearchKnowledgeQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<SearchKnowledgeQuery | undefined, SearchKnowledgeQueryVariables>;
export function useSearchKnowledgeSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<SearchKnowledgeQuery, SearchKnowledgeQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<SearchKnowledgeQuery, SearchKnowledgeQueryVariables>(
        SearchKnowledgeDocument,
        options,
    );
}
export type SearchKnowledgeQueryHookResult = ReturnType<typeof useSearchKnowledgeQuery>;
export type SearchKnowledgeLazyQueryHookResult = ReturnType<typeof useSearchKnowledgeLazyQuery>;
export type SearchKnowledgeSuspenseQueryHookResult = ReturnType<typeof useSearchKnowledgeSuspenseQuery>;
export type SearchKnowledgeQueryResult = ApolloReactCommon.QueryResult<
    SearchKnowledgeQuery,
    SearchKnowledgeQueryVariables
>;
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
    baseOptions?: ApolloReactHooks.QueryHookOptions<SettingsUserQuery, SettingsUserQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<SettingsUserQuery, SettingsUserQueryVariables>(SettingsUserDocument, options);
}
export function useSettingsUserLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<SettingsUserQuery, SettingsUserQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<SettingsUserQuery, SettingsUserQueryVariables>(SettingsUserDocument, options);
}
// @ts-ignore
export function useSettingsUserSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<SettingsUserQuery, SettingsUserQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<SettingsUserQuery, SettingsUserQueryVariables>;
export function useSettingsUserSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<SettingsUserQuery, SettingsUserQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<SettingsUserQuery | undefined, SettingsUserQueryVariables>;
export function useSettingsUserSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<SettingsUserQuery, SettingsUserQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<SettingsUserQuery, SettingsUserQueryVariables>(
        SettingsUserDocument,
        options,
    );
}
export type SettingsUserQueryHookResult = ReturnType<typeof useSettingsUserQuery>;
export type SettingsUserLazyQueryHookResult = ReturnType<typeof useSettingsUserLazyQuery>;
export type SettingsUserSuspenseQueryHookResult = ReturnType<typeof useSettingsUserSuspenseQuery>;
export type SettingsUserQueryResult = ApolloReactCommon.QueryResult<SettingsUserQuery, SettingsUserQueryVariables>;
export const AddFavoriteFlowDocument = gql`
    mutation addFavoriteFlow($flowId: ID!) {
        addFavoriteFlow(flowId: $flowId)
    }
`;
export type AddFavoriteFlowMutationFn = ApolloReactCommon.MutationFunction<
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
    baseOptions?: ApolloReactHooks.MutationHookOptions<AddFavoriteFlowMutation, AddFavoriteFlowMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<AddFavoriteFlowMutation, AddFavoriteFlowMutationVariables>(
        AddFavoriteFlowDocument,
        options,
    );
}
export type AddFavoriteFlowMutationHookResult = ReturnType<typeof useAddFavoriteFlowMutation>;
export type AddFavoriteFlowMutationResult = ApolloReactCommon.MutationResult<AddFavoriteFlowMutation>;
export type AddFavoriteFlowMutationOptions = ApolloReactCommon.BaseMutationOptions<
    AddFavoriteFlowMutation,
    AddFavoriteFlowMutationVariables
>;
export const DeleteFavoriteFlowDocument = gql`
    mutation deleteFavoriteFlow($flowId: ID!) {
        deleteFavoriteFlow(flowId: $flowId)
    }
`;
export type DeleteFavoriteFlowMutationFn = ApolloReactCommon.MutationFunction<
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
    baseOptions?: ApolloReactHooks.MutationHookOptions<DeleteFavoriteFlowMutation, DeleteFavoriteFlowMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<DeleteFavoriteFlowMutation, DeleteFavoriteFlowMutationVariables>(
        DeleteFavoriteFlowDocument,
        options,
    );
}
export type DeleteFavoriteFlowMutationHookResult = ReturnType<typeof useDeleteFavoriteFlowMutation>;
export type DeleteFavoriteFlowMutationResult = ApolloReactCommon.MutationResult<DeleteFavoriteFlowMutation>;
export type DeleteFavoriteFlowMutationOptions = ApolloReactCommon.BaseMutationOptions<
    DeleteFavoriteFlowMutation,
    DeleteFavoriteFlowMutationVariables
>;
export const AnonymizeTextDocument = gql`
    mutation anonymizeText($text: String!) {
        anonymizeText(text: $text)
    }
`;
export type AnonymizeTextMutationFn = ApolloReactCommon.MutationFunction<
    AnonymizeTextMutation,
    AnonymizeTextMutationVariables
>;

/**
 * __useAnonymizeTextMutation__
 *
 * To run a mutation, you first call `useAnonymizeTextMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAnonymizeTextMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [anonymizeTextMutation, { data, loading, error }] = useAnonymizeTextMutation({
 *   variables: {
 *      text: // value for 'text'
 *   },
 * });
 */
export function useAnonymizeTextMutation(
    baseOptions?: ApolloReactHooks.MutationHookOptions<AnonymizeTextMutation, AnonymizeTextMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<AnonymizeTextMutation, AnonymizeTextMutationVariables>(
        AnonymizeTextDocument,
        options,
    );
}
export type AnonymizeTextMutationHookResult = ReturnType<typeof useAnonymizeTextMutation>;
export type AnonymizeTextMutationResult = ApolloReactCommon.MutationResult<AnonymizeTextMutation>;
export type AnonymizeTextMutationOptions = ApolloReactCommon.BaseMutationOptions<
    AnonymizeTextMutation,
    AnonymizeTextMutationVariables
>;
export const FlowTemplatesDocument = gql`
    query flowTemplates {
        flowTemplates {
            ...flowTemplateFragment
        }
    }
    ${FlowTemplateFragmentFragmentDoc}
`;

/**
 * __useFlowTemplatesQuery__
 *
 * To run a query within a React component, call `useFlowTemplatesQuery` and pass it any options that fit your needs.
 * When your component renders, `useFlowTemplatesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowTemplatesQuery({
 *   variables: {
 *   },
 * });
 */
export function useFlowTemplatesQuery(
    baseOptions?: ApolloReactHooks.QueryHookOptions<FlowTemplatesQuery, FlowTemplatesQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<FlowTemplatesQuery, FlowTemplatesQueryVariables>(FlowTemplatesDocument, options);
}
export function useFlowTemplatesLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<FlowTemplatesQuery, FlowTemplatesQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<FlowTemplatesQuery, FlowTemplatesQueryVariables>(
        FlowTemplatesDocument,
        options,
    );
}
// @ts-ignore
export function useFlowTemplatesSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<FlowTemplatesQuery, FlowTemplatesQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowTemplatesQuery, FlowTemplatesQueryVariables>;
export function useFlowTemplatesSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<FlowTemplatesQuery, FlowTemplatesQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowTemplatesQuery | undefined, FlowTemplatesQueryVariables>;
export function useFlowTemplatesSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<FlowTemplatesQuery, FlowTemplatesQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<FlowTemplatesQuery, FlowTemplatesQueryVariables>(
        FlowTemplatesDocument,
        options,
    );
}
export type FlowTemplatesQueryHookResult = ReturnType<typeof useFlowTemplatesQuery>;
export type FlowTemplatesLazyQueryHookResult = ReturnType<typeof useFlowTemplatesLazyQuery>;
export type FlowTemplatesSuspenseQueryHookResult = ReturnType<typeof useFlowTemplatesSuspenseQuery>;
export type FlowTemplatesQueryResult = ApolloReactCommon.QueryResult<FlowTemplatesQuery, FlowTemplatesQueryVariables>;
export const FlowTemplateDocument = gql`
    query flowTemplate($templateId: ID!) {
        flowTemplate(templateId: $templateId) {
            ...flowTemplateFragment
        }
    }
    ${FlowTemplateFragmentFragmentDoc}
`;

/**
 * __useFlowTemplateQuery__
 *
 * To run a query within a React component, call `useFlowTemplateQuery` and pass it any options that fit your needs.
 * When your component renders, `useFlowTemplateQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowTemplateQuery({
 *   variables: {
 *      templateId: // value for 'templateId'
 *   },
 * });
 */
export function useFlowTemplateQuery(
    baseOptions: ApolloReactHooks.QueryHookOptions<FlowTemplateQuery, FlowTemplateQueryVariables> &
        ({ variables: FlowTemplateQueryVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useQuery<FlowTemplateQuery, FlowTemplateQueryVariables>(FlowTemplateDocument, options);
}
export function useFlowTemplateLazyQuery(
    baseOptions?: ApolloReactHooks.LazyQueryHookOptions<FlowTemplateQuery, FlowTemplateQueryVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useLazyQuery<FlowTemplateQuery, FlowTemplateQueryVariables>(FlowTemplateDocument, options);
}
// @ts-ignore
export function useFlowTemplateSuspenseQuery(
    baseOptions?: ApolloReactHooks.SuspenseQueryHookOptions<FlowTemplateQuery, FlowTemplateQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowTemplateQuery, FlowTemplateQueryVariables>;
export function useFlowTemplateSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<FlowTemplateQuery, FlowTemplateQueryVariables>,
): ApolloReactHooks.UseSuspenseQueryResult<FlowTemplateQuery | undefined, FlowTemplateQueryVariables>;
export function useFlowTemplateSuspenseQuery(
    baseOptions?:
        | ApolloReactHooks.SkipToken
        | ApolloReactHooks.SuspenseQueryHookOptions<FlowTemplateQuery, FlowTemplateQueryVariables>,
) {
    const options = baseOptions === ApolloReactHooks.skipToken ? baseOptions : { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSuspenseQuery<FlowTemplateQuery, FlowTemplateQueryVariables>(
        FlowTemplateDocument,
        options,
    );
}
export type FlowTemplateQueryHookResult = ReturnType<typeof useFlowTemplateQuery>;
export type FlowTemplateLazyQueryHookResult = ReturnType<typeof useFlowTemplateLazyQuery>;
export type FlowTemplateSuspenseQueryHookResult = ReturnType<typeof useFlowTemplateSuspenseQuery>;
export type FlowTemplateQueryResult = ApolloReactCommon.QueryResult<FlowTemplateQuery, FlowTemplateQueryVariables>;
export const CreateFlowTemplateDocument = gql`
    mutation createFlowTemplate($input: CreateFlowTemplateInput!) {
        createFlowTemplate(input: $input) {
            ...flowTemplateFragment
        }
    }
    ${FlowTemplateFragmentFragmentDoc}
`;
export type CreateFlowTemplateMutationFn = ApolloReactCommon.MutationFunction<
    CreateFlowTemplateMutation,
    CreateFlowTemplateMutationVariables
>;

/**
 * __useCreateFlowTemplateMutation__
 *
 * To run a mutation, you first call `useCreateFlowTemplateMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateFlowTemplateMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createFlowTemplateMutation, { data, loading, error }] = useCreateFlowTemplateMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateFlowTemplateMutation(
    baseOptions?: ApolloReactHooks.MutationHookOptions<CreateFlowTemplateMutation, CreateFlowTemplateMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<CreateFlowTemplateMutation, CreateFlowTemplateMutationVariables>(
        CreateFlowTemplateDocument,
        options,
    );
}
export type CreateFlowTemplateMutationHookResult = ReturnType<typeof useCreateFlowTemplateMutation>;
export type CreateFlowTemplateMutationResult = ApolloReactCommon.MutationResult<CreateFlowTemplateMutation>;
export type CreateFlowTemplateMutationOptions = ApolloReactCommon.BaseMutationOptions<
    CreateFlowTemplateMutation,
    CreateFlowTemplateMutationVariables
>;
export const UpdateFlowTemplateDocument = gql`
    mutation updateFlowTemplate($templateId: ID!, $input: UpdateFlowTemplateInput!) {
        updateFlowTemplate(templateId: $templateId, input: $input) {
            ...flowTemplateFragment
        }
    }
    ${FlowTemplateFragmentFragmentDoc}
`;
export type UpdateFlowTemplateMutationFn = ApolloReactCommon.MutationFunction<
    UpdateFlowTemplateMutation,
    UpdateFlowTemplateMutationVariables
>;

/**
 * __useUpdateFlowTemplateMutation__
 *
 * To run a mutation, you first call `useUpdateFlowTemplateMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateFlowTemplateMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateFlowTemplateMutation, { data, loading, error }] = useUpdateFlowTemplateMutation({
 *   variables: {
 *      templateId: // value for 'templateId'
 *      input: // value for 'input'
 *   },
 * });
 */
export function useUpdateFlowTemplateMutation(
    baseOptions?: ApolloReactHooks.MutationHookOptions<UpdateFlowTemplateMutation, UpdateFlowTemplateMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<UpdateFlowTemplateMutation, UpdateFlowTemplateMutationVariables>(
        UpdateFlowTemplateDocument,
        options,
    );
}
export type UpdateFlowTemplateMutationHookResult = ReturnType<typeof useUpdateFlowTemplateMutation>;
export type UpdateFlowTemplateMutationResult = ApolloReactCommon.MutationResult<UpdateFlowTemplateMutation>;
export type UpdateFlowTemplateMutationOptions = ApolloReactCommon.BaseMutationOptions<
    UpdateFlowTemplateMutation,
    UpdateFlowTemplateMutationVariables
>;
export const DeleteFlowTemplateDocument = gql`
    mutation deleteFlowTemplate($templateId: ID!) {
        deleteFlowTemplate(templateId: $templateId)
    }
`;
export type DeleteFlowTemplateMutationFn = ApolloReactCommon.MutationFunction<
    DeleteFlowTemplateMutation,
    DeleteFlowTemplateMutationVariables
>;

/**
 * __useDeleteFlowTemplateMutation__
 *
 * To run a mutation, you first call `useDeleteFlowTemplateMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteFlowTemplateMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteFlowTemplateMutation, { data, loading, error }] = useDeleteFlowTemplateMutation({
 *   variables: {
 *      templateId: // value for 'templateId'
 *   },
 * });
 */
export function useDeleteFlowTemplateMutation(
    baseOptions?: ApolloReactHooks.MutationHookOptions<DeleteFlowTemplateMutation, DeleteFlowTemplateMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<DeleteFlowTemplateMutation, DeleteFlowTemplateMutationVariables>(
        DeleteFlowTemplateDocument,
        options,
    );
}
export type DeleteFlowTemplateMutationHookResult = ReturnType<typeof useDeleteFlowTemplateMutation>;
export type DeleteFlowTemplateMutationResult = ApolloReactCommon.MutationResult<DeleteFlowTemplateMutation>;
export type DeleteFlowTemplateMutationOptions = ApolloReactCommon.BaseMutationOptions<
    DeleteFlowTemplateMutation,
    DeleteFlowTemplateMutationVariables
>;
export const CreateFlowDocument = gql`
    mutation createFlow($modelProvider: String!, $input: String!, $resourceIds: [ID!]) {
        createFlow(modelProvider: $modelProvider, input: $input, resourceIds: $resourceIds) {
            ...flowFragment
        }
    }
    ${FlowFragmentFragmentDoc}
`;
export type CreateFlowMutationFn = ApolloReactCommon.MutationFunction<CreateFlowMutation, CreateFlowMutationVariables>;

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
 *      resourceIds: // value for 'resourceIds'
 *   },
 * });
 */
export function useCreateFlowMutation(
    baseOptions?: ApolloReactHooks.MutationHookOptions<CreateFlowMutation, CreateFlowMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<CreateFlowMutation, CreateFlowMutationVariables>(CreateFlowDocument, options);
}
export type CreateFlowMutationHookResult = ReturnType<typeof useCreateFlowMutation>;
export type CreateFlowMutationResult = ApolloReactCommon.MutationResult<CreateFlowMutation>;
export type CreateFlowMutationOptions = ApolloReactCommon.BaseMutationOptions<
    CreateFlowMutation,
    CreateFlowMutationVariables
>;
export const DeleteFlowDocument = gql`
    mutation deleteFlow($flowId: ID!) {
        deleteFlow(flowId: $flowId)
    }
`;
export type DeleteFlowMutationFn = ApolloReactCommon.MutationFunction<DeleteFlowMutation, DeleteFlowMutationVariables>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<DeleteFlowMutation, DeleteFlowMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<DeleteFlowMutation, DeleteFlowMutationVariables>(DeleteFlowDocument, options);
}
export type DeleteFlowMutationHookResult = ReturnType<typeof useDeleteFlowMutation>;
export type DeleteFlowMutationResult = ApolloReactCommon.MutationResult<DeleteFlowMutation>;
export type DeleteFlowMutationOptions = ApolloReactCommon.BaseMutationOptions<
    DeleteFlowMutation,
    DeleteFlowMutationVariables
>;
export const PutUserInputDocument = gql`
    mutation putUserInput($flowId: ID!, $input: String!, $modelProvider: String, $resourceIds: [ID!]) {
        putUserInput(flowId: $flowId, input: $input, modelProvider: $modelProvider, resourceIds: $resourceIds)
    }
`;
export type PutUserInputMutationFn = ApolloReactCommon.MutationFunction<
    PutUserInputMutation,
    PutUserInputMutationVariables
>;

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
 *      modelProvider: // value for 'modelProvider'
 *      resourceIds: // value for 'resourceIds'
 *   },
 * });
 */
export function usePutUserInputMutation(
    baseOptions?: ApolloReactHooks.MutationHookOptions<PutUserInputMutation, PutUserInputMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<PutUserInputMutation, PutUserInputMutationVariables>(
        PutUserInputDocument,
        options,
    );
}
export type PutUserInputMutationHookResult = ReturnType<typeof usePutUserInputMutation>;
export type PutUserInputMutationResult = ApolloReactCommon.MutationResult<PutUserInputMutation>;
export type PutUserInputMutationOptions = ApolloReactCommon.BaseMutationOptions<
    PutUserInputMutation,
    PutUserInputMutationVariables
>;
export const FinishFlowDocument = gql`
    mutation finishFlow($flowId: ID!) {
        finishFlow(flowId: $flowId)
    }
`;
export type FinishFlowMutationFn = ApolloReactCommon.MutationFunction<FinishFlowMutation, FinishFlowMutationVariables>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<FinishFlowMutation, FinishFlowMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<FinishFlowMutation, FinishFlowMutationVariables>(FinishFlowDocument, options);
}
export type FinishFlowMutationHookResult = ReturnType<typeof useFinishFlowMutation>;
export type FinishFlowMutationResult = ApolloReactCommon.MutationResult<FinishFlowMutation>;
export type FinishFlowMutationOptions = ApolloReactCommon.BaseMutationOptions<
    FinishFlowMutation,
    FinishFlowMutationVariables
>;
export const StopFlowDocument = gql`
    mutation stopFlow($flowId: ID!) {
        stopFlow(flowId: $flowId)
    }
`;
export type StopFlowMutationFn = ApolloReactCommon.MutationFunction<StopFlowMutation, StopFlowMutationVariables>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<StopFlowMutation, StopFlowMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<StopFlowMutation, StopFlowMutationVariables>(StopFlowDocument, options);
}
export type StopFlowMutationHookResult = ReturnType<typeof useStopFlowMutation>;
export type StopFlowMutationResult = ApolloReactCommon.MutationResult<StopFlowMutation>;
export type StopFlowMutationOptions = ApolloReactCommon.BaseMutationOptions<
    StopFlowMutation,
    StopFlowMutationVariables
>;
export const RenameFlowDocument = gql`
    mutation renameFlow($flowId: ID!, $title: String!) {
        renameFlow(flowId: $flowId, title: $title)
    }
`;
export type RenameFlowMutationFn = ApolloReactCommon.MutationFunction<RenameFlowMutation, RenameFlowMutationVariables>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<RenameFlowMutation, RenameFlowMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<RenameFlowMutation, RenameFlowMutationVariables>(RenameFlowDocument, options);
}
export type RenameFlowMutationHookResult = ReturnType<typeof useRenameFlowMutation>;
export type RenameFlowMutationResult = ApolloReactCommon.MutationResult<RenameFlowMutation>;
export type RenameFlowMutationOptions = ApolloReactCommon.BaseMutationOptions<
    RenameFlowMutation,
    RenameFlowMutationVariables
>;
export const CreateAssistantDocument = gql`
    mutation createAssistant(
        $flowId: ID!
        $modelProvider: String!
        $input: String!
        $useAgents: Boolean!
        $resourceIds: [ID!]
    ) {
        createAssistant(
            flowId: $flowId
            modelProvider: $modelProvider
            input: $input
            useAgents: $useAgents
            resourceIds: $resourceIds
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
    ${AssistantFragmentFragmentDoc}
`;
export type CreateAssistantMutationFn = ApolloReactCommon.MutationFunction<
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
 *      resourceIds: // value for 'resourceIds'
 *   },
 * });
 */
export function useCreateAssistantMutation(
    baseOptions?: ApolloReactHooks.MutationHookOptions<CreateAssistantMutation, CreateAssistantMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<CreateAssistantMutation, CreateAssistantMutationVariables>(
        CreateAssistantDocument,
        options,
    );
}
export type CreateAssistantMutationHookResult = ReturnType<typeof useCreateAssistantMutation>;
export type CreateAssistantMutationResult = ApolloReactCommon.MutationResult<CreateAssistantMutation>;
export type CreateAssistantMutationOptions = ApolloReactCommon.BaseMutationOptions<
    CreateAssistantMutation,
    CreateAssistantMutationVariables
>;
export const CallAssistantDocument = gql`
    mutation callAssistant(
        $flowId: ID!
        $assistantId: ID!
        $input: String!
        $useAgents: Boolean!
        $resourceIds: [ID!]
    ) {
        callAssistant(
            flowId: $flowId
            assistantId: $assistantId
            input: $input
            useAgents: $useAgents
            resourceIds: $resourceIds
        )
    }
`;
export type CallAssistantMutationFn = ApolloReactCommon.MutationFunction<
    CallAssistantMutation,
    CallAssistantMutationVariables
>;

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
 *      resourceIds: // value for 'resourceIds'
 *   },
 * });
 */
export function useCallAssistantMutation(
    baseOptions?: ApolloReactHooks.MutationHookOptions<CallAssistantMutation, CallAssistantMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<CallAssistantMutation, CallAssistantMutationVariables>(
        CallAssistantDocument,
        options,
    );
}
export type CallAssistantMutationHookResult = ReturnType<typeof useCallAssistantMutation>;
export type CallAssistantMutationResult = ApolloReactCommon.MutationResult<CallAssistantMutation>;
export type CallAssistantMutationOptions = ApolloReactCommon.BaseMutationOptions<
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
`;
export type StopAssistantMutationFn = ApolloReactCommon.MutationFunction<
    StopAssistantMutation,
    StopAssistantMutationVariables
>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<StopAssistantMutation, StopAssistantMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<StopAssistantMutation, StopAssistantMutationVariables>(
        StopAssistantDocument,
        options,
    );
}
export type StopAssistantMutationHookResult = ReturnType<typeof useStopAssistantMutation>;
export type StopAssistantMutationResult = ApolloReactCommon.MutationResult<StopAssistantMutation>;
export type StopAssistantMutationOptions = ApolloReactCommon.BaseMutationOptions<
    StopAssistantMutation,
    StopAssistantMutationVariables
>;
export const DeleteAssistantDocument = gql`
    mutation deleteAssistant($flowId: ID!, $assistantId: ID!) {
        deleteAssistant(flowId: $flowId, assistantId: $assistantId)
    }
`;
export type DeleteAssistantMutationFn = ApolloReactCommon.MutationFunction<
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
    baseOptions?: ApolloReactHooks.MutationHookOptions<DeleteAssistantMutation, DeleteAssistantMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<DeleteAssistantMutation, DeleteAssistantMutationVariables>(
        DeleteAssistantDocument,
        options,
    );
}
export type DeleteAssistantMutationHookResult = ReturnType<typeof useDeleteAssistantMutation>;
export type DeleteAssistantMutationResult = ApolloReactCommon.MutationResult<DeleteAssistantMutation>;
export type DeleteAssistantMutationOptions = ApolloReactCommon.BaseMutationOptions<
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
`;
export type TestAgentMutationFn = ApolloReactCommon.MutationFunction<TestAgentMutation, TestAgentMutationVariables>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<TestAgentMutation, TestAgentMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<TestAgentMutation, TestAgentMutationVariables>(TestAgentDocument, options);
}
export type TestAgentMutationHookResult = ReturnType<typeof useTestAgentMutation>;
export type TestAgentMutationResult = ApolloReactCommon.MutationResult<TestAgentMutation>;
export type TestAgentMutationOptions = ApolloReactCommon.BaseMutationOptions<
    TestAgentMutation,
    TestAgentMutationVariables
>;
export const TestProviderDocument = gql`
    mutation testProvider($type: ProviderType!, $agents: AgentsConfigInput!) {
        testProvider(type: $type, agents: $agents) {
            ...providerTestResultFragment
        }
    }
    ${ProviderTestResultFragmentFragmentDoc}
`;
export type TestProviderMutationFn = ApolloReactCommon.MutationFunction<
    TestProviderMutation,
    TestProviderMutationVariables
>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<TestProviderMutation, TestProviderMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<TestProviderMutation, TestProviderMutationVariables>(
        TestProviderDocument,
        options,
    );
}
export type TestProviderMutationHookResult = ReturnType<typeof useTestProviderMutation>;
export type TestProviderMutationResult = ApolloReactCommon.MutationResult<TestProviderMutation>;
export type TestProviderMutationOptions = ApolloReactCommon.BaseMutationOptions<
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
`;
export type CreateProviderMutationFn = ApolloReactCommon.MutationFunction<
    CreateProviderMutation,
    CreateProviderMutationVariables
>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<CreateProviderMutation, CreateProviderMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<CreateProviderMutation, CreateProviderMutationVariables>(
        CreateProviderDocument,
        options,
    );
}
export type CreateProviderMutationHookResult = ReturnType<typeof useCreateProviderMutation>;
export type CreateProviderMutationResult = ApolloReactCommon.MutationResult<CreateProviderMutation>;
export type CreateProviderMutationOptions = ApolloReactCommon.BaseMutationOptions<
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
`;
export type UpdateProviderMutationFn = ApolloReactCommon.MutationFunction<
    UpdateProviderMutation,
    UpdateProviderMutationVariables
>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<UpdateProviderMutation, UpdateProviderMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<UpdateProviderMutation, UpdateProviderMutationVariables>(
        UpdateProviderDocument,
        options,
    );
}
export type UpdateProviderMutationHookResult = ReturnType<typeof useUpdateProviderMutation>;
export type UpdateProviderMutationResult = ApolloReactCommon.MutationResult<UpdateProviderMutation>;
export type UpdateProviderMutationOptions = ApolloReactCommon.BaseMutationOptions<
    UpdateProviderMutation,
    UpdateProviderMutationVariables
>;
export const DeleteProviderDocument = gql`
    mutation deleteProvider($providerId: ID!) {
        deleteProvider(providerId: $providerId)
    }
`;
export type DeleteProviderMutationFn = ApolloReactCommon.MutationFunction<
    DeleteProviderMutation,
    DeleteProviderMutationVariables
>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<DeleteProviderMutation, DeleteProviderMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<DeleteProviderMutation, DeleteProviderMutationVariables>(
        DeleteProviderDocument,
        options,
    );
}
export type DeleteProviderMutationHookResult = ReturnType<typeof useDeleteProviderMutation>;
export type DeleteProviderMutationResult = ApolloReactCommon.MutationResult<DeleteProviderMutation>;
export type DeleteProviderMutationOptions = ApolloReactCommon.BaseMutationOptions<
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
export type ValidatePromptMutationFn = ApolloReactCommon.MutationFunction<
    ValidatePromptMutation,
    ValidatePromptMutationVariables
>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<ValidatePromptMutation, ValidatePromptMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<ValidatePromptMutation, ValidatePromptMutationVariables>(
        ValidatePromptDocument,
        options,
    );
}
export type ValidatePromptMutationHookResult = ReturnType<typeof useValidatePromptMutation>;
export type ValidatePromptMutationResult = ApolloReactCommon.MutationResult<ValidatePromptMutation>;
export type ValidatePromptMutationOptions = ApolloReactCommon.BaseMutationOptions<
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
export type CreatePromptMutationFn = ApolloReactCommon.MutationFunction<
    CreatePromptMutation,
    CreatePromptMutationVariables
>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<CreatePromptMutation, CreatePromptMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<CreatePromptMutation, CreatePromptMutationVariables>(
        CreatePromptDocument,
        options,
    );
}
export type CreatePromptMutationHookResult = ReturnType<typeof useCreatePromptMutation>;
export type CreatePromptMutationResult = ApolloReactCommon.MutationResult<CreatePromptMutation>;
export type CreatePromptMutationOptions = ApolloReactCommon.BaseMutationOptions<
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
export type UpdatePromptMutationFn = ApolloReactCommon.MutationFunction<
    UpdatePromptMutation,
    UpdatePromptMutationVariables
>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<UpdatePromptMutation, UpdatePromptMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<UpdatePromptMutation, UpdatePromptMutationVariables>(
        UpdatePromptDocument,
        options,
    );
}
export type UpdatePromptMutationHookResult = ReturnType<typeof useUpdatePromptMutation>;
export type UpdatePromptMutationResult = ApolloReactCommon.MutationResult<UpdatePromptMutation>;
export type UpdatePromptMutationOptions = ApolloReactCommon.BaseMutationOptions<
    UpdatePromptMutation,
    UpdatePromptMutationVariables
>;
export const DeletePromptDocument = gql`
    mutation deletePrompt($promptId: ID!) {
        deletePrompt(promptId: $promptId)
    }
`;
export type DeletePromptMutationFn = ApolloReactCommon.MutationFunction<
    DeletePromptMutation,
    DeletePromptMutationVariables
>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<DeletePromptMutation, DeletePromptMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<DeletePromptMutation, DeletePromptMutationVariables>(
        DeletePromptDocument,
        options,
    );
}
export type DeletePromptMutationHookResult = ReturnType<typeof useDeletePromptMutation>;
export type DeletePromptMutationResult = ApolloReactCommon.MutationResult<DeletePromptMutation>;
export type DeletePromptMutationOptions = ApolloReactCommon.BaseMutationOptions<
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
export type CreateApiTokenMutationFn = ApolloReactCommon.MutationFunction<
    CreateApiTokenMutation,
    CreateApiTokenMutationVariables
>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<CreateApiTokenMutation, CreateApiTokenMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<CreateApiTokenMutation, CreateApiTokenMutationVariables>(
        CreateApiTokenDocument,
        options,
    );
}
export type CreateApiTokenMutationHookResult = ReturnType<typeof useCreateApiTokenMutation>;
export type CreateApiTokenMutationResult = ApolloReactCommon.MutationResult<CreateApiTokenMutation>;
export type CreateApiTokenMutationOptions = ApolloReactCommon.BaseMutationOptions<
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
export type UpdateApiTokenMutationFn = ApolloReactCommon.MutationFunction<
    UpdateApiTokenMutation,
    UpdateApiTokenMutationVariables
>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<UpdateApiTokenMutation, UpdateApiTokenMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<UpdateApiTokenMutation, UpdateApiTokenMutationVariables>(
        UpdateApiTokenDocument,
        options,
    );
}
export type UpdateApiTokenMutationHookResult = ReturnType<typeof useUpdateApiTokenMutation>;
export type UpdateApiTokenMutationResult = ApolloReactCommon.MutationResult<UpdateApiTokenMutation>;
export type UpdateApiTokenMutationOptions = ApolloReactCommon.BaseMutationOptions<
    UpdateApiTokenMutation,
    UpdateApiTokenMutationVariables
>;
export const DeleteApiTokenDocument = gql`
    mutation deleteAPIToken($tokenId: String!) {
        deleteAPIToken(tokenId: $tokenId)
    }
`;
export type DeleteApiTokenMutationFn = ApolloReactCommon.MutationFunction<
    DeleteApiTokenMutation,
    DeleteApiTokenMutationVariables
>;

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
    baseOptions?: ApolloReactHooks.MutationHookOptions<DeleteApiTokenMutation, DeleteApiTokenMutationVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<DeleteApiTokenMutation, DeleteApiTokenMutationVariables>(
        DeleteApiTokenDocument,
        options,
    );
}
export type DeleteApiTokenMutationHookResult = ReturnType<typeof useDeleteApiTokenMutation>;
export type DeleteApiTokenMutationResult = ApolloReactCommon.MutationResult<DeleteApiTokenMutation>;
export type DeleteApiTokenMutationOptions = ApolloReactCommon.BaseMutationOptions<
    DeleteApiTokenMutation,
    DeleteApiTokenMutationVariables
>;
export const CreateKnowledgeDocumentDocument = gql`
    mutation createKnowledgeDocument($input: CreateKnowledgeDocumentInput!) {
        createKnowledgeDocument(input: $input) {
            ...knowledgeDocumentFragment
        }
    }
    ${KnowledgeDocumentFragmentFragmentDoc}
`;
export type CreateKnowledgeDocumentMutationFn = ApolloReactCommon.MutationFunction<
    CreateKnowledgeDocumentMutation,
    CreateKnowledgeDocumentMutationVariables
>;

/**
 * __useCreateKnowledgeDocumentMutation__
 *
 * To run a mutation, you first call `useCreateKnowledgeDocumentMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateKnowledgeDocumentMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createKnowledgeDocumentMutation, { data, loading, error }] = useCreateKnowledgeDocumentMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateKnowledgeDocumentMutation(
    baseOptions?: ApolloReactHooks.MutationHookOptions<
        CreateKnowledgeDocumentMutation,
        CreateKnowledgeDocumentMutationVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<CreateKnowledgeDocumentMutation, CreateKnowledgeDocumentMutationVariables>(
        CreateKnowledgeDocumentDocument,
        options,
    );
}
export type CreateKnowledgeDocumentMutationHookResult = ReturnType<typeof useCreateKnowledgeDocumentMutation>;
export type CreateKnowledgeDocumentMutationResult = ApolloReactCommon.MutationResult<CreateKnowledgeDocumentMutation>;
export type CreateKnowledgeDocumentMutationOptions = ApolloReactCommon.BaseMutationOptions<
    CreateKnowledgeDocumentMutation,
    CreateKnowledgeDocumentMutationVariables
>;
export const UpdateKnowledgeDocumentDocument = gql`
    mutation updateKnowledgeDocument($id: String!, $input: UpdateKnowledgeDocumentInput!) {
        updateKnowledgeDocument(id: $id, input: $input) {
            ...knowledgeDocumentFragment
        }
    }
    ${KnowledgeDocumentFragmentFragmentDoc}
`;
export type UpdateKnowledgeDocumentMutationFn = ApolloReactCommon.MutationFunction<
    UpdateKnowledgeDocumentMutation,
    UpdateKnowledgeDocumentMutationVariables
>;

/**
 * __useUpdateKnowledgeDocumentMutation__
 *
 * To run a mutation, you first call `useUpdateKnowledgeDocumentMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateKnowledgeDocumentMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateKnowledgeDocumentMutation, { data, loading, error }] = useUpdateKnowledgeDocumentMutation({
 *   variables: {
 *      id: // value for 'id'
 *      input: // value for 'input'
 *   },
 * });
 */
export function useUpdateKnowledgeDocumentMutation(
    baseOptions?: ApolloReactHooks.MutationHookOptions<
        UpdateKnowledgeDocumentMutation,
        UpdateKnowledgeDocumentMutationVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<UpdateKnowledgeDocumentMutation, UpdateKnowledgeDocumentMutationVariables>(
        UpdateKnowledgeDocumentDocument,
        options,
    );
}
export type UpdateKnowledgeDocumentMutationHookResult = ReturnType<typeof useUpdateKnowledgeDocumentMutation>;
export type UpdateKnowledgeDocumentMutationResult = ApolloReactCommon.MutationResult<UpdateKnowledgeDocumentMutation>;
export type UpdateKnowledgeDocumentMutationOptions = ApolloReactCommon.BaseMutationOptions<
    UpdateKnowledgeDocumentMutation,
    UpdateKnowledgeDocumentMutationVariables
>;
export const DeleteKnowledgeDocumentDocument = gql`
    mutation deleteKnowledgeDocument($id: String!) {
        deleteKnowledgeDocument(id: $id)
    }
`;
export type DeleteKnowledgeDocumentMutationFn = ApolloReactCommon.MutationFunction<
    DeleteKnowledgeDocumentMutation,
    DeleteKnowledgeDocumentMutationVariables
>;

/**
 * __useDeleteKnowledgeDocumentMutation__
 *
 * To run a mutation, you first call `useDeleteKnowledgeDocumentMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteKnowledgeDocumentMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteKnowledgeDocumentMutation, { data, loading, error }] = useDeleteKnowledgeDocumentMutation({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useDeleteKnowledgeDocumentMutation(
    baseOptions?: ApolloReactHooks.MutationHookOptions<
        DeleteKnowledgeDocumentMutation,
        DeleteKnowledgeDocumentMutationVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useMutation<DeleteKnowledgeDocumentMutation, DeleteKnowledgeDocumentMutationVariables>(
        DeleteKnowledgeDocumentDocument,
        options,
    );
}
export type DeleteKnowledgeDocumentMutationHookResult = ReturnType<typeof useDeleteKnowledgeDocumentMutation>;
export type DeleteKnowledgeDocumentMutationResult = ApolloReactCommon.MutationResult<DeleteKnowledgeDocumentMutation>;
export type DeleteKnowledgeDocumentMutationOptions = ApolloReactCommon.BaseMutationOptions<
    DeleteKnowledgeDocumentMutation,
    DeleteKnowledgeDocumentMutationVariables
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
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<
        TerminalLogAddedSubscription,
        TerminalLogAddedSubscriptionVariables
    > &
        ({ variables: TerminalLogAddedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<TerminalLogAddedSubscription, TerminalLogAddedSubscriptionVariables>(
        TerminalLogAddedDocument,
        options,
    );
}
export type TerminalLogAddedSubscriptionHookResult = ReturnType<typeof useTerminalLogAddedSubscription>;
export type TerminalLogAddedSubscriptionResult = ApolloReactCommon.SubscriptionResult<TerminalLogAddedSubscription>;
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
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<
        MessageLogAddedSubscription,
        MessageLogAddedSubscriptionVariables
    > &
        ({ variables: MessageLogAddedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<MessageLogAddedSubscription, MessageLogAddedSubscriptionVariables>(
        MessageLogAddedDocument,
        options,
    );
}
export type MessageLogAddedSubscriptionHookResult = ReturnType<typeof useMessageLogAddedSubscription>;
export type MessageLogAddedSubscriptionResult = ApolloReactCommon.SubscriptionResult<MessageLogAddedSubscription>;
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
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<
        MessageLogUpdatedSubscription,
        MessageLogUpdatedSubscriptionVariables
    > &
        ({ variables: MessageLogUpdatedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<MessageLogUpdatedSubscription, MessageLogUpdatedSubscriptionVariables>(
        MessageLogUpdatedDocument,
        options,
    );
}
export type MessageLogUpdatedSubscriptionHookResult = ReturnType<typeof useMessageLogUpdatedSubscription>;
export type MessageLogUpdatedSubscriptionResult = ApolloReactCommon.SubscriptionResult<MessageLogUpdatedSubscription>;
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
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<
        ScreenshotAddedSubscription,
        ScreenshotAddedSubscriptionVariables
    > &
        ({ variables: ScreenshotAddedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<ScreenshotAddedSubscription, ScreenshotAddedSubscriptionVariables>(
        ScreenshotAddedDocument,
        options,
    );
}
export type ScreenshotAddedSubscriptionHookResult = ReturnType<typeof useScreenshotAddedSubscription>;
export type ScreenshotAddedSubscriptionResult = ApolloReactCommon.SubscriptionResult<ScreenshotAddedSubscription>;
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
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<
        AgentLogAddedSubscription,
        AgentLogAddedSubscriptionVariables
    > &
        ({ variables: AgentLogAddedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<AgentLogAddedSubscription, AgentLogAddedSubscriptionVariables>(
        AgentLogAddedDocument,
        options,
    );
}
export type AgentLogAddedSubscriptionHookResult = ReturnType<typeof useAgentLogAddedSubscription>;
export type AgentLogAddedSubscriptionResult = ApolloReactCommon.SubscriptionResult<AgentLogAddedSubscription>;
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
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<
        SearchLogAddedSubscription,
        SearchLogAddedSubscriptionVariables
    > &
        ({ variables: SearchLogAddedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<SearchLogAddedSubscription, SearchLogAddedSubscriptionVariables>(
        SearchLogAddedDocument,
        options,
    );
}
export type SearchLogAddedSubscriptionHookResult = ReturnType<typeof useSearchLogAddedSubscription>;
export type SearchLogAddedSubscriptionResult = ApolloReactCommon.SubscriptionResult<SearchLogAddedSubscription>;
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
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<
        VectorStoreLogAddedSubscription,
        VectorStoreLogAddedSubscriptionVariables
    > &
        ({ variables: VectorStoreLogAddedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<VectorStoreLogAddedSubscription, VectorStoreLogAddedSubscriptionVariables>(
        VectorStoreLogAddedDocument,
        options,
    );
}
export type VectorStoreLogAddedSubscriptionHookResult = ReturnType<typeof useVectorStoreLogAddedSubscription>;
export type VectorStoreLogAddedSubscriptionResult =
    ApolloReactCommon.SubscriptionResult<VectorStoreLogAddedSubscription>;
export const AssistantCreatedDocument = gql`
    subscription assistantCreated($flowId: ID!) {
        assistantCreated(flowId: $flowId) {
            ...assistantFragment
        }
    }
    ${AssistantFragmentFragmentDoc}
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
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<
        AssistantCreatedSubscription,
        AssistantCreatedSubscriptionVariables
    > &
        ({ variables: AssistantCreatedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<AssistantCreatedSubscription, AssistantCreatedSubscriptionVariables>(
        AssistantCreatedDocument,
        options,
    );
}
export type AssistantCreatedSubscriptionHookResult = ReturnType<typeof useAssistantCreatedSubscription>;
export type AssistantCreatedSubscriptionResult = ApolloReactCommon.SubscriptionResult<AssistantCreatedSubscription>;
export const AssistantUpdatedDocument = gql`
    subscription assistantUpdated($flowId: ID!) {
        assistantUpdated(flowId: $flowId) {
            ...assistantFragment
        }
    }
    ${AssistantFragmentFragmentDoc}
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
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<
        AssistantUpdatedSubscription,
        AssistantUpdatedSubscriptionVariables
    > &
        ({ variables: AssistantUpdatedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<AssistantUpdatedSubscription, AssistantUpdatedSubscriptionVariables>(
        AssistantUpdatedDocument,
        options,
    );
}
export type AssistantUpdatedSubscriptionHookResult = ReturnType<typeof useAssistantUpdatedSubscription>;
export type AssistantUpdatedSubscriptionResult = ApolloReactCommon.SubscriptionResult<AssistantUpdatedSubscription>;
export const AssistantDeletedDocument = gql`
    subscription assistantDeleted($flowId: ID!) {
        assistantDeleted(flowId: $flowId) {
            ...assistantFragment
        }
    }
    ${AssistantFragmentFragmentDoc}
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
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<
        AssistantDeletedSubscription,
        AssistantDeletedSubscriptionVariables
    > &
        ({ variables: AssistantDeletedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<AssistantDeletedSubscription, AssistantDeletedSubscriptionVariables>(
        AssistantDeletedDocument,
        options,
    );
}
export type AssistantDeletedSubscriptionHookResult = ReturnType<typeof useAssistantDeletedSubscription>;
export type AssistantDeletedSubscriptionResult = ApolloReactCommon.SubscriptionResult<AssistantDeletedSubscription>;
export const FlowFileAddedDocument = gql`
    subscription flowFileAdded($flowId: ID!) {
        flowFileAdded(flowId: $flowId) {
            ...flowFileFragment
        }
    }
    ${FlowFileFragmentFragmentDoc}
`;

/**
 * __useFlowFileAddedSubscription__
 *
 * To run a query within a React component, call `useFlowFileAddedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useFlowFileAddedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowFileAddedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useFlowFileAddedSubscription(
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<
        FlowFileAddedSubscription,
        FlowFileAddedSubscriptionVariables
    > &
        ({ variables: FlowFileAddedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<FlowFileAddedSubscription, FlowFileAddedSubscriptionVariables>(
        FlowFileAddedDocument,
        options,
    );
}
export type FlowFileAddedSubscriptionHookResult = ReturnType<typeof useFlowFileAddedSubscription>;
export type FlowFileAddedSubscriptionResult = ApolloReactCommon.SubscriptionResult<FlowFileAddedSubscription>;
export const FlowFileUpdatedDocument = gql`
    subscription flowFileUpdated($flowId: ID!) {
        flowFileUpdated(flowId: $flowId) {
            ...flowFileFragment
        }
    }
    ${FlowFileFragmentFragmentDoc}
`;

/**
 * __useFlowFileUpdatedSubscription__
 *
 * To run a query within a React component, call `useFlowFileUpdatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useFlowFileUpdatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowFileUpdatedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useFlowFileUpdatedSubscription(
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<
        FlowFileUpdatedSubscription,
        FlowFileUpdatedSubscriptionVariables
    > &
        ({ variables: FlowFileUpdatedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<FlowFileUpdatedSubscription, FlowFileUpdatedSubscriptionVariables>(
        FlowFileUpdatedDocument,
        options,
    );
}
export type FlowFileUpdatedSubscriptionHookResult = ReturnType<typeof useFlowFileUpdatedSubscription>;
export type FlowFileUpdatedSubscriptionResult = ApolloReactCommon.SubscriptionResult<FlowFileUpdatedSubscription>;
export const FlowFileDeletedDocument = gql`
    subscription flowFileDeleted($flowId: ID!) {
        flowFileDeleted(flowId: $flowId) {
            ...flowFileFragment
        }
    }
    ${FlowFileFragmentFragmentDoc}
`;

/**
 * __useFlowFileDeletedSubscription__
 *
 * To run a query within a React component, call `useFlowFileDeletedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useFlowFileDeletedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowFileDeletedSubscription({
 *   variables: {
 *      flowId: // value for 'flowId'
 *   },
 * });
 */
export function useFlowFileDeletedSubscription(
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<
        FlowFileDeletedSubscription,
        FlowFileDeletedSubscriptionVariables
    > &
        ({ variables: FlowFileDeletedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<FlowFileDeletedSubscription, FlowFileDeletedSubscriptionVariables>(
        FlowFileDeletedDocument,
        options,
    );
}
export type FlowFileDeletedSubscriptionHookResult = ReturnType<typeof useFlowFileDeletedSubscription>;
export type FlowFileDeletedSubscriptionResult = ApolloReactCommon.SubscriptionResult<FlowFileDeletedSubscription>;
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
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<
        AssistantLogAddedSubscription,
        AssistantLogAddedSubscriptionVariables
    > &
        ({ variables: AssistantLogAddedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<AssistantLogAddedSubscription, AssistantLogAddedSubscriptionVariables>(
        AssistantLogAddedDocument,
        options,
    );
}
export type AssistantLogAddedSubscriptionHookResult = ReturnType<typeof useAssistantLogAddedSubscription>;
export type AssistantLogAddedSubscriptionResult = ApolloReactCommon.SubscriptionResult<AssistantLogAddedSubscription>;
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
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<
        AssistantLogUpdatedSubscription,
        AssistantLogUpdatedSubscriptionVariables
    > &
        ({ variables: AssistantLogUpdatedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<AssistantLogUpdatedSubscription, AssistantLogUpdatedSubscriptionVariables>(
        AssistantLogUpdatedDocument,
        options,
    );
}
export type AssistantLogUpdatedSubscriptionHookResult = ReturnType<typeof useAssistantLogUpdatedSubscription>;
export type AssistantLogUpdatedSubscriptionResult =
    ApolloReactCommon.SubscriptionResult<AssistantLogUpdatedSubscription>;
export const FlowCreatedDocument = gql`
    subscription flowCreated {
        flowCreated {
            ...flowFragment
        }
    }
    ${FlowFragmentFragmentDoc}
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
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<FlowCreatedSubscription, FlowCreatedSubscriptionVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<FlowCreatedSubscription, FlowCreatedSubscriptionVariables>(
        FlowCreatedDocument,
        options,
    );
}
export type FlowCreatedSubscriptionHookResult = ReturnType<typeof useFlowCreatedSubscription>;
export type FlowCreatedSubscriptionResult = ApolloReactCommon.SubscriptionResult<FlowCreatedSubscription>;
export const FlowDeletedDocument = gql`
    subscription flowDeleted {
        flowDeleted {
            ...flowFragment
        }
    }
    ${FlowFragmentFragmentDoc}
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
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<FlowDeletedSubscription, FlowDeletedSubscriptionVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<FlowDeletedSubscription, FlowDeletedSubscriptionVariables>(
        FlowDeletedDocument,
        options,
    );
}
export type FlowDeletedSubscriptionHookResult = ReturnType<typeof useFlowDeletedSubscription>;
export type FlowDeletedSubscriptionResult = ApolloReactCommon.SubscriptionResult<FlowDeletedSubscription>;
export const FlowUpdatedDocument = gql`
    subscription flowUpdated {
        flowUpdated {
            ...flowFragment
        }
    }
    ${FlowFragmentFragmentDoc}
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
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<FlowUpdatedSubscription, FlowUpdatedSubscriptionVariables>,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<FlowUpdatedSubscription, FlowUpdatedSubscriptionVariables>(
        FlowUpdatedDocument,
        options,
    );
}
export type FlowUpdatedSubscriptionHookResult = ReturnType<typeof useFlowUpdatedSubscription>;
export type FlowUpdatedSubscriptionResult = ApolloReactCommon.SubscriptionResult<FlowUpdatedSubscription>;
export const TaskCreatedDocument = gql`
    subscription taskCreated($flowId: ID!) {
        taskCreated(flowId: $flowId) {
            ...taskFragment
        }
    }
    ${TaskFragmentFragmentDoc}
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
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<TaskCreatedSubscription, TaskCreatedSubscriptionVariables> &
        ({ variables: TaskCreatedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<TaskCreatedSubscription, TaskCreatedSubscriptionVariables>(
        TaskCreatedDocument,
        options,
    );
}
export type TaskCreatedSubscriptionHookResult = ReturnType<typeof useTaskCreatedSubscription>;
export type TaskCreatedSubscriptionResult = ApolloReactCommon.SubscriptionResult<TaskCreatedSubscription>;
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
    baseOptions: ApolloReactHooks.SubscriptionHookOptions<TaskUpdatedSubscription, TaskUpdatedSubscriptionVariables> &
        ({ variables: TaskUpdatedSubscriptionVariables; skip?: boolean } | { skip: boolean }),
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<TaskUpdatedSubscription, TaskUpdatedSubscriptionVariables>(
        TaskUpdatedDocument,
        options,
    );
}
export type TaskUpdatedSubscriptionHookResult = ReturnType<typeof useTaskUpdatedSubscription>;
export type TaskUpdatedSubscriptionResult = ApolloReactCommon.SubscriptionResult<TaskUpdatedSubscription>;
export const ProviderCreatedDocument = gql`
    subscription providerCreated {
        providerCreated {
            ...providerConfigFragment
        }
    }
    ${ProviderConfigFragmentFragmentDoc}
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
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<
        ProviderCreatedSubscription,
        ProviderCreatedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<ProviderCreatedSubscription, ProviderCreatedSubscriptionVariables>(
        ProviderCreatedDocument,
        options,
    );
}
export type ProviderCreatedSubscriptionHookResult = ReturnType<typeof useProviderCreatedSubscription>;
export type ProviderCreatedSubscriptionResult = ApolloReactCommon.SubscriptionResult<ProviderCreatedSubscription>;
export const ProviderUpdatedDocument = gql`
    subscription providerUpdated {
        providerUpdated {
            ...providerConfigFragment
        }
    }
    ${ProviderConfigFragmentFragmentDoc}
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
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<
        ProviderUpdatedSubscription,
        ProviderUpdatedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<ProviderUpdatedSubscription, ProviderUpdatedSubscriptionVariables>(
        ProviderUpdatedDocument,
        options,
    );
}
export type ProviderUpdatedSubscriptionHookResult = ReturnType<typeof useProviderUpdatedSubscription>;
export type ProviderUpdatedSubscriptionResult = ApolloReactCommon.SubscriptionResult<ProviderUpdatedSubscription>;
export const ProviderDeletedDocument = gql`
    subscription providerDeleted {
        providerDeleted {
            ...providerConfigFragment
        }
    }
    ${ProviderConfigFragmentFragmentDoc}
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
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<
        ProviderDeletedSubscription,
        ProviderDeletedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<ProviderDeletedSubscription, ProviderDeletedSubscriptionVariables>(
        ProviderDeletedDocument,
        options,
    );
}
export type ProviderDeletedSubscriptionHookResult = ReturnType<typeof useProviderDeletedSubscription>;
export type ProviderDeletedSubscriptionResult = ApolloReactCommon.SubscriptionResult<ProviderDeletedSubscription>;
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
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<
        ApiTokenCreatedSubscription,
        ApiTokenCreatedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<ApiTokenCreatedSubscription, ApiTokenCreatedSubscriptionVariables>(
        ApiTokenCreatedDocument,
        options,
    );
}
export type ApiTokenCreatedSubscriptionHookResult = ReturnType<typeof useApiTokenCreatedSubscription>;
export type ApiTokenCreatedSubscriptionResult = ApolloReactCommon.SubscriptionResult<ApiTokenCreatedSubscription>;
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
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<
        ApiTokenUpdatedSubscription,
        ApiTokenUpdatedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<ApiTokenUpdatedSubscription, ApiTokenUpdatedSubscriptionVariables>(
        ApiTokenUpdatedDocument,
        options,
    );
}
export type ApiTokenUpdatedSubscriptionHookResult = ReturnType<typeof useApiTokenUpdatedSubscription>;
export type ApiTokenUpdatedSubscriptionResult = ApolloReactCommon.SubscriptionResult<ApiTokenUpdatedSubscription>;
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
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<
        ApiTokenDeletedSubscription,
        ApiTokenDeletedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<ApiTokenDeletedSubscription, ApiTokenDeletedSubscriptionVariables>(
        ApiTokenDeletedDocument,
        options,
    );
}
export type ApiTokenDeletedSubscriptionHookResult = ReturnType<typeof useApiTokenDeletedSubscription>;
export type ApiTokenDeletedSubscriptionResult = ApolloReactCommon.SubscriptionResult<ApiTokenDeletedSubscription>;
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
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<
        SettingsUserUpdatedSubscription,
        SettingsUserUpdatedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<SettingsUserUpdatedSubscription, SettingsUserUpdatedSubscriptionVariables>(
        SettingsUserUpdatedDocument,
        options,
    );
}
export type SettingsUserUpdatedSubscriptionHookResult = ReturnType<typeof useSettingsUserUpdatedSubscription>;
export type SettingsUserUpdatedSubscriptionResult =
    ApolloReactCommon.SubscriptionResult<SettingsUserUpdatedSubscription>;
export const FlowTemplateCreatedDocument = gql`
    subscription flowTemplateCreated {
        flowTemplateCreated {
            ...flowTemplateFragment
        }
    }
    ${FlowTemplateFragmentFragmentDoc}
`;

/**
 * __useFlowTemplateCreatedSubscription__
 *
 * To run a query within a React component, call `useFlowTemplateCreatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useFlowTemplateCreatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowTemplateCreatedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useFlowTemplateCreatedSubscription(
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<
        FlowTemplateCreatedSubscription,
        FlowTemplateCreatedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<FlowTemplateCreatedSubscription, FlowTemplateCreatedSubscriptionVariables>(
        FlowTemplateCreatedDocument,
        options,
    );
}
export type FlowTemplateCreatedSubscriptionHookResult = ReturnType<typeof useFlowTemplateCreatedSubscription>;
export type FlowTemplateCreatedSubscriptionResult =
    ApolloReactCommon.SubscriptionResult<FlowTemplateCreatedSubscription>;
export const FlowTemplateUpdatedDocument = gql`
    subscription flowTemplateUpdated {
        flowTemplateUpdated {
            ...flowTemplateFragment
        }
    }
    ${FlowTemplateFragmentFragmentDoc}
`;

/**
 * __useFlowTemplateUpdatedSubscription__
 *
 * To run a query within a React component, call `useFlowTemplateUpdatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useFlowTemplateUpdatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowTemplateUpdatedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useFlowTemplateUpdatedSubscription(
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<
        FlowTemplateUpdatedSubscription,
        FlowTemplateUpdatedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<FlowTemplateUpdatedSubscription, FlowTemplateUpdatedSubscriptionVariables>(
        FlowTemplateUpdatedDocument,
        options,
    );
}
export type FlowTemplateUpdatedSubscriptionHookResult = ReturnType<typeof useFlowTemplateUpdatedSubscription>;
export type FlowTemplateUpdatedSubscriptionResult =
    ApolloReactCommon.SubscriptionResult<FlowTemplateUpdatedSubscription>;
export const FlowTemplateDeletedDocument = gql`
    subscription flowTemplateDeleted {
        flowTemplateDeleted {
            ...flowTemplateFragment
        }
    }
    ${FlowTemplateFragmentFragmentDoc}
`;

/**
 * __useFlowTemplateDeletedSubscription__
 *
 * To run a query within a React component, call `useFlowTemplateDeletedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useFlowTemplateDeletedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFlowTemplateDeletedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useFlowTemplateDeletedSubscription(
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<
        FlowTemplateDeletedSubscription,
        FlowTemplateDeletedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<FlowTemplateDeletedSubscription, FlowTemplateDeletedSubscriptionVariables>(
        FlowTemplateDeletedDocument,
        options,
    );
}
export type FlowTemplateDeletedSubscriptionHookResult = ReturnType<typeof useFlowTemplateDeletedSubscription>;
export type FlowTemplateDeletedSubscriptionResult =
    ApolloReactCommon.SubscriptionResult<FlowTemplateDeletedSubscription>;
export const ResourceAddedDocument = gql`
    subscription resourceAdded {
        resourceAdded {
            ...userResourceFragment
        }
    }
    ${UserResourceFragmentFragmentDoc}
`;

/**
 * __useResourceAddedSubscription__
 *
 * To run a query within a React component, call `useResourceAddedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useResourceAddedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useResourceAddedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useResourceAddedSubscription(
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<
        ResourceAddedSubscription,
        ResourceAddedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<ResourceAddedSubscription, ResourceAddedSubscriptionVariables>(
        ResourceAddedDocument,
        options,
    );
}
export type ResourceAddedSubscriptionHookResult = ReturnType<typeof useResourceAddedSubscription>;
export type ResourceAddedSubscriptionResult = ApolloReactCommon.SubscriptionResult<ResourceAddedSubscription>;
export const ResourceUpdatedDocument = gql`
    subscription resourceUpdated {
        resourceUpdated {
            ...userResourceFragment
        }
    }
    ${UserResourceFragmentFragmentDoc}
`;

/**
 * __useResourceUpdatedSubscription__
 *
 * To run a query within a React component, call `useResourceUpdatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useResourceUpdatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useResourceUpdatedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useResourceUpdatedSubscription(
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<
        ResourceUpdatedSubscription,
        ResourceUpdatedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<ResourceUpdatedSubscription, ResourceUpdatedSubscriptionVariables>(
        ResourceUpdatedDocument,
        options,
    );
}
export type ResourceUpdatedSubscriptionHookResult = ReturnType<typeof useResourceUpdatedSubscription>;
export type ResourceUpdatedSubscriptionResult = ApolloReactCommon.SubscriptionResult<ResourceUpdatedSubscription>;
export const ResourceDeletedDocument = gql`
    subscription resourceDeleted {
        resourceDeleted {
            ...userResourceFragment
        }
    }
    ${UserResourceFragmentFragmentDoc}
`;

/**
 * __useResourceDeletedSubscription__
 *
 * To run a query within a React component, call `useResourceDeletedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useResourceDeletedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useResourceDeletedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useResourceDeletedSubscription(
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<
        ResourceDeletedSubscription,
        ResourceDeletedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<ResourceDeletedSubscription, ResourceDeletedSubscriptionVariables>(
        ResourceDeletedDocument,
        options,
    );
}
export type ResourceDeletedSubscriptionHookResult = ReturnType<typeof useResourceDeletedSubscription>;
export type ResourceDeletedSubscriptionResult = ApolloReactCommon.SubscriptionResult<ResourceDeletedSubscription>;
export const KnowledgeDocumentCreatedDocument = gql`
    subscription knowledgeDocumentCreated {
        knowledgeDocumentCreated {
            ...knowledgeDocumentFragment
        }
    }
    ${KnowledgeDocumentFragmentFragmentDoc}
`;

/**
 * __useKnowledgeDocumentCreatedSubscription__
 *
 * To run a query within a React component, call `useKnowledgeDocumentCreatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useKnowledgeDocumentCreatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useKnowledgeDocumentCreatedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useKnowledgeDocumentCreatedSubscription(
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<
        KnowledgeDocumentCreatedSubscription,
        KnowledgeDocumentCreatedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<
        KnowledgeDocumentCreatedSubscription,
        KnowledgeDocumentCreatedSubscriptionVariables
    >(KnowledgeDocumentCreatedDocument, options);
}
export type KnowledgeDocumentCreatedSubscriptionHookResult = ReturnType<typeof useKnowledgeDocumentCreatedSubscription>;
export type KnowledgeDocumentCreatedSubscriptionResult =
    ApolloReactCommon.SubscriptionResult<KnowledgeDocumentCreatedSubscription>;
export const KnowledgeDocumentUpdatedDocument = gql`
    subscription knowledgeDocumentUpdated {
        knowledgeDocumentUpdated {
            ...knowledgeDocumentFragment
        }
    }
    ${KnowledgeDocumentFragmentFragmentDoc}
`;

/**
 * __useKnowledgeDocumentUpdatedSubscription__
 *
 * To run a query within a React component, call `useKnowledgeDocumentUpdatedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useKnowledgeDocumentUpdatedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useKnowledgeDocumentUpdatedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useKnowledgeDocumentUpdatedSubscription(
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<
        KnowledgeDocumentUpdatedSubscription,
        KnowledgeDocumentUpdatedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<
        KnowledgeDocumentUpdatedSubscription,
        KnowledgeDocumentUpdatedSubscriptionVariables
    >(KnowledgeDocumentUpdatedDocument, options);
}
export type KnowledgeDocumentUpdatedSubscriptionHookResult = ReturnType<typeof useKnowledgeDocumentUpdatedSubscription>;
export type KnowledgeDocumentUpdatedSubscriptionResult =
    ApolloReactCommon.SubscriptionResult<KnowledgeDocumentUpdatedSubscription>;
export const KnowledgeDocumentDeletedDocument = gql`
    subscription knowledgeDocumentDeleted {
        knowledgeDocumentDeleted {
            ...knowledgeDocumentFragment
        }
    }
    ${KnowledgeDocumentFragmentFragmentDoc}
`;

/**
 * __useKnowledgeDocumentDeletedSubscription__
 *
 * To run a query within a React component, call `useKnowledgeDocumentDeletedSubscription` and pass it any options that fit your needs.
 * When your component renders, `useKnowledgeDocumentDeletedSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useKnowledgeDocumentDeletedSubscription({
 *   variables: {
 *   },
 * });
 */
export function useKnowledgeDocumentDeletedSubscription(
    baseOptions?: ApolloReactHooks.SubscriptionHookOptions<
        KnowledgeDocumentDeletedSubscription,
        KnowledgeDocumentDeletedSubscriptionVariables
    >,
) {
    const options = { ...defaultOptions, ...baseOptions };
    return ApolloReactHooks.useSubscription<
        KnowledgeDocumentDeletedSubscription,
        KnowledgeDocumentDeletedSubscriptionVariables
    >(KnowledgeDocumentDeletedDocument, options);
}
export type KnowledgeDocumentDeletedSubscriptionHookResult = ReturnType<typeof useKnowledgeDocumentDeletedSubscription>;
export type KnowledgeDocumentDeletedSubscriptionResult =
    ApolloReactCommon.SubscriptionResult<KnowledgeDocumentDeletedSubscription>;
