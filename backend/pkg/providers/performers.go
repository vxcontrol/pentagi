package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"pentagi/pkg/cast"
	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/templates"
	"pentagi/pkg/tools"

	"github.com/sirupsen/logrus"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func (fp *flowProvider) performTaskResultReporter(
	ctx context.Context,
	taskID, subtaskID *int64,
	systemReporterTmpl, userReporterTmpl, input string,
) (*tools.TaskResult, error) {
	var (
		taskResult   tools.TaskResult
		optAgentType = pconfig.OptionsTypeSimple
		msgChainType = database.MsgchainTypeReporter
	)

	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemReporterTmpl),
		llms.TextParts(llms.ChatMessageTypeHuman, userReporterTmpl),
	}

	ctx = tools.PutAgentContext(ctx, msgChainType)
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

	msgChain, err := fp.db.CreateMsgChain(ctx, database.CreateMsgChainParams{
		Type:          msgChainType,
		Model:         fp.Model(optAgentType),
		ModelProvider: string(fp.Type()),
		Chain:         chainBlob,
		FlowID:        fp.flowID,
		TaskID:        database.Int64ToNullInt64(taskID),
		SubtaskID:     database.Int64ToNullInt64(subtaskID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create msg chain: %w", err)
	}

	err = fp.performAgentChain(ctx, optAgentType, msgChain.ID, taskID, subtaskID, chain, executor, fp.summarizer)
	if err != nil {
		return nil, fmt.Errorf("failed to get task reporter result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.putAgentLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			input,
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
	systemGeneratorTmpl, userGeneratorTmpl, input string,
) ([]tools.SubtaskInfo, error) {
	var (
		subtaskList  tools.SubtaskList
		optAgentType = pconfig.OptionsTypeGenerator
		msgChainType = database.MsgchainTypeGenerator
	)

	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemGeneratorTmpl),
		llms.TextParts(llms.ChatMessageTypeHuman, userGeneratorTmpl),
	}

	memorist, err := fp.GetMemoristHandler(ctx, &taskID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get memorist handler: %w", err)
	}

	searcher, err := fp.GetTaskSearcherHandler(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get searcher handler: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, msgChainType)
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

	msgChain, err := fp.db.CreateMsgChain(ctx, database.CreateMsgChainParams{
		Type:          msgChainType,
		Model:         fp.Model(optAgentType),
		ModelProvider: string(fp.Type()),
		Chain:         chainBlob,
		FlowID:        fp.flowID,
		TaskID:        database.Int64ToNullInt64(&taskID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create msg chain: %w", err)
	}

	err = fp.performAgentChain(ctx, optAgentType, msgChain.ID, &taskID, nil, chain, executor, fp.summarizer)
	if err != nil {
		return nil, fmt.Errorf("failed to get subtasks generator result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.putAgentLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			input,
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
	plannedSubtasks []database.Subtask,
	systemRefinerTmpl, userRefinerTmpl, input string,
) ([]tools.SubtaskInfo, error) {
	var (
		subtaskPatch tools.SubtaskPatch
		chain        []llms.MessageContent
		optAgentType = pconfig.OptionsTypeRefiner
		msgChainType = database.MsgchainTypeRefiner
	)

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"task_id":        taskID,
		"planned_count":  len(plannedSubtasks),
		"msg_chain_type": msgChainType,
		"opt_agent_type": optAgentType,
	})

	logger.Debug("starting subtasks refiner")

	// Track execution time for duration calculation
	startTime := time.Now()

	restoreChain := func(msgChain json.RawMessage) ([]llms.MessageContent, error) {
		var msgList []llms.MessageContent
		err := json.Unmarshal(msgChain, &msgList)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal msg chain: %w", err)
		}

		ast, err := cast.NewChainAST(msgList, true)
		if err != nil {
			return nil, fmt.Errorf("failed to create refiner chain ast: %w", err)
		}

		if len(ast.Sections) == 0 {
			return nil, fmt.Errorf("failed to get sections from refiner chain ast")
		}

		systemSection := ast.Sections[0] // there may be multiple sections due to reflector agent
		systemMessage := llms.TextParts(llms.ChatMessageTypeSystem, systemRefinerTmpl)
		systemSection.Header.SystemMessage = &systemMessage

		// remove the last report with subtasks list/patch
		for idx := len(systemSection.Body) - 1; idx >= 0; idx-- {
			if systemSection.Body[idx].Type == cast.RequestResponse {
				systemSection.Body = systemSection.Body[:idx]
				break
			}
		}

		// build human message with tool calls history
		// we combine the history into single part for better LLMs compatibility
		toolCalls := extractToolCallsFromChain(systemSection.Messages())
		toolCallsHistory := extractHistoryFromHumanMessage(systemSection.Header.HumanMessage)
		combinedToolCallsHistory := appendNewToolCallsToHistory(toolCallsHistory, toolCalls)
		combinedUserRefinerTmpl := combineHistoryToolCallsToHumanMessage(combinedToolCallsHistory, userRefinerTmpl)
		humanMessage := llms.TextParts(llms.ChatMessageTypeHuman, combinedUserRefinerTmpl)
		systemSection.Header.HumanMessage = &humanMessage

		// reset messages in the chain, it's already saved in the header
		systemSection.Body = []*cast.BodyPair{}

		// restore the chain
		return systemSection.Messages(), nil
	}

	msgChain, err := fp.db.GetFlowTaskTypeLastMsgChain(ctx, database.GetFlowTaskTypeLastMsgChainParams{
		FlowID: fp.flowID,
		TaskID: database.Int64ToNullInt64(&taskID),
		Type:   msgChainType,
	})
	if err != nil || isEmptyChain(msgChain.Chain) {
		// fallback to generator chain if refiner chain is not found or empty
		msgChain, err = fp.db.GetFlowTaskTypeLastMsgChain(ctx, database.GetFlowTaskTypeLastMsgChainParams{
			FlowID: fp.flowID,
			TaskID: database.Int64ToNullInt64(&taskID),
			Type:   database.MsgchainTypeGenerator,
		})
		if err != nil || isEmptyChain(msgChain.Chain) {
			// is unexpected, but we should fallback to empty chain
			chain = []llms.MessageContent{
				llms.TextParts(llms.ChatMessageTypeSystem, systemRefinerTmpl),
				llms.TextParts(llms.ChatMessageTypeHuman, userRefinerTmpl),
			}
		} else {
			if chain, err = restoreChain(msgChain.Chain); err != nil {
				return nil, fmt.Errorf("failed to restore chain from generator state: %w", err)
			}
		}
	} else {
		if chain, err = restoreChain(msgChain.Chain); err != nil {
			return nil, fmt.Errorf("failed to restore chain from refiner state: %w", err)
		}
	}

	memorist, err := fp.GetMemoristHandler(ctx, &taskID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get memorist handler: %w", err)
	}

	searcher, err := fp.GetTaskSearcherHandler(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get searcher handler: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, msgChainType)
	cfg := tools.RefinerExecutorConfig{
		TaskID:   taskID,
		Memorist: memorist,
		Searcher: searcher,
		SubtaskPatch: func(ctx context.Context, name string, args json.RawMessage) (string, error) {
			logger.WithField("args_len", len(args)).Debug("received subtask patch")
			if err := json.Unmarshal(args, &subtaskPatch); err != nil {
				logger.WithError(err).Error("failed to unmarshal subtask patch")
				return "", fmt.Errorf("failed to unmarshal subtask patch: %w", err)
			}
			if err := ValidateSubtaskPatch(subtaskPatch); err != nil {
				logger.WithError(err).Error("invalid subtask patch")
				return "", fmt.Errorf("invalid subtask patch: %w", err)
			}
			logger.WithField("operations_count", len(subtaskPatch.Operations)).Debug("subtask patch validated")
			return "subtask patch successfully processed", nil
		},
	}
	executor, err := fp.executor.GetRefinerExecutor(cfg)
	if err != nil {
		logger.WithError(err).Error("failed to get refiner executor")
		return nil, fmt.Errorf("failed to get refiner executor: %w", err)
	}

	chainBlob, err := json.Marshal(chain)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal msg chain: %w", err)
	}

	msgChain, err = fp.db.CreateMsgChain(ctx, database.CreateMsgChainParams{
		Type:            msgChainType,
		Model:           fp.Model(optAgentType),
		ModelProvider:   string(fp.Type()),
		Chain:           chainBlob,
		FlowID:          fp.flowID,
		TaskID:          database.Int64ToNullInt64(&taskID),
		DurationSeconds: time.Since(startTime).Seconds(),
	})
	if err != nil {
		logger.WithError(err).Error("failed to create msg chain")
		return nil, fmt.Errorf("failed to create msg chain: %w", err)
	}

	logger.WithField("msg_chain_id", msgChain.ID).Debug("created msg chain for refiner")

	err = fp.performAgentChain(ctx, optAgentType, msgChain.ID, &taskID, nil, chain, executor, fp.summarizer)
	if err != nil {
		logger.WithError(err).Error("failed to perform subtasks refiner agent chain")
		return nil, fmt.Errorf("failed to get subtasks refiner result: %w", err)
	}

	// Apply the patch operations to the planned subtasks
	result, err := applySubtaskOperations(plannedSubtasks, subtaskPatch, logger)
	if err != nil {
		logger.WithError(err).Error("failed to apply subtask operations")
		return nil, fmt.Errorf("failed to apply subtask operations: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"input_count":  len(plannedSubtasks),
		"output_count": len(result),
		"operations":   len(subtaskPatch.Operations),
	}).Debug("successfully applied subtask patch")

	subtasks := convertSubtaskInfoPatch(result)
	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.putAgentLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			input,
			fp.subtasksToMarkdown(subtasks),
			&taskID,
			nil,
		)
	}

	return subtasks, nil
}

