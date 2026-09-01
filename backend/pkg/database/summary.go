package database

import (
	"context"
	"encoding/json"
	"runtime"
	"sort"
	"time"

	"pentagi/pkg/version"
)

// InstanceSummarySchema is the version of the document GetInstanceSummary
// produces. It travels inside the document so a reader that meets a newer one
// can tell that it is newer rather than malformed.
//
// Bump it when a field changes meaning or disappears. Adding a field does not
// need a bump: a reader that does not know a key ignores it, which is the whole
// reason the document is JSON rather than a fixed record.
const InstanceSummarySchema = 1

// summaryQueryTimeout bounds each aggregate on its own.
//
// The point of this document is to describe a server cheaply. A count over a
// table that has grown for a year is a sequential scan in PostgreSQL, and an
// installation big enough for that to be slow is exactly the one that must not
// be disturbed to produce a number. So every aggregate gets its own deadline and
// a missed deadline costs one field, named in Skipped, rather than the document.
const summaryQueryTimeout = 10 * time.Second

// InstanceSummary describes what this server currently holds and how it is set
// up. Quantities and closed vocabularies only — no names, no addresses, no
// contents of anything a user wrote.
type InstanceSummary struct {
	Schema  int    `json:"schema"`
	Version string `json:"version"`
	Develop bool   `json:"develop"`

	// OS, Arch and BinaryHash describe the process that produced this document, and
	// they are in the document because nothing that carries it can produce them.
	//
	// A caller that runs this binary inside a container to obtain the description
	// knows the platform of the HOST — on a macOS host that is `darwin` while this
	// process is linux — and the image digest it can see names a filesystem rather
	// than the file that is executing. Both are inferences about a process drawn from
	// its surroundings, and both are wrong in ways that are hard to notice. The
	// process states them instead: the platform it was compiled for, and the sha256
	// of the executable it was started from.
	//
	// BinaryHash is what tells two builds carrying the same version string apart,
	// which outside a release is every build. Empty when it could not be read.
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	BinaryHash string `json:"binary_hash"`

	Features InstanceFeatures `json:"features"`

	Providers []InstanceProvider `json:"providers"`
	OAuth     []string           `json:"oauth"`

	// EmbeddingProvider is the provider type actually used for embeddings, as
	// resolved by the caller — empty when embeddings are disabled ("none").
	EmbeddingProvider string `json:"embedding_provider"`

	Counts InstanceCounts  `json:"counts"`
	Usage  []InstanceUsage `json:"usage"`

	// Skipped names the aggregates that could not be produced — a timeout, a
	// table that is not there yet on a partially migrated instance. It exists
	// because a zero and a failure are indistinguishable otherwise, and the two
	// mean opposite things about an installation.
	Skipped []string `json:"skipped,omitempty"`
}

// InstanceFeatures are the switches this server runs with. They come from the
// configuration rather than from the database, so the caller supplies them.
type InstanceFeatures struct {
	Debug              bool `json:"debug"`
	AskUser            bool `json:"ask_user"`
	DockerInside       bool `json:"docker_inside"`
	AssistantUseAgents bool `json:"assistant_use_agents"`
}

// InstanceProvider is one configured provider type and how many of it there are.
// The type only: the names and credentials of individual providers are not part
// of a description of the server and must never be put in one.
type InstanceProvider struct {
	Type  string `json:"type"`
	Count uint32 `json:"count"`
}

// InstanceCounts are the sizes of the things this server holds.
type InstanceCounts struct {
	Users         uint32 `json:"users"`
	Flows         uint32 `json:"flows"`
	FlowsActive   uint32 `json:"flows_active"`
	Tasks         uint32 `json:"tasks"`
	Subtasks      uint32 `json:"subtasks"`
	Containers    uint32 `json:"containers"`
	Assistants    uint32 `json:"assistants"`
	FlowTemplates uint32 `json:"flow_templates"`
	Prompts       uint32 `json:"prompts"`
	APITokens     uint32 `json:"api_tokens"`
	Toolcalls     uint64 `json:"toolcalls"`
}

