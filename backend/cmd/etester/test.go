package main

import (
	"database/sql"
	"fmt"
	"strings"

	"pentagi/pkg/terminal"
)

const (
	testText  = "This is a test text for embedding"
	testTexts = "This is a test text for embedding\nThis is another test text for embedding"
)

// test checks connectivity to the database and tests the embedder.
func (t *Tester) test() error {
	terminal.Info("Testing connection to PostgreSQL database... ")
	err := t.conn.Ping(t.ctx)
	if err != nil {
		terminal.Error("FAILED")
		return fmt.Errorf("database connection test failed: %w", err)
	}
	terminal.Success("OK")

	terminal.Info("Testing pgvector extension... ")
	var result string
	err = t.conn.QueryRow(t.ctx, "SELECT extname FROM pg_extension WHERE extname = 'vector'").Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			terminal.Error("FAILED")
			return fmt.Errorf("pgvector extension is not installed")
		}
		terminal.Error("FAILED")
		return fmt.Errorf("failed to check pgvector extension: %w", err)
	}
	terminal.Success("OK")

	terminal.Info("Testing embedding table existence... ")
	var tableExists bool
	err = t.conn.QueryRow(t.ctx,
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)",
		t.embeddingTableName).Scan(&tableExists)
	if err != nil {
		terminal.Error("FAILED")
		return fmt.Errorf("failed to check embedding table: %w", err)
	}
	if !tableExists {
		terminal.Error("FAILED")
		return fmt.Errorf("embedding table '%s' does not exist", t.embeddingTableName)
	}
	terminal.Success("OK")

	terminal.Info("Testing embedder with single query... ")
	if !t.embedder.IsAvailable() {
		terminal.Error("FAILED")
		return fmt.Errorf("embedder is not available")
	}

	embedVector, err := t.embedder.EmbedQuery(t.ctx, testText)
	if err != nil {
		terminal.Error("FAILED")
		return fmt.Errorf("embedder test failed: %w", err)
	}
	if len(embedVector) == 0 {
		terminal.Error("FAILED")
		return fmt.Errorf("embedder returned empty vector")
	}
	terminal.Success(fmt.Sprintf("OK (%d dimensions)", len(embedVector)))

	terminal.Info("Testing embedder with multiple documents... ")
	texts := strings.Split(testTexts, "\n")
	embedVectors, err := t.embedder.EmbedDocuments(t.ctx, texts)
	if err != nil {
		terminal.Error("FAILED")
		return fmt.Errorf("embedder multi-text test failed: %w", err)
	}
	if len(embedVectors) != len(texts) {
		terminal.Error("FAILED")
		return fmt.Errorf("embedder returned wrong number of vectors: got %d, expected %d",
			len(embedVectors), len(texts))
	}
	if len(embedVectors[0]) == 0 || len(embedVectors[1]) == 0 {
		terminal.Error("FAILED")
		return fmt.Errorf("embedder returned empty vectors")
	}
	terminal.Success(fmt.Sprintf("OK (%d documents, %d dimensions each)",
		len(embedVectors), len(embedVectors[0])))

	if t.verbose {
		terminal.PrintHeader("\nVerbose output:")
		terminal.PrintKeyValue("Embedding provider", t.cfg.EmbeddingProvider)
		terminal.PrintKeyValueFormat("Vector dimensions", "%d", len(embedVector))

		// Display a sample of vector values for inspection
		vectorSample := embedVector
		if len(vectorSample) > 5 {
			vectorSample = vectorSample[:5]
		}
		terminal.PrintKeyValue("First values of test vector",
			fmt.Sprintf("%v", vectorSample))
	}

	terminal.Success("\nAll tests passed successfully!")
	return nil
}
