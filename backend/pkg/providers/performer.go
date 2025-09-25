package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"pentagi/pkg/cast"
	"pentagi/pkg/csum"
	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

const (
	maxRetriesToCallSimpleChain = 3
	maxRetriesToCallAgentChain  = 3
	maxRetriesToCallFunction    = 3
	maxReflectorCallsPerChain   = 3
	delayBetweenRetries         = 5 * time.Second
)

type callResult struct {
	streamID  int64
	funcCalls []llms.ToolCall
	info      map[string]any
	thinking  string
	content   string
}

func (fp *flowProvider) performAgentChain(
	ctx context.Context,
	optAgentType pconfig.ProviderOptionsType,
	chainID int64,
	taskID, subtaskID *int64,
	chain []llms.MessageContent,
	executor tools.ContextToolsExecutor,
	summarizer csum.Summarizer,
) error {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.performAgentChain")
	defer span.End()

	var (
		wantToStop        bool
		detector          = &repeatingDetector{}
		summarizerHandler = fp.GetSummarizeResultHandler(taskID, subtaskID)
	)

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"agent":        fp.Type(),
		"flow_id":      fp.flowID,
		"task_id":      taskID,
		"subtask_id":   subtaskID,
		"msg_chain_id": chainID,
	})

	executionContext, err := fp.getExecutionContext(ctx, taskID, subtaskID)
	if err != nil {
		logger.WithError(err).Error("failed to get execution context")
		return fmt.Errorf("failed to get execution context: %w", err)
	}

	for {
		result, err := fp.callWithRetries(ctx, chain, optAgentType, executor)
		if err != nil {
			logger.WithError(err).Error("failed to call agent chain")
			return err
		}

		if err := fp.updateMsgChainUsage(ctx, chainID, result.info); err != nil {
			logger.WithError(err).Error("failed to update msg chain usage")
			return err
		}

		if len(result.funcCalls) == 0 {
			if optAgentType == pconfig.OptionsTypeAssistant {
				return fp.processAssistantResult(ctx, logger, chainID, chain, result, summarizer, summarizerHandler)
			} else {
				result, err = fp.performReflector(
					ctx, optAgentType, chainID, taskID, subtaskID,
					append(chain, llms.TextParts(llms.ChatMessageTypeAI, result.content)),
					fp.getLastHumanMessage(chain), result.content, executionContext, executor, 1)
				if err != nil {
					fields := make(logrus.Fields)
					if result != nil {
						fields["content"] = result.content[:min(1000, len(result.content))]
						fields["thinking"] = result.thinking[:min(1000, len(result.thinking))]
						fields["execution"] = executionContext[:min(1000, len(executionContext))]
					}
					logger.WithError(err).WithFields(fields).Error("failed to perform reflector")
					return err
				}
			}
		}

		msg := llms.MessageContent{Role: llms.ChatMessageTypeAI}
		for _, toolCall := range result.funcCalls {
			msg.Parts = append(msg.Parts, toolCall)
		}
		chain = append(chain, msg)

		if err := fp.updateMsgChain(ctx, chainID, chain); err != nil {
			logger.WithError(err).Error("failed to update msg chain")
			return err
		}

		for idx, toolCall := range result.funcCalls {
			if toolCall.FunctionCall == nil {
				continue
			}

			funcName := toolCall.FunctionCall.Name
			response, err := fp.execToolCall(ctx, chainID, idx, result, detector, executor)
			if err != nil {
				logger.WithError(err).WithFields(logrus.Fields{
					"func_name": funcName,
					"func_args": toolCall.FunctionCall.Arguments,
				}).Error("failed to exec tool call")
				return err
			}

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
			if err := fp.updateMsgChain(ctx, chainID, chain); err != nil {
				logger.WithError(err).Error("failed to update msg chain")
				return err
			}

			if executor.IsBarrierFunction(funcName) {
				wantToStop = true
			}
		}

		if wantToStop {
			return nil
		}

		if summarizer != nil {
			// it returns the same chain state if error occurs
			chain, err = summarizer.SummarizeChain(ctx, summarizerHandler, chain)
			if err != nil {
				logger.WithError(err).Warn("failed to summarize chain")
			} else if err := fp.updateMsgChain(ctx, chainID, chain); err != nil {
				logger.WithError(err).Error("failed to update msg chain")
				return err
			}
		}
	}
}

