package tools

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/tools/searchers"
)

// searchLogProviderMock is a test double for SearchLogProvider. It moved here from the
// engines' proxy_test.go when those engines were extracted into the searchers
// subpackage: the searchers no longer write to the DB search-log (the web_search
// orchestrator does), so the mock now lives alongside the orchestrator's tests.
var _ SearchLogProvider = &searchLogProviderMock{}

type searchLogProviderMock struct {
	calls      int64
	engine     database.SearchengineType
	query      string
	result     string
	taskID     *int64
	subtaskID  *int64
	parentType database.MsgchainType
	currType   database.MsgchainType
}

func (m *searchLogProviderMock) PutLog(
	_ context.Context,
	initiator database.MsgchainType,
	executor database.MsgchainType,
	engine database.SearchengineType,
	query string,
	result string,
	taskID *int64,
	subtaskID *int64,
) (int64, error) {
	m.calls++
	m.parentType = initiator
	m.currType = executor
	m.engine = engine
	m.query = query
	m.result = result
	m.taskID = taskID
	m.subtaskID = subtaskID
	return m.calls, nil
}

// fakeSearcher is a scriptable searchers.Searcher for orchestrator tests. It records
// the global call order (via order) so tests can assert the fallback sequence.
type fakeSearcher struct {
	engine    database.SearchengineType
	available bool
	handler   func(call int) (string, error)
	calls     int
	order     *[]database.SearchengineType
}

func (f *fakeSearcher) Engine() database.SearchengineType { return f.engine }
func (f *fakeSearcher) IsAvailable() bool                 { return f.available }

func (f *fakeSearcher) Handle(_ context.Context, _ searchers.Request) (string, error) {
	if f.order != nil {
		*f.order = append(*f.order, f.engine)
	}
	c := f.calls
	f.calls++
	if f.handler == nil {
		return "", searchers.Fatal(errors.New("no handler"))
	}
	return f.handler(c)
}

func newTestWebSearch(slp SearchLogProvider, engines map[database.SearchengineType]searchers.Searcher) *webSearch {
	return &webSearch{
		cfg:     &config.Config{},
		flowID:  1,
		slp:     slp,
		engines: engines,
	}
}

func webSearchArgs(t *testing.T, mode, query string) []byte {
	t.Helper()
	if mode == "" {
		return []byte(`{"query":"` + query + `","message":"m"}`)
	}
	return []byte(`{"query":"` + query + `","mode":"` + mode + `","message":"m"}`)
}

func ctxWithAgent() context.Context {
	return PutAgentContext(context.Background(), database.MsgchainTypeSearcher)
}

// TestWebSearchStrategyOrder verifies the orchestrator walks fallbackStrategy in order
// and skips unavailable engines.
func TestWebSearchStrategyOrder(t *testing.T) {
	for _, mode := range []SearchMode{ModeLinks, ModeAnswer, ModeResearch, ModeExploit} {
		t.Run(string(mode), func(t *testing.T) {
			var order []database.SearchengineType
			engines := map[database.SearchengineType]searchers.Searcher{}
			for _, id := range fallbackStrategy[mode] {
				engines[id] = &fakeSearcher{
					engine:    id,
					available: true,
					order:     &order,
					handler:   func(int) (string, error) { return "", searchers.Fatal(errors.New("boom")) },
				}
			}

			ws := newTestWebSearch(&searchLogProviderMock{}, engines)
			_, err := ws.Handle(ctxWithAgent(), WebSearchToolName, webSearchArgs(t, string(mode), "q"))
			if err != nil {
				t.Fatalf("Handle() unexpected hard error: %v", err)
			}

			want := fallbackStrategy[mode]
			if !reflect.DeepEqual(order, want) {
				t.Errorf("call order = %v, want %v", order, want)
			}
		})
	}
}

