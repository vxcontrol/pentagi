package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type screenshotProviderMock struct {
	mu        sync.Mutex
	calls     int
	lastName  string
	lastURL   string
	lastTask  *int64
	lastSub   *int64
	returnErr error
}

func (m *screenshotProviderMock) PutScreenshot(_ context.Context, name, url string, taskID, subtaskID *int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls++
	m.lastName = name
	m.lastURL = url
	m.lastTask = taskID
	m.lastSub = subtaskID
	return 1, m.returnErr
}

func TestBrowserResolveUrl(t *testing.T) {
	tests := []struct {
		name      string
		scPrvURL  string
		scPubURL  string
		targetURL string
		wantURL   string
		wantErr   bool
	}{
		{
			name:      "both URLs set, private target uses private",
			scPrvURL:  "http://scraper-prv:8080",
			scPubURL:  "http://scraper-pub:8080",
			targetURL: "http://192.168.1.1/test",
			wantURL:   "http://scraper-prv:8080",
			wantErr:   false,
		},
		{
			name:      "both URLs set, public target uses public",
			scPrvURL:  "http://scraper-prv:8080",
			scPubURL:  "http://scraper-pub:8080",
			targetURL: "https://google.com",
			wantURL:   "http://scraper-pub:8080",
			wantErr:   false,
		},
		{
			name:      "only private URL set, private target uses private",
			scPrvURL:  "http://scraper-prv:8080",
			scPubURL:  "",
			targetURL: "http://localhost:3000",
			wantURL:   "http://scraper-prv:8080",
			wantErr:   false,
		},
		{
			name:      "only private URL set, public target falls back to private",
			scPrvURL:  "http://scraper-prv:8080",
			scPubURL:  "",
			targetURL: "https://example.com",
			wantURL:   "http://scraper-prv:8080",
			wantErr:   false,
		},
		{
			name:      "only public URL set, public target uses public",
			scPrvURL:  "",
			scPubURL:  "http://scraper-pub:8080",
			targetURL: "https://google.com",
			wantURL:   "http://scraper-pub:8080",
			wantErr:   false,
		},
		{
			name:      "only public URL set, private target falls back to public",
			scPrvURL:  "",
			scPubURL:  "http://scraper-pub:8080",
			targetURL: "http://10.0.0.1",
			wantURL:   "http://scraper-pub:8080",
			wantErr:   false,
		},
		{
			name:      "no URLs set, returns error",
			scPrvURL:  "",
			scPubURL:  "",
			targetURL: "https://example.com",
			wantURL:   "",
			wantErr:   true,
		},
		{
			name:      "localhost target uses private",
			scPrvURL:  "http://scraper-prv:8080",
			scPubURL:  "http://scraper-pub:8080",
			targetURL: "http://localhost:8080",
			wantURL:   "http://scraper-prv:8080",
			wantErr:   false,
		},
		{
			name:      "local zone target uses private",
			scPrvURL:  "http://scraper-prv:8080",
			scPubURL:  "http://scraper-pub:8080",
			targetURL: "http://myapp.local",
			wantURL:   "http://scraper-prv:8080",
			wantErr:   false,
		},
		{
			name:      "10.x.x.x private IP uses private",
			scPrvURL:  "http://scraper-prv:8080",
			scPubURL:  "http://scraper-pub:8080",
			targetURL: "http://10.1.2.3:8000",
			wantURL:   "http://scraper-prv:8080",
			wantErr:   false,
		},
		{
			name:      "172.16.x.x private IP uses private",
			scPrvURL:  "http://scraper-prv:8080",
			scPubURL:  "http://scraper-pub:8080",
			targetURL: "http://172.16.0.1",
			wantURL:   "http://scraper-prv:8080",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &browser{
				scPrvURL: tt.scPrvURL,
				scPubURL: tt.scPubURL,
			}

			gotURL, err := b.resolveUrl(tt.targetURL)

			if (err != nil) != tt.wantErr {
				t.Errorf("resolveUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				got := gotURL.Scheme + "://" + gotURL.Host
				if got != tt.wantURL {
					t.Errorf("resolveUrl() = %v, want %v", got, tt.wantURL)
				}
			}
		})
	}
}

func TestBrowserIsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		scPrvURL string
		scPubURL string
		want     bool
	}{
		{
			name:     "both URLs set",
			scPrvURL: "http://scraper-prv:8080",
			scPubURL: "http://scraper-pub:8080",
			want:     true,
		},
		{
			name:     "only private URL set",
			scPrvURL: "http://scraper-prv:8080",
			scPubURL: "",
			want:     true,
		},
		{
			name:     "only public URL set",
			scPrvURL: "",
			scPubURL: "http://scraper-pub:8080",
			want:     true,
		},
		{
			name:     "no URLs set",
			scPrvURL: "",
			scPubURL: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &browser{
				scPrvURL: tt.scPrvURL,
				scPubURL: tt.scPubURL,
			}

			if got := b.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// newTestScraper creates an httptest server that simulates the scraper service.
// screenshotBehavior controls the /screenshot endpoint: "ok", "fail", or "small".
func newTestScraper(t *testing.T, screenshotBehavior string) *httptest.Server {
	t.Helper()
	validMD := strings.Repeat("A", minMdContentSize+1)
	validHTML := strings.Repeat("<p>x</p>", minHtmlContentSize/8+1)
	validLinks := `[{"Title":"Example","Link":"https://example.com"}]`
	validScreenshot := strings.Repeat("\x89PNG", minImgContentSize/4+1) // exceeds minImgContentSize

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/markdown":
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, validMD)
		case "/html":
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, validHTML)
		case "/links":
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, validLinks)
		case "/screenshot":
			switch screenshotBehavior {
			case "ok":
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, validScreenshot)
			case "fail":
				w.WriteHeader(http.StatusInternalServerError)
			case "small":
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "tiny") // below minImgContentSize
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestContentMD_ScreenshotFailure_ReturnsContent(t *testing.T) {
	ts := newTestScraper(t, "fail")
	defer ts.Close()

	dataDir := t.TempDir()
	b := &browser{
		flowID:   1,
		dataDir:  dataDir,
		scPubURL: ts.URL,
	}

	content, screenshot, err := b.ContentMD(t.Context(), "https://example.com/page")
	if err != nil {
		t.Fatalf("ContentMD() returned unexpected error: %v", err)
	}
	if content == "" {
		t.Error("ContentMD() returned empty content despite successful fetch")
	}
	if screenshot != "" {
		t.Errorf("ContentMD() screenshot = %q, want empty string on failure", screenshot)
	}
}

func TestContentHTML_ScreenshotFailure_ReturnsContent(t *testing.T) {
	ts := newTestScraper(t, "fail")
	defer ts.Close()

	dataDir := t.TempDir()
	b := &browser{
		flowID:   1,
		dataDir:  dataDir,
		scPubURL: ts.URL,
	}

	content, screenshot, err := b.ContentHTML(t.Context(), "https://example.com/page")
	if err != nil {
		t.Fatalf("ContentHTML() returned unexpected error: %v", err)
	}
	if content == "" {
		t.Error("ContentHTML() returned empty content despite successful fetch")
	}
	if screenshot != "" {
		t.Errorf("ContentHTML() screenshot = %q, want empty string on failure", screenshot)
	}
}

func TestLinks_ScreenshotFailure_ReturnsContent(t *testing.T) {
	ts := newTestScraper(t, "fail")
	defer ts.Close()

	dataDir := t.TempDir()
	b := &browser{
		flowID:   1,
		dataDir:  dataDir,
		scPubURL: ts.URL,
	}

	links, screenshot, err := b.Links(t.Context(), "https://example.com/page")
	if err != nil {
		t.Fatalf("Links() returned unexpected error: %v", err)
	}
	if links == "" {
		t.Error("Links() returned empty content despite successful fetch")
	}
	if screenshot != "" {
		t.Errorf("Links() screenshot = %q, want empty string on failure", screenshot)
	}
}

func TestContentMD_ScreenshotSmall_ReturnsContent(t *testing.T) {
	ts := newTestScraper(t, "small")
	defer ts.Close()

	dataDir := t.TempDir()
	b := &browser{
		flowID:   1,
		dataDir:  dataDir,
		scPubURL: ts.URL,
	}

	content, screenshot, err := b.ContentMD(t.Context(), "https://example.com/page")
	if err != nil {
		t.Fatalf("ContentMD() returned unexpected error: %v", err)
	}
	if content == "" {
		t.Error("ContentMD() returned empty content when screenshot was too small")
	}
	if screenshot != "" {
		t.Errorf("ContentMD() screenshot = %q, want empty string for undersized image", screenshot)
	}
}

