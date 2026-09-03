package aimlapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
)

func testConfig(serverURL string) *config.Config {
	return &config.Config{
		AIMLAPIKey:       "test-key",
		AIMLAPIServerURL: serverURL,
	}
}

func newTestProvider(t *testing.T, cfg *config.Config) provider.Provider {
	t.Helper()

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(cfg, provider.DefaultProviderNameAIMLAPI, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	return prov
}

func TestConfigLoading(t *testing.T) {
	prov := newTestProvider(t, testConfig("https://api.aimlapi.com/v1"))

	if len(prov.GetRawConfig()) == 0 {
		t.Fatal("Raw config should not be empty")
	}

	if prov.GetProviderConfig() == nil {
		t.Fatal("Provider config should not be nil")
	}

	for _, agentType := range pconfig.AllAgentTypes {
		if model := prov.Model(agentType); model == "" {
			t.Errorf("Agent type %v should have a model assigned", agentType)
		}
	}

	for _, agentType := range pconfig.AllAgentTypes {
		priceInfo := prov.GetPriceInfo(agentType)
		if priceInfo == nil {
			t.Errorf("Agent type %v should have price information", agentType)
			continue
		}
		if priceInfo.Input <= 0 || priceInfo.Output <= 0 {
			t.Errorf("Agent type %v should have positive input (%f) and output (%f) prices",
				agentType, priceInfo.Input, priceInfo.Output)
		}
	}
}

func TestProviderType(t *testing.T) {
	prov := newTestProvider(t, testConfig("https://api.aimlapi.com/v1"))

	if prov.Type() != provider.ProviderAIMLAPI {
		t.Errorf("Expected provider type %q, got %q", provider.ProviderAIMLAPI, prov.Type())
	}
	if prov.Name() != provider.DefaultProviderNameAIMLAPI {
		t.Errorf("Expected provider name %q, got %q", provider.DefaultProviderNameAIMLAPI, prov.Name())
	}
}

func TestMissingAPIKey(t *testing.T) {
	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	cfg := &config.Config{AIMLAPIServerURL: "https://api.aimlapi.com/v1"}
	if _, err := New(cfg, provider.DefaultProviderNameAIMLAPI, providerConfig); err == nil {
		t.Fatal("Expected error when API key is missing")
	}
}

// TestModelsAreCatalogIds guards the models the config assigns to agent roles
// against the models.yml catalog. A model id that is not in the catalog is either
// a typo or an id retired upstream; either way the agent using it 404s at runtime
// with nothing in the test suite to catch it.
func TestModelsAreCatalogIds(t *testing.T) {
	models, err := DefaultModels()
	if err != nil {
		t.Fatalf("Failed to load models: %v", err)
	}

	known := make(map[string]struct{}, len(models))
	for _, model := range models {
		known[model.Name] = struct{}{}
	}

	prov := newTestProvider(t, testConfig("https://api.aimlapi.com/v1"))
	for _, agentType := range pconfig.AllAgentTypes {
		model := prov.Model(agentType)
		if _, ok := known[model]; !ok {
			t.Errorf("Agent type %v uses model %q, which is not in models.yml", agentType, model)
		}
	}
}

// TestPartnerIDShape asserts the gateway's partner-id contract, /^part_[A-Za-z0-9]{1,64}$/.
// A malformed id is not rejected by the API — it is silently treated as untagged
// traffic — so a typo here is invisible at runtime and only a test can catch it.
func TestPartnerIDShape(t *testing.T) {
	pattern := regexp.MustCompile(`^part_[A-Za-z0-9]{1,64}$`)
	if !pattern.MatchString(attributionPartnerID) {
		t.Errorf("Partner ID %q does not match %s", attributionPartnerID, pattern)
	}

	if got := attributionHeaders()["X-AIMLAPI-Partner-ID"]; got != attributionPartnerID {
		t.Errorf("Header carries partner ID %q, want %q", got, attributionPartnerID)
	}
}

// TestAttributionHeadersAreNotShared catches the classic mistake of handing every
// provider instance the same map and letting one mutation leak into all of them.
func TestAttributionHeadersAreNotShared(t *testing.T) {
	first := attributionHeaders()
	first["X-AIMLAPI-Partner-ID"] = "part_mutated"

	if got := attributionHeaders()["X-AIMLAPI-Partner-ID"]; got != attributionPartnerID {
		t.Errorf("Mutating one header map changed the next one: got %q", got)
	}
}

func TestAttributionHost(t *testing.T) {
	tests := []struct {
		baseURL string
		want    string
	}{
		{"https://api.aimlapi.com/v1", "api.aimlapi.com"},
		{"https://API.AIMLAPI.COM/v1", "api.aimlapi.com"},
		{"https://aimlapi.com/v1", "aimlapi.com"},
		// A proxy or self-hosted gateway that fronts us must not be tagged: its
		// traffic is not this integration's to claim.
		{"https://litellm.internal:4000/v1", ""},
		{"http://llm-server:8000/v1", ""},
		{"https://notaimlapi.com/v1", ""},
		{"https://aimlapi.com.evil.example/v1", ""},
		{"", ""},
	}

	for _, tt := range tests {
		if got := attributionHost(tt.baseURL); got != tt.want {
			t.Errorf("attributionHost(%q) = %q, want %q", tt.baseURL, got, tt.want)
		}
	}
}

func TestAttributionTransportTagsOurOriginOnly(t *testing.T) {
	var seen http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &attributionTransport{
		host:    "127.0.0.1",
		headers: attributionHeaders(),
	}
	client := &http.Client{Transport: transport}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	for key, want := range attributionHeaders() {
		if got := seen.Get(key); got != want {
			t.Errorf("Header %s = %q, want %q", key, got, want)
		}
	}

	// Same client, a host that is not ours: nothing may be attached.
	transport.host = "api.aimlapi.com"
	resp, err = client.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	for key := range attributionHeaders() {
		if got := seen.Get(key); got != "" {
			t.Errorf("Header %s leaked to a foreign host with value %q", key, got)
		}
	}
}

func TestAttributionDoesNotOverrideCallerHeaders(t *testing.T) {
	var seen http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{Transport: &attributionTransport{host: "127.0.0.1", headers: attributionHeaders()}}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("X-Title", "operator-supplied")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if got := seen.Get("X-Title"); got != "operator-supplied" {
		t.Errorf("Attribution overwrote a caller header: X-Title = %q", got)
	}
	if got := seen.Get("X-AIMLAPI-Partner-ID"); got != attributionPartnerID {
		t.Errorf("Attribution dropped the partner id when a caller header was present: %q", got)
	}
	if req.Header.Get("X-AIMLAPI-Partner-ID") != "" {
		t.Error("RoundTrip mutated the request it was given instead of tagging a clone")
	}
}

func TestWithAttributionDoesNotMutateSharedClient(t *testing.T) {
	base := &http.Client{}

	tagged := withAttribution(base, "https://api.aimlapi.com/v1")
	if tagged == base {
		t.Fatal("withAttribution returned the client it was given instead of a copy")
	}
	if base.Transport != nil {
		t.Error("withAttribution mutated the shared HTTP client's transport")
	}

	// A non-aimlapi base URL leaves the client untouched, attribution disabled.
	if got := withAttribution(base, "http://llm-server:8000/v1"); got != base {
		t.Error("withAttribution wrapped a client pointed at a foreign endpoint")
	}
}

// TestRequestOmitsUnsetSamplingParams is a regression test for a live failure
// mode of the gateway: it rejects `temperature: null`, `top_p: null` and
// `seed: null` with HTTP 400, while accepting them as absent. Clients that
// serialise unset optionals as explicit nulls therefore fail on every real call
// while a mocked test suite stays green. This drives the actual provider and
// inspects the bytes on the wire.
func TestRequestOmitsUnsetSamplingParams(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"1","object":"chat.completion","choices":[{"index":0,` +
			`"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
			`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer server.Close()

	prov := newTestProvider(t, testConfig(server.URL))

	// The searcher role sets temperature but no top_p and no seed, so those two
	// must be absent from the payload rather than present and null.
	if _, err := prov.Call(context.Background(), pconfig.OptionsTypeSearcher, "ping"); err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	for key, value := range body {
		if value == nil {
			t.Errorf("Request serialised %q as null; the gateway rejects null sampling params", key)
		}
	}
	for _, key := range []string{"top_p", "seed"} {
		if _, present := body[key]; present {
			t.Errorf("Request carries %q although the agent config leaves it unset", key)
		}
	}
	if _, present := body["temperature"]; !present {
		t.Error("Request dropped temperature although the agent config sets it")
	}
}

// TestToolsAreOmittedNotNulledOnFollowUpTurn is the agent-loop half of the same
// gateway behaviour: `"tools": null` is a 400. A turn that carries tools followed
// by a turn that clears them is the ordinary shape of an agent loop, so a client
// that nulls the cleared field succeeds on turn one and fails on every turn two.
func TestToolsAreOmittedNotNulledOnFollowUpTurn(t *testing.T) {
	var bodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
		}

		var body map[string]any
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		bodies = append(bodies, body)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"1","object":"chat.completion","choices":[{"index":0,` +
			`"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],` +
			`"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer server.Close()

	prov := newTestProvider(t, testConfig(server.URL))
	chain := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "scan 10.0.0.7")}
	tools := []llms.Tool{{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_open_ports",
			Description: "Return the list of open TCP ports on a host.",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{"host": map[string]any{"type": "string"}},
			},
		},
	}}

	ctx := context.Background()
	if _, err := prov.CallWithTools(ctx, pconfig.OptionsTypePentester, chain, tools, nil); err != nil {
		t.Fatalf("first turn failed: %v", err)
	}
	if _, err := prov.CallWithTools(ctx, pconfig.OptionsTypePentester, chain, nil, nil); err != nil {
		t.Fatalf("second turn failed: %v", err)
	}

	if len(bodies) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(bodies))
	}
	if _, present := bodies[0]["tools"]; !present {
		t.Error("First turn dropped the tools it was given")
	}
	if _, present := bodies[1]["tools"]; present {
		t.Errorf("Second turn serialised cleared tools as %v; the gateway rejects a null tools field",
			bodies[1]["tools"])
	}
	for i, body := range bodies {
		for key, value := range body {
			if value == nil {
				t.Errorf("Turn %d serialised %q as null; the gateway rejects null for most fields", i+1, key)
			}
		}
	}
}