func (fp *flowProvider) performCoder(
	ctx context.Context,
	taskID, subtaskID *int64,
	systemCoderTmpl, userCoderTmpl, question string,
) (string, error) {
	var (
		codeResult   tools.CodeResult
		optAgentType = pconfig.OptionsTypeCoder
		msgChainType = database.MsgchainTypeCoder
	)

	adviser, err := fp.GetAskAdviceHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get adviser handler: %w", err)
	}

	installer, err := fp.GetInstallerHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get installer handler: %w", err)
	}

	memorist, err := fp.GetMemoristHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get memorist handler: %w", err)
	}

	searcher, err := fp.GetSubtaskSearcherHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get searcher handler: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, msgChainType)
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
		Summarizer: fp.GetSummarizeResultHandler(taskID, subtaskID),
	}
	executor, err := fp.executor.GetCoderExecutor(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get coder executor: %w", err)
	}

	if fp.planning {
		userCoderTmplWithPlan, err := fp.performPlanner(
			ctx, taskID, subtaskID, optAgentType, executor, userCoderTmpl, question,
		)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Warn("failed to get task plan from planner, proceeding without plan")
		} else {
			userCoderTmpl = userCoderTmplWithPlan
		}
	}

	msgChainID, chain, err := fp.restoreChain(
		ctx, taskID, subtaskID, optAgentType, msgChainType, systemCoderTmpl, userCoderTmpl,
	)
	if err != nil {
		return "", fmt.Errorf("failed to restore chain: %w", err)
	}

	err = fp.performAgentChain(ctx, optAgentType, msgChainID, taskID, subtaskID, chain, executor, fp.summarizer)
	if err != nil {
		return "", fmt.Errorf("failed to get task coder result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.putAgentLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			question,
			codeResult.Result,
			taskID,
			subtaskID,
		)
	}

	return codeResult.Result, nil
}

