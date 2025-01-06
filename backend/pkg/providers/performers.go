package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
)

const (
	maxRetriesToCallSimpleChain = 3
	maxRetriesToCallAgentChain  = 3
	maxRetriesToCallFunction    = 3
	toolCallsPrefix             = "[TOOL_CALLS] "
	delayBetweenRetries         = 5 * time.Second
)

func (fp *flowProvider) performAgentChain(
	ctx context.Context,
	opt provider.ProviderOptionsType,
	chainID int64,
	taskID, subtaskID *int64,
	chain []llms.MessageContent,
	executor tools.ContextToolsExecutor,
) ([]llms.MessageContent, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.performAgentChain")
	defer span.End()

	ticker := time.NewTicker(delayBetweenRetries)
	defer ticker.Stop()

	var (
		wantToStop bool
		detector   = &repeatingDetector{}
		resp       *llms.ContentResponse
		err        error
	)

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"agent":        fp.Type(),
		"flow_id":      fp.flowID,
		"task_id":      taskID,
		"subtask_id":   subtaskID,
		"msg_chain_id": chainID,
	})

	for {
		for idx := 0; idx <= maxRetriesToCallAgentChain; idx++ {
			if idx == maxRetriesToCallAgentChain {
				msg := fmt.Sprintf("failed to call agent chain: max retries reached, %d", idx)
				logger.WithError(err).Error(msg)
				return chain, fmt.Errorf(msg+": %w", err)
			}

			resp, err = fp.CallWithTools(ctx, opt, chain, executor.Tools())
			if err == nil {
				break
			}

			ticker.Reset(delayBetweenRetries)
			select {
			case <-ticker.C:
			case <-ctx.Done():
				logger.WithError(ctx.Err()).Error("context canceled while waiting for retry")
				return chain, fmt.Errorf("context canceled: %w", ctx.Err())
			}
		}

		if len(resp.Choices) == 0 {
			logger.Error("no choices in response")
			return chain, fmt.Errorf("no choices in response")
		}

		var (
			content   string
			funcCalls []llms.ToolCall
			info      map[string]any
		)
		for _, choice := range resp.Choices {
			choice.Content = strings.TrimSpace(choice.Content)
			if choice.Content != "" {
				content = choice.Content
			}

			if choice.GenerationInfo != nil {
				info = choice.GenerationInfo
			}

			for _, toolCall := range choice.ToolCalls {
				if toolCall.FunctionCall == nil {
					continue
				}
				funcCalls = append(funcCalls, toolCall)
			}
		}

		msg := llms.MessageContent{Role: llms.ChatMessageTypeAI}
		for _, toolCall := range funcCalls {
			msg.Parts = append(msg.Parts, llms.ToolCall{
				ID:           toolCall.ID,
				Type:         toolCall.Type,
				FunctionCall: toolCall.FunctionCall,
			})
		}
		if len(funcCalls) == 0 {
			msg.Parts = append(msg.Parts, llms.TextContent{Text: content})
		}

		chain = append(chain, msg)
		if len(funcCalls) == 0 {
			reflectorContext := map[string]any{
				"BarrierToolNames": executor.GetBarrierToolNames(),
				"Message":          content,
			}
			logger.WithField("content", content[:min(1000, len(content))]).Warn("got message instead of tool call")
			if len(chain) > 2 && chain[1].Role == llms.ChatMessageTypeHuman {
				if part, ok := chain[1].Parts[0].(llms.TextContent); ok {
					reflectorContext["Subtask"] = part.Text
				}
			}

			rctx, observation := obs.Observer.NewObservation(ctx)
			reflectorSpan := observation.Span(
				langfuse.WithStartSpanName("reflector agent"),
				langfuse.WithStartSpanInput(content),
				langfuse.WithStartSpanMetadata(reflectorContext),
			)
			rctx, _ = reflectorSpan.Observation(rctx)

			reflectorTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeReflector, reflectorContext)
			if err != nil {
				return chain, wrapErrorEndSpan(rctx, reflectorSpan, "failed to render reflector template", err)
			}

			opt := provider.OptionsTypeReflector
			msgChainType := database.MsgchainTypeReflector
			advice, err := fp.performSimpleChain(rctx, taskID, subtaskID, opt, msgChainType, reflectorTmpl)
			if err != nil {
				advice = ToolPlaceholder
			}

			reflectorSpan.End(langfuse.WithEndSpanOutput(advice))

			chain = append(chain, llms.TextParts(llms.ChatMessageTypeHuman, advice))
		}

		for _, toolCall := range funcCalls {
			if toolCall.FunctionCall == nil {
				continue
			}

			funcName := toolCall.FunctionCall.Name
			funcArgs := json.RawMessage(toolCall.FunctionCall.Arguments)
			if detector.detect(toolCall) {
				response := fmt.Sprintf("tool call '%s' is repeating, please try another tool", funcName)
				chain = append(chain, llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: toolCall.ID,
							Name:       funcName,
							Content:    response,
						},
					},
				})
				continue
			}

			ectx, observation := obs.Observer.NewObservation(ctx)
			opts := []langfuse.EventStartOption{
				langfuse.WithStartEventName(fmt.Sprintf("tool call %s", funcName)),
				langfuse.WithStartEventInput(funcArgs),
				langfuse.WithStartEventMetadata(map[string]any{
					"tool_call_id": toolCall.ID,
					"tool_name":    funcName,
				}),
			}

			var response string
			for idx := 0; idx <= maxRetriesToCallFunction; idx++ {
				if idx == maxRetriesToCallFunction {
					logger.WithError(err).WithField("args", string(funcArgs)).Error("failed to exec function")
					return chain, fmt.Errorf("failed to exec function '%s': %w", funcName, err)
				}

				response, err = executor.Execute(ectx, toolCall.ID, funcName, funcArgs)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return chain, err
					}

					observation.Event(append(opts,
						langfuse.WithStartEventStatus(err.Error()),
						langfuse.WithStartEventLevel(langfuse.ObservationLevelError),
					)...)
					logger.WithError(err).WithField("args", string(funcArgs)).Warn("failed to exec function")

					funcExecErr := err
					funcSchema, err := executor.GetToolSchema(funcName)
					if err != nil {
						logger.WithError(err).WithField("args", string(funcArgs)).Error("failed to get tool schema")
						return chain, fmt.Errorf("failed to get tool schema: %w", err)
					}

					funcArgs, err = fp.fixToolCallArgs(ctx, funcName, funcArgs, funcSchema, funcExecErr)
					if err != nil {
						logger.WithError(err).WithField("args", string(funcArgs)).Error("failed to fix tool call args")
						return chain, fmt.Errorf("failed to fix tool call args: %w", err)
					}
				} else {
					break
				}
			}

			observation.Event(append(opts,
				langfuse.WithStartEventStatus("success"),
				langfuse.WithStartEventOutput(response),
			)...)

			chain = append(chain, llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: toolCall.ID,
						Name:       funcName,
						Content:    response,
					},
				},
			})

			if executor.IsBarrierFunction(funcName) {
				wantToStop = true
			}
		}

		// TODO: here need to check chain length and summarize if it's too long

		chainBlob, err := json.Marshal(chain)
		if err != nil {
			logger.WithError(err).Error("failed to marshal msg chain")
			return chain, fmt.Errorf("failed to marshal msg chain: %w", err)
		}

		inputTokens, outputTokens := fp.GetUsage(info)
		_, err = fp.db.UpdateMsgChain(ctx, database.UpdateMsgChainParams{
			UsageIn:  inputTokens,
			UsageOut: outputTokens,
			Chain:    chainBlob,
			ID:       chainID,
		})
		if err != nil {
			logger.WithError(err).Error("failed to update msg chain")
			return chain, fmt.Errorf("failed to update msg chain: %w", err)
		}

		if wantToStop {
			return chain, nil
		}
	}
}

