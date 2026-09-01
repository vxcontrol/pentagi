package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"runtime"
	"strings"
	"sync"
	"testing"

	"pentagi/pkg/version"
)

// The stand-in for a database that is reachable but answers nothing: a driver
// rather than a hand-rolled DBTX.
//
// A fake DBTX cannot be built honestly. *sql.Row carries its error in unexported
// fields, so the only *sql.Row a test can construct is the zero value — and that
// one PANICS on Scan rather than returning an error, which is a different thing
// entirely from what a broken database does. Going through database/sql with a
// failing driver exercises the real machinery and records what was asked.
var (
	errStubDriver     = errors.New("stub driver: nothing is answered here")
	registerStubOnce  sync.Once
	stubQueryRecorder = &queryRecorder{}
)

type queryRecorder struct {
	mu      sync.Mutex
	queries []string
}

func (r *queryRecorder) record(query string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.queries = append(r.queries, query)
}

func (r *queryRecorder) taken() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	taken := append([]string(nil), r.queries...)
	r.queries = nil
	return taken
}

type stubDriver struct{}

func (stubDriver) Open(string) (driver.Conn, error) { return stubConn{}, nil }

type stubConn struct{}

func (stubConn) Prepare(query string) (driver.Stmt, error) {
	stubQueryRecorder.record(query)
	return nil, errStubDriver
}
func (stubConn) Close() error              { return nil }
func (stubConn) Begin() (driver.Tx, error) { return nil, errStubDriver }