func (fp *flowProvider) execToolCall(
	ctx context.Context,
	chainID int64,
	toolCallIDx int,
	result *callResult,
	detector *repeatingDetector,
	executor tools.ContextToolsExecutor,
) (string, error) {
	var (
		streamID int64
		thinking string
	)

	// use streamID and thinking only for first tool call to minimize content
	if toolCallIDx == 0 {
		streamID = result.streamID
		thinking = result.thinking
	}

	toolCall := result.funcCalls[toolCallIDx]
	funcName := toolCall.FunctionCall.Name
	funcArgs := json.RawMessage(toolCall.FunctionCall.Arguments)

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"agent":        fp.Type(),
		"flow_id":      fp.flowID,
		"func_name":    funcName,
		"func_args":    string(funcArgs)[:min(1000, len(funcArgs))],
		"tool_call_id": toolCall.ID,
		"msg_chain_id": chainID,
	})

	ctx, observation := obs.Observer.NewObservation(ctx)
	opts := []langfuse.EventStartOption{
		langfuse.WithStartEventName(fmt.Sprintf("tool call %s", funcName)),
		langfuse.WithStartEventInput(funcArgs),
		langfuse.WithStartEventMetadata(map[string]any{
			"tool_call_id": toolCall.ID,
			"tool_name":    funcName,
		}),
	}

	if detector.detect(toolCall) {
		response := fmt.Sprintf("tool call '%s' is repeating, please try another tool", funcName)

		observation.Event(append(opts,
			langfuse.WithStartEventStatus("failed"),
			langfuse.WithStartEventLevel(langfuse.ObservationLevelError),
			langfuse.WithStartEventOutput(response),
		)...)
		logger.Warn("failed to exec function: tool call is repeating")

		return response, nil
	}

	var (
		err      error
		response string
	)

	for idx := 0; idx <= maxRetriesToCallFunction; idx++ {
		if idx == maxRetriesToCallFunction {
			err = fmt.Errorf("reached max retries to call function: %w", err)
			logger.WithError(err).Error("failed to exec function")
			return "", fmt.Errorf("failed to exec function '%s': %w", funcName, err)
		}

		response, err = executor.Execute(ctx, streamID, toolCall.ID, funcName, thinking, funcArgs)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return "", err
			}

			observation.Event(append(opts,
				langfuse.WithStartEventStatus(err.Error()),
				langfuse.WithStartEventLevel(langfuse.ObservationLevelError),
			)...)
			logger.WithError(err).Warn("failed to exec function")

			funcExecErr := err
			funcSchema, err := executor.GetToolSchema(funcName)
			if err != nil {
				logger.WithError(err).Error("failed to get tool schema")
				return "", fmt.Errorf("failed to get tool schema: %w", err)
			}

			funcArgs, err = fp.fixToolCallArgs(ctx, funcName, funcArgs, funcSchema, funcExecErr)
			if err != nil {
				logger.WithError(err).Error("failed to fix tool call args")
				return "", fmt.Errorf("failed to fix tool call args: %w", err)
			}
		} else {
			break
		}
	}

	observation.Event(append(opts,
		langfuse.WithStartEventStatus("success"),
		langfuse.WithStartEventOutput(response),
	)...)

	return response, nil
}

