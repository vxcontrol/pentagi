package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

const (
	duckduckgoMaxResults = 10
	duckduckgoMaxRetries = 3
	duckduckgoSearchURL  = "https://html.duckduckgo.com/html/"
	duckduckgoTimeout    = 30 * time.Second
	duckduckgoUserAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// Region constants for DuckDuckGo search
const (
	RegionUS = "us-en" // USA
	RegionUK = "uk-en" // United Kingdom
	RegionDE = "de-de" // Germany
	RegionFR = "fr-fr" // France
	RegionJP = "jp-jp" // Japan
	RegionCN = "cn-zh" // China
	RegionRU = "ru-ru" // Russia
)

// Safe search levels for DuckDuckGo
const (
	DuckDuckGoSafeSearchStrict   = "strict"   // Strict filtering
	DuckDuckGoSafeSearchModerate = "moderate" // Moderate filtering
	DuckDuckGoSafeSearchOff      = "off"      // No filtering
)

// Time range constants for DuckDuckGo search
const (
	TimeRangeDay   = "d" // Day
	TimeRangeWeek  = "w" // Week
	TimeRangeMonth = "m" // Month
	TimeRangeYear  = "y" // Year
)

// searchResult represents a single search result from DuckDuckGo
type searchResult struct {
	Title       string `json:"t"`
	URL         string `json:"u"`
	Description string `json:"a"`
}

// searchResponse represents the response from DuckDuckGo search API
type searchResponse struct {
	Results   []searchResult `json:"results"`
	NoResults bool           `json:"noResults"`
}

type duckduckgo struct {
	flowID     int64
	taskID     *int64
	subtaskID  *int64
	enabled    bool
	proxyURL   string
	region     string
	safeSearch string
	timeRange  string
	slp        SearchLogProvider
}

func NewDuckDuckGoTool(flowID int64, taskID, subtaskID *int64, enabled bool,
	proxyURL, region, safeSearch, timeRange string, slp SearchLogProvider,
) Tool {
	return &duckduckgo{
		flowID:     flowID,
		taskID:     taskID,
		subtaskID:  subtaskID,
		enabled:    enabled,
		proxyURL:   proxyURL,
		region:     region,
		safeSearch: safeSearch,
		timeRange:  timeRange,
		slp:        slp,
	}
}

// Handle processes the search request from an AI agent
func (d *duckduckgo) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	var action SearchAction
	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(d.flowID, d.taskID, d.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	if err := json.Unmarshal(args, &action); err != nil {
		logger.WithError(err).Error("failed to unmarshal duckduckgo search action")
		return "", fmt.Errorf("failed to unmarshal %s search action arguments: %w", name, err)
	}

	// Set default number of results if invalid
	numResults := int(action.MaxResults)
	if numResults < 1 || numResults > duckduckgoMaxResults {
		numResults = duckduckgoMaxResults
	}

	logger = logger.WithFields(logrus.Fields{
		"query":       action.Query[:min(len(action.Query), 1000)],
		"num_results": numResults,
		"region":      d.region,
	})

	// Perform search
	result, err := d.search(ctx, action.Query, numResults)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("search engine error swallowed"),
			langfuse.WithEventInput(action.Query),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"tool_name":   DuckDuckGoToolName,
				"engine":      "duckduckgo",
				"query":       action.Query,
				"max_results": numResults,
				"region":      d.region,
				"error":       err.Error(),
			}),
		)

		logger.WithError(err).Error("failed to search in DuckDuckGo")
		return fmt.Sprintf("failed to search in DuckDuckGo: %v", err), nil
	}

	// Log search results if configured
	if agentCtx, ok := GetAgentContext(ctx); ok {
		_, _ = d.slp.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			database.SearchengineTypeDuckduckgo,
			action.Query,
			result,
			d.taskID,
			d.subtaskID,
		)
	}

	return result, nil
}

