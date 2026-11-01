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

const outcrawlSearchURL = "https://api.outcrawl.ai/v1/search"

type outcrawlRequest struct {
	Query   string   `json:"query"`
	Limit   int      `json:"limit,omitempty"`
	Scrape  bool     `json:"scrape"`
	Formats []string `json:"formats,omitempty"`
}

type outcrawlSearchResult struct {
	Hits []outcrawlHit `json:"hits"`
}

type outcrawlHit struct {
	URL      string `json:"url"`
	Title    string `json:"title"`
	Snippet  string `json:"snippet"`
	Rank     int    `json:"rank"`
	Markdown string `json:"markdown"`
}

// outcrawlError is the error body Outcrawl returns with any non-2xx status.
type outcrawlError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type outcrawl struct {
	cfg        *config.Config
	summarizer SummarizeHandler
}

func NewOutcrawl(cfg *config.Config, summarizer SummarizeHandler) Searcher {
	return &outcrawl{
		cfg:        cfg,
		summarizer: summarizer,
	}
}

func (o *outcrawl) Engine() database.SearchengineType {
	return database.SearchengineTypeOutcrawl
}

func (o *outcrawl) Handle(ctx context.Context, req Request) (string, error) {
	if !o.IsAvailable() {
		return "", ErrNotConfigured
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"engine":      "outcrawl",
		"query":       req.Query[:min(len(req.Query), 1000)],
		"max_results": req.MaxResults,
	})

	result, err := o.search(ctx, req.Query, req.MaxResults)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("search engine error"),
			langfuse.WithEventInput(req.Query),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"engine":      "outcrawl",
				"query":       req.Query,
				"max_results": req.MaxResults,
				"error":       err.Error(),
			}),
		)

		obs.LogErrorOrCancel(logger, err, "failed to search in outcrawl")
		return "", err
	}

	return result, nil
}

func (o *outcrawl) search(ctx context.Context, query string, maxResults int) (string, error) {
	client, err := system.GetHTTPClient(o.cfg)
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to create http client: %w", err))
	}

	// scrape asks Outcrawl to return each result page as markdown in the same
	// call, so the summarizer gets the page content without a second request.
	reqPayload := outcrawlRequest{
		Query:   query,
		Limit:   maxResults,
		Scrape:  true,
		Formats: []string{"markdown"},
	}
	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to marshal request body: %v", err))
	}

	req, err := http.NewRequest(http.MethodPost, outcrawlSearchURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to build request: %v", err))
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey())

	resp, err := client.Do(req)
	if err != nil {
		return "", Retryable(fmt.Errorf("failed to do request: %v", err), 0)
	}
	defer resp.Body.Close()

	return o.parseHTTPResponse(ctx, query, resp)
}

func (o *outcrawl) parseHTTPResponse(ctx context.Context, query string, resp *http.Response) (string, error) {
	if resp.StatusCode == http.StatusOK {
		var respBody outcrawlSearchResult
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return "", Fatal(fmt.Errorf("failed to decode response body: %v", err))
		}
		return o.buildOutcrawlResult(ctx, query, &respBody), nil
	}

	// Outcrawl explains every failure in a {code, message} body; keep it so the
	// agent log shows the reason rather than only the status code.
	var apiErr outcrawlError
	_ = json.NewDecoder(resp.Body).Decode(&apiErr)
	detail := ""
	if apiErr.Message != "" {
		detail = ": " + apiErr.Message
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		return "", Fatal(fmt.Errorf("request is invalid%s", detail))
	case http.StatusUnauthorized:
		return "", Fatal(fmt.Errorf("API key is wrong%s", detail))
	case http.StatusPaymentRequired:
		return "", Fatal(fmt.Errorf("insufficient credits to perform this request%s", detail))
	case http.StatusForbidden:
		return "", Fatal(fmt.Errorf("the API key is not allowed to use search%s", detail))
	case http.StatusNotFound:
		return "", Fatal(fmt.Errorf("the specified endpoint could not be found%s", detail))
	case http.StatusRequestTimeout:
		return "", Retryable(fmt.Errorf("the request timed out. try again later%s", detail), 0)
	case http.StatusTooManyRequests:
		return "", Retryable(fmt.Errorf("rate limit reached, too many requests%s", detail), 0)
	case http.StatusInternalServerError, http.StatusBadGateway:
		return "", Retryable(fmt.Errorf("there was a problem with the server. try again later%s", detail), 0)
	case http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return "", Retryable(fmt.Errorf("the service is temporarily unavailable. try again later%s", detail), 0)
	default:
		return "", Fatal(fmt.Errorf("unexpected status code: %d%s", resp.StatusCode, detail))
	}
}

func (o *outcrawl) buildOutcrawlResult(ctx context.Context, query string, result *outcrawlSearchResult) string {
	var writer strings.Builder
	writer.WriteString("# Links\n\n")

	isMarkdownExists := false
	for i, hit := range result.Hits {
		fmt.Fprintf(&writer, "## %d. %s\n\n", i+1, hit.Title)
		fmt.Fprintf(&writer, "* URL %s\n\n", hit.URL)
		if hit.Snippet != "" {
			fmt.Fprintf(&writer, "### Short content\n\n%s\n\n", hit.Snippet)
		}
		if hit.Markdown != "" {
			isMarkdownExists = true
		}
	}

	if isMarkdownExists && o.summarizer != nil {
		summarizePrompt, err := o.getSummarizePrompt(query, result)
		if err != nil {
			writer.WriteString(o.getContentFromResults(result.Hits))
		} else {
			summarizedContents, err := o.summarizer(ctx, summarizePrompt)
			if err != nil {
				writer.WriteString(o.getContentFromResults(result.Hits))
			} else {
				fmt.Fprintf(&writer, "### Summarized Content\n\n%s\n\n", summarizedContents)
			}
		}
	} else {
		writer.WriteString(o.getContentFromResults(result.Hits))
	}

	return writer.String()
}

func (o *outcrawl) getContentFromResults(hits []outcrawlHit) string {
	var writer strings.Builder
	for i, hit := range hits {
		if hit.Markdown != "" {
			markdown := hit.Markdown[:min(len(hit.Markdown), maxRawContentLength)]
			fmt.Fprintf(&writer, "### Raw content for %d. %s\n\n%s\n\n", i+1, hit.Title, markdown)
		}
	}
	return writer.String()
}

// outcrawlPromptDoc is a single source record fed to the summarizer. IDs are
// one-based so they line up with the numbered links in the tool output, and
// markdown is bounded per document so an oversized page can't blow up the prompt.
type outcrawlPromptDoc struct {
	ID       int
	Title    string
	URL      string
	Markdown string
}

func (o *outcrawl) getSummarizePrompt(query string, result *outcrawlSearchResult) (string, error) {
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
	docs := make([]outcrawlPromptDoc, 0, len(result.Hits))
	for i, hit := range result.Hits {
		if hit.Markdown == "" {
			continue
		}
		docs = append(docs, outcrawlPromptDoc{
			ID:       i + 1,
			Title:    hit.Title,
			URL:      hit.URL,
			Markdown: hit.Markdown[:min(len(hit.Markdown), maxRawContentLength)],
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

func (o *outcrawl) IsAvailable() bool {
	return o.apiKey() != ""
}

func (o *outcrawl) apiKey() string {
	if o.cfg == nil {
		return ""
	}

	return o.cfg.OutcrawlAPIKey
}
