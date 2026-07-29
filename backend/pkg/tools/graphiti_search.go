package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"pentagi/pkg/graphiti"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// maxGraphitiResponsePreviewBytes bounds how much of a non-2xx Graphiti
// response body is ever shown to the LLM (via truncateString), so a verbose
// error page or stack trace from the Graphiti backend can't blow up the
// agent's context window.
const maxGraphitiResponsePreviewBytes = 512

// graphitiAPIStatusErrorRe matches the one fixed error shape the vendored
// graphiti-go-client emits for a non-2xx HTTP response: a plain fmt.Errorf
// (see client.go's `do`) with no typed status code, so this is the only way
// to recover the status and body without forking the dependency.
var graphitiAPIStatusErrorRe = regexp.MustCompile(`(?s)API request failed with status (\d+): (.*)`)

type GraphitiSearcher interface {
	IsEnabled() bool
	TemporalWindowSearch(ctx context.Context, req graphiti.TemporalSearchRequest) (*graphiti.TemporalSearchResponse, error)
	EntityRelationshipsSearch(ctx context.Context, req graphiti.EntityRelationshipSearchRequest) (*graphiti.EntityRelationshipSearchResponse, error)
	DiverseResultsSearch(ctx context.Context, req graphiti.DiverseSearchRequest) (*graphiti.DiverseSearchResponse, error)
	EpisodeContextSearch(ctx context.Context, req graphiti.EpisodeContextSearchRequest) (*graphiti.EpisodeContextSearchResponse, error)
	SuccessfulToolsSearch(ctx context.Context, req graphiti.SuccessfulToolsSearchRequest) (*graphiti.SuccessfulToolsSearchResponse, error)
	RecentContextSearch(ctx context.Context, req graphiti.RecentContextSearchRequest) (*graphiti.RecentContextSearchResponse, error)
	EntityByLabelSearch(ctx context.Context, req graphiti.EntityByLabelSearchRequest) (*graphiti.EntityByLabelSearchResponse, error)
}

const (
	// Default values for search parameters
	DefaultTemporalMaxResults     = 15
	DefaultRecentMaxResults       = 10
	DefaultSuccessfulMaxResults   = 15
	DefaultEpisodeMaxResults      = 10
	DefaultRelationshipMaxResults = 20
	DefaultDiverseMaxResults      = 10
	DefaultLabelMaxResults        = 25

	DefaultMaxDepth       = 2
	DefaultMinMentions    = 2
	DefaultDiversityLevel = "medium"
	DefaultRecencyWindow  = "24h"
)

// graphitiTimeFormats are the layouts accepted for temporal_window's
// time_start/time_end, tried in order. RFC3339 is the documented format;
// the remaining layouts tolerate a common LLM near-miss - a timestamp
// missing the timezone designator (e.g. "2026-07-24T11:53:34") - which
// would otherwise fail with a confusing parse error even though the value
// is unambiguous. Layouts without a zone parse as UTC, matching the
// graph's own UTC timestamps.
var graphitiTimeFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
}

// parseGraphitiTime parses a temporal_window time_start/time_end value,
// trying each of graphitiTimeFormats in order and returning the error from
// the strict RFC3339 attempt if none match (the most informative for the LLM
// tool-call fixer, since it's the documented/expected format).
func parseGraphitiTime(value string) (time.Time, error) {
	rfc3339Err := error(nil)
	for i, format := range graphitiTimeFormats {
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		} else if i == 0 {
			rfc3339Err = err
		}
	}
	return time.Time{}, rfc3339Err
}

var (
	allowedRecencyWindows = map[string]struct{}{
		"1h":  {},
		"6h":  {},
		"24h": {},
		"7d":  {},
	}
	allowedDiversityLevels = map[string]struct{}{
		"low":    {},
		"medium": {},
		"high":   {},
	}

	// graphitiRetrieverTitles maps each search_type to a human-readable
	// langfuse observation title so the retrieval intent is clear in traces.
	graphitiRetrieverTitles = map[string]string{
		"temporal_window":      "retrieve temporal window context from graphiti knowledge graph",
		"entity_relationships": "retrieve entity relationships from graphiti knowledge graph",
		"diverse_results":      "retrieve diverse results from graphiti knowledge graph",
		"episode_context":      "retrieve episode context from graphiti knowledge graph",
		"successful_tools":     "retrieve successful tools from graphiti knowledge graph",
		"recent_context":       "retrieve recent context from graphiti knowledge graph",
		"entity_by_label":      "retrieve entities by label from graphiti knowledge graph",
	}
)

// graphitiSearchTool provides search access to Graphiti knowledge graph
type graphitiSearchTool struct {
	flowID         int64
	taskID         *int64
	subtaskID      *int64
	groupID        string
	graphitiClient GraphitiSearcher
}

