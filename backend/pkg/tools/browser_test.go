package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// browserTarget is the only page the stand-in scraper serves.
const browserTarget = "https://example.com/page"

const browserLinksJSON = `[` +
	`{"Title":"Example","Link":"https://example.com"},` +
	`{"Title":"  ","Link":"https://example.com/untitled"},` +
	`{"Title":"No link","Link":"  "}` +
	`]`

var (
	browserMarkdownPage  = strings.Repeat("A", minMdContentSize+1)
	browserHTMLPage      = strings.Repeat("<p>x</p>", minHtmlContentSize/8+1)
	browserScreenshotPNG = strings.Repeat("\x89PNG", minImgContentSize/4+1)
)

// browserScreenshotRecorder stands in for the screenshot table and, like it, refuses a done context.
type browserScreenshotRecorder struct {
	mu       sync.Mutex
	err      error
	calls    int
	lastName string
	lastURL  string
	lastTask *int64
	lastSub  *int64
}

func (m *browserScreenshotRecorder) PutScreenshot(ctx context.Context, name, url string, taskID, subtaskID *int64) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls++
	m.lastName = name
	m.lastURL = url
	m.lastTask = taskID
	m.lastSub = subtaskID
	return 1, m.err
}

func browserIDString(id *int64) string {
	if id == nil {
		return "<nil>"
	}
	return strconv.FormatInt(*id, 10)
}

// browserScraper serves pages every fetch accepts unless answers override a path.
func browserScraper(t *testing.T, answers map[string]scraperAnswer) *httptest.Server {
	t.Helper()

	pages := map[string]scraperAnswer{
		"/markdown":   {http.StatusOK, browserMarkdownPage},
		"/html":       {http.StatusOK, browserHTMLPage},
		"/links":      {http.StatusOK, browserLinksJSON},
		"/screenshot": {http.StatusOK, browserScreenshotPNG},
	}
	for path, answer := range answers {
		pages[path] = answer
	}

	return fakeScraper(t, browserTarget, pages)
}

func TestBrowser_Handle_RefusesACallItCannotServe(t *testing.T) {
	configured := &browser{scPrvURL: "http://127.0.0.1:8080"}

	tests := []struct {
		name string
		b    *browser
		tool string
		args string
		want string
	}{
		{
			name: "a browser with no scraper configured",
			b:    &browser{},
			tool: "browser",
			args: `{"url":"https://example.com/page","action":"markdown","message":"m"}`,
			want: "browser is not available",
		},
		{
			name: "another tool's name",
			b:    configured,
			tool: "not-browser",
			args: `{}`,
			want: "unknown tool: not-browser",
		},
		{
			name: "arguments that are not json",
			b:    configured,
			tool: "browser",
			args: `{`,
			want: "failed to unmarshal browser action",
		},
		{
			name: "an action it does not know",
			b:    configured,
			tool: "browser",
			args: `{"url":"https://example.com","action":"unknown","message":"m"}`,
			want: "unknown browser action: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.b.Handle(t.Context(), tt.tool, json.RawMessage(tt.args))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Handle() error = %v, want it to contain %q", err, tt.want)
			}
			if got != "" {
				t.Errorf("Handle() result = %q alongside an error, want none", got)
			}
		})
	}
}

