package gemini

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/httputil"
)

func TestConfigLoading(t *testing.T) {
	cfg := &config.Config{
		GeminiAPIKey:    "test-key",
		GeminiServerURL: "https://generativelanguage.googleapis.com",
	}

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	rawConfig := prov.GetRawConfig()
	if len(rawConfig) == 0 {
		t.Fatal("Raw config should not be empty")
	}

	providerConfig = prov.GetProviderConfig()
	if providerConfig == nil {
		t.Fatal("Provider config should not be nil")
	}

	for _, agentType := range pconfig.AllAgentTypes {
		model := prov.Model(agentType)
		if model == "" {
			t.Errorf("Agent type %v should have a model assigned", agentType)
		}
	}

	for _, agentType := range pconfig.AllAgentTypes {
		priceInfo := prov.GetPriceInfo(agentType)
		if priceInfo == nil {
			t.Errorf("Agent type %v should have price information", agentType)
		} else {
			if priceInfo.Input <= 0 || priceInfo.Output <= 0 {
				t.Errorf("Agent type %v should have positive input (%f) and output (%f) prices",
					agentType, priceInfo.Input, priceInfo.Output)
			}
		}
	}
}

func TestProviderType(t *testing.T) {
	cfg := &config.Config{
		GeminiAPIKey:    "test-key",
		GeminiServerURL: "https://generativelanguage.googleapis.com",
	}

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	if prov.Type() != provider.ProviderGemini {
		t.Errorf("Expected provider type %v, got %v", provider.ProviderGemini, prov.Type())
	}
}

func TestModelsLoading(t *testing.T) {
	models, err := DefaultModels()
	if err != nil {
		t.Fatalf("Failed to load models: %v", err)
	}

	if len(models) == 0 {
		t.Fatal("Models list should not be empty")
	}

	for _, model := range models {
		if model.Name == "" {
			t.Error("Model name should not be empty")
		}

		if model.Price == nil {
			t.Errorf("Model %s should have price information", model.Name)
			continue
		}

		if model.Price.Input != 0 || model.Price.Output != 0 { // exclude totally free models
			if model.Price.Input <= 0 {
				t.Errorf("Model %s should have positive input price", model.Name)
			}

			if model.Price.Output <= 0 {
				t.Errorf("Model %s should have positive output price", model.Name)
			}
		}
	}
}