// NewGraphitiSearchTool creates a new Graphiti search tool
func NewGraphitiSearchTool(
	flowID int64,
	taskID, subtaskID *int64,
	groupID string,
	graphitiClient GraphitiSearcher,
) Tool {
	return &graphitiSearchTool{
		flowID:         flowID,
		taskID:         taskID,
		subtaskID:      subtaskID,
		groupID:        groupID,
		graphitiClient: graphitiClient,
	}
}

// IsAvailable checks if the tool is available
func (t *graphitiSearchTool) IsAvailable() bool {
	return t.graphitiClient != nil && t.graphitiClient.IsEnabled()
}

// Handle executes the search based on search_type
func (t *graphitiSearchTool) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if !t.IsAvailable() {
		return "Graphiti knowledge graph is not enabled. No historical context or memory data is available for this search.", nil
	}

	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(t.flowID, t.taskID, t.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	var searchArgs GraphitiSearchAction
	if err := json.Unmarshal(args, &searchArgs); err != nil {
		logger.WithError(err).Error("failed to unmarshal search arguments")
		return "", fmt.Errorf("failed to unmarshal search arguments: %w", err)
	}

	searchArgs.Query = strings.TrimSpace(searchArgs.Query)

	if searchArgs.Query == "" {
		logger.Error("query parameter is required")
		return "", fmt.Errorf("query parameter is required")
	}
	if searchArgs.SearchType == "" {
		logger.Error("search_type parameter is required")
		return "", fmt.Errorf("search_type parameter is required")
	}

	ctx, observation := obs.Observer.NewObservation(ctx)

	retrieverTitle, ok := graphitiRetrieverTitles[searchArgs.SearchType.String()]
	if !ok {
		retrieverTitle = "retrieve context from graphiti knowledge graph"
	}

	retriever := observation.Retriever(
		langfuse.WithRetrieverName(retrieverTitle),
		langfuse.WithRetrieverInput(searchArgs.retrieverInput(t.groupID)),
		langfuse.WithRetrieverMetadata(langfuse.Metadata{
			"tool_name":   name,
			"engine":      "graphiti",
			"search_type": searchArgs.SearchType,
			"group_id":    t.groupID,
			"query":       searchArgs.Query,
		}),
	)
	ctx, observation = retriever.Observation(ctx)

	// Nest Graphiti server-side observations under the retriever observation
	observationObject := &graphiti.Observation{
		ID:      observation.ID(),
		TraceID: observation.TraceID(),
		Time:    time.Now().UTC(),
	}

	var (
		err    error
		result string
	)
	switch searchArgs.SearchType {
	case "temporal_window":
		result, err = t.handleTemporalWindowSearch(ctx, t.groupID, searchArgs, observationObject)
	case "entity_relationships":
		result, err = t.handleEntityRelationshipsSearch(ctx, t.groupID, searchArgs, observationObject)
	case "diverse_results":
		result, err = t.handleDiverseResultsSearch(ctx, t.groupID, searchArgs, observationObject)
	case "episode_context":
		result, err = t.handleEpisodeContextSearch(ctx, t.groupID, searchArgs, observationObject)
	case "successful_tools":
		result, err = t.handleSuccessfulToolsSearch(ctx, t.groupID, searchArgs, observationObject)
	case "recent_context":
		result, err = t.handleRecentContextSearch(ctx, t.groupID, searchArgs, observationObject)
	case "entity_by_label":
		result, err = t.handleEntityByLabelSearch(ctx, t.groupID, searchArgs, observationObject)
	default:
		err = fmt.Errorf("unknown search_type: %s", searchArgs.SearchType)
	}

	if err != nil {
		// Transport-level failures (connection refused, DNS, timeout, TLS handshake
		// timeout, context deadline exceeded) surface from the underlying http.Client
		// as *url.Error. This is not a malformed-arguments problem, so it must not be
		// routed through the tool-call arg-fixer (which cannot fix a network outage
		// and would burn 3 retries doing so) — degrade gracefully instead, matching
		// the terminal/browser tools' handling of the same class of failure.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			softMsg := fmt.Sprintf(
				"Graphiti knowledge graph is temporarily unavailable (%s); continuing without historical context.",
				truncateString(err.Error(), maxGraphitiResponsePreviewBytes),
			)
			retriever.End(
				langfuse.WithRetrieverStatus(softMsg),
				langfuse.WithRetrieverLevel(langfuse.ObservationLevelWarning),
			)
			logger.WithError(err).Warnf("graphiti search '%s' unavailable, degrading gracefully", searchArgs.SearchType)
			return softMsg, nil
		}

		// The Graphiti server responded (connection succeeded), but with a
		// non-2xx status. graphiti-go-client has no typed error for this - it
		// embeds the raw, unbounded response body in a plain fmt.Errorf - so
		// recover the status/body via the one fixed message shape it emits.
		if m := graphitiAPIStatusErrorRe.FindStringSubmatch(err.Error()); m != nil {
			statusCode, convErr := strconv.Atoi(m[1])
			body := truncateString(strings.TrimSpace(m[2]), maxGraphitiResponsePreviewBytes)

			if convErr == nil && statusCode >= 500 {
				// A 5xx from Graphiti's own backend is the same class of
				// problem as a transport failure: not fixable by editing
				// arguments, so degrade gracefully and show the LLM whatever
				// of the response body we could read.
				softMsg := fmt.Sprintf(
					"Graphiti knowledge graph returned HTTP %d and is likely temporarily unavailable; continuing without historical context. Response: %s",
					statusCode, body,
				)
				retriever.End(
					langfuse.WithRetrieverStatus(softMsg),
					langfuse.WithRetrieverLevel(langfuse.ObservationLevelWarning),
				)
				logger.WithError(err).Warnf("graphiti search '%s' returned HTTP %d, degrading gracefully", searchArgs.SearchType, statusCode)
				return softMsg, nil
			}

			// 4xx (or an unparseable status) likely means our own request was
			// malformed - stays a hard failure so the tool-call fixer can act
			// on it, but capped so a verbose error page can't blow up its context.
			err = fmt.Errorf("graphiti API request failed with status %s: %s", m[1], body)
		}

		retriever.End(
			langfuse.WithRetrieverStatus(err.Error()),
			langfuse.WithRetrieverLevel(langfuse.ObservationLevelError),
		)
		logger.WithError(err).Errorf("failed to perform graphiti search '%s'", searchArgs.SearchType)
		return "", err
	}

	retriever.End(
		langfuse.WithRetrieverStatus("success"),
		langfuse.WithRetrieverLevel(langfuse.ObservationLevelDebug),
		langfuse.WithRetrieverOutput(result),
	)

	return result, nil
}

