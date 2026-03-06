package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"pentagi/pkg/cast"
	"pentagi/pkg/csum"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/providers/embeddings"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
	"github.com/vxcontrol/langchaingo/llms"
)

type AssistantProvider interface {
	Type() provider.ProviderType
	Model(opt pconfig.ProviderOptionsType) string
	Title() string
	Language() string
	ToolCallIDTemplate() string
	Embedder() embeddings.Embedder

	SetMsgChainID(msgChainID int64)
	SetAgentLogProvider(agentLog tools.AgentLogProvider)
	SetMsgLogProvider(msgLog tools.MsgLogProvider)

	PrepareAgentChain(ctx context.Context) (int64, error)
	PerformAgentChain(ctx context.Context) error
	PutInputToAgentChain(ctx context.Context, input string) error
	EnsureChainConsistency(ctx context.Context) error
}

type assistantProvider struct {
	id         int64
	msgChainID int64
	summarizer csum.Summarizer
	fp         flowProvider
}

func (ap *assistantProvider) Type() provider.ProviderType {
	return ap.fp.Type()
}

func (ap *assistantProvider) Model(opt pconfig.ProviderOptionsType) string {
	return ap.fp.Model(opt)
}

func (ap *assistantProvider) Title() string {
	return ap.fp.Title()
}

func (ap *assistantProvider) Language() string {
	return ap.fp.Language()
}

func (ap *assistantProvider) ToolCallIDTemplate() string {
	return ap.fp.ToolCallIDTemplate()
}

func (ap *assistantProvider) Embedder() embeddings.Embedder {
	return ap.fp.Embedder()
}

func (ap *assistantProvider) SetMsgChainID(msgChainID int64) {
	ap.msgChainID = msgChainID
}

func (ap *assistantProvider) SetAgentLogProvider(agentLog tools.AgentLogProvider) {
	ap.fp.SetAgentLogProvider(agentLog)
}

func (ap *assistantProvider) SetMsgLogProvider(msgLog tools.MsgLogProvider) {
	ap.fp.SetMsgLogProvider(msgLog)
}

func (ap *assistantProvider) PrepareAgentChain(ctx context.Context) (int64, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.PrepareAssistantChain")
	defer span.End()

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"provider":     ap.fp.Type(),
		"assistant_id": ap.id,
		"flow_id":      ap.fp.ID(),
	})

	systemPrompt, err := ap.getAssistantSystemPrompt(ctx)
	if err != nil {
		logger.WithError(err).Error("failed to get assistant system prompt")
		return 0, fmt.Errorf("failed to get assistant system prompt: %w", err)
	}

	optAgentType := pconfig.OptionsTypeAssistant
	msgChainType := database.MsgchainTypeAssistant
	ap.msgChainID, _, err = ap.fp.restoreChain(
		ctx, nil, nil, optAgentType, msgChainType, systemPrompt, "",
	)
	if err != nil {
		logger.WithError(err).Error("failed to restore assistant msg chain")
		return 0, fmt.Errorf("failed to restore assistant msg chain: %w", err)
	}

	return ap.msgChainID, nil
}

