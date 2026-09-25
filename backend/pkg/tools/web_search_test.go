package tools

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/tools/searchers"
)

// webSearchLogRecorder stands in for the search log, keeping the last row; like the database, it refuses a done context.
var _ SearchLogProvider = &webSearchLogRecorder{}

type webSearchLogRecorder struct {
	calls      int64
	engine     database.SearchengineType
	query      string
	result     string
	taskID     *int64
	subtaskID  *int64
	parentType database.MsgchainType
	currType   database.MsgchainType
}

func (m *webSearchLogRecorder) PutLog(
	ctx context.Context,
	initiator database.MsgchainType,
	executor database.MsgchainType,
	engine database.SearchengineType,
	query string,
	result string,
	taskID *int64,
	subtaskID *int64,
) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
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

// webSearchEngineDouble is a scriptable searchers.Searcher that, like an engine's HTTP request, fails on a done context.
type webSearchEngineDouble struct {
	engine    database.SearchengineType
	available bool
	handler   func(call int) (string, error)
	calls     int
	requests  []searchers.Request
	contexts  []context.Context
	order     *[]database.SearchengineType
}

func (f *webSearchEngineDouble) Engine() database.SearchengineType { return f.engine }
func (f *webSearchEngineDouble) IsAvailable() bool                 { return f.available }

func (f *webSearchEngineDouble) Handle(ctx context.Context, req searchers.Request) (string, error) {
	if f.order != nil {
		*f.order = append(*f.order, f.engine)
	}
	f.requests = append(f.requests, req)
	f.contexts = append(f.contexts, ctx)
	c := f.calls
	f.calls++
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if f.handler == nil {
		return "", searchers.Fatal(errors.New("no handler"))
	}
	return f.handler(c)
}

// webSearchLogRow prints the ids by value, so two rows naming the same task compare equal.
func webSearchLogRow(m *webSearchLogRecorder) string {
	id := func(p *int64) string {
		if p == nil {
			return "nil"
		}
		return strconv.FormatInt(*p, 10)
	}
	return fmt.Sprintf("calls=%d engine=%q query=%q result=%q task=%s subtask=%s initiator=%q executor=%q",
		m.calls, m.engine, m.query, m.result, id(m.taskID), id(m.subtaskID), m.parentType, m.currType)
}

func webSearchUnderTest(slp SearchLogProvider, engines map[database.SearchengineType]searchers.Searcher) *webSearch {
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

func webSearchAgentCtx() context.Context {
	return PutAgentContext(context.Background(), database.MsgchainTypeSearcher)
}

func webSearchFails(msg string) func(int) (string, error) {
	return func(int) (string, error) { return "", searchers.Fatal(errors.New(msg)) }
}

func webSearchAnswers(result string) func(int) (string, error) {
	return func(int) (string, error) { return result, nil }
}

// Every engine of every chain is registered and fails, so the calls are exactly the resolved mode's chain.
func TestWebSearch_Handle_WalksTheChainOfTheModeItResolves(t *testing.T) {
	seen := map[string]SearchMode{}
	for mode, chain := range fallbackStrategy {
		key := fmt.Sprint(chain)
		if other, ok := seen[key]; ok {
			t.Fatalf("modes %q and %q share one chain, so no row can tell which of them Handle resolved", mode, other)
		}
		seen[key] = mode
	}

	tests := []struct {
		name string
		mode string
		want SearchMode
	}{
		{name: "a request for links", mode: "links", want: ModeLinks},
		{name: "a request for an answer", mode: "answer", want: ModeAnswer},
		{name: "a request for research", mode: "research", want: ModeResearch},
		{name: "a request for exploits", mode: "exploit", want: ModeExploit},
		{name: "an omitted mode is answer", mode: "", want: ModeAnswer},
		{name: "an unknown mode is answer", mode: "bogus", want: ModeAnswer},
		{name: "an upper-case mode", mode: "LINKS", want: ModeLinks},
		{name: "a padded mode", mode: " research ", want: ModeResearch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var order []database.SearchengineType
			engines := map[database.SearchengineType]searchers.Searcher{}
			for _, chain := range fallbackStrategy {
				for _, id := range chain {
					engines[id] = &webSearchEngineDouble{
						engine: id, available: true, order: &order, handler: webSearchFails("boom"),
					}
				}
			}

			ws := webSearchUnderTest(&webSearchLogRecorder{}, engines)
			if _, err := ws.Handle(webSearchAgentCtx(), WebSearchToolName, webSearchArgs(t, tt.mode, "q")); err != nil {
				t.Fatalf("Handle() unexpected hard error: %v", err)
			}

			if want := fallbackStrategy[tt.want]; !reflect.DeepEqual(order, want) {
				t.Errorf("call order = %v, want the %s chain %v", order, tt.want, want)
			}
		})
	}
}