func TestGeminiSpecificFeatures(t *testing.T) {
	models, err := DefaultModels()
	if err != nil {
		t.Fatalf("Failed to load models: %v", err)
	}

	// Test that we have current Gemini models
	expectedModels := []string{"gemini-2.5-flash", "gemini-2.5-pro", "gemini-2.0-flash"}
	for _, expectedModel := range expectedModels {
		found := false
		for _, model := range models {
			if model.Name == expectedModel {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected model %s not found in models list", expectedModel)
		}
	}

	// Test default agent model
	if GeminiAgentModel != "gemini-2.5-flash" {
		t.Errorf("Expected default agent model to be gemini-2.5-flash, got %s", GeminiAgentModel)
	}
}

func TestGetUsage(t *testing.T) {
	cfg := &config.Config{
		GeminiAPIKey:    "test-key",
		GeminiServerURL: "https://generativelanguage.googleapis.com",
	}

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Test usage parsing with Google AI format
	usageInfo := map[string]any{
		"PromptTokens":     int32(100),
		"CompletionTokens": int32(50),
	}

	usage := prov.GetUsage(usageInfo)
	if usage.Input != 100 {
		t.Errorf("Expected input tokens 100, got %d", usage.Input)
	}
	if usage.Output != 50 {
		t.Errorf("Expected output tokens 50, got %d", usage.Output)
	}

	// Test with missing usage info
	emptyInfo := map[string]any{}
	usage = prov.GetUsage(emptyInfo)
	if !usage.IsZero() {
		t.Errorf("Expected zero tokens with empty usage info, got %s", usage.String())
	}
}

func TestAPIKeyTransportRoundTrip(t *testing.T) {
	tests := []struct {
		name             string
		serverURL        string
		apiKey           string
		requestURL       string
		requestQuery     string
		expectedScheme   string
		expectedHost     string
		expectedPath     string
		expectedQueryKey string
	}{
		{
			name:             "no custom server, adds API key to query only (no auth header for default host)",
			serverURL:        "",
			apiKey:           "test-api-key-123",
			requestURL:       "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent",
			requestQuery:     "",
			expectedScheme:   "https",
			expectedHost:     "generativelanguage.googleapis.com",
			expectedPath:     "/v1beta/models/gemini-pro:generateContent",
			expectedQueryKey: "test-api-key-123",
		},
		{
			name:             "custom server URL replaces base URL",
			serverURL:        "https://proxy.example.com/gemini",
			apiKey:           "my-key",
			requestURL:       "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent",
			requestQuery:     "",
			expectedScheme:   "https",
			expectedHost:     "proxy.example.com",
			expectedPath:     "/gemini/v1beta/models/gemini-pro:generateContent",
			expectedQueryKey: "my-key",
		},
		{
			name:             "custom server URL with trailing slash replaces base URL",
			serverURL:        "https://proxy.example.com/gemini/",
			apiKey:           "my-key",
			requestURL:       "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent",
			requestQuery:     "",
			expectedScheme:   "https",
			expectedHost:     "proxy.example.com",
			expectedPath:     "/gemini/v1beta/models/gemini-pro:generateContent",
			expectedQueryKey: "my-key",
		},
		{
			name:             "preserves existing query parameters",
			serverURL:        "https://proxy.example.com",
			apiKey:           "api-key",
			requestURL:       "https://generativelanguage.googleapis.com/v1/models",
			requestQuery:     "foo=bar&baz=qux",
			expectedScheme:   "https",
			expectedHost:     "proxy.example.com",
			expectedPath:     "/v1/models",
			expectedQueryKey: "api-key",
		},
		{
			name:             "does not override existing API key in query",
			serverURL:        "",
			apiKey:           "new-key",
			requestURL:       "https://generativelanguage.googleapis.com/v1/models",
			requestQuery:     "key=existing-key",
			expectedScheme:   "https",
			expectedHost:     "generativelanguage.googleapis.com",
			expectedPath:     "/v1/models",
			expectedQueryKey: "existing-key", // should keep existing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create mock round tripper that captures the request
			var capturedReq *http.Request
			mockRT := &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					capturedReq = req
					return &http.Response{
						StatusCode: 200,
						Body:       http.NoBody,
						Header:     make(http.Header),
					}, nil
				},
			}

			transport := &httputil.ApiKeyTransport{
				Transport: mockRT,
				APIKey:    tt.apiKey,
				BaseURL:   tt.serverURL,
				ProxyURL:  "",
			}

			// create test request
			reqURL := tt.requestURL
			if tt.requestQuery != "" {
				reqURL += "?" + tt.requestQuery
			}
			req, err := http.NewRequest("POST", reqURL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// execute RoundTrip
			_, err = transport.RoundTrip(req)
			if err != nil {
				t.Fatalf("RoundTrip failed: %v", err)
			}

			// verify captured request
			if capturedReq == nil {
				t.Fatal("Request was not captured")
			}

			if capturedReq.URL.Scheme != tt.expectedScheme {
				t.Errorf("Expected scheme %s, got %s", tt.expectedScheme, capturedReq.URL.Scheme)
			}

			if capturedReq.URL.Host != tt.expectedHost {
				t.Errorf("Expected host %s, got %s", tt.expectedHost, capturedReq.URL.Host)
			}

			if capturedReq.URL.Path != tt.expectedPath {
				t.Errorf("Expected path %s, got %s", tt.expectedPath, capturedReq.URL.Path)
			}

			queryKey := capturedReq.URL.Query().Get("key")
			if queryKey != tt.expectedQueryKey {
				t.Errorf("Expected query key %s, got %s", tt.expectedQueryKey, queryKey)
			}

			// verify original query parameters are preserved
			if tt.requestQuery != "" {
				originalQuery, _ := url.ParseQuery(tt.requestQuery)
				for k, v := range originalQuery {
					if k == "key" {
						continue // key may be added by transport
					}
					capturedValues := capturedReq.URL.Query()[k]
					if len(capturedValues) != len(v) {
						t.Errorf("Query parameter %s: expected %v, got %v", k, v, capturedValues)
					}
				}
			}
		})
	}
}

