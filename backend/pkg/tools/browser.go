package tools

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

func NewBrowserTool(flowID int64, taskID, subtaskID *int64, dataDir, scPrvURL, scPubURL string, scp ScreenshotProvider) Tool {
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

		logrus.WithContext(ctx).WithError(err).WithFields(logrus.Fields{
			"tool":   name,
			"url":    url,
			"screen": screen,
			"result": result[:min(len(result), 1000)],
		}).Error("browser tool failed")
		return fmt.Sprintf("browser tool '%s' handled with error: %v", name, err), nil
	}
	_, _ = b.scp.PutScreenshot(ctx, screen, url, b.taskID, b.subtaskID)
	return result, nil
}

func (b *browser) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	var action Browser
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"tool": name,
		"args": string(args),
	})

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
		result, screen, err := b.ContentMD(action.Url)
		return b.wrapCommandResult(ctx, name, result, action.Url, screen, err)
	case HTML:
		result, screen, err := b.ContentHTML(action.Url)
		return b.wrapCommandResult(ctx, name, result, action.Url, screen, err)
	case Links:
		result, screen, err := b.Links(action.Url)
		return b.wrapCommandResult(ctx, name, result, action.Url, screen, err)
	default:
		logger.Error("unknown file action")
		return "", fmt.Errorf("unknown file action: %s", action.Action)
	}
}

func (b *browser) ContentMD(url string) (string, string, error) {
	log.Println("Trying to get content from", url)

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
		return "", "", errScreenshot
	}

	return content, screenshotName, nil
}

func (b *browser) ContentHTML(url string) (string, string, error) {
	log.Println("Trying to get content from", url)

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
		return "", "", errScreenshot
	}

	return content, screenshotName, nil
}

func (b *browser) Links(url string) (string, string, error) {
	log.Println("Trying to get urls from", url)

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
		return "", "", errScreenshot
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

func (b *browser) writeScreenshotToFile(screenshot []byte) (string, error) {
	// Write screenshot to file
	flowDirName := fmt.Sprintf("flow-%d", b.flowID)
	err := os.MkdirAll(filepath.Join(b.dataDir, "screenshots", flowDirName), os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("error creating directory: %w", err)
	}

	screenshotName := fmt.Sprintf("%s.png", time.Now().Format("2006-01-02-15-04-05"))
	path := filepath.Join(b.dataDir, "screenshots", flowDirName, screenshotName)

	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("error creating file: %w", err)
	}

	defer file.Close()

	_, err = file.Write(screenshot)
	if err != nil {
		return "", fmt.Errorf("error writing to file: %w", err)
	}

	return screenshotName, nil
}

func (b *browser) getMD(targetURL string) (string, error) {
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
		return "", fmt.Errorf("content size is less than minimum: %d bytes", minMdContentSize)
	}

	return string(content), nil
}

func (b *browser) getHTML(targetURL string) (string, error) {
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
	if len(content) < minMdContentSize {
		return "", fmt.Errorf("content size is less than minimum: %d bytes", minMdContentSize)
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

	return b.writeScreenshotToFile(content)
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