// Other ids of the answer chain have no engine and are skipped; the backoff is a lower bound so a slow runner cannot fail it.
func TestWebSearch_Handle_FallsBackAlongTheChain(t *testing.T) {
	tests := []struct {
		name            string
		tavilyAvailable bool
		tavilyHandler   func(call int) (string, error)
		wantResult      string
		wantOrder       []database.SearchengineType
		wantLogged      database.SearchengineType
		wantMinWait     time.Duration
	}{
		{
			name:            "an unavailable engine is skipped",
			tavilyAvailable: false,
			wantResult:      "second-result",
			wantOrder:       []database.SearchengineType{EngineTraversaal},
			wantLogged:      EngineTraversaal,
		},
		{
			name:            "a fatal error moves on without a retry",
			tavilyAvailable: true,
			tavilyHandler:   webSearchFails("401 auth"),
			wantResult:      "second-result",
			wantOrder:       []database.SearchengineType{EngineTavily, EngineTraversaal},
			wantLogged:      EngineTraversaal,
		},
		{
			name:            "a retryable error is retried on the same engine",
			tavilyAvailable: true,
			tavilyHandler: func(call int) (string, error) {
				if call == 0 {
					return "", searchers.Retryable(errors.New("429"), 0)
				}
				return "recovered", nil
			},
			wantResult:  "recovered",
			wantOrder:   []database.SearchengineType{EngineTavily, EngineTavily},
			wantLogged:  EngineTavily,
			wantMinWait: 300 * time.Millisecond,
		},
		{
			name:            "a retryable error that persists moves on once the two tries are spent",
			tavilyAvailable: true,
			tavilyHandler: func(int) (string, error) {
				return "", searchers.Retryable(errors.New("429"), 0)
			},
			wantResult:  "second-result",
			wantOrder:   []database.SearchengineType{EngineTavily, EngineTavily, EngineTraversaal},
			wantLogged:  EngineTraversaal,
			wantMinWait: 300 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var order []database.SearchengineType
			engines := map[database.SearchengineType]searchers.Searcher{
				EngineTavily: &webSearchEngineDouble{
					engine: EngineTavily, available: tt.tavilyAvailable, order: &order, handler: tt.tavilyHandler,
				},
				EngineTraversaal: &webSearchEngineDouble{
					engine: EngineTraversaal, available: true, order: &order, handler: webSearchAnswers("second-result"),
				},
			}
			slp := &webSearchLogRecorder{}
			ws := webSearchUnderTest(slp, engines)

			start := time.Now()
			got, err := ws.Handle(webSearchAgentCtx(), WebSearchToolName, webSearchArgs(t, "answer", "q"))
			elapsed := time.Since(start)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantResult {
				t.Errorf("result = %q, want %q", got, tt.wantResult)
			}
			if !reflect.DeepEqual(order, tt.wantOrder) {
				t.Errorf("call order = %v, want %v", order, tt.wantOrder)
			}
			if elapsed < tt.wantMinWait {
				t.Errorf("Handle took %v, want at least %v of backoff before the retry", elapsed, tt.wantMinWait)
			}
			if slp.calls != 1 || slp.engine != tt.wantLogged || slp.result != tt.wantResult {
				t.Errorf("search log: calls=%d engine=%q result=%q, want 1 row for %q with %q",
					slp.calls, slp.engine, slp.result, tt.wantLogged, tt.wantResult)
			}
		})
	}
}