func (fp *flowProvider) performInstaller(
	ctx context.Context,
	taskID, subtaskID *int64,
	systemInstallerTmpl, userInstallerTmpl, question string,
) (string, error) {
	var (
		maintenanceResult tools.MaintenanceResult
		optAgentType      = pconfig.OptionsTypeInstaller
		msgChainType      = database.MsgchainTypeInstaller
	)

	adviser, err := fp.GetAskAdviceHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get adviser handler: %w", err)
	}

	memorist, err := fp.GetMemoristHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get memorist handler: %w", err)
	}

	searcher, err := fp.GetSubtaskSearcherHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get searcher handler: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, msgChainType)
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
		Summarizer: fp.GetSummarizeResultHandler(taskID, subtaskID),
	}
	executor, err := fp.executor.GetInstallerExecutor(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get installer executor: %w", err)
	}

	if fp.planning {
		userInstallerTmplWithPlan, err := fp.performPlanner(
			ctx, taskID, subtaskID, optAgentType, executor, userInstallerTmpl, question,
		)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Warn("failed to get task plan from planner, proceeding without plan")
		} else {
			userInstallerTmpl = userInstallerTmplWithPlan
		}
	}

	msgChainID, chain, err := fp.restoreChain(
		ctx, taskID, subtaskID, optAgentType, msgChainType, systemInstallerTmpl, userInstallerTmpl,
	)
	if err != nil {
		return "", fmt.Errorf("failed to restore chain: %w", err)
	}

	err = fp.performAgentChain(ctx, optAgentType, msgChainID, taskID, subtaskID, chain, executor, fp.summarizer)
	if err != nil {
		return "", fmt.Errorf("failed to get task installer result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.putAgentLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			question,
			maintenanceResult.Result,
			taskID,
			subtaskID,
		)
	}

	return maintenanceResult.Result, nil
}