func (fp *flowProvider) performTaskResultReporter(
	ctx context.Context,
	taskID, subtaskID *int64,
	systemReporterTmpl, reporterTmpl, input string,
) (*tools.TaskResult, error) {
	var taskResult tools.TaskResult
	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemReporterTmpl),
		llms.TextParts(llms.ChatMessageTypeHuman, reporterTmpl),
	}
	cfg := tools.ReporterExecutorConfig{
		TaskID:    taskID,
		SubtaskID: subtaskID,
		ReportResult: func(ctx context.Context, name string, args json.RawMessage) (string, error) {
			err := json.Unmarshal(args, &taskResult)
			if err != nil {
				return "", fmt.Errorf("failed to unmarshal task result: %w", err)
			}
			return "report result successfully processed", nil
		},
	}
	executor, err := fp.executor.GetReporterExecutor(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get reporter executor: %w", err)
	}

	chainBlob, err := json.Marshal(chain)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal msg chain: %w", err)
	}

	opt := provider.OptionsTypeSimple
	msgChain, err := fp.db.CreateMsgChain(ctx, database.CreateMsgChainParams{
		Type:          database.MsgchainTypeReporter,
		Model:         fp.Model(opt),
		ModelProvider: string(fp.Type()),
		Chain:         chainBlob,
		FlowID:        fp.flowID,
		TaskID:        database.Int64ToNullInt64(taskID),
		SubtaskID:     database.Int64ToNullInt64(subtaskID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create msg chain: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, database.MsgchainTypeReporter)
	chain, err = fp.performAgentChain(ctx, opt, msgChain.ID, taskID, subtaskID, chain, executor)
	if err != nil {
		return nil, fmt.Errorf("failed to get task reporter result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.agentLog.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			fmt.Sprintf("Report task result for user's task:\n\n%s", input),
			taskResult.Result,
			taskID,
			subtaskID,
		)
	}

	return &taskResult, nil
}

func (fp *flowProvider) performSubtasksGenerator(
	ctx context.Context,
	taskID int64,
	systemGeneratorTmpl, generatorTmpl, input string,
) ([]tools.SubtaskInfo, error) {
	var subtaskList tools.SubtaskList
	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemGeneratorTmpl),
		llms.TextParts(llms.ChatMessageTypeHuman, generatorTmpl),
	}

	memorist, err := fp.getMemoristHandler(ctx, taskID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get memorist handler: %w", err)
	}

	searcher, err := fp.getTaskSearcherHandler(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get searcher handler: %w", err)
	}

	cfg := tools.GeneratorExecutorConfig{
		TaskID:   taskID,
		Memorist: memorist,
		Searcher: searcher,
		SubtaskList: func(ctx context.Context, name string, args json.RawMessage) (string, error) {
			err := json.Unmarshal(args, &subtaskList)
			if err != nil {
				return "", fmt.Errorf("failed to unmarshal subtask list: %w", err)
			}
			return "subtask list successfully processed", nil
		},
	}
	executor, err := fp.executor.GetGeneratorExecutor(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get generator executor: %w", err)
	}

	chainBlob, err := json.Marshal(chain)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal msg chain: %w", err)
	}

	opt := provider.OptionsTypeGenerator
	msgChain, err := fp.db.CreateMsgChain(ctx, database.CreateMsgChainParams{
		Type:          database.MsgchainTypeGenerator,
		Model:         fp.Model(opt),
		ModelProvider: string(fp.Type()),
		Chain:         chainBlob,
		FlowID:        fp.flowID,
		TaskID:        database.Int64ToNullInt64(&taskID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create msg chain: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, database.MsgchainTypeGenerator)
	chain, err = fp.performAgentChain(ctx, opt, msgChain.ID, &taskID, nil, chain, executor)
	if err != nil {
		return nil, fmt.Errorf("failed to get subtasks generator result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.agentLog.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			fmt.Sprintf("Generate new subtasks for user's task:\n\n%s", input),
			fp.subtasksToMarkdown(subtaskList.Subtasks),
			&taskID,
			nil,
		)
	}

	return subtaskList.Subtasks, nil
}

func (fp *flowProvider) performSubtasksRefiner(
	ctx context.Context,
	taskID int64,
	systemRefinerTmpl, refinerTmpl, input string,
) ([]tools.SubtaskInfo, error) {
	var (
		subtaskList tools.SubtaskList
		chain       []llms.MessageContent
	)

	msgChainType := database.MsgchainTypeRefiner
	msgChains, err := fp.db.GetTaskTypeMsgChains(ctx, database.GetTaskTypeMsgChainsParams{
		TaskID: database.Int64ToNullInt64(&taskID),
		Type:   msgChainType,
	})
	if err != nil || len(msgChains) == 0 {
		chain = []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, systemRefinerTmpl),
			llms.TextParts(llms.ChatMessageTypeHuman, refinerTmpl),
		}
	} else {
		err = json.Unmarshal(msgChains[0].Chain, &chain)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg chain: %w", err)
		}

		chain[0] = llms.TextParts(llms.ChatMessageTypeSystem, systemRefinerTmpl)
		chain = append(chain, llms.TextParts(llms.ChatMessageTypeHuman, refinerTmpl))
	}

	memorist, err := fp.getMemoristHandler(ctx, taskID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get memorist handler: %w", err)
	}

	searcher, err := fp.getTaskSearcherHandler(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get searcher handler: %w", err)
	}

	cfg := tools.GeneratorExecutorConfig{
		TaskID:   taskID,
		Memorist: memorist,
		Searcher: searcher,
		SubtaskList: func(ctx context.Context, name string, args json.RawMessage) (string, error) {
			err := json.Unmarshal(args, &subtaskList)
			if err != nil {
				return "", fmt.Errorf("failed to unmarshal subtask list: %w", err)
			}
			return "subtask list successfully processed", nil
		},
	}
	executor, err := fp.executor.GetGeneratorExecutor(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get generator executor: %w", err)
	}

	chainBlob, err := json.Marshal(chain)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal msg chain: %w", err)
	}

	opt := provider.OptionsTypeRefiner
	msgChain, err := fp.db.CreateMsgChain(ctx, database.CreateMsgChainParams{
		Type:          msgChainType,
		Model:         fp.Model(opt),
		ModelProvider: string(fp.Type()),
		Chain:         chainBlob,
		FlowID:        fp.flowID,
		TaskID:        database.Int64ToNullInt64(&taskID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create msg chain: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, database.MsgchainTypeRefiner)
	chain, err = fp.performAgentChain(ctx, opt, msgChain.ID, &taskID, nil, chain, executor)
	if err != nil {
		return nil, fmt.Errorf("failed to get subtasks refiner result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.agentLog.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			fmt.Sprintf("Refine subtasks list for user's task:\n\n%s", input),
			fp.subtasksToMarkdown(subtaskList.Subtasks),
			&taskID,
			nil,
		)
	}

	return subtaskList.Subtasks, nil
}

func (fp *flowProvider) performCoder(
	ctx context.Context,
	taskID, subtaskID int64,
	systemCoderTmpl, question string,
) (string, error) {
	var codeResult tools.CodeResult
	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemCoderTmpl),
		llms.TextParts(llms.ChatMessageTypeHuman, question),
	}

	adviser, err := fp.getAskAdviceHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get adviser handler: %w", err)
	}

	installer, err := fp.getInstallerHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get installer handler: %w", err)
	}

	memorist, err := fp.getMemoristHandler(ctx, taskID, &subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get memorist handler: %w", err)
	}

	searcher, err := fp.getSubtaskSearcherHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get searcher handler: %w", err)
	}

	cfg := tools.CoderExecutorConfig{
		TaskID:    taskID,
		SubtaskID: subtaskID,
		Adviser:   adviser,
		Installer: installer,
		Memorist:  memorist,
		Searcher:  searcher,
		CodeResult: func(ctx context.Context, name string, args json.RawMessage) (string, error) {
			err := json.Unmarshal(args, &codeResult)
			if err != nil {
				return "", fmt.Errorf("failed to unmarshal result: %w", err)
			}
			return "code result successfully processed", nil
		},
		Summarizer: fp.getSummarizeResultHandler(&taskID, &subtaskID),
	}
	executor, err := fp.executor.GetCoderExecutor(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get coder executor: %w", err)
	}

	chainBlob, err := json.Marshal(chain)
	if err != nil {
		return "", fmt.Errorf("failed to marshal msg chain: %w", err)
	}

	opt := provider.OptionsTypeCoder
	msgChain, err := fp.db.CreateMsgChain(ctx, database.CreateMsgChainParams{
		Type:          database.MsgchainTypeCoder,
		Model:         fp.Model(opt),
		ModelProvider: string(fp.Type()),
		Chain:         chainBlob,
		FlowID:        fp.flowID,
		TaskID:        database.Int64ToNullInt64(&taskID),
		SubtaskID:     database.Int64ToNullInt64(&subtaskID),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create msg chain: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, database.MsgchainTypeCoder)
	chain, err = fp.performAgentChain(ctx, opt, msgChain.ID, &taskID, &subtaskID, chain, executor)
	if err != nil {
		return "", fmt.Errorf("failed to get task coder result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.agentLog.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			question,
			codeResult.Result,
			&taskID,
			&subtaskID,
		)
	}

	return codeResult.Result, nil
}

