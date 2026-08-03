package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/providers/embeddings"
	"pentagi/pkg/terminal"
	"pentagi/pkg/version"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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

	// Create this tenant's schema and repoint DATABASE_URL at it before any
	// consumer reads the DSN. No-op when TENANT_ID is empty.
	if err := database.EnsureTenantSchema(ctx, cfg); err != nil {
		log.Fatalf("Tenant schema initialization failed: %v", err)
	}

	// Verify search_path on a short-lived database/sql connection before the
	// long-lived pgxpool is opened — same guard the main server uses.
	verifyDB, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to open database for schema verification: %v", err)
	}
	verifyDB.SetMaxOpenConns(1)
	if err := database.VerifySearchPath(ctx, verifyDB, cfg); err != nil {
		verifyDB.Close()
		log.Fatalf("Tenant schema verification failed: %v", err)
	}
	verifyDB.Close()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to parse database URL: %v", err)
	}

	poolConfig.MaxConns = min(int32(cfg.DBVectorMaxConns), 10)
	poolConfig.MinConns = min(int32(cfg.DBMaxIdleConns), 2, poolConfig.MaxConns)
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