func (fp *flowProvider) performMemorist(
	ctx context.Context,
	taskID, subtaskID *int64,
	systemMemoristTmpl, userMemoristTmpl, question string,
) (string, error) {
	var (
		memoristResult tools.MemoristResult
		optAgentType   = pconfig.OptionsTypeSearcher
		msgChainType   = database.MsgchainTypeMemorist
	)

	ctx = tools.PutAgentContext(ctx, msgChainType)
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
		Summarizer: fp.GetSummarizeResultHandler(taskID, subtaskID),
	}
	executor, err := fp.executor.GetMemoristExecutor(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get memorist executor: %w", err)
	}

	msgChainID, chain, err := fp.restoreChain(
		ctx, taskID, subtaskID, optAgentType, msgChainType, systemMemoristTmpl, userMemoristTmpl,
	)
	if err != nil {
		return "", fmt.Errorf("failed to restore chain: %w", err)
	}

	err = fp.performAgentChain(ctx, optAgentType, msgChainID, taskID, subtaskID, chain, executor, fp.summarizer)
	if err != nil {
		return "", fmt.Errorf("failed to get task memorist result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.putAgentLog(
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
	taskID, subtaskID *int64,
	systemPentesterTmpl, userPentesterTmpl, question string,
) (string, error) {
	var (
		hackResult   tools.HackResult
		optAgentType = pconfig.OptionsTypePentester
		msgChainType = database.MsgchainTypePentester
	)

	adviser, err := fp.GetAskAdviceHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get adviser handler: %w", err)
	}

	coder, err := fp.GetCoderHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get coder handler: %w", err)
	}

	installer, err := fp.GetInstallerHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get installer handler: %w", err)
	}

	memorist, err := fp.GetMemoristHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get memorist handler: %w", err)
	}

	searcher, err := fp.GetSubtaskSearcherHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get searcher handler: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, msgChainType)
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
		Summarizer: fp.GetSummarizeResultHandler(taskID, subtaskID),
	}
	executor, err := fp.executor.GetPentesterExecutor(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get pentester executor: %w", err)
	}

	if fp.planning {
		userPentesterTmplWithPlan, err := fp.performPlanner(
			ctx, taskID, subtaskID, optAgentType, executor, userPentesterTmpl, question,
		)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Warn("failed to get task plan from planner, proceeding without plan")
		} else {
			userPentesterTmpl = userPentesterTmplWithPlan
		}
	}

	msgChainID, chain, err := fp.restoreChain(
		ctx, taskID, subtaskID, optAgentType, msgChainType, systemPentesterTmpl, userPentesterTmpl,
	)
	if err != nil {
		return "", fmt.Errorf("failed to restore chain: %w", err)
	}

	err = fp.performAgentChain(ctx, optAgentType, msgChainID, taskID, subtaskID, chain, executor, fp.summarizer)
	if err != nil {
		return "", fmt.Errorf("failed to get task pentester result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.putAgentLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			question,
			hackResult.Result,
			taskID,
			subtaskID,
		)
	}

	return hackResult.Result, nil
}