func (ap *assistantProvider) PerformAgentChain(ctx context.Context) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.assistantProvider.PerformAgentChain")
	defer span.End()

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"provider":     ap.fp.Type(),
		"assistant_id": ap.id,
		"flow_id":      ap.fp.ID(),
		"msg_chain_id": ap.msgChainID,
	})

	useAgents, err := ap.getAssistantUseAgents(ctx)
	if err != nil {
		logger.WithError(err).Error("failed to get assistant use agents")
		return fmt.Errorf("failed to get assistant use agents: %w", err)
	}

	msgChain, err := ap.fp.DB().GetMsgChain(ctx, ap.msgChainID)
	if err != nil {
		logger.WithError(err).Error("failed to get primary agent msg chain")
		return fmt.Errorf("failed to get primary agent msg chain %d: %w", ap.msgChainID, err)
	}

	var chain []llms.MessageContent
	if err := json.Unmarshal(msgChain.Chain, &chain); err != nil {
		logger.WithError(err).Error("failed to unmarshal primary agent msg chain")
		return fmt.Errorf("failed to unmarshal primary agent msg chain %d: %w", ap.msgChainID, err)
	}

	adviser, err := ap.fp.GetAskAdviceHandler(ctx, nil, nil)
	if err != nil {
		logger.WithError(err).Error("failed to get ask advice handler")
		return fmt.Errorf("failed to get ask advice handler: %w", err)
	}

	coder, err := ap.fp.GetCoderHandler(ctx, nil, nil)
	if err != nil {
		logger.WithError(err).Error("failed to get coder handler")
		return fmt.Errorf("failed to get coder handler: %w", err)
	}

	installer, err := ap.fp.GetInstallerHandler(ctx, nil, nil)
	if err != nil {
		logger.WithError(err).Error("failed to get installer handler")
		return fmt.Errorf("failed to get installer handler: %w", err)
	}

	memorist, err := ap.fp.GetMemoristHandler(ctx, nil, nil)
	if err != nil {
		logger.WithError(err).Error("failed to get memorist handler")
		return fmt.Errorf("failed to get memorist handler: %w", err)
	}

	pentester, err := ap.fp.GetPentesterHandler(ctx, nil, nil)
	if err != nil {
		logger.WithError(err).Error("failed to get pentester handler")
		return fmt.Errorf("failed to get pentester handler: %w", err)
	}

	searcher, err := ap.fp.GetSubtaskSearcherHandler(ctx, nil, nil)
	if err != nil {
		logger.WithError(err).Error("failed to get searcher handler")
		return fmt.Errorf("failed to get searcher handler: %w", err)
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	executorAgent := observation.Agent(
		langfuse.WithAgentName(fmt.Sprintf("assistant %d for flow %d: %s", ap.id, ap.fp.ID(), ap.fp.Title())),
		langfuse.WithAgentInput(chain),
		langfuse.WithAgentMetadata(langfuse.Metadata{
			"assistant_id": ap.id,
			"flow_id":      ap.fp.ID(),
			"msg_chain_id": ap.msgChainID,
			"provider":     ap.fp.Type(),
			"image":        ap.fp.Image(),
			"lang":         ap.fp.Language(),
		}),
	)
	ctx, _ = executorAgent.Observation(ctx)

	cfg := tools.AssistantExecutorConfig{
		UseAgents:  useAgents,
		Adviser:    adviser,
		Coder:      coder,
		Installer:  installer,
		Memorist:   memorist,
		Pentester:  pentester,
		Searcher:   searcher,
		Summarizer: ap.fp.GetSummarizeResultHandler(nil, nil),
	}

	executor, err := ap.fp.Executor().GetAssistantExecutor(cfg)
	if err != nil {
		return wrapErrorEndAgentSpan(ctx, executorAgent, "failed to get assistant executor", err)
	}

	ctx = tools.PutAgentContext(ctx, database.MsgchainTypeAssistant)
	err = ap.fp.performAgentChain(
		ctx, pconfig.OptionsTypeAssistant, msgChain.ID, nil, nil, chain, executor, ap.summarizer,
	)
	if err != nil {
		return wrapErrorEndAgentSpan(ctx, executorAgent, "failed to perform assistant agent chain", err)
	}

	executorAgent.End()

	return nil
}

func (ap *assistantProvider) PutInputToAgentChain(ctx context.Context, input string) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.assistantProvider.PutInputToAgentChain")
	defer span.End()

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"provider":     ap.fp.Type(),
		"assistant_id": ap.id,
		"flow_id":      ap.fp.ID(),
		"msg_chain_id": ap.msgChainID,
		"input":        input[:min(len(input), 1000)],
	})

	return ap.fp.processChain(ctx, ap.msgChainID, logger, func(chain []llms.MessageContent) ([]llms.MessageContent, error) {
		return ap.updateAssistantChain(ctx, chain, input)
	})
}

func (ap *assistantProvider) EnsureChainConsistency(ctx context.Context) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.assistantProvider.EnsureChainConsistency")
	defer span.End()

	return ap.fp.EnsureChainConsistency(ctx, ap.msgChainID)
}

func (ap *assistantProvider) updateAssistantChain(
	ctx context.Context, chain []llms.MessageContent, humanPrompt string,
) ([]llms.MessageContent, error) {
	systemPrompt, err := ap.getAssistantSystemPrompt(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get assistant system prompt: %w", err)
	}

	if len(chain) == 0 {
		return []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
			llms.TextParts(llms.ChatMessageTypeHuman, humanPrompt),
		}, nil
	}

	ast, err := cast.NewChainAST(chain, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create chain ast: %w", err)
	}

	systemMessage := llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt)
	ast.Sections[0].Header.SystemMessage = &systemMessage

	ast.AppendHumanMessage(humanPrompt)

	return ast.Messages(), nil
}

