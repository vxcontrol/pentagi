package tools

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"

	"github.com/sirupsen/logrus"
)

const (
	minMdContentSize   = 50
	minHtmlContentSize = 300
	minImgContentSize  = 2048
)

// nonHTMLExtensions lists URL path suffixes that point to resources the scraper
// cannot render as text. When a URL matches, the browser tool returns a
// descriptive hint instead of a generic "content too small" error.
var nonHTMLExtensions = []string{
	".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
	".zip", ".tar", ".gz", ".bz2", ".rar", ".7z",
	".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp", ".svg", ".ico",
	".mp3", ".mp4", ".avi", ".mov", ".mkv", ".wav",
	".exe", ".bin", ".dll", ".so", ".dmg", ".apk",
}

// isBinaryURL returns true when the URL points to a known non-HTML resource.
func isBinaryURL(rawURL string) bool {
	lower := strings.ToLower(rawURL)
	// strip query string for extension matching
	if idx := strings.Index(lower, "?"); idx != -1 {
		lower = lower[:idx]
	}
	for _, ext := range nonHTMLExtensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

var localZones = []string{
	".localdomain",
	".local",
	".lan",
	".htb",
	".dev",
	".test",
	".corp",
	".example",
	".invalid",
	".internal",
	".home.arpa",
}

type browser struct {
	flowID    int64
	taskID    *int64
	subtaskID *int64
	dataDir   string
	scPrvURL  string
	scPubURL  string
	scp       ScreenshotProvider
}

func NewBrowserTool(
	flowID int64, taskID, subtaskID *int64,
	dataDir, scPrvURL, scPubURL string,
	scp ScreenshotProvider,
) Tool {
	return &browser{
		flowID:    flowID,
		taskID:    taskID,
		subtaskID: subtaskID,
		dataDir:   dataDir,
		scPrvURL:  scPrvURL,
		scPubURL:  scPubURL,
		scp:       scp,
	}
}

func (b *browser) wrapCommandResult(ctx context.Context, name, result, url, screen string, err error) (string, error) {
	ctx, observation := obs.Observer.NewObservation(ctx)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("browser tool error swallowed"),
			langfuse.WithEventInput(map[string]any{
				"url":    url,
				"action": name,
			}),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"tool_name": BrowserToolName,
				"url":       url,
				"screen":    screen,
				"error":     err.Error(),
			}),
		)

		logrus.WithContext(ctx).WithError(err).WithFields(enrichLogrusFields(b.flowID, b.taskID, b.subtaskID, logrus.Fields{
			"tool":   name,
			"url":    url,
			"screen": screen,
			"result": result[:min(len(result), 1000)],
		})).Error("browser tool failed")
		return fmt.Sprintf("browser tool '%s' handled with error: %v", name, err), nil
	}
	if screen != "" {
		_, _ = b.scp.PutScreenshot(ctx, screen, url, b.taskID, b.subtaskID)
	}
	return result, nil
}

func (b *browser) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if !b.IsAvailable() {
		return "", fmt.Errorf("browser is not available")
	}

	var action Browser
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(b.flowID, b.taskID, b.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	if name != "browser" {
		logger.Error("unknown tool")
		return "", fmt.Errorf("unknown tool: %s", name)
	}

	if err := json.Unmarshal(args, &action); err != nil {
		logger.WithError(err).Error("failed to unmarshal browser action")
		return "", fmt.Errorf("failed to unmarshal browser action: %w", err)
	}

	logger = logger.WithFields(logrus.Fields{
		"action": action.Action,
		"url":    action.Url,
	})

	switch action.Action {
	case Markdown:
		result, screen, err := b.ContentMD(ctx, action.Url)
		return b.wrapCommandResult(ctx, name, result, action.Url, screen, err)
	case HTML:
		result, screen, err := b.ContentHTML(ctx, action.Url)
		return b.wrapCommandResult(ctx, name, result, action.Url, screen, err)
	case Links:
		result, screen, err := b.Links(ctx, action.Url)
		return b.wrapCommandResult(ctx, name, result, action.Url, screen, err)
	default:
		logger.Error("unknown file action")
		return "", fmt.Errorf("unknown file action: %s", action.Action)
	}
}

