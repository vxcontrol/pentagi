package tools

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
)

const testTraversaalAPIKey = "test-key"

func testTraversaalConfig() *config.Config {
	return &config.Config{TraversaalAPIKey: testTraversaalAPIKey}
}

func TestTraversaalHandle(t *testing.T) {
	var seenRequest bool
	var receivedMethod string
	var receivedContentType string
	var receivedAPIKey string
	var receivedBody []byte

	mockMux := http.NewServeMux()
	mockMux.HandleFunc("/live/predict", func(w http.ResponseWriter, r *http.Request) {
		seenRequest = true
		receivedMethod = r.Method
		receivedContentType = r.Header.Get("Content-Type")
		receivedAPIKey = r.Header.Get("x-api-key")

		var err error
		receivedBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"response_text":"answer text","web_url":["https://a.com","https://b.com"]}}`))
	})

	proxy, err := newTestProxy("api-ares.traversaal.ai", mockMux)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
	defer proxy.Close()

	flowID := int64(1)
	taskID := int64(10)
	subtaskID := int64(20)
	slp := &searchLogProviderMock{}

	cfg := &config.Config{
		TraversaalAPIKey:    testTraversaalAPIKey,
		ProxyURL:            proxy.URL(),
		ExternalSSLCAPath:   proxy.CACertPath(),
		ExternalSSLInsecure: false,
	}

	trav := NewTraversaalTool(cfg, flowID, &taskID, &subtaskID, slp)

	ctx := PutAgentContext(t.Context(), database.MsgchainTypeSearcher)
	got, err := trav.Handle(
		ctx,
		TraversaalToolName,
		[]byte(`{"query":"test query","max_results":5,"message":"m"}`),
	)
	if err != nil {
		t.Fatalf("Handle() unexpected error: %v", err)
	}

	// Verify mock handler was called
	if !seenRequest {
		t.Fatal("request was not intercepted by proxy - mock handler was not called")
	}

	// Verify request was built correctly
	if receivedMethod != http.MethodPost {
		t.Errorf("request method = %q, want POST", receivedMethod)
	}
	if receivedContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", receivedContentType)
	}
	if receivedAPIKey != testTraversaalAPIKey {
		t.Errorf("x-api-key = %q, want %q", receivedAPIKey, testTraversaalAPIKey)
	}
	if !strings.Contains(string(receivedBody), `"query":"test query"`) {
		t.Errorf("request body = %q, expected to contain query", string(receivedBody))
	}

	// Verify response was parsed correctly
	if !strings.Contains(got, "# Answer") {
		t.Errorf("result missing '# Answer' section: %q", got)
	}
	if !strings.Contains(got, "# Links") {
		t.Errorf("result missing '# Links' section: %q", got)
	}
	if !strings.Contains(got, "answer text") {
		t.Errorf("result missing expected text 'answer text': %q", got)
	}
	if !strings.Contains(got, "https://a.com") {
		t.Errorf("result missing expected link 'https://a.com': %q", got)
	}
	if !strings.Contains(got, "https://b.com") {
		t.Errorf("result missing expected link 'https://b.com': %q", got)
	}

	// Verify search log was written with agent context
	if slp.calls != 1 {
		t.Errorf("PutLog() calls = %d, want 1", slp.calls)
	}
	if slp.engine != database.SearchengineTypeTraversaal {
		t.Errorf("engine = %q, want %q", slp.engine, database.SearchengineTypeTraversaal)
	}
	if slp.query != "test query" {
		t.Errorf("logged query = %q, want %q", slp.query, "test query")
	}
	if slp.parentType != database.MsgchainTypeSearcher {
		t.Errorf("parent agent type = %q, want %q", slp.parentType, database.MsgchainTypeSearcher)
	}
	if slp.currType != database.MsgchainTypeSearcher {
		t.Errorf("current agent type = %q, want %q", slp.currType, database.MsgchainTypeSearcher)
	}
	if slp.taskID == nil || *slp.taskID != taskID {
		t.Errorf("task ID = %v, want %d", slp.taskID, taskID)
	}
	if slp.subtaskID == nil || *slp.subtaskID != subtaskID {
		t.Errorf("subtask ID = %v, want %d", slp.subtaskID, subtaskID)
	}
}

func TestTraversaalIsAvailable(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{
			name: "available when API key is set",
			cfg:  testTraversaalConfig(),
			want: true,
		},
		{
			name: "unavailable when API key is empty",
			cfg:  &config.Config{},
			want: false,
		},
		{
			name: "unavailable when nil config",
			cfg:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trav := &traversaal{cfg: tt.cfg}
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
	t.Run("invalid json", func(t *testing.T) {
		trav := &traversaal{cfg: testTraversaalConfig()}
		_, err := trav.Handle(t.Context(), TraversaalToolName, []byte("{"))
		if err == nil || !strings.Contains(err.Error(), "failed to unmarshal") {
			t.Fatalf("expected unmarshal error, got: %v", err)
		}
	})

	t.Run("search error swallowed", func(t *testing.T) {
		var seenRequest bool
		mockMux := http.NewServeMux()
		mockMux.HandleFunc("/live/predict", func(w http.ResponseWriter, r *http.Request) {
			seenRequest = true
			w.WriteHeader(http.StatusBadGateway)
		})

		proxy, err := newTestProxy("api-ares.traversaal.ai", mockMux)
		if err != nil {
			t.Fatalf("failed to create proxy: %v", err)
		}
		defer proxy.Close()

		trav := &traversaal{
			flowID: 1,
			cfg: &config.Config{
				TraversaalAPIKey:    testTraversaalAPIKey,
				ProxyURL:            proxy.URL(),
				ExternalSSLCAPath:   proxy.CACertPath(),
				ExternalSSLInsecure: false,
			},
		}

		result, err := trav.Handle(
			t.Context(),
			TraversaalToolName,
			[]byte(`{"query":"q","max_results":5,"message":"m"}`),
		)
		if err != nil {
			t.Fatalf("Handle() unexpected error: %v", err)
		}

		// Verify mock handler was called (request was intercepted)
		if !seenRequest {
			t.Error("request was not intercepted by proxy - mock handler was not called")
		}

		// Verify error was swallowed and returned as string
		if !strings.Contains(result, "failed to search in traversaal") {
			t.Errorf("Handle() = %q, expected swallowed error message", result)
		}
	})
}