func (fp *flowProvider) performInstaller(
	ctx context.Context,
	taskID, subtaskID int64,
	systemInstallerTmpl, question string,
) (string, error) {
	var maintenanceResult tools.MaintenanceResult
	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemInstallerTmpl),
		llms.TextParts(llms.ChatMessageTypeHuman, question),
	}

	adviser, err := fp.getAskAdviceHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get adviser handler: %w", err)
	}

	memorist, err := fp.getMemoristHandler(ctx, taskID, &subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get memorist handler: %w", err)
	}

	searcher, err := fp.getSubtaskSearcherHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get searcher handler: %w", err)
	}

	cfg := tools.InstallerExecutorConfig{
		TaskID:    taskID,
		SubtaskID: subtaskID,
		Adviser:   adviser,
		Memorist:  memorist,
		Searcher:  searcher,
		MaintenanceResult: func(ctx context.Context, name string, args json.RawMessage) (string, error) {
			err := json.Unmarshal(args, &maintenanceResult)
			if err != nil {
				return "", fmt.Errorf("failed to unmarshal result: %w", err)
			}
			return "maintenance result successfully processed", nil
		},
		Summarizer: fp.getSummarizeResultHandler(&taskID, &subtaskID),
	}
	executor, err := fp.executor.GetInstallerExecutor(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get installer executor: %w", err)
	}

	chainBlob, err := json.Marshal(chain)
	if err != nil {
		return "", fmt.Errorf("failed to marshal msg chain: %w", err)
	}

	opt := provider.OptionsTypeInstaller
	msgChain, err := fp.db.CreateMsgChain(ctx, database.CreateMsgChainParams{
		Type:          database.MsgchainTypeInstaller,
		Model:         fp.Model(opt),
		ModelProvider: string(fp.Type()),
		Chain:         chainBlob,
		FlowID:        fp.flowID,
		TaskID:        database.Int64ToNullInt64(&taskID),
		SubtaskID:     database.Int64ToNullInt64(&subtaskID),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create msg chain: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, database.MsgchainTypeInstaller)
	chain, err = fp.performAgentChain(ctx, opt, msgChain.ID, &taskID, &subtaskID, chain, executor)
	if err != nil {
		return "", fmt.Errorf("failed to get task installer result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.agentLog.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			question,
			maintenanceResult.Result,
			&taskID,
			&subtaskID,
		)
	}

	return maintenanceResult.Result, nil
}