func (fp *flowProvider) callWithRetries(
	ctx context.Context,
	chain []llms.MessageContent,
	optAgentType pconfig.ProviderOptionsType,
	executor tools.ContextToolsExecutor,
) (*callResult, error) {
	var (
		err     error
		errs    []error
		msgType = database.MsglogTypeAnswer
		resp    *llms.ContentResponse
		result  callResult
	)

	ticker := time.NewTicker(delayBetweenRetries)
	defer ticker.Stop()

	fillResult := func(resp *llms.ContentResponse) error {
		var parts []string

		if resp == nil || len(resp.Choices) == 0 {
			return fmt.Errorf("no choices in response")
		}

		for _, choice := range resp.Choices {
			if strings.TrimSpace(choice.Content) != "" {
				parts = append(parts, choice.Content)
			}

			if choice.GenerationInfo != nil {
				result.info = choice.GenerationInfo
			}

			for _, toolCall := range choice.ToolCalls {
				if toolCall.FunctionCall == nil {
					continue
				}
				result.funcCalls = append(result.funcCalls, toolCall)
			}

			if choice.ReasoningContent != "" {
				result.thinking = choice.ReasoningContent
			}
		}

		result.content = strings.Join(parts, "\n")
		if strings.Trim(result.content, "' \"\n\r\t") == "" && len(result.funcCalls) == 0 {
			return fmt.Errorf("no content and tool calls in response")
		}

		return nil
	}

	for idx := 0; idx <= maxRetriesToCallAgentChain; idx++ {
		if idx == maxRetriesToCallAgentChain {
			msg := fmt.Sprintf("failed to call agent chain: max retries reached, %d", idx)
			return nil, fmt.Errorf(msg+": %w", errors.Join(errs...))
		}

		var streamCb streaming.Callback
		if fp.streamCb != nil {
			result.streamID = fp.callCounter.Add(1)
			streamCb = func(ctx context.Context, chunk streaming.Chunk) error {
				switch chunk.Type {
				case streaming.ChunkTypeReasoning:
					return fp.streamCb(ctx, &StreamMessageChunk{
						Type:     StreamMessageChunkTypeThinking,
						MsgType:  msgType,
						Thinking: chunk.ReasoningContent,
						StreamID: result.streamID,
					})
				case streaming.ChunkTypeText:
					return fp.streamCb(ctx, &StreamMessageChunk{
						Type:     StreamMessageChunkTypeContent,
						MsgType:  msgType,
						Content:  chunk.Content,
						StreamID: result.streamID,
					})
				case streaming.ChunkTypeToolCall:
					// skip tool call chunks (we don't need them for now)
				case streaming.ChunkTypeDone:
					return fp.streamCb(ctx, &StreamMessageChunk{
						Type:     StreamMessageChunkTypeFlush,
						MsgType:  msgType,
						StreamID: result.streamID,
					})
				}
				return nil
			}
		}

		resp, err = fp.CallWithTools(ctx, optAgentType, chain, executor.Tools(), streamCb)
		if err == nil {
			err = fillResult(resp)
		}
		if err == nil {
			break
		} else {
			errs = append(errs, err)
		}

		ticker.Reset(delayBetweenRetries)
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return nil, fmt.Errorf("context canceled while waiting for retry: %w", ctx.Err())
		}
	}

	if fp.streamCb != nil && result.streamID != 0 {
		fp.streamCb(ctx, &StreamMessageChunk{
			Type:     StreamMessageChunkTypeUpdate,
			MsgType:  msgType,
			Content:  result.content,
			Thinking: result.thinking,
			StreamID: result.streamID,
		})
		// don't update stream by ID if we got content separately from tool calls
		// because we stored thinking and content into standalone messages
		if len(result.funcCalls) > 0 && result.content != "" {
			result.streamID = 0
			result.thinking = ""
		}
	}

	return &result, nil
}

func (fp *flowProvider) performReflector(
	ctx context.Context,
	optOriginType pconfig.ProviderOptionsType,
	chainID int64,
	taskID, subtaskID *int64,
	chain []llms.MessageContent,
	humanMessage, content, executionContext string,
	executor tools.ContextToolsExecutor,
	iteration int,
) (*callResult, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.performReflector")
	defer span.End()

	var (
		optAgentType = pconfig.OptionsTypeReflector
		msgChainType = database.MsgchainTypeReflector
	)

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"agent":      fp.Type(),
		"flow_id":    fp.flowID,
		"task_id":    taskID,
		"subtask_id": subtaskID,
		"iteration":  iteration,
	})

	if iteration > maxReflectorCallsPerChain {
		msg := "reflector called too many times"
		_, observation := obs.Observer.NewObservation(ctx)
		observation.Event(
			langfuse.WithStartEventName("reflector limit calls reached"),
			langfuse.WithStartEventInput(content),
			langfuse.WithStartEventStatus("failed"),
			langfuse.WithStartEventLevel(langfuse.ObservationLevelError),
			langfuse.WithStartEventOutput(msg),
		)
		logger.WithField("content", content[:min(1000, len(content))]).Warn(msg)
		return nil, errors.New(msg)
	}

	logger.WithField("content", content[:min(1000, len(content))]).Warn("got message instead of tool call")

	reflectorContext := map[string]map[string]any{
		"user": {
			"Message":          content,
			"BarrierToolNames": executor.GetBarrierToolNames(),
		},
		"system": {
			"BarrierTools":     executor.GetBarrierTools(),
			"CurrentTime":      getCurrentTime(),
			"ExecutionContext": executionContext,
		},
	}

	if humanMessage != "" {
		reflectorContext["system"]["Request"] = humanMessage
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	reflectorSpan := observation.Span(
		langfuse.WithStartSpanName("reflector agent"),
		langfuse.WithStartSpanInput(content),
		langfuse.WithStartSpanMetadata(langfuse.Metadata{
			"user_context":   reflectorContext["user"],
			"system_context": reflectorContext["system"],
		}),
	)
	ctx, _ = reflectorSpan.Observation(ctx)

	userReflectorTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeQuestionReflector, reflectorContext["user"])
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, reflectorSpan, "failed to get user reflector template", err)
	}

	systemReflectorTmpl, err := fp.prompter.RenderTemplate(templates.PromptTypeReflector, reflectorContext["system"])
	if err != nil {
		return nil, wrapErrorEndSpan(ctx, reflectorSpan, "failed to get system reflector template", err)
	}

	advice, err := fp.performSimpleChain(ctx, taskID, subtaskID, optAgentType,
		msgChainType, systemReflectorTmpl, userReflectorTmpl)
	if err != nil {
		advice = ToolPlaceholder
	}

	opts := []langfuse.SpanEndOption{
		langfuse.WithEndSpanStatus("failed"),
		langfuse.WithEndSpanOutput(advice),
		langfuse.WithEndSpanLevel(langfuse.ObservationLevelWarning),
	}
	defer reflectorSpan.End(opts...)

	chain = append(chain, llms.TextParts(llms.ChatMessageTypeHuman, advice))
	result, err := fp.callWithRetries(ctx, chain, optOriginType, executor)
	if err != nil {
		logger.WithError(err).Error("failed to call agent chain by reflector")
		opts = append(opts, langfuse.WithEndSpanStatus(err.Error()))
		return nil, err
	}

	if err := fp.updateMsgChainUsage(ctx, chainID, result.info); err != nil {
		logger.WithError(err).Error("failed to update msg chain usage")
		return nil, err
	}

	chain = append(chain, llms.TextParts(llms.ChatMessageTypeAI, result.content))
	if len(result.funcCalls) == 0 {
		return fp.performReflector(ctx, optOriginType, chainID, taskID, subtaskID, chain,
			humanMessage, result.content, executionContext, executor, iteration+1)
	}

	opts = append(opts, langfuse.WithEndSpanStatus("success"))
	return result, nil
}

