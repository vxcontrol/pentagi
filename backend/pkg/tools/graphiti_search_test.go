package tools

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/graphiti"
)

const graphitiSearchGroupID = "test-group"

// graphitiSearchStub stands in for the Graphiti backend and, like its HTTP request, fails on a done context.
type graphitiSearchStub struct {
	enabled bool
	err     error

	calls []string
	// request is the last request received, with its Observation cleared so it compares by value.
	request any
	// observed reports whether that request carried an Observation to nest server-side spans under.
	observed bool
}

func graphitiSearchAnswer[R any](
	ctx context.Context, s *graphitiSearchStub, method string, request any, observation *graphiti.Observation,
) (*R, error) {
	s.calls = append(s.calls, method)
	s.request = request
	s.observed = observation != nil
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.err != nil {
		return nil, s.err
	}
	return new(R), nil
}

func (s *graphitiSearchStub) IsEnabled() bool { return s.enabled }

func (s *graphitiSearchStub) TemporalWindowSearch(
	ctx context.Context, req graphiti.TemporalSearchRequest,
) (*graphiti.TemporalSearchResponse, error) {
	observation := req.Observation
	req.Observation = nil
	return graphitiSearchAnswer[graphiti.TemporalSearchResponse](ctx, s, "TemporalWindowSearch", req, observation)
}

func (s *graphitiSearchStub) EntityRelationshipsSearch(
	ctx context.Context, req graphiti.EntityRelationshipSearchRequest,
) (*graphiti.EntityRelationshipSearchResponse, error) {
	observation := req.Observation
	req.Observation = nil
	return graphitiSearchAnswer[graphiti.EntityRelationshipSearchResponse](ctx, s, "EntityRelationshipsSearch", req, observation)
}

func (s *graphitiSearchStub) DiverseResultsSearch(
	ctx context.Context, req graphiti.DiverseSearchRequest,
) (*graphiti.DiverseSearchResponse, error) {
	observation := req.Observation
	req.Observation = nil
	return graphitiSearchAnswer[graphiti.DiverseSearchResponse](ctx, s, "DiverseResultsSearch", req, observation)
}

func (s *graphitiSearchStub) EpisodeContextSearch(
	ctx context.Context, req graphiti.EpisodeContextSearchRequest,
) (*graphiti.EpisodeContextSearchResponse, error) {
	observation := req.Observation
	req.Observation = nil
	return graphitiSearchAnswer[graphiti.EpisodeContextSearchResponse](ctx, s, "EpisodeContextSearch", req, observation)
}

func (s *graphitiSearchStub) SuccessfulToolsSearch(
	ctx context.Context, req graphiti.SuccessfulToolsSearchRequest,
) (*graphiti.SuccessfulToolsSearchResponse, error) {
	observation := req.Observation
	req.Observation = nil
	return graphitiSearchAnswer[graphiti.SuccessfulToolsSearchResponse](ctx, s, "SuccessfulToolsSearch", req, observation)
}

func (s *graphitiSearchStub) RecentContextSearch(
	ctx context.Context, req graphiti.RecentContextSearchRequest,
) (*graphiti.RecentContextSearchResponse, error) {
	observation := req.Observation
	req.Observation = nil
	return graphitiSearchAnswer[graphiti.RecentContextSearchResponse](ctx, s, "RecentContextSearch", req, observation)
}

func (s *graphitiSearchStub) EntityByLabelSearch(
	ctx context.Context, req graphiti.EntityByLabelSearchRequest,
) (*graphiti.EntityByLabelSearchResponse, error) {
	observation := req.Observation
	req.Observation = nil
	return graphitiSearchAnswer[graphiti.EntityByLabelSearchResponse](ctx, s, "EntityByLabelSearch", req, observation)
}

func TestGraphitiSearch_Handle_AnswersWithoutTheBackendWhenDisabled(t *testing.T) {
	disabled := &graphitiSearchStub{enabled: false}
	tests := []struct {
		name   string
		client GraphitiSearcher
	}{
		{"no client configured", nil},
		{"a client that is switched off", disabled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewGraphitiSearchTool(1, nil, nil, graphitiSearchGroupID, tt.client)

			args := []byte(`{"search_type":"recent_context","query":"open ports","message":"m"}`)
			result, err := tool.Handle(t.Context(), GraphitiSearchToolName, args)
			if err != nil {
				t.Fatalf("a disabled knowledge graph is not the model's mistake, yet Handle failed: %v", err)
			}
			want := "Graphiti knowledge graph is not enabled. No historical context or memory data is available for this search."
			if result != want {
				t.Fatalf("result = %q, want %q", result, want)
			}
		})
	}
	if len(disabled.calls) != 0 {
		t.Fatalf("a switched-off client was still called: %v", disabled.calls)
	}
}

