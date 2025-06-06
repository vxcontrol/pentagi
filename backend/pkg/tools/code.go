package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"

	"github.com/sirupsen/logrus"
	"github.com/vxcontrol/langchaingo/documentloaders"
	"github.com/vxcontrol/langchaingo/vectorstores"
	"github.com/vxcontrol/langchaingo/vectorstores/pgvector"
)

const (
	codeVectorStoreThreshold   = 0.2
	codeVectorStoreResultLimit = 3
	codeVectorStoreDefaultType = "code"
	codeNotFoundMessage        = "nothing found in code samples store and you need to store it after figure out this case"
)

type code struct {
	flowID    int64
	taskID    *int64
	subtaskID *int64
	store     *pgvector.Store
	vslp      VectorStoreLogProvider
}

func NewCodeTool(flowID int64, taskID, subtaskID *int64, store *pgvector.Store, vslp VectorStoreLogProvider) Tool {
	return &code{
		flowID:    flowID,
		taskID:    taskID,
		subtaskID: subtaskID,
		store:     store,
		vslp:      vslp,
	}
}

func (c *code) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"tool": name,
		"args": string(args),
	})

	if c.store == nil {
		logger.Error("pgvector store is not initialized")
		return "", fmt.Errorf("pgvector store is not initialized")
	}

	switch name {
	case SearchCodeToolName:
		var action SearchCodeAction
		if err := json.Unmarshal(args, &action); err != nil {
			logger.WithError(err).Error("failed to unmarshal search code action")
			return "", fmt.Errorf("failed to unmarshal %s search code action arguments: %w", name, err)
		}

		opts := []langfuse.EventStartOption{
			langfuse.WithStartEventName("retrieve code samples from vector store"),
			langfuse.WithStartEventInput(action.Question),
			langfuse.WithStartEventMetadata(map[string]any{
				"tool_name": name,
				"code_lang": action.Lang,
				"message":   action.Message,
				"limit":     codeVectorStoreResultLimit,
				"threshold": codeVectorStoreThreshold,
				"doc_type":  codeVectorStoreDefaultType,
			}),
		}

		filters := map[string]any{
			"doc_type":  codeVectorStoreDefaultType,
			"code_lang": action.Lang,
		}

		logger = logger.WithFields(logrus.Fields{
			"query":   action.Question[:min(len(action.Question), 1000)],
			"lang":    action.Lang,
			"filters": filters,
		})

		docs, err := c.store.SimilaritySearch(
			ctx,
			action.Question,
			codeVectorStoreResultLimit,
			vectorstores.WithScoreThreshold(codeVectorStoreThreshold),
			vectorstores.WithFilters(filters),
		)
		if err != nil {
			observation.Event(append(opts,
				langfuse.WithStartEventStatus(err.Error()),
				langfuse.WithStartEventLevel(langfuse.ObservationLevelError),
			)...)
			logger.WithError(err).Error("failed to search code samples for question")
			return "", fmt.Errorf("failed to search code samples for question: %w", err)
		}

		if len(docs) == 0 {
			event := observation.Event(append(opts,
				langfuse.WithStartEventStatus("no code samples found"),
				langfuse.WithStartEventLevel(langfuse.ObservationLevelWarning),
			)...)
			_, observation = event.Observation(ctx)
			observation.Score(
				langfuse.WithScoreComment("no code samples found"),
				langfuse.WithScoreName("code_search_result"),
				langfuse.WithScoreStringValue("not_found"),
			)
			return codeNotFoundMessage, nil
		}

		event := observation.Event(append(opts,
			langfuse.WithStartEventStatus("success"),
			langfuse.WithStartEventLevel(langfuse.ObservationLevelDebug),
			langfuse.WithStartEventOutput(docs),
		)...)
		_, observation = event.Observation(ctx)

		buffer := strings.Builder{}
		for _, doc := range docs {
			observation.Score(
				langfuse.WithScoreComment("code samples vector store result"),
				langfuse.WithScoreName("code_search_result"),
				langfuse.WithScoreFloatValue(float64(doc.Score)),
			)
			buffer.WriteString(fmt.Sprintf("# Original Code Question\n\n%s\n\n", doc.Metadata["question"]))
			buffer.WriteString(fmt.Sprintf("# Original Code Description\n\n%s\n\n", doc.Metadata["description"]))
			buffer.WriteString(doc.PageContent)
			buffer.WriteString("\n\n")
		}

		if agentCtx, ok := GetAgentContext(ctx); ok {
			filtersData, err := json.Marshal(filters)
			if err != nil {
				logger.WithError(err).Error("failed to marshal filters")
				return "", fmt.Errorf("failed to marshal filters: %w", err)
			}
			_, _ = c.vslp.PutLog(
				ctx,
				agentCtx.ParentAgentType,
				agentCtx.CurrentAgentType,
				filtersData,
				action.Question,
				database.VecstoreActionTypeRetrieve,
				buffer.String(),
				c.taskID,
				c.subtaskID,
			)
		}

		return buffer.String(), nil

	case StoreCodeToolName:
		var action StoreCodeAction
		if err := json.Unmarshal(args, &action); err != nil {
			logger.WithError(err).Error("failed to unmarshal store code action")
			return "", fmt.Errorf("failed to unmarshal %s store code action arguments: %w", name, err)
		}

		buffer := strings.Builder{}
		buffer.WriteString(action.Explanation)
		buffer.WriteString(fmt.Sprintf("\n\n```%s\n\n", action.Lang))
		buffer.WriteString(action.Code)
		buffer.WriteString("\n```")

		opts := []langfuse.EventStartOption{
			langfuse.WithStartEventName("store code samples to vector store"),
			langfuse.WithStartEventInput(action.Question),
			langfuse.WithStartEventOutput(buffer.String()),
			langfuse.WithStartEventMetadata(map[string]any{
				"tool_name": name,
				"code_lang": action.Lang,
				"message":   action.Message,
				"doc_type":  codeVectorStoreDefaultType,
			}),
		}

		logger = logger.WithFields(logrus.Fields{
			"query": action.Question[:min(len(action.Question), 1000)],
			"lang":  action.Lang,
			"code":  action.Code[:min(len(action.Code), 1000)],
		})

		docs, err := documentloaders.NewText(strings.NewReader(buffer.String())).Load(ctx)
		if err != nil {
			observation.Event(append(opts,
				langfuse.WithStartEventStatus(err.Error()),
				langfuse.WithStartEventLevel(langfuse.ObservationLevelError),
			)...)
			logger.WithError(err).Error("failed to load document")
			return "", fmt.Errorf("failed to load document: %w", err)
		}

		for _, doc := range docs {
			if doc.Metadata == nil {
				doc.Metadata = map[string]any{}
			}
			doc.Metadata["flow_id"] = c.flowID
			if c.taskID != nil {
				doc.Metadata["task_id"] = *c.taskID
			}
			if c.subtaskID != nil {
				doc.Metadata["subtask_id"] = *c.subtaskID
			}
			doc.Metadata["doc_type"] = codeVectorStoreDefaultType
			doc.Metadata["code_lang"] = action.Lang
			doc.Metadata["question"] = action.Question
			doc.Metadata["description"] = action.Description
			doc.Metadata["part_size"] = len(doc.PageContent)
			doc.Metadata["total_size"] = len(action.Code)
		}

		if _, err := c.store.AddDocuments(ctx, docs); err != nil {
			observation.Event(append(opts,
				langfuse.WithStartEventStatus(err.Error()),
				langfuse.WithStartEventLevel(langfuse.ObservationLevelError),
			)...)
			logger.WithError(err).Error("failed to store code sample")
			return "", fmt.Errorf("failed to store code sample: %w", err)
		}

		observation.Event(append(opts,
			langfuse.WithStartEventStatus("success"),
			langfuse.WithStartEventLevel(langfuse.ObservationLevelDebug),
			langfuse.WithStartEventOutput(docs),
		)...)

		if agentCtx, ok := GetAgentContext(ctx); ok {
			data := map[string]any{
				"doc_type":  codeVectorStoreDefaultType,
				"code_lang": action.Lang,
			}
			if c.taskID != nil {
				data["task_id"] = *c.taskID
			}
			if c.subtaskID != nil {
				data["subtask_id"] = *c.subtaskID
			}
			filtersData, err := json.Marshal(data)
			if err != nil {
				logger.WithError(err).Error("failed to marshal filters")
				return "", fmt.Errorf("failed to marshal filters: %w", err)
			}
			_, _ = c.vslp.PutLog(
				ctx,
				agentCtx.ParentAgentType,
				agentCtx.CurrentAgentType,
				filtersData,
				action.Question,
				database.VecstoreActionTypeStore,
				buffer.String(),
				c.taskID,
				c.subtaskID,
			)
		}

		return "code sample stored successfully", nil

	default:
		logger.Error("unknown tool")
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (c *code) IsAvailable() bool {
	return c.store != nil
}
