package tools

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"pentagi/pkg/database"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type searchLogProviderMock struct {
	calls      int
	engine     database.SearchengineType
	query      string
	result     string
	taskID     *int64
	subtaskID  *int64
	parentType database.MsgchainType
	currType   database.MsgchainType
}

func (m *searchLogProviderMock) PutLog(
	_ context.Context,
	initiator database.MsgchainType,
	executor database.MsgchainType,
	engine database.SearchengineType,
	query string,
	result string,
	taskID *int64,
	subtaskID *int64,
) (int64, error) {
	m.calls++
	m.parentType = initiator
	m.currType = executor
	m.engine = engine
	m.query = query
	m.result = result
	m.taskID = taskID
	m.subtaskID = subtaskID
	return 1, nil
}

func TestTraversaalSearchDoesNotMutateDefaultClient(t *testing.T) {
	originalTransport := http.DefaultClient.Transport

	trav := &traversaal{
		flowID:   1,
		apiKey:   "test-key",
		proxyURL: "http://127.0.0.1:19999", // non-empty to trigger new client path
	}

	// search will fail to connect (expected); the assertion is on global state.
	_, _ = trav.search(t.Context(), "test query")

	if http.DefaultClient.Transport != originalTransport {
		t.Error("http.DefaultClient.Transport was mutated; search must create a new http.Client instead")
	}
}

func TestTraversaalSearch_RequestBuildAndParse(t *testing.T) {
	originalTransport := http.DefaultClient.Transport
	defer func() { http.DefaultClient.Transport = originalTransport }()
	var seenRequest bool
	http.DefaultClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		seenRequest = true
		if req.Method != http.MethodPost {
			t.Fatalf("request method = %s, want POST", req.Method)
		}
		if req.URL.String() != traversaalURL {
			t.Fatalf("request url = %s, want %s", req.URL.String(), traversaalURL)
		}
		if got := req.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		if got := req.Header.Get("x-api-key"); got != "test-key" {
			t.Fatalf("x-api-key = %q, want test-key", got)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if !strings.Contains(string(body), `"query":"test query"`) {
			t.Fatalf("request body = %q, expected query payload", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(
				`{"data":{"response_text":"answer text","web_url":["https://a.com","https://b.com"]}}`,
			)),
			Header: make(http.Header),
		}, nil
	})

	trav := &traversaal{
		flowID:   1,
		apiKey:   "test-key",
		proxyURL: "",
	}

	got, err := trav.search(t.Context(), "test query")
	if err != nil {
		t.Fatalf("search() unexpected error: %v", err)
	}
	if !seenRequest {
		t.Fatal("custom transport was not used")
	}
	if !strings.Contains(got, "# Answer") || !strings.Contains(got, "https://a.com") {
		t.Fatalf("search() unexpected result format: %q", got)
	}
}

func TestTraversaalIsAvailable(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		want   bool
	}{
		{
			name:   "available when API key is set",
			apiKey: "test-key",
			want:   true,
		},
		{
			name:   "unavailable when API key is empty",
			apiKey: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trav := &traversaal{apiKey: tt.apiKey}
			if got := trav.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTraversaalParseHTTPResponse_StatusAndDecodeErrors(t *testing.T) {
	trav := &traversaal{flowID: 1}

	t.Run("status error", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("")),
		}
		_, err := trav.parseHTTPResponse(resp)
		if err == nil || !strings.Contains(err.Error(), "unexpected status code") {
			t.Fatalf("expected status code error, got: %v", err)
		}
	})

	t.Run("decode error", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("{invalid json")),
		}
		_, err := trav.parseHTTPResponse(resp)
		if err == nil || !strings.Contains(err.Error(), "failed to decode response body") {
			t.Fatalf("expected decode error, got: %v", err)
		}
	})
}

func TestTraversaalHandle_ValidationAndSwallowedError(t *testing.T) {
	trav := &traversaal{apiKey: "test-key"}

	t.Run("invalid json", func(t *testing.T) {
		_, err := trav.Handle(t.Context(), TraversaalToolName, []byte("{"))
		if err == nil || !strings.Contains(err.Error(), "failed to unmarshal") {
			t.Fatalf("expected unmarshal error, got: %v", err)
		}
	})

	t.Run("search error swallowed", func(t *testing.T) {
		originalTransport := http.DefaultClient.Transport
		defer func() { http.DefaultClient.Transport = originalTransport }()
		http.DefaultClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("dial failure")
		})

		result, err := trav.Handle(
			t.Context(),
			TraversaalToolName,
			[]byte(`{"query":"q","max_results":5,"message":"m"}`),
		)
		if err != nil {
			t.Fatalf("Handle() unexpected error: %v", err)
		}
		if !strings.Contains(result, "failed to search in traversaal") {
			t.Fatalf("Handle() = %q, expected swallowed error message", result)
		}
	})
}

func TestTraversaalHandle_Success_WritesSearchLogWithAgentContext(t *testing.T) {
	originalTransport := http.DefaultClient.Transport
	defer func() { http.DefaultClient.Transport = originalTransport }()
	http.DefaultClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(
				`{"data":{"response_text":"answer text","web_url":["https://a.com"]}}`,
			)),
			Header: make(http.Header),
		}, nil
	})

	taskID := int64(10)
	subtaskID := int64(20)
	slp := &searchLogProviderMock{}
	trav := &traversaal{
		flowID:    1,
		taskID:    &taskID,
		subtaskID: &subtaskID,
		apiKey:    "test-key",
		slp:       slp,
	}

	ctx := PutAgentContext(t.Context(), database.MsgchainTypeSearcher)
	got, err := trav.Handle(
		ctx,
		TraversaalToolName,
		[]byte(`{"query":"q","max_results":5,"message":"m"}`),
	)
	if err != nil {
		t.Fatalf("Handle() unexpected error: %v", err)
	}
	if !strings.Contains(got, "answer text") {
		t.Fatalf("Handle() result = %q, expected answer text", got)
	}
	if slp.calls != 1 {
		t.Fatalf("PutLog() calls = %d, want 1", slp.calls)
	}
	if slp.engine != database.SearchengineTypeTraversaal {
		t.Fatalf("engine = %q, want %q", slp.engine, database.SearchengineTypeTraversaal)
	}
	if slp.query != "q" {
		t.Fatalf("query = %q, want q", slp.query)
	}
}