func TestGraphitiSearch_Handle_RejectsBadArgumentsBeforeTheBackend(t *testing.T) {
	tests := []struct {
		name string
		args string
		want []string
	}{
		{
			name: "arguments that are not json",
			args: `{"search_type":`,
			want: []string{"failed to unmarshal search arguments"},
		},
		{
			name: "a query of blanks",
			args: `{"search_type":"recent_context","query":"   ","message":"m"}`,
			want: []string{"query parameter is required"},
		},
		{
			name: "no search type",
			args: `{"query":"open ports","message":"m"}`,
			want: []string{"search_type parameter is required"},
		},
		{
			name: "an unknown search type",
			args: `{"search_type":"not_a_real_type","query":"open ports","message":"m"}`,
			want: []string{"unknown search_type: not_a_real_type"},
		},
		{
			name: "temporal_window without an end",
			args: `{"search_type":"temporal_window","query":"open ports","time_start":"2026-07-24T11:53:34Z","message":"m"}`,
			want: []string{"time_start and time_end are required for temporal_window search"},
		},
		{
			name: "temporal_window with an unparseable start",
			args: `{"search_type":"temporal_window","query":"open ports","time_start":"not-a-date","time_end":"2026-07-25T11:53:34Z","message":"m"}`,
			want: []string{"invalid time_start format"},
		},
		{
			name: "temporal_window with an unparseable end",
			args: `{"search_type":"temporal_window","query":"open ports","time_start":"2026-07-24T11:53:34Z","time_end":"yesterday","message":"m"}`,
			want: []string{"invalid time_end format"},
		},
		{
			name: "temporal_window ending before it starts",
			args: `{"search_type":"temporal_window","query":"open ports","time_start":"2026-07-25T11:53:34Z","time_end":"2026-07-24T11:53:34Z","message":"m"}`,
			want: []string{"time_end must be after time_start"},
		},
		{
			name: "entity_relationships without a center node",
			args: `{"search_type":"entity_relationships","query":"open ports","message":"m"}`,
			want: []string{"center_node_uuid is required"},
		},
		{
			name: "entity_relationships with a center node that is not a uuid",
			args: `{"search_type":"entity_relationships","query":"open ports","center_node_uuid":"not-a-real-uuid","message":"m"}`,
			want: []string{"must be a valid UUID", `"not-a-real-uuid"`},
		},
		{
			name: "diverse_results with an unknown diversity level",
			args: `{"search_type":"diverse_results","query":"open ports","diversity_level":"extreme","message":"m"}`,
			want: []string{"invalid diversity_level: extreme"},
		},
		{
			name: "recent_context with an unknown recency window",
			args: `{"search_type":"recent_context","query":"open ports","recency_window":"not-a-window","message":"m"}`,
			want: []string{"invalid recency_window: not-a-window"},
		},
		{
			name: "entity_by_label without node labels",
			args: `{"search_type":"entity_by_label","query":"open ports","message":"m"}`,
			want: []string{"node_labels is required", "Vulnerability"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := &graphitiSearchStub{enabled: true}
			tool := NewGraphitiSearchTool(1, nil, nil, graphitiSearchGroupID, backend)

			result, err := tool.Handle(t.Context(), GraphitiSearchToolName, []byte(tt.args))
			if err == nil {
				t.Fatalf("a bad argument must be a hard failure, got result %q", result)
			}
			for _, want := range tt.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not contain %q", err, want)
				}
			}
			if len(backend.calls) != 0 {
				t.Errorf("a refused search still reached the backend: %v", backend.calls)
			}
		})
	}
}

