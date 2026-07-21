/** Internal type. DO NOT USE DIRECTLY. */
type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
/** Internal type. DO NOT USE DIRECTLY. */
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
import type { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';
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

export enum ModelReasoningMode {
    Adaptive = 'adaptive',
    AdaptiveOnly = 'adaptive_only',
    Budget = 'budget',
}

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
    Minimax = 'minimax',
    Ollama = 'ollama',
    Openai = 'openai',
    Qwen = 'qwen',
}

export type ReasoningConfigInput = {
    effort?: ReasoningEffort | null | undefined;
    maxTokens?: number | null | undefined;
    mode?: ReasoningMode | null | undefined;
};

export enum ReasoningEffort {
    High = 'high',
    Low = 'low',
    Max = 'max',
    Medium = 'medium',
    Xhigh = 'xhigh',
}

export enum ReasoningMode {
    Adaptive = 'adaptive',
    Budget = 'budget',
    Off = 'off',
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
    createdAt: string;
    updatedAt: string;
    terminals: Array<TerminalFragmentFragment> | null;
    provider: ProviderFragmentFragment;
};

export type TerminalFragmentFragment = {
    id: string;
    type: TerminalType;
    name: string;
    image: string;
    connected: boolean;
    createdAt: string;
};

export type TaskFragmentFragment = {
    id: string;
    title: string;
    status: StatusType;
    input: string;
    result: string;
    flowId: string;
    createdAt: string;
    updatedAt: string;
    subtasks: Array<SubtaskFragmentFragment> | null;
};

export type SubtaskFragmentFragment = {
    id: string;
    status: StatusType;
    title: string;
    description: string;
    result: string;
    taskId: string;
    createdAt: string;
    updatedAt: string;
};

export type TerminalLogFragmentFragment = {
    id: string;
    flowId: string;
    taskId: string | null;
    subtaskId: string | null;
    type: TerminalLogType;
    text: string;
    terminal: string;
    createdAt: string;
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
    createdAt: string;
};

export type ScreenshotFragmentFragment = {
    id: string;
    flowId: string;
    taskId: string | null;
    subtaskId: string | null;
    name: string;
    url: string;
    createdAt: string;
};

export type FlowFileFragmentFragment = {
    id: string;
    name: string;
    path: string;
    size: number;
    isDir: boolean;
    modifiedAt: string;
};

export type UserResourceFragmentFragment = {
    id: string;
    userId: string;
    name: string;
    path: string;
    size: number;
    isDir: boolean;
    createdAt: string;
    updatedAt: string;
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
    createdAt: string;
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
    createdAt: string;
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
    createdAt: string;
};

export type AssistantFragmentFragment = {
    id: string;
    title: string;
    status: StatusType;
    flowId: string;
    useAgents: boolean;
    createdAt: string;
    updatedAt: string;
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
    createdAt: string;
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
    thinking: boolean | null;
    reasoning: {
        mode: ModelReasoningMode | null;
        efforts: Array<ReasoningEffort> | null;
        supported: boolean | null;
        cannotDisable: boolean | null;
        defaultOn: boolean | null;
    } | null;
    price: { input: number; output: number; cacheRead: number; cacheWrite: number } | null;
};

export type ProviderFragmentFragment = { name: string; type: ProviderType };

export type ProviderConfigFragmentFragment = {
    id: string;
    name: string;
    type: ProviderType;
    createdAt: string;
    updatedAt: string;
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
    reasoning: { mode: ReasoningMode | null; effort: ReasoningEffort | null; maxTokens: number | null } | null;
    price: { input: number; output: number; cacheRead: number; cacheWrite: number } | null;
};

export type UserPromptFragmentFragment = {
    id: string;
    type: PromptType;
    template: string;
    createdAt: string;
    updatedAt: string;
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
    createdAt: string;
    updatedAt: string;
};

export type ApiTokenWithSecretFragmentFragment = {
    id: string;
    tokenId: string;
    userId: string;
    roleId: string;
    name: string | null;
    ttl: number;
    status: TokenStatus;
    createdAt: string;
    updatedAt: string;
    token: string;
};

export type FlowTemplateFragmentFragment = {
    id: string;
    userId: string;
    title: string;
    text: string;
    createdAt: string;
    updatedAt: string;
};

export type UsageStatsFragmentFragment = {
    totalUsageIn: number;
    totalUsageOut: number;
    totalUsageCacheIn: number;
    totalUsageCacheOut: number;
    totalUsageCostIn: number;
    totalUsageCostOut: number;
};

export type DailyUsageStatsFragmentFragment = { date: string; stats: UsageStatsFragmentFragment };

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

export type DailyToolcallsStatsFragmentFragment = { date: string; stats: ToolcallsStatsFragmentFragment };

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

export type DailyFlowsStatsFragmentFragment = { date: string; stats: FlowsStatsFragmentFragment };

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
            minimax: boolean;
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
            minimax: ProviderConfigFragmentFragment | null;
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
            minimax: Array<ModelConfigFragmentFragment> | null;
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

export type RenameKnowledgeDocumentMutationVariables = Exact<{
    id: string;
    question: string;
}>;

export type RenameKnowledgeDocumentMutation = { renameKnowledgeDocument: KnowledgeDocumentFragmentFragment };

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
        updatedAt: string;
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

export const SettingsFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'settingsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Settings' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'debug' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'askUser' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'version' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'dockerInside' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'isDevelopMode' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'assistantUseAgents' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<SettingsFragmentFragment, unknown>;
export const TerminalFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'terminalFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Terminal' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'image' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'connected' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<TerminalFragmentFragment, unknown>;
export const ProviderFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ProviderFragmentFragment, unknown>;
export const FlowFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Flow' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'terminals' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'terminalFragment' } }],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'provider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'terminalFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Terminal' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'image' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'connected' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowFragmentFragment, unknown>;
export const SubtaskFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'subtaskFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Subtask' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<SubtaskFragmentFragment, unknown>;
export const TaskFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'taskFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Task' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'input' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'subtasks' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'subtaskFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'subtaskFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Subtask' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<TaskFragmentFragment, unknown>;
export const TerminalLogFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'terminalLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'TerminalLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'text' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'terminal' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<TerminalLogFragmentFragment, unknown>;
export const MessageLogFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'messageLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'MessageLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'message' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'thinking' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'resultFormat' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<MessageLogFragmentFragment, unknown>;
export const ScreenshotFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'screenshotFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Screenshot' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'url' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ScreenshotFragmentFragment, unknown>;
export const FlowFileFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowFileFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowFile' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'path' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'size' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'isDir' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'modifiedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowFileFragmentFragment, unknown>;
export const UserResourceFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'userResourceFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UserResource' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'path' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'size' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'isDir' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UserResourceFragmentFragment, unknown>;
export const AgentLogFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'initiator' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'executor' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'task' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AgentLogFragmentFragment, unknown>;
export const SearchLogFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'searchLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'SearchLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'initiator' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'executor' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'engine' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'query' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<SearchLogFragmentFragment, unknown>;
export const VectorStoreLogFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'vectorStoreLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'VectorStoreLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'initiator' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'executor' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'filter' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'query' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'action' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<VectorStoreLogFragmentFragment, unknown>;
export const AssistantFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'assistantFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Assistant' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'provider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'useAgents' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AssistantFragmentFragment, unknown>;
export const AssistantLogFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'assistantLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AssistantLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'message' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'thinking' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'resultFormat' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'appendPart' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'assistantId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AssistantLogFragmentFragment, unknown>;
export const TestResultFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'testResultFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'TestResult' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'reasoning' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'streaming' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'latency' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'error' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<TestResultFragmentFragment, unknown>;
export const AgentTestResultFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentTestResultFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentTestResult' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'tests' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'testResultFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'testResultFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'TestResult' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'reasoning' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'streaming' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'latency' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'error' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AgentTestResultFragmentFragment, unknown>;
export const ProviderTestResultFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerTestResultFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ProviderTestResult' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simple' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simpleJson' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'primaryAgent' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistant' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'generator' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'refiner' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'adviser' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reflector' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'searcher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'enricher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'coder' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'installer' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'pentester' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'testResultFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'TestResult' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'reasoning' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'streaming' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'latency' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'error' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentTestResultFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentTestResult' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'tests' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'testResultFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ProviderTestResultFragmentFragment, unknown>;
export const ModelConfigFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'modelConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ModelConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'thinking' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reasoning' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'mode' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'efforts' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'supported' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cannotDisable' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'defaultOn' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'price' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'input' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'output' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheRead' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheWrite' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ModelConfigFragmentFragment, unknown>;
export const AgentConfigFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'model' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'temperature' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topK' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topP' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'minLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'repetitionPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'frequencyPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'presencePenalty' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reasoning' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'mode' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'effort' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'price' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'input' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'output' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheRead' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheWrite' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AgentConfigFragmentFragment, unknown>;
export const AgentsConfigFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentsConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentsConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simple' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simpleJson' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'primaryAgent' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistant' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'generator' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'refiner' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'adviser' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reflector' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'searcher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'enricher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'coder' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'installer' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'pentester' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'model' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'temperature' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topK' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topP' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'minLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'repetitionPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'frequencyPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'presencePenalty' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reasoning' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'mode' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'effort' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'price' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'input' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'output' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheRead' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheWrite' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AgentsConfigFragmentFragment, unknown>;
export const ProviderConfigFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ProviderConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'agents' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentsConfigFragment' } },
                            ],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'model' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'temperature' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topK' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topP' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'minLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'repetitionPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'frequencyPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'presencePenalty' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reasoning' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'mode' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'effort' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'price' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'input' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'output' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheRead' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheWrite' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentsConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentsConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simple' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simpleJson' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'primaryAgent' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistant' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'generator' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'refiner' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'adviser' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reflector' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'searcher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'enricher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'coder' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'installer' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'pentester' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ProviderConfigFragmentFragment, unknown>;