func (fp *flowProvider) performSearcher(
	ctx context.Context,
	taskID, subtaskID *int64,
	systemSearcherTmpl, userSearcherTmpl, question string,
) (string, error) {
	var (
		searchResult tools.SearchResult
		optAgentType = pconfig.OptionsTypeSearcher
		msgChainType = database.MsgchainTypeSearcher
	)

	memorist, err := fp.GetMemoristHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get memorist handler: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, msgChainType)
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
		Summarizer: fp.GetSummarizeResultHandler(taskID, subtaskID),
	}
	executor, err := fp.executor.GetSearcherExecutor(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get searcher executor: %w", err)
	}

	msgChainID, chain, err := fp.restoreChain(
		ctx, taskID, subtaskID, optAgentType, msgChainType, systemSearcherTmpl, userSearcherTmpl,
	)
	if err != nil {
		return "", fmt.Errorf("failed to restore chain: %w", err)
	}

	err = fp.performAgentChain(ctx, optAgentType, msgChainID, taskID, subtaskID, chain, executor, fp.summarizer)
	if err != nil {
		return "", fmt.Errorf("failed to get task searcher result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.putAgentLog(
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
	taskID, subtaskID *int64,
	systemEnricherTmpl, userEnricherTmpl, question string,
) (string, error) {
	var (
		enricherResult tools.EnricherResult
		optAgentType   = pconfig.OptionsTypeEnricher
		msgChainType   = database.MsgchainTypeEnricher
	)

	ctx = tools.PutAgentContext(ctx, msgChainType)
	cfg := tools.EnricherExecutorConfig{
		TaskID:    taskID,
		SubtaskID: subtaskID,
		EnricherResult: func(ctx context.Context, name string, args json.RawMessage) (string, error) {
			err := json.Unmarshal(args, &enricherResult)
			if err != nil {
				return "", fmt.Errorf("failed to unmarshal result: %w", err)
			}
			return "enrich result successfully processed", nil
		},
		Summarizer: fp.GetSummarizeResultHandler(taskID, subtaskID),
	}
	executor, err := fp.executor.GetEnricherExecutor(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to get enricher executor: %w", err)
	}

	msgChainID, chain, err := fp.restoreChain(
		ctx, taskID, subtaskID, optAgentType, msgChainType, systemEnricherTmpl, userEnricherTmpl,
	)
	if err != nil {
		return "", fmt.Errorf("failed to restore chain: %w", err)
	}

	err = fp.performAgentChain(ctx, optAgentType, msgChainID, taskID, subtaskID, chain, executor, fp.summarizer)
	if err != nil {
		return "", fmt.Errorf("failed to get task enricher result: %w", err)
	}

	if agentCtx, ok := tools.GetAgentContext(ctx); ok {
		fp.putAgentLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			question,
			enricherResult.Result,
			taskID,
			subtaskID,
		)
	}

	return enricherResult.Result, nil
}