func (b *browser) ContentMD(ctx context.Context, url string) (string, string, error) {
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(b.flowID, b.taskID, b.subtaskID, logrus.Fields{
		"tool":   "browser",
		"action": "markdown",
		"url":    url,
	}))
	logger.Debug("trying to get markdown content")

	var (
		wg                        sync.WaitGroup
		content, screenshotName   string
		errContent, errScreenshot error
	)
	wg.Add(2)

	go func() {
		defer wg.Done()
		content, errContent = b.getMD(url)
	}()

	go func() {
		defer wg.Done()
		screenshotName, errScreenshot = b.getScreenshot(url)
	}()

	wg.Wait()

	if errContent != nil {
		return "", "", errContent
	}
	if errScreenshot != nil {
		logger.WithError(errScreenshot).Warn("failed to capture screenshot, continuing without it")
		screenshotName = ""
	}

	return content, screenshotName, nil
}

func (b *browser) ContentHTML(ctx context.Context, url string) (string, string, error) {
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(b.flowID, b.taskID, b.subtaskID, logrus.Fields{
		"tool":   "browser",
		"action": "html",
		"url":    url,
	}))
	logger.Debug("trying to get HTML content")

	var (
		wg                        sync.WaitGroup
		content, screenshotName   string
		errContent, errScreenshot error
	)
	wg.Add(2)

	go func() {
		defer wg.Done()
		content, errContent = b.getHTML(url)
	}()

	go func() {
		defer wg.Done()
		screenshotName, errScreenshot = b.getScreenshot(url)
	}()

	wg.Wait()

	if errContent != nil {
		return "", "", errContent
	}
	if errScreenshot != nil {
		logger.WithError(errScreenshot).Warn("failed to capture screenshot, continuing without it")
		screenshotName = ""
	}

	return content, screenshotName, nil
}

func (b *browser) Links(ctx context.Context, url string) (string, string, error) {
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(b.flowID, b.taskID, b.subtaskID, logrus.Fields{
		"tool":   "browser",
		"action": "links",
		"url":    url,
	}))
	logger.Debug("trying to get links")

	var (
		wg                      sync.WaitGroup
		links, screenshotName   string
		errLinks, errScreenshot error
	)
	wg.Add(2)

	go func() {
		defer wg.Done()
		links, errLinks = b.getLinks(url)
	}()

	go func() {
		defer wg.Done()
		screenshotName, errScreenshot = b.getScreenshot(url)
	}()

	wg.Wait()

	if errLinks != nil {
		return "", "", errLinks
	}
	if errScreenshot != nil {
		logger.WithError(errScreenshot).Warn("failed to capture screenshot, continuing without it")
		screenshotName = ""
	}

	return links, screenshotName, nil
}

func (b *browser) resolveUrl(targetURL string) (*url.URL, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host
	}

	// determine if target is private or public
	isPrivate := false

	hostIP := net.ParseIP(host)
	if hostIP != nil {
		isPrivate = hostIP.IsPrivate() || hostIP.IsLoopback()
	} else {
		ip, err := net.ResolveIPAddr("ip", host)
		if err == nil {
			isPrivate = ip.IP.IsPrivate() || ip.IP.IsLoopback()
		} else {
			lowerHost := strings.ToLower(host)
			if strings.Contains(lowerHost, "localhost") || !strings.Contains(lowerHost, ".") {
				isPrivate = true
			} else {
				for _, zone := range localZones {
					if strings.HasSuffix(lowerHost, zone) {
						isPrivate = true
						break
					}
				}
			}
		}
	}

	// select appropriate scraper URL with fallback
	var scraperURL string
	if isPrivate {
		scraperURL = b.scPrvURL
		if scraperURL == "" {
			scraperURL = b.scPubURL
		}
	} else {
		scraperURL = b.scPubURL
		if scraperURL == "" {
			scraperURL = b.scPrvURL
		}
	}

	if scraperURL == "" {
		return nil, fmt.Errorf("no scraper URL configured")
	}

	return url.Parse(scraperURL)
}

func (b *browser) saveScreenshotData(screenshot []byte) (string, error) {
	flowDirName := fmt.Sprintf("flow-%d", b.flowID)
	screenshotDir := filepath.Join(b.dataDir, "screenshots", flowDirName)

	err := os.MkdirAll(screenshotDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to prepare screenshot directory: %w", err)
	}

	screenshotName := fmt.Sprintf("screenshot-%d.png", time.Now().Unix())
	path := filepath.Join(screenshotDir, screenshotName)

	err = os.WriteFile(path, screenshot, 0644)
	if err != nil {
		return "", fmt.Errorf("screenshot write operation failed: %w", err)
	}

	return screenshotName, nil
}