export const UserPromptFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'userPromptFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UserPrompt' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'template' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UserPromptFragmentFragment, unknown>;
export const DefaultPromptFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'defaultPromptFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'DefaultPrompt' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'template' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'variables' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<DefaultPromptFragmentFragment, unknown>;
export const PromptValidationResultFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'promptValidationResultFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'PromptValidationResult' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'errorType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'message' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'line' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'details' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<PromptValidationResultFragmentFragment, unknown>;
export const ApiTokenFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'apiTokenFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'APIToken' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'tokenId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'roleId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'ttl' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ApiTokenFragmentFragment, unknown>;
export const ApiTokenWithSecretFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'apiTokenWithSecretFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'APITokenWithSecret' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'tokenId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'roleId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'ttl' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'token' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ApiTokenWithSecretFragmentFragment, unknown>;
export const FlowTemplateFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowTemplateFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowTemplate' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'text' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowTemplateFragmentFragment, unknown>;
export const UsageStatsFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'usageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostOut' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UsageStatsFragmentFragment, unknown>;
export const DailyUsageStatsFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'dailyUsageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'DailyUsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'date' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stats' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'usageStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'usageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostOut' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<DailyUsageStatsFragmentFragment, unknown>;
export const ProviderUsageStatsFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerUsageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ProviderUsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'provider' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stats' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'usageStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'usageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostOut' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ProviderUsageStatsFragmentFragment, unknown>;
export const ModelUsageStatsFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'modelUsageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ModelUsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'model' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'provider' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stats' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'usageStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'usageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostOut' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ModelUsageStatsFragmentFragment, unknown>;
export const AgentTypeUsageStatsFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentTypeUsageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentTypeUsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'agentType' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stats' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'usageStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'usageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostOut' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AgentTypeUsageStatsFragmentFragment, unknown>;
export const ModelAgentsUsageStatsFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'modelAgentsUsageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ModelAgentsUsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'model' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'provider' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'agentTypes' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stats' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'usageStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'usageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostOut' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ModelAgentsUsageStatsFragmentFragment, unknown>;
export const ToolcallsStatsFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'toolcallsStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ToolcallsStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ToolcallsStatsFragmentFragment, unknown>;
export const DailyToolcallsStatsFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'dailyToolcallsStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'DailyToolcallsStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'date' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stats' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'toolcallsStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'toolcallsStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ToolcallsStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<DailyToolcallsStatsFragmentFragment, unknown>;
export const FunctionToolcallsStatsFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'functionToolcallsStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FunctionToolcallsStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'functionName' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'isAgent' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'avgDurationSeconds' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FunctionToolcallsStatsFragmentFragment, unknown>;
export const FlowStatsFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalTasksCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSubtasksCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalAssistantsCount' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowStatsFragmentFragment, unknown>;
export const FlowsStatsFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowsStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowsStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalFlowsCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalTasksCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSubtasksCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalAssistantsCount' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowsStatsFragmentFragment, unknown>;
export const DailyFlowsStatsFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'dailyFlowsStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'DailyFlowsStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'date' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stats' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowsStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowsStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowsStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalFlowsCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalTasksCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSubtasksCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalAssistantsCount' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<DailyFlowsStatsFragmentFragment, unknown>;
export const SubtaskExecutionStatsFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'subtaskExecutionStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'SubtaskExecutionStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskTitle' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalToolcallsCount' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<SubtaskExecutionStatsFragmentFragment, unknown>;
export const TaskExecutionStatsFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'taskExecutionStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'TaskExecutionStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskTitle' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalToolcallsCount' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'subtasks' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                {
                                    kind: 'FragmentSpread',
                                    name: { kind: 'Name', value: 'subtaskExecutionStatsFragment' },
                                },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'subtaskExecutionStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'SubtaskExecutionStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskTitle' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalToolcallsCount' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<TaskExecutionStatsFragmentFragment, unknown>;
export const FlowExecutionStatsFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowExecutionStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowExecutionStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowTitle' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalToolcallsCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalAssistantsCount' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'tasks' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'taskExecutionStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'subtaskExecutionStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'SubtaskExecutionStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskTitle' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalToolcallsCount' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'taskExecutionStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'TaskExecutionStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskTitle' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalToolcallsCount' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'subtasks' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                {
                                    kind: 'FragmentSpread',
                                    name: { kind: 'Name', value: 'subtaskExecutionStatsFragment' },
                                },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowExecutionStatsFragmentFragment, unknown>;
