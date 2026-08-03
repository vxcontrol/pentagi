package searchers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/system"

	"github.com/sirupsen/logrus"
)

const (
	firecrawlDefaultURL = "https://api.firecrawl.dev"
	firecrawlSearchPath = "/v2/search"
)

type firecrawlRequest struct {
	Query         string                  `json:"query"`
	Limit         int                     `json:"limit,omitempty"`
	ScrapeOptions *firecrawlScrapeOptions `json:"scrapeOptions,omitempty"`
}

type firecrawlScrapeOptions struct {
	Formats         []string `json:"formats"`
	OnlyMainContent bool     `json:"onlyMainContent"`
}

type firecrawlSearchResult struct {
	Success bool          `json:"success"`
	Warning string        `json:"warning"`
	Error   string        `json:"error"`
	Data    firecrawlData `json:"data"`
}

type firecrawlData struct {
	Web []firecrawlResult `json:"web"`
}

type firecrawlResult struct {
	Title       string             `json:"title"`
	Description string             `json:"description"`
	URL         string             `json:"url"`
	Markdown    string             `json:"markdown"`
	Metadata    *firecrawlMetadata `json:"metadata"`
}

type firecrawlMetadata struct {
	Title     string `json:"title"`
	SourceURL string `json:"sourceURL"`
}

// resolvedURL prefers the top-level URL and falls back to the metadata source
// URL, which is where it lands when scrapeOptions turn a result into a document.
func (r firecrawlResult) resolvedURL() string {
	if r.URL != "" {
		return r.URL
	}
	if r.Metadata != nil {
		return r.Metadata.SourceURL
	}
	return ""
}

// resolvedTitle prefers the top-level title and falls back to the metadata title.
func (r firecrawlResult) resolvedTitle() string {
	if r.Title != "" {
		return r.Title
	}
	if r.Metadata != nil {
		return r.Metadata.Title
	}
	return ""
}

type firecrawl struct {
	cfg        *config.Config
	summarizer SummarizeHandler
}

func NewFirecrawl(cfg *config.Config, summarizer SummarizeHandler) Searcher {
	return &firecrawl{
		cfg:        cfg,
		summarizer: summarizer,
	}
}

func (f *firecrawl) Engine() database.SearchengineType {
	return database.SearchengineTypeFirecrawl
}

func (f *firecrawl) Handle(ctx context.Context, req Request) (string, error) {
	if !f.IsAvailable() {
		return "", ErrNotConfigured
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"engine":      "firecrawl",
		"query":       req.Query[:min(len(req.Query), 1000)],
		"max_results": req.MaxResults,
	})

	result, err := f.search(ctx, req.Query, req.MaxResults)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("search engine error"),
			langfuse.WithEventInput(req.Query),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"engine":      "firecrawl",
				"query":       req.Query,
				"max_results": req.MaxResults,
				"error":       err.Error(),
			}),
		)

		obs.LogErrorOrCancel(logger, err, "failed to search in firecrawl")
		return "", err
	}

	return result, nil
}

func (f *firecrawl) search(ctx context.Context, query string, maxResults int) (string, error) {
	client, err := system.GetHTTPClient(f.cfg)
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to create http client: %w", err))
	}

	// sources is intentionally omitted: /v2/search defaults to ["web"], which is
	// all this tool consumes, and it avoids the string-vs-object shape ambiguity
	// that field carries across the API/SDK layers.
	reqPayload := firecrawlRequest{
		Query: query,
		Limit: maxResults,
		ScrapeOptions: &firecrawlScrapeOptions{
			Formats:         []string{"markdown"},
			OnlyMainContent: true,
		},
	}
	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to marshal request body: %v", err))
	}

	req, err := http.NewRequest(http.MethodPost, f.searchURL(), bytes.NewBuffer(reqBody))
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to build request: %v", err))
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+f.apiKey())

	resp, err := client.Do(req)
	if err != nil {
		return "", Retryable(fmt.Errorf("failed to do request: %v", err), 0)
	}
	defer resp.Body.Close()

	return f.parseHTTPResponse(ctx, query, resp)
}

func (f *firecrawl) parseHTTPResponse(ctx context.Context, query string, resp *http.Response) (string, error) {
	switch resp.StatusCode {
	case http.StatusOK:
		var respBody firecrawlSearchResult
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return "", Fatal(fmt.Errorf("failed to decode response body: %v", err))
		}
		if !respBody.Success {
			if respBody.Error != "" {
				return "", Fatal(fmt.Errorf("request failed: %s", respBody.Error))
			}
			return "", Fatal(fmt.Errorf("request failed"))
		}
		return f.buildFirecrawlResult(ctx, query, &respBody), nil
	case http.StatusBadRequest:
		return "", Fatal(fmt.Errorf("request is invalid"))
	case http.StatusUnauthorized:
		return "", Fatal(fmt.Errorf("API key is wrong"))
	case http.StatusPaymentRequired:
		return "", Fatal(fmt.Errorf("insufficient credits to perform this request"))
	case http.StatusForbidden:
		return "", Fatal(fmt.Errorf("the endpoint requested is hidden for administrators only"))
	case http.StatusNotFound:
		return "", Fatal(fmt.Errorf("the specified endpoint could not be found"))
	case http.StatusMethodNotAllowed:
		return "", Fatal(fmt.Errorf("there need to try to access an endpoint with an invalid method"))
	case http.StatusRequestTimeout:
		return "", Retryable(fmt.Errorf("the request timed out. try again later"), 0)
	case http.StatusTooManyRequests:
		return "", Retryable(fmt.Errorf("there are requesting too many results"), 0)
	case http.StatusInternalServerError:
		return "", Retryable(fmt.Errorf("there had a problem with our server. try again later"), 0)
	case http.StatusBadGateway:
		return "", Retryable(fmt.Errorf("there was a problem with the server. Please try again later"), 0)
	case http.StatusServiceUnavailable:
		return "", Retryable(fmt.Errorf("there are temporarily offline for maintenance. please try again later"), 0)
	case http.StatusGatewayTimeout:
		return "", Retryable(fmt.Errorf("there are temporarily offline for maintenance. please try again later"), 0)
	default:
		return "", Fatal(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
	}
}