// Every request must carry the group id; the defaults and the depth cap are the ones the schema promises.
func TestGraphitiSearch_Handle_SendsEachSearchTypeToItsBackendCall(t *testing.T) {
	groupID := graphitiSearchGroupID
	labels := func(values ...string) *[]string { return &values }
	const center = "f7b95dfc-ee58-4a8b-8d85-582cf117b4df"

	tests := []struct {
		name        string
		args        string
		wantCall    string
		wantRequest any
		wantResult  []string
	}{
		{
			name:     "temporal_window with zone-less timestamps",
			args:     `{"search_type":"temporal_window","query":"open ports","time_start":"2026-07-24T11:53:34","time_end":"2026-07-25T11:53:34","message":"m"}`,
			wantCall: "TemporalWindowSearch",
			wantRequest: graphiti.TemporalSearchRequest{
				Query: "open ports", GroupID: &groupID, MaxResults: 15,
				TimeStart: time.Date(2026, 7, 24, 11, 53, 34, 0, time.UTC),
				TimeEnd:   time.Date(2026, 7, 25, 11, 53, 34, 0, time.UTC),
			},
			wantResult: []string{"# Temporal Search Results", "**Query:** open ports", "No results found in the specified time window."},
		},
		{
			name:     "entity_relationships with only a center node",
			args:     `{"search_type":"entity_relationships","query":"open ports","center_node_uuid":"` + center + `","message":"m"}`,
			wantCall: "EntityRelationshipsSearch",
			wantRequest: graphiti.EntityRelationshipSearchRequest{
				Query: "open ports", GroupID: &groupID, CenterNodeUUID: center, MaxDepth: 2, MaxResults: 20,
			},
			wantResult: []string{"# Entity Relationship Search Results", "No relationships found matching criteria."},
		},
		{
			name:     "entity_relationships deeper than the cap, with filters",
			args:     `{"search_type":"entity_relationships","query":"open ports","center_node_uuid":"` + center + `","max_depth":7,"max_results":5,"node_labels":["Host"],"edge_types":["HAS_PORT"],"message":"m"}`,
			wantCall: "EntityRelationshipsSearch",
			wantRequest: graphiti.EntityRelationshipSearchRequest{
				Query: "open ports", GroupID: &groupID, CenterNodeUUID: center, MaxDepth: 3, MaxResults: 5,
				NodeLabels: labels("Host"), EdgeTypes: labels("HAS_PORT"),
			},
			wantResult: []string{"# Entity Relationship Search Results"},
		},
		{
			name:     "diverse_results with the default diversity",
			args:     `{"search_type":"diverse_results","query":"open ports","message":"m"}`,
			wantCall: "DiverseResultsSearch",
			wantRequest: graphiti.DiverseSearchRequest{
				Query: "open ports", GroupID: &groupID, DiversityLevel: "medium", MaxResults: 10,
			},
			wantResult: []string{"# Diverse Search Results", "**Query:** open ports"},
		},
		{
			name:     "diverse_results asking for high diversity",
			args:     `{"search_type":"diverse_results","query":"open ports","diversity_level":"high","message":"m"}`,
			wantCall: "DiverseResultsSearch",
			wantRequest: graphiti.DiverseSearchRequest{
				Query: "open ports", GroupID: &groupID, DiversityLevel: "high", MaxResults: 10,
			},
			wantResult: []string{"# Diverse Search Results"},
		},
		{
			name:        "episode_context with the default limit",
			args:        `{"search_type":"episode_context","query":"open ports","message":"m"}`,
			wantCall:    "EpisodeContextSearch",
			wantRequest: graphiti.EpisodeContextSearchRequest{Query: "open ports", GroupID: &groupID, MaxResults: 10},
			wantResult:  []string{"# Episode Context Results", "No episode context found."},
		},
		{
			name:     "successful_tools with the default thresholds",
			args:     `{"search_type":"successful_tools","query":"open ports","message":"m"}`,
			wantCall: "SuccessfulToolsSearch",
			wantRequest: graphiti.SuccessfulToolsSearchRequest{
				Query: "open ports", GroupID: &groupID, MinMentions: 2, MaxResults: 15,
			},
			wantResult: []string{"# Successful Tools & Techniques", "No successful tool executions found matching criteria."},
		},
		{
			name:     "successful_tools with its own thresholds",
			args:     `{"search_type":"successful_tools","query":"open ports","min_mentions":4,"max_results":3,"message":"m"}`,
			wantCall: "SuccessfulToolsSearch",
			wantRequest: graphiti.SuccessfulToolsSearchRequest{
				Query: "open ports", GroupID: &groupID, MinMentions: 4, MaxResults: 3,
			},
			wantResult: []string{"# Successful Tools & Techniques"},
		},
		{
			name:     "recent_context with a padded query",
			args:     `{"search_type":"recent_context","query":"  open ports \n","message":"m"}`,
			wantCall: "RecentContextSearch",
			wantRequest: graphiti.RecentContextSearchRequest{
				Query: "open ports", GroupID: &groupID, RecencyWindow: "24h", MaxResults: 10,
			},
			wantResult: []string{"# Recent Context", "**Query:** open ports\n", "No recent context found in the specified window."},
		},
		{
			name:     "recent_context over a week",
			args:     `{"search_type":"recent_context","query":"open ports","recency_window":"7d","message":"m"}`,
			wantCall: "RecentContextSearch",
			wantRequest: graphiti.RecentContextSearchRequest{
				Query: "open ports", GroupID: &groupID, RecencyWindow: "7d", MaxResults: 10,
			},
			wantResult: []string{"# Recent Context"},
		},
		{
			name:     "entity_by_label with one label and the default limit",
			args:     `{"search_type":"entity_by_label","query":"open ports","node_labels":["Vulnerability"],"message":"m"}`,
			wantCall: "EntityByLabelSearch",
			wantRequest: graphiti.EntityByLabelSearchRequest{
				Query: "open ports", GroupID: &groupID, NodeLabels: []string{"Vulnerability"}, MaxResults: 25,
			},
			wantResult: []string{"# Entity Inventory Search", "No entities found matching the specified labels/query."},
		},
		{
			name:     "entity_by_label with edge filters",
			args:     `{"search_type":"entity_by_label","query":"open ports","node_labels":["Host"],"edge_types":["HAS_VULNERABILITY"],"max_results":50,"message":"m"}`,
			wantCall: "EntityByLabelSearch",
			wantRequest: graphiti.EntityByLabelSearchRequest{
				Query: "open ports", GroupID: &groupID, NodeLabels: []string{"Host"},
				EdgeTypes: labels("HAS_VULNERABILITY"), MaxResults: 50,
			},
			wantResult: []string{"# Entity Inventory Search"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := &graphitiSearchStub{enabled: true}
			tool := NewGraphitiSearchTool(1, nil, nil, graphitiSearchGroupID, backend)

			result, err := tool.Handle(t.Context(), GraphitiSearchToolName, []byte(tt.args))
			if err != nil {
				t.Fatalf("Handle failed: %v", err)
			}
			if !slices.Equal(backend.calls, []string{tt.wantCall}) {
				t.Fatalf("backend calls = %v, want exactly [%s]", backend.calls, tt.wantCall)
			}
			if !reflect.DeepEqual(backend.request, tt.wantRequest) {
				t.Errorf("request sent:\n%#v\nwant:\n%#v", backend.request, tt.wantRequest)
			}
			if !backend.observed {
				t.Error("the request carried no observation to nest the backend's spans under")
			}
			for _, want := range tt.wantResult {
				if !strings.Contains(result, want) {
					t.Errorf("result does not contain %q:\n%s", want, result)
				}
			}
		})
	}
}

