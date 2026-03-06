package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"pentagi/pkg/database"

	"github.com/sirupsen/logrus"
)

const (
	defaultSearxngTimeout = 30 * time.Second
)

// SearxngTool represents the Searxng search tool
type SearxngTool struct {
	flowID     int64
	taskID     *int64
	subtaskID  *int64
	baseURL    string
	categories string
	language   string
	safeSearch string
	timeRange  string
	proxyURL   string
	timeout    time.Duration
	slp        SearchLogProvider
	summarizer SummarizeHandler
}

// NewSearxngTool creates a new Searxng tool instance
func NewSearxngTool(
	flowID int64,
	taskID, subtaskID *int64,
	baseURL, categories, language, safeSearch, timeRange, proxyURL string,
	timeout int,
	slp SearchLogProvider,
	summarizer SummarizeHandler,
) *SearxngTool {
	tool := &SearxngTool{
		flowID:     flowID,
		taskID:     taskID,
		subtaskID:  subtaskID,
		baseURL:    baseURL,
		categories: categories,
		language:   language,
		safeSearch: safeSearch,
		timeRange:  timeRange,
		proxyURL:   proxyURL,
		slp:        slp,
		summarizer: summarizer,
	}

	if timeout > 0 {
		tool.timeout = time.Duration(timeout) * time.Second
	} else {
		tool.timeout = defaultSearxngTimeout
	}

	return tool
}

// IsAvailable checks if the Searxng tool is available
func (s *SearxngTool) IsAvailable() bool {
	return s.baseURL != "" && s.slp != nil
}

// Handle handles the Searxng search tool execution
func (s *SearxngTool) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if !s.IsAvailable() {
		return "", fmt.Errorf("searxng tool is not available: missing base URL or search log provider")
	}

	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(s.flowID, s.taskID, s.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	var searchArgs SearchAction
	if err := json.Unmarshal(args, &searchArgs); err != nil {
		logger.WithError(err).Error("failed to unmarshal searxng search action")
		return "", fmt.Errorf("error unmarshaling search arguments: %w", err)
	}

	// Validate required parameters
	if searchArgs.Query == "" {
		logger.Error("query parameter is required")
		return "", fmt.Errorf("query parameter is required")
	}

	// Log the search
	var searchLogID int64
	var err error
	if agentCtx, ok := GetAgentContext(ctx); ok {
		searchLogID, err = s.slp.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			database.SearchengineTypeSearxng,
			searchArgs.Query,
			"",
			s.taskID,
			s.subtaskID,
		)
		if err != nil {
			logger.WithError(err).Error("failed to create search log")
		}
	}

	// Perform the search
	results, err := s.performSearxngSearch(ctx, searchArgs.Query, searchArgs.MaxResults.Int())
	if err != nil {
		// Update search log with error
		if searchLogID > 0 {
			if agentCtx, ok := GetAgentContext(ctx); ok {
				_, updateErr := s.slp.PutLog(
					ctx,
					agentCtx.ParentAgentType,
					agentCtx.CurrentAgentType,
					database.SearchengineTypeSearxng,
					searchArgs.Query,
					err.Error(),
					s.taskID,
					s.subtaskID,
				)
				if updateErr != nil {
					logger.WithError(updateErr).Error("failed to update search log with error")
				}
			}
		}
		logger.WithError(err).Error("failed to perform searxng search")
		return "", fmt.Errorf("searxng search failed: %w", err)
	}

	// Update search log with results
	if searchLogID > 0 {
		resultJSON, _ := json.Marshal(results)
		if agentCtx, ok := GetAgentContext(ctx); ok {
			_, updateErr := s.slp.PutLog(
				ctx,
				agentCtx.ParentAgentType,
				agentCtx.CurrentAgentType,
				database.SearchengineTypeSearxng,
				searchArgs.Query,
				string(resultJSON),
				s.taskID,
				s.subtaskID,
			)
			if updateErr != nil {
				logger.WithError(updateErr).Error("failed to update search log with results")
			}
		}
	}

	// Format the results
	return s.formatSearchResults(results, searchArgs.Query), nil
}

