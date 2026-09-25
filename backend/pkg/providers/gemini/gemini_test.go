package gemini

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/providers/pconfig"
	"pentagi/pkg/providers/provider"

	"github.com/vxcontrol/langchaingo/llms"
)

func geminiWriteCAFile(t *testing.T, srv *httptest.Server) string {
	t.Helper()

	cert := srv.Certificate()
	if cert == nil {
		t.Fatal("test server has no certificate")
	}

	path := filepath.Join(t.TempDir(), "ca.pem")
	body := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write CA file: %v", err)
	}

	return path
}

func geminiNew(t *testing.T) (provider.Provider, *pconfig.ProviderConfig) {
	t.Helper()

	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	prov, err := New(&config.Config{
		GeminiAPIKey:    "test-key",
		GeminiServerURL: "https://generativelanguage.googleapis.com",
	}, provider.DefaultProviderNameGemini, providerConfig)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	return prov, providerConfig
}

func TestGemini_BundledRolesUseTheReplayedModelsAndPrices(t *testing.T) {
	cfg, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("load bundled config: %v", err)
	}

	wantModels := map[pconfig.ProviderOptionsType]string{
		pconfig.OptionsTypeSimple:       "gemini-3.1-flash-lite",
		pconfig.OptionsTypeSimpleJSON:   "gemini-3.1-flash-lite",
		pconfig.OptionsTypePrimaryAgent: "gemini-3.5-flash-lite",
		pconfig.OptionsTypeAssistant:    "gemini-3.5-flash-lite",
		pconfig.OptionsTypeGenerator:    "gemini-3.5-flash-lite",
		pconfig.OptionsTypeRefiner:      "gemini-3.5-flash-lite",
		pconfig.OptionsTypeAdviser:      "gemini-3.5-flash-lite",
		pconfig.OptionsTypeReflector:    "gemini-3.1-flash-lite",
		pconfig.OptionsTypeSearcher:     "gemini-3.1-flash-lite",
		pconfig.OptionsTypeEnricher:     "gemini-3.1-flash-lite",
		pconfig.OptionsTypeCoder:        "gemini-3.5-flash-lite",
		pconfig.OptionsTypeInstaller:    "gemini-3.5-flash-lite",
		pconfig.OptionsTypePentester:    "gemini-3.5-flash-lite",
	}
	if len(wantModels) != len(pconfig.AllAgentTypes) {
		t.Fatalf("test covers %d bundled roles, want %d", len(wantModels), len(pconfig.AllAgentTypes))
	}

	for _, role := range pconfig.AllAgentTypes {
		wantModel, ok := wantModels[role]
		if !ok {
			t.Errorf("missing expected model for %s", role)
			continue
		}
		options := llms.CallOptions{}
		for _, option := range cfg.GetOptionsForType(role) {
			option(&options)
		}
		if got := options.GetModel(); got != wantModel {
			t.Errorf("%s model = %q, want %q", role, got, wantModel)
		}
		wantInput, wantOutput, wantCacheRead := 0.25, 1.50, 0.025
		if wantModel == "gemini-3.5-flash-lite" {
			wantInput, wantOutput, wantCacheRead = 0.30, 2.50, 0.03
		}
		if got := cfg.GetPriceInfoForType(role); got == nil ||
			got.Input != wantInput || got.Output != wantOutput || got.CacheRead != wantCacheRead {
			t.Errorf("%s price = %+v, want %s catalogue price", role, got, wantModel)
		}
	}
}

func TestGemini_New_BuildsTheDoorWithAPricedModelForEveryAgent(t *testing.T) {
	prov, providerConfig := geminiNew(t)

	if prov.Type() != provider.ProviderGemini {
		t.Errorf("Expected provider type %v, got %v", provider.ProviderGemini, prov.Type())
	}
	if prov.Name() != provider.DefaultProviderNameGemini {
		t.Errorf("Expected provider name %v, got %v", provider.DefaultProviderNameGemini, prov.Name())
	}
	if len(prov.GetRawConfig()) == 0 {
		t.Fatal("Raw config should not be empty")
	}
	if prov.GetProviderConfig() != providerConfig {
		t.Fatal("Provider config should be the one New was given")
	}

	for _, agentType := range pconfig.AllAgentTypes {
		if prov.Model(agentType) == "" {
			t.Errorf("Agent type %v should have a model assigned", agentType)
		}

		priceInfo := prov.GetPriceInfo(agentType)
		if priceInfo == nil {
			t.Errorf("Agent type %v should have price information", agentType)
		} else if priceInfo.Input <= 0 || priceInfo.Output <= 0 {
			t.Errorf("Agent type %v should have positive input (%f) and output (%f) prices",
				agentType, priceInfo.Input, priceInfo.Output)
		}
	}
}

