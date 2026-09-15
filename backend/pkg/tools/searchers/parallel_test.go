package searchers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	"pentagi/pkg/version"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type parallelRPCRequest struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params map[string]any  `json:"params"`
}

// Exercise the fixed production URL through the project's existing proxy/CA
// path. Only the MCP server is mocked; the provider and SDK run normally.
func newParallelTestSearcher(t *testing.T, call func(http.ResponseWriter, *http.Request, parallelRPCRequest)) (Searcher, *atomic.Int64) {
	t.Helper()
	requests := &atomic.Int64{}
	proxy, err := newTestProxy("search.parallel.ai", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The shared test proxy serves one request per CONNECT tunnel.
		w.Header().Set("Connection", "close")
		requests.Add(1)
		if r.URL.Path != "/mcp" || r.URL.RawQuery != "" {
			t.Errorf("unexpected MCP URL: %s", r.URL)
		}
		if r.Header.Get("User-Agent") != "pentagi/"+version.GetBinaryVersion() {
			t.Errorf("project identity not sent on %s: %q", r.Method, r.Header.Get("User-Agent"))
		}
		if r.Header.Get("Authorization") != "" || r.Header.Get("X-Api-Key") != "" {
			t.Error("free MCP request carried authentication")
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" || !strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
			t.Errorf("unexpected MCP request headers: %s %v", r.Method, r.Header)
		}
		var req parallelRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("invalid request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		respond := func(result any) {
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
		}
		switch req.Method {
		case "server/discover":
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID,
				"error": map[string]any{"code": -32601, "message": "legacy server"}})
		case "initialize":
			w.Header().Set("Mcp-Session-Id", "server-owned-test-session")
			respond(map[string]any{"protocolVersion": req.Params["protocolVersion"],
				"serverInfo": map[string]string{"name": "mock", "version": "1"}, "capabilities": map[string]any{"tools": map[string]any{}}})
		case "notifications/initialized", "notifications/cancelled":
			w.WriteHeader(http.StatusAccepted)
		case "tools/list":
			if r.Header.Get("Mcp-Session-Id") != "server-owned-test-session" || r.Header.Get("Mcp-Protocol-Version") == "" {
				t.Errorf("negotiated MCP headers missing: %v", r.Header)
			}
			if req.Params["cursor"] == "search-page" {
				respond(map[string]any{"tools": []map[string]any{{"name": "web_search", "inputSchema": map[string]any{"type": "object"}}}})
			} else {
				respond(map[string]any{"tools": []map[string]any{{"name": "web_fetch", "inputSchema": map[string]any{"type": "object"}}}, "nextCursor": "search-page"})
			}
		case "tools/call":
			call(w, r, req)
		default:
			t.Errorf("unexpected MCP method: %s", req.Method)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = proxy.Close() })
	return NewParallel(&config.Config{
		ParallelSearchEnabled: true, HTTPClientTimeout: 2,
		ProxyURL: proxy.URL(), ExternalSSLCAPath: proxy.CACertPath(),
		TavilyAPIKey: "another-provider-key",
	}), requests
}

func parallelReply(w http.ResponseWriter, req parallelRPCRequest, result any) {
	_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
}

func TestParallelHandleDiscoveryResultsAndRepeatedAttribution(t *testing.T) {
	var calls atomic.Int64
	searcher, _ := newParallelTestSearcher(t, func(w http.ResponseWriter, r *http.Request, req parallelRPCRequest) {
		calls.Add(1)
		if req.Params["name"] != "web_search" {
			t.Errorf("wrong MCP tool: %v", req.Params)
		}
		arguments := req.Params["arguments"].(map[string]any)
		if len(arguments) != 2 || arguments["objective"] != "official documentation" || arguments["search_queries"].([]any)[0] != "official documentation" {
			t.Errorf("unsupported or missing MCP arguments: %v", arguments)
		}
		if r.Header.Get("Mcp-Session-Id") != "server-owned-test-session" {
			t.Error("tool call lost the transport session")
		}
		parallelReply(w, req, map[string]any{
			"content": []map[string]string{{"type": "text", "text": "duplicate representation must not be used"}},
			"structuredContent": map[string]any{
				"results": []map[string]any{
					{"url": "https://example.com/first", "title": "First source", "publish_date": "2026-09-14", "excerpts": []string{"Useful excerpt"}},
					{"url": "https://example.com/second", "title": "Second source", "excerpts": []string{"Another excerpt"}},
				},
				"warnings": []map[string]string{{"type": "input_validation_warning", "message": "Server warning retained"}},
			},
		})
	})
	for range 2 {
		got, err := searcher.Handle(context.Background(), Request{Query: "official documentation", MaxResults: 1})
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"First source", "https://example.com/first", "Useful excerpt", "2026-09-14", "Server warning retained", "limited locally to 1"} {
			if !strings.Contains(got, want) {
				t.Errorf("result lost %q: %s", want, got)
			}
		}
		if strings.Contains(got, "Second source") || strings.Contains(got, "duplicate representation") {
			t.Errorf("result limit or representation selection failed: %s", got)
		}
	}
	if calls.Load() != 2 || searcher.Engine() != database.SearchengineTypeParallel {
		t.Fatalf("unexpected calls or attribution: %d %s", calls.Load(), searcher.Engine())
	}
}

