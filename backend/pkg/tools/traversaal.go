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
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"

	"github.com/sirupsen/logrus"
)

const traversaalURL = "https://api-ares.traversaal.ai/live/predict"

type traversaalSearchResult struct {
	Response string   `json:"response_text"`
	Links    []string `json:"web_url"`
}

type traversaal struct {
	flowID    int64
	taskID    *int64
	subtaskID *int64
	apiKey    string
	proxyURL  string
	slp       SearchLogProvider
}

func NewTraversaalTool(flowID int64, taskID, subtaskID *int64, apiKey, proxyURL string, slp SearchLogProvider) Tool {
	return &traversaal{
		flowID:    flowID,
		taskID:    taskID,
		subtaskID: subtaskID,
		apiKey:    apiKey,
		proxyURL:  proxyURL,
		slp:       slp,
	}
}

func (t *traversaal) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	var action SearchAction
	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(t.flowID, t.taskID, t.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	if err := json.Unmarshal(args, &action); err != nil {
		logger.WithError(err).Error("failed to unmarshal traversaal search action")
		return "", fmt.Errorf("failed to unmarshal %s search action arguments: %w", name, err)
	}

	logger = logger.WithFields(logrus.Fields{
		"query":       action.Query[:min(len(action.Query), 1000)],
		"max_results": action.MaxResults,
	})

	result, err := t.search(ctx, action.Query)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("search engine error swallowed"),
			langfuse.WithEventInput(action.Query),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"tool_name":   TraversaalToolName,
				"engine":      "traversaal",
				"query":       action.Query,
				"max_results": action.MaxResults.Int(),
				"error":       err.Error(),
			}),
		)

		logger.WithError(err).Error("failed to search in traversaal")
		return fmt.Sprintf("failed to search in traversaal: %v", err), nil
	}

	if agentCtx, ok := GetAgentContext(ctx); ok {
		_, _ = t.slp.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			database.SearchengineTypeTraversaal,
			action.Query,
			result,
			t.taskID,
			t.subtaskID,
		)
	}

	return result, nil
}

func (t *traversaal) search(ctx context.Context, query string) (string, error) {
	client := http.DefaultClient
	if t.proxyURL != "" {
		client.Transport = &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(t.proxyURL)
			},
		}
	}

	reqBody, err := json.Marshal(struct {
		Query string `json:"query"`
	}{
		Query: query,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, traversaalURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to build request: %v", err)
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", t.apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to do request: %v", err)
	}
	defer resp.Body.Close()

	return t.parseHTTPResponse(resp)
}

func (t *traversaal) parseHTTPResponse(resp *http.Response) (string, error) {
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	var respBody struct {
		Data traversaalSearchResult `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", fmt.Errorf("failed to decode response body: %v", err)
	}

	var writer strings.Builder
	writer.WriteString("# Answer\n\n")
	writer.WriteString(respBody.Data.Response)
	writer.WriteString("\n\n# Links\n\n")

	for i, resultLink := range respBody.Data.Links {
		writer.WriteString(fmt.Sprintf("%d. %s\n", i+1, resultLink))
	}

	return writer.String(), nil
}

func (t *traversaal) IsAvailable() bool {
	return t.apiKey != ""
}
