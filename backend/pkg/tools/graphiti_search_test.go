package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"
	"testing"

	"pentagi/pkg/graphiti"
)

const testGroupID = "test-group"

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
	if s.err != nil {
		return nil, s.err
	}
	return &graphiti.TemporalSearchResponse{}, nil
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
	tool := NewGraphitiSearchTool(1, nil, nil, testGroupID, &stubGraphitiSearcher{
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
	tool := NewGraphitiSearchTool(1, nil, nil, testGroupID, &stubGraphitiSearcher{
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
	tool := NewGraphitiSearchTool(1, nil, nil, testGroupID, &stubGraphitiSearcher{
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
	tool := NewGraphitiSearchTool(1, nil, nil, testGroupID, &stubGraphitiSearcher{enabled: true, err: fakeNetError()})

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
	tool := NewGraphitiSearchTool(1, nil, nil, testGroupID, &stubGraphitiSearcher{enabled: true})

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
	tool := NewGraphitiSearchTool(1, nil, nil, testGroupID, &stubGraphitiSearcher{enabled: true})

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
	tool := NewGraphitiSearchTool(1, nil, nil, testGroupID, &stubGraphitiSearcher{enabled: true})

	args := []byte(`{"search_type":"entity_by_label","query":"test query","node_labels":["Vulnerability"],"message":"m"}`)
	_, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err != nil {
		t.Fatalf("expected no error when node_labels is present, got: %v", err)
	}
}

func TestGraphitiSearchTool_Handle_EntityRelationships_MissingCenterNodeUUID_StaysHard(t *testing.T) {
	tool := NewGraphitiSearchTool(1, nil, nil, testGroupID, &stubGraphitiSearcher{enabled: true})

	args := []byte(`{"search_type":"entity_relationships","query":"test query","message":"m"}`)
	_, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err == nil || !strings.Contains(err.Error(), "center_node_uuid is required") {
		t.Fatalf("expected hard 'center_node_uuid is required' error, got: %v", err)
	}
}

func TestGraphitiSearchTool_Handle_EntityRelationships_MalformedCenterNodeUUID_GivesActionableError(t *testing.T) {
	tool := NewGraphitiSearchTool(1, nil, nil, testGroupID, &stubGraphitiSearcher{enabled: true})

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
	tool := NewGraphitiSearchTool(1, nil, nil, testGroupID, &stubGraphitiSearcher{enabled: true})

	args := []byte(`{"search_type":"entity_relationships","query":"test query","center_node_uuid":"f7b95dfc-ee58-4a8b-8d85-582cf117b4df","message":"m"}`)
	_, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err != nil {
		t.Fatalf("expected no error for a well-formed center_node_uuid, got: %v", err)
	}
}

func TestParseGraphitiTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"RFC3339 with Z", "2026-07-25T11:53:34Z", false},
		{"RFC3339 with offset", "2026-07-25T11:53:34+03:00", false},
		{"missing timezone designator (production near-miss)", "2026-07-24T11:53:34", false},
		{"space-separated, no timezone", "2026-07-24 11:53:34", false},
		{"garbage", "not-a-date", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseGraphitiTime(tt.input)
			if tt.wantErr && err == nil {
				t.Fatalf("expected an error for input %q, got nil", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error for input %q, got: %v", tt.input, err)
			}
		})
	}
}

func TestGraphitiSearchTool_Handle_TemporalWindow_MissingTimezone_Succeeds(t *testing.T) {
	tool := NewGraphitiSearchTool(1, nil, nil, testGroupID, &stubGraphitiSearcher{enabled: true})

	// Reproduces the exact production near-miss: an LLM omitted the trailing
	// 'Z'/offset on an otherwise well-formed timestamp.
	args := []byte(`{"search_type":"temporal_window","query":"test query","time_start":"2026-07-24T11:53:34","time_end":"2026-07-25T11:53:34","message":"m"}`)
	_, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err != nil {
		t.Fatalf("expected no error for a timestamp missing its timezone designator, got: %v", err)
	}
}

func TestGraphitiSearchTool_Handle_TemporalWindow_GarbageTime_StaysHard(t *testing.T) {
	tool := NewGraphitiSearchTool(1, nil, nil, testGroupID, &stubGraphitiSearcher{enabled: true})

	args := []byte(`{"search_type":"temporal_window","query":"test query","time_start":"not-a-date","time_end":"2026-07-25T11:53:34Z","message":"m"}`)
	_, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err == nil || !strings.Contains(err.Error(), "invalid time_start format") {
		t.Fatalf("expected hard 'invalid time_start format' error, got: %v", err)
	}
}

func TestGraphitiSearchTool_Handle_InvalidRecencyWindow_StaysHard(t *testing.T) {
	tool := NewGraphitiSearchTool(1, nil, nil, testGroupID, &stubGraphitiSearcher{enabled: true})

	// Argument-validation errors (not network errors) must still be treated as
	// hard failures so the tool-call arg-fixer can actually help here.
	args := []byte(`{"search_type":"recent_context","query":"test query","message":"m","recency_window":"not-a-window"}`)
	_, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)

	if err == nil || !strings.Contains(err.Error(), "invalid recency_window") {
		t.Fatalf("expected hard 'invalid recency_window' error, got: %v", err)
	}
}

// These three tests are regression guards for a real production bug: the
// entity/community listings for recent_context, diverse_results, and
// episode_context silently omitted the "UUID:" field that
// FormatGraphitiTemporalResults, FormatGraphitiEntityRelationshipResults, and
// FormatGraphitiEntityByLabelResults already included. Since entity_relationships
// requires a real center_node_uuid copied from an EARLIER result, agents that
// started their research with recent_context (the documented default) never
// saw a UUID to copy and fabricated one instead (a flow ID, a hostname, a
// page title) - which is exactly what graphiti_search's new UUID-format
// validation now (correctly) rejects. Fixing only the validation without also
// restoring the missing UUID field would leave every recent_context-first
// workflow permanently unable to reach entity_relationships.

func TestFormatGraphitiRecentContextResults_IncludesNodeUUID(t *testing.T) {
	resp := &graphiti.RecentContextSearchResponse{
		Nodes: []graphiti.NodeResult{
			{UUID: "f7b95dfc-ee58-4a8b-8d85-582cf117b4df", Name: "500", Labels: []string{"Entity", "Port"}, Summary: "Host has port 500"},
		},
		NodeScores: []float64{0.5},
	}

	result := FormatGraphitiRecentContextResults(resp, "test query")

	if !strings.Contains(result, "UUID: f7b95dfc-ee58-4a8b-8d85-582cf117b4df") {
		t.Fatalf("expected recent_context entity listing to include the node UUID, got:\n%s", result)
	}
}

func TestFormatGraphitiDiverseResults_IncludesCommunityUUID(t *testing.T) {
	resp := &graphiti.DiverseSearchResponse{
		Communities: []graphiti.CommunityResult{
			{UUID: "a1b2c3d4-e5f6-4789-a012-3456789abcde", Name: "Trading Platform Cluster", Summary: "Cluster of related findings"},
		},
		CommunityMMRScores: []float64{0.9},
	}

	result := FormatGraphitiDiverseResults(resp, "test query")

	if !strings.Contains(result, "UUID: a1b2c3d4-e5f6-4789-a012-3456789abcde") {
		t.Fatalf("expected diverse_results community listing to include the community UUID, got:\n%s", result)
	}
}

func TestFormatGraphitiEpisodeContextResults_IncludesMentionedNodeUUID(t *testing.T) {
	resp := &graphiti.EpisodeContextSearchResponse{
		MentionedNodes: []graphiti.NodeResult{
			{UUID: "11111111-2222-4333-8444-555555555555", Name: "NoSQL Injection", Summary: "Confirmed vulnerability"},
		},
		MentionedNodeScores: []float64{0.8},
	}

	result := FormatGraphitiEpisodeContextResults(resp, "test query")

	if !strings.Contains(result, "UUID: 11111111-2222-4333-8444-555555555555") {
		t.Fatalf("expected episode_context mentioned-entities listing to include the node UUID, got:\n%s", result)
	}
}

// TestGraphitiSearchAction_JSONSchema_IncludesAllFields serializes the exact
// JSON schema the LLM is shown for the graphiti_search tool (the same
// reflector.Reflect(&GraphitiSearchAction{}) call registry.go uses to build
// the live tool definition) and verifies every struct field survives into
// the "properties" object under its declared json name, and that "required"
// matches the struct tags exactly.
//
// This exists to answer a concrete production question: agents were sending
// hallucinated values (a flow ID, a hostname, a page title, even a stray
// markdown line) for center_node_uuid instead of a real Graphiti UUID. This
// test rules out "the LLM never actually saw the field / its guidance" as a
// cause — if it ever regresses (a field silently dropped, unexported, or
// renamed without updating the json tag), this test fails at build/test time
// instead of surfacing only as a confusing runtime tool-call error.
func TestGraphitiSearchAction_JSONSchema_IncludesAllFields(t *testing.T) {
	schema := reflector.Reflect(&GraphitiSearchAction{})

	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("failed to unmarshal schema into a generic map: %v", err)
	}

	properties, ok := doc["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema has no 'properties' object, got: %s", raw)
	}

	// Every field of GraphitiSearchAction, by its json tag name.
	wantFields := []string{
		"search_type", "query", "max_results", "time_start", "time_end",
		"center_node_uuid", "max_depth", "node_labels", "edge_types",
		"diversity_level", "min_mentions", "recency_window", "message",
	}
	for _, field := range wantFields {
		if _, ok := properties[field]; !ok {
			t.Errorf("expected field %q in the JSON schema 'properties' but it is missing (full schema below)\n%s", field, raw)
		}
	}
	if len(properties) != len(wantFields) {
		t.Errorf("expected exactly %d properties, got %d: %v", len(wantFields), len(properties), slices.Collect(maps.Keys(properties)))
	}

	requiredRaw, ok := doc["required"].([]any)
	if !ok {
		t.Fatalf("schema has no 'required' array, got: %s", raw)
	}
	var required []string
	for _, r := range requiredRaw {
		s, ok := r.(string)
		if !ok {
			t.Fatalf("required entry is not a string: %v", r)
		}
		required = append(required, s)
	}
	wantRequired := []string{"search_type", "query", "message"}
	if len(required) != len(wantRequired) {
		t.Fatalf("expected required=%v, got %v", wantRequired, required)
	}
	for _, field := range wantRequired {
		if !slices.Contains(required, field) {
			t.Errorf("expected %q to be in the required list, got %v", field, required)
		}
	}

	// Spot-check the exact guidance an LLM would read for the two fields
	// implicated in the production failure, so a future edit that weakens
	// or removes this wording is caught here rather than in a live flow.
	centerNodeUUID, ok := properties["center_node_uuid"].(map[string]any)
	if !ok {
		t.Fatalf("center_node_uuid property is not an object: %v", properties["center_node_uuid"])
	}
	centerDesc, _ := centerNodeUUID["description"].(string)
	if !strings.Contains(centerDesc, "NEVER invent") {
		t.Errorf("expected center_node_uuid description to warn against inventing a value, got: %q", centerDesc)
	}
	if !strings.Contains(centerDesc, "UUID:") {
		t.Errorf("expected center_node_uuid description to reference the 'UUID:' field agents must copy from, got: %q", centerDesc)
	}

	nodeLabels, ok := properties["node_labels"].(map[string]any)
	if !ok {
		t.Fatalf("node_labels property is not an object: %v", properties["node_labels"])
	}
	nodeLabelsDesc, _ := nodeLabels["description"].(string)
	if !strings.Contains(nodeLabelsDesc, "PascalCase") {
		t.Errorf("expected node_labels description to mention PascalCase casing, got: %q", nodeLabelsDesc)
	}
}