// answersNothing opens a database that connects and then refuses every statement,
// and hands back the statements it was asked, in order.
func answersNothing(t *testing.T) (*Queries, func() []string) {
	t.Helper()

	registerStubOnce.Do(func() { sql.Register("pentagi-summary-stub", stubDriver{}) })

	db, err := sql.Open("pentagi-summary-stub", "")
	if err != nil {
		t.Fatalf("open stub database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	stubQueryRecorder.taken()
	return New(db), stubQueryRecorder.taken
}

// TestADescriptionSurvivesADatabaseThatAnswersNothing.
//
// The document is produced to be read by something that has already decided to
// ask. Returning an error instead of a document turns "this server is having
// trouble" into "this server said nothing", and those are the two readings the
// whole exercise exists to tell apart.
func TestADescriptionSurvivesADatabaseThatAnswersNothing(t *testing.T) {
	queries, _ := answersNothing(t)

	summary := queries.GetInstanceSummary(context.Background(), SummaryOptions{})

	if summary.Schema != InstanceSummarySchema {
		t.Errorf("schema lost: %d", summary.Schema)
	}
	if summary.Version == "" {
		t.Error("the version must be present even when nothing else is: it is what a reader groups by")
	}
	if len(summary.Skipped) == 0 {
		t.Fatal("nothing could be gathered, yet nothing was reported as skipped — a page of zeroes " +
			"would read as a measured empty installation")
	}
	for _, name := range []string{"users", "flows", "toolcalls", "providers", "usage"} {
		if !contains(summary.Skipped, name) {
			t.Errorf("%q could not be gathered but is not named in skipped: %v", name, summary.Skipped)
		}
	}

	// Empty rather than nil, all the way out. A nil slice marshals to `null`, and
	// a reader that walks the list then has to special-case a value that means the
	// same thing as an empty list.
	if summary.Providers == nil || summary.OAuth == nil || summary.Usage == nil {
		t.Error("list fields left nil; they must marshal as empty lists, not as null")
	}
}

// TestTheDescriptionSelectsOnlyQuantitiesAndClosedVocabularies.
//
// The promise this document makes is that it holds counts and closed vocabularies
// and nothing else. That promise lives in the SELECT list of every statement, and
// nowhere else: reading a row of `prompts` to count it is a quantity, reading its
// text is a person's words, and the two statements differ only in their
// projection.
//
// An allowlist rather than a blocklist, deliberately. A blocklist passes every
// column nobody thought to forbid, which is exactly the column that gets added
// next year; this fails on anything not named here, so widening the document is a
// decision somebody has to write down.
func TestTheDescriptionSelectsOnlyQuantitiesAndClosedVocabularies(t *testing.T) {
	queries, taken := answersNothing(t)
	queries.GetInstanceSummary(context.Background(), SummaryOptions{})
	issued := taken()

	if len(issued) == 0 {
		t.Fatal("no statements were issued; the check below would pass vacuously")
	}

	// The only bare columns a description may project. Both are closed
	// vocabularies defined in this repository — a provider type and the provider a
	// chain ran through — so neither can carry anything a user typed.
	allowed := map[string]bool{"type": true, "model_provider": true}

	for _, query := range issued {
		projection := selectList(t, query)
		if projection == "" {
			t.Errorf("could not read the projection of:\n%s", query)
			continue
		}
		for _, item := range splitProjection(projection) {
			item = strings.ToLower(strings.TrimSpace(item))
			if item == "" {
				continue
			}
			// Anything that is an aggregate is a quantity by construction: it
			// reduces however many rows to one number.
			if strings.HasPrefix(item, "count(") || strings.HasPrefix(item, "coalesce(sum(") ||
				strings.HasPrefix(item, "sum(") {
				continue
			}
			if allowed[item] {
				continue
			}
			t.Errorf("the description projects %q, which is neither a quantity nor a closed "+
				"vocabulary — this is how a person's data reaches a document that promises "+
				"neither:\n%s", item, query)
		}
	}
}

// splitProjection splits a SELECT list on the commas that separate its items —
// which is not every comma. `COALESCE(SUM(usage_in), 0)` contains one that
// belongs to the call, and splitting on it turns a legitimate aggregate into two
// fragments, neither of which is anything.
func splitProjection(projection string) []string {
	var (
		items []string
		depth int
		start int
	)
	for i, char := range projection {
		switch char {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				items = append(items, projection[start:i])
				start = i + 1
			}
		}
	}
	return append(items, projection[start:])
}

// selectList returns everything between SELECT and the matching FROM.
func selectList(t *testing.T, query string) string {
	t.Helper()

	flat := strings.Join(strings.Fields(query), " ")
	lowered := strings.ToLower(flat)

	start := strings.Index(lowered, "select ")
	if start < 0 {
		return ""
	}
	end := strings.Index(lowered, " from ")
	if end < 0 || end <= start {
		return ""
	}
	return flat[start+len("select ") : end]
}

// TestTheDescriptionAsksOneTableAtATime.
//
// The cost of this document is the reason it can be produced on a schedule. The
// analytics behind the product's own dashboard join msgchains to subtasks to
// tasks to flows in order to attribute consumption per flow; none of that is
// needed to answer "how much went through each provider", and on an installation
// large enough to be interesting those joins are the difference between a
// description and an outage.
func TestTheDescriptionAsksOneTableAtATime(t *testing.T) {
	queries, taken := answersNothing(t)
	queries.GetInstanceSummary(context.Background(), SummaryOptions{})

	for _, query := range taken() {
		lowered := strings.ToLower(query)
		if strings.Contains(lowered, " join ") {
			t.Errorf("a join crept into the description; it is meant to stay cheap enough to run "+
				"unattended:\n%s", query)
		}
	}
}

// TestOptionsAreCarriedThroughRatherThanGuessed pins that the facts which are not
// in the database arrive from the caller. They describe how the server is
// configured, and the database has no opinion about that.
func TestOptionsAreCarriedThroughRatherThanGuessed(t *testing.T) {
	queries, _ := answersNothing(t)

	summary := queries.GetInstanceSummary(context.Background(), SummaryOptions{
		Features:          InstanceFeatures{AskUser: true, DockerInside: true},
		OAuth:             []string{"google", "github"},
		DefaultProviders:  []string{"anthropic", "custom"},
		EmbeddingProvider: "custom",
	})

	if !summary.Features.AskUser || !summary.Features.DockerInside {
		t.Errorf("configuration switches lost: %+v", summary.Features)
	}
	if summary.Features.Debug || summary.Features.AssistantUseAgents {
		t.Errorf("switches invented that the caller did not set: %+v", summary.Features)
	}
	// Sorted, so that two installations with the same providers produce the same
	// document and a diff between two reports means something changed.
	if len(summary.OAuth) != 2 || summary.OAuth[0] != "github" || summary.OAuth[1] != "google" {
		t.Errorf("oauth providers lost or unordered: %v", summary.OAuth)
	}
	if summary.EmbeddingProvider != "custom" {
		t.Errorf("embedding provider lost: %q", summary.EmbeddingProvider)
	}
	if !contains(providerTypes(summary.Providers), "anthropic") || !contains(providerTypes(summary.Providers), "custom") {
		t.Errorf("default providers lost: %+v", summary.Providers)
	}
}

// TestTheProcessDescribesItselfRatherThanBeingTold.
//
// The platform and the executable digest are read off the running process, not taken
// from SummaryOptions like the configuration switches beside them, and the difference
// is the point. Every caller would compute the same value, so a caller that got it
// wrong — or one that was updated to supply it while another was not — would put two
// answers to "which build is running" into circulation. There are two callers today:
// the `-info` flag and the periodic report.
//
// And they have to be in the DOCUMENT, not only in the request the report sends,
// because a description obtained by executing this binary inside a container is
// forwarded by somebody who cannot produce either fact: that caller knows the
// platform of the HOST, and the image digest it can see names a filesystem rather
// than the file that is executing.
func TestTheProcessDescribesItselfRatherThanBeingTold(t *testing.T) {
	queries, _ := answersNothing(t)

	summary := queries.GetInstanceSummary(context.Background(), SummaryOptions{})

	if summary.OS != runtime.GOOS || summary.Arch != runtime.GOARCH {
		t.Errorf("platform = %q/%q, want the platform this binary was compiled for (%q/%q)",
			summary.OS, summary.Arch, runtime.GOOS, runtime.GOARCH)
	}
	if summary.BinaryHash != version.BinaryHash() {
		t.Errorf("binary hash = %q, want the digest of the running executable", summary.BinaryHash)
	}
	if len(summary.BinaryHash) != 64 {
		t.Errorf("binary hash = %q, want a sha256 in hex", summary.BinaryHash)
	}

	// And all three have to be IN the rendered document, not merely on the struct.
	// An `omitempty` added later would make an unknown platform indistinguishable
	// from a document too old to carry one, on the route where nothing else can
	// supply it.
	document, err := summary.MarshalIndent()
	if err != nil {
		t.Fatalf("the description must serialise: %v", err)
	}
	for _, key := range []string{`"os"`, `"arch"`, `"binary_hash"`} {
		if !strings.Contains(string(document), key) {
			t.Errorf("the document does not carry %s: %s", key, document)
		}
	}
}

// TestDefaultProvidersFillTheGapTheDatabaseCannot pins the reason
// DefaultProviders exists at all: an installation configured entirely through
// the environment never gets a row in the `providers` table (that table only
// holds providers added through the Settings UI), so without this the summary
// would report zero providers despite every one of them being usable.
func TestDefaultProvidersFillTheGapTheDatabaseCannot(t *testing.T) {
	queries, _ := answersNothing(t)

	summary := queries.GetInstanceSummary(context.Background(), SummaryOptions{
		DefaultProviders: []string{"openai", "openai", "anthropic"},
	})

	// A type present only via configuration must still show up, with a count
	// of one per default provider entry — here "openai" appears twice on
	// purpose, to pin that repeats accumulate rather than collapsing to one.
	got := make(map[string]uint32, len(summary.Providers))
	for _, p := range summary.Providers {
		got[p.Type] = p.Count
	}
	if got["openai"] != 2 {
		t.Errorf("expected openai count 2 from two default entries, got %+v", summary.Providers)
	}
	if got["anthropic"] != 1 {
		t.Errorf("expected anthropic count 1, got %+v", summary.Providers)
	}
}

func providerTypes(providers []InstanceProvider) []string {
	types := make([]string, 0, len(providers))
	for _, p := range providers {
		types = append(types, p.Type)
	}
	return types
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
