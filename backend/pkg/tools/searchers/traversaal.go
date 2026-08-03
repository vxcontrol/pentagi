package searchers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/system"

	"github.com/sirupsen/logrus"
)

const traversaalURL = "https://api-ares.traversaal.ai/live/predict"

type traversaalSearchResult struct {
	Response string   `json:"response_text"`
	Links    []string `json:"web_url"`
}

type traversaal struct {
	cfg *config.Config
}

func NewTraversaal(cfg *config.Config) Searcher {
	return &traversaal{cfg: cfg}
}

func (t *traversaal) Engine() database.SearchengineType {
	return database.SearchengineTypeTraversaal
}

func (t *traversaal) Handle(ctx context.Context, req Request) (string, error) {
	if !t.IsAvailable() {
		return "", ErrNotConfigured
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"engine": "traversaal",
		"query":  req.Query[:min(len(req.Query), 1000)],
	})

	result, err := t.search(ctx, req.Query)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("search engine error"),
			langfuse.WithEventInput(req.Query),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"engine": "traversaal",
				"query":  req.Query,
				"error":  err.Error(),
			}),
		)

		obs.LogErrorOrCancel(logger, err, "failed to search in traversaal")
		return "", err
	}

	return result, nil
}

func (t *traversaal) search(ctx context.Context, query string) (string, error) {
	client, err := system.GetHTTPClient(t.cfg)
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to create http client: %w", err))
	}

	reqBody, err := json.Marshal(struct {
		Query string `json:"query"`
	}{
		Query: query,
	})
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to marshal request body: %v", err))
	}

	req, err := http.NewRequest(http.MethodPost, traversaalURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to build request: %v", err))
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", t.apiKey())

	resp, err := client.Do(req)
	if err != nil {
		return "", Retryable(fmt.Errorf("failed to do request: %v", err), 0)
	}
	defer resp.Body.Close()

	return t.parseHTTPResponse(resp)
}

func (t *traversaal) parseHTTPResponse(resp *http.Response) (string, error) {
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			return "", Retryable(err, 0)
		}
		return "", Fatal(err)
	}
	var respBody struct {
		Data traversaalSearchResult `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", Fatal(fmt.Errorf("failed to decode response body: %v", err))
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
	return t.apiKey() != ""
}

func (t *traversaal) apiKey() string {
	if t.cfg == nil {
		return ""
	}

	return t.cfg.TraversaalAPIKey
}