func TestWebSearchSkipsUnavailable(t *testing.T) {
	var order []database.SearchengineType
	// answer chain: tavily(unavail), traversaal(avail, succeeds)
	engines := map[database.SearchengineType]searchers.Searcher{
		EngineTavily:     &fakeSearcher{engine: EngineTavily, available: false, order: &order},
		EngineTraversaal: &fakeSearcher{engine: EngineTraversaal, available: true, order: &order, handler: func(int) (string, error) { return "ok", nil }},
	}
	slp := &searchLogProviderMock{}
	ws := newTestWebSearch(slp, engines)

	got, err := ws.Handle(ctxWithAgent(), WebSearchToolName, webSearchArgs(t, "answer", "q"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ok" {
		t.Errorf("result = %q, want %q", got, "ok")
	}
	if len(order) != 1 || order[0] != EngineTraversaal {
		t.Errorf("call order = %v, want [traversaal] (tavily unavailable must be skipped)", order)
	}
	if slp.calls != 1 || slp.engine != EngineTraversaal {
		t.Errorf("search log: calls=%d engine=%q, want 1 traversaal", slp.calls, slp.engine)
	}
}

func TestWebSearchFatalMovesToNextEngineWithoutRetry(t *testing.T) {
	var order []database.SearchengineType
	first := &fakeSearcher{engine: EngineTavily, available: true, order: &order,
		handler: func(int) (string, error) { return "", searchers.Fatal(errors.New("401 auth")) }}
	second := &fakeSearcher{engine: EngineTraversaal, available: true, order: &order,
		handler: func(int) (string, error) { return "second-result", nil }}
	engines := map[database.SearchengineType]searchers.Searcher{
		EngineTavily:     first,
		EngineTraversaal: second,
	}
	slp := &searchLogProviderMock{}
	ws := newTestWebSearch(slp, engines)

	got, err := ws.Handle(ctxWithAgent(), WebSearchToolName, webSearchArgs(t, "answer", "q"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "second-result" {
		t.Errorf("result = %q, want second-result", got)
	}
	if first.calls != 1 {
		t.Errorf("first engine called %d times, want 1 (fatal must not retry)", first.calls)
	}
	if slp.engine != EngineTraversaal {
		t.Errorf("search log engine = %q, want traversaal (the winner)", slp.engine)
	}
}

func TestWebSearchRetriesRetryableThenSucceeds(t *testing.T) {
	engine := &fakeSearcher{engine: EngineTavily, available: true, handler: func(call int) (string, error) {
		if call == 0 {
			return "", searchers.Retryable(errors.New("429"), 0)
		}
		return "recovered", nil
	}}
	engines := map[database.SearchengineType]searchers.Searcher{EngineTavily: engine}
	ws := newTestWebSearch(&searchLogProviderMock{}, engines)

	got, err := ws.Handle(ctxWithAgent(), WebSearchToolName, webSearchArgs(t, "answer", "q"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "recovered" {
		t.Errorf("result = %q, want recovered", got)
	}
	if engine.calls != 2 {
		t.Errorf("engine called %d times, want 2 (one retry)", engine.calls)
	}
}

func TestWebSearchRetryableExhaustsThenMovesOn(t *testing.T) {
	persistent429 := &fakeSearcher{engine: EngineTavily, available: true,
		handler: func(int) (string, error) { return "", searchers.Retryable(errors.New("429"), 0) }}
	next := &fakeSearcher{engine: EngineTraversaal, available: true,
		handler: func(int) (string, error) { return "ok", nil }}
	engines := map[database.SearchengineType]searchers.Searcher{
		EngineTavily:     persistent429,
		EngineTraversaal: next,
	}
	ws := newTestWebSearch(&searchLogProviderMock{}, engines)

	got, err := ws.Handle(ctxWithAgent(), WebSearchToolName, webSearchArgs(t, "answer", "q"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ok" {
		t.Errorf("result = %q, want ok", got)
	}
	if persistent429.calls != maxEngineAttempts {
		t.Errorf("retryable engine called %d times, want %d", persistent429.calls, maxEngineAttempts)
	}
}

func TestWebSearchAllExhaustedReturnsSoftError(t *testing.T) {
	engines := map[database.SearchengineType]searchers.Searcher{
		EngineTavily:     &fakeSearcher{engine: EngineTavily, available: true, handler: func(int) (string, error) { return "", searchers.Fatal(errors.New("down")) }},
		EngineTraversaal: &fakeSearcher{engine: EngineTraversaal, available: true, handler: func(int) (string, error) { return "", searchers.Fatal(errors.New("down")) }},
	}
	slp := &searchLogProviderMock{}
	ws := newTestWebSearch(slp, engines)

	got, err := ws.Handle(ctxWithAgent(), WebSearchToolName, webSearchArgs(t, "answer", "q"))
	// Exhaustion is a soft failure: message + nil error (must NOT burn agent retries).
	if err != nil {
		t.Fatalf("expected soft (nil) error on exhaustion, got hard error: %v", err)
	}
	if got == "" {
		t.Fatal("expected a soft message, got empty string")
	}
	if !strings.Contains(got, "all") || !strings.Contains(got, "failed") {
		t.Errorf("soft message = %q, want it to explain the exhaustion", got)
	}
	if slp.calls != 1 {
		t.Errorf("search log calls = %d, want 1 failure row", slp.calls)
	}
}

func TestWebSearchNothingConfiguredReturnsSoftError(t *testing.T) {
	// No engine available for the mode.
	engines := map[database.SearchengineType]searchers.Searcher{
		EngineTavily: &fakeSearcher{engine: EngineTavily, available: false},
	}
	slp := &searchLogProviderMock{}
	ws := newTestWebSearch(slp, engines)

	got, err := ws.Handle(ctxWithAgent(), WebSearchToolName, webSearchArgs(t, "answer", "q"))
	if err != nil {
		t.Fatalf("expected soft error, got hard error: %v", err)
	}
	if !strings.Contains(got, "no search engine is configured") {
		t.Errorf("message = %q, want the 'not configured' guidance", got)
	}
	if slp.calls != 0 {
		t.Errorf("search log calls = %d, want 0 (nothing was attempted)", slp.calls)
	}
}

func TestWebSearchMalformedArgsIsHardError(t *testing.T) {
	ws := newTestWebSearch(&searchLogProviderMock{}, map[database.SearchengineType]searchers.Searcher{})
	_, err := ws.Handle(ctxWithAgent(), WebSearchToolName, []byte("{not json"))
	if err == nil {
		t.Fatal("expected a hard error for malformed args (fixable by the arg-fixer)")
	}
}

func TestWebSearchEmptyQueryIsHardError(t *testing.T) {
	ws := newTestWebSearch(&searchLogProviderMock{}, map[database.SearchengineType]searchers.Searcher{})
	_, err := ws.Handle(ctxWithAgent(), WebSearchToolName, []byte(`{"query":"  ","message":"m"}`))
	if err == nil {
		t.Fatal("expected a hard error for an empty query")
	}
}

func TestWebSearchDefaultMode(t *testing.T) {
	// Omitting mode must resolve to defaultMode (answer): the answer chain leads with
	// tavily, so a tavily-only registry should be used.
	var order []database.SearchengineType
	engines := map[database.SearchengineType]searchers.Searcher{
		EngineTavily: &fakeSearcher{engine: EngineTavily, available: true, order: &order, handler: func(int) (string, error) { return "ok", nil }},
	}
	ws := newTestWebSearch(&searchLogProviderMock{}, engines)

	got, err := ws.Handle(ctxWithAgent(), WebSearchToolName, webSearchArgs(t, "", "q"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ok" {
		t.Errorf("result = %q, want ok", got)
	}
	if len(order) == 0 || order[0] != EngineTavily {
		t.Errorf("default mode did not use the answer chain (order=%v)", order)
	}
}

func TestNormalizeMode(t *testing.T) {
	cases := map[string]SearchMode{
		"links": ModeLinks, "answer": ModeAnswer, "research": ModeResearch, "exploit": ModeExploit,
		"": defaultMode, "bogus": defaultMode, "ANSWER": ModeAnswer, " research ": ModeResearch,
	}
	for in, want := range cases {
		if got := normalizeMode(in); got != want {
			t.Errorf("normalizeMode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestClampResults(t *testing.T) {
	cases := map[int]int{0: webSearchDefaultResults, -5: webSearchDefaultResults, 1: 1, 25: 25, 100: webSearchMaxResults}
	for in, want := range cases {
		if got := clampResults(in); got != want {
			t.Errorf("clampResults(%d) = %d, want %d", in, got, want)
		}
	}
}