// The page is the result whatever becomes of its screenshot.
func TestBrowser_Handle_ReturnsTheScrapedPage(t *testing.T) {
	const (
		markdownArgs = `{"url":"https://example.com/page","action":"markdown","message":"m"}`
		htmlArgs     = `{"url":"https://example.com/page","action":"html","message":"m"}`
		linksArgs    = `{"url":"https://example.com/page","action":"links","message":"m"}`
		wantLinks    = "Links list from URL 'https://example.com/page'\n" +
			"[Example](https://example.com)\n" +
			"[UNTITLED](https://example.com/untitled)\n"
	)
	failedScreenshot := map[string]scraperAnswer{"/screenshot": {http.StatusInternalServerError, ""}}

	tests := []struct {
		name           string
		args           string
		answers        map[string]scraperAnswer
		dataDirIsAFile bool
		registerErr    error
		want           string
		wantScreenshot bool
	}{
		{
			name:           "markdown",
			args:           markdownArgs,
			want:           browserMarkdownPage,
			wantScreenshot: true,
		},
		{
			name:           "html",
			args:           htmlArgs,
			want:           browserHTMLPage,
			wantScreenshot: true,
		},
		{
			name:           "links, skipping a blank link and naming an untitled one",
			args:           linksArgs,
			want:           wantLinks,
			wantScreenshot: true,
		},
		{
			name:           "no action given reads the page as markdown",
			args:           `{"url":"https://example.com/page","message":"m"}`,
			want:           browserMarkdownPage,
			wantScreenshot: true,
		},
		{
			name:           "a url wrapped in newlines",
			args:           `{"url":"\nhttps://example.com/page \n","action":"markdown","message":"m"}`,
			want:           browserMarkdownPage,
			wantScreenshot: true,
		},
		{
			name:    "an almost empty markdown page comes with a warning",
			args:    markdownArgs,
			answers: map[string]scraperAnswer{"/markdown": {http.StatusOK, strings.Repeat("x", 49)}},
			want: "[WARNING: page returned very little content (49 bytes), it may be a redirect, error page, or near-empty]\n\n" +
				strings.Repeat("x", 49),
			wantScreenshot: true,
		},
		{
			// 60 bytes clears the markdown floor but not the html one.
			name:    "an html page under the html floor comes with a warning",
			args:    htmlArgs,
			answers: map[string]scraperAnswer{"/html": {http.StatusOK, strings.Repeat("x", 60)}},
			want: "[WARNING: page returned very little HTML content (60 bytes), it may be a redirect, error page, or near-empty]\n\n" +
				strings.Repeat("x", 60),
			wantScreenshot: true,
		},
		{
			name:    "markdown when the scraper fails the screenshot",
			args:    markdownArgs,
			answers: failedScreenshot,
			want:    browserMarkdownPage,
		},
		{
			name:    "html when the scraper fails the screenshot",
			args:    htmlArgs,
			answers: failedScreenshot,
			want:    browserHTMLPage,
		},
		{
			name:    "links when the scraper fails the screenshot",
			args:    linksArgs,
			answers: failedScreenshot,
			want:    wantLinks,
		},
		{
			name:    "markdown when the screenshot is too small to be an image",
			args:    markdownArgs,
			answers: map[string]scraperAnswer{"/screenshot": {http.StatusOK, "tiny"}},
			want:    browserMarkdownPage,
		},
		{
			name:           "markdown when the screenshot cannot be saved",
			args:           markdownArgs,
			dataDirIsAFile: true,
			want:           browserMarkdownPage,
		},
		{
			name:           "markdown when the screenshot cannot be registered",
			args:           markdownArgs,
			registerErr:    errors.New("database is down"),
			want:           browserMarkdownPage,
			wantScreenshot: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := browserScraper(t, tt.answers)

			dataDir := t.TempDir()
			if tt.dataDirIsAFile {
				dataDir = filepath.Join(dataDir, "a-file")
				if err := os.WriteFile(dataDir, nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			taskID, subtaskID := int64(7), int64(11)
			scp := &browserScreenshotRecorder{err: tt.registerErr}
			b := &browser{
				flowID:    1,
				taskID:    &taskID,
				subtaskID: &subtaskID,
				dataDir:   dataDir,
				scPubURL:  ts.URL,
				scp:       scp,
			}

			got, err := b.Handle(t.Context(), "browser", json.RawMessage(tt.args))
			if err != nil {
				t.Fatalf("Handle() returned an error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Handle() result = %q, want %q", got, tt.want)
			}

			scp.mu.Lock()
			defer scp.mu.Unlock()

			if !tt.wantScreenshot {
				if scp.calls != 0 {
					t.Errorf("PutScreenshot() calls = %d, want none", scp.calls)
				}
				return
			}
			if scp.calls != 1 {
				t.Fatalf("PutScreenshot() calls = %d, want 1", scp.calls)
			}
			if scp.lastURL != browserTarget {
				t.Errorf("PutScreenshot() url = %q, want %q", scp.lastURL, browserTarget)
			}
			if scp.lastTask == nil || *scp.lastTask != taskID || scp.lastSub == nil || *scp.lastSub != subtaskID {
				t.Errorf("PutScreenshot() task/subtask = %s/%s, want %d/%d",
					browserIDString(scp.lastTask), browserIDString(scp.lastSub), taskID, subtaskID)
			}
			if scp.lastName == "" {
				t.Fatal("PutScreenshot() screenshot name is empty")
			}
			saved, err := os.ReadFile(filepath.Join(dataDir, "screenshots", "flow-1", scp.lastName))
			if err != nil {
				t.Fatalf("the registered screenshot is not on disk: %v", err)
			}
			if string(saved) != browserScreenshotPNG {
				t.Errorf("the screenshot on disk holds %d bytes, want the %d the scraper served", len(saved), len(browserScreenshotPNG))
			}
		})
	}
}

func TestBrowser_Handle_AnswersAFailedFetchWithTheCause(t *testing.T) {
	const (
		markdownArgs = `{"url":"https://example.com/page","action":"markdown","message":"m"}`
		htmlArgs     = `{"url":"https://example.com/page","action":"html","message":"m"}`
		linksArgs    = `{"url":"https://example.com/page","action":"links","message":"m"}`
	)

	tests := []struct {
		name    string
		args    string
		answers map[string]scraperAnswer
		want    []string
	}{
		{
			name: "markdown of a url that does not parse",
			args: `{"url":"http://exa mple.com/","action":"markdown","message":"m"}`,
			want: []string{"failed to resolve url: failed to parse url"},
		},
		{
			name: "html of a url that does not parse",
			args: `{"url":"http://exa mple.com/","action":"html","message":"m"}`,
			want: []string{"failed to resolve url: failed to parse url"},
		},
		{
			name: "links of a url that does not parse",
			args: `{"url":"http://exa mple.com/","action":"links","message":"m"}`,
			want: []string{"failed to resolve url: failed to parse url"},
		},
		{
			name: "markdown of a pdf",
			args: `{"url":"https://example.com/report.pdf","action":"markdown","message":"m"}`,
			want: []string{
				"binary/non-HTML resource",
				"cannot be rendered as markdown. Use the terminal tool with curl/wget",
			},
		},
		{
			name: "html of a pdf",
			args: `{"url":"https://example.com/report.pdf","action":"html","message":"m"}`,
			want: []string{
				"binary/non-HTML resource",
				"cannot be rendered as HTML. Use the terminal tool with curl/wget",
			},
		},
		{
			name:    "an empty markdown page",
			args:    markdownArgs,
			answers: map[string]scraperAnswer{"/markdown": {http.StatusOK, ""}},
			want:    []string{"failed to fetch content by url 'https://example.com/page': empty response body for scraper"},
		},
		{
			name:    "an empty html page",
			args:    htmlArgs,
			answers: map[string]scraperAnswer{"/html": {http.StatusOK, ""}},
			want:    []string{"failed to fetch content by url 'https://example.com/page': empty response body for scraper"},
		},
		{
			name:    "links when the scraper fails with a body",
			args:    linksArgs,
			answers: map[string]scraperAnswer{"/links": {http.StatusBadGateway, "<html><body>502 Bad Gateway</body></html>"}},
			want: []string{
				"failed to fetch links by url 'https://example.com/page': unexpected resp code for scraper",
				": 502, response: <html><body>502 Bad Gateway</body></html>",
			},
		},
		{
			name:    "links that are not a list",
			args:    linksArgs,
			answers: map[string]scraperAnswer{"/links": {http.StatusOK, `{"Title":"x"}`}},
			want:    []string{"failed to unmarshal links"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := browserScraper(t, tt.answers)
			scp := &browserScreenshotRecorder{}
			b := &browser{flowID: 1, dataDir: t.TempDir(), scPubURL: ts.URL, scp: scp}

			got, err := b.Handle(t.Context(), "browser", json.RawMessage(tt.args))
			if err != nil {
				t.Fatalf("a failed fetch must not fail the call, got: %v", err)
			}
			if !strings.HasPrefix(got, "browser tool 'browser' handled with error: ") {
				t.Fatalf("Handle() result = %q, want it to say the fetch failed", got)
			}
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("Handle() result = %q, want it to contain %q", got, want)
				}
			}

			scp.mu.Lock()
			defer scp.mu.Unlock()
			if scp.calls != 0 {
				t.Errorf("PutScreenshot() calls = %d for a page that failed, want none", scp.calls)
			}
		})
	}
}

