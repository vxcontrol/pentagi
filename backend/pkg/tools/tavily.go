package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"pentagi/pkg/database"

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
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"tool": name,
		"args": string(args),
	})

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

	for i, result := range result.Results {
		writer.WriteString(fmt.Sprintf("## %d. %s\n\n", i+1, result.Title))
		writer.WriteString(fmt.Sprintf("* URL %s\n", result.URL))
		writer.WriteString(fmt.Sprintf("* Match score %3.3f\n\n", result.Score))
		writer.WriteString(fmt.Sprintf("### Short content\n\n%s\n\n", result.Content))
		if result.RawContent != nil {
			if t.summarizer != nil {
				summarizedContent, err := t.summarizer(ctx, *result.RawContent)
				if err != nil {
					rawContent := *result.RawContent
					rawContent = rawContent[:min(len(rawContent), maxRawContentLength)]
					writer.WriteString(fmt.Sprintf("### Content\n\n%s\n\n", rawContent))
				} else {
					writer.WriteString(fmt.Sprintf("### Content\n\n%s\n\n", summarizedContent))
				}
			} else {
				rawContent := *result.RawContent
				rawContent = rawContent[:min(len(rawContent), maxRawContentLength)]
				writer.WriteString(fmt.Sprintf("### Content (not summarized)\n\n%s\n\n", rawContent))
			}
		}
	}

	return writer.String()
}

func (t *tavily) IsAvailable() bool {
	return t.apiKey != ""
}
