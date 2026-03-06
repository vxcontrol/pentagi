package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"pentagi/pkg/graphiti"
	obs "pentagi/pkg/observability"

	"github.com/sirupsen/logrus"
)

type graphitiSearcher interface {
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
)

// GraphitiSearchTool provides search access to Graphiti knowledge graph
type GraphitiSearchTool struct {
	flowID         int64
	taskID         *int64
	subtaskID      *int64
	graphitiClient graphitiSearcher
}

// NewGraphitiSearchTool creates a new Graphiti search tool
func NewGraphitiSearchTool(
	flowID int64,
	taskID, subtaskID *int64,
	graphitiClient graphitiSearcher,
) *GraphitiSearchTool {
	return &GraphitiSearchTool{
		flowID:         flowID,
		taskID:         taskID,
		subtaskID:      subtaskID,
		graphitiClient: graphitiClient,
	}
}

// IsAvailable checks if the tool is available
func (t *GraphitiSearchTool) IsAvailable() bool {
	return t.graphitiClient != nil && t.graphitiClient.IsEnabled()
}

// Handle executes the search based on search_type
func (t *GraphitiSearchTool) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
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

	// Validate required parameters
	if searchArgs.Query == "" {
		logger.Error("query parameter is required")
		return "", fmt.Errorf("query parameter is required")
	}
	if searchArgs.SearchType == "" {
		logger.Error("search_type parameter is required")
		return "", fmt.Errorf("search_type parameter is required")
	}

	ctx, observation := obs.Observer.NewObservation(ctx)
	observationObject := &graphiti.Observation{
		ID:      observation.ID(),
		TraceID: observation.TraceID(),
		Time:    time.Now().UTC(),
	}

	// Get group ID from flow context
	groupID := fmt.Sprintf("flow-%d", t.flowID)

	// Route to appropriate search method
	var (
		err    error
		result string
	)
	switch searchArgs.SearchType {
	case "temporal_window":
		result, err = t.handleTemporalWindowSearch(ctx, groupID, searchArgs, observationObject)
	case "entity_relationships":
		result, err = t.handleEntityRelationshipsSearch(ctx, groupID, searchArgs, observationObject)
	case "diverse_results":
		result, err = t.handleDiverseResultsSearch(ctx, groupID, searchArgs, observationObject)
	case "episode_context":
		result, err = t.handleEpisodeContextSearch(ctx, groupID, searchArgs, observationObject)
	case "successful_tools":
		result, err = t.handleSuccessfulToolsSearch(ctx, groupID, searchArgs, observationObject)
	case "recent_context":
		result, err = t.handleRecentContextSearch(ctx, groupID, searchArgs, observationObject)
	case "entity_by_label":
		result, err = t.handleEntityByLabelSearch(ctx, groupID, searchArgs, observationObject)
	default:
		err = fmt.Errorf("unknown search_type: %s", searchArgs.SearchType)
	}

	if err != nil {
		logger.WithError(err).Errorf("failed to perform graphiti search '%s'", searchArgs.SearchType)
		return "", err
	}

	return result, nil
}

// handleTemporalWindowSearch performs time-bounded search
func (t *GraphitiSearchTool) handleTemporalWindowSearch(
	ctx context.Context,
	groupID string,
	args GraphitiSearchAction,
	observationObject *graphiti.Observation,
) (string, error) {
	// Validate temporal parameters
	if args.TimeStart == "" || args.TimeEnd == "" {
		return "", fmt.Errorf("time_start and time_end are required for temporal_window search")
	}

	timeStart, err := time.Parse(time.RFC3339, args.TimeStart)
	if err != nil {
		return "", fmt.Errorf("invalid time_start format (use ISO 8601): %w", err)
	}

	timeEnd, err := time.Parse(time.RFC3339, args.TimeEnd)
	if err != nil {
		return "", fmt.Errorf("invalid time_end format (use ISO 8601): %w", err)
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
func (t *GraphitiSearchTool) handleEntityRelationshipsSearch(
	ctx context.Context,
	groupID string,
	args GraphitiSearchAction,
	observationObject *graphiti.Observation,
) (string, error) {
	if args.CenterNodeUUID == "" {
		return "", fmt.Errorf("center_node_uuid is required for entity_relationships search")
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
func (t *GraphitiSearchTool) handleDiverseResultsSearch(
	ctx context.Context,
	groupID string,
	args GraphitiSearchAction,
	observationObject *graphiti.Observation,
) (string, error) {
	maxResults := args.MaxResults.Int()
	if maxResults <= 0 {
		maxResults = DefaultDiverseMaxResults
	}

	diversityLevel := args.DiversityLevel
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
func (t *GraphitiSearchTool) handleEpisodeContextSearch(
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
func (t *GraphitiSearchTool) handleSuccessfulToolsSearch(
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
func (t *GraphitiSearchTool) handleRecentContextSearch(
	ctx context.Context,
	groupID string,
	args GraphitiSearchAction,
	observationObject *graphiti.Observation,
) (string, error) {
	maxResults := args.MaxResults.Int()
	if maxResults <= 0 {
		maxResults = DefaultRecentMaxResults
	}

	recencyWindow := args.RecencyWindow
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
func (t *GraphitiSearchTool) handleEntityByLabelSearch(
	ctx context.Context,
	groupID string,
	args GraphitiSearchAction,
	observationObject *graphiti.Observation,
) (string, error) {
	if len(args.NodeLabels) == 0 {
		return "", fmt.Errorf("node_labels is required for entity_by_label search")
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
			builder.WriteString(fmt.Sprintf("- **%s**%s: %s\n", node.Name, score, node.Summary))
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