// Handle never passes a screenshot with an error, so only a direct call can build that pair.
func TestBrowser_WrapCommandResult_StoresNoScreenshotForAFailedFetch(t *testing.T) {
	scp := &browserScreenshotRecorder{}
	b := &browser{scp: scp}

	got, err := b.wrapCommandResult(t.Context(), "browser", "payload", "https://example.com", "screen.png", errors.New("boom"))
	if err != nil {
		t.Fatalf("wrapCommandResult() returned an error: %v", err)
	}
	if want := "browser tool 'browser' handled with error: boom"; got != want {
		t.Fatalf("wrapCommandResult() = %q, want %q", got, want)
	}

	scp.mu.Lock()
	defer scp.mu.Unlock()
	if scp.calls != 0 {
		t.Fatalf("PutScreenshot() calls = %d on the error branch, want none", scp.calls)
	}
}

func TestBrowser_CallScraper_ReportsAnAnswerItCannotUse(t *testing.T) {
	answering := func(status int, body string) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
			fmt.Fprint(w, body)
		}
	}

	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    string // %s is the scraper URL
	}{
		{
			name:    "a 5xx answer carries its body",
			handler: answering(http.StatusBadGateway, "<html><body>502 Bad Gateway</body></html>"),
			want:    "unexpected resp code for scraper '%s': 502, response: <html><body>502 Bad Gateway</body></html>",
		},
		{
			name:    "a 5xx body over the cap is cut",
			handler: answering(http.StatusInternalServerError, strings.Repeat("x", 2000)),
			want:    "unexpected resp code for scraper '%s': 500, response: " + strings.Repeat("x", 512) + "... [truncated]",
		},
		{
			name:    "a 5xx body of blank space is not quoted",
			handler: answering(http.StatusServiceUnavailable, " \n "),
			want:    "unexpected resp code for scraper '%s': 503",
		},
		{
			name:    "a 4xx answer does not quote its body",
			handler: answering(http.StatusNotFound, "not found body"),
			want:    "unexpected resp code for scraper '%s': 404",
		},
		{
			name:    "an empty answer",
			handler: answering(http.StatusOK, ""),
			want:    "empty response body for scraper '%s'",
		},
		{
			name: "an answer cut short",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Length", "100")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "0123456789")
			},
			want: "failed to read response body for scraper '%s': unexpected EOF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			b := &browser{flowID: 1}
			got, err := b.callScraper(t.Context(), ts.URL)
			if err == nil {
				t.Fatalf("callScraper() = %q, want an error", got)
			}
			if want := fmt.Sprintf(tt.want, ts.URL); err.Error() != want {
				t.Errorf("callScraper() error = %q, want %q", err.Error(), want)
			}
			if got != nil {
				t.Errorf("callScraper() returned %q alongside an error", got)
			}
		})
	}
}

