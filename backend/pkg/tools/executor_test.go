package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"pentagi/pkg/cast"
	"pentagi/pkg/config"
	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"

	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
	"github.com/vxcontrol/langchaingo/llms"
)

// executorMsgLog stands in for the message log and, like the database, refuses a write through a done context.
type executorMsgLog struct {
	putErr    error
	updateErr error

	result string
	format database.MsglogResultFormat
}

func (l *executorMsgLog) PutMsg(
	ctx context.Context, _ database.MsglogType, _, _ *int64, _ int64, _, _ string,
) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if l.putErr != nil {
		return 0, l.putErr
	}
	return 1, nil
}

func (l *executorMsgLog) UpdateMsgResult(
	ctx context.Context, _ int64, _ int64, result string, format database.MsglogResultFormat,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	l.result, l.format = result, format
	return l.updateErr
}

// executorToolCallLog stands in for the tool call log and, like the database, refuses a write through a done context.
type executorToolCallLog struct {
	putErr     error
	successErr error

	success []string
	failed  []string
}

func (l *executorToolCallLog) PutLog(
	ctx context.Context, _, _ string, _ json.RawMessage, _, _ *int64,
) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if l.putErr != nil {
		return 0, l.putErr
	}
	return 1, nil
}

func (l *executorToolCallLog) UpdateLogSuccess(ctx context.Context, _ int64, result string, _ float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	l.success = append(l.success, result)
	return l.successErr
}

func (l *executorToolCallLog) UpdateLogFailed(ctx context.Context, _ int64, result string, _ float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	l.failed = append(l.failed, result)
	return nil
}

// executorLangfuse stands in for the Langfuse server and client behind a real observer, keeping every event sent.
type executorLangfuse struct {
	observer langfuse.Observer
	markers  int

	mu     sync.Mutex
	events []executorLangfuseEvent
}

type executorLangfuseEvent struct {
	Type string `json:"type"`
	Body struct {
		ID                  string         `json:"id"`
		TraceID             string         `json:"traceId"`
		ParentObservationID string         `json:"parentObservationId"`
		Name                string         `json:"name"`
		Metadata            map[string]any `json:"metadata"`
		Level               string         `json:"level"`
		StatusMessage       string         `json:"statusMessage"`
	} `json:"body"`
}

func (l *executorLangfuse) RoundTrip(r *http.Request) (*http.Response, error) {
	answer := `{"data":[{"id":"project","name":"executor"}]}`
	if r.Method == http.MethodPost {
		defer r.Body.Close()
		var ingestion struct {
			Batch []executorLangfuseEvent `json:"batch"`
		}
		if err := json.NewDecoder(r.Body).Decode(&ingestion); err != nil {
			return nil, err
		}
		l.mu.Lock()
		l.events = append(l.events, ingestion.Batch...)
		l.mu.Unlock()
		answer = `{"successes":[],"errors":[]}`
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       io.NopCloser(strings.NewReader(answer)),
		Request:    r,
	}, nil
}

func (l *executorLangfuse) API() langfuse.Client                 { return langfuse.Client{} }
func (l *executorLangfuse) Observer() langfuse.Observer          { return l.observer }
func (l *executorLangfuse) Shutdown(ctx context.Context) error   { return l.observer.Shutdown(ctx) }
func (l *executorLangfuse) ForceFlush(ctx context.Context) error { return l.observer.ForceFlush(ctx) }

// sent waits for a marker queued after the events, since a flush can go out before the last events reach it.
func (l *executorLangfuse) sent(t *testing.T) []executorLangfuseEvent {
	t.Helper()

	l.markers++
	marker := fmt.Sprintf("marker-%d", l.markers)
	_, observation := l.observer.NewObservation(context.Background(), langfuse.WithObservationTraceID(marker))
	observation.Event(langfuse.WithEventName(marker))

	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		if err := l.observer.ForceFlush(t.Context()); err != nil {
			t.Fatalf("sending the observations to Langfuse: %v", err)
		}
		l.mu.Lock()
		events := slices.Clone(l.events)
		l.mu.Unlock()
		if slices.ContainsFunc(events, func(e executorLangfuseEvent) bool { return e.Body.TraceID == marker }) {
			return events
		}
	}
	t.Fatal("the observer never sent the events queued before the marker")

	return nil
}

// executorObserveLangfuse installs a real Langfuse observer process-wide for the rest of the test.
func executorObserveLangfuse(t *testing.T) *executorLangfuse {
	t.Helper()

	lf := &executorLangfuse{}
	lfClient, err := langfuse.NewClient(
		langfuse.WithBaseURL("http://langfuse.invalid"),
		langfuse.WithPublicKey("pk"),
		langfuse.WithSecretKey("sk"),
		langfuse.WithProjectID("project"),
		langfuse.WithHTTPClient(&http.Client{Transport: lf}),
	)
	if err != nil {
		t.Fatalf("building the Langfuse client: %v", err)
	}
	// Nothing is sent on a timer or a full queue, only when the test asks.
	lf.observer = langfuse.NewObserver(lfClient,
		langfuse.WithSendInterval(10*time.Minute), langfuse.WithQueueSize(1000))

	obs.InitObserver(context.Background(), lf, nil, []logrus.Level{})
	t.Cleanup(func() {
		obs.InitObserver(context.Background(), nil, nil, []logrus.Level{})
		_ = lf.observer.Shutdown(context.Background())
	})

	return lf
}

