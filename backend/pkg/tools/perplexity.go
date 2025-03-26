package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"pentagi/pkg/database"

	"github.com/sirupsen/logrus"
)

// Constants for Perplexity API
const (
	perplexityURL         = "https://api.perplexity.ai/chat/completions"
	perplexityTimeout     = 60 * time.Second
	perplexityModel       = "sonar"
	perplexityTemperature = 0.5
	perplexityTopP        = 0.9
	perplexityMaxTokens   = 4000
)

// Message - structure for Perplexity API message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest - request to Perplexity API
type CompletionRequest struct {
	Messages               []Message `json:"messages"`
	Model                  string    `json:"model"`
	MaxTokens              int       `json:"max_tokens"`
	Temperature            float64   `json:"temperature"`
	TopP                   float64   `json:"top_p"`
	SearchContextSize      string    `json:"search_context_size"`
	SearchDomainFilter     []string  `json:"search_domain_filter,omitempty"`
	ReturnImages           bool      `json:"return_images"`
	ReturnRelatedQuestions bool      `json:"return_related_questions"`
	SearchRecencyFilter    string    `json:"search_recency_filter,omitempty"`
	TopK                   int       `json:"top_k,omitempty"`
	Stream                 bool      `json:"stream"`
	PresencePenalty        float64   `json:"presence_penalty,omitempty"`
	FrequencyPenalty       float64   `json:"frequency_penalty,omitempty"`
}

// CompletionResponse - response from Perplexity API
type CompletionResponse struct {
	ID        string    `json:"id"`
	Model     string    `json:"model"`
	Created   int       `json:"created"`
	Object    string    `json:"object"`
	Choices   []Choice  `json:"choices"`
	Usage     Usage     `json:"usage"`
	Citations *[]string `json:"citations,omitempty"`
}

// Choice - choice from Perplexity API response
type Choice struct {
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
}

// Usage - information about used tokens
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// perplexity - structure for working with Perplexity API
type perplexity struct {
	flowID      int64
	taskID      *int64
	subtaskID   *int64
	apiKey      string
	proxyURL    string
	model       string
	contextSize string
	temperature float64
	topP        float64
	maxTokens   int
	timeout     time.Duration
	slp         SearchLogProvider
	summarizer  SummarizeHandler
}

func NewPerplexityTool(flowID int64, taskID, subtaskID *int64,
	apiKey, proxyURL, model, contextSize string, temperature, topP float64,
	maxTokens int, timeout time.Duration, slp SearchLogProvider, summarizer SummarizeHandler,
) Tool {
	if model == "" {
		model = perplexityModel
	}

	if temperature == 0 {
		temperature = perplexityTemperature
	}

	if topP == 0 {
		topP = perplexityTopP
	}

	if maxTokens == 0 {
		maxTokens = perplexityMaxTokens
	}

	if timeout == 0 {
		timeout = perplexityTimeout
	}

	return &perplexity{
		flowID:      flowID,
		taskID:      taskID,
		subtaskID:   subtaskID,
		apiKey:      apiKey,
		proxyURL:    proxyURL,
		model:       model,
		contextSize: contextSize,
		temperature: temperature,
		topP:        topP,
		maxTokens:   maxTokens,
		timeout:     timeout,
		slp:         slp,
		summarizer:  summarizer,
	}
}

// Handle processes a search request through Perplexity API
func (t *perplexity) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	var action SearchAction
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"tool": name,
		"args": string(args),
	})

	if err := json.Unmarshal(args, &action); err != nil {
		logger.WithError(err).Error("failed to unmarshal perplexity search action")
		return "", fmt.Errorf("failed to unmarshal %s search action arguments: %w", name, err)
	}

	logger = logger.WithFields(logrus.Fields{
		"query":       action.Query[:min(len(action.Query), 1000)],
		"max_results": action.MaxResults,
	})

	result, err := t.search(ctx, action.Query, action.MaxResults.Int())
	if err != nil {
		logger.WithError(err).Error("failed to search in perplexity")
		return fmt.Sprintf("failed to search in perplexity: %v", err), nil
	}

	if agentCtx, ok := GetAgentContext(ctx); ok {
		_, _ = t.slp.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			database.SearchengineTypePerplexity,
			action.Query,
			result,
			t.taskID,
			t.subtaskID,
		)
	}

	return result, nil
}

// search performs a request to Perplexity API
func (t *perplexity) search(ctx context.Context, query string, maxResults int) (string, error) {
	// Setting up HTTP client with timeout
	httpClient := &http.Client{
		Timeout: t.timeout,
	}

	// Setting up proxy if specified
	if t.proxyURL != "" {
		httpClient.Transport = &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(t.proxyURL)
			},
		}
	}

	// Creating message for the request
	messages := []Message{
		{
			Role:    "user",
			Content: query,
		},
	}

	// Forming the request
	reqPayload := CompletionRequest{
		Messages:               messages,
		Model:                  t.model,
		SearchContextSize:      t.contextSize,
		MaxTokens:              t.maxTokens,
		Temperature:            t.temperature,
		TopP:                   t.topP,
		ReturnImages:           false,
		ReturnRelatedQuestions: false,
		Stream:                 false,
	}

	// Serializing the request
	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Creating HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, perplexityURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Setting request headers
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Sending the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handling the response
	if resp.StatusCode != http.StatusOK {
		return "", t.handleErrorResponse(resp.StatusCode)
	}

	// Reading the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Deserializing the response
	var response CompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Forming the result
	result := t.formatResponse(ctx, &response, maxResults)
	return result, nil
}

// handleErrorResponse handles erroneous HTTP statuses
func (t *perplexity) handleErrorResponse(statusCode int) error {
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

// formatResponse formats the API response into readable text
func (t *perplexity) formatResponse(ctx context.Context, response *CompletionResponse, maxResults int) string {
	var builder strings.Builder

	// Checking for response choices
	if len(response.Choices) == 0 {
		return "No response received from Perplexity API"
	}

	// Getting the response content
	content := response.Choices[0].Message.Content
	builder.WriteString("# Answer\n\n")
	builder.WriteString(content)

	// Adding citations if available and within maxResults limit
	if response.Citations != nil && len(*response.Citations) > 0 {
		builder.WriteString("\n\n# Citations\n\n")
		for i, citation := range *response.Citations {
			if i >= maxResults {
				break
			}
			builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, citation))
		}
	}

	rawContent := builder.String()
	if len(rawContent) > maxRawContentLength {
		// Check if summarizer is available
		if t.summarizer != nil {
			summarizedContent, err := t.summarizer(ctx, rawContent)
			if err == nil {
				return summarizedContent
			}
		}
		// If summarizer is nil or failed, truncate content
		return rawContent[:min(len(rawContent), maxRawContentLength)]
	}

	return rawContent
}

// isAvailable checks the availability of the API
func (t *perplexity) IsAvailable() bool {
	return t.apiKey != ""
}