// performSearxngSearch performs the actual search against the Searxng API
func (s *SearxngTool) performSearxngSearch(ctx context.Context, query string, maxResults int) ([]SearxngResult, error) {
	// Build the Searxng API URL
	apiURL, err := url.Parse(s.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Searxng base URL: %w", err)
	}

	// Add /search path if not present
	if !strings.HasSuffix(apiURL.Path, "/search") {
		apiURL.Path = strings.TrimSuffix(apiURL.Path, "/") + "/search"
	}

	// Prepare query parameters
	params := url.Values{}
	params.Add("q", query)
	params.Add("format", "json")
	params.Add("language", s.language)
	params.Add("categories", s.categories)
	params.Add("safesearch", s.safeSearch)

	if s.timeRange != "" {
		params.Add("time_range", s.timeRange)
	}

	if maxResults > 0 {
		params.Add("limit", strconv.Itoa(maxResults))
	} else {
		params.Add("limit", "10") // Default limit
	}

	apiURL.RawQuery = params.Encode()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: s.timeout,
	}

	// If proxy URL is provided, configure proxy
	if s.proxyURL != "" {
		proxyURL, err := url.Parse(s.proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	// Make the request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "PentAGI/1.0")

	logrus.WithFields(enrichLogrusFields(s.flowID, s.taskID, s.subtaskID, logrus.Fields{
		"url":    apiURL.String(),
		"query":  query,
		"limit":  params.Get("limit"),
		"engine": "searxng",
	})).Debug("Performing Searxng search")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("searxng API returned status code: %d", resp.StatusCode)
	}

	// Parse the response
	var searxngResponse SearxngResponse
	if err := json.NewDecoder(resp.Body).Decode(&searxngResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return searxngResponse.Results, nil
}

// formatSearchResults formats the Searxng results for display
func (s *SearxngTool) formatSearchResults(results []SearxngResult, query string) string {
	if len(results) == 0 {
		return fmt.Sprintf("# No Results Found\n\nNo results were found for query: %s", query)
	}

	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("# Searxng Search Results\n\n## Query: %s\n\n", query))
	builder.WriteString("Results from Searxng meta search engine (aggregated from multiple search engines):\n\n")

	for i, result := range results {
		builder.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, result.Title))

		if result.URL != "" {
			builder.WriteString(fmt.Sprintf("**URL:** [%s](%s)\n\n", result.URL, result.URL))
		}

		if result.Content != "" {
			builder.WriteString(fmt.Sprintf("**Content:** %s\n\n", result.Content))
		}

		if result.Author != "" {
			builder.WriteString(fmt.Sprintf("**Author:** %s\n\n", result.Author))
		}

		if resultPublished := result.PublishedDate; resultPublished != "" {
			builder.WriteString(fmt.Sprintf("**Published:** %s\n\n", resultPublished))
		}

		if result.Engine != "" {
			builder.WriteString(fmt.Sprintf("**Source Engine:** %s\n\n", result.Engine))
		}

		builder.WriteString("---\n\n")
	}

	return builder.String()
}

// SearxngResult represents a single result from Searxng
type SearxngResult struct {
	Title         string `json:"title"`
	URL           string `json:"url"`
	Content       string `json:"content"`
	Author        string `json:"author"`
	PublishedDate string `json:"publishedDate"`
	Engine        string `json:"engine"`
}

// SearxngResponse represents the response from Searxng API
type SearxngResponse struct {
	Query   string          `json:"query"`
	Results []SearxngResult `json:"results"`
	Info    SearxngInfo     `json:"info"`
}

// SearxngInfo contains additional information about the search
type SearxngInfo struct {
	Timings     map[string]interface{} `json:"timings"`
	Results     int                    `json:"results"`
	Engine      string                 `json:"engine"`
	Suggestions []string               `json:"suggestions"`
}