// search performs a web search using DuckDuckGo
func (d *duckduckgo) search(ctx context.Context, query string, maxResults int) (string, error) {
	// Build form data for POST request
	formData := d.buildFormData(query)

	// Create HTTP client with proper configuration
	client := d.createHTTPClient()

	// Execute request with retry logic
	var response *searchResponse
	for attempt := 0; attempt < duckduckgoMaxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", duckduckgoSearchURL, strings.NewReader(formData))
		if err != nil {
			return "", fmt.Errorf("failed to create search request: %w", err)
		}

		// Add necessary headers for POST request
		req.Header.Set("User-Agent", duckduckgoUserAgent)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		resp, err := client.Do(req)
		if err != nil {
			if attempt == duckduckgoMaxRetries-1 {
				return "", fmt.Errorf("failed to execute search after %d attempts: %w", duckduckgoMaxRetries, err)
			}
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(time.Second):
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			if attempt == duckduckgoMaxRetries-1 {
				return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			}
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(time.Second):
			}
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %w", err)
		}

		response, err = d.parseHTMLResponse(body)
		if err != nil {
			return "", fmt.Errorf("failed to parse search response: %w", err)
		}

		break
	}

	if response == nil || len(response.Results) == 0 {
		return "No results found", nil
	}

	// Limit results to requested number
	if len(response.Results) > maxResults {
		response.Results = response.Results[:maxResults]
	}

	// Format results in readable text format
	return d.formatSearchResults(response.Results), nil
}

// buildFormData creates form data for DuckDuckGo POST request
func (d *duckduckgo) buildFormData(query string) string {
	params := url.Values{}
	params.Set("q", query)
	params.Set("b", "")
	params.Set("df", "")

	if d.region != "" {
		params.Set("kl", d.region)
	} else {
		params.Set("kl", RegionUS)
	}

	if d.safeSearch != "" {
		switch d.safeSearch {
		case DuckDuckGoSafeSearchStrict:
			params.Set("kp", "1")
		case DuckDuckGoSafeSearchModerate:
			params.Set("kp", "0")
		case DuckDuckGoSafeSearchOff:
			params.Set("kp", "-1")
		}
	}

	if d.timeRange != "" {
		params.Set("df", d.timeRange)
	}

	return params.Encode()
}

// parseHTMLResponse parses the HTML search response from DuckDuckGo
func (d *duckduckgo) parseHTMLResponse(body []byte) (*searchResponse, error) {
	// Try structured HTML parsing first
	results, err := d.parseHTMLStructured(body)
	if err == nil && len(results) > 0 {
		return &searchResponse{
			Results:   results,
			NoResults: false,
		}, nil
	}

	// Fallback to regex-based parsing
	results, err = d.parseHTMLRegex(body)
	if err != nil {
		return nil, err
	}

	return &searchResponse{
		Results:   results,
		NoResults: len(results) == 0,
	}, nil
}

// parseHTMLStructured uses golang.org/x/net/html for structured HTML parsing
func (d *duckduckgo) parseHTMLStructured(body []byte) ([]searchResult, error) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	results := make([]searchResult, 0)
	d.findResultNodes(doc, &results)

	return results, nil
}

// findResultNodes recursively finds and extracts search result nodes
func (d *duckduckgo) findResultNodes(n *html.Node, results *[]searchResult) {
	// Look for div with class "result results_links"
	if n.Type == html.ElementNode && n.Data == "div" {
		if d.hasClass(n, "result") && d.hasClass(n, "results_links") {
			result := d.extractResultFromNode(n)
			if result.Title != "" && result.URL != "" {
				*results = append(*results, result)
			}
		}
	}

	// Recurse through children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		d.findResultNodes(c, results)
	}
}

// extractResultFromNode extracts title, URL, and description from a result node
func (d *duckduckgo) extractResultFromNode(n *html.Node) searchResult {
	result := searchResult{}

	// Find title link (a.result__a)
	d.findElement(n, func(node *html.Node) bool {
		if node.Type == html.ElementNode && node.Data == "a" && d.hasClass(node, "result__a") {
			result.URL = d.getAttr(node, "href")
			result.Title = d.getTextContent(node)
			return true
		}
		return false
	})

	// Find snippet (a.result__snippet)
	d.findElement(n, func(node *html.Node) bool {
		if node.Type == html.ElementNode && node.Data == "a" && d.hasClass(node, "result__snippet") {
			result.Description = d.getTextContent(node)
			return true
		}
		return false
	})

	// Clean text
	result.Title = d.cleanText(result.Title)
	result.Description = d.cleanText(result.Description)

	return result
}

// findElement finds the first element matching the predicate
func (d *duckduckgo) findElement(n *html.Node, predicate func(*html.Node) bool) bool {
	if predicate(n) {
		return true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if d.findElement(c, predicate) {
			return true
		}
	}

	return false
}

// hasClass checks if a node has a specific CSS class
func (d *duckduckgo) hasClass(n *html.Node, className string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			classes := strings.Fields(attr.Val)
			for _, c := range classes {
				if c == className {
					return true
				}
			}
		}
	}
	return false
}

// getAttr gets an attribute value from a node
func (d *duckduckgo) getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// getTextContent extracts all text content from a node and its children
func (d *duckduckgo) getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	var text strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text.WriteString(d.getTextContent(c))
	}

	return text.String()
}