func TestWebSearch_Handle_FailsHardOnlyOnArgumentsTheModelCanFix(t *testing.T) {
	tests := []struct {
		name          string
		engines       func() map[database.SearchengineType]searchers.Searcher
		args          []byte
		wantErr       string
		wantContains  []string
		wantLogs      int64
		wantLogEngine database.SearchengineType
		wantLogResult string
	}{
		{
			name:    "malformed arguments",
			args:    []byte("{not json"),
			wantErr: "failed to unmarshal web_search arguments",
		},
		{
			name:    "a blank query",
			args:    []byte(`{"query":"  ","message":"m"}`),
			wantErr: "'query' is required",
		},
		{
			name: "no engine configured for the mode",
			engines: func() map[database.SearchengineType]searchers.Searcher {
				return map[database.SearchengineType]searchers.Searcher{
					EngineTavily: &webSearchEngineDouble{engine: EngineTavily, available: false},
				}
			},
			args:         webSearchArgs(t, "answer", "q"),
			wantContains: []string{"no search engine is configured for mode 'answer'"},
		},
		{
			name: "every configured engine failed",
			engines: func() map[database.SearchengineType]searchers.Searcher {
				return map[database.SearchengineType]searchers.Searcher{
					EngineTavily:     &webSearchEngineDouble{engine: EngineTavily, available: true, handler: webSearchFails("down")},
					EngineTraversaal: &webSearchEngineDouble{engine: EngineTraversaal, available: true, handler: webSearchFails("down")},
				}
			},
			args: webSearchArgs(t, "answer", "q"),
			wantContains: []string{
				"all 2 configured engine(s) failed for mode 'answer'",
				"Last error (traversaal): down",
			},
			wantLogs:      1,
			wantLogEngine: EngineTraversaal,
			wantLogResult: "web_search failed: down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engines := map[database.SearchengineType]searchers.Searcher{}
			if tt.engines != nil {
				engines = tt.engines()
			}
			slp := &webSearchLogRecorder{}
			ws := webSearchUnderTest(slp, engines)

			got, err := ws.Handle(webSearchAgentCtx(), WebSearchToolName, tt.args)
			switch {
			case tt.wantErr != "":
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want a hard error containing %q", err, tt.wantErr)
				}
			case err != nil:
				t.Fatalf("hard error %v, want a message for the agent", err)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("message = %q, want it to contain %q", got, want)
				}
			}
			if slp.calls != tt.wantLogs {
				t.Errorf("search log calls = %d, want %d", slp.calls, tt.wantLogs)
			}
			if tt.wantLogs > 0 && (slp.engine != tt.wantLogEngine || slp.result != tt.wantLogResult) {
				t.Errorf("search log row: engine=%q result=%q, want %q with %q",
					slp.engine, slp.result, tt.wantLogEngine, tt.wantLogResult)
			}
		})
	}
}

// The limits are the ones the tool's schema promises the model: 1-25, 5 by default.
func TestWebSearch_Handle_BuildsTheEngineRequestFromTheArguments(t *testing.T) {
	tests := []struct {
		name string
		args string
		want searchers.Request
	}{
		{
			name: "an omitted max_results is five",
			args: `{"query":"q","message":"m"}`,
			want: searchers.Request{Query: "q", MaxResults: 5},
		},
		{
			name: "a zero max_results is five",
			args: `{"query":"q","max_results":0,"message":"m"}`,
			want: searchers.Request{Query: "q", MaxResults: 5},
		},
		{
			name: "a negative max_results is five",
			args: `{"query":"q","max_results":-5,"message":"m"}`,
			want: searchers.Request{Query: "q", MaxResults: 5},
		},
		{
			name: "a max_results of one is kept",
			args: `{"query":"q","max_results":1,"message":"m"}`,
			want: searchers.Request{Query: "q", MaxResults: 1},
		},
		{
			name: "a max_results at the ceiling is kept",
			args: `{"query":"q","max_results":25,"message":"m"}`,
			want: searchers.Request{Query: "q", MaxResults: 25},
		},
		{
			name: "a max_results above the ceiling is clamped",
			args: `{"query":"q","max_results":100,"message":"m"}`,
			want: searchers.Request{Query: "q", MaxResults: 25},
		},
		{
			name: "padding around the query, exploit type and sort is trimmed",
			args: `{"query":"  cve-2024-3094  ","mode":"exploit","exploit_type":" tools ","sort":" date ","message":"m"}`,
			want: searchers.Request{Query: "cve-2024-3094", MaxResults: 5, ExploitType: "tools", Sort: "date"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &webSearchEngineDouble{engine: EngineTavily, available: true, handler: webSearchAnswers("ok")}
			ws := webSearchUnderTest(&webSearchLogRecorder{}, map[database.SearchengineType]searchers.Searcher{
				EngineTavily: engine,
			})

			if _, err := ws.Handle(webSearchAgentCtx(), WebSearchToolName, []byte(tt.args)); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if want := []searchers.Request{tt.want}; !reflect.DeepEqual(engine.requests, want) {
				t.Errorf("engine requests = %+v, want %+v", engine.requests, want)
			}
		})
	}
}

