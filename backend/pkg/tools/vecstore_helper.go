package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"pentagi/pkg/database"
	"pentagi/pkg/providers/embeddings"

	"github.com/google/uuid"
)

// truncateForEmbedding truncates text to maxBytes bytes for sending to the
// embedding model API. Returns text unchanged if it is already within the limit.
// This avoids token-limit errors (e.g. OpenAI: max 8192 tokens) when a heavy
// field exceeds the model's context window.
func truncateForEmbedding(text string, maxBytes int) string {
	if maxBytes <= 0 || len(text) <= maxBytes {
		return text
	}
	return text[:maxBytes]
}

// storeDocumentWithEmbeddingLimit computes the embedding vector from embeddingText
// (the truncated/prefix version of the content, ≤ maxBytes) and inserts the
// document into the vector store with the full original text as the stored content.
//
// Use this when len(fullText) > maxBytes to avoid embedding API token-limit errors
// while still persisting the complete document for retrieval.
//
// Returns the UUID string of the newly inserted document.
func storeDocumentWithEmbeddingLimit(
	ctx context.Context,
	db database.Querier,
	embedder embeddings.Embedder,
	embeddingText string,
	fullText string,
	metadata map[string]any,
) (string, error) {
	vecs, err := embedder.EmbedDocuments(ctx, []string{embeddingText})
	if err != nil {
		return "", fmt.Errorf("failed to compute embedding: %w", err)
	}
	if len(vecs) == 0 {
		return "", fmt.Errorf("embedder returned no vectors")
	}

	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	id := uuid.New()
	docID, err := db.InsertKnowledgeDocument(ctx, database.InsertKnowledgeDocumentParams{
		Uuid:      id,
		Document:  sql.NullString{String: fullText, Valid: true},
		Embedding: formatVectorFromFloat32s(vecs[0]),
		Cmetadata: json.RawMessage(metaJSON),
	})
	if err != nil {
		return "", fmt.Errorf("failed to insert document with pre-computed embedding: %w", err)
	}

	return docID, nil
}

// formatVectorFromFloat32s converts a float32 slice into a PostgreSQL vector literal '[f1,f2,...]'.
func formatVectorFromFloat32s(v []float32) string {
	strs := make([]string, len(v))
	for i, f := range v {
		strs[i] = strconv.FormatFloat(float64(f), 'f', -1, 32)
	}
	return "[" + strings.Join(strs, ",") + "]"
}
