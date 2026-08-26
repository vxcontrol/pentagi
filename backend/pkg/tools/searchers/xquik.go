package searchers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/system"

	"github.com/sirupsen/logrus"
)

const (
	xquikSearchURL        = "https://xquik.com/api/v1/x/tweets/search"
	xquikMaxResponseBytes = 1 << 20
)

type xquikErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type xquik struct {
	cfg *config.Config
}

func NewXquik(cfg *config.Config) Searcher {
	return &xquik{cfg: cfg}
}

func (x *xquik) Engine() database.SearchengineType {
	return database.SearchengineTypeXquik
}

func (x *xquik) IsAvailable() bool {
	return x.apiKey() != ""
}

func (x *xquik) Handle(ctx context.Context, req Request) (string, error) {
	if !x.IsAvailable() {
		return "", ErrNotConfigured
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	limit := clampXquikLimit(req.MaxResults)
	observedQuery := normalizeXquikField(req.Query, 1000)
	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"engine":      "xquik",
		"query":       observedQuery,
		"max_results": limit,
	})

	result, err := x.search(ctx, req.Query, limit)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("search engine error"),
			langfuse.WithEventInput(observedQuery),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"engine":      "xquik",
				"query":       observedQuery,
				"max_results": limit,
				"error":       err.Error(),
			}),
		)
		obs.LogErrorOrCancel(logger, err, "failed to search X through Xquik")
		return "", err
	}

	return result, nil
}

func (x *xquik) search(ctx context.Context, query string, limit int) (string, error) {
	client, err := system.GetHTTPClient(x.cfg)
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to create HTTP client: %w", err))
	}

	endpoint, err := url.Parse(xquikSearchURL)
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to parse Xquik search URL: %w", err))
	}
	values := endpoint.Query()
	values.Set("q", query)
	values.Set("queryType", "Latest")
	values.Set("limit", strconv.Itoa(limit))
	endpoint.RawQuery = values.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return "", Fatal(fmt.Errorf("failed to build Xquik request: %w", err))
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("x-api-key", x.apiKey())

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", Retryable(fmt.Errorf("failed to call Xquik: %w", err), 0)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, xquikMaxResponseBytes+1))
	if err != nil {
		return "", Retryable(fmt.Errorf("failed to read Xquik response: %w", err), 0)
	}
	if len(body) > xquikMaxResponseBytes {
		return "", Fatal(fmt.Errorf("Xquik response exceeds %d bytes", xquikMaxResponseBytes))
	}

	if resp.StatusCode != http.StatusOK {
		return "", classifyXquikError(resp, body)
	}

	var result xquikSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", Fatal(fmt.Errorf("failed to decode Xquik response: %w", err))
	}

	return formatXquikResults(query, limit, result), nil
}

func (x *xquik) apiKey() string {
	if x == nil || x.cfg == nil {
		return ""
	}
	return strings.TrimSpace(x.cfg.XquikAPIKey)
}

func classifyXquikError(resp *http.Response, body []byte) error {
	message := http.StatusText(resp.StatusCode)
	var payload xquikErrorResponse
	if json.Unmarshal(body, &payload) == nil {
		if payload.Message != "" {
			message = payload.Message
		} else if payload.Error != "" {
			message = payload.Error
		}
	}
	message = normalizeXquikField(message, xquikMaxErrorBytes)
	if message == "" {
		message = "Xquik request failed"
	}

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return Retryable(fmt.Errorf("%s (HTTP %d)", message, resp.StatusCode), parseRetryAfter(resp.Header.Get("Retry-After")))
	}
	return Fatal(fmt.Errorf("%s (HTTP %d)", message, resp.StatusCode))
}