export const KnowledgeDocumentFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'knowledgeDocumentFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'KnowledgeDocument' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'docType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'content' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'question' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'guideType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'answerType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'codeLang' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'partSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'manual' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<KnowledgeDocumentFragmentFragment, unknown>;
export const KnowledgeDocumentWithScoreFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'knowledgeDocumentWithScoreFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'KnowledgeDocumentWithScore' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'score' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'document' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'knowledgeDocumentFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'knowledgeDocumentFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'KnowledgeDocument' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'docType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'content' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'question' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'guideType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'answerType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'codeLang' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'partSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'manual' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<KnowledgeDocumentWithScoreFragmentFragment, unknown>;
export const UserPreferencesFragmentFragmentDoc = {
    kind: 'Document',
    definitions: [
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'userPreferencesFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UserPreferences' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'favoriteFlows' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UserPreferencesFragmentFragment, unknown>;
export const FlowsDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'flows' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flows' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'terminalFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Terminal' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'image' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'connected' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Flow' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'terminals' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'terminalFragment' } }],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'provider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowsQuery, FlowsQueryVariables>;
export const ProvidersDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'providers' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'providers' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ProvidersQuery, ProvidersQueryVariables>;
export const SettingsDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'settings' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'settings' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'settingsFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'settingsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Settings' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'debug' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'askUser' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'version' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'dockerInside' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'isDevelopMode' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'assistantUseAgents' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<SettingsQuery, SettingsQueryVariables>;
export const SettingsProvidersDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'settingsProviders' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'settingsProviders' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                {
                                    kind: 'Field',
                                    name: { kind: 'Name', value: 'enabled' },
                                    selectionSet: {
                                        kind: 'SelectionSet',
                                        selections: [
                                            { kind: 'Field', name: { kind: 'Name', value: 'openai' } },
                                            { kind: 'Field', name: { kind: 'Name', value: 'anthropic' } },
                                            { kind: 'Field', name: { kind: 'Name', value: 'gemini' } },
                                            { kind: 'Field', name: { kind: 'Name', value: 'bedrock' } },
                                            { kind: 'Field', name: { kind: 'Name', value: 'ollama' } },
                                            { kind: 'Field', name: { kind: 'Name', value: 'custom' } },
                                            { kind: 'Field', name: { kind: 'Name', value: 'deepseek' } },
                                            { kind: 'Field', name: { kind: 'Name', value: 'glm' } },
                                            { kind: 'Field', name: { kind: 'Name', value: 'kimi' } },
                                            { kind: 'Field', name: { kind: 'Name', value: 'qwen' } },
                                            { kind: 'Field', name: { kind: 'Name', value: 'minimax' } },
                                        ],
                                    },
                                },
                                {
                                    kind: 'Field',
                                    name: { kind: 'Name', value: 'default' },
                                    selectionSet: {
                                        kind: 'SelectionSet',
                                        selections: [
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'openai' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'providerConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'anthropic' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'providerConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'gemini' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'providerConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'bedrock' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'providerConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'ollama' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'providerConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'custom' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'providerConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'deepseek' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'providerConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'glm' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'providerConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'kimi' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'providerConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'qwen' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'providerConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'minimax' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'providerConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                        ],
                                    },
                                },
                                {
                                    kind: 'Field',
                                    name: { kind: 'Name', value: 'userDefined' },
                                    selectionSet: {
                                        kind: 'SelectionSet',
                                        selections: [
                                            {
                                                kind: 'FragmentSpread',
                                                name: { kind: 'Name', value: 'providerConfigFragment' },
                                            },
                                        ],
                                    },
                                },
                                {
                                    kind: 'Field',
                                    name: { kind: 'Name', value: 'models' },
                                    selectionSet: {
                                        kind: 'SelectionSet',
                                        selections: [
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'openai' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'modelConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'anthropic' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'modelConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'gemini' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'modelConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'bedrock' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'modelConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'ollama' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'modelConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'custom' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'modelConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'deepseek' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'modelConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'glm' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'modelConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'kimi' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'modelConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'qwen' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'modelConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'minimax' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'FragmentSpread',
                                                            name: { kind: 'Name', value: 'modelConfigFragment' },
                                                        },
                                                    ],
                                                },
                                            },
                                        ],
                                    },
                                },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'model' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'temperature' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topK' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topP' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'minLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'repetitionPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'frequencyPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'presencePenalty' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reasoning' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'mode' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'effort' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'price' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'input' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'output' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheRead' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheWrite' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentsConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentsConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simple' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simpleJson' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'primaryAgent' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistant' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'generator' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'refiner' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'adviser' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reflector' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'searcher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'enricher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'coder' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'installer' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'pentester' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ProviderConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'agents' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentsConfigFragment' } },
                            ],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'modelConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ModelConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'thinking' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reasoning' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'mode' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'efforts' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'supported' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cannotDisable' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'defaultOn' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'price' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'input' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'output' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheRead' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheWrite' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<SettingsProvidersQuery, SettingsProvidersQueryVariables>;
