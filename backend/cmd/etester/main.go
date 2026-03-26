package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/embeddings"
	"pentagi/pkg/terminal"
	"pentagi/pkg/version"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

const (
	defaultEmbeddingTableName  = "langchain_pg_embedding"
	defaultCollectionTableName = "langchain_pg_collection"
)

func main() {
	// Define flags (but don't include command as a flag)
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	envFile := flag.String("env", ".env", "Path to environment file")
	help := flag.Bool("help", false, "Show help information")
	flag.Parse()

	logrus.Infof("Starting PentAGI Embedding Tester %s", version.GetBinaryVersion())

	// Extract command from first non-flag argument
	args := flag.Args()
	var command string
	if len(args) > 0 {
		command = args[0]
		args = args[1:] // Remove command from args
	} else {
		command = "test" // Default command
	}

	if *help {
		showHelp()
		return
	}

	// Load environment from .env file
	err := godotenv.Load(*envFile)
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database connection pool
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to parse database URL: %v", err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	connPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer connPool.Close()

	embedder, err := embeddings.New(cfg)
	if err != nil {
		log.Fatalf("Unable to create embedder: %v", err)
	}

	// Initialize tester with the parsed command
	tester := NewTester(
		connPool,
		embedder,
		*verbose,
		command,
		ctx,
		cfg,
	)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		terminal.Info("Shutting down gracefully...")
		cancel()
	}()

	// Execute the command with remaining arguments
	if err := tester.executeCommand(args); err != nil {
		terminal.Error("Error executing command: %v", err)
		os.Exit(1)
	}
}

func showHelp() {
	terminal.PrintHeader("Embedding Tester (etester) - A tool for testing and managing embeddings")
	terminal.Info("\nUsage:")
	terminal.Info("  ./etester [flags] [command] [args]")
	terminal.Info("\nFlags:")
	terminal.Info("  -env string       Path to environment file (default \".env\")")
	terminal.Info("  -verbose          Enable verbose output")
	terminal.Info("  -help             Show this help message")
	terminal.Info("\nCommands:")
	terminal.PrintKeyValue("  test    ", "Test embedding provider and pgvector connection")
	terminal.PrintKeyValue("  info    ", "Display statistics about the embedding database")
	terminal.PrintKeyValue("  flush   ", "Delete all documents from the embedding database")
	terminal.PrintKeyValue("  reindex ", "Recalculate embeddings for all documents")
	terminal.PrintKeyValue("  search  ", "Search for documents in the embedding database")
	terminal.Info("\nExamples:")
	terminal.Info("  ./etester test -verbose         Test with verbose output")
	terminal.Info("  ./etester info                  Show database statistics")
	terminal.Info("  ./etester flush                 Delete all documents")
	terminal.Info("  ./etester reindex               Reindex all documents")
	terminal.Info("  ./etester search -query \"How to install PostgreSQL\"  Search for documents")
	terminal.Info("")
}