// graphitiSearchNetError is what graphiti-go-client's `do` returns when http.Client.Do fails below HTTP.
func graphitiSearchNetError(cause error) error {
	return fmt.Errorf("failed to perform request: %w", &url.Error{
		Op:  "Post",
		URL: "http://graphiti-neo4j/search/recent-context",
		Err: cause,
	})
}

// graphitiSearchStatusError is what graphiti-go-client's `do` returns for a non-2xx response, whole body included.
func graphitiSearchStatusError(statusCode int, body string) error {
	return fmt.Errorf("API request failed with status %d: %s", statusCode, body)
}

// Every search type wraps the backend's error itself, so each gets its own transport-failure row.
func TestGraphitiSearch_Handle_ClassifiesBackendFailures(t *testing.T) {
	const recentContext = `{"search_type":"recent_context","query":"open ports","message":"m"}`
	cut := func(c string) string {
		return strings.Repeat(c, 512) + "... [truncated full size is 2000 bytes]"
	}
	unavailable := func(search string) string {
		return "Graphiti knowledge graph is temporarily unavailable (" + search + " search failed: failed to perform request: " +
			`Post "http://graphiti-neo4j/search/recent-context": context deadline exceeded); continuing without historical context.`
	}

	tests := []struct {
		name     string
		args     string
		err      error
		wantHard bool
		want     string
	}{
		{
			name: "a transport failure during temporal_window",
			args: `{"search_type":"temporal_window","query":"open ports","time_start":"2026-07-24T11:53:34Z","time_end":"2026-07-25T11:53:34Z","message":"m"}`,
			err:  graphitiSearchNetError(context.DeadlineExceeded),
			want: unavailable("temporal window"),
		},
		{
			name: "a transport failure during entity_relationships",
			args: `{"search_type":"entity_relationships","query":"open ports","center_node_uuid":"f7b95dfc-ee58-4a8b-8d85-582cf117b4df","message":"m"}`,
			err:  graphitiSearchNetError(context.DeadlineExceeded),
			want: unavailable("entity relationships"),
		},
		{
			name: "a transport failure during diverse_results",
			args: `{"search_type":"diverse_results","query":"open ports","message":"m"}`,
			err:  graphitiSearchNetError(context.DeadlineExceeded),
			want: unavailable("diverse results"),
		},
		{
			name: "a transport failure during episode_context",
			args: `{"search_type":"episode_context","query":"open ports","message":"m"}`,
			err:  graphitiSearchNetError(context.DeadlineExceeded),
			want: unavailable("episode context"),
		},
		{
			name: "a transport failure during successful_tools",
			args: `{"search_type":"successful_tools","query":"open ports","message":"m"}`,
			err:  graphitiSearchNetError(context.DeadlineExceeded),
			want: unavailable("successful tools"),
		},
		{
			name: "a transport failure during recent_context",
			args: recentContext,
			err:  graphitiSearchNetError(context.DeadlineExceeded),
			want: unavailable("recent context"),
		},
		{
			name: "a transport failure during entity_by_label",
			args: `{"search_type":"entity_by_label","query":"open ports","node_labels":["Host"],"message":"m"}`,
			err:  graphitiSearchNetError(context.DeadlineExceeded),
			want: unavailable("entity by label"),
		},
		{
			name: "a transport failure whose message runs over the cap",
			args: recentContext,
			err:  graphitiSearchNetError(errors.New("dial tcp: lookup " + strings.Repeat("g", 600) + ": no such host")),
			want: "Graphiti knowledge graph is temporarily unavailable (recent context search failed: failed to perform request: " +
				`Post "http://graphiti-neo4j/search/recent-context": dial tcp: lookup ` + strings.Repeat("g", 386) +
				"... [truncated full size is 740 bytes]); continuing without historical context.",
		},
		{
			name: "a 5xx with a short body",
			args: recentContext,
			err:  graphitiSearchStatusError(502, "<html><body>502 Bad Gateway</body></html>\n"),
			want: "Graphiti knowledge graph returned HTTP 502 and is likely temporarily unavailable; continuing without historical context. " +
				"Response: <html><body>502 Bad Gateway</body></html>",
		},
		{
			name: "a 5xx with a body over the cap",
			args: recentContext,
			err:  graphitiSearchStatusError(500, strings.Repeat("x", 2000)),
			want: "Graphiti knowledge graph returned HTTP 500 and is likely temporarily unavailable; continuing without historical context. " +
				"Response: " + cut("x"),
		},
		{
			name:     "a 4xx with a body over the cap",
			args:     recentContext,
			err:      graphitiSearchStatusError(400, strings.Repeat("y", 2000)),
			wantHard: true,
			want:     "graphiti API request failed with status 400: " + cut("y"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewGraphitiSearchTool(1, nil, nil, graphitiSearchGroupID, &graphitiSearchStub{enabled: true, err: tt.err})

			result, err := tool.Handle(t.Context(), GraphitiSearchToolName, []byte(tt.args))

			if tt.wantHard {
				if err == nil {
					t.Fatalf("a 4xx must stay a hard failure, got result %q", result)
				}
				if err.Error() != tt.want {
					t.Fatalf("error:\n%q\nwant:\n%q", err, tt.want)
				}
				return
			}
			if err != nil {
				t.Fatalf("an outage must degrade to a result, got error: %v", err)
			}
			if result != tt.want {
				t.Fatalf("result:\n%q\nwant:\n%q", result, tt.want)
			}
		})
	}
}

