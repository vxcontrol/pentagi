package searchers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/system"

	"github.com/sirupsen/logrus"
)

// Constants for Perplexity API
const (
	perplexityURL = "https://api.perplexity.ai/chat/completions"
	// A single web-grounded completion, not an agent run: sonar answers in
	// seconds, but a large search_context_size plus synthesis can take longer,
	// so the ceiling is generous. PERPLEXITY_TIMEOUT overrides it.
	defaultPerplexityTimeout = 120 * time.Second
	// The Sonar model chat/completions serves when PERPLEXITY_MODEL is unset.
	defaultPerplexityModel = "sonar"
	// search_context_size when PERPLEXITY_CONTEXT_SIZE is unset.
	defaultPerplexityContextSize = "low"
)

type chatRequest struct {
	Model            string            `json:"model"`
	Messages         []chatMessage     `json:"messages"`
	WebSearchOptions *webSearchOptions `json:"web_search_options,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type webSearchOptions struct {
	SearchContextSize string `json:"search_context_size,omitempty"`
}

type chatResponse struct {
	ID      string       `json:"id"`
	Model   string       `json:"model"`
	Choices []chatChoice `json:"choices"`
	// Perplexity returns sources both as a flat list of URLs (citations) and,
	// on newer responses, as structured search_results carrying titles. Either
	// may be present; both are read and merged.
	Citations     []string           `json:"citations"`
	SearchResults []chatSearchResult `json:"search_results"`
}

type chatChoice struct {
	Message      chatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type chatSearchResult struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Date  string `json:"date"`
}

// perplexity - structure for working with Perplexity API
type perplexity struct {
	cfg        *config.Config
	summarizer SummarizeHandler
}

func NewPerplexity(cfg *config.Config, summarizer SummarizeHandler) Searcher {
	return &perplexity{
		cfg:        cfg,
		summarizer: summarizer,
	}
}

func (p *perplexity) Engine() database.SearchengineType {
	return database.SearchengineTypePerplexity
}

// Handle processes a search request through Perplexity API
func (p *perplexity) Handle(ctx context.Context, req Request) (string, error) {
	if !p.IsAvailable() {
		return "", ErrNotConfigured
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"engine": "perplexity",
		"query":  req.Query[:min(len(req.Query), 1000)],
		"model":  p.model(),
	})

	result, err := p.search(ctx, req.Query)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("search engine error"),
			langfuse.WithEventInput(req.Query),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"engine": "perplexity",
				"query":  req.Query,
				"model":  p.model(),
				"error":  err.Error(),
			}),
		)

		obs.LogErrorOrCancel(logger, err, "failed to search in perplexity")
		return "", err
	}

	return result, nil
}

// search performs a request to Perplexity API
func (p *perplexity) search(ctx context.Context, query string) (string, error) {
	client, err := system.GetHTTPClient(p.cfg)
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to create http client: %w", err))
	}

	client.Timeout = p.timeout()

	// Forming the request
	reqPayload := chatRequest{
		Model:            p.model(),
		Messages:         []chatMessage{{Role: "user", Content: query}},
		WebSearchOptions: &webSearchOptions{SearchContextSize: p.contextSize()},
	}

	// Serializing the request
	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to marshal request body: %w", err))
	}

	// Creating HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, perplexityURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to create request: %w", err))
	}

	// Setting request headers
	req.Header.Set("Authorization", "Bearer "+p.apiKey())
	req.Header.Set("Content-Type", "application/json")

	// Sending the request
	resp, err := client.Do(req)
	if err != nil {
		// Our own budget running out is not a transient the way a 502 is: asking
		// again spends the budget a second time to fail the same way. Fall
		// through to the next engine instead.
		if errors.Is(err, context.DeadlineExceeded) {
			return "", Fatal(fmt.Errorf("failed to send request: %w", err))
		}

		return "", Retryable(fmt.Errorf("failed to send request: %w", err), 0)
	}
	defer resp.Body.Close()

	// Handling the response. handleErrorResponse keeps the per-status message; the
	// retryable/fatal classification is decided here from the status code.
	if resp.StatusCode != http.StatusOK {
		baseErr := p.handleErrorResponse(resp.StatusCode)
		if detail := readErrorDetail(resp.Body); detail != "" {
			baseErr = fmt.Errorf("%w: %s", baseErr, detail)
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			return "", Retryable(baseErr, 0)
		}
		return "", Fatal(baseErr)
	}

	// Reading the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", Retryable(fmt.Errorf("failed to read response body: %w", err), 0)
	}

	// Deserializing the response
	var response chatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", Fatal(fmt.Errorf("failed to unmarshal response: %w", err))
	}

	// Forming the result
	result := p.formatResponse(ctx, &response, query)
	if result == "" {
		return "", Fatal(fmt.Errorf("the completion carries no answer"))
	}

	return result, nil
}

// handleErrorResponse handles erroneous HTTP statuses
func (p *perplexity) handleErrorResponse(statusCode int) error {
	switch statusCode {
	case http.StatusBadRequest:
		return errors.New("request is invalid")
	case http.StatusUnauthorized:
		return errors.New("API key is wrong")
	case http.StatusForbidden:
		return errors.New("the endpoint requested is hidden for administrators only")
	case http.StatusNotFound:
		return errors.New("the specified endpoint could not be found")
	case http.StatusMethodNotAllowed:
		return errors.New("there need to try to access an endpoint with an invalid method")
	case http.StatusTooManyRequests:
		return errors.New("there are requesting too many results")
	case http.StatusInternalServerError:
		return errors.New("there had a problem with our server. try again later")
	case http.StatusBadGateway:
		return errors.New("there was a problem with the server. Please try again later")
	case http.StatusServiceUnavailable:
		return errors.New("there are temporarily offline for maintenance. please try again later")
	case http.StatusGatewayTimeout:
		return errors.New("there are temporarily offline for maintenance. please try again later")
	default:
		return fmt.Errorf("unexpected status code: %d", statusCode)
	}
}

func answerAndCitations(response *chatResponse) (string, []string) {
	var answer strings.Builder
	for _, choice := range response.Choices {
		answer.WriteString(choice.Message.Content)
	}

	var citations []string
	seen := make(map[string]struct{})
	addCitation := func(url string) {
		if url == "" {
			return
		}
		if _, ok := seen[url]; ok {
			return
		}
		seen[url] = struct{}{}
		citations = append(citations, url)
	}

	// search_results first: they carry the same URLs in the same order as
	// citations but are the newer, richer field. citations backfills anything a
	// response reports only there.
	for _, result := range response.SearchResults {
		addCitation(result.URL)
	}
	for _, url := range response.Citations {
		addCitation(url)
	}

	return answer.String(), citations
}

func (p *perplexity) formatResponse(ctx context.Context, response *chatResponse, query string) string {
	content, citations := answerAndCitations(response)
	if content == "" {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("# Answer\n\n")
	builder.WriteString(content)

	if len(citations) > 0 {
		builder.WriteString("\n\n# Citations\n\n")
		for i, citation := range citations {
			fmt.Fprintf(&builder, "%d. %s\n", i+1, citation)
		}
	}

	rawContent := builder.String()
	if len(rawContent) > maxRawContentLength {
		// Check if summarizer is available
		if p.summarizer != nil {
			summarizePrompt, err := p.getSummarizePrompt(query, rawContent, citations)
			if err == nil {
				if summarizedContent, err := p.summarizer(ctx, summarizePrompt); err == nil {
					return summarizedContent
				}
			}
		}
		// If summarizer is nil or failed, truncate content
		return rawContent[:min(len(rawContent), maxRawContentLength)]
	}

	return rawContent
}

// getSummarizePrompt creates a prompt for summarizing Perplexity search results
func (p *perplexity) getSummarizePrompt(query string, content string, citations []string) (string, error) {
	templateText := `<instructions>
TASK: Summarize Perplexity search results for the following user query:

USER QUERY: "{{.Query}}"

DATA:
- <answer> contains the AI-generated response to the user's query
- <citations> contains source references that support the response

REQUIREMENTS:
1. Create focused summary (max {{.MaxLength}} chars) that DIRECTLY answers the user query
2. Preserve all critical facts, technical details, and numerical data from the answer
3. Maintain all actionable insights, procedures, or recommendations
4. Keep ALL query-relevant information even if reducing overall length
5. Retain important source attributions when specific facts are kept
6. Ensure the user query is fully addressed in the summary
7. NEVER remove information that answers the user's original question

FORMAT:
- Begin with a direct answer to the user query
- Maintain the original answer's structure and flow where possible
- Preserve hierarchical organization with headings when present
- Keep bullet points and numbered lists for clarity
- Include the most important citations that support key claims

The summary MUST provide complete answers to the user's query, preserving all relevant information.
</instructions>

<answer>
{{.Content}}
</answer>

{{if .HasCitations}}
<citations>
{{range $index, $citation := .Citations}}{{$index | inc}}. {{$citation}}
{{end}}</citations>
{{end}}`

	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}

	templateContext := map[string]any{
		"Query":        query,
		"MaxLength":    maxRawContentLength,
		"Content":      content,
		"HasCitations": len(citations) > 0,
	}

	if len(citations) > 0 {
		templateContext["Citations"] = citations
	}

	tmpl, err := template.New("summarize").Funcs(funcMap).Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("error creating template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateContext); err != nil {
		return "", fmt.Errorf("error executing template: %v", err)
	}

	return buf.String(), nil
}

func (p *perplexity) IsAvailable() bool {
	return p.apiKey() != ""
}

func (p *perplexity) apiKey() string {
	if p.cfg == nil {
		return ""
	}

	return p.cfg.PerplexityAPIKey
}

// model returns the model to send, verbatim. Perplexity names its own models
// bare (sonar, sonar-pro, ...); the value is passed through untouched so an
// operator pointing PERPLEXITY at a gateway can still use that gateway's own
// naming. An unset value falls back to the default.
func (p *perplexity) model() string {
	if p.cfg == nil {
		return defaultPerplexityModel
	}

	if configured := strings.TrimSpace(p.cfg.PerplexityModel); configured != "" {
		return configured
	}

	return defaultPerplexityModel
}

// readErrorDetail returns the API's own explanation of a rejected request, or
// an empty string when the body carries none.
func readErrorDetail(r io.Reader) string {
	body, err := io.ReadAll(io.LimitReader(r, 8192))
	if err != nil || len(body) == 0 {
		return ""
	}

	var payload struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && payload.Error.Message != "" {
		return payload.Error.Message
	}

	text := strings.TrimSpace(string(body))
	if strings.HasPrefix(text, "<") {
		return ""
	}
	if line, _, found := strings.Cut(text, "\n"); found {
		text = strings.TrimSpace(line)
	}
	if runes := []rune(text); len(runes) > maxErrorDetailRunes {
		text = string(runes[:maxErrorDetailRunes]) + "…"
	}
	return text
}

const maxErrorDetailRunes = 200

func (p *perplexity) contextSize() string {
	if p.cfg == nil || p.cfg.PerplexityContextSize == "" {
		return defaultPerplexityContextSize
	}

	return p.cfg.PerplexityContextSize
}

func (p *perplexity) timeout() time.Duration {
	if p.cfg == nil || p.cfg.PerplexityTimeout <= 0 {
		return defaultPerplexityTimeout
	}

	return time.Duration(p.cfg.PerplexityTimeout) * time.Second
}
