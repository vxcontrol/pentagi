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
	if s.err != nil {
		return nil, s.err
	}
	return &graphiti.EntityRelationshipSearchResponse{}, nil
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
	if s.err != nil {
		return nil, s.err
	}
	return &graphiti.EntityByLabelSearchResponse{}, nil
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

// fakeStatusError mimics the exact error shape graphiti-go-client's `do`
// produces for a non-2xx HTTP response: a plain fmt.Errorf embedding the raw
// response body, further wrapped by the handle*Search method's own %w wrap.
func fakeStatusError(statusCode int, body string) error {
	return fmt.Errorf(
		"recent context search failed: API request failed with status %d: %s",
		statusCode, body,
	)
}

func TestGraphitiSearchTool_Handle_ServerError5xx_DegradesGracefullyWithBody(t *testing.T) {
	tool := NewGraphitiSearchTool(1, nil, nil, &stubGraphitiSearcher{
		enabled: true,
		err:     fakeStatusError(502, "<html><body>502 Bad Gateway</body></html>"),
	})

	args := []byte(`{"search_type":"recent_context","query":"test query","message":"m"}`)
	result, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err != nil {
		t.Fatalf("expected graceful degradation (nil error) on 5xx, got error: %v", err)
	}
	if !strings.Contains(result, "HTTP 502") {
		t.Fatalf("expected result to mention the status code, got: %q", result)
	}
	if !strings.Contains(result, "502 Bad Gateway") {
		t.Fatalf("expected result to include the response body, got: %q", result)
	}
}

func TestGraphitiSearchTool_Handle_ServerError5xx_BodyTruncatedAt512Bytes(t *testing.T) {
	hugeBody := strings.Repeat("x", 2000)
	tool := NewGraphitiSearchTool(1, nil, nil, &stubGraphitiSearcher{
		enabled: true,
		err:     fakeStatusError(500, hugeBody),
	})

	args := []byte(`{"search_type":"recent_context","query":"test query","message":"m"}`)
	result, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err != nil {
		t.Fatalf("expected graceful degradation (nil error) on 5xx, got error: %v", err)
	}
	if strings.Count(result, "x") >= 2000 {
		t.Fatalf("expected the 2000-byte body to be truncated to the 512-byte cap, got a result of length %d", len(result))
	}
	if !strings.Contains(result, "truncated") {
		t.Fatalf("expected result to indicate truncation, got: %q", result)
	}
}

func TestGraphitiSearchTool_Handle_ClientError4xx_StaysHardWithTruncatedBody(t *testing.T) {
	hugeBody := strings.Repeat("y", 2000)
	tool := NewGraphitiSearchTool(1, nil, nil, &stubGraphitiSearcher{
		enabled: true,
		err:     fakeStatusError(400, hugeBody),
	})

	args := []byte(`{"search_type":"recent_context","query":"test query","message":"m"}`)
	_, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err == nil {
		t.Fatal("expected a hard failure for a 4xx status, got nil error")
	}
	if !strings.Contains(err.Error(), "status 400") {
		t.Fatalf("expected error to mention the status code, got: %v", err)
	}
	if strings.Count(err.Error(), "y") >= 2000 {
		t.Fatalf("expected the 2000-byte body to be truncated to the 512-byte cap even on the hard-fail path, got error of length %d", len(err.Error()))
	}
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

func TestGraphitiSearchTool_Handle_EntityByLabel_MissingNodeLabels_GivesActionableError(t *testing.T) {
	tool := NewGraphitiSearchTool(1, nil, nil, &stubGraphitiSearcher{enabled: true})

	args := []byte(`{"search_type":"entity_by_label","query":"test query","message":"m"}`)
	_, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err == nil {
		t.Fatal("expected a hard failure when node_labels is missing for entity_by_label, got nil error")
	}
	if !strings.Contains(err.Error(), "node_labels is required") {
		t.Fatalf("expected error to state node_labels is required, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Vulnerability") {
		t.Fatalf("expected error to include a real taxonomy example value to guide the LLM, got: %v", err)
	}
}

func TestGraphitiSearchTool_Handle_EntityByLabel_WithNodeLabels_Succeeds(t *testing.T) {
	tool := NewGraphitiSearchTool(1, nil, nil, &stubGraphitiSearcher{enabled: true})

	args := []byte(`{"search_type":"entity_by_label","query":"test query","node_labels":["Vulnerability"],"message":"m"}`)
	_, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err != nil {
		t.Fatalf("expected no error when node_labels is present, got: %v", err)
	}
}

func TestGraphitiSearchTool_Handle_EntityRelationships_MissingCenterNodeUUID_StaysHard(t *testing.T) {
	tool := NewGraphitiSearchTool(1, nil, nil, &stubGraphitiSearcher{enabled: true})

	args := []byte(`{"search_type":"entity_relationships","query":"test query","message":"m"}`)
	_, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err == nil || !strings.Contains(err.Error(), "center_node_uuid is required") {
		t.Fatalf("expected hard 'center_node_uuid is required' error, got: %v", err)
	}
}

func TestGraphitiSearchTool_Handle_EntityRelationships_MalformedCenterNodeUUID_GivesActionableError(t *testing.T) {
	tool := NewGraphitiSearchTool(1, nil, nil, &stubGraphitiSearcher{enabled: true})

	// Simulates a diagnostic string, truncated ID, or otherwise non-UUID value
	// ending up in center_node_uuid (e.g. a hallucinated or mangled value) -
	// this must be rejected before ever reaching the graph backend.
	args := []byte(`{"search_type":"entity_relationships","query":"test query","center_node_uuid":"not-a-real-uuid","message":"m"}`)
	_, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err == nil {
		t.Fatal("expected a hard failure for a malformed center_node_uuid, got nil error")
	}
	if !strings.Contains(err.Error(), "must be a valid UUID") {
		t.Fatalf("expected error to explain the UUID requirement, got: %v", err)
	}
	if !strings.Contains(err.Error(), "not-a-real-uuid") {
		t.Fatalf("expected error to echo back the offending value, got: %v", err)
	}
}

func TestGraphitiSearchTool_Handle_EntityRelationships_ValidCenterNodeUUID_Succeeds(t *testing.T) {
	tool := NewGraphitiSearchTool(1, nil, nil, &stubGraphitiSearcher{enabled: true})

	args := []byte(`{"search_type":"entity_relationships","query":"test query","center_node_uuid":"f7b95dfc-ee58-4a8b-8d85-582cf117b4df","message":"m"}`)
	_, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err != nil {
		t.Fatalf("expected no error for a well-formed center_node_uuid, got: %v", err)
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