func TestBrowser_CallScraper_KeepsBasicAuthButNeverReturnsItsPassword(t *testing.T) {
	const (
		username = "scraper-user"
		password = "audit-secret/with space"
	)
	var authenticated int
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != username || pass != password {
			t.Errorf("scraper request did not carry the configured Basic Auth credentials")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		authenticated++
		if r.URL.Path == "/failure" {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprint(w, "render failed for "+password)
			return
		}
		if r.URL.Path == "/redirect" {
			w.Header().Set("Location", "https://redirect-user:redirect-secret@%ZZ")
			w.WriteHeader(http.StatusFound)
			return
		}
		if r.URL.Path == "/markdown" && r.URL.Query().Get("url") == "https://example.com/fail" {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprint(w, "render failed for "+password)
			return
		}
		fmt.Fprint(w, browserMarkdownPage)
	}))
	defer ts.Close()

	endpoint, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	endpoint.User = url.UserPassword(username, password)
	base := endpoint.String()
	b := &browser{}
	content, err := b.callScraper(t.Context(), base+"/markdown")
	if err != nil || string(content) != browserMarkdownPage {
		t.Fatalf("authenticated TLS request failed: content length %d, error %v", len(content), err)
	}
	configured := &browser{scPubURL: base}
	rendered, err := configured.getMD(t.Context(), browserTarget)
	if err != nil || rendered != browserMarkdownPage {
		t.Fatalf("configured browser could not fetch through Basic Auth: content length %d, error %v", len(rendered), err)
	}
	_, err = configured.getMD(t.Context(), "https://example.com/fail")
	if err == nil || strings.Contains(err.Error(), password) || !strings.Contains(err.Error(), "502") {
		t.Fatalf("configured browser returned an unsafe scraper failure: %v", err)
	}

	_, err = b.callScraper(t.Context(), base+"/failure?url=private-target-token")
	if err == nil {
		t.Fatal("scraper failure returned no error")
	}
	for _, secret := range []string{username, password, endpoint.User.String(), "private-target-token"} {
		if strings.Contains(err.Error(), secret) {
			t.Errorf("scraper failure disclosed %q", secret)
		}
	}
	if !strings.Contains(err.Error(), "502") || strings.Contains(err.Error(), "render failed") {
		t.Errorf("authenticated scraper failure must expose the status but not an untrusted body: %v", err)
	}
	_, err = b.callScraper(t.Context(), base+"/redirect")
	if err == nil || !strings.Contains(err.Error(), "network or TLS error") {
		t.Fatalf("malformed redirect = %v, want a safe transport error", err)
	}
	for _, secret := range []string{username, password, "redirect-user", "redirect-secret"} {
		if strings.Contains(err.Error(), secret) {
			t.Errorf("redirect error disclosed %q", secret)
		}
	}
	if authenticated != 5 {
		t.Errorf("authenticated scraper requests = %d, want 5", authenticated)
	}

	ts.Close()
	_, err = b.callScraper(t.Context(), base+"/transport")
	if err == nil {
		t.Fatal("closed scraper returned no transport error")
	}
	for _, secret := range []string{username, password, endpoint.User.String()} {
		if strings.Contains(err.Error(), secret) {
			t.Errorf("scraper transport error disclosed %q", secret)
		}
	}
}

