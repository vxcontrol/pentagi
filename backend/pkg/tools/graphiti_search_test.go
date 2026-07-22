package tools

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"pentagi/pkg/graphiti"
)

// stubGraphitiSearcher is a minimal GraphitiSearcher test double: every method
// returns whatever error/response was configured for it, so tests can exercise
// the Handle() error-classification logic without a real Graphiti/Neo4j backend.
type stubGraphitiSearcher struct {
	enabled bool
	err     error
}

func (s *stubGraphitiSearcher) IsEnabled() bool { return s.enabled }

func (s *stubGraphitiSearcher) TemporalWindowSearch(
	ctx context.Context, req graphiti.TemporalSearchRequest,
) (*graphiti.TemporalSearchResponse, error) {
	return nil, s.err
}

func (s *stubGraphitiSearcher) EntityRelationshipsSearch(
	ctx context.Context, req graphiti.EntityRelationshipSearchRequest,
) (*graphiti.EntityRelationshipSearchResponse, error) {
	return nil, s.err
}

func (s *stubGraphitiSearcher) DiverseResultsSearch(
	ctx context.Context, req graphiti.DiverseSearchRequest,
) (*graphiti.DiverseSearchResponse, error) {
	return nil, s.err
}

func (s *stubGraphitiSearcher) EpisodeContextSearch(
	ctx context.Context, req graphiti.EpisodeContextSearchRequest,
) (*graphiti.EpisodeContextSearchResponse, error) {
	return nil, s.err
}

func (s *stubGraphitiSearcher) SuccessfulToolsSearch(
	ctx context.Context, req graphiti.SuccessfulToolsSearchRequest,
) (*graphiti.SuccessfulToolsSearchResponse, error) {
	return nil, s.err
}

func (s *stubGraphitiSearcher) RecentContextSearch(
	ctx context.Context, req graphiti.RecentContextSearchRequest,
) (*graphiti.RecentContextSearchResponse, error) {
	return nil, s.err
}

func (s *stubGraphitiSearcher) EntityByLabelSearch(
	ctx context.Context, req graphiti.EntityByLabelSearchRequest,
) (*graphiti.EntityByLabelSearchResponse, error) {
	return nil, s.err
}

// fakeNetError mimics the *url.Error shape produced by http.Client.Do on a
// transport-level failure (timeout, TLS handshake timeout, connection refused).
func fakeNetError() error {
	return fmt.Errorf(
		"recent context search failed: failed to perform request: %w",
		&url.Error{
			Op:  "Post",
			URL: "http://graphiti-neo4j/search/recent-context",
			Err: context.DeadlineExceeded,
		},
	)
}

func TestGraphitiSearchTool_Handle_NetworkFailure_DegradesGracefully(t *testing.T) {
	tool := NewGraphitiSearchTool(1, nil, nil, &stubGraphitiSearcher{enabled: true, err: fakeNetError()})

	args := []byte(`{"search_type":"recent_context","query":"test query","message":"m"}`)
	result, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err != nil {
		t.Fatalf("expected graceful degradation (nil error) on transport failure, got error: %v", err)
	}
	if !strings.Contains(result, "temporarily unavailable") {
		t.Fatalf("expected soft-fail message about temporary unavailability, got: %q", result)
	}
}

func TestGraphitiSearchTool_Handle_ValidationError_StaysHard(t *testing.T) {
	tool := NewGraphitiSearchTool(1, nil, nil, &stubGraphitiSearcher{enabled: true})

	// Invalid search_type never reaches the graphiti client - it is rejected by
	// Handle() itself, so this must remain a hard failure regardless of the
	// network-error leniency added for transport failures.
	args := []byte(`{"search_type":"not_a_real_type","query":"test query","message":"m"}`)
	_, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err == nil || !strings.Contains(err.Error(), "unknown search_type") {
		t.Fatalf("expected hard 'unknown search_type' error, got: %v", err)
	}
}

func TestGraphitiSearchTool_Handle_InvalidRecencyWindow_StaysHard(t *testing.T) {
	tool := NewGraphitiSearchTool(1, nil, nil, &stubGraphitiSearcher{enabled: true})

	// Argument-validation errors (not network errors) must still be treated as
	// hard failures so the tool-call arg-fixer can actually help here.
	args := []byte(`{"search_type":"recent_context","query":"test query","message":"m","recency_window":"not-a-window"}`)
	_, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err == nil || !strings.Contains(err.Error(), "invalid recency_window") {
		t.Fatalf("expected hard 'invalid recency_window' error, got: %v", err)
	}
}