export const SettingsPromptsDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'settingsPrompts' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'settingsPrompts' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                {
                                    kind: 'Field',
                                    name: { kind: 'Name', value: 'default' },
                                    selectionSet: {
                                        kind: 'SelectionSet',
                                        selections: [
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'agents' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'primaryAgent' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'system' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'assistant' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'system' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'pentester' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'system' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'human' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'coder' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'system' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'human' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'installer' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'system' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'human' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'searcher' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'system' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'human' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'memorist' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'system' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'human' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'adviser' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'system' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'human' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'generator' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'system' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'human' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'refiner' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'system' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'human' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'reporter' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'system' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'human' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'reflector' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'system' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'human' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'enricher' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'system' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'human' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'toolCallFixer' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'system' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'human' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'summarizer' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'Field',
                                                                        name: { kind: 'Name', value: 'system' },
                                                                        selectionSet: {
                                                                            kind: 'SelectionSet',
                                                                            selections: [
                                                                                {
                                                                                    kind: 'FragmentSpread',
                                                                                    name: {
                                                                                        kind: 'Name',
                                                                                        value: 'defaultPromptFragment',
                                                                                    },
                                                                                },
                                                                            ],
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                    ],
                                                },
                                            },
                                            {
                                                kind: 'Field',
                                                name: { kind: 'Name', value: 'tools' },
                                                selectionSet: {
                                                    kind: 'SelectionSet',
                                                    selections: [
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'getFlowDescription' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'FragmentSpread',
                                                                        name: {
                                                                            kind: 'Name',
                                                                            value: 'defaultPromptFragment',
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'getTaskDescription' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'FragmentSpread',
                                                                        name: {
                                                                            kind: 'Name',
                                                                            value: 'defaultPromptFragment',
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'getExecutionLogs' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'FragmentSpread',
                                                                        name: {
                                                                            kind: 'Name',
                                                                            value: 'defaultPromptFragment',
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'getFullExecutionContext' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'FragmentSpread',
                                                                        name: {
                                                                            kind: 'Name',
                                                                            value: 'defaultPromptFragment',
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'getShortExecutionContext' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'FragmentSpread',
                                                                        name: {
                                                                            kind: 'Name',
                                                                            value: 'defaultPromptFragment',
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'chooseDockerImage' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'FragmentSpread',
                                                                        name: {
                                                                            kind: 'Name',
                                                                            value: 'defaultPromptFragment',
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'chooseUserLanguage' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'FragmentSpread',
                                                                        name: {
                                                                            kind: 'Name',
                                                                            value: 'defaultPromptFragment',
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'collectToolCallId' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'FragmentSpread',
                                                                        name: {
                                                                            kind: 'Name',
                                                                            value: 'defaultPromptFragment',
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'detectToolCallIdPattern' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'FragmentSpread',
                                                                        name: {
                                                                            kind: 'Name',
                                                                            value: 'defaultPromptFragment',
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'monitorAgentExecution' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'FragmentSpread',
                                                                        name: {
                                                                            kind: 'Name',
                                                                            value: 'defaultPromptFragment',
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'planAgentTask' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'FragmentSpread',
                                                                        name: {
                                                                            kind: 'Name',
                                                                            value: 'defaultPromptFragment',
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                        {
                                                            kind: 'Field',
                                                            name: { kind: 'Name', value: 'wrapAgentTask' },
                                                            selectionSet: {
                                                                kind: 'SelectionSet',
                                                                selections: [
                                                                    {
                                                                        kind: 'FragmentSpread',
                                                                        name: {
                                                                            kind: 'Name',
                                                                            value: 'defaultPromptFragment',
                                                                        },
                                                                    },
                                                                ],
                                                            },
                                                        },
                                                    ],
                                                },
                                            },
                                        ],
                                    },
                                },
                                {
                                    kind: 'Field',
                                    name: { kind: 'Name', value: 'userDefined' },
                                    selectionSet: {
                                        kind: 'SelectionSet',
                                        selections: [
                                            {
                                                kind: 'FragmentSpread',
                                                name: { kind: 'Name', value: 'userPromptFragment' },
                                            },
                                        ],
                                    },
                                },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'defaultPromptFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'DefaultPrompt' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'template' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'variables' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'userPromptFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UserPrompt' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'template' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<SettingsPromptsQuery, SettingsPromptsQueryVariables>;
export const FlowDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'flow' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flow' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowFragment' } }],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'tasks' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'taskFragment' } }],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'screenshots' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'screenshotFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'terminalLogs' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'terminalLogFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'messageLogs' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'messageLogFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'agentLogs' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentLogFragment' } }],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'searchLogs' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'searchLogFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'vectorStoreLogs' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'vectorStoreLogFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'terminalFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Terminal' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'image' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'connected' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'subtaskFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Subtask' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Flow' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'terminals' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'terminalFragment' } }],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'provider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'taskFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Task' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'input' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'subtasks' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'subtaskFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'screenshotFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Screenshot' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'url' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'terminalLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'TerminalLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'text' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'terminal' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'messageLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'MessageLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'message' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'thinking' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'resultFormat' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'initiator' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'executor' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'task' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'searchLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'SearchLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'initiator' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'executor' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'engine' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'query' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'vectorStoreLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'VectorStoreLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'initiator' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'executor' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'filter' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'query' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'action' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowQuery, FlowQueryVariables>;
export const TasksDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'tasks' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'tasks' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'taskFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'subtaskFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Subtask' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'taskFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Task' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'input' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'subtasks' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'subtaskFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<TasksQuery, TasksQueryVariables>;
export const FlowFilesDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'flowFiles' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flowFiles' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowFileFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowFileFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowFile' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'path' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'size' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'isDir' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'modifiedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowFilesQuery, FlowFilesQueryVariables>;
export const ResourcesDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'resources' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'path' } },
                    type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'recursive' } },
                    type: { kind: 'NamedType', name: { kind: 'Name', value: 'Boolean' } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'resources' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'path' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'path' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'recursive' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'recursive' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'userResourceFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'userResourceFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UserResource' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'path' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'size' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'isDir' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ResourcesQuery, ResourcesQueryVariables>;
export const AssistantsDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'assistants' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistants' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'assistantFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'assistantFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Assistant' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'provider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'useAgents' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AssistantsQuery, AssistantsQueryVariables>;
export const AssistantLogsDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'assistantLogs' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'assistantId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistantLogs' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'assistantId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'assistantId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'assistantLogFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'assistantLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AssistantLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'message' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'thinking' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'resultFormat' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'appendPart' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'assistantId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AssistantLogsQuery, AssistantLogsQueryVariables>;
export const FlowReportDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'flowReport' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flow' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowFragment' } }],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'tasks' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'taskFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'terminalFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Terminal' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'image' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'connected' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'subtaskFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Subtask' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Flow' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'terminals' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'terminalFragment' } }],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'provider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'taskFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Task' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'input' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'subtasks' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'subtaskFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowReportQuery, FlowReportQueryVariables>;
export const UsageStatsTotalDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'usageStatsTotal' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'usageStatsTotal' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'usageStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'usageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostOut' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UsageStatsTotalQuery, UsageStatsTotalQueryVariables>;
export const UsageStatsByPeriodDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'usageStatsByPeriod' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'period' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStatsPeriod' } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'usageStatsByPeriod' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'period' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'period' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'dailyUsageStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'usageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostOut' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'dailyUsageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'DailyUsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'date' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stats' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'usageStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UsageStatsByPeriodQuery, UsageStatsByPeriodQueryVariables>;
export const UsageStatsByProviderDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'usageStatsByProvider' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'usageStatsByProvider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerUsageStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'usageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostOut' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerUsageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ProviderUsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'provider' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stats' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'usageStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UsageStatsByProviderQuery, UsageStatsByProviderQueryVariables>;
export const UsageStatsByModelDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'usageStatsByModel' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'usageStatsByModel' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'modelUsageStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'usageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostOut' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'modelUsageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ModelUsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'model' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'provider' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stats' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'usageStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UsageStatsByModelQuery, UsageStatsByModelQueryVariables>;
export const UsageStatsByAgentTypeDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'usageStatsByAgentType' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'usageStatsByAgentType' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                {
                                    kind: 'FragmentSpread',
                                    name: { kind: 'Name', value: 'agentTypeUsageStatsFragment' },
                                },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'usageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostOut' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentTypeUsageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentTypeUsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'agentType' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stats' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'usageStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UsageStatsByAgentTypeQuery, UsageStatsByAgentTypeQueryVariables>;
export const UsageStatsByFlowDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'usageStatsByFlow' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'usageStatsByFlow' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'usageStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'usageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostOut' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UsageStatsByFlowQuery, UsageStatsByFlowQueryVariables>;
export const UsageStatsByAgentTypeForFlowDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'usageStatsByAgentTypeForFlow' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'usageStatsByAgentTypeForFlow' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                {
                                    kind: 'FragmentSpread',
                                    name: { kind: 'Name', value: 'agentTypeUsageStatsFragment' },
                                },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'usageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostOut' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentTypeUsageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentTypeUsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'agentType' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stats' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'usageStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UsageStatsByAgentTypeForFlowQuery, UsageStatsByAgentTypeForFlowQueryVariables>;
export const UsageStatsByModelAgentsForFlowDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'usageStatsByModelAgentsForFlow' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'usageStatsByModelAgentsForFlow' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                {
                                    kind: 'FragmentSpread',
                                    name: { kind: 'Name', value: 'modelAgentsUsageStatsFragment' },
                                },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'usageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCacheOut' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostIn' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalUsageCostOut' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'modelAgentsUsageStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ModelAgentsUsageStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'model' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'provider' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'agentTypes' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stats' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'usageStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UsageStatsByModelAgentsForFlowQuery, UsageStatsByModelAgentsForFlowQueryVariables>;
export const ToolcallsStatsTotalDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'toolcallsStatsTotal' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'toolcallsStatsTotal' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'toolcallsStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'toolcallsStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ToolcallsStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ToolcallsStatsTotalQuery, ToolcallsStatsTotalQueryVariables>;
export const ToolcallsStatsByPeriodDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'toolcallsStatsByPeriod' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'period' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStatsPeriod' } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'toolcallsStatsByPeriod' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'period' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'period' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                {
                                    kind: 'FragmentSpread',
                                    name: { kind: 'Name', value: 'dailyToolcallsStatsFragment' },
                                },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'toolcallsStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ToolcallsStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'dailyToolcallsStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'DailyToolcallsStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'date' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stats' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'toolcallsStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ToolcallsStatsByPeriodQuery, ToolcallsStatsByPeriodQueryVariables>;
export const ToolcallsStatsByFunctionDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'toolcallsStatsByFunction' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'toolcallsStatsByFunction' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                {
                                    kind: 'FragmentSpread',
                                    name: { kind: 'Name', value: 'functionToolcallsStatsFragment' },
                                },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'functionToolcallsStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FunctionToolcallsStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'functionName' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'isAgent' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'avgDurationSeconds' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ToolcallsStatsByFunctionQuery, ToolcallsStatsByFunctionQueryVariables>;
export const ToolcallsStatsByFlowDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'toolcallsStatsByFlow' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'toolcallsStatsByFlow' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'toolcallsStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'toolcallsStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ToolcallsStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ToolcallsStatsByFlowQuery, ToolcallsStatsByFlowQueryVariables>;
export const ToolcallsStatsByFunctionForFlowDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'toolcallsStatsByFunctionForFlow' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'toolcallsStatsByFunctionForFlow' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                {
                                    kind: 'FragmentSpread',
                                    name: { kind: 'Name', value: 'functionToolcallsStatsFragment' },
                                },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'functionToolcallsStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FunctionToolcallsStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'functionName' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'isAgent' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'avgDurationSeconds' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ToolcallsStatsByFunctionForFlowQuery, ToolcallsStatsByFunctionForFlowQueryVariables>;
export const FlowsStatsTotalDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'flowsStatsTotal' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flowsStatsTotal' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowsStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowsStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowsStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalFlowsCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalTasksCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSubtasksCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalAssistantsCount' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowsStatsTotalQuery, FlowsStatsTotalQueryVariables>;
export const FlowsStatsByPeriodDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'flowsStatsByPeriod' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'period' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStatsPeriod' } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flowsStatsByPeriod' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'period' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'period' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'dailyFlowsStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowsStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowsStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalFlowsCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalTasksCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSubtasksCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalAssistantsCount' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'dailyFlowsStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'DailyFlowsStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'date' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stats' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowsStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowsStatsByPeriodQuery, FlowsStatsByPeriodQueryVariables>;
export const FlowStatsByFlowDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'flowStatsByFlow' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flowStatsByFlow' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'totalTasksCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSubtasksCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalAssistantsCount' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowStatsByFlowQuery, FlowStatsByFlowQueryVariables>;
export const FlowsExecutionStatsByPeriodDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'flowsExecutionStatsByPeriod' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'period' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'UsageStatsPeriod' } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flowsExecutionStatsByPeriod' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'period' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'period' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowExecutionStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'subtaskExecutionStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'SubtaskExecutionStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskTitle' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalToolcallsCount' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'taskExecutionStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'TaskExecutionStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskTitle' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalToolcallsCount' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'subtasks' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                {
                                    kind: 'FragmentSpread',
                                    name: { kind: 'Name', value: 'subtaskExecutionStatsFragment' },
                                },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowExecutionStatsFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowExecutionStats' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowTitle' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalDurationSeconds' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalToolcallsCount' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalAssistantsCount' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'tasks' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'taskExecutionStatsFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowsExecutionStatsByPeriodQuery, FlowsExecutionStatsByPeriodQueryVariables>;
export const ApiTokensDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'apiTokens' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'apiTokens' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'apiTokenFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'apiTokenFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'APIToken' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'tokenId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'roleId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'ttl' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ApiTokensQuery, ApiTokensQueryVariables>;
export const ApiTokenDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'apiToken' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'tokenId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'apiToken' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'tokenId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'tokenId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'apiTokenFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'apiTokenFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'APIToken' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'tokenId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'roleId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'ttl' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ApiTokenQuery, ApiTokenQueryVariables>;
export const KnowledgeDocumentsDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'knowledgeDocuments' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'filter' } },
                    type: { kind: 'NamedType', name: { kind: 'Name', value: 'KnowledgeFilter' } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'withContent' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'Boolean' } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'knowledgeDocuments' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'filter' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'filter' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'withContent' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'withContent' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'knowledgeDocumentFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'knowledgeDocumentFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'KnowledgeDocument' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'docType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'content' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'question' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'guideType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'answerType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'codeLang' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'partSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'manual' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<KnowledgeDocumentsQuery, KnowledgeDocumentsQueryVariables>;
export const KnowledgeDocumentDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'knowledgeDocument' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'knowledgeDocument' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'id' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'knowledgeDocumentFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'knowledgeDocumentFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'KnowledgeDocument' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'docType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'content' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'question' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'guideType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'answerType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'codeLang' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'partSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'manual' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<KnowledgeDocumentQuery, KnowledgeDocumentQueryVariables>;
export const SearchKnowledgeDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'searchKnowledge' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'query' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'filter' } },
                    type: { kind: 'NamedType', name: { kind: 'Name', value: 'KnowledgeFilter' } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'limit' } },
                    type: { kind: 'NamedType', name: { kind: 'Name', value: 'Int' } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'searchKnowledge' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'query' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'query' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'filter' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'filter' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'limit' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'limit' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                {
                                    kind: 'FragmentSpread',
                                    name: { kind: 'Name', value: 'knowledgeDocumentWithScoreFragment' },
                                },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'knowledgeDocumentFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'KnowledgeDocument' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'docType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'content' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'question' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'guideType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'answerType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'codeLang' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'partSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'manual' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'knowledgeDocumentWithScoreFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'KnowledgeDocumentWithScore' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'score' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'document' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'knowledgeDocumentFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<SearchKnowledgeQuery, SearchKnowledgeQueryVariables>;
export const SettingsUserDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'settingsUser' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'settingsUser' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'userPreferencesFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'userPreferencesFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UserPreferences' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'favoriteFlows' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<SettingsUserQuery, SettingsUserQueryVariables>;
export const AddFavoriteFlowDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'addFavoriteFlow' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'addFavoriteFlow' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AddFavoriteFlowMutation, AddFavoriteFlowMutationVariables>;
export const DeleteFavoriteFlowDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'deleteFavoriteFlow' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'deleteFavoriteFlow' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<DeleteFavoriteFlowMutation, DeleteFavoriteFlowMutationVariables>;
export const AnonymizeTextDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'anonymizeText' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'text' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'anonymizeText' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'text' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'text' } },
                            },
                        ],
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AnonymizeTextMutation, AnonymizeTextMutationVariables>;
export const FlowTemplatesDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'flowTemplates' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flowTemplates' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowTemplateFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowTemplateFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowTemplate' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'text' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowTemplatesQuery, FlowTemplatesQueryVariables>;
export const FlowTemplateDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'query',
            name: { kind: 'Name', value: 'flowTemplate' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'templateId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flowTemplate' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'templateId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'templateId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowTemplateFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowTemplateFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowTemplate' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'text' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowTemplateQuery, FlowTemplateQueryVariables>;
export const CreateFlowTemplateDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'createFlowTemplate' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'CreateFlowTemplateInput' } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'createFlowTemplate' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'input' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowTemplateFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowTemplateFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowTemplate' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'text' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<CreateFlowTemplateMutation, CreateFlowTemplateMutationVariables>;
export const UpdateFlowTemplateDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'updateFlowTemplate' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'templateId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'UpdateFlowTemplateInput' } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'updateFlowTemplate' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'templateId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'templateId' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'input' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowTemplateFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowTemplateFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowTemplate' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'text' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UpdateFlowTemplateMutation, UpdateFlowTemplateMutationVariables>;
export const DeleteFlowTemplateDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'deleteFlowTemplate' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'templateId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'deleteFlowTemplate' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'templateId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'templateId' } },
                            },
                        ],
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<DeleteFlowTemplateMutation, DeleteFlowTemplateMutationVariables>;
export const CreateFlowDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'createFlow' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'modelProvider' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'resourceIds' } },
                    type: {
                        kind: 'ListType',
                        type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'createFlow' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'modelProvider' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'modelProvider' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'input' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'resourceIds' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'resourceIds' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'terminalFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Terminal' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'image' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'connected' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Flow' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'terminals' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'terminalFragment' } }],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'provider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<CreateFlowMutation, CreateFlowMutationVariables>;
export const DeleteFlowDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'deleteFlow' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'deleteFlow' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<DeleteFlowMutation, DeleteFlowMutationVariables>;
export const PutUserInputDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'putUserInput' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'modelProvider' } },
                    type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'resourceIds' } },
                    type: {
                        kind: 'ListType',
                        type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'putUserInput' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'input' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'modelProvider' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'modelProvider' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'resourceIds' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'resourceIds' } },
                            },
                        ],
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<PutUserInputMutation, PutUserInputMutationVariables>;
export const FinishFlowDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'finishFlow' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'finishFlow' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FinishFlowMutation, FinishFlowMutationVariables>;
export const StopFlowDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'stopFlow' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stopFlow' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<StopFlowMutation, StopFlowMutationVariables>;
export const RenameFlowDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'renameFlow' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'title' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'renameFlow' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'title' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'title' } },
                            },
                        ],
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<RenameFlowMutation, RenameFlowMutationVariables>;
export const CreateAssistantDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'createAssistant' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'modelProvider' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'useAgents' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'Boolean' } },
                    },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'resourceIds' } },
                    type: {
                        kind: 'ListType',
                        type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'createAssistant' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'modelProvider' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'modelProvider' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'input' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'useAgents' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'useAgents' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'resourceIds' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'resourceIds' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                {
                                    kind: 'Field',
                                    name: { kind: 'Name', value: 'flow' },
                                    selectionSet: {
                                        kind: 'SelectionSet',
                                        selections: [
                                            { kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowFragment' } },
                                        ],
                                    },
                                },
                                {
                                    kind: 'Field',
                                    name: { kind: 'Name', value: 'assistant' },
                                    selectionSet: {
                                        kind: 'SelectionSet',
                                        selections: [
                                            {
                                                kind: 'FragmentSpread',
                                                name: { kind: 'Name', value: 'assistantFragment' },
                                            },
                                        ],
                                    },
                                },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'terminalFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Terminal' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'image' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'connected' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Flow' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'terminals' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'terminalFragment' } }],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'provider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'assistantFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Assistant' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'provider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'useAgents' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<CreateAssistantMutation, CreateAssistantMutationVariables>;
export const CallAssistantDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'callAssistant' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'assistantId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'useAgents' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'Boolean' } },
                    },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'resourceIds' } },
                    type: {
                        kind: 'ListType',
                        type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'callAssistant' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'assistantId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'assistantId' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'input' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'useAgents' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'useAgents' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'resourceIds' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'resourceIds' } },
                            },
                        ],
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<CallAssistantMutation, CallAssistantMutationVariables>;
export const StopAssistantDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'stopAssistant' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'assistantId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'stopAssistant' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'assistantId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'assistantId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'assistantFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'assistantFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Assistant' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'provider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'useAgents' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<StopAssistantMutation, StopAssistantMutationVariables>;
export const DeleteAssistantDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'deleteAssistant' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'assistantId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'deleteAssistant' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'assistantId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'assistantId' } },
                            },
                        ],
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<DeleteAssistantMutation, DeleteAssistantMutationVariables>;
export const TestAgentDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'testAgent' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'type' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'ProviderType' } },
                    },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'agentType' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentConfigType' } },
                    },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'agent' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentConfigInput' } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'testAgent' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'type' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'type' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'agentType' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'agentType' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'agent' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'agent' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'testResultFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'TestResult' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'reasoning' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'streaming' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'latency' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'error' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentTestResultFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentTestResult' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'tests' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'testResultFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<TestAgentMutation, TestAgentMutationVariables>;
