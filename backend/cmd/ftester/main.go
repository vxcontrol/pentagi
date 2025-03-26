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
	"strings"
	"syscall"

	"pentagi/cmd/ftester/terminal"
	"pentagi/cmd/ftester/worker"
	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/provider"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	envFile := flag.String("env", ".env", "Path to environment file")
	providerTypeStr := flag.String("provider", "custom", "Provider type (openai, custom, anthropic)")
	flowID := flag.Int64("flow", 0, "Flow ID for testing functions that require it (0 means using mocks)")
	taskID := flag.Int64("task", 0, "Task ID for testing functions with default unset")
	subtaskID := flag.Int64("subtask", 0, "Subtask ID for testing functions with default unset")
	flag.Parse()

	if *taskID == 0 {
		taskID = nil
	}
	if *subtaskID == 0 {
		subtaskID = nil
	}

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
	defer lfclient.ForceFlush(context.Background())

	otelclient, err := obs.NewTelemetryClient(ctx, cfg)
	if err != nil && !errors.Is(err, obs.ErrNotConfigured) {
		log.Fatalf("Unable to create telemetry client: %v\n", err)
	}
	defer otelclient.ForceFlush(context.Background())

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
	queries := database.New(db)

	terminal.PrintHeader("Function Tester (ftester)")
	terminal.PrintInfo("Starting ftester with the following parameters:")
	terminal.PrintKeyValue("Environment file", *envFile)
	terminal.PrintKeyValue("Provider", *providerTypeStr)
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
	providerController, err := providers.NewProviderController(cfg, dockerClient)
	if err != nil {
		log.Fatalf("Failed to initialize provider controller: %v", err)
	}

	// Convert provider type string to correct provider type
	*providerTypeStr = strings.ToLower(*providerTypeStr)
	var prvType provider.ProviderType

	switch *providerTypeStr {
	case "custom":
		prvType = provider.ProviderCustom
	case "openai":
		prvType = provider.ProviderOpenAI
	case "anthropic":
		prvType = provider.ProviderAnthropic
	default:
		terminal.PrintWarning("Unknown provider type '%s', using 'custom' instead", *providerTypeStr)
		prvType = provider.ProviderCustom
	}

	// Initialize tester with appropriate proxy interfaces
	tester, err := worker.NewTester(
		queries,
		cfg,
		ctx,
		dockerClient,
		providerController,
		*flowID,
		taskID,
		subtaskID,
		prvType,
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