func (fp *flowProvider) performMemorist(
	ctx context.Context,
	taskID, subtaskID *int64,
	systemMemoristTmpl, question string,
) (string, error) {
	var memoristResult tools.MemoristResult
	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemMemoristTmpl),
		llms.TextParts(llms.ChatMessageTypeHuman, question),
	}
	cfg := tools.MemoristExecutorConfig{
		TaskID:    taskID,
		SubtaskID: subtaskID,
		SearchResult: func(ctx context.Context, name string, args json.RawMessage) (string, error) {
			err := json.Unmarshal(args, &memoristResult)
			if err != nil {
				return "", fmt.Errorf("failed to unmarshal result: %w", err)
			}
			return "memorist result successfully processed", nil
		},
		Summarizer: fp.getSummarizeResultHandler(taskID, subtaskID),
	}
	executor, err := fp.executor.GetMemoristExecutor(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get memorist executor: %w", err)
	}

	chainBlob, err := json.Marshal(chain)
	if err != nil {
		return "", fmt.Errorf("failed to marshal msg chain: %w", err)
	}

	opt := provider.OptionsTypeSearcher
	msgChain, err := fp.db.CreateMsgChain(ctx, database.CreateMsgChainParams{
		Type:          database.MsgchainTypeMemorist,
		Model:         fp.Model(opt),
		ModelProvider: string(fp.Type()),
		Chain:         chainBlob,
		FlowID:        fp.flowID,
		TaskID:        database.Int64ToNullInt64(taskID),
		SubtaskID:     database.Int64ToNullInt64(subtaskID),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create msg chain: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, database.MsgchainTypeMemorist)
	chain, err = fp.performAgentChain(ctx, opt, msgChain.ID, taskID, subtaskID, chain, executor)
	if err != nil {
		return "", fmt.Errorf("failed to get task memorist result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.agentLog.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			question,
			memoristResult.Result,
			taskID,
			subtaskID,
		)
	}

	return memoristResult.Result, nil
}