func TestGraphitiSearch_ParseGraphitiTime_AcceptsRFC3339AndReadsZoneLessNearMissesAsUTC(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantUTC bool
	}{
		{
			name: "rfc3339 in utc", input: "2026-07-25T11:53:34Z",
			want: time.Date(2026, 7, 25, 11, 53, 34, 0, time.UTC), wantUTC: true,
		},
		{
			name: "rfc3339 with an offset", input: "2026-07-25T11:53:34+03:00",
			want: time.Date(2026, 7, 25, 8, 53, 34, 0, time.UTC),
		},
		{
			name: "no zone designator", input: "2026-07-24T11:53:34",
			want: time.Date(2026, 7, 24, 11, 53, 34, 0, time.UTC), wantUTC: true,
		},
		{
			name: "space-separated without a zone", input: "2026-07-24 11:53:34",
			want: time.Date(2026, 7, 24, 11, 53, 34, 0, time.UTC), wantUTC: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGraphitiTime(tt.input)
			if err != nil {
				t.Fatalf("parseGraphitiTime(%q) failed: %v", tt.input, err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("parseGraphitiTime(%q) = %s, want %s", tt.input, got, tt.want)
			}
			if tt.wantUTC && got.Location() != time.UTC {
				t.Errorf("parseGraphitiTime(%q) is in %s, want UTC like the graph's own timestamps", tt.input, got.Location())
			}
		})
	}
}