// retrieverInput builds the langfuse retriever input payload, including only
// the search parameters relevant to the requested search_type.
func (args GraphitiSearchAction) retrieverInput(groupID string) map[string]any {
	input := map[string]any{
		"query":       args.Query,
		"search_type": args.SearchType,
		"group_id":    groupID,
	}
	if args.MaxResults != nil {
		input["max_results"] = args.MaxResults.Int()
	}
	if args.TimeStart != "" {
		input["time_start"] = args.TimeStart
	}
	if args.TimeEnd != "" {
		input["time_end"] = args.TimeEnd
	}
	if args.CenterNodeUUID != "" {
		input["center_node_uuid"] = args.CenterNodeUUID
	}
	if args.MaxDepth != nil {
		input["max_depth"] = args.MaxDepth.Int()
	}
	if len(args.NodeLabels) > 0 {
		input["node_labels"] = args.NodeLabels
	}
	if len(args.EdgeTypes) > 0 {
		input["edge_types"] = args.EdgeTypes
	}
	if args.DiversityLevel != "" {
		input["diversity_level"] = args.DiversityLevel
	}
	if args.MinMentions != nil {
		input["min_mentions"] = args.MinMentions.Int()
	}
	if args.RecencyWindow != "" {
		input["recency_window"] = args.RecencyWindow
	}
	return input
}

// handleTemporalWindowSearch performs time-bounded search
func (t *graphitiSearchTool) handleTemporalWindowSearch(
	ctx context.Context,
	groupID string,
	args GraphitiSearchAction,
	observationObject *graphiti.Observation,
) (string, error) {
	// Validate temporal parameters
	if args.TimeStart == "" || args.TimeEnd == "" {
		return "", fmt.Errorf("time_start and time_end are required for temporal_window search")
	}

	timeStart, err := parseGraphitiTime(args.TimeStart)
	if err != nil {
		return "", fmt.Errorf("invalid time_start format (use ISO 8601, e.g. 2026-01-02T15:04:05Z): %w", err)
	}

	timeEnd, err := parseGraphitiTime(args.TimeEnd)
	if err != nil {
		return "", fmt.Errorf("invalid time_end format (use ISO 8601, e.g. 2026-01-02T15:04:05Z): %w", err)
	}

	if timeEnd.Before(timeStart) {
		return "", fmt.Errorf("time_end must be after time_start")
	}

	maxResults := args.MaxResults.Int()
	if maxResults <= 0 {
		maxResults = DefaultTemporalMaxResults
	}

	req := graphiti.TemporalSearchRequest{
		Query:       args.Query,
		GroupID:     &groupID,
		TimeStart:   timeStart,
		TimeEnd:     timeEnd,
		MaxResults:  maxResults,
		Observation: observationObject,
	}

	resp, err := t.graphitiClient.TemporalWindowSearch(ctx, req)
	if err != nil {
		return "", fmt.Errorf("temporal window search failed: %w", err)
	}

	return FormatGraphitiTemporalResults(resp, args.Query), nil
}

