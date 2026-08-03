package searchers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/system"

	"github.com/sirupsen/logrus"
)

const (
	defaultSearxngTimeout = 30 * time.Second
)

type searxng struct {
	cfg        *config.Config
	summarizer SummarizeHandler
}

func NewSearxng(cfg *config.Config, summarizer SummarizeHandler) Searcher {
	return &searxng{
		cfg:        cfg,
		summarizer: summarizer,
	}
}

func (s *searxng) Engine() database.SearchengineType {
	return database.SearchengineTypeSearxng
}

func (s *searxng) IsAvailable() bool {
	return s.baseURL() != ""
}

func (s *searxng) Handle(ctx context.Context, req Request) (string, error) {
	if !s.IsAvailable() {
		return "", ErrNotConfigured
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"engine":      "searxng",
		"query":       req.Query[:min(len(req.Query), 1000)],
		"max_results": req.MaxResults,
	})

	result, err := s.search(ctx, req.Query, req.MaxResults)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("search engine error"),
			langfuse.WithEventInput(req.Query),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"engine":      "searxng",
				"query":       req.Query,
				"max_results": req.MaxResults,
				"error":       err.Error(),
			}),
		)

		obs.LogErrorOrCancel(logger, err, "failed to search in searxng")
		return "", err
	}

	return result, nil
}

func (s *searxng) search(ctx context.Context, query string, maxResults int) (string, error) {
	apiURL, err := url.Parse(s.baseURL())
	if err != nil {
		return "", Fatal(fmt.Errorf("invalid searxng base URL: %w", err))
	}

	if !strings.HasSuffix(apiURL.Path, "/search") {
		apiURL.Path = strings.TrimSuffix(apiURL.Path, "/") + "/search"
	}

	params := url.Values{}
	params.Add("q", query)
	params.Add("format", "json")
	params.Add("language", s.language())
	params.Add("categories", s.categories())
	params.Add("safesearch", s.safeSearch())

	if timeRange := s.timeRange(); timeRange != "" {
		params.Add("time_range", timeRange)
	}

	if maxResults > 0 {
		params.Add("limit", strconv.Itoa(maxResults))
	} else {
		params.Add("limit", "10")
	}

	apiURL.RawQuery = params.Encode()

	client, err := system.GetHTTPClient(s.cfg)
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to create http client: %w", err))
	}

	client.Timeout = s.timeout()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to create request: %w", err))
	}

	req.Header.Set("User-Agent", "PentAGI/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", Retryable(fmt.Errorf("failed to do request: %w", err), 0)
	}
	defer resp.Body.Close()

	return s.parseHTTPResponse(resp, query)
}

func (s *searxng) parseHTTPResponse(resp *http.Response, query string) (string, error) {
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			return "", Retryable(err, 0)
		}
		return "", Fatal(err)
	}

	var searxngResponse SearxngResponse
	if err := json.NewDecoder(resp.Body).Decode(&searxngResponse); err != nil {
		return "", Fatal(fmt.Errorf("failed to decode response body: %w", err))
	}

	return s.formatResults(searxngResponse.Results, query), nil
}

func (s *searxng) formatResults(results []SearxngResult, query string) string {
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

func (s *searxng) baseURL() string {
	if s.cfg == nil {
		return ""
	}

	return s.cfg.SearxngURL
}

func (s *searxng) categories() string {
	if s.cfg == nil {
		return ""
	}

	return s.cfg.SearxngCategories
}

func (s *searxng) language() string {
	if s.cfg == nil {
		return ""
	}

	return s.cfg.SearxngLanguage
}

func (s *searxng) safeSearch() string {
	if s.cfg == nil {
		return ""
	}

	return s.cfg.SearxngSafeSearch
}

func (s *searxng) timeRange() string {
	if s.cfg == nil {
		return ""
	}

	return s.cfg.SearxngTimeRange
}

func (s *searxng) timeout() time.Duration {
	if s.cfg == nil || s.cfg.SearxngTimeout <= 0 {
		return defaultSearxngTimeout
	}

	return time.Duration(s.cfg.SearxngTimeout) * time.Second
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
