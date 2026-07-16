package tools

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
	Sources       []string                `json:"sources,omitempty"`
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
	flowID     int64
	taskID     *int64
	subtaskID  *int64
	slp        SearchLogProvider
	summarizer SummarizeHandler
}

func NewFirecrawlTool(
	cfg *config.Config,
	flowID int64,
	taskID, subtaskID *int64,
	slp SearchLogProvider,
	summarizer SummarizeHandler,
) Tool {
	return &firecrawl{
		cfg:        cfg,
		flowID:     flowID,
		taskID:     taskID,
		subtaskID:  subtaskID,
		slp:        slp,
		summarizer: summarizer,
	}
}

func (f *firecrawl) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if !f.IsAvailable() {
		return "", fmt.Errorf("firecrawl is not available")
	}

	var action SearchAction
	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(f.flowID, f.taskID, f.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	if err := json.Unmarshal(args, &action); err != nil {
		logger.WithError(err).Error("failed to unmarshal firecrawl search action")
		return "", fmt.Errorf("failed to unmarshal %s search action arguments: %w", name, err)
	}

	logger = logger.WithFields(logrus.Fields{
		"query":       action.Query[:min(len(action.Query), 1000)],
		"max_results": action.MaxResults,
	})

	result, err := f.search(ctx, action.Query, action.MaxResults.Int())
	if err != nil {
		observation.Event(
			langfuse.WithEventName("search engine error swallowed"),
			langfuse.WithEventInput(action.Query),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"tool_name":   FirecrawlToolName,
				"engine":      "firecrawl",
				"query":       action.Query,
				"max_results": action.MaxResults.Int(),
				"error":       err.Error(),
			}),
		)

		logger.WithError(err).Error("failed to search in firecrawl")
		return fmt.Sprintf("failed to search in firecrawl: %v", err), nil
	}

	if agentCtx, ok := GetAgentContext(ctx); ok {
		_, _ = f.slp.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			database.SearchengineTypeFirecrawl,
			action.Query,
			result,
			f.taskID,
			f.subtaskID,
		)
	}

	return result, nil
}

func (f *firecrawl) search(ctx context.Context, query string, maxResults int) (string, error) {
	client, err := system.GetHTTPClient(f.cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create http client: %w", err)
	}

	reqPayload := firecrawlRequest{
		Query:   query,
		Limit:   maxResults,
		Sources: []string{"web"},
		ScrapeOptions: &firecrawlScrapeOptions{
			Formats:         []string{"markdown"},
			OnlyMainContent: true,
		},
	}
	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, f.searchURL(), bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to build request: %v", err)
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+f.apiKey())

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to do request: %v", err)
	}
	defer resp.Body.Close()

	return f.parseHTTPResponse(ctx, resp)
}

func (f *firecrawl) parseHTTPResponse(ctx context.Context, resp *http.Response) (string, error) {
	switch resp.StatusCode {
	case http.StatusOK:
		var respBody firecrawlSearchResult
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return "", fmt.Errorf("failed to decode response body: %v", err)
		}
		if !respBody.Success {
			if respBody.Error != "" {
				return "", fmt.Errorf("request failed: %s", respBody.Error)
			}
			return "", fmt.Errorf("request failed")
		}
		return f.buildFirecrawlResult(ctx, &respBody), nil
	case http.StatusBadRequest:
		return "", fmt.Errorf("request is invalid")
	case http.StatusUnauthorized:
		return "", fmt.Errorf("API key is wrong")
	case http.StatusPaymentRequired:
		return "", fmt.Errorf("insufficient credits to perform this request")
	case http.StatusForbidden:
		return "", fmt.Errorf("the endpoint requested is hidden for administrators only")
	case http.StatusNotFound:
		return "", fmt.Errorf("the specified endpoint could not be found")
	case http.StatusMethodNotAllowed:
		return "", fmt.Errorf("there need to try to access an endpoint with an invalid method")
	case http.StatusRequestTimeout:
		return "", fmt.Errorf("the request timed out. try again later")
	case http.StatusTooManyRequests:
		return "", fmt.Errorf("there are requesting too many results")
	case http.StatusInternalServerError:
		return "", fmt.Errorf("there had a problem with our server. try again later")
	case http.StatusBadGateway:
		return "", fmt.Errorf("there was a problem with the server. Please try again later")
	case http.StatusServiceUnavailable:
		return "", fmt.Errorf("there are temporarily offline for maintenance. please try again later")
	case http.StatusGatewayTimeout:
		return "", fmt.Errorf("there are temporarily offline for maintenance. please try again later")
	default:
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (f *firecrawl) buildFirecrawlResult(ctx context.Context, result *firecrawlSearchResult) string {
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
		summarizePrompt, err := f.getSummarizePrompt(result)
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

func (f *firecrawl) getSummarizePrompt(result *firecrawlSearchResult) (string, error) {
	templateText := `<instructions>
TASK: Summarize web search results for the following user query:

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

{{range $index, $result := .Results}}
{{if $result.Markdown}}
<raw_content id="{{$index}}" title="{{$result.Title}}" url="{{$result.URL}}">
{{$result.Markdown}}
</raw_content>
{{end}}
{{end}}`

	templateContext := map[string]any{
		"MaxLength": maxRawContentLength,
		"Results":   result.Data.Web,
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