// InstanceUsage is what has been consumed through one provider type over the
// whole life of this server.
type InstanceUsage struct {
	Provider  string  `json:"provider"`
	Flows     uint32  `json:"flows"`
	Chains    uint64  `json:"chains"`
	TokensIn  uint64  `json:"tokens_in"`
	TokensOut uint64  `json:"tokens_out"`
	CacheIn   uint64  `json:"cache_in"`
	CacheOut  uint64  `json:"cache_out"`
	CostIn    float64 `json:"cost_in"`
	CostOut   float64 `json:"cost_out"`
}

// SummaryOptions carries the facts that are not in the database.
type SummaryOptions struct {
	Features InstanceFeatures
	// OAuth is the set of configured OAuth provider names — the names of the
	// providers, not of any user.
	OAuth []string
	// DefaultProviders lists the built-in provider types this instance would
	// construct from its environment configuration (see
	// providers.EnabledDefaultProviderTypes). They never get a row in the
	// `providers` table — that table only holds per-user providers added
	// through the Settings UI — so without this the summary would undercount,
	// or on an installation that configures everything through the
	// environment, report no providers at all despite active usage.
	DefaultProviders []string
	// EmbeddingProvider is the provider type actually used for embeddings, as
	// resolved by the caller (see embeddings.ResolvedProviderType).
	EmbeddingProvider string
}

// GetInstanceSummary describes this server.
//
// Every aggregate is a single statement over a single table. The analytics the
// UI runs are shaped for a different question — they walk one flow at a time and
// combine the rows in Go — and reusing them here would make describing a server
// cost as much as rendering a dashboard for it.
//
// It does not fail as a whole. An aggregate that cannot be produced is named in
// Skipped and the rest of the document is still returned, because a description
// missing one number is useful and an error instead of a description is not.
func (q *Queries) GetInstanceSummary(ctx context.Context, opts SummaryOptions) InstanceSummary {
	summary := InstanceSummary{
		Schema:  InstanceSummarySchema,
		Version: version.GetBinaryVersion(),
		Develop: version.IsDevelopMode(),
		// Read off the running process rather than taken from the caller, for the
		// same reason the version is: every caller would compute the identical
		// value, and one that got it wrong — or that was updated to fill it and one
		// that was not — would put two answers to the same question in circulation.
		OS:                runtime.GOOS,
		Arch:              runtime.GOARCH,
		BinaryHash:        version.BinaryHash(),
		Features:          opts.Features,
		Providers:         []InstanceProvider{},
		OAuth:             []string{},
		EmbeddingProvider: opts.EmbeddingProvider,
		Usage:             []InstanceUsage{},
	}
	if opts.OAuth != nil {
		summary.OAuth = append(summary.OAuth, opts.OAuth...)
		sort.Strings(summary.OAuth)
	}

	// Counters, each independent of the others: one that cannot be produced must
	// not cost the ones that can.
	counters := []struct {
		name  string
		query func(context.Context) (int64, error)
		into  func(uint64)
	}{
		{"users", q.CountUsers,
			func(v uint64) { summary.Counts.Users = uint32(v) }},
		{"flows", q.CountFlows,
			func(v uint64) { summary.Counts.Flows = uint32(v) }},
		{"flows_active", q.CountActiveFlows,
			func(v uint64) { summary.Counts.FlowsActive = uint32(v) }},
		{"tasks", q.CountTasks,
			func(v uint64) { summary.Counts.Tasks = uint32(v) }},
		{"subtasks", q.CountSubtasks,
			func(v uint64) { summary.Counts.Subtasks = uint32(v) }},
		{"containers", q.CountContainers,
			func(v uint64) { summary.Counts.Containers = uint32(v) }},
		{"assistants", q.CountAssistants,
			func(v uint64) { summary.Counts.Assistants = uint32(v) }},
		{"flow_templates", q.CountFlowTemplates,
			func(v uint64) { summary.Counts.FlowTemplates = uint32(v) }},
		{"prompts", q.CountPrompts,
			func(v uint64) { summary.Counts.Prompts = uint32(v) }},
		{"api_tokens", q.CountAPITokens,
			func(v uint64) { summary.Counts.APITokens = uint32(v) }},
		{"toolcalls", q.CountToolcalls,
			func(v uint64) { summary.Counts.Toolcalls = v }},
	}
	for _, counter := range counters {
		value, err := q.countOne(ctx, counter.query)
		if err != nil {
			summary.Skipped = append(summary.Skipped, counter.name)
			continue
		}
		counter.into(value)
	}

	if providers, err := q.summaryProviders(ctx); err != nil {
		summary.Skipped = append(summary.Skipped, "providers")
	} else {
		summary.Providers = providers
	}
	summary.Providers = mergeDefaultProviderCounts(summary.Providers, opts.DefaultProviders)

	if usage, err := q.summaryUsage(ctx); err != nil {
		summary.Skipped = append(summary.Skipped, "usage")
	} else {
		summary.Usage = usage
	}

	return summary
}

