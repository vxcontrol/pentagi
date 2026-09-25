package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	// The runtime image is alpine without the tzdata package, so IANA names the
	// dashboard sends would not resolve unless the binary carries the database.
	_ "time/tzdata"

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
	"pentagi/pkg/server/update"
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

	// Before anything is logged or started. `-info` describes this server and
	// exits, and its output is a JSON document on stdout — a greeting printed
	// first would be the first thing a reader has to strip back out.
	if wantsInfo(os.Args[1:]) {
		os.Exit(runInfo(ctx))
	}

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

	_ = obs.Observer.StartProcessMetricCollect(attribute.String("component", "server"))
	_ = obs.Observer.StartGoRuntimeMetricCollect(attribute.String("component", "server"))

	// Create this tenant's schema and repoint DATABASE_URL at it before any
	// consumer reads the DSN. No-op when TENANT_ID is empty.
	if err := database.EnsureTenantSchema(ctx, cfg); err != nil {
		logrus.WithError(err).Fatal("Tenant schema initialization failed")
	}

	boundedURL, err := database.WithStatementTimeout(cfg.DatabaseURL, database.StatementTimeout)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to bound database statements")
	}

	db, err := sql.Open("postgres", boundedURL)
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
	pgPoolConfig, err := pgxpool.ParseConfig(boundedURL)
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
	// Migrations get a connection of their own, without the statement ceiling:
	// one of them may legitimately run longer than any request would.
	migrator, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to open database for migrations")
	}

	migrationErr := database.RunMigrations(ctx, migrator, cfg, func(db *sql.DB) error {
		return goose.Up(db, "sql")
	})

	if err := migrator.Close(); err != nil {
		logrus.WithError(err).Warn("failed to close the migration connection")
	}

	if err := migrationErr; err != nil {
		// Fatal: continuing on a half-migrated schema and then serving traffic is
		// strictly worse than refusing to start.
		logrus.WithError(err).Fatal("Schema migration execution failed")
	}

	logrus.Info("Database schema updated successfully")

	// Keep the update service informed about this installation, periodically.
	// After the migrations, deliberately: on a first boot the tables it reads do
	// not exist until this point, and a summary of a schema that is not there yet
	// describes an installation with nothing in it — which is also what an
	// installation nobody uses looks like.
	//
	// A goroutine that owns nothing: it reads, it sends, and every failure along
	// the way is a debug line. A server must not fail to start because something
	// it only talks to is unreachable. What it learns is handed to the router, so
	// the UI can show whether this build is current; nil means it never runs, and
	// the UI says so.
	var updates *update.Service
	if built, err := update.New(cfg, queries); err != nil {
		logrus.WithError(err).Debug("update service integration is not configured")
	} else {
		updates = built
		go updates.Run(ctx)
	}

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

	r := router.NewRouter(queries, orm, cfg, providers, controller, subscriptions, client, updates)

	// Launch HTTP/HTTPS server in background goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		listen := net.JoinHostPort(cfg.ServerHost, strconv.Itoa(cfg.ServerPort))
		logrus.Infof("API server listening on %s", listen)

		srv := &http.Server{
			Addr:              listen,
			Handler:           r.Handler(),
			ReadHeaderTimeout: router.ReadHeaderTimeout,
			IdleTimeout:       router.IdleTimeout,
		}

		var startErr error
		if cfg.ServerUseSSL && cfg.ServerSSLCrt != "" && cfg.ServerSSLKey != "" {
			logrus.Info("Starting server with TLS enabled")
			startErr = srv.ListenAndServeTLS(cfg.ServerSSLCrt, cfg.ServerSSLKey)
		} else {
			logrus.Info("Starting server without TLS (HTTP only)")
			startErr = srv.ListenAndServe()
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