func (fp *flowProvider) getLastHumanMessage(chain []llms.MessageContent) string {
	ast, err := cast.NewChainAST(chain, true)
	if err != nil {
		return ""
	}

	slices.Reverse(ast.Sections)
	for _, section := range ast.Sections {
		if section.Header.HumanMessage != nil {
			var hparts []string
			for _, part := range section.Header.HumanMessage.Parts {
				if text, ok := part.(llms.TextContent); ok {
					hparts = append(hparts, text.Text)
				}
			}
			return strings.Join(hparts, "\n")
		}
	}

	return ""
}

func (fp *flowProvider) processAssistantResult(
	ctx context.Context,
	logger *logrus.Entry,
	chainID int64,
	chain []llms.MessageContent,
	result *callResult,
	summarizer csum.Summarizer,
	summarizerHandler tools.SummarizeHandler,
) error {
	var err error

	if fp.streamCb != nil {
		if result.streamID == 0 {
			result.streamID = fp.callCounter.Add(1)
		}
		err := fp.streamCb(ctx, &StreamMessageChunk{
			Type:     StreamMessageChunkTypeUpdate,
			MsgType:  database.MsglogTypeAnswer,
			Content:  result.content,
			Thinking: result.thinking,
			StreamID: result.streamID,
		})
		if err != nil {
			return fmt.Errorf("failed to stream assistant result: %w", err)
		}
	}

	if summarizer != nil {
		// it returns the same chain state if error occurs
		chain, err = summarizer.SummarizeChain(ctx, summarizerHandler, chain)
		if err != nil {
			logger.WithError(err).Warn("failed to summarize chain")
		}
	}

	chain = append(chain, llms.TextParts(llms.ChatMessageTypeAI, result.content))
	if err := fp.updateMsgChain(ctx, chainID, chain); err != nil {
		return fmt.Errorf("failed to update msg chain: %w", err)
	}

	return nil
}

func (fp *flowProvider) updateMsgChain(ctx context.Context, chainID int64, chain []llms.MessageContent) error {
	chainBlob, err := json.Marshal(chain)
	if err != nil {
		return fmt.Errorf("failed to marshal msg chain: %w", err)
	}

	_, err = fp.db.UpdateMsgChain(ctx, database.UpdateMsgChainParams{
		Chain: chainBlob,
		ID:    chainID,
	})
	if err != nil {
		return fmt.Errorf("failed to update msg chain in DB: %w", err)
	}

	return nil
}

func (fp *flowProvider) updateMsgChainUsage(ctx context.Context, chainID int64, info map[string]any) error {
	inputTokens, outputTokens := int64(0), int64(0)
	if info != nil {
		inputTokens, outputTokens = fp.GetUsage(info)
	}

	_, err := fp.db.UpdateMsgChainUsage(ctx, database.UpdateMsgChainUsageParams{
		UsageIn:  inputTokens,
		UsageOut: outputTokens,
		ID:       chainID,
	})
	if err != nil {
		return fmt.Errorf("failed to update msg chain usage in DB: %w", err)
	}

	return nil
}
