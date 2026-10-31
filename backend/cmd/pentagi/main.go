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

// HTTP shutdown and telemetry drain run on fresh contexts: the signal context
// is already cancelled by the time either starts.
const (
	httpShutdownTimeout   = 15 * time.Second
	telemetryDrainTimeout = 5 * time.Second
)

func main() {
	// os.Exit skips deferred calls, so every deferred cleanup lives in
	// runApplication and the exit status is only decided once it has returned.
	os.Exit(runApplication())
}

func runApplication() int {
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
		return runInfo(ctx)
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

	// One pool shared by sqlc and GORM, closed here exactly once. Deferred, so
	// it runs only after the HTTP server has stopped (and after the pgxpool
	// defer below, which was registered later).
	defer func() {
		if err := db.Close(); err != nil {
			logrus.WithError(err).Warn("failed to close the database connection pool")
		}
	}()

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

	listen := net.JoinHostPort(cfg.ServerHost, strconv.Itoa(cfg.ServerPort))
	srv := &http.Server{
		Addr:              listen,
		Handler:           r.Handler(),
		ReadHeaderTimeout: router.ReadHeaderTimeout,
		IdleTimeout:       router.IdleTimeout,
	}

	serve := func() error {
		logrus.Infof("API server listening on %s", listen)

		if cfg.ServerUseSSL && cfg.ServerSSLCrt != "" && cfg.ServerSSLKey != "" {
			logrus.Info("Starting server with TLS enabled")
			return srv.ListenAndServeTLS(cfg.ServerSSLCrt, cfg.ServerSSLKey)
		}

		logrus.Info("Starting server without TLS (HTTP only)")
		return srv.ListenAndServe()
	}

	// Blocks until a shutdown signal or a server failure, then stops the HTTP
	// server and drains telemetry. The exit code is returned rather than raised
	// with Fatal so the deferred pool cleanup above still runs.
	exitCode := serveUntilShutdown(ctx, srv, serve, httpShutdownTimeout, func(ctx context.Context) error {
		return obs.Observer.Drain(ctx)
	})

	// On a server failure ctx is still live; cancel it so work bound to it (the
	// update service, flows) stops before the deferred calls close the pools.
	// Idempotent after a signal.
	cancelOnSignal()

	if exitCode == 0 {
		logrus.Info("Application shutdown completed successfully")
	}

	return exitCode
}

// httpServer is the part of *http.Server the shutdown logic needs, so tests can
// drive the deadline and failure paths without waiting on real timers.
type httpServer interface {
	Shutdown(ctx context.Context) error
	Close() error
}

// serveUntilShutdown runs serve until ctx is cancelled or serve fails, stops the
// HTTP server, and only then drains telemetry, so requests that finish during
// graceful shutdown can still emit it. It returns the process exit code and
// never exits itself, which keeps the caller's deferred cleanup reachable.
func serveUntilShutdown(
	ctx context.Context,
	srv httpServer,
	serve func() error,
	shutdownTimeout time.Duration,
	drain func(context.Context) error,
) int {
	err := superviseHTTPServer(ctx, srv, serve, shutdownTimeout)

	drainTelemetry(drain)

	if err != nil {
		logrus.WithError(err).Error("HTTP server terminated with an error")
		return 1
	}

	return 0
}

// superviseHTTPServer runs serve in the background and returns once the server
// has been stopped. It returns nil only for a shutdown it requested itself and
// that completed gracefully.
func superviseHTTPServer(
	ctx context.Context,
	srv httpServer,
	serve func() error,
	shutdownTimeout time.Duration,
) error {
	// Buffered so the goroutine can always deliver its result and exit, even
	// when nobody is left to receive it.
	serveErr := make(chan error, 1)
	go func() { serveErr <- serve() }()

	select {
	case <-ctx.Done():
		logrus.Warn("Shutdown signal received, cleaning up resources...")
		shutdownErr := shutdownHTTPServer(srv, shutdownTimeout)

		// Shutdown makes serve return http.ErrServerClosed. Wait for it so a real
		// error that raced the signal is reported instead of lost.
		var failure error
		if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
			failure = fmt.Errorf("API server failed: %w", err)
		}

		return errors.Join(failure, shutdownErr)

	case err := <-serveErr:
		// Nobody asked the server to stop, so even http.ErrServerClosed or a nil
		// return means it died on its own.
		if err == nil {
			err = errors.New("serve returned without an error")
		}
		failure := fmt.Errorf("API server failed: %w", err)

		// Connections accepted before the failure are still being served; give
		// them the same bounded shutdown before their resources are released.
		return errors.Join(failure, shutdownHTTPServer(srv, shutdownTimeout))
	}
}

// shutdownHTTPServer gracefully stops srv within timeout and force-closes what
// remains if that fails. Shutdown does not wait for hijacked connections
// (WebSockets), nor does it stop background services or running flows.
func shutdownHTTPServer(srv httpServer, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := srv.Shutdown(ctx)
	if err == nil {
		return nil
	}

	logrus.WithError(err).Warn("Graceful HTTP shutdown failed, closing remaining connections")
	shutdownErr := fmt.Errorf("graceful HTTP shutdown: %w", err)

	if closeErr := srv.Close(); closeErr != nil {
		logrus.WithError(closeErr).Error("Forced HTTP server close failed")
		return errors.Join(shutdownErr, fmt.Errorf("force close HTTP server: %w", closeErr))
	}

	return shutdownErr
}

// drainTelemetry flushes buffered telemetry on its own deadline. A failure is
// only logged: losing a final batch must not change the exit status.
func drainTelemetry(drain func(context.Context) error) {
	ctx, cancel := context.WithTimeout(context.Background(), telemetryDrainTimeout)
	defer cancel()

	if err := drain(ctx); err != nil {
		logrus.WithError(err).Warn("Telemetry drain incomplete")
	}
}