// parseHTMLRegex is a fallback regex-based parser
func (d *duckduckgo) parseHTMLRegex(body []byte) ([]searchResult, error) {
	htmlStr := string(body)

	// Check for "no results" message
	if strings.Contains(htmlStr, "No results found") || strings.Contains(htmlStr, "noResults") {
		return []searchResult{}, nil
	}

	results := make([]searchResult, 0)

	// Pattern to find result blocks (web results)
	// Each block starts with <div class="result results_links..."> and ends with <div class="clear"></div>
	// followed by closing tags </div></div>
	resultPattern := regexp.MustCompile(`(?s)<div class="result results_links[^"]*">.*?<div class="clear"></div>\s*</div>\s*</div>`)
	resultBlocks := resultPattern.FindAllString(htmlStr, -1)

	// Extract title, URL, and description from each result block
	titlePattern := regexp.MustCompile(`<a[^>]+class="result__a"[^>]+href="([^"]+)"[^>]*>([^<]+)</a>`)
	snippetPattern := regexp.MustCompile(`(?s)<a[^>]+class="result__snippet"[^>]+href="[^"]*">(.+?)</a>`)

	for _, block := range resultBlocks {
		// Extract title and URL
		titleMatches := titlePattern.FindStringSubmatch(block)
		if len(titleMatches) < 3 {
			continue
		}

		resultURL := titleMatches[1]
		title := d.cleanText(titleMatches[2])

		// Extract description
		description := ""
		snippetMatches := snippetPattern.FindStringSubmatch(block)
		if len(snippetMatches) > 1 {
			description = d.cleanText(snippetMatches[1])
		}

		if title == "" || resultURL == "" {
			continue
		}

		results = append(results, searchResult{
			Title:       title,
			URL:         resultURL,
			Description: description,
		})
	}

	return results, nil
}

// cleanText removes HTML tags and decodes HTML entities
func (d *duckduckgo) cleanText(text string) string {
	// Remove HTML tags (like <b>, </b>, etc.)
	re := regexp.MustCompile(`<[^>]*>`)
	text = re.ReplaceAllString(text, "")

	// Decode common HTML entities
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&#x27;", "'")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&apos;", "'")

	// Decode hex HTML entities (&#xNN;)
	hexEntityRe := regexp.MustCompile(`&#x([0-9A-Fa-f]+);`)
	text = hexEntityRe.ReplaceAllStringFunc(text, func(match string) string {
		// Extract hex value
		hex := hexEntityRe.FindStringSubmatch(match)
		if len(hex) > 1 {
			var codePoint int
			_, err := fmt.Sscanf(hex[1], "%x", &codePoint)
			if err == nil && codePoint < 128 {
				return string(rune(codePoint))
			}
		}
		return match
	})

	// Decode decimal HTML entities (&#NNN;)
	decEntityRe := regexp.MustCompile(`&#([0-9]+);`)
	text = decEntityRe.ReplaceAllStringFunc(text, func(match string) string {
		dec := decEntityRe.FindStringSubmatch(match)
		if len(dec) > 1 {
			var codePoint int
			_, err := fmt.Sscanf(dec[1], "%d", &codePoint)
			if err == nil && codePoint < 128 {
				return string(rune(codePoint))
			}
		}
		return match
	})

	// Trim whitespace and normalize spaces
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	return text
}

// formatSearchResults formats search results in a readable text format
func (d *duckduckgo) formatSearchResults(results []searchResult) string {
	var builder strings.Builder

	for i, result := range results {
		builder.WriteString(fmt.Sprintf("# %d. %s\n\n", i+1, result.Title))
		builder.WriteString(fmt.Sprintf("## URL\n%s\n\n", result.URL))
		builder.WriteString(fmt.Sprintf("## Description\n\n%s\n\n", result.Description))

		if i < len(results)-1 {
			builder.WriteString("---\n\n")
		}
	}

	return builder.String()
}

// createHTTPClient creates an HTTP client with configured proxy and timeout
func (d *duckduckgo) createHTTPClient() *http.Client {
	client := &http.Client{
		Timeout: duckduckgoTimeout,
	}

	// Configure proxy if specified
	if d.proxyURL != "" {
		proxyURL, err := url.Parse(d.proxyURL)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}

	return client
}

// isAvailable checks if the DuckDuckGo search client is properly configured
func (d *duckduckgo) IsAvailable() bool {
	// DuckDuckGo is a free search engine that doesn't require API keys or additional configuration.
	// We only need to check if it's enabled in the settings according to the user config.
	return d.enabled
}
