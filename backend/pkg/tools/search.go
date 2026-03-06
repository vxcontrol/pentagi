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
	searchVectorStoreThreshold   = 0.2
	searchVectorStoreResultLimit = 3
	searchVectorStoreDefaultType = "answer"
	searchNotFoundMessage        = "nothing found in answer store and you need to store it after figure out this case"
)

type search struct {
	flowID    int64
	taskID    *int64
	subtaskID *int64
	store     *pgvector.Store
	vslp      VectorStoreLogProvider
}

func NewSearchTool(flowID int64, taskID, subtaskID *int64, store *pgvector.Store, vslp VectorStoreLogProvider) Tool {
	return &search{
		flowID:    flowID,
		taskID:    taskID,
		subtaskID: subtaskID,
		store:     store,
		vslp:      vslp,
	}
}

func (s *search) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(s.flowID, s.taskID, s.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	if s.store == nil {
		logger.Error("pgvector store is not initialized")
		return "", fmt.Errorf("pgvector store is not initialized")
	}

	switch name {
	case SearchAnswerToolName:
		var action SearchAnswerAction
		if err := json.Unmarshal(args, &action); err != nil {
			logger.WithError(err).Error("failed to unmarshal search answer action arguments")
			return "", fmt.Errorf("failed to unmarshal %s search answer action arguments: %w", name, err)
		}

		filters := map[string]any{
			"doc_type":    searchVectorStoreDefaultType,
			"answer_type": action.Type,
		}

		metadata := langfuse.Metadata{
			"tool_name":   name,
			"message":     action.Message,
			"limit":       searchVectorStoreResultLimit,
			"threshold":   searchVectorStoreThreshold,
			"doc_type":    searchVectorStoreDefaultType,
			"answer_type": action.Type,
		}

		retriever := observation.Retriever(
			langfuse.WithRetrieverName("retrieve search answer from vector store"),
			langfuse.WithRetrieverInput(map[string]any{
				"query":       action.Question,
				"threshold":   searchVectorStoreThreshold,
				"max_results": searchVectorStoreResultLimit,
				"filters":     filters,
			}),
			langfuse.WithRetrieverMetadata(metadata),
		)
		ctx, observation = retriever.Observation(ctx)

		logger = logger.WithFields(logrus.Fields{
			"query":       action.Question[:min(len(action.Question), 1000)],
			"answer_type": action.Type,
		})

		docs, err := s.store.SimilaritySearch(
			ctx,
			action.Question,
			searchVectorStoreResultLimit,
			vectorstores.WithScoreThreshold(searchVectorStoreThreshold),
			vectorstores.WithFilters(filters),
		)
		if err != nil {
			retriever.End(
				langfuse.WithRetrieverStatus(err.Error()),
				langfuse.WithRetrieverLevel(langfuse.ObservationLevelError),
			)
			logger.WithError(err).Error("failed to search answer for question")
			return "", fmt.Errorf("failed to search answer for question: %w", err)
		}

		if len(docs) == 0 {
			retriever.End(
				langfuse.WithRetrieverStatus("no search answer found"),
				langfuse.WithRetrieverLevel(langfuse.ObservationLevelWarning),
				langfuse.WithRetrieverOutput([]any{}),
			)
			observation.Score(
				langfuse.WithScoreComment("no search answer found"),
				langfuse.WithScoreName("search_answer_result"),
				langfuse.WithScoreStringValue("not_found"),
			)
			return searchNotFoundMessage, nil
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
				langfuse.WithScoreComment("search answer vector store result"),
				langfuse.WithScoreName("search_answer_result"),
				langfuse.WithScoreFloatValue(float64(doc.Score)),
			)
			buffer.WriteString(fmt.Sprintf("# Document %d Search Score: %f\n\n", i+1, doc.Score))
			buffer.WriteString(fmt.Sprintf("## Original Answer Type: %s\n\n", doc.Metadata["answer_type"]))
			buffer.WriteString(fmt.Sprintf("## Original Search Question\n\n%s\n\n", doc.Metadata["question"]))
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
			_, _ = s.vslp.PutLog(
				ctx,
				agentCtx.ParentAgentType,
				agentCtx.CurrentAgentType,
				filtersData,
				action.Question,
				database.VecstoreActionTypeRetrieve,
				buffer.String(),
				s.taskID,
				s.subtaskID,
			)
		}

		return buffer.String(), nil

	case StoreAnswerToolName:
		var action StoreAnswerAction
		if err := json.Unmarshal(args, &action); err != nil {
			logger.WithError(err).Error("failed to unmarshal search answer action arguments")
			return "", fmt.Errorf("failed to unmarshal %s store answer action arguments: %w", name, err)
		}

		opts := []langfuse.EventOption{
			langfuse.WithEventName("store search answer to vector store"),
			langfuse.WithEventInput(action.Question),
			langfuse.WithEventOutput(action.Answer),
			langfuse.WithEventMetadata(map[string]any{
				"tool_name":   name,
				"message":     action.Message,
				"doc_type":    searchVectorStoreDefaultType,
				"answer_type": action.Type,
			}),
		}

		logger = logger.WithFields(logrus.Fields{
			"query":       action.Question[:min(len(action.Question), 1000)],
			"answer_type": action.Type,
			"answer":      action.Answer[:min(len(action.Answer), 1000)],
		})

		docs, err := documentloaders.NewText(strings.NewReader(action.Answer)).Load(ctx)
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
			doc.Metadata["flow_id"] = s.flowID
			doc.Metadata["task_id"] = s.taskID
			doc.Metadata["subtask_id"] = s.subtaskID
			doc.Metadata["doc_type"] = searchVectorStoreDefaultType
			doc.Metadata["answer_type"] = action.Type
			doc.Metadata["question"] = action.Question
			doc.Metadata["part_size"] = len(doc.PageContent)
			doc.Metadata["total_size"] = len(action.Answer)
		}

		if _, err := s.store.AddDocuments(ctx, docs); err != nil {
			observation.Event(append(opts,
				langfuse.WithEventStatus(err.Error()),
				langfuse.WithEventLevel(langfuse.ObservationLevelError),
			)...)
			logger.WithError(err).Error("failed to store answer for question")
			return "", fmt.Errorf("failed to store answer for question: %w", err)
		}

		observation.Event(append(opts,
			langfuse.WithEventStatus("success"),
			langfuse.WithEventLevel(langfuse.ObservationLevelDebug),
			langfuse.WithEventOutput(docs),
		)...)

		if agentCtx, ok := GetAgentContext(ctx); ok {
			filtersData, err := json.Marshal(map[string]any{
				"doc_type":    searchVectorStoreDefaultType,
				"answer_type": action.Type,
				"task_id":     s.taskID,
				"subtask_id":  s.subtaskID,
			})
			if err != nil {
				logger.WithError(err).Error("failed to marshal filters")
				return "", fmt.Errorf("failed to marshal filters: %w", err)
			}
			_, _ = s.vslp.PutLog(
				ctx,
				agentCtx.ParentAgentType,
				agentCtx.CurrentAgentType,
				filtersData,
				action.Question,
				database.VecstoreActionTypeStore,
				action.Answer,
				s.taskID,
				s.subtaskID,
			)
		}

		return "answer for question stored successfully", nil

	default:
		logger.Error("unknown tool")
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *search) IsAvailable() bool {
	return s.store != nil
}