// synctest's clock moves only while the bubble waits, so the wait is measured exactly.
func TestWebSearch_Handle_WaitsTheRetryAfterTheEngineAsksFor(t *testing.T) {
	tests := []struct {
		name       string
		retryAfter time.Duration
		wantWait   time.Duration
	}{
		{
			name:       "a Retry-After longer than the backoff lengthens the wait",
			retryAfter: 500 * time.Millisecond,
			wantWait:   500 * time.Millisecond,
		},
		{
			name:       "a Retry-After shorter than the backoff shortens it",
			retryAfter: time.Millisecond,
			wantWait:   time.Millisecond,
		},
		{
			name:     "without a Retry-After the first retry waits the backoff",
			wantWait: 300 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				engine := &webSearchEngineDouble{engine: EngineTavily, available: true, handler: func(call int) (string, error) {
					if call == 0 {
						return "", searchers.Retryable(errors.New("429"), tt.retryAfter)
					}
					return "recovered", nil
				}}
				ws := webSearchUnderTest(&webSearchLogRecorder{}, map[database.SearchengineType]searchers.Searcher{EngineTavily: engine})

				start := time.Now()
				got, err := ws.Handle(webSearchAgentCtx(), WebSearchToolName, webSearchArgs(t, "answer", "q"))
				waited := time.Since(start)

				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != "recovered" || engine.calls != 2 {
					t.Fatalf("result = %q after %d calls, want %q after 2", got, engine.calls, "recovered")
				}
				if waited != tt.wantWait {
					t.Errorf("Handle retried after %v, want exactly %v", waited, tt.wantWait)
				}
			})
		})
	}
}