func TestExecutor_Execute_AnswersACallItCannotRunWithACorrection(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		tool  string
		args  string
		check func(t *testing.T, answer string)
	}{
		{
			name: "an unknown tool is answered with the tools the executor has",
			tool: "reed_file",
			args: `{}`,
			check: func(t *testing.T, answer string) {
				if !strings.HasPrefix(answer, "function 'reed_file' not found") {
					t.Errorf("the answer must name the tool that was called, got %q", answer)
				}
				_, list, found := strings.Cut(answer, "available: ")
				if !found {
					t.Fatalf("the answer must list the tools the model may call, got %q", answer)
				}
				if list != "browser, file, terminal" {
					t.Errorf("available tools = %q, want every handler in a stable order", list)
				}
			},
		},
		{
			name: "the summarization marker is explained rather than refused",
			tool: cast.SummarizationToolName,
			args: `{}`,
			check: func(t *testing.T, answer string) {
				if !strings.Contains(answer, cast.SummarizationToolName) {
					t.Errorf("the answer must name the marker, got %q", answer)
				}
				if !strings.Contains(answer, "automatically") {
					t.Errorf("the answer must say the compaction happens on its own, got %q", answer)
				}
				if strings.Contains(answer, "not found in available tools list") {
					t.Errorf("the marker must not be reported as a missing tool, got %q", answer)
				}
			},
		},
		{
			name: "malformed arguments are sent back to be fixed",
			tool: TerminalToolName,
			args: `{invalid`,
			check: func(t *testing.T, answer string) {
				if !strings.Contains(answer, "failed to unmarshal 'terminal' tool call arguments") ||
					!strings.HasSuffix(answer, "fix it") {
					t.Errorf("the answer must ask for the arguments to be fixed, got %q", answer)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var runs atomic.Int32
			handler := func(context.Context, string, json.RawMessage) (string, error) {
				runs.Add(1)
				return "ran", nil
			}
			ce := &customExecutor{handlers: map[string]ExecutorHandler{
				TerminalToolName: handler,
				FileToolName:     handler,
				BrowserToolName:  handler,
			}}

			answer, err := ce.Execute(t.Context(), 1, "call-1", tc.tool, "", "", json.RawMessage(tc.args))
			if err != nil {
				t.Fatalf("the model has to be answered, yet Execute failed: %v", err)
			}
			if runs.Load() != 0 {
				t.Errorf("a call the executor cannot run reached a handler")
			}
			tc.check(t, answer)
		})
	}
}

func TestExecutor_Execute_ReportsALoggingFailureOnlyBeforeTheToolRuns(t *testing.T) {
	obs.InitObserver(context.Background(), nil, nil, []logrus.Level{})

	errWrite := errors.New("connection reset by peer")
	args := mustJSON(map[string]any{
		"input": "id", "cwd": "", "detach": false, "timeout": 0,
		"message": "Check the user.",
	})

	for name, tc := range map[string]struct {
		mlp      *executorMsgLog
		tclp     *executorToolCallLog
		embedder *fakeEmbedder
		wantRun  bool
	}{
		"the message log refuses the call": {
			mlp: &executorMsgLog{putErr: errWrite}, tclp: &executorToolCallLog{},
		},
		"the tool call log refuses the call": {
			mlp: &executorMsgLog{}, tclp: &executorToolCallLog{putErr: errWrite},
		},
		"the embedder rejects the memory write": {
			mlp: &executorMsgLog{}, tclp: &executorToolCallLog{}, wantRun: true,
			embedder: &fakeEmbedder{err: errors.New("401 Unauthorized: invalid api key")},
		},
		"the tool call log update fails": {
			mlp: &executorMsgLog{}, tclp: &executorToolCallLog{successErr: errWrite}, wantRun: true,
		},
		"the message result update fails": {
			mlp: &executorMsgLog{updateErr: errWrite}, tclp: &executorToolCallLog{}, wantRun: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			var runs atomic.Int32
			ce := &customExecutor{
				flowID:   1,
				mlp:      tc.mlp,
				tclp:     tc.tclp,
				replacer: identityReplacer{},
				handlers: map[string]ExecutorHandler{
					TerminalToolName: func(context.Context, string, json.RawMessage) (string, error) {
						runs.Add(1)
						return "uid=0(root)", nil
					},
				},
			}
			if tc.embedder != nil {
				ce.store = newFakeVectorStore(t, &fakeVectorConn{}, tc.embedder)
			}

			result, err := ce.Execute(t.Context(), 1, "call-1", TerminalToolName, TerminalToolName, "", args)

			if !tc.wantRun {
				if !errors.Is(err, errWrite) || runs.Load() != 0 {
					t.Fatalf("a call that could not be logged must be reported failed without running, "+
						"got result %q and error %v after %d runs", result, err, runs.Load())
				}
				return
			}
			if err != nil {
				t.Fatalf("the command ran, yet Execute reported a failure the performer answers by running it again: %v", err)
			}
			if result != "uid=0(root)" || runs.Load() != 1 {
				t.Fatalf("result %q after %d runs", result, runs.Load())
			}
		})
	}
}

