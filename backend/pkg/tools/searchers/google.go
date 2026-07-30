package searchers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/system"

	"github.com/sirupsen/logrus"
	customsearch "google.golang.org/api/customsearch/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const googleMaxResults = 10

type google struct {
	cfg *config.Config
}

func NewGoogle(cfg *config.Config) Searcher {
	return &google{cfg: cfg}
}

func (g *google) Engine() database.SearchengineType {
	return database.SearchengineTypeGoogle
}

func (g *google) Handle(ctx context.Context, req Request) (string, error) {
	if !g.IsAvailable() {
		return "", ErrNotConfigured
	}

	ctx, observation := obs.Observer.NewObservation(ctx)

	numResults := int64(req.MaxResults)
	if numResults < 1 || numResults > googleMaxResults {
		numResults = googleMaxResults
	}

	logger := logrus.WithContext(ctx).WithFields(logrus.Fields{
		"engine":      "google",
		"query":       req.Query[:min(len(req.Query), 1000)],
		"num_results": numResults,
	})

	svc, err := g.newSearchService(ctx)
	if err != nil {
		logger.WithError(err).Error("failed to create google search service")
		return "", Fatal(fmt.Errorf("failed to create google search service: %w", err))
	}

	result, err := g.search(ctx, svc, req.Query, numResults)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("search engine error"),
			langfuse.WithEventInput(req.Query),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"engine":      "google",
				"query":       req.Query,
				"max_results": numResults,
				"error":       err.Error(),
			}),
		)

		obs.LogErrorOrCancel(logger, err, "failed to search in google")
		return "", classifyGoogleError(err)
	}

	return result, nil
}

// classifyGoogleError maps a Google Custom Search failure to a retryable/fatal error.
// A *googleapi.Error carries the upstream HTTP status (429/5xx retryable, 4xx fatal);
// anything else is a transport-level failure, which may clear on retry.
func classifyGoogleError(err error) error {
	var gerr *googleapi.Error
	if errors.As(err, &gerr) {
		return ClassifyHTTPStatus(gerr.Code, "google search failed")
	}
	return Retryable(fmt.Errorf("google search failed: %w", err), 0)
}

func (g *google) search(ctx context.Context, svc *customsearch.Service, query string, numResults int64) (string, error) {
	resp, err := svc.Cse.List().Context(ctx).Cx(g.cxKey()).Q(query).Lr(g.lrKey()).Num(numResults).Do()
	if err != nil {
		return "", fmt.Errorf("failed to do request: %w", err)
	}

	return g.formatResults(resp), nil
}

func (g *google) formatResults(res *customsearch.Search) string {
	var writer strings.Builder
	for i, item := range res.Items {
		writer.WriteString(fmt.Sprintf("# %d. %s\n\n", i+1, item.Title))
		writer.WriteString(fmt.Sprintf("## URL\n%s\n\n", item.Link))
		writer.WriteString(fmt.Sprintf("## Snippet\n\n%s\n\n", item.Snippet))
	}

	return writer.String()
}

func (g *google) newSearchService(ctx context.Context) (*customsearch.Service, error) {
	client, err := system.GetHTTPClient(g.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	// google.golang.org/api normally injects the API key through the HTTP transport it
	// builds itself. But we MUST supply our own proxy/TLS client via WithHTTPClient, and
	// WithHTTPClient takes precedence — it replaces that transport, so option.WithAPIKey
	// is silently dropped and requests go out unauthenticated (HTTP 403 "unregistered
	// caller"). Attach the key ourselves as the `key` query parameter (the documented
	// Custom Search auth) by wrapping the proxy client's transport.
	client.Transport = &googleAPIKeyTransport{
		key:  g.apiKey(),
		base: client.Transport,
	}

	svc, err := customsearch.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create google search service: %v", err)
	}

	return svc, nil
}

// googleAPIKeyTransport attaches the Google API key as the `key` query parameter to
// every outgoing request, so authentication survives the proxy/TLS client that
// option.WithHTTPClient forces us to use (see newSearchService).
type googleAPIKeyTransport struct {
	key  string
	base http.RoundTripper
}

func (t *googleAPIKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone so the caller's request is never mutated.
	r := req.Clone(req.Context())
	q := r.URL.Query()
	q.Set("key", t.key)
	r.URL.RawQuery = q.Encode()

	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(r)
}

func (g *google) IsAvailable() bool {
	return g.apiKey() != "" && g.cxKey() != ""
}

func (g *google) apiKey() string {
	if g.cfg == nil {
		return ""
	}

	return g.cfg.GoogleAPIKey
}

func (g *google) cxKey() string {
	if g.cfg == nil {
		return ""
	}

	return g.cfg.GoogleCXKey
}

func (g *google) lrKey() string {
	if g.cfg == nil {
		return ""
	}

	return g.cfg.GoogleLRKey
}
