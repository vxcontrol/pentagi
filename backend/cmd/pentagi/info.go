package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/providers"
	"pentagi/pkg/providers/embeddings"

	"github.com/sirupsen/logrus"
)

// infoFlag prints a description of this server and exits.
const infoFlag = "-info"

// infoDeadline bounds the whole description. Chosen so that a caller which gave
// this process ten seconds still gets a document rather than a killed process,
// and so that a database answering slowly costs a `skipped` entry instead of a
// wait nobody budgeted for.
const infoDeadline = 8 * time.Second

// wantsInfo reports whether the process was asked to describe itself.
//
// Argument parsing rather than the flag package: the server takes no flags at
// all, and introducing a FlagSet would make every unrecognised argument a fatal
// error for a program that has always ignored them.
func wantsInfo(args []string) bool {
	for _, arg := range args {
		if arg == infoFlag || arg == "--info" {
			return true
		}
	}
	return false
}

// runInfo writes the current state of this server to stdout as JSON.
//
// stdout carries the document and nothing else. Framework logging is silenced
// rather than merely redirected: a library that decides to greet the operator
// would otherwise land in the middle of the document, and whoever reads this
// output is entitled to hand it straight to a JSON parser. Diagnostics go to
// stderr, where they belong and where they cannot corrupt the result.
//
// A document is produced even when the database cannot be reached. The parts
// that could not be gathered are named in `skipped`, which is a description of a
// server that is having trouble — still more useful than an error, and honest in
// a way that a page of zeroes would not be.
func runInfo(ctx context.Context) int {
	logrus.SetOutput(io.Discard)

	// A ceiling on the whole description, not just on each part of it. Every
	// aggregate has its own deadline, but a database that accepts connections and
	// then stops answering would let them run one after another — thirteen times
	// three seconds is not a flag anybody wants to wait for, and whoever runs this
	// programmatically gives up long before that and reports nothing at all.
	ctx, cancel := context.WithTimeout(ctx, infoDeadline)
	defer cancel()

	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot load configuration: %v\n", err)
		return 1
	}

	// The tenant's schema has to be on the search_path or every count reads the
	// wrong rows — or none. This is the read-only half of the tenant bootstrap:
	// it rewrites the DSN and never touches the catalog, which is what a flag
	// that only describes things is allowed to do.
	if err := database.RewriteDatabaseURLForTenant(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "cannot resolve tenant schema: %v\n", err)
		return 1
	}

	options := database.SummaryOptions{
		Features: database.InstanceFeatures{
			Debug:              cfg.Debug,
			AskUser:            cfg.AskUser,
			DockerInside:       cfg.DockerInside,
			AssistantUseAgents: cfg.AssistantUseAgents,
		},
		OAuth:             configuredOAuthProviders(cfg),
		DefaultProviders:  providers.EnabledDefaultProviderTypes(cfg),
		EmbeddingProvider: embeddings.ResolvedProviderType(cfg),
	}

	// sql.Open does not connect, so an unreachable database costs nothing here
	// and shows up as skipped aggregates below — which is the intended shape.
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open database: %v\n", err)
		return 1
	}
	defer db.Close()

	// One connection and a short life: this process exists for a fraction of a
	// second and must not take a slot from the server that is actually running.
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(time.Minute)

	summary := database.New(db).GetInstanceSummary(ctx, options)

	encoded, err := summary.MarshalIndent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot render the summary: %v\n", err)
		return 1
	}

	fmt.Println(string(encoded))
	return 0
}

// configuredOAuthProviders names the sign-in providers this server offers.
//
// The names of the providers, never of a user. A provider counts as configured
// under exactly the conditions the router uses to register it, so this list
// cannot claim a sign-in route that does not exist.
func configuredOAuthProviders(cfg *config.Config) []string {
	providers := []string{}
	if cfg.PublicURL == "" {
		return providers
	}
	if cfg.OAuthGoogleClientID != "" && cfg.OAuthGoogleClientSecret != "" {
		providers = append(providers, "google")
	}
	if cfg.OAuthGithubClientID != "" && cfg.OAuthGithubClientSecret != "" {
		providers = append(providers, "github")
	}
	return providers
}