// handleEntityRelationshipsSearch finds relationships from a center node
func (t *graphitiSearchTool) handleEntityRelationshipsSearch(
	ctx context.Context,
	groupID string,
	args GraphitiSearchAction,
	observationObject *graphiti.Observation,
) (string, error) {
	if args.CenterNodeUUID == "" {
		return "", fmt.Errorf("center_node_uuid is required for entity_relationships search")
	}
	if _, err := uuid.Parse(args.CenterNodeUUID); err != nil {
		return "", fmt.Errorf(
			"center_node_uuid must be a valid UUID copied verbatim from the 'UUID:' field of a prior "+
				"graphiti_search result, got %q", args.CenterNodeUUID,
		)
	}

	maxResults := args.MaxResults.Int()
	if maxResults <= 0 {
		maxResults = DefaultRelationshipMaxResults
	}

	maxDepth := args.MaxDepth.Int()
	if maxDepth <= 0 {
		maxDepth = DefaultMaxDepth
	}
	if maxDepth > 3 {
		maxDepth = 3
	}

	var nodeLabels *[]string
	if len(args.NodeLabels) > 0 {
		nodeLabels = &args.NodeLabels
	}

	var edgeTypes *[]string
	if len(args.EdgeTypes) > 0 {
		edgeTypes = &args.EdgeTypes
	}

	req := graphiti.EntityRelationshipSearchRequest{
		Query:          args.Query,
		GroupID:        &groupID,
		CenterNodeUUID: args.CenterNodeUUID,
		MaxDepth:       maxDepth,
		NodeLabels:     nodeLabels,
		EdgeTypes:      edgeTypes,
		MaxResults:     maxResults,
		Observation:    observationObject,
	}

	resp, err := t.graphitiClient.EntityRelationshipsSearch(ctx, req)
	if err != nil {
		return "", fmt.Errorf("entity relationships search failed: %w", err)
	}

	return FormatGraphitiEntityRelationshipResults(resp, args.Query), nil
}

// handleDiverseResultsSearch gets diverse, non-redundant results
func (t *graphitiSearchTool) handleDiverseResultsSearch(
	ctx context.Context,
	groupID string,
	args GraphitiSearchAction,
	observationObject *graphiti.Observation,
) (string, error) {
	maxResults := args.MaxResults.Int()
	if maxResults <= 0 {
		maxResults = DefaultDiverseMaxResults
	}

	diversityLevel := args.DiversityLevel.String()
	if diversityLevel == "" {
		diversityLevel = DefaultDiversityLevel
	}
	if _, ok := allowedDiversityLevels[diversityLevel]; !ok {
		return "", fmt.Errorf("invalid diversity_level: %s", diversityLevel)
	}

	req := graphiti.DiverseSearchRequest{
		Query:          args.Query,
		GroupID:        &groupID,
		DiversityLevel: diversityLevel,
		MaxResults:     maxResults,
		Observation:    observationObject,
	}

	resp, err := t.graphitiClient.DiverseResultsSearch(ctx, req)
	if err != nil {
		return "", fmt.Errorf("diverse results search failed: %w", err)
	}

	return FormatGraphitiDiverseResults(resp, args.Query), nil
}

// handleEpisodeContextSearch searches through agent responses and tool execution records
func (t *graphitiSearchTool) handleEpisodeContextSearch(
	ctx context.Context,
	groupID string,
	args GraphitiSearchAction,
	observationObject *graphiti.Observation,
) (string, error) {
	maxResults := args.MaxResults.Int()
	if maxResults <= 0 {
		maxResults = DefaultEpisodeMaxResults
	}

	req := graphiti.EpisodeContextSearchRequest{
		Query:       args.Query,
		GroupID:     &groupID,
		MaxResults:  maxResults,
		Observation: observationObject,
	}

	resp, err := t.graphitiClient.EpisodeContextSearch(ctx, req)
	if err != nil {
		return "", fmt.Errorf("episode context search failed: %w", err)
	}

	return FormatGraphitiEpisodeContextResults(resp, args.Query), nil
}