func (fp *flowProvider) performPentester(
	ctx context.Context,
	taskID, subtaskID int64,
	systemPentesterTmpl, question string,
) (string, error) {
	var hackResult tools.HackResult
	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPentesterTmpl),
		llms.TextParts(llms.ChatMessageTypeHuman, question),
	}

	adviser, err := fp.getAskAdviceHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get adviser handler: %w", err)
	}

	coder, err := fp.getCoderHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get coder handler: %w", err)
	}

	installer, err := fp.getInstallerHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get installer handler: %w", err)
	}

	memorist, err := fp.getMemoristHandler(ctx, taskID, &subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get memorist handler: %w", err)
	}

	searcher, err := fp.getSubtaskSearcherHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get searcher handler: %w", err)
	}

	cfg := tools.PentesterExecutorConfig{
		TaskID:    taskID,
		SubtaskID: subtaskID,
		Adviser:   adviser,
		Coder:     coder,
		Installer: installer,
		Memorist:  memorist,
		Searcher:  searcher,
		HackResult: func(ctx context.Context, name string, args json.RawMessage) (string, error) {
			err := json.Unmarshal(args, &hackResult)
			if err != nil {
				return "", fmt.Errorf("failed to unmarshal result: %w", err)
			}
			return "hack result successfully processed", nil
		},
		Summarizer: fp.getSummarizeResultHandler(&taskID, &subtaskID),
	}
	executor, err := fp.executor.GetPentesterExecutor(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get pentester executor: %w", err)
	}

	chainBlob, err := json.Marshal(chain)
	if err != nil {
		return "", fmt.Errorf("failed to marshal msg chain: %w", err)
	}

	opt := provider.OptionsTypePentester
	msgChain, err := fp.db.CreateMsgChain(ctx, database.CreateMsgChainParams{
		Type:          database.MsgchainTypePentester,
		Model:         fp.Model(opt),
		ModelProvider: string(fp.Type()),
		Chain:         chainBlob,
		FlowID:        fp.flowID,
		TaskID:        database.Int64ToNullInt64(&taskID),
		SubtaskID:     database.Int64ToNullInt64(&subtaskID),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create msg chain: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, database.MsgchainTypePentester)
	chain, err = fp.performAgentChain(ctx, opt, msgChain.ID, &taskID, &subtaskID, chain, executor)
	if err != nil {
		return "", fmt.Errorf("failed to get task pentester result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.agentLog.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			question,
			hackResult.Result,
			&taskID,
			&subtaskID,
		)
	}

	return hackResult.Result, nil
}

