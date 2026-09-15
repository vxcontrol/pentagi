package searchers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/system"
	"pentagi/pkg/version"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	parallelMCPURL           = "https://search.parallel.ai/mcp"
	parallelMaxResponseBytes = 2 << 20
	parallelMaxOutputBytes   = 32 << 10
)

type parallel struct {
	cfg *config.Config
}

// NewParallel provides anonymous Search MCP access only when explicitly enabled.
func NewParallel(cfg *config.Config) Searcher {
	return &parallel{cfg: cfg}
}

func (p *parallel) Engine() database.SearchengineType {
	return database.SearchengineTypeParallel
}

func (p *parallel) IsAvailable() bool {
	return p.cfg != nil && p.cfg.ParallelSearchEnabled
}

func (p *parallel) Handle(ctx context.Context, req Request) (string, error) {
	if !p.IsAvailable() {
		return "", ErrNotConfigured
	}
	if strings.TrimSpace(req.Query) == "" {
		return "", Fatal(fmt.Errorf("parallel: query cannot be empty"))
	}
	if req.ExploitType != "" || req.Sort != "" {
		return "", Fatal(fmt.Errorf("parallel: exploit_type and sort are unsupported"))
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	result, err := p.search(ctx, req)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("search engine error"),
			langfuse.WithEventInput(req.Query),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"engine":      "parallel",
				"query":       req.Query,
				"max_results": req.MaxResults,
				"error":       err.Error(),
			}),
		)
	}
	return result, err
}

func (p *parallel) search(ctx context.Context, req Request) (string, error) {
	httpClient, err := system.GetHTTPClient(p.cfg)
	if err != nil {
		return "", Fatal(fmt.Errorf("parallel: failed to create HTTP client: %w", err))
	}
	transport := &parallelTransport{base: httpClient.Transport}
	httpClient.Transport = transport
	defer httpClient.CloseIdleConnections()

	// Bound the entire initialize/discover/search operation by the host's timeout.
	// An explicit zero keeps the host's unlimited timeout and caller cancellation.
	var cancel context.CancelFunc
	if httpClient.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, httpClient.Timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{
		Name: "pentagi", Version: version.GetBinaryVersion(),
	}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint: parallelMCPURL, HTTPClient: httpClient,
		DisableStandaloneSSE: true, MaxRetries: -1,
		MaxEventSize: parallelMaxResponseBytes,
	}, nil)
	if err != nil {
		return "", transport.classify(ctx, err)
	}
	defer func() {
		// Stop any open POST stream before best-effort SDK session cleanup. The SDK
		// gives DELETE its own bounded deadline; cleanup never replaces the result.
		cancel()
		_ = session.Close()
	}()

	found := false
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			return "", transport.classify(ctx, err)
		}
		if tool.Name == "web_search" {
			found = true
			break
		}
	}
	if !found {
		return "", Fatal(fmt.Errorf("parallel: MCP server does not expose web_search"))
	}

	// Searcher has no conversation identity or model metadata. Omit these optional
	// MCP arguments rather than inventing shared or per-request conversation state.
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "web_search",
		Arguments: map[string]any{
			"objective": req.Query, "search_queries": []string{req.Query},
		},
	})
	if err != nil {
		return "", transport.classify(ctx, err)
	}
	if result.IsError {
		return "", Fatal(fmt.Errorf("parallel: MCP tool error: %s", parallelToolError(result)))
	}
	return parseParallelResult(result, req.MaxResults)
}

// parallelTransport adds host-owned HTTP policy while the SDK owns MCP protocol,
// response correlation, SSE events, pagination and transport sessions.
type parallelTransport struct {
	base    http.RoundTripper
	mu      sync.Mutex
	bodyErr error
}

