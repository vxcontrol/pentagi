package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"

	"github.com/sirupsen/logrus"
)

const tavilyURL = "https://api.tavily.com/search"

const maxRawContentLength = 3000

type tavilyRequest struct {
	ApiKey            string   `json:"api_key"`
	Query             string   `json:"query"`
	Topic             string   `json:"topic"`
	SearchDepth       string   `json:"search_depth,omitempty"`
	IncludeImages     bool     `json:"include_images,omitempty"`
	IncludeAnswer     bool     `json:"include_answer,omitempty"`
	IncludeRawContent bool     `json:"include_raw_content,omitempty"`
	MaxResults        int      `json:"max_results,omitempty"`
	IncludeDomains    []string `json:"include_domains,omitempty"`
	ExcludeDomains    []string `json:"exclude_domains,omitempty"`
}

type tavilySearchResult struct {
	Answer       string         `json:"answer"`
	Query        string         `json:"query"`
	ResponseTime float64        `json:"response_time"`
	Results      []tavilyResult `json:"results"`
}

type tavilyResult struct {
	Title      string  `json:"title"`
	URL        string  `json:"url"`
	Content    string  `json:"content"`
	RawContent *string `json:"raw_content"`
	Score      float64 `json:"score"`
}

type tavily struct {
	flowID     int64
	taskID     *int64
	subtaskID  *int64
	apiKey     string
	proxyURL   string
	slp        SearchLogProvider
	summarizer SummarizeHandler
}

func NewTavilyTool(flowID int64, taskID, subtaskID *int64, apiKey, proxyURL string,
	slp SearchLogProvider, summarizer SummarizeHandler,
) Tool {
	return &tavily{
		flowID:     flowID,
		taskID:     taskID,
		subtaskID:  subtaskID,
		apiKey:     apiKey,
		proxyURL:   proxyURL,
		slp:        slp,
		summarizer: summarizer,
	}
}

func (t *tavily) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	var action SearchAction
	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(t.flowID, t.taskID, t.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	if err := json.Unmarshal(args, &action); err != nil {
		logger.WithError(err).Error("failed to unmarshal tavily search action")
		return "", fmt.Errorf("failed to unmarshal %s search action arguments: %w", name, err)
	}

	logger = logger.WithFields(logrus.Fields{
		"query":       action.Query[:min(len(action.Query), 1000)],
		"max_results": action.MaxResults,
	})

	result, err := t.search(ctx, action.Query, action.MaxResults.Int())
	if err != nil {
		observation.Event(
			langfuse.WithEventName("search engine error swallowed"),
			langfuse.WithEventInput(action.Query),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"tool_name":   TavilyToolName,
				"engine":      "tavily",
				"query":       action.Query,
				"max_results": action.MaxResults.Int(),
				"error":       err.Error(),
			}),
		)

		logger.WithError(err).Error("failed to search in tavily")
		return fmt.Sprintf("failed to search in tavily: %v", err), nil
	}

	if agentCtx, ok := GetAgentContext(ctx); ok {
		_, _ = t.slp.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			database.SearchengineTypeTavily,
			action.Query,
			result,
			t.taskID,
			t.subtaskID,
		)
	}

	return result, nil
}

func (t *tavily) search(ctx context.Context, query string, maxResults int) (string, error) {
	client := http.DefaultClient
	if t.proxyURL != "" {
		client.Transport = &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(t.proxyURL)
			},
		}
	}

	reqPayload := tavilyRequest{
		Query:             query,
		ApiKey:            t.apiKey,
		Topic:             "general",
		SearchDepth:       "advanced",
		IncludeImages:     false,
		IncludeAnswer:     true,
		IncludeRawContent: true,
		MaxResults:        maxResults,
	}
	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, tavilyURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to build request: %v", err)
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to do request: %v", err)
	}
	defer resp.Body.Close()

	return t.parseHTTPResponse(ctx, resp)
}

func (t *tavily) parseHTTPResponse(ctx context.Context, resp *http.Response) (string, error) {
	switch resp.StatusCode {
	case http.StatusOK:
		var respBody tavilySearchResult
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return "", fmt.Errorf("failed to decode response body: %v", err)
		}
		return t.buildTavilyResult(ctx, &respBody), nil
	case http.StatusBadRequest:
		return "", fmt.Errorf("request is invalid")
	case http.StatusUnauthorized:
		return "", fmt.Errorf("API key is wrong")
	case http.StatusForbidden:
		return "", fmt.Errorf("the endpoint requested is hidden for administrators only")
	case http.StatusNotFound:
		return "", fmt.Errorf("the specified endpoint could not be found")
	case http.StatusMethodNotAllowed:
		return "", fmt.Errorf("there need to try to access an endpoint with an invalid method")
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

func (t *tavily) buildTavilyResult(ctx context.Context, result *tavilySearchResult) string {
	var writer strings.Builder
	writer.WriteString("# Answer\n\n")
	writer.WriteString(result.Answer)
	writer.WriteString("\n\n# Links\n\n")

	isRawContentExists := false
	for i, result := range result.Results {
		writer.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, result.Title))
		writer.WriteString(fmt.Sprintf("* URL %s\n", result.URL))
		writer.WriteString(fmt.Sprintf("* Match score %3.3f\n\n", result.Score))
		writer.WriteString(fmt.Sprintf("### Short content\n\n%s\n\n", result.Content))
		if result.RawContent != nil {
			isRawContentExists = true
		}
	}

	if isRawContentExists && t.summarizer != nil {
		summarizePrompt, err := t.getSummarizePrompt(result.Query, result)
		if err != nil {
			writer.WriteString(t.getRawContentFromResults(result.Results))
		} else {
			summarizedContents, err := t.summarizer(ctx, summarizePrompt)
			if err != nil {
				writer.WriteString(t.getRawContentFromResults(result.Results))
			} else {
				writer.WriteString(fmt.Sprintf("### Summarized Content\n\n%s\n\n", summarizedContents))
			}
		}
	} else {
		writer.WriteString(t.getRawContentFromResults(result.Results))
	}

	return writer.String()
}

func (t *tavily) getRawContentFromResults(results []tavilyResult) string {
	var writer strings.Builder
	for i, result := range results {
		if result.RawContent != nil {
			rawContent := *result.RawContent
			rawContent = rawContent[:min(len(rawContent), maxRawContentLength)]
			writer.WriteString(fmt.Sprintf("### Raw content for %d. %s\n\n%s\n\n", i+1, result.Title, rawContent))
		}
	}
	return writer.String()
}

func (t *tavily) getSummarizePrompt(query string, result *tavilySearchResult) (string, error) {
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

{{range $index, $result := .Results}}
{{if $result.RawContent}}
<raw_content id="{{$index}}" title="{{$result.Title}}" url="{{$result.URL}}">
{{$result.RawContent}}
</raw_content>
{{end}}
{{end}}`

	templateContext := map[string]any{
		"Query":     query,
		"MaxLength": maxRawContentLength,
		"Results":   result.Results,
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

func (t *tavily) IsAvailable() bool {
	return t.apiKey != ""
}