// handleSuccessfulToolsSearch finds successful tool executions and attack patterns
func (t *graphitiSearchTool) handleSuccessfulToolsSearch(
	ctx context.Context,
	groupID string,
	args GraphitiSearchAction,
	observationObject *graphiti.Observation,
) (string, error) {
	maxResults := args.MaxResults.Int()
	if maxResults <= 0 {
		maxResults = DefaultSuccessfulMaxResults
	}

	minMentions := args.MinMentions.Int()
	if minMentions <= 0 {
		minMentions = DefaultMinMentions
	}

	req := graphiti.SuccessfulToolsSearchRequest{
		Query:       args.Query,
		GroupID:     &groupID,
		MinMentions: minMentions,
		MaxResults:  maxResults,
		Observation: observationObject,
	}

	resp, err := t.graphitiClient.SuccessfulToolsSearch(ctx, req)
	if err != nil {
		return "", fmt.Errorf("successful tools search failed: %w", err)
	}

	return FormatGraphitiSuccessfulToolsResults(resp, args.Query), nil
}

// handleRecentContextSearch retrieves recent relevant context
func (t *graphitiSearchTool) handleRecentContextSearch(
	ctx context.Context,
	groupID string,
	args GraphitiSearchAction,
	observationObject *graphiti.Observation,
) (string, error) {
	maxResults := args.MaxResults.Int()
	if maxResults <= 0 {
		maxResults = DefaultRecentMaxResults
	}

	recencyWindow := args.RecencyWindow.String()
	if recencyWindow == "" {
		recencyWindow = DefaultRecencyWindow
	}
	if _, ok := allowedRecencyWindows[recencyWindow]; !ok {
		return "", fmt.Errorf("invalid recency_window: %s", recencyWindow)
	}

	req := graphiti.RecentContextSearchRequest{
		Query:         args.Query,
		GroupID:       &groupID,
		RecencyWindow: recencyWindow,
		MaxResults:    maxResults,
		Observation:   observationObject,
	}

	resp, err := t.graphitiClient.RecentContextSearch(ctx, req)
	if err != nil {
		return "", fmt.Errorf("recent context search failed: %w", err)
	}

	return FormatGraphitiRecentContextResults(resp, args.Query), nil
}

// handleEntityByLabelSearch searches for entities by label/type
func (t *graphitiSearchTool) handleEntityByLabelSearch(
	ctx context.Context,
	groupID string,
	args GraphitiSearchAction,
	observationObject *graphiti.Observation,
) (string, error) {
	if len(args.NodeLabels) == 0 {
		return "", fmt.Errorf(
			"node_labels is required for entity_by_label search: pass one or more EXACT taxonomy node names " +
				`(PascalCase singular), e.g. node_labels: ["Host", "Service", "Vulnerability"]`,
		)
	}

	maxResults := args.MaxResults.Int()
	if maxResults <= 0 {
		maxResults = DefaultLabelMaxResults
	}

	var edgeTypes *[]string
	if len(args.EdgeTypes) > 0 {
		edgeTypes = &args.EdgeTypes
	}

	req := graphiti.EntityByLabelSearchRequest{
		Query:       args.Query,
		GroupID:     &groupID,
		NodeLabels:  args.NodeLabels,
		EdgeTypes:   edgeTypes,
		MaxResults:  maxResults,
		Observation: observationObject,
	}

	resp, err := t.graphitiClient.EntityByLabelSearch(ctx, req)
	if err != nil {
		return "", fmt.Errorf("entity by label search failed: %w", err)
	}

	return FormatGraphitiEntityByLabelResults(resp, args.Query), nil
}

