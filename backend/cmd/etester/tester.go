package main

import (
	"context"
	"fmt"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/embeddings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Tester represents the main application structure for the etester tool
type Tester struct {
	conn                *pgxpool.Pool
	embedder            embeddings.Embedder
	embeddingTableName  string
	collectionTableName string
	verbose             bool
	command             string
	ctx                 context.Context
	cfg                 *config.Config
}

// NewTester creates a new instance of the Tester with the provided configuration
func NewTester(
	conn *pgxpool.Pool,
	embedder embeddings.Embedder,
	verbose bool,
	command string,
	ctx context.Context,
	cfg *config.Config,
) *Tester {
	return &Tester{
		conn:                conn,
		embedder:            embedder,
		embeddingTableName:  defaultEmbeddingTableName,
		collectionTableName: defaultCollectionTableName,
		verbose:             verbose,
		command:             command,
		ctx:                 ctx,
		cfg:                 cfg,
	}
}

// executeCommand executes the appropriate command based on the command string
func (t *Tester) executeCommand(args []string) error {
	switch t.command {
	case "test":
		return t.test()
	case "info":
		return t.info()
	case "flush":
		return t.flush()
	case "reindex":
		return t.reindex()
	case "search":
		return t.search(args)
	default:
		return fmt.Errorf("unknown command: %s", t.command)
	}
}

// formatSize formats a file size in bytes to a human-readable string
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