func (fp *flowProvider) performSearcher(
	ctx context.Context,
	taskID, subtaskID *int64,
	systemSearcherTmpl, question string,
) (string, error) {
	var searchResult tools.SearchResult
	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemSearcherTmpl),
		llms.TextParts(llms.ChatMessageTypeHuman, question),
	}

	tID := int64(0)
	if taskID != nil {
		tID = *taskID
	}
	memorist, err := fp.getMemoristHandler(ctx, tID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get memorist handler: %w", err)
	}

	cfg := tools.SearcherExecutorConfig{
		TaskID:    taskID,
		SubtaskID: subtaskID,
		Memorist:  memorist,
		SearchResult: func(ctx context.Context, name string, args json.RawMessage) (string, error) {
			err := json.Unmarshal(args, &searchResult)
			if err != nil {
				return "", fmt.Errorf("failed to unmarshal result: %w", err)
			}
			return "search result successfully processed", nil
		},
		Summarizer: fp.getSummarizeResultHandler(taskID, subtaskID),
	}
	executor, err := fp.executor.GetSearcherExecutor(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get searcher executor: %w", err)
	}

	chainBlob, err := json.Marshal(chain)
	if err != nil {
		return "", fmt.Errorf("failed to marshal msg chain: %w", err)
	}

	opt := provider.OptionsTypeSearcher
	msgChain, err := fp.db.CreateMsgChain(ctx, database.CreateMsgChainParams{
		Type:          database.MsgchainTypeSearcher,
		Model:         fp.Model(opt),
		ModelProvider: string(fp.Type()),
		Chain:         chainBlob,
		FlowID:        fp.flowID,
		TaskID:        database.Int64ToNullInt64(taskID),
		SubtaskID:     database.Int64ToNullInt64(subtaskID),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create msg chain: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, database.MsgchainTypeSearcher)
	chain, err = fp.performAgentChain(ctx, opt, msgChain.ID, taskID, subtaskID, chain, executor)
	if err != nil {
		return "", fmt.Errorf("failed to get task searcher result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.agentLog.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			question,
			searchResult.Result,
			taskID,
			subtaskID,
		)
	}

	return searchResult.Result, nil
}