func TestBrowser_InvalidScraperURLDoesNotRevealItsCredentials(t *testing.T) {
	b := &browser{scPubURL: "https://user:secret%ZZ@host.invalid"}
	_, err := b.resolveUrl(browserTarget)
	if err == nil || !strings.Contains(err.Error(), "invalid scraper URL") {
		t.Fatalf("resolveUrl() = %v, want a safe configuration error", err)
	}
	if strings.Contains(err.Error(), "user") || strings.Contains(err.Error(), "secret") {
		t.Fatal("invalid scraper URL exposed its credentials")
	}

	_, err = b.callScraper(t.Context(), b.scPubURL)
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatal("malformed request exposed its credentials")
	}
}

func TestBrowser_CallScraper_StopsWhenTheCallerIsCancelled(t *testing.T) {
	served := make(chan struct{})
	// testDone frees the handler if the fetch ignores cancellation, so ts.Close cannot hang.
	testDone := make(chan struct{})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(served)
		select {
		case <-r.Context().Done():
		case <-testDone:
		}
	}))
	defer ts.Close()
	defer close(testDone)

	b := &browser{flowID: 1}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := b.callScraper(ctx, ts.URL)
		done <- err
	}()

	select {
	case <-served:
	case <-time.After(5 * time.Second):
		t.Fatal("the scraper was never asked for the page")
	}
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("a cancelled fetch must report cancellation, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the scraper kept fetching after the caller was cancelled")
	}
}