func TestExecutor_Execute_ClosesTheToolCallLogEvenAfterTheFlowStops(t *testing.T) {
	errHandler := errors.New("docker daemon is not reachable")
	withMessage := mustJSON(map[string]any{"input": "id", "message": "Check the user."})

	for _, tc := range []struct {
		name        string
		args        json.RawMessage
		stopFlow    bool
		handlerErr  error
		wantSuccess []string
		wantFailed  []string
	}{
		{
			name: "a failed call that carries a message", args: withMessage, handlerErr: errHandler,
			wantFailed: []string{"failed to execute handler: docker daemon is not reachable"},
		},
		{
			name: "a failed call without one", args: mustJSON(map[string]any{"input": "id"}), handlerErr: errHandler,
			wantFailed: []string{"failed to execute handler: docker daemon is not reachable"},
		},
		{
			name: "a call that fails once its flow is stopped", args: withMessage, stopFlow: true, handlerErr: errHandler,
			wantFailed: []string{"failed to execute handler: docker daemon is not reachable"},
		},
		{
			name: "a result that arrives once its flow is stopped", args: withMessage, stopFlow: true,
			wantSuccess: []string{"uid=0(root)"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, stop := context.WithCancel(t.Context())
			defer stop()

			tclp := &executorToolCallLog{}
			ce := &customExecutor{
				flowID:   1,
				mlp:      &executorMsgLog{},
				tclp:     tclp,
				replacer: identityReplacer{},
				handlers: map[string]ExecutorHandler{
					TerminalToolName: func(context.Context, string, json.RawMessage) (string, error) {
						if tc.stopFlow {
							stop()
						}
						if tc.handlerErr != nil {
							return "", tc.handlerErr
						}
						return "uid=0(root)", nil
					},
				},
			}

			result, err := ce.Execute(ctx, 1, "call-1", TerminalToolName, TerminalToolName, "", tc.args)
			if tc.handlerErr != nil {
				if !errors.Is(err, tc.handlerErr) || result != "" {
					t.Fatalf("the handler's failure must reach the caller, got result %q and error %v", result, err)
				}
			} else if err != nil || result != "uid=0(root)" {
				t.Fatalf("the command ran, got result %q and error %v", result, err)
			}
			if !slices.Equal(tclp.failed, tc.wantFailed) || !slices.Equal(tclp.success, tc.wantSuccess) {
				t.Fatalf("the tool call log was closed with failed %q and success %q, want failed %q and success %q",
					tclp.failed, tclp.success, tc.wantFailed, tc.wantSuccess)
			}
		})
	}
}

func TestExecutor_Execute_CleansTheHandlersOutputOfBytesPostgresRejects(t *testing.T) {
	tclp, mlp := &executorToolCallLog{}, &executorMsgLog{}
	ce := &customExecutor{
		mlp:      mlp,
		tclp:     tclp,
		replacer: identityReplacer{},
		handlers: map[string]ExecutorHandler{
			TerminalToolName: func(context.Context, string, json.RawMessage) (string, error) {
				return "uid=0\xff(root)\x00", nil
			},
		},
	}

	args := mustJSON(map[string]any{"input": "id", "message": "Check the user."})
	result, err := ce.Execute(t.Context(), 1, "call-1", TerminalToolName, TerminalToolName, "", args)
	if err != nil {
		t.Fatalf("the command ran, yet Execute reported a failure: %v", err)
	}

	const want = "uid=0�(root)"
	if result != want {
		t.Errorf("the agent received %q, want %q", result, want)
	}
	if !slices.Equal(tclp.success, []string{want}) || mlp.result != want {
		t.Errorf("the logs hold %q and %q, want %q in both", tclp.success, mlp.result, want)
	}
}

// Both logs must hold exactly what the agent receives: the summary, or the page cut down.
func TestExecutor_Execute_SummarizesALargeResultOrTruncatesIt(t *testing.T) {
	page := strings.Repeat("x", DefaultResultSizeLimit*3)
	longURL := "https://example.com/?q=" + strings.Repeat("u", maxArgValueLength)
	args := mustJSON(map[string]any{"url": longURL, "action": "markdown", "message": "Read the page."})

	truncated := func(t *testing.T, result string) {
		if !strings.Contains(result, "[truncated]") || len(result) >= len(page) {
			t.Fatalf("the unsummarised result must be cut down, got %d bytes of a %d-byte page", len(result), len(page))
		}
	}

	for name, tc := range map[string]struct {
		summarize  bool
		summary    string
		summaryErr error
		wantFormat database.MsglogResultFormat
		check      func(t *testing.T, result string)
	}{
		"the summary arrives": {
			summarize:  true,
			summary:    "The page lists three hosts.",
			wantFormat: database.MsglogResultFormatMarkdown,
			check: func(t *testing.T, result string) {
				if result != "The page lists three hosts." {
					t.Fatalf("the agent must receive the summary, got %d bytes", len(result))
				}
			},
		},
		"the summarizer fails": {
			summarize:  true,
			summaryErr: errors.New("502 bad gateway"),
			wantFormat: database.MsglogResultFormatPlain,
			check:      truncated,
		},
		"no summarizer is configured": {
			wantFormat: database.MsglogResultFormatPlain,
			check:      truncated,
		},
	} {
		t.Run(name, func(t *testing.T) {
			var prompt string
			tclp, mlp := &executorToolCallLog{}, &executorMsgLog{}
			ce := &customExecutor{
				mlp:      mlp,
				tclp:     tclp,
				replacer: identityReplacer{},
				handlers: map[string]ExecutorHandler{
					BrowserToolName: func(context.Context, string, json.RawMessage) (string, error) {
						return page, nil
					},
				},
			}
			if tc.summarize {
				ce.summarizer = func(_ context.Context, p string) (string, error) {
					prompt = p
					return tc.summary, tc.summaryErr
				}
			}

			result, err := ce.Execute(t.Context(), 1, "call-1", BrowserToolName, BrowserToolName, "", args)
			if err != nil {
				t.Fatalf("the page was fetched, yet Execute reported a failure: %v", err)
			}
			tc.check(t, result)

			if !slices.Equal(tclp.success, []string{result}) {
				t.Fatal("the tool call log must hold what the agent received")
			}
			if mlp.result != result || mlp.format != tc.wantFormat {
				t.Fatalf("the message log holds a %q result of %d bytes, want the %q result the agent received",
					mlp.format, len(mlp.result), tc.wantFormat)
			}
			if !tc.summarize {
				return
			}

			for _, want := range []string{
				`<function name="browser">`,
				"url: " + longURL[:1024] + "... [truncated]\n",
				`"url": {`,
				page,
			} {
				if !strings.Contains(prompt, want) {
					t.Errorf("the summarizer's prompt lacks %.80q", want)
				}
			}
			if strings.Contains(prompt, longURL) {
				t.Error("an argument value longer than the limit reached the summarizer whole")
			}
		})
	}
}