func TestParallelFailuresRemainFailures(t *testing.T) {
	cases := []struct {
		name    string
		handler func(http.ResponseWriter, parallelRPCRequest)
		retry   bool
		message string
	}{
		{"http 400", func(w http.ResponseWriter, _ parallelRPCRequest) { w.WriteHeader(400) }, false, "HTTP 400"},
		{"http 429", func(w http.ResponseWriter, _ parallelRPCRequest) {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(429)
		}, true, "HTTP 429"},
		{"http 503", func(w http.ResponseWriter, _ parallelRPCRequest) { w.WriteHeader(503) }, true, "HTTP 503"},
		{"rpc error", func(w http.ResponseWriter, req parallelRPCRequest) {
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "error": map[string]any{"code": -32000, "message": "service error"}})
		}, false, "service error"},
		{"null id rpc error", func(w http.ResponseWriter, _ parallelRPCRequest) {
			_, _ = io.WriteString(w, `{"jsonrpc":"2.0","id":null,"error":{"code":-32000,"message":"quota error"}}`)
		}, false, "quota error"},
		{"tool error", func(w http.ResponseWriter, req parallelRPCRequest) {
			parallelReply(w, req, map[string]any{"isError": true, "content": []map[string]string{{"type": "text", "text": "tool failed"}}})
		}, false, "tool failed"},
		{"malformed envelope", func(w http.ResponseWriter, _ parallelRPCRequest) { _, _ = io.WriteString(w, "invalid json") }, false, "decode"},
		{"missing results", func(w http.ResponseWriter, req parallelRPCRequest) {
			parallelReply(w, req, map[string]any{"structuredContent": map[string]any{"message": "bad"}})
		}, false, "results array"},
		{"null results", func(w http.ResponseWriter, req parallelRPCRequest) {
			parallelReply(w, req, map[string]any{"structuredContent": map[string]any{"results": nil}})
		}, false, "results array"},
		{"invalid source", func(w http.ResponseWriter, req parallelRPCRequest) {
			parallelReply(w, req, map[string]any{"structuredContent": map[string]any{"results": []map[string]any{{"url": "invalid", "excerpts": []string{"text"}}}}})
		}, false, "source URL"},
		{"oversized chunked response", func(w http.ResponseWriter, _ parallelRPCRequest) {
			w.(http.Flusher).Flush()
			_, _ = io.WriteString(w, strings.Repeat("x", parallelMaxResponseBytes+1))
		}, false, "exceeds"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			searcher, _ := newParallelTestSearcher(t, func(w http.ResponseWriter, _ *http.Request, req parallelRPCRequest) { tc.handler(w, req) })
			got, err := searcher.Handle(context.Background(), Request{Query: "public query", MaxResults: 5})
			if err == nil || got != "" || IsRetryable(err) != tc.retry || (!tc.retry && !IsFatal(err)) || !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("wrong failure: got=%q err=%v retry=%v fatal=%v", got, err, IsRetryable(err), IsFatal(err))
			}
			if tc.name == "http 429" {
				var retry *RetryableError
				if !errors.As(err, &retry) || retry.RetryAfter != 2*time.Second {
					t.Fatalf("Retry-After not preserved: %v", err)
				}
			}
		})
	}
}

func TestParallelTextEmptyAndBoundedOutput(t *testing.T) {
	for _, text := range []string{
		`{"results":[],"warnings":[{"type":"warning","message":"empty warning"}]}`,
		`{"results":[{"url":"https://example.com/first","title":null,"excerpts":["` + strings.Repeat("é", parallelMaxOutputBytes) + `"]},{"url":"https://example.com/second","excerpts":["second"]}]}`,
	} {
		searcher, _ := newParallelTestSearcher(t, func(w http.ResponseWriter, _ *http.Request, req parallelRPCRequest) {
			parallelReply(w, req, map[string]any{"content": []map[string]string{{"type": "text", "text": text}}})
		})
		got, err := searcher.Handle(context.Background(), Request{Query: "public query", MaxResults: 5})
		if err != nil || len(got) > parallelMaxOutputBytes || !utf8.ValidString(got) {
			t.Fatalf("invalid text result: bytes=%d err=%v", len(got), err)
		}
		if strings.Contains(text, "empty warning") {
			if !strings.Contains(got, "No results found") || !strings.Contains(got, "empty warning") {
				t.Fatalf("empty success or its warning lost: %s", got)
			}
		} else if !strings.Contains(got, "output truncated") || !strings.Contains(got, "https://example.com/first") || !strings.Contains(got, "https://example.com/second") {
			t.Fatal("bounded output lost citations or its truncation notice")
		}
	}
}

