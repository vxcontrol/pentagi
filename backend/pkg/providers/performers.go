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
	"pentagi/pkg/providers/pconfig"
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

	ctx = tools.PutAgentContext(ctx, msgChainType)
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

	ctx = tools.PutAgentContext(ctx, msgChainType)
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
		humanMessage := llms.TextParts(llms.ChatMessageTypeHuman, userRefinerTmpl)
		systemSection.Header.HumanMessage = &humanMessage
		// remove the last report with subtasks list/patch
		for idx := len(systemSection.Body) - 1; idx >= 0; idx-- {
			if systemSection.Body[idx].Type == cast.RequestResponse {
				systemSection.Body = systemSection.Body[:idx]
				break
			}
		}
		// remove all past completions
		for idx := len(systemSection.Body) - 1; idx >= 0; idx-- {
			if systemSection.Body[idx].Type != cast.Completion {
				systemSection.Body = systemSection.Body[:idx+1]
				break
			}
		}

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

	ctx = tools.PutAgentContext(ctx, msgChainType)
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

	msgChainID, chain, err := fp.restoreChain(
		ctx, taskID, subtaskID, optAgentType, msgChainType, systemCoderTmpl, userCoderTmpl,
	)
	if err != nil {
		return "", fmt.Errorf("failed to restore chain: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, msgChainType)
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

	msgChainID, chain, err := fp.restoreChain(
		ctx, taskID, subtaskID, optAgentType, msgChainType, systemInstallerTmpl, userInstallerTmpl,
	)
	if err != nil {
		return "", fmt.Errorf("failed to restore chain: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, msgChainType)
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

	ctx = tools.PutAgentContext(ctx, msgChainType)
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

	msgChainID, chain, err := fp.restoreChain(
		ctx, taskID, subtaskID, optAgentType, msgChainType, systemPentesterTmpl, userPentesterTmpl,
	)
	if err != nil {
		return "", fmt.Errorf("failed to restore chain: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, msgChainType)
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

	ctx = tools.PutAgentContext(ctx, msgChainType)
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

	memorist, err := fp.GetMemoristHandler(ctx, taskID, subtaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get memorist handler: %w", err)
	}

	searcher, err := fp.GetSubtaskSearcherHandler(ctx, taskID, subtaskID)
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

	msgChainID, chain, err := fp.restoreChain(
		ctx, taskID, subtaskID, optAgentType, msgChainType, systemEnricherTmpl, userEnricherTmpl,
	)
	if err != nil {
		return "", fmt.Errorf("failed to restore chain: %w", err)
	}

	ctx = tools.PutAgentContext(ctx, msgChainType)
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