func (ap *assistantProvider) getAssistantUseAgents(ctx context.Context) (bool, error) {
	return ap.fp.DB().GetAssistantUseAgents(ctx, ap.id)
}

func (ap *assistantProvider) getAssistantSystemPrompt(ctx context.Context) (string, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"provider":     ap.fp.Type(),
		"assistant_id": ap.id,
		"flow_id":      ap.fp.ID(),
	})

	useAgents, err := ap.getAssistantUseAgents(ctx)
	if err != nil {
		logger.WithError(err).Error("failed to get assistant use agents")
		return "", fmt.Errorf("failed to get assistant use agents: %w", err)
	}

	executionContext, err := ap.getAssistantExecutionContext(ctx)
	if err != nil {
		logger.WithError(err).Error("failed to get assistant execution context")
		return "", fmt.Errorf("failed to get assistant execution context: %w", err)
	}

	systemAssistantTmpl, err := ap.fp.Prompter().RenderTemplate(templates.PromptTypeAssistant, map[string]any{
		"SearchToolName":          tools.SearchToolName,
		"PentesterToolName":       tools.PentesterToolName,
		"CoderToolName":           tools.CoderToolName,
		"AdviceToolName":          tools.AdviceToolName,
		"MemoristToolName":        tools.MemoristToolName,
		"MaintenanceToolName":     tools.MaintenanceToolName,
		"TerminalToolName":        tools.TerminalToolName,
		"FileToolName":            tools.FileToolName,
		"GoogleToolName":          tools.GoogleToolName,
		"DuckDuckGoToolName":      tools.DuckDuckGoToolName,
		"TavilyToolName":          tools.TavilyToolName,
		"TraversaalToolName":      tools.TraversaalToolName,
		"PerplexityToolName":      tools.PerplexityToolName,
		"BrowserToolName":         tools.BrowserToolName,
		"SearchInMemoryToolName":  tools.SearchInMemoryToolName,
		"SearchGuideToolName":     tools.SearchGuideToolName,
		"SearchAnswerToolName":    tools.SearchAnswerToolName,
		"SearchCodeToolName":      tools.SearchCodeToolName,
		"SummarizationToolName":   cast.SummarizationToolName,
		"SummarizedContentPrefix": strings.ReplaceAll(csum.SummarizedContentPrefix, "\n", "\\n"),
		"UseAgents":               useAgents,
		"DockerImage":             ap.fp.Image(),
		"Cwd":                     docker.WorkFolderPathInContainer,
		"ContainerPorts":          ap.fp.getContainerPortsDescription(),
		"ExecutionContext":        executionContext,
		"Lang":                    ap.fp.Language(),
		"CurrentTime":             getCurrentTime(),
	})
	if err != nil {
		logger.WithError(err).Error("failed to get system prompt for assistant template")
		return "", fmt.Errorf("failed to get system prompt for assistant template: %w", err)
	}

	return systemAssistantTmpl, nil
}

func (ap *assistantProvider) getAssistantExecutionContext(ctx context.Context) (string, error) {
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"provider":     ap.fp.Type(),
		"assistant_id": ap.id,
		"flow_id":      ap.fp.ID(),
	})

	subtasks, err := ap.fp.DB().GetFlowSubtasks(ctx, ap.fp.ID())
	if err != nil {
		logger.WithError(err).Error("failed to get flow subtasks")
		return "", fmt.Errorf("failed to get flow subtasks: %w", err)
	}

	slices.SortFunc(subtasks, func(a, b database.Subtask) int {
		return int(a.ID - b.ID)
	})

	var (
		executionContext     string
		lastActiveSubtaskIDX int = -1
	)
	for sdx, subtask := range subtasks {
		if subtask.Status != database.SubtaskStatusCreated {
			lastActiveSubtaskIDX = sdx
		}

		if subtask.Context != "" {
			executionContext = subtask.Context
		}
	}

	if executionContext == "" && len(subtasks) > 0 {
		if lastActiveSubtaskIDX == -1 {
			lastActiveSubtaskIDX = len(subtasks) - 1
		}

		lastSubtask := subtasks[lastActiveSubtaskIDX]
		executionContext, err = ap.fp.prepareExecutionContext(ctx, lastSubtask.TaskID, lastSubtask.ID)
		if err != nil {
			logger.WithError(err).Error("failed to prepare execution context")
			return "", fmt.Errorf("failed to prepare execution context: %w", err)
		}
	}

	return executionContext, nil
}