// FormatGraphitiTemporalResults formats results for agent consumption
func FormatGraphitiTemporalResults(
	resp *graphiti.TemporalSearchResponse,
	query string,
) string {
	var builder strings.Builder

	builder.WriteString("# Temporal Search Results\n\n")
	builder.WriteString(fmt.Sprintf("**Query:** %s\n\n", query))
	builder.WriteString(fmt.Sprintf("**Time Window:** %s to %s\n\n",
		resp.TimeWindow.Start.Format(time.RFC3339),
		resp.TimeWindow.End.Format(time.RFC3339)))

	// Format edges (facts/relationships)
	if len(resp.Edges) > 0 {
		builder.WriteString("## Facts & Relationships\n\n")
		for i, edge := range resp.Edges {
			score := ""
			if i < len(resp.EdgeScores) {
				score = fmt.Sprintf(" (score: %.3f)", resp.EdgeScores[i])
			}
			builder.WriteString(fmt.Sprintf("%d. **%s**%s\n", i+1, edge.Name, score))
			builder.WriteString(fmt.Sprintf("   - Fact: %s\n", edge.Fact))
			builder.WriteString(fmt.Sprintf("   - Created: %s\n", edge.CreatedAt.Format(time.RFC3339)))
			if edge.ValidAt != nil {
				builder.WriteString(fmt.Sprintf("   - Valid At: %s\n", edge.ValidAt.Format(time.RFC3339)))
			}
			builder.WriteString("\n")
		}
	}

	// Format nodes (entities)
	if len(resp.Nodes) > 0 {
		builder.WriteString("## Entities\n\n")
		for i, node := range resp.Nodes {
			score := ""
			if i < len(resp.NodeScores) {
				score = fmt.Sprintf(" (score: %.3f)", resp.NodeScores[i])
			}
			builder.WriteString(fmt.Sprintf("%d. **%s**%s\n", i+1, node.Name, score))
			builder.WriteString(fmt.Sprintf("   - UUID: %s\n", node.UUID))
			builder.WriteString(fmt.Sprintf("   - Labels: %v\n", node.Labels))
			builder.WriteString(fmt.Sprintf("   - Summary: %s\n", node.Summary))
			if len(node.Attributes) > 0 {
				builder.WriteString(fmt.Sprintf("   - Attributes: %v\n", node.Attributes))
			}
			builder.WriteString("\n")
		}
	}

	// Format episodes (agent responses & tool executions)
	if len(resp.Episodes) > 0 {
		builder.WriteString("## Agent Responses & Tool Executions\n\n")
		for i, episode := range resp.Episodes {
			score := ""
			if i < len(resp.EpisodeScores) {
				score = fmt.Sprintf(" (score: %.3f)", resp.EpisodeScores[i])
			}
			builder.WriteString(fmt.Sprintf("%d. **%s**%s\n", i+1, episode.Source, score))
			builder.WriteString(fmt.Sprintf("   - Description: %s\n", episode.SourceDescription))
			builder.WriteString(fmt.Sprintf("   - Created: %s\n", episode.CreatedAt.Format(time.RFC3339)))
			builder.WriteString(fmt.Sprintf("   - Content:\n```\n%s\n```\n", episode.Content))
			builder.WriteString("\n")
		}
	}

	if len(resp.Edges) == 0 && len(resp.Nodes) == 0 && len(resp.Episodes) == 0 {
		builder.WriteString("No results found in the specified time window.\n")
	}

	return builder.String()
}

// FormatGraphitiEntityRelationshipResults formats entity relationship results
func FormatGraphitiEntityRelationshipResults(
	resp *graphiti.EntityRelationshipSearchResponse,
	query string,
) string {
	var builder strings.Builder

	builder.WriteString("# Entity Relationship Search Results\n\n")
	builder.WriteString(fmt.Sprintf("**Query:** %s\n\n", query))

	if resp.CenterNode != nil {
		builder.WriteString(fmt.Sprintf("## Center Node: %s\n", resp.CenterNode.Name))
		builder.WriteString(fmt.Sprintf("- UUID: %s\n", resp.CenterNode.UUID))
		builder.WriteString(fmt.Sprintf("- Summary: %s\n\n", resp.CenterNode.Summary))
	}

	if len(resp.Edges) > 0 {
		builder.WriteString("## Related Facts & Relationships\n\n")
		for i, edge := range resp.Edges {
			dist := ""
			if i < len(resp.EdgeDistances) {
				dist = fmt.Sprintf(" (distance: %.3f)", resp.EdgeDistances[i])
			}
			builder.WriteString(fmt.Sprintf("%d. **%s**%s\n", i+1, edge.Name, dist))
			builder.WriteString(fmt.Sprintf("   - Fact: %s\n", edge.Fact))
			builder.WriteString(fmt.Sprintf("   - Source: %s\n", edge.SourceNodeUUID))
			builder.WriteString(fmt.Sprintf("   - Target: %s\n", edge.TargetNodeUUID))
			builder.WriteString("\n")
		}
	}

	if len(resp.Nodes) > 0 {
		builder.WriteString("## Related Entities\n\n")
		for i, node := range resp.Nodes {
			dist := ""
			if i < len(resp.NodeDistances) {
				dist = fmt.Sprintf(" (distance: %.3f)", resp.NodeDistances[i])
			}
			builder.WriteString(fmt.Sprintf("%d. **%s**%s\n", i+1, node.Name, dist))
			builder.WriteString(fmt.Sprintf("   - UUID: %s\n", node.UUID))
			builder.WriteString(fmt.Sprintf("   - Labels: %v\n", node.Labels))
			builder.WriteString(fmt.Sprintf("   - Summary: %s\n", node.Summary))
			builder.WriteString("\n")
		}
	}

	if len(resp.Edges) == 0 && len(resp.Nodes) == 0 {
		builder.WriteString("No relationships found matching criteria.\n")
	}

	return builder.String()
}