// The engine's own request must also go through the flow's context, so a stop cuts it off too.
func TestWebSearch_Handle_StopsWaitingToRetryWhenTheContextEnds(t *testing.T) {
	ctx, cancel := context.WithCancel(webSearchAgentCtx())
	defer cancel()

	engine := &webSearchEngineDouble{engine: EngineTavily, available: true, handler: func(int) (string, error) {
		cancel()
		return "", searchers.Retryable(errors.New("429"), time.Hour)
	}}
	ws := webSearchUnderTest(&webSearchLogRecorder{}, map[database.SearchengineType]searchers.Searcher{EngineTavily: engine})

	type webSearchOutcome struct {
		result string
		err    error
	}
	done := make(chan webSearchOutcome, 1)
	go func() {
		result, err := ws.Handle(ctx, WebSearchToolName, webSearchArgs(t, "answer", "q"))
		done <- webSearchOutcome{result: result, err: err}
	}()

	select {
	case out := <-done:
		if out.err != nil {
			t.Fatalf("unexpected hard error: %v", out.err)
		}
		if !strings.Contains(out.result, context.Canceled.Error()) {
			t.Errorf("message = %q, want it to name %q", out.result, context.Canceled)
		}
		if engine.calls != 1 {
			t.Errorf("engine called %d times, want 1: the retry must not run once the context ended", engine.calls)
		}
		if len(engine.contexts) != 1 || !errors.Is(engine.contexts[0].Err(), context.Canceled) {
			t.Errorf("the engine was not asked through the flow's context: stopping the flow left its request running")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Handle is still waiting to retry after its context ended")
	}
}

// The query arrives padded: the row must hold the trimmed query the engines searched.
func TestWebSearch_Handle_AttributesTheSearchLogToTheAgentsInTheContext(t *testing.T) {
	taskID, subtaskID := int64(7), int64(8)
	primaryThenSearcher := PutAgentContext(
		PutAgentContext(context.Background(), database.MsgchainTypePrimaryAgent),
		database.MsgchainTypeSearcher,
	)

	tests := []struct {
		name string
		ctx  context.Context
		want *webSearchLogRecorder
	}{
		{
			name: "a searcher the primary agent called",
			ctx:  primaryThenSearcher,
			want: &webSearchLogRecorder{
				calls:      1,
				engine:     EngineTavily,
				query:      "q",
				result:     "found",
				taskID:     &taskID,
				subtaskID:  &subtaskID,
				parentType: database.MsgchainTypePrimaryAgent,
				currType:   database.MsgchainTypeSearcher,
			},
		},
		{
			name: "a context without an agent writes no row",
			ctx:  context.Background(),
			want: &webSearchLogRecorder{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slp := &webSearchLogRecorder{}
			ws := webSearchUnderTest(slp, map[database.SearchengineType]searchers.Searcher{
				EngineTavily: &webSearchEngineDouble{engine: EngineTavily, available: true, handler: webSearchAnswers("found")},
			})
			ws.taskID, ws.subtaskID = &taskID, &subtaskID

			got, err := ws.Handle(tt.ctx, WebSearchToolName, webSearchArgs(t, "answer", "  q  "))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != "found" {
				t.Errorf("result = %q, want %q", got, "found")
			}
			if got, want := webSearchLogRow(slp), webSearchLogRow(tt.want); got != want {
				t.Errorf("search log = %s\nwant         %s", got, want)
			}
		})
	}
}

func TestWebSearch_IsAvailable_NeedsOneConfiguredEngine(t *testing.T) {
	tests := []struct {
		name    string
		engines map[database.SearchengineType]searchers.Searcher
		want    bool
	}{
		{name: "no engine", engines: map[database.SearchengineType]searchers.Searcher{}, want: false},
		{
			name: "a nil engine and an unconfigured one",
			engines: map[database.SearchengineType]searchers.Searcher{
				EngineTavily: nil,
				EngineGoogle: &webSearchEngineDouble{engine: EngineGoogle, available: false},
			},
			want: false,
		},
		{
			name: "one configured engine among unconfigured ones",
			engines: map[database.SearchengineType]searchers.Searcher{
				EngineTavily: &webSearchEngineDouble{engine: EngineTavily, available: false},
				EngineGoogle: &webSearchEngineDouble{engine: EngineGoogle, available: true},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := webSearchUnderTest(&webSearchLogRecorder{}, tt.engines).IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWebSearch_NewBrowserPageFetcher_ReadsPagesOnlyThroughAConfiguredScraper(t *testing.T) {
	const page = "# A page\n\nlong enough for the browser to take it for a rendered page"

	t.Run("no scraper configured", func(t *testing.T) {
		if f := NewBrowserPageFetcher(1, nil, nil, &config.Config{DataDir: t.TempDir()}, nil); f != nil {
			t.Errorf("NewBrowserPageFetcher() = %T, want nil", f)
		}
	})

	tests := []struct {
		name           string
		markdownStatus int
		wantMD         string
		wantErrParts   []string
	}{
		{name: "a scraper that renders the page", markdownStatus: http.StatusOK, wantMD: page},
		{
			name:           "a failing scraper passes its status on",
			markdownStatus: http.StatusInternalServerError,
			wantErrParts:   []string{"failed to fetch content by url 'http://127.0.0.1:1/page'", ": 500"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := ""
			if tt.markdownStatus == http.StatusOK {
				body = page
			}
			scraper := fakeScraper(t, "http://127.0.0.1:1/page", map[string]scraperAnswer{
				"/markdown": {tt.markdownStatus, body},
			})

			cfg := &config.Config{DataDir: t.TempDir(), ScraperPrivateURL: scraper.URL}
			fetcher := NewBrowserPageFetcher(1, nil, nil, cfg, nil)
			if fetcher == nil {
				t.Fatal("NewBrowserPageFetcher() = nil with a scraper configured")
			}

			md, err := fetcher.FetchMarkdown(t.Context(), "http://127.0.0.1:1/page")
			switch {
			case len(tt.wantErrParts) == 0 && err != nil:
				t.Fatalf("FetchMarkdown() unexpected error: %v", err)
			case len(tt.wantErrParts) > 0 && err == nil:
				t.Fatalf("FetchMarkdown() error = nil, want one containing %q", tt.wantErrParts)
			}
			for _, part := range tt.wantErrParts {
				if !strings.Contains(err.Error(), part) {
					t.Errorf("FetchMarkdown() error = %q, want it to contain %q", err, part)
				}
			}
			if md != tt.wantMD {
				t.Errorf("FetchMarkdown() = %q, want %q", md, tt.wantMD)
			}
		})
	}
}