func (b *browser) getMD(targetURL string) (string, error) {
	if isBinaryURL(targetURL) {
		return "", fmt.Errorf(
			"the URL appears to point to a binary/non-HTML resource (e.g. PDF, image, archive) " +
				"that cannot be rendered as markdown. Use the terminal tool with curl/wget to download it instead",
		)
	}

	scraperURL, err := b.resolveUrl(targetURL)
	if err != nil {
		return "", fmt.Errorf("failed to resolve url: %w", err)
	}

	query := scraperURL.Query()
	query.Add("url", targetURL)
	scraperURL.Path = "/markdown"
	scraperURL.RawQuery = query.Encode()

	content, err := b.callScraper(scraperURL.String())
	if err != nil {
		return "", fmt.Errorf("failed to fetch content by url '%s': %w", targetURL, err)
	}
	if len(content) < minMdContentSize {
		return fmt.Sprintf(
			"[WARNING: page returned very little content (%d bytes), it may be a redirect, error page, or near-empty]\n\n%s",
			len(content), string(content),
		), nil
	}

	return string(content), nil
}

func (b *browser) getHTML(targetURL string) (string, error) {
	if isBinaryURL(targetURL) {
		return "", fmt.Errorf(
			"the URL appears to point to a binary/non-HTML resource (e.g. PDF, image, archive) " +
				"that cannot be rendered as HTML. Use the terminal tool with curl/wget to download it instead",
		)
	}

	scraperURL, err := b.resolveUrl(targetURL)
	if err != nil {
		return "", fmt.Errorf("failed to resolve url: %w", err)
	}

	query := scraperURL.Query()
	query.Add("url", targetURL)
	scraperURL.Path = "/html"
	scraperURL.RawQuery = query.Encode()

	content, err := b.callScraper(scraperURL.String())
	if err != nil {
		return "", fmt.Errorf("failed to fetch content by url '%s': %w", targetURL, err)
	}
	if len(content) < minHtmlContentSize {
		return fmt.Sprintf(
			"[WARNING: page returned very little HTML content (%d bytes), it may be a redirect, error page, or near-empty]\n\n%s",
			len(content), string(content),
		), nil
	}

	return string(content), nil
}

func (b *browser) getLinks(targetURL string) (string, error) {
	scraperURL, err := b.resolveUrl(targetURL)
	if err != nil {
		return "", fmt.Errorf("failed to resolve url: %w", err)
	}

	query := scraperURL.Query()
	query.Add("url", targetURL)
	scraperURL.Path = "/links"
	scraperURL.RawQuery = query.Encode()

	content, err := b.callScraper(scraperURL.String())
	if err != nil {
		return "", fmt.Errorf("failed to fetch links by url '%s': %w", targetURL, err)
	}

	links := []struct {
		Title string
		Link  string
	}{}
	err = json.Unmarshal(content, &links)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal links: %w", err)
	}

	var buffer strings.Builder
	buffer.WriteString(fmt.Sprintf("Links list from URL '%s'\n", targetURL))
	for _, l := range links {
		link := strings.TrimSpace(l.Link)
		if link == "" {
			continue
		}
		title := strings.TrimSpace(l.Title)
		if title == "" {
			title = "UNTITLED"
		}
		buffer.WriteString(fmt.Sprintf("[%s](%s)\n", title, l.Link))
	}

	return buffer.String(), nil
}

func (b *browser) getScreenshot(targetURL string) (string, error) {
	scraperURL, err := b.resolveUrl(targetURL)
	if err != nil {
		return "", fmt.Errorf("failed to resolve url: %w", err)
	}

	query := scraperURL.Query()
	query.Add("fullPage", "true")
	query.Add("url", targetURL)
	scraperURL.Path = "/screenshot"
	scraperURL.RawQuery = query.Encode()

	content, err := b.callScraper(scraperURL.String())
	if err != nil {
		return "", fmt.Errorf("failed to fetch screenshot by url '%s': %w", targetURL, err)
	}
	if len(content) < minImgContentSize {
		return "", fmt.Errorf("image size is less than minimum: %d bytes", minImgContentSize)
	}

	return b.saveScreenshotData(content)
}

func (b *browser) callScraper(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 65 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data by scraper '%s': %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected resp code for scraper '%s': %d", url, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body for scraper '%s': %w", url, err)
	} else if len(content) == 0 {
		return nil, fmt.Errorf("empty response body for scraper '%s'", url)
	}

	return content, nil
}

func (b *browser) IsAvailable() bool {
	return b.scPrvURL != "" || b.scPubURL != ""
}