func (f *firecrawl) buildFirecrawlResult(ctx context.Context, query string, result *firecrawlSearchResult) string {
	var writer strings.Builder
	writer.WriteString("# Links\n\n")

	isMarkdownExists := false
	for i, res := range result.Data.Web {
		writer.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, res.resolvedTitle()))
		writer.WriteString(fmt.Sprintf("* URL %s\n\n", res.resolvedURL()))
		if res.Description != "" {
			writer.WriteString(fmt.Sprintf("### Short content\n\n%s\n\n", res.Description))
		}
		if res.Markdown != "" {
			isMarkdownExists = true
		}
	}

	if isMarkdownExists && f.summarizer != nil {
		summarizePrompt, err := f.getSummarizePrompt(query, result)
		if err != nil {
			writer.WriteString(f.getContentFromResults(result.Data.Web))
		} else {
			summarizedContents, err := f.summarizer(ctx, summarizePrompt)
			if err != nil {
				writer.WriteString(f.getContentFromResults(result.Data.Web))
			} else {
				writer.WriteString(fmt.Sprintf("### Summarized Content\n\n%s\n\n", summarizedContents))
			}
		}
	} else {
		writer.WriteString(f.getContentFromResults(result.Data.Web))
	}

	return writer.String()
}

func (f *firecrawl) getContentFromResults(results []firecrawlResult) string {
	var writer strings.Builder
	for i, res := range results {
		if res.Markdown != "" {
			markdown := res.Markdown
			markdown = markdown[:min(len(markdown), maxRawContentLength)]
			writer.WriteString(fmt.Sprintf("### Raw content for %d. %s\n\n%s\n\n", i+1, res.resolvedTitle(), markdown))
		}
	}
	return writer.String()
}

// firecrawlPromptDoc is a single source record fed to the summarizer. IDs are
// one-based so they line up with the numbered links in the tool output, and the
// title/URL use the metadata fallbacks so nothing renders empty. Markdown is
// bounded per document so an oversized or hostile page can't blow up the prompt.
type firecrawlPromptDoc struct {
	ID       int
	Title    string
	URL      string
	Markdown string
}

func (f *firecrawl) getSummarizePrompt(query string, result *firecrawlSearchResult) (string, error) {
	templateText := `<instructions>
TASK: Summarize web search results for the following user query:

USER QUERY: "{{.Query}}"

DATA:
- <raw_content> tags contain web page content with attributes: id, title, url
- Content may include HTML, structured data, tables, or plain text

REQUIREMENTS:
1. Create concise summary (max {{.MaxLength}} chars) that DIRECTLY answers the user query
2. Preserve ALL critical facts, statistics, technical details, and numerical data
3. Maintain all actionable insights, procedures, or code examples exactly as presented
4. Keep ALL query-relevant information even if reducing overall length
5. Highlight authoritative information and note contradictions between sources
6. Cite sources using [Source #] format when presenting specific claims
7. Ensure the user query is fully addressed in the summary
8. NEVER remove information that answers the user's original question

FORMAT:
- Begin with a direct answer to the user query
- Organize thematically with clear structure using headings
- Keep bullet points and numbered lists for clarity and steps
- Include brief "Sources Overview" section identifying key references

The summary MUST provide complete answers to the user's query, preserving all relevant information.
</instructions>

{{range .Results}}
<raw_content id="{{.ID}}" title="{{.Title}}" url="{{.URL}}">
{{.Markdown}}
</raw_content>
{{end}}`

	// Keep source IDs aligned with the one-based numbering of the links section,
	// so a result without markdown still leaves a gap rather than renumbering.
	docs := make([]firecrawlPromptDoc, 0, len(result.Data.Web))
	for i, res := range result.Data.Web {
		if res.Markdown == "" {
			continue
		}
		markdown := res.Markdown[:min(len(res.Markdown), maxRawContentLength)]
		docs = append(docs, firecrawlPromptDoc{
			ID:       i + 1,
			Title:    res.resolvedTitle(),
			URL:      res.resolvedURL(),
			Markdown: markdown,
		})
	}

	templateContext := map[string]any{
		"Query":     query,
		"MaxLength": maxRawContentLength,
		"Results":   docs,
	}

	tmpl, err := template.New("summarize").Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("error creating template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateContext); err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}

	return buf.String(), nil
}

func (f *firecrawl) IsAvailable() bool {
	return f.apiKey() != ""
}

func (f *firecrawl) apiKey() string {
	if f.cfg == nil {
		return ""
	}

	return f.cfg.FirecrawlAPIKey
}

func (f *firecrawl) searchURL() string {
	baseURL := firecrawlDefaultURL
	if f.cfg != nil && f.cfg.FirecrawlAPIURL != "" {
		baseURL = f.cfg.FirecrawlAPIURL
	}

	return strings.TrimRight(baseURL, "/") + firecrawlSearchPath
}