func TestGemini_New_RefusesAnUnparsableServerURL(t *testing.T) {
	providerConfig, err := DefaultProviderConfig()
	if err != nil {
		t.Fatalf("Failed to create provider config: %v", err)
	}

	_, err = New(&config.Config{GeminiAPIKey: "test-key", GeminiServerURL: "://invalid-url"},
		provider.DefaultProviderNameGemini, providerConfig)
	if err == nil || !strings.Contains(err.Error(), "failed to parse Gemini server URL") {
		t.Fatalf("Expected the server URL parse error, got %v", err)
	}
}

func TestGemini_GetUsage_ReadsInt32TokenCounts(t *testing.T) {
	prov, _ := geminiNew(t)

	usage := prov.GetUsage(map[string]any{
		"PromptTokens":     int32(100),
		"CompletionTokens": int32(50),
	})
	if usage.Input != 100 {
		t.Errorf("Expected input tokens 100, got %d", usage.Input)
	}
	if usage.Output != 50 {
		t.Errorf("Expected output tokens 50, got %d", usage.Output)
	}
}

func TestGemini_DefaultModels_ShipsANonEmptyPricedCatalogue(t *testing.T) {
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

func TestGemini_DefaultModels_ListsTheCurrentModelsAndTheDefaultOne(t *testing.T) {
	models, err := DefaultModels()
	if err != nil {
		t.Fatalf("Failed to load models: %v", err)
	}

	expectedModels := []string{
		"gemini-3.8-flash", "gemini-3.7-flash", "gemini-3.6-flash",
		"gemini-3.5-flash", "gemini-3.5-flash-lite",
		"gemini-3.1-pro-preview", "gemini-3.1-pro-preview-customtools",
		"gemini-3.1-flash-lite", "gemini-3-flash-preview",
		"gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.5-flash-lite",
		"gemma-4-31b-it", "gemma-4-26b-a4b-it",
	}
	if len(models) != len(expectedModels) {
		t.Errorf("catalogue has %d models, want %d", len(models), len(expectedModels))
	}
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

	if GeminiAgentModel != "gemini-3.1-flash-lite" {
		t.Errorf("Expected default agent model to be gemini-3.1-flash-lite, got %s", GeminiAgentModel)
	}
}

func TestGemini_NewHTTPClient_AppliesTheConfiguredTimeout(t *testing.T) {
	released := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		<-released
	}))
	t.Cleanup(func() {
		close(released)
		srv.Close()
	})

	client, err := newHTTPClient(&config.Config{
		HTTPClientTimeout: 1,
		GeminiServerURL:   srv.URL,
	})
	if err != nil {
		t.Fatalf("build client: %v", err)
	}

	failed := make(chan error, 1)
	go func() {
		resp, err := client.Get(srv.URL) //nolint:noctx
		if resp != nil {
			_ = resp.Body.Close()
		}
		failed <- err
	}()

	select {
	case err := <-failed:
		var netErr net.Error
		if !errors.As(err, &netErr) || !netErr.Timeout() {
			t.Fatalf("a hung response must be cut off by the configured timeout, got %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("request outlived the configured timeout, so it never reached the client")
	}
}

func TestGemini_NewHTTPClient_TrustsTheConfiguredCA(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	caPath := geminiWriteCAFile(t, srv)

	trusting, err := newHTTPClient(&config.Config{
		HTTPClientTimeout: 5,
		ExternalSSLCAPath: caPath,
		GeminiServerURL:   srv.URL,
	})
	if err != nil {
		t.Fatalf("build client with CA: %v", err)
	}

	resp, err := trusting.Get(srv.URL) //nolint:noctx
	if err != nil {
		t.Fatalf("a certificate signed by the configured CA must be accepted: %v", err)
	}
	_ = resp.Body.Close()

	plain, err := newHTTPClient(&config.Config{
		HTTPClientTimeout: 5,
		GeminiServerURL:   srv.URL,
	})
	if err != nil {
		t.Fatalf("build client without CA: %v", err)
	}

	resp, err = plain.Get(srv.URL) //nolint:noctx
	if err == nil {
		_ = resp.Body.Close()
	}
	var unknownAuthority x509.UnknownAuthorityError
	if !errors.As(err, &unknownAuthority) {
		t.Fatalf("without the configured CA the same certificate must be rejected as signed by an unknown authority, got %v", err)
	}
}

func TestGemini_NewHTTPClient_InjectsTheKeyAndRewritesTheBase(t *testing.T) {
	var (
		gotPath string
		gotKey  string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.URL.Query().Get("key")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client, err := newHTTPClient(&config.Config{
		HTTPClientTimeout: 5,
		GeminiAPIKey:      "gemini-key",
		GeminiServerURL:   srv.URL,
	})
	if err != nil {
		t.Fatalf("build client: %v", err)
	}

	resp, err := client.Get("https://generativelanguage.googleapis.com/v1beta/models") //nolint:noctx
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	_ = resp.Body.Close()

	if gotPath != "/v1beta/models" {
		t.Fatalf("request path %q lost the original path during the base URL rewrite", gotPath)
	}
	if gotKey != "gemini-key" {
		t.Fatalf("api key reached the wire as %q", gotKey)
	}
}

func geminiThinkingCall(t *testing.T, configured string, asked int) map[string]any {
	t.Helper()

	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"candidates":[{"content":{"parts":[{"text":"ok"}],"role":"model"},`+
			`"finishReason":"STOP","index":0}],`+
			`"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1,"totalTokenCount":2}}`)
	}))
	t.Cleanup(srv.Close)

	providerConfig, err := BuildProviderConfig(fmt.Appendf(nil,
		"simple:\n  model: gemini-3-pro\n%s  reasoning:\n    mode: budget\n    max_tokens: %d\n", configured, asked))
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	prov, err := New(&config.Config{GeminiAPIKey: "k", GeminiServerURL: srv.URL},
		provider.DefaultProviderNameGemini, providerConfig)
	if err != nil {
		t.Fatalf("provider: %v", err)
	}
	if _, err := prov.CallEx(context.Background(), pconfig.OptionsTypeSimple,
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, nil); err != nil {
		t.Fatalf("call: %v", err)
	}

	generation, _ := body["generationConfig"].(map[string]any)
	if generation == nil {
		t.Fatalf("no generation config reached the wire; body=%v", body)
	}
	return generation
}