func TestGraphitiSearch_ParseGraphitiTime_ReportsTheRFC3339ErrorWhenNothingMatches(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{name: "not a date", input: "not-a-date", wantErr: `parsing time "not-a-date" as "2006-01-02T15:04:05Z07:00"`},
		{name: "an empty value", input: "", wantErr: `parsing time "" as "2006-01-02T15:04:05Z07:00"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGraphitiTime(tt.input)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("parseGraphitiTime(%q) = %s, %v; want an error containing %q", tt.input, got, err, tt.wantErr)
			}
		})
	}
}

// The message is set in both rows but is not a search parameter, so it is left out.
func TestGraphitiSearch_RetrieverInput_CarriesOnlyTheParametersThatWereSet(t *testing.T) {
	five, three, two := Int64(5), Int64(3), Int64(2)
	tests := []struct {
		name string
		args GraphitiSearchAction
		want map[string]any
	}{
		{
			name: "only the required parameters",
			args: GraphitiSearchAction{SearchType: "recent_context", Query: "open ports", Message: "m"},
			want: map[string]any{"query": "open ports", "search_type": String("recent_context"), "group_id": "g"},
		},
		{
			name: "every parameter",
			args: GraphitiSearchAction{
				SearchType: "entity_relationships", Query: "open ports", MaxResults: &five,
				TimeStart: "2026-07-24T11:53:34Z", TimeEnd: "2026-07-25T11:53:34Z",
				CenterNodeUUID: "f7b95dfc-ee58-4a8b-8d85-582cf117b4df", MaxDepth: &three,
				NodeLabels: []string{"Host"}, EdgeTypes: []string{"HAS_PORT"}, DiversityLevel: "high",
				MinMentions: &two, RecencyWindow: "7d", Message: "m",
			},
			want: map[string]any{
				"query": "open ports", "search_type": String("entity_relationships"), "group_id": "g",
				"max_results": 5, "time_start": "2026-07-24T11:53:34Z", "time_end": "2026-07-25T11:53:34Z",
				"center_node_uuid": "f7b95dfc-ee58-4a8b-8d85-582cf117b4df", "max_depth": 3,
				"node_labels": []string{"Host"}, "edge_types": []string{"HAS_PORT"},
				"diversity_level": String("high"), "min_mentions": 2, "recency_window": String("7d"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.retrieverInput("g"); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("retrieverInput =\n%#v\nwant\n%#v", got, tt.want)
			}
		})
	}
}

func graphitiSearchLines(lines ...string) string { return strings.Join(lines, "\n") }

// Subtests are keyed by formatter; every entity and community listing shows the UUID entity_relationships needs.
func TestGraphitiSearch_FormatGraphiti_RendersTheWholeListingTheAgentReads(t *testing.T) {
	start := time.Date(2026, 7, 24, 11, 53, 34, 0, time.UTC)
	validAt := time.Date(2026, 7, 24, 9, 0, 0, 0, time.UTC)
	window := graphiti.TimeWindow{Start: start, End: start.Add(24 * time.Hour)}
	edges := []graphiti.EdgeResult{
		{
			Name: "HAS_PORT", Fact: "10.0.0.5 exposes 443", SourceNodeUUID: "node-a", TargetNodeUUID: "node-p",
			CreatedAt: start, ValidAt: &validAt,
		},
		{
			Name: "RUNS", Fact: "443 runs nginx", SourceNodeUUID: "node-p", TargetNodeUUID: "node-b",
			CreatedAt: time.Date(2026, 7, 24, 12, 10, 0, 0, time.UTC),
		},
	}
	center := graphiti.NodeResult{UUID: "node-c", Name: "dmz", Summary: "perimeter segment"}
	nodes := []graphiti.NodeResult{
		{UUID: "node-a", Name: "10.0.0.5", Labels: []string{"Host"}, Summary: "web host", Attributes: map[string]any{"os": "linux"}},
		{UUID: "node-b", Name: "nginx", Labels: []string{"Service"}, Summary: "web server"},
	}
	longContent := strings.Repeat("a", 250)
	episodes := []graphiti.EpisodeResult{
		{
			Source: "pentester", SourceDescription: "nmap scan", Content: "443/tcp open https",
			CreatedAt: time.Date(2026, 7, 24, 11, 50, 0, 0, time.UTC),
		},
		{
			Source: "coder", SourceDescription: "exploit draft", Content: longContent,
			CreatedAt: time.Date(2026, 7, 24, 12, 5, 0, 0, time.UTC),
		},
	}
	atCap := strings.Repeat("c", 200)
	diverseEpisodes := []graphiti.EpisodeResult{
		{Source: "pentester", SourceDescription: "nmap scan", Content: atCap},
		{Source: "coder", SourceDescription: "exploit draft", Content: strings.Repeat("d", 201)},
	}
	one := []float64{0.9}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "temporal window results",
			got: FormatGraphitiTemporalResults(&graphiti.TemporalSearchResponse{
				Edges: edges, EdgeScores: one, Nodes: nodes, NodeScores: one,
				Episodes: episodes, EpisodeScores: one, TimeWindow: window,
			}, "open ports"),
			want: graphitiSearchLines(
				"# Temporal Search Results",
				"",
				"**Query:** open ports",
				"",
				"**Time Window:** 2026-07-24T11:53:34Z to 2026-07-25T11:53:34Z",
				"",
				"## Facts & Relationships",
				"",
				"1. **HAS_PORT** (score: 0.900)",
				"   - Fact: 10.0.0.5 exposes 443",
				"   - Created: 2026-07-24T11:53:34Z",
				"   - Valid At: 2026-07-24T09:00:00Z",
				"",
				"2. **RUNS**",
				"   - Fact: 443 runs nginx",
				"   - Created: 2026-07-24T12:10:00Z",
				"",
				"## Entities",
				"",
				"1. **10.0.0.5** (score: 0.900)",
				"   - UUID: node-a",
				"   - Labels: [Host]",
				"   - Summary: web host",
				"   - Attributes: map[os:linux]",
				"",
				"2. **nginx**",
				"   - UUID: node-b",
				"   - Labels: [Service]",
				"   - Summary: web server",
				"",
				"## Agent Responses & Tool Executions",
				"",
				"1. **pentester** (score: 0.900)",
				"   - Description: nmap scan",
				"   - Created: 2026-07-24T11:50:00Z",
				"   - Content:",
				"```",
				"443/tcp open https",
				"```",
				"",
				"2. **coder**",
				"   - Description: exploit draft",
				"   - Created: 2026-07-24T12:05:00Z",
				"   - Content:",
				"```",
				longContent,
				"```",
				"",
				"",
			),
		},
		{
			name: "entity relationship results",
			got: FormatGraphitiEntityRelationshipResults(&graphiti.EntityRelationshipSearchResponse{
				CenterNode: &center, Edges: edges, EdgeDistances: one, Nodes: nodes, NodeDistances: one,
			}, "open ports"),
			want: graphitiSearchLines(
				"# Entity Relationship Search Results",
				"",
				"**Query:** open ports",
				"",
				"## Center Node: dmz",
				"- UUID: node-c",
				"- Summary: perimeter segment",
				"",
				"## Related Facts & Relationships",
				"",
				"1. **HAS_PORT** (distance: 0.900)",
				"   - Fact: 10.0.0.5 exposes 443",
				"   - Source: node-a",
				"   - Target: node-p",
				"",
				"2. **RUNS**",
				"   - Fact: 443 runs nginx",
				"   - Source: node-p",
				"   - Target: node-b",
				"",
				"## Related Entities",
				"",
				"1. **10.0.0.5** (distance: 0.900)",
				"   - UUID: node-a",
				"   - Labels: [Host]",
				"   - Summary: web host",
				"",
				"2. **nginx**",
				"   - UUID: node-b",
				"   - Labels: [Service]",
				"   - Summary: web server",
				"",
				"",
			),
		},
		{
			name: "diverse results",
			got: FormatGraphitiDiverseResults(&graphiti.DiverseSearchResponse{
				Communities: []graphiti.CommunityResult{
					{UUID: "comm-a", Name: "web tier", Summary: "hosts serving 443"},
					{UUID: "comm-b", Name: "db tier", Summary: "hosts serving 5432"},
				},
				CommunityMMRScores: one, Edges: edges, EdgeMMRScores: one, Episodes: diverseEpisodes, EpisodeScores: one,
			}, "open ports"),
			want: graphitiSearchLines(
				"# Diverse Search Results",
				"",
				"**Query:** open ports",
				"",
				"## Communities (Context Clusters)",
				"",
				"1. **web tier** (MMR score: 0.900)",
				"   - UUID: comm-a",
				"   - Summary: hosts serving 443",
				"",
				"2. **db tier**",
				"   - UUID: comm-b",
				"   - Summary: hosts serving 5432",
				"",
				"## Diverse Facts",
				"",
				"1. **HAS_PORT** (MMR score: 0.900)",
				"   - Fact: 10.0.0.5 exposes 443",
				"",
				"2. **RUNS**",
				"   - Fact: 443 runs nginx",
				"",
				"## Diverse Agent Activity",
				"",
				"1. **pentester** (score: 0.900)",
				"   - Description: nmap scan",
				"   - Content: "+atCap,
				"",
				"2. **coder**",
				"   - Description: exploit draft",
				"   - Content: "+strings.Repeat("d", 200)+"...",
				"",
				"",
			),
		},
		{
			name: "episode context results",
			got: FormatGraphitiEpisodeContextResults(&graphiti.EpisodeContextSearchResponse{
				Episodes: episodes, RerankerScores: one, MentionedNodes: nodes, MentionedNodeScores: one,
			}, "open ports"),
			want: graphitiSearchLines(
				"# Episode Context Results",
				"",
				"**Query:** open ports",
				"",
				"## Relevant Agent Activity",
				"",
				"1. **pentester** (relevance: 0.900)",
				"   - Time: 2026-07-24T11:50:00Z",
				"   - Description: nmap scan",
				"   - Content:",
				"```",
				"443/tcp open https",
				"```",
				"",
				"2. **coder**",
				"   - Time: 2026-07-24T12:05:00Z",
				"   - Description: exploit draft",
				"   - Content:",
				"```",
				longContent,
				"```",
				"",
				"## Mentioned Entities",
				"",
				"- **10.0.0.5** (relevance: 0.900) (UUID: node-a): web host",
				"- **nginx** (UUID: node-b): web server",
				"",
			),
		},
		{
			name: "successful tools results",
			got: FormatGraphitiSuccessfulToolsResults(&graphiti.SuccessfulToolsSearchResponse{
				Episodes: episodes, EpisodeScores: one, Edges: edges, EdgeMentionCounts: []float64{3},
			}, "open ports"),
			want: graphitiSearchLines(
				"# Successful Tools & Techniques",
				"",
				"**Query:** open ports",
				"",
				"## Successful Executions",
				"",
				"1. **pentester** (score: 0.900)",
				"   - Description: nmap scan",
				"   - Command/Output:",
				"```",
				"443/tcp open https",
				"```",
				"",
				"2. **coder**",
				"   - Description: exploit draft",
				"   - Command/Output:",
				"```",
				longContent,
				"```",
				"",
				"## Related Facts (Success Indicators)",
				"",
				"- **HAS_PORT** (mentions: 3): 10.0.0.5 exposes 443",
				"- **RUNS**: 443 runs nginx",
				"",
			),
		},
		{
			name: "recent context results",
			got: FormatGraphitiRecentContextResults(&graphiti.RecentContextSearchResponse{
				Nodes: nodes, NodeScores: one, Edges: edges, EdgeScores: one,
				Episodes: episodes, EpisodeScores: one, TimeWindow: window,
			}, "open ports"),
			want: graphitiSearchLines(
				"# Recent Context",
				"",
				"**Query:** open ports",
				"",
				"**Time Window:** 2026-07-24T11:53:34Z to 2026-07-25T11:53:34Z",
				"",
				"## Recently Discovered Entities",
				"",
				"1. **10.0.0.5** (score: 0.900)",
				"   - UUID: node-a",
				"   - Labels: [Host]",
				"   - Summary: web host",
				"",
				"2. **nginx**",
				"   - UUID: node-b",
				"   - Labels: [Service]",
				"   - Summary: web server",
				"",
				"## Recent Facts",
				"",
				"- **HAS_PORT** (score: 0.900): 10.0.0.5 exposes 443",
				"- **RUNS**: 443 runs nginx",
				"## Recent Activity",
				"",
				"- **pentester** (score: 0.900): nmap scan",
				"- **coder**: exploit draft",
				"",
			),
		},
		{
			name: "entity by label results",
			got: FormatGraphitiEntityByLabelResults(&graphiti.EntityByLabelSearchResponse{
				Nodes: nodes, NodeScores: one, Edges: edges, EdgeScores: one,
			}, "open ports"),
			want: graphitiSearchLines(
				"# Entity Inventory Search",
				"",
				"**Query:** open ports",
				"",
				"## Matching Entities",
				"",
				"1. **10.0.0.5** (score: 0.900)",
				"   - UUID: node-a",
				"   - Labels: [Host]",
				"   - Summary: web host",
				"   - Attributes: map[os:linux]",
				"",
				"2. **nginx**",
				"   - UUID: node-b",
				"   - Labels: [Service]",
				"   - Summary: web server",
				"",
				"## Associated Facts",
				"",
				"- **HAS_PORT** (score: 0.900): 10.0.0.5 exposes 443",
				"- **RUNS**: 443 runs nginx",
				"",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got == tt.want {
				return
			}
			got, want := strings.Split(tt.got, "\n"), strings.Split(tt.want, "\n")
			line := 0
			for line < len(got) && line < len(want) && got[line] == want[line] {
				line++
			}
			at := func(lines []string) string {
				if line < len(lines) {
					return fmt.Sprintf("%q", lines[line])
				}
				return "the end of the text"
			}
			t.Fatalf("the listing first differs at line %d: got %s, want %s\nwhole listing:\n%s", line+1, at(got), at(want), tt.got)
		})
	}
}