// MarshalIndent renders the summary as the document callers print or forward.
//
// Indented because the first consumer of this is a person reading a terminal,
// and a wall of one line is not a description of anything.
func (s InstanceSummary) MarshalIndent() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

func (q *Queries) countOne(ctx context.Context, query func(context.Context) (int64, error)) (uint64, error) {
	queryCtx, cancel := context.WithTimeout(ctx, summaryQueryTimeout)
	defer cancel()

	count, err := query(queryCtx)
	if err != nil {
		return 0, err
	}
	if count < 0 {
		return 0, nil
	}
	return uint64(count), nil
}

// mergeDefaultProviderCounts folds the environment-configured (built-in)
// provider types into the per-type counts read from the `providers` table.
// Those types never get a database row — see SummaryOptions.DefaultProviders
// — so without this step a type used only through the environment would be
// invisible here even while it accumulates real usage.
func mergeDefaultProviderCounts(dbProviders []InstanceProvider, defaultTypes []string) []InstanceProvider {
	counts := make(map[string]uint32, len(dbProviders)+len(defaultTypes))
	for _, p := range dbProviders {
		counts[p.Type] = p.Count
	}
	for _, t := range defaultTypes {
		counts[t]++
	}

	merged := make([]InstanceProvider, 0, len(counts))
	for t, c := range counts {
		merged = append(merged, InstanceProvider{Type: t, Count: c})
	}
	sort.Slice(merged, func(i, j int) bool { return merged[i].Type < merged[j].Type })

	return merged
}

// summaryProviders reports which provider types are configured and how many of
// each. One grouped scan of a small table.
func (q *Queries) summaryProviders(ctx context.Context) ([]InstanceProvider, error) {
	queryCtx, cancel := context.WithTimeout(ctx, summaryQueryTimeout)
	defer cancel()

	rows, err := q.GetSummaryProviderCounts(queryCtx)
	if err != nil {
		return nil, err
	}

	providers := make([]InstanceProvider, 0, len(rows))
	for _, row := range rows {
		providers = append(providers, InstanceProvider{Type: string(row.Type), Count: uint32(row.Total)})
	}

	return providers, nil
}

// summaryUsage reports consumption grouped by provider type.
//
// One grouped scan of msgchains and nothing else. The per-flow variant of this
// question joins subtasks, tasks and flows to attribute each chain to a flow;
// none of that is needed to answer "how much has gone through each provider",
// and on a busy instance those joins are the difference between a description
// and an outage.
func (q *Queries) summaryUsage(ctx context.Context) ([]InstanceUsage, error) {
	queryCtx, cancel := context.WithTimeout(ctx, summaryQueryTimeout)
	defer cancel()

	rows, err := q.GetSummaryUsageByProvider(queryCtx)
	if err != nil {
		return nil, err
	}

	usage := make([]InstanceUsage, 0, len(rows))
	for _, row := range rows {
		usage = append(usage, InstanceUsage{
			Provider:  row.ModelProvider,
			Flows:     uint32(row.Flows),
			Chains:    uint64(row.Chains),
			TokensIn:  uint64(row.TokensIn),
			TokensOut: uint64(row.TokensOut),
			CacheIn:   uint64(row.CacheIn),
			CacheOut:  uint64(row.CacheOut),
			CostIn:    row.CostIn,
			CostOut:   row.CostOut,
		})
	}

	return usage, nil
}