func (fp *flowProvider) performEnricher(
	ctx context.Context,
	taskID, subtaskID int64,
	systemEnricherTmpl, question string,
) (string, error) {
	var enricherResult tools.EnricherResult
	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemEnricherTmpl),
		llms.TextParts(llms.ChatMessageTypeHuman, question),
	}

	memorist, err := fp.getMemoristHandler(ctx, taskID, &subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get memorist handler: %w", err)
	}

	searcher, err := fp.getSubtaskSearcherHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get searcher handler: %w", err)
	}

	cfg := tools.EnricherExecutorConfig{
		TaskID:    taskID,
		SubtaskID: subtaskID,
		Memorist:  memorist,
		Searcher:  searcher,
		EnricherResult: func(ctx context.Context, name string, args json.RawMessage) (string, error) {
			err := json.Unmarshal(args, &enricherResult)
			if err != nil {
				return "", fmt.Errorf("failed to unmarshal result: %w", err)
			}
			return "enrich result successfully processed", nil
		},
	}
	executor, err := fp.executor.GetEnricherExecutor(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get enricher executor: %w", err)
	}

	chainBlob, err := json.Marshal(chain)
	if err != nil {
		return "", fmt.Errorf("failed to marshal msg chain: %w", err)
	}

	opt := provider.OptionsTypeEnricher
	msgChain, err := fp.db.CreateMsgChain(ctx, database.CreateMsgChainParams{
		Type:          database.MsgchainTypeEnricher,
		Model:         fp.Model(opt),
		ModelProvider: string(fp.Type()),
		Chain:         chainBlob,
		FlowID:        fp.flowID,
		TaskID:        database.Int64ToNullInt64(&taskID),
		SubtaskID:     database.Int64ToNullInt64(&subtaskID),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create msg chain: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, database.MsgchainTypeEnricher)
	chain, err = fp.performAgentChain(ctx, opt, msgChain.ID, &taskID, &subtaskID, chain, executor)
	if err != nil {
		return "", fmt.Errorf("failed to get task searcher result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.agentLog.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			question,
			enricherResult.Result,
			&taskID,
			&subtaskID,
		)
	}

	return enricherResult.Result, nil
}

func (fp *flowProvider) performSimpleChain(
	ctx context.Context,
	taskID, subtaskID *int64,
	opt provider.ProviderOptionsType,
	msgChainType database.MsgchainType,
	prompt string,
) (string, error) {
	var (
		resp *llms.ContentResponse
		err  error
	)
	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}

	for idx := 0; idx <= maxRetriesToCallSimpleChain; idx++ {
		if idx == maxRetriesToCallSimpleChain {
			return "", fmt.Errorf("failed to call simple chain: %w", err)
		}

		resp, err = fp.CallEx(ctx, opt, chain)
		if err == nil {
			break
		}
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	choice := resp.Choices[0]
	chain = append(chain, llms.TextParts(llms.ChatMessageTypeAI, choice.Content))

	chainBlob, err := json.Marshal(chain)
	if err != nil {
		return "", fmt.Errorf("failed to marshal summarizer msg chain: %w", err)
	}

	inputTokens, outputTokens := fp.GetUsage(choice.GenerationInfo)
	_, err = fp.db.CreateMsgChain(ctx, database.CreateMsgChainParams{
		Type:          msgChainType,
		Model:         fp.Model(opt),
		ModelProvider: string(fp.Type()),
		UsageIn:       inputTokens,
		UsageOut:      outputTokens,
		Chain:         chainBlob,
		FlowID:        fp.flowID,
		TaskID:        database.Int64ToNullInt64(taskID),
		SubtaskID:     database.Int64ToNullInt64(subtaskID),
	})

	return choice.Content, nil
}