func TestBrowser_ResolveUrl_PicksTheScraperForTheTarget(t *testing.T) {
	const (
		prv = "http://scraper-prv:8080"
		pub = "http://scraper-pub:8080"
	)

	tests := []struct {
		name      string
		scPrvURL  string
		scPubURL  string
		targetURL string
		want      string
	}{
		{
			name:      "a private ip goes to the private scraper",
			scPrvURL:  prv,
			scPubURL:  pub,
			targetURL: "http://192.168.1.1/test",
			want:      prv,
		},
		{
			name:      "a private ip with a port goes to the private scraper",
			scPrvURL:  prv,
			scPubURL:  pub,
			targetURL: "http://10.1.2.3:8000",
			want:      prv,
		},
		{
			name:      "a loopback ip goes to the private scraper",
			scPrvURL:  prv,
			scPubURL:  pub,
			targetURL: "http://127.0.0.1:8080",
			want:      prv,
		},
		{
			name:      "localhost goes to the private scraper",
			scPrvURL:  prv,
			scPubURL:  pub,
			targetURL: "http://localhost:8080",
			want:      prv,
		},
		{
			name:      "a host in a local zone goes to the private scraper",
			scPrvURL:  prv,
			scPubURL:  pub,
			targetURL: "http://target.htb",
			want:      prv,
		},
		{
			// The empty label fails the lookup at once; a valid .local name waits 5 s on mDNS on macOS.
			name:      "a host in the .local zone goes to the private scraper",
			scPrvURL:  prv,
			scPubURL:  pub,
			targetURL: "http://dc01..corp.local/",
			want:      prv,
		},
		{
			name:      "a single-label host nothing resolves goes to the private scraper",
			scPrvURL:  prv,
			scPubURL:  pub,
			targetURL: "http://pentagi-no-such-host:8080",
			want:      prv,
		},
		{
			name:      "a public host goes to the public scraper",
			scPrvURL:  prv,
			scPubURL:  pub,
			targetURL: "https://google.com",
			want:      pub,
		},
		{
			name:      "only a private scraper, a private target",
			scPrvURL:  prv,
			targetURL: "http://localhost:3000",
			want:      prv,
		},
		{
			name:      "only a private scraper, a public target falls back to it",
			scPrvURL:  prv,
			targetURL: "https://example.com",
			want:      prv,
		},
		{
			name:      "only a public scraper, a public target",
			scPubURL:  pub,
			targetURL: "https://google.com",
			want:      pub,
		},
		{
			name:      "only a public scraper, a private target falls back to it",
			scPubURL:  pub,
			targetURL: "http://10.0.0.1",
			want:      pub,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &browser{scPrvURL: tt.scPrvURL, scPubURL: tt.scPubURL}

			got, err := b.resolveUrl(tt.targetURL)
			if err != nil {
				t.Fatalf("resolveUrl() error = %v", err)
			}
			if scraper := got.Scheme + "://" + got.Host; scraper != tt.want {
				t.Errorf("resolveUrl() = %v, want %v", scraper, tt.want)
			}
		})
	}

	t.Run("no scraper at all", func(t *testing.T) {
		b := &browser{}
		_, err := b.resolveUrl("https://example.com")
		if err == nil || !strings.Contains(err.Error(), "no scraper URL configured") {
			t.Fatalf("resolveUrl() error = %v, want no scraper URL configured", err)
		}
	})
}

func TestBrowser_IsAvailable_NeedsAScraperURL(t *testing.T) {
	tests := []struct {
		name     string
		scPrvURL string
		scPubURL string
		want     bool
	}{
		{
			name:     "both urls set",
			scPrvURL: "http://scraper-prv:8080",
			scPubURL: "http://scraper-pub:8080",
			want:     true,
		},
		{
			name:     "only the private url set",
			scPrvURL: "http://scraper-prv:8080",
			want:     true,
		},
		{
			name:     "only the public url set",
			scPubURL: "http://scraper-pub:8080",
			want:     true,
		},
		{
			name: "no url set",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &browser{scPrvURL: tt.scPrvURL, scPubURL: tt.scPubURL}
			if got := b.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// The extensions are listed by hand so that dropping one from nonHTMLExtensions fails a case.
func TestBrowser_IsBinaryURL_RecognizesANonHTMLResource(t *testing.T) {
	type binaryURLCase struct {
		name string
		url  string
		want bool
	}

	var tests []binaryURLCase
	for _, ext := range []string{
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".zip", ".tar", ".gz", ".bz2", ".rar", ".7z",
		".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp", ".svg", ".ico",
		".mp3", ".mp4", ".avi", ".mov", ".mkv", ".wav",
		".exe", ".bin", ".dll", ".so", ".dmg", ".apk",
	} {
		tests = append(tests, binaryURLCase{
			name: "a file ending in " + ext,
			url:  "https://example.com/files/download" + ext,
			want: true,
		})
	}
	tests = append(tests,
		binaryURLCase{"an upper-case extension", "https://example.com/REPORT.PDF", true},
		binaryURLCase{"an extension followed by a query", "https://example.com/file.pdf?token=abc&download=1", true},
		binaryURLCase{"a path without an extension", "https://example.com/page", false},
		binaryURLCase{"an html page", "https://example.com/report.html", false},
		binaryURLCase{"a json answer", "https://example.com/api/data.json", false},
		binaryURLCase{"an extension inside the path, not at its end", "https://example.com/pdf-guide/intro", false},
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBinaryURL(tt.url); got != tt.want {
				t.Errorf("isBinaryURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}
