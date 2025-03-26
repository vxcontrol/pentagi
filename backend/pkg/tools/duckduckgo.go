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

	"github.com/sirupsen/logrus"
)

const (
	duckduckgoMaxResults = 10
	duckduckgoMaxRetries = 3
	duckduckgoBaseURL    = "https://duckduckgo.com"
	duckduckgoSearchURL  = "https://links.duckduckgo.com/d.js"
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
	proxyURL   string
	region     string
	safeSearch string
	timeRange  string
	slp        SearchLogProvider
}

func NewDuckDuckGoTool(flowID int64, taskID, subtaskID *int64,
	proxyURL, region, safeSearch, timeRange string, slp SearchLogProvider,
) Tool {
	return &duckduckgo{
		flowID:     flowID,
		taskID:     taskID,
		subtaskID:  subtaskID,
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
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"tool": name,
		"args": string(args),
	})

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
	// Get VQD token required for DuckDuckGo search
	vqd, err := d.getVQDToken(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to get VQD token: %w", err)
	}

	// Build search URL with all parameters
	searchURL := d.buildSearchURL(query, vqd, 1)

	// Create HTTP client with proper configuration
	client := d.createHTTPClient()
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create search request: %w", err)
	}

	// Add necessary headers
	req.Header.Set("User-Agent", duckduckgoUserAgent)
	req.Header.Set("Accept", "application/json")

	// Execute request with retry logic
	var response *searchResponse
	for attempt := 0; attempt < duckduckgoMaxRetries; attempt++ {
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
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
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
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %w", err)
		}

		response, err = d.parseSearchResponse(body)
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

// getVQDToken retrieves the VQD token required for DuckDuckGo search queries
func (d *duckduckgo) getVQDToken(ctx context.Context, query string) (string, error) {
	client := d.createHTTPClient()

	// Create URL with query parameter
	reqURL := fmt.Sprintf("%s?q=%s", duckduckgoBaseURL, url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create VQD request: %w", err)
	}

	req.Header.Set("User-Agent", duckduckgoUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute VQD request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code from VQD request: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read VQD response body: %w", err)
	}

	vqd := d.extractVQDToken(string(body))
	if vqd == "" {
		return "", fmt.Errorf("failed to extract VQD token")
	}

	return vqd, nil
}

// extractVQDToken extracts the VQD token from the HTML response
func (d *duckduckgo) extractVQDToken(html string) string {
	re := regexp.MustCompile(`vqd=["']([^"']+)["']`)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// buildSearchURL creates a properly formatted URL for DuckDuckGo search
func (d *duckduckgo) buildSearchURL(query, vqd string, page int) string {
	params := url.Values{}
	params.Set("q", query)
	params.Set("vqd", vqd)
	params.Set("l", "us-en")
	params.Set("dl", "en")
	params.Set("ct", "US")
	params.Set("ss", "1")
	params.Set("sp", "1")
	params.Set("sc", "1")
	params.Set("o", "json")

	if d.region != "" {
		params.Set("kl", d.region)
	}

	if d.safeSearch != "" {
		params.Set("p", d.safeSearch)
	}

	if d.timeRange != "" {
		params.Set("df", d.timeRange)
	}

	if page > 1 {
		params.Set("s", fmt.Sprintf("%d", (page-1)*25))
	}

	return duckduckgoSearchURL + "?" + params.Encode()
}

// parseSearchResponse parses the JSON search response from DuckDuckGo
func (d *duckduckgo) parseSearchResponse(body []byte) (*searchResponse, error) {
	var response struct {
		Results []struct {
			Title       string `json:"t"`
			URL         string `json:"u"`
			Description string `json:"a"`
		} `json:"results"`
		NoResults bool `json:"noResults"`
	}

	err := json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	results := make([]searchResult, 0, len(response.Results))
	for _, r := range response.Results {
		if r.Description == "" && r.URL == "" && r.Title == "" {
			continue
		}

		results = append(results, searchResult{
			Title:       r.Title,
			URL:         r.URL,
			Description: r.Description,
		})
	}

	return &searchResponse{
		Results:   results,
		NoResults: response.NoResults,
	}, nil
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
	return true // DuckDuckGo doesn't require API keys
}