func TestContentMD_BothSucceed_ReturnsContentAndScreenshot(t *testing.T) {
	ts := newTestScraper(t, "ok")
	defer ts.Close()

	dataDir := t.TempDir()
	b := &browser{
		flowID:   1,
		dataDir:  dataDir,
		scPubURL: ts.URL,
	}

	content, screenshot, err := b.ContentMD(t.Context(), "https://example.com/page")
	if err != nil {
		t.Fatalf("ContentMD() returned unexpected error: %v", err)
	}
	if content == "" {
		t.Error("ContentMD() returned empty content")
	}
	if screenshot == "" {
		t.Error("ContentMD() returned empty screenshot when both should succeed")
	}
	// Verify screenshot file was written
	screenshotPath := filepath.Join(dataDir, "screenshots", "flow-1", screenshot)
	if _, err := os.Stat(screenshotPath); os.IsNotExist(err) {
		t.Errorf("screenshot file not written: %s", screenshotPath)
	}
}

func TestGetHTML_UsesCorrectMinContentSize(t *testing.T) {
	// Serve HTML content that is larger than minMdContentSize (50)
	// but smaller than minHtmlContentSize (300).
	// With the fix, getHTML should reject this; before the fix it would accept it.
	smallHTML := strings.Repeat("x", minMdContentSize+10) // 60 bytes: > 50, < 300

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, smallHTML)
	}))
	defer ts.Close()

	b := &browser{
		flowID:   1,
		scPubURL: ts.URL,
	}

	_, err := b.getHTML("https://example.com/page")
	if err == nil {
		t.Fatal("getHTML() should reject content smaller than minHtmlContentSize")
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("%d bytes", minHtmlContentSize)) {
		t.Errorf("getHTML() error should reference minHtmlContentSize (%d), got: %v", minHtmlContentSize, err)
	}
}

func TestBrowserHandle_ValidationErrors(t *testing.T) {
	b := &browser{
		scPrvURL: "http://127.0.0.1:8080",
	}

	t.Run("unknown tool", func(t *testing.T) {
		_, err := b.Handle(t.Context(), "not-browser", json.RawMessage(`{}`))
		if err == nil || !strings.Contains(err.Error(), "unknown tool") {
			t.Fatalf("expected unknown tool error, got: %v", err)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := b.Handle(t.Context(), "browser", json.RawMessage(`{`))
		if err == nil || !strings.Contains(err.Error(), "failed to unmarshal browser action") {
			t.Fatalf("expected unmarshal error, got: %v", err)
		}
	})

	t.Run("unknown action", func(t *testing.T) {
		_, err := b.Handle(t.Context(), "browser", json.RawMessage(`{"url":"https://example.com","action":"unknown","message":"m"}`))
		if err == nil || !strings.Contains(err.Error(), "unknown file action") {
			t.Fatalf("expected unknown action error, got: %v", err)
		}
	})
}

func TestBrowserHandle_MarkdownSuccess_StoresScreenshot(t *testing.T) {
	ts := newTestScraper(t, "ok")
	defer ts.Close()

	scp := &screenshotProviderMock{}
	dataDir := t.TempDir()
	b := &browser{
		flowID:   1,
		dataDir:  dataDir,
		scPubURL: ts.URL,
		scp:      scp,
	}

	result, err := b.Handle(t.Context(), "browser", json.RawMessage(`{"url":"https://example.com/page","action":"markdown","message":"m"}`))
	if err != nil {
		t.Fatalf("Handle() returned unexpected error: %v", err)
	}
	if result == "" {
		t.Fatal("Handle() returned empty markdown result")
	}

	scp.mu.Lock()
	defer scp.mu.Unlock()

	if scp.calls != 1 {
		t.Fatalf("PutScreenshot() calls = %d, want 1", scp.calls)
	}
	if scp.lastURL != "https://example.com/page" {
		t.Fatalf("PutScreenshot() url = %q, want %q", scp.lastURL, "https://example.com/page")
	}
	if scp.lastName == "" {
		t.Fatal("PutScreenshot() screenshot name should not be empty")
	}
}

func TestWrapCommandResult_ErrorIsSwallowed(t *testing.T) {
	scp := &screenshotProviderMock{}
	b := &browser{scp: scp}

	result, err := b.wrapCommandResult(
		t.Context(),
		"browser",
		"payload",
		"https://example.com",
		"screen.png",
		errors.New("boom"),
	)
	if err != nil {
		t.Fatalf("wrapCommandResult() returned unexpected error: %v", err)
	}
	if !strings.Contains(result, "handled with error") {
		t.Fatalf("wrapCommandResult() = %q, want handled with error message", result)
	}

	scp.mu.Lock()
	defer scp.mu.Unlock()
	if scp.calls != 0 {
		t.Fatalf("PutScreenshot() should not be called on error branch, got %d calls", scp.calls)
	}
}