func TestExecutor_Execute_StopsSummarizingWhenTheFlowStops(t *testing.T) {
	entered := make(chan struct{})
	tclp := &executorToolCallLog{}

	ce := &customExecutor{
		tclp: tclp,
		handlers: map[string]ExecutorHandler{
			BrowserToolName: func(context.Context, string, json.RawMessage) (string, error) {
				return strings.Repeat("x", DefaultResultSizeLimit+1), nil
			},
		},
		summarizer: func(ctx context.Context, _ string) (string, error) {
			close(entered)
			<-ctx.Done()

			return "", ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := ce.Execute(ctx, 1, "id", BrowserToolName, "", "", json.RawMessage(`{"action":"markdown"}`))
		done <- err
	}()

	select {
	case <-entered:
	case err := <-done:
		t.Fatalf("the result was never handed to the summarizer, Execute returned %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("the result was never handed to the summarizer")
	}
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("a cancelled summary must report cancellation, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("summarising kept running after the flow was stopped")
	}

	if len(tclp.failed) != 1 || !strings.Contains(tclp.failed[0], "summarize") {
		t.Fatalf("the tool call log must still record why the result never arrived, got: %q", tclp.failed)
	}
}

func TestExecutor_Execute_AnonymizesWhatEveryAgentStoresInMemory(t *testing.T) {
	obs.InitObserver(context.Background(), nil, nil, []logrus.Level{})

	handler := func(context.Context, string, json.RawMessage) (string, error) { return "done", nil }

	cases := map[string]struct {
		agent database.MsgchainType
		build func(fte *flowToolsExecutor) (ContextToolsExecutor, error)
	}{
		"assistant": {
			agent: database.MsgchainTypeAssistant,
			build: func(fte *flowToolsExecutor) (ContextToolsExecutor, error) {
				return fte.GetAssistantExecutor(AssistantExecutorConfig{
					Adviser: handler, Coder: handler, Installer: handler,
					Memorist: handler, Pentester: handler, Searcher: handler,
				})
			},
		},
		"installer": {
			agent: database.MsgchainTypeInstaller,
			build: func(fte *flowToolsExecutor) (ContextToolsExecutor, error) {
				return fte.GetInstallerExecutor(InstallerExecutorConfig{
					Adviser: handler, Memorist: handler, Searcher: handler, MaintenanceResult: handler,
				})
			},
		},
		"coder": {
			agent: database.MsgchainTypeCoder,
			build: func(fte *flowToolsExecutor) (ContextToolsExecutor, error) {
				return fte.GetCoderExecutor(CoderExecutorConfig{
					Adviser: handler, Installer: handler, Memorist: handler, Searcher: handler, CodeResult: handler,
				})
			},
		},
		"pentester": {
			agent: database.MsgchainTypePentester,
			build: func(fte *flowToolsExecutor) (ContextToolsExecutor, error) {
				return fte.GetPentesterExecutor(PentesterExecutorConfig{
					Adviser: handler, Coder: handler, Installer: handler,
					Memorist: handler, Searcher: handler, HackResult: handler,
				})
			},
		},
		"generator": {
			agent: database.MsgchainTypeGenerator,
			build: func(fte *flowToolsExecutor) (ContextToolsExecutor, error) {
				return fte.GetGeneratorExecutor(GeneratorExecutorConfig{
					TaskID: 1, Memorist: handler, Searcher: handler, SubtaskList: handler,
				})
			},
		},
		"refiner": {
			agent: database.MsgchainTypeRefiner,
			build: func(fte *flowToolsExecutor) (ContextToolsExecutor, error) {
				return fte.GetRefinerExecutor(RefinerExecutorConfig{
					TaskID: 1, Memorist: handler, Searcher: handler, SubtaskPatch: handler,
				})
			},
		},
		"memorist": {
			agent: database.MsgchainTypeMemorist,
			build: func(fte *flowToolsExecutor) (ContextToolsExecutor, error) {
				return fte.GetMemoristExecutor(MemoristExecutorConfig{SearchResult: handler})
			},
		},
		"enricher": {
			agent: database.MsgchainTypeEnricher,
			build: func(fte *flowToolsExecutor) (ContextToolsExecutor, error) {
				return fte.GetEnricherExecutor(EnricherExecutorConfig{EnricherResult: handler})
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			conn := &fakeVectorConn{}
			store := newFakeVectorStore(t, conn, &fakeEmbedder{})

			log := &recordingVectorStoreLog{}
			fte := &flowToolsExecutor{
				flowID: 1,
				cfg:    &config.Config{},
				db:     &fakeContainerDB{row: &database.Container{ID: 1, LocalID: sql.NullString{String: "primary", Valid: true}}},
				docker: &fakeDockerClient{
					isRunning:      true,
					execCreateResp: client.ExecCreateResult{ID: "exec"},
					attachOutput:   []byte("inet " + secretHost + "/24 scope global eth0"),
				},
				tlp:      &recordingTermLog{},
				mlp:      &executorMsgLog{},
				tclp:     &executorToolCallLog{},
				vslp:     log,
				store:    store,
				replacer: hostReplacer{},
			}

			executor, err := tc.build(fte)
			if err != nil {
				t.Fatalf("building the executor: %v", err)
			}

			args := mustJSON(map[string]any{
				"input": "ip addr | grep " + secretHost, "cwd": "", "detach": false, "timeout": 0,
				"message": "Check which interface carries the address.",
			})
			ctx := PutAgentContext(t.Context(), tc.agent)

			if _, err := executor.Execute(ctx, 1, "call-1", TerminalToolName, TerminalToolName, "", args); err != nil {
				t.Fatalf("the terminal call failed: %v", err)
			}

			if len(conn.added) == 0 || len(log.entries) == 0 || log.entries[len(log.entries)-1].result == "" {
				t.Fatal("the terminal result never reached the vector store, so this case proves nothing")
			}

			for _, doc := range conn.added {
				if written := fmt.Sprint(doc.args...); strings.Contains(written, secretHost) {
					t.Errorf("the vector store received %s verbatim: %q", secretHost, written)
				}
			}
			for _, entry := range log.entries {
				if strings.Contains(entry.query, secretHost) || strings.Contains(entry.result, secretHost) {
					t.Errorf("the vector store log carries %s verbatim: query %q, result %q", secretHost, entry.query, entry.result)
				}
			}
		})
	}
}

func TestExecutor_Execute_StoresAMemoryToolResultUnderItsScope(t *testing.T) {
	taskID, subtaskID := int64(2), int64(3)

	for _, tc := range []struct {
		name       string
		tool       string
		message    string
		wantStored bool
	}{
		{
			name: "a terminal result is stored under its flow, task and subtask",
			tool: TerminalToolName, message: "Check the user.", wantStored: true,
		},
		{
			name: "a browser result is not a memory tool's and is not stored",
			tool: BrowserToolName, message: "Check the user.",
		},
		{
			name: "a terminal call without a message is not stored",
			tool: TerminalToolName,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conn := &fakeVectorConn{}
			store := newFakeVectorStore(t, conn, &fakeEmbedder{})
			vslp := &recordingVectorStoreLog{}
			ce := &customExecutor{
				userID:    7,
				flowID:    1,
				taskID:    &taskID,
				subtaskID: &subtaskID,
				mlp:       &executorMsgLog{},
				tclp:      &executorToolCallLog{},
				replacer:  identityReplacer{},
				store:     store,
				vslp:      vslp,
				handlers: map[string]ExecutorHandler{
					tc.tool: func(context.Context, string, json.RawMessage) (string, error) { return "uid=0(root)", nil },
				},
			}

			fields := map[string]any{"input": "id"}
			if tc.message != "" {
				fields["message"] = tc.message
			}
			ctx := PutAgentContext(t.Context(), database.MsgchainTypePentester)
			if _, err := ce.Execute(ctx, 1, "call-1", tc.tool, tc.tool, "", mustJSON(fields)); err != nil {
				t.Fatalf("the call failed: %v", err)
			}

			if !tc.wantStored {
				if len(conn.added) != 0 || len(vslp.entries) != 0 {
					t.Fatalf("the %s result reached memory: %d writes, %d log entries", tc.tool, len(conn.added), len(vslp.entries))
				}
				return
			}

			if len(conn.added) == 0 || len(vslp.entries) != 1 {
				t.Fatal("the result never reached the vector store, so this case proves nothing")
			}
			for _, doc := range conn.added {
				for key, want := range map[string]any{
					"user_id": int64(7), "flow_id": int64(1), "task_id": int64(2), "subtask_id": int64(3),
					"tool_name": "terminal", "doc_type": "memory",
				} {
					if doc.meta[key] != want {
						t.Errorf("a stored part carries %s = %v, want %v", key, doc.meta[key], want)
					}
				}
			}

			var filters map[string]any
			if err := json.Unmarshal([]byte(vslp.entries[0].filter), &filters); err != nil {
				t.Fatalf("the vector store log was given filters that are not JSON: %v", err)
			}
			for key, want := range map[string]any{
				"flow_id": 1.0, "task_id": 2.0, "subtask_id": 3.0, "tool_name": "terminal", "doc_type": "memory",
			} {
				if filters[key] != want {
					t.Errorf("the vector store log shows %s = %v, want %v", key, filters[key], want)
				}
			}
		})
	}
}

// Each observation Execute opens is filed by its kind, carries flow, task and subtask, and marks failure by level.
func TestExecutor_Execute_ObservesEachToolTypeUnderTheCallersTrace(t *testing.T) {
	lf := executorObserveLangfuse(t)

	taskID, subtaskID := int64(2), int64(3)
	errHandler := errors.New("docker daemon is not reachable")
	failing := json.RawMessage(`{"fail":true}`)

	for _, tc := range []struct {
		name    string
		tool    string
		opens   string
		nameKey string
	}{
		{name: "an environment tool is a tool", tool: TerminalToolName, opens: "tool-create", nameKey: "tool_name"},
		{name: "a network search is a tool", tool: BrowserToolName, opens: "tool-create", nameKey: "tool_name"},
		{name: "an agent's result is a tool", tool: HackResultToolName, opens: "tool-create", nameKey: "tool_name"},
		{name: "a vector-store write is a tool", tool: StoreGuideToolName, opens: "tool-create", nameKey: "tool_name"},
		{name: "an agent is an agent", tool: CoderToolName, opens: "agent-create", nameKey: "agent_name"},
		{name: "a barrier is a span", tool: FinalyToolName, opens: "span-create", nameKey: "barrier_name"},
		{name: "a vector-store search opens nothing", tool: SearchInMemoryToolName},
		{name: "a tool with no type opens nothing", tool: "custom_tool"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			traceID, callerID := "trace-of-"+tc.tool, "caller-of-"+tc.tool
			caller, _ := obs.Observer.NewObservation(t.Context(),
				langfuse.WithObservationTraceID(traceID), langfuse.WithObservationID(callerID))
			caller = PutAgentContext(caller, database.MsgchainTypePentester)

			var ranIn []langfuse.Observation
			ce := &customExecutor{
				flowID:    1,
				taskID:    &taskID,
				subtaskID: &subtaskID,
				tclp:      &executorToolCallLog{},
				handlers: map[string]ExecutorHandler{
					tc.tool: func(ctx context.Context, _ string, args json.RawMessage) (string, error) {
						_, observation := obs.Observer.NewObservation(ctx)
						ranIn = append(ranIn, observation)
						if agent, ok := GetAgentContext(ctx); !ok || agent.CurrentAgentType != database.MsgchainTypePentester {
							t.Errorf("the handler lost the calling agent, got %+v", agent)
						}
						if string(args) == string(failing) {
							return "", errHandler
						}
						return "handled", nil
					},
				},
			}

			result, err := ce.Execute(caller, 1, "call-1", tc.tool, tc.tool, "", json.RawMessage(`{}`))
			if err != nil || result != "handled" {
				t.Fatalf("Execute must return what the handler did, got %q and %v", result, err)
			}
			if _, err := ce.Execute(caller, 1, "call-2", tc.tool, tc.tool, "", failing); !errors.Is(err, errHandler) {
				t.Fatalf("Execute must return the handler's failure, got %v", err)
			}

			sent := lf.sent(t)
			var opened []executorLangfuseEvent
			for _, event := range sent {
				if event.Body.TraceID == traceID {
					opened = append(opened, event)
				}
			}

			for _, observation := range ranIn {
				if observation.TraceID() != traceID {
					t.Errorf("the handler ran in trace %q, want the caller's %q", observation.TraceID(), traceID)
				}
			}
			if tc.opens == "" {
				if len(opened) != 0 {
					t.Errorf("Execute sent Langfuse %d observations, want none: %+v", len(opened), opened)
				}
				for _, observation := range ranIn {
					if observation.ID() != callerID {
						t.Errorf("the handler ran in observation %q, want the caller's %q with nothing around it",
							observation.ID(), callerID)
					}
				}
				return
			}

			if len(opened) != 2 || len(ranIn) != 2 {
				t.Fatalf("two calls opened %d observations and ran %d handlers, want one of each per call",
					len(opened), len(ranIn))
			}
			for i, want := range []struct{ level, status string }{
				{level: "DEFAULT", status: "success"},
				{level: "ERROR", status: "failed to execute handler: docker daemon is not reachable"},
			} {
				id := ranIn[i].ID()
				idx := slices.IndexFunc(opened, func(e executorLangfuseEvent) bool { return e.Body.ID == id })
				if idx < 0 {
					t.Errorf("call %d: the handler ran in observation %q, which Execute never sent to Langfuse", i+1, id)
					continue
				}
				opening := opened[idx]
				if opening.Type != tc.opens || opening.Body.Name != tc.tool || opening.Body.ParentObservationID != callerID {
					t.Errorf("call %d: Execute opened %s %q under %q, want %s %q under the caller's %q", i+1,
						opening.Type, opening.Body.Name, opening.Body.ParentObservationID, tc.opens, tc.tool, callerID)
				}
				for key, value := range map[string]any{
					tc.nameKey: tc.tool, "flow_id": 1.0, "task_id": 2.0, "subtask_id": 3.0,
				} {
					if opening.Body.Metadata[key] != value {
						t.Errorf("call %d: the observation carries %s = %v, want %v", i+1, key, opening.Body.Metadata[key], value)
					}
				}

				// One goroutine sends events in the order Execute made them, so an observation's second event is its close.
				var lifecycle []executorLangfuseEvent
				for _, event := range sent {
					if event.Body.ID == id {
						lifecycle = append(lifecycle, event)
					}
				}
				if len(lifecycle) != 2 {
					t.Errorf("call %d: observation %q was sent %d times, want once opened and once closed",
						i+1, id, len(lifecycle))
					continue
				}
				if closing := lifecycle[1].Body; closing.Level != want.level || closing.StatusMessage != want.status {
					t.Errorf("call %d: the observation was closed at level %q with %q, want %q with %q",
						i+1, closing.Level, closing.StatusMessage, want.level, want.status)
				}
			}
		})
	}
}