export const TestProviderDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'testProvider' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'type' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'ProviderType' } },
                    },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'agents' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentsConfigInput' } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'testProvider' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'type' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'type' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'agents' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'agents' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerTestResultFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'testResultFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'TestResult' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'reasoning' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'streaming' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'latency' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'error' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentTestResultFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentTestResult' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'tests' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'testResultFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerTestResultFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ProviderTestResult' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simple' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simpleJson' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'primaryAgent' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistant' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'generator' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'refiner' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'adviser' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reflector' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'searcher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'enricher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'coder' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'installer' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'pentester' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentTestResultFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<TestProviderMutation, TestProviderMutationVariables>;
export const CreateProviderDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'createProvider' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'name' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'type' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'ProviderType' } },
                    },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'agents' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentsConfigInput' } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'createProvider' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'name' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'name' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'type' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'type' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'agents' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'agents' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerConfigFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'model' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'temperature' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topK' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topP' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'minLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'repetitionPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'frequencyPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'presencePenalty' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reasoning' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'mode' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'effort' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'price' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'input' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'output' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheRead' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheWrite' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentsConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentsConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simple' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simpleJson' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'primaryAgent' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistant' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'generator' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'refiner' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'adviser' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reflector' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'searcher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'enricher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'coder' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'installer' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'pentester' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ProviderConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'agents' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentsConfigFragment' } },
                            ],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<CreateProviderMutation, CreateProviderMutationVariables>;
export const UpdateProviderDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'updateProvider' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'providerId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'name' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'agents' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentsConfigInput' } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'updateProvider' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'providerId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'providerId' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'name' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'name' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'agents' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'agents' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerConfigFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'model' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'temperature' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topK' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topP' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'minLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'repetitionPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'frequencyPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'presencePenalty' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reasoning' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'mode' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'effort' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'price' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'input' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'output' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheRead' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheWrite' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentsConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentsConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simple' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simpleJson' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'primaryAgent' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistant' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'generator' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'refiner' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'adviser' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reflector' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'searcher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'enricher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'coder' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'installer' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'pentester' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ProviderConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'agents' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentsConfigFragment' } },
                            ],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UpdateProviderMutation, UpdateProviderMutationVariables>;
export const DeleteProviderDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'deleteProvider' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'providerId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'deleteProvider' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'providerId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'providerId' } },
                            },
                        ],
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<DeleteProviderMutation, DeleteProviderMutationVariables>;
export const ValidatePromptDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'validatePrompt' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'type' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'PromptType' } },
                    },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'template' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'validatePrompt' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'type' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'type' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'template' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'template' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                {
                                    kind: 'FragmentSpread',
                                    name: { kind: 'Name', value: 'promptValidationResultFragment' },
                                },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'promptValidationResultFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'PromptValidationResult' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'errorType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'message' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'line' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'details' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ValidatePromptMutation, ValidatePromptMutationVariables>;
export const CreatePromptDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'createPrompt' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'type' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'PromptType' } },
                    },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'template' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'createPrompt' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'type' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'type' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'template' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'template' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'userPromptFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'userPromptFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UserPrompt' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'template' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<CreatePromptMutation, CreatePromptMutationVariables>;
