package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"pentagi/pkg/database"
	"pentagi/pkg/database/knowledge"
	"pentagi/pkg/database/knowledge/limits"
	"pentagi/pkg/graph/model"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/providers/embeddings"

	"github.com/sirupsen/logrus"
	"github.com/vxcontrol/cloud/anonymizer"
	"github.com/vxcontrol/langchaingo/vectorstores/pgvector"
)

const (
	knowledgeSearchDefaultLimit = 5
)

type knowledgeSearch struct {
	userID            int64
	flowID            int64
	taskID            *int64
	subtaskID         *int64
	replacer          anonymizer.Replacer
	store             *pgvector.Store
	embedder          embeddings.Embedder
	db                database.Querier
	maxEmbeddingBytes int
	vslp              VectorStoreLogProvider
	knp               KnowledgeProvider
}

// NewKnowledgeSearchTool builds a tool that retrieves documents from the SHARED
// KNOWLEDGE BASE (the curated documents users uploaded) via semantic search. It
// is distinct from the agent memory/answer/guide/code vector stores: those hold
// what the agent produced during work, while the knowledge base holds what users
// explicitly added. The KnowledgeStore is constructed from the same db/store/embedder
// the executor already holds.
func NewKnowledgeSearchTool(
	userID int64,
	flowID int64,
	taskID, subtaskID *int64,
	replacer anonymizer.Replacer,
	store *pgvector.Store,
	embedder embeddings.Embedder,
	db database.Querier,
	maxEmbeddingBytes int,
	vslp VectorStoreLogProvider,
	knp KnowledgeProvider,
) Tool {
	return &knowledgeSearch{
		userID:            userID,
		flowID:            flowID,
		taskID:            taskID,
		subtaskID:         subtaskID,
		replacer:          replacer,
		store:             store,
		embedder:          embedder,
		db:                db,
		maxEmbeddingBytes: maxEmbeddingBytes,
		vslp:              vslp,
		knp:               knp,
	}
}

func (ks *knowledgeSearch) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(ks.flowID, ks.taskID, ks.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	// NOTE: deliberately mirrors the sibling store-backed tools (memory/guide/
	// search/code), which gate only on `store != nil`. The embedder is required at
	// search time, but its IsAvailable() helper is a thin `e.Embedder != nil` probe;
	// gating on it here would silently drop this tool from the assistant's tool list
	// even when the store is fully functional. As long as the store is set the
	// executor's embedder is guaranteed non-nil (SetEmbedder sets both together).
	if ks.store == nil || ks.embedder == nil {
		logger.Error("knowledge store or embedder is not available")
		return "", fmt.Errorf("knowledge store or embedder is not available")
	}

	switch name {
	case KnowledgeSearchToolName:
		logger.Infof("MARKER_KNOWLEDGE_SEARCH_CALLED")
		var action KnowledgeSearchAction
		if err := json.Unmarshal(args, &action); err != nil {
			logger.WithError(err).Error("failed to unmarshal knowledge search action arguments")
			return "", fmt.Errorf("failed to unmarshal %s knowledge search action arguments: %w", name, err)
		}

		limit := action.Limit
		if limit <= 0 {
			limit = knowledgeSearchDefaultLimit
		}

		// Scope to the user's CURATED knowledge base only (Manual = true): those are
		// the documents users explicitly uploaded, as opposed to the agent's own
		// generated answers/guides/code (Manual = false), which the search_answer /
		// search_guide / search_code tools already cover. Scoping here (Go-side
		// applyGoFilters) is safe regardless of whether agent docs carry the manual
		// metadata key at all.
		manual := true
		filter := &model.KnowledgeFilter{Manual: &manual}

		store := knowledge.NewKnowledgeStore(ks.db, ks.store, ks.embedder, nil, ks.maxEmbeddingBytes)

		results, err := store.SearchDocuments(ctx, action.Query, filter, limit)
		if err != nil {
			obs.LogErrorOrCancel(logger, err, "failed to search knowledge base")
			return "", fmt.Errorf("failed to search knowledge base: %w", err)
		}

		if len(results) == 0 {
			observation.Event(
				langfuse.WithEventName("search knowledge base"),
				langfuse.WithEventStatus("no documents found"),
				langfuse.WithEventInput(action.Query),
				langfuse.WithEventOutput([]any{}),
			)
			return "nothing found in the knowledge base for this query", nil
		}

		buffer := strings.Builder{}
		for i, res := range results {
			doc := res.Document
			fmt.Fprintf(&buffer, "# Document %d (relevance score: %.4f)\n\n", i+1, res.Score)
			if doc.Question != "" {
				fmt.Fprintf(&buffer, "## Question\n%s\n\n", doc.Question)
			}
			buffer.WriteString("## Content\n\n")
			buffer.WriteString(doc.Content)
			buffer.WriteString("\n\n")
		}

		anonymizedQuery := limits.TruncateToLimit(ks.replacer.ReplaceString(action.Query), limits.MaxQuestionLen)
		if agentCtx, ok := GetAgentContext(ctx); ok {
			data, _ := json.Marshal(map[string]any{"limit": limit})
			_, _ = ks.vslp.PutLog(
				ctx,
				agentCtx.ParentAgentType,
				agentCtx.CurrentAgentType,
				data,
				anonymizedQuery,
				database.VecstoreActionTypeRetrieve,
				buffer.String(),
				ks.taskID,
				ks.subtaskID,
			)
		}

		return buffer.String(), nil

	default:
		logger.Error("unknown tool")
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (ks *knowledgeSearch) IsAvailable() bool {
	return ks.store != nil && ks.embedder != nil
}