func TestExecutor_IsFunctionExists_AnswersFromTheHandlers(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{
		handlers: map[string]ExecutorHandler{TerminalToolName: nil, FinalyToolName: nil},
		barriers: map[string]struct{}{FinalyToolName: {}, AskUserToolName: {}},
	}

	for name, want := range map[string]bool{
		TerminalToolName: true, FinalyToolName: true, AskUserToolName: false, AdviceToolName: false,
	} {
		if got := ce.IsFunctionExists(name); got != want {
			t.Errorf("IsFunctionExists(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestExecutor_IsBarrierFunction_RecognisesOnlyTheConfiguredBarriers(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{
		handlers: map[string]ExecutorHandler{TerminalToolName: nil, FinalyToolName: nil},
		barriers: map[string]struct{}{FinalyToolName: {}, AskUserToolName: {}},
	}

	for name, want := range map[string]bool{
		FinalyToolName: true, AskUserToolName: true, TerminalToolName: false, "": false,
	} {
		if got := ce.IsBarrierFunction(name); got != want {
			t.Errorf("IsBarrierFunction(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestExecutor_GetBarrierToolNames_ListsEveryBarrier(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{barriers: map[string]struct{}{FinalyToolName: {}, AskUserToolName: {}}}

	names := ce.GetBarrierToolNames()
	slices.Sort(names)
	if want := []string{"ask", "done"}; !slices.Equal(names, want) {
		t.Errorf("GetBarrierToolNames() = %q, want %q", names, want)
	}
}

func TestExecutor_GetBarrierTools_DescribesEachBarrierByItsArguments(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		barriers []string
		want     map[string]string
	}{
		{
			name:     "every barrier is described with the schema of its arguments",
			barriers: []string{FinalyToolName, AskUserToolName},
			want:     map[string]string{"done": "result", "ask": "message"},
		},
		{
			name:     "a barrier with no definition is left out",
			barriers: []string{FinalyToolName, "unknown_tool"},
			want:     map[string]string{"done": "result"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ce := &customExecutor{barriers: map[string]struct{}{}}
			for _, name := range tc.barriers {
				ce.barriers[name] = struct{}{}
			}

			tools := ce.GetBarrierTools()
			if len(tools) != len(tc.want) {
				t.Fatalf("GetBarrierTools() described %d barriers, want %d: %+v", len(tools), len(tc.want), tools)
			}
			for _, tool := range tools {
				argument, ok := tc.want[tool.Name]
				if !ok {
					t.Errorf("GetBarrierTools() described %q, want only %v", tool.Name, tc.want)
					continue
				}
				var described struct {
					Properties map[string]json.RawMessage `json:"properties"`
				}
				if err := json.Unmarshal([]byte(tool.Schema), &described); err != nil {
					t.Fatalf("the schema of %q is not JSON: %v", tool.Name, err)
				}
				if _, ok := described.Properties[argument]; !ok {
					t.Errorf("the schema of %q lacks its %q argument: %s", tool.Name, argument, tool.Schema)
				}
			}
		})
	}
}

func TestExecutor_Tools_OffersEveryDefinitionAsAFunctionInOrder(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{definitions: []llms.FunctionDefinition{
		{Name: TerminalToolName, Description: "terminal"},
		{Name: FileToolName, Description: "file"},
	}}

	var names []string
	for _, tool := range ce.Tools() {
		if tool.Type != "function" || tool.Function == nil {
			t.Fatalf("a tool the model is offered must be a function, got %+v", tool)
		}
		names = append(names, tool.Function.Name)
	}
	if want := []string{"terminal", "file"}; !slices.Equal(names, want) {
		t.Errorf("Tools() = %q, want %q", names, want)
	}
}

func TestExecutor_GetToolSchema_PrefersTheExecutorsDefinitionToTheRegistry(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{definitions: []llms.FunctionDefinition{
		{Name: TerminalToolName, Parameters: map[string]any{
			"type":       "object",
			"properties": map[string]any{"only_here": map[string]any{"type": "string"}},
		}},
		{Name: "unmarshalable_tool", Parameters: make(chan int)},
		{Name: "not_a_schema_tool", Parameters: map[string]any{"type": 5}},
	}}

	for _, tc := range []struct {
		name         string
		tool         string
		wantProperty string
		wantAbsent   string
		wantErr      string
	}{
		{
			name: "the executor's own definition wins over the registry's",
			tool: TerminalToolName, wantProperty: "only_here", wantAbsent: "input",
		},
		{
			name: "a tool the executor does not define falls back to the registry",
			tool: BrowserToolName, wantProperty: "url",
		},
		{
			name: "an unknown tool is an error",
			tool: "unknown_tool", wantErr: "tool unknown_tool not found",
		},
		{
			name: "parameters that do not marshal are an error",
			tool: "unmarshalable_tool", wantErr: "failed to marshal parameters",
		},
		{
			name: "parameters that are not a JSON schema are an error",
			tool: "not_a_schema_tool", wantErr: "failed to unmarshal schema",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ce.GetToolSchema(tc.tool)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("GetToolSchema(%q) error = %v, want %q", tc.tool, err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetToolSchema(%q) unexpected error: %v", tc.tool, err)
			}
			if _, ok := got.Properties[tc.wantProperty]; !ok {
				t.Errorf("GetToolSchema(%q) lacks %q", tc.tool, tc.wantProperty)
			}
			if _, ok := got.Properties[tc.wantAbsent]; tc.wantAbsent != "" && ok {
				t.Errorf("GetToolSchema(%q) has the registry's %q", tc.tool, tc.wantAbsent)
			}
		})
	}
}

func TestExecutor_TruncateResult_CutsOnRuneBoundaries(t *testing.T) {
	t.Parallel()

	limit := DefaultResultSizeLimit
	for _, tc := range []struct {
		name  string
		input string
		want  []string
	}{
		{
			// "ж" is two bytes, so after the leading "a" the head cut at the even limit lands inside one.
			name:  "a cut through a character at the head",
			input: "a" + strings.Repeat("ж", limit),
			want:  []string{"[0:16383 bytes]", "[16385:32769 bytes]"},
		},
		{
			name:  "a cut through a character at the tail",
			input: strings.Repeat("ж", limit) + "a",
			want:  []string{"[0:16384 bytes]", "[16386:32769 bytes]"},
		},
		{
			name:  "a result of twice the limit is left whole",
			input: strings.Repeat("x", 2*limit),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := truncateResult(tc.input)
			if !utf8.ValidString(got) {
				t.Fatal("a cut through a multibyte character is rejected by postgres when the tool call is stored")
			}
			if tc.want == nil {
				if got != tc.input {
					t.Fatalf("a result within the limit was cut to %d bytes", len(got))
				}
				return
			}
			if !strings.Contains(got, "[truncated]") {
				t.Fatal("the result was not truncated")
			}
			for _, mark := range tc.want {
				if !strings.Contains(got, mark) {
					t.Errorf("the truncated result lacks the mark %q", mark)
				}
			}
		})
	}
}

func TestExecutor_GetMessage_ReadsTheMessageFieldOrNothing(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{}

	for _, tt := range []struct {
		name string
		args string
		want string
	}{
		{name: "a message field", args: `{"message": "hello world", "other": "data"}`, want: "hello world"},
		{name: "an empty message", args: `{"message": ""}`, want: ""},
		{name: "no message field", args: `{"other": "data"}`, want: ""},
		{name: "invalid json", args: `{invalid}`, want: ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ce.getMessage(json.RawMessage(tt.args)); got != tt.want {
				t.Errorf("getMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExecutor_ArgsToMarkdown_ListsEveryArgumentButTheMessage(t *testing.T) {
	t.Parallel()

	ce := &customExecutor{}

	for _, tt := range []struct {
		name    string
		args    string
		want    string
		wantErr bool
	}{
		{name: "an argument becomes a bullet", args: `{"query": "test search"}`, want: "* query: test search\n"},
		{name: "the message is left out", args: `{"query": "test", "message": "should be skipped"}`, want: "* query: test\n"},
		{name: "no arguments", args: `{}`, want: ""},
		{name: "invalid json", args: `{invalid}`, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ce.argsToMarkdown(json.RawMessage(tt.args))
			if (err != nil) != tt.wantErr {
				t.Fatalf("argsToMarkdown() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("argsToMarkdown() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Not parallel: captureLogrus swaps the standard logger's hooks.
func TestExecutor_Execute_LogsAMissingToolButNotTheSummarizationMarker(t *testing.T) {
	for _, tc := range []struct {
		name    string
		tool    string
		wantLog bool
	}{
		{name: "a tool that does not exist is worth an operator's attention", tool: "reed_file", wantLog: true},
		{name: "the marker the transcript itself taught the model is not", tool: cast.SummarizationToolName},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hook := captureLogrus(t)
			ce := &customExecutor{
				flowID: 1, mlp: &executorMsgLog{}, tclp: &executorToolCallLog{},
				replacer: identityReplacer{},
				handlers: map[string]ExecutorHandler{TerminalToolName: func(
					context.Context, string, json.RawMessage,
				) (string, error) {
					return "", nil
				}},
			}

			if _, err := ce.Execute(t.Context(), 1, "call-1", tc.tool, tc.tool, "", json.RawMessage(`{}`)); err != nil {
				t.Fatalf("an unknown tool is answered, not failed: %v", err)
			}

			var logged bool
			for _, entry := range hook.AllEntries() {
				if strings.Contains(entry.Message, "model called a tool that does not exist") {
					logged = true
				}
			}
			if logged != tc.wantLog {
				t.Errorf("logged the missing tool = %v, want %v", logged, tc.wantLog)
			}
		})
	}
}

func TestExecutor_Execute_WritesTheMessageResultEvenAfterTheFlowStops(t *testing.T) {
	t.Parallel()

	// The toolcall row and the message log are two halves of one record. A flow
	// cancelled between them used to leave the toolcall finished with a result
	// and the message the operator reads empty.
	ctx, stop := context.WithCancel(t.Context())
	defer stop()

	tclp, mlp := &executorToolCallLog{}, &executorMsgLog{}
	ce := &customExecutor{
		flowID: 1, mlp: mlp, tclp: tclp, replacer: identityReplacer{},
		handlers: map[string]ExecutorHandler{TerminalToolName: func(
			context.Context, string, json.RawMessage,
		) (string, error) {
			stop()

			return "uid=0(root)", nil
		}},
	}

	args := mustJSON(map[string]any{"input": "id", "message": "Check the user."})
	if _, err := ce.Execute(ctx, 1, "call-1", TerminalToolName, TerminalToolName, "", args); err != nil {
		t.Fatalf("the command ran: %v", err)
	}

	if !slices.Equal(tclp.success, []string{"uid=0(root)"}) {
		t.Fatalf("the tool call log holds %q, want the result", tclp.success)
	}
	if mlp.result != "uid=0(root)" {
		t.Errorf("the message log holds %q, want the same result the tool call log has", mlp.result)
	}
}