export const UpdatePromptDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'updatePrompt' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'promptId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'template' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'updatePrompt' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'promptId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'promptId' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'template' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'template' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'userPromptFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'userPromptFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UserPrompt' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'template' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UpdatePromptMutation, UpdatePromptMutationVariables>;
export const DeletePromptDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'deletePrompt' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'promptId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'deletePrompt' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'promptId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'promptId' } },
                            },
                        ],
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<DeletePromptMutation, DeletePromptMutationVariables>;
export const CreateApiTokenDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'createAPIToken' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'CreateAPITokenInput' } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'createAPIToken' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'input' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'apiTokenWithSecretFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'apiTokenWithSecretFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'APITokenWithSecret' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'tokenId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'roleId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'ttl' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'token' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<CreateApiTokenMutation, CreateApiTokenMutationVariables>;
export const UpdateApiTokenDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'updateAPIToken' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'tokenId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'UpdateAPITokenInput' } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'updateAPIToken' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'tokenId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'tokenId' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'input' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'apiTokenFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'apiTokenFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'APIToken' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'tokenId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'roleId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'ttl' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UpdateApiTokenMutation, UpdateApiTokenMutationVariables>;
export const DeleteApiTokenDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'deleteAPIToken' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'tokenId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'deleteAPIToken' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'tokenId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'tokenId' } },
                            },
                        ],
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<DeleteApiTokenMutation, DeleteApiTokenMutationVariables>;
export const CreateKnowledgeDocumentDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'createKnowledgeDocument' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'CreateKnowledgeDocumentInput' } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'createKnowledgeDocument' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'input' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'knowledgeDocumentFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'knowledgeDocumentFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'KnowledgeDocument' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'docType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'content' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'question' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'guideType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'answerType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'codeLang' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'partSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'manual' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<CreateKnowledgeDocumentMutation, CreateKnowledgeDocumentMutationVariables>;
export const UpdateKnowledgeDocumentDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'updateKnowledgeDocument' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                    type: {
                        kind: 'NonNullType',
                        type: { kind: 'NamedType', name: { kind: 'Name', value: 'UpdateKnowledgeDocumentInput' } },
                    },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'updateKnowledgeDocument' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'id' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'input' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'input' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'knowledgeDocumentFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'knowledgeDocumentFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'KnowledgeDocument' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'docType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'content' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'question' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'guideType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'answerType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'codeLang' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'partSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'manual' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<UpdateKnowledgeDocumentMutation, UpdateKnowledgeDocumentMutationVariables>;
export const RenameKnowledgeDocumentDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'renameKnowledgeDocument' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'question' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'renameKnowledgeDocument' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'id' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                            },
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'question' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'question' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'knowledgeDocumentFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'knowledgeDocumentFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'KnowledgeDocument' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'docType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'content' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'question' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'guideType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'answerType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'codeLang' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'partSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'manual' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<RenameKnowledgeDocumentMutation, RenameKnowledgeDocumentMutationVariables>;
export const DeleteKnowledgeDocumentDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'mutation',
            name: { kind: 'Name', value: 'deleteKnowledgeDocument' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'String' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'deleteKnowledgeDocument' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'id' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'id' } },
                            },
                        ],
                    },
                ],
            },
        },
    ],
} as unknown as DocumentNode<DeleteKnowledgeDocumentMutation, DeleteKnowledgeDocumentMutationVariables>;
export const TerminalLogAddedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'terminalLogAdded' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'terminalLogAdded' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'terminalLogFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'terminalLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'TerminalLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'text' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'terminal' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<TerminalLogAddedSubscription, TerminalLogAddedSubscriptionVariables>;
export const MessageLogAddedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'messageLogAdded' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'messageLogAdded' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'messageLogFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'messageLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'MessageLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'message' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'thinking' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'resultFormat' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<MessageLogAddedSubscription, MessageLogAddedSubscriptionVariables>;
export const MessageLogUpdatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'messageLogUpdated' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'messageLogUpdated' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'messageLogFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'messageLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'MessageLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'message' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'thinking' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'resultFormat' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<MessageLogUpdatedSubscription, MessageLogUpdatedSubscriptionVariables>;
export const ScreenshotAddedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'screenshotAdded' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'screenshotAdded' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'screenshotFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'screenshotFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Screenshot' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'url' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ScreenshotAddedSubscription, ScreenshotAddedSubscriptionVariables>;
export const AgentLogAddedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'agentLogAdded' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'agentLogAdded' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentLogFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'initiator' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'executor' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'task' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AgentLogAddedSubscription, AgentLogAddedSubscriptionVariables>;
export const SearchLogAddedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'searchLogAdded' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'searchLogAdded' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'searchLogFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'searchLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'SearchLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'initiator' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'executor' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'engine' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'query' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<SearchLogAddedSubscription, SearchLogAddedSubscriptionVariables>;
export const VectorStoreLogAddedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'vectorStoreLogAdded' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'vectorStoreLogAdded' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'vectorStoreLogFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'vectorStoreLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'VectorStoreLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'initiator' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'executor' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'filter' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'query' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'action' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<VectorStoreLogAddedSubscription, VectorStoreLogAddedSubscriptionVariables>;
export const AssistantCreatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'assistantCreated' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistantCreated' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'assistantFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'assistantFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Assistant' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'provider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'useAgents' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AssistantCreatedSubscription, AssistantCreatedSubscriptionVariables>;
export const AssistantUpdatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'assistantUpdated' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistantUpdated' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'assistantFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'assistantFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Assistant' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'provider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'useAgents' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AssistantUpdatedSubscription, AssistantUpdatedSubscriptionVariables>;
export const AssistantDeletedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'assistantDeleted' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistantDeleted' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'assistantFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'assistantFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Assistant' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'provider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'useAgents' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AssistantDeletedSubscription, AssistantDeletedSubscriptionVariables>;
export const FlowFileAddedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'flowFileAdded' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flowFileAdded' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowFileFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowFileFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowFile' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'path' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'size' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'isDir' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'modifiedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowFileAddedSubscription, FlowFileAddedSubscriptionVariables>;
export const FlowFileUpdatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'flowFileUpdated' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flowFileUpdated' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowFileFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowFileFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowFile' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'path' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'size' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'isDir' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'modifiedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowFileUpdatedSubscription, FlowFileUpdatedSubscriptionVariables>;
export const FlowFileDeletedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'flowFileDeleted' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flowFileDeleted' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowFileFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowFileFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowFile' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'path' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'size' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'isDir' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'modifiedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowFileDeletedSubscription, FlowFileDeletedSubscriptionVariables>;
export const AssistantLogAddedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'assistantLogAdded' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistantLogAdded' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'assistantLogFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'assistantLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AssistantLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'message' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'thinking' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'resultFormat' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'appendPart' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'assistantId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AssistantLogAddedSubscription, AssistantLogAddedSubscriptionVariables>;
export const AssistantLogUpdatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'assistantLogUpdated' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistantLogUpdated' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'assistantLogFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'assistantLogFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AssistantLog' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'message' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'thinking' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'resultFormat' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'appendPart' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'assistantId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<AssistantLogUpdatedSubscription, AssistantLogUpdatedSubscriptionVariables>;
export const FlowCreatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'flowCreated' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flowCreated' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'terminalFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Terminal' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'image' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'connected' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Flow' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'terminals' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'terminalFragment' } }],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'provider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowCreatedSubscription, FlowCreatedSubscriptionVariables>;
export const FlowDeletedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'flowDeleted' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flowDeleted' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'terminalFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Terminal' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'image' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'connected' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Flow' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'terminals' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'terminalFragment' } }],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'provider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowDeletedSubscription, FlowDeletedSubscriptionVariables>;
export const FlowUpdatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'flowUpdated' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flowUpdated' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'terminalFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Terminal' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'image' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'connected' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Provider' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Flow' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'terminals' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'terminalFragment' } }],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'provider' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowUpdatedSubscription, FlowUpdatedSubscriptionVariables>;
export const TaskCreatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'taskCreated' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'taskCreated' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'taskFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'subtaskFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Subtask' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'taskFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Task' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'input' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'subtasks' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'subtaskFragment' } }],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<TaskCreatedSubscription, TaskCreatedSubscriptionVariables>;
export const TaskUpdatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'taskUpdated' },
            variableDefinitions: [
                {
                    kind: 'VariableDefinition',
                    variable: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                    type: { kind: 'NonNullType', type: { kind: 'NamedType', name: { kind: 'Name', value: 'ID' } } },
                },
            ],
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'taskUpdated' },
                        arguments: [
                            {
                                kind: 'Argument',
                                name: { kind: 'Name', value: 'flowId' },
                                value: { kind: 'Variable', name: { kind: 'Name', value: 'flowId' } },
                            },
                        ],
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                                {
                                    kind: 'Field',
                                    name: { kind: 'Name', value: 'subtasks' },
                                    selectionSet: {
                                        kind: 'SelectionSet',
                                        selections: [
                                            {
                                                kind: 'FragmentSpread',
                                                name: { kind: 'Name', value: 'subtaskFragment' },
                                            },
                                        ],
                                    },
                                },
                                { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'subtaskFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'Subtask' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'result' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<TaskUpdatedSubscription, TaskUpdatedSubscriptionVariables>;
export const ProviderCreatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'providerCreated' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'providerCreated' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerConfigFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'model' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'temperature' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topK' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topP' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'minLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'repetitionPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'frequencyPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'presencePenalty' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reasoning' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'mode' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'effort' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'price' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'input' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'output' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheRead' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheWrite' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentsConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentsConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simple' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simpleJson' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'primaryAgent' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistant' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'generator' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'refiner' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'adviser' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reflector' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'searcher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'enricher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'coder' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'installer' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'pentester' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ProviderConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'agents' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentsConfigFragment' } },
                            ],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ProviderCreatedSubscription, ProviderCreatedSubscriptionVariables>;
