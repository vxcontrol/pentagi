package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/system"

	"github.com/sirupsen/logrus"
	customsearch "google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

const googleMaxResults = 10

type google struct {
	cfg       *config.Config
	flowID    int64
	taskID    *int64
	subtaskID *int64
	slp       SearchLogProvider
}

func NewGoogleTool(
	cfg *config.Config,
	flowID int64,
	taskID, subtaskID *int64,
	slp SearchLogProvider,
) Tool {
	return &google{
		cfg:       cfg,
		flowID:    flowID,
		taskID:    taskID,
		subtaskID: subtaskID,
		slp:       slp,
	}
}

func (g *google) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if !g.IsAvailable() {
		return "", fmt.Errorf("google is not available")
	}

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

	result, err := g.search(ctx, svc, action.Query, numResults)
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

		logger.WithError(err).Error("failed to search in google")
		result = fmt.Sprintf("failed to search in google: %v", err)
	}

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

	opts := []option.ClientOption{
		option.WithAPIKey(g.apiKey()),
		option.WithHTTPClient(client),
	}

	svc, err := customsearch.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create google search service: %v", err)
	}

	return svc, nil
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