func (t *parallelTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	// Identify PentAGI so Parallel can measure aggregate free MCP usage. Keep this
	// project-wide and preserve it across transport changes, without user identifiers.
	identity := "pentagi/" + version.GetBinaryVersion()
	if existing := req.Header.Get("User-Agent"); existing != "" {
		identity = existing + " " + identity
	}
	req.Header.Set("User-Agent", identity)

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, Retryable(fmt.Errorf("parallel: HTTP request failed: %w", err), 0)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = resp.Body.Close()
		// Returning an error here also rejects redirects before http.Client can
		// forward the query, MCP headers or proxy configuration to another origin.
		err := ClassifyHTTPStatus(resp.StatusCode, "parallel: MCP request failed")
		if resp.StatusCode == http.StatusRequestTimeout {
			err = Retryable(fmt.Errorf("parallel: MCP request timed out (HTTP 408)"), 0)
		}
		var retry *RetryableError
		if errors.As(err, &retry) {
			retry.RetryAfter = parallelRetryAfter(resp.Header.Get("Retry-After"))
		}
		return nil, err
	}

	mediaType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if mediaType == "application/json" {
		body, err := io.ReadAll(io.LimitReader(resp.Body, parallelMaxResponseBytes+1))
		_ = resp.Body.Close()
		if err != nil {
			return nil, Retryable(fmt.Errorf("parallel: response read failed: %w", err), 0)
		}
		if len(body) > parallelMaxResponseBytes {
			return nil, Fatal(fmt.Errorf("parallel: response exceeds %d bytes", parallelMaxResponseBytes))
		}
		// Some service errors have a null ID. Surface them immediately instead of
		// leaving the SDK waiting for a correlated success that cannot arrive.
		var envelope struct {
			ID    json.RawMessage `json:"id"`
			Error *jsonrpc.Error  `json:"error"`
		}
		if json.Unmarshal(body, &envelope) == nil && envelope.Error != nil && (len(envelope.ID) == 0 || string(envelope.ID) == "null") {
			return nil, Fatal(fmt.Errorf("parallel: MCP error: %w", envelope.Error))
		}
		resp.Body = io.NopCloser(bytes.NewReader(body))
	} else {
		resp.Body = &parallelBoundedBody{ReadCloser: resp.Body, remaining: parallelMaxResponseBytes, transport: t}
	}
	return resp, nil
}

func (t *parallelTransport) CloseIdleConnections() {
	if closer, ok := t.base.(interface{ CloseIdleConnections() }); ok {
		closer.CloseIdleConnections()
	}
}

func (t *parallelTransport) classify(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return ctx.Err()
		}
		return Retryable(ctx.Err(), 0)
	}
	if IsRetryable(err) || IsFatal(err) {
		return err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.bodyErr != nil {
		return t.bodyErr
	}
	return Fatal(fmt.Errorf("parallel: MCP request failed: %w", err))
}

type parallelBoundedBody struct {
	io.ReadCloser
	remaining int64
	transport *parallelTransport
}

func (b *parallelBoundedBody) Read(p []byte) (int, error) {
	if b.remaining < 0 {
		return 0, Fatal(fmt.Errorf("parallel: response exceeds %d bytes", parallelMaxResponseBytes))
	}
	p = p[:min(int64(len(p)), b.remaining+1)]
	n, err := b.ReadCloser.Read(p)
	b.remaining -= int64(n)
	if b.remaining < 0 {
		err = Fatal(fmt.Errorf("parallel: response exceeds %d bytes", parallelMaxResponseBytes))
		n = 0
	} else if err != nil && !errors.Is(err, io.EOF) {
		err = Retryable(fmt.Errorf("parallel: response read failed: %w", err), 0)
	}
	if err != nil && !errors.Is(err, io.EOF) {
		b.transport.mu.Lock()
		b.transport.bodyErr = err
		b.transport.mu.Unlock()
	}
	return n, err
}

func parallelRetryAfter(value string) time.Duration {
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil && seconds > 0 {
		// Avoid overflowing a duration on an untrusted header.
		return time.Duration(min(seconds, 86400)) * time.Second
	}
	if deadline, err := http.ParseTime(value); err == nil {
		return max(time.Until(deadline), 0)
	}
	return 0
}

