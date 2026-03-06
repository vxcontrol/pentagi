package tools

import (
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
	customsearch "google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

const googleMaxResults = 10

type google struct {
	flowID    int64
	taskID    *int64
	subtaskID *int64
	apiKey    string
	cxKey     string
	lrKey     string
	proxyURL  string
	slp       SearchLogProvider
}

func NewGoogleTool(flowID int64, taskID, subtaskID *int64,
	apiKey, cxKey, lrKey, proxyURL string, slp SearchLogProvider,
) Tool {
	return &google{
		flowID:    flowID,
		taskID:    taskID,
		subtaskID: subtaskID,
		apiKey:    apiKey,
		cxKey:     cxKey,
		lrKey:     lrKey,
		proxyURL:  proxyURL,
		slp:       slp,
	}
}

func (g *google) parseGoogleSearchResult(res *customsearch.Search) string {
	var writer strings.Builder
	for i, item := range res.Items {
		writer.WriteString(fmt.Sprintf("# %d. %s\n\n", i+1, item.Title))
		writer.WriteString(fmt.Sprintf("## URL\n%s\n\n", item.Link))
		writer.WriteString(fmt.Sprintf("## Snippet\n\n%s\n\n", item.Snippet))
	}

	return writer.String()
}

func (g *google) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	var action SearchAction
	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(g.flowID, g.taskID, g.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	if err := json.Unmarshal(args, &action); err != nil {
		logger.WithError(err).Error("failed to unmarshal google search action")
		return "", fmt.Errorf("failed to unmarshal %s search action arguments: %w", name, err)
	}

	numResults := int64(action.MaxResults)
	if numResults < 1 || numResults > googleMaxResults {
		numResults = googleMaxResults
	}

	logger = logger.WithFields(logrus.Fields{
		"query":       action.Query[:min(len(action.Query), 1000)],
		"num_results": numResults,
	})

	svc, err := g.newSearchService(ctx)
	if err != nil {
		logger.WithError(err).Error("failed to create google search service")
		return "", err
	}

	resp, err := svc.Cse.List().Context(ctx).Cx(g.cxKey).Q(action.Query).Lr(g.lrKey).Num(numResults).Do()
	if err != nil {
		observation.Event(
			langfuse.WithEventName("search engine error swallowed"),
			langfuse.WithEventInput(action.Query),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"tool_name":   GoogleToolName,
				"engine":      "google",
				"query":       action.Query,
				"max_results": numResults,
				"error":       err.Error(),
			}),
		)

		logger.WithError(err).Error("failed to call tool to search in google results")
		return fmt.Sprintf("failed to call tool %s to search in google results: %v", name, err), nil
	}

	result := g.parseGoogleSearchResult(resp)

	if agentCtx, ok := GetAgentContext(ctx); ok {
		_, _ = g.slp.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			database.SearchengineTypeGoogle,
			action.Query,
			result,
			g.taskID,
			g.subtaskID,
		)
	}

	return result, nil
}

func (g *google) newSearchService(ctx context.Context) (*customsearch.Service, error) {
	opts := []option.ClientOption{
		option.WithAPIKey(g.apiKey),
	}

	if g.proxyURL != "" {
		opts = append(opts, option.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(g.proxyURL)
				},
			},
		}))
	}

	svc, err := customsearch.NewService(ctx, option.WithAPIKey(g.apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create google search service: %v", err)
	}

	return svc, nil
}

func (g *google) IsAvailable() bool {
	return g.apiKey != "" && g.cxKey != ""
}
