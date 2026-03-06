package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pentagi/cmd/ftester/worker"
	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/provider"
	"pentagi/pkg/terminal"
	"pentagi/pkg/version"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	envFile := flag.String("env", ".env", "Path to environment file")
	providerName := flag.String("provider", "custom", "Provider name (openai, anthropic, gemini, bedrock, ollama, custom)")
	flowID := flag.Int64("flow", 0, "Flow ID for testing functions that require it (0 means using mocks)")
	userID := flag.Int64("user", 0, "User ID for testing functions that require it (1 is default admin user)")
	taskID := flag.Int64("task", 0, "Task ID for testing functions with default unset")
	subtaskID := flag.Int64("subtask", 0, "Subtask ID for testing functions with default unset")
	flag.Parse()

	if *taskID == 0 {
		taskID = nil
	}
	if *subtaskID == 0 {
		subtaskID = nil
	}

	logrus.Infof("Starting PentAGI Function Tester %s", version.GetBinaryVersion())

	err := godotenv.Load(*envFile)
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lfclient, err := obs.NewLangfuseClient(ctx, cfg)
	if err != nil && !errors.Is(err, obs.ErrNotConfigured) {
		log.Fatalf("Unable to create langfuse client: %v\n", err)
	}
	defer func() {
		if lfclient != nil {
			lfclient.ForceFlush(context.Background())
		}
	}()

	otelclient, err := obs.NewTelemetryClient(ctx, cfg)
	if err != nil && !errors.Is(err, obs.ErrNotConfigured) {
		log.Fatalf("Unable to create telemetry client: %v\n", err)
	}
	defer func() {
		if otelclient != nil {
			otelclient.ForceFlush(context.Background())
		}
	}()

	obs.InitObserver(ctx, lfclient, otelclient, []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
	})

	// Initialize database connection
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to open database: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Hour)

	queries := database.New(db)

	terminal.PrintHeader("Function Tester (ftester)")
	terminal.PrintInfo("Starting ftester with the following parameters:")
	terminal.PrintKeyValue("Environment file", *envFile)
	terminal.PrintKeyValue("Provider", *providerName)
	if *flowID != 0 {
		terminal.PrintKeyValue("Flow ID", fmt.Sprintf("%d", *flowID))
	} else {
		terminal.PrintInfo("Using mock mode (flowID=0)")
	}

	if taskID != nil {
		terminal.PrintKeyValueFormat("Task ID", "%d", *taskID)
	}
	if subtaskID != nil {
		terminal.PrintKeyValueFormat("Subtask ID", "%d", *subtaskID)
	}
	terminal.PrintThinSeparator()

	// Initialize docker client
	dockerClient, err := docker.NewDockerClient(context.Background(), queries, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Docker client: %v", err)
	}

	// Initialize provider controller
	providerController, err := providers.NewProviderController(cfg, queries, dockerClient)
	if err != nil {
		log.Fatalf("Failed to initialize provider controller: %v", err)
	}

	// Initialize tester with appropriate proxy interfaces
	tester, err := worker.NewTester(
		queries,
		cfg,
		ctx,
		dockerClient,
		providerController,
		*flowID,
		*userID,
		taskID,
		subtaskID,
		provider.ProviderName(*providerName),
	)
	if err != nil {
		log.Fatalf("Failed to initialize tester worker: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		cancel()
	}()

	// Execute the tester with the parsed arguments
	if err := tester.Execute(flag.Args()); err != nil {
		terminal.PrintError("Error executing function: %v", err)
		os.Exit(1)
	}
}
