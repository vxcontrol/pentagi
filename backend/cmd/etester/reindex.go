package main

import (
	"fmt"
	"os"

	"pentagi/pkg/terminal"

	"github.com/jackc/pgx/v5"
	"github.com/pgvector/pgvector-go"
)

// Document represents a document in the embedding store
type Document struct {
	UUID    string
	Content string
}

// reindex recalculates embeddings for all documents in the store
func (t *Tester) reindex() error {
	terminal.Warning("This will reindex ALL documents in the embedding store.")
	terminal.Warning("This operation may take a long time depending on the number of documents.")
	response, err := terminal.GetYesNoInputContext(t.ctx, "Are you sure you want to continue?", os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to get yes/no input: %w", err)
	}

	if !response {
		terminal.Info("Operation cancelled.")
		return nil
	}

	// Get total document count
	var totalDocs int
	err = t.conn.QueryRow(t.ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", t.embeddingTableName)).Scan(&totalDocs)
	if err != nil {
		return fmt.Errorf("failed to get document count: %w", err)
	}

	if totalDocs == 0 {
		terminal.Info("No documents found in the embedding store.")
		return nil
	}

	terminal.Info(fmt.Sprintf("Found %d documents to reindex.", totalDocs))

	// Calculate batch size for processing
	batchSize := t.cfg.EmbeddingBatchSize
	if batchSize <= 0 {
		batchSize = 10 // Default batch size
	}

	rows, err := t.conn.Query(t.ctx, fmt.Sprintf("SELECT uuid, document FROM %s", t.embeddingTableName))
	if err != nil {
		return fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()

	// Collect documents
	documents := []Document{}
	for rows.Next() {
		var doc Document
		if err := rows.Scan(&doc.UUID, &doc.Content); err != nil {
			return fmt.Errorf("failed to scan document row: %w", err)
		}
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating document rows: %w", err)
	}

	totalBatches := (len(documents) + batchSize - 1) / batchSize
	processedDocs := 0

	// Process documents in batches to avoid memory issues
	for i := 0; i < totalBatches; i++ {
		start := i * batchSize
		end := min((i+1)*batchSize, len(documents))
		batchDocs := documents[start:end]

		// Extract content for embedding
		texts := make([]string, len(batchDocs))
		for j, doc := range batchDocs {
			texts[j] = doc.Content
		}

		// Generate embeddings
		terminal.Info(fmt.Sprintf("Processing batch %d/%d (%d documents)...",
			i+1, totalBatches, len(batchDocs)))

		vectors, err := t.embedder.EmbedDocuments(t.ctx, texts)
		if err != nil {
			return fmt.Errorf("failed to generate embeddings for batch %d: %w", i+1, err)
		}

		if len(vectors) != len(batchDocs) {
			return fmt.Errorf("embedder returned wrong number of vectors: got %d, expected %d",
				len(vectors), len(batchDocs))
		}

		// Update documents in database
		batch := &pgx.Batch{}
		for j, doc := range batchDocs {
			batch.Queue(
				fmt.Sprintf("UPDATE %s SET embedding = $1 WHERE uuid = $2", t.embeddingTableName),
				pgvector.NewVector(vectors[j]), doc.UUID)
		}

		results := t.conn.SendBatch(t.ctx, batch)
		if err := results.Close(); err != nil {
			return fmt.Errorf("failed to update embeddings for batch %d: %w", i+1, err)
		}

		processedDocs += len(batchDocs)
		progressPercent := float64(processedDocs) / float64(totalDocs) * 100
		terminal.Info("Progress: %.2f%% (%d/%d documents processed)", progressPercent, processedDocs, totalDocs)
	}

	terminal.Success("\nReindexing completed successfully! %d documents were updated.", processedDocs)
	return nil
}