func TestParallelRejectsUnsupportedInputsBeforeDispatch(t *testing.T) {
	searcher, requests := newParallelTestSearcher(t, func(http.ResponseWriter, *http.Request, parallelRPCRequest) { t.Error("unexpected call") })
	for _, req := range []Request{{Query: " "}, {Query: "query", ExploitType: "tools"}, {Query: "query", Sort: "date"}} {
		if _, err := searcher.Handle(context.Background(), req); !IsFatal(err) {
			t.Fatalf("unsupported input was accepted: %+v %v", req, err)
		}
	}
	if requests.Load() != 0 {
		t.Fatal("unsupported input reached Parallel")
	}
	for _, cfg := range []*config.Config{nil, {}} {
		if _, err := NewParallel(cfg).Handle(context.Background(), Request{Query: "query"}); !errors.Is(err, ErrNotConfigured) {
			t.Fatalf("disabled provider was dispatched: %v", err)
		}
	}
}

func TestParallelCancellationAndHostTimeout(t *testing.T) {
	searcher, _ := newParallelTestSearcher(t, func(w http.ResponseWriter, _ *http.Request, req parallelRPCRequest) {
		time.Sleep(1500 * time.Millisecond)
		parallelReply(w, req, map[string]any{"structuredContent": map[string]any{"results": []any{}}})
	})
	searcher.(*parallel).cfg.HTTPClientTimeout = 1
	started := time.Now()
	if _, err := searcher.Handle(context.Background(), Request{Query: "query"}); !IsRetryable(err) || time.Since(started) > 1400*time.Millisecond {
		t.Fatalf("host timeout not honored: elapsed=%v err=%v", time.Since(started), err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := searcher.Handle(ctx, Request{Query: "query"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation not preserved: %v", err)
	}
}

func TestParallelRedirectRejectsDestinationAndPreservesHeaders(t *testing.T) {
	var destinationRequests atomic.Int64
	destination := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { destinationRequests.Add(1) }))
	defer destination.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "existing-client pentagi/"+version.GetBinaryVersion() || r.Header.Get("X-Test") != "preserved" {
			t.Errorf("existing client headers were lost: %v", r.Header)
		}
		w.Header().Set("Location", destination.URL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))
	defer source.Close()
	client := &http.Client{Transport: &parallelTransport{base: http.DefaultTransport}}
	req, err := http.NewRequest(http.MethodPost, source.URL, strings.NewReader("public query"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "existing-client")
	req.Header.Set("X-Test", "preserved")
	if _, err := client.Do(req); !IsFatal(err) || destinationRequests.Load() != 0 {
		t.Fatalf("redirect sent data to the destination: err=%v requests=%d", err, destinationRequests.Load())
	}
}

func TestParallelSSEMatchingMultilineResponseDoesNotWaitForEOF(t *testing.T) {
	streamDone := make(chan struct{})
	defer close(streamDone)
	searcher, _ := newParallelTestSearcher(t, func(w http.ResponseWriter, _ *http.Request, req parallelRPCRequest) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, ": "+strings.Repeat(" ", 8<<10)+"\n\n")
		_, _ = fmt.Fprint(w, "data: {\"jsonrpc\":\"2.0\",\"method\":\"notifications/progress\",\"params\":{}}\n\n")
		_, _ = fmt.Fprint(w, "data: {\"jsonrpc\":\"2.0\",\"id\":9999,\"result\":{\"structuredContent\":{\"results\":[]}}}\n\n")
		_, _ = fmt.Fprintf(w, "data: {\"jsonrpc\":\"2.0\",\"id\":%s,\n", req.ID)
		_, _ = fmt.Fprint(w, "data: \"result\":{\"structuredContent\":{\"results\":[]},\"content\":[]}}\n\n")
		_, _ = fmt.Fprint(w, ": "+strings.Repeat(" ", 8<<10)+"\n\n")
		w.(http.Flusher).Flush()
		// Keep the response open until Handle returns, so waiting for EOF cannot pass.
		<-streamDone
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	got, err := searcher.Handle(ctx, Request{Query: "query"})
	if err != nil || !strings.Contains(got, "No results found") {
		t.Fatalf("SSE waited for EOF or failed: result=%q err=%v", got, err)
	}
}

func TestParallelMissingPayloadIsNotEmptySuccess(t *testing.T) {
	if _, err := parseParallelResult(&mcp.CallToolResult{}, 5); !IsFatal(err) {
		t.Fatalf("missing payload treated as empty success: %v", err)
	}
}