// performPlanner invokes adviser to create an execution plan for agent tasks
func (fp *flowProvider) performPlanner(
	ctx context.Context,
	taskID, subtaskID *int64,
	opt pconfig.ProviderOptionsType,
	executor tools.ContextToolsExecutor,
	userTmpl, question string,
) (string, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.performPlanner")
	defer span.End()

	toolCallID := templates.GenerateFromPattern(fp.tcIDTemplate, tools.AdviceToolName)
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"task_id":      taskID,
		"subtask_id":   subtaskID,
		"agent_type":   string(opt),
		"tool_call_id": toolCallID,
	})

	logger.Debug("requesting task plan from adviser (planner)")

	// 1. Format Question for task planning
	planQuestionData := map[string]any{
		"AgentType":    string(opt),
		"TaskQuestion": question,
	}

	planQuestion, err := fp.prompter.RenderTemplate(templates.PromptTypeQuestionTaskPlanner, planQuestionData)
	if err != nil {
		return "", fmt.Errorf("failed to render task planner question: %w", err)
	}

	// 2. Call adviser handler with custom observation name "planner"
	askAdvice := tools.AskAdvice{
		Question: planQuestion,
	}

	askAdviceJSON, err := json.Marshal(askAdvice)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ask advice: %w", err)
	}

	logger.Debug("executing adviser handler for task planning")
	plan, err := executor.Execute(ctx, 0, toolCallID, tools.AdviceToolName, "planner", "", askAdviceJSON)
	if err != nil {
		return "", fmt.Errorf("failed to execute adviser handler: %w", err)
	}

	logger.WithField("plan_length", len(plan)).Debug("task plan created successfully")

	// Wrap original request with execution plan using template
	taskAssignment, err := fp.prompter.RenderTemplate(templates.PromptTypeTaskAssignmentWrapper, map[string]any{
		"OriginalRequest": userTmpl,
		"ExecutionPlan":   plan,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render task assignment wrapper: %w", err)
	}

	return taskAssignment, nil
}

// performMentor invokes adviser to monitor agent execution progress
func (fp *flowProvider) performMentor(
	ctx context.Context,
	opt pconfig.ProviderOptionsType,
	chainID int64,
	taskID, subtaskID *int64,
	chain []llms.MessageContent,
	executor tools.ContextToolsExecutor,
	lastToolCall llms.ToolCall,
	lastToolResult string,
) (string, error) {
	ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindInternal, "providers.flowProvider.performMentor")
	defer span.End()

	if lastToolCall.FunctionCall == nil {
		return "", fmt.Errorf("last tool call function call is nil")
	}

	toolCallID := templates.GenerateFromPattern(fp.tcIDTemplate, tools.AdviceToolName)
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"chain_id":       chainID,
		"task_id":        taskID,
		"subtask_id":     subtaskID,
		"last_tool_name": lastToolCall.FunctionCall.Name,
		"agent_type":     string(opt),
		"tool_call_id":   toolCallID,
	})

	logger.Debug("invoking execution adviser for progress monitoring (mentor)")

	// 1. Collect recent messages from chain
	recentMessages := getRecentMessages(chain)

	// 2. Extract all executed tool calls from chain
	executedToolCalls := extractToolCallsFromChain(chain)

	// 3. Get subtask description
	subtaskDesc := ""
	if subtaskID != nil {
		if subtask, err := fp.db.GetSubtask(ctx, *subtaskID); err == nil {
			subtaskDesc = subtask.Description
		}
	}

	// 4. Extract original agent prompt from chain
	agentPrompt := extractAgentPromptFromChain(chain)

	// 5. Format Question through new template
	questionData := map[string]any{
		"SubtaskDescription": subtaskDesc,
		"AgentType":          string(opt),
		"AgentPrompt":        agentPrompt,
		"RecentMessages":     recentMessages,
		"ExecutedToolCalls":  executedToolCalls,
		"LastToolName":       lastToolCall.FunctionCall.Name,
		"LastToolArgs":       formatToolCallArguments(lastToolCall.FunctionCall.Arguments),
		"LastToolResult":     cutString(lastToolResult, 4096),
	}

	question, err := fp.prompter.RenderTemplate(templates.PromptTypeQuestionExecutionMonitor, questionData)
	if err != nil {
		return "", fmt.Errorf("failed to render execution monitor question: %w", err)
	}

	// 6. Call adviser handler with custom observation name "mentor"
	askAdvice := tools.AskAdvice{
		Question: question,
	}

	askAdviceJSON, err := json.Marshal(askAdvice)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ask advice: %w", err)
	}

	logger.Debug("executing adviser handler for execution monitoring")
	result, err := executor.Execute(ctx, 0, toolCallID, tools.AdviceToolName, "mentor", "", askAdviceJSON)
	if err != nil {
		return "", fmt.Errorf("failed to execute adviser handler: %w", err)
	}

	logger.WithField("result_length", len(result)).Debug("execution mentor completed successfully")
	return result, nil
}

