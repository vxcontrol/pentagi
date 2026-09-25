package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"runtime"
	"strings"
	"sync"
	"testing"

	"pentagi/pkg/version"

	"github.com/stretchr/testify/assert"
	"github.com/vxcontrol/cloud/models"
)

// A driver rather than a fake DBTX: the only *sql.Row a test can build panics on Scan instead of failing.
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

// answersNothing opens a database that connects, refuses every statement and hands back the statements asked, in order.
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

func TestSummary_GetInstanceSummary_SurvivesADatabaseThatAnswersNothing(t *testing.T) {
	queries, _ := answersNothing(t)

	summary := queries.GetInstanceSummary(context.Background(), SummaryOptions{})

	assert.Equal(t, 1, summary.Schema)
	assert.NotEmpty(t, summary.Version, "the version is what a reader groups by, so it is there even when nothing else is")
	assert.Equal(t, []string{
		"users", "flows", "flows_active", "tasks", "subtasks", "containers", "assistants",
		"flow_templates", "prompts", "api_tokens", "toolcalls", "providers", "usage",
	}, summary.Skipped, "a zero nobody could measure must not read as a measured empty installation")
	// A nil list marshals to null, which a reader would have to special-case.
	assert.NotNil(t, summary.Providers)
	assert.NotNil(t, summary.OAuth)
	assert.NotNil(t, summary.Usage)
}

// An allowlist: widening what the document may carry is a decision somebody has to write down here.
func TestSummary_GetInstanceSummary_SelectsOnlyQuantitiesFromOneTableAtATime(t *testing.T) {
	queries, taken := answersNothing(t)
	queries.GetInstanceSummary(context.Background(), SummaryOptions{})
	issued := taken()

	if len(issued) == 0 {
		t.Fatal("no statements were issued; the checks below would pass vacuously")
	}

	// The only bare columns: closed vocabularies defined in this repository, never anything a user typed.
	allowed := map[string]bool{"type": true, "model_provider": true}

	for _, query := range issued {
		if strings.Contains(strings.ToLower(strings.Join(strings.Fields(query), " ")), " join ") {
			t.Errorf("a join crept into the description; it has to stay cheap enough to run unattended:\n%s", query)
		}

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
			if strings.HasPrefix(item, "count(") || strings.HasPrefix(item, "coalesce(sum(") ||
				strings.HasPrefix(item, "sum(") {
				continue
			}
			if allowed[item] {
				continue
			}
			t.Errorf("the description projects %q, which is neither a quantity nor a closed "+
				"vocabulary:\n%s", item, query)
		}
	}
}

// splitProjection splits on the commas outside parentheses: the one in COALESCE(SUM(x), 0) belongs to the call.
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

func TestSummary_GetInstanceSummary_CarriesWhatOnlyTheCallerKnows(t *testing.T) {
	queries, _ := answersNothing(t)

	summary := queries.GetInstanceSummary(context.Background(), SummaryOptions{
		Features:          InstanceFeatures{AskUser: true, DockerInside: true},
		OAuth:             []string{"google", "github"},
		DefaultProviders:  []string{"openai", "openai", "anthropic"},
		EmbeddingProvider: "custom",
	})

	assert.Equal(t, InstanceFeatures{AskUser: true, DockerInside: true}, summary.Features)
	assert.Equal(t, []string{"github", "google"}, summary.OAuth, "sorted, so a diff between two reports means something changed")
	assert.Equal(t, "custom", summary.EmbeddingProvider)
	assert.Equal(t, []InstanceProvider{{Type: "anthropic", Count: 1}, {Type: "openai", Count: 2}}, summary.Providers,
		"a provider configured only through the environment has no row, and each entry counts")
}

func TestSummary_MergeDefaultProviderCounts_AddsDefaultsToTheStoredCounts(t *testing.T) {
	merged := mergeDefaultProviderCounts(
		[]InstanceProvider{{Type: "openai", Count: 3}, {Type: "custom", Count: 1}},
		[]string{"openai", "anthropic"},
	)

	assert.Equal(t, []InstanceProvider{
		{Type: "anthropic", Count: 1}, {Type: "custom", Count: 1}, {Type: "openai", Count: 4},
	}, merged)
}

// The platform and the digest describe the process that ran, which a caller outside a container cannot supply.
func TestSummary_GetInstanceSummary_DescribesTheRunningProcess(t *testing.T) {
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

	// In the rendered document too: an omitempty would make an unknown platform look like an old document.
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

func TestSummary_InstanceSummary_FitsWhatTheUpdateContractAccepts(t *testing.T) {
	summary := InstanceSummary{
		Schema:  InstanceSummarySchema,
		Version: "2.1.0",
	}
	for _, provider := range []string{
		"openai", "anthropic", "gemini", "bedrock", "ollama", "custom",
		"deepseek", "glm", "kimi", "qwen", "minimax",
	} {
		summary.Providers = append(summary.Providers,
			InstanceProvider{Type: provider, Count: 3})
		summary.Usage = append(summary.Usage, InstanceUsage{
			Provider: provider, Chains: 99999, TokensIn: 123456789, TokensOut: 987654321,
			CacheIn: 1234567, CacheOut: 7654321, CostIn: 12345.67, CostOut: 7654.32,
		})
	}
	summary.OAuth = []string{"google", "github"}
	summary.Skipped = []string{"toolcalls", "usage"}

	document, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("the description must serialise: %v", err)
	}
	if len(document) > models.MaxProductInfoBytes {
		t.Errorf("a full description is %d bytes, over the %d the contract accepts — "+
			"every report from a fully configured installation would be refused whole",
			len(document), models.MaxProductInfoBytes)
	}
}

func TestSummary_InstanceSummary_CarriesOnlyKeysSomebodyDecidedToSend(t *testing.T) {
	summary := InstanceSummary{
		Schema:    InstanceSummarySchema,
		Version:   "2.1.0",
		Providers: []InstanceProvider{{Type: "openai", Count: 2}},
		OAuth:     []string{"google"},
		Usage:     []InstanceUsage{{Provider: "openai", TokensIn: 10}},
	}
	summary.Counts.Users = 3
	summary.Counts.Flows = 12

	document, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("serialise: %v", err)
	}

	var shape map[string]json.RawMessage
	if err := json.Unmarshal(document, &shape); err != nil {
		t.Fatalf("the document must be an object: %v", err)
	}

	allowed := map[string]bool{
		"schema": true, "version": true, "develop": true, "features": true,
		"providers": true, "oauth": true, "counts": true, "usage": true, "skipped": true,
		"embedding_provider": true, "os": true, "arch": true, "binary_hash": true,
	}
	for key := range shape {
		if !allowed[key] {
			t.Errorf("the document carries %q, which nobody decided to send — every key here "+
				"has to be a deliberate choice, because this leaves the machine", key)
		}
	}

	if strings.Contains(strings.ToLower(string(document)), "@") {
		t.Errorf("the document contains an address-shaped value: %s", document)
	}
}
