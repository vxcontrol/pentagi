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

func TestGetHTML_SmallContent_ReturnsWarning(t *testing.T) {
	// Serve HTML content that is larger than minMdContentSize (50)
	// but smaller than minHtmlContentSize (300). With the updated behaviour,
	// getHTML must return the content with a WARNING prefix instead of an error.
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

	result, err := b.getHTML("https://example.com/page")
	if err != nil {
		t.Fatalf("getHTML() should not error for small-but-non-empty content, got: %v", err)
	}
	if !strings.Contains(result, "[WARNING:") {
		t.Errorf("getHTML() should prefix result with WARNING for small content, got: %q", result)
	}
	if !strings.Contains(result, smallHTML) {
		t.Errorf("getHTML() should include the actual content in the result")
	}
}

func TestGetHTML_EmptyContent_ReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// intentionally write nothing — callScraper detects empty body and errors
	}))
	defer ts.Close()

	b := &browser{flowID: 1, scPubURL: ts.URL}

	_, err := b.getHTML("https://example.com/page")
	if err == nil {
		t.Fatal("getHTML() should error for empty content")
	}
	// callScraper wraps the error as "failed to fetch content by url '...': empty response body..."
	if !strings.Contains(err.Error(), "failed to fetch content") {
		t.Errorf("getHTML() error should come from callScraper, got: %v", err)
	}
}

func TestGetHTML_BinaryURL_ReturnsError(t *testing.T) {
	b := &browser{flowID: 1, scPubURL: "http://127.0.0.1:1"}

	_, err := b.getHTML("https://example.com/report.pdf")
	if err == nil {
		t.Fatal("getHTML() should error for binary URL")
	}
	if !strings.Contains(err.Error(), "binary") {
		t.Errorf("getHTML() error should mention binary resource, got: %v", err)
	}
	if !strings.Contains(err.Error(), "curl") {
		t.Errorf("getHTML() error should hint about curl/wget, got: %v", err)
	}
}

func TestGetMD_SmallContent_ReturnsWarning(t *testing.T) {
	tiny := strings.Repeat("x", minMdContentSize-1) // just below minimum, but not 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, tiny)
	}))
	defer ts.Close()

	b := &browser{flowID: 1, scPubURL: ts.URL}

	result, err := b.getMD("https://example.com/page")
	if err != nil {
		t.Fatalf("getMD() should not error for small-but-non-empty content, got: %v", err)
	}
	if !strings.Contains(result, "[WARNING:") {
		t.Errorf("getMD() should prefix result with WARNING for small content, got: %q", result)
	}
	if !strings.Contains(result, tiny) {
		t.Errorf("getMD() should include the actual content in the result")
	}
}

func TestGetMD_EmptyContent_ReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// intentionally write nothing — callScraper detects empty body and errors
	}))
	defer ts.Close()

	b := &browser{flowID: 1, scPubURL: ts.URL}

	_, err := b.getMD("https://example.com/page")
	if err == nil {
		t.Fatal("getMD() should error for empty content")
	}
	if !strings.Contains(err.Error(), "failed to fetch content") {
		t.Errorf("getMD() error should come from callScraper, got: %v", err)
	}
}

func TestGetMD_BinaryURL_ReturnsError(t *testing.T) {
	b := &browser{flowID: 1, scPubURL: "http://127.0.0.1:1"}

	for _, url := range []string{
		"https://example.com/file.pdf",
		"https://host.io/archive.zip",
		"https://cdn.example.com/image.png",
		"https://example.com/data.xlsx?token=abc",
	} {
		_, err := b.getMD(url)
		if err == nil {
			t.Errorf("getMD(%q) should error for binary URL", url)
			continue
		}
		if !strings.Contains(err.Error(), "binary") {
			t.Errorf("getMD(%q) error should mention binary resource, got: %v", url, err)
		}
	}
}

func TestIsBinaryURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://example.com/report.pdf", true},
		{"https://example.com/REPORT.PDF", true},
		{"https://cdn.example.com/image.png", true},
		{"https://cdn.example.com/photo.jpg", true},
		{"https://cdn.example.com/photo.JPEG", true},
		{"https://example.com/archive.zip", true},
		{"https://example.com/data.xlsx", true},
		{"https://example.com/slides.pptx", true},
		{"https://example.com/file.exe", true},
		{"https://example.com/lib.so", true},
		{"https://example.com/video.mp4", true},
		// query string is stripped before extension check
		{"https://example.com/file.pdf?token=abc&download=1", true},
		// non-binary URLs
		{"https://example.com/page", false},
		{"https://example.com/report.html", false},
		{"https://example.com/api/data.json", false},
		{"https://example.com/path/to/index", false},
		// pdf in path component but not as suffix
		{"https://example.com/pdf-guide/intro", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := isBinaryURL(tt.url)
			if got != tt.want {
				t.Errorf("isBinaryURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
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
