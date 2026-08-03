package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
		// Fatal rather than Error: cfg is nil on failure, so continuing would
		// panic anyway, and an invalid TENANT_ID must never boot into a namespace
		// it did not intend to own.
		logrus.WithError(err).Fatal("Unable to load config")
	}

	// Configure logrus log level based on DEBUG env variable
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("Debug logging enabled")
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Record the identity of this instance for the record: which tenant namespace
	// it owns and which data directory it writes to. Both are operator-supplied
	// and must differ between instances that share a host; having them in the
	// startup log makes an accidental overlap obvious after the fact.
	logrus.WithFields(logrus.Fields{
		"tenant_id":       cfg.TenantID,
		"data_dir":        cfg.DataDir,
		"schema":          cfg.SchemaName(),
		"installation_id": cfg.InstallationID,
	}).Info("Instance identity")

	// Telemetry is optional — degrade to a no-op observer on init failure so an
	// unreachable collector can't take the app down.
	lfclient, err := obs.NewLangfuseClient(ctx, cfg)
	if err != nil && !errors.Is(err, obs.ErrNotConfigured) {
		logrus.WithError(err).Warn("langfuse telemetry disabled: client init failed")
		lfclient = nil
	}

	otelclient, err := obs.NewTelemetryClient(ctx, cfg)
	if err != nil && !errors.Is(err, obs.ErrNotConfigured) {
		logrus.WithError(err).Warn("opentelemetry disabled: client init failed")
		otelclient = nil
	}

	obs.InitObserver(ctx, lfclient, otelclient, []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
	})

	obs.Observer.StartProcessMetricCollect(attribute.String("component", "server"))
	obs.Observer.StartGoRuntimeMetricCollect(attribute.String("component", "server"))

	// Create this tenant's schema and repoint DATABASE_URL at it before any
	// consumer reads the DSN. No-op when TENANT_ID is empty.
	if err := database.EnsureTenantSchema(ctx, cfg); err != nil {
		logrus.WithError(err).Fatal("Tenant schema initialization failed")
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to open database")
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	if err := database.VerifySearchPath(ctx, db, cfg); err != nil {
		logrus.WithError(err).Fatal("Tenant schema verification failed")
	}

	queries := database.New(db)

	// Pass the same *sql.DB to GORM so both sqlc and GORM share one connection
	// pool. Previously each opened its own *sql.DB, consuming up to 40
	// Postgres connections. Now together they consume at most DBMaxOpenConns.
	orm, err := database.NewGorm(db, cfg.Debug)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to open database with gorm")
	}

	// Create a shared pgxpool for all pgvector stores so that each executor
	// reuses pooled connections instead of opening a dedicated pgx.Connect.
	pgPoolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to parse pgxpool config")
	}
	pgPoolConfig.MaxConns = int32(cfg.DBVectorMaxConns)
	pgPool, err := pgxpool.NewWithConfig(ctx, pgPoolConfig)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create pgxpool")
	}
	defer pgPool.Close()

	// Attach the live pool to the config so router and tool executors can use
	// pgvector.WithConn(cfg.PgxPool) without any interface changes.
	cfg.PgxPool = pgPool

	goose.SetBaseFS(migrations.EmbedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		logrus.WithError(err).Fatal("Database dialect configuration failed")
	}

	// goose's own queries are unqualified, so without this a fresh tenant
	// schema silently inherits public's version table via search_path and
	// skips its migrations. See pkg/database/tenant.go for the
	// schema/search_path setup.
	goose.SetTableName(cfg.SchemaName() + ".goose_db_version")

	// Hold an advisory lock so simultaneous boots cannot execute the same
	// migration set concurrently; the initial migration uses bare CREATE TABLE,
	// so the loser would otherwise abort on "relation already exists".
	if err := database.RunMigrations(ctx, db, cfg, func(db *sql.DB) error {
		return goose.Up(db, "sql")
	}); err != nil {
		// Fatal: continuing on a half-migrated schema and then serving traffic is
		// strictly worse than refusing to start.
		logrus.WithError(err).Fatal("Schema migration execution failed")
	}

	logrus.Info("Database schema updated successfully")

	if cfg.PprofAddr != "" {
		go profiling.Start(cfg.PprofAddr)
	}

	client, err := docker.NewDockerClient(ctx, queries, cfg)
	if err != nil {
		logrus.WithError(err).Fatal("Docker runtime client initialization failed")
	}

	providers, err := providers.NewProviderController(cfg, queries, client)
	if err != nil {
		logrus.WithError(err).Fatal("LLM provider controller initialization failed")
	}
	subscriptions := subscriptions.NewSubscriptionsController()
	controller := controller.NewFlowController(queries, cfg, client, providers, subscriptions)

	if err := controller.LoadFlows(ctx); err != nil {
		logrus.WithError(err).Fatal("Active flows restoration failed")
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
		// ctx is already cancelled here — drain telemetry on a fresh deadline so the
		// final batch still reaches the collector before the process exits.
		drainCtx, cancelDrain := context.WithTimeout(context.Background(), 5*time.Second)
		if err := obs.Observer.Drain(drainCtx); err != nil {
			logrus.WithError(err).Warn("Telemetry drain incomplete")
		}
		cancelDrain()
	case err := <-serverErrChan:
		logrus.Fatalf("Server terminated unexpectedly: %v", err)
	}

	logrus.Info("Application shutdown completed successfully")
}