// FormatGraphitiDiverseResults formats diverse results
func FormatGraphitiDiverseResults(
	resp *graphiti.DiverseSearchResponse,
	query string,
) string {
	var builder strings.Builder

	builder.WriteString("# Diverse Search Results\n\n")
	builder.WriteString(fmt.Sprintf("**Query:** %s\n\n", query))

	if len(resp.Communities) > 0 {
		builder.WriteString("## Communities (Context Clusters)\n\n")
		for i, comm := range resp.Communities {
			score := ""
			if i < len(resp.CommunityMMRScores) {
				score = fmt.Sprintf(" (MMR score: %.3f)", resp.CommunityMMRScores[i])
			}
			builder.WriteString(fmt.Sprintf("%d. **%s**%s\n", i+1, comm.Name, score))
			builder.WriteString(fmt.Sprintf("   - UUID: %s\n", comm.UUID))
			builder.WriteString(fmt.Sprintf("   - Summary: %s\n\n", comm.Summary))
		}
	}

	if len(resp.Edges) > 0 {
		builder.WriteString("## Diverse Facts\n\n")
		for i, edge := range resp.Edges {
			score := ""
			if i < len(resp.EdgeMMRScores) {
				score = fmt.Sprintf(" (MMR score: %.3f)", resp.EdgeMMRScores[i])
			}
			builder.WriteString(fmt.Sprintf("%d. **%s**%s\n", i+1, edge.Name, score))
			builder.WriteString(fmt.Sprintf("   - Fact: %s\n\n", edge.Fact))
		}
	}

	if len(resp.Episodes) > 0 {
		builder.WriteString("## Diverse Agent Activity\n\n")
		for i, ep := range resp.Episodes {
			score := ""
			if i < len(resp.EpisodeScores) { // Using raw scores for episodes as MMR scores might not be available in same format
				score = fmt.Sprintf(" (score: %.3f)", resp.EpisodeScores[i])
			}
			builder.WriteString(fmt.Sprintf("%d. **%s**%s\n", i+1, ep.Source, score))
			builder.WriteString(fmt.Sprintf("   - Description: %s\n", ep.SourceDescription))
			builder.WriteString(fmt.Sprintf("   - Content: %s\n\n", truncate(ep.Content, 200)))
		}
	}

	return builder.String()
}

// FormatGraphitiEpisodeContextResults formats episode context results
func FormatGraphitiEpisodeContextResults(
	resp *graphiti.EpisodeContextSearchResponse,
	query string,
) string {
	var builder strings.Builder

	builder.WriteString("# Episode Context Results\n\n")
	builder.WriteString(fmt.Sprintf("**Query:** %s\n\n", query))

	if len(resp.Episodes) > 0 {
		builder.WriteString("## Relevant Agent Activity\n\n")
		for i, ep := range resp.Episodes {
			score := ""
			if i < len(resp.RerankerScores) {
				score = fmt.Sprintf(" (relevance: %.3f)", resp.RerankerScores[i])
			}
			builder.WriteString(fmt.Sprintf("%d. **%s**%s\n", i+1, ep.Source, score))
			builder.WriteString(fmt.Sprintf("   - Time: %s\n", ep.CreatedAt.Format(time.RFC3339)))
			builder.WriteString(fmt.Sprintf("   - Description: %s\n", ep.SourceDescription))
			builder.WriteString(fmt.Sprintf("   - Content:\n```\n%s\n```\n\n", ep.Content))
		}
	}

	if len(resp.MentionedNodes) > 0 {
		builder.WriteString("## Mentioned Entities\n\n")
		for i, node := range resp.MentionedNodes {
			score := ""
			if i < len(resp.MentionedNodeScores) {
				score = fmt.Sprintf(" (relevance: %.3f)", resp.MentionedNodeScores[i])
			}
			builder.WriteString(fmt.Sprintf("- **%s**%s (UUID: %s): %s\n", node.Name, score, node.UUID, node.Summary))
		}
	}

	if len(resp.Episodes) == 0 {
		builder.WriteString("No episode context found.\n")
	}

	return builder.String()
}

// FormatGraphitiSuccessfulToolsResults formats successful tools results
func FormatGraphitiSuccessfulToolsResults(
	resp *graphiti.SuccessfulToolsSearchResponse,
	query string,
) string {
	var builder strings.Builder

	builder.WriteString("# Successful Tools & Techniques\n\n")
	builder.WriteString(fmt.Sprintf("**Query:** %s\n\n", query))

	if len(resp.Episodes) > 0 {
		builder.WriteString("## Successful Executions\n\n")
		for i, ep := range resp.Episodes {
			score := ""
			if i < len(resp.EpisodeScores) {
				score = fmt.Sprintf(" (score: %.3f)", resp.EpisodeScores[i])
			}
			builder.WriteString(fmt.Sprintf("%d. **%s**%s\n", i+1, ep.Source, score))
			builder.WriteString(fmt.Sprintf("   - Description: %s\n", ep.SourceDescription))
			builder.WriteString(fmt.Sprintf("   - Command/Output:\n```\n%s\n```\n\n", ep.Content))
		}
	}

	if len(resp.Edges) > 0 {
		builder.WriteString("## Related Facts (Success Indicators)\n\n")
		for i, edge := range resp.Edges {
			count := ""
			if i < len(resp.EdgeMentionCounts) {
				count = fmt.Sprintf(" (mentions: %.0f)", resp.EdgeMentionCounts[i])
			}
			builder.WriteString(fmt.Sprintf("- **%s**%s: %s\n", edge.Name, count, edge.Fact))
		}
	}

	if len(resp.Episodes) == 0 {
		builder.WriteString("No successful tool executions found matching criteria.\n")
	}

	return builder.String()
}