export const ProviderUpdatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'providerUpdated' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'providerUpdated' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerConfigFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'model' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'temperature' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topK' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topP' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'minLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'repetitionPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'frequencyPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'presencePenalty' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reasoning' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'mode' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'effort' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'price' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'input' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'output' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheRead' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheWrite' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentsConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentsConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simple' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simpleJson' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'primaryAgent' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistant' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'generator' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'refiner' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'adviser' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reflector' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'searcher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'enricher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'coder' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'installer' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'pentester' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ProviderConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'agents' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentsConfigFragment' } },
                            ],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ProviderUpdatedSubscription, ProviderUpdatedSubscriptionVariables>;
export const ProviderDeletedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'providerDeleted' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'providerDeleted' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'providerConfigFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'model' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'temperature' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topK' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'topP' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'minLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'maxLength' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'repetitionPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'frequencyPenalty' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'presencePenalty' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reasoning' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'mode' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'effort' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'maxTokens' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'price' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'Field', name: { kind: 'Name', value: 'input' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'output' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheRead' } },
                                { kind: 'Field', name: { kind: 'Name', value: 'cacheWrite' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'agentsConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'AgentsConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simple' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'simpleJson' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'primaryAgent' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'assistant' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'generator' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'refiner' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'adviser' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'reflector' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'searcher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'enricher' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'coder' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'installer' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'pentester' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentConfigFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'providerConfigFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'ProviderConfig' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'type' } },
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'agents' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'agentsConfigFragment' } },
                            ],
                        },
                    },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ProviderDeletedSubscription, ProviderDeletedSubscriptionVariables>;
export const ApiTokenCreatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'apiTokenCreated' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'apiTokenCreated' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'apiTokenFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'apiTokenFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'APIToken' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'tokenId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'roleId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'ttl' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ApiTokenCreatedSubscription, ApiTokenCreatedSubscriptionVariables>;
export const ApiTokenUpdatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'apiTokenUpdated' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'apiTokenUpdated' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'apiTokenFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'apiTokenFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'APIToken' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'tokenId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'roleId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'ttl' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ApiTokenUpdatedSubscription, ApiTokenUpdatedSubscriptionVariables>;
export const ApiTokenDeletedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'apiTokenDeleted' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'apiTokenDeleted' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [{ kind: 'FragmentSpread', name: { kind: 'Name', value: 'apiTokenFragment' } }],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'apiTokenFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'APIToken' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'tokenId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'roleId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'ttl' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'status' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ApiTokenDeletedSubscription, ApiTokenDeletedSubscriptionVariables>;
export const SettingsUserUpdatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'settingsUserUpdated' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'settingsUserUpdated' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'userPreferencesFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'userPreferencesFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UserPreferences' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'favoriteFlows' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<SettingsUserUpdatedSubscription, SettingsUserUpdatedSubscriptionVariables>;
export const FlowTemplateCreatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'flowTemplateCreated' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flowTemplateCreated' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowTemplateFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowTemplateFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowTemplate' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'text' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowTemplateCreatedSubscription, FlowTemplateCreatedSubscriptionVariables>;
export const FlowTemplateUpdatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'flowTemplateUpdated' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flowTemplateUpdated' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowTemplateFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowTemplateFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowTemplate' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'text' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowTemplateUpdatedSubscription, FlowTemplateUpdatedSubscriptionVariables>;
export const FlowTemplateDeletedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'flowTemplateDeleted' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'flowTemplateDeleted' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'flowTemplateFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'flowTemplateFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'FlowTemplate' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'title' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'text' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<FlowTemplateDeletedSubscription, FlowTemplateDeletedSubscriptionVariables>;
export const ResourceAddedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'resourceAdded' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'resourceAdded' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'userResourceFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'userResourceFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UserResource' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'path' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'size' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'isDir' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ResourceAddedSubscription, ResourceAddedSubscriptionVariables>;
export const ResourceUpdatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'resourceUpdated' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'resourceUpdated' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'userResourceFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'userResourceFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UserResource' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'path' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'size' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'isDir' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ResourceUpdatedSubscription, ResourceUpdatedSubscriptionVariables>;
export const ResourceDeletedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'resourceDeleted' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'resourceDeleted' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'userResourceFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'userResourceFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'UserResource' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'name' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'path' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'size' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'isDir' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'createdAt' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'updatedAt' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<ResourceDeletedSubscription, ResourceDeletedSubscriptionVariables>;
export const KnowledgeDocumentCreatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'knowledgeDocumentCreated' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'knowledgeDocumentCreated' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'knowledgeDocumentFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'knowledgeDocumentFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'KnowledgeDocument' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'docType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'content' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'question' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'guideType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'answerType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'codeLang' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'partSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'manual' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<KnowledgeDocumentCreatedSubscription, KnowledgeDocumentCreatedSubscriptionVariables>;
export const KnowledgeDocumentUpdatedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'knowledgeDocumentUpdated' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'knowledgeDocumentUpdated' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'knowledgeDocumentFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'knowledgeDocumentFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'KnowledgeDocument' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'docType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'content' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'question' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'guideType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'answerType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'codeLang' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'partSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'manual' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<KnowledgeDocumentUpdatedSubscription, KnowledgeDocumentUpdatedSubscriptionVariables>;
export const KnowledgeDocumentDeletedDocument = {
    kind: 'Document',
    definitions: [
        {
            kind: 'OperationDefinition',
            operation: 'subscription',
            name: { kind: 'Name', value: 'knowledgeDocumentDeleted' },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    {
                        kind: 'Field',
                        name: { kind: 'Name', value: 'knowledgeDocumentDeleted' },
                        selectionSet: {
                            kind: 'SelectionSet',
                            selections: [
                                { kind: 'FragmentSpread', name: { kind: 'Name', value: 'knowledgeDocumentFragment' } },
                            ],
                        },
                    },
                ],
            },
        },
        {
            kind: 'FragmentDefinition',
            name: { kind: 'Name', value: 'knowledgeDocumentFragment' },
            typeCondition: { kind: 'NamedType', name: { kind: 'Name', value: 'KnowledgeDocument' } },
            selectionSet: {
                kind: 'SelectionSet',
                selections: [
                    { kind: 'Field', name: { kind: 'Name', value: 'id' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'docType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'content' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'question' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'description' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'userId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'flowId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'taskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'subtaskId' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'guideType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'answerType' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'codeLang' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'partSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'totalSize' } },
                    { kind: 'Field', name: { kind: 'Name', value: 'manual' } },
                ],
            },
        },
    ],
} as unknown as DocumentNode<KnowledgeDocumentDeletedSubscription, KnowledgeDocumentDeletedSubscriptionVariables>;