// mockRoundTripper is a mock implementation of http.RoundTripper for testing
type mockRoundTripper struct {
	roundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestAPIKeyTransportWithMockServer(t *testing.T) {
	// track received requests
	var receivedRequests []*http.Request
	var mu sync.Mutex

	// create test HTTP server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		receivedRequests = append(receivedRequests, r.Clone(r.Context()))
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer testServer.Close()

	// parse test server URL
	serverURL, err := url.Parse(testServer.URL)
	if err != nil {
		t.Fatalf("Failed to parse test server URL: %v", err)
	}

	// create transport with custom server
	transport := &httputil.ApiKeyTransport{
		Transport: http.DefaultTransport,
		APIKey:    "test-api-key-789",
		BaseURL:   testServer.URL,
		ProxyURL:  "",
	}

	// create HTTP client with our transport
	client := &http.Client{Transport: transport}

	// make request to Google API endpoint (will be redirected to test server)
	req, err := http.NewRequest("GET", "https://generativelanguage.googleapis.com/v1beta/models/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// verify request was received by test server
	mu.Lock()
	defer mu.Unlock()

	if len(receivedRequests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(receivedRequests))
	}

	capturedReq := receivedRequests[0]

	// verify URL was rewritten to test server
	if capturedReq.Host != serverURL.Host {
		t.Errorf("Expected host %s, got %s", serverURL.Host, capturedReq.Host)
	}

	// verify API key was added
	if key := capturedReq.URL.Query().Get("key"); key != "test-api-key-789" {
		t.Errorf("Expected API key test-api-key-789, got %s", key)
	}

	// verify original path was preserved
	if !strings.Contains(capturedReq.URL.Path, "/v1beta/models/test") {
		t.Errorf("Expected path to contain /v1beta/models/test, got %s", capturedReq.URL.Path)
	}
}

func TestGeminiProviderWithProxyConfiguration(t *testing.T) {
	// create provider config
	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	// test provider creation with proxy settings
	testCases := []struct {
		name      string
		proxyURL  string
		serverURL string
		wantErr   bool
	}{
		{
			name:      "valid configuration without proxy",
			proxyURL:  "",
			serverURL: "https://generativelanguage.googleapis.com",
			wantErr:   false,
		},
		{
			name:      "valid configuration with proxy",
			proxyURL:  "http://proxy.example.com:8080",
			serverURL: "https://generativelanguage.googleapis.com",
			wantErr:   false,
		},
		{
			name:      "valid configuration with custom server and proxy",
			proxyURL:  "http://localhost:8888",
			serverURL: "https://litellm.proxy.com/v1",
			wantErr:   false,
		},
		{
			name:      "invalid server URL",
			proxyURL:  "",
			serverURL: "://invalid-url",
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				GeminiAPIKey:    "test-key-" + tc.name,
				GeminiServerURL: tc.serverURL,
				ProxyURL:        tc.proxyURL,
			}

			prov, err := New(cfg, providerConfig)

			if tc.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if prov == nil {
				t.Fatal("Provider should not be nil")
			}

			if prov.Type() != provider.ProviderGemini {
				t.Errorf("Expected provider type Gemini, got %v", prov.Type())
			}
		})
	}
}