// FormatGraphitiRecentContextResults formats recent context results
func FormatGraphitiRecentContextResults(
	resp *graphiti.RecentContextSearchResponse,
	query string,
) string {
	var builder strings.Builder

	builder.WriteString("# Recent Context\n\n")
	builder.WriteString(fmt.Sprintf("**Query:** %s\n\n", query))
	builder.WriteString(fmt.Sprintf("**Time Window:** %s to %s\n\n",
		resp.TimeWindow.Start.Format(time.RFC3339),
		resp.TimeWindow.End.Format(time.RFC3339)))

	if len(resp.Nodes) > 0 {
		builder.WriteString("## Recently Discovered Entities\n\n")
		for i, node := range resp.Nodes {
			score := ""
			if i < len(resp.NodeScores) {
				score = fmt.Sprintf(" (score: %.3f)", resp.NodeScores[i])
			}
			builder.WriteString(fmt.Sprintf("%d. **%s**%s\n", i+1, node.Name, score))
			builder.WriteString(fmt.Sprintf("   - UUID: %s\n", node.UUID))
			builder.WriteString(fmt.Sprintf("   - Labels: %v\n", node.Labels))
			builder.WriteString(fmt.Sprintf("   - Summary: %s\n\n", node.Summary))
		}
	}

	if len(resp.Edges) > 0 {
		builder.WriteString("## Recent Facts\n\n")
		for i, edge := range resp.Edges {
			score := ""
			if i < len(resp.EdgeScores) {
				score = fmt.Sprintf(" (score: %.3f)", resp.EdgeScores[i])
			}
			builder.WriteString(fmt.Sprintf("- **%s**%s: %s\n", edge.Name, score, edge.Fact))
		}
	}

	if len(resp.Episodes) > 0 {
		builder.WriteString("## Recent Activity\n\n")
		for i, ep := range resp.Episodes {
			score := ""
			if i < len(resp.EpisodeScores) {
				score = fmt.Sprintf(" (score: %.3f)", resp.EpisodeScores[i])
			}
			builder.WriteString(fmt.Sprintf("- **%s**%s: %s\n", ep.Source, score, ep.SourceDescription))
		}
	}

	if len(resp.Nodes) == 0 && len(resp.Edges) == 0 && len(resp.Episodes) == 0 {
		builder.WriteString("No recent context found in the specified window.\n")
	}

	return builder.String()
}

// FormatGraphitiEntityByLabelResults formats entity by label results
func FormatGraphitiEntityByLabelResults(
	resp *graphiti.EntityByLabelSearchResponse,
	query string,
) string {
	var builder strings.Builder

	builder.WriteString("# Entity Inventory Search\n\n")
	builder.WriteString(fmt.Sprintf("**Query:** %s\n\n", query))

	if len(resp.Nodes) > 0 {
		builder.WriteString("## Matching Entities\n\n")
		for i, node := range resp.Nodes {
			score := ""
			if i < len(resp.NodeScores) {
				score = fmt.Sprintf(" (score: %.3f)", resp.NodeScores[i])
			}
			builder.WriteString(fmt.Sprintf("%d. **%s**%s\n", i+1, node.Name, score))
			builder.WriteString(fmt.Sprintf("   - UUID: %s\n", node.UUID))
			builder.WriteString(fmt.Sprintf("   - Labels: %v\n", node.Labels))
			builder.WriteString(fmt.Sprintf("   - Summary: %s\n", node.Summary))
			if len(node.Attributes) > 0 {
				builder.WriteString(fmt.Sprintf("   - Attributes: %v\n", node.Attributes))
			}
			builder.WriteString("\n")
		}
	}

	if len(resp.Edges) > 0 {
		builder.WriteString("## Associated Facts\n\n")
		for i, edge := range resp.Edges {
			score := ""
			if i < len(resp.EdgeScores) {
				score = fmt.Sprintf(" (score: %.3f)", resp.EdgeScores[i])
			}
			builder.WriteString(fmt.Sprintf("- **%s**%s: %s\n", edge.Name, score, edge.Fact))
		}
	}

	if len(resp.Nodes) == 0 {
		builder.WriteString("No entities found matching the specified labels/query.\n")
	}

	return builder.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
