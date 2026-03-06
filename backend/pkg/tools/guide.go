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
	guideVectorStoreThreshold   = 0.2
	guideVectorStoreResultLimit = 3
	guideVectorStoreDefaultType = "guide"
	guideNotFoundMessage        = "nothing found in guide store and you need to store it after figure out this case"
)

type guide struct {
	flowID    int64
	taskID    *int64
	subtaskID *int64
	store     *pgvector.Store
	vslp      VectorStoreLogProvider
}

func NewGuideTool(flowID int64, taskID, subtaskID *int64, store *pgvector.Store, vslp VectorStoreLogProvider) Tool {
	return &guide{
		flowID:    flowID,
		taskID:    taskID,
		subtaskID: subtaskID,
		store:     store,
		vslp:      vslp,
	}
}

func (g *guide) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(g.flowID, g.taskID, g.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	if g.store == nil {
		logger.Error("pgvector store is not initialized")
		return "", fmt.Errorf("pgvector store is not initialized")
	}

	switch name {
	case SearchGuideToolName:
		var action SearchGuideAction
		if err := json.Unmarshal(args, &action); err != nil {
			logger.WithError(err).Error("failed to unmarshal search guide action")
			return "", fmt.Errorf("failed to unmarshal %s search guide action arguments: %w", name, err)
		}

		filters := map[string]any{
			"doc_type":   guideVectorStoreDefaultType,
			"guide_type": action.Type,
		}

		metadata := langfuse.Metadata{
			"tool_name":  name,
			"message":    action.Message,
			"limit":      guideVectorStoreResultLimit,
			"threshold":  guideVectorStoreThreshold,
			"doc_type":   guideVectorStoreDefaultType,
			"guide_type": action.Type,
		}

		retriever := observation.Retriever(
			langfuse.WithRetrieverName("retrieve guide from vector store"),
			langfuse.WithRetrieverInput(map[string]any{
				"query":       action.Question,
				"threshold":   guideVectorStoreThreshold,
				"max_results": guideVectorStoreResultLimit,
				"filters":     filters,
			}),
			langfuse.WithRetrieverMetadata(metadata),
		)
		ctx, observation = retriever.Observation(ctx)

		logger = logger.WithFields(logrus.Fields{
			"query": action.Question[:min(len(action.Question), 1000)],
			"type":  action.Type,
		})

		docs, err := g.store.SimilaritySearch(
			ctx,
			action.Question,
			guideVectorStoreResultLimit,
			vectorstores.WithScoreThreshold(guideVectorStoreThreshold),
			vectorstores.WithFilters(filters),
		)
		if err != nil {
			retriever.End(
				langfuse.WithRetrieverStatus(err.Error()),
				langfuse.WithRetrieverLevel(langfuse.ObservationLevelError),
			)
			logger.WithError(err).Error("failed to search for similar documents")
			return "", fmt.Errorf("failed to search for similar documents: %w", err)
		}

		if len(docs) == 0 {
			retriever.End(
				langfuse.WithRetrieverStatus("no guide found"),
				langfuse.WithRetrieverLevel(langfuse.ObservationLevelWarning),
				langfuse.WithRetrieverOutput([]any{}),
			)
			observation.Score(
				langfuse.WithScoreComment("no guide found"),
				langfuse.WithScoreName("guide_search_result"),
				langfuse.WithScoreStringValue("not_found"),
			)
			return guideNotFoundMessage, nil
		}

		retriever.End(
			langfuse.WithRetrieverStatus("success"),
			langfuse.WithRetrieverLevel(langfuse.ObservationLevelDebug),
			langfuse.WithRetrieverOutput(docs),
		)

		// TODO: here need to rerank and filter the docs based on the question
		// use evaluator observation type to process each document and to get a score

		buffer := strings.Builder{}
		for i, doc := range docs {
			observation.Score(
				langfuse.WithScoreComment("guide vector store result"),
				langfuse.WithScoreName("guide_search_result"),
				langfuse.WithScoreFloatValue(float64(doc.Score)),
			)
			buffer.WriteString(fmt.Sprintf("# Document %d Match score: %f\n\n", i+1, doc.Score))
			buffer.WriteString(fmt.Sprintf("## Original Guide Type: %s\n\n", doc.Metadata["guide_type"]))
			buffer.WriteString(fmt.Sprintf("## Original Guide Question\n\n%s\n\n", doc.Metadata["question"]))
			buffer.WriteString("## Content\n\n")
			buffer.WriteString(doc.PageContent)
			buffer.WriteString("\n\n")
		}

		if agentCtx, ok := GetAgentContext(ctx); ok {
			filtersData, err := json.Marshal(filters)
			if err != nil {
				logger.WithError(err).Error("failed to marshal filters")
				return "", fmt.Errorf("failed to marshal filters: %w", err)
			}
			_, _ = g.vslp.PutLog(
				ctx,
				agentCtx.ParentAgentType,
				agentCtx.CurrentAgentType,
				filtersData,
				action.Question,
				database.VecstoreActionTypeRetrieve,
				buffer.String(),
				g.taskID,
				g.subtaskID,
			)
		}

		return buffer.String(), nil

	case StoreGuideToolName:
		var action StoreGuideAction
		if err := json.Unmarshal(args, &action); err != nil {
			logger.WithError(err).Error("failed to unmarshal store guide action")
			return "", fmt.Errorf("failed to unmarshal %s store guide action arguments: %w", name, err)
		}

		guide := fmt.Sprintf("Question:\n%s\n\nGuide:\n%s", action.Question, action.Guide)

		opts := []langfuse.EventOption{
			langfuse.WithEventName("store guide to vector store"),
			langfuse.WithEventInput(action.Question),
			langfuse.WithEventOutput(guide),
			langfuse.WithEventMetadata(map[string]any{
				"tool_name":  name,
				"message":    action.Message,
				"doc_type":   guideVectorStoreDefaultType,
				"guide_type": action.Type,
			}),
		}

		logger = logger.WithFields(logrus.Fields{
			"query": action.Question[:min(len(action.Question), 1000)],
			"type":  action.Type,
			"guide": action.Guide[:min(len(action.Guide), 1000)],
		})

		docs, err := documentloaders.NewText(strings.NewReader(guide)).Load(ctx)
		if err != nil {
			observation.Event(append(opts,
				langfuse.WithEventStatus(err.Error()),
				langfuse.WithEventLevel(langfuse.ObservationLevelError),
			)...)
			logger.WithError(err).Error("failed to load document")
			return "", fmt.Errorf("failed to load document: %w", err)
		}

		for _, doc := range docs {
			if doc.Metadata == nil {
				doc.Metadata = map[string]any{}
			}
			doc.Metadata["flow_id"] = g.flowID
			if g.taskID != nil {
				doc.Metadata["task_id"] = *g.taskID
			}
			if g.subtaskID != nil {
				doc.Metadata["subtask_id"] = *g.subtaskID
			}
			doc.Metadata["doc_type"] = guideVectorStoreDefaultType
			doc.Metadata["guide_type"] = action.Type
			doc.Metadata["question"] = action.Question
			doc.Metadata["part_size"] = len(doc.PageContent)
			doc.Metadata["total_size"] = len(action.Guide)
		}

		if _, err := g.store.AddDocuments(ctx, docs); err != nil {
			observation.Event(append(opts,
				langfuse.WithEventStatus(err.Error()),
				langfuse.WithEventLevel(langfuse.ObservationLevelError),
			)...)
			logger.WithError(err).Error("failed to store guide")
			return "", fmt.Errorf("failed to store guide: %w", err)
		}

		observation.Event(append(opts,
			langfuse.WithEventStatus("success"),
			langfuse.WithEventLevel(langfuse.ObservationLevelDebug),
			langfuse.WithEventOutput(docs),
		)...)

		if agentCtx, ok := GetAgentContext(ctx); ok {
			data := map[string]any{
				"doc_type":   guideVectorStoreDefaultType,
				"guide_type": action.Type,
			}
			if g.taskID != nil {
				data["task_id"] = *g.taskID
			}
			if g.subtaskID != nil {
				data["subtask_id"] = *g.subtaskID
			}
			filtersData, err := json.Marshal(data)
			if err != nil {
				logger.WithError(err).Error("failed to marshal filters")
				return "", fmt.Errorf("failed to marshal filters: %w", err)
			}
			_, _ = g.vslp.PutLog(
				ctx,
				agentCtx.ParentAgentType,
				agentCtx.CurrentAgentType,
				filtersData,
				action.Question,
				database.VecstoreActionTypeStore,
				guide,
				g.taskID,
				g.subtaskID,
			)
		}

		return "guide stored successfully", nil

	default:
		logger.Error("unknown tool")
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (g *guide) IsAvailable() bool {
	return g.store != nil
}