func parallelToolError(result *mcp.CallToolResult) string {
	for _, content := range result.Content {
		if text, ok := content.(*mcp.TextContent); ok {
			return parallelTrim(text.Text, 1000)
		}
	}
	return "the server reported isError=true"
}

type parallelSearchResult struct {
	Results  []parallelResult `json:"results"`
	Warnings []struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"warnings"`
}

type parallelResult struct {
	URL         string   `json:"url"`
	Title       *string  `json:"title"`
	PublishDate *string  `json:"publish_date"`
	Excerpts    []string `json:"excerpts"`
}

func parseParallelResult(result *mcp.CallToolResult, maxResults int) (string, error) {
	if result == nil {
		return "", Fatal(fmt.Errorf("parallel: missing MCP tool result"))
	}
	var body []byte
	if result.StructuredContent != nil {
		var err error
		body, err = json.Marshal(result.StructuredContent)
		if err != nil {
			return "", Fatal(fmt.Errorf("parallel: invalid structured result: %w", err))
		}
	} else {
		for _, content := range result.Content {
			if text, ok := content.(*mcp.TextContent); ok {
				body = []byte(text.Text)
				break
			}
		}
	}
	var payload parallelSearchResult
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", Fatal(fmt.Errorf("parallel: invalid search result: %w", err))
	}
	if payload.Results == nil {
		return "", Fatal(fmt.Errorf("parallel: search result must contain a results array"))
	}
	for _, result := range payload.Results {
		u, err := url.Parse(result.URL)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || result.Excerpts == nil {
			return "", Fatal(fmt.Errorf("parallel: invalid source URL or excerpts array"))
		}
	}
	if maxResults < 1 {
		maxResults = 5
	}
	maxResults = min(maxResults, 25)
	return formatParallelResult(payload, maxResults), nil
}

func formatParallelResult(payload parallelSearchResult, maxResults int) string {
	const notice = "\n\n[Parallel search output truncated.]\n"
	budget := parallelMaxOutputBytes - len(notice)
	var writer strings.Builder
	truncated := false
	write := func(text string) bool {
		remaining := budget - writer.Len()
		if len(text) > remaining {
			writer.WriteString(parallelTrim(text, remaining))
			truncated = true
			return false
		}
		writer.WriteString(text)
		return true
	}
	write("# Links\n\n")
	results := payload.Results[:min(len(payload.Results), maxResults)]
	if len(payload.Results) > maxResults {
		write(fmt.Sprintf("[Parallel returned %d results; limited locally to %d.]\n\n", len(payload.Results), maxResults))
	}
	if len(results) == 0 {
		write("No results found.\n")
	}
	// Put citations before excerpts so one oversized excerpt cannot consume the
	// space needed to identify the other returned sources. Never slice a URL.
	linked := 0
	for i, result := range results {
		title := result.URL
		if result.Title != nil && *result.Title != "" {
			title = *result.Title
		}
		link := fmt.Sprintf("%d. %s\n   URL: %s\n\n", i+1, title, result.URL)
		if len(link) > budget-writer.Len() {
			truncated = true
			break
		}
		write(link)
		linked++
	}
	for _, warning := range payload.Warnings {
		if !write(fmt.Sprintf("Warning (%s): %s\n\n", warning.Type, warning.Message)) {
			break
		}
	}
	for i, result := range results[:linked] {
		if !write(fmt.Sprintf("## Source %d\n\n", i+1)) {
			break
		}
		if result.PublishDate != nil {
			if !write("Published: " + *result.PublishDate + "\n\n") {
				break
			}
		}
		for _, excerpt := range result.Excerpts {
			if !write(excerpt + "\n\n") {
				break
			}
		}
	}
	if truncated {
		writer.WriteString(notice)
	}
	return writer.String()
}

func parallelTrim(text string, limit int) string {
	if len(text) <= limit {
		return text
	}
	for limit > 0 && !utf8.RuneStart(text[limit]) {
		limit--
	}
	return text[:limit]
}
