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
	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"tool": name,
		"args": string(args),
	})

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

		opts := []langfuse.EventStartOption{
			langfuse.WithStartEventName("retrieve memory facts from vector store"),
			langfuse.WithStartEventInput(action.Question),
			langfuse.WithStartEventMetadata(map[string]any{
				"tool_name":  name,
				"message":    action.Message,
				"limit":      memoryVectorStoreResultLimit,
				"threshold":  memoryVectorStoreThreshold,
				"doc_type":   memoryVectorStoreDefaultType,
				"task_id":    action.TaskID,
				"subtask_id": action.SubtaskID,
			}),
		}

		fields := logrus.Fields{
			"query": action.Question[:min(len(action.Question), 1000)],
		}
		if action.TaskID != nil {
			fields["task_id"] = action.TaskID.Int64()
		}
		if action.SubtaskID != nil {
			fields["subtask_id"] = action.SubtaskID.Int64()
		}

		logger = logger.WithFields(fields)

		docs, err := m.store.SimilaritySearch(
			ctx,
			action.Question,
			memoryVectorStoreResultLimit,
			vectorstores.WithScoreThreshold(memoryVectorStoreThreshold),
			vectorstores.WithFilters(filters),
		)
		if err != nil {
			observation.Event(append(opts,
				langfuse.WithStartEventStatus(err.Error()),
				langfuse.WithStartEventLevel(langfuse.ObservationLevelError),
			)...)
			logger.WithError(err).Error("failed to search for similar documents")
			return "", fmt.Errorf("failed to search for similar documents: %w", err)
		}

		// fallback to search in the flow only if task or subtask id is provided
		if isSpecificFilters, globalFilters := getGlobalFilters(filters); isSpecificFilters && len(docs) == 0 {
			docs, err = m.store.SimilaritySearch(
				ctx,
				action.Question,
				memoryVectorStoreResultLimit,
				vectorstores.WithScoreThreshold(memoryVectorStoreThreshold),
				vectorstores.WithFilters(globalFilters),
			)
			if err != nil {
				observation.Event(append(opts,
					langfuse.WithStartEventStatus(err.Error()),
					langfuse.WithStartEventLevel(langfuse.ObservationLevelError),
				)...)
				logger.WithError(err).Error("failed to search for similar documents by global filters")
				return "", fmt.Errorf("failed to search for similar documents by global filters: %w", err)
			}
		}

		if len(docs) == 0 {
			event := observation.Event(append(opts,
				langfuse.WithStartEventStatus("no memory facts found"),
				langfuse.WithStartEventLevel(langfuse.ObservationLevelWarning),
			)...)
			_, observation = event.Observation(ctx)
			observation.Score(
				langfuse.WithScoreComment("no memory facts found"),
				langfuse.WithScoreName("memory_search_result"),
				langfuse.WithScoreStringValue("not_found"),
			)
			return memoryNotFoundMessage, nil
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
				langfuse.WithScoreComment("memory facts vector store result"),
				langfuse.WithScoreName("memory_search_result"),
				langfuse.WithScoreFloatValue(float64(doc.Score)),
			)
			buffer.WriteString(fmt.Sprintf("# Match score %f\n\n", doc.Score))
			if taskID, ok := doc.Metadata["task_id"]; ok {
				buffer.WriteString(fmt.Sprintf("# Task ID %v\n\n", taskID))
			}
			if subtaskID, ok := doc.Metadata["subtask_id"]; ok {
				buffer.WriteString(fmt.Sprintf("# Subtask ID %v\n\n", subtaskID))
			}
			buffer.WriteString(fmt.Sprintf("# Tool Name '%s'\n\n", doc.Metadata["tool_name"]))
			if toolDescription, ok := doc.Metadata["tool_description"]; ok {
				buffer.WriteString(fmt.Sprintf("# Tool Description\n\n%s\n\n", toolDescription))
			}
			buffer.WriteString("# Chunk\n\n")
			buffer.WriteString(doc.PageContent)
			buffer.WriteString("\n---------------------------\n")
		}

		if agentCtx, ok := GetAgentContext(ctx); ok {
			filtersData, err := json.Marshal(filters)
			if err != nil {
				logger.WithError(err).Error("failed to marshal filters")
				return "", fmt.Errorf("failed to marshal filters: %w", err)
			}
			_, _ = m.vslp.PutLog(
				ctx,
				agentCtx.ParentAgentType,
				agentCtx.CurrentAgentType,
				filtersData,
				action.Question,
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