func (fp *flowProvider) performSimpleChain(
	ctx context.Context,
	taskID, subtaskID *int64,
	opt pconfig.ProviderOptionsType,
	msgChainType database.MsgchainType,
	systemTmpl, userTmpl string,
) (string, error) {
	var (
		resp *llms.ContentResponse
		err  error
	)

	startTime := time.Now()

	chain := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemTmpl),
		llms.TextParts(llms.ChatMessageTypeHuman, userTmpl),
	}

	for idx := 0; idx <= maxRetriesToCallSimpleChain; idx++ {
		if idx == maxRetriesToCallSimpleChain {
			return "", fmt.Errorf("failed to call simple chain: %w", err)
		}

		resp, err = fp.CallEx(ctx, opt, chain, nil)
		if err == nil {
			break
		} else {
			if errors.Is(err, context.Canceled) {
				return "", err
			}

			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(time.Second * 5):
			default:
			}
		}
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	var parts []string
	var usage pconfig.CallUsage
	var reasoning *reasoning.ContentReasoning
	for _, choice := range resp.Choices {
		parts = append(parts, choice.Content)
		usage.Merge(fp.GetUsage(choice.GenerationInfo))
		// Preserve reasoning from first choice for simple chains (safe for all providers)
		if reasoning == nil && !choice.Reasoning.IsEmpty() {
			reasoning = choice.Reasoning
		}
	}

	// Update cost based on price info
	usage.UpdateCost(fp.GetPriceInfo(opt))

	// Universal pattern for simple chains - preserve reasoning if present
	msg := llms.MessageContent{Role: llms.ChatMessageTypeAI}
	content := strings.Join(parts, "\n")
	if content != "" || reasoning != nil {
		msg.Parts = append(msg.Parts, llms.TextPartWithReasoning(content, reasoning))
	}
	chain = append(chain, msg)

	chainBlob, err := json.Marshal(chain)
	if err != nil {
		return "", fmt.Errorf("failed to marshal summarizer msg chain: %w", err)
	}

	_, err = fp.db.CreateMsgChain(ctx, database.CreateMsgChainParams{
		Type:            msgChainType,
		Model:           fp.Model(opt),
		ModelProvider:   string(fp.Type()),
		UsageIn:         usage.Input,
		UsageOut:        usage.Output,
		UsageCacheIn:    usage.CacheRead,
		UsageCacheOut:   usage.CacheWrite,
		UsageCostIn:     usage.CostInput,
		UsageCostOut:    usage.CostOutput,
		DurationSeconds: time.Since(startTime).Seconds(),
		Chain:           chainBlob,
		FlowID:          fp.flowID,
		TaskID:          database.Int64ToNullInt64(taskID),
		SubtaskID:       database.Int64ToNullInt64(subtaskID),
	})

	return strings.Join(parts, "\n\n"), nil
}
