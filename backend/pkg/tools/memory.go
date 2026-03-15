package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strconv"
	"strings"

	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"

	"github.com/sirupsen/logrus"
	"github.com/vxcontrol/langchaingo/schema"
	"github.com/vxcontrol/langchaingo/vectorstores"
	"github.com/vxcontrol/langchaingo/vectorstores/pgvector"
)

const (
	memoryVectorStoreThreshold   = 0.2
	memoryVectorStoreResultLimit = 3
	memoryVectorStoreDefaultType = "memory"
	memoryNotFoundMessage        = "nothing found in memory store by this question"
)

type memory struct {
	flowID int64
	store  *pgvector.Store
	vslp   VectorStoreLogProvider
}

func NewMemoryTool(flowID int64, store *pgvector.Store, vslp VectorStoreLogProvider) Tool {
	return &memory{
		flowID: flowID,
		store:  store,
		vslp:   vslp,
	}
}

func (m *memory) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if !m.IsAvailable() {
		return "", fmt.Errorf("memory is not available")
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(m.flowID, nil, nil, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	if m.store == nil {
		logger.Error("pgvector store is not initialized")
		return "", fmt.Errorf("pgvector store is not initialized")
	}

	switch name {
	case SearchInMemoryToolName:
		var action SearchInMemoryAction
		if err := json.Unmarshal(args, &action); err != nil {
			logger.WithError(err).Error("failed to unmarshal search in memory action arguments")
			return "", fmt.Errorf("failed to unmarshal %s search in memory action arguments: %w", name, err)
		}

		filters := map[string]any{
			"flow_id":  strconv.FormatInt(m.flowID, 10),
			"doc_type": memoryVectorStoreDefaultType,
		}
		if action.TaskID != nil && *action.TaskID != 0 {
			filters["task_id"] = action.TaskID.String()
		}
		if action.SubtaskID != nil && *action.SubtaskID != 0 {
			filters["subtask_id"] = action.SubtaskID.String()
		}

		isSpecificFilters, globalFilters := getGlobalFilters(filters)
		metadata := langfuse.Metadata{
			"tool_name":        name,
			"message":          action.Message,
			"limit":            memoryVectorStoreResultLimit,
			"threshold":        memoryVectorStoreThreshold,
			"doc_type":         memoryVectorStoreDefaultType,
			"task_id":          action.TaskID,
			"subtask_id":       action.SubtaskID,
			"specific_filters": isSpecificFilters,
			"queries_count":    len(action.Questions),
		}

		retriever := observation.Retriever(
			langfuse.WithRetrieverName("retrieve memory facts from vector store"),
			langfuse.WithRetrieverInput(map[string]any{
				"queries":     action.Questions,
				"threshold":   memoryVectorStoreThreshold,
				"max_results": memoryVectorStoreResultLimit,
				"filters":     filters,
			}),
			langfuse.WithRetrieverMetadata(metadata),
		)
		ctx, observation = retriever.Observation(ctx)

		fields := logrus.Fields{
			"queries_count": len(action.Questions),
		}
		if action.TaskID != nil {
			fields["task_id"] = action.TaskID.Int64()
		}
		if action.SubtaskID != nil {
			fields["subtask_id"] = action.SubtaskID.Int64()
		}

		logger = logger.WithFields(fields)

		// Execute multiple queries and collect all documents
		var allDocs []schema.Document
		for i, query := range action.Questions {
			queryLogger := logger.WithFields(logrus.Fields{
				"query_index": i + 1,
				"query":       query[:min(len(query), 1000)],
			})

			docs, err := m.store.SimilaritySearch(
				ctx,
				query,
				memoryVectorStoreResultLimit,
				vectorstores.WithScoreThreshold(memoryVectorStoreThreshold),
				vectorstores.WithFilters(filters),
			)
			if err != nil {
				queryLogger.WithError(err).Error("failed to search for similar documents")
				continue // Continue with other queries even if one fails
			}

			// Fallback to global filters if specific filters yielded no results
			if isSpecificFilters && len(docs) == 0 {
				docs, err = m.store.SimilaritySearch(
					ctx,
					query,
					memoryVectorStoreResultLimit,
					vectorstores.WithScoreThreshold(memoryVectorStoreThreshold),
					vectorstores.WithFilters(globalFilters),
				)
				if err != nil {
					queryLogger.WithError(err).Error("failed to search with global filters")
					continue
				}
				if len(docs) > 0 {
					observation.Event(
						langfuse.WithEventName("memory search fallback to global filters"),
						langfuse.WithEventInput(map[string]any{
							"query":       query,
							"query_index": i + 1,
							"threshold":   memoryVectorStoreThreshold,
							"max_results": memoryVectorStoreResultLimit,
							"filters":     globalFilters,
						}),
						langfuse.WithEventOutput(docs),
						langfuse.WithEventLevel(langfuse.ObservationLevelDebug),
					)
				}
			}

			queryLogger.WithField("docs_found", len(docs)).Debug("query executed")
			allDocs = append(allDocs, docs...)
		}

		logger.WithFields(logrus.Fields{
			"total_docs_before_dedup": len(allDocs),
		}).Debug("all queries completed")

		// Merge, deduplicate, sort by score, and limit results
		docs := MergeAndDeduplicateDocs(allDocs, memoryVectorStoreResultLimit)

		logger.WithFields(logrus.Fields{
			"docs_after_dedup": len(docs),
		}).Debug("documents deduplicated and sorted")

		if len(docs) == 0 {
			retriever.End(
				langfuse.WithRetrieverStatus("no memory facts found"),
				langfuse.WithRetrieverLevel(langfuse.ObservationLevelWarning),
				langfuse.WithRetrieverOutput([]any{}),
			)
			observation.Score(
				langfuse.WithScoreComment("no memory facts found"),
				langfuse.WithScoreName("memory_search_result"),
				langfuse.WithScoreStringValue("not_found"),
			)
			return memoryNotFoundMessage, nil
		}

		retriever.End(
			langfuse.WithRetrieverStatus("success"),
			langfuse.WithRetrieverLevel(langfuse.ObservationLevelDebug),
			langfuse.WithRetrieverOutput(docs),
		)

		buffer := strings.Builder{}
		for i, doc := range docs {
			observation.Score(
				langfuse.WithScoreComment("memory facts vector store result"),
				langfuse.WithScoreName("memory_search_result"),
				langfuse.WithScoreFloatValue(float64(doc.Score)),
			)
			buffer.WriteString(fmt.Sprintf("# Retrieved Memory Fact %d Match score: %f\n\n", i+1, doc.Score))
			if taskID, ok := doc.Metadata["task_id"]; ok {
				buffer.WriteString(fmt.Sprintf("## Task ID %v\n\n", taskID))
			}
			if subtaskID, ok := doc.Metadata["subtask_id"]; ok {
				buffer.WriteString(fmt.Sprintf("## Subtask ID %v\n\n", subtaskID))
			}
			buffer.WriteString(fmt.Sprintf("## Tool Name '%s'\n\n", doc.Metadata["tool_name"]))
			if toolDescription, ok := doc.Metadata["tool_description"]; ok {
				buffer.WriteString(fmt.Sprintf("## Tool Description\n\n%s\n\n", toolDescription))
			}
			buffer.WriteString("## Content\n\n")
			buffer.WriteString(doc.PageContent)
			buffer.WriteString("\n---------------------------\n")
		}

		if agentCtx, ok := GetAgentContext(ctx); ok {
			filtersData, err := json.Marshal(filters)
			if err != nil {
				logger.WithError(err).Error("failed to marshal filters")
				return "", fmt.Errorf("failed to marshal filters: %w", err)
			}
			// Join all queries for logging
			queriesText := strings.Join(action.Questions, "\n--------------------------------\n")
			_, _ = m.vslp.PutLog(
				ctx,
				agentCtx.ParentAgentType,
				agentCtx.CurrentAgentType,
				filtersData,
				queriesText,
				database.VecstoreActionTypeRetrieve,
				buffer.String(),
				action.TaskID.PtrInt64(),
				action.SubtaskID.PtrInt64(),
			)
		}

		return buffer.String(), nil

	default:
		logger.Error("unknown tool")
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (m *memory) IsAvailable() bool {
	return m.store != nil
}

func getGlobalFilters(filters map[string]any) (bool, map[string]any) {
	globalFilters := maps.Clone(filters)
	delete(globalFilters, "task_id")
	delete(globalFilters, "subtask_id")
	return len(globalFilters) != len(filters), globalFilters
}
