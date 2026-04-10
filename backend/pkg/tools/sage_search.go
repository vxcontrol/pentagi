package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"pentagi/pkg/sage"
	obs "pentagi/pkg/observability"

	"github.com/sirupsen/logrus"
)

// SageSearcher provides an interface for SAGE persistent memory operations
type SageSearcher interface {
	IsEnabled() bool
	Recall(ctx context.Context, req sage.RecallRequest) (*sage.RecallResponse, error)
	Remember(ctx context.Context, req sage.RememberRequest) (*sage.RememberResponse, error)
}

// sageSearchTool provides search and store access to SAGE persistent memory
type sageSearchTool struct {
	flowID     int64
	taskID     *int64
	subtaskID  *int64
	sageClient SageSearcher
}

// NewSageSearchTool creates a new SAGE search tool
func NewSageSearchTool(flowID int64, taskID, subtaskID *int64, sageClient SageSearcher) Tool {
	return &sageSearchTool{
		flowID:     flowID,
		taskID:     taskID,
		subtaskID:  subtaskID,
		sageClient: sageClient,
	}
}

// IsAvailable checks if the tool is available
func (t *sageSearchTool) IsAvailable() bool {
	return t.sageClient != nil && t.sageClient.IsEnabled()
}

// Handle executes the SAGE operation based on tool name
func (t *sageSearchTool) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if !t.IsAvailable() {
		return "SAGE persistent memory is not enabled. Cross-session knowledge is not available.", nil
	}

	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(t.flowID, t.taskID, t.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	ctx, _ = obs.Observer.NewObservation(ctx)

	switch name {
	case SageRecallToolName:
		return t.handleRecall(ctx, logger, args)
	case SageRememberToolName:
		return t.handleRemember(ctx, logger, args)
	default:
		logger.Error("unknown SAGE tool")
		return "", fmt.Errorf("unknown SAGE tool: %s", name)
	}
}

// handleRecall searches SAGE persistent memory for cross-session knowledge
func (t *sageSearchTool) handleRecall(ctx context.Context, logger *logrus.Entry, args json.RawMessage) (string, error) {
	var action SageRecallAction
	if err := json.Unmarshal(args, &action); err != nil {
		logger.WithError(err).Error("failed to unmarshal sage recall arguments")
		return "", fmt.Errorf("failed to unmarshal sage recall arguments: %w", err)
	}

	action.Query = strings.TrimSpace(action.Query)
	if action.Query == "" {
		logger.Error("query parameter is required")
		return "", fmt.Errorf("query parameter is required for sage_recall")
	}

	// Apply defaults
	minConfidence := action.MinConfidence
	if minConfidence <= 0 {
		minConfidence = 0.6
	}

	maxResults := 5
	if action.MaxResults != nil && action.MaxResults.Int() > 0 {
		maxResults = action.MaxResults.Int()
		if maxResults > 10 {
			maxResults = 10
		}
	}

	req := sage.RecallRequest{
		Query:         action.Query,
		Domain:        action.Domain,
		MinConfidence: minConfidence,
		MaxResults:    maxResults,
	}

	resp, err := t.sageClient.Recall(ctx, req)
	if err != nil {
		logger.WithError(err).Error("failed to recall from SAGE")
		return "", fmt.Errorf("SAGE recall failed: %w", err)
	}

	return FormatSageRecallResults(resp, action.Query), nil
}

// handleRemember stores knowledge in SAGE persistent memory
func (t *sageSearchTool) handleRemember(ctx context.Context, logger *logrus.Entry, args json.RawMessage) (string, error) {
	var action SageRememberAction
	if err := json.Unmarshal(args, &action); err != nil {
		logger.WithError(err).Error("failed to unmarshal sage remember arguments")
		return "", fmt.Errorf("failed to unmarshal sage remember arguments: %w", err)
	}

	action.Content = strings.TrimSpace(action.Content)
	if action.Content == "" {
		logger.Error("content parameter is required")
		return "", fmt.Errorf("content parameter is required for sage_remember")
	}

	if action.MemoryType == "" {
		logger.Error("memory_type parameter is required")
		return "", fmt.Errorf("memory_type parameter is required for sage_remember")
	}

	if action.Domain == "" {
		logger.Error("domain parameter is required")
		return "", fmt.Errorf("domain parameter is required for sage_remember")
	}

	if action.Confidence <= 0 || action.Confidence > 1.0 {
		action.Confidence = 0.8
	}

	req := sage.RememberRequest{
		Content:    action.Content,
		MemoryType: action.MemoryType,
		Domain:     action.Domain,
		Confidence: action.Confidence,
	}

	resp, err := t.sageClient.Remember(ctx, req)
	if err != nil {
		logger.WithError(err).Error("failed to store in SAGE")
		return "", fmt.Errorf("SAGE remember failed: %w", err)
	}

	return FormatSageRememberResult(resp), nil
}

// FormatSageRecallResults formats SAGE recall results for agent consumption
func FormatSageRecallResults(resp *sage.RecallResponse, query string) string {
	var builder strings.Builder

	builder.WriteString("# SAGE Cross-Session Memory Results\n\n")
	builder.WriteString(fmt.Sprintf("**Query:** %s\n\n", query))

	if len(resp.Memories) == 0 {
		builder.WriteString("No cross-session memories found matching your query.\n")
		return builder.String()
	}

	for i, mem := range resp.Memories {
		builder.WriteString(fmt.Sprintf("## Memory %d (confidence: %.2f, type: %s)\n",
			i+1, mem.Confidence, mem.MemoryType))
		builder.WriteString(fmt.Sprintf("**Domain:** %s\n", mem.Domain))
		if mem.StoredAt != "" {
			builder.WriteString(fmt.Sprintf("**Stored:** %s\n", mem.StoredAt))
		}
		builder.WriteString("\n")
		builder.WriteString(mem.Content)
		builder.WriteString("\n---------------------------\n")
	}

	return builder.String()
}

// FormatSageRememberResult formats the confirmation of a SAGE remember operation
func FormatSageRememberResult(resp *sage.RememberResponse) string {
	var builder strings.Builder

	builder.WriteString("# SAGE Memory Stored\n\n")

	if resp.Success {
		builder.WriteString("Knowledge successfully submitted to SAGE persistent memory.\n")
		if resp.MemoryID != "" {
			builder.WriteString(fmt.Sprintf("**Memory ID:** %s\n", resp.MemoryID))
		}
		if resp.Message != "" {
			builder.WriteString(fmt.Sprintf("**Status:** %s\n", resp.Message))
		}
	} else {
		builder.WriteString("Failed to store knowledge in SAGE persistent memory.\n")
		if resp.Message != "" {
			builder.WriteString(fmt.Sprintf("**Reason:** %s\n", resp.Message))
		}
	}

	return builder.String()
}
