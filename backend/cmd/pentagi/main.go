package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"pentagi/migrations"
	"pentagi/pkg/config"
	"pentagi/pkg/controller"
	"pentagi/pkg/database"
	"pentagi/pkg/docker"
	"pentagi/pkg/graph/subscriptions"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/profiling"
	"pentagi/pkg/providers"
	router "pentagi/pkg/server"
	"pentagi/pkg/version"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	// Setup graceful shutdown context with signal handling
	ctx, cancelOnSignal := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancelOnSignal()

	logrus.Infof("Starting PentAGI %s", version.GetBinaryVersion())

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Unable to load config: %v\n", err)
	}

	// Configure logrus log level based on DEBUG env variable
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("Debug logging enabled")
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	lfclient, err := obs.NewLangfuseClient(ctx, cfg)
	if err != nil && !errors.Is(err, obs.ErrNotConfigured) {
		log.Fatalf("Unable to create langfuse client: %v\n", err)
	}

	otelclient, err := obs.NewTelemetryClient(ctx, cfg)
	if err != nil && !errors.Is(err, obs.ErrNotConfigured) {
		log.Fatalf("Unable to create telemetry client: %v\n", err)
	}

	obs.InitObserver(ctx, lfclient, otelclient, []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
	})

	obs.Observer.StartProcessMetricCollect(attribute.String("component", "server"))
	obs.Observer.StartGoRuntimeMetricCollect(attribute.String("component", "server"))

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to open database: %v\n", err)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	queries := database.New(db)

	// Pass the same *sql.DB to GORM so both sqlc and GORM share one connection
	// pool. Previously each opened its own *sql.DB, consuming up to 40
	// Postgres connections. Now together they consume at most DBMaxOpenConns.
	orm, err := database.NewGorm(db, cfg.Debug)
	if err != nil {
		log.Fatalf("Unable to open database with gorm: %v\n", err)
	}

	// Create a shared pgxpool for all pgvector stores so that each executor
	// reuses pooled connections instead of opening a dedicated pgx.Connect.
	pgPoolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to parse pgxpool config: %v\n", err)
	}
	pgPoolConfig.MaxConns = int32(cfg.DBVectorMaxConns)
	pgPool, err := pgxpool.NewWithConfig(ctx, pgPoolConfig)
	if err != nil {
		log.Fatalf("Failed to create pgxpool: %v\n", err)
	}
	defer pgPool.Close()

	// Attach the live pool to the config so router and tool executors can use
	// pgvector.WithConn(cfg.PgxPool) without any interface changes.
	cfg.PgxPool = pgPool

	goose.SetBaseFS(migrations.EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Database dialect configuration failed: %v\n", err)
	}

	if err := goose.Up(db, "sql"); err != nil {
		log.Fatalf("Schema migration execution failed: %v\n", err)
	}

	log.Println("Database schema updated successfully")

	go profiling.Start()

	client, err := docker.NewDockerClient(ctx, queries, cfg)
	if err != nil {
		log.Fatalf("Docker runtime client initialization failed: %v", err)
	}

	providers, err := providers.NewProviderController(cfg, queries, client)
	if err != nil {
		log.Fatalf("LLM provider controller initialization failed: %v", err)
	}
	subscriptions := subscriptions.NewSubscriptionsController()
	controller := controller.NewFlowController(queries, cfg, client, providers, subscriptions)

	if err := controller.LoadFlows(ctx); err != nil {
		log.Fatalf("Active flows restoration failed: %v", err)
	}

	r := router.NewRouter(queries, orm, cfg, providers, controller, subscriptions, client)

	// Launch HTTP/HTTPS server in background goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		listen := net.JoinHostPort(cfg.ServerHost, strconv.Itoa(cfg.ServerPort))
		logrus.Infof("API server listening on %s", listen)

		var startErr error
		if cfg.ServerUseSSL && cfg.ServerSSLCrt != "" && cfg.ServerSSLKey != "" {
			logrus.Info("Starting server with TLS enabled")
			startErr = r.RunTLS(listen, cfg.ServerSSLCrt, cfg.ServerSSLKey)
		} else {
			logrus.Info("Starting server without TLS (HTTP only)")
			startErr = r.Run(listen)
		}

		if startErr != nil {
			serverErrChan <- fmt.Errorf("API server startup failed: %w", startErr)
		}
	}()

	// Block until shutdown signal received or server error occurs
	select {
	case <-ctx.Done():
		logrus.Warn("Shutdown signal received, cleaning up resources...")
	case err := <-serverErrChan:
		logrus.Fatalf("Server terminated unexpectedly: %v", err)
	}

	logrus.Info("Application shutdown completed successfully")
}