func TestGemini_CallEx_KeepsTheThinkingBudgetUnderTheOutputCap(t *testing.T) {
	for _, tc := range []struct {
		name        string
		configured  string
		asked       float64
		wantAsIs    bool
		wantClamped bool
	}{
		{name: "a budget under an explicit cap", configured: "  max_tokens: 2048\n", asked: 1024, wantAsIs: true},
		{name: "a budget over an explicit cap", configured: "  max_tokens: 2048\n", asked: 4096, wantClamped: true},
		{name: "no cap configured at all", asked: 4096},
		{name: "a budget larger than any model takes", asked: 32000, wantClamped: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			generation := geminiThinkingCall(t, tc.configured, int(tc.asked))
			thinking, _ := generation["thinkingConfig"].(map[string]any)
			budget, hasBudget := thinking["thinkingBudget"].(float64)
			limit, hasLimit := generation["maxOutputTokens"].(float64)
			if !hasBudget || !hasLimit {
				t.Fatalf("budget %v and output cap %v must both reach the wire", thinking, generation)
			}

			if budget <= 0 {
				t.Errorf("budget on the wire = %v, which asks for no thinking at all", budget)
			}
			if budget >= limit {
				t.Errorf("budget %v is not below the output cap %v, which the API rejects", budget, limit)
			}
			if strings.Contains(tc.configured, "max_tokens") && limit < 2048 {
				t.Errorf("output cap on the wire = %v, below the 2048 the operator configured", limit)
			}
			if tc.wantAsIs && budget != tc.asked {
				t.Errorf("budget on the wire = %v, want the %v asked for: it fits under the cap untouched", budget, tc.asked)
			}
			if tc.wantClamped && budget >= tc.asked {
				t.Errorf("budget on the wire = %v, want less than the %v asked for: it does not fit", budget, tc.asked)
			}
		})
	}
}

func TestGemini_CallEx_KeepsTopPBesideTheThinkingBudget(t *testing.T) {
	generation := geminiThinkingCall(t, "  top_p: 0.3\n", 2048)

	if got := generation["topP"]; got != 0.3 {
		t.Errorf("topP = %v, want 0.3: this door takes sampling alongside thinking", got)
	}
	thinking, _ := generation["thinkingConfig"].(map[string]any)
	if budget, _ := thinking["thinkingBudget"].(float64); budget <= 0 {
		t.Errorf("thinking config on the wire = %v: sampling survived, but the thinking it rides with did not", thinking)
	}
}
